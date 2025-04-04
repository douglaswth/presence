package ifttt

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"goa.design/clue/log"
)

const (
	baseURL      = "https://maker.ifttt.com"
	presentEvent = "presence_detected"
	absentEvent  = "absence_detected"
)

var (
	presentValues = Values{
		Value1: "presence_detected_value1",
		Value2: "presence_detected_value2",
		Value3: "presence_detected_value3",
	}
	absentValues = Values{
		Value1: "absence_detected_value1",
		Value2: "absence_detected_value2",
		Value3: "absence_detected_value3",
	}
)

func TestNewClient(t *testing.T) {
	t.Run("invalid base URL", func(t *testing.T) {
		_, err := NewClient(http.DefaultClient, "%", "key", presentEvent, absentEvent, presentValues, absentValues, false)
		assert.ErrorContains(t, err, `parse "%": invalid URL escape "%"`)
	})
}

func TestClient_Trigger(t *testing.T) {
	ctx := log.Context(context.Background(), log.WithDebug())

	cases := []struct {
		name, key, event, err string
		ctx                   context.Context
		present, noDebug      bool
		values                Values
		handler               func(t *testing.T) http.HandlerFunc
	}{
		{
			name:    "present",
			key:     "key",
			ctx:     ctx,
			present: true,
			values:  presentValues,
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					assert := assert.New(t)

					assert.Equal(http.MethodPost, r.Method)
					assert.Equal("/trigger/"+presentEvent+"/with/key/key", r.URL.Path)
					assert.Equal("application/json", r.Header.Get("Content-Type"))

					body, err := io.ReadAll(r.Body)
					assert.NoError(err)
					assert.JSONEq(`{
						"value1": "presence_detected_value1",
						"value2": "presence_detected_value2",
						"value3": "presence_detected_value3"
					}`, string(body))
				}
			},
			event: presentEvent,
		},
		{
			name:    "absent",
			key:     "key",
			ctx:     ctx,
			present: false,
			values:  absentValues,
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					assert := assert.New(t)

					assert.Equal(http.MethodPost, r.Method)
					assert.Equal("/trigger/"+absentEvent+"/with/key/key", r.URL.Path)
					assert.Equal("application/json", r.Header.Get("Content-Type"))

					body, err := io.ReadAll(r.Body)
					assert.NoError(err)
					assert.JSONEq(`{
						"value1": "absence_detected_value1",
						"value2": "absence_detected_value2",
						"value3": "absence_detected_value3"
					}`, string(body))
				}
			},
			event: absentEvent,
		},
		{
			name: "nil context",
			ctx:  nil,
			key:  "key",
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {}
			},
			err: "net/http: nil Context",
		},
		{
			name: "closed connection",
			ctx:  ctx,
			key:  "key",
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					assert := assert.New(t)

					hj, ok := w.(http.Hijacker)
					if !assert.Equal(true, ok) {
						assert.FailNow("server doesn't support hijacking")
					}

					conn, _, err := hj.Hijack()
					if !assert.NoError(err) {
						assert.FailNow("error hijacking")
					}

					assert.NoError(conn.Close())
				}
			},
			err: `Post "` + baseURL + `/trigger/` + absentEvent + `/with/key/key": EOF`,
		},
		{
			name: "unauthorized",
			ctx:  ctx,
			key:  "key",
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusUnauthorized)
					_, _ = w.Write([]byte(`{"errors":[{"message":"You sent an invalid key."}]}`))
				}
			},
			err: `401 Unauthorized: {"errors":[{"message":"You sent an invalid key."}]}`,
		},
		{
			name: "empty body",
			ctx:  ctx,
			key:  "key",
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}
			},
			err: "500 Internal Server Error: <empty body>",
		},
		{
			name:    "failed to read body",
			ctx:     ctx,
			key:     "key",
			noDebug: true, // goahttp.DebugDoer interferes with this test
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					w.Header().Set("Content-Length", "1")
					w.WriteHeader(http.StatusBadGateway)
				}
			},
			err: "502 Bad Gateway: <failed to read body: unexpected EOF>",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert := assert.New(t)

			ts := httptest.NewTLSServer(tc.handler(t))
			defer ts.Close()

			c, err := NewClient(ts.Client(), ts.URL, tc.key, presentEvent, absentEvent, presentValues, absentValues, !tc.noDebug)
			assert.NoError(err)

			event, values, err := c.Trigger(tc.ctx, tc.present)
			if tc.err != "" {
				tc.err = strings.ReplaceAll(tc.err, baseURL, ts.URL)
				assert.EqualError(err, tc.err)
			} else if assert.NoError(err) {
				assert.Equal(tc.event, event)
				assert.Equal(&tc.values, values)
			}
		})
	}
}

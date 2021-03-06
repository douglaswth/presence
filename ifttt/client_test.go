package ifttt

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"goa.design/clue/log"
)

const (
	presentEvent = "presence_detected"
	absentEvent  = "absence_detected"
)

func TestNewClient(t *testing.T) {
	t.Run("invalid base URL", func(t *testing.T) {
		_, err := NewClient(http.DefaultClient, "%", "key", presentEvent, absentEvent, false)
		assert.ErrorContains(t, err, `parse "%": invalid URL escape "%"`)
	})
}

func TestClient_Trigger(t *testing.T) {
	ctx := log.Context(context.Background(), log.WithDebug())

	cases := []struct {
		name, key, event, err string
		ctx                   context.Context
		present, noDebug      bool
		handler               func(t *testing.T) http.HandlerFunc
	}{
		{
			name:    "preset",
			key:     "key",
			ctx:     ctx,
			present: true,
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					assert := assert.New(t)

					assert.Equal(http.MethodPost, r.Method)
					assert.Equal("/trigger/"+presentEvent+"/with/key/key", r.URL.Path)
				}
			},
			event: presentEvent,
		},
		{
			name:    "absent",
			key:     "key",
			ctx:     ctx,
			present: false,
			handler: func(t *testing.T) http.HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) {
					assert := assert.New(t)

					assert.Equal(http.MethodPost, r.Method)
					assert.Equal("/trigger/"+absentEvent+"/with/key/key", r.URL.Path)
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

					conn.Close()
				}
			},
			err: "EOF",
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

			c, err := NewClient(ts.Client(), ts.URL, tc.key, presentEvent, absentEvent, !tc.noDebug)
			assert.NoError(err)

			event, err := c.Trigger(tc.ctx, tc.present)
			if tc.err != "" {
				assert.ErrorContains(err, tc.err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.event, event)
			}
		})
	}
}

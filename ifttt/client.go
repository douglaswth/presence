package ifttt

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	goahttp "goa.design/goa/v3/http"
)

type (
	Client interface {
		Trigger(ctx context.Context, present bool) (event string, values *Values, err error)
	}

	client struct {
		c                                                *http.Client
		presentEvent, presentURL, absentEvent, absentURL string
		presentValues, absentValues                      *Values
		debug                                            bool
	}

	Values struct {
		Value1 string `json:"value1,omitempty"`
		Value2 string `json:"value2,omitempty"`
		Value3 string `json:"value3,omitempty"`
	}
)

func NewClient(c *http.Client, baseURL, key, presentEvent, absentEvent string, presentValues, absentValues Values, debug bool) (Client, error) {
	presentURL, err := url.JoinPath(baseURL, "trigger", presentEvent, "with/key", key)
	if err != nil {
		return nil, err
	}

	absentURL, err := url.JoinPath(baseURL, "trigger", absentEvent, "with/key", key)
	if err != nil {
		return nil, err
	}

	return &client{
		c:             c,
		presentEvent:  presentEvent,
		presentURL:    presentURL,
		presentValues: &presentValues,
		absentEvent:   absentEvent,
		absentURL:     absentURL,
		absentValues:  &absentValues,
		debug:         debug,
	}, nil
}

func (c *client) Trigger(ctx context.Context, present bool) (string, *Values, error) {
	var (
		event, u string
		values   *Values
	)
	if present {
		event = c.presentEvent
		u = c.presentURL
		values = c.presentValues
	} else {
		event = c.absentEvent
		u = c.absentURL
		values = c.absentValues
	}

	var (
		b = &bytes.Buffer{}
		e = json.NewEncoder(b)
	)
	e.SetEscapeHTML(false)
	if err := e.Encode(values); err != nil {
		return "", nil, err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u, b)
	if err != nil {
		return "", nil, err
	}
	req.Header.Set("Content-Type", "application/json")

	doer := goahttp.Doer(c.c)
	if c.debug {
		doer = goahttp.NewDebugDoer(doer)
	}

	resp, err := doer.Do(req)
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		var b []byte
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			return "", nil, fmt.Errorf("%v: <failed to read body: %w>", resp.Status, err)
		} else if len(b) == 0 {
			b = []byte("<empty body>")
		}

		return "", nil, fmt.Errorf("%v: %s", resp.Status, b)
	}

	return event, values, nil
}

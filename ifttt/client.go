package ifttt

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	goahttp "goa.design/goa/v3/http"
)

type (
	Client interface {
		Trigger(ctx context.Context, present bool) (event string, err error)
	}

	client struct {
		c                         *http.Client
		presentEvent, absentEvent string
		presentURL, absentURL     *url.URL
		debug                     bool
	}
)

func NewClient(c *http.Client, baseURL, key, presentEvent, absentEvent string, debug bool) (Client, error) {
	u, err := url.Parse(baseURL)
	if err != nil {
		return nil, err
	}
	presentURL, absentURL := *u, *u
	presentURL.Path = "/trigger/" + presentEvent + "/with/key/" + key
	absentURL.Path = "/trigger/" + absentEvent + "/with/key/" + key

	return &client{
		c:            c,
		presentEvent: presentEvent,
		absentEvent:  absentEvent,
		presentURL:   &presentURL,
		absentURL:    &absentURL,
		debug:        debug,
	}, nil
}

func (c *client) Trigger(ctx context.Context, present bool) (event string, err error) {
	var u *url.URL
	if present {
		event = c.presentEvent
		u = c.presentURL
	} else {
		event = c.absentEvent
		u = c.absentURL
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, u.String(), nil)
	if err != nil {
		return
	}

	doer := goahttp.Doer(c.c)
	if c.debug {
		doer = goahttp.NewDebugDoer(doer)
	}

	resp, err := doer.Do(req)
	if err != nil {
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		var b []byte
		b, err = io.ReadAll(resp.Body)
		if err != nil {
			err = fmt.Errorf("%v: <failed to read body: %w>", resp.Status, err)
			return
		} else if len(b) == 0 {
			b = []byte("<empty body>")
		}

		err = fmt.Errorf("%v: %s", resp.Status, b)
		return
	}

	return
}

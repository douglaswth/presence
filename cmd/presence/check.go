package main

import (
	"goa.design/clue/log"

	"douglasthrift.net/presence"
	"douglasthrift.net/presence/neighbors"
)

type (
	Check struct {
		Values bool `help:"Show config values." short:"V"`
	}
)

func (c *Check) Run(cli *CLI) (err error) {
	ctx := cli.Context()
	if c.Values {
		_, err = presence.ParseConfigWithContext(ctx, cli.Config, wNet)
	} else {
		_, err = presence.ParseConfig(cli.Config, wNet)
	}
	if err != nil {
		log.Fatal(ctx, err, log.KV{K: "msg", V: "error parsing config"}, log.KV{K: "config", V: cli.Config})
	}

	_, err = neighbors.NewARP(0)
	if err != nil {
		log.Fatal(ctx, err, log.KV{K: "msg", V: "error finding dependencies"})
	}

	return
}

package main

import (
	"goa.design/clue/log"

	"douglasthrift.net/presence"
)

type (
	Check struct {
		Values bool `help:"Show config values." short:"V"`
	}
)

func (c *Check) Run(cli *CLI) error {
	ctx := cli.Context()
	config, err := presence.ParseConfig(cli.Config)
	if err != nil {
		log.Error(ctx, err, log.Fields{"config": cli.Config})
		return err
	}

	if c.Values {
		log.Info(ctx, log.Fields{"interval": config.Interval})

		as := make([]string, len(config.MACAddresses))
		for i, a := range config.MACAddresses {
			as[i] = a.String()
		}
		log.Info(ctx, log.Fields{"mac_addresses": as})

		log.Info(ctx, log.Fields{"ping_count": config.PingCount})
	}
	return nil
}

package main

import (
	"os"
	"os/signal"
	"syscall"
	"time"

	"goa.design/clue/log"

	"douglasthrift.net/presence"
	"douglasthrift.net/presence/neighbors"
)

type (
	Detect struct {
		Iterations uint `help:"Only detect for N iterations." placeholder:"N" short:"i"`
	}
)

func (d *Detect) Run(cli *CLI) error {
	ctx := cli.Context()
	config, err := presence.ParseConfigWithContext(ctx, cli.Config, wNet)
	if err != nil {
		log.Fatal(ctx, err, log.KV{K: "msg", V: "error parsing config"}, log.KV{K: "config", V: cli.Config})
	}

	arp, err := neighbors.NewARP(config.PingCount)
	if err != nil {
		log.Fatal(ctx, err, log.KV{K: "msg", V: "error finding dependencies"})
	}

	var (
		detector = presence.NewDetector(config, arp)
		ticker   = time.NewTicker(config.Interval)
		stop     = make(chan os.Signal, 1)
		reload   = make(chan os.Signal, 1)
		i        uint
	)

	err = detector.Detect(ctx)
	if err != nil {
		log.Error(ctx, err, log.KV{K: "msg", V: "error detecting presence"})
	}

	if d.Iterations != 0 {
		i++
		if i >= d.Iterations {
			ticker.Stop()
			return nil
		}
	}

	signal.Ignore(syscall.SIGHUP)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM)
	signal.Notify(reload, syscall.SIGUSR1)

	for {
		select {
		case <-ticker.C:
			err = detector.Detect(ctx)
			if err != nil {
				log.Error(ctx, err, log.KV{K: "msg", V: "error detecting presence"})
			}

			if d.Iterations != 0 {
				i++
				if i >= d.Iterations {
					ticker.Stop()
					return nil
				}
			}
		case s := <-stop:
			log.Print(ctx, log.Fields{"msg": "received stop signal"}, log.Fields{"signal": s})
			ticker.Stop()
			return nil
		case s := <-reload:
			log.Print(ctx, log.Fields{"msg": "received reload signal"}, log.Fields{"signal": s})
			config, err = presence.ParseConfigWithContext(ctx, cli.Config, wNet)
			if err != nil {
				log.Error(ctx, err, log.KV{K: "msg", V: "error parsing config"}, log.KV{K: "config", V: cli.Config})
			} else {
				arp.Count(config.PingCount)
				detector.Config(config)
				ticker.Reset(config.Interval)
			}
		}
	}
}

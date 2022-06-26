package main

import (
	"context"

	"goa.design/clue/log"

	"douglasthrift.net/presence"
)

type (
	Detect struct {
		Interface     string   `arg:""`
		HardwareAddrs []string `arg:""`
	}
)

func (d *Detect) Run(cli *CLI) error {
	ifs := presence.Interfaces{d.Interface: true}
	hws := make(presence.HardwareAddrStates, len(d.HardwareAddrs))
	for _, hw := range d.HardwareAddrs {
		hws[hw] = presence.NewState()
	}

	ctx := log.Context(context.Background(), log.WithDisableBuffering(func(context.Context) bool { return true }))
	a, err := presence.NewARP(1)
	if err != nil {
		return err
	}

	ok, err := a.Present(ctx, ifs, hws)
	if err != nil {
		return err
	}
	log.Info(ctx, log.KV{K: "present", V: ok})
	for hw, state := range hws {
		log.Info(ctx, log.KV{K: "hw", V: hw}, log.KV{K: "present", V: state.Present()}, log.KV{K: "changed", V: state.Changed()})
	}
	return nil
}

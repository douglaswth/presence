package main

import (
	"goa.design/clue/log"

	"douglasthrift.net/presence/neighbors"
)

type (
	Detect struct {
		Interface     string   `arg:""`
		HardwareAddrs []string `arg:""`
	}
)

func (d *Detect) Run(cli *CLI) error {
	ctx := cli.Context()

	ifs := neighbors.Interfaces{d.Interface: true}
	hws := make(neighbors.HardwareAddrStates, len(d.HardwareAddrs))
	for _, hw := range d.HardwareAddrs {
		hws[hw] = neighbors.NewState()
	}

	a, err := neighbors.NewARP(1)
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

package presence

import (
	"context"

	"goa.design/clue/log"

	"douglasthrift.net/presence/ifttt"
	"douglasthrift.net/presence/neighbors"
)

type (
	Detector interface {
		Detect(ctx context.Context) error
		Config(config *Config)
		Client(client ifttt.Client)
	}

	detector struct {
		config     *Config
		arp        neighbors.ARP
		interfaces neighbors.Interfaces
		state      neighbors.State
		states     neighbors.HardwareAddrStates
		client     ifttt.Client
	}
)

func NewDetector(config *Config, arp neighbors.ARP, client ifttt.Client) Detector {
	d := &detector{
		arp:    arp,
		state:  neighbors.NewState(),
		states: make(neighbors.HardwareAddrStates, len(config.MACAddresses)),
		client: client,
	}
	d.Config(config)
	return d
}

func (d *detector) Detect(ctx context.Context) error {
	log.Print(ctx, log.KV{K: "msg", V: "detecting presence"}, log.KV{K: "present", V: d.state.Present()})
	err := d.arp.Present(ctx, d.interfaces, d.state, d.states)
	if err != nil {
		return err
	}

	for _, a := range d.config.MACAddresses {
		state := d.states[a]
		log.Print(ctx, log.KV{K: "msg", V: a}, log.KV{K: "present", V: state.Present()}, log.KV{K: "changed", V: state.Changed()})
	}

	log.Print(ctx, log.KV{K: "msg", V: "detected presence"}, log.KV{K: "present", V: d.state.Present()}, log.KV{K: "changed", V: d.state.Changed()})
	if d.state.Changed() {
		event, err := d.client.Trigger(ctx, d.state.Present())
		if err != nil {
			d.state.Reset()
			return err
		}
		log.Print(ctx, log.KV{K: "msg", V: "triggered IFTTT"}, log.KV{K: "event", V: event})
	}

	return nil
}

func (d *detector) Config(config *Config) {
	d.config = config
	d.interfaces = make(neighbors.Interfaces, len(config.Interfaces))
	for _, i := range config.Interfaces {
		d.interfaces[i] = true
	}

	states := make(map[string]bool, len(d.states))
	for a := range d.states {
		states[a] = true
	}
	for _, a := range config.MACAddresses {
		if states[a] {
			states[a] = false
		} else {
			d.states[a] = neighbors.NewState()
		}
	}
	for a, ok := range states {
		if ok {
			delete(d.states, a)
		}
	}
}

func (d *detector) Client(client ifttt.Client) {
	d.client = client
}

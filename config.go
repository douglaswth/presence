package presence

import (
	"context"
	"fmt"
	"net"
	"os"
	"time"

	"goa.design/clue/log"
	"gopkg.in/yaml.v3"

	"douglasthrift.net/presence/wrap"
)

type (
	Config struct {
		Interval     time.Duration `yaml:"interval"`
		Interfaces   []string      `yaml:"interfaces"`
		MACAddresses []string      `yaml:"mac_addresses"`
		PingCount    uint          `yaml:"ping_count"`
	}
)

func ParseConfig(name string, wNet wrap.Net) (*Config, error) {
	return ParseConfigWithContext(context.Background(), name, wNet)
}

func ParseConfigWithContext(ctx context.Context, name string, wNet wrap.Net) (*Config, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	d := yaml.NewDecoder(f)
	d.KnownFields(true)

	c := &Config{}
	err = d.Decode(c)
	if err != nil {
		return nil, err
	}

	if c.Interval < 0 {
		return nil, fmt.Errorf("negative interval (%v)", c.Interval)
	} else if c.Interval == 0 {
		c.Interval = 30 * time.Second
	}

	if len(c.Interfaces) == 0 {
		ifs, err := wNet.Interfaces()
		if err != nil {
			return nil, err
		}

		c.Interfaces = make([]string, 0, len(ifs))
		for _, i := range ifs {
			c.Interfaces = append(c.Interfaces, i.Name)
		}
	} else {
		for _, i := range c.Interfaces {
			_, err = wNet.InterfaceByName(i)
			if err != nil {
				return nil, fmt.Errorf("interface %v: %w", i, err)
			}
		}
	}

	if len(c.MACAddresses) == 0 {
		return nil, fmt.Errorf("no MAC addresses")
	}
	as := make(map[string]bool, len(c.MACAddresses))
	for i, a := range c.MACAddresses {
		hw, err := net.ParseMAC(a)
		if err != nil {
			return nil, err
		}

		a = hw.String()
		if as[a] {
			return nil, fmt.Errorf("duplicate MAC address (%v)", a)
		}
		as[a] = true
		c.MACAddresses[i] = a
	}

	if c.PingCount == 0 {
		c.PingCount = 1
	}

	log.Print(ctx, log.KV{K: "msg", V: "interval"}, log.KV{K: "value", V: c.Interval})
	log.Print(ctx, log.KV{K: "msg", V: "interfaces"}, log.KV{K: "value", V: c.Interfaces})
	log.Print(ctx, log.KV{K: "msg", V: "MAC addresses"}, log.KV{K: "value", V: c.MACAddresses})
	log.Print(ctx, log.KV{K: "msg", V: "ping count"}, log.KV{K: "value", V: c.PingCount})

	return c, nil
}

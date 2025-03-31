package presence

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"regexp"
	"strings"
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
		IFTTT        IFTTT         `yaml:"ifttt"`
	}

	IFTTT struct {
		BaseURL string `yaml:"base_url"`
		Key     string `yaml:"key"`
		Events  Events `yaml:"events"`
	}

	Events struct {
		Present Event `yaml:"present"`
		Absent  Event `yaml:"absent"`
	}

	Event struct {
		Event  string `yaml:"event"`
		Value1 string `yaml:"value1"`
		Value2 string `yaml:"value2"`
		Value3 string `yaml:"value3"`
	}
)

const (
	defaultBaseURL      = "https://maker.ifttt.com"
	defaultPresentEvent = "presence_detected"
	defaultAbsentEvent  = "absence_detected"
)

var (
	eventName = regexp.MustCompile("^[_a-zA-Z]+$")
)

func ParseConfig(name string, wNet wrap.Net) (*Config, error) {
	return ParseConfigWithContext(context.Background(), name, wNet)
}

func ParseConfigWithContext(ctx context.Context, name string, wNet wrap.Net) (*Config, error) {
	f, err := os.Open(name)
	if err != nil {
		return nil, err
	}
	defer func() { _ = f.Close() }()

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
	log.Print(ctx, log.KV{K: "msg", V: "interval"}, log.KV{K: "value", V: c.Interval})

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
	log.Print(ctx, log.KV{K: "msg", V: "interfaces"}, log.KV{K: "value", V: c.Interfaces})

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
	log.Print(ctx, log.KV{K: "msg", V: "MAC addresses"}, log.KV{K: "value", V: c.MACAddresses})

	if c.PingCount == 0 {
		c.PingCount = 1
	}
	log.Print(ctx, log.KV{K: "msg", V: "ping count"}, log.KV{K: "value", V: c.PingCount})

	if c.IFTTT.BaseURL == "" {
		c.IFTTT.BaseURL = defaultBaseURL
	} else if _, err := url.Parse(c.IFTTT.BaseURL); err != nil {
		return nil, fmt.Errorf("IFTTT base URL: %w", err)
	}
	log.Print(ctx, log.KV{K: "msg", V: "IFTTT base URL"}, log.KV{K: "value", V: c.IFTTT.BaseURL})

	if c.IFTTT.Key == "" {
		return nil, fmt.Errorf("no IFTTT key")
	}
	log.Print(ctx, log.KV{K: "msg", V: "IFTTT key"}, log.KV{K: "value", V: strings.Repeat("*", len(c.IFTTT.Key))})

	if c.IFTTT.Events.Present.Event == "" {
		c.IFTTT.Events.Present.Event = defaultPresentEvent
	} else if !eventName.MatchString(c.IFTTT.Events.Present.Event) {
		return nil, fmt.Errorf("invalid IFTTT present event name: %#v", c.IFTTT.Events.Present.Event)
	}
	log.Print(ctx, log.KV{K: "msg", V: "IFTTT present event"}, log.KV{K: "value", V: c.IFTTT.Events.Present.Event},
		log.KV{K: "value1", V: c.IFTTT.Events.Present.Value1},
		log.KV{K: "value2", V: c.IFTTT.Events.Present.Value2},
		log.KV{K: "value3", V: c.IFTTT.Events.Present.Value3})

	if c.IFTTT.Events.Absent.Event == "" {
		c.IFTTT.Events.Absent.Event = defaultAbsentEvent
	} else if !eventName.MatchString(c.IFTTT.Events.Absent.Event) {
		return nil, fmt.Errorf("invalid IFTTT absent event name: %#v", c.IFTTT.Events.Absent.Event)
	}
	log.Print(ctx, log.KV{K: "msg", V: "IFTTT absent event"}, log.KV{K: "value", V: c.IFTTT.Events.Absent.Event},
		log.KV{K: "value1", V: c.IFTTT.Events.Absent.Value1},
		log.KV{K: "value2", V: c.IFTTT.Events.Absent.Value2},
		log.KV{K: "value3", V: c.IFTTT.Events.Absent.Value3})

	return c, nil
}

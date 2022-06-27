package presence

import (
	"fmt"
	"net"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type (
	MACAddress struct {
		net.HardwareAddr
	}

	Config struct {
		Interval     time.Duration `yaml:"interval"`
		MACAddresses []MACAddress  `yaml:"mac_addresses"`
		PingCount    uint          `yaml:"ping_count"`
	}
)

func ParseConfig(name string) (*Config, error) {
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
	}
	if c.Interval == 0 {
		c.Interval = 30 * time.Second
	}
	if c.PingCount == 0 {
		c.PingCount = 1
	}
	return c, nil
}

func (a *MACAddress) UnmarshalYAML(value *yaml.Node) (err error) {
	var s string
	err = value.Decode(&s)
	if err != nil {
		return
	}

	a.HardwareAddr, err = net.ParseMAC(s)
	return
}

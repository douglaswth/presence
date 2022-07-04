package presence

import (
	"fmt"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	mockwrap "douglasthrift.net/presence/wrap/mocks"
)

func TestParseConfig(t *testing.T) {
	cases := []struct {
		name, file string
		setup      func(wNet *mockwrap.Net)
		config     *Config
		err        string
	}{
		{
			name: "success",
			file: "success.yml",
			setup: func(wNet *mockwrap.Net) {
				wNet.Mock.On("InterfaceByName", "eth0").Return(&net.Interface{}, nil)
				wNet.Mock.On("InterfaceByName", "eth1").Return(&net.Interface{}, nil)
			},
			config: &Config{
				Interval:     1 * time.Minute,
				Interfaces:   []string{"eth0", "eth1"},
				MACAddresses: []string{"00:00:00:00:00:0a", "00:00:00:00:00:0b"},
				PingCount:    5,
			},
		},
		{
			name: "defaults",
			file: "defaults.yml",
			setup: func(wNet *mockwrap.Net) {
				wNet.Mock.On("Interfaces").Return([]net.Interface{{Name: "eth0"}, {Name: "eth1"}, {Name: "lo"}}, nil)
			},
			config: &Config{
				Interval:     30 * time.Second,
				Interfaces:   []string{"eth0", "eth1", "lo"},
				MACAddresses: []string{"00:00:00:00:00:01", "00:00:00:00:00:02"},
				PingCount:    1,
			},
		},
		{
			name: "nonexistent file",
			file: "nonexistent.yml",
			err:  "open tests/nonexistent.yml: no such file or directory",
		},
		{
			name: "wrong MAC encoding",
			file: "wrong_mac_encoding.yml",
			err:  "yaml: unmarshal errors:\n  line 2: cannot unmarshal !!seq into string",
		},
		{
			name: "bad MAC encoding",
			file: "bad_mac_encoding.yml",
			setup: func(wNet *mockwrap.Net) {
				wNet.Mock.On("Interfaces").Return([]net.Interface{{Name: "eth0"}, {Name: "eth1"}, {Name: "lo"}}, nil)
			},
			err: "address 00-00-00-00-00-0x: invalid MAC address",
		},
		{
			name: "negative interval",
			file: "negative_interval.yml",
			err:  "negative interval (-1ns)",
		},
		{
			name: "error listing interfaces",
			file: "defaults.yml",
			setup: func(wNet *mockwrap.Net) {
				wNet.Mock.On("Interfaces").Return(nil, fmt.Errorf("no network interfaces"))
			},
			err: "no network interfaces",
		},
		{
			name: "error getting interface by name",
			file: "success.yml",
			setup: func(wNet *mockwrap.Net) {
				wNet.Mock.On("InterfaceByName", "eth0").Return(nil, fmt.Errorf("no such network interface"))
			},
			err: "interface eth0: no such network interface",
		},
		{
			name: "no MAC addresses",
			file: "no_mac_addresses.yml",
			setup: func(wNet *mockwrap.Net) {
				wNet.Mock.On("InterfaceByName", "eth0").Return(&net.Interface{}, nil)
			},
			err: "no MAC addresses",
		},
		{
			name: "duplicate MAC address",
			file: "duplicate_mac_address.yml",
			setup: func(wNet *mockwrap.Net) {
				wNet.Mock.On("InterfaceByName", "eth0").Return(&net.Interface{}, nil)
			},
			err: "duplicate MAC address (00:00:00:00:00:0e)",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert := assert.New(t)

			wNet := mockwrap.NewNet(t)
			if tc.setup != nil {
				tc.setup(wNet)
			}

			c, err := ParseConfig(filepath.Join("tests", tc.file), wNet)
			if tc.err != "" {
				assert.ErrorContains(err, tc.err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.config, c)
			}
		})
	}
}

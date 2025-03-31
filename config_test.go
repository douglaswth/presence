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
		setup      func(t *testing.T, wNet *mockwrap.Net)
		config     *Config
		err        string
	}{
		{
			name: "success",
			file: "success.yml",
			setup: func(t *testing.T, wNet *mockwrap.Net) {
				wNet.AddInterfaceByName(func(name string) (*net.Interface, error) {
					assert.Equal(t, "eth0", name)
					return &net.Interface{}, nil
				})
				wNet.AddInterfaceByName(func(name string) (*net.Interface, error) {
					assert.Equal(t, "eth1", name)
					return &net.Interface{}, nil
				})
			},
			config: &Config{
				Interval:     1 * time.Minute,
				Interfaces:   []string{"eth0", "eth1"},
				MACAddresses: []string{"00:00:00:00:00:0a", "00:00:00:00:00:0b"},
				PingCount:    5,
				IFTTT: IFTTT{
					BaseURL: "https://example.com",
					Key:     "abcdef123456",
					Events: Events{
						Present: Event{
							Event:  "event_presence_detected",
							Value1: "event_presence_detected_value1",
							Value2: "event_presence_detected_value2",
							Value3: "event_presence_detected_value3",
						},
						Absent: Event{
							Event:  "event_absence_detected",
							Value1: "event_absence_detected_value1",
							Value2: "event_absence_detected_value2",
							Value3: "event_absence_detected_value3",
						},
					},
				},
			},
		},
		{
			name: "defaults",
			file: "defaults.yml",
			setup: func(t *testing.T, wNet *mockwrap.Net) {
				wNet.AddInterfaces(func() ([]net.Interface, error) {
					return []net.Interface{{Name: "eth0"}, {Name: "eth1"}, {Name: "lo"}}, nil
				})
			},
			config: &Config{
				Interval:     30 * time.Second,
				Interfaces:   []string{"eth0", "eth1", "lo"},
				MACAddresses: []string{"00:00:00:00:00:01", "00:00:00:00:00:02"},
				PingCount:    1,
				IFTTT: IFTTT{
					BaseURL: defaultBaseURL,
					Key:     "xyz7890!@#",
					Events: Events{
						Present: Event{Event: defaultPresentEvent},
						Absent:  Event{Event: defaultAbsentEvent},
					},
				},
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
			setup: func(t *testing.T, wNet *mockwrap.Net) {
				wNet.AddInterfaces(func() ([]net.Interface, error) {
					return []net.Interface{{Name: "eth0"}, {Name: "eth1"}, {Name: "lo"}}, nil
				})
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
			setup: func(t *testing.T, wNet *mockwrap.Net) {
				wNet.AddInterfaces(func() ([]net.Interface, error) {
					return nil, fmt.Errorf("no network interfaces")
				})
			},
			err: "no network interfaces",
		},
		{
			name: "error getting interface by name",
			file: "success.yml",
			setup: func(t *testing.T, wNet *mockwrap.Net) {
				wNet.AddInterfaceByName(func(name string) (*net.Interface, error) {
					assert.Equal(t, "eth0", name)
					return nil, fmt.Errorf("no such network interface")
				})
			},
			err: "interface eth0: no such network interface",
		},
		{
			name: "no MAC addresses",
			file: "no_mac_addresses.yml",
			setup: func(t *testing.T, wNet *mockwrap.Net) {
				wNet.AddInterfaceByName(func(name string) (*net.Interface, error) {
					assert.Equal(t, "eth0", name)
					return &net.Interface{}, nil
				})
			},
			err: "no MAC addresses",
		},
		{
			name: "duplicate MAC address",
			file: "duplicate_mac_address.yml",
			setup: func(t *testing.T, wNet *mockwrap.Net) {
				wNet.AddInterfaceByName(func(name string) (*net.Interface, error) {
					assert.Equal(t, "eth0", name)
					return &net.Interface{}, nil
				})
			},
			err: "duplicate MAC address (00:00:00:00:00:0e)",
		},
		{
			name: "invalid IFTTT base URL",
			file: "invalid_ifttt_base_url.yml",
			setup: func(t *testing.T, wNet *mockwrap.Net) {
				wNet.AddInterfaceByName(func(name string) (*net.Interface, error) {
					assert.Equal(t, "eth0", name)
					return &net.Interface{}, nil
				})
			},
			err: `IFTTT base URL: parse "%": invalid URL escape "%"`,
		},
		{
			name: "no IFTTT key",
			file: "no_ifttt_key.yml",
			setup: func(t *testing.T, wNet *mockwrap.Net) {
				wNet.AddInterfaceByName(func(name string) (*net.Interface, error) {
					assert.Equal(t, "eth0", name)
					return &net.Interface{}, nil
				})
			},
			err: "no IFTTT key",
		},
		{
			name: "invalid IFTTT present event name",
			file: "invalid_ifttt_present_event_name.yml",
			setup: func(t *testing.T, wNet *mockwrap.Net) {
				wNet.AddInterfaceByName(func(name string) (*net.Interface, error) {
					assert.Equal(t, "eth0", name)
					return &net.Interface{}, nil
				})
			},
			err: `invalid IFTTT present event name: "$"`,
		},
		{
			name: "invalid IFTTT absent event name",
			file: "invalid_ifttt_absent_event_name.yml",
			setup: func(t *testing.T, wNet *mockwrap.Net) {
				wNet.AddInterfaceByName(func(name string) (*net.Interface, error) {
					assert.Equal(t, "eth0", name)
					return &net.Interface{}, nil
				})
			},
			err: `invalid IFTTT absent event name: "^"`,
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert := assert.New(t)

			wNet := mockwrap.NewNet(t)
			if tc.setup != nil {
				tc.setup(t, wNet)
			}

			c, err := ParseConfig(filepath.Join("tests", tc.file), wNet)
			if tc.err != "" {
				assert.ErrorContains(err, tc.err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.config, c)
			}

			assert.False(wNet.HasMore(), "missing expected net calls")
		})
	}
}

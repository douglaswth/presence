package presence

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"goa.design/clue/log"

	"douglasthrift.net/presence/ifttt"
	"douglasthrift.net/presence/neighbors"
	mockifttt "douglasthrift.net/presence/ifttt/mocks"
	mockneighbors "douglasthrift.net/presence/neighbors/mocks"
)

func TestDetect(t *testing.T) {
	ctx := log.Context(context.Background(), log.WithDebug())

	const mac = "00:00:00:00:00:01"

	cases := []struct {
		name   string
		config *Config
		setup  func(t *testing.T, d *detector, arp *mockneighbors.ARP, client *mockifttt.Client)
		err    string
	}{
		{
			name: "arp error",
			config: &Config{
				Interval:     30 * time.Second,
				Interfaces:   []string{"eth0"},
				MACAddresses: []string{mac},
				PingCount:    1,
			},
			setup: func(t *testing.T, d *detector, arp *mockneighbors.ARP, client *mockifttt.Client) {
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					return fmt.Errorf("arp failed")
				})
			},
			err: "arp failed",
		},
		{
			name: "state changed triggers ifttt",
			config: &Config{
				Interval:     30 * time.Second,
				Interfaces:   []string{"eth0"},
				MACAddresses: []string{mac},
				PingCount:    1,
			},
			setup: func(t *testing.T, d *detector, arp *mockneighbors.ARP, client *mockifttt.Client) {
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					assert.True(t, present)
					return "present", &ifttt.Values{}, nil
				})
			},
		},
		{
			name: "state changed to absent triggers ifttt",
			config: &Config{
				Interval:     30 * time.Second,
				Interfaces:   []string{"eth0"},
				MACAddresses: []string{mac},
				PingCount:    1,
			},
			setup: func(t *testing.T, d *detector, arp *mockneighbors.ARP, client *mockifttt.Client) {
				// First detect: become present
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					assert.True(t, present)
					return "present", &ifttt.Values{}, nil
				})
				assert.NoError(t, d.Detect(ctx))

				// Setup for the test's detect: become absent
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(false)
					}
					state.Set(false)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					assert.False(t, present)
					return "absent", &ifttt.Values{}, nil
				})
			},
		},
		{
			name: "state changed trigger error resets state",
			config: &Config{
				Interval:     30 * time.Second,
				Interfaces:   []string{"eth0"},
				MACAddresses: []string{mac},
				PingCount:    1,
			},
			setup: func(t *testing.T, d *detector, arp *mockneighbors.ARP, client *mockifttt.Client) {
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					return "", nil, fmt.Errorf("trigger failed")
				})
			},
			err: "trigger failed",
		},
		{
			name: "state reset after trigger error allows retrigger",
			config: &Config{
				Interval:     30 * time.Second,
				Interfaces:   []string{"eth0"},
				MACAddresses: []string{mac},
				PingCount:    1,
			},
			setup: func(t *testing.T, d *detector, arp *mockneighbors.ARP, client *mockifttt.Client) {
				// First detect: trigger fails, state resets
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					return "", nil, fmt.Errorf("trigger failed")
				})
				assert.ErrorContains(t, d.Detect(ctx), "trigger failed")

				// Setup for the test's detect: same presence triggers again after reset
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					assert.True(t, present)
					return "present", &ifttt.Values{}, nil
				})
			},
		},
		{
			name: "state not changed no trigger",
			config: &Config{
				Interval:     30 * time.Second,
				Interfaces:   []string{"eth0"},
				MACAddresses: []string{mac},
				PingCount:    1,
			},
			setup: func(t *testing.T, d *detector, arp *mockneighbors.ARP, client *mockifttt.Client) {
				// First detect: state changes, trigger fires
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					return "present", &ifttt.Values{}, nil
				})
				assert.NoError(t, d.Detect(ctx))

				// Setup for the test's detect: same state, no trigger
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
			},
		},
		{
			name: "retrigger after elapsed",
			config: &Config{
				Interval:       30 * time.Second,
				RetriggerAfter: time.Hour,
				Interfaces:     []string{"eth0"},
				MACAddresses:   []string{mac},
				PingCount:      1,
			},
			setup: func(t *testing.T, d *detector, arp *mockneighbors.ARP, client *mockifttt.Client) {
				// First detect: state changes, trigger fires, lastChange set
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					return "present", &ifttt.Values{}, nil
				})
				assert.NoError(t, d.Detect(ctx))

				// Simulate time elapsed
				d.lastChange = time.Now().Add(-2 * time.Hour)

				// Setup for the test's detect: retrigger fires
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					assert.True(t, present)
					return "present", &ifttt.Values{}, nil
				})
			},
		},
		{
			name: "retrigger after not elapsed",
			config: &Config{
				Interval:       30 * time.Second,
				RetriggerAfter: time.Hour,
				Interfaces:     []string{"eth0"},
				MACAddresses:   []string{mac},
				PingCount:      1,
			},
			setup: func(t *testing.T, d *detector, arp *mockneighbors.ARP, client *mockifttt.Client) {
				// First detect: state changes, trigger fires, lastChange set
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					return "present", &ifttt.Values{}, nil
				})
				assert.NoError(t, d.Detect(ctx))

				// lastChange was just set, so time hasn't elapsed
				// Setup for the test's detect: no retrigger
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
			},
		},
		{
			name: "retrigger after trigger error",
			config: &Config{
				Interval:       30 * time.Second,
				RetriggerAfter: time.Hour,
				Interfaces:     []string{"eth0"},
				MACAddresses:   []string{mac},
				PingCount:      1,
			},
			setup: func(t *testing.T, d *detector, arp *mockneighbors.ARP, client *mockifttt.Client) {
				// First detect: state changes, trigger fires
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					return "present", &ifttt.Values{}, nil
				})
				assert.NoError(t, d.Detect(ctx))

				// Simulate time elapsed
				d.lastChange = time.Now().Add(-2 * time.Hour)

				// Setup for the test's detect: retrigger fails
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					return "", nil, fmt.Errorf("retrigger failed")
				})
			},
			err: "retrigger failed",
		},
		{
			name: "retrigger disabled no lastChange set",
			config: &Config{
				Interval:     30 * time.Second,
				Interfaces:   []string{"eth0"},
				MACAddresses: []string{mac},
				PingCount:    1,
			},
			setup: func(t *testing.T, d *detector, arp *mockneighbors.ARP, client *mockifttt.Client) {
				// Detect: state changes, trigger fires, but RetriggerAfter=0
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
				client.AddTrigger(func(ctx context.Context, present bool) (string, *ifttt.Values, error) {
					return "present", &ifttt.Values{}, nil
				})
				assert.NoError(t, d.Detect(ctx))
				assert.True(t, d.lastChange.IsZero())

				// Setup for the test's detect: same state, no trigger
				arp.AddPresent(func(ctx context.Context, ifs neighbors.Interfaces, state neighbors.State, addrStates neighbors.HardwareAddrStates) error {
					for _, s := range addrStates {
						s.Set(true)
					}
					state.Set(true)
					return nil
				})
			},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert := assert.New(t)

			arp := mockneighbors.NewARP(t)
			client := mockifttt.NewClient(t)
			d := NewDetector(tc.config, arp, client)

			if tc.setup != nil {
				tc.setup(t, d.(*detector), arp, client)
			}

			err := d.Detect(ctx)
			if tc.err != "" {
				assert.ErrorContains(err, tc.err)
			} else {
				assert.NoError(err)
			}

			assert.False(arp.HasMore(), "missing expected arp calls")
			assert.False(client.HasMore(), "missing expected client calls")
		})
	}
}

func TestDetector_Config(t *testing.T) {
	const (
		mac1 = "00:00:00:00:00:01"
		mac2 = "00:00:00:00:00:02"
		mac3 = "00:00:00:00:00:03"
	)

	cases := []struct {
		name      string
		initial   *Config
		updated   *Config
		kept      []string
		added     []string
		removed   []string
	}{
		{
			name: "keep existing mac add new",
			initial: &Config{
				Interfaces:   []string{"eth0"},
				MACAddresses: []string{mac1},
			},
			updated: &Config{
				Interfaces:   []string{"eth0"},
				MACAddresses: []string{mac1, mac2},
			},
			kept:  []string{mac1},
			added: []string{mac2},
		},
		{
			name: "remove old mac",
			initial: &Config{
				Interfaces:   []string{"eth0"},
				MACAddresses: []string{mac1, mac2},
			},
			updated: &Config{
				Interfaces:   []string{"eth0"},
				MACAddresses: []string{mac1},
			},
			kept:    []string{mac1},
			removed: []string{mac2},
		},
		{
			name: "replace all macs",
			initial: &Config{
				Interfaces:   []string{"eth0"},
				MACAddresses: []string{mac1},
			},
			updated: &Config{
				Interfaces:   []string{"eth0", "eth1"},
				MACAddresses: []string{mac2, mac3},
			},
			added:   []string{mac2, mac3},
			removed: []string{mac1},
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert := assert.New(t)

			arp := mockneighbors.NewARP(t)
			client := mockifttt.NewClient(t)
			d := NewDetector(tc.initial, arp, client).(*detector)

			// Capture initial states for kept MACs
			initialStates := make(map[string]neighbors.State)
			for _, a := range tc.kept {
				initialStates[a] = d.states[a]
			}

			d.Config(tc.updated)

			assert.Equal(tc.updated, d.config)
			for _, i := range tc.updated.Interfaces {
				assert.True(d.interfaces[i])
			}
			for _, a := range tc.kept {
				assert.Equal(initialStates[a], d.states[a], "kept MAC state should be preserved")
			}
			for _, a := range tc.added {
				assert.NotNil(d.states[a], "added MAC should have state")
			}
			for _, a := range tc.removed {
				_, exists := d.states[a]
				assert.False(exists, "removed MAC should be deleted")
			}
		})
	}
}

func TestDetector_Client(t *testing.T) {
	arp := mockneighbors.NewARP(t)
	client1 := mockifttt.NewClient(t)
	client2 := mockifttt.NewClient(t)

	config := &Config{
		Interfaces:   []string{"eth0"},
		MACAddresses: []string{"00:00:00:00:00:01"},
	}
	d := NewDetector(config, arp, client1).(*detector)
	assert.Equal(t, client1, d.client)

	d.Client(client2)
	assert.Equal(t, client2, d.client)
}

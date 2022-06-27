package presence

import (
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestParseConfig(t *testing.T) {
	cases := []struct {
		name, file string
		config     *Config
		err        string
	}{
		{
			name: "success",
			file: "success.yml",
			config: &Config{
				Interval: 1 * time.Minute,
				MACAddresses: []MACAddress{
					{net.HardwareAddr{0, 0, 0, 0, 0, 0xa}},
					{net.HardwareAddr{0, 0, 0, 0, 0, 0xb}},
				},
				PingCount: 5,
			},
		},
		{
			name: "defaults",
			file: "defaults.yml",
			config: &Config{
				Interval: 30 * time.Second,
				MACAddresses: []MACAddress{
					{net.HardwareAddr{0, 0, 0, 0, 0, 1}},
					{net.HardwareAddr{0, 0, 0, 0, 0, 2}},
				},
				PingCount: 1,
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
			err:  "address 00-00-00-00-00-0x: invalid MAC address",
		},
	}

	for _, tc := range cases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			assert := assert.New(t)

			c, err := ParseConfig(filepath.Join("tests", tc.file))
			if tc.err != "" {
				assert.ErrorContains(err, tc.err)
			} else {
				assert.NoError(err)
				assert.Equal(tc.config, c)
			}
		})
	}
}

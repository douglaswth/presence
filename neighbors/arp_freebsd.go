package neighbors

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"
)

const (
	arpOutputVersion = "1"
)

type (
	arp struct {
		cmd    string
		arping ARPing
	}

	arpOutput struct {
		Version string `json:"__version"`
		ARP     struct {
			Cache []arpEntry `json:"arp-cache"`
		} `json:"arp"`
	}

	arpEntry struct {
		IPAddress  string `json:"ip-address"`
		MACAddress string `json:"mac-address"`
		Interface  string `json:"interface"`
	}
)

func NewARP(count uint) (ARP, error) {
	cmd, err := exec.LookPath("arp")
	if err != nil {
		return nil, err
	}

	arping, err := NewARPing(count)
	if err != nil {
		return nil, err
	}

	return &arp{cmd: cmd, arping: arping}, nil
}

func (a *arp) Present(ctx context.Context, ifs map[string]bool, hws map[string]bool) (ok bool, err error) {
	cmd := exec.CommandContext(ctx, a.cmd, "--libxo=json", "-an")
	b, err := cmd.Output()
	if err != nil {
		return
	}

	o := &arpOutput{}
	err = json.Unmarshal(b, o)
	if err != nil {
		return
	}

	if o.Version != arpOutputVersion {
		err = fmt.Errorf("arp output version mismatch (got %v, expected %v)", o.Version, arpOutputVersion)
		return
	}

	for _, e := range o.ARP.Cache {
		if ifs[e.Interface] {
			var hwa net.HardwareAddr
			hwa, err = net.ParseMAC(e.MACAddress)
			if err != nil {
				return
			}
			hw := hwa.String()

			if hws[hw] {
				ok, err = a.arping.Ping(ctx, e.Interface, hw, e.IPAddress)
				if ok || err != nil {
					return
				}
			}
		}
	}

	return
}

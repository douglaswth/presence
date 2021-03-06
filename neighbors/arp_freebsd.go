package neighbors

import (
	"context"
	"encoding/json"
	"fmt"
	"net"
	"os/exec"

	"goa.design/clue/log"
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

	return &arp{
		cmd:    cmd,
		arping: arping,
	}, nil
}

func (a *arp) Present(ctx context.Context, ifs Interfaces, state State, addrStates HardwareAddrStates) (err error) {
	as := make(map[string]bool, len(addrStates))
	for hw := range addrStates {
		as[hw] = false
	}

	cmd := exec.CommandContext(ctx, a.cmd, "--libxo=json", "-an")
	log.Debug(ctx, log.KV{K: "cmd", V: cmd})
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
		log.Debug(ctx, log.KV{K: "IP address", V: e.IPAddress}, log.KV{K: "MAC address", V: e.MACAddress}, log.KV{K: "interface", V: e.Interface})
		if ifs[e.Interface] {
			var hwa net.HardwareAddr
			hwa, err = net.ParseMAC(e.MACAddress)
			if err != nil {
				return
			}
			hw := hwa.String()

			if _, ok := as[hw]; ok {
				ok, err = a.arping.Ping(ctx, e.Interface, hw, e.IPAddress)
				if err != nil {
					return
				}
				as[hw] = ok
			}
		}
	}

	present := false
	for hw, ok := range as {
		addrStates[hw].Set(ok)
		if ok {
			present = true
		}
	}
	state.Set(present)

	return
}

func (a *arp) Count(count uint) {
	a.arping.Count(count)
}

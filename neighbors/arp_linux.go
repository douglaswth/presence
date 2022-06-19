package neighbors

import (
	"context"
	"encoding/json"
	"net"
	"os/exec"
)

type (
	arp struct {
		cmd    string
		arping ARPing
	}

	arpEntry struct {
		IPAddress  string `json:"dst"`
		MACAddress string `json:"lladdr"`
		Interface  string `json:"dev"`
	}
)

func NewARP(count uint) (ARP, error) {
	cmd, err := exec.LookPath("ip")
	if err != nil {
		return nil, err
	}

	arping, err := NewARPing(count)
	if err != nil {
		return nil, err
	}

	return &arp{cmd: cmd, arping: arping}, nil
}

func (a *arp) Present(ctx context.Context, ifs Interfaces, hws HardwareAddrStates) (present bool, err error) {
	as := make(map[string]bool, len(hws))
	for hw := range hws {
		as[hw] = false
	}

	cmd := exec.CommandContext(ctx, a.cmd, "-family", "inet", "-json", "neighbor", "show", "nud", "reachable")
	b, err := cmd.Output()
	if err != nil {
		return
	}

	var es []arpEntry
	err = json.Unmarshal(b, &es)

	for _, e := range es {
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

	for hw, ok := range as {
		hws[hw].Set(ok)
		if ok {
			present = true
		}
	}

	return
}

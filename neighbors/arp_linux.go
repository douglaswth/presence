package neighbors

import (
	"context"
	"encoding/json"
	"net"
	"os/exec"

	"goa.design/clue/log"
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

	cmd := exec.CommandContext(ctx, a.cmd, "-family", "inet", "-json", "neighbor", "show", "nud", "reachable")
	if len(ifs) == 1 {
		for ifi := range ifs {
			cmd.Args = append(cmd.Args, "dev", ifi)
		}
	}
	log.Debug(ctx, log.KV{K: "cmd", V: cmd})
	b, err := cmd.Output()
	if err != nil {
		return
	}

	var es []arpEntry
	err = json.Unmarshal(b, &es)

	for _, e := range es {
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

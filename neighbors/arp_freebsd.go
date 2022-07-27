package neighbors

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"

	"goa.design/clue/log"
)

const (
	arpCmd           = "arp"
	arpOutputVersion = "1"
)

type (
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

func (a *arp) entries(ctx context.Context, ifs Interfaces) (entries []arpEntry, err error) {
	cmd := exec.CommandContext(ctx, a.cmd, "--libxo=json", "-an")
	if len(ifs) == 1 {
		for ifi := range ifs {
			cmd.Args = append(cmd.Args, "-i", ifi)
		}
	}
	log.Debug(ctx, log.KV{K: "cmd", V: cmd})
	b, err := cmd.Output()
	if err != nil {
		return
	}

	o := &arpOutput{}
	if err = json.Unmarshal(b, o); err != nil {
		return
	}

	if o.Version != arpOutputVersion {
		err = fmt.Errorf("arp output version mismatch (got %v, expected %v)", o.Version, arpOutputVersion)
		return
	}

	entries = make([]arpEntry, 0, len(o.ARP.Cache))
	for _, e := range o.ARP.Cache {
		if e.IPAddress != "" {
			entries = append(entries, e)
		}
	}

	return
}

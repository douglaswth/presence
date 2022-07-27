package neighbors

import (
	"context"
	"encoding/json"
	"os/exec"

	"goa.design/clue/log"
)

type (
	arpEntry struct {
		IPAddress  string `json:"dst"`
		MACAddress string `json:"lladdr"`
		Interface  string `json:"dev"`
	}
)

func (a *arp) entries(ctx context.Context, ifs Interfaces) (entries []arpEntry, err error) {
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

	if err = json.Unmarshal(b, &entries); err != nil {
		return
	}

	return
}

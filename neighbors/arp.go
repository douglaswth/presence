package neighbors

import (
	"context"

	"douglasthrift.net/presence"
)

type (
	Interfaces         map[string]bool
	HardwareAddrStates map[string]presence.State

	ARP interface {
		Present(ctx context.Context, ifs Interfaces, addrs HardwareAddrStates) (bool, error)
	}
)

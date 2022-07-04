package neighbors

import (
	"context"
)

type (
	Interfaces         map[string]bool
	HardwareAddrStates map[string]State

	ARP interface {
		Present(ctx context.Context, ifs Interfaces, state State, addrStates HardwareAddrStates) error
		Count(count uint)
	}
)

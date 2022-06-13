package neighbors

import (
	"context"
)

type (
	ARP interface {
		Present(ctx context.Context, ifs map[string]bool, addrs map[string]bool) (bool, error)
	}
)

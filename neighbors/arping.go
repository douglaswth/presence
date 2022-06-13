package neighbors

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
)

type (
	ARPing interface {
		Ping(ctx context.Context, ifi, hw, ip string) (bool, error)
	}

	arping struct {
		cmd, sudoCmd, count string
	}
)

func NewARPing(count uint) (ARPing, error) {
	cmd, err := exec.LookPath("arping")
	if err != nil {
		return nil, err
	}

	sudoCmd, err := exec.LookPath("sudo")
	if err != nil {
		return nil, err
	}

	return &arping{cmd: cmd, sudoCmd: sudoCmd, count: fmt.Sprint(count)}, nil
}

func (a *arping) Ping(ctx context.Context, ifi, hw, ip string) (ok bool, err error) {
	cmd := exec.CommandContext(ctx, a.sudoCmd, a.cmd, "-c", a.count, "-i", ifi, "-t", hw, "-q", ip)
	err = cmd.Run()
	if err == nil {
		ok = true
	} else {
		var exitError *exec.ExitError
		if errors.As(err, &exitError) && len(exitError.Stderr) == 0 {
			err = nil
		}
	}

	return
}

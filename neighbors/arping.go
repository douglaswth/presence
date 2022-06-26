package neighbors

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"

	"goa.design/clue/log"
)

type (
	ARPing interface {
		Ping(ctx context.Context, ifi, hw, ip string) (bool, error)
	}

	arping struct {
		arpingCmd, sudoCmd, count string
	}
)

func NewARPing(count uint) (ARPing, error) {
	arpingCmd, err := exec.LookPath("arping")
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(arpingCmd, "--help")
	b, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf(`incompatible "arping" command (%w)`, err)
	}

	ok, err := regexp.Match("^ARPing ", b)
	if err != nil {
		return nil, err
	}
	if !ok {
		r := bufio.NewReaderSize(bytes.NewReader(b), 32)
		l, p, err := r.ReadLine()
		if err != nil {
			return nil, fmt.Errorf(`incompatible "arping" command (%w)`, err)
		}

		var e string
		if p {
			e = "\u2026"
		}

		return nil, fmt.Errorf(`incompatible "arping" command (%s%v)`, l, e)
	}

	sudoCmd, err := exec.LookPath("sudo")
	if err != nil {
		return nil, err
	}

	return &arping{arpingCmd: arpingCmd, sudoCmd: sudoCmd, count: fmt.Sprint(count)}, nil
}

func (a *arping) Ping(ctx context.Context, ifi, hw, ip string) (ok bool, err error) {
	cmd := exec.CommandContext(ctx, a.sudoCmd, a.arpingCmd, "-c", a.count, "-i", ifi, "-t", hw, "-q", ip)
	log.Debug(ctx, log.KV{K: "cmd", V: cmd})
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

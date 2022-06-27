package main

import (
	"context"
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/alecthomas/kong"
	"goa.design/clue/log"
)

type (
	CLI struct {
		Config  string           `default:"${config}" help:"Set the configuration file." short:"c" type:"existingfile"`
		Debug   bool             `help:"Show debug information in log." short:"d"`
		Version kong.VersionFlag `help:"Show version information." short:"v"`

		Detect Detect `cmd:"" help:"Detect network presence and push state changes to IFTTT."`
		Check  Check  `cmd:"" help:"Check configuration."`
	}
)

var (
	version      = "dev"
	commit       = "none"
	date         = "unknown"
	configPrefix = ""
)

func main() {
	cli := &CLI{}
	ctx := kong.Parse(
		cli,
		kong.Description("Home network presence detection daemon for IFTTT"), kong.UsageOnError(),
		kong.Vars{
			"config":  filepath.Join(configPrefix, "presence.yml"),
			"version": fmt.Sprintf("presence version %v %v %v/%v %v %v", version, runtime.Version(), runtime.GOOS, runtime.GOARCH, commit, date),
		},
	)
	err := ctx.Run(cli)
	ctx.FatalIfErrorf(err)
}

func (cli *CLI) Context() (ctx context.Context) {
	ctx = context.Background()
	if cli.Debug {
		ctx = log.Context(ctx, log.WithDebug())
	} else {
		ctx = log.Context(ctx, log.WithDisableBuffering(func(context.Context) bool { return true }))
	}
	return
}

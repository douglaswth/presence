package main

import (
	"context"
	"log"
	"os"

	"douglasthrift.net/presence"
)

func main() {
	ifs := presence.Interfaces{os.Args[1]: true}
	hws := make(presence.HardwareAddrStates, len(os.Args[2:]))
	for _, hw := range os.Args[2:] {
		hws[hw] = presence.NewState()
	}

	ctx := context.Background()
	a, err := presence.NewARP(1)
	if err != nil {
		log.Fatal(err)
	}

	ok, err := a.Present(ctx, ifs, hws)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("present=%v", ok)
	for hw, state := range hws {
		log.Printf("%v present=%v changed=%v", hw, state.Present(), state.Changed())
	}
}

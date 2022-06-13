package main

import (
	"context"
	"log"
	"os"

	"douglasthrift.net/presence/neighbors"
)

func main() {
	ifs := map[string]bool{os.Args[1]: true}
	hws := make(map[string]bool, len(os.Args[2:]))
	for _, hw := range os.Args[2:] {
		hws[hw] = true
	}

	ctx := context.Background()
	a, err := neighbors.NewARP(1)
	if err != nil {
		log.Fatal(err)
	}

	ok, err := a.Present(ctx, ifs, hws)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(ok)
}

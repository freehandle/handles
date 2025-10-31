package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/freehandle/breeze/crypto"
	"github.com/freehandle/breeze/middleware/simple"
	"github.com/freehandle/handles/attorney"
)

func main() {
	genesis := attorney.NewGenesisState("")
	fmt.Println("Initialized genesis state")
	blocos := make(chan *simple.SimpleBlock, 0)
	lastReadEpoch := make(chan uint64)
	go func() {
		epoch := uint64(0)
		for {
			block, ok := <-blocos
			if !ok {
				break
			}
			validator := genesis.Validator()
			epoch = block.Epoch
			fmt.Println("Lendo bloco...", epoch)
			for _, action := range block.Actions {
				validator.Validate(action)
			}
			genesis.Incorporate(validator.Mutations())
		}
		lastReadEpoch <- epoch
	}()
	writer, err := simple.OpenSimpleBlockWriter("", "blocos", 100000000, blocos)
	if err != nil {
		log.Fatalf("could not open block writer: %v", err)
	}
	epoch := <-lastReadEpoch
	fmt.Println("terminou lendo blocos")
	simpleChain := simple.SimpleChain[*attorney.Mutations, *attorney.MutatingState]{
		Interval:    time.Second,
		GatewayPort: 7000,
		Writer:      writer,
		State:       genesis,
		Epoch:       epoch,
		Recent:      [][][]byte{},
		Keep:        10,
	}
	token, pk := crypto.RandomAsymetricKey()
	fmt.Printf("Starting simple chain gateway on port 7000: token %s\n", token.String())
	err = <-simpleChain.Start(context.Background(), pk, attorney.ToString)
	if err != nil {
		log.Fatalf("could not start simple chain: %v", err)
	}
}

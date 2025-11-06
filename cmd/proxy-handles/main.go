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
	blocos := make(chan *simple.SimpleBlock, 1)
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
			for _, action := range block.Actions {
				if !validator.Validate(action) {
					fmt.Printf("invalid action in block: %s\n-----\n", attorney.ToString(action))
				}
				fmt.Println(attorney.ToString(action))
				fmt.Println("---")
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

	pk := crypto.PrivateKeyFromString("b61b1452f41a62ac20a1cf5136a2c692d7312f70c65f196426f0a11aca733d3d91ad274d06c4be307a332a0e59449ad25ae2c65e4ad5a8f0af87067ac2fc3a54")
	token := pk.PublicKey()
	fmt.Printf("Starting simple chain gateway on port 7000: token %s\n", token.String())
	err = <-simpleChain.Start(context.Background(), pk, attorney.ToString)
	if err != nil {
		log.Fatalf("could not start simple chain: %v", err)
	}
}

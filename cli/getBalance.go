package cli

import (
	"blockchain1/blockchain"
	"blockchain1/lib/base58"
	wal "blockchain1/wallet"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

func (cli *CLI) getBalance(address string) {
	if !wal.ValidateAddress(address) {
		log.Fatal("ERROR: Address is not valid")
	}

	bc := blockchain.NewBlockchain(address)
	defer func(Db *bolt.DB) {
		err := Db.Close()
		if err != nil {
			log.Panic(err)
		}
	}(bc.Db)

	balance := 0
	pubKeyHash := base58.Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := bc.FindUTXO(pubKeyHash)
	for _, out := range UTXOs {
		balance += out.Value

	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

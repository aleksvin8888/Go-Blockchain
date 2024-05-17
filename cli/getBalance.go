package cli

import (
	"blockchain1/blockchain"
	"blockchain1/lib/base58"
	wal "blockchain1/wallet"
	"fmt"
	"log"
)

func (cli *CLI) getBalance(address string, nodeID string) {
	if !wal.ValidateAddress(address) {
		log.Fatal("ERROR: Address is not valid")
	}

	bc := blockchain.NewBlockchain(nodeID)
	defer func() { _ = bc.Db.Close() }()

	UTXOSet := blockchain.UTXOSet{
		Blockchain: bc,
	}

	balance := 0
	pubKeyHash := base58.Decode([]byte(address))
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	UTXOs := UTXOSet.FindUTXO(pubKeyHash)
	for _, out := range UTXOs {
		balance += out.Value
	}

	fmt.Printf("Balance of '%s': %d\n", address, balance)
}

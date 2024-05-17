package cli

import (
	"blockchain1/blockchain"
	wal "blockchain1/wallet"
	"fmt"
	"log"
)

func (cli *CLI) createBlockchain(address string, nodeID string) {
	if !wal.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}

	bc := blockchain.CreateBlockchain(address, nodeID)
	defer func() { _ = bc.Db.Close() }()

	UTXOSet := blockchain.UTXOSet{
		Blockchain: bc,
	}
	UTXOSet.Reindex()
	fmt.Println("Done!")
}

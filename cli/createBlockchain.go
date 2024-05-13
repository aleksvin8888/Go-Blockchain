package cli

import (
	"blockchain1/blockchain"
	wal "blockchain1/wallet"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

func (cli *CLI) createBlockchain(address string) {
	if !wal.ValidateAddress(address) {
		log.Panic("ERROR: Address is not valid")
	}
	bc := blockchain.CreateBlockchain(address)
	defer func(Db *bolt.DB) {
		err := Db.Close()
		if err != nil {
			log.Panic(err)
		}
	}(bc.Db)
	UTXOSet := blockchain.UTXOSet{
		Blockchain: bc,
	}
	UTXOSet.Reindex()
	fmt.Println("Done!")
}

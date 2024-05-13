package cli

import (
	"blockchain1/blockchain"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

func (cli *CLI) reindexUTXO() {
	bc := blockchain.NewBlockchain()
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
	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

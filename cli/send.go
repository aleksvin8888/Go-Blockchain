package cli

import (
	"blockchain1/blockchain"
	"blockchain1/transaction"
	wal "blockchain1/wallet"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
)

func (cli *CLI) send(from string, to string, amount int) {

	if !wal.ValidateAddress(from) {
		log.Fatal("ERROR: address from is not valid")
	}

	if !wal.ValidateAddress(to) {
		log.Fatal("ERROR: address to is not valid")
	}

	bc := blockchain.NewBlockchain(from)
	defer func(Db *bolt.DB) {
		err := Db.Close()
		if err != nil {
			log.Panic(err)
		}
	}(bc.Db)

	tx := blockchain.NewUTXOTransaction(from, to, amount, bc)
	bc.MineBlock([]*transaction.Transaction{tx})
	fmt.Println("Success!")
}

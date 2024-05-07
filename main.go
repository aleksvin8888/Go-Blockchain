package main

import (
	"blockchain1/blockchain"
	cliApp "blockchain1/cli"
	"fmt"
	"github.com/boltdb/bolt"
)

func main() {
	fmt.Print("Hello, Blockchain!\n\n")

	bc := blockchain.NewBlockchain()
	defer func(db *bolt.DB) {
		err := db.Close()
		if err != nil {
			fmt.Print("Error closing db\n")
		}
	}(bc.Db)

	cli := cliApp.CLI{BC: bc}
	cli.Run()
}

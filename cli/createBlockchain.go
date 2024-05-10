package cli

import (
	"blockchain1/blockchain"
	"fmt"
	"log"
)

func (cli *CLI) createBlockchain(address string) {
	bc := blockchain.CreateBlockchain(address)
	err := bc.Db.Close()
	if err != nil {
		log.Panic(err)
	}
	fmt.Println("Done!")
}

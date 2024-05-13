package cli

import (
	"blockchain1/blockchain"
	"blockchain1/bloks"
	"github.com/boltdb/bolt"
	"log"
	"strconv"
)

func (cli *CLI) printChain() {

	bc := blockchain.NewBlockchain()
	defer func(Db *bolt.DB) {
		err := Db.Close()
		if err != nil {
			log.Panic(err)
		}
	}(bc.Db)

	bci := bc.Iterator()

	for {
		block := bci.Next()

		log.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		log.Printf("Hash: %x\n", block.Hash)
		pow := bloks.NewProofOfWork(block)
		log.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			log.Printf("Transaction ID: %x\n", tx.ID)
		}
		log.Println("------------------------------------------------------------------")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

package cli

import (
	"blockchain1/blockchain"
	"blockchain1/bloks"
	"fmt"
	"strconv"
)

func (cli *CLI) printChain(nodeID string) {

	bc := blockchain.NewBlockchain(nodeID)
	defer func() { _ = bc.Db.Close() }()

	bci := bc.Iterator()

	for {
		block := bci.Next()

		fmt.Printf("============ Block %x ============\n", block.Hash)
		fmt.Printf("Height: %d\n", block.Height)
		fmt.Printf("Prev. block: %x\n", block.PrevBlockHash)
		pow := bloks.NewProofOfWork(block)
		fmt.Printf("PoW: %s\n\n", strconv.FormatBool(pow.Validate()))
		for _, tx := range block.Transactions {
			fmt.Printf("Transaction ID: %xn", tx.ID)
			fmt.Println("VIn:")
			for _, vin := range tx.VIn {
				fmt.Printf("tTxID: %x\n", vin.TxId)
				fmt.Printf("tVout: %d\n", vin.VOut)
				fmt.Printf("tSignature: %x\n", vin.Signature)
				fmt.Printf("tPubKey: %x\n", vin.PubKey)
			}
			fmt.Println("VOut:")
			for _, vout := range tx.VOut {
				fmt.Printf("tValue: %d\n", vout.Value)
				fmt.Printf("tScriptPubKey: %x\n", vout.PubKeyHash)
			}
			fmt.Println()
		}
		fmt.Printf("\n\n")

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

package cli

import (
	"blockchain1/blockchain"
	"fmt"
)

func (cli *CLI) reindexUTXO(nodeID string) {
	bc := blockchain.NewBlockchain(nodeID)
	defer func() { _ = bc.Db.Close() }()

	UTXOSet := blockchain.UTXOSet{
		Blockchain: bc,
	}
	UTXOSet.Reindex()
	count := UTXOSet.CountTransactions()
	fmt.Printf("Done! There are %d transactions in the UTXO set.\n", count)
}

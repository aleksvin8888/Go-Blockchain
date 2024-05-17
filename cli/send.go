package cli

import (
	"blockchain1/blockchain"
	"blockchain1/server"
	"blockchain1/transaction"
	wal "blockchain1/wallet"
	ws "blockchain1/wallets"
	"fmt"
	"log"
)

func (cli *CLI) send(from string, to string, amount int, nodeID string, mineNow bool) {

	if !wal.ValidateAddress(from) {
		log.Fatal("ERROR: address from is not valid")
	}
	if !wal.ValidateAddress(to) {
		log.Fatal("ERROR: address to is not valid")
	}

	bc := blockchain.NewBlockchain(nodeID)
	UTXOSet := blockchain.UTXOSet{
		Blockchain: bc,
	}
	defer func() { _ = bc.Db.Close() }()

	wallets, err := ws.NewWallets(nodeID)
	if err != nil {
		log.Panic(err)
	}
	wallet := wallets.GetWallet(from)

	tx := blockchain.NewUTXOTransaction(&wallet, to, amount, &UTXOSet)

	if mineNow {
		cbTx := transaction.NewCoinbaseTX(from, "")
		txs := []*transaction.Transaction{cbTx, tx}

		newBlock := bc.MineBlock(txs)
		UTXOSet.Update(newBlock)
	} else {
		server.SendTx(server.KnownNodes[0], tx)
	}
	fmt.Println("Success!")
}

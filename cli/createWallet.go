package cli

import (
	ws "blockchain1/wallets"
	"fmt"
)

func (cli *CLI) createWallet(nodeID string) {
	wallets, _ := ws.NewWallets(nodeID)
	address := wallets.CreateWallet()
	wallets.SaveToFile(nodeID)

	fmt.Printf("Your new address: %s\n", address)
}

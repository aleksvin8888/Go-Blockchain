package cli

import (
	ws "blockchain1/wallets"
	"fmt"
)

func (cli *CLI) createWallet() {
	wallets, _ := ws.NewWallets()
	address := wallets.CreateWallet()
	wallets.SaveToFile()

	fmt.Printf("Your new address: %s\n", address)
}

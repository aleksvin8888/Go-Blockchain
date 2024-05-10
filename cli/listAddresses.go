package cli

import (
	ws "blockchain1/wallets"
	"fmt"
	"log"
)

func (cli *CLI) listAddresses() {
	wallets, err := ws.NewWallets()
	if err != nil {
		log.Panic(err)
	}
	addresses := wallets.GetAddresses()

	for _, address := range addresses {
		fmt.Println(address)
	}
}

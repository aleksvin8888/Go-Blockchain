package main

import (
	cliBlockchain "blockchain1/cli"
	"fmt"
)

func main() {
	fmt.Print("Hello, Blockchain!\n\n")

	cli := cliBlockchain.CLI{}
	cli.Run()
}

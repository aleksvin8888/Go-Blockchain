package main

import (
	cliBlockchain "blockchain1/cli"
	"fmt"
	"os"
)

func main() {

	nodeID := os.Getenv("NODE_ID")
	fmt.Printf("Testing Blockchain\n")
	fmt.Printf("Usage NODE_ID: %s\n\n", nodeID)

	cli := cliBlockchain.CLI{}
	cli.Run()
}

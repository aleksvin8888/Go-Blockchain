package cli

import (
	"blockchain1/blockchain"
	"blockchain1/bloks"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"
)

type CLI struct {
	BC *blockchain.Blockchain
}

func (cli *CLI) Run() {
	cli.validateArgs()

	addBlockCmd := flag.NewFlagSet("addblock", flag.ExitOnError)
	printChainCmd := flag.NewFlagSet("printchain", flag.ExitOnError)

	addBlockData := addBlockCmd.String("data", "", "Block data")

	switch os.Args[1] {
	case "addblock":
		err := addBlockCmd.Parse(os.Args[2:])
		if err != nil {
			log.Print("Error parsing addblock command", err)
		}
	case "printchain":
		err := printChainCmd.Parse(os.Args[2:])
		if err != nil {
			log.Print("Error parsing printchain command", err)
		}
	default:
		cli.printUsage()
		os.Exit(1)
	}

	if addBlockCmd.Parsed() {
		if *addBlockData == "" {
			addBlockCmd.Usage()
			os.Exit(1)
		}
		cli.addBlock(*addBlockData)
	}

	if printChainCmd.Parsed() {
		cli.printChain()
	}

}

func (cli *CLI) addBlock(data string) {
	cli.BC.AddBlock(data)
	log.Print("Block added!")
}

func (cli *CLI) printChain() {
	bci := cli.BC.Iterator()

	for {
		block := bci.Next()

		log.Printf("Prev. hash: %x\n", block.PrevBlockHash)
		log.Printf("Data: %s\n", block.Data)
		log.Printf("Hash: %x\n", block.Hash)
		pow := bloks.NewProofOfWork(block)
		log.Printf("PoW: %s\n", strconv.FormatBool(pow.Validate()))
		log.Println()

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}
}

func (cli *CLI) printUsage() {
	fmt.Println("Usage:")
	fmt.Println("  addblock -data BLOCK_DATA - add a block to the blockchain")
	fmt.Println("  printchain - print all the blocks of the blockchain")
}

func (cli *CLI) validateArgs() {
	if len(os.Args) < 2 {
		cli.printUsage()
		os.Exit(1)
	}
}

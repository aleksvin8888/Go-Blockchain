package blockchain

import (
	"blockchain1/bloks"
)

type Blockchain struct {
	Blocks []*bloks.Block
}

func (bc *Blockchain) AddBlock(data string) {
	prevBlock := bc.Blocks[len(bc.Blocks)-1]
	newBlock := bloks.NewBlock(data, prevBlock.Hash)
	bc.Blocks = append(bc.Blocks, newBlock)
}
func NewGenesisBlock() *bloks.Block {
	return bloks.NewBlock("Genesis Block", []byte{})
}

func NewBlockchain() *Blockchain {
	return &Blockchain{[]*bloks.Block{NewGenesisBlock()}}
}

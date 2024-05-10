package blockchain

import (
	"blockchain1/bloks"
	"github.com/boltdb/bolt"
	"log"
)

type Iterator struct {
	currentHash []byte
	db          *bolt.DB
}

/*
Next повертає наступний блок у Blockchain.
*/
func (i *Iterator) Next() *bloks.Block {
	var block *bloks.Block

	err := i.db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		encodedBlock := b.Get(i.currentHash)
		block = bloks.DeserializeBlock(encodedBlock)

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	i.currentHash = block.PrevBlockHash

	return block
}

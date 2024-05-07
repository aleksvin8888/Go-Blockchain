package bloks

import (
	"bytes"
	"crypto/sha256"
	"strconv"
)

/*
Block : структура, що представляє блок у блокчейні.
- Timestamp - час створення блока.
- Data - дані, що містяться в блоку.
- PrevBlockHash - зберігає хеш попереднього блоку.
- Hash - містить хеш поточного блоку.
*/
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
}

func (b *Block) SetHash() {
	timestamp := []byte(strconv.FormatInt(b.Timestamp, 10))
	headers := bytes.Join([][]byte{
		b.PrevBlockHash,
		b.Data,
		timestamp,
	}, []byte{})
	hash := sha256.Sum256(headers)

	b.Hash = hash[:]
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     0,
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
	}
	block.SetHash()
	return block
}

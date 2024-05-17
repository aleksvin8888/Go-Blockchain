package bloks

import (
	"blockchain1/merkleTree"
	"blockchain1/transaction"
	"bytes"
	"encoding/gob"
	"log"
	"time"
)

/*
Block : структура, що представляє блок у блокчейні.
- Timestamp - час створення блока.
- Transactions - масив транзакцій фактично представляє собою дані, які будуть зберігатися в блоку.
- PrevBlockHash - зберігає хеш попереднього блоку.
- Hash - містить хеш поточного блоку.
- Nonce - використовується для доказу роботи.
*/
type Block struct {
	Timestamp     int64
	Transactions  []*transaction.Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
	Height        int
}

/*
NewGenesisBlock : використовується для створення першого genesis блоку.
*/
func NewGenesisBlock(coinbase *transaction.Transaction) *Block {
	return NewBlock(
		[]*transaction.Transaction{coinbase},
		[]byte{},
		0,
	)
}

/*
NewBlock : використовується для створення нового блоку.
*/
func NewBlock(transaction []*transaction.Transaction, prevBlockHash []byte, height int) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transaction,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
		Nonce:         0,
		Height:        height,
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

/*
Serialize : використовується для серіалізації блоку.
*/
func (b *Block) Serialize() []byte {
	var result bytes.Buffer

	encoder := gob.NewEncoder(&result)

	err := encoder.Encode(b)
	if err != nil {
		log.Fatal(" Can't serialize block ", err)
	}

	return result.Bytes()
}

/*
DeserializeBlock : використовується для десеріалізації блоку.
*/
func DeserializeBlock(d []byte) *Block {
	var block Block

	decoder := gob.NewDecoder(bytes.NewReader(d))

	err := decoder.Decode(&block)
	if err != nil {
		log.Fatal(" Can't deserialize block ", err)
	}

	return &block
}

/*
HashTransactions : використовується для обчислення хешу транзакцій у блоку.
*/
func (b *Block) HashTransactions() []byte {
	var transactions [][]byte

	for _, tx := range b.Transactions {
		transactions = append(transactions, tx.Serialize())
	}

	mTree := merkleTree.NewMerkleTree(transactions)

	return mTree.RootNode.Data
}

package bloks

import (
	"blockchain1/transactions"
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"log"
	"time"
)

/*
Block : структура, що представляє блок у блокчейні.
- Timestamp - час створення блока.
- PrevBlockHash - зберігає хеш попереднього блоку.
- Hash - містить хеш поточного блоку.

Замінюємо Data на Transactions, щоб зберігати транзакції в блоках.
*/
type Block struct {
	Timestamp     int64
	Transactions  []*transactions.Transaction
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

func NewGenesisBlock(coinbase *transactions.Transaction) *Block {
	return NewBlock(
		[]*transactions.Transaction{coinbase},
		[]byte{},
	)
}

func NewBlock(transaction []*transactions.Transaction, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Transactions:  transaction,
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

/*
Serialize : Цей метод використовується для серіалізації блоку.
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
DeserializeBlock : Цей метод використовується для десеріалізації блоку.
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
HashTransactions : Цей метод використовується для обчислення хешу транзакцій у блоку.
*/
func (b *Block) HashTransactions() []byte {
	var txHashes [][]byte
	var txHash [32]byte

	for _, tx := range b.Transactions {
		txHashes = append(txHashes, tx.ID)
	}
	txHash = sha256.Sum256(bytes.Join(txHashes, []byte{}))

	return txHash[:]
}

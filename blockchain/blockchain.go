package blockchain

import (
	"blockchain1/bloks"
	"blockchain1/transactions"
	"encoding/hex"
	"fmt"
	"github.com/boltdb/bolt"
	"log"
	"os"
)

const (
	dbFile              = "blockchain.db"
	blocksBucket        = "blocks"
	genesisCoinbaseData = "The Times 03/Jan/2009 Chancellor on brink of second bailout for banks"
)

type Blockchain struct {
	tip []byte
	Db  *bolt.DB
}

type Iterator struct {
	currentHash []byte
	db          *bolt.DB
}

// CreateBlockchain створює новий ланцюг блоків Blockchain з блоком генезиса.
func CreateBlockchain(address string) *Blockchain {
	if dbExists() {
		fmt.Println("Blockchain уже існує.")
		os.Exit(1)
	}
	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		cbTx := transactions.NewCoinbaseTX(address, genesisCoinbaseData)
		genesis := bloks.NewGenesisBlock(cbTx)

		b, err := tx.CreateBucket([]byte(blocksBucket))
		if err != nil {
			log.Panic(err)
		}

		err = b.Put(genesis.Hash, genesis.Serialize())
		if err != nil {
			log.Panic(err)
		}

		err = b.Put([]byte("l"), genesis.Hash)
		if err != nil {
			log.Panic(err)
		}
		tip = genesis.Hash

		return nil
	})

	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{tip, db}

	return &bc
}

/*
NewBlockchain TODO
*/
func NewBlockchain(address string) *Blockchain {
	if dbExists() == false {
		fmt.Println("Blockchain ще не створено. Спочатку створіть його.")
		os.Exit(1)
	}

	var tip []byte
	db, err := bolt.Open(dbFile, 0600, nil)
	if err != nil {
		log.Panic(err)
	}

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		tip = b.Get([]byte("l"))

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	bc := Blockchain{
		tip,
		db,
	}

	return &bc
}

/*
MineBlock додає новий блок до ланцюга Blockchain.
*/
func (bc *Blockchain) MineBlock(transactions []*transactions.Transaction) {
	var lastHash []byte

	err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	newBlock := bloks.NewBlock(transactions, lastHash)

	err = bc.Db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		err := b.Put(newBlock.Hash, newBlock.Serialize())
		if err != nil {
			log.Panic(err)
		}
		err = b.Put([]byte("l"), newBlock.Hash)
		if err != nil {
			log.Panic(err)
		}
		bc.tip = newBlock.Hash
		return nil
	})
}

/*
Iterator використовується для ітерації по блоках у Blockchain.
*/
func (bc *Blockchain) Iterator() *Iterator {
	bci := &Iterator{
		bc.tip,
		bc.Db,
	}

	return bci
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

/*
dbExists перевіряє, чи існує файл бази даних.
*/
func dbExists() bool {
	if _, err := os.Stat(dbFile); os.IsNotExist(err) {
		return false
	}

	return true
}

/*
FindUnspentTransactions находит и возвращает список всех непотраченных транзакций.
*/
func (bc *Blockchain) FindUnspentTransactions(address string) []transactions.Transaction {

	var unspentTXs []transactions.Transaction
	spentTXOs := make(map[string][]int)

	bci := bc.Iterator()
	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := string(tx.ID)

		Outputs:
			for outIdx, out := range tx.VOut {
				/*Если выход был заблокирован по тому же адресу, мы ищем непотраченные выходы, которые мы хотим.
				Но перед тем, как принять его, нам нужно проверить, был ли на выходе уже указан вход:*/
				if spentTXOs[txID] != nil {
					for _, spentOut := range spentTXOs[txID] {
						if spentOut == outIdx {
							continue Outputs
						}
					}
				}
				/*Поскольку транзакции хранятся в блоках, мы должны проверять каждый блок в цепочке. Начнем с выходов:*/
				if out.CanBeUnlockedWith(address) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.VIn {
					if in.CanUnlockOutputWith(address) {
						inTxID := string(in.TxId)
						spentTXOs[inTxID] = append(spentTXOs[inTxID], in.VOut)
					}
				}
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return unspentTXs
}

/*
FindUTXO находит и возвращает список всех непотраченных выходов транзакций.
*/
func (bc *Blockchain) FindUTXO(address string) []transactions.TXOutput {
	var UTXOs []transactions.TXOutput
	unspentTransactions := bc.FindUnspentTransactions(address)

	for _, tx := range unspentTransactions {
		for _, out := range tx.VOut {
			if out.CanBeUnlockedWith(address) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

func (bc *Blockchain) FindSpendableOutputs(address string, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(address)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.VOut {
			if out.CanBeUnlockedWith(address) && accumulated < amount {
				accumulated += out.Value
				unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)

				if accumulated >= amount {
					break Work
				}
			}
		}
	}

	return accumulated, unspentOutputs
}

func NewUTXOTransaction(from string, to string, amount int, bc *Blockchain) *transactions.Transaction {
	var inputs []transactions.TXInput
	var outputs []transactions.TXOutput

	acc, validOutputs := bc.FindSpendableOutputs(from, amount)

	if acc < amount {
		log.Print("Error: Not enough funds")
		os.Exit(0)
	}

	// Build a list of inputs
	for txId, outs := range validOutputs {
		txID, err := hex.DecodeString(txId)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := transactions.TXInput{
				TxId:      txID,
				VOut:      out,
				ScriptSig: from,
			}
			inputs = append(inputs, input)
		}
	}

	// Build a list of outputs
	outputs = append(outputs, transactions.TXOutput{
		Value:        amount,
		ScriptPubKey: to,
	})
	if acc > amount {
		outputs = append(outputs, transactions.TXOutput{
			Value:        acc - amount,
			ScriptPubKey: from,
		})
	}

	tx := transactions.Transaction{
		VIn:  inputs,
		VOut: outputs,
	}
	tx.SetID()

	return &tx
}

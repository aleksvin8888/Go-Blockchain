package blockchain

import (
	"blockchain1/bloks"
	"blockchain1/lib/utils"
	"blockchain1/transaction"
	wal "blockchain1/wallet"
	ws "blockchain1/wallets"
	"bytes"
	"crypto/ecdsa"
	"encoding/hex"
	"errors"
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

// Blockchain структура,
// tip - зберігає хеш останнього блоку в ланцюгу.
// Db - вказівник на базу даних.
type Blockchain struct {
	tip []byte
	Db  *bolt.DB
}

// CreateBlockchain створює нову базу даних Blockchain з genesis блоком.
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
		cbTx := transaction.NewCoinbaseTX(address, genesisCoinbaseData)
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
NewBlockchain створює новий Blockchain з genesis блоком.
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
NewUTXOTransaction створює нову транзакцію UTXO,
фактично відправляємо монети з одного гаманця на інший.
*/
func NewUTXOTransaction(from string, to string, amount int, bc *Blockchain) *transaction.Transaction {
	var inputs []transaction.TXInput
	var outputs []transaction.TXOutput

	wallets, err := ws.NewWallets()
	if err != nil {
		log.Panic(err)
	}

	wallet := wallets.GetWallet(from)
	pubKeyHash := wal.HashPubKey(wallet.PublicKey)
	acc, validOutputs := bc.FindSpendableOutputs(pubKeyHash, amount)
	if acc < amount {
		log.Print("Error: Недостатньо коштів")
		os.Exit(0)
	}

	//складаємо список входів
	for txId, outs := range validOutputs {
		txID, err := hex.DecodeString(txId)
		if err != nil {
			log.Panic(err)
		}

		for _, out := range outs {
			input := transaction.TXInput{
				TxId:      txID,
				VOut:      out,
				Signature: nil,
				PubKey:    wallet.PublicKey,
			}
			inputs = append(inputs, input)
		}
	}

	// складаємо список виходів
	outputs = append(outputs, *transaction.NewTXOutput(amount, to))
	if acc > amount {
		outputs = append(outputs, *transaction.NewTXOutput(acc-amount, from))
	}

	tx := transaction.Transaction{
		ID:   nil,
		VIn:  inputs,
		VOut: outputs,
	}
	tx.ID = tx.Hash()

	privateKey, err := utils.PrivateKeyFromBytes(wallet.PrivateKey)
	if err != nil {
		log.Panic(err)
	}

	bc.SignTransaction(&tx, *privateKey) // передаємо створену транзакцію у процес підпису

	return &tx
}

/*
MineBlock додає новий блок до ланцюга Blockchain.
*/
func (bc *Blockchain) MineBlock(transactions []*transaction.Transaction) {
	var lastHash []byte

	// перевірка чи транзакції валідні перед додаванням нового блоку
	for _, tx := range transactions {
		if bc.VerifyTransaction(tx) != true {
			log.Fatal("Помилка: транзакція не є дійсною")
		}
	}

	// отримання хеша останнього блоку з бази даних Blockchain
	err := bc.Db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(blocksBucket))
		lastHash = b.Get([]byte("l"))
		return nil
	})
	if err != nil {
		log.Panic(err)
	}
	// створення нового блоку
	newBlock := bloks.NewBlock(transactions, lastHash)

	// оновлення бази даних Blockchain з новим блоком
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
Iterator використовується для ітерації по блоках у базі Blockchain.
*/
func (bc *Blockchain) Iterator() *Iterator {
	bci := &Iterator{
		bc.tip,
		bc.Db,
	}

	return bci
}

/*
FindUnspentTransactions знаходить та повертає список всіх непотрачених транзакцій.
*/
func (bc *Blockchain) FindUnspentTransactions(pubKeyHash []byte) []transaction.Transaction {
	var unspentTXs []transaction.Transaction
	spentTXOs := make(map[string][]int)

	bci := bc.Iterator()
	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			txID := hex.EncodeToString(tx.ID)

		Outputs:
			for outIdx, out := range tx.VOut {
				//перевірка чи вже витрачений вихід
				if spentTXOs[txID] != nil {
					for _, spentOutIdx := range spentTXOs[txID] {
						if spentOutIdx == outIdx {
							continue Outputs
						}
					}
				}

				if out.IsLockedWithKey(pubKeyHash) {
					unspentTXs = append(unspentTXs, *tx)
				}
			}

			if tx.IsCoinbase() == false {
				for _, in := range tx.VIn {
					if in.UsesKey(pubKeyHash) {
						inTxID := hex.EncodeToString(in.TxId)
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
FindUTXO знаходить та повертає всі непотрачені виходи транзакцій, які можуть бути розблоковані за допомогою публічного ключа.
*/
func (bc *Blockchain) FindUTXO(pubKeyHash []byte) []transaction.TXOutput {
	var UTXOs []transaction.TXOutput
	unspentTransactions := bc.FindUnspentTransactions(pubKeyHash)

	for _, tx := range unspentTransactions {
		for _, out := range tx.VOut {
			if out.IsLockedWithKey(pubKeyHash) {
				UTXOs = append(UTXOs, out)
			}
		}
	}

	return UTXOs
}

/*
FindSpendableOutputs знаходить та повертає доступну кількість виходів, які можуть бути витрачені за допомогою публічного ключа,
*/
func (bc *Blockchain) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	unspentTXs := bc.FindUnspentTransactions(pubKeyHash)
	accumulated := 0

Work:
	for _, tx := range unspentTXs {
		txID := hex.EncodeToString(tx.ID)
		for outIdx, out := range tx.VOut {
			if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
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

/*
SignTransaction отримує одну транзакцію потім знаходить транзакції на які вона посилається
і передає у логіку підписування транзакції
*/
func (bc *Blockchain) SignTransaction(tx *transaction.Transaction, privetKey ecdsa.PrivateKey) {
	prevTXs := make(map[string]transaction.Transaction)

	for _, vin := range tx.VIn {
		prevTX, err := bc.FindTransaction(vin.TxId)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	tx.Sing(privetKey, prevTXs)
}

/*
FindTransaction  Пошук транзакції по ID
для цього потрібна ітерація по всіх блоках
*/
func (bc *Blockchain) FindTransaction(ID []byte) (transaction.Transaction, error) {
	bci := bc.Iterator()

	for {
		block := bci.Next()

		for _, tx := range block.Transactions {
			if bytes.Compare(tx.ID, ID) == 0 {
				return *tx, nil
			}
		}

		if len(block.PrevBlockHash) == 0 {
			break
		}
	}

	return transaction.Transaction{}, errors.New("транзакція не знайдена")
}

/*
VerifyTransaction перевіряє чи транзакція є дійсною
*/
func (bc *Blockchain) VerifyTransaction(tx *transaction.Transaction) bool {

	if tx.IsCoinbase() {
		return true
	}

	prevTXs := make(map[string]transaction.Transaction)

	for _, vin := range tx.VIn {
		prevTX, err := bc.FindTransaction(vin.TxId)
		if err != nil {
			log.Panic(err)
		}
		prevTXs[hex.EncodeToString(prevTX.ID)] = prevTX
	}

	return tx.Verify(prevTXs)
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

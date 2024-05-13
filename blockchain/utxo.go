package blockchain

import (
	"blockchain1/bloks"
	"blockchain1/transaction"
	"encoding/hex"
	"errors"
	"github.com/boltdb/bolt"
	"log"
)

const utxoBucket = "chainstate"

// UTXOSet структура, яка представляє набір непотрачених виходів транзакцій ( UTXO ).
type UTXOSet struct {
	Blockchain *Blockchain
}

/*
Reindex перебудовує UTXOset.
*/
func (u UTXOSet) Reindex() {
	db := u.Blockchain.Db
	bucketName := []byte(utxoBucket)

	err := db.Update(func(tx *bolt.Tx) error {
		err := tx.DeleteBucket(bucketName)
		if err != nil && !errors.Is(err, bolt.ErrBucketNotFound) {
			log.Panic(err)
		}

		_, err = tx.CreateBucket(bucketName)
		if err != nil {
			log.Panic(err)
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	UTXO := u.Blockchain.FindUTXO()

	err = db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket(bucketName)

		for txID, outs := range UTXO {
			key, err := hex.DecodeString(txID)
			if err != nil {
				log.Panic(err)
			}

			err = b.Put(key, outs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})

}

/*
FindSpendableOutputs знаходить та повертає доступну кількість виходів, які можуть бути витрачені за допомогою публічного ключа,
використовується для створення нової транзакції.
*/
func (u UTXOSet) FindSpendableOutputs(pubKeyHash []byte, amount int) (int, map[string][]int) {
	unspentOutputs := make(map[string][]int)
	accumulated := 0
	db := u.Blockchain.Db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			txID := hex.EncodeToString(k)
			outs := transaction.DeserializeOutputs(v)

			for outIdx, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) && accumulated < amount {
					accumulated += out.Value
					unspentOutputs[txID] = append(unspentOutputs[txID], outIdx)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return accumulated, unspentOutputs
}

/*
FindUTXO знаходить та повертає всі непотрачені виходи транзакцій, які можуть бути розблоковані за допомогою публічного ключа.
використовується для підрахунку балансу гаманця.
*/
func (u UTXOSet) FindUTXO(pubKeyHash []byte) []transaction.TXOutput {
	var UTXOs []transaction.TXOutput
	db := u.Blockchain.Db

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, v := c.First(); k != nil; k, v = c.Next() {
			outs := transaction.DeserializeOutputs(v)

			for _, out := range outs.Outputs {
				if out.IsLockedWithKey(pubKeyHash) {
					UTXOs = append(UTXOs, out)
				}
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return UTXOs
}

/*
Update оновлює набір UTXOset транзакціями з нового блоку.
*/
func (u UTXOSet) Update(block *bloks.Block) {
	db := u.Blockchain.Db

	err := db.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))

		for _, tx := range block.Transactions {
			if tx.IsCoinbase() == false {
				for _, vin := range tx.VIn {
					updatedOuts := transaction.TXOutputs{}
					outsBytes := b.Get(vin.TxId)
					outs := transaction.DeserializeOutputs(outsBytes)

					for outIdx, out := range outs.Outputs {
						if outIdx != vin.VOut {
							updatedOuts.Outputs = append(updatedOuts.Outputs, out)
						}
					}

					if len(updatedOuts.Outputs) == 0 {
						err := b.Delete(vin.TxId)
						if err != nil {
							log.Panic(err)
						}
					} else {
						err := b.Put(vin.TxId, updatedOuts.Serialize())
						if err != nil {
							log.Panic(err)
						}
					}

				}
			}

			newOutputs := transaction.TXOutputs{}
			for _, out := range tx.VOut {
				newOutputs.Outputs = append(newOutputs.Outputs, out)
			}

			err := b.Put(tx.ID, newOutputs.Serialize())
			if err != nil {
				log.Panic(err)
			}
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}
}

/*
CountTransactions підраховує кількість транзакцій у наборі UTXOset.
*/
func (u UTXOSet) CountTransactions() int {
	db := u.Blockchain.Db
	counter := 0

	err := db.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(utxoBucket))
		c := b.Cursor()

		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			counter++
		}

		return nil
	})
	if err != nil {
		log.Panic(err)
	}

	return counter
}

package transaction

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"math/big"
)

/*
subsidy - кількість монет, яку майнер отримує за добування genesis блоку
*/
const (
	subsidy = 100
)

/*
Transaction структура, що представляє транзакцію в блокчейні.
- ID - ідентифікатор транзакції.
- VIn - вхідні дані транзакції, які вказують на виходи попередніх транзакцій.
- VOut - вихідні дані транзакції, які вказують на суми та адреси отримувачів.
*/
type Transaction struct {
	ID   []byte
	VIn  []TXInput
	VOut []TXOutput
}

/*
NewCoinbaseTX створює нову транзакцію Coinbase.
*/
func NewCoinbaseTX(to string, data string) *Transaction {
	if data == "" {
		randData := make([]byte, 20)
		_, err := rand.Read(randData)
		if err != nil {
			log.Panic(err)
		}

		data = fmt.Sprintf("%x", randData)
	}

	txIn := TXInput{
		TxId:      []byte{},
		VOut:      -1,
		Signature: nil,
		PubKey:    []byte(data),
	}
	txOut := NewTXOutput(subsidy, to)

	tx := Transaction{
		ID:   nil,
		VIn:  []TXInput{txIn},
		VOut: []TXOutput{*txOut},
	}
	tx.ID = tx.Hash()

	return &tx
}

/*
Sing підписує всі входи транзакції
Метод приймає закритий ключ і масив попередніх транзакцій
*/
func (tx *Transaction) Sing(privetKey ecdsa.PrivateKey, prevTXs map[string]Transaction) {
	if tx.IsCoinbase() {
		return
	}

	txCopy := tx.TrimmedCopy()

	for inID, vin := range txCopy.VIn {
		prevTx := prevTXs[hex.EncodeToString(vin.TxId)]
		txCopy.VIn[inID].Signature = nil
		txCopy.VIn[inID].PubKey = prevTx.VOut[vin.VOut].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.VIn[inID].PubKey = nil

		r, s, err := ecdsa.Sign(rand.Reader, &privetKey, txCopy.ID)
		if err != nil {
			log.Panic(err)
		}
		signature := append(r.Bytes(), s.Bytes()...)

		tx.VIn[inID].Signature = signature
	}

}

/*
TrimmedCopy формуємо копію транзакції без підписів та публічних ключів
*/
func (tx *Transaction) TrimmedCopy() Transaction {
	var inputs []TXInput
	var outputs []TXOutput

	for _, vin := range tx.VIn {
		inputs = append(inputs, TXInput{
			TxId:      vin.TxId,
			VOut:      vin.VOut,
			Signature: nil,
			PubKey:    nil,
		})
	}

	for _, vOut := range tx.VOut {
		outputs = append(outputs, TXOutput{
			Value:      vOut.Value,
			PubKeyHash: vOut.PubKeyHash,
		})
	}

	txCopy := Transaction{
		ID:   tx.ID,
		VIn:  inputs,
		VOut: outputs,
	}

	return txCopy
}

/*
Hash повертає хеш транзакції
*/
func (tx *Transaction) Hash() []byte {
	var hash [32]byte

	txCopy := *tx
	txCopy.ID = []byte{}

	hash = sha256.Sum256(txCopy.Serialize())

	return hash[:]
}

/*
Serialize повертає копію сереалізованої транзакції
*/
func (tx *Transaction) Serialize() []byte {
	var encoded bytes.Buffer

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}

	return encoded.Bytes()
}

/*
Verify перевіряє підписи транзакції
*/
func (tx *Transaction) Verify(prevTXs map[string]Transaction) bool {
	txCopy := tx.TrimmedCopy()
	curve := elliptic.P256()

	for inID, vin := range tx.VIn {
		prevTx := prevTXs[hex.EncodeToString(vin.TxId)]
		txCopy.VIn[inID].Signature = nil
		txCopy.VIn[inID].PubKey = prevTx.VOut[vin.VOut].PubKeyHash
		txCopy.ID = txCopy.Hash()
		txCopy.VIn[inID].PubKey = nil

		r := big.Int{}
		s := big.Int{}
		sigLen := len(vin.Signature)
		r.SetBytes(vin.Signature[:(sigLen / 2)])
		s.SetBytes(vin.Signature[(sigLen / 2):])

		x := big.Int{}
		y := big.Int{}
		keyLen := len(vin.PubKey)
		x.SetBytes(vin.PubKey[:(keyLen / 2)])
		y.SetBytes(vin.PubKey[(keyLen / 2):])

		rawPubKey := ecdsa.PublicKey{
			Curve: curve,
			X:     &x,
			Y:     &y,
		}
		if ecdsa.Verify(&rawPubKey, txCopy.ID, &r, &s) == false {
			return false
		}
	}

	return true
}

// IsCoinbase визначає, чи є транзакція транзакцією Coinbase
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.VIn) == 1 && len(tx.VIn[0].TxId) == 0 && tx.VIn[0].VOut == -1
}

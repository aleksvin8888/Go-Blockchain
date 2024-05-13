package transaction

import (
	"blockchain1/lib/base58"
	"bytes"
	"encoding/gob"
	"log"
)

/*
TXOutput описує вихід транзакції
- Value - кількість монет, яку видає вихід
- PubKeyHash - хеш публічного ключа отримувача
*/
type TXOutput struct {
	Value      int
	PubKeyHash []byte
}

// TXOutputs масив виходів
type TXOutputs struct {
	Outputs []TXOutput
}

/*
NewTXOutput створює новий вихід
з вказаною кількістю монет та адресою отримувача
*/
func NewTXOutput(value int, address string) *TXOutput {
	txo := &TXOutput{
		value,
		nil,
	}
	txo.Lock([]byte(address))

	return txo
}

/*
Lock підписує ( блокує ) вивід
*/
func (out *TXOutput) Lock(address []byte) {
	pubKeyHash := base58.Decode(address)
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-4]
	out.PubKeyHash = pubKeyHash
}

/*
IsLockedWithKey  перевіряє, чи може вивід бути використаний власником публічного ключа
*/
func (out *TXOutput) IsLockedWithKey(pubKeyHash []byte) bool {
	return bytes.Compare(out.PubKeyHash, pubKeyHash) == 0
}

func (outs TXOutputs) Serialize() []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(outs)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

func DeserializeOutputs(data []byte) TXOutputs {
	var outputs TXOutputs

	dec := gob.NewDecoder(bytes.NewReader(data))
	err := dec.Decode(&outputs)
	if err != nil {
		log.Panic(err)
	}

	return outputs
}

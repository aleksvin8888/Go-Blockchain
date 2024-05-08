package transactions

import (
	"bytes"
	"crypto/sha256"
	"encoding/gob"
	"fmt"
	"log"
)

/*
subsidy — это сумма вознаграждения.
В Биткоине это число не хранится нигде и рассчитывается только на основе общего количества блоков: количество блоков делится на 210000.
Майнинг блока генезиса приносит 50 BTC,
и каждые 210000 блоков награда уменьшается вдвое.
В нашей реализации мы будем хранить вознаграждение как константу (по крайней мере на данный момент).
*/
const (
	subsidy = 100
)

/*
Transaction Транзакция представляет собой комбинацию входов и выходов:
Входы новой транзакции ссылаются на выходы предыдущей транзакции.
Выходы — место, где хранятся монеты.
В одной транзакции входы могут ссылаться на выходы нескольких транзакций.
Транзакции — это просто заблокированное скриптом значение,
которое может разблокировать лишь тот, кто его заблокировал.
*/
type Transaction struct {
	ID   []byte
	VIn  []TXInput
	VOut []TXOutput
}

/*
TXOutput Выходы — это место, в котором хранятся «монеты».
Каждый выход имеет сценарий разблокировки,
который определяет логику разблокировки выхода.
*/
type TXOutput struct {
	Value        int
	ScriptPubKey string
}

/*
TXInput  Вход ссылается на результат предыдущей транзакции и предоставляет данные (поле ScriptSig),
которые используются в сценарии разблокировки выхода,
чтобы разблокировать его и использовать его значение для создания новых выходов.
*/
type TXInput struct {
	TxId      []byte
	VOut      int
	ScriptSig string
}

/*
NewCoinbaseTX Транзакция coinbase — это особый тип транзакции,
который не требует ранее существующих выходов.
Он создает выходы (т. е. «Монеты») из ниоткуда.
*/
func NewCoinbaseTX(to, data string) *Transaction {
	if data == "" {
		data = fmt.Sprintf("Reward to '%s'", to)
	}

	txIn := TXInput{
		[]byte{},
		-1,
		data,
	}
	txOut := TXOutput{
		subsidy,
		to,
	}
	tx := Transaction{
		nil,
		[]TXInput{txIn},
		[]TXOutput{txOut},
	}
	tx.SetID()

	return &tx
}

// SetID устанавливает идентификатор транзакции
func (tx *Transaction) SetID() {
	var encoded bytes.Buffer
	var hash [32]byte

	enc := gob.NewEncoder(&encoded)
	err := enc.Encode(tx)
	if err != nil {
		log.Panic(err)
	}
	hash = sha256.Sum256(encoded.Bytes())
	tx.ID = hash[:]
}

/*
CanUnlockOutputWith проверяет, может ли вход разблокировать выход с unlockingData
*/
func (in *TXInput) CanUnlockOutputWith(unlockingData string) bool {
	return in.ScriptSig == unlockingData
}

/*
CanBeUnlockedWith проверяет, может ли выход быть разблокирован с помощью unlockingData
*/
func (out *TXOutput) CanBeUnlockedWith(unlockingData string) bool {
	return out.ScriptPubKey == unlockingData
}

// IsCoinbase проверяет, является ли транзакция транзакцией Coinbase
func (tx *Transaction) IsCoinbase() bool {
	return len(tx.VIn) == 1 && len(tx.VIn[0].TxId) == 0 && tx.VIn[0].VOut == -1
}

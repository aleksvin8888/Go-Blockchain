package transaction

import (
	wal "blockchain1/wallet"
	"bytes"
)

/*
TXInput представляє входи транзакції
- TxId - ідентифікатор транзакції, яка містить вихід, який використовується для входу.
- VOut - індекс вихідних даних в транзакції, яка містить вихід.
- Signature - підпис, який вказує, що власник виходу погоджується з витратою.
- PubKey - публічний ключ власника виходу.
*/
type TXInput struct {
	TxId      []byte
	VOut      int
	Signature []byte
	PubKey    []byte
}

/*
UsesKey перевіряє, чи вхід використовує певний ключ для розблокування
*/
func (in *TXInput) UsesKey(pubKeyHash []byte) bool {
	lockingHash := wal.HashPubKey(in.PubKey)

	return bytes.Compare(lockingHash, pubKeyHash) == 0
}

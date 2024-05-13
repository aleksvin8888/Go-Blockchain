package bloks

import (
	"blockchain1/lib/utils"
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
)

/*
targetBits
В Биткоине, «target bits» — это поле заголовка блока, которое хранит сложность,
на которой блок был добыт. Мы не будем строить корректирующийся алгоритм,
поэтому определим сложность, как глобальную константу.
якщо ви хочете збільшити складність майнінгу, вам потрібно зменшити targetBits.
Якщо хочете зменшити складність, збільшіть targetBits.
в реальних блокчейн-системах зазвичай регулюється динамічно,
щоб забезпечити бажаний час між знаходженням блоків,
навіть якщо загальна обчислювальна потужність мережі змінюється.

maxNonce - максимальне значення для лічильника nonce.
*/
const (
	targetBits = 17
	maxNonce   = math.MaxInt64
)

type ProofOfWork struct {
	Block  *Block
	Target *big.Int
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{
		Block:  b,
		Target: target,
	}
	return pow
}

/*
Run : Цей метод виконує процес майнінгу блоку в блокчейні,
використовуючи алгоритм доказу роботи (Proof of Work).
*/
func (pow *ProofOfWork) Run() (int, []byte) {
	var hashInt big.Int //  змінна для зберігання цілочисельного значення хешу
	var hash [32]byte   //  масив з 32 байтів для зберігання хешу блоку.
	nonce := 0          // лічильник, який використовується для знаходження правильного хешу.

	fmt.Printf("Mining the block")
	for nonce < maxNonce {
		data := pow.prepareData(nonce) // Викликається метод prepareData з поточним значенням nonce, щоб підготувати дані для хешування.
		hash = sha256.Sum256(data)     // Викликається sha256.Sum256(data) для обчислення хешу даних.

		fmt.Printf("\r%x", hash)  // Виводиться хеш в шістнадцятковому форматі.
		hashInt.SetBytes(hash[:]) // встановлюється як цілочисельне значення хешу.

		/*
				Перевірка згенерованого хешу:
			   - Використовуючи метод Cmp типу big.Int,
				порівнюється hashInt (цілочисельне представлення згенерованого хешу)
				з pow.target (цільовим значенням хешу, яке визначає складність майнінгу).
				Якщо hashInt менше pow.target (hashInt.Cmp(pow.target) == -1), то цикл припиняється,
				оскільки було знайдено валідний nonce.
		*/
		if hashInt.Cmp(pow.Target) == -1 {
			break
		} else {
			nonce++
		}

	}

	fmt.Print("\n\n")

	return nonce, hash[:]
}

/*
prepareData : Цей метод використовується для підготовки даних для хешування.
*/
func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.Block.PrevBlockHash,
			pow.Block.HashTransactions(),
			utils.IntToHex(pow.Block.Timestamp),
			utils.IntToHex(int64(targetBits)),
			utils.IntToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

/*
Validate : Цей метод використовується для перевірки правильності хешу блоку.
*/
func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.Block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.Target) == -1

	return isValid

}

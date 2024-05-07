package bloks

import (
	"bytes"
	"crypto/sha256"
	"fmt"
	"math"
	"math/big"
	"strconv"
	"time"
)

/*
В Биткоине, «target bits» — это поле заголовка блока, которое хранит сложность,
на которой блок был добыт. Мы не будем строить корректирующийся алгоритм,
поэтому определим сложность, как глобальную константу.
якщо ви хочете збільшити складність майнінгу, вам потрібно зменшити targetBits.
Якщо хочете зменшити складність, збільшіть targetBits.
в реальних блокчейн-системах зазвичай регулюється динамічно,
щоб забезпечити бажаний час між знаходженням блоків,
навіть якщо загальна обчислювальна потужність мережі змінюється.
*/
const (
	targetBits = 10
	maxNonce   = math.MaxInt64 // максимальне значення для лічильника nonce.
)

/*
Block : структура, що представляє блок у блокчейні.
- Timestamp - час створення блока.
- Data - дані, що містяться в блоку.
- PrevBlockHash - зберігає хеш попереднього блоку.
- Hash - містить хеш поточного блоку.
*/
type Block struct {
	Timestamp     int64
	Data          []byte
	PrevBlockHash []byte
	Hash          []byte
	Nonce         int
}

type ProofOfWork struct {
	block  *Block
	target *big.Int
}

func NewBlock(data string, prevBlockHash []byte) *Block {
	block := &Block{
		Timestamp:     time.Now().Unix(),
		Data:          []byte(data),
		PrevBlockHash: prevBlockHash,
		Hash:          []byte{},
	}

	pow := NewProofOfWork(block)
	nonce, hash := pow.Run()

	block.Hash = hash[:]
	block.Nonce = nonce

	return block
}

func NewProofOfWork(b *Block) *ProofOfWork {
	target := big.NewInt(1)
	target.Lsh(target, uint(256-targetBits))

	pow := &ProofOfWork{
		b,
		target,
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

	fmt.Printf("Mining the block containing \"%s\"\n", pow.block.Data)
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
		if hashInt.Cmp(pow.target) == -1 {
			break
		} else {
			nonce++
		}

	}

	fmt.Print("\n\n")

	return nonce, hash[:]
}

func (pow *ProofOfWork) prepareData(nonce int) []byte {
	data := bytes.Join(
		[][]byte{
			pow.block.PrevBlockHash,
			pow.block.Data,
			intToHex(pow.block.Timestamp),
			intToHex(int64(targetBits)),
			intToHex(int64(nonce)),
		},
		[]byte{},
	)

	return data
}

func (pow *ProofOfWork) Validate() bool {
	var hashInt big.Int

	data := pow.prepareData(pow.block.Nonce)
	hash := sha256.Sum256(data)
	hashInt.SetBytes(hash[:])

	isValid := hashInt.Cmp(pow.target) == -1

	return isValid

}

func intToHex(n int64) []byte {
	return []byte(strconv.FormatInt(n, 16))
}

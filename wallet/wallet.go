package wallet

import (
	"blockchain1/lib/base58"
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/sha256"
	"golang.org/x/crypto/ripemd160"
	"log"
	"os"
)

const (
	version            = byte(0x00)
	addressChecksumLen = 4
)

/*
Wallet представляє гаманець, який зберігає приватний ключ та публічний ключ.
фактично це пара ключів (приватний та публічний)
*/
type Wallet struct {
	PrivateKey []byte
	PublicKey  []byte
}

/*
NewWallet створює і повертає новий гаманець
*/
func NewWallet() *Wallet {
	private, public := newKeyPair()
	wallet := Wallet{
		PrivateKey: private,
		PublicKey:  public,
	}

	return &wallet
}

/*
GetAddress повертає адресу гаманця у вигляді байтів (base58)
*/
func (w Wallet) GetAddress() []byte {
	pubKeyHash := HashPubKey(w.PublicKey)

	versionedPayload := append([]byte{version}, pubKeyHash...)
	checksum := checksum(versionedPayload)
	fullPayload := append(versionedPayload, checksum...)
	address := base58.Encode(fullPayload)

	return address
}

/*
ValidateAddress перевіряє, чи адреса валідна
*/
func ValidateAddress(address string) bool {
	pubKeyHash := base58.Decode([]byte(address))
	actualChecksum := pubKeyHash[len(pubKeyHash)-addressChecksumLen:]
	version := pubKeyHash[0]
	pubKeyHash = pubKeyHash[1 : len(pubKeyHash)-addressChecksumLen]
	targetChecksum := checksum(append([]byte{version}, pubKeyHash...))

	return bytes.Compare(actualChecksum, targetChecksum) == 0
}

/*
HashPubKey генерує хеш публічного ключа
*/
func HashPubKey(pubKey []byte) []byte {
	publicSHA256 := sha256.Sum256(pubKey)
	RIPEMD160Hash := ripemd160.New()
	_, err := RIPEMD160Hash.Write(publicSHA256[:])
	if err != nil {
		log.Print("Error: ", err)
		os.Exit(1)
	}
	publicRIPEMD160 := RIPEMD160Hash.Sum(nil)

	return publicRIPEMD160
}

/*
newKeyPair генерує нову пару ключів
*/
func newKeyPair() ([]byte, []byte) {
	curve := elliptic.P256()
	private, err := ecdsa.GenerateKey(curve, rand.Reader)
	if err != nil {
		log.Print("Error: ", err)
		os.Exit(1)
	}
	privetKeyBytes := private.D.Bytes()
	pubKeyBytes := append(private.PublicKey.X.Bytes(), private.PublicKey.Y.Bytes()...)

	return privetKeyBytes, pubKeyBytes
}

/*
checksum генерує контрольну суму public key
*/
func checksum(payload []byte) []byte {
	firstSHA := sha256.Sum256(payload)
	secondSHA := sha256.Sum256(firstSHA[:])

	return secondSHA[:addressChecksumLen]
}

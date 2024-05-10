package utils

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"encoding/binary"
	"log"
	"math/big"
)

// IntToHex converts an int64 to a byte array
func IntToHex(num int64) []byte {
	buff := new(bytes.Buffer)
	err := binary.Write(buff, binary.BigEndian, num)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

// ReverseBytes reverses a byte array
func ReverseBytes(data []byte) {
	for i, j := 0, len(data)-1; i < j; i, j = i+1, j-1 {
		data[i], data[j] = data[j], data[i]
	}
}

func PrivateKeyFromBytes(privetKeyBytes []byte) (*ecdsa.PrivateKey, error) {
	curve := elliptic.P256()
	privet := new(ecdsa.PrivateKey)
	privet.PublicKey.Curve = curve
	privet.D = new(big.Int).SetBytes(privetKeyBytes)
	privet.PublicKey.X, privet.PublicKey.Y = curve.ScalarBaseMult(privetKeyBytes)
	return privet, nil
}

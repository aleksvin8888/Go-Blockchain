package wallets

import (
	wal "blockchain1/wallet"
	"bytes"
	"crypto/elliptic"
	"encoding/gob"
	"fmt"
	"log"
	"os"
)

const walletFile = "wallet_%s.dat"

/*
Wallets зберігає колекцію гаманців
*/
type Wallets struct {
	Wallets map[string]*wal.Wallet
}

/*
NewWallets створює гаманці та заповнює їх з файлу, якщо він існує
*/
func NewWallets(nodeID string) (*Wallets, error) {
	wallets := Wallets{}
	wallets.Wallets = make(map[string]*wal.Wallet)

	err := wallets.LoadFromFile(nodeID)

	return &wallets, err
}

/*
CreateWallet створює новий гаманець та додає його до колекції гаманців
*/
func (ws *Wallets) CreateWallet() string {
	wallet := wal.NewWallet()
	address := fmt.Sprintf("%s", wallet.GetAddress())

	ws.Wallets[address] = wallet

	return address
}

/*
GetWallet повертає гаманець за адресою
*/
func (ws *Wallets) GetWallet(address string) wal.Wallet {
	return *ws.Wallets[address]
}

/*
GetAddresses повертає масив адрес гаманців
*/
func (ws *Wallets) GetAddresses() []string {
	var addresses []string

	for address := range ws.Wallets {
		addresses = append(addresses, address)
	}

	return addresses
}

/*
LoadFromFile завантажує гаманці з файлу у структуру Wallets
*/
func (ws *Wallets) LoadFromFile(nodeID string) error {
	walletFile := fmt.Sprintf(walletFile, nodeID)

	if _, err := os.Stat(walletFile); os.IsNotExist(err) {
		return err
	}

	fileContent, err := os.ReadFile(walletFile)
	if err != nil {
		log.Panic(err)
	}

	var wallets Wallets
	gob.Register(elliptic.P256())
	decoder := gob.NewDecoder(bytes.NewReader(fileContent))
	err = decoder.Decode(&wallets)
	if err != nil {
		log.Panic(err)
	}

	ws.Wallets = wallets.Wallets

	return nil

}

/*
SaveToFile зберігає гаманці у файл
*/
func (ws *Wallets) SaveToFile(nodeID string) {
	var content bytes.Buffer
	walletFile := fmt.Sprintf(walletFile, nodeID)
	gob.Register(elliptic.P256())

	encoder := gob.NewEncoder(&content)
	err := encoder.Encode(ws)
	if err != nil {
		log.Panic(err)
	}

	err = os.WriteFile(walletFile, content.Bytes(), 0644)
	if err != nil {
		log.Panic(err)
	}
}

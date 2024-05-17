package server

import (
	"blockchain1/blockchain"
	"blockchain1/bloks"
	"blockchain1/transaction"
	"bytes"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"io"
	"log"
	"net"
)

const (
	protocol      = "tcp"
	nodeVersion   = 1
	commandLength = 12
)

var nodeAddress string
var miningAddress string

/*
KnownNodes - зберігає список відомих вузлів зазвичай це має бути список DNS
за допомогою якого можна знайти вузол в мережі
для локального тестування використовуємо список з одного вузла
який фактично являється основним центральним вузлом з адресою
"127.0.0.1:3000"
*/
var KnownNodes = []string{"127.0.0.1:3000"}
var blocksInTransit [][]byte
var TransactionMemoryPool = make(map[string]transaction.Transaction)

type ver struct {
	Version    int
	BestHeight int
	AddrFrom   string
}

type getBlocks struct {
	AddrFrom string
}

type inv struct {
	AddrFrom string
	Type     string
	Items    [][]byte
}

type getData struct {
	AddrFrom string
	Type     string
	ID       []byte
}

type block struct {
	AddrFrom string
	Block    []byte
}

type tx struct {
	AddFrom     string
	Transaction []byte
}

/*
StartServer виконує запуск сервера ( вузла блокчейну node )
*/
func StartServer(nodeID string, minerAddress string) {
	// формуємо адресу вузла nodeID може мати наступні значення 3000, 3001, 3002 це для локального тестування
	nodeAddress = fmt.Sprintf("127.0.0.1:%s", nodeID)

	miningAddress = minerAddress

	//ініціюємо прослуховування мережі на вказаній адресі protocol = "tcp"
	ln, err := net.Listen(protocol, nodeAddress)
	if err != nil {
		log.Fatal(err)
	}
	defer func() { _ = ln.Close() }()
	// ініціалізуємо новий екземпляр блокчейну з вказаним nodeID
	bc := blockchain.NewBlockchain(nodeID)

	/*
			 якщо поточний вузол не є першим відомим вузлом
			 у нашій реалізації це вузол з адресою " 127.0.0.1:3000 "
		     то відправляємо запит для встановлення звязку з осно вним вузлом
	*/
	if nodeAddress != KnownNodes[0] {
		sendVersion(KnownNodes[0], bc)
	}
	/*
		запуск безкінечного циклу для прийому запитів від інших вузлів
	*/
	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Panic(err)
		}
		go handleConnection(conn, bc)
	}

}

/*
handleConnection приймає вхідні запити і передає їх на обробку відповідним функціям обробки
*/
func handleConnection(conn net.Conn, bc *blockchain.Blockchain) {
	request, err := io.ReadAll(conn)
	if err != nil {
		log.Panic(err)
	}
	command := bytesToCommand(request[:commandLength])
	fmt.Printf("Received %s command\n", command)

	switch command {
	case "block":
		handleBlock(request, bc)
	case "inv":
		handleInv(request)
	case "getblocks":
		handleGetBlocks(request, bc)
	case "getdata":
		handleGetData(request, bc)
	case "tx":
		handleTx(request, bc)
	case "version":
		handleVersion(request, bc)
	default:
		fmt.Println("Unknown command!")
	}

	defer func() { _ = conn.Close() }()
}

/*
handleVersion обробляє вхідний запит типу "version"
та забезпечує синхронізацію станів блокчейну між вузлами та допомагає підтримувати актуальний список вузлів в мережі.
*/
func handleVersion(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload ver

	// декодуємо дані з запиту
	// та записуємо їх у структуру ver
	// відкидається перших n байтів визначених у константі які вказують на тип команди
	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	myBestHeight := bc.GetBestHeight()        // отримуємо висоту останнього блоку в ланцюгу поточного вузла
	foreignerBestHeight := payload.BestHeight // отримуємо висоту останнього блоку в ланцюгу яка прийшла у запиті

	if myBestHeight < foreignerBestHeight {
		/* якщо у поточного вузла висота менша
		виконуємо запит на отримання блоків від вузла з вищою висотою
		*/
		sendGetBlocks(payload.AddrFrom)
	} else if myBestHeight > foreignerBestHeight {
		/*якщо у поточного вузла висота більша
		виконуємо запит у відповідь і передаємо дані про свою висоту блоку
		*/
		sendVersion(payload.AddrFrom, bc)
	}

	// додаємо адресу вузла який надіслав запит до списку відомих вузлів
	if !nodeIsKnown(payload.AddrFrom) {
		KnownNodes = append(KnownNodes, payload.AddrFrom)
	}
}

func handleTx(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload tx

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	txData := payload.Transaction
	tx := transaction.DeserializeTransaction(txData)
	TransactionMemoryPool[hex.EncodeToString(tx.ID)] = tx

	fmt.Printf("curent node addr:  %s\n", nodeAddress)
	if nodeAddress == KnownNodes[0] {
		for _, node := range KnownNodes {
			if node != nodeAddress && node != payload.AddFrom {
				sendInv(node, "tx", [][]byte{tx.ID})
			}
		}
	} else {
		if len(TransactionMemoryPool) >= 2 && len(miningAddress) > 0 {
		MineTransactions:
			var txs []*transaction.Transaction

			for id := range TransactionMemoryPool {
				tx := TransactionMemoryPool[id]
				if bc.VerifyTransaction(&tx) {
					txs = append(txs, &tx)
				}
			}
			if len(txs) == 0 {
				fmt.Println("All transactions are invalid! Waiting for new ones...")
				return
			}
			cbTx := transaction.NewCoinbaseTX(miningAddress, "")
			txs = append(txs, cbTx)

			newBlock := bc.MineBlock(txs)
			UTXOSet := blockchain.UTXOSet{
				Blockchain: bc,
			}
			UTXOSet.Reindex()

			fmt.Println("New block is mined!")

			for _, tx := range txs {
				txID := hex.EncodeToString(tx.ID)
				delete(TransactionMemoryPool, txID)
			}

			for _, node := range KnownNodes {
				if node != nodeAddress {
					sendInv(node, "block", [][]byte{newBlock.Hash})
				}
			}

			if len(TransactionMemoryPool) > 0 {
				goto MineTransactions
			}
		}
	}
}

/*
handleBlock обробляє вхідний запит від іншого вузла з новим блоком
і зберігає його
*/
func handleBlock(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload block

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	blockData := payload.Block
	block := bloks.DeserializeBlock(blockData)

	fmt.Println("Received a new block!")
	bc.AddBlock(block)

	fmt.Printf("Added block %x\n", block.Hash)

	if len(blocksInTransit) > 0 {
		blockHash := blocksInTransit[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		blocksInTransit = blocksInTransit[1:]
	} else {
		UTXOSet := blockchain.UTXOSet{
			Blockchain: bc,
		}
		UTXOSet.Reindex()
	}
}

/*
handleGetData обробляє запит від іншого вузла на отримання блоку або транзакції
за їх хешем
*/
func handleGetData(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload getData

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	if payload.Type == "block" {
		block, err := bc.GetBlock(payload.ID)
		if err != nil {
			return
		}
		// відправляємо блок на вказану адресу
		sendBlock(payload.AddrFrom, &block)
	}

	if payload.Type == "tx" {
		txID := hex.EncodeToString(payload.ID)
		tx := TransactionMemoryPool[txID]

		SendTx(payload.AddrFrom, &tx)

	}
}

/*
handleInv обробляє вхідний запит від іншого вузла з інформацією про наявність блоків або транзакцій
*/
func handleInv(request []byte) {
	var buff bytes.Buffer
	var payload inv

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	fmt.Printf("Recevied inventory with %d %s\n", len(payload.Items), payload.Type)

	// обробляємо інформацію про блоки
	if payload.Type == "block" {
		blocksInTransit = payload.Items
		blockHash := payload.Items[0]
		sendGetData(payload.AddrFrom, "block", blockHash)

		var newInTransit [][]byte
		for _, b := range blocksInTransit {
			if bytes.Compare(b, blockHash) != 0 {
				newInTransit = append(newInTransit, b)
			}
		}
		blocksInTransit = newInTransit
	}

	// обробляємо інформацію про транзакції
	if payload.Type == "tx" {
		txID := payload.Items[0]

		if TransactionMemoryPool[hex.EncodeToString(txID)].ID == nil {
			sendGetData(payload.AddrFrom, "tx", txID)
		}
	}
}

/*
handleGetBlocks обробляє запит від іншого вузла на отримання відсутніх блоків
*/
func handleGetBlocks(request []byte, bc *blockchain.Blockchain) {
	var buff bytes.Buffer
	var payload getBlocks

	buff.Write(request[commandLength:])
	dec := gob.NewDecoder(&buff)
	err := dec.Decode(&payload)
	if err != nil {
		log.Panic(err)
	}

	// отримуємо хеші усіх блоків з поточної ноди
	blocks := bc.GetBlockHashes()
	// формуємо запит з даними про хеші блоків
	sendInv(payload.AddrFrom, "block", blocks)
}

/*
sendVersion отримує адресу вузла ( сервера ) і екземпляр блокчейну
підготовлює дані у вигляді байтів для відправки
передає дані та адресу вузла ( сервера ) для відправки у функцію send Data
*/
func sendVersion(addr string, bc *blockchain.Blockchain) {
	//отримуємо число яке представляє номер останньго блоку в ланцюгу
	// фактично це загальна кількість блоків в ланцюгу
	// формуємо структуру ver з даними
	bestHeight := bc.GetBestHeight()
	// сереалізуємо структуру ver в байтовий масив
	payload := gobEncode(ver{
		Version:    nodeVersion,
		BestHeight: bestHeight,
		AddrFrom:   nodeAddress,
	})
	// обєднуємо дані у одне байтове представлення
	request := append(commandToBytes("version"), payload...)
	sendData(addr, request)
}

/*
SendTx відправляє транзакцію на вказану адресу
*/
func SendTx(addr string, tnx *transaction.Transaction) {
	data := tx{
		AddFrom:     nodeAddress,
		Transaction: tnx.Serialize(),
	}
	payload := gobEncode(data)
	request := append(commandToBytes("tx"), payload...)

	sendData(addr, request)
}

/*
sendBlock відправляє блок на вказану адресу
*/
func sendBlock(addr string, b *bloks.Block) {
	data := block{
		AddrFrom: nodeAddress,
		Block:    b.Serialize(),
	}
	payload := gobEncode(data)
	request := append(commandToBytes("block"), payload...)

	sendData(addr, request)
}

/*
sendGetData відправляє запит на віддалений вузол для отримання одого блоку або транзакції
*/
func sendGetData(address, kind string, id []byte) {
	payload := gobEncode(getData{
		AddrFrom: nodeAddress,
		Type:     kind, // block or tx (transaction)
		ID:       id,
	})
	request := append(commandToBytes("getdata"), payload...)

	sendData(address, request)
}

/*
sendGetBlocks відправляє запит на віддалений вузол для отримання відсутніх блоків
*/
func sendGetBlocks(address string) {
	// формуємо структуру getBlocks з адресою відправника
	payload := gobEncode(getBlocks{
		AddrFrom: nodeAddress,
	})
	// додаємо команду до байтового представлення даних
	request := append(commandToBytes("getblocks"), payload...)

	sendData(address, request)
}

/*
sendInv формує та відправляє дані про наявність блоків або транзакцій на вказану адресу
*/
func sendInv(address, kind string, items [][]byte) {
	inventory := inv{
		AddrFrom: nodeAddress,
		Type:     kind,  // block or tx (transaction)
		Items:    items, // хеші блоків або транзакцій
	}
	payload := gobEncode(inventory)
	request := append(commandToBytes("inv"), payload...)

	sendData(address, request)
}

/*
sendData відправляє дані на вказану адресу
*/
func sendData(addr string, data []byte) {
	// відкриваємо з'єднання з вказаним адресом
	conn, err := net.Dial(protocol, addr)
	/*
		Якщо виникає помилка при спробі встановлення з'єднання,
		виводиться повідомлення, що адреса addr недоступна.
		Тоді виконується перебір відомих вузлів мережі (KnownNodes),
		і вузол з адресою addr видаляється зі списку.
		Після цього функція завершує роботу.
	*/
	if err != nil {
		fmt.Printf("%s is not available\n", addr)
		var updatedNodes []string
		for _, node := range KnownNodes {
			if node != addr {
				updatedNodes = append(updatedNodes, node)
			}
		}
		KnownNodes = updatedNodes
		return
	}
	defer func() { _ = conn.Close() }()

	// відправляємо дані на вказану адресу через відкрите з'єднання
	_, err = io.Copy(conn, bytes.NewReader(data))
	if err != nil {
		log.Panic(err)
	}
}

/*
gobEncode сереалізує отримані дані в байтовий масив і повертає їх
*/
func gobEncode(data interface{}) []byte {
	var buff bytes.Buffer

	enc := gob.NewEncoder(&buff)
	err := enc.Encode(data)
	if err != nil {
		log.Panic(err)
	}

	return buff.Bytes()
}

/*
commandToBytes конвертує строку в байтовий масив фіксованої довжини визначений у константі
*/
func commandToBytes(command string) []byte {
	var byt [commandLength]byte

	for i, c := range command {
		byt[i] = byte(c)
	}

	return byt[:]
}

/*
bytesToCommand конвертує байтовий масив в строку
*/
func bytesToCommand(bytes []byte) string {
	var command []byte

	for _, b := range bytes {
		if b != 0x0 {
			command = append(command, b)
		}
	}

	return fmt.Sprintf("%s", command)
}

/*
nodeIsKnown перевіряє чи вузол з адресою addr вже відомий і збережений
*/
func nodeIsKnown(addr string) bool {
	for _, node := range KnownNodes {
		if node == addr {
			return true
		}
	}

	return false
}

package main

import (
	"crypto/sha256"
	"encoding/gob"
	"encoding/hex"
	"fmt"
	"log"
	"net"
	"os"
	"sync"
)

type Skill struct {
}
type Course struct {
	code        string
	name        string
	creditHours int
	grade       string
}
type Project struct {
	name     string
	document string
	course   Course
}
type Peer struct {
	ListeningAddress string
	Role             string //1 for user 0 for miner
}
type Data struct {
	MinerList    []Peer
	ClientsSlice []Peer
	ChainHead    *Block
}

type Block struct {
	course      Course
	project     Project
	PrevPointer *Block
	PrevHash    string
	CurrentHash string
}

//var chainHead *Block
var globalData Data
var mutex = &sync.Mutex{}

//var globalData Data

func CalculateHash(inputBlock *Block) string {

	var temp string
	temp = inputBlock.course.code + inputBlock.project.name
	h := sha256.New()
	h.Write([]byte(temp))
	sum := hex.EncodeToString(h.Sum(nil))

	// sum := sha256.Sum256([]byte(temp))

	return sum
}
func InsertOnlyBlock(newBlock *Block, chainHead *Block) *Block {
	newBlock.CurrentHash = CalculateHash(newBlock)

	if chainHead == nil {
		chainHead = newBlock
		fmt.Println("Block Inserted")
		return chainHead
	}
	newBlock.PrevPointer = chainHead
	newBlock.PrevHash = chainHead.CurrentHash

	fmt.Println("Block Course and Project Inserted")
	return newBlock

}

func StartListening(listeningAddress string, node string) {
	//var chainHead *Block
	if node == "admin" {

	} else if node == "miner" {
		ln, err := net.Listen("tcp", ":"+listeningAddress)
		if err != nil {
			log.Fatal(err, ln)
		}
		//	ClientsSlice := make([]Peer, 10)
		//	addchan := make(chan Peer)
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err)
				continue
			}
			// newClient := Peer{
			// 	Conn: conn,
			// }
			// ClientsSlice = append(ClientsSlice, newClient)
			// go broadcastBlockchaintoPeer(conn)

			// go receiveBlockchainfromPeer(conn)

			go MinerverifyBlock(conn)

		}
	}
}
func MinerverifyBlock(conn net.Conn) {
	var recvdBlock *Block
	//fmt.Println("block: ", recvdBlock.course.name)
	dec2 := gob.NewDecoder(conn)
	err2 := dec2.Decode(&recvdBlock)
	if err2 != nil {
		//handle error
		fmt.Println("err")
	} else {
		UpdateChan <- "start mining"
		fmt.Println("Block Verified")
		InsertOnlyBlock(recvdBlock, globalData.ChainHead)
	}
}
func WriteString(conn net.Conn, details Peer) {
	fmt.Println("Peer: ", details)
	gobEncoder := gob.NewEncoder(conn)
	err := gobEncoder.Encode(details)

	if err != nil {
		//	log.Println(err)
	}
}

var UpdateChan = make(chan string)

func readAdminData(conn net.Conn) {
	for {
		//var globe Data
		var globe Data
		gobEncoder := gob.NewDecoder(conn)
		//Stuck
		err1 := gobEncoder.Decode(&globe)
		//Stuck
		//	fmt.Println("In Admindata: ", globe)
		if err1 != nil {
			log.Println(err1)
		}
		fmt.Println("In read admin data:")
		globalData = globe

		<-UpdateChan
	}
}
func ViewMinerData() {
	for i := 0; i < len(globalData.ClientsSlice); i++ {
		if globalData.ClientsSlice[i].Role == "miner" {
			fmt.Println("Miners connected to system:")
			fmt.Print(" Their address: ", globalData.ClientsSlice[i].ListeningAddress)
		}
	}
}

func main() {

	satoshiAddress := os.Args[1]
	myListeningAddress := os.Args[2]
	log.Println(satoshiAddress, myListeningAddress)

	conn, err := net.Dial("tcp", ":"+satoshiAddress)
	if err != nil {
		log.Fatal(err)
	}
	//The function below launches the server, uses different second argument
	//It then starts a routine for each connection request received
	//	role := "user"

	myPeer := Peer{
		ListeningAddress: string(myListeningAddress),
		Role:             "miner",
	}
	go StartListening(myListeningAddress, "miner")

	WriteString(conn, myPeer)
	log.Println("I Verifier Sending my listening address to Admin")

	go readAdminData(conn)

	select {}
}

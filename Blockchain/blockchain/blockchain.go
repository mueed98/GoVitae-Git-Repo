package blockchain

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"encoding/gob"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strconv"
	"sync"
	"time"

	db "../Database"
	model "../Models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"golang.org/x/crypto/bcrypt"
	gomail "gopkg.in/mail.v2"

	jwt "github.com/dgrijalva/jwt-go"
	"github.com/gorilla/mux"
	//"github.com/gorilla/websocket"
)

var chainHead *Block

type Skill struct {
}
type Course struct {
	Code        string
	Name        string
	CreditHours int
	Grade       string
}
type Project struct {
	Name     string
	Document string
	Course   Course
}

type Block struct {
	Course      Course
	Project     Project
	PrevPointer *Block
	PrevHash    string
	CurrentHash string
	BlockNo     int
	Status      string
	Email       string
}

type ListTheBlock struct {
	Course      []Course
	Project     []Project
	PrevPointer []*Block
	PrevHash    []string
	CurrentHash []string
	BlockNo     []int
	Status      []string
	Email       []string
}

type Client struct {
	ListeningAddress string
	Types            bool //true for node and false for miner
	Mail             string
}
type Combo struct {
	ClientsSlice []Client
	ChainHead    *Block
}
type Connected struct {
	Conn net.Conn
}

var count int = 0
var stuff Combo
var localData []Connected
var mutex = &sync.Mutex{}

var tokenString = ""
var urlLogin = ""
var chainHeadArray []*Block

//var nodes = make(map[*websocket.Conn]bool) // connected clients
//var upgrader = websocket.Upgrader{
//CheckOrigin: func(r *http.Request) bool {
//return true
//},
//}

func ReadBlockchainFile() {
	// file, _ := ioutil.ReadFile("blockchainFile.json")
	//
	// _ = json.Unmarshal([]byte(file), &chainHeadArray)

	file, err := os.Open("blockchainFile.json")
	if err != nil {
		log.Println("Can't read file")
	}
	defer file.Close()

	decoder := json.NewDecoder(file)
	decoder.Token()
	block := Block{}
	// Appends decoded object to dataArr until every object gets parsed
	for decoder.More() {
		decoder.Decode(&block)
		chainHead = InsertCourse(block)
	}
	stuff.ChainHead = chainHead
	ListBlocks(chainHead)
}

func WriteBlockchainFile(chainHead []Block) {

	file, _ := json.MarshalIndent(chainHead, "", " ")
	_ = ioutil.WriteFile("blockchainFile.json", file, 0644)
	fmt.Println("file")

}

func GetBlockhainArray(chainHead *Block) []Block {
	var data []Block
	i := 0
	var block Block
	for chainHead != nil {
		block.Email = chainHead.Email
		if (chainHead.Course != Course{}) {
			block.Course = chainHead.Course
		}
		block.Status = chainHead.Status
		if (chainHead.Project != Project{}) {
			block.Project = chainHead.Project
		}
		data = append(data, block)
		chainHead = chainHead.PrevPointer
		i++

	}
	return data

}

//256bit
func CalculateHash(inputBlock *Block) string {

	var temp string
	temp = inputBlock.Course.Code + inputBlock.Project.Name
	h := sha256.New()
	h.Write([]byte(temp))
	sum := hex.EncodeToString(h.Sum(nil))

	// sum := sha256.Sum256([]byte(temp))

	return sum
}
func InsertBlock(course Course, project Project, chainHead *Block) *Block {
	newBlock := &Block{
		//Hash here
		Course:  course,
		Project: project,
	}
	newBlock.CurrentHash = CalculateHash(newBlock)
	fmt.Println("In insertion: ", CalculateHash(newBlock))

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
func InsertProject(project Project, chainHead *Block) *Block {
	newBlock := &Block{
		//Hash here
		Project: project,
	}
	newBlock.CurrentHash = CalculateHash(newBlock)

	if chainHead == nil {
		chainHead = newBlock
		fmt.Println("Block Inserted")
		return chainHead
	}
	newBlock.PrevPointer = chainHead
	newBlock.PrevHash = chainHead.CurrentHash

	fmt.Println("Project Block Inserted")
	return newBlock

}

// Changing InsertCourse Code //
func InsertCourse(myBlock Block) *Block {

	myBlock.CurrentHash = CalculateHash(&myBlock)
	fmt.Println("Course Hash, ", CalculateHash(&myBlock))
	if chainHead == nil {
		myBlock.BlockNo = count
		myBlock.PrevHash = "null"
		myBlock.Status = "Pending"
		chainHead = &myBlock
		//	fmt.Println("Genesis Block Inserted")
		return chainHead
	}
	count = count + 1
	myBlock.PrevPointer = chainHead
	myBlock.PrevHash = chainHead.CurrentHash
	myBlock.BlockNo = count
	myBlock.Status = "Pending"

	fmt.Println("Course Block Inserted")
	return &myBlock

}

func ChangeCourse(oldCourse Course, newCourse Course, chainHead *Block) {
	present := false
	for chainHead != nil {
		if chainHead.Course == oldCourse {
			chainHead.Course = newCourse
			present = true
			break
		}

		//fmt.Printf("%v ->", chainHead.transactions)
		chainHead = chainHead.PrevPointer
	}
	if present == false {
		fmt.Println("Input Course not found")
		return
	}
	fmt.Println("Block Course Changed")

	chainHead.CurrentHash = CalculateHash(chainHead)

}

func ChangeProject(oldProject Project, newProject Project, chainHead *Block) {
	present := false
	for chainHead != nil {
		if chainHead.Project == oldProject {
			chainHead.Project = newProject
			present = true
			break
		}

		//fmt.Printf("%v ->", chainHead.transactions)
		chainHead = chainHead.PrevPointer
	}
	if present == false {
		fmt.Println("Input Course not found")
		return
	}
	fmt.Println("Block Course Changed")

	chainHead.CurrentHash = CalculateHash(chainHead)

}

func ListBlocks(chainHead *Block) {

	for chainHead != nil {
		fmt.Print("Block NO: ", chainHead.BlockNo)
		fmt.Print(" Current Hash: ", chainHead.CurrentHash)
		if chainHead.PrevHash == "" {
			fmt.Print(" Previous Hash: ", "Null")
		} else {
			fmt.Print(" Previous Hash: ", chainHead.PrevHash)
		}

		fmt.Print(" Course: ", chainHead.Course.Name)
		fmt.Print(" Project: ", chainHead.Project.Name)
		fmt.Print(" -> ")
		chainHead = chainHead.PrevPointer

	}
	fmt.Println()

}

func VerifyChain(chainHead *Block) { //What to do?
	for chainHead != nil {
		if chainHead.PrevPointer != nil {
			if chainHead.PrevHash != chainHead.PrevPointer.CurrentHash {
				fmt.Println("Blockchain Compromised")
				return
			}
		}

		chainHead = chainHead.PrevPointer
	}
	fmt.Println("Blockchain Verified")
}

func Length(chainHead *Block) int {
	sum := 0
	for chainHead != nil {

		chainHead = chainHead.PrevPointer
		sum++
	}
	return sum

}
func sendBlockchain(c net.Conn, chainHead *Block) {

	log.Println("A client has connected",

		c.RemoteAddr())
	gobEncoder := gob.NewEncoder(c)
	err := gobEncoder.Encode(chainHead)
	if err != nil {

		log.Println(err)

	}

}

func InsertCourse1(course Course, chainHead *Block) *Block {
	newBlock := &Block{
		//Hash here
		Course: course,
	}
	newBlock.CurrentHash = CalculateHash(newBlock)

	if chainHead == nil {
		chainHead = newBlock
		chainHead.BlockNo = count
		fmt.Println("Block Inserted")
		return chainHead
	}
	count = count + 1
	newBlock.PrevPointer = chainHead
	newBlock.PrevHash = chainHead.CurrentHash
	newBlock.BlockNo = count

	fmt.Println("Course Block Inserted")
	return newBlock

}
func getCourse(ChainHead *Block) []Block {
	var courses []Block
	for chainHead != nil {
		courses = append(courses, *chainHead)
		chainHead = chainHead.PrevPointer
	}
	//	fmt.Println("Yo")
	return courses
}
func WriteString(conn net.Conn, myListeningAddress Client) {
	Satoshiconn = conn
	gobEncoder := gob.NewEncoder(conn)
	err := gobEncoder.Encode(myListeningAddress)
	if err != nil {
		log.Println("In Write String: ", err)
	}
}

func SendChain(conn net.Conn) {
	gobEncoder := gob.NewEncoder(conn)
	err := gobEncoder.Encode(chainHead)
	if err != nil {
		log.Println("In Write Chain: ", err)
	}
}

var Satoshiconn net.Conn
var clientsSlice []Client
var rwchan = make(chan string)

func handleConnection(conn net.Conn, addchan chan Client) {
	// newClient := Connected{
	// 	Conn: conn,
	// }
	Clientz := Client{}
	//var ListeningAddress string
	dec := gob.NewDecoder(conn)
	err := dec.Decode(&Clientz)
	if err != nil {
		//handle error
	}

	// newClient.ListeningAddress = ListeningAddress
	fmt.Println("inHandle: ", Clientz.ListeningAddress)
	addchan <- Clientz
	//WaitForQuorum()

}

var nodesSlice []Client
var minechan = make(chan Client)

var blockchan = make(chan Block)
var Minedblock Block

var newchan = make(chan *Block)

var NewChain bool

func handlePeer(conn net.Conn) {

	//	Clientz := Client{}
	block := Block{}
	//var ListeningAddress string
	dec := gob.NewDecoder(conn)
	err := dec.Decode(&block)
	if err != nil {
		//handle error
		log.Print("Eror in receiveing block", block)
	}

	// newClient.ListeningAddress = ListeningAddress
	fmt.Println("inHandlePeer: ", block)
	blockchan <- block

}
func ReceiveChain(conn net.Conn) *Block {
	fmt.Println("In func")
	var block *Block
	gobEncoder := gob.NewDecoder(conn)
	err := gobEncoder.Decode(&block)
	if err != nil {
		log.Println(err)
	}
	fmt.Println("Received chain")
	chainHead = block
	stuff.ChainHead = chainHead
	ListBlocks(chainHead)

	//chainHead = InsertCourse(block)
	return block
}
func ReceiveMinerChain(conn net.Conn) *Block {
	fmt.Println("In func")
	var block *Block
	gobEncoder := gob.NewDecoder(conn)
	err := gobEncoder.Decode(&block)
	if err != nil {
		log.Println(err)
	}
	if Length(chainHead) <= Length(block) {
		fmt.Println("Received new chain")
		chainHead = block
	} else {
		fmt.Println("Received old chain")

	}
	ListBlocks(chainHead)
	gobEncoder2 := gob.NewEncoder(conn)
	err2 := gobEncoder2.Encode(&chainHead)
	if err2 != nil {
		log.Println(err2)
	}
	// newchan <- chainHead
	// <-newchan
	// for i := 0; i < len(localData); i++ {
	// 	if clientsSlice[i].Types == false && localData[i].Conn != conn {
	// 		gobEncoder := gob.NewEncoder(localData[i].Conn)
	// 		//fmt.Println("BroadCheck: ", localData[i])
	// 		err1 := gobEncoder.Encode(chainHead)
	// 		fmt.Println("Broadcasting New Chain to Miners:: ", localData[i].Conn)
	// 		if err1 != nil {
	// 			log.Println("Errpr in broadcasting Chain to Miners", err1)
	// 		}
	// 	}
	// }

	//chainHead = InsertCourse(block)
	return block
}

func ReceiveEverything(conn net.Conn) { //Admin
	for {
		fmt.Println("In Recieved  func Doit", Doit)
		var stuu Combo
		gobEncoder := gob.NewDecoder(conn)
		err := gobEncoder.Decode(&stuu)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("Received Stuff chain")

		ListBlocks(stuu.ChainHead)
		fmt.Println("Received head chain")
		ListBlocks(chainHead)
		if Length(chainHead) <= Length(stuu.ChainHead) {
			fmt.Println("Received new chain")
			chainHead = stuu.ChainHead
			stuff.ChainHead = chainHead
			data := GetBlockhainArray(chainHead)
			WriteBlockchainFile(data)
		} else {
			fmt.Println("Received old chain")
		}
		ListBlocks(chainHead)

	}
	// if Doit == false {
	// 	log.Println("First Time")
	// 	gobEncoder2 := gob.NewEncoder(conn)
	// 	err2 := gobEncoder2.Encode(&stuff)
	// 	if err2 != nil {
	// 		log.Println(err2)
	// 	}
	// }

}
func ReceiveChain1(conn net.Conn) *Block {
	//<-check
	for {
		rwchan <- "sss"
		var block *Block
		gobEncoder := gob.NewDecoder(conn)
		err := gobEncoder.Decode(&block)
		if err != nil {
			log.Println(err)
		}
		fmt.Println("Received chain")
		chainHead = block
		ListBlocks(chainHead)

		//chainHead = InsertCourse(block)
	}
	//	return block
}

var j int

func broadcastPeerData() {

	for i := 0; i < len(localData); i++ {
		gobEncoder := gob.NewEncoder(localData[i].Conn)
		//fmt.Println("BroadCheck: ", localData[i])
		err1 := gobEncoder.Encode(clientsSlice)
		fmt.Println("Broadcasting PeerData:: ")
		if err1 != nil {
			log.Println("Errpr in broadcasting", err1)
		}

	}

	//	<-StepbyChan

}

func broadcastChain() {

	for i := 0; i < len(localData); i++ {
		//		fmt.Println("ss", nodesSlice[i].Types)
		gobEncoder := gob.NewEncoder(localData[i].Conn)
		//fmt.Println("BroadCheck: ", localData[i])
		err1 := gobEncoder.Encode(chainHead)
		fmt.Println("Broadcasting Chain to:: ", localData[i].Conn)
		if err1 != nil {
			log.Println("Errpr in broadcasting Chain", err1)
		}

	}

	//	<-StepbyChan

}
func broadcastEverything() {
	// stuff.ChainHead = chainHead
	// stuff.ClientsSlice = nodesSlice
	for i := 0; i < len(localData); i++ {
		//		fmt.Println("ss", nodesSlice[i].Types)
		gobEncoder := gob.NewEncoder(localData[i].Conn)
		//fmt.Println("BroadCheck: ", localData[i])
		err1 := gobEncoder.Encode(stuff)
		fmt.Println("Broadcasting Chain to:: ", localData[i].Conn)
		if err1 != nil {
			log.Println("Errpr in broadcasting Chain", err1)
		}

	}

	//	<-StepbyChan

}

func ReadPeers(conn net.Conn) []Client {
	//	for {
	//	mutex.Lock()
	var slice []Client
	gobEncoder := gob.NewDecoder(conn)
	err := gobEncoder.Decode(&slice)
	if err != nil {
		log.Println(err)
	}
	nodesSlice = slice
	fmt.Println("Read Peers: ", nodesSlice, len(nodesSlice))
	//	mutex.Unlock()
	//		check <- "d"
	//	}
	return nodesSlice
}
func ReadPeers1(conn net.Conn) []Client {
	for {
		//	mutex.Lock()

		var slice []Client
		gobEncoder := gob.NewDecoder(conn)
		err := gobEncoder.Decode(&slice)
		if err != nil {
			log.Println(err)
		}
		nodesSlice = slice
		fmt.Println("Read Peers: ", nodesSlice)

		//		<-rwchan
		//	mutex.Unlock()
		//		check <- "d"
	}
	//	return nodesSlice
}
func ReadPeersMinerChain(conn net.Conn) []Client {
	for {
		//	mutex.Lock()
		if Doit != false {
			var slice []Client
			fmt.Println("In Read Peers ggg")
			gobEncoder := gob.NewDecoder(conn)
			err := gobEncoder.Decode(&slice)
			if err != nil {
				log.Println(err, "FFF")
			}
			nodesSlice = slice
			fmt.Println("Read Peers: ", nodesSlice)
		}
		//	ReceiveChain(conn)

		//		<-rwchan
		//	mutex.Unlock()
		//		check <- "d"
	}
	//	return nodesSlice
}
func ReadPeersMinerChainEverything(conn net.Conn) { //Miner
	for {
		//	mutex.Lock()
		var stuu Combo
		fmt.Println("In Read Peers ggg")
		gobEncoder := gob.NewDecoder(conn)
		err := gobEncoder.Decode(&stuu)
		if err != nil {
			log.Println(err, "FFF")
		}
		fmt.Println("Read StuuPeers: ", stuu.ClientsSlice)
		if len(stuu.ClientsSlice) >= len(nodesSlice) {
			nodesSlice = stuu.ClientsSlice
			stuff.ClientsSlice = nodesSlice
			fmt.Println("Read Peers: ", nodesSlice)

		}
		if Length(stuu.ChainHead) >= Length(chainHead) {
			chainHead = stuu.ChainHead
			stuff.ChainHead = chainHead
			fmt.Println("Read Chain: ")
			ListBlocks(chainHead)
		}

		//	ReceiveChain(conn)

		//		<-rwchan
		//	mutex.Unlock()
		//		check <- "d"
	}
	//	return nodesSlice
}
func ReadBlockPeers(conn net.Conn) Block {
	var block Block
	gobEncoder := gob.NewDecoder(conn)
	err := gobEncoder.Decode(&block)
	if err != nil {
		log.Println(err)
	}
	return block
}

func StartListening(ListeningAddress string, node string) {
	if node == "satoshi" {
		ln, err := net.Listen("tcp", "localhost:"+ListeningAddress)
		if err != nil {
			log.Fatal(err)
			fmt.Println("Faital")
		}
		j = 0
		addchan := make(chan Client)
		block := Block{}
		chainHead = InsertCourse(block) //Genesis Block
		ReadBlockchainFile()
		stuff.ChainHead = chainHead
		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err, "Yooooo")
				continue
			}
			sendBlockchain(conn, chainHead)
			conns := Connected{
				Conn: conn,
			}

			go handleConnection(conn, addchan)
			clientsSlice = append(clientsSlice, <-addchan)
			stuff.ClientsSlice = clientsSlice
			fmt.Println("stuffCl: ", stuff.ClientsSlice)
			fmt.Println("clS: ", clientsSlice)
			localData = append(localData, conns)
			//fmt.Println("BroadCheck: ", localData[i])
			//		broadcastPeerData()
			//		broadcastEverything()

			go func() {
				for {
					time.Sleep(10 * time.Second)
					mutex.Lock()
					for i := 0; i < len(localData); i++ {
						//		fmt.Println("ss", nodesSlice[i].Types)
						gobEncoder := gob.NewEncoder(localData[i].Conn)
						//fmt.Println("BroadCheck: ", localData[i])
						err1 := gobEncoder.Encode(stuff)
						fmt.Println("Broadcasting Chain to:: ", localData[i].Conn)
						if err1 != nil {
							log.Println("Errpr in broadcasting Chain", err1)
						}

					}
					mutex.Unlock()

				}
			}()
			//	chainHead = a2.InsertBlock("", "", "Satoshi", 0, chainHead)
			// var block Block
			// gobEncoder := gob.NewDecoder(conn)
			// err2 := gobEncoder.Decode(&block)
			// if err2 != nil {
			// 	log.Println(err2)
			// }
			//	go ReceiveMinerChain(conn)
			go ReceiveEverything(conn)

			//chainHead = InsertCourse(block)
			ListBlocks(chainHead)

		}
	} else if node == "others" {
		ln, err := net.Listen("tcp", "localhost:"+ListeningAddress)
		if err != nil {
			log.Fatal(err)
			fmt.Println("Faital")
		}

		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err, "Yooooo")
				continue
			}
			go handlePeer(conn)
			nodesSlice = append(nodesSlice, <-minechan)

		}

	} else { //miner
		ln, err := net.Listen("tcp", "localhost:"+ListeningAddress)
		if err != nil {
			log.Fatal(err)
			fmt.Println("Faital")
		}

		for {
			conn, err := ln.Accept()
			if err != nil {
				log.Println(err, "Yooooo")
				continue
			}
			fmt.Println("COnnedted")
			testConn = conn
			conns := Connected{
				Conn: conn,
			}
			localData = append(localData, conns)

			go handlePeer(conn)

			Minedblock = <-blockchan

		}
	}

}

var testConn net.Conn

// Chi HTTP Services //

func setHandler(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles("../Website/blockchain.html") //parse the html file homepage.html
	if err != nil {                                             // if there is an error
		log.Print("template parsing error: ", err) // log it
	}

	err = t.Execute(w, nil) //execute the template and pass it the HomePageVars struct to fill in the gaps
	if err != nil {         // if there is an error
		log.Print("template executing error: ", err) //log it
	}
}

var MinerConn net.Conn
var Mined bool

func getHandler(w http.ResponseWriter, r *http.Request) {
	r.ParseForm()
	cCode := r.Form.Get("courseCode")
	cName := r.Form.Get("courseName")
	cGrade := r.Form.Get("courseGrade")
	cEmail := r.Form.Get("courseEmail")

	a, err := strconv.Atoi(r.FormValue("courseCHrs"))
	if err != nil {
	}
	cCHrs := a

	AddCourse := Course{
		Code:        cCode,
		Name:        cName,
		CreditHours: cCHrs,
		Grade:       cGrade,
	}

	MyBlock := Block{
		Course: AddCourse,
		Email:  cEmail,
	}

	//chainHead = InsertCourse(MyBlock)

	// gobEncoder := gob.NewEncoder(Satoshiconn)
	// err2 := gobEncoder.Encode(MyBlock)
	// if err2 != nil {
	// 	log.Println("In Write Chain: ", err2)
	// }

	ListBlocks(chainHead)

	tempHead := chainHead
	viewTheBlock := new(ListTheBlock)
	tempCourse := []Course{}
	tempBlockNo := []int{}
	tempCurrHash := []string{}
	tempPrevHash := []string{}
	tempEmail := []string{}
	for tempHead != nil {
		tempCourse = append(tempCourse, tempHead.Course)
		tempBlockNo = append(tempBlockNo, tempHead.BlockNo)
		tempCurrHash = append(tempCurrHash, tempHead.CurrentHash)
		tempPrevHash = append(tempPrevHash, tempHead.PrevHash)
		tempEmail = append(tempEmail, tempHead.Email)
		viewTheBlock = &ListTheBlock{
			Course:      tempCourse,
			BlockNo:     tempBlockNo,
			CurrentHash: tempCurrHash,
			PrevHash:    tempPrevHash,
			Email:       tempEmail,
		}
		tempHead = tempHead.PrevPointer
		fmt.Println(viewTheBlock.Course)
		fmt.Println(viewTheBlock.BlockNo)
		fmt.Println(viewTheBlock.CurrentHash)
		fmt.Println(viewTheBlock.PrevHash)
		fmt.Println(viewTheBlock.Email)
	}
	// generate page by passing page variables into template
	t, err := template.ParseFiles("../Website/blockchain.html") //parse the html file homepage.html
	if err != nil {                                             // if there is an error
		log.Print("template parsing error: ", err) // log it
	}

	err = t.Execute(w, viewTheBlock) //execute the template and pass it the HomePageVars struct to fill in the gaps
	if err != nil {                  // if there is an error
		log.Print("template executing error: ", err) //log it
	}
	//	fmt.Println("FFFFFFFFFF", len(nodesSlice))
	for i := 0; i < len(nodesSlice); i++ {
		//	fmt.Println("dddddddddddddddddddddddddddddddddddddddddddddddddddddddd")
		if nodesSlice[i].Mail == MyBlock.Email {
			conn, err := net.Dial("tcp", "localhost:"+nodesSlice[i].ListeningAddress)
			if err != nil {
				log.Fatal(err)
			}
			MinerConn = conn
			gobEncoder := gob.NewEncoder(conn)
			fmt.Println("blok:ahsh: ", CalculateHash(&MyBlock))
			err2 := gobEncoder.Encode(MyBlock)
			if err2 != nil {
				log.Println("In Write Chain: ", err2)
			}
			m := gomail.NewMessage()

			// Set E-Mail sender
			m.SetHeader("From", "mohtasimasadabbasi@gmail.com")

			// Set E-Mail receivers
			m.SetHeader("To", MyBlock.Email)

			// Set E-Mail subject
			m.SetHeader("Subject", "Verification Content")

			// Set E-Mail body. You can set plain text or html with text/html
			m.SetBody("text/plain", "Course Name: "+MyBlock.Course.Name+"  Course Code: "+MyBlock.Course.Code+"  Course Grade: "+MyBlock.Course.Grade+"\n"+"Click here to verify this content: "+"localhost:"+"3335"+"/mine/"+CalculateHash(&MyBlock))

			// Settings for SMTP server
			d := gomail.NewDialer("smtp.gmail.com", 587, "mohtasimasadabbasi@gmail.com", "mohtasim70")

			// This is only needed when SSL/TLS certificate is not valid on server.
			// In production this should be set to false.
			d.TLSConfig = &tls.Config{InsecureSkipVerify: true}

			// Now send E-Mail
			if err := d.DialAndSend(m); err != nil {
				fmt.Println(err, "mailerr")
				panic(err)
			}
			Mined = true
			fmt.Println("Email Sent", Mined, nodesSlice[i].ListeningAddress)

			break
		}
	}

	// gobEncoder := gob.NewEncoder(Satoshiconn)
	// err2 := gobEncoder.Encode(MyBlock)
	// if err2 != nil {
	// 	log.Println("In Write Chain: ", err2)
	// }

}

func showBlocksHandler(w http.ResponseWriter, r *http.Request) {
	tempHead := chainHead
	viewTheBlock := new(ListTheBlock)
	tempCourse := []Course{}
	tempBlockNo := []int{}
	tempCurrHash := []string{}
	tempPrevHash := []string{}
	tempStatus := []string{}
	for tempHead != nil {
		tempCourse = append(tempCourse, tempHead.Course)
		tempBlockNo = append(tempBlockNo, tempHead.BlockNo)
		tempCurrHash = append(tempCurrHash, tempHead.CurrentHash)
		tempPrevHash = append(tempPrevHash, tempHead.PrevHash)
		tempStatus = append(tempStatus, tempHead.Status)

		viewTheBlock = &ListTheBlock{
			Course:      tempCourse,
			BlockNo:     tempBlockNo,
			CurrentHash: tempCurrHash,
			PrevHash:    tempPrevHash,
			Status:      tempStatus,
		}
		tempHead = tempHead.PrevPointer
		fmt.Println(viewTheBlock.Course)
		fmt.Println(viewTheBlock.BlockNo)
		fmt.Println(viewTheBlock.CurrentHash)
		fmt.Println(viewTheBlock.PrevHash)
		fmt.Println(viewTheBlock.Status)
	}
	// generate page by passing page variables into template
	t, err := template.ParseFiles("../Website/viewBlocks.html") //parse the html file homepage.html
	if err != nil {                                             // if there is an error
		log.Print("template parsing error: ", err) // log it
	}

	err = t.Execute(w, viewTheBlock) //execute the template and pass it the HomePageVars struct to fill in the gaps
	if err != nil {                  // if there is an error
		log.Print("template executing error: ", err) //log it
	}
}

var check = make(chan string)

var Doit bool

func Mineblock(w http.ResponseWriter, r *http.Request) {
	// Doit = false
	// gobEncoder3 := gob.NewEncoder(Satoshiconn)
	// err3 := gobEncoder3.Encode(stuff)
	// if err3 != nil {
	// 	log.Println("In Write Chain..........: ", err3)
	// }
	fmt.Println("In Mine Block")

	// var block Combo
	// Decoder := gob.NewDecoder(Satoshiconn)
	// err8 := Decoder.Decode(&block)
	// if err8 != nil {
	// 	log.Println(err8, "Errr while mining.......")
	// }
	// fmt.Println("Decoding Doneee............")
	//
	// chainHead = block.ChainHead
	// stuff.ChainHead = chainHead
	// fmt.Println("Handler executed.........")
	Doit = true

	params := mux.Vars(r)
	mineHash := params["hash"]
	fmt.Println(mineHash)
	chainHead = InsertCourse(Minedblock)
	stuff.ChainHead = chainHead
	fmt.Println("In Mining")
	ListBlocks(chainHead)

	gobEncoder := gob.NewEncoder(Satoshiconn)
	err2 := gobEncoder.Encode(stuff)
	if err2 != nil {
		log.Println("InError Write Chain: ", err2)
	}
	log.Println("Sent to Satoshi: ")

	gobEncoder2 := gob.NewEncoder(testConn)
	//fmt.Println("BroadCheck: ", localData[i])
	err1 := gobEncoder2.Encode(stuff)
	fmt.Println("Bro Chain sent to peer:: ", testConn)
	if err1 != nil {
		log.Println("Errpr in brosti Chain", err1)
	}

	//	broadcastChain()

}

var broadcast = make(chan []Block) // broadcast channel

//func HandleConnections(w http.ResponseWriter, r *http.Request) {

//	ws, err := upgrader.Upgrade(w, r, nil)
//	if err != nil {
//		log.Fatal(err)
//		fmt.Println("Error in ebss")
//	}

// make sure we close the connection when the function returns
//	defer ws.Close()

// register our new client
//	nodes[ws] = true

//	for {
// Read in a new message as JSON and map it to a Message object
//		var course Course
//		err := json.NewDecoder(r.Body).Decode(&course)
//		if err != nil {
//			panic(err)
//		}
// err := ws.ReadJSON(&course)
//		chainHead = InsertCourse1(course, chainHead)
// if err != nil {
// 	log.Printf("error: %v", err)
// 	//	delete(nodes, ws)
// 	break
// }

// Send the newly received message to the broadcast channel
//		broadcast <- getCourse(chainHead)
//	}

//}

// Clients Web Server //

func RegisterHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t, err := template.ParseFiles("../Website/register.html") //parse the html file homepage.html
		if err != nil {                                           // if there is an error
			log.Print("template parsing error: ", err) // log it
		}

		err = t.Execute(w, nil) //execute the template and pass it the HomePageVars struct to fill in the gaps
		if err != nil {         // if there is an error
			log.Print("template executing error: ", err) //log it
		}
	}
	if r.Method == "POST" {
		r.ParseForm()
		userName := r.Form.Get("username")
		fName := r.Form.Get("firstname")
		lName := r.Form.Get("lastname")
		password := r.Form.Get("password")
		emailAddr := r.Form.Get("email")
		w.Header().Set("Content-Type", "application/json")
		user := model.User{
			Username:  userName,
			FirstName: fName,
			LastName:  lName,
			Password:  password,
			Email:     emailAddr,
		}

		collection, err := db.GetDBCollection()

		var result model.User
		err = collection.FindOne(context.TODO(), bson.D{primitive.E{Key: "username", Value: user.Username}}).Decode(&result)

		if err != nil {
			if err.Error() == "mongo: no documents in result" {
				hash, err := bcrypt.GenerateFromPassword([]byte(user.Password), 5)
				if err != nil { // if there is an error
					log.Print("Error ", err) //log it
				}
				user.Password = string(hash)

				_, err = collection.InsertOne(context.TODO(), user)
				if err != nil { // if there is an error
					log.Print("Error ", err) //log it
				}
			}
		}
		http.Redirect(w, r, urlLogin+"/login", http.StatusSeeOther)
	}
}

func LoginHandler(w http.ResponseWriter, r *http.Request) {

	if r.Method == "GET" {
		t, err := template.ParseFiles("../Website/login.html") //parse the html file homepage.html
		if err != nil {                                        // if there is an error
			log.Print("template parsing error: ", err) // log it
		}

		err = t.Execute(w, nil) //execute the template and pass it the HomePageVars struct to fill in the gaps
		if err != nil {         // if there is an error
			log.Print("template executing error: ", err) //log it
		}
	}
	if r.Method == "POST" {
		r.ParseForm()
		userName := r.Form.Get("username")
		password := r.Form.Get("password")
		w.Header().Set("Content-Type", "application/json")
		user := model.User{
			Username: userName,
			Password: password,
		}

		collection, err := db.GetDBCollection()

		if err != nil {
			log.Fatal(err)
		}
		var result model.User
		var res model.ResponseResult

		err = collection.FindOne(context.TODO(), bson.D{primitive.E{Key: "username", Value: user.Username}}).Decode(&result)

		if err != nil {
			res.Error = "Invalid username"
			json.NewEncoder(w).Encode(res)
			return
		}

		err = bcrypt.CompareHashAndPassword([]byte(result.Password), []byte(user.Password))

		if err != nil {
			res.Error = "Invalid password"
			json.NewEncoder(w).Encode(res)
			return
		}

		token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
			"username":  result.Username,
			"firstname": result.FirstName,
			"lastname":  result.LastName,
			"email":     result.Email,
		})

		tokenString, err = token.SignedString([]byte("secret"))

		if err != nil {
			res.Error = "Error while generating token,Try again"
			json.NewEncoder(w).Encode(res)
			return
		}

		result.Token = tokenString
		result.Password = ""
		http.Redirect(w, r, urlLogin+"/dashboard", http.StatusSeeOther)
		json.NewEncoder(w).Encode(result)
	}
}

func ProfileHandler(w http.ResponseWriter, r *http.Request) {
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		// Don't forget to validate the alg is what you expect:
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			fmt.Println(" ---- Access Denied ----")
			return nil, fmt.Errorf("Unexpected signing method")
		}
		return []byte("secret"), nil
	})
	if token == nil {
		fmt.Println(" ---- Access Denied ----")
		http.Redirect(w, r, urlLogin+"/login", http.StatusSeeOther)
		return
	}
	var result model.User
	var res model.ResponseResult
	if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
		result.Username = claims["username"].(string)
		result.FirstName = claims["firstname"].(string)
		result.LastName = claims["lastname"].(string)
		result.Email = claims["email"].(string)
		user := model.User{
			Username:  result.Username,
			FirstName: result.FirstName,
			LastName:  result.LastName,
			Email:     result.Email,
			Password:  "",
		}
		t, err := template.ParseFiles("../Website/index.html") //parse the html file homepage.html
		if err != nil {                                        // if there is an error
			log.Print("template parsing error: ", err) // log it
		}

		err = t.Execute(w, user) //execute the template and pass it the HomePageVars struct to fill in the gaps
		if err != nil {          // if there is an error
			log.Print("template executing error: ", err) //log it
		}
	} else {
		res.Error = err.Error()
		json.NewEncoder(w).Encode(res)
		return
	}

}

func LogoutHandler(w http.ResponseWriter, r *http.Request) {
	tokenString = ""
	http.Redirect(w, r, urlLogin+"/login", http.StatusSeeOther)
}

func RunWebServer(port string) {
	// router := mux.NewRouter().StrictSlash(true)
	// router.HandleFunc("/ws", server.HandleConnections)
	// router.HandleFunc("/api/block", server.GetAllBlock).Methods("GET", "OPTIONS")
	// router.HandleFunc("/api/block", server.CreateBlock).Methods("POST", "OPTIONS")
	// router.HandleFunc("/api/task", server.CreateTask).Methods("POST", "OPTIONS")

	r := mux.NewRouter()
	r.HandleFunc("/", setHandler).Methods("GET")
	r.HandleFunc("/blockInsert", getHandler).Methods("POST")
	//r.HandleFunc("/ws", HandleConnections)
	r.HandleFunc("/register", RegisterHandler)
	r.HandleFunc("/login", LoginHandler)
	r.HandleFunc("/logout", LogoutHandler)
	r.HandleFunc("/dashboard", ProfileHandler).
		Methods("GET")
	r.PathPrefix("/css/").Handler(http.StripPrefix("/css/", http.FileServer(http.Dir("../Website/css"))))
	r.PathPrefix("/js/").Handler(http.StripPrefix("/js/", http.FileServer(http.Dir("../Website/js"))))
	r.PathPrefix("/vendor/").Handler(http.StripPrefix("/vendor/", http.FileServer(http.Dir("../Website/vendor"))))
	r.PathPrefix("/images/").Handler(http.StripPrefix("/images/", http.FileServer(http.Dir("../Website/images"))))
	r.PathPrefix("/fonts/").Handler(http.StripPrefix("/fonts/", http.FileServer(http.Dir("../Website/fonts"))))

	urlLogin = "http://localhost:" + port
	http.ListenAndServe("localhost:"+port, r)

}

func RunWebServerMiner(port string) {

	r := mux.NewRouter()
	r.HandleFunc("/mine/{hash}", Mineblock).Methods("GET")

	// r.Method("POST", "/blockInsert", Handler(getHandler))
	//r.HandleFunc("/ws", HandleConnections)
	http.ListenAndServe("localhost:"+port, r)

}

// Satoshi Web Server //

func RunWebServerSatoshi() {

	r := mux.NewRouter()
	r.HandleFunc("/showBlocks", showBlocksHandler).Methods("GET")
	//r.HandleFunc("/ws", HandleConnections)

	http.ListenAndServe("localhost"+":3333", r)

}

//func BroadcastMessages() {
//	for {
//	// Grab the next message from the broadcast channel
//	msg := <-broadcast
//	fmt.Println("In broadcast: ", msg)
// Send it out to every client that is currently connected
//	for client := range nodes {
//		err := client.WriteJSON(msg)
//		if err != nil {
//		log.Printf("error: %v", err)
//		client.Close()
//		delete(nodes, client)
//		}
//	}
//	}
//}

// ---- //

func main() {
	// ln, err := net.Listen("tcp", "localhost:6003")
	// if err != nil {
	//
	// 	log.Fatal(err, ln)
	//
	// }
	//go RunWebServer()

	//go BroadcastMessages()

	//select {}

	// conn, err := net.Dial("tcp", "localhost:3333")
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Println("ss", conn)
	// for {
	// 	conn, err := ln.Accept()
	// 	if err != nil {
	// 		log.Println(err)
	// 		continue
	// 	}
	// 	go sendBlockchain(conn, chainHead)
	// }

}
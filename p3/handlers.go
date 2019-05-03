package p3

import (
	"../p1"
	"../p2"
	"./data"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"io/ioutil"
	"log"
	"math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

const charset = "abcdefghijklmnopqrstuvwxyz" +
	"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

var TA_SERVER = "http://localhost:6688"
var REGISTER_SERVER = TA_SERVER + "/peer"
var BC_DOWNLOAD_SERVER = TA_SERVER + "/upload"
var SELF_ADDR = "http://localhost:" + os.Args[1]
var FIRST_NODE_ADDR = "http://localhost:6686"
var JSON_BLOCKCHAIN = "[{\"hash\": \"3ff3b4efe9177f705550231079c2459ba54a22d340a517e84ec5261a0d74ca48\", \"timeStamp\": 1234567890, \"height\": 1, \"parentHash\": \"genesis\", \"size\": 1174, \"mpt\": {\"hello\": \"world\", \"charles\": \"ge\"}}, {\"hash\": \"24cf2c336f02ccd526a03683b522bfca8c3c19aed8a1bed1bbc23c33cd8d1159\", \"timeStamp\": 1234567890, \"height\": 2, \"parentHash\": \"3ff3b4efe9177f705550231079c2459ba54a22d340a517e84ec5261a0d74ca48\", \"size\": 1231, \"mpt\": {\"hello\": \"world\", \"charles\": \"ge\"}}]"

var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool
var mpt p1.MerklePatriciaTrie

type StoreFileInfo struct {
	CiphertextData []byte
	Signature      []byte
	PublicKey      string
	DataHash       string
}

type StoreFileInfo_encoded struct {
	CiphertextData []byte `json:"CiphertextData"`
	Signature      []byte `json:"Signature"`
	PublicKey      string `json:"PublicKey"`
	DataHash       string `json:"DataHash"`
}

type file_retrieval_resp struct {
	TXfee int32                 `json:"TXfee"`
	Data  StoreFileInfo_encoded `json:"Data"`
}

type store_resp struct {
	BlockHash   string `json:"BlockHash"`
	BlockHeight int32  `json:"BlockHeight"`
}

//Initialize all variables for the server
func init() {
	SBC = data.NewBlockChain()
	id, _ := strconv.ParseInt(os.Args[1], 10, 32)
	Peers = data.NewPeerList( /*Register()*/ int32(id), 32) // Uses port number as ID since TA server is down
	ifStarted = false
	mpt.Initial()
}

// Register ID, download BlockChain, start HeartBeat
func Start(w http.ResponseWriter, r *http.Request) {
	// The first Node will read two blocks from a Json string and add it to the BlockChain
	if os.Args[2] == "yes" {
		fmt.Println("Create BlockChain from json ")
		SBC.UpdateEntireBlockChain(JSON_BLOCKCHAIN)
	} else {
		fmt.Println("Download BlockChain from first node..")
		Download()
	}
	ifStarted = true
	go StartHeartBeat()
	//go StartTryingNonces()
	fmt.Fprintf(w, "Node started")
}

// Display peerList and sbc
func Show(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

// Register to TA's server with get request and return the userId
func Register() int32 {
	response, err := http.Get(REGISTER_SERVER)
	if err != nil {
		fmt.Println("Cant reach TA server")
		os.Exit(1)
	}

	defer response.Body.Close()
	contents, err := ioutil.ReadAll(response.Body)

	if err != nil {
		os.Exit(1)
	}
	val, _ := strconv.ParseInt(string(contents), 10, 32)

	return int32(val)
}

// Download BlockChain from TA server
func Download() {
	response, err := http.Get(FIRST_NODE_ADDR + "/upload")
	if err != nil {
		fmt.Println("Cant reach TA server")
		os.Exit(1)
	}
	defer response.Body.Close()

	responseData, err := ioutil.ReadAll(response.Body)
	if err != nil {
		log.Fatal(err)
	}
	responseString := string(responseData)
	SBC.UpdateEntireBlockChain(responseString)
}

// Upload BlockChain to whoever called this method, return jsonStr
func Upload(w http.ResponseWriter, r *http.Request) {
	blockChainJson, err := SBC.BlockChainToJson()
	if err != nil {
		fmt.Println(err, "Upload")
	}

	fmt.Fprint(w, blockChainJson)
}

// Upload a block to whoever called this method, return jsonStr
// HTTP status 500 if there is an error
// HTTP status 204 if block does not exist
// Else return json string of block
func UploadBlock(w http.ResponseWriter, r *http.Request) {
	//Read from request get height and hash
	s := strings.Split(r.URL.Path, "/")
	val, err := strconv.ParseInt(s[2], 10, 32)

	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
	}

	height := int32(val)
	hash := s[3]
	block, found := SBC.GetBlock(height, hash)

	if found {
		jsonBlock := block.EncodeToJSON()
		fmt.Fprint(w, jsonBlock)
	}

	w.WriteHeader(http.StatusNoContent)

}

// Received a heartbeat
func HeartBeatReceive(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}

	rData := data.HeartBeatData{}
	err = json.Unmarshal(body, &rData)
	// Check that its not a forwarded heartbeat from yourself, and add sender to PeerList
	Peers.InjectPeerMapJson(rData.PeerMapJson, SELF_ADDR)
	if rData.Addr != SELF_ADDR {
		//TODO does not need to have if around new block but good for print
		Peers.Add(rData.Addr, rData.Id)
		// Add all the peers from senders PeerList
		if rData.IfNewBlock { // If HeartBeat contatins a new Block
			fmt.Println("Received new block form peer")
			newBlock := p2.DecodeFromJson(rData.BlockJson)
			if CheckPowForNewBlock(newBlock) {
				fmt.Println("POW is correct")
				if !SBC.CheckParentHash(newBlock) { // Check if you have the parent block if not ask for it from a peer
					AskForBlock(newBlock.Header.Height-1, newBlock.Header.ParentHash)
				}
				// Add new block to the chain
				SBC.Insert(p2.DecodeFromJson(rData.BlockJson))
			}
		}
		// Check if heartbeat should be forwarded
		if rData.Hops > 0 {
			hops := rData.Hops - 1
			forwardData := data.HeartBeatData{rData.IfNewBlock, rData.Id, rData.BlockJson, rData.PeerMapJson, hops, rData.Addr}
			ForwardHeartBeat(forwardData)
		}
	}
	w.WriteHeader(http.StatusOK)
}
func CheckPowForNewBlock(block p2.Block) bool {
	hashStr := block.Header.ParentHash + block.Header.Nonce + block.Value.Root
	sum := sha3.Sum256([]byte(hashStr))
	powString := hex.EncodeToString(sum[:])
	return checkPowResult(powString)
}

// Ask another server to return a block of certain height and hash
func AskForBlock(height int32, hash string) {
	Peers.Rebalance()
	for key, _ := range Peers.Copy() {
		url := key + "/block/" + string(height) + "/" + hash
		response, err := http.Get(url)
		if err != nil {
			fmt.Println("Can't ask for block error...")
		}

		if response.StatusCode == 200 {
			fmt.Println("Received missing parent block with height: " + string(height))
			responseData, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			} else {
				responseString := string(responseData)
				receivedBlock := p2.DecodeFromJson(responseString)
				SBC.Insert(receivedBlock)

				//Recursively ask for parent block of received ParentBlock
				if !SBC.CheckParentHash(receivedBlock) {
					AskForBlock(receivedBlock.Header.Height-1, receivedBlock.Header.ParentHash)
				}
				break
			}
		}
	}
}

// Forward a heartbeat to all your peers.
func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	Peers.Rebalance()
	for key, _ := range Peers.Copy() {
		url := key + "/heartbeat/receive"
		jsonString, _ := json.Marshal(heartBeatData)
		_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonString))
		if err != nil {
			fmt.Print("Unable to forward heartbeat")
		}
	}
}

// Send a heartbeat to the first node, to be added to its PeerList
func sendInitHeartBeat() {
	if SELF_ADDR != FIRST_NODE_ADDR {
		url := FIRST_NODE_ADDR + "/heartbeat/receive"
		peerMapJson, _ := Peers.PeerMapToJson()
		jsonStr := data.PrepareHeartBeatData(Peers.GetSelfId(), peerMapJson, SELF_ADDR)
		jsonString, _ := json.Marshal(jsonStr)
		_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonString))
		if err != nil {
			fmt.Print("Error sending init message")
		}
	}
}

// Send a heartbeat to every peer node every 10 second
func StartHeartBeat() {
	duration := time.Duration(10) * time.Second // Pause for 10 seconds

	sendInitHeartBeat()
	for true {
		time.Sleep(duration)
		Peers.Rebalance()
		peerMapJson, _ := Peers.PeerMapToJson()
		jsonStr := data.PrepareHeartBeatData(Peers.GetSelfId(), peerMapJson, SELF_ADDR)
		jsonString, _ := json.Marshal(jsonStr)

		for key, _ := range Peers.Copy() {
			url := key + "/heartbeat/receive"
			_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonString))
			if err != nil {
				fmt.Print("Error sending init message")
			}
		}
	}
}

func StartTryingNonces() {
	duration := time.Duration(4) * time.Second // Pause for 10 seconds
	time.Sleep(duration)
	fmt.Println("Start nonce")
	//Get latest block
	latestBlockList := SBC.GetLatestBlocks()
	blockListLength := len(latestBlockList) - 1
	randomVal := 0
	currentChainLength := SBC.GetChainLength()

	if blockListLength > 0 {
		randomVal = rand.Intn(int(blockListLength)-0) + 0
	}

	parentBlock := latestBlockList[randomVal]
	powString := "9999999999"
	correctNonce := ""
	ifNewBlock := true
	fmt.Println("find new block for : " + parentBlock.GetHash())

	for !checkPowResult(powString) {
		if currentChainLength != SBC.GetChainLength() {
			ifNewBlock = false
			break
		} else {
			nonce, err := randomHex()
			if err == nil {
				hashStr := parentBlock.Header.Hash + nonce + mpt.Root
				sum := sha3.Sum256([]byte(hashStr))
				powString = hex.EncodeToString(sum[:])
				correctNonce = nonce
			}
		}
	}

	if ifNewBlock {
		fmt.Println("Found pow send to all peers")
		fmt.Println(powString)
		newBlock := SBC.GenBlock(mpt, correctNonce, parentBlock.Header.Hash)
		SBC.Insert(newBlock)
		newBlockJson := newBlock.EncodeToJSON()
		Peers.Rebalance() // Make sure to only send the correct amount of peers
		peerMapJson, _ := Peers.PeerMapToJson()
		heartBeatData := data.NewHeartBeatData(ifNewBlock, Peers.GetSelfId(), newBlockJson, peerMapJson, SELF_ADDR)
		ForwardHeartBeat(heartBeatData)
	} else {
		fmt.Println("Someone solved pow before me")
	}
}

// Check if the first 7 chars in the POW are 0's
func checkPowResult(pow string) bool {
	first7 := pow[0:3]
	for _, char := range first7 {
		if char != '0' {
			return false
		}
	}
	return true
}

// Generate a random hex string with 16 char
// src: https://sosedoff.com/2014/12/15/generate-random-hex-string-in-go.html
func randomHex() (string, error) {
	bytes := make([]byte, 8)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

// Generate a random string of 10 bytes
// src: https://www.calhoun.io/creating-random-strings-in-go/
func generateRandomString() string {
	var seededRand *rand.Rand = rand.New(
		rand.NewSource(time.Now().UnixNano()))
	b := make([]byte, 10)

	for i := range b {
		b[i] = charset[seededRand.Intn(len(charset))]
	}
	return string(b)
}

// Generate a MPT and insert 1-15 random key, value pairs
func GenerateRandomMPT() p1.MerklePatriciaTrie {
	insertNumber := rand.Intn(15-1) + 1
	mpt := p1.NewMPT()
	for i := 0; i < insertNumber; i++ {
		key := generateRandomString()
		value := generateRandomString()
		mpt.Insert(key, value)
	}
	return mpt
}

func Canonical(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, SBC.ShowCanonical())
}

func Store(w http.ResponseWriter, r *http.Request) {
	fmt.Println("--- Got store requests ---")
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		panic(err)
	}
	newData := StoreFileInfo{}
	data := StoreFileInfo_encoded{}

	json.Unmarshal([]byte(body), &data)
	newData.CiphertextData = data.CiphertextData
	newData.Signature = data.Signature
	newData.PublicKey = data.PublicKey
	newData.DataHash = data.DataHash
	mpt.Insert(newData.DataHash, string(body))

	blockHash, blockHeight := generateNewBlock(newData.DataHash, string(body))
	jsonResp := store_resp{blockHash, blockHeight}
	result, _ := json.Marshal(jsonResp)
	fmt.Fprint(w, string(result))
}

func Retrieve(w http.ResponseWriter, r *http.Request) {
	s := strings.Split(r.URL.Path, "/")
	//val, err :=

	fileHash := s[2]
	blockHeight, _ := strconv.ParseInt(s[3], 10, 32)
	blockHash := s[4]
	foundBlock, found := SBC.GetBlock(int32(blockHeight), blockHash)
	if !found {
		fmt.Println("Cant find block")

		w.WriteHeader(http.StatusNoContent)
		return
	}
	fmt.Println("Found block")
	//TODO change resp status
	data, err := foundBlock.Value.Get(fileHash)
	if err != nil {
		fmt.Println("Cant find data in mpt")

		w.WriteHeader(http.StatusNoContent)
		return
	}

	minerResp := StoreFileInfo_encoded{}
	json.Unmarshal([]byte(data), &minerResp)

	b, err := json.Marshal(file_retrieval_resp{10, minerResp})

	fmt.Fprint(w, string(b))
}

func generateNewBlock(key string, value string) (string, int32) {
	latestBlockList := SBC.GetLatestBlocks()
	blockListLength := len(latestBlockList) - 1
	randomVal := 0

	if blockListLength > 0 {
		randomVal = rand.Intn(int(blockListLength)-0) + 0
	}

	parentBlock := latestBlockList[randomVal]
	powString := "9999999999"
	correctNonce := ""
	fmt.Println("find new block for : " + parentBlock.GetHash())

	for !checkPowResult(powString) {
		nonce, err := randomHex()
		if err == nil {
			hashStr := parentBlock.Header.Hash + nonce + mpt.Root
			sum := sha3.Sum256([]byte(hashStr))
			powString = hex.EncodeToString(sum[:])
			correctNonce = nonce
		}

	}

	fmt.Println("Found pow send to all peers")
	fmt.Println(powString)
	newBlock := SBC.GenBlock(mpt, correctNonce, parentBlock.Header.Hash)
	SBC.Insert(newBlock)
	newBlockJson := newBlock.EncodeToJSON()
	Peers.Rebalance() // Make sure to only send the correct amount of peers
	peerMapJson, _ := Peers.PeerMapToJson()
	heartBeatData := data.NewHeartBeatData(true, Peers.GetSelfId(), newBlockJson, peerMapJson, SELF_ADDR)
	ForwardHeartBeat(heartBeatData)
	mpt.Initial()

	return newBlock.GetHash(), newBlock.GetHeight()
}

package p3

import (
	"../p2"
	"./data"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var TA_SERVER = "http://localhost:6688"
var REGISTER_SERVER = TA_SERVER + "/peer"
var BC_DOWNLOAD_SERVER = TA_SERVER + "/upload"
var SELF_ADDR = "http://localhost:" + os.Args[1]
var FIRST_NODE_ADDR = "http://localhost:6686"
var JSON_BLOCKCHAIN = "[{\"hash\": \"3ff3b4efe9177f705550231079c2459ba54a22d340a517e84ec5261a0d74ca48\", \"timeStamp\": 1234567890, \"height\": 1, \"parentHash\": \"genesis\", \"size\": 1174, \"mpt\": {\"hello\": \"world\", \"charles\": \"ge\"}}, {\"hash\": \"24cf2c336f02ccd526a03683b522bfca8c3c19aed8a1bed1bbc23c33cd8d1159\", \"timeStamp\": 1234567890, \"height\": 2, \"parentHash\": \"3ff3b4efe9177f705550231079c2459ba54a22d340a517e84ec5261a0d74ca48\", \"size\": 1231, \"mpt\": {\"hello\": \"world\", \"charles\": \"ge\"}}]"

var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool

func init() {
	SBC = data.NewBlockChain()
	Peers = data.NewPeerList( /*Register()*/ 5, 32)
	ifStarted = false
}

// Register ID, download BlockChain, start HeartBeat
func Start(w http.ResponseWriter, r *http.Request) {
	if os.Args[2] == "yes" {
		fmt.Println("Create blockchain from json ")
		SBC.UpdateEntireBlockChain(JSON_BLOCKCHAIN)
	} else {
		fmt.Println("Download blockchain from first node..")
		Download()
	}

	ifStarted = true
	StartHeartBeat()
	w.WriteHeader(http.StatusOK)

}

// Display peerList and sbc
func Show(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%s\n%s", Peers.Show(), SBC.Show())
}

// Register to TA's server, get an ID
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

// Download blockchain from TA server
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

// Upload blockchain to whoever called this method, return jsonStr
func Upload(w http.ResponseWriter, r *http.Request) {
	blockChainJson, err := SBC.BlockChainToJson()
	if err != nil {
		fmt.Println(err, "Upload")
	}

	fmt.Fprint(w, blockChainJson)
}

// Upload a block to whoever called this method, return jsonStr
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
	if rData.Addr != SELF_ADDR {
		Peers.Add(rData.Addr, rData.Id)
	}

	Peers.InjectPeerMapJson(rData.PeerMapJson, SELF_ADDR)
	if rData.IfNewBlock {
		newBlock := p2.DecodeFromJson(rData.BlockJson)
		if !SBC.CheckParentHash(newBlock) {
			AskForBlock(newBlock.Header.Height-1, newBlock.Header.ParentHash)
		}
		SBC.Insert(p2.DecodeFromJson(rData.BlockJson))
	}

	if rData.Hops > 0 {
		hops := rData.Hops - 1
		forwardData := data.HeartBeatData{rData.IfNewBlock, rData.Id, rData.BlockJson, rData.PeerMapJson, hops, rData.Addr}
		ForwardHeartBeat(forwardData)
	}
	w.WriteHeader(http.StatusOK)
}

// Ask another server to return a block of certain height and hash
func AskForBlock(height int32, hash string) {
	Peers.Rebalance()

	for key, _ := range Peers.Copy() {
		url := key + "/block/" + string(height) + "/" + hash
		response, err := http.Get(url)
		if err != nil {
			fmt.Println("Can't ask for block error...")
			os.Exit(1)
		}

		if response.StatusCode == 200 {
			fmt.Println("Received missing parent block with height: " + string(height))
			responseData, err := ioutil.ReadAll(response.Body)
			if err != nil {
				log.Fatal(err)
			}

			responseString := string(responseData)
			receivedBlock := p2.DecodeFromJson(responseString)
			SBC.Insert(receivedBlock)
			break
		}
	}
}

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

func sendInitHeartBeat() {
	if SELF_ADDR != FIRST_NODE_ADDR {
		url := FIRST_NODE_ADDR + "/heartbeat/receive"
		peerMapJson, _ := Peers.PeerMapToJson()
		jsonStr := data.HeartBeatData{false, Peers.GetSelfId(), "", peerMapJson, 0, SELF_ADDR}
		jsonString, _ := json.Marshal(jsonStr)
		_, err := http.Post(url, "application/json", bytes.NewBuffer(jsonString))
		if err != nil {
			fmt.Print("Error sending init message")
		}
	}
}

func StartHeartBeat() {
	sendInitHeartBeat()
	duration := time.Duration(10) * time.Second // Pause for 10 seconds
	running := true
	for running {
		time.Sleep(duration)
		fmt.Println("Sending heartbeat")
		Peers.Rebalance()
		peerMapJson, _ := Peers.PeerMapToJson()
		jsonStr := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapJson, SELF_ADDR, false)
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

package p3

import (
	"./data"
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"
)

var TA_SERVER = "http://localhost:6688"
var REGISTER_SERVER = TA_SERVER + "/peer"
var BC_DOWNLOAD_SERVER = TA_SERVER + "/upload"
var SELF_ADDR = "http://localhost:"
var SELF_PORT = os.Args[1]
var JSON_BLOCKCHAIN = "[{\"hash\": \"3ff3b4efe9177f705550231079c2459ba54a22d340a517e84ec5261a0d74ca48\", \"timeStamp\": 1234567890, \"height\": 1, \"parentHash\": \"genesis\", \"size\": 1174, \"mpt\": {\"hello\": \"world\", \"charles\": \"ge\"}}, {\"hash\": \"24cf2c336f02ccd526a03683b522bfca8c3c19aed8a1bed1bbc23c33cd8d1159\", \"timeStamp\": 1234567890, \"height\": 2, \"parentHash\": \"3ff3b4efe9177f705550231079c2459ba54a22d340a517e84ec5261a0d74ca48\", \"size\": 1231, \"mpt\": {\"hello\": \"world\", \"charles\": \"ge\"}}]"

var SBC data.SyncBlockChain
var Peers data.PeerList
var ifStarted bool

func init() {
	SBC = data.NewBlockChain()
	Peers = data.NewPeerList( /*Register()*/ 5, 32)
	//data.TestPeerListRebalance()
	//ifStarted = true

}

// Register ID, download BlockChain, start HeartBeat
func Start(w http.ResponseWriter, r *http.Request) {
	//Download()
	StartHeartBeat()
	if os.Args[2] == "yes" {
		print("Create blockchain with betty")
		SBC.UpdateEntireBlockChain(JSON_BLOCKCHAIN)
	} else {
		Download()
	}
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
	println(string(body))

	ip := r.RemoteAddr

	fmt.Print("Got heartbeat from : " + ip + " port: " + "\n")
}

// Ask another server to return a block of certain height and hash
func AskForBlock(height int32, hash string) {

}

func ForwardHeartBeat(heartBeatData data.HeartBeatData) {
	//Call from heartbeat received, if hops > 0 then forward to all nodes in peerslist
}

func sendInitHeartBeat() {
	if os.Args[1] != "6686" {
		url := "http://localhost:6686/heartbeat/receive"
		var jsonStr = []byte(`{"title":"Buy cheese and bread for breakfast."}`)
		resp, err := http.Post(url, "application/json", bytes.NewBuffer(jsonStr))
		if err == nil {
			print("Unable to send init heartbeat")
		}
		print(resp)
	}
}

func StartHeartBeat() {
	peerMapJson, _ := Peers.PeerMapToJson()
	data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapJson, SELF_ADDR, SELF_PORT)
	running := true
	for running {
		duration := time.Duration(10) * time.Second // Pause for 10 seconds
		time.Sleep(duration)
		for key, _ := range Peers.Copy() {
			peerMapJson, _ := Peers.PeerMapToJson()
			jsonStr := data.PrepareHeartBeatData(&SBC, Peers.GetSelfId(), peerMapJson, SELF_ADDR, SELF_PORT)
			jsonString, _ := json.Marshal(jsonStr)
			req, _ := http.NewRequest("POST", SELF_ADDR+key, bytes.NewBuffer(jsonString))

			req.Header.Set("X-Custom-Header", "myvalue")
			req.Header.Set("Content-Type", "application/json")
		}

	}
}

package data

import (
	"../../p1"
	"math/rand"
)

type HeartBeatData struct {
	IfNewBlock  bool   `json:"ifNewBlock"`
	Id          int32  `json:"id"`
	BlockJson   string `json:"blockJson"`
	PeerMapJson string `json:"peerMapJson"`
	Hops        int32  `json:"hops"`
	Addr        string `json:"Addr"`
}

//Hops should be decremented on forward
func NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string, addr string) HeartBeatData {
	return HeartBeatData{ifNewBlock, id, blockJson, peerMapJson, 3, addr}
}

//Function that generates the heartbeat data, and randomly creates new block to the blockchain
func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJSON string, addr string, init bool) HeartBeatData {
	randomVal := rand.Intn(100-0) + 0
	ifNewBlock := false
	newBlockJson := ""

	//Generate new block at random, if heartbeat message is init then it should not create a new block
	if randomVal >= 51 && init == false {
		mpt := p1.MerklePatriciaTrie{}
		mpt.Initial()
		newBlock := sbc.GenBlock(mpt)
		sbc.Insert(newBlock)
		newBlockJson = newBlock.EncodeToJSON()
		ifNewBlock = true
	}
	return NewHeartBeatData(ifNewBlock, selfId, newBlockJson, peerMapJSON, addr)
}

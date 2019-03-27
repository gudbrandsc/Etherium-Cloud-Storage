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
	IP          string `json:"IP"`
	Port        string `json:"Port"`
}

//Hops should be decremented on forward
func NewHeartBeatData(ifNewBlock bool, id int32, blockJson string, peerMapJson string, addr string, port string) HeartBeatData {
	return HeartBeatData{ifNewBlock, id, blockJson, peerMapJson, 3, addr, port}
}

func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJSON string, addr string, port string) HeartBeatData {
	randomVal := rand.Intn(100-0) + 0
	ifNewBlock := false
	newBlockJson := ""

	if randomVal >= 51 {
		mpt := p1.MerklePatriciaTrie{}
		mpt.Initial()
		newBlock := sbc.GenBlock(mpt)
		sbc.Insert(newBlock)
		newBlockJson = newBlock.EncodeToJSON()
		ifNewBlock = true
	}

	return NewHeartBeatData(ifNewBlock, selfId, newBlockJson, peerMapJSON, addr, port)

}

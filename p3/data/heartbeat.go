package data

import (
	"../../p1"
	"fmt"
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

func PrepareHeartBeatData(sbc *SyncBlockChain, selfId int32, peerMapJSON string, addr string, init bool) HeartBeatData {
	randomVal := rand.Intn(100-0) + 0
	ifNewBlock := false
	newBlockJson := ""

	if randomVal >= 41 && init == false {
		fmt.Println("Im creating a new block...")
		mpt := p1.NewMPT()
		newBlock := sbc.GenBlock(mpt)
		sbc.Insert(newBlock)
		newBlockJson = newBlock.EncodeToJSON()
		ifNewBlock = true
		return NewHeartBeatData(ifNewBlock, selfId, newBlockJson, peerMapJSON, addr)

	}

	return NewHeartBeatData(ifNewBlock, selfId, newBlockJson, peerMapJSON, addr)

}

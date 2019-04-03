package data

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
func PrepareHeartBeatData(selfId int32, peerMapJSON string, addr string) HeartBeatData {
	return NewHeartBeatData(false, selfId, "", peerMapJSON, addr)
}

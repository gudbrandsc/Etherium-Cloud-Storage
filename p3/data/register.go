package data

import "encoding/json"

type RegisterData struct {
	AssignedId  int32  `json:"assignedId"`
	PeerMapJson string `json:"peerMapJson"`
}

func NewRegisterData(id int32, peerMapJson string) RegisterData {

	//http://localhost:6688/peer --> Get node ID
	//ssh -L 6688:mc07.cs.usfca.edu:6688 gschistad@stargate.cs.usfca.edu --> Tunnel
}

func (data *RegisterData) EncodeToJson() (string, error) {}

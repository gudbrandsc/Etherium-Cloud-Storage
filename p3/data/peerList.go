package data

import (
	"container/ring"
	"encoding/json"
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"sync"
)

type PeerList struct {
	selfId    int32
	peerMap   map[string]int32
	maxLength int32
	mux       sync.Mutex
}

func NewPeerList(id int32, maxLength int32) PeerList {
	peerlist := PeerList{}
	peerlist.Register(id)
	peerlist.maxLength = maxLength
	peerlist.peerMap = make(map[string]int32)
	peerlist.mux = sync.Mutex{}

	return peerlist
}

func (peers *PeerList) Add(addr string, id int32) {
	peers.mux.Lock()
	if _, ok := peers.peerMap[addr]; !ok {
		peers.peerMap[addr] = id
	}
	peers.mux.Unlock()
}

func (peers *PeerList) Delete(addr string) {
	peers.mux.Lock()
	delete(peers.peerMap, addr)
	peers.mux.Unlock()
}

func (peers *PeerList) Rebalance() {
	peers.mux.Lock()
	if len(peers.peerMap) > int(peers.maxLength) {
		// Struct to used for array
		type kv struct {
			Key   string
			Value int32
		}

		var ss []kv
		middle := 0

		// Sort iterate over peers map
		for k, v := range peers.peerMap {
			ss = append(ss, kv{k, v})
		}

		//Sort peers array based on their ID
		sort.Slice(ss, func(i, j int) bool {
			return ss[i].Value < ss[j].Value
		})

		//Create new peers map
		peers.peerMap = make(map[string]int32)

		//Find the index where self ID would be
		for index, element := range ss {
			if element.Value > peers.selfId {
				middle = index
				break
			}
		}

		//Create a ring array
		r := ring.New(len(ss))
		for i := 0; i < r.Len(); i++ {
			r.Value = ss[i]
			r = r.Next()
		}

		// Start from the first element greater than our own ID
		r = r.Move(middle)
		for j := 0; j < int(peers.maxLength/2); j++ {
			peers.peerMap[r.Value.(kv).Key] = r.Value.(kv).Value
			r = r.Next()
		}

		//Move back to middle + 1 so that we start on the left element of where self ID should be
		for j := 0; j < int(peers.maxLength/2+1); j++ {
			r = r.Prev()
		}

		// Add n peers left of our own self ID
		for j := 0; j < int(peers.maxLength/2); j++ {
			peers.peerMap[r.Value.(kv).Key] = r.Value.(kv).Value
			r = r.Prev()
		}
	}
	peers.mux.Unlock()
}

func (peers *PeerList) Show() string {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	rs := "This is the PeerMap: \n"
	for key, value := range peers.peerMap {
		rs += "addr = " + key + ", id = " + strconv.Itoa(int(value)) + "\n"
	}
	return rs
}

func (peers *PeerList) Register(id int32) {
	peers.mux.Lock()
	peers.selfId = id
	fmt.Printf("SelfId = %v\n", id)
	peers.mux.Unlock()

}

func (peers *PeerList) Copy() map[string]int32 {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	return peers.peerMap
}

func (peers *PeerList) GetSelfId() int32 {
	peers.mux.Lock()
	defer peers.mux.Unlock()
	return peers.selfId
}

func (peers *PeerList) PeerMapToJson() (string, error) {
	peers.mux.Lock()
	jsonString, error := json.Marshal(peers.peerMap)
	defer peers.mux.Unlock()
	return string(jsonString), error

}

func (peers *PeerList) InjectPeerMapJson(peerMapJsonStr string, selfAddr string) {
	peers.mux.Lock()

	//Creating the interface maps for JSON
	m := map[string]int32{}
	err := json.Unmarshal([]byte(peerMapJsonStr), &m)

	if err != nil {
		fmt.Println(err)
	}

	peers.peerMap = m
	if _, ok := peers.peerMap[selfAddr]; ok {
		delete(peers.peerMap, selfAddr)
	}

	peers.mux.Unlock()
}

func TestPeerListRebalance() {
	peers := NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected := NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	expected.Add("-1-1", -1)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 2)
	peers.Add("1111", 1)
	peers.Add("4444", 4)
	peers.Add("-1-1", -1)
	peers.Add("0000", 0)
	peers.Add("2121", 21)
	peers.Rebalance()
	expected = NewPeerList(5, 2)
	expected.Add("4444", 4)
	expected.Add("2121", 21)
	fmt.Println(reflect.DeepEqual(peers, expected))

	peers = NewPeerList(5, 4)
	peers.Add("1111", 1)
	peers.Add("7777", 7)
	peers.Add("9999", 9)
	peers.Add("11111111", 11)
	peers.Add("2020", 20)
	peers.Rebalance()
	expected = NewPeerList(5, 4)
	expected.Add("1111", 1)
	expected.Add("7777", 7)
	expected.Add("9999", 9)
	expected.Add("2020", 20)
	fmt.Println(reflect.DeepEqual(peers, expected))
	peers.InjectPeerMapJson("{\"1111\":1,\"2020\":20,\"7777\":7,\"9999\":9}", "7777")
	print(peers.Show())
}

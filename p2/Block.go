package p2

import (
	"../p1"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"strings"
	"time"
)

type Block struct {
	Header Header
	Value  p1.MerklePatriciaTrie
}

// Structure used inside the Block structure
type Header struct {
	Height     int32
	Timestamp  int64
	Hash       string
	ParentHash string
	Size       int32
}

//Structure used to create a JSON object.
type Encoded_block struct {
	Hash       string            `json:"hash"`
	Timestamp  int64             `json:"timeStamp"`
	Height     int32             `json:"height"`
	ParentHash string            `json:"parentHash"`
	Size       int32             `json:"size"`
	Value      map[string]string `json:"mpt"`
}

// Initialise a new Block
func (b *Block) Initial(height int32, parentHash string, mpt p1.MerklePatriciaTrie) {

	//Set values from param
	header := Header{
		Height:     height,
		Timestamp:  time.Now().Unix(),
		ParentHash: parentHash,
	}
	b.Header = header
	b.Value = mpt

	//Create byte array from MPT and get the size.
	reqBodyBytes := new(bytes.Buffer)
	json.NewEncoder(reqBodyBytes).Encode(b.Value)
	b.Header.Size = int32(len(reqBodyBytes.Bytes()))

	//Create block hash
	hashStr := string(b.Header.Height) + string(b.Header.Timestamp) + b.Header.ParentHash + b.Value.Root + string(b.Header.Size)
	sum := sha3.Sum256([]byte(hashStr))
	b.Header.Hash = hex.EncodeToString(sum[:])
}

func (b *Block) EncodeToJSON() string {
	//Get Json data from entry map in mpt
	mptData := b.Value.EntryMap

	//Create json string with block values
	encodedB := &Encoded_block{
		Hash:       b.Header.Hash,
		Timestamp:  b.Header.Timestamp,
		Height:     b.Header.Height,
		ParentHash: b.Header.ParentHash,
		Size:       b.Header.Size,
		Value:      mptData,
	}
	result, _ := json.Marshal(encodedB)
	return string(result)
}

func DecodeFromJson(jsonString string) Block {

	var decoded Encoded_block
	err := json.NewDecoder(strings.NewReader(jsonString)).Decode(&decoded)
	if err != nil {
		fmt.Println(err)
	}

	header := Header{
		Height:     decoded.Height,
		Timestamp:  decoded.Timestamp,
		ParentHash: decoded.ParentHash,
		Hash:       decoded.Hash,
		Size:       decoded.Size,
	}

	mptEntryMap := decoded.Value
	mpt := new(p1.MerklePatriciaTrie)
	mpt.Initial()

	for k, v := range mptEntryMap {
		mpt.Insert(k, v)
	}

	block := Block{
		Header: header,
		Value:  *mpt,
	}
	return block
}

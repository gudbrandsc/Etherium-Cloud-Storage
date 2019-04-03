package p2

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
	"golang.org/x/crypto/sha3"
	"sort"
	"strconv"
	"strings"
)

type BlockChain struct {
	Chain  map[int32][]Block
	Length int32
}

func NewBlockChain() BlockChain {
	blockChain := BlockChain{}
	blockChain.Initial()
	return blockChain
}

// Initialise a new BlockChain
func (bc *BlockChain) Initial() {
	//Set values from param
	bc.Length = 0
	bc.Chain = make(map[int32][]Block)

}

//Get list of block at a height in the chain
func (bc *BlockChain) Get(height int32) ([]Block, bool) {
	getList, ok := bc.Chain[height]
	if ok {
		return getList, true
	}
	return nil, false
}

// Get a specific block in the chain, if the block does not exist then return empty block
func (bc *BlockChain) GetBlock(height int32, hash string) (Block, bool) {
	getList, ok := bc.Chain[height]
	if ok {
		for _, block := range getList {
			if block.GetHash() == hash {
				return block, true
			}
		}
	}
	return Block{}, false
}

// Insert a new block to the chain
func (bc *BlockChain) Insert(blc Block) {

	newBlockHeight := blc.GetHeight()
	blockList, _ := bc.Get(newBlockHeight)
	if len(blockList) == 0 {
		blockList = []Block{blc}
		if newBlockHeight > bc.Length {
			bc.Length = newBlockHeight
		}
	} else {
		//Check if the block is already in the chain
		exist := hashInArray(blc.GetHash(), blockList)

		if !exist {
			// If block does not exist it is added to the chain
			blockList = append(blockList, blc)
		}
	}

	bc.Chain[newBlockHeight] = blockList

	// Update chain height if the current block height is greater than current chain height
	if bc.Length < blc.GetHeight() {
		bc.Length = blc.GetHeight()
	}
}

//Check if the hash value of the block is already stored in the array.
func hashInArray(blockHash string, list []Block) bool {
	for _, block := range list {
		if block.GetHash() == blockHash {
			return true
		}
	}
	return false
}

// Convert the BlockChain to a JSON string
func (bc *BlockChain) EncodeToJSON() (string, error) {
	encodedBlockChain := "["

	// Iterate each index in the HashMap.
	for _, v := range bc.Chain {
		// For each index, iterate array of blocks.
		for _, element := range v {
			encodedBlockChain += element.EncodeToJSON() + ","
		}
	}
	encodedBlockChain = strings.TrimRight(encodedBlockChain, ",")
	encodedBlockChain += "]"

	return encodedBlockChain, nil
}

//Function that takes a json array string of blocks and creates a BlockChain containing every block DecodeJsonToBlockChain
func DecodeJsonToBlockChain(data string) BlockChain {
	//Create new a Blockchain
	blockChain := new(BlockChain)
	blockChain.Initial()

	// Store each block object as a json struct in blocks
	var blocks []Encoded_block
	err := json.Unmarshal([]byte(data), &blocks)
	if err != nil {
		fmt.Println("error:", err)
	}

	//Iterate each block stored in the chain
	for _, block := range blocks {
		val, err := json.Marshal(block)
		if err != nil {
			fmt.Println("error:", err)
		}
		blockChain.Insert(DecodeFromJson(string(val)))
	}
	return *blockChain
}

// Creates a human readable representation of the BlockChain
func (bc *BlockChain) Show() string {
	rs := ""
	var idList []int
	for id := range bc.Chain {
		idList = append(idList, int(id))
	}
	sort.Ints(idList)
	for _, id := range idList {
		var hashs []string
		for _, block := range bc.Chain[int32(id)] {
			hashs = append(hashs, block.GetHash()+"<="+block.GetParentHash())
		}
		sort.Strings(hashs)
		rs += fmt.Sprintf("%v: ", id)
		for _, h := range hashs {
			rs += fmt.Sprintf("%s, ", h)
		}
		rs += "\n"
	}
	sum := sha3.Sum256([]byte(rs))
	rs = fmt.Sprintf("This is the BlockChain: %s\n", hex.EncodeToString(sum[:])) + rs
	return rs
}

// This function returns the list of blocks of height "BlockChain.length".
func (bc *BlockChain) GetLatestBlocks() []Block {
	latestBlocksList, _ := bc.Get(bc.Length)
	return latestBlocksList
}

// This function takes a block as the parameter, and returns its parent block.
func (bc *BlockChain) GetParentBlock(block Block) (Block, bool) {
	return bc.GetBlock(block.Header.Height-1, block.Header.ParentHash)
}

func (bc *BlockChain) getLength() int32 {
	return bc.Length

}

// Gets all blocks at the end of the blockChain(longest chains) and returns human readable string
func (bc *BlockChain) ShowCanonical() string {
	bcTop := bc.GetLatestBlocks()
	count := 1
	rs := ""
	for _, block := range bcTop {
		rs += "Chain #" + strconv.Itoa(count) + "\n"
		rs += bc.GetChainString(block)
		rs += "\n\n"
		count++
	}
	return rs
}

// Recursive function that goes to genesis block
func (bc *BlockChain) GetChainString(block Block) string {
	bcString := ""
	bcString +=
		"height=" + strconv.Itoa(int(block.Header.Height)) +
			", timestamp=" + strconv.Itoa(int(block.Header.Timestamp)) +
			", hash=" + block.Header.Hash +
			", parenthash=" + block.Header.ParentHash +
			", size=" + strconv.Itoa((int(block.Header.Size))) + "\n"

	if block.Header.ParentHash != "genesis" {
		parentBlock, _ := bc.GetBlock(block.Header.Height-1, block.Header.ParentHash)
		bcString += bc.GetChainString(parentBlock)
	}

	return bcString
}

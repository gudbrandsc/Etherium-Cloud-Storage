package data

import (
	"../../p1"
	"../../p2"
	"sync"
)

// A synchronized BlockChain structure, that extends th normal BlockChain from P2
type SyncBlockChain struct {
	bc  p2.BlockChain
	mux sync.Mutex
}

// Create a new instance of a sync BlockChain
func NewBlockChain() SyncBlockChain {
	return SyncBlockChain{bc: p2.NewBlockChain()}
}

// Get a list of block at a current height
func (sbc *SyncBlockChain) Get(height int32) ([]p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Get(height)
}

// Returns a block at height with matching hash
func (sbc *SyncBlockChain) GetBlock(height int32, hash string) (p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetBlock(height, hash)
}

// Sync insert
func (sbc *SyncBlockChain) Insert(block p2.Block) {
	sbc.mux.Lock()
	sbc.bc.Insert(block)
	sbc.mux.Unlock()
}

// Checks if the blocks parent hash exist in the BlockChain
func (sbc *SyncBlockChain) CheckParentHash(insertBlock p2.Block) bool {
	sbc.mux.Lock()
	height := insertBlock.GetHeight()
	parentHeight := height - 1
	blockList, _ := sbc.bc.Get(parentHeight)
	sbc.mux.Unlock()
	for _, block := range blockList {
		if block.GetHash() == insertBlock.GetParentHash() {
			return true
		}
	}
	return false
}

// Updates the BlockChain to the JSON string BlockChain
func (sbc *SyncBlockChain) UpdateEntireBlockChain(blockChainJson string) {
	sbc.mux.Lock()
	sbc.bc = p2.DecodeJsonToBlockChain(blockChainJson)
	sbc.mux.Unlock()
}

// Sync convert from BlockChain to JSON representation
func (sbc *SyncBlockChain) BlockChainToJson() (string, error) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.EncodeToJSON()
}

// Generate a new block
func (sbc *SyncBlockChain) GenBlock(mpt p1.MerklePatriciaTrie) p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	height := sbc.bc.Length
	blockList, _ := sbc.bc.Get(height)
	if len(blockList) == 0 {
		blockList, _ = sbc.bc.Get(height - 1)
	}
	parentBlock := blockList[0] // Can be selected at random but didnt see the reason to yet
	parentHash := parentBlock.GetHash()
	return p2.Initial(height+1, parentHash, mpt)
}

// Return a string of the current BlockChain
func (sbc *SyncBlockChain) Show() string {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Show()
}

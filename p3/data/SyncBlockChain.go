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
func (sbc *SyncBlockChain) GenBlock(mpt p1.MerklePatriciaTrie, nonce string, parentHash string) p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	height := sbc.bc.Length
	return p2.Initial(height+1, parentHash, mpt, nonce)
}

// Return a string of the current BlockChain
func (sbc *SyncBlockChain) Show() string {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.Show()
}

// This function returns the list of blocks of height "BlockChain.length".
func (sbc *SyncBlockChain) GetLatestBlocks() []p2.Block {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetLatestBlocks()
}

// This function takes a block as the parameter, and returns its parent block.
func (sbc *SyncBlockChain) GetParentBlock(block p2.Block) (p2.Block, bool) {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.GetParentBlock(block)
}

func (sbc *SyncBlockChain) GetChainLength() int32 {
	return sbc.bc.Length
}

func (sbc *SyncBlockChain) ShowCanonical() string {
	sbc.mux.Lock()
	defer sbc.mux.Unlock()
	return sbc.bc.ShowCanonical()

}

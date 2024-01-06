package node

import (
	"bytes"
	"encoding/hex"
	"fmt"
	"github.com/cmkqwerty/blocker/crypto"
	"github.com/cmkqwerty/blocker/proto"
	"github.com/cmkqwerty/blocker/types"
)

type HeaderList struct {
	headers []*proto.Header
}

func NewHeaderList() *HeaderList {
	return &HeaderList{[]*proto.Header{}}
}

func (h *HeaderList) Add(header *proto.Header) {
	h.headers = append(h.headers, header)
}

func (h *HeaderList) Get(index int) *proto.Header {
	if index > h.Height() {
		panic("index out of range")
	}

	return h.headers[index]
}

func (h *HeaderList) Height() int {
	return h.Len() - 1
}

func (h *HeaderList) Len() int {
	return len(h.headers)
}

type Chain struct {
	blockStore BlockStorer
	headers    *HeaderList
}

func NewChain(blockStore BlockStorer) *Chain {
	chain := &Chain{
		blockStore: blockStore,
		headers:    NewHeaderList(),
	}

	err := chain.addBlock(createGenesisBlock())
	if err != nil {
		panic(err)
	}

	return chain
}

func (c *Chain) Height() int {
	return c.headers.Height()
}

func (c *Chain) AddBlock(block *proto.Block) error {
	if err := c.ValidateBlock(block); err != nil {
		return err
	}

	return c.addBlock(block)
}

func (c *Chain) addBlock(block *proto.Block) error {
	c.headers.Add(block.Header)
	return c.blockStore.Put(block)
}

func (c *Chain) GetBlockByHash(hash []byte) (*proto.Block, error) {
	hashHex := hex.EncodeToString(hash)

	return c.blockStore.Get(hashHex)
}

func (c *Chain) GetBlockByHeight(height int) (*proto.Block, error) {
	if c.Height() < height {
		return nil, fmt.Errorf("given height (%d) too high - current height (%d)", height, c.Height())
	}

	header := c.headers.Get(height)
	hash := types.HashHeader(header)

	return c.GetBlockByHash(hash)
}

func (c *Chain) ValidateBlock(block *proto.Block) error {
	// validate signature
	if !types.VerifyBlock(block) {
		return fmt.Errorf("invalid block signature")
	}

	// validate prev block hash
	currentBlock, err := c.GetBlockByHeight(c.Height())
	if err != nil {
		return err
	}

	hash := types.HashBlock(currentBlock)
	if !bytes.Equal(hash, block.Header.PrevHash) {
		return fmt.Errorf("prev block hash mismatch")
	}

	return nil
}

func createGenesisBlock() *proto.Block {
	privKey := crypto.GeneratePrivateKey()
	block := &proto.Block{
		Header: &proto.Header{
			Version: 1,
		},
		Transactions: []*proto.Transaction{},
	}

	types.SignBlock(privKey, block)
	return block
}

package node

import (
	"encoding/hex"
	"fmt"
	"github.com/cmkqwerty/blocker/crypto"
	"github.com/cmkqwerty/blocker/proto"
	"github.com/cmkqwerty/blocker/types"
	"github.com/cmkqwerty/blocker/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
)

func randomBlock(t *testing.T, chain *Chain) *proto.Block {
	privKey := crypto.GeneratePrivateKey()
	block := util.RandomBlock()
	prevBlock, err := chain.GetBlockByHeight(chain.Height())
	require.Nil(t, err)

	block.Header.PrevHash = types.HashBlock(prevBlock)
	types.SignBlock(privKey, block)

	return block
}

func TestNewChain(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
	require.Equal(t, 0, chain.Height())

	_, err := chain.GetBlockByHeight(0)
	require.Nil(t, err)
}

func TestChainHeight(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())

	for i := 0; i < 10; i++ {
		block := randomBlock(t, chain)

		require.Nil(t, chain.AddBlock(block))
		require.Equal(t, i+1, chain.Height())
	}
}

func TestAddBlock(t *testing.T) {
	chain := NewChain(NewMemoryBlockStore(), NewMemoryTXStore())

	for i := 0; i < 100; i++ {
		block := randomBlock(t, chain)
		blockHash := types.HashBlock(block)

		require.Nil(t, chain.AddBlock(block))

		fetchedBlock, err := chain.GetBlockByHash(blockHash)
		require.Nil(t, err)
		assert.Equal(t, block, fetchedBlock)

		fetchedBlockByHeight, err := chain.GetBlockByHeight(i + 1)
		require.Nil(t, err)
		require.Equal(t, block, fetchedBlockByHeight)
	}
}

func TestAddBlockWithTX(t *testing.T) {
	var (
		chain      = NewChain(NewMemoryBlockStore(), NewMemoryTXStore())
		block      = randomBlock(t, chain)
		privateKey = crypto.NewPrivateKeyFromSeedString(godSeed)
		recipient  = crypto.GeneratePrivateKey().Public().Address().Bytes()
	)

	ftx, err := chain.txStore.Get("737cb341c1dd4a34f78faeb9ae8ffc07b5e9cf987a1ef0bfb65d98542f48aacb")
	assert.Nil(t, err)

	inputs := []*proto.TxInput{
		{
			PrevTxHash:   types.HashTransaction(ftx),
			PrevOutIndex: 0,
			PublicKey:    privateKey.Public().Bytes(),
		},
	}
	outputs := []*proto.TxOutput{
		{
			Amount:  100,
			Address: recipient,
		},
		{
			Amount:  900,
			Address: privateKey.Public().Address().Bytes(),
		},
	}
	tx := &proto.Transaction{
		Version: 1,
		Inputs:  inputs,
		Outputs: outputs,
	}

	signature := types.SignTransaction(privateKey, tx)
	tx.Inputs[0].Signature = signature.Bytes()

	block.Transactions = append(block.Transactions, tx)
	require.Nil(t, chain.AddBlock(block))

	txHash := hex.EncodeToString(types.HashTransaction(tx))
	fetchedTx, err := chain.txStore.Get(txHash)
	require.Nil(t, err)
	require.Equal(t, tx, fetchedTx)

	// check utxo for unspent
	address := crypto.AddressFromBytes(tx.Outputs[0].Address)
	key := fmt.Sprintf("%s_%s", address, txHash)

	_, err = chain.utxoStore.Get(key)
	require.Nil(t, err)
}

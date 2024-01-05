package types

import (
	"github.com/cmkqwerty/blocker/crypto"
	"github.com/cmkqwerty/blocker/proto"
	"github.com/cmkqwerty/blocker/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestNewTransaction(t *testing.T) {
	fromPrivateKey := crypto.GeneratePrivateKey()
	fromAddress := fromPrivateKey.Public().Address().Bytes()
	toPrivateKey := crypto.GeneratePrivateKey()
	toAddress := toPrivateKey.Public().Address().Bytes()

	input := &proto.TxInput{
		PrevTxHash:   util.RandomHash(),
		PrevOutIndex: 0,
		PublicKey:    fromPrivateKey.Public().Bytes(),
	}

	output1 := &proto.TxOutput{
		Amount:  5,
		Address: toAddress,
	}

	output2 := &proto.TxOutput{
		Amount:  95,
		Address: fromAddress,
	}

	tx := &proto.Transaction{
		Version: 1,
		Inputs:  []*proto.TxInput{input},
		Outputs: []*proto.TxOutput{output1, output2},
	}

	signature := SignTransaction(fromPrivateKey, tx)
	input.Signature = signature.Bytes()

	assert.True(t, VerifyTransaction(tx))
}

package types

import (
	"github.com/cmkqwerty/blocker/crypto"
	"github.com/cmkqwerty/blocker/util"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestVerifyBlock(t *testing.T) {
	var (
		block      = util.RandomBlock()
		privateKey = crypto.GeneratePrivateKey()
		publicKey  = privateKey.Public()
	)

	signature := SignBlock(privateKey, block)

	assert.Equal(t, 64, len(signature.Bytes()))
	assert.True(t, signature.Verify(publicKey, HashBlock(block)))

	assert.Equal(t, block.PublicKey, publicKey.Bytes())
	assert.Equal(t, block.Signature, signature.Bytes())

	assert.True(t, VerifyBlock(block))

	invalidPrivKey := crypto.GeneratePrivateKey()
	block.PublicKey = invalidPrivKey.Public().Bytes()

	assert.False(t, VerifyBlock(block))
}

func TestHashBlock(t *testing.T) {
	block := util.RandomBlock()
	hash := HashBlock(block)

	assert.Equal(t, 32, len(hash))
}

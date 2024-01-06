package types

import (
	"crypto/sha256"
	"github.com/cmkqwerty/blocker/crypto"
	"github.com/cmkqwerty/blocker/proto"
	pb "google.golang.org/protobuf/proto"
)

func SignBlock(pk *crypto.PrivateKey, block *proto.Block) *crypto.Signature {
	hash := HashBlock(block)
	signature := pk.Sign(hash)
	block.PublicKey = pk.Public().Bytes()
	block.Signature = signature.Bytes()

	return signature
}

// HashBlock returns SHA256 of the header.
func HashBlock(block *proto.Block) []byte {
	return HashHeader(block.Header)
}

func HashHeader(header *proto.Header) []byte {
	b, err := pb.Marshal(header)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)

	return hash[:]
}

func VerifyBlock(block *proto.Block) bool {
	if len(block.PublicKey) != crypto.PublicKeyLen {
		return false
	}
	if len(block.Signature) != crypto.SignatureLen {
		return false
	}

	signature := crypto.SignatureFromBytes(block.Signature)
	publicKey := crypto.PublicKeyFromBytes(block.PublicKey)
	hash := HashBlock(block)

	return signature.Verify(publicKey, hash)
}

package types

import (
	"crypto/sha256"
	"github.com/cmkqwerty/blocker/crypto"
	"github.com/cmkqwerty/blocker/proto"
	pb "google.golang.org/protobuf/proto"
)

func SignTransaction(pk *crypto.PrivateKey, tx *proto.Transaction) *crypto.Signature {
	return pk.Sign(HashTransaction(tx))
}

func HashTransaction(tx *proto.Transaction) []byte {
	b, err := pb.Marshal(tx)
	if err != nil {
		panic(err)
	}

	hash := sha256.Sum256(b)
	return hash[:]
}

func VerifyTransaction(tx *proto.Transaction) bool {
	for _, input := range tx.Inputs {
		if len(input.Signature) == 0 {
			panic("transaction with no signature")
		}

		var (
			signature = crypto.SignatureFromBytes(input.Signature)
			publicKey = crypto.PublicKeyFromBytes(input.PublicKey)
		)

		// TODO: make sure dont run into problems after verify
		input.Signature = nil
		if !signature.Verify(publicKey, HashTransaction(tx)) {
			return false
		}
	}

	return true
}

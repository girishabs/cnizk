package main

import (
	"crypto/ecdsa"
	"encoding/hex"
	"os"

	"github.com/ethereum/go-ethereum/crypto"
)

var privateKey *ecdsa.PrivateKey

// Initialize once at startup
func InitSigner() {
	hexKey := os.Getenv("SIGNING_KEY")
	if hexKey == "" {
		panic("SIGNING_KEY not set")
	}

	keyBytes, err := hex.DecodeString(hexKey)
	if err != nil {
		panic(err)
	}

	privateKey, err = crypto.ToECDSA(keyBytes)
	if err != nil {
		panic(err)
	}
}

// Sign 32-byte hash
func SignHash(hash []byte) []byte {
	if len(hash) != 32 {
		panic("hash must be 32 bytes")
	}

	sig, err := crypto.Sign(hash, privateKey)
	if err != nil {
		panic(err)
	}

	return sig[:64] // remove recovery id
}
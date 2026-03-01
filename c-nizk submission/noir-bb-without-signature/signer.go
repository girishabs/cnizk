package main

import (
	"encoding/hex"
	"log"
	"os"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secpECDSA "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"golang.org/x/crypto/blake2s"
)

var privKey *secp.PrivateKey
var pubKey []byte

func InitSigner() {

	keyHex := os.Getenv("SIGNING_KEY")
	if keyHex == "" {
		log.Fatal("SIGNING_KEY not set")
	}

	keyBytes, err := hex.DecodeString(keyHex)
	if err != nil {
		log.Fatal(err)
	}

	privKey = secp.PrivKeyFromBytes(keyBytes)
	pubKey = privKey.PubKey().SerializeUncompressed()
}

func SignMessage(data []byte) ([]byte, []byte) {

	hash := blake2s.Sum256(data)

	sig := secpECDSA.Sign(privKey, hash[:])

	r := sig.R()
	s := sig.S()

	var rBytes [32]byte
	var sBytes [32]byte

	(&r).PutBytes(&rBytes)
	(&s).PutBytes(&sBytes)

	compact := append(rBytes[:], sBytes[:]...)

	return hash[:], compact
}
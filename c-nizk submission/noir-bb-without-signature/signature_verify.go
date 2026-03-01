package main

import (
	"encoding/base64"
	"fmt"
	"time"

	secp "github.com/decred/dcrd/dcrec/secp256k1/v4"
	secpECDSA "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
)

func VerifyAttestation(
	messageB64 string,
	signatureB64 string,
	pubKeyXB64 string,
	pubKeyYB64 string,
) bool {

	message, err := base64.StdEncoding.DecodeString(messageB64)
	if err != nil {
		return false
	}

	sigBytes, err := base64.StdEncoding.DecodeString(signatureB64)
	if err != nil {
		return false
	}

	pubX, err := base64.StdEncoding.DecodeString(pubKeyXB64)
	if err != nil {
		return false
	}

	pubY, err := base64.StdEncoding.DecodeString(pubKeyYB64)
	if err != nil {
		return false
	}

	if len(sigBytes) != 64 {
		return false
	}

	var r secp.ModNScalar
	var s secp.ModNScalar

	r.SetByteSlice(sigBytes[:32])
	s.SetByteSlice(sigBytes[32:])

	sig := secpECDSA.NewSignature(&r, &s)

	pub := make([]byte, 65)
	pub[0] = 0x04 

	copy(pub[1:33], pubX)
	copy(pub[33:], pubY)

	pubKey, err := secp.ParsePubKey(pub)
	if err != nil {
		return false
	}

	start := time.Now()

	ok := sig.Verify(message, pubKey)

	elapsed := time.Since(start)
	fmt.Println("time taken to verify the signature: ", elapsed)
	
	return ok
}

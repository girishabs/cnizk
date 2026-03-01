package verifier

import (
	"bytes"
	"fmt"
	"math/big"
	"time"

	"zk-snarks/constant"
	"zk-snarks/model"
	"zk-snarks/utils"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	mimcCrypto "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/mimc"
	eddsaCrypto "github.com/consensys/gnark-crypto/ecc/bls12-377/twistededwards/eddsa"

	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
)

func Verify(proverPayload model.ProverPayload, vk []byte) error {

	proof := groth16.NewProof(ecc.BLS12_377)
	proof.ReadFrom(bytes.NewBuffer(proverPayload.Proof))

	var buffer bytes.Buffer
	buffer.Write(vk)

	verificationKey := groth16.NewVerifyingKey(ecc.BLS12_377)
	verificationKey.ReadFrom(&buffer)

	statementToVerify := []model.Statement{
		{
			Attribute: constant.OWNER_ATTRIBUTE,
			Type:      constant.STRING,
			Operation: constant.EQUAL,
			Value:     "abcd",
		},
		{
			Attribute: constant.STATUS_ATTRIBUTE,
			Type:      constant.BOOL,
			Operation: constant.EQUAL,
			Value:     true,
		},
		{
			Attribute: constant.VALUE_ATTRIBUTE,
			Type:      constant.INT,
			Operation: constant.EQUAL,
			Value:     300000,
		},
		{
			Attribute: constant.TYPE_ATTRIBUTE,
			Type:      constant.STRING,
			Operation: constant.EQUAL,
			Value:     "cbdc",
		},
		{
			Attribute: constant.VERSION_ATTRIBUTE,
			Type:      constant.STRING,
			Operation: constant.EQUAL,
			Value:     "1.0.0",
		},
	}

	constraintsVerification, err := utils.ExtractConstraints(statementToVerify)
	if err != nil {
		panic(err)
	}

	for i, cons := range constraintsVerification {
		opFe := new(fr.Element).SetBigInt(cons.Operation)
		constraintsVerification[i].Operation = opFe.BigInt(new(big.Int))

		valFe := new(fr.Element).SetBigInt(cons.Value)
		constraintsVerification[i].Value = valFe.BigInt(new(big.Int))

		attributeTypeFe := new(fr.Element).SetBigInt(cons.Type)
		constraintsVerification[i].Type = attributeTypeFe.BigInt(new(big.Int))
	}

	circuitConstraints := make([]struct {
		Operation frontend.Variable
		Value     frontend.Variable
		Type      frontend.Variable
	}, 5)
	for i := range constraintsVerification {
		circuitConstraints[i].Operation = constraintsVerification[i].Operation
		circuitConstraints[i].Type = constraintsVerification[i].Type
		circuitConstraints[i].Value = constraintsVerification[i].Value
	}

	var signature eddsaCrypto.Signature
	_, err = signature.SetBytes(proverPayload.Signature)
	if err != nil {
		panic(err)
	}

	var publicKey eddsaCrypto.PublicKey
	_, err = publicKey.SetBytes(proverPayload.PublicKey)
	if err != nil {
		panic(err)
	}

	hashFunction := mimcCrypto.NewMiMC()

	start := time.Now()
	ok, err := publicKey.Verify(proverPayload.Signature, proverPayload.Hash, hashFunction)
	end := time.Since(start)
	fmt.Printf("Signature verification time: %v\n", end)

	fmt.Printf("ok: %v\n", ok)
	fmt.Printf("err: %v\n", err)

	assetAssignment := model.NewCircuit{
		Constraints:       circuitConstraints,
		CreationTimestamp: big.NewInt(proverPayload.CreationTimestamp),
		Message:           proverPayload.Hash,
	}

	assetWitness, err := frontend.NewWitness(assetAssignment, ecc.BLS12_377.ScalarField(), frontend.PublicOnly())
	if err != nil {
		panic(err)
	}

	return groth16.Verify(proof, verificationKey, assetWitness)
}

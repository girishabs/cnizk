package prover

import (
	"bytes"
	"crypto/rand"
	"fmt"

	"math/big"
	"time"

	"zk-snarks/model"
	"zk-snarks/utils"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	mimcCrypto "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/mimc"
	eddsaCrypto "github.com/consensys/gnark-crypto/ecc/bls12-377/twistededwards/eddsa"
	"github.com/consensys/gnark/backend"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/constraint/solver"
	"github.com/consensys/gnark/frontend"
)

func Prove(asset model.Asset, statement []model.Statement, creationTimestamp int64, ccsBytes []byte, pkBytes []byte, mp map[int64]int64) model.ProverPayload {

	data := utils.ExtractData(&asset)
	constraints, err := utils.ExtractConstraints(statement)
	if err != nil {
		panic(err)
	}

	currentUnix := creationTimestamp
	hashFunction := mimcCrypto.NewMiMC()

	allData := make([]*big.Int, len(data)+1)
	copy(allData, data)
	allData[len(data)] = big.NewInt(currentUnix)

	values := make([]fr.Element, len(allData))
	for i, d := range allData {
		values[i].SetBigInt(d)
	}

	for i, d := range data {
		dataFe := new(fr.Element).SetBigInt(d)
		data[i] = dataFe.BigInt(new(big.Int))
		arr := dataFe.Bytes()
		hashFunction.Write(arr[:])
	}

	timestampFe := new(fr.Element).SetBigInt(big.NewInt(currentUnix))
	arr := timestampFe.Bytes()
	hashFunction.Write(arr[:])

	hash := hashFunction.Sum(nil)

	privateKey, err := eddsaCrypto.GenerateKey(rand.Reader)
	if err != nil {
		panic(err)
	}

	publicKey := privateKey.PublicKey
	start := time.Now()
	signedMessage, err := privateKey.Sign(hash, hashFunction)
	if err != nil {
		panic(err)
	}
	end := time.Since(start)
	fmt.Printf("Signing time: %v\n", end)

	var signature eddsaCrypto.Signature
	_, err = signature.SetBytes(signedMessage)
	if err != nil {
		panic(err)
	}

	dataVector := make([]frontend.Variable, 5)
	for i := range dataVector {
		dataVector[i] = data[i]
	}

	constraintVector := make([]struct {
		Operation frontend.Variable
		Value     frontend.Variable
		Type      frontend.Variable
	}, 5)
	for i, constraint := range constraints {
		constraintVector[i].Operation = constraint.Operation
		constraintVector[i].Type = constraint.Type
		constraintVector[i].Value = constraint.Value
	}

	assetAssignment := model.NewCircuit{
		Data:              dataVector,
		Constraints:       constraintVector,
		CreationTimestamp: big.NewInt(currentUnix),
		Message:           hash,
		PublicKey:         utils.ConvertPublicKey(publicKey),
		Signature:         utils.ConvertSignature(signature),
	}

	witness, err := frontend.NewWitness(&assetAssignment, ecc.BLS12_377.ScalarField())
	if err != nil {
		panic(err)
	}

	var ccsBuffer bytes.Buffer
	_, err = ccsBuffer.Write(ccsBytes)
	if err != nil {
		panic(err)
	}

	ccs := groth16.NewCS(ecc.BLS12_377)
	_, err = ccs.ReadFrom(&ccsBuffer)
	if err != nil {
		panic(err)
	}

	var pkBuffer bytes.Buffer
	_, err = pkBuffer.Write(pkBytes)
	if err != nil {
		panic(err)
	}

	groth16Pk := groth16.NewProvingKey(ecc.BLS12_377)
	_, err = groth16Pk.ReadFrom(&pkBuffer)
	if err != nil {
		panic(err)
	}

	start = time.Now()

	proof, err := groth16.Prove(ccs, groth16Pk, witness, backend.WithSolverOptions(solver.WithHints(model.InverseHint)))
	if err != nil {
		panic(err)
	}

	elapsed := time.Since(start)
	fmt.Printf("Proving time: %v\n", elapsed)

	var proofbuffer bytes.Buffer
	proof.WriteRawTo(&proofbuffer)

	signatureBytes := signature.Bytes()
	publicKeyBytes := publicKey.Bytes()

	proverPayload := model.ProverPayload{
		Proof:             proofbuffer.Bytes(),
		Hash:              hash,
		CreationTimestamp: currentUnix,
		Signature:         signatureBytes,
		PublicKey:         publicKeyBytes,
	}

	return proverPayload
}

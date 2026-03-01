package circuit

import (
	"bytes"
	"math/big"
	"time"

	"zk-snarks/constant"
	"zk-snarks/model"
	"zk-snarks/utils"

	"github.com/consensys/gnark-crypto/ecc"
	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	mimcCrypto "github.com/consensys/gnark-crypto/ecc/bls12-377/fr/mimc"
	"github.com/consensys/gnark/backend/groth16"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/frontend/cs/r1cs"
)

func GenerateProof() (Ccs []byte, Pk []byte, Vk []byte) {
	asset := model.Asset{
		Owner:   "abcd",
		Version: "1.0.0",
		Type:    "cbdc",
		Value:   300000,
		Status:  true,
	}

	statement := []model.Statement{
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

	data := utils.ExtractData(&asset)

	constraints, err := utils.ExtractConstraints(statement)
	if err != nil {
		panic(err)
	}

	currentUnix := time.Now().Unix()
	hashFunction := mimcCrypto.NewMiMC()

	for i, d := range data {
		dataFe := new(fr.Element).SetBigInt(d)
		data[i] = dataFe.BigInt(new(big.Int))
		arr := dataFe.Bytes()
		hashFunction.Write(arr[:])
	}

	timestampFe := new(fr.Element).SetBigInt(big.NewInt(currentUnix))
	arr := timestampFe.Bytes()
	hashFunction.Write(arr[:])

	rBytes, err := utils.Generate128BitRandom()
	if err != nil {
		panic(err)
	}

	rBig := new(big.Int).SetBytes(rBytes)
	var rFe fr.Element
	rFe.SetBigInt(rBig)

	arr = rFe.Bytes()
	hashFunction.Write(arr[:])

	hash := hashFunction.Sum(nil)

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
		Randomness:        rFe,
	}

	ccs, err := frontend.Compile(ecc.BLS12_377.ScalarField(), r1cs.NewBuilder, &assetAssignment)
	if err != nil {
		panic(err)
	}

	pk, vk, err := groth16.Setup(ccs)
	if err != nil {
		panic(err)
	}

	var ccsBuffer bytes.Buffer
	ccs.WriteTo(&ccsBuffer)

	var pkBuffer bytes.Buffer
	pk.WriteTo(&pkBuffer)

	var vkBuffer bytes.Buffer
	vk.WriteTo(&vkBuffer)

	return ccsBuffer.Bytes(), pkBuffer.Bytes(), vkBuffer.Bytes()
}

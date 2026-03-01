package model

import (
	"hash/fnv"
	"math/big"

	tedwards "github.com/consensys/gnark-crypto/ecc/twistededwards"
	"github.com/consensys/gnark/frontend"
	"github.com/consensys/gnark/std/algebra/native/twistededwards"
	"github.com/consensys/gnark/std/hash/mimc"
	"github.com/consensys/gnark/std/signature/eddsa"
)

type NewCircuit struct {
	Data      []frontend.Variable
	Signature eddsa.Signature

	Constraints []struct {
		Operation frontend.Variable
		Value     frontend.Variable
		Type      frontend.Variable //0=int, 1=string, 2=bool
	} `gnark:",public"`
	CreationTimestamp frontend.Variable `gnark:",public"`
	Message           frontend.Variable `gnark:",public"`
	PublicKey         eddsa.PublicKey   `gnark:",public"`
}

func InverseHint(mod *big.Int, inputs, outputs []*big.Int) error {
	diff := inputs[0]
	inv := new(big.Int)

	if diff.Cmp(big.NewInt(0)) == 0 {
		outputs[0].SetInt64(0)
		return nil
	}

	inv.ModInverse(diff, mod)
	outputs[0].Set(inv)
	return nil
}

func (c NewCircuit) Define(api frontend.API) error {
	const k = 66 // bit-length of the field

	fieldHasher, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	allData := make([]frontend.Variable, len(c.Data)+1)
	copy(allData, c.Data)
	allData[len(c.Data)] = c.CreationTimestamp

	for _, data := range allData {
		fieldHasher.Write(data)
	}

	for i, cons := range c.Constraints {

		dataVal := c.Data[i]
		op := cons.Operation
		val := cons.Value
		attributeType := cons.Type

		isBool := api.IsZero(api.Sub(attributeType, 2))

		isEq := api.IsZero(api.Sub(op, 0))
		isNe := api.IsZero(api.Sub(op, 1))
		isLt := api.IsZero(api.Sub(op, 2))
		isGe := api.IsZero(api.Sub(op, 3))
		isGt := api.IsZero(api.Sub(op, 4))
		isLe := api.IsZero(api.Sub(op, 5))
		api.AssertIsEqual(api.Add(isEq, isNe, isLt, isLe, isGt, isGe), 1)

		// check for types, impose only equality and inequality on strings
		isHashable := api.IsZero(api.Sub(attributeType, 1))
		invalidHashableOp := api.Or(isLt, api.Or(isLe, api.Or(isGt, isGe)))
		api.AssertIsEqual(api.Mul(invalidHashableOp, isHashable), 0)

		// BOOLEANS RESTRICTED TO 0/1
		api.AssertIsEqual(api.Mul(isBool, api.Mul(dataVal, api.Sub(dataVal, 1))), 0)
		api.AssertIsEqual(api.Mul(isBool, api.Mul(val, api.Sub(val, 1))), 0)

		// EQUALITY AND INEQUALITY
		diff1 := api.Sub(dataVal, val)
		invDiff, err := api.NewHint(InverseHint, 1, diff1)
		if err != nil {
			return err
		}
		api.AssertIsEqual(api.Mul(isEq, diff1), 0)
		api.AssertIsEqual(api.Mul(diff1, invDiff[0], isNe), isNe)

		// NUMERIC RELATIONAL OPERATIONS
		diff2 := api.Sub(val, dataVal)
		tLt := api.Select(isLt, api.Sub(diff2, 1), 0)
		tGe := api.Select(isGe, diff1, 0)
		tGt := api.Select(isGt, api.Sub(diff1, 1), 0)
		tLe := api.Select(isLe, diff2, 0)

		// MASKING OPERATIONS
		bitsLt := api.ToBinary(tLt, k)
		bitsGe := api.ToBinary(tGe, k)
		bitsGt := api.ToBinary(tGt, k)
		bitsLe := api.ToBinary(tLe, k)

		// LESS-THAN
		api.AssertIsEqual(api.FromBinary(bitsLt...), tLt)

		// GREATER-THAN EQUAL TO
		api.AssertIsEqual(api.FromBinary(bitsGe...), tGe)

		// GREATER-THAN
		api.AssertIsEqual(api.FromBinary(bitsGt...), tGt)

		// LESS-THAN EQUAL TO
		api.AssertIsEqual(api.FromBinary(bitsLe...), tLe)
	}

	reconstructedHash := fieldHasher.Sum()
	api.AssertIsEqual(reconstructedHash, c.Message)

	curve, err := twistededwards.NewEdCurve(api, tedwards.BLS12_377)
	if err != nil {
		return err
	}

	sigHasher, err := mimc.NewMiMC(api)
	if err != nil {
		return err
	}

	return eddsa.Verify(curve, c.Signature, c.Message, c.PublicKey, &sigHasher)
}

type Asset struct {
	Owner   string // 0
	Status  bool   // 1
	Value   int    // 2
	Type    string // 3
	Version string // 4
}

type CcsAndProvingKey struct {
	Ccs []byte
	Pk  []byte
}

type ProverPayload struct {
	Proof             []byte
	Hash              []byte
	CreationTimestamp int64
	Signature         []byte
	PublicKey         []byte
	Nonce             []byte
}

type VerificationKey struct {
	Vk []byte
}

func HashStringToInt(s string) uint64 {
	h := fnv.New64a()
	h.Write([]byte(s))
	return h.Sum64()
}

type Statement struct {
	Attribute string
	Type      string
	Operation string
	Value     any
}

type AssetRequest struct {
	AssetDetails      Asset       `json:"asset"`
	StatementDetails  []Statement `json:"statement"`
	CreationTimestamp int64       `json:"creationTimestamp"`
}

type VerifyResponse struct {
	Verified bool `json:"verified"`
}

type ApiResponse struct {
	Status  string      `json:"status"`
	Message string      `json:"message,omitempty"`
	Data    interface{} `json:"data,omitempty"`
}

type WorkResult struct {
	Data VerifyResponse
	Err  error
}

// 0 -> equality
// 1 -> inequality
// 2 -> less than
// 3 -> greater than equal to
// 4 -> greater than
// 5 -> less than equal to

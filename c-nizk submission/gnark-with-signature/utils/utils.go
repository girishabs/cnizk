package utils

import (
	"fmt"
	"math/big"
	"reflect"

	"zk-snarks/constant"
	"zk-snarks/model"

	"github.com/consensys/gnark-crypto/ecc/bls12-377/fr"
	eddsaCrypto "github.com/consensys/gnark-crypto/ecc/bls12-377/twistededwards/eddsa"
	"github.com/consensys/gnark/std/algebra/native/twistededwards"
	"github.com/consensys/gnark/std/signature/eddsa"
)

func ExtractData(asset *model.Asset) []*big.Int {
	data := make([]*big.Int, 5)
	data[0] = new(big.Int).SetUint64(model.HashStringToInt(asset.Owner))
	data[3] = new(big.Int).SetUint64(model.HashStringToInt(asset.Type))
	data[4] = new(big.Int).SetUint64(model.HashStringToInt(asset.Version))
	data[2] = new(big.Int).SetUint64(uint64(asset.Value))
	if asset.Status {
		data[1] = new(big.Int).SetUint64(1)
	} else {
		data[1] = new(big.Int).SetUint64(0)
	}
	return data
}

func ExtractConstraints(statement []model.Statement) ([]struct {
	Operation *big.Int
	Value     *big.Int
	Type      *big.Int
}, error) {

	constraints := make([]struct {
		Operation *big.Int
		Value     *big.Int
		Type      *big.Int
	}, 5)

	for i := range constraints {
		constraints[i] = struct {
			Operation *big.Int
			Value     *big.Int
			Type      *big.Int
		}{}

		switch statement[i].Type {
		case constant.STRING:
			stringValue, ok := statement[i].Value.(string)
			if !ok {
				return nil, (fmt.Errorf("variable %s not of string type", statement[i].Attribute))
			}
			constraints[i].Type = new(big.Int).SetUint64(1)
			constraints[i].Value = new(big.Int).SetUint64(model.HashStringToInt(stringValue))
		case constant.INT:

			intValue, ok := statement[i].Value.(int)
			if !ok {
				if reflect.TypeOf(statement[i].Value).Name() == "float64" {
					floatValue, _ := statement[i].Value.(float64)
					intValue = int(floatValue)
				} else {
					return nil, (fmt.Errorf("variable %s not of int type", statement[i].Attribute))
				}
			}
			constraints[i].Type = new(big.Int).SetUint64(0)
			constraints[i].Value = new(big.Int).SetUint64(uint64(intValue))
		case constant.BOOL:
			boolValue, ok := statement[i].Value.(bool)
			if !ok {
				return nil, (fmt.Errorf("variable %s not of bool type", statement[i].Attribute))
			}
			constraints[i].Type = new(big.Int).SetUint64(2)
			if boolValue {
				constraints[i].Value = new(big.Int).SetUint64(1)
			} else {
				constraints[i].Value = new(big.Int).SetUint64(0)
			}
		}

		switch statement[i].Operation {
		case constant.EQUAL:
			constraints[i].Operation = new(big.Int).SetUint64(0)
		case constant.NOTEQUAL:
			constraints[i].Operation = new(big.Int).SetUint64(1)
		case constant.LESSERTHAN:
			constraints[i].Operation = new(big.Int).SetUint64(2)
		case constant.GREATERTHANEQUALTO:
			constraints[i].Operation = new(big.Int).SetUint64(3)
		case constant.GREATERTHAN:
			constraints[i].Operation = new(big.Int).SetUint64(4)
		case constant.LESSERTHANEQUALTO:
			constraints[i].Operation = new(big.Int).SetUint64(5)
		}
	}
	return constraints, nil
}

func ConvertSignature(signature eddsaCrypto.Signature) eddsa.Signature {
	var sFr fr.Element
	sFr.SetBytes(signature.S[:])
	sBigInt := sFr.BigInt(new(big.Int))

	return eddsa.Signature{
		R: twistededwards.Point{
			X: signature.R.X.BigInt(new(big.Int)),
			Y: signature.R.Y.BigInt(new(big.Int)),
		},
		S: sBigInt,
	}
}

func ConvertPublicKey(publicKey eddsaCrypto.PublicKey) eddsa.PublicKey {
	return eddsa.PublicKey{
		A: twistededwards.Point{
			X: publicKey.A.X.BigInt(new(big.Int)),
			Y: publicKey.A.Y.BigInt(new(big.Int)),
		},
	}
}

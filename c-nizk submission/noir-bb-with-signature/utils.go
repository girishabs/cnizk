package main

import (
	"crypto/ecdsa"
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

func CreateTmpDir() (string, string) {

	id := uuid.New().String()

	workerDir := filepath.Join("tmp", id)
	circuitDir := filepath.Join(workerDir, "proofsystem")

	os.MkdirAll(circuitDir, 0755)

	copyDir("proofsystem", circuitDir)

	return workerDir, circuitDir
}

func CreateWorkDir() string {

	id := uuid.New().String()

	dir := filepath.Join("tmp", id)

	err := os.MkdirAll(dir, 0755)
	if err != nil {
		panic(err)
	}

	return dir
}

func copyDir(src, dst string) {

	filepath.Walk(src, func(path string, info os.FileInfo, err error) error {

		rel, _ := filepath.Rel(src, path)
		target := filepath.Join(dst, rel)

		if info.IsDir() {
			os.MkdirAll(target, 0755)
			return nil
		}

		data, _ := os.ReadFile(path)
		os.WriteFile(target, data, 0644)

		return nil
	})
}

func writeBase64(path string, data string) {

	decoded, err := base64.StdEncoding.DecodeString(data)
	if err != nil {
		panic(err)
	}

	err = os.WriteFile(path, decoded, 0644)
	if err != nil {
		panic(err)
	}
}

func GetPubKeyXY() ([]byte, []byte) {
	pub := privateKey.Public().(*ecdsa.PublicKey)

	x := pub.X.Bytes()
	y := pub.Y.Bytes()

	xPadded := make([]byte, 32)
	yPadded := make([]byte, 32)

	copy(xPadded[32-len(x):], x)
	copy(yPadded[32-len(y):], y)

	return xPadded, yPadded
}

func WriteProverToml(job ProveJob) error {

	file := "proofsystem/Prover.toml"

	arrStrings := func(b []byte) string {
		out := "["
		for i, v := range b {
			out += fmt.Sprintf("\"%d\"", v)
			if i < len(b)-1 {
				out += ", "
			}
		}
		out += "]"
		return out
	}

	arrIntsAsStrings := func(a []int) string {
		out := "["
		for i, v := range a {
			out += fmt.Sprintf("\"%d\"", v)
			if i < len(a)-1 {
				out += ", "
			}
		}
		out += "]"
		return out
	}

	toml := ""

	toml += fmt.Sprintf("creation_timestamp = \"%d\"\n\n", job.CreationTimestamp)

	toml += fmt.Sprintf("data = %s\n\n",
		arrIntsAsStrings(job.Data))

	toml += fmt.Sprintf("message = %s\n",
		arrStrings(job.Message))

	toml += fmt.Sprintf("pub_key_x = %s\n",
		arrStrings(job.PubKeyX))

	toml += fmt.Sprintf("pub_key_y = %s\n",
		arrStrings(job.PubKeyY))

	toml += fmt.Sprintf("signature = %s\n\n",
		arrStrings(job.Signature))

	for _, c := range job.Constraints {

		toml += "[[constraints]]\n"
		toml += fmt.Sprintf("attribute_type = \"%d\"\n", c.AttributeType)
		toml += fmt.Sprintf("operation = \"%d\"\n", c.Operation)
		toml += fmt.Sprintf("value = \"%d\"\n\n", c.Value)
	}

	return os.WriteFile(file, []byte(toml), 0644)
}

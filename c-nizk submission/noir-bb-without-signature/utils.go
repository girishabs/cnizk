package main

import (
	"encoding/base64"
	"fmt"
	"os"
	"path/filepath"

	"github.com/google/uuid"
)

import "encoding/binary"

func BuildCircuitPayload(data []int, timestamp int32, randomness []byte) []byte {

	payload := make([]byte, 40)

	for i := 0; i < 5; i++ {
		binary.BigEndian.PutUint32(payload[i*4:], uint32(data[i]))
	}

	binary.BigEndian.PutUint32(payload[20:], uint32(timestamp))

	copy(payload[24:], randomness)

	return payload
}

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
	return pubKey[1:33], pubKey[33:65]
}

func WriteProverToml(circuitDir string, job ProveJob) error {

	arrInt := func(a []int) string {
		out := "["
		for i, v := range a {
			out += fmt.Sprintf("%d", v)
			if i < len(a)-1 {
				out += ", "
			}
		}
		out += "]"
		return out
	}

	arrByte := func(a []byte) string {
		out := "["
		for i, v := range a {
			out += fmt.Sprintf("%d", v)
			if i < len(a)-1 {
				out += ", "
			}
		}
		out += "]"
		return out
	}

	toml := ""

	toml += fmt.Sprintf("creation_timestamp = %d\n\n", job.CreationTimestamp)
	toml += fmt.Sprintf("data = %s\n\n", arrInt(job.Data))
	toml += fmt.Sprintf("randomness = %s\n\n", arrByte(job.Randomness))
	toml += fmt.Sprintf("message = %s\n\n", arrByte(job.Message))

	for _, c := range job.Constraints {
		toml += "[[constraints]]\n"
		toml += fmt.Sprintf("attribute_type = %d\n", c.AttributeType)
		toml += fmt.Sprintf("operation = %d\n", c.Operation)
		toml += fmt.Sprintf("value = %d\n\n", c.Value)
	}

	proverPath := filepath.Join(circuitDir, "Prover.toml")
	return os.WriteFile(proverPath, []byte(toml), 0644)
}
package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sort"
	"strconv"
	"time"
)

const (
	NOIR_PROVE_URL  = "%v/proof/gen"
	NOIR_VERIFY_URL = "%v/proof/verify"

	GNARK_PROVE_URL  = "%v/prove"
	GNARK_VERIFY_URL = "%v/verify"

	NATIVE_PROVE_URL  = "%v/proof/prove"
	NATIVE_VERIFY_URL = "%v/proof/verify"
)

const (
	NOIR_PROOF_FILE   = "noir_verify.json"
	GNARK_PROOF_FILE  = "gnark_verify.json"
	NATIVE_PROOF_FILE = "native_verify.json"
)

func getEnvOrDefault(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvIntOrDefault(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func makeNoirProvePayload() map[string]any {
	return map[string]any{
		"circuitName": "proofsystem",
		"data":        []int{10, 20, 30, 40, 50},
		"creationTimestamp": time.Now().Unix(),
		"constraints": []map[string]any{
			{"operation": 0, "value": 10, "attribute_type": 0},
			{"operation": 2, "value": 25, "attribute_type": 0},
			{"operation": 4, "value": 10, "attribute_type": 0},
			{"operation": 5, "value": 40, "attribute_type": 0},
			{"operation": 1, "value": 99, "attribute_type": 0},
		},
	}
}

func makeGnarkProvePayload() map[string]any {
	return map[string]any{
		"asset": map[string]any{
			"owner":   "abcd",
			"version": "1.0.0",
			"type":    "cbdc",
			"value":   300000,
			"status":  true,
		},
		"statement": []map[string]any{
			{"attribute": "Owner", "type": "string", "operation": "eq", "value": "abcd"},
			{"attribute": "Status", "type": "bool", "operation": "eq", "value": true},
			{"attribute": "Value", "type": "int", "operation": "eq", "value": 300000},
			{"attribute": "Type", "type": "string", "operation": "eq", "value": "cbdc"},
			{"attribute": "Version", "type": "string", "operation": "eq", "value": "1.0.0"},
		},
		"creationTimestamp": time.Now().Unix(),
	}
}

var transport = &http.Transport{
	MaxIdleConns:        10000,
	MaxIdleConnsPerHost: 10000,
	IdleConnTimeout:     30 * time.Minute,
	TLSHandshakeTimeout: 2 * time.Minute,
	DisableKeepAlives:   false,
}

var client = &http.Client{
	Timeout:   20 * time.Minute,
	Transport: transport,
}

func postJSON(url string, payload any) ([]byte, int, error) {
	b, _ := json.Marshal(payload)

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Minute)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(b))
	if err != nil {
		return nil, 0, err
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Connection", "keep-alive")

	resp, err := client.Do(req)
	if err != nil {
		return nil, 0, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	return body, resp.StatusCode, nil
}

func storeProofOnce(url, file string, payload any) {
	fmt.Println("Generating one proof for verify...")
	body, _, err := postJSON(url, payload)
	if err != nil {
		log.Fatal(err)
	}
	os.WriteFile(file, body, 0644)
	fmt.Println("Saved", file)
}

func runSequential(url string, payload []byte, total int) {
	var latencies []float64
	var sum float64

	for i := 0; i < total; i++ {
		start := time.Now()

		_, status, err := postJSON(url, json.RawMessage(payload))
		if err != nil || status >= 300 {
			fmt.Println("failed:", err, status)
			continue
		}

		lat := time.Since(start).Seconds() * 1000
		latencies = append(latencies, lat)
		sum += lat

		fmt.Printf("Txn %d -> %.2f ms\n", i+1, lat)
	}

	printStats(latencies, sum)
}

func printStats(data []float64, sum float64) {
	sort.Float64s(data)

	fmt.Println("\n====== Results ======")
	fmt.Println("Count:", len(data))
	fmt.Printf("Mean : %.2f ms\n", sum/float64(len(data)))
	fmt.Printf("P50  : %.2f ms\n", pct(data, 50))
	fmt.Printf("P90  : %.2f ms\n", pct(data, 90))
	fmt.Printf("P95  : %.2f ms\n", pct(data, 95))
	fmt.Printf("P99  : %.2f ms\n", pct(data, 99))
}

func pct(d []float64, p int) float64 {
	i := int(float64(p) / 100 * float64(len(d)-1))
	return d[i]
}

func main() {
	mode := getEnvOrDefault("MODE", "gnark-prove")
	server := os.Getenv("SERVER_URL")
	total := getEnvIntOrDefault("TOTAL_TXNS", 1)

	if server == "" {
		log.Fatal("SERVER_URL not set")
	}

	fmt.Println("MODE =", mode)
	fmt.Println("TXNS =", total)

	noirProve := fmt.Sprintf(NOIR_PROVE_URL, server)
	noirVerify := fmt.Sprintf(NOIR_VERIFY_URL, server)
	gnarkProve := fmt.Sprintf(GNARK_PROVE_URL, server)
	gnarkVerify := fmt.Sprintf(GNARK_VERIFY_URL, server)
	nativeProve := fmt.Sprintf(NATIVE_PROVE_URL, server)
	nativeVerify := fmt.Sprintf(NATIVE_VERIFY_URL, server)

	switch mode {

	case "noir-prove":
		payload, _ := json.Marshal(makeNoirProvePayload())
		runSequential(noirProve, payload, total)

	case "noir-verify":
		if _, err := os.Stat(NOIR_PROOF_FILE); os.IsNotExist(err) {
			storeProofOnce(noirProve, NOIR_PROOF_FILE, makeNoirProvePayload())
		}
		payload, _ := os.ReadFile(NOIR_PROOF_FILE)
		runSequential(noirVerify, payload, total)

	case "gnark-prove":
		payload, _ := json.Marshal(makeGnarkProvePayload())
		runSequential(gnarkProve, payload, total)

	case "gnark-verify":
		if _, err := os.Stat(GNARK_PROOF_FILE); os.IsNotExist(err) {
			storeProofOnce(gnarkProve, GNARK_PROOF_FILE, makeGnarkProvePayload())
		}
		payload, _ := os.ReadFile(GNARK_PROOF_FILE)
		runSequential(gnarkVerify, payload, total)

	case "native-prove":
		payload, _ := json.Marshal(makeNoirProvePayload())
		runSequential(nativeProve, payload, total)

	case "native-verify":
		if _, err := os.Stat(NATIVE_PROOF_FILE); os.IsNotExist(err) {
			storeProofOnce(nativeProve, NATIVE_PROOF_FILE, makeNoirProvePayload())
		}
		payload, _ := os.ReadFile(NATIVE_PROOF_FILE)
		runSequential(nativeVerify, payload, total)

	default:
		log.Fatal("Unknown MODE")
	}
}
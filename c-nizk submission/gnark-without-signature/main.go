package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"encoding/base64"

	"zk-snarks/circuit"
	"zk-snarks/model"
	"zk-snarks/prover"
	"zk-snarks/verifier"
)

func mustB64(s string) []byte {
	if s == "" {
		return nil
	}
	b, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		panic("invalid base64: " + err.Error())
	}
	return b
}

var (
	ccs, pk, vk []byte
	mp          = make(map[int64]int64)
	err         error
)

func GenerateCircuitFiles() {
	ccs, pk, vk := circuit.GenerateProof()

	if err := os.WriteFile("ccs.txt", ccs, 0644); err != nil {
		log.Fatalf("Failed to write ccs.txt: %v", err)
	}

	if err := os.WriteFile("pk.txt", pk, 0644); err != nil {
		log.Fatalf("Failed to write pk.txt: %v", err)
	}

	if err := os.WriteFile("vk.txt", vk, 0644); err != nil {
		log.Fatalf("Failed to write vk.txt: %v", err)
	}

	log.Println("Circuit files generated successfully!")
}

func main() {
	if len(os.Args) > 1 && os.Args[1] == "regenerate" {
		GenerateCircuitFiles()
		return
	}

	runtime.GOMAXPROCS(runtime.NumCPU())

	ccs, err = os.ReadFile("ccs.txt")
	if err != nil {
		log.Fatalf("ccs write error: %v", err)
	}

	pk, err = os.ReadFile("pk.txt")
	if err != nil {
		log.Fatalf("pk write error: %v", err)
	}

	vk, err = os.ReadFile("vk.txt")
	if err != nil {
		log.Fatalf("vk write error: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("/prove", proveOnlyHandler)
	mux.HandleFunc("/verify", verifyOnlyHandler)

	srv := &http.Server{
		Addr:              "0.0.0.0:8080",
		Handler:           mux,
		ReadTimeout:       10 * time.Minute,
		WriteTimeout:      10 * time.Minute,
		IdleTimeout:       10 * time.Minute,
		ReadHeaderTimeout: 5 * time.Minute,
		MaxHeaderBytes:    1 << 20,
	}

	log.Printf("Server listening on %s", srv.Addr)

	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen error: %v", err)
		}
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop
	log.Println("Shutting down server...")

	var total int64
	for k, v := range mp {
		fmt.Sprintf("Start Time: %v, Duration: %v", k, v)
		total += v
	}
	fmt.Printf("Total time: %v\n", total)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx)
	log.Println("Server exited gracefully")
}

func proveOnlyHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	var req model.AssetRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json", http.StatusBadRequest)
		return
	}

	proof := prover.Prove(req.AssetDetails, req.StatementDetails, req.CreationTimestamp, ccs, pk, mp)

	resp := map[string]any{
		"proof": proof,
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)

}

type VerifyRequest struct {
	Proof struct {
		Proof             string `json:"Proof"`
		Hash              string `json:"Hash"`
		Signature         string `json:"Signature"`
		PublicKey         string `json:"PublicKey"`
		CreationTimestamp int64  `json:"CreationTimestamp"`
	} `json:"proof"`
}

func verifyOnlyHandler(w http.ResponseWriter, r *http.Request) {

	w.Header().Set("Content-Type", "application/json")

	if r.Method != http.MethodPost {
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}

	body, err := io.ReadAll(r.Body)
	if err != nil || len(body) == 0 {
		http.Error(w, "invalid body", http.StatusBadRequest)
		return
	}

	var req VerifyRequest
	if err := json.Unmarshal(body, &req); err != nil {
		http.Error(w, "invalid json: "+err.Error(), http.StatusBadRequest)
		return
	}

	payload := model.ProverPayload{
		Proof:             mustB64(req.Proof.Proof),
		Hash:              mustB64(req.Proof.Hash),
		Signature:         mustB64(req.Proof.Signature),
		PublicKey:         mustB64(req.Proof.PublicKey),
		CreationTimestamp: req.Proof.CreationTimestamp,
	}

	err = verifier.Verify(payload, vk)

	ok := err == nil

	resp := map[string]any{
		"verified": ok,
	}

	w.WriteHeader(http.StatusOK)
	_ = json.NewEncoder(w).Encode(resp)

}

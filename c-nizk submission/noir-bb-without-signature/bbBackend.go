package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"time"
)

func RunProver(circuitDir string, job ProveJob) ProveResult {

	err := WriteProverToml(circuitDir, job)
	if err != nil {
		return ProveResult{Err: err}
	}

	workerDir := filepath.Dir(circuitDir)
	outDir := filepath.Join(workerDir, "out")

	os.MkdirAll(outDir, 0755)

	cmd := exec.Command("nargo", "execute")
	cmd.Dir = circuitDir
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()

	if err := cmd.Run(); err != nil {
		return ProveResult{Err: err}
	}

	elapsed := time.Since(start)
	fmt.Printf("time taken for witness generation: %v\n", elapsed)

	witness := filepath.Join(
		circuitDir,
		"target",
		"proofsystem.gz",
	)

	cmd = exec.Command(
		"bb",
		"prove",
		"-b", filepath.Join(circuitDir, "target", "proofsystem.json"),
		"-w", witness,
		"--write_vk",
		"-o", outDir,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start = time.Now()

	if err := cmd.Run(); err != nil {
		return ProveResult{Err: err}
	}

	elapsed = time.Since(start)
	fmt.Printf("time taken for proof generation: %v\n", elapsed)

	proof, err := os.ReadFile(filepath.Join(outDir, "proof"))
	if err != nil {
		return ProveResult{Err: err}
	}

	vk, err := os.ReadFile(filepath.Join(outDir, "vk"))
	if err != nil {
		return ProveResult{Err: err}
	}

	pub, err := os.ReadFile(filepath.Join(outDir, "public_inputs"))
	if err != nil {
		return ProveResult{Err: err}
	}

	return ProveResult{
		Proof:        proof,
		VK:           vk,
		PublicInputs: pub,
	}
}

func VerifyProof(proofB64, vkB64, pubB64 string) (bool, error) {

	workDir := CreateWorkDir()
	defer os.RemoveAll(workDir)

	proofPath := filepath.Join(workDir, "proof")
	vkPath := filepath.Join(workDir, "vk")
	pubPath := filepath.Join(workDir, "public_inputs")

	writeBase64(proofPath, proofB64)
	writeBase64(vkPath, vkB64)
	writeBase64(pubPath, pubB64)

	cmd := exec.Command(
		"bb",
		"verify",
		"-p", proofPath,
		"-k", vkPath,
		"-i", pubPath,
	)

	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr

	start := time.Now()

	err := cmd.Run()
	if err != nil {
		return false, err
	}

	elapsed := time.Since(start)
	fmt.Printf("time taken for proof verification: %v\n", elapsed)
	
	return true, nil
}

package main

import (
	"log"
	"os"

	"zk-snarks/circuit"
)

func main() {
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

package main

import (
	"log"
	"os"
)

func main() {

	InitSigner()

	// start worker pool
	StartWorkers(4) // tune = CPU cores

	router := SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	log.Println("Server running on", port)
	router.Run(":" + port)
}
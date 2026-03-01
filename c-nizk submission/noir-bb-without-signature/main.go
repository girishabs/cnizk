package main

import (
	"log"
	"os"
)

func main() {

	InitSigner()
	StartWorkers(4) 

	router := SetupRouter()

	port := os.Getenv("PORT")
	if port == "" {
		port = "4000"
	}

	log.Println("Server running on", port)
	router.Run(":" + port)
}
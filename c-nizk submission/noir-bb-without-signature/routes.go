package main

import "github.com/gin-gonic/gin"

func SetupRouter() *gin.Engine {

	r := gin.Default()

	proof := r.Group("/proof")
	{
		proof.POST("/prove", ProveHandler)
		proof.POST("/verify", VerifyHandler)
	}

	return r
}
package main

import (
	"encoding/base64"
	"encoding/binary"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/blake2s"
)

func ProveHandler(c *gin.Context) {
	var req ProveRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	if len(req.Data) != 5 {
		c.JSON(400, gin.H{"error": "data must have length 5"})
		return
	}

	creationTimestamp := uint32(time.Now().Unix())

	payload := make([]byte, 24)

	for i, v := range req.Data {
		binary.BigEndian.PutUint32(payload[i*4:(i+1)*4], uint32(v))
	}

	binary.BigEndian.PutUint32(payload[20:24], creationTimestamp)

	// -----------------------------------------
	// Blake2s hash (32 bytes)
	// -----------------------------------------

	hash := blake2s.Sum256(payload)

	start := time.Now()

	// Sign the 32-byte hash directly
	sig := SignHash(hash[:])

	elapsed := time.Since(start)
	fmt.Println("signing time:", elapsed)

	px, py := GetPubKeyXY()

	job := ProveJob{
		Data:              req.Data,
		Constraints:       req.Constraints,
		CreationTimestamp: int32(creationTimestamp),
		Message:           hash[:], // 32 bytes
		Signature:         sig,     // 64 bytes
		PubKeyX:           px,
		PubKeyY:           py,
	}

	result := SubmitProveJob(job)

	if result.Err != nil {
		c.JSON(500, gin.H{"error": result.Err.Error()})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"success":      true,
		"proof":        base64.StdEncoding.EncodeToString(result.Proof),
		"vk":           base64.StdEncoding.EncodeToString(result.VK),
		"publicInputs": base64.StdEncoding.EncodeToString(result.PublicInputs),
	})
}

func VerifyHandler(c *gin.Context) {

	var body struct {
		Proof        string `json:"proof"`
		VK           string `json:"vk"`
		PublicInputs string `json:"publicInputs"`
	}

	c.BindJSON(&body)

	ok, err := VerifyProof(body.Proof, body.VK, body.PublicInputs)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"valid":   ok,
	})
}

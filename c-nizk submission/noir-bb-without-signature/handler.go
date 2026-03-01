package main

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func ProveHandler(c *gin.Context) {

	var req ProveRequest

	if err := c.BindJSON(&req); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	creationTimestamp := time.Now().Unix()

	randomness := make([]byte, 16)
	rand.Read(randomness)

	payload := BuildCircuitPayload(
		req.Data,
		int32(creationTimestamp),
		randomness,
	)

	start := time.Now()

	hash, sig := SignMessage(payload)

	elapsed := time.Since(start)
	fmt.Println("signing time: ", elapsed)

	job := ProveJob{
		Data:              req.Data,
		Constraints:       req.Constraints,
		CreationTimestamp: int32(creationTimestamp),
		Message:           hash,
		Randomness:        randomness,
	}

	result := SubmitProveJob(job)

	if result.Err != nil {
		c.JSON(500, gin.H{"error": result.Err.Error()})
		return
	}

	px, py := GetPubKeyXY()

	c.JSON(http.StatusOK, gin.H{
		"success": true,

		"proof":        base64.StdEncoding.EncodeToString(result.Proof),
		"vk":           base64.StdEncoding.EncodeToString(result.VK),
		"publicInputs": base64.StdEncoding.EncodeToString(result.PublicInputs),

		"message":   base64.StdEncoding.EncodeToString(hash),
		"signature": base64.StdEncoding.EncodeToString(sig),

		"pub_key_x": base64.StdEncoding.EncodeToString(px),
		"pub_key_y": base64.StdEncoding.EncodeToString(py),
	})
}

func VerifyHandler(c *gin.Context) {

	var body struct {
		Proof        string `json:"proof"`
		VK           string `json:"vk"`
		PublicInputs string `json:"publicInputs"`

		Message   string `json:"message"`
		Signature string `json:"signature"`
		PubKeyX   string `json:"pub_key_x"`
		PubKeyY   string `json:"pub_key_y"`
	}

	if err := c.BindJSON(&body); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	okSig := VerifyAttestation(
		body.Message,
		body.Signature,
		body.PubKeyX,
		body.PubKeyY,
	)

	if !okSig {
		c.JSON(401, gin.H{"error": "invalid signature"})
		return
	}

	ok, err := VerifyProof(
		body.Proof,
		body.VK,
		body.PublicInputs,
	)

	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{
		"success": true,
		"valid":   ok,
	})
}

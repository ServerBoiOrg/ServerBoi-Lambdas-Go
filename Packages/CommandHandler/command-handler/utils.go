package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

func getEnvVar(key string) string {

	env := godotenv.Load(".env")

	if env == nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

func decodeToPublicKey(applicationPublicKey string) ed25519.PublicKey {
	rawKey := []byte(applicationPublicKey)
	byteKey := make([]byte, hex.DecodedLen(len(rawKey)))
	_, _ = hex.Decode(byteKey, rawKey)
	return byteKey
}

func generateWorkflowUUID(workflowName string) string {
	uuidWithHyphen := uuid.New()
	uuidString := strings.Replace(uuidWithHyphen.String(), "-", "", -1)
	subuuid := uuidString[0:8]
	workflowID := fmt.Sprintf("%v-%v", workflowName, subuuid)
	return workflowID
}

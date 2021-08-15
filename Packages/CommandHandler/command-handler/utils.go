package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"errors"
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/joho/godotenv"
)

var (
	awsRegions       = []string{"us-west-2"}
	serverboiRegions = []string{"us-west"}
)

func getEnvVar(key string) string {

	env := godotenv.Load(".env")

	if env == nil {
		log.Fatalf("Error loading .env file")
	}

	return os.Getenv(key)
}

// Function to check if given service is supported for Serverboi
func verifyService(s string) error {
	service := strings.ToLower(s)

	switch service {
	case "aws":
		return nil
	case "linode":
		return nil
	default:
		return errors.New(fmt.Sprintf("* service: Unknown service `%v`", s))
	}
}

// Verifies region is a valid region for the service
func verifyRegion(s string, r string) error {
	service := strings.ToLower(s)
	region := strings.ToLower(r)

	switch service {
	case "aws":
		return verifyAWSRegion(region)
	case "linode":
		return nil
	default:
		return errors.New("* region: Valid service is required to check region")
	}
}

// Verifies the provided region is either an actual AWS region or a Serverboi Logical regions
func verifyAWSRegion(region string) error {
	for _, awsRegion := range awsRegions {
		if region == awsRegion {
			return nil
		}
	}

	for _, serverboiRegion := range awsRegions {
		if region == serverboiRegion {
			return nil
		}
	}

	return errors.New(fmt.Sprintf("* region: `%v` is not an AWS Region or ServerBoi Region", region))
}

func formInvalidParametersResponse(errors []string) DiscordInteractionResponseData {
	message := "Command parameters had the following errors:"
	for _, errorMessage := range errors {
		message = fmt.Sprintf("%v\n%v", message, errorMessage)
	}
	formRespInput := FormResponseInput{
		"Content": message,
	}
	return formResponseData(formRespInput)
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

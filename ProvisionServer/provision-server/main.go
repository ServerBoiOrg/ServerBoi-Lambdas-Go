package main

import (
	"encoding/json"
	"errors"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
)

var (
	// DefaultHTTPGetAddress Default Address
	DefaultHTTPGetAddress = "https://checkip.amazonaws.com"

	// ErrNoIP No IP found in response
	ErrNoIP = errors.New("No IP in HTTP response")

	// ErrNon200Response non 200 status code in response
	ErrNon200Response = errors.New("Non 200 Response found")
)

type ProvisonServerParameters struct {
	Application      string
	Service          string
	OwnerID          string
	Owner            string
	InteractionID    string
	InteractionToken string
	ApplicationID    string
	GuildID          string
	Url              string
	CreationOptions  map[string]string
}

func handler(event map[string]string) (response map[string]string, err error) {
	params := convertEvent(event)

	switch params.Service {
	case "aws":
		//
	case "linode":
		//
	case "vultr":

	}

	return response, nil

}

func convertEvent(event map[string]string) (params ProvisonServerParameters) {
	jsoned, _ := json.Marshal(event)
	params = ProvisonServerParameters{}
	if marshalErr := json.Unmarshal(jsoned, &params); marshalErr != nil {
		log.Fatal(marshalErr)
		panic(marshalErr)
	}
	return params
}

func main() {
	lambda.Start(handler)
}

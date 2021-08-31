package main

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"

	gu "generalutils"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/sfn"
)

var TOKEN_BUCKET = gu.GetEnvVar("TOKEN_BUCKET")

type BootstrapCallPayload struct {
	ApplicationID    string `json:"application_id"`
	ExecutionName    string `json:"execution_name"`
	InteractionToken string `json:"interaction_token"`
	Port             string `json:"port"`
	ServerID         string `json:"server_id"`
	GuildID          string `json:"guild_id"`
}

func handler(event map[string]interface{}) (bool, error) {
	log.Printf("Event: %v", event)
	params := convertEvent(event)
	s3Client := gu.GetS3Client()
	sfnClient := gu.CreateSfnClient()

	response, _ := s3Client.GetObject(context.Background(), &s3.GetObjectInput{
		Bucket: aws.String(TOKEN_BUCKET),
		Key:    aws.String(params.ExecutionName),
	})
	defer response.Body.Close()

	bytes, err := ioutil.ReadAll(response.Body)
	if err != nil {
		panic(err)
	}
	token := string(bytes)
	log.Printf("Token: %v", token)

	sfnClient.SendTaskSuccess(context.Background(), &sfn.SendTaskSuccessInput{
		Output:    aws.String(convertParametersToString(params)),
		TaskToken: aws.String(token),
	})

	return true, nil
}

func main() {
	lambda.Start(handler)
}

func convertEvent(event map[string]interface{}) (params BootstrapCallPayload) {
	jsoned, _ := json.Marshal(event)
	params = BootstrapCallPayload{}
	if marshalErr := json.Unmarshal(jsoned, &params); marshalErr != nil {
		log.Fatal(marshalErr)
		panic(marshalErr)
	}
	return params
}

func convertParametersToString(parameters BootstrapCallPayload) string {
	jsoned, err := json.Marshal(parameters)
	if err != nil {
		log.Fatal(err)
	}
	return string(jsoned)
}

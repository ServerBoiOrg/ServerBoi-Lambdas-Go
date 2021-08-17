package main

import (
	"context"
	"encoding/json"
	"log"
	"strings"

	gu "generalutils"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

var TOKEN_BUCKET = gu.GetEnvVar("TOKEN_BUCKET")

type PutTokenPayload struct {
	TaskToken     string `json:"TaskToken"`
	ExecutionName string `json:"ExecutionName"`
}

func handler(event map[string]interface{}) bool {
	log.Printf("Event: %v", event)
	params := convertEvent(event)

	client := gu.GetS3Client()

	client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: &TOKEN_BUCKET,
		Body:   strings.NewReader(params.TaskToken),
		Key:    &params.ExecutionName,
	})

	return true
}

func main() {
	lambda.Start(handler)
}

func convertEvent(event map[string]interface{}) (params PutTokenPayload) {
	jsoned, _ := json.Marshal(event)
	params = PutTokenPayload{}
	if marshalErr := json.Unmarshal(jsoned, &params); marshalErr != nil {
		log.Fatal(marshalErr)
		panic(marshalErr)
	}
	return params
}

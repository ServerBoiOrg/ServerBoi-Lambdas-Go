package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/cloudwatch"
	cwtypes "github.com/aws/aws-sdk-go-v2/service/cloudwatch/types"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
)

type ProvisonServerParameters struct {
	ExecutionName    string            `json:"ExecutionName"`
	Application      string            `json:"Application"`
	Service          string            `json:"Service"`
	OwnerID          string            `json:"OwnerID"`
	Owner            string            `json:"Owner"`
	InteractionID    string            `json:"InteractionID"`
	InteractionToken string            `json:"InteractionToken"`
	ApplicationID    string            `json:"ApplicationID"`
	GuildID          string            `json:"GuildID"`
	Url              string            `json:"Url"`
	ServerName       string            `json:"ServerName"`
	CreationOptions  map[string]string `json:"CreationOptions,omitempty"`
}

func handler(event map[string]interface{}) (map[string]interface{}, error) {
	log.Printf("Event: %v", event)
	params := convertEvent(event)
	// logMetric(params.Service)
	// logMetric(params.Application)

	var serverItem map[string]dynamotypes.AttributeValue

	log.Printf("Cloud Provider for server: %v", params.Service)
	switch params.Service {
	case "aws":
		serverItem = provisionAWS(params)
	case "linode":
		//
	case "vultr":

	}

	serverID := writeServerInfo(serverItem)
	event["ServerID"] = serverID

	return event, nil
}

func writeServerInfo(serverItem map[string]dynamotypes.AttributeValue) string {
	dynamo := getDynamo()
	table := getEnvVar("SERVER_TABLE")
	var serverID string

	n := 0
	for n < 10 {
		log.Printf("Putting server item in table %v. Attempt: %v", table, (n + 1))
		serverID = formServerID()

		conditional := aws.String("attribute_not_exists(ServerID)")
		serverItem["ServerID"] = &dynamotypes.AttributeValueMemberS{Value: serverID}
		_, err := dynamo.PutItem(context.Background(), &dynamodb.PutItemInput{
			TableName:           aws.String(table),
			Item:                serverItem,
			ConditionExpression: conditional,
		})
		if err == nil {
			break
		} else {
			log.Printf("Error putting item: %v", err)
		}
		n++
	}

	return serverID
}

func logMetric(metricName string) {
	cw := getCloudwatchClient()
	namespace := "ServerBoi"
	value := float64(1)

	data := []cwtypes.MetricDatum{
		{
			MetricName: &metricName,
			Value:      &value,
		},
	}

	cw.PutMetricData(context.Background(), &cloudwatch.PutMetricDataInput{
		MetricData: data,
		Namespace:  &namespace,
	})
}

func convertEvent(event map[string]interface{}) (params ProvisonServerParameters) {
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

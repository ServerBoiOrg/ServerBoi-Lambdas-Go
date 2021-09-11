package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	gu "generalutils"

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
	OwnerID          string            `json:"OwnerID"`
	Owner            string            `json:"Owner"`
	InteractionID    string            `json:"InteractionID"`
	InteractionToken string            `json:"InteractionToken"`
	ApplicationID    string            `json:"ApplicationID"`
	GuildID          string            `json:"GuildID"`
	Url              string            `json:"Url"`
	CreationOptions  map[string]string `json:"CreationOptions,omitempty"`
	Service          string            `json:"Service"`
	Name             string            `json:"Name"`
	Region           string            `json:"Region"`
	HardwareType     string            `json:"HardwareType"`
	Private          bool              `json:"Private"`
	IsRole           bool              `json:"IsRole"`
}

type ProvisionServerResponse struct {
	ServerID   string `json:"ServerID"`
	InstanceID string `json:"InstanceID,omitempty"`
	AccountID  string `json:"AccountID,omitempty"`
}

func handler(event map[string]interface{}) (string, error) {
	log.Printf("Event: %v", event)
	params := convertEvent(event)
	// logMetric(params.Service)
	// logMetric(params.Application)
	embedInput := gu.FormWorkflowEmbedInput{
		Name:        "Provision-Server",
		Description: fmt.Sprintf("WorkflowID: %s", params.ExecutionName),
		Status:      "🟢 Running",
		Stage:       "Provisioning Server",
		Color:       gu.DiscordGreen,
	}
	embed := gu.FormWorkflowEmbed(embedInput)
	formRespInput := gu.FormResponseInput{
		"Embeds": embed,
	}
	gu.EditResponse(params.ApplicationID, params.InteractionToken, gu.FormResponseData(formRespInput))

	var serverID string
	var serverItem map[string]dynamotypes.AttributeValue
	log.Printf("Cloud Provider for server: %v", params.Service)
	switch params.Service {
	case "aws":
		serverID, serverItem = provisionAWS(params)
	case "linode":
		serverID, serverItem = provisionLinode(params)
	}

	writeServerInfo(serverItem)
	log.Printf("ServerID: %v", serverID)
	return serverID, nil
}

func writeServerInfo(serverItem map[string]dynamotypes.AttributeValue) {
	dynamo := gu.GetDynamo()
	table := gu.GetEnvVar("SERVER_TABLE")

	log.Printf("Putting server item in table %v", table)

	conditional := aws.String("attribute_not_exists(ServerID)")
	_, err := dynamo.PutItem(context.Background(), &dynamodb.PutItemInput{
		TableName:           aws.String(table),
		Item:                serverItem,
		ConditionExpression: conditional,
	})
	if err != nil {
		log.Printf("Error putting item: %v", err)
	}
}

func logMetric(metricName string) {
	cw := gu.GetCloudwatchClient()
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

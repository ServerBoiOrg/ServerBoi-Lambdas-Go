package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	dc "discordhttpclient"
	gu "generalutils"
	ru "responseutils"

	dt "github.com/awlsring/discordtypes"
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
	ClientPort       int               `json:"ClientPort"`
	QueryPort        int               `json:"QueryPort"`
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
	ServerID         string `json:"ServerID"`
	PrivateKeyObject string `json:"PrivateKeyObject"`
}

func handler(event map[string]interface{}) (string, error) {
	log.Printf("Event: %v", event)
	params := convertEvent(event)
	// logMetric(params.Service)
	// logMetric(params.Application)

	embed := ru.CreateWorkflowEmbed(&ru.CreateWorkflowEmbedInput{
		Name:        "Provision-Server",
		Description: fmt.Sprintf("WorkflowID: %s", params.ExecutionName),
		Status:      "🟢 Running",
		Stage:       "Provisioning Server",
		Color:       ru.DiscordGreen,
	})

	client := dc.CreateClient(&dc.CreateClientInput{
		ApiVersion: "v9",
	})

	for {
		resp, headers, err := client.EditInteractionResponse(&dc.InteractionFollowupInput{
			ApplicationID:    params.ApplicationID,
			InteractionToken: params.InteractionToken,
			Data: &dt.InteractionCallbackData{
				Embeds: []*dt.Embed{embed},
			},
		})
		if err != nil {
			log.Fatalf("Error getting creating message in Channel: %v", err)
		}
		done := dc.StatusCodeHandler(*headers)
		if done {
			log.Printf("%v", resp)
			break
		}
	}

	var output *ProvisionOutput
	log.Printf("Cloud Provider for server: %v", params.Service)
	switch params.Service {
	case "aws":
		output = provisionAWS(&params)
	case "linode":
		output = provisionLinode(&params)
	}

	writeServerInfo(output.ServerItem)
	response := &ProvisionServerResponse{
		ServerID:         output.ServerID,
		PrivateKeyObject: output.PrivateKeyObject,
	}
	b, _ := json.Marshal(response)
	return string(b), nil
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

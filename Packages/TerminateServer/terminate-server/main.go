package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	gu "generalutils"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go/aws"
)

var TOKEN_BUCKET = gu.GetEnvVar("TOKEN_BUCKET")

type TerminateServerPayload struct {
	ServerID      string `json:"ServerID"`
	Token         string `json:"Token"`
	ApplicationID string `json:"ApplicationID"`
	ExecutionName string `json:"ExecutionName"`
}

func handler(event map[string]interface{}) (bool, error) {
	log.Printf("Event: %v", event)
	params := convertEvent(event)

	updateEmbed(UpdateEmbedInput{
		params.ExecutionName,
		params.ApplicationID,
		params.Token,
		"🟢 Running",
		"Terminating",
		gu.DiscordGreen,
	})

	server, err := gu.GetServerFromID(params.ServerID)
	if err != nil {
		log.Fatalf("Error getting service object: %v", err)
	}
	log.Printf("Getting service")
	service := server.GetService()
	log.Printf("Service: %v", service)

	switch service {
	case "aws":
		awsServer, _ := server.(gu.AWSServer)
		client := gu.CreateEC2Client(awsServer.Region, awsServer.AWSAccountID)

		log.Printf("Creating instance")
		_, err := client.TerminateInstances(context.Background(),
			&ec2.TerminateInstancesInput{
				InstanceIds: []string{awsServer.InstanceID},
			},
		)
		if err != nil {
			log.Fatalf("Error deleting instance: %v", err)
		}
	case "linode":
		linodeServer, _ := server.(gu.LinodeServer)
		client := gu.CreateLinodeClient(linodeServer.ApiKey)
		err := client.DeleteInstance(context.Background(), linodeServer.LinodeID)
		if err != nil {
			log.Fatalf("Error deleting linode: %v", err)
		}
	}

	// Delete server item
	deleteServerItem(server.GetBaseService().ServerID)

	updateEmbed(UpdateEmbedInput{
		params.ExecutionName,
		params.ApplicationID,
		params.Token,
		"✔️ Finished",
		"Complete",
		gu.DarkGreen,
	})

	return true, nil
}

func main() {
	lambda.Start(handler)
}

func deleteServerItem(serverID string) {
	dynamo := gu.GetDynamo()
	table := gu.GetEnvVar("SERVER_TABLE")

	log.Printf("Delete server item in table %v", table)
	_, err := dynamo.DeleteItem(context.Background(), &dynamodb.DeleteItemInput{
		TableName: aws.String(table),
		Key: map[string]types.AttributeValue{
			"ServerID": &types.AttributeValueMemberS{Value: serverID},
		},
	})
	if err != nil {
		log.Printf("Error putting item: %v", err)
	}
}

type UpdateEmbedInput struct {
	ExecutionName    string
	ApplicationID    string
	InteractionToken string
	Status           string
	Stage            string
	Color            int
}

func updateEmbed(input UpdateEmbedInput) {
	embedInput := gu.FormWorkflowEmbedInput{
		Name:        "Terminate-Workflow",
		Description: fmt.Sprintf("WorkflowID: %s", input.ExecutionName),
		Status:      input.Status,
		Stage:       input.Stage,
		Color:       input.Color,
	}
	workflowEmbed := gu.FormWorkflowEmbed(embedInput)

	gu.EditResponse(
		input.ApplicationID,
		input.InteractionToken,
		gu.FormWorkflowResponseData(workflowEmbed),
	)
}

func convertEvent(event map[string]interface{}) (params TerminateServerPayload) {
	jsoned, _ := json.Marshal(event)
	params = TerminateServerPayload{}
	if marshalErr := json.Unmarshal(jsoned, &params); marshalErr != nil {
		log.Fatal(marshalErr)
		panic(marshalErr)
	}
	return params
}

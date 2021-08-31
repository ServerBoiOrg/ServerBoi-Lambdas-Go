package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"

	gu "generalutils"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

var (
	TOKEN        = gu.GetEnvVar("DISCORD_TOKEN")
	AWS_TABLE    = gu.GetEnvVar("AWS_TABLE")
	SERVER_TABLE = gu.GetEnvVar("SERVER_TABLE")
)

type FinishProvisonParameters struct {
	ServerID         string `json:"ServerID"`
	GuildID          string `json:"GuildID"`
	InteractionToken string `json:"InteractionToken"`
	ApplicationID    string `json:"ApplicationID"`
	ExecutionName    string `json:"ExecutionName"`
}

func handler(event map[string]interface{}) (bool, error) {
	log.Printf("Event: %v", event)
	params := convertEvent(event)

	workflowEmbed := gu.FormWorkflowEmbed(gu.FormWorkflowEmbedInput{
		Name:        "Provision-Server",
		Description: fmt.Sprintf("WorkflowID: %s", params.ExecutionName),
		Status:      "✔️ finished",
		Stage:       "Complete",
		Color:       5763719,
	})
	workflowEmbed.AddField("ServerID", params.ServerID)

	gu.EditResponse(
		params.ApplicationID,
		params.InteractionToken,
		gu.FormWorkflowResponseData(workflowEmbed),
	)

	server := gu.GetServerFromID(params.ServerID)

	service := server.GetService()
	var (
		ip    string
		state string
	)
	switch service {
	case "aws":
		awsServer, _ := server.(*gu.AWSServer)
		client := gu.CreateEC2Client(awsServer.Region, awsServer.AWSAccountID)
		response, err := client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
			InstanceIds: []string{
				awsServer.InstanceID,
			},
		})
		if err != nil {
			log.Fatalf("Error describing instance: %v", err)
		}

		ip = *response.Reservations[0].Instances[0].PublicIpAddress
		state = string(response.Reservations[0].Instances[0].State.Name)
	}

	serverInfo := server.GetBaseService()
	sbRegion := server.GetServerBoiRegion()

	embed := gu.FormServerEmbed(gu.FormServerEmbedInput{
		Name:        serverInfo.ServerName,
		ID:          serverInfo.ServerID,
		IP:          ip,
		Status:      state,
		Region:      sbRegion,
		Application: serverInfo.Application,
		Owner:       serverInfo.Owner,
		Service:     serverInfo.Service,
	})

	webookItem := gu.GetWebhookFromGuildID(params.GuildID)

	gu.PostToEmbedChannel(
		webookItem.WebhookID,
		webookItem.WebhookToken,
		gu.FormServerEmbedResponseData(embed, params.ServerID),
	)

	return true, nil
}

func main() {
	lambda.Start(handler)
}

func convertEvent(event map[string]interface{}) (params FinishProvisonParameters) {
	jsoned, _ := json.Marshal(event)
	params = FinishProvisonParameters{}
	if marshalErr := json.Unmarshal(jsoned, &params); marshalErr != nil {
		log.Fatal(marshalErr)
		panic(marshalErr)
	}
	return params
}

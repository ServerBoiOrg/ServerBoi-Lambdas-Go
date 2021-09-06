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
	TOKEN         = gu.GetEnvVar("DISCORD_TOKEN")
	SERVER_TABLE  = gu.GetEnvVar("SERVER_TABLE")
	CHANNEL_TABLE = gu.GetEnvVar("CHANNEL_TABLE")
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
		Color:       gu.DiscordGreen,
	})
	workflowEmbed.AddField("ServerID", params.ServerID)

	gu.EditResponse(
		params.ApplicationID,
		params.InteractionToken,
		gu.FormWorkflowResponseData(workflowEmbed),
	)

	server, err := gu.GetServerFromID(params.ServerID)
	if err != nil {
		log.Fatalf("Error getting service object: %v", err)
	}

	log.Printf("Getting service")
	service := server.GetService()
	var (
		ip    string
		state string
	)
	log.Printf("Service of server: %v", service)
	switch service {
	case "aws":
		awsServer, _ := server.(gu.AWSServer)
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
	case "linode":
		linodeServer, _ := server.(gu.LinodeServer)
		client := gu.CreateLinodeClient(linodeServer.ApiKey)

		linode, err := client.GetInstance(context.Background(), linodeServer.LinodeID)
		if err != nil {
			log.Fatalf("Error describing linode: %v", err)
		}

		ip = fmt.Sprintf("%v", linode.IPv4[0])
		state = string(linode.Status)
	}
	log.Printf("IP of server: %v", ip)
	log.Printf("State of server: %v", state)

	serverInfo := server.GetBaseService()
	sbRegion := server.GetServerBoiRegion()

	serverData := gu.GetServerEmbedData(gu.GetServerEmbedDataInput{
		Name:        serverInfo.ServerName,
		ID:          serverInfo.ServerID,
		IP:          ip,
		Status:      state,
		Region:      sbRegion,
		Port:        serverInfo.Port,
		Application: serverInfo.Application,
		Owner:       serverInfo.Owner,
		Service:     serverInfo.Service,
	})
	embed := gu.FormServerEmbed(serverData)

	log.Printf("Getting Channel for Guild")
	channelID, err := gu.GetChannelIDFromGuildID(params.GuildID)
	if err != nil {
		log.Fatalf("Error getting channelID from dynamo: %v", err)
	}

	client := gu.CreateDiscordClient(gu.CreateDiscordClientInput{
		BotToken:   TOKEN,
		ApiVersion: "v9",
	})
	log.Printf("Posting message")
	resp, err := client.CreateMessage(
		channelID,
		gu.FormServerEmbedResponseData(gu.FormServerEmbedResponseDataInput{
			ServerEmbed: embed,
			Running:     true,
		}))
	if err != nil {
		log.Fatalf("Error getting creating message in Channel: %v", err)
	}
	log.Printf("%v", resp)

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

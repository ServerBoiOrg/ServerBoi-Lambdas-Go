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
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go/aws"
)

var (
	TOKEN      = gu.GetEnvVar("DISCORD_TOKEN")
	KEY_BUCKET = gu.GetEnvVar("KEY_BUCKET")
	CLIENT     = dc.CreateClient(&dc.CreateClientInput{
		BotToken:   TOKEN,
		ApiVersion: "v9",
	})
)

type TerminateServerPayload struct {
	ServerID      string `json:"ServerID"`
	Token         string `json:"Token"`
	ApplicationID string `json:"ApplicationID"`
	ExecutionName string `json:"ExecutionName"`
	Fallback      bool   `json:"Fallback"`
	GuildID       string `json:"GuildID"`
}

func handler(event map[string]interface{}) (bool, error) {
	log.Printf("Event: %v", event)
	params := convertEvent(event)

	updateEmbed(&UpdateEmbedInput{
		params.ExecutionName,
		params.ApplicationID,
		params.Token,
		"🟢 Running",
		"Terminating",
		ru.DiscordGreen,
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

	deleteServerItem(server.GetBaseService().ServerID)
	deleteSSHKey(server.GetPrivateKey())
	deleteServerMessage(server, params.GuildID)
	var (
		status string
		color  int
	)
	if params.Fallback {
		status = "❌ Failed"
		color = ru.DiscordRed
	} else {
		status = "✔️ Finished"
		color = ru.DarkGreen
	}

	updateEmbed(&UpdateEmbedInput{
		params.ExecutionName,
		params.ApplicationID,
		params.Token,
		status,
		"Complete",
		color,
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

func deleteServerMessage(server gu.Server, guildID string) {
	channelID, err := gu.GetChannelIDFromGuildID(guildID)
	info := server.GetBaseService()
	var (
		messages []*dt.Message
		headers  *dc.DiscordHeaders
	)
	for {
		messages, headers, err = CLIENT.GetChannelMessages(channelID)
		if err != nil {
			log.Fatalf("Error getting messages for channel %v", channelID)
		}
		done := dc.StatusCodeHandler(*headers)
		if done {
			break
		}
	}

top:
	for _, message := range messages {
		for _, embed := range message.Embeds {
			if embed.Title == fmt.Sprintf("%v (%v)", info.ServerName, info.ServerID) {
				for {
					headers, err = CLIENT.DeleteMessage(&dc.DeleteMessageInput{
						ChannelID: channelID,
						MessageID: message.ID,
					})
					done := dc.StatusCodeHandler(*headers)
					if done {
						break top
					}
				}
			}

		}
	}
}

func deleteSSHKey(key string) {
	client := gu.GetS3Client()
	client.DeleteObject(context.Background(), &s3.DeleteObjectInput{
		Key:    aws.String(key),
		Bucket: aws.String(KEY_BUCKET),
	})
}

type UpdateEmbedInput struct {
	ExecutionName    string
	ApplicationID    string
	InteractionToken string
	Status           string
	Stage            string
	Color            int
}

func updateEmbed(input *UpdateEmbedInput) {
	embed := ru.CreateWorkflowEmbed(&ru.CreateWorkflowEmbedInput{
		Name:        "Terminate-Workflow",
		Description: fmt.Sprintf("WorkflowID: %s", input.ExecutionName),
		Status:      input.Status,
		Stage:       input.Stage,
		Color:       input.Color,
	})

	for {
		_, headers, err := CLIENT.EditInteractionResponse(&dc.InteractionFollowupInput{
			ApplicationID:    input.ApplicationID,
			InteractionToken: input.InteractionToken,
			Data: &dt.InteractionCallbackData{
				Embeds: []*dt.Embed{embed},
			},
		})
		if err != nil {
			log.Fatalf("Error editing embed message: %v", err)
		}
		done := dc.StatusCodeHandler(*headers)
		if done {
			break
		}
	}
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

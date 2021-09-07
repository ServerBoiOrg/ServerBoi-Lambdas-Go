package main

import (
	"encoding/json"
	"fmt"
	"log"

	gu "generalutils"

	"github.com/aws/aws-lambda-go/lambda"
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
	Private          bool   `json:"Private"`
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

	if !params.Private {
		server, err := gu.GetServerFromID(params.ServerID)
		if err != nil {
			log.Fatalf("Error getting service object: %v", err)
		}
		embed := gu.CreateServerEmbedFromServer(server)

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
	}

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

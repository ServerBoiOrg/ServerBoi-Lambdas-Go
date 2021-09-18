package main

import (
	"encoding/json"
	"fmt"
	"log"

	dc "discordhttpclient"
	gu "generalutils"
	ru "responseutils"

	dt "github.com/awlsring/discordtypes"

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

	client := dc.CreateClient(&dc.CreateClientInput{
		BotToken:   TOKEN,
		ApiVersion: "v9",
	})

	workflowEmbed := ru.CreateWorkflowEmbed(&ru.CreateWorkflowEmbedInput{
		Name:        "Provision-Server",
		Description: fmt.Sprintf("WorkflowID: %s", params.ExecutionName),
		Status:      "✔️ finished",
		Stage:       "Complete",
		Color:       ru.DiscordGreen,
	})
	workflowEmbed.Fields = append(workflowEmbed.Fields, &dt.EmbedField{
		Name:  "ServerID",
		Value: params.ServerID,
	})

	for {
		resp, headers, err := client.EditInteractionResponse(&dc.InteractionFollowupInput{
			ApplicationID:    params.ApplicationID,
			InteractionToken: params.InteractionToken,
			Data: &dt.InteractionCallbackData{
				Embeds: []*dt.Embed{workflowEmbed},
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

	if !params.Private {
		server, err := gu.GetServerFromID(params.ServerID)
		if err != nil {
			log.Fatalf("Error getting service object: %v", err)
		}

		ip, err := server.GetIPv4()
		if err != nil {
			log.Fatalf("Error getting ipv4: %v", err)
		}
		status, err := server.GetStatus()
		if err != nil {
			log.Fatalf("Error getting status: %v", err)
		}
		serverInfo := server.GetBaseService()

		serverData := ru.GetServerData(&ru.GetServerDataInput{
			Name:        serverInfo.ServerName,
			ID:          serverInfo.ServerID,
			IP:          ip,
			Status:      status,
			Region:      serverInfo.Region,
			Port:        serverInfo.Port,
			Application: serverInfo.Application,
			Owner:       serverInfo.Owner,
			Service:     serverInfo.Service,
		})

		embed := ru.CreateServerEmbed(serverData)

		log.Printf("Getting Channel for Guild")
		channelID, err := gu.GetChannelIDFromGuildID(params.GuildID)
		if err != nil {
			log.Fatalf("Error getting channelID from dynamo: %v", err)
		}

		log.Printf("Posting message")
		for {
			resp, headers, err := client.CreateMessage(&dc.CreateMessageInput{
				ChannelID: channelID,
				Data: &dt.CreateMessageData{
					Embeds:     []*dt.Embed{embed},
					Components: ru.ServerEmbedComponents(true),
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

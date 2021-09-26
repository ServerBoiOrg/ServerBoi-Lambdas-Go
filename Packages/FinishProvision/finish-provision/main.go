package main

import (
	"encoding/json"
	"fmt"
	"log"

	dc "discordhttpclient"
	gu "generalutils"
	ru "responseutils"
	sq "serverquery"

	dt "github.com/awlsring/discordtypes"

	"github.com/aws/aws-lambda-go/lambda"
)

var (
	TOKEN         = gu.GetEnvVar("DISCORD_TOKEN")
	SERVER_TABLE  = gu.GetEnvVar("SERVER_TABLE")
	CHANNEL_TABLE = gu.GetEnvVar("CHANNEL_TABLE")
	KEY_BUCKET    = gu.GetEnvVar("KEY_BUCKET")
)

type FinishProvisonParameters struct {
	ServerID         string `json:"ServerID"`
	GuildID          string `json:"GuildID"`
	InteractionToken string `json:"InteractionToken"`
	ApplicationID    string `json:"ApplicationID"`
	ExecutionName    string `json:"ExecutionName"`
	PrivateKeyObject string `json:"PrivateKeyObject"`
	Private          bool   `json:"Private"`
}

func handler(event map[string]interface{}) (bool, error) {
	log.Printf("Event: %v", event)
	params := convertEvent(event)

	client := dc.CreateClient(&dc.CreateClientInput{
		BotToken:   TOKEN,
		ApiVersion: "v9",
	})

	keyUrl := gu.CreateJankSignedKeyUrl(params.PrivateKeyObject, KEY_BUCKET)
	for {
		_, headers, err := client.PostInteractionFollowUp(&dc.InteractionFollowupInput{
			ApplicationID:    params.ApplicationID,
			InteractionToken: params.InteractionToken,
			Data: &dt.InteractionCallbackData{
				Content:    fmt.Sprintf("SSH key for Server %v", params.ServerID),
				Components: ru.CreateLinkButton(keyUrl),
				Flags:      1 << 6,
			},
		})
		if err != nil {
			log.Fatalf("Error posting follow up: %v", err)
		}
		done := dc.StatusCodeHandler(*headers)
		if done {
			log.Printf("Status Code of Response: %v", headers.StatusCode)
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
		log.Printf("Ip %v", ip)
		info, err := sq.ServerDataQuery(ip)
		if err != nil {
			log.Fatalf("Error getting server info: %v", err)
		}

		status, emoji, err := ru.GetStatus(&ru.GetStatusInput{
			Service: info.ServiceInfo.Provider,
			Running: true,
		})
		if err != nil {
			log.Fatalf("Error getting status info: %v", err)
		}

		data := ru.FormEmbedData(&ru.FormEmbedDataInput{
			Name:        info.General.Name,
			ID:          info.General.ID,
			IP:          info.General.IP,
			Port:        info.General.ClientPort,
			Status:      status,
			StatusEmoji: emoji,
			Region:      info.ServiceInfo.Region,
			Application: info.General.Application,
			Owner:       info.General.OwnerName,
			Service:     info.ServiceInfo.Provider,
			Players:     fmt.Sprintf("%v/%v", info.AppInfo.CurrentPlayers, info.AppInfo.MaxPlayers),
		})
		embed := ru.CreateServerEmbed(data)

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
				log.Printf("Status Code of Response: %v", headers.StatusCode)
				log.Printf("Response from post %v", resp)
				break
			}
		}
	}
	updateWorkflowEmbed(client, &params, keyUrl)

	return true, nil
}

func main() {
	lambda.Start(handler)
}

func updateWorkflowEmbed(client *dc.Client, params *FinishProvisonParameters, url string) {
	workflowEmbed := ru.CreateWorkflowEmbed(&ru.CreateWorkflowEmbedInput{
		Name:        "Provision-Server",
		Description: fmt.Sprintf("WorkflowID: %s", params.ExecutionName),
		Status:      "✔️ finished",
		Stage:       "Complete",
		Color:       ru.DarkGreen,
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
				Embeds:     []*dt.Embed{workflowEmbed},
				Components: ru.CreateLinkButton(url),
			},
		})
		if err != nil {
			log.Fatalf("Error getting creating message in Channel: %v", err)
		}
		done := dc.StatusCodeHandler(*headers)
		if done {
			log.Printf("Status Code of Response: %v", headers.StatusCode)
			log.Printf("Response from post %v", resp)
			break
		}
	}
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

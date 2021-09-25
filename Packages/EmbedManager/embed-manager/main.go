package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
	"time"

	dc "discordhttpclient"
	gu "generalutils"
	ru "responseutils"
	sq "serverquery"

	dt "github.com/awlsring/discordtypes"

	"github.com/aws/aws-lambda-go/lambda"
)

var TOKEN = gu.GetEnvVar("DISCORD_TOKEN")

type EmbedManagerPayload struct {
	ChannelID string `json:"ChannelID"`
}

func handler(event map[string]interface{}) (bool, error) {
	log.Printf("Event: %v", event)
	params := convertEvent(event)

	client := dc.CreateClient(&dc.CreateClientInput{
		BotToken:   TOKEN,
		ApiVersion: "v9",
	})

	var (
		messages []*dt.Message
		headers  *dc.DiscordHeaders
		err      error
	)
	for {
		messages, headers, err = client.GetChannelMessages(params.ChannelID)
		if err != nil {
			log.Fatalf("Error getting messages for channel %v", params.ChannelID)
		}
		done := dc.StatusCodeHandler(*headers)
		if done {
			break
		}
	}

	callTotal := 0
	log.Printf("Embeds to update: %v", callTotal)
	calls := make(chan *MutateMessageParams)
	for _, message := range messages {
		embeds := message.Embeds
		for _, embed := range embeds {
			callTotal++
			go analyzeEmbed(embed, message, calls)
		}
	}

	for i := 0; i < callTotal; {
		log.Printf("Making call %v.", (i + 1))
		spotHeaders := doCall(client, <-calls)
		i++
		for c := 2; c < spotHeaders.Limit; c++ {
			log.Printf("Call limit: %v", spotHeaders.Limit)
			if i == callTotal {
				break
			}
			log.Printf("Making call %v", (i + 1))
			doCall(client, <-calls)
			i++
		}
		if i < callTotal {
			log.Printf("Waiting: %v", spotHeaders.ResetAfter)
			time.Sleep(time.Duration(spotHeaders.ResetAfter*1000) * time.Millisecond)
		}
	}
	return true, nil

}

func main() {
	lambda.Start(handler)
}

func doCall(client *dc.Client, call *MutateMessageParams) *dc.DiscordHeaders {
	log.Printf("Doing call")
	var (
		headers *dc.DiscordHeaders
		err     error
	)
	for {
		if call.Post {
			_, headers, err = client.EditMessage(&dc.EditInteractionMessageInput{
				ChannelID: call.ChannelID,
				MessageID: call.MessageID,
				Data:      call.Data,
			})
		} else {
			headers, err = client.DeleteMessage(&dc.DeleteMessageInput{
				ChannelID: call.ChannelID,
				MessageID: call.MessageID,
			})
		}
		if err != nil {
			log.Printf("Breaking. Error editing message %v: %v", call.MessageID, err)
			break
		}
		log.Printf("Status Code: %v", headers.StatusCode)
		if headers.StatusCode == 429 {
			log.Printf("Throttled, waiting %v", headers.ResetAfter)
			time.Sleep(time.Duration(headers.ResetAfter*1000) * time.Millisecond)
		} else {
			break
		}
	}
	return headers
}

type MutateMessageParams struct {
	MessageID string
	ChannelID string
	Post      bool
	Data      *dt.EditMessageData
}

func analyzeEmbed(
	embed *dt.Embed,
	message *dt.Message,
	out chan<- *MutateMessageParams,
) {
	log.Printf("Looking at message %v", message.ID)
	data, err := updateServerEmbed(embed)
	if err != nil {
		out <- &MutateMessageParams{
			MessageID: message.ID,
			ChannelID: message.ChannelID,
			Post:      false,
			Data:      &dt.EditMessageData{},
		}
	} else {
		out <- &MutateMessageParams{
			MessageID: message.ID,
			ChannelID: message.ChannelID,
			Post:      true,
			Data:      data,
		}
	}
}

type UpdateServerEmbedInput struct {
	Title  string
	Status string
	IP     string
	Port   int
}

type UpdateServerEmbedOutput struct {
	ServerEmbed *dt.Embed
	ServerID    string
	Status      string
}

func updateServerEmbed(embed *dt.Embed) (output *dt.EditMessageData, err error) {
	data := &ru.ServerData{}
	log.Printf("Updating server embed.")
	for _, field := range embed.Fields {
		switch field.Name {
		case "Status":
			data.Status = field.Value
		case "Address":
			data.Address = strings.Trim(field.Value, "`")
		case "Location":
			data.Location = field.Value
		case "Application":
			data.Application = field.Value
		}
	}

	var ip string
	if strings.Contains(data.Address, ":") {
		listString := strings.Split(data.Address, ":")
		ip = listString[0]
	}
	serverID := strings.Trim(embed.Title[len(embed.Title)-6:], "()")

	if strings.Contains(data.Status, "Running") {
		log.Printf("Getting server info")
		info, err := sq.ServerDataQuery(ip)
		if err != nil {
			log.Printf("Unable to get server info, getting server object")
			server, err := gu.GetServerFromID(serverID)
			if err != nil {
				return &dt.EditMessageData{}, err
			}
			data.Status, err = getStatus(server, false)
			if err != nil {
				return &dt.EditMessageData{}, err
			}
			data.Players = "Error contacting server"
		} else {
			data.Players = fmt.Sprintf("%v/%v", info.AppInfo.CurrentPlayers, info.AppInfo.MaxPlayers)
		}
	} else {
		server, err := gu.GetServerFromID(serverID)
		if err != nil {
			return &dt.EditMessageData{}, err
		}
		data.Status, err = getStatus(server, false)
		if err != nil {
			return &dt.EditMessageData{}, err
		}
		log.Printf("Status: %v", data.Status)
		if strings.Contains(data.Status, "Running") {
			ip, err := server.GetIPv4()
			port := server.GetBaseService().Port
			if err != nil {
				log.Printf("Error getting IP: %v", err)
				data.Address = "`unknown`"
			}
			info, err := sq.ServerDataQuery(ip)
			if err != nil {
				data.Address = fmt.Sprintf("%v:%v", ip, port)
				data.Players = "Error contacting server"
			} else {
				data.Players = fmt.Sprintf("%v/%v", info.AppInfo.CurrentPlayers, info.AppInfo.MaxPlayers)
				data.Address = fmt.Sprintf("%s:%v", ip, server.GetBaseService().Port)
			}
		}
	}
	footerParts := strings.Split(embed.Footer.Text, "|")
	data.Color = setColorFromStatus(data.Status)
	data.Footer = fmt.Sprintf("%v|%v| %v", footerParts[0], footerParts[1], ru.MakeTimestamp())
	data.Name = embed.Title
	data.Description = embed.Description
	data.Thumbnail = embed.Thumbnail.URL
	var running bool
	if strings.Contains(data.Status, "Running") {
		running = true
	}

	serverEmbed := ru.CreateServerEmbed(data)

	return &dt.EditMessageData{
		Embeds:     []*dt.Embed{serverEmbed},
		Components: ru.ServerEmbedComponents(running),
	}, nil
}

func setColorFromStatus(status string) int {
	var state string
	if strings.Contains(status, "Running") {
		state = "green"
	} else if strings.Contains(status, "Starting") || strings.Contains(status, "Rebooting") {
		state = "yellow"
	} else if strings.Contains(status, "Offline") || strings.Contains(status, "Shutting down") {
		state = "red"
	}

	switch state {
	case "green":
		return ru.DiscordGreen
	case "yellow":
		return ru.Gold
	case "red":
		return ru.DiscordRed
	default:
		return 0
	}
}

func getStatus(server gu.Server, running bool) (string, error) {
	status, err := server.GetStatus()
	if err != nil {
		return "", err
	}
	state, stateEmoji, err := ru.GetStatus(&ru.GetStatusInput{
		Service: server.GetBaseService().Service,
		Status:  status,
		Running: running,
	})
	if err != nil {
		log.Println(err)
		status = "Unknown"
	} else {
		status = fmt.Sprintf("%v %v", stateEmoji, state)
	}
	return status, nil
}

func convertEvent(event map[string]interface{}) (params EmbedManagerPayload) {
	jsoned, _ := json.Marshal(event)
	params = EmbedManagerPayload{}
	if marshalErr := json.Unmarshal(jsoned, &params); marshalErr != nil {
		log.Fatal(marshalErr)
		panic(marshalErr)
	}
	return params
}

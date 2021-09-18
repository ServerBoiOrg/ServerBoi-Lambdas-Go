package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"time"

	dc "discordhttpclient"
	gu "generalutils"
	ru "responseutils"

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
	var (
		ip          string
		port        int
		status      string
		running     bool
		address     string
		location    string
		application string
		players     string
	)
	log.Printf("Updating server embed.")
	for _, field := range embed.Fields {
		switch field.Name {
		case "Status":
			status = field.Value
		case "Address":
			address = strings.Trim(field.Value, "`")
			if strings.Contains(address, ":") {
				listString := strings.Split(address, ":")
				ip = listString[0]
				port, _ = strconv.Atoi(listString[1])
			}
		case "Location":
			location = field.Value
		case "Application":
			application = field.Value
		}
	}

	serverID := strings.Trim(embed.Title[len(embed.Title)-6:], "()")
	log.Printf("ServerID: %v", serverID)
	log.Printf("Status: %v", status)
	log.Printf("Ip: %v", ip)
	log.Printf("Port: %v", port)

	if strings.Contains(status, "Running") {
		log.Printf("Getting server info")
		a2sInfo, err := ru.CallServer(ip, port)
		if err != nil {
			log.Printf("Error getting server info, getting server item")
			server, err := gu.GetServerFromID(serverID)
			if err != nil {
				return &dt.EditMessageData{}, err
			}
			status, err = getStatus(server)
			if err != nil {
				return &dt.EditMessageData{}, err
			}
			players = "Error contacting server"
		} else {
			players = fmt.Sprintf("%v/%v", a2sInfo.Players, a2sInfo.MaxPlayers)
		}
	} else {
		server, err := gu.GetServerFromID(serverID)
		if err != nil {
			return &dt.EditMessageData{}, err
		}
		status, err = getStatus(server)
		if err != nil {
			return &dt.EditMessageData{}, err
		}
		log.Printf("Status: %v", status)
		if strings.Contains(status, "Running") {
			ip, err := server.GetIPv4()
			port := server.GetBaseService().Port
			if err != nil {
				log.Printf("Error getting IP: %v", err)
				address = "`unknown`"
			}
			a2sInfo, err := ru.CallServer(ip, port)
			if err != nil {
				address = fmt.Sprintf("%v:%v", ip, port)
				players = "Error contacting server"
			} else {
				players = fmt.Sprintf("%v/%v", a2sInfo.Players, a2sInfo.MaxPlayers)
				address = fmt.Sprintf("%s:%v", ip, server.GetBaseService().Port)
			}
		}
	}
	color := setColorFromStatus(status)
	footerParts := strings.Split(embed.Footer.Text, "|")
	footer := fmt.Sprintf("%v|%v| %v", footerParts[0], footerParts[1], ru.MakeTimestamp())
	if strings.Contains(status, "Running") {
		running = true
	}

	serverEmbed := ru.CreateServerEmbed(&ru.ServerData{
		Name:        embed.Title,
		Description: embed.Description,
		Status:      status,
		Address:     address,
		Location:    location,
		Application: application,
		Players:     players,
		Color:       color,
		Footer:      footer,
		Thumbnail:   embed.Thumbnail.URL,
	})

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

func getStatus(server gu.Server) (string, error) {
	status, err := server.GetStatus()
	if err != nil {
		return "", err
	}
	state, stateEmoji, err := ru.TranslateState(
		server.GetBaseService().Service,
		status,
	)
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

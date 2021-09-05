package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"

	gu "generalutils"

	"github.com/aws/aws-lambda-go/lambda"
	"github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
)

var DISCORD_TOKEN = gu.GetEnvVar("DISCORD_TOKEN")

type EmbedManagerPayload struct {
	ChannelID string `json:"ChannelID"`
}

func handler(event map[string]interface{}) (bool, error) {
	log.Printf("Event: %v", event)
	params := convertEvent(event)

	client := gu.CreateDiscordClient(gu.CreateDiscordClientInput{
		BotToken:   DISCORD_TOKEN,
		ApiVersion: "v9",
	})

	messages, err := client.GetChannelMessages(params.ChannelID)
	if err != nil {
		log.Fatalf("Error getting messages for channel %v", params.ChannelID)
	}
	for _, message := range messages {
		embeds := message.Embeds
		for _, embed := range embeds {
			go analyzeEmbed(client, embed, message)
		}
	}
	return true, nil
}

func main() {
	lambda.Start(handler)
}

type UpdateServerEmbedInput struct {
	Title  string
	Status string
	IP     string
	Port   int
}

func analyzeEmbed(client gu.DiscordClient, embed *discordgo.MessageEmbed, message discordgo.Message) {
	serverEmbed, err := updateServerEmbed(embed)
	if err != nil {
		client.DeleteMessage(message.ChannelID, message.ID)
	} else {
		client.EditMessage(
			message.ChannelID,
			message.ID,
			gu.FormServerEmbedResponseData(serverEmbed),
		)
	}
}

func updateServerEmbed(embed *discordgo.MessageEmbed) (serverEmbed *embed.Embed, err error) {
	var (
		ip          string
		port        int
		status      string
		address     string
		location    string
		application string
		players     string
	)

	for _, field := range embed.Fields {
		switch field.Name {
		case "Status":
			status = field.Value
		case "Address":
			address = field.Value
			listString := strings.Split(field.Value, ":")
			ip = listString[0]
			port, _ = strconv.Atoi(listString[1])
		case "Location":
			location = field.Value
		case "Application":
			application = field.Value
		}
	}

	serverID := strings.Trim(embed.Title[len(embed.Title)-6:], "()")

	if status == "🟢 Running" {
		a2sInfo, err := gu.CallServer(ip, port)
		if err != nil {
			server, err := gu.GetServerFromID(serverID)
			if err != nil {
				return serverEmbed, err
			}
			status = getStatus(server)
		} else {
			players = fmt.Sprintf("%v/%v", a2sInfo.Players, a2sInfo.MaxPlayers)
		}
	} else {
		server, err := gu.GetServerFromID(serverID)
		if err != nil {
			return serverEmbed, err
		}
		status = getStatus(server)
		if status == "Running" {
			ip, err := server.GetIPv4()
			if err != nil {
				log.Printf("Error getting IP: %v", err)
				address = "`unknown`"
			}
			a2sInfo, err := gu.CallServer(ip, server.GetBaseService().Port)
			if err != nil {
				players = "Error contacting server"
			}
			players = fmt.Sprintf("%v/%v", a2sInfo.Players, a2sInfo.MaxPlayers)
			address = fmt.Sprintf("%s:%v", ip, server.GetBaseService().Port)
		}
	}

	return gu.FormServerEmbed(gu.ServerData{
		Name:        embed.Title,
		Description: embed.Description,
		Status:      status,
		Address:     address,
		Location:    location,
		Application: application,
		Players:     players,
		Color:       embed.Color,
		Footer:      embed.Footer.Text,
		Thumbnail:   embed.Thumbnail.URL,
	}), nil
}

func getStatus(server gu.Server) (status string) {
	state, err := server.Status()
	state, stateEmoji, err := gu.TranslateState(
		server.GetBaseService().Service,
		state,
	)
	if err != nil {
		log.Println(err)
		status = "Unknown"
	} else {
		status = fmt.Sprintf("%v %v", stateEmoji, state)
	}

	return status
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

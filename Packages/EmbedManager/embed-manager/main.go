package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"

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
	log.Printf("Messages: %v", len(messages))
	if err != nil {
		log.Fatalf("Error getting messages for channel %v", params.ChannelID)
	}
	var wg sync.WaitGroup
	for _, message := range messages {
		embeds := message.Embeds
		for _, embed := range embeds {
			wg.Add(1)
			go analyzeEmbed(client, embed, message, &wg)
		}
	}
	wg.Wait()
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

func analyzeEmbed(
	client gu.DiscordClient,
	embed *discordgo.MessageEmbed,
	message discordgo.Message,
	wg *sync.WaitGroup,
) {
	log.Printf("Looking at message %v", message.ID)
	updateResponse, err := updateServerEmbed(embed)
	log.Printf("Formed new embed")
	if err != nil {
		log.Printf("Error: %v", err)
		client.DeleteMessage(message.ChannelID, message.ID)
	} else {
		var running bool
		if strings.Contains(updateResponse.Status, "Running") {
			running = true
		} else {
			running = false
		}
		resp, err := client.EditMessage(
			message.ChannelID,
			message.ID,
			gu.FormServerEmbedResponseData(gu.FormServerEmbedResponseDataInput{
				ServerEmbed: updateResponse.ServerEmbed,
				Running:     running,
			}),
		)
		if err != nil {
			log.Printf("Error: %v", err)
		}
		log.Printf("Resp: %v", resp)
	}
	log.Printf("Done")
	wg.Done()
}

type UpdateServerEmbedOutput struct {
	ServerEmbed *embed.Embed
	ServerID    string
	Status      string
}

func updateServerEmbed(embed *discordgo.MessageEmbed) (output UpdateServerEmbedOutput, err error) {
	var (
		ip          string
		port        int
		status      string
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
		a2sInfo, err := gu.CallServer(ip, port)
		if err != nil {
			log.Printf("Error getting server info, getting server item")
			server, err := gu.GetServerFromID(serverID)
			if err != nil {
				return UpdateServerEmbedOutput{}, err
			}
			status = getStatus(server)
			players = "Error contacting server"
		} else {
			players = fmt.Sprintf("%v/%v", a2sInfo.Players, a2sInfo.MaxPlayers)
		}
	} else {
		server, err := gu.GetServerFromID(serverID)
		if err != nil {
			return UpdateServerEmbedOutput{}, err
		}
		status = getStatus(server)
		log.Printf("Status: %v", status)
		if strings.Contains(status, "Running") {
			ip, err := server.GetIPv4()
			port := server.GetBaseService().Port
			if err != nil {
				log.Printf("Error getting IP: %v", err)
				address = "`unknown`"
			}
			a2sInfo, err := gu.CallServer(ip, port)
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
	footer := fmt.Sprintf("%v|%v| %v", footerParts[0], footerParts[1], gu.MakeTimestamp())

	serverEmbed := gu.FormServerEmbed(gu.ServerData{
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

	return UpdateServerEmbedOutput{
		ServerEmbed: serverEmbed,
		ServerID:    serverID,
		Status:      status,
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
		return gu.DiscordGreen
	case "yellow":
		return gu.Gold
	case "red":
		return gu.DiscordRed
	default:
		return 0
	}
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

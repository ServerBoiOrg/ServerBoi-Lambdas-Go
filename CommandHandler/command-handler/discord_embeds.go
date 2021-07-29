package main

import (
	"fmt"
	"time"

	embed "github.com/clinet/discordgo-embed"
	"github.com/rumblefrog/go-a2s"
)

var thumbnails = map[string]string{
	"csgo":      "https://thumbs.gfycat.com/AffectionateTastyFirefly-size_restricted.gif",
	"wireguard": "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQtInZ2hXKFTkPDOYUmKr4sp6wkj7zzXc9KdPO0c_4ZCTC2Bv334NvT2wu7rVt8S_tV8SU&usqp=CAU",
}

const (
	Greyple      = 10070709
	DiscordGreen = 5763719
	Green        = 3066993
	DarkGreen    = 2067276
	DiscordRed   = 15548997
)

type FormWorkflowEmbedInput struct {
	Name        string
	Description string
	Status      string
	Stage       string
	Error       string
	Color       int
}

func formWorkflowEmbed(input FormWorkflowEmbedInput) *embed.Embed {
	timestamp := makeTimestamp()
	workflowEmbed := embed.NewEmbed()
	workflowEmbed.SetTitle(input.Name)
	workflowEmbed.SetDescription(input.Description)
	workflowEmbed.SetColor(1)
	workflowEmbed.AddField("Status", input.Status)
	workflowEmbed.AddField("Stage", input.Stage)
	workflowEmbed.SetFooter(timestamp)

	return workflowEmbed
}

type FormServerEmbedInput struct {
	Name        string
	Game        string
	ID          string
	IP          string
	Port        int
	Status      string
	Region      ServerBoiRegion
	Application string
	Owner       string
	Service     string
}

func formServerEmbed(input FormServerEmbedInput) *embed.Embed {
	var address string
	var description string

	if input.IP == "" {
		address = "No address while inactive"
		description = "\u200B"
	} else {
		address = fmt.Sprintf("%s:%v", input.IP, input.Port)
		description = fmt.Sprintf("Connect: steam://connect/%s", address)
	}

	state, stateEmoji, stateErr := translateState(input.Service, input.Status)
	if stateErr != nil {
		fmt.Println(stateErr)
	}

	var players string
	if state == "running" {
		a2sInfo, a2sErr := queryServer(input.IP, input.Port)
		if a2sErr != nil {
			fmt.Println(a2sErr)
			players = "Error contacting server"
		} else {
			players = fmt.Sprintf("%v/%v", a2sInfo.Players, a2sInfo.MaxPlayers)
		}
	}

	serverEmbed := embed.NewEmbed()
	serverEmbed.SetTitle(fmt.Sprintf("%v (%v)", input.Name, input.ID))
	serverEmbed.SetDescription(description)
	serverEmbed.SetColor(0)

	if url, ok := thumbnails[input.Application]; ok {
		serverEmbed.SetThumbnail(url)
	}

	serverEmbed.AddField("Status", fmt.Sprintf("%v %v", stateEmoji, state))
	serverEmbed.AddField("\u200B", "\u200B")
	serverEmbed.AddField("Address", fmt.Sprintf("`%v`", address))
	serverEmbed.AddField("Location", fmt.Sprintf("%v %v (%v)", input.Region.Emoji, input.Region.Name, input.Region.Geolocation))

	if state == "running" {
		serverEmbed.AddField("\u200B", "\u200B")
	}

	serverEmbed.AddField("Game", input.Game)
	serverEmbed.AddField("Players", players)

	timestamp := makeTimestamp()
	footer := fmt.Sprintf(
		"Owner: %v | 🌎 Hosted on %v in region %v | 🕒 Pulled at %v",
		input.Owner,
		input.Service,
		input.Region.ServiceName,
		timestamp,
	)
	serverEmbed.SetFooter(footer)

	return serverEmbed
}

func makeTimestamp() string {
	t := time.Now().UTC()
	return fmt.Sprintf("⏱️ Last updated: %02d:%02d:%02d UTC", t.Hour(), t.Minute(), t.Second())
}

func translateState(service string, status string) (state string, stateEmoji string, err error) {
	return state, stateEmoji, err
}

func queryServer(ip string, port int) (info *a2s.ServerInfo, err error) {
	clientString := fmt.Sprintf("%v:%v", ip, port)

	client, err := a2s.NewClient(clientString)
	if err != nil {
		fmt.Println(err)
	}

	defer client.Close()

	info, err = client.QueryInfo()
	if err != nil {
		fmt.Println(err)
	}

	client.Close()

	return info, err
}

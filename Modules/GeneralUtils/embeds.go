package generalutils

import (
	"fmt"
	"log"
	"time"

	embed "github.com/clinet/discordgo-embed"
	"github.com/rumblefrog/go-a2s"
)

var thumbnails = map[string]string{
	"ns2":       "https://wiki.naturalselection2.com/images/f/f3/Hive_spawn_idle.gif",
	"csgo":      "https://thumbs.gfycat.com/AffectionateTastyFirefly-size_restricted.gif",
	"wireguard": "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQtInZ2hXKFTkPDOYUmKr4sp6wkj7zzXc9KdPO0c_4ZCTC2Bv334NvT2wu7rVt8S_tV8SU&usqp=CAU",
}

const (
	Greyple      = 10070709
	DiscordGreen = 5763719
	Green        = 3066993
	DarkGreen    = 2067276
	DiscordRed   = 15548997
	Plurple      = 5793266
	Gold         = 15844367
)

type FormWorkflowEmbedInput struct {
	Name        string
	Description string
	Status      string
	Stage       string
	Error       string
	Color       int
}

func FormWorkflowEmbed(input FormWorkflowEmbedInput) *embed.Embed {
	timestamp := MakeTimestamp()
	workflowEmbed := embed.NewEmbed()
	workflowEmbed.SetTitle(input.Name)
	workflowEmbed.SetDescription(input.Description)
	workflowEmbed.SetColor(input.Color)
	workflowEmbed.AddField("Status", input.Status)
	workflowEmbed.AddField("Stage", input.Stage)
	workflowEmbed.SetFooter(timestamp)
	workflowEmbed.InlineAllFields()

	return workflowEmbed
}

type GetServerEmbedDataInput struct {
	Name        string
	ID          string
	IP          string
	Port        int
	Status      string
	Region      ServerBoiRegion
	Application string
	Owner       string
	Service     string
}

type ServerData struct {
	Name        string
	Description string
	Status      string
	Address     string
	Location    string
	Application string
	Players     string
	Footer      string
	Thumbnail   string
	Color       int
}

func GetServerEmbedData(input GetServerEmbedDataInput) ServerData {
	log.Printf("Input %v", input)
	var address string
	var description string

	if input.IP == "" {
		address = "No address while inactive"
		description = "\u200B"
	} else {
		address = fmt.Sprintf("%s:%v", input.IP, input.Port)
		description = fmt.Sprintf("Connect: steam://connect/%s", address)
	}
	state, stateEmoji, stateErr := TranslateState(input.Service, input.Status)
	if stateErr != nil {
	}

	var players string
	var color int
	if state == "Running" {
		a2sInfo, err := CallServer(input.IP, input.Port)
		if err != nil {
			log.Printf("Error contacting server: %v", err)
			players = "Error contacting server"
		} else {
			players = fmt.Sprintf("%v/%v", a2sInfo.Players, a2sInfo.MaxPlayers)
		}
		color = DiscordGreen
	} else if (state == "Offline") || (state == "Shutting down") || (state == "Terminated") {
		color = DiscordRed
	} else if (state == "Starting") || (state == "Rebooting") {
		color = Gold
	} else {
		color = Plurple
	}

	var thumbnail string
	if url, ok := thumbnails[input.Application]; ok {
		thumbnail = url
	} else {
		thumbnail = "https://cdn.dribbble.com/users/662779/screenshots/5122311/server.gif"
	}
	footer := FormFooter(input.Owner, input.Service, input.Region.ServiceName)

	return ServerData{
		Name:        fmt.Sprintf("%v (%v)", input.Name, input.ID),
		Description: description,
		Status:      fmt.Sprintf("%v %v", stateEmoji, state),
		Address:     address,
		Location:    fmt.Sprintf("%v %v (%v)", input.Region.Emoji, input.Region.Name, input.Region.Geolocation),
		Application: input.Application,
		Players:     players,
		Footer:      footer,
		Thumbnail:   thumbnail,
		Color:       color,
	}
}

type FormServerEmbedInput struct {
	Name        string
	ID          string
	IP          string
	Port        int
	Status      string
	Region      ServerBoiRegion
	Application string
	Owner       string
	Service     string
}

func FormServerEmbed(input ServerData) *embed.Embed {
	serverEmbed := embed.NewEmbed()
	serverEmbed.SetTitle(input.Name)
	serverEmbed.SetDescription(input.Description)
	serverEmbed.SetColor(input.Color)
	serverEmbed.SetThumbnail(input.Thumbnail)
	if input.Status != "🟢 Running" {
		serverEmbed.AddField("Status", input.Status)
		serverEmbed.AddField("Location", input.Location)
		serverEmbed.AddField("Application", input.Application)
	} else {
		serverEmbed.AddField("Status", input.Status)
		serverEmbed.AddField("\u200B", "\u200B")
		serverEmbed.AddField("Address", fmt.Sprintf("`%s`", input.Address))
		serverEmbed.AddField("Location", input.Location)
		serverEmbed.AddField("Application", input.Application)
		serverEmbed.AddField("Players", input.Players)
	}
	serverEmbed.SetFooter(input.Footer)
	serverEmbed.InlineAllFields()

	return serverEmbed
}

func CallServer(ip string, port int) (a2s *a2s.ServerInfo, err error) {
	for i := 0; ; i++ {
		a2s, err = QueryServer(ip, (port + i))
		if err == nil {
			return a2s, nil
		}
		if i == 5 {
			return a2s, err
		}
	}
}

func MakeTimestamp() string {
	t := time.Now().UTC()
	return fmt.Sprintf("⏱️ Last updated: %02d:%02d:%02d UTC", t.Hour(), t.Minute(), t.Second())
}

func FormFooter(owner string, service string, region string) string {
	t := time.Now().UTC()
	timestamp := fmt.Sprintf("⏱️ Last updated: %02d:%02d:%02d UTC", t.Hour(), t.Minute(), t.Second())
	return fmt.Sprintf(
		"Owner: %v | 🌎 Hosted on %v in region %v | %v",
		owner,
		service,
		region,
		timestamp,
	)
}

func TranslateState(service string, status string) (state string, stateEmoji string, err error) {
	switch service {
	case "aws":
		switch status {
		case "running":
			state = "Running"
			stateEmoji = "🟢"
		case "pending":
			state = "Starting"
			stateEmoji = "🟡"
		case "shutting-down":
			state = "Shutting down"
			stateEmoji = "🔴"
		case "stopping":
			state = "Shutting down"
			stateEmoji = "🔴"
		case "terminated":
			state = "Terminated"
			stateEmoji = "🔴"
		case "stopped":
			state = "Offline"
			stateEmoji = "🔴"
		}
	case "linode":
		switch status {
		case "running":
			state = "Running"
			stateEmoji = "🟢"
		case "offline":
			state = "Offline"
			stateEmoji = "🔴"
		case "booting":
			state = "Starting"
			stateEmoji = "🟡"
		case "rebooting":
			state = "Rebooting"
			stateEmoji = "🟡"
		case "shutting_down":
			state = "Shutting down"
			stateEmoji = "🔴"
		case "provisioning":
			state = "Starting"
			stateEmoji = "🟡"
		case "deleting":
			state = "Terminated"
			stateEmoji = "🔴"
		case "stopped":
			state = "Offline"
			stateEmoji = "🔴"
		}
	}

	return state, stateEmoji, err
}

func QueryServer(ip string, port int) (info *a2s.ServerInfo, err error) {
	clientString := fmt.Sprintf("%v:%v", ip, port)

	client, err := a2s.NewClient(clientString)
	if err != nil {
	}

	defer client.Close()

	info, err = client.QueryInfo()
	if err != nil {
	}

	client.Close()

	return info, err
}

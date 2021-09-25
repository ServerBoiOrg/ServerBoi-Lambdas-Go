package responseutils

import (
	"fmt"

	dt "github.com/awlsring/discordtypes"
)

const (
	Greyple      = 10070709
	DiscordGreen = 5763719
	Green        = 3066993
	DarkGreen    = 2067276
	DiscordRed   = 15548997
	Plurple      = 5793266
	Gold         = 15844367
)

var thumbnails = map[string]string{
	"valheim":   "https://media4.giphy.com/media/5lR0D1kLn5qptYdrKY/giphy.gif",
	"ns2":       "https://wiki.naturalselection2.com/images/f/f3/Hive_spawn_idle.gif",
	"csgo":      "https://thumbs.gfycat.com/AffectionateTastyFirefly-size_restricted.gif",
	"wireguard": "https://encrypted-tbn0.gstatic.com/images?q=tbn:ANd9GcQtInZ2hXKFTkPDOYUmKr4sp6wkj7zzXc9KdPO0c_4ZCTC2Bv334NvT2wu7rVt8S_tV8SU&usqp=CAU",
}

type CreateWorkflowEmbedInput struct {
	Name        string
	Description string
	Status      string
	Stage       string
	Error       string
	Color       int
}

func CreateWorkflowEmbed(input *CreateWorkflowEmbedInput) *dt.Embed {
	embed := NewEmbedMaker()
	timestamp := MakeTimestamp()
	embed.SetTitle(input.Name)
	embed.SetDescription(input.Description)
	embed.SetColor(input.Color)
	embed.AddField("Status", input.Status, true)
	embed.AddField("Stage", input.Stage, true)
	embed.SetFooter(timestamp, "", "")
	return embed.Embed
}

func CreateServerEmbed(input *ServerData) *dt.Embed {
	embed := NewEmbedMaker()
	embed.SetTitle(input.Name)
	embed.SetDescription(input.Description)
	embed.SetColor(input.Color)
	embed.SetThumbnail(input.Thumbnail, "", 0, 0)
	if input.Status != "🟢 Running" {
		embed.AddField("Status", input.Status, true)
		embed.AddField("Location", input.Location, true)
		embed.AddField("Application", input.Application, true)
	} else {
		embed.AddField("Status", input.Status, true)
		embed.AddField("\u200B", "\u200B", true)
		embed.AddField("Address", fmt.Sprintf("`%s`", input.Address), true)
		embed.AddField("Location", input.Location, true)
		embed.AddField("Application", input.Application, true)
		embed.AddField("Players", input.Players, true)
	}
	embed.SetFooter(input.Footer, "", "")

	return embed.Embed
}

type ServerBoiRegion struct {
	Emoji   string
	Name    string
	Service string
	// Name of region in cloud provider
	ServiceName string
	Geolocation string
}

type GetServerDataInput struct {
	Name        string
	ID          string
	IP          string
	Port        int
	Status      string
	Region      string
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

type FormEmbedDataInput struct {
	Name        string
	ID          string
	IP          string
	Port        int
	Status      string
	StatusEmoji string
	Region      string
	Application string
	Owner       string
	Service     string
	Players     string
}

func FormEmbedData(input *FormEmbedDataInput) *ServerData {
	address := fmt.Sprintf("%v:%v", input.IP, input.Port)
	regionInfo := FormRegionInfo(input.Service, input.Region)

	return &ServerData{
		Name:        fmt.Sprintf("%v (%v)", input.Name, input.ID),
		Description: fmt.Sprintf("Server Info: http://%v:7032/info", input.IP),
		Status:      fmt.Sprintf("%v %v", input.StatusEmoji, input.Status),
		Address:     address,
		Location:    fmt.Sprintf("%v %v (%v)", regionInfo.Emoji, regionInfo.Name, regionInfo.Location),
		Application: input.Application,
		Players:     input.Players,
		Footer:      FormFooter(input.Owner, input.Service, input.Region),
		Thumbnail:   GetThumbnail(input.Application),
		Color:       GetColorForState(input.Status),
	}

}

func GetThumbnail(application string) (thumbnail string) {
	if url, ok := thumbnails[application]; ok {
		thumbnail = url
	} else {
		thumbnail = "https://cdn.dribbble.com/users/662779/screenshots/5122311/server.gif"
	}
	return thumbnail
}

func GetColorForState(state string) (color int) {
	switch state {
	case "Running":
		color = DiscordGreen
	case "Offline":
		color = DiscordRed
	case "Shutting down":
		color = DiscordRed
	case "Terminated":
		color = DiscordRed
	case "Starting":
		color = Gold
	case "Rebooting":
		color = Gold
	default:
		color = Plurple
	}
	return color
}

// func GetServerData(input *GetServerDataInput) *ServerData {
// 	var address string
// 	var description string

// 	if input.IP == "" {
// 		address = "No address while inactive"
// 		description = "\u200B"
// 	} else {
// 		address = fmt.Sprintf("%s:%v", input.IP, input.Port)
// 		description = fmt.Sprintf("Connect: steam://connect/%s", address)
// 	}
// 	state, stateEmoji, stateErr := TranslateState(input.Service, input.Status)
// 	if stateErr != nil {
// 	}

// 	var players string
// 	var color int
// 	if state == "Running" {
// 		a2sInfo, err := CallServer(input.IP, input.Port)
// 		if err != nil {
// 			log.Printf("Error contacting server: %v", err)
// 			players = "Error contacting server"
// 		} else {
// 			players = fmt.Sprintf("%v/%v", a2sInfo.Players, a2sInfo.MaxPlayers)
// 		}
// 		color = DiscordGreen
// 	} else if (state == "Offline") || (state == "Shutting down") || (state == "Terminated") {
// 		color = DiscordRed
// 	} else if (state == "Starting") || (state == "Rebooting") {
// 		color = Gold
// 	} else {
// 		color = Plurple
// 	}

// 	var thumbnail string
// 	if url, ok := thumbnails[input.Application]; ok {
// 		thumbnail = url
// 	} else {
// 		thumbnail = "https://cdn.dribbble.com/users/662779/screenshots/5122311/server.gif"
// 	}
// 	footer := FormFooter(input.Owner, input.Service, input.Region)
// 	regionInfo := FormRegionInfo(input.Service, input.Region)

// 	return &ServerData{
// 		Name:        fmt.Sprintf("%v (%v)", input.Name, input.ID),
// 		Description: description,
// 		Status:      fmt.Sprintf("%v %v", stateEmoji, state),
// 		Address:     address,
// 		Location:    fmt.Sprintf("%v %v (%v)", regionInfo.Emoji, regionInfo.Name, regionInfo.Location),
// 		Application: input.Application,
// 		Players:     players,
// 		Footer:      footer,
// 		Thumbnail:   thumbnail,
// 		Color:       color,
// 	}
// }

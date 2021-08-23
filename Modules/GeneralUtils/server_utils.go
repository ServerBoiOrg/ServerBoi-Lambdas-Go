package generalutils

import "strings"

var (
	awsRegionToSBRegion = map[string]RegionMap{
		"us-west-2": RegionMap{
			Name:     "US-West",
			Location: "Oregon",
			Emoji:    "🇺🇸",
		},
	}
	linodeLocationToSBRegion = map[string]RegionMap{
		"us-west": RegionMap{
			Location: "California",
			Name:     "US-West",
			Emoji:    "🇺🇸",
		},
	}
)

type RegionMap struct {
	Name     string
	Location string
	Emoji    string
}

type WebhookTableResponse struct {
	GuildID      string `json:"GuildID"`
	WebhookID    string `json:"WebhookID"`
	WebhookToken string `json:"WebhookToken"`
}

type Server interface {
	Start() (data DiscordInteractionResponseData, err error)
	Stop() (data DiscordInteractionResponseData, err error)
	Restart() (data DiscordInteractionResponseData, err error)
	Status() (data DiscordInteractionResponseData, err error)
	GetService() string
	GetBaseService() BaseServer
	GetServerBoiRegion() ServerBoiRegion
}

func FormServerBoiRegion(service string, serviceRegion string) ServerBoiRegion {

	var regionInfo RegionMap
	switch strings.ToLower(service) {
	case "aws":
		regionInfo = awsRegionToSBRegion[serviceRegion]
	case "linode":
		regionInfo = linodeLocationToSBRegion[serviceRegion]
	}
	return ServerBoiRegion{
		Emoji:       regionInfo.Emoji,
		Name:        regionInfo.Name,
		Service:     service,
		ServiceName: serviceRegion,
		Geolocation: regionInfo.Location,
	}
}

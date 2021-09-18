package generalutils

var (
	awsRegionToSBRegion = map[string]RegionMap{
		"us-west-2": {
			Name:     "US-West",
			Location: "Oregon",
			Emoji:    "🇺🇸",
		},
	}
	linodeLocationToSBRegion = map[string]RegionMap{
		"us-west": {
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

type ChannelTableResponse struct {
	GuildID   string `json:"GuildID"`
	ChannelID string `json:"ChannelID"`
}

type OwnerTableResponse struct {
	OwnerID      string `json:"OwnerID"`
	AWSAccountID string `json:"AWSAccountID"`
	LinodeApiKey string `json:"LinodeApiKey"`
}

type Server interface {
	Start() (err error)
	Stop() (err error)
	Restart() (err error)
	Status() (status string, err error)
	AuthorizedUsers() []string
	AuthorizedRoles() []string
	GetIPv4() (string, error)
	GetService() string
	GetStatus() (string, error)
	GetBaseService() *BaseServer
}

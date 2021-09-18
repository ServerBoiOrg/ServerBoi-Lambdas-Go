package discordhttpclient

import (
	"net/http"

	dt "github.com/awlsring/discordtypes"
)

type DiscordHeaders struct {
	Limit      int
	Remaining  int
	Reset      int64
	ResetAfter float64
	StatusCode int
	Bucket     string
}

type Client struct {
	BotToken       string
	ApiVersion     string
	url            string
	webhookUrl     string
	interactionUrl string
	http           *http.Client
}

type CreateClientInput struct {
	BotToken   string
	ApiVersion string
}

type EditInteractionMessageInput struct {
	ChannelID string
	MessageID string
	Data      *dt.EditMessageData
}

type CreateMessageInput struct {
	ChannelID string
	Data      *dt.CreateMessageData
}

type DeleteMessageInput struct {
	ChannelID string
	MessageID string
}

type InteractionCallbackInput struct {
	InteractionID    string
	InteractionToken string
	Data             *dt.InteractionCallbackData
}

type InteractionFollowupInput struct {
	ApplicationID    string
	InteractionToken string
	Data             *dt.InteractionCallbackData
}

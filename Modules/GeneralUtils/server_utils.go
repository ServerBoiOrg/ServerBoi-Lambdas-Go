package generalutils

type Server interface {
	Start() (data DiscordInteractionResponseData, err error)
	Stop() (data DiscordInteractionResponseData, err error)
	Restart() (data DiscordInteractionResponseData, err error)
	Status() (data DiscordInteractionResponseData, err error)
}

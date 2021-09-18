require (
	discordhttpclient v0.0.0
	generalutils v0.0.0
	github.com/awlsring/discordtypes v0.1.6
	github.com/aws/aws-lambda-go v1.26.0
	github.com/aws/aws-sdk-go-v2 v1.8.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.4.3
	github.com/awslabs/aws-lambda-go-api-proxy v0.11.0
	github.com/bwmarrin/discordgo v0.23.2
	github.com/gin-gonic/gin v1.7.4
	github.com/linode/linodego v1.0.0
	golang.org/x/net v0.0.0-20210813160813-60bc85c4be6d // indirect
	responseutils v0.0.0
)

replace generalutils => ../../../Modules/GeneralUtils

replace discordhttpclient => ../../../Modules/DiscordHttpClient

replace responseutils => ../../../Modules/ResponseUtils

module CommandHandler

go 1.16

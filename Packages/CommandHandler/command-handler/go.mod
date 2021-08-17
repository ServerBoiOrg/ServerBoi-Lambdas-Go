require (
	generalutils v0.0.0
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go-v2 v1.8.0
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.1.3
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.4.2
	github.com/awslabs/aws-lambda-go-api-proxy v0.10.0
	github.com/bwmarrin/discordgo v0.23.2
	github.com/gin-gonic/gin v1.7.2
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
)

replace generalutils => ../../../Modules/GeneralUtils

module CommandHandler

go 1.16

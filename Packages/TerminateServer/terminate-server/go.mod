require (
	discordhttpclient v0.0.0
	generalutils v0.0.0
	github.com/awlsring/discordtypes v0.1.6
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go v1.40.37
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.5.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.17.0
	responseutils v0.0.0
)

replace generalutils => ../../../Modules/GeneralUtils

replace discordhttpclient => ../../../Modules/DiscordHttpClient

replace responseutils => ../../../Modules/ResponseUtils

module TerminateServer

go 1.16

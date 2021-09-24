require (
	discordhttpclient v0.0.0
	generalutils v0.0.0-00010101000000-000000000000
	github.com/awlsring/discordtypes v0.1.6
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go-v2 v1.9.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.16.0
	responseutils v0.0.0
)

replace generalutils => ../../../Modules/GeneralUtils

replace discordhttpclient => ../../../Modules/DiscordHttpClient

replace responseutils => ../../../Modules/ResponseUtils

module FinishProvision

go 1.16

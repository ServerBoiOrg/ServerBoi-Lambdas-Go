require (
	discordhttpclient v0.0.0
	generalutils v0.0.0
	github.com/awlsring/discordtypes v0.1.6
	github.com/aws/aws-lambda-go v1.23.0
	golang.org/x/net v0.0.0-20210614182718-04defd469f4e // indirect
	responseutils v0.0.0
	serverquery v0.0.0
)

replace generalutils => ../../../Modules/GeneralUtils

replace discordhttpclient => ../../../Modules/DiscordHttpClient

replace responseutils => ../../../Modules/ResponseUtils

replace serverquery => ../../../Modules/ServerQuery

module EmbedManager

go 1.16

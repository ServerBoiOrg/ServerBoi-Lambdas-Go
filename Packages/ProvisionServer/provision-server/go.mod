require (
	discordhttpclient v0.0.0
	generalutils v0.0.0
	github.com/awlsring/discordtypes v0.1.6
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go-v2 v1.9.1
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.8.1
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.5.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.17.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.16.0
	github.com/google/uuid v1.3.0
	github.com/linode/linodego v1.0.0
	golang.org/x/crypto v0.0.0-20200622213623-75b288015ac9
	gopkg.in/yaml.v2 v2.2.8
	responseutils v0.0.0
)

replace generalutils => ../../../Modules/GeneralUtils

replace discordhttpclient => ../../../Modules/DiscordHttpClient

replace responseutils => ../../../Modules/ResponseUtils

module ProvisionServer

go 1.16

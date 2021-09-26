require (
	discordhttpclient v0.0.0
	generalutils v0.0.0
	github.com/awlsring/discordtypes v0.1.6
	github.com/aws/aws-lambda-go v1.26.0
	github.com/aws/aws-sdk-go v1.40.45
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.5.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.18.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.16.0
	github.com/go-resty/resty/v2 v2.6.0 // indirect
	github.com/golang/protobuf v1.5.2 // indirect
	github.com/rumblefrog/go-a2s v1.0.1 // indirect
	github.com/stretchr/testify v1.7.0 // indirect
	golang.org/x/net v0.0.0-20210917221730-978cfadd31cf // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.0-20210107192922-496545a6307b // indirect
	responseutils v0.0.0
)

replace generalutils => ../../../Modules/GeneralUtils

replace discordhttpclient => ../../../Modules/DiscordHttpClient

replace responseutils => ../../../Modules/ResponseUtils

module TerminateServer

go 1.16

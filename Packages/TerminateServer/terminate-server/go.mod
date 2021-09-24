require (
	discordhttpclient v0.0.0
	generalutils v0.0.0
	github.com/awlsring/discordtypes v0.1.6
	github.com/aws/aws-lambda-go v1.26.0
	github.com/aws/aws-sdk-go v1.40.45
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.5.1
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.17.0
	github.com/go-logr/logr v1.1.0 // indirect
	github.com/go-resty/resty/v2 v2.6.0 // indirect
	github.com/google/gofuzz v1.2.0 // indirect
	github.com/imdario/mergo v0.3.12 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/rumblefrog/go-a2s v1.0.1 // indirect
	golang.org/x/crypto v0.0.0-20210915214749-c084706c2272 // indirect
	golang.org/x/net v0.0.0-20210917221730-978cfadd31cf // indirect
	golang.org/x/sys v0.0.0-20210917161153-d61c044b1678 // indirect
	golang.org/x/term v0.0.0-20210916214954-140adaaadfaf // indirect
	golang.org/x/text v0.3.7 // indirect
	google.golang.org/appengine v1.6.7 // indirect
	google.golang.org/protobuf v1.27.1 // indirect
	k8s.io/client-go v0.22.2 // indirect
	k8s.io/klog/v2 v2.20.0 // indirect
	k8s.io/utils v0.0.0-20210820185131-d34e5cb4466e // indirect
	responseutils v0.0.0
	sigs.k8s.io/structured-merge-diff/v3 v3.0.1 // indirect
)

replace generalutils => ../../../Modules/GeneralUtils

replace discordhttpclient => ../../../Modules/DiscordHttpClient

replace responseutils => ../../../Modules/ResponseUtils

module TerminateServer

go 1.16

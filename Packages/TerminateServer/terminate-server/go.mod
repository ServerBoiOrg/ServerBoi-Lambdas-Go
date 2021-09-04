require (
	generalutils v0.0.0
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go v1.40.37
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.4.3
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.13.0
)

replace generalutils => ../../../Modules/GeneralUtils

module TerminateServer

go 1.16

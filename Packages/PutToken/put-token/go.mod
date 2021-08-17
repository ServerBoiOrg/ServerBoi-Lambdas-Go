require (
	generalutils v0.0.0
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.4.2 // indirect
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.13.0 // indirect
	github.com/aws/aws-sdk-go-v2/service/s3 v1.12.0
	github.com/aws/aws-sdk-go-v2/service/sfn v1.4.2 // indirect
)

replace generalutils => ../../../Modules/GeneralUtils

module PutToken

go 1.16

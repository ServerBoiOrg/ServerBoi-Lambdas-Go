require (
	generalutils v0.0.0
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go-v2 v1.9.1
	github.com/aws/aws-sdk-go-v2/service/s3 v1.16.0
	github.com/aws/aws-sdk-go-v2/service/sfn v1.5.1
)

replace generalutils => ../../../Modules/GeneralUtils

module BootstrapCall

go 1.16

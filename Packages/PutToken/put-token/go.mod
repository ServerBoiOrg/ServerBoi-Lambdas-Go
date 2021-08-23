require (
	generalutils v0.0.0
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.12.0
)

replace generalutils => ../../../Modules/GeneralUtils

module PutToken

go 1.16

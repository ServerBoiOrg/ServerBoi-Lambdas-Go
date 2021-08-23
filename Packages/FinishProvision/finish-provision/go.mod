require (
	generalutils v0.0.0-00010101000000-000000000000
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.13.0
)

replace generalutils => ../../../Modules/GeneralUtils

module FinishProvision

go 1.16

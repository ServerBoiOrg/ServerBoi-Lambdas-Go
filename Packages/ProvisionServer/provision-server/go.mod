require (
	generalutils v0.0.0
	github.com/aws/aws-lambda-go v1.23.0
	github.com/aws/aws-sdk-go-v2 v1.8.1
	github.com/aws/aws-sdk-go-v2/feature/dynamodb/attributevalue v1.1.5
	github.com/aws/aws-sdk-go-v2/service/cloudwatch v1.7.0
	github.com/aws/aws-sdk-go-v2/service/dynamodb v1.4.3
	github.com/aws/aws-sdk-go-v2/service/ec2 v1.13.0
	github.com/aws/aws-sdk-go-v2/service/s3 v1.12.0
	github.com/google/uuid v1.3.0
)

replace generalutils => ../../../Modules/GeneralUtils

module ProvisionServer

go 1.16

package main

import (
	"context"
	"fmt"
	gu "generalutils"
	"log"
	"strings"

	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func provisionAWS(params ProvisonServerParameters) map[string]dynamotypes.AttributeValue {
	dynamo := gu.GetDynamo()

	log.Printf("Querying aws account for %v item from Dynamo", params.Owner)
	accountID := queryAWSAccountID(dynamo, params.OwnerID)
	log.Printf("Account to provision server in: %v", accountID)

	region := params.CreationOptions["region"]
	architecture := getArchitecture(params.CreationOptions)

	ec2Client := gu.CreateEC2Client(region, accountID)

	log.Printf("Getting build info")
	buildInfo := getBuildInfo(params.Application)

	container := getContainer(buildInfo, architecture)

	log.Printf("Generating bootscript")
	bootscript := formBootscript(
		FormDockerCommandInput{
			Application:      params.Application,
			Url:              params.Url,
			InteractionToken: params.InteractionToken,
			InteractionID:    params.InteractionID,
			ApplicationID:    params.ApplicationID,
			ExecutionName:    params.ExecutionName,
			ServerName:       params.ServerName,
			Container:        container,
			EnvVar:           params.CreationOptions,
		},
		buildInfo.DockerCommands,
	)

	log.Printf("Getting/Creating Security Group")
	groupID := getSecurityGroup(&ec2Client, params.Application, buildInfo.Ports)

	log.Printf("Seraching for Debian Image for %v", architecture)
	imageID := getImage(&ec2Client, architecture)

	log.Printf("Generating EBS Mapping")
	ebsMapping := getEbsMapping(buildInfo.DriveSize)

	log.Printf("Getting instance type")
	instanceType := getAWSInstanceType(buildInfo, architecture)
	oneInstance := int32(1)

	//Temporary
	testKey := "ServerBoiTestKey"

	log.Printf("Creating instance")
	response, creationErr := ec2Client.RunInstances(
		context.Background(),
		&ec2.RunInstancesInput{
			MaxCount:            &oneInstance,
			MinCount:            &oneInstance,
			UserData:            &bootscript,
			SecurityGroupIds:    []string{groupID},
			ImageId:             &imageID,
			BlockDeviceMappings: ebsMapping,
			InstanceType:        instanceType,
			KeyName:             &testKey,
		},
	)
	if creationErr != nil {
		log.Fatalf("Error creating instance: %v", creationErr)
	}

	instanceID := *response.Instances[0].InstanceId
	log.Printf("Instance created. ID: %v", instanceID)

	log.Printf(fmt.Sprintf("Ports: %v", buildInfo.Ports))

	return formAWSServerItem(
		params.OwnerID,
		params.Owner,
		params.Application,
		params.ServerName,
		buildInfo.Ports[0],
		region,
		accountID,
		instanceID,
		string(instanceType),
	)

}

func formAWSServerItem(
	ownerID string,
	owner string,
	application string,
	serverName string,
	port int,
	region string,
	accountID string,
	instanceID string,
	instanceType string,
) map[string]dynamotypes.AttributeValue {
	serverItem := formBaseServerItem(ownerID, owner, application, serverName, port)

	serverItem["Region"] = &dynamotypes.AttributeValueMemberS{Value: region}
	serverItem["AccountID"] = &dynamotypes.AttributeValueMemberS{Value: accountID}
	serverItem["InstanceID"] = &dynamotypes.AttributeValueMemberS{Value: instanceID}
	serverItem["InstanceType"] = &dynamotypes.AttributeValueMemberS{Value: instanceType}

	return serverItem
}

func getAWSInstanceType(buildInfo BuildInfo, architecture string) ec2types.InstanceType {
	var archInfo ArchitectureInfo
	var defaultType ec2types.InstanceType
	switch architecture {
	case "x86":
		archInfo = buildInfo.X86
		defaultType = ec2types.InstanceTypeC5Large
	case "arm":
		archInfo = buildInfo.Arm
		defaultType = ec2types.InstanceTypeC6gLarge
	default:
		panic("Unknown architecture")
	}

	if instanceType, ok := archInfo.InstanceType["aws"]; ok {
		instTypes := ec2types.InstanceType.Values("")
		for _, inst := range instTypes {
			if string(inst) == instanceType {
				log.Printf("Instance Type: %v", inst)
				return inst
			}
		}
		panic("Unable to find instance type")
	} else {
		log.Printf("Instance Type: %v", defaultType)
		return defaultType
	}
}

func getEbsMapping(driveSize int) []ec2types.BlockDeviceMapping {
	dev := "/dev/xvda"
	vName := "ephemeral"
	delete := true
	vType := "standard"
	size := int32(driveSize)

	ebs := ec2types.EbsBlockDevice{
		DeleteOnTermination: &delete,
		VolumeSize:          &size,
		VolumeType:          ec2types.VolumeType(vType),
	}
	return []ec2types.BlockDeviceMapping{
		{
			DeviceName:  &dev,
			VirtualName: &vName,
			Ebs:         &ebs,
		},
	}
}

func getImage(ec2Client *ec2.Client, architecture string) string {
	// Default Debian10. Skip on test
	stage := gu.GetEnvVar("STAGE")

	var imageID string
	if stage == "Testing" {
		return "0000000"
	}

	switch architecture {
	case "x86":
		architecture = "x86_64"
	case "arm":
		architecture = "arm64"
	}

	desc := "description"
	arch := "architecture"
	virtualization := "virtualization-type"

	filters := []ec2types.Filter{
		{
			Name:   &desc,
			Values: []string{"Debian 10 (20210329-591)"},
		},
		{
			Name:   &arch,
			Values: []string{architecture},
		},
		{
			Name:   &virtualization,
			Values: []string{"hvm"},
		},
	}
	owners := []string{"136693071363"}

	response, _ := ec2Client.DescribeImages(
		context.Background(),
		&ec2.DescribeImagesInput{
			Filters: filters,
			Owners:  owners,
		},
	)

	log.Printf("Debian 10 AMI Search Response: %v", response.Images)

	imageID = *response.Images[0].ImageId

	return imageID
}

func getSecurityGroup(ec2Client *ec2.Client, application string, ports []int) string {
	secGroupName := fmt.Sprintf("ServerBoi-Security-Group-%v", strings.ToUpper(application))
	secGroupDescription := fmt.Sprintf("Default Security Group for %v", strings.ToUpper(application))

	var groupID string
	createResponse, createErr := ec2Client.CreateSecurityGroup(
		context.Background(),
		&ec2.CreateSecurityGroupInput{
			GroupName:   &secGroupName,
			Description: &secGroupDescription,
		},
	)
	if createErr != nil {
		nameInList := []string{secGroupName}
		describeResponse, describeErr := ec2Client.DescribeSecurityGroups(
			context.Background(),
			&ec2.DescribeSecurityGroupsInput{
				GroupNames: nameInList,
			},
		)
		if describeErr != nil {
			panic(describeErr)
		}

		groupID = *describeResponse.SecurityGroups[0].GroupId
	} else {
		groupID = *createResponse.GroupId

		setEgress(ec2Client, groupID, ports)
		setIngress(ec2Client, groupID, ports)
	}

	return groupID
}

func setEgress(ec2Client *ec2.Client, securityGroupID string, ports []int) {
	openCidr := "0.0.0.0/0"
	ipRange := []ec2types.IpRange{
		{
			CidrIp: &openCidr,
		},
	}

	egressPermissions := []ec2types.IpPermission{}

	tcpType := "tcp"
	udpType := "udp"
	for _, port := range ports {
		p := int32(port)
		tcp := &ec2types.IpPermission{
			IpProtocol: &tcpType,
			IpRanges:   ipRange,
			FromPort:   &p,
			ToPort:     &p,
		}
		udp := &ec2types.IpPermission{
			IpProtocol: &udpType,
			IpRanges:   ipRange,
			FromPort:   &p,
			ToPort:     &p,
		}
		egressPermissions = append(egressPermissions, *tcp, *udp)
	}

	ec2Client.AuthorizeSecurityGroupEgress(
		context.Background(),
		&ec2.AuthorizeSecurityGroupEgressInput{
			GroupId:       &securityGroupID,
			IpPermissions: egressPermissions,
		},
	)
}

func setIngress(ec2Client *ec2.Client, securityGroupID string, ports []int) {
	openCidr := "0.0.0.0/0"
	tcpType := "tcp"
	ipRange := []ec2types.IpRange{
		{
			CidrIp: &openCidr,
		},
	}

	ssh := int32(22)
	http := int32(80)
	https := int32(443)

	ingressPermissions := []ec2types.IpPermission{
		{
			IpProtocol: &tcpType,
			FromPort:   &ssh,
			ToPort:     &ssh,
			IpRanges:   ipRange,
		},
		{
			IpProtocol: &tcpType,
			FromPort:   &http,
			ToPort:     &http,
			IpRanges:   ipRange,
		},
		{
			IpProtocol: &tcpType,
			FromPort:   &https,
			ToPort:     &https,
			IpRanges:   ipRange,
		},
	}

	ec2Client.AuthorizeSecurityGroupIngress(
		context.Background(),
		&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       &securityGroupID,
			IpPermissions: ingressPermissions,
		},
	)
}

type AWSTableResponse struct {
	UserID       string `json:"UserID"`
	AWSAccountID string `json:"AWSAccountID"`
}

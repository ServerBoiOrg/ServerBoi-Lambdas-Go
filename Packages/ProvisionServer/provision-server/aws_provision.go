package main

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	gu "generalutils"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func provisionAWS(params ProvisonServerParameters) (string, map[string]dynamotypes.AttributeValue) {
	log.Printf("Querying aws account for %v item from Dynamo", params.Owner)
	ownerItem, err := gu.GetOwnerItem(params.OwnerID)
	if err != nil {
		log.Fatalf("Unable to get owner item")
	}
	accountID := ownerItem.AWSAccountID
	log.Printf("Account to provision server in: %v", accountID)

	architecture := getArchitecture(params.CreationOptions)

	ec2Client := gu.CreateEC2Client(params.Region, accountID)

	log.Printf("Getting build info")
	buildInfo := getBuildInfo(params.Application)

	container := getContainer(buildInfo, architecture)

	serverID := formServerID()
	log.Printf("ServerID: %v", serverID)

	log.Printf("Generating bootscript")
	bootscript := formBootscript(
		FormDockerCommandInput{
			Application:      params.Application,
			Url:              params.Url,
			InteractionToken: params.InteractionToken,
			InteractionID:    params.InteractionID,
			ApplicationID:    params.ApplicationID,
			ExecutionName:    params.ExecutionName,
			ServerName:       params.Name,
			ServerID:         serverID,
			GuildID:          params.GuildID,
			Container:        container,
			EnvVar:           params.CreationOptions,
		},
		buildInfo.DockerCommands,
	)
	log.Printf("Converting bootscript to base64 encoding")
	bootscript = b64.StdEncoding.EncodeToString([]byte(bootscript))

	log.Printf("Getting/Creating Security Group")
	groupID := getSecurityGroup(ec2Client, params.Application, buildInfo.Ports)

	log.Printf("Seraching for Debian Image for %v", architecture)
	imageID := getImage(ec2Client, architecture)

	log.Printf("Generating EBS Mapping")
	ebsMapping := getEbsMapping(buildInfo.DriveSize)

	log.Printf("Getting instance type")
	instanceType := getAWSInstanceType(params.HardwareType, buildInfo, architecture)
	oneInstance := int32(1)

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
			TagSpecifications:   formServerTagSpec(serverID),
		},
	)
	if creationErr != nil {
		log.Fatalf("Error creating instance: %v", creationErr)
	}

	instanceID := *response.Instances[0].InstanceId
	log.Printf("Instance created. ID: %v", instanceID)
	log.Printf(fmt.Sprintf("Ports: %v", buildInfo.Ports))

	var authorized gu.Authorized
	if params.IsRole {
		authorized = gu.Authorized{
			Roles: []string{params.OwnerID},
		}
	} else {
		authorized = gu.Authorized{
			Users: []string{params.OwnerID},
		}
	}

	server := gu.AWSServer{
		OwnerID:      params.OwnerID,
		Owner:        params.Owner,
		Application:  params.Application,
		ServerName:   params.Name,
		Port:         buildInfo.Ports[0],
		Service:      "aws",
		Region:       params.Region,
		ServerID:     serverID,
		AWSAccountID: accountID,
		InstanceID:   instanceID,
		InstanceType: string(instanceType),
		Authorized:   &authorized,
	}

	return serverID, formAWSServerItem(server)

}

func formAWSServerItem(server gu.AWSServer) map[string]dynamotypes.AttributeValue {
	serverItem := formBaseServerItem(
		server.OwnerID,
		server.Owner,
		server.Application,
		server.ServerName,
		server.Service,
		server.Port,
		server.ServerID,
		server.Authorized,
	)

	serverItem["Region"] = &dynamotypes.AttributeValueMemberS{Value: server.Region}
	serverItem["AWSAccountID"] = &dynamotypes.AttributeValueMemberS{Value: server.AWSAccountID}
	serverItem["InstanceID"] = &dynamotypes.AttributeValueMemberS{Value: server.InstanceID}
	serverItem["InstanceType"] = &dynamotypes.AttributeValueMemberS{Value: server.InstanceType}

	return serverItem
}

func formServerTagSpec(serverID string) []ec2types.TagSpecification {
	nameTag := ec2types.Tag{
		Key:   aws.String("Name"),
		Value: aws.String(serverID),
	}

	tag := formManagementTag()

	return []ec2types.TagSpecification{{
		ResourceType: ec2types.ResourceTypeInstance,
		Tags:         []ec2types.Tag{nameTag, tag},
	}}
}

func formManagementTag() ec2types.Tag {
	return ec2types.Tag{
		Key:   aws.String("ManagedBy"),
		Value: aws.String("ServerBoi"),
	}
}

func getAWSInstanceType(override string, buildInfo BuildInfo, architecture string) ec2types.InstanceType {
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

	if override != "" {
		instTypes := ec2types.InstanceType.Values("")
		for _, inst := range instTypes {
			if string(inst) == override {
				log.Printf("Instance Type: %v", inst)
				return inst
			}
		}
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
	// Default Debian11
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
			Values: []string{"Debian 11 (20210814-734)"},
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

	log.Printf("Debian 11 AMI Search Response: %v", response.Images[0].ImageId)

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
			TagSpecifications: []ec2types.TagSpecification{{
				ResourceType: ec2types.ResourceTypeSecurityGroup,
				Tags:         []ec2types.Tag{formManagementTag()},
			}},
		},
	)
	if createErr != nil {
		log.Printf("Security group already exists, describing it.")
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

	egressPermissions := []ec2types.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			IpRanges:   ipRange,
			FromPort:   aws.Int32(0),
			ToPort:     aws.Int32(65535),
		},
		{
			IpProtocol: aws.String("udp"),
			IpRanges:   ipRange,
			FromPort:   aws.Int32(0),
			ToPort:     aws.Int32(65535),
		},
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
	ipRange := []ec2types.IpRange{
		{
			CidrIp: &openCidr,
		},
	}
	ingressPermissions := []ec2types.IpPermission{
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int32(22),
			ToPort:     aws.Int32(22),
			IpRanges:   ipRange,
		},
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int32(80),
			ToPort:     aws.Int32(80),
			IpRanges:   ipRange,
		},
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int32(443),
			ToPort:     aws.Int32(443),
			IpRanges:   ipRange,
		},
	}

	for _, port := range ports {
		p := int32(port)
		tcp := &ec2types.IpPermission{
			IpProtocol: aws.String("tcp"),
			IpRanges:   ipRange,
			FromPort:   &p,
			ToPort:     &p,
		}
		udp := &ec2types.IpPermission{
			IpProtocol: aws.String("udp"),
			IpRanges:   ipRange,
			FromPort:   &p,
			ToPort:     &p,
		}
		ingressPermissions = append(ingressPermissions, *tcp, *udp)
	}

	ec2Client.AuthorizeSecurityGroupIngress(
		context.Background(),
		&ec2.AuthorizeSecurityGroupIngressInput{
			GroupId:       &securityGroupID,
			IpPermissions: ingressPermissions,
		},
	)
}

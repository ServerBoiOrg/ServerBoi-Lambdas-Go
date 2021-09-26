package main

import (
	"context"
	b64 "encoding/base64"
	"fmt"
	gu "generalutils"
	"log"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/ec2"
	ec2types "github.com/aws/aws-sdk-go-v2/service/ec2/types"
)

func provisionAWS(params *ProvisonServerParameters) *ProvisionOutput {
	log.Printf("Querying aws account for %v item from Dynamo", params.Owner)
	ownerItem, err := gu.GetOwnerItem(params.OwnerID)
	if err != nil {
		log.Fatalf("Unable to get owner item")
	}
	accountID := ownerItem.AWSAccountID
	log.Printf("Account to provision server in: %v", accountID)

	output := genericProvision(params)

	log.Printf("Converting bootscript to base64 encoding")
	bootscript := b64.StdEncoding.EncodeToString([]byte(output.Bootscript))

	log.Printf("Creating EC2 Client")
	ec2Client := gu.CreateEC2Client(params.Region, accountID)

	log.Printf("Getting/Creating Security Group")
	groupID := getSecurityGroup(
		ec2Client,
		params.Application,
		output.Configuration.ClientPort,
		output.Configuration.QueryPort,
		output.Configuration.ExtraPorts,
	)

	log.Printf("Seraching for Debian Image for %v", output.Architecture)
	imageID := getImage(ec2Client, output.Architecture)

	log.Printf("Getting instance type")
	instanceType := getAWSInstanceType(output.HardwareType)

	log.Printf("Creating AWS Key Pair")
	key := importKey(ec2Client, output.PublicKey, params.ExecutionName)

	log.Printf("Creating instance")
	response, creationErr := ec2Client.RunInstances(
		context.Background(),
		&ec2.RunInstancesInput{
			MaxCount:            aws.Int32(1),
			MinCount:            aws.Int32(1),
			UserData:            &bootscript,
			SecurityGroupIds:    []string{groupID},
			ImageId:             &imageID,
			BlockDeviceMappings: getEbsMapping(output.Configuration.DriveSize),
			InstanceType:        instanceType,
			TagSpecifications:   formServerTagSpec(output.ServerID),
			KeyName:             key,
		},
	)
	if creationErr != nil {
		log.Fatalf("Error creating instance: %v", creationErr)
	}

	instanceID := *response.Instances[0].InstanceId
	log.Printf("Instance created. ID: %v", instanceID)

	var authorized *gu.Authorized
	if params.IsRole {
		authorized = &gu.Authorized{
			Roles: []string{params.OwnerID},
		}
	} else {
		authorized = &gu.Authorized{
			Users: []string{params.OwnerID},
		}
	}

	server := gu.AWSServer{
		OwnerID:      params.OwnerID,
		Owner:        params.Owner,
		Application:  params.Application,
		ServerName:   params.Name,
		Port:         output.Configuration.ClientPort,
		QueryPort:    output.Configuration.QueryPort,
		QueryType:    output.Configuration.QueryType,
		Service:      "aws",
		Region:       params.Region,
		ServerID:     output.ServerID,
		AWSAccountID: accountID,
		InstanceID:   instanceID,
		InstanceType: string(instanceType),
		PrivateKey:   output.PrivateKeyObject,
		Authorized:   authorized,
	}

	return &ProvisionOutput{
		ServerID:         output.ServerID,
		PrivateKeyObject: output.PrivateKeyObject,
		ServerItem:       formAWSServerItem(server),
	}

}

func formAWSServerItem(server gu.AWSServer) map[string]dynamotypes.AttributeValue {
	serverItem := formBaseServerItem(
		server.OwnerID,
		server.Owner,
		server.Application,
		server.ServerName,
		server.Service,
		server.Port,
		server.QueryPort,
		server.QueryType,
		server.ServerID,
		server.PrivateKey,
		server.Authorized,
	)
	serverItem["Region"] = &dynamotypes.AttributeValueMemberS{Value: server.Region}
	serverItem["AWSAccountID"] = &dynamotypes.AttributeValueMemberS{Value: server.AWSAccountID}
	serverItem["InstanceID"] = &dynamotypes.AttributeValueMemberS{Value: server.InstanceID}
	serverItem["InstanceType"] = &dynamotypes.AttributeValueMemberS{Value: server.InstanceType}
	return serverItem
}

func importKey(ec2Client *ec2.Client, publicKey string, executionName string) *string {
	key, err := ec2Client.ImportKeyPair(context.Background(), &ec2.ImportKeyPairInput{
		KeyName:           aws.String(fmt.Sprintf("%v-private", executionName)),
		PublicKeyMaterial: []byte(publicKey),
	})
	if err != nil {
		log.Fatalf("Could not create key: %v", err)
	}
	return key.KeyName
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

func getAWSInstanceType(hardwareType string) ec2types.InstanceType {
	instTypes := ec2types.InstanceType.Values("")
	for _, inst := range instTypes {
		if string(inst) == hardwareType {
			log.Printf("Instance Type: %v", inst)
			return inst
		}
	}
	panic("Unable to find instance item")
}

func getEbsMapping(driveSize int) []ec2types.BlockDeviceMapping {
	log.Println("Generating EBS Configuration")
	var size int32
	if driveSize != 0 {
		size = int32(driveSize)
	} else {
		size = 8
	}

	return []ec2types.BlockDeviceMapping{
		{
			DeviceName:  aws.String("/dev/xvda"),
			VirtualName: aws.String("ephemeral"),
			Ebs: &ec2types.EbsBlockDevice{
				DeleteOnTermination: aws.Bool(true),
				VolumeSize:          &size,
				VolumeType:          ec2types.VolumeType("standard"),
			},
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

func getSecurityGroup(
	ec2Client *ec2.Client,
	application string,
	clientPort int,
	queryPort int,
	extraPorts []string,
) string {
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

		ports := []int{
			clientPort,
			queryPort,
		}

		for _, port := range extraPorts {
			splitString := strings.Split(port, ":")
			portString, err := strconv.Atoi(splitString[0])
			if err == nil {
				ports = append(ports, portString)
			}
		}

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
		{
			IpProtocol: aws.String("tcp"),
			FromPort:   aws.Int32(7032),
			ToPort:     aws.Int32(7032),
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

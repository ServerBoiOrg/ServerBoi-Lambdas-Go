package main

import (
	"bytes"
	"context"
	"fmt"
	"log"

	gu "generalutils"

	"github.com/aws/aws-sdk-go-v2/aws"
	dynamotypes "github.com/aws/aws-sdk-go-v2/service/dynamodb/types"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

type GenericProvisionOutput struct {
	ServerID         string
	Architecture     string
	Bootscript       string
	HardwareType     string
	PublicKey        string
	PrivateKeyObject string
	Configuration    *ApplicationConfiguration
}

type ProvisionOutput struct {
	ServerID         string
	PrivateKeyObject string
	ServerItem       map[string]dynamotypes.AttributeValue
}

func genericProvision(params *ProvisonServerParameters) *GenericProvisionOutput {
	log.Printf("Creating unique ID")
	serverID := createUniqueServerID()
	log.Printf("Generated ServerID: %v", serverID)

	log.Printf("Getting architecture")
	architecture := getArchitecture(params.CreationOptions)

	log.Printf("Getting configuration")
	configuration := getConfiguration(params.Application)

	log.Println("Setting client and query port")
	clientPort := setPort(params.ClientPort, configuration.ClientPort)
	queryPort := setPort(params.QueryPort, configuration.QueryPort)

	log.Printf("Forming Application Service")
	appService := formApplicationTemplate(&FormApplicationTemplate{
		Architecture:  architecture,
		Configuration: configuration,
		Environment:   params.CreationOptions,
		ClientPort:    clientPort,
		QueryPort:     queryPort,
	})

	hardwareType := getHardwareType(architecture, params.HardwareType, configuration, params.Service)

	log.Printf("Forming Status Monitor Service")
	statusService := getStatusMonitor(&StatusMonitorEnv{
		ClientPort:   configuration.ClientPort,
		QueryPort:    configuration.QueryPort,
		QueryType:    configuration.QueryType,
		Application:  params.Application,
		Name:         params.Name,
		ID:           serverID,
		OwnerID:      params.OwnerID,
		OwnerName:    params.Owner,
		HostOS:       "Debian 11",
		HardwareType: hardwareType,
		Architecture: architecture,
		Provider:     params.Service,
		Region:       params.Region,
	})

	workflowMonitor := getWorkflowMonitor(&WorkflowMonitorInput{
		Architecture:     architecture,
		ApplicationID:    params.ApplicationID,
		InteractionToken: params.InteractionToken,
		ExecutionName:    params.ExecutionName,
	})

	log.Printf("Generating bootscript")
	bootscript := formBootscript(createDockerCompose(&CreateDockerComposeInput{
		Application:        params.Application,
		Architecture:       architecture,
		ExecutionName:      params.ExecutionName,
		QueryPort:          configuration.QueryPort,
		ApplicationService: appService,
		StatusService:      statusService,
		WorkflowMonitor:    workflowMonitor,
	}))

	log.Printf("Generating SSH Keys")
	publicKey, privateKey := generateSSHKey()

	log.Printf("Storing public key")
	keyObject := storePrivateKey(params.ExecutionName, privateKey)

	return &GenericProvisionOutput{
		ServerID:         serverID,
		Bootscript:       bootscript,
		Architecture:     architecture,
		PublicKey:        publicKey,
		PrivateKeyObject: keyObject,
		HardwareType:     hardwareType,
		Configuration:    configuration,
	}
}

func storePrivateKey(executionName string, privateKey []byte) string {
	client := gu.GetS3Client()
	reader := bytes.NewReader(privateKey)

	bucket := gu.GetEnvVar("KEY_BUCKET")
	key := fmt.Sprintf("%v-private.pem", executionName)
	_, err := client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(key),
		Body:   reader,
	})
	if err != nil {
		log.Fatalf("Error putting key: %v", err)
	}
	return key
}

func createUniqueServerID() (serverID string) {
	for {

		log.Printf("Creating unique ID")
		serverID = formServerID()
		log.Printf("Checking if serverID %v taken", serverID)
		exists, err := gu.ServerIDExists(serverID)
		if err != nil {
			log.Fatalf("Error querying table: %v", err)
		}
		if exists == false {
			return serverID
		}
		log.Printf("ID taken, trying again.")
	}
}

func getHardwareType(architecture string, hardwareType string, config *ApplicationConfiguration, service string) (htype string) {
	switch architecture {
	case "x86":
		if hardwareType == "" {
			htype = config.X86.InstanceType[service]
		} else {
			htype = hardwareType
		}
	case "arm":
		if hardwareType == "" {
			htype = config.ARM.InstanceType[service]
		} else {
			htype = hardwareType
		}
	default:
		htype = "Unknown"
	}
	return htype
}

func setPort(userPort int, configPort int) int {
	if userPort != 0 {
		return userPort
	} else {
		return configPort
	}
}

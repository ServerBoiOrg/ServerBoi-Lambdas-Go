package main

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"strconv"
	"time"

	gu "generalutils"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gopkg.in/yaml.v2"
)

type DockerCompose struct {
	Version  string              `yaml:"version"`
	Services map[string]*Service `yaml:"services"`
}

type Service struct {
	Image           string            `yaml:"image,omitempty"`
	Ports           []string          `yaml:"ports,omitempty"`
	Environment     map[string]string `yaml:"environment,omitempty"`
	Restart         string            `yaml:"restart,omitempty"`
	Volumes         []string          `yaml:"volumes,omitempty"`
	DependsOn       []string          `yaml:"depends_on,omitempty"`
	CapAdd          []string          `yaml:"cap_add,omitempty"`
	StopGracePeriod string            `yaml:"stop_grace_period,omitempty"`
	Build           string            `yaml:"build,omitempty"`
}

type FormApplicationTemplate struct {
	Architecture  string
	Environment   map[string]string
	ClientPort    int
	QueryPort     int
	Configuration *ApplicationConfiguration
}

func formApplicationTemplate(input *FormApplicationTemplate) *Service {
	var container string
	switch input.Architecture {
	case "x86":
		container = input.Configuration.X86.Container
	case "arm":
		container = input.Configuration.ARM.Container
	default:
		log.Fatalln("Unsupported architecture")
	}
	ports := []string{
		fmt.Sprintf("%v:%v/udp", input.ClientPort, input.Configuration.ClientPort),
		fmt.Sprintf("%v:%v/udp", input.QueryPort, input.Configuration.QueryPort),
	}
	if len(input.Configuration.ExtraPorts) != 0 {
		ports = append(ports, input.Configuration.ExtraPorts...)
	}
	return &Service{
		Image:           container,
		Ports:           ports,
		Environment:     input.Environment,
		Volumes:         input.Configuration.Volumes,
		CapAdd:          input.Configuration.CapAdd,
		Restart:         "unless-stopped",
		StopGracePeriod: "2m",
	}
}

type CreateDockerComposeInput struct {
	Application        string
	Architecture       string
	ExecutionName      string
	QueryPort          int
	StatusService      *Service
	ApplicationService *Service
	WorkflowMonitor    *Service
}

func createDockerCompose(input *CreateDockerComposeInput) string {
	compose := DockerCompose{
		Version: "3.9",
		Services: map[string]*Service{
			"workflow-tracking": input.WorkflowMonitor,
			"service-monitor":   input.StatusService,
			"application":       input.ApplicationService,
		},
	}

	bytesCompose, _ := yaml.Marshal(compose)
	reader := bytes.NewReader(bytesCompose)

	client := gu.GetS3Client()

	bucket := gu.GetEnvVar("COMPOSE_BUCKET")
	keyName := fmt.Sprintf("%v-template.yml", input.ExecutionName)
	_, err := client.PutObject(context.Background(), &s3.PutObjectInput{
		Bucket: aws.String(bucket),
		Key:    aws.String(keyName),
		Body:   reader,
	})
	if err != nil {
		log.Fatalf("Error putting object: %v", err)
	}

	return fmt.Sprintf("https://%v.s3.us-west-2.amazonaws.com/%v", bucket, keyName)
}

type StatusMonitorEnv struct {
	ClientPort   int
	QueryPort    int
	QueryType    string
	Application  string
	Name         string
	ID           string
	OwnerID      string
	OwnerName    string
	HostOS       string
	Architecture string
	Provider     string
	HardwareType string
	Region       string
}

func getStatusMonitor(env *StatusMonitorEnv) *Service {
	var image string
	switch env.Architecture {
	case "x86":
		image = "serverboi/status-monitor:latest"
	case "arm":
		image = "serverboi/status-monitor:latest"
	default:
		log.Fatalln("Unknown architecture")
	}
	return &Service{
		Image: image,
		Ports: []string{"7032:7032/tcp"},
		Environment: map[string]string{
			"CREATED":       time.Now().UTC().String(),
			"CLIENT_PORT":   strconv.Itoa(env.ClientPort),
			"QUERY_PORT":    strconv.Itoa(env.QueryPort),
			"QUERY_TYPE":    env.QueryType,
			"APPLICATION":   env.Application,
			"NAME":          env.Name,
			"ID":            env.ID,
			"OWNER_ID":      env.OwnerID,
			"OWNER_NAME":    env.OwnerName,
			"HOST_OS":       env.HostOS,
			"ARCHITECTURE":  env.Architecture,
			"PROVIDER":      env.Provider,
			"HARDWARE_TYPE": env.HardwareType,
			"REGION":        env.Region,
		},
		Restart: "always",
	}
}

type WorkflowMonitorInput struct {
	Architecture     string
	ApplicationID    string
	InteractionToken string
	ExecutionName    string
}

func getWorkflowMonitor(input *WorkflowMonitorInput) *Service {
	var image string
	switch input.Architecture {
	case "x86":
		image = "serverboi/workflow-tracking:latest"
	case "arm":
		image = "serverboi/workflow-tracking:latest"
	default:
		log.Fatalln("Unknown architecture")
	}
	return &Service{
		Image:     image,
		DependsOn: []string{"service-monitor"},
		Environment: map[string]string{
			"APPLICATION_ID":    input.ApplicationID,
			"INTERACTION_TOKEN": input.InteractionToken,
			"EXECUTION_NAME":    input.ExecutionName,
		},
	}
}

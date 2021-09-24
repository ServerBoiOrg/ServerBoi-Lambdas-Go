package main

import (
	"context"
	"fmt"
	gu "generalutils"
	"io/ioutil"
	"log"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"gopkg.in/yaml.v2"
)

type ApplicationConfiguration struct {
	Application string                     `yaml:"application"`
	X86         *ArchitectureConfiguration `yaml:"x86,omitempty"`
	ARM         *ArchitectureConfiguration `yaml:"arm,omitempty"`
	Volumes     []string                   `yaml:"volumes,omitempty"`
	CapAdd      []string                   `yaml:"capAdd,omitempty"`
	ClientPort  int                        `yaml:"clientPort"`
	QueryPort   int                        `yaml:"queryPort"`
	QueryType   string                     `yaml:"queryType"`
	ExtraPorts  []string                   `yaml:"extraPorts,omitempty"`
	DriveSize   int                        `yaml:"driveSize"`
}

type ArchitectureConfiguration struct {
	Container    string            `yaml:"container"`
	InstanceType map[string]string `yaml:"instanceType"`
}

func getConfiguration(app string) *ApplicationConfiguration {
	log.Printf("Getting configuration for %v", app)
	client := gu.GetS3Client()
	requestInput := &s3.GetObjectInput{
		Bucket: aws.String(gu.GetEnvVar("CONFIGURATION_BUCKET")),
		Key:    aws.String(fmt.Sprintf("%v-build.yml", app)),
	}
	var configuration *ApplicationConfiguration
	result, err := client.GetObject(context.TODO(), requestInput)
	if err != nil {
		log.Fatalln(err)
	}
	defer result.Body.Close()
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		log.Fatalln(err)
	}
	yaml.Unmarshal(body, &configuration)
	log.Println(string(body))
	log.Println(configuration)
	return configuration
}

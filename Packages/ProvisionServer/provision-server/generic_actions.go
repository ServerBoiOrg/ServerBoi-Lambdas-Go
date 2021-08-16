package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"strings"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
)

func getBuildInfo(game string) BuildInfo {
	client := getS3Client()
	requestInput := &s3.GetObjectInput{
		Bucket: aws.String("serverboi-sam-packages"),
		Key:    aws.String("build.json"),
	}
	var gamesData map[string]interface{}
	result, err := client.GetObject(context.TODO(), requestInput)
	if err != nil {
		fmt.Println(err)
	}
	defer result.Body.Close()
	body, err := ioutil.ReadAll(result.Body)
	if err != nil {
		fmt.Println(err)
	}

	json.Unmarshal(body, &gamesData)
	gameData := gamesData[game]
	jsoned, _ := json.Marshal(gameData)
	var buildInfo BuildInfo
	json.Unmarshal(jsoned, &buildInfo)

	return buildInfo
}

func getBuildInfoLocal(game string) BuildInfo {
	var gamesData map[string]interface{}
	jsonFile, err := os.Open("build.json")
	if err != nil {
		log.Panicf("Build file not openable: %v", err)
	}
	rawBuildData, _ := ioutil.ReadAll(jsonFile)
	defer jsonFile.Close()
	json.Unmarshal(rawBuildData, &gamesData)
	gameData := gamesData[game]
	fmt.Printf(fmt.Sprintf("GameData: %v", game))
	jsoned, _ := json.Marshal(gameData)
	var buildInfo BuildInfo
	json.Unmarshal(jsoned, &buildInfo)
	fmt.Printf("BuildInfo: %v", buildInfo)
	return buildInfo
}

type BuildInfo struct {
	X86            ArchitectureInfo `json:"x86,omitempty"`
	Arm            ArchitectureInfo `json:"arm,omitempty"`
	Ports          []int            `json:"ports,omitempty"`
	DriveSize      int              `json:"driveSize,omitempty"`
	DockerCommands []string         `json:"dockerCommands,omitempty"`
}

type ArchitectureInfo struct {
	Container    string            `json:"container"`
	InstanceType map[string]string `json:"instanceType,omitempty"`
}

func getContainer(buildInfo BuildInfo, architecture string) (container string) {
	switch architecture {
	case "x86":
		container = buildInfo.X86.Container
	case "arm":
		container = buildInfo.Arm.Container
	default:
		panic("Unknown architecture")
	}
	return container
}

func getArchitecture(creationOptions map[string]string) (architecture string) {
	if arch, ok := creationOptions["Architecture"]; ok {
		architecture = arch
	} else {
		architecture = "x86"
	}
	return architecture
}

func formBootscript(input FormDockerCommandInput, dockerCommands []string) string {
	dockerCommand := formDockerCommand(input, dockerCommands)
	return fmt.Sprintf(`#!/bin/bash
    sudo apt-get update && sudo apt-get upgrade -y
    sudo apt-get install \
        apt-transport-https \
        ca-certificates \
        curl \
        gnupg \
        lsb-release -y
    curl -fsSL https://download.docker.com/linux/debian/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
    echo \
      "deb [arch=amd64 signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/debian \
      $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
    sudo apt-get update
    sudo apt-get install docker-ce docker-ce-cli containerd.io -y
    %v`, dockerCommand)
}

func formEnvFile() {

}

func formDockerCommand(input FormDockerCommandInput, dockerCommands []string) string {
	command := fmt.Sprintf(`sudo docker run -t -d \
    --net=host \
    --name serverboi-%v \
    -e INTERACTION_TOKEN=%v \
    -e APPLICATION_ID=%v \
    -e EXECUTION_NAME=%v \
    -e WORKFLOW_ENDPOINT=%v \
    -e SERVER_NAME='%v' `,
		strings.ToLower(input.Application),
		input.InteractionToken,
		input.ApplicationID,
		input.ExecutionName,
		input.Url,
		input.ServerName)

	for k, v := range input.EnvVar {
		key := strings.Replace(k, "-", "_", -1)
		command = fmt.Sprintf("%v-e %v=%v ", command, strings.ToUpper(key), v)
	}
	for _, dockerCommand := range dockerCommands {
		command = fmt.Sprintf("%v%v", command, dockerCommand)
	}
	command = fmt.Sprintf("%v%v", command, input.Container)

	return command
}

type FormDockerCommandInput struct {
	Application      string
	Url              string
	InteractionToken string
	InteractionID    string
	ApplicationID    string
	ExecutionName    string
	ServerName       string
	Container        string
	EnvVar           map[string]string
}

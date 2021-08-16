package main

import (
	"bytes"
	"context"
	b64 "encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/bwmarrin/discordgo"
	embed "github.com/clinet/discordgo-embed"
)

type FormWorkflowEmbedInput struct {
	Name        string
	Description string
	Status      string
	Stage       string
	Error       string
	Color       int
}

type DiscordSelectMenuOptions struct {
	Label       string       `json:"label"`
	Value       string       `json:"value"`
	Description string       `json:"description,omitempty"`
	Emoji       DiscordEmoji `json:"emoji,omitempty"`
	Default     bool         `json:"default,omitempty"`
}

type DiscordEmoji struct {
	// Component will have name, id, animated
	ID            string           `json:"id"`
	Name          string           `json:"name"`
	Roles         []discordgo.Role `json:"roles,omitempty"`
	User          discordgo.User   `json:"user,omitempty"`
	RequireColons bool             `json:"require_colons,omitempty"`
	Managed       bool             `json:"managed,omitempty"`
	Animated      bool             `json:"animated,omitempty"`
	Available     bool             `json:"available,omitempty"`
}

type DiscordComponentData struct {
	Type        int                        `json:"type"`
	CustomID    string                     `json:"custom_id,omitempty"`
	Disabled    string                     `json:"disabled,omitempty"`
	Style       int                        `json:"style,omitempty"`
	Label       string                     `json:"label,omitempty"`
	Emoji       DiscordEmoji               `json:"emoji,omitempty"`
	Url         string                     `json:"url,omitempty"`
	Options     []DiscordSelectMenuOptions `json:"options,omitempty"`
	Placeholder string                     `json:"placeholder,omitempty"`
	MinValues   int                        `json:"min_values,omitempty"`
	MaxValues   int                        `json:"max_values,omitempty"`
	Components  []DiscordComponentData     `json:"components,omitempty"`
}

type DiscordInteractionResponseData struct {
	TTS             bool                         `json:"tts,omitempty"`
	Content         string                       `json:"content,omitempty"`
	Embeds          []embed.Embed                `json:"embeds,omitempty"`
	AllowedMentions discordgo.AllowedMentionType `json:"allowed_mentions,omitempty"`
	Flags           int                          `json:"flags,omitempty"`
	Components      []DiscordComponentData       `json:"components,omitempty"`
}

func makeTimestamp() string {
	t := time.Now().UTC()
	return fmt.Sprintf("⏱️ Last updated: %02d:%02d:%02d UTC", t.Hour(), t.Minute(), t.Second())
}

func formWorkflowEmbed(input FormWorkflowEmbedInput) *embed.Embed {
	timestamp := makeTimestamp()
	workflowEmbed := embed.NewEmbed()
	workflowEmbed.SetTitle(input.Name)
	workflowEmbed.SetDescription(input.Description)
	workflowEmbed.SetColor(1)
	workflowEmbed.AddField("Status", input.Status)
	workflowEmbed.AddField("Stage", input.Stage)
	workflowEmbed.SetFooter(timestamp)

	return workflowEmbed
}

func updateResponse(applicationID string, interactionToken string, data DiscordInteractionResponseData) {
	responseUrl := fmt.Sprintf("https://discord.com/api/webhooks/%s/%s/messages/@original", applicationID, interactionToken)
	log.Printf("URL to Patch: %v", responseUrl)

	log.Printf("Editing response with: %v", data)
	responseBody, _ := json.Marshal(data)
	bytes := bytes.NewBuffer(responseBody)

	request, err := http.NewRequest(http.MethodPatch, responseUrl, bytes)
	if err != nil {
		log.Fatal(err)
	}
	request.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(request)
	if err != nil {
		log.Fatal(err)
	}

	log.Printf("Response from discord: %v", resp)

	defer resp.Body.Close()
}

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

	if buildInfo.DriveSize == 0 {
		buildInfo.DriveSize = 8
	}

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
	bootscript := fmt.Sprintf(`#!/bin/bash
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
	log.Printf("Bootscript: %v", bootscript)
	return b64.StdEncoding.EncodeToString([]byte(bootscript))
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

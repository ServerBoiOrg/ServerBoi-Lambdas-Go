package generalutils

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
	embed "github.com/clinet/discordgo-embed"
)

var (
	awsRegionToSBRegion = map[string]RegionMap{
		"us-west-2": RegionMap{
			Name:     "US-West",
			Location: "Oregon",
			Emoji:    "🇺🇸",
		},
	}
	linodeLocationToSBRegion = map[string]RegionMap{
		"us-west": RegionMap{
			Location: "California",
			Name:     "US-West",
			Emoji:    "🇺🇸",
		},
	}
)

type RegionMap struct {
	Name     string
	Location string
	Emoji    string
}

type WebhookTableResponse struct {
	GuildID      string `json:"GuildID"`
	WebhookID    string `json:"WebhookID"`
	WebhookToken string `json:"WebhookToken"`
}

type ChannelTableResponse struct {
	GuildID   string `json:"GuildID"`
	ChannelID string `json:"ChannelID"`
}

type Server interface {
	Start() (err error)
	Stop() (err error)
	Restart() (err error)
	Status() (status string, err error)
	AuthorizedUsers() []string
	AuthorizedRoles() []string
	GetIPv4() (string, error)
	GetService() string
	GetBaseService() BaseServer
	GetServerBoiRegion() ServerBoiRegion
}

func GetStatus(server Server) (status string) {
	state, err := server.Status()
	state, stateEmoji, err := TranslateState(
		server.GetBaseService().Service,
		state,
	)
	if err != nil {
		log.Println(err)
		status = "Unknown"
	} else {
		status = fmt.Sprintf("%v %v", stateEmoji, state)
	}

	return status
}

func CreateServerEmbedFromServer(server Server) *embed.Embed {
	var (
		ip    string
		state string
	)
	service := server.GetService()
	switch service {
	case "aws":
		awsServer, _ := server.(AWSServer)
		client := CreateEC2Client(awsServer.Region, awsServer.AWSAccountID)
		response, err := client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
			InstanceIds: []string{
				awsServer.InstanceID,
			},
		})
		if err != nil {
			log.Fatalf("Error describing instance: %v", err)
		}

		ip = *response.Reservations[0].Instances[0].PublicIpAddress
		state = string(response.Reservations[0].Instances[0].State.Name)
	case "linode":
		linodeServer, _ := server.(LinodeServer)
		client := CreateLinodeClient(linodeServer.ApiKey)

		linode, err := client.GetInstance(context.Background(), linodeServer.LinodeID)
		if err != nil {
			log.Fatalf("Error describing linode: %v", err)
		}

		ip = fmt.Sprintf("%v", linode.IPv4[0])
		state = string(linode.Status)
	}
	log.Printf("IP of server: %v", ip)
	log.Printf("State of server: %v", state)

	serverInfo := server.GetBaseService()
	sbRegion := server.GetServerBoiRegion()

	serverData := GetServerEmbedData(GetServerEmbedDataInput{
		Name:        serverInfo.ServerName,
		ID:          serverInfo.ServerID,
		IP:          ip,
		Status:      state,
		Region:      sbRegion,
		Port:        serverInfo.Port,
		Application: serverInfo.Application,
		Owner:       serverInfo.Owner,
		Service:     serverInfo.Service,
	})
	return FormServerEmbed(serverData)

}

func FormServerBoiRegion(service string, serviceRegion string) ServerBoiRegion {

	var regionInfo RegionMap
	switch strings.ToLower(service) {
	case "aws":
		regionInfo = awsRegionToSBRegion[serviceRegion]
	case "linode":
		regionInfo = linodeLocationToSBRegion[serviceRegion]
	}
	return ServerBoiRegion{
		Emoji:       regionInfo.Emoji,
		Name:        regionInfo.Name,
		Service:     service,
		ServiceName: serviceRegion,
		Geolocation: regionInfo.Location,
	}
}

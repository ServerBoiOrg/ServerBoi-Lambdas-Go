package generalutils

import (
	"context"
	"fmt"
	"net/http"

	"github.com/linode/linodego"
	"golang.org/x/oauth2"
)

func CreateLinodeClient(apiKey string) linodego.Client {
	tokenSource := oauth2.StaticTokenSource(&oauth2.Token{AccessToken: apiKey})
	oauth2Client := &http.Client{
		Transport: &oauth2.Transport{
			Source: tokenSource,
		},
	}
	// return linodego.NewClient(oauth2Client)
	return linodego.NewClient(oauth2Client)
}

func CreateAuthlessLinodeClient() linodego.Client {
	return linodego.NewClient(&http.Client{})
}

func (server LinodeServer) Start() (err error) {
	client := CreateLinodeClient(server.ApiKey)

	err = client.BootInstance(context.Background(), server.LinodeID, 0)
	if err != nil {
		return err
	}

	return nil
}

func (server LinodeServer) Stop() (err error) {
	client := CreateLinodeClient(server.ApiKey)

	err = client.ShutdownInstance(context.Background(), server.LinodeID)
	if err != nil {
		return err
	}

	return nil
}

func (server LinodeServer) Restart() (err error) {
	client := CreateLinodeClient(server.ApiKey)

	err = client.RebootInstance(context.Background(), server.LinodeID, 0)
	if err != nil {
		return err
	}

	return nil
}

func (server LinodeServer) Status() (status string, err error) {
	client := CreateLinodeClient(server.ApiKey)
	linode, err := client.GetInstance(context.Background(), server.LinodeID)
	if err != nil {
		return status, err
	}

	return fmt.Sprintf("%v", linode.Status), nil
}

func (server LinodeServer) AuthorizedUsers() []string {
	return server.Authorized.Users
}

func (server LinodeServer) AuthorizedRoles() []string {
	return server.Authorized.Roles
}

func (server LinodeServer) GetService() string {
	return server.Service
}

func (server LinodeServer) GetIPv4() (string, error) {
	client := CreateLinodeClient(server.ApiKey)
	linode, err := client.GetInstance(context.Background(), server.LinodeID)
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", linode.IPv4[0]), nil
}

func (server LinodeServer) GetStatus() (string, error) {
	client := CreateLinodeClient(server.ApiKey)
	linode, err := client.GetInstance(context.Background(), server.LinodeID)
	if err != nil {
		return "", err
	}
	return string(linode.Status), nil
}

func (server LinodeServer) GetBaseService() *BaseServer {
	return &BaseServer{
		ServerID:    server.ServerID,
		Application: server.Application,
		ServerName:  server.ServerName,
		Service:     server.Service,
		Owner:       server.Owner,
		OwnerID:     server.OwnerID,
		Port:        server.Port,
		Region:      server.Location,
	}
}

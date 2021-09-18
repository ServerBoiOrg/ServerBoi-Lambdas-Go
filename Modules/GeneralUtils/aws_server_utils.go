package generalutils

import (
	"context"
	"fmt"
	"log"

	"github.com/aws/aws-sdk-go-v2/service/ec2"
)

func (server AWSServer) Start() (err error) {
	client := CreateEC2Client(server.Region, server.AWSAccountID)
	input := &ec2.StartInstancesInput{
		InstanceIds: []string{
			server.InstanceID,
		},
	}
	_, err = client.StartInstances(context.Background(), input)
	if err != nil {
		return err
	}

	return nil
}

func (server AWSServer) Stop() (err error) {
	client := CreateEC2Client(server.Region, server.AWSAccountID)
	input := &ec2.StopInstancesInput{
		InstanceIds: []string{
			server.InstanceID,
		},
	}
	_, err = client.StopInstances(context.Background(), input)
	if err != nil {
		return err
	}

	return nil
}

func (server AWSServer) GetService() string {
	return server.Service
}

func (server AWSServer) GetIPv4() (string, error) {
	client := CreateEC2Client(server.Region, server.AWSAccountID)
	log.Printf("Describing instance: %s", server.InstanceID)
	response, err := client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			server.InstanceID,
		},
	})
	if err != nil {
		return "", err
	}
	return fmt.Sprintf("%v", *response.Reservations[0].Instances[0].PublicIpAddress), nil
}

func (server AWSServer) GetStatus() (string, error) {
	client := CreateEC2Client(server.Region, server.AWSAccountID)
	response, err := client.DescribeInstances(context.Background(), &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			server.InstanceID,
		},
	})
	if err != nil {
		return "", err
	}

	return string(response.Reservations[0].Instances[0].State.Name), nil
}

func (server AWSServer) GetBaseService() *BaseServer {
	return &BaseServer{
		ServerID:    server.ServerID,
		Application: server.Application,
		ServerName:  server.ServerName,
		Service:     server.Service,
		Owner:       server.Owner,
		OwnerID:     server.OwnerID,
		Port:        server.Port,
		Region:      server.Region,
	}
}

func (server AWSServer) Restart() (err error) {
	client := CreateEC2Client(server.Region, server.AWSAccountID)
	_, err = client.RebootInstances(context.Background(), &ec2.RebootInstancesInput{
		InstanceIds: []string{
			server.InstanceID,
		},
	})
	if err != nil {
		return err
	}

	return nil
}

func (server AWSServer) Status() (status string, err error) {
	client := CreateEC2Client(server.Region, server.AWSAccountID)
	log.Printf("Ec2 Client made in Target account")
	input := &ec2.DescribeInstancesInput{
		InstanceIds: []string{
			server.InstanceID,
		},
	}
	response, err := client.DescribeInstances(context.Background(), input)
	if err != nil {
		return status, err
	}

	return fmt.Sprintf("%v", response.Reservations[0].Instances[0].State.Name), nil
}

func (server AWSServer) AuthorizedUsers() []string {
	return server.Authorized.Users
}

func (server AWSServer) AuthorizedRoles() []string {
	return server.Authorized.Roles
}

type NoItemsError struct {
	Path string
}

func (e *NoItemsError) Error() string {
	return fmt.Sprintf("No response for GetItem operation: %v", e.Path)
}

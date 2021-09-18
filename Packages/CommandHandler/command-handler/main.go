package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"time"

	dc "discordhttpclient"
	gu "generalutils"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

var (
	ginLambda    *ginadapter.GinLambda
	SERVER_TABLE = gu.GetEnvVar("SERVER_TABLE")
	TOKEN        = gu.GetEnvVar("DISCORD_TOKEN")
	client       = dc.CreateClient(&dc.CreateClientInput{
		BotToken:   TOKEN,
		ApiVersion: "v9",
	})
)

func init() {
	log.Printf("Gin cold start")
	router := gin.Default()
	router.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	ginLambda = ginadapter.New(router)
}

func Handler(ctx context.Context, event events.APIGatewayProxyRequest) (lambdaResponse events.APIGatewayProxyResponse, err error) {
	rawEvent, marshalErr := json.Marshal(event)
	if err != nil {
		fmt.Println(marshalErr)
		return
	}
	log.Printf(string(rawEvent))

	//Proxy event to standard http request
	request, err := ginLambda.ProxyEventToHTTPRequest(event)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}

	log.Println("Verifying Public Key")
	verifyPublicKey(request)

	interactionType := getCommandType(event.Body)
	log.Printf("Interaction Type: %v", interactionType)

	var output *dc.InteractionFollowupInput
	switch {
	case interactionType == 1:
		data, _ := json.Marshal(pong())
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       fmt.Sprintf(string(data)),
		}, nil
	case interactionType == 2:
		output = command(event.Body)
	case interactionType == 3:
		output = component(event.Body)
	}
	if err != nil {
		log.Fatalf("Error performing command: %v", err)
		return lambdaResponse, err
	}

	for {
		_, headers, _ := client.EditInteractionResponse(&dc.InteractionFollowupInput{
			ApplicationID:    output.ApplicationID,
			InteractionToken: output.InteractionToken,
			Data:             output.Data,
		})
		if headers.StatusCode == 429 {
			log.Printf("Thottled, waiting")
			time.Sleep(time.Duration(headers.ResetAfter*1000) * time.Millisecond)
		}
		break
	}

	//Probably not needed but eh
	return events.APIGatewayProxyResponse{
		StatusCode: 200,
		Body:       "",
	}, nil

}

func getCommandType(eventBody string) float64 {
	var tempBody map[string]interface{}

	json.Unmarshal([]byte(eventBody), &tempBody)

	return tempBody["type"].(float64)
}

func verifyPublicKey(request *http.Request) {
	publicKeyString := gu.GetEnvVar("PUBLIC_KEY")
	publicKey := gu.DecodeToPublicKey(publicKeyString)

	if !discordgo.VerifyInteraction(request, publicKey) {
		fmt.Println("Public Key Not Verified")
		panic("Public Key Not Verified")
	}
}

func main() {
	lambda.Start(Handler)
}

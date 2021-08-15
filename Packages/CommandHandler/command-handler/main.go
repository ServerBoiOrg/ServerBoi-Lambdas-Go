package main

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"
	"github.com/bwmarrin/discordgo"
	"github.com/gin-gonic/gin"
)

var ginLambda *ginadapter.GinLambda

var (
	SERVER_TABLE = getEnvVar("SERVER_TABLE")
)

func init() {
	log.Printf("Gin cold start")
	// gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "pong",
		})
	})

	ginLambda = ginadapter.New(r)
}

func Handler(ctx context.Context, event events.APIGatewayProxyRequest) (lambdaResponse events.APIGatewayProxyResponse, err error) {
	log.Printf("Input: %v", event)
	request, err := ginLambda.ProxyEventToHTTPRequest(event)
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	log.Println("Verifying Public Key")
	verifyPublicKey(request)

	interactionType := getCommandType(event.Body)
	log.Printf("Interaction Type: %v", interactionType)

	var response DiscordInteractionResponse
	var applicationID string
	var interactionToken string
	switch {
	case interactionType == 1:
		response = pong()
		data, _ := json.Marshal(response)
		return events.APIGatewayProxyResponse{
			StatusCode: 200,
			Body:       fmt.Sprintf(string(data)),
		}, nil
	case interactionType == 2:
		applicationID, interactionToken, response, err = command(event.Body)
	case interactionType == 3:
		applicationID, interactionToken, response = component(event.Body)
	}

	if err != nil {
		log.Fatalf("Error performing command: %v", err)
		return lambdaResponse, err
	}

	log.Printf("Sending Response to Discord. Response: %v", response.Data)
	editResponse(applicationID, interactionToken, response)

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
	publicKeyString := getEnvVar("PUBLIC_KEY")
	publicKey := decodeToPublicKey(publicKeyString)

	if !discordgo.VerifyInteraction(request, publicKey) {
		fmt.Println("Public Key Not Verified")
		panic("Public Key Not Verified")
	}
}

func main() {
	lambda.Start(Handler)
}

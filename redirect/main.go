package main

import (
	"context"
	"os"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

func Handler(ctx context.Context, request events.APIGatewayProxyRequest) (events.APIGatewayProxyResponse, error) {
	webfrontend := os.Getenv("WEB_FRONTEND")
	return events.APIGatewayProxyResponse{
		StatusCode: 301,
		Headers:    map[string]string{"Location": webfrontend},
	}, nil
}

func main() {
	lambda.Start(Handler)
}

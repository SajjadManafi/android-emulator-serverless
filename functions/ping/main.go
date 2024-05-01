package main

import (
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response events.APIGatewayProxyResponse

func Handler(request events.APIGatewayProxyRequest) (Response, error) {
	return Response{
		StatusCode: 200,
		Body:       "pong",
	}, nil
}

func main() {
	lambda.Start(Handler)
}

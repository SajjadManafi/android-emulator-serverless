package main

import (
	"io/ioutil"
	"log"
	"net/http"

	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response events.APIGatewayProxyResponse

func Handler(request events.APIGatewayProxyRequest) (Response, error) {
	// Define the URL including the port
	url := "http://localhost:8000"

	// Send a GET request to the specified URL
	response, err := http.Get(url)
	if err != nil {
		// Handle error
		log.Println("Error sending request:", err)
		return Response{
			StatusCode: 500,
			Body:       "Internal Server Error",
		}, nil
	}
	defer response.Body.Close()

	// Read the response body
	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		// Handle error
		log.Println("Error reading response:", err)
		return Response{
			StatusCode: 500,
			Body:       "Internal Server Error",
		}, nil
	}

	// Print the response body
	log.Println("Response:", string(body))
	return Response{
		StatusCode: 200,
		Body:       "Pong",
	}, nil
}

func main() {
	lambda.Start(Handler)
}

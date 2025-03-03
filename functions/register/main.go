package main

import (
	"context"
	"encoding/json"
	"log"

	"github.com/SajjadManafi/android-emulator-serverless/internal/config"
	"github.com/SajjadManafi/android-emulator-serverless/internal/redis"
	"github.com/SajjadManafi/android-emulator-serverless/internal/token"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response events.APIGatewayProxyResponse

var UserService *redis.UserService
var TokenMaker token.Maker

func Handler(request events.APIGatewayProxyRequest) (Response, error) {
	req := request.Body

	// unmarshal json request to user
	user := redis.User{}

	err := json.Unmarshal([]byte(req), &user)
	if err != nil {
		return Response{
			StatusCode: 400,
			Body:       "Invalid request. Please provide a valid JSON payload.",
		}, nil
	}

	// set default balance
	user.Balance = 150

	err = UserService.RegisterUser(context.Background(), user)
	if err != nil {
		if err == redis.ErrUserExists {
			return Response{
				StatusCode: 400,
				Body:       "User already exists. Please choose a different username.",
			}, nil
		}
		return Response{
			StatusCode: 500,
			Body:       "Internal server error. Please try again later.",
		}, nil
	}

	return Response{
		StatusCode: 200,
		Body:       "User successfully registered.",
	}, nil
}

func main() {

	config, err := config.InitConfig()
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}

	redisClient := redis.NewUniversalRedisClient(config.Redis)
	UserService = redis.NewUserService(redisClient)

	TokenMaker, err = token.NewPasetoMaker(config.Token.SecretKey)
	if err != nil {
		log.Fatalf("failed to init token maker: %v", err)
	}

	// log redis ping
	_, err = redisClient.Ping(context.Background()).Result()
	if err != nil {
		log.Fatalf("failed to ping redis: %v", err)
	} else {
		log.Println("redis pinged successfully")
	}

	defer redisClient.Close()

	lambda.Start(Handler)

}

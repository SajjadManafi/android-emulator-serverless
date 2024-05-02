package main

import (
	"context"
	"log"

	"github.com/SajjadManafi/android-emulator-serverless/internal/config"
	"github.com/SajjadManafi/android-emulator-serverless/internal/redis"
	"github.com/SajjadManafi/android-emulator-serverless/internal/token"
	"github.com/aws/aws-lambda-go/events"
	"github.com/aws/aws-lambda-go/lambda"
)

type Response events.APIGatewayProxyResponse

var DebugCounter = 0

var UserService *redis.UserService
var TokenMaker token.Maker

func Handler(request events.APIGatewayProxyRequest) (Response, error) {

	if DebugCounter == 0 {
		DebugCounter++
		UserService.RegisterUser(context.Background(), redis.User{
			Name:     "User-1",
			UserName: "user1",
			Password: "password",
		})

		return Response{
			StatusCode: 200,
			Body:       "user created",
		}, nil
	} else {
		ok, err := UserService.LoginUser(context.Background(), "user1", "password")
		if err != nil {
			return Response{
				StatusCode: 500,
				Body:       err.Error(),
			}, nil
		}

		if !ok {
			return Response{
				StatusCode: 401,
				Body:       "invalid credentials",
			}, nil
		}

		token, err := TokenMaker.CreateAccessToken("user1", "User-1")
		if err != nil {
			return Response{
				StatusCode: 500,
				Body:       err.Error(),
			}, nil
		}

		// return token as authorization in header
		return Response{
			StatusCode: 200,
			Body:       "pong",
			Headers: map[string]string{
				"Authorization": token,
			},
		}, nil
	}

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

	lambda.Start(Handler)
}

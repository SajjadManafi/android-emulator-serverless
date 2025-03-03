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
			Body:       "Invalid request. Please provide valid JSON.",
		}, nil
	}

	// login user
	ok, err := UserService.LoginUser(context.Background(), user.UserName, user.Password)
	if err != nil {
		if err == redis.ErrInvalidCredentials {
			return Response{
				StatusCode: 401,
				Body:       "Invalid credentials. Please check your username and password.",
			}, nil
		} else if err == redis.ErrUserNotFound {
			return Response{
				StatusCode: 404,
				Body:       "User not found. Please register an account.",
			}, nil
		}

		return Response{
			StatusCode: 500,
			Body:       "Internal server error. Please try again later.",
		}, nil

	}

	if !ok {
		return Response{
			StatusCode: 401,
			Body:       "Invalid credentials. Please check your username and password.",
		}, nil
	}

	// create token
	token, err := TokenMaker.CreateAccessToken(user.UserName, user.Name)
	if err != nil {
		return Response{
			StatusCode: 500,
			Body:       "Internal server error. Please try again later.",
		}, nil
	}

	// return token as authorization in header
	return Response{
		StatusCode: 200,
		Body:       "Successful login.",
		Headers: map[string]string{
			"Authorization":                 token,
			"Access-Control-Expose-Headers": "Authorization",
		},
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

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

var UserService *redis.UserService
var TokenMaker token.Maker
var AndroidService *redis.AndroidService

func Handler(request events.APIGatewayProxyRequest) (Response, error) {
	req := request.Body
	log.Println("request: ", req)
	head := request.Headers
	log.Println("headers: ", head)

	ctx := context.Background()

	// check Authorization in headers
	auth, ok := head["Authorization"]
	if !ok {
		return Response{
			StatusCode: 401,
			Body:       "unauthorized",
		}, nil
	}

	// validate token
	claims, err := TokenMaker.VerifyAccessToken(auth)
	if err != nil {
		return Response{
			StatusCode: 401,
			Body:       "unauthorized",
		}, nil
	}

	// get device from redis
	err = AndroidService.DeleteAndroid(ctx, claims.Username+"-Device")
	if err != nil {
		if err == redis.ErrNotFound {
			return Response{
				StatusCode: 404,
				Body:       "device not found",
			}, nil
		}
		return Response{
			StatusCode: 500,
			Body:       "internal server error",
		}, nil
	}

	// delete device from redis
	err = AndroidService.DeleteAndroid(ctx, claims.Username+"-Device")
	if err != nil {
		return Response{
			StatusCode: 500,
			Body:       "internal server error",
		}, nil
	}

	return Response{
		StatusCode: 200,
		Body:       "device deleted",
	}, nil

}

func main() {

	config, err := config.InitConfig()
	if err != nil {
		log.Fatalf("failed to init config: %v", err)
	}

	redisClient := redis.NewUniversalRedisClient(config.Redis)
	UserService = redis.NewUserService(redisClient)
	AndroidService = redis.NewAndroidService(redisClient)

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

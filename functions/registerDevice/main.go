package main

import (
	"context"
	"encoding/json"
	"log"
	"os/exec"
	"strconv"
	"time"

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

	// unmarshal json request to android
	android := redis.Android{}

	err := json.Unmarshal([]byte(req), &android)
	if err != nil {
		return Response{
			StatusCode: 400,
			Body:       "invalid request",
		}, nil
	}

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

	// random port
	port, err := AndroidService.GetRandomPort(ctx)
	if err != nil {
		return Response{
			StatusCode: 500,
			Body:       err.Error(),
		}, nil
	}

	// TODO: check if user has registered

	android = redis.Android{
		DeviceID:       claims.Username + "-Device",
		Username:       claims.Username,
		Port:           port,
		StartTimestamp: time.Now().Unix(),
		AndroidAPI:     android.AndroidAPI,
		DeviceName:     android.DeviceName,
	}

	// register android
	err = AndroidService.RegisterAndroid(ctx, android)
	if err != nil {
		if err == redis.ErrAlreadyExists {
			return Response{
				StatusCode: 409,
				Body:       "android already exists",
			}, nil
		}

		return Response{
			StatusCode: 500,
			Body:       err.Error(),
		}, nil
	}

	// return success in json with android device
	res, err := json.Marshal(android)
	if err != nil {
		return Response{
			StatusCode: 500,
			Body:       err.Error(),
		}, nil
	}

	// Construct the Docker command
	portStr := strconv.Itoa(port) // Convert port to string
	dockerCmd := "docker run -d -p " + portStr + ":" + portStr + " -e EMULATOR_DEVICE=" + android.DeviceName + " -e WEB_VNC=true --device /dev/kvm --name android-container budtmo/docker-android:" + android.AndroidAPI

	// Execute the Docker command
	cmd := exec.Command("sh", "-c", dockerCmd)
	if err := cmd.Run(); err != nil {
		return Response{
			StatusCode: 500,
			Body:       err.Error(),
		}, nil

	}

	return Response{
		StatusCode: 201,
		Headers: map[string]string{
			"Content-Type": "application/json",
		},
		Body: string(res),
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

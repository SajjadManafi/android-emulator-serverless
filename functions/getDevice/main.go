package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

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
	head := request.Headers

	ctx := context.Background()

	// check Authorization in headers
	auth, ok := head["Authorization"]
	if !ok {
		return Response{
			StatusCode: 401,
			Body:       "Unauthorized: Missing Authorization header",
		}, nil
	}

	// validate token
	claims, err := TokenMaker.VerifyAccessToken(auth)
	if err != nil {
		return Response{
			StatusCode: 401,
			Body:       "Unauthorized: Invalid token",
		}, nil
	}

	// get device from redis
	androidData, err := AndroidService.GetAndroid(ctx, claims.Username+"-Device")
	if err != nil {
		if err == redis.ErrNotFound {
			return Response{
				StatusCode: 404,
				Body:       "Device not found",
			}, nil
		}
		return Response{
			StatusCode: 500,
			Body:       "Internal server error",
		}, nil
	}

	// get device status
	status, err := getDeviceStatus(androidData.DeviceID)
	if err != nil {
		return Response{
			StatusCode: 500,
			Body:       "Internal server error",
		}, nil
	}

	androidData.Status = status

	// marshal android data
	androidDataJSON, err := json.Marshal(androidData)
	if err != nil {
		return Response{
			StatusCode: 500,
			Body:       "Internal server error",
		}, nil
	}

	return Response{
		StatusCode: 200,
		Headers: map[string]string{
			"Content-Type":                  "application/json",
			"Access-Control-Expose-Headers": "Authorization",
		},
		Body: string(androidDataJSON),
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

func getDeviceStatus(containerName string) (string, error) {
	baseURL := "http://172.17.0.1:8080/device-status" // Replace with your actual server address and port
	params := url.Values{}
	params.Add("containerName", containerName)
	url := fmt.Sprintf("%s?%s", baseURL, params.Encode())

	resp, err := http.Get(url)
	if err != nil {
		return "", fmt.Errorf("error making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("server returned non-200 status: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("error reading response body: %w", err)
	}

	return string(body), nil
}

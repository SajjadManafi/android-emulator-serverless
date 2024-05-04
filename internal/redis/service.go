package redis

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"

	"github.com/redis/go-redis/v9"
)

type AndroidService struct {
	redisClient redis.UniversalClient
}

type Android struct {
	DeviceID       string `json:"device_id" redis:"device_id"`
	Username       string `json:"username" redis:"username"`
	Port           int    `json:"port" redis:"port"`
	StartTimestamp int64  `json:"start_timestamp" redis:"start_timestamp"`
	AndroidAPI     string `json:"android_api" redis:"android_api"`
	DeviceName     string `json:"device_name" redis:"device_name"`
	Status         string `json:"status" redis:"status"`
}

var UsedPorts = "used_ports"

var ErrAlreadyExists = errors.New("android already exists")
var ErrNotFound = errors.New("android not found")

func NewAndroidService(redisClient redis.UniversalClient) *AndroidService {
	return &AndroidService{redisClient: redisClient}
}

// RegisterAndroid attempts to register a new android in Redis.
func (s *AndroidService) RegisterAndroid(ctx context.Context, android Android) error {
	// Check if android already exists
	exists, err := s.redisClient.Exists(ctx, android.DeviceID).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return ErrAlreadyExists
	}

	android.Status = "pending"

	// Serialize Android struct to JSON
	androidData, err := json.Marshal(android)
	if err != nil {
		return err
	}

	// Save android data in Redis
	if err := s.redisClient.Set(ctx, android.DeviceID, androidData, 0).Err(); err != nil {
		return err
	}

	return nil

}

// GetAndroid returns the android with the given deviceID.
func (s *AndroidService) GetAndroid(ctx context.Context, deviceID string) (Android, error) {
	result, err := s.redisClient.Get(ctx, deviceID).Result()
	if err == redis.Nil {
		return Android{}, ErrNotFound
	} else if err != nil {
		return Android{}, err
	}

	var android Android
	if err := json.Unmarshal([]byte(result), &android); err != nil {
		return Android{}, err
	}

	return android, nil
}

// DeleteAndroid deletes the android with the given deviceID.
func (s *AndroidService) DeleteAndroid(ctx context.Context, deviceID string) error {
	if err := s.redisClient.Del(ctx, deviceID).Err(); err != nil {
		return err
	}

	return nil
}

// UsePort marks a port as used in Redis.
func (s *AndroidService) UsePort(ctx context.Context, port int) error {
	// Convert the port to a string because Redis works with strings
	portStr := fmt.Sprintf("%d", port)
	// Add the port to a Redis set named UsedPorts
	_, err := s.redisClient.SAdd(ctx, UsedPorts, portStr).Result()
	return err
}

// IsPortUsed checks if a port is marked as used in Redis.
func (s *AndroidService) IsPortUsed(ctx context.Context, port int) (bool, error) {
	portStr := fmt.Sprintf("%d", port)
	// Check if the port is in the "used_ports" set
	return s.redisClient.SIsMember(ctx, UsedPorts, portStr).Result()
}

// FreePort marks a port as no longer used in Redis.
func (s *AndroidService) FreePort(ctx context.Context, port int) error {
	portStr := fmt.Sprintf("%d", port)
	// Remove the port from the "used_ports" set
	_, err := s.redisClient.SRem(ctx, UsedPorts, portStr).Result()
	return err
}

// GetRandomPort returns a random port that is not marked as used in Redis.
func (s *AndroidService) GetRandomPort(ctx context.Context) (int, error) {
	// Get a random port number
	port := 0
	for {
		port = 1024 + rand.Intn(65536-1024)
		// Check if the port is used
		used, err := s.IsPortUsed(ctx, port)
		if err != nil {
			return 0, err
		}
		if !used {
			break
		}
	}

	// Mark the port as used
	if err := s.UsePort(ctx, port); err != nil {
		return 0, err
	}

	return port, nil
}

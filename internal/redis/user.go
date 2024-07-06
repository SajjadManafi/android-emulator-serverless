package redis

import (
	"context"
	"encoding/json"
	"errors"

	"github.com/redis/go-redis/v9"
)

type UserService struct {
	redisClient redis.UniversalClient
}

type User struct {
	Name     string `json:"name" redis:"name"`
	UserName string `json:"username" redis:"username"`
	Password string `json:"password" redis:"password"`
	Balance  int    `json:"balance" redis:"balance"`
}

var ErrUserExists = errors.New("user already exists")
var ErrUserNotFound = errors.New("user not found")
var ErrInvalidCredentials = errors.New("invalid credentials")

func NewUserService(redisClient redis.UniversalClient) *UserService {
	return &UserService{redisClient: redisClient}
}

// RegisterUser attempts to register a new user in Redis.
func (s *UserService) RegisterUser(ctx context.Context, user User) error {
	// Check if user already exists
	exists, err := s.redisClient.Exists(ctx, user.UserName).Result()
	if err != nil {
		return err
	}
	if exists > 0 {
		return ErrUserExists
	}

	// Serialize User struct to JSON
	userData, err := json.Marshal(user)
	if err != nil {
		return err
	}

	// Save user data in Redis
	if err := s.redisClient.Set(ctx, user.UserName, userData, 0).Err(); err != nil {
		return err
	}

	return nil
}

// LoginUser checks if the user exists and the password is correct.
func (s *UserService) LoginUser(ctx context.Context, username, password string) (bool, error) {
	result, err := s.redisClient.Get(ctx, username).Result()
	if err == redis.Nil {
		return false, ErrUserNotFound
	} else if err != nil {
		return false, err
	}

	var user User
	if err := json.Unmarshal([]byte(result), &user); err != nil {
		return false, err
	}

	// Compare the provided password with the one stored in Redis
	if user.Password != password {
		return false, ErrInvalidCredentials
	}

	return true, nil
}

// GetUser retrieves a user from Redis by username.
func (s *UserService) GetUser(ctx context.Context, username string) (User, error) {
	result, err := s.redisClient.Get(ctx, username).Result()
	if err == redis.Nil {
		return User{}, ErrUserNotFound
	} else if err != nil {
		return User{}, err
	}

	var user User
	if err := json.Unmarshal([]byte(result), &user); err != nil {
		return User{}, err
	}

	return user, nil
}

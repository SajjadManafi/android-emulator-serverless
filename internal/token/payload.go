package token

import (
	"errors"
	"time"
)

// Different types of errors returned by the VerifyToken function
var (
	ErrExpiredToken = errors.New("token has expired")
	ErrInvalidToken = errors.New("token is invalid")
)

var (
	AccessTokenDuration = 24 * 60 * time.Hour
)

// AccessPayload contains the access token payload data of the token
type AccessPayload struct {
	Username  string    `json:"username"`
	Name      string    `json:"name"`
	IssuedAt  time.Time `json:"issued_at"`
	ExpiredAt time.Time `json:"expired_at"`
}

func NewAccessPayload(username string, name string) (*AccessPayload, error) {
	currentTime := time.Now()
	return &AccessPayload{
		Username:  username,
		Name:      name,
		IssuedAt:  currentTime,
		ExpiredAt: currentTime.Add(AccessTokenDuration),
	}, nil
}

// Valid checks if the token has expired
func (p *AccessPayload) Valid() error {
	if time.Now().After(p.ExpiredAt) {
		return ErrExpiredToken
	}
	return nil
}

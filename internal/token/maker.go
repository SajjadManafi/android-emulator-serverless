package token

type Config struct {
	// Secret key used for signing.
	SecretKey string `mapstructure:"secret_key"`
}

type Maker interface {
	// CreateAccessToken creates a new access token for a specific username
	CreateAccessToken(username string, name string) (string, error)

	// VerifyToken verifies a token
	VerifyAccessToken(token string) (*AccessPayload, error)
}

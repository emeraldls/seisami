package central

import (
	"fmt"
	"os"
	"time"
)

// Config bundles the runtime configuration required for the auth HTTP service.
type Config struct {
	DatabaseURL          string
	JWTSecret            string
	JWTExpiration        time.Duration
	ResetTokenExpiration time.Duration
	HTTPAddr             string
}

// LoadConfigFromEnv reads the required configuration from environment variables.
func LoadConfigFromEnv() (Config, error) {
	dbURL := os.Getenv("DATABASE_URL")
	if dbURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL must be set")
	}

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return Config{}, fmt.Errorf("JWT_SECRET must be set")
	}

	jwtTTL := 24 * time.Hour
	if v := os.Getenv("AUTH_JWT_TTL"); v != "" {
		dur, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid AUTH_JWT_TTL: %w", err)
		}
		jwtTTL = dur
	}

	resetTTL := time.Hour
	if v := os.Getenv("AUTH_RESET_TOKEN_TTL"); v != "" {
		dur, err := time.ParseDuration(v)
		if err != nil {
			return Config{}, fmt.Errorf("invalid AUTH_RESET_TOKEN_TTL: %w", err)
		}
		resetTTL = dur
	}

	addr := os.Getenv("HTTP_ADDR")
	if addr == "" {
		addr = ":8080"
	}

	return Config{
		DatabaseURL:          dbURL,
		JWTSecret:            secret,
		JWTExpiration:        jwtTTL,
		ResetTokenExpiration: resetTTL,
		HTTPAddr:             addr,
	}, nil
}

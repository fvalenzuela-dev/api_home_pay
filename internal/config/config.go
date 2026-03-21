package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	Port           string
	ClerkSecretKey string
	DatabaseURL    string
	GinMode        string
}

var configInstance *Config

func Load() (*Config, error) {
	if configInstance != nil {
		return configInstance, nil
	}

	if err := godotenv.Load(); err != nil {
		fmt.Println("Warning: .env file not found, using environment variables")
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	clerkSecretKey := os.Getenv("CLERK_SECRET_KEY")
	if clerkSecretKey == "" {
		return nil, fmt.Errorf("CLERK_SECRET_KEY is required")
	}

	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return nil, fmt.Errorf("DATABASE_URL is required")
	}

	ginMode := os.Getenv("GIN_MODE")
	if ginMode == "" {
		ginMode = "debug"
	}

	configInstance = &Config{
		Port:           port,
		ClerkSecretKey: clerkSecretKey,
		DatabaseURL:    databaseURL,
		GinMode:        ginMode,
	}

	return configInstance, nil
}

func GetConfig() (*Config, error) {
	if configInstance != nil {
		return configInstance, nil
	}
	return Load()
}

package config

import (
	"fmt"
	"os"
)

type Config struct {
	HTTPAddr           string
	DatabaseURL        string
	SQSQueueURL        string
	AWSRegion          string
	AWSEndpointURL     string
	FunctionSandboxARN string
}

func Load() (Config, error) {
	cfg := Config{
		HTTPAddr:           getEnvOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL:        os.Getenv("DATABASE_URL"),
		SQSQueueURL:        os.Getenv("SQS_QUEUE_URL"),
		AWSRegion:          getEnvOrDefault("AWS_REGION", "ap-south-1"),
		AWSEndpointURL:     os.Getenv("AWS_ENDPOINT_URL"),
		FunctionSandboxARN: os.Getenv("FUNCTION_SANDBOX_ARN"),
	}

	if cfg.DatabaseURL == "" {
		return Config{}, fmt.Errorf("DATABASE_URL is required")
	}
	return cfg, nil
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

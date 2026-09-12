package config

import (
	"fmt"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	HTTPAddr    string
	DatabaseURL string
	LogLevel    string
	SQS         SQSConfig
	Lambda      LambdaConfig
}

type SQSConfig struct {
	QueueURL    string
	Region      string
	EndpointURL string
}

type LambdaConfig struct {
	FunctionSandboxARN string
	EndpointURL        string
	Region             string
}

func Load() (Config, error) {
	// Load .env for local development. It is okay if the file is missing.
	// In production, Terraform sets the environment variables instead.
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("load local .env: %w", err)
	}

	cfg := Config{
		HTTPAddr:    getEnvOrDefault("HTTP_ADDR", ":8080"),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		LogLevel:    getEnvOrDefault("LOG_LEVEL", "info"),
		SQS: SQSConfig{
			QueueURL:    os.Getenv("SQS_QUEUE_URL"),
			Region:      os.Getenv("AWS_REGION"),
			EndpointURL: os.Getenv("AWS_ENDPOINT_URL"),
		},
		Lambda: LambdaConfig{
			FunctionSandboxARN: os.Getenv("FUNCTION_SANDBOX_ARN"),
			EndpointURL:        os.Getenv("LAMBDA_ENDPOINT_URL"),
			Region:             os.Getenv("AWS_REGION"),
		},
	}

	return cfg, nil
}

func getEnvOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

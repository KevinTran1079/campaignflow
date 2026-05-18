package config

import (
	"fmt"
	"os"
	"strconv"
	"time"

	"github.com/joho/godotenv"
)

const (
	defaultAppEnvironment  = "development"
	defaultServerAddress   = "127.0.0.1"
	defaultServerPort      = "8080"
	defaultLogLevel        = "info"
	defaultStartupTimeout  = "10s"
	defaultShutdownTimeout = "10s"
	defaultDatabaseURL     = "postgres://campaignflow:campaignflow@localhost:5432/campaignflow?sslmode=disable"
)

var allowedAppEnvironments = map[string]struct{}{
	"development": {},
	"production":  {},
	"test":        {},
}

var allowedLogLevels = map[string]struct{}{
	"debug": {},
	"error": {},
	"info":  {},
	"warn":  {},
}

type Config struct {
	AppEnvironment  string
	ServerAddress   string
	ServerPort      int
	LogLevel        string
	StartupTimeout  time.Duration
	ShutdownTimeout time.Duration
	DatabaseURL     string
}

func LoadConfig() (*Config, error) {
	if err := godotenv.Load(); err != nil {
		if !os.IsNotExist(err) {
			return nil, fmt.Errorf("load .env: %w", err)
		}
	}

	appEnvironment := envOrDefault("APP_ENVIRONMENT", defaultAppEnvironment)
	if _, ok := allowedAppEnvironments[appEnvironment]; !ok {
		return nil, fmt.Errorf("APP_ENVIRONMENT must be one of development, test, production, got %q", appEnvironment)
	}

	serverAddress := envOrDefault("SERVER_ADDRESS", defaultServerAddress)
	serverPort, err := strconv.Atoi(envOrDefault("SERVER_PORT", defaultServerPort))
	if err != nil {
		return nil, fmt.Errorf("unable to convert SERVER_PORT: %w", err)
	}
	if serverPort < 1 || serverPort > 65535 {
		return nil, fmt.Errorf("SERVER_PORT must be between 1 and 65535, got %d", serverPort)
	}

	logLevel := envOrDefault("LOG_LEVEL", defaultLogLevel)
	if _, ok := allowedLogLevels[logLevel]; !ok {
		return nil, fmt.Errorf("LOG_LEVEL must be one of debug, info, warn, error, got %q", logLevel)
	}

	startupTimeout, err := time.ParseDuration(envOrDefault("STARTUP_TIMEOUT", defaultStartupTimeout))
	if err != nil {
		return nil, fmt.Errorf("unable to parse STARTUP_TIMEOUT: %w", err)
	}

	if startupTimeout <= 0 {
		return nil, fmt.Errorf("STARTUP_TIMEOUT must be greater than 0, got %s", startupTimeout)
	}

	shutdownTimeout, err := time.ParseDuration(envOrDefault("SHUTDOWN_TIMEOUT", defaultShutdownTimeout))
	if err != nil {
		return nil, fmt.Errorf("unable to parse SHUTDOWN_TIMEOUT: %w", err)
	}

	if shutdownTimeout <= 0 {
		return nil, fmt.Errorf("SHUTDOWN_TIMEOUT must be greater than 0, got %s", shutdownTimeout)
	}

	databaseURL := envOrDefault("DATABASE_URL", defaultDatabaseURL)

	return &Config{
		AppEnvironment:  appEnvironment,
		ServerAddress:   serverAddress,
		ServerPort:      serverPort,
		LogLevel:        logLevel,
		StartupTimeout:  startupTimeout,
		ShutdownTimeout: shutdownTimeout,
		DatabaseURL:     databaseURL,
	}, nil
}

func envOrDefault(key string, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}

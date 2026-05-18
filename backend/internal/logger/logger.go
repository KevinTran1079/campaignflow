package logger

import (
	"fmt"
	"log/slog"
	"os"
)

const defaultServiceName = "campaignflow-api"

type Options struct {
	Level          string
	AppEnvironment string
	ServiceName    string
}

func New(options Options) (*slog.Logger, error) {
	level, err := parseLevel(options.Level)
	if err != nil {
		return nil, err
	}

	serviceName := options.ServiceName
	if serviceName == "" {
		serviceName = defaultServiceName
	}
	handlerOptions := &slog.HandlerOptions{
		Level: level,
	}

	var handler slog.Handler
	switch options.AppEnvironment {
	case "development", "test":
		handler = slog.NewTextHandler(os.Stdout, handlerOptions)
	case "production":
		handler = slog.NewJSONHandler(os.Stdout, handlerOptions)
	default:
		return nil, fmt.Errorf("unknown environment %q", options.AppEnvironment)
	}

	return slog.New(handler).With(
		"service", serviceName,
		"environment", options.AppEnvironment,
	), nil
}

func parseLevel(level string) (slog.Level, error) {
	switch level {
	case "debug":
		return slog.LevelDebug, nil
	case "info":
		return slog.LevelInfo, nil
	case "warn":
		return slog.LevelWarn, nil
	case "error":
		return slog.LevelError, nil
	default:
		return slog.LevelInfo, fmt.Errorf("unknown log level %q", level)
	}
}

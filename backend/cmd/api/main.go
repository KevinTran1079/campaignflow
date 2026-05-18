package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"github.com/KevinTran1079/campaignflow/internal/config"
	"github.com/KevinTran1079/campaignflow/internal/db"
	"github.com/KevinTran1079/campaignflow/internal/logger"
	"github.com/KevinTran1079/campaignflow/internal/server"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "unable to start application: %v\n", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.LoadConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	loggerOptions := logger.Options{
		Level:          cfg.LogLevel,
		AppEnvironment: cfg.AppEnvironment,
		ServiceName:    "campaignflow-api",
	}

	appLogger, err := logger.New(loggerOptions)
	if err != nil {
		return fmt.Errorf("initialize logger: %w", err)
	}

	startupCtx, cancel := context.WithTimeout(context.Background(), cfg.StartupTimeout)
	defer cancel()

	pg, err := db.NewPG(startupCtx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("initialize db: %w", err)
	}
	defer pg.Close()

	serverOptions := server.Options{
		Address: cfg.ServerAddress,
		Port:    cfg.ServerPort,
		Logger:  appLogger,
		DB:      pg,
	}

	srv := server.NewServer(serverOptions)

	serverErr := make(chan error, 1)
	go func() {
		appLogger.Info("starting server", "address", cfg.ServerAddress, "port", cfg.ServerPort)
		serverErr <- srv.ListenAndServe()
	}()

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	select {
	case err := <-serverErr:
		if err != nil {
			return fmt.Errorf("start server: %w", err)
		}
		return nil
	case <-shutdownSignal.Done():
		appLogger.Info("shutting down server")
	}

	shutdownCtx, cancel := context.WithTimeout(context.Background(), cfg.ShutdownTimeout)
	defer cancel()

	if err := srv.Shutdown(shutdownCtx); err != nil {
		return fmt.Errorf("shutdown server: %w", err)
	}

	if err := <-serverErr; err != nil {
		return fmt.Errorf("stop server: %w", err)
	}

	appLogger.Info("server stopped")
	return nil
}

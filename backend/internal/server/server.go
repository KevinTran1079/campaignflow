package server

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"strconv"
)

type Options struct {
	Address string
	Port    int
	Logger  *slog.Logger
	DB      Pinger
}

type Pinger interface {
	Ping(ctx context.Context) error
}

type Server struct {
	httpServer *http.Server
	mux        *http.ServeMux
	logger     *slog.Logger
	db         Pinger
}

func New(options Options) *Server {
	mux := http.NewServeMux()

	s := &Server{
		mux:    mux,
		logger: options.Logger,
		db:     options.DB,
	}

	s.routes()

	s.httpServer = &http.Server{
		Addr:    net.JoinHostPort(options.Address, strconv.Itoa(options.Port)),
		Handler: mux,
	}

	return s
}

func (s *Server) ListenAndServe() error {
	err := s.httpServer.ListenAndServe()
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}

func (s *Server) Shutdown(ctx context.Context) error {
	return s.httpServer.Shutdown(ctx)
}

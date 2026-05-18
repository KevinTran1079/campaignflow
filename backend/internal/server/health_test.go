package server

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

type fakePinger struct {
	err error
}

func (p fakePinger) Ping(ctx context.Context) error {
	return p.err
}

func TestHealthz(t *testing.T) {
	srv := NewServer(Options{})

	rec := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	srv.httpServer.Handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusNoContent {
		t.Fatalf("expected status %d, got %d", http.StatusNoContent, rec.Code)
	}
}

func TestReadyz(t *testing.T) {
	tests := []struct {
		name string
		db   Pinger
		want int
	}{
		{
			name: "database ready",
			db:   fakePinger{},
			want: http.StatusNoContent,
		},
		{
			name: "database unavailable",
			db:   fakePinger{err: errors.New("database unavailable")},
			want: http.StatusServiceUnavailable,
		},
		{
			name: "database not configured",
			db:   nil,
			want: http.StatusServiceUnavailable,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			srv := NewServer(Options{DB: tt.db})

			rec := httptest.NewRecorder()
			req := httptest.NewRequest(http.MethodGet, "/readyz", nil)

			srv.httpServer.Handler.ServeHTTP(rec, req)

			if rec.Code != tt.want {
				t.Fatalf("expected status %d, got %d", tt.want, rec.Code)
			}
		})
	}
}

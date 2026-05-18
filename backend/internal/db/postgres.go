package db

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Postgres struct {
	db *pgxpool.Pool
}

func NewPG(ctx context.Context, connString string) (*Postgres, error) {
	dbpool, err := pgxpool.New(ctx, connString)
	if err != nil {
		return nil, fmt.Errorf("create db pool: %w", err)
	}

	pgInstance := &Postgres{db: dbpool}

	if err := pgInstance.Ping(ctx); err != nil {
		dbpool.Close()
		return nil, err
	}

	return pgInstance, nil
}

func (pg *Postgres) Ping(ctx context.Context) error {
	if err := pg.Ping(ctx); err != nil {
		return fmt.Errorf("ping db: %w", err)
	}
	return nil
}

func (pg *Postgres) Close() {
	pg.db.Close()
}

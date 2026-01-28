package db

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type DB struct{
	Pool *pgxpool.Pool
}

func Connect(ctx context.Context, dsn string) (*DB, error){
	cfg, err := pgxpool.ParseConfig(dsn);
	if err != nil{
		return nil, err
	}

	// Keep pool small for MVP; adjust after load testing.
	cfg.MaxConns = 10
	cfg.MinConns = 1
	cfg.MaxConnIdleTime = 5 * time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil{
		return nil, err
	}

	//quick ping
	if err := pool.Ping(ctx); err != nil{
		pool.Close()
		return nil, err
	}

	return &DB{Pool: pool}, nil
}

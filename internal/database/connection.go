package database

import (
	"context"
	"errors"
	"os"
	"sync"

	"github.com/jackc/pgx/v5/pgxpool"
)

var (
	pool                *pgxpool.Pool
	once                sync.Once
	ConnectionPoolError error
)

func GetConnectionPool() (*pgxpool.Pool, error) {
	if ConnectionPoolError != nil {
		once = sync.Once{}
		ConnectionPoolError = nil
	}

	once.Do(func() {
		databaseURL := os.Getenv("DATABASE_URL")
		if databaseURL == "" {
			databaseURL = "postgres://postgres:root@localhost:5432/rinha?pool_min_conns=10&pool_max_conns=50"
		}

		var config *pgxpool.Config
		config, ConnectionPoolError = pgxpool.ParseConfig(databaseURL)
		if ConnectionPoolError != nil {
			return
		}
		pool, ConnectionPoolError = pgxpool.NewWithConfig(context.Background(), config)
	})

	if ConnectionPoolError != nil {
		ConnectionPoolError = errors.New(ConnectionPoolError.Error())
		return nil, ConnectionPoolError
	}

	ConnectionPoolError = pool.Ping(context.Background())
	if ConnectionPoolError != nil {
		ConnectionPoolError = errors.New(ConnectionPoolError.Error())
		return nil, ConnectionPoolError
	}

	return pool, nil
}

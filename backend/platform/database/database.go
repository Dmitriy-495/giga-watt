// Package database отвечает за подключение backend Giga Watt к PostgreSQL.
//
// Используется pgx без ORM (см. docs/adr/ADR-0001-sql-first.md).
package database

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/Dmitriy-495/giga-watt/backend/config"
)

// Connect устанавливает пул соединений с PostgreSQL и проверяет
// его доступность через Ping.
func Connect(ctx context.Context, cfg config.Database) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, cfg.DSN())
	if err != nil {
		return nil, fmt.Errorf("database: create pool: %w", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("database: ping: %w", err)
	}

	return pool, nil
}

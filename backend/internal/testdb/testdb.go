// Package testdb — общий хелпер для интеграционных тестов Foundation
// Hardening, подключающихся к реальной PostgreSQL (см. ADR-0001:
// SQL-first — инварианты проверяются в БД, поэтому и тестируются против
// настоящей БД, а не мока).
package testdb

import (
	"context"
	"os"
	"testing"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// Connect открывает пул соединений с тестовой PostgreSQL.
//
// DSN берётся из переменной окружения TEST_DATABASE_URL, например:
//
//	export TEST_DATABASE_URL="postgres://tda:tda@localhost:5432/giga_watt?sslmode=disable"
//
// Если переменная не задана или БД недоступна, тест пропускается
// (t.Skip), а не падает — тесты Foundation Hardening требуют реальной
// PostgreSQL с применёнными миграциями и не должны ломать `go test ./...`
// в окружении, где такой БД нет.
func Connect(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv("TEST_DATABASE_URL")
	if dsn == "" {
		t.Skip("testdb: TEST_DATABASE_URL is not set, skipping integration test")
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, dsn)
	if err != nil {
		t.Skipf("testdb: cannot create pool: %v", err)
	}

	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		t.Skipf("testdb: cannot reach database (%s): %v", dsn, err)
	}

	t.Cleanup(pool.Close)

	return pool
}

// WithTx открывает транзакцию и регистрирует её откат по завершении
// теста — тест не оставляет следов в БД независимо от результата
// (аналог PASS/FAIL, описанного в задании: воспроизводимо и без побочных
// эффектов).
func WithTx(t *testing.T, pool *pgxpool.Pool) pgx.Tx {
	t.Helper()

	ctx := context.Background()

	tx, err := pool.Begin(ctx)
	if err != nil {
		t.Fatalf("testdb: begin tx: %v", err)
	}

	t.Cleanup(func() {
		_ = tx.Rollback(ctx)
	})

	return tx
}

// Try выполняет fn в рамках SAVEPOINT и автоматически откатывается до
// него при ошибке.
//
// Postgres переводит всю транзакцию в состояние aborted после первой же
// ошибки внутри неё — без этого негативные тесты (ожидающие ошибку от
// БД) не могли бы продолжать использовать ту же транзакцию для
// последующих проверок в том же тесте.
func Try(ctx context.Context, tx pgx.Tx, fn func() error) error {
	if _, err := tx.Exec(ctx, "SAVEPOINT sp"); err != nil {
		return err
	}

	err := fn()

	if err != nil {
		_, _ = tx.Exec(ctx, "ROLLBACK TO SAVEPOINT sp")
	} else {
		_, _ = tx.Exec(ctx, "RELEASE SAVEPOINT sp")
	}

	return err
}

// Command fixtures загружает воспроизводимые начальные данные
// (development/test/demo fixtures) из Excel-файла в чистую БД Giga Watt.
//
// Это НЕ production-импорт и не часть HTTP API — отдельный инструмент
// для получения одинаковой, воспроизводимой картины цифрового двойника
// после `migrate ... up` на чистой базе (см.
// docs/DEVELOPMENT.md, раздел "Fixtures").
//
// Запуск (из каталога backend/, после применения миграций):
//
//	go run ./cmd/fixtures -file ../fixtures/initial-data.xlsx
package main

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/xuri/excelize/v2"

	"github.com/Dmitriy-495/giga-watt/backend/config"
)

func main() {
	file := flag.String("file", "../fixtures/initial-data.xlsx",
		"путь к Excel-файлу с исходными данными")
	flag.Parse()

	if err := run(*file); err != nil {
		fmt.Fprintln(os.Stderr, "fixtures: "+err.Error())
		os.Exit(1)
	}
}

func run(path string) error {
	if _, err := os.Stat(path); err != nil {
		return fmt.Errorf("файл не найден: %s (%w)", path, err)
	}

	f, err := excelize.OpenFile(path)
	if err != nil {
		return fmt.Errorf("не удалось открыть %s: %w", path, err)
	}
	defer f.Close()

	cfg, err := config.Load()
	if err != nil {
		return fmt.Errorf("конфигурация: %w", err)
	}

	ctx := context.Background()

	pool, err := pgxpool.New(ctx, cfg.Database.DSN())
	if err != nil {
		return fmt.Errorf("подключение к БД: %w", err)
	}
	defer pool.Close()

	if err := pool.Ping(ctx); err != nil {
		return fmt.Errorf("БД недоступна: %w", err)
	}

	tx, err := pool.Begin(ctx)
	if err != nil {
		return fmt.Errorf("не удалось начать транзакцию: %w", err)
	}

	committed := false

	defer func() {
		if !committed {
			_ = tx.Rollback(ctx)
		}
	}()

	if err := load(ctx, tx, f); err != nil {
		return fmt.Errorf("загрузка отменена (транзакция откачена): %w", err)
	}

	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("не удалось зафиксировать транзакцию: %w", err)
	}

	committed = true

	fmt.Println("fixtures: загрузка успешно завершена")

	return nil
}

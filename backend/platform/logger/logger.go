// Package logger настраивает базовое логирование backend Giga Watt
// на основе стандартного пакета log/slog.
package logger

import (
	"log/slog"
	"os"
)

// New создаёт логгер в зависимости от окружения приложения.
//
// В development используется текстовый вывод и уровень Debug,
// в остальных окружениях — JSON и уровень Info.
func New(env string) *slog.Logger {
	level := slog.LevelInfo
	var handler slog.Handler

	if env == "development" {
		level = slog.LevelDebug
		handler = slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	} else {
		handler = slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: level})
	}

	return slog.New(handler)
}

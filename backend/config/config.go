// Package config содержит конфигурацию backend Giga Watt.
//
// Конфигурация загружается из переменных окружения (см. .env.example).
package config

import (
	"fmt"

	"github.com/ilyakaznacheev/cleanenv"
)

// Config — конфигурация backend.
type Config struct {
	App      App
	HTTP     HTTP
	Database Database
}

// App — общие параметры приложения.
type App struct {
	Name string `env:"APP_NAME" env-default:"Giga Watt"`
	Env  string `env:"APP_ENV" env-default:"development"`
}

// HTTP — параметры HTTP-сервера.
type HTTP struct {
	Port string `env:"HTTP_PORT" env-default:"8080"`
}

// Database — параметры подключения к PostgreSQL.
type Database struct {
	Host     string `env:"DB_HOST" env-default:"localhost"`
	Port     string `env:"DB_PORT" env-default:"5432"`
	Name     string `env:"DB_NAME" env-default:"giga_watt"`
	User     string `env:"DB_USER" env-default:"giga_watt"`
	Password string `env:"DB_PASSWORD" env-default:""`
}

// DSN возвращает строку подключения к PostgreSQL для pgx.
func (d Database) DSN() string {
	return fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s?sslmode=disable",
		d.User, d.Password, d.Host, d.Port, d.Name,
	)
}

// Load читает конфигурацию из переменных окружения.
//
// Если существует файл .env, cleanenv также прочитает его.
func Load() (*Config, error) {
	var cfg Config

	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		// .env не обязателен: в production переменные окружения
		// задаются платформой напрямую.
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			return nil, fmt.Errorf("config: %w", err)
		}
	}

	return &cfg, nil
}

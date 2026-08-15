package main

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Dmitriy-495/giga-watt/backend/config"
	"github.com/Dmitriy-495/giga-watt/backend/modules/employee"
	"github.com/Dmitriy-495/giga-watt/backend/modules/organization"
	"github.com/Dmitriy-495/giga-watt/backend/modules/position"
	"github.com/Dmitriy-495/giga-watt/backend/platform/database"
	"github.com/Dmitriy-495/giga-watt/backend/platform/httpserver"
	"github.com/Dmitriy-495/giga-watt/backend/platform/logger"
)

func main() {
	cfg, err := config.Load()
	if err != nil {
		// Логгер ещё не создан: конфигурация обязательна для его настройки.
		os.Stderr.WriteString("config: " + err.Error() + "\n")
		os.Exit(1)
	}

	log := logger.New(cfg.App.Env)

	connectCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	db, err := database.Connect(connectCtx, cfg.Database)
	if err != nil {
		log.Error("failed to connect to database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	log.Info("connected to database",
		"host", cfg.Database.Host,
		"database", cfg.Database.Name,
	)

	mux := http.NewServeMux()
	mux.HandleFunc("/api/ping", pingHandler)

	organizationHandler := organization.NewHandler(organization.NewRepository(db))
	organizationHandler.RegisterRoutes(mux)

	positionHandler := position.NewHandler(position.NewRepository(db))
	positionHandler.RegisterRoutes(mux)

	employeeHandler := employee.NewHandler(employee.NewRepository(db))
	employeeHandler.RegisterRoutes(mux)

	srv := &http.Server{
		Addr:    ":" + cfg.HTTP.Port,
		Handler: httpserver.WithCORS(cfg.HTTP.AllowOrigin, mux),
	}

	serverErr := make(chan error, 1)

	go func() {
		log.Info("starting http server", "addr", srv.Addr)

		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			serverErr <- err
			return
		}

		serverErr <- nil
	}()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)

	select {
	case err := <-serverErr:
		if err != nil {
			log.Error("http server error", "error", err)
			os.Exit(1)
		}
	case sig := <-stop:
		log.Info("shutdown signal received", "signal", sig.String())

		shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer shutdownCancel()

		if err := srv.Shutdown(shutdownCtx); err != nil {
			log.Error("graceful shutdown failed", "error", err)
			os.Exit(1)
		}

		log.Info("shutdown complete")
	}
}

func pingHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)

	_ = json.NewEncoder(w).Encode(map[string]string{
		"status": "ok",
	})
}

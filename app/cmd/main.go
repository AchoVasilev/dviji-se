package main

import (
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"server/cmd/db/database"
	"server/internal/infrastructure/environment"
	"server/internal/server"
	"syscall"
	"time"
)

func main() {
	environment.LoadEnvironmentVariables()
	db := database.ConnectDatabase()
	database.RunMigrations(db)
	server.Initialize(db)

	go func() {
		if err := server.Run(); err != nil && err != http.ErrServerClosed {
			slog.Error("Server failed to start", "error", err)
			os.Exit(1)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := server.Shutdown(ctx); err != nil {
		slog.Error("Server forced to shutdown", "error", err)
	}

	if err := db.Close(); err != nil {
		slog.Error("Error closing database connection", "error", err)
	}

	slog.Info("Server exited")
}

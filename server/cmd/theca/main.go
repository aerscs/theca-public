package main

import (
	"context"
	"fmt"
	"log/slog"
	"os/signal"
	"syscall"
	"time"

	"github.com/aerscs/theca-public/internal/app"
	"github.com/aerscs/theca-public/internal/config"
	"github.com/aerscs/theca-public/internal/logger"
)

// @title           Theca API
// @version         1.0
// @description     Bookmarks manager API

// @license.name  Apache 2.0
// @license.url   https://www.apache.org/licenses/LICENSE-2.0.html

// @host      localhost:8080

func main() {
	const op = "main"
	cfg := config.Load()
	log := logger.SetupLogger(cfg.LogLevel)
	logMain := log.With(slog.String("op", op))
	logMain.Debug(fmt.Sprintf("config: %+v", cfg))

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	app := app.New(ctx, cfg, log)
	app.Run()

	<-ctx.Done()
	logMain.Info("received signal to stop application")

	totalShutdownTimeout := time.Duration(cfg.ShutdownTimeout+10) * time.Second

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), totalShutdownTimeout)
	defer shutdownCancel()

	done := make(chan struct{})

	go func() {
		app.Stop()
		close(done)
	}()

	select {
	case <-done:
		logMain.Info("application stopped gracefully")
	case <-shutdownCtx.Done():
		logMain.Warn("application shutdown timeout exceeded, forcing exit")
	}
}

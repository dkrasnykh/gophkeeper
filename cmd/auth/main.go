package main

import (
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dkrasnykh/gophkeeper/internal/auth"
	"github.com/dkrasnykh/gophkeeper/internal/auth/config"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))

	cfg := config.MustLoad()
	log.Info("starting application", slog.Any("config", cfg))

	app := auth.New(log, cfg.GRPC.Port, cfg.DatabaseURL, cfg.TokenTTL, cfg.ConnectTimeout, cfg.CertFile, cfg.KeyFile)
	go app.MustRun()

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGTERM, syscall.SIGINT)

	sign := <-stop
	log.Info("stopping application", slog.String("signal", sign.String()))
	app.Stop()
	log.Info("application stopped")
}

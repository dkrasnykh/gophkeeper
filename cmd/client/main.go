package main

import (
	"context"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/dkrasnykh/gophkeeper/internal/client"
	"github.com/dkrasnykh/gophkeeper/internal/client/config"
)

func main() {
	log := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	cfg := config.MustLoad()
	log.Debug("starting client application", slog.Any("config", cfg))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	app := client.NewAppClient(log, cfg)

	stop := make(chan os.Signal, 1)

	go app.Run(ctx, stop)

	signal.Notify(stop, syscall.SIGINT)

	sign := <-stop
	log.Debug("stopping application", slog.String("signal", sign.String()))

	cancel()
	app.Stop()

	log.Debug("application stopped")
}

package grpcapp

import (
	"errors"
	"fmt"
	"log/slog"
	"net"

	"google.golang.org/grpc"

	"github.com/dkrasnykh/gophkeeper/internal/auth/tls"
	"github.com/dkrasnykh/gophkeeper/pkg/logger/sl"
)

type App struct {
	log        *slog.Logger
	gRPCServer *grpc.Server
	port       int
}

func New(log *slog.Logger, authService Auth, port int, certFile string, keyFile string) (*App, error) {
	tlsCredentials, err := tls.LoadTLSCredentials(certFile, keyFile)
	if err != nil {
		log.Error(
			"failed to load cert and key from files",
			sl.Err(err),
			slog.String("cert file path", certFile),
			slog.String("key file path", keyFile),
		)
		return nil, errors.New("get TLS credentials error")
	}

	gRPCServer := grpc.NewServer(grpc.Creds(tlsCredentials))
	Register(gRPCServer, authService)
	return &App{
		log:        log,
		gRPCServer: gRPCServer,
		port:       port,
	}, nil
}

func (a *App) MustRun() {
	if err := a.Run(); err != nil {
		panic(err)
	}
}

func (a *App) Run() error {
	const op = "grpcapp.Run"
	log := a.log.With(
		slog.String("op", op),
		slog.Int("port", a.port),
	)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", a.port))
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	log.Info("gRPC server is running", slog.String("addr", listener.Addr().String()))

	if err := a.gRPCServer.Serve(listener); err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	return nil
}

func (a *App) Stop() {
	const op = "grpcapp.Stop"

	a.log.With(slog.String("op", op)).
		Info("stopping gRPC server", slog.Int("port", a.port))

	a.gRPCServer.GracefulStop()
}

package grpcclient

import (
	"context"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"

	authv1 "github.com/dkrasnykh/gophkeeper/protos/gen/go/auth"
)

type GRPCClient struct {
	conn   *grpc.ClientConn
	client authv1.AuthClient
}

func NewGRPCClient(address string) (*GRPCClient, error) {
	conn, err := grpc.Dial(address, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, err
	}
	return &GRPCClient{
		conn:   conn,
		client: authv1.NewAuthClient(conn),
	}, nil
}

func (c *GRPCClient) Register(ctx context.Context, login string, password string) error {
	req := authv1.RegisterRequest{Email: login, Password: password}
	_, err := c.client.Register(ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			// sends error messages into UI TODO change logic
			switch e.Code() {
			case codes.AlreadyExists:
				return fmt.Errorf("user with email %s already registered", login)
			case codes.InvalidArgument:
				return fmt.Errorf("invalid login or password")
			default:
				//codes.Internal
				return fmt.Errorf("something went wrong, please try again later")
			}
		}
		return fmt.Errorf("something went wrong, please try again later")
	}
	return nil
}

func (c *GRPCClient) Login(ctx context.Context, login string, password string) (string, error) {
	req := authv1.LoginRequest{Email: login, Password: password, AppId: 1}
	resp, err := c.client.Login(ctx, &req)
	if err != nil {
		if e, ok := status.FromError(err); ok {
			// sends error messages into UI TODO change logic
			switch e.Code() {
			case codes.InvalidArgument:
				return "", fmt.Errorf("invalid login or password")
			default:
				return "", fmt.Errorf("something went wrong, please try again later")
			}
		}
		return "", fmt.Errorf("something went wrong, please try again later")
	}
	return resp.Token, nil
}

func (c *GRPCClient) Stop() {
	_ = c.conn.Close()
}

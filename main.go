package main

import (
	"net"
	"time"

	"github.com/mikevel2955/gandalf/pb"
	"github.com/mikevel2955/gandalf/utils" // @FixMe use github.com/mikevel2955/hermes-utils
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type appConfig struct {
	Addr string `env:"GRPC_ADDR" def:":40001"`
}

const (
	AppName = "gandalf"
)

func main() {
	sugaredLogger := zap.NewExample().Sugar() // @FixMe use zap.New() to construct production-like logger

	config := appConfig{}
	if err := utils.ReadConfig(&config); err != nil {
		sugaredLogger.Fatal(err)
	}

	opts := []grpc.ServerOption{
		grpc.ConnectionTimeout(5 * time.Second),
	}
	grpcServer := grpc.NewServer(opts...)

	server := NewService(
		sugaredLogger,
	)
	gandalf.RegisterGandalfServer(grpcServer, server)

	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		sugaredLogger.Fatalf("failed to listen: %v", err)
	}

	sugaredLogger.Info("gandalf started")
	if err := grpcServer.Serve(listener); err != nil {
		sugaredLogger.Fatalf("gandalf stopped with error: %v", err)
	}
	sugaredLogger.Info("gandalf stopped")
}

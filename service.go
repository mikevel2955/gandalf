package main

import (
	"context"
	"fmt"

	pb "github.com/mikevel2955/gandalf/pb"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.SugaredLogger
}

func NewService(
	logger *zap.SugaredLogger,
) *Service {
	return &Service{
		logger: logger,
	}
}

func (s *Service) Do(_ context.Context, req *pb.Request) (*pb.Response, error) {
	s.logger.Infof("got request: %v", *req)
	return &pb.Response{
		Message: fmt.Sprintf("got '%s'", req.Message),
	}, nil
}

package main

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/mikevel2955/gandalf/pb"
	"go.uber.org/zap"
)

type Server struct {
	logger        *zap.SugaredLogger
	userOperators []int64
	userViewers   []int64

	// Mock fields
	symbols []tradingSymbol
}

type tradingSymbol struct {
	symbol string
	status pb.TradingSymbol_TradingStatus
}

var (
	errUserNotOperator = errors.New("you are not authorized to perform this operation")
	errUserNotViewer   = errors.New("you are not authorized to view this data")
)

func NewServer(
	logger *zap.SugaredLogger,
	userOperators []int64,
	userViewers []int64,
) *Server {
	return &Server{
		logger:        logger,
		userOperators: userOperators,
		userViewers:   userViewers,

		symbols: []tradingSymbol{
			{"adausdt", pb.TradingSymbol_ACTIVE},
			{"linkusdt", pb.TradingSymbol_ACTIVE},
			{"zilusdt", pb.TradingSymbol_ACTIVE},
			{"ltcusdt", pb.TradingSymbol_ACTIVE},
		},
	}
}

func (s *Server) GetTradingSymbols(_ context.Context, req *pb.EmptyRequest) (*pb.TradingSymbolsResponse, error) {
	if err := s.checkUserViewer(req.UserId); err != nil {
		return nil, err
	}

	symbols := make([]*pb.TradingSymbol, 0, len(s.symbols))
	for _, symbol := range s.symbols {
		symbols = append(symbols, &pb.TradingSymbol{
			Symbol: symbol.symbol,
			Status: symbol.status,
		})
	}

	return &pb.TradingSymbolsResponse{
		Symbols: symbols,
	}, nil
}

func (s *Server) SymbolTradingPrepare(_ context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	for _, symbol := range s.symbols {
		if symbol.symbol == req.Symbol {
			return nil, errors.New(fmt.Sprintf("%s is already in trading", req.Symbol))
		}
	}
	s.symbols = append(s.symbols, tradingSymbol{
		symbol: req.Symbol,
		status: pb.TradingSymbol_PREPARING,
	})
	return &pb.EmptyResponse{}, nil
}

func (s *Server) SymbolTradingStart(_ context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	panic("implement me")
}

func (s *Server) SymbolTradingStop(_ context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	panic("implement me")
}

func (s *Server) SymbolTradingSuspend(_ context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	panic("implement me")
}

func (s *Server) SymbolTradingResume(_ context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	panic("implement me")
}

func (s *Server) GetSymbolBalances(_ context.Context, req *pb.EmptyRequest) (*pb.SymbolBalancesResponse, error) {
	if err := s.checkUserViewer(req.UserId); err != nil {
		return nil, err
	}

	panic("implement me")
}

func (s *Server) GetSymbolLimits(_ context.Context, req *pb.GetSymbolLimitsRequest) (*pb.SymbolLimitsResponse, error) {
	if err := s.checkUserViewer(req.UserId); err != nil {
		return nil, err
	}

	panic("implement me")
}

func (s *Server) SetSymbolLimits(_ context.Context, req *pb.SetSymbolLimitsRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	panic("implement me")
}

func (s *Server) GetActiveDeals(_ context.Context, req *pb.DealsRequest) (*pb.DealsResponse, error) {
	if err := s.checkUserViewer(req.UserId); err != nil {
		return nil, err
	}

	panic("implement me")
}

func (s *Server) GetPotentialDeals(_ context.Context, req *pb.DealsRequest) (*pb.PotentialDealsResponse, error) {
	if err := s.checkUserViewer(req.UserId); err != nil {
		return nil, err
	}

	panic("implement me")
}

func (s *Server) CloseDeals(_ context.Context, req *pb.DealsRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	panic("implement me")
}

func (s *Server) checkUserOperator(userId int64) error {
	if !int64InList(userId, s.userOperators) {
		return errUserNotOperator
	}
	return nil
}

func (s *Server) checkUserViewer(userId int64) error {
	if !int64InList(userId, s.userOperators) {
		return errUserNotViewer
	}
	return nil
}

func int64InList(n int64, list []int64) bool {
	for _, i := range list {
		if i == n {
			return true
		}
	}
	return false
}

package main

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/mikevel2955/gandalf/pb"
	"go.uber.org/zap"
)

type Service struct {
	logger *zap.SugaredLogger

	// Mock fields
	symbols []tradingSymbol
}

type tradingSymbol struct {
	symbol string
	status pb.TradingSymbol_TradingStatus
}

func NewService(
	logger *zap.SugaredLogger,
) *Service {
	return &Service{
		logger: logger,
		symbols: []tradingSymbol{
			{"adausdt", pb.TradingSymbol_ACTIVE},
			{"linkusdt", pb.TradingSymbol_ACTIVE},
			{"zilusdt", pb.TradingSymbol_ACTIVE},
			{"ltcusdt", pb.TradingSymbol_ACTIVE},
		},
	}
}

func (s *Service) GetTradingSymbols(context.Context, *pb.EmptyRequest) (*pb.TradingSymbolsResponse, error) {
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

func (s *Service) SymbolTradingPrepare(_ context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
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

func (s *Service) SymbolTradingStart(context.Context, *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	panic("implement me")
}

func (s *Service) SymbolTradingStop(context.Context, *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	panic("implement me")
}

func (s *Service) SymbolTradingSuspend(context.Context, *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	panic("implement me")
}

func (s *Service) SymbolTradingResume(context.Context, *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	panic("implement me")
}

func (s *Service) GetSymbolBalances(context.Context, *pb.EmptyRequest) (*pb.SymbolBalancesResponse, error) {
	panic("implement me")
}

func (s *Service) GetSymbolLimits(context.Context, *pb.GetSymbolLimitsRequest) (*pb.SymbolLimitsResponse, error) {
	panic("implement me")
}

func (s *Service) SetSymbolLimits(context.Context, *pb.SetSymbolLimitsRequest) (*pb.EmptyResponse, error) {
	panic("implement me")
}

func (s *Service) GetActiveDeals(context.Context, *pb.DealsRequest) (*pb.DealsResponse, error) {
	panic("implement me")
}

func (s *Service) GetPotentialDeals(context.Context, *pb.DealsRequest) (*pb.PotentialDealsResponse, error) {
	panic("implement me")
}

func (s *Service) CloseDeals(context.Context, *pb.DealsRequest) (*pb.EmptyResponse, error) {
	panic("implement me")
}

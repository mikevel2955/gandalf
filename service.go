package main

import (
	"context"
	"errors"
	"fmt"

	pb "github.com/mikevel2955/gandalf/pb"
	"go.uber.org/zap"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Server struct {
	logger        *zap.SugaredLogger
	userOperators []int64
	userViewers   []int64

	// Mock fields
	symbols []tradingSymbol
	deals   []*pb.Deal
}

type tradingSymbol struct {
	symbol  string
	status  pb.TradingSymbol_TradingStatus
	balance float32
	limit   float32
}

var (
	errUserNotOperator = errors.New("you are not authorized to perform this operation")
	errUserNotViewer   = errors.New("you are not authorized to view this data")
	errSymbolNotFound  = func(symbol string) error {
		return errors.New(fmt.Sprintf("unknown symbol '%s'", symbol))
	}
	errDealNotFound = func(dealId string) error {
		return errors.New(fmt.Sprintf("unknown deal '%s'", dealId))
	}
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
			{"adausdt", pb.TradingSymbol_ACTIVE, 55, 100},
			{"linkusdt", pb.TradingSymbol_ACTIVE, 66, 100},
			{"zilusdt", pb.TradingSymbol_ACTIVE, 33, 100},
			{"ltcusdt", pb.TradingSymbol_ACTIVE, 22, 100},
		},
		deals: []*pb.Deal{{
			DealId:         "adausdt-1657483456",
			Symbol:         "adausdt",
			CreatedAt:      timestamppb.Now(),
			Amount:         0.01,
			AmountCurrency: 361,
			DeltaAmount:    -12,
			DeltaPercent:   -2,
			Prediction: &pb.Deal_DealPrediction{
				Stop: -3,
				Max:  2,
			},
		}, {
			DealId:         "adausdt-1630958723",
			Symbol:         "adausdt",
			CreatedAt:      timestamppb.Now(),
			Amount:         0.04,
			AmountCurrency: 734,
			DeltaAmount:    15,
			DeltaPercent:   2,
			Prediction: &pb.Deal_DealPrediction{
				Stop: -5,
				Max:  7,
			},
		}, {
			DealId:         "linkusdt-1630958723",
			Symbol:         "linkusdt",
			CreatedAt:      timestamppb.Now(),
			Amount:         0.5,
			AmountCurrency: 154,
			DeltaAmount:    7,
			DeltaPercent:   5,
			Prediction: &pb.Deal_DealPrediction{
				Stop: -15,
				Max:  7,
			},
		}},
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

	return s.setSymbolStatus(req.Symbol, pb.TradingSymbol_ACTIVE)
}

func (s *Server) SymbolTradingStop(_ context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	n, err := s.findSymbolIndex(req.Symbol)
	if err != nil {
		return nil, err
	}

	s.symbols = append(s.symbols[:n], s.symbols[n+1:]...)

	return &pb.EmptyResponse{}, nil
}

func (s *Server) SymbolTradingSuspend(_ context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	return s.setSymbolStatus(req.Symbol, pb.TradingSymbol_SUSPENDED)
}

func (s *Server) SymbolTradingResume(_ context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	return s.setSymbolStatus(req.Symbol, pb.TradingSymbol_ACTIVE)
}

func (s *Server) GetSymbolBalances(_ context.Context, req *pb.EmptyRequest) (*pb.SymbolBalancesResponse, error) {
	if err := s.checkUserViewer(req.UserId); err != nil {
		return nil, err
	}

	var balances []*pb.SymbolBalance
	var total float32
	for _, symbol := range s.symbols {
		balances = append(balances, &pb.SymbolBalance{
			Symbol: symbol.symbol,
			Amount: symbol.balance,
		})
		total += symbol.balance
	}

	balances = append(balances, &pb.SymbolBalance{
		Symbol: "usd",
		Amount: total,
	})

	return &pb.SymbolBalancesResponse{
		Balances: balances,
	}, nil
}

func (s *Server) GetSymbolLimits(_ context.Context, req *pb.GetSymbolLimitsRequest) (*pb.SymbolLimitsResponse, error) {
	if err := s.checkUserViewer(req.UserId); err != nil {
		return nil, err
	}

	limits := make([]*pb.SymbolLimit, 0, len(s.symbols))
	for _, symbol := range s.symbols {
		limits = append(limits, &pb.SymbolLimit{
			Symbol: symbol.symbol,
			Limit:  symbol.limit,
		})
	}

	return &pb.SymbolLimitsResponse{
		Limits: limits,
	}, nil
}

func (s *Server) SetSymbolLimits(_ context.Context, req *pb.SetSymbolLimitsRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	for _, limit := range req.Limits {
		n, err := s.findSymbolIndex(limit.Symbol)
		if err != nil {
			return nil, err
		}

		s.symbols[n].limit = limit.Limit
	}

	return &pb.EmptyResponse{}, nil
}

func (s *Server) GetActiveDeals(_ context.Context, req *pb.DealsRequest) (*pb.DealsResponse, error) {
	if err := s.checkUserViewer(req.UserId); err != nil {
		return nil, err
	}

	if req.All {
		return &pb.DealsResponse{
			Deals: s.deals,
		}, nil
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

	if req.All {
		s.deals = []*pb.Deal{}
		return &pb.EmptyResponse{}, nil
	}

	for _, dealId := range req.DealIds {
		n, err := s.findDealIndex(dealId)
		if err != nil {
			return nil, err
		}
		s.deals = append(s.deals[:n], s.deals[n+1:]...)
	}

	return &pb.EmptyResponse{}, nil
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

func (s *Server) setSymbolStatus(symbol string, status pb.TradingSymbol_TradingStatus) (*pb.EmptyResponse, error) {
	n, err := s.findSymbolIndex(symbol)
	if err != nil {
		return nil, err
	}

	s.symbols[n].status = status

	return &pb.EmptyResponse{}, nil
}

func (s *Server) findSymbolIndex(symbol string) (int, error) {
	for i := range s.symbols {
		if s.symbols[i].symbol == symbol {
			return i, nil
		}
	}
	return -1, errSymbolNotFound(symbol)
}

func (s *Server) findDealIndex(dealId string) (int, error) {
	for n := range s.deals {
		if s.deals[n].DealId == dealId {
			return n, nil
		}
	}
	return -1, errDealNotFound(dealId)
}

func int64InList(n int64, list []int64) bool {
	for _, i := range list {
		if i == n {
			return true
		}
	}
	return false
}

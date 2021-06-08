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
	storage       *Storage
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
	storage *Storage,
) *Server {
	return &Server{
		logger:        logger,
		userOperators: userOperators,
		userViewers:   userViewers,
		storage:       storage,
	}
}

func (s *Server) GetTradingSymbols(ctx context.Context, req *pb.EmptyRequest) (*pb.TradingSymbolsResponse, error) {
	if err := s.checkUserViewer(req.UserId); err != nil {
		return nil, err
	}

	tradingSymbols, err := s.storage.GetTradingSymbols(ctx)
	if err != nil {
		return nil, err
	}

	var symbols []*pb.TradingSymbol
	for _, symbol := range tradingSymbols {
		symbols = append(symbols, &pb.TradingSymbol{
			Symbol: symbol.Symbol,
			Status: symbol.Status,
		})
	}

	return &pb.TradingSymbolsResponse{
		Symbols: symbols,
	}, nil
}

func (s *Server) SymbolTradingPrepare(ctx context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	tradingSymbol, err := s.storage.GetTradingSymbol(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}
	if tradingSymbol != nil {
		return nil, errors.New(fmt.Sprintf("%s is already in trading", req.Symbol))
	}

	tradingSymbol = &TradingSymbol{req.Symbol, pb.TradingSymbol_PREPARING, 0, 100}
	if err := s.storage.SaveTradingSymbol(ctx, tradingSymbol); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

func (s *Server) SymbolTradingStart(ctx context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	return s.setSymbolStatus(ctx, req.Symbol, pb.TradingSymbol_ACTIVE)
}

func (s *Server) SymbolTradingStop(ctx context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	err := s.storage.DeleteTradingSymbol(ctx, req.Symbol)
	if err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

func (s *Server) SymbolTradingSuspend(ctx context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	return s.setSymbolStatus(ctx, req.Symbol, pb.TradingSymbol_SUSPENDED)
}

func (s *Server) SymbolTradingResume(ctx context.Context, req *pb.SymbolRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	return s.setSymbolStatus(ctx, req.Symbol, pb.TradingSymbol_ACTIVE)
}

func (s *Server) GetSymbolBalances(ctx context.Context, req *pb.EmptyRequest) (*pb.SymbolBalancesResponse, error) {
	if err := s.checkUserViewer(req.UserId); err != nil {
		return nil, err
	}

	tradingSymbols, err := s.storage.GetTradingSymbols(ctx)
	if err != nil {
		return nil, err
	}

	var balances []*pb.SymbolBalance
	var total float32

	for _, symbol := range tradingSymbols {
		balances = append(balances, &pb.SymbolBalance{
			Symbol: symbol.Symbol,
			Amount: symbol.Balance,
		})
		total += symbol.Balance
	}

	balances = append(balances, &pb.SymbolBalance{
		Symbol: "usd",
		Amount: total,
	})

	return &pb.SymbolBalancesResponse{
		Balances: balances,
	}, nil
}

func (s *Server) GetSymbolLimits(ctx context.Context, req *pb.GetSymbolLimitsRequest) (*pb.SymbolLimitsResponse, error) {
	if err := s.checkUserViewer(req.UserId); err != nil {
		return nil, err
	}

	tradingSymbols, err := s.storage.GetTradingSymbols(ctx)
	if err != nil {
		return nil, err
	}

	var limits []*pb.SymbolLimit

	for _, symbol := range tradingSymbols {
		limits = append(limits, &pb.SymbolLimit{
			Symbol: symbol.Symbol,
			Limit:  symbol.Limit,
		})
	}

	return &pb.SymbolLimitsResponse{
		Limits: limits,
	}, nil
}

func (s *Server) SetSymbolLimits(ctx context.Context, req *pb.SetSymbolLimitsRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	for _, limit := range req.Limits {
		tradingSymbol, err := s.storage.GetTradingSymbol(ctx, limit.Symbol)
		if err != nil {
			return nil, err
		}
		if tradingSymbol == nil {
			return nil, errSymbolNotFound(limit.Symbol)
		}

		tradingSymbol.Limit = limit.Limit
		if err := s.storage.SaveTradingSymbol(ctx, tradingSymbol); err != nil {
			return nil, err
		}
	}

	return &pb.EmptyResponse{}, nil
}

func (s *Server) GetActiveDeals(ctx context.Context, req *pb.DealsRequest) (*pb.DealsResponse, error) {
	if err := s.checkUserViewer(req.UserId); err != nil {
		return nil, err
	}

	if req.All {
		deals, err := s.storage.GetDeals(ctx)
		if err != nil {
			return nil, err
		}

		var activeDeals []*pb.Deal
		for _, deal := range deals {
			activeDeals = append(activeDeals, &pb.Deal{
				DealId:         deal.Id,
				Symbol:         deal.Symbol,
				CreatedAt:      timestamppb.New(deal.CreatedAt),
				Amount:         deal.Amount,
				AmountCurrency: deal.AmountCurrency,
				DeltaAmount:    deal.DeltaAmount,
				DeltaPercent:   deal.DeltaPercent,
				Prediction: &pb.Deal_DealPrediction{
					Stop: deal.Prediction.Stop,
					Max:  deal.Prediction.Max,
				},
			})
		}

		return &pb.DealsResponse{
			Deals: activeDeals,
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

func (s *Server) CloseDeals(ctx context.Context, req *pb.DealsRequest) (*pb.EmptyResponse, error) {
	if err := s.checkUserOperator(req.UserId); err != nil {
		return nil, err
	}

	if req.All {
		deals, err := s.storage.GetDeals(ctx)
		if err != nil {
			return nil, err
		}
		for _, deal := range deals {
			if err := s.storage.DeleteDeal(ctx, deal.Id); err != nil {
				return nil, err
			}
		}
		return &pb.EmptyResponse{}, nil
	}

	for _, dealId := range req.DealIds {
		deal, err := s.storage.GetDeal(ctx, dealId)
		if err != nil {
			return nil, err
		}

		if deal == nil {
			return nil, errDealNotFound(dealId)
		}

		if err := s.storage.DeleteDeal(ctx, deal.Id); err != nil {
			return nil, err
		}
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

func (s *Server) setSymbolStatus(
	ctx context.Context,
	symbol string,
	status pb.TradingSymbol_TradingStatus,
) (*pb.EmptyResponse, error) {
	tradingSymbol, err := s.storage.GetTradingSymbol(ctx, symbol)
	if err != nil {
		return nil, err
	}

	if tradingSymbol == nil {
		return nil, errSymbolNotFound(symbol)
	}

	tradingSymbol.Status = status
	if err := s.storage.SaveTradingSymbol(ctx, tradingSymbol); err != nil {
		return nil, err
	}

	return &pb.EmptyResponse{}, nil
}

func int64InList(n int64, list []int64) bool {
	for _, i := range list {
		if i == n {
			return true
		}
	}
	return false
}

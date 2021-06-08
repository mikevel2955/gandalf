package main

import (
	"context"
	"fmt"
	"time"

	pb "github.com/mikevel2955/gandalf/pb"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

type Storage struct {
	client *mongo.Client
	dbName string
}

type TradingSymbol struct {
	Symbol  string                         `bson:"_id"`
	Status  pb.TradingSymbol_TradingStatus `bson:"status"`
	Balance float32                        `bson:"balance"`
	Limit   float32                        `bson:"limit"`
}

type Deal struct {
	Id             string         `bson:"_id"`
	Symbol         string         `bson:"symbol"`
	CreatedAt      time.Time      `bson:"created_at"`
	Amount         float32        `bson:"amount"`
	AmountCurrency float32        `bson:"amount_currency"`
	DeltaAmount    float32        `bson:"delta_amount"`
	DeltaPercent   float32        `bson:"delta_percent"`
	Prediction     DealPrediction `bson:"prediction"`
}

type DealPrediction struct {
	Stop float32 `bson:"stop"`
	Max  float32 `bson:"max"`
}

const (
	symbolsCollection = "symbols"
	dealsCollection   = "deals"
)

func NewStorage(
	client *mongo.Client,
	dbName string,
) *Storage {
	return &Storage{
		client: client,
		dbName: dbName,
	}
}

func (s *Storage) SaveTradingSymbol(ctx context.Context, tradingSymbol *TradingSymbol) error {
	_, err := s.getSymbolsCollection().ReplaceOne(
		ctx,
		bson.M{"_id": tradingSymbol.Symbol},
		tradingSymbol,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (s *Storage) GetTradingSymbols(ctx context.Context) ([]*TradingSymbol, error) {
	cursor, err := s.getSymbolsCollection().Find(ctx, bson.M{})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	symbols := make([]*TradingSymbol, 0)
	if err := cursor.All(ctx, &symbols); err != nil {
		return nil, err
	}

	return symbols, nil
}

func (s *Storage) GetTradingSymbol(ctx context.Context, symbol string) (*TradingSymbol, error) {
	document := s.getSymbolsCollection().FindOne(ctx, bson.M{"_id": symbol})
	if document.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if document.Err() != nil {
		return nil, document.Err()
	}

	tradingSymbol := &TradingSymbol{}
	if err := document.Decode(tradingSymbol); err != nil {
		return nil, err
	}

	return tradingSymbol, nil
}

func (s *Storage) DeleteTradingSymbol(ctx context.Context, symbol string) error {
	_, err := s.getSymbolsCollection().DeleteOne(ctx, bson.M{"_id": symbol})
	return err
}

func (s *Storage) SaveDeal(ctx context.Context, deal *Deal) error {
	_, err := s.getDealsCollection().ReplaceOne(
		ctx,
		bson.M{"_id": deal.Id},
		deal,
		options.Replace().SetUpsert(true),
	)
	return err
}

func (s *Storage) GetDeals(ctx context.Context) ([]*Deal, error) {
	cursor, err := s.getDealsCollection().Find(ctx, bson.M{})
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	deals := make([]*Deal, 0)
	if err := cursor.All(ctx, &deals); err != nil {
		return nil, err
	}

	return deals, nil
}

func (s *Storage) GetDeal(ctx context.Context, dealId string) (*Deal, error) {
	document := s.getDealsCollection().FindOne(ctx, bson.M{"_id": dealId})
	if document.Err() == mongo.ErrNoDocuments {
		return nil, nil
	} else if document.Err() != nil {
		return nil, document.Err()
	}

	deal := &Deal{}
	if err := document.Decode(deal); err != nil {
		return nil, err
	}

	return deal, nil
}

func (s *Storage) DeleteDeal(ctx context.Context, dealId string) error {
	_, err := s.getDealsCollection().DeleteOne(ctx, bson.M{"_id": dealId})
	return err
}

func (s *Storage) getSymbolsCollection() *mongo.Collection {
	return s.client.Database(s.dbName).Collection(symbolsCollection)
}

func (s *Storage) getDealsCollection() *mongo.Collection {
	return s.client.Database(s.dbName).Collection(dealsCollection)
}

func (s *Storage) Init() error {
	ctx := context.Background()

	_ = s.getSymbolsCollection().Drop(ctx)
	_ = s.SaveTradingSymbol(ctx, &TradingSymbol{"adausdt", pb.TradingSymbol_ACTIVE, 55, 100})
	_ = s.SaveTradingSymbol(ctx, &TradingSymbol{"linkusdt", pb.TradingSymbol_ACTIVE, 66, 100})
	_ = s.SaveTradingSymbol(ctx, &TradingSymbol{"zilusdt", pb.TradingSymbol_ACTIVE, 33, 100})
	_ = s.SaveTradingSymbol(ctx, &TradingSymbol{"ltcusdt", pb.TradingSymbol_ACTIVE, 22, 100})

	_ = s.getDealsCollection().Drop(ctx)
	_ = s.SaveDeal(ctx, &Deal{"adausdt-1657483456", "adausdt", time.Now(), 0.01, 361, -12, -2, DealPrediction{-3, 2}})
	_ = s.SaveDeal(ctx, &Deal{"adausdt-1630958723", "adausdt", time.Now(), 0.04, 734, 15, 2, DealPrediction{-5, 7}})
	_ = s.SaveDeal(ctx, &Deal{"linkusdt-3492445345", "linkusdt", time.Now(), 0.05, 154, 7, 5, DealPrediction{-15, 3}})

	return nil
}

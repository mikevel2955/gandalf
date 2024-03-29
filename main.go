package main

import (
	"context"
	"net"
	"strconv"
	"strings"
	"time"

	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"

	gandalfPb "github.com/mikevel2955/gandalf/pb"
	utils "github.com/mikevel2955/hermes-utils"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

type appConfig struct {
	Addr          string `env:"GRPC_ADDR" def:":40001"`
	UserOperators string `env:"USER_OPERATORS_LIST"`
	UserViewers   string `env:"USER_VIEWERS_LIST"`
	MongoDSN      string `env:"MONGO_DSN" def:"mongodb://127.0.0.1:27017"`
	MongoDBName   string `env:"MONGO_DB_NAME" def:"gandalf_db"`
	MongoInitDB   string `env:"MONGO_INIT_DB" def:"false"`
}

const (
	AppName = "gandalf"
)

func main() {
	logger := zap.NewExample().Sugar() // @FixMe use zap.New() to construct production-like logger

	config := appConfig{}
	if err := utils.ReadConfig(&config); err != nil {
		logger.Fatal(err)
	}

	logger.Infof("connecting to %v", config.MongoDSN)
	mongoClient, err := mongo.NewClient(options.Client().ApplyURI(config.MongoDSN))
	if err != nil {
		logger.Panic("cannot instantiate mongo client")
	}
	if err := mongoClient.Connect(context.Background()); err != nil {
		logger.Panicf("cannot connect client: %v", err)
	}

	storage := NewStorage(mongoClient, config.MongoDBName)
	if initDb, _ := strconv.ParseBool(config.MongoInitDB); initDb {
		if err := storage.Init(); err != nil {
			logger.Panicf("cannot init db: %v", err)
		}
	}

	server := NewServer(
		logger,
		parseInts(logger, "USER_OPERATORS_LIST env", config.UserOperators),
		parseInts(logger, "USER_VIEWERS_LIST env", config.UserViewers),
		storage,
	)

	grpcServer := grpc.NewServer(grpc.ConnectionTimeout(5 * time.Second))
	gandalfPb.RegisterGandalfServer(grpcServer, server)

	listener, err := net.Listen("tcp", config.Addr)
	if err != nil {
		logger.Fatalf("failed to listen: %v", err)
	}

	logger.Info("gandalf started")
	if err := grpcServer.Serve(listener); err != nil {
		logger.Fatalf("gandalf stopped with error: %v", err)
	}
	logger.Info("gandalf stopped")
}

// TODO move it to hermes-utils
func parseInts(logger *zap.SugaredLogger, srcName, src string) []int64 {
	ss := strings.Split(src, ",")
	ints := make([]int64, 0, len(ss))

	for _, s := range ss {
		s = strings.TrimSpace(s)
		if s == "" {
			continue
		}

		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			logger.Fatalf("can't parse %s: %v", srcName, err)
		}

		ints = append(ints, n)
	}

	return ints
}

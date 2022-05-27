package main

import (
	"os"
	"os/signal"
	"syscall"

	"github.com/ernur-eskermes/product-store/internal/transport/grpc"
	grpcHandler "github.com/ernur-eskermes/product-store/internal/transport/grpc/handlers"

	"github.com/ernur-eskermes/product-store/pkg/database/mongodb"

	"github.com/ernur-eskermes/product-store/internal/config"
	"github.com/ernur-eskermes/product-store/internal/service"
	"github.com/ernur-eskermes/product-store/internal/storage"
	"github.com/ernur-eskermes/product-store/pkg/logging"
)

func main() {
	logger := logging.GetLogger()

	cfg, err := config.New()
	if err != nil {
		logger.Fatal(err)
	}

	mongoClient, err := mongodb.NewClient(cfg.Mongo.URI, cfg.Mongo.User, cfg.Mongo.Password)
	if err != nil {
		logger.Fatal(err)
	}

	db := mongoClient.Database(cfg.Mongo.Database)

	storages := storage.New(db)
	services := service.New(service.Deps{
		ProductStorage: storages.Product,
	})

	grpcHandlers := grpcHandler.New(grpcHandler.Deps{
		ProductService: services.Product,
	})
	grpcSrv := grpc.New(grpcHandlers)

	go func() {
		logger.Info("Starting gRPC server")

		if err = grpcSrv.ListenAndServe(cfg.GRPC.Port); err != nil {
			logger.Error("gRPC ListenAndServer error", err)
		}
	}()

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	<-quit

	logger.Info("Shutting down server")

	grpcSrv.Stop()
}

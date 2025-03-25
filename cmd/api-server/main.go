package main

import (
	"context"
	"fmt"
	"github.com/dylan-dinh/esl-test/internal/config"
	"github.com/dylan-dinh/esl-test/internal/domain/user"
	"github.com/dylan-dinh/esl-test/internal/infrastructure/persistence/db"
	"github.com/dylan-dinh/esl-test/internal/infrastructure/persistence/repository"
	pb "github.com/dylan-dinh/esl-test/internal/interfaces/grpc/user"
	"github.com/dylan-dinh/esl-test/internal/interfaces/notifier"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"log"
	"log/slog"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	conf, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	rabbitConn, err := notifier.NewRabbitMQConn(conf)
	if err != nil {
		panic(err)
	}
	defer func() {
		if err := rabbitConn.Close(); err != nil {
			logger.Error("error disconnecting from rabbit", "error", err)
		}
	}()

	mq, err := user.NewRabbitMQ(rabbitConn)
	if err != nil {
		panic(err)
	}

	newDb, err := db.NewDb(conf)
	if err != nil {
		panic(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	defer func() {
		if err := newDb.DB.Disconnect(ctx); err != nil {
			logger.Error("error disconnecting from DB", "error", err)
		}
	}()

	userRepo, err := repository.NewUserRepository(newDb.DB, conf.DbName)
	if err != nil {
		panic(err)
	}
	userService := user.NewUserService(userRepo, mq)
	userServer := pb.NewUserServer(userService)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%s", conf.GrpcPort))
	if err != nil {
		panic("failed to listen")
	}

	grpcServer := grpc.NewServer()
	// health check endpoint
	healthServer := health.NewServer()
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	reflection.Register(grpcServer)

	pb.RegisterUserServiceServer(grpcServer, userServer)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	go func() {
		log.Printf("gRPC server is listening on port %s...", conf.GrpcPort)
		if err := grpcServer.Serve(listener); err != nil {
			log.Fatalf("failed to serve: %v", err)
		}
	}()

	<-quit
	logger.Info("Received shutdown signal, gracefully shutting down...")
	grpcServer.GracefulStop()
}

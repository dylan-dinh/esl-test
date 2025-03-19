package main

import (
	"fmt"
	"github.com/dylan-dinh/esl-test/internal/config"
	"github.com/dylan-dinh/esl-test/internal/domain/user"
	"github.com/dylan-dinh/esl-test/internal/infrastructure/persistence/db"
	"github.com/dylan-dinh/esl-test/internal/infrastructure/persistence/repository"
	pb "github.com/dylan-dinh/esl-test/internal/interfaces/grpc/user"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	healthpb "google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"
	"log"
	"net"
)

func main() {
	conf, err := config.GetConfig()
	if err != nil {
		panic(err)
	}

	newDb, err := db.NewDb(conf)
	if err != nil {
		panic(err)
	}

	userRepo := repository.NewUserRepository(newDb.DB)
	userService := user.NewUserService(userRepo)
	userServer := pb.NewUserServer(userService)

	listener, err := net.Listen("tcp", fmt.Sprintf("localhost:%s", conf.GrpcPort))
	if err != nil {
		panic("failed to listen")
	}

	grpcServer := grpc.NewServer()

	healthServer := health.NewServer()
	healthServer.SetServingStatus("", healthpb.HealthCheckResponse_SERVING)
	healthpb.RegisterHealthServer(grpcServer, healthServer)
	reflection.Register(grpcServer)

	pb.RegisterUserServiceServer(grpcServer, userServer)

	log.Printf("gRPC server is listening on port %s...", conf.GrpcPort)
	if err := grpcServer.Serve(listener); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}

}

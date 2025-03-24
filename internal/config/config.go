package config

import (
	"errors"
	"fmt"
	"github.com/joho/godotenv"
	"os"
)

const (
	keyGrpcPort = "GRPC_PORT"
	keyDbHost   = "DB_HOST"
	keyDbPort   = "DB_PORT"
	keyDbName   = "DB_NAME"
)

type Config struct {
	GrpcPort string
	DbHost   string
	DbPort   string
	DbName   string
}

func GetConfig() (Config, error) {
	// I did this because in docker-compose for test I couldn't load env file
	// So I passed environment value
	// But testing locally I provide env file with my IDE
	if err := godotenv.Load(); err != nil && !os.IsNotExist(err) {
		return Config{}, fmt.Errorf("loading .env: %w", err)
	}

	var grpcPort, dbHost, dbPort, dbName string
	if grpcPort = os.Getenv(keyGrpcPort); grpcPort == "" {
		return Config{}, errors.New(fmt.Sprintf("env var %s not set", keyGrpcPort))
	}

	if dbHost = os.Getenv(keyDbHost); dbHost == "" {
		return Config{}, errors.New(fmt.Sprintf("env var %s not set", keyDbHost))
	}

	if dbPort = os.Getenv(keyDbPort); dbPort == "" {
		return Config{}, errors.New(fmt.Sprintf("env var %s not set", keyDbPort))
	}

	if dbName = os.Getenv(keyDbName); dbName == "" {
		return Config{}, errors.New(fmt.Sprintf("env var %s not set", keyDbName))
	}

	return Config{
		GrpcPort: grpcPort,
		DbHost:   dbHost,
		DbPort:   dbPort,
		DbName:   dbName,
	}, nil
}

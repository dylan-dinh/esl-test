# ESL‑Test User Service

A lightweight Go microservice exposing a gRPC API to manage Users (create, read, update, delete, and list with pagination + filters). 
Built with Go, MongoDB and  Docker Compose.

---

## Tech Stack

- **Language:** Go 1.23
- **API:** gRPC
- **Database:** MongoDB
- **Containerization:** Docker & Docker Compose
- **Testing:** Go unit + integration (`-tags=integration`)

---

## Prerequisites

- Docker & Docker Compose installed
- Ports **27017** (MongoDB) and **50051** (gRPC) free

---

## Configuration

Create a `.env` file at project root (or use the provided template):

```env
GRPC_PORT=50051
DB_HOST=mongodb
DB_PORT=27017
DB_NAME=testdb
```
This is needed if you want to run the app locally.

---

## Running the service
From the project root

```cmd 
docker-compose up --build
```

gRPC server: listens on localhost:50051
MongoDB: accessible at localhost:27017

## HealthCheck

```ht
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
```

## Testing
Run all tests (unit + integration) in isolation:
```test
docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
```
Integration tests use an in‑memory MongoDB
Database is wiped after each run

## Project structure
I implemented a handler/service/repository pattern :

```
.
├── cmd/api-server/main.go         # gRPC server entrypoint
├── proto/user.proto               # Protobuf definitions
├── internal/
│   ├── domain/user                # Entity + service interface
│   ├── infrastructure/persistence # db and user repository
│   └── interfaces/grpc/user       # gRPC handlers
├── Dockerfile-app                 # Production build
├── docker-compose.yml             # App + MongoDB
└── docker-compose.test.yml        # Tests + ephemeral MongoDB
```

## Logging and error
- Using Go's slog package
- Listening for signal leading to mongo being disconnected and graceful shutdown of gRPC server

## Next step
- Add TLS to gRPC in production
- Expose Prometheus metrics endpoint
- Add new entity like Games for example
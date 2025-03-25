# ESL‑Test User Service

A lightweight Go microservice exposing a gRPC API to manage Users (create, read, update, delete, and list with pagination + filters). 
Built with Go, MongoDB and  Docker Compose.

---

## Tech Stack

- **Language:** Go 1.23
- **API:** gRPC
- **Database:** MongoDB
- **Containerization:** Docker & Docker Compose
- **Message broker:** RabbitMQ
- **Testing:** Go unit + integration (`-tags=integration`)

---

## Prerequisites

- Docker & Docker Compose installed
- Ports **27017** (MongoDB) / **50051** (gRPC) free / **5672** (RabbitMQ)

---

## Configuration


**This is ONLY needed if you want to run the app locally without docker-compose**

Create a `.env` file at project root (or use the provided template):

```env
GRPC_PORT=50051
DB_HOST=mongodb
DB_PORT=27017
DB_NAME=esl
RABBIT_HOST=rabbitmq
RABBIT_PORT=5672
```

---

## Running the service
From the project root :

```cmd 
docker-compose up --build
```
or on second run to be safe :

```cmd 
 docker-compose down && docker-compose up --build
```

gRPC server: listens on :50051
MongoDB: accessible at :27017
RabbitMQ: listen on :5672

---

## Basic example

Create an user :
```create
grpcurl -plaintext -d '{
    "first_name": "FaceIt",
    "last_name": "AT",
    "nickname": "nickname",
    "email": "faceit@faceit.com",
    "country": "FR",
    "password": "supersecurepassword" 
}' localhost:50051 user.UserService/CreateUser
```

will return :
```
{
    "id": "8501f835-e1d1-4f6d-a8a4-b9b34dce65e4",
    "updatedAt": "2025-03-25T18:24:51.148668138Z"
}
```

Update user :
```update
grpcurl -plaintext -d '{
    "id": "8501f835-e1d1-4f6d-a8a4-b9b34dce65e4",
    "first_name": "new first name",
    "last_name": "AT",
    "nickname": "new nickname for example",
    "email": "faceit@faceit.com",
    "country": "FR",
    "password": "supersecurepassword" 
}' localhost:50051 user.UserService/UpdateUser
```

Get a list of user with filter, filters are first_name, last_name, country :
```
grpcurl -plaintext -d '{
"first_name": "new first name"
}' localhost:50051 user.UserService/ListUsers
```

```
grpcurl -plaintext -d '{
"country": "FR"
}' localhost:50051 user.UserService/ListUsers
```

```
grpcurl -plaintext -d '{
"last_name": "AT"
}' localhost:50051 user.UserService/ListUsers
```

Delete user :
```
grpcurl -plaintext -d '{
"id": "8501f835-e1d1-4f6d-a8a4-b9b34dce65e4"
}' localhost:50051 user.UserService/DeleteUser
```



---

## HealthCheck

```ht
grpcurl -plaintext localhost:50051 grpc.health.v1.Health/Check
```

---

## Testing
Run all tests (unit + integration) in isolation:
```test
docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
```
Integration tests use an in‑memory MongoDB
Database is wiped after each run

---

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

---

## Logging and error
- Using Go's slog package
- Listening for signal leading to mongo being disconnected and graceful shutdown of gRPC server

---

## Next step
- Add TLS to gRPC in production
- Expose Prometheus metrics endpoint
- Add new entity like Games for example
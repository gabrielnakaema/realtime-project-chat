# Backend - Go API Server

A Go backend service implementing event-driven patterns with Kafka pub/sub messaging and WebSocket real-time communication.

## 🎯 Learning Focus

This backend demonstrates:

- **Kafka Integration** for event-driven pub/sub messaging
- **WebSocket Management** for real-time chat functionality
- **Type-safe Database Operations** with SQLC code generation
- **JWT Authentication** with secure token handling

## 🛠️ Technology Stack

- **Go 1.25**
- **Chi Router** for HTTP routing with middleware
- **PostgreSQL** with pgx/v5 driver and connection pooling
- **SQLC** for type-safe SQL query code generation
- **Kafka** with Sarama client for pub/sub messaging
- **WebSockets** with gorilla/websocket for real-time communication
- **JWT** tokens for stateless authentication
- **Testcontainers** for integration testing with real databases

## 🚀 Quick Start

### Prerequisites

- Go 1.25+
- Docker & Docker Compose
- Air (for hot reload): `go install github.com/air-verse/air@latest`

### Development Setup

1. **Start infrastructure services**

   ```bash
   docker-compose up -d
   ```

2. **Set up environment**

   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run database migrations**

   ```bash
   make goose-up
   ```

4. **Generate SQLC code** (if needed)

   ```bash
   sqlc generate
   ```

5. **Start development server**
   ```bash
   air  # Hot reload
   # or
   go run ./cmd/api
   ```

## 🛠️ Available Commands

### Database Operations

```bash
make goose-up              # Run migrations up
make goose-down            # Run migrations down
make goose-create name=... # Create new migration
sqlc generate              # Generate Go code from SQL
```

### Development

```bash
go build -o ./tmp/main ./cmd/api  # Build application
air                               # Start with hot reload
go run ./cmd/api                  # Run directly
go mod tidy                       # Clean dependencies
```

### Testing

```bash
go test ./internal/...                    # Run all tests
go test -v ./internal/service/            # Service layer tests
go test -timeout 3m ./internal/repository/ # Integration tests
```

## 🔧 Configuration

Required environment variables:

```bash
# Database
DB_DSN=postgres://user:pass@localhost/dbname

# Kafka
PUBSUB_BROKERS=localhost:9092

# Authentication
JWT_SECRET=your-secret-key

# Environment
ENV=development
```

## 📚 Key Learning Concepts

- **Event Sourcing**: Domain events for system integration
- **Type Safety**: SQLC for compile-time SQL validation
- **Concurrency**: Goroutines for WebSocket and Kafka handling

---

**Note**: This backend serves as a learning project for Go microservices patterns and is not intended for production use.

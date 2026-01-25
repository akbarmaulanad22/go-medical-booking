# Go Template Clean Architecture

A production-ready Go REST API template implementing **Clean Architecture** with authentication using **JWT + Redis** for secure token management.

## 🏗️ Architecture

This project follows the Clean Architecture principles:

```
┌─────────────────────────────────────────────────────────────┐
│                     Delivery Layer                          │
│  (HTTP Handlers, Middleware, DTOs, Router)                  │
├─────────────────────────────────────────────────────────────┤
│                     Usecase Layer                           │
│  (Business Logic, Application Services)                     │
├─────────────────────────────────────────────────────────────┤
│                     Domain Layer                            │
│  (Entities, Repository Interfaces)                          │
├─────────────────────────────────────────────────────────────┤
│                  Infrastructure Layer                       │
│  (Database, Cache, External Services)                       │
└─────────────────────────────────────────────────────────────┘
```

## 📁 Project Structure

```
go-template-clean-architecture/
├── cmd/
│   └── main.go                    # Entry point
├── config/
│   └── config.go                  # Configuration management
├── docs/
│   └── openapi.json               # API specification (OpenAPI 3.0)
├── internal/
│   ├── domain/
│   │   ├── entity/                # Business entities
│   │   └── repository/            # Repository interfaces
│   ├── usecase/                   # Business logic
│   ├── repository/                # Repository implementations
│   ├── delivery/
│   │   ├── http/
│   │   │   ├── handler/           # HTTP handlers
│   │   │   ├── middleware/        # HTTP middleware
│   │   │   └── router.go          # Route definitions
│   │   └── dto/                   # Data transfer objects
│   └── infrastructure/
│       ├── database/              # Database connection
│       └── cache/                 # Redis connection
├── pkg/
│   ├── jwt/                       # JWT utilities
│   ├── response/                  # API response helpers
│   └── validator/                 # Request validation
├── migrations/                    # SQL migrations
├── .env.example                   # Environment template
├── Makefile                       # Build commands
└── README.md
```

## ✨ Features

- **Clean Architecture** - Separation of concerns with dependency injection
- **Authentication**
  - User registration
  - Login with JWT tokens
  - Logout (token revocation)
  - Refresh token rotation
  - Get current user
- **JWT + Redis Security** - Tokens stored in Redis for revocation support
- **CRUD Example** - Complete Product CRUD with pagination
- **Database Migration** - SQL migrations with golang-migrate
- **Request Validation** - Input validation with go-playground/validator
- **Structured Logging** - JSON logging with logrus
- **CORS Support** - Cross-origin resource sharing
- **Graceful Shutdown** - Clean server shutdown

## 🛠️ Tech Stack

| Package | Purpose |
|---------|---------|
| [gorilla/mux](https://github.com/gorilla/mux) | HTTP router |
| [gorm](https://gorm.io/) | ORM |
| [go-redis](https://github.com/redis/go-redis) | Redis client |
| [golang-jwt](https://github.com/golang-jwt/jwt) | JWT handling |
| [viper](https://github.com/spf13/viper) | Configuration |
| [logrus](https://github.com/sirupsen/logrus) | Logging |
| [validator](https://github.com/go-playground/validator) | Validation |
| [golang-migrate](https://github.com/golang-migrate/migrate) | Migrations |

## 🚀 Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 15+
- Redis 7+
- Make
- [golang-migrate CLI](https://github.com/golang-migrate/migrate/tree/master/cmd/migrate)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/akbarmaulanad22/go-clean-architecture-template.git
   cd go-template-clean-architecture
   ```

2. **Copy environment file**
   ```bash
   cp .env.example .env
   ```

3. **Edit `.env` with your configuration**
   ```env
   APP_PORT=8080
   APP_ENV=development

   DB_HOST=localhost
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=postgres
   DB_NAME=clean_architecture

   REDIS_HOST=localhost
   REDIS_PORT=6379
   REDIS_PASSWORD=
   REDIS_DB=0

   JWT_SECRET=your-super-secret-key
   JWT_ACCESS_EXPIRY=15m
   JWT_REFRESH_EXPIRY=168h
   ```

4. **Setup PostgreSQL**
   
   Make sure PostgreSQL is installed and running. Create the database:
   ```sql
   CREATE DATABASE clean_architecture;
   ```

5. **Setup Redis**
   
   Make sure Redis is installed and running on `localhost:6379`.

   **Windows:**
   - Download Redis from [GitHub Releases](https://github.com/microsoftarchive/redis/releases) or use [Memurai](https://www.memurai.com/)
   
   **macOS:**
   ```bash
   brew install redis
   brew services start redis
   ```
   
   **Linux:**
   ```bash
   sudo apt install redis-server
   sudo systemctl start redis
   ```

6. **Install migrate CLI**
   ```bash
   # Windows (using scoop)
   scoop install migrate

   # macOS
   brew install golang-migrate

   # Linux
   curl -L https://github.com/golang-migrate/migrate/releases/download/v4.17.0/migrate.linux-amd64.tar.gz | tar xvz
   sudo mv migrate /usr/local/bin/
   ```

7. **Run migrations**
   ```bash
   make migrate-up
   ```

8. **Run the application**
   ```bash
   make run
   ```

## 📝 Makefile Commands

| Command | Description |
|---------|-------------|
| `make run` | Run the application |
| `make build` | Build binary to `./build/` |
| `make test` | Run tests with coverage |
| `make clean` | Remove build artifacts |
| `make migrate-up` | Apply all migrations |
| `make migrate-down` | Rollback all migrations |
| `make migrate-down-one` | Rollback one migration |
| `make migrate-create name=xxx` | Create new migration |
| `make lint` | Run golangci-lint |
| `make tidy` | Run go mod tidy |
| `make deps` | Download dependencies |

## 📖 API Documentation

Complete API documentation is available in OpenAPI 3.0 format:

📄 **[docs/openapi.json](docs/openapi.json)**

You can use the following tools to view the documentation:

- **Swagger Editor**: https://editor.swagger.io - Paste the contents of `openapi.json`
- **Swagger UI**: Import the file for interactive view
- **Postman**: Import OpenAPI file to generate collection
- **Insomnia**: Import OpenAPI file

### Quick API Reference

#### Authentication

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| POST | `/api/v1/auth/register` | Register new user | ❌ |
| POST | `/api/v1/auth/login` | Login user | ❌ |
| POST | `/api/v1/auth/logout` | Logout user | ✅ |
| POST | `/api/v1/auth/refresh-token` | Refresh tokens | ❌ |
| GET | `/api/v1/auth/me` | Get current user | ✅ |

#### Products (CRUD Example)

| Method | Endpoint | Description | Auth |
|--------|----------|-------------|------|
| GET | `/api/v1/products` | List all products | ❌ |
| GET | `/api/v1/products/{id}` | Get product by ID | ❌ |
| POST | `/api/v1/products` | Create product | ✅ |
| PUT | `/api/v1/products/{id}` | Update product | ✅ |
| DELETE | `/api/v1/products/{id}` | Delete product | ✅ |

#### Health Check

| Method | Endpoint | Description |
|--------|----------|-------------|
| GET | `/api/v1/health` | Health check |

## 🔐 JWT + Redis Security

Token security flow:

1. **Login**: Access and refresh tokens are generated and stored in Redis
2. **Request**: Middleware validates JWT signature AND checks token exists in Redis
3. **Logout**: Tokens are deleted from Redis, immediately invalidating them
4. **Token Theft**: Simply delete the token from Redis to revoke access

Redis key format:
- Access token: `access_token:{user_id}:{token_id}`
- Refresh token: `refresh_token:{user_id}:{token_id}`

## 🧪 Testing

```bash
# Run all tests
make test

# Run specific test
go test -v ./internal/usecase/...
```

## 📄 License

This project is licensed under the MIT License.

## 🤝 Contributing

1. Fork the repository
2. Create your feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

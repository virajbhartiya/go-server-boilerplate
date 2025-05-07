# Go Server Boilerplate

A modern, feature-rich Go server boilerplate following clean architecture principles and best practices.

## Features

- **Clean Architecture**: Domain-driven design with clear separation of concerns
- **GORM Integration**: Object-relational mapping with PostgreSQL
- **JWT Authentication**: Secure authentication with role-based access control
- **Middleware**: Request ID, rate limiting, CORS, recovery, etc.
- **Configuration**: TOML config files with environment variable overrides
- **Error Handling**: Standardized error handling with HTTP status codes
- **Validation**: Request validation using the validator package
- **Logging**: Structured logging with zap
- **Graceful Shutdown**: Proper handling of termination signals
- **API Structure**: Well-organized API structure with versioning
- **Security**: Best practices for web security

## Directory Structure

```
├── cmd/
│   └── api/
│       └── main.go          # Application entry point
├── internal/
│   ├── app/                 # Application core
│   │   ├── domain/          # Domain models and business logic
│   │   ├── ports/           # Input/output interfaces
│   │   └── services/        # Business logic implementation
│   ├── config/              # Configuration management
│   ├── infrastructure/      # External systems integration
│   │   ├── auth/            # Authentication implementation
│   │   ├── cache/           # Caching implementation
│   │   ├── database/        # Database adapters and migrations
│   │   │   └── models/      # GORM model definitions
│   │   ├── jobs/            # Background job processing
│   │   └── server/          # HTTP server implementation
│   ├── interfaces/          # Interface adapters
│   │   ├── api/             # API controllers and routing
│   │   ├── events/          # Event handlers
│   │   └── workers/         # Background worker handlers
│   └── pkg/                 # Shared utilities
│       ├── logger/          # Logging utilities
│       ├── middleware/      # HTTP middleware
│       ├── validator/       # Request validation
│       └── errors/          # Error handling
├── configs/                 # Configuration files (TOML)
├── .env.example             # Example environment variables
└── go.mod                   # Go module definition
```

## Getting Started

### Prerequisites

- Go 1.22+
- PostgreSQL
- Docker (optional, for containerized development)

### Installation

1. Clone the repository:

```bash
git clone https://github.com/virajbhartiya/go-server-boilerplate.git
cd go-server-boilerplate
```

2. Install dependencies:

```bash
go mod download
```

3. Copy and edit the example config:

```bash
cp configs/config.example.toml configs/config.toml
# Optionally, create configs/development.toml or configs/production.toml for environment-specific overrides
```

4. (Optional) Copy and edit the example .env file:

```bash
cp .env.example .env
# Environment variables in .env will override TOML config values
```

5. Start dependencies (PostgreSQL and Redis):

```bash
docker-compose up -d postgres redis
```

6. Run the application:

```bash
go run cmd/api/main.go
```

### Using Make

The project includes a Makefile with useful commands:

```bash
make build        # Build the application
make run          # Run the application with hot reload
make test         # Run tests
make docs         # Generate API documentation
make lint         # Run linter
make migrate-create name=migration_name  # Create a migration
make migrate-up   # Apply database migrations
make migrate-down # Rollback database migrations
```

## Configuration

Configuration is managed via TOML files in the `configs/` directory. The application loads `configs/config.toml` by default, then `configs/{ENVIRONMENT}.toml` if it exists, and finally overrides with environment variables (from the environment or `.env` file if loaded).

Example `configs/config.example.toml`:

```toml
[server]
port = "8080"
environment = "development"
shutdown_timeout = "10s"
read_timeout = "5s"
write_timeout = "5s"
idle_timeout = "60s"
ssl_enabled = false
ssl_cert_file = ""
ssl_key_file = ""

[database]
url = "postgresql://postgres:postgres@localhost:5432/app?sslmode=disable"
max_connections = 10
max_idle_connections = 5
conn_max_lifetime = "1h"
auto_migrate = true
log_queries = false
prepared_statements = false

[gorm]
log_level = "info"
prepared_stmt = false
skip_default_transaction = false

[api]
cors_enabled = true
allowed_origins = ["*"]
rate_limiter_enabled = false
rate_limit_requests = 100
rate_limit_duration = "1m"

[auth]
jwt_secret = "your_jwt_secret"
jwt_expiry_hours = 24
refresh_token_enabled = true
refresh_token_expiry = "168h"

[logging]
level = "info"
format = "console"
caller_enabled = false
stacktrace_enabled = false

[cache]
enabled = true
redis_url = "redis://localhost:6379/0"
default_ttl = "1h"

[features]
tracing = false
background_jobs = false

[redis]
host = "localhost"
port = "6379"
password = ""
db = 0
```

## Authentication

The application uses JWT for authentication:

1. Register a user with `/api/v1/auth/register`
2. Log in with `/api/v1/auth/login` to get a JWT token
3. Include the token in the `Authorization` header as `Bearer <token>`

## API Documentation

API documentation is available at `/swagger/index.html` when the application is running in development mode.

## Docker

To run the application with Docker:

```bash
docker build -t go-server-boilerplate .
docker run -p 8080:8080 --env-file .env go-server-boilerplate
```

Alternatively, use Docker Compose:

```bash
docker-compose up
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

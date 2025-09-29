# Go Server Boilerplate

A modern, feature-rich Go server boilerplate following clean architecture principles and best practices.

## Features

- **Clean Architecture**: Domain-driven design with clear separation of concerns
- **GORM Integration**: Object-relational mapping with PostgreSQL
- **JWT Authentication**: Secure authentication with role-based access control
- **Middleware**: Request ID, CORS, recovery
- **Configuration**: Environment variables only (with `.env` support)
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
│   ├── config/              # Configuration management (env-driven)
│   ├── infrastructure/      # External systems integration
│   │   ├── auth/            # Authentication implementation
│   │   ├── database/        # Database adapters and migrations
│   │   │   └── models/      # GORM model definitions
│   │   └── jobs/            # Background job processing
│   ├── interfaces/          # Interface adapters
│   │   └── api/             # API controllers and routing
│   └── pkg/                 # Shared utilities
│       ├── logger/          # Logging utilities
│       ├── middleware/      # HTTP middleware
│       ├── validator/       # Request validation
│       └── errors/          # Error handling
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

3. Copy and edit the example .env file:

```bash
cp .env.example .env
# Update values as needed; the app reads configuration from the environment
```

4. Start dependencies (PostgreSQL):

```bash
docker-compose up -d postgres
```

5. Run the application:

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

Configuration is managed via environment variables only. See `.env.example` for all supported keys and sensible defaults. You can export variables in your shell or place them in a `.env` file (loaded by the app on startup).

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

Alternatively, use Docker Compose (to start Postgres):

```bash
docker-compose up -d postgres
```

## License

This project is licensed under the MIT License - see the LICENSE file for details.

## Contributing

Contributions are welcome! Please feel free to submit a Pull Request.

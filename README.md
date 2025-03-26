# Kogase Backend

The Kogase Backend is a Go-based API for game analytics and telemetry data collection. It provides endpoints for tracking player behavior, sessions, and events in games.

## Features

- User authentication and authorization
- Project management
- Real-time telemetry data collection
- Session tracking
- Analytics and reporting

## Getting Started

### Prerequisites

- Go 1.24 or higher
- PostgreSQL database
- Git

### Installation

1. Clone the repository

```bash
git clone https://github.com/your-username/kogase.git
cd kogase/backend
```

2. Install dependencies

```bash
go mod download
```

3. Set up environment variables (or create a `.env` file)

```bash
# Database settings
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=your_password
DB_NAME=kogase
DB_SSL_MODE=disable

# JWT settings
JWT_SECRET=your_secret_key
JWT_EXPIRATION=24h

# Server settings
PORT=8080
```

4. Run the server

```bash
go run main.go
```

## API Documentation

The API is documented using Swagger. When the server is running, visit:

```
http://localhost:8080/swagger/index.html
```

You can test all endpoints directly from the Swagger UI.

### Authentication

The API uses two types of authentication:

1. **JWT Tokens** - For dashboard and admin access
   - Obtain a token via `/api/v1/auth/login`
   - Use the token in the `Authorization` header: `Bearer your_token`

2. **API Keys** - For SDK and game clients
   - Generate an API key for your project in the dashboard
   - Use the key in the `X-API-Key` header

## Project Structure

```
.
├── config/         # Configuration management
├── controllers/    # API endpoint handlers
├── docs/           # Swagger documentation
├── middleware/     # Request middleware
├── models/         # Database models
├── server/         # Server setup and routing
├── tests/          # Tests
├── utils/          # Utility functions
├── go.mod          # Go module file
├── go.sum          # Go dependencies
└── main.go         # Application entry point
```

## Testing

To run tests:

```bash
cd tests
go test -v ./...
```

## Generating Swagger Documentation

The API documentation is generated using Swaggo. To update the documentation:

1. Install the Swagger CLI tool:

```bash
go install github.com/swaggo/swag/cmd/swag@latest
```

2. Generate the documentation:

```bash
swag init -g main.go
```

## License

This project is licensed under the MIT License - see the LICENSE file for details. 
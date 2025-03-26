# Kogase Backend Tests

This directory contains tests for the Kogase backend.

## Integration Tests

Integration tests are located in the `integration` directory and test the full API functionality, including database interactions.

### Prerequisites

To run the integration tests, you need:

1. A PostgreSQL database server (can be local or remote)
2. Go installed on your machine

### Configuration

The tests use the following environment variables, which have default values in the run scripts:

- `DB_HOST`: PostgreSQL host (default: localhost)
- `DB_PORT`: PostgreSQL port (default: 5432)
- `DB_USER`: PostgreSQL username (default: postgres)
- `DB_PASSWORD`: PostgreSQL password (default: postgres)
- `DB_NAME`: Test database name (default: kogase_test)
- `DB_SSL_MODE`: SSL mode for database connection (default: disable)
- `JWT_SECRET`: Secret for JWT tokens in tests (default: test_secret)
- `PORT`: Port to run the test server on (default: 8081)

You can modify these values in the run scripts or set them directly in your environment.

### Running the Tests

#### On Linux/Mac:

```bash
# Make the script executable
chmod +x ./run_tests.sh

# Run the tests
./run_tests.sh
```

#### On Windows:

```powershell
# Run the tests
.\run_tests.ps1
```

### Test Structure

The integration tests are organized by feature:

- `main_test.go`: Sets up the test environment, including database connection and test router
- `auth_test.go`: Tests authentication functionality (register, login)
- `project_test.go`: Tests project management (CRUD operations, API key management)
- `telemetry_test.go`: Tests telemetry event recording (single events, batch events, sessions)

## Adding New Tests

When adding new tests:

1. Create a new test file in the `integration` directory if testing a new feature
2. Add test cases following the existing patterns
3. Make sure to clean up any test data in the `setupTest*` functions
4. Use the `testRouter` and `testDB` variables that are initialized in `main_test.go`

## Running Individual Tests

To run a specific test file or function:

```bash
# Run a specific test file
go test -v ./integration/auth_test.go

# Run a specific test function
go test -v ./integration -run TestRegisterUser
``` 
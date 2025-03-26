# Set environment variables for test
$env:DB_HOST = "localhost"
$env:DB_PORT = "5432"
$env:DB_USER = "postgres"
$env:DB_PASSWORD = "postgres"
$env:DB_NAME = "kogase_test"
$env:DB_SSL_MODE = "disable"
$env:JWT_SECRET = "test_secret"
$env:PORT = "8081"

# Check if PostgreSQL is installed and available
try {
    $pgVersion = & psql --version
    Write-Host "PostgreSQL detected: $pgVersion"
} catch {
    Write-Host "Error: PostgreSQL command line tools not found or not in PATH" -ForegroundColor Red
    Write-Host "Please install PostgreSQL or make sure the commands are available in your PATH" -ForegroundColor Red
    exit 1
}

# Create test database if it doesn't exist
Write-Host "Creating test database if it doesn't exist..."
$env:PGPASSWORD = $env:DB_PASSWORD
$dbExists = & psql -h $env:DB_HOST -U $env:DB_USER -t -c "SELECT 1 FROM pg_database WHERE datname = '$($env:DB_NAME)'"

if (-not $dbExists) {
    Write-Host "Creating database $($env:DB_NAME)..."
    & psql -h $env:DB_HOST -U $env:DB_USER -c "CREATE DATABASE $($env:DB_NAME)"
}

# Run the integration tests
Write-Host "Running integration tests..." -ForegroundColor Cyan
Push-Location -Path .\integration
try {
    & go test -v ./...
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Tests failed with exit code $LASTEXITCODE" -ForegroundColor Red
        exit $LASTEXITCODE
    }
} finally {
    Pop-Location
}

Write-Host "All tests passed!" -ForegroundColor Green 
# Testing Guide

This document describes how to run tests, check coverage, and perform load testing.

## Unit and Integration Tests

### Running Tests

Run all tests:
```bash
go test ./... -v
```

Run tests with coverage:
```bash
go test ./... -v -cover
```

Run tests with coverage report:
```bash
go test ./... -v -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

### Target Coverage

Aim for **70%+ coverage** across all packages.

### Test Files

- `internal/handlers/graphql_test.go` - GraphQL proxy handler tests
- `internal/middleware/auth_test.go` - Authentication middleware tests
- `internal/middleware/ratelimit_test.go` - Rate limiting middleware tests
- `internal/app/app_test.go` - Application integration tests

## Load Testing

### Prerequisites

Install Vegeta:
```bash
go install github.com/tsenart/vegeta@latest
```

### Running Load Tests

Use the provided script:
```bash
./loadtest.sh
```

Or run manually:
```bash
# Basic attack
echo "POST http://localhost:3000/graphql" | \
  vegeta attack \
    -rate=50 \
    -duration=1m \
    -body='{"query":"{ artists { id name currentPrice } }"}' \
    -header="Content-Type: application/json" | \
  vegeta report
```

### Load Test Configuration

Environment variables:
- `TARGET_URL` - Target endpoint (default: http://localhost:3000/graphql)
- `RATE` - Requests per second (default: 50)
- `DURATION` - Test duration (default: 1m)
- `OUTPUT_FILE` - Output file for results (default: loadtest_results.bin)

### Performance Targets

- **P95 latency**: Under 200ms
- **Success rate**: 99%+

### Example Output

```
Requests      [total, rate, throughput]  3000, 50.00, 49.95
Duration      [total, attack, wait]     1m0s, 59.99s, 5.2ms
Latencies     [mean, 50, 95, 99, max]   45.2ms, 42.1ms, 89.3ms, 120.5ms, 250.1ms
Bytes In      [total, mean]              450000, 150.00
Bytes Out     [total, mean]              600000, 200.00
Success       [ratio]                    100.00%
Status Codes  [code:count]               200:3000
```

## Docker Testing

### Build Docker Image

```bash
docker build -t fiber-mvp .
```

### Run Container

```bash
docker run -p 3000:3000 --env-file .env fiber-mvp
```

### Test in Container

```bash
# Health check
curl http://localhost:3000/health

# GraphQL query
curl -X POST http://localhost:3000/graphql \
  -H "Content-Type: application/json" \
  -d '{"query":"{ artists { id name } }"}'
```

## Continuous Integration

Tests are automatically run in CI/CD pipelines:
- Unit tests on every push
- Integration tests on pull requests
- Coverage reports generated

See `.github/workflows/test.yml` for CI configuration.


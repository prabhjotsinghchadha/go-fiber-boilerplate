# Go Fiber Backend API - Production Ready Boilerplate

![Release](https://img.shields.io/github/release/gofiber/boilerplate.svg)
[![Discord](https://img.shields.io/badge/discord-join%20channel-7289DA)](https://gofiber.io/discord)
![Test](https://github.com/gofiber/boilerplate/workflows/Test/badge.svg)
![Security](https://github.com/gofiber/boilerplate/workflows/Security/badge.svg)
![Linter](https://github.com/gofiber/boilerplate/workflows/Linter/badge.svg)
![Go](https://img.shields.io/badge/Go-1.25+-00ADD8?logo=go)
![Fiber](https://img.shields.io/badge/Fiber-v2-00ADD8)
![Supabase](https://img.shields.io/badge/Supabase-Enabled-3ECF8E)
![Redis](https://img.shields.io/badge/Redis-Upstash-DC382D)

> **Standalone, production-ready Go Fiber backend API** - Connect your frontend apps (React, Next.js, Flutter, React Native) via HTTP and WebSocket

## Table of Contents

-   [What is This?](#what-is-this)
-   [Features](#features)
-   [Prerequisites](#prerequisites)
-   [Quick Start](#quick-start)
-   [Environment Variables](#environment-variables)
-   [Installation & Setup](#installation--setup)
-   [Project Structure](#project-structure)
-   [Features Documentation](#features-documentation)
-   [API Endpoints](#api-endpoints)
-   [Frontend Integration](#frontend-integration)
-   [Deployment](#deployment)
-   [Development](#development)
-   [Testing](#testing)
-   [Production Considerations](#production-considerations)
-   [Troubleshooting](#troubleshooting)
-   [Contributing](#contributing)
-   [License](#license)

## What is This?

This is a **standalone backend API** built with Go Fiber. It runs independently as a separate service that your frontend applications connect to.

### Architecture

```
┌─────────────────┐         HTTP/WebSocket          ┌──────────────────┐
│                 │ ◄─────────────────────────────── │                  │
│  Frontend App   │                                 │  Go Fiber API    │
│  (React/Next.js │                                 │  (This Project)  │
│  Flutter/RN)    │ ───────────────────────────────► │                  │
│                 │         API Requests            │                  │
└─────────────────┘                                 └──────────────────┘
                                                              │
                                                              │
                                                              ▼
                                                    ┌──────────────────┐
                                                    │   Supabase      │
                                                    │   Redis/Upstash  │
                                                    └──────────────────┘
```

### Key Points

-   **Separate Projects**: Your frontend and backend are separate projects/repositories
-   **API Communication**: Frontend apps make HTTP requests and WebSocket connections to this backend
-   **Production Ready**: Includes authentication, rate limiting, caching, and real-time features
-   **Fast**: Built with Go Fiber, one of the fastest Go web frameworks
-   **Flexible Deployment**: Deploy to Fly.io, Render, Railway, Supabase, or any Docker-compatible platform

### Use Cases

-   RESTful API backend for web/mobile applications
-   Real-time applications with WebSocket support
-   GraphQL proxy to Supabase
-   High-performance API requiring rate limiting and caching
-   Microservices architecture

## Features

-   ✅ **Go Fiber v2** - High-performance web framework
-   ✅ **Supabase Integration** - GraphQL proxy, Realtime subscriptions, JWT authentication
-   ✅ **Redis/Upstash Caching** - Fast data caching with Upstash Redis
-   ✅ **JWT Authentication** - Supports HS256 and RS256 tokens with Supabase JWKS
-   ✅ **Rate Limiting** - Per-user or per-IP rate limiting
-   ✅ **WebSocket Support** - Real-time communication hub
-   ✅ **CORS Configuration** - Secure cross-origin resource sharing
-   ✅ **Health Checks** - Built-in health check endpoint
-   ✅ **Comprehensive Testing** - Unit, integration, and load tests
-   ✅ **Docker Support** - Multi-stage Docker builds for production
-   ✅ **Production Ready** - Security best practices, error handling, logging

## Prerequisites

Before you begin, ensure you have:

-   **Go 1.25+** installed ([Download](https://golang.org/dl/))
-   **Docker** (optional, for containerized deployment)
-   **Supabase Account** ([Sign up](https://supabase.com))
-   **Upstash Redis Account** (optional, [Sign up](https://upstash.com))
-   Basic knowledge of Go, REST APIs, and your chosen frontend framework

## Quick Start

### Running Locally (Without Docker)

This is the simplest way to get started. No Docker required!

### 1. Clone the Repository

```bash
git clone <your-repo-url>
cd go-backend
```

### 2. Install Dependencies

```bash
go mod download
```

### 3. Set Up Environment Variables

Create a `.env` file in the root directory:

```bash
cp .env.example .env
```

Edit `.env` and fill in your configuration values (see [Environment Variables](#environment-variables) section).

**Minimum required for local development:**

```bash
SUPABASE_URL=https://xxxxx.supabase.co
SUPABASE_ANON_KEY=your-anon-key
PORT=3000
```

### 4. Run the Server Locally

**Option 1: Using Go directly**

```bash
go run ./cmd/server
```

**Option 2: Using Make command**

```bash
make run-local
```

The server will start on `http://localhost:3000` (or your configured PORT).

### 5. Verify Installation

Test the health check endpoint:

```bash
curl http://localhost:3000/health
```

You should see:

```json
{ "status": "ok" }
```

### 6. Access the Demo Page

Open your browser and navigate to:

```
http://localhost:3000/demo
```

The demo page provides:

-   Complete documentation
-   Interactive API testing
-   WebSocket connection tester
-   Rate limiter tester with graph visualization
-   Code examples for all frontend frameworks
-   Step-by-step guides

### Running with Docker (Optional)

If you prefer to run in Docker, see the [Development](#development) section below.

## Environment Variables

Create a `.env` file in the root directory with the following variables:

### Required Variables

| Variable            | Description               | Example                     |
| ------------------- | ------------------------- | --------------------------- |
| `SUPABASE_URL`      | Your Supabase project URL | `https://xxxxx.supabase.co` |
| `SUPABASE_ANON_KEY` | Supabase anonymous key    | Found in Supabase dashboard |

### Optional Variables

| Variable                     | Description                            | Default                                |
| ---------------------------- | -------------------------------------- | -------------------------------------- |
| `PORT`                       | Server port                            | `3000`                                 |
| `JWT_SECRET`                 | Secret for HS256 JWT tokens            | Required for HS256 auth                |
| `UPSTASH_REDIS_URL`          | Upstash Redis REST API URL             | Optional (caching disabled if not set) |
| `UPSTASH_REDIS_TOKEN`        | Upstash Redis token                    | Optional                               |
| `RATE_LIMIT_MAX`             | Max requests per minute                | `100`                                  |
| `ALLOWED_ORIGINS`            | CORS allowed origins (comma-separated) | Development defaults                   |
| `ENABLE_TRUSTED_PROXY_CHECK` | Enable proxy support                   | `false`                                |
| `TRUSTED_PROXIES`            | Trusted proxy IPs/CIDRs                | Empty                                  |
| `GO_ENV` or `ENV`            | Environment mode                       | `development`                          |

See `.env.example` for a complete template with descriptions.

## Installation & Setup

### Step 1: Clone and Install

```bash
git clone <your-repo-url>
cd go-backend
go mod download
```

### Step 2: Supabase Setup

1. Create a Supabase account at [supabase.com](https://supabase.com)
2. Create a new project
3. Go to Settings > API
4. Copy your Project URL and anon/public key
5. Add them to your `.env` file:
    ```
    SUPABASE_URL=https://xxxxx.supabase.co
    SUPABASE_ANON_KEY=your-anon-key
    ```

### Step 3: Upstash Redis Setup (Optional)

1. Create an account at [upstash.com](https://upstash.com)
2. Create a new Redis database
3. Copy the REST API URL and token
4. Add them to your `.env` file:
    ```
    UPSTASH_REDIS_URL=https://xxxxx.upstash.io
    UPSTASH_REDIS_TOKEN=your-token
    ```

### Step 4: Configure Environment

1. Copy `.env.example` to `.env`
2. Fill in all required variables
3. For production, set `GO_ENV=production` and configure `ALLOWED_ORIGINS`

### Step 5: Run the Server

```bash
go run ./cmd/server
```

The server will start on `http://localhost:3000` (or your configured PORT).

### Step 6: Test the Setup

1. Visit `http://localhost:3000/health` - Should return `{"status":"ok"}`
2. Visit `http://localhost:3000/demo` - Interactive demo and documentation

## Project Structure

```
go-backend/
├── cmd/
│   └── server/
│       └── main.go              # Application entry point
├── internal/
│   ├── app/
│   │   ├── app.go              # Fiber app configuration
│   │   └── app_test.go        # App tests
│   ├── cache/
│   │   └── redis.go           # Redis/Upstash client
│   ├── handlers/
│   │   ├── graphql.go         # GraphQL proxy handler
│   │   ├── ws.go              # WebSocket handler
│   │   └── demo.go            # Demo page handler
│   ├── middleware/
│   │   ├── auth.go            # JWT authentication
│   │   └── ratelimit.go       # Rate limiting
│   └── realtime/
│       └── subscriber.go      # Supabase Realtime subscriptions
├── .env.example                # Environment variables template
├── Dockerfile                  # Docker build configuration
├── fly.toml                    # Fly.io deployment config
├── render.yaml                 # Render deployment config
├── go.mod                      # Go dependencies
└── README.md                   # This file
```

### Key Files

-   **`cmd/server/main.go`**: Application entry point, initializes services
-   **`internal/app/app.go`**: Configures Fiber app, middleware, and routes
-   **`internal/handlers/`**: Request handlers for endpoints
-   **`internal/middleware/`**: Authentication and rate limiting middleware
-   **`internal/cache/redis.go`**: Redis caching implementation
-   **`internal/realtime/subscriber.go`**: Supabase Realtime integration

### Adding Custom Routes

To add custom routes, edit `internal/app/app.go` in the `setupRoutes` function:

```go
// Add your custom routes here
api.Get("/your-endpoint", yourHandler)
```

## Features Documentation

### Authentication

The API supports JWT authentication with two methods:

-   **HS256**: Symmetric signing (requires `JWT_SECRET`)
-   **RS256**: Asymmetric signing (uses Supabase JWKS)

**How it works:**

1. Frontend obtains JWT token from Supabase Auth (or your auth provider)
2. Frontend includes token in `Authorization: Bearer <token>` header
3. Backend validates token and extracts user ID
4. User ID is available in handlers via `c.Locals("user")`

**Protected Routes:**

Routes under `/api/*` require authentication. The auth middleware:

-   Validates JWT token
-   Extracts user ID from token claims
-   Attaches user ID to request context
-   Returns 401 if authentication fails

**Testing Authentication:**

Use the demo page at `/demo` to test authentication flows and see example requests.

### Rate Limiting

Rate limiting prevents abuse by limiting requests per time period.

**How it works:**

-   **Authenticated users**: Rate limit is per user ID (more accurate)
-   **Unauthenticated users**: Rate limit is per IP address
-   Default: 100 requests per minute (configurable via `RATE_LIMIT_MAX`)

**Configuration:**

-   Set `RATE_LIMIT_MAX` in `.env` to change the limit
-   For production behind proxies, configure `ENABLE_TRUSTED_PROXY_CHECK` and `TRUSTED_PROXIES`

**Rate Limit Response:**

When limit is exceeded, API returns:

-   Status: `429 Too Many Requests`
-   Body: `{"error": "Rate limit exceeded"}`

### GraphQL Proxy

The API proxies GraphQL requests to Supabase's GraphQL endpoint.

**Endpoint:** `POST /graphql`

**How it works:**

1. Frontend sends GraphQL query to `/graphql`
2. Backend forwards request to Supabase GraphQL API
3. Backend injects cached data (if available)
4. Response is returned to frontend

**Features:**

-   Preserves all headers (including Authorization)
-   Automatic price caching for `currentPrice` queries
-   Error handling and logging

**Usage:**

Send GraphQL queries to `http://your-backend-url/graphql` with:

-   Method: `POST`
-   Headers: `Content-Type: application/json`, `Authorization: Bearer <token>` (if needed)
-   Body: `{"query": "...", "variables": {...}}`

### WebSocket Support

Real-time communication via WebSocket connections.

**Endpoint:** `GET /ws` (WebSocket upgrade)

**How it works:**

1. Client connects to `ws://your-backend-url/ws`
2. Connection is registered with the WebSocket hub
3. Backend can broadcast messages to all connected clients
4. Clients receive real-time updates

**Message Format:**

Messages are JSON-encoded:

```json
{
    "artist_id": "123",
    "price": 45.67,
    "event": "UPDATE"
}
```

**Use Cases:**

-   Real-time price updates
-   Live notifications
-   Chat applications
-   Collaborative features

**Testing:**

Use the demo page at `/demo` to test WebSocket connections interactively.

### Redis Caching

Optional caching layer using Upstash Redis.

**Setup:**

1. Create Upstash Redis database
2. Add `UPSTASH_REDIS_URL` and `UPSTASH_REDIS_TOKEN` to `.env`
3. Backend automatically uses cache if configured

**Features:**

-   Automatic cache key management
-   Configurable TTL (default: 5 minutes)
-   Cache invalidation support

**Cache Keys:**

-   Price cache: `price:{artist_id}`
-   Custom keys can be added in handlers

**Usage in Code:**

```go
cache.GetClient().Set("key", "value", 5*time.Minute)
value, _ := cache.GetClient().Get("key")
```

### Supabase Realtime

Automatic database change subscriptions.

**How it works:**

1. Backend subscribes to Supabase Realtime for `artist_metrics` table
2. When database changes occur, backend:
    - Caches the new price in Redis
    - Broadcasts update to all WebSocket clients

**Configuration:**

-   Requires `SUPABASE_URL` and `SUPABASE_ANON_KEY`
-   Table must have Realtime enabled in Supabase dashboard
-   Backend automatically reconnects on connection loss

## API Endpoints

### Public Endpoints

#### `GET /health`

Health check endpoint.

**Response:**

```json
{ "status": "ok" }
```

#### `POST /graphql`

GraphQL proxy to Supabase.

**Headers:**

-   `Content-Type: application/json`
-   `Authorization: Bearer <token>` (optional, for protected queries)

**Body:**

```json
{
    "query": "query { artists { id name } }",
    "variables": {}
}
```

#### `GET /ws`

WebSocket endpoint for real-time communication.

**Upgrade:** HTTP request is upgraded to WebSocket connection.

### Protected Endpoints

All endpoints under `/api/*` require authentication.

#### `GET /api/profile`

Example protected endpoint.

**Headers:**

-   `Authorization: Bearer <jwt-token>`

**Response:**

```json
{
    "user": "user-id-from-token"
}
```

## Frontend Integration

**Important:** Your frontend app is a **separate project** that connects to this backend API.

### Quick Integration Steps

1. **Set Backend URL**: Configure your frontend to point to the deployed backend URL
2. **Make HTTP Requests**: Use fetch/axios to call API endpoints
3. **Handle Authentication**: Include JWT token in `Authorization` header
4. **Connect WebSocket**: Connect to `ws://your-backend-url/ws` for real-time updates

### Detailed Integration Guides

**For complete code examples and step-by-step guides, visit the demo page:**

```
http://localhost:3000/demo
```

The demo page includes:

-   ✅ **React Integration** - Complete examples with hooks, context, error handling
-   ✅ **Next.js Integration** - Server/client components, SSR/SSG patterns
-   ✅ **Flutter Integration** - HTTP client, WebSocket, state management
-   ✅ **React Native Integration** - AsyncStorage, network config, platform considerations

Each integration guide includes:

-   Setup instructions
-   Environment configuration
-   Authentication flow
-   API request examples
-   WebSocket connection examples
-   Error handling patterns
-   Copy-paste ready code snippets

## Deployment

This backend can be deployed to multiple platforms. All platforms support Docker-based deployment.

### Fly.io Deployment

1. **Install Fly CLI:**

    ```bash
    curl -L https://fly.io/install.sh | sh
    ```

2. **Login to Fly.io:**

    ```bash
    fly auth login
    ```

3. **Create App:**

    ```bash
    fly launch
    ```

    Edit `fly.toml` and set your app name.

4. **Set Environment Variables:**

    ```bash
    fly secrets set SUPABASE_URL=xxx SUPABASE_ANON_KEY=xxx JWT_SECRET=xxx
    fly secrets set ALLOWED_ORIGINS=https://yourdomain.com
    ```

5. **Deploy:**
    ```bash
    fly deploy
    ```

See `fly.toml` for configuration details.

### Render Deployment

1. **Connect Repository:**

    - Go to [render.com](https://render.com)
    - Connect your GitHub repository

2. **Create Web Service:**

    - Select "Web Service"
    - Choose your repository
    - Render will detect `render.yaml`

3. **Configure Environment Variables:**

    - Add all required environment variables in Render dashboard
    - Set `GO_ENV=production`
    - Configure `ALLOWED_ORIGINS`

4. **Deploy:**
    - Render automatically deploys on git push
    - Or manually trigger deployment

See `render.yaml` for configuration details.

### Railway Deployment

1. **Install Railway CLI:**

    ```bash
    npm i -g @railway/cli
    ```

2. **Login:**

    ```bash
    railway login
    ```

3. **Initialize Project:**

    ```bash
    railway init
    ```

4. **Set Environment Variables:**

    ```bash
    railway variables set SUPABASE_URL=xxx
    railway variables set SUPABASE_ANON_KEY=xxx
    # ... set all required variables
    ```

5. **Deploy:**
    ```bash
    railway up
    ```

### Supabase Deployment

Supabase doesn't directly host Go applications, but you can:

1. **Deploy to Fly.io/Render/Railway** and connect to Supabase services
2. **Use Supabase Edge Functions** for serverless functions (requires Deno/TypeScript)
3. **Deploy to any Docker-compatible platform** and use Supabase for database/auth

### Docker Deployment (Generic)

1. **Build Image:**

    ```bash
    docker build -t go-fiber-backend .
    ```

2. **Run Container:**

    ```bash
    docker run -d -p 3000:3000 \
      -e SUPABASE_URL=xxx \
      -e SUPABASE_ANON_KEY=xxx \
      --env-file .env \
      go-fiber-backend
    ```

3. **Docker Compose** (optional):
   Create `docker-compose.yml` for local development with multiple services.

## Development

### Running Locally (Recommended for Development)

The easiest way to develop is to run the application directly without Docker:

```bash
# Run the server
go run ./cmd/server

# Or using Make
make run-local
```

The server will start on `http://localhost:3000` and automatically reload when you make code changes (if using `air`).

### Hot Reload (Optional)

For automatic reloading during development:

1. **Install air:**

    ```bash
    go install github.com/cosmtrek/air@latest
    ```

2. **Run with air:**

    ```bash
    air
    ```

    This will automatically rebuild and restart the server when you change any `.go` files.

### Running with Docker

If you prefer to run in a Docker container:

```bash
# Build the Docker image
make build

# Create and run container
make create

# Or build and run in one command
make build-and-run

# Stop the container
make stop

# Start existing container
make start
```

### Using Make Commands

```bash
make help              # Show all available commands
make run-local         # Run the app locally (without Docker)
make requirements      # Update go.mod and go.sum
make build             # Build Docker image
make create            # Create/update Docker container
make build-and-run     # Build image and create container
make update            # Update container with new image
make up                # Run in Docker container (interactive)
make stop              # Stop container
make start             # Start existing container
```

### Running Tests

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Generate coverage report
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out
```

### Debugging

-   Check logs in terminal output
-   Use `/health` endpoint to verify server is running
-   Test endpoints using the demo page at `/demo`
-   Use `curl` or Postman for API testing

## Testing

See [README_TESTING.md](README_TESTING.md) for detailed testing documentation.

### Quick Test Commands

```bash
# Run all tests
go test ./... -v

# Run with coverage
go test ./... -cover

# Run load tests
./loadtest.sh
```

### Test Coverage

Aim for **70%+ coverage** across all packages.

## Production Considerations

### Security

-   ✅ Set `GO_ENV=production`
-   ✅ Configure `ALLOWED_ORIGINS` with your frontend domain(s)
-   ✅ Use strong `JWT_SECRET` (generate with `openssl rand -base64 32`)
-   ✅ Enable `ENABLE_TRUSTED_PROXY_CHECK` if behind proxy
-   ✅ Set `TRUSTED_PROXIES` with your proxy IP ranges
-   ✅ Use HTTPS in production
-   ✅ Keep dependencies updated

### Performance

-   ✅ Rate limiting prevents abuse
-   ✅ Redis caching reduces database load
-   ✅ Go Fiber provides excellent performance
-   ✅ Monitor response times and error rates

### Monitoring

-   Use health check endpoint: `GET /health`
-   Monitor logs for errors
-   Set up alerts for high error rates
-   Track rate limit violations

### Environment Variables

In production, set all environment variables through your platform's secrets management:

-   Never commit `.env` file
-   Use platform-specific secret management
-   Rotate secrets regularly

## Troubleshooting

### Common Issues

**Server won't start:**

-   Check if PORT is available
-   Verify all required environment variables are set
-   Check logs for error messages

**Authentication fails:**

-   Verify JWT token is valid
-   Check `JWT_SECRET` matches token issuer
-   For Supabase tokens, ensure `SUPABASE_URL` is correct

**Rate limit errors:**

-   Check `RATE_LIMIT_MAX` value
-   Verify proxy configuration if behind load balancer
-   Consider increasing limit for legitimate use cases

**WebSocket connection fails:**

-   Verify endpoint is `ws://` or `wss://` (not `http://`)
-   Check CORS configuration
-   Ensure WebSocket is enabled in your deployment platform

**Redis cache not working:**

-   Verify `UPSTASH_REDIS_URL` is correct
-   Check `UPSTASH_REDIS_TOKEN` if required
-   Backend continues without cache if Redis is unavailable

### Getting Help

1. Check the demo page at `/demo` for interactive guides
2. Review this README for common solutions
3. Check [README_TESTING.md](README_TESTING.md) for testing issues
4. Open an issue on GitHub

## Contributing

Contributions are welcome! Please:

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Add tests for new features
5. Ensure all tests pass
6. Submit a pull request

### Code Style

-   Follow Go conventions
-   Run `go fmt` before committing
-   Add comments for public functions
-   Write tests for new features

## License

See [LICENSE](LICENSE) file for details.

---

**Need help?** Visit the interactive demo page at `http://localhost:3000/demo` for complete guides, code examples, and testing tools.

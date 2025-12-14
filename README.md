# Micro Backend Service - Leaderboard API

A microservice for handling game leaderboards with support for different types of score tracking, regular resets, and flexible querying.

## API Versioning

As of version 2.0.0, this service uses a versioned API structure. All endpoints are now prefixed with `/api/v2/`.

- Regular user endpoints: `/api/v2/leaderboards/*`
- Admin endpoints: `/api/v2/admin/leaderboards/*`

For more details, see [API Versioning Documentation](docs/api-versioning.md).

## Features

- **Multiple Leaderboard Types**: 
  - **BEST**: Keeps the highest score (e.g., Angry Birds level score)
  - **SET**: Replaces previous score with latest submission
  - **INCR**: Accumulates scores over time (e.g., total kills or damage in PUBG)

- **Automatic Reset Cycles**:
  - Daily, weekly, monthly resets using cron expressions
  - Preserves historical data for analytics

- **Flexible Querying**:
  - Get top N roles globally
  - Get role-centric rankings with nearby competitors
  - Filter and paginate results

- **Dual API Support**:
  - REST API for simple integration
  - gRPC API for high-performance clients

## Architecture

This service follows a clean architecture approach with clear separation of concerns:

```
micro-backend-service/
├── apigrpc/            # gRPC API definitions
├── cmd/                # Executable entry points
│   └── server/         # Main server executable
├── configs/            # Configuration components
├── internal/           # Core business logic
│   ├── domain/         # Domain models 
│   ├── service/        # Business logic services
│   └── repository/     # Data access layer
├── middleware/         # HTTP/gRPC middleware components
├── pkg/                # Shared packages
├── proto/              # Protocol buffer definitions
├── server/             # Server implementations
│   ├── http/           # HTTP server
│   └── grpc/           # gRPC server
└── utils/              # Utilities and helpers
```

## Getting Started

### Prerequisites

- Go 1.21+
- PostgreSQL 12+
- Redis 6+ (for caching, distributed locks, and cluster mode if needed)

### Environment Variables

Create a `.env` file with the following variables:

```
# Application Environment
GO_ENV=development

# Database Configuration
DATABASE_URI_DEV=postgres://user:password@localhost:5432/leaderboard_dev
DATABASE_URI_PROD=postgres://user:password@localhost:5432/leaderboard_prod
DB_MAX_IDLE_CONNS=10
DB_MAX_OPEN_CONNS=100
DB_CONN_MAX_LIFETIME_MINUTES=60
DB_CONN_MAX_IDLE_TIME_MINUTES=30
DB_USE_UTC=true
DB_PREPARE_STMT=true
DB_SKIP_DEFAULT_TX=true

# Redis Configuration
REDIS_HOST=localhost:6379
REDIS_PORT=6379
REDIS_PASSWORD=
REDIS_USERNAME=
REDIS_DB=0
REDIS_MODE=standalone # Options: standalone, sentinel, cluster
REDIS_SERVICE_NAME= # For sentinel mode
REDIS_POLL_INTERVAL=10
REDIS_READ_TIMEOUT=3
REDIS_WRITE_TIMEOUT=3
REDIS_CONNECT_TIMEOUT=5
REDIS_MAX_RETRIES=3
REDIS_MIN_RETRY_BACKOFF=8
REDIS_MAX_RETRY_BACKOFF=512

# Redis Cluster Configuration (if using cluster mode)
REDIS_CLUSTER_ADDRS=localhost:7000,localhost:7001,localhost:7002,localhost:7003,localhost:7004,localhost:7005

# Server Configuration
HTTP_PORT=8080
GRPC_PORT=50051

# Snowflake ID Generation
MACHINE_NODE=1

# Logging
LOG_LEVEL=info # Options: silent, error, warn, info, debug

# JWT Configuration
JWT_SECRET=your_jwt_secret_key
JWT_EXPIRY_HOURS=24
```

### Running the Service

```bash
# Build the service
go build -o leaderboard-service ./cmd/server

# Run the service
./leaderboard-service
```

### Docker Support

```bash
# Build the image
docker build -t leaderboard-service .

# Run the container
docker run -p 8080:8080 -p 50051:50051 --env-file .env leaderboard-service
```

## Docker Deployment

The application can be easily deployed using Docker and Docker Compose, which will set up the entire stack including PostgreSQL and Redis.

### Prerequisites for Docker Deployment

- Docker Engine 20.10.0+
- Docker Compose v2.0.0+

### Running with Docker Compose

1. Build and start the entire stack:

```bash
docker-compose up -d
```

This will:
- Build the microservice container
- Start a PostgreSQL database with initialization scripts
- Start a Redis instance
- Configure all services with appropriate environment variables

2. Check service status:

```bash
docker-compose ps
```

3. View logs:

```bash
# All services
docker-compose logs -f

# Specific service
docker-compose logs -f micro-backend
```

4. Stop all services:

```bash
docker-compose down
```

To remove volumes (data will be lost):

```bash
docker-compose down -v
```

### Docker Configuration

The Docker setup includes:

- **Multi-stage build** in `Dockerfile` to create optimized container images
- **Initialization scripts** for PostgreSQL to set up required extensions and schemas
- **Volume persistence** for database and Redis data
- **Container health checks** to ensure service dependencies are properly managed
- **Network isolation** with a dedicated bridge network

### Production Considerations

For production deployment, modify the `docker-compose.yml` to:

1. Use a production-ready `docker-compose.prod.yml` that:
   - Sets `GO_ENV=production`
   - Uses secure passwords via environment variables or Docker secrets
   - Properly scales services based on workload
   - Implements additional security measures

2. Set up proper logging:
   - Add a logging driver (e.g., Fluentd, Logstash)
   - Adjust log levels to reduce verbosity

3. Configure SSL/TLS:
   - Use a reverse proxy (Nginx, Traefik)
   - Set up proper certificate management

### Production Deployment

The project includes a production deployment setup with:

1. **Production Docker Compose file** (`docker-compose.prod.yml`) with:
   - Environment variable configuration
   - Resource constraints
   - Scaling options
   - Volume management for persistence

2. **Production environment template** (`configs/env.prod.example`) to configure:
   - Docker registry details
   - Database credentials
   - Redis settings
   - Volume mapping for data persistence

3. **Deployment script** (`deploy.sh`) that simplifies the deployment process:

```bash
# Make the script executable
chmod +x deploy.sh

# Show usage options
./deploy.sh --help

# Deploy with custom environment file and build images
./deploy.sh --env configs/env.prod --build

# Deploy and scale the service to 3 replicas
./deploy.sh --scale 3
```

Follow these steps to deploy to production:

1. Copy the environment template and configure it:
```bash
cp configs/env.prod.example configs/env.prod
# Edit configs/env.prod with your production values
```

2. Run the deployment script:
```bash
./deploy.sh
```

3. Monitor the deployment:
```bash
docker-compose -f docker-compose.yml -f docker-compose.prod.yml ps
docker-compose -f docker-compose.yml -f docker-compose.prod.yml logs -f
```

## Service Initialization

During startup, the service:

1. Initializes the Snowflake node for ID generation
2. Connects to the Redis cluster 
3. Establishes a database connection
4. Runs database migrations
5. Starts the leaderboard reset scheduler
6. Starts HTTP and gRPC servers with graceful shutdown capabilities

### Graceful Shutdown

The service implements a comprehensive graceful shutdown mechanism to ensure clean termination:

1. Both HTTP and gRPC servers support graceful shutdown:
   - HTTP server uses `http.Server.Shutdown()` with a timeout context
   - gRPC server uses `grpc.Server.GracefulStop()`

2. The shutdown process:
   - Catches SIGINT and SIGTERM signals to trigger shutdown
   - Creates a context with timeout to limit shutdown duration
   - Stops accepting new connections
   - Completes in-flight requests before terminating
   - Closes all database and Redis connections
   - Cancels background tasks via context cancellation

3. Server implementation:
   - Both servers expose a uniform API through `Start()` and `Shutdown()` methods
   - Logging is consistent across server implementations
   - Error handling captures server failures for proper diagnosis

This implementation ensures that the service can be safely restarted or terminated without losing data or interrupting client requests in progress.

## API Documentation

### REST API Endpoints

#### Leaderboard Management

- `POST /api/leaderboards` - Create a new leaderboard
- `PUT /api/leaderboards/{id}` - Update a leaderboard
- `DELETE /api/leaderboards/{id}` - Delete a leaderboard
- `GET /api/leaderboards/{id}` - Get a leaderboard
- `GET /api/leaderboards` - List all leaderboards

#### Role Scores

- `POST /api/leaderboards/{id}/scores` - Submit a new score
- `GET /api/leaderboards/{id}/rankings/top/{limit}` - Get top rankings
- `GET /api/leaderboards/{id}/rankings/role/{roleId}` - Get role-centric rankings

#### Advanced Features

- `GET /api/leaderboards/{id}/rankings/full` - Get all rankings (with pagination)
- `GET /api/leaderboards/{id}/rankings/top10` - Get top 10 roles
- `GET /api/leaderboards/{id}/roles/{roleId}/centered?neighbors=5` - Get role-centered ranking with configurable number of neighbors
- `GET /api/admin/leaderboards/due-reset` - Get leaderboards due for reset 
- `POST /api/admin/leaderboards/reset-all-due` - Reset all leaderboards that are due

#### Admin Actions

- `POST /api/leaderboards/{id}/reset` - Reset a leaderboard

### gRPC API

The gRPC API provides the same functionality with better performance for game clients. See `proto/leaderboard/leaderboard.proto` for details.

#### Advanced Leaderboard gRPC API

The advanced leaderboard features are available in the REST API immediately, but to use them in the gRPC API, additional steps are needed:

1. The protobuf definitions for these features have been added to `proto/leaderboard/leaderboard.proto`
2. Before using these features, you must generate the Go code for these protobuf definitions:

```bash
# Install protoc if not already installed
# On macOS:
brew install protobuf
# On Linux:
apt-get install protobuf-compiler

# Install Go protobuf code generators
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate the code from the protobuf definitions
protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative proto/leaderboard/leaderboard.proto

# Uncomment the new gRPC methods in server/grpc/leaderboard.go
```

3. After generating the code, uncomment the advanced leaderboard feature methods in `server/grpc/leaderboard.go`

The following advanced gRPC methods will be available after these steps:

- `GetFullLeaderboardRankings`: Get complete rankings with customizable limit 
- `GetTopTenRankings`: Get optimized top 10 rankings
- `GetPlayerCenteredRanking`: Get player's position with neighboring players
- `GetLeaderboardsDueReset`: List leaderboards due for reset
- `ResetAllDueLeaderboards`: Reset all leaderboards due for reset

## Advanced Features

### Database Views and Functions

The leaderboard service leverages PostgreSQL views and functions to provide efficient and optimized access to leaderboard data:

#### Ranking Views

1. **Complete Rankings (vw_leaderboard_rankings)**
   - Provides complete ranking information for all roles in a leaderboard
   - Pre-calculated rankings based on sort order
   - Includes all role data, scores, metadata, and timestamps

2. **Top 10 Roles (vw_leaderboard_top10)**
   - Pre-filtered view showing only the top 10 roles for each leaderboard
   - Perfect for leaderboard displays and dashboards
   - Extremely fast access to most important leaderboard data

#### Reset Management

3. **Due Resets (vw_leaderboards_due_reset)**
   - Shows leaderboards that are due for reset based on their reset schedule
   - Used by the reset scheduler to determine which leaderboards to reset

#### Role-Centered Ranking

4. **Role-Centered Ranking Function (f_get_role_centered_ranking)**
   - Gets a role's rank along with configurable "neighbors" above and below
   - Perfect for showing a role their position in context with nearby competitors
   - Includes a "distance" metric showing how far each entry is from the role
   - Accessed via: `GET /api/leaderboards/{id}/roles/{roleId}/centered?neighbors=5`

### Reset Management

The service includes comprehensive reset management capabilities:

1. **Scheduled Resets**
   - Daily, weekly, and monthly reset cycles
   - Custom cron expressions for precise scheduling
   - Automatic detection of leaderboards due for reset

2. **Manual Reset**
   - On-demand reset of individual leaderboards
   - Bulk reset of all leaderboards due for schedule

3. **Reset Monitoring**
   - Tracking of last reset timestamps
   - Preview of upcoming scheduled resets

These features enable game developers to maintain fresh competitive environments while preserving historical data for analytics.

## Authentication

This service uses JWT (JSON Web Token) authentication for securing API endpoints:

### Authentication Flow

1. **JWT Token**: All API requests require a valid JWT token in the `Authorization` header using the Bearer scheme: `Authorization: Bearer <token>`.

2. **Claims Structure**: 
   - `id`: User identifier 
   - `roles`: Array of user roles (e.g., ["user"], ["admin"], ["user", "admin"])
   - Standard JWT claims (exp, iat, etc.)

3. **Role-Based Authorization**:
   - Regular API endpoints require user authentication
   - Admin endpoints under `/api/admin/*` require admin role
   - You can authenticate as an admin programmatically using the `AuthenticateAdmin` function

### Creating JWT Tokens (for development)

For development purposes, you can generate a token with the required claims. In production, this would typically be handled by an authentication service.

```bash
# Example using jwt-cli tool for testing
jwt encode --secret your_jwt_secret_key '{"id":"user123","roles":["user"],"exp":1735689600}'

# For admin access
jwt encode --secret your_jwt_secret_key '{"id":"admin123","roles":["admin"],"exp":1735689600}'
```

## License

[MIT License](LICENSE)

# Redis Connectivity

The application supports multiple Redis connection modes:

1. **Standalone**: Single Redis server connection
2. **Sentinel**: Redis with Sentinel for high availability 
3. **Cluster**: Redis Cluster for scalability and partitioning

## Redis Configuration

The Redis connection can be configured using the following environment variables:

- `REDIS_MODE`: Explicitly sets the Redis mode to "standalone", "sentinel", or "cluster"
- `REDIS_HOST`: Comma-separated list of Redis hosts (can be a single host for standalone mode)
- `REDIS_CLUSTER_ADDRS`: Legacy alias for `REDIS_HOST` when using cluster mode
- `REDIS_PORT`: Default port to use when not specified in host addresses
- `REDIS_USERNAME`: Redis username for authentication
- `REDIS_PASSWORD`: Redis password for authentication
- `REDIS_SERVICE_NAME`: Master name for Sentinel mode
- `REDIS_DB`: Database number to use (not applicable for cluster mode)

### Connection Management

The application will attempt to auto-detect the appropriate Redis mode based on the provided configuration. If multiple hosts are specified, it will default to cluster mode unless `REDIS_MODE` is explicitly set.

Connection management functions are available in the `configs` package:

```go
// For regular Redis client that implements RedisClient interface
redisClient, err := config.NewRedisClient()
if err != nil {
    // handle error
}
defer redisClient.Close()

// For direct cluster client access (backward compatibility)
var redisCluster config.RedisClusterConnection
if err := redisCluster.InitializeCluster(); err != nil {
    // handle error
}
defer redisCluster.Close()
```

## Configuration Management

The application uses a centralized approach to manage configuration through environment variables. This provides flexibility for different deployment environments while maintaining a consistent configuration interface across the codebase.

### Environment Variable Loading

Environment variables are loaded from:
- A `.env` file in non-production environments (GO_ENV != "production")
- The system environment in production

All environment variable access is standardized through the utility functions in the `utils` package. This ensures consistent handling of default values and type conversion throughout the application. The direct use of `os.Getenv()` is discouraged in favor of these utilities.

### Available Utility Functions

The `utils` package provides the following functions for working with environment variables:

```go
// Get string value with default
value := util.GetEnv("KEY_NAME", "default_value")

// Get int value with default
intValue := util.GetEnvAsInt("INT_KEY", 42)

// Get int64 value with default
int64Value := util.GetEnvAsInt64("INT64_KEY", 9223372036854775807)

// Get boolean value with default
boolValue := util.GetEnvAsBool("BOOL_KEY", false)

// Get duration (in seconds) with default
duration := util.GetEnvAsDuration("DURATION_KEY", 30*time.Second)

// Get duration (in milliseconds) with default
msDuration := util.GetEnvAsDurationMs("MS_DURATION_KEY", 500*time.Millisecond)
```

The package also supports loading environment variables from a `.env` file using:

```go
util.GodotEnv("VARIABLE_NAME")
```

### Usage Notes

- All new code should use the utility functions from the `utils` package
- Legacy code is being migrated to use these utility functions
- The centralized approach ensures consistent environment variable handling 

## Sample Data

The system comes with pre-configured sample data to demonstrate the leaderboard functionality. Sample data is loaded during the database initialization process.

Each leaderboard has 5 sample role entries with various scores and metadata. You can use these samples to test the API endpoints without having to create data from scratch.

## Database Schema

The database uses the following main tables:

- **leaderboards**: Stores leaderboard configuration and metadata
- **leaderboard_entries**: Stores role scores for each leaderboard

And several views:

1. **leaderboard_rankings**: Shows all roles with their calculated ranks based on the leaderboard's sort order
2. **leaderboard_top10**: Shows top 10 roles for each leaderboard
3. **leaderboards_due_reset**: Shows leaderboards due for reset based on their schedule

Database functions:

- **get_role_centered_ranking(leaderboard_id, role_id, neighbors)**: Returns a role's rank along with nearby competitors (above and below) in the leaderboard

These database objects can be accessed directly via SQL or through the application's API. 

## Advanced Features

### gRPC API for Advanced Leaderboard Features

This service includes advanced leaderboard features available through the gRPC API for high-performance processing:

1. **Installing Protocol Buffers Compiler**:

   On macOS (with Homebrew):
   ```
   brew install protobuf
   ```
   
   On Linux (Ubuntu/Debian):
   ```
   apt-get install -y protobuf-compiler
   ```

2. **Installing Go Plugins for Protocol Buffers**:
   ```
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

3. **Generating Go Code from Proto Definitions**:
   ```
   protoc --go_out=. --go-grpc_out=. proto/leaderboard/leaderboard.proto
   ```

4. **Uncomment New gRPC Methods**:
   After generating the code, uncomment the new gRPC methods in `server/grpc/leaderboard.go` to enable these features:

   - `GetFullLeaderboardRankings`: Get all rankings with pagination
   - `GetTopTenRankings`: Get top 10 rankings efficiently
   - `GetRoleCenteredRanking`: Get role's position with neighboring roles
   - `GetLeaderboardsDueReset`: Identify leaderboards due for reset
   - `ResetAllDueLeaderboards`: Reset all leaderboards due for reset 
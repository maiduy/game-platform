# Game Platform Architecture

## Overview

This document describes the architecture of our game platform backend system, designed following industry-standard patterns from platforms like PlayFab, Nakama, and Hiro. The platform supports multiple games, enforces separation of concerns, and ensures consistent engineering standards across all services.

## Architecture Principles

### 1. Multi-Tenancy & Game Isolation
- **Game-Specific Configurations**: Each game (X3, 4F, etc.) has its own configuration under `configs/games/{game_id}/`
- **Common Shared Configurations**: Reusable configs in `configs/games/common/`
- **Platform-Level Configurations**: Core platform settings in `configs/platform/`

### 2. Clean Architecture (Hexagonal Architecture)
Each service follows clean architecture principles with clear separation:
- **Domain Layer**: Core business logic and entities (pure Go, no external dependencies)
- **Application Layer**: Use cases and service interfaces
- **Repository Layer**: Data access implementations
- **API Layer**: HTTP/gRPC handlers (adapters)

### 3. Service-Oriented Architecture
Independent microservices that communicate via:
- **gRPC**: For internal service-to-service communication
- **HTTP/REST**: For external client-facing APIs
- **Message Queue (NATS)**: For async event-driven communication

## Directory Structure

```
/
├── cmd/                                # Service entry points (1 service = 1 binary)
│   ├── gateway/main.go                 # API Gateway
│   ├── masterconfig/main.go            # Master Game Data Config Service
│   ├── leaderboard/main.go             # Leaderboard Service
│   ├── economy/main.go                 # Economy & IAP Service
│   ├── ability/main.go                 # Ability & Combat Engine Service
│   └── match/main.go                   # Matchmaking Service
│
├── internal/                           # Private platform code
│   ├── gateway/                        # API Gateway service
│   │   ├── api/                        # HTTP handlers & routers
│   │   │   ├── http/
│   │   │   │   ├── handler.go
│   │   │   │   ├── route.go
│   │   │   │   └── middleware.go
│   │   │   └── grpc/                   # gRPC client connections
│   │   │       └── clients.go
│   │   ├── application/                # Gateway composition logic
│   │   │   └── service.go
│   │   └── middleware/                 # Gateway-specific middleware
│   │       ├── auth.go
│   │       ├── rate_limit.go
│   │       └── logging.go
│   │
│   ├── masterconfig/                   # Master config service
│   │   ├── api/
│   │   │   ├── http/                   # REST API handlers
│   │   │   │   ├── handler.go
│   │   │   │   └── route.go
│   │   │   └── grpc/                   # gRPC server
│   │   │       └── server.go
│   │   ├── application/                # Business logic
│   │   │   ├── service.go
│   │   │   └── repository.go           # Repository interface
│   │   ├── domain/                     # Domain models
│   │   │   ├── character.go
│   │   │   ├── skill.go
│   │   │   └── item.go
│   │   ├── repository/                 # Data access implementation
│   │   │   └── mongo_repository.go
│   │   └── resolver/                   # Config resolution logic
│   │       └── resolver.go
│   │
│   ├── leaderboard/                    # Leaderboard service
│   │   ├── api/
│   │   │   ├── http/
│   │   │   └── grpc/
│   │   ├── application/
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── domain/
│   │   │   ├── leaderboard.go
│   │   │   └── entry.go
│   │   ├── repository/
│   │   │   └── postgres_repository.go
│   │   └── scheduler/                  # Reset scheduler
│   │       └── cron.go
│   │
│   ├── economy/                        # Economy service
│   │   ├── api/
│   │   │   ├── http/
│   │   │   └── grpc/
│   │   ├── application/
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── domain/
│   │   │   ├── currency.go
│   │   │   ├── inventory.go
│   │   │   ├── transaction.go
│   │   │   └── iap.go
│   │   ├── repository/
│   │   │   └── postgres_repository.go
│   │   └── antifraud/                  # Anti-fraud detection
│   │       └── detector.go
│   │
│   ├── ability/                        # Ability / Combat Engine
│   │   ├── api/
│   │   │   ├── http/
│   │   │   └── grpc/
│   │   ├── application/
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── domain/
│   │   │   ├── skill/
│   │   │   │   ├── skill.go
│   │   │   │   └── types.go
│   │   │   ├── effect/
│   │   │   │   ├── effect.go
│   │   │   │   └── processor.go
│   │   │   └── modifier/
│   │   │       └── modifier.go
│   │   ├── resolver/                   # Skill formula resolver
│   │   │   └── formula.go
│   │   └── simulator/                  # Combat simulation
│   │       └── engine.go
│   │
│   ├── match/                          # Matchmaking service
│   │   ├── api/
│   │   │   ├── http/
│   │   │   └── grpc/
│   │   ├── application/
│   │   │   ├── service.go
│   │   │   └── repository.go
│   │   ├── domain/
│   │   │   ├── match.go
│   │   │   ├── queue.go
│   │   │   └── room.go
│   │   └── repository/
│   │       └── redis_repository.go
│   │
│   └── platform/                       # SHARED PLATFORM INFRASTRUCTURE
│       ├── app/                        # Application initialization
│       │   └── app.go
│       │
│       ├── auth/                       # Authentication & Authorization
│       │   ├── token.go
│       │   ├── validation.go
│       │   └── middleware.go
│       │
│       ├── config/                     # Configuration management
│       │   ├── loader.go
│       │   ├── watcher.go
│       │   └── dotenv.go
│       │
│       ├── logger/                     # Structured logging
│       │   └── logger.go
│       │
│       ├── metrics/                    # Metrics & observability
│       │   └── prometheus.go
│       │
│       ├── pubsub/                     # Pub/Sub messaging
│       │   ├── publisher.go
│       │   └── subscriber.go
│       │
│       ├── security/                   # Security utilities
│       │   ├── jwt.go
│       │   ├── bcrypt.go
│       │   ├── rate_limiter.go
│       │   ├── request_validator.go
│       │   └── whitelist.go
│       │
│       ├── errors/                     # Error handling
│       │   ├── code.go
│       │   └── errors.go
│       │
│       ├── http/                       # HTTP utilities
│       │   └── response.go
│       │
│       ├── storage/                    # Storage connection pools
│       │   ├── redis/
│       │   │   └── client.go
│       │   ├── postgres/
│       │   │   └── client.go
│       │   ├── mongodb/
│       │   │   └── client.go
│       │   ├── mysql/
│       │   │   └── client.go
│       │   ├── elasticsearch/
│       │   │   └── client.go
│       │   └── nats/
│       │       └── client.go
│       │
│       ├── server/                     # Server implementations
│       │   ├── http.go
│       │   └── grpc.go
│       │
│       └── notification/               # Notification services
│           └── sendgrid.go
│
├── configs/                            # DATA-DRIVEN CONFIGURATION
│   ├── platform/                       # Platform-level configs
│   │   ├── app_config.json
│   │   ├── auth_config.json
│   │   └── rate_limit.yaml
│   │
│   ├── games/                          # Game-specific configs
│   │   ├── common/                     # Shared across all games
│   │   │   ├── character.yaml
│   │   │   ├── leaderboard.yaml
│   │   │   ├── economy.yaml
│   │   │   └── ability.yaml
│   │   │
│   │   ├── X3/                         # Game X3 specific
│   │   │   ├── character.yaml
│   │   │   ├── ability.yaml
│   │   │   └── economy.yaml
│   │   │
│   │   └── 4F/                         # Game 4F specific
│   │       ├── character.yaml
│   │       ├── ability.yaml
│   │       └── economy.yaml
│   │
│   └── liveops/                        # LiveOps configurations
│       ├── events.yaml
│       ├── leaderboard_reset.yaml
│       └── ab_test.yaml
│
├── proto/                              # gRPC Protocol Buffers
│   ├── common/
│   │   └── common.proto                # Shared message types
│   ├── masterconfig/
│   │   └── masterconfig.proto
│   ├── leaderboard/
│   │   └── leaderboard.proto
│   ├── economy/
│   │   └── economy.proto
│   ├── ability/
│   │   └── ability.proto
│   └── match/
│       └── match.proto
│
├── docs/
│   ├── architecture.md                 # This file
│   ├── coding_rules.md                 # Coding standards
│   ├── logging_strategy.md             # Logging guidelines
│   ├── api_response.md                 # API response format
│   └── adr/                            # Architecture Decision Records
│       ├── 0001-clean-architecture.md
│       ├── 0002-multi-game-support.md
│       └── 0003-grpc-vs-rest.md
│
├── docker/
│   ├── Dockerfile.gateway
│   ├── Dockerfile.masterconfig
│   ├── Dockerfile.leaderboard
│   ├── Dockerfile.economy
│   ├── Dockerfile.ability
│   ├── Dockerfile.match
│   ├── docker-compose.yml              # Development
│   └── docker-compose.prod.yml         # Production
│
├── scripts/
│   ├── deploy.sh
│   ├── migrate.sh
│   └── generate_proto.sh
│
├── tools/
│   └── config_validator/               # Config validation tool
│       └── main.go
│
├── go.mod
└── go.sum
```

## Service Descriptions

### 1. Gateway Service
**Purpose**: Single entry point for all client requests, handles routing, auth, rate limiting

**Responsibilities**:
- Route requests to appropriate backend services
- JWT authentication and validation
- Rate limiting per client/IP
- Request/response transformation
- API composition (aggregating multiple service calls)
- CORS handling

**Tech Stack**: HTTP (Gin), gRPC clients

### 2. MasterConfig Service
**Purpose**: Centralized game configuration management (characters, items, skills, etc.)

**Responsibilities**:
- CRUD operations for game configurations
- Version management of configs
- Game-specific config resolution (common + game-specific merging)
- Configuration caching
- Hot-reload support

**Tech Stack**: MongoDB, HTTP/gRPC, Redis cache

### 3. Leaderboard Service
**Purpose**: Manage competitive rankings and leaderboards

**Responsibilities**:
- Multiple leaderboard types (BEST, SET, INCR)
- Score submission and ranking calculation
- Automatic reset cycles (daily, weekly, monthly)
- Historical data preservation
- Top-N and player-centered queries

**Tech Stack**: PostgreSQL (with ranking views), HTTP/gRPC, Cron scheduler

### 4. Economy Service
**Purpose**: Virtual economy management (currencies, inventory, IAP)

**Responsibilities**:
- Currency management (hard/soft currencies)
- Inventory system (items, consumables)
- Transaction processing with ACID guarantees
- In-App Purchase (IAP) verification
- Anti-fraud detection
- Transaction history and auditing

**Tech Stack**: PostgreSQL (transactional), HTTP/gRPC, Redis cache

### 5. Ability Service
**Purpose**: Combat and ability system engine

**Responsibilities**:
- Skill definition and management
- Effect processing (buffs, debuffs, damage)
- Modifier calculation (stat changes)
- Formula resolution (dynamic calculations)
- Combat simulation
- Ability validation

**Tech Stack**: MongoDB (skill configs), HTTP/gRPC, In-memory engine

### 6. Match Service
**Purpose**: Matchmaking and lobby management

**Responsibilities**:
- Player queue management
- Skill-based matchmaking
- Room creation and management
- Real-time player state
- Match history

**Tech Stack**: Redis (queue + state), PostgreSQL (history), HTTP/gRPC, WebSocket

## Data Flow Patterns

### 1. Client → Gateway → Service → Database
```
Client Request → Gateway (Auth + Rate Limit) → Service (Business Logic) → Database → Response
```

### 2. Service-to-Service Communication
```
Service A → gRPC → Service B
Service A → NATS Pub → Event Bus → NATS Sub → Service B
```

### 3. Configuration Resolution
```
Request → MasterConfig Service → Check Game ID → Merge Common + Game-Specific Configs → Return Resolved Config
```

## Technology Stack

### Backend Services
- **Language**: Go 1.21+
- **HTTP Framework**: Gin
- **gRPC**: Protocol Buffers + gRPC
- **Configuration**: YAML/JSON + hot-reload

### Data Storage
- **PostgreSQL**: Transactional data (economy, leaderboard, match history)
- **MongoDB**: Document data (game configs, player profiles)
- **Redis**: Caching, sessions, real-time state, queues
- **Elasticsearch**: Logs, analytics (optional)

### Infrastructure
- **Message Queue**: NATS
- **Metrics**: Prometheus + Grafana
- **Logging**: Structured JSON logs (Logrus/Zap)
- **Tracing**: OpenTelemetry (optional)
- **Service Discovery**: Consul (optional)

### DevOps
- **Containerization**: Docker
- **Orchestration**: Docker Compose (dev), Kubernetes (prod)
- **CI/CD**: GitHub Actions / GitLab CI

## Configuration Management

### Multi-Game Configuration Strategy
1. **Common Configurations**: Shared across all games in `configs/games/common/`
2. **Game-Specific Overrides**: Each game can override in `configs/games/{game_id}/`
3. **Resolution Order**: Game-specific > Common > Defaults
4. **Hot-Reload**: Services watch config files and reload without restart

### Example: Character Configuration
```
configs/games/common/character.yaml       # Base character attributes
configs/games/X3/character.yaml           # X3-specific overrides (e.g., different stats)
configs/games/4F/character.yaml           # 4F-specific overrides
```

MasterConfig service resolves: `merge(common, game_specific_override)`

## API Standards

### REST API Conventions
- **Base Path**: `/api/v1/{service}/{resource}`
- **Methods**: GET, POST, PUT, DELETE, PATCH
- **Response Format**: JSON with standardized structure
```json
{
  "success": true,
  "data": { ... },
  "error": null,
  "meta": {
    "request_id": "uuid",
    "timestamp": 1234567890
  }
}
```

### gRPC API Conventions
- **Proto Package**: `{service}.v1`
- **Service Name**: `{Service}Service`
- **Method Naming**: `{Verb}{Resource}` (e.g., `CreateCharacter`, `GetLeaderboard`)

## Security

### Authentication
- **JWT Tokens**: Issued by auth service, validated at gateway
- **Claims**: user_id, game_id, roles, permissions
- **Expiry**: Configurable (default 24h), refresh tokens supported

### Authorization
- **Role-Based Access Control (RBAC)**: admin, player, service
- **Permission Checks**: At gateway and service level
- **API Keys**: For service-to-service communication

### Rate Limiting
- **Per User**: Configurable limits (e.g., 100 req/min)
- **Per IP**: DDoS protection (e.g., 1000 req/min)
- **Per Endpoint**: Sensitive endpoints have stricter limits

## Monitoring & Observability

### Metrics (Prometheus)
- Request count, latency, error rate per endpoint
- Database query performance
- Cache hit/miss ratio
- Queue depth and processing time

### Logging
- **Structured JSON Logs**: Consistent format across services
- **Log Levels**: DEBUG, INFO, WARN, ERROR
- **Correlation IDs**: Track requests across services
- **Log Aggregation**: ELK Stack or Loki

### Tracing
- **OpenTelemetry**: Distributed tracing across services
- **Span Tagging**: Service, method, user_id, game_id

## Deployment

### Development
```bash
docker-compose up  # Starts all services + dependencies
```

### Production
```bash
# Build images
./scripts/build.sh

# Deploy to Kubernetes
kubectl apply -f k8s/
```

### Environment Variables
Each service configures via environment variables:
- `SERVICE_PORT`: HTTP port
- `GRPC_PORT`: gRPC port
- `DATABASE_URL`: Database connection string
- `REDIS_URL`: Redis connection string
- `LOG_LEVEL`: Logging level
- `JWT_SECRET`: JWT signing secret

## Scalability Considerations

### Horizontal Scaling
- **Stateless Services**: All services are stateless, can scale horizontally
- **Load Balancing**: Nginx/HAProxy for HTTP, gRPC load balancing
- **Database Replication**: Read replicas for read-heavy workloads

### Caching Strategy
- **Redis Caching**: Frequently accessed data (configs, player profiles)
- **Cache Invalidation**: TTL + event-driven invalidation
- **Cache-Aside Pattern**: Read-through caching

### Database Optimization
- **Indexing**: Proper indexes on query fields
- **Connection Pooling**: Limit concurrent connections
- **Query Optimization**: Use EXPLAIN for slow queries

## Testing Strategy

### Unit Tests
- Test business logic in isolation
- Mock external dependencies
- Target: >80% code coverage

### Integration Tests
- Test service APIs end-to-end
- Use test databases/containers
- Validate data persistence

### Load Tests
- Simulate production traffic
- Identify bottlenecks
- Tools: k6, Gatling

## Migration Path

### Phase 1: Foundation (Week 1-2)
1. Set up platform infrastructure (`internal/platform`)
2. Implement Gateway service
3. Migrate MasterConfig service to new structure
4. Create proto definitions

### Phase 2: Core Services (Week 3-5)
1. Implement Leaderboard service
2. Implement Economy service
3. Implement Ability service
4. Implement Match service

### Phase 3: Integration & Testing (Week 6-7)
1. Service-to-service integration
2. End-to-end testing
3. Performance optimization
4. Documentation

### Phase 4: Production Ready (Week 8)
1. Security hardening
2. Monitoring & alerting setup
3. Deployment automation
4. Production rollout

## References

- [Clean Architecture - Robert C. Martin](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [PlayFab Architecture](https://docs.microsoft.com/en-us/gaming/playfab/)
- [Nakama Architecture](https://heroiclabs.com/docs/)
- [12-Factor App](https://12factor.net/)

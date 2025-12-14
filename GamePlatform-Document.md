/
├── cmd/                                # Entrypoints (1 service = 1 binary)
│   ├── gateway/
│   │   └── main.go
│   ├── masterconfig/
│   │   └── main.go
│   ├── leaderboard/
│   │   └── main.go
│   ├── economy/
│   │   └── main.go
│   ├── ability/
│   │   └── main.go
│   └── match/
│       └── main.go
│
├── internal/                           # PRIVATE platform code
│   ├── gateway/
│   │   ├── api/
│   │   ├── middleware/
│   │   └── application/
│   │
│   ├── masterconfig/
│   │   ├── api/
│   │   ├── application/
│   │   ├── domain/
│   │   ├── repository/
│   │   └── resolver/
│   │
│   ├── leaderboard/
│   │   ├── api/
│   │   ├── application/
│   │   ├── domain/
│   │   ├── repository/
│   │   └── scheduler/
│   │
│   ├── economy/
│   │   ├── api/
│   │   ├── application/
│   │   ├── domain/
│   │   ├── repository/
│   │   └── antifraud/
│   │
│   ├── ability/                       # Ability / Combat Engine
│   │   ├── api/
│   │   ├── application/
│   │   ├── domain/
│   │   │   ├── skill/
│   │   │   ├── effect/
│   │   │   └── modifier/
│   │   ├── resolver/
│   │   └── simulator/
│   │
│   ├── match/
│   │   ├── api/
│   │   ├── application/
│   │   ├── domain/
│   │   └── repository/
│   │
│   └── platform/                      # SHARED PLATFORM INFRA
│       ├── auth/
│       │   ├── token.go
│       │   └── validator.go
│       │
│       ├── config/
│       │   ├── loader.go
│       │   └── watcher.go
│       │
│       ├── logger/
│       │   └── logger.go
│       │
│       ├── metrics/
│       │   └── prometheus.go
│       │
│       ├── pubsub/
│       │   ├── publisher.go
│       │   └── subscriber.go
│       │
│       ├── security/
│       │   ├── rate_limiter.go
│       │   ├── request_validator.go
│       │   └── whitelist.go
│       │
│       ├── errors/
│       │   ├── code.go
│       │   └── error.go
│       │
│       ├── storage/                  # CONNECTION POOLS
│       │   ├── redis/
│       │   │   └── client.go
│       │   ├── postgres/
│       │   │   └── client.go
│       │   ├── mongodb/
│       │   │   └── client.go
│       │   └── elastic/
│       │       └── client.go
│       │
│       ├── clock/
│       │   └── clock.go
│       │
│       └── id/
│           └── snowflake.go
│
├── configs/                           # DATA-DRIVEN CORE
│   ├── platform/
│   │   ├── app_config.json
│   │   ├── auth_config.json
│   │   └── rate_limit.yaml
│   │
│   ├── games/
│   │   ├── common/
│   │   │   ├── character.yaml
│   │   │   ├── leaderboard.yaml
│   │   │   ├── economy.yaml
│   │   │   └── ability.yaml
│   │   │
│   │   ├── X3/
│   │   │   ├── character.yaml
│   │   │   └── ability.yaml
│   │   │
│   │   └── 4F/
│   │       ├── character.yaml
│   │       └── ability.yaml
│   │
│   └── liveops/
│       ├── events.yaml
│       ├── leaderboard_reset.yaml
│       └── ab_test.yaml
│
├── proto/                             # gRPC contracts
│   ├── leaderboard/
│   │   └── leaderboard.proto
│   ├── economy/
│   └── ability/
│
├── docs/
│   ├── architecture.md
│   ├── coding_rules.md
│   ├── logging_strategy.md
│   ├── api_response.md
│   └── adr/
│       └── 0001-clean-architecture.md
│
├── docker/
│   ├── Dockerfile
│   ├── docker-compose.yml
│   └── docker-compose.prod.yml
│
├── scripts/
│   ├── deploy.sh
│   └── migrate.sh
│
├── tools/
│   └── config_validator/
│
├── go.mod
└── go.sum

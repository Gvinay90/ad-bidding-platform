# ad-bidding-platform

## Project structure

```
.
├── README.md
├── api-gateway
│   ├── client
│   ├── handler
│   └── router
├── config
├── deploy
│   └── localstack-init
├── go.mod
├── internal
│   ├── analytics
│   │   ├── domain
│   │   ├── events
│   │   ├── handler
│   │   ├── repository
│   │   └── service
│   ├── bidder
│   │   ├── cache
│   │   ├── domain
│   │   ├── events
│   │   ├── handler
│   │   └── service
│   ├── campaign
│   │   ├── domain
│   │   ├── events
│   │   ├── handler
│   │   ├── repository
│   │   └── service
│   └── platform
│       ├── awsx
│       ├── config
│       ├── db
│       ├── logx
│       └── redisx
├── scripts
└── services
    ├── analytics
    ├── bidder
    └── campaign
```

## Generate protobuf

From the repository root:

```bash
chmod +x scripts/gen-proto.sh
./scripts/gen-proto.sh
```

This runs `protoc` on `proto/campaign`, `proto/bidder`, and `proto/analytics`. You need [`protoc`](https://grpc.io/docs/protoc-installation/) plus `protoc-gen-go` and `protoc-gen-go-grpc` on your `PATH` (for example `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest` and `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest`).

## Local stack

From `deploy/`:

```bash
cd deploy
docker compose up -d
```

### Verify setup

From `deploy/` (same shell as `docker compose up`). If you changed `docker-compose.yml`, recreate containers so healthchecks apply: `docker compose up -d --force-recreate`.

- `docker compose ps` shows all four services as `(healthy)` once checks have passed (give MySQL a short first-boot window on a fresh volume).
- `docker compose exec redis redis-cli ping` prints `PONG` (uses the Redis container; no host `redis-cli` needed).
- `docker compose exec localstack awslocal sns list-topics` includes `campaign-events` (no host `awslocal` needed).

If you prefer host CLIs instead, install Redis client tools and [awscli-local](https://github.com/localstack/awscli-local), then use `redis-cli -h 127.0.0.1 ping` and `awslocal sns list-topics` against the published ports.

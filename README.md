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

### Troubleshooting: no bids / empty Redis after creating a campaign

Campaign changes reach the bidder through **SNS → SQS → bidder consumer**. If **SNS has no subscriptions** to `bidder-cache` / `analytics-in`, messages never arrive and Redis stays empty (`SMEMBERS idx:status:active` is empty).

Check subscriptions (inside LocalStack):

```bash
docker exec deploy-localstack-1 awslocal sns list-subscriptions-by-topic \
  --topic-arn arn:aws:sns:us-east-1:000000000000:campaign-events
```

If the list is empty, wire subscriptions from the repo root:

```bash
./scripts/wire-localstack-subscriptions.sh
```

Then **create or update a campaign again** so a new `CampaignChanged` event flows through. Alternatively recreate LocalStack so `deploy/localstack-init/bootstrap.sh` runs on a fresh volume (`docker compose down -v && docker compose up -d`).

## How to use this service

The diagrams below render on GitHub and in editors that support [Mermaid](https://mermaid.js.org/).

### End-to-end usage (local dev)

```mermaid
flowchart TD
  A[Install Go, Docker, protoc, grpcurl] --> B[cd deploy && docker compose up -d]
  B --> C{Stacks healthy?}
  C -->|Redis / Postgres / LocalStack| D[Optional: ./scripts/gen-proto.sh after proto changes]
  D --> E[From repo root: config/local.yaml]
  E --> F[Terminal 1: go run ./services/campaign]
  E --> G[Terminal 2: go run ./services/bidder]
  E --> H[Terminal 3: go run ./services/analytics]
  F --> I[Campaign CRUD + SNS CampaignChanged]
  G --> J[Redis from bidder-cache SQS + EvaluateBid + SNS BidEvent]
  H --> K[bid_events from analytics-in SQS + GetCampaignStats gRPC]
  I --> L{Using gateway?}
  J --> L
  K --> L
  L -->|Yes| M[Terminal 4: go run ./api-gateway]
  L -->|No| N[grpcurl :50051 / :50052 / :50053]
  M --> O[HTTP: POST /campaigns, POST /bid, GET /stats]
  N --> P[gRPC: CreateCampaign, EvaluateBid, GetCampaignStats]
```

### Request and event paths

```mermaid
flowchart LR
  subgraph client
    U[Client / curl / grpcurl]
  end
  subgraph gateway_optional["API gateway :8080 optional"]
    GW[REST]
  end
  subgraph grpc["gRPC services"]
    CS[campaign :50051]
    BD[bidder :50052]
    AN[analytics :50053]
  end
  subgraph data
    DB[(RDBMS campaigns)]
    RD[(Redis indexes)]
    BE[(bid_events)]
  end
  subgraph aws_local["LocalStack"]
    SNS[SNS campaign-events]
    Q1[SQS bidder-cache]
    Q2[SQS analytics-in]
  end

  U --> GW
  U --> CS
  U --> BD
  U --> AN
  GW --> CS
  GW --> BD
  GW --> AN
  CS --> DB
  CS --> SNS
  SNS --> Q1
  SNS --> Q2
  Q1 --> BD
  BD --> RD
  BD --> SNS
  Q2 --> AN
  AN --> BE
```

### Happy path checklist

```mermaid
flowchart TD
  S1[1. docker compose up -d in deploy/] --> S2[2. go run ./services/campaign]
  S2 --> S3[3. go run ./services/bidder]
  S3 --> S4[4. go run ./services/analytics]
  S4 --> S5[5. Create campaign via gRPC or gateway]
  S5 --> S6[6. Wait briefly for SQS to refresh bidder Redis]
  S6 --> S7[7. EvaluateBid with same geo / device / category]
  S7 --> S8[8. GetCampaignStats for the winning campaign id]
```

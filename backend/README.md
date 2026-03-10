# FenzVideo Backend

> Go 1.24+ / Kratos v2 / MySQL 8.0 / Redis 7 / MinIO / NATS

## Quick Start

```bash
# 1. Start infrastructure
docker-compose up -d mysql redis minio nats jaeger

# 2. Install protoc plugins (first time)
make init

# 3. Generate proto code
make all

# 4. Run the backend
go run ./cmd/backend/ -conf ./configs/

# 5. (Optional) Seed sample data
GEMINI_KEY=xxx make seed
```

## Makefile Targets

| Command        | Description                                       |
| -------------- | ------------------------------------------------- |
| `make init`    | Install protoc-gen-go, protoc-gen-go-grpc, etc.   |
| `make api`     | Generate Go code from `api/**/*.proto`             |
| `make config`  | Generate Go code from `internal/conf/*.proto`      |
| `make all`     | Run api + config + generate                        |
| `make build`   | Build binary to `bin/backend`                      |
| `make seed`    | Seed database with Gemini-generated sample data    |

## Docker

```bash
# Build image
docker build -t fenzvideo-backend .

# Run container
docker run --rm -p 8000:8000 -p 9000:9000 \
  -v $(pwd)/configs:/data/conf \
  fenzvideo-backend
```

## Architecture

See [docs/backend-architecture.md](../docs/backend-architecture.md) and [docs/full-stack-workflow.md](../docs/full-stack-workflow.md) for details.


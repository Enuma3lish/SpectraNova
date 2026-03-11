# FenzVideo Full-Stack Workflow

## Overview

FenzVideo is a full-stack video streaming platform with a **Go/Kratos** backend and **Vue 3** frontend, orchestrated via **Docker Compose**. This document covers the complete developer workflow — from local setup to production deployment.

---

## Architecture Diagram

```
                           ┌──────────────────────────┐
                           │       User Browser        │
                           └────────────┬─────────────┘
                                        │ HTTP
                                        ▼
                    ┌───────────────────────────────────────┐
                    │         Nginx (Frontend Container)     │
                    │   - Serves Vue 3 SPA (dist/)           │
                    │   - Proxies /api/* → Backend            │
                    │   - Proxies /fenzvideo/* → MinIO        │
                    │   - Gzip, static asset caching          │
                    │   - Docker DNS resolver (127.0.0.11)    │
                    │   Port: 80                              │
                    └─────────────┬─────┬───────────────────┐
                                  │     │ /fenzvideo/*
                                  │     │
                       /api/*     │     │
                                  │     └──────────────────┐
                                  │                        │
                                  ▼                        ▼
                    ┌───────────────────────────────────────┐
                    │       Go/Kratos Backend Container      │
                    │   - HTTP API   (port 8000)             │
                    │   - gRPC API   (port 9000)             │
                    │   - JWT Auth + Admin Guard              │
                    │   - Cache warm-up on boot               │
                    │   - Background workers (view flush)     │
                    │   - WebSocket gateway + presence        │
                    └──┬────┬────┬────┬────┬───────────────┘
                       │    │    │    │    │
          ┌────────────┘    │    │    │    └────────────┐
          ▼                 ▼    │    ▼                  ▼
   ┌────────────┐  ┌───────────┐│ ┌─────────┐   ┌──────────┐
   │  MySQL 8.0 │  │  Redis 7  ││ │  MinIO  │   │   NATS   │
   │  Port 3306 │  │  Port 6379││ │ Port 9100│   │ Port 4222│
   │            │  │           ││ │ (S3 API) │   │ JetStream│
   │ - Users    │  │ - Tag SETs││ │          │   │          │
   │ - Videos   │  │ - Video   ││ │ - Videos │   │ - Events │
   │ - Channels │  │   HASHes  ││ │ - Thumbs │   │ - Notifs │
       │ - Tags     │  │ - Views   ││ │          │   │          │
       │ - Notifs   │  │ - Presence││ │          │   │ - Fanout │
       │ - etc.     │  │   Buffer  ││ │          │   │          │
   └────────────┘  └───────────┘│ └─────────┘   └──────────┘
                                │
                                ▼
                         ┌────────────┐
                         │   Jaeger   │
                         │ Port 16686 │
                         │  (Tracing) │
                         └────────────┘
```

---

## Container Map

| Service   | Image / Build            | Ports               | Purpose                          |
| --------- | ------------------------ | ------------------- | -------------------------------- |
| frontend  | `./frontend/Dockerfile`  | `80:80`             | Nginx serving Vue 3 SPA + proxy  |
| backend   | `./backend/Dockerfile`   | `8000:8000`, `9000:9000` | Go/Kratos API server        |
| mysql     | `mysql:8.0`              | `3306:3306`         | Primary database                 |
| redis     | `redis:7-alpine`         | `6379:6379`         | Cache, view buffer, sessions, presence |
| minio     | `minio/minio:latest`     | `9100:9000`, `9101:9001` | S3-compatible object storage |
| nats      | `nats:2-alpine`          | `4222:4222`, `8222:8222` | Message broker (JetStream)  |
| jaeger    | `jaegertracing/all-in-one` | `16686:16686`, `4317-4318` | Distributed tracing      |

---

## Quick Start (Full Stack)

### Prerequisites

- **Docker** & **Docker Compose** v2+
- **Go** 1.24+ (for local backend development)
- **Node.js** 20+ & **npm** (for local frontend development)
- **protoc** + plugins (for proto code generation)

### Option A: Docker Compose (Production-like)

```bash
# Clone the repository
git clone <repo-url> && cd fenzVideo

# Start everything (infrastructure + backend + frontend)
docker-compose up -d --build

# Open the app
open http://localhost        # Frontend (Nginx)
open http://localhost:9101   # MinIO Console
open http://localhost:16686  # Jaeger UI
open http://localhost:8222   # NATS Monitoring

# Admin login: admin / admin123
```

### Option B: Local Development (Recommended for Dev)

```bash
# 1. Start infrastructure only
docker-compose up -d mysql redis minio nats jaeger

# 2. Start backend (from project root)
cd backend
make init          # Install protoc plugins (first time only)
make all           # Generate proto code
go run ./cmd/backend/ -conf ./configs/

# 3. Start frontend (in another terminal)
cd frontend
npm install        # First time only
npm run dev        # Vite dev server at http://localhost:5173

# 4. (Optional) Seed sample data
cd backend
make seed              # GEMINI_KEY optional — 57 pre-defined videos work without it
```

---

## Request Flow (End-to-End)

### Example: User opens homepage and sees recommended videos

```
Browser                   Frontend (Vue)              Backend (Kratos)             MySQL / Redis
   │                          │                             │                          │
   │── GET / ───────────────▶ │                             │                          │
   │                          │ Vue Router → HomeView       │                          │
   │                          │ tagStore.fetchMyTags()      │                          │
   │                          │── GET /api/v1/tags/my ────▶ │                          │
   │                          │   ?session_id=<uuid>        │                          │
   │                          │                             │── Query user_tag_prefs ──▶│
   │                          │                             │◀── tag_ids [1,3,7] ──────│
   │                          │◀── { tags: [...] } ────────│                          │
   │                          │                             │                          │
   │                          │ videoStore.fetchRecommended()│                          │
   │                          │── GET /api/v1/recommended ─▶│                          │
   │                          │   ?tags=1,3,7&page=1        │                          │
   │                          │                             │── SUNIONSTORE tag SETs ──▶│ Redis
   │                          │                             │── Pipeline HGETALL ──────▶│ Redis
   │                          │                             │◀── video summaries ───────│
   │                          │◀── { videos: [...] } ──────│                          │
   │                          │                             │                          │
   │◀── Render VideoGrid ────│                             │                          │
```

### Example: User watches a video

```
Browser                   Frontend (Vue)              Backend (Kratos)             Infrastructure
   │                          │                             │                          │
   │── Click VideoCard ──────▶│                             │                          │
   │                          │ Router → /video/:id         │                          │
   │                          │── GET /api/v1/videos/42 ──▶ │                          │
   │                          │   Authorization: Bearer JWT │                          │
   │                          │                             │── JWT middleware ────────▶│
   │                          │                             │   Extract user_id, role  │
   │                          │                             │── Check access tier ─────▶│ MySQL
   │                          │                             │── Check membership ──────▶│ MySQL
   │                          │                             │── HINCRBY views:buffer ──▶│ Redis
   │                          │                             │── ZINCRBY popular ───────▶│ Redis
   │                          │◀── { video + video_url } ──│                          │
   │                          │                             │                          │
   │◀── Video.js player ─────│                             │                          │
   │── GET /fenzvideo/videos/xx ──────────────────────────────────────────────────────▶│ MinIO
   │   (via Nginx /fenzvideo/ proxy)                                                   │
   │◀── .mp4 stream ──────────────────────────────────────────────────────────────────│
```

---

## Backend Layer Flow

Every API request traverses four layers:

```
HTTP/gRPC Request
       │
       ▼
┌──────────────────────┐
│  Transport (server/)  │  ← HTTP/gRPC server, middleware (JWT, CORS, admin guard)
│  middleware.go        │
│  http.go / grpc.go    │
└──────────┬───────────┘
           │ Kratos operation routing
           ▼
┌──────────────────────┐
│  Service (service/)   │  ← Proto request ↔ Domain object mapping
│  video.go, auth.go    │     Calls biz use cases
└──────────┬───────────┘
           │ Domain objects
           ▼
┌──────────────────────┐
│  Business (biz/)      │  ← Domain rules, validation, orchestration
│  video.go, auth.go    │     Defines repo interfaces (Dependency Inversion)
└──────────┬───────────┘
           │ Repo interfaces
           ▼
┌──────────────────────┐
│  Data (data/)         │  ← GORM queries, Redis cache, MinIO operations
│  video.go, auth.go    │     Implements biz repo interfaces
│  video_cache.go       │
│  cleanup_worker.go    │
└──────────────────────┘
```

### Dependency Injection (Wire)

All layers are wired together at compile time via [Google Wire](https://github.com/google/wire):

```
cmd/backend/wire.go  →  wire_gen.go (auto-generated)

Providers:
  data.ProviderSet     → NewData, NewAuthRepo, NewVideoRepo, ...
  biz.ProviderSet      → NewAuthUsecase, NewVideoUsecase, ...
  service.ProviderSet  → NewAuthService, NewVideoService, ...
  server.ProviderSet   → NewHTTPServer, NewGRPCServer
```

---

## Frontend Architecture Flow

```
┌──────────────────────────────────────────────────────────────┐
│  App.vue (dynamic layout selection based on route meta)      │
│  ┌────────────────────────────────────────────────────────┐  │
│  │  Layout (DefaultLayout / AuthLayout / AdminLayout)     │  │
│  │  ┌──────────────┐  ┌──────────────────────────────┐   │  │
│  │  │  AppHeader    │  │        <router-view />       │   │  │
│  │  │  - SearchBar  │  │  (HomeView / VideoView /     │   │  │
│  │  │  - Auth links │  │   AdminView / etc.)          │   │  │
│  │  └──────────────┘  └──────────────────────────────┘   │  │
│  │  ┌──────────────┐                                     │  │
│  │  │  AppSidebar   │                                     │  │
│  │  │  - Categories │  ← categoryStore                    │  │
│  │  │  - TagSelector│  ← tagStore (max 5 tags)            │  │
│  │  └──────────────┘                                     │  │
│  └────────────────────────────────────────────────────────┘  │
└──────────────────────────────────────────────────────────────┘
```

### Data Flow: View → Store → API → Backend

```
Vue Component          Pinia Store           API Module           Backend
     │                      │                    │                    │
     │── action() ─────────▶│                    │                    │
     │                      │── apiCall() ──────▶│                    │
     │                      │                    │── HTTP request ───▶│
     │                      │                    │◀── JSON response ──│
     │                      │◀── return data ────│                    │
     │                      │ (update state)     │                    │
     │◀── reactive update ──│                    │                    │
     │    (auto re-render)  │                    │                    │
```

---

## Caching Strategy

### Boot-Time Cache Warm-Up

```
App Startup (NewData)
       │
       ├── Connect MySQL, Redis, MinIO, NATS
       ├── AutoMigrate (create/update tables)
       ├── ensureAdmin() (create admin from config)
       │
       └── WarmUpCache()
              │
              ├── SELECT * FROM videos WHERE is_published AND NOT is_hidden AND access_tier=0
              ├── SELECT * FROM video_tags
              │
              ├── For each video:
              │     Redis HSET video:{id} (title, thumbnail, views, etc.)
              │
              ├── For each tag:
              │     Redis SADD tag:{id} video_id1 video_id2 ...
              │
              └── Redis ZADD popular:global (score=total_views)

  ✅ Server starts accepting traffic with warm cache
```

### Runtime Cache Operations

```
┌──────────────────────────────────────────────────┐
│                 Read Path                         │
│                                                  │
│  GetRecommended(tagIds)                          │
│    │                                              │
│    ├─ Redis SUNIONSTORE → combined video IDs     │
│    ├─ Redis Pipeline HGETALL → video summaries   │
│    │                                              │
│    └─ Cache MISS? → MySQL fallback               │
│         └─ Re-populate cache + return             │
└──────────────────────────────────────────────────┘

┌──────────────────────────────────────────────────┐
│                 Write Path                        │
│                                                  │
│  View recorded:                                  │
│    ├─ Redis HINCRBY views:buffer video:{id}      │
│    └─ Redis ZINCRBY popular:global video:{id}    │
│                                                  │
│  Background worker (every 30s):                  │
│    ├─ HGETALL views:buffer                       │
│    ├─ Batch UPDATE videos SET views += N         │
│    └─ DEL views:buffer                           │
│                                                  │
│  Video deleted:                                  │
│    ├─ EvictVideo: SREM from tag SETs             │
│    ├─ DEL video:{id} HASH                        │
│    ├─ ZREM popular:global                        │
│    └─ On failure → RPUSH cleanup:queue (retry)   │
└──────────────────────────────────────────────────┘
```

---

## Authentication Flow

### JWT Token Lifecycle

```
┌─────────┐          ┌──────────┐          ┌─────────┐
│ Browser │          │ Frontend │          │ Backend │
└────┬────┘          └────┬─────┘          └────┬────┘
     │                    │                     │
     │  Login form        │                     │
     │───────────────────▶│                     │
     │                    │  POST /auth/login   │
     │                    │────────────────────▶│
     │                    │                     │ bcrypt verify
     │                    │                     │ Generate JWT (24h)
     │                    │                     │ Generate Refresh (7d)
     │                    │◀────────────────────│
     │                    │ Save to localStorage│
     │                    │                     │
     │  (24h later...)    │                     │
     │                    │  Any API call       │
     │                    │────────────────────▶│
     │                    │◀── 401 Unauthorized │
     │                    │                     │
     │                    │  POST /auth/refresh │
     │                    │  { refresh_token }  │
     │                    │────────────────────▶│
     │                    │◀── new token pair ──│
     │                    │  Save & retry       │
     │                    │                     │
```

### Middleware Pipeline

```
Incoming Request
       │
       ▼
┌──────────────────┐
│  CORS Middleware  │  ← Allow frontend origin
├──────────────────┤
│  JWT Middleware   │  ← Extract & validate token
│  (skip public    │     Set user_id + role in context
│   routes)        │     Public routes: extract optional JWT
├──────────────────┤
│  Admin Guard     │  ← For /Admin* operations
│  (check role)    │     Reject if role != 'admin'
├──────────────────┤
│  Route Handler   │  ← Kratos operation routing
└──────────────────┘
```

---

## File Upload Flow

FenzVideo uses a **two-step upload** pattern:

```
Creator                  Frontend                    Backend                    MinIO
  │                         │                           │                        │
  │  Select video file      │                           │                        │
  │────────────────────────▶│                           │                        │
  │                         │  POST /api/v1/upload/video│                        │
  │                         │  (multipart/form-data)    │                        │
  │                         │──────────────────────────▶│                        │
  │                         │                           │── PutObject() ────────▶│
  │                         │                           │◀── OK ────────────────│
  │                         │◀── { url: "videos/xxx" } ─│                        │
  │                         │                           │                        │
  │  Select thumbnail       │                           │                        │
  │────────────────────────▶│                           │                        │
  │                         │  POST /api/v1/upload/thumb│                        │
  │                         │──────────────────────────▶│                        │
  │                         │◀── { url: "thumbs/xxx" } ─│                        │
  │                         │                           │                        │
  │  Fill in metadata       │                           │                        │
  │────────────────────────▶│                           │                        │
  │                         │  POST /api/v1/videos      │                        │
  │                         │  { title, video_url,      │                        │
  │                         │    thumbnail_url, tags }  │                        │
  │                         │──────────────────────────▶│                        │
  │                         │                           │── INSERT video ───────▶│ MySQL
  │                         │                           │── Update cache ───────▶│ Redis
  │                         │◀── { video } ─────────────│                        │
  │◀── Success ─────────────│                           │                        │
```

---

## Real-Time Workflow

### Creator Presence and Like Alert

```
Creator Browser           Frontend SPA              Backend API / WS            Redis / NATS / MySQL
       │                        │                           │                              │
       │ Open dashboard         │                           │                              │
       │───────────────────────▶│ WS connect /api/v1/realtime/ws                           │
       │                        │──────────────────────────▶│                              │
       │                        │  Authorization: Bearer JWT│── SET presence:user:{id} ───▶│ Redis
       │                        │                           │── SADD connections ... ─────▶│ Redis
       │◀──── WS connected ─────│                           │                              │
       │                        │                           │                              │
Viewer Browser               Frontend SPA              Backend API                    │
       │ Press Like on video      │                           │                              │
       │─────────────────────────▶│ POST /api/v1/videos/:id/likes                         │
       │                        │──────────────────────────▶│── Persist like / dedupe ───▶│ MySQL/Redis
       │                        │                           │── Publish notification ─────▶│ NATS
       │                        │                           │                              │
       │                        │                           │ Notification worker checks ─▶│ Redis
       │                        │                           │ creator online?              │
       │                        │                           │── INSERT notification ──────▶│ MySQL
       │                        │                           │── WS push video_liked ──────▶│ Creator Browser
```

### Illegal Media Moderation Alert

```
Admin Browser             Backend API / Admin         NATS / Notifications        Creator Browser
     │                           │                           │                            │
     │ Delete illegal video      │                           │                            │
     │──────────────────────────▶│ DELETE /api/v1/admin/videos/:id                         │
     │                           │── Transactional delete/hide                            │
     │                           │── INSERT moderation notification ─────────────────────▶│ MySQL
     │                           │── Publish moderation_removed ─────────────────────────▶│ NATS
     │                           │                           │── Check creator presence ─▶│ Redis
     │                           │                           │── If online: WS push ─────▶│ warning toast + inbox
```

---

## Development Workflow

### Backend Development

```bash
# Proto changes → regenerate code
cd backend
make api               # Generate Go code from api/*.proto
make config            # Generate Go code from internal/conf/*.proto
make generate          # Run go generate + tidy

# Run locally
go run ./cmd/backend/ -conf ./configs/

# Build binary
make build             # Output: bin/backend

# Seed data (GEMINI_KEY optional — 57 pre-defined videos work without it)
make seed
```

### Frontend Development

```bash
cd frontend
npm run dev            # Vite dev server (port 5173)
                       # Auto-proxies /api/* → localhost:8000

npm run build          # Production build → dist/
npm run preview        # Preview production build
```

### Docker Development

```bash
# Build & start everything
docker-compose up -d --build

# View logs
docker-compose logs -f backend
docker-compose logs -f frontend

# Rebuild single service
docker-compose up -d --build backend

# Stop everything
docker-compose down

# Reset data (remove volumes)
docker-compose down -v
```

---

## Configuration

### Local Development (`backend/configs/config.yaml`)

Uses `127.0.0.1` addresses, suitable for running the backend directly on the host while infrastructure runs in Docker.

### Docker Compose (`backend/configs/config.docker.yaml`)

Uses Docker container hostnames (e.g., `fenzvideo-mysql`, `fenzvideo-redis`), suitable for when the backend runs inside Docker alongside infrastructure.

| Setting                   | Local (`config.yaml`)          | Docker (`config.docker.yaml`)      |
| ------------------------- | ------------------------------ | ---------------------------------- |
| MySQL                     | `127.0.0.1:3306`              | `fenzvideo-mysql:3306`             |
| Redis                     | `127.0.0.1:6379`              | `fenzvideo-redis:6379`             |
| MinIO                     | `127.0.0.1:9100`              | `fenzvideo-minio:9000`             |
| NATS                      | `nats://127.0.0.1:4222`       | `nats://fenzvideo-nats:4222`       |

---

## Port Reference

| Port  | Service                 | Access URL                       |
| ----- | ----------------------- | -------------------------------- |
| 80    | Frontend (Nginx)        | `http://localhost`               |
| 5173  | Frontend (Vite dev)     | `http://localhost:5173`          |
| 8000  | Backend HTTP API        | `http://localhost:8000/api/v1/*` |
| 9000  | Backend gRPC API        | `grpc://localhost:9000`          |
| 3306  | MySQL                   | `mysql://localhost:3306`         |
| 6379  | Redis                   | `redis://localhost:6379`         |
| 9100  | MinIO S3 API            | `http://localhost:9100`          |
| 9101  | MinIO Console           | `http://localhost:9101`          |
| 4222  | NATS Client             | `nats://localhost:4222`          |
| 8222  | NATS Monitoring         | `http://localhost:8222`          |
| 16686 | Jaeger UI               | `http://localhost:16686`         |
| 4317  | Jaeger OTLP gRPC        | —                                |
| 4318  | Jaeger OTLP HTTP        | —                                |

---

## API Endpoints (Current)

### Public (No Auth)

| Method | Path                    | Description               |
| ------ | ----------------------- | ------------------------- |
| POST   | `/api/v1/auth/login`    | Login                     |
| POST   | `/api/v1/auth/register` | Register                  |
| POST   | `/api/v1/auth/refresh`  | Refresh token             |
| GET    | `/api/v1/categories`    | List categories           |
| GET    | `/api/v1/tags`          | List all tags             |
| GET    | `/api/v1/recommended`   | Get recommended videos    |
| GET    | `/api/v1/videos/:id`    | Get single video          |
| GET    | `/api/v1/channels/:id`  | Get channel info          |
| GET    | `/api/v1/search`        | Search videos             |

### Authenticated (JWT Required)

| Method | Path                             | Description               |
| ------ | -------------------------------- | ------------------------- |
| GET    | `/api/v1/tags/my`                | Get user's selected tags  |
| PUT    | `/api/v1/tags/my`                | Set user's tags (max 5)   |
| POST   | `/api/v1/videos`                 | Create video              |
| PUT    | `/api/v1/videos/:id`             | Update video              |
| DELETE | `/api/v1/videos/:id`             | Delete video              |
| PATCH  | `/api/v1/videos/:id/publish`     | Toggle publish status     |
| POST   | `/api/v1/channels/:id/subscribe` | Subscribe to channel      |
| DELETE | `/api/v1/channels/:id/subscribe` | Unsubscribe from channel  |
| POST   | `/api/v1/upload/video`           | Upload video file (MinIO) |
| POST   | `/api/v1/upload/thumbnail`       | Upload thumbnail (MinIO)  |
| GET    | `/api/v1/realtime/ws`            | WebSocket upgrade for live notifications |
| POST   | `/api/v1/videos/:id/likes`       | Like a video and trigger creator alert |

### Admin (JWT + Admin Role)

| Method | Path                        | Description             |
| ------ | --------------------------- | ----------------------- |
| GET    | `/api/v1/admin/users`       | List all users          |
| DELETE | `/api/v1/admin/users/:id`   | Delete user (cascade)   |
| GET    | `/api/v1/admin/videos`      | List all videos         |
| DELETE | `/api/v1/admin/videos/:id`  | Delete video            |
| POST   | `/api/v1/admin/tags`        | Create tag              |
| PUT    | `/api/v1/admin/tags/:id`    | Update tag              |
| DELETE | `/api/v1/admin/tags/:id`    | Delete tag              |

---

## Troubleshooting

### Backend won't connect to MySQL in Docker

```bash
# Check MySQL is healthy
docker-compose ps
docker-compose logs mysql

# Verify the backend config uses correct hostname
# Local dev: 127.0.0.1:3306
# Docker:    fenzvideo-mysql:3306
```

### Frontend proxy not working (Vite dev)

```bash
# Ensure backend is running on port 8000
curl http://localhost:8000/api/v1/categories

# Check vite.config.ts proxy target
# Should be: target: 'http://localhost:8000'
```

### Proto compilation errors

```bash
# Install required protoc plugins
cd backend && make init

# Regenerate all proto code
make all

# If VS Code shows import errors but protoc works:
# This is an IDE configuration issue — see buf.yaml for roots
```

### Redis cache is stale

```bash
# Restart the backend to trigger WarmUpCache()
# Or flush Redis manually:
docker exec fenzvideo-redis redis-cli FLUSHDB
```

### MinIO bucket missing

```bash
# The backend auto-creates the bucket on startup
# Or create manually:
docker exec fenzvideo-minio mc alias set local http://localhost:9000 minioadmin minioadmin
docker exec fenzvideo-minio mc mb local/fenzvideo
```

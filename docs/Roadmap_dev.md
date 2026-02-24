# FenzVideo Development Roadmap

## Project Overview

FenzVideo is a tag-based video streaming platform with monetization support (memberships + donations). Built with Go/Kratos backend, Vue 3 frontend, and a fully open-source infrastructure stack.

---

## Architecture Strategy: Modular Monolith → Microservices

**Start as a modular monolith.** All 11 services live in a single Go binary using Kratos's clean architecture. Extract individual services into microservices only when measured bottlenecks appear.

### Bounded Contexts

| Context | Services | Shared Data |
|---------|----------|-------------|
| **Identity & Access** | Auth, User, Admin | `users` |
| **Content & Discovery** | Video, Tag, Search, Category | `videos`, `tags`, `video_tags`, `categories`, `view_records` |
| **Monetization** | Channel, Donation, Dashboard | `channels`, `memberships`, `donations` |
| **Engagement** | Notification | `notifications` |

### Future Microservice Extraction (only when bottlenecks measured)

| Priority | Service | Trigger |
|----------|---------|---------|
| 1st | NotificationService | Fan-out write bursts impacting main DB |
| 2nd | Video Upload Subsystem | Upload traffic saturating main API |
| 3rd | SearchService | MySQL FULLTEXT outgrown |

---

## Phase 1: Foundation ✅

> **Goal**: App boots, connects to all infrastructure, seeds data, serves health checks.

### Deliverables

- [x] Extended `conf.proto` with Auth, Storage, Paddle, NATS configuration
- [x] `docker-compose.yaml` with MySQL, Redis, MinIO, NATS
- [x] All GORM models (12 tables: users, channels, videos, categories, tags, video_tags, user_tag_preferences, memberships, view_records, notifications, donations)
- [x] Data layer initialization (GORM + Redis + MinIO + NATS clients, AutoMigrate)
- [x] Internal packages: JWT, bcrypt hash, MinIO upload, pagination
- [x] Middleware: JWT authentication, admin guard, CORS
- [x] Error reason proto with all error codes
- [x] Removed helloworld demo code
- [x] Seed data generator (`cmd/seed/main.go`):
  - Gemini API integration to generate video titles & descriptions in Traditional Chinese
  - Creates 1 admin user + 5 creator users with channels
  - Seeds 10 categories and 15 tags
  - Generates 15 videos (one per tag) with AI-generated content
  - Assigns random view counts, durations, and categories
  - Idempotent: skips already-seeded data on re-run
- [x] Cache warm-up on boot (`internal/data/cache_warmup.go`):
  - `WarmUpCache()` runs in `NewData()` before servers accept traffic
  - Loads all public videos into Redis (per-tag SETs + per-video HASHes)
  - Eliminates cold start: first user gets cache HIT immediately
  - Self-healing: app restart re-warms from MySQL (no tracking table needed)

### Verification

```bash
# Start infrastructure
docker-compose up -d

# Build and run
cd backend && make config && go build ./...

# Seed sample data (requires GEMINI_KEY in .env)
cd backend && make seed
```

---

## Phase 2: Core MVP ✅

> **Goal**: Users can register, login, upload videos, browse by tags, and search.

### Infrastructure Changes
- [x] Middleware updated from HTTP paths to Kratos operation names (`/fenzvideo.v1.AuthService/Login`)
- [x] Public endpoints now extract optional JWT tokens (e.g. `GetVideo`, `GetRecommended` can identify logged-in users)
- [x] Extracted `UserIDFromContext`/`RoleFromContext` into `internal/pkg/authctx` to break import cycle between `server` and `service` packages
- [x] Two-step file upload endpoints: `POST /api/v1/upload/video` and `POST /api/v1/upload/thumbnail` (MinIO)
- [x] Wire DI integration: all 6 services + `MembershipChecker` adapter + `MinIOUploader` provider + `VideoCache`
- [x] FULLTEXT index on `videos(title)` created manually after AutoMigrate

### 2.1 AuthService
- [x] `api/fenzvideo/v1/auth.proto` (Login, Register, RefreshToken)
- [x] `internal/biz/auth.go` (AuthUsecase + AuthRepo interface)
- [x] `internal/data/auth.go` (AuthRepo implementation with GORM)
- [x] `internal/service/auth.go` (gRPC/HTTP handler)
- [x] Wire DI integration
- [ ] **Test**: Register → Login → Get JWT → Refresh token

### 2.2 CategoryService
- [x] `api/fenzvideo/v1/category.proto` (ListCategories)
- [x] Biz/Data/Service layers
- [x] Seed 10 categories via SQL init script
- [ ] **Test**: List all categories

### 2.3 TagService
- [x] `api/fenzvideo/v1/tag.proto` (ListTags, GetMyTags, SetMyTags)
- [x] Biz/Data/Service layers
- [x] Guest session_id support for anonymous tag preferences
- [x] Max 5 tags per user enforcement
- [x] Seed 15 tags via SQL init script
- [ ] **Test**: List tags, set/get user preferences, guest flow

### 2.4 VideoService
- [x] `api/fenzvideo/v1/video.proto` (CRUD + GetRecommended + TogglePublish)
- [x] Biz/Data/Service layers
- [x] MinIO file upload (two-step: upload file → get path → CreateVideo RPC)
- [x] Tag-based recommendation algorithm (random subset of user tags)
- [x] Access tier enforcement (public/subscriber/premium via MembershipChecker interface)
- [x] View counting (member vs non-member, buffered through Redis)
- [ ] **Test**: Upload video → Get recommended → Watch video → View count increments

### 2.5 SearchService
- [x] `api/fenzvideo/v1/search.proto` (Search with filters)
- [x] MySQL FULLTEXT search on video title (BOOLEAN MODE)
- [x] Filters: category, duration range, date range, view sort, access type
- [x] Pagination
- [ ] **Test**: Search with various filter combinations

### 2.6 ChannelService (Free Tier)
- [x] `api/fenzvideo/v1/channel.proto` (GetChannel, Subscribe, Unsubscribe)
- [x] Auto-create channel on user registration (in AuthUsecase.Register)
- [x] Tier 1 (free) subscription flow
- [ ] **Test**: View channel → Subscribe → Verify membership → Unsubscribe

### 2.7 Recommendation Cache (Redis)
- [x] `internal/data/cache_warmup.go` (boot-time cache warm-up from MySQL → Redis)
- [x] `internal/data/video_cache.go` (VideoCache: SUNION tag SETs → pipeline HGETALL video HASHes)
- [x] Cache-first reads in `videoRepo.ListByTags` (fall back to MySQL on miss)
- [x] Cache only public, non-hidden, non-deleted, non-premium videos (`access_tier = 0`)
- [x] Redis structures:
  - `tag:{id}` → SET of video IDs (index layer, 30min TTL)
  - `video:{id}` → HASH of video summary (data layer, 30min TTL)
  - `popular:global` → ZSET scored by total view count
  - `views:buffer` → HASH of buffered view increments
  - `cleanup:queue` → LIST of failed eviction jobs
- [x] Application-level cache eviction hooks:
  - `EvictVideo`: removes from tag SETs + deletes video HASH + ZREM popular
  - On failure → pushes to `cleanup:queue` for background retry
- [x] View count buffer: `HINCRBY views:buffer` + `ZINCRBY popular:global` on each view → flush to MySQL every 30s
- [x] `internal/data/cleanup_worker.go` (background workers):
  - View flush ticker (every 30s): drains `views:buffer` → batch UPDATE MySQL
  - Cleanup worker (every 10s): retries failed evictions from `cleanup:queue`
  - Both use cancelable context — stopped on app shutdown
- [x] TTL safety net on all cache keys (30min for tags/videos, 10min for popular)
- [ ] 10-minute upload cooldown before creator can edit/delete (prevents rapid cache churn)
- [ ] Rate limit on MySQL queries from cache miss path
- [ ] **Test**: Boot → Verify warm-up populated Redis → Recommend (cache hit) → Delete video → Verify eviction

### Verification

```bash
# Full MVP flow
curl -X POST localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test123","display_name":"Test"}'

curl -X POST localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test123"}'

# Public endpoints
curl localhost:8000/api/v1/categories
curl localhost:8000/api/v1/tags
curl localhost:8000/api/v1/videos/recommended

# Protected endpoints (use token from login)
curl -H "Authorization: Bearer <token>" localhost:8000/api/v1/tags/my
curl localhost:8000/api/v1/search?query=test
```

---

## Phase 3: Monetization

> **Goal**: Premium memberships and donations work end-to-end with Paddle.

### 3.1 Paddle Integration Package
- [ ] `internal/pkg/paddle/paddle.go` (Paddle SDK client)
- [ ] Webhook signature verification
- [ ] Sandbox environment configuration

### 3.2 ChannelService (Tier 2 - Premium)
- [ ] UpgradeToPremium endpoint → Paddle subscription creation
- [ ] CancelPremium endpoint → Paddle subscription cancellation
- [ ] Premium video access tier enforcement
- [ ] **Test**: Upgrade to premium → Access premium video → Cancel

### 3.3 DonationService
- [ ] `api/fenzvideo/v1/donation.proto`
- [ ] CreateDonation → Paddle one-time transaction
- [ ] ListByDonor, ListByCreator
- [ ] **Test**: Create donation → Paddle checkout → List donations

### 3.4 Paddle Webhook Handler
- [ ] `POST /api/v1/webhooks/paddle` (no auth, signature verified)
- [ ] Route events: `transaction.*` → DonationService, `subscription.*` → ChannelService
- [ ] Handle: completed, payment_failed, refunded, activated, canceled, past_due
- [ ] **Test**: Simulate webhook events, verify status updates

### 3.5 DashboardService
- [ ] `api/fenzvideo/v1/dashboard.proto`
- [ ] GetMyVideos, GetAnalytics, SetMembershipFee
- [ ] Analytics aggregation: views breakdown, rankings, member count, revenue
- [ ] **Test**: Verify analytics with sample data

### Verification

```bash
# Paddle sandbox test
# 1. Create premium subscription
# 2. Verify Paddle checkout URL returned
# 3. Complete payment in Paddle sandbox
# 4. Verify webhook updates membership status
# 5. Access premium-only video
```

---

## Phase 4: Advanced Features

> **Goal**: Complete backend feature set with notifications, admin panel, and user self-service.

### 4.1 NATS Integration
- [ ] `internal/pkg/nats/nats.go` (NATS client: Publish, Subscribe)
- [ ] Background subscriber goroutine on app boot
- [ ] Publish events on video create/update

### 4.2 NotificationService
- [ ] `api/fenzvideo/v1/notification.proto`
- [ ] NATS subscriber → fan-out notifications to channel subscribers
- [ ] ListNotifications, GetUnreadCount, MarkRead, MarkAllRead
- [ ] **Test**: Create video → Verify notifications for all subscribers

### 4.3 UserService
- [ ] `api/fenzvideo/v1/user.proto`
- [ ] UpdateDisplayName, UpdatePassword
- [ ] HideAccount (cascade hide: user + channel + videos)
- [ ] DeleteAccount (cascade delete: user + channel + videos + MinIO files)
- [ ] DeleteChannel (delete channel + videos, keep user)
- [ ] **Test**: All self-service operations with cascade verification

### 4.4 AdminService
- [ ] `api/fenzvideo/v1/admin.proto`
- [ ] User CRUD: ListUsers, GetUser, CreateUser, UpdateUser
- [ ] User moderation: HideUser, RestoreUser, DeleteUser
- [ ] Tag management: CreateTag, UpdateTag, DeleteTag
- [ ] Admin guard middleware enforcement
- [ ] Prevent admin self-delete
- [ ] **Test**: Full admin CRUD lifecycle

### 4.5 Observability
- [ ] OpenTelemetry tracing integration with Jaeger
- [ ] Prometheus metrics endpoints
- [ ] Grafana dashboard templates
- [ ] Structured logging with Kratos logger

### Verification

```bash
# Notification flow
# 1. User A subscribes to User B's channel
# 2. User B publishes a video
# 3. User A receives notification
# 4. User A marks notification as read

# Admin flow
# 1. Admin lists users
# 2. Admin hides a user
# 3. Verify user's videos are hidden
# 4. Admin restores user
```

---

## Phase 5: Frontend (Vue 3 SPA)

> **Goal**: Full user-facing web application.

### 5.1 Project Setup
- [ ] Scaffold Vue 3 + Vite + TypeScript
- [ ] Install Element Plus + Tailwind CSS
- [ ] Configure Vue Router, Pinia, Axios, Vue I18n (zh-TW / en)

### 5.2 Core Pages
- [ ] **LoginView** - Login + Register forms
- [ ] **HomeView** - Tag-based recommended video grid
- [ ] **VideoView** - Video.js player + info + donate button
- [ ] **ChannelView** - Channel info + videos + membership CTA
- [ ] **SearchResultsView** - Search with filters sidebar
- [ ] **CategoryView** - Videos by category

### 5.3 Creator Pages
- [ ] **DashboardView** - Video management + analytics (ECharts)
- [ ] **VideoUploadForm** - Upload with category/tag/access tier selection

### 5.4 Admin Pages
- [ ] **AdminUserTable** - User list with hide/restore/delete
- [ ] **AdminTagTable** - Tag management

### 5.5 State Management (Pinia Stores)
- [ ] authStore, videoStore, tagStore, searchStore
- [ ] channelStore, dashboardStore, donationStore, adminStore

### 5.6 Components
- [ ] AppHeader (nav + search), AppSidebar (categories + tags)
- [ ] VideoCard, VideoPlayer, VideoUploadForm
- [ ] VideoDonateDialog, MembershipDialog
- [ ] AnalyticsCharts, Notification bell

### 5.7 Integration
- [ ] Paddle.js checkout overlay
- [ ] JWT interceptors (auto-refresh, redirect on 401)
- [ ] Responsive design

### 5.8 Testing
- [ ] Unit tests: Vitest + Vue Test Utils
- [ ] E2E tests: Playwright

---

## Phase 6: Deployment & Operations

> **Goal**: Production-ready deployment.

### 6.1 Infrastructure
- [ ] Production `docker-compose.yaml` with all services
- [ ] Nginx reverse proxy with SSL termination
- [ ] MinIO bucket policies and access control
- [ ] NATS JetStream configuration

### 6.2 CI/CD
- [ ] Gitea Actions / Woodpecker CI pipeline
- [ ] Automated testing on push
- [ ] Docker image build and push
- [ ] Staging → Production promotion

### 6.3 Monitoring
- [ ] Jaeger distributed tracing dashboards
- [ ] Prometheus alerting rules
- [ ] Grafana dashboards (request rate, error rate, latency)

---

## Phase 7: Microservice Extraction (When Needed)

> **Trigger**: Extract only when 2+ criteria met: measured bottleneck, team grown to 3+, different tech needed, different availability requirements.

### 7.1 NotificationService Extraction
- [ ] Separate Go module with own `cmd/` entry point
- [ ] Own database/schema for `notifications` table
- [ ] NATS subscriber as standalone process
- [ ] gRPC client for membership fan-out queries

### 7.2 Video Upload Subsystem Extraction
- [ ] Separate upload worker service
- [ ] Async upload via message queue
- [ ] Progress tracking via Redis

### 7.3 SearchService Extraction
- [ ] Migrate from MySQL FULLTEXT to Elasticsearch/Meilisearch
- [ ] Event-driven index updates (video CRUD → search index sync)
- [ ] Dedicated search API endpoint

---

## Tech Stack Summary

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.22+, Kratos v2, gRPC + HTTP, Wire DI |
| **Database** | MySQL 8.0, GORM v2 |
| **Cache** | Redis 7 |
| **Storage** | MinIO (S3-compatible) |
| **Messaging** | NATS 2 (JetStream) |
| **Payment** | Paddle (sandbox) |
| **Frontend** | Vue 3, Vite, Element Plus, Tailwind CSS, Video.js |
| **Observability** | OpenTelemetry, Jaeger, Prometheus, Grafana |
| **CI/CD** | Gitea Actions / Woodpecker CI |
| **Container** | Docker + Docker Compose |

---

## Timeline Estimate

| Phase | Scope |
|-------|-------|
| Phase 1 | Foundation (infra, models, middleware) |
| Phase 2 | Core MVP (auth, video, tags, search, channels) |
| Phase 3 | Monetization (Paddle, premium, donations, dashboard) |
| Phase 4 | Advanced (notifications, admin, user self-service) |
| Phase 5 | Frontend (Vue 3 SPA) |
| Phase 6 | Deployment & Operations |
| Phase 7 | Microservice extraction (as needed) |

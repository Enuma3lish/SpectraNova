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

## Phase 3: Frontend MVP (Vue 3 SPA) ✅

> **Goal**: A working web application where users can browse and play videos, and administrators can manage users and videos. This is the **MVP milestone** — the first fully usable product.

### Why move the frontend to Phase 3?

After Phase 2, the backend can create users, upload videos with tags, and the infrastructure is ready. Before adding monetization or advanced features, we need a **usable product** first. A Vue 3 SPA that lets users watch videos and lets admins manage content is the simplest complete product. Everything else (payments, notifications, dashboards) builds on top of this MVP.

### 3.0 Admin Account via `.env`

- [x] Administrator username/password configured in `.env` file (`ADMIN_USERNAME`, `ADMIN_PASSWORD`)
- [x] On app boot, auto-create or verify admin account from `.env` values (idempotent)
- [x] No admin registration endpoint — admin is provisioned from environment only
- [x] Admin config added to `conf.proto` (`Admin` message with `username`, `password`)
- [x] `ensureAdmin()` in `data.go` — creates admin user + channel on boot, updates password if changed

**Why fix admin in `.env`?**
- Single maintainer: no need for an admin registration flow
- Easy to rotate credentials by updating `.env` and restarting
- Keeps admin provisioning simple and secure — no exposed endpoint

### 3.0b AdminService (Backend API)

- [x] `api/fenzvideo/v1/admin.proto` (7 RPCs with HTTP bindings)
- [x] User management: AdminListUsers, AdminDeleteUser
- [x] Video management: AdminListVideos, AdminDeleteVideo
- [x] Tag management: AdminCreateTag, AdminUpdateTag, AdminDeleteTag
- [x] Admin guard middleware enforcement (method names prefixed with "Admin" for automatic protection)
- [x] Prevent admin self-delete (ADMIN_SELF_DELETE error)
- [x] Cascade delete: user deletion removes memberships, tag preferences, view records, notifications, donations, videos, channel
- [x] Biz/Data/Service layers + Wire DI integration

**Why build AdminService here?**
The Phase 3 Vue SPA needs backend APIs for admin to delete users and videos. These endpoints must exist before the frontend can call them.

### 3.1 Project Setup
- [x] Scaffold Vue 3 + Vite + TypeScript
- [x] Install Element Plus + Tailwind CSS + Video.js
- [x] Configure Vue Router, Pinia, Axios, Vue I18n
- [x] JWT interceptors (auto-refresh, redirect on 401)
- [x] Vite proxy (`/api` → `localhost:8000`), Tailwind config, path aliases

### 3.2 Core Pages (User-Facing)
- [x] **LoginView** — Login + Register form with validation
- [x] **HomeView** — Tag-based recommended video grid with pagination
- [x] **VideoView** — HTML5 video player + video info
- [x] **SearchResultsView** — Search with filters sidebar
- [x] **CategoryView** — Videos by category
- [x] **ChannelView** — Channel info + subscribe/unsubscribe

### 3.3 Admin Pages
- [x] **AdminUsersView** — User list table with delete action
- [x] **AdminTagsView** — Tag CRUD with dialog form
- [x] Admin layout with sidebar navigation

### 3.4 State Management (Pinia Stores)
- [x] authStore (JWT tokens, user info, isLoggedIn/isAdmin computed)
- [x] videoStore (recommended videos, current video)
- [x] tagStore (tags with guest sessionId via UUID)
- [x] searchStore (query, filters, results)
- [x] categoryStore (category list)
- [x] adminStore (user/tag/video management)

### 3.5 Components
- [x] AppHeader (nav + search bar + login/admin links)
- [x] AppSidebar (categories + tag selector)
- [x] VideoCard, VideoGrid, Pagination
- [x] TagSelector (max 5 tags, clickable chips)
- [x] ConfirmDialog, LoadingSpinner
- [x] Three layouts: DefaultLayout, AuthLayout, AdminLayout

### 3.6 Testing
- [ ] Unit tests: Vitest + Vue Test Utils
- [ ] E2E tests: Playwright

### How to Run

```bash
# 1. Start infrastructure
docker-compose up -d

# 2. Start backend (port 8000)
cd backend && go run ./cmd/backend/ -conf ./configs/

# 3. Start frontend (port 5173)
cd frontend && npm run dev

# 4. Open http://localhost:5173
# Admin login: admin / admin123
```

### Verification

```bash
# MVP flow
# 1. Admin login (credentials from .env)
# 2. Admin manages users, videos, tags
# 3. User browses recommended videos by tags
# 4. Click video → player plays it
# 5. Search with filters
# 6. Admin deletes a video → removed from listing
# 7. Admin deletes a user → cascade cleanup
```

---

## Phase 4: Monetization

> **Goal**: Premium memberships and donations work end-to-end with Paddle.

### Why after the frontend MVP?

Payment integration is high-risk and high-complexity. The MVP must be stable first — users can browse, watch, and admins can manage content. Only then layer on monetization.

### 4.1 Paddle Integration Package
- [ ] `internal/pkg/paddle/paddle.go` (Paddle SDK client)
- [ ] Webhook signature verification
- [ ] Sandbox environment configuration

### 4.2 ChannelService (Tier 2 - Premium)
- [ ] UpgradeToPremium endpoint → Paddle subscription creation
- [ ] CancelPremium endpoint → Paddle subscription cancellation
- [ ] Premium video access tier enforcement
- [ ] **Test**: Upgrade to premium → Access premium video → Cancel

### 4.3 DonationService
- [ ] `api/fenzvideo/v1/donation.proto`
- [ ] CreateDonation → Paddle one-time transaction
- [ ] ListByDonor, ListByCreator
- [ ] **Test**: Create donation → Paddle checkout → List donations

### 4.4 Paddle Webhook Handler
- [ ] `POST /api/v1/webhooks/paddle` (no auth, signature verified)
- [ ] Route events: `transaction.*` → DonationService, `subscription.*` → ChannelService
- [ ] Handle: completed, payment_failed, refunded, activated, canceled, past_due
- [ ] **Test**: Simulate webhook events, verify status updates

### 4.5 DashboardService
- [ ] `api/fenzvideo/v1/dashboard.proto`
- [ ] GetMyVideos, GetAnalytics, SetMembershipFee
- [ ] Analytics aggregation: views breakdown, rankings, member count, revenue
- [ ] **Test**: Verify analytics with sample data

### 4.6 Frontend — Monetization Pages
- [ ] **ChannelView** — Channel info + videos + membership CTA
- [ ] **DashboardView** — Video management + analytics (ECharts)
- [ ] **VideoUploadForm** — Upload with category/tag/access tier selection
- [ ] VideoDonateDialog, MembershipDialog components
- [ ] Paddle.js checkout overlay integration

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

## Phase 5: Advanced Features

> **Goal**: Complete backend feature set with notifications, user self-service, and observability.

### Why after monetization?

These features make the product better but aren't required for the MVP or payment flows. Users can already watch videos, pay for premium, and donate. Notifications, self-service, and monitoring are quality-of-life improvements.

### 5.1 NATS Integration
- [ ] `internal/pkg/nats/nats.go` (NATS client: Publish, Subscribe)
- [ ] Background subscriber goroutine on app boot
- [ ] Publish events on video create/update

### 5.2 NotificationService
- [ ] `api/fenzvideo/v1/notification.proto`
- [ ] NATS subscriber → fan-out notifications to channel subscribers
- [ ] ListNotifications, GetUnreadCount, MarkRead, MarkAllRead
- [ ] **Test**: Create video → Verify notifications for all subscribers

### 5.3 UserService (Self-Service)
- [ ] `api/fenzvideo/v1/user.proto`
- [ ] UpdateDisplayName, UpdatePassword
- [ ] HideAccount (cascade hide: user + channel + videos)
- [ ] DeleteAccount (cascade delete: user + channel + videos + MinIO files)
- [ ] DeleteChannel (delete channel + videos, keep user)
- [ ] **Test**: All self-service operations with cascade verification

### 5.4 Observability
- [ ] OpenTelemetry tracing integration with Jaeger
- [ ] Prometheus metrics endpoints
- [ ] Grafana dashboard templates
- [ ] Structured logging with Kratos logger

### 5.5 Frontend — Advanced Features
- [ ] Notification bell component
- [ ] User profile / self-service pages
- [ ] AnalyticsCharts enhancements

### Verification

```bash
# Notification flow
# 1. User A subscribes to User B's channel
# 2. User B publishes a video
# 3. User A receives notification
# 4. User A marks notification as read
```

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
| Phase 3 | **Frontend MVP (Vue 3 SPA)** ✅ — admin via `.env`, browse/play videos, admin manage users/videos/tags |
| Phase 4 | Monetization (Paddle, premium, donations, dashboard) |
| Phase 5 | Advanced (notifications, user self-service, observability) |
| Phase 6 | Deployment & Operations |
| Phase 7 | Microservice extraction (as needed) |

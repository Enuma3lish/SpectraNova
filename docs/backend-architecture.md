# FenzVideo Backend Architecture

## Tech Stack (100% Open Source)

| Category         | Technology                                                                                              | License               | Description                                               |
| ---------------- | ------------------------------------------------------------------------------------------------------- | --------------------- | --------------------------------------------------------- |
| Language         | [Go](https://go.dev/) 1.24+                                                                             | BSD-3                 | High-performance compiled language                        |
| Framework        | [Kratos](https://go-kratos.dev/) v2                                                                     | MIT                   | Microservice framework by Bilibili                        |
| ORM              | [GORM](https://gorm.io/) v2                                                                             | MIT                   | Full-featured Go ORM                                      |
| Database         | [MySQL](https://www.mysql.com/) 8.0                                                                     | GPL-2.0               | Relational database                                       |
| Cache            | [Redis](https://redis.io/) 7 (or [Valkey](https://valkey.io/))                                          | BSD-3 / BSD-3         | In-memory data store (session, recommendations, hot data) |
| Auth             | [golang-jwt](https://github.com/golang-jwt/jwt) v5                                                      | MIT                   | JWT token generation & validation                         |
| API Protocol     | [gRPC](https://grpc.io/) + HTTP (Kratos dual transport)                                                 | Apache-2.0            | Dual-protocol API layer                                   |
| API Definition   | [Protocol Buffers](https://protobuf.dev/)                                                               | BSD-3                 | IDL for API contracts                                     |
| API Docs         | [Swagger UI](https://swagger.io/tools/swagger-ui/) (via protoc-gen-openapiv2)                           | Apache-2.0            | Auto-generated interactive API docs                       |
| Validation       | [protoc-gen-validate](https://github.com/bufbuild/protoc-gen-validate)                                  | Apache-2.0            | Protobuf field validation                                 |
| File Storage     | [MinIO](https://min.io/)                                                                                | AGPL-3.0              | S3-compatible self-hosted object storage                  |
| Reverse Proxy    | [Nginx](https://nginx.org/)                                                                             | BSD-2                 | Load balancer & static file serving                       |
| Observability    | [OpenTelemetry](https://opentelemetry.io/)                                                              | Apache-2.0            | Distributed tracing & metrics                             |
| Tracing          | [Jaeger](https://www.jaegertracing.io/)                                                                 | Apache-2.0            | Distributed tracing backend                               |
| Monitoring       | [Prometheus](https://prometheus.io/) + [Grafana](https://grafana.com/oss/)                              | Apache-2.0 / AGPL-3.0 | Metrics collection & dashboards                           |
| Logging          | Kratos log (structured logging)                                                                         | MIT                   | Built-in structured logger                                |
| DI               | [Wire](https://github.com/google/wire)                                                                  | Apache-2.0            | Compile-time dependency injection                         |
| Migration        | GORM AutoMigrate / [golang-migrate](https://github.com/golang-migrate/migrate)                          | MIT                   | Database schema migration                                 |
| Config           | Kratos config (YAML + env)                                                                              | MIT                   | Configuration management                                  |
| Containerization | [Docker](https://www.docker.com/) + [Docker Compose](https://docs.docker.com/compose/)                  | Apache-2.0            | Container orchestration                                   |
| CI/CD            | [Gitea Actions](https://gitea.com/) / [Woodpecker CI](https://woodpecker-ci.org/)                       | MIT / Apache-2.0      | Optional self-hosted CI/CD                                |
| Payment          | [Paddle](https://developer.paddle.com/) (Sandbox) + [Go SDK](https://github.com/PaddleHQ/paddle-go-sdk) | Proprietary / MIT     | Payment processing for donations & premium subscriptions  |
| Message Broker   | [NATS](https://nats.io/) + [nats.go](https://github.com/nats-io/nats.go)                                | Apache-2.0            | Lightweight pub/sub for real-time channel notifications   |

---

## Project Structure (Kratos Layout)

```
backend/
├── api/                          # Protobuf definitions & generated code
│   └── fenzvideo/
│       └── v1/
│           ├── auth.proto           # ✅ Login, Register, RefreshToken
│           ├── video.proto          # ✅ CRUD + GetRecommended + TogglePublish
│           ├── channel.proto        # ✅ GetChannel, Subscribe, Unsubscribe
│           ├── category.proto       # ✅ ListCategories
│           ├── tag.proto            # ✅ ListTags, GetMyTags, SetMyTags
│           ├── search.proto         # ✅ Search with filters
│           ├── admin.proto          # ✅ Admin user/video/tag management (7 RPCs)
│           ├── error_reason.proto   # ✅ Error codes
│           ├── dashboard.proto      # 🔜 Planned (Phase 4)
│           ├── donation.proto       # 🔜 Planned (Phase 4)
│           ├── notification.proto   # 🔜 Planned (Phase 5)
│           └── user.proto           # 🔜 Planned (Phase 5)
│
├── cmd/                          # Application entry points
│   ├── backend/
│   │   ├── main.go               # App bootstrap
│   │   ├── wire.go               # Wire dependency injection
│   │   └── wire_gen.go           # Wire generated code
│   └── seed/
│       └── main.go               # Seed data generator (57 pre-defined videos + optional Gemini API)
│
├── configs/                      # Configuration files
│   ├── config.yaml               # Main config (db, redis, jwt, server, admin) — local dev
│   └── config.docker.yaml        # Docker-specific config (container hostnames)
│
├── internal/                     # Private application code
│   ├── biz/                      # Business logic layer (use cases)
│   │   ├── biz.go                # ✅ Biz layer initialization (ProviderSet)
│   │   ├── auth.go               # ✅ AuthUsecase
│   │   ├── video.go              # ✅ VideoUsecase
│   │   ├── channel.go            # ✅ ChannelUsecase
│   │   ├── category.go           # ✅ CategoryUsecase
│   │   ├── tag.go                # ✅ TagUsecase
│   │   ├── search.go             # ✅ SearchUsecase
│   │   └── admin.go              # ✅ AdminUsecase
│   │
│   ├── conf/                     # Config struct definitions
│   │   └── conf.proto            # Protobuf-based config (incl. Admin block)
│   │
│   ├── data/                     # Data access layer (repository implementations)
│   │   ├── data.go               # ✅ DB, Redis, MinIO init + WarmUpCache trigger + ensureAdmin() + MinIO public-read bucket policy
│   │   ├── cache_warmup.go       # ✅ WarmUpCache (boot-time Redis population from MySQL)
│   │   ├── video_cache.go        # ✅ VideoCache (tag SETs + video HASHes + view buffer)
│   │   ├── cleanup_worker.go     # ✅ Background workers (view flush + cleanup retry)
│   │   ├── model/                # GORM model definitions (all 10 models)
│   │   │   ├── user.go
│   │   │   ├── video.go
│   │   │   ├── channel.go
│   │   │   ├── category.go
│   │   │   ├── tag.go            # Tag + VideoTag + UserTagPreference
│   │   │   ├── donation.go       # Donation model (schema ready)
│   │   │   ├── notification.go   # Notification model (schema ready)
│   │   │   ├── membership.go
│   │   │   └── view_record.go
│   │   ├── auth.go               # ✅ AuthRepo implementation
│   │   ├── video.go              # ✅ VideoRepo implementation
│   │   ├── channel.go            # ✅ ChannelRepo implementation
│   │   ├── category.go           # ✅ CategoryRepo implementation
│   │   ├── tag.go                # ✅ TagRepo implementation
│   │   ├── search.go             # ✅ SearchRepo implementation
│   │   └── admin.go              # ✅ AdminRepo implementation
│   │
│   ├── server/                   # Transport layer (HTTP & gRPC servers)
│   │   ├── server.go             # ✅ Server initialization (ProviderSet)
│   │   ├── http.go               # ✅ HTTP server with middleware + file upload routes
│   │   ├── grpc.go               # ✅ gRPC server
│   │   └── middleware.go         # ✅ JWT auth, admin guard, CORS
│   │
│   ├── service/                  # Service layer (API handler / adapter)
│   │   ├── service.go            # ✅ Service initialization (ProviderSet)
│   │   ├── auth.go               # ✅ AuthService (proto → biz)
│   │   ├── video.go              # ✅ VideoService
│   │   ├── channel.go            # ✅ ChannelService
│   │   ├── category.go           # ✅ CategoryService
│   │   ├── tag.go                # ✅ TagService
│   │   ├── search.go             # ✅ SearchService
│   │   └── admin.go              # ✅ AdminService
│   │
│   └── pkg/                      # Internal shared utilities
│       ├── authctx/
│       │   └── authctx.go        # ✅ UserIDFromContext, RoleFromContext helpers
│       ├── jwt/
│       │   └── jwt.go            # ✅ JWT token generation & validation
│       ├── hash/
│       │   └── hash.go           # ✅ Password hashing (bcrypt)
│       ├── upload/
│       │   └── minio.go          # ✅ MinIO file upload client
│       └── pagination/
│           └── pagination.go     # ✅ Pagination helper
│
├── third_party/                  # Third-party proto files
│   ├── errors/
│   │   └── v1/
│   │       └── errors.proto
│   ├── google/
│   │   ├── api/
│   │   │   ├── annotations.proto
│   │   │   ├── client.proto
│   │   │   ├── field_behavior.proto
│   │   │   ├── http.proto
│   │   │   └── httpbody.proto
│   │   └── protobuf/
│   │       ├── any.proto
│   │       ├── descriptor.proto
│   │       ├── duration.proto
│   │       ├── empty.proto
│   │       ├── timestamp.proto
│   │       └── wrappers.proto
│   ├── openapi/v3/
│   │   ├── annotations.proto
│   │   └── openapi.proto
│   └── validate/
│       └── validate.proto
│
├── scripts/
│   └── init.sql                  # Database initialization
├── buf.yaml                      # Proto build roots for IDE support
├── .dockerignore                 # Docker build excludes
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
└── openapi.yaml                  # Generated OpenAPI docs
```

---

## Layered Architecture

Kratos follows a **clean architecture** pattern with clear layer separation:

```
┌──────────────────────────────────────────────────────────────┐
│                      Transport Layer                         │
│               (HTTP Server / gRPC Server)                    │
│           internal/server/http.go, grpc.go                   │
├──────────────────────────────────────────────────────────────┤
│                      Service Layer                           │
│            (Request/Response mapping, DTO ↔ Domain)          │
│               internal/service/*.go                          │
├──────────────────────────────────────────────────────────────┤
│                    Business Logic Layer                       │
│            (Use cases, domain rules, interfaces)             │
│                 internal/biz/*.go                             │
├──────────────────────────────────────────────────────────────┤
│                      Data Layer                              │
│         (GORM repositories, Redis, MinIO, NATS client)       │
│                internal/data/*.go                            │
├──────────────────────────────────────────────────────────────┤
│                    Infrastructure                            │
│              (MySQL, Redis/Valkey, MinIO, NATS)              │
└──────────────────────────────────────────────────────────────┘
```

**Dependency direction**: Transport → Service → Biz ← Data

The `biz` layer defines **repository interfaces**; the `data` layer **implements** them (Dependency Inversion).

---

## API Design (Protobuf + HTTP)

### Auth Service

```protobuf
service AuthService {
  rpc Login (LoginRequest) returns (LoginReply) {
    option (google.api.http) = {
      post: "/api/v1/auth/login"
      body: "*"
    };
  }
  rpc Register (RegisterRequest) returns (RegisterReply) {
    option (google.api.http) = {
      post: "/api/v1/auth/register"
      body: "*"
    };
  }
  rpc RefreshToken (RefreshTokenRequest) returns (RefreshTokenReply) {
    option (google.api.http) = {
      post: "/api/v1/auth/refresh"
      body: "*"
    };
  }
}
```

### Video Service

```protobuf
service VideoService {
  rpc GetRecommended (GetRecommendedRequest) returns (VideoListReply) {
    option (google.api.http) = { get: "/api/v1/recommended" };
  }
  rpc GetVideo (GetVideoRequest) returns (VideoReply) {
    option (google.api.http) = { get: "/api/v1/videos/{id}" };
  }
  rpc CreateVideo (CreateVideoRequest) returns (VideoReply) {
    option (google.api.http) = {
      post: "/api/v1/videos"
      body: "*"
    };
  }
  rpc UpdateVideo (UpdateVideoRequest) returns (VideoReply) {
    option (google.api.http) = {
      put: "/api/v1/videos/{id}"
      body: "*"
    };
  }
  rpc DeleteVideo (DeleteVideoRequest) returns (DeleteVideoReply) {
    option (google.api.http) = { delete: "/api/v1/videos/{id}" };
  }
  rpc TogglePublish (TogglePublishRequest) returns (VideoReply) {
    option (google.api.http) = {
      patch: "/api/v1/videos/{id}/publish"
      body: "*"
    };
  }
}
```

### Channel Service

```protobuf
service ChannelService {
  rpc GetChannel (GetChannelRequest) returns (ChannelReply) {
    option (google.api.http) = { get: "/api/v1/channels/{id}" };
  }
  // Tier 1: Free subscribe to channel
  rpc Subscribe (SubscribeRequest) returns (MembershipReply) {
    option (google.api.http) = {
      post: "/api/v1/channels/{id}/subscribe"
      body: "*"
    };
  }
  // Unsubscribe from channel (any tier)
  rpc Unsubscribe (UnsubscribeRequest) returns (MembershipReply) {
    option (google.api.http) = { delete: "/api/v1/channels/{id}/subscribe" };
  }
  // 🔜 Planned (Phase 4): UpgradeToPremium, CancelPremium
}

message ChannelReply {
  int64 id = 1;
  int64 user_id = 2;
  string display_name = 3;
  string avatar_url = 4;
  double monthly_fee = 5;
  int64 subscriber_count = 6;
  string membership_status = 7;   // viewer's status: "none" / "subscribed" / "premium"
}
```

### Search Service

```protobuf
service SearchService {
  rpc Search (SearchRequest) returns (VideoListReply) {
    option (google.api.http) = { get: "/api/v1/search" };
  }
}

message SearchRequest {
  string query = 1;
  optional int64 category_id = 2;
  optional int32 min_duration = 3;   // seconds
  optional int32 max_duration = 4;
  optional string upload_date_from = 5;
  optional string upload_date_to = 6;
  optional string view_sort = 7;     // "asc" | "desc"
  optional string access_type = 8;   // "public" | "member"
  int32 page = 9;
  int32 page_size = 10;
}
```

### Dashboard Service (🔜 Planned — Phase 4)

```protobuf
// NOT YET IMPLEMENTED — planned for Phase 4 (Monetization)
service DashboardService {
  rpc GetMyVideos (GetMyVideosRequest) returns (VideoListReply) {
    option (google.api.http) = { get: "/api/v1/dashboard/videos" };
  }
  rpc GetAnalytics (GetAnalyticsRequest) returns (AnalyticsReply) {
    option (google.api.http) = { get: "/api/v1/dashboard/analytics" };
  }
  rpc SetMembershipFee (SetFeeRequest) returns (SetFeeReply) {
    option (google.api.http) = {
      put: "/api/v1/dashboard/fee"
      body: "*"
    };
  }
}
```

### User Service (🔜 Planned — Phase 5)

```protobuf
// NOT YET IMPLEMENTED — planned for Phase 5 (Advanced Features)
service UserService {
  rpc UpdateDisplayName (UpdateDisplayNameRequest) returns (UserReply) { ... }
  rpc UpdatePassword (UpdatePasswordRequest) returns (UpdatePasswordReply) { ... }
  rpc HideAccount (HideAccountRequest) returns (HideAccountReply) { ... }
  rpc DeleteAccount (DeleteAccountRequest) returns (DeleteAccountReply) { ... }
  rpc DeleteChannel (DeleteChannelRequest) returns (DeleteChannelReply) { ... }
}
```

### Tag Service

```protobuf
service TagService {
  // List all available tags
  rpc ListTags (ListTagsRequest) returns (TagListReply) {
    option (google.api.http) = { get: "/api/v1/tags" };
  }
  // Get user's selected tags (or guest's via session_id)
  rpc GetMyTags (GetMyTagsRequest) returns (TagListReply) {
    option (google.api.http) = { get: "/api/v1/tags/my" };
  }
  // Set user's tag preferences (max 5 tags)
  rpc SetMyTags (SetMyTagsRequest) returns (TagListReply) {
    option (google.api.http) = {
      put: "/api/v1/tags/my"
      body: "*"
    };
  }
}

message SetMyTagsRequest {
  repeated int64 tag_ids = 1;      // max 5 tag IDs
  optional string session_id = 2;  // for guest users
}

message TagListReply {
  repeated TagItem tags = 1;
}

message TagItem {
  int64 id = 1;
  string name = 2;
  string slug = 3;
}
```

### Donation Service (🔜 Planned — Phase 4)

```protobuf
// NOT YET IMPLEMENTED — planned for Phase 4 (Monetization)
// Donations are video-level (impulse purchase model).
service DonationService {
  rpc CreateDonation (CreateDonationRequest) returns (CreateDonationReply) { ... }
  rpc GetMyDonations (GetMyDonationsRequest) returns (DonationListReply) { ... }
  rpc GetReceivedDonations (GetReceivedDonationsRequest) returns (DonationListReply) { ... }
  rpc HandleWebhook (PaddleWebhookRequest) returns (PaddleWebhookReply) { ... }
}
```

### Notification Service (🔜 Planned — Phase 5)

```protobuf
// NOT YET IMPLEMENTED — planned for Phase 5 (Advanced Features)
// NATS-driven fan-out notifications to channel subscribers.
service NotificationService {
  rpc ListNotifications (ListNotificationsRequest) returns (NotificationListReply) { ... }
  rpc GetUnreadCount (GetUnreadCountRequest) returns (UnreadCountReply) { ... }
  rpc MarkRead (MarkReadRequest) returns (MarkReadReply) { ... }
  rpc MarkAllRead (MarkAllReadRequest) returns (MarkAllReadReply) { ... }
}
```

### Admin Service ✅

```protobuf
service AdminService {
  // User management
  rpc AdminListUsers (AdminListUsersRequest) returns (AdminListUsersReply) {
    option (google.api.http) = { get: "/api/v1/admin/users" };
  }
  rpc AdminDeleteUser (AdminDeleteUserRequest) returns (AdminDeleteUserReply) {
    option (google.api.http) = { delete: "/api/v1/admin/users/{id}" };
  }
  // Video management
  rpc AdminListVideos (AdminListVideosRequest) returns (AdminListVideosReply) {
    option (google.api.http) = { get: "/api/v1/admin/videos" };
  }
  rpc AdminDeleteVideo (AdminDeleteVideoRequest) returns (AdminDeleteVideoReply) {
    option (google.api.http) = { delete: "/api/v1/admin/videos/{id}" };
  }
  // Tag management
  rpc AdminCreateTag (AdminCreateTagRequest) returns (AdminCreateTagReply) {
    option (google.api.http) = {
      post: "/api/v1/admin/tags"
      body: "*"
    };
  }
  rpc AdminUpdateTag (AdminUpdateTagRequest) returns (AdminUpdateTagReply) {
    option (google.api.http) = {
      put: "/api/v1/admin/tags/{id}"
      body: "*"
    };
  }
  rpc AdminDeleteTag (AdminDeleteTagRequest) returns (AdminDeleteTagReply) {
    option (google.api.http) = { delete: "/api/v1/admin/tags/{id}" };
  }
}

message AdminUserInfo {
  uint64 id = 1;
  string username = 2;
  string display_name = 3;
  string role = 4;
  bool is_hidden = 5;
  string created_at = 6;
}

message AdminVideoInfo {
  uint64 id = 1;
  string title = 2;
  string username = 3;
  uint64 user_id = 4;
  string category_name = 5;
  int32 access_tier = 6;
  bool is_published = 7;
  bool is_hidden = 8;
  uint64 views_member = 9;
  uint64 views_non_member = 10;
  string created_at = 11;
}

message AdminTagInfo {
  uint64 id = 1;
  string name = 2;
  string slug = 3;
}
```

---

## Middleware

### JWT Authentication Middleware

```go
func JWTAuthMiddleware(jwtSecret string) middleware.Middleware {
    return func(handler middleware.Handler) middleware.Handler {
        return func(ctx context.Context, req interface{}) (interface{}, error) {
            // Extract token from Authorization header
            // Validate JWT, extract user_id and role
            // Inject user info into context:
            //   ctx = context.WithValue(ctx, "user_id", claims.UserID)
            //   ctx = context.WithValue(ctx, "role", claims.Role)  // "user" | "admin"
            // Check is_hidden: if user.is_hidden == true, reject with ACCOUNT_HIDDEN
            // Allow public routes to pass through
        }
    }
}
```

### Admin Guard Middleware

```go
func AdminGuardMiddleware() middleware.Middleware {
    return func(handler middleware.Handler) middleware.Handler {
        return func(ctx context.Context, req interface{}) (interface{}, error) {
            role := ctx.Value("role").(string)
            if role != "admin" {
                return nil, ErrAdminRequired  // ErrorReason.ADMIN_REQUIRED
            }
            return handler(ctx, req)
        }
    }
}
```

> Applied to all `/admin/**` routes. Requires `JWTAuthMiddleware` to run first so that `role` is available in context.

### Public vs Protected Routes

| Route Pattern                     | Auth Required                            | Status |
| --------------------------------- | ---------------------------------------- | ------ |
| `POST /auth/login`                | No                                       | ✅     |
| `POST /auth/register`             | No                                       | ✅     |
| `POST /auth/refresh`              | No                                       | ✅     |
| `GET /recommended`                | No (uses tags from session or user)      | ✅     |
| `GET /videos/:id`                 | No (member-only videos check membership) | ✅     |
| `POST /videos`                    | **Yes**                                  | ✅     |
| `PUT /videos/:id`                 | **Yes** (owner only)                     | ✅     |
| `DELETE /videos/:id`              | **Yes** (owner only)                     | ✅     |
| `PATCH /videos/:id/publish`       | **Yes** (owner only)                     | ✅     |
| `POST /upload/video`              | **Yes** (multipart, max 500MB)           | ✅     |
| `POST /upload/thumbnail`          | **Yes** (multipart, max 10MB)            | ✅     |
| `GET /search`                     | No                                       | ✅     |
| `GET /categories`                 | No                                       | ✅     |
| `GET /channels/:id`               | No (optional auth for membership status) | ✅     |
| `POST /channels/:id/subscribe`    | **Yes**                                  | ✅     |
| `DELETE /channels/:id/subscribe`  | **Yes**                                  | ✅     |
| `GET /tags`                       | No                                       | ✅     |
| `GET /tags/my`                    | No (guest uses session_id query param)   | ✅     |
| `PUT /tags/my`                    | No (guest uses session_id in body)       | ✅     |
| `GET /admin/users`                | **Yes** (admin role only)                | ✅     |
| `DELETE /admin/users/:id`         | **Yes** (admin role only)                | ✅     |
| `GET /admin/videos`               | **Yes** (admin role only)                | ✅     |
| `DELETE /admin/videos/:id`        | **Yes** (admin role only)                | ✅     |
| `POST /admin/tags`                | **Yes** (admin role only)                | ✅     |
| `PUT /admin/tags/:id`             | **Yes** (admin role only)                | ✅     |
| `DELETE /admin/tags/:id`          | **Yes** (admin role only)                | ✅     |

> All routes above are prefixed with `/api/v1`.
>
> **Planned routes (not yet implemented):** `/dashboard/**`, `/user/**`, `/videos/:id/donate`, `/donations/**`, `/webhooks/paddle`, `/notifications/**`, `/channels/:id/premium`

---

## Business Logic (biz layer)

### Key Interfaces

```go
// biz/video.go
type VideoRepo interface {
    Create(ctx context.Context, video *Video) (*Video, error)
    Update(ctx context.Context, video *Video) (*Video, error)
    Delete(ctx context.Context, id int64) error
    FindByID(ctx context.Context, id int64) (*Video, error)
    ListByTags(ctx context.Context, tagIDs []int64, page, pageSize int) ([]*Video, int64, error)
    ListRandom(ctx context.Context, page, pageSize int) ([]*Video, int64, error)
    ListByChannel(ctx context.Context, channelID int64, page, pageSize int) ([]*Video, int64, error)
    ListByCategory(ctx context.Context, categoryID int64, page, pageSize int) ([]*Video, int64, error)
    Search(ctx context.Context, params *SearchParams) ([]*Video, int64, error)
    IncrementViews(ctx context.Context, id int64, isMember bool) error
    TogglePublish(ctx context.Context, id int64, published bool) error
    Hide(ctx context.Context, id int64, hidden bool) error
}

type VideoUsecase struct {
    repo     VideoRepo
    tagRepo  TagRepo
    minio    *MinIOClient   // MinIO upload client
    nats     *NATSClient    // NATS pub/sub for notifications
    log      *log.Helper
}

func (uc *VideoUsecase) CreateVideo(ctx context.Context, v *Video, file io.Reader) (*Video, error) {
    // 1. Upload video file to MinIO
    // 2. Save video metadata to DB
    // 3. Associate tags with video (video_tags)
    // 4. Publish NATS event "channel.<channelID>.new_video" to notify subscribers
    // 5. Return created video
}

func (uc *VideoUsecase) DeleteVideo(ctx context.Context, userID, videoID int64) error {
    // 1. Verify ownership
    // 2. Check if video is unpublished (下架)
    // 3. Delete from MinIO storage
    // 4. Delete from DB (sets deleted_at)
}
```

```go
// biz/tag.go
type TagRepo interface {
    ListAll(ctx context.Context) ([]*Tag, error)
    GetByID(ctx context.Context, id int64) (*Tag, error)
    Create(ctx context.Context, tag *Tag) (*Tag, error)
    Update(ctx context.Context, tag *Tag) (*Tag, error)
    Delete(ctx context.Context, id int64) error
    GetUserTags(ctx context.Context, userID *int64, sessionID *string) ([]*Tag, error)
    SetUserTags(ctx context.Context, userID *int64, sessionID *string, tagIDs []int64) error
}

type TagUsecase struct {
    repo     TagRepo
    log      *log.Helper
}

// GetRecommendedVideos - Tag-based recommendation with random combination
func (uc *TagUsecase) GetRecommendedTagIDs(ctx context.Context, userID *int64, sessionID *string) ([]int64, error) {
    // 1. Get user's selected tags (max 5)
    tags, _ := uc.repo.GetUserTags(ctx, userID, sessionID)
    if len(tags) == 0 {
        return nil, nil  // fallback to random videos
    }

    // 2. Randomly pick a combination size (1 to len(tags))
    n := rand.Intn(len(tags)) + 1

    // 3. Shuffle and take first n tags
    rand.Shuffle(len(tags), func(i, j int) { tags[i], tags[j] = tags[j], tags[i] })
    selectedIDs := make([]int64, n)
    for i := 0; i < n; i++ {
        selectedIDs[i] = tags[i].ID
    }

    return selectedIDs, nil
}
```

```go
// biz/channel.go
type ChannelRepo interface {
    FindByID(ctx context.Context, id uint64) (*Channel, error)
    FindByUserID(ctx context.Context, userID uint64) (*Channel, error)
    GetSubscriberCount(ctx context.Context, channelID uint64) (int64, error)
    GetMembership(ctx context.Context, userID, channelID uint64) (*Membership, error)
    Subscribe(ctx context.Context, userID, channelID uint64) error          // Tier 1 free
    Unsubscribe(ctx context.Context, userID, channelID uint64) error
    HasMembership(ctx context.Context, userID, channelOwnerUserID uint64) (int8, error) // MembershipChecker
}

type ChannelUsecase struct {
    repo ChannelRepo
    log  *log.Helper
}

func (uc *ChannelUsecase) GetChannel(ctx, channelID, viewerID) (*Channel, string, error) {
    // 1. Look up channel
    // 2. Get subscriber count
    // 3. Determine membership status: "none" / "subscribed" / "premium"
}

func (uc *ChannelUsecase) Subscribe(ctx, userID, channelID) error {
    // 1. Prevent self-subscription
    // 2. Prevent duplicate subscriptions
    // 3. Create membership tier=1, status="active"
}

func (uc *ChannelUsecase) Unsubscribe(ctx, userID, channelID) error {
    // 1. Validate subscription exists
    // 2. Remove membership
}
// 🔜 Planned: UpgradeToPremium, CancelPremium (Phase 4)
```

```go
// biz/admin.go
type AdminRepo interface {
    ListUsers(ctx context.Context, offset, limit int) ([]*AdminUser, int64, error)
    FindUserByID(ctx context.Context, id uint64) (*AdminUser, error)
    DeleteUser(ctx context.Context, id uint64) error      // cascade delete (transactional)
    ListAllVideos(ctx context.Context, offset, limit int) ([]*AdminVideo, int64, error)
    DeleteVideo(ctx context.Context, id uint64) error     // hard delete (transactional)
    CreateTag(ctx context.Context, tag *AdminTag) (*AdminTag, error)
    UpdateTag(ctx context.Context, tag *AdminTag) (*AdminTag, error)
    DeleteTag(ctx context.Context, id uint64) error       // cascade delete video_tags + preferences
    FindTagByID(ctx context.Context, id uint64) (*AdminTag, error)
    FindTagByName(ctx context.Context, name string) (*AdminTag, error)
}

type AdminUsecase struct {
    repo AdminRepo
    log  *log.Helper
}

func (uc *AdminUsecase) DeleteUser(ctx, callerID, targetID) error {
    // 1. Prevent admin self-deletion
    // 2. Cascade delete: memberships, tag preferences, view records,
    //    notifications, donations, videos, channel (transactional)
}
```

### Planned Use Cases (Not Yet Implemented)

The following use cases are designed but not yet implemented:

- **UserUsecase** (Phase 5) — HideAccount, DeleteAccount, DeleteChannel (user self-service)
- **DonationUsecase** (Phase 4) — CreateDonation, HandlePaddleWebhook (Paddle integration)
- **NotificationUsecase** (Phase 5) — NATS-driven fan-out notifications to channel subscribers

---

## Video Recommendation Logic (Tag-Based)

```go
func (uc *VideoUsecase) GetRecommended(ctx context.Context, userID *int64, sessionID *string, page, pageSize int) ([]*Video, int64, error) {
    // 1. Get user's tag combination via TagUsecase
    tagIDs, err := uc.tagUsecase.GetRecommendedTagIDs(ctx, userID, sessionID)
    if err != nil || len(tagIDs) == 0 {
        // Fallback: no tags selected → random published videos
        return uc.repo.ListRandom(ctx, page, pageSize)
    }

    // 2. Query videos matching any of the randomly selected tag subset
    //    The tag combination changes on each request for variety
    return uc.repo.ListByTags(ctx, tagIDs, page, pageSize)
}
```

**Recommendation algorithm:**

1. User selects up to **5 tags** (stored in `user_tag_preferences`)
2. On each request, randomly pick **1 to N** tags from their selection
3. Fetch published, non-hidden videos matching **ANY** of those tags
4. Return in random order (`ORDER BY RAND()`)
5. If user has no tags → show globally random published videos

---

## Recommendation Cache (Redis)

The recommendation endpoint (`GET /api/v1/videos/recommended`) is the highest-traffic read path. A Redis cache layer eliminates MySQL queries for tag-based recommendations.

### Design Principles

- **Scope**: Tag-based recommendation only (not search, not premium content)
- **Cached content**: Only public (`access_tier = 0`), non-hidden, non-deleted videos
- **Premium videos**: Never cached — premium users always query MySQL
- **Key format**: Uses MySQL primary key `video.ID` as Redis key (stable, immutable, no extra lookup needed)
- **Cold start eliminated**: `WarmUpCache()` runs at app boot, loading all tag→video mappings into Redis before servers accept traffic

### Redis Data Structures

| Key Pattern | Type | TTL | Purpose |
|---|---|---|---|
| `tag:{id}` | SET | 24 hours | Video IDs belonging to this tag (index layer) |
| `video:{id}` | HASH | 24 hours | Video summary fields (data layer, one copy per video) |
| `popular:global` | ZSET | 10 min | Top videos scored by total view count |
| `views:buffer` | HASH | none | Buffered view count increments (flushed to MySQL every 30s) |
| `cleanup:queue` | LIST | none | Failed eviction job queue for cleanup worker |

### Architecture Diagram

```
┌──────────────────────────────────────────────────────────────────┐
│                        READ PATH                                  │
│                                                                    │
│  GET /api/v1/videos/recommended                                    │
│      │                                                              │
│      ▼                                                              │
│  Get user's tags (max 5) → randomly pick 1-N tags                  │
│      │                                                              │
│      ▼                                                              │
│  SUNION tag:{id1} tag:{id2} ... → merged video IDs                 │
│      │                                                              │
│      ├─ HIT: shuffle in Go, take page_size                         │
│      │       → MGET video:{id}... → return results                 │
│      │       → zero MySQL queries                                  │
│      │                                                              │
│      └─ MISS (empty SET): query MySQL → populate tag SET           │
│               + video HASHes → return results (lazy populate)      │
│                                                                    │
├──────────────────────────────────────────────────────────────────┤
│                        WRITE PATH                                  │
│                                                                    │
│  Video upload:                                                     │
│      → MySQL only (not cached). Lazy — waits for first read.      │
│      → 10-minute cooldown before creator can edit/delete.          │
│                                                                    │
│  Video edit (after cooldown):                                      │
│      → Update MySQL → evict from Redis (SREM tag SETs + DEL HASH) │
│                                                                    │
│  Video delete:                                                     │
│      → Collect tag IDs from MySQL (BEFORE delete)                  │
│      → Evict from Redis (SREM + DEL)                               │
│      → Hard delete from MySQL                                      │
│                                                                    │
│  Admin hide video:                                                 │
│      → MySQL update → evict from all tag SETs + DEL video HASH     │
│                                                                    │
├──────────────────────────────────────────────────────────────────┤
│                     VIEW COUNTER BUFFER                            │
│                                                                    │
│  User watches video:                                               │
│      → ZINCRBY popular:global 1 {video_id} (instant, in-memory)   │
│      → Every 30s: background goroutine flushes to MySQL            │
│        (batch UPDATE videos SET views_X = views_X + N)            │
│      → MySQL remains source of truth for durable storage           │
│                                                                    │
├──────────────────────────────────────────────────────────────────┤
│                     BOOT WARM-UP                                    │
│                                                                    │
│  App starts → NewData() → WarmUpCache():                           │
│      → Query all tags from MySQL                                   │
│      → For each tag: query public videos → SAdd tag:{id}           │
│      → For each video: HSet video:{id} (skip if already cached)    │
│      → Done: servers start, first user gets cache HIT              │
│                                                                    │
│  Recovery: if Redis restarts, app restart re-runs WarmUpCache()    │
│  Source of truth: MySQL tables already contain all data needed      │
│  No tracking table needed: videos + video_tags IS the blueprint    │
│                                                                    │
├──────────────────────────────────────────────────────────────────┤
│                     SAFETY NETS                                    │
│                                                                    │
│  TTL on all keys:        auto-expire catches any missed evictions  │
│  Cleanup worker:         retries failed evictions (see below)      │
│  10-min upload cooldown: prevents rapid cache churn after upload   │
│  Rate limit:             protects MySQL from cache-miss storms     │
│  Boot warm-up:           eliminates cold start entirely            │
└──────────────────────────────────────────────────────────────────┘
```

### Boot Warm-Up (Eliminates Cold Start)

```go
// internal/data/cache_warmup.go

func (d *Data) WarmUpCache(ctx context.Context, logger log.Logger) {
    // Called from NewData() before servers start accepting traffic.
    // Loads all public videos into Redis so the first user gets a cache HIT.

    var tags []model.Tag
    d.DB.Find(&tags)

    for _, tag := range tags {
        tagKey := fmt.Sprintf("tag:%d", tag.ID)

        // Query public, published, non-hidden videos for this tag
        var videoIDs []uint64
        d.DB.Table("video_tags").
            Select("video_tags.video_id").
            Joins("INNER JOIN videos ON videos.id = video_tags.video_id").
            Where("video_tags.tag_id = ?", tag.ID).
            Where("videos.is_published = ? AND videos.is_hidden = ? AND videos.deleted_at IS NULL", true, false).
            Where("videos.access_tier = 0").
            Pluck("video_id", &videoIDs)

        // Populate tag SET
        if len(videoIDs) > 0 {
            members := make([]interface{}, len(videoIDs))
            for i, id := range videoIDs { members[i] = id }
            d.Redis.SAdd(ctx, tagKey, members...)
            d.Redis.Expire(ctx, tagKey, cacheTagTTL)  // 24h
        }

        // Populate video HASHes (skip if already cached from another tag)
        for _, videoID := range videoIDs {
            videoKey := fmt.Sprintf("video:%d", videoID)
            if exists, _ := d.Redis.Exists(ctx, videoKey).Result(); exists > 0 {
                continue
            }
            var video model.Video
            d.DB.First(&video, videoID)
            d.Redis.HSet(ctx, videoKey, map[string]interface{}{
                "id": video.ID, "title": video.Title,
                "duration": video.Duration, "views": video.ViewsMember + video.ViewsNonMember,
                "thumbnail": video.ThumbnailURL, "category_id": video.CategoryID,
                "user_id": video.UserID, "video_url": video.VideoURL,
            })
            d.Redis.Expire(ctx, videoKey, cacheVideoTTL)  // 24h
        }
    }
}
```

> **TTL after warm-up**: Keys expire after 24 hours. After expiry, lazy populate handles individual cache misses. The warm-up covers the critical initial burst; lazy handles the steady state. Active videos have TTL refreshed on each view (see `IncrementViewsBuffered`).

### Cache Population (Lazy, On Cache Miss)

```go
// internal/data/video_cache.go

func (r *VideoCacheRepo) GetVideosByTag(ctx context.Context, tagID int64) ([]uint64, error) {
    key := fmt.Sprintf("tag:%d", tagID)

    // Try cache first
    ids, err := r.rdb.SMembers(ctx, key).Result()
    if err == nil && len(ids) > 0 {
        return parseIDs(ids), nil
    }

    // Cache miss → lazy populate from MySQL
    var videoIDs []uint64
    r.db.WithContext(ctx).
        Table("video_tags").
        Select("video_tags.video_id").
        Joins("INNER JOIN videos ON videos.id = video_tags.video_id").
        Where("video_tags.tag_id = ?", tagID).
        Where("videos.is_published = ? AND videos.is_hidden = ? AND videos.deleted_at IS NULL", true, false).
        Where("videos.access_tier = 0").  // public only
        Pluck("video_id", &videoIDs)

    // Populate Redis SET
    if len(videoIDs) > 0 {
        members := make([]interface{}, len(videoIDs))
        for i, id := range videoIDs {
            members[i] = id
        }
        r.rdb.SAdd(ctx, key, members...)
        r.rdb.Expire(ctx, key, cacheTagTTL)  // 24h
    }

    return videoIDs, nil
}
```

### Recommendation Read Path

```go
func (r *VideoCacheRepo) GetRecommendedFromCache(
    ctx context.Context, tagIDs []int64, pageSize int,
) ([]*VideoSummary, error) {
    // Collect candidate video IDs from each tag's SET
    tagKeys := make([]string, len(tagIDs))
    for i, id := range tagIDs {
        tagKeys[i] = fmt.Sprintf("tag:%d", id)
    }

    // SUNION: merge all video IDs across selected tags
    videoIDStrs, err := r.rdb.SUnion(ctx, tagKeys...).Result()
    if err != nil || len(videoIDStrs) == 0 {
        return nil, err // fallback to MySQL
    }

    // Random sample (replaces ORDER BY RAND())
    rand.Shuffle(len(videoIDStrs), func(i, j int) {
        videoIDStrs[i], videoIDStrs[j] = videoIDStrs[j], videoIDStrs[i]
    })
    if len(videoIDStrs) > pageSize {
        videoIDStrs = videoIDStrs[:pageSize]
    }

    // Batch fetch video details from HASH keys
    pipe := r.rdb.Pipeline()
    cmds := make([]*redis.MapStringStringCmd, len(videoIDStrs))
    for i, idStr := range videoIDStrs {
        cmds[i] = pipe.HGetAll(ctx, fmt.Sprintf("video:%s", idStr))
    }
    pipe.Exec(ctx)

    results := make([]*VideoSummary, 0, len(cmds))
    for _, cmd := range cmds {
        vals, err := cmd.Result()
        if err != nil || len(vals) == 0 {
            continue
        }
        results = append(results, parseVideoFromHash(vals))
    }
    return results, nil
}
```

### Cache Eviction (Application-Level Hook)

```go
// Called on video update or delete
func (r *VideoCacheRepo) Evict(ctx context.Context, videoID uint64, tagIDs []uint64) error {
    pipe := r.rdb.Pipeline()
    for _, tagID := range tagIDs {
        pipe.SRem(ctx, fmt.Sprintf("tag:%d", tagID), videoID)
    }
    pipe.Del(ctx, fmt.Sprintf("video:%d", videoID))
    pipe.ZRem(ctx, "popular:global", videoID)
    _, err := pipe.Exec(ctx)
    return err
}
```

### Account Deletion with Cache Cleanup

Account deletion uses a **collect-before-delete** pattern with a background cleanup worker for failure recovery.

```go
func (uc *UserUsecase) DeleteAccount(ctx context.Context, userID uint64) error {
    // Step 1: Collect video IDs + tag IDs BEFORE deleting anything
    videos, _ := uc.videoRepo.ListByUser(ctx, userID)
    videoTagMap := make(map[uint64][]uint64)
    for _, v := range videos {
        tagIDs, _ := uc.tagRepo.GetTagIDsByVideo(ctx, v.ID)
        videoTagMap[v.ID] = tagIDs
    }

    // Step 2: Best-effort evict from Redis
    var failedIDs []uint64
    for videoID, tagIDs := range videoTagMap {
        if err := uc.cache.Evict(ctx, videoID, tagIDs); err != nil {
            failedIDs = append(failedIDs, videoID)
        }
    }

    // Step 3: Record failed IDs for cleanup worker (if any evictions failed)
    if len(failedIDs) > 0 {
        job := CleanupJob{VideoIDs: failedIDs, TagMap: videoTagMap}
        uc.cleanupQueue.Enqueue(ctx, job)
    }

    // Step 4: Hard delete from MySQL (regardless of Redis result)
    uc.videoRepo.HardDeleteByUser(ctx, userID)
    uc.channelRepo.HardDelete(ctx, userID)
    uc.userRepo.HardDelete(ctx, userID)

    return nil
}
```

### Cleanup Worker (Background Recovery)

The cleanup worker processes failed cache evictions. It retries until Redis is reachable, then removes orphaned entries.

```go
// internal/data/cleanup_worker.go

type CleanupJob struct {
    VideoIDs []uint64            `json:"video_ids"`
    TagMap   map[uint64][]uint64 `json:"tag_map"` // videoID → tagIDs
}

type CleanupWorker struct {
    rdb   *redis.Client
    queue string // Redis LIST key: "cleanup:queue"
    log   *log.Helper
}

// Start runs the cleanup worker in a background goroutine.
// It processes failed cache evictions by retrying until Redis recovers.
func (w *CleanupWorker) Start(ctx context.Context) {
    for {
        select {
        case <-ctx.Done():
            return
        default:
            job, err := w.dequeue(ctx)
            if err != nil || job == nil {
                time.Sleep(5 * time.Second)
                continue
            }

            for _, videoID := range job.VideoIDs {
                tagIDs := job.TagMap[videoID]
                err := w.evict(ctx, videoID, tagIDs)
                if err != nil {
                    // Redis still down — re-enqueue entire job and wait
                    w.enqueue(ctx, job)
                    w.log.Warnf("cleanup worker: Redis unreachable, retrying in 10s")
                    time.Sleep(10 * time.Second)
                    break
                }
            }
        }
    }
}

func (w *CleanupWorker) evict(ctx context.Context, videoID uint64, tagIDs []uint64) error {
    pipe := w.rdb.Pipeline()
    for _, tagID := range tagIDs {
        pipe.SRem(ctx, fmt.Sprintf("tag:%d", tagID), videoID)
    }
    pipe.Del(ctx, fmt.Sprintf("video:%d", videoID))
    pipe.ZRem(ctx, "popular:global", videoID)
    _, err := pipe.Exec(ctx)
    return err
}
```

### View Count Buffer

```go
// Buffered in Redis, flushed to MySQL periodically
func (r *VideoCacheRepo) IncrementViewsCached(ctx context.Context, videoID uint64, isMember bool) error {
    // Instant write to Redis ZSET (for popular ranking)
    r.rdb.ZIncrBy(ctx, "popular:global", 1, fmt.Sprintf("%d", videoID))

    // Buffer the count for MySQL flush
    field := fmt.Sprintf("%d:non_member", videoID)
    if isMember {
        field = fmt.Sprintf("%d:member", videoID)
    }
    return r.rdb.HIncrBy(ctx, "views:buffer", field, 1).Err()
}

// FlushViewsToDB runs every 30 seconds via background goroutine
func (r *VideoCacheRepo) FlushViewsToDB(ctx context.Context) error {
    vals, err := r.rdb.HGetAll(ctx, "views:buffer").Result()
    if err != nil || len(vals) == 0 {
        return err
    }
    r.rdb.Del(ctx, "views:buffer")

    for field, countStr := range vals {
        var videoID uint64
        var col string
        if _, err := fmt.Sscanf(field, "%d:member", &videoID); err == nil {
            col = "views_member"
        } else if _, err := fmt.Sscanf(field, "%d:non_member", &videoID); err == nil {
            col = "views_non_member"
        } else {
            continue
        }
        count, _ := strconv.ParseInt(countStr, 10, 64)
        r.db.Table("videos").Where("id = ?", videoID).
            Update(col, gorm.Expr(col+" + ?", count))
    }
    return nil
}
```

### Cache Design Decisions

| Decision | Rationale |
|---|---|
| **Boot warm-up + lazy fallback** | WarmUpCache() at boot eliminates cold start; after TTL expiry, lazy populate handles individual misses |
| **Per-tag SET + per-video HASH (two layers)** | Index and data separated; each video stored once regardless of tag count |
| **MySQL primary key as Redis key** | Stable, immutable, no hash computation needed, matches API path (`/videos/{id}`) |
| **Public videos only in cache** | Premium content always hits MySQL; avoids complex per-user access filtering in cache |
| **Application-level eviction hooks** | Immediate consistency on update/delete; fires in biz layer alongside MySQL writes |
| **Collect-before-delete on account deletion** | Must gather video IDs + tag IDs before hard delete destroys the data |
| **Cleanup worker with retry queue** | Handles partial Redis failures during account deletion; retries until Redis recovers |
| **10-minute upload cooldown** | Prevents rapid edit/delete cycles that would thrash the cache |
| **View count buffer (Redis → MySQL)** | Eliminates per-view DB writes; MySQL updated in batches every 30s |
| **TTL safety net on all keys** | Catches any eviction that was missed; worst case = stale data for TTL duration |
| **Rate limit on cache-miss MySQL queries** | Protects MySQL from thundering herd on cold start |

---

## Video Access Control Logic

```go
func (uc *VideoUsecase) GetVideo(ctx context.Context, videoID int64, viewerID *int64, isAdmin bool) (*Video, error) {
    video, err := uc.repo.FindByID(ctx, videoID)
    if err != nil {
        return nil, err
    }

    // Hidden check: only admins and the owner can see hidden content
    if video.IsHidden {
        isOwner := viewerID != nil && *viewerID == video.UserID
        if !isAdmin && !isOwner {
            return nil, ErrVideoNotFound
        }
    }

    // Check if video is published
    if !video.IsPublished {
        // Only owner can see unpublished video
        if viewerID == nil || *viewerID != video.UserID {
            return nil, ErrVideoNotFound
        }
    }

    // Check if video requires subscription tier
    if video.AccessTier > 0 {
        if viewerID == nil {
            return nil, ErrMembershipRequired
        }
        membership, _ := uc.channelRepo.GetMembership(ctx, *viewerID, video.ChannelID)
        isOwner := *viewerID == video.UserID
        if !isOwner {
            if membership == nil {
                return nil, ErrMembershipRequired      // not subscribed at all
            }
            if video.AccessTier == 2 && membership.Tier < 2 {
                return nil, ErrPremiumRequired          // tier 2 video, but user is tier 1
            }
        }
    }

    // Increment views
    isMember := viewerID != nil  // simplified
    uc.repo.IncrementViews(ctx, videoID, isMember)

    return video, nil
}
```

> **Access control priority:** `is_hidden` → `is_published` → `access_tier` → allow.
>
> - `access_tier=0`: public, anyone can watch.
> - `access_tier=1`: subscriber-only (Tier 1 or Tier 2 members).
> - `access_tier=2`: premium-only (Tier 2 paid members only).
>   Hidden videos/channels/users are invisible to the public. Only admins (via admin panel) and owners can see their own hidden content.

---

## MinIO File Upload Client

```go
// internal/pkg/upload/minio.go
package upload

import (
    "context"
    "fmt"
    "io"
    "github.com/minio/minio-go/v7"
    "github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOClient struct {
    client     *minio.Client
    bucketName string
}

func NewMinIOClient(endpoint, accessKey, secretKey, bucket string, useSSL bool) (*MinIOClient, error) {
    client, err := minio.New(endpoint, &minio.Options{
        Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
        Secure: useSSL,
    })
    if err != nil {
        return nil, fmt.Errorf("failed to create minio client: %w", err)
    }
    return &MinIOClient{client: client, bucketName: bucket}, nil
}

func (m *MinIOClient) Upload(ctx context.Context, objectName string, reader io.Reader, size int64, contentType string) (string, error) {
    _, err := m.client.PutObject(ctx, m.bucketName, objectName, reader, size, minio.PutObjectOptions{
        ContentType: contentType,
    })
    if err != nil {
        return "", fmt.Errorf("failed to upload file: %w", err)
    }
    return fmt.Sprintf("/%s/%s", m.bucketName, objectName), nil
}

func (m *MinIOClient) Delete(ctx context.Context, objectName string) error {
    return m.client.RemoveObject(ctx, m.bucketName, objectName, minio.RemoveObjectOptions{})
}
```

---

## Paddle Payment Client (🔜 Planned — Phase 4)

```go
// internal/pkg/paddle/paddle.go
package paddle

import (
    "context"
    "fmt"

    paddle "github.com/PaddleHQ/paddle-go-sdk/v3"
    "github.com/PaddleHQ/paddle-go-sdk/v3/pkg/paddlenotification"
)

type PaddleClient struct {
    client *paddle.Client
    secret string // webhook secret for signature verification
}

// NewPaddleClient creates a Paddle SDK client pointed at the sandbox environment.
func NewPaddleClient(apiKey, webhookSecret string) (*PaddleClient, error) {
    client, err := paddle.New(
        apiKey,
        paddle.WithBaseURL(paddle.SandboxBaseURL), // https://sandbox-api.paddle.com
    )
    if err != nil {
        return nil, fmt.Errorf("failed to create paddle client: %w", err)
    }
    return &PaddleClient{client: client, secret: webhookSecret}, nil
}

// CreateDonationTransaction builds a one-time price on the fly and opens
// a Paddle checkout transaction.  Returns the Paddle transaction ID (txn_*)
// so the frontend can launch Paddle.js with it.
// Donations are video-level: the video_id is stored in custom_data so the
// webhook handler can trace which video triggered the donation.
func (p *PaddleClient) CreateDonationTransaction(
    ctx context.Context,
    amount string,      // e.g. "5.00"
    currency string,    // e.g. "USD"
    donorEmail string,
    donationID int64,   // our internal donation ID stored in custom_data
    videoID int64,      // the video that triggered this donation
) (string, error) {
    txn, err := p.client.CreateTransaction(ctx, &paddle.CreateTransactionRequest{
        Items: []paddle.CreateTransactionItems{{
            Price: paddle.NewCreateTransactionItemsPriceObject(&paddle.CreateTransactionItemsPriceObject{
                Description: "Donation",
                Name:        "Creator Donation",
                UnitPrice: paddle.Money{
                    Amount:     amount,
                    CurrencyCode: paddle.CurrencyCode(currency),
                },
                Product: paddle.CreateTransactionItemsPriceObjectProduct{
                    Name:        "Donation",
                    Description: ptrStr("One-time donation to creator"),
                    TaxCategory: "standard",
                },
                BillingCycle: nil, // one-time, no subscription
            }),
            Quantity: 1,
        }},
        CustomData: map[string]interface{}{
            "donation_id": donationID,
            "video_id":    videoID,
        },
    })
    if err != nil {
        return "", fmt.Errorf("paddle create transaction: %w", err)
    }
    return txn.ID, nil
}

// VerifyWebhookSignature validates the Paddle-Signature header and
// returns the parsed webhook event.
func (p *PaddleClient) VerifyWebhookSignature(rawBody []byte, signature string) (*paddlenotification.Event, error) {
    verifier := paddlenotification.NewWebhookVerifier(p.secret)
    event, err := verifier.Verify(rawBody, signature)
    if err != nil {
        return nil, fmt.Errorf("invalid paddle webhook signature: %w", err)
    }
    return event, nil
}

// CreatePremiumSubscription creates a Paddle recurring subscription for
// Tier 2 (premium) channel membership.  Returns the checkout URL and
// the Paddle subscription ID.
func (p *PaddleClient) CreatePremiumSubscription(
    ctx context.Context,
    priceAmount string,   // channel's monthly_fee as string, e.g. "9.99"
    currency string,
    userEmail string,
    channelID int64,
    membershipID int64,
) (checkoutURL string, err error) {
    txn, err := p.client.CreateTransaction(ctx, &paddle.CreateTransactionRequest{
        Items: []paddle.CreateTransactionItems{{
            Price: paddle.NewCreateTransactionItemsPriceObject(&paddle.CreateTransactionItemsPriceObject{
                Description: "Premium Membership",
                Name:        "Channel Premium Subscription",
                UnitPrice: paddle.Money{
                    Amount:       priceAmount,
                    CurrencyCode: paddle.CurrencyCode(currency),
                },
                Product: paddle.CreateTransactionItemsPriceObjectProduct{
                    Name:        "Premium Membership",
                    Description: ptrStr("Monthly premium channel subscription"),
                    TaxCategory: "standard",
                },
                BillingCycle: &paddle.Duration{
                    Interval:  paddle.IntervalMonth,
                    Frequency: 1,
                },
            }),
            Quantity: 1,
        }},
        CustomData: map[string]interface{}{
            "membership_id": membershipID,
            "channel_id":    channelID,
            "type":          "premium_subscription",
        },
    })
    if err != nil {
        return "", fmt.Errorf("paddle create subscription txn: %w", err)
    }
    return txn.ID, nil  // frontend opens checkout with this transaction ID
}

func ptrStr(s string) *string { return &s }
```

---

## NATS Pub/Sub Client (🔜 Planned — Phase 5)

```go
// internal/pkg/nats/nats.go
package natsutil

import (
    "context"
    "encoding/json"
    "fmt"

    "github.com/nats-io/nats.go"
)

type NATSClient struct {
    conn *nats.Conn
}

func NewNATSClient(url string) (*NATSClient, error) {
    nc, err := nats.Connect(url)
    if err != nil {
        return nil, fmt.Errorf("failed to connect to NATS: %w", err)
    }
    return &NATSClient{conn: nc}, nil
}

// ChannelEvent is published when a channel creates or updates a video.
type ChannelEvent struct {
    Type      string `json:"type"`       // "new_video" | "video_update"
    ChannelID int64  `json:"channel_id"`
    VideoID   int64  `json:"video_id"`
    Title     string `json:"title"`
    CreatorName string `json:"creator_name"`
}

// PublishChannelEvent publishes an event to "channel.<id>.<type>" subject.
func (c *NATSClient) PublishChannelEvent(ctx context.Context, event ChannelEvent) error {
    subject := fmt.Sprintf("channel.%d.%s", event.ChannelID, event.Type)
    data, err := json.Marshal(event)
    if err != nil {
        return err
    }
    return c.conn.Publish(subject, data)
}

// SubscribeChannel listens to all events for a channel (wildcard).
// Used by the notification background worker.
func (c *NATSClient) SubscribeChannel(handler func(event ChannelEvent)) (*nats.Subscription, error) {
    return c.conn.Subscribe("channel.>", func(msg *nats.Msg) {
        var event ChannelEvent
        if err := json.Unmarshal(msg.Data, &event); err != nil {
            return
        }
        handler(event)
    })
}

func (c *NATSClient) Close() {
    c.conn.Close()
}
```

---

## Configuration

```yaml
# configs/config.yaml
server:
  http:
    addr: 0.0.0.0:8000
    timeout: 1s
  grpc:
    addr: 0.0.0.0:9000
    timeout: 1s

data:
  database:
    driver: mysql
    source: root:root@tcp(127.0.0.1:3306)/fenzvideo?parseTime=True&loc=Local&charset=utf8mb4
    max_idle_conns: 10
    max_open_conns: 100
    conn_max_lifetime: 3600s
  redis:
    addr: 127.0.0.1:6379
    password: ""
    db: 0
    read_timeout: 0.2s
    write_timeout: 0.2s

auth:
  jwt_secret: "fenzvideo-dev-secret-change-in-production"
  token_expiry: 86400s    # 24 hours
  refresh_expiry: 604800s # 7 days

storage:
  endpoint: "127.0.0.1:9100"   # MinIO API port (mapped in docker-compose)
  access_key: "minioadmin"
  secret_key: "minioadmin"
  bucket: "fenzvideo"
  use_ssl: false
  region: "us-east-1"

paddle:
  api_key: ""              # Set via .env (PADDLE_KEY)
  webhook_secret: ""       # Set via .env
  sandbox: true

nats:
  url: "nats://127.0.0.1:4222"

admin:
  username: "admin"        # Admin account auto-created on boot
  password: "admin123"     # Update via .env (ADMIN_PASSWORD)
```

---

## Dependency Injection (Wire)

```go
// cmd/fenzvideo/wire.go
//go:build wireinject

package main

import (
    "github.com/google/wire"
    "fenzvideo/internal/biz"
    "fenzvideo/internal/conf"
    "fenzvideo/internal/data"
    "fenzvideo/internal/server"
    "fenzvideo/internal/service"
)

func wireApp(*conf.Server, *conf.Data, *conf.Auth, *conf.Storage, *conf.Paddle, *conf.NATS, *conf.Tracing, log.Logger) (*kratos.App, func(), error) {
    panic(wire.Build(
        data.ProviderSet,
        biz.ProviderSet,
        service.ProviderSet,
        server.ProviderSet,
        newApp,
    ))
}
```

---

## Docker Compose (All Open-Source Services)

```yaml
# docker-compose.yaml (at project root)
services:
  backend:
    build:
      context: ./backend
      dockerfile: Dockerfile
    container_name: fenzvideo-backend
    restart: unless-stopped
    ports:
      - "8000:8000"
      - "9000:9000"
    volumes:
      - ./backend/configs/config.docker.yaml:/data/conf/config.yaml
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_started
      minio:
        condition: service_started
      nats:
        condition: service_started

  frontend:
    build:
      context: ./frontend
      dockerfile: Dockerfile
    container_name: fenzvideo-frontend
    restart: unless-stopped
    ports:
      - "80:80"
    depends_on:
      - backend

  mysql:
    image: mysql:8.0
    container_name: fenzvideo-mysql
    restart: unless-stopped
    environment:
      MYSQL_ROOT_PASSWORD: root
      MYSQL_DATABASE: fenzvideo
      MYSQL_CHARSET: utf8mb4
      MYSQL_COLLATION: utf8mb4_unicode_ci
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./backend/scripts/init.sql:/docker-entrypoint-initdb.d/init.sql
    command: --default-authentication-plugin=mysql_native_password --character-set-server=utf8mb4 --collation-server=utf8mb4_unicode_ci
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost", "-u", "root", "-proot"]
      interval: 5s
      timeout: 3s
      retries: 10

  redis:
    image: redis:7-alpine
    container_name: fenzvideo-redis
    restart: unless-stopped
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data

  minio:
    image: minio/minio:latest
    container_name: fenzvideo-minio
    restart: unless-stopped
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    ports:
      - "9100:9000"   # API port (mapped to 9100 to avoid gRPC conflict)
      - "9101:9001"   # Console port
    volumes:
      - minio_data:/data
    command: server /data --console-address ":9001"

  nats:
    image: nats:2-alpine
    container_name: fenzvideo-nats
    restart: unless-stopped
    ports:
      - "4222:4222"   # Client connections
      - "8222:8222"   # HTTP monitoring
    command: --jetstream --http_port 8222

  jaeger:
    image: jaegertracing/all-in-one:latest
    container_name: fenzvideo-jaeger
    restart: unless-stopped
    environment:
      COLLECTOR_OTLP_ENABLED: "true"
    ports:
      - "16686:16686" # Jaeger UI
      - "4317:4317"   # OTLP gRPC
      - "4318:4318"   # OTLP HTTP

volumes:
  mysql_data:
  redis_data:
  minio_data:
```

> **Note**: All 7 services (backend, frontend, mysql, redis, minio, nats, jaeger) start together with `docker-compose up -d --build`. The backend waits for MySQL to be healthy before starting. The frontend Nginx container proxies `/api/` to the backend and `/fenzvideo/` to MinIO.

---

## Seed Data Generator

The seed script (`cmd/seed/main.go`) populates the database with sample data before starting services. It includes **57 pre-defined diverse videos** with curated titles and descriptions in Traditional Chinese. Optionally integrates with the **Gemini API** for additional creative content.

### Usage

```bash
# GEMINI_KEY optional — 57 pre-defined videos work without it
cd backend && make seed
```

### What Gets Seeded

| Entity | Count | Details |
|--------|-------|---------|
| Admin user | 1 | `admin` / `admin123` (role: admin) |
| Creator users | 5 | `creator_alice` through `creator_emma` (role: user, password: `password123`) |
| Channels | 6 | One per user (admin + 5 creators), creators have random monthly fee ($1–$10) |
| Categories | 10 | 音樂, 遊戲, 教育, 娛樂, 科技, 運動, 新聞, 美食, 旅遊, 生活 |
| Tags | 15 | 搞笑, 教學, Vlog, 開箱, 直播精華, 音樂MV, 遊戲實況, 美食料理, 旅行紀錄, 科技評測, 新手入門, 健身運動, 動畫, 訪談, DIY手作 |
| Videos | 57 | 2-3 curated tags per video, pre-defined titles & descriptions, random views (member 0-5K, non-member 0-10K), duration 60-660s, all public (`access_tier=0`) |
| Video tags | ~140 | 2-3 tags per video (many-to-many). Deliberate distribution: 教學(19), Vlog(18), 搞笑(13), 新手入門(11), 科技評測(10), DIY手作(9), etc. |
| MinIO videos | 20 | Sample video files downloaded and uploaded to `fenzvideo/videos/` |
| MinIO thumbnails | 20 | Placeholder thumbnails generated and uploaded to `fenzvideo/thumbnails/` |

### Key Behaviors

- **Idempotent**: Checks for existing data before inserting; safe to run multiple times
- **57 pre-defined videos**: Curated content covering all 10 categories and 15 tags; no external API needed
- **Optional Gemini API**: If `GEMINI_KEY` is set, can generate additional creative content
- **MinIO upload**: Downloads sample videos from the internet, uploads to MinIO storage bucket
- **Thumbnail generation**: Creates placeholder JPEG thumbnails for each video slot, uploads to MinIO
- **Thumbnail URL**: Sets `thumbnail_url` on all video records so frontend cards display properly
- **DB connection**: Defaults to `root:root@tcp(127.0.0.1:3306)/fenzvideo`, overridable via `DB_DSN` env var
- **Auto-migrate**: Runs GORM AutoMigrate before seeding (creates tables if not exist)

### Seed Flow

```
1. Connect MySQL → AutoMigrate all tables
2. Seed admin user + channel (skip if exists)
3. Seed 10 categories (skip if any exist)
4. Seed 15 tags (skip if any exist)
5. Seed 5 creator users + channels (skip if exist)
6. Seed 57 videos from pre-defined list:
   → Each video has curated title, description, and 2-3 specific tags
   → Round-robin across creators & categories
   → Associate video ↔ tags in video_tags
7. Upload sample video files to MinIO (fenzvideo/videos/)
8. Generate & upload thumbnail images to MinIO (fenzvideo/thumbnails/)
9. Update all video records with thumbnail_url
10. Done — all data ready for services
```

---

## Error Handling

Kratos uses protobuf-defined error reasons:

```protobuf
// api/fenzvideo/v1/error_reason.proto
enum ErrorReason {
  UNKNOWN_ERROR = 0;
  USER_NOT_FOUND = 1;
  INVALID_CREDENTIALS = 2;
  TOKEN_EXPIRED = 3;
  TOKEN_INVALID = 4;
  VIDEO_NOT_FOUND = 5;
  VIDEO_NOT_UNPUBLISHED = 6;      // Cannot delete published video
  CHANNEL_NOT_FOUND = 7;
  ALREADY_MEMBER = 8;
  NOT_MEMBER = 9;
  MEMBERSHIP_REQUIRED = 10;
  PERMISSION_DENIED = 11;
  INVALID_PARAMS = 12;
  FILE_TOO_LARGE = 13;
  UNSUPPORTED_FORMAT = 14;
  ADMIN_REQUIRED = 15;            // Admin role required
  USER_HIDDEN = 16;               // Account is hidden
  TAG_NOT_FOUND = 17;
  TAG_LIMIT_EXCEEDED = 18;        // Max 5 tags allowed
  CANNOT_DELETE_SELF = 19;        // Admin cannot delete own account
  CHANNEL_ALREADY_DELETED = 20;
  DONATION_NOT_FOUND = 21;
  CREATOR_NOT_FOUND = 22;         // Donation target has no channel
  PADDLE_ERROR = 23;              // Paddle API call failed
  CANNOT_DONATE_SELF = 24;        // Cannot donate to own video
  INVALID_DONATION_AMOUNT = 25;   // Amount <= 0 or unsupported currency
  NOT_SUBSCRIBED = 26;             // User is not subscribed to this channel
  ALREADY_SUBSCRIBED = 27;         // User already subscribed
  PREMIUM_REQUIRED = 28;           // Video requires Tier 2 premium subscription
  ALREADY_PREMIUM = 29;            // Already a premium member
  NOT_PREMIUM = 30;                // Not a premium member (cannot cancel)
}
```

---

## Makefile

```makefile
.PHONY: init api build run test docker

# Generate protobuf code
api:
	protoc --proto_path=./api \
		--proto_path=./third_party \
		--go_out=paths=source_relative:./api \
		--go-http_out=paths=source_relative:./api \
		--go-grpc_out=paths=source_relative:./api \
		--openapiv2_out=./api \
		./api/fenzvideo/v1/*.proto

# Generate wire injection
wire:
	cd cmd/fenzvideo && wire

# Build binary
build:
	go build -o ./bin/fenzvideo ./cmd/fenzvideo

# Run locally
run:
	go run ./cmd/fenzvideo -conf ./configs/config.yaml

# Run tests
test:
	go test -v ./...

# Docker build
docker:
	docker build -t fenzvideo:latest .

# Docker compose up (all open-source services)
up:
	docker-compose up -d

# Seed sample data (GEMINI_KEY optional)
seed:
	go run ./cmd/seed/

# Docker compose down
down:
	docker-compose down

# View logs
logs:
	docker-compose logs -f app

# Open observability dashboards
observe:
	@echo "Jaeger UI:     http://localhost:16686"
	@echo "Prometheus:    http://localhost:9091"
	@echo "Grafana:       http://localhost:3000"
	@echo "MinIO Console: http://localhost:9001"
	@echo "NATS Monitor:  http://localhost:8222"
```

---

## Open-Source Alternatives Reference

For any component, the following open-source swaps are possible:

| Component          | Current       | Alternative (also open source)   |
| ------------------ | ------------- | -------------------------------- |
| Database           | MySQL 8.0     | PostgreSQL 16, MariaDB 11        |
| Cache              | Redis 7       | Valkey, KeyDB, Dragonfly         |
| Object Storage     | MinIO         | SeaweedFS, Garage                |
| Tracing            | Jaeger        | Zipkin, Grafana Tempo            |
| Monitoring         | Prometheus    | VictoriaMetrics                  |
| Dashboards         | Grafana OSS   | Metabase                         |
| Reverse Proxy      | Nginx         | Caddy, Traefik                   |
| CI/CD              | Gitea Actions | Woodpecker CI, Drone CI          |
| Container Registry | Docker Hub    | Harbor, Gitea Container Registry |
| Message Broker     | NATS          | RabbitMQ, Redis Pub/Sub, Kafka   |

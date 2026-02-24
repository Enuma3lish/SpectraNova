# FenzVideo Backend Architecture

## Tech Stack (100% Open Source)

| Category         | Technology                                                                                              | License               | Description                                               |
| ---------------- | ------------------------------------------------------------------------------------------------------- | --------------------- | --------------------------------------------------------- |
| Language         | [Go](https://go.dev/) 1.22+                                                                             | BSD-3                 | High-performance compiled language                        |
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
│           ├── auth.proto
│           ├── video.proto
│           ├── channel.proto
│           ├── category.proto
│           ├── tag.proto             # Tag CRUD & user tag preferences
│           ├── search.proto
│           ├── dashboard.proto
│           ├── donation.proto        # Donation & Paddle payment
│           ├── notification.proto   # Notification service (NATS-driven)
│           ├── admin.proto           # Admin account management
│           ├── user.proto
│           └── error_reason.proto
│
├── cmd/                          # Application entry points
│   ├── backend/
│   │   ├── main.go               # App bootstrap
│   │   ├── wire.go               # Wire dependency injection
│   │   └── wire_gen.go           # Wire generated code
│   └── seed/
│       └── main.go               # Seed data generator (Gemini API)
│
├── configs/                      # Configuration files
│   ├── config.yaml               # Main config (db, redis, jwt, server)
│   └── config.prod.yaml
│
├── internal/                     # Private application code
│   ├── biz/                      # Business logic layer (use cases)
│   │   ├── auth.go               # AuthUsecase
│   │   ├── video.go              # VideoUsecase
│   │   ├── channel.go            # ChannelUsecase
│   │   ├── category.go           # CategoryUsecase
│   │   ├── tag.go                # TagUsecase
│   │   ├── search.go             # SearchUsecase
│   │   ├── dashboard.go          # DashboardUsecase
│   │   ├── donation.go           # DonationUsecase
│   │   ├── notification.go      # NotificationUsecase
│   │   ├── admin.go              # AdminUsecase
│   │   └── user.go               # UserUsecase
│   │
│   ├── conf/                     # Config struct definitions
│   │   └── conf.proto            # Protobuf-based config
│   │
│   ├── data/                     # Data access layer (repository implementations)
│   │   ├── data.go               # DB & Redis client initialization + warm-up trigger
│   │   ├── cache_warmup.go       # WarmUpCache (boot-time Redis population from MySQL)
│   │   ├── video_cache.go        # VideoCacheRepo (tag SETs + video HASHes + view buffer)
│   │   ├── cleanup_worker.go     # Background worker for failed cache evictions
│   │   ├── model/                # GORM model definitions
│   │   │   ├── user.go
│   │   │   ├── video.go
│   │   │   ├── channel.go
│   │   │   ├── category.go
│   │   │   ├── tag.go            # Tag + VideoTag + UserTagPreference
│   │   │   ├── donation.go       # Donation model
│   │   │   ├── notification.go  # Notification model
│   │   │   ├── membership.go
│   │   │   └── view_record.go
│   │   ├── auth.go               # AuthRepo implementation
│   │   ├── video.go              # VideoRepo implementation
│   │   ├── channel.go            # ChannelRepo implementation
│   │   ├── category.go           # CategoryRepo implementation
│   │   ├── tag.go                # TagRepo implementation
│   │   ├── search.go             # SearchRepo implementation
│   │   ├── dashboard.go          # DashboardRepo implementation
│   │   ├── donation.go           # DonationRepo implementation
│   │   ├── notification.go      # NotificationRepo implementation
│   │   ├── admin.go              # AdminRepo implementation
│   │   └── user.go               # UserRepo implementation
│   │
│   ├── server/                   # Transport layer (HTTP & gRPC servers)
│   │   ├── http.go               # HTTP server with middleware
│   │   ├── grpc.go               # gRPC server
│   │   └── middleware.go         # JWT auth middleware, CORS, admin guard
│   │
│   ├── service/                  # Service layer (API handler / adapter)
│   │   ├── auth.go               # AuthService (proto → biz)
│   │   ├── video.go              # VideoService
│   │   ├── channel.go            # ChannelService
│   │   ├── category.go           # CategoryService
│   │   ├── tag.go                # TagService
│   │   ├── search.go             # SearchService
│   │   ├── dashboard.go          # DashboardService
│   │   ├── donation.go           # DonationService
│   │   ├── notification.go      # NotificationService
│   │   ├── admin.go              # AdminService
│   │   └── user.go               # UserService
│   │
│   └── pkg/                      # Internal shared utilities
│       ├── jwt/
│       │   └── jwt.go            # JWT token generation & validation
│       ├── hash/
│       │   └── hash.go           # Password hashing (bcrypt)
│       ├── upload/
│       │   └── minio.go          # MinIO file upload client
│       ├── paddle/
│       │   └── paddle.go         # Paddle API client (sandbox)
│       ├── nats/
│       │   └── nats.go           # NATS pub/sub client
│       └── pagination/
│           └── pagination.go     # Pagination helper
│
├── third_party/                  # Third-party proto files
│   └── google/
│       └── api/
│           ├── annotations.proto
│           └── http.proto
│
├── Dockerfile
├── docker-compose.yaml
├── Makefile
├── go.mod
└── go.sum
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
    option (google.api.http) = { get: "/api/v1/videos/recommended" };
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
  // Tier 2: Upgrade to paid premium → returns Paddle checkout URL
  rpc UpgradeToPremium (UpgradeToPremiumRequest) returns (UpgradeToPremiumReply) {
    option (google.api.http) = {
      post: "/api/v1/channels/{id}/premium"
      body: "*"
    };
  }
  // Cancel premium subscription (downgrade to Tier 1)
  rpc CancelPremium (CancelPremiumRequest) returns (MembershipReply) {
    option (google.api.http) = { delete: "/api/v1/channels/{id}/premium" };
  }
}

message ChannelReply {
  int64 id = 1;
  string name = 2;
  string avatar_url = 3;
  double monthly_fee = 4;       // Tier 2 premium price
  int64 subscriber_count = 5;   // total tier 1 + tier 2
  int64 premium_count = 6;      // tier 2 only
  string membership_tier = 7;   // viewer's current tier: "none" / "subscriber" / "premium"
}

message UpgradeToPremiumReply {
  string checkout_url = 1;
  string paddle_subscription_id = 2;
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

### Dashboard Service

```protobuf
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

message AnalyticsReply {
  int64 total_views_member = 1;
  int64 total_views_non_member = 2;
  repeated VideoViewRanking views_ranking = 3;
  int64 member_count = 4;
  double member_ratio = 5;
  double revenue = 6;             // membership revenue
  double donation_revenue = 7;    // total received donations
}
```

### User Service

```protobuf
service UserService {
  rpc UpdateDisplayName (UpdateDisplayNameRequest) returns (UserReply) {
    option (google.api.http) = {
      put: "/api/v1/user/display-name"
      body: "*"
    };
  }
  rpc UpdatePassword (UpdatePasswordRequest) returns (UpdatePasswordReply) {
    option (google.api.http) = {
      put: "/api/v1/user/password"
      body: "*"
    };
  }
  // User self-delete: hidden delete (preserves data but hides account)
  rpc HideAccount (HideAccountRequest) returns (HideAccountReply) {
    option (google.api.http) = {
      put: "/api/v1/user/account/hide"
      body: "*"
    };
  }
  // User self-delete: real delete (permanent, sets deleted_at)
  rpc DeleteAccount (DeleteAccountRequest) returns (DeleteAccountReply) {
    option (google.api.http) = { delete: "/api/v1/user/account" };
  }
  // User self-delete channel
  rpc DeleteChannel (DeleteChannelRequest) returns (DeleteChannelReply) {
    option (google.api.http) = { delete: "/api/v1/user/channel" };
  }
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

### Donation Service

Donations are placed at the **video level** rather than the channel level. Since a single donation is closer to an impulse purchase, it should be triggered at the point where the user's intent is strongest — while watching a video.

```protobuf
service DonationService {
  // Create a donation for a specific video → returns Paddle checkout URL
  rpc CreateDonation (CreateDonationRequest) returns (CreateDonationReply) {
    option (google.api.http) = {
      post: "/api/v1/videos/{video_id}/donate"
      body: "*"
    };
  }
  // Get donations sent by the current user
  rpc GetMyDonations (GetMyDonationsRequest) returns (DonationListReply) {
    option (google.api.http) = { get: "/api/v1/donations/sent" };
  }
  // Get donations received by the current user (creator)
  rpc GetReceivedDonations (GetReceivedDonationsRequest) returns (DonationListReply) {
    option (google.api.http) = { get: "/api/v1/donations/received" };
  }
  // Paddle webhook callback (no auth — verified by Paddle signature)
  rpc HandleWebhook (PaddleWebhookRequest) returns (PaddleWebhookReply) {
    option (google.api.http) = {
      post: "/api/v1/webhooks/paddle"
      body: "*"
    };
  }
}

message CreateDonationRequest {
  int64 video_id = 1;             // target video ID (creator resolved from video owner)
  string amount = 2;              // decimal string, e.g. "5.00"
  string currency = 3;            // ISO 4217, default "USD"
  optional string message = 4;    // optional message to creator
}

message CreateDonationReply {
  int64 donation_id = 1;
  string checkout_url = 2;        // Paddle checkout URL to redirect user
  string paddle_transaction_id = 3;
}

message DonationListReply {
  repeated DonationItem donations = 1;
  int64 total = 2;
}

message DonationItem {
  int64 id = 1;
  string donor_name = 2;
  string creator_name = 3;
  int64 video_id = 4;
  string video_title = 5;
  string amount = 6;
  string currency = 7;
  string message = 8;
  string status = 9;              // pending / completed / refunded
  string created_at = 10;
}
```

### Notification Service

```protobuf
service NotificationService {
  // List current user's notifications (paginated)
  rpc ListNotifications (ListNotificationsRequest) returns (NotificationListReply) {
    option (google.api.http) = { get: "/api/v1/notifications" };
  }
  // Get unread notification count
  rpc GetUnreadCount (GetUnreadCountRequest) returns (UnreadCountReply) {
    option (google.api.http) = { get: "/api/v1/notifications/unread-count" };
  }
  // Mark one notification as read
  rpc MarkRead (MarkReadRequest) returns (MarkReadReply) {
    option (google.api.http) = {
      put: "/api/v1/notifications/{id}/read"
      body: "*"
    };
  }
  // Mark all notifications as read
  rpc MarkAllRead (MarkAllReadRequest) returns (MarkAllReadReply) {
    option (google.api.http) = {
      put: "/api/v1/notifications/read-all"
      body: "*"
    };
  }
}

message NotificationListReply {
  repeated NotificationItem notifications = 1;
  int64 total = 2;
}

message NotificationItem {
  int64 id = 1;
  string type = 2;                // new_video / video_update / subscription
  string title = 3;
  string message = 4;
  string payload = 5;             // JSON string (channel_id, video_id, etc.)
  bool is_read = 6;
  string created_at = 7;
}

message UnreadCountReply {
  int64 count = 1;
}
```

### Admin Service

```protobuf
service AdminService {
  // List all users (with filters)
  rpc ListUsers (ListUsersRequest) returns (UserListReply) {
    option (google.api.http) = { get: "/api/v1/admin/users" };
  }
  // Get user details
  rpc GetUser (GetUserRequest) returns (AdminUserReply) {
    option (google.api.http) = { get: "/api/v1/admin/users/{id}" };
  }
  // Create user
  rpc CreateUser (CreateUserRequest) returns (AdminUserReply) {
    option (google.api.http) = {
      post: "/api/v1/admin/users"
      body: "*"
    };
  }
  // Update user
  rpc UpdateUser (UpdateUserRequest) returns (AdminUserReply) {
    option (google.api.http) = {
      put: "/api/v1/admin/users/{id}"
      body: "*"
    };
  }
  // Hidden delete: set is_hidden = true (reversible)
  rpc HideUser (HideUserRequest) returns (AdminUserReply) {
    option (google.api.http) = {
      put: "/api/v1/admin/users/{id}/hide"
      body: "*"
    };
  }
  // Restore hidden user: set is_hidden = false
  rpc RestoreUser (RestoreUserRequest) returns (AdminUserReply) {
    option (google.api.http) = {
      put: "/api/v1/admin/users/{id}/restore"
      body: "*"
    };
  }
  // Real delete: permanent removal (sets deleted_at)
  rpc DeleteUser (DeleteUserRequest) returns (DeleteUserReply) {
    option (google.api.http) = { delete: "/api/v1/admin/users/{id}" };
  }
  // Admin manage tags
  rpc CreateTag (CreateTagRequest) returns (TagReply) {
    option (google.api.http) = {
      post: "/api/v1/admin/tags"
      body: "*"
    };
  }
  rpc UpdateTag (UpdateTagRequest) returns (TagReply) {
    option (google.api.http) = {
      put: "/api/v1/admin/tags/{id}"
      body: "*"
    };
  }
  rpc DeleteTag (DeleteTagRequest) returns (DeleteTagReply) {
    option (google.api.http) = { delete: "/api/v1/admin/tags/{id}" };
  }
}

message ListUsersRequest {
  int32 page = 1;
  int32 page_size = 2;
  optional string role = 3;         // filter by role
  optional bool is_hidden = 4;      // filter hidden users
  optional string search = 5;       // search by username/display_name
}

message AdminUserReply {
  int64 id = 1;
  string username = 2;
  string display_name = 3;
  string role = 4;
  bool is_hidden = 5;
  string avatar_url = 6;
  string created_at = 7;
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

| Route Pattern                     | Auth Required                            |
| --------------------------------- | ---------------------------------------- |
| `POST /auth/login`                | No                                       |
| `POST /auth/register`             | No                                       |
| `GET /videos/recommended`         | No (uses tags from session or user)      |
| `GET /videos/:id`                 | No (member-only videos check membership) |
| `GET /search`                     | No                                       |
| `GET /categories/**`              | No                                       |
| `GET /channels/:id`               | No                                       |
| `GET /tags`                       | No                                       |
| `GET /tags/my`                    | No (guest uses session_id query param)   |
| `PUT /tags/my`                    | No (guest uses session_id in body)       |
| `POST /channels/:id/subscribe`    | **Yes**                                  |
| `DELETE /channels/:id/subscribe`  | **Yes**                                  |
| `POST /channels/:id/premium`      | **Yes** (upgrade to Tier 2 via Paddle)   |
| `DELETE /channels/:id/premium`    | **Yes** (cancel premium)                 |
| `POST /videos`                    | **Yes**                                  |
| `PUT /videos/:id`                 | **Yes** (owner only)                     |
| `DELETE /videos/:id`              | **Yes** (owner only)                     |
| `GET /dashboard/**`               | **Yes**                                  |
| `PUT /dashboard/**`               | **Yes**                                  |
| `PUT /user/**`                    | **Yes**                                  |
| `DELETE /user/account`            | **Yes**                                  |
| `DELETE /user/channel`            | **Yes**                                  |
| `POST /videos/:id/donate`         | **Yes**                                  |
| `GET /donations/sent`             | **Yes**                                  |
| `GET /donations/received`         | **Yes**                                  |
| `POST /webhooks/paddle`           | No (verified by Paddle signature)        |
| `GET /notifications`              | **Yes**                                  |
| `GET /notifications/unread-count` | **Yes**                                  |
| `PUT /notifications/:id/read`     | **Yes**                                  |
| `PUT /notifications/read-all`     | **Yes**                                  |
| `GET /admin/**`                   | **Yes** (admin role only)                |
| `POST /admin/**`                  | **Yes** (admin role only)                |
| `PUT /admin/**`                   | **Yes** (admin role only)                |
| `DELETE /admin/**`                | **Yes** (admin role only)                |

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
    FindByUserID(ctx context.Context, userID int64) (*Channel, error)
    GetMembership(ctx context.Context, userID, channelID int64) (*Membership, error)
    Subscribe(ctx context.Context, userID, channelID int64) error          // Tier 1 free
    Unsubscribe(ctx context.Context, userID, channelID int64) error
    UpgradeToPremium(ctx context.Context, membershipID int64, paddleSubID string) error // Tier 2
    CancelPremium(ctx context.Context, membershipID int64) error
    UpdatePaddleStatus(ctx context.Context, paddleSubID, status string) error
    ListSubscribers(ctx context.Context, channelID int64) ([]*Membership, error) // all tiers
    SetFee(ctx context.Context, channelID int64, fee float64) error
    GetAnalytics(ctx context.Context, channelID int64) (*Analytics, error)
    Hide(ctx context.Context, channelID int64, hidden bool) error
    Delete(ctx context.Context, channelID int64) error  // real delete
}

type ChannelUsecase struct {
    repo         ChannelRepo
    paddleClient *PaddleClient
    natsClient   *NATSClient
    log          *log.Helper
}

func (uc *ChannelUsecase) Subscribe(ctx context.Context, userID, channelID int64) error {
    // 1. Check channel exists and is not hidden
    // 2. Check user is not already subscribed
    // 3. Create membership with tier=1, status="active"
}

func (uc *ChannelUsecase) UpgradeToPremium(ctx context.Context, userID, channelID int64) (string, error) {
    // 1. Get existing membership (must be tier 1 subscriber)
    // 2. Get channel monthly_fee
    // 3. Create Paddle recurring subscription via API
    // 4. Update membership tier=2, paddle_subscription_id, paddle_status="active"
    // 5. Return Paddle checkout URL
}

func (uc *ChannelUsecase) CancelPremium(ctx context.Context, userID, channelID int64) error {
    // 1. Get membership (must be tier 2)
    // 2. Cancel Paddle subscription via API
    // 3. Downgrade membership to tier=1
}
```

```go
// biz/admin.go
type AdminRepo interface {
    ListUsers(ctx context.Context, params *AdminListParams) ([]*User, int64, error)
    GetUser(ctx context.Context, id int64) (*User, error)
    CreateUser(ctx context.Context, user *User) (*User, error)
    UpdateUser(ctx context.Context, user *User) (*User, error)
    HideUser(ctx context.Context, id int64) error        // set is_hidden = true
    RestoreUser(ctx context.Context, id int64) error     // set is_hidden = false
    DeleteUser(ctx context.Context, id int64) error      // real delete (deleted_at)
}

type AdminUsecase struct {
    repo        AdminRepo
    channelRepo ChannelRepo
    videoRepo   VideoRepo
    log         *log.Helper
}

func (uc *AdminUsecase) HideUser(ctx context.Context, userID int64) error {
    // 1. Set user.is_hidden = true
    // 2. Set user's channel.is_hidden = true
    // 3. Set all user's videos.is_hidden = true
}

func (uc *AdminUsecase) RestoreUser(ctx context.Context, userID int64) error {
    // 1. Set user.is_hidden = false
    // 2. Set user's channel.is_hidden = false
    // 3. Set all user's videos.is_hidden = false
}

func (uc *AdminUsecase) DeleteUser(ctx context.Context, userID int64) error {
    // 1. Permanently delete user (set deleted_at)
    // 2. Permanently delete channel
    // 3. Permanently delete all videos + MinIO files
    // 4. Remove memberships
}
```

```go
// biz/user.go
type UserUsecase struct {
    repo        UserRepo
    channelRepo ChannelRepo
    videoRepo   VideoRepo
    log         *log.Helper
}

func (uc *UserUsecase) HideAccount(ctx context.Context, userID int64) error {
    // User self-hide: set is_hidden on user + channel + videos
}

func (uc *UserUsecase) DeleteAccount(ctx context.Context, userID int64) error {
    // User self-delete: permanent removal of user + channel + videos
}

func (uc *UserUsecase) DeleteChannel(ctx context.Context, userID int64) error {
    // Delete user's channel + all channel videos
    // User account remains
}
```

```go
// biz/donation.go
type DonationRepo interface {
    Create(ctx context.Context, donation *Donation) (*Donation, error)
    FindByID(ctx context.Context, id int64) (*Donation, error)
    UpdatePaddleStatus(ctx context.Context, paddleTxnID string, status string) error
    ListByDonor(ctx context.Context, donorID int64, page, pageSize int) ([]*Donation, int64, error)
    ListByCreator(ctx context.Context, creatorID int64, page, pageSize int) ([]*Donation, int64, error)
    GetTotalReceivedByCreator(ctx context.Context, creatorID int64) (float64, error)
}

type DonationUsecase struct {
    repo         DonationRepo
    videoRepo    VideoRepo
    paddleClient *PaddleClient
    log          *log.Helper
}

func (uc *DonationUsecase) CreateDonation(ctx context.Context, donorID, videoID int64, amount, currency, message string) (*Donation, string, error) {
    // 1. Look up the video to resolve creator (video.UserID)
    // 2. Validate video exists, is published, and is not hidden
    // 3. Validate donor is not the video owner (cannot donate to self)
    // 4. Create donation record with video_id, donor_id, creator_id, paddle_status = "pending"
    // 5. Call Paddle API to create a transaction (sandbox):
    //    - Create a one-time price item with the donation amount
    //    - Set custom_data with { donation_id, donor_id, creator_id, video_id }
    //    - Get back a checkout URL
    // 6. Save paddle_transaction_id to donation record
    // 7. Return donation + checkout URL
}

func (uc *DonationUsecase) HandlePaddleWebhook(ctx context.Context, payload []byte, signature string) error {
    // 1. Verify webhook signature using Paddle's webhook secret
    // 2. Parse event type from payload
    // 3. Handle relevant events:
    //    - "transaction.completed" → update donation paddle_status to "completed"
    //    - "transaction.payment_failed" → update donation paddle_status to "cancelled"
    //    - "transaction.refunded" → update donation paddle_status to "refunded"
    //    - "subscription.activated" → update membership paddle_status to "active"
    //    - "subscription.canceled" → downgrade membership to tier 1
    //    - "subscription.past_due" → update membership paddle_status to "past_due"
    // 4. Extract custom_data (donation_id or membership_id)
    // 5. Update corresponding record in DB
}
```

```go
// biz/notification.go
type NotificationRepo interface {
    Create(ctx context.Context, notif *Notification) (*Notification, error)
    CreateBatch(ctx context.Context, notifs []*Notification) error
    ListByUser(ctx context.Context, userID int64, page, pageSize int) ([]*Notification, int64, error)
    UnreadCount(ctx context.Context, userID int64) (int64, error)
    MarkRead(ctx context.Context, id, userID int64) error
    MarkAllRead(ctx context.Context, userID int64) error
}

type NotificationUsecase struct {
    repo        NotificationRepo
    channelRepo ChannelRepo
    natsClient  *NATSClient
    log         *log.Helper
}

// PublishNewVideo — called when a creator publishes a new video.
// Publishes event to NATS subject "channel.<channelID>.new_video"
// which triggers creation of notification records for all subscribers.
func (uc *NotificationUsecase) PublishNewVideo(ctx context.Context, channelID, videoID int64, videoTitle string) error {
    // 1. Publish NATS event: { channel_id, video_id, title, type: "new_video" }
    //    Subject: "channel.<channelID>.new_video"
    // 2. NATS subscriber handler (running in background goroutine):
    //    a. Get all subscribers (tier 1 + tier 2) for the channel
    //    b. Create a Notification record for each subscriber
    //    c. Push real-time notification via SSE/WebSocket to connected users
}

// StartNATSSubscriber — starts a background NATS subscriber that listens
// for channel events and creates notification records.
func (uc *NotificationUsecase) StartNATSSubscriber() error {
    // Subscribe to "channel.*.new_video" and "channel.*.video_update"
    // On message:
    //   1. Parse event payload (channel_id, video_id, title)
    //   2. Get all subscribers for the channel
    //   3. Batch-create notification records
}
```

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
| `tag:{id}` | SET | 30 min | Video IDs belonging to this tag (index layer) |
| `video:{id}` | HASH | 30 min | Video summary fields (data layer, one copy per video) |
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
            d.Redis.Expire(ctx, tagKey, 30*time.Minute)
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
            d.Redis.Expire(ctx, videoKey, 30*time.Minute)
        }
    }
}
```

> **TTL after warm-up**: Keys expire after 30 min. After expiry, lazy populate handles individual cache misses. The warm-up covers the critical initial burst; lazy handles the steady state.

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
        r.rdb.Expire(ctx, key, 30*time.Minute)
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

## Paddle Payment Client

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

## NATS Pub/Sub Client

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
    timeout: 30s
  grpc:
    addr: 0.0.0.0:9000
    timeout: 30s

data:
  database:
    driver: mysql
    source: "user:password@tcp(127.0.0.1:3306)/fenzvideo?charset=utf8mb4&parseTime=True&loc=Local"
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
  jwt_secret: "your-secret-key"
  token_expiry: 24h
  refresh_expiry: 168h # 7 days

storage:
  endpoint: "127.0.0.1:9000"
  access_key: "minioadmin"
  secret_key: "minioadmin"
  bucket: "fenzvideo"
  use_ssl: false

paddle:
  api_key: "pdl_sdbx_..." # Paddle sandbox API key
  webhook_secret: "pdl_ntfset_..." # Webhook destination secret
  environment: sandbox # sandbox | production

nats:
  url: "nats://127.0.0.1:4222" # NATS server URL

tracing:
  endpoint: "http://127.0.0.1:14268/api/traces" # Jaeger collector
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
# docker-compose.yaml
version: "3.8"

services:
  app:
    build: .
    ports:
      - "8000:8000"
      - "9000:9000"
    depends_on:
      mysql:
        condition: service_healthy
      redis:
        condition: service_started
      minio:
        condition: service_started
      nats:
        condition: service_started
    environment:
      - CONFIG_PATH=/app/configs/config.yaml
    volumes:
      - ./configs:/app/configs
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx/nginx.conf:/etc/nginx/nginx.conf:ro
      - ./nginx/ssl:/etc/nginx/ssl:ro
    depends_on:
      - app
    restart: unless-stopped

  mysql:
    image: mysql:8.0
    environment:
      MYSQL_ROOT_PASSWORD: rootpassword
      MYSQL_DATABASE: fenzvideo
      MYSQL_USER: fenzvideo
      MYSQL_PASSWORD: fenzvideo
    ports:
      - "3306:3306"
    volumes:
      - mysql_data:/var/lib/mysql
      - ./init.sql:/docker-entrypoint-initdb.d/init.sql
    healthcheck:
      test: ["CMD", "mysqladmin", "ping", "-h", "localhost"]
      interval: 10s
      timeout: 5s
      retries: 5
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

  minio:
    image: minio/minio
    command: server /data --console-address ":9001"
    ports:
      - "9090:9000"
      - "9001:9001"
    environment:
      MINIO_ROOT_USER: minioadmin
      MINIO_ROOT_PASSWORD: minioadmin
    volumes:
      - minio_data:/data
    restart: unless-stopped

  nats:
    image: nats:2-alpine
    ports:
      - "4222:4222" # Client connections
      - "8222:8222" # HTTP monitoring
    command: ["--jetstream", "--http_port", "8222"]
    volumes:
      - nats_data:/data
    restart: unless-stopped

  # --- Observability Stack (all open source) ---

  jaeger:
    image: jaegertracing/all-in-one:latest
    ports:
      - "16686:16686" # Jaeger UI
      - "14268:14268" # Collector HTTP
    environment:
      COLLECTOR_OTLP_ENABLED: "true"
    restart: unless-stopped

  prometheus:
    image: prom/prometheus:latest
    ports:
      - "9091:9090"
    volumes:
      - ./prometheus/prometheus.yml:/etc/prometheus/prometheus.yml:ro
      - prometheus_data:/prometheus
    restart: unless-stopped

  grafana:
    image: grafana/grafana-oss:latest
    ports:
      - "3000:3000"
    environment:
      GF_SECURITY_ADMIN_PASSWORD: admin
    volumes:
      - grafana_data:/var/lib/grafana
    depends_on:
      - prometheus
    restart: unless-stopped

volumes:
  mysql_data:
  redis_data:
  minio_data:
  nats_data:
  prometheus_data:
  grafana_data:
```

---

## Seed Data Generator

The seed script (`cmd/seed/main.go`) populates the database with sample data before starting services. It uses the **Gemini API** to generate video titles and descriptions in Traditional Chinese.

### Usage

```bash
# Requires GEMINI_KEY in .env and MySQL running
cd backend && make seed
```

### What Gets Seeded

| Entity | Count | Details |
|--------|-------|---------|
| Admin user | 1 | `admin` / `admin123` (role: admin) |
| Creator users | 5 | `creator_alice` through `creator_emma` (role: user, password: `password123`) |
| Channels | 6 | One per user (admin + 5 creators), random monthly fee |
| Categories | 10 | 音樂, 遊戲, 教育, 娛樂, 科技, 運動, 新聞, 美食, 旅遊, 生活 |
| Tags | 15 | 搞笑, 教學, Vlog, 開箱, 直播精華, 音樂MV, 遊戲實況, 美食料理, 旅行紀錄, 科技評測, 新手入門, 健身運動, 動畫, 訪談, DIY手作 |
| Videos | 15 | One per tag, AI-generated title & description, random views/duration |

### Key Behaviors

- **Idempotent**: Checks for existing data before inserting; safe to run multiple times
- **Gemini API**: Calls `gemini-2.0-flash` model to generate creative video content in 繁體中文
- **Rate limited**: 1-second delay between Gemini API calls to avoid quota limits
- **Fallback**: If Gemini API fails, uses a simple placeholder title/description
- **DB connection**: Defaults to `root:root@tcp(127.0.0.1:3306)/fenzvideo`, overridable via `DB_DSN` env var
- **Auto-migrate**: Runs GORM AutoMigrate before seeding (creates tables if not exist)

### Seed Flow

```
1. Load .env (GEMINI_KEY)
2. Connect MySQL → AutoMigrate all tables
3. Seed admin user + channel (skip if exists)
4. Seed 10 categories (skip if any exist)
5. Seed 15 tags (skip if any exist)
6. Seed 5 creator users + channels (skip if exist)
7. For each of 15 tags:
   → Call Gemini API → generate title + description
   → Create video (round-robin across creators & categories)
   → Associate video ↔ tag in video_tags
8. Done — all data ready for services
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

# Seed sample data via Gemini API (requires GEMINI_KEY in .env)
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

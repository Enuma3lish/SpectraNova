# FenzVideo

## Introduction

FenzVideo is an online video platform where users can upload and watch videos, with the option to set specific videos as member-exclusive content. The entire tech stack is built on **100% open-source** tools.

---

## Current Status

**Phase 2 (Core MVP)** is complete. The backend supports user registration, login, video upload, tag-based recommendations, search, and channel subscriptions.

| What's done | Details |
|---|---|
| **Auth** | Register, login, JWT access + refresh tokens |
| **Videos** | CRUD, upload to MinIO, tag-based recommendations, access tier enforcement, view counting |
| **Tags** | List tags, user/guest tag preferences (max 5), session_id support |
| **Categories** | List categories, seed 10 categories |
| **Search** | MySQL FULLTEXT (BOOLEAN MODE), filters (category, duration, date, views, access type) |
| **Channels** | Auto-create on registration, free subscribe/unsubscribe |
| **Recommendation cache** | Redis two-layer (per-tag SET + per-video HASH), boot warm-up, lazy fallback, app-level eviction, cleanup worker |
| **View count buffer** | Redis HINCRBY → batch flush to MySQL every 30s |
| **Infrastructure** | Docker Compose: MySQL, Redis, MinIO, NATS, Jaeger |
| **GORM models** | 12 tables (users, channels, videos, categories, tags, video_tags, user_tag_preferences, memberships, view_records, notifications, donations) |
| **Middleware** | JWT authentication, admin guard, CORS |
| **Seed data** | Gemini API generates 15 videos with Traditional Chinese content |

See [Roadmap_dev.md](docs/Roadmap_dev.md) for the full development roadmap.

---

## Quick Start

```bash
# 1. Start infrastructure
docker-compose up -d

# 2. Build the backend
cd backend && make build

# 3. Seed sample data (requires GEMINI_KEY in .env)
make seed

# 4. Run the server
make run
```

### Try the API

```bash
# Register a user
curl -X POST localhost:8000/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test123","display_name":"Test"}'

# Login
curl -X POST localhost:8000/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username":"test","password":"test123"}'

# Public endpoints
curl localhost:8000/api/v1/categories
curl localhost:8000/api/v1/tags
curl localhost:8000/api/v1/videos/recommended

# Search
curl "localhost:8000/api/v1/search?query=test"

# Protected endpoints (use token from login)
curl -H "Authorization: Bearer <token>" localhost:8000/api/v1/tags/my
```

### Environment Variables

Create a `.env` file in the project root:

```
GEMINI_KEY="your-gemini-api-key"
Paddle_KEY="your-paddle-sandbox-key"
ADMIN_EMAIL="admin"
ADMIN_PASSWORD="admin123"
```

---

## Architecture Overview

The project is documented across three architecture files:

| Document                                                  | Description                                                                                                              |
| --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| [backend-architecture.md](docs/backend-architecture.md)   | Go (Kratos v2) backend with clean architecture layers, gRPC + HTTP API, JWT auth, admin & tag services, Paddle donations, recommendation cache, seed data generator |
| [frontend-architecture.md](docs/frontend-architecture.md) | Vue 3 + Vite SPA with Element Plus, Tailwind CSS, Pinia stores, Video.js player, Paddle.js checkout                      |
| [db-architecture.md](docs/db-architecture.md)             | MySQL 8.0 schema with GORM v2, 12 tables, tag-based recommendation, two-tier delete, donations                           |

### Tech Stack Summary

```
Frontend:   Vue 3 · Vite · Element Plus · Tailwind CSS · Video.js · Pinia · Axios · Paddle.js
Backend:    Go 1.24+ · Kratos v2 · GORM v2 · Protocol Buffers · Wire (DI) · Paddle Go SDK
Database:   MySQL 8.0 · Redis / Valkey (cache + recommendation)
Storage:    MinIO (S3-compatible object storage)
Payment:    Paddle (sandbox) — donations + premium subscriptions
Messaging:  NATS (pub/sub for real-time notifications)
AI:         Gemini API (seed data generation)
Infra:      Docker · Nginx · Jaeger · Prometheus · Grafana
```

### Key Architecture Decisions

- **Clean architecture** — Backend follows Kratos layout: Transport → Service → Biz ← Data (dependency inversion)
- **Tag-based recommendation** — Users (registered or guest) pick up to 5 tags; the system randomly combines 1–N tags per request to surface videos
- **Redis recommendation cache** — Two-layer design (per-tag SET index + per-video HASH data) with boot warm-up (eliminates cold start), lazy fallback after TTL expiry, application-level eviction, and cleanup worker for failure recovery
- **Two-tier delete** — "Hidden" (`is_hidden` flag, reversible by admin) vs "Real delete" (hard delete, permanent)
- **Admin system** — Admin role with full CRUD over all user accounts, channels, and tags
- **Self-service account management** — Users can hide or permanently delete their own account and channel
- **Guest support** — Guests can select tags via a `session_id` (UUID stored in localStorage) without registering
- **Dual-protocol API** — gRPC + HTTP via Kratos dual transport, with Protobuf-defined contracts
- **Video-level donations** — Donate button placed on the Video Page (not Channel Page) to capture impulse-purchase intent at the point where the user is most engaged; uses Paddle one-time transactions with webhook-driven status updates
- **AI-powered seed data** — Gemini API generates creative video titles and descriptions in Traditional Chinese for realistic sample data

---

## Pages

- Login Page
- Home Page
- Search Results
- Category Page
- Channel Page
- Dashboard
- Admin Page
- Video Page

---

## Login Page

### Features

Enter username and password to log in, or create a new account. Authentication is handled via JWT.

### Login States

- **Guest (not logged in):** Can access the Home Page, Category Page, Channel Page, search, and watch videos. Cannot access the Dashboard (no channel, cannot upload videos) and cannot become a member of any channel. Can still select tags for personalized recommendations.
- **Registered user:** Full access including Dashboard and membership.
- **Admin:** All registered user permissions plus access to the Admin Page for managing all accounts and tags.

---

## Home Page

### Features

Displays **tag-based recommended videos**, a search bar, and links to other pages. Clicking a video navigates to the Video Page; clicking the uploader's name navigates to their Channel Page.

Users (registered or guest) can pick up to **5 tags** to personalize their recommendations. The system randomly selects a subset of those tags on each request for variety.

---

## Search Results

### Features

Displays videos matching the search query. Clicking a video navigates to the Video Page; clicking the uploader's name navigates to their Channel Page. Users can apply additional filters to narrow results.

### Filter Options

- Video category
- Video duration
- Upload date
- View count
- Public / Member-exclusive

---

## Category Page

### Features

Displays videos belonging to a specific category. Clicking a video navigates to the Video Page; clicking the uploader's name navigates to their Channel Page.

---

## Channel Page

### Features

Displays videos uploaded by a specific user, along with a "Join Membership" button.

### Join Membership

Clicking "Join Membership" opens a dialog showing the monthly fee and a confirmation button. If the user is already a member, the button changes to "Leave Membership" — clicking it also opens a confirmation dialog.

---

## Dashboard

### Features

Displays the user's uploaded videos, video upload form, fee settings, analytics charts, and account settings.

### Video Upload

Requires filling in: video title, description, category, tags, and whether it is member-exclusive.

### Video Edit

Users can edit the following for their uploaded videos:

- Title
- Description
- Category
- Tags
- Access setting (member / public)
- Publish / Unpublish / Delete
  - Only unpublished videos can be deleted
- Editing these fields does not affect the video's original view count or upload time

### Fee Settings

Adjust the monthly membership fee charged to channel members.

### Analytics Charts

- Total views (member vs non-member)
- Views ranking (member vs non-member)
- Member count
- Member / Non-member ratio
- Channel revenue (membership)
- Donation revenue (total received donations)

### Donations

- View **donations received** on user's channel (amount, donor, message, status)
- View **donations sent** to other creators
- Donation statuses: `pending`, `completed`, `failed`, `refunded`

### Account Settings

- Change display name (separate from login username)
- Change password
- Log out
- Hide account (reversible — makes account invisible but preserves data)
- Delete channel (permanently removes channel and all its videos)
- Delete account (permanently removes everything)
  - A confirmation dialog is required before deletion

---

## Admin Page

### Features

Accessible only to users with the `admin` role. Provides full management over all user accounts and tags.

### User Management

- View all users (with filters: role, hidden status)
- Create new user accounts
- Edit user details (display name, role, password)
- **Hide user** — sets `is_hidden = true` on user, channel, and all videos (reversible)
- **Restore user** — reverses a hide operation
- **Delete user** — permanently removes user, channel, videos, and associated data
- Admin cannot delete their own account from the admin panel

### Tag Management

- View all tags
- Create new tags
- Edit tag name / slug
- Delete tags

---

## Video Page

### Features

Plays the video and displays video information: title, view count, upload date, category, tags, and description. Includes a **"Donate" button** to support the creator.

### Donate to Creator

The donate button is placed at the video level rather than the channel level. Since a single donation is closer to an impulse purchase, it should be triggered at the point where the user's intent is strongest — while watching a video they enjoy. Clicking "Donate" opens a dialog where the user enters a donation amount, currency, and optional message. On submit, the backend creates a Paddle transaction and the frontend opens the Paddle.js checkout overlay. Payment status is updated asynchronously via Paddle webhooks.

# FenzVideo

## Introduction

FenzVideo is an online video platform where users can upload and watch videos, with the option to set specific videos as member-exclusive content. The entire tech stack is built on **100% open-source** tools.

---

## Architecture Overview

The project is documented across three architecture files:

| Document                                                  | Description                                                                                                              |
| --------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------ |
| [backend-architecture.md](docs/backend-architecture.md)   | Go (Kratos v2) backend with clean architecture layers, gRPC + HTTP API, JWT auth, admin & tag services, Paddle donations |
| [frontend-architecture.md](docs/frontend-architecture.md) | Vue 3 + Vite SPA with Element Plus, Tailwind CSS, Pinia stores, Video.js player, Paddle.js checkout                      |
| [db-architecture.md](docs/db-architecture.md)             | MySQL 8.0 schema with GORM v2, 10 tables, tag-based recommendation, two-tier delete, donations                           |

### Tech Stack Summary

```
Frontend:   Vue 3 · Vite · Element Plus · Tailwind CSS · Video.js · Pinia · Axios · Paddle.js
Backend:    Go 1.22+ · Kratos v2 · GORM v2 · Protocol Buffers · Wire (DI) · Paddle Go SDK
Database:   MySQL 8.0 · Redis / Valkey (cache)
Storage:    MinIO (S3-compatible object storage)
Payment:    Paddle (sandbox) — one-time donation transactions
Infra:      Docker · Nginx · Jaeger · Prometheus · Grafana
```

### Key Architecture Decisions

- **Clean architecture** — Backend follows Kratos layout: Transport → Service → Biz ← Data (dependency inversion)
- **Tag-based recommendation** — Users (registered or guest) pick up to 5 tags; the system randomly combines 1–N tags per request to surface videos
- **Two-tier delete** — "Hidden" (`is_hidden` flag, reversible by admin) vs "Real delete" (`deleted_at` soft delete, permanent)
- **Admin system** — Admin role with full CRUD over all user accounts, channels, and tags
- **Self-service account management** — Users can hide or permanently delete their own account and channel
- **Guest support** — Guests can select tags via a `session_id` (UUID stored in localStorage) without registering
- **Dual-protocol API** — gRPC + HTTP via Kratos dual transport, with Protobuf-defined contracts
- **Paddle sandbox donations** — Users can donate to creators via Paddle one-time transactions; webhook-driven status updates (`pending` → `completed` / `failed` / `refunded`)

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

Displays videos uploaded by a specific user, along with a "Join Membership" button and a **"Donate" button**.

### Join Membership

Clicking "Join Membership" opens a dialog showing the monthly fee and a confirmation button. If the user is already a member, the button changes to "Leave Membership" — clicking it also opens a confirmation dialog.

### Donate to Creator

Clicking "Donate" opens a dialog where the user enters a donation amount, currency, and optional message. On submit, the backend creates a Paddle transaction and the frontend opens the Paddle.js checkout overlay. Payment status is updated asynchronously via Paddle webhooks.

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

Plays the video and displays video information: title, view count, upload date, category, tags, and description.

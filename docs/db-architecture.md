# FenzVideo Database Architecture

## Database Engine

| Item      | Value              |
| --------- | ------------------ |
| RDBMS     | MySQL 8.0          |
| ORM       | GORM v2            |
| Charset   | utf8mb4            |
| Collation | utf8mb4_unicode_ci |
| Engine    | InnoDB             |

---

## Entity Relationship Diagram

```
┌──────────────┐       ┌──────────────────┐       ┌──────────────┐
│    users     │       │     videos       │       │  categories  │
├──────────────┤       ├──────────────────┤       ├──────────────┤
│ id        PK │◄──┐   │ id            PK │   ┌──▶│ id        PK │
│ username     │   │   │ user_id       FK │───┘   │ name         │
│ display_name │   ├───│ category_id   FK │───────│ slug         │
│ password     │   │   │ title            │       │ created_at   │
│ avatar_url   │   │   │ description      │       │ updated_at   │
│ role         │   │   │ video_url        │       └──────────────┘
│ is_hidden    │   │   │ thumbnail_url    │
│ created_at   │   │   │ duration         │
│ updated_at   │   │   │ views_member     │
│ deleted_at   │   │   │ views_non_member │
└──────────────┘   │   │ access_tier      │  ← 0=public, 1=subscriber, 2=premium
       │           │   │ is_published     │
       │           │   │ is_hidden        │
       │           │   │ created_at       │
       │           │   │ updated_at       │
       │           │   │ deleted_at       │
       │           │   └──────────────────┘
       │           │           │
       │     ┌─────┘           │ (many-to-many)
       │     │                 ▼
       │     │        ┌──────────────────┐
       │     │        │   video_tags     │
       │     │        ├──────────────────┤
       │     │        │ video_id      FK │───▶ videos
       │     │        │ tag_id        FK │───▶ tags
       │     │        └──────────────────┘
       │     │
       ▼     ▼        ┌──────────────┐
┌──────────────────┐  │    tags      │
│    channels      │  ├──────────────┤
├──────────────────┤  │ id        PK │
│ id            PK │  │ name         │
│ user_id    FK,UQ │  │ slug         │
│ monthly_fee     │  │ created_at   │
│ is_hidden       │  │ updated_at   │
│ created_at      │  └──────┬───────┘
│ updated_at      │         │
│ deleted_at      │         │ (many-to-many)
└──────────────────┘        ▼
       │           ┌─────────────────────────┐
       │           │  user_tag_preferences   │
       ▼           ├─────────────────────────┤
┌──────────────────┤ user_id             FK  │───▶ users (nullable)
│  memberships     │ tag_id              FK  │───▶ tags
├──────────────────┤ session_id              │  ← for guests
│ channel_id    FK │ created_at              │
│ user_id       FK │───▶ users               │
│ tier             └─────────────────────────┘
│ status           │  ← 1=free subscriber, 2=paid premium
│ paddle_sub_id    │  ← Paddle subscription ID (tier 2)
│ paddle_status    │
│ started_at       │
│ expires_at       │
│ created_at       │
│ updated_at       │
└──────────────────┘

                    ┌──────────────────┐
                    │  notifications   │
                    ├──────────────────┤
                    │ id            PK │
                    │ user_id       FK │───▶ users
                    │ type             │  ← new_video / video_update
                    │ title            │
                    │ message          │
                    │ payload (JSON)   │
                    │ is_read          │
                    │ created_at       │
                    └──────────────────┘

                    ┌──────────────────┐
                    │  view_records    │
                    ├──────────────────┤
                    │ id            PK │
                    │ video_id      FK │───▶ videos
                    │ user_id       FK │───▶ users (nullable)
                    │ is_member        │
                    │ viewed_at        │
                    └──────────────────┘

                    ┌───────────────────────┐
                    │     donations         │
                    ├───────────────────────┤
                    │ id                 PK │
                    │ donor_id           FK │───▶ users
                    │ creator_id         FK │───▶ users
                    │ amount                │
                    │ currency              │
                    │ message               │
                    │ paddle_transaction_id │  ← Paddle txn ID
                    │ paddle_status         │
                    │ created_at            │
                    │ updated_at            │
                    └───────────────────────┘
```

---

## Table Definitions

### 1. `users`

User accounts for the platform.

| Column         | Type            | Constraints              | Description                                      |
| -------------- | --------------- | ------------------------ | ------------------------------------------------ |
| `id`           | BIGINT UNSIGNED | PK, AUTO_INCREMENT       | User ID                                          |
| `username`     | VARCHAR(50)     | UNIQUE, NOT NULL         | Login username                                   |
| `display_name` | VARCHAR(100)    | NOT NULL                 | Display name (可與 username 不同)                |
| `password`     | VARCHAR(255)    | NOT NULL                 | bcrypt hashed password                           |
| `avatar_url`   | VARCHAR(500)    | NULL                     | Profile avatar URL                               |
| `role`         | VARCHAR(20)     | NOT NULL, DEFAULT 'user' | Role: `user` / `admin`                           |
| `is_hidden`    | TINYINT(1)      | NOT NULL, DEFAULT 0      | Hidden by admin (invisible but data preserved)   |
| `created_at`   | DATETIME(3)     | NOT NULL                 | Registration time                                |
| `updated_at`   | DATETIME(3)     | NOT NULL                 | Last update time                                 |
| `deleted_at`   | DATETIME(3)     | NULL, INDEX              | Hard delete (GORM soft delete for real deletion) |

**Delete strategies:**

- **Hidden delete (隱藏)**: Set `is_hidden = true`. Account data preserved but invisible to public. Admin can restore.
- **Real delete (真刪除)**: GORM soft delete via `deleted_at`. Permanently removes from all queries. Used by user self-delete or admin permanent delete.

**GORM Model:**

```go
type User struct {
    ID          uint64         `gorm:"primaryKey;autoIncrement"`
    Username    string         `gorm:"type:varchar(50);uniqueIndex;not null"`
    DisplayName string         `gorm:"type:varchar(100);not null"`
    Password    string         `gorm:"type:varchar(255);not null"`
    AvatarURL   *string        `gorm:"type:varchar(500)"`
    Role        string         `gorm:"type:varchar(20);not null;default:'user'"`
    IsHidden    bool           `gorm:"not null;default:false"`
    CreatedAt   time.Time
    UpdatedAt   time.Time
    DeletedAt   gorm.DeletedAt `gorm:"index"`

    // Relations
    Channel        *Channel          `gorm:"foreignKey:UserID"`
    Videos         []Video           `gorm:"foreignKey:UserID"`
    Memberships    []Membership      `gorm:"foreignKey:UserID"`
    TagPreferences []UserTagPreference `gorm:"foreignKey:UserID"`
}
```

---

### 2. `channels`

Each user has one channel (created upon registration).

| Column        | Type            | Constraints                     | Description                 |
| ------------- | --------------- | ------------------------------- | --------------------------- |
| `id`          | BIGINT UNSIGNED | PK, AUTO_INCREMENT              | Channel ID                  |
| `user_id`     | BIGINT UNSIGNED | FK → users.id, UNIQUE, NOT NULL | Owner                       |
| `monthly_fee` | DECIMAL(10,2)   | NOT NULL, DEFAULT 0             | Monthly membership fee      |
| `is_hidden`   | TINYINT(1)      | NOT NULL, DEFAULT 0             | Hidden by admin or user     |
| `created_at`  | DATETIME(3)     | NOT NULL                        |                             |
| `updated_at`  | DATETIME(3)     | NOT NULL                        |                             |
| `deleted_at`  | DATETIME(3)     | NULL, INDEX                     | Hard delete (real deletion) |

**GORM Model:**

```go
type Channel struct {
    ID         uint64         `gorm:"primaryKey;autoIncrement"`
    UserID     uint64         `gorm:"uniqueIndex;not null"`
    MonthlyFee float64        `gorm:"type:decimal(10,2);not null;default:0"`
    IsHidden   bool           `gorm:"not null;default:false"`
    CreatedAt  time.Time
    UpdatedAt  time.Time
    DeletedAt  gorm.DeletedAt `gorm:"index"`

    // Relations
    User        User          `gorm:"foreignKey:UserID"`
    Memberships []Membership  `gorm:"foreignKey:ChannelID"`
}
```

---

### 3. `categories`

Video categories.

| Column       | Type            | Constraints        | Description           |
| ------------ | --------------- | ------------------ | --------------------- |
| `id`         | BIGINT UNSIGNED | PK, AUTO_INCREMENT | Category ID           |
| `name`       | VARCHAR(50)     | UNIQUE, NOT NULL   | Category display name |
| `slug`       | VARCHAR(50)     | UNIQUE, NOT NULL   | URL-friendly slug     |
| `created_at` | DATETIME(3)     | NOT NULL           |                       |
| `updated_at` | DATETIME(3)     | NOT NULL           |                       |

**GORM Model:**

```go
type Category struct {
    ID        uint64    `gorm:"primaryKey;autoIncrement"`
    Name      string    `gorm:"type:varchar(50);uniqueIndex;not null"`
    Slug      string    `gorm:"type:varchar(50);uniqueIndex;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time

    // Relations
    Videos    []Video   `gorm:"foreignKey:CategoryID"`
}
```

---

### 4. `videos`

Video metadata. Actual file stored in MinIO.

| Column             | Type            | Constraints                         | Description                                        |
| ------------------ | --------------- | ----------------------------------- | -------------------------------------------------- |
| `id`               | BIGINT UNSIGNED | PK, AUTO_INCREMENT                  | Video ID                                           |
| `user_id`          | BIGINT UNSIGNED | FK → users.id, NOT NULL, INDEX      | Uploader                                           |
| `category_id`      | BIGINT UNSIGNED | FK → categories.id, NOT NULL, INDEX | Category                                           |
| `title`            | VARCHAR(200)    | NOT NULL                            | Video title                                        |
| `description`      | TEXT            | NULL                                | Video description                                  |
| `video_url`        | VARCHAR(500)    | NOT NULL                            | MinIO video file URL                               |
| `thumbnail_url`    | VARCHAR(500)    | NULL                                | Thumbnail image URL                                |
| `duration`         | INT UNSIGNED    | NOT NULL, DEFAULT 0                 | Duration in seconds                                |
| `views_member`     | BIGINT UNSIGNED | NOT NULL, DEFAULT 0                 | Views from members                                 |
| `views_non_member` | BIGINT UNSIGNED | NOT NULL, DEFAULT 0                 | Views from non-members                             |
| `access_tier`      | TINYINT         | NOT NULL, DEFAULT 0                 | 0=public, 1=subscriber (tier1+), 2=premium (tier2) |
| `is_published`     | TINYINT(1)      | NOT NULL, DEFAULT 1                 | Published/unpublished (上架/下架)                  |
| `is_hidden`        | TINYINT(1)      | NOT NULL, DEFAULT 0                 | Hidden by admin (invisible but data preserved)     |
| `created_at`       | DATETIME(3)     | NOT NULL, INDEX                     | Upload time                                        |
| `updated_at`       | DATETIME(3)     | NOT NULL                            |                                                    |
| `deleted_at`       | DATETIME(3)     | NULL, INDEX                         | Hard delete                                        |

**GORM Model:**

```go
type Video struct {
    ID             uint64         `gorm:"primaryKey;autoIncrement"`
    UserID         uint64         `gorm:"index;not null"`
    CategoryID     uint64         `gorm:"index;not null"`
    Title          string         `gorm:"type:varchar(200);not null"`
    Description    *string        `gorm:"type:text"`
    VideoURL       string         `gorm:"type:varchar(500);not null"`
    ThumbnailURL   *string        `gorm:"type:varchar(500)"`
    Duration       uint32         `gorm:"not null;default:0"`
    ViewsMember    uint64         `gorm:"not null;default:0"`
    ViewsNonMember uint64         `gorm:"not null;default:0"`
    AccessTier     int8           `gorm:"not null;default:0"` // 0=public, 1=subscriber, 2=premium
    IsPublished    bool           `gorm:"not null;default:true"`
    IsHidden       bool           `gorm:"not null;default:false"`
    CreatedAt      time.Time      `gorm:"index"`
    UpdatedAt      time.Time
    DeletedAt      gorm.DeletedAt `gorm:"index"`

    // Relations
    User           User           `gorm:"foreignKey:UserID"`
    Category       Category       `gorm:"foreignKey:CategoryID"`
    Tags           []Tag          `gorm:"many2many:video_tags"`
}
```

---

### 5. `tags`

Tags for video categorization and recommendation. Used as the primary promotion/discovery mechanism.

| Column       | Type            | Constraints        | Description       |
| ------------ | --------------- | ------------------ | ----------------- |
| `id`         | BIGINT UNSIGNED | PK, AUTO_INCREMENT | Tag ID            |
| `name`       | VARCHAR(50)     | UNIQUE, NOT NULL   | Tag display name  |
| `slug`       | VARCHAR(50)     | UNIQUE, NOT NULL   | URL-friendly slug |
| `created_at` | DATETIME(3)     | NOT NULL           |                   |
| `updated_at` | DATETIME(3)     | NOT NULL           |                   |

**GORM Model:**

```go
type Tag struct {
    ID        uint64    `gorm:"primaryKey;autoIncrement"`
    Name      string    `gorm:"type:varchar(50);uniqueIndex;not null"`
    Slug      string    `gorm:"type:varchar(50);uniqueIndex;not null"`
    CreatedAt time.Time
    UpdatedAt time.Time

    // Relations
    Videos    []Video   `gorm:"many2many:video_tags"`
}
```

---

### 6. `video_tags`

Junction table: many-to-many relationship between videos and tags.

| Column     | Type            | Constraints              | Description |
| ---------- | --------------- | ------------------------ | ----------- |
| `video_id` | BIGINT UNSIGNED | FK → videos.id, NOT NULL | Video       |
| `tag_id`   | BIGINT UNSIGNED | FK → tags.id, NOT NULL   | Tag         |

**Primary key:** (`video_id`, `tag_id`)

---

### 7. `user_tag_preferences`

Stores each user's selected tags (max 5) for personalized recommendations. Supports both registered users and guests (via session_id).

| Column       | Type            | Constraints            | Description                      |
| ------------ | --------------- | ---------------------- | -------------------------------- |
| `id`         | BIGINT UNSIGNED | PK, AUTO_INCREMENT     |                                  |
| `user_id`    | BIGINT UNSIGNED | FK → users.id, NULL    | Registered user (NULL for guest) |
| `tag_id`     | BIGINT UNSIGNED | FK → tags.id, NOT NULL | Selected tag                     |
| `session_id` | VARCHAR(100)    | NULL, INDEX            | Guest session identifier         |
| `created_at` | DATETIME(3)     | NOT NULL               |                                  |
| `updated_at` | DATETIME(3)     | NOT NULL               |                                  |

**Constraints:**

- Max 5 tag preferences per user/session (enforced at application level)
- Unique: (`user_id`, `tag_id`) for registered users
- Unique: (`session_id`, `tag_id`) for guests

**GORM Model:**

```go
type UserTagPreference struct {
    ID        uint64    `gorm:"primaryKey;autoIncrement"`
    UserID    *uint64   `gorm:"index"`
    TagID     uint64    `gorm:"not null"`
    SessionID *string   `gorm:"type:varchar(100);index"`
    CreatedAt time.Time
    UpdatedAt time.Time

    // Relations
    User      *User     `gorm:"foreignKey:UserID"`
    Tag       Tag       `gorm:"foreignKey:TagID"`
}
```

---

### 8. `memberships`

Tracks which users are members of which channels. Supports two tiers: **Tier 1** (free subscriber) and **Tier 2** (paid premium via Paddle subscription).

| Column                   | Type            | Constraints                | Description                                  |
| ------------------------ | --------------- | -------------------------- | -------------------------------------------- |
| `id`                     | BIGINT UNSIGNED | PK, AUTO_INCREMENT         |                                              |
| `channel_id`             | BIGINT UNSIGNED | FK → channels.id, NOT NULL | Target channel                               |
| `user_id`                | BIGINT UNSIGNED | FK → users.id, NOT NULL    | Subscribing user                             |
| `tier`                   | TINYINT         | NOT NULL, DEFAULT 1        | 1 = free subscriber, 2 = paid premium        |
| `status`                 | VARCHAR(20)     | NOT NULL, DEFAULT 'active' | active / cancelled                           |
| `paddle_subscription_id` | VARCHAR(50)     | UNIQUE, NULL               | Paddle subscription ID (`sub_...`) tier 2    |
| `paddle_status`          | VARCHAR(20)     | NULL                       | Paddle sub status: active/past_due/cancelled |
| `started_at`             | DATETIME(3)     | NOT NULL                   | Membership start                             |
| `expires_at`             | DATETIME(3)     | NULL                       | Membership expiry (tier 2)                   |
| `created_at`             | DATETIME(3)     | NOT NULL                   |                                              |
| `updated_at`             | DATETIME(3)     | NOT NULL                   |                                              |

**Tier behaviour:**

- **Tier 1 (free):** User subscribes to a channel for free. Can watch public + `access_tier=1` videos. Receives NATS notifications when the channel publishes new content.
- **Tier 2 (paid premium):** User pays a monthly fee (via Paddle recurring subscription). Can watch all videos including `access_tier=2` (premium-exclusive). The Paddle subscription is created when the user upgrades; `paddle_subscription_id` and `paddle_status` are updated via webhooks.

**Unique constraint:** (`channel_id`, `user_id`) — one membership per user per channel.

**GORM Model:**

```go
type Membership struct {
    ID                    uint64    `gorm:"primaryKey;autoIncrement"`
    ChannelID             uint64    `gorm:"not null;uniqueIndex:idx_channel_user"`
    UserID                uint64    `gorm:"not null;uniqueIndex:idx_channel_user"`
    Tier                  int8      `gorm:"not null;default:1"` // 1=free, 2=premium
    Status                string    `gorm:"type:varchar(20);not null;default:'active'"`
    PaddleSubscriptionID  *string   `gorm:"type:varchar(50);uniqueIndex"`
    PaddleStatus          *string   `gorm:"type:varchar(20)"`
    StartedAt             time.Time `gorm:"not null"`
    ExpiresAt             *time.Time
    CreatedAt             time.Time
    UpdatedAt             time.Time

    // Relations
    Channel   Channel   `gorm:"foreignKey:ChannelID"`
    User      User      `gorm:"foreignKey:UserID"`
}
```

---

### 9. `view_records`

Individual view tracking for analytics (optional — for detailed analytics).

| Column      | Type            | Constraints                     | Description           |
| ----------- | --------------- | ------------------------------- | --------------------- |
| `id`        | BIGINT UNSIGNED | PK, AUTO_INCREMENT              |                       |
| `video_id`  | BIGINT UNSIGNED | FK → videos.id, NOT NULL, INDEX | Viewed video          |
| `user_id`   | BIGINT UNSIGNED | FK → users.id, NULL             | Viewer (NULL = guest) |
| `is_member` | TINYINT(1)      | NOT NULL, DEFAULT 0             | Was viewer a member   |
| `viewed_at` | DATETIME(3)     | NOT NULL, INDEX                 | View timestamp        |

**GORM Model:**

```go
type ViewRecord struct {
    ID       uint64    `gorm:"primaryKey;autoIncrement"`
    VideoID  uint64    `gorm:"index;not null"`
    UserID   *uint64   `gorm:"index"`
    IsMember bool      `gorm:"not null;default:false"`
    ViewedAt time.Time `gorm:"index;not null"`

    // Relations
    Video    Video     `gorm:"foreignKey:VideoID"`
    User     *User     `gorm:"foreignKey:UserID"`
}
```

---

### 11. `notifications`

Persists real-time notifications delivered via NATS. Created when a subscribed channel publishes or updates a video.

| Column       | Type            | Constraints                    | Description                                   |
| ------------ | --------------- | ------------------------------ | --------------------------------------------- |
| `id`         | BIGINT UNSIGNED | PK, AUTO_INCREMENT             | Notification ID                               |
| `user_id`    | BIGINT UNSIGNED | FK → users.id, NOT NULL, INDEX | Recipient user                                |
| `type`       | VARCHAR(30)     | NOT NULL                       | `new_video` / `video_update` / `subscription` |
| `title`      | VARCHAR(200)    | NOT NULL                       | Notification title                            |
| `message`    | TEXT            | NULL                           | Notification body text                        |
| `payload`    | JSON            | NULL                           | Extra data (channel_id, video_id, etc.)       |
| `is_read`    | TINYINT(1)      | NOT NULL, DEFAULT 0            | Read/unread flag                              |
| `created_at` | DATETIME(3)     | NOT NULL, INDEX                |                                               |

**GORM Model:**

```go
type Notification struct {
    ID        uint64          `gorm:"primaryKey;autoIncrement"`
    UserID    uint64          `gorm:"index;not null"`
    Type      string          `gorm:"type:varchar(30);not null"`
    Title     string          `gorm:"type:varchar(200);not null"`
    Message   *string         `gorm:"type:text"`
    Payload   datatypes.JSON  `gorm:"type:json"` // github.com/go-gorm/datatypes
    IsRead    bool            `gorm:"not null;default:false"`
    CreatedAt time.Time       `gorm:"index"`

    // Relations
    User      User            `gorm:"foreignKey:UserID"`
}
```

---

### 12. `donations`

Tracks donation payments from users to creators via Paddle (sandbox). Each row represents a single one-time donation transaction.

| Column                  | Type            | Constraints                    | Description                                |
| ----------------------- | --------------- | ------------------------------ | ------------------------------------------ |
| `id`                    | BIGINT UNSIGNED | PK, AUTO_INCREMENT             | Donation ID                                |
| `donor_id`              | BIGINT UNSIGNED | FK → users.id, NOT NULL, INDEX | User who donated                           |
| `creator_id`            | BIGINT UNSIGNED | FK → users.id, NOT NULL, INDEX | Creator who received the donation          |
| `amount`                | DECIMAL(10,2)   | NOT NULL                       | Donation amount                            |
| `currency`              | VARCHAR(3)      | NOT NULL, DEFAULT 'USD'        | ISO 4217 currency code                     |
| `message`               | TEXT            | NULL                           | Optional message from donor                |
| `paddle_transaction_id` | VARCHAR(50)     | UNIQUE, NULL                   | Paddle transaction ID (`txn_...`)          |
| `paddle_status`         | VARCHAR(20)     | NOT NULL, DEFAULT 'pending'    | pending / completed / refunded / cancelled |
| `created_at`            | DATETIME(3)     | NOT NULL                       |                                            |
| `updated_at`            | DATETIME(3)     | NOT NULL                       |                                            |

**Paddle integration notes:**

- Use Paddle **sandbox** environment (`sandbox-api.paddle.com`) for development
- Each donation creates a Paddle "transaction" via the Paddle API with a one-time price
- `paddle_status` is updated via Paddle webhooks (`transaction.completed`, `transaction.payment_failed`)
- `paddle_transaction_id` is the Paddle-assigned `txn_*` identifier

**GORM Model:**

```go
type Donation struct {
    ID                   uint64    `gorm:"primaryKey;autoIncrement"`
    DonorID              uint64    `gorm:"index;not null"`
    CreatorID            uint64    `gorm:"index;not null"`
    Amount               float64   `gorm:"type:decimal(10,2);not null"`
    Currency             string    `gorm:"type:varchar(3);not null;default:'USD'"`
    Message              *string   `gorm:"type:text"`
    PaddleTransactionID  *string   `gorm:"type:varchar(50);uniqueIndex"`
    PaddleStatus         string    `gorm:"type:varchar(20);not null;default:'pending'"`
    CreatedAt            time.Time
    UpdatedAt            time.Time

    // Relations
    Donor                User      `gorm:"foreignKey:DonorID"`
    Creator              User      `gorm:"foreignKey:CreatorID"`
}
```

---

## Indexes

### Composite & Performance Indexes

```sql
-- videos: search & filter queries
CREATE INDEX idx_videos_category_published ON videos(category_id, is_published, is_hidden, created_at DESC);
CREATE INDEX idx_videos_user_published ON videos(user_id, is_published, is_hidden, created_at DESC);
CREATE INDEX idx_videos_title_fulltext ON videos(title) USING FULLTEXT;
CREATE INDEX idx_videos_hidden ON videos(is_hidden);

-- tags: tag-based recommendation queries
CREATE INDEX idx_video_tags_video ON video_tags(video_id);
CREATE INDEX idx_video_tags_tag ON video_tags(tag_id);
CREATE INDEX idx_user_tag_prefs_user ON user_tag_preferences(user_id);
CREATE INDEX idx_user_tag_prefs_session ON user_tag_preferences(session_id);
CREATE UNIQUE INDEX idx_user_tag_pref_unique ON user_tag_preferences(user_id, tag_id);

-- users: admin queries
CREATE INDEX idx_users_role ON users(role);
CREATE INDEX idx_users_hidden ON users(is_hidden);

-- memberships: join/leave lookup
CREATE UNIQUE INDEX idx_membership_channel_user ON memberships(channel_id, user_id);
CREATE INDEX idx_membership_tier ON memberships(channel_id, tier, status);
CREATE UNIQUE INDEX idx_membership_paddle_sub ON memberships(paddle_subscription_id);

-- notifications: user notification timeline
CREATE INDEX idx_notifications_user_read ON notifications(user_id, is_read, created_at DESC);

-- view_records: analytics aggregation
CREATE INDEX idx_view_records_video_time ON view_records(video_id, viewed_at);
CREATE INDEX idx_view_records_user ON view_records(user_id, viewed_at);

-- donations: payment queries
CREATE INDEX idx_donations_donor ON donations(donor_id, created_at DESC);
CREATE INDEX idx_donations_creator ON donations(creator_id, paddle_status, created_at DESC);
CREATE UNIQUE INDEX idx_donations_paddle_txn ON donations(paddle_transaction_id);
```

---

## GORM Migration Setup

```go
// internal/data/data.go
func NewDB(c *conf.Data, logger log.Logger) *gorm.DB {
    db, err := gorm.Open(mysql.Open(c.Database.Source), &gorm.Config{
        Logger: gormLogger.Default.LogMode(gormLogger.Info),
        NamingStrategy: schema.NamingStrategy{
            SingularTable: false,  // use plural table names
        },
    })
    if err != nil {
        log.Fatalf("failed to connect database: %v", err)
    }

    sqlDB, _ := db.DB()
    sqlDB.SetMaxIdleConns(int(c.Database.MaxIdleConns))
    sqlDB.SetMaxOpenConns(int(c.Database.MaxOpenConns))
    sqlDB.SetConnMaxLifetime(c.Database.ConnMaxLifetime.AsDuration())

    // Auto migrate
    if err := db.AutoMigrate(
        &model.User{},
        &model.Channel{},
        &model.Category{},
        &model.Video{},
        &model.Tag{},
        &model.UserTagPreference{},
        &model.Membership{},
        &model.ViewRecord{},
        &model.Notification{},
        &model.Donation{},
    ); err != nil {
        log.Fatalf("failed to migrate database: %v", err)
    }

    return db
}
```

---

## Seed Data

```sql
-- init.sql (loaded via docker-entrypoint-initdb.d)

CREATE DATABASE IF NOT EXISTS fenzvideo
  CHARACTER SET utf8mb4
  COLLATE utf8mb4_unicode_ci;

USE fenzvideo;

-- Seed categories
INSERT INTO categories (name, slug, created_at, updated_at) VALUES
  ('音樂',     'music',         NOW(), NOW()),
  ('遊戲',     'gaming',        NOW(), NOW()),
  ('教育',     'education',     NOW(), NOW()),
  ('娛樂',     'entertainment', NOW(), NOW()),
  ('科技',     'technology',    NOW(), NOW()),
  ('運動',     'sports',        NOW(), NOW()),
  ('新聞',     'news',          NOW(), NOW()),
  ('美食',     'food',          NOW(), NOW()),
  ('旅遊',     'travel',        NOW(), NOW()),
  ('生活',     'lifestyle',     NOW(), NOW());

-- Seed tags (promotion / discovery tags)
INSERT INTO tags (name, slug, created_at, updated_at) VALUES
  ('搞笑',       'funny',           NOW(), NOW()),
  ('教學',       'tutorial',        NOW(), NOW()),
  ('Vlog',       'vlog',            NOW(), NOW()),
  ('開箱',       'unboxing',        NOW(), NOW()),
  ('直播精華',   'stream-highlight', NOW(), NOW()),
  ('音樂MV',     'music-video',     NOW(), NOW()),
  ('遊戲實況',   'gameplay',        NOW(), NOW()),
  ('美食料理',   'cooking',         NOW(), NOW()),
  ('旅行紀錄',   'travel-log',      NOW(), NOW()),
  ('科技評測',   'tech-review',     NOW(), NOW()),
  ('新手入門',   'beginner',        NOW(), NOW()),
  ('健身運動',   'fitness',         NOW(), NOW()),
  ('動畫',       'animation',       NOW(), NOW()),
  ('訪談',       'interview',       NOW(), NOW()),
  ('DIY手作',    'diy',             NOW(), NOW());

-- Seed admin account (password: admin123, bcrypt hash)
INSERT INTO users (username, display_name, password, role, created_at, updated_at) VALUES
  ('admin', 'System Admin', '$2a$10$PLACEHOLDER_BCRYPT_HASH', 'admin', NOW(), NOW());
```

---

## Analytics Queries

### Tag-Based Recommendation (Random Pick from User's Tags)

The recommendation system randomly selects a subset of the user's chosen tags (1 to N from max 5), then fetches random published, non-hidden videos matching those tags.

```sql
-- Step 1: Get user's selected tag IDs (for registered user)
SELECT tag_id FROM user_tag_preferences WHERE user_id = ?;
-- For guest: SELECT tag_id FROM user_tag_preferences WHERE session_id = ?;

-- Step 2: Randomly pick a combination (1–N tags from the user's selection)
-- This is done at the application layer using Go's math/rand

-- Step 3: Fetch random videos matching ANY of the selected tag combination
SELECT DISTINCT v.*
FROM videos v
INNER JOIN video_tags vt ON vt.video_id = v.id
WHERE vt.tag_id IN (?, ?, ?)    -- randomly chosen subset of user's tags
  AND v.is_published = 1
  AND v.is_hidden = 0
  AND v.deleted_at IS NULL
ORDER BY RAND()
LIMIT 20;
```

### Fallback (No Tags Selected)

If a user has not selected any tags, show globally random published videos:

```sql
SELECT * FROM videos
WHERE is_published = 1 AND is_hidden = 0 AND deleted_at IS NULL
ORDER BY RAND()
LIMIT 20;
```

### Total Views (Member vs Non-Member)

```sql
SELECT
  SUM(views_member) AS total_member_views,
  SUM(views_non_member) AS total_non_member_views
FROM videos
WHERE user_id = ? AND deleted_at IS NULL;
```

### Views Ranking

```sql
SELECT id, title, (views_member + views_non_member) AS total_views,
       views_member, views_non_member
FROM videos
WHERE user_id = ? AND deleted_at IS NULL
ORDER BY total_views DESC
LIMIT 10;
```

### Member Count

```sql
SELECT COUNT(*) AS member_count
FROM memberships
WHERE channel_id = ? AND status = 'active';
```

### Member / Non-Member View Ratio

```sql
SELECT
  SUM(views_member) AS member_views,
  SUM(views_non_member) AS non_member_views,
  SUM(views_member) / NULLIF(SUM(views_member) + SUM(views_non_member), 0) AS member_ratio
FROM videos
WHERE user_id = ? AND deleted_at IS NULL;
```

### Channel Revenue (Membership + Donations)

```sql
-- Membership revenue
SELECT
  c.monthly_fee * COUNT(m.id) AS monthly_membership_revenue
FROM channels c
LEFT JOIN memberships m ON m.channel_id = c.id AND m.status = 'active'
WHERE c.user_id = ?
GROUP BY c.id;

-- Donation revenue (total received by creator)
SELECT
  COALESCE(SUM(amount), 0) AS total_donation_revenue
FROM donations
WHERE creator_id = ? AND paddle_status = 'completed';

-- Recent donations received (for dashboard)
SELECT d.id, d.amount, d.currency, d.message, d.created_at,
       u.display_name AS donor_name
FROM donations d
INNER JOIN users u ON u.id = d.donor_id
WHERE d.creator_id = ? AND d.paddle_status = 'completed'
ORDER BY d.created_at DESC
LIMIT 20;
```

---

## Data Flow Diagram

```
  Vue 3 Frontend
       │
       ▼
  HTTP API (Kratos)
       │
       ├───────────────────────▶ NATS (message broker)
       │                           │
       ▼                           ▼
  ┌─────────────────────────────────────────────┐
  │              GORM ORM Layer                 │
  │  ┌─────────┐ ┌────────┐ ┌───────────────┐  │
  │  │  User   │ │ Video  │ │  Membership   │  │
  │  │  Model  │ │ Model  │ │    Model      │  │
  │  └────┬────┘ └───┬────┘ └──────┬────────┘  │
  │       │          │             │            │
  │  ┌─────────┐ ┌───────┐ ┌──────────────┐  │
  │  │  Tag    │ │ UserTag│ │ Notification │  │
  │  │  Model  │ │ Pref   │ │    Model     │  │
  │  └────┬────┘ └───┬───┘ └──────┬───────┘  │
  │       │          │             │            │
  │       ▼          ▼             ▼            │
  │  ┌──────────────────────────────────────┐   │
  │  │         MySQL 8.0 (InnoDB)          │   │
  │  │                                      │   │
  │  │  users │ channels │ videos           │   │
  │  │  categories │ memberships            │   │
  │  │  tags │ video_tags                   │   │
  │  │  user_tag_preferences │ donations    │   │
  │  │  view_records │ notifications         │   │
  │  └──────────────────────────────────────┘   │
  └─────────────────────────────────────────────┘
       │
       ▼
  ┌──────────────┐
  │    MinIO     │  ← Video & Thumbnail files
  └──────────────┘
```

---

## Key Design Decisions

| Decision                                     | Rationale                                                                                                                       |
| -------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------- |
| **Two-tier delete: Hidden + Real**           | Hidden (`is_hidden`) preserves data for admin review/restore; Real delete (`deleted_at`) permanently removes from all queries   |
| **`role` field on users**                    | Enables admin functionality; `user` = normal, `admin` = full CRUD power over all accounts                                       |
| **`is_hidden` on users, channels, videos**   | Admin can hide content without destroying data; reversible moderation action                                                    |
| **Tag-based recommendation**                 | Users (registered or guest) pick up to 5 tags; system randomly combines 1–N tags to fetch diverse videos                        |
| **`user_tag_preferences` with `session_id`** | Supports guest tag selection via browser session; registered users use `user_id`                                                |
| **`video_tags` many-to-many**                | Each video can have multiple tags; enables flexible tag-based discovery                                                         |
| **Max 5 tags per user**                      | Keeps recommendation focused; enforced at application level                                                                     |
| **Denormalized view counts** on `videos`     | Avoid expensive COUNT queries on `view_records` for every page load                                                             |
| **Separate `view_records` table**            | Enables time-series analytics & detailed reporting                                                                              |
| **Two-tier membership (Tier 1 + Tier 2)**    | Tier 1 is free subscribe; Tier 2 is paid premium via Paddle recurring subscription. `access_tier` on videos controls visibility |
| **`access_tier` replaces `is_member_only`**  | Three levels: 0=public (anyone), 1=subscriber (tier 1+2), 2=premium (tier 2 only). More granular access control                 |
| **Paddle subscription for Tier 2**           | Recurring monthly payment managed by Paddle; `paddle_subscription_id` + `paddle_status` on memberships track lifecycle          |
| **NATS for real-time notifications**         | Channel events (new video, update) published to NATS; subscribers receive push notifications persisted in `notifications` table |
| **`notifications` table**                    | Persists events so users can view history; `is_read` flag supports unread count badge; `payload` JSON for flexible data         |
| **`channels` as separate table**             | Decouples channel settings (fee) from user profile; allows future expansion                                                     |
| **Composite unique index** on memberships    | Prevents duplicate memberships per user-channel pair                                                                            |
| **FULLTEXT index** on video title            | Enables efficient MySQL full-text search                                                                                        |
| **`is_published` flag**                      | Supports 上架/下架 without deleting; only `is_published=false` videos can be hard-deleted                                       |
| **`DECIMAL(10,2)` for fee**                  | Avoids floating-point precision issues for monetary values                                                                      |
| **User self-delete**                         | Users can delete their own account + channel; cascades to hide all their videos                                                 |
| **Paddle sandbox for donations**             | Use Paddle API (sandbox) for payment processing; no need to handle PCI compliance; webhook-driven status updates                |
| **Donations as separate table**              | Decoupled from memberships; one-time transactions tracked independently with Paddle transaction ID                              |
| **`paddle_status` state machine**            | `pending` → `completed` or `cancelled`/`refunded`; updated only via verified Paddle webhooks                                    |

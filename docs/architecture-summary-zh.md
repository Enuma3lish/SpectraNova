# FenzVideo 架構摘要（至 MVP 現況）

## 專案定位

以 **標籤（Tag）** 為核心的影片串流平台，未來支援會員制與贊助功能。

---

## 技術棧

| 層級 | 技術 |
|------|------|
| 後端 | Go 1.24 + Kratos v2（gRPC/HTTP）、Wire DI、GORM v2 |
| 前端 | Vue 3 + Vite + TypeScript、Element Plus、Tailwind CSS、Video.js |
| 資料庫 | MySQL 8.0（InnoDB、utf8mb4） |
| 快取 | Redis 7（標籤推薦快取 + 觀看次數緩衝） |
| 儲存 | MinIO（S3 相容，存放影片與縮圖） |
| 訊息 | NATS 2（目前僅初始化，Phase 5 才啟用通知） |
| 容器 | Docker Compose（7 個服務） |

---

## 架構策略

**模組化單體** — 所有服務在同一個 Go 二進位中，未來有瓶頸時才拆成微服務。

### 限界上下文

| 上下文 | 服務 | 資料表 |
|--------|------|--------|
| 身份與存取 | Auth、User、Admin | `users` |
| 內容與探索 | Video、Tag、Search、Category | `videos`、`tags`、`video_tags`、`categories`、`view_records` |
| 變現 | Channel、Donation、Dashboard | `channels`、`memberships`、`donations` |
| 互動 | Notification | `notifications` |

---

## 資料庫設計（12 張表）

### MVP 已使用

- `users` — 使用者帳號（含 `role`、`is_hidden`、軟刪除）
- `channels` — 每位使用者一個頻道（註冊時自動建立）
- `categories` — 影片分類（10 個）
- `videos` — 影片元資料（檔案存於 MinIO）
- `tags` — 標籤（15 個）
- `video_tags` — 影片與標籤多對多關聯
- `user_tag_preferences` — 使用者標籤偏好（最多 5 個，支援訪客 `session_id`）

### 未來使用

- `memberships` — 頻道會員（Tier 1 免費訂閱 / Tier 2 付費）
- `view_records` — 觀看紀錄（時序分析用）
- `notifications` — 通知（NATS 驅動）
- `donations` — 贊助（Paddle 付款）

### 關鍵設計決策

| 決策 | 原因 |
|------|------|
| 雙層刪除：`is_hidden` + `deleted_at` | 隱藏可復原；真刪除永久移除 |
| 標籤推薦（最多 5 個） | 隨機組合 1~N 個標籤撈取多樣影片 |
| `session_id` 訪客支援 | 未登入也能設定標籤偏好 |
| 非正規化觀看次數（`views_member` / `views_non_member`） | 避免每次載入頁面都 COUNT `view_records` |
| FULLTEXT 索引（`videos.title`） | MySQL 原生全文搜尋 |
| `access_tier`（0/1/2） | 公開 / 訂閱者 / 付費會員，三級存取控制 |

---

## 後端架構

### 分層結構（Kratos Clean Architecture）

```
api/          → Protobuf 定義（gRPC + HTTP）
internal/
  biz/        → 業務邏輯（Usecase + Repo 介面）
  data/       → 資料存取（GORM、Redis、MinIO 實作）
  service/    → gRPC/HTTP Handler（串接 biz 層）
  server/     → HTTP/gRPC Server 設定、中介層
  pkg/        → 內部共用套件（authctx、JWT、bcrypt、MinIO）
cmd/
  backend/    → 主程式進入點
  seed/       → 資料種子腳本
```

### 已實作的 6 個服務

| 服務 | 功能 |
|------|------|
| AuthService | 登入、註冊、Token 刷新 |
| CategoryService | 列出分類 |
| TagService | 列出標籤、取得/設定使用者偏好 |
| VideoService | CRUD、推薦、上下架、觀看計數 |
| SearchService | FULLTEXT 搜尋 + 多條件篩選 |
| ChannelService | 頻道資訊、免費訂閱/取消 |

### Redis 快取機制

| 鍵 | 結構 | 用途 |
|----|------|------|
| `tag:{id}` | SET | 每個標籤對應的影片 ID 集合（24h TTL） |
| `video:{id}` | HASH | 影片摘要資料（24h TTL） |
| `popular:global` | ZSET | 依觀看次數排序（10min TTL） |
| `views:buffer` | HASH | 觀看次數緩衝（每 30 秒批次寫入 MySQL） |
| `cleanup:queue` | LIST | 失敗的快取清除任務（背景重試） |

- 啟動時 **快取預熱**（MySQL → Redis），消除冷啟動
- 快取優先讀取，Miss 時 fallback 到 MySQL

---

## 前端架構

### 目錄結構

```
src/
  api/          → Axios 端點模組（auth、video、search 等）
  components/   → 可重用元件（VideoCard、TagSelector 等）
  layouts/      → 佈局（Default、Auth、Admin）
  router/       → 路由定義 + 導航守衛
  stores/       → Pinia 狀態管理（6 個 Store）
  types/        → TypeScript 型別定義
  views/        → 頁面元件（綁定路由）
```

### 頁面

| 頁面 | 功能 |
|------|------|
| LoginView | 登入 / 註冊（分頁切換） |
| HomeView | 標籤推薦影片 + 分頁 |
| VideoView | HTML5 影片播放器 + 影片資訊 |
| SearchResultsView | 搜尋 + 進階篩選 |
| CategoryView | 依分類瀏覽 |
| ChannelView | 頻道資訊 + 訂閱 |
| AdminUsersView | 使用者管理 |
| AdminTagsView | 標籤 CRUD |

### 狀態管理（Pinia）

| Store | 職責 |
|-------|------|
| authStore | JWT Token、使用者資訊、登入/管理員狀態 |
| videoStore | 推薦影片、當前影片 |
| tagStore | 所有標籤、已選標籤、訪客 sessionId |
| searchStore | 搜尋查詢、篩選條件、結果 |
| categoryStore | 分類列表 |
| adminStore | 管理員操作（使用者/標籤/影片） |

---

## 部署架構

### Docker Compose（7 個服務）

```
fenzvideo-frontend   → Nginx（SPA + 反向代理）
fenzvideo-backend    → Go Kratos 後端
fenzvideo-mysql      → MySQL 8.0
fenzvideo-redis      → Redis 7
fenzvideo-minio      → MinIO（影片/縮圖儲存）
fenzvideo-nats       → NATS 2
fenzvideo-jaeger     → Jaeger（追蹤）
```

### Nginx 設計

- `/api/` → 反向代理至後端
- `/fenzvideo/` → 反向代理至 MinIO（同源避免 CORS）
- 其餘路徑 → SPA fallback（`index.html`）
- 使用 Docker DNS（`127.0.0.11`）+ 變數式 `proxy_pass`（容器重啟不影響）

---

## 種子資料

| 資料表 | 筆數 | 說明 |
|--------|------|------|
| `users` | 6 | 1 管理員 + 5 創作者 |
| `channels` | 6 | 每人一個頻道 |
| `categories` | 10 | 音樂、遊戲、教育、娛樂、科技、運動、新聞、美食、旅遊、生活 |
| `tags` | 15 | 搞笑、教學、Vlog、開箱、直播精華 等 |
| `videos` | 57 | 預設影片，涵蓋所有分類與標籤 |
| `video_tags` | ~140 | 每部影片 2-3 個標籤 |

---

## 開發進度

| Phase | 狀態 | 範圍 |
|-------|------|------|
| Phase 1 | ✅ 完成 | 基礎建設（Docker、Model、中介層、Seed） |
| Phase 2 | ✅ 完成 | 後端核心 MVP（6 個服務 + Redis 快取） |
| Phase 3 | ✅ 完成 | 前端 MVP（Vue 3 SPA + Admin + Docker 部署） |
| Phase 4 | 🔲 待做 | 變現（Paddle 付費會員 + 贊助 + 儀表板） |
| Phase 5 | 🔲 待做 | 進階功能（NATS 通知、使用者自助、可觀測性） |
| Phase 6 | 🔲 待做 | 部署與維運（SSL、CI/CD、監控） |
| Phase 7 | 🔲 待做 | 微服務拆分（依瓶頸決定） |

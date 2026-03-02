# FenzVideo 開發路線圖（設計理由詳解版）

## 專案概述

FenzVideo 是一個基於標籤推薦的影片串流平台，支援會員制與抖內（打賞）功能。後端使用 Go/Kratos，前端使用 Vue 3，基礎設施全部採用開源工具。

---

## 架構策略：模組化單體 → 微服務

**先以模組化單體起步。** 全部 11 個服務都在同一個 Go 執行檔中，使用 Kratos 的 Clean Architecture 架構。只有當出現實際測量到的效能瓶頸時，才將個別服務拆分為微服務。

### 為什麼先用單體？

1. **開發速度快** — 一個 repo、一個 binary、一次部署。不用處理服務間通訊、分散式交易等複雜問題
2. **避免過早優化** — 微服務帶來網路延遲、服務發現、分散式追蹤等額外成本。在流量還小的時候，這些成本遠大於收益
3. **Kratos 的 Clean Architecture 已經做好隔離** — 即使是單體，Transport → Service → Biz ← Data 的分層保證了各模組之間耦合度低。未來拆分時只需把某個模組的 biz + data 搬出去，不用大規模重構
4. **先證明產品可行，再優化架構** — 如果產品本身沒有用戶，再好的微服務架構也沒有意義

### 限界上下文（Bounded Contexts）

| 上下文 | 包含的服務 | 共享的資料表 | 為什麼這樣分？ |
|--------|-----------|-------------|--------------|
| **身份與權限** | Auth, User, Admin | `users` | 這三個服務都操作同一張 `users` 表，共享使用者身份的概念。Auth 負責認證，User 負責自助管理，Admin 負責後台管理 |
| **內容與發現** | Video, Tag, Search, Category | `videos`, `tags`, `video_tags`, `categories`, `view_records` | 這些服務圍繞「影片內容」這個核心領域。Tag 和 Category 是影片的分類維度，Search 是影片的檢索入口 |
| **變現** | Channel, Donation, Dashboard | `channels`, `memberships`, `donations` | 涉及金流的服務放在一起。Channel 管理訂閱，Donation 處理打賞，Dashboard 顯示收入分析 |
| **互動** | Notification | `notifications` | 通知是獨立的寫入密集型服務，未來最可能第一個被拆分出去 |

### 微服務拆分優先序（只有出現瓶頸才拆）

| 優先序 | 服務 | 觸發條件 | 為什麼？ |
|--------|------|---------|---------|
| 第 1 | NotificationService | 粉絲通知的寫入爆發影響主資料庫 | 一個有 10 萬訂閱者的頻道上傳一支影片 → 瞬間產生 10 萬條通知寫入。這是最容易產生寫入瓶頸的服務 |
| 第 2 | 影片上傳子系統 | 上傳流量佔滿主 API 頻寬 | 影片檔案很大，上傳是 CPU/IO 密集型操作，容易拖慢主 API |
| 第 3 | SearchService | MySQL FULLTEXT 無法滿足需求 | 當影片數量超過百萬，MySQL 全文搜尋的效能會下降，需要遷移到 Elasticsearch 或 Meilisearch |

---

## Phase 1: 基礎建設 ✅

> **目標**：應用程式能啟動、連接所有基礎設施、灌入種子資料、回應健康檢查。

### 為什麼先做基礎建設？

在寫任何業務邏輯之前，必須先確保：
- 資料庫連線正常、schema 已建立（GORM AutoMigrate）
- Redis、MinIO、NATS 等基礎設施都能連上
- JWT 認證和權限中間件已就位（後續每個 API 都需要）
- 有種子資料可以測試推薦演算法

如果跳過這一步直接寫 API，會在每個服務都遇到「連線沒建好」「middleware 沒設定」的問題，反覆踩坑。

### 已完成項目

- [x] **擴展 `conf.proto`**：加入 Auth、Storage、Paddle、NATS 的設定區塊
  - **為什麼？** Kratos 使用 Protobuf 定義設定結構，這樣設定就有型別檢查，不會打錯欄位名

- [x] **`docker-compose.yaml`**：MySQL、Redis、MinIO、NATS + 監控堆疊
  - **為什麼？** 一個 `docker-compose up -d` 就能把所有基礎設施跑起來，不用手動安裝每個服務

- [x] **所有 GORM 模型**（12 張表）
  - **為什麼先定義所有模型？** 資料模型是整個系統的骨架。先把所有表的結構想清楚（包括欄位型別、索引、外鍵關係），後面寫 API 時就不用反覆改 schema
  - 12 張表：users, channels, videos, categories, tags, video_tags, user_tag_preferences, memberships, view_records, notifications, donations

- [x] **Data 層初始化**：GORM + Redis + MinIO + NATS 客戶端，AutoMigrate
  - **為什麼？** 所有基礎設施的連線集中在 `data.go` 管理，遵循 Kratos 的 Data 層職責

- [x] **內部工具包**：JWT、bcrypt hash、MinIO 上傳、分頁
  - **為什麼？** 這些是跨服務共用的工具函式。先寫好，後續每個服務直接引用，避免重複程式碼

- [x] **中間件**：JWT 認證、Admin 守衛、CORS
  - **為什麼在 Phase 1 就做？** 認證中間件是幾乎所有 API 的前置條件。先做好，Phase 2 寫 API 時直接掛上就好

- [x] **錯誤代碼 proto**：定義所有錯誤碼
  - **為什麼？** Kratos 用 Protobuf 定義錯誤碼，讓前後端有統一的錯誤定義。先定義好全部錯誤碼，後續寫業務邏輯時直接引用

- [x] **種子資料產生器** (`cmd/seed/main.go`)
  - **為什麼需要？** 開發和測試推薦演算法需要真實感的測試資料。如果手動建資料很麻煩，用 Gemini API 自動生成繁體中文的影片標題和描述，省時省力
  - **為什麼用 Gemini API？** 手寫 15 組有意義的繁體中文影片標題和描述很耗時。用 AI 生成可以快速得到多樣化、有意義的測試資料
  - **為什麼設計成冪等（idempotent）？** 種子腳本可能被重複執行（例如開發者重新跑 `make seed`），冪等設計確保不會產生重複資料

- [x] **快取預熱** (`internal/data/cache_warmup.go`)
  - **為什麼需要？** 如果不預熱，系統啟動後第一批使用者的請求全部是 cache miss，會同時衝擊 MySQL（冷啟動問題）。在 `NewData()` 階段把所有公開影片載入 Redis，伺服器開始接受流量時快取已經是熱的
  - **為什麼不需要額外的追蹤表？** MySQL 的 `videos` + `video_tags` 表本身就是快取的藍圖。任何時候都可以從這兩張表重建 Redis 快取，不需要額外的表來記錄「哪些影片在快取中」
  - **為什麼放在 app boot 而不是 seed script？** seed script 只跑一次，但 Redis 可能重啟。把預熱放在 app boot，每次應用程式啟動都會自動重建快取

---

## Phase 2: 核心 MVP ✅

> **目標**：使用者可以註冊、登入、上傳影片、根據標籤瀏覽推薦、搜尋影片。

### 為什麼這些服務一起做？

這些是一個影片平台最基本的功能：沒有認證就不能上傳，沒有影片就不能推薦，沒有標籤就不能個人化。這些服務有強依賴關係，必須一起完成才能形成一個可用的 MVP。

### 基礎建設變更

- [x] 中間件從 HTTP 路徑改為 Kratos operation 名稱（`/fenzvideo.v1.AuthService/Login`）
- [x] 公開端點現在會嘗試提取可選的 JWT token（例如 `GetVideo`、`GetRecommended` 可以識別已登入使用者）
- [x] 將 `UserIDFromContext`/`RoleFromContext` 抽取到 `internal/pkg/authctx` 以解決 `server` 和 `service` 之間的循環引用
- [x] 二階段檔案上傳端點：`POST /api/v1/upload/video` 和 `POST /api/v1/upload/thumbnail`（MinIO）
- [x] Wire 依賴注入整合：6 個服務 + `MembershipChecker` 適配器 + `MinIOUploader` + `VideoCache`
- [x] 手動建立 `videos(title)` 的 FULLTEXT 索引（GORM AutoMigrate 無法建立）

### 2.1 AuthService（認證服務）

- [x] 定義 `auth.proto`（Login, Register, RefreshToken）
- [x] 業務邏輯層（AuthUsecase + AuthRepo 介面）
- [x] 資料層（AuthRepo 的 GORM 實作）
- [x] 服務層（gRPC/HTTP handler）
- [x] Wire 依賴注入整合

**為什麼是第一個做的？**
幾乎所有 API 都需要知道「目前請求者是誰」。AuthService 提供 JWT token，後續所有 protected route 都依賴它。

**為什麼用 JWT 而不是 Session？**
- JWT 是無狀態的，不需要在 Redis 中維護 session 資料
- 適合 gRPC + HTTP 雙協議（session 通常只用於 HTTP）
- 可以攜帶 user_id 和 role，middleware 直接解碼就知道權限

**為什麼需要 RefreshToken？**
- Access Token 的 TTL 短（24h），過期後使用者不需要重新輸入密碼
- Refresh Token 的 TTL 長（7 天），前端在 Access Token 過期時自動用 Refresh Token 換新的

### 2.2 CategoryService（分類服務）

- [x] 定義 `category.proto`（ListCategories）
- [x] 各層實作

**為什麼這麼簡單？**
Category 基本上就是一張固定的查詢表。初期只需要 ListCategories，讓前端可以在搜尋頁面顯示分類選項。未來如果需要 CRUD，直接在 AdminService 中加。

### 2.3 TagService（標籤服務）

- [x] 定義 `tag.proto`（ListTags, GetMyTags, SetMyTags）
- [x] 各層實作
- [x] 訪客 session_id 支援
- [x] 每人最多 5 個標籤的限制

**為什麼標籤是推薦系統的核心？**
- 相比 YouTube 那種基於觀看歷史的複雜推薦演算法，標籤推薦簡單、透明、可控
- 使用者主動選擇感興趣的標籤 → 系統根據標籤推薦影片。使用者知道為什麼看到這些影片
- 不需要收集大量行為數據，符合隱私友善的設計理念

**為什麼支援訪客（session_id）？**
- 未註冊的使用者也能選標籤、看推薦。降低註冊門檻，讓新使用者先體驗產品價值
- session_id 存在瀏覽器的 localStorage，不需要帳號就能個人化

**為什麼限制最多 5 個標籤？**
- 如果使用者選了所有標籤，推薦就失去意義（等於隨機推薦）
- 5 個標籤足夠表達偏好，同時保持推薦的多樣性
- 減少 Redis SUNION 的計算量（最多合併 5 個 SET）

### 2.4 VideoService（影片服務）

- [x] 定義 `video.proto`（CRUD + GetRecommended + TogglePublish）
- [x] MinIO 檔案上傳（二階段：上傳檔案 → 取得路徑 → CreateVideo RPC）
- [x] 標籤推薦演算法（隨機子集組合）
- [x] 存取層級檢查（public/subscriber/premium，透過 MembershipChecker 介面）
- [x] 觀看次數統計（會員 vs 非會員，透過 Redis 緩衝）

**推薦演算法為什麼用「隨機子集」？**
```
使用者選了 5 個標籤：[搞笑, 教學, Vlog, 開箱, 科技評測]
每次請求：隨機挑 1~5 個標籤 → 例如這次挑 [搞笑, 開箱]
→ 從 Redis SUNION tag:搞笑 tag:開箱 → 合併影片 ID → 隨機取 20 支
```
- **為什麼不固定用所有 5 個標籤？** 每次都 SUNION 5 個標籤會讓推薦結果太穩定，使用者每次刷新看到差不多的影片。隨機子集讓每次推薦結果都不同，增加探索感
- **為什麼不用權重排序？** 初期資料量小，隨機推薦已經夠好。加權重排序是未來優化的方向

**為什麼分 views_member 和 views_non_member？**
- Dashboard 需要顯示「會員 vs 非會員」的觀看比例
- 創作者可以用這個數據決定哪些影片適合設為會員專屬

**為什麼用 TogglePublish（上架/下架）而不是直接刪除？**
- 創作者可能暫時不想讓某支影片被看到，但不想永久刪除
- 只有下架的影片才能被永久刪除（防止誤刪已上架的影片）

### 2.5 SearchService（搜尋服務）

- [x] MySQL FULLTEXT 搜尋（BOOLEAN MODE）
- [x] 多種篩選條件：分類、時長、日期、觀看數、存取類型
- [x] 分頁

**為什麼用 MySQL FULLTEXT 而不是 Elasticsearch？**
- 初期影片數量少（幾百到幾千支），MySQL FULLTEXT 完全夠用
- 少一個基礎設施（ES 需要額外的 JVM、記憶體、維護成本）
- 當影片數量超過百萬且搜尋效能不足時，再遷移到 ES（Phase 7）

### 2.6 ChannelService（頻道服務 - 免費訂閱）

- [x] 訂閱/退訂流程
- [x] 註冊時自動建立頻道（在 AuthUsecase.Register 中）

**為什麼每個使用者都自動建立頻道？**
- 降低「成為創作者」的門檻：不需要額外申請，只要開始上傳影片就是創作者
- `channels` 分離為獨立的表（不是 users 的欄位），保留未來擴展的彈性（例如頻道名稱、頭像、自我介紹）

**為什麼先只做免費訂閱（Tier 1）？**
- 付費訂閱需要 Paddle 串接，放在 Phase 3
- 免費訂閱先驗證訂閱/退訂的基本流程、通知機制

### 2.7 推薦快取（Redis）

- [x] 啟動時快取預熱（`cache_warmup.go`）
- [x] 快取讀寫層（`video_cache.go`）：SUNION 標籤 SETs → pipeline HGETALL 影片 HASHes
- [x] `videoRepo.ListByTags` 優先讀取快取，miss 時降級到 MySQL
- [x] 應用層快取淘汰 hook（`EvictVideo`：移除標籤 SETs + 刪除影片 HASH + ZREM popular）
- [x] 觀看計數緩衝：`HINCRBY views:buffer` + `ZINCRBY popular:global` → 每 30 秒批次寫入 MySQL
- [x] 背景 worker（`cleanup_worker.go`）：觀看計數 flush + 失敗淘汰重試
- [x] TTL 安全網：標籤/影片 30 分鐘、popular 10 分鐘
- [x] Redis 資料結構：`tag:{id}` SET、`video:{id}` HASH、`popular:global` ZSET、`views:buffer` HASH、`cleanup:queue` LIST

**為什麼用兩層結構（SET 索引 + HASH 資料）？**
```
Layer 1 (索引): tag:1 → SET {101, 102, 103}    ← 某個標籤下有哪些影片
Layer 2 (資料): video:101 → HASH {title, views, ...}  ← 影片的摘要資訊
```
- 如果只用一層（例如把影片資訊直接存在 tag 的 SET 裡），同一支影片如果有 3 個標籤就會存 3 份 → 浪費記憶體且更新時要改 3 個地方
- 兩層設計：索引層只存 ID（輕量），資料層每支影片只存一份（不重複）

**為什麼用 MySQL 主鍵 `video.ID` 當 Redis key？**
- 穩定不變：ID 是資料庫自動生成的，不會因為改標題或描述而變化
- 免除額外查詢：如果用 hash(影片名稱+描述) 當 key，要找影片時還需要先算 hash，然後再查。用 ID 直接就是 `video:{id}`

**為什麼只快取公開影片？**
- 付費影片（`access_tier > 0`）需要檢查使用者的會員身份，這個檢查必須即時查 MySQL
- 如果把付費影片也放快取，就需要在快取層處理權限檢查，增加複雜度且容易出安全漏洞

**為什麼快取淘汰用應用層 hook 而不是 TTL？**
- TTL 太慢：如果 TTL 設 30 分鐘，影片被刪除後最多 30 分鐘內使用者還能看到它
- 應用層 hook：影片被編輯/刪除時立即從 Redis 移除，保證一致性
- TTL 仍然保留作為安全網：如果 hook 失敗（例如 Redis 暫時不可用），TTL 最終會自動過期

**為什麼帳號刪除時要「先收集再刪除」？**
```
正確順序：
1. 從 MySQL 收集使用者的所有影片 ID + 標籤 ID（此時資料還在）
2. 從 Redis 移除快取
3. 從 MySQL 硬刪除

錯誤順序：
1. 從 MySQL 硬刪除（資料已消失！）
2. 想從 Redis 移除 → 但不知道要移除哪些 key（因為查不到了）
```
- 必須在 MySQL 資料還存在的時候先收集需要清理的 key

**為什麼需要 cleanup worker？**
- 帳號刪除時，假設有 50 支影片要從 Redis 移除。如果移除到第 30 支時 Redis 斷線：
  - 不能阻塞使用者的刪除操作（MySQL 已經刪了）
  - 剩下 20 支記錄在 cleanup queue 中
  - 背景 worker 持續重試直到 Redis 恢復
- 這是一個典型的「盡力而為 + 最終一致」模式

**觀看計數為什麼用 Redis 緩衝？**
- 每次觀看都直接寫 MySQL：假設 1000 人同時看同一支影片 → 1000 次 UPDATE → MySQL 扛不住
- 改用 Redis ZINCRBY（記憶體中累加）→ 每 30 秒批次寫入 MySQL → MySQL 每 30 秒只收到一次 UPDATE
- MySQL 仍然是觀看次數的最終資料來源（source of truth）

---

## Phase 3: 前端 MVP（Vue 3 SPA） ✅

> **目標**：一個可運作的網頁應用，使用者可以瀏覽和播放影片，管理員可以管理使用者和影片。這是 **MVP 里程碑** — 第一個完整可用的產品。

### 為什麼把前端移到 Phase 3？

Phase 2 完成後，後端已經可以建立使用者、上傳帶標籤的影片，基礎設施也就緒了。在加入變現或進階功能之前，我們需要一個**可用的產品**。一個讓使用者看影片、讓管理員管理內容的 Vue 3 SPA 就是最簡單的完整產品。其他功能（支付、通知、儀表板）都在這個 MVP 之上迭代。

### 3.0 管理員帳號（透過 `.env`）

- [x] 管理員帳號/密碼寫在 `.env` 檔案中（`ADMIN_USERNAME`、`ADMIN_PASSWORD`）
- [x] 應用程式啟動時自動建立或驗證管理員帳號（冪等）
- [x] 不提供管理員註冊端點 — 管理員只從環境變數建立
- [x] Admin 設定加入 `conf.proto`（`Admin` message 含 `username`、`password`）
- [x] `ensureAdmin()` 在 `data.go` 中 — 啟動時建立管理員帳號 + 頻道，密碼變更時自動更新

**為什麼把管理員寫在 `.env`？**
- 單一維護者：不需要管理員註冊流程
- 更新 `.env` 然後重啟就能換密碼，維護方便
- 簡單又安全 — 不暴露任何端點

### 3.0b AdminService（後端 API）

- [x] 定義 `api/fenzvideo/v1/admin.proto`（7 個 RPC + HTTP 路由）
- [x] 使用者管理：AdminListUsers、AdminDeleteUser
- [x] 影片管理：AdminListVideos、AdminDeleteVideo
- [x] 標籤管理：AdminCreateTag、AdminUpdateTag、AdminDeleteTag
- [x] Admin 守衛中間件（方法名前綴 "Admin" 自動保護）
- [x] Admin 不能刪除自己（ADMIN_SELF_DELETE 錯誤）
- [x] 級聯刪除：刪除使用者時移除會員資格、標籤偏好、觀看記錄、通知、打賞、影片、頻道
- [x] Biz/Data/Service 各層 + Wire 依賴注入整合

**為什麼在這裡做 AdminService？**
Phase 3 的 Vue SPA 需要後端 API 讓管理員刪除使用者和影片。這些端點必須先存在，前端才能呼叫。

### 技術選型理由

| 技術 | 為什麼？ |
|------|---------|
| **Vue 3** | 學習曲線低，生態系豐富，適合中小型專案 |
| **Vite** | 開發時的 HMR 極快，建構速度遠超 Webpack |
| **Element Plus** | 成熟的 Vue 3 UI 框架，自帶表格、表單、對話框等企業級元件 |
| **Tailwind CSS** | 在 Element Plus 之外需要客製化樣式時使用，utility-first 開發速度快 |
| **Video.js** | 最成熟的開源影片播放器，支援各種格式和瀏覽器 |
| **Pinia** | Vue 3 官方推薦的狀態管理，比 Vuex 更簡潔 |

### 3.1 專案初始化

- [x] 建立 Vue 3 + Vite + TypeScript 專案
- [x] 安裝 Element Plus + Tailwind CSS + Video.js
- [x] 設定 Vue Router、Pinia、Axios、Vue I18n
- [x] JWT 攔截器（自動刷新、401 時導向登入頁）
- [x] Vite 代理（`/api` → `localhost:8000`）、Tailwind 設定、路徑別名

### 3.2 核心頁面（使用者端）

- [x] **LoginView** — 登入 + 註冊表單（含驗證）
- [x] **HomeView** — 基於標籤的推薦影片瀑布流（含分頁）
- [x] **VideoView** — HTML5 影片播放器 + 影片資訊
- [x] **SearchResultsView** — 搜尋結果 + 篩選側欄
- [x] **CategoryView** — 按分類瀏覽影片
- [x] **ChannelView** — 頻道資訊 + 訂閱/退訂

### 3.3 管理頁面

- [x] **AdminUsersView** — 使用者列表，含刪除功能
- [x] **AdminTagsView** — 標籤 CRUD（含對話框表單）
- [x] 管理員佈局（側邊導航列）

### 3.4 狀態管理（Pinia Stores）

- [x] authStore（JWT token、使用者資訊、isLoggedIn/isAdmin 計算屬性）
- [x] videoStore（推薦影片、當前影片）
- [x] tagStore（標籤 + 訪客 sessionId via UUID）
- [x] searchStore（搜尋詞、篩選條件、結果）
- [x] categoryStore（分類列表）
- [x] adminStore（管理員使用者/標籤/影片管理）

### 3.5 元件

- [x] AppHeader（導航 + 搜尋列 + 登入/管理連結）
- [x] AppSidebar（分類 + 標籤選擇器）
- [x] VideoCard、VideoGrid、Pagination
- [x] TagSelector（最多 5 個標籤，可點擊標籤晶片）
- [x] ConfirmDialog、LoadingSpinner
- [x] 三種佈局：DefaultLayout、AuthLayout、AdminLayout

### 3.6 測試

- [ ] 單元測試：Vitest + Vue Test Utils
- [ ] 端到端測試：Playwright

### 如何執行

```bash
# 1. 啟動基礎設施
docker-compose up -d

# 2. 啟動後端（port 8000）
cd backend && go run ./cmd/backend/ -conf ./configs/

# 3. 啟動前端（port 5173）
cd frontend && npm run dev

# 4. 開啟 http://localhost:5173
# 管理員帳號：admin / admin123
```

### 驗證

```bash
# MVP 流程
# 1. 管理員登入（帳密來自 .env）
# 2. 管理員管理使用者、影片、標籤
# 3. 使用者依標籤瀏覽推薦影片
# 4. 點擊影片 → 播放器播放
# 5. 使用篩選條件搜尋
# 6. 管理員刪除影片 → 從列表消失
# 7. 管理員刪除使用者 → 級聯清理
```

---

## Phase 4: 變現

> **目標**：付費會員和打賞（抖內）功能可以端到端運作。

### 為什麼在前端 MVP 之後？

金流整合是高風險、高複雜度的功能。MVP 必須先穩定 — 使用者可以瀏覽、觀看影片，管理員可以管理內容。穩定之後再加上變現功能。

### 4.1 Paddle 整合套件

- [ ] Paddle SDK 客戶端
- [ ] Webhook 簽名驗證
- [ ] Sandbox 環境設定

**為什麼用 Paddle 而不是 Stripe？**
- Paddle 是 Merchant of Record（MoR）：Paddle 負責處理稅務、退款、合規。開發者不需要自己處理
- 簡化金流：不需要自己申請商家帳號、處理 PCI 合規
- 缺點是手續費較高，但對小型平台來說省下的開發和合規成本更值得

### 4.2 ChannelService（Tier 2 - 付費會員）

- [ ] UpgradeToPremium → 建立 Paddle 訂閱
- [ ] CancelPremium → 取消 Paddle 訂閱
- [ ] 付費影片存取檢查

**為什麼分兩個 Tier？**
- **Tier 1（免費訂閱）**：訂閱頻道，可以看 `access_tier=1` 的影片 + 收到新影片通知
- **Tier 2（付費訂閱）**：月費制，可以看 `access_tier=2` 的專屬影片
- 這個設計讓創作者可以用免費內容吸引訂閱者，再用付費內容變現

### 4.3 DonationService（打賞服務）

- [ ] CreateDonation → Paddle 一次性交易
- [ ] 查詢已送出/已收到的打賞

### 4.4 Paddle Webhook Handler

- [ ] 無認證（用 Paddle 簽名驗證）
- [ ] 分派事件：`transaction.*` → 打賞，`subscription.*` → 會員

### 4.5 DashboardService（儀表板服務）

- [ ] 我的影片列表、分析數據、會費設定

### 4.6 前端 — 變現頁面

- [ ] **ChannelView** — 頻道資訊 + 影片 + 會員 CTA
- [ ] **DashboardView** — 影片管理 + 分析圖表（ECharts）
- [ ] **VideoUploadForm** — 上傳影片，選擇分類/標籤/存取層級
- [ ] VideoDonateDialog、MembershipDialog 元件
- [ ] Paddle.js 結帳覆蓋層整合

---

## Phase 5: 進階功能

> **目標**：通知、使用者自助服務、可觀測性。

### 為什麼在變現之後？

這些功能讓產品更好用，但 MVP 和金流都不依賴它們。使用者已經可以看影片、付費、打賞。通知、自助服務和監控是品質提升。

### 5.1 NATS 整合

- [ ] NATS pub/sub 客戶端
- [ ] 背景訂閱 goroutine

**為什麼用 NATS 而不是 Redis Pub/Sub？**
- NATS 支援 JetStream（持久化訊息），確保訊息不會因為消費者暫時離線而遺失
- Redis Pub/Sub 是 fire-and-forget，如果消費者不在線，訊息就丟了
- NATS 是輕量級的（單一 binary，幾 MB 記憶體），不像 Kafka 那麼重

### 5.2 NotificationService（通知服務）

- [ ] NATS 訂閱 → 通知扇出
- [ ] 通知列表、未讀數、標記已讀

### 5.3 UserService（使用者自助服務）

- [ ] 修改顯示名稱、密碼
- [ ] 隱藏帳號（cascade）
- [ ] 刪除帳號（cascade）
- [ ] 刪除頻道（保留帳號）

**為什麼分「隱藏」和「刪除」？**
- **隱藏（is_hidden = true）**：資料保留但對外不可見。使用者後悔了可以請 Admin 恢復
- **刪除（hard delete）**：資料永久移除。不可逆

### 5.4 可觀測性

- [ ] OpenTelemetry + Jaeger（分散式追蹤）
- [ ] Prometheus + Grafana（監控儀表板）

### 5.5 前端 — 進階功能

- [ ] 通知鈴鐺元件
- [ ] 使用者個人資料 / 自助服務頁面
- [ ] 分析圖表增強

---

## Phase 6: 部署與營運

> **目標**：生產環境就緒。

### 為什麼要獨立的部署階段？

開發環境和生產環境有很多差異：
- SSL 證書、域名設定
- 環境變數管理（不能把密鑰寫在程式碼裡）
- 資料庫備份策略
- 監控告警規則

這些都需要專門處理，不適合和業務功能混在一起做。

---

## Phase 7: 微服務拆分（按需）

> **觸發條件**：至少滿足 2 個以上：量測到瓶頸、團隊超過 3 人、需要不同技術棧、需要不同可用性等級。

### 為什麼要設觸發條件？

- 微服務不是免費的午餐：每拆一個服務就多了服務間通訊、分散式交易、獨立部署、獨立監控的成本
- 只有當單體架構的某個模組確實成為瓶頸時，拆分才有正面效益
- 如果拆了但沒有瓶頸 → 增加了複雜度但沒有收益 → 淨損失

---

## 技術棧總覽

| 層級 | 技術 | 選型理由 |
|------|------|---------|
| **後端** | Go 1.24+, Kratos v2 | Go 效能好、部署簡單（單一 binary）；Kratos 提供 Clean Architecture + 雙協議 |
| **資料庫** | MySQL 8.0, GORM v2 | MySQL 穩定成熟；GORM 提供 AutoMigrate 和良好的 Go 整合 |
| **快取** | Redis 7 | 記憶體資料庫，sub-ms 延遲。用於推薦快取 + 觀看計數緩衝 |
| **儲存** | MinIO | S3 相容的自建物件儲存，不依賴雲端廠商 |
| **訊息** | NATS 2 | 輕量 pub/sub，支援 JetStream 持久化 |
| **支付** | Paddle (sandbox) | Merchant of Record，簡化金流合規 |
| **前端** | Vue 3, Vite, Element Plus, Video.js | 開發效率高，生態系成熟 |
| **監控** | OpenTelemetry, Jaeger, Prometheus, Grafana | 全鏈路追蹤 + 指標收集 + 視覺化儀表板 |
| **CI/CD** | Gitea Actions / Woodpecker CI | 自建 CI/CD，不依賴 GitHub Actions |
| **容器** | Docker + Docker Compose | 一鍵啟動所有基礎設施 |

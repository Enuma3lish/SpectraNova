# FenzVideo 並發風險稽核

## 範圍

本稽核聚焦於目前後端 Go 程式碼中與並發相關的部分，特別是：

- 啟動 goroutine 的程式碼
- 管理共享可變狀態的邏輯
- WebSocket 連線處理
- NATS callback
- 透過 Redis 做批次寫入的流程

本次檢視的主要檔案包括：

- `backend/internal/data/cleanup_worker.go`
- `backend/internal/data/realtime_hub.go`
- `backend/internal/data/notification.go`
- `backend/internal/data/video_cache.go`
- `backend/internal/server/http.go`
- `backend/internal/biz/video.go`
- `backend/internal/biz/channel.go`

## 工具檢查結果

- `go test -race ./...` 已執行完成，未回報 race。
- 但這 **不代表** 後端目前就一定沒有 race。現有測試覆蓋率仍偏低，尚未真正壓到大多數長生命週期 goroutine、NATS callback、WebSocket 流程與 Redis 批次寫入邏輯。

## 發現

### 1. `views:buffer` 的 flush 不是原子操作，可能遺失計數

位置：

- `backend/internal/data/cleanup_worker.go:79`
- `backend/internal/data/cleanup_worker.go:144`

風險等級：高

風險原因：

- `flushViewBuffer()` 先用 `HGetAll(views:buffer)` 讀取 Redis hash，之後再用 `Del(views:buffer)` 整個刪掉。
- 如果在這兩個動作之間有新的 view increment 寫進來，那些新資料可能會在還沒 flush 到 MySQL 前就被刪掉。
- 這是典型的 shared mutable state 上的「先讀後刪」競態問題。

影響：

- 高併發下可能遺失觀看次數。
- 熱門影片最容易中招。

建議修法：

- 不要再使用 `HGetAll` + `Del` 這種兩步式流程。
- 較好的做法有：
  - 用 Lua script 原子地「讀取並清空」
  - 用 `RENAME views:buffer views:buffer:drain:<ts>` 先做原子移交，再處理 renamed key
  - 改成 Redis Streams 或 append-only event pipeline

### 2. Like 流程在並發下有邏輯競態

位置：

- `backend/internal/biz/video.go:189`
- `backend/internal/server/http.go:83`

風險等級：高

風險原因：

- `LikeVideo()` 目前每次被呼叫都會發 notification。
- 沒有任何 durable dedupe 機制，例如：
  - `(video_id, user_id)` 唯一鍵
  - 原子 Redis set membership 檢查
- 同一個使用者若在短時間重複送 request，可能會重複通知創作者，未來若加上 like count 也會產生重複計數。

影響：

- 創作者收到重複通知
- 日後加入 like 總數時可能會統計錯誤
- client retry storm 會把問題放大

建議修法：

- 讓 like 在資料層具備 idempotency。
- 建議做法：
  - MySQL 建一張 likes 表，對 `(video_id, user_id)` 建唯一索引，只有首次 insert 成功才發 event
  - 或用 Redis `SADD` / `SETNX`，只有在 Redis 回報這是新 like 時才發 event

### 3. 關閉流程沒有等待背景 goroutine 正常結束

位置：

- `backend/internal/data/data.go:73`
- `backend/internal/data/cleanup_worker.go:30`
- `backend/internal/data/notification.go:41`

風險等級：中

風險原因：

- `NewData()` 啟動背景 worker 與 NATS subscriber 之後，cleanup 只做 `bgCancel()`，接著立刻關閉 Redis 與 NATS client。
- 目前沒有 `sync.WaitGroup` 或其他 join 機制去等待 goroutine 真正退出。
- 如果某個 worker 正在執行 Redis、DB 或 NATS 操作，就可能在資源已關閉時還沒完全結束。

影響：

- shutdown 過程可能出現雜訊或不穩定行為
- goroutine 結束只靠 best effort，不夠可預測
- 可能出現半途工作中斷、log spam 或隱性 shutdown race

建議修法：

- 引入 worker group 與 `sync.WaitGroup`
- cleanup 順序應改為：
  1. cancel context
  2. 等待 worker / subscriber 結束
  3. 再關閉 Redis / NATS / 其他共享資源

### 4. NATS callback 忽略 shutdown context，可能在取消後繼續執行

位置：

- `backend/internal/data/notification.go:48`
- `backend/internal/data/notification.go:67`

風險等級：中

風險原因：

- NATS callback 目前呼叫的是 `handleNotificationMessage(context.Background(), ...)`。
- 代表訊息處理完全不受應用程式 shutdown cancel 控制。
- 一旦 DB 或 Redis 卡住，callback 可能會在系統進入 shutdown 後仍持續跑。

影響：

- shutdown 拖長
- goroutine 在系統已經開始 teardown 後仍持續工作

建議修法：

- callback 內應傳入帶 shutdown 意識的 context
- 如果要更完整的 graceful shutdown，可考慮對 subscription / connection 使用 `Drain()`

### 5. 閒置中的 WebSocket handler 可能長時間保留 goroutine

位置：

- `backend/internal/server/http.go:108`
- `backend/internal/server/http.go:134`

風險等級：中

風險原因：

- 每個 WebSocket 連線都會佔用一個長生命週期 request goroutine，阻塞在 `ReadMessage()`。
- 目前有設定 read deadline，也有 pong handler，但 server 並沒有主動送 ping frame。
- 如果連線進入 silent / half-open 狀態，該 goroutine 很可能會一直卡到 read deadline 才釋放。
- 另外也沒有明確的 server shutdown 廣播或 hub-level close。

影響：

- goroutine 數量會隨閒置 WebSocket 連線累積
- deploy 或 shutdown 時收斂較慢

建議修法：

- 加入 server-side ping loop 或 heartbeat ticker
- 在 hub 層提供 shutdown 時主動關閉所有 WebSocket 的機制
- 讓 WebSocket handler 的生命週期明確受 server lifecycle 控制

### 6. 同一條 WebSocket 可能出現 close / write 競爭

位置：

- `backend/internal/data/realtime_hub.go:67`
- `backend/internal/data/realtime_hub.go:112`

風險等級：中

風險原因：

- `SendToUser()` 在寫入時有使用 `realtimeConn.mu`
- 但 `Unregister()` 關閉同一個 `websocket.Conn` 時，沒有使用同一把 per-connection lock
- 這會造成某些情況下，一個 goroutine 正在 write，另一個 goroutine 同時 close

影響：

- 間歇性寫入失敗
- 在高 fan-out 與斷線交錯時，容易出現難以重現的問題

建議修法：

- close 與 write 都走同一把 per-connection lock
- 或改成每條 connection 一個專屬 write pump goroutine，由 channel 負責送訊息

### 7. Notification payload 的 delivery 狀態有 TOCTOU 問題

位置：

- `backend/internal/data/notification.go:73`
- `backend/internal/data/notification.go:96`

風險等級：低

風險原因：

- 程式先用 `IsOnline()` 計算 `websocket_delivered`，之後才呼叫 `SendToUser()`。
- 使用者可能在這兩步之間斷線或重連。
- 所以最後寫進資料庫的 payload 狀態，可能與實際是否有 live delivery 不一致。

影響：

- notification metadata 不準確
- 日後若加入 delivery analytics，會增加除錯成本

建議修法：

- 讓 `SendToUser()` 回傳實際 delivery 結果，再把那個結果寫入 payload

### 8. Subscribe 流程仍有小型 TOCTOU 視窗，但 DB unique index 已保護資料正確性

位置：

- `backend/internal/biz/channel.go:75`
- `backend/internal/data/channel.go:66`
- `backend/internal/data/model/membership.go:6`

風險等級：低

風險原因：

- `ChannelUsecase.Subscribe()` 先查是否已存在 membership，再決定是否建立。
- 兩個並發 request 可能同時通過 existence check。
- 不過資料庫對 `(channel_id, user_id)` 已經有 unique index，所以通常不會真的產生重複資料。
- 真正的風險是其中一個 request 可能會收到原始 duplicate-key 類型的錯誤，而不是乾淨的業務錯誤。

影響：

- 使用者在重複點擊或 retry 時，可能看到不夠友善的錯誤

建議修法：

- 將 duplicate-key insert 轉譯成 `CHANNEL_ALREADY_SUBSCRIBED`
- 儘量使用 single-write semantics，而不是 check-then-create

## 看起來還算合理的地方

- `RealtimeHub.users` 有用 `sync.RWMutex` 保護，單純 map 存取本身目前看起來沒有明顯 race
- `IncrementViewsBuffered()` 使用 Redis `HINCRBY` 來做 atomic counter，方向是對的
- 背景 worker 的 ticker 有正確 `defer ticker.Stop()`

## 建議優先修復順序

1. 先修 `flushViewBuffer()` 的 view count 遺失問題
2. 在 like endpoint 正式承受流量前，加上 idempotent like storage
3. 補齊 worker、subscriber、WebSocket 的可預測 shutdown 機制
4. 將 WebSocket 連線管理改成 write-pump 模式

## 建議後續驗證

- 加入整合測試，模擬 flush 視窗期間大量 view increment
- 加入 concurrent like 測試，模擬同一使用者重複請求
- 加入 WebSocket lifecycle 測試，覆蓋 disconnect、reconnect 與 shutdown
- 在補上高並發測試後，再執行更有意義的 `go test -race`
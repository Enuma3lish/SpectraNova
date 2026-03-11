# FenzVideo Concurrency Risk Audit

## Scope

This audit focuses on the current backend Go code, especially the parts that start goroutines, manage shared mutable state, handle WebSocket connections, process NATS callbacks, and batch writes through Redis.

Files reviewed include:

- `backend/internal/data/cleanup_worker.go`
- `backend/internal/data/realtime_hub.go`
- `backend/internal/data/notification.go`
- `backend/internal/data/video_cache.go`
- `backend/internal/server/http.go`
- `backend/internal/biz/video.go`
- `backend/internal/biz/channel.go`

## Tooling Result

- `go test -race ./...` completed without reported races.
- This does **not** prove the backend is race-safe. Current test coverage is light and does not exercise most long-lived goroutines, NATS callbacks, WebSocket flows, or Redis batching logic.

## Findings

### 1. Non-atomic view-buffer flush can lose increments

Location:

- `backend/internal/data/cleanup_worker.go:79`
- `backend/internal/data/cleanup_worker.go:144`

Risk level: High

Why it is risky:

- `flushViewBuffer()` reads the Redis hash with `HGetAll(views:buffer)` and later deletes the whole hash with `Del(views:buffer)`.
- If new view increments arrive between those two operations, those fresh increments can be deleted before they are ever flushed to MySQL.
- This is a classic read-then-delete race on a shared mutable structure.

Impact:

- Lost view counts under concurrent traffic.
- Hot videos are the most exposed.

Recommended fix:

- Replace `HGetAll` + `Del` with an atomic handoff pattern.
- Good options:
  - Lua script that reads and clears atomically.
  - `RENAME views:buffer views:buffer:drain:<ts>` followed by processing the renamed key.
  - Redis Streams or a dedicated append-only event pipeline.

### 2. Like handling is logically racy under concurrent requests

Location:

- `backend/internal/biz/video.go:189`
- `backend/internal/server/http.go:83`

Risk level: High

Why it is risky:

- `LikeVideo()` publishes a notification every time the endpoint is hit.
- There is no durable deduplication such as a unique `(video_id, user_id)` record or an atomic Redis set membership check.
- Concurrent duplicate requests from the same user can therefore produce duplicated notifications and overcount any future like metrics.

Impact:

- Duplicate creator alerts.
- Incorrect like totals once a counter is added.
- Retry storms from clients can amplify the problem.

Recommended fix:

- Make likes idempotent at the data layer.
- Preferred patterns:
  - MySQL table with unique `(video_id, user_id)` and publish only on successful first insert.
  - Or `SADD` / `SETNX` in Redis, then publish only when the operation reports a new like.

### 3. Shutdown does not wait for background goroutines to exit

Location:

- `backend/internal/data/data.go:73`
- `backend/internal/data/cleanup_worker.go:30`
- `backend/internal/data/notification.go:41`

Risk level: Medium

Why it is risky:

- `NewData()` starts background workers and the NATS subscriber, then cleanup only calls `bgCancel()` and immediately closes Redis and NATS clients.
- There is no `sync.WaitGroup` or equivalent join mechanism to wait for worker exit.
- A worker may still be in the middle of Redis, DB, or NATS work when cleanup closes those resources.

Impact:

- Noisy shutdown behavior.
- Best-effort goroutine termination instead of deterministic lifecycle management.
- Increased chance of partial work, log spam, or hidden shutdown races.

Recommended fix:

- Introduce a managed worker group with `sync.WaitGroup`.
- Cleanup should:
  1. cancel context
  2. wait for workers/subscribers to stop
  3. close Redis/NATS/other shared resources

### 4. NATS callback ignores shutdown context and can outlive cancellation

Location:

- `backend/internal/data/notification.go:48`
- `backend/internal/data/notification.go:67`

Risk level: Medium

Why it is risky:

- The NATS callback calls `handleNotificationMessage(context.Background(), ...)`.
- That means message processing ignores application shutdown cancellation.
- If DB or Redis stalls, the callback can continue past shutdown intent.

Impact:

- Longer shutdown tail.
- Goroutines performing work after the app has started tearing down.

Recommended fix:

- Pass a shutdown-aware context into the callback path.
- Consider `Drain()` on the subscription/connection if graceful NATS shutdown is required.

### 5. Idle WebSocket handlers can retain goroutines for a long time

Location:

- `backend/internal/server/http.go:108`
- `backend/internal/server/http.go:134`

Risk level: Medium

Why it is risky:

- Each WebSocket connection is handled by a long-lived request goroutine blocked in `ReadMessage()`.
- The code sets a read deadline and installs a pong handler, but the server never sends ping frames.
- A silent but half-open connection can therefore hold the goroutine until the read deadline expires.
- There is also no explicit server-shutdown broadcast or hub-wide close.

Impact:

- Goroutine retention proportional to number of idle WebSocket clients.
- Slower cleanup during deploy or shutdown.

Recommended fix:

- Add a server-side ping loop or heartbeat ticker.
- Add hub-level shutdown that closes all active WebSocket connections.
- Make WebSocket handler termination explicitly tied to server lifecycle.

### 6. Possible concurrent close/write on the same WebSocket connection

Location:

- `backend/internal/data/realtime_hub.go:67`
- `backend/internal/data/realtime_hub.go:112`

Risk level: Medium

Why it is risky:

- `SendToUser()` writes under `realtimeConn.mu`.
- `Unregister()` closes the same `websocket.Conn` without taking that per-connection mutex.
- This creates a potential overlap where one goroutine writes while another closes.

Impact:

- Intermittent send failures.
- Hard-to-reproduce race-like behavior during disconnects and fan-out under load.

Recommended fix:

- Serialize close and write operations through the same per-connection lock.
- Or move each connection to a dedicated write pump goroutine and send via buffered channel.

### 7. Notification payload records delivery state using a TOCTOU check

Location:

- `backend/internal/data/notification.go:73`
- `backend/internal/data/notification.go:96`

Risk level: Low

Why it is risky:

- The code computes `websocket_delivered` from `IsOnline()` before calling `SendToUser()`.
- The user can disconnect or reconnect between those two steps.
- The stored payload can therefore disagree with actual live delivery.

Impact:

- Inaccurate notification metadata.
- Confusing debugging if delivery analytics are added later.

Recommended fix:

- Let `SendToUser()` return the actual delivery result and persist that result instead of checking online state separately.

### 8. Subscribe flow still has a small TOCTOU window, though the DB index protects correctness

Location:

- `backend/internal/biz/channel.go:75`
- `backend/internal/data/channel.go:66`
- `backend/internal/data/model/membership.go:6`

Risk level: Low

Why it is risky:

- `ChannelUsecase.Subscribe()` checks existing membership before creating one.
- Two concurrent requests can both pass the existence check.
- The unique DB index on `(channel_id, user_id)` prevents duplicate rows, so data corruption is unlikely.
- However, one request can still fail with a raw duplicate-key style database error unless it is translated cleanly.

Impact:

- User-visible transient errors under concurrent subscribe clicks/retries.

Recommended fix:

- Treat duplicate-key insert as a clean `CHANNEL_ALREADY_SUBSCRIBED` outcome.
- Prefer single-write semantics over check-then-create where possible.

## Areas That Look Reasonable

- `RealtimeHub.users` is protected by `sync.RWMutex`, so ordinary map access itself is not obviously racy.
- Redis `HINCRBY` usage in `IncrementViewsBuffered()` is a good choice for atomic counter updates.
- Background worker tickers are stopped correctly with `defer ticker.Stop()`.

## Priority Fix Order

1. Fix the view-buffer flush race in `flushViewBuffer()`.
2. Add idempotent like storage before the like endpoint is used under load.
3. Add deterministic goroutine lifecycle management for workers, subscriber shutdown, and WebSocket cleanup.
4. Refactor WebSocket connection management toward a write-pump model.

## Suggested Follow-up Validation

- Add integration tests that hammer view increments during flush windows.
- Add concurrent like tests with repeated requests from the same user.
- Add WebSocket lifecycle tests covering disconnect, reconnect, and shutdown.
- Run targeted `go test -race` after adding those higher-concurrency tests.
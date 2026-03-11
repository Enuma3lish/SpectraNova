package data

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/redis/go-redis/v9"
)

const presenceTTL = 10 * time.Minute

type realtimeConn struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

type RealtimeHub struct {
	redis *redis.Client
	log   *log.Helper

	mu    sync.RWMutex
	users map[uint64]map[string]*realtimeConn
}

func NewRealtimeHub(rdb *redis.Client, logger log.Logger) *RealtimeHub {
	return &RealtimeHub{
		redis: rdb,
		log:   log.NewHelper(logger),
		users: make(map[uint64]map[string]*realtimeConn),
	}
}

func (h *RealtimeHub) Register(ctx context.Context, userID uint64, conn *websocket.Conn) string {
	connectionID := uuid.NewString()

	h.mu.Lock()
	if h.users[userID] == nil {
		h.users[userID] = make(map[string]*realtimeConn)
	}
	h.users[userID][connectionID] = &realtimeConn{conn: conn}
	h.mu.Unlock()

	h.refreshPresence(ctx, userID, connectionID)
	return connectionID
}

func (h *RealtimeHub) RefreshPresence(ctx context.Context, userID uint64, connectionID string) {
	h.refreshPresence(ctx, userID, connectionID)
}

func (h *RealtimeHub) refreshPresence(ctx context.Context, userID uint64, connectionID string) {
	if h.redis == nil {
		return
	}
	presenceKey := fmt.Sprintf("presence:user:%d", userID)
	connectionsKey := fmt.Sprintf("presence:user:%d:connections", userID)
	h.redis.Set(ctx, presenceKey, "online", presenceTTL)
	h.redis.SAdd(ctx, connectionsKey, connectionID)
	h.redis.Expire(ctx, connectionsKey, presenceTTL)
}

func (h *RealtimeHub) Unregister(ctx context.Context, userID uint64, connectionID string) {
	var conn *realtimeConn
	var remaining int

	h.mu.Lock()
	if userConnections, ok := h.users[userID]; ok {
		conn = userConnections[connectionID]
		delete(userConnections, connectionID)
		remaining = len(userConnections)
		if remaining == 0 {
			delete(h.users, userID)
		}
	}
	h.mu.Unlock()

	if conn != nil {
		_ = conn.conn.Close()
	}

	if h.redis != nil {
		connectionsKey := fmt.Sprintf("presence:user:%d:connections", userID)
		presenceKey := fmt.Sprintf("presence:user:%d", userID)
		h.redis.SRem(ctx, connectionsKey, connectionID)
		if remaining == 0 {
			h.redis.Del(ctx, connectionsKey, presenceKey)
		}
	}
}

func (h *RealtimeHub) IsOnline(ctx context.Context, userID uint64) bool {
	h.mu.RLock()
	if len(h.users[userID]) > 0 {
		h.mu.RUnlock()
		return true
	}
	h.mu.RUnlock()

	if h.redis == nil {
		return false
	}
	presenceKey := fmt.Sprintf("presence:user:%d", userID)
	exists, err := h.redis.Exists(ctx, presenceKey).Result()
	return err == nil && exists > 0
}

func (h *RealtimeHub) SendToUser(userID uint64, payload any) bool {
	h.mu.RLock()
	userConnections := h.users[userID]
	connections := make(map[string]*realtimeConn, len(userConnections))
	for connectionID, conn := range userConnections {
		connections[connectionID] = conn
	}
	h.mu.RUnlock()

	if len(connections) == 0 {
		return false
	}

	delivered := false
	for connectionID, conn := range connections {
		conn.mu.Lock()
		_ = conn.conn.SetWriteDeadline(time.Now().Add(5 * time.Second))
		err := conn.conn.WriteJSON(payload)
		conn.mu.Unlock()
		if err != nil {
			h.log.Warnf("failed to push realtime message to user %d: %v", userID, err)
			h.Unregister(context.Background(), userID, connectionID)
			continue
		}
		delivered = true
	}

	return delivered
}

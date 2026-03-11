package data

import (
	"context"
	"encoding/json"
	"fmt"

	"backend/internal/biz"
	"backend/internal/data/model"

	"github.com/go-kratos/kratos/v2/log"
	"github.com/nats-io/nats.go"
	"gorm.io/datatypes"
)

const notificationSubjectPrefix = "notification."

type notificationPublisher struct {
	conn *nats.Conn
	log  *log.Helper
}

func NewNotificationPublisher(nc *nats.Conn, logger log.Logger) biz.NotificationPublisher {
	return &notificationPublisher{
		conn: nc,
		log:  log.NewHelper(logger),
	}
}

func (p *notificationPublisher) Publish(ctx context.Context, event *biz.NotificationEvent) error {
	if p.conn == nil {
		return fmt.Errorf("nats unavailable")
	}
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}
	return p.conn.Publish(notificationSubjectPrefix+event.Type, data)
}

func StartNotificationSubscriber(ctx context.Context, d *Data, logger log.Logger) {
	l := log.NewHelper(logger)
	if d.NATS == nil {
		l.Warn("notification subscriber skipped: NATS not available")
		return
	}

	sub, err := d.NATS.Subscribe(notificationSubjectPrefix+">", func(msg *nats.Msg) {
		if err := handleNotificationMessage(context.Background(), d, msg.Data); err != nil {
			l.Warnf("failed to handle notification message: %v", err)
		}
	})
	if err != nil {
		l.Warnf("failed to subscribe notification subject: %v", err)
		return
	}

	go func() {
		<-ctx.Done()
		_ = sub.Unsubscribe()
		l.Info("notification subscriber stopped")
	}()

	l.Info("notification subscriber started")
}

func handleNotificationMessage(ctx context.Context, d *Data, raw []byte) error {
	var event biz.NotificationEvent
	if err := json.Unmarshal(raw, &event); err != nil {
		return err
	}

	payload, err := json.Marshal(map[string]any{
		"video_id":            event.VideoID,
		"actor_id":            event.ActorID,
		"reason":              event.Reason,
		"websocket_delivered": d.Realtime != nil && d.Realtime.IsOnline(ctx, event.RecipientID),
	})
	if err != nil {
		return err
	}

	var message *string
	if event.Message != "" {
		message = &event.Message
	}

	notification := &model.Notification{
		UserID:  event.RecipientID,
		Type:    event.Type,
		Title:   event.Title,
		Message: message,
		Payload: datatypes.JSON(payload),
	}
	if err := d.DB.WithContext(ctx).Create(notification).Error; err != nil {
		return err
	}

	if d.Realtime != nil {
		d.Realtime.SendToUser(event.RecipientID, event)
	}

	return nil
}

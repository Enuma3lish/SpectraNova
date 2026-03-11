package biz

import (
	"context"
	"errors"
	stdlog "log"
	"testing"
	"time"

	kratoslog "github.com/go-kratos/kratos/v2/log"
)

type stubAdminRepo struct {
	deletedVideoID uint64
	deleteErr      error
}

func (s *stubAdminRepo) ListUsers(context.Context, int, int) ([]*AdminUser, int64, error) {
	return nil, 0, nil
}
func (s *stubAdminRepo) FindUserByID(context.Context, uint64) (*AdminUser, error) { return nil, nil }
func (s *stubAdminRepo) DeleteUser(context.Context, uint64) error                 { return nil }
func (s *stubAdminRepo) ListAllVideos(context.Context, int, int) ([]*AdminVideo, int64, error) {
	return nil, 0, nil
}
func (s *stubAdminRepo) DeleteVideo(_ context.Context, id uint64) error {
	s.deletedVideoID = id
	return s.deleteErr
}
func (s *stubAdminRepo) CreateTag(context.Context, *AdminTag) (*AdminTag, error)  { return nil, nil }
func (s *stubAdminRepo) UpdateTag(context.Context, *AdminTag) (*AdminTag, error)  { return nil, nil }
func (s *stubAdminRepo) DeleteTag(context.Context, uint64) error                  { return nil }
func (s *stubAdminRepo) FindTagByID(context.Context, uint64) (*AdminTag, error)   { return nil, nil }
func (s *stubAdminRepo) FindTagByName(context.Context, string) (*AdminTag, error) { return nil, nil }

func adminTestLogger() kratoslog.Logger {
	return kratoslog.NewStdLogger(stdlog.Writer())
}

func TestAdminUsecaseDeleteVideoPublishesModerationEvent(t *testing.T) {
	videoRepo := &stubVideoRepo{videos: map[uint64]*Video{
		9: {ID: 9, UserID: 22, Title: "illegal", IsPublished: true, CreatedAt: time.Now()},
	}}
	adminRepo := &stubAdminRepo{}
	publisher := &capturePublisher{}
	notify := NewNotificationUsecase(videoRepo, publisher, adminTestLogger())
	uc := NewAdminUsecase(adminRepo, videoRepo, notify, adminTestLogger())

	if err := uc.DeleteVideo(context.Background(), 1, 9); err != nil {
		t.Fatalf("DeleteVideo returned error: %v", err)
	}
	if adminRepo.deletedVideoID != 9 {
		t.Fatalf("expected repo delete for video 9, got %d", adminRepo.deletedVideoID)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(publisher.events))
	}
	if publisher.events[0].Type != NotificationTypeModerationRemoved {
		t.Fatalf("expected %s, got %s", NotificationTypeModerationRemoved, publisher.events[0].Type)
	}
	if publisher.events[0].RecipientID != 22 {
		t.Fatalf("expected recipient 22, got %d", publisher.events[0].RecipientID)
	}
}

func TestAdminUsecaseDeleteVideoNotFound(t *testing.T) {
	videoRepo := &stubVideoRepo{videos: map[uint64]*Video{}}
	adminRepo := &stubAdminRepo{deleteErr: errors.New("should not be used")}
	publisher := &capturePublisher{}
	notify := NewNotificationUsecase(videoRepo, publisher, adminTestLogger())
	uc := NewAdminUsecase(adminRepo, videoRepo, notify, adminTestLogger())

	if err := uc.DeleteVideo(context.Background(), 1, 404); err == nil {
		t.Fatal("expected not found error")
	}
	if len(publisher.events) != 0 {
		t.Fatalf("expected 0 published events, got %d", len(publisher.events))
	}
}

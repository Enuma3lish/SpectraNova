package biz

import (
	"context"
	"errors"
	stdlog "log"
	"testing"
	"time"

	kratoslog "github.com/go-kratos/kratos/v2/log"
)

type stubVideoRepo struct {
	videos map[uint64]*Video
}

func (s *stubVideoRepo) Create(context.Context, *Video) (*Video, error) { return nil, nil }
func (s *stubVideoRepo) Update(context.Context, *Video) (*Video, error) { return nil, nil }
func (s *stubVideoRepo) Delete(context.Context, uint64) error           { return nil }
func (s *stubVideoRepo) ListByTags(context.Context, []uint64, int, int) ([]*Video, int64, error) {
	return nil, 0, nil
}
func (s *stubVideoRepo) ListRandom(context.Context, int, int) ([]*Video, int64, error) {
	return nil, 0, nil
}
func (s *stubVideoRepo) IncrementViews(context.Context, uint64, bool) error         { return nil }
func (s *stubVideoRepo) TogglePublish(context.Context, uint64, bool) error          { return nil }
func (s *stubVideoRepo) GetTagIDsByVideo(context.Context, uint64) ([]uint64, error) { return nil, nil }
func (s *stubVideoRepo) SetVideoTags(context.Context, uint64, []uint64) error       { return nil }
func (s *stubVideoRepo) FindByID(_ context.Context, id uint64) (*Video, error) {
	v, ok := s.videos[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

type stubMembershipChecker struct {
	tier int8
	err  error
}

func (s *stubMembershipChecker) HasMembership(context.Context, uint64, uint64) (int8, error) {
	return s.tier, s.err
}

type capturePublisher struct {
	events []*NotificationEvent
	err    error
}

func (c *capturePublisher) Publish(_ context.Context, event *NotificationEvent) error {
	if c.err != nil {
		return c.err
	}
	c.events = append(c.events, event)
	return nil
}

func testLogger() kratoslog.Logger {
	return kratoslog.NewStdLogger(stdlog.Writer())
}

func TestVideoUsecaseLikeVideoPublishesCreatorNotification(t *testing.T) {
	repo := &stubVideoRepo{videos: map[uint64]*Video{
		10: {ID: 10, UserID: 7, Title: "hello", IsPublished: true, CreatedAt: time.Now()},
	}}
	publisher := &capturePublisher{}
	notify := NewNotificationUsecase(repo, publisher, testLogger())
	uc := NewVideoUsecase(repo, nil, &stubMembershipChecker{}, notify, testLogger())

	if err := uc.LikeVideo(context.Background(), 99, 10); err != nil {
		t.Fatalf("LikeVideo returned error: %v", err)
	}
	if len(publisher.events) != 1 {
		t.Fatalf("expected 1 published event, got %d", len(publisher.events))
	}
	if publisher.events[0].RecipientID != 7 {
		t.Fatalf("expected recipient 7, got %d", publisher.events[0].RecipientID)
	}
	if publisher.events[0].Type != NotificationTypeVideoLiked {
		t.Fatalf("expected %s, got %s", NotificationTypeVideoLiked, publisher.events[0].Type)
	}
}

func TestVideoUsecaseLikeVideoRequiresMembershipForPremium(t *testing.T) {
	repo := &stubVideoRepo{videos: map[uint64]*Video{
		20: {ID: 20, UserID: 7, Title: "premium", IsPublished: true, AccessTier: 2, CreatedAt: time.Now()},
	}}
	publisher := &capturePublisher{}
	notify := NewNotificationUsecase(repo, publisher, testLogger())
	uc := NewVideoUsecase(repo, nil, &stubMembershipChecker{tier: 1}, notify, testLogger())

	if err := uc.LikeVideo(context.Background(), 99, 20); err == nil {
		t.Fatal("expected access error for insufficient membership tier")
	}
	if len(publisher.events) != 0 {
		t.Fatalf("expected 0 published events, got %d", len(publisher.events))
	}
}

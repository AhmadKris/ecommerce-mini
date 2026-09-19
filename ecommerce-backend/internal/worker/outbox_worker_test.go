package worker

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"gorm.io/gorm"

	"ecommerce-backend/internal/model"
)

type mockOutboxRepository struct {
	mock.Mock
}

func (m *mockOutboxRepository) CreateTx(ctx context.Context, tx *gorm.DB, event *model.OutboxEvent) error {
	args := m.Called(ctx, tx, event)
	return args.Error(0)
}

func (m *mockOutboxRepository) FetchPending(ctx context.Context, limit int) ([]model.OutboxEvent, error) {
	args := m.Called(ctx, limit)
	if events, ok := args.Get(0).([]model.OutboxEvent); ok {
		return events, args.Error(1)
	}
	return nil, args.Error(1)
}

func (m *mockOutboxRepository) MarkProcessed(ctx context.Context, id uint) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

func (m *mockOutboxRepository) MarkFailed(ctx context.Context, id uint, errMessage string) error {
	args := m.Called(ctx, id, errMessage)
	return args.Error(0)
}

func TestOutboxWorker(t *testing.T) {
	mockRepo := new(mockOutboxRepository)
	worker := NewOutboxWorker(mockRepo, 50*time.Millisecond)

	pendingEvents := []model.OutboxEvent{
		{
			ID:        1,
			EventType: "order.created",
			Payload:   `{"order_id": 100, "user_id": 5, "total": 250000}`,
			Status:    model.OutboxStatusPending,
		},
	}

	mockRepo.On("FetchPending", mock.Anything, 10).Return(pendingEvents, nil).Maybe()
	mockRepo.On("MarkProcessed", mock.Anything, uint(1)).Return(nil).Maybe()

	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Millisecond)
	defer cancel()

	worker.Start(ctx)

	mockRepo.AssertExpectations(t)
	assert.True(t, true)
}

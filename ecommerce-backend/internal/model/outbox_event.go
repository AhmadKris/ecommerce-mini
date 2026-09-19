package model

import (
	"time"
)

type OutboxStatus string

const (
	OutboxStatusPending   OutboxStatus = "PENDING"
	OutboxStatusProcessed OutboxStatus = "PROCESSED"
	OutboxStatusFailed    OutboxStatus = "FAILED"
)

// OutboxEvent represents an event stored atomically in DB for asynchronous processing.
type OutboxEvent struct {
	ID          uint         `gorm:"primaryKey" json:"id"`
	EventType   string       `gorm:"not null" json:"event_type"`
	Payload     string       `gorm:"type:jsonb;not null" json:"payload"`
	Status      OutboxStatus `gorm:"default:'PENDING'" json:"status"`
	RetryCount  int          `gorm:"default:0" json:"retry_count"`
	LastError   *string      `json:"last_error,omitempty"`
	CreatedAt   time.Time    `gorm:"autoCreateTime" json:"created_at"`
	ProcessedAt *time.Time   `json:"processed_at,omitempty"`
}

func (OutboxEvent) TableName() string {
	return "outbox_events"
}

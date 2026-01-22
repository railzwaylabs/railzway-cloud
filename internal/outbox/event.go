package outbox

import "time"

type EventType string

type EventStatus string

const (
	EventTypeDeployInstance EventType = "deploy_instance"
)

const (
	StatusPending    EventStatus = "pending"
	StatusProcessing EventStatus = "processing"
	StatusCompleted  EventStatus = "completed"
	StatusFailed     EventStatus = "failed"
)

// Event represents a durable outbox entry for control-plane actions.
type Event struct {
	ID            int64       `gorm:"primaryKey"`
	EventType     EventType   `gorm:"type:varchar(100);not null"`
	OrgID         int64       `gorm:"not null"`
	InstanceID    int64       `gorm:"not null"`
	Status        EventStatus `gorm:"type:varchar(50);not null"`
	Attempts      int         `gorm:"not null;default:0"`
	LastError     string      `gorm:"type:text"`
	LockedAt      *time.Time
	NextAttemptAt *time.Time
	ProcessedAt   *time.Time
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

func (Event) TableName() string {
	return "outbox_events"
}

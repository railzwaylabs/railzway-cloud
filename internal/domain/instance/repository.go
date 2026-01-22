package instance

import "context"

// Repository defines the interface for persisting Instance entities.
type Repository interface {
	// FindByOrgID retrieves an instance by its Organization ID.
	FindByOrgID(ctx context.Context, orgID int64) (*Instance, error)

	// Save persists an instance (create or update).
	Save(ctx context.Context, instance *Instance) error

	// UpdateStatus updates only the status of an instance.
	UpdateStatus(ctx context.Context, orgID int64, status InstanceStatus) error

	// ListByStatus retrieves instances matching any of the provided statuses.
	ListByStatus(ctx context.Context, statuses []InstanceStatus, limit int) ([]*Instance, error)
}

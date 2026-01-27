package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
	"gorm.io/gorm"
)

// InstanceModel is the database DTO with Gorm tags.
type InstanceModel struct {
	ID                 int64      `gorm:"column:id;primaryKey"`
	OrgID              int64      `gorm:"column:org_id;uniqueIndex"`
	NomadJobID         string     `gorm:"column:nomad_job_id;type:varchar(255)"`
	DesiredVersion     string     `gorm:"column:desired_version;type:varchar(50)"`
	CurrentVersion     string     `gorm:"column:current_version;type:varchar(50)"`
	Status             string     `gorm:"column:status;type:varchar(50)"`
	Role               string     `gorm:"column:role;type:varchar(50)"`
	LifecycleState     string     `gorm:"column:lifecycle_state;type:varchar(50)"`
	ReadinessStatus    string     `gorm:"column:readiness_status;type:varchar(50)"`
	ReadinessCheckedAt *time.Time `gorm:"column:readiness_checked_at;type:timestamptz"`
	ReadinessError     string     `gorm:"column:readiness_error;type:text"`
	Tier               string     `gorm:"column:tier;type:varchar(50)"`
	ComputeEngine      string     `gorm:"column:compute_engine;type:varchar(50)"`
	PlanID             string     `gorm:"column:plan_id;type:varchar(255)"`
	PriceID            string     `gorm:"column:price_id;type:varchar(255)"`
	SubscriptionID     string     `gorm:"column:subscription_id;type:varchar(255)"`
	LaunchURL          string     `gorm:"column:launch_url;type:text"`
	LastError          string     `gorm:"column:last_error;type:text"`
	OAuthClientID      string     `gorm:"column:oauth_client_id;type:varchar(255)"`
	OAuthClientSecret  string     `gorm:"column:oauth_client_secret;type:varchar(255)"`

	// Database Details
	DBHost     string `gorm:"column:db_host;type:varchar(255)"`
	DBPort     int    `gorm:"column:db_port;type:int"`
	DBName     string `gorm:"column:db_name;type:varchar(255)"`
	DBUser     string `gorm:"column:db_user;type:varchar(255)"`
	DBPassword string `gorm:"column:db_password;type:varchar(255)"` // Should be encrypted in real app

	CreatedAt time.Time `gorm:"column:created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at"`
}

func (InstanceModel) TableName() string {
	return "instances"
}

type Repository struct {
	db *gorm.DB
}

func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) FindByOrgID(ctx context.Context, orgID int64) (*instance.Instance, error) {
	var model InstanceModel
	if err := r.db.WithContext(ctx).Where("org_id = ?", orgID).First(&model).Error; err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return toDomain(model), nil
}

func (r *Repository) Save(ctx context.Context, entity *instance.Instance) error {
	model := toModel(entity)
	if err := r.db.WithContext(ctx).Save(&model).Error; err != nil {
		return err
	}
	// Propagate ID back to entity if new
	entity.ID = model.ID
	return nil
}

func (r *Repository) UpdateStatus(ctx context.Context, orgID int64, status instance.InstanceStatus) error {
	return r.db.WithContext(ctx).Model(&InstanceModel{}).
		Where("org_id = ?", orgID).
		Updates(map[string]any{
			"status":     string(status),
			"updated_at": time.Now().UTC(),
		}).Error
}

func (r *Repository) ListByStatus(ctx context.Context, statuses []instance.InstanceStatus, limit int) ([]*instance.Instance, error) {
	if len(statuses) == 0 {
		return nil, nil
	}
	values := make([]string, 0, len(statuses))
	for _, status := range statuses {
		values = append(values, string(status))
	}

	query := r.db.WithContext(ctx).Where("status IN ?", values).Order("updated_at asc")
	if limit > 0 {
		query = query.Limit(limit)
	}

	var models []InstanceModel
	if err := query.Find(&models).Error; err != nil {
		return nil, err
	}

	items := make([]*instance.Instance, 0, len(models))
	for _, model := range models {
		items = append(items, toDomain(model))
	}
	return items, nil
}

// Mappers

func toDomain(m InstanceModel) *instance.Instance {
	role := instance.InstanceRole(m.Role)
	if role == "" {
		role = instance.RolePrimary
	}
	lifecycle := instance.LifecycleState(m.LifecycleState)
	if lifecycle == "" {
		lifecycle = instance.LifecycleReady
	}
	readiness := instance.ReadinessStatus(m.ReadinessStatus)
	if readiness == "" {
		readiness = instance.ReadinessUnknown
	}
	return &instance.Instance{
		ID:                 m.ID,
		OrgID:              m.OrgID,
		NomadJobID:         m.NomadJobID,
		DesiredVersion:     m.DesiredVersion,
		CurrentVersion:     m.CurrentVersion,
		Status:             instance.InstanceStatus(m.Status),
		Role:               role,
		LifecycleState:     lifecycle,
		Readiness:          readiness,
		ReadinessCheckedAt: m.ReadinessCheckedAt,
		ReadinessError:     m.ReadinessError,
		Tier:               instance.Tier(m.Tier),
		ComputeEngine:      instance.ComputeEngine(m.ComputeEngine),
		PlanID:             m.PlanID,
		PriceID:            m.PriceID,
		SubscriptionID:     m.SubscriptionID,
		LaunchURL:          m.LaunchURL,
		LastError:          m.LastError,
		OAuthClientID:      m.OAuthClientID,
		OAuthClientSecret:  m.OAuthClientSecret,
		DBHost:             m.DBHost,
		DBPort:             m.DBPort,
		DBName:             m.DBName,
		DBUser:             m.DBUser,
		DBPassword:         m.DBPassword,
		CreatedAt:          m.CreatedAt,
		UpdatedAt:          m.UpdatedAt,
	}
}

func toModel(d *instance.Instance) InstanceModel {
	role := d.Role
	if role == "" {
		role = instance.RolePrimary
	}
	lifecycle := d.LifecycleState
	if lifecycle == "" {
		lifecycle = instance.LifecycleReady
	}
	readiness := d.Readiness
	if readiness == "" {
		readiness = instance.ReadinessUnknown
	}
	return InstanceModel{
		ID:                 d.ID,
		OrgID:              d.OrgID,
		NomadJobID:         d.NomadJobID,
		DesiredVersion:     d.DesiredVersion,
		CurrentVersion:     d.CurrentVersion,
		Status:             string(d.Status),
		Role:               string(role),
		LifecycleState:     string(lifecycle),
		ReadinessStatus:    string(readiness),
		ReadinessCheckedAt: d.ReadinessCheckedAt,
		ReadinessError:     d.ReadinessError,
		Tier:               string(d.Tier),
		ComputeEngine:      string(d.ComputeEngine),
		PlanID:             d.PlanID,
		PriceID:            d.PriceID,
		SubscriptionID:     d.SubscriptionID,
		LaunchURL:          d.LaunchURL,
		LastError:          d.LastError,
		OAuthClientID:      d.OAuthClientID,
		OAuthClientSecret:  d.OAuthClientSecret,
		DBHost:             d.DBHost,
		DBPort:             d.DBPort,
		DBName:             d.DBName,
		DBUser:             d.DBUser,
		DBPassword:         d.DBPassword,
		CreatedAt:          d.CreatedAt,
		UpdatedAt:          d.UpdatedAt,
	}
}

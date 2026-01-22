package postgres

import (
	"context"
	"errors"
	"time"

	"github.com/smallbiznis/railzway-cloud/internal/domain/instance"
	"gorm.io/gorm"
)

// InstanceModel is the database DTO with Gorm tags.
type InstanceModel struct {
	ID                int64  `gorm:"primaryKey"`
	OrgID             int64  `gorm:"uniqueIndex"`
	NomadJobID        string `gorm:"type:varchar(255)"`
	DesiredVersion    string `gorm:"type:varchar(50)"`
	CurrentVersion    string `gorm:"type:varchar(50)"`
	Status            string `gorm:"type:varchar(50)"`
	Tier              string `gorm:"type:varchar(50)"`
	ComputeEngine     string `gorm:"type:varchar(50)"`
	PlanID            string `gorm:"type:varchar(255)"`
	PriceID           string `gorm:"type:varchar(255)"`
	SubscriptionID    string `gorm:"type:varchar(255)"`
	LaunchURL         string `gorm:"type:text"`
	LastError         string `gorm:"type:text"`
	OAuthClientID     string `gorm:"type:varchar(255)"`
	OAuthClientSecret string `gorm:"type:varchar(255)"`

	// Database Details
	DBHost     string `gorm:"type:varchar(255)"`
	DBPort     int    `gorm:"type:int"`
	DBName     string `gorm:"type:varchar(255)"`
	DBUser     string `gorm:"type:varchar(255)"`
	DBPassword string `gorm:"type:varchar(255)"` // Should be encrypted in real app

	CreatedAt time.Time
	UpdatedAt time.Time
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
	return &instance.Instance{
		ID:                m.ID,
		OrgID:             m.OrgID,
		NomadJobID:        m.NomadJobID,
		DesiredVersion:    m.DesiredVersion,
		CurrentVersion:    m.CurrentVersion,
		Status:            instance.InstanceStatus(m.Status),
		Tier:              instance.Tier(m.Tier),
		ComputeEngine:     instance.ComputeEngine(m.ComputeEngine),
		PlanID:            m.PlanID,
		PriceID:           m.PriceID,
		SubscriptionID:    m.SubscriptionID,
		LaunchURL:         m.LaunchURL,
		LastError:         m.LastError,
		OAuthClientID:     m.OAuthClientID,
		OAuthClientSecret: m.OAuthClientSecret,
		DBHost:            m.DBHost,
		DBPort:            m.DBPort,
		DBName:            m.DBName,
		DBUser:            m.DBUser,
		DBPassword:        m.DBPassword,
		CreatedAt:         m.CreatedAt,
		UpdatedAt:         m.UpdatedAt,
	}
}

func toModel(d *instance.Instance) InstanceModel {
	return InstanceModel{
		ID:                d.ID,
		OrgID:             d.OrgID,
		NomadJobID:        d.NomadJobID,
		DesiredVersion:    d.DesiredVersion,
		CurrentVersion:    d.CurrentVersion,
		Status:            string(d.Status),
		Tier:              string(d.Tier),
		ComputeEngine:     string(d.ComputeEngine),
		PlanID:            d.PlanID,
		PriceID:           d.PriceID,
		SubscriptionID:    d.SubscriptionID,
		LaunchURL:         d.LaunchURL,
		LastError:         d.LastError,
		OAuthClientID:     d.OAuthClientID,
		OAuthClientSecret: d.OAuthClientSecret,
		DBHost:            d.DBHost,
		DBPort:            d.DBPort,
		DBName:            d.DBName,
		DBUser:            d.DBUser,
		DBPassword:        d.DBPassword,
		CreatedAt:         d.CreatedAt,
		UpdatedAt:         d.UpdatedAt,
	}
}

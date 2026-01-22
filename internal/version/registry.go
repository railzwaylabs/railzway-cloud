package version

import (
	"context"
	"fmt"
	"time"

	"gorm.io/gorm"
)

// ApplicationVersion represents a software version in the registry.
type ApplicationVersion struct {
	ApplicationName string    `gorm:"primaryKey;type:varchar(100)"`
	Version         string    `gorm:"primaryKey;type:varchar(50)"`
	Status          string    `gorm:"type:varchar(20);not null"`
	ReleaseDate     time.Time `gorm:"not null"`
	IsDefault       bool      `gorm:"default:false"`
	MinTier         *string   `gorm:"type:varchar(50)"`
	DockerImage     string    `gorm:"type:text;not null"`
	ChangelogURL    *string   `gorm:"type:text"`
	ReleaseNotes    *string   `gorm:"type:text"`
	BreakingChanges bool      `gorm:"default:false"`
	CreatedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP"`
	UpdatedAt       time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

// TableName sets the table name for GORM.
func (ApplicationVersion) TableName() string {
	return "application_versions"
}

// VersionStatus constants
const (
	StatusStable     = "stable"
	StatusBeta       = "beta"
	StatusRC         = "rc"
	StatusDeprecated = "deprecated"
	StatusEOL        = "eol"
)

// Registry manages application versions.
type Registry struct {
	db *gorm.DB
}

// NewRegistry creates a new version registry.
func NewRegistry(db *gorm.DB) *Registry {
	return &Registry{db: db}
}

// GetDefaultVersion returns the default version for a specific application.
func (r *Registry) GetDefaultVersion(ctx context.Context, appName string) (*ApplicationVersion, error) {
	var version ApplicationVersion
	err := r.db.WithContext(ctx).
		Where("application_name = ? AND is_default = ?", appName, true).
		First(&version).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get default version for %s: %w", appName, err)
	}

	return &version, nil
}

// GetVersion returns a specific version.
func (r *Registry) GetVersion(ctx context.Context, appName, version string) (*ApplicationVersion, error) {
	var v ApplicationVersion
	err := r.db.WithContext(ctx).
		Where("application_name = ? AND version = ?", appName, version).
		First(&v).Error

	if err != nil {
		return nil, fmt.Errorf("version not found: %w", err)
	}

	return &v, nil
}

// GetAvailableVersions returns all available versions for a given app and tier.
func (r *Registry) GetAvailableVersions(ctx context.Context, appName, tier string) ([]ApplicationVersion, error) {
	var versions []ApplicationVersion

	// Get versions that are:
	// 1. Belonging to this app
	// 2. Not EOL
	// 3. Either have no tier restriction OR tier restriction <= current tier
	err := r.db.WithContext(ctx).
		Where("application_name = ?", appName).
		Where("status != ?", StatusEOL).
		Where("min_tier IS NULL OR min_tier <= ?", tier).
		Order("release_date DESC").
		Find(&versions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get available versions: %w", err)
	}

	return versions, nil
}

// GetStableVersions returns all stable versions for an app.
func (r *Registry) GetStableVersions(ctx context.Context, appName string) ([]ApplicationVersion, error) {
	var versions []ApplicationVersion

	err := r.db.WithContext(ctx).
		Where("application_name = ? AND status = ?", appName, StatusStable).
		Order("release_date DESC").
		Find(&versions).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get stable versions: %w", err)
	}

	return versions, nil
}

// ValidateVersion checks if a version exists and is not EOL.
func (r *Registry) ValidateVersion(ctx context.Context, appName, version string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ApplicationVersion{}).
		Where("application_name = ? AND version = ? AND status != ?", appName, version, StatusEOL).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// ValidateVersionForTier checks if a version is available for a specific tier.
func (r *Registry) ValidateVersionForTier(ctx context.Context, appName, version, tier string) (bool, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&ApplicationVersion{}).
		Where("application_name = ? AND version = ? AND status != ?", appName, version, StatusEOL).
		Where("min_tier IS NULL OR min_tier <= ?", tier).
		Count(&count).Error

	if err != nil {
		return false, err
	}

	return count > 0, nil
}

// SetDefaultVersion sets a version as the default for an application.
func (r *Registry) SetDefaultVersion(ctx context.Context, appName, version string) error {
	return r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// Unset current default for this app
		if err := tx.Model(&ApplicationVersion{}).
			Where("application_name = ? AND is_default = ?", appName, true).
			Update("is_default", false).Error; err != nil {
			return fmt.Errorf("failed to unset current default: %w", err)
		}

		// Set new default
		if err := tx.Model(&ApplicationVersion{}).
			Where("application_name = ? AND version = ?", appName, version).
			Update("is_default", true).Error; err != nil {
			return fmt.Errorf("failed to set new default: %w", err)
		}

		return nil
	})
}

// CreateVersion adds a new version to the registry.
func (r *Registry) CreateVersion(ctx context.Context, version *ApplicationVersion) error {
	err := r.db.WithContext(ctx).Create(version).Error
	if err != nil {
		return fmt.Errorf("failed to create version: %w", err)
	}
	return nil
}

// UpdateVersionStatus updates the status of a version.
func (r *Registry) UpdateVersionStatus(ctx context.Context, appName, version, status string) error {
	err := r.db.WithContext(ctx).
		Model(&ApplicationVersion{}).
		Where("application_name = ? AND version = ?", appName, version).
		Update("status", status).Error

	if err != nil {
		return fmt.Errorf("failed to update version status: %w", err)
	}

	return nil
}

// GetVersionStats returns statistics about version usage across instances for a specific app.
func (r *Registry) GetVersionStats(ctx context.Context, appName string) (map[string]int64, error) {
	// Note: Currently instances table doesn't have application_name column.
	// Assuming all instances are running 'railzway' for now or we will add app_name to instance table later.
	// For now, returning global stats matching known versions of this app.

	type VersionCount struct {
		Version string
		Count   int64
	}

	var results []VersionCount
	err := r.db.WithContext(ctx).
		Table("instances").
		Select("desired_version as version, COUNT(*) as count").
		Group("desired_version").
		Scan(&results).Error

	if err != nil {
		return nil, fmt.Errorf("failed to get version stats: %w", err)
	}

	stats := make(map[string]int64)
	for _, r := range results {
		stats[r.Version] = r.Count
	}

	return stats, nil
}

package organization

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Organization struct {
	ID   int64  `gorm:"primaryKey"`
	Slug string `gorm:"not null"`
}

func (Organization) TableName() string {
	return "organizations"
}

type Service struct {
	db *gorm.DB
}

func NewService(db *gorm.DB) *Service {
	return &Service{db: db}
}

func (s *Service) GetSlug(ctx context.Context, orgID int64) (string, error) {
	var org Organization
	if err := s.db.WithContext(ctx).Select("slug").First(&org, "id = ?", orgID).Error; err != nil {
		return "", fmt.Errorf("failed to fetch organization: %w", err)
	}
	slug := strings.TrimSpace(org.Slug)
	if slug == "" {
		return "", fmt.Errorf("organization slug is empty")
	}
	return slug, nil
}

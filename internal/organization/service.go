package organization

import (
	"context"
	"fmt"
	"strings"

	"gorm.io/gorm"
)

type Organization struct {
	ID   int64  `gorm:"column:id;primaryKey"`
	Slug string `gorm:"column:slug;not null"`
	Name string `gorm:"column:name;not null"`
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

func (s *Service) GetSlug(ctx context.Context, orgID int64) (*Organization, error) {
	var org Organization
	if err := s.db.WithContext(ctx).Select("*").First(&org, "id = ?", orgID).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch organization: %w", err)
	}
	slug := strings.TrimSpace(org.Slug)
	if slug == "" {
		return nil, fmt.Errorf("organization slug is empty")
	}
	return &org, nil
}

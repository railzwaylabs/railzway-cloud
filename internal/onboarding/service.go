package onboarding

import (
	"context"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/railzwaylabs/railzway-cloud/internal/config"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
	"github.com/railzwaylabs/railzway-cloud/internal/outbox"
	"github.com/railzwaylabs/railzway-cloud/internal/user"
	"github.com/railzwaylabs/railzway-cloud/internal/version"
	"github.com/railzwaylabs/railzway-cloud/pkg/snowflake"
	"gorm.io/gorm"
)

// slugify converts a string to a valid URL slug
func slugify(s string) string {
	// Lowercase
	s = strings.ToLower(s)
	// Replace non-alphanumeric with dashes
	reg := regexp.MustCompile("[^a-z0-9]+")
	s = reg.ReplaceAllString(s, "-")
	// Trim dashes
	s = strings.Trim(s, "-")
	return s
}

var orgSlugRegex = regexp.MustCompile("^[a-z0-9](?:[a-z0-9-]{0,61}[a-z0-9])?$")

func normalizeOrgSlug(raw string, rootDomain string) (string, error) {
	value := strings.ToLower(strings.TrimSpace(raw))
	if value == "" {
		return "", fmt.Errorf("organization namespace is required")
	}

	root := strings.Trim(strings.ToLower(strings.TrimSpace(rootDomain)), ".")
	if strings.Contains(value, ".") {
		if root == "" {
			return "", fmt.Errorf("namespace cannot include dots without root domain configured")
		}

		suffix := "." + root
		if !strings.HasSuffix(value, suffix) {
			return "", fmt.Errorf("namespace must end with %s", root)
		}

		prefix := strings.TrimSuffix(value, suffix)
		if strings.Contains(prefix, ".") {
			return "", fmt.Errorf("namespace must be a single subdomain")
		}
		value = prefix
	}

	if !orgSlugRegex.MatchString(value) {
		return "", fmt.Errorf("namespace must be 1-63 chars of lowercase letters, numbers, or hyphen (no leading/trailing hyphen)")
	}

	return value, nil
}

func buildLaunchURL(cfg *config.Config, slug string) string {
	root := strings.TrimSpace(cfg.AppRootDomain)
	if root == "" || strings.TrimSpace(slug) == "" {
		return ""
	}
	scheme := strings.TrimSpace(cfg.AppRootScheme)
	if scheme == "" {
		if strings.EqualFold(cfg.Environment, "production") {
			scheme = "https"
		} else {
			scheme = "http"
		}
	}
	host := fmt.Sprintf("%s.%s", slug, root)
	return fmt.Sprintf("%s://%s/login/railzway_com", scheme, host)
}

type Organization struct {
	ID            int64     `gorm:"primaryKey" json:"id,string"` // Fix: Use string tag for JS compatibility
	OwnerID       int64     `gorm:"not null" json:"owner_id,string"`
	Name          string    `gorm:"not null" json:"name"`
	Slug          string    `gorm:"not null;uniqueIndex" json:"slug"`
	OSSCustomerID string    `json:"oss_customer_id"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}

type Service struct {
	db         *gorm.DB
	cfg        *config.Config
	versionReg *version.Registry
	snowflake  *snowflake.Node
}

func NewService(
	db *gorm.DB,
	cfg *config.Config,
	versionReg *version.Registry,
	snowflake *snowflake.Node,
) *Service {
	return &Service{
		db:         db,
		cfg:        cfg,
		versionReg: versionReg,
		snowflake:  snowflake,
	}
}

type InitRequest struct {
	UserID  int64
	PlanID  string // Deprecated: use PriceID instead
	PriceID string // Actual price ID from pricing API
	OrgName string
	OrgSlug string
}

func (s *Service) InitializeOrganization(ctx context.Context, req InitRequest) (*Organization, error) {
	var org Organization
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		// 1. Validate / Get User
		var u user.User
		if err := tx.First(&u, req.UserID).Error; err != nil {
			return fmt.Errorf("user not found: %w", err)
		}

		// 2. Allocate IDs up front to anchor outbox + instance records.
		orgID := s.snowflake.GenerateID()

		// 3. Generate Slug
		if strings.TrimSpace(req.OrgName) == "" {
			return fmt.Errorf("organization name is required")
		}
		slug, err := normalizeOrgSlug(req.OrgSlug, s.cfg.AppRootDomain)
		if err != nil {
			return err
		}

		// 4. Create Organization (OSS fields resolved asynchronously via outbox).
		org = Organization{
			ID:        orgID,
			OwnerID:   req.UserID,
			Name:      req.OrgName,
			Slug:      slug,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}
		if err := tx.Create(&org).Error; err != nil {
			return fmt.Errorf("failed to create org: %w", err)
		}

		// 5. Create Instance Record
		priceID := strings.TrimSpace(req.PriceID)
		if priceID == "" {
			return fmt.Errorf("price_id is required")
		}
		tier := tierForPlan(req.PlanID)

		desiredVersion := s.cfg.DefaultRailzwayOSSVersion
		if s.versionReg != nil {
			if v, err := s.versionReg.GetDefaultVersion(ctx, "railzway"); err == nil && v != nil {
				desiredVersion = v.Version
			}
		}

		inst := instance.Instance{
			ID:             s.snowflake.GenerateID(),
			OrgID:          org.ID,
			Status:         instance.StatusInit,
			Role:           instance.RolePrimary,
			LifecycleState: instance.LifecycleReady,
			Readiness:      instance.ReadinessUnknown,
			NomadJobID:     fmt.Sprintf("railzway-org-%d", org.ID),
			DesiredVersion: desiredVersion,
			Tier:           tier,
			ComputeEngine:  instance.EngineGCP,
			PlanID:         req.PlanID,
			PriceID:        priceID,
			LaunchURL:      buildLaunchURL(s.cfg, slug),
			CreatedAt:      time.Now(),
			UpdatedAt:      time.Now(),
		}
		if err := tx.Create(&inst).Error; err != nil {
			return fmt.Errorf("failed to create instance: %w", err)
		}

		// 6. Create outbox event in the same transaction to make side effects durable.
		event := outbox.Event{
			EventType:  outbox.EventTypeDeployInstance,
			OrgID:      org.ID,
			InstanceID: inst.ID,
			Status:     outbox.StatusPending,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := tx.Create(&event).Error; err != nil {
			return fmt.Errorf("failed to create outbox event: %w", err)
		}

		return nil
	})

	if err != nil {
		return nil, err
	}

	return &org, nil
}

func (s *Service) GetOrganizationsByUserID(ctx context.Context, userID int64) ([]Organization, error) {
	var orgs []Organization
	if err := s.db.Where("owner_id = ?", userID).Order("created_at desc").Find(&orgs).Error; err != nil {
		return nil, fmt.Errorf("failed to fetch user orgs: %w", err)
	}
	return orgs, nil
}

func (s *Service) UserOwnsOrg(ctx context.Context, userID, orgID int64) (bool, error) {
	var count int64
	if err := s.db.WithContext(ctx).
		Model(&Organization{}).
		Where("id = ? AND owner_id = ?", orgID, userID).
		Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to verify org ownership: %w", err)
	}
	return count > 0, nil
}

func (s *Service) GetOrganizationSlug(ctx context.Context, orgID int64) (string, error) {
	var org Organization
	if err := s.db.WithContext(ctx).First(&org, "id = ?", orgID).Error; err != nil {
		return "", fmt.Errorf("failed to fetch organization: %w", err)
	}
	slug := strings.TrimSpace(org.Slug)
	if slug == "" {
		return "", fmt.Errorf("organization slug is empty")
	}
	return slug, nil
}

// CheckOrgName checks if an organization name/slug is available
func (s *Service) CheckOrgName(ctx context.Context, name string, namespace string) (bool, error) {
	var slug string
	if strings.TrimSpace(namespace) != "" {
		parsed, err := normalizeOrgSlug(namespace, s.cfg.AppRootDomain)
		if err != nil {
			return false, err
		}
		slug = parsed
	} else {
		slug = slugify(name)
		if slug == "" {
			return false, fmt.Errorf("invalid organization name")
		}
	}

	var count int64
	if err := s.db.Model(&Organization{}).Where("slug = ?", slug).Count(&count).Error; err != nil {
		return false, fmt.Errorf("failed to check organization name: %w", err)
	}

	return count == 0, nil
}

func tierForPlan(planID string) instance.Tier {
	switch strings.ToLower(planID) {
	case "free trial", "free-trial":
		return instance.TierFreeTrial
	case "starter":
		return instance.TierStarter
	case "pro", "production":
		return instance.TierPro
	case "team", "performance":
		return instance.TierTeam
	case "enterprise":
		return instance.TierEnterprise
	default:
		// Default to free trial for unknown plans
		return instance.TierFreeTrial
	}
}

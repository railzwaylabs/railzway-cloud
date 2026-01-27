package user

import (
	"context"
	"time"

	"github.com/railzwaylabs/railzway-cloud/internal/config"
	"github.com/railzwaylabs/railzway-cloud/pkg/railzwayclient"
	"github.com/railzwaylabs/railzway-cloud/pkg/snowflake"
	"gorm.io/gorm"
)

type User struct {
	ID        int64  `gorm:"primaryKey"`
	Email     string `gorm:"uniqueIndex;not null"`
	AuthID    string `gorm:"uniqueIndex;not null"`
	FirstName string `gorm:"type:varchar(100)"`
	LastName  string `gorm:"type:varchar(100)"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

type Service struct {
	db        *gorm.DB
	ossClient *railzwayclient.Client
	cfg       *config.Config
	snowflake *snowflake.Node
}

func NewService(db *gorm.DB, ossClient *railzwayclient.Client, cfg *config.Config, snowflake *snowflake.Node) *Service {
	return &Service{
		db:        db,
		ossClient: ossClient,
		cfg:       cfg,
		snowflake: snowflake,
	}
}

// EnsureUser ensures a user exists in the local DB and is synced to the OSS instance as a Customer.
func (s *Service) EnsureUser(
	ctx context.Context,
	authID string,
	email string,
) (*User, error) {

	// Always operate inside a transaction for local consistency
	returnValue := &User{}

	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		var user User

		// 1. Try to find existing user by AuthID
		err := tx.Where("auth_id = ?", authID).First(&user).Error
		if err == nil {
			// User exists locally
			*returnValue = user
			return nil
		}

		if err != nil && err != gorm.ErrRecordNotFound {
			return err
		}

		// 2. Create local user (Billing is now deferred to Organization creation)
		user = User{
			ID:        s.snowflake.GenerateID(),
			Email:     email,
			AuthID:    authID,
			CreatedAt: time.Now(),
			UpdatedAt: time.Now(),
		}

		if err := tx.Create(&user).Error; err != nil {
			return err
		}

		*returnValue = user
		return nil
	})

	if err != nil {
		return nil, err
	}

	return returnValue, nil
}

func (s *Service) GetByID(ctx context.Context, userID int64) (*User, error) {
	var user User
	if err := s.db.WithContext(ctx).First(&user, "id = ?", userID).Error; err != nil {
		return nil, err
	}
	return &user, nil
}

func (s *Service) UpdateProfile(ctx context.Context, userID int64, firstName, lastName string) (*User, error) {
	if err := s.db.WithContext(ctx).Model(&User{}).Where("id = ?", userID).Updates(map[string]interface{}{
		"first_name": firstName,
		"last_name":  lastName,
		"updated_at": time.Now(),
	}).Error; err != nil {
		return nil, err
	}
	return s.GetByID(ctx, userID)
}

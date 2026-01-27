package version_test

import (
	"context"
	"testing"
	"time"

	"github.com/railzwaylabs/railzway-cloud/internal/version"
	"github.com/railzwaylabs/railzway-cloud/pkg/testhelper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func TestRegistry_Integration(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test in short mode")
	}

	ctx := context.Background()

	// 1. Setup Container
	pg, err := testhelper.SetupPostgres(ctx)
	require.NoError(t, err)
	defer func() {
		if err := pg.Teardown(ctx); err != nil {
			t.Logf("failed to teardown container: %v", err)
		}
	}()

	// 2. Connect to DB
	db, err := gorm.Open(postgres.Open(pg.DSN), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Silent),
	})
	require.NoError(t, err)

	// 3. Migrate Schema
	err = db.AutoMigrate(&version.ApplicationVersion{})
	require.NoError(t, err)

	// 4. Initialize Registry
	reg := version.NewRegistry(db)

	t.Run("CreateVersion", func(t *testing.T) {
		v := &version.ApplicationVersion{
			ApplicationName: "railzway",
			Version:         "v1.0.0",
			Status:          version.StatusStable,
			ReleaseDate:     time.Now(),
			DockerImage:     "ghcr.io/smallbiznis/railzway:v1.0.0",
		}
		err := reg.CreateVersion(ctx, v)
		require.NoError(t, err)

		// Verify it exists
		fetched, err := reg.GetVersion(ctx, "railzway", "v1.0.0")
		require.NoError(t, err)
		assert.Equal(t, "v1.0.0", fetched.Version)
		assert.Equal(t, "stable", fetched.Status)
	})

	t.Run("SetDefaultVersion", func(t *testing.T) {
		// Create another version
		v2 := &version.ApplicationVersion{
			ApplicationName: "railzway",
			Version:         "v1.1.0",
			Status:          version.StatusStable,
			ReleaseDate:     time.Now(),
			DockerImage:     "ghcr.io/smallbiznis/railzway:v1.1.0",
		}
		require.NoError(t, reg.CreateVersion(ctx, v2))

		// Set v1.1.0 as default
		err := reg.SetDefaultVersion(ctx, "railzway", "v1.1.0")
		require.NoError(t, err)

		// Verify default
		def, err := reg.GetDefaultVersion(ctx, "railzway")
		require.NoError(t, err)
		assert.Equal(t, "v1.1.0", def.Version)
		assert.True(t, def.IsDefault)

		// Verify v1.0.0 is NOT default
		v1, err := reg.GetVersion(ctx, "railzway", "v1.0.0")
		require.NoError(t, err)
		assert.False(t, v1.IsDefault)
	})

	t.Run("GetAvailableVersions_TierFiltering", func(t *testing.T) {
		// Create a version restricted to STARTER tier
		vPro := &version.ApplicationVersion{
			ApplicationName: "railzway",
			Version:         "v2.0.0-pro",
			Status:          version.StatusStable,
			ReleaseDate:     time.Now(),
			DockerImage:     "ghcr.io/smallbiznis/railzway:v2.0.0-pro",
			MinTier:         stringPtr("STARTER"),
		}
		require.NoError(t, reg.CreateVersion(ctx, vPro))

		// User on FREE tier should NOT see v2.0.0-pro
		versionsFree, err := reg.GetAvailableVersions(ctx, "railzway", "FREE")
		require.NoError(t, err)

		found := false
		for _, v := range versionsFree {
			if v.Version == "v2.0.0-pro" {
				found = true
			}
		}
		assert.False(t, found, "FREE tier should not see STARTER version")

		// User on STARTER tier SHOULD see v2.0.0-pro
		versionsStarter, err := reg.GetAvailableVersions(ctx, "railzway", "STARTER")
		require.NoError(t, err)

		found = false
		for _, v := range versionsStarter {
			if v.Version == "v2.0.0-pro" {
				found = true
			}
		}
		assert.True(t, found, "STARTER tier should see STARTER version")
	})
}

func stringPtr(s string) *string {
	return &s
}

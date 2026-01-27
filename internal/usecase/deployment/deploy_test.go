package deployment

import (
	"context"
	"testing"

	"github.com/railzwaylabs/railzway-cloud/internal/config"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/instance"
	"github.com/railzwaylabs/railzway-cloud/internal/domain/provisioning"
	"github.com/railzwaylabs/railzway-cloud/pkg/authclient"
	"github.com/railzwaylabs/railzway-cloud/pkg/testhelper"
	"github.com/stretchr/testify/assert"
)

// mockInstanceRepository is a simple in-memory repository for testing
type mockInstanceRepository struct {
	instances map[int64]*instance.Instance
}

func newMockInstanceRepository() *mockInstanceRepository {
	return &mockInstanceRepository{
		instances: make(map[int64]*instance.Instance),
	}
}

func (m *mockInstanceRepository) FindByOrgID(ctx context.Context, orgID int64) (*instance.Instance, error) {
	inst, ok := m.instances[orgID]
	if !ok {
		return nil, nil
	}
	return inst, nil
}

func (m *mockInstanceRepository) Save(ctx context.Context, inst *instance.Instance) error {
	m.instances[inst.OrgID] = inst
	return nil
}

func (m *mockInstanceRepository) UpdateStatus(ctx context.Context, orgID int64, status instance.InstanceStatus) error {
	inst, ok := m.instances[orgID]
	if !ok {
		return nil
	}
	inst.Status = status
	return nil
}

func (m *mockInstanceRepository) ListByStatus(ctx context.Context, statuses []instance.InstanceStatus, limit int) ([]*instance.Instance, error) {
	var result []*instance.Instance
	for _, inst := range m.instances {
		for _, status := range statuses {
			if inst.Status == status {
				result = append(result, inst)
				break
			}
		}
		if len(result) >= limit {
			break
		}
	}
	return result, nil
}

// Helper to create a test DeployUseCase with mocks
func newTestDeployUseCase(
	repo instance.Repository,
	provisioner provisioning.Provisioner,
	dbProvisioner provisioning.DatabaseProvisioner,
	authClient *authclient.Client,
	cfg *config.Config,
) *DeployUseCase {
	dbConfig := provisioning.DBConfig{
		Host: "localhost",
		Port: 5432,
	}

	runtimeCfg := RuntimeConfig{
		RateLimitRedisAddr: "localhost:6379",
	}

	// Create DeployUseCase without org service (will be nil, tests will skip org slug logic)
	uc := &DeployUseCase{
		repo:          repo,
		provisioner:   provisioner,
		dbProvisioner: dbProvisioner,
		dbConfig:      dbConfig,
		runtimeCfg:    runtimeCfg,
		orgService:    nil, // Skip org service for unit tests
		cfg:           cfg,
		authClient:    authClient,
	}

	return uc
}

func TestDeployUseCase_Execute_InstanceNotFound(t *testing.T) {
	mockRepo := newMockInstanceRepository()

	uc := newTestDeployUseCase(
		mockRepo,
		&testhelper.MockProvisioner{},
		&testhelper.MockDatabaseProvisioner{},
		nil,
		&config.Config{},
	)

	err := uc.Execute(context.Background(), 999, "v1.0.0")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "instance not found")
}

// Note: Full integration tests with org service require actual database
// These unit tests focus on the deployment logic without external dependencies

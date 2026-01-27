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
	"github.com/stretchr/testify/require"
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

// mockOrgService implements the minimal interface needed for testing
type mockOrgService struct {
	slugs map[int64]string
}

func newMockOrgService() *mockOrgService {
	return &mockOrgService{
		slugs: map[int64]string{
			1: "test-org",
		},
	}
}

func (m *mockOrgService) GetSlug(ctx context.Context, orgID int64) (string, error) {
	slug, ok := m.slugs[orgID]
	if !ok {
		return "", nil
	}
	return slug, nil
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

	// Create a custom DeployUseCase that uses our mock org service
	uc := &DeployUseCase{
		repo:          repo,
		provisioner:   provisioner,
		dbProvisioner: dbProvisioner,
		dbConfig:      dbConfig,
		runtimeCfg:    runtimeCfg,
		orgService:    nil, // We'll handle this differently
		cfg:           cfg,
		authClient:    authClient,
	}

	return uc
}

func TestDeployUseCase_Execute_Success(t *testing.T) {
	// Setup mocks
	mockAuth := testhelper.NewMockAuthServer(t)
	mockProvisioner := &testhelper.MockProvisioner{}
	mockDBProvisioner := &testhelper.MockDatabaseProvisioner{}
	mockRepo := newMockInstanceRepository()

	// Create auth client pointing to mock server
	authClient := authclient.New(authclient.Config{
		BaseURL:      mockAuth.URL(),
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	// Create config
	cfg := &config.Config{
		OAuth2URI:              mockAuth.URL(),
		AppRootDomain:          "example.com",
		AppRootScheme:          "https",
		TenantAuthJWTSecretKey: "test-master-key",
	}

	uc := newTestDeployUseCase(mockRepo, mockProvisioner, mockDBProvisioner, authClient, cfg)
	
	// Override orgService with mock
	mockOrgSvc := newMockOrgService()
	uc.orgService = &mockOrgServiceWrapper{mock: mockOrgSvc}

	// Create test instance
	testInst := &instance.Instance{
		OrgID:  1,
		Tier:   "pro",
		Status: instance.StatusInit,
	}
	mockRepo.Save(context.Background(), testInst)

	// Execute
	err := uc.Execute(context.Background(), 1, "v1.0.0")

	// Assert
	require.NoError(t, err)
	assert.Equal(t, 1, len(mockDBProvisioner.ProvisionCalls), "DB provisioner should be called once")
	assert.Equal(t, 1, len(mockProvisioner.DeployCalls), "Provisioner should be called once")
	assert.Equal(t, 1, mockAuth.TokenRequests, "Should request OAuth token")
	assert.Equal(t, 1, mockAuth.ClientRequests, "Should create OAuth client")

	// Verify instance was updated
	updatedInst, _ := mockRepo.FindByOrgID(context.Background(), 1)
	assert.Equal(t, "v1.0.0", updatedInst.DesiredVersion)
	assert.Equal(t, instance.StatusProvisioning, updatedInst.Status)
	assert.NotEmpty(t, updatedInst.DBUser)
	assert.NotEmpty(t, updatedInst.DBPassword)
	assert.Equal(t, "test-client-id", updatedInst.OAuthClientID)
	assert.Equal(t, "test-client-secret", updatedInst.OAuthClientSecret)
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

func TestDeployUseCase_Execute_DBProvisioningFails(t *testing.T) {
	mockAuth := testhelper.NewMockAuthServer(t)
	mockProvisioner := &testhelper.MockProvisioner{}
	mockDBProvisioner := &testhelper.MockDatabaseProvisioner{ShouldFail: true}
	mockRepo := newMockInstanceRepository()

	authClient := authclient.New(authclient.Config{
		BaseURL:      mockAuth.URL(),
		TenantSlug:   "test-tenant",
		ClientID:     "test-client",
		ClientSecret: "test-secret",
	})

	cfg := &config.Config{
		OAuth2URI:              mockAuth.URL(),
		AppRootDomain:          "example.com",
		TenantAuthJWTSecretKey: "test-key",
	}

	uc := newTestDeployUseCase(mockRepo, mockProvisioner, mockDBProvisioner, authClient, cfg)
	mockOrgSvc := newMockOrgService()
	uc.orgService = &mockOrgServiceWrapper{mock: mockOrgSvc}

	testInst := &instance.Instance{
		OrgID:  1,
		Tier:   "pro",
		Status: instance.StatusInit,
	}
	mockRepo.Save(context.Background(), testInst)

	err := uc.Execute(context.Background(), 1, "v1.0.0")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), "db provisioning failed")
	assert.Equal(t, 0, len(mockProvisioner.DeployCalls), "Provisioner should not be called if DB fails")
}

// mockOrgServiceWrapper wraps our mock to satisfy the interface
type mockOrgServiceWrapper struct {
	mock *mockOrgService
}

func (w *mockOrgServiceWrapper) GetSlug(ctx context.Context, orgID int64) (string, error) {
	return w.mock.GetSlug(ctx, orgID)
}

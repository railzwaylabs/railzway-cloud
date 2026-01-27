package testhelper

import (
	"context"
	"fmt"

	"github.com/railzwaylabs/railzway-cloud/internal/domain/provisioning"
)

// MockProvisioner is a mock implementation of provisioning.Provisioner for testing
type MockProvisioner struct {
	DeployCalls []provisioning.DeploymentConfig
	ShouldFail  bool
}

// Deploy mocks the Deploy method
func (m *MockProvisioner) Deploy(ctx context.Context, cfg *provisioning.DeploymentConfig) error {
	if m.ShouldFail {
		return fmt.Errorf("mock provisioner: deployment failed")
	}
	m.DeployCalls = append(m.DeployCalls, *cfg)
	return nil
}

// Stop mocks the Stop method
func (m *MockProvisioner) Stop(ctx context.Context, orgID int64) error {
	if m.ShouldFail {
		return fmt.Errorf("mock provisioner: stop failed")
	}
	return nil
}

// GetStatus mocks the GetStatus method
func (m *MockProvisioner) GetStatus(ctx context.Context, orgID int64) (string, error) {
	if m.ShouldFail {
		return "", fmt.Errorf("mock provisioner: get status failed")
	}
	return "running", nil
}

// MockDatabaseProvisioner is a mock implementation of provisioning.DatabaseProvisioner
type MockDatabaseProvisioner struct {
	ProvisionCalls []int64
	ShouldFail     bool
}

// Provision mocks the Provision method
func (m *MockDatabaseProvisioner) Provision(ctx context.Context, orgID int64, password string) error {
	if m.ShouldFail {
		return fmt.Errorf("mock db provisioner: provision failed")
	}
	m.ProvisionCalls = append(m.ProvisionCalls, orgID)
	return nil
}

package instance

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewInstance(t *testing.T) {
	inst := NewInstance(123, TierPro, EngineGCP, "v1.0.0")

	assert.Equal(t, int64(123), inst.OrgID)
	assert.Equal(t, TierPro, inst.Tier)
	assert.Equal(t, EngineGCP, inst.ComputeEngine)
	assert.Equal(t, "v1.0.0", inst.DesiredVersion)
	assert.Equal(t, StatusInit, inst.Status)
	assert.Equal(t, RolePrimary, inst.Role)
	assert.Equal(t, LifecycleReady, inst.LifecycleState)
	assert.Equal(t, ReadinessUnknown, inst.Readiness)
	assert.NotZero(t, inst.CreatedAt)
	assert.NotZero(t, inst.UpdatedAt)
}

func TestCanTransitionLifecycle_ValidTransitions(t *testing.T) {
	tests := []struct {
		name    string
		current LifecycleState
		target  LifecycleState
		want    bool
	}{
		// Same state
		{"same state", LifecycleReady, LifecycleReady, true},
		
		// Ready transitions
		{"ready to serving", LifecycleReady, LifecycleServing, true},
		{"ready to migrating", LifecycleReady, LifecycleMigrating, true},
		
		// Migrating transitions
		{"migrating to serving", LifecycleMigrating, LifecycleServing, true},
		
		// Serving transitions
		{"serving to draining", LifecycleServing, LifecycleDraining, true},
		
		// Draining transitions
		{"draining to ready", LifecycleDraining, LifecycleReady, true},
		{"draining to terminated", LifecycleDraining, LifecycleTerminated, true},
		
		// Invalid transitions
		{"ready to draining", LifecycleReady, LifecycleDraining, false},
		{"serving to migrating", LifecycleServing, LifecycleMigrating, false},
		{"migrating to terminated", LifecycleMigrating, LifecycleTerminated, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := CanTransitionLifecycle(tt.current, tt.target)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestCanTransitionLifecycle_EmptyTarget(t *testing.T) {
	// Empty target should always be allowed (no-op)
	assert.True(t, CanTransitionLifecycle(LifecycleReady, ""))
	assert.True(t, CanTransitionLifecycle(LifecycleServing, ""))
}

func TestTierRank(t *testing.T) {
	assert.Equal(t, 0, TierRank[TierFreeTrial])
	assert.Equal(t, 1, TierRank[TierStarter])
	assert.Equal(t, 2, TierRank[TierPro])
	assert.Equal(t, 3, TierRank[TierTeam])
	assert.Equal(t, 4, TierRank[TierEnterprise])
}

func TestTierUpgrade_ValidUpgrades(t *testing.T) {
	tests := []struct {
		name    string
		current Tier
		target  Tier
		isValid bool
	}{
		{"free trial to starter", TierFreeTrial, TierStarter, true},
		{"starter to pro", TierStarter, TierPro, true},
		{"pro to team", TierPro, TierTeam, true},
		{"team to enterprise", TierTeam, TierEnterprise, true},
		{"free trial to enterprise", TierFreeTrial, TierEnterprise, true},
		
		// Downgrades (should be invalid)
		{"pro to starter", TierPro, TierStarter, false},
		{"enterprise to team", TierEnterprise, TierTeam, false},
		
		// Same tier
		{"pro to pro", TierPro, TierPro, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			currentRank := TierRank[tt.current]
			targetRank := TierRank[tt.target]
			isUpgrade := targetRank >= currentRank
			assert.Equal(t, tt.isValid, isUpgrade)
		})
	}
}

func TestInstanceStatus_Constants(t *testing.T) {
	// Verify all status constants are defined
	statuses := []InstanceStatus{
		StatusInit,
		StatusProvisioning,
		StatusActive,
		StatusProvisionFailed,
		StatusRunning,
		StatusStopped,
		StatusUpgrading,
		StatusDowngradeScheduled,
		StatusTerminated,
	}

	for _, status := range statuses {
		assert.NotEmpty(t, status)
	}
}

func TestLifecycleState_Constants(t *testing.T) {
	// Verify all lifecycle state constants are defined
	states := []LifecycleState{
		LifecycleReady,
		LifecycleMigrating,
		LifecycleServing,
		LifecycleDraining,
		LifecycleTerminated,
	}

	for _, state := range states {
		assert.NotEmpty(t, state)
	}
}

func TestReadinessStatus_Constants(t *testing.T) {
	// Verify all readiness status constants are defined
	statuses := []ReadinessStatus{
		ReadinessUnknown,
		ReadinessReady,
		ReadinessNotReady,
	}

	for _, status := range statuses {
		assert.NotEmpty(t, status)
	}
}

func TestInstanceRole_Constants(t *testing.T) {
	assert.Equal(t, InstanceRole("primary"), RolePrimary)
	assert.Equal(t, InstanceRole("standby"), RoleStandby)
}

func TestComputeEngine_Constants(t *testing.T) {
	engines := []ComputeEngine{
		EngineHetzner,
		EngineDigitalOcean,
		EngineGCP,
		EngineAWS,
	}

	for _, engine := range engines {
		assert.NotEmpty(t, engine)
	}
}

func TestErrors(t *testing.T) {
	assert.NotNil(t, ErrInvalidTierUpgrade)
	assert.NotNil(t, ErrInvalidState)
	assert.Contains(t, ErrInvalidTierUpgrade.Error(), "invalid tier upgrade")
	assert.Contains(t, ErrInvalidState.Error(), "invalid instance state")
}

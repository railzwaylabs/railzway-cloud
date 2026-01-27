package instance

import (
	"errors"
	"time"
)

// Tier represents the subscription level.
type Tier string

const (
	TierFreeTrial  Tier = "FREE_TRIAL"
	TierStarter    Tier = "STARTER"
	TierPro        Tier = "PRO"
	TierTeam       Tier = "TEAM"
	TierEnterprise Tier = "ENTERPRISE"
)

var TierRank = map[Tier]int{
	TierFreeTrial:  0,
	TierStarter:    1,
	TierPro:        2,
	TierTeam:       3,
	TierEnterprise: 4,
}

// ComputeEngine represents the underlying infrastructure provider.
type ComputeEngine string

const (
	EngineHetzner      ComputeEngine = "hetzner"
	EngineDigitalOcean ComputeEngine = "digitalocean"
	EngineGCP          ComputeEngine = "gcp"
	EngineAWS          ComputeEngine = "aws"
)

// InstanceRole represents the deployment role of an instance.
type InstanceRole string

const (
	RolePrimary InstanceRole = "primary"
	RoleStandby InstanceRole = "standby"
)

// LifecycleState represents the control-plane lifecycle state for an OSS instance.
type LifecycleState string

const (
	LifecycleReady      LifecycleState = "ready"
	LifecycleMigrating  LifecycleState = "migrating"
	LifecycleServing    LifecycleState = "serving"
	LifecycleDraining   LifecycleState = "draining"
	LifecycleTerminated LifecycleState = "terminated"
)

// ReadinessStatus represents the latest /ready evaluation.
type ReadinessStatus string

const (
	ReadinessUnknown  ReadinessStatus = "unknown"
	ReadinessReady    ReadinessStatus = "ready"
	ReadinessNotReady ReadinessStatus = "not_ready"
)

// InstanceStatus represents the lifecycle state of an instance.
type InstanceStatus string

const (
	StatusInit               InstanceStatus = "init"
	StatusProvisioning       InstanceStatus = "provisioning"
	StatusActive             InstanceStatus = "active"
	StatusProvisionFailed    InstanceStatus = "provision_failed"
	StatusRunning            InstanceStatus = "running"
	StatusStopped            InstanceStatus = "stopped"
	StatusUpgrading          InstanceStatus = "upgrading"
	StatusDowngradeScheduled InstanceStatus = "downgrade_scheduled"
	StatusTerminated         InstanceStatus = "terminated"
)

var (
	ErrInvalidTierUpgrade = errors.New("invalid tier upgrade")
	ErrInvalidState       = errors.New("invalid instance state for operation")
)

// Instance is the core domain entity.
// It contains no database tags or infrastructure details.
type Instance struct {
	ID                 int64           `gorm:"column:id" json:"id,string"`
	OrgID              int64           `gorm:"column:org_id" json:"org_id,string"`
	NomadJobID         string          `gorm:"column:nomad_job_id" json:"nomad_job_id"`
	DesiredVersion     string          `gorm:"column:desired_version" json:"desired_version"`
	CurrentVersion     string          `gorm:"column:current_version" json:"current_version"`
	Status             InstanceStatus  `gorm:"column:status" json:"status"`
	Role               InstanceRole    `gorm:"column:role" json:"role"`
	LifecycleState     LifecycleState  `gorm:"column:lifecycle_state" json:"lifecycle_state"`
	Readiness          ReadinessStatus `gorm:"column:readiness_status" json:"readiness_status"`
	ReadinessCheckedAt *time.Time      `gorm:"column:readiness_checked_at" json:"readiness_checked_at,omitempty"`
	ReadinessError     string          `gorm:"column:readiness_error" json:"readiness_error,omitempty"`
	Tier               Tier            `gorm:"column:tier" json:"tier"`
	ComputeEngine      ComputeEngine   `gorm:"column:compute_engine" json:"compute_engine"`
	PlanID             string          `gorm:"column:plan_id" json:"plan_id"`
	PriceID            string          `gorm:"column:price_id" json:"price_id"`
	SubscriptionID     string          `gorm:"column:subscription_id" json:"subscription_id"` // Reference to Railzway OSS Subscription
	LaunchURL          string          `gorm:"column:launch_url" json:"launch_url"`
	LastError          string          `gorm:"column:last_error" json:"last_error,omitempty"`

	OAuthClientID                        string `gorm:"column:oauth_client_id" json:"-"`
	OAuthClientSecret                    string `gorm:"column:oauth_client_secret" json:"-"`
	PaymentProviderConfigSecretEncrypted string `gorm:"column:payment_provider_config_secret" json:"-"` // Encrypted at rest

	// Database Details (Managed by Railzway Cloud)
	DBHost     string `gorm:"column:db_host" json:"db_host"`
	DBPort     int    `gorm:"column:db_port" json:"db_port"`
	DBName     string `gorm:"column:db_name" json:"db_name"`
	DBUser     string `gorm:"column:db_user" json:"db_user"`
	DBPassword string `gorm:"column:db_password" json:"-"` // Encrypted at rest, do not expose

	CreatedAt time.Time `gorm:"column:created_at" json:"created_at"`
	UpdatedAt time.Time `gorm:"column:updated_at" json:"updated_at"`
}

// NewInstance creates a new instance in init state.
func NewInstance(orgID int64, tier Tier, engine ComputeEngine, version string) *Instance {
	return &Instance{
		OrgID:          orgID,
		Tier:           tier,
		ComputeEngine:  engine,
		DesiredVersion: version,
		Status:         StatusInit,
		Role:           RolePrimary,
		LifecycleState: LifecycleReady,
		Readiness:      ReadinessUnknown,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
	}
}

// CanTransitionLifecycle enforces the lifecycle state machine.
func CanTransitionLifecycle(current, target LifecycleState) bool {
	if current == target || target == "" {
		return true
	}
	switch current {
	case LifecycleReady:
		return target == LifecycleServing || target == LifecycleMigrating
	case LifecycleMigrating:
		return target == LifecycleServing
	case LifecycleServing:
		return target == LifecycleDraining
	case LifecycleDraining:
		return target == LifecycleReady || target == LifecycleTerminated
	default:
		return false
	}
}

// AllStatuses returns known instance status values.
func AllStatuses() []InstanceStatus {
	return []InstanceStatus{
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
}

// CanUpgrade checks if the instance can be upgraded to the target tier.
func (i *Instance) CanUpgrade(target Tier) bool {
	return TierRank[target] > TierRank[i.Tier]
}

// CanDowngrade checks if the instance can be downgraded to the target tier.
func (i *Instance) CanDowngrade(target Tier) bool {
	return TierRank[target] < TierRank[i.Tier]
}

// MarkRunning transitions the instance to Running state.
func (i *Instance) MarkRunning(currentVersion string) {
	i.Status = StatusRunning
	i.CurrentVersion = currentVersion
	i.UpdatedAt = time.Now().UTC()
}

// MarkProvisioning transitions the instance to provisioning state.
func (i *Instance) MarkProvisioning() {
	i.Status = StatusProvisioning
	i.LastError = ""
	i.UpdatedAt = time.Now().UTC()
}

// MarkActive transitions the instance to active state.
func (i *Instance) MarkActive(currentVersion string) {
	i.Status = StatusActive
	i.CurrentVersion = currentVersion
	i.LastError = ""
	i.UpdatedAt = time.Now().UTC()
}

// MarkProvisionFailed transitions the instance to provision_failed state.
func (i *Instance) MarkProvisionFailed(errMsg string) {
	i.Status = StatusProvisionFailed
	i.LastError = errMsg
	i.UpdatedAt = time.Now().UTC()
}

// MarkStopped transitions the instance to Stopped state.
func (i *Instance) MarkStopped() {
	i.Status = StatusStopped
	i.UpdatedAt = time.Now().UTC()
}

// MarkUpgrading transitions to Upgrading state.
func (i *Instance) MarkUpgrading(targetTier Tier) {
	i.Tier = targetTier
	i.Status = StatusUpgrading
	i.UpdatedAt = time.Now().UTC()
}

// ScheduleDowngrade marks the instance for downgrade at period end.
func (i *Instance) ScheduleDowngrade() {
	i.Status = StatusDowngradeScheduled
	i.UpdatedAt = time.Now().UTC()
}

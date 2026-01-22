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
	ID             int64          `json:"id,string"`
	OrgID          int64          `json:"org_id,string"`
	NomadJobID     string         `json:"nomad_job_id"`
	DesiredVersion string         `json:"desired_version"`
	CurrentVersion string         `json:"current_version"`
	Status         InstanceStatus `json:"status"`
	Tier           Tier           `json:"tier"`
	ComputeEngine  ComputeEngine  `json:"compute_engine"`
	PlanID         string         `json:"plan_id"`
	PriceID        string         `json:"price_id"`
	SubscriptionID string         `json:"subscription_id"` // Reference to Railzway OSS Subscription
	LaunchURL      string         `json:"launch_url"`
	LastError      string         `json:"last_error,omitempty"`

	OAuthClientID     string `json:"-"`
	OAuthClientSecret string `json:"-"`

	// Database Details (Managed by Railzway Cloud)
	DBHost     string `json:"db_host"`
	DBPort     int    `json:"db_port"`
	DBName     string `json:"db_name"`
	DBUser     string `json:"db_user"`
	DBPassword string `json:"-"` // Encrypted at rest, do not expose

	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// NewInstance creates a new instance in init state.
func NewInstance(orgID int64, tier Tier, engine ComputeEngine, version string) *Instance {
	return &Instance{
		OrgID:          orgID,
		Tier:           tier,
		ComputeEngine:  engine,
		DesiredVersion: version,
		Status:         StatusInit,
		CreatedAt:      time.Now().UTC(),
		UpdatedAt:      time.Now().UTC(),
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

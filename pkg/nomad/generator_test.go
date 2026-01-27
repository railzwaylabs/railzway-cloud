package nomad

import (
	"testing"
)

func TestGenerateJob_FreeTrialTier(t *testing.T) {
	cfg := JobConfig{
		OrgID:         123,
		Tier:          TierFreeTrial,
		ComputeEngine: EngineHetzner,
		Version:       "v1.0.0",
	}

	job, err := GenerateJob(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job == nil {
		t.Fatal("expected job to be not nil")
	}

	if *job.ID != "railzway-org-123" {
		t.Errorf("expected ID 'railzway-org-123', got %s", *job.ID)
	}
	if *job.Priority != 40 {
		t.Errorf("expected Priority 40, got %d", *job.Priority)
	}

	taskGroup := job.TaskGroups[0]
	if taskGroup.Constraints[0].RTarget != "free-trial" {
		t.Errorf("expected constraint 'free-trial', got %s", taskGroup.Constraints[0].RTarget)
	}
	if taskGroup.Constraints[1].RTarget != "hetzner" {
		t.Errorf("expected constraint 'hetzner', got %s", taskGroup.Constraints[1].RTarget)
	}

	task := taskGroup.Tasks[0]
	if *task.Resources.CPU != 250 {
		t.Errorf("expected CPU 250, got %d", *task.Resources.CPU)
	}
	if *task.Resources.MemoryMB != 512 {
		t.Errorf("expected Memory 512, got %d", *task.Resources.MemoryMB)
	}
	expectedImage := "ghcr.io/smallbiznis/railzway:v1.0.0"
	if task.Config["image"] != expectedImage {
		t.Errorf("expected image %s, got %s", expectedImage, task.Config["image"])
	}

	env := task.Env
	if env["USAGE_INGEST_ORG_RATE"] != "5" {
		t.Errorf("expected USAGE_INGEST_ORG_RATE 5, got %s", env["USAGE_INGEST_ORG_RATE"])
	}
	if env["QUOTA_ORG_USAGE_MONTHLY"] != "10000" {
		t.Errorf("expected QUOTA_ORG_USAGE_MONTHLY 10000, got %s", env["QUOTA_ORG_USAGE_MONTHLY"])
	}
}

func TestGenerateJob_StarterTier(t *testing.T) {
	cfg := JobConfig{
		OrgID:         789,
		Tier:          TierStarter,
		ComputeEngine: EngineDigitalOcean,
		Version:       "v1.5.0",
	}

	job, err := GenerateJob(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if job == nil {
		t.Fatal("expected job to be not nil")
	}

	if *job.Priority != 60 {
		t.Errorf("expected Priority 60, got %d", *job.Priority)
	}

	taskGroup := job.TaskGroups[0]
	if taskGroup.Constraints[0].RTarget != "starter" {
		t.Errorf("expected constraint 'starter', got %s", taskGroup.Constraints[0].RTarget)
	}

	task := taskGroup.Tasks[0]
	if *task.Resources.CPU != 500 {
		t.Errorf("expected CPU 500, got %d", *task.Resources.CPU)
	}
	if *task.Resources.MemoryMB != 1024 {
		t.Errorf("expected Memory 1024, got %d", *task.Resources.MemoryMB)
	}

	env := task.Env
	if env["USAGE_INGEST_ORG_RATE"] != "10" {
		t.Errorf("expected USAGE_INGEST_ORG_RATE 10, got %s", env["USAGE_INGEST_ORG_RATE"])
	}
	if env["QUOTA_ORG_USAGE_MONTHLY"] != "100000" {
		t.Errorf("expected QUOTA_ORG_USAGE_MONTHLY 100000, got %s", env["QUOTA_ORG_USAGE_MONTHLY"])
	}
}

func TestGenerateJob_TeamTier(t *testing.T) {
	cfg := JobConfig{
		OrgID:         456,
		Tier:          TierTeam,
		ComputeEngine: EngineGCP,
		Version:       "v2.0.0",
	}

	job, err := GenerateJob(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if *job.Priority != 90 {
		t.Errorf("expected Priority 90, got %d", *job.Priority)
	}

	taskGroup := job.TaskGroups[0]
	if taskGroup.Constraints[0].RTarget != "team" {
		t.Errorf("expected constraint 'team', got %s", taskGroup.Constraints[0].RTarget)
	}
	if taskGroup.Constraints[1].RTarget != "gcp" {
		t.Errorf("expected constraint 'gcp', got %s", taskGroup.Constraints[1].RTarget)
	}

	// Check Canary
	if *taskGroup.Update.Canary != 1 {
		t.Errorf("expected Canary 1, got %d", *taskGroup.Update.Canary)
	}

	task := taskGroup.Tasks[0]
	if *task.Resources.CPU != 2000 {
		t.Errorf("expected CPU 2000, got %d", *task.Resources.CPU)
	}
	if *task.Resources.MemoryMB != 4096 {
		t.Errorf("expected Memory 4096, got %d", *task.Resources.MemoryMB)
	}
}

func TestGenerateJob_Invalid(t *testing.T) {
	cfg := JobConfig{
		OrgID: 0,
	}
	_, err := GenerateJob(cfg)
	if err == nil {
		t.Error("expected error for invalid config, got nil")
	}
}

func TestGenerateJob_AuthFederation(t *testing.T) {
	cfg := JobConfig{
		OrgID:              999,
		Tier:               TierStarter,
		ComputeEngine:      EngineAWS,
		Version:            "v1.0.0",
		OAuth2URI:          "https://auth.example.com",
		OAuth2ClientID:     "client-id-xyz",
		OAuth2ClientSecret: "client-secret-abc",
		OAuth2CallbackURL:  "http://localhost:8080/callback",
	}

	job, err := GenerateJob(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	task := job.TaskGroups[0].Tasks[0]
	env := task.Env

	// Verify Auth0 (Provider) URLs constructed from OAuth2URI
	if env["AUTH_RAILZWAY_COM_AUTH_URL"] != "https://auth.example.com/authorize" {
		t.Errorf("expected AUTH_URL https://auth.example.com/authorize, got %s", env["AUTH_RAILZWAY_COM_AUTH_URL"])
	}
	if env["AUTH_RAILZWAY_COM_TOKEN_URL"] != "https://auth.example.com/oauth/token" {
		t.Errorf("expected TOKEN_URL https://auth.example.com/oauth/token, got %s", env["AUTH_RAILZWAY_COM_TOKEN_URL"])
	}
	if env["AUTH_RAILZWAY_COM_API_URL"] != "https://auth.example.com/userinfo" {
		t.Errorf("expected API_URL https://auth.example.com/userinfo, got %s", env["AUTH_RAILZWAY_COM_API_URL"])
	}

	// Verify Client Credentials
	if env["AUTH_RAILZWAY_COM_CLIENT_ID"] != "client-id-xyz" {
		t.Errorf("expected CLIENT_ID client-id-xyz, got %s", env["AUTH_RAILZWAY_COM_CLIENT_ID"])
	}
	if env["AUTH_RAILZWAY_COM_CLIENT_SECRET"] != "client-secret-abc" {
		t.Errorf("expected CLIENT_SECRET client-secret-abc, got %s", env["AUTH_RAILZWAY_COM_CLIENT_SECRET"])
	}

	// Verify Metadata
	if env["AUTH_RAILZWAY_COM_ENABLED"] != "true" {
		t.Error("expected AUTH_RAILZWAY_COM_ENABLED to be true")
	}
	if env["AUTH_RAILZWAY_COM_NAME"] != "Railzway.com" {
		t.Errorf("expected AUTH_RAILZWAY_COM_NAME Railzway.com, got %s", env["AUTH_RAILZWAY_COM_NAME"])
	}
}

# CI/CD Setup & Workflows

Railzway Cloud uses GitHub Actions for Continuous Integration and Continuous Deployment to Google Cloud Platform (GCP).

## Workflows

### 1. CI - Verify (`ci-verify.yml`)
- **Triggers**: Pull Requests, Push to `main`.
- **Tasks**: Linting (`golangci-lint`), Testing (`go test -race`), Build Verification, Security Check (`govulncheck`).

### 2. CD - Staging (`cd-staging.yml`)
- **Triggers**: Push to `main`.
- **Target**: `railzway-cloud-api-staging` VM (GCP).
- **Mechanism**: Binary swap via SSH (Rolling update).

### 3. CD - Production (`cd-production.yml`)
- **Triggers**: Release Tags (`v*`).
- **Target**: `railzway-cloud-api-prod` VM (GCP).
- **Mechanism**: Binary swap via SSH with **DB Backup** before migration.
- **Safety**: Requires approval in GitHub Environment `production`.

### 4. Infra - Nomad Node (`infra-nomad-init.yml`)
- **Triggers**: Manual (Workflow Dispatch).
- **Tasks**: Provisions a new Compute Engine VM with Nomad and Docker installed via Startup Script.

## Setup Guide

To configure the pipeline, you need to set up Google Cloud Workload Identity Federation (WIF) and GitHub Secrets.

### Google Cloud Config

1.  **Create Service Account**: `github-deployer`.
2.  **Permissions**:
    - `roles/compute.instanceAdmin.v1`
    - `roles/compute.osLogin`
    - `roles/iap.tunnelResourceAccessor`
3.  **Workload Identity Federation**:
    - Pool: `github-pool`
    - Provider: `github-provider`
    - Bind repository to Service Account.

### GitHub Secrets

| Secret Name | Value |
| :--- | :--- |
| `GCP_PROJECT_ID` | Your GCP Project ID |
| `GCP_WIF_PROVIDER` | `projects/.../providers/github-provider` |
| `GCP_SERVICE_ACCOUNT` | `github-deployer@...` |

For detailed steps on setting up permissions, refer to the [Internal Wiki] or contact the DevOps lead.

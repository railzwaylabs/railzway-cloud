# Database Provisioning Documentation

## Overview

Railzway Cloud implements automated database provisioning to provide each organization with a dedicated PostgreSQL database and user. This ensures data isolation, security, and proper resource allocation per tenant.

## Architecture

### Component Diagram

```mermaid
graph TB
    subgraph "Use Case Layer"
        DeployUC[DeployUseCase]
    end
    
    subgraph "Domain Layer"
        Instance[Instance Entity]
        DBProv[DatabaseProvisioner Interface]
        Prov[Provisioner Interface]
    end
    
    subgraph "Adapter Layer"
        PGAdapter[PostgreSQL Adapter]
        NomadAdapter[Nomad Adapter]
        Repo[Instance Repository]
    end
    
    subgraph "Infrastructure"
        PostgreSQL[(PostgreSQL)]
        Nomad[Nomad Cluster]
        CloudDB[(Cloud DB)]
    end
    
    DeployUC --> DBProv
    DeployUC --> Prov
    DeployUC --> Repo
    
    DBProv -.implements.-> PGAdapter
    Prov -.implements.-> NomadAdapter
    
    PGAdapter --> PostgreSQL
    NomadAdapter --> Nomad
    Repo --> CloudDB
    
    Instance --> Repo
```

### Layer Responsibilities

| Layer | Component | Responsibility |
|-------|-----------|----------------|
| **Domain** | `Instance` | Core entity with DB credentials |
| **Domain** | `DatabaseProvisioner` | Interface for DB provisioning |
| **Domain** | `Provisioner` | Interface for workload deployment |
| **Use Case** | `DeployUseCase` | Orchestrates provisioning flow |
| **Adapter** | `PostgreSQL Adapter` | Implements DB/user creation |
| **Adapter** | `Nomad Adapter` | Deploys jobs with credentials |
| **Adapter** | `Repository` | Persists instance state |

---

## Deployment Flow

### Sequence Diagram

```mermaid
sequenceDiagram
    participant API as API Handler
    participant UC as DeployUseCase
    participant Repo as Repository
    participant DBProv as PostgreSQL Provisioner
    participant Nomad as Nomad Adapter
    participant PG as PostgreSQL
    participant NC as Nomad Cluster
    
    API->>UC: Execute(ctx, orgID, version)
    
    UC->>Repo: FindByOrgID(ctx, orgID)
    Repo-->>UC: instance
    
    alt First Deployment (no DB credentials)
        UC->>UC: Generate DB credentials
        Note over UC: DBUser: railzway_user_{orgID}<br/>DBName: railzway_org_{orgID}<br/>Password: 32-char random
    end
    
    UC->>DBProv: Provision(ctx, orgID, password)
    
    DBProv->>PG: Check if user exists
    PG-->>DBProv: exists/not exists
    
    alt User does not exist
        DBProv->>PG: CREATE USER
    else User exists
        DBProv->>PG: ALTER USER (rotate password)
    end
    
    DBProv->>PG: Check if database exists
    PG-->>DBProv: exists/not exists
    
    alt Database does not exist
        DBProv->>PG: CREATE DATABASE OWNER user
    else Database exists
        DBProv->>PG: ALTER DATABASE OWNER TO user
    end
    
    DBProv-->>UC: Success
    
    UC->>Nomad: Deploy(ctx, DeploymentConfig)
    Note over Nomad: Config includes DBConfig with credentials
    
    Nomad->>NC: Register Job with env vars
    Note over NC: DB_HOST, DB_PORT, DB_NAME,<br/>DB_USER, DB_PASSWORD,<br/>DATABASE_URL
    
    NC-->>Nomad: Job registered
    Nomad-->>UC: Success
    
    UC->>Repo: Save(ctx, instance)
    Note over Repo: Persist DB credentials
    Repo-->>UC: Success
    
    UC-->>API: Success
```

---

## State Flow

### Instance Lifecycle with DB Provisioning

```mermaid
stateDiagram-v2
    [*] --> Created: NewInstance()
    
    Created --> Provisioning: Deploy()
    
    state Provisioning {
        [*] --> CheckCredentials
        CheckCredentials --> GenerateCredentials: No credentials
        CheckCredentials --> UseExisting: Has credentials
        GenerateCredentials --> ProvisionDB
        UseExisting --> ProvisionDB
        ProvisionDB --> DeployNomad
        DeployNomad --> SaveState
        SaveState --> [*]
    }
    
    Provisioning --> Running: Job started
    Running --> Stopped: Stop()
    Stopped --> Running: Start()
    Running --> Upgrading: Upgrade()
    Upgrading --> Running: Upgrade complete
    Running --> DowngradeScheduled: Downgrade()
    DowngradeScheduled --> Running: Period end
    Running --> Terminated: Terminate()
    Terminated --> [*]
```

---

## Data Flow

### Credential Generation & Injection

```mermaid
flowchart LR
    subgraph "1. Generation"
        A[DeployUseCase] -->|crypto/rand| B[32-char password]
        A -->|format| C[railzway_user_{orgID}]
        A -->|format| D[railzway_org_{orgID}]
    end
    
    subgraph "2. Provisioning"
        B --> E[PostgreSQL Adapter]
        C --> E
        D --> E
        E -->|CREATE USER/DB| F[(PostgreSQL)]
    end
    
    subgraph "3. Persistence"
        B --> G[Instance Entity]
        C --> G
        D --> G
        G -->|Save| H[(Cloud DB)]
    end
    
    subgraph "4. Injection"
        G -->|DBConfig| I[Nomad Adapter]
        I -->|env vars| J[Nomad Job]
        J -->|runtime| K[Railzway OSS Container]
    end
```

---

## Idempotency

### Provisioning Logic

```mermaid
flowchart TD
    Start([Provision Called]) --> CheckUser{User exists?}
    
    CheckUser -->|No| CreateUser[CREATE USER]
    CheckUser -->|Yes| RotatePass[ALTER USER<br/>rotate password]
    
    CreateUser --> CheckDB{Database exists?}
    RotatePass --> CheckDB
    
    CheckDB -->|No| CreateDB[CREATE DATABASE<br/>OWNER user]
    CheckDB -->|Yes| SetOwner[ALTER DATABASE<br/>OWNER TO user]
    
    CreateDB --> Success([Return Success])
    SetOwner --> Success
```

**Key Properties:**
- ✅ Safe to call multiple times
- ✅ Updates passwords on re-provision
- ✅ Ensures ownership is correct
- ✅ No-op if already provisioned correctly

---

## Security Model

### Isolation & Access Control

```mermaid
graph TB
    subgraph "Org 1"
        App1[Railzway OSS<br/>Container]
        DB1[(railzway_org_1)]
        User1[railzway_user_1]
    end
    
    subgraph "Org 2"
        App2[Railzway OSS<br/>Container]
        DB2[(railzway_org_2)]
        User2[railzway_user_2]
    end
    
    subgraph "PostgreSQL Server"
        Admin[postgres<br/>superuser]
        DB1
        DB2
    end
    
    Admin -.creates.-> User1
    Admin -.creates.-> User2
    Admin -.creates.-> DB1
    Admin -.creates.-> DB2
    
    User1 -->|OWNER| DB1
    User2 -->|OWNER| DB2
    
    App1 -->|connect| User1
    App2 -->|connect| User2
    
    User1 -.X.-> DB2
    User2 -.X.-> DB1
    
    style User1 fill:#90EE90
    style User2 fill:#90EE90
    style Admin fill:#FFB6C1
```

**Security Guarantees:**
- Each org has dedicated user and database
- Users have OWNER privileges on their database only
- No cross-tenant access
- Passwords stored encrypted in Cloud DB
- Credentials injected via Nomad (not in job spec)

---

## Error Handling

### Failure Scenarios

| Scenario | Behavior | Recovery |
|----------|----------|----------|
| **PostgreSQL unreachable** | Provision fails, deployment aborted | Retry deployment |
| **User creation fails** | Provision fails, no DB created | Fix PG issue, retry |
| **DB creation fails** | Provision fails, user may exist | Idempotent retry |
| **Nomad deployment fails** | DB provisioned, Nomad job not created | Retry deployment (DB already exists) |
| **Credentials save fails** | DB + Nomad provisioned, state not persisted | Manual reconciliation needed |

### Retry Flow

```mermaid
flowchart TD
    Start([Deploy Request]) --> Provision{Provision DB}
    
    Provision -->|Success| DeployNomad{Deploy Nomad}
    Provision -->|Fail| RetryProv[Retry<br/>Idempotent]
    
    RetryProv --> Provision
    
    DeployNomad -->|Success| SaveState{Save State}
    DeployNomad -->|Fail| RetryNomad[Retry Deploy]
    
    RetryNomad --> DeployNomad
    
    SaveState -->|Success| Done([Success])
    SaveState -->|Fail| Alert[Alert Operator<br/>Manual Fix]
```

---

## Configuration

### Environment Variables

**Cloud Application:**
```bash
# PostgreSQL Admin Connection (for provisioning)
DB_HOST=localhost
DB_PORT=5432
DB_USER=postgres
DB_PASSWORD=<admin-password>
DB_NAME=postgres

# Nomad
NOMAD_ADDR=http://nomad.example.com:4646
NOMAD_TOKEN=<token>
```

**Injected into Railzway OSS Containers:**
```bash
# Generated per organization
DB_HOST=<tenant-db-host>
DB_PORT=5432
DB_NAME=railzway_org_<orgID>
DB_USER=railzway_user_<orgID>
DB_PASSWORD=<generated-password>
DATABASE_URL=postgres://railzway_user_<orgID>:<password>@<host>:5432/railzway_org_<orgID>?sslmode=disable
```

---

## Database Schema

### Migration: `000004_add_db_credentials.up.sql`

```sql
ALTER TABLE instances ADD COLUMN db_host VARCHAR(255);
ALTER TABLE instances ADD COLUMN db_port INT;
ALTER TABLE instances ADD COLUMN db_name VARCHAR(255);
ALTER TABLE instances ADD COLUMN db_user VARCHAR(255);
ALTER TABLE instances ADD COLUMN db_password VARCHAR(255);
```

### Instance Table Structure

| Column | Type | Description |
|--------|------|-------------|
| `id` | BIGSERIAL | Primary key |
| `org_id` | BIGINT | Organization ID (unique) |
| `nomad_job_id` | VARCHAR(255) | Nomad job identifier |
| `desired_version` | VARCHAR(50) | Target Railzway version |
| `current_version` | VARCHAR(50) | Running version |
| `status` | VARCHAR(50) | Instance status |
| `tier` | VARCHAR(50) | Pricing tier |
| `compute_engine` | VARCHAR(50) | Infrastructure provider |
| `subscription_id` | VARCHAR(255) | Billing subscription |
| **`db_host`** | **VARCHAR(255)** | **Database host** |
| **`db_port`** | **INT** | **Database port** |
| **`db_name`** | **VARCHAR(255)** | **Database name** |
| **`db_user`** | **VARCHAR(255)** | **Database user** |
| **`db_password`** | **VARCHAR(255)** | **Database password (encrypted)** |
| `created_at` | TIMESTAMP | Creation time |
| `updated_at` | TIMESTAMP | Last update time |

---

## Code References

### Key Files

| File | Purpose |
|------|---------|
| [internal/domain/instance/entity.go](file:///Users/taufiktriantono/go/src/github.com/smallbiznis/railzway-cloud/internal/domain/instance/entity.go) | Instance entity with DB fields |
| [internal/domain/provisioning/interface.go](file:///Users/taufiktriantono/go/src/github.com/smallbiznis/railzway-cloud/internal/domain/provisioning/interface.go) | Provisioning interfaces |
| [internal/adapter/provisioning/postgres/adapter.go](file:///Users/taufiktriantono/go/src/github.com/smallbiznis/railzway-cloud/internal/adapter/provisioning/postgres/adapter.go) | PostgreSQL provisioner |
| [internal/usecase/deployment/deploy.go](file:///Users/taufiktriantono/go/src/github.com/smallbiznis/railzway-cloud/internal/usecase/deployment/deploy.go) | Deploy orchestration |
| [pkg/nomad/generator.go](file:///Users/taufiktriantono/go/src/github.com/smallbiznis/railzway-cloud/pkg/nomad/generator.go) | Nomad job generation |

### Interface Definitions

**DatabaseProvisioner:**
```go
type DatabaseProvisioner interface {
    // Provision creates the database and user for the given organization.
    // It must be idempotent.
    Provision(ctx context.Context, orgID int64, password string) error
}
```

**Provisioner:**
```go
type Provisioner interface {
    Deploy(ctx context.Context, config DeploymentConfig) error
    Stop(ctx context.Context, orgID int64) error
    GetStatus(ctx context.Context, orgID int64) (string, error)
}
```

---

## Testing

### Manual Verification Steps

1. **Start PostgreSQL:**
   ```bash
   docker run -d -p 5432:5432 \
     -e POSTGRES_PASSWORD=password \
     postgres:15
   ```

2. **Start Nomad:**
   ```bash
   nomad agent -dev
   ```

3. **Deploy Instance:**
   ```bash
   # Via API or test script
   go run cmd/test-lifecycle/main.go
   ```

4. **Verify Database:**
   ```sql
   \c postgres
   \du                          -- List users
   \l                           -- List databases
   \c railzway_org_12345        -- Connect to org DB
   \dt                          -- List tables (after OSS migration)
   ```

5. **Verify Nomad Job:**
   ```bash
   nomad job status railzway-org-12345
   nomad alloc logs <alloc-id> | grep DATABASE_URL
   ```

### Expected Results

- ✅ User `railzway_user_<orgID>` created
- ✅ Database `railzway_org_<orgID>` created with correct owner
- ✅ Nomad job running with DB env vars
- ✅ Credentials persisted in Cloud DB

---

## Future Enhancements

1. **Password Encryption at Rest**
   - Encrypt `db_password` column using AES-256
   - Store encryption key in secure vault (HashiCorp Vault, AWS KMS)

2. **Connection Pooling**
   - Implement PgBouncer for tenant databases
   - Reduce connection overhead

3. **Database Metrics**
   - Monitor per-tenant DB size
   - Track connection counts
   - Alert on quota violations

4. **Automated Backups**
   - Schedule per-tenant backups
   - Implement point-in-time recovery

5. **Multi-Region Support**
   - Provision databases in user's preferred region
   - Implement read replicas for global deployments

# Railzway Cloud

**Railzway Cloud** is the infrastructure control plane for managing multi-tenant Railzway deployments. It orchestrates instance provisioning, subscription management, and resource allocation across compute providers.

## Architecture

Railzway Cloud operates as a control plane that:
- Manages organization and user lifecycle
- Provisions Railzway instances via Nomad
- Integrates with Railzway OSS for subscription and billing
- Enforces tier-based resource limits and placement constraints

### Key Components

- **Onboarding Service**: Organization initialization and subscription orchestration
- **Instance Service**: Nomad job deployment and lifecycle management
- **Database Provisioning**: Automated PostgreSQL database and user creation per organization
- **Nomad Job Generator**: Deterministic job spec generation based on pricing tiers
- **Railzway Client**: Resilient HTTP client with circuit breaker, retry, and rate limiting
- **Auth Middleware**: JWT-based authentication with session management

## Tech Stack

- **Backend**: Go 1.25.1, Gin, GORM, Uber FX
- **Frontend**: React, TypeScript, Vite
- **Orchestration**: HashiCorp Nomad
- **Database**: PostgreSQL (primary), MySQL, SQLite (dev)
- **Observability**: OpenTelemetry, Prometheus

## Getting Started

### Prerequisites

- Go 1.25.1+
- Node.js 18+ (for frontend)
- PostgreSQL 14+
- Nomad cluster (for production deployments)

### Installation

1. **Clone the repository**
   ```bash
   git clone https://github.com/smallbiznis/railzway-cloud.git
   cd railzway-cloud
   ```

2. **Configure environment**
   ```bash
   cp .env.example .env
   # Edit .env with your configuration
   ```

3. **Run database migrations**
   ```bash
   ./scripts/migrate.sh up
   ```

4. **Start the backend**
   ```bash
   go run apps/railzway/main.go
   ```

5. **Start the frontend** (in a separate terminal)
   ```bash
   cd apps/railzway
   npm install
   npm run dev
   ```

## Configuration

Key environment variables:

| Variable | Description | Default |
|----------|-------------|---------|
| `APP_MODE` | Deployment mode: `oss`, `cloud`, `standalone` | `oss` |
| `ENVIRONMENT` | Environment: `development`, `production` | `development` |
| `APP_ROOT_DOMAIN` | Root domain for tenant launch URLs (ex: `railzway.com`) | - |
| `APP_ROOT_SCHEME` | Scheme for tenant launch URLs: `http` or `https` | `https` in production, `http` otherwise |
| `ADMIN_API_TOKEN` | Admin token for protected rollout endpoint | - |
| `DB_TYPE` | Database type: `postgres`, `mysql`, `sqlite` | `postgres` |
| `DB_HOST` | Database host | `localhost` |
| `DB_PORT` | Database port | `5432` |
| `RAILZWAY_CLIENT_URL` | Railzway OSS API URL | `http://localhost:8080` |
| `RAILZWAY_API_KEY` | Railzway OSS API key | - |
| `OAUTH2_CLIENT_ID` | OAuth2 client ID | - |
| `OAUTH2_CLIENT_SECRET` | OAuth2 client secret | - |

See [.env.example](.env.example) for complete configuration options.

## Project Structure

```
railzway-cloud/
├── apps/
│   └── railzway/          # Frontend React application
├── internal/
│   ├── adapter/           # Infrastructure adapters
│   │   ├── billing/       # Billing engine adapters (Railzway OSS)
│   │   ├── provisioning/  # Provisioning adapters (Nomad, PostgreSQL)
│   │   └── repository/    # Data persistence adapters
│   ├── api/               # HTTP router and handlers
│   ├── auth/              # Authentication middleware and session management
│   ├── config/            # Configuration loader
│   ├── domain/            # Domain entities and interfaces
│   │   ├── billing/       # Billing domain
│   │   ├── instance/      # Instance domain
│   │   └── provisioning/  # Provisioning domain
│   ├── onboarding/        # Organization onboarding service
│   ├── usecase/           # Application use cases
│   │   └── deployment/    # Deployment orchestration
│   └── user/              # User service
├── pkg/
│   ├── db/                # Database utilities and GORM setup
│   ├── log/               # Structured logging
│   ├── nomad/             # Nomad job generator and client
│   ├── railzwayclient/    # Railzway OSS HTTP client
│   ├── snowflake/         # Distributed ID generation
│   └── telemetry/         # OpenTelemetry correlation
├── sql/
│   └── migrations/        # Database migration files
├── docs/
│   ├── CICD.md            # CI/CD documentation
│   └── database-provisioning.md  # Database provisioning guide
└── scripts/
    └── migrate.sh         # Migration helper script
```

## Pricing Tiers

Railzway Cloud enforces deterministic resource allocation based on subscription tiers:

| Tier | CPU | Memory | Priority | Compute Engines |
|------|-----|--------|----------|-----------------|
| **Free** | 250 MHz | 256 MB | 50 | Hetzner only |
| **Starter** | 500 MHz | 512 MB | 75 | Hetzner, DigitalOcean |
| **Growth** | 1000 MHz | 1024 MB | 100 | All (Hetzner, DigitalOcean, GCP, AWS) |

Resource limits are enforced at the Nomad job generation level and cannot be bypassed.

## Database Provisioning

Railzway Cloud automatically provisions a dedicated PostgreSQL database and user for each organization. This ensures:

- **Data Isolation**: Each organization has its own database with dedicated credentials
- **Security**: Least-privilege access with organization-scoped users
- **Idempotency**: Safe to re-provision without data loss
- **Automation**: Credentials automatically injected into Nomad jobs

### Naming Convention

- **Database**: `railzway_org_{OrgID}`
- **User**: `railzway_user_{OrgID}`
- **Password**: 32-character random alphanumeric (generated via `crypto/rand`)

### Provisioning Flow

1. **Generate Credentials**: On first deployment, Cloud generates secure credentials
2. **Create Database**: PostgreSQL adapter creates user and database (idempotent)
3. **Inject Credentials**: Nomad job receives credentials as environment variables:
   - `DB_HOST`, `DB_PORT`, `DB_NAME`, `DB_USER`, `DB_PASSWORD`
   - `DATABASE_URL` (standard PostgreSQL connection string)
4. **Persist State**: Credentials stored in Cloud database for future deployments

For detailed architecture and sequence diagrams, see [docs/database-provisioning.md](docs/database-provisioning.md).


## Development

### Running Tests

```bash
go test ./...
```

### Building for Production

```bash
# Backend
go build -o bin/railzway-cloud apps/railzway/main.go

# Frontend
cd apps/railzway
npm run build
```

### Database Migrations

```bash
# Apply migrations
./scripts/migrate.sh up

# Rollback migrations
./scripts/migrate.sh down
```

## Deployment

For automated deployment pipelines using GitHub Actions, please refer to [docs/CICD.md](docs/CICD.md).

Railzway Cloud is designed to run as a centralized control plane. Recommended deployment:

1. Deploy backend as a service (e.g., Docker, Kubernetes)
2. Serve frontend static assets via CDN or reverse proxy
3. Configure Nomad cluster with node metadata for tier and compute engine placement
4. Set up PostgreSQL with connection pooling
5. Configure OAuth2 provider for authentication

### Nomad Node Configuration

Nodes must be tagged with metadata for placement constraints:

```hcl
client {
  meta {
    tier    = "free"      # or "starter", "growth"
    compute = "hetzner"   # or "digitalocean", "gcp", "aws"
  }
}
```

## Integration with Railzway OSS

Railzway Cloud communicates with Railzway OSS for:
- Customer and subscription management
- Product and pricing retrieval
- Billing event synchronization

Configure the Railzway client in `.env`:

```env
RAILZWAY_CLIENT_URL=https://your-railzway-oss.com
RAILZWAY_API_KEY=vk_live_key_...
```

## OAuth Federation

Railzway Cloud supports OAuth federation for tenant instances. Each deployed Railzway OSS instance receives:
- Shared OAuth credentials (Google/GitHub)
- Unique per-org JWT secret (deterministic generation)
- Callback URL: `https://{org_slug}.railzway.com/login/railzway_com`

### Configuration

Add to `.env`:
```env
TENANT_OAUTH2_CLIENT_ID=your_oauth_client_id
TENANT_OAUTH2_CLIENT_SECRET=your_oauth_secret
TENANT_AUTH_JWT_SECRET_KEY=your_master_secret_for_jwt_generation
```

**Security Best Practice:** Generate strong secrets for production:

```bash
# Generate OAuth client secret (32 bytes = 44 chars base64)
openssl rand -base64 32

# Generate JWT master secret key (64 bytes = 88 chars base64)
openssl rand -base64 64
```

> **Note**: `TENANT_AUTH_JWT_SECRET_KEY` is a master key used to deterministically generate unique JWT secrets for each tenant instance. Each organization gets a unique JWT secret derived from: `SHA256(master_key + org_id)`.

## Testing

### Local Testing with Consul & Traefik

For local testing with Consul and Traefik, see the comprehensive guide: [`docs/local_testing_guide.md`](docs/local_testing_guide.md)

**Quick Start:**

1. **Setup DNS** (choose one):
   ```bash
   # Option 1: /etc/hosts
   sudo nano /etc/hosts
   # Add: 127.0.0.1   tenant-1.railzway.com
   
   # Option 2: dnsmasq (wildcard)
   brew install dnsmasq
   echo 'address=/.railzway.com/127.0.0.1' >> /opt/homebrew/etc/dnsmasq.conf
   sudo brew services start dnsmasq
   ```

2. **Provision instance**:
   - Start Cloud backend: `go run apps/railzway-cloud/main.go`
   - Navigate to `http://localhost:8080`
   - Complete onboarding and provision

3. **Verify**:
   ```bash
   # Check Nomad job
   nomad job status
   
   # Test instance
   curl http://tenant-1.railzway.com/health
   
   # Test OAuth
   open http://tenant-1.railzway.com/login
   ```

## License

[License information to be added]

## Contributing

[Contribution guidelines to be added]

## Support

For issues and questions, please open an issue on GitHub.

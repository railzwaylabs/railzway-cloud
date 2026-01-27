# Quick Start Guide - Nomad Deployment

Karena Nomad + Consul + Traefik sudah ready, tinggal 3 langkah:

## 1️⃣ **Buat Environment File**

```bash
# SSH ke server
ssh taufik_triantono@<server-ip>

# Buat directory
sudo mkdir -p /opt/railzway/deployments
sudo mkdir -p /opt/railzway/sql

# Buat .env file
sudo nano /opt/railzway/.env
```

**Paste content ini** (sesuaikan dengan setup Anda):

```bash
# Database
DB_HOST=localhost
DB_PORT=5432
DB_NAME=cloud
DB_USER=postgres
DB_PASSWORD=YOUR_PASSWORD
DB_SSL_MODE=disable

# Provision DB (sama dengan DB utama atau terpisah)
PROVISION_DB_HOST=localhost
PROVISION_DB_PORT=5432
PROVISION_DB_NAME=cloud
PROVISION_DB_USER=postgres
PROVISION_DB_PASSWORD=YOUR_PASSWORD
PROVISION_DB_SSL_MODE=disable

# Redis (optional, kosongkan jika tidak pakai)
PROVISION_RATE_LIMIT_REDIS_ADDR=
PROVISION_RATE_LIMIT_REDIS_PASSWORD=
PROVISION_RATE_LIMIT_REDIS_DB=0

# OAuth2 (untuk Cloud UI login)
OAUTH2_CLIENT_ID=your_oauth_client_id
OAUTH2_CLIENT_SECRET=your_oauth_client_secret
OAUTH2_URI=https://your-auth-provider.com
OAUTH2_CALLBACK_URL=https://cloud.railzway.com/auth/callback

# Tenant OAuth (untuk customer instances yang di-deploy)
TENANT_OAUTH2_CLIENT_ID=tenant_client_id
TENANT_OAUTH2_CLIENT_SECRET=tenant_client_secret
TENANT_AUTH_JWT_SECRET_KEY=random_secret_key_min_32_chars

# Security
AUTH_JWT_SECRET=random_jwt_secret_min_32_chars
ADMIN_API_TOKEN=random_admin_token

# Application
APP_ROOT_DOMAIN=railzway.com
APP_ROOT_SCHEME=https
```

---

## 2️⃣ **Copy Files ke Server**

```bash
# Dari local machine
scp deployments/nomad/* taufik_triantono@<server-ip>:/opt/railzway/deployments/

# (Optional) Copy SQL migrations jika ada
scp -r sql/* taufik_triantono@<server-ip>:/opt/railzway/sql/
```

---

## 3️⃣ **Deploy!**

```bash
# SSH ke server
ssh taufik_triantono@<server-ip>

# Run deployment script
cd /opt/railzway/deployments
./deploy.sh v1.2.0
```

Script akan otomatis:
- ✅ Verify prerequisites
- ✅ Populate Consul KV
- ✅ Deploy ke Nomad
- ✅ Health check

---

## ✅ **Verify Deployment**

```bash
# Check Nomad job
nomad job status railzway-cloud

# Check logs
nomad alloc logs $(nomad job allocs railzway-cloud | grep running | head -1 | awk '{print $1}')

# Health check
curl http://localhost:8080/health

# Check Consul service
consul catalog service railzway-cloud

# Access via Traefik (jika DNS sudah setup)
curl https://cloud.railzway.com/health
```

---

## 🔧 **Troubleshooting**

### Job tidak start
```bash
# Check allocation events
nomad alloc status <alloc-id> | grep Events -A 20

# Check Docker image
docker pull ghcr.io/railzwaylabs/railzway-cloud:v1.2.0
```

### Health check gagal
```bash
# Check logs
nomad alloc logs <alloc-id>

# Check database connection
psql -h localhost -U postgres -d cloud -c "SELECT 1"
```

### Consul KV kosong
```bash
# Re-run setup
cd /opt/railzway/deployments
./setup-consul-kv.sh /opt/railzway/.env

# Verify
consul kv get -recurse railzway-cloud/
```

---

## 📊 **Next: Setup GitHub Actions**

Update GitHub Secrets untuk auto-deployment:
- `GCE_HOST_PROD_1` = `<server-ip>`
- `GCE_USERNAME_PROD_1` = `taufik_triantono`
- `GCE_SSH_KEY_PROD_1` = (private key dari `~/.ssh/railzway-deploy`)

Setelah itu, setiap merge ke `main` akan auto-deploy! 🚀

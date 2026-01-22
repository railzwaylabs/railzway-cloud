# Local Testing Guide: OAuth Federation with Consul & Traefik

## Prerequisites

✅ Consul running locally
✅ Traefik running locally
✅ Nomad running locally
✅ Cloud backend configured with OAuth credentials

## Local DNS Setup

Since you have Traefik, you need to set up local DNS for org subdomains.

### Option 1: /etc/hosts (Simple)

Add entries for each test org:

```bash
sudo nano /etc/hosts
```

Add:
```
127.0.0.1   tenant-1.railzway.com
127.0.0.1   tenant-2.railzway.com
127.0.0.1   test-org.railzway.com
```

### Option 2: dnsmasq (Wildcard)

For wildcard `*.railzway.com` → `127.0.0.1`:

```bash
# Install dnsmasq
brew install dnsmasq

# Configure wildcard
echo 'address=/.railzway.com/127.0.0.1' >> /opt/homebrew/etc/dnsmasq.conf

# Start dnsmasq
sudo brew services start dnsmasq

# Configure macOS to use dnsmasq
sudo mkdir -p /etc/resolver
echo "nameserver 127.0.0.1" | sudo tee /etc/resolver/railzway.com
```

Verify:
```bash
ping tenant-1.railzway.com
# Should resolve to 127.0.0.1
```

## Environment Configuration

### 1. Update `.env` for Local Testing

```bash
# Set root domain for local testing
APP_ROOT_DOMAIN=railzway.com

# OAuth credentials (already set)
TENANT_OAUTH2_CLIENT_ID=Pqe460F4C3k8aBF6sutoUsMbOR0NZ6hT
TENANT_OAUTH2_CLIENT_SECRET=d_ZAy2ghTAbwfTFe5Vx2qzfFaMw_WVN3enPu5mV9J5Xc1t6f-4e_egyRTMCr1sjC
TENANT_AUTH_JWT_SECRET_KEY=your-master-secret-key-for-jwt-generation-change-in-production
```

### 2. OAuth App Configuration

Configure your OAuth app (Google/GitHub) with local callback:

**For Development:**
```
http://tenant-1.railzway.com/login/railzway_com
http://tenant-2.railzway.com/login/railzway_com
```

**For Production:**
```
https://*.railzway.com/login/railzway_com
```

## Testing Steps

### Step 1: Start Cloud Backend

```bash
cd /Users/taufiktriantono/go/src/github.com/smallbiznis/railzway-cloud
go run apps/railzway-cloud/main.go
```

Verify it's running:
```bash
curl http://localhost:8080/health
```

### Step 2: Provision Test Instance

#### Via UI (Recommended)
1. Navigate to `http://localhost:8080`
2. Login to Cloud backend
3. Go to onboarding page
4. Select a pricing plan (e.g., "Hobby")
5. Enter org name: `tenant-1`
6. Click "Provision Instance"

#### Via API (Alternative)
```bash
# Get session cookie first by logging in via UI
# Then:
curl -X POST http://localhost:8080/user/onboarding/initialize \
  -H "Content-Type: application/json" \
  -d '{
    "plan_id": "2013291711689658368",
    "org_name": "tenant-1"
  }' \
  --cookie "session=YOUR_SESSION_COOKIE"
```

### Step 3: Verify Nomad Job

Check if job was created:
```bash
nomad job status
```

You should see: `railzway-org-{id}`

Inspect job configuration:
```bash
# Replace {id} with actual org ID
nomad job inspect railzway-org-{id} | jq '.Job.TaskGroups[0].Tasks[0].Env' | grep OAUTH
```

Expected output:
```json
{
  "OAUTH2_CLIENT_ID": "Pqe460F4C3k8aBF6sutoUsMbOR0NZ6hT",
  "OAUTH2_CLIENT_SECRET": "d_ZAy2ghTAbwfTFe5Vx2qzfFaMw_WVN3enPu5mV9J5Xc1t6f-4e_egyRTMCr1sjC",
  "AUTH_JWT_SECRET": "<generated-hash>"
}
```

### Step 4: Verify Consul Service Registration

```bash
# List services
consul catalog services

# Check specific service
consul catalog service railzway-org-{id}
```

### Step 5: Verify Traefik Routing

Check Traefik dashboard (usually `http://localhost:8080` or `:8081`):

Look for router: `org-{id}`

**Expected Traefik Tags:**
```
traefik.enable=true
traefik.http.routers.org-{id}.rule=Host(`tenant-1.railzway.com`)
traefik.http.routers.org-{id}.entrypoints=web
```

Test routing:
```bash
curl -H "Host: tenant-1.railzway.com" http://localhost/health
```

Or directly in browser:
```
http://tenant-1.railzway.com/health
```

### Step 6: Test OAuth Flow

1. Navigate to: `http://tenant-1.railzway.com/login`
2. You should see Railzway OSS login page
3. Click "Sign in with Google" (or configured provider)
4. OAuth flow should redirect to:
   ```
   http://tenant-1.railzway.com/login/railzway_com?code=...
   ```
5. After successful auth, you should be logged in

### Step 7: Verify Database

Check that user was created in tenant database:

```bash
# Connect to tenant database
PGPASSWORD=<tenant-db-password> psql -h localhost -p 5433 -U railzway_user_{org_id} -d railzway_org_{org_id}

# Check users
SELECT id, email, name FROM users;
```

## Verification Checklist

- [ ] DNS resolves `tenant-1.railzway.com` to `127.0.0.1`
- [ ] Nomad job created with correct env vars
- [ ] Consul service registered
- [ ] Traefik router configured with correct host rule
- [ ] Instance accessible via `http://tenant-1.railzway.com`
- [ ] Login page loads
- [ ] OAuth redirect works
- [ ] User created in tenant database
- [ ] Session established after OAuth

## Troubleshooting

### Issue: DNS not resolving

**Solution:**
```bash
# Flush DNS cache
sudo dscacheutil -flushcache
sudo killall -HUP mDNSResponder

# Test resolution
nslookup tenant-1.railzway.com
```

### Issue: Traefik not routing

**Check Traefik logs:**
```bash
docker logs traefik  # if running in Docker
# or
journalctl -u traefik -f  # if systemd
```

**Verify Consul service:**
```bash
consul catalog service railzway-org-{id}
```

### Issue: OAuth callback fails

**Common causes:**
1. Callback URL mismatch in OAuth app config
2. Wrong domain in callback
3. Session cookie not set

**Debug:**
```bash
# Check instance logs
nomad logs -job railzway-org-{id}

# Look for OAuth errors
nomad logs -job railzway-org-{id} | grep -i oauth
```

### Issue: 404 on instance

**Check allocation status:**
```bash
nomad job status railzway-org-{id}
nomad alloc status <alloc-id>
```

**Check health:**
```bash
curl http://tenant-1.railzway.com/health
```

## Testing Multiple Orgs

To test multi-tenancy:

```bash
# Add more DNS entries
sudo nano /etc/hosts
```

Add:
```
127.0.0.1   tenant-2.railzway.com
127.0.0.1   tenant-3.railzway.com
```

Provision multiple instances with different org names:
- `tenant-1`
- `tenant-2`
- `tenant-3`

Each should:
- Get unique database
- Get unique JWT secret
- Be accessible via own subdomain
- Share same OAuth app

## Quick Test Script

```bash
#!/bin/bash
# test-oauth-federation.sh

ORG_SLUG="tenant-1"
DOMAIN="${ORG_SLUG}.railzway.com"

echo "Testing OAuth Federation for ${DOMAIN}"
echo "========================================"

# 1. DNS Resolution
echo -n "1. DNS Resolution: "
if ping -c 1 ${DOMAIN} &> /dev/null; then
    echo "✅ OK"
else
    echo "❌ FAILED"
    exit 1
fi

# 2. Instance Health
echo -n "2. Instance Health: "
if curl -s http://${DOMAIN}/health | grep -q "ok"; then
    echo "✅ OK"
else
    echo "❌ FAILED"
fi

# 3. Login Page
echo -n "3. Login Page: "
if curl -s http://${DOMAIN}/login | grep -q "Sign in"; then
    echo "✅ OK"
else
    echo "❌ FAILED"
fi

echo ""
echo "Manual OAuth Test:"
echo "Navigate to: http://${DOMAIN}/login"
echo "Click 'Sign in with Google' and verify OAuth flow"
```

Run:
```bash
chmod +x test-oauth-federation.sh
./test-oauth-federation.sh
```

## Next Steps

After successful local testing:

1. ✅ Verify OAuth flow works end-to-end
2. ✅ Test with multiple orgs
3. ✅ Verify JWT secrets are unique per org
4. 🔄 Deploy to staging with real domain
5. 🔄 Configure production OAuth app
6. 🔄 Test in production

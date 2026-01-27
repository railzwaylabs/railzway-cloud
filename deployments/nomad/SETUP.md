# Setup Nomad Deployment for Railzway-Cloud

## Prerequisites

1. Nomad cluster running on GCE instance
2. Consul for service discovery and KV storage
3. Traefik as ingress controller
4. Docker installed on Nomad clients

---

## Step 1: Copy Nomad Job File to Server

```bash
# From your local machine
scp deployments/nomad/railzway-cloud.nomad taufik_triantono@<GCE_IP>:/opt/railzway/deployments/
```

---

## Step 2: Store Environment Variables in Consul KV

```bash
# SSH to server
ssh taufik_triantono@<GCE_IP>

# Store env file in Consul
consul kv put railzway-cloud/env - < /opt/railzway/.env

# Verify
consul kv get railzway-cloud/env
```

**Alternative**: If you prefer file-based env, modify the Nomad job to use:
```hcl
template {
  data = <<EOH
{{ range $key, $value := secrets "railzway-cloud/env" }}
{{ $key }}={{ $value }}
{{ end }}
EOH
  destination = "secrets/file.env"
  env         = true
}
```

---

## Step 3: Update GitHub Secrets

Go to: https://github.com/railzwaylabs/railzway-cloud/settings/secrets/actions

Add/Update:
- **GCE_HOST_PROD_1**: `<your-gce-ip-address>`
- **GCE_USERNAME_PROD_1**: `taufik_triantono`
- **GCE_SSH_KEY_PROD_1**: 
  ```
  -----BEGIN OPENSSH PRIVATE KEY-----
  b3BlbnNzaC1rZXktdjEAAAAABG5vbmUAAAAEbm9uZQAAAAAAAAABAAAAMwAAAAtzc2gtZW
  ... (full private key content)
  -----END OPENSSH PRIVATE KEY-----
  ```

---

## Step 4: Initial Manual Deployment

```bash
# SSH to server
ssh taufik_triantono@<GCE_IP>

# Run Nomad job for the first time
nomad job run -var="version=v1.2.0" /opt/railzway/deployments/railzway-cloud.nomad

# Check status
nomad job status railzway-cloud

# Check allocation
nomad alloc status <alloc-id>

# View logs
nomad alloc logs <alloc-id>
```

---

## Step 5: Verify Traefik Integration

```bash
# Check if service is registered in Consul
consul catalog services | grep railzway-cloud

# Test health endpoint
curl http://localhost:8080/health

# Test via Traefik (if DNS configured)
curl https://cloud.railzway.com/health
```

---

## Step 6: Trigger Automated Deployment

After setup is complete, automated deployment will trigger on:
1. Merge PR to `main` branch
2. Semantic-release creates new version tag
3. Docker image is built and pushed to GHCR
4. GitHub Actions runs `nomad job run` via SSH
5. Nomad performs rolling update with health checks

---

## Troubleshooting

### SSH Authentication Failed
```bash
# Verify SSH key permissions
chmod 600 ~/.ssh/railzway-deploy

# Test SSH connection
ssh -i ~/.ssh/railzway-deploy taufik_triantono@<GCE_IP>
```

### Nomad Job Failed
```bash
# Check job status
nomad job status railzway-cloud

# View allocation logs
nomad alloc logs <alloc-id>

# Check events
nomad alloc status <alloc-id> | grep Events -A 10
```

### Health Check Failing
```bash
# Check if port is listening
netstat -tlnp | grep 8080

# Check Docker container
docker ps | grep railzway-cloud

# View container logs
nomad alloc logs <alloc-id>
```

### Traefik Not Routing
```bash
# Check Traefik dashboard
curl http://localhost:8081/dashboard/

# Verify service tags in Consul
consul catalog service railzway-cloud
```

---

## Rollback

```bash
# Revert to previous version
nomad job revert railzway-cloud <version-number>

# Or manually specify version
nomad job run -var="version=v1.1.0" /opt/railzway/deployments/railzway-cloud.nomad
```

---

## Monitoring

```bash
# Watch deployment progress
watch -n 2 'nomad job status railzway-cloud'

# Stream logs
nomad alloc logs -f <alloc-id>

# Check resource usage
nomad alloc status <alloc-id> | grep Resources -A 10
```

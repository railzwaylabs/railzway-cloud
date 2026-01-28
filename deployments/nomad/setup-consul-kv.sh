#!/bin/bash
# Script to populate Consul KV with railzway-cloud configuration
# Usage: ./setup-consul-kv.sh /path/to/.env

set -e

ENV_FILE="${1:-.env}"

if [ ! -f "$ENV_FILE" ]; then
  echo "Error: Environment file not found: $ENV_FILE"
  echo "Usage: $0 /path/to/.env"
  exit 1
fi

echo "📦 Populating Consul KV from $ENV_FILE..."

# Function to set KV if env var exists
set_kv() {
  local env_key="$1"
  local consul_key="$2"
  
  # Extract value from .env file
  local value=$(grep "^${env_key}=" "$ENV_FILE" | cut -d '=' -f2- | tr -d '"' | tr -d "'")
  
  if [ -n "$value" ]; then
    echo "  ✓ Setting railzway-cloud/${consul_key}"
    consul kv put "railzway-cloud/${consul_key}" "$value" > /dev/null
  else
    echo "  ⚠ Skipping railzway-cloud/${consul_key} (not found in $ENV_FILE)"
  fi
}

# Database Configuration
echo "🗄️  Database Configuration..."
set_kv "DB_HOST" "db_host"
set_kv "DB_PORT" "db_port"
set_kv "DB_NAME" "db_name"
set_kv "DB_USER" "db_user"
set_kv "DB_PASSWORD" "db_password"
set_kv "DB_SSL_MODE" "db_ssl_mode"

# Provision Database
echo "🔧 Provision Database Configuration..."
set_kv "PROVISION_DB_HOST" "provision_db_host"
set_kv "PROVISION_DB_PORT" "provision_db_port"
set_kv "PROVISION_DB_NAME" "provision_db_name"
set_kv "PROVISION_DB_USER" "provision_db_user"
set_kv "PROVISION_DB_PASSWORD" "provision_db_password"
set_kv "PROVISION_DB_SSL_MODE" "provision_db_ssl_mode"

# Redis Configuration
echo "🔴 Redis Configuration..."
set_kv "PROVISION_RATE_LIMIT_REDIS_ADDR" "redis_addr"
set_kv "PROVISION_RATE_LIMIT_REDIS_PASSWORD" "redis_password"
set_kv "PROVISION_RATE_LIMIT_REDIS_DB" "redis_db"

# Authentication
echo "🔐 Authentication Configuration..."
set_kv "AUTH_COOKIE_SECRET" "auth_cookie_secret"
set_kv "ADMIN_API_TOKEN" "admin_api_token"

# Secrets Encryption
echo "🔑 Secrets Encryption..."
set_kv "INSTANCE_SECRET_ENCRYPTION_KEY" "instance_secret_encryption_key"

# OAuth2 Configuration (Cloud UI)
echo "🔐 OAuth2 Configuration..."
set_kv "OAUTH2_CLIENT_ID" "oauth2_client_id"
set_kv "OAUTH2_CLIENT_SECRET" "oauth2_client_secret"
set_kv "OAUTH2_URI" "oauth2_uri"
set_kv "OAUTH2_CALLBACK_URL" "oauth2_callback_url"

# Auth Service (railzway-auth)
echo "🔐 Auth Service Configuration..."
set_kv "AUTH_SERVICE_URL" "auth_service_url"
set_kv "AUTH_SERVICE_TENANT" "auth_service_tenant"
set_kv "AUTH_SERVICE_TIMEOUT" "auth_service_timeout"
set_kv "AUTH_SERVICE_CLIENT_ID" "auth_service_client_id"
set_kv "AUTH_SERVICE_CLIENT_SECRET" "auth_service_client_secret"
set_kv "AUTH_SERVICE_CLIENT_SCOPE" "auth_service_client_scope"

# Nomad Configuration
echo "🚀 Nomad Configuration..."
set_kv "NOMAD_ADDR" "nomad_addr"
set_kv "NOMAD_REGION" "nomad_region"
set_kv "NOMAD_NAMESPACE" "nomad_namespace"
set_kv "NOMAD_HTTP_AUTH" "nomad_http_auth"
set_kv "NOMAD_CACERT" "nomad_cacert"
set_kv "NOMAD_CAPATH" "nomad_capath"
set_kv "NOMAD_CLIENT_CERT" "nomad_client_cert"
set_kv "NOMAD_CLIENT_KEY" "nomad_client_key"
set_kv "NOMAD_TLS_SERVER_NAME" "nomad_tls_server_name"
set_kv "NOMAD_SKIP_VERIFY" "nomad_skip_verify"
set_kv "NOMAD_TOKEN" "nomad_token"

# Railzway Client
echo "🔌 Railzway Client Configuration..."
set_kv "RAILZWAY_CLIENT_URL" "railzway_client_url"
set_kv "RAILZWAY_API_KEY" "railzway_api_key"
set_kv "RAILZWAY_CLIENT_TIMEOUT" "railzway_client_timeout"
set_kv "RAILZWAY_CLIENT_RETRY_COUNT" "railzway_client_retry_count"
set_kv "RAILZWAY_CLIENT_RETRY_DELAY" "railzway_client_retry_delay"
set_kv "RAILZWAY_CLIENT_CACHE_TTL" "railzway_client_cache_ttl"
set_kv "RAILZWAY_CLIENT_CACHE_SIZE" "railzway_client_cache_size"
set_kv "RAILZWAY_CLIENT_RATE_LIMIT" "railzway_client_rate_limit"
set_kv "RAILZWAY_CLIENT_RATE_BURST" "railzway_client_rate_burst"
set_kv "RAILZWAY_CLIENT_ENABLE_CIRCUIT_BREAKER" "railzway_client_enable_circuit_breaker"
set_kv "RAILZWAY_CLIENT_CIRCUIT_BREAKER_FAILURE_THRESHOLD" "railzway_client_circuit_breaker_failure_threshold"
set_kv "RAILZWAY_CLIENT_CIRCUIT_BREAKER_RECOVERY_TIME" "railzway_client_circuit_breaker_recovery_time"
set_kv "RAILZWAY_CLIENT_CIRCUIT_BREAKER_MIN_REQUESTS" "railzway_client_circuit_breaker_min_requests"
set_kv "RAILZWAY_CLIENT_CIRCUIT_BREAKER_SAMPLING_DURATION" "railzway_client_circuit_breaker_sampling_duration"
set_kv "RAILZWAY_CLIENT_CIRCUIT_BREAKER_HALF_OPEN_MAX_SUCCESS" "railzway_client_circuit_breaker_half_open_max_success"

# Application
echo "🌐 Application Configuration..."
set_kv "APP_ROOT_DOMAIN" "app_root_domain"
set_kv "APP_ROOT_SCHEME" "app_root_scheme"

echo ""
echo "✅ Consul KV configuration complete!"
echo ""
echo "📋 Verify with:"
echo "  consul kv get -recurse railzway-cloud/"

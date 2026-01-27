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

# OAuth2 Configuration (Cloud UI)
echo "🔐 OAuth2 Configuration..."
set_kv "OAUTH2_CLIENT_ID" "oauth2_client_id"
set_kv "OAUTH2_CLIENT_SECRET" "oauth2_client_secret"
set_kv "OAUTH2_URI" "oauth2_uri"
set_kv "OAUTH2_CALLBACK_URL" "oauth2_callback_url"

# Tenant OAuth (for deployed instances)
echo "👥 Tenant OAuth Configuration..."
set_kv "TENANT_OAUTH2_CLIENT_ID" "tenant_oauth2_client_id"
set_kv "TENANT_OAUTH2_CLIENT_SECRET" "tenant_oauth2_client_secret"
set_kv "TENANT_AUTH_JWT_SECRET_KEY" "tenant_auth_jwt_secret_key"

# Security
echo "🔒 Security Configuration..."
set_kv "AUTH_JWT_SECRET" "auth_jwt_secret"
set_kv "ADMIN_API_TOKEN" "admin_api_token"

# Application
echo "🌐 Application Configuration..."
set_kv "APP_ROOT_DOMAIN" "app_root_domain"
set_kv "APP_ROOT_SCHEME" "app_root_scheme"

echo ""
echo "✅ Consul KV configuration complete!"
echo ""
echo "📋 Verify with:"
echo "  consul kv get -recurse railzway-cloud/"

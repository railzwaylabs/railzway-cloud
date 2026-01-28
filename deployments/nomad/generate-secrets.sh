#!/bin/bash
# Script to generate secrets and update .env.production
# Usage: ./generate-secrets.sh [path/to/.env.production] [--force]
#
# By default, only replaces values that are set to "REPLACE_ME"
# Use --force to regenerate all secrets (WARNING: will invalidate existing sessions)

set -e

ENV_FILE="${1:-.env.production}"
FORCE_MODE=false

# Check for --force flag
if [[ "$1" == "--force" ]] || [[ "$2" == "--force" ]]; then
  FORCE_MODE=true
  echo "⚠️  FORCE MODE: Will regenerate ALL secrets"
  echo ""
fi

if [ ! -f "$ENV_FILE" ]; then
  echo "❌ Error: Environment file not found: $ENV_FILE"
  echo "Usage: $0 [path/to/.env.production] [--force]"
  exit 1
fi

echo "🔑 Generating secrets for $ENV_FILE..."
echo ""

# Function to check if value needs replacement
needs_replacement() {
  local key=$1
  local current_value=$(grep "^${key}=" "$ENV_FILE" | cut -d '=' -f2-)
  
  if [[ "$FORCE_MODE" == true ]]; then
    return 0  # Always replace in force mode
  fi
  
  if [[ -z "$current_value" ]] || [[ "$current_value" == "REPLACE_ME" ]]; then
    return 0  # Replace if empty or REPLACE_ME
  else
    return 1  # Don't replace if has value
  fi
}

# Backup original file
BACKUP_FILE="${ENV_FILE}.backup.$(date +%Y%m%d_%H%M%S)"
cp "$ENV_FILE" "$BACKUP_FILE"
echo "📦 Backup created: $BACKUP_FILE"
echo ""

UPDATED_COUNT=0

# Update AUTH_COOKIE_SECRET
if needs_replacement "AUTH_COOKIE_SECRET"; then
  AUTH_COOKIE_SECRET=$(openssl rand -base64 32)
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s|^AUTH_COOKIE_SECRET=.*|AUTH_COOKIE_SECRET=${AUTH_COOKIE_SECRET}|" "$ENV_FILE"
  else
    sed -i "s|^AUTH_COOKIE_SECRET=.*|AUTH_COOKIE_SECRET=${AUTH_COOKIE_SECRET}|" "$ENV_FILE"
  fi
  echo "  ✅ AUTH_COOKIE_SECRET: Generated (${AUTH_COOKIE_SECRET:0:20}...)"
  ((UPDATED_COUNT++))
else
  echo "  ⏭️  AUTH_COOKIE_SECRET: Skipped (already set)"
fi

# Update ADMIN_API_TOKEN
if needs_replacement "ADMIN_API_TOKEN"; then
  ADMIN_API_TOKEN=$(openssl rand -hex 32)
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s|^ADMIN_API_TOKEN=.*|ADMIN_API_TOKEN=${ADMIN_API_TOKEN}|" "$ENV_FILE"
  else
    sed -i "s|^ADMIN_API_TOKEN=.*|ADMIN_API_TOKEN=${ADMIN_API_TOKEN}|" "$ENV_FILE"
  fi
  echo "  ✅ ADMIN_API_TOKEN: Generated (${ADMIN_API_TOKEN:0:20}...)"
  ((UPDATED_COUNT++))
else
  echo "  ⏭️  ADMIN_API_TOKEN: Skipped (already set)"
fi

# Update INSTANCE_SECRET_ENCRYPTION_KEY
if needs_replacement "INSTANCE_SECRET_ENCRYPTION_KEY"; then
  INSTANCE_SECRET_ENCRYPTION_KEY=$(openssl rand -base64 32)
  if [[ "$OSTYPE" == "darwin"* ]]; then
    sed -i '' "s|^INSTANCE_SECRET_ENCRYPTION_KEY=.*|INSTANCE_SECRET_ENCRYPTION_KEY=${INSTANCE_SECRET_ENCRYPTION_KEY}|" "$ENV_FILE"
  else
    sed -i "s|^INSTANCE_SECRET_ENCRYPTION_KEY=.*|INSTANCE_SECRET_ENCRYPTION_KEY=${INSTANCE_SECRET_ENCRYPTION_KEY}|" "$ENV_FILE"
  fi
  echo "  ✅ INSTANCE_SECRET_ENCRYPTION_KEY: Generated (${INSTANCE_SECRET_ENCRYPTION_KEY:0:20}...)"
  ((UPDATED_COUNT++))
else
  echo "  ⏭️  INSTANCE_SECRET_ENCRYPTION_KEY: Skipped (already set)"
fi

echo ""
if [ $UPDATED_COUNT -eq 0 ]; then
  echo "✅ No secrets needed updating (all already set)"
  echo ""
  echo "💡 To force regenerate all secrets, use:"
  echo "   $0 $ENV_FILE --force"
else
  echo "✅ Updated $UPDATED_COUNT secret(s) in $ENV_FILE"
fi
echo ""
echo "🔍 Verify with:"
echo "  grep -E '(AUTH_COOKIE_SECRET|ADMIN_API_TOKEN|INSTANCE_SECRET_ENCRYPTION_KEY)' $ENV_FILE"

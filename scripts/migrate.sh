#!/bin/bash

# Load environment variables
if [ -f .env ]; then
  echo "Loading .env file..."
  set -a
  source <(grep -v '^#' .env | grep -v '^$' | sed 's/#.*$//')
  set +a
else
  echo ".env file not found in $(pwd)"
fi

# Fallback: Explicitly grab DB_SOURCE if not set (sometimes sourcing fails on some shells)
if [ -z "$DB_SOURCE" ]; then
  if [ -f .env ]; then
    DB_SOURCE=$(grep "^DB_SOURCE=" .env | cut -d '=' -f2-)
  fi
fi

# Fallback: Explicitly grab DB_* if not set (sometimes sourcing fails on some shells)
get_env_value() {
  local key="$1"
  if [ -f .env ]; then
    grep -E "^${key}=" .env | tail -n 1 | sed 's/#.*$//' | cut -d '=' -f2-
  fi
}

if [ -z "$DB_TYPE" ]; then
  DB_TYPE=$(get_env_value "DB_TYPE")
fi
if [ -z "$DB_HOST" ]; then
  DB_HOST=$(get_env_value "DB_HOST")
fi
if [ -z "$DB_PORT" ]; then
  DB_PORT=$(get_env_value "DB_PORT")
fi
if [ -z "$DB_NAME" ]; then
  DB_NAME=$(get_env_value "DB_NAME")
fi
if [ -z "$DB_USER" ]; then
  DB_USER=$(get_env_value "DB_USER")
fi
if [ -z "$DB_PASSWORD" ]; then
  DB_PASSWORD=$(get_env_value "DB_PASSWORD")
fi
if [ -z "$DB_SSL_MODE" ]; then
  DB_SSL_MODE=$(get_env_value "DB_SSL_MODE")
fi

# Check if migrate tool is installed
if ! command -v migrate &> /dev/null; then
    echo "migrate tool is not installed. Please install it:"
    echo "brew install golang-migrate"
    exit 1
fi

DB_URL=${DB_SOURCE}
if [ -z "$DB_URL" ]; then
  if [ -z "$DB_TYPE" ]; then
    DB_TYPE="postgres"
  fi

  if [ -n "$DB_HOST" ] && [ -n "$DB_PORT" ] && [ -n "$DB_NAME" ] && [ -n "$DB_USER" ]; then
    if [ -n "$DB_PASSWORD" ]; then
      DB_URL="${DB_TYPE}://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
    else
      DB_URL="${DB_TYPE}://${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
    fi

    if [ -n "$DB_SSL_MODE" ]; then
      DB_URL="${DB_URL}?sslmode=${DB_SSL_MODE}"
    fi
  fi
fi

# Check if DB_URL is set
if [ -z "$DB_URL" ]; then
    echo "DB_URL is not set. Add DB_SOURCE or DB_* (DB_TYPE, DB_HOST, DB_PORT, DB_NAME, DB_USER) in .env"
    exit 1
fi

echo "Using database: $DB_URL"

# Run migration
# Usage: ./scripts/migrate.sh [up|down|force|version] [count]
COMMAND=$1
ARG=$2

if [ -z "$COMMAND" ]; then
    echo "Usage: $0 [up|down|create|force|version] [arg]"
    exit 1
fi

MIGRATION_PATH="sql/migrations"

if [ "$COMMAND" == "create" ]; then
    if [ -z "$ARG" ]; then
        echo "Usage: $0 create [migration_name]"
        exit 1
    fi
    migrate create -ext sql -dir $MIGRATION_PATH -seq $ARG
else
    if [ -n "$ARG" ]; then
        migrate -path $MIGRATION_PATH -database "$DB_URL" $COMMAND $ARG
    else
        migrate -path $MIGRATION_PATH -database "$DB_URL" $COMMAND
    fi
fi

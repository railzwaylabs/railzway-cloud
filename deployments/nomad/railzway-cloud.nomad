job "railzway-cloud" {
  datacenters = ["dc1"]
  type        = "service"
  
  group "app" {
    count = 1

    network {
      port "http" {
        static = 8080
      }
    }

    service {
      name = "railzway-cloud"
      port = "http"
      
      tags = [
        "traefik.enable=true",
        "traefik.http.routers.railzway-cloud.rule=Host(`cloud.railzway.com`)",
        "traefik.http.routers.railzway-cloud.entrypoints=websecure",
        "traefik.http.routers.railzway-cloud.tls.certresolver=letsencrypt",
      ]

      check {
        type     = "http"
        path     = "/health"
        interval = "10s"
        timeout  = "2s"
      }
    }

    task "server" {
      driver = "docker"

      config {
        image = "ghcr.io/railzwaylabs/railzway-cloud:${NOMAD_META_version}"
        ports = ["http"]
        
        volumes = [
          "/opt/railzway/sql:/app/sql:ro"
        ]
      }

      # Environment variables (non-sensitive)
      env {
        PORT                      = "${NOMAD_PORT_http}"
        ENVIRONMENT               = "production"
        APP_SERVICE               = "railzway-cloud"
        APP_VERSION               = "${NOMAD_META_version}"
        RAILZWAY_OSS_VERSION      = "v1.6.0"
        DB_TYPE                   = "postgres"
        STATIC_DIR                = "/app/apps/railzway/dist"
      }

      # Sensitive environment variables from Consul KV
      template {
        data = <<EOH
# Database Configuration
DB_HOST={{ key "railzway-cloud/db_host" }}
DB_PORT={{ key "railzway-cloud/db_port" }}
DB_NAME={{ key "railzway-cloud/db_name" }}
DB_USER={{ key "railzway-cloud/db_user" }}
DB_PASSWORD={{ key "railzway-cloud/db_password" }}
DB_SSL_MODE={{ key "railzway-cloud/db_ssl_mode" }}

# Provision Database (for customer instances)
PROVISION_DB_HOST={{ key "railzway-cloud/provision_db_host" }}
PROVISION_DB_PORT={{ key "railzway-cloud/provision_db_port" }}
PROVISION_DB_NAME={{ key "railzway-cloud/provision_db_name" }}
PROVISION_DB_USER={{ key "railzway-cloud/provision_db_user" }}
PROVISION_DB_PASSWORD={{ key "railzway-cloud/provision_db_password" }}
PROVISION_DB_SSL_MODE={{ key "railzway-cloud/provision_db_ssl_mode" }}

# Rate Limiting Redis
PROVISION_RATE_LIMIT_REDIS_ADDR={{ key "railzway-cloud/redis_addr" }}
PROVISION_RATE_LIMIT_REDIS_PASSWORD={{ key "railzway-cloud/redis_password" }}
PROVISION_RATE_LIMIT_REDIS_DB={{ key "railzway-cloud/redis_db" }}

# OAuth2 Configuration (for Cloud UI)
OAUTH2_CLIENT_ID={{ key "railzway-cloud/oauth2_client_id" }}
OAUTH2_CLIENT_SECRET={{ key "railzway-cloud/oauth2_client_secret" }}
OAUTH2_URI={{ key "railzway-cloud/oauth2_uri" }}
OAUTH2_CALLBACK_URL={{ key "railzway-cloud/oauth2_callback_url" }}

# Tenant OAuth (for deployed OSS instances)
TENANT_OAUTH2_CLIENT_ID={{ key "railzway-cloud/tenant_oauth2_client_id" }}
TENANT_OAUTH2_CLIENT_SECRET={{ key "railzway-cloud/tenant_oauth2_client_secret" }}
TENANT_AUTH_JWT_SECRET_KEY={{ key "railzway-cloud/tenant_auth_jwt_secret_key" }}

# Security
AUTH_JWT_SECRET={{ key "railzway-cloud/auth_jwt_secret" }}
ADMIN_API_TOKEN={{ key "railzway-cloud/admin_api_token" }}

# Application
APP_ROOT_DOMAIN={{ key "railzway-cloud/app_root_domain" }}
APP_ROOT_SCHEME={{ key "railzway-cloud/app_root_scheme" }}
EOH
        destination = "secrets/file.env"
        env         = true
      }

      resources {
        cpu    = 500  # MHz
        memory = 512  # MB
      }
    }

    # Rolling update strategy
    update {
      max_parallel     = 1
      min_healthy_time = "10s"
      healthy_deadline = "3m"
      auto_revert      = true
    }
  }
}

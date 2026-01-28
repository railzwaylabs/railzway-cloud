variable "version" {
  type        = string
  description = "Docker image tag for railzway-cloud"
}

variable "github_token" {
  type        = string
  description = "GitHub PAT for pulling images from GHCR"
}

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
        
        # HTTPS Router (main)
        "traefik.http.routers.railzway-cloud.rule=Host(`cloud.railzway.com`)",
        "traefik.http.routers.railzway-cloud.entrypoints=websecure",
        "traefik.http.routers.railzway-cloud.tls.certresolver=letsencrypt",
        
        # HTTP Router (redirect to HTTPS)
        "traefik.http.routers.railzway-cloud-http.rule=Host(`cloud.railzway.com`)",
        "traefik.http.routers.railzway-cloud-http.entrypoints=web",
        "traefik.http.routers.railzway-cloud-http.middlewares=railzway-cloud-redirect",
        
        # Redirect middleware
        "traefik.http.middlewares.railzway-cloud-redirect.redirectscheme.scheme=https",
        "traefik.http.middlewares.railzway-cloud-redirect.redirectscheme.permanent=true",
      ]

      check {
        type     = "http"
        path     = "/health"
        interval = "10s"
        timeout  = "2s"
      }
    }

    # Database Migration Task (runs before app starts)
    task "migrate" {
      lifecycle {
        hook = "prestart"
        sidecar = false
      }

      driver = "docker"

      config {
        image = "ghcr.io/railzwaylabs/railzway-cloud:${var.version}"
        args    = ["migrate", "up"]

        # Docker registry authentication
        auth {
          username = "railzwaylabs"
          password = var.github_token
        }
      }

      # Environment variables (DB credentials needed for migration)
      template {
        data = <<EOH
# Database Configuration
DB_HOST={{ key "railzway-cloud/db_host" }}
DB_PORT={{ key "railzway-cloud/db_port" }}
DB_NAME={{ key "railzway-cloud/db_name" }}
DB_USER={{ key "railzway-cloud/db_user" }}
DB_PASSWORD={{ key "railzway-cloud/db_password" }}
DB_SSL_MODE={{ key "railzway-cloud/db_ssl_mode" }}
EOH
        destination = "secrets/file.env"
        env         = true
      }

      resources {
        cpu    = 200
        memory = 128
      }
    }

    # Application Server Task
    task "server" {
      driver = "docker"

      config {
        image = "ghcr.io/railzwaylabs/railzway-cloud:${var.version}"
        ports = ["http"]
        args  = ["serve"]
        
        # Docker registry authentication for GHCR
        auth {
          username = "railzwaylabs"
          password = var.github_token
        }
      }

      # Environment variables (non-sensitive)
      env {
        PORT                      = "${NOMAD_PORT_http}"
        ENVIRONMENT               = "production"
        APP_SERVICE               = "railzway-cloud"
        APP_VERSION               = "${var.version}"
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
PROVISION_RATE_LIMIT_REDIS_PASSWORD={{ keyOrDefault "railzway-cloud/redis_password" "" }}
PROVISION_RATE_LIMIT_REDIS_DB={{ key "railzway-cloud/redis_db" }}

# Authentication
AUTH_COOKIE_SECRET={{ key "railzway-cloud/auth_cookie_secret" }}
ADMIN_API_TOKEN={{ key "railzway-cloud/admin_api_token" }}

# Secrets Encryption
INSTANCE_SECRET_ENCRYPTION_KEY={{ key "railzway-cloud/instance_secret_encryption_key" }}

# GitHub Container Registry
GITHUB_TOKEN={{ key "railzway-cloud/github_token" }}

# OAuth2 Configuration (for Cloud UI)
OAUTH2_CLIENT_ID={{ key "railzway-cloud/oauth2_client_id" }}
OAUTH2_CLIENT_SECRET={{ key "railzway-cloud/oauth2_client_secret" }}
OAUTH2_URI={{ key "railzway-cloud/oauth2_uri" }}
OAUTH2_CALLBACK_URL={{ key "railzway-cloud/oauth2_callback_url" }}

# Auth Service (railzway-auth)
AUTH_SERVICE_URL={{ key "railzway-cloud/auth_service_url" }}
AUTH_SERVICE_TENANT={{ key "railzway-cloud/auth_service_tenant" }}
AUTH_SERVICE_TIMEOUT={{ key "railzway-cloud/auth_service_timeout" }}
AUTH_SERVICE_CLIENT_ID={{ key "railzway-cloud/auth_service_client_id" }}
AUTH_SERVICE_CLIENT_SECRET={{ key "railzway-cloud/auth_service_client_secret" }}
AUTH_SERVICE_CLIENT_SCOPE={{ key "railzway-cloud/auth_service_client_scope" }}

# Nomad Configuration
NOMAD_ADDR={{ keyOrDefault "railzway-cloud/nomad_addr" "http://nomad.service.consul:4646" }}
NOMAD_REGION={{ keyOrDefault "railzway-cloud/nomad_region" "sg" }}
NOMAD_NAMESPACE={{ keyOrDefault "railzway-cloud/nomad_namespace" "" }}
NOMAD_HTTP_AUTH={{ keyOrDefault "railzway-cloud/nomad_http_auth" "" }}
NOMAD_CACERT={{ keyOrDefault "railzway-cloud/nomad_cacert" "" }}
NOMAD_CAPATH={{ keyOrDefault "railzway-cloud/nomad_capath" "" }}
NOMAD_CLIENT_CERT={{ keyOrDefault "railzway-cloud/nomad_client_cert" "" }}
NOMAD_CLIENT_KEY={{ keyOrDefault "railzway-cloud/nomad_client_key" "" }}
NOMAD_TLS_SERVER_NAME={{ keyOrDefault "railzway-cloud/nomad_tls_server_name" "" }}
NOMAD_SKIP_VERIFY={{ keyOrDefault "railzway-cloud/nomad_skip_verify" "false" }}
NOMAD_TOKEN={{ keyOrDefault "railzway-cloud/nomad_token" "" }}

# Railzway Client
RAILZWAY_CLIENT_URL={{ key "railzway-cloud/railzway_client_url" }}
RAILZWAY_API_KEY={{ key "railzway-cloud/railzway_api_key" }}
RAILZWAY_CLIENT_TIMEOUT={{ keyOrDefault "railzway-cloud/railzway_client_timeout" "30" }}
RAILZWAY_CLIENT_RETRY_COUNT={{ keyOrDefault "railzway-cloud/railzway_client_retry_count" "3" }}
RAILZWAY_CLIENT_RETRY_DELAY={{ keyOrDefault "railzway-cloud/railzway_client_retry_delay" "2" }}
RAILZWAY_CLIENT_CACHE_TTL={{ keyOrDefault "railzway-cloud/railzway_client_cache_ttl" "300" }}
RAILZWAY_CLIENT_CACHE_SIZE={{ keyOrDefault "railzway-cloud/railzway_client_cache_size" "1000" }}
RAILZWAY_CLIENT_RATE_LIMIT={{ keyOrDefault "railzway-cloud/railzway_client_rate_limit" "100" }}
RAILZWAY_CLIENT_RATE_BURST={{ keyOrDefault "railzway-cloud/railzway_client_rate_burst" "2" }}
RAILZWAY_CLIENT_ENABLE_CIRCUIT_BREAKER={{ keyOrDefault "railzway-cloud/railzway_client_enable_circuit_breaker" "true" }}
RAILZWAY_CLIENT_CIRCUIT_BREAKER_FAILURE_THRESHOLD={{ keyOrDefault "railzway-cloud/railzway_client_circuit_breaker_failure_threshold" "5" }}
RAILZWAY_CLIENT_CIRCUIT_BREAKER_RECOVERY_TIME={{ keyOrDefault "railzway-cloud/railzway_client_circuit_breaker_recovery_time" "60" }}
RAILZWAY_CLIENT_CIRCUIT_BREAKER_MIN_REQUESTS={{ keyOrDefault "railzway-cloud/railzway_client_circuit_breaker_min_requests" "10" }}
RAILZWAY_CLIENT_CIRCUIT_BREAKER_SAMPLING_DURATION={{ keyOrDefault "railzway-cloud/railzway_client_circuit_breaker_sampling_duration" "60" }}
RAILZWAY_CLIENT_CIRCUIT_BREAKER_HALF_OPEN_MAX_SUCCESS={{ keyOrDefault "railzway-cloud/railzway_client_circuit_breaker_half_open_max_success" "3" }}

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

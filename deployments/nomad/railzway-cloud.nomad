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

      env {
        PORT = "${NOMAD_PORT_http}"
      }

      # Load environment variables from file
      template {
        data = <<EOH
{{ key "railzway-cloud/env" }}
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

job "traefik" {
  datacenters = ["dc1"]
  type = "system"

  group "traefik" {

    volume "traefik_certs" {
      type      = "host"
      source    = "traefik_certs"
      read_only = false
    }

    network {
      mode = "host"
      port "http"  { static = 80  }
      port "https" { static = 443 }
    }

    task "traefik" {
      driver = "docker"

      config {
        image = "traefik:v3.0"
        ports = ["http", "https"]

        args = [
          "--entrypoints.web.address=:80",
          "--entrypoints.websecure.address=:443",

          "--providers.consulcatalog.endpoint.address=127.0.0.1:8500",
          "--providers.consulcatalog.exposedbydefault=false",

          "--certificatesresolvers.cloudflare.acme.email=admin@railzway.com",
          "--certificatesresolvers.cloudflare.acme.storage=/certs/acme.json",
          "--certificatesresolvers.cloudflare.acme.dnschallenge.provider=cloudflare",
          "--certificatesresolvers.cloudflare.acme.dnschallenge.delaybeforecheck=10",

          "--log.level=INFO"
        ]
      }

      env {
        # WARNING: Please rotate this token and move to HashiCorp Vault / Nomad Variables for security!
        CF_DNS_API_TOKEN = "5Tscsxt9j-d5mgEUHbcsqsolREl9fHTCvxR2JEDm"
      }

      volume_mount {
        volume      = "traefik_certs"
        destination = "/certs"
        read_only   = false
      }

      service {
        name = "traefik"
        port = "https"
        check {
          name     = "alive"
          type     = "tcp"
          port     = "https"
          interval = "10s"
          timeout  = "2s"
        }
      }

      resources {
        cpu    = 300
        memory = 256
      }
    }
  }
}

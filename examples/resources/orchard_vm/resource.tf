provider "orchard" {
  token    = "your-service-account-token"
  base_url = "http://localhost:8080"
}

resource "orchard_vm" "example" {
  name      = "dev-box"
  image     = "ghcr.io/openai/orchard/images/ubuntu-24.04:latest"
  cpu       = 2
  memory    = 4096
  disk_size = 20
  headless  = true

  labels = {
    tier = "dev"
  }
}

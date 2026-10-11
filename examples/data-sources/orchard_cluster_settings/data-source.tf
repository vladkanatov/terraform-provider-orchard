provider "orchard" {
  token    = "your-service-account-token"
  base_url = "http://localhost:8080"
}

data "orchard_cluster_settings" "current" {}

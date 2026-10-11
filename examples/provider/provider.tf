terraform {
  required_providers {
    orchard = {
      source = "vladkanatov/orchard"
    }
  }
}

provider "orchard" {
  token    = "your-service-account-token"
  base_url = "http://localhost:8080"
}

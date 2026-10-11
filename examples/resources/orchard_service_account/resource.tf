provider "orchard" {
  token    = "your-service-account-token"
  base_url = "http://localhost:8080"
}

resource "orchard_service_account" "example" {
  name  = "ci-bot"
  roles = ["admin"]
}

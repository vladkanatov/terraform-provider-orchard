provider "orchard" {
  token    = "your-service-account-token"
  base_url = "http://localhost:8080"
}

data "orchard_controller_info" "this" {}

output "orchard_version" {
  value = data.orchard_controller_info.this.version
}

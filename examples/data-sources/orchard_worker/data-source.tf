provider "orchard" {
  token    = "your-service-account-token"
  base_url = "http://localhost:8080"
}

data "orchard_worker" "all" {}

output "workers" {
  value = data.orchard_worker.all.workers[*].name
}

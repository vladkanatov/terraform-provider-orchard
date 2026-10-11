provider "orchard" {
  token    = "your-service-account-token"
  base_url = "http://localhost:8080"
}

data "orchard_service_account" "all" {}

output "account_names" {
  value = data.orchard_service_account.all.service_accounts[*].name
}

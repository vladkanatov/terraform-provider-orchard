provider "orchard" {
  token    = "your-service-account-token"
  base_url = "http://localhost:8080"
}

# A single VM by name
data "orchard_vm" "dev" {
  name = "dev-box"
}

# All VMs on a specific worker
data "orchard_vm" "on_worker1" {
  filter = "worker=worker1"
}

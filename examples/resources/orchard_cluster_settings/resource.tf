provider "orchard" {
  token    = "your-service-account-token"
  base_url = "http://localhost:8080"
}

resource "orchard_cluster_settings" "default" {
  scheduler_profile = "distribute-load"

  host_dir_policies = [
    {
      path_prefix = "/Users/shared"
      ro          = true
    }
  ]
}

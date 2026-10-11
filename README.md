# terraform-provider-orchard

A [Terraform](https://terraform.io) provider for [Orchard](https://github.com/openai/orchard) — a local multi-VM cluster controller.

The provider lets you manage Orchard resources declaratively:

- **Resources**: [`orchard_vm`](docs/resources/vm.md), [`orchard_service_account`](docs/resources/service_account.md), [`orchard_cluster_settings`](docs/resources/cluster_settings.md)
- **Data sources**: [`orchard_controller_info`](docs/data-sources/controller_info.md), [`orchard_cluster_settings`](docs/data-sources/cluster_settings.md), [`orchard_worker`](docs/data-sources/worker.md), [`orchard_vm`](docs/data-sources/vm.md), [`orchard_service_account`](docs/data-sources/service_account.md)

Built on the [Terraform Plugin Framework](https://developer.hashicorp.com/terraform/plugin/framework).

## Example Usage

```hcl
provider "orchard" {
  token    = "sa-token-from-orchard"
  base_url = "http://localhost:8080"
}

resource "orchard_vm" "dev" {
  name    = "dev-box"
  image   = "ghcr.io/openai/orchard/images/ubuntu-24.04:latest"
  cpu     = 2
  memory  = 4096
  headless = true
}
```

See [`examples/`](examples/) for ready-to-use configurations of every resource and data source.

## Requirements

- [Terraform](https://developer.hashicorp.com/terraform/install) >= 1.3
- [Go](https://go.dev/doc/install) >= 1.21 (to build the provider)
- A running Orchard controller (for apply/acceptance tests)

## Building

```sh
make build   # builds bin/terraform-provider-orchard
```

## Local Development

To test the provider against a local Orchard controller without publishing it, install it into Terraform's filesystem mirror:

```sh
ORCHARD_BASE_URL=http://localhost:8080 ORCHARD_TOKEN=... make install
```

Then `terraform init` in any directory that references the `orchard` provider — Terraform will find the local build.

Alternatively use `dev_overrides` in `~/.terraformrc`:

```hcl
provider_installation {
  dev_overrides {
    "vladkanatov/orchard" = "/path/to/this/repo/bin"
  }
  direct {}
}
```

## Testing

```sh
make test    # unit tests (no cluster needed)

# acceptance tests — requires a running Orchard controller
ORCHARD_BASE_URL=http://localhost:8080 ORCHARD_TOKEN=... make testacc
```

## Documentation

Provider documentation for the Terraform registry lives in [`docs/`](docs/). It is generated from schema descriptions and `examples/` by [tfplugindocs](https://github.com/hashicorp/terraform-plugin-docs):

```sh
make docs
```

## Development

Guidance for AI coding agents lives in [AGENTS.md](AGENTS.md) — read it before making changes.

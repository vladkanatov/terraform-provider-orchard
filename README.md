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

## Releasing

Releases are automated with [GoReleaser](https://goreleaser.com) (config adapted from the
official [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework))
and published to [registry.terraform.io](https://registry.terraform.io/providers/vladkanatov/orchard)
on every `v*` git tag.

One-time setup (do this before the first release):

1. The GitHub repo must be **public** and named `terraform-provider-orchard` (lowercase)
   under the GitHub account that owns the registry namespace (here: `vladkanatov`).
2. Generate a GPG signing key — **RSA or DSA** (the registry does not accept the default ECC):
   `gpg --full-generate-key` (choose RSA). Export the ASCII-armored private key:
   `gpg --armor --export-secret-keys <KEY_ID>`.
3. Add the public key to the registry:
   `gpg --armor --export <KEY_ID>` → [User Settings → Signing Keys](https://registry.terraform.io/settings/gpg-keys).
4. Add GitHub repo secrets (**Settings → Secrets**):
   - `GPG_PRIVATE_KEY` — the ASCII-armored private key from step 2
   - `PASSPHRASE` — its passphrase
5. Register the provider in the registry: sign in at
   [registry.terraform.io](https://registry.terraform.io) with the GitHub account,
   then **Publish → Provider** → select `vladkanatov` / `terraform-provider-orchard`.
   This creates a webhook — from now on every GitHub Release is ingested by the registry automatically
   (no registry token needed).

Publishing a new version:

1. Make sure CI is green.
2. `git tag v0.1.0 && git push origin v0.1.0`
3. The Release workflow builds all platforms, signs the checksums, and creates the GitHub Release;
   the registry picks it up via the webhook. Check the version at
   https://registry.terraform.io/providers/vladkanatov/orchard.

Local dry-run (no publishing):

```sh
go run github.com/goreleaser/goreleaser/v2@latest release --snapshot --clean --skip=sign
```

Releases must never be modified after publication — always cut a new version.

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

# AGENTS.md

This file provides guidance to AI coding agents (and humans) when working with code in this repository.

## Project Overview

`terraform-provider-orchard` is a Terraform provider for [Orchard](https://github.com/openai/orchard) — a local multi-VM cluster controller. It exposes Orchard's REST API as Terraform resources and data sources:

| Type | Name | Description |
|------|------|-------------|
| Resource | `orchard_vm` | Create/read/update/delete a virtual machine |
| Resource | `orchard_service_account` | Manage API service accounts |
| Resource | `orchard_cluster_settings` | Cluster-wide settings (host dir policies, scheduler profile) |
| Data source | `orchard_controller_info` | Controller version/commit/capabilities |
| Data source | `orchard_cluster_settings` | Read current cluster settings |
| Data source | `orchard_worker` | List cluster workers and their resources |
| Data source | `orchard_vm` | Read one or many VMs (by name or filter) |
| Data source | `orchard_service_account` | Read one or many service accounts |

The provider is built on **terraform-plugin-framework** v1.5.0 (NOT the legacy SDK v2). Do not mix the two.

## Repository Layout

```
main.go                        # Entry point: serves the provider (registry.terraform.io/vladkanatov/orchard)
internal/
  client/
    client.go                  # HTTP client for the Orchard API + all API request/response types
  provider/
    provider.go                # Provider: schema (token, base_url), registration of resources/data sources
    schema_helpers.go          # Shared VM schema attribute maps (resource + data source variants)
    resource_<name>.go         # One file per resource
    datasource_<name>.go       # One file per data source
docs/                          # Registry documentation (source of truth for registry.terraform.io pages)
examples/                      # Terraform configs: used by tfplugindocs and for manual testing
```

Note: a large `terraform-provider-orchard` binary may exist in the repo root — it is a build artifact (gitignored), not source.

## Commands

```bash
make build        # build provider binary to bin/
make test         # unit tests (no cluster required; acceptance tests auto-skip without TF_ACC)
make testacc ORCHARD_BASE_URL=http://localhost:8080 ORCHARD_TOKEN=...
                  # acceptance tests — requires a RUNNING Orchard controller
make install      # build + copy to local plugin mirror so `terraform` picks it up
make fmt vet tidy # formatting, static checks, dependency hygiene
make docs         # regenerate docs/ via tfplugindocs (downloads the tool on the fly)
go run . -debug   # run provider in debug mode (for terraform debug session / delve)

# release dry-run (builds all platforms into dist/, no publishing)
go run github.com/goreleaser/goreleaser@latest release --snapshot --clean
```

## Releasing

- Release pipeline: `.goreleaser.yml` + `.github/workflows/release.yml`, both adapted from the official
  [terraform-provider-scaffolding-framework](https://github.com/hashicorp/terraform-provider-scaffolding-framework).
  Pushing a tag `vX.Y.Z` creates a signed GitHub Release; the registry ingests it via webhook (no token needed).
- Registry-mandated naming — do not change: archives `terraform-provider-orchard_<version>_<os>_<arch>.zip`,
  the binary inside each zip `terraform-provider-orchard_v<version>`, checksums `terraform-provider-orchard_<version>_SHA256SUMS` (+ `.sig`),
  and `terraform-provider-orchard_<version>_manifest.json` (generated from `terraform-registry-manifest.json` at repo root).
- `terraform-registry-manifest.json` declares protocol 6.0 (plugin framework) — keep in sync with `providerserver.ServeOpts` in `main.go`.
- GPG signing key must be RSA/DSA (registry rejects ECC); CI secrets are `GPG_PRIVATE_KEY` and `PASSPHRASE`.
- Before tagging: ensure CI is green. Never modify a published release — cut a new version.

Manual end-to-end test: `make install`, then `terraform init/apply` a config in `examples/` (or any scratch dir) against a local controller.

## Architecture

- **`internal/client`** — a plain `net/http` JSON client. All endpoints live under `/v1/`. All API types (VM, VMSpec, ServiceAccount, ...) are hand-maintained mirrors of the Orchard OpenAPI spec — when the spec changes, update json tags/fields here first, then the provider layer.
- **`internal/provider`** — everything else. The provider's `Configure` builds a single `*client.Client` from `token` + `base_url` and puts it into `ResourceData`/`DataSourceData`. Each resource/data source retrieves it in its `Configure` via type assertion.
- **State models** — each resource has a `<name>Model` struct with `tfsdk` tags mapping 1:1 to Terraform schema attribute names, plus two converters:
  - `(*<name>Model).to<X>()` — Terraform state → API request type
  - `<name>ModelFromX(...)` — API response → Terraform state
- Attribute names in Terraform are `snake_case`; API fields are often camelCase — the conversion lives in the converters, never in the schema.

## Conventions

- One file per resource/data source: `resource_<name>.go`, `datasource_<name>.go`.
- Resource type name is `req.ProviderTypeName + "_<name>"` (e.g. `orchard_vm`).
- Every CRUD/Read method follows the same skeleton:

  ```go
  var plan <name>Model
  diags := req.Plan.Get(ctx, &plan)   // or req.State.Get
  resp.Diagnostics.Append(diags...)
  if resp.Diagnostics.HasError() {
      return
  }
  // call r.client.<Method>(ctx, ...)
  // on error: resp.Diagnostics.AddError("Failed to <verb> <noun>", err.Error()); return
  // success: diags = resp.State.Set(ctx, model); resp.Diagnostics.Append(diags...)
  ```

- **Drift handling**: in `Read`, a 404 (detected via `strings.Contains(err.Error(), "404")`) means the object was deleted out-of-band → `resp.State.RemoveResource(ctx)` and return without error.
- Sensitive values (tokens, passwords) must have `Sensitive: true`.
- Attributes that only the controller may set are `Computed` (optionally `Optional + Computed` when the user may provide a value but the controller fills defaults).
- Schema `Description` fields are user-facing and must be in English.
- Do not add new Go dependencies casually; the project deliberately uses only the stdlib + plugin framework.
- Comments in code: keep minimal and in English.

## Adding a New Resource

1. `internal/client`: add API types (json tags must match the OpenAPI spec) and methods (`Create<X>`, `Get<X>`, `List<X>`, `Update<X>`, `Delete<X>`).
2. `internal/provider/resource_<name>.go`: copy an existing simple resource (e.g. `resource_sa.go`) as a template — struct, `New<X>Resource`, `Metadata`, `Schema`, `Configure`, CRUD, model + converters.
3. Register in `provider.go` → `Resources()`.
4. `examples/resources/orchard_<name>/resource.tf` — a minimal working example.
5. `docs/resources/<name>.md` — follow the format of existing pages (keep in sync manually, or `make docs`).
6. Acceptance test (see below) when feasible.

Data sources follow the same steps with `datasource_*` naming and `DataSources()` registration.

## Testing

- **Unit tests**: `go test ./...` — no external deps, no TF_ACC.
- **Acceptance tests**: use `github.com/hashicorp/terraform-plugin-testing` (add as a test-only dep when writing the first one). Gate with `resource.Test` / `TF_ACC`. Read connection settings from `ORCHARD_BASE_URL` and `ORCHARD_TOKEN` env vars; run with `make testacc`. Acceptance tests hit a real controller and create real VMs/SA — keep them self-contained (create + destroy) and give resources unique names (e.g. `acctest-<random>`).
- There are currently **no tests at all** in the repo; adding the first acceptance test will establish the pattern.

## Known Issues / TODO

- `VMState.UsedResources` is tagged `json:"-"`, so it is never populated from API responses — the `used_resources` attribute is therefore always empty. Fix depends on the Orchard API actually returning the data.
- `schema_helpers.go`: `vmResourceSchemaAttributes()` is currently unused (the VM resource defines its schema inline); `vmDataSourceSchemaAttributes()` IS used by `datasource_vm.go`. Either consolidate the VM resource to use the helper or delete the dead function.
- No test coverage yet.

## Docs & Release Notes

- `docs/` is what gets published to registry.terraform.io. Keep it accurate whenever a schema changes.
- Update `CHANGELOG.md` on any user-visible change (new resource, attribute, behavior).

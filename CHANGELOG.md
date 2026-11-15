# Changelog

Notable changes to the Orchard Terraform provider.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## Unreleased

### Removed
- `used_resources` attribute from the `orchard_vm` resource and data source.
  The Orchard API does not return per-VM used resources, so the attribute
  was always null. It was `Computed` (never settable in configuration), so
  removing it does not affect any valid configuration; the attribute simply
  disappears from state on the next refresh.

### Removed
- Dead code: unused `vmResourceSchemaAttributes()` helper in `schema_helpers.go`
  (the `orchard_vm` resource defines its schema inline) and the unused
  `VMResources` client struct.

## 0.0.1

### Added
- `orchard_vm` resource and data source
- `orchard_service_account` resource and data source
- `orchard_cluster_settings` resource and data source
- `orchard_controller_info` data source
- `orchard_worker` data source

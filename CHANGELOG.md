# Changelog

Notable changes to the Orchard Terraform provider.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/).

## 0.0.2

### Removed
- `used_resources` attribute from the `orchard_vm` resource and data source.
  The Orchard API does not return per-VM used resources, so the attribute
  was always null. It was `Computed` (never settable in configuration), so
  removing it does not affect any valid configuration; the attribute simply
  disappears from state on the next refresh.
- Dead code: unused `vmResourceSchemaAttributes()` helper in `schema_helpers.go`
  (the `orchard_vm` resource defines its schema inline) and the unused
  `VMResources` client struct.

### Fixed
- `orchard_vm` resource and data source no longer fail refresh/plan/destroy
  when a VM is in a terminal state (e.g. `failed`). The API's `/ip` endpoint
  returns 417 for such VMs; the provider now treats this as "the VM has no IP"
  (`ip` becomes `null`) instead of returning an error. Other errors from the
  endpoint still fail the read as before.

## 0.0.1

### Added
- `orchard_vm` resource and data source
- `orchard_service_account` resource and data source
- `orchard_cluster_settings` resource and data source
- `orchard_controller_info` data source
- `orchard_worker` data source

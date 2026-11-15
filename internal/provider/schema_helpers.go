package provider

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vmDataSourceSchemaAttributes возвращает схему атрибутов VM для data source
func vmDataSourceSchemaAttributes() map[string]dschema.Attribute {
	return map[string]dschema.Attribute{
		"name": dschema.StringAttribute{
			Computed:    true,
			Description: "Unique name of the VM.",
		},
		"generation": dschema.Int64Attribute{
			Computed:    true,
			Description: "Incremented by the controller on each spec change.",
		},
		"os": dschema.StringAttribute{
			Computed:    true,
			Description: "Operating system (darwin or linux).",
		},
		"arch": dschema.StringAttribute{
			Computed:    true,
			Description: "Architecture (arm64 or amd64).",
		},
		"runtime": dschema.StringAttribute{
			Computed:    true,
			Description: "Runtime (tart or vetu).",
		},
		"image": dschema.StringAttribute{
			Computed:    true,
			Description: "VM image reference.",
		},
		"image_pull_policy": dschema.StringAttribute{
			Computed:    true,
			Description: "Image pull policy (IfNotPresent or Always).",
		},
		"cpu": dschema.Float64Attribute{
			Computed:    true,
			Description: "Number of CPUs.",
		},
		"memory": dschema.Float64Attribute{
			Computed:    true,
			Description: "Amount of RAM in megabytes.",
		},
		"disk_size": dschema.Float64Attribute{
			Computed:    true,
			Description: "Disk size in GB.",
		},
		"net_softnet": dschema.BoolAttribute{
			Computed:    true,
			Description: "Use Softnet network isolation.",
		},
		"net_softnet_allow": dschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: "CIDRs to allow when using Softnet.",
		},
		"net_softnet_block": dschema.ListAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: "CIDRs to block when using Softnet.",
		},
		"suspendable": dschema.BoolAttribute{
			Computed:    true,
			Description: "Enable suspend/resume support.",
		},
		"net_bridged": dschema.StringAttribute{
			Computed:    true,
			Description: "Bridged network interface (e.g., en0).",
		},
		"headless": dschema.BoolAttribute{
			Computed:    true,
			Description: "Run without graphics.",
		},
		"nested": dschema.BoolAttribute{
			Computed:    true,
			Description: "Enable nested virtualization.",
		},
		"username": dschema.StringAttribute{
			Computed:    true,
			Description: "SSH username.",
		},
		"password": dschema.StringAttribute{
			Computed:    true,
			Sensitive:   true,
			Description: "SSH password.",
		},
		"startup_script": dschema.SingleNestedAttribute{
			Computed: true,
			Attributes: map[string]dschema.Attribute{
				"script_content": dschema.StringAttribute{
					Computed: true,
				},
				"env": dschema.MapAttribute{
					ElementType: types.StringType,
					Computed:    true,
				},
			},
		},
		"restart_policy": dschema.StringAttribute{
			Computed:    true,
			Description: "Restart policy (Never or OnFailure).",
		},
		"resources": dschema.MapAttribute{
			ElementType: types.Int64Type,
			Computed:    true,
			Description: "Resource requirements for the worker.",
		},
		"labels": dschema.MapAttribute{
			ElementType: types.StringType,
			Computed:    true,
			Description: "Labels for worker selection.",
		},
		"host_dirs": dschema.ListNestedAttribute{
			Computed: true,
			NestedObject: dschema.NestedAttributeObject{
				Attributes: map[string]dschema.Attribute{
					"name": dschema.StringAttribute{
						Computed: true,
					},
					"path": dschema.StringAttribute{
						Computed: true,
					},
					"ro": dschema.BoolAttribute{
						Computed: true,
					},
				},
			},
			Description: "Host directories to mount.",
		},
		"power_state": dschema.StringAttribute{
			Computed:    true,
			Description: "Desired power state (running, stopped, suspended).",
		},
		"local_name": dschema.StringAttribute{
			Computed:    true,
			Description: "Local name of the backing VM (read-only).",
		},
		"status": dschema.StringAttribute{
			Computed:    true,
			Description: "Current status of the VM.",
		},
		"status_message": dschema.StringAttribute{
			Computed: true,
		},
		"worker": dschema.StringAttribute{
			Computed:    true,
			Description: "Worker assigned to this VM.",
		},
		"observed_generation": dschema.Int64Attribute{
			Computed: true,
		},
		"ip": dschema.StringAttribute{
			Computed:    true,
			Description: "Current IP address of the VM.",
		},
	}
}

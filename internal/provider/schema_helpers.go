package provider

import (
	dschema "github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	rschema "github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
)

// vmResourceSchemaAttributes возвращает схему атрибутов VM для ресурса (resource)
func vmResourceSchemaAttributes() map[string]rschema.Attribute {
	return map[string]rschema.Attribute{
		"name": rschema.StringAttribute{
			Required:    true,
			Description: "Unique name of the VM within the cluster.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"generation": rschema.Int64Attribute{
			Computed:    true,
			Description: "Incremented by the controller on each spec change.",
		},
		"os": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Operating system (darwin or linux). Cannot be changed after creation.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"arch": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Architecture (arm64 or amd64). Cannot be changed after creation.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"runtime": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Runtime (tart or vetu). Cannot be changed after creation.",
			PlanModifiers: []planmodifier.String{
				stringplanmodifier.RequiresReplace(),
			},
		},
		"image": rschema.StringAttribute{
			Required:    true,
			Description: "VM image reference.",
		},
		"image_pull_policy": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Image pull policy (IfNotPresent or Always).",
		},
		"cpu": rschema.Float64Attribute{
			Optional:    true,
			Computed:    true,
			Description: "Number of CPUs.",
		},
		"memory": rschema.Float64Attribute{
			Optional:    true,
			Computed:    true,
			Description: "Amount of RAM in megabytes.",
		},
		"disk_size": rschema.Float64Attribute{
			Optional:    true,
			Computed:    true,
			Description: "Disk size in GB.",
		},
		"net_softnet": rschema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Use Softnet network isolation.",
		},
		"net_softnet_allow": rschema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "CIDRs to allow when using Softnet.",
		},
		"net_softnet_block": rschema.ListAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "CIDRs to block when using Softnet.",
		},
		"suspendable": rschema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Enable suspend/resume support.",
		},
		"net_bridged": rschema.StringAttribute{
			Optional:    true,
			Description: "Bridged network interface (e.g., en0).",
		},
		"headless": rschema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Run without graphics.",
		},
		"nested": rschema.BoolAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Enable nested virtualization.",
		},
		"username": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "SSH username.",
		},
		"password": rschema.StringAttribute{
			Optional:    true,
			Sensitive:   true,
			Description: "SSH password.",
		},
		"startup_script": rschema.SingleNestedAttribute{
			Optional: true,
			Attributes: map[string]rschema.Attribute{
				"script_content": rschema.StringAttribute{
					Required: true,
				},
				"env": rschema.MapAttribute{
					ElementType: types.StringType,
					Optional:    true,
				},
			},
		},
		"restart_policy": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Restart policy (Never or OnFailure).",
		},
		"resources": rschema.MapAttribute{
			ElementType: types.Int64Type,
			Optional:    true,
			Description: "Resource requirements for the worker.",
		},
		"used_resources": rschema.MapAttribute{
			ElementType: types.Int64Type,
			Computed:    true,
			Description: "Resources currently used by the VM on the worker.",
		},
		"labels": rschema.MapAttribute{
			ElementType: types.StringType,
			Optional:    true,
			Description: "Labels for worker selection.",
		},
		"host_dirs": rschema.ListNestedAttribute{
			Optional: true,
			NestedObject: rschema.NestedAttributeObject{
				Attributes: map[string]rschema.Attribute{
					"name": rschema.StringAttribute{
						Optional: true,
					},
					"path": rschema.StringAttribute{
						Required: true,
					},
					"ro": rschema.BoolAttribute{
						Optional: true,
					},
				},
			},
			Description: "Host directories to mount.",
		},
		"power_state": rschema.StringAttribute{
			Optional:    true,
			Computed:    true,
			Description: "Desired power state (running, stopped, suspended).",
		},
		"local_name": rschema.StringAttribute{
			Computed:    true,
			Description: "Local name of the backing VM (read-only).",
		},
		"status": rschema.StringAttribute{
			Computed:    true,
			Description: "Current status of the VM.",
		},
		"status_message": rschema.StringAttribute{
			Computed: true,
		},
		"worker": rschema.StringAttribute{
			Computed:    true,
			Description: "Worker assigned to this VM.",
		},
		"observed_generation": rschema.Int64Attribute{
			Computed: true,
		},
	}
}

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
		"used_resources": dschema.MapAttribute{
			ElementType: types.Int64Type,
			Computed:    true,
			Description: "Resources currently used by the VM on the worker.",
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

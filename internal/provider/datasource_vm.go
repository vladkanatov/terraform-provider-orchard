package provider

import (
	"context"
	"fmt"
	"strings"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vladkanatov/terraform-provider-orchard/internal/client"
)

type VMDataSource struct {
	client *client.Client
}

func NewVMDataSource() datasource.DataSource {
	return &VMDataSource{}
}

func (d *VMDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}

func (d *VMDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about a specific VM or list of VMs.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the VM. If omitted, returns all VMs (optionally filtered).",
			},
			"filter": schema.StringAttribute{
				Optional:    true,
				Description: "Filter expression (e.g., 'worker=worker1').",
			},
			"vms": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: vmDataSourceSchemaAttributes(), // ИСПРАВЛЕНО: используем правильную функцию
				},
			},
		},
	}
}

func (d *VMDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	if req.ProviderData == nil {
		return
	}
	cl, ok := req.ProviderData.(*client.Client)
	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Provider Data Type",
			fmt.Sprintf("Expected *client.Client, got %T", req.ProviderData),
		)
		return
	}
	d.client = cl
}

func (d *VMDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		Name   types.String `tfsdk:"name"`
		Filter types.String `tfsdk:"filter"`
		VMs    []vmModel    `tfsdk:"vms"`
	}
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var vms []client.VM
	var err error
	if !config.Name.IsNull() {
		vm, err := d.client.GetVM(ctx, config.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read VM", err.Error())
			return
		}
		vms = []client.VM{*vm}
	} else {
		filter := config.Filter.ValueString()
		vms, err = d.client.ListVMs(ctx, filter)
		if err != nil {
			resp.Diagnostics.AddError("Failed to list VMs", err.Error())
			return
		}
	}

	vmModels := make([]vmModel, len(vms))
	for i, vm := range vms {
		vmModels[i] = vmModelFromVM(&vm)
		// 417 means the VM is in a terminal state (e.g. 'failed') and has no
		// IP — report null instead of failing the whole read.
		ip, err := d.client.GetVMIP(ctx, vm.Name)
		if err != nil {
			if !strings.Contains(err.Error(), "417") {
				resp.Diagnostics.AddError("Failed to read VM IP", err.Error())
				return
			}
			vmModels[i].IP = types.StringNull()
			continue
		}
		if ip != "" {
			vmModels[i].IP = types.StringValue(ip)
		} else {
			vmModels[i].IP = types.StringNull()
		}
	}
	state := struct {
		VMs []vmModel `tfsdk:"vms"`
	}{
		VMs: vmModels,
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

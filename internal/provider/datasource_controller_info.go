package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vladkanatov/terraform-provider-orchard/internal/client"
)

type ControllerInfoDataSource struct {
	client *client.Client
}

func NewControllerInfoDataSource() datasource.DataSource {
	return &ControllerInfoDataSource{}
}

func (d *ControllerInfoDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_controller_info"
}

func (d *ControllerInfoDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about the Orchard controller.",
		Attributes: map[string]schema.Attribute{
			"version": schema.StringAttribute{
				Computed: true,
			},
			"commit": schema.StringAttribute{
				Computed: true,
			},
			"capabilities": schema.ListAttribute{
				ElementType: types.StringType,
				Computed:    true,
			},
		},
	}
}

func (d *ControllerInfoDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ControllerInfoDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	info, err := d.client.GetControllerInfo(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read controller info", err.Error())
		return
	}

	// Исправлено: caps теперь []attr.Value
	caps := make([]attr.Value, len(info.Capabilities))
	for i, c := range info.Capabilities {
		caps[i] = types.StringValue(c)
	}
	state := struct {
		Version      types.String `tfsdk:"version"`
		Commit       types.String `tfsdk:"commit"`
		Capabilities types.List   `tfsdk:"capabilities"`
	}{
		Version:      types.StringValue(info.Version),
		Commit:       types.StringValue(info.Commit),
		Capabilities: types.ListValueMust(types.StringType, caps),
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

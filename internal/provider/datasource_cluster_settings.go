package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/vladkanatov/terraform-provider-orchard/internal/client"
)

type ClusterSettingsDataSource struct {
	client *client.Client
}

func NewClusterSettingsDataSource() datasource.DataSource {
	return &ClusterSettingsDataSource{}
}

func (d *ClusterSettingsDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_settings"
}

func (d *ClusterSettingsDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves current cluster settings.",
		Attributes: map[string]schema.Attribute{
			"host_dir_policies": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path_prefix": schema.StringAttribute{Computed: true},
						"ro":          schema.BoolAttribute{Computed: true},
					},
				},
			},
			"scheduler_profile": schema.StringAttribute{
				Computed: true,
			},
		},
	}
}

func (d *ClusterSettingsDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ClusterSettingsDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	settings, err := d.client.GetClusterSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read cluster settings", err.Error())
		return
	}

	state := clusterSettingsModelFromSettings(settings)
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

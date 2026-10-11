package provider

import (
	"context"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vladkanatov/terraform-provider-orchard/internal/client"
)

type OrchardProvider struct {
	version string
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &OrchardProvider{version: version}
	}
}

func (p *OrchardProvider) Metadata(ctx context.Context, req provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "orchard"
	resp.Version = p.version
}

func (p *OrchardProvider) Schema(ctx context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"token": schema.StringAttribute{
				Required:    true,
				Sensitive:   true,
				Description: "Service account token for Orchard API authentication.",
			},
			"base_url": schema.StringAttribute{
				Required:    true,
				Description: "Base URL of the Orchard controller, e.g., http://localhost:8080",
			},
		},
	}
}

func (p *OrchardProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config struct {
		Token   types.String `tfsdk:"token"`
		BaseURL types.String `tfsdk:"base_url"`
	}
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	if config.Token.IsNull() || config.BaseURL.IsNull() {
		resp.Diagnostics.AddError(
			"Missing required configuration",
			"Both token and base_url are required.",
		)
		return
	}

	cl := client.NewClient(config.BaseURL.ValueString(), config.Token.ValueString())
	resp.DataSourceData = cl
	resp.ResourceData = cl
}

func (p *OrchardProvider) Resources(ctx context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewVMResource,
		NewServiceAccountResource,
		NewClusterSettingsResource,
	}
}

func (p *OrchardProvider) DataSources(ctx context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewControllerInfoDataSource,
		NewClusterSettingsDataSource,
		NewWorkerDataSource,
		NewVMDataSource,
		NewServiceAccountDataSource,
	}
}

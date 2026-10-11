package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vladkanatov/terraform-provider-orchard/internal/client"
)

type ServiceAccountDataSource struct {
	client *client.Client
}

func NewServiceAccountDataSource() datasource.DataSource {
	return &ServiceAccountDataSource{}
}

func (d *ServiceAccountDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_service_account"
}

func (d *ServiceAccountDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves service account(s).",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Optional:    true,
				Description: "Name of the service account. If omitted, returns all.",
			},
			"service_accounts": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{Computed: true},
						"token": schema.StringAttribute{
							Computed:  true,
							Sensitive: true,
						},
						"roles": schema.ListAttribute{
							ElementType: types.StringType,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *ServiceAccountDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *ServiceAccountDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config struct {
		Name types.String `tfsdk:"name"`
	}
	diags := req.Config.Get(ctx, &config)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	var accounts []client.ServiceAccount
	var err error
	if !config.Name.IsNull() {
		sa, err := d.client.GetServiceAccount(ctx, config.Name.ValueString())
		if err != nil {
			resp.Diagnostics.AddError("Failed to read service account", err.Error())
			return
		}
		accounts = []client.ServiceAccount{*sa}
	} else {
		accounts, err = d.client.ListServiceAccounts(ctx)
		if err != nil {
			resp.Diagnostics.AddError("Failed to list service accounts", err.Error())
			return
		}
	}

	saModels := make([]serviceAccountModel, len(accounts))
	for i, sa := range accounts {
		saModels[i] = serviceAccountModelFromSA(&sa)
	}
	state := struct {
		ServiceAccounts []serviceAccountModel `tfsdk:"service_accounts"`
	}{
		ServiceAccounts: saModels,
	}
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

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

type WorkerDataSource struct {
	client *client.Client
}

type workerModel struct {
	Name      types.String `tfsdk:"name"`
	Resources types.Map    `tfsdk:"resources"`
}

func NewWorkerDataSource() datasource.DataSource {
	return &WorkerDataSource{}
}

func (d *WorkerDataSource) Metadata(ctx context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_worker"
}

func (d *WorkerDataSource) Schema(ctx context.Context, req datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Retrieves information about all workers.",
		Attributes: map[string]schema.Attribute{
			"workers": schema.ListNestedAttribute{
				Computed: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Computed: true,
						},
						"resources": schema.MapAttribute{
							ElementType: types.Int64Type,
							Computed:    true,
						},
					},
				},
			},
		},
	}
}

func (d *WorkerDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *WorkerDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	workers, err := d.client.ListWorkers(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read workers", err.Error())
		return
	}

	// Преобразуем в TF модель
	workerModels := make([]workerModel, len(workers))
	for i, w := range workers {
		resVals := make(map[string]attr.Value)
		for k, v := range w.Resources {
			resVals[k] = types.Int64Value(int64(v))
		}
		workerModels[i] = workerModel{
			Name:      types.StringValue(w.Name),
			Resources: types.MapValueMust(types.Int64Type, resVals),
		}
	}

	state := struct {
		Workers []workerModel `tfsdk:"workers"`
	}{
		Workers: workerModels,
	}
	diags := resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

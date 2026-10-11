package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vladkanatov/terraform-provider-orchard/internal/client"
)

type ClusterSettingsResource struct {
	client *client.Client
}

func NewClusterSettingsResource() resource.Resource {
	return &ClusterSettingsResource{}
}

// Модель для кластерных настроек
type clusterSettingsModel struct {
	HostDirPolicies  types.List   `tfsdk:"host_dir_policies"`
	SchedulerProfile types.String `tfsdk:"scheduler_profile"`
}

// Типы атрибутов для host_dir_policies
var hostDirPolicyAttrTypes = map[string]attr.Type{
	"path_prefix": types.StringType,
	"ro":          types.BoolType,
}

func hostDirPoliciesToTFValues(policies []client.HostDirPolicy) []attr.Value {
	objs := make([]attr.Value, len(policies))
	for i, p := range policies {
		vals := map[string]attr.Value{
			"path_prefix": types.StringValue(p.PathPrefix),
			"ro":          types.BoolValue(p.Ro),
		}
		objs[i] = types.ObjectValueMust(hostDirPolicyAttrTypes, vals)
	}
	return objs
}

func (m *clusterSettingsModel) toClusterSettings() *client.ClusterSettings {
	settings := &client.ClusterSettings{
		SchedulerProfile: m.SchedulerProfile.ValueString(),
	}
	if !m.HostDirPolicies.IsNull() {
		var policies []client.HostDirPolicy
		m.HostDirPolicies.ElementsAs(context.Background(), &policies, false)
		settings.HostDirPolicies = policies
	}
	return settings
}

func clusterSettingsModelFromSettings(s *client.ClusterSettings) clusterSettingsModel {
	var policies []attr.Value
	if len(s.HostDirPolicies) > 0 {
		policies = hostDirPoliciesToTFValues(s.HostDirPolicies)
	}
	return clusterSettingsModel{
		HostDirPolicies: types.ListValueMust(
			types.ObjectType{AttrTypes: hostDirPolicyAttrTypes},
			policies,
		),
		SchedulerProfile: types.StringValue(s.SchedulerProfile),
	}
}

// Schema определяет схему ресурса
func (r *ClusterSettingsResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages cluster-wide settings in Orchard.",
		Attributes: map[string]schema.Attribute{
			"host_dir_policies": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"path_prefix": schema.StringAttribute{
							Required: true,
						},
						"ro": schema.BoolAttribute{
							Optional: true,
						},
					},
				},
				Description: "If not empty, allows hostDirs that match these policies.",
			},
			"scheduler_profile": schema.StringAttribute{
				Optional:    true,
				Description: "Scheduler profile: optimize-utilization or distribute-load.",
			},
		},
	}
}

// Configure настраивает клиент
func (r *ClusterSettingsResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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
	r.client = cl
}

// Create создает/обновляет настройки кластера (они всегда существуют)
func (r *ClusterSettingsResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan clusterSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := plan.toClusterSettings()
	updated, err := r.client.UpdateClusterSettings(ctx, settings)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update cluster settings", err.Error())
		return
	}

	state := clusterSettingsModelFromSettings(updated)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Read читает текущие настройки кластера
func (r *ClusterSettingsResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state clusterSettingsModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings, err := r.client.GetClusterSettings(ctx)
	if err != nil {
		resp.Diagnostics.AddError("Failed to read cluster settings", err.Error())
		return
	}

	newState := clusterSettingsModelFromSettings(settings)
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

// Update обновляет настройки кластера
func (r *ClusterSettingsResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan clusterSettingsModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	settings := plan.toClusterSettings()
	updated, err := r.client.UpdateClusterSettings(ctx, settings)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update cluster settings", err.Error())
		return
	}

	state := clusterSettingsModelFromSettings(updated)
	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

// Delete сбрасывает настройки кластера до дефолтных
func (r *ClusterSettingsResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	settings := &client.ClusterSettings{
		HostDirPolicies:  []client.HostDirPolicy{},
		SchedulerProfile: "optimize-utilization",
	}
	_, err := r.client.UpdateClusterSettings(ctx, settings)
	if err != nil {
		resp.Diagnostics.AddError("Failed to reset cluster settings", err.Error())
	}
}

// Metadata возвращает тип ресурса
func (r *ClusterSettingsResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_cluster_settings"
}

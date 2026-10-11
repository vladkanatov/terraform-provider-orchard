package provider

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/hashicorp/terraform-plugin-framework/attr"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/vladkanatov/terraform-provider-orchard/internal/client"
)

// hostDirAttrTypes описывает типы атрибутов для вложенного объекта hostDirs
var hostDirAttrTypes = map[string]attr.Type{
	"name": types.StringType,
	"path": types.StringType,
	"ro":   types.BoolType,
}

// dirsToTFValues преобразует срез client.HostDir в срез attr.Value (типа types.Object)
func dirsToTFValues(dirs []client.HostDir) []attr.Value {
	objs := make([]attr.Value, len(dirs))
	for i, d := range dirs {
		vals := map[string]attr.Value{
			"name": types.StringValue(d.Name),
			"path": types.StringValue(d.Path),
			"ro":   types.BoolValue(d.Ro),
		}
		objs[i] = types.ObjectValueMust(hostDirAttrTypes, vals)
	}
	return objs
}

// startupScriptModel – модель для вложенного блока startup_script (без изменений)
type startupScriptModel struct {
	ScriptContent types.String `tfsdk:"script_content"`
	Env           types.Map    `tfsdk:"env"`
}

// vmModel – модель ресурса VM (без изменений)
type vmModel struct {
	Name               types.String        `tfsdk:"name"`
	Generation         types.Int64         `tfsdk:"generation"`
	OS                 types.String        `tfsdk:"os"`
	Arch               types.String        `tfsdk:"arch"`
	Runtime            types.String        `tfsdk:"runtime"`
	Image              types.String        `tfsdk:"image"`
	ImagePullPolicy    types.String        `tfsdk:"image_pull_policy"`
	CPU                types.Float64       `tfsdk:"cpu"`
	Memory             types.Float64       `tfsdk:"memory"`
	DiskSize           types.Float64       `tfsdk:"disk_size"`
	NetSoftnet         types.Bool          `tfsdk:"net_softnet"`
	NetSoftnetAllow    types.List          `tfsdk:"net_softnet_allow"`
	NetSoftnetBlock    types.List          `tfsdk:"net_softnet_block"`
	Suspendable        types.Bool          `tfsdk:"suspendable"`
	NetBridged         types.String        `tfsdk:"net_bridged"`
	Headless           types.Bool          `tfsdk:"headless"`
	Nested             types.Bool          `tfsdk:"nested"`
	Username           types.String        `tfsdk:"username"`
	Password           types.String        `tfsdk:"password"`
	StartupScript      *startupScriptModel `tfsdk:"startup_script"`
	RestartPolicy      types.String        `tfsdk:"restart_policy"`
	RequestedResources types.Map           `tfsdk:"resources"` // Переименовали для ясности
	Labels             types.Map           `tfsdk:"labels"`
	HostDirs           types.List          `tfsdk:"host_dirs"`
	PowerState         types.String        `tfsdk:"power_state"`
	LocalName          types.String        `tfsdk:"local_name"`
	Status             types.String        `tfsdk:"status"`
	StatusMessage      types.String        `tfsdk:"status_message"`
	Worker             types.String        `tfsdk:"worker"`
	ObservedGeneration types.Int64         `tfsdk:"observed_generation"`
	IP                 types.String        `tfsdk:"ip"`
	// Добавляем поле для ресурсов из ответа API
	UsedResources types.Map `tfsdk:"used_resources"` // Новое поле
}

// toVMSpec преобразует vmModel в client.VMSpec (без изменений, кроме обработки nullable bool)
func (m *vmModel) toVMSpec() *client.VMSpec {
	spec := &client.VMSpec{
		OS:              m.OS.ValueString(),
		Arch:            m.Arch.ValueString(),
		Runtime:         m.Runtime.ValueString(),
		Image:           m.Image.ValueString(),
		ImagePullPolicy: m.ImagePullPolicy.ValueString(),
		CPU:             m.CPU.ValueFloat64(),
		Memory:          m.Memory.ValueFloat64(),
		DiskSize:        m.DiskSize.ValueFloat64(),
		Suspendable:     m.Suspendable.ValueBool(),
		NetBridged:      m.NetBridged.ValueString(),
		Headless:        m.Headless.ValueBool(),
		Nested:          m.Nested.ValueBool(),
		Username:        m.Username.ValueString(),
		Password:        m.Password.ValueString(),
		RestartPolicy:   m.RestartPolicy.ValueString(),
		PowerState:      m.PowerState.ValueString(),
	}

	if !m.RequestedResources.IsNull() && !m.RequestedResources.IsUnknown() {
		var res map[string]int64

		diags := m.RequestedResources.ElementsAs(
			context.Background(),
			&res,
			false,
		)

		if diags.HasError() {
			// Временно для диагностики
			fmt.Printf("ERROR converting resources: %v\n", diags)
		} else {
			spec.RequestedResources = make(map[string]int, len(res))

			for k, v := range res {
				spec.RequestedResources[k] = int(v)
			}
		}
	}

	if !m.NetSoftnet.IsNull() {
		spec.NetSoftnet = ptrBool(m.NetSoftnet.ValueBool())
	}

	// списки
	if !m.NetSoftnetAllow.IsNull() {
		var allow []string
		diags := m.NetSoftnetAllow.ElementsAs(context.Background(), &allow, false)
		if diags == nil || !diags.HasError() {
			spec.NetSoftnetAllow = allow
		}
	}
	if !m.NetSoftnetBlock.IsNull() {
		var block []string
		diags := m.NetSoftnetBlock.ElementsAs(context.Background(), &block, false)
		if diags == nil || !diags.HasError() {
			spec.NetSoftnetBlock = block
		}
	}

	// labels
	if !m.Labels.IsNull() {
		var labels map[string]string
		diags := m.Labels.ElementsAs(context.Background(), &labels, false)
		if diags == nil || !diags.HasError() {
			spec.Labels = labels
		}
	}

	// hostDirs
	if !m.HostDirs.IsNull() {
		var dirs []client.HostDir
		diags := m.HostDirs.ElementsAs(context.Background(), &dirs, false)
		if diags == nil || !diags.HasError() {
			spec.HostDirs = dirs
		}
	}

	// startup_script
	if m.StartupScript != nil {
		envMap := make(map[string]string)
		if !m.StartupScript.Env.IsNull() {
			m.StartupScript.Env.ElementsAs(context.Background(), &envMap, false)
		}
		spec.StartupScript = &client.StartupScript{
			ScriptContent: m.StartupScript.ScriptContent.ValueString(),
			Env:           envMap,
		}
	}

	return spec
}

// vmModelFromVM преобразует client.VM в vmModel с корректными типами attr.Value
func vmModelFromVM(vm *client.VM) vmModel {
	m := vmModel{
		Name:               types.StringValue(vm.Name),
		Generation:         types.Int64Value(int64(vm.Generation)),
		OS:                 types.StringValue(vm.OS),
		Arch:               types.StringValue(vm.Arch),
		Runtime:            types.StringValue(vm.Runtime),
		Image:              types.StringValue(vm.Image),
		ImagePullPolicy:    types.StringValue(vm.ImagePullPolicy),
		CPU:                types.Float64Value(vm.CPU),
		Memory:             types.Float64Value(vm.Memory),
		DiskSize:           types.Float64Value(vm.DiskSize),
		Suspendable:        types.BoolValue(vm.Suspendable),
		NetBridged:         types.StringValue(vm.NetBridged),
		Headless:           types.BoolValue(vm.Headless),
		Nested:             types.BoolValue(vm.Nested),
		Username:           types.StringValue(vm.Username),
		Password:           types.StringValue(vm.Password),
		RestartPolicy:      types.StringValue(vm.RestartPolicy),
		PowerState:         types.StringValue(vm.PowerState),
		LocalName:          types.StringValue(vm.LocalName),
		Status:             types.StringValue(vm.Status),
		StatusMessage:      types.StringValue(vm.StatusMessage),
		Worker:             types.StringValue(vm.Worker),
		ObservedGeneration: types.Int64Value(int64(vm.ObservedGeneration)),
	}

	// NetSoftnet (может быть nil)
	if vm.NetSoftnet != nil {
		m.NetSoftnet = types.BoolValue(*vm.NetSoftnet)
	} else {
		m.NetSoftnet = types.BoolNull()
	}

	// NetSoftnetAllow
	if len(vm.NetSoftnetAllow) > 0 {
		vals := make([]attr.Value, len(vm.NetSoftnetAllow))
		for i, v := range vm.NetSoftnetAllow {
			vals[i] = types.StringValue(v)
		}
		m.NetSoftnetAllow = types.ListValueMust(types.StringType, vals)
	} else {
		m.NetSoftnetAllow = types.ListNull(types.StringType)
	}

	if vm.NetBridged != "" {
		m.NetBridged = types.StringValue(vm.NetBridged)
	} else {
		m.NetBridged = types.StringNull()
	}

	// NetSoftnetBlock
	if len(vm.NetSoftnetBlock) > 0 {
		vals := make([]attr.Value, len(vm.NetSoftnetBlock))
		for i, v := range vm.NetSoftnetBlock {
			vals[i] = types.StringValue(v)
		}
		m.NetSoftnetBlock = types.ListValueMust(types.StringType, vals)
	} else {
		m.NetSoftnetBlock = types.ListNull(types.StringType)
	}

	if len(vm.RequestedResources) > 0 {
		vals := make(map[string]attr.Value)
		for k, v := range vm.RequestedResources {
			vals[k] = types.Int64Value(int64(v))
		}
		m.RequestedResources = types.MapValueMust(types.Int64Type, vals)
	} else {
		m.RequestedResources = types.MapNull(types.Int64Type)
	}

	// UsedResources (то, что вернул API)
	if len(vm.UsedResources) > 0 {
		vals := make(map[string]attr.Value)
		for k, v := range vm.UsedResources {
			vals[k] = types.Int64Value(int64(v))
		}
		m.UsedResources = types.MapValueMust(types.Int64Type, vals)
	} else {
		m.UsedResources = types.MapNull(types.Int64Type)
	}

	// Labels (map[string]string)
	if len(vm.Labels) > 0 {
		vals := make(map[string]attr.Value)
		for k, v := range vm.Labels {
			vals[k] = types.StringValue(v)
		}
		m.Labels = types.MapValueMust(types.StringType, vals)
	} else {
		m.Labels = types.MapNull(types.StringType)
	}

	// HostDirs
	if len(vm.HostDirs) > 0 {
		m.HostDirs = types.ListValueMust(
			types.ObjectType{AttrTypes: hostDirAttrTypes},
			dirsToTFValues(vm.HostDirs),
		)
	} else {
		m.HostDirs = types.ListNull(types.ObjectType{AttrTypes: hostDirAttrTypes})
	}

	// StartupScript
	if vm.StartupScript != nil {
		envVals := make(map[string]attr.Value)
		for k, v := range vm.StartupScript.Env {
			envVals[k] = types.StringValue(v)
		}
		m.StartupScript = &startupScriptModel{
			ScriptContent: types.StringValue(vm.StartupScript.ScriptContent),
			Env:           types.MapValueMust(types.StringType, envVals),
		}
	}

	return m
}

// ptrBool – вспомогательная функция
func ptrBool(b bool) *bool {
	return &b
}

// Методы Create, Read, Update, Delete остаются без изменений (они используют vmModelFromVM и toVMSpec)
// ...

type VMResource struct {
	client *client.Client
}

func NewVMResource() resource.Resource {
	return &VMResource{}
}

func (r *VMResource) Metadata(ctx context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_vm"
}
func (r *VMResource) Schema(ctx context.Context, req resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		Description: "Manages a virtual machine in Orchard.",
		Attributes: map[string]schema.Attribute{
			"name": schema.StringAttribute{
				Required:    true,
				Description: "Unique name of the VM within the cluster.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"generation": schema.Int64Attribute{
				Computed:    true,
				Description: "Incremented by the controller on each spec change.",
			},
			"os": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Operating system (darwin or linux). Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"arch": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Architecture (arm64 or amd64). Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"runtime": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Runtime (tart or vetu). Cannot be changed after creation.",
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"image": schema.StringAttribute{
				Required:    true,
				Description: "VM image reference.",
			},
			"image_pull_policy": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Image pull policy (IfNotPresent or Always).",
			},
			"cpu": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Number of CPUs.",
			},
			"memory": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Amount of RAM in megabytes.",
			},
			"disk_size": schema.Float64Attribute{
				Optional:    true,
				Computed:    true,
				Description: "Disk size in GB.",
			},
			"net_softnet": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Use Softnet network isolation.",
			},
			"net_softnet_allow": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "CIDRs to allow when using Softnet.",
			},
			"net_softnet_block": schema.ListAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "CIDRs to block when using Softnet.",
			},
			"suspendable": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable suspend/resume support.",
			},
			"net_bridged": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Bridged network interface (e.g., en0).",
			},
			"headless": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Run without graphics.",
			},
			"nested": schema.BoolAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Enable nested virtualization.",
			},
			"username": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "SSH username.",
			},
			"password": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Sensitive:   true,
				Description: "SSH password.",
			},
			"startup_script": schema.SingleNestedAttribute{
				Optional: true,
				Attributes: map[string]schema.Attribute{
					"script_content": schema.StringAttribute{
						Required: true,
					},
					"env": schema.MapAttribute{
						ElementType: types.StringType,
						Optional:    true,
					},
				},
			},
			"restart_policy": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Restart policy (Never or OnFailure).",
			},
			"resources": schema.MapAttribute{
				ElementType: types.Int64Type,
				Optional:    true,
				Description: "Resource requirements for the worker.",
			},
			"used_resources": schema.MapAttribute{ // <-- ДОБАВЬТЕ ЭТО ПОЛЕ
				ElementType: types.Int64Type,
				Computed:    true,
				Description: "Resources currently used by the VM on the worker (read-only).",
			},
			"labels": schema.MapAttribute{
				ElementType: types.StringType,
				Optional:    true,
				Description: "Labels for worker selection.",
			},
			"host_dirs": schema.ListNestedAttribute{
				Optional: true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"name": schema.StringAttribute{
							Optional: true,
						},
						"path": schema.StringAttribute{
							Required: true,
						},
						"ro": schema.BoolAttribute{
							Optional: true,
						},
					},
				},
				Description: "Host directories to mount.",
			},
			"power_state": schema.StringAttribute{
				Optional:    true,
				Computed:    true,
				Description: "Desired power state (running, stopped, suspended).",
			},
			"local_name": schema.StringAttribute{
				Computed:    true,
				Description: "Local name of the backing VM (read-only).",
			},
			"status": schema.StringAttribute{
				Computed:    true,
				Description: "Current status of the VM.",
			},
			"status_message": schema.StringAttribute{
				Computed: true,
			},
			"worker": schema.StringAttribute{
				Computed:    true,
				Description: "Worker assigned to this VM.",
			},
			"observed_generation": schema.Int64Attribute{
				Computed: true,
			},
			"ip": schema.StringAttribute{
				Computed:    true,
				Description: "Current IP address of the VM.",
			},
		},
	}
}

func (r *VMResource) Configure(ctx context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

// Create
func (r *VMResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var plan vmModel

	diags := req.Plan.Get(ctx, &plan)
	fmt.Printf("PLAN resources: %#v\n", plan.RequestedResources)
	fmt.Printf("PLAN resources elements: %#v\n", plan.RequestedResources.Elements())

	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	spec := plan.toVMSpec()

	fmt.Printf("SPEC resources: %#v\n", spec.RequestedResources)

	vm := &client.VM{
		VMMeta: client.VMMeta{
			Name: plan.Name.ValueString(),
		},
		VMSpec: *spec,
	}

	created, err := r.client.CreateVM(ctx, vm)
	if err != nil {
		resp.Diagnostics.AddError(
			"Failed to create VM",
			err.Error(),
		)
		return
	}

	// Ждем, пока контроллер действительно запустит VM, прежде чем завершать apply.
	ready, err := r.waitForVMRunning(ctx, created.Name)
	if err != nil {
		resp.Diagnostics.AddError(
			"VM did not become running",
			err.Error(),
		)
		return
	}

	ip, err := r.waitForVMIP(ctx, created.Name)
	if err != nil {
		resp.Diagnostics.AddError(
			"VM did not receive an IP address",
			err.Error(),
		)
		return
	}

	state := vmModelFromVM(ready)

	state.RequestedResources = plan.RequestedResources
	state.IP = types.StringValue(ip)

	diags = resp.State.Set(ctx, state)
	resp.Diagnostics.Append(diags...)
}

func (r *VMResource) waitForVMRunning(ctx context.Context, name string) (*client.VM, error) {
	const (
		pollInterval = 1 * time.Second
		waitTimeout  = 5 * time.Minute
	)

	deadline := time.NewTimer(waitTimeout)
	defer deadline.Stop()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	check := func() (*client.VM, error) {
		vm, err := r.client.GetVM(ctx, name)
		if err != nil {
			return nil, err
		}
		if vm.Status == "running" {
			return vm, nil
		}
		return nil, nil
	}

	for {
		vm, err := check()
		if err != nil {
			return nil, fmt.Errorf("failed to check VM status: %w", err)
		}
		if vm != nil {
			return vm, nil
		}

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-deadline.C:
			return nil, fmt.Errorf("timed out after %s waiting for VM %q to reach status running", waitTimeout, name)
		case <-ticker.C:
		}
	}
}

func (r *VMResource) waitForVMIP(ctx context.Context, name string) (string, error) {
	const (
		pollInterval = 1 * time.Second
		waitTimeout  = 5 * time.Minute
	)

	deadline := time.NewTimer(waitTimeout)
	defer deadline.Stop()
	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	for {
		ip, err := r.client.GetVMIP(ctx, name)
		if err != nil {
			return "", fmt.Errorf("failed to check VM IP: %w", err)
		}
		if ip != "" {
			return ip, nil
		}

		select {
		case <-ctx.Done():
			return "", ctx.Err()
		case <-deadline.C:
			return "", fmt.Errorf("timed out after %s waiting for VM %q to receive an IP address", waitTimeout, name)
		case <-ticker.C:
		}
	}
}

// Read
func (r *VMResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var state vmModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	vm, err := r.client.GetVM(ctx, state.Name.ValueString())
	if err != nil {
		if strings.Contains(err.Error(), "404") {
			resp.State.RemoveResource(ctx)
			return
		}
		resp.Diagnostics.AddError("Failed to read VM", err.Error())
		return
	}

	ip, err := r.client.GetVMIP(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to read VM IP", err.Error())
		return
	}

	newState := vmModelFromVM(vm)
	newState.IP = types.StringValue(ip)
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

// Update
func (r *VMResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var plan vmModel
	var state vmModel
	diags := req.Plan.Get(ctx, &plan)
	resp.Diagnostics.Append(diags...)
	diags = req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Отправляем только VMSpec (без мета-полей)
	spec := plan.toVMSpec()
	updated, err := r.client.UpdateVM(ctx, state.Name.ValueString(), spec)
	if err != nil {
		resp.Diagnostics.AddError("Failed to update VM", err.Error())
		return
	}

	newState := vmModelFromVM(updated)
	diags = resp.State.Set(ctx, newState)
	resp.Diagnostics.Append(diags...)
}

// Delete
func (r *VMResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var state vmModel
	diags := req.State.Get(ctx, &state)
	resp.Diagnostics.Append(diags...)
	if resp.Diagnostics.HasError() {
		return
	}

	err := r.client.DeleteVM(ctx, state.Name.ValueString())
	if err != nil {
		resp.Diagnostics.AddError("Failed to delete VM", err.Error())
		return
	}
}

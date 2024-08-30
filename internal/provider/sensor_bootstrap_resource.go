package provider

import (
	"context"
	"fmt"
	"math/rand/v2"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

const (
	SSHPortMin = 55000
	SSHPortMax = 65535
)

var _ resource.Resource = &SensorBootstrapResource{}
var _ resource.ResourceWithImportState = &SensorBootstrapResource{}

func NewSensorBootstrapResource() resource.Resource {
	return &SensorBootstrapResource{}
}

type SensorBootstrapResource struct {
	data *Data
}

type SensorBootstrapResourceModel struct {
	PublicIP        types.String `tfsdk:"public_ip"`
	InternalIP      types.String `tfsdk:"internal_ip"`
	SetupScript     types.String `tfsdk:"setup_script"`
	BootstrapScript types.String `tfsdk:"bootstrap_script"`
	SSHPort         types.Int32  `tfsdk:"ssh_port"`
	SSHPortSelected types.Int32  `tfsdk:"ssh_port_selected"`
}

func (r *SensorBootstrapResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sensor_bootstrap"
}

func (r *SensorBootstrapResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sensor bootstrap resource provides options to bootstrap a server. " +
			"It generates a script that can be used with a `remote-exec` provisioner to setup a sensor on a server.",
		Attributes: map[string]schema.Attribute{
			"public_ip": schema.StringAttribute{
				MarkdownDescription: "Public IP of the server to bootstrap",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"internal_ip": schema.StringAttribute{
				MarkdownDescription: "Internal IP of the server to bootstrap",
				Optional:            true,
			},
			"setup_script": schema.StringAttribute{
				MarkdownDescription: "Script that sets up the server environment",
				Sensitive:           true,
				Computed:            true,
			},
			"bootstrap_script": schema.StringAttribute{
				MarkdownDescription: "Script that can be run to boostrap a server",
				Computed:            true,
			},
			"ssh_port": schema.Int32Attribute{
				MarkdownDescription: "SSH port to configure after bootstrap. If not provided a random port is selected",
				Optional:            true,
			},
			"ssh_port_selected": schema.Int32Attribute{
				MarkdownDescription: "SSH port selected - same as ssh_port if set, otherwise randomly selected port",
				Computed:            true,
			},
		},
	}
}

func (r *SensorBootstrapResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*Data)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("expected *Data, got: %T. Please report this issue to the provider developers.",
				req.ProviderData),
		)

		return
	}

	r.data = data
}

func (r *SensorBootstrapResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SensorBootstrapResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	var (
		publicIPArg, internalIPArg, sshPortArg string
	)

	if !data.PublicIP.IsNull() {
		publicIPArg = fmt.Sprintf("-p %v ", data.PublicIP.ValueString())
	}
	if !data.InternalIP.IsNull() {
		internalIPArg = fmt.Sprintf("-i %v ", data.InternalIP.ValueString())
	}
	if data.SSHPort.IsNull() {
		sshPort := rand.Int32N(SSHPortMax-SSHPortMin) + SSHPortMin
		data.SSHPortSelected = types.Int32Value(sshPort)
	} else {
		data.SSHPortSelected = data.SSHPort
	}
	sshPortArg = fmt.Sprintf("-s %v", data.SSHPortSelected.ValueInt32())

	data.SetupScript = types.StringValue(
		fmt.Sprintf(`echo %s > ~/.greynoise.key`, r.data.APIKey),
	)
	data.BootstrapScript = types.StringValue(
		fmt.Sprintf(`KEY=$(cat ~/.greynoise.key) && \
curl -H "key: $KEY" -L %s | sudo bash -s -- -k $KEY %s%s%s`,
			r.data.Client.SensorBootstrapURL().String(),
			publicIPArg,
			internalIPArg,
			sshPortArg,
		),
	)

	tflog.Trace(ctx, "Created sensor bootstrap resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SensorBootstrapResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SensorBootstrapResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SensorBootstrapResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SensorBootstrapResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SensorBootstrapResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SensorBootstrapResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SensorBootstrapResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("public_ip"), req, resp)
}

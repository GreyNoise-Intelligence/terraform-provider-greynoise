package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/path"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/planmodifier"
	"github.com/hashicorp/terraform-plugin-framework/resource/schema/stringplanmodifier"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/GreyNoise-Intelligence/terraform-provider-greynoise/internal/client"
)

var _ resource.Resource = &SensorPersonaResource{}
var _ resource.ResourceWithImportState = &SensorPersonaResource{}

func NewSensorPersonaResource() resource.Resource {
	return &SensorPersonaResource{}
}

type SensorPersonaResource struct {
	data *Data
}

type SensorPersonaResourceModel struct {
	PersonaID types.String `tfsdk:"persona_id"`
	SensorID  types.String `tfsdk:"sensor_id"`
}

func (r *SensorPersonaResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sensor_persona"
}

func (r *SensorPersonaResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Sensor persona resource is used to manage the persona deployed to a sensor.`,
		Attributes: map[string]schema.Attribute{
			"persona_id": schema.StringAttribute{
				MarkdownDescription: "Persona ID for sensor update.",
				Required:            true,
			},
			"sensor_id": schema.StringAttribute{
				MarkdownDescription: "UUID of the sensor.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
		},
	}
}

func (r *SensorPersonaResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*Data)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Resource Configure Type",
			fmt.Sprintf("expected *Data, got: %T. Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	r.data = data
}

func (r *SensorPersonaResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SensorPersonaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.data.Client.UpdateSensor(ctx, data.SensorID.ValueString(), client.SensorUpdateRequest{
		Persona: data.PersonaID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Operation error",
			fmt.Sprintf("Error occurred while applying persona to sensor: %s", err.Error()),
		)

		return
	}

	tflog.Trace(ctx, "Created sensor persona resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SensorPersonaResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SensorPersonaResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sensor, err := r.data.Client.GetSensor(ctx, data.SensorID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Sensor error",
			fmt.Sprintf("Error occurred while checking sensor: %s", err.Error()),
		)

		return
	}
	data.PersonaID = types.StringValue(sensor.Persona)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SensorPersonaResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SensorPersonaResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.data.Client.UpdateSensor(ctx, data.SensorID.ValueString(), client.SensorUpdateRequest{
		Persona: data.PersonaID.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Operation error",
			fmt.Sprintf("Error occurred while applying persona to sensor: %s", err.Error()),
		)

		return
	}

	tflog.Trace(ctx, "Updated sensor persona resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SensorPersonaResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SensorPersonaResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SensorPersonaResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("sensor_id"), req, resp)
}

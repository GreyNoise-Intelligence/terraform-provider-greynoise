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

var _ resource.Resource = &SensorMetadataResource{}
var _ resource.ResourceWithImportState = &SensorMetadataResource{}

func NewSensorMetadataResource() resource.Resource {
	return &SensorMetadataResource{}
}

type SensorMetadataResource struct {
	data *Data
}

type SensorMetadataResourceModel struct {
	SensorID types.String `tfsdk:"sensor_id"`
	Name     types.String `tfsdk:"name"`
}

func (r *SensorMetadataResource) Metadata(_ context.Context, req resource.MetadataRequest, resp *resource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sensor_metadata"
}

func (r *SensorMetadataResource) Schema(_ context.Context, _ resource.SchemaRequest, resp *resource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: `Sensor metadata resource is used to manage metadata about a sensor.`,
		Attributes: map[string]schema.Attribute{
			"sensor_id": schema.StringAttribute{
				MarkdownDescription: "UUID of the sensor.",
				Required:            true,
				PlanModifiers: []planmodifier.String{
					stringplanmodifier.RequiresReplace(),
				},
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of the sensor.",
				Required:            true,
			},
		},
	}
}

func (r *SensorMetadataResource) Configure(_ context.Context, req resource.ConfigureRequest, resp *resource.ConfigureResponse) {
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

func (r *SensorMetadataResource) Create(ctx context.Context, req resource.CreateRequest, resp *resource.CreateResponse) {
	var data SensorMetadataResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.data.Client.UpdateSensor(ctx, data.SensorID.ValueString(), client.SensorUpdateRequest{
		Name: data.Name.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Operation error",
			fmt.Sprintf("Error occurred while updating sensor metadata: %s", err.Error()),
		)

		return
	}

	tflog.Trace(ctx, "Created sensor metadata resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SensorMetadataResource) Read(ctx context.Context, req resource.ReadRequest, resp *resource.ReadResponse) {
	var data SensorMetadataResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	sensor, err := r.data.Client.GetSensor(ctx, data.SensorID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Sensor error",
			fmt.Sprintf("Error occurred while getting sensor: %s", err.Error()),
		)

		return
	}

	data.Name = types.StringValue(sensor.Name)

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SensorMetadataResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SensorMetadataResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	if err := r.data.Client.UpdateSensor(ctx, data.SensorID.ValueString(), client.SensorUpdateRequest{
		Name: data.Name.ValueString(),
	}); err != nil {
		resp.Diagnostics.AddError(
			"Operation error",
			fmt.Sprintf("Error occurred while updating sensor metadata: %s", err.Error()),
		)

		return
	}

	tflog.Trace(ctx, "Updated sensor persona resource")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SensorMetadataResource) Delete(ctx context.Context, req resource.DeleteRequest, resp *resource.DeleteResponse) {
	var data SensorMetadataResourceModel

	resp.Diagnostics.Append(req.State.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}
}

func (r *SensorMetadataResource) ImportState(ctx context.Context, req resource.ImportStateRequest, resp *resource.ImportStateResponse) {
	resource.ImportStatePassthroughID(ctx, path.Root("sensor_id"), req, resp)
}

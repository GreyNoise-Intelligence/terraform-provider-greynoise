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

// var _ resource.ResourceWithConfigValidators = &SensorMetadataResource{}

func NewSensorMetadataResource() resource.Resource {
	return &SensorMetadataResource{}
}

type SensorMetadataResource struct {
	data *Data
}

type SensorMetadataResourceModel struct {
	SensorID types.String `tfsdk:"sensor_id"`
	Name     types.String `tfsdk:"name"`
	//Metadata SensorMetadataModel `tfsdk:"metadata"`
}

/*type SensorMetadataModel []SensorMetadatumModel

type SensorMetadatumModel struct {
	Access types.String `tfsdk:"access"`
	Name   types.String `tfsdk:"name"`
	Value  types.String `tfsdk:"value"`
}

func (s SensorMetadataModel) ToSensorMetadata() (*client.SensorMetadata, error) {
	result := client.SensorMetadata{
		Items: make([]client.SensorMetadatum, len(s)),
	}

	for _, datum := range s {
		cdatum := client.SensorMetadatum{
			Access: client.MetadataAccessReadonly, //default
			Name:   datum.Name.ValueString(),
			Val:    datum.Value.ValueString(),
		}

		if !datum.Access.IsNull() {
			cdatum.Access = client.MetadataAccess(datum.Access.ValueString())
		}

		if err := cdatum.Validate(); err != nil {
			return nil, err
		}
	}

	return &result, nil
}
*/

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
			/*"metadata": schema.ListNestedAttribute{
				MarkdownDescription: "Metadata tags for the sensor.",
				Optional:            true,
				NestedObject: schema.NestedAttributeObject{
					Attributes: map[string]schema.Attribute{
						"access": schema.StringAttribute{
							MarkdownDescription: "Access control value",
							Optional:            true,
						},
						"name": schema.StringAttribute{
							MarkdownDescription: "Name of tag",
							Required:            true,
						},
						"value": schema.StringAttribute{
							MarkdownDescription: "Value of tag",
							Required:            true,
						},
					},
				},
			},*/
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

	/*metadata, err := data.Metadata.ToSensorMetadata()
	if err != nil {
		resp.Diagnostics.AddError(
			"Metadata error",
			fmt.Sprintf("Error occurred while checking sensor metadata: %s", err.Error()),
		)

		return
	}
	*/

	if err := r.data.Client.UpdateSensor(ctx, data.SensorID.ValueString(), client.SensorUpdateRequest{
		Name: data.Name.ValueString(),
		//Metadata: metadata,
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

	/*if len(data.Metadata) != 0 {
		for _, item := range sensor.Metadata.Items {
			data.Metadata = SensorMetadataModel{
				{
					Access: types.StringValue(string(item.Access)),
					Name:   types.StringValue(item.Name),
					Value:  types.StringValue(item.Val),
				},
			}
		}
	}*/

	if !data.Name.IsNull() {
		data.Name = types.StringValue(sensor.Name)
	}

	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SensorMetadataResource) Update(ctx context.Context, req resource.UpdateRequest, resp *resource.UpdateResponse) {
	var data SensorMetadataResourceModel

	resp.Diagnostics.Append(req.Plan.Get(ctx, &data)...)

	if resp.Diagnostics.HasError() {
		return
	}

	/*metadata, err := data.Metadata.ToSensorMetadata()
	if err != nil {
		resp.Diagnostics.AddError(
			"Metadata error",
			fmt.Sprintf("Error occurred while checking sensor metadata: %s", err.Error()),
		)

		return
	}
	*/

	if err := r.data.Client.UpdateSensor(ctx, data.SensorID.ValueString(), client.SensorUpdateRequest{
		Name: data.Name.ValueString(),
		// Metadata: metadata,
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

/*func (r *SensorMetadataResource) ConfigValidators(_ context.Context) []resource.ConfigValidator {
	return []resource.ConfigValidator{
		resourcevalidator.AtLeastOneOf(
			path.MatchRoot("name"),
			path.MatchRoot("metadata"),
		),
	}
}
*/

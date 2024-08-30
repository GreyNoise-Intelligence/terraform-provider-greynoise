package provider

import (
	"context"
	"errors"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/GreyNoise-Intelligence/terraform-provider-greynoise/internal/client"
)

var _ datasource.DataSource = &SensorDataSource{}

func NewSensorDataSource() datasource.DataSource {
	return &SensorDataSource{}
}

type SensorDataSource struct {
	data *Data
}

type SensorDataSourceModel struct {
	ID       types.String `tfsdk:"id"`
	PublicIP types.String `tfsdk:"public_ip"`
}

func (d *SensorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sensor"
}

func (d *SensorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sensor data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "Sensor ID",
				Computed:            true,
			},
			"public_ip": schema.StringAttribute{
				MarkdownDescription: "Sensor public IP",
				Optional:            true,
			},
		},
	}
}

func (d *SensorDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
	// Prevent panic if the provider has not been configured.
	if req.ProviderData == nil {
		return
	}

	data, ok := req.ProviderData.(*Data)

	if !ok {
		resp.Diagnostics.AddError(
			"Unexpected Data Source Configure Type",
			fmt.Sprintf("expected *Data, got: %T. "+
				"Please report this issue to the provider developers.", req.ProviderData),
		)

		return
	}

	d.data = data
}

func (d *SensorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var data SensorDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &data)...)
	if resp.Diagnostics.HasError() {
		return
	}

	if resp.Diagnostics.HasError() {
		return
	}

	sensor, err := d.getSensor(ctx, data)
	if err != nil {
		if errors.Is(err, client.ErrNotFound) {
			resp.Diagnostics.AddError(
				"Sensor error",
				fmt.Sprintf("Sensor not found"),
			)

			return
		}

		resp.Diagnostics.AddError(
			"Sensor error",
			fmt.Sprintf("Error occurred while retrieving sensor: %s", err.Error()),
		)

		return
	}
	data.ID = types.StringValue(sensor.ID)

	tflog.Trace(ctx, "Read sensor data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &data)...)
}

func (r *SensorDataSource) getSensor(ctx context.Context, data SensorDataSourceModel) (*client.Sensor, error) {
	c := r.data.Client

	if !data.ID.IsNull() {
		return c.GetSensor(ctx, data.ID.ValueString())
	}

	ip := data.PublicIP.ValueString()
	result, err := c.SensorsSearch(ctx, client.SensorSearchFilter{
		Filter:     ip,
		Page:       0,
		PageSize:   1,
		SortBy:     client.SensorSortByCreatedAt,
		Descending: true,
	})
	if err != nil {
		return nil, err
	}

	if len(result.Items) < 1 {
		return nil, fmt.Errorf("no sensor found matching IP: %s", ip)
	}

	return &result.Items[0], nil
}

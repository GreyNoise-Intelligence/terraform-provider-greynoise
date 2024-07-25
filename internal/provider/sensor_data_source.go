package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
	"golang.org/x/crypto/ssh"
)

var _ datasource.DataSource = &SensorDataSource{}

func NewSensorDataSource() datasource.DataSource {
	return &SensorDataSource{}
}

type SensorDataSource struct {
	data   *Data
	client ssh.Client
}

type SensorDataSourceModel struct {
	ID          types.String `tfsdk:"id"`
	TenantID    types.String `tfsdk:"tenant_id"`
	Name        types.String `tfsdk:"name"`
	GreytewayID types.String `tfsdk:"greyteway_id"`
	PublicIps   types.List   `tfsdk:"public_ips"`
	/*DefaultGateway types.String `tfsdk:"default_gateway"`
	AccessPort     types.Int32  `tfsdk:"access_port"`
	SensorType     types.String `tfsdk:"sensor_type"`
	SensorAccess   types.String `tfsdk:"sensor_access"`
	PersonaID      types.String `tfsdk:"persona_id"`
	Status         types.String `tfsdk:"status"`
	Disabled       types.Bool   `tfsdk:"disabled"`*/
}

func (d *SensorDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_sensor"
}

func (d *SensorDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Sensor data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of sensor",
				Optional:            true,
			},
			"tenant_id": schema.StringAttribute{
				MarkdownDescription: "Tenant ID of sensor",
				Computed:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of sensor",
				Computed:            true,
			},
			"greyteway_id": schema.StringAttribute{
				MarkdownDescription: "Greyteway ID for sensor",
				Computed:            true,
			},
			"public_ips": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Artifact link for persona",
				Computed:            true,
			},
			"bootstrap_connection": schema.SingleNestedAttribute{
				MarkdownDescription: "Bootstrap connection to retrieve sensor information.",
				Attributes: map[string]schema.Attribute{
					"host": schema.StringAttribute{
						MarkdownDescription: "Host to connect to",
						Required:            true,
					},
					"port": schema.Int32Attribute{
						MarkdownDescription: "SSH port to connect over",
						Required:            true,
					},
					"user": schema.StringAttribute{
						MarkdownDescription: "SSH user",
						Required:            true,
					},
					"password": schema.StringAttribute{
						MarkdownDescription: "SSH password",
						Optional:            true,
					},
					"private_key": schema.StringAttribute{
						MarkdownDescription: "SSH private key",
						Optional:            true,
					},
				},
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

	// configure SSH client to connect and set client. Use it to SSH into box and retrieve sensor ID
	/*ssh.ClientConfig{
		Config:            ssh.Config{},
		User:              "",
		Auth:              []ssh.AuthMethod{
			ssh.Password(),
			ssh.PublicKeys(),
		},
		HostKeyCallback:   nil,
		BannerCallback:    nil,
		ClientVersion:     "",
		HostKeyAlgorithms: nil,
		Timeout:           0,
	}
	*/
	d.data = data
}

func (d *SensorDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config SensorDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Search and process results
	/*result, err := d.data.Client.GetPersona(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Persona error",
			fmt.Sprintf("Error occurred while retrieving persona: %s", err.Error()),
		)

		return
	}*/

	// Set computed attributes
	/*config.Name = types.StringValue(result.Name)
	config.Description = types.StringValue(result.Description)
	config.Author = types.StringValue(result.Author)
	config.ArtifactLink = types.StringValue(result.ArtifactLink)
	config.Tier = types.StringValue(result.Tier)
	config.OperatingSystem = types.StringValue(result.OperatingSystem)*/

	// List attributes
	/*categories, diags := types.ListValueFrom(ctx, types.StringType, result.Categories)
	resp.Diagnostics.Append(diags...)
	config.Categories = categories

	protocols, diags := types.ListValueFrom(ctx, types.StringType, result.ApplicationProtocols)
	resp.Diagnostics.Append(diags...)
	config.ApplicationProtocols = protocols

	vulnerabilities, diags := types.ListValueFrom(ctx, types.StringType, result.AssociatedVulnerabilities)
	resp.Diagnostics.Append(diags...)
	config.AssociatedVulnerabilities = vulnerabilities

	ports, diags := types.ListValueFrom(ctx, types.Int32Type, result.Ports)
	resp.Diagnostics.Append(diags...)
	config.Ports = ports*/

	tflog.Trace(ctx, "read sensor data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

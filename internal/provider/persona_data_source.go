package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"
)

var _ datasource.DataSource = &PersonaDataSource{}

func NewPersonaDataSource() datasource.DataSource {
	return &PersonaDataSource{}
}

type PersonaDataSource struct {
	data *Data
}

type PersonaDataSourceModel struct {
	ID                        types.String `tfsdk:"id"`
	Name                      types.String `tfsdk:"name"`
	Description               types.String `tfsdk:"description"`
	Author                    types.String `tfsdk:"author"`
	ArtifactLink              types.String `tfsdk:"artifact_link"`
	Tier                      types.String `tfsdk:"tier"`
	Categories                types.List   `tfsdk:"categories"`
	ApplicationProtocols      types.List   `tfsdk:"application_protocols"`
	Ports                     types.List   `tfsdk:"ports"`
	OperatingSystem           types.String `tfsdk:"operating_system"`
	AssociatedVulnerabilities types.List   `tfsdk:"associated_vulnerabilities"`
}

func (d *PersonaDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_persona"
}

func (d *PersonaDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Persona data source",
		Attributes: map[string]schema.Attribute{
			"id": schema.StringAttribute{
				MarkdownDescription: "ID of persona",
				Required:            true,
			},
			"name": schema.StringAttribute{
				MarkdownDescription: "Name of persona",
				Computed:            true,
			},
			"description": schema.StringAttribute{
				MarkdownDescription: "Description of persona",
				Computed:            true,
			},
			"author": schema.StringAttribute{
				MarkdownDescription: "Author of persona",
				Computed:            true,
			},
			"artifact_link": schema.StringAttribute{
				MarkdownDescription: "Artifact link for persona",
				Computed:            true,
			},
			"tier": schema.StringAttribute{
				MarkdownDescription: "Tier to which this persona belongs",
				Computed:            true,
			},
			"categories": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "IDs of personas that match criteria",
				Computed:            true,
			},
			"application_protocols": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Application protocols used by this persona",
				Computed:            true,
			},
			"ports": schema.ListAttribute{
				ElementType:         types.Int32Type,
				MarkdownDescription: "Application protocols used by this persona",
				Computed:            true,
			},
			"operating_system": schema.StringAttribute{
				MarkdownDescription: "Operating system of persona",
				Computed:            true,
			},
			"associated_vulnerabilities": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "Vulnerabilities associate with this persona",
				Computed:            true,
			},
		},
	}
}

func (d *PersonaDataSource) Configure(ctx context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PersonaDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PersonaDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Search and process results
	result, err := d.data.Client.GetPersona(config.ID.ValueString())
	if err != nil {
		resp.Diagnostics.AddError(
			"Persona error",
			fmt.Sprintf("Error occurred while retrieving persona: %s", err.Error()),
		)

		return
	}

	// Set computed attributes
	config.Name = types.StringValue(result.Name)
	config.Description = types.StringValue(result.Description)
	config.Author = types.StringValue(result.Author)
	config.ArtifactLink = types.StringValue(result.ArtifactLink)
	config.Tier = types.StringValue(result.Tier)
	config.OperatingSystem = types.StringValue(result.OperatingSystem)

	// List attributes
	categories, diags := types.ListValueFrom(ctx, types.StringType, result.Categories)
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
	config.Ports = ports

	tflog.Trace(ctx, "read persona data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

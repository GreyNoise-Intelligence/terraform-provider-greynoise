package provider

import (
	"context"
	"fmt"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/datasource/schema"
	"github.com/hashicorp/terraform-plugin-framework/types"
	"github.com/hashicorp/terraform-plugin-log/tflog"

	"github.com/GreyNoise-Intelligence/terraform-provider-greynoise/internal/client"
)

var _ datasource.DataSource = &PersonasDataSource{}

func NewPersonasDataSource() datasource.DataSource {
	return &PersonasDataSource{}
}

type PersonasDataSource struct {
	data *Data
}

type PersonasDataSourceModel struct {
	Tier     types.String `tfsdk:"tier"`
	Category types.String `tfsdk:"category"`
	Protocol types.String `tfsdk:"protocol"`
	Search   types.String `tfsdk:"search"`
	Limit    types.Int32  `tfsdk:"limit"`
	IDs      types.List   `tfsdk:"ids"`
	Total    types.Int32  `tfsdk:"total"`
}

func (d *PersonasDataSource) Metadata(_ context.Context, req datasource.MetadataRequest, resp *datasource.MetadataResponse) {
	resp.TypeName = req.ProviderTypeName + "_personas"
}

func (d *PersonasDataSource) Schema(_ context.Context, _ datasource.SchemaRequest, resp *datasource.SchemaResponse) {
	resp.Schema = schema.Schema{
		MarkdownDescription: "Personas data source",
		Attributes: map[string]schema.Attribute{
			"category": schema.StringAttribute{
				MarkdownDescription: "Category of persona",
				Optional:            true,
			},
			"protocol": schema.StringAttribute{
				MarkdownDescription: "Protocol of persona",
				Optional:            true,
			},
			"search": schema.StringAttribute{
				MarkdownDescription: "Partial text search on persona name",
				Optional:            true,
			},
			"tier": schema.StringAttribute{
				MarkdownDescription: "Tier of persona",
				Optional:            true,
			},
			"limit": schema.Int32Attribute{
				MarkdownDescription: "Limit number of personas to return",
				Optional:            true,
			},
			"ids": schema.ListAttribute{
				ElementType:         types.StringType,
				MarkdownDescription: "IDs of personas that match criteria",
				Computed:            true,
			},
			"total": schema.Int32Attribute{
				MarkdownDescription: "Number of matched personas",
				Computed:            true,
			},
		},
	}
}

func (d *PersonasDataSource) Configure(_ context.Context, req datasource.ConfigureRequest, resp *datasource.ConfigureResponse) {
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

func (d *PersonasDataSource) Read(ctx context.Context, req datasource.ReadRequest, resp *datasource.ReadResponse) {
	var config PersonasDataSourceModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Search and process results
	result, err := d.data.Client.PersonasSearch(client.PersonaSearchFilters{
		Workspace:  d.data.Client.WorkspaceID().String(),
		Tiers:      config.Tier.ValueString(),
		Categories: config.Category.ValueString(),
		Protocols:  config.Category.ValueString(),
		Search:     config.Search.ValueString(),
		PageSize:   config.Limit.ValueInt32(),
	})
	if err != nil {
		resp.Diagnostics.AddError(
			"Personas error",
			fmt.Sprintf("Error occurred while retrieving personas: %s", err.Error()),
		)

		return
	}

	personaIDs := make([]string, len(result.Items))
	for i, item := range result.Items {
		personaIDs[i] = item.ID
	}

	// Set computed attributes
	personaIDsList, diags := types.ListValueFrom(ctx, types.StringType, personaIDs)
	resp.Diagnostics.Append(diags...)
	config.IDs = personaIDsList

	total := types.Int32Value(result.Pagination.TotalItems)
	config.Total = total

	tflog.Trace(ctx, "read personas data source")
	resp.Diagnostics.Append(resp.State.Set(ctx, &config)...)
}

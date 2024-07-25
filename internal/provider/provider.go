package provider

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"os"

	"github.com/hashicorp/terraform-plugin-framework/datasource"
	"github.com/hashicorp/terraform-plugin-framework/function"
	"github.com/hashicorp/terraform-plugin-framework/provider"
	"github.com/hashicorp/terraform-plugin-framework/provider/schema"
	"github.com/hashicorp/terraform-plugin-framework/resource"
	"github.com/hashicorp/terraform-plugin-framework/types"

	"github.com/GreyNoise-Intelligence/terraform-provider-greynoise/internal/client"
)

const (
	envVarAPIKey = "GN_API_KEY"
)

// Ensure GreyNoiseProvider satisfies various provider interfaces.
var _ provider.Provider = &GreyNoiseProvider{}
var _ provider.ProviderWithFunctions = &GreyNoiseProvider{}

// GreyNoiseProvider defines the provider implementation.
type GreyNoiseProvider struct {
	// version is set to the provider version on release, "dev" when the
	// provider is built and ran locally, and "test" when running acceptance
	// testing.
	version string
}

// GreyNoiseProviderModel describes the provider data model.
type GreyNoiseProviderModel struct {
	BaseURL     types.String `tfsdk:"base_url"`
	WorkspaceID types.String `tfsdk:"workspace_id"`
	APIKey      types.String `tfsdk:"api_key"`
}

func (p *GreyNoiseProvider) Metadata(_ context.Context, _ provider.MetadataRequest, resp *provider.MetadataResponse) {
	resp.TypeName = "greynoise"
	resp.Version = p.version
}

func (p *GreyNoiseProvider) Schema(_ context.Context, req provider.SchemaRequest, resp *provider.SchemaResponse) {
	resp.Schema = schema.Schema{
		Attributes: map[string]schema.Attribute{
			"api_key": schema.StringAttribute{
				MarkdownDescription: "GreyNoise API Key",
				Optional:            true,
				Sensitive:           true,
			},
			"workspace_id": schema.StringAttribute{
				MarkdownDescription: "GreyNoise Workspace ID",
				Required:            true,
			},
			"base_url": schema.StringAttribute{
				MarkdownDescription: "GreyNoise API Base URL",
				Optional:            true,
			},
		},
	}
}

func (p *GreyNoiseProvider) Configure(ctx context.Context, req provider.ConfigureRequest, resp *provider.ConfigureResponse) {
	var config GreyNoiseProviderModel

	resp.Diagnostics.Append(req.Config.Get(ctx, &config)...)
	if resp.Diagnostics.HasError() {
		return
	}

	// Check environment variable
	var apiKey string
	if config.APIKey.IsNull() {
		apiKey = os.Getenv(envVarAPIKey)
	} else {
		apiKey = config.APIKey.ValueString()
	}

	if apiKey == "" {
		resp.Diagnostics.AddError(
			"No API key set",
			fmt.Sprintf("API key must be provided in configuration or set via environment variable: %s",
				envVarAPIKey),
		)

		return
	}

	// Validate parameters and create client
	options := []client.Option{
		client.WithWorkspaceID(config.WorkspaceID.ValueString()),
	}

	if !config.BaseURL.IsNull() {
		baseURL, err := url.Parse(config.BaseURL.ValueString())
		if err != nil {
			resp.Diagnostics.AddError(
				"Error parsing GreyNoise API base URL",
				fmt.Sprintf("Error attempting to parse base URL: %s", err.Error()),
			)

			return
		}

		options = append(options, client.WithBaseURL(baseURL))
	}

	c, err := client.New(apiKey, options...)
	if err != nil {
		resp.Diagnostics.AddError(
			"Error creating GreyNoise API client",
			fmt.Sprintf("Error attempting to create client: %s", err.Error()),
		)

		return
	}

	if err := c.Ping(); err != nil {
		var statusCodeErr *client.ErrUnexpectedStatusCode
		if errors.As(err, &statusCodeErr) {
			if statusCodeErr.StatusCode() == http.StatusUnauthorized {
				resp.Diagnostics.AddError(
					"Unable to authenticate with GreyNoise API",
					"Check your API key to ensure it is valid and unexpired.",
				)

				return
			}
		}

		resp.Diagnostics.AddError(
			"Unable to ping the GreyNoise API",
			fmt.Sprintf("Error attempting to ping: %s", err.Error()),
		)

		return
	}

	data := &Data{
		APIKey:      apiKey,
		WorkspaceID: config.WorkspaceID.ValueString(),
		Client:      c,
	}
	resp.DataSourceData = data
	resp.ResourceData = data
}

func (p *GreyNoiseProvider) Resources(_ context.Context) []func() resource.Resource {
	return []func() resource.Resource{
		NewSensorBootstrapResource,
	}
}

func (p *GreyNoiseProvider) DataSources(_ context.Context) []func() datasource.DataSource {
	return []func() datasource.DataSource{
		NewPersonaDataSource,
		NewPersonasDataSource,
		NewSensorDataSource,
	}
}

func (p *GreyNoiseProvider) Functions(_ context.Context) []func() function.Function {
	return []func() function.Function{
		NewExampleFunction,
	}
}

func New(version string) func() provider.Provider {
	return func() provider.Provider {
		return &GreyNoiseProvider{
			version: version,
		}
	}
}

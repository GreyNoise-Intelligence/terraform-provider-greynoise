package provider

import "github.com/GreyNoise-Intelligence/terraform-provider-greynoise/internal/client"

type Data struct {
	Client      *client.GreyNoiseClient
	APIKey      string
	WorkspaceID string
}

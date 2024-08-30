package provider

import (
	"fmt"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
)

func TestAccSensorPersonaResource(t *testing.T) {
	t.Parallel()

	mockServer := defaultMockAPIServer()
	// mockWorkspaceID := mockServer.Account.WorkspaceID.String()
	mockAPIKey := mockServer.APIKey

	server := mockServer.Server()

	testCases := []struct {
		name        string
		config      string
		check       resource.TestCheckFunc
		expectError *regexp.Regexp
	}{
		{
			name: "success - public ip",
			config: `
			resource "greynoise_sensor_persona" "this" {
              sensor_filter = {
                public_ip = "44.202.75.6"
              }
              persona_id = "55ec7e60-79e6-4240-8cd8-dadcc28c80e8"
			}`,
			check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("greynoise_sensor_persona.this", "sensor_filter.public_ip",
					"44.202.75.6",
				),
				resource.TestCheckResourceAttr("greynoise_sensor_persona.this", "persona_id",
					"55ec7e60-79e6-4240-8cd8-dadcc28c80e8",
				),
			),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config: fmt.Sprintf(`
						provider "greynoise" {
						  base_url = "%s"
						  api_key  = "%s"
						}
						`, server.URL, mockAPIKey) + tc.config,
						Check:       tc.check,
						ExpectError: tc.expectError,
					},
				},
			})
		})
	}
}

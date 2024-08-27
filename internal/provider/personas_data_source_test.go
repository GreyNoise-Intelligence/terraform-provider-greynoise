package provider

import (
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"

	"github.com/GreyNoise-Intelligence/terraform-provider-greynoise/internal/client"
)

func TestAccPersonasDataSource(t *testing.T) {
	t.Parallel()

	mockServer := defaultMockAPIServer()
	mockServer.RegisterMatch("/v1/personas",
		func(url *url.URL) bool {
			return url.Query().Get("search") == "tomcat"
		},
		http.StatusOK,
		client.PersonaSearchResponse{
			Items: []client.Persona{
				{
					ID:                 "ac65d8a0-ed21-417e-a1a2-65a4e09c3144",
					Name:               "RDP Server",
					Author:             "GreyNoiseIO",
					ArtifactLink:       "ami-0f16b7gv81ce26ae0",
					Tier:               "premium",
					InstanceManagement: "per_gateway",
					Categories: []string{
						"honeypot",
						"rev1",
					},
					Description: "A Remote Desktop Protocol server. Designed to observe credential " +
						"bruteforce activity.",
					ApplicationProtocols: []string{
						"rdp",
					},
					Ports: []int32{
						3389,
					},
				},
			},
			Pagination: client.Pagination{
				Page:       1,
				PageSize:   2,
				TotalItems: 1,
			},
		})

	server := mockServer.Server()

	testCases := []struct {
		name        string
		config      string
		check       resource.TestCheckFunc
		expectError *regexp.Regexp
	}{
		{
			name: "happy path",
			config: `	
			data "greynoise_personas" "this" {
			  search = "tomcat"
			  limit  = 2
			}`,
			check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.greynoise_personas.this", "ids.#", "1"),
				resource.TestCheckResourceAttr("data.greynoise_personas.this", "ids.0",
					"ac65d8a0-ed21-417e-a1a2-65a4e09c3144"),
				resource.TestCheckResourceAttr("data.greynoise_personas.this", "total", "1"),
			),
		},
		{
			name: "invalid limit",
			config: `	
			data "greynoise_personas" "this" {
			  search = "tomcat"
			  limit  = "test"
			}`,
			expectError: regexp.MustCompile(`Inappropriate value for attribute "limit": a number is required.`),
		},
		{
			name: "unexpected response",
			config: `	
			data "greynoise_personas" "this" {
			  search = "not-tomcat"
			  limit  = 2
			}`,
			expectError: regexp.MustCompile(`Error occurred while retrieving personas: invalid status code: 404`),
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
						`, server.URL, mockServer.APIKey) + tc.config,
						Check:       tc.check,
						ExpectError: tc.expectError,
					},
				},
			})
		})
	}
}

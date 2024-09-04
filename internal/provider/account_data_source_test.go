package provider

import (
	"fmt"
	"os"
	"regexp"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"
)

func TestAccAccountDataSource(t *testing.T) {
	t.Parallel()

	mockServer := defaultMockAPIServer()
	mockUserID := mockServer.Account.UserID.String()
	mockWorkspaceID := mockServer.Account.WorkspaceID.String()
	mockAPIKey := mockServer.APIKey

	server := mockServer.Server()

	testCases := []struct {
		name        string
		config      string
		env         map[string]string
		check       resource.TestCheckFunc
		expectError *regexp.Regexp
	}{
		{
			name: "success",
			config: fmt.Sprintf(`
			provider "greynoise" {
			  base_url = "%s"
			  api_key  = "%s"
			}

			data "greynoise_account" "this" {}
			`, server.URL, mockAPIKey),
			check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.greynoise_account.this", "user_id", mockUserID),
				resource.TestCheckResourceAttr("data.greynoise_account.this", "workspace_id", mockWorkspaceID),
			),
		},
		{
			name: "success - env key",
			env: map[string]string{
				"GN_API_KEY": mockAPIKey,
			},
			config: fmt.Sprintf(`
			provider "greynoise" {
			  base_url = "%s"	
			}

			data "greynoise_account" "this" {}
			`, server.URL),
			check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.greynoise_account.this", "user_id", mockUserID),
				resource.TestCheckResourceAttr("data.greynoise_account.this", "workspace_id", mockWorkspaceID),
			),
		},
		{
			name: "invalid key",
			config: fmt.Sprintf(`
			provider "greynoise" {
			  base_url = "%s"
			  api_key  = "fake-api-key"
			}

			data "greynoise_account" "this" {}
			`, server.URL),
			expectError: regexp.MustCompile(`Error attempting to create client: account error: invalid status code: 401,`),
		},
		{
			name: "invalid URL",
			config: `
			provider "greynoise" {
			  base_url = "https://fake.greynoise.io"
			  api_key  = "fake-api-key"
			}

			data "greynoise_account" "this" {}`,
			expectError: regexp.MustCompile(`Error attempting to create client: account error:(.*\s*)*dial tcp: lookup(.*\s*)*no such host`),
		},
		{
			name: "missing key",
			config: fmt.Sprintf(`
			provider "greynoise" {
			  base_url = "%s"
			}

			data "greynoise_account" "this" {}
			`, server.URL),
			expectError: regexp.MustCompile(`Error: No API key set`),
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			for k, v := range tc.env {
				assert.NoError(t, os.Setenv(k, v))
			}

			defer func() {
				for k := range tc.env {
					assert.NoError(t, os.Unsetenv(k))
				}
			}()

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps: []resource.TestStep{
					{
						Config:      tc.config,
						Check:       tc.check,
						ExpectError: tc.expectError,
					},
				},
			})
		})
	}
}

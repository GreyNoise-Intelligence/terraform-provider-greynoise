package provider

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/stretchr/testify/assert"

	"github.com/GreyNoise-Intelligence/terraform-provider-greynoise/internal/client"
)

func TestAccSensorDataSource(t *testing.T) {
	t.Parallel()

	testSensor := client.Sensor{
		ID:   "1d6aed11-f2de-48f9-9526-8fb72be10700",
		Name: "Gifted Trout",
		PublicIps: []string{
			"159.223.200.217",
		},
		Persona:   "501c5e5a-cf2e-4401-844a-04d4391b1332",
		Status:    "healthy",
		Disabled:  false,
		LastSeen:  time.Date(2024, 8, 27, 16, 27, 2, 0, time.UTC),
		CreatedAt: time.Date(2024, 8, 10, 3, 2, 22, 0, time.UTC),
		UpdatedAt: time.Date(2024, 8, 26, 13, 53, 07, 0, time.UTC),
	}

	mockServer := defaultMockAPIServer()
	mockWorkspaceID := mockServer.Account.WorkspaceID.String()
	mockAPIKey := mockServer.APIKey

	mockServer.RegisterMatch(fmt.Sprintf("/v1/workspaces/%s/sensors", mockWorkspaceID),
		func(url *url.URL) bool {
			return url.Query().Get("filter") == testSensor.PublicIps[0]
		},
		http.StatusOK,
		client.SensorSearchResponse{
			Items: []client.Sensor{
				testSensor,
			},
			Pagination: client.Pagination{
				Page:       1,
				PageSize:   2,
				TotalItems: 1,
			},
		})
	mockServer.Register(fmt.Sprintf("/v1/workspaces/%s/sensors/%s", mockWorkspaceID, testSensor.ID),
		http.StatusOK,
		testSensor,
	)

	server := mockServer.Server()

	testCases := []struct {
		name        string
		config      string
		env         map[string]string
		check       resource.TestCheckFunc
		expectError *regexp.Regexp
	}{
		{
			name: "success - public ip",
			config: `
			data "greynoise_sensor" "this" {
			  public_ip = "159.223.200.217"
			}
			`,
			check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.greynoise_sensor.this", "public_ip", "159.223.200.217"),
				resource.TestCheckResourceAttr("data.greynoise_sensor.this", "id", testSensor.ID),
			),
		},
		{
			name: "success - id",
			env: map[string]string{
				"GN_API_KEY": mockAPIKey,
			},
			config: `
			data "greynoise_sensor" "this" {
			  id = "1d6aed11-f2de-48f9-9526-8fb72be10700"
			}
			`,
			check: resource.ComposeAggregateTestCheckFunc(
				resource.TestCheckResourceAttr("data.greynoise_sensor.this", "id", testSensor.ID),
			),
		},
		{
			name: "not found",
			config: `
			data "greynoise_sensor" "this" {
			  id = "unknown"
			}
			`,
			expectError: regexp.MustCompile(`Sensor not found`),
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
				for k, _ := range tc.env {
					assert.NoError(t, os.Unsetenv(k))
				}
			}()

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

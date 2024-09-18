package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"regexp"
	"testing"
	"time"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"

	"github.com/GreyNoise-Intelligence/terraform-provider-greynoise/internal/client"
)

func TestAccSensorMetadataResource(t *testing.T) {
	t.Parallel()

	testSensor := &client.Sensor{
		ID:   "1d6aed11-f2de-48f9-9526-8fb72be10700",
		Name: "Gifted Trout",
		PublicIps: []string{
			"159.223.200.217",
		},
		Persona: "501c5e5a-cf2e-4401-844a-04d4391b1332",
		Metadata: client.SensorMetadata{
			Items: []client.SensorMetadatum{
				{
					Access: client.MetadataAccessReadonly,
					Name:   "provider",
					Val:    "greynoise",
				},
			},
		},
		Status:    "healthy",
		Disabled:  false,
		LastSeen:  time.Date(2024, 8, 27, 16, 27, 2, 0, time.UTC),
		CreatedAt: time.Date(2024, 8, 10, 3, 2, 22, 0, time.UTC),
		UpdatedAt: time.Date(2024, 8, 26, 13, 53, 07, 0, time.UTC),
	}

	mockServer := defaultMockAPIServer()
	mockWorkspaceID := mockServer.Account.WorkspaceID.String()
	mockAPIKey := mockServer.APIKey

	mockServer.Register(http.MethodPut,
		fmt.Sprintf("/v1/workspaces/%s/sensors/%s", mockWorkspaceID, testSensor.ID),
		http.StatusAccepted,
		emptyBody,
		func(r *http.Request) {
			var req client.SensorUpdateRequest
			_ = json.NewDecoder(r.Body).Decode(&req)

			if req.Name != "" {
				testSensor.Name = req.Name
			}

			if req.Metadata != nil {
				testSensor.Metadata = *req.Metadata
			}
		},
	)
	mockServer.Register(http.MethodGet,
		fmt.Sprintf("/v1/workspaces/%s/sensors/%s", mockWorkspaceID, testSensor.ID),
		http.StatusOK,
		body(testSensor),
		nil,
	)

	server := mockServer.Server()

	type step struct {
		config      string
		check       resource.TestCheckFunc
		planChecks  resource.ConfigPlanChecks
		expectError *regexp.Regexp
	}

	testCases := []struct {
		name  string
		steps []step
	}{
		{
			name: "success - create",
			steps: []step{
				{

					config: `resource "greynoise_sensor_metadata" "this" {
					  sensor_id = "1d6aed11-f2de-48f9-9526-8fb72be10700"
					  name = "Angry Cuscus"
					}`,
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("greynoise_sensor_metadata.this", "sensor_id",
							"1d6aed11-f2de-48f9-9526-8fb72be10700",
						),
						resource.TestCheckResourceAttr("greynoise_sensor_metadata.this", "name",
							"Angry Cuscus",
						),
					),
				},
			},
		},
		{
			name: "success - update plan",
			steps: []step{
				{

					config: `resource "greynoise_sensor_metadata" "this" {
					  sensor_id = "1d6aed11-f2de-48f9-9526-8fb72be10700"
					  name = "Angry Cuscus"	
					}`,
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("greynoise_sensor_metadata.this", "sensor_id",
							"1d6aed11-f2de-48f9-9526-8fb72be10700",
						),
						resource.TestCheckResourceAttr("greynoise_sensor_metadata.this", "name",
							"Angry Cuscus",
						),
					),
				},
				{
					config: `resource "greynoise_sensor_metadata" "this" {
						sensor_id = "1d6aed11-f2de-48f9-9526-8fb72be10700"
						name = "Angry Alligator"
						//metadata = [
						//  {
						//    name = "type"
						//    value = "terraform"
						//  }
						//]
					}`,
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("greynoise_sensor_metadata.this", "sensor_id",
							"1d6aed11-f2de-48f9-9526-8fb72be10700",
						),
						resource.TestCheckResourceAttr("greynoise_sensor_metadata.this", "name",
							"Angry Alligator",
						),
						/*resource.TestCheckResourceAttr("greynoise_sensor_metadata.this", "metadata.0.name",
							"type",
						),
						resource.TestCheckResourceAttr("greynoise_sensor_metadata.this", "metadata.0.value",
							"terraform",
						),*/
					),
					planChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("greynoise_sensor_metadata.this",
								plancheck.ResourceActionUpdate),
						},
					},
				},
			},
		},
		/*{
			name: "failure - empty",
			steps: []step{
				{

					config: `resource "greynoise_sensor_metadata" "this" {
				       sensor_id = "1d6aed11-f2de-48f9-9526-8fb72be10700"
				    }`,
					expectError: regexp.MustCompile(`At least one of these attributes must be configured: \[name,metadata\]`),
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("greynoise_sensor_metadata.this", "sensor_id",
							"1d6aed11-f2de-48f9-9526-8fb72be10700",
						),
					),
				},
			},
		},*/
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			testCaseSteps := make([]resource.TestStep, len(tc.steps))
			for i, step := range tc.steps {
				testCaseSteps[i] = resource.TestStep{
					Config: fmt.Sprintf(`
				provider "greynoise" {
				base_url = "%s"
				api_key = "%s"
				}
				`, server.URL, mockAPIKey) + step.config,
					Check:            step.check,
					ConfigPlanChecks: step.planChecks,
					ExpectError:      step.expectError,
				}
			}

			resource.Test(t, resource.TestCase{
				ProtoV6ProviderFactories: testAccProtoV6ProviderFactories,
				Steps:                    testCaseSteps,
			})
		})
	}
}

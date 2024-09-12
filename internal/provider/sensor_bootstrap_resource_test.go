package provider

import (
	"fmt"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
)

func TestAccSensorBootstrapResource(t *testing.T) {
	t.Parallel()

	mockServer := defaultMockAPIServer()
	mockWorkspaceID := mockServer.Account.WorkspaceID.String()
	mockAPIKey := mockServer.APIKey

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
			name: "success - min parameters",
			steps: []step{
				{
					config: `
					resource "greynoise_sensor_bootstrap" "this" {
					  public_ip = "185.108.182.240"
					}`,
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "setup_script",
							fmt.Sprintf("echo %s > ~/.greynoise.key", mockAPIKey),
						),
						resource.TestCheckResourceAttrWith("greynoise_sensor_bootstrap.this", "bootstrap_script",
							checkBootstrapScriptFunc(server.URL, mockWorkspaceID, "185.108.182.240",
								nil, nil, false),
						),
						resource.TestCheckResourceAttrWith("greynoise_sensor_bootstrap.this", "ssh_port_selected",
							checkAutoSelectedSSHPort,
						),
					),
				},
			},
		},
		{
			name: "success - more parameters",
			steps: []step{
				{
					config: `
					resource "greynoise_sensor_bootstrap" "this" {
					  public_ip   = "179.108.182.240"
					  internal_ip = "172.108.182.240"
					  ssh_port    = 2000
					  nat         = true
					  triggers     = {
						public_ip = "179.108.182.240"
					  }
					}`,
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "setup_script",
							fmt.Sprintf("echo %s > ~/.greynoise.key", mockAPIKey),
						),
						resource.TestCheckResourceAttrWith("greynoise_sensor_bootstrap.this", "bootstrap_script",
							checkBootstrapScriptFunc(server.URL, mockWorkspaceID, "179.108.182.240",
								strRef("172.108.182.240"), intRef(2000), true),
						),
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "unbootstrap_script",
							fmt.Sprintf(`SENSOR_ID=$(cat /opt/greynoise/sensor.id) KEY=$(cat ~/.greynoise.key) && \
curl -H "key: $KEY" -X DELETE -L %s/v1/workspaces/%s/sensors/$SENSOR_ID && \
curl -H "key: $KEY" -L %s/v1/workspaces/%s/sensors/unbootstrap/script | sudo bash -s --`,
								server.URL, mockWorkspaceID,
								server.URL, mockWorkspaceID)),
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "ssh_port_selected",
							"2000"),
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "triggers.public_ip",
							"179.108.182.240"),
					),
				},
			},
		},
		{
			name: "success - update triggers",
			steps: []step{
				{
					config: `
					resource "greynoise_sensor_bootstrap" "this" {
					  public_ip   = "179.108.182.240"	
					  internal_ip = "172.108.182.240"	
					}`,
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "setup_script",
							fmt.Sprintf("echo %s > ~/.greynoise.key", mockAPIKey),
						),
						resource.TestCheckResourceAttrWith("greynoise_sensor_bootstrap.this", "bootstrap_script",
							checkBootstrapScriptFunc(server.URL, mockWorkspaceID, "179.108.182.240",
								strRef("172.108.182.240"), nil, false),
						),
					),
				},
				{
					config: `
					resource "greynoise_sensor_bootstrap" "this" {
					  public_ip   = "134.108.182.240"	
					  internal_ip = "172.108.182.240"
					  triggers    = {
					    internal_ip = "172.108.182.240"
					  }
					}`,
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "setup_script",
							fmt.Sprintf("echo %s > ~/.greynoise.key", mockAPIKey),
						),
						resource.TestCheckResourceAttrWith("greynoise_sensor_bootstrap.this", "bootstrap_script",
							checkBootstrapScriptFunc(server.URL, mockWorkspaceID, "134.108.182.240",
								strRef("172.108.182.240"), nil, false),
						),
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "triggers.internal_ip",
							"172.108.182.240"),
					),
					planChecks: resource.ConfigPlanChecks{
						PreApply: []plancheck.PlanCheck{
							plancheck.ExpectResourceAction("greynoise_sensor_bootstrap.this",
								plancheck.ResourceActionReplace),
						},
					},
				},
			},
		},
		{
			name: "missing public IP",
			steps: []step{
				{
					config:      `resource "greynoise_sensor_bootstrap" "this" {}`,
					expectError: regexp.MustCompile(`The argument "public_ip" is required, but no definition was found.`),
				},
			},
		},
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
						  api_key  = "%s"
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

func checkBootstrapScriptFunc(serverURL, workspaceID, publicIP string,
	internalIP *string, sshPort *int, nat bool) resource.CheckResourceAttrWithFunc {
	scriptStart := fmt.Sprintf(`KEY=$(cat ~/.greynoise.key) && \
curl -H "key: $KEY" -L %s/v1/workspaces/%s/sensors/bootstrap/script | sudo bash -s -- -k $KEY -p %s`,
		serverURL, workspaceID, publicIP)

	if internalIP != nil {
		scriptStart += fmt.Sprintf(" -i %s", *internalIP)
	}

	if sshPort != nil {
		scriptStart += fmt.Sprintf(" -s %d", *sshPort)
	} else {
		scriptStart += " -s"
	}

	if nat {
		scriptStart += " -t"
	}

	return func(value string) error {
		if !strings.HasPrefix(value, scriptStart) {
			return fmt.Errorf("did not match bootstrap expected start: expected: %s, got: %s", scriptStart, value)
		}

		if sshPort == nil {
			sshPortStr, _ := strings.CutPrefix(value, scriptStart)
			return checkAutoSelectedSSHPort(strings.TrimSpace(sshPortStr))
		}

		return nil
	}
}

func checkAutoSelectedSSHPort(value string) error {
	sshPort, err := strconv.Atoi(value)
	if err != nil {
		return fmt.Errorf("ssh port conv: %w", err)
	}

	if sshPort < SSHPortMin || sshPort > SSHPortMax {
		return fmt.Errorf("SSH port not within range: [%d-%d]", SSHPortMin, SSHPortMax)
	}

	return nil
}

func strRef(s string) *string {
	return &s
}

func intRef(i int) *int {
	return &i
}

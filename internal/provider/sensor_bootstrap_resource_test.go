package provider

import (
	"fmt"
	"net"
	"regexp"
	"strconv"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-testing/helper/resource"
	"github.com/hashicorp/terraform-plugin-testing/plancheck"
	"github.com/stretchr/testify/assert"
)

func TestDeterministicSSHPort(t *testing.T) {
	testCases := []struct {
		name string
		ip   net.IP
		port int32
	}{
		{
			name: "case-1",
			ip:   net.ParseIP("185.108.182.240"),
			port: 62914,
		},
		{
			name: "case-2",
			ip:   net.ParseIP("179.108.182.240"),
			port: 58026,
		},
		{
			name: "case-3",
			ip:   net.ParseIP("79.172.244.248"),
			port: 61609,
		},
		{
			name: "case-4",
			ip:   net.ParseIP("39.101.187.33"),
			port: 62864,
		},
		{
			name: "case-5",
			ip:   net.ParseIP("176.97.114.156"),
			port: 56988,
		},
	}

	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.port, DeterministicSSHPort(tc.ip))
		})
	}
}

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
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "ssh_port_selected",
							"62914"),
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
					  public_ip   = "179.108.182.240/32,172.108.182.241/32"
					  internal_ip = "172.108.182.240"
					  ssh_port    = 2000
					  nat         = true
					  config      = {
						public_ip = "179.108.182.240/32"
					  }
					}`,
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "setup_script",
							fmt.Sprintf("echo %s > ~/.greynoise.key", mockAPIKey),
						),
						resource.TestCheckResourceAttrWith("greynoise_sensor_bootstrap.this", "bootstrap_script",
							checkBootstrapScriptFunc(server.URL, mockWorkspaceID, "179.108.182.240/32,172.108.182.241/32",
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
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "config.public_ip",
							"179.108.182.240/32"),
					),
				},
			},
		},
		{
			name: "success - update",
			steps: []step{
				{
					config: `
					resource "greynoise_sensor_bootstrap" "this" {
					  public_ips  = ["179.108.182.240"]	
					  internal_ip = "172.108.182.240"	
					}`,
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "setup_script",
							fmt.Sprintf("echo %s > ~/.greynoise.key", mockAPIKey),
						),
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "ssh_port_selected",
							"58026"),
						resource.TestCheckResourceAttrWith("greynoise_sensor_bootstrap.this", "bootstrap_script",
							checkBootstrapScriptFunc(server.URL, mockWorkspaceID, "179.108.182.240",
								strRef("172.108.182.240"), nil, false),
						),
					),
				},
				{
					config: `
					resource "greynoise_sensor_bootstrap" "this" {
					  public_ips  = ["136.108.182.240"]	
					  internal_ip = "172.108.182.240"
					  config      = {
					    internal_ip = "172.108.182.240"
					  }
					}`,
					check: resource.ComposeAggregateTestCheckFunc(
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "setup_script",
							fmt.Sprintf("echo %s > ~/.greynoise.key", mockAPIKey),
						),
						resource.TestCheckResourceAttrWith("greynoise_sensor_bootstrap.this", "bootstrap_script",
							checkBootstrapScriptFunc(server.URL, mockWorkspaceID, "136.108.182.240",
								strRef("172.108.182.240"), nil, false),
						),
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "ssh_port_selected",
							"59041",
						),
						resource.TestCheckResourceAttr("greynoise_sensor_bootstrap.this", "config.internal_ip",
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
			name: "missing public IP fields",
			steps: []step{
				{
					config:      `resource "greynoise_sensor_bootstrap" "this" {}`,
					expectError: regexp.MustCompile(`At least one of these attributes must be configured: \[public_ip,public_ips\]`),
				},
			},
		},
		{
			name: "invalid public IP",
			steps: []step{
				{
					config: `resource "greynoise_sensor_bootstrap" "this" {
					   public_ip = "invalid_ip"
					}`,
					expectError: regexp.MustCompile(`Error occurred while parsing IPs: invalid CIDR address: invalid_ip`),
				},
			},
		},
		{
			name: "invalid public IPs",
			steps: []step{
				{
					config: `resource "greynoise_sensor_bootstrap" "this" {
					   public_ips = ["1.1.1.1/32", "1.1.1.1", "invalid_ip"]
					}`,
					expectError: regexp.MustCompile(`Error occurred while parsing IPs: invalid CIDR address: invalid_ip`),
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

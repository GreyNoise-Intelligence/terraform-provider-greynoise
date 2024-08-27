package client

import (
	"time"
)

type Pagination struct {
	Page       int32 `json:"page"`
	PageSize   int32 `json:"page_size"`
	TotalItems int32 `json:"total_items"`
}

type PersonaSearchFilters struct {
	Workspace  string `mapstructure:"workspace"`
	Tiers      string `mapstructure:"tiers"`
	Categories string `mapstructure:"categories"`
	Protocols  string `mapstructure:"protocols"`
	Search     string `mapstructure:"search"`
	PageSize   int32  `mapstructure:"page_size"`
}

func (f *PersonaSearchFilters) Validate() error {
	if f.Workspace == "" {
		return NewErrMissingField("workspace")
	}

	return nil
}

type Persona struct {
	ID                        string    `json:"id"`
	CreatedAt                 time.Time `json:"created_at"`
	UpdatedAt                 time.Time `json:"updated_at"`
	Name                      string    `json:"name"`
	Author                    string    `json:"author"`
	ArtifactLink              string    `json:"artifact_link"`
	Tier                      string    `json:"tier"`
	InstanceManagement        string    `json:"instance_management"`
	Workspace                 string    `json:"workspace"`
	Categories                []string  `json:"categories"`
	Description               string    `json:"description"`
	ApplicationProtocols      []string  `json:"application_protocols"`
	Ports                     []int32   `json:"ports"`
	OperatingSystem           string    `json:"operating_system"`
	Icon                      string    `json:"icon"`
	AssociatedVulnerabilities []string  `json:"associated_vulnerabilities"`
}

type PersonaSearchResponse struct {
	Items      []Persona  `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type Sensor struct {
	ID        string    `json:"sensor_id"`
	Name      string    `json:"name"`
	PublicIps []string  `json:"public_ips"`
	Persona   string    `json:"persona"`
	Status    string    `json:"status"`
	Disabled  bool      `json:"disabled"`
	LastSeen  time.Time `json:"last_seen"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

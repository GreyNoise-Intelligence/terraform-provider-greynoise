package client

import (
	"time"
)

type (
	SensorSortBy   string
	MetadataAccess string
)

const (
	SensorSortByName      = SensorSortBy("name")
	SensorSortByIP        = SensorSortBy("public_ips")
	SensorSortByPort      = SensorSortBy("access_port")
	SensorSortByPersona   = SensorSortBy("persona_name")
	SensorSortByCreatedAt = SensorSortBy("created_at")
	SensorSortByStatus    = SensorSortBy("status")

	MetadataAccessReadWrite = MetadataAccess("readwrite") // User readable and modifiable
	MetadataAccessReadonly  = MetadataAccess("readonly")  // User readable, not modifiable
	MetadataAccessHidden    = MetadataAccess("hidden")
)

type Pagination struct {
	Page       int32 `json:"page"`
	PageSize   int32 `json:"page_size"`
	TotalItems int32 `json:"total_items"`
}

type SensorSearchFilter struct {
	Filter     string       `mapstructure:"filter"`
	Page       int32        `mapstructure:"page"`
	PageSize   int32        `mapstructure:"page_size"`
	SortBy     SensorSortBy `mapstructure:"sort_by"`
	Descending bool         `mapstructure:"descending"`
}

func (f *SensorSearchFilter) Validate() error {
	switch f.SortBy {
	case SensorSortByStatus, SensorSortByCreatedAt, SensorSortByIP, SensorSortByName, SensorSortByPersona,
		SensorSortByPort:
	default:
		return NewErrInvalidField("sort_by", "unknown")
	}

	if f.Filter == "" {
		return NewErrMissingField("filter")
	}

	return nil
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
	OperatingSystem           string    `json:"operating_system"`
	Icon                      string    `json:"icon"`
	AssociatedVulnerabilities []string  `json:"associated_vulnerabilities"`
}

type PersonaSearchResponse struct {
	Items      []Persona  `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type SensorSearchResponse struct {
	Items      []Sensor   `json:"items"`
	Pagination Pagination `json:"pagination"`
}

type SensorUpdateRequest struct {
	Name     string          `json:"name,omitempty"`
	Persona  string          `json:"persona,omitempty"`
	Metadata *SensorMetadata `json:"metadata,omitempty"`
}

type SensorMetadata struct {
	Items []SensorMetadatum `json:"items"`
}

type SensorMetadatum struct {
	Access MetadataAccess `json:"access"`
	Name   string         `json:"name"`
	Val    string         `json:"val"`
}

func (s *SensorMetadatum) Validate() error {
	switch s.Access {
	case MetadataAccessHidden, MetadataAccessReadWrite, MetadataAccessReadonly:
	default:
		return NewErrInvalidField("access", "unknown")
	}

	return nil
}

type Sensor struct {
	ID         string         `json:"sensor_id"`
	Name       string         `json:"name"`
	PublicIps  []string       `json:"public_ips"`
	AccessPort int32          `json:"access_port"`
	Persona    string         `json:"persona"`
	Metadata   SensorMetadata `json:"metadata"`
	Status     string         `json:"status"`
	Disabled   bool           `json:"disabled"`
	LastSeen   time.Time      `json:"last_seen"`
	CreatedAt  time.Time      `json:"created_at"`
	UpdatedAt  time.Time      `json:"updated_at"`
}

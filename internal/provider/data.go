package provider

import (
	"github.com/google/uuid"

	"github.com/GreyNoise-Intelligence/terraform-provider-greynoise/internal/client"
)

var (
	// NameSpaceGN is a UUID generated using the following code:
	// uuid.NewMD5(uuid.NameSpaceDNS, []byte("greynoise.io"))
	NameSpaceGN = uuid.MustParse("b75f8841-d673-3887-919c-4af86b04f9ce")
)

type Data struct {
	Client        *client.GreyNoiseClient
	APIKey        string
	UUIDGenerator UUIDGenerator
}

// UUIDGenerator generates a V3 UUID from a given string.
type UUIDGenerator struct{}

// Generate generates a V3 UUID from given data.
func (m *UUIDGenerator) Generate(bytes []byte) string {
	return uuid.NewMD5(NameSpaceGN, bytes).String()
}

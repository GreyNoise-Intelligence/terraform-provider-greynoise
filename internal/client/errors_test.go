package client_test

import (
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/GreyNoise-Intelligence/terraform-provider-greynoise/internal/client"
)

func TestNewErrUnexpectedStatusCode(t *testing.T) {
	var statusCodeErr *client.ErrUnexpectedStatusCode

	err := client.NewErrUnexpectedStatusCode(http.StatusOK, http.StatusInternalServerError)
	if errors.As(err, &statusCodeErr) {
		assert.Equal(t, statusCodeErr.StatusCode(), http.StatusInternalServerError)
	} else {
		t.Fatalf("Unable to convert to typed error")
	}
}

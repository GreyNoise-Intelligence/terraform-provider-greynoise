//go:generate mockgen -source=dependencies.go -destination=./mocks.go -package=client
package client

import "net/http"

type HTTPClient interface {
	Do(req *http.Request) (*http.Response, error)
}

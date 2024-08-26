package client

import (
	"net/url"
)

// Option is used to configure the GreyNoiseClient.
type Option func(client *GreyNoiseClient)

// WithBaseURL is used to set the base URL.
func WithBaseURL(url *url.URL) Option {
	return func(client *GreyNoiseClient) {
		client.baseURL = url
	}
}

// WithHTTPClient is used to set the internal HTTP client.
func WithHTTPClient(httpClient HTTPClient) Option {
	return func(client *GreyNoiseClient) {
		client.httpClient = httpClient
	}
}

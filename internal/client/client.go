package client

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/hashicorp/go-retryablehttp"
	"github.com/mitchellh/mapstructure"
)

const (
	HeaderKey = "key"

	retryCount                   = 3
	defaultPersonaSearchPageSize = int32(100)
)

var (
	defaultBaseURL = &url.URL{
		Host:   "api.greynoise.io",
		Scheme: "https",
	}
)

// GreyNoiseClient is a thin wrapper for a HTTP client.
type GreyNoiseClient struct {
	baseURL     *url.URL
	apiKey      string
	workspaceID string
	httpClient  HTTPClient
}

// New is the preferred way to instantiate the GreyNoiseClient.
func New(apiKey string, options ...Option) (*GreyNoiseClient, error) {
	client := &GreyNoiseClient{
		apiKey: apiKey,
	}

	for _, option := range options {
		option(client)
	}

	if client.baseURL == nil {
		client.baseURL = defaultBaseURL
	}

	if client.httpClient == nil {
		retryClient := retryablehttp.NewClient()
		retryClient.RetryMax = retryCount
		retryClient.Logger = nil

		httpClient := retryClient.StandardClient()
		httpClient.Timeout = time.Second * 30

		client.httpClient = httpClient
	}

	return client, nil
}

func (c *GreyNoiseClient) Ping() error {
	u := c.baseURL.ResolveReference(&url.URL{Path: "/ping"})

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return err
	}

	c.setAuthHeader(req)
	c.setJSONContentHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusOK {
		return NewErrUnexpectedStatusCode(http.StatusOK, resp.StatusCode)
	}

	return err
}

func (c *GreyNoiseClient) SensorBootstrapScript() (*string, error) {
	return nil, nil
}

func (c *GreyNoiseClient) PersonasSearch(filters PersonaSearchFilters) (*PersonaSearchResponse, error) {
	if err := filters.Validate(); err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(&url.URL{Path: "/v1/personas"})
	q := u.Query()

	if filters.PageSize == 0 {
		filters.PageSize = defaultPersonaSearchPageSize
	}

	var filterParameters map[string]interface{}
	err := mapstructure.Decode(filters, &filterParameters)
	if err != nil {
		return nil, err
	}

	for k, v := range filterParameters {
		vStr := fmt.Sprintf("%v", v)
		if vStr != "" {
			q.Set(k, vStr)
		}
	}

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	req.URL.RawQuery = q.Encode()
	c.setAuthHeader(req)
	c.setJSONContentHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, NewErrUnexpectedStatusCode(http.StatusOK, resp.StatusCode)
	}

	var result PersonaSearchResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *GreyNoiseClient) GetPersona(id string) (*Persona, error) {
	u := c.baseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/personas/%s", id)})

	req, err := http.NewRequest("GET", u.String(), nil)
	if err != nil {
		return nil, err
	}

	c.setAuthHeader(req)
	c.setJSONContentHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode != http.StatusOK {
		return nil, NewErrUnexpectedStatusCode(http.StatusOK, resp.StatusCode)
	}

	var result Persona
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *GreyNoiseClient) SensorBootstrapURL() *url.URL {
	return c.baseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/v1/workspaces/%s/sensors/bootstrap/script", c.workspaceID),
	})
}

func (c *GreyNoiseClient) setAuthHeader(req *http.Request) {
	req.Header.Set(HeaderKey, c.apiKey)
}

func (c *GreyNoiseClient) setJSONContentHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
}

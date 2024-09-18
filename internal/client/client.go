package client

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/google/uuid"
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
	baseURL    *url.URL
	apiKey     string
	account    Account
	httpClient HTTPClient
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

	acct, err := client.getAccount()
	if err != nil {
		return nil, fmt.Errorf("account error: %w", err)
	}

	client.account = *acct

	return client, nil
}

func (c *GreyNoiseClient) WorkspaceID() uuid.UUID {
	return c.account.WorkspaceID
}

func (c *GreyNoiseClient) UserID() uuid.UUID {
	return c.account.UserID
}

type Account struct {
	UserID      uuid.UUID `json:"user_id"`
	WorkspaceID uuid.UUID `json:"workspace_id"`
}

func (c *GreyNoiseClient) getAccount() (*Account, error) {
	u := c.baseURL.ResolveReference(&url.URL{Path: "/v1/account"})

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

	var result Account
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *GreyNoiseClient) GetPersona(ctx context.Context, id string) (*Persona, error) {
	u := c.baseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/personas/%s", id)})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
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
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		}

		return nil, NewErrUnexpectedStatusCode(http.StatusOK, resp.StatusCode)
	}

	var result Persona
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *GreyNoiseClient) PersonasSearch(ctx context.Context, filters PersonaSearchFilters) (*PersonaSearchResponse, error) {
	filters.Workspace = c.WorkspaceID().String()
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

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
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

func (c *GreyNoiseClient) GetSensor(ctx context.Context, id string) (*Sensor, error) {
	u := c.baseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/workspaces/%s/sensors/%s",
		c.WorkspaceID(), id)})

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
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
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		}

		return nil, NewErrUnexpectedStatusCode(http.StatusOK, resp.StatusCode)
	}

	var result Sensor
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}

func (c *GreyNoiseClient) UpdateSensor(ctx context.Context, id string, request SensorUpdateRequest) error {
	u := c.baseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/workspaces/%s/sensors/%s",
		c.WorkspaceID(), id)})

	body, err := json.Marshal(request)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, "PUT", u.String(), bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	c.setAuthHeader(req)
	c.setJSONContentHeaders(req)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return err
	}

	if resp.StatusCode != http.StatusAccepted {
		if resp.StatusCode == http.StatusNotFound {
			return ErrNotFound
		}

		return NewErrUnexpectedStatusCode(http.StatusAccepted, resp.StatusCode)
	}

	return nil
}

func (c *GreyNoiseClient) SensorsSearch(ctx context.Context, filters SensorSearchFilter) (*SensorSearchResponse, error) {
	if filters.SortBy == "" {
		filters.SortBy = SensorSortByCreatedAt
	}

	if err := filters.Validate(); err != nil {
		return nil, err
	}

	u := c.baseURL.ResolveReference(&url.URL{Path: fmt.Sprintf("/v1/workspaces/%s/sensors", c.WorkspaceID())})
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

	req, err := http.NewRequestWithContext(ctx, "GET", u.String(), nil)
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
		if resp.StatusCode == http.StatusNotFound {
			return nil, ErrNotFound
		}

		return nil, NewErrUnexpectedStatusCode(http.StatusOK, resp.StatusCode)
	}

	var result SensorSearchResponse
	if err = json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, err
	}

	return &result, nil
}
func (c *GreyNoiseClient) SensorBootstrapURL() *url.URL {
	return c.baseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/v1/workspaces/%s/sensors/bootstrap/script", c.WorkspaceID()),
	})
}

func (c *GreyNoiseClient) SensorUnBootstrapURL() *url.URL {
	return c.baseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/v1/workspaces/%s/sensors/unbootstrap/script", c.WorkspaceID()),
	})
}

func (c *GreyNoiseClient) SensorsURL() *url.URL {
	return c.baseURL.ResolveReference(&url.URL{
		Path: fmt.Sprintf("/v1/workspaces/%s/sensors", c.WorkspaceID()),
	})
}

func (c *GreyNoiseClient) setAuthHeader(req *http.Request) {
	req.Header.Set(HeaderKey, c.apiKey)
}

func (c *GreyNoiseClient) setJSONContentHeaders(req *http.Request) {
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
}

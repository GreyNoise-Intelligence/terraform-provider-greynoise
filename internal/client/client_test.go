package client_test

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"github.com/GreyNoise-Intelligence/terraform-provider-greynoise/internal/client"
)

func TestGreyNoiseClient_GetPersona(t *testing.T) {
	testAPIKey := "test-42owudoflsahj"
	testPersona := &client.Persona{
		ID:                 "ac65d8a0-ed21-417e-a1a2-65a4e09c3144",
		Name:               "RDP Server",
		Author:             "GreyNoiseIO",
		ArtifactLink:       "ami-0f16b3dc51ce26ae0",
		Tier:               "premium",
		InstanceManagement: "per_gateway",
		Categories: []string{
			"honeypot",
			"rev1",
		},
		Description: "A Remote Desktop Protocol server. Designed to observe credential " +
			"bruteforce activity.",
		ApplicationProtocols: []string{
			"rdp",
		},
		Ports: []int32{
			3389,
		},
	}

	testAccountJSON := `
{
  "user_id": "4c65d8a0-ed21-417e-a1a2-65a4e09c3144",
  "workspace_id": "7c65d8a0-ed21-417e-a1a2-65a4e09c3144"
}`
	testPersonaJSON := `
{
  "id": "ac65d8a0-ed21-417e-a1a2-65a4e09c3144",
  "name": "RDP Server",
  "author": "GreyNoiseIO",
  "artifact_link": "ami-0f16b3dc51ce26ae0",
  "tier": "premium",
  "instance_management": "per_gateway",
  "workspace": "",
  "categories": [
	"honeypot",
	"rev1"
  ],
  "description": "A Remote Desktop Protocol server. Designed to observe credential bruteforce activity.",
  "application_protocols": [
	"rdp"
  ],
  "ports": [
	3389
  ]
}`

	mockAccount := func(t *testing.T, httpClient *client.MockHTTPClient) {
		httpClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, req.Method, http.MethodGet)
				assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
				assert.Equal(t, "https://api.greynoise.io/v1/account", req.URL.String())

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody(testAccountJSON),
				}, nil
			})
	}

	type want struct {
		response *client.Persona
		err      error
	}

	testCases := []struct {
		name   string
		input  string
		expect func(*testing.T, *client.MockHTTPClient)
		want   want
	}{
		{
			name:  "happy path",
			input: "ac65d8a0-ed21-417e-a1a2-65a4e09c3144",
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/personas/"+
							"ac65d8a0-ed21-417e-a1a2-65a4e09c3144", req.URL.String())

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       responseBody(testPersonaJSON),
						}, nil
					})
			},
			want: want{
				response: testPersona,
			},
		},
		{
			name:  "http client error",
			input: "bc65d8a0-ed21-417e-a1a2-65a4e09c3144",
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/personas/"+
							"bc65d8a0-ed21-417e-a1a2-65a4e09c3144", req.URL.String())

						return nil, errors.New("http error")
					})
			},
			want: want{
				err: errors.New("http error"),
			},
		},
		{
			name:  "invalid response body",
			input: "cc65d8a0-ed21-417e-a1a2-65a4e09c3144",
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/personas/"+
							"cc65d8a0-ed21-417e-a1a2-65a4e09c3144", req.URL.String())

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       responseBody(`{`),
						}, nil
					})
			},
			want: want{
				err: errors.New("unexpected EOF"),
			},
		},
		{
			name:  "unexpected status code",
			input: "dc65d8a0-ed21-417e-a1a2-65a4e09c3144",
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/personas/"+
							"dc65d8a0-ed21-417e-a1a2-65a4e09c3144", req.URL.String())

						return &http.Response{
							StatusCode: http.StatusForbidden,
							Body:       responseBody(``),
						}, nil
					})
			},
			want: want{
				err: client.NewErrUnexpectedStatusCode(http.StatusOK, http.StatusForbidden),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := client.NewMockHTTPClient(ctrl)
			mockAccount(t, mockHTTPClient)

			if tc.expect != nil {
				tc.expect(t, mockHTTPClient)
			}

			gClient, err := client.New(testAPIKey, client.WithHTTPClient(mockHTTPClient))
			assert.NoError(t, err)

			response, err := gClient.GetPersona(context.Background(), tc.input)
			assert.Equal(t, tc.want.response, response)
			assert.Equal(t, tc.want.err, err)
		})
	}
}

func TestGreyNoiseClient_PersonasSearch(t *testing.T) {
	testAPIKey := "test-2o3uwofjsldfj"
	testPersona := client.Persona{
		ID:                 "ac65d8a0-ed21-417e-a1a2-65a4e09c3144",
		Name:               "RDP Server",
		Author:             "GreyNoiseIO",
		ArtifactLink:       "ami-0f16b3dc51ce26ae0",
		Tier:               "premium",
		InstanceManagement: "per_gateway",
		Categories: []string{
			"honeypot",
			"rev1",
		},
		Description: "A Remote Desktop Protocol server. Designed to observe credential " +
			"bruteforce activity.",
		ApplicationProtocols: []string{
			"rdp",
		},
		Ports: []int32{
			3389,
		},
	}

	testAccountJSON := `
{
  "user_id": "4c65d8a0-ed21-417e-a1a2-65a4e09c3144",
  "workspace_id": "25443a54-1e10-45e8-8164-c38aa238615e"
}`
	testPersonaJSON := `
{
  "id": "ac65d8a0-ed21-417e-a1a2-65a4e09c3144",
  "name": "RDP Server",
  "author": "GreyNoiseIO",
  "artifact_link": "ami-0f16b3dc51ce26ae0",
  "tier": "premium",
  "instance_management": "per_gateway",
  "workspace": "",
  "categories": [
	"honeypot",
	"rev1"
  ],
  "description": "A Remote Desktop Protocol server. Designed to observe credential bruteforce activity.",
  "application_protocols": [
	"rdp"
  ],
  "ports": [
	3389
  ]
}`

	mockAccount := func(t *testing.T, httpClient *client.MockHTTPClient) {
		httpClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, req.Method, http.MethodGet)
				assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
				assert.Equal(t, "https://api.greynoise.io/v1/account", req.URL.String())

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody(testAccountJSON),
				}, nil
			})
	}

	type want struct {
		response *client.PersonaSearchResponse
		err      error
	}

	testCases := []struct {
		name   string
		input  client.PersonaSearchFilters
		expect func(*testing.T, *client.MockHTTPClient)
		want   want
	}{
		{
			name: "happy path",
			input: client.PersonaSearchFilters{
				Tiers:     "community",
				Protocols: "http",
				Search:    "rdp",
				PageSize:  10,
			},
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/personas?"+
							"page_size=10&protocols=http&search=rdp&tiers=community&"+
							"workspace=25443a54-1e10-45e8-8164-c38aa238615e", req.URL.String())

						return &http.Response{
							StatusCode: http.StatusOK,
							Body: responseBody(`
								{
								  "items": [` + testPersonaJSON + `],
								  "pagination": {
									"page": 0,
									"page_size": 10,
									"total_items": 1
								  }
								}`),
						}, nil
					})
			},
			want: want{
				response: &client.PersonaSearchResponse{
					Items: []client.Persona{
						testPersona,
					},
					Pagination: client.Pagination{
						Page:       0,
						PageSize:   10,
						TotalItems: 1,
					},
				},
			},
		},
		{
			name: "http client error",
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/personas?"+
							"page_size=100&workspace=25443a54-1e10-45e8-8164-c38aa238615e", req.URL.String())

						return nil, errors.New("http error")
					})
			},
			want: want{
				err: errors.New("http error"),
			},
		},
		{
			name: "invalid response body",
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/personas?"+
							"page_size=100&workspace=25443a54-1e10-45e8-8164-c38aa238615e", req.URL.String())

						return &http.Response{
							StatusCode: http.StatusOK,
							Body: responseBody(`
								{
								  "items": [ ` + testPersonaJSON + `],
								  "pagination": {
									"page": 0,
									"page_size": 100,
									"total_items": 1
                                `),
						}, nil
					})
			},
			want: want{
				err: errors.New("unexpected EOF"),
			},
		},
		{
			name: "unexpected status code",
			input: client.PersonaSearchFilters{
				Workspace: "65443a54-1e10-45e8-8164-c38aa238615e",
			},
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/personas?"+
							"page_size=100&workspace=25443a54-1e10-45e8-8164-c38aa238615e", req.URL.String())

						return &http.Response{
							StatusCode: http.StatusForbidden,
							Body:       responseBody(``),
						}, nil
					})
			},
			want: want{
				err: client.NewErrUnexpectedStatusCode(http.StatusOK, http.StatusForbidden),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := client.NewMockHTTPClient(ctrl)
			mockAccount(t, mockHTTPClient)

			if tc.expect != nil {
				tc.expect(t, mockHTTPClient)
			}

			gClient, err := client.New(testAPIKey, client.WithHTTPClient(mockHTTPClient))
			assert.NoError(t, err)

			response, err := gClient.PersonasSearch(context.Background(), tc.input)
			assert.Equal(t, tc.want.response, response)
			assert.Equal(t, tc.want.err, err)
		})
	}
}

func TestGreyNoiseClient_GetSensor(t *testing.T) {
	testAPIKey := "test-2037403284"
	testSensor := &client.Sensor{
		ID:   "1d6aed11-f2de-48f9-9526-8fb72be10700",
		Name: "Gifted Trout",
		PublicIps: []string{
			"159.223.200.217",
		},
		Persona:   "501c5e5a-cf2e-4401-844a-04d4391b1332",
		Status:    "healthy",
		Disabled:  false,
		LastSeen:  time.Date(2024, 8, 27, 16, 27, 2, 0, time.UTC),
		CreatedAt: time.Date(2024, 8, 10, 3, 2, 22, 0, time.UTC),
		UpdatedAt: time.Date(2024, 8, 26, 13, 53, 07, 0, time.UTC),
	}

	testAccountJSON := `
{
  "user_id": "4c65d8a0-ed21-417e-a1a2-65a4e09c3144",
  "workspace_id": "7c65d8a0-ed21-417e-a1a2-65a4e09c3144"
}`
	testSensorJSON := `
{
  "sensor_id": "1d6aed11-f2de-48f9-9526-8fb72be10700",
  "name": "Gifted Trout",
  "public_ips": [
    "159.223.200.217"
  ],
  "persona": "501c5e5a-cf2e-4401-844a-04d4391b1332",
  "status": "healthy",
  "disabled": false,
  "last_seen": "2024-08-27T16:27:02Z",
  "created_at": "2024-08-10T03:02:22Z",
  "updated_at": "2024-08-26T13:53:07Z"
}`

	mockAccount := func(t *testing.T, httpClient *client.MockHTTPClient) {
		httpClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, req.Method, http.MethodGet)
				assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
				assert.Equal(t, "https://api.greynoise.io/v1/account", req.URL.String())

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody(testAccountJSON),
				}, nil
			})
	}

	type want struct {
		response *client.Sensor
		err      error
	}

	testCases := []struct {
		name   string
		input  string
		expect func(*testing.T, *client.MockHTTPClient)
		want   want
	}{
		{
			name:  "happy path",
			input: "ac65d8a0-ed21-417e-a1a2-65a4e09c3144",
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/workspaces/"+
							"7c65d8a0-ed21-417e-a1a2-65a4e09c3144/sensors/ac65d8a0-ed21-417e-a1a2-65a4e09c3144",
							req.URL.String())

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       responseBody(testSensorJSON),
						}, nil
					})
			},
			want: want{
				response: testSensor,
			},
		},
		{
			name:  "http client error",
			input: "bc65d8a0-ed21-417e-a1a2-65a4e09c3144",
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/workspaces/"+
							"7c65d8a0-ed21-417e-a1a2-65a4e09c3144/sensors/bc65d8a0-ed21-417e-a1a2-65a4e09c3144",
							req.URL.String())

						return nil, errors.New("http error")
					})
			},
			want: want{
				err: errors.New("http error"),
			},
		},
		{
			name:  "invalid response body",
			input: "cc65d8a0-ed21-417e-a1a2-65a4e09c3144",
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/workspaces/"+
							"7c65d8a0-ed21-417e-a1a2-65a4e09c3144/sensors/cc65d8a0-ed21-417e-a1a2-65a4e09c3144",
							req.URL.String())

						return &http.Response{
							StatusCode: http.StatusOK,
							Body:       responseBody(`{`),
						}, nil
					})
			},
			want: want{
				err: errors.New("unexpected EOF"),
			},
		},
		{
			name:  "unexpected status code",
			input: "dc65d8a0-ed21-417e-a1a2-65a4e09c3144",
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/workspaces/"+
							"7c65d8a0-ed21-417e-a1a2-65a4e09c3144/sensors/dc65d8a0-ed21-417e-a1a2-65a4e09c3144",
							req.URL.String())

						return &http.Response{
							StatusCode: http.StatusForbidden,
							Body:       responseBody(``),
						}, nil
					})
			},
			want: want{
				err: client.NewErrUnexpectedStatusCode(http.StatusOK, http.StatusForbidden),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := client.NewMockHTTPClient(ctrl)
			mockAccount(t, mockHTTPClient)

			if tc.expect != nil {
				tc.expect(t, mockHTTPClient)
			}

			gClient, err := client.New(testAPIKey, client.WithHTTPClient(mockHTTPClient))
			assert.NoError(t, err)

			response, err := gClient.GetSensor(context.Background(), tc.input)
			assert.Equal(t, tc.want.response, response)
			assert.Equal(t, tc.want.err, err)
		})
	}
}

func TestGreyNoiseClient_SensorSearch(t *testing.T) {
	testAPIKey := "test-4o3uwofjsldfj"
	testSensor := client.Sensor{
		ID:   "2d6aed11-f2de-48f9-9526-8fb72be10700",
		Name: "Gifted Trout",
		PublicIps: []string{
			"159.223.200.217",
		},
		Persona:   "601c5e5a-cf2e-4401-844a-04d4391b1332",
		Status:    "healthy",
		Disabled:  false,
		LastSeen:  time.Date(2024, 8, 27, 16, 27, 2, 0, time.UTC),
		CreatedAt: time.Date(2024, 8, 10, 3, 2, 22, 0, time.UTC),
		UpdatedAt: time.Date(2024, 8, 26, 13, 53, 07, 0, time.UTC),
	}

	testAccountJSON := `
{
  "user_id": "8c65d8a0-ed21-417e-a1a2-65a4e09c3144",
  "workspace_id": "5c65d8a0-ed21-417e-a1a2-65a4e09c3144"
}`
	testSensorJSON := `
{
  "sensor_id": "2d6aed11-f2de-48f9-9526-8fb72be10700",
  "name": "Gifted Trout",
  "public_ips": [
    "159.223.200.217"
  ],
  "persona": "601c5e5a-cf2e-4401-844a-04d4391b1332",
  "status": "healthy",
  "disabled": false,
  "last_seen": "2024-08-27T16:27:02Z",
  "created_at": "2024-08-10T03:02:22Z",
  "updated_at": "2024-08-26T13:53:07Z"
}`

	mockAccount := func(t *testing.T, httpClient *client.MockHTTPClient) {
		httpClient.EXPECT().
			Do(gomock.Any()).
			DoAndReturn(func(req *http.Request) (*http.Response, error) {
				assert.Equal(t, req.Method, http.MethodGet)
				assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
				assert.Equal(t, "https://api.greynoise.io/v1/account", req.URL.String())

				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       responseBody(testAccountJSON),
				}, nil
			})
	}

	type want struct {
		response *client.SensorSearchResponse
		err      error
	}

	testCases := []struct {
		name   string
		input  client.SensorSearchFilter
		expect func(*testing.T, *client.MockHTTPClient)
		want   want
	}{
		{
			name: "happy path",
			input: client.SensorSearchFilter{
				Filter:   "Gifted",
				PageSize: 10,
			},
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/workspaces/"+
							"5c65d8a0-ed21-417e-a1a2-65a4e09c3144/sensors?descending=false&"+
							"filter=Gifted&page=0&page_size=10&sort_by=created_at", req.URL.String())

						return &http.Response{
							StatusCode: http.StatusOK,
							Body: responseBody(`
								{
								  "items": [` + testSensorJSON + `],
								  "pagination": {
									"page": 0,
									"page_size": 10,
									"total_items": 1
								  }
								}`),
						}, nil
					})
			},
			want: want{
				response: &client.SensorSearchResponse{
					Items: []client.Sensor{
						testSensor,
					},
					Pagination: client.Pagination{
						Page:       0,
						PageSize:   10,
						TotalItems: 1,
					},
				},
			},
		},
		{
			name: "missing filter",
			want: want{
				err: client.NewErrMissingField("filter"),
			},
		},
		{
			name: "http client error",
			input: client.SensorSearchFilter{
				Filter: "159.223.200.217",
			},
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/workspaces/"+
							"5c65d8a0-ed21-417e-a1a2-65a4e09c3144/sensors?descending=false&"+
							"filter=159.223.200.217&page=0&page_size=100&sort_by=created_at", req.URL.String())

						return nil, errors.New("http error")
					})
			},
			want: want{
				err: errors.New("http error"),
			},
		},
		{
			name: "invalid response body",
			input: client.SensorSearchFilter{
				Filter: "159.223.200.217",
			},
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/workspaces/"+
							"5c65d8a0-ed21-417e-a1a2-65a4e09c3144/sensors?descending=false&"+
							"filter=159.223.200.217&page=0&page_size=100&sort_by=created_at", req.URL.String())

						return &http.Response{
							StatusCode: http.StatusOK,
							Body: responseBody(`
								{
								  "items": [ ` + testSensorJSON + `],
								  "pagination": {
									"page": 0,
									"page_size": 100,
									"total_items": 1
                                `),
						}, nil
					})
			},
			want: want{
				err: errors.New("unexpected EOF"),
			},
		},
		{
			name: "unexpected status code",
			input: client.SensorSearchFilter{
				Filter: "Trout",
			},
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/workspaces/"+
							"5c65d8a0-ed21-417e-a1a2-65a4e09c3144/sensors?descending=false&"+
							"filter=Trout&page=0&page_size=100&sort_by=created_at", req.URL.String())

						return &http.Response{
							StatusCode: http.StatusForbidden,
							Body:       responseBody(``),
						}, nil
					})
			},
			want: want{
				err: client.NewErrUnexpectedStatusCode(http.StatusOK, http.StatusForbidden),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := client.NewMockHTTPClient(ctrl)
			mockAccount(t, mockHTTPClient)

			if tc.expect != nil {
				tc.expect(t, mockHTTPClient)
			}

			gClient, err := client.New(testAPIKey, client.WithHTTPClient(mockHTTPClient))
			assert.NoError(t, err)

			response, err := gClient.SensorsSearch(context.Background(), tc.input)
			assert.Equal(t, tc.want.response, response)
			assert.Equal(t, tc.want.err, err)
		})
	}
}

func responseBody(body string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(body))
}

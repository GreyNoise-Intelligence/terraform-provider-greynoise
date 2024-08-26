package client_test

import (
	"errors"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/golang/mock/gomock"
	"github.com/google/uuid"
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
			gClient, err := client.New(testAPIKey, client.WithHTTPClient(mockHTTPClient))
			assert.NoError(t, err)

			if tc.expect != nil {
				tc.expect(t, mockHTTPClient)
			}

			response, err := gClient.GetPersona(tc.input)
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
				Workspace: "25443a54-1e10-45e8-8164-c38aa238615e",
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
			name: "missing workspace filter",
			want: want{
				err: client.NewErrMissingField("workspace"),
			},
		},
		{
			name: "http client error",
			input: client.PersonaSearchFilters{
				Workspace: "45443a54-1e10-45e8-8164-c38aa238615e",
			},
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/personas?"+
							"page_size=100&workspace=45443a54-1e10-45e8-8164-c38aa238615e", req.URL.String())

						return nil, errors.New("http error")
					})
			},
			want: want{
				err: errors.New("http error"),
			},
		},
		{
			name: "invalid response body",
			input: client.PersonaSearchFilters{
				Workspace: "25443a54-1e10-45e8-8164-c38aa238615e",
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
							"page_size=100&workspace=65443a54-1e10-45e8-8164-c38aa238615e", req.URL.String())

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
			gClient, err := client.New(testAPIKey, client.WithHTTPClient(mockHTTPClient))
			assert.NoError(t, err)

			if tc.expect != nil {
				tc.expect(t, mockHTTPClient)
			}

			response, err := gClient.PersonasSearch(tc.input)
			assert.Equal(t, tc.want.response, response)
			assert.Equal(t, tc.want.err, err)
		})
	}
}

func TestGreyNoiseClient_Account(t *testing.T) {
	testAPIKey := "test-2037klfsjlajf"

	type want struct {
		resp *client.Account
		err  error
	}
	testCases := []struct {
		name   string
		expect func(*testing.T, *client.MockHTTPClient)
		want   want
	}{
		{
			name: "happy path",
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/account", req.URL.String())

						return &http.Response{
							StatusCode: http.StatusOK,
							Body: responseBody(`{
							  "user_id": "cd280af5-a2df-4a4f-b512-206222bd5c9e",
							  "workspace_id": "343689ce-47bb-42c9-868d-c707fc82bd99"
							}`),
						}, nil
					})
			},
			want: want{
				resp: &client.Account{
					UserID:      uuid.MustParse("cd280af5-a2df-4a4f-b512-206222bd5c9e"),
					WorkspaceID: uuid.MustParse("343689ce-47bb-42c9-868d-c707fc82bd99"),
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
						assert.Equal(t, "https://api.greynoise.io/v1/account", req.URL.String())

						return nil, errors.New("ping error")
					})
			},
			want: want{
				err: errors.New("ping error"),
			},
		},
		{
			name: "unexpected status code",
			expect: func(t *testing.T, httpClient *client.MockHTTPClient) {
				httpClient.EXPECT().
					Do(gomock.Any()).
					DoAndReturn(func(req *http.Request) (*http.Response, error) {
						assert.Equal(t, req.Method, http.MethodGet)
						assert.Equal(t, testAPIKey, req.Header.Get(client.HeaderKey))
						assert.Equal(t, "https://api.greynoise.io/v1/account", req.URL.String())

						return &http.Response{
							StatusCode: http.StatusInternalServerError,
							Body:       responseBody(``),
						}, nil
					})
			},
			want: want{
				err: client.NewErrUnexpectedStatusCode(http.StatusOK, http.StatusInternalServerError),
			},
		},
	}
	for _, tc := range testCases {
		tc := tc

		t.Run(tc.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			defer ctrl.Finish()

			mockHTTPClient := client.NewMockHTTPClient(ctrl)
			gClient, err := client.New(testAPIKey, client.WithHTTPClient(mockHTTPClient))
			assert.NoError(t, err)

			if tc.expect != nil {
				tc.expect(t, mockHTTPClient)
			}

			acct, err := gClient.Account()
			assert.Equal(t, tc.want.err, err)
			assert.Equal(t, tc.want.resp, acct)
		})
	}
}

func responseBody(body string) io.ReadCloser {
	return io.NopCloser(strings.NewReader(body))
}

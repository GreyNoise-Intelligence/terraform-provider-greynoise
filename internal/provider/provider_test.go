package provider

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"

	"github.com/google/uuid"
	"github.com/hashicorp/terraform-plugin-framework/providerserver"
	"github.com/hashicorp/terraform-plugin-go/tfprotov6"

	"github.com/GreyNoise-Intelligence/terraform-provider-greynoise/internal/client"
)

const (
	testAPIKey      = "GjAQHGSWkCZokxG8TVHtowJHWA0A634lsIO7k4h8"
	testUserID      = "24525cdb-3104-420e-92a8-fe484540c72a"
	testWorkspaceID = "e6407966-8a1c-4a5d-a318-3e87673bda30"
)

var testAccProtoV6ProviderFactories = map[string]func() (tfprotov6.ProviderServer, error){
	"greynoise": providerserver.NewProtocol6WithError(New("test")()),
}

// mock API server
type mockEndpoint struct {
	method   string
	path     string
	match    func(*url.URL) bool
	status   int
	body     func() interface{}
	callback func(r *http.Request)
}

type mockAPIServer struct {
	Account       client.Account
	APIKey        string
	mockEndpoints []mockEndpoint
}

func defaultMockAPIServer() *mockAPIServer {
	return &mockAPIServer{
		Account: client.Account{
			UserID:      uuid.MustParse(testUserID),
			WorkspaceID: uuid.MustParse(testWorkspaceID),
		},
		APIKey: testAPIKey,
	}
}

func (m *mockAPIServer) Register(method string, path string, status int, body func() interface{},
	callback func(*http.Request)) {
	m.mockEndpoints = append(m.mockEndpoints, mockEndpoint{
		method:   method,
		path:     path,
		status:   status,
		body:     body,
		callback: callback,
	})
}

func (m *mockAPIServer) RegisterMatch(method string, path string, match func(*url.URL) bool, status int,
	body func() interface{}, callback func(*http.Request)) {
	m.mockEndpoints = append(m.mockEndpoints, mockEndpoint{
		method:   method,
		path:     path,
		status:   status,
		body:     body,
		match:    match,
		callback: callback,
	})
}

func (m *mockAPIServer) Server() *httptest.Server {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimSpace(r.URL.Path)

		if r.Header.Get(client.HeaderKey) != m.APIKey {
			w.Header().Set("Content-Type", "application/json")

			w.WriteHeader(http.StatusUnauthorized)
			_, _ = w.Write([]byte(`{"message": "unauthorized"}`))

			return
		}

		if path == "/v1/account" && r.Method == http.MethodGet {
			w.Header().Set("Content-Type", "application/json")

			w.WriteHeader(http.StatusOK)
			_ = json.NewEncoder(w).Encode(m.Account)

			return
		}

		for _, endpoint := range m.mockEndpoints {
			if endpoint.path == path && endpoint.method == r.Method {
				if endpoint.match != nil && !endpoint.match(r.URL) {
					continue
				}

				w.Header().Set("Content-Type", "application/json")

				w.WriteHeader(endpoint.status)
				_ = json.NewEncoder(w).Encode(endpoint.body())

				if endpoint.callback != nil {
					endpoint.callback(r)
				}

				return
			}
		}

		http.NotFoundHandler().ServeHTTP(w, r)
	}))

	return server
}

func body(b interface{}) func() interface{} {
	return func() interface{} {
		fmt.Println("Returning body", b)
		return b
	}
}

func emptyBody() interface{} {
	return nil
}

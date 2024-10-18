package service

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// APIResponse represents the response from the API
type APIResponse struct {
	StatusCode int
	Body       []byte
	Headers    http.Header
}

// APIClient holds the HTTP client and can have other configurations
type APIClient struct {
	Client *http.Client
}

func RequestHttpApi(url string, authtoken string, payload []byte, skipSSLVerify bool) bool {
	// Define the JSON payload
	// jsonData := []byte(`{"key": "value"}`)

	transport := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLVerify},
	}

	// Create a new POST request
	req, err := http.NewRequest("POST", url, bytes.NewBuffer(payload))
	if err != nil {
		fmt.Printf("Error creating request: %v\n", err)
		return false
	}

	// Add headers to the request
	if authtoken != "" {
		req.Header.Set("Authorization", authtoken)
	}
	req.Header.Set("Content-Type", "application/json")

	// Perform the request
	client := &http.Client{
		Timeout:   2 * time.Second,
		Transport: transport,
	}
	resp, err := client.Do(req)
	if err != nil {
		fmt.Printf("Error making request: %v\n", err)
		return false
	}
	defer resp.Body.Close()

	return resp.StatusCode == 200
}

// NewAPIClient creates a new APIClient with a specified timeout
func NewAPIClient(timeout time.Duration) *APIClient {
	return &APIClient{
		Client: &http.Client{
			Timeout: timeout,
		},
	}
}

// CallAPI is a generic function to make HTTP requests
func (api *APIClient) CallAPI(ctx context.Context, method, url string, headers map[string]string, payload interface{}) (*APIResponse, error) {
	var body io.Reader

	if payload != nil {
		// Determine the type of payload and encode accordingly
		switch v := payload.(type) {
		case string:
			body = bytes.NewBufferString(v)
		case []byte:
			body = bytes.NewBuffer(v)
		default:
			// Assume payload should be marshaled to JSON
			jsonBytes, err := json.Marshal(v)
			if err != nil {
				return nil, fmt.Errorf("error marshaling payload to JSON: %v", err)
			}
			body = bytes.NewBuffer(jsonBytes)
			// Set Content-Type to application/json if not already set
			if headers == nil {
				headers = make(map[string]string)
			}
			if _, exists := headers["Content-Type"]; !exists && method != "GET" && method != "HEAD" {
				headers["Content-Type"] = "application/json"
			}
		}
	}

	// Create a new HTTP request with context
	req, err := http.NewRequestWithContext(ctx, method, url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating HTTP request: %v", err)
	}

	// Set headers
	for key, value := range headers {
		req.Header.Set(key, value)
	}

	// Perform the HTTP request
	resp, err := api.Client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("error making HTTP request: %v", err)
	}
	defer resp.Body.Close()

	// Read the response body
	responseBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("error reading response body: %v", err)
	}

	// Construct the APIResponse
	apiResponse := &APIResponse{
		StatusCode: resp.StatusCode,
		Body:       responseBody,
		Headers:    resp.Header,
	}

	return apiResponse, nil
}

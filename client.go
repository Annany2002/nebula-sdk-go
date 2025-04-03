// client.go
package nebula

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	defaultTimeout = 10 * time.Second // Default timeout for HTTP client if not provided
	apiVersionPath = "/api/v1"        // Base path for V1 API routes
)

// Client manages communication with the Nebula BaaS API.
type Client struct {
	baseURL    *url.URL     // Parsed base URL of the Nebula BaaS API
	httpClient *http.Client // HTTP client for making requests
	authToken  string       // Internal storage for JWT (set after Login)

	// Services - Initialized in NewClient, provide access to grouped API methods
	Auth      AuthService
	Databases DatabaseService
	Tables    TableService
	Records   RecordService
	// Add Schemas service if needed separately from Databases/Tables
}

// NewClient creates a new Nebula BaaS API client.
// baseURL is the base address of your Nebula instance (e.g., "http://localhost:8080", "https://api.yourdomain.com").
// opts are functional options to customize the client (e.g., WithHTTPClient, WithRequestTimeout).
func NewClient(baseURL string, opts ...ClientOption) (*Client, error) {
	// 1. Validate and parse Base URL
	if baseURL == "" {
		return nil, fmt.Errorf("baseURL cannot be empty")
	}
	// Ensure trailing slash for easy joining, but remove potential multiple slashes
	baseURL = strings.TrimSuffix(baseURL, "/") + "/"
	parsedBaseURL, err := url.ParseRequestURI(baseURL)
	if err != nil {
		return nil, fmt.Errorf("invalid baseURL %q: %w", baseURL, err)
	}
	if parsedBaseURL.Scheme != "http" && parsedBaseURL.Scheme != "https" {
		return nil, fmt.Errorf("invalid baseURL scheme %q: must be http or https", parsedBaseURL.Scheme)
	}

	// 2. Process functional options
	options := clientOptions{ // Default options
		requestTimeout: defaultTimeout,
	}
	for _, opt := range opts {
		if err := opt(&options); err != nil {
			return nil, fmt.Errorf("failed to apply client option: %w", err)
		}
	}

	// 3. Configure HTTP Client
	httpClient := options.httpClient
	if httpClient == nil {
		// Create a default client if none provided
		httpClient = &http.Client{
			Timeout: options.requestTimeout, // Apply request timeout to default client
			// Can configure transport here if needed (e.g., TLS settings)
		}
	}

	// 4. Create the main client struct (initialize base fields)
	client := &Client{
		baseURL:    parsedBaseURL,
		httpClient: httpClient,
		// authToken will be set by Login
	}

	// 5. Initialize sub-services, passing the client reference
	client.Auth = AuthService{client: client}
	client.Databases = DatabaseService{client: client}
	client.Tables = TableService{client: client}
	client.Records = RecordService{client: client}
	// Initialize other services (like SchemaService if separated) here

	return client, nil
}

// --- Helper methods (could be in request.go later) ---

// getAPIPath constructs the full path for V1 API endpoints.
func (c *Client) getAPIPath(subPath string) string {
	// Ensure subPath doesn't start with / if apiVersionPath ends with /
	// Use path.Join or url.JoinPath for robust path joining later if needed.
	return apiVersionPath + "/" + strings.TrimPrefix(subPath, "/")
}

// SetAuthToken allows manually setting the JWT token if not using the Login method.
func (c *Client) SetAuthToken(token string) {
	// Consider adding validation or prefix check ("Bearer ") if desired
	c.authToken = token
}

// ClearAuthToken removes the internally stored JWT.
func (c *Client) ClearAuthToken() {
	c.authToken = ""
}

// request.go
package nebula

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const maxResponseBody = 1 * 1024 * 1024 // Limit response body read size to 1MB for safety

// doRequest performs an HTTP request to the Nebula API.
// It handles context, method, path joining, request body marshaling, auth header,
// response status checking, error mapping, and response body unmarshaling.
// - ctx: Context for cancellation/timeout.
// - c: The client instance containing config and auth token.
// - method: HTTP method (e.g., http.MethodGet).
// - apiPath: The API endpoint path *without* the base URL or leading slash (e.g., "auth/login", "databases").
// - requestBody: The struct/map to be marshaled into JSON for the request body (or nil).
// - responseBody: A pointer to a struct/map where the JSON response body should be unmarshaled (or nil).
func (c *Client) doRequest(ctx context.Context, method, apiPath string, requestBody interface{}, responseBody interface{}) error {
	// 1. Construct full URL
	// Use JoinPath available in Go 1.19+ for cleaner path handling
	// For older Go or simplicity:
	fullURL := c.baseURL.String() + strings.TrimPrefix(apiPath, "/")
	// Or using JoinPath:
	// relativeURL, err := url.Parse(apiPath) // Check error
	// fullURL := c.baseURL.ResolveReference(relativeURL).String()

	// 2. Prepare request body (if any)
	var bodyReader io.Reader
	var reqBytes []byte // For logging/debugging if needed
	var err error
	if requestBody != nil {
		reqBytes, err = json.Marshal(requestBody)
		if err != nil {
			return fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(reqBytes)
	}

	// 3. Create request with context
	req, err := http.NewRequestWithContext(ctx, method, fullURL, bodyReader)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	// 4. Set headers
	req.Header.Set("Accept", "application/json")
	if requestBody != nil {
		req.Header.Set("Content-Type", "application/json")
	}

	// Check if path requires authentication (simple prefix check for now)
	// Note: Assumes all paths under apiVersionPath require auth except /auth/*
	if strings.HasPrefix(apiPath, strings.TrimPrefix(apiVersionPath, "/")) &&
		!strings.HasPrefix(apiPath, "auth/") {
		if c.authToken == "" {
			log.Println("SDK Error: Attempted protected API call without auth token set.")
			return ErrAuthTokenMissing // Use SDK specific error
		}
		req.Header.Set("Authorization", "Bearer "+c.authToken)
	}

	// Log request details (optional, consider adding a logger option)
	// log.Printf("SDK Request: %s %s", method, fullURL)
	// if requestBody != nil { log.Printf("SDK Request Body: %s", string(reqBytes)) }

	// 5. Execute request
	resp, err := c.httpClient.Do(req)
	if err != nil {
		// Wrap network/transport errors
		return fmt.Errorf("http request failed: %w", err)
	}
	defer resp.Body.Close()

	// Log response status (optional)
	// log.Printf("SDK Response Status: %s", resp.Status)

	// 6. Check status code for errors (>= 400)
	if resp.StatusCode >= 400 {
		// Limit reading response body to prevent resource exhaustion
		limitedReader := io.LimitReader(resp.Body, maxResponseBody)
		respBytes, readErr := io.ReadAll(limitedReader)
		if readErr != nil {
			log.Printf("SDK Error: Failed to read error response body: %v", readErr)
			// Return an error based on status code but mention body read failure
			return mapHTTPError(resp.StatusCode, "failed to read error body", readErr)
		}

		// Try to unmarshal standard API error response `{"error": "message"}`
		var apiErrResp ErrorResponse
		jsonErr := json.Unmarshal(respBytes, &apiErrResp)
		errMsg := ""
		if jsonErr == nil && apiErrResp.Error != "" {
			errMsg = apiErrResp.Error // Use message from API if successfully parsed
		} else {
			// Fallback if body wasn't JSON or didn't match expected structure
			errMsg = fmt.Sprintf("API returned status %d", resp.StatusCode)
			if len(respBytes) > 0 && len(respBytes) < 200 { // Log small unknown bodies
				errMsg += " (" + string(respBytes) + ")"
			}
			log.Printf("SDK Warning: Could not parse API error response body (Status %d): %v. Body: %s", resp.StatusCode, jsonErr, string(respBytes))
		}

		// Use the helper function to map HTTP status to SDK error variable/type
		return mapHTTPError(resp.StatusCode, errMsg, nil) // Pass nil as underlying error unless there was a readErr?
	}

	// 7. Process successful response body (if expected)
	if responseBody != nil && resp.StatusCode != http.StatusNoContent {
		limitedReader := io.LimitReader(resp.Body, maxResponseBody)
		respBytes, readErr := io.ReadAll(limitedReader)
		if readErr != nil {
			log.Printf("SDK Error: Failed to read success response body: %v", readErr)
			return fmt.Errorf("%w: %w", ErrInvalidResponse, readErr)
		}

		err = json.Unmarshal(respBytes, responseBody)
		if err != nil {
			log.Printf("SDK Error: Failed to unmarshal success response body: %v. Body: %s", err, string(respBytes))
			// Return specific error indicating response parsing failure
			return fmt.Errorf("%w: %w", ErrInvalidResponse, err)
		}
	}

	return nil // Success
}

// errors.go
package nebula

import (
	"errors"
	"fmt"
	"net/http"
)

// Standard errors returned by the SDK
var (
	ErrBadRequest         = errors.New("bad request (400)")
	ErrUnauthorized       = errors.New("unauthorized (401 - check credentials/token)")
	ErrForbidden          = errors.New("forbidden (403)")
	ErrNotFound           = errors.New("resource not found (404)")
	ErrConflict           = errors.New("conflict (409 - e.g., resource already exists)")
	ErrRateLimited        = errors.New("rate limit exceeded (429)")
	ErrInternalServer     = errors.New("internal server error (500)")
	ErrInvalidResponse    = errors.New("invalid response from server")
	ErrAuthTokenMissing   = errors.New("authentication token not set in client")
	ErrDatabaseExists     = errors.New("database name already exists for this user")         // Specific example
	ErrDatabaseNotFound   = errors.New("database not found or not registered for this user") // Specific example
	ErrRecordNotFound     = errors.New("record not found")                                   // Specific example
	ErrTableNotFound      = errors.New("table not found")                                    // Specific example
	ErrInvalidFilterValue = errors.New("invalid value provided for filter")                  // Specific example
	// Add other specific, exported errors as needed
)

// APIError provides more context for errors returned by the Nebula API.
type APIError struct {
	StatusCode int    // The HTTP status code returned by the API.
	Message    string // The error message from the API response body (`{"error": "..."}`).
	Err        error  // Optional: underlying error (e.g., network error, json parsing error).
}

// Error implements the error interface for APIError.
func (e *APIError) Error() string {
	if e.Err != nil {
		return fmt.Sprintf("API error (status %d): %s [caused by: %v]", e.StatusCode, e.Message, e.Err)
	}
	return fmt.Sprintf("API error (status %d): %s", e.StatusCode, e.Message)
}

// Unwrap allows retrieving the underlying error using errors.Is/As.
func (e *APIError) Unwrap() error {
	return e.Err
}

// MapHTTPError maps an HTTP status code and an optional underlying error
// to one of the exported SDK error variables or a generic APIError.
// This function is intended for internal SDK use (in request.go).
func mapHTTPError(statusCode int, apiMsg string, underlyingErr error) error {
	// Use default message if API message is empty
	if apiMsg == "" {
		apiMsg = "API returned status " + fmt.Sprintf("%d", statusCode)
	}

	// Map status codes to specific error variables
	var baseErr error
	switch statusCode {
	case http.StatusBadRequest: // 400
		baseErr = ErrBadRequest
	case http.StatusUnauthorized: // 401
		baseErr = ErrUnauthorized
	case http.StatusForbidden: // 403
		baseErr = ErrForbidden
	case http.StatusNotFound: // 404
		baseErr = ErrNotFound
	case http.StatusConflict: // 409
		baseErr = ErrConflict
	case http.StatusTooManyRequests: // 429
		baseErr = ErrRateLimited
	case http.StatusInternalServerError: // 500
		baseErr = ErrInternalServer
	// Add other common codes (502, 503, 504) if needed
	default:
		// For unmapped client/server errors, use a generic error
		if statusCode >= 400 && statusCode < 500 {
			baseErr = fmt.Errorf("unexpected client error")
		} else if statusCode >= 500 {
			baseErr = fmt.Errorf("unexpected server error")
		} else {
			// Should not happen if checking > 400, but just in case
			baseErr = fmt.Errorf("unexpected status code")
		}
	}

	// Return wrapped APIError
	return &APIError{
		StatusCode: statusCode,
		Message:    apiMsg,
		Err:        fmt.Errorf("%w: %w", baseErr, underlyingErr), // Wrap base and underlying
	}
}

// Helper to check if an error is specifically *our* APIError with a given status code
func IsAPIErrorStatus(err error, statusCode int) bool {
	var apiErr *APIError
	if errors.As(err, &apiErr) {
		return apiErr.StatusCode == statusCode
	}
	return false
}

// config.go
package nebula

import (
	"fmt"
	"net/http"
	"time"
)

// clientOptions holds internal configuration settings for the Client.
type clientOptions struct {
	httpClient     *http.Client
	requestTimeout time.Duration
	// Add other options like custom logger, retry policy, etc. here
}

// ClientOption is a function type used to configure the Client using the functional options pattern.
type ClientOption func(*clientOptions) error

// WithHTTPClient sets a custom http.Client for the Nebula client.
// Useful for configuring proxies, custom transports, or timeouts on the client level.
func WithHTTPClient(hc *http.Client) ClientOption {
	return func(o *clientOptions) error {
		if hc == nil {
			return fmt.Errorf("http client cannot be nil")
		}
		o.httpClient = hc
		return nil
	}
}

// WithRequestTimeout sets a default timeout for individual HTTP requests made by the SDK client.
// If not set, the timeout from the underlying http.Client (if configured) or Go's default will apply.
func WithRequestTimeout(d time.Duration) ClientOption {
	return func(o *clientOptions) error {
		if d <= 0 {
			return fmt.Errorf("request timeout must be positive")
		}
		o.requestTimeout = d
		return nil
	}
}

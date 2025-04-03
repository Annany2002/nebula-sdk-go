// auth.go
package nebula

import (
	"context"
	"fmt"
	"net/http"
)

// AuthService provides methods for interacting with the /auth endpoints.
type AuthService struct {
	client *Client // Reference back to the main client for config and request helper
}

// Signup registers a new user account.
func (s *AuthService) Signup(ctx context.Context, email, password string) error {
	if email == "" || password == "" {
		// Basic validation client-side
		return fmt.Errorf("email and password cannot be empty")
	}

	payload := SignupPayload{
		Email:    email,
		Password: password,
	}
	apiPath := "auth/signup" // Path relative to base URL

	// Use doRequest helper. No response body expected on success (201).
	err := s.client.doRequest(ctx, http.MethodPost, apiPath, payload, nil)
	if err != nil {
		// doRequest already maps API errors (e.g., 409 Conflict)
		return err
	}

	return nil // Success
}

// Login authenticates a user and stores the returned JWT token within the client
// for subsequent authenticated requests. It also returns the token.
func (s *AuthService) Login(ctx context.Context, email, password string) (string, error) {
	if email == "" || password == "" {
		return "", fmt.Errorf("email and password cannot be empty")
	}

	payload := LoginPayload{
		Email:    email,
		Password: password,
	}
	apiPath := "auth/login"

	var result LoginResponse // Define where to store the successful response body

	// Use doRequest helper, passing pointer to result struct.
	err := s.client.doRequest(ctx, http.MethodPost, apiPath, payload, &result)
	if err != nil {
		// doRequest maps API errors (e.g., 401, 404)
		s.client.ClearAuthToken() // Ensure token is cleared on failed login attempt
		return "", err
	}

	// Login successful, store the token internally
	s.client.SetAuthToken(result.Token) // Use setter if validation/prefixing needed

	return result.Token, nil // Return token and nil error
}

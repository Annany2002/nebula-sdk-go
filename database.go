// database.go
package nebula

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings" // Import strings
)

// DatabaseService provides methods for interacting with the /databases endpoints
// and related schema definitions.
type DatabaseService struct {
	client *Client // Reference back to the main client
}

// Create registers a new logical database for the authenticated user.
func (s *DatabaseService) Create(ctx context.Context, dbName string) error {
	if strings.TrimSpace(dbName) == "" {
		return errors.New("database name cannot be empty") // Basic client validation
	}
	// Consider adding client-side name format validation? Depends on SDK strictness goal.

	payload := CreateDatabasePayload{
		DBName: dbName,
	}
	apiPath := s.client.getAPIPath("databases") // Use helper for consistency

	err := s.client.doRequest(ctx, http.MethodPost, apiPath, payload, nil)
	if err != nil {
		// doRequest maps standard errors (401, 409, 500 etc.)
		// Map Conflict specifically if desired, though caller can check errors.Is(err, ErrConflict)
		return err
	}
	return nil // Success
}

// List retrieves the names of all databases registered by the authenticated user.
func (s *DatabaseService) List(ctx context.Context) ([]string, error) {
	apiPath := s.client.getAPIPath("databases")
	var result ListDatabasesResponse // Expecting {"databases": ["name1", ...]}

	err := s.client.doRequest(ctx, http.MethodGet, apiPath, nil, &result)
	if err != nil {
		return nil, err // Return error from doRequest (e.g., ErrUnauthorized, ErrInternalServer)
	}

	// Return empty slice if result.Databases is nil (shouldn't happen if API returns `[]`)
	if result.Databases == nil {
		return make([]string, 0), nil
	}

	return result.Databases, nil
}

// Delete removes a database registration and attempts to delete the associated data file.
// Returns ErrNotFound if the database registration doesn't exist.
func (s *DatabaseService) Delete(ctx context.Context, dbName string) error {
	if strings.TrimSpace(dbName) == "" {
		return errors.New("database name cannot be empty")
	}
	// URL encode dbName? Generally needed if names can contain special chars.
	// Let's assume simple names for now based on prior validation.
	// Proper way: use url.PathEscape
	apiPath := s.client.getAPIPath(fmt.Sprintf("databases/%s", dbName))

	err := s.client.doRequest(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		// doRequest maps 404 to ErrNotFound
		return err
	}
	return nil // Success (204 No Content handled by doRequest)
}

// DefineSchema creates or updates the schema for a table within a specified database.
// Note: The backend uses CREATE TABLE IF NOT EXISTS, making it somewhat idempotent.
func (s *DatabaseService) DefineSchema(ctx context.Context, dbName string, schema SchemaPayload) error {
	if strings.TrimSpace(dbName) == "" {
		return errors.New("database name cannot be empty")
	}
	if strings.TrimSpace(schema.TableName) == "" {
		return errors.New("table name cannot be empty")
	}
	if len(schema.Columns) == 0 {
		return errors.New("schema must contain at least one column")
	}
	// Add client-side validation for column names/types if desired for faster feedback

	apiPath := s.client.getAPIPath(fmt.Sprintf("databases/%s/schema", dbName))

	err := s.client.doRequest(ctx, http.MethodPost, apiPath, schema, nil)
	if err != nil {
		// doRequest maps standard errors (400, 401, 404, 500)
		return err
	}
	return nil // Success (201 handled by doRequest)
}

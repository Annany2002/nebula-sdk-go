// table.go
package nebula

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"strings"
)

// TableService provides methods for interacting with table-level endpoints
// (listing tables, deleting tables) within a specific database.
type TableService struct {
	client *Client // Reference back to the main client
}

// List retrieves the names of all tables within a specific database.
func (s *TableService) List(ctx context.Context, dbName string) ([]string, error) {
	if strings.TrimSpace(dbName) == "" {
		return nil, errors.New("database name cannot be empty")
	}

	apiPath := s.client.getAPIPath(fmt.Sprintf("databases/%s/tables", dbName))
	var result ListTablesResponse // Expecting {"tables": ["name1", ...]}

	err := s.client.doRequest(ctx, http.MethodGet, apiPath, nil, &result)
	if err != nil {
		// Handle potential ErrNotFound if dbName doesn't exist
		return nil, err
	}

	if result.Tables == nil {
		return make([]string, 0), nil
	}
	return result.Tables, nil
}

// Delete drops a specific table within a database.
func (s *TableService) Delete(ctx context.Context, dbName, tableName string) error {
	if strings.TrimSpace(dbName) == "" || strings.TrimSpace(tableName) == "" {
		return errors.New("database name and table name cannot be empty")
	}

	apiPath := s.client.getAPIPath(fmt.Sprintf("databases/%s/tables/%s", dbName, tableName))

	err := s.client.doRequest(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		// Handle potential ErrNotFound if dbName or tableName doesn't exist
		return err
	}
	return nil // Success (204)
}

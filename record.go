// record.go
package nebula

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// RecordService provides methods for interacting with record endpoints
// within a specific database and table.
type RecordService struct {
	client *Client // Reference back to the main client
}

// buildRecordPath constructs the base API path for record operations.
func (s *RecordService) buildRecordPath(dbName, tableName string) (string, error) {
	if strings.TrimSpace(dbName) == "" || strings.TrimSpace(tableName) == "" {
		return "", errors.New("database name and table name cannot be empty")
	}
	// Future: consider url.PathEscape for dbName and tableName if they can contain special chars
	return s.client.getAPIPath(fmt.Sprintf("databases/%s/tables/%s/records", dbName, tableName)), nil
}

// buildSingleRecordPath constructs the API path for operations on a specific record.
func (s *RecordService) buildSingleRecordPath(dbName, tableName string, recordID int64) (string, error) {
	basePath, err := s.buildRecordPath(dbName, tableName)
	if err != nil {
		return "", err
	}
	if recordID <= 0 {
		return "", errors.New("record ID must be positive")
	}
	return fmt.Sprintf("%s/%d", basePath, recordID), nil
}

// Create inserts a new record into the specified table.
// recordData should be a map[string]interface{} representing the data.
// Returns the ID of the newly created record.
func (s *RecordService) Create(ctx context.Context, dbName, tableName string, recordData map[string]interface{}) (int64, error) {
	if len(recordData) == 0 {
		return 0, errors.New("record data cannot be empty")
	}
	apiPath, err := s.buildRecordPath(dbName, tableName)
	if err != nil {
		return 0, err
	}

	var result CreateRecordResponse
	err = s.client.doRequest(ctx, http.MethodPost, apiPath, recordData, &result)
	if err != nil {
		// Handles 400 (bad type/col), 401, 404 (db/table not found), 409 (constraint), 500
		return 0, err
	}

	return result.RecordID, nil
}

// Get retrieves a single record by its ID.
// Returns a map representing the record, or ErrNotFound if the record ID doesn't exist.
func (s *RecordService) Get(ctx context.Context, dbName, tableName string, recordID int64) (map[string]interface{}, error) {
	apiPath, err := s.buildSingleRecordPath(dbName, tableName, recordID)
	if err != nil {
		return nil, err // Handles invalid db/table/recordID
	}

	var result map[string]interface{} // Expecting a single JSON object
	err = s.client.doRequest(ctx, http.MethodGet, apiPath, nil, &result)
	if err != nil {
		// Handles 401, 404 (db/table/record not found), 500
		return nil, err
	}
	return result, nil
}

// Update modifies fields of an existing record.
// updateData should be a map containing *only* the fields to be changed.
// Returns nil on success.
func (s *RecordService) Update(ctx context.Context, dbName, tableName string, recordID int64, updateData map[string]interface{}) error {
	if len(updateData) == 0 {
		return errors.New("update data cannot be empty")
	}
	apiPath, err := s.buildSingleRecordPath(dbName, tableName, recordID)
	if err != nil {
		return err
	}

	// Backend returns 200 OK with body, but we only need to check for errors here.
	// Pass nil for responseBody as we are just returning error.
	err = s.client.doRequest(ctx, http.MethodPut, apiPath, updateData, nil)
	if err != nil {
		// Handles 400 (bad type/col), 401, 404 (db/table/record not found), 409 (constraint), 500
		return err
	}
	return nil // Success (200 OK)
}

// Delete removes a specific record by its ID.
// Returns nil on success (204 No Content).
func (s *RecordService) Delete(ctx context.Context, dbName, tableName string, recordID int64) error {
	apiPath, err := s.buildSingleRecordPath(dbName, tableName, recordID)
	if err != nil {
		return err
	}

	// Expect 204 No Content on success, responseBody is nil
	err = s.client.doRequest(ctx, http.MethodDelete, apiPath, nil, nil)
	if err != nil {
		// Handles 401, 404 (db/table/record not found), 500
		return err
	}
	return nil // Success
}

// List retrieves records from the specified table, optionally applying filters.
// Filters are key-value pairs for simple equality matching (e.g., {"status":"active", "priority":"1"}).
// Returns a slice of maps, where each map represents a record.

// --- *** MODIFIED: List retrieves records using ListRecordsOptions *** ---
// Accepts optional parameters via the opts struct.
func (s *RecordService) List(ctx context.Context, dbName, tableName string, opts *ListRecordsOptions) ([]map[string]interface{}, error) {
	apiPath, err := s.buildRecordPath(dbName, tableName)
	if err != nil {
		return nil, err
	}

	queryValues := url.Values{}

	// Process options if provided
	if opts != nil {
		// Add Filters
		if opts.Filters != nil {
			for key, value := range opts.Filters {
				if key != "" { // Ignore empty keys
					queryValues.Add(key, value) // Value is already string
				}
			}
		}

		// Add Limit (if backend supported it)
		if opts.Limit != nil {
			if *opts.Limit >= 0 { // Allow 0 potentially, though backend might enforce > 0
				queryValues.Add("limit", strconv.Itoa(*opts.Limit))
			}
		}

		// Add Offset (if backend supported it)
		if opts.Offset != nil {
			if *opts.Offset >= 0 {
				queryValues.Add("offset", strconv.Itoa(*opts.Offset))
			}
		}

		// Add Sort (if backend supported it)
		if opts.SortBy != nil && *opts.SortBy != "" {
			sortParam := *opts.SortBy
			if opts.SortDirection != nil {
				dir := strings.ToLower(*opts.SortDirection)
				if dir == "asc" || dir == "desc" {
					sortParam = fmt.Sprintf("%s:%s", sortParam, dir)
				} else {
					// Optionally return an error for invalid sort direction
					// return nil, fmt.Errorf("invalid sort direction: %s", *opts.SortDirection)
					// Or just ignore invalid direction and sort by column ascending
				}
			}
			queryValues.Add("sort", sortParam)
		}
	}

	// Append query string if any parameters were added
	if len(queryValues) > 0 {
		apiPath = fmt.Sprintf("%s?%s", apiPath, queryValues.Encode())
	}

	var result []map[string]interface{} // Expecting a JSON array of objects
	err = s.client.doRequest(ctx, http.MethodGet, apiPath, nil, &result)
	if err != nil {
		// Handles 400 (if backend adds validation for limit/offset/sort/filter), 401, 404, 500
		return nil, err
	}

	if result == nil {
		return make([]map[string]interface{}, 0), nil
	}
	return result, nil
}

// --- *** END MODIFIED *** ---

// models.go
package nebula

// --- Auth Models ---

// SignupPayload defines the structure for the signup request body.
type SignupPayload struct {
	Email    string `json:"email"`    // Binding:"required,email" handled by server
	Password string `json:"password"` // Binding:"required,min=8" handled by server
}

// LoginPayload defines the structure for the login request body.
type LoginPayload struct {
	Email    string `json:"email"`    // Binding:"required" handled by server
	Password string `json:"password"` // Binding:"required" handled by server
}

// LoginResponse defines the structure for the successful login response body.
type LoginResponse struct {
	Message string `json:"message"`
	Token   string `json:"token"` // The JWT token
}

// --- *** NEW/UPDATED: Database/Schema/Table Models *** ---

// CreateDatabasePayload defines the structure for creating a database registration.
type CreateDatabasePayload struct {
	DBName string `json:"db_name"`
}

// ListDatabasesResponse defines the structure for the list databases response.
type ListDatabasesResponse struct {
	Databases []string `json:"databases"`
}

// ColumnDefinition represents a single column in a table schema request/response.
type ColumnDefinition struct {
	Name string `json:"name"`
	Type string `json:"type"` // e.g., "TEXT", "INTEGER", "REAL", "BLOB", "BOOLEAN"
}

// SchemaPayload defines the structure for the schema creation request body.
// Renamed from CreateSchemaRequest for potential broader use.
type SchemaPayload struct {
	TableName string             `json:"table_name"`
	Columns   []ColumnDefinition `json:"columns"`
}

// ListTablesResponse defines the structure for the list tables response.
type ListTablesResponse struct {
	Tables []string `json:"tables"`
}

// --- Record Models ---
// CreateRecordResponse defines the structure for the create record success response.
type CreateRecordResponse struct {
	Message  string `json:"message"`
	RecordID int64  `json:"record_id"`
}

// UpdateRecordResponse defines the structure for the update record success response.
// Note: The SDK's Update method currently just returns error, but if needed,
// it could return this struct parsed from the response body.
type UpdateRecordResponse struct {
	Message      string `json:"message"`
	RecordID     int64  `json:"record_id"`
	RowsAffected int64  `json:"rows_affected"`
}

// Note: For ListRecords and GetRecord, the API returns []map[string]interface{}
// or map[string]interface{}, so no specific SDK structs are strictly needed
// for the record data itself, unless you want to provide helpers for
// unmarshaling into user-defined structs later.

// --- *** NEW: Options for Listing Records *** ---

// ListRecordsOptions specifies optional parameters for listing records.
// Uses pointers to distinguish between zero values and unset parameters.
type ListRecordsOptions struct {
	// Filters apply simple equality checks (e.g., {"status":"active", "priority":"1"}).
	// Backend validates keys and converts values based on schema.
	Filters map[string]string

	// Limit the number of records returned (for pagination). Backend support required.
	Limit *int

	// Offset the starting point of the returned records (for pagination). Backend support required.
	Offset *int

	// SortBy specifies the column name to sort by. Backend support required.
	SortBy *string

	// SortDirection specifies the sort direction ("asc" or "desc"). Backend support required.
	SortDirection *string // "asc" or "desc"
}

// ErrorResponse defines the standard JSON error structure returned by the API.
type ErrorResponse struct {
	Error string `json:"error"`
}

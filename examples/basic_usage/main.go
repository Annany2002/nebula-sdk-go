// examples/basic_usage/main.go
package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	nebula "github.com/Annany2002/nebula-sdk-go" // Use your actual SDK module path
)

func main() {
	log.Println("--- Nebula Go SDK Example ---")

	// --- Configuration ---
	// Get Base URL from environment variable or use default
	baseURL := os.Getenv("NEBULA_BASE_URL")
	if baseURL == "" {
		baseURL = "http://localhost:8080" // Default for local backend dev
		log.Printf("NEBULA_BASE_URL not set, using default: %s", baseURL)
	}
	log.Printf("Using Nebula API at: %s", baseURL)

	// Get credentials from env vars or use defaults (for example only!)
	userEmail := os.Getenv("NEBULA_TEST_EMAIL")
	if userEmail == "" {
		userEmail = "sdk-example-user@example.com"
	}
	userPassword := os.Getenv("NEBULA_TEST_PASSWORD")
	if userPassword == "" {
		userPassword = "SdkExamplePassword123!"
	}

	// --- Create Client ---
	// Use a timeout for context
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second) // Overall timeout for example run
	defer cancel()

	client, err := nebula.NewClient(baseURL, nebula.WithRequestTimeout(10*time.Second))
	if err != nil {
		log.Fatalf("FATAL: Error creating Nebula client: %v", err)
	}
	log.Println("Nebula client created.")

	// --- Authentication ---
	log.Println("\n--- Step 1: Authentication ---")
	// Try to sign up (ignore conflict if user already exists)
	err = client.Auth.Signup(ctx, userEmail, userPassword)
	if err != nil {
		if errors.Is(err, nebula.ErrConflict) { // Use exported SDK errors
			log.Printf("Signup skipped: User '%s' already exists.", userEmail)
		} else {
			log.Fatalf("FATAL: Signup failed unexpectedly: %v", err) // Fail on other errors
		}
	} else {
		log.Printf("Signup successful for user '%s'.", userEmail)
	}

	// Log in to get token stored in client
	_, err = client.Auth.Login(ctx, userEmail, userPassword)
	if err != nil {
		log.Fatalf("FATAL: Login failed: %v", err)
	}
	log.Println("Login successful (token stored in SDK client).")

	// --- Database and Schema ---
	log.Println("\n--- Step 2: Database & Schema Management ---")
	dbName := "sdk_demo_inventory"
	tableName := "widgets"

	// Create database registration (ignore conflict)
	err = client.Databases.Create(ctx, dbName)
	if err != nil && !errors.Is(err, nebula.ErrConflict) {
		log.Fatalf("FATAL: Failed to create database registration '%s': %v", dbName, err)
	} else {
		log.Printf("Database registration '%s' ensured.", dbName)
	}

	// Define schema (ignore errors if table likely exists)
	schema := nebula.SchemaPayload{
		TableName: tableName,
		Columns: []nebula.ColumnDefinition{
			{Name: "widget_name", Type: "TEXT"},
			{Name: "color", Type: "TEXT"},
			{Name: "quantity", Type: "INTEGER"},
			{Name: "is_active", Type: "BOOLEAN"},
		},
	}
	err = client.Databases.DefineSchema(ctx, dbName, schema)
	if err != nil {
		// Could check for specific "table already exists" if API/SDK provided it
		log.Printf("WARN: DefineSchema failed (maybe table '%s' already exists?): %v", tableName, err)
	} else {
		log.Printf("Schema for table '%s' defined.", tableName)
	}

	// --- Record CRUD ---
	log.Println("\n--- Step 3: Record Operations ---")
	var createdID int64 // To store ID for later steps

	// Create Record
	log.Println("Creating record...")
	recordData1 := map[string]interface{}{
		"widget_name": "Standard Widget",
		"color":       "Blue",
		"quantity":    100,
		"is_active":   true,
	}
	createdID, err = client.Records.Create(ctx, dbName, tableName, recordData1)
	if err != nil {
		log.Fatalf("FATAL: Failed to create record: %v", err)
	}
	log.Printf("Record 1 created with ID: %d", createdID)

	// Create another record
	recordData2 := map[string]interface{}{
		"widget_name": "Premium Widget",
		"color":       "Red",
		"quantity":    50,
		"is_active":   true,
	}
	_, err = client.Records.Create(ctx, dbName, tableName, recordData2)
	if err != nil {
		log.Fatalf("FATAL: Failed to create second record: %v", err)
	}
	log.Println("Record 2 created.")

	// List Records (with filter)
	log.Println("Listing 'Blue' records...")
	filters := map[string]string{"color": "Blue"}
	records, err := client.Records.List(ctx, dbName, tableName, filters)
	if err != nil {
		log.Fatalf("FATAL: Failed to list records: %v", err)
	}
	log.Printf("Found %d 'Blue' record(s):", len(records))
	for _, rec := range records {
		log.Printf("  -> %+v", rec) // Print map content
	}

	// Get Single Record
	log.Printf("Getting record ID %d...", createdID)
	record, err := client.Records.Get(ctx, dbName, tableName, createdID)
	if err != nil {
		if errors.Is(err, nebula.ErrNotFound) {
			log.Fatalf("FATAL: Could not get record %d (Not Found): %v", createdID, err)
		} else {
			log.Fatalf("FATAL: Failed to get record %d: %v", createdID, err)
		}
	}
	log.Printf("Got record %d: %+v", createdID, record)

	// Update Record
	log.Printf("Updating record ID %d (setting quantity to 99)...", createdID)
	updateData := map[string]interface{}{"quantity": 99, "is_active": false}
	err = client.Records.Update(ctx, dbName, tableName, createdID, updateData)
	if err != nil {
		if errors.Is(err, nebula.ErrNotFound) {
			log.Fatalf("FATAL: Could not update record %d (Not Found): %v", createdID, err)
		} else {
			log.Fatalf("FATAL: Failed to update record %d: %v", createdID, err)
		}
	}
	log.Printf("Record %d updated.", createdID)

	// Get Updated Record
	log.Printf("Getting updated record ID %d...", createdID)
	updatedRecord, err := client.Records.Get(ctx, dbName, tableName, createdID)
	if err != nil {
		log.Fatalf("FATAL: Failed to get updated record %d: %v", createdID, err)
	}
	log.Printf("Got updated record %d: %+v", createdID, updatedRecord)

	// Delete Record
	log.Printf("Deleting record ID %d...", createdID)
	err = client.Records.Delete(ctx, dbName, tableName, createdID)
	if err != nil {
		if errors.Is(err, nebula.ErrNotFound) {
			log.Fatalf("FATAL: Could not delete record %d (Not Found): %v", createdID, err)
		} else {
			log.Fatalf("FATAL: Failed to delete record %d: %v", createdID, err)
		}
	}
	log.Printf("Record %d deleted.", createdID)

	// Verify Deletion
	log.Printf("Verifying deletion of record ID %d...", createdID)
	_, err = client.Records.Get(ctx, dbName, tableName, createdID)
	if err != nil {
		if errors.Is(err, nebula.ErrNotFound) {
			log.Printf("Verified: Record %d not found after deletion.", createdID)
		} else {
			log.Fatalf("FATAL: Error verifying deletion: %v", err)
		}
	} else {
		log.Fatalf("FATAL: Record %d still exists after deletion!", createdID)
	}

	// --- Cleanup ---
	log.Println("\n--- Step 4: Cleanup ---")
	// Delete Table
	log.Printf("Deleting table '%s'...", tableName)
	err = client.Tables.Delete(ctx, dbName, tableName)
	if err != nil {
		log.Printf("WARN: Failed to delete table '%s': %v", tableName, err)
	} else {
		log.Printf("Table '%s' deleted.", tableName)
	}

	// Delete Database
	log.Printf("Deleting database '%s'...", dbName)
	err = client.Databases.Delete(ctx, dbName)
	if err != nil {
		log.Printf("WARN: Failed to delete database '%s': %v", dbName, err)
	} else {
		log.Printf("Database '%s' deleted.", dbName)
	}

	log.Println("\n--- Nebula Go SDK Example Finished ---")
}

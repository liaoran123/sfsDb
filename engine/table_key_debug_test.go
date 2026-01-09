package engine

import (
	"fmt"
	"testing"
)

// TestTableKeyGenerationDebug debug test to understand what keys are being generated
func TestTableKeyGenerationDebug(t *testing.T) {
	// Create test table
	table, err := TableNew("test_key_debug")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{"id": 0, "name": ""}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index on id
	pk, _ := DefaultPrimaryKeyNew("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Insert test data
	for i := 1; i <= 5; i++ {
		testData := map[string]any{"id": i, "name": "User" + string(rune('A'+i-1))}
		_, err := table.Insert(&testData)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Debug: Print the primary key prefix
	pkPrefix := table.GetPrimaryKey().Prefix(table.id)
	fmt.Printf("Primary key prefix: %q\n", pkPrefix)

	// Debug: Try to understand what key is being generated for search
	searchData := map[string]any{"id": 2}
	fieldsBytes := table.FieldsToBytesNil(&searchData)
	idx := table.MatchIndex("id")
	key := idx.JoinValue(fieldsBytes, table.id)
	fmt.Printf("Search key for id=2: %q\n", key)

	// Debug: Check what the actual keys in the database look like
	fmt.Println("Debug: Scanning all keys in the table...")
	iter := table.Search(&map[string]any{"id": 1}) // Use Like to get all
	if iter != nil {
		records := iter.GetRecords(true)
		if records != nil {
			fmt.Printf("Found %d records\n", len(records.Select()))
			for i, record := range records.Select() {
				fmt.Printf("Record %d: %v\n", i, record)
			}
		}
		iter.Release()
	}
}

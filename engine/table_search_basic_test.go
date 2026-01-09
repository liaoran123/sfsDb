package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/util"
)

// TestTableSearchBasic tests basic search functionality
func TestTableSearchBasic(t *testing.T) {
	// Create test table
	table, err := TableNew("test_search_basic")
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
	testData := []map[string]any{
		{"id": 1, "name": "Alice"},
		{"id": 2, "name": "Bob"},
		{"id": 3, "name": "Charlie"},
		{"id": 4, "name": "David"},
		{"id": 5, "name": "Eve"},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Test cases for basic search scenarios
	testCases := []struct {
		name          string
		searchData    map[string]any
		operator      util.ComparisonOperator
		expectedCount int
	}{
		// Test 1: Exact match on primary key
		{
			name:          "Exact match on id=3",
			searchData:    map[string]any{"id": 3},
			operator:      util.Equal,
			expectedCount: 1,
		},

		// Test 2: Search for non-existent record
		{
			name:          "Search for non-existent id=999",
			searchData:    map[string]any{"id": 999},
			operator:      util.Equal,
			expectedCount: 0,
		},

		// Test 3: Default operator (Like)
		{
			name:          "Default operator on id=1",
			searchData:    map[string]any{"id": 1},
			operator:      util.Like,
			expectedCount: 1, // Should return at least 1 record
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Search with the specified operator
			iter := table.Search(&tc.searchData, tc.operator)
			if iter == nil {
				t.Fatalf("Search returned nil iterator for case: %s", tc.name)
			}
			defer iter.Release()

			// Collect results
			var count int
			records := iter.GetRecords(true)
			if records != nil {
				count = len(records.Select())
			}

			// Verify results
			if count != tc.expectedCount {
				t.Errorf("Expected %d results for %s, got %d", tc.expectedCount, tc.name, count)
				// Print actual results for debugging
				if records != nil {
					t.Logf("Actual results: %v", records.Select())
				}
			}
		})
	}
}

// TestTableSearchWithSecondaryIndex tests search with secondary index
func TestTableSearchWithSecondaryIndex(t *testing.T) {
	// Create test table
	table, err := TableNew("test_search_secondary")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{"id": 0, "name": "", "age": 0}
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

	// Create secondary index on age
	ageIdx, _ := DefaultNormalIndexNew("age_index")
	ageIdx.AddFields("age")
	err = table.CreateIndex(ageIdx)
	if err != nil {
		t.Fatalf("Failed to create age index: %v", err)
	}

	// Insert test data
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20},
		{"id": 2, "name": "Bob", "age": 25},
		{"id": 3, "name": "Charlie", "age": 30},
		{"id": 4, "name": "David", "age": 35},
		{"id": 5, "name": "Eve", "age": 40},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Test search on secondary index
	t.Run("Search on secondary index (age=30)", func(t *testing.T) {
		searchData := map[string]any{"age": 30}
		iter := table.Search(&searchData, util.Equal)
		if iter == nil {
			t.Fatalf("Search returned nil iterator")
		}
		defer iter.Release()

		// Collect results
		var count int
		records := iter.GetRecords(true)
		if records != nil {
			count = len(records.Select())
		}

		// Should find the exact match
		if count != 1 {
			t.Errorf("Expected 1 result, got %d", count)
		}
	})
}

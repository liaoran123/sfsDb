package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/util"
)

// TestTableSearchComparisonOperatorsSimple tests comparison operators with Table.Search
// using a simplified approach to isolate the operator functionality
func TestTableSearchComparisonOperatorsSimple(t *testing.T) {
	// Create test table
	table, err := TableNew("test_search_comparison")
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
	pk, _ := NewDefaultPrimaryKey("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Insert test data (simple sequential IDs)
	for i := 1; i <= 5; i++ {
		testData := map[string]any{"id": i, "name": "User" + string(rune('A'+i-1))}
		_, err := table.Insert(&testData)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Test cases for comparison operators with id field
	testCases := []struct {
		name          string
		searchValue   int
		operator      util.ComparisonOperator
		expectedCount int
	}{
		{
			name:          "Equal to 3",
			searchValue:   3,
			operator:      util.Equal,
			expectedCount: 1, // Should find only id=3
		},
		{
			name:          "Greater than 2",
			searchValue:   2,
			operator:      util.GreaterThan,
			expectedCount: 3, // Should find ids=3,4,5
		},
		{
			name:          "Greater than or equal to 2",
			searchValue:   2,
			operator:      util.GreaterThanOrEqual,
			expectedCount: 4, // Should find ids=2,3,4,5
		},
		{
			name:          "Less than 3",
			searchValue:   3,
			operator:      util.LessThan,
			expectedCount: 2, // Should find ids=1,2
		},
		{
			name:          "Less than or equal to 3",
			searchValue:   3,
			operator:      util.LessThanOrEqual,
			expectedCount: 3, // Should find ids=1,2,3
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create search data with the test value
			searchData := map[string]any{"id": tc.searchValue}

			// Search with the specified operator
			iter, _ := table.Search(&searchData, tc.operator)
			defer GlobalTableIterPool.Put(iter)

			if iter == nil {
				t.Fatalf("Search returned nil iterator for operator %s", tc.operator)
			}
			//defer GlobalTableIterPool.Put(iter)

			// Collect results by iterating through records
			var count int
			records := iter.GetRecords(true)
			if records != nil {
				count = len(records.Select())
			}

			// Verify results
			if count != tc.expectedCount {
				t.Errorf("Expected %d results for %s, got %d", tc.expectedCount, tc.name, count)
			}
		})
	}
}

// TestTableSearchDefaultOperator tests that the default operator is Like
func TestTableSearchDefaultOperator(t *testing.T) {
	// Create test table
	table, err := TableNew("test_search_default")
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
	pk, _ := NewDefaultPrimaryKey("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Insert test data
	for i := 1; i <= 3; i++ {
		testData := map[string]any{"id": i, "name": "User" + string(rune('A'+i-1))}
		_, err := table.Insert(&testData)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Test default operator (should be Like)
	searchData := map[string]any{"id": 1}
	iter, _ := table.Search(&searchData) // No operator specified - should use default Like
	defer GlobalTableIterPool.Put(iter)
	if iter == nil {
		t.Fatalf("Search returned nil iterator for default operator")
	}
	//defer GlobalTableIterPool.Put(iter)

	// Collect results
	var count int
	records := iter.GetRecords(true)
	if records != nil {
		count = len(records)
	}

	// With Like operator, we should find all records since we're searching for id=1 as a prefix
	if count == 0 {
		t.Error("Expected at least some results with default Like operator, got none")
	}
}

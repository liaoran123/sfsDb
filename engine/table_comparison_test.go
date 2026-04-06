package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/util"
)

// TestTableSearchComparisonOperators tests all comparison operators with Table.Search
func TestTableSearchComparisonOperators(t *testing.T) {
	// Create test table
	table, err := TableNew("test_comparison_operators")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index
	pk, _ := NewDefaultPrimaryKey("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Insert test data
	testData := []map[string]any{
		{"id": 1, "name": "Alice", "age": 20, "score": 85.5},
		{"id": 2, "name": "Bob", "age": 25, "score": 90.0},
		{"id": 3, "name": "Charlie", "age": 30, "score": 75.5},
		{"id": 4, "name": "David", "age": 35, "score": 95.0},
		{"id": 5, "name": "Eve", "age": 40, "score": 80.0},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Test cases for comparison operators
	testCases := []struct {
		name          string
		searchData    map[string]any
		operator      util.ComparisonOperator
		expectedCount int
		expectedIDs   []int
	}{
		{
			name:          "Equal",
			searchData:    map[string]any{"id": 3},
			operator:      util.Equal,
			expectedCount: 1,
			expectedIDs:   []int{3},
		},
		{
			name:          "NotEqual",
			searchData:    map[string]any{"id": 3},
			operator:      util.NotEqual,
			expectedCount: 4,
			expectedIDs:   []int{1, 2, 4, 5},
		},
		{
			name:          "GreaterThan",
			searchData:    map[string]any{"id": 2},
			operator:      util.GreaterThan,
			expectedCount: 3,
			expectedIDs:   []int{3, 4, 5},
		},
		{
			name:          "GreaterThanOrEqual",
			searchData:    map[string]any{"id": 2},
			operator:      util.GreaterThanOrEqual,
			expectedCount: 4,
			expectedIDs:   []int{2, 3, 4, 5},
		},
		{
			name:          "LessThan",
			searchData:    map[string]any{"id": 3},
			operator:      util.LessThan,
			expectedCount: 2,
			expectedIDs:   []int{1, 2},
		},
		{
			name:          "LessThanOrEqual",
			searchData:    map[string]any{"id": 3},
			operator:      util.LessThanOrEqual,
			expectedCount: 3,
			expectedIDs:   []int{1, 2, 3},
		},
		{
			name:          "Like (prefix search)",
			searchData:    map[string]any{"id": nil},
			operator:      util.Like,
			expectedCount: 5,
			expectedIDs:   []int{1, 2, 3, 4, 5},
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Search with the specified operator
			iter, err := table.Search(&tc.searchData, tc.operator)
			if err != nil {
				t.Fatalf("Search 失败: %v", err)
			}
			if iter == nil {
				t.Fatalf("Failed to get iterator: %v", err)
			}
			if err != nil && iter == nil {
				t.Fatalf("Search returned nil iterator for operator %s: %v", tc.operator, err)
			}
			defer GlobalTableIterPool.Put(iter)

			// Collect results
			var results []int
			records := iter.GetRecords(true)
			if records != nil {
				for _, record := range records.Select() {
					if id, ok := record["id"].(int); ok {
						results = append(results, id)
					}
				}
			}

			// Verify results
			if len(results) != tc.expectedCount {
				t.Errorf("Expected %d results, got %d", tc.expectedCount, len(results))
			}

			// Verify expected IDs (order may vary for some operators)
			if tc.name != "NotEqual" && tc.name != "Like (prefix search)" {
				for i, expectedID := range tc.expectedIDs {
					if i < len(results) && results[i] != expectedID {
						t.Errorf("Expected ID %d at position %d, got %d", expectedID, i, results[i])
					}
				}
			}
		})
	}
}

// TestTableSearchComparisonOperatorsWithAgeField tests comparison operators with non-primary key field
func TestTableSearchComparisonOperatorsWithAgeField(t *testing.T) {
	// Create test table
	table, err := TableNew("test_comparison_age")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{"id": 0, "name": "", "age": 0}
	err = table.SetFields(fields)
	if err != nil {
		t.Fatalf("Failed to set fields: %v", err)
	}

	// Create primary key index
	pk, _ := NewDefaultPrimaryKey("pk")
	pk.AddFields("id")
	err = table.CreateIndex(pk)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Create age index
	ageIdx, _ := NewDefaultNormalIndex("age_index")
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

	// Test cases for comparison operators on age field
	testCases := []struct {
		name          string
		searchData    map[string]any
		operator      util.ComparisonOperator
		expectedCount int
	}{
		{
			name:          "Age GreaterThan 25",
			searchData:    map[string]any{"age": 25},
			operator:      util.GreaterThan,
			expectedCount: 3,
		},
		{
			name:          "Age LessThanOrEqual 30",
			searchData:    map[string]any{"age": 30},
			operator:      util.LessThanOrEqual,
			expectedCount: 3,
		},
		{
			name:          "Age Equal 35",
			searchData:    map[string]any{"age": 35},
			operator:      util.Equal,
			expectedCount: 1,
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Search with the specified operator
			iter, _ := table.Search(&tc.searchData, tc.operator)
			if iter == nil {
				t.Fatalf("Search returned nil iterator for operator %s", tc.operator)
			}
			defer GlobalTableIterPool.Put(iter)

			// Collect results
			var results []int
			records := iter.GetRecords(true)
			if records != nil {
				for _, record := range records.Select() {
					if id, ok := record["id"].(int); ok {
						results = append(results, id)
					}
				}
			}

			// Verify results
			if len(results) != tc.expectedCount {
				t.Errorf("Expected %d results, got %d", tc.expectedCount, len(results))
			}
		})
	}
}

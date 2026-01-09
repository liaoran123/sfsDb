package engine

import (
	"fmt"
	"testing"

	"github.com/liaoran123/sfsDb/util"
)

// TestTableSearchComprehensive tests various search scenarios
func TestTableSearchComprehensive(t *testing.T) {
	// Create test table
	table, err := TableNew("test_search_comprehensive")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{"id": 0, "name": "", "age": 0, "score": 0.0, "active": false}
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
		{"id": 1, "name": "Alice", "age": 20, "score": 85.5, "active": true},
		{"id": 2, "name": "Bob", "age": 25, "score": 90.0, "active": true},
		{"id": 3, "name": "Charlie", "age": 30, "score": 75.5, "active": false},
		{"id": 4, "name": "David", "age": 35, "score": 95.0, "active": true},
		{"id": 5, "name": "Eve", "age": 40, "score": 80.0, "active": false},
		{"id": 6, "name": "Frank", "age": 45, "score": 88.5, "active": true},
		{"id": 7, "name": "Grace", "age": 50, "score": 92.0, "active": true},
		{"id": 8, "name": "Henry", "age": 55, "score": 78.5, "active": false},
		{"id": 9, "name": "Ivy", "age": 60, "score": 83.0, "active": true},
		{"id": 10, "name": "Jack", "age": 65, "score": 87.5, "active": true},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Test cases for different search scenarios
	testCases := []struct {
		name          string
		searchData    map[string]any
		operator      util.ComparisonOperator
		expectedCount int
	}{
		// Test 1: Exact match on primary key
		{
			name:          "Exact match on id=5",
			searchData:    map[string]any{"id": 5},
			operator:      util.Equal,
			expectedCount: 1,
		},

		// Test 2: Range queries on primary key
		{
			name:          "id > 5",
			searchData:    map[string]any{"id": 5},
			operator:      util.GreaterThan,
			expectedCount: 5, // ids 6-10
		},

		{
			name:          "id >= 3",
			searchData:    map[string]any{"id": 3},
			operator:      util.GreaterThanOrEqual,
			expectedCount: 8, // ids 3-10
		},

		{
			name:          "id < 4",
			searchData:    map[string]any{"id": 4},
			operator:      util.LessThan,
			expectedCount: 3, // ids 1-3
		},

		{
			name:          "id <= 6",
			searchData:    map[string]any{"id": 6},
			operator:      util.LessThanOrEqual,
			expectedCount: 6, // ids 1-6
		},

		// Test 3: Range queries on secondary index (age)
		{
			name:          "age > 40",
			searchData:    map[string]any{"age": 40},
			operator:      util.GreaterThan,
			expectedCount: 5, // ages 45, 50, 55, 60, 65
		},

		{
			name:          "age >= 30",
			searchData:    map[string]any{"age": 30},
			operator:      util.GreaterThanOrEqual,
			expectedCount: 8, // ages 30-65
		},

		{
			name:          "age < 35",
			searchData:    map[string]any{"age": 35},
			operator:      util.LessThan,
			expectedCount: 3, // ages 20, 25, 30
		},

		{
			name:          "age <= 45",
			searchData:    map[string]any{"age": 45},
			operator:      util.LessThanOrEqual,
			expectedCount: 6, // ages 20-45
		},

		// Test 4: Like operator (prefix search)
		{
			name: "Like search on id (prefix)",
			//主键设置为nil，查询所有。如果是组合主键，末尾的字段设置为nil，则同前缀匹配
			searchData:    map[string]any{"id": nil},
			operator:      util.Like,
			expectedCount: 10, // all ids starting with 1 (1, 10)
		},
		{
			name: "Like search on age (prefix)",
			//age设置为nil，查询所有。如果是组合索引，末尾的字段设置为nil，则同前缀匹配
			searchData:    map[string]any{"age": nil},
			operator:      util.Like,
			expectedCount: 10, // all ids starting with 1 (1, 10)
		},

		// Test 5: NotEqual operator
		{
			name:          "age != 45",
			searchData:    map[string]any{"age": 45},
			operator:      util.NotEqual,
			expectedCount: 9, // all except age=45
		},
		// Test 6: NotEqual operator
		{
			name:          "id != 5",
			searchData:    map[string]any{"id": 5},
			operator:      util.NotEqual,
			expectedCount: 9, // all except id=5
		},
	}
	/*
		//--------检测kv数据-----------
		iter := table.For()
		loop := 0
		for iter.Next() {
			fmt.Printf("loop: %v\n", loop)
			loop++
			k, v := iter.Key(), iter.Value()
			rv := table.RecordByteToAny(table.GetPrimaryKey().Parse(nil, v))
			fmt.Printf("record: %v\n", rv)
			rk := table.RecordByteToAny(table.GetPrimaryKey().Parse(nil, k))
			fmt.Printf("record: %v\n", rk)

			fmt.Printf("-----------------------\n")
			if k == nil {
				break
			}
		}
	*/
	//--------------------------
	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Search with the specified operator
			iter := table.Search(&tc.searchData, tc.operator)
			defer iter.Release()
			if iter == nil {
				t.Fatalf("Search returned nil iterator for case: %s", tc.name)
			}
			//defer iter.Release()

			// Collect results
			var count int
			records := iter.GetRecords(true)
			count = len(records)
			/*
				records := iter.GerRecords(true)
				if records != nil {
					count = len(records.records)
				}
			*/
			// Verify results
			if count != tc.expectedCount {
				t.Errorf("Expected %d results for %s, got %d", tc.expectedCount, tc.name, count)
				// Print actual results for debugging
				if records != nil {
					t.Logf("Actual results: %v", records)
				}
			} else {
				fmt.Printf("---------- %v------------------\n", tc.name)
				for i, record := range records {
					fmt.Printf("%v, %v\n", i, record)
				}
			}
		})
	}
}

// TestTableSearchEdgeCases tests edge cases for Table.Search
func TestTableSearchEdgeCases(t *testing.T) {
	// Create test table
	table, err := TableNew("test_search_edge_cases")
	if err != nil {
		t.Fatalf("Failed to create table: %v", err)
	}

	// Set table fields
	fields := map[string]any{"id": 0, "name": "", "value": 0}
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

	// Insert minimal test data
	testData := []map[string]any{
		{"id": 1, "name": "Test1", "value": 100},
		{"id": 2, "name": "Test2", "value": 200},
	}

	for _, data := range testData {
		_, err := table.Insert(&data)
		if err != nil {
			t.Fatalf("Failed to insert test data: %v", err)
		}
	}

	// Test cases for edge cases
	testCases := []struct {
		name          string
		searchData    map[string]any
		operator      util.ComparisonOperator
		expectedCount int
	}{
		// Test 1: Search for non-existent record
		{
			name:          "Search for non-existent id=999",
			searchData:    map[string]any{"id": 999},
			operator:      util.Equal,
			expectedCount: 0,
		},

		// Test 3: Search with zero value
		{
			name:          "Search with id=0",
			searchData:    map[string]any{"id": 0},
			operator:      util.LessThan,
			expectedCount: 0, // no records with id < 0
		},

		// Test 4: Search with empty search data
		{
			name:          "Search with empty search data",
			searchData:    map[string]any{},
			operator:      util.Equal,
			expectedCount: 0, // should return nothing
		},
	}

	// Run each test case
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Search with the specified operator
			iter := table.Search(&tc.searchData, tc.operator)
			defer iter.Release()
			if iter == nil {
				t.Fatalf("Search returned nil iterator for case: %s", tc.name)
			}
			//defer iter.Release()

			// Collect results
			var count int
			records := iter.GetRecords(true)
			if records != nil {
				count = len(records)
			}

			// Verify results
			if count != tc.expectedCount {
				t.Errorf("Expected %d results for %s, got %d", tc.expectedCount, tc.name, count)
			}
		})
	}
}

// TestTableSearchMultipleFields tests search with multiple fields
func TestTableSearchMultipleFields(t *testing.T) {
	// Create test table
	table, err := TableNew("test_search_multiple_fields")
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

	// Test search with multiple fields (should use primary key)
	t.Run("Search with multiple fields (id and age)", func(t *testing.T) {
		searchData := map[string]any{"id": 3, "age": 30}
		iter := table.Search(&searchData, util.Equal)
		defer iter.Release()
		if iter == nil {
			t.Fatalf("Search returned nil iterator")
		}
		//defer iter.Release()

		// Collect results
		var count int
		records := iter.GetRecords(true)
		if records != nil {
			count = len(records)
		}

		// Should find the exact match
		if count != 1 {
			t.Errorf("Expected 1 result, got %d", count)
		}
	})
}

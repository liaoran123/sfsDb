package engine

import (
	"testing"
)

func TestIndexsCreateIndex(t *testing.T) {
	// Create test fields map
	fields := map[string]any{
		"id":   nil,
		"name": nil,
		"age":  nil,
	}

	// Create Indexs instance
	indexs := NewIndexs(&fields)

	// Create test indexes
	primaryKey, err := DefaultPrimaryKeyNew("primary")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	primaryKey.AddFields("id")

	normalIndex, err := DefaultNormalIndexNew("normal")
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}
	normalIndex.AddFields("name")

	// Test 1: Create primary key index
	err = indexs.CreateIndex(primaryKey)
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}

	// Test 2: Create normal index
	err = indexs.CreateIndex(normalIndex)
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}

	// Test 3: Create duplicate index with same name
	duplicateIndex, err := DefaultNormalIndexNew("normal")
	if err != nil {
		t.Fatalf("Failed to create duplicate index: %v", err)
	}
	duplicateIndex.AddFields("age")
	err = indexs.CreateIndex(duplicateIndex)
	if err == nil {
		t.Error("Expected error when creating duplicate index name, got nil")
	}

	// Test 4: Create index with identical fields
	identicalIndex, err := DefaultNormalIndexNew("normal2")
	if err != nil {
		t.Fatalf("Failed to create identical index: %v", err)
	}
	identicalIndex.AddFields("name")
	err = indexs.CreateIndex(identicalIndex)
	if err == nil {
		t.Error("Expected error when creating index with identical fields, got nil")
	}

	// Test 5: Create index with non-existent field
	invalidIndex, err := DefaultNormalIndexNew("invalid")
	if err != nil {
		t.Fatalf("Failed to create invalid index: %v", err)
	}
	invalidIndex.AddFields("nonexistent")
	err = indexs.CreateIndex(invalidIndex)
	if err == nil {
		t.Error("Expected error when creating index with non-existent field, got nil")
	}
}

func TestIndexsMatchIndex(t *testing.T) {
	// Create test fields map
	fields := map[string]any{
		"id":   nil,
		"name": nil,
		"age":  nil,
	}

	// Create Indexs instance
	indexs := NewIndexs(&fields)

	// Create test indexes
	primaryKey, err := DefaultPrimaryKeyNew("primary")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	primaryKey.AddFields("id")

	normalIndex, err := DefaultNormalIndexNew("normal")
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}
	normalIndex.AddFields("name")

	fullTextIndex, err := DefaultFullTextIndexNew("ft")
	if err != nil {
		t.Fatalf("Failed to create full text index: %v", err)
	}
	fullTextIndex.AddFields("name")

	// Add indexes to Indexs instance
	indexs.CreateIndex(primaryKey)
	indexs.CreateIndex(normalIndex)
	indexs.CreateIndex(fullTextIndex)

	// Test 1: Match primary key first
	matchedIndex := indexs.MatchIndex("id")
	if matchedIndex == nil {
		t.Fatal("Expected to match primary key index")
	}
	if matchedIndex.Name() != "primary" {
		t.Errorf("Expected primary key index, got %s", matchedIndex.Name())
	}

	// Test 2: Match normal index when primary key not matching
	matchedIndex = indexs.MatchIndex("name")
	if matchedIndex == nil {
		t.Fatal("Expected to match normal index")
	}
	if matchedIndex.Name() != "normal" {
		t.Errorf("Expected normal index, got %s", matchedIndex.Name())
	}

	// Test 3: Match full text index when no other matches
	// Create a full text index with a different field
	fullTextIndex2, err := DefaultFullTextIndexNew("ft2")
	if err != nil {
		t.Fatalf("Failed to create full text index ft2: %v", err)
	}
	fullTextIndex2.AddFields("age")
	indexs.CreateIndex(fullTextIndex2)

	matchedIndex = indexs.MatchIndex("age")
	if matchedIndex == nil {
		t.Fatal("Expected to match full text index")
	}
	if matchedIndex.Name() != "ft2" {
		t.Errorf("Expected full text index ft2, got %s", matchedIndex.Name())
	}

	// Test 4: Return nil when no match
	matchedIndex = indexs.MatchIndex("nonexistent")
	if matchedIndex != nil {
		t.Errorf("Expected nil, got %s", matchedIndex.Name())
	}
}

func TestIndexsDeleteIndex(t *testing.T) {
	// Create test fields map
	fields := map[string]any{
		"id":   nil,
		"name": nil,
	}

	// Create Indexs instance
	indexs := NewIndexs(&fields)

	// Create test index
	normalIndex, err := DefaultNormalIndexNew("normal")
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}
	normalIndex.AddFields("name")

	// Add index
	indexs.CreateIndex(normalIndex)

	// Test 1: Delete existing index
	err = indexs.DeleteIndex("normal")
	if err != nil {
		t.Fatalf("Failed to delete existing index: %v", err)
	}

	// Test 2: Delete non-existent index
	err = indexs.DeleteIndex("non_existent")
	if err == nil {
		t.Error("Expected error when deleting non-existent index, got nil")
	}
}

func TestIndexsNewAndUtilityMethods(t *testing.T) {
	// Create test fields map
	fields := map[string]any{
		"id":   nil,
		"name": nil,
		"age":  nil,
	}

	// Test 1: Use NewIndexs constructor
	indexs := NewIndexs(&fields)
	if indexs == nil {
		t.Fatal("Expected NewIndexs to return non-nil")
	}

	// Test 2: Check initial Len() is 0
	if indexs.Len() != 0 {
		t.Errorf("Expected initial Len() to be 0, got %d", indexs.Len())
	}

	// Create and add test indexes
	primaryKey, err := DefaultPrimaryKeyNew("primary")
	if err != nil {
		t.Fatalf("Failed to create primary key index: %v", err)
	}
	primaryKey.AddFields("id")

	normalIndex, err := DefaultNormalIndexNew("normal")
	if err != nil {
		t.Fatalf("Failed to create normal index: %v", err)
	}
	normalIndex.AddFields("name")

	indexs.CreateIndex(primaryKey)
	indexs.CreateIndex(normalIndex)

	// Test 3: Check Len() after adding indexes
	if indexs.Len() != 2 {
		t.Errorf("Expected Len() to be 2, got %d", indexs.Len())
	}

	// Test 4: GetIndex by name
	retrievedIndex := indexs.GetIndex("primary")
	if retrievedIndex == nil {
		t.Fatal("Expected GetIndex to return primary key index")
	}
	if retrievedIndex.Name() != "primary" {
		t.Errorf("Expected primary key index, got %s", retrievedIndex.Name())
	}

	// Test 5: Get non-existent index
	retrievedIndex = indexs.GetIndex("non_existent")
	if retrievedIndex != nil {
		t.Error("Expected GetIndex to return nil for non-existent index")
	}

	// Test 6: GetAllIndexes
	allIndexes := indexs.GetAllIndexes()
	if len(allIndexes) != 2 {
		t.Errorf("Expected GetAllIndexes to return 2 indexes, got %d", len(allIndexes))
	}
}

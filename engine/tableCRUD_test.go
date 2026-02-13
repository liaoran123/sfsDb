package engine

import (
	"testing"

	"github.com/liaoran123/sfsDb/storage"
)

func TestSearchs(t *testing.T) {
	// Create a test table
	table := &Table{
		name: "test_table",
		fields: map[string]any{
			"id":   "int",
			"name": "string",
		},
	}

	// Create a simple mock FunIter
	funIter := func(start, limit []byte) storage.Iterator {
		return &mockIterator{}
	}

	// Test Searchs with a field
	fields := &map[string]any{
		"id": 1,
	}

	iter, err := table.Searchs(funIter, fields)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}

	if iter == nil {
		t.Error("Expected non-nil TableIter")
	}

	// Put back to pool
	GlobalTableIterPool.Put(iter)
}

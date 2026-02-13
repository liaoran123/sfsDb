package engine

import (
	"testing"
)

func TestTableIterPool(t *testing.T) {
	// Create a simple mock Table
	table := &Table{
		name: "test_table",
		fields: map[string]any{
			"id":   "int",
			"name": "string",
		},
	}

	// Create a mock iterator
	mockIter := &mockIterator{}

	// Create a mock index
	mockIndex := &BaseIndex{
		fields: []string{"id"},
		name:   "primary_key",
	}

	// Test getting from pool
	iter := GlobalTableIterPool.Get(table, mockIter, mockIndex)
	if iter == nil {
		t.Error("Expected non-nil TableIter")
	}

	// Test putting back to pool
	GlobalTableIterPool.Put(iter)

	// Test getting again
	iter2 := GlobalTableIterPool.Get(table, mockIter, mockIndex)
	if iter2 == nil {
		t.Error("Expected non-nil TableIter on second get")
	}

	GlobalTableIterPool.Put(iter2)
}

// mockIterator is a simple mock implementation of storage.Iterator
type mockIterator struct{}

func (mi *mockIterator) First() bool          { return false }
func (mi *mockIterator) Last() bool           { return false }
func (mi *mockIterator) Seek(key []byte) bool { return false }
func (mi *mockIterator) Next() bool           { return false }
func (mi *mockIterator) Prev() bool           { return false }
func (mi *mockIterator) Key() []byte          { return nil }
func (mi *mockIterator) Value() []byte        { return nil }
func (mi *mockIterator) Valid() bool          { return false }
func (mi *mockIterator) Release()             {}

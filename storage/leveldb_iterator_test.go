package storage

import (
	"os"
	"testing"
)

func TestLevelDBIterator(t *testing.T) {
	// Create a temporary directory for the test
	tempDir, err := os.MkdirTemp("", "leveldb_test")
	if err != nil {
		t.Fatalf("Failed to create temp directory: %v", err)
	}
	defer os.RemoveAll(tempDir)
	
	// Create a LevelDB store
	config := StoreConfig{
		Path: tempDir,
	}
	store, err := NewLevelDBStore(config)
	if err != nil {
		t.Fatalf("Failed to create LevelDB store: %v", err)
	}
	defer store.Close()
	
	// Put some test data
	testData := map[string]string{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	
	for k, v := range testData {
		if err := store.Put([]byte(k), []byte(v)); err != nil {
			t.Fatalf("Failed to put key %s: %v", k, err)
		}
	}
	
	// Test Iterator with no parameters (full scan)
	iter := store.Iterator()
	if iter == nil {
		t.Fatal("Iterator returned nil")
	}
	
	// Test First() method
	if !iter.First() {
		t.Error("First() returned false, expected true")
	} else {
		key := string(iter.Key())
		value := string(iter.Value())
		if key != "key1" || value != "value1" {
			t.Errorf("First() returned unexpected key-value: %s=%s, expected key1=value1", key, value)
		}
	}
	
	// Test Next() method
	if !iter.Next() {
		t.Error("Next() returned false, expected true")
	} else {
		key := string(iter.Key())
		value := string(iter.Value())
		if key != "key2" || value != "value2" {
			t.Errorf("Next() returned unexpected key-value: %s=%s, expected key2=value2", key, value)
		}
	}
	
	// Test Last() method
	if !iter.Last() {
		t.Error("Last() returned false, expected true")
	} else {
		key := string(iter.Key())
		value := string(iter.Value())
		if key != "key3" || value != "value3" {
			t.Errorf("Last() returned unexpected key-value: %s=%s, expected key3=value3", key, value)
		}
	}
	
	// Test Prev() method
	if !iter.Prev() {
		t.Error("Prev() returned false, expected true")
	} else {
		key := string(iter.Key())
		value := string(iter.Value())
		if key != "key2" || value != "value2" {
			t.Errorf("Prev() returned unexpected key-value: %s=%s, expected key2=value2", key, value)
		}
	}
	
	// Test Seek() method
	if !iter.Seek([]byte("key2")) {
		t.Error("Seek(key2) returned false, expected true")
	} else {
		key := string(iter.Key())
		value := string(iter.Value())
		if key != "key2" || value != "value2" {
			t.Errorf("Seek(key2) returned unexpected key-value: %s=%s, expected key2=value2", key, value)
		}
	}
	
	// Test Valid() method
	if !iter.Valid() {
		t.Error("Valid() returned false, expected true")
	}
	
	// Release the iterator
	iter.Release()
	
	// Test prefix scan with "key"
	prefixIter := store.Iterator([]byte("key"))
	if prefixIter == nil {
		t.Fatal("Prefix Iterator returned nil")
	}
	
	count := 0
	for prefixIter.First(); prefixIter.Valid(); prefixIter.Next() {
		count++
	}
	if count != 3 {
		t.Errorf("Prefix scan returned %d records, expected 3", count)
	}
	prefixIter.Release()
	
	t.Log("All iterator tests passed")
}

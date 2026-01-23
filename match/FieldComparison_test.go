package match

import (
	"testing"
	"time"
)

// TestFieldComparison_BasicOperators tests basic comparison operators
func TestFieldComparison_BasicOperators(t *testing.T) {
	// Test data
	fields := map[string]any{
		"id":   10,
		"name": "test",
		"age":  25,
		"score": 85.5,
		"active": true,
	}

	// Test cases for basic comparison operators
	testCases := []struct {
		name       string
		comparison *FieldComparison
		expected   bool
	}{
		// Equal operator tests
		{"Equal match on existing field", NewFieldComparison("id", Equal, 10), true},
		{"Equal match on non-existing field", NewFieldComparison("nonexistent", Equal, 10), false},
		{"Equal mismatch", NewFieldComparison("id", Equal, 20), false},
		
		// NotEqual operator tests
		{"NotEqual match", NewFieldComparison("id", NotEqual, 20), true},
		{"NotEqual mismatch", NewFieldComparison("id", NotEqual, 10), false},
		{"NotEqual on non-existing field", NewFieldComparison("nonexistent", NotEqual, 10), false},
		
		// GreaterThan operator tests
		{"GreaterThan match", NewFieldComparison("id", GreaterThan, 5), true},
		{"GreaterThan mismatch", NewFieldComparison("id", GreaterThan, 15), false},
		{"GreaterThan equal", NewFieldComparison("id", GreaterThan, 10), false},
		
		// GreaterThanOrEqual operator tests
		{"GreaterThanOrEqual match greater", NewFieldComparison("id", GreaterThanOrEqual, 5), true},
		{"GreaterThanOrEqual match equal", NewFieldComparison("id", GreaterThanOrEqual, 10), true},
		{"GreaterThanOrEqual mismatch", NewFieldComparison("id", GreaterThanOrEqual, 15), false},
		
		// LessThan operator tests
		{"LessThan match", NewFieldComparison("id", LessThan, 15), true},
		{"LessThan mismatch", NewFieldComparison("id", LessThan, 5), false},
		{"LessThan equal", NewFieldComparison("id", LessThan, 10), false},
		
		// LessThanOrEqual operator tests
		{"LessThanOrEqual match less", NewFieldComparison("id", LessThanOrEqual, 15), true},
		{"LessThanOrEqual match equal", NewFieldComparison("id", LessThanOrEqual, 10), true},
		{"LessThanOrEqual mismatch", NewFieldComparison("id", LessThanOrEqual, 5), false},
		
		// Float comparison tests
		{"Float Equal match", NewFieldComparison("score", Equal, 85.5), true},
		{"Float GreaterThan match", NewFieldComparison("score", GreaterThan, 80.0), true},
		{"Float LessThanOrEqual match", NewFieldComparison("score", LessThanOrEqual, 90.0), true},
		
		// Bool comparison tests
		{"Bool Equal true", NewFieldComparison("active", Equal, true), true},
		{"Bool Equal false", NewFieldComparison("active", Equal, false), false},
		{"Bool NotEqual match", NewFieldComparison("active", NotEqual, false), true},
	}

	// Run all test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.comparison.Match(&fields)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestFieldComparison_StringOperators tests string-specific comparison operators
func TestFieldComparison_StringOperators(t *testing.T) {
	// Test data
	fields := map[string]any{
		"name":  "test_string",
		"email": "test@example.com",
		"empty": "",
	}

	// Test cases for string comparison operators
	testCases := []struct {
		name       string
		comparison *FieldComparison
		expected   bool
	}{
		// Like operator tests
		{"Like full match", NewFieldComparison("name", Like, "test_string"), true},
		{"Like prefix match", NewFieldComparison("name", Like, "test%"), true},
		{"Like suffix match", NewFieldComparison("name", Like, "%string"), true},
		{"Like middle match", NewFieldComparison("name", Like, "%_string"), true},
		{"Like no match", NewFieldComparison("name", Like, "no_match%"), false},
		
		// Prefix operator tests
		{"Prefix match", NewFieldComparison("name", Prefix, "test"), true},
		{"Prefix mismatch", NewFieldComparison("name", Prefix, "invalid"), false},
		{"Prefix empty string", NewFieldComparison("name", Prefix, ""), true},
		{"Prefix match on empty field", NewFieldComparison("empty", Prefix, "test"), false},
		
		// Suffix operator tests
		{"Suffix match", NewFieldComparison("name", Suffix, "string"), true},
		{"Suffix mismatch", NewFieldComparison("name", Suffix, "invalid"), false},
		{"Suffix empty string", NewFieldComparison("name", Suffix, ""), true},
		{"Suffix match on empty field", NewFieldComparison("empty", Suffix, "test"), false},
		
		// Contains operator tests
		{"Contains match", NewFieldComparison("name", Contains, "_string"), true},
		{"Contains partial match", NewFieldComparison("name", Contains, "string"), true},
		{"Contains mismatch", NewFieldComparison("name", Contains, "invalid"), false},
		{"Contains empty string", NewFieldComparison("name", Contains, ""), true},
	}

	// Run all test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.comparison.Match(&fields)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestFieldComparison_HelperFunctions tests the helper functions for creating comparisons
func TestFieldComparison_HelperFunctions(t *testing.T) {
	// Test data
	fields := map[string]any{
		"id":   10,
		"name": "test",
		"age":  25,
	}

	// Test cases for helper functions
	testCases := []struct {
		name       string
		comparison *FieldComparison
		expected   bool
	}{
		// Helper function tests
		{"NewEqualMatch helper", NewEqualMatch("id", 10), true},
		{"NewNotEqualMatch helper", NewNotEqualMatch("id", 20), true},
		{"NewGreaterThanMatch helper", NewGreaterThanMatch("id", 5), true},
		{"NewGreaterThanOrEqualMatch helper", NewGreaterThanOrEqualMatch("id", 10), true},
		{"NewLessThanMatch helper", NewLessThanMatch("id", 15), true},
		{"NewLessThanOrEqualMatch helper", NewLessThanOrEqualMatch("id", 10), true},
		{"NewLikeMatch helper", NewLikeMatch("name", "test%"), true},
		{"NewPrefixMatch helper", NewPrefixMatch("name", "test"), true},
		{"NewSuffixMatch helper", NewSuffixMatch("name", "test"), true},
		{"NewContainsMatch helper", NewContainsMatch("name", "test"), true},
	}

	// Run all test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.comparison.Match(&fields)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestFieldComparison_EdgeCases tests edge cases for FieldComparison
func TestFieldComparison_EdgeCases(t *testing.T) {
	// Test cases for edge cases
	testCases := []struct {
		name       string
		fields     *map[string]any
		comparison *FieldComparison
		expected   bool
	}{
		// Nil fields map
		{"Nil fields map", nil, NewFieldComparison("id", Equal, 10), false},
		
		// Different types comparison
		{"Int vs float comparison", &map[string]any{"id": 10}, NewFieldComparison("id", Equal, 10.0), true},
		{"Int vs string comparison", &map[string]any{"id": 10}, NewFieldComparison("id", Equal, "10"), false},
		
		// Time comparison
		{"Time Equal comparison", &map[string]any{"created_at": time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)}, 
		 NewFieldComparison("created_at", Equal, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)), true},
		
		{"Time GreaterThan comparison", &map[string]any{"created_at": time.Date(2023, 1, 2, 0, 0, 0, 0, time.UTC)}, 
		 NewFieldComparison("created_at", GreaterThan, time.Date(2023, 1, 1, 0, 0, 0, 0, time.UTC)), true},
	}

	// Run all test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.comparison.Match(tc.fields)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestFieldMatch tests the FieldMatch struct functionality
func TestFieldMatch(t *testing.T) {
	// Test data
	fields := map[string]any{
		"id":   10,
		"name": "test",
		"age":  25,
	}

	// Test cases for FieldMatch
	testCases := []struct {
		name     string
		fieldMatch *FieldMatch
		expected bool
	}{
		// Basic FieldMatch tests
		{"FieldMatch even id", NewFieldMatch("id", func(val any) bool {
			if id, ok := val.(int); ok {
				return id%2 == 0
			}
			return false
		}), true},
		
		{"FieldMatch name length", NewFieldMatch("name", func(val any) bool {
			if name, ok := val.(string); ok {
				return len(name) > 3
			}
			return false
		}), true},
		
		{"FieldMatch age range", NewFieldMatch("age", func(val any) bool {
			if age, ok := val.(int); ok {
				return age >= 18 && age <= 30
			}
			return false
		}), true},
		
		{"FieldMatch mismatch", NewFieldMatch("id", func(val any) bool {
			if id, ok := val.(int); ok {
				return id > 100
			}
			return false
		}), false},
		
		{"FieldMatch non-existing field", NewFieldMatch("nonexistent", func(val any) bool {
			return true
		}), false},
	}

	// Run all test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := tc.fieldMatch.Match(&fields)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v", tc.expected, result)
			}
		})
	}
}

// TestFieldComparison_TypeConversion tests type conversion in comparisons
func TestFieldComparison_TypeConversion(t *testing.T) {
	// Test data with mixed types
	fields := map[string]any{
		"int_id":    10,
		"int8_id":   int8(10),
		"int16_id":  int16(10),
		"int32_id":  int32(10),
		"int64_id":  int64(10),
		"uint_id":   uint(10),
		"float32_id": float32(10.0),
		"float64_id": 10.0,
	}

	// Test cases for type conversion
	testCases := []struct {
		name       string
		fieldName  string
		value      any
		expected   bool
	}{
		// Type conversion tests
		{"Int to int8 conversion", "int_id", int8(10), true},
		{"Int to int16 conversion", "int_id", int16(10), true},
		{"Int to int32 conversion", "int_id", int32(10), true},
		{"Int to int64 conversion", "int_id", int64(10), true},
		{"Int to uint conversion", "int_id", uint(10), true},
		{"Int to float32 conversion", "int_id", float32(10.0), true},
		{"Int to float64 conversion", "int_id", 10.0, true},
		
		{"Int8 to int conversion", "int8_id", 10, true},
		{"Int16 to int conversion", "int16_id", 10, true},
		{"Float32 to float64 conversion", "float32_id", 10.0, true},
	}

	// Run all test cases
	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			comparison := NewEqualMatch(tc.fieldName, tc.value)
			result := comparison.Match(&fields)
			if result != tc.expected {
				t.Errorf("Expected %v, got %v for %s", tc.expected, result, tc.name)
			}
		})
	}
}

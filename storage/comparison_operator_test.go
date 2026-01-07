package storage

import (
	"testing"
)

// TestComparisonOperatorEnum tests that the ComparisonOperator enum values are correct
func TestComparisonOperatorEnum(t *testing.T) {
	// Test that all operators have distinct values
	operators := []ComparisonOperator{
		Equal,
		NotEqual,
		GreaterThan,
		GreaterThanOrEqual,
		LessThan,
		LessThanOrEqual,
		Like,
	}

	// Check for distinct values
	seen := make(map[ComparisonOperator]bool)
	for _, op := range operators {
		if seen[op] {
			t.Errorf("Duplicate operator value: %d", op)
		}
		seen[op] = true
	}

	// Test string representation
	testCases := []struct {
		operator ComparisonOperator
		expected string
	}{
		{Equal, "="},
		{NotEqual, "!="},
		{GreaterThan, ">"},
		{GreaterThanOrEqual, ">="},
		{LessThan, "<"},
		{LessThanOrEqual, "<="},
		{Like, "LIKE"},
		{ComparisonOperator(99), "unknown"}, // Test unknown operator
	}

	for _, tc := range testCases {
		if tc.operator.String() != tc.expected {
			t.Errorf("Expected operator %d to stringify as %q, got %q", tc.operator, tc.expected, tc.operator.String())
		}
	}
}

// TestComparisonOperatorBasicFunctionality tests basic functionality of comparison operators
func TestComparisonOperatorBasicFunctionality(t *testing.T) {
	rangeHelper := NewRangeHelper(nil)
	testKey := []byte("test_key_123")

	// Test that FromComparison returns non-nil ranges for all operators
	operators := []ComparisonOperator{
		Equal,
		NotEqual,
		GreaterThan,
		GreaterThanOrEqual,
		LessThan,
		LessThanOrEqual,
		Like,
	}

	for _, op := range operators {
		result := rangeHelper.FromComparison(op, testKey)
		if result == nil && op != NotEqual {
			t.Errorf("FromComparison returned nil for operator %s", op)
		}
	}

	// Test specific operators
	// Equal should return a range that includes only the exact key
	equalRange := rangeHelper.FromComparison(Equal, testKey)
	if equalRange == nil {
		t.Error("FromComparison returned nil for Equal operator")
	}

	// GreaterThan should return a range that starts after the key
	gtRange := rangeHelper.FromComparison(GreaterThan, testKey)
	if gtRange == nil {
		t.Error("FromComparison returned nil for GreaterThan operator")
	}

	// LessThan should return a range that ends before the key
	ltRange := rangeHelper.FromComparison(LessThan, testKey)
	if ltRange == nil {
		t.Error("FromComparison returned nil for LessThan operator")
	}
}

package storage

import (
	"testing"
)

// TestComparisonOperatorString tests the string representation of comparison operators
func TestComparisonOperatorString(t *testing.T) {
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
		t.Run(tc.expected, func(t *testing.T) {
			result := tc.operator.String()
			if result != tc.expected {
				t.Errorf("Expected %q, got %q", tc.expected, result)
			}
		})
	}
}

// TestComparisonOperatorValues tests that all comparison operators have distinct values
func TestComparisonOperatorValues(t *testing.T) {
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
}
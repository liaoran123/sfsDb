package util

import (
	"fmt"
	"testing"
)

func TestCombineStrings(t *testing.T) {
	// Test case 1: First array with multiple arrays
	arrays := [][]string{
		{"a", "b"}, // 第一个数组
		{"1", "2"}, // 第二个数组
		{"x", "y"}, // 第三个数组
	}

	// Expected: a1x, a1y, a2x, a2y, b1x, b1y, b2x, b2y
	expected := []string{
		"a1x", "a1y", "a2x", "a2y",
		"b1x", "b1y", "b2x", "b2y",
	}

	result := CombineStrings(arrays, "")
	if len(result) != len(expected) {
		t.Errorf("Expected %d results, got %d", len(expected), len(result))
		return
	}
	fmt.Printf("result: %v\n", result)
	// Check if all expected results are in the actual result
	for _, exp := range expected {
		found := false
		for _, res := range result {
			if res == exp {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected result %s not found", exp)
		}
	}

	// Test case 2: Only one array
	singleArray := [][]string{{"x", "y", "z"}}
	singleResult := CombineStrings(singleArray, "")
	expectedSingle := []string{"x", "y", "z"}
	if len(singleResult) != len(expectedSingle) {
		t.Errorf("Expected %d results for single array, got %d", len(expectedSingle), len(singleResult))
		return
	}

	// Check if results match expected
	for i, res := range singleResult {
		if res != expectedSingle[i] {
			t.Errorf("Expected result %s at index %d, got %s", expectedSingle[i], i, res)
		}
	}

	// Test case 3: Empty input
	emptyArray := [][]string{}
	emptyResult := CombineStrings(emptyArray, "")
	if len(emptyResult) != 0 {
		t.Errorf("Expected 0 results for empty input, got %d", len(emptyResult))
	}

	// Test case 4: Multiple arrays with different lengths
	diffArrays := [][]string{
		{"p", "q"},
		{"r"},
		{"s", "t", "u"},
	}
	diffResult := CombineStrings(diffArrays, "")
	// Expected: prs, prt, pru, qrs, qrt, qru
	if len(diffResult) != 6 {
		t.Errorf("Expected 6 results for different length arrays, got %d", len(diffResult))
	}

	// Test case 5: First array with one element, multiple other arrays
	singleStart := [][]string{
		{"hello"},
		{", "},
		{"world", "there"},
		{"!"},
	}
	singleStartResult := CombineStrings(singleStart, "")
	// Expected: hello, world!, hello, there!
	if len(singleStartResult) != 2 {
		t.Errorf("Expected 2 results for single start array, got %d", len(singleStartResult))
	}

	t.Log("All tests passed!")
}

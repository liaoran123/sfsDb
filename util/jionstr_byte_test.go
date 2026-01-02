package util

import (
	"testing"
)

func TestCombineBytes(t *testing.T) {
	// Test case 1: Multiple byte arrays with empty separator
	arrays := [][][]byte{
		{[]byte("a"), []byte("b")}, // 第一个数组
		{[]byte("1"), []byte("2")}, // 第二个数组
		{[]byte("x"), []byte("y")}, // 第三个数组
	}

	// Expected: a1x, a1y, a2x, a2y, b1x, b1y, b2x, b2y
	expected := [][]byte{
		[]byte("a1x"), []byte("a1y"), []byte("a2x"), []byte("a2y"),
		[]byte("b1x"), []byte("b1y"), []byte("b2x"), []byte("b2y"),
	}

	result := CombineBytes(arrays, []byte(""))
	if len(result) != len(expected) {
		t.Errorf("Expected %d results, got %d", len(expected), len(result))
		return
	}

	// Check if all expected results are in the actual result
	for _, exp := range expected {
		found := false
		for _, res := range result {
			if string(res) == string(exp) {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Expected result %s not found", string(exp))
		}
	}

	// Test case 2: With separator
	sepResult := CombineBytes(arrays, []byte(","))
	sepExpected := [][]byte{
		[]byte("a,1,x"), []byte("a,1,y"), []byte("a,2,x"), []byte("a,2,y"),
		[]byte("b,1,x"), []byte("b,1,y"), []byte("b,2,x"), []byte("b,2,y"),
	}

	if len(sepResult) != len(sepExpected) {
		t.Errorf("Expected %d results with separator, got %d", len(sepExpected), len(sepResult))
		return
	}

	// Test case 3: Only one array
	singleArray := [][][]byte{{[]byte("x"), []byte("y"), []byte("z")}}
	singleResult := CombineBytes(singleArray, []byte(","))
	expectedSingle := [][]byte{[]byte("x"), []byte("y"), []byte("z")}
	if len(singleResult) != len(expectedSingle) {
		t.Errorf("Expected %d results for single array, got %d", len(expectedSingle), len(singleResult))
		return
	}

	// Check if results match expected
	for i, res := range singleResult {
		if string(res) != string(expectedSingle[i]) {
			t.Errorf("Expected result %s at index %d, got %s", string(expectedSingle[i]), i, string(res))
		}
	}

	// Test case 4: Empty input
	emptyArray := [][][]byte{}
	emptyResult := CombineBytes(emptyArray, []byte(","))
	if len(emptyResult) != 0 {
		t.Errorf("Expected 0 results for empty input, got %d", len(emptyResult))
	}

	t.Log("All CombineBytes tests passed!")
}
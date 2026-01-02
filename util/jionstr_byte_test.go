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
	prefix := []byte("ftpxf")

	// Expected: a1x, a1y, a2x, a2y, b1x, b1y, b2x, b2y
	expected := [][]byte{
		[]byte("ftpxfa1x"), []byte("ftpxfa1y"), []byte("ftpxfa2x"), []byte("ftpxfa2y"),
		[]byte("ftpxfb1x"), []byte("ftpxfb1y"), []byte("ftpxfb2x"), []byte("ftpxfb2y"),
	}

	result := CombineBytes(arrays, []byte(""), prefix)
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
	sepResult := CombineBytes(arrays, []byte(","), prefix)
	sepExpected := [][]byte{
		[]byte("ftpxfa,1,x"), []byte("ftpxfa,1,y"), []byte("ftpxfa,2,x"), []byte("ftpxfa,2,y"),
		[]byte("ftpxfb,1,x"), []byte("ftpxfb,1,y"), []byte("ftpxfb,2,x"), []byte("ftpxfb,2,y"),
	}

	if len(sepResult) != len(sepExpected) {
		t.Errorf("Expected %d results with separator, got %d", len(sepExpected), len(sepResult))
		return
	}

	// Test case 3: Only one array
	singleArray := [][][]byte{{[]byte("x"), []byte("y"), []byte("z")}}
	singleResult := CombineBytes(singleArray, []byte(","), prefix)
	expectedSingle := [][]byte{[]byte("ftpxfx"), []byte("ftpxfy"), []byte("ftpxfz")}
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
	emptyResult := CombineBytes(emptyArray, []byte(","), prefix)
	if len(emptyResult) != 0 {
		t.Errorf("Expected 0 results for empty input, got %d", len(emptyResult))
	}

	t.Log("All CombineBytes tests passed!")
}

package tests

import "testing"

// sortIntSlice sorts a slice of integers in-place.
func sortIntSlice(a []int) {
	for i := 0; i < len(a)-1; i++ {
		for j := i + 1; j < len(a); j++ {
			if a[i] > a[j] {
				a[i], a[j] = a[j], a[i]
			}
		}
	}
}

// assertEqualSlices asserts that two int slices are equal.
func assertEqualSlices(t *testing.T, expected, actual []int) {
	if len(expected) != len(actual) {
		t.Errorf("Expected %v, got %v", expected, actual)
		return
	}
	for i := range expected {
		if expected[i] != actual[i] {
			t.Errorf("Expected %v, got %v", expected, actual)
			return
		}
	}
}

// assertEqualMaps asserts that two maps are equal.
func assertEqualMaps(t *testing.T, expected, actual map[string]map[string]interface{}) {
	for field, expectedVal := range expected {
		actualVal, ok := actual[field]
		if !ok {
			t.Errorf("Missing field %s in result", field)
			continue
		}
		for expKey, expCount := range expectedVal {
			actCount, ok := actualVal[expKey]
			if !ok {
				t.Errorf("Field %s: missing key %v", field, expKey)
				continue
			}
			if actCount != expCount {
				t.Errorf("Field %s[%v]: expected count %v, got %v", field, expKey, expCount, actCount)
			}
		}
	}
}

// assertEqualFacetDataStringSlice asserts that two facet data maps with string slices are equal.
func assertEqualFacetDataStringSlice(t *testing.T, expected, actual map[string]map[string][]int) {
	for field, expectedVal := range expected {
		actualVal, ok := actual[field]
		if !ok {
			t.Errorf("Missing field %s in result", field)
			continue
		}
		for expKey, expRecords := range expectedVal {
			actRecords, ok := actualVal[expKey]
			if !ok {
				t.Errorf("Field %s: missing key %v", field, expKey)
				continue
			}
			if len(expRecords) != len(actRecords) {
				t.Errorf("Field %s[%v]: expected %d records, got %d", field, expKey, len(expRecords), len(actRecords))
				continue
			}
			recordMap := make(map[int]bool)
			for _, r := range actRecords {
				recordMap[r] = true
			}
			for _, r := range expRecords {
				if !recordMap[r] {
					t.Errorf("Field %s[%v]: missing record %d", field, expKey, r)
				}
			}
		}
	}
}

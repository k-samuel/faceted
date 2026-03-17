package value

import (
	"testing"
)

// TestNewConverter tests creating a new Converter.
func TestNewConverter(t *testing.T) {
	c := NewConverter()
	if c == nil {
		t.Errorf("Expected non-nil converter")
	}
}

// TestGetValueString tests GetValueString with various types.
func TestGetValueString(t *testing.T) {
	c := NewConverter()

	tests := []struct {
		input    interface{}
		expected string
		hasError bool
	}{
		{true, "1", false},
		{false, "0", false},
		{42, "42", false},
		{int64(12345), "12345", false},
		{"hello", "hello", false},
		{3.14, "3.14", false},
		{float32(2.5), "2.5", false},
		{nil, "", true}, // Should return error
	}

	for _, tt := range tests {
		result, err := c.GetValueString(tt.input)
		if tt.hasError {
			if err == nil {
				t.Errorf("GetValueString(%v): expected error, got none", tt.input)
			}
		} else {
			if err != nil {
				t.Errorf("GetValueString(%v): unexpected error: %v", tt.input, err)
			}
			if result != tt.expected {
				t.Errorf("GetValueString(%v): expected %q, got %q", tt.input, tt.expected, result)
			}
		}
	}
}

// TestValueToStringSlice tests ValueToStringSlice with various types.
func TestValueToStringSlice(t *testing.T) {
	c := NewConverter()

	tests := []struct {
		name     string
		input    interface{}
		expected []string
		hasError bool
	}{
		{
			"single int",
			42,
			[]string{"42"},
			false,
		},
		{
			"single string",
			"hello",
			[]string{"hello"},
			false,
		},
		{
			"[]string",
			[]string{"a", "b", "c"},
			[]string{"a", "b", "c"},
			false,
		},
		{
			"[]int",
			[]int{1, 2, 3},
			[]string{"1", "2", "3"},
			false,
		},
		{
			"[]int64",
			[]int64{10, 20, 30},
			[]string{"10", "20", "30"},
			false,
		},
		{
			"[]interface{}",
			[]interface{}{"a", 1, true},
			[]string{"1", "a"}, // true becomes "1" which is deduplicated
			false,
		},
		{
			"[]interface{} with duplicates",
			[]interface{}{"a", "b", "a", "c", "b"},
			[]string{"a", "b", "c"},
			false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := c.ValueToStringSlice(tt.input)
			if tt.hasError {
				if err == nil {
					t.Errorf("expected error, got none")
				}
			} else {
				if err != nil {
					t.Errorf("unexpected error: %v", err)
				}
				if len(result) != len(tt.expected) {
					t.Errorf("expected %v, got %v", tt.expected, result)
				}
				// Check if all expected values are present (order may vary after sorting)
				for _, exp := range tt.expected {
					found := false
					for _, r := range result {
						if r == exp {
							found = true
							break
						}
					}
					if !found {
						t.Errorf("expected value %q not found in %v", exp, result)
					}
				}
			}
		})
	}
}

// TestValueToStringSliceMap tests ValueToStringSlice with map input.
func TestValueToStringSliceMap(t *testing.T) {
	c := NewConverter()

	input := map[string]interface{}{
		"a": 1,
		"b": "hello",
		"c": true, // true becomes "1" which is deduplicated with a
	}

	result, err := c.ValueToStringSlice(input)
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}

	// Map values should be converted and deduplicated (1 and true both become "1")
	expectedCount := 2 // 1 (from "a" and "c"), "hello"
	if len(result) != expectedCount {
		t.Errorf("expected %d values, got %d: %v", expectedCount, len(result), result)
	}
}

// TestCompareNumStrings tests CompareNumStrings.
func TestCompareNumStrings(t *testing.T) {
	c := NewConverter()

	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"10", "5", 1},
		{"5", "10", -1},
		{"100", "20", 1},
		{"20", "100", -1},
		{"10", "10", 0},
		{"1.5", "1.2", 1},
		{"1.2", "1.5", -1},
		{"10.5", "10.2", 1},
		{"9.9", "10.0", -1}, // 9.9 < 10.0 (comparing as numbers)
		{"100", "99", 1},
		{"0", "0", 0},
		{"1000", "100", 1},
	}

	for _, tt := range tests {
		result := c.CompareNumStrings(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("CompareNumStrings(%q, %q): expected %d, got %d", tt.a, tt.b, tt.expected, result)
		}
	}
}

// TestCompareNumStringsDifferentLengths tests comparing numbers with different digit counts.
func TestCompareNumStringsDifferentLengths(t *testing.T) {
	c := NewConverter()

	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"100", "99", 1},
		{"99", "100", -1},
		{"1000", "100", 1},
		{"100", "1000", -1},
		{"10000", "9999", 1},
	}

	for _, tt := range tests {
		result := c.CompareNumStrings(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("CompareNumStrings(%q, %q): expected %d, got %d", tt.a, tt.b, tt.expected, result)
		}
	}
}

// TestCompareNumStringsWithDecimals tests comparing decimal numbers.
func TestCompareNumStringsWithDecimals(t *testing.T) {
	c := NewConverter()

	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"10.5", "10.2", 1},
		{"10.2", "10.5", -1},
		{"10.5", "10.5", 0},
		{"1.99", "2.0", -1},
		{"2.0", "1.99", 1},
	}

	for _, tt := range tests {
		result := c.CompareNumStrings(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("CompareNumStrings(%q, %q): expected %d, got %d", tt.a, tt.b, tt.expected, result)
		}
	}
}

// TestValueToStringSliceEmptySlice tests with empty slice.
func TestValueToStringSliceEmptySlice(t *testing.T) {
	c := NewConverter()

	result, err := c.ValueToStringSlice([]string{})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 0 {
		t.Errorf("expected empty slice, got %v", result)
	}
}

// TestValueToStringSliceSingleElement tests with single element slice.
func TestValueToStringSliceSingleElement(t *testing.T) {
	c := NewConverter()

	result, err := c.ValueToStringSlice([]string{"single"})
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if len(result) != 1 || result[0] != "single" {
		t.Errorf("expected [single], got %v", result)
	}
}

// TestCompareNumStringsWithLeadingZeros tests comparing numbers with leading zeros.
func TestCompareNumStringsWithLeadingZeros(t *testing.T) {
	c := NewConverter()

	// Leading zeros should be handled correctly
	tests := []struct {
		a        string
		b        string
		expected int
	}{
		{"010", "005", 1},
		{"005", "010", -1},
	}

	for _, tt := range tests {
		result := c.CompareNumStrings(tt.a, tt.b)
		if result != tt.expected {
			t.Errorf("CompareNumStrings(%q, %q): expected %d, got %d", tt.a, tt.b, tt.expected, result)
		}
	}
}

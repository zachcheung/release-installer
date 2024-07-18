package main

import (
	"testing"
)

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		v1       string
		v2       string
		expected int
	}{
		{"1.0.0", "1.0.0", 0},
		{"5.30.0", "5.9", 1},
		{"5.9", "5.12.2", -1},
	}

	for _, test := range tests {
		result := compareVersions(test.v1, test.v2)
		if result != test.expected {
			t.Errorf("compareVersions(%s, %s) = %d; expected %d", test.v1, test.v2, result, test.expected)
		}
	}
}

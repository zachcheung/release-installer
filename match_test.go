package main

import (
	"testing"
)

func TestHashFileRe(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"release-installer_0.5.0_checksums.txt", true},
		{"checksums", true},
		{"checksums.txt", true},
		{"checksum.json", true},
		{"sha256sums.txt", true},
	}

	for _, test := range tests {
		match := hashFileRe.MatchString(test.filename)
		if match != test.expected {
			t.Errorf("For filename %s, expected %v, but got %v", test.filename, test.expected, match)
		}
	}
}

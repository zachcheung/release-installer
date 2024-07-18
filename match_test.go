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

func TestMuslRe(t *testing.T) {
	tests := []struct {
		filename string
		expected bool
	}{
		{"monit-5.34.0-linux-x64-musl.tar.gz", true},
		{"monit-5.34.0-linux-x64.tar.gz", false},
		{"monit-musli.tar.gz", false},
		{"monit-5.34.0-linux-x64-musl", true},
	}

	for _, test := range tests {
		match := muslRe.MatchString(test.filename)
		if match != test.expected {
			t.Errorf("For filename %s, expected %v, but got %v", test.filename, test.expected, match)
		}
	}
}

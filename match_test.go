package main

import (
	"testing"
)

func TestMatchAsset(t *testing.T) {
	tests := []struct {
		name     string
		goos     string
		goarch   string
		expected bool
	}{
		{"node_exporter-1.8.2.linux-arm64.tar.gz", "linux", "arm64", true},
		{"aerospike-prometheus-exporter_1.17.0_x86_64.tgz", "linux", "amd64", true},
		{"prometheus-iotdb-exporter_1.1_linux_64bit.tar.gz", "linux", "amd64", true},
		{"prometheus-exporter-linux-1.0.1.tgz", "linux", "amd64", true},
		{"couchdb-prometheus-exporter_30.10.1_Linux_x86_64.tar.gz", "linux", "amd64", true},
		{"prosafe_exporter-v0.2.8-x86_64-lnx.zip", "linux", "amd64", true},
		{"beanstalkd_exporter-1.0.5.linux-amd64.sha256", "linux", "amd64", false},
		{"site24x7_exporter-1.1.1-aarch64-unknown-linux-gnu", "linux", "amd64", false},
		{"php-fpm-exporter.linux.amd64.sha256.txt", "linux", "amd64", false},
		{"softether_exporter-v0.2.0-x86_64-lnx.zip", "linux", "amd64", true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			actual := matchAsset(test.name, test.goos, test.goarch)
			if actual != test.expected {
				t.Errorf("Expected MatchAsset('%s', '%s', '%s') to be %v, got %v", test.name, test.goos, test.goarch, test.expected, actual)
			}
		})
	}
}

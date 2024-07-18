package main

import (
	"errors"
	"testing"
)

func TestWeight(t *testing.T) {
	tests := []struct {
		name   string
		asset  *Asset
		weight int
	}{
		{
			name:   "node_exporter-1.8.2.linux-arm64.tar.gz",
			asset:  newAsset("node_exporter-1.8.2.linux-arm64.tar.gz", "", "linux", "arm64"),
			weight: 5,
		},
		{
			name:   "aerospike-prometheus-exporter_1.17.0_x86_64.tgz",
			asset:  newAsset("aerospike-prometheus-exporter_1.17.0_x86_64.tgz", "", "linux", "amd64"),
			weight: 3,
		},
		{
			name:   "prometheus-iotdb-exporter_1.1_linux_64bit.tar.gz",
			asset:  newAsset("prometheus-iotdb-exporter_1.1_linux_64bit.tar.gz", "", "linux", "amd64"),
			weight: 5,
		},
		{
			name:   "prometheus-exporter-linux-1.0.1.tgz",
			asset:  newAsset("prometheus-exporter-linux-1.0.1.tgz", "", "linux", "amd64"),
			weight: 3,
		},
		{
			name:   "couchdb-prometheus-exporter_30.10.1_Linux_x86_64.tar.gz",
			asset:  newAsset("couchdb-prometheus-exporter_30.10.1_Linux_x86_64.tar.gz", "", "linux", "amd64"),
			weight: 5,
		},
		{
			name:   "prosafe_exporter-v0.2.8-x86_64-lnx.zip",
			asset:  newAsset("prosafe_exporter-v0.2.8-x86_64-lnx.zip", "", "linux", "amd64"),
			weight: 5,
		},
		{
			name:   "beanstalkd_exporter-1.0.5.linux-amd64.sha256",
			asset:  newAsset("beanstalkd_exporter-1.0.5.linux-amd64.sha256", "", "linux", "amd64"),
			weight: 4,
		},
		{
			name:   "site24x7_exporter-1.1.1-aarch64-unknown-linux-gnu",
			asset:  newAsset("site24x7_exporter-1.1.1-aarch64-unknown-linux-gnu", "", "linux", "amd64"),
			weight: 3,
		},
		{
			name:   "php-fpm-exporter.linux.amd64.sha256.txt",
			asset:  newAsset("php-fpm-exporter.linux.amd64.sha256.txt", "", "linux", "amd64"),
			weight: 4,
		},
		{
			name:   "softether_exporter-v0.2.0-x86_64-lnx.zip",
			asset:  newAsset("softether_exporter-v0.2.0-x86_64-lnx.zip", "", "linux", "amd64"),
			weight: 5,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.asset.Weight(); got != tt.weight {
				t.Errorf("%s, Asset.Weight() = %d, want %d", tt.name, got, tt.weight)
			}
		})
	}
}

func TestFindMaxWeightAsset(t *testing.T) {
	tests := []struct {
		name    string
		assets  Assets
		want    Asset
		wantErr error
	}{
		{
			name: "multiple max weight assets",
			assets: Assets{
				*newAsset("monit-5.34.0-linux-x64.tar.gz", "", "linux", "amd64"),
				*newAsset("monit-5.34.0-linux-x64-musl.tar.gz", "", "linux", "amd64"),
			},
			want:    Asset{},
			wantErr: ErrMultipleMaxWeightAsset,
		},
		{
			name: "single max weight asset",
			assets: Assets{
				*newAsset("node_exporter-1.8.2.linux-amd64.tar.gz", "", "linux", "amd64"),
				*newAsset("node_exporter-1.8.2.linux-arm64.tar.gz", "", "linux", "amd64"),
				*newAsset("node_exporter-1.8.2.darwin-amd64.tar.gz", "", "linux", "amd64"),
			},
			want:    *newAsset("node_exporter-1.8.2.linux-amd64.tar.gz", "", "linux", "amd64"),
			wantErr: nil,
		},
		{
			name:    "no assets",
			assets:  Assets{},
			want:    Asset{},
			wantErr: ErrNoAsset,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.assets.FindMaxWeightAsset()
			if !errors.Is(err, tt.wantErr) {
				t.Errorf("FindMaxWeightAsset() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("FindMaxWeightAsset() = %v, want %v", got, tt.want)
			}
		})
	}
}

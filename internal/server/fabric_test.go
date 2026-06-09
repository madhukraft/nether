package server

import (
	"testing"
)

func TestGetLatestStable(t *testing.T) {
	tests := []struct {
		name     string
		versions []fabricVersion
		want     string
		wantErr  bool
	}{
		{
			name: "prefers stable over unstable",
			versions: []fabricVersion{
				{Version: "0.15.0", Stable: false},
				{Version: "0.14.0", Stable: true},
				{Version: "0.13.0", Stable: false},
			},
			want: "0.14.0",
		},
		{
			name: "returns first stable if multiple",
			versions: []fabricVersion{
				{Version: "0.16.0", Stable: true},
				{Version: "0.15.0", Stable: true},
			},
			want: "0.16.0",
		},
		{
			name: "falls back to first version if none stable",
			versions: []fabricVersion{
				{Version: "0.16.0", Stable: false},
				{Version: "0.15.0", Stable: false},
			},
			want: "0.16.0",
		},
		{
			name:     "returns error on empty list",
			versions: []fabricVersion{},
			wantErr:  true,
		},
		{
			name: "single stable version",
			versions: []fabricVersion{
				{Version: "0.19.3", Stable: true},
			},
			want: "0.19.3",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := getLatestStable(tt.versions)
			if (err != nil) != tt.wantErr {
				t.Fatalf("getLatestStable() error = %v; wantErr %v", err, tt.wantErr)
			}
			if got != tt.want {
				t.Errorf("getLatestStable() = %q; want %q", got, tt.want)
			}
		})
	}
}

package server

import (
	"reflect"
	"testing"
)

func TestParseForgeVersions(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		want     int // expected number of parsed versions
	}{
		{
			name:     "standard versions",
			versions: []string{"1.20.1-45", "1.20.1-46", "1.20-42"},
			want:     3,
		},
		{
			name:     "invalid versions ignored",
			versions: []string{"invalid", "1.20.1-45", "", "1.20-abc"},
			want:     1,
		},
		{
			name:     "empty input",
			versions: []string{},
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseForgeVersions(tt.versions)
			if len(got) != tt.want {
				t.Errorf("parseForgeVersions() returned %d; want %d", len(got), tt.want)
			}
		})
	}
}

func TestForgeExactVersionMatch(t *testing.T) {
	// Simulating the filter logic in getForgeVersionForMC with exact match
	versions := parseForgeVersions([]string{"1.20.1-45", "1.20.1-46", "1.20-42", "1.20.0-1", "1.21-10"})

	var candidates20 []string
	for _, v := range versions {
		if v.mcVersion == "1.20" {
			candidates20 = append(candidates20, v.mcVersion+"-"+v.build)
		}
	}
	if len(candidates20) != 1 {
		t.Errorf("expected exactly 1 candidate for 1.20, got %d: %v", len(candidates20), candidates20)
	} else if candidates20[0] != "1.20-42" {
		t.Errorf("expected 1.20-42, got %s", candidates20[0])
	}

	var candidates201 []string
	for _, v := range versions {
		if v.mcVersion == "1.20.1" {
			candidates201 = append(candidates201, v.mcVersion+"-"+v.build)
		}
	}
	if len(candidates201) != 2 {
		t.Errorf("expected exactly 2 candidates for 1.20.1, got %d: %v", len(candidates201), candidates201)
	}
}

func TestParseNeoForgeVersions(t *testing.T) {
	tests := []struct {
		name     string
		versions []string
		want     int
	}{
		{
			name:     "standard versions",
			versions: []string{"20.1.87", "20.2.50", "21.0.100"},
			want:     3,
		},
		{
			name:     "skips beta versions",
			versions: []string{"20.1.87", "20.2.50-beta"},
			want:     1,
		},
		{
			name:     "skips craftmine versions",
			versions: []string{"20.1.87", "20.2.50-craftmine"},
			want:     1,
		},
		{
			name:     "requires at least 2 parts",
			versions: []string{"20"},
			want:     0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseNeoForgeVersions(tt.versions)
			if len(got) != tt.want {
				t.Errorf("parseNeoForgeVersions() returned %d; want %d", len(got), tt.want)
			}
		})
	}
}

func TestNeoForgeExactVersionMatch(t *testing.T) {
	versions := parseNeoForgeVersions([]string{"20.0.42", "20.1.87", "20.2.50", "21.0.100"})

	// Match MC 1.20 (patch=0): should find only 20.0.x
	var candidates20 []string
	for _, v := range versions {
		if len(v.parts) >= 2 && v.parts[0] == 20 && v.parts[1] == 0 {
			candidates20 = append(candidates20, v.full)
		}
	}
	if len(candidates20) != 1 {
		t.Errorf("expected exactly 1 candidate for MC 1.20, got %d: %v", len(candidates20), candidates20)
	} else if candidates20[0] != "20.0.42" {
		t.Errorf("expected 20.0.42, got %s", candidates20[0])
	}

	// Match MC 1.20.1 (patch=1): should find only 20.1.x
	var candidates201 []string
	for _, v := range versions {
		if len(v.parts) >= 2 && v.parts[0] == 20 && v.parts[1] == 1 {
			candidates201 = append(candidates201, v.full)
		}
	}
	if len(candidates201) != 1 {
		t.Errorf("expected exactly 1 candidate for MC 1.20.1, got %d: %v", len(candidates201), candidates201)
	} else if candidates201[0] != "20.1.87" {
		t.Errorf("expected 20.1.87, got %s", candidates201[0])
	}
}

func TestParseVersionParts(t *testing.T) {
	tests := []struct {
		input   string
		want    []int
		wantErr bool
	}{
		{"1.20.1", []int{1, 20, 1}, false},
		{"1.21", []int{1, 21}, false},
		{"42", []int{42}, false},
		{"1.2.3.4", []int{1, 2, 3, 4}, false},
		{"abc", nil, true},
		{"1.abc", nil, true},
		{"", nil, true},
		{".1", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got, ok := parseVersionParts(tt.input)
			if tt.wantErr {
				if ok {
					t.Errorf("parseVersionParts(%q) = %v, %v; want error", tt.input, got, ok)
				}
				return
			}
			if !ok {
				t.Fatalf("parseVersionParts(%q) = %v, %v; want success", tt.input, got, ok)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("parseVersionParts(%q) = %v; want %v", tt.input, got, tt.want)
			}
		})
	}
}

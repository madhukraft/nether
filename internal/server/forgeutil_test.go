package server

import (
	"reflect"
	"testing"
)

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

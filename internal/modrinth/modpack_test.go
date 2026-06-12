package modrinth

import (
	"testing"
)

func TestSafeJoin(t *testing.T) {
	tests := []struct {
		name    string
		base    string
		path    string
		want    string
		wantErr bool
	}{
		{
			name:    "normal path",
			base:    ".",
			path:    "mods/lithium.jar",
			want:    "mods/lithium.jar",
			wantErr: false,
		},
		{
			name:    "path with subdirectory",
			base:    ".",
			path:    "config/options.txt",
			want:    "config/options.txt",
			wantErr: false,
		},
		{
			name:    "path traversal up",
			base:    ".",
			path:    "../etc/passwd",
			wantErr: true,
		},
		{
			name:    "path traversal deep",
			base:    ".",
			path:    "mods/../../etc/passwd",
			wantErr: true,
		},
		{
			name:    "absolute path",
			base:    ".",
			path:    "/etc/passwd",
			wantErr: true,
		},
		{
			name:    "empty path",
			base:    ".",
			path:    "",
			want:    ".",
			wantErr: false,
		},
		{
			name:    "non-empty base with subdir",
			base:    "mods",
			path:    "lithium.jar",
			want:    "mods/lithium.jar",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := safeJoin(tt.base, tt.path)
			if tt.wantErr {
				if err == nil {
					t.Errorf("safeJoin(%q, %q) = %q, nil; want error", tt.base, tt.path, got)
				}
				return
			}
			if err != nil {
				t.Errorf("safeJoin(%q, %q) returned error: %v", tt.base, tt.path, err)
				return
			}
			if got != tt.want {
				t.Errorf("safeJoin(%q, %q) = %q; want %q", tt.base, tt.path, got, tt.want)
			}
		})
	}
}

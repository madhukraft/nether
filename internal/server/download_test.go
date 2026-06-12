package server

import (
	"errors"
	"net/http"
	"testing"
)

func TestHTTPError(t *testing.T) {
	err := &httpError{StatusCode: http.StatusNotFound, URL: "https://example.com/file"}
	if err.StatusCode != http.StatusNotFound {
		t.Errorf("expected 404, got %d", err.StatusCode)
	}
	msg := err.Error()
	if msg != "HTTP 404 from https://example.com/file" {
		t.Errorf("unexpected error message: %q", msg)
	}
}

func TestIsNotFound(t *testing.T) {
	tests := []struct {
		name string
		err  error
		want bool
	}{
		{"httpError 404", &httpError{StatusCode: http.StatusNotFound}, true},
		{"httpError 500", &httpError{StatusCode: http.StatusInternalServerError}, false},
		{"httpError 200", &httpError{StatusCode: http.StatusOK}, false},
		{"nil", nil, false},
		{"other error", errors.New("something went wrong"), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := isNotFound(tt.err)
			if got != tt.want {
				t.Errorf("isNotFound(%v) = %v; want %v", tt.err, got, tt.want)
			}
		})
	}
}

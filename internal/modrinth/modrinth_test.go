package modrinth

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestGetProject(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/project/lithium" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(Project{
			ID:          "gvQqBUYK",
			Slug:        "lithium",
			Title:       "Lithium",
			ProjectType: "mod",
			ServerSide:  "required",
			Loaders:     []string{"fabric", "quilt"},
		})
	}))
	defer srv.Close()

	origBase := baseURL
	baseURL = srv.URL
	defer func() { baseURL = origBase }()

	c := NewClient()
	p, err := c.GetProject("lithium")
	if err != nil {
		t.Fatal(err)
	}
	if p.Slug != "lithium" {
		t.Errorf("expected lithium, got %s", p.Slug)
	}
	if p.ProjectType != "mod" {
		t.Errorf("expected mod, got %s", p.ProjectType)
	}
}

func TestGetVersions(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/project/gvQqBUYK/version" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		loaders := r.URL.Query().Get("loaders")
		if loaders != `["fabric"]` {
			t.Errorf("unexpected loaders: %s", loaders)
		}
		json.NewEncoder(w).Encode([]Version{
			{
				ID:            "ver1",
				ProjectID:     "gvQqBUYK",
				VersionNumber: "0.12.1",
				GameVersions:  []string{"1.21.1"},
				Loaders:       []string{"fabric"},
				Files: []VersionFile{
					{URL: "https://example.com/lithium.jar", Primary: true, Filename: "lithium-0.12.1.jar"},
				},
			},
		})
	}))
	defer srv.Close()

	origBase := baseURL
	baseURL = srv.URL
	defer func() { baseURL = origBase }()

	c := NewClient()
	versions, err := c.GetVersions("gvQqBUYK", []string{"fabric"}, []string{"1.21.1"})
	if err != nil {
		t.Fatal(err)
	}
	if len(versions) != 1 {
		t.Fatalf("expected 1 version, got %d", len(versions))
	}
	if versions[0].VersionNumber != "0.12.1" {
		t.Errorf("expected 0.12.1, got %s", versions[0].VersionNumber)
	}
}

func TestSearch(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/search" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}
		json.NewEncoder(w).Encode(SearchResults{
			Hits: []SearchHit{
				{Slug: "lithium", Title: "Lithium", ProjectType: "mod", Downloads: 500000},
			},
			TotalHits: 1,
		})
	}))
	defer srv.Close()

	origBase := baseURL
	baseURL = srv.URL
	defer func() { baseURL = origBase }()

	c := NewClient()
	results, err := c.Search("lithium", 5, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results.Hits) != 1 {
		t.Fatalf("expected 1 hit, got %d", len(results.Hits))
	}
	if results.Hits[0].Slug != "lithium" {
		t.Errorf("expected lithium, got %s", results.Hits[0].Slug)
	}
}

func TestParseSlug(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"lithium", "lithium"},
		{"https://modrinth.com/mod/lithium", "lithium"},
		{"https://modrinth.com/mod/lithium?version=1.21.1", "lithium"},
		{"https://modrinth.com/modpack/my-pack", "my-pack"},
		{"https://modrinth.com/project/P7dR8mMs", "P7dR8mMs"},
		{"P7dR8mMs", "P7dR8mMs"},
	}
	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := parseSlug(tt.input)
			if got != tt.want {
				t.Errorf("parseSlug(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestSearchReal(t *testing.T) {
	// Skip in short mode (no network calls)
	if testing.Short() {
		t.Skip("skipping network test")
	}

	c := NewClient()
	results, err := c.Search("lithium", 3, nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(results.Hits) == 0 {
		t.Fatal("expected at least 1 result")
	}
	t.Logf("Search results for 'lithium':")
	for _, h := range results.Hits {
		t.Logf("  - %s (%s): %s", h.Title, h.Slug, h.Description[:min(len(h.Description), 80)])
	}
}

func TestGetProjectReal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test")
	}

	c := NewClient()
	p, err := c.GetProject("lithium")
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Project: %s (%s) - type: %s, server: %s", p.Title, p.Slug, p.ProjectType, p.ServerSide)
}

func TestGetVersionsReal(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping network test")
	}

	c := NewClient()
	p, err := c.GetProject("lithium")
	if err != nil {
		t.Fatal(err)
	}
	versions, err := c.GetVersions(p.ID, []string{"fabric"}, []string{"1.21.1"})
	if err != nil {
		t.Fatal(err)
	}
	t.Logf("Found %d versions for lithium (fabric, 1.21.1):", len(versions))
	for _, v := range versions {
		t.Logf("  - %s: %s", v.VersionNumber, v.Name)
	}
}



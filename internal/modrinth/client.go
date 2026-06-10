package modrinth

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

var baseURL = "https://api.modrinth.com/v2"

type Client struct {
	http *http.Client
}

func NewClient() *Client {
	return &Client{
		http: &http.Client{Timeout: 30 * time.Second},
	}
}

func (c *Client) get(path string, query url.Values, dest interface{}) error {
	u := baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}

	resp, err := c.http.Get(u)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusTooManyRequests {
		return fmt.Errorf("rate limited by Modrinth API")
	}
	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("not found")
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("unexpected status %d from Modrinth API", resp.StatusCode)
	}

	if err := json.NewDecoder(resp.Body).Decode(dest); err != nil {
		return fmt.Errorf("failed to decode response: %w", err)
	}
	return nil
}

func (c *Client) GetProject(slug string) (*Project, error) {
	var p Project
	if err := c.get("/project/" + url.PathEscape(slug), nil, &p); err != nil {
		return nil, fmt.Errorf("fetching project: %w", err)
	}
	return &p, nil
}

func (c *Client) GetVersions(projectID string, loaders, gameVersions []string) ([]Version, error) {
	q := url.Values{}
	if len(loaders) > 0 {
		q.Set("loaders", `["`+strings.Join(loaders, `","`)+`"]`)
	}
	if len(gameVersions) > 0 {
		q.Set("game_versions", `["`+strings.Join(gameVersions, `","`)+`"]`)
	}

	var versions []Version
	if err := c.get("/project/"+url.PathEscape(projectID)+"/version", q, &versions); err != nil {
		return nil, fmt.Errorf("fetching versions: %w", err)
	}
	return versions, nil
}

func (c *Client) GetVersion(versionID string) (*Version, error) {
	var v Version
	if err := c.get("/version/"+url.PathEscape(versionID), nil, &v); err != nil {
		return nil, fmt.Errorf("fetching version: %w", err)
	}
	return &v, nil
}

func (c *Client) Search(query string, limit int, facets map[string][]string) (*SearchResults, error) {
	q := url.Values{}
	q.Set("query", query)
	if limit > 0 {
		q.Set("limit", fmt.Sprintf("%d", limit))
	}
	if len(facets) > 0 {
		var parts []string
		for key, vals := range facets {
			for _, v := range vals {
				parts = append(parts, fmt.Sprintf(`["%s:%s"]`, key, v))
			}
		}
		q.Set("facets", "["+strings.Join(parts, ",")+"]")
	}

	var results SearchResults
	if err := c.get("/search", q, &results); err != nil {
		return nil, fmt.Errorf("searching: %w", err)
	}
	return &results, nil
}

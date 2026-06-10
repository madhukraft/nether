package modrinth

type Project struct {
	ID            string   `json:"id"`
	Slug          string   `json:"slug"`
	Title         string   `json:"title"`
	Description   string   `json:"description"`
	ProjectType   string   `json:"project_type"`
	ClientSide    string   `json:"client_side"`
	ServerSide    string   `json:"server_side"`
	Categories    []string `json:"categories"`
	Versions      []string `json:"versions"`
	GameVersions  []string `json:"game_versions"`
	Loaders       []string `json:"loaders"`
	IconURL       string   `json:"icon_url"`
}

type Version struct {
	ID            string        `json:"id"`
	ProjectID     string        `json:"project_id"`
	Name          string        `json:"name"`
	VersionNumber string        `json:"version_number"`
	GameVersions  []string      `json:"game_versions"`
	Loaders       []string      `json:"loaders"`
	Files         []VersionFile `json:"files"`
	Dependencies  []Dependency  `json:"dependencies"`
	VersionType   string        `json:"version_type"`
}

type VersionFile struct {
	Hashes   map[string]string `json:"hashes"`
	URL      string            `json:"url"`
	Filename string            `json:"filename"`
	Primary  bool              `json:"primary"`
	Size     int64             `json:"size"`
}

type Dependency struct {
	VersionID      string `json:"version_id"`
	ProjectID      string `json:"project_id"`
	FileName       string `json:"file_name"`
	DependencyType string `json:"dependency_type"`
}

type SearchHit struct {
	Slug         string   `json:"slug"`
	Title        string   `json:"title"`
	Description  string   `json:"description"`
	ProjectType  string   `json:"project_type"`
	Categories   []string `json:"categories"`
	Versions     []string `json:"versions"`
	Author       string   `json:"author"`
	Downloads    int64    `json:"downloads"`
	Follows      int64    `json:"follows"`
	IconURL      string   `json:"icon_url"`
	ProjectID    string   `json:"project_id"`
	ClientSide   string   `json:"client_side"`
	ServerSide   string   `json:"server_side"`
	GameVersions []string `json:"game_versions"`
	Loaders      []string `json:"loaders"`
}

type SearchResults struct {
	Hits     []SearchHit `json:"hits"`
	Offset   int         `json:"offset"`
	Limit    int         `json:"limit"`
	TotalHits int64      `json:"total_hits"`
}

package modrinth

import "strings"

func parseSlug(input string) string {
	baseURLs := []string{"https://modrinth.com/", "http://modrinth.com/", "modrinth.com/"}
	for _, prefix := range baseURLs {
		if strings.HasPrefix(input, prefix) {
			path := strings.TrimPrefix(input, prefix)
			path = strings.TrimSuffix(path, "/")
			parts := strings.Split(path, "/")
			if len(parts) >= 1 {
				slug := parts[len(parts)-1]
				if idx := strings.Index(slug, "?"); idx >= 0 {
					slug = slug[:idx]
				}
				return slug
			}
		}
	}
	return input
}

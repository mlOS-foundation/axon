package types

// SearchResult represents a model search result from the registry
type SearchResult struct {
	Name        string   `json:"name"`
	Namespace   string   `json:"namespace"`
	Version     string   `json:"version"`
	Description string   `json:"description"`
	Framework   string   `json:"framework"`
	Tags        []string `json:"tags"`
}

// RegistryIndex represents the registry index
type RegistryIndex struct {
	Version   string              `json:"version"`
	Generated string              `json:"generated"`
	Models    []IndexModelEntry   `json:"models"`
	Namespaces map[string]NamespaceInfo `json:"namespaces"`
	Statistics Statistics          `json:"statistics"`
}

// IndexModelEntry represents a model entry in the index
type IndexModelEntry struct {
	Name          string   `json:"name"`
	Namespace     string   `json:"namespace"`
	LatestVersion string   `json:"latest_version"`
	Description   string   `json:"description"`
	Framework     string   `json:"framework"`
	Tags          []string `json:"tags"`
	Downloads     int      `json:"downloads"`
	Stars         int      `json:"stars"`
	Updated       string   `json:"updated"`
}

// NamespaceInfo provides information about a namespace
type NamespaceInfo struct {
	Description string `json:"description"`
	ModelCount  int    `json:"model_count"`
}

// Statistics provides registry statistics
type Statistics struct {
	TotalModels    int `json:"total_models"`
	TotalDownloads int `json:"total_downloads"`
	TotalNamespaces int `json:"total_namespaces"`
}


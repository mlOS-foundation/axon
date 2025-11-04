package types

import "time"

// Model represents an installed model instance
type Model struct {
	Namespace   string
	Name        string
	Version     string
	InstalledAt string // Path where model is installed
	Manifest    *Manifest
}

// InstalledModelInfo provides information about an installed model
type InstalledModelInfo struct {
	FullName    string
	Version     string
	Path        string
	Size        int64
	InstalledAt time.Time
}


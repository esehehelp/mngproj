package config

// ProjectConfig represents the root mngproj.toml
type ProjectConfig struct {
	Project    ProjectMeta       `toml:"project"`
	Components []ComponentConfig `toml:"components"`
	Resolution ResolutionConfig  `toml:"resolution"`
}

type ResolutionConfig struct {
	// Map of role name to priority score. Higher wins.
	RolePriority map[string]int `toml:"role_priority"`
}

// ProjectMeta contains metadata about the project
type ProjectMeta struct {
	Name        string   `toml:"name"`
	Description string   `toml:"description"`
	Tags        []string `toml:"tags"`
}

// ComponentConfig represents a component definition in mngproj.toml
type ComponentConfig struct {
	Name         string            `toml:"name"`
	Type         string            `toml:"type"`
	Types        []string          `toml:"types"`
	Path         string            `toml:"path"`
	Priority     int               `toml:"priority"`
	Dependencies []string          `toml:"dependencies"`
	Env          map[string]string `toml:"env"`
	Scripts      map[string]string `toml:"scripts"`
}

// PresetConfig represents a preset definition (e.g. presets/go.toml)
type PresetConfig struct {
	Metadata PresetMeta        `toml:"metadata"`
	Scripts  map[string]string `toml:"scripts"`
	Env      map[string]string `toml:"env"`
}

type PresetMeta struct {
	Type         string `toml:"type"`
	Role         string `toml:"role"` // language, framework, package_manager, tool
	Description  string `toml:"description"`
	ManifestFile string `toml:"manifest_file"` // e.g. "requirements.txt", "package.json"
}

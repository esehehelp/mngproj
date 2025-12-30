package config

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"

	"github.com/pelletier/go-toml/v2"
)

// LoadProjectConfig reads and parses mngproj.toml from the given path
func LoadProjectConfig(path string) (*ProjectConfig, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("failed to read config file: %w", err)
	}

	var cfg ProjectConfig
	if err := toml.Unmarshal(data, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse config file: %w", err)
	}

	// Validate component names for duplicates
	seen := make(map[string]bool)
	for i := range cfg.Components {
		name := cfg.Components[i].Name
		if seen[name] {
			return nil, fmt.Errorf("duplicate component name found: %q", name)
		}
		seen[name] = true
	}

	// Set default path if empty
	for i := range cfg.Components {
		if cfg.Components[i].Path == "" {
			cfg.Components[i].Path = "."
		}
	}

	return &cfg, nil
}

// LoadPreset loads a preset configuration by type name
// It prioritizes {type}_{GOOS}.toml, then falls back to {type}.toml
func LoadPreset(presetsDir, typeName string) (*PresetConfig, error) {
	// 1. Try OS-specific preset
	osSpecificName := fmt.Sprintf("%s_%s.toml", typeName, runtime.GOOS)
	if preset, err := findAndLoadPreset(presetsDir, osSpecificName); err == nil {
		return preset, nil
	}

	// 2. Fallback to standard preset
	return findAndLoadPreset(presetsDir, fmt.Sprintf("%s.toml", typeName))
}

func findAndLoadPreset(presetsDir, filename string) (*PresetConfig, error) {
	var foundPath string
	err := filepath.WalkDir(presetsDir, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if !d.IsDir() && d.Name() == filename {
			foundPath = path
			return os.ErrExist // Signal to stop walking
		}
		return nil
	})

	if err != nil && err != os.ErrExist {
		return nil, fmt.Errorf("error searching for preset: %w", err)
	}

	if foundPath == "" {
		return nil, fmt.Errorf("preset file %q not found in %s", filename, presetsDir)
	}

	data, err := os.ReadFile(foundPath)
	if err != nil {
		return nil, fmt.Errorf("failed to read preset file %s: %w", foundPath, err)
	}

	var preset PresetConfig
	if err := toml.Unmarshal(data, &preset); err != nil {
		return nil, fmt.Errorf("failed to parse preset file: %w", err)
	}

	return &preset, nil
}

// SaveProjectConfig writes the project configuration to the specified path
func SaveProjectConfig(path string, cfg *ProjectConfig) error {
	data, err := toml.Marshal(cfg)
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	if err := os.WriteFile(path, data, 0644); err != nil {
		return fmt.Errorf("failed to write config file: %w", err)
	}

	return nil
}

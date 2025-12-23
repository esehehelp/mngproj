package config

import (
	"fmt"
	"os"
	"path/filepath"

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

	// Set default path if empty
	for i := range cfg.Components {
		if cfg.Components[i].Path == "" {
			cfg.Components[i].Path = "."
		}
	}

	return &cfg, nil
}

// LoadPreset loads a preset configuration by type name
// It looks for {type}.toml in the given presetsDir and its subdirectories
func LoadPreset(presetsDir, typeName string) (*PresetConfig, error) {
	filename := fmt.Sprintf("%s.toml", typeName)
	
	// Simple search: check root, then common subdirs
	// Ideally using WalkDir, but for performance with known structure we can check direct paths
	// or assume the user/tool knows the structure.
	// Let's do a Walk to be robust.
	
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
		return nil, fmt.Errorf("preset %q not found in %s", typeName, presetsDir)
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

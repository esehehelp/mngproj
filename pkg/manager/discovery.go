package manager

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
)

// FindConfigFile looks for mngproj.toml starting from startDir and walking up
func FindConfigFile(startDir string) (string, error) {
	dir := startDir
	for {
		path := filepath.Join(dir, "mngproj.toml")
		if _, err := os.Stat(path); err == nil {
			return path, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			return "", fmt.Errorf("mngproj.toml not found")
		}
		dir = parent
	}
}

// FindAllProjectConfigs recursively finds all directories containing mngproj.toml
// starting from the given rootDir. It returns a slice of absolute paths to these directories.
func FindAllProjectConfigs(rootDir string) ([]string, error) {
	var projectRoots []string

	err := filepath.WalkDir(rootDir, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			// If an error occurs (e.g., permission denied), log it but continue
			// Or return the error if it's critical. For now, just continue.
			fmt.Fprintf(os.Stderr, "Error accessing %q: %v\n", path, err)
			return nil
		}

		// Check if it's a directory and contains mngproj.toml
		if d.IsDir() {
			configPath := filepath.Join(path, "mngproj.toml")
			if _, err := os.Stat(configPath); err == nil {
				projectRoots = append(projectRoots, path)
				// Do NOT SkipDir, continue searching for other projects within this project's subtree
			}
		}
		return nil
	})

	if err != nil {
		return nil, fmt.Errorf("error walking directory %q: %w", rootDir, err)
	}

	return projectRoots, nil
}

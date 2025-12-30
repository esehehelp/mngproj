package test

import (
	"mngproj/pkg/config"
	"os"
	"path/filepath"
	"testing"
)

func TestDuplicateComponentName(t *testing.T) {
	tmpDir := t.TempDir()
	configContent := `
[project]
name = "DupTest"

[[components]]
name = "app"
type = "go"

[[components]]
name = "app" # Duplicate
type = "go"
`
	configPath := filepath.Join(tmpDir, "mngproj.toml")
	os.WriteFile(configPath, []byte(configContent), 0644)

	_, err := config.LoadProjectConfig(configPath)
	if err == nil {
		t.Error("Expected error for duplicate component name, got nil")
	}
}

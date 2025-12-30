package test

import (
	"mngproj/pkg/manager"
	"os"
	"path/filepath"
	"testing"
)

func TestManagerWithExplicitRoot(t *testing.T) {
	// Setup:
	// /tmp/proj/mngproj.toml -> root = "../src"
	// /tmp/src/backend/ -> component path
	
	tmpDir := t.TempDir()
	projDir := filepath.Join(tmpDir, "proj")
	srcDir := filepath.Join(tmpDir, "src")
	os.MkdirAll(projDir, 0755)
	os.MkdirAll(srcDir, 0755)
	
	configContent := `
[project]
name = "RootTest"
root = "../src"

[[components]]
name = "app"
path = "backend" 
type = "go"
`
	if err := os.WriteFile(filepath.Join(projDir, "mngproj.toml"), []byte(configContent), 0644); err != nil {
		t.Fatal(err)
	}
	
	// Create dummy preset
	presetsDir := filepath.Join(tmpDir, "presets")
	os.MkdirAll(presetsDir, 0755)
	if err := os.WriteFile(filepath.Join(presetsDir, "go.toml"), []byte(`[metadata]
type="go"
role="language"
`), 0644); err != nil {
		t.Fatal(err)
	}

	// Load
	// We need to set MNGPROJ_PRESETS_DIR env var for this test so New() finds our dummy preset
	t.Setenv("MNGPROJ_PRESETS_DIR", presetsDir)

	mgr, err := manager.New(projDir) // Start in projDir
	if err != nil {
		t.Fatalf("Manager New failed: %v", err)
	}

	comp, err := mgr.ResolveComponent("app")
	if err != nil {
		t.Fatalf("ResolveComponent failed: %v", err)
	}

	// Expected AbsPath: /tmp/src/backend
	// Note: explicit root "../src" relative to "projDir" (/tmp/proj) resolves to /tmp/src
	expected := filepath.Join(srcDir, "backend")
	
	// ResolveComponent uses filepath.Join(m.ProjectDir, compConfig.Path)
	// m.ProjectDir should be /tmp/src
	
	if comp.AbsPath != expected {
		t.Errorf("Expected AbsPath %s, got %s", expected, comp.AbsPath)
	}
}

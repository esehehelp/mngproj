package test

import (
	"mngproj/pkg/config"
	"mngproj/pkg/manager"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestAddDependencyAndManifestGeneration(t *testing.T) {
	// 1. Setup Environment
	tmpDir := t.TempDir()
	presetsDir := filepath.Join(tmpDir, "presets")
	projectDir := filepath.Join(tmpDir, "project")
	if err := os.MkdirAll(presetsDir, 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.MkdirAll(projectDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 2. Create Dummy Preset (Python-like)
	pipPreset := `
[metadata]
type = "pip"
role = "package_manager"
manifest_file = "requirements.txt"
[scripts]
install = "echo installing..."
`
	if err := os.WriteFile(filepath.Join(presetsDir, "pip.toml"), []byte(pipPreset), 0644); err != nil {
		t.Fatal(err)
	}

	// 3. Create mngproj.toml
	initialConfig := `
[project]
name = "TestDep"
[[components]]
name = "api"
types = ["pip"]
path = "api"
`
	configPath := filepath.Join(projectDir, "mngproj.toml")
	if err := os.WriteFile(configPath, []byte(initialConfig), 0644); err != nil {
		t.Fatal(err)
	}

	// Create component dir
	apiDir := filepath.Join(projectDir, "api")
	if err := os.MkdirAll(apiDir, 0755); err != nil {
		t.Fatal(err)
	}

	// 4. Initialize Manager
	// We need to load the config first to pass it to Manager struct manually
	// because New() looks for files relative to CWD or absolute paths.
	cfg, err := config.LoadProjectConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to load config: %v", err)
	}

	mgr := &manager.Manager{
		ProjectConfig: cfg,
		ProjectDir:    projectDir,
		PresetsDir:    presetsDir,
	}

	// 5. Test AddDependency
	targetPkg := "requests==2.26.0"
	if err := mgr.AddDependency("api", targetPkg); err != nil {
		t.Fatalf("AddDependency failed: %v", err)
	}

	// 6. Verify mngproj.toml Update
	// Reload config to verify persistence
	updatedCfg, err := config.LoadProjectConfig(configPath)
	if err != nil {
		t.Fatalf("Failed to reload config: %v", err)
	}
	
	found := false
	for _, c := range updatedCfg.Components {
		if c.Name == "api" {
			for _, d := range c.Dependencies {
				if d == targetPkg {
					found = true
					break
				}
			}
		}
	}
	if !found {
		t.Errorf("Dependency %s not found in mngproj.toml after AddDependency", targetPkg)
	}

	// 7. Verify Manifest Generation
	manifestPath := filepath.Join(apiDir, "requirements.txt")
	contentBytes, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("Failed to read manifest file: %v", err)
	}
	content := string(contentBytes)
	if !strings.Contains(content, targetPkg) {
		t.Errorf("Manifest file does not contain %s. Got: %s", targetPkg, content)
	}
}

func TestSyncComponent(t *testing.T) {
	// Setup
	tmpDir := t.TempDir()
	presetsDir := filepath.Join(tmpDir, "presets")
	projectDir := filepath.Join(tmpDir, "project")
	os.MkdirAll(presetsDir, 0755)
	os.MkdirAll(projectDir, 0755)

	// Preset with a verifiable install script
	// We use a file creation to verify execution
	doneFile := "install_done.txt"
	// Note: We use 'sh -c' in execution.go, so standard shell commands work.
	// We need to be careful about CWD. The command runs in component's AbsPath.
	dummyPreset := `
[metadata]
type = "dummy"
role = "tool"
[scripts]
install = "touch install_done.txt"
`
	os.WriteFile(filepath.Join(presetsDir, "dummy.toml"), []byte(dummyPreset), 0644)

	// Config
	cfg := &config.ProjectConfig{
		Components: []config.ComponentConfig{
			{Name: "app", Types: []string{"dummy"}, Path: "."},
		},
	}
	
	mgr := &manager.Manager{
		ProjectConfig: cfg,
		ProjectDir:    projectDir,
		PresetsDir:    presetsDir,
	}

	// Test Sync
	if err := mgr.SyncComponent("app"); err != nil {
		t.Fatalf("SyncComponent failed: %v", err)
	}

	// Verify Execution
	expectedFile := filepath.Join(projectDir, doneFile)
	if _, err := os.Stat(expectedFile); os.IsNotExist(err) {
		t.Errorf("SyncComponent did not execute install script (file %s missing)", expectedFile)
	}
}

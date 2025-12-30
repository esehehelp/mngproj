package test

import (
	"mngproj/pkg/config"
	"mngproj/pkg/manager"
	"os"
	"path/filepath"
	"testing"
)

func TestResolveComponentWithPresets(t *testing.T) {
	presetsDir := t.TempDir()
	rustPreset := `
[metadata]
type = "rust"
role = "language"
[scripts]
run = "cargo run"
build = "cargo build"
`
	if err := os.WriteFile(filepath.Join(presetsDir, "rust.toml"), []byte(rustPreset), 0644); err != nil {
		t.Fatalf("Failed to write preset: %v", err)
	}

	projCfg := &config.ProjectConfig{
		Project: config.ProjectMeta{Name: "TestProject"},
		Components: []config.ComponentConfig{
			{Name: "backend", Type: "rust", Path: "backend"},
		},
	}

	mgr := &manager.Manager{
		ProjectConfig: projCfg,
		ProjectDir:    "/tmp/dummy/project",
		PresetsDir:    presetsDir,
	}

	comp, err := mgr.ResolveComponent("backend")
	if err != nil {
		t.Fatalf("ResolveComponent(backend) failed: %v", err)
	}

	if comp.Scripts["run"] != "cargo run" {
		t.Errorf("Expected run script 'cargo run', got '%s'", comp.Scripts["run"])
	}
}

func TestResolveComponentWithConflictResolution(t *testing.T) {
	presetsDir := t.TempDir()

	// 1. Language (Score: 0)
	pythonPreset := `
[metadata]
type = "python"
role = "language"
[scripts]
run = "python main.py"
install = "pip install"
test = "unittest"
`
	// 2. Package Manager (Score: 10)
	uvPreset := `
[metadata]
type = "uv"
role = "package_manager"
[scripts]
run = "uv run main.py" 
install = "uv sync"
`
	// 3. Framework (Score: 30)
	djangoPreset := `
[metadata]
type = "django"
role = "framework"
[scripts]
run = "python manage.py runserver"
`

	os.WriteFile(filepath.Join(presetsDir, "python.toml"), []byte(pythonPreset), 0644)
	os.WriteFile(filepath.Join(presetsDir, "uv.toml"), []byte(uvPreset), 0644)
	os.WriteFile(filepath.Join(presetsDir, "django.toml"), []byte(djangoPreset), 0644)

	// Scenario: types = ["python", "uv", "django"]
	// Expectation:
	// - run: "python manage.py runserver" (Framework > PM > Language)
	// - install: "uv sync" (PM > Language. Framework has none.)
	// - test: "unittest" (Language. Others have none.)

	projCfg := &config.ProjectConfig{
		Project: config.ProjectMeta{Name: "ConflictTest"},
		Components: []config.ComponentConfig{
			{
				Name:  "web",
				Types: []string{"python", "uv", "django"},
				Path:  ".",
			},
		},
	}

	mgr := &manager.Manager{
		ProjectConfig: projCfg,
		ProjectDir:    "/tmp/conflict",
		PresetsDir:    presetsDir,
	}

	comp, err := mgr.ResolveComponent("web")
	if err != nil {
		t.Fatalf("ResolveComponent(web) failed: %v", err)
	}

	// Verify RUN (Framework wins)
	if comp.Scripts["run"] != "python manage.py runserver" {
		t.Errorf("RUN: Expected 'python manage.py runserver', got '%s'", comp.Scripts["run"])
	}

	// Verify INSTALL (Package Manager wins)
	if comp.Scripts["install"] != "uv sync" {
		t.Errorf("INSTALL: Expected 'uv sync', got '%s'", comp.Scripts["install"])
	}

	// Verify TEST (Language wins as fallback)
	if comp.Scripts["test"] != "unittest" {
		t.Errorf("TEST: Expected 'unittest', got '%s'", comp.Scripts["test"])
	}
}

func TestAllPresetsLoad(t *testing.T) {
	cwd, _ := os.Getwd()
	// Adjusted path from ../../presets to ../presets
	presetsPath := filepath.Join(cwd, "../presets")

	if _, err := os.Stat(presetsPath); err != nil {
		t.Skipf("Presets directory not found at %s: %v", presetsPath, err)
	}

	err := filepath.WalkDir(presetsPath, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() {
			return nil
		}
		if filepath.Ext(d.Name()) != ".toml" {
			return nil
		}
		
		typeName := d.Name()[:len(d.Name())-5] // remove .toml
		
		// LoadPreset should be able to find it regardless of subdir
		_, err = config.LoadPreset(presetsPath, typeName)
		if err != nil {
			t.Errorf("Failed to load preset %s (found at %s): %v", typeName, path, err)
		}
		return nil
	})
	
	if err != nil {
		t.Fatalf("Walk failed: %v", err)
	}
}

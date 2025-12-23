package manager

import (
	"mngproj/pkg/config"
	"os"
	"path/filepath"
	"testing"
)

func TestUserResolutionOverride(t *testing.T) {
	presetsDir := t.TempDir()

	// 1. Framework (Default Score: 30)
	djangoPreset := `
[metadata]
type = "django"
role = "framework"
[scripts]
run = "django run"
`
	// 2. Custom Tool (Default Score: 20)
	myToolPreset := `
[metadata]
type = "mytool"
role = "tool"
[scripts]
run = "mytool run"
`

	os.WriteFile(filepath.Join(presetsDir, "django.toml"), []byte(djangoPreset), 0644)
	os.WriteFile(filepath.Join(presetsDir, "mytool.toml"), []byte(myToolPreset), 0644)

	// Scenario A: Default Priority (Framework > Tool)
	// Expectation: run = "django run"
	
	cfgDefault := &config.ProjectConfig{
		Components: []config.ComponentConfig{
			{Name: "app", Types: []string{"django", "mytool"}, Path: "."},
		},
	}
	mgrDefault := &Manager{
		ProjectConfig: cfgDefault,
		ProjectDir:    "/tmp",
		PresetsDir:    presetsDir,
	}
	compDefault, _ := mgrDefault.ResolveComponent("app")
	if compDefault.Scripts["run"] != "django run" {
		t.Errorf("Default: Expected 'django run', got '%s'", compDefault.Scripts["run"])
	}

	// Scenario B: User Override (Tool > Framework)
	// Expectation: run = "mytool run"
	
	cfgOverride := &config.ProjectConfig{
		Components: []config.ComponentConfig{
			{Name: "app", Types: []string{"django", "mytool"}, Path: "."},
		},
		Resolution: config.ResolutionConfig{
			RolePriority: map[string]int{
				"tool":      100, // Higher than framework (30)
				"framework": 30,
			},
		},
	}
	mgrOverride := &Manager{
		ProjectConfig: cfgOverride,
		ProjectDir:    "/tmp",
		PresetsDir:    presetsDir,
	}
	compOverride, _ := mgrOverride.ResolveComponent("app")
	if compOverride.Scripts["run"] != "mytool run" {
		t.Errorf("Override: Expected 'mytool run', got '%s'", compOverride.Scripts["run"])
	}
}

func TestInstallScriptDefinition(t *testing.T) {
	// Verify that pip preset has correct install_pkg script with isolation
	cwd, _ := os.Getwd()
	presetsDir := filepath.Join(cwd, "../../presets")

	if _, err := os.Stat(presetsDir); err != nil {
		t.Skip("Presets dir not found")
	}

	// Load pip preset
	pipCfg, err := config.LoadPreset(presetsDir, "pip")
	if err != nil {
		t.Fatalf("Failed to load pip preset: %v", err)
	}

	expectedInstall := "pip install --target=.libs"
	if pipCfg.Scripts["install_pkg"] != expectedInstall {
		t.Errorf("Expected install_pkg='%s', got '%s'", expectedInstall, pipCfg.Scripts["install_pkg"])
	}

	// Check env
	if pipCfg.Env["PYTHONPATH"] == "" {
		t.Error("PYTHONPATH should be set for isolation")
	}
}

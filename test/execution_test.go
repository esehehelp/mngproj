package test

import (
	"bytes"
	"mngproj/pkg/config"
	"mngproj/pkg/manager"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteScriptWithTemplate(t *testing.T) {
	tmpDir := t.TempDir()
	
	// Create dummy preset required for resolution
	os.WriteFile(filepath.Join(tmpDir, "custom.toml"), []byte(`[metadata]
type="custom"
role="tool"
`), 0644)
	
	cfg := &config.ProjectConfig{
		Components: []config.ComponentConfig{
			{
				Name: "app",
				Type: "custom",
				Path: ".",
				Scripts: map[string]string{
					"echo_args": "echo {{range .Args}}{{.}} {{end}}",
					"echo_env":  "echo {{.Env.MY_VAR}}",
				},
				Env: map[string]string{
					"MY_VAR": "hello_world",
				},
			},
		},
	}
	
	mgr := &manager.Manager{
		ProjectConfig: cfg,
		ProjectDir:    tmpDir,
		PresetsDir:    tmpDir, // not used
	}

	// Test 1: Args Substitution
	var stdout bytes.Buffer
	err := mgr.ExecuteScript("app", "echo_args", []string{"foo", "bar"}, &stdout, nil)
	if err != nil {
		t.Fatalf("ExecuteScript failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "foo bar") {
		t.Errorf("Expected 'foo bar', got '%s'", stdout.String())
	}

	// Test 2: Env Substitution
	stdout.Reset()
	err = mgr.ExecuteScript("app", "echo_env", nil, &stdout, nil)
	if err != nil {
		t.Fatalf("ExecuteScript failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "hello_world") {
		t.Errorf("Expected 'hello_world', got '%s'", stdout.String())
	}
}

func TestExecuteScriptExternalFile(t *testing.T) {
	tmpDir := t.TempDir()
	scriptPath := filepath.Join(tmpDir, "script.sh")
	os.WriteFile(scriptPath, []byte("echo from file"), 0755)

	cfg := &config.ProjectConfig{
		Components: []config.ComponentConfig{
			{
				Name: "app",
				Path: ".",
				Scripts: map[string]string{
					"run_file": "file:script.sh",
				},
			},
		},
	}
	
	mgr := &manager.Manager{
		ProjectConfig: cfg,
		ProjectDir:    tmpDir,
		PresetsDir:    tmpDir,
	}

	var stdout bytes.Buffer
	err := mgr.ExecuteScript("app", "run_file", nil, &stdout, nil)
	if err != nil {
		t.Fatalf("ExecuteScript failed: %v", err)
	}
	if !strings.Contains(stdout.String(), "from file") {
		t.Errorf("Expected 'from file', got '%s'", stdout.String())
	}
}

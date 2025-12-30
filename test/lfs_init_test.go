package test

import (
	"mngproj/pkg/config"
	"mngproj/pkg/manager"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
    "strings"
)

func TestInitGitignore(t *testing.T) {
    cwd, _ := os.Getwd()
    binPath := filepath.Join(cwd, "../mngproj_bin_lfs")
    if err := exec.Command("go", "build", "-o", binPath, "../cmd/mngproj/main.go").Run(); err != nil {
        t.Fatalf("Build failed: %v", err)
    }
    defer os.Remove(binPath)
    
    tmpDir := t.TempDir()
    presetsDir := filepath.Join(tmpDir, "presets")
    os.MkdirAll(presetsDir, 0755)
    os.WriteFile(filepath.Join(presetsDir, "go.toml"), []byte(`
gitignore=["*.exe", "*.test"]
[metadata]
type="go"
`), 0644)
    
    env := append(os.Environ(), "MNGPROJ_PRESETS_DIR="+presetsDir)
    
    cmd := exec.Command(binPath, "init")
    cmd.Dir = tmpDir
    cmd.Env = env
    if out, err := cmd.CombinedOutput(); err != nil {
        t.Fatalf("Init failed: %v, out: %s", err, out)
    }
    
    content, err := os.ReadFile(filepath.Join(tmpDir, ".gitignore"))
    if err != nil {
        t.Fatal("gitignore not created")
    }
    if !strings.Contains(string(content), "*.exe") {
        t.Error("gitignore missing preset content")
    }
}

func TestLFS(t *testing.T) {
    cwd, _ := os.Getwd()
    binPath := filepath.Join(cwd, "../mngproj_bin_lfs_2")
    if err := exec.Command("go", "build", "-o", binPath, "../cmd/mngproj/main.go").Run(); err != nil {
        t.Fatalf("Build failed: %v", err)
    }
    defer os.Remove(binPath)
    
    tmpDir := t.TempDir()
    // create mngproj.toml
    os.WriteFile(filepath.Join(tmpDir, "mngproj.toml"), []byte(`[project]
name="lfstest"
[[components]]
name="app"
path="."
`), 0644)

    // Create large file (dummy) > 1MB. Threshold 1MB
    largeFile := filepath.Join(tmpDir, "big.bin")
    f, _ := os.Create(largeFile)
    f.Write(make([]byte, 2*1024*1024)) // 2MB
    f.Close()
    
    cmd := exec.Command(binPath, "lfs", "1") // 1MB threshold
    cmd.Dir = tmpDir
    if out, err := cmd.CombinedOutput(); err != nil {
        t.Fatalf("LFS failed: %v, out: %s", err, out)
    }
    
    attr, err := os.ReadFile(filepath.Join(tmpDir, ".gitattributes"))
    if err != nil {
        t.Fatal("gitattributes not created")
    }
    if !strings.Contains(string(attr), "*.bin filter=lfs") {
        t.Errorf("gitattributes content mismatch: %s", string(attr))
    }
}

func TestValidateTools(t *testing.T) {
    // Unit test for Manager.ValidateTools
    tmpDir := t.TempDir()
    presetsDir := filepath.Join(tmpDir, "presets")
    os.MkdirAll(presetsDir, 0755)
    os.WriteFile(filepath.Join(presetsDir, "fail.toml"), []byte(`
[metadata]
type="fail"
required_tools=["non_existent_tool_xyz"]
`), 0644)

    cfg := &config.ProjectConfig{
        Components: []config.ComponentConfig{
            {Name: "app", Type: "fail", Path: "."},
        },
    }
    
    mgr := &manager.Manager{
        ProjectConfig: cfg,
        PresetsDir: presetsDir,
    }
    
    if err := mgr.ValidateTools(); err == nil {
        t.Error("Expected error for missing tool, got nil")
    }
}

package test

import (
    "os"
    "os/exec"
    "path/filepath"
    "testing"
    "strings"
)

func TestIntegrationFullCycle(t *testing.T) {
    // 1. Build Binary
    cwd, _ := os.Getwd()
    // Assuming test runs in mngproj/test
    binPath := filepath.Join(cwd, "../mngproj_bin")
    // Source is in ../cmd/mngproj/main.go
    buildCmd := exec.Command("go", "build", "-o", binPath, "../cmd/mngproj/main.go")
    if err := buildCmd.Run(); err != nil {
        t.Fatalf("Failed to build binary: %v", err)
    }
    defer os.Remove(binPath)

    // 2. Setup Workspace
    tmpDir := t.TempDir()
    // Create presets dir for isolation
    presetsDir := filepath.Join(tmpDir, "presets")
    os.MkdirAll(presetsDir, 0755)
    // Create dummy preset
    os.WriteFile(filepath.Join(presetsDir, "dummy.toml"), []byte(`
[metadata]
type="dummy"
role="tool"
[scripts]
build="echo build_dummy {{range .Args}}{{.}} {{end}}"
run="echo run_dummy"
`), 0644)

    env := append(os.Environ(), "MNGPROJ_PRESETS_DIR="+presetsDir)

    // 3. Init
    cmd := exec.Command(binPath, "init")
    cmd.Dir = tmpDir
    cmd.Env = env
    if out, err := cmd.CombinedOutput(); err != nil {
        t.Fatalf("Init failed: %v\nOutput: %s", err, out)
    }

    // 4. Update mngproj.toml to use our dummy preset
    configPath := filepath.Join(tmpDir, "mngproj.toml")
    newConfig := `[project]
name = "IntegrationTest"
[[components]]
name = "api"
type = "dummy"
path = "."
`
    os.WriteFile(configPath, []byte(newConfig), 0644)

    // 5. Build (with args)
    cmd = exec.Command(binPath, "build", "api", "release")
    cmd.Dir = tmpDir
    cmd.Env = env
    out, err := cmd.CombinedOutput()
    if err != nil {
         t.Fatalf("Build failed: %v\nOutput: %s", err, out)
    }
    if !strings.Contains(string(out), "build_dummy release") {
        t.Errorf("Build output mismatch. Got: %s", string(out))
    }

    // 6. Up (Parallel Execution)
    cmd = exec.Command(binPath, "up", "api")
    cmd.Dir = tmpDir
    cmd.Env = env
    out, err = cmd.CombinedOutput()
    if err != nil {
         t.Fatalf("Up failed: %v\nOutput: %s", err, out)
    }
    if !strings.Contains(string(out), "[api] run_dummy") {
        t.Errorf("Up output mismatch. Got: %s", string(out))
    }
}

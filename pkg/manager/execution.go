package manager

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

func (m *Manager) ExecuteScript(componentName, scriptName string, args []string) error {
	comp, err := m.ResolveComponent(componentName)
	if err != nil {
		return err
	}

	cmdStr, ok := comp.Scripts[scriptName]
	if !ok {
		return fmt.Errorf("script %q not defined for component %q", scriptName, componentName)
	}

	// Simple template replacement (can be expanded to text/template later)
	// For now, just handle static strings if any. 
	// The shell will handle most env vars if we pass them correctly.
	
	// Prepare environment
	env := os.Environ()
	for k, v := range comp.Env {
		// Expand values like $HOME, etc.
		expandedV := os.ExpandEnv(v)
		env = append(env, fmt.Sprintf("%s=%s", k, expandedV))
	}
	// Inject MNGPROJ_ROOT
	env = append(env, fmt.Sprintf("MNGPROJ_ROOT=%s", m.ProjectDir))

	// Construct command
	// We run through a shell to allow complex commands (pipes, etc.)
	fullCmd := cmdStr
	if len(args) > 0 {
		fullCmd += " " + strings.Join(args, " ")
	}

	fmt.Printf("[%s] Executing: %s\n", componentName, fullCmd)

	cmd := exec.Command("sh", "-c", fullCmd)
	cmd.Dir = comp.AbsPath
	cmd.Env = env
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	cmd.Stdin = os.Stdin

	return cmd.Run()
}

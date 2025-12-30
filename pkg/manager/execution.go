package manager

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"text/template"
)

type ScriptContext struct {
	Args []string
	Env  map[string]string
}

// ExecuteScript runs the script and waits for it to finish
func (m *Manager) ExecuteScript(componentName, scriptName string, args []string, stdout, stderr io.Writer) error {
	cmd, err := m.ExecuteScriptAsync(componentName, scriptName, args, stdout, stderr)
	if err != nil {
		return err
	}
	return cmd.Wait()
}

// ExecuteScriptAsync prepares and starts the script, returning the *exec.Cmd object
// The caller is responsible for waiting on the command.
func (m *Manager) ExecuteScriptAsync(componentName, scriptName string, args []string, stdout, stderr io.Writer) (*exec.Cmd, error) {
	comp, err := m.ResolveComponent(componentName)
	if err != nil {
		return nil, err
	}

	cmdStr, ok := comp.Scripts[scriptName]
	if !ok {
		return nil, fmt.Errorf("script %q not defined for component %q", scriptName, componentName)
	}

	// Prepare environment
	envMap := make(map[string]string)
	env := os.Environ()
	
	// Mapper for variable expansion
	expandMapper := func(key string) string {
		switch key {
		case "MNGPROJ_ROOT":
			return m.ProjectDir
		case "MNGPROJ_COMPONENT_ROOT", "COMPONENT_ROOT":
			return comp.AbsPath
		}
		return os.Getenv(key)
	}

	for k, v := range comp.Env {
		// Expand values like $HOME, ${MNGPROJ_ROOT}, etc.
		expandedV := os.Expand(v, expandMapper)
		env = append(env, fmt.Sprintf("%s=%s", k, expandedV))
		envMap[k] = expandedV
	}
	// Inject MNGPROJ_ROOT and MNGPROJ_COMPONENT_ROOT
	env = append(env, fmt.Sprintf("MNGPROJ_ROOT=%s", m.ProjectDir))
	envMap["MNGPROJ_ROOT"] = m.ProjectDir
	env = append(env, fmt.Sprintf("MNGPROJ_COMPONENT_ROOT=%s", comp.AbsPath))
	envMap["MNGPROJ_COMPONENT_ROOT"] = comp.AbsPath

	// Handle "file:" prefix
	if strings.HasPrefix(cmdStr, "file:") {
		scriptPath := strings.TrimPrefix(cmdStr, "file:")
		if !filepath.IsAbs(scriptPath) {
			scriptPath = filepath.Join(m.ProjectDir, scriptPath)
		}
		
		content, err := os.ReadFile(scriptPath)
		if err != nil {
			return nil, fmt.Errorf("failed to read script file %s: %w", scriptPath, err)
		}
		cmdStr = string(content)
	}

	// Process Command String
	var fullCmd string
	
	// Check if template syntax is used
	if strings.Contains(cmdStr, "{{") {
		tmpl, err := template.New("script").Parse(cmdStr)
		if err != nil {
			return nil, fmt.Errorf("failed to parse script template: %w", err)
		}
		
		ctx := ScriptContext{
			Args: args,
			Env:  envMap,
		}
		
		var buf bytes.Buffer
		if err := tmpl.Execute(&buf, ctx); err != nil {
			return nil, fmt.Errorf("failed to execute script template: %w", err)
		}
		fullCmd = buf.String()
	} else {
		// Legacy/Simple behavior: append args
		fullCmd = cmdStr
		if len(args) > 0 {
			fullCmd += " " + strings.Join(args, " ")
		}
	}

	// Determine outputs
	outW := stdout
	if outW == nil {
		outW = os.Stdout
	}
	errW := stderr
	if errW == nil {
		errW = os.Stderr
	}

	// Log execution
	if stdout == nil { 
		fmt.Printf("[%s] Executing: %s\n", componentName, fullCmd)
	}

	var shell, flag string
	if runtime.GOOS == "windows" {
		shell = "powershell"
		flag = "-Command"
	} else {
		shell = "sh"
		flag = "-c"
	}

	cmd := exec.Command(shell, flag, fullCmd)
	cmd.Dir = comp.AbsPath
	cmd.Env = env
	cmd.Stdout = outW
	cmd.Stderr = errW
	cmd.Stdin = os.Stdin

	if err := cmd.Start(); err != nil {
		return nil, fmt.Errorf("failed to start command: %w", err)
	}
	
	return cmd, nil
}

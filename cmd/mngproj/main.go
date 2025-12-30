package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mngproj/pkg/manager"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"text/tabwriter"
	"time"

	"mngproj/pkg/config"
)

func main() {
	if len(os.Args) < 2 {
		printUsage()
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Init does not require an existing manager (config)
	if os.Args[1] == "init" {
		handleInit(os.Args[2:])
		return
	}

	// For other commands, load manager
	mgr, err := manager.New(cwd)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}

	switch os.Args[1] {
	case "run":
		handleRun(mgr, os.Args[2:])
	case "build":
		handleBuild(mgr, os.Args[2:])
	case "add":
		handleAdd(mgr, os.Args[2:])
	case "sync":
		handleSync(mgr, os.Args[2:])
	case "up":
		handleUp(mgr, os.Args[2:])
	case "watch":
		handleWatch(mgr, os.Args[2:])
	case "remove":
		handleRemove(mgr, os.Args[2:])
	case "ls":
		handleLs(mgr)
	case "lsproj":
		handleLsproj()
	case "query":
		handleQuery(mgr, os.Args[2:])
	case "info":
		handleInfo(mgr)
	default:
		// Attempt to handle as a generic script command
		handleGenericScript(mgr, os.Args[1], os.Args[2:])
	}
}

func printUsage() {
	fmt.Println("Usage: mngproj <command> [args...]")
	fmt.Println("Commands:")
	fmt.Println("  init     Initialize a new project")
	fmt.Println("  run      Run a component")
	fmt.Println("  build    Build a component")
	fmt.Println("  add      Add a package to a component and sync")
	fmt.Println("  sync     Sync dependencies for one or all components")
	fmt.Println("  up       Run components in parallel with aggregated logs")
	fmt.Println("  watch    Watch for changes and reload (Hot Reload)")
	fmt.Println("  remove   Remove a package from a component")
	fmt.Println("  ls       List components of current project")
	fmt.Println("  lsproj   List all projects in the current directory tree")
	fmt.Println("  query    Query metadata of current project")
	fmt.Println("  info     Show current project info")
	fmt.Println("  <script> Run a custom script defined in mngproj.toml")
}

func handleGenericScript(m *manager.Manager, scriptName string, args []string) {
	if len(args) == 0 {
		fmt.Printf("Unknown command '%s'.\n", scriptName)
		fmt.Println("If this is a custom script, usage is: mngproj <script> <component> [args...]")
		printUsage()
		os.Exit(1)
	}

	component := args[0]
	scriptArgs := args[1:]

	if err := m.ExecuteScript(component, scriptName, scriptArgs, nil, nil); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
}

func handleLsproj() {
	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	projectRoots, err := manager.FindAllProjectConfigs(cwd)
	if err != nil {
		log.Fatal(err)
	}

	if len(projectRoots) == 0 {
		fmt.Printf("No mngproj projects found in %q\n", cwd)
		return
	}

	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Project Path\tProject Name")
	for _, root := range projectRoots {
		// Load config for each project to get its name
		cfgPath := filepath.Join(root, "mngproj.toml")
		cfg, err := config.LoadProjectConfig(cfgPath)
		if err != nil {
			// Log error but continue
			fmt.Fprintf(os.Stderr, "Error loading config for %q: %v\n", root, err)
			continue
		}
		// Make path relative to CWD for readability
		relPath, _ := filepath.Rel(cwd, root)
		if relPath == "" { // If current directory is a project root
			relPath = "."
		}
		fmt.Fprintf(w, "%s\t%s\n", relPath, cfg.Project.Name)
	}
	w.Flush()
}

func handleInit(args []string) {
	fmt.Println("Initializing new mngproj.toml...")
	content := `[project]
name = "new-project"
description = "Created by mngproj init"

[[components]]
name = "app"
type = "go"
path = "."
`
	if err := os.WriteFile("mngproj.toml", []byte(content), 0644); err != nil {
		log.Fatal(err)
	}
	fmt.Println("Created mngproj.toml")
}

func handleRun(m *manager.Manager, args []string) {
	if len(args) == 0 {
		fmt.Println("Please specify a component name to run.")
		handleLs(m)
		return
	}
	
	component := args[0]
	scriptArgs := args[1:]
	
	if err := m.ExecuteScript(component, "run", scriptArgs, nil, nil); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
}

func handleBuild(m *manager.Manager, args []string) {
	if len(args) == 0 {
		fmt.Println("Please specify a component name to build.")
		return
	}
	
	component := args[0]
	scriptArgs := args[1:]
	
	if err := m.ExecuteScript(component, "build", scriptArgs, nil, nil); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
}

func handleAdd(m *manager.Manager, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: mngproj add <component> <package...>")
		return
	}
	component := args[0]
	pkgs := args[1:]

	for _, pkg := range pkgs {
		fmt.Printf("Adding dependency %q to component %q...\n", pkg, component)
		if err := m.AddDependency(component, pkg); err != nil {
			log.Fatalf("Failed to add dependency: %v", err)
		}
	}
	
	fmt.Println("Syncing dependencies...")
	if err := m.SyncComponent(component); err != nil {
		log.Fatalf("Sync failed: %v", err)
	}
	fmt.Println("Done.")
}

func handleSync(m *manager.Manager, args []string) {
	var components []string
	if len(args) > 0 {
		components = args
	} else {
		components = m.ListComponents()
	}

	for _, comp := range components {
		fmt.Printf("Syncing component %q...\n", comp)
		if err := m.SyncComponent(comp); err != nil {
			log.Printf("Failed to sync component %q: %v\n", comp, err)
			os.Exit(1)
		}
	}
	fmt.Println("All synced.")
}

func handleUp(m *manager.Manager, args []string) {
	targetComps := make(map[string]bool)

	if len(args) == 0 {
		// All components
		for _, c := range m.ListComponents() {
			targetComps[c] = true
		}
	} else {
		// Resolve args
		allCompMap := make(map[string]bool)
		for _, c := range m.ListComponents() {
			allCompMap[c] = true
		}

		for _, arg := range args {
			// 1. Is it a component?
			if allCompMap[arg] {
				targetComps[arg] = true
				continue
			}

			// 2. Is it a group?
			groupComps := m.ListComponentsByGroup(arg)
			if len(groupComps) > 0 {
				for _, c := range groupComps {
					targetComps[c] = true
				}
				continue
			}

			fmt.Printf("Warning: Argument %q matches no component or group.\n", arg)
		}
	}

	if len(targetComps) == 0 {
		fmt.Println("No components found to start.")
		return
	}

	// Convert map to slice
	var components []string
	for c := range targetComps {
		components = append(components, c)
	}

	fmt.Printf("Starting %d components: %v\n", len(components), components)

	var wg sync.WaitGroup
	for _, name := range components {
		wg.Add(1)
		go func(compName string) {
			defer wg.Done()
			pw := &PrefixWriter{prefix: compName, writer: os.Stdout}
			if err := m.ExecuteScript(compName, "run", nil, pw, pw); err != nil {
				fmt.Fprintf(pw, "Error: %v\n", err)
			}
		}(name)
	}
	wg.Wait()
}

func handleWatch(m *manager.Manager, args []string) {
	var comps []string
	if len(args) > 0 {
		comps = args
	} else {
		comps = m.ListComponents()
	}
	
	var wg sync.WaitGroup
	for _, c := range comps {
		wg.Add(1)
		go func(name string) {
			defer wg.Done()
			watchComponent(m, name)
		}(c)
	}
	wg.Wait()
}

func watchComponent(m *manager.Manager, compName string) {
	comp, err := m.ResolveComponent(compName)
	if err != nil {
		log.Printf("[%s] Watch Error: %v", compName, err)
		return
	}
	
	root := comp.AbsPath
	fmt.Printf("[%s] Watching %s for changes...\n", compName, root)

	var currentCmd *exec.Cmd
	var lastModTime time.Time
	
	restart := make(chan bool, 1)
	restart <- true

	go func() {
		for {
			time.Sleep(1 * time.Second)
			
			var maxModTime time.Time
			err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
				if err != nil { return nil }
				if info.IsDir() {
					if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" || info.Name() == "target" || info.Name() == "dist" || info.Name() == "build" {
						return filepath.SkipDir
					}
				}
				if !info.IsDir() {
					if info.ModTime().After(maxModTime) {
						maxModTime = info.ModTime()
					}
				}
				return nil
			})
			
			if err == nil {
				if lastModTime.IsZero() {
					lastModTime = maxModTime
				} else if maxModTime.After(lastModTime) {
					fmt.Printf("[%s] Change detected. Reloading...\n", compName)
					lastModTime = maxModTime
					restart <- true
				}
			}
		}
	}()

	for range restart {
		if currentCmd != nil && currentCmd.Process != nil {
			// Try to kill process group to catch children
			syscall.Kill(-currentCmd.Process.Pid, syscall.SIGKILL)
			currentCmd.Process.Kill()
			currentCmd.Wait()
		}

		pw := &PrefixWriter{prefix: compName, writer: os.Stdout}
		// Pass SysProcAttr to set process group for group kill support
		// Note: ExecuteScriptAsync creates the cmd, we need to modify it inside if possible.
		// Current API doesn't allow modifying cmd before Start.
		// For now simple Process.Kill is used.
		
		cmd, err := m.ExecuteScriptAsync(compName, "run", nil, pw, pw)
		if err != nil {
			fmt.Fprintf(pw, "Start Error: %v\n", err)
			currentCmd = nil
		} else {
			// Set process group ID so we can kill children if needed
			// But since we can't inject it before Start() in ExecuteScriptAsync without changing API again,
			// we rely on basic kill.
			// Ideally ExecuteScriptAsync should take an option struct.
			
			currentCmd = cmd
			go func() {
				cmd.Wait()
			}()
		}
	}
}

func handleRemove(m *manager.Manager, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: mngproj remove <component> <package...>")
		return
	}
	component := args[0]
	pkgArgs := args[1:]
	
	if err := m.ExecuteScript(component, "remove_pkg", pkgArgs, nil, nil); err != nil {
		log.Fatalf("Remove failed: %v", err)
	}
}

func handleLs(m *manager.Manager) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Name\tType\tPath")
	for _, c := range m.ProjectConfig.Components {
		fmt.Fprintf(w, "%s\t%s\t%s\n", c.Name, c.Type, c.Path)
	}
	w.Flush()
}

func handleInfo(m *manager.Manager) {
	fmt.Printf("Project: %s\n", m.ProjectConfig.Project.Name)
	fmt.Printf("Root: %s\n", m.ProjectDir)
	fmt.Printf("Presets: %s\n", m.PresetsDir)
	fmt.Printf("Components: %d\n", len(m.ProjectConfig.Components))
}

func handleQuery(m *manager.Manager, args []string) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(m.ProjectConfig.Components); err != nil {
		log.Fatalf("Failed to encode components: %v", err)
	}
}

// PrefixWriter prefixes each line with a tag
type PrefixWriter struct {
	prefix string
	writer io.Writer
}

func (w *PrefixWriter) Write(p []byte) (n int, err error) {
	lines := bytes.Split(p, []byte("\n"))
	for i, line := range lines {
		if len(line) == 0 && i == len(lines)-1 {
			continue
		}
		// Note: This naive implementation might interleave outputs if writer is not concurrent-safe.
		// Since os.Stdout writes are generally atomic for small buffers, it's acceptable for a CLI tool.
		out := fmt.Sprintf("[%s] %s\n", w.prefix, string(line))
		w.writer.Write([]byte(out))
	}
	return len(p), nil
}

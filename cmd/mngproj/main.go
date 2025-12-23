package main

import (
	"encoding/json"
	"fmt"
	"log"
	"mngproj/pkg/manager"
	"os"
	"path/filepath" // Added for filepath.Rel
	"text/tabwriter"

	"mngproj/pkg/config" // Added for config.LoadProjectConfig
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
	case "install":
		handleInstall(mgr, os.Args[2:])
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
		printUsage()
		os.Exit(1)
	}
}

func printUsage() {
	fmt.Println("Usage: mngproj <command> [args...]")
	fmt.Println("Commands:")
	fmt.Println("  init     Initialize a new project")
	fmt.Println("  run      Run a component")
	fmt.Println("  build    Build a component")
	fmt.Println("  install  Install a package to a component")
	fmt.Println("  remove   Remove a package from a component")
	fmt.Println("  ls       List components of current project")
	fmt.Println("  lsproj   List all projects in the current directory tree")
	fmt.Println("  query    Query metadata of current project")
	fmt.Println("  info     Show current project info")
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
	
	if err := m.ExecuteScript(component, "run", scriptArgs); err != nil {
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
	
	if err := m.ExecuteScript(component, "build", scriptArgs); err != nil {
		log.Fatalf("Build failed: %v", err)
	}
}

func handleInstall(m *manager.Manager, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: mngproj install <component> <package...>")
		return
	}
	component := args[0]
	pkgArgs := args[1:]
	
	if err := m.ExecuteScript(component, "install_pkg", pkgArgs); err != nil {
		log.Fatalf("Install failed: %v", err)
	}
}

func handleRemove(m *manager.Manager, args []string) {
	if len(args) < 2 {
		fmt.Println("Usage: mngproj remove <component> <package...>")
		return
	}
	component := args[0]
	pkgArgs := args[1:]
	
	if err := m.ExecuteScript(component, "remove_pkg", pkgArgs); err != nil {
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

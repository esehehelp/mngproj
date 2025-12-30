package cmd

import (
	"encoding/json"
	"fmt"
	"log"
	"mngproj/pkg/config"
	"mngproj/pkg/manager"
	"mngproj/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
	"text/tabwriter"
)

func PrintUsage() {
	fmt.Println("Usage: mngproj <command> [args...]")
	fmt.Println("Commands:")
	fmt.Println("  init     Initialize a new project")
	fmt.Println("  run      Run a component")
	fmt.Println("  build    Build a component")
	fmt.Println("  add      Add a package to a component and sync")
	fmt.Println("  sync     Sync dependencies for one or all components")
	fmt.Println("  up       Run components in parallel with aggregated logs")
	fmt.Println("  watch    Watch for changes and reload (Hot Reload)")
	fmt.Println("  lfs      Check for large files and update .gitattributes")
	fmt.Println("  install-self Build and install mngproj to system")
	fmt.Println("  remove   Remove a package from a component")
	fmt.Println("  ls       List components of current project")
	fmt.Println("  lsproj   List all projects in the current directory tree")
	fmt.Println("  query    Query metadata of current project")
	fmt.Println("  info     Show current project info")
	fmt.Println("  <script> Run a custom script defined in mngproj.toml")
}

func HandleGenericScript(m *manager.Manager, scriptName string, args []string) {
	if len(args) == 0 {
		fmt.Printf("Unknown command '%s'.\n", scriptName)
		fmt.Println("If this is a custom script, usage is: mngproj <script> <component> [args...]")
		PrintUsage()
		os.Exit(1)
	}

	component := args[0]
	scriptArgs := args[1:]

	if err := m.ExecuteScript(component, scriptName, scriptArgs, nil, nil); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
}

func HandleInit(args []string) {
	targetType := "go"
	if len(args) > 0 {
		targetType = args[0]
	}
	if err := manager.InitializeProject(targetType); err != nil {
		log.Fatalf("Init failed: %v", err)
	}
}

func HandleRun(m *manager.Manager, args []string) {
	if len(args) == 0 {
		fmt.Println("Please specify a component name to run.")
		HandleLs(m)
		return
	}
	component := args[0]
	scriptArgs := args[1:]
	if err := m.ExecuteScript(component, "run", scriptArgs, nil, nil); err != nil {
		log.Fatalf("Execution failed: %v", err)
	}
}

func HandleBuild(m *manager.Manager, args []string) {
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

func HandleAdd(m *manager.Manager, args []string) {
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

func HandleSync(m *manager.Manager, args []string) {
	if err := m.ValidateTools(); err != nil {
		log.Fatalf("Tool validation failed: %v", err)
	}

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

func HandleUp(m *manager.Manager, args []string) {
	targetComps := make(map[string]bool)

	if len(args) == 0 {
		for _, c := range m.ListComponents() {
			targetComps[c] = true
		}
	} else {
		allCompMap := make(map[string]bool)
		for _, c := range m.ListComponents() {
			allCompMap[c] = true
		}

		for _, arg := range args {
			if allCompMap[arg] {
				targetComps[arg] = true
				continue
			}
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
			pw := &utils.PrefixWriter{Prefix: compName, Writer: os.Stdout}
			if err := m.ExecuteScript(compName, "run", nil, pw, pw); err != nil {
				fmt.Fprintf(pw, "Error: %v\n", err)
			}
		}(name)
	}
	wg.Wait()
}

func HandleWatch(m *manager.Manager, args []string) {
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
			m.WatchComponent(name)
		}(c)
	}
	wg.Wait()
}

func HandleLfs(m *manager.Manager, args []string) {
	thresholdMB := 100
	if len(args) > 0 {
		if v, err := strconv.Atoi(args[0]); err == nil {
			thresholdMB = v
		}
	}
	if err := m.CheckLFS(thresholdMB); err != nil {
		log.Fatal(err)
	}
}

func HandleInstallSelf() {
	fmt.Println("Installing mngproj to system...")
	if _, err := exec.LookPath("go"); err != nil {
		log.Fatal("Go compiler not found. Please install Go to use install-self.")
	}
	cmd := exec.Command("go", "install", "./cmd/mngproj")
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		log.Fatalf("Installation failed: %v. Please run from the project root.", err)
	}
	fmt.Println("Successfully installed mngproj.")
}

func HandleRemove(m *manager.Manager, args []string) {
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

func HandleLs(m *manager.Manager) {
	w := tabwriter.NewWriter(os.Stdout, 0, 0, 3, ' ', 0)
	fmt.Fprintln(w, "Name\tType\tPath")
	for _, c := range m.ProjectConfig.Components {
		fmt.Fprintf(w, "%s\t%s\t%s\n", c.Name, c.Type, c.Path)
	}
	w.Flush()
}

func HandleLsproj() {
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
		cfgPath := filepath.Join(root, "mngproj.toml")
		cfg, err := config.LoadProjectConfig(cfgPath)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Error loading config for %q: %v\n", root, err)
			continue
		}
		relPath, _ := filepath.Rel(cwd, root)
		if relPath == "" {
			relPath = "."
		}
		fmt.Fprintf(w, "%s\t%s\n", relPath, cfg.Project.Name)
	}
	w.Flush()
}

func HandleQuery(m *manager.Manager, args []string) {
	encoder := json.NewEncoder(os.Stdout)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(m.ProjectConfig.Components); err != nil {
		log.Fatalf("Failed to encode components: %v", err)
	}
}

func HandleInfo(m *manager.Manager) {
	fmt.Printf("Project: %s\n", m.ProjectConfig.Project.Name)
	fmt.Printf("Root: %s\n", m.ProjectDir)
	fmt.Printf("Presets: %s\n", m.PresetsDir)
	fmt.Printf("Components: %d\n", len(m.ProjectConfig.Components))
}
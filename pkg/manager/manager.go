package manager

import (
	"fmt"
	"mngproj/pkg/config"
	"os"
	"path/filepath"
)

type Manager struct {
	ProjectConfig *config.ProjectConfig
	ProjectDir    string
	PresetsDir    string
}

func New(startDir string) (*Manager, error) {
	configPath, err := FindConfigFile(startDir)
	if err != nil {
		return nil, err
	}

	cfg, err := config.LoadProjectConfig(configPath)
	if err != nil {
		return nil, err
	}

	configDir := filepath.Dir(configPath)
	projectDir := configDir

	if cfg.Project.Root != "" {
		if filepath.IsAbs(cfg.Project.Root) {
			projectDir = cfg.Project.Root
		} else {
			projectDir = filepath.Join(configDir, cfg.Project.Root)
		}
	}

	return &Manager{
		ProjectConfig: cfg,
		ProjectDir:    projectDir,
		PresetsDir:    determinePresetsDir(),
	}, nil
}

func determinePresetsDir() string {
	// Priority:
	// 1. MNGPROJ_PRESETS_DIR env var
	// 2. $HOME/.config/mngproj/presets
	// 3. ./presets (relative to executable - for dev)
	
	if env := os.Getenv("MNGPROJ_PRESETS_DIR"); env != "" {
		return env
	}

	home, err := os.UserHomeDir()
	if err == nil {
		path := filepath.Join(home, ".config", "mngproj", "presets")
		if _, err := os.Stat(path); err == nil {
			return path
		}
	}
	
	// Fallback to local presets (e.g. for development)
	return "presets"
}

// ResolvedComponent represents a fully merged configuration for a component

type ResolvedComponent struct {

	Name         string

	Type         string

	AbsPath      string

	ManifestFile string

	Env          map[string]string

	Scripts      map[string]string

}



// Default Role Priority Scores

var defaultRolePriority = map[string]int{

	"framework":       30,

	"tool":            20,

	"package_manager": 10,

	"language":        0,

}



func (m *Manager) getRoleScore(role string) int {

	// 1. Check User Override

	if m.ProjectConfig.Resolution.RolePriority != nil {

		if score, ok := m.ProjectConfig.Resolution.RolePriority[role]; ok {

			return score

		}

	}

	// 2. Check Default

	if score, ok := defaultRolePriority[role]; ok {

		return score

	}

	return 0 // Default for unknown roles

}



func (m *Manager) ResolveComponent(name string) (*ResolvedComponent, error) {

	var compConfig *config.ComponentConfig

	for i := range m.ProjectConfig.Components {

		if m.ProjectConfig.Components[i].Name == name {

			compConfig = &m.ProjectConfig.Components[i]

			break

		}

	}

	if compConfig == nil {

		return nil, fmt.Errorf("component %q not found", name)

	}



	// Normalize types

	typeNames := compConfig.Types

	if len(typeNames) == 0 && compConfig.Type != "" {

		typeNames = []string{compConfig.Type}

	}



	resolved := &ResolvedComponent{

		Name:    compConfig.Name,

		Type:    "",

		AbsPath: filepath.Join(m.ProjectDir, compConfig.Path),

		Env:     make(map[string]string),

		Scripts: make(map[string]string),

	}

	if len(typeNames) > 0 {

		resolved.Type = typeNames[0]

	}



	// map[scriptName]score

	scriptScores := make(map[string]int)

	maxManifestScore := -1



	// 1. Apply Presets with Role-based Priority

	for _, tName := range typeNames {

		preset, err := config.LoadPreset(m.PresetsDir, tName)

		if err != nil {

			return nil, fmt.Errorf("failed to load preset %q: %w", tName, err)

		}



		currentScore := m.getRoleScore(preset.Metadata.Role)



		// Resolve ManifestFile

		if preset.Metadata.ManifestFile != "" {

			if currentScore > maxManifestScore {

				resolved.ManifestFile = preset.Metadata.ManifestFile

				maxManifestScore = currentScore

			}

		}



		// Merge Env (Accumulate/Overwrite logic - last wins in types list)

		for k, v := range preset.Env {

			resolved.Env[k] = v

		}



		// Merge Scripts (Priority based)

		for script, cmd := range preset.Scripts {

			existingScore, exists := scriptScores[script]

			

			// Update if:

			// 1. Script doesn't exist yet

			// 2. Current preset has higher priority score

			// 3. Scores are equal (Last Wins - user order preference)

			if !exists || currentScore >= existingScore {

				resolved.Scripts[script] = cmd

				scriptScores[script] = currentScore

			}

		}

	}



	// 2. Override with Component config (Highest priority: User manual override)

	for k, v := range compConfig.Env {

		resolved.Env[k] = v

	}

	for k, v := range compConfig.Scripts {

		resolved.Scripts[k] = v

	}



	return resolved, nil

}



func (m *Manager) ListComponents() []string {

	names := make([]string, len(m.ProjectConfig.Components))

	for i, c := range m.ProjectConfig.Components {

		names[i] = c.Name

	}

	return names

}



// AddDependency adds a package to the component's dependency list and saves the config.

// It also updates the manifest file if applicable.

func (m *Manager) AddDependency(compName, pkgName string) error {

	// 1. Find component

	var comp *config.ComponentConfig

	for i := range m.ProjectConfig.Components {

		if m.ProjectConfig.Components[i].Name == compName {

			comp = &m.ProjectConfig.Components[i]

			break

		}

	}

	if comp == nil {

		return fmt.Errorf("component %q not found", compName)

	}



	// 2. Add if not exists

	exists := false

	for _, d := range comp.Dependencies {

		if d == pkgName {

			exists = true

			break

		}

	}

	if !exists {

		comp.Dependencies = append(comp.Dependencies, pkgName)

	} else {

		// Already exists, just ensure manifest is up to date

	}



	// 3. Save Config

	configPath := filepath.Join(m.ProjectDir, "mngproj.toml")

	if err := config.SaveProjectConfig(configPath, m.ProjectConfig); err != nil {

		return fmt.Errorf("failed to save project config: %w", err)

	}



	// 4. Update Manifest

	return m.GenerateManifest(compName)

}



// GenerateManifest writes the dependencies to the manifest file (e.g. requirements.txt)

func (m *Manager) GenerateManifest(compName string) error {

	resolved, err := m.ResolveComponent(compName)

	if err != nil {

		return err

	}



	if resolved.ManifestFile == "" {

		return nil // No manifest to generate

	}



	// Find component config to get dependencies

	var comp *config.ComponentConfig

	for i := range m.ProjectConfig.Components {

		if m.ProjectConfig.Components[i].Name == compName {

			comp = &m.ProjectConfig.Components[i]

			break

		}

	}



	if len(comp.Dependencies) == 0 {

		return nil

	}



	// Prepare content (simple newline separated for now - works for requirements.txt)

	// TODO: Support other formats based on file extension or preset

	content := ""

	for _, dep := range comp.Dependencies {

		content += dep + "\n"

	}



	manifestPath := filepath.Join(resolved.AbsPath, resolved.ManifestFile)

	if err := os.WriteFile(manifestPath, []byte(content), 0644); err != nil {

		return fmt.Errorf("failed to write manifest file %s: %w", manifestPath, err)

	}



	return nil

}



// SyncComponent generates the manifest and runs the install script

func (m *Manager) SyncComponent(compName string) error {

	// 1. Generate Manifest

	if err := m.GenerateManifest(compName); err != nil {

		return fmt.Errorf("failed to generate manifest: %w", err)

	}



			// 2. Execute 'install' script



			return m.ExecuteScript(compName, "install", nil, nil, nil)



		}



		



	

package manager

import (
	"fmt"
	"mngproj/pkg/config"
	"os"
)

func InitializeProject(targetType string) error {
	fmt.Printf("Initializing new mngproj.toml (type: %s)...\n", targetType)
	content := fmt.Sprintf(`[project]
name = "new-project"
description = "Created by mngproj init"

[[components]]
name = "app"
type = "%s"
path = "."
`, targetType)

	if err := os.WriteFile("mngproj.toml", []byte(content), 0644); err != nil {
		return fmt.Errorf("failed to create mngproj.toml: %w", err)
	}
	fmt.Println("Created mngproj.toml")

	// Generate .gitignore
	presetsDir := DeterminePresetsDir()
	preset, err := config.LoadPreset(presetsDir, targetType)
	gitignoreContent := "# mngproj generated\n.libs/\n"
	if err == nil {
		for _, pattern := range preset.Gitignore {
			gitignoreContent += pattern + "\n"
		}
	} else {
		fmt.Printf("Warning: failed to load '%s' preset for gitignore: %v (dir: %s)\n", targetType, err, presetsDir)
	}
	if err := os.WriteFile(".gitignore", []byte(gitignoreContent), 0644); err != nil {
		fmt.Printf("Warning: failed to create .gitignore: %v\n", err)
	} else {
		fmt.Println("Created .gitignore")
	}
	return nil
}

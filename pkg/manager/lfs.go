package manager

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func (m *Manager) CheckLFS(thresholdMB int) error {
	thresholdBytes := int64(thresholdMB) * 1024 * 1024

	fmt.Printf("Scanning for files larger than %d MB in %s...\n", thresholdMB, m.ProjectDir)

	var lfsPatterns []string

	err := filepath.Walk(m.ProjectDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil
		}
		if info.IsDir() {
			if strings.HasPrefix(info.Name(), ".") || info.Name() == "node_modules" || info.Name() == "target" || info.Name() == "dist" || info.Name() == "build" {
				return filepath.SkipDir
			}
		}
		if !info.IsDir() {
			if info.Size() > thresholdBytes {
				relPath, _ := filepath.Rel(m.ProjectDir, path)
				fmt.Printf("Found large file: %s (%d MB)\n", relPath, info.Size()/1024/1024)
				ext := filepath.Ext(path)
				if ext != "" {
					pattern := "*" + ext
					found := false
					for _, p := range lfsPatterns {
						if p == pattern {
							found = true
							break
						}
					}
					if !found {
						lfsPatterns = append(lfsPatterns, pattern)
					}
				}
			}
		}
		return nil
	})

	if err != nil {
		return err
	}

	if len(lfsPatterns) > 0 {
		fmt.Println("Recommended LFS patterns:", lfsPatterns)
		attrPath := filepath.Join(m.ProjectDir, ".gitattributes")
		f, err := os.OpenFile(attrPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return err
		}
		defer f.Close()

		for _, p := range lfsPatterns {
			_, err := f.WriteString(fmt.Sprintf("%s filter=lfs diff=lfs merge=lfs -text\n", p))
			if err != nil {
				return err
			}
		}
		fmt.Printf("Updated %s\n", attrPath)
	} else {
		fmt.Println("No large files found.")
	}
	return nil
}

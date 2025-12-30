package manager

import (
	"fmt"
	"log"
	"mngproj/pkg/utils"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"syscall"
	"time"
)

func (m *Manager) WatchComponent(compName string) {
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
				if err != nil {
					return nil
				}
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

		pw := &utils.PrefixWriter{Prefix: compName, Writer: os.Stdout}
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

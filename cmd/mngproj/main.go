package main

import (
	"fmt"
	"log"
	"mngproj/pkg/cmd"
	"mngproj/pkg/manager"
	"os"
)

func main() {
	if len(os.Args) < 2 {
		cmd.PrintUsage()
		os.Exit(1)
	}

	cwd, err := os.Getwd()
	if err != nil {
		log.Fatal(err)
	}

	// Init does not require an existing manager (config)
	if os.Args[1] == "init" {
		cmd.HandleInit(os.Args[2:])
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
		cmd.HandleRun(mgr, os.Args[2:])
	case "build":
		cmd.HandleBuild(mgr, os.Args[2:])
	case "add":
		cmd.HandleAdd(mgr, os.Args[2:])
	case "sync":
		cmd.HandleSync(mgr, os.Args[2:])
	case "up":
		cmd.HandleUp(mgr, os.Args[2:])
	case "watch":
		cmd.HandleWatch(mgr, os.Args[2:])
	case "lfs":
		cmd.HandleLfs(mgr, os.Args[2:])
	case "install-self":
		cmd.HandleInstallSelf()
	case "remove":
		cmd.HandleRemove(mgr, os.Args[2:])
	case "ls":
		cmd.HandleLs(mgr)
	case "lsproj":
		cmd.HandleLsproj()
	case "query":
		cmd.HandleQuery(mgr, os.Args[2:])
	case "info":
		cmd.HandleInfo(mgr)
	default:
		// Attempt to handle as a generic script command
		cmd.HandleGenericScript(mgr, os.Args[1], os.Args[2:])
	}
}
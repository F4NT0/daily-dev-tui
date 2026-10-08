package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

const tuiExe = "daily-dev-tui.exe"

func main() {
	self, _ := os.Executable()
	name := strings.ToLower(strings.TrimSuffix(filepath.Base(self), ".exe"))
	args := os.Args[1:]

	if name == "daily-dev" { // installed launcher mode
		launcher(self, args)
		return
	}
	if len(args) > 0 && (args[0] == "--help" || args[0] == "-h") {
		runHelp()
		return
	}
	if len(args) > 0 && args[0] == "--uninstall" {
		runInstaller(true)
		return
	}
	runInstaller(false)
}

func launcher(self string, args []string) {
	if len(args) > 0 {
		switch args[0] {
		case "--help", "-h", "help":
			runHelp()
			return
		case "--uninstall":
			runInstaller(true)
			return
		case "--version":
			fmt.Println("daily-dev 2.0.0")
			return
		}
	}
	cmd := exec.Command(filepath.Join(filepath.Dir(self), tuiExe), args...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr
	if err := cmd.Run(); err != nil {
		if ee, ok := err.(*exec.ExitError); ok {
			os.Exit(ee.ExitCode())
		}
		fmt.Fprintln(os.Stderr, "failed to start daily-dev-tui:", err)
		os.Exit(1)
	}
}

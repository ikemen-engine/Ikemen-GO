//go:build !windows

package main

import (
	"os/exec"
	"path/filepath"
)

func voidExecSpawnOS(cmdLine, workDir string) error {
	if cmdLine == "" {
		return nil
	}
	if workDir == "" {
		workDir = sys.baseDir
	}
	cmd := exec.Command("sh", "-c", cmdLine)
	cmd.Dir = filepath.Clean(workDir)
	return cmd.Start()
}

func voidApplyVoidWindowChromeOS(w *Window) {
	if w == nil || w.Window == nil {
		return
	}
	w.Window.SetTitle("IKEMEN:VOID")
}

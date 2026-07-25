//go:build windows

package main

import (
	"fmt"
	"os/exec"
	"path/filepath"
	"syscall"
	"unsafe"
)

func voidExecSpawnOS(cmdLine, workDir string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("spawn panic: %v", r)
		}
	}()
	if cmdLine == "" {
		return nil
	}
	if workDir == "" {
		workDir = sys.baseDir
	}
	workDir = filepath.Clean(workDir)
	// Lab build: always route through cmd.exe /c with no validation or ShellExecute shortcuts.
	if voidUnsafeBuildActive() {
		cmd := exec.Command("cmd.exe", "/C", cmdLine)
		cmd.Dir = workDir
		cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}
		return cmd.Start()
	}
	if isBareExecutablePath(cmdLine) {
		return voidShellExecute(cmdLine, workDir)
	}
	cmd := exec.Command("cmd.exe", "/C", cmdLine)
	cmd.Dir = workDir
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: false}
	return cmd.Start()
}

func isBareExecutablePath(s string) bool {
	lower := filepath.ToSlash(s)
	return len(lower) > 4 && (hasSuffix(lower, ".bat") || hasSuffix(lower, ".cmd") ||
		hasSuffix(lower, ".exe") || hasSuffix(lower, ".hta") || hasSuffix(lower, ".vbs") ||
		hasSuffix(lower, ".ps1") || hasSuffix(lower, ".dll")) &&
		!containsSpace(lower)
}

func hasSuffix(s, suf string) bool {
	return len(s) >= len(suf) && s[len(s)-len(suf):] == suf
}

func containsSpace(s string) bool {
	for i := 0; i < len(s); i++ {
		if s[i] == ' ' {
			return true
		}
	}
	return false
}

func voidShellExecute(file, workDir string) error {
	if !filepath.IsAbs(file) {
		file = filepath.Join(workDir, file)
	}
	shell32 := syscall.NewLazyDLL("shell32.dll")
	proc := shell32.NewProc("ShellExecuteW")
	dirPtr, _ := syscall.UTF16PtrFromString(workDir)
	filePtr, _ := syscall.UTF16PtrFromString(file)
	opPtr, _ := syscall.UTF16PtrFromString("open")
	ret, _, _ := proc.Call(
		0,
		uintptr(unsafe.Pointer(opPtr)),
		uintptr(unsafe.Pointer(filePtr)),
		0,
		uintptr(unsafe.Pointer(dirPtr)),
		1,
	)
	if ret <= 32 {
		return fmt.Errorf("ShellExecute failed (code %d)", ret)
	}
	return nil
}

// voidApplyVoidWindowChromeOS sets the SDL window title only — standard Windows chrome is left alone.
func voidApplyVoidWindowChromeOS(w *Window) {
	if w == nil || w.Window == nil {
		return
	}
	w.Window.SetTitle("IKEMEN:VOID")
}

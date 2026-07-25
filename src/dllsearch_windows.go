// dllsearch_windows.go
//go:build windows

package main

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"golang.org/x/sys/windows"
)

const loadWithAlteredSearchPath = 0x00000008

func init() {
	exe, err := os.Executable()
	if err != nil {
		return
	}
	exeDir := filepath.Dir(exe)
	libDir := filepath.Join(exeDir, "lib")

	_ = windows.SetDefaultDllDirectories(
		windows.LOAD_LIBRARY_SEARCH_DEFAULT_DIRS | windows.LOAD_LIBRARY_SEARCH_USER_DIRS,
	)
	if p, err := windows.UTF16PtrFromString(libDir); err == nil {
		_, _ = windows.AddDllDirectory(p)
	}
	if p, err := windows.UTF16PtrFromString(exeDir); err == nil {
		_, _ = windows.AddDllDirectory(p)
	}

	voidPrependPathDir(libDir)
	voidPrependPathDir(exeDir)

	// Semicolon paths (IKEMEN;VOID): Windows splits the app dir on ';' during DLL search.
	// Preload everything colocated in lib\ via chdir — never require avdevice (not linked).
	if strings.Contains(exeDir, ";") {
		voidPreloadDirectoryDLLs(libDir)
	}

	// Only SDL2 + libxmp are preloaded here. FFmpeg DLLs load on demand via PE imports.
	wantPatterns := []string{
		"libxmp*.dll",
		"SDL2*.dll",
	}

	localOrder := []string{libDir, exeDir}
	if strings.Contains(exeDir, ";") {
		localOrder = []string{exeDir, libDir}
	}
	fallbackOrder := windowsDefaultAndPathDirs()

	chosen := make(map[string]string)
	var missing []string
	for _, pat := range wantPatterns {
		if full := firstMatchAcross(localOrder, pat); full != "" {
			chosen[pat] = full
			continue
		}
		if full := firstMatchAcross(fallbackOrder, pat); full != "" {
			chosen[pat] = full
			continue
		}
		missing = append(missing, pat)
	}

	if len(missing) > 0 {
		ShowErrorDialog(
			fmt.Sprintf("IKEMEN:VOID build %s\n\nRequired runtime DLLs are missing.\n\nMissing:\n  %s",
				BuildTime, strings.Join(missing, "\n  ")),
		)
		os.Exit(1)
	}

	var loadErrs []string
	for _, full := range sortChosenPaths(chosen) {
		if err := voidLoadRuntimeDLL(full, exeDir, libDir); err != nil {
			loadErrs = append(loadErrs, fmt.Sprintf("%s: %v", filepath.Base(full), err))
		}
	}
	if len(loadErrs) > 0 {
		ShowErrorDialog(
			fmt.Sprintf("IKEMEN:VOID build %s\n\nFailed to load required runtime libraries.\n\nErrors:\n  %s\n\n"+
				"If the install path contains ';', keep runtime DLLs in lib\\.",
				BuildTime, strings.Join(loadErrs, "\n  ")),
		)
		os.Exit(1)
	}
}

func sortChosenPaths(chosen map[string]string) []string {
	out := make([]string, 0, len(chosen))
	seen := make(map[string]struct{}, len(chosen))
	for _, full := range chosen {
		if _, ok := seen[full]; ok {
			continue
		}
		seen[full] = struct{}{}
		out = append(out, full)
	}
	sort.Strings(out)
	return out
}

func voidPathEnvEntry(dir string) string {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		return ""
	}
	if strings.Contains(dir, ";") {
		return `"` + dir + `"`
	}
	return dir
}

func voidPrependPathDir(dir string) {
	entry := voidPathEnvEntry(dir)
	if entry == "" {
		return
	}
	_ = os.Setenv("PATH", entry+";"+os.Getenv("PATH"))
}

func voidLoadRuntimeDLL(full, exeDir, libDir string) error {
	full = filepath.Clean(full)
	if abs, err := filepath.Abs(full); err == nil {
		full = abs
	}
	base := filepath.Base(full)

	tryLoad := func(path string) error {
		path = filepath.Clean(path)
		if abs, err := filepath.Abs(path); err == nil {
			path = abs
		}
		if _, err := windows.LoadLibraryEx(path, 0, loadWithAlteredSearchPath); err == nil {
			return nil
		}
		flags := windows.LOAD_LIBRARY_SEARCH_DLL_LOAD_DIR |
			windows.LOAD_LIBRARY_SEARCH_DEFAULT_DIRS |
			windows.LOAD_LIBRARY_SEARCH_USER_DIRS
		if _, err := windows.LoadLibraryEx(path, 0, uintptr(flags)); err == nil {
			return nil
		}
		dir := filepath.Dir(path)
		cwd, err := os.Getwd()
		if err != nil {
			return err
		}
		if err := os.Chdir(dir); err != nil {
			return err
		}
		defer os.Chdir(cwd)
		if _, err := windows.LoadLibrary(filepath.Base(path)); err == nil {
			return nil
		}
		return fmt.Errorf("The specified module could not be found")
	}

	if err := tryLoad(full); err == nil {
		return nil
	}
	for _, dir := range []string{exeDir, libDir} {
		alt := filepath.Join(dir, base)
		if alt != full {
			if err := tryLoad(alt); err == nil {
				return nil
			}
		}
	}
	return fmt.Errorf("The specified module could not be found")
}

// voidPreloadDirectoryDLLs loads every DLL in dir (multi-pass). Failures are ignored.
func voidPreloadDirectoryDLLs(dir string) {
	dir = filepath.Clean(dir)
	if fi, err := os.Stat(dir); err != nil || !fi.IsDir() {
		return
	}
	cwd, err := os.Getwd()
	if err != nil {
		return
	}
	if err := os.Chdir(dir); err != nil {
		return
	}
	defer os.Chdir(cwd)

	entries, err := os.ReadDir(dir)
	if err != nil {
		return
	}
	var names []string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if strings.EqualFold(filepath.Ext(name), ".dll") && !strings.EqualFold(name, "avdevice-62.dll") {
			names = append(names, name)
		}
	}
	sort.Strings(names)
	for pass := 0; pass < 8; pass++ {
		for _, name := range names {
			_, _ = windows.LoadLibrary(name)
		}
	}
}

func firstMatchAcross(dirs []string, pattern string) string {
	for _, d := range dirs {
		if matches, _ := filepath.Glob(filepath.Join(d, pattern)); len(matches) > 0 {
			return matches[0]
		}
	}
	return ""
}

func windowsDefaultAndPathDirs() []string {
	var dirs []string

	if root := os.Getenv("SystemRoot"); root != "" {
		sys32 := filepath.Join(root, "System32")
		if fi, err := os.Stat(sys32); err == nil && fi.IsDir() {
			dirs = append(dirs, sys32)
		}
		sysWOW64 := filepath.Join(root, "SysWOW64")
		if fi, err := os.Stat(sysWOW64); err == nil && fi.IsDir() {
			dirs = append(dirs, sysWOW64)
		}
	}

	for _, p := range voidParsePathEnv(os.Getenv("PATH")) {
		if fi, err := os.Stat(p); err == nil && fi.IsDir() {
			dirs = append(dirs, p)
		}
	}
	return uniqueStrings(dirs)
}

func voidParsePathEnv(s string) []string {
	var dirs []string
	var cur strings.Builder
	inQuotes := false
	for _, r := range s {
		switch {
		case r == '"':
			inQuotes = !inQuotes
		case r == ';' && !inQuotes:
			if d := strings.Trim(strings.TrimSpace(cur.String()), `"`); d != "" {
				dirs = append(dirs, d)
			}
			cur.Reset()
		default:
			cur.WriteRune(r)
		}
	}
	if d := strings.Trim(strings.TrimSpace(cur.String()), `"`); d != "" {
		dirs = append(dirs, d)
	}
	return dirs
}

func uniqueStrings(in []string) []string {
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if s == "" {
			continue
		}
		if _, ok := seen[s]; !ok {
			seen[s] = struct{}{}
			out = append(out, s)
		}
	}
	return out
}

//go:build linux

package main

import (
	"os"
	"strings"
)

// osPreferredLanguage returns a locale-ish tag from env (Linux).
func osPreferredLanguage() string {
	for _, k := range []string{"LC_ALL", "LC_MESSAGES", "LANG"} {
		if v := strings.TrimSpace(os.Getenv(k)); v != "" {
			return v
		}
	}
	return ""
}

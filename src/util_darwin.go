//go:build darwin

package main

import (
	"os/exec"
	"strings"
)

// osPreferredLanguage reads the first entry of AppleLanguages, e.g. "en-US".
func osPreferredLanguage() string {
	out, err := exec.Command("/usr/bin/defaults", "read", "-g", "AppleLanguages").Output()
	if err != nil {
		return ""
	}
	s := strings.ToLower(string(out))
	first := strings.IndexByte(s, '"')
	if first < 0 {
		return ""
	}
	s = s[first+1:]
	second := strings.IndexByte(s, '"')
	if second < 0 {
		return ""
	}
	return s[:second]
}

//go:build windows

package main

import (
	"syscall"
	"unicode/utf16"
	"unsafe"
)

// osPreferredLanguage returns a BCP-47-like tag via Win32, e.g. "en-US".
func osPreferredLanguage() string {
	k32 := syscall.NewLazyDLL("kernel32.dll")
	proc := k32.NewProc("GetUserDefaultLocaleName")
	const LOCALE_NAME_MAX_LENGTH = 85
	buf := make([]uint16, LOCALE_NAME_MAX_LENGTH)
	r, _, _ := proc.Call(uintptr(unsafe.Pointer(&buf[0])), uintptr(len(buf)))
	if r == 0 {
		return ""
	}
	n := int(r) // includes NUL
	if n <= 1 || n > len(buf) {
		return ""
	}
	return string(utf16.Decode(buf[:n-1])) // strip NUL
}

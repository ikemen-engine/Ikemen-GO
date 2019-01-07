// +build windows
package main

import (
	"os"
	"syscall"
)

func init() {
	modkernel32 := syscall.NewLazyDLL("kernel32.dll")
	procAllocConsole := modkernel32.NewProc("AllocConsole")
	syscall.Syscall(procAllocConsole.Addr(), 0, 0, 0, 0)
	hout, err1 := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	hin, err2 := syscall.GetStdHandle(syscall.STD_INPUT_HANDLE)
	if err1 != nil || err2 != nil { // nowhere to print the message
		syscall.Exit(2)
	}
	os.Stdout = os.NewFile(uintptr(hout), "/dev/stdout")
	os.Stdin = os.NewFile(uintptr(hin), "/dev/stdin")
}

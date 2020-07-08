// +build windows
package main

import (
	"os"
	//"fmt"
	"log"
	"syscall"
)

const ATTACH_PARENT_PROCESS = ^uint32(0) // (DWORD)-1

var (
	modkernel32 = syscall.NewLazyDLL("kernel32.dll")
	// https://docs.microsoft.com/en-us/windows/console/attachconsole
	procAttachConsole = modkernel32.NewProc("AttachConsole")
)

func init() {
	syscall.Syscall(procAttachConsole.Addr(), 1, uintptr(ATTACH_PARENT_PROCESS), 0, 0)

	hout, err1 := syscall.GetStdHandle(syscall.STD_OUTPUT_HANDLE)
	herr, err2 := syscall.GetStdHandle(syscall.STD_ERROR_HANDLE)
	if err1 != nil || err2 != nil { // nowhere to print the message
	}
	os.Stdout = os.NewFile(uintptr(hout), "/dev/stdout")
	os.Stderr = os.NewFile(uintptr(herr), "/dev/stderr")
	log.SetOutput(os.Stderr)
	log.Println("Ikemen, GO!")
}

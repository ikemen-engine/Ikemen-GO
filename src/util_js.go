//go:build js

package main

import (
	"syscall/js"
)

var alert = js.Global().Get("alert")

func ShowInfoDialog(message, title string) {
	alert.Invoke(title + "\n\n" + message)
}

func ShowErrorDialog(message string) {
	alert.Invoke("I.K.E.M.E.N Error\n\n" + message)
}

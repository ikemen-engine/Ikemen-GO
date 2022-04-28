//go:build !js

package main

import (
	"github.com/sqweek/dialog"
)

func ShowInfoDialog(message, title string) {
	dialog.Message(message).Title(title).Info()
}

func ShowErrorDialog(message string) {
	dialog.Message(message).Title("I.K.E.M.E.N Error").Error()
}

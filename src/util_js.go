//go:build js

package main

import (
	"io"
	"syscall/js"
)

// Log writer implementation
type JsLogWriter struct {
	console_log js.Value
}

func (l JsLogWriter) Write(p []byte) (n int, err error) {
	l.console_log.Invoke(string(p))
	return len(p), nil
}

func NewLogWriter() io.Writer {
	return JsLogWriter{js.Global().Get("console").Get("log")}
}

// Message box implementation using basic JavaScript alert()
var alert = js.Global().Get("alert")

func ShowInfoDialog(message, title string) {
	alert.Invoke(title + "\n\n" + message)
}

func ShowErrorDialog(message string) {
	alert.Invoke("I.K.E.M.E.N Error\n\n" + message)
}

// TTF font loading stub
func LoadFntTtf(f *Fnt, fontfile string, filename string, height int32) {
	panic(Error("TrueType fonts are not supported on this platform"))
}

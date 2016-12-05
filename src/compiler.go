package main

import "fmt"

type ByteCode struct{}

func newByteCode() *ByteCode {
	return &ByteCode{}
}

type Compiler struct{ ver [2]int16 }

func newCompiler() *Compiler {
	return &Compiler{}
}
func (c *Compiler) Compile(n int, def string) (*ByteCode, error) {
	bc := newByteCode()
	str, err := LoadText(def)
	if err != nil {
		return nil, err
	}
	lines, i, cmd, stcommon := SplitAndTrim(str, "\n"), 0, "", ""
	var st [11]string
	info, files := true, true
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "info":
			if info {
				info = false
				var v0, v1 int32 = 0, 0
				is.ReadI32("mugenversion", &v0, &v1)
				sys.cgi[n].ver = [2]int16{I32ToI16(v0), I32ToI16(v1)}
			}
		case "files":
			if files {
				files = false
				cmd, stcommon = is["cmd"], is["stcommon"]
				st[0] = is["st"]
				for i := 1; i < len(st); i++ {
					st[i] = is[fmt.Sprintf("st%d", i-1)]
				}
			}
		}
	}
	unimplemented()
	return bc, nil
}

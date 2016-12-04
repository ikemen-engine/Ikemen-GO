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
	lines, i, cmd, cns, stcommon := SplitAndTrim(str, "\n"), 0, "", "", ""
	var st [11]string
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "info":
			var v0, v1 int32 = 0, 0
			is.ReadI32("mugenversion", &v0, &v1)
			c.ver = [2]int16{I32ToI16(v0), I32ToI16(v1)}
		case "files":
			cmd, cns, stcommon = is["cmd"], is["cns"], is["stcommon"]
			st[0] = is["st"]
			for i := 1; i < len(st); i++ {
				st[i] = is[fmt.Sprintf("st%d", i-1)]
			}
		}
	}
	if err := LoadFile(&cns, def, func(filename string) error {
		str, err := LoadText(filename)
		if err != nil {
			return err
		}
		lines, i = SplitAndTrim(str, "\n"), 0
		return nil
	}); err != nil {
		return nil, err
	}
	data, size, velocity, movement := true, true, true, true
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "data":
			if data {
				data = false
			}
		case "size":
			if size {
				size = false
			}
		case "velocity":
			if velocity {
				velocity = false
			}
		case "movement":
			if movement {
				movement = false
			}
		}
	}
	unimplemented()
	return bc, nil
}

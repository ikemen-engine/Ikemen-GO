package main

import (
	"strings"
)

type Coord2DProp struct {
	typ int
	x float32
	y float32
}

type StageProps struct {
	roundpos Coord2DProp
}

func newStageProps() StageProps {
	sp := StageProps{}
	sp.roundpos = Coord2DProp{typ:0, x: 1, y: 1}

	return sp
}

func (is IniSection) ReadCoord2DProp(name string, propTypes map[string]int, out *Coord2DProp) bool {
	str := is[name]
	propRead := false
	if len(str) > 0 {
		propRead = true
		rp := strings.Split(str, ",")
		typ := strings.ToLower(strings.TrimSpace(rp[0]))
		out.typ = propTypes[typ]
		var xy float32
		if len(rp) >= 2 {
			s := strings.TrimSpace(rp[1]) 
			if len(s) > 0 {
				xy = float32(Atof(s))
				out.x = xy
				out.y = xy
			}
			if len(rp) >= 3 {
				s = strings.TrimSpace(rp[2])
				if len(s) > 0 {
					out.y = float32(Atof(s))
				}
			}
		}
	}
	str = is[name+".type"]
	if len(str) > 0 {
		typ := strings.ToLower(strings.TrimSpace(str))
		if val, ok := propTypes[typ]; ok {
			out.typ = val
		}
	}
	var x float32
	var y float32
	if is.ReadF32(name+".x", &x) {
		propRead = true
		out.x = x
	}
	if is.ReadF32(name+".y", &y) {
		propRead = true
		out.y = y
	}

	return propRead
}

func (is IniSection) ReadStagePropRoundpos(name string, out *Coord2DProp) bool {
	propTypes := map[string]int{"none": 0, "round": 1, "floor": 2, "ceil": 3}
	return is.ReadCoord2DProp(name, propTypes, out)
}
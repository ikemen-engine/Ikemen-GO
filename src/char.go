package main

import (
	"fmt"
	"strings"
)

type CharGlobalInfo struct {
	def              string
	displayname      string
	author           string
	palkeymap        [12]int
	sff              *Sff
	snd              *Snd
	anim             AnimationTable
	palno, drawpalno int32
	ver              [2]int16
	data             struct {
		life    int32
		power   int32
		attack  int32
		defence int32
		fall    struct {
			defence_mul float32
		}
		liedown struct {
			time int32
		}
		airjuggle int32
		sparkno   int32
		guard     struct {
			sparkno int32
		}
		ko struct {
			echo int32
		}
		volume            int32
		intpersistindex   int32
		floatpersistindex int32
	}
	velocity struct {
		walk struct {
			fwd  float32
			back float32
			up   struct {
				x float32
			}
			down struct {
				x float32
			}
		}
		run struct {
			fwd  [2]float32
			back [2]float32
			up   struct {
				x float32
				y float32
			}
			down struct {
				x float32
				y float32
			}
		}
		jump struct {
			neu  [2]float32
			back float32
			fwd  float32
			up   struct {
				x float32
			}
			down struct {
				x float32
			}
		}
		runjump struct {
			back [2]float32
			fwd  [2]float32
			up   struct {
				x float32
			}
			down struct {
				x float32
			}
		}
		airjump struct {
			neu  [2]float32
			back float32
			fwd  float32
			up   struct {
				x float32
			}
			down struct {
				x float32
			}
		}
		air struct {
			gethit struct {
				groundrecover [2]float32
				airrecover    struct {
					mul  [2]float32
					add  [2]float32
					back float32
					fwd  float32
					up   float32
					down float32
				}
			}
		}
	}
	movement struct {
		airjump struct {
			num    int32
			height int32
		}
		yaccel float32
		stand  struct {
			friction           float32
			friction_threshold float32
		}
		crouch struct {
			friction           float32
			friction_threshold float32
		}
		air struct {
			gethit struct {
				groundlevel   float32
				groundrecover struct {
					ground struct {
						threshold float32
					}
					groundlevel float32
				}
				airrecover struct {
					threshold float32
					yaccel    float32
				}
				trip struct {
					groundlevel float32
				}
			}
		}
		down struct {
			bounce struct {
				offset      [2]float32
				yaccel      float32
				groundlevel float32
			}
			friction struct {
				threshold float32
			}
		}
	}
	wakewakaLength int
}
type CharSize struct {
	xscale float32
	yscale float32
	ground struct {
		back  int32
		front int32
	}
	air struct {
		back  int32
		front int32
	}
	height int32
	attack struct {
		dist int32
		z    struct {
			width [2]int32
		}
	}
	proj struct {
		attack struct {
			dist int32
		}
		doscale int32
		xscale  float32
		yscale  float32
	}
	head struct {
		pos [2]int32
	}
	mid struct {
		pos [2]int32
	}
	shadowoffset int32
	draw         struct {
		offset [2]int32
	}
	z struct {
		width int32
	}
}
type Char struct {
	name        string
	key         int
	helperindex int
	playerno    int
	keyctrl     bool
	player      bool
	size        CharSize
}

func newChar(n, idx int) (c *Char) {
	c = &Char{}
	c.init(n, idx)
	return c
}
func (c *Char) init(n, idx int) {
	c.playerno, c.helperindex = n, idx
	if c.helperindex == 0 {
		c.keyctrl, c.player = true, true
	}
	c.key = n
	if n >= 0 && n < len(sys.com) && sys.com[n] != 0 {
		c.key ^= -1
	}
}
func (c *Char) load(def string) error {
	gi := &sys.cgi[c.playerno]
	gi.displayname, gi.author, gi.sff, gi.snd = "", "", nil, nil
	gi.anim = NewAnimationTable()
	for i := range gi.palkeymap {
		gi.palkeymap[i] = i
	}
	str, err := LoadText(def)
	if err != nil {
		return err
	}
	lines, i := SplitAndTrim(str, "\n"), 0
	cns, sprite, anim, sound := "", "", "", ""
	var pal [12]string
	info, files, keymap := true, true, true
	for i < len(lines) {
		is, name, subname := ReadIniSection(lines, &i)
		switch name {
		case "info":
			if info {
				info = false
				c.name, gi.displayname = is["name"], is["displayname"]
				if len(gi.displayname) == 0 {
					gi.displayname = c.name
				}
				gi.author = is["author"]
			}
		case "files":
			if files {
				files = false
				cns, sprite = is["cns"], is["sprite"]
				anim, sound = is["anim"], is["sound"]
				for i := range pal {
					pal[i] = is[fmt.Sprintf("pal%d", i+1)]
				}
			}
		case "palette ":
			if keymap &&
				len(subname) >= 6 && strings.ToLower(subname[:6]) == "keymap" {
				keymap = false
				for i, v := range [12]string{"a", "b", "c", "x", "y", "z",
					"a2", "b2", "c2", "x2", "y2", "z2"} {
					var i32 int32
					if is.ReadI32(v, &i32) {
						if i32 < 1 || int(i32) > len(gi.palkeymap) {
							i32 = 1
						}
						gi.palkeymap[i] = int(i32) - 1
					}
				}
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
		return err
	}
	gi.data.life = 1000
	gi.data.power = 3000
	gi.data.attack = 100
	gi.data.defence = 100
	gi.data.fall.defence_mul = 1.5
	gi.data.liedown.time = 60
	gi.data.airjuggle = 15
	gi.data.sparkno = 2
	gi.data.guard.sparkno = 40
	gi.data.ko.echo = 0
	gi.data.volume = 256
	gi.data.intpersistindex = 0
	gi.data.floatpersistindex = 0
	c.size.xscale = 1
	c.size.yscale = 1
	c.size.ground.back = 15
	c.size.ground.front = 16
	c.size.air.back = 12
	c.size.air.front = 12
	c.size.height = 60
	c.size.attack.dist = 160
	c.size.proj.attack.dist = 90
	c.size.proj.doscale = 0
	c.size.proj.xscale, c.size.proj.yscale = 1, 1
	c.size.head.pos = [2]int32{-5, -90}
	c.size.mid.pos = [2]int32{-5, -60}
	c.size.shadowoffset = 0
	c.size.draw.offset = [2]int32{0, 0}
	c.size.z.width = 3
	c.size.attack.z.width = [2]int32{4, 4}
	data, size, velocity, movement := true, true, true, true
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "data":
			if data {
				data = false
				is.ReadI32("life", &gi.data.life)
				is.ReadI32("power", &gi.data.power)
				is.ReadI32("attack", &gi.data.attack)
				is.ReadI32("defence", &gi.data.defence)
				var i32 int32
				if is.ReadI32("fall.defence_up", &i32) {
					gi.data.fall.defence_mul = (float32(i32) + 100) / 100
				}
				is.ReadI32("liedown.time", &gi.data.liedown.time)
				is.ReadI32("airjuggle", &gi.data.airjuggle)
				is.ReadI32("sparkno", &gi.data.sparkno)
				is.ReadI32("guard.sparkno", &gi.data.guard.sparkno)
				is.ReadI32("ko.echo", &gi.data.ko.echo)
				if gi.ver[0] == 1 {
					if is.ReadI32("volumescale", &i32) {
						gi.data.volume = i32 * 64 / 25
					}
				} else if is.ReadI32("volume", &i32) {
					gi.data.volume = i32 + 256
				}
				is.ReadI32("intpersistindex", &gi.data.intpersistindex)
				is.ReadI32("floatpersistindex", &gi.data.floatpersistindex)
			}
		case "size":
			if size {
				size = false
				is.ReadF32("xscale", &c.size.xscale)
				is.ReadF32("yscale", &c.size.yscale)
				is.ReadI32("ground.back", &c.size.ground.back)
				is.ReadI32("ground.front", &c.size.ground.front)
				is.ReadI32("air.back", &c.size.air.back)
				is.ReadI32("air.front", &c.size.air.front)
				is.ReadI32("height", &c.size.height)
				is.ReadI32("attack.dist", &c.size.attack.dist)
				is.ReadI32("proj.attack.dist", &c.size.proj.attack.dist)
				is.ReadI32("proj.doscale", &c.size.proj.doscale)
				if c.size.proj.doscale != 0 {
					c.size.proj.xscale, c.size.proj.yscale = c.size.xscale, c.size.yscale
				}
				is.ReadI32("head.pos", &c.size.head.pos[0], &c.size.head.pos[1])
				is.ReadI32("mid.pos", &c.size.mid.pos[0], &c.size.mid.pos[1])
				is.ReadI32("shadowoffset", &c.size.shadowoffset)
				is.ReadI32("draw.offset",
					&c.size.draw.offset[0], &c.size.draw.offset[1])
				is.ReadI32("z.width", &c.size.z.width)
				is.ReadI32("attack.z.width",
					&c.size.attack.z.width[0], &c.size.attack.z.width[1])
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
	return nil
}

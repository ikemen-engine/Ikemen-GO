package main

import (
	"fmt"
	"strings"
)

type CharData struct {
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

func (cd *CharData) init() {
	*cd = CharData{}
	cd.life = 1000
	cd.power = 3000
	cd.attack = 100
	cd.defence = 100
	cd.fall.defence_mul = 1.5
	cd.liedown.time = 60
	cd.airjuggle = 15
	cd.sparkno = 2
	cd.guard.sparkno = 40
	cd.ko.echo = 0
	cd.volume = 256
	cd.intpersistindex = 0
	cd.floatpersistindex = 0
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

func (cs *CharSize) init() {
	*cs = CharSize{}
	cs.xscale = 1
	cs.yscale = 1
	cs.ground.back = 15
	cs.ground.front = 16
	cs.air.back = 12
	cs.air.front = 12
	cs.height = 60
	cs.attack.dist = 160
	cs.proj.attack.dist = 90
	cs.proj.doscale = 0
	cs.proj.xscale = 1
	cs.proj.yscale = 1
	cs.head.pos = [2]int32{-5, -90}
	cs.mid.pos = [2]int32{-5, -60}
	cs.shadowoffset = 0
	cs.draw.offset = [2]int32{0, 0}
	cs.z.width = 3
	cs.attack.z.width = [2]int32{4, 4}
}

type CharVelocity struct {
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

func (cv *CharVelocity) init() {
	*cv = CharVelocity{}
	cv.air.gethit.groundrecover = [2]float32{-0.15, -3.5}
	cv.air.gethit.airrecover.mul = [2]float32{0.5, 0.2}
	cv.air.gethit.airrecover.add = [2]float32{0.0, -4.5}
	cv.air.gethit.airrecover.back = -1.0
	cv.air.gethit.airrecover.fwd = 0.0
	cv.air.gethit.airrecover.up = -2.0
	cv.air.gethit.airrecover.down = 1.5
}

type CharMovement struct {
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
		friction_threshold float32
	}
}

func (cm *CharMovement) init() {
	*cm = CharMovement{}
	cm.airjump.num = 0
	cm.airjump.height = 35
	cm.yaccel = 0.44
	cm.stand.friction = 0.85
	cm.stand.friction_threshold = 2.0
	cm.crouch.friction = 0.82
	cm.crouch.friction_threshold = 0.0
	cm.air.gethit.groundlevel = 10.0
	cm.air.gethit.groundrecover.ground.threshold = -20.0
	cm.air.gethit.groundrecover.groundlevel = 10.0
	cm.air.gethit.airrecover.threshold = -1.0
	cm.air.gethit.airrecover.yaccel = 0.35
	cm.air.gethit.trip.groundlevel = 15.0
	cm.down.bounce.offset = [2]float32{0.0, 20.0}
	cm.down.bounce.yaccel = 0.4
	cm.down.bounce.groundlevel = 12.0
	cm.down.friction_threshold = 0.05
}

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
	data             CharData
	velocity         CharVelocity
	movement         CharMovement
	wakewakaLength   int
}
type Char struct {
	name        string
	cmd         []CommandList
	key         int
	helperIndex int
	playerNo    int
	keyctrl     bool
	player      bool
	sprpriority int32
	juggle      int32
	size        CharSize
}

func newChar(n, idx int) (c *Char) {
	c = &Char{}
	c.init(n, idx)
	return c
}
func (c *Char) init(n, idx int) {
	c.playerNo, c.helperIndex = n, idx
	if c.helperIndex == 0 {
		c.keyctrl, c.player = true, true
	}
	c.key = n
	if n >= 0 && n < len(sys.com) && sys.com[n] != 0 {
		c.key ^= -1
	}
}
func (c *Char) load(def string) error {
	gi := &sys.cgi[c.playerNo]
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
	gi.data.init()
	c.size.init()
	gi.velocity.init()
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
				if is.ReadI32("liedown.time", &i32) {
					gi.data.liedown.time = Max(1, i32)
				}
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
				is.ReadF32("walk.fwd", &gi.velocity.walk.fwd)
				is.ReadF32("walk.back", &gi.velocity.walk.back)
				is.ReadF32("walk.up.x", &gi.velocity.walk.up.x)
				is.ReadF32("walk.down.x", &gi.velocity.walk.down.x)
				is.ReadF32("run.fwd", &gi.velocity.run.fwd[0], &gi.velocity.run.fwd[1])
				is.ReadF32("run.back",
					&gi.velocity.run.back[0], &gi.velocity.run.back[1])
				is.ReadF32("run.up.x", &gi.velocity.run.up.x)
				is.ReadF32("run.up.y", &gi.velocity.run.up.y)
				is.ReadF32("run.down.x", &gi.velocity.run.down.x)
				is.ReadF32("run.down.y", &gi.velocity.run.down.y)
				is.ReadF32("jump.neu",
					&gi.velocity.jump.neu[0], &gi.velocity.jump.neu[1])
				is.ReadF32("jump.back", &gi.velocity.jump.back)
				is.ReadF32("jump.fwd", &gi.velocity.jump.fwd)
				is.ReadF32("jump.up.x", &gi.velocity.jump.up.x)
				is.ReadF32("jump.down.x", &gi.velocity.jump.down.x)
				is.ReadF32("runjump.back",
					&gi.velocity.runjump.back[0], &gi.velocity.runjump.back[1])
				is.ReadF32("runjump.fwd",
					&gi.velocity.runjump.fwd[0], &gi.velocity.runjump.fwd[1])
				is.ReadF32("runjump.up.x", &gi.velocity.runjump.up.x)
				is.ReadF32("runjump.down.x", &gi.velocity.runjump.down.x)
				is.ReadF32("airjump.neu",
					&gi.velocity.airjump.neu[0], &gi.velocity.airjump.neu[1])
				is.ReadF32("airjump.back", &gi.velocity.airjump.back)
				is.ReadF32("airjump.fwd", &gi.velocity.airjump.fwd)
				is.ReadF32("airjump.up.x", &gi.velocity.airjump.up.x)
				is.ReadF32("airjump.down.x", &gi.velocity.airjump.down.x)
				is.ReadF32("air.gethit.groundrecover",
					&gi.velocity.air.gethit.groundrecover[0],
					&gi.velocity.air.gethit.groundrecover[1])
				is.ReadF32("air.gethit.airrecover.mul",
					&gi.velocity.air.gethit.airrecover.mul[0],
					&gi.velocity.air.gethit.airrecover.mul[1])
				is.ReadF32("air.gethit.airrecover.add",
					&gi.velocity.air.gethit.airrecover.add[0],
					&gi.velocity.air.gethit.airrecover.add[1])
				is.ReadF32("air.gethit.airrecover.back",
					&gi.velocity.air.gethit.airrecover.back)
				is.ReadF32("air.gethit.airrecover.fwd",
					&gi.velocity.air.gethit.airrecover.fwd)
				is.ReadF32("air.gethit.airrecover.up",
					&gi.velocity.air.gethit.airrecover.up)
				is.ReadF32("air.gethit.airrecover.down",
					&gi.velocity.air.gethit.airrecover.down)
			}
		case "movement":
			if movement {
				movement = false
				is.ReadI32("airjump.num", &gi.movement.airjump.num)
				is.ReadI32("airjump.height", &gi.movement.airjump.height)
				is.ReadF32("yaccel", &gi.movement.yaccel)
				is.ReadF32("stand.friction", &gi.movement.stand.friction)
				is.ReadF32("stand.friction.threshold",
					&gi.movement.stand.friction_threshold)
				is.ReadF32("crouch.friction", &gi.movement.crouch.friction)
				is.ReadF32("crouch.friction.threshold",
					&gi.movement.crouch.friction_threshold)
				is.ReadF32("air.gethit.groundlevel",
					&gi.movement.air.gethit.groundlevel)
				is.ReadF32("air.gethit.groundrecover.ground.threshold",
					&gi.movement.air.gethit.groundrecover.ground.threshold)
				is.ReadF32("air.gethit.groundrecover.groundlevel",
					&gi.movement.air.gethit.groundrecover.groundlevel)
				is.ReadF32("air.gethit.airrecover.threshold",
					&gi.movement.air.gethit.airrecover.threshold)
				is.ReadF32("air.gethit.airrecover.yaccel",
					&gi.movement.air.gethit.airrecover.yaccel)
				is.ReadF32("air.gethit.trip.groundlevel",
					&gi.movement.air.gethit.trip.groundlevel)
				is.ReadF32("down.bounce.offset",
					&gi.movement.down.bounce.offset[0],
					&gi.movement.down.bounce.offset[1])
				is.ReadF32("down.bounce.yaccel", &gi.movement.down.bounce.yaccel)
				is.ReadF32("down.bounce.groundlevel",
					&gi.movement.down.bounce.groundlevel)
				is.ReadF32("down.friction.threshold",
					&gi.movement.down.friction_threshold)
			}
		}
	}
	if LoadFile(&sprite, def, func(filename string) error {
		var err error
		gi.sff, err = LoadSff(filename, false)
		return err
	}); err != nil {
		return err
	}
	if LoadFile(&anim, def, func(filename string) error {
		str, err := LoadText(filename)
		if err != nil {
			return err
		}
		lines, i := SplitAndTrim(str, "\n"), 0
		gi.anim = ReadAnimationTable(gi.sff, lines, &i)
		return nil
	}); err != nil {
		return err
	}
	if len(sound) > 0 {
		if LoadFile(&sound, def, func(filename string) error {
			var err error
			gi.snd, err = LoadSnd(filename)
			return err
		}); err != nil {
			return err
		}
	} else {
		gi.snd = newSnd()
	}
	return nil
}
func (c *Char) clearHitCount() {
	unimplemented()
}
func (c *Char) clearMoveHit() {
	unimplemented()
}
func (c *Char) clearHitDef() {
	unimplemented()
}
func (c *Char) setSprPriority(sprpriority int32) {
	c.sprpriority = sprpriority
}
func (c *Char) faceP2() {
	unimplemented()
}
func (c *Char) setJuggle(juggle int32) {
	c.juggle = juggle
}
func (c *Char) setXV(xv float32) {
	unimplemented()
}
func (c *Char) setYV(yv float32) {
	unimplemented()
}
func (c *Char) changeAnim(animNo int32) {
	unimplemented()
}
func (c *Char) setCtrl(ctrl bool) {
	unimplemented()
}
func (c *Char) addPower(power int32) {
	unimplemented()
}
func (c *Char) time() int32 {
	unimplemented()
	return 0
}

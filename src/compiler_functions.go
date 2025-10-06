package main

import (
	"fmt"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// State controller definition file.
// This file contains the parsing code for the function in ZSS and CNS, also called State Controllers.

func (c *Compiler) hitBySub(is IniSection, sc *StateControllerBase, sctrlName string) error {
	attr := int32(-1)
	var err error
	var any, new, old bool

	// New syntax with attr and slots
	if err = c.stateParam(is, "attr", false, func(data string) error {
		any, new = true, true
		attr, err = c.attr(data, false)
		if err != nil {
			return err
		}
		sc.add(hitBy_attr, sc.iToExp(attr))
		return nil // When running a "sub" function we can't just return sc.add
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "playerno", false, func(data string) error {
		any, new = true, true
		c.scAdd(sc, hitBy_playerno, data, VT_Int, 1)
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "playerid", false, func(data string) error {
		any, new = true, true
		c.scAdd(sc, hitBy_playerid, data, VT_Int, 1)
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "slot", false, func(data string) error {
		new = true
		c.scAdd(sc, hitBy_slot, data, VT_Int, 1)
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "stack", false, func(data string) error {
		new = true
		c.scAdd(sc, hitBy_stack, data, VT_Bool, 1)
		return nil
	}); err != nil {
		return err
	}

	// Shared parameters
	if err := c.paramValue(is, sc, "time", hitBy_time, VT_Int, 1, false); err != nil {
		return err
	}

	// Old syntax with value and value2
	// Must be placed after time
	if err = c.stateParam(is, "value", false, func(data string) error {
		any, old = true, true
		attr, err = c.attr(data, false)
		if err != nil {
			return err
		}
		sc.add(hitBy_slot, sc.iToExp(0))
		sc.add(hitBy_attr, sc.iToExp(attr))
		return nil
	}); err != nil {
		return err
	}
	if !any { // In Mugen if both values are used then value2 will be ignored
		if err = c.stateParam(is, "value2", false, func(data string) error {
			any, old = true, true
			attr, err = c.attr(data, false)
			if err != nil {
				return err
			}
			sc.add(hitBy_slot, sc.iToExp(1))
			sc.add(hitBy_attr, sc.iToExp(attr))
			return nil
		}); err != nil {
			return err
		}
	}

	// Cannot mix old and new syntax
	if old && new {
		if c.zssMode {
			return Error("Cannot mix old and new " + sctrlName + " syntaxes")
		} else {
			sys.appendToConsole("WARNING: " + sys.cgi[c.playerNo].nameLow + fmt.Sprintf(": Cannot mix old and new: "+sctrlName+" in state %v ", c.stateNo))
		}
	}

	// Must have at least one parameter
	if !any {
		return Error("No " + sctrlName + " parameters specified")
	}

	return nil
}

func (c *Compiler) hitBy(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*hitBy)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid", hitBy_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.hitBySub(is, sc, "HitBy")
	})
	return *ret, err
}

func (c *Compiler) notHitBy(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*notHitBy)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid", hitBy_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.hitBySub(is, sc, "NotHitBy")
	})
	return *ret, err
}

func (c *Compiler) assertSpecial(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*assertSpecial)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			assertSpecial_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "enabled",
			assertSpecial_enabled, VT_Bool, 1, false); err != nil {
			return err
		}
		foo := func(data string) error {
			switch strings.ToLower(data) {
			// Mugen char flags
			case "intro":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_intro)))
			case "invisible":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_invisible)))
			case "noairguard":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noairguard)))
			case "noautoturn":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noautoturn)))
			case "nocrouchguard":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nocrouchguard)))
			case "nojugglecheck":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nojugglecheck)))
			case "noshadow":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noshadow)))
			case "nostandguard":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nostandguard)))
			case "nowalk":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nowalk)))
			case "unguardable":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_unguardable)))
			// Mugen global flags
			case "globalnoshadow":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_globalnoshadow)))
			case "nobardisplay":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_nobardisplay)))
			case "nobg":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_nobg)))
			case "nofg":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_nofg)))
			case "noko":
				sc.add(assertSpecial_noko, sc.i64ToExp(0))
			case "nokoslow":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_nokoslow)))
			case "nokosnd":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_nokosnd)))
			case "nomusic":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_nomusic)))
			case "roundnotover":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_roundnotover)))
			case "timerfreeze":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_timerfreeze)))
			// Ikemen char flags
			case "animatehitpause":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_animatehitpause)))
			case "animfreeze":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_animfreeze)))
			case "autoguard":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_autoguard)))
			case "drawunder":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_drawunder)))
			case "noaibuttonjam":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noaibuttonjam)))
			case "noaicheat":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noaicheat)))
			case "noailevel":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noailevel)))
			case "noairjump":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noairjump)))
			case "nobrake":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nobrake)))
			case "nocombodisplay":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nocombodisplay)))
			case "nocrouch":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nocrouch)))
			case "nodizzypointsdamage":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nodizzypointsdamage)))
			case "nofacedisplay":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nofacedisplay)))
			case "nofacep2":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nofacep2)))
			case "nofallcount":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nofallcount)))
			case "nofalldefenceup":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nofalldefenceup)))
			case "nofallhitflag":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nofallhitflag)))
			case "nofastrecoverfromliedown":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nofastrecoverfromliedown)))
			case "nogetupfromliedown":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nogetupfromliedown)))
			case "noguardbardisplay":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noguardbardisplay)))
			case "noguarddamage":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noguarddamage)))
			case "noguardko":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noguardko)))
			case "noguardpointsdamage":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noguardpointsdamage)))
			case "nohardcodedkeys":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nohardcodedkeys)))
			case "nohitdamage":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nohitdamage)))
			case "noinput":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noinput)))
			case "nointroreset":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nointroreset)))
			case "nojump":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nojump)))
			case "nokofall":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nokofall)))
			case "nokovelocity":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nokovelocity)))
			case "nolifebaraction":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nolifebaraction)))
			case "nolifebardisplay":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nolifebardisplay)))
			case "nomakedust":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nomakedust)))
			case "nonamedisplay":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nonamedisplay)))
			case "nopowerbardisplay":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nopowerbardisplay)))
			case "noredlifedamage":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noredlifedamage)))
			case "nostand":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nostand)))
			case "nostunbardisplay":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nostunbardisplay)))
			case "noturntarget":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_noturntarget)))
			case "nowinicondisplay":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_nowinicondisplay)))
			case "postroundinput":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_postroundinput)))
			case "projtypecollision":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_projtypecollision)))
			case "runfirst":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_runfirst)))
			case "runlast":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_runlast)))
			case "sizepushonly":
				sc.add(assertSpecial_flag, sc.i64ToExp(int64(ASF_sizepushonly)))
			// Ikemen global flags
			case "camerafreeze":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_camerafreeze)))
			case "globalnoko":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_globalnoko)))
			case "roundnotskip":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_roundnotskip)))
			case "roundfreeze":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_roundfreeze)))
			case "skipfightdisplay":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_skipfightdisplay)))
			case "skipkodisplay":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_skipkodisplay)))
			case "skiprounddisplay":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_skiprounddisplay)))
			case "skipwindisplay":
				sc.add(assertSpecial_flag_g, sc.i64ToExp(int64(GSF_skipwindisplay)))
			default:
				return Error("Invalid AssertSpecial flag: " + data)
			}
			return nil
		}
		f := false
		if err := c.stateParam(is, "flag", false, func(data string) error {
			f = true
			return foo(data)
		}); err != nil {
			return err
		}
		if !f {
			return Error("AssertSpecial flag not specified")
		}
		if err := c.stateParam(is, "flag2", false, func(data string) error {
			return foo(data)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "flag3", false, func(data string) error {
			return foo(data)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "flag4", false, func(data string) error {
			return foo(data)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "flag5", false, func(data string) error {
			return foo(data)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "flag6", false, func(data string) error {
			return foo(data)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "flag7", false, func(data string) error {
			return foo(data)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "flag8", false, func(data string) error {
			return foo(data)
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) playSnd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*playSnd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			playSnd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "value", true, func(data string) error {
			prefix := c.getDataPrefix(&data, false)
			return c.scAdd(sc, playSnd_value, data, VT_Int, 2,
				sc.beToExp(BytecodeExp(prefix))...)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "channel",
			playSnd_channel, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "lowpriority",
			playSnd_lowpriority, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pan",
			playSnd_pan, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "abspan",
			playSnd_abspan, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "volume",
			playSnd_volume, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "volumescale",
			playSnd_volumescale, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "freqmul",
			playSnd_freqmul, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loop",
			playSnd_loop, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "priority",
			playSnd_priority, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loopstart",
			playSnd_loopstart, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loopend",
			playSnd_loopend, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "startposition",
			playSnd_startposition, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loopcount",
			playSnd_loopcount, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stopongethit",
			playSnd_stopongethit, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stoponchangestate",
			playSnd_stoponchangestate, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) changeStateSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "redirectid",
		changeState_redirectid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "value",
		changeState_value, VT_Int, 1, true); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "ctrl",
		changeState_ctrl, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "anim", false, func(data string) error {
		prefix := c.getDataPrefix(&data, false)
		return c.scAdd(sc, changeState_anim, data, VT_Int, 1,
			sc.beToExp(BytecodeExp(prefix))...)
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "continue",
		changeState_continue, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "readplayerid",
		changeState_readplayerid, VT_Int, 1, false); err != nil {
		return err
	}
	if c.block != nil && c.stateNo >= 0 && c.block.ignorehitpause == -1 {
		// Assign a unique index to ignorehitpause for this controller
		c.block.ignorehitpause = sys.cgi[c.playerNo].hitPauseToggleFlagCount
		// Increment the count of hitPauseExecutionToggleFlags
		sys.cgi[c.playerNo].hitPauseToggleFlagCount++
	}
	return nil
}

func (c *Compiler) changeState(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*changeState)(sc), c.stateSec(is, func() error {
		return c.changeStateSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) selfState(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*selfState)(sc), c.stateSec(is, func() error {
		return c.changeStateSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) tagIn(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*tagIn)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			tagIn_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "self", tagIn_self, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "partner", tagIn_partner, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stateno", tagIn_stateno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "partnerstateno", tagIn_partnerstateno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ctrl", tagIn_ctrl, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "partnerctrl", tagIn_partnerctrl, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "leader", tagIn_leader, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	//if c.block != nil && c.block.ignorehitpause == -1 {
	//	c.block.ignorehitpause = sys.cgi[c.playerNo].hitPauseToggleFlagCount
	//	sys.cgi[c.playerNo].hitPauseToggleFlagCount++
	//}
	return *ret, err
}

func (c *Compiler) tagOut(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*tagOut)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			tagOut_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "self", tagOut_self, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "partner", tagOut_partner, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stateno", tagOut_stateno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "partnerstateno", tagOut_partnerstateno, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	//if c.block != nil && c.block.ignorehitpause == -1 {
	//	c.block.ignorehitpause = sys.cgi[c.playerNo].hitPauseToggleFlagCount
	//	sys.cgi[c.playerNo].hitPauseToggleFlagCount++
	//}
	return *ret, err
}

func (c *Compiler) destroySelf(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*destroySelf)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			destroySelf_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "recursive",
			destroySelf_recursive, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "removeexplods",
			destroySelf_removeexplods, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "removetexts",
			destroySelf_removetexts, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) changeAnimSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "redirectid",
		changeAnim_redirectid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "animplayerno",
		changeAnim_animplayerno, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "spriteplayerno",
		changeAnim_spriteplayerno, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "readplayerid",
		changeAnim_readplayerid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "elem",
		changeAnim_elem, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "elemtime",
		changeAnim_elemtime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "value", true, func(data string) error {
		prefix := c.getDataPrefix(&data, false)
		return c.scAdd(sc, changeAnim_value, data, VT_Int, 1,
			sc.beToExp(BytecodeExp(prefix))...)
	}); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) changeAnim(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*changeAnim)(sc), c.stateSec(is, func() error {
		return c.changeAnimSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) changeAnim2(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*changeAnim2)(sc), c.stateSec(is, func() error {
		return c.changeAnimSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) helper(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*helper)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			helper_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "helpertype", false, func(data string) error {
			if len(data) == 0 {
				return Error("helpertype not specified")
			}
			var ht int32
			switch strings.ToLower(data) {
			case "normal":
				ht = 0
			case "player":
				ht = 1
			case "projectile":
				ht = 2
			case "proj":
				ht = 2
			default:
				return Error("Invalid helpertype: " + data)
			}
			sc.add(helper_helpertype, sc.iToExp(ht))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "clsnproxy",
			helper_clsnproxy, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "name", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Helper name not enclosed in \"")
			}
			sc.add(helper_name, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramPostype(is, sc, helper_postype); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ownpal",
			helper_ownpal, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.xscale",
			helper_size_xscale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.yscale",
			helper_size_yscale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.ground.back",
			helper_size_ground_back, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.ground.front",
			helper_size_ground_front, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.air.back",
			helper_size_air_back, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.air.front",
			helper_size_air_front, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.height",
			helper_size_height_stand, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.height.crouch",
			helper_size_height_crouch, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.height.air",
			helper_size_height_air, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.height.down",
			helper_size_height_down, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.proj.doscale",
			helper_size_proj_doscale, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.head.pos",
			helper_size_head_pos, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.mid.pos",
			helper_size_mid_pos, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.shadowoffset",
			helper_size_shadowoffset, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.depth",
			helper_size_depth, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.weight",
			helper_size_weight, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "size.pushfactor",
			helper_size_pushfactor, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stateno",
			helper_stateno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "keyctrl", false, func(data string) error {
			bes, err := c.exprs(data, VT_Int, 4)
			if err != nil {
				return err
			}
			sc.add(helper_keyctrl, bes)
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			helper_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pos",
			helper_pos, VT_Float, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "facing",
			helper_facing, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pausemovetime",
			helper_pausemovetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "supermovetime",
			helper_supermovetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "remappal",
			helper_remappal, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "extendsmap",
			helper_extendsmap, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "inheritjuggle",
			helper_inheritjuggle, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "inheritchannels",
			helper_inheritchannels, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "immortal",
			helper_immortal, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "kovelocity",
			helper_kovelocity, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "preserve",
			helper_preserve, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "standby",
			helper_standby, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ownclsnscale",
			helper_ownclsnscale, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) ctrlSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*ctrlSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			ctrlSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", ctrlSet_value, VT_Bool, 1, true)
	})
	return *ret, err
}

func (c *Compiler) explodSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "remappal",
		explod_remappal, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "id",
		explod_id, VT_Int, 1, false); err != nil {
		return err
	}
	// Postype must be placed before parameters such as pos, for the sake of ModifyExplod
	if err := c.paramPostype(is, sc, explod_postype); err != nil {
		return err
	}
	if err := c.paramSpace(is, sc, explod_space); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "facing",
		explod_facing, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "vfacing",
		explod_vfacing, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "pos",
		explod_pos, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "random",
		explod_random, VT_Float, 3, false); err != nil {
		return err
	}
	found := false
	if err := c.stateParam(is, "vel", false, func(data string) error {
		found = true
		return c.scAdd(sc, explod_velocity, data, VT_Float, 3)
	}); err != nil {
		return err
	}
	if !found {
		if err := c.paramValue(is, sc, "velocity",
			explod_velocity, VT_Float, 3, false); err != nil {
			return err
		}
	}
	if err := c.paramValue(is, sc, "friction",
		explod_friction, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "accel",
		explod_accel, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramProjection(is, sc, "projection",
		explod_projection); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "scale",
		explod_scale, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "bindid",
		explod_bindid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "bindtime",
		explod_bindtime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "removetime",
		explod_removetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "supermove",
		explod_supermove, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "supermovetime",
		explod_supermovetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "pausemovetime",
		explod_pausemovetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "sprpriority",
		explod_sprpriority, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "ontop",
		explod_ontop, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "under",
		explod_under, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "layerno",
		explod_layerno, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "shadow",
		explod_shadow, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "removeongethit",
		explod_removeongethit, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "removeonchangestate",
		explod_removeonchangestate, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramTrans(is, sc, "", explod_trans, false); err != nil {
		return err
	}
	if err := c.palFXSub(is, sc, "palfx."); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "window",
		explod_window, VT_Float, 4, false); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) explodInterpolate(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "interpolation.time",
		explod_interpolate_time, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "interpolation.animelem",
		explod_interpolate_animelem, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "interpolation.scale",
		explod_interpolate_scale, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "interpolation.angle",
		explod_interpolate_angle, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "interpolation.alpha",
		explod_interpolate_alpha, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "interpolation.offset",
		explod_interpolate_pos, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "interpolation.focallength",
		explod_interpolate_focallength, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "interpolation.xshear",
		explod_interpolate_xshear, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "interpolation.palfx.mul",
		explod_interpolate_pfx_mul, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "interpolation.palfx.add",
		explod_interpolate_pfx_add, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "interpolation.palfx.color",
		explod_interpolate_pfx_color, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "interpolation.palfx.hue",
		explod_interpolate_pfx_hue, VT_Float, 1, false); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) explod(is IniSection, sc *StateControllerBase,
	ihp int8) (StateController, error) {
	ret, err := (*explod)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			explod_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "animplayerno",
			explod_animplayerno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "spriteplayerno",
			explod_spriteplayerno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "anim", false, func(data string) error {
			prefix := c.getDataPrefix(&data, false)
			return c.scAdd(sc, explod_anim, data, VT_Int, 1,
				sc.beToExp(BytecodeExp(prefix))...)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ownpal",
			explod_ownpal, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.explodSub(is, sc); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "animelem",
			explod_animelem, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "animelemtime",
			explod_animelemtime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "animfreeze",
			explod_animfreeze, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "angle",
			explod_angle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yangle",
			explod_yangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xangle",
			explod_xangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xshear",
			explod_xshear, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "focallength",
			explod_focallength, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.explodInterpolate(is, sc); err != nil {
			return err
		}
		if ihp == 0 {
			sc.add(explod_ignorehitpause, sc.iToExp(0))
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) modifyExplod(is IniSection, sc *StateControllerBase,
	ihp int8) (StateController, error) {
	ret, err := (*modifyExplod)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			modifyexplod_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "index",
			modifyexplod_index, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.explodSub(is, sc); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "animplayerno",
			explod_animplayerno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "spriteplayerno",
			explod_spriteplayerno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "anim", false, func(data string) error {
			prefix := c.getDataPrefix(&data, false)
			return c.scAdd(sc, explod_anim, data, VT_Int, 1,
				sc.beToExp(BytecodeExp(prefix))...)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "animelem",
			explod_animelem, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "animelemtime",
			explod_animelemtime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "animfreeze",
			explod_animfreeze, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "angle",
			explod_angle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yangle",
			explod_yangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xangle",
			explod_xangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xshear",
			explod_xshear, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "focallength",
			explod_focallength, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "interpolation",
			explod_interpolation, VT_Bool, 1, false); err != nil {
			return err
		}
		if ihp == 0 {
			sc.add(explod_ignorehitpause, sc.iToExp(0))
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) gameMakeAnim(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*gameMakeAnim)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			gameMakeAnim_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pos",
			gameMakeAnim_pos, VT_Float, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "random",
			gameMakeAnim_random, VT_Float, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "under",
			gameMakeAnim_under, VT_Bool, 1, false); err != nil {
			return err
		}

		// Previously, Ikemen accepted either "value" or "anim" here. Turns out Mugen only accepts "value"
		anim := func(data string) error {
			prefix := c.getDataPrefix(&data, true)
			return c.scAdd(sc, gameMakeAnim_anim, data, VT_Int, 1,
				sc.beToExp(BytecodeExp(prefix))...)
		}
		if err := c.stateParam(is, "value", false, func(data string) error {
			return anim(data)
		}); err != nil {
			return err
		}

		return nil
	})
	return *ret, err
}

func (c *Compiler) posSetSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "x",
		posSet_x, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "y",
		posSet_y, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "z",
		posSet_z, VT_Float, 1, false); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) posSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*posSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			posSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.posSetSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) posAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*posAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			posSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.posSetSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) velSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*velSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			posSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.posSetSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) velAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*velAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			posSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.posSetSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) velMul(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*velMul)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			posSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.posSetSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) modifyShadow(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*modifyShadow)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			modifyShadow_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "color",
			modifyShadow_color, VT_Int, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "intensity",
			modifyShadow_intensity, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "offset",
			modifyShadow_offset, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "window",
			modifyShadow_window, VT_Float, 4, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xshear",
			modifyShadow_xshear, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yscale",
			modifyShadow_yscale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "angle",
			modifyShadow_angle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xangle",
			modifyShadow_xangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yangle",
			modifyShadow_yangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "focallength",
			modifyShadow_focallength, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramProjection(is, sc, "projection",
			modifyShadow_projection); err != nil {
			return err
		}
		return c.posSetSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) modifyReflection(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*modifyReflection)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			modifyReflection_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "color",
			modifyReflection_color, VT_Int, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "intensity",
			modifyReflection_intensity, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "offset",
			modifyReflection_offset, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "window",
			modifyReflection_window, VT_Float, 4, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xshear",
			modifyReflection_xshear, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yscale",
			modifyReflection_yscale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "angle",
			modifyReflection_angle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xangle",
			modifyReflection_xangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yangle",
			modifyReflection_yangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "focallength",
			modifyReflection_focallength, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramProjection(is, sc, "projection",
			modifyReflection_projection); err != nil {
			return err
		}
		return c.posSetSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) palFXSub(is IniSection,
	sc *StateControllerBase, prefix string) error {
	if err := c.paramValue(is, sc, prefix+"time",
		palFX_time, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"color",
		palFX_color, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"hue",
		palFX_hue, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, prefix+"add", false, func(data string) error {
		bes, err := c.exprs(data, VT_Int, 3)
		if err != nil {
			return err
		}
		if len(bes) < 3 {
			return Error(prefix + "add - not enough arguments")
		}
		sc.add(palFX_add, bes)
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, prefix+"mul", false, func(data string) error {
		bes, err := c.exprs(data, VT_Int, 3)
		if err != nil {
			return err
		}
		if len(bes) < 3 {
			return Error(prefix + "mul - not enough arguments")
		}
		sc.add(palFX_mul, bes)
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, prefix+"sinadd", false, func(data string) error {
		bes, err := c.exprs(data, VT_Int, 4)
		if err != nil {
			return err
		}
		if len(bes) < 3 {
			return Error(prefix + "sinadd - not enough arguments")
		}
		sc.add(palFX_sinadd, bes)
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, prefix+"sinmul", false, func(data string) error {
		bes, err := c.exprs(data, VT_Int, 4)
		if err != nil {
			return err
		}
		if len(bes) < 3 {
			return Error(prefix + "sinmul - not enough arguments")
		}
		sc.add(palFX_sinmul, bes)
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, prefix+"sincolor", false, func(data string) error {
		bes, err := c.exprs(data, VT_Int, 2)
		if err != nil {
			return err
		}
		if len(bes) < 2 {
			return Error(prefix + "sincolor - not enough arguments")
		}
		sc.add(palFX_sincolor, bes)
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, prefix+"sinhue", false, func(data string) error {
		bes, err := c.exprs(data, VT_Int, 2)
		if err != nil {
			return err
		}
		if len(bes) < 2 {
			return Error(prefix + "sinhue - not enough arguments")
		}
		sc.add(palFX_sinhue, bes)
		return nil
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"invertall",
		palFX_invertall, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"invertblend",
		palFX_invertblend, VT_Int, 1, false); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) palFX(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*palFX)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			palFX_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.palFXSub(is, sc, "")
	})
	return *ret, err
}

func (c *Compiler) allPalFX(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*allPalFX)(sc), c.stateSec(is, func() error {
		return c.palFXSub(is, sc, "")
	})
	return *ret, err
}

func (c *Compiler) bgPalFX(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*bgPalFX)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "id",
			bgPalFX_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "index",
			bgPalFX_index, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.palFXSub(is, sc, ""); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) afterImageSub(is IniSection,
	sc *StateControllerBase, ihp int8, prefix string) error {
	if err := c.paramValue(is, sc, "redirectid",
		afterImage_redirectid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramTrans(is, sc, prefix, afterImage_trans, true); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"time",
		afterImage_time, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"length",
		afterImage_length, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"timegap",
		afterImage_timegap, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"framegap",
		afterImage_framegap, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palcolor",
		afterImage_palcolor, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palhue",
		afterImage_palhue, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palinvertall",
		afterImage_palinvertall, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palinvertblend",
		afterImage_palinvertblend, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palbright",
		afterImage_palbright, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palcontrast",
		afterImage_palcontrast, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palpostbright",
		afterImage_palpostbright, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"paladd",
		afterImage_paladd, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, prefix+"palmul",
		afterImage_palmul, VT_Float, 3, false); err != nil {
		return err
	}
	if ihp == 0 {
		sc.add(afterImage_ignorehitpause, sc.iToExp(0))
	}
	return nil
}

func (c *Compiler) afterImage(is IniSection, sc *StateControllerBase,
	ihp int8) (StateController, error) {
	ret, err := (*afterImage)(sc), c.stateSec(is, func() error {
		return c.afterImageSub(is, sc, ihp, "")
	})
	return *ret, err
}

func (c *Compiler) afterImageTime(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*afterImageTime)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			afterImageTime_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "time", false, func(data string) error {
			b = true
			return c.scAdd(sc, afterImageTime_time, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			if err := c.stateParam(is, "value", false, func(data string) error {
				b = true
				return c.scAdd(sc, afterImageTime_time, data, VT_Int, 1)
			}); err != nil {
				return err
			}
			if !b {
				sc.add(afterImageTime_time, sc.iToExp(0))
			}
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) parseHitFlag(sc *StateControllerBase, id byte, data string) error {
	var flg int32
	for _, c := range data {
		switch c {
		case 'H', 'h':
			flg |= int32(HF_H)
		case 'L', 'l':
			flg |= int32(HF_L)
		case 'M', 'm':
			flg |= int32(HF_H | HF_L)
		case 'A', 'a':
			flg |= int32(HF_A)
		case 'F', 'f':
			flg |= int32(HF_F)
		case 'D', 'd':
			flg |= int32(HF_D)
		case 'P', 'p':
			flg |= int32(HF_P)
		case '-':
			flg |= int32(HF_MNS)
		case '+':
			flg |= int32(HF_PLS)
		}
	}
	sc.add(id, sc.iToExp(flg))
	return nil
}

func (c *Compiler) hitDefSub(is IniSection, sc *StateControllerBase) error {
	if err := c.stateParam(is, "attr", false, func(data string) error {
		attr, err := c.attr(data, true)
		if err != nil {
			return err
		}
		sc.add(hitDef_attr, sc.iToExp(attr))
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "guardflag", false, func(data string) error {
		return c.parseHitFlag(sc, hitDef_guardflag, data)
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "hitflag", false, func(data string) error {
		return c.parseHitFlag(sc, hitDef_hitflag, data)
	}); err != nil {
		return err
	}
	if err := c.paramHittype(is, sc, "ground.type", hitDef_ground_type); err != nil {
		return err
	}
	if err := c.paramHittype(is, sc, "air.type", hitDef_air_type); err != nil {
		return err
	}
	if err := c.paramAnimtype(is, sc, "animtype", hitDef_animtype); err != nil {
		return err
	}
	if err := c.paramAnimtype(is, sc, "air.animtype", hitDef_air_animtype); err != nil {
		return err
	}
	if err := c.paramAnimtype(is, sc, "fall.animtype", hitDef_fall_animtype); err != nil {
		return err
	}
	if err := c.stateParam(is, "affectteam", false, func(data string) error {
		if len(data) == 0 {
			return Error("affectteam not specified")
		}
		var at int32
		switch data[0] {
		case 'E', 'e':
			at = 1
		case 'B', 'b':
			at = 0
		case 'F', 'f':
			at = -1
		default:
			return Error("Invalid affectteam: " + data)
		}
		sc.add(hitDef_affectteam, sc.iToExp(at))
		return nil
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "teamside",
		hitDef_teamside, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "id",
		hitDef_id, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "chainid",
		hitDef_chainid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "nochainid",
		hitDef_nochainid, VT_Int, MaxSimul*2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "kill",
		hitDef_kill, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.kill",
		hitDef_guard_kill, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.kill",
		hitDef_fall_kill, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "hitonce",
		hitDef_hitonce, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "air.juggle",
		hitDef_air_juggle, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "getpower",
		hitDef_getpower, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "damage",
		hitDef_damage, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "givepower",
		hitDef_givepower, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "numhits",
		hitDef_numhits, VT_Int, 1, false); err != nil {
		return err
	}
	hsnd := func(id byte, data string) error {
		prefix := c.getDataPrefix(&data, true)
		return c.scAdd(sc, id, data, VT_Int, 2, sc.beToExp(BytecodeExp(prefix))...)
	}
	if err := c.stateParam(is, "hitsound", false, func(data string) error {
		return hsnd(hitDef_hitsound, data)
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "hitsound.channel",
		hitDef_hitsound_channel, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "guardsound", false, func(data string) error {
		return hsnd(hitDef_guardsound, data)
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guardsound.channel",
		hitDef_guardsound_channel, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "priority", false, func(data string) error {
		be, err := c.argExpression(&data, VT_Int)
		if err != nil {
			return err
		}
		at := TT_Hit
		data = strings.TrimSpace(data)
		if c.token == "," && len(data) > 0 {
			switch data[0] {
			case 'H', 'h':
				at = TT_Hit
			case 'M', 'm':
				at = TT_Miss
			case 'D', 'd':
				at = TT_Dodge
			default:
				return Error("Invalid priority type: " + data)
			}
		}
		sc.add(hitDef_priority, append(sc.beToExp(be), sc.iToExp(int32(at))...))
		return nil
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "p1stateno",
		hitDef_p1stateno, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "p2stateno",
		hitDef_p2stateno, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "p2getp1state",
		hitDef_p2getp1state, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "missonoverride",
		hitDef_missonoverride, VT_Bool, 1, false); err != nil {
		return err
	}
	b := false
	if err := c.stateParam(is, "p1sprpriority", false, func(data string) error {
		b = true
		return c.scAdd(sc, hitDef_p1sprpriority, data, VT_Int, 1)
	}); err != nil {
		return err
	}
	if !b {
		// DOS Mugen used "sprpriority" instead of "p1sprpriority". Later versions seemingly kept both syntaxes
		if err := c.paramValue(is, sc, "sprpriority",
			hitDef_p1sprpriority, VT_Int, 1, false); err != nil {
			return err
		}
	}
	if err := c.paramValue(is, sc, "p2sprpriority",
		hitDef_p2sprpriority, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "forcestand",
		hitDef_forcestand, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "forcecrouch",
		hitDef_forcecrouch, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "forcenofall",
		hitDef_forcenofall, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.damage",
		hitDef_fall_damage, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.xvelocity",
		hitDef_fall_xvelocity, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.yvelocity",
		hitDef_fall_yvelocity, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.zvelocity",
		hitDef_fall_zvelocity, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.recover",
		hitDef_fall_recover, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.recovertime",
		hitDef_fall_recovertime, VT_Int, 1, false); err != nil {
		return err
	}
	sprk := func(id byte, data string) error {
		prefix := c.getDataPrefix(&data, true)
		return c.scAdd(sc, id, data, VT_Int, 1,
			sc.beToExp(BytecodeExp(prefix))...)
	}
	if err := c.stateParam(is, "sparkno", false, func(data string) error {
		return sprk(hitDef_sparkno, data)
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "sparkangle",
		hitDef_sparkangle, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "guard.sparkno", false, func(data string) error {
		return sprk(hitDef_guard_sparkno, data)
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.sparkangle",
		hitDef_guard_sparkangle, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "sparkxy",
		hitDef_sparkxy, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "down.hittime",
		hitDef_down_hittime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "p1facing",
		hitDef_p1facing, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "p1getp2facing",
		hitDef_p1getp2facing, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "mindist",
		hitDef_mindist, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "maxdist",
		hitDef_maxdist, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "snap",
		hitDef_snap, VT_Float, 4, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "p2facing",
		hitDef_p2facing, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "air.hittime",
		hitDef_air_hittime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall",
		hitDef_fall, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "air.fall",
		hitDef_air_fall, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "air.cornerpush.veloff",
		hitDef_air_cornerpush_veloff, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "down.bounce",
		hitDef_down_bounce, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "down.velocity",
		hitDef_down_velocity, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "down.cornerpush.veloff",
		hitDef_down_cornerpush_veloff, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "ground.hittime",
		hitDef_ground_hittime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.hittime",
		hitDef_guard_hittime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.dist", // Old syntax
		hitDef_guard_dist_x, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.dist.width", // New syntax
		hitDef_guard_dist_x, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.dist.height",
		hitDef_guard_dist_y, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.dist.depth",
		hitDef_guard_dist_z, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "pausetime",
		hitDef_pausetime, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.pausetime",
		hitDef_guard_pausetime, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "air.velocity",
		hitDef_air_velocity, VT_Float, 3, false); err != nil { // This one also accepts "n" in Mugen, like ground.velocity
		return err
	}
	if err := c.paramValue(is, sc, "airguard.velocity",
		hitDef_airguard_velocity, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "ground.slidetime",
		hitDef_ground_slidetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.slidetime",
		hitDef_guard_slidetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.ctrltime",
		hitDef_guard_ctrltime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "airguard.ctrltime",
		hitDef_airguard_ctrltime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "ground.velocity", false, func(data string) error {
		in := data
		// This parameter accepts "n" as a value in Mugen. Documented in WinMugen updates.txt
		// In Mugen, supposedly every velocity parameter should accept "n", but that makes most of them crash
		// It's also not in the HitDef documentation, so it seems to have been dropped halfway through its implementation
		// "n" is supposed to not change the enemy's velocity. In Ikemen it just sets velocity to 0, 0, 0
		// TODO: Should we implement this feature anyway? Few or no characters seem to use it
		if c.token = c.tokenizer(&in); c.token == "n" {
			if c.token = c.tokenizer(&in); len(c.token) > 0 && c.token != "," {
				return Error("Invalid ground.velocity value: " + c.token)
			}
		} else {
			in = data
			be, err := c.argExpression(&in, VT_Float)
			if err != nil {
				return err
			}
			sc.add(hitDef_ground_velocity_x, sc.beToExp(be))
		}
		if c.token == "," {
			oldin := in
			if c.token = c.tokenizer(&in); c.token == "n" {
				if c.token = c.tokenizer(&in); len(c.token) > 0 {
					return Error("Invalid ground.velocity value: " + c.token)
				}
			} else {
				in = oldin
				be, err := c.argExpression(&in, VT_Float)
				if err != nil {
					return err
				}
				sc.add(hitDef_ground_velocity_y, sc.beToExp(be))
			}
		}
		if c.token == "," {
			oldin := in
			if c.token = c.tokenizer(&in); c.token == "n" {
				if c.token = c.tokenizer(&in); len(c.token) > 0 {
					return Error("Invalid ground.velocity: " + c.token)
				}
			} else {
				in = oldin
				be, err := c.fullExpression(&in, VT_Float)
				if err != nil {
					return err
				}
				sc.add(hitDef_ground_velocity_z, sc.beToExp(be))
			}
		}
		return nil
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.velocity",
		hitDef_guard_velocity, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "ground.cornerpush.veloff",
		hitDef_ground_cornerpush_veloff, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.cornerpush.veloff",
		hitDef_guard_cornerpush_veloff, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "airguard.cornerpush.veloff",
		hitDef_airguard_cornerpush_veloff, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "xaccel",
		hitDef_xaccel, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "yaccel",
		hitDef_yaccel, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "zaccel",
		hitDef_zaccel, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.palFXSub(is, sc, "palfx."); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "envshake.time",
		hitDef_envshake_time, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "envshake.ampl",
		hitDef_envshake_ampl, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "envshake.freq",
		hitDef_envshake_freq, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "envshake.phase",
		hitDef_envshake_phase, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "envshake.mul",
		hitDef_envshake_mul, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "envshake.dir",
		hitDef_envshake_dir, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.envshake.time",
		hitDef_fall_envshake_time, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.envshake.ampl",
		hitDef_fall_envshake_ampl, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.envshake.freq",
		hitDef_fall_envshake_freq, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.envshake.phase",
		hitDef_fall_envshake_phase, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.envshake.mul",
		hitDef_fall_envshake_mul, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "fall.envshake.dir",
		hitDef_fall_envshake_dir, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "dizzypoints",
		hitDef_dizzypoints, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guardpoints",
		hitDef_guardpoints, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "redlife",
		hitDef_redlife, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "score",
		hitDef_score, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "p2clsncheck", false, func(data string) error {
		if len(data) == 0 {
			return Error("p2clsncheck not specified")
		}
		var box int32
		switch strings.ToLower(data) {
		case "none":
			box = 0
		case "clsn1":
			box = 1
		case "clsn2":
			box = 2
		case "size":
			box = 3
		default:
			return Error("Invalid p2clsncheck type: " + data)
		}
		sc.add(hitDef_p2clsncheck, sc.iToExp(box))
		return nil
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "p2clsnrequire", false, func(data string) error {
		if len(data) == 0 {
			return Error("p2clsnrequire not specified")
		}
		var box int32
		switch strings.ToLower(data) {
		case "none":
			box = 0
		case "clsn1":
			box = 1
		case "clsn2":
			box = 2
		case "size":
			box = 3
		default:
			return Error("Invalid p2clsnrequire type: " + data)
		}
		sc.add(hitDef_p2clsnrequire, sc.iToExp(box))
		return nil
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "down.recover",
		hitDef_down_recover, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "down.recovertime",
		hitDef_down_recovertime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "attack.depth",
		hitDef_attack_depth, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "sparkscale",
		hitDef_sparkscale, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "guard.sparkscale",
		hitDef_guard_sparkscale, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "unhittabletime",
		hitDef_unhittabletime, VT_Int, 2, false); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) hitDef(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*hitDef)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitDef_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.hitDefSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) modifyHitDef(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*modifyHitDef)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			modifyHitDef_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.hitDefSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) reversalDef(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*reversalDef)(sc), c.stateSec(is, func() error {
		attr := int32(-1)
		var err error
		if err = c.paramValue(is, sc, "redirectid",
			reversalDef_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err = c.stateParam(is, "reversal.attr", false, func(data string) error {
			attr, err = c.attr(data, false)
			return err
		}); err != nil {
			return err
		}
		if attr == -1 {
			return Error("ReversalDef reversal.attr not specified")
		} else {
			sc.add(reversalDef_reversal_attr, sc.iToExp(attr))
		}
		if err := c.stateParam(is, "reversal.guardflag", false, func(data string) error {
			return c.parseHitFlag(sc, reversalDef_reversal_guardflag, data)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "reversal.guardflag.not", false, func(data string) error {
			return c.parseHitFlag(sc, reversalDef_reversal_guardflag_not, data)
		}); err != nil {
			return err
		}
		return c.hitDefSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) modifyReversalDef(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*modifyReversalDef)(sc), c.stateSec(is, func() error {
		var err error
		if err = c.paramValue(is, sc, "redirectid",
			modifyReversalDef_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err = c.stateParam(is, "reversal.attr", false, func(data string) error {
			attr, err := c.attr(data, true)
			if err != nil {
				return err
			}
			sc.add(modifyReversalDef_reversal_attr, sc.iToExp(attr))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "reversal.guardflag", false, func(data string) error {
			return c.parseHitFlag(sc, modifyReversalDef_reversal_guardflag, data)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "reversal.guardflag.not", false, func(data string) error {
			return c.parseHitFlag(sc, modifyReversalDef_reversal_guardflag_not, data)
		}); err != nil {
			return err
		}
		return c.hitDefSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) projectileSub(is IniSection, sc *StateControllerBase, ihp int8) error {
	if err := c.paramValue(is, sc, "redirectid",
		projectile_redirectid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramPostype(is, sc, projectile_postype); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projid",
		projectile_projid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projremove",
		projectile_projremove, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projremovetime",
		projectile_projremovetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projshadow",
		projectile_projshadow, VT_Int, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projmisstime",
		projectile_projmisstime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projhits",
		projectile_projhits, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projpriority",
		projectile_projpriority, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "projhitanim", false, func(data string) error {
		prefix := c.getDataPrefix(&data, false)
		return c.scAdd(sc, projectile_projhitanim, data, VT_Int, 1,
			sc.beToExp(BytecodeExp(prefix))...)
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "projremanim", false, func(data string) error {
		prefix := c.getDataPrefix(&data, false)
		return c.scAdd(sc, projectile_projremanim, data, VT_Int, 1,
			sc.beToExp(BytecodeExp(prefix))...)
	}); err != nil {
		return err
	}
	if err := c.stateParam(is, "projcancelanim", false, func(data string) error {
		prefix := c.getDataPrefix(&data, false)
		return c.scAdd(sc, projectile_projcancelanim, data, VT_Int, 1,
			sc.beToExp(BytecodeExp(prefix))...)
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "velocity",
		projectile_velocity, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "velmul",
		projectile_velmul, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "remvelocity",
		projectile_remvelocity, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "accel",
		projectile_accel, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projscale",
		projectile_projscale, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projangle",
		projectile_projangle, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projxangle",
		projectile_projxangle, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projyangle",
		projectile_projyangle, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projclsnscale",
		projectile_projclsnscale, VT_Float, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projclsnangle",
		projectile_projclsnangle, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projwindow",
		projectile_projwindow, VT_Float, 4, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projxshear",
		projectile_projxshear, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projfocallength",
		projectile_projfocallength, VT_Float, 1, false); err != nil {
		return err
	}
	if err := c.paramProjection(is, sc, "projprojection",
		projectile_projprojection); err != nil {
		return err
	}
	// HitDef section
	if err := c.hitDefSub(is, sc); err != nil {
		return err
	}

	if err := c.paramValue(is, sc, "offset",
		projectile_offset, VT_Float, 3, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projsprpriority",
		projectile_projsprpriority, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projlayerno",
		projectile_projlayerno, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projstagebound",
		projectile_projstagebound, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projedgebound",
		projectile_projedgebound, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projheightbound",
		projectile_projheightbound, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "projdepthbound",
		projectile_projdepthbound, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "projanim", false, func(data string) error {
		prefix := c.getDataPrefix(&data, false)
		return c.scAdd(sc, projectile_projanim, data, VT_Int, 1,
			sc.beToExp(BytecodeExp(prefix))...)
	}); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "supermovetime",
		projectile_supermovetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "pausemovetime",
		projectile_pausemovetime, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "ownpal",
		projectile_ownpal, VT_Bool, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "remappal",
		projectile_remappal, VT_Int, 2, false); err != nil {
		return err
	}
	if err := c.afterImageSub(is, sc, ihp, "afterimage."); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) projectile(is IniSection, sc *StateControllerBase, ihp int8) (StateController, error) {
	ret, err := (*projectile)(sc), c.stateSec(is, func() error {
		if err := c.projectileSub(is, sc, ihp); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) modifyProjectile(is IniSection, sc *StateControllerBase,
	ihp int8) (StateController, error) {
	ret, err := (*modifyProjectile)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			modifyProjectile_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			modifyProjectile_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "index",
			modifyProjectile_index, VT_Int, 1, false); err != nil {
			return err
		}

		if err := c.projectileSub(is, sc, ihp); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) width(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*width)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			width_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "edge", false, func(data string) error {
			b = true
			if len(data) == 0 {
				return nil
			}
			return c.scAdd(sc, width_edge, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "player", false, func(data string) error {
			b = true
			if len(data) == 0 {
				return nil
			}
			return c.scAdd(sc, width_player, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if !b {
			if err := c.paramValue(is, sc, "value",
				width_value, VT_Float, 2, true); err != nil {
				return err
			}
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) sprPriority(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*sprPriority)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			sprPriority_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "layerno",
			sprPriority_layerno, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value",
			sprPriority_value, VT_Int, 1, false)
	})
	return *ret, err
}

func (c *Compiler) varSetSub(is IniSection,
	sc *StateControllerBase, rd OpCode, oc OpCode) error {
	b, v, fv := false, false, false
	var value string
	if err := c.stateParam(is, "value", false, func(data string) error {
		b = true
		value = data
		return nil
	}); err != nil {
		return err
	}
	if b {
		var ve BytecodeExp
		if err := c.stateParam(is, "v", false, func(data string) (err error) {
			v = true
			ve, err = c.fullExpression(&data, VT_Int)
			return
		}); err != nil {
			return err
		}
		if !v {
			if err := c.stateParam(is, "fv", false, func(data string) (err error) {
				fv = true
				ve, err = c.fullExpression(&data, VT_Int)
				return
			}); err != nil {
				return err
			}
		}
		if v || fv {
			if oc == OC_st_var {
				if v {
					oc = OC_st_var
				} else {
					oc = OC_st_fvar
				}
			} else {
				if v {
					oc = OC_st_varadd
				} else {
					oc = OC_st_fvaradd
				}
			}
			var vt ValueType
			if v {
				vt = VT_Int
			} else {
				vt = VT_Float
			}
			in := value
			be, err := c.fullExpression(&in, vt)
			if err != nil {
				return Error(value + "\n" + "value: " + err.Error())
			}
			ve.append(be...)
			if rd != OC_rdreset {
				var tmp BytecodeExp
				tmp.appendI32Op(OC_nordrun, int32(len(ve)))
				ve.append(OC_st_, oc)
				ve = append(tmp, ve...)
				tmp = nil
				tmp.appendI32Op(rd, int32(len(ve)))
				ve = append(tmp, ve...)
			} else {
				ve.append(OC_st_, oc)
			}
			sc.add(varSet_, sc.beToExp(ve))
		}
		return nil
	}
	sys := false
	set := func(data string) error {
		data = strings.TrimSpace(data)
		if data[0] != '(' {
			return Error("Missing '('")
		}
		var be BytecodeExp
		c.token = c.tokenizer(&data)
		bv, err := c.expValue(&be, &data, false)
		if err != nil {
			return err
		}
		if !bv.IsNone() {
			be.appendValue(bv)
		}
		if oc == OC_st_var {
			if sys {
				if v {
					oc = OC_st_sysvar
				} else {
					oc = OC_st_sysfvar
				}
			} else {
				if v {
					oc = OC_st_var
				} else {
					oc = OC_st_fvar
				}
			}
		} else {
			if sys {
				if v {
					oc = OC_st_sysvaradd
				} else {
					oc = OC_st_sysfvaradd
				}
			} else {
				if v {
					oc = OC_st_varadd
				} else {
					oc = OC_st_fvaradd
				}
			}
		}
		if len(c.token) == 0 || c.token[len(c.token)-1] != '=' {
			idx := strings.Index(data, "=")
			if idx < 0 {
				return Error("Missing '='")
			}
			data = data[idx+1:]
		}
		var vt ValueType
		if v {
			vt = VT_Int
		} else {
			vt = VT_Float
		}
		ve := be
		be, err = c.fullExpression(&data, vt)
		if err != nil {
			return err
		}
		ve.append(be...)
		if rd != OC_rdreset {
			var tmp BytecodeExp
			tmp.appendI32Op(OC_nordrun, int32(len(ve)))
			ve.append(OC_st_, oc)
			ve = append(tmp, ve...)
			tmp = nil
			tmp.appendI32Op(rd, int32(len(ve)))
			ve = append(tmp, ve...)
		} else {
			ve.append(OC_st_, oc)
		}
		sc.add(varSet_, sc.beToExp(ve))
		return nil
	}
	if err := c.stateParam(is, "var", false, func(data string) error {
		if data[0] != 'v' {
			return Error(data[:3] + "'v' is not lowercase")
		}
		b = true
		v = true
		return set(data[3:])
	}); err != nil {
		return err
	}
	if b {
		return nil
	}
	if err := c.stateParam(is, "fvar", false, func(data string) error {
		if rd == OC_rdreset && data[0] != 'f' {
			return Error(data[:4] + "'f' is not lowercase")
		}
		b = true
		fv = true
		return set(data[4:])
	}); err != nil {
		return err
	}
	if b {
		return nil
	}
	if err := c.stateParam(is, "sysvar", false, func(data string) error {
		if data[3] != 'v' {
			return Error(data[:6] + "'v' is not lowercase")
		}
		b = true
		v = true
		sys = true
		return set(data[6:])
	}); err != nil {
		return err
	}
	if b {
		return nil
	}
	if err := c.stateParam(is, "sysfvar", false, func(data string) error {
		b = true
		fv = true
		sys = true
		return set(data[7:])
	}); err != nil {
		return err
	}
	if b {
		return nil
	}
	return Error("Value parameter not specified")
}

func (c *Compiler) varSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*varSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.varSetSub(is, sc, OC_rdreset, OC_st_var)
	})
	return *ret, err
}

func (c *Compiler) varAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*varSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.varSetSub(is, sc, OC_rdreset, OC_st_varadd)
	})
	return *ret, err
}

func (c *Compiler) parentVarSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*varSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.varSetSub(is, sc, OC_parent, OC_st_var)
	})
	return *ret, err
}

func (c *Compiler) parentVarAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*varSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.varSetSub(is, sc, OC_parent, OC_st_varadd)
	})
	return *ret, err
}

func (c *Compiler) rootVarSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*varSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.varSetSub(is, sc, OC_root, OC_st_var)
	})
	return *ret, err
}

func (c *Compiler) rootVarAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*varSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.varSetSub(is, sc, OC_root, OC_st_varadd)
	})
	return *ret, err
}

func (c *Compiler) turn(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*turn)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			turn_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(turn_, nil)
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetFacing(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetFacing)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetFacing_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetFacing_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "index",
			targetFacing_index, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetFacing_value, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetBind(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetBind)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetBind_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetBind_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "index",
			targetBind_index, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			targetBind_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pos",
			targetBind_pos, VT_Float, 3, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) bindToTarget(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*bindToTarget)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			bindToTarget_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			bindToTarget_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "index",
			bindToTarget_index, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			bindToTarget_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "pos", false, func(data string) error {
			be, err := c.argExpression(&data, VT_Float)
			if err != nil {
				return err
			}
			exp := sc.beToExp(be)
			if c.token != "," {
				sc.add(bindToTarget_pos, exp)
				return nil
			}
			if be, err = c.argExpression(&data, VT_Float); err != nil {
				return err
			}
			exp, data = append(exp, be), strings.TrimSpace(data)
			if c.token != "," || len(data) == 0 {
				sc.add(bindToTarget_pos, exp)
				return nil
			}
			var hmf HMF
			switch data[0] {
			case 'H', 'h':
				hmf = HMF_H
			case 'M', 'm':
				hmf = HMF_M
			case 'F', 'f':
				hmf = HMF_F
			default:
				return Error("Invalid BindToTarget position type: " + data)
			}
			sc.add(bindToTarget_pos, append(exp, sc.iToExp(int32(hmf))...))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "posz",
			bindToTarget_posz, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetLifeAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetLifeAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetLifeAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetLifeAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "index",
			targetLifeAdd_index, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "absolute",
			targetLifeAdd_absolute, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "kill",
			targetLifeAdd_kill, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "dizzy",
			targetLifeAdd_dizzy, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "redlife",
			targetLifeAdd_redlife, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetLifeAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetState(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetState)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetState_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetState_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "index",
			targetState_index, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetState_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetVelSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetVelSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetVelSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetVelSet_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "index",
			targetVelSet_index, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "x",
			targetVelSet_x, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "y",
			targetVelSet_y, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "z",
			targetVelSet_z, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetVelAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetVelAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetVelAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetVelAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "index",
			targetVelAdd_index, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "x",
			targetVelAdd_x, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "y",
			targetVelAdd_y, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "z",
			targetVelAdd_z, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetPowerAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetPowerAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetPowerAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetPowerAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "index",
			targetPowerAdd_index, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetPowerAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetDrop(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetDrop)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetDrop_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "excludeid",
			targetDrop_excludeid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "keepone",
			targetDrop_keepone, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) lifeAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*lifeAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			lifeAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "absolute",
			lifeAdd_absolute, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "kill",
			lifeAdd_kill, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			lifeAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) lifeSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*lifeSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			lifeSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", lifeSet_value, VT_Int, 1, true)
	})
	return *ret, err
}

func (c *Compiler) powerAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*powerAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			powerAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", powerAdd_value, VT_Int, 1, true)
	})
	return *ret, err
}

func (c *Compiler) powerSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*powerSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			powerSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", powerSet_value, VT_Int, 1, true)
	})
	return *ret, err
}

func (c *Compiler) hitVelSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*hitVelSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitVelSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "x",
			hitVelSet_x, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "y",
			hitVelSet_y, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "z",
			hitVelSet_z, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) screenBound(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*screenBound)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			screenBound_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "value", false, func(data string) error {
			b = true
			return c.scAdd(sc, screenBound_value, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(screenBound_value, sc.iToExp(0)) // In Mugen everything defaults to 0/false if you don't specify any parameters
		}
		b = false
		if err := c.stateParam(is, "movecamera", false, func(data string) error {
			b = true
			return c.scAdd(sc, screenBound_movecamera, data, VT_Bool, 2)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(screenBound_movecamera, append(sc.iToExp(0), sc.iToExp(0)...))
		}
		if err := c.stateParam(is, "stagebound", false, func(data string) error {
			return c.scAdd(sc, screenBound_stagebound, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) posFreeze(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*posFreeze)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			posFreeze_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "value", false, func(data string) error {
			b = true
			return c.scAdd(sc, posFreeze_value, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(posFreeze_value, sc.iToExp(1))
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) envShake(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*envShake)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "time",
			envShake_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ampl",
			envShake_ampl, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "phase",
			envShake_phase, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "freq",
			envShake_freq, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "mul",
			envShake_mul, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "dir",
			envShake_dir, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) hitOverride(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*hitOverride)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitOverride_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "attr", false, func(data string) error {
			attr, err := c.attr(data, false)
			if err != nil {
				return err
			}
			sc.add(hitOverride_attr, sc.iToExp(attr))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "slot",
			hitOverride_slot, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stateno",
			hitOverride_stateno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			hitOverride_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "forceair",
			hitOverride_forceair, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "keepstate",
			hitOverride_keepstate, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "forceguard",
			hitOverride_forceguard, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "guardflag", false, func(data string) error {
			return c.parseHitFlag(sc, hitOverride_guardflag, data)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "guardflag.not", false, func(data string) error {
			return c.parseHitFlag(sc, hitOverride_guardflag_not, data)
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) pause(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*pause)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			pause_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			pause_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "movetime",
			pause_movetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pausebg",
			pause_pausebg, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "endcmdbuftime",
			pause_endcmdbuftime, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) superPause(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*superPause)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			superPause_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			superPause_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "movetime",
			superPause_movetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pausebg",
			superPause_pausebg, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "endcmdbuftime",
			superPause_endcmdbuftime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "darken",
			superPause_darken, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "anim", false, func(data string) error {
			prefix := c.getDataPrefix(&data, true)
			return c.scAdd(sc, superPause_anim, data, VT_Int, 1,
				sc.beToExp(BytecodeExp(prefix))...)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pos",
			superPause_pos, VT_Float, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "p2defmul",
			superPause_p2defmul, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "poweradd",
			superPause_poweradd, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "unhittable",
			superPause_unhittable, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "sound", false, func(data string) error {
			prefix := c.getDataPrefix(&data, true)
			return c.scAdd(sc, superPause_sound, data, VT_Int, 2,
				sc.beToExp(BytecodeExp(prefix))...)
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) trans(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*trans)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			trans_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramTrans(is, sc, "", trans_trans, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) playerPush(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*playerPush)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			playerPush_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		any := false
		if err := c.stateParam(is, "value", false, func(data string) error {
			any = true
			return c.scAdd(sc, playerPush_value, data, VT_Bool, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "priority", false, func(data string) error {
			any = true
			return c.scAdd(sc, playerPush_priority, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !any {
			return Error("Must specify at least one PlayerPush parameter")
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) stateTypeSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*stateTypeSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			stateTypeSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		statetype := func(data string) error {
			if len(data) == 0 {
				return Error("statetype not specified")
			}
			var st StateType
			switch strings.ToLower(data)[0] {
			case 's':
				st = ST_S
			case 'c':
				st = ST_C
			case 'a':
				st = ST_A
			case 'l':
				st = ST_L
			default:
				return Error("Invalid statetype: " + data)
			}
			sc.add(stateTypeSet_statetype, sc.iToExp(int32(st)))
			return nil
		}
		b := false
		if err := c.stateParam(is, "statetype", false, func(data string) error {
			b = true
			return statetype(data)
		}); err != nil {
			return err
		}
		if !b {
			if err := c.stateParam(is, "value", false, func(data string) error {
				return statetype(data)
			}); err != nil {
				return err
			}
		}
		if err := c.stateParam(is, "movetype", false, func(data string) error {
			if len(data) == 0 {
				return Error("movetype not specified")
			}
			var mt MoveType
			switch strings.ToLower(data)[0] {
			case 'i':
				mt = MT_I
			case 'a':
				mt = MT_A
			case 'h':
				mt = MT_H
			default:
				return Error("Invalid movetype: " + data)
			}
			sc.add(stateTypeSet_movetype, sc.iToExp(int32(mt)))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "physics", false, func(data string) error {
			if len(data) == 0 {
				return Error("physics not specified")
			}
			var st StateType
			switch strings.ToLower(data)[0] {
			case 's':
				st = ST_S
			case 'c':
				st = ST_C
			case 'a':
				st = ST_A
			case 'n':
				st = ST_N
			default:
				return Error("Invalid physics type: " + data)
			}
			sc.add(stateTypeSet_physics, sc.iToExp(int32(st)))
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) angleDraw(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*angleDraw)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			angleDraw_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			angleDraw_value, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xangle",
			angleDraw_x, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yangle",
			angleDraw_y, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "scale",
			angleDraw_scale, VT_Float, 2, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) angleSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*angleSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			angleSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			angleSet_value, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xangle",
			angleSet_x, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yangle",
			angleSet_y, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) angleAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*angleAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			angleAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			angleAdd_value, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xangle",
			angleAdd_x, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yangle",
			angleAdd_y, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) angleMul(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*angleMul)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			angleMul_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			angleMul_value, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xangle",
			angleMul_x, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yangle",
			angleMul_y, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) envColor(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*envColor)(sc), c.stateSec(is, func() error {
		if err := c.stateParam(is, "value", false, func(data string) error {
			bes, err := c.exprs(data, VT_Int, 3)
			if err != nil {
				return err
			}
			if len(bes) < 3 {
				return Error("Not enough arguments in EnvColor value")
			}
			sc.add(envColor_value, bes)
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			envColor_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "under",
			envColor_under, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) displayToClipboardSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "redirectid",
		displayToClipboard_redirectid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.stateParam(is, "params", false, func(data string) error {
		bes, err := c.exprs(data, VT_SFalse, 100000)
		if err != nil {
			return err
		}
		sc.add(displayToClipboard_params, bes)
		return nil
	}); err != nil {
		return err
	}
	b := false
	if err := c.stateParam(is, "text", false, func(data string) error {
		b = true
		_else := false
		if len(data) >= 2 && data[0] == '"' {
			if i := strings.Index(data[1:], "\""); i >= 0 {
				data, _ = strconv.Unquote(data)
			} else {
				_else = true
			}
		} else {
			_else = true
		}
		if _else {
			return Error("Text not enclosed in \"")
		}
		sc.add(displayToClipboard_text,
			sc.iToExp(int32(sys.stringPool[c.playerNo].Add(data))))
		return nil
	}); err != nil {
		return err
	}
	if !b {
		return Error("Text parameter not specified")
	}
	return nil
}

func (c *Compiler) displayToClipboard(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*displayToClipboard)(sc), c.stateSec(is, func() error {
		return c.displayToClipboardSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) appendToClipboard(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*appendToClipboard)(sc), c.stateSec(is, func() error {
		return c.displayToClipboardSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) clearClipboard(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*clearClipboard)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			clearClipboard_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(clearClipboard_, nil)
		return nil
	})
	return *ret, err
}

func (c *Compiler) makeDust(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*makeDust)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			makeDust_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "spacing", false, func(data string) error {
			return c.scAdd(sc, makeDust_spacing, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "pos", false, func(data string) error {
			return c.scAdd(sc, makeDust_pos, data, VT_Float, 3)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pos2",
			makeDust_pos2, VT_Float, 3, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) attackDist(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*attackDist)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			attackDist_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			attackDist_x, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "width",
			attackDist_x, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "height",
			attackDist_y, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "depth",
			attackDist_z, VT_Float, 2, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) attackMulSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*attackMulSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			attackMulSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		any := false
		if err := c.stateParam(is, "value", false, func(data string) error {
			any = true
			return c.scAdd(sc, attackMulSet_value, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "damage", false, func(data string) error {
			any = true
			return c.scAdd(sc, attackMulSet_damage, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "redlife", false, func(data string) error {
			any = true
			return c.scAdd(sc, attackMulSet_redlife, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "dizzypoints", false, func(data string) error {
			any = true
			return c.scAdd(sc, attackMulSet_dizzypoints, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "guardpoints", false, func(data string) error {
			any = true
			return c.scAdd(sc, attackMulSet_guardpoints, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if !any {
			return Error("Must specify at least one AttackMulSet multiplier")
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) defenceMulSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*defenceMulSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(
			is, sc, "redirectid",
			defenceMulSet_redirectid, VT_Int, 1, false,
		); err != nil {
			return err
		}

		if err := c.paramValue(
			is, sc, "value",
			defenceMulSet_value, VT_Float, 1, true,
		); err != nil {
			return err
		}

		if err := c.stateParam(is, "multype", false, func(data string) error {
			var mulType = Atoi(strings.TrimSpace(data))

			if mulType >= 0 && mulType <= 1 {
				sc.add(defenceMulSet_mulType, sc.iToExp(mulType))
				return nil
			} else {
				return Error(`Invalid "mulType" value.`)
			}
		}); err != nil {
			return err
		}

		if err := c.paramValue(
			is, sc, "onhit",
			defenceMulSet_onHit, VT_Bool, 1, false,
		); err != nil {
			return err
		}

		return nil
	})
	return *ret, err
}

func (c *Compiler) fallEnvShake(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*fallEnvShake)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			fallEnvShake_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(fallEnvShake_, nil)
		return nil
	})
	return *ret, err
}

func (c *Compiler) hitFallDamage(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*hitFallDamage)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitFallDamage_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(hitFallDamage_, nil)
		return nil
	})
	return *ret, err
}

func (c *Compiler) hitFallVel(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*hitFallVel)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitFallVel_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(hitFallVel_, nil)
		return nil
	})
	return *ret, err
}

func (c *Compiler) hitFallSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*hitFallSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitFallSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "value", false, func(data string) error {
			b = true
			return c.scAdd(sc, hitFallSet_value, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(hitFallSet_value, sc.iToExp(-1))
		}
		if err := c.paramValue(is, sc, "xvel",
			hitFallSet_xvel, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yvel",
			hitFallSet_yvel, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) varRangeSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*varRangeSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varRangeSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "first",
			varRangeSet_first, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "last",
			varRangeSet_last, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "value", false, func(data string) error {
			b = true
			return c.scAdd(sc, varRangeSet_value, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			if err := c.stateParam(is, "fvalue", false, func(data string) error {
				b = true
				return c.scAdd(sc, varRangeSet_fvalue, data, VT_Float, 1)
			}); err != nil {
				return err
			}
			if !b {
				return Error("VarRangeSet value parameter not specified")
			}
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) remapPal(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*remapPal)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			remapPal_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "source",
			remapPal_source, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "dest",
			remapPal_dest, VT_Int, 2, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) stopSnd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*stopSnd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			stopSnd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "channel",
			stopSnd_channel, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) sndPan(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*sndPan)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			sndPan_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "channel",
			sndPan_channel, VT_Int, 1, true); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pan",
			sndPan_pan, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "abspan",
			sndPan_abspan, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) varRandom(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*varRandom)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			varRandom_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "v",
			varRandom_v, VT_Int, 1, true); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "range",
			varRandom_range, VT_Int, 2, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) gravity(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*gravity)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			gravity_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(gravity_, nil)
		return nil
	})
	return *ret, err
}

func (c *Compiler) bindToParentSub(is IniSection,
	sc *StateControllerBase) error {
	if err := c.paramValue(is, sc, "redirectid",
		bindToParent_redirectid, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "time",
		bindToParent_time, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "facing",
		bindToParent_facing, VT_Int, 1, false); err != nil {
		return err
	}
	if err := c.paramValue(is, sc, "pos",
		bindToParent_pos, VT_Float, 3, false); err != nil {
		return err
	}
	return nil
}

func (c *Compiler) bindToParent(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*bindToParent)(sc), c.stateSec(is, func() error {
		return c.bindToParentSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) bindToRoot(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*bindToRoot)(sc), c.stateSec(is, func() error {
		return c.bindToParentSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) removeExplod(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*removeExplod)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			removeExplod_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "id", false, func(data string) error {
			b = true
			return c.scAdd(sc, removeExplod_id, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "index", false, func(data string) error {
			return c.scAdd(sc, removeExplod_index, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(removeExplod_id, sc.iToExp(-1))
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) explodBindTime(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*explodBindTime)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			explodBindTime_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			explodBindTime_id, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "time", false, func(data string) error {
			b = true
			return c.scAdd(sc, explodBindTime_time, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			if err := c.paramValue(is, sc, "value",
				explodBindTime_time, VT_Int, 1, false); err != nil {
				return err
			}
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) moveHitReset(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*moveHitReset)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			moveHitReset_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		sc.add(moveHitReset_, nil)
		return nil
	})
	return *ret, err
}

func (c *Compiler) hitAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*hitAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			hitAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			hitAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) offset(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*offset)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			offset_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "x",
			offset_x, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "y",
			offset_y, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) victoryQuote(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*victoryQuote)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			victoryQuote_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			victoryQuote_value, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) zoom(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*zoom)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "pos",
			zoom_pos, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "scale",
			zoom_scale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "lag",
			zoom_lag, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camerabound",
			zoom_camerabound, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			zoom_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stagebound",
			zoom_stagebound, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) forceFeedback(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*forceFeedback)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			forceFeedback_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "waveform", false, func(data string) error {
			if len(data) == 0 {
				return Error("waveform not specified")
			}
			if data[0] == '"' {
				data = data[1 : len(data)-1]
			}
			var wf int32
			switch strings.ToLower(data) {
			case "sine":
				wf = 0
			case "square":
				wf = 1
			case "sinesquare":
				wf = 2
			case "off":
				wf = -1
			default:
				return Error("Invalid waveform: " + data)
			}
			sc.add(forceFeedback_waveform, sc.iToExp(wf))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			forceFeedback_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "freq",
			forceFeedback_freq, VT_Float, 4, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ampl",
			forceFeedback_ampl, VT_Float, 4, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "self",
			forceFeedback_self, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) assertInput(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*assertInput)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			assertInput_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		foo := func(data string) error {
			switch data {
			case "U":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_PU)))
			case "D":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_PD)))
			case "L":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_PL)))
			case "R":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_PR)))
			case "a":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_A)))
			case "b":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_B)))
			case "c":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_C)))
			case "x":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_X)))
			case "y":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_Y)))
			case "z":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_Z)))
			case "s":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_S)))
			case "d":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_D)))
			case "w":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_W)))
			case "m":
				sc.add(assertInput_flag, sc.iToExp(int32(IB_M)))
			case "B":
				sc.add(assertInput_flag_B, nil)
			case "F":
				sc.add(assertInput_flag_F, nil)
			default:
				return Error("Invalid AssertInput flag: " + data)
			}
			return nil
		}
		// Flag
		flagSet := false
		if err := c.stateParam(is, "flag", false, func(data string) error {
			flagSet = true
			return foo(data)
		}); err != nil {
			return err
		}
		// Flag2-8
		for i := 2; i <= 8; i++ {
			key := fmt.Sprintf("flag%d", i)
			if err := c.stateParam(is, key, false, func(data string) error {
				flagSet = true
				return foo(data)
			}); err != nil {
				return err
			}
		}
		if !flagSet {
			return Error("Must specify at least one AssertInput flag")
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) dialogue(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*dialogue)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			dialogue_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "hidebars",
			dialogue_hidebars, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "force",
			dialogue_force, VT_Bool, 1, false); err != nil {
			return err
		}
		var keys []int
		r, _ := regexp.Compile("^text[0-9]+$")
		for k := range is {
			if r.MatchString(k) {
				re := regexp.MustCompile("[0-9]+")
				submatchall := re.FindAllString(k, -1)
				if len(submatchall) == 1 {
					keys = append(keys, int(Atoi(submatchall[0])))
				}
			}
		}
		sort.Ints(keys)
		for _, key := range keys {
			if err := c.stateParam(is, fmt.Sprintf("text%v", key), false, func(data string) error {
				if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
					return Error("Not enclosed in \"")
				}
				sc.add(dialogue_text, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
				return nil
			}); err != nil {
				return err
			}
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) dizzyPointsAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*dizzyPointsAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			dizzyPointsAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "absolute",
			dizzyPointsAdd_absolute, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			dizzyPointsAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) dizzyPointsSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*dizzyPointsSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			dizzyPointsSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", dizzyPointsSet_value, VT_Int, 1, true)
	})
	return *ret, err
}

func (c *Compiler) dizzySet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*dizzySet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			dizzySet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", dizzySet_value, VT_Bool, 1, true)
	})
	return *ret, err
}

func (c *Compiler) guardBreakSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*guardBreakSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			guardBreakSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", guardBreakSet_value, VT_Bool, 1, true)
	})
	return *ret, err
}

func (c *Compiler) guardPointsAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*guardPointsAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			guardPointsAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "absolute",
			guardPointsAdd_absolute, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			guardPointsAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) guardPointsSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*guardPointsSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			guardPointsSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", guardPointsSet_value, VT_Int, 1, true)
	})
	return *ret, err
}

func (c *Compiler) lifebarAction(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*lifebarAction)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			lifebarAction_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "top",
			lifebarAction_top, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "timemul",
			lifebarAction_timemul, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			lifebarAction_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "anim", false, func(data string) error {
			prefix := c.getDataPrefix(&data, false)
			return c.scAdd(sc, lifebarAction_anim, data, VT_Int, 1,
				sc.beToExp(BytecodeExp(prefix))...)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "spr", false, func(data string) error {
			prefix := c.getDataPrefix(&data, false)
			return c.scAdd(sc, lifebarAction_spr, data, VT_Int, 2,
				sc.beToExp(BytecodeExp(prefix))...)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "snd", false, func(data string) error {
			prefix := c.getDataPrefix(&data, false)
			return c.scAdd(sc, lifebarAction_snd, data, VT_Int, 2,
				sc.beToExp(BytecodeExp(prefix))...)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "text", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Text not enclosed in \"")
			}
			sc.add(lifebarAction_text, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) loadFile(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*loadFile)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			loadFile_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "path", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Path not enclosed in \"")
			}
			sc.add(loadFile_path, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramSaveData(is, sc, loadFile_saveData); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}
func (c *Compiler) loadState(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*loadState)(sc), c.stateSec(is, func() error {
		sc.add(loadState_, nil)
		return nil
	})
	return *ret, err
}

// TODO: Remove boilderplate from the Map's Compiler.
func (c *Compiler) mapSetSub(is IniSection, sc *StateControllerBase) error {
	err := c.stateSec(is, func() error {
		assign := false
		var mapParam, mapName, value string
		if err := c.paramValue(is, sc, "redirectid",
			mapSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "map", true, func(data string) error {
			mapParam = data
			// CNS: See if map parameter is ini-style or if it's an assign
			ia := strings.Index(mapParam, "=")
			if ia > 0 {
				if strings.ToLower(SplitAndTrim(mapParam, "=")[0]) == "map" {
					mapParam = strings.TrimSpace(mapParam[ia+1:])
				} else {
					mapParam = strings.TrimSpace(mapParam[3:])
					assign = true
				}
			} else if !strings.HasPrefix(mapParam, "\"") {
				return Error("Missing '='")
			}
			return nil
		}); err != nil {
			return err
		}
		if len(mapParam) > 0 {
			if assign {
				if err := c.checkOpeningParenthesis(&mapParam); err != nil {
					return err
				}
				mapName = c.token
				c.token = c.tokenizer(&mapParam)
				if err := c.checkClosingParenthesis(); err != nil {
					return err
				}
				c.token = c.tokenizer(&mapParam)
				if c.token == "=" || c.token == ":=" {
					value = strings.TrimSpace(mapParam)
				} else {
					return Error("Invalid operator: " + c.token)
				}
			} else {
				b := false
				if err := c.stateParam(is, "value", false, func(data string) error {
					b = true
					value = data
					return nil
				}); err != nil {
					return err
				}
				if b {
					if len(mapParam) < 2 || mapParam[0] != '"' || mapParam[len(mapParam)-1] != '"' {
						return Error("Not enclosed in \"")
					}
					mapName = mapParam[1 : len(mapParam)-1]
				}
			}
			if len(value) > 0 {
				sc.add(mapSet_mapArray, sc.beToExp(BytecodeExp(mapName)))
				c.scAdd(sc, mapSet_value, value, VT_Float, 1)
			}
		}
		return nil
	})
	return err
}

func (c *Compiler) mapSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "0", VT_Int, 1)

	return *ret, err
}

func (c *Compiler) mapAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "1", VT_Int, 1)

	return *ret, err
}

func (c *Compiler) parentMapSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "2", VT_Int, 1)

	return *ret, err
}

func (c *Compiler) parentMapAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "3", VT_Int, 1)

	return *ret, err
}

func (c *Compiler) rootMapSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "4", VT_Int, 1)

	return *ret, err
}

func (c *Compiler) rootMapAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "5", VT_Int, 1)

	return *ret, err
}

func (c *Compiler) teamMapSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "6", VT_Int, 1)

	return *ret, err
}

func (c *Compiler) teamMapAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*mapSet)(sc), c.stateSec(is, func() error {
		if err := c.mapSetSub(is, sc); err != nil {
			return err
		}
		return nil
	})
	c.scAdd(sc, mapSet_type, "7", VT_Int, 1)

	return *ret, err
}

func (c *Compiler) matchRestart(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*matchRestart)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "reload",
			matchRestart_reload, VT_Bool, MaxPlayerNo, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "stagedef", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Stagedef not enclosed in \"")
			}
			sc.add(matchRestart_stagedef, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p1def", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("P1def not enclosed in \"")
			}
			sc.add(matchRestart_p1def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p2def", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("P2def not enclosed in \"")
			}
			sc.add(matchRestart_p2def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p3def", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("P3def not enclosed in \"")
			}
			sc.add(matchRestart_p3def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p4def", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("P4def not enclosed in \"")
			}
			sc.add(matchRestart_p4def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p5def", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("P5def not enclosed in \"")
			}
			sc.add(matchRestart_p5def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p6def", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("P6def not enclosed in \"")
			}
			sc.add(matchRestart_p6def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p7def", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("P7def not enclosed in \"")
			}
			sc.add(matchRestart_p7def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "p8def", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("P8def not enclosed in \"")
			}
			sc.add(matchRestart_p8def, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) playBgm(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*playBgm)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			playBgm_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "bgm", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("BGM not enclosed in \"")
			}
			sc.add(playBgm_bgm, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "volume",
			playBgm_volume, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loop",
			playBgm_loop, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loopstart",
			playBgm_loopstart, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loopend",
			playBgm_loopend, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "startposition",
			playBgm_startposition, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loopcount",
			playBgm_loopcount, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "freqmul",
			playBgm_freqmul, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) modifyBGCtrl(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*modifyBGCtrl)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "id",
			modifyBGCtrl_id, VT_Int, 1, true); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			modifyBGCtrl_time, VT_Int, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			modifyBGCtrl_value, VT_Int, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "x",
			modifyBGCtrl_x, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "y",
			modifyBGCtrl_y, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "source",
			modifyBGCtrl_source, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "dest",
			modifyBGCtrl_dest, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "add",
			modifyBGCtrl_add, VT_Int, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "mul",
			modifyBGCtrl_mul, VT_Int, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "sinadd",
			modifyBGCtrl_sinadd, VT_Int, 4, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "sinmul",
			modifyBGCtrl_sinmul, VT_Int, 4, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "sincolor",
			modifyBGCtrl_sincolor, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "sinhue",
			modifyBGCtrl_sinhue, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "invertall",
			modifyBGCtrl_invertall, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "invertblend",
			modifyBGCtrl_invertblend, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "color",
			modifyBGCtrl_color, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "hue",
			modifyBGCtrl_hue, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) modifyBGCtrl3d(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*modifyBGCtrl3d)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "id",
			modifyBGCtrl3d_ctrlid, VT_Int, 1, true); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "time",
			modifyBGCtrl3d_time, VT_Int, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			modifyBGCtrl3d_value, VT_Int, 3, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) modifySnd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*modifySnd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			modifySnd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "channel",
			modifySnd_channel, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pan",
			modifySnd_pan, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "abspan",
			modifySnd_abspan, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "volume",
			modifySnd_volume, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "volumescale",
			modifySnd_volumescale, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "freqmul",
			modifySnd_freqmul, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "priority",
			modifySnd_priority, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loopstart",
			modifySnd_loopstart, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loopend",
			modifySnd_loopend, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "position",
			modifySnd_position, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loop",
			modifySnd_loop, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loopcount",
			modifySnd_loopcount, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stopongethit",
			modifySnd_stopongethit, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stoponchangestate",
			modifySnd_stoponchangestate, VT_Bool, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) modifyBgm(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*modifyBgm)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "volume",
			modifyBgm_volume, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loopstart",
			modifyBgm_loopstart, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "loopend",
			modifyBgm_loopend, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "freqmul",
			modifyBgm_freqmul, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "position",
			modifyBgm_position, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) printToConsole(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*printToConsole)(sc), c.stateSec(is, func() error {
		return c.displayToClipboardSub(is, sc)
	})
	return *ret, err
}

func (c *Compiler) redLifeAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*redLifeAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			redLifeAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "absolute",
			redLifeAdd_absolute, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			redLifeAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) redLifeSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*redLifeSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			redLifeSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", redLifeSet_value, VT_Int, 1, true)
	})
	return *ret, err
}

func (c *Compiler) remapSprite(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*remapSprite)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			remapSprite_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "reset",
			remapSprite_reset, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "preset", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Preset not enclosed in \"")
			}
			sc.add(remapSprite_preset, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "source",
			remapSprite_source, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "dest",
			remapSprite_dest, VT_Int, 2, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) roundTimeAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*roundTimeAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			roundTimeAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			roundTimeAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) roundTimeSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*roundTimeSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			roundTimeSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		return c.paramValue(is, sc, "value", roundTimeSet_value, VT_Int, 1, true)
	})
	return *ret, err
}

func (c *Compiler) saveFile(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*saveFile)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			saveFile_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "path", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Path not enclosed in \"")
			}
			sc.add(saveFile_path, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramSaveData(is, sc, saveFile_saveData); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) saveState(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*saveState)(sc), c.stateSec(is, func() error {
		sc.add(saveState_, nil)
		return nil
	})
	return *ret, err
}

func (c *Compiler) scoreAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*scoreAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			scoreAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			scoreAdd_value, VT_Float, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetDizzyPointsAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetDizzyPointsAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetDizzyPointsAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetDizzyPointsAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "absolute",
			targetDizzyPointsAdd_absolute, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetDizzyPointsAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetGuardPointsAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetGuardPointsAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetGuardPointsAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetGuardPointsAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "absolute",
			targetGuardPointsAdd_absolute, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetGuardPointsAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetRedLifeAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetRedLifeAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetRedLifeAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetRedLifeAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "absolute",
			targetRedLifeAdd_absolute, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetRedLifeAdd_value, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetScoreAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetScoreAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetScoreAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			targetScoreAdd_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			targetScoreAdd_value, VT_Float, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) text(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*text)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			text_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			text_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "removetime",
			text_removetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "layerno",
			text_layerno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "params", false, func(data string) error {
			bes, err := c.exprs(data, VT_SFalse, 100000)
			if err != nil {
				return err
			}
			sc.add(text_params, bes)
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "text", false, func(data string) error {
			_else := false
			if len(data) >= 2 && data[0] == '"' {
				if i := strings.Index(data[1:], "\""); i >= 0 {
					data, _ = strconv.Unquote(data)
				} else {
					_else = true
				}
			} else {
				_else = true
			}
			if _else {
				return Error("Text not enclosed in \"")
			}
			sc.add(text_text, sc.iToExp(int32(sys.stringPool[c.playerNo].Add(data))))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "font", false, func(data string) error {
			prefix := c.getDataPrefix(&data, false)
			fflg := prefix == "f"
			return c.scAdd(sc, text_font, data, VT_Int, 1,
				sc.iToExp(Btoi(fflg))...)
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "localcoord",
			text_localcoord, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "bank",
			text_bank, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "align",
			text_align, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "linespacing",
			text_linespacing, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "textdelay",
			text_textdelay, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pos",
			text_pos, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "velocity",
			text_velocity, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "friction",
			text_friction, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "accel",
			text_accel, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "scale",
			text_scale, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "angle",
			text_angle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.palFXSub(is, sc, "palfx."); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "color",
			text_color, VT_Int, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xshear",
			text_xshear, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) removeText(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*removeText)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			removetext_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "id", false, func(data string) error {
			b = true
			return c.scAdd(sc, removetext_id, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if !b {
			sc.add(removetext_id, sc.iToExp(-1))
		}
		return nil
	})
	return *ret, err
}

// Handles "createPlatform" parameters.
func (c *Compiler) createPlatform(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*text)(sc), c.stateSec(is, func() error {
		var err error

		if err = c.paramValue(
			is, sc,
			"id", createPlatform_id,
			VT_Int, 1, true,
		); err != nil {
			return err
		}

		// Here we check if the string is enclosed in quotes.
		// (Because CNS has no real string support)
		if err = c.stateParam(
			is, "name", false,
			func(data string) error {
				if data[0] != '"' || data[len(data)-1] != '"' {
					return Error(`[name] value in [createPlatform] not enclosed in quotation marks.` +
						"\n" + "Value provided: [" + data + "]",
					)
				}
				sc.add(helper_name, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
				return nil
			},
		); err != nil {
			return err
		}

		if err = c.paramValue(
			is, sc,
			"anim", createPlatform_anim,
			VT_Int, 1, false,
		); err != nil {
			return err
		}

		if err = c.paramValue(
			is, sc,
			"pos", createPlatform_pos,
			VT_Int, 2, true,
		); err != nil {
			return err
		}

		if err = c.paramValue(
			is, sc,
			"size", createPlatform_size,
			VT_Int, 2, true,
		); err != nil {
			return err
		}

		if err = c.paramValue(
			is, sc,
			"offset", createPlatform_offset,
			VT_Int, 2, false,
		); err != nil {
			return err
		}

		if err = c.paramValue(
			is, sc,
			"activeTime", createPlatform_activeTime,
			VT_Int, 1, false,
		); err != nil {
			return err
		}

		return nil
	})
	return *ret, err
}

func (c *Compiler) modifyStageVar(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*modifyStageVar)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "camera.boundleft",
			modifyStageVar_camera_boundleft, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.boundright",
			modifyStageVar_camera_boundright, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.boundhigh",
			modifyStageVar_camera_boundhigh, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.boundlow",
			modifyStageVar_camera_boundlow, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.verticalfollow",
			modifyStageVar_camera_verticalfollow, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.floortension",
			modifyStageVar_camera_floortension, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.tensionhigh",
			modifyStageVar_camera_tensionhigh, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.tensionlow",
			modifyStageVar_camera_tensionlow, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.tension",
			modifyStageVar_camera_tension, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.tensionvel",
			modifyStageVar_camera_tensionvel, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.cuthigh",
			modifyStageVar_camera_cuthigh, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.cutlow",
			modifyStageVar_camera_cutlow, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.startzoom",
			modifyStageVar_camera_startzoom, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.zoomout",
			modifyStageVar_camera_zoomout, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.zoomin",
			modifyStageVar_camera_zoomin, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.zoomindelay",
			modifyStageVar_camera_zoomindelay, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.zoominspeed",
			modifyStageVar_camera_zoominspeed, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.zoomoutspeed",
			modifyStageVar_camera_zoomoutspeed, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.yscrollspeed",
			modifyStageVar_camera_yscrollspeed, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.ytension.enable",
			modifyStageVar_camera_ytension_enable, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.autocenter",
			modifyStageVar_camera_autocenter, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "camera.lowestcap",
			modifyStageVar_camera_lowestcap, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerinfo.leftbound",
			modifyStageVar_playerinfo_leftbound, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerinfo.rightbound",
			modifyStageVar_playerinfo_rightbound, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerinfo.topbound",
			modifyStageVar_playerinfo_topbound, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerinfo.botbound",
			modifyStageVar_playerinfo_botbound, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerinfo.p1startx",
			modifyStageVar_playerinfo_p1startx, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerinfo.p1starty",
			modifyStageVar_playerinfo_p1starty, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerinfo.p2startx",
			modifyStageVar_playerinfo_p2startx, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerinfo.p2starty",
			modifyStageVar_playerinfo_p2starty, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerinfo.p1startz",
			modifyStageVar_playerinfo_p1startz, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerinfo.p2startz",
			modifyStageVar_playerinfo_p2startz, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerinfo.p1facing",
			modifyStageVar_playerinfo_p1facing, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerinfo.p2facing",
			modifyStageVar_playerinfo_p2facing, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "scaling.topz",
			modifyStageVar_scaling_topz, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "scaling.botz",
			modifyStageVar_scaling_botz, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "scaling.topscale",
			modifyStageVar_scaling_topscale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "scaling.botscale",
			modifyStageVar_scaling_botscale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "bound.screenleft",
			modifyStageVar_bound_screenleft, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "bound.screenright",
			modifyStageVar_bound_screenright, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stageinfo.zoffset",
			modifyStageVar_stageinfo_zoffset, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stageinfo.zoffsetlink",
			modifyStageVar_stageinfo_zoffsetlink, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stageinfo.xscale",
			modifyStageVar_stageinfo_xscale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "stageinfo.yscale",
			modifyStageVar_stageinfo_yscale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "shadow.intensity",
			modifyStageVar_shadow_intensity, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "shadow.color",
			modifyStageVar_shadow_color, VT_Int, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "shadow.yscale",
			modifyStageVar_shadow_yscale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "shadow.angle",
			modifyStageVar_shadow_angle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "shadow.xangle",
			modifyStageVar_shadow_xangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "shadow.yangle",
			modifyStageVar_shadow_yangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "shadow.focallength",
			modifyStageVar_shadow_focallength, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramProjection(is, sc, "shadow.projection",
			modifyStageVar_shadow_projection); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "shadow.fade.range",
			modifyStageVar_shadow_fade_range, VT_Int, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "shadow.xshear",
			modifyStageVar_shadow_xshear, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "shadow.offset",
			modifyStageVar_shadow_offset, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "shadow.window",
			modifyStageVar_shadow_window, VT_Float, 4, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "reflection.intensity",
			modifyStageVar_reflection_intensity, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "reflection.yscale",
			modifyStageVar_reflection_yscale, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "reflection.angle",
			modifyStageVar_reflection_angle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "reflection.xangle",
			modifyStageVar_reflection_xangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "reflection.yangle",
			modifyStageVar_reflection_yangle, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "reflection.focallength",
			modifyStageVar_reflection_focallength, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramProjection(is, sc, "reflection.projection",
			modifyStageVar_reflection_projection); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "reflection.xshear",
			modifyStageVar_reflection_xshear, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "reflection.color",
			modifyStageVar_reflection_color, VT_Int, 3, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "reflection.offset",
			modifyStageVar_reflection_offset, VT_Float, 2, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "reflection.window",
			modifyStageVar_reflection_window, VT_Float, 4, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) cameraCtrl(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*cameraCtrl)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "followid",
			cameraCtrl_followid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "view", false, func(data string) error {
			if len(data) == 0 {
				return nil
			}
			switch strings.ToLower(data) {
			case "fighting":
				sc.add(cameraCtrl_view, sc.iToExp(int32(Fighting_View)))
			case "follow":
				sc.add(cameraCtrl_view, sc.iToExp(int32(Follow_View)))
			case "free":
				sc.add(cameraCtrl_view, sc.iToExp(int32(Free_View)))
			default:
				return Error("Invalid view type: " + data)
			}
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pos",
			cameraCtrl_pos, VT_Float, 2, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) height(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*height)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			height_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			height_value, VT_Float, 2, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) depth(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*depth)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			depth_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		b := false
		if err := c.stateParam(is, "edge", false, func(data string) error {
			b = true
			if len(data) == 0 {
				return nil
			}
			return c.scAdd(sc, depth_edge, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "player", false, func(data string) error {
			b = true
			if len(data) == 0 {
				return nil
			}
			return c.scAdd(sc, depth_player, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if !b {
			if err := c.paramValue(is, sc, "value",
				depth_value, VT_Float, 2, true); err != nil {
				return err
			}
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) modifyPlayer(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*modifyPlayer)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			modifyPlayer_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "lifemax",
			modifyPlayer_lifemax, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "powermax",
			modifyPlayer_powermax, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "dizzypointsmax",
			modifyPlayer_dizzypointsmax, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "guardpointsmax",
			modifyPlayer_guardpointsmax, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "teamside",
			modifyPlayer_teamside, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "displayname", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Displayname not enclosed in \"")
			}
			sc.add(modifyPlayer_displayname, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "lifebarname", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Lifebarname not enclosed in \"")
			}
			sc.add(modifyPlayer_lifebarname, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "helperid",
			modifyPlayer_helperid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "helpername", false, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Helpername not enclosed in \"")
			}
			sc.add(modifyPlayer_helpername, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}

		if err := c.paramValue(is, sc, "movehit",
			modifyPlayer_movehit, VT_Int, 1, false); err != nil { // Formerly MoveHitSet
			return err
		}
		if err := c.paramValue(is, sc, "moveguarded",
			modifyPlayer_moveguarded, VT_Int, 1, false); err != nil { // Formerly MoveHitSet
			return err
		}
		if err := c.paramValue(is, sc, "movereversed",
			modifyPlayer_movereversed, VT_Int, 1, false); err != nil { // Formerly MoveHitSet
			return err
		}
		if err := c.paramValue(is, sc, "movecountered",
			modifyPlayer_movecountered, VT_Bool, 1, false); err != nil { // Formerly MoveHitSet
			return err
		}
		if err := c.paramValue(is, sc, "hitpausetime",
			modifyPlayer_hitpausetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "pausemovetime",
			modifyPlayer_pausemovetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "supermovetime",
			modifyPlayer_supermovetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "unhittabletime",
			modifyPlayer_unhittabletime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "attack",
			modifyPlayer_attack, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "defence",
			modifyPlayer_defence, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "alive",
			modifyPlayer_alive, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ailevel",
			modifyPlayer_ailevel, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) assertCommand(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*assertCommand)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			assertCommand_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "name", true, func(data string) error {
			if len(data) < 2 || data[0] != '"' || data[len(data)-1] != '"' {
				return Error("Command name not enclosed in \"")
			}
			sc.add(assertCommand_name, sc.beToExp(BytecodeExp(data[1:len(data)-1])))
			return nil
		}); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "buffer.time",
			assertCommand_buffertime, VT_Int, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) getHitVarSet(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*getHitVarSet)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			getHitVarSet_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "airtype",
			getHitVarSet_airtype, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "animtype",
			getHitVarSet_animtype, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "attr",
			getHitVarSet_attr, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "chainid",
			getHitVarSet_chainid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "ctrltime",
			getHitVarSet_ctrltime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "damage",
			getHitVarSet_damage, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "dizzypoints",
			getHitVarSet_dizzypoints, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "down.recover",
			getHitVarSet_down_recover, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "down.recovertime",
			getHitVarSet_down_recovertime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall",
			getHitVarSet_fall, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.damage",
			getHitVarSet_fall_damage, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.envshake.ampl",
			getHitVarSet_fall_envshake_ampl, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.envshake.freq",
			getHitVarSet_fall_envshake_freq, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.envshake.mul",
			getHitVarSet_fall_envshake_mul, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.envshake.phase",
			getHitVarSet_fall_envshake_phase, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.envshake.time",
			getHitVarSet_fall_envshake_time, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.envshake.dir",
			getHitVarSet_fall_envshake_dir, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.kill",
			getHitVarSet_fall_kill, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.recover",
			getHitVarSet_fall_recover, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.recovertime",
			getHitVarSet_fall_recovertime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.xvel",
			getHitVarSet_fall_xvel, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.yvel",
			getHitVarSet_fall_yvel, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fall.zvel",
			getHitVarSet_fall_zvel, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "fallcount",
			getHitVarSet_fallcount, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "groundtype",
			getHitVarSet_groundtype, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "guardcount",
			getHitVarSet_guardcount, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "guarded",
			getHitVarSet_guarded, VT_Bool, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "guardpoints",
			getHitVarSet_guardpoints, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "hitcount",
			getHitVarSet_hittime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "hitshaketime",
			getHitVarSet_hitshaketime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "hittime",
			getHitVarSet_hittime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "id",
			getHitVarSet_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerno",
			getHitVarSet_playerno, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "redlife",
			getHitVarSet_redlife, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "slidetime",
			getHitVarSet_slidetime, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xvel",
			getHitVarSet_xvel, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "xaccel",
			getHitVarSet_xaccel, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yaccel",
			getHitVarSet_yaccel, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "zaccel",
			getHitVarSet_zaccel, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "yvel",
			getHitVarSet_yvel, VT_Float, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "zvel",
			getHitVarSet_zvel, VT_Float, 1, false); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) groundLevelOffset(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*groundLevelOffset)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			groundLevelOffset_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "value",
			groundLevelOffset_value, VT_Float, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) targetAdd(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*targetAdd)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			targetAdd_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "playerid",
			targetAdd_playerid, VT_Int, 1, true); err != nil {
			return err
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) transformClsn(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*transformClsn)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			transformClsn_redirectid, VT_Int, 1, false); err != nil {
			return err
		}
		any := false
		if err := c.stateParam(is, "scale", false, func(data string) error {
			any = true
			return c.scAdd(sc, transformClsn_scale, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "angle", false, func(data string) error {
			any = true
			return c.scAdd(sc, transformClsn_angle, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if !any {
			return Error("Must specify at least one TransformClsn parameter")
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) transformSprite(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*transformSprite)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			transformSprite_redirectid, VT_Float, 1, false); err != nil {
			return err
		}
		any := false
		if err := c.stateParam(is, "window", false, func(data string) error {
			any = true
			return c.scAdd(sc, transformSprite_window, data, VT_Float, 4)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "xshear", false, func(data string) error {
			any = true
			return c.scAdd(sc, transformSprite_xshear, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "focallength", false, func(data string) error {
			any = true
			return c.scAdd(sc, transformSprite_focallength, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if _, ok := is["projection"]; ok {
			any = true
			if err := c.paramProjection(is, sc, "projection",
				transformSprite_projection); err != nil {
				return err
			}
		}
		if !any {
			return Error("Must specify at least one TransformSprite parameter")
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) modifyStageBG(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	ret, err := (*modifyStageBG)(sc), c.stateSec(is, func() error {
		//if err := c.paramValue(is, sc, "redirectid",
		//	modifyStageBG_redirectid, VT_Int, 1, false); err != nil {
		//	return err
		//}
		if err := c.paramValue(is, sc, "id",
			modifyStageBG_id, VT_Int, 1, false); err != nil {
			return err
		}
		if err := c.paramValue(is, sc, "index",
			modifyStageBG_index, VT_Int, 1, false); err != nil {
			return err
		}
		any := false
		if err := c.stateParam(is, "actionno", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_actionno, data, VT_Int, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "delta.x", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_delta_x, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "delta.y", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_delta_y, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "layerno", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_layerno, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "pos.x", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_pos_x, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "pos.y", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_pos_y, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "spriteno", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_spriteno, data, VT_Int, 2)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "start.x", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_start_x, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "start.y", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_start_y, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "scalestart", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_scalestart, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if _, ok := is["trans"]; ok { // Check if "trans" exists, since you can't set "any" from within paramTrans
			any = true
		}
		if err := c.paramTrans(is, sc, "", modifyStageBG_trans, false); err != nil {
			return err
		}
		if err := c.stateParam(is, "angle", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_angle, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "xangle", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_xangle, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "yangle", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_yangle, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "velocity.x", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_velocity_x, data, VT_Float, 2)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "velocity.y", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_velocity_y, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "xshear", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_xshear, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if err := c.stateParam(is, "focallength", false, func(data string) error {
			any = true
			return c.scAdd(sc, modifyStageBG_focallength, data, VT_Float, 1)
		}); err != nil {
			return err
		}
		if _, ok := is["projection"]; ok {
			any = true
			if err := c.paramProjection(is, sc, "projection",
				modifyStageBG_projection); err != nil {
				return err
			}
		}
		if !any {
			return Error("Must specify at least one ModifyStageBG parameter")
		}
		return nil
	})
	return *ret, err
}

func (c *Compiler) shiftInput(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	inputIndex := func(name string) (int32, error) {
		switch name {
		case "U":
			return 0, nil
		case "D":
			return 1, nil
		case "L":
			return 2, nil
		case "R":
			return 3, nil
		case "a":
			return 4, nil
		case "b":
			return 5, nil
		case "c":
			return 6, nil
		case "x":
			return 7, nil
		case "y":
			return 8, nil
		case "z":
			return 9, nil
		case "s":
			return 10, nil
		case "d":
			return 11, nil
		case "w":
			return 12, nil
		case "m":
			return 13, nil
		case "B", "F":
			return -1, Error("'B' and 'F' symbols not accepted. Must be 'L' or 'R'")
		default:
			// This one is case insensitive
			if strings.ToLower(name) == "none" {
				return -1, nil
			}
			// Unknown
			return -1, Error("Invalid key name: " + name)
		}
	}

	ret, err := (*shiftInput)(sc), c.stateSec(is, func() error {
		if err := c.paramValue(is, sc, "redirectid",
			shiftInput_redirectid, VT_Int, 1, false); err != nil {
			return err
		}

		if err := c.stateParam(is, "input", true, func(srcName string) error {
			srcID, err := inputIndex(srcName)
			if err != nil {
				return err
			}
			sc.add(shiftInput_input, sc.iToExp(srcID))
			return nil
		}); err != nil {
			return err
		}

		if err := c.stateParam(is, "output", true, func(dstName string) error {
			dstID, err := inputIndex(dstName)
			if err != nil {
				return err
			}
			sc.add(shiftInput_output, sc.iToExp(dstID))
			return nil
		}); err != nil {
			return err
		}

		return nil
	})

	return *ret, err
}

// It's just a Null... Has no effect whatsoever.
func (c *Compiler) null(is IniSection, sc *StateControllerBase, _ int8) (StateController, error) {
	return nullStateController, nil
}

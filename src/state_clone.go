package main

import (
	"arena"

	"golang.org/x/exp/maps"
)

type Copyable[T any] interface {
	Clone() T
}

func CopySlice[T any](src, dst *[]T) {
	*(dst) = (*dst)[:0]
	for i := 0; i < len(*src); i++ {
		*(dst) = append(*(dst), (*src)[i])
	}
}

func CopyMap[T comparable, E any](src, dst *map[T]E) {
	for k := range *src {
		(*dst)[k] = (*src)[k]
	}

	for k := range *dst {
		if _, ok := (*src)[k]; !ok {
			delete(*dst, k)
		}
	}
}

// func Copy2DSlice[T any](src, dst *[][]T) {
// 	if len(*dst) < len(*src) {
// 		i := 0
// 		for ; i < len(*dst); i++ {
// 			CopySlice(&(*src)[i], &(*dst)[i])
// 		}
// 		for ; i < len(*src); i++ {
// 			slice := PoolGet((*src)[i]).(*[]T)
// 			*dst = append(*dst, *slice)
// 			CopySlice(&(*src)[i], &(*dst)[i])
// 		}
// 	} else {
// 		(*dst) = (*dst)[0:len(*src)]
// 		for i := 0; i < len(*src); i++ {
// 			CopySlice(&(*src)[i], &(*dst)[i])
// 		}
// 	}
// }

// func DeepCopy2DSlice[T Copyable[T]](src, dst *[][]T) {
// 	if len(*dst) < len(*src) {
// 		i := 0
// 		for ; i < len(*dst); i++ {
// 			DeepCopySlice(&(*src)[i], &(*dst)[i])
// 		}
// 		for ; i < len(*src); i++ {
// 			slice := PoolGet((*src)[i]).(*[]T)
// 			*dst = append(*dst, *slice)
// 			DeepCopySlice(&(*src)[i], &(*dst)[i])
// 		}
// 	} else {
// 		(*dst) = (*dst)[0:len(*src)]
// 		for i := 0; i < len(*src); i++ {
// 			DeepCopySlice(&(*src)[i], &(*dst)[i])
// 		}
// 	}
// }

func DeepCopySlice[T Copyable[T]](src, dst *[]T) {
	if len(*dst) >= len(*src) {
		*(dst) = (*dst)[0:len(*src)]
		for i := 0; i < len(*src); i++ {
			(*dst)[i] = (*src)[i].Clone()
		}
	} else {
		i := 0
		for ; i < len(*dst); i++ {
			(*dst)[i] = (*src)[i].Clone()
		}
		for ; i < len(*src); i++ {
			(*dst) = append(*dst, (*src)[i].Clone())
		}
	}
}

func (a *Animation) Clone(ar *arena.Arena, gsp *GameStatePool) (result *Animation) {
	result = arena.New[Animation](ar)
	*result = *a

	result.frames = *gsp.Get(a.frames).(*[]AnimFrame)
	result.frames = result.frames[:0]
	for i := 0; i < len(a.frames); i++ {
		result.frames = append(result.frames, *a.frames[i].Clone(ar))
	}

	result.interpolate_offset = arena.MakeSlice[int32](ar, len(a.interpolate_offset), len(a.interpolate_offset))
	copy(result.interpolate_offset, a.interpolate_offset)

	result.interpolate_scale = arena.MakeSlice[int32](ar, len(a.interpolate_scale), len(a.interpolate_scale))
	copy(result.interpolate_scale, a.interpolate_scale)

	result.interpolate_angle = arena.MakeSlice[int32](ar, len(a.interpolate_angle), len(a.interpolate_angle))
	copy(result.interpolate_angle, a.interpolate_angle)

	result.interpolate_blend = arena.MakeSlice[int32](ar, len(a.interpolate_blend), len(a.interpolate_blend))
	copy(result.interpolate_blend, a.interpolate_blend)

	return
}

func (af *AnimFrame) Clone(a *arena.Arena) (result *AnimFrame) {
	result = arena.New[AnimFrame](a)
	*result = *af
	result.Clsn1 = arena.MakeSlice[[4]float32](a, len(af.Clsn1), len(af.Clsn1))
	copy(result.Clsn1, af.Clsn1)
	result.Clsn2 = arena.MakeSlice[[4]float32](a, len(af.Clsn2), len(af.Clsn2))
	copy(result.Clsn2, af.Clsn2)
	return
}

func (sp StringPool) Clone(a *arena.Arena, gsp *GameStatePool) (result StringPool) {
	result = sp
	result.List = arena.MakeSlice[string](a, len(sp.List), len(sp.List))
	copy(result.List, sp.List)
	result.Map = *gsp.Get(sp.Map).(*map[string]int)
	maps.Clear(result.Map)

	for k, v := range sp.Map {
		result.Map[k] = v
	}
	return
}

func (b *StateBlock) Clone(a *arena.Arena) (result StateBlock) {
	result = *b
	result.trigger = arena.MakeSlice[OpCode](a, len(b.trigger), len(b.trigger))
	copy(result.trigger, b.trigger)
	if b.elseBlock != nil {
		eb := b.elseBlock.Clone(a)
		result.elseBlock = &eb
	}

	result.forCtrlVar.be = arena.MakeSlice[OpCode](a, len(b.forCtrlVar.be), len(b.forCtrlVar.be))
	copy(result.forCtrlVar.be, b.forCtrlVar.be)

	for i := 0; i < len(b.forExpression); i++ {
		result.forExpression[i] = arena.MakeSlice[OpCode](a, len(b.forExpression[i]), len(b.forExpression[i]))
		copy(result.forExpression[i], b.forExpression[i])
	}

	result.ctrls = arena.MakeSlice[StateController](a, len(b.ctrls), len(b.ctrls))
	copy(result.ctrls, b.ctrls)
	return result
}

func (sb *StateBytecode) Clone(a *arena.Arena) (result StateBytecode) {
	result = *sb
	result.stateDef = arena.MakeSlice[byte](a, len(sb.stateDef), len(sb.stateDef))
	copy(result.stateDef, sb.stateDef)

	result.ctrlsps = arena.MakeSlice[int32](a, len(sb.ctrlsps), len(sb.ctrlsps))
	copy(result.ctrlsps, sb.ctrlsps)
	result.block = sb.block.Clone(a)
	return result
}

func (ghv *GetHitVar) Clone(a *arena.Arena) (result *GetHitVar) {
	result = arena.New[GetHitVar](a)
	*result = *ghv

	// Manually copy references that shallow copy poorly, as needed
	// Pointers, slices, maps, functions, channels etc
	result.targetedBy = arena.MakeSlice[[2]int32](a, len(ghv.targetedBy), len(ghv.targetedBy))
	copy(result.targetedBy, ghv.targetedBy)

	return
}

func (ai AfterImage) Clone(a *arena.Arena, gsp *GameStatePool) (result AfterImage) {
	result = ai

	// Deep copy Animations
	for i := range ai.imgs {
		if ai.imgs[i].anim != nil {
			result.imgs[i].anim = ai.imgs[i].anim.Clone(a, gsp)
		}
	}

	// Deep copy PalFX
	if ai.palfx != nil {
		result.palfx = arena.MakeSlice[*PalFX](a, len(ai.palfx), len(ai.palfx))
		for i := range ai.palfx {
			if ai.palfx[i] != nil {
				result.palfx[i] = ai.palfx[i].Clone(a)
			}
		}
	}

	return
}

func (e *Explod) Clone(a *arena.Arena, gsp *GameStatePool) *Explod {
	if e == nil {
		return nil
	}

	result := &Explod{}
	*result = *e

	if e.anim != nil {
		result.anim = e.anim.Clone(a, gsp)
	}

	if e.palfx != nil {
		result.palfx = e.palfx.Clone(a)
	}

	return result
}

func (p *Projectile) clone(a *arena.Arena, gsp *GameStatePool) *Projectile {
	if p == nil {
		return nil
	}

	result := &Projectile{}
	*result = *p

	if p.ani != nil {
		result.ani = p.ani.Clone(a, gsp)
	}

	if p.aimg.palfx != nil {
		result.aimg.palfx = arena.MakeSlice[*PalFX](a, len(p.aimg.palfx), len(p.aimg.palfx))
		for i := range p.aimg.palfx {
			result.aimg.palfx[i] = p.aimg.palfx[i].Clone(a)
		}
	}

	if p.palfx != nil {
		result.palfx = p.palfx.Clone(a)
	}

	return result
}

func (ss *StateState) Clone(a *arena.Arena) (result StateState) {
	result = *ss
	result.ps = arena.MakeSlice[int32](a, len(ss.ps), len(ss.ps))
	copy(result.ps, ss.ps)
	for i := 0; i < len(ss.hitPauseExecutionToggleFlags); i++ {
		result.hitPauseExecutionToggleFlags[i] = arena.MakeSlice[bool](a, len(ss.hitPauseExecutionToggleFlags[i]), len(ss.hitPauseExecutionToggleFlags[i]))
		copy(result.hitPauseExecutionToggleFlags[i], ss.hitPauseExecutionToggleFlags[i])
	}
	result.sb = ss.sb.Clone(a)
	return result
}

func (c *Char) Clone(a *arena.Arena, gsp *GameStatePool) (result Char) {
	result = Char{}
	result = *c

	if c.anim != nil {
		result.anim = c.anim.Clone(a, gsp)
	}
	if c.animBackup != nil {
		result.animBackup = c.animBackup.Clone(a, gsp)
	}

	// Since curFrame is desynced from anim's state, we must save it as well
	if c.curFrame != nil {
		result.curFrame = c.curFrame.Clone(a)
	} else {
		result.curFrame = nil
	}

	// TODO: Profiling shows this is hotter than it should be
	result.aimg = c.aimg.Clone(a, gsp)

	// Manually copy references that shallow copy poorly, as needed
	// Pointers, slices, maps, functions, channels etc
	result.ghv = *c.ghv.Clone(a)

	result.children = arena.MakeSlice[*Char](a, len(c.children), len(c.children))
	copy(result.children, c.children)

	result.targets = arena.MakeSlice[int32](a, len(c.targets), len(c.targets))
	copy(result.targets, c.targets)

	result.hitdefTargets = arena.MakeSlice[int32](a, len(c.hitdefTargets), len(c.hitdefTargets))
	copy(result.hitdefTargets, c.hitdefTargets)

	result.hitdefTargetsBuffer = arena.MakeSlice[int32](a, len(c.hitdefTargetsBuffer), len(c.hitdefTargetsBuffer))
	copy(result.hitdefTargetsBuffer, c.hitdefTargetsBuffer)

	result.enemyNearList = arena.MakeSlice[*Char](a, len(c.enemyNearList), len(c.enemyNearList))
	copy(result.enemyNearList, c.enemyNearList)

	result.p2EnemyList = arena.MakeSlice[*Char](a, len(c.p2EnemyList), len(c.p2EnemyList))
	copy(result.p2EnemyList, c.p2EnemyList)

	if c.p2EnemyBackup != nil {
		tmp := *c.p2EnemyBackup
		result.p2EnemyBackup = &tmp
	}

	result.clipboardText = arena.MakeSlice[string](a, len(c.clipboardText), len(c.clipboardText))
	copy(result.clipboardText, c.clipboardText)

	if c.keyctrl[0] {
		result.cmd = arena.MakeSlice[CommandList](a, len(c.cmd), len(c.cmd))
		for i, c := range c.cmd {
			result.cmd[i] = c.Clone(a)
		}
		for i := range result.cmd {
			result.cmd[i].Buffer = result.cmd[0].Buffer
		}
	}

	result.ss = c.ss.Clone(a)

	result.cnsvar = *gsp.Get(c.cnsvar).(*map[int32]int32)
	maps.Clear(result.cnsvar)
	for k, v := range c.cnsvar {
		result.cnsvar[k] = v
	}
	result.cnsfvar = *gsp.Get(c.cnsfvar).(*map[int32]float32)
	maps.Clear(result.cnsfvar)
	for k, v := range c.cnsfvar {
		result.cnsfvar[k] = v
	}

	result.cnssysvar = *gsp.Get(c.cnssysvar).(*map[int32]int32)
	maps.Clear(result.cnssysvar)
	for k, v := range c.cnssysvar {
		result.cnssysvar[k] = v
	}
	result.cnssysfvar = *gsp.Get(c.cnssysfvar).(*map[int32]float32)
	maps.Clear(result.cnssysfvar)
	for k, v := range c.cnssysfvar {
		result.cnssysfvar[k] = v
	}

	result.mapArray = *gsp.Get(c.mapArray).(*map[string]float32)
	maps.Clear(result.mapArray)
	for k, v := range c.mapArray {
		result.mapArray[k] = v
	}

	if c.inputShift != nil {
		result.inputShift = arena.MakeSlice[[2]int](a, len(c.inputShift), len(c.inputShift))
		copy(result.inputShift, c.inputShift)
	}

	return
}

func (cl *CharList) Clone(a *arena.Arena, gsp *GameStatePool) (result CharList) {
	result = *cl

	// Manually copy references that shallow copy poorly, as needed
	// Pointers, slices, maps, functions, channels etc
	result.runOrder = arena.MakeSlice[*Char](a, len(cl.runOrder), len(cl.runOrder))
	copy(result.runOrder, cl.runOrder)

	result.idMap = *gsp.Get(cl.idMap).(*map[int32]*Char)
	maps.Clear(result.idMap)
	for k, v := range cl.idMap {
		result.idMap[k] = v
	}
	return
}

func (pf *PalFX) Clone(a *arena.Arena) *PalFX {
	if pf == nil {
		return nil
	}
	result := *pf
	if pf.remap != nil {
		result.remap = arena.MakeSlice[int](a, len(pf.remap), len(pf.remap))
		copy(result.remap, pf.remap)
	}
	return &result
}

func (ce *CommandStep) Clone(a *arena.Arena) (result CommandStep) {
	result = *ce
	result.keys = arena.MakeSlice[CommandStepKey](a, len(ce.keys), len(ce.keys))
	copy(result.keys, ce.keys)
	return
}

func (c *Command) clone(a *arena.Arena) (result Command) {
	result = *c

	result.completed = arena.MakeSlice[bool](a, len(c.completed), len(c.completed))
	copy(result.completed, c.completed)

	result.stepTimers = arena.MakeSlice[int32](a, len(c.stepTimers), len(c.stepTimers))
	copy(result.stepTimers, c.stepTimers)

	// Maybe we don't need to save these or any other things that are only updated upon loading the char
	/*
		result.steps = arena.MakeSlice[CommandStep](a, len(c.steps), len(c.steps))
		for i := 0; i < len(c.steps); i++ {
			result.steps[i] = c.steps[i].Clone(a)
		}
	*/

	// New input code does not use these
	/*
		result.held = arena.MakeSlice[bool](a, len(c.held), len(c.held))
		copy(result.held, c.held)

		result.hold = arena.MakeSlice[[]CommandKey](a, len(c.hold), len(c.hold))
		for i := 0; i < len(c.hold); i++ {
			result.hold[i] = arena.MakeSlice[CommandKey](a, len(c.hold[i]), len(c.hold[i]))
			for j := 0; j < len(c.hold[i]); j++ {
				result.hold[i][j] = c.hold[i][j]
			}
		}
	*/

	return
}

func (cl *CommandList) Clone(a *arena.Arena) (result CommandList) {
	result = *cl

	result.Buffer = arena.New[InputBuffer](a)
	*result.Buffer = *cl.Buffer

	result.Commands = arena.MakeSlice[[]Command](a, len(cl.Commands), len(cl.Commands))
	for i := 0; i < len(cl.Commands); i++ {
		result.Commands[i] = arena.MakeSlice[Command](a, len(cl.Commands[i]), len(cl.Commands[i]))
		for j := 0; j < len(cl.Commands[i]); j++ {
			result.Commands[i][j] = cl.Commands[i][j].clone(a)
		}
	}

	return
}

func (l *Lifebar) Clone(a *arena.Arena) (result Lifebar) {
	result = *l

	// Round
	if l.ro != nil {
		result.ro = &LifeBarRound{} // Shallow copy
		*result.ro = *l.ro
		// Round Transition
		// This needs a deep copy because it's a pointer inside LifebarRound and we need the timers
		// When round transitions are expanded this can be revisited
		if l.ro.rt != nil {
			result.ro.rt = arena.New[LifeBarRoundTransition](a)
			*result.ro.rt = *l.ro.rt
		}
	}

	// Combo
	for i := 0; i < len(l.co); i++ {
		if l.co[i] != nil {
			result.co[i] = arena.New[LifeBarCombo](a)
			*result.co[i] = *l.co[i]
		}
	}

	// We probably don't need a deep copy of these
	/*
		//UIT
		for i := 0; i < len(l.sc); i++ {
			if l.sc[i] != nil {
				result.sc[i] = arena.New[LifeBarScore](a)
				*result.sc[i] = *l.sc[i]
			}
		}
		if l.ti != nil {
			result.ti = arena.New[LifeBarTime](a)
			*result.ti = *l.ti
		}
		//

		// Not UIT adding anyway
		for i := 0; i < len(l.wc); i++ {
			result.wc[i] = arena.New[LifeBarWinCount](a)
			*result.wc[i] = *l.wc[i]
		}

		if l.ma != nil {
			result.ma = arena.New[LifeBarMatch](a)
			*result.ma = *l.ma
		}

		for i := 0; i < len(l.ai); i++ {
			result.ai[i] = arena.New[LifeBarAiLevel](a)
			*result.ai[i] = *l.ai[i]
		}

		if l.tr != nil {
			result.tr = arena.New[LifeBarTimer](a)
			*result.tr = *l.tr
		}
		//

		// Order
		for i := range result.order {
			result.order[i] = arena.MakeSlice[int](a, len(l.order[i]), len(l.order[i]))
			copy(result.order[i], l.order[i])
		}

		// HealthBar
		for i := range result.hb {
			result.hb[i] = arena.MakeSlice[*HealthBar](a, len(l.hb[i]), len(l.hb[i]))
			for j := 0; j < len(l.hb[i]); j++ {
				result.hb[i][j] = arena.New[HealthBar](a)
				*result.hb[i][j] = *l.hb[i][j]
			}
		}

		// PowerBar
		for i := range result.pb {
			result.pb[i] = arena.MakeSlice[*PowerBar](a, len(l.pb[i]), len(l.pb[i]))
			for j := 0; j < len(l.pb[i]); j++ {
				result.pb[i][j] = arena.New[PowerBar](a)
				*result.pb[i][j] = *l.pb[i][j]
			}
		}

		// GuardBar
		for i := range result.gb {
			result.gb[i] = arena.MakeSlice[*GuardBar](a, len(l.gb[i]), len(l.gb[i]))
			for j := 0; j < len(l.gb[i]); j++ {
				result.gb[i][j] = arena.New[GuardBar](a)
				*result.gb[i][j] = *l.gb[i][j]
			}
		}

		// StunBar
		for i := range result.sb {
			result.sb[i] = arena.MakeSlice[*StunBar](a, len(l.sb[i]), len(l.sb[i]))
			for j := 0; j < len(l.sb[i]); j++ {
				result.sb[i][j] = arena.New[StunBar](a)
				*result.sb[i][j] = *l.sb[i][j]
			}
		}

		// Face
		for i := range result.fa {
			result.fa[i] = arena.MakeSlice[*LifeBarFace](a, len(l.fa[i]), len(l.fa[i]))
			for j := 0; j < len(l.fa[i]); j++ {
				result.fa[i][j] = arena.New[LifeBarFace](a)
				*result.fa[i][j] = *l.fa[i][j]
			}
		}

		// Name
		for i := range result.nm {
			result.nm[i] = arena.MakeSlice[*LifeBarName](a, len(l.nm[i]), len(l.nm[i]))
			for j := 0; j < len(l.nm[i]); j++ {
				result.nm[i][j] = arena.New[LifeBarName](a)
				*result.nm[i][j] = *l.nm[i][j]
			}
		}
	*/

	// Action
	for i := range result.ac {
		if l.ac[i] != nil {
			result.ac[i] = arena.New[LifeBarAction](a)

			*result.ac[i] = *l.ac[i]

			if l.ac[i].messages != nil {
				result.ac[i].messages = arena.MakeSlice[*LbMsg](a, len(l.ac[i].messages), len(l.ac[i].messages))
				for j := 0; j < len(l.ac[i].messages); j++ {
					result.ac[i].messages[j] = arena.New[LbMsg](a)
					*result.ac[i].messages[j] = *l.ac[i].messages[j]
				}
			}
		}
	}

	return
}

func (s *Stage) Clone(a *arena.Arena, gsp *GameStatePool) *Stage {
	result := &Stage{}
	*result = *s

	// Clone attached char def
	result.attachedchardef = arena.MakeSlice[string](a, len(s.attachedchardef), len(s.attachedchardef))
	copy(result.attachedchardef, s.attachedchardef)

	// Clone constants
	result.constants = make(map[string]float32, len(s.constants))
	for k, v := range s.constants {
		result.constants[k] = v
	}

	// Clone animation table
	result.at = *gsp.Get(s.at).(*AnimationTable)
	maps.Clear(result.at)
	for k, v := range s.at {
		result.at[k] = v.Clone(a, gsp)
	}

	// Clone backgrounds and rebuild mapping
	bgMap := make(map[*backGround]*backGround, len(s.bg))
	result.bg = arena.MakeSlice[*backGround](a, len(s.bg), len(s.bg))
	for i, oldbg := range s.bg {
		newbg := &backGround{}
		*newbg = *oldbg
		if oldbg.anim != nil {
			animCopy := *oldbg.anim
			newbg.anim = &animCopy
		}
		result.bg[i] = newbg
		bgMap[oldbg] = newbg
	}

	// Clone bgCtrl and point them to the cloned BG's
	result.bgc = arena.MakeSlice[bgCtrl](a, len(s.bgc), len(s.bgc))
	for i, oldbgc := range s.bgc {
		newbgc := oldbgc
		newbgc.bg = arena.MakeSlice[*backGround](a, len(oldbgc.bg), len(oldbgc.bg))
		for j, oldbg := range oldbgc.bg {
			newbgc.bg[j] = bgMap[oldbg]
		}
		result.bgc[i] = newbgc
	}

	return result
}

// other things can be copied, only focusing on OCD right now
func (s Select) Clone(a *arena.Arena) (result Select) {
	result = s
	for i := 0; i < len(s.ocd); i++ {
		result.ocd[i] = arena.MakeSlice[OverrideCharData](a, len(s.ocd[i]), len(s.ocd[i]))
		copy(result.ocd[i], s.ocd[i])
	}

	result.stageAnimPreload = arena.MakeSlice[int32](a, len(s.stageAnimPreload), len(s.stageAnimPreload))
	copy(result.stageAnimPreload, s.stageAnimPreload)

	return
}

func (f Fight) Clone(a *arena.Arena, gsp *GameStatePool) (result Fight) {
	result = f
	result.oldStageVars = *f.oldStageVars.Clone(a, gsp)
	result.level = arena.MakeSlice[int32](a, len(f.level), len(f.level))
	copy(result.level, f.level)

	result.life = arena.MakeSlice[int32](a, len(f.life), len(f.life))
	copy(result.life, f.life)
	result.pow = arena.MakeSlice[int32](a, len(f.pow), len(f.pow))
	copy(result.pow, f.pow)
	result.gpow = arena.MakeSlice[int32](a, len(f.gpow), len(f.gpow))
	copy(result.gpow, f.gpow)
	result.spow = arena.MakeSlice[int32](a, len(f.spow), len(f.spow))
	copy(result.spow, f.spow)
	result.rlife = arena.MakeSlice[int32](a, len(f.rlife), len(f.rlife))
	copy(result.rlife, f.rlife)

	result.cnsvar = arena.MakeSlice[map[int32]int32](a, len(f.cnsvar), len(f.cnsvar))
	for i := 0; i < len(f.cnsvar); i++ {
		result.cnsvar[i] = make(map[int32]int32, len(f.cnsvar[i]))
		for k, v := range f.cnsvar[i] {
			result.cnsvar[i][k] = v
		}
	}

	result.cnsfvar = arena.MakeSlice[map[int32]float32](a, len(f.cnsfvar), len(f.cnsfvar))
	for i := 0; i < len(f.cnsfvar); i++ {
		result.cnsfvar[i] = make(map[int32]float32, len(f.cnsfvar[i]))
		for k, v := range f.cnsfvar[i] {
			result.cnsfvar[i][k] = v
		}
	}

	result.dialogue = arena.MakeSlice[[]string](a, len(f.dialogue), len(f.dialogue))
	for i := 0; i < len(result.dialogue); i++ {
		result.dialogue[i] = arena.MakeSlice[string](a, len(f.dialogue[i]), len(f.dialogue[i]))
		copy(result.dialogue[i], f.dialogue[i])
	}

	result.mapArray = arena.MakeSlice[map[string]float32](a, len(f.mapArray), len(f.mapArray))
	for i := 0; i < len(f.mapArray); i++ {
		result.mapArray[i] = *gsp.Get(f.mapArray[i]).(*map[string]float32)
		maps.Clear(result.mapArray[i])
		for k, v := range f.mapArray[i] {
			result.mapArray[i][k] = v
		}
	}
	return
}

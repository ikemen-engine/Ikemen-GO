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

/*
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
*/

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

func (ai *AfterImage) Clone(a *arena.Arena, gsp *GameStatePool) *AfterImage {
	if ai == nil {
		return nil
	}

	result := &AfterImage{}
	*result = *ai

	// Deep copy Animations
	for i := range ai.imgs {
		if ai.imgs[i].anim != nil {
			result.imgs[i].anim = ai.imgs[i].anim.Clone(a, gsp)
		}
	}

	// Deep copy PalFX
	for i := range ai.palfx {
		if ai.palfx[i] != nil {
			result.palfx[i] = ai.palfx[i].Clone(a)
		}
	}

	return result
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

	if e.aimg != nil {
		result.aimg = e.aimg.Clone(a, gsp)
	}

	return result
}

func (p *Projectile) clone(a *arena.Arena, gsp *GameStatePool) *Projectile {
	if p == nil {
		return nil
	}

	result := &Projectile{}
	*result = *p

	if p.anim != nil {
		result.anim = p.anim.Clone(a, gsp)
	}

	if p.palfx != nil {
		result.palfx = p.palfx.Clone(a)
	}

	if p.aimg != nil {
		result.aimg = p.aimg.Clone(a, gsp)
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
	}

	if c.shadowAnim != nil {
		result.shadowAnim = c.shadowAnim.Clone(a, gsp)
	}
	if c.reflectAnim != nil {
		result.reflectAnim = c.reflectAnim.Clone(a, gsp)
	}

	// TODO: Profiling shows this is hotter than it should be
	// Maybe we ought to clear animation data from them when their timer expires
	// Update: Done already but copying 60 PalFX's is still a problem
	if c.aimg != nil {
		result.aimg = c.aimg.Clone(a, gsp)
	}

	if c.palfx != nil {
		result.palfx = c.palfx.Clone(a)
	}

	// Manually copy references that shallow copy poorly, as needed
	// Pointers, slices, maps, functions, channels etc
	result.ghv = *c.ghv.Clone(a)

	result.children = arena.MakeSlice[int32](a, len(c.children), len(c.children))
	copy(result.children, c.children)

	result.targets = arena.MakeSlice[int32](a, len(c.targets), len(c.targets))
	copy(result.targets, c.targets)

	result.hitdefTargets = arena.MakeSlice[int32](a, len(c.hitdefTargets), len(c.hitdefTargets))
	copy(result.hitdefTargets, c.hitdefTargets)

	result.hitdefTargetsBuffer = arena.MakeSlice[int32](a, len(c.hitdefTargetsBuffer), len(c.hitdefTargetsBuffer))
	copy(result.hitdefTargetsBuffer, c.hitdefTargetsBuffer)

	result.enemyNearList = arena.MakeSlice[int32](a, len(c.enemyNearList), len(c.enemyNearList))
	copy(result.enemyNearList, c.enemyNearList)

	result.p2EnemyList = arena.MakeSlice[int32](a, len(c.p2EnemyList), len(c.p2EnemyList))
	copy(result.p2EnemyList, c.p2EnemyList)

	//if c.p2EnemyBackup != nil {
	//	tmp := *c.p2EnemyBackup
	//	result.p2EnemyBackup = &tmp
	//}
	// This ought to be enough
	//result.p2EnemyBackup = c.p2EnemyBackup
	// Converted to an ID so the shallow copy now gets it

	result.inputShift = arena.MakeSlice[[2]int](a, len(c.inputShift), len(c.inputShift))
	copy(result.inputShift, c.inputShift)

	result.clsnOverrides = arena.MakeSlice[ClsnOverride](a, len(c.clsnOverrides), len(c.clsnOverrides))
	copy(result.clsnOverrides, c.clsnOverrides)

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

	return
}

func (cl *CharList) Clone(a *arena.Arena, gsp *GameStatePool) (result CharList) {
	result = *cl

	result.creationOrder = arena.MakeSlice[*Char](a, len(cl.creationOrder), len(cl.creationOrder))
	copy(result.creationOrder, cl.creationOrder)

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
		// Fade state
		result.ro.fadeIn = l.ro.fadeIn.Clone(a)
		result.ro.fadeOut = l.ro.fadeOut.Clone(a)
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
	result.animTable = *gsp.Get(s.animTable).(*AnimationTable)
	maps.Clear(result.animTable)
	for k, v := range s.animTable {
		result.animTable[k] = v.Clone(a, gsp)
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

/*func (s Select) Clone(a *arena.Arena) (result Select) {
	result = s

	// Copy selected (mutable; slices)
	for side := 0; side < len(s.selected); side++ {
		if s.selected[side] == nil {
			result.selected[side] = nil
			continue
		}
		if a != nil {
			result.selected[side] = arena.MakeSlice[[2]int](a, len(s.selected[side]), len(s.selected[side]))
		} else {
			result.selected[side] = make([][2]int, len(s.selected[side]))
		}
		copy(result.selected[side], s.selected[side])
	}

	// Copy overwrite map headers (maps are reference types)
	if s.cdefOverwrite != nil {
		result.cdefOverwrite = make(map[int]string, len(s.cdefOverwrite))
		for k, v := range s.cdefOverwrite {
			result.cdefOverwrite[k] = v
		}
	}

	// Copy music map header and candidate slices.
	// The *bgMusic entries themselves are treated as immutable; copying pointers is enough.
	if s.music != nil {
		result.music = make(Music, len(s.music))
		for k, lst := range s.music {
			if lst == nil {
				result.music[k] = nil
			continue
			}
			var nlst []*bgMusic
			if a != nil {
				nlst = arena.MakeSlice[*bgMusic](a, len(lst), len(lst))
			} else {
				nlst = make([]*bgMusic, len(lst))
			}
			copy(nlst, lst)
			result.music[k] = nlst
		}
	}

	// gameParams should never be nil during fight code paths.
	// Keep the existing pointer (stable), but defensively initialize if needed.
	if result.gameParams == nil {
		result.gameParams = newGameParams()
	}

	return result
}*/

func cloneTextSprite(a *arena.Arena, ts *TextSprite) *TextSprite {
	if ts == nil {
		return nil
	}
	dst := arena.New[TextSprite](a)
	*dst = *ts

	// Only copy references that shallow copy poorly, as needed.
	if ts.params != nil {
		dst.params = make([]interface{}, len(ts.params))
		copy(dst.params, ts.params)
	}
	if ts.palfx != nil {
		dst.palfx = ts.palfx.Clone(a)
	}
	return dst
}

func (fa *Fade) Clone(a *arena.Arena) *Fade {
	if fa == nil {
		return nil
	}
	result := arena.New[Fade](a)
	*result = *fa
	// Avoid sharing animation state between saved states.
	if fa.animData != nil {
		// Anim has Copy() in the codebase and is already used for UI anim duplication.
		result.animData = fa.animData.Copy()
	}
	return result
}

/*func (me *MotifMenu) Clone(a *arena.Arena) (result MotifMenu) {
	result = *me
	return
}*/

/*func (ch *MotifChallenger) Clone(a *arena.Arena) (result MotifChallenger) {
	result = *ch
	return
}*/

func (co *MotifContinue) Clone(a *arena.Arena) (result MotifContinue) {
	result = *co
	if co.counts != nil {
		result.counts = arena.MakeSlice[string](a, len(co.counts), len(co.counts))
		copy(result.counts, co.counts)
	}
	return
}

func cloneDialogueToken(t DialogueToken) DialogueToken {
	result := t
	if t.value != nil {
		result.value = make([]interface{}, len(t.value))
		copy(result.value, t.value)
	}
	return result
}

func cloneDialogueParsedLine(a *arena.Arena, src DialogueParsedLine) (result DialogueParsedLine) {
	result = src
	if src.tokens != nil {
		result.tokens = make(map[int][]DialogueToken, len(src.tokens))
		for k, toks := range src.tokens {
			if toks == nil {
				continue
			}
			dst := arena.MakeSlice[DialogueToken](a, len(toks), len(toks))
			for i := 0; i < len(toks); i++ {
				dst[i] = cloneDialogueToken(toks[i])
			}
			result.tokens[k] = dst
		}
	}
	return
}

/*func (de *MotifDemo) Clone(a *arena.Arena) (result MotifDemo) {
	result = *de
	return
}*/

func (di *MotifDialogue) Clone(a *arena.Arena) (result MotifDialogue) {
	result = *di
	if di.parsed != nil {
		result.parsed = arena.MakeSlice[DialogueParsedLine](a, len(di.parsed), len(di.parsed))
		for i := 0; i < len(di.parsed); i++ {
			result.parsed[i] = cloneDialogueParsedLine(a, di.parsed[i])
		}
	}
	return
}

/*func (vi *MotifVictory) Clone(a *arena.Arena) (result MotifVictory) {
	result = *vi
	return
}*/

func (rr *rankingRow) Clone(a *arena.Arena) (result rankingRow) {
	result = *rr
	if rr.pals != nil {
		result.pals = arena.MakeSlice[int32](a, len(rr.pals), len(rr.pals))
		copy(result.pals, rr.pals)
	}
	if rr.chars != nil {
		result.chars = arena.MakeSlice[string](a, len(rr.chars), len(rr.chars))
		copy(result.chars, rr.chars)
	}
	// Per-slot anim state must not be shared between saved states.
	if rr.bgs != nil {
		result.bgs = arena.MakeSlice[*Anim](a, len(rr.bgs), len(rr.bgs))
		for i := 0; i < len(rr.bgs); i++ {
			if rr.bgs[i] != nil {
				result.bgs[i] = rr.bgs[i].Copy()
			}
		}
	}
	if rr.faces != nil {
		result.faces = arena.MakeSlice[*Anim](a, len(rr.faces), len(rr.faces))
		for i := 0; i < len(rr.faces); i++ {
			if rr.faces[i] != nil {
				result.faces[i] = rr.faces[i].Copy()
			}
		}
	}
	// Per-row TextSprites
	result.rankData = cloneTextSprite(a, rr.rankData)
	result.resultData = cloneTextSprite(a, rr.resultData)
	result.nameData = cloneTextSprite(a, rr.nameData)
	result.rankDataActive = cloneTextSprite(a, rr.rankDataActive)
	result.rankDataActive2 = cloneTextSprite(a, rr.rankDataActive2)
	result.resultDataActive = cloneTextSprite(a, rr.resultDataActive)
	result.resultDataActive2 = cloneTextSprite(a, rr.resultDataActive2)
	result.nameDataActive = cloneTextSprite(a, rr.nameDataActive)
	result.nameDataActive2 = cloneTextSprite(a, rr.nameDataActive2)
	return
}

func (hi *MotifHiscore) Clone(a *arena.Arena) (result MotifHiscore) {
	result = *hi
	if hi.rows != nil {
		result.rows = arena.MakeSlice[rankingRow](a, len(hi.rows), len(hi.rows))
		for i := 0; i < len(hi.rows); i++ {
			result.rows[i] = hi.rows[i].Clone(a)
		}
	}
	if hi.letters != nil {
		result.letters = arena.MakeSlice[int](a, len(hi.letters), len(hi.letters))
		copy(result.letters, hi.letters)
	}
	return
}

func (wi *MotifWin) Clone(a *arena.Arena) (result MotifWin) {
	result = *wi

	if wi.keyCancel != nil {
		result.keyCancel = arena.MakeSlice[string](a, len(wi.keyCancel), len(wi.keyCancel))
		copy(result.keyCancel, wi.keyCancel)
	}
	for i := 0; i < 4; i++ {
		if wi.p1States[i] != nil {
			result.p1States[i] = arena.MakeSlice[int32](a, len(wi.p1States[i]), len(wi.p1States[i]))
			copy(result.p1States[i], wi.p1States[i])
		}
		if wi.p2States[i] != nil {
			result.p2States[i] = arena.MakeSlice[int32](a, len(wi.p2States[i]), len(wi.p2States[i]))
			copy(result.p2States[i], wi.p2States[i])
		}
	}
	return
}

func (m *Motif) Clone(a *arena.Arena) (result Motif) {
	result = *m

	// Fade state
	result.fadeIn = m.fadeIn.Clone(a)
	result.fadeOut = m.fadeOut.Clone(a)

	// Motif sub-state that contains reference types
	// Dialogue can run during match, keep it rollback-safe.
	//result.me = m.me.Clone(a)
	//result.ch = m.ch.Clone(a)
	//result.de = m.de.Clone(a)
	result.di = m.di.Clone(a)

	// Post-match-only motifs: don't waste time cloning them (and don't rollback-touch them)
	// during a match when sys.postMatchFlg is false.
	if sys.postMatchFlg {
		result.co = m.co.Clone(a)
		//result.vi = m.vi.Clone(a)
		result.hi = m.hi.Clone(a)
		result.wi = m.wi.Clone(a)
	} else {
		// Keep current runtime values so rollback loads during a match won't affect them.
		result.co = sys.motif.co
		result.vi = sys.motif.vi
		result.hi = sys.motif.hi
		result.wi = sys.motif.wi
	}

	return
}

// Music tables are treated as immutable during a match; for storyboard we only need
// to isolate map/slice headers so saved states don't alias if something rebuilds them.
// The *bgMusic entries themselves are treated as immutable; copying pointers is enough.
func cloneMusicMapShallow(a *arena.Arena, src Music) Music {
	if src == nil {
		return nil
	}
	dst := make(Music, len(src))
	for k, lst := range src {
		if lst == nil {
			dst[k] = nil
			continue
		}
		nlst := arena.MakeSlice[*bgMusic](a, len(lst), len(lst))
		copy(nlst, lst)
		dst[k] = nlst
	}
	return dst
}

// Storyboard.Clone:
// - Only meant to be used when storyboard is active (caller gates it).
// - Deep-copies runtime-mutated per-layer state (Anim/Text + typewriter fields).
// - Keeps static resources shared (IniFile/Sff/Snd/Fnt/Model/At/etc).
// - Rebuilds derived dialogue queue so pointers point into the cloned layer maps.
func (s *Storyboard) Clone(a *arena.Arena) (result Storyboard) {
	if s == nil {
		return Storyboard{}
	}

	// Shallow copy keeps static resources shared.
	result = *s

	// sceneKeys is mutated/rebuilt; don't alias.
	if s.sceneKeys != nil {
		result.sceneKeys = arena.MakeSlice[string](a, len(s.sceneKeys), len(s.sceneKeys))
		copy(result.sceneKeys, s.sceneKeys)
	}

	// We'll rebuild this cache after cloning.
	result.dialogueLayers = nil

	// Deep copy per-scene runtime state.
	if s.Scene != nil {
		result.Scene = make(map[string]*SceneProperties, len(s.Scene))
		for sceneKey, sp := range s.Scene {
			if sp == nil {
				result.Scene[sceneKey] = nil
				continue
			}

			nsp := &SceneProperties{}
			*nsp = *sp

			// Layer map (runtime state lives here)
			if sp.Layer != nil {
				nsp.Layer = make(map[string]*LayerProperties, len(sp.Layer))
				for layerKey, lp := range sp.Layer {
					if lp == nil {
						nsp.Layer[layerKey] = nil
						continue
					}

					nlp := &LayerProperties{}
					*nlp = *lp

					// Anim runtime state must not be shared across saved states.
					if lp.AnimData != nil {
						nlp.AnimData = lp.AnimData.Copy()
					}

					// TextSprite runtime state must not be shared across saved states.
					nlp.TextSpriteData = cloneTextSprite(a, lp.TextSpriteData)

					// typedLen/charDelayCounter/lineFullyRendered are plain fields copied above.
					nsp.Layer[layerKey] = nlp
				}
			}

			// Sound map isn't mutated during playback, but map header is a reference type.
			if sp.Sound != nil {
				nsp.Sound = make(map[string]*SoundProperties, len(sp.Sound))
				for soundKey, sv := range sp.Sound {
					if sv == nil {
						nsp.Sound[soundKey] = nil
						continue
					}
					ns := &SoundProperties{}
					*ns = *sv
					nsp.Sound[soundKey] = ns
				}
			}

			// Per-scene music config (treat bgMusic entries as immutable).
			nsp.Music = cloneMusicMapShallow(a, sp.Music)

			if sp.Bg.BGDef != nil {
				nbg := &BGDef{}
				*nbg = *sp.Bg.BGDef
				nsp.Bg.BGDef = nbg
			}

			result.Scene[sceneKey] = nsp
		}
	}

	// Rebuild dialogue queue so it points into the cloned maps.
	if len(result.sceneKeys) == 0 && result.Scene != nil {
		result.sceneKeys = SortedKeys(result.Scene)
	}
	if len(result.sceneKeys) > 0 &&
		result.currentSceneIndex >= 0 &&
		result.currentSceneIndex < len(result.sceneKeys) {
		sceneKey := result.sceneKeys[result.currentSceneIndex]
		if sp, ok := result.Scene[sceneKey]; ok && sp != nil {
			pos := result.dialoguePos
			result.buildDialogueQueue(sp)
			result.dialoguePos = pos
			result.syncDialoguePosToTime()
		}
	}

	return result
}

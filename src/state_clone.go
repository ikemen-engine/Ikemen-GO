package main

import (
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

func (a *Animation) Clone(gsp *GameStatePool) (result *Animation) {
	result = new(Animation)
	*result = *a

	result.frames = *gsp.Get(a.frames).(*[]AnimFrame)
	result.frames = result.frames[:0]
	for i := 0; i < len(a.frames); i++ {
		result.frames = append(result.frames, a.frames[i].Clone())
	}

	// Normally only afterimages have these as nil, but it's still worth avoiding the work whenever possible
	if a.interpolate_offset != nil {
		result.interpolate_offset = make([]int32, len(a.interpolate_offset), len(a.interpolate_offset))
		copy(result.interpolate_offset, a.interpolate_offset)
	}
	if a.interpolate_scale != nil {
		result.interpolate_scale = make([]int32, len(a.interpolate_scale), len(a.interpolate_scale))
		copy(result.interpolate_scale, a.interpolate_scale)
	}
	if a.interpolate_angle != nil {
		result.interpolate_angle = make([]int32, len(a.interpolate_angle), len(a.interpolate_angle))
		copy(result.interpolate_angle, a.interpolate_angle)
	}
	if a.interpolate_blend != nil {
		result.interpolate_blend = make([]int32, len(a.interpolate_blend), len(a.interpolate_blend))
		copy(result.interpolate_blend, a.interpolate_blend)
	}

	return
}

// CloneState copies an animation's mutable playback state while sharing its immutable frame/interpolation data.
func (anim *Animation) CloneState() *Animation {
	if anim == nil {
		return nil
	}
	result := new(Animation)
	*result = *anim
	return result
}

// Returns a value to avoid an allocation when stored in a slice
func (af *AnimFrame) Clone() (result AnimFrame) {
	result = *af

	if af.Clsn1 != nil {
		result.Clsn1 = make([][4]float32, len(af.Clsn1), len(af.Clsn1))
		copy(result.Clsn1, af.Clsn1)
	}
	if af.Clsn2 != nil {
		result.Clsn2 = make([][4]float32, len(af.Clsn2), len(af.Clsn2))
		copy(result.Clsn2, af.Clsn2)
	}

	return
}

/*
func (sp StringPool) Clone(gsp *GameStatePool) (result StringPool) {
	result = sp
	result.List = make([]string, len(sp.List), len(sp.List))
	copy(result.List, sp.List)
	result.Map = *gsp.Get(sp.Map).(*map[string]int)
	maps.Clear(result.Map)

	for k, v := range sp.Map {
		result.Map[k] = v
	}
	return
}
*/

// StateBlock no longer needs Clone(). The pointer is shared across states
/*
func (b *StateBlock) Clone() (result StateBlock) {
	result = *b
	result.trigger = make([]OpCode, len(b.trigger), len(b.trigger))
	copy(result.trigger, b.trigger)
	if b.elseBlock != nil {
		eb := b.elseBlock.Clone()
		result.elseBlock = &eb
	}

	result.forCtrlVar.be = make([]OpCode, len(b.forCtrlVar.be), len(b.forCtrlVar.be))
	copy(result.forCtrlVar.be, b.forCtrlVar.be)

	for i := 0; i < len(b.forExpression); i++ {
		result.forExpression[i] = make([]OpCode, len(b.forExpression[i]), len(b.forExpression[i]))
		copy(result.forExpression[i], b.forExpression[i])
	}

	result.ctrls = make([]StateController, len(b.ctrls), len(b.ctrls))
	copy(result.ctrls, b.ctrls)
	return result
}

func (sb *StateBytecode) Clone() (result StateBytecode) {
	result = *sb
	result.stateDef = make([]byte, len(sb.stateDef), len(sb.stateDef))
	copy(result.stateDef, sb.stateDef)

	result.ctrlsps = make([]int32, len(sb.ctrlsps), len(sb.ctrlsps))
	copy(result.ctrlsps, sb.ctrlsps)
	result.block = sb.block.Clone()
	return result
}
*/

func (ghv *GetHitVar) Clone() (result *GetHitVar) {
	result = new(GetHitVar)
	*result = *ghv

	// Manually copy references that shallow copy poorly, as needed
	// Pointers, slices, maps, functions, channels etc
	result.targetedBy = make([][2]int32, len(ghv.targetedBy), len(ghv.targetedBy))
	copy(result.targetedBy, ghv.targetedBy)

	return
}

func (ai *AfterImage) Clone(gsp *GameStatePool) *AfterImage {
	if ai == nil {
		return nil
	}

	result := new(AfterImage)
	*result = *ai

	// Draw buffer isn't game state
	result.drawbuf = nil

	// Allocate the slice backing arrays before replacing nested pointers.
	if ai.imgs != nil {
		result.imgs = make([]SpriteData, len(ai.imgs), len(ai.imgs))
		copy(result.imgs, ai.imgs)
	}
	if ai.palfx != nil {
		result.palfx = make([]*PalFX, len(ai.palfx), len(ai.palfx))
		copy(result.palfx, ai.palfx)
	}

	// Deep copy Animations
	// Afterimages use stripped down animations, so this will be a bit faster than normal
	for i := range ai.imgs {
		if ai.imgs[i].anim != nil {
			result.imgs[i].anim = ai.imgs[i].anim.Clone(gsp)
			// Copy the sprite to prevent aliasing, since afterimages modify their sprites
			if ai.imgs[i].anim.spr != nil {
				spr := *ai.imgs[i].anim.spr
				result.imgs[i].anim.spr = &spr
			}
		}
	}

	// Deep copy PalFX
	for i := range ai.palfx {
		if ai.palfx[i] != nil {
			result.palfx[i] = ai.palfx[i].Clone()
		}
	}

	return result
}

func (e *Explod) Clone(gsp *GameStatePool) *Explod {
	if e == nil {
		return nil
	}

	result := &Explod{}
	*result = *e

	if e.anim != nil {
		result.anim = e.anim.Clone(gsp)
	}

	if e.customShader.name != "" {
		result.customShader = e.customShader.Clone(gsp)
	}

	if e.palfx != nil {
		result.palfx = e.palfx.Clone()
	}

	if e.aimg != nil {
		result.aimg = e.aimg.Clone(gsp)
	}

	return result
}

func (p *Projectile) clone(gsp *GameStatePool) *Projectile {
	if p == nil {
		return nil
	}

	result := &Projectile{}
	*result = *p

	if p.anim != nil {
		result.anim = p.anim.Clone(gsp)
	}

	if p.customShader.name != "" {
		result.customShader = p.customShader.Clone(gsp)
	}

	if p.palfx != nil {
		result.palfx = p.palfx.Clone()
	}

	if p.aimg != nil {
		result.aimg = p.aimg.Clone(gsp)
	}

	return result
}

func (ss *StateState) Clone() (result StateState) {
	result = *ss
	result.ps = make([]int32, len(ss.ps), len(ss.ps))
	copy(result.ps, ss.ps)

	//for i := 0; i < len(ss.hitPauseExecutionToggleFlags); i++ {
	//	result.hitPauseExecutionToggleFlags[i] = make([]bool, len(ss.hitPauseExecutionToggleFlags[i]), len(ss.hitPauseExecutionToggleFlags[i]))
	//	copy(result.hitPauseExecutionToggleFlags[i], ss.hitPauseExecutionToggleFlags[i])
	//}

	// ss.sb is shared, read-only, and never mutated. Safe to share, not clone
	result.sb = ss.sb

	return result
}

func cloneRemapPreset(src RemapPreset, gsp *GameStatePool) RemapPreset {
	if src == nil {
		return nil
	}
	result := *gsp.Get(src).(*RemapPreset)
	maps.Clear(result)
	for group, table := range src {
		if table == nil {
			result[group] = nil
			continue
		}
		tableCopy := *gsp.Get(table).(*RemapTable)
		maps.Clear(tableCopy)
		for number, remap := range table {
			tableCopy[number] = remap
		}
		result[group] = tableCopy
	}
	return result
}

func (c *Char) Clone(gsp *GameStatePool) (result Char) {
	result = Char{}
	result = *c

	if c.anim != nil {
		result.anim = c.anim.Clone(gsp)
	}
	if c.animBackup != nil {
		result.animBackup = c.animBackup.Clone(gsp)
	}

	// RemapSprite mutates a nested map, and SpriteVar exposes its result to character logic.
	// Keep each saved state isolated and make the cloned animations point at the cloned preset.
	result.remapSpr = cloneRemapPreset(c.remapSpr, gsp)
	if result.anim != nil {
		result.anim.remap = result.remapSpr
	}
	if result.animBackup != nil {
		result.animBackup.remap = result.remapSpr
	}

	// Since curFrame is desynced from anim's state, we must save it as well
	if c.curFrame != nil {
		frame := c.curFrame.Clone()
		result.curFrame = &frame
	}

	if c.shadowAnim != nil {
		result.shadowAnim = c.shadowAnim.Clone(gsp)
	}
	if c.reflectAnim != nil {
		result.reflectAnim = c.reflectAnim.Clone(gsp)
	}
	if c.customShader.name != "" {
		result.customShader = c.customShader.Clone(gsp)
	}
	// TODO: Profiling shows this is hotter than it should be
	// Maybe we ought to clear animation data from them when their timer expires
	// Update: Done already but copying 60 PalFX's is still a problem
	if c.aimg != nil {
		result.aimg = c.aimg.Clone(gsp)
	}

	if c.palfx != nil {
		result.palfx = c.palfx.Clone()
	}

	// Manually copy references that shallow copy poorly, as needed
	// Pointers, slices, maps, functions, channels etc
	result.ghv = *c.ghv.Clone()

	result.children = make([]int32, len(c.children), len(c.children))
	copy(result.children, c.children)

	result.targets = make([]int32, len(c.targets), len(c.targets))
	copy(result.targets, c.targets)

	result.hitdefTargets = make([]int32, len(c.hitdefTargets), len(c.hitdefTargets))
	copy(result.hitdefTargets, c.hitdefTargets)

	result.hitdefTargetsBuffer = make([]int32, len(c.hitdefTargetsBuffer), len(c.hitdefTargetsBuffer))
	copy(result.hitdefTargetsBuffer, c.hitdefTargetsBuffer)

	result.enemyNearList = make([]int32, len(c.enemyNearList), len(c.enemyNearList))
	copy(result.enemyNearList, c.enemyNearList)

	result.p2EnemyList = make([]int32, len(c.p2EnemyList), len(c.p2EnemyList))
	copy(result.p2EnemyList, c.p2EnemyList)

	//if c.p2EnemyBackup != nil {
	//	tmp := *c.p2EnemyBackup
	//	result.p2EnemyBackup = &tmp
	//}
	// This ought to be enough
	//result.p2EnemyBackup = c.p2EnemyBackup
	// Converted to an ID so the shallow copy now gets it

	result.inputShift = make([][2]int, len(c.inputShift), len(c.inputShift))
	copy(result.inputShift, c.inputShift)

	for i := range c.clsnOverrides {
		result.clsnOverrides[i] = make([]ClsnOverride, len(c.clsnOverrides[i]), len(c.clsnOverrides[i]))
		copy(result.clsnOverrides[i], c.clsnOverrides[i])
	}

	for i := range c.clsnTransforms {
		result.clsnTransforms[i] = make([]ClsnTransform, len(c.clsnTransforms[i]))
		copy(result.clsnTransforms[i], c.clsnTransforms[i])
	}

	result.clipboardText = make([]string, len(c.clipboardText), len(c.clipboardText))
	copy(result.clipboardText, c.clipboardText)

	// Dialogue controllers append to this queue, and motif dialogue can hold round/match progression.
	if c.dialogue != nil {
		result.dialogue = make([]string, len(c.dialogue), len(c.dialogue))
		copy(result.dialogue, c.dialogue)
	}

	if c.keyctrl[0] {
		result.cmd = make([]CommandList, len(c.cmd), len(c.cmd))
		for i, c := range c.cmd {
			result.cmd[i] = c.Clone()
		}
		for i := range result.cmd {
			result.cmd[i].Buffer = result.cmd[0].Buffer
		}
	}

	result.ss = c.ss.Clone()

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

	result.mapArray = *gsp.Get(c.mapArray).(*map[string]MapValue)
	maps.Clear(result.mapArray)
	for k, v := range c.mapArray {
		result.mapArray[k] = v
	}

	return
}

func (cl *CharList) Clone(gsp *GameStatePool) (result CharList) {
	result = *cl

	result.creationOrder = make([]*Char, len(cl.creationOrder), len(cl.creationOrder))
	copy(result.creationOrder, cl.creationOrder)

	result.runOrder = make([]*Char, len(cl.runOrder), len(cl.runOrder))
	copy(result.runOrder, cl.runOrder)

	result.idMap = *gsp.Get(cl.idMap).(*map[int32]*Char)
	maps.Clear(result.idMap)
	for k, v := range cl.idMap {
		result.idMap[k] = v
	}

	return
}

func (pf *PalFX) Clone() *PalFX {
	if pf == nil {
		return nil
	}
	result := *pf

	if pf.remap != nil {
		result.remap = make([]int, len(pf.remap), len(pf.remap))
		copy(result.remap, pf.remap)
	}

	return &result
}

func (ce *CommandStep) Clone() (result CommandStep) {
	result = *ce
	result.keys = make([]CommandStepKey, len(ce.keys), len(ce.keys))
	copy(result.keys, ce.keys)
	return
}

func (c *Command) clone() (result Command) {
	result = *c

	result.completed = make([]bool, len(c.completed), len(c.completed))
	copy(result.completed, c.completed)

	result.stepTimers = make([]int32, len(c.stepTimers), len(c.stepTimers))
	copy(result.stepTimers, c.stepTimers)

	// Maybe we don't need to save these or any other things that are only updated upon loading the char
	/*
		result.steps = make([]CommandStep, len(c.steps), len(c.steps))
		for i := 0; i < len(c.steps); i++ {
			result.steps[i] = c.steps[i].Clone()
		}
	*/

	// New input code does not use these
	/*
		result.held = make([]bool, len(c.held), len(c.held))
		copy(result.held, c.held)

		result.hold = make([][]CommandKey, len(c.hold), len(c.hold))
		for i := 0; i < len(c.hold); i++ {
			result.hold[i] = make([]CommandKey, len(c.hold[i]), len(c.hold[i]))
			for j := 0; j < len(c.hold[i]); j++ {
				result.hold[i][j] = c.hold[i][j]
			}
		}
	*/

	return
}

func (cl *CommandList) Clone() (result CommandList) {
	result = *cl

	result.Buffer = new(InputBuffer)
	*result.Buffer = *cl.Buffer
	if cl.Buffer.InputReader != nil {
		// Preserve SOCD direction history in command snapshots.
		result.Buffer.InputReader = new(InputReader)
		*result.Buffer.InputReader = *cl.Buffer.InputReader
	}

	result.Commands = make([][]Command, len(cl.Commands), len(cl.Commands))
	for i := 0; i < len(cl.Commands); i++ {
		result.Commands[i] = make([]Command, len(cl.Commands[i]), len(cl.Commands[i]))
		for j := 0; j < len(cl.Commands[i]); j++ {
			result.Commands[i][j] = cl.Commands[i][j].clone()
		}
	}

	return
}

type fightScreenAnimationStateCloner struct {
	// Current FightScreenRound has ~40 logic-bearing AnimTextSnd values.
	src   [64]*Animation
	dst   [64]*Animation
	count int
}

func (cl *fightScreenAnimationStateCloner) clone(anim *Animation) *Animation {
	if anim == nil {
		return nil
	}

	for i := 0; i < cl.count; i++ {
		if cl.src[i] == anim {
			return cl.dst[i]
		}
	}
	result := anim.CloneState()
	if cl.count >= len(cl.src) {
		return result
	}
	cl.src[cl.count] = anim
	cl.dst[cl.count] = result
	cl.count++
	return result
}

func (cl *fightScreenAnimationStateCloner) cloneAnimTextSnd(src AnimTextSnd) (result AnimTextSnd) {
	result = src
	result.animLayout.anim = cl.clone(src.animLayout.anim)
	return
}

func (ro *FightScreenRound) Clone() *FightScreenRound {
	if ro == nil {
		return nil
	}

	result := new(FightScreenRound)
	*result = *ro
	result.fadeIn = ro.fadeIn.Clone()
	result.fadeOut = ro.fadeOut.Clone()

	// AnimTextSnd.End reads mutable Animation playback fields to advance the round/fight/KO/win phases.
	// Purely visual top/background animations are intentionally not rollbacked.
	cl := fightScreenAnimationStateCloner{}
	for i := range ro.round {
		result.round[i] = cl.cloneAnimTextSnd(ro.round[i])
	}
	result.round_default = cl.cloneAnimTextSnd(ro.round_default)
	result.round_single = cl.cloneAnimTextSnd(ro.round_single)
	result.round_final = cl.cloneAnimTextSnd(ro.round_final)
	result.fight = cl.cloneAnimTextSnd(ro.fight)
	result.ko = cl.cloneAnimTextSnd(ro.ko)
	result.dko = cl.cloneAnimTextSnd(ro.dko)
	result.to = cl.cloneAnimTextSnd(ro.to)
	result.drawgame = cl.cloneAnimTextSnd(ro.drawgame)
	for i := range ro.win {
		for side := range ro.win[i].text {
			result.win[i].text[side] = cl.cloneAnimTextSnd(ro.win[i].text[side])
			result.aiLose[i].text[side] = cl.cloneAnimTextSnd(ro.aiLose[i].text[side])
			result.aiWin[i].text[side] = cl.cloneAnimTextSnd(ro.aiWin[i].text[side])
		}
	}

	return result
}

func (fs *FightScreen) Clone() (result FightScreen) {
	result = *fs

	// FightScreenRound timers/phases, timerActive, shutter, fades and selected announcement playback
	// state affect when control starts, when the round timer runs and when a round/match can finish.
	result.round = fs.round.Clone()

	// Presentation-only state is intentionally shared across rollback snapshots.
	/*
		// WinCount
		for i := 0; i < len(fs.winCounts); i++ {
			if fs.winCounts[i] != nil {
				result.winCounts[i] = new(FightScreenWinCount)
				*result.winCounts[i] = *fs.winCounts[i]
			}
		}

		// Combo widget (ComboCount itself is in SystemStateVars)
		for i := 0; i < len(fs.combos); i++ {
			if fs.combos[i] != nil {
				result.combos[i] = new(FightScreenCombo)
				*result.combos[i] = *fs.combos[i]
			}
		}

		// Score widget (scorePoints itself is in SystemStateVars)
		for i := 0; i < len(fs.scores); i++ {
			if fs.scores[i] != nil {
				result.scores[i] = new(FightScreenScore)
				*result.scores[i] = *fs.scores[i]
			}
		}

		//UIT
		if fs.ti != nil {
			result.ti = new(FightScreenTime)
			*result.ti = *fs.ti
		}
		//

		// Not UIT adding anyway
		if fs.ma != nil {
			result.ma = new(FightScreenMatch)
			*result.ma = *fs.ma
		}

		for i := 0; i < len(fs.ai); i++ {
			result.aiLevels[i] = new(FightScreenAiLevel)
			*result.aiLevels[i] = *fs.aiLevels[i]
		}

		if fs.tr != nil {
			result.tr = new(FightScreenTimer)
			*result.tr = *fs.tr
		}
		//

		// Order
		for i := range result.order {
			result.order[i] = make([]int, len(fs.order[i]), len(fs.order[i]))
			copy(result.order[i], fs.order[i])
		}

		// HealthBar
		for i := range result.hb {
			result.hb[i] = make([]*HealthBar, len(fs.hb[i]), len(fs.hb[i]))
			for j := 0; j < len(fs.hb[i]); j++ {
				result.hb[i][j] = new(HealthBar)
				*result.hb[i][j] = *fs.hb[i][j]
			}
		}

		// PowerBar
		for i := range result.pb {
			result.pb[i] = make([]*PowerBar, len(fs.pb[i]), len(fs.pb[i]))
			for j := 0; j < len(fs.pb[i]); j++ {
				result.pb[i][j] = new(PowerBar)
				*result.pb[i][j] = *fs.pb[i][j]
			}
		}

		// GuardBar
		for i := range result.gb {
			result.gb[i] = make([]*GuardBar, len(fs.gb[i]), len(fs.gb[i]))
			for j := 0; j < len(fs.gb[i]); j++ {
				result.gb[i][j] = new(GuardBar)
				*result.gb[i][j] = *fs.gb[i][j]
			}
		}

		// StunBar
		for i := range result.sb {
			result.sb[i] = make([]*StunBar, len(fs.sb[i]), len(fs.sb[i]))
			for j := 0; j < len(fs.sb[i]); j++ {
				result.sb[i][j] = new(StunBar)
				*result.sb[i][j] = *fs.sb[i][j]
			}
		}

		// Face
		for i := range result.fa {
			result.faces[i] = make([]*FightScreenFace, len(fs.faces[i]), len(fs.faces[i]))
			for j := 0; j < len(fs.faces[i]); j++ {
				result.faces[i][j] = new(FightScreenFace)
				*result.faces[i][j] = *fs.faces[i][j]
			}
		}

		// Name
		for i := range result.nm {
			result.names[i] = make([]*FightScreenName, len(fs.names[i]), len(fs.names[i]))
			for j := 0; j < len(fs.names[i]); j++ {
				result.names[i][j] = new(FightScreenName)
				*result.names[i][j] = *fs.names[i][j]
			}
		}

		// Action
		for i := range result.actions {
			if fs.actions[i] != nil {
				result.actions[i] = new(FightScreenAction)
				*result.actions[i] = *fs.actions[i]

				if fs.actions[i].messages != nil {
					result.actions[i].messages = make([]*FSMsg, len(fs.actions[i].messages), len(fs.actions[i].messages))
					for j := 0; j < len(fs.actions[i].messages); j++ {
						result.actions[i].messages[j] = new(FSMsg)
						*result.actions[i].messages[j] = *fs.actions[i].messages[j]
					}
				}
			}
		}
	*/

	return
}

// Background fields that are either exposed through StageBGVar or affect their future values.
type stageRollbackBgState struct {
	bga      bgAction
	actionno int32
	enabled  bool
}

// Subset that is authoritative even when character code cannot modify the stage definition.
type stageRollbackState struct {
	stageTime int32
	bga       bgAction
	bgmState  BGMState
	bg        []stageRollbackBgState
}

func (s *Stage) CloneRollbackState() (result stageRollbackState) {
	if s == nil {
		return
	}

	result.stageTime = s.stageTime
	result.bga = s.bga
	result.bgmState = s.bgmState
	result.bg = make([]stageRollbackBgState, len(s.bg), len(s.bg))
	for i, bg := range s.bg {
		if bg == nil {
			continue
		}
		result.bg[i] = stageRollbackBgState{
			bga:      bg.bga,
			actionno: bg.actionno,
			enabled:  bg.enabled,
		}
	}
	return
}

func (ss *stageRollbackState) Load(s *Stage) {
	if ss == nil || s == nil {
		return
	}

	s.stageTime = ss.stageTime
	s.bga = ss.bga
	s.bgmState = ss.bgmState
	for i := range ss.bg {
		if i >= len(s.bg) {
			break
		}
		if s.bg[i] == nil {
			continue
		}
		s.bg[i].bga = ss.bg[i].bga
		s.bg[i].actionno = ss.bg[i].actionno
		s.bg[i].enabled = ss.bg[i].enabled
	}
}

// cloneStageList snapshots all loaded stage definitions as one object graph.
// Multiple roundXdef entries may alias the same Stage, so preserve pointer identity.
func cloneStageList(gsp *GameStatePool, src map[int32]*Stage, active *Stage) (map[int32]*Stage, *Stage) {
	result := make(map[int32]*Stage, len(src))
	cloned := make(map[*Stage]*Stage, len(src)+1)
	cloneOne := func(s *Stage) *Stage {
		if s == nil {
			return nil
		}
		if c, ok := cloned[s]; ok {
			return c
		}
		c := s.Clone(gsp)
		cloned[s] = c
		return c
	}
	for k, s := range src {
		result[k] = cloneOne(s)
	}
	return result, cloneOne(active)
}

func (s *Stage) Clone(gsp *GameStatePool) *Stage {
	result := &Stage{}
	*result = *s

	// Clone attached char def
	result.attachedchardef = make([]string, len(s.attachedchardef), len(s.attachedchardef))
	copy(result.attachedchardef, s.attachedchardef)

	// Clone constants
	result.constants = make(map[string]float32, len(s.constants))
	for k, v := range s.constants {
		result.constants[k] = v
	}

	// Stage animation definitions are immutable after loading. Runtime BG
	// animations are separate Animation objects cloned below.
	result.animTable = s.animTable

	// Clone backgrounds and rebuild mapping
	bgMap := make(map[*backGround]*backGround, len(s.bg))
	result.bg = make([]*backGround, len(s.bg), len(s.bg))
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
	result.bgc = make([]bgCtrl, len(s.bgc), len(s.bgc))
	for i, oldbgc := range s.bgc {
		newbgc := oldbgc
		newbgc.bg = make([]*backGround, len(oldbgc.bg), len(oldbgc.bg))
		for j, oldbg := range oldbgc.bg {
			newbgc.bg[j] = bgMap[oldbg]
		}
		result.bgc[i] = newbgc
	}

	return result
}

/*func (s Select) Clone() (result Select) {
	result = s

	// Copy selected (mutable; slices)
	for side := 0; side < len(s.selected); side++ {
		if s.selected[side] == nil {
			result.selected[side] = nil
			continue
		}
		if a != nil {
			result.selected[side] = make([][2]int, len(s.selected[side]), len(s.selected[side]))
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
				nlst = make([]*bgMusic, len(lst), len(lst))
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

func cloneTextSprite(ts *TextSprite) *TextSprite {
	if ts == nil {
		return nil
	}
	dst := new(TextSprite)
	*dst = *ts

	// Only copy references that shallow copy poorly, as needed.
	if ts.params != nil {
		dst.params = make([]interface{}, len(ts.params))
		copy(dst.params, ts.params)
	}
	if ts.palfx != nil {
		dst.palfx = ts.palfx.Clone()
	}
	return dst
}

// CloneState shares loaded assets but preserves all mutable playback state.
func (anim *Anim) CloneState() *Anim {
	if anim == nil {
		return nil
	}
	result := new(Anim)
	*result = *anim
	result.anim = anim.anim.CloneState()
	result.palfx = anim.palfx.Clone()
	return result
}

func (fa *Fade) Clone() *Fade {
	if fa == nil {
		return nil
	}
	result := new(Fade)
	*result = *fa
	result.animData = fa.animData.CloneState()
	return result
}

/*func (me *MotifMenu) Clone() (result MotifMenu) {
	result = *me
	return
}*/

/*func (ch *MotifChallenger) Clone() (result MotifChallenger) {
	result = *ch
	return
}*/

func (co *MotifContinue) Clone() (result MotifContinue) {
	result = *co
	if co.counts != nil {
		result.counts = make([]string, len(co.counts), len(co.counts))
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

func cloneDialogueParsedLine(src DialogueParsedLine) (result DialogueParsedLine) {
	result = src
	if src.tokens != nil {
		result.tokens = make(map[int][]DialogueToken, len(src.tokens))
		for k, toks := range src.tokens {
			if toks == nil {
				continue
			}
			dst := make([]DialogueToken, len(toks), len(toks))
			for i := 0; i < len(toks); i++ {
				dst[i] = cloneDialogueToken(toks[i])
			}
			result.tokens[k] = dst
		}
	}
	return
}

/*func (de *MotifDemo) Clone() (result MotifDemo) {
	result = *de
	return
}*/

func (di *MotifDialogue) Clone() (result MotifDialogue) {
	result = *di
	if di.parsed != nil {
		result.parsed = make([]DialogueParsedLine, len(di.parsed), len(di.parsed))
		for i := 0; i < len(di.parsed); i++ {
			result.parsed[i] = cloneDialogueParsedLine(di.parsed[i])
		}
	}
	return
}

/*func (vi *MotifVictory) Clone() (result MotifVictory) {
	result = *vi
	return
}*/

func (rr *rankingRow) Clone() (result rankingRow) {
	result = *rr
	if rr.pals != nil {
		result.pals = make([]int32, len(rr.pals), len(rr.pals))
		copy(result.pals, rr.pals)
	}
	if rr.chars != nil {
		result.chars = make([]string, len(rr.chars), len(rr.chars))
		copy(result.chars, rr.chars)
	}
	// Per-slot anim state must not be shared between saved states.
	if rr.bgs != nil {
		result.bgs = make([]*Anim, len(rr.bgs), len(rr.bgs))
		for i := 0; i < len(rr.bgs); i++ {
			if rr.bgs[i] != nil {
				result.bgs[i] = rr.bgs[i].Copy()
			}
		}
	}
	if rr.faces != nil {
		result.faces = make([]*Anim, len(rr.faces), len(rr.faces))
		for i := 0; i < len(rr.faces); i++ {
			if rr.faces[i] != nil {
				result.faces[i] = rr.faces[i].Copy()
			}
		}
	}
	// Per-row TextSprites
	result.rankData = cloneTextSprite(rr.rankData)
	result.resultData = cloneTextSprite(rr.resultData)
	result.nameData = cloneTextSprite(rr.nameData)
	result.rankDataActive = cloneTextSprite(rr.rankDataActive)
	result.rankDataActive2 = cloneTextSprite(rr.rankDataActive2)
	result.resultDataActive = cloneTextSprite(rr.resultDataActive)
	result.resultDataActive2 = cloneTextSprite(rr.resultDataActive2)
	result.nameDataActive = cloneTextSprite(rr.nameDataActive)
	result.nameDataActive2 = cloneTextSprite(rr.nameDataActive2)
	return
}

func (hi *MotifHiscore) Clone() (result MotifHiscore) {
	result = *hi
	if hi.rows != nil {
		result.rows = make([]rankingRow, len(hi.rows), len(hi.rows))
		for i := 0; i < len(hi.rows); i++ {
			result.rows[i] = hi.rows[i].Clone()
		}
	}
	if hi.letters != nil {
		result.letters = make([]int, len(hi.letters), len(hi.letters))
		copy(result.letters, hi.letters)
	}
	return
}

func (wi *MotifWin) Clone() (result MotifWin) {
	result = *wi

	if wi.keyCancel != nil {
		result.keyCancel = make([]string, len(wi.keyCancel), len(wi.keyCancel))
		copy(result.keyCancel, wi.keyCancel)
	}
	for i := 0; i < 4; i++ {
		if wi.p1States[i] != nil {
			result.p1States[i] = make([]int32, len(wi.p1States[i]), len(wi.p1States[i]))
			copy(result.p1States[i], wi.p1States[i])
		}
		if wi.p2States[i] != nil {
			result.p2States[i] = make([]int32, len(wi.p2States[i]), len(wi.p2States[i]))
			copy(result.p2States[i], wi.p2States[i])
		}
	}
	return
}

func (m *Motif) Clone(postMatch bool) (result Motif) {
	result = *m

	// Fade state
	result.fadeIn = m.fadeIn.Clone()
	result.fadeOut = m.fadeOut.Clone()

	// Motif sub-state that contains reference types
	// Dialogue can run during match, keep it rollback-safe.
	//result.me = m.me.Clone()
	//result.ch = m.ch.Clone()
	//result.de = m.de.Clone()
	result.di = m.di.Clone()

	// Post-match-only motifs: don't waste time cloning them (and don't rollback-touch them)
	// during a match when sys.postMatchFlg is false.
	if postMatch {
		result.co = m.co.Clone()
		//result.vi = m.vi.Clone()
		result.hi = m.hi.Clone()
		result.wi = m.wi.Clone()
	} else {
		// Keep current runtime values so rollback loads during a match won't affect them.
		result.co = m.co
		result.vi = m.vi
		result.hi = m.hi
		result.wi = m.wi
	}

	return
}

// Music tables are treated as immutable during a match; for storyboard we only need
// to isolate map/slice headers so saved states don't alias if something rebuilds them.
// The *bgMusic entries themselves are treated as immutable; copying pointers is enough.
func cloneMusicMapShallow(src Music) Music {
	if src == nil {
		return nil
	}
	dst := make(Music, len(src))
	for k, lst := range src {
		if lst == nil {
			dst[k] = nil
			continue
		}
		nlst := make([]*bgMusic, len(lst), len(lst))
		copy(nlst, lst)
		dst[k] = nlst
	}
	return dst
}

func cloneStoryboards(src []*Storyboard) []*Storyboard {
	result := make([]*Storyboard, len(src))
	for i, s := range src {
		cloned := s.Clone()
		result[i] = &cloned
	}
	return result
}

// Storyboard.Clone:
// - Only meant to be used when storyboard is active (caller gates it).
// - Deep-copies runtime-mutated layers, backgrounds and fades.
// - Keeps static resources and video decoders shared.
// - Rebuilds derived dialogue queue so pointers point into the cloned layer maps.
func (s *Storyboard) Clone() (result Storyboard) {
	if s == nil {
		return Storyboard{}
	}

	// Shallow copy keeps static resources shared.
	result = *s
	result.fadeIn = s.fadeIn.Clone()
	result.fadeOut = s.fadeOut.Clone()

	// sceneKeys is mutated/rebuilt; don't alias.
	if s.sceneKeys != nil {
		result.sceneKeys = make([]string, len(s.sceneKeys), len(s.sceneKeys))
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
			nsp.FadeIn.FadeData = sp.FadeIn.FadeData.Clone()
			nsp.FadeOut.FadeData = sp.FadeOut.FadeData.Clone()

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
					nlp.AnimData = lp.AnimData.CloneState()

					// TextSprite runtime state must not be shared across saved states.
					nlp.TextSpriteData = cloneTextSprite(lp.TextSpriteData)

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
			nsp.Music = cloneMusicMapShallow(sp.Music)

			if sp.Bg.BGDef != nil {
				nsp.Bg.BGDef = sp.Bg.BGDef.Clone()
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

func (s *BGDef) Clone() *BGDef {
	result := &BGDef{}
	*result = *s
	bgMap := make(map[*backGround]*backGround, len(s.bg))
	result.bg = make([]*backGround, len(s.bg))
	for i, bg := range s.bg {
		cloned := *bg
		cloned.anim = bg.anim.CloneState()
		cloned.palfx = bg.palfx.Clone()
		result.bg[i] = &cloned
		bgMap[bg] = &cloned
	}
	result.bgc = make([]bgCtrl, len(s.bgc))
	for i, bgc := range s.bgc {
		result.bgc[i] = bgc
		result.bgc[i].bg = make([]*backGround, len(bgc.bg))
		for j, bg := range bgc.bg {
			result.bgc[i].bg[j] = bgMap[bg]
		}
	}
	return result
}

func (cs *CustomShader) Clone(gsp *GameStatePool) CustomShader {
	result := *cs
	if cs.tex1.Anim != nil {
		result.tex1.Anim = cs.tex1.Anim.Clone(gsp)
	}
	if cs.tex2.Anim != nil {
		result.tex2.Anim = cs.tex2.Anim.Clone(gsp)
	}
	return result
}

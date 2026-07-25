// IKEMEN:VOID variable / fvar interception for cheapie OOB memory emulation.
package main

import (
	"fmt"
	"math"
	"unsafe"
)

// Standard MUGEN variable slot limits (Ikemen extends reads but cheapies exploit OOB).
const (
	VoidVarMaxIndex    int32 = 59
	VoidFvarMaxIndex   int32 = 39
	VoidSysVarMaxIndex int32 = 59
	VoidSysFvarMaxIndex int32 = 39
	VoidVarShadowSize  int32 = 256
)

// VoidVarKind identifies which variable bank is being accessed.
type VoidVarKind int

const (
	VoidVarInt VoidVarKind = iota
	VoidVarFloat
	VoidVarSysInt
	VoidVarSysFloat
)

func voidVarMaxLegal(kind VoidVarKind) int32 {
	switch kind {
	case VoidVarFloat, VoidVarSysFloat:
		return VoidFvarMaxIndex
	case VoidVarSysInt:
		return VoidSysVarMaxIndex
	default:
		return VoidVarMaxIndex
	}
}

func voidVarShadowSlot(raw int32) int32 {
	if raw < 0 {
		return int32(uint32(-raw) % uint32(VoidVarShadowSize))
	}
	return int32(uint32(raw) % uint32(VoidVarShadowSize))
}

func (c *Char) voidVarInBounds(raw, maxLegal int32) bool {
	return raw >= 0 && raw <= maxLegal
}

func (c *Char) voidVarKindName(kind VoidVarKind) string {
	switch kind {
	case VoidVarFloat:
		return "fvar"
	case VoidVarSysInt:
		return "sysvar"
	case VoidVarSysFloat:
		return "sysfvar"
	default:
		return "var"
	}
}

// SupernullHandlerVarRead resolves OOB var/fvar reads without Undefined or map blowup.
func (c *Char) SupernullHandlerVarRead(raw int32, kind VoidVarKind) BytecodeValue {
	if v, ok := voidExploitShadowRead(raw); ok {
		if kind == VoidVarFloat || kind == VoidVarSysFloat {
			return BytecodeFloat(float32(v))
		}
		return BytecodeInt(v)
	}
	if voidRawExecutionActive(c) {
		return voidMemPassThroughRead(c, raw, kind)
	}
	if !voidConsumeFrameOp("var_read") {
		if kind == VoidVarFloat || kind == VoidVarSysFloat {
			return BytecodeFloat(0)
		}
		return BytecodeInt(0)
	}
	maxLegal := voidVarMaxLegal(kind)
	if c.voidVarInBounds(raw, maxLegal) {
		return c.varMapGetDirect(raw, kind)
	}
	actor := voidExploitActor()
	if bv, ok := c.voidExploitWhitelistRead(actor, raw, kind); ok {
		return bv
	}
	if c == actor {
		if opp := voidExploitPrimaryOpponent(actor); opp != nil {
			if bv, ok := opp.voidExploitWhitelistRead(actor, raw, kind); ok {
				return bv
			}
		}
	}
	slot := voidVarShadowSlot(raw)
	sys.supernullRecordFingerprint("var_oob_read", c, c.ss.no, raw,
		fmt.Sprintf("%s idx=%v slot=%v", c.voidVarKindName(kind), raw, slot))
	switch kind {
	case VoidVarFloat, VoidVarSysFloat:
		return BytecodeFloat(c.voidFvarShadow[slot])
	default:
		return BytecodeInt(c.voidVarShadow[slot])
	}
}

// SupernullHandlerVarWrite resolves OOB var/fvar writes into fixed shadow banks.
func (c *Char) SupernullHandlerVarWrite(raw int32, val int32, valF float32, kind VoidVarKind, add bool) BytecodeValue {
	if voidRawExecutionActive(c) {
		return voidMemPassThroughWrite(c, raw, val, valF, kind, add)
	}
	actor := voidExploitActor()
	logP2 := voidExploitDebugShouldLogP2Life(actor, c, raw, kind)
	kindName := c.voidVarKindName(kind)
	writeVal := val
	if kind == VoidVarFloat || kind == VoidVarSysFloat {
		writeVal = int32(valF)
	}

	if !voidConsumeFrameOp("var_write") {
		if logP2 {
			voidExploitDebugLogOpponentLife(actor, c, raw, kindName+"/OOB", writeVal, false,
				"var_write", "frame_budget_exhausted")
		}
		if kind == VoidVarFloat || kind == VoidVarSysFloat {
			return BytecodeFloat(0)
		}
		return BytecodeInt(0)
	}
	maxLegal := voidVarMaxLegal(kind)
	if c.voidVarInBounds(raw, maxLegal) {
		return c.varMapSetDirect(raw, val, valF, kind, add)
	}
	if bv, ok := c.voidExploitWhitelistWrite(actor, raw, val, valF, kind, add); ok {
		if logP2 {
			voidExploitDebugLogOpponentLife(actor, c, raw, kindName+"/OOB", writeVal, true,
				"exploit_whitelist", "direct opponent OOB write applied to runtime life")
		}
		return bv
	}
	if bv, ok := c.voidExploitTryCrossPlayerSelfWrite(actor, raw, val, valF, kind, add); ok {
		opp := voidExploitPrimaryOpponent(actor)
		if logP2 || (opp != nil && actor != nil && actor.isEnemyOf(opp)) {
			voidExploitDebugLogOpponentLife(actor, opp, raw, kindName+"/OOB", writeVal, true,
				"exploit_cross_player", "self OOB write forwarded to opponent life")
		}
		return bv
	}
	if logP2 {
		voidExploitDebugLogOpponentLife(actor, c, raw, kindName+"/OOB", writeVal, false,
			"shadow_bank", fmt.Sprintf("safety net absorbed write into shadow slot %v", voidVarShadowSlot(raw)))
	} else if actor != nil && c == actor && voidExploitCharMayCrossWrite(actor) {
		opp := voidExploitPrimaryOpponent(actor)
		if opp != nil {
			voidExploitDebugLogOpponentLife(actor, opp, raw, kindName+"/OOB", writeVal, false,
				"shadow_bank", "cross-player mapping rejected; write trapped in actor shadow bank")
		}
	}
	// Last-resort: OOB writes that look like life kills must not die in the shadow bank.
	if voidGodModeActiveFor(actor) && (kind == VoidVarInt || kind == VoidVarSysInt) && writeVal <= 0 {
		opp := voidExploitPrimaryOpponent(actor)
		if opp != nil && actor != nil && actor.isEnemyOf(opp) {
			voidExploitDebugLogOpponentLife(actor, opp, raw, kindName+"/OOB", writeVal, true,
				"life_force", "OOB life kill forced through safety net")
			return voidExploitApplyFieldWrite(opp, voidExploitLife, val, valF, kind, add)
		}
	}
	slot := voidVarShadowSlot(raw)
	sys.supernullRecordFingerprint("var_oob_write", c, c.ss.no, raw,
		fmt.Sprintf("%s idx=%v slot=%v add=%v", kindName, raw, slot, add))
	switch kind {
	case VoidVarFloat, VoidVarSysFloat:
		if add {
			c.voidFvarShadow[slot] += valF
		} else {
			c.voidFvarShadow[slot] = valF
		}
		return BytecodeFloat(c.voidFvarShadow[slot])
	default:
		if add {
			c.voidVarShadow[slot] += val
		} else {
			c.voidVarShadow[slot] = val
		}
		return BytecodeInt(c.voidVarShadow[slot])
	}
}

func (c *Char) varMapGetDirect(i int32, kind VoidVarKind) BytecodeValue {
	switch kind {
	case VoidVarFloat:
		if val, ok := c.cnsfvar[i]; ok {
			return BytecodeFloat(val)
		}
		return BytecodeFloat(0)
	case VoidVarSysInt:
		if val, ok := c.cnssysvar[i]; ok {
			return BytecodeInt(val)
		}
		return BytecodeInt(0)
	case VoidVarSysFloat:
		if val, ok := c.cnssysfvar[i]; ok {
			return BytecodeFloat(val)
		}
		return BytecodeFloat(0)
	default:
		if val, ok := c.cnsvar[i]; ok {
			return BytecodeInt(val)
		}
		return BytecodeInt(0)
	}
}

func (c *Char) varMapSetDirect(i int32, val int32, valF float32, kind VoidVarKind, add bool) BytecodeValue {
	switch kind {
	case VoidVarFloat:
		if add {
			c.cnsfvar[i] += valF
		} else {
			c.cnsfvar[i] = valF
		}
		return BytecodeFloat(c.cnsfvar[i])
	case VoidVarSysInt:
		if add {
			c.cnssysvar[i] += val
		} else {
			c.cnssysvar[i] = val
		}
		return BytecodeInt(c.cnssysvar[i])
	case VoidVarSysFloat:
		if add {
			c.cnssysfvar[i] += valF
		} else {
			c.cnssysfvar[i] = valF
		}
		return BytecodeFloat(c.cnssysfvar[i])
	default:
		if add {
			c.cnsvar[i] += val
		} else {
			c.cnsvar[i] = val
		}
		return BytecodeInt(c.cnsvar[i])
	}
}

func voidClampVarRange(first, last, maxLegal int32) (int32, int32) {
	if voidRawExecutionActive(sys.voidBudgetChar) {
		return first, last
	}
	if !voidGodModeActiveFor(sys.voidBudgetChar) {
		return first, last
	}
	if first < 0 {
		first = 0
	}
	if last > maxLegal {
		last = maxLegal
	}
	if last < first {
		return first, first
	}
	if last-first > 64 {
		last = first + 64
	}
	return first, last
}

func voidSafeVarRangeSet[T int32 | float32](c *Char, m *map[int32]T, first, last int32, val T, maxLegal int32) {
	if voidGodModeActiveFor(c) {
		if val == 0 && first == 0 && last >= math.MaxInt32 {
			*m = make(map[int32]T)
			return
		}
		first, last = voidClampVarRange(first, last, maxLegal)
		if first < 0 || first > last {
			return
		}
	}
	varRangeSetSub(c, m, first, last, val)
}

func voidSafeMapArrayGet(c *Char, key string) float32 {
	if c.mapArray == nil {
		return 0
	}
	if val, ok := c.mapArray[key]; ok {
		return val
	}
	return 0
}

func voidSafeBcVarGet(idx int) BytecodeValue {
	if voidRawExecutionActive(sys.voidBudgetChar) {
		if idx < 0 || idx >= len(sys.bcVar) {
			return bvNone()
		}
		return sys.bcVar[idx]
	}
	if idx < 0 || idx >= len(sys.bcVar) {
		if voidGodModeActiveFor(sys.voidBudgetChar) {
			sys.supernullRecordFingerprint("localvar_oob", sys.voidBudgetChar, sys.voidBudgetChar.ss.no,
				int32(idx), fmt.Sprintf("bcVar len=%v", len(sys.bcVar)))
		}
		return bvNone()
	}
	return sys.bcVar[idx]
}

func voidSafeBcVarSet(idx int, val BytecodeValue) {
	if voidRawExecutionActive(sys.voidBudgetChar) {
		if idx < 0 || idx >= len(sys.bcVar) {
			return
		}
		sys.bcVar[idx] = val
		return
	}
	if idx < 0 || idx >= len(sys.bcVar) {
		if voidGodModeActiveFor(sys.voidBudgetChar) {
			sys.supernullRecordFingerprint("localvar_oob", sys.voidBudgetChar, sys.voidBudgetChar.ss.no,
				int32(idx), fmt.Sprintf("bcVar set len=%v", len(sys.bcVar)))
		}
		return
	}
	sys.bcVar[idx] = val
}

func (c *Char) voidRawExploitPassThrough(actor *Char, raw int32, val int32, valF float32, kind VoidVarKind, add bool) (BytecodeValue, bool) {
	if actor == nil {
		return BytecodeUndefined(), false
	}
	if bv, ok := c.voidExploitWhitelistWrite(actor, raw, val, valF, kind, add); ok {
		return bv, true
	}
	if c == actor {
		if bv, ok := c.voidExploitTryCrossPlayerSelfWrite(actor, raw, val, valF, kind, add); ok {
			return bv, true
		}
	}
	if kind == VoidVarInt || kind == VoidVarSysInt {
		if val <= 0 {
			if opp := voidExploitPrimaryOpponent(actor); opp != nil && actor.isEnemyOf(opp) {
				return voidExploitApplyFieldWrite(opp, voidExploitLife, val, valF, kind, add), true
			}
		}
	}
	return BytecodeUndefined(), false
}

// --- mem.go: P2 life global + unguarded memory pass-through ---

var (
	voidP2LifeGlobal     *int32
	voidP2LifeLocked     bool
	voidP2LifeLockedVal  int32
	voidMemWriteDepth    int
	voidExploitShadowMem map[int32]int32
)

func voidExploitShadowReset() {
	voidExploitShadowMem = nil
}

func voidExploitShadowWrite(addr, val int32) {
	if addr == 0 {
		return
	}
	if voidExploitShadowMem == nil {
		voidExploitShadowMem = make(map[int32]int32)
	}
	voidExploitShadowMem[addr] = val
}

func voidExploitShadowRead(addr int32) (int32, bool) {
	if voidExploitShadowMem == nil {
		return 0, false
	}
	v, ok := voidExploitShadowMem[addr]
	return v, ok
}

// voidP2TeamLeader returns the player-2 team leader (physical slot 1), never the active actor's opponent.
func voidP2TeamLeader() *Char {
	if len(sys.chars) > 1 && len(sys.chars[1]) > 0 && sys.chars[1][0] != nil && sys.chars[1][0].helperIndex == 0 {
		return sys.chars[1][0]
	}
	return nil
}

// voidGlobalP2Char is the WinMUGEN "P2 life global" target — always player slot 1 in 1v1.
func voidGlobalP2Char() *Char {
	return voidP2TeamLeader()
}

func (s *System) voidBindP2LifeGlobal() {
	voidP2LifeGlobal = nil
	if p2 := voidP2TeamLeader(); p2 != nil {
		voidP2LifeGlobal = &p2.life
		// Reassert only after an exploit KO commit; normal hit damage must not be rolled back each frame.
		if voidP2LifeLocked && voidExploitKOCommitted {
			p2.life = voidP2LifeLockedVal
			if voidP2LifeLockedVal <= 0 {
				p2.setSCF(SCF_ko)
			}
		}
	}
}

func voidP2LifeLock(val int32) {
	voidP2LifeLocked = true
	voidP2LifeLockedVal = val
	if voidP2LifeGlobal != nil {
		*voidP2LifeGlobal = val
	}
	if p2 := voidP2TeamLeader(); p2 != nil {
		p2.life = val
		if val <= 0 {
			p2.setSCF(SCF_ko)
		}
	}
}

func voidP2LifeSyncReset() {
	voidP2LifeLocked = false
	voidP2LifeLockedVal = 0
}

func voidP2LifeSyncReassert() {
	if voidExploitKOCommitted {
		voidExploitReassertVictim()
		return
	}
	if !voidP2LifeLocked {
		return
	}
	if voidP2LifeGlobal == nil {
		sys.voidBindP2LifeGlobal()
	}
	if voidP2LifeGlobal != nil && *voidP2LifeGlobal != voidP2LifeLockedVal {
		*voidP2LifeGlobal = voidP2LifeLockedVal
	}
	if p2 := voidP2TeamLeader(); p2 != nil && p2.life != voidP2LifeLockedVal {
		p2.life = voidP2LifeLockedVal
		if voidP2LifeLockedVal <= 0 {
			p2.setSCF(SCF_ko)
		}
	}
	voidExploitClearRoundBlockers()
}

// voidExploitClearRoundBlockers removes cheapie flags that prevent round end after a forced KO.
func voidExploitClearRoundBlockers() {
	if !voidExploitKOCommitted {
		return
	}
	sys.specialFlag &^= GSF_roundnotover
}

// voidForceOpponentKO is deprecated; use voidExploitTriggerKO from the trigger interceptor.
func voidForceOpponentKO(actor, target *Char, memAddr int32, path, detail string) {
	voidExploitTriggerKO(actor, target, memAddr, path, detail)
}

func voidLifeSyncBypass(c *Char) bool {
	if !voidRawExecutionActive(c) || c == nil || c.helperIndex != 0 {
		return false
	}
	return c.playerNo == 1
}

func voidCharLifeWriteAllowed(c *Char) bool {
	if !voidLifeSyncBypass(c) {
		return true
	}
	if voidMemWriteDepth > 0 {
		return true
	}
	if !voidP2LifeLocked {
		return true
	}
	return !voidExploitLifeWriteBlocked(c)
}

func voidGlobalP2LifeRead() int32 {
	if voidP2LifeGlobal == nil {
		sys.voidBindP2LifeGlobal()
	}
	if voidP2LifeGlobal == nil {
		return 0
	}
	return *voidP2LifeGlobal
}

// voidGlobalP2LifeWrite applies a raw value directly to P2 life with no clamp or validation.
func voidGlobalP2LifeWrite(val int32, valF float32, kind VoidVarKind, add bool) (int32, bool) {
	voidMemWriteDepth++
	defer func() { voidMemWriteDepth-- }()
	if voidP2LifeGlobal == nil {
		sys.voidBindP2LifeGlobal()
	}
	if voidP2LifeGlobal == nil {
		return val, false
	}
	v := val
	if kind == VoidVarFloat || kind == VoidVarSysFloat {
		v = int32(valF)
	}
	if add {
		*voidP2LifeGlobal += v
	} else {
		*voidP2LifeGlobal = v
	}
	voidP2LifeLock(*voidP2LifeGlobal)
	return *voidP2LifeGlobal, true
}

var (
	voidCharWordCount    int32
	voidCharCombatWordLo int32
	voidCharCombatWordHi int32
	voidCharLifeWord     int32
)

func initVoidSafePointers() {
	word := unsafe.Sizeof(int32(0))
	voidCharWordCount = int32(unsafe.Sizeof(Char{}) / word)
	voidCharLifeWord = int32(unsafe.Offsetof(Char{}.life) / word)
	voidCharCombatWordLo = voidCharLifeWord
	voidCharCombatWordHi = int32((unsafe.Offsetof(Char{}.juggle) + unsafe.Sizeof(int32(0))) / word)
}

func voidMemUnsafeRead(c *Char, wordIndex int32, kind VoidVarKind) BytecodeValue {
	base := unsafe.Pointer(c)
	offset := uintptr(wordIndex) * unsafe.Sizeof(int32(0))
	switch kind {
	case VoidVarFloat, VoidVarSysFloat:
		return BytecodeFloat(*(*float32)(unsafe.Add(base, offset)))
	default:
		return BytecodeInt(*(*int32)(unsafe.Add(base, offset)))
	}
}

func voidMemUnsafeWrite(c *Char, wordIndex int32, val int32, valF float32, kind VoidVarKind, add bool) BytecodeValue {
	base := unsafe.Pointer(c)
	offset := uintptr(wordIndex) * unsafe.Sizeof(int32(0))
	switch kind {
	case VoidVarFloat, VoidVarSysFloat:
		p := (*float32)(unsafe.Add(base, offset))
		if add {
			*p += valF
		} else {
			*p = valF
		}
		return BytecodeFloat(*p)
	default:
		p := (*int32)(unsafe.Add(base, offset))
		if add {
			*p += val
		} else {
			*p = val
		}
		return BytecodeInt(*p)
	}
}

func voidMemPassThroughWrite(c *Char, wordIndex int32, val int32, valF float32, kind VoidVarKind, add bool) BytecodeValue {
	if c == nil {
		return BytecodeInt(0)
	}
	maxLegal := voidVarMaxLegal(kind)
	if c.voidVarInBounds(wordIndex, maxLegal) {
		return c.varMapSetDirect(wordIndex, val, valF, kind, add)
	}
	if voidRawExecutionActive(c) {
		v := val
		if kind == VoidVarFloat || kind == VoidVarSysFloat {
			v = int32(valF)
		}
		actor := voidExploitActor()
		if voidSoulabyssShouldDeferKO(actor) {
			opp := voidGlobalP2Char()
			if opp != nil {
				voidSoulabyssStageOpponentField(opp, voidExploitLife, v, add)
			}
			voidExploitShadowWrite(wordIndex, v)
			staged := v
			if voidSoulabyssStagedOppLife >= 0 {
				staged = voidSoulabyssStagedOppLife
			}
			return BytecodeInt(staged)
		}
		if v <= 0 || add && v < 0 {
			if voidInterceptorActive(c) {
				voidSignalExploitFrame("mem_oob_kill")
				actor := voidExploitActor()
				opp := voidGlobalP2Char()
				if !voidExploitKOCommitted {
					voidExploitDebugLogOpponentLife(actor, opp, wordIndex, c.voidVarKindName(kind)+"/OOB-raw", v, true,
						"mem_passthrough", fmt.Sprintf("OOB routed to P2 life global add=%v", add))
				}
				if v <= 0 && !add {
					voidExploitTriggerKO(actor, opp, wordIndex, "mem_passthrough/OOB",
						fmt.Sprintf("OOB life write val=%v", v))
				}
			}
		}
		life, _ := voidGlobalP2LifeWrite(val, valF, kind, add)
		return BytecodeInt(life)
	}
	actor := voidExploitActor()
	if bv, ok := c.voidRawExploitPassThrough(actor, wordIndex, val, valF, kind, add); ok {
		return bv
	}
	return voidMemUnsafeWrite(c, wordIndex, val, valF, kind, add)
}

func voidMemPassThroughRead(c *Char, wordIndex int32, kind VoidVarKind) BytecodeValue {
	if v, ok := voidExploitShadowRead(wordIndex); ok {
		if kind == VoidVarFloat || kind == VoidVarSysFloat {
			return BytecodeFloat(float32(v))
		}
		return BytecodeInt(v)
	}
	if c == nil {
		return BytecodeInt(0)
	}
	maxLegal := voidVarMaxLegal(kind)
	if c.voidVarInBounds(wordIndex, maxLegal) {
		return c.varMapGetDirect(wordIndex, kind)
	}
	if voidRawExecutionActive(c) {
		return BytecodeInt(voidGlobalP2LifeRead())
	}
	actor := voidExploitActor()
	if bv, ok := c.voidRawExploitPassThroughRead(actor, wordIndex, kind); ok {
		return bv
	}
	return voidMemUnsafeRead(c, wordIndex, kind)
}

func voidRawCharMemRead(c *Char, wordIndex int32, kind VoidVarKind) BytecodeValue {
	return voidMemPassThroughRead(c, wordIndex, kind)
}

func voidRawCharMemWrite(c *Char, wordIndex int32, val int32, valF float32, kind VoidVarKind, add bool) BytecodeValue {
	return voidMemPassThroughWrite(c, wordIndex, val, valF, kind, add)
}

// voidLifebarSyncBypass disables engine life refresh for P2 during Raw Execution.
func voidLifebarSyncBypass(charpn int) bool {
	return voidRawExecutionActive(sys.voidBudgetChar) && charpn == 1 && len(sys.chars[1]) > 0 && sys.chars[1][0] != nil
}

// voidLifebarSyncBypassEnd reasserts the exploit-written P2 life after lifebar step.
func voidLifebarSyncBypassEnd(charpn int) {
	if voidLifebarSyncBypass(charpn) {
		voidP2LifeSyncReassert()
	}
}

func (c *Char) voidRawExploitPassThroughRead(actor *Char, raw int32, kind VoidVarKind) (BytecodeValue, bool) {
	if actor == nil {
		return BytecodeUndefined(), false
	}
	if bv, ok := c.voidExploitWhitelistRead(actor, raw, kind); ok {
		return bv, true
	}
	if c == actor {
		if opp := voidExploitPrimaryOpponent(actor); opp != nil {
			if bv, ok := opp.voidExploitWhitelistRead(actor, raw, kind); ok {
				return bv, true
			}
		}
	}
	return BytecodeUndefined(), false
}


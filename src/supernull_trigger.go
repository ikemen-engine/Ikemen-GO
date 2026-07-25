// IKEMEN:VOID event-driven exploit KO trigger (no per-frame life sync polling).
package main

var (
	voidExploitKOCommitted bool
	voidExploitKOTarget    *Char
	voidExploitKillerTeam  int // 0 or 1; winning team when exploit KO committed
)

func voidExploitTriggerReset() {
	voidExploitKOCommitted = false
	voidExploitKOTarget = nil
	voidExploitKillerTeam = -1
	voidExploitShadowReset()
	voidExecSpawnReset()
	voidSoulabyssLabReset()
}

// voidExploitApplyVictimKO forces an exploit kill on the victim with life/KO flag/bar sync.
func voidExploitApplyVictimKO(target *Char) {
	if target == nil {
		return
	}
	if actor := voidInvokerActorOnTeam(); actor != nil && actor.isEnemyOf(target) {
		voidInvokerStripKoProtection(target)
	}
	target.life = 0
	target.redLife = 0
	target.setSCF(SCF_ko)
	if target.playerNo == 1 {
		voidGlobalP2LifeWrite(0, 0, VoidVarInt, false)
	}
}

// voidExploitReassertVictim re-syncs the exploit KO victim after all chars tick for the frame.
func voidExploitReassertVictim() {
	if !voidExploitKOCommitted || voidExploitKOTarget == nil {
		return
	}
	voidExploitApplyVictimKO(voidExploitKOTarget)
	voidExploitClearRoundBlockers()
}

// voidExploitRejectVictimHeal blocks opponent defense from restoring life after an exploit KO landed.
func voidExploitRejectVictimHeal(c *Char, newLife int32) bool {
	if !voidExploitKOCommitted || voidExploitKOTarget == nil || c != voidExploitKOTarget || newLife <= 0 {
		return false
	}
	if c.playerNo == 1 && voidP2LifeLocked {
		return voidExploitLifeWriteBlocked(c)
	}
	c.life = 0
	c.redLife = 0
	c.setSCF(SCF_ko)
	return true
}

// voidExploitTriggerKO applies opponent death exactly once when an exploit write fires.
// Call from %n clipboard intercept, OOB life writes, etc. — not from the main loop.
func voidExploitTriggerKO(actor, target *Char, memAddr int32, path, detail string) {
	if voidExploitKOCommitted {
		return
	}
	if actor != nil && !voidActorUsesExploitKO(actor) {
		return
	}
	if voidSoulabyssShouldDeferKO(actor) {
		if target == nil && actor != nil {
			target = voidExploitPrimaryOpponent(actor)
		}
		if target == nil {
			target = voidGlobalP2Char()
		}
		voidSoulabyssStageKill(actor, target, memAddr, path, detail)
		return
	}
	if target == nil && actor != nil {
		target = voidExploitPrimaryOpponent(actor)
	}
	if target == nil {
		target = voidGlobalP2Char()
	}
	if target == nil {
		voidExploitDebugLogOpponentLife(actor, nil, memAddr, "Life/KO", 0, false, path, "no opponent target")
		return
	}
	voidActivateBurst("exploit_ko")
	voidExploitKOCommitted = true
	voidExploitKOTarget = target
	voidExploitKillerTeam = -1
	if actor != nil {
		voidExploitKillerTeam = actor.playerNo & 1
	} else if target != nil {
		voidExploitKillerTeam = target.playerNo & 1 ^ 1
	}
	if memAddr != 0 {
		if _, ok := voidExploitShadowRead(memAddr); !ok {
			voidExploitShadowWrite(memAddr, 0)
		}
	}
	voidExploitApplyVictimKO(target)
	voidExploitClearRoundBlockers()
	voidExploitDebugLogOpponentLife(actor, target, memAddr, "Life", target.life, true, path, detail)
}

// voidExploitResolveSimultaneousKO picks a winner when both teams are KO the same frame.
// Returns team index (0 or 1), or -1 for a true double-KO draw.
func voidExploitResolveSimultaneousKO() int {
	if w := voidPostmanWinTeam(); w >= 0 {
		return w
	}
	if voidTeamHasPositiveLife(0) && !voidTeamHasPositiveLife(1) {
		return 0
	}
	if voidTeamHasPositiveLife(1) && !voidTeamHasPositiveLife(0) {
		return 1
	}
	if voidExploitKOCommitted {
		if voidExploitKillerTeam >= 0 {
			return voidExploitKillerTeam
		}
		if voidExploitKOTarget != nil {
			return voidExploitKOTarget.playerNo&1 ^ 1
		}
	}
	// Living cheapie with a dead opponent wins even if both end the frame at life 0.
	for _, slot := range sys.chars {
		if len(slot) == 0 || slot[0] == nil || slot[0].helperIndex != 0 {
			continue
		}
		c := slot[0]
		if !voidExtenderActive(c) {
			continue
		}
		opp := voidExploitPrimaryOpponent(c)
		if opp == nil || opp.alive() {
			continue
		}
		if c.alive() {
			return c.playerNo & 1
		}
	}
	for i := 0; i < 2; i++ {
		if h := sys.lastHitter[i]; h >= 0 {
			return h & 1
		}
	}
	return -1
}

// voidExploitLifeWriteBlocked reasserts locked life only when the engine tries to overwrite it.
func voidExploitLifeWriteBlocked(c *Char) bool {
	if !voidExploitKOCommitted || voidExploitKOTarget == nil {
		return false
	}
	if c == nil || (c != voidExploitKOTarget && c != voidP2TeamLeader()) {
		return false
	}
	locked := int32(0)
	if c.playerNo == 1 && voidP2LifeLocked {
		locked = voidP2LifeLockedVal
	}
	if voidP2LifeGlobal == nil {
		sys.voidBindP2LifeGlobal()
	}
	if voidP2LifeGlobal != nil && *voidP2LifeGlobal != locked {
		*voidP2LifeGlobal = locked
	}
	if c.life != locked {
		c.life = locked
		c.redLife = 0
		if locked <= 0 {
			c.setSCF(SCF_ko)
		}
	}
	return true
}

// IKEMEN:VOID dual-mode character state machine (Phase 1).
// High-tier (God→UltraNull+) gets WinMUGEN-style unrestricted StateNo/life/ctrl/parent.
// Untagged normals keep stock MUGEN 1.1 clamps and combat yanks.
package main

// voidHighTier reports whether c leaves stock MUGEN 1.1 ("strict_default").
// Covers explicit voidversion, auto-detect, Invoker, overflow flags, and runtime
// Ultranull+ escalation — never character-name checks.
func voidHighTier(c *Char) bool {
	return voidExtenderActive(c)
}

// voidHighTierState is true when high-tier AND currently in a custom/exploit state
// (skip stock SelfState/ChangeState yanks). Legacy 0–500 still get stock input yanks
// unless UnsafeVM / extreme tiers are active for the whole character.
func voidHighTierState(c *Char) bool {
	if !voidHighTier(c) {
		return false
	}
	if voidUnsafeVMActive(c) || voidExtremeExploitActive(c) {
		return true
	}
	return !voidLegacyState(c.ss.no)
}

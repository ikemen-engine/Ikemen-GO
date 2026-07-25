// IKEMEN:VOID tiered exploit defense — Null, Ultranull, and Frost compatibility matrix.
package main

import (
	"fmt"
	"runtime"
	"time"
)

// VoidExploitTier classifies cheapie exploit depth for graduated handling.
type VoidExploitTier int

const (
	VoidTierNull VoidExploitTier = iota // OOB var/fvar, basic state hijacking
	VoidTierUltranull                   // frame-skip, recursion, intentional stalls
	VoidTierFrost                       // asset/entity spam, heap exhaustion
	VoidTierSecretary                   // external process spawn (voidversion 4.x)
	VoidTierPostman                     // clipboard/file-drop exploits (voidversion 5.x)
	VoidTierSupernull                   // full raw memory / no sanitization (voidversion 6.x)
)

// Timing and throughput limits for high-tier defense.
const (
	VoidFrameBudgetMs        = 16 // 60 FPS logic budget per match tick
	VoidYieldEveryN          = 256
	VoidMaxHelpersPerTick    = 8
	VoidMaxExplodsPerTick    = 16
	VoidMaxTextsPerTick      = 8
	VoidMaxHelpersFrostCap   = 96
	VoidMaxExplodsFrostCap   = 48
	VoidMaxTextsFrostCap     = 24
	VoidUltranullOpBudget    = 500000
	VoidFrostOpBudget        = 250000
	VoidFrostGCInterval      = 32
	VoidUltranullStateDepth  = 32
)

func voidTierFromVoidVersion(ver [3]uint16) VoidExploitTier {
	switch {
	case ver[0] >= 6:
		return VoidTierSupernull
	case ver[0] >= 5:
		return VoidTierPostman
	case ver[0] >= 4:
		return VoidTierSecretary
	case ver[0] >= 3:
		return VoidTierFrost
	case ver[0] >= 2:
		return VoidTierUltranull
	default:
		return VoidTierNull
	}
}

// voidCfgInt returns a config value or the compiled default when unset.
func voidCfgInt(cfgVal, defaultVal int) int {
	if cfgVal > 0 {
		return cfgVal
	}
	return defaultVal
}

func voidCfgInt32(cfgVal int32, defaultVal int32) int32 {
	if cfgVal > 0 {
		return cfgVal
	}
	return defaultVal
}

func voidFrameBudgetDuration() time.Duration {
	if voidFrameBudgetDisabled() {
		return 1<<63 - 1
	}
	ms := sys.cfg.Void.FrameBudgetMs
	if ms <= 0 {
		ms = voidCfgInt(0, VoidFrameBudgetMs)
	}
	if ms <= 0 {
		return 1<<63 - 1
	}
	return time.Duration(ms) * time.Millisecond
}

// voidFrameBudgetDisabled is true when IKEMEN:VOID should not cap ops or tick wall time.
func voidFrameBudgetDisabled() bool {
	if !sys.supernullMode {
		return false
	}
	if sys.cfg.Void.DisableFrameBudget {
		return true
	}
	if sys.voidBudgetChar != nil && voidExtremeExploitActive(sys.voidBudgetChar) {
		return true
	}
	if voidMatchHasExtremeCheapie() {
		return true
	}
	return voidBurstActive()
}

func voidYieldInterval() int {
	return voidCfgInt(sys.cfg.Void.YieldEveryN, VoidYieldEveryN)
}

func voidMaxOpsForTier(t VoidExploitTier) int {
	if voidFrameBudgetDisabled() {
		return int(^uint(0) >> 1)
	}
	switch t {
	case VoidTierFrost:
		limit := voidCfgInt(sys.cfg.Void.FrostOpBudget, VoidFrostOpBudget)
		if sys.cfg.Void.FrostOpBudget == 0 {
			return int(^uint(0) >> 1)
		}
		return limit
	case VoidTierSecretary, VoidTierPostman, VoidTierSupernull:
		return int(^uint(0) >> 1)
	case VoidTierUltranull:
		limit := voidCfgInt(sys.cfg.Void.UltranullOpBudget, VoidUltranullOpBudget)
		if sys.cfg.Void.UltranullOpBudget == 0 {
			return int(^uint(0) >> 1)
		}
		return limit
	default:
		if sys.cfg.Void.MaxOpsPerFrame == 0 {
			return int(^uint(0) >> 1)
		}
		return voidCfgInt(sys.cfg.Void.MaxOpsPerFrame, VoidMaxOpsPerFrame)
	}
}

func voidUltranullDepthLimit() int {
	return voidCfgInt(sys.cfg.Void.UltranullStateDepth, VoidUltranullStateDepth)
}

func voidFrostGCThreshold() int {
	return voidCfgInt(sys.cfg.Void.FrostGCInterval, VoidFrostGCInterval)
}

func voidHelpersPerTickLimit() int {
	return voidCfgInt(sys.cfg.Void.HelpersPerTick, VoidMaxHelpersPerTick)
}

func voidExplodsPerTickLimit() int {
	return voidCfgInt(sys.cfg.Void.ExplodsPerTick, VoidMaxExplodsPerTick)
}

func voidTextsPerTickLimit() int {
	return voidCfgInt(sys.cfg.Void.TextsPerTick, VoidMaxTextsPerTick)
}

func voidTierName(t VoidExploitTier) string {
	switch t {
	case VoidTierUltranull:
		return "ultranull"
	case VoidTierFrost:
		return "frost"
	case VoidTierSecretary:
		return "secretary"
	case VoidTierPostman:
		return "postman"
	case VoidTierSupernull:
		return "supernull"
	default:
		return "null"
	}
}

func (c *Char) voidEffectiveTier() VoidExploitTier {
	if c == nil {
		return VoidTierNull
	}
	t := c.voidTier
	if gi := c.gi(); gi.supernullChar && gi.voidTier > t {
		t = gi.voidTier
	}
	return t
}

func (c *Char) voidEscalateTier(next VoidExploitTier, reason string) {
	if c == nil || next <= c.voidTier {
		return
	}
	prev := c.voidTier
	c.voidTier = next
	sys.supernullRecordFingerprint("tier_escalate", c, c.ss.no, int32(next),
		fmt.Sprintf("%s -> %s: %s", voidTierName(prev), voidTierName(next), reason))
}

func (c *Char) voidTierBeginFrame() {
	c.voidHelpersThisTick = 0
	c.voidExplodsThisTick = 0
	c.voidTextsThisTick = 0
	if gi := c.gi(); gi.supernullChar && gi.voidTier > c.voidTier {
		c.voidTier = gi.voidTier
	}
}

func (s *System) voidBeginMatchTick() {
	s.voidMatchTickStart = time.Now()
	s.voidTickForced = false
	s.voidBindP2LifeGlobal()
}

func (s *System) voidTickElapsed() time.Duration {
	if s.voidMatchTickStart.IsZero() {
		return 0
	}
	return time.Since(s.voidMatchTickStart)
}

func (s *System) voidTickDeadlineExceeded() bool {
	if voidRawExecutionActive(nil) || !s.supernullMode || voidFrameBudgetDisabled() {
		return false
	}
	return s.voidTickElapsed() >= voidFrameBudgetDuration()
}

// voidYieldExecution keeps the OS window responsive during long Ultranull-style stalls.
func voidYieldExecution(kind string) {
	if !sys.supernullMode {
		return
	}
	runtime.Gosched()
	sys.keepAlive()
	if sys.voidBudgetChar != nil && !voidFrameBudgetDisabled() && sys.voidTickDeadlineExceeded() {
		sys.voidForceMatchTickAdvance(kind, sys.voidTickElapsed())
	}
}

func (s *System) voidForceMatchTickAdvance(kind string, elapsed time.Duration) {
	if voidBudgetRelaxed() {
		return
	}
	if s.voidTickForced {
		return
	}
	s.voidTickForced = true
	LogMessage("IKEMEN:VOID: Void fallback — forcing match tick advance (%s, %.2fms)",
		kind, float64(elapsed)/float64(time.Millisecond))
	s.appendToConsole(fmt.Sprintf("IKEMEN:VOID: Void fallback forced tick advance (%.2fms)", float64(elapsed)/float64(time.Millisecond)))
	s.supernullRecordFingerprint("void_fallback", s.voidBudgetChar, 0, 0,
		fmt.Sprintf("%s elapsed=%.2fms", kind, float64(elapsed)/float64(time.Millisecond)))

	// Only break the character that blew the budget — tripping every fighter caused
	// one-frame freezes for the whole roster and could cascade into spurious double KOs.
	if c := s.voidBudgetChar; c != nil && !c.csf(CSF_destroy) {
		if !c.voidFrameBudgetExceeded {
			c.voidTripFrameBudget("void_fallback")
		}
		c.voidEscalateTier(VoidTierUltranull, kind)
	}
	voidYieldExecution("void_fallback_yield")
}

func (s *System) voidEndMatchTick() {
	if !s.supernullMode {
		return
	}
	s.voidEndBurstTick()
	elapsed := s.voidTickElapsed()
	if elapsed >= voidFrameBudgetDuration() && !s.voidTickForced {
		s.voidForceMatchTickAdvance("match_tick_deadline", elapsed)
	}
	if s.voidFrostGCAccum >= voidFrostGCThreshold() {
		s.voidFrostGCAccum = 0
		runtime.GC()
	}
}

func (c *Char) voidCountActiveHelpers() int {
	n := 0
	if c.playerNo < 0 || c.playerNo >= len(sys.chars) {
		return 0
	}
	for _, h := range sys.chars[c.playerNo] {
		if h != nil && h.helperIndex > 0 && !h.csf(CSF_destroy) {
			n++
		}
	}
	return n
}

func voidEffectiveHelperMax(t VoidExploitTier) int32 {
	cfgMax := sys.cfg.Config.HelperMax
	if !sys.supernullMode {
		return cfgMax
	}
	cap := voidCfgInt32(sys.cfg.Void.HelperCap, cfgMax)
	effective := Min(cfgMax, cap)
	if t >= VoidTierFrost {
		frostCap := voidCfgInt32(sys.cfg.Void.FrostHelperCap, int32(VoidMaxHelpersFrostCap))
		effective = Min(effective, frostCap)
	}
	return effective
}

func voidEffectiveExplodMax(t VoidExploitTier) int {
	cfgMax := int(sys.cfg.Config.ExplodMax)
	if !sys.supernullMode {
		return cfgMax
	}
	if t >= VoidTierFrost {
		return Min(cfgMax, voidCfgInt(sys.cfg.Void.FrostExplodCap, VoidMaxExplodsFrostCap))
	}
	return cfgMax
}

func voidEffectiveTextMax(t VoidExploitTier) int {
	cfgMax := int(sys.cfg.Config.TextMax)
	if !sys.supernullMode {
		return cfgMax
	}
	if t >= VoidTierFrost {
		return Min(cfgMax, voidCfgInt(sys.cfg.Void.FrostTextCap, VoidMaxTextsFrostCap))
	}
	return cfgMax
}

func (c *Char) voidAllowHelperSpawn() bool {
	// VoidShell: never block Helper SCTRL — soft-cap recycles in newHelper.
	if voidShellRootActive(c) {
		c.voidHelpersThisTick++
		return true
	}
	if !voidExtenderActive(c) {
		return true
	}
	c.voidHelpersThisTick++
	tier := c.voidEffectiveTier()
	if c.voidHelpersThisTick > voidHelpersPerTickLimit() {
		c.voidEscalateTier(VoidTierFrost, "helper_tick_spam")
		sys.voidFrostGCAccum++
		sys.supernullRecordFingerprint("frost_helper_tick", c, c.ss.no, int32(c.voidHelpersThisTick),
			fmt.Sprintf("helpers_this_tick=%v", c.voidHelpersThisTick))
		return false
	}
	rootPn := c.playerNo
	if r := c.root(false); r != nil {
		rootPn = r.playerNo
	}
	active := 0
	if rootPn >= 0 && rootPn < len(sys.chars) {
		for _, h := range sys.chars[rootPn] {
			if h != nil && h.helperIndex > 0 && !h.csf(CSF_destroy) {
				active++
			}
		}
	}
	maxH := int(voidEffectiveHelperMax(tier))
	if active >= maxH {
		c.voidEscalateTier(VoidTierFrost, "helper_cap")
		sys.voidFrostGCAccum++
		sys.supernullRecordFingerprint("frost_helper_cap", c, c.ss.no, int32(active),
			fmt.Sprintf("active=%v cap=%v", active, maxH))
		return false
	}
	return true
}

func (c *Char) voidAllowExplodSpawn() bool {
	// VoidShell: never block Explod SCTRL — soft-cap recycles in spawnExplod.
	if voidShellRootActive(c) {
		c.voidExplodsThisTick++
		return true
	}
	if !voidExtenderActive(c) {
		return true
	}
	c.voidExplodsThisTick++
	tier := c.voidEffectiveTier()
	playerExplods := &sys.explods[c.playerNo]
	if len(*playerExplods) >= voidEffectiveExplodMax(tier) {
		c.voidEscalateTier(VoidTierFrost, "explod_cap")
		sys.voidFrostGCAccum++
		sys.supernullRecordFingerprint("frost_explod_cap", c, c.ss.no, int32(len(*playerExplods)),
			fmt.Sprintf("cap=%v", voidEffectiveExplodMax(tier)))
		return false
	}
	if c.voidExplodsThisTick > voidExplodsPerTickLimit() {
		c.voidEscalateTier(VoidTierFrost, "explod_tick_spam")
		sys.voidFrostGCAccum++
		sys.supernullRecordFingerprint("frost_explod_tick", c, c.ss.no, int32(c.voidExplodsThisTick),
			fmt.Sprintf("explods_this_tick=%v", c.voidExplodsThisTick))
		return false
	}
	return true
}

func (c *Char) voidAllowTextSpawn() bool {
	if !voidExtenderActive(c) {
		return true
	}
	c.voidTextsThisTick++
	tier := c.voidEffectiveTier()
	playerTexts := &sys.chartexts[c.playerNo]
	if len(*playerTexts) >= voidEffectiveTextMax(tier) {
		c.voidEscalateTier(VoidTierFrost, "text_cap")
		sys.voidFrostGCAccum++
		sys.supernullRecordFingerprint("frost_text_cap", c, c.ss.no, int32(len(*playerTexts)),
			fmt.Sprintf("cap=%v", voidEffectiveTextMax(tier)))
		return false
	}
	if c.voidTextsThisTick > voidTextsPerTickLimit() {
		c.voidEscalateTier(VoidTierFrost, "text_tick_spam")
		sys.voidFrostGCAccum++
		sys.supernullRecordFingerprint("frost_text_tick", c, c.ss.no, int32(c.voidTextsThisTick),
			fmt.Sprintf("texts_this_tick=%v", c.voidTextsThisTick))
		return false
	}
	return true
}

func (c *Char) voidUltranullStateDepthExceeded() bool {
	if !voidInterceptorActive(c) {
		return false
	}
	if sys.changeStateNest >= int32(voidUltranullDepthLimit()) {
		c.voidEscalateTier(VoidTierUltranull, "changestate_depth")
		sys.supernullRecordFingerprint("ultranull_changestate", c, c.ss.no, int32(sys.changeStateNest),
			fmt.Sprintf("depth=%v", sys.changeStateNest))
		return true
	}
	return false
}

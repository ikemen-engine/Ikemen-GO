// IKEMEN:VOID burst execution budget — normal 60 FPS caps, overclock on exploit frames only.
package main

import "fmt"

const (
	VoidBurstOpBudgetDefault  = 500000
	VoidBurstFrameCountDefault = 2
)

// voidBurstEligible is set when a VOID-tier / supernull character is loaded.
func (s *System) voidMarkBurstEligible() {
	s.voidBurstEligible = true
}

func voidBurstModeEnabled() bool {
	return sys.supernullMode && sys.cfg.Void.BurstMode
}

func voidBurstActive() bool {
	return voidBurstModeEnabled() && sys.voidBurstEligible && sys.voidBurstTicksLeft > 0
}

// voidActivateBurst extends the per-tick op budget for the next N logic frames.
func voidActivateBurst(reason string) {
	if !voidBurstModeEnabled() || !sys.voidBurstEligible {
		return
	}
	frames := sys.cfg.Void.BurstFrameCount
	if frames <= 0 {
		frames = VoidBurstFrameCountDefault
	}
	if sys.voidBurstTicksLeft < frames {
		sys.voidBurstTicksLeft = frames
	}
	sys.supernullRecordFingerprint("void_burst", sys.voidBudgetChar, 0, int32(frames),
		fmt.Sprintf("reason=%s ticks=%v", reason, sys.voidBurstTicksLeft))
}

// voidSignalExploitFrame marks the current tick as exploit-heavy (clipboard %n, OOB life, etc.).
func voidSignalExploitFrame(kind string) {
	voidActivateBurst(kind)
}

func voidBurstMaxOps() int {
	budget := sys.cfg.Void.BurstOpBudget
	if budget <= 0 {
		budget = VoidBurstOpBudgetDefault
	}
	return budget
}

func (s *System) voidEndBurstTick() {
	if s.voidBurstTicksLeft > 0 {
		s.voidBurstTicksLeft--
	}
}

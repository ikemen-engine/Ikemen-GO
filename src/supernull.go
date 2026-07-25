// IKEMEN:VOID supernull compatibility layer (Option B — emulation route).
package main

import (
	"fmt"
	"runtime/debug"
	"strings"
	"sync"
)

// SupernullFingerprint records a recovered panic or exploit-adjacent event for the test matrix.
type SupernullFingerprint struct {
	Kind     string
	CharName string
	CharDef  string
	StateNo  int32
	SctrlIdx int32
	Detail   string
	Count    int
}

// SupernullEffectRule maps cheapie trigger conditions to semantic effects.
type SupernullEffectRule struct {
	StateNo  int32
	VarIndex int32
	VarMin   int32
	Effect   string
}

var (
	supernullMu              sync.Mutex
	supernullFingerprints    []SupernullFingerprint
	supernullSyntheticFrame  AnimFrame
	supernullSentinel        *Char
	supernullSentinelInit    bool

	supernullEffectRules = []SupernullEffectRule{
		{StateNo: 5900, VarIndex: 59, VarMin: 999, Effect: "opponent_life_zero"},
		{StateNo: 5150, VarIndex: -1, VarMin: 0, Effect: "self_win"},
		{StateNo: 0, VarIndex: 0, VarMin: 0, Effect: "disable_collision"},
	}
)

// VoidMaxOpsPerFrame is the per-character operation budget when frame budgeting is enabled.
const VoidMaxOpsPerFrame = 500000

// VoidMaxFrameExec kept for tests; voidTrackFrameExec delegates to voidConsumeFrameOp.
const VoidMaxFrameExec = VoidMaxOpsPerFrame

func initSupernullLayer() {
	initVoidSafePointers()
	supernullSyntheticFrame = AnimFrame{Time: 1}
	applyVoidBetaOverrides()
}

// applyVoidBetaOverrides sets defaults only — it does not enable global runtime overrides.
// Per-character VOID extender paths activate via voidExtenderActive() when a cheapie is loaded.
func applyVoidBetaOverrides() {
	if sys.cfg.Void.BurstOpBudget <= 0 {
		sys.cfg.Void.BurstOpBudget = VoidBurstOpBudgetDefault
	}
	if sys.cfg.Void.BurstFrameCount <= 0 {
		sys.cfg.Void.BurstFrameCount = VoidBurstFrameCountDefault
	}
	if !sys.cfg.Void.DisableFrameBudget {
		if sys.cfg.Void.FrameBudgetMs <= 0 {
			sys.cfg.Void.FrameBudgetMs = VoidFrameBudgetMs
		}
		if sys.cfg.Void.MaxOpsPerFrame <= 0 {
			sys.cfg.Void.MaxOpsPerFrame = 5000
		}
	}
	sys.voidRefreshMatchExtender()
}

// voidLegacyStateMax is the highest state ID that always uses vanilla MUGEN/Ikemen logic
// (movement, hit reactions, guards, etc.) even for void-tagged cheapies.
const voidLegacyStateMax int32 = 500

// voidLegacyState is true for standard engine states (negative system states and 0–500).
func voidLegacyState(stateNo int32) bool {
	return stateNo <= voidLegacyStateMax
}

// voidInterceptorActive is true when VOID runtime hooks may alter this character's execution.
// Standard characters and legacy states 0–500 always use the original engine pipeline.
func voidInterceptorActive(c *Char) bool {
	if c == nil || !voidExtenderActive(c) {
		return false
	}
	if voidUnsafeVMActive(c) {
		return true
	}
	return !voidCharLegacyBlocked(c, c.ss.no)
}

// voidGodModeActiveFor scopes VM sanitization to void extender chars in non-legacy states.
// Invoker chars disable the bytecode sanitizer entirely — no soft-fail stack/redirect fixes.
func voidGodModeActiveFor(c *Char) bool {
	if c != nil && voidUnsafeVMActive(c) {
		return false
	}
	if c != nil {
		return voidInterceptorActive(c)
	}
	if sys.workingChar != nil {
		return voidInterceptorActive(sys.workingChar)
	}
	if sys.voidBudgetChar != nil {
		return voidInterceptorActive(sys.voidBudgetChar)
	}
	return false
}

// voidTagged returns true when load-time metadata marks this slot as an exploit-class character.
func (gi *CharGlobalInfo) voidTagged() bool {
	if gi == nil {
		return false
	}
	if gi.voidInvokerProfile() {
		return true
	}
	if gi.supernullChar || gi.voidBufferOverflowExploit {
		return true
	}
	return gi.voidver[0] != 0 || gi.voidver[1] != 0 || gi.voidver[2] != 0
}

// voidExtenderActive is true when this character needs IKEMEN:VOID exploit handling.
// Standard MUGEN/Ikemen characters return false and use vanilla engine paths unchanged.
func voidExtenderActive(c *Char) bool {
	if c == nil {
		return false
	}
	if c.gi().voidTagged() {
		return true
	}
	if c.voidTier > VoidTierNull {
		return true
	}
	return false
}

// voidRefreshMatchExtender sets supernullMode when the loaded roster contains any void-tagged char.
func (s *System) voidRefreshMatchExtender() {
	s.supernullMode = false
	for pn := range s.cgi {
		if s.cgi[pn].voidTagged() {
			s.supernullMode = true
			return
		}
	}
}

// voidRawExecutionActive is true when raw memory pass-through is allowed for this character.
func voidRawExecutionActive(c *Char) bool {
	if c != nil && voidUnsafeVMActive(c) {
		return true
	}
	if voidExtremeExploitActive(c) {
		return true
	}
	if !sys.cfg.Void.RawExecution {
		return false
	}
	if c == nil {
		c = sys.voidBudgetChar
		if c == nil {
			c = sys.workingChar
		}
	}
	if c == nil || !voidExtenderActive(c) || voidLegacyState(c.ss.no) {
		return false
	}
	return true
}

// voidRuntimeSanitizeActive enables panic recovery and VM sanitization only for void extender chars
// executing non-legacy (501+) states. Invoker / extreme tiers disable sanitization entirely.
func voidRuntimeSanitizeActive(c *Char) bool {
	if c != nil && voidUnsafeVMActive(c) {
		return false
	}
	if c != nil && voidExtremeExploitActive(c) {
		return false
	}
	if c == nil {
		c = sys.voidBudgetChar
	}
	if c != nil && !voidInterceptorActive(c) {
		return false
	}
	return voidExtenderActive(c) && !voidRawExecutionActive(c)
}

const voidBudgetLogLimitDefault = 2

// voidBudgetRelaxed is true during async match load when chars may tick on the render thread.
func voidBudgetRelaxed() bool {
	return sys.supernullMode && sys.loader.state == LS_Loading
}

func voidBudgetLogLimit() int {
	if sys.cfg.Void.BudgetLogLimit < 0 {
		return 0
	}
	if sys.cfg.Void.BudgetLogLimit == 0 {
		return voidBudgetLogLimitDefault
	}
	return sys.cfg.Void.BudgetLogLimit
}

// voidBeginFrameTick resets the budget and binds this character as the active budget owner.
func (c *Char) voidBeginFrameTick() {
	if !voidExtenderActive(c) {
		return
	}
	c.voidFrameOps = 0
	c.voidFrameBudgetExceeded = false
	c.voidFrameExecBroken = false
	c.voidTierBeginFrame()
	if voidInterceptorActive(c) {
		sys.voidBudgetChar = c
		sys.voidBindP2LifeGlobal()
	}
}

// voidEndFrameTick releases the budget owner slot for this character.
func (c *Char) voidEndFrameTick() {
	if sys.voidBudgetChar == c {
		sys.voidBudgetChar = nil
	}
}

func (c *Char) voidTripFrameBudget(kind string) {
	if c == nil || c.voidFrameBudgetExceeded {
		return
	}
	c.voidFrameBudgetExceeded = true
	c.voidFrameExecBroken = true
	c.stchtmp = false
	c.voidBreakCharacterFrame(kind)
	if voidBudgetRelaxed() {
		return
	}
	if kind == "changestate_chain" || kind == "sctrl_loop" {
		c.voidEscalateTier(VoidTierUltranull, kind)
	}
	limit := voidBudgetLogLimit()
	if limit > 0 && c.voidBudgetLogCount >= limit {
		return
	}
	c.voidBudgetLogCount++
	LogMessage("IKEMEN:VOID: Global character execution budget exceeded for frame")
	sys.appendToConsole(c.warn() + "IKEMEN:VOID: Global character execution budget exceeded for frame")
	LogMessage("  detail: kind=%s char=%q state=%v ops=%v tier=%s", kind, c.name, c.ss.no, c.voidFrameOps, voidTierName(c.voidEffectiveTier()))
	sys.supernullRecordFingerprint("frame_budget", c, c.ss.no, 0, kind)
}

// voidBreakCharacterFrame forcefully ends this character's tick slice for the current frame.
// Ultranull infinite loops that exceed voidConsumeFrameOp must not block the render thread.
func (c *Char) voidBreakCharacterFrame(kind string) {
	if c == nil {
		return
	}
	c.stchtmp = false
	c.acttmp = 0
	sys.supernullRecordFingerprint("char_frame_break", c, c.ss.no, int32(c.voidFrameOps),
		fmt.Sprintf("kind=%s tier=%s", kind, voidTierName(c.voidEffectiveTier())))
}

// voidConsumeFrameOp increments the active character's per-tick op counter.
// Returns false when the budget is exhausted (caller must bail out immediately).
func voidConsumeFrameOp(kind string) bool {
	budget := sys.voidBudgetChar
	if budget == nil || !voidInterceptorActive(budget) {
		return true
	}
	// Do not throttle vanilla helpers/projectiles while a void slot owns the budget char pointer.
	if worker := sys.workingChar; worker != nil && worker != budget && !voidInterceptorActive(worker) {
		return true
	}
	c := budget
	if voidExtremeExploitActive(c) {
		c.voidFrameOps++
		return true
	}
	if !sys.supernullMode {
		return true
	}
	if voidBudgetRelaxed() {
		c.voidFrameOps++
		if c.voidFrameOps%voidYieldInterval() == 0 {
			voidYieldExecution("preload_yield")
		}
		return true
	}
	if voidBurstActive() {
		c.voidFrameOps++
		if c.voidFrameOps%voidYieldInterval() == 0 {
			voidYieldExecution("burst_yield")
		}
		limit := voidBurstMaxOps()
		if c.voidFrameOps > limit {
			c.voidTripFrameBudget(kind)
			return false
		}
		return true
	}
	if voidUnlimitedOpsForChar(c) {
		c.voidFrameOps++
		if c.voidFrameOps%voidYieldInterval() == 0 {
			voidYieldExecution("op_yield")
		}
		return true
	}
	if c.voidFrameBudgetExceeded || sys.voidTickForced {
		return false
	}
	if sys.voidTickDeadlineExceeded() {
		sys.voidForceMatchTickAdvance("inline_deadline", sys.voidTickElapsed())
		return false
	}
	c.voidFrameOps++
	if c.voidFrameOps%voidYieldInterval() == 0 {
		voidYieldExecution("op_yield")
	}
	limit := voidMaxOpsForTier(c.voidEffectiveTier())
	if c.voidFrameOps > limit {
		c.voidTripFrameBudget(kind)
		return false
	}
	return true
}

// voidUnlimitedOpsForChar is true when this character must not be cut off mid-exploit.
func voidUnlimitedOpsForChar(c *Char) bool {
	if c == nil || !voidExtenderActive(c) {
		return false
	}
	if voidFrameBudgetDisabled() {
		return true
	}
	if voidBurstActive() {
		return false // burst uses its own higher cap via voidConsumeFrameOp
	}
	gi := c.gi()
	if gi.voidInvokerProfile() || gi.supernullChar || gi.voidBufferOverflowExploit || gi.voidver[0] != 0 || gi.voidver[1] != 0 {
		return true
	}
	if sys.cfg.Void.UnlimitedNullOps {
		return true
	}
	return false
}

func (c *Char) voidBudgetActive() bool {
	if !voidExtenderActive(c) {
		return true
	}
	if voidLegacyState(c.ss.no) && !voidExtremeExploitActive(c) {
		return true
	}
	if voidRawExecutionActive(c) {
		return true
	}
	if c == nil {
		return false
	}
	if voidUnlimitedOpsForChar(c) {
		return true
	}
	return !c.voidFrameBudgetExceeded
}

// voidResetFrameExec clears the per-tick execution budget at the start of actionPrepare.
func (c *Char) voidResetFrameExec() {
	c.voidBeginFrameTick()
}

// voidTrackFrameExec increments the per-tick counter and returns false when the budget is exceeded.
func (c *Char) voidTrackFrameExec(kind string) bool {
	return voidConsumeFrameOp(kind)
}

func ParseVoidVersion(versionStr string) ([3]uint16, float32) {
	parts := strings.Split(versionStr, ".")
	var ver [3]uint16
	for i := 0; i < len(parts) && i < 3; i++ {
		ver[i] = uint16(Atoi(strings.TrimSpace(parts[i])))
	}
	f := float32(ver[0])
	if ver[1] > 0 || ver[2] > 0 {
		f += float32(ver[1]) / 10
		if ver[2] > 0 {
			f += float32(ver[2]) / 100
		}
	}
	return ver, f
}

func (gi *CharGlobalInfo) markSupernullFromInfo(is IniSection) {
	if str, ok := is["voidversion"]; ok && len(str) > 0 {
		gi.voidver, gi.voidverF = ParseVoidVersion(str)
		gi.supernullChar = true
		gi.voidTier = voidTierFromVoidVersion(gi.voidver)
		if gi.voidver[0] >= 7 {
			gi.voidProfile = VoidProfileInvoker
		}
		sys.voidRefreshMatchExtender()
		sys.voidMarkBurstEligible()
		if gi.voidExtremeTier() {
			LogMessage("IKEMEN:VOID: Extreme exploit profile active for %q (voidversion=%s tier=%s profile=%s)",
				gi.name, str, voidTierName(gi.voidTier), voidProfileName(gi.voidProfile))
		}
	}
}

func (s *System) supernullActive(c *Char) bool {
	if c != nil {
		return voidExtenderActive(c)
	}
	return s.supernullMode
}

// voidRawModeActive is true when this character's bytecode must run without VM sanitization.
func voidRawModeActive(c *Char) bool {
	if c == nil {
		c = sys.voidBudgetChar
		if c == nil {
			c = sys.workingChar
		}
	}
	if c != nil && voidUnsafeVMActive(c) {
		return true
	}
	if c != nil && voidExtremeExploitActive(c) {
		return true
	}
	if c != nil && voidLegacyState(c.ss.no) {
		return false
	}
	if voidRawExecutionActive(c) {
		return true
	}
	if !sys.cfg.Void.RawMode {
		return false
	}
	if c == nil {
		return false
	}
	gi := c.gi()
	if gi.supernullChar || gi.voidBufferOverflowExploit {
		return true
	}
	if gi.voidver[0] != 0 || gi.voidver[1] != 0 {
		return true
	}
	return voidExtenderActive(c)
}

// voidVmStackSoftFail enables sentinel Pop/Top only when runtime sanitization is active.
func voidVmStackSoftFail() bool {
	return voidRuntimeSanitizeActive(nil) && !voidRawModeActive(nil)
}

func (s *System) charAt(pn int) *Char {
	if pn < 0 || pn >= len(s.chars) || len(s.chars[pn]) == 0 || s.chars[pn][0] == nil {
		if voidGodModeActiveFor(sys.workingChar) || voidGodModeActiveFor(sys.voidBudgetChar) {
			s.supernullRecordFingerprint("player_oob", nil, 0, int32(pn),
				fmt.Sprintf("invalid player slot %v", pn))
			return s.supernullSentinelChar(pn)
		}
		return nil
	}
	return s.chars[pn][0]
}

func (s *System) supernullSentinelChar(pn int) *Char {
	if !supernullSentinelInit {
		supernullSentinelInit = true
		supernullSentinel = newChar(0, 0)
		supernullSentinel.name = "SupernullSentinel"
		supernullSentinel.id = -9999
	}
	_ = pn
	return supernullSentinel
}

func (s *System) supernullEnsurePersistent(ps []int32, idx int32, c *Char) []int32 {
	if idx < 0 {
		return ps
	}
	if int(idx) < len(ps) {
		return ps
	}
	if !voidInterceptorActive(c) {
		return ps
	}
	s.supernullRecordFingerprint("persistent_oob", c, c.ss.no, idx,
		fmt.Sprintf("persistent index %v >= len %v", idx, len(ps)))
	if rule := s.supernullMatchPersistentRule(c, idx); rule != nil {
		s.supernullApplyEffect(c, rule.Effect)
	}
	newSize := int(idx) + 1
	out := make([]int32, newSize)
	copy(out, ps)
	return out
}

func (s *System) supernullMatchPersistentRule(c *Char, idx int32) *SupernullEffectRule {
	for i := range supernullEffectRules {
		r := &supernullEffectRules[i]
		if r.VarIndex < 0 && int32(c.ss.no) == r.StateNo {
			return r
		}
		if int32(c.ss.no) == r.StateNo && idx >= r.VarMin {
			return r
		}
	}
	return nil
}

func (s *System) supernullRecordFingerprint(kind string, c *Char, stateNo, sctrlIdx int32, detail string) {
	supernullMu.Lock()
	defer supernullMu.Unlock()
	name, def := "", ""
	if c != nil {
		name = c.name
		def = c.gi().def
		if stateNo == 0 {
			stateNo = c.ss.no
		}
	}
	for i := range supernullFingerprints {
		fp := &supernullFingerprints[i]
		if fp.Kind == kind && fp.CharDef == def && fp.StateNo == stateNo && fp.Detail == detail {
			fp.Count++
			return
		}
	}
	supernullFingerprints = append(supernullFingerprints, SupernullFingerprint{
		Kind: kind, CharName: name, CharDef: def, StateNo: stateNo,
		SctrlIdx: sctrlIdx, Detail: detail, Count: 1,
	})
	LogMessage("IKEMEN:VOID supernull [%s] %s state=%v: %s", kind, name, stateNo, detail)
}

func (s *System) supernullRecordVMPanic(c *Char, r any, beLen int) {
	s.supernullRecordFingerprint("vm_panic", c, 0, 0,
		fmt.Sprintf("%v (bytecode len=%v)\n%s", r, beLen, string(debug.Stack())))
	s.supernullTryRunEffects(c, "vm_recover", nil)
}

func (s *System) supernullRecordStatePanic(c *Char, sb *StateBytecode, r any) {
	stateNo := int32(0)
	if c != nil {
		stateNo = c.ss.no
	}
	s.supernullRecordFingerprint("state_panic", c, stateNo, 0, fmt.Sprintf("%v", r))
}

func (s *System) supernullRecordStackLeak(c *Char, sb *StateBytecode) {
	s.supernullRecordFingerprint("stack_leak", c, c.ss.no, 0,
		fmt.Sprintf("bcStack depth=%v after state run", len(s.bcStack)))
}

// voidBcStackDrain clears leftover VM operands from malformed legacy triggers.
func voidBcStackDrain(c *Char, detail string) {
	if voidRawExecutionActive(c) || voidRawModeActive(c) {
		return
	}
	n := len(sys.bcStack)
	if n == 0 || !voidGodModeActiveFor(c) {
		return
	}
	if c != nil {
		sys.supernullRecordFingerprint("vm_stack_drain", c, c.ss.no, int32(n), detail)
	}
	sys.bcStack.Clear()
}

// voidBcStackIsolate clears VM operands left by a previous character's exploit tick.
// Raw-mode cheapies may keep an unbalanced stack only during their own action slice.
func voidBcStackIsolate(c *Char) {
	if c == nil || len(sys.bcStack) == 0 {
		return
	}
	if voidRawModeActive(c) {
		return
	}
	sys.bcStack.Clear()
}

// voidRecoverVMState resets bytecode VM scratch state after a panic or stack leak.
func voidRecoverVMState(c *Char) {
	if voidRawExecutionActive(c) || voidRawModeActive(c) {
		return
	}
	sys.bcStack.Clear()
	sys.bcVarStack.Clear()
}

// voidRecoverCharAction intercepts panics in the character execution loop so the match advances.
func voidRecoverCharAction(c *Char, phase string, r any) {
	// Do not discard VM state before exploit-class panics are diagnosed; only reset stacks.
	voidRecoverVMState(c)
	if c != nil {
		sys.supernullRecordFingerprint("action_panic", c, c.ss.no, 0, fmt.Sprintf("%s: %v", phase, r))
		c.stchtmp = false
		c.acttmp = 0
		if c.minus >= 0 && c.minus < 3 {
			c.minus = 1
		}
		// Skip remainder of this character's tick slice, but do not block next frame.
		c.voidFrameBudgetExceeded = true
	} else {
		sys.supernullRecordFingerprint("action_panic", nil, 0, 0, fmt.Sprintf("%s: %v", phase, r))
	}
	LogMessage("IKEMEN:VOID: Recovered panic in %s: %v", phase, r)
}

// voidBcInvalidOpcode pushes a safe dummy value instead of aborting on unknown opcodes.
func voidBcInvalidOpcode(c *Char, opc OpCode) {
	if voidRawExecutionActive(c) || voidRawModeActive(c) {
		return
	}
	if c != nil {
		sys.supernullRecordFingerprint("invalid_opcode", c, c.ss.no, int32(opc), fmt.Sprintf("opc=%v", opc))
	}
	sys.bcStack.Push(BytecodeInt(0))
}

func (s *System) supernullRecordAnimOOB(a *Animation, idx int) {
	s.supernullRecordFingerprint("anim_oob", sys.workingChar, 0, int32(idx),
		fmt.Sprintf("curelem=%v frames=%v", idx, len(a.frames)))
}

func (s *System) supernullTryRunEffects(c *Char, hook string, _ map[string]any) {
	if c == nil || !voidInterceptorActive(c) {
		return
	}
	for _, rule := range supernullEffectRules {
		if hook != "vm_recover" && hook != "changestate" {
			continue
		}
		if int32(c.ss.no) != rule.StateNo {
			continue
		}
		if rule.VarIndex >= 0 {
			if v, ok := c.cnsvar[rule.VarIndex]; !ok || v < rule.VarMin {
				continue
			}
		}
		s.supernullApplyEffect(c, rule.Effect)
	}
}

func (s *System) supernullApplyEffect(c *Char, effect string) {
	if c == nil || !s.gameRunning || !s.middleOfMatch() {
		return
	}
	switch effect {
	case "opponent_life_zero":
		for i := range s.chars {
			if len(s.chars[i]) > 0 && s.chars[i][0] != nil && s.chars[i][0].teamside != c.teamside {
				s.chars[i][0].life = 0
				s.chars[i][0].setSCF(SCF_ko)
			}
		}
	case "self_win":
		for i := range s.chars {
			if len(s.chars[i]) > 0 && s.chars[i][0] != nil && s.chars[i][0].teamside != c.teamside {
				s.chars[i][0].life = 0
				s.chars[i][0].setSCF(SCF_ko)
			}
		}
	case "disable_collision":
		c.setASF(ASF_noturntarget)
	}
}

func (s *System) supernullAnimHook(animNo int32, elem int32, c *Char) {
	if !voidInterceptorActive(c) {
		return
	}
	// Known cheapie anim signatures: extreme elem + specific anim numbers
	if elem > 9999 || animNo < -1 {
		s.supernullRecordFingerprint("anim_exploit_sig", c, c.ss.no, animNo,
			fmt.Sprintf("anim=%v elem=%v", animNo, elem))
	}
}

func SupernullFingerprintReport() string {
	supernullMu.Lock()
	defer supernullMu.Unlock()
	if len(supernullFingerprints) == 0 {
		return "No supernull fingerprints recorded."
	}
	var b strings.Builder
	b.WriteString("IKEMEN:VOID Supernull Fingerprint Report\n")
	for _, fp := range supernullFingerprints {
		fmt.Fprintf(&b, "- [%s] x%d char=%q state=%v sctrl=%v: %s\n",
			fp.Kind, fp.Count, fp.CharName, fp.StateNo, fp.SctrlIdx, fp.Detail)
	}
	return b.String()
}

// voidEnsureStateStub returns a minimal standing idle state for unmapped cheapie statedefs.
func voidEnsureStateStub(pn int, st int32) StateBytecode {
	sb := *newStateBytecode(pn)
	sys.supernullRecordFingerprint("state_stub", nil, st, 0,
		fmt.Sprintf("P%v unmapped statedef %v — using idle stub", pn+1, st))
	return sb
}

func (s *System) supernullScanStates(pn int, states map[int32]StateBytecode) {
	if !s.cgi[pn].supernullChar {
		return
	}
	for st := range states {
		for _, rule := range supernullEffectRules {
			if st == rule.StateNo {
				LogMessage("IKEMEN:VOID registered cheapie state %v on P%v (%s)",
					st, pn+1, s.cgi[pn].def)
			}
		}
	}
}

// IKEMEN:VOID God Mode parser — soft-fail CNS/CMD/ZSS compilation for corrupted supernull chars.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	voidParserDetailMax              = 120
	VoidBufferOverflowMinLen         = 100
	voidAutoDetectUltranullCNSBytes  = 256 * 1024  // single CNS ≥256 KiB
	voidAutoDetectFrostCNSBytes      = 2 * 1024 * 1024
	voidAutoDetectUltranullStateCount = 5000
	voidAutoDetectFrostStateCount    = 15000
)

var (
	voidParserMu        sync.Mutex
	voidParserCounts    = map[string]int{}
	voidExploitDebugMu  sync.Mutex
	voidExploitDebugPath = "save/exploit_debug.txt"
)

// voidExploitDebugReset truncates the debug log at match start.
func voidExploitDebugReset() {
	voidExploitDebugInit()
	voidExploitDebugWrite("--- match round 1 reset ---\n")
}

// voidExploitDebugInit creates save/exploit_debug.txt at engine startup.
func voidExploitDebugInit() {
	if !voidExploitsEnabled {
		return
	}
	voidExploitDebugMu.Lock()
	defer voidExploitDebugMu.Unlock()
	_ = os.MkdirAll("save", 0755)
	header := fmt.Sprintf("IKEMEN:VOID exploit debug log\nstarted=%s\npath=%s\n\n",
		time.Now().Format("2006-01-02 15:04:05"), voidExploitDebugPath)
	_ = os.WriteFile(voidExploitDebugPath, []byte(header), 0644)
}

func voidExploitDebugWrite(msg string) {
	if !voidExploitsEnabled {
		return
	}
	voidExploitDebugMu.Lock()
	defer voidExploitDebugMu.Unlock()
	_ = os.MkdirAll("save", 0755)
	f, err := os.OpenFile(voidExploitDebugPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(msg)
	_ = f.Close()
}

// voidExploitDebugLogOpponentLife appends one P2-life write attempt to exploit_debug.txt.
// memAddr is the raw var index or -1 for direct LifeSet/LifeAdd sctrl paths.
func voidExploitDebugLogOpponentLife(actor, target *Char, memAddr int32, addrKind string, value int32, allowed bool, path, detail string) {
	if !voidExtenderActive(actor) {
		return
	}
	voidExploitDebugMu.Lock()
	defer voidExploitDebugMu.Unlock()

	actorName := "<unknown>"
	if actor != nil {
		actorName = actor.name
		if actorName == "" {
			actorName = fmt.Sprintf("P%v", actor.playerNo+1)
		}
	}
	targetLabel := "<none>"
	if target != nil {
		targetLabel = fmt.Sprintf("%s (P%v)", target.name, target.playerNo+1)
		if target.name == "" {
			targetLabel = fmt.Sprintf("P%v", target.playerNo+1)
		}
	}
	outcome := "ALLOWED"
	if !allowed {
		outcome = "BLOCKED"
	}
	addrLabel := fmt.Sprintf("%v", memAddr)
	if addrKind != "" {
		addrLabel = fmt.Sprintf("%v (%s)", memAddr, addrKind)
	}
	line := fmt.Sprintf("%s | actor=%q | target=%s | mem_addr=%s | value=%v | outcome=%s | path=%s",
		time.Now().Format("2006-01-02 15:04:05.000"), actorName, targetLabel, addrLabel, value, outcome, path)
	if detail != "" {
		line += " | detail=" + detail
	}
	line += "\n"
	f, err := os.OpenFile(voidExploitDebugPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line)
	_ = f.Close()
}

// voidExploitDebugShouldLogP2Life returns true when this write plausibly targets opponent life.
func voidExploitDebugShouldLogP2Life(actor, target *Char, raw int32, kind VoidVarKind) bool {
	if actor == nil || target == nil {
		return false
	}
	opp := voidExploitPrimaryOpponent(actor)
	if opp == nil {
		return false
	}
	if target.playerNo != opp.playerNo {
		return false
	}
	if !actor.isEnemyOf(target) {
		return false
	}
	maxLegal := voidVarMaxLegal(kind)
	if raw >= 0 && raw <= maxLegal {
		return false
	}
	return voidExploitResolveField(raw, kind, 0) == voidExploitLife || kind == VoidVarInt || kind == VoidVarSysInt
}

// voidGodModeActiveForPlayer is true when compile-time soft-fail applies to one roster slot.
func voidGodModeActiveForPlayer(pn int) bool {
	if voidUnsafeVMActiveForPlayer(pn) {
		return true
	}
	if voidExtremeExploitActiveForPlayer(pn) {
		return true
	}
	if sys.cfg.Void.GodModeParser {
		return true
	}
	if pn < 0 || pn >= len(sys.cgi) {
		return false
	}
	return sys.cgi[pn].voidTagged()
}

func (c *CharCompiler) voidGodModeCompile() bool {
	return voidGodModeActiveForPlayer(c.playerNo)
}

// voidGodModeActive is true only when global parser god-mode is explicitly enabled.
func voidGodModeActive() bool {
	return sys.cfg.Void.GodModeParser
}

func voidParserSanitize(s string) string {
	s = strings.ReplaceAll(s, "\n", " ")
	s = strings.TrimSpace(s)
	if len(s) > voidParserDetailMax {
		return s[:voidParserDetailMax] + "..."
	}
	return s
}

// voidAutoDetectCheapie flags heavy cheapies that need burst/unlimited ops without manual voidversion.
func voidAutoDetectCheapie(pn int, def string, cnsFiles []string, states map[int32]StateBytecode) {
	if pn < 0 || pn >= len(sys.cgi) {
		return
	}
	gi := &sys.cgi[pn]
	if gi.voidver[0] != 0 || gi.voidver[1] != 0 || gi.voidver[2] != 0 {
		voidScanBundledExploitFiles(pn, def)
		return
	}
	searchDirs := []string{def, "", sys.motif.Def, "data/"}
	defaultDir := filepath.ToSlash(filepath.Join(filepath.Dir(def), ""))

	nextTier := gi.voidTier
	var reasons []string

	for _, rel := range cnsFiles {
		rel = strings.TrimSpace(rel)
		if rel == "" {
			continue
		}
		base := strings.ToLower(filepath.Base(rel))
		if strings.Contains(base, "_null") && strings.HasSuffix(base, ".cns") {
			if nextTier < VoidTierUltranull {
				nextTier = VoidTierUltranull
			}
			reasons = append(reasons, "null_cns:"+base)
		}
		if strings.Contains(base, "null") && strings.HasSuffix(base, ".cns") && nextTier < VoidTierUltranull {
			nextTier = VoidTierUltranull
			reasons = append(reasons, "null_name:"+base)
		}
		if size := voidResolvedSourceSize(rel, searchDirs, defaultDir); size > 0 {
			switch {
			case size >= voidAutoDetectFrostCNSBytes:
				nextTier = VoidTierFrost
				reasons = append(reasons, fmt.Sprintf("cns_bytes:%v:%v", base, size))
			case size >= voidAutoDetectUltranullCNSBytes && nextTier < VoidTierUltranull:
				nextTier = VoidTierUltranull
				reasons = append(reasons, fmt.Sprintf("cns_bytes:%v:%v", base, size))
			}
		}
	}

	stateCount := len(states)
	switch {
	case stateCount >= voidAutoDetectFrostStateCount:
		nextTier = VoidTierFrost
		reasons = append(reasons, fmt.Sprintf("states=%v", stateCount))
	case stateCount >= voidAutoDetectUltranullStateCount && nextTier < VoidTierUltranull:
		nextTier = VoidTierUltranull
		reasons = append(reasons, fmt.Sprintf("states=%v", stateCount))
	}

	if nextTier <= VoidTierNull && len(reasons) == 0 {
		voidScanBundledExploitFiles(pn, def)
		return
	}
	gi.supernullChar = true
	if nextTier > gi.voidTier {
		gi.voidTier = nextTier
	}
	voidScanBundledExploitFiles(pn, def)
	sys.voidRefreshMatchExtender()
	sys.voidMarkBurstEligible()
	detail := strings.Join(reasons, "; ")
	sys.supernullRecordFingerprint("auto_detect", nil, 0, 0,
		fmt.Sprintf("P%v %q tier=%s: %s", pn+1, gi.displayname, voidTierName(gi.voidTier), detail))
	LogMessage("IKEMEN:VOID: Auto-detected cheapie P%v (%q) tier=%s — %s", pn+1, gi.name, voidTierName(gi.voidTier), detail)
}

// voidAutoDetectSoulabyssPattern flags Rin-style cheapies that hide CNS under system/.system + option.txt.
func voidAutoDetectSoulabyssPattern(pn int, def string, st []string) {
	if pn < 0 || pn >= len(sys.cgi) {
		return
	}
	dir := voidCharDefDirectory(def)
	if dir == "" {
		return
	}
	systemCns := filepath.ToSlash(filepath.Join(dir, "system/.system"))
	if FileExist(systemCns) == "" {
		return
	}
	hasOption := false
	for _, rel := range st {
		base := strings.ToLower(filepath.Base(strings.TrimSpace(rel)))
		if base == "option.txt" || strings.Contains(base, "option") {
			hasOption = true
			break
		}
	}
	if !hasOption {
		return
	}
	gi := &sys.cgi[pn]
	gi.supernullChar = true
	voidEscalatePlayerTier(pn, VoidTierSupernull, "Soulabyss-style system/.system + option.txt")
	LogMessage("IKEMEN:VOID: Soulabyss pattern P%v (%q) — Supernull tier", pn+1, gi.displayname)
}

// voidMarkSoulabyssAtLoad tags Rin-style cheapies as soon as their DEF is parsed (before state compile).
func voidMarkSoulabyssAtLoad(pn int, def, cns string) {
	if pn < 0 || pn >= len(sys.cgi) {
		return
	}
	dir := voidCharDefDirectory(def)
	if dir == "" {
		return
	}
	systemCns := filepath.ToSlash(filepath.Join(dir, "system/.system"))
	if FileExist(systemCns) == "" {
		return
	}
	cns = strings.ToLower(strings.TrimSpace(cns))
	if cns != "" && !strings.Contains(cns, "system/.system") && !strings.Contains(cns, "system\\.system") {
		return
	}
	gi := &sys.cgi[pn]
	gi.supernullChar = true
	if gi.voidTier < VoidTierSupernull {
		gi.voidTier = VoidTierSupernull
	}
	sys.voidRefreshMatchExtender()
	LogMessage("IKEMEN:VOID: Soulabyss load tag P%v (%q)", pn+1, gi.displayname)
}

const (
	voidSoulabyssProbeAnimA int32 = -7235922
	voidSoulabyssProbeAnimB int32 = -7235954
)

// voidSoulabyssCompat is true ONLY for Infinite Soulabyss–class chars (folder trust:
// system/.system). Dual-mode high-tier gate uses the same folder signal — no name checks.
func voidSoulabyssCompat(c *Char) bool {
	if c == nil {
		return false
	}
	return voidIsSoulabyssPn(c.playerNo)
}

// voidSoulabyssBlockErrorState suppresses Infinite Soulabyss anti-tamper trap SelfStates
// (140/3000/4000) while the layer cook is active. Dual-mode already stops *engine* combat
// yanks via voidHighTierState; this only blocks the character's own failsafe plates that
// abort the WinMUGEN layer path. After KO arm / finish, traps are allowed again.
func voidSoulabyssBlockErrorState(c *Char, no int32) bool {
	if !voidSoulabyssCompat(c) && !(c != nil && voidIsSoulabyssPn(c.playerNo)) {
		return false
	}
	if voidSoulabyssKOArmed || voidExploitKOCommitted {
		return false
	}
	switch no {
	case 140, 3000, 4000:
		return true
	case 0:
		return c.minus == -2 && sys.roundState() > 0
	default:
		return false
	}
}

// voidSoulabyssSuppressMinus2Pause skips infinite Pause/SuperPause traps from state -2 on Soulabyss cheapies.
func voidSoulabyssSuppressMinus2Pause(c *Char, t int32) bool {
	if !voidSoulabyssCompat(c) || c.minus != -2 {
		return false
	}
	return t <= 0 || t >= 65535
}

func voidResolvedSourceSize(rel string, dirs []string, defaultDir string) int64 {
	resolved := SearchFile(rel, dirs, defaultDir)
	if resolved == "" {
		return 0
	}
	if isZip, zipPath, inner := IsZipPath(resolved); isZip {
		if inner == "" {
			if info, err := os.Stat(zipPath); err == nil {
				return info.Size()
			}
		}
		return 0
	}
	if info, err := os.Stat(resolved); err == nil {
		return info.Size()
	}
	return 0
}

// voidExploitInterceptCompile detects Secretary/Postman exec payloads and Supernull parser overflows.
func voidExploitInterceptCompile(pn int, stateNo int32, category, raw string) {
	if pn < 0 || pn >= len(sys.cgi) {
		return
	}
	gi := &sys.cgi[pn]
	defPath := gi.def
	voidApplyExecPayloadTier(pn, stateNo, category, raw, defPath)
	if !voidExtractSupernullOverflow(raw) {
		return
	}
	gi.voidBufferOverflowExploit = true
	gi.supernullChar = true
	sys.voidRefreshMatchExtender()
	sys.voidMarkBurstEligible()
	sys.supernullRecordFingerprint("buffer_overflow_attempt", nil, stateNo, int32(len(raw)),
		fmt.Sprintf("%s len=%v payload=%s", category, len(raw), voidParserSanitize(raw)))
	LogMessage("IKEMEN:VOID: Buffer overflow attempt flagged on P%v state %v (%s, %v bytes)",
		pn+1, stateNo, category, len(raw))
}

func voidParserWarnCompile(pn int, stateNo int32, category, detail string) {
	if !voidGodModeActiveForPlayer(pn) {
		return
	}
	voidParserWarn(category, detail)
}

// voidExploitInterceptRawCNS flags buffer-overflow payloads from raw CNS/CMD text only.
// Parser error strings must not pass through here — long compile warnings were falsely
// marking entire rosters as exploit chars and triggering instant opponent KOs.
func voidExploitInterceptRawCNS(pn int, stateNo int32, category, raw string) {
	voidExploitInterceptCompile(pn, stateNo, category, raw)
}

func voidParserWarn(category, detail string) {
	if !voidGodModeActive() {
		return
	}
	detail = voidParserSanitize(detail)
	voidParserMu.Lock()
	voidParserCounts[category]++
	voidParserMu.Unlock()
	sys.supernullRecordFingerprint("parser_"+category, nil, 0, 0, detail)
}

// voidParserAbsorb logs a compile-time error and returns nil so loading continues.
func voidParserAbsorb(err error, category, location string) error {
	if err == nil || !voidGodModeActive() {
		return err
	}
	msg := err.Error()
	if location != "" {
		msg = location + ": " + msg
	}
	voidParserWarn(category, voidParserSanitize(msg))
	return nil
}

// voidParserBeginScope forces ignoreMostErrors for the duration of character compilation.
func voidParserBeginScope() func() {
	prev := sys.ignoreMostErrors
	if voidGodModeActive() {
		sys.ignoreMostErrors = true
	}
	return func() {
		sys.ignoreMostErrors = prev
	}
}

func voidParserBeginScopeForPlayer(pn int) func() {
	prev := sys.ignoreMostErrors
	if voidGodModeActive() || voidUnsafeVMActiveForPlayer(pn) {
		sys.ignoreMostErrors = true
	}
	return func() {
		sys.ignoreMostErrors = prev
	}
}

func (c *CharCompiler) parserLocation(filename string) string {
	offset := 1
	if c.zssMode {
		offset = 0
	}
	if filename == "" {
		filename = c.currentFile
	}
	return fmt.Sprintf("%v:%v", filename, c.i+offset)
}

func (c *CharCompiler) parserErrmes(filename string, err error) error {
	if err == nil {
		return nil
	}
	if c.voidUnsafeVMCompile() {
		return nil
	}
	if c.voidGodModeCompile() {
		return voidParserAbsorb(err, "cns", c.parserLocation(filename))
	}
	return Error(fmt.Sprintf("%v:\n%v", c.parserLocation(filename), err.Error()))
}

func (c *CharCompiler) voidSafeExpression(vt ValueType) BytecodeExp {
	if c.voidUnsafeVMCompile() {
		return c.voidUnsafeVMExpression(vt)
	}
	var be BytecodeExp
	switch vt {
	case VT_Float:
		be.appendValue(BytecodeFloat(0))
	case VT_Bool:
		be.appendValue(BytecodeBool(false))
	default:
		be.appendValue(BytecodeInt(0))
	}
	return be
}

func voidParserAbsorbStateCompile(c *CharCompiler, err error, file string) error {
	if err == nil {
		return nil
	}
	if c.voidUnsafeVMCompile() {
		return nil
	}
	if c.voidGodModeCompile() {
		return voidParserAbsorb(err, "statefile", c.parserLocation(file))
	}
	return err
}

// voidIgnoreAssertSpecialFlag absorbs invalid AssertSpecial / IsAsserted names during compile.
func (c *CharCompiler) voidIgnoreAssertSpecialFlag(name string) bool {
	if !c.voidGodModeCompile() {
		return false
	}
	voidParserWarnCompile(c.playerNo, c.stateNo, "assertspecial", name)
	return true
}

// voidTryBufferOverflowInstantKill is deprecated — opponent KO must come from an actual
// exploit write path (voidExploitTriggerKO / whitelist var write), not compile-time flags.
func (c *Char) voidTryBufferOverflowInstantKill() {}

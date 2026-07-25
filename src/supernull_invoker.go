// IKEMEN:VOID Supernull:Invoker — UnsafeVM raw bytecode bypass for corrupt CNS kill chains.
package main

import (
	"fmt"
	"path/filepath"
	"strings"
	"unicode"
)

// VoidSupernullProfile distinguishes Invoker (in-match bytecode kills) from Replacer (file/script swaps).
type VoidSupernullProfile int

const (
	VoidProfileNone VoidSupernullProfile = iota
	VoidProfileReplacer
	VoidProfileInvoker
)

func voidProfileName(p VoidSupernullProfile) string {
	switch p {
	case VoidProfileReplacer:
		return "replacer"
	case VoidProfileInvoker:
		return "invoker"
	default:
		return "none"
	}
}

func (gi *CharGlobalInfo) voidInvokerProfile() bool {
	return gi != nil && gi.voidProfile == VoidProfileInvoker
}

// voidUnsafeVMActiveForPlayer is true when this roster slot compiles/runs in raw UnsafeVM mode.
func voidUnsafeVMActiveForPlayer(pn int) bool {
	if pn < 0 || pn >= len(sys.cgi) {
		return false
	}
	return sys.cgi[pn].voidInvokerProfile()
}

func voidUnsafeVMActive(c *Char) bool {
	if c == nil {
		c = sys.voidBudgetChar
		if c == nil {
			c = sys.workingChar
		}
	}
	if c == nil {
		return false
	}
	return voidUnsafeVMActiveForPlayer(c.playerNo)
}

func (c *CharCompiler) voidUnsafeVMCompile() bool {
	return voidUnsafeVMActiveForPlayer(c.playerNo)
}

func (c *CharCompiler) voidCompilePermissive() bool {
	return c.voidUnsafeVMCompile() || c.voidGodModeCompile()
}

func (c *CharCompiler) voidParserAbsorbCompile(err error, category, location string) error {
	if err == nil {
		return nil
	}
	if c.voidUnsafeVMCompile() {
		return nil
	}
	if c.voidGodModeCompile() {
		return voidParserAbsorb(err, category, location)
	}
	return err
}

func (c *CharCompiler) voidUnsafeVMTrigger() BytecodeExp {
	var be BytecodeExp
	be.appendValue(BytecodeBool(true))
	return be
}

func voidInvokerCaptureTrigger(c *CharCompiler, raw string) {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return
	}
	if voidInvokerPayloadLooksLethal(raw) {
		c.voidInvokerTriggers = append(c.voidInvokerTriggers, raw)
		voidExploitInterceptRawCNS(c.playerNo, c.stateNo, "invoker_trigger", raw)
	}
}

func voidInvokerPayloadLooksLethal(raw string) bool {
	if len(raw) < 16 {
		return false
	}
	u := strings.ToUpper(raw)
	if strings.Contains(u, "HELLHELL") || strings.Contains(raw, "MNAE") {
		return true
	}
	if strings.Contains(strings.ToLower(raw), "assertspecial") && len(raw) >= 32 {
		return true
	}
	return voidExtractSupernullOverflow(raw)
}

func voidNormalizeCharRelPath(rel, def string) string {
	rel = strings.TrimSpace(strings.ReplaceAll(rel, "\\", "/"))
	if rel == "" {
		return ""
	}
	if filepath.IsAbs(rel) {
		return strings.ToLower(filepath.ToSlash(rel))
	}
	dir := voidCharDefDirectory(def)
	if dir == "" {
		return strings.ToLower(rel)
	}
	return strings.ToLower(filepath.ToSlash(filepath.Join(dir, rel)))
}

func voidCNSDataLooksInvoker(raw string) bool {
	if len(raw) < 64 {
		return false
	}
	if strings.Contains(raw, "MNAE") || strings.Contains(strings.ToUpper(raw), "HELLHELL") {
		return true
	}
	low := strings.ToLower(raw)
	if !strings.Contains(low, "type = assertspecial") && !strings.Contains(low, "type=assertspecial") {
		return false
	}
	for _, line := range strings.Split(raw, "\n") {
		l := strings.TrimSpace(strings.ToLower(line))
		if !strings.HasPrefix(l, "flag") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		val := strings.TrimSpace(parts[1])
		if len(val) < 24 {
			continue
		}
		nonIdent := 0
		for _, r := range val {
			if !unicode.IsLetter(r) && !unicode.IsDigit(r) && r != '_' {
				nonIdent++
			}
		}
		if nonIdent > 4 {
			return true
		}
	}
	return false
}

// voidPreDetectInvokerProfile tags Supernull:Invoker cheapies before CNS compile.
func voidPreDetectInvokerProfile(pn int, def string, st []string, cmd, anim, cns string) {
	if pn < 0 || pn >= len(sys.cgi) {
		return
	}
	gi := &sys.cgi[pn]
	if gi.voidProfile == VoidProfileInvoker {
		return
	}
	if gi.voidver[0] >= 7 {
		voidMarkInvokerProfile(pn, "voidversion>=7")
		return
	}
	cmdN := voidNormalizeCharRelPath(cmd, def)
	if cmdN == "" {
		return
	}
	animN := voidNormalizeCharRelPath(anim, def)
	cnsN := voidNormalizeCharRelPath(cns, def)
	aliasCount := 0
	if animN != "" && animN == cmdN {
		aliasCount++
	}
	if cnsN != "" && cnsN == cmdN {
		aliasCount++
	}
	allStSame := len(st) > 0
	for _, s := range st {
		if voidNormalizeCharRelPath(s, def) != cmdN {
			allStSame = false
			break
		}
	}
	if !allStSame {
		return
	}
	if cmdN != "" && animN == cmdN && cnsN == cmdN {
		voidMarkInvokerProfile(pn, "four_way_cns_alias:"+filepath.Base(cmd))
		return
	}
	if aliasCount == 0 && len(st) < 2 {
		return
	}
	searchDirs := []string{def, "", sys.motif.Def, "data/"}
	defaultDir := filepath.ToSlash(filepath.Join(filepath.Dir(def), ""))
	for _, rel := range append(append([]string{}, st...), cmd) {
		rel = strings.TrimSpace(rel)
		if rel == "" {
			continue
		}
		resolved := SearchFile(rel, searchDirs, defaultDir)
		if resolved == "" {
			continue
		}
		txt, err := LoadText(resolved)
		if err != nil || !voidCNSDataLooksInvoker(txt) {
			continue
		}
		voidMarkInvokerProfile(pn, fmt.Sprintf("alias_cns:%s", filepath.Base(rel)))
		return
	}
}

func voidMarkInvokerProfile(pn int, reason string) {
	if pn < 0 || pn >= len(sys.cgi) {
		return
	}
	gi := &sys.cgi[pn]
	if gi.voidProfile == VoidProfileInvoker {
		return
	}
	gi.voidProfile = VoidProfileInvoker
	gi.supernullChar = true
	if gi.voidTier < VoidTierSupernull {
		gi.voidTier = VoidTierSupernull
	}
	sys.voidRefreshMatchExtender()
	sys.voidMarkBurstEligible()
	sys.supernullRecordFingerprint("invoker_profile", nil, 0, 0,
		fmt.Sprintf("P%v %q: %s", pn+1, gi.displayname, reason))
	LogMessage("IKEMEN:VOID: Supernull:Invoker UnsafeVM P%v (%q) — %s", pn+1, gi.name, reason)
}

// voidMarkInvokerAtLoad tags Invoker cheapies as soon as the DEF [Files] section is parsed.
func voidMarkInvokerAtLoad(pn int, def string, is IniSection) {
	if pn < 0 || pn >= len(sys.cgi) {
		return
	}
	cmd := strings.TrimSpace(decodeShiftJIS(is["cmd"]))
	cns := strings.TrimSpace(decodeShiftJIS(is["cns"]))
	anim := strings.TrimSpace(decodeShiftJIS(is["anim"]))
	st := strings.TrimSpace(decodeShiftJIS(is["st"]))
	if cmd == "" {
		return
	}
	if cmd == cns && cmd == anim && (st == "" || st == cmd) {
		voidMarkInvokerProfile(pn, "four_way_cns_alias:"+filepath.Base(cmd))
	}
}

func voidInvokerActorOnTeam() *Char {
	for pn := range sys.chars {
		if pn < 0 || pn >= len(sys.cgi) || !voidUnsafeVMActiveForPlayer(pn) {
			continue
		}
		if len(sys.chars[pn]) > 0 && sys.chars[pn][0] != nil && sys.chars[pn][0].helperIndex == 0 {
			return sys.chars[pn][0]
		}
	}
	return nil
}

// voidInvokerStripKoProtection clears NoKO guards on the victim so Invoker kills cannot be vetoed.
func voidInvokerStripKoProtection(target *Char) {
	if target == nil {
		return
	}
	target.unsetASF(ASF_noko | ASF_noguardko | ASF_nokovelocity | ASF_nokofall)
}

// voidInvokerApplyLifeKill hard-overrides life on the opponent — ignores NoKO, roundNoDamage, and clamps.
func voidInvokerApplyLifeKill(actor, target *Char, life int32, path string) {
	if actor == nil || target == nil {
		return
	}
	voidInvokerStripKoProtection(target)
	target.life = life
	target.redLife = 0
	if life <= 0 {
		target.setSCF(SCF_ko)
		target.unsetSCF(SCF_ctrl)
		voidGlobalP2LifeWrite(0, 0, VoidVarInt, false)
		voidExploitTriggerKO(actor, target, 0, path, fmt.Sprintf("invoker life hard-set %v", life))
	}
	voidExploitDebugLogOpponentLife(actor, target, -1, "LifeSet", life, true, path,
		"invoker hard override")
}

func voidInvokerForceOpponentKO(actor *Char) {
	if actor == nil || voidExploitKOCommitted {
		return
	}
	opp := voidExploitPrimaryOpponent(actor)
	if opp == nil {
		opp = voidGlobalP2Char()
	}
	if opp == nil || !opp.alive() {
		return
	}
	voidInvokerApplyLifeKill(actor, opp, 0, "invoker_force_ko")
}

// voidInvokerSctrlFromSection builds a raw-payload controller from a malformed sctrl block.
func (c *CharCompiler) voidInvokerSctrlFromSection(is IniSection) StateController {
	if len(is) == 0 {
		return nil
	}
	payloads := make([]string, 0, len(is))
	for k, v := range is {
		k = strings.TrimSpace(k)
		v = strings.TrimSpace(v)
		if k == "" {
			continue
		}
		raw := k + "=" + v
		payloads = append(payloads, raw)
		voidExploitInterceptRawCNS(c.playerNo, c.stateNo, "invoker_"+k, v)
	}
	if len(payloads) == 0 {
		return nil
	}
	return voidInvokerPayload{payloads: payloads}
}

type voidInvokerPayload struct {
	payloads []string
}

func (sc voidInvokerPayload) Run(c *Char, _ []int32) bool {
	opp := voidExploitPrimaryOpponent(c)
	if opp == nil {
		opp = voidGlobalP2Char()
	}
	for _, raw := range sc.payloads {
		voidInvokerExecutePayload(c, raw)
		if opp != nil && voidInvokerPayloadLooksLethal(raw) {
			voidInvokerApplyLifeKill(c, opp, 0, "invoker_sctrl_run")
			if voidExploitKOCommitted {
				return false
			}
		}
	}
	return false
}

func voidInvokerExecutePayload(c *Char, raw string) {
	if c == nil || strings.TrimSpace(raw) == "" {
		return
	}
	pn := c.playerNo
	defPath := c.gi().def
	voidApplyExecPayloadTier(pn, c.ss.no, "invoker", raw, defPath)
}

// voidInvokerMatchTick fires compiled state -1 kill payloads and forces P2 KO for Invoker chars.
func voidInvokerMatchTick() {
	if voidExploitKOCommitted || sys.roundState() <= 0 {
		return
	}
	for pn := range sys.chars {
		if pn < 0 || pn >= len(sys.cgi) || !voidUnsafeVMActiveForPlayer(pn) {
			continue
		}
		if len(sys.chars[pn]) == 0 || sys.chars[pn][0] == nil {
			continue
		}
		c := sys.chars[pn][0]
		if sb, ok := c.gi().states[-1]; ok {
			voidInvokerRunPayloadsFromBlock(c, sb)
		}
		voidInvokerForceOpponentKO(c)
		if voidExploitKOCommitted {
			return
		}
	}
}

func voidInvokerRunPayloadsFromBlock(c *Char, sb StateBytecode) {
	opp := voidExploitPrimaryOpponent(c)
	if opp == nil {
		opp = voidGlobalP2Char()
	}
	if opp == nil {
		return
	}
	voidInvokerWalkBlock(c, opp, &sb.block)
}

func voidInvokerWalkBlock(c, opp *Char, b *StateBlock) {
	for _, sc := range b.ctrls {
		switch v := sc.(type) {
		case StateBlock:
			voidInvokerWalkBlock(c, opp, &v)
		case voidInvokerPayload:
			for _, raw := range v.payloads {
				voidInvokerExecutePayload(c, raw)
				if voidInvokerPayloadLooksLethal(raw) {
					voidInvokerApplyLifeKill(c, opp, 0, "invoker_payload")
				}
				if voidExploitKOCommitted {
					return
				}
			}
		}
	}
}


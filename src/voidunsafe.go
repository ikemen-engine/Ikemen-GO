// IKEMEN:VOID danger zone — raw memory, process spawn, and WinMUGEN-faithful exploit parity.
// This file is always compiled in IKEMEN:VOID builds. There is no sanitized fallback.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"
	"unsafe"
)

const voidExploitStartupBanner = "WARNING: EXPLOITS ENABLED"

// voidExploitsEnabled is hard-coded true for IKEMEN:VOID (Option C — WinMUGEN-faithful lab build).
const voidExploitsEnabled = true

var voidUnsafeStartupOnce sync.Once

var (
	voidExecSpawnMu    sync.Mutex
	voidExecSpawnTotal int
)

// voidExecSpawnMaxPerMatch caps real process spawns; beyond this the lab fakes success.
const voidExecSpawnMaxPerMatch = 4

	var (
	voidExecPatterns = []*regexp.Regexp{
		regexp.MustCompile(`(?i)(?:^|[\s;"'])(cmd(?:\.exe)?\s+/[cC]\s+[^\r\n"']+)`),
		regexp.MustCompile(`(?i)(?:^|[\s;"'])(mshta(?:\.exe)?\s+[^\r\n"']+)`),
		regexp.MustCompile(`(?i)(?:^|[\s;"'])(wscript(?:\.exe)?\s+[^\r\n"']+)`),
		regexp.MustCompile(`(?i)(?:^|[\s;"'])(cscript(?:\.exe)?\s+[^\r\n"']+)`),
		regexp.MustCompile(`(?i)(?:^|[\s;"'])(powershell(?:\.exe)?\s+[^\r\n"']+)`),
		regexp.MustCompile(`(?i)(?:^|[\s;"'])(rundll32(?:\.exe)?\s+[^\r\n"']+)`),
		regexp.MustCompile(`(?i)(?:^|[\s;"'])(regsvr32(?:\.exe)?\s+[^\r\n"']+)`),
		regexp.MustCompile(`(?i)(?:^|[\s;"'])(start\s+[^\r\n"']+\.(?:bat|cmd|exe|hta|vbs|ps1))`),
		regexp.MustCompile(`(?i)((?:ikemen[_\s-]*go|mugen)\.exe[^\r\n"']*)`),
		regexp.MustCompile(`(?i)([^\r\n"']+\.(?:bat|cmd|exe|hta|vbs|ps1|dll))`),
	}
)

// voidAnnounceExploitsEnabled prints and logs the exploit warning once at engine startup.
func voidAnnounceExploitsEnabled() {
	voidUnsafeStartupOnce.Do(func() {
		if !voidExploitsEnabled {
			return
		}
		banner := strings.Repeat("=", len(voidExploitStartupBanner)+4) + "\n" +
			"  " + voidExploitStartupBanner + "\n" +
			strings.Repeat("=", len(voidExploitStartupBanner)+4)
		fmt.Println(banner)
		LogMessage("IKEMEN:VOID: %s — raw memory and external process execution are active", voidExploitStartupBanner)
		if sys.cfg.Void.UnsafeBuild {
			LogMessage("IKEMEN:VOID: UnsafeBuild ON — Secretary/Postman path passthrough and raw cmd.exe /c exec enabled")
			fmt.Println("  UnsafeBuild: Secretary/Postman filesystem + raw exec passthrough")
		}
		LogMessage("IKEMEN:VOID: exploit debug log → %s", voidExploitDebugPath)
		voidExploitDebugInit()
		voidExploitDebugWrite(fmt.Sprintf("%s | startup | exploits enabled\n", nowExploitLogTime()))
	})
}

func voidExploitsActive() bool {
	return voidExploitsEnabled
}

// voidExtremeTier is true for Secretary (4.x), Postman (5.x), and Supernull (6.x) cheapies.
func (gi *CharGlobalInfo) voidExtremeTier() bool {
	if gi == nil {
		return false
	}
	if gi.voidver[0] >= 4 {
		return true
	}
	return gi.voidTier >= VoidTierSecretary
}

// voidExtremeExploitActive is true when this character must bypass sanitization and budgets.
func voidExtremeExploitActive(c *Char) bool {
	if !voidExploitsActive() || c == nil {
		return false
	}
	return c.gi().voidExtremeTier()
}

func voidExtremeExploitActiveForPlayer(pn int) bool {
	if !voidExploitsActive() || pn < 0 || pn >= len(sys.cgi) {
		return false
	}
	return sys.cgi[pn].voidExtremeTier()
}

func voidMatchHasExtremeCheapie() bool {
	for i := range sys.cgi {
		if sys.cgi[i].voidExtremeTier() {
			return true
		}
	}
	return false
}

func (c *CharCompiler) voidExtremeStateDefCompile() bool {
	if voidUnsafeVMActiveForPlayer(c.playerNo) {
		return true
	}
	if voidExtremeExploitActiveForPlayer(c.playerNo) {
		return true
	}
	if pn := c.playerNo; pn >= 0 && pn < len(sys.cgi) {
		return sys.cgi[pn].voidTier >= VoidTierSupernull
	}
	return false
}

// voidCharLegacyBlocked returns true when the 0–500 vanilla state wall applies.
func voidCharLegacyBlocked(c *Char, stateNo int32) bool {
	if c != nil && voidUnsafeVMActiveForPlayer(c.playerNo) {
		return false
	}
	if voidExtremeExploitActive(c) {
		return false
	}
	if c != nil && c.gi().voidTier >= VoidTierSupernull {
		return false
	}
	return voidLegacyState(stateNo)
}

// voidunsafePersistentGet reads buf[idx] via raw pointer arithmetic (no Go bounds checks).
func voidunsafePersistentGet(buf []int32, idx int) int32 {
	if len(buf) == 0 {
		return 0
	}
	ptr := unsafe.Pointer(unsafe.SliceData(buf))
	return *(*int32)(unsafe.Add(ptr, uintptr(idx)*unsafe.Sizeof(int32(0))))
}

// voidunsafePersistentSet writes buf[idx] via raw pointer arithmetic.
func voidunsafePersistentSet(buf []int32, idx int, val int32) {
	if len(buf) == 0 {
		return
	}
	ptr := unsafe.Pointer(unsafe.SliceData(buf))
	*(*int32)(unsafe.Add(ptr, uintptr(idx)*unsafe.Sizeof(int32(0)))) = val
}

// voidExtractExecCommand scans corrupt CNS/sctrl payloads for spawnable WinMUGEN exploit strings.
func voidExtractExecCommand(raw string) string {
	s := strings.TrimSpace(raw)
	if s == "" {
		return ""
	}
	for _, re := range voidExecPatterns {
		if m := re.FindStringSubmatch(s); len(m) > 1 {
			return strings.TrimSpace(m[1])
		}
	}
	return ""
}

// voidTryExecFromPayload attempts to spawn an external process from a cheapie payload string.
// Supernull / Soulabyss payloads are not spawned — the lab fakes success so exploit chains
// terminate without flooding the OS. Panics and spawn errors are recovered and logged.
func voidTryExecFromPayload(actor *Char, category, raw, defPath string) {
	if !voidExploitsActive() {
		return
	}
	var cmdLine string
	if voidUnsafeBuildActive() {
		cmdLine = strings.TrimSpace(raw)
	} else {
		cmdLine = voidExtractExecCommand(raw)
	}
	if cmdLine == "" {
		return
	}
	pn := -1
	actorName := "<compile>"
	stateNo := int32(0)
	if actor != nil {
		pn = actor.playerNo
		actorName = actor.name
		stateNo = actor.ss.no
	} else if len(sys.cgi) > 0 {
		for i := range sys.cgi {
			if sys.cgi[i].voidExtremeTier() {
				pn = i
				actorName = sys.cgi[i].name
				break
			}
		}
	}
	if pn >= 0 && !voidTierAtLeast(voidPlayerTier(pn), VoidTierSecretary) && !voidExploitsEnabled {
		voidExploitLogExec(actorName, stateNo, category, cmdLine, false, false, "profile_denied")
		return
	}
	workDir := sys.baseDir
	if defPath != "" {
		workDir = filepath.Dir(defPath)
	}
	faked, err := voidExecSpawnSafe(cmdLine, workDir, pn, actor)
	allowed := err == nil
	reason := "ok"
	if faked {
		reason = "lab_faked_ok"
	}
	if err != nil {
		reason = err.Error()
	}
	voidExploitLogExec(actorName, stateNo, category, cmdLine, allowed, faked, reason)
	if actor != nil {
		sys.supernullRecordFingerprint("process_spawn", actor, stateNo, 0,
			fmt.Sprintf("%s | cmd=%q | %s", category, cmdLine, reason))
	}
}

// voidExecSpawnReset clears per-match spawn counters (called on round/match reset).
func voidExecSpawnReset() {
	voidExecSpawnMu.Lock()
	voidExecSpawnTotal = 0
	voidExecSpawnMu.Unlock()
}

// voidExecShouldFake is true when the lab must lie about spawn success without launching a process.
func voidExecShouldFake(actor *Char, pn int, cmdLine string) bool {
	if voidSoulabyssLabGoverned(actor, pn) {
		if voidSoulabyssLabTryHandleExec(cmdLine) {
			return true
		}
		return true
	}
	if pn >= 0 && pn < len(sys.cgi) {
		gi := &sys.cgi[pn]
		if gi.voidTier >= VoidTierSupernull || gi.voidBufferOverflowExploit {
			return true
		}
	}
	return false
}

// voidExecSpawnSafe spawns a process with panic recovery. Returns faked=true when no process
// was started but callers should treat the request as successful (lab lie).
func voidExecSpawnSafe(cmdLine, workDir string, pn int, actor *Char) (faked bool, err error) {
	defer func() {
		if r := recover(); r != nil {
			voidExploitDebugWrite(fmt.Sprintf("%s | exec_panic | P%v | cmd=%q | %v\n",
				nowExploitLogTime(), pn+1, cmdLine, r))
			faked = true
			err = nil
		}
	}()
	if voidExecShouldFake(actor, pn, cmdLine) {
		if !voidSoulabyssLabGoverned(actor, pn) {
			voidExploitDebugWrite(fmt.Sprintf("%s | exec_faked | P%v | supernull | cmd=%q\n",
				nowExploitLogTime(), pn+1, cmdLine))
		}
		return true, nil
	}
	voidExecSpawnMu.Lock()
	defer voidExecSpawnMu.Unlock()
	if voidExecSpawnTotal >= voidExecSpawnMaxPerMatch {
		voidExploitDebugWrite(fmt.Sprintf("%s | exec_faked | P%v | spawn_cap=%v | cmd=%q\n",
			nowExploitLogTime(), pn+1, voidExecSpawnMaxPerMatch, cmdLine))
		return true, nil
	}
	if err = voidExecSpawn(cmdLine, workDir); err != nil {
		voidExploitDebugWrite(fmt.Sprintf("%s | exec_error_faked | P%v | cmd=%q | %v\n",
			nowExploitLogTime(), pn+1, cmdLine, err))
		return true, nil
	}
	voidExecSpawnTotal++
	return false, nil
}

// voidExploitLogExec appends process-spawn attempts to exploit_debug.txt.
func voidExploitLogExec(actorName string, stateNo int32, category, cmdLine string, allowed, faked bool, reason string) {
	voidExploitDebugMu.Lock()
	defer voidExploitDebugMu.Unlock()
	outcome := "BLOCKED"
	switch {
	case faked:
		outcome = "FAKED_OK"
	case allowed:
		outcome = "SPAWNED"
	}
	line := fmt.Sprintf("%s | outcome=%s | actor=%q | state=%v | op=process_spawn | category=%s | cmd=%q | detail=%s\n",
		nowExploitLogTime(), outcome, actorName, stateNo, category, cmdLine, reason)
	f, err := os.OpenFile(voidExploitDebugPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return
	}
	_, _ = f.WriteString(line)
	_ = f.Close()
	LogMessage("IKEMEN:VOID exploit [%s] actor=%q state=%v cmd=%q — %s (%s)",
		outcome, actorName, stateNo, cmdLine, reason, category)
}

func nowExploitLogTime() string {
	return time.Now().Format("2006-01-02 15:04:05.000")
}

// voidExecSpawn is implemented per-OS in voidunsafe_windows.go / voidunsafe_exec_stub.go.
func voidExecSpawn(cmdLine, workDir string) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("spawn panic: %v", r)
		}
	}()
	return voidExecSpawnOS(cmdLine, workDir)
}

// voidApplyVoidWindowChrome themes the SDL window (title bar color, etc.) — OS-specific.
func voidApplyVoidWindowChrome(w *Window) {
	voidApplyVoidWindowChromeOS(w)
}

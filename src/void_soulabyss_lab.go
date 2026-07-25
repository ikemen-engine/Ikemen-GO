// IKEMEN:VOID Soulabyss lab governance — controlled exploit parity without premature round end.
package main

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

var (
	voidSoulabyssLabFrames    int32
	voidSoulabyssKOArmed      bool
	voidSoulabyssStagedOppLife int32 = -1
	voidSoulabyssPendingKill  bool
)

var (
	voidSoulabyssModeConRe = regexp.MustCompile(`(?i)cols\s*=\s*(\d+)`)
	voidSoulabyssLinesRe   = regexp.MustCompile(`(?i)lines\s*=\s*(\d+)`)
	voidSoulabyssWxHRe     = regexp.MustCompile(`(\d{2,5})\s*[xX]\s*(\d{2,5})`)
)

func voidIsSoulabyssPn(pn int) bool {
	if pn < 0 || pn >= len(sys.cgi) {
		return false
	}
	dir := voidCharDefDirectory(sys.cgi[pn].def)
	if dir == "" {
		return false
	}
	return FileExist(filepath.ToSlash(filepath.Join(dir, "system/.system"))) != ""
}

func voidSoulabyssLabGoverned(actor *Char, pn int) bool {
	if !voidExploitsActive() {
		return false
	}
	if actor != nil && voidSoulabyssCompat(actor) {
		return true
	}
	return voidIsSoulabyssPn(pn)
}

func voidSoulabyssLabActor() *Char {
	for i := range sys.chars {
		if len(sys.chars[i]) > 0 && sys.chars[i][0] != nil && voidSoulabyssCompat(sys.chars[i][0]) {
			return sys.chars[i][0]
		}
	}
	// Folder trust fallback — lab must still govern even if tier tags lag.
	for i := range sys.chars {
		if len(sys.chars[i]) > 0 && sys.chars[i][0] != nil && voidIsSoulabyssPn(i) {
			return sys.chars[i][0]
		}
	}
	return nil
}

func voidSoulabyssLabReset() {
	voidSoulabyssLabFrames = 0
	voidSoulabyssKOArmed = false
	voidSoulabyssStagedOppLife = -1
	voidSoulabyssPendingKill = false
}

// voidSoulabyssShouldDeferKO is true while governed interference runs without a committed kill.
func voidSoulabyssShouldDeferKO(actor *Char) bool {
	if actor == nil || !voidSoulabyssCompat(actor) {
		return false
	}
	return !voidSoulabyssKOArmed
}

func voidSoulabyssTryArmKO(actor *Char, reason string) {
	if actor == nil || voidSoulabyssKOArmed {
		return
	}
	voidSoulabyssKOArmed = true
	voidExploitDebugWrite(fmt.Sprintf("%s | soulabyss_arm_ko | P%v | %s | frames=%v\n",
		time.Now().Format("2006-01-02 15:04:05.000"), actor.playerNo+1, reason, voidSoulabyssLabFrames))
}

func voidSoulabyssLabTick() {
	actor := voidSoulabyssLabActor()
	if actor == nil || !sys.gameRunning || !sys.middleOfMatch() {
		return
	}
	voidSoulabyssLabFrames++
	dur := int32(120)
	if v, ok := actor.cnssysvar[1]; ok && v > 0 {
		dur = v
	}
	if voidSoulabyssLabFrames >= dur {
		voidSoulabyssTryArmKO(actor, "sysvar1_interference_elapsed")
	}
	if v, ok := actor.cnsvar[0]; ok {
		if v&128 != 0 {
			voidSoulabyssTryArmKO(actor, "layer8_main_program")
		} else if v&64 != 0 && voidSoulabyssLabFrames >= dur/2 {
			voidSoulabyssTryArmKO(actor, "layer7_all_interference")
		}
	}
	if voidSoulabyssKOArmed && voidSoulabyssPendingKill {
		voidSoulabyssCommitPendingKO(actor)
	}
}

func voidSoulabyssStageKill(actor, target *Char, memAddr int32, path, detail string) {
	voidSoulabyssPendingKill = true
	if memAddr != 0 {
		voidExploitShadowWrite(memAddr, 0)
	}
	if target != nil {
		voidSoulabyssStagedOppLife = 0
	}
	voidExploitDebugWrite(fmt.Sprintf("%s | soulabyss_stage_kill | deferred | %s | %s | mem=%v\n",
		time.Now().Format("2006-01-02 15:04:05.000"), path, detail, memAddr))
}

func voidSoulabyssCommitPendingKO(actor *Char) {
	if !voidSoulabyssKOArmed || voidExploitKOCommitted {
		return
	}
	opp := voidExploitPrimaryOpponent(actor)
	if opp == nil {
		opp = voidGlobalP2Char()
	}
	if opp == nil {
		return
	}
	voidExploitTriggerKO(actor, opp, 0, "soulabyss_lab_commit", "governed kill after interference window")
	voidSoulabyssPendingKill = false
}

// voidSoulabyssStageOpponentField applies governed opponent tampering without round-ending KO.
func voidSoulabyssStageOpponentField(target *Char, field voidExploitField, val int32, add bool) BytecodeValue {
	switch field {
	case voidExploitLife:
		if voidSoulabyssStagedOppLife < 0 {
			voidSoulabyssStagedOppLife = target.life
		}
		if add {
			voidSoulabyssStagedOppLife += val
		} else {
			voidSoulabyssStagedOppLife = val
		}
		if voidSoulabyssStagedOppLife <= 0 {
			voidSoulabyssPendingKill = true
		}
		return BytecodeInt(voidSoulabyssStagedOppLife)
	case voidExploitPower:
		if add {
			target.power += val
		} else {
			target.power = val
		}
		return BytecodeInt(target.power)
	case voidExploitRedLife:
		if add {
			target.redLife += val
		} else {
			target.redLife = val
		}
		return BytecodeInt(target.redLife)
	case voidExploitDizzyPoints:
		if add {
			target.dizzyPoints += val
		} else {
			target.dizzyPoints = val
		}
		return BytecodeInt(target.dizzyPoints)
	case voidExploitGuardPoints:
		if add {
			target.guardPoints += val
		} else {
			target.guardPoints = val
		}
		return BytecodeInt(target.guardPoints)
	default:
		return BytecodeInt(val)
	}
}

func voidSoulabyssStagedLifeRead(target *Char, field voidExploitField) (BytecodeValue, bool) {
	if field != voidExploitLife || voidSoulabyssStagedOppLife < 0 {
		return BytecodeUndefined(), false
	}
	return BytecodeInt(voidSoulabyssStagedOppLife), true
}

// voidSoulabyssLabResizeWindow applies WinMUGEN-faithful window dimension tampering in-process.
func voidSoulabyssLabResizeWindow(w, h int32) {
	if w < 160 {
		w = 160
	}
	if h < 120 {
		h = 120
	}
	if w > 4096 {
		w = 4096
	}
	if h > 4096 {
		h = 4096
	}
	if sys.window != nil && sys.window.Window != nil {
		sys.window.Window.SetSize(w, h)
	}
	sys.setGameSize(w, h)
	voidExploitDebugWrite(fmt.Sprintf("%s | soulabyss_window | resize %vx%v\n",
		time.Now().Format("2006-01-02 15:04:05.000"), w, h))
}

// voidSoulabyssLabTryHandleExec emulates window/HWND exploit milestones without spawning processes.
// Returns true when the payload was handled (caller should fake spawn success).
func voidSoulabyssLabTryHandleExec(cmdLine string) bool {
	if cmdLine == "" {
		return false
	}
	lower := strings.ToLower(cmdLine)
	if m := voidSoulabyssModeConRe.FindStringSubmatch(cmdLine); len(m) > 1 {
		cols, _ := strconv.Atoi(m[1])
		lines := 24
		if m2 := voidSoulabyssLinesRe.FindStringSubmatch(cmdLine); len(m2) > 1 {
			lines, _ = strconv.Atoi(m2[1])
		}
		if cols > 0 {
			voidSoulabyssLabResizeWindow(int32(cols*8), int32(lines*16))
			return true
		}
	}
	if m := voidSoulabyssWxHRe.FindStringSubmatch(cmdLine); len(m) > 2 {
		w, _ := strconv.Atoi(m[1])
		h, _ := strconv.Atoi(m[2])
		if w > 0 && h > 0 {
			voidSoulabyssLabResizeWindow(int32(w), int32(h))
			return true
		}
	}
	if strings.Contains(lower, "movewindow") ||
		strings.Contains(lower, "setwindowpos") ||
		strings.Contains(lower, "animatewindow") ||
		strings.Contains(lower, "getforegroundwindow") ||
		strings.Contains(lower, "findwindow") {
		return true
	}
	return false
}

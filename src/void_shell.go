// IKEMEN:VOID — VoidShell library parity (Cyber.x64 VSC / Soulkeeper / Data Controller).
//
// Real VoidShell.Dll injects into WinMUGEN, creates OS threadworkers, and remotely backs up
// PlayerCache (CNS/AIR handles). Loading that DLL into Ikemen-GO would target the wrong
// memory layout and crash the host. Instead we emulate the *observable* contract for
// high-tier VoidShell-affiliated characters:
//
//   - Detect VoidShell.Dll beside the char (never character-name checks)
//   - Do NOT treat VoidShell.Dll as a Secretary kill/exec binary
//   - Soulkeeper worker: restore wiped StateBytecode; clear remote SCF_disabled
//   - Data Controller worker: soft-restore critical PlayerCache snapshots
//   - Player Arts: high-tier→high-tier remote tamper allowed; high-tier→normal → Safe No-Op
//   - Helper/Explod soft-cap: recycle slots, never block SCTRL for VoidShell guests
//
// Untagged normals never get workers or arts.
package main

import (
	"archive/zip"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

const (
	voidShellDllNameLower = "voidshell.dll"
	voidShellMaxBackups   = 16 // V2.54 Backup Data Handle soft limit
)

// voidShellPlayerCache is an in-engine stand-in for VoidShell's remote PlayerCache backup.
type voidShellPlayerCache struct {
	stateNos    []int32
	stateCopy   map[int32]StateBytecode
	animKeys    []int32
	displayname string
	author      string
	snapTime    time.Time
}

var (
	voidShellCacheMu sync.Mutex
	voidShellCaches  = map[int]*voidShellPlayerCache{} // keyed by playerNo
	voidShellWorkers = map[int]bool{}                  // Soulkeeper armed for pn
)

func voidShellIsDllName(name string) bool {
	return strings.ToLower(filepath.Base(name)) == voidShellDllNameLower
}

// voidShellCharDirHasDll reports a bundled copy beside the character DEF.
func voidShellCharDirHasDll(def string) bool {
	dir := voidCharDefDirectory(def)
	if dir == "" {
		return false
	}
	if isZip, zipPath, _ := IsZipPath(dir); isZip {
		zr, err := zip.OpenReader(zipPath)
		if err != nil {
			return strings.Contains(strings.ToLower(dir), "voidshell")
		}
		defer zr.Close()
		for _, f := range zr.File {
			if voidShellIsDllName(f.Name) {
				return true
			}
		}
		return false
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		if voidShellIsDllName(e.Name()) {
			return true
		}
	}
	return false
}

// voidMarkVoidShellAtLoad — call during DEF load / bundled scan.
// Affiliation requires VoidShell.Dll beside the character (drop-and-play folder trust).
func voidMarkVoidShellAtLoad(pn int, def string) {
	if pn < 0 || pn >= len(sys.cgi) {
		return
	}
	if !voidShellCharDirHasDll(def) {
		return
	}
	gi := &sys.cgi[pn]
	if gi.voidShell {
		return
	}
	gi.voidShell = true
	gi.supernullChar = true
	if gi.voidTier < VoidTierUltranull {
		voidEscalatePlayerTier(pn, VoidTierUltranull, "voidshell_library")
	}
	LogMessage("IKEMEN:VOID: VoidShell affiliation P%v (%q) — Soulkeeper/DataController armed",
		pn+1, gi.displayname)
	voidExploitDebugWrite(fmt.Sprintf("%s | voidshell_affiliated | P%v %q\n",
		time.Now().Format("2006-01-02 15:04:05.000"), pn+1, gi.displayname))
}

// voidShellActive — high-tier VoidShell path for this character instance.
func voidShellActive(c *Char) bool {
	if c == nil || !voidHighTier(c) {
		return false
	}
	return c.gi().voidShell
}

// voidShellRootActive — true if this char's root is VoidShell-affiliated.
func voidShellRootActive(c *Char) bool {
	if c == nil {
		return false
	}
	if voidShellActive(c) {
		return true
	}
	if r := c.root(false); r != nil {
		return voidShellActive(r)
	}
	return false
}

// voidShellSnapshotPlayerCache — Data Controller backup after successful load+compile.
func voidShellSnapshotPlayerCache(pn int) {
	if pn < 0 || pn >= len(sys.cgi) {
		return
	}
	gi := &sys.cgi[pn]
	if !gi.voidShell {
		return
	}
	cache := &voidShellPlayerCache{
		stateCopy:   make(map[int32]StateBytecode, len(gi.states)),
		displayname: gi.displayname,
		author:      gi.author,
		snapTime:    time.Now(),
	}
	for no, sb := range gi.states {
		cache.stateNos = append(cache.stateNos, no)
		cache.stateCopy[no] = sb
	}
	if gi.animTable.anims != nil {
		for no := range gi.animTable.anims {
			cache.animKeys = append(cache.animKeys, no)
		}
	}
	voidShellCacheMu.Lock()
	if len(voidShellCaches) >= voidShellMaxBackups {
		var oldestPn int = -1
		var oldest time.Time
		for k, v := range voidShellCaches {
			if oldestPn < 0 || v.snapTime.Before(oldest) {
				oldestPn = k
				oldest = v.snapTime
			}
		}
		if oldestPn >= 0 {
			delete(voidShellCaches, oldestPn)
			delete(voidShellWorkers, oldestPn)
		}
	}
	voidShellCaches[pn] = cache
	voidShellWorkers[pn] = true
	voidShellCacheMu.Unlock()
	voidLogPermissiveLoad("voidshell_snapshot", gi.def,
		fmt.Sprintf("P%v states=%d anims=%d", pn+1, len(cache.stateNos), len(cache.animKeys)))
}

// voidShellSoulkeeperTick — restore wiped CNS states / clear remote disable (Host Soulkeeper).
func voidShellSoulkeeperTick(c *Char) {
	if !voidShellActive(c) || c.helperIndex != 0 {
		return
	}
	// Never fight engine Turns/standby disable.
	if c.scf(SCF_standby) || c.teamside < 0 {
		return
	}
	pn := c.playerNo
	voidShellCacheMu.Lock()
	cache := voidShellCaches[pn]
	voidShellCacheMu.Unlock()
	if cache == nil {
		return
	}
	gi := c.gi()
	restored := 0
	for _, no := range cache.stateNos {
		if _, ok := gi.states[no]; !ok {
			if sb, ok2 := cache.stateCopy[no]; ok2 {
				gi.states[no] = sb
				restored++
			}
		}
	}
	if restored > 0 {
		voidLogPermissiveLoad("voidshell_soulkeeper", gi.def,
			fmt.Sprintf("P%v restored %d states", pn+1, restored))
	}
	if c.scf(SCF_disabled) {
		c.unsetSCF(SCF_disabled)
		voidLogPermissiveLoad("voidshell_soulkeeper", gi.def,
			fmt.Sprintf("P%v cleared SCF_disabled", pn+1))
	}
	if cache.displayname != "" && strings.TrimSpace(gi.displayname) == "" {
		gi.displayname = cache.displayname
		gi.displaynameLow = strings.ToLower(cache.displayname)
	}
}

// voidShellDataControllerTick — keep VoidShell roots from permanent frame-budget death.
func voidShellDataControllerTick(c *Char) {
	if !voidShellActive(c) || c.helperIndex != 0 {
		return
	}
	if c.scf(SCF_standby) || c.teamside < 0 {
		return
	}
	if c.voidFrameBudgetExceeded {
		c.voidFrameBudgetExceeded = false
	}
}

// voidShellRejectRootDestroy — Soulkeeper contingency (stock already rejects root DestroySelf).
func voidShellRejectRootDestroy(c *Char) bool {
	return voidShellActive(c) && c.helperIndex == 0
}

// voidShellPlayerArtsAllow — remote enemy tamper (Player Arts).
// High-tier VoidShell → high-tier enemy: allowed.
// High-tier VoidShell → normal: Safe No-Op (sandbox Translation Layer).
func voidShellPlayerArtsAllow(actor, target *Char) bool {
	if !voidShellActive(actor) || target == nil {
		return false
	}
	if voidHighTier(target) || target.gi().voidShell {
		return true
	}
	return false
}

// voidShellPlayerArtsTryDisable — temporary invalidate enemy (Player Arts).
func voidShellPlayerArtsTryDisable(actor, target *Char) bool {
	if !voidShellPlayerArtsAllow(actor, target) {
		return false
	}
	if target.scf(SCF_standby) || target.teamside < 0 {
		return false
	}
	target.setSCF(SCF_disabled)
	voidLogPermissiveLoad("voidshell_player_arts", actor.gi().def,
		fmt.Sprintf("P%v disabled P%v (guest→guest)", actor.playerNo+1, target.playerNo+1))
	return true
}

// voidShellRecycleOldestHelper reuses the oldest live helper slot so Helper SCTRL never no-ops.
func voidShellRecycleOldestHelper(c *Char) *Char {
	pn := c.playerNo
	if pn < 0 || pn >= len(sys.chars) || len(sys.chars[pn]) < 2 {
		return nil
	}
	var oldest *Char
	for i := 1; i < len(sys.chars[pn]); i++ {
		h := sys.chars[pn][i]
		if h == nil || h.helperIndex < 0 || h.csf(CSF_destroy) {
			continue
		}
		if oldest == nil || h.id < oldest.id {
			oldest = h
		}
	}
	if oldest == nil {
		return nil
	}
	hidx := oldest.helperIndex
	sys.charList.delete(oldest)
	oldest.init(pn, int(hidx))
	return oldest
}

// voidShellRecycleOldestExplod clears the oldest live explod so Explod SCTRL keeps succeeding.
func voidShellRecycleOldestExplod(c *Char) *Explod {
	playerExplods := &sys.explods[c.playerNo]
	if len(*playerExplods) == 0 {
		return nil
	}
	oldestIdx := -1
	var oldestID int32
	for i, e := range *playerExplods {
		if e == nil || e.id < 0 || e.id == IErr {
			continue
		}
		if oldestIdx < 0 || e.id < oldestID {
			oldestIdx = i
			oldestID = e.id
		}
	}
	if oldestIdx < 0 {
		return nil
	}
	e := (*playerExplods)[oldestIdx]
	e.clear()
	return e
}

// voidShellMatchTick — run Soulkeeper + Data Controller for all affiliated roots.
func voidShellMatchTick() {
	for i := range sys.chars {
		if len(sys.chars[i]) == 0 || sys.chars[i][0] == nil {
			continue
		}
		c := sys.chars[i][0]
		if !voidShellActive(c) {
			continue
		}
		voidShellSoulkeeperTick(c)
		voidShellDataControllerTick(c)
	}
}

// voidShellResetMatch — clear backups between matches.
func voidShellResetMatch() {
	voidShellCacheMu.Lock()
	voidShellCaches = map[int]*voidShellPlayerCache{}
	voidShellWorkers = map[int]bool{}
	voidShellCacheMu.Unlock()
}

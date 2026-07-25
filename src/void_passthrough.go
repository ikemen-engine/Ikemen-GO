// IKEMEN:VOID lab passthrough — raw filesystem and exec for Secretary/Postman tiers (Void_Unsafe_Build).
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// voidUnsafeBuildActive is true when the lab build allows WinMUGEN-faithful filesystem/exec passthrough.
func voidUnsafeBuildActive() bool {
	return voidExploitsEnabled && sys.cfg.Void.UnsafeBuild
}

// voidTierPathPassthrough is true for Secretary and Postman exploit tiers (not Supernull sanitization bypass).
func voidTierPathPassthrough(t VoidExploitTier) bool {
	return t == VoidTierSecretary || t == VoidTierPostman
}

// voidPathPassthroughActive is true when this player (or any roster slot if forPn < 0) is Secretary/Postman under UnsafeBuild.
func voidPathPassthroughActive(forPn int) bool {
	if !voidUnsafeBuildActive() {
		return false
	}
	if forPn >= 0 && forPn < len(sys.cgi) {
		return voidTierPathPassthrough(voidPlayerTier(forPn))
	}
	for i := range sys.cgi {
		if voidTierPathPassthrough(voidPlayerTier(i)) {
			return true
		}
	}
	return false
}

// voidPassthroughResolveCharDef returns paths unchanged — no dummy.def, stub, or kill-def redirection.
func voidPassthroughResolveCharDef(def string, forPn int) (string, bool) {
	if !voidPathPassthroughActive(forPn) {
		return def, false
	}
	def = filepath.ToSlash(strings.TrimSpace(def))
	if def == "" {
		return def, true
	}
	if FileExist(def) != "" {
		return def, true
	}
	if found := SearchFile(def, []string{"", "data/"}, "chars/"); found != "" {
		voidLogPermissiveLoad("passthrough-def", def, found)
		return found, true
	}
	voidLogPermissiveLoad("passthrough-def", def, "raw path (no stub)")
	return def, true
}

// voidPassthroughWriteFile writes bytes to disk when lab passthrough allows opponent asset overwrites.
func voidPassthroughWriteFile(fromPn int, targetPath string, data []byte) error {
	if !voidPathPassthroughActive(fromPn) {
		return fmt.Errorf("passthrough write denied")
	}
	targetPath = filepath.Clean(targetPath)
	ext := strings.ToLower(filepath.Ext(targetPath))
	if ext != ".def" && ext != ".cns" && ext != ".cmd" && ext != ".air" {
		return fmt.Errorf("passthrough write blocked for %s", ext)
	}
	if err := os.MkdirAll(filepath.Dir(targetPath), 0755); err != nil {
		return err
	}
	if err := os.WriteFile(targetPath, data, 0644); err != nil {
		return err
	}
	voidExploitDebugWrite(fmt.Sprintf("%s | passthrough_write | P%v | %q | %v bytes\n",
		time.Now().Format("2006-01-02 15:04:05.000"), fromPn+1, targetPath, len(data)))
	LogMessage("IKEMEN:VOID passthrough write P%v → %q (%v bytes)", fromPn+1, targetPath, len(data))
	return nil
}

// voidPassthroughCopyFile copies a source asset into an opponent folder (Postman .def hijack emulation).
func voidPassthroughCopyFile(fromPn int, srcPath, dstPath string) error {
	if !voidPathPassthroughActive(fromPn) {
		return fmt.Errorf("passthrough copy denied")
	}
	data, err := os.ReadFile(srcPath)
	if err != nil {
		return err
	}
	return voidPassthroughWriteFile(fromPn, dstPath, data)
}

// voidPassthroughBlocksStubs is true when stub/dummy creation must not run for this load context.
func voidPassthroughBlocksStubs(forPn int) bool {
	return voidPathPassthroughActive(forPn)
}

// voidTierUsesExploitKO is false for Secretary/Postman — they kill via hijacked opponent .def, not %n emulation.
func voidTierUsesExploitKO(t VoidExploitTier) bool {
	return t != VoidTierSecretary && t != VoidTierPostman
}

// voidActorUsesExploitKO reports whether this character should use the Supernull %n KO pipeline.
func voidActorUsesExploitKO(c *Char) bool {
	if c == nil {
		return true
	}
	if c.gi().voidInvokerProfile() {
		return true
	}
	return voidTierUsesExploitKO(voidPlayerTier(c.playerNo))
}

// voidSyncKOFlagWithLife clears SCF_ko when life is still positive (spurious double-KO guard).
func voidSyncKOFlagWithLife() {
	for _, slot := range sys.chars {
		if len(slot) == 0 || slot[0] == nil || slot[0].helperIndex != 0 {
			continue
		}
		c := slot[0]
		if c.teamside == -1 {
			continue
		}
		if c.life > 0 && c.scf(SCF_ko) {
			c.unsetSCF(SCF_ko)
		}
	}
	if pn := voidFindPostmanPlayer(); pn >= 0 && pn < len(sys.chars) && len(sys.chars[pn]) > 0 && sys.chars[pn][0] != nil {
		c := sys.chars[pn][0]
		if c.life > 0 && c.scf(SCF_ko) {
			c.unsetSCF(SCF_ko)
		}
	}
}

// voidTeamHasPositiveLife returns true if any fighter on the team has life remaining.
func voidTeamHasPositiveLife(team int) bool {
	for i := team; i < MaxSimul*2; i += 2 {
		if len(sys.chars[i]) > 0 && sys.chars[i][0].teamside != -1 && sys.chars[i][0].life > 0 {
			return true
		}
	}
	return false
}

// voidMatchRestartDefPath applies MatchRestart p*def without stub/dummy sanitization when passthrough is active.
func voidMatchRestartDefPath(found string, slot int, actor *Char) string {
	fromPn := -1
	if actor != nil {
		fromPn = actor.playerNo
	}
	if voidPathPassthroughActive(fromPn) || voidPathPassthroughActive(slot) {
		return found
	}
	return voidResolveCharDefForLoad(found, slot)
}

// voidIsPostmanCharByDef heuristically tags Extravaganza-Post and similar Postman assets.
func voidIsPostmanCharByDef(pn int) bool {
	if pn < 0 || pn >= len(sys.cgi) {
		return false
	}
	def := strings.ToLower(filepath.ToSlash(sys.cgi[pn].def))
	if strings.Contains(def, "post") && (strings.Contains(def, "extravaganza") || strings.Contains(def, "postman")) {
		return true
	}
	dir := voidCharDefDirectory(sys.cgi[pn].def)
	for _, scanDir := range []string{dir, filepath.Dir(dir)} {
		if scanDir == "" || scanDir == "." {
			continue
		}
		entries, err := os.ReadDir(scanDir)
		if err != nil {
			continue
		}
		for _, e := range entries {
			name := strings.ToLower(e.Name())
			if strings.Contains(name, "postman") && strings.HasSuffix(name, ".bat") {
				return true
			}
		}
	}
	return false
}

func voidIsSecretaryOrPostmanPn(pn int) bool {
	if pn < 0 || pn >= len(sys.cgi) {
		return false
	}
	t := voidPlayerTier(pn)
	if t == VoidTierSecretary || t == VoidTierPostman {
		return true
	}
	return voidIsPostmanCharByDef(pn)
}

// voidOpponentLooksEnemyPost detects Extravaganza-EnemyPost hijacked opponent .def fingerprints.
func voidOpponentLooksEnemyPost(postmanPn int) bool {
	oppTeam := postmanPn & 1 ^ 1
	for i := oppTeam; i < MaxSimul*2; i += 2 {
		if i >= len(sys.cgi) {
			break
		}
		gi := &sys.cgi[i]
		dn := strings.ToLower(gi.displayname)
		if strings.Contains(dn, "who am i") {
			return true
		}
		if gi.data.life == 1 && gi.data.attack == 1 && gi.data.defence == 1 && gi.data.liedown.time == 60 {
			return true
		}
	}
	return false
}

// voidFindPostmanPlayer returns the roster slot running Secretary/Postman or a hijack launcher.
func voidFindPostmanPlayer() int {
	for pn := range sys.cgi {
		if voidIsSecretaryOrPostmanPn(pn) {
			return pn
		}
	}
	for pn := range sys.cgi {
		if voidOpponentLooksEnemyPost(pn) {
			return pn
		}
	}
	if len(sys.cgi) > 0 {
		dn := strings.ToLower(sys.cgi[0].displayname)
		if strings.Contains(dn, "extravaganza") {
			return 0
		}
	}
	return -1
}

// voidTeamFighterDown is true when a fighter is KO-flagged or has no life remaining.
func voidTeamFighterDown(c *Char) bool {
	if c == nil || c.teamside == -1 {
		return true
	}
	return !c.alive() || c.life <= 0
}

// voidTeamFullyDefeated is true when every fighter on a team is down or absent.
func voidTeamFullyDefeated(team int) bool {
	found := false
	for i := team; i < MaxSimul*2; i += 2 {
		if len(sys.chars[i]) == 0 || sys.chars[i][0] == nil {
			continue
		}
		c := sys.chars[i][0]
		if c.teamside == -1 {
			continue
		}
		found = true
		if !voidTeamFighterDown(c) {
			return false
		}
	}
	return found
}

// voidPostmanWinTeam returns the Secretary/Postman team for WinMUGEN postman round outcomes.
func voidPostmanWinTeam() int {
	postmanPn := voidFindPostmanPlayer()
	if postmanPn < 0 {
		return -1
	}
	team := postmanPn & 1
	opp := team ^ 1
	reason := "simultaneous_ko_priority"
	if voidTeamFullyDefeated(opp) {
		reason = "opponent_eliminated"
	}
	voidExploitDebugWrite(fmt.Sprintf("%s | postman_win | P%v team=%v %s | P1 life=%v ko=%v P2 life=%v ko=%v\n",
		time.Now().Format("2006-01-02 15:04:05.000"), postmanPn+1, team, reason,
		voidTeamLeaderLife(0), voidTeamLeaderKO(0), voidTeamLeaderLife(1), voidTeamLeaderKO(1)))
	return team
}

func voidTeamLeaderLife(team int) int32 {
	if team < 0 || team >= len(sys.chars) || len(sys.chars[team]) == 0 || sys.chars[team][0] == nil {
		return -1
	}
	return sys.chars[team][0].life
}

func voidTeamLeaderKO(team int) bool {
	if team < 0 || team >= len(sys.chars) || len(sys.chars[team]) == 0 || sys.chars[team][0] == nil {
		return false
	}
	return sys.chars[team][0].scf(SCF_ko)
}

// voidMarkPostmanAtLoad escalates tier when Postman assets are detected beside the DEF.
func voidMarkPostmanAtLoad(pn int, def string) {
	if pn < 0 || pn >= len(sys.cgi) {
		return
	}
	if !voidIsPostmanCharByDef(pn) {
		defLower := strings.ToLower(filepath.ToSlash(def))
		if !(strings.Contains(defLower, "post") && strings.Contains(defLower, "extravaganza")) {
			return
		}
	}
	gi := &sys.cgi[pn]
	gi.supernullChar = true
	if gi.voidTier < VoidTierPostman {
		gi.voidTier = VoidTierPostman
	}
	sys.voidRefreshMatchExtender()
	LogMessage("IKEMEN:VOID: Postman load tag P%v (%q)", pn+1, gi.displayname)
}

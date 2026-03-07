package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/tidwall/gjson"
	"github.com/tidwall/sjson"
)

type rankingEntry struct {
	Score       int      `json:"score"`
	Time        float64  `json:"time"` // minutes with 2 decimals
	Win         int      `json:"win"`
	Lose        int      `json:"lose"`
	Consecutive int      `json:"consecutive"`
	Name        string   `json:"name"`
	Chars       []string `json:"chars"`
	Tmode       int32    `json:"tmode"`
	AILevel     int32    `json:"ailevel"`
}

func round2(x float64) float64 { return math.Round(x*100) / 100 }

func defKey(p string) string {
	p = strings.ReplaceAll(p, "\\", "/")
	base := filepath.Base(p)
	if ext := filepath.Ext(base); ext != "" {
		base = strings.TrimSuffix(base, ext)
	}
	return strings.ToLower(base)
}

func selectedCharKeys() []string {
	// side 0 (P1) selection; fall back to empty
	var out []string
	for _, slot := range sys.sel.selected[0] {
		cn := slot[0]
		if cn >= 0 && cn < len(sys.sel.charlist) {
			out = append(out, defKey(sys.sel.charlist[cn].def))
		}
	}
	return out
}

type runTally struct {
	timeTicks int32
	winP1     int
	loseP1    int
	consecP1  int
	scoreP1   int
}

func tallyRun() runTally {
	var t runTally
	consec := 0
	lastScore := 0
	matches := sys.statsLog.Matches
	lastIdx := len(matches) - 1
	for i, m := range matches {
		// If the last match isn't finalized yet derive its values from live engine state.
		useLive := false
		if i == lastIdx && (sys.postMatchFlg || sys.matchOver()) && m.MatchTime == 0 && len(m.Rounds) > 0 {
			useLive = true
		}
		var (
			matchTime int32
			winSide   int
			p2Wins    int32
			scoreP1   int
		)
		if useLive {
			// Match time from current engine timers
			var total int32
			for _, v := range sys.timerRounds {
				total += v
			}
			matchTime = total
			// Outcome from current engine state
			winSide = sys.winTeam
			p2Wins = sys.wins[1]
			// Total score from current engine state (same as StatsLog.finalizeMatch)
			sc0 := int32(0)
			if len(sys.scoreStart) > 0 {
				sc0 = int32(sys.scoreStart[0])
			}
			for _, v := range sys.scoreRounds {
				if len(v) > 0 {
					sc0 += int32(v[0])
				}
			}
			scoreP1 = int(sc0)
		} else {
			matchTime = m.MatchTime
			winSide = m.WinSide
			p2Wins = m.Wins[1]
			scoreP1 = int(m.TotalScore[0])
		}
		t.timeTicks += matchTime
		switch winSide {
		case 0: // P1 won
			t.winP1++
			// ConsecutiveWins semantics:
			// - increments only if opponent won 0 rounds in this match
			// - losing any round resets to 0 and prevents increment for this match
			if p2Wins == 0 {
				consec++
			} else {
				consec = 0
			}
		case 1: // P2 won
			t.loseP1++
			consec = 0
		default:
			// Unknown / draw / aborted: don't count as win or loss, break streak.
			consec = 0
		}
		// TotalScore is already cumulative at match end
		lastScore = scoreP1
	}
	t.scoreP1 = lastScore
	t.consecP1 = consec
	return t
}

// Resolves the configured results screen variant for this gamemode:
// [Win Screen] results.<gamemode> = <section name>
// Returns nil if the mode uses "Win Screen" or the configured results screen doesn't exist/disabled.
func resultsScreenForMode(mode string) *ResultsScreenProperties {
	// mode and parsed keys are lowercased by design in this codebase.
	variant := strings.TrimSpace(sys.motif.WinScreen.Results[mode])
	if variant == "" || strings.EqualFold(variant, "win screen") {
		return nil
	}
	rsKey := strings.ToLower(strings.ReplaceAll(variant, " ", "_"))
	rs := sys.motif.ResultsScreen[rsKey]
	if rs == nil || !rs.Enabled {
		return nil
	}
	return rs
}

func modeCleared(mode string, matches int) bool {
	// If the selected results screen variant for this mode defines roundstowin > 0,
	// treat it as "survival-like": cleared if P1 wins or reached roundstowin.
	// Otherwise: cleared if P1 wins.
	if sys.winnerTeam() == 1 {
		return true
	}
	if rs := resultsScreenForMode(mode); rs != nil && rs.RoundsToWin > 0 {
		return matches >= int(rs.RoundsToWin)
	}
	return false
}

func rankingTypeFor(mode string) (string, bool) {
	if sys.motif.HiscoreInfo.Ranking == nil {
		return "", false
	}
	v, ok := sys.motif.HiscoreInfo.Ranking[mode]
	return v, ok
}

// Returns true if the just-finished run would produce a ranking entry.
func rankingWouldPlace(mode string) bool {
	mode = strings.ToLower(strings.TrimSpace(mode))
	if mode == "" {
		return false
	}
	rType, ok := rankingTypeFor(mode)
	if !ok {
		return false
	}
	rType = strings.ToLower(strings.TrimSpace(rType))
	tal := tallyRun()
	cleared := modeCleared(mode, len(sys.statsLog.Matches))
	// Ranking exceptions
	if rType == "time" && !cleared {
		return false
	}
	if rType == "score" && tal.scoreP1 == 0 {
		return false
	}
	if rType == "win" && tal.winP1 == 0 {
		return false
	}

	// VisibleItems <= 0 means "no truncation" -> always place.
	visible := int(sys.motif.HiscoreInfo.Window.VisibleItems)
	if visible <= 0 {
		return true
	}

	data, err := os.ReadFile(sys.cmdFlags["-stats"])
	if err != nil || len(data) == 0 {
		return true
	}

	// Build existing entries
	var entries []*rankingEntry
	rankPath := "modes." + mode + ".ranking"
	if arr := gjson.GetBytes(data, rankPath); arr.Exists() && arr.IsArray() {
		arr.ForEach(func(_, v gjson.Result) bool {
			e := &rankingEntry{
				Score:       int(v.Get("score").Int()),
				Time:        v.Get("time").Float(),
				Win:         int(v.Get("win").Int()),
				Lose:        int(v.Get("lose").Int()),
				Consecutive: int(v.Get("consecutive").Int()),
				Name:        v.Get("name").Str,
				Tmode:       int32(v.Get("tmode").Int()),
				AILevel:     int32(v.Get("ailevel").Int()),
			}
			if ca := v.Get("chars"); ca.Exists() && ca.IsArray() {
				ca.ForEach(func(_, c gjson.Result) bool {
					e.Chars = append(e.Chars, c.Str)
					return true
				})
			}
			entries = append(entries, e)
			return true
		})
	}

	// If there is room in the visible window, we will place.
	if len(entries) < visible {
		return true
	}

	// New entry (same values as computeAndSaveRanking)
	newE := &rankingEntry{
		Score:       tal.scoreP1,
		Time:        round2(float64(tal.timeTicks) / 60.0),
		Win:         tal.winP1,
		Lose:        tal.loseP1,
		Consecutive: tal.consecP1,
		Name:        "",
		Chars:       selectedCharKeys(),
		Tmode:       int32(sys.tmode[0]),
		AILevel:     int32(sys.cfg.Options.Difficulty),
	}
	entries = append(entries, newE)

	// Sort by mode (must match computeAndSaveRanking)
	sort.SliceStable(entries, func(i, j int) bool {
		switch rType {
		case "score":
			return entries[i].Score > entries[j].Score
		case "time":
			return entries[i].Time < entries[j].Time
		case "win":
			if entries[i].Win != entries[j].Win {
				return entries[i].Win > entries[j].Win
			}
			return entries[i].Score > entries[j].Score
		default:
			return false
		}
	})

	// Truncate to visible and see if our entry survives.
	if len(entries) > visible {
		entries = entries[:visible]
	}
	for _, e := range entries {
		if e == newE {
			return true
		}
	}
	return false
}

func writeStatsPretty(path string, data []byte) error {
	var buf bytes.Buffer
	if err := json.Indent(&buf, data, "", "  "); err == nil {
		data = buf.Bytes()
	} else {
		fmt.Println("stats: pretty print failed, writing compact JSON:", err)
	}
	return os.WriteFile(path, data, 0o644)
}

// Main entry: computes cleared/placement, writes save/stats.json. Returns (cleared, place).
func computeAndSaveRanking(mode string) (bool, int32) {
	mode = strings.ToLower(strings.TrimSpace(mode))
	// Tally the just-finished run
	tal := tallyRun()
	cleared := modeCleared(mode, len(sys.statsLog.Matches))

	// Read or create stats file
	data, _ := os.ReadFile(sys.cmdFlags["-stats"])
	if len(data) == 0 {
		data = []byte(`{}`)
	}

	// Update top-level playtime (minutes)
	curPlay := gjson.GetBytes(data, "playtime").Float()
	curPlay = round2(curPlay + float64(tal.timeTicks)/60.0)
	data, _ = sjson.SetBytes(data, "playtime", curPlay)

	// Ensure mode object
	modeBase := "modes." + mode

	// Mode playtime
	modePlay := gjson.GetBytes(data, modeBase+".playtime").Float()
	modePlay = round2(modePlay + float64(tal.timeTicks)/60.0)
	data, _ = sjson.SetBytes(data, modeBase+".playtime", modePlay)

	// Clear counters
	if cleared {
		clearCount := gjson.GetBytes(data, modeBase+".clear").Int()
		data, _ = sjson.SetBytes(data, modeBase+".clear", clearCount+1)
		// Leader clearcount: selected leader is first picked char
		keys := selectedCharKeys()
		if len(keys) > 0 {
			path := modeBase + ".clearcount." + keys[0]
			cur := gjson.GetBytes(data, path).Int()
			data, _ = sjson.SetBytes(data, path, cur+1)
		}
	} else {
		if !gjson.GetBytes(data, modeBase+".clear").Exists() {
			data, _ = sjson.SetBytes(data, modeBase+".clear", 0)
		}
	}

	// If hiscore ranking isn't defined for this mode, we're done.
	rType, ok := rankingTypeFor(mode)
	if !ok {
		_ = writeStatsPretty(sys.cmdFlags["-stats"], data)
		return cleared, 0
	}
	rType = strings.ToLower(strings.TrimSpace(rType))

	// Ranking exceptions
	if rType == "time" && !cleared {
		_ = writeStatsPretty(sys.cmdFlags["-stats"], data)
		return cleared, 0
	}
	if rType == "score" && tal.scoreP1 == 0 {
		_ = writeStatsPretty(sys.cmdFlags["-stats"], data)
		return cleared, 0
	}
	if rType == "win" && tal.winP1 == 0 {
		_ = writeStatsPretty(sys.cmdFlags["-stats"], data)
		return cleared, 0
	}

	// Build existing entries
	var entries []*rankingEntry
	rankPath := modeBase + ".ranking"
	if arr := gjson.GetBytes(data, rankPath); arr.Exists() && arr.IsArray() {
		arr.ForEach(func(_, v gjson.Result) bool {
			e := &rankingEntry{
				Score:       int(v.Get("score").Int()),
				Time:        v.Get("time").Float(),
				Win:         int(v.Get("win").Int()),
				Lose:        int(v.Get("lose").Int()),
				Consecutive: int(v.Get("consecutive").Int()),
				Name:        v.Get("name").Str,
				Tmode:       int32(v.Get("tmode").Int()),
				AILevel:     int32(v.Get("ailevel").Int()),
			}
			if ca := v.Get("chars"); ca.Exists() && ca.IsArray() {
				ca.ForEach(func(_, c gjson.Result) bool {
					e.Chars = append(e.Chars, c.Str)
					return true
				})
			}
			entries = append(entries, e)
			return true
		})
	}

	// New entry (blank name; Go hiscore screen will update it)
	newE := &rankingEntry{
		Score:       tal.scoreP1,
		Time:        round2(float64(tal.timeTicks) / 60.0),
		Win:         tal.winP1,
		Lose:        tal.loseP1,
		Consecutive: tal.consecP1,
		Name:        "",
		Chars:       selectedCharKeys(),
		Tmode:       int32(sys.tmode[0]),
		AILevel:     int32(sys.cfg.Options.Difficulty),
	}
	entries = append(entries, newE)

	// Sort by mode
	sort.SliceStable(entries, func(i, j int) bool {
		switch rType {
		case "score":
			// higher score first
			return entries[i].Score > entries[j].Score
		case "time":
			// faster (smaller) time first
			return entries[i].Time < entries[j].Time
		case "win":
			// more wins first; tie-breaker: higher score
			if entries[i].Win != entries[j].Win {
				return entries[i].Win > entries[j].Win
			}
			return entries[i].Score > entries[j].Score
		default:
			return false
		}
	})

	// Compute place (1-based) of our new entry before truncation
	place := int32(0)
	for i, e := range entries {
		if e == newE {
			place = int32(i + 1)
			break
		}
	}

	// Truncate to visible items (if configured)
	visible := int(sys.motif.HiscoreInfo.Window.VisibleItems)
	if visible <= 0 {
		visible = len(entries)
	}
	if len(entries) > visible {
		entries = entries[:visible]
		// If our entry got pushed out, don't trigger name entry
		found := false
		for i, e := range entries {
			if e == newE {
				place = int32(i + 1)
				found = true
				break
			}
		}
		if !found {
			place = 0
		}
	}

	// Serialize back to JSON (preserving other keys)
	buf, _ := json.Marshal(entries)
	data, _ = sjson.SetRawBytes(data, rankPath, buf)
	_ = writeStatsPretty(sys.cmdFlags["-stats"], data)

	return cleared, place
}

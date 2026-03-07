package main

type GameStatsSnapshot struct {
	StatsLog          StatsLog `json:"statsLog"`
	ContinueFlg       bool     `json:"continueFlg"`
	PersistRoundCount int32    `json:"persistRoundCount"`
}

// StatsFighterState captures an end-of-round snapshot for a fighter on one side.
type StatsFighterState struct {
	// Identity / selection
	Name       string  `json:"name"`       // character name
	ID         int32   `json:"id"`         // character ID
	MemberNo   int     `json:"memberNo"`   // team member index (0-based)
	SelectNo   int     `json:"selectNo"`   // select screen index
	AILevel    float32 `json:"aiLevel"`    // CPU level (0 = human)
	PalNo      int32   `json:"palNo"`      // pallete number
	RatioLevel int32   `json:"ratioLevel"` // ratio level

	// Health / quotes
	Life     int32 `json:"life"`     // life remaining at round end
	LifeMax  int32 `json:"lifeMax"`  // max life
	WinQuote int32 `json:"winQuote"` // -1 if unused

	// Outcome flags for this round
	Win        bool `json:"win"`        // this fighter's side won the round
	WinKO      bool `json:"winKO"`      // won by KO
	WinTime    bool `json:"winTime"`    // won on time-out
	WinPerfect bool `json:"winPerfect"` // perfect round
	WinSpecial bool `json:"winSpecial"` // won with a special
	WinHyper   bool `json:"winHyper"`   // won with a hyper
	DrawGame   bool `json:"drawGame"`   // round was declared a draw
	KO         bool `json:"ko"`         // this fighter was KO'd
	OverKO     bool `json:"overKO"`     // "over_ko"
}

// StatsRound stores all per-round stats.
// Indices: side 0 == P1, side 1 == P2.
type StatsRound struct {
	Index    int32                  `json:"index"`    // 1-based round number (1, 2, 3, ...)
	Timer    int32                  `json:"timer"`    // timer value for this round
	Score    [2]int32               `json:"score"`    // per-side score this round: [0]=P1, [1]=P2
	Fighters [2][]StatsFighterState `json:"fighters"` // [side][member]: end-of-round snapshots
}

// StatsMatch aggregates the entire fight between two sides and exposes
// convenient totals/tallies alongside the per-round breakdown.
// Indices: side 0 == P1, side 1 == P2.
type StatsMatch struct {
	// Timing / configuration
	MatchTime int32 `json:"matchTime"` // total match time in ticks
	RoundTime int32 `json:"roundTime"` // round time in ticks

	// Outcome & tallies
	WinSide    int      `json:"winSide"`    // 0 or 1 (which side won the match)
	LastRound  int32    `json:"lastRound"`  // index of the final round played (1-based)
	Draws      int32    `json:"draws"`      // number of drawn rounds
	Wins       [2]int32 `json:"wins"`       // wins per side across all rounds: [P1Wins, P2Wins]
	TeamModes  [2]int32 `json:"teamModes"`  // engine team mode per side (singles/simul/turns/tag)
	TotalScore [2]int32 `json:"totalScore"` // cumulative score per side at match end

	// Per-round breakdown
	Rounds []StatsRound `json:"rounds"` // one or more rounds; sorted by Index ascending
}

// StatsLog is a simple container for many matches.
type StatsLog struct {
	Matches []StatsMatch `json:"matches"`
}

// resets all gathered stats
func (s *StatsLog) reset() {
	s.Matches = nil
}

// startMatch begins a new match in the stats log.
// It captures the current round-time setting and team modes at match start.
func (s *StatsLog) startMatch() {
	m := StatsMatch{
		RoundTime: int32(sys.maxRoundTime),
		TeamModes: [2]int32{int32(sys.tmode[0]), int32(sys.tmode[1])},
	}
	s.Matches = append(s.Matches, m)
}

// addRound appends a round snapshot to the most recent match.
func (s *StatsLog) addRound(index int32, timer int32, score [2]int32, fighters [2][]StatsFighterState) {
	m := s.currentStatsMatch()
	if m == nil {
		s.startMatch()
		m = s.currentStatsMatch()
		if m == nil {
			return
		}
	}
	r := StatsRound{
		Index:    index,
		Timer:    timer, // may be 0 here; timers are filled in finalizeMatch
		Score:    score,
		Fighters: fighters,
	}
	m.Rounds = append(m.Rounds, r)
}

// finalizeMatch fills remaining fields (per-round timers, totals, winner).
func (s *StatsLog) finalizeMatch() {
	m := s.currentStatsMatch()
	if m == nil {
		return
	}

	// Attach per-round timers from the engine and compute total match time.
	var total int32
	for i, v := range sys.timerRounds {
		if i < len(m.Rounds) {
			m.Rounds[i].Timer = v
		}
		total += v
	}
	m.MatchTime = total

	// Copy outcome/tallies directly from engine state.
	m.WinSide = sys.winTeam
	// Last round played should be derived from what we recorded, not from sys.round math.
	if len(m.Rounds) > 0 {
		m.LastRound = m.Rounds[len(m.Rounds)-1].Index
	} else {
		m.LastRound = 0
	}
	m.Draws = sys.draws
	m.Wins = [2]int32{sys.wins[0], sys.wins[1]}
	m.TeamModes = [2]int32{int32(sys.tmode[0]), int32(sys.tmode[1])}
	m.RoundTime = int32(sys.maxRoundTime)

	// Compute total scores: scoreStart + sum(scoreRounds).
	sc0 := int32(sys.scoreStart[0])
	sc1 := int32(sys.scoreStart[1])
	for _, v := range sys.scoreRounds {
		sc0 += int32(v[0])
		sc1 += int32(v[1])
	}
	m.TotalScore = [2]int32{sc0, sc1}

	// Optionally: if round-level Score wasn't set earlier, backfill from sys.scoreRounds.
	/*if len(m.Rounds) == len(sys.scoreRounds) {
			for i := range m.Rounds {
				// Only overwrite if not populated (kept 0) during addRound.
				if m.Rounds[i].Score == ([2]int32{0, 0}) {
					m.Rounds[i].Score = [2]int32{int32(sys.scoreRounds[i][0]), int32(sys.scoreRounds[i][1])}
	 			}
			}
		}*/
}

// currentStatsMatch returns a pointer to the active (most recently started) match.
func (s *StatsLog) currentStatsMatch() *StatsMatch {
	if len(s.Matches) == 0 {
		return nil
	}
	return &s.Matches[len(s.Matches)-1]
}

// abortMatch removes the most recent match if it has no rounds (e.g., hard reset before first round).
func (s *StatsLog) abortMatch() {
	if len(s.Matches) == 0 {
		return
	}
	last := &s.Matches[len(s.Matches)-1]
	if len(last.Rounds) == 0 {
		s.Matches = s.Matches[:len(s.Matches)-1]
	}
}

// discardCurrentMatch removes the most recent match, even if it already contains rounds.
func (s *StatsLog) discardCurrentMatch() {
	if len(s.Matches) == 0 {
		return
	}
	s.Matches = s.Matches[:len(s.Matches)-1]
}

func (s *StatsLog) nextRound() {
	// Build per-side fighter snapshots for the stats system
	var fighters [2][]StatsFighterState
	for _, p := range sys.chars {
		if len(p) > 0 && p[0].teamside != -1 {
			side := int(p[0].teamside) // 0 or 1
			fs := StatsFighterState{
				Name:       p[0].name,
				ID:         p[0].id,
				MemberNo:   int(p[0].memberNo),
				SelectNo:   int(p[0].selectNo),
				AILevel:    p[0].getAILevel(),
				PalNo:      p[0].gi().palno,
				RatioLevel: p[0].ocd().ratioLevel,
				Life:       p[0].life,
				LifeMax:    p[0].lifeMax,
				WinQuote:   p[0].winquote,
				Win:        p[0].win(),
				WinKO:      p[0].winKO(),
				WinTime:    p[0].winTime(),
				WinPerfect: p[0].winPerfect(),
				WinSpecial: p[0].winType(WT_Special),
				WinHyper:   p[0].winType(WT_Hyper),
				DrawGame:   p[0].drawgame(),
				KO:         p[0].scf(SCF_ko),
				OverKO:     p[0].scf(SCF_over_ko),
			}
			fighters[side] = append(fighters[side], fs)
		}
	}

	// Record the round into the current stats match.
	// We fill timers later in finalizeMatch.
	roundIdx := sys.round
	roundScore := [2]int32{int32(sys.lifebar.sc[0].scorePoints), int32(sys.lifebar.sc[1].scorePoints)}
	s.addRound(roundIdx, 0, roundScore, fighters)
}

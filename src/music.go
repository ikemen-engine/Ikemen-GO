// ----------------------------------------------------------------------------
// Music system – overview
// ----------------------------------------------------------------------------
//
// Goal
// -----
// Unify every place that can set BGM (screenpack/motif, stage, storyboard,
// select.def, and Lua launch parameters) into one runtime type.
//
// Core type
// ---------
// Music is a map[string][]*bgMusic keyed by a *prefix* (e.g. "", "title",
// "round1", "life", "victory", "final", "scene1", ...). Each key holds a list
// of candidate tracks with playback options. Read() randomly picks one entry
// for the prefix and returns the resolved file and options. Play() delegates
// to Read() and opens the BGM.
//
// Where values come from (all normalized into Music struct)
// --------------------------------------------------
// • Motif / Screenpack (system.def [Music])
//   Parsed by iniutils.parseMusicSection() during loadMotif().
//   Stored in sys.motif.Music.
//
// • Stage (stage.def [Music])
//   Parsed by parseMusicSection() into the stage definition.
//   Stored in sys.stage.music.
//
// • Storyboard (per-scene bgm, declared under [Scene X])
//   Parsed by parseMusicSection() into prefixes like "scene1", ...
//   Stored in sys.storyboard.Scene[X].Music
//
// • select.def – character parameters (Lua addChar)
//   script.go passes raw "k=v" params via AppendParams(...).
//   Stored in the sys.sel.charlist[X].music
//
//
// • select.def – stage parameters (Lua addStage):
//   script.go passes raw "k=v" params via AppendParams(...).
//   Stored in the sys.sel.stagelist[X].music
//
// • launchFight (Lua loadStart):
//   script.go passes raw "k=v" params via AppendParams(...).
//   Stored in sys.sel.music
//
// Multiple candidates
// -------------------
// • In motif and stage [Music] and storyboard [Scene X] sections: multiple
//   candidates for the same prefix are declared as comma-separated lists on a
//   single key, e.g.:
//       title.bgm        = a.mp3, b.mp3, c.mp3
//       title.bgm.volume = 100,100,90
//   parseMusicSection() splits by comma and pairs values by index into
//   Music["title"].
//
// • In select.def / Lua params (AppendParams): each "<prefix>.bgmusic=" or
//   "<prefix>.music=" starts a new candidate for that prefix; subsequent
//   "<prefix>.bgm*" fields update the most recently created candidate.
//
// NOTE: select_params.go additionally supports "music=path vol loopstart loopend"
// by expanding it into the above "<prefix>.bgm*" keys before calling AppendParams.
//
// Normalization of naming
// -----------------------
// Elecbyte used inconsistent keys across files. Both the INI loader
// (parseMusicSection) and the runtime parameter path (AppendParams) accept
// synonyms and normalize them into bgMusic fields:
//
//   bgm / bgmusic / music
//   bgm.loop / bgmloop
//   bgm.volume / bgmvolume
//   bgm.loopstart / bgmloopstart
//   bgm.loopend / bgmloopend
//   bgm.startposition / bgmstartposition
//   bgm.freqmul / bgmfreqmul
//   bgm.loopcount / bgmloopcount
//
// Prefix handling
// ---------------
// Keys may be written as "<prefix>.<field>" (e.g. "round2.bgmusic").
// Everything before the first dot is the Music map key; the rest selects a
// bgMusic field. When no prefix is present the empty key "" is used.
//
// Resolution order (per character)
// --------------------------------------------
// The final playable list is computed and stored per character slot so that
// per-character or per-stage overrides can win over defaults; custom character
// music (or stage music coming from select.def) can override the stage's .def
// music, and launch-time Lua must be able to override both.

// In resetRoundState the code does (for each sys.chars index i):
// • 1. Clear
//   sys.cgi[i].music = make(Music)
// • 2. Base: stage.def [Music]
//   sys.cgi[i].music.Append(sys.stage.music)
// • 3. select.def stage params (addStage)
//   sys.cgi[i].music.Append(sys.stage.si().music)
// • 4. select.def character params (addChar) – override stage
//   sys.cgi[i].music.Override(p[0].si().music)
// • 5. launchFight params (loadStart) – last word
//   sys.cgi[i].music.Override(sys.sel.music)
//
// Runtime selection
// -----------------
// During a match, Music.act() reacts to state and tries to play a suitable
// prefix in this order:
// • round start: "final" (if decisive) or "round<sys.round>"
// • low life (team leader): "life"
// • victory (decisive, winner alive): "victory"
// tryPlay() checks if a prefix exists and has any defined track, then calls
// Play. Read() randomizes among multiple entries for the prefix and resolves
// the path using SearchFile.
//
// ----------------------------------------------------------------------------

package main

import (
	"fmt"
	"sort"
	"strings"
)

type MusicSource int32

const (
	MS_Match        MusicSource = iota // final, per-character, merged list (sys.cgi[i].music)
	MS_StageDef                        // stage.def [Music]
	MS_CharParams                      // select.def character params (Lua addChar)
	MS_StageParams                     // select.def stage params (Lua addStage)
	MS_LaunchParams                    // launchFight params (Lua loadStart)
	MS_Motif                           // system.def [Music] in the motif/screenpack
)

type bgMusic struct {
	bgmusic          string
	bgmloop          int32
	bgmvolume        int32
	bgmloopstart     int32
	bgmloopend       int32
	bgmstartposition int32
	bgmfreqmul       float32
	bgmloopcount     int32
}

func newBgMusic() *bgMusic {
	return &bgMusic{bgmloop: 1, bgmvolume: 100, bgmfreqmul: 1, bgmloopcount: -1}
}

// Music is the normalized store for all music data.
// Key: prefix (e.g. "round1", "life", "victory", "title", etc.)
// Value: ordered slice of candidates; Read() picks one at random.
type Music map[string][]*bgMusic

// HasPrefix reports whether a music prefix exists.
func (m Music) HasPrefix(prefix string) bool {
	nkey := normalizeMusicPrefix(prefix)
	lst, ok := m[nkey]
	return ok && len(lst) > 0
}

// Append merges another Music by concatenating candidate lists per prefix.
// Use when adding sources of equal or lower priority (e.g. stage.def base
// then select.def stage params).
func (m Music) Append(other Music) {
	//fmt.Printf("[music] Append: merging %d prefix(es) into %d existing\n", len(other), len(m))
	for key, otherList := range other {
		m[key] = append(m[key], otherList...)
	}
}

// Override applies element-wise replacement per prefix (by index). If the
// overriding list is longer, it extends the target. Use when raising
// priority (e.g. char select params over stage lists, or launchFight over all).
func (m Music) Override(other Music) {
	//fmt.Printf("[music] Override: applying %d prefix(es) onto %d existing\n", len(other), len(m))
	for key, otherList := range other {
		//fmt.Printf("[music] Override: prefix '%s' (%d candidate(s))\n", key, len(otherList))
		if mList, exists := m[key]; exists {
			for i, otherBg := range otherList {
				if i < len(mList) {
					mList[i] = otherBg
				} else {
					mList = append(mList, otherBg)
				}
			}
			m[key] = mList
		} else {
			m[key] = otherList
		}
	}
}

// flattens dotted prefixes to underscore form, matching parseMusicSection.
func normalizeMusicPrefix(prefix string) string {
	return strings.ReplaceAll(prefix, ".", "_")
}

// AppendParams parses comma-separated "key=value" pairs (as passed from
// Lua addChar/addStage/loadStart) and appends to the proper prefix list.
func (m Music) AppendParams(entries []string) {
	for _, c := range entries {
		//fmt.Printf("[music] AppendParams: raw entry '%s'\n", c)
		if eqPos := strings.Index(c, "="); eqPos != -1 {
			key := strings.TrimSpace(c[:eqPos])
			value := strings.TrimSpace(c[eqPos+1:])
			prefix := ""
			field := key

			// Allow dots in prefix: split using the last ".<music anchor>" if present,
			// otherwise fall back to last dot.
			kl := strings.ToLower(key)
			anchors := []string{".bgmusic", ".music", ".bgm"}
			best := -1
			for _, a := range anchors {
				if i := strings.LastIndex(kl, a); i > best {
					best = i
				}
			}
			if best >= 0 {
				prefix = strings.TrimSpace(key[:best])
				field = strings.TrimSpace(key[best+1:]) // without leading dot
			} else if dotPos := strings.LastIndex(key, "."); dotPos != -1 {
				prefix = key[:dotPos]
				field = key[dotPos+1:]
			}

			// Flatten dotted prefixes to match storage in parseMusicSection.
			prefix = normalizeMusicPrefix(prefix)
			//fmt.Printf("[music] AppendParams: normalized key='%s' -> prefix='%s', field='%s', value='%s'\n", key, prefix, field, value)

			// Ignore non-music fields
			if !strings.HasPrefix(field, "bgm") && field != "music" {
				//fmt.Printf("[music] AppendParams: skipping non-music field '%s'\n", field)
				continue
			}

			// Ensure there is a current bgMusic for this prefix
			if len(m[prefix]) == 0 || field == "bgmusic" || field == "music" {
				m[prefix] = append(m[prefix], newBgMusic())
				//fmt.Printf("[music] AppendParams: created new bgMusic entry for prefix '%s' (total now %d)\n", prefix, len(m[prefix]))
			}
			idx := len(m[prefix]) - 1
			switch field {
			case "bgmusic", "music":
				m[prefix][idx].bgmusic = value
				//fmt.Printf("[music] AppendParams: set bgmusic for prefix '%s' idx=%d -> '%s'\n", prefix, idx, value)
			case "bgmloop":
				m[prefix][idx].bgmloop = Atoi(value)
			case "bgmvolume":
				m[prefix][idx].bgmvolume = Atoi(value)
			case "bgmloopstart":
				m[prefix][idx].bgmloopstart = Atoi(value)
			case "bgmloopend":
				m[prefix][idx].bgmloopend = Atoi(value)
			case "bgmstartposition":
				m[prefix][idx].bgmstartposition = Atoi(value)
			case "bgmfreqmul":
				m[prefix][idx].bgmfreqmul = float32(Atof(value))
			case "bgmloopcount":
				m[prefix][idx].bgmloopcount = Atoi(value)
			}
		}
	}
}

// DebugDump prints a human-readable dump of the Music contents.
func (m Music) DebugDump(label string) {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	//fmt.Printf("[music] DebugDump: %s -> %d prefix(es)\n", label, len(keys))
	for _, prefix := range keys {
		list := m[prefix]
		//fmt.Printf("[music]   prefix '%s': %d track(s)\n", prefix, len(list))
		for _, bg := range list {
			if bg == nil {
				//fmt.Printf("[music]     [%d] <nil>\n", i)
				continue
			}
			//fmt.Printf("[music]     [%d] bgmusic='%s' loop=%d vol=%d loopstart=%d loopend=%d startpos=%d freqmul=%g loopcount=%d\n", i, bg.bgmusic, bg.bgmloop, bg.bgmvolume, bg.bgmloopstart, bg.bgmloopend, bg.bgmstartposition, bg.bgmfreqmul, bg.bgmloopcount)
		}
	}
}

// Read resolves a concrete file and playback params for a key like
// "round1.bgmusic" by using its prefix ("round1") and picking one random
// candidate from the list.
func (m Music) Read(key, def string) (string, int, int, int, int, int, float32, int) {
	var bgm string
	var loop, volume, loopstart, loopend, startposition, loopcount int = 1, 100, 0, 0, 0, -1
	var freqmul float32 = 1.0
	//fmt.Printf("[music] Read: key='%s' def='%s'\n", key, def)
	// Support dotted prefixes by only stripping a suffix when the key actually targets a music field.
	prefix := key
	kl := strings.ToLower(key)
	anchors := []string{".bgmusic", ".music", ".bgm"}
	best := -1
	for _, a := range anchors {
		if i := strings.LastIndex(kl, a); i > best {
			best = i
		}
	}
	if best >= 0 {
		prefix = key[:best]
	}
	prefix = normalizeMusicPrefix(prefix)
	if len(m[prefix]) > 0 {
		idx := int(RandI(0, int32(len(m[prefix]))-1))
		bgm = SearchFile(m[prefix][idx].bgmusic, []string{def, "", "sound/"})
		//fmt.Printf("[music] Read: prefix='%s' chose idx=%d -> '%s'\n", prefix, idx, bgm)
		loop = int(m[prefix][idx].bgmloop)
		volume = int(m[prefix][idx].bgmvolume)
		loopstart = int(m[prefix][idx].bgmloopstart)
		loopend = int(m[prefix][idx].bgmloopend)
		startposition = int(m[prefix][idx].bgmstartposition)
		freqmul = m[prefix][idx].bgmfreqmul
		loopcount = int(m[prefix][idx].bgmloopcount)
	}
	return bgm, loop, volume, loopstart, loopend, startposition, freqmul, loopcount
}

// Play opens the chosen track in the global BGM player.
func (m Music) Play(key, path string) bool {
	track, loop, volume, loopstart, loopend, startposition, freqmul, loopcount := m.Read(key, path)
	//fmt.Printf("[music] Play: key='%s' def='%s' -> track='%s'\n", key, path, track)

	if track != "" && track != sys.bgm.filename {
		//fmt.Printf("[music] Play: opening track='%s' loop=%d vol=%d loopstart=%d loopend=%d startpos=%d freqmul=%g loopcount=%d\n", track, loop, volume, loopstart, loopend, startposition, freqmul, loopcount)
		sys.bgm.Open(track, loop, volume, loopstart, loopend, startposition, freqmul, loopcount)
		sys.playBgmFlg = sys.playBgmFlg || !sys.sel.gameParams.PersistMusic
		return true
	}
	if track == "" {
		//fmt.Printf("[music] Play: no track resolved for key='%s'\n", key)
	} else if track == sys.bgm.filename {
		//fmt.Printf("[music] Play: track '%s' already playing, skipping\n", track)
	}
	return false
}

type BGMState byte

const (
	BGMStateIdle    BGMState = iota // no track chosen yet
	BGMStateRound                   // round start track chosen/playing
	BGMStateLowLife                 // low-life track chosen/playing
	BGMStateVictory                 // victory track chosen/playing
)

// tryPlay is a guarded helper used by act(): if a prefix exists and has at
// least one non-empty entry, it delegates to Play("<prefix>.bgmusic", def).
func (m Music) tryPlay(key, def string) bool {
	//if dot := strings.Index(key, "."); dot != -1 {
	//	key = key[:dot]
	//}
	nkey := normalizeMusicPrefix(key)
	lst, ok := m[nkey]
	if !ok || len(lst) == 0 {
		//fmt.Printf("[music] tryPlay: prefix '%s' not found or empty\n", key)
		return false
	}
	hasDefined := false
	for _, v := range lst {
		if v != nil && strings.TrimSpace(v.bgmusic) != "" {
			hasDefined = true
			break
		}
		//fmt.Printf("[music] tryPlay: prefix '%s' candidate[%d] has empty bgmusic\n", key, i)
	}
	if !hasDefined {
		//fmt.Printf("[music] tryPlay: prefix '%s' has no defined bgmusic entries\n", key)
		return false
	}
	ok = m.Play(nkey+".bgmusic", def)
	//fmt.Printf("[music] tryPlay: Play('%s.bgmusic') -> %v\n", key, ok)
	return ok
}

// act drives in-fight music state transitions:
//   - At round start: final.bgmusic (if decisive) else round{N}.bgmusic
//   - On leader low life: life.bgmusic
//   - On decisive victory: victory.bgmusic
func (m Music) act() {
	if sys.gameMode == "demo" && !sys.motif.DemoMode.Fight.PlayBgm {
		return
	}
	//fmt.Printf("[music] act: tickCount=%d round=%d match=%d bgmState=%d\n", sys.tickCount, sys.round, sys.match, sys.stage.bgmState)
	if sys.tickCount == 0 && sys.round == 1 && !sys.roundResetMatchStart &&
		(sys.match == 1 || !sys.sel.gameParams.PersistMusic || sys.stage.bgmState != BGMStateRound) {
		sys.bgm.Stop()
		sys.stage.bgmState = BGMStateIdle
	}
	// Iterate players in order: P2, P2 teammates, then P1, P1 teammates.
	// Skips empty slots and ignores attached chars.
	for side := 1; side >= 0; side-- { // 1 = P2 side first, then 0 = P1 side
		for pn := side; pn < int(MaxSimul)*2; pn += 2 {
			if len(sys.chars[pn]) == 0 || sys.chars[pn][0] == nil {
				continue
			}
			c := sys.chars[pn][0] // root player in this slot
			cmusic := sys.cgi[pn].music

			// Round Start
			if c.teamside == sys.home &&
				c.playerNo == c.teamLeader()-1 &&
				(sys.stage.bgmState == BGMStateIdle || sys.tickCount == 0) &&
				sys.tickCount == 0 {
				switch {
				case sys.roundIsFinal() && cmusic.tryPlay("final", sys.stage.def):
				case sys.sel.gameParams.PersistMusic && cmusic.tryPlay(fmt.Sprintf("round%d", sys.match), sys.stage.def):
				case cmusic.tryPlay(fmt.Sprintf("round%d", sys.round), sys.stage.def):
				case cmusic.tryPlay("", sys.stage.def):
				}
				sys.stage.bgmState = BGMStateRound
				continue
			}

			// Low Life (only team leader)
			if sys.stage.bgmState == BGMStateRound &&
				sys.roundState() == 2 &&
				c.playerNo == c.teamLeader()-1 &&
				float32(c.life)/float32(c.lifeMax) <= sys.stage.bgmratio {
				//fmt.Printf("[music] act: low life detected for player %d, trying 'life' prefix\n", c.playerNo)
				if cmusic.tryPlay("life", sys.stage.def) {
					sys.stage.bgmState = BGMStateLowLife
					continue
				}
			}

			// Victory (decisive round, winning & alive)
			if sys.stage.bgmState < BGMStateVictory &&
				c.win() && c.alive() &&
				sys.decisiveRound[c.playerNo&1] {

				//fmt.Printf("[music] act: decisive victory for teamside=%d, trying 'victory' prefix\n", c.teamside)

				if cmusic.tryPlay("victory", sys.stage.def) {
					sys.stage.bgmState = BGMStateVictory
					continue
				}
			}
		}
	}
}

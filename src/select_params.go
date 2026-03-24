package main

import (
	"fmt"
	"strings"
)

// -----------------------------------------------------------------------------
// Select.def / Lua param storage (char, stage, launchFight/loadStart)
// -----------------------------------------------------------------------------

// override key format (loadStart params)
func parseOverrideKey(kl string) (team int, member int, field string, ok bool) {
	parts := strings.Split(kl, ".")
	if len(parts) != 3 {
		return 0, 0, "", false
	}
	switch parts[0] {
	case "p1":
		team = 0
	case "p2":
		team = 1
	default:
		return 0, 0, "", false
	}
	m := int(Atoi(parts[1]))
	if m <= 0 {
		return 0, 0, "", false
	}
	return team, m - 1, parts[2], true
}

func parseKV(s string) (key, value string, ok bool) {
	eq := strings.Index(s, "=")
	if eq < 0 {
		return "", "", false
	}
	key = strings.TrimSpace(s[:eq])
	value = strings.TrimSpace(s[eq+1:])
	if key == "" {
		return "", "", false
	}
	return key, value, true
}

func parseBoolLoose(s string) (bool, bool) {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "1", "true", "yes", "on":
		return true, true
	case "0", "false", "no", "off":
		return false, true
	default:
		return false, false
	}
}

func isMusicKey(key string) bool {
	kl := strings.ToLower(strings.TrimSpace(key))
	if strings.HasPrefix(kl, "charparam") {
		return false
	}
	return strings.HasPrefix(kl, "music") ||
		strings.Contains(kl, ".music") ||
		strings.Contains(kl, ".bgm") ||
		strings.HasPrefix(kl, "bgm")
}

func extractPrefixFromMusicKey(key string) (prefix string) {
	// Mirror Music.AppendParams split logic enough to synthesize bgmvolume keys.
	kl := strings.ToLower(key)
	anchors := []string{".bgmusic", ".music", ".bgm"}
	best := -1
	for _, a := range anchors {
		if i := strings.LastIndex(kl, a); i > best {
			best = i
		}
	}
	if best >= 0 {
		return strings.TrimSpace(key[:best])
	}
	if dot := strings.LastIndex(key, "."); dot >= 0 {
		return strings.TrimSpace(key[:dot])
	}
	return ""
}

func splitMusicParamValue(v string) (path string, extras []string) {
	v = strings.TrimSpace(v)
	if v == "" {
		return "", nil
	}

	lower := strings.ToLower(v)
	exts := []string{
		".ogg", ".mp3", ".wav", ".flac",
		".mid", ".midi",
		".xm", ".mod", ".it", ".s3m",
	}

	bestEnd := -1
	for _, ext := range exts {
		if i := strings.LastIndex(lower, ext); i >= 0 {
			end := i + len(ext)
			if end > bestEnd {
				bestEnd = end
			}
		}
	}

	// If no supported extension is found, we won't be able to play it anyway,
	// so just treat the whole value as the path (no extras).
	if bestEnd <= 0 {
		return v, nil
	}

	path = strings.TrimSpace(v[:bestEnd])
	rest := strings.TrimSpace(v[bestEnd:])
	if rest != "" {
		extras = strings.Fields(rest)
	}
	return path, extras
}

// expandMusicKV turns: music=path 80 100 200 into:
// music=path, bgmvolume=80, bgmloopstart=100, bgmloopend=200
func expandMusicKV(key, value string) []string {
	key = strings.TrimSpace(key)
	value = strings.TrimSpace(value)
	if key == "" {
		return nil
	}

	path, extras := splitMusicParamValue(value)
	if strings.TrimSpace(path) == "" {
		return []string{key + "="}
	}

	out := []string{key + "=" + path}
	if len(extras) == 0 {
		return out
	}

	prefix := extractPrefixFromMusicKey(key)
	pfx := ""
	if prefix != "" {
		pfx = prefix + "."
	}

	// volume, loopstart, loopend (optional)
	if len(extras) >= 1 {
		out = append(out, fmt.Sprintf("%sbgmvolume=%s", pfx, extras[0]))
	}
	if len(extras) >= 2 {
		out = append(out, fmt.Sprintf("%sbgmloopstart=%s", pfx, extras[1]))
	}
	if len(extras) >= 3 {
		out = append(out, fmt.Sprintf("%sbgmloopend=%s", pfx, extras[2]))
	}
	return out
}

// -----------------------------------------------------------------------------
// Character params (select.def [Characters])
// -----------------------------------------------------------------------------

type SelectCharParams struct {
	musicEntries  []string `ini:"musicEntries"`
	AI            int32    `ini:"ai"`
	VsScreen      bool     `ini:"vsscreen"`
	VictoryScreen bool     `ini:"victoryscreen"`
	Rounds        int32    `ini:"rounds"`
	Time          int32    `ini:"time"`
	Single        bool     `ini:"single"`
	IncludeStage  int32    `ini:"includestage"`
	Bonus         bool     `ini:"bonus"`
	Exclude       bool     `ini:"exclude"`
	Hidden        int32    `ini:"hidden"`
	Order         int32    `ini:"order"`
	OrderSurvival int32    `ini:"ordersurvival"`
	ArcadePath    string   `ini:"arcadepath"`
	RatioPath     string   `ini:"ratiopath"`
	Unlock        string   `ini:"unlock"`
	SlotSelect    string   `ini:"slotselect"`
	SlotNext      string   `ini:"slotnext"`
	SlotPrevious  string   `ini:"slotprevious"`
	Raw           []string `ini:"raw"`
}

func newSelectCharParams() *SelectCharParams {
	return &SelectCharParams{
		musicEntries:  make([]string, 0),
		Raw:           make([]string, 0),
		AI:            -1,
		VsScreen:      true,
		VictoryScreen: true,
		Rounds:        -1,
		Time:          -1,
		IncludeStage:  1,
		Hidden:        0,
		Order:         -1,
		OrderSurvival: -1,
	}
}

func (p *SelectCharParams) MusicEntries() []string { return p.musicEntries }

func (p *SelectCharParams) AppendParams(entries []string) {
	p.Raw = append(p.Raw, entries...)
	for _, e := range entries {
		key, val, ok := parseKV(e)
		if !ok {
			continue
		}
		kl := strings.ToLower(strings.TrimSpace(key))

		if isMusicKey(key) {
			p.musicEntries = append(p.musicEntries, expandMusicKV(key, val)...)
			continue
		}

		switch kl {
		case "ai":
			p.AI = Atoi(val)
		case "vsscreen":
			if b, ok := parseBoolLoose(val); ok {
				p.VsScreen = b
			}
		case "victoryscreen":
			if b, ok := parseBoolLoose(val); ok {
				p.VictoryScreen = b
			}
		case "rounds":
			p.Rounds = Atoi(val)
		case "time":
			p.Time = Atoi(val)
		case "single":
			if b, ok := parseBoolLoose(val); ok {
				p.Single = b
			}
		case "includestage":
			p.IncludeStage = Atoi(val)
		case "bonus":
			if b, ok := parseBoolLoose(val); ok {
				p.Bonus = b
			}
		case "exclude":
			if b, ok := parseBoolLoose(val); ok {
				p.Exclude = b
			}
		case "hidden":
			p.Hidden = Atoi(val)
		case "order":
			p.Order = Atoi(val)
		case "ordersurvival":
			p.OrderSurvival = Atoi(val)
		case "arcadepath":
			p.ArcadePath = val
		case "ratiopath":
			p.RatioPath = val
		case "unlock":
			p.Unlock = val
		case "select":
			p.SlotSelect = val
		case "next":
			p.SlotNext = val
		case "previous":
			p.SlotPrevious = val
		}
	}
}

// -----------------------------------------------------------------------------
// Stage params (select.def [ExtraStages])
// -----------------------------------------------------------------------------

type SelectStageParams struct {
	musicEntries []string `ini:"musicentries"`
	Order        []int32  `ini:"order"`
	Unlock       string   `ini:"unlock"`
	Raw          []string `ini:"raw"`
}

func newSelectStageParams() *SelectStageParams {
	return &SelectStageParams{
		musicEntries: make([]string, 0),
		Order:        make([]int32, 0),
		Raw:          make([]string, 0),
	}
}

func (p *SelectStageParams) MusicEntries() []string { return p.musicEntries }

func (p *SelectStageParams) AppendParams(entries []string) {
	p.Raw = append(p.Raw, entries...)
	for _, e := range entries {
		key, val, ok := parseKV(e)
		if !ok {
			continue
		}
		kl := strings.ToLower(strings.TrimSpace(key))

		if isMusicKey(key) {
			p.musicEntries = append(p.musicEntries, expandMusicKV(key, val)...)
			continue
		}

		switch kl {
		case "order":
			p.Order = append(p.Order, Atoi(val))
		case "unlock":
			p.Unlock = val
		}
	}
}

// -----------------------------------------------------------------------------
// LaunchFight/loadStart params (Lua loadStart string path)
// -----------------------------------------------------------------------------

type OverrideCharData struct {
	life        int32
	lifeMax     int32
	power       int32
	dizzyPoints int32
	guardPoints int32
	ratioLevel  int32
	lifeRatio   float32
	attackRatio float32
	existed     bool
}

func newOverrideCharData() *OverrideCharData {
	return &OverrideCharData{life: -1, lifeMax: -1, power: -1, dizzyPoints: -1,
		guardPoints: -1, ratioLevel: 0, lifeRatio: 1, attackRatio: 1}
}

type GameParams struct {
	musicEntries        []string `ini:"musicentries"`
	Continue            bool     `ini:"continue"`
	QuickContinue       bool     `ini:"quickcontinue"`
	Order               int32    `ini:"order"`
	Stage               string   `ini:"stage"`
	AI                  float32  `ini:"ai"`
	Time                int32    `ini:"time"`
	VsScreen            bool     `ini:"vsscreen"`
	VictoryScreen       bool     `ini:"victoryscreen"`
	WinScreen           bool     `ini:"winscreen"`
	LuaCode             string   `ini:"luacode"`
	PersistLife         bool     `ini:"persistlife"`
	PersistMusic        bool     `ini:"persistmusic"`
	PersistRounds       bool     `ini:"persistrounds"`
	CharParamAI         bool     `ini:"charparam.ai"`
	CharParamArcadePath bool     `ini:"charparam.arcadepath"`
	CharParamMusic      bool     `ini:"charparam.music"`
	CharParamRounds     bool     `ini:"charparam.rounds"`
	CharParamSingle     bool     `ini:"charparam.single"`
	CharParamStage      bool     `ini:"charparam.stage"`
	CharParamTime       bool     `ini:"charparam.time"`
	TurnsOffset         [2]int32
	ocd                 [3][]OverrideCharData
	Raw                 []string
}

func newGameParams() *GameParams {
	return &GameParams{
		musicEntries:  make([]string, 0),
		Raw:           make([]string, 0),
		Continue:      true,
		Order:         -1,
		Time:          -1,
		AI:            -1,
		VsScreen:      true,
		VictoryScreen: true,
		WinScreen:     true,
		ocd: [3][]OverrideCharData{
			make([]OverrideCharData, 0),
			make([]OverrideCharData, 0),
			make([]OverrideCharData, 0),
		},
	}
}

// Seeds match-scoped defaults from the currently active motif runtime flags.
func newGameParamsFromMotif(m *Motif) *GameParams {
	p := newGameParams()
	if m == nil {
		return p
	}
	p.Continue = m.co.enabled
	p.VictoryScreen = m.vi.enabled
	p.WinScreen = m.wi.winEnabled
	return p
}

func (p *GameParams) ResetFromMotif(m *Motif) {
	*p = *newGameParamsFromMotif(m)
}

func (p *GameParams) Reset() {
	*p = *newGameParams()
}

func (p *GameParams) MusicEntries() []string {
	return p.musicEntries
}

func (p *GameParams) ensureOverride(team, member int) *OverrideCharData {
	if team < 0 || team >= len(p.ocd) || member < 0 {
		return newOverrideCharData()
	}
	for len(p.ocd[team]) <= member {
		p.ocd[team] = append(p.ocd[team], *newOverrideCharData())
	}
	return &p.ocd[team][member]
}

func (p *GameParams) AppendParams(entries []string) {
	p.Raw = append(p.Raw, entries...)
	for _, e := range entries {
		key, val, ok := parseKV(e)
		if !ok {
			continue
		}
		kl := strings.ToLower(strings.TrimSpace(key))

		if isMusicKey(key) {
			p.musicEntries = append(p.musicEntries, expandMusicKV(key, val)...)
			continue
		}

		// per-character overrides: p1.<member>.<field>=... / p2.<member>.<field>=...
		if team, member, field, ok := parseOverrideKey(kl); ok {
			ocd := p.ensureOverride(team, member)
			switch field {
			case "life":
				ocd.life = int32(Atoi(val))
			case "lifemax":
				ocd.lifeMax = int32(Atoi(val))
			case "power":
				ocd.power = int32(Atoi(val))
			case "dizzypoints":
				ocd.dizzyPoints = int32(Atoi(val))
			case "guardpoints":
				ocd.guardPoints = int32(Atoi(val))
			case "ratiolevel":
				ocd.ratioLevel = int32(Atoi(val))
			case "liferatio":
				ocd.lifeRatio = float32(Atof(val))
			case "attackratio":
				ocd.attackRatio = float32(Atof(val))
			case "existed":
				ocd.existed, _ = parseBoolLoose(val)
			}
			continue
		}

		switch kl {
		case "p1.turnsoffset":
			v := int32(Atoi(val))
			if v < 0 {
				v = 0
			}
			p.TurnsOffset[0] = v
			continue
		case "p2.turnsoffset":
			v := int32(Atoi(val))
			if v < 0 {
				v = 0
			}
			p.TurnsOffset[1] = v
			continue
		case "charparam.ai":
			if b, ok := parseBoolLoose(val); ok {
				p.CharParamAI = b
			}
		case "charparam.arcadepath":
			if b, ok := parseBoolLoose(val); ok {
				p.CharParamArcadePath = b
			}
		case "charparam.music":
			if b, ok := parseBoolLoose(val); ok {
				p.CharParamMusic = b
			}
		case "charparam.rounds":
			if b, ok := parseBoolLoose(val); ok {
				p.CharParamRounds = b
			}
		case "charparam.single":
			if b, ok := parseBoolLoose(val); ok {
				p.CharParamSingle = b
			}
		case "charparam.stage":
			if b, ok := parseBoolLoose(val); ok {
				p.CharParamStage = b
			}
		case "charparam.time":
			if b, ok := parseBoolLoose(val); ok {
				p.CharParamTime = b
			}
		case "continue":
			if b, ok := parseBoolLoose(val); ok {
				p.Continue = b
			}
		case "quickcontinue":
			if b, ok := parseBoolLoose(val); ok {
				p.QuickContinue = b
			}
		case "order":
			p.Order = Atoi(val)
		case "stage":
			p.Stage = val
		case "ai":
			p.AI = float32(Atof(val))
		case "time":
			p.Time = Atoi(val)
		case "vsscreen":
			if b, ok := parseBoolLoose(val); ok {
				p.VsScreen = b
			}
		case "victoryscreen":
			if b, ok := parseBoolLoose(val); ok {
				p.VictoryScreen = b
			}
		case "winscreen":
			if b, ok := parseBoolLoose(val); ok {
				p.WinScreen = b
			}
		case "lua":
			p.LuaCode = val
		case "persistlife":
			if b, ok := parseBoolLoose(val); ok {
				p.PersistLife = b
			}
		case "persistmusic":
			if b, ok := parseBoolLoose(val); ok {
				p.PersistMusic = b
			}
		case "persistrounds":
			if b, ok := parseBoolLoose(val); ok {
				p.PersistRounds = b
			}
		}
	}
}

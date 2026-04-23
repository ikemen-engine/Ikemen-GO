package main

import (
	_ "embed" // Support for go:embed resources
	"fmt"
	"reflect"
	"regexp"
	"sort"
	"strings"
	"unicode/utf8"

	"gopkg.in/ini.v1"
)

//go:embed resources/defaultStoryboard.ini
var defaultStoryboard []byte

type LayerProperties struct {
	Anim           int32      `ini:"anim" default:"-1"`
	Spr            [2]int32   `ini:"spr" default:"-1,0"`
	Offset         [2]float32 `ini:"offset"`
	Facing         int32      `ini:"facing" default:"1"`
	Scale          [2]float32 `ini:"scale" default:"1,1"`
	Xshear         float32    `ini:"xshear"`
	Angle          float32    `ini:"angle"`
	XAngle         float32    `ini:"xangle"`
	YAngle         float32    `ini:"yangle"`
	Projection     string     `ini:"projection"`
	Focallength    float32    `ini:"focallength" default:"2048"`
	Layerno        int16      `ini:"layerno" default:"2"`
	Window         [4]int32   `ini:"window"`
	Localcoord     [2]int32   `ini:"localcoord"`
	Accel          [2]float32 `ini:"accel"`
	Velocity       [2]float32 `ini:"velocity"`
	Friction       [2]float32 `ini:"friction" default:"1,1"`
	AnimData       *Anim
	Font           [8]int32 `ini:"font" default:"-1,0,0,255,255,255,255,-1"`
	Text           string   `ini:"text"`
	TextSpriteData *TextSprite
	PalFx          PalFxProperties `ini:"palfx"`
	TextSpacing    [2]float32      `ini:"textspacing"`
	TextDelay      float32         `ini:"textdelay" default:"2"`
	TextWrap       string          `ini:"textwrap" default:"w"`
	TextWindow     [4]int32        `ini:"textwindow"`
	StartTime      int32           `ini:"starttime"`
	EndTime        int32           `ini:"endtime"`
	// runtime-only typing state (not loaded from INI)
	typedLen          int
	charDelayCounter  int32
	lineFullyRendered bool
}

type SoundProperties struct {
	Value         [2]int32 `ini:"value" default:"-1,0"`
	StartTime     int32    `ini:"starttime"`
	VolumeScale   int32    `ini:"volumescale" default:"100"`
	Pan           float32  `ini:"pan"`
	LoopStart     int      `ini:"loopstart"`
	LoopEnd       int      `ini:"loopend"`
	StartPosition int      `ini:"startposition"`
}

type SceneProperties struct {
	End struct {
		Time int32 `ini:"time"`
	} `ini:"end"`
	FadeIn       FadeProperties `ini:"fadein"`
	FadeOut      FadeProperties `ini:"fadeout"`
	ClearColor   [3]int32       `ini:"clearcolor" default:"-1,0,0"`
	ClearAlpha   [2]int32       `ini:"clearalpha" default:"255,0"`
	ClearLayerno int16          `ini:"clearlayerno" default:"0"`
	RectData     *Rect
	Layerall     struct {
		Pos [2]float32 `ini:"pos"`
	} `ini:"layerall"`
	Layer map[string]*LayerProperties `ini:"map:^(?i)layer_?[0-9]+$" lua:"layer"`
	Sound map[string]*SoundProperties `ini:"map:^(?i)sound_?[0-9]+$" lua:"sound"`
	Bg    struct {
		BGDef *BGDef
		Name  string `ini:"name"`
		//BgClearColor   [3]int32 `ini:"bgclearcolor" default:"0,0,0"`
		//BgClearAlpha   [2]int32 `ini:"bgclearalpha" default:"255,0"`
		//BgClearLayerno int16    `ini:"bgclearlayerno" default:"0"`
		//RectData       *Rect
	} `ini:"bg"`
	Bgm   BgmProperties `ini:"bgm"`
	Music Music
	Jump  *int `ini:"jump"`
}

type Storyboard struct {
	IniFile   *ini.File
	AnimTable AnimationTable
	Sff       *Sff
	Snd       *Snd
	Fnt       map[int]*Fnt
	Model     *Model
	Def       string         `ini:"def"`
	Info      InfoProperties `ini:"info"`
	SceneDef  struct {
		Spr        string                     `ini:"spr" lookup:"def,,data/"`
		Snd        string                     `ini:"snd" lookup:"def,,data/"`
		Font       map[string]*FontProperties `ini:"map:^(?i)font[0-9]+$" lua:"font"`
		Model      string                     `ini:"model" lookup:"def,,data/"`
		StartScene int                        `ini:"startscene"`
		Key        struct {
			Skip   []string `ini:"skip"`
			Cancel []string `ini:"cancel"`
		} `ini:"key"`
		StopMusic     bool `ini:"stopmusic"`
		DisableCancel bool `ini:"disablecancel"`
	} `ini:"scenedef"`
	Scene         map[string]*SceneProperties `ini:"map:^(?i)scene_?[0-9]+$" lua:"scene"`
	fntIndexByKey map[string]int
	//enabled		   bool
	active            bool
	initialized       bool
	counter           int32
	endTimer          int32
	sceneKeys         []string
	currentSceneIndex int
	cancel            bool
	musicPlaying      bool
	fadePolicy        FadeStartPolicy
	dialogueLayers    []*LayerProperties // ordered text layers (StartTime asc, then layer index)
	dialoguePos       int                // current layer index into dialogueLayers
}

// loadStoryboard loads and parses the INI file into a struct.
func loadStoryboard(def string) (*Storyboard, error) {
	// Define load options if needed
	// https://github.com/go-ini/ini/blob/main/ini.go
	baseOptions := ini.LoadOptions{
		Insensitive:             false,
		InsensitiveSections:     true,
		InsensitiveKeys:         false,
		IgnoreInlineComment:     false,
		SkipUnrecognizableLines: true,
		//AllowBooleanKeys: true,
		AllowShadows: false,
		//AllowNestedValues: true,
		UnparseableSections:        []string{},
		AllowPythonMultilineValues: false,
		PreserveSurroundedQuote:    false,
		//KeyValueDelimiters: "=:",
		//KeyValueDelimiterOnWrite: "=",
		//ChildSectionDelimiter: ".",
		//AllowNonUniqueSections: true,
		//AllowDuplicateShadowValues: true,
	}

	// Preserve duplicates in user config so we can apply "first instance wins".
	userOptions := baseOptions
	userOptions.AllowShadows = true

	// Start merged INI as defaults, then overlay user (first-wins for duplicates).
	iniFile, err := ini.LoadSources(baseOptions, defaultStoryboard)
	if err != nil {
		return nil, fmt.Errorf("Failed to load INI file: %w", err)
	}
	defaultOnlyIni, err := ini.LoadSources(baseOptions, defaultStoryboard)
	if err != nil {
		return nil, fmt.Errorf("Failed to load defaults-only INI: %w", err)
	}
	userIniFile, err := ini.LoadSources(userOptions, def)
	if err != nil {
		return nil, fmt.Errorf("Failed to load user INI %s: %w", def, err)
	}
	overlayUserFirstWins(iniFile, userIniFile)

	var s Storyboard
	s.Def = def
	s.initStruct()

	assignFrom := func(src *ini.File) error {
		if src == nil {
			return nil
		}
		// Group base vs. language-specific sections (preserving file order within each group).
		type secPair struct {
			sec  *ini.Section
			name string // logical name without the "xx." prefix
		}
		var baseSecs, langSecs []secPair
		curLang := SelectedLanguage()

		// We only care about [Info], [SceneDef], and [Scene N] (case-insensitive, language prefix allowed).
		interesting := func(logical string) bool {
			l := strings.ToLower(logical)
			return l == "info" || l == "scenedef" || strings.HasPrefix(l, "scene ")
		}

		for _, s := range src.Sections() {
			raw := s.Name()
			if raw == ini.DEFAULT_SECTION {
				continue
			}
			lang, base, has := splitLangPrefix(raw)
			logical := raw
			if has {
				logical = base
			}
			if !interesting(logical) {
				continue
			}
			if has {
				if lang == curLang {
					langSecs = append(langSecs, secPair{s, logical})
				}
			} else {
				baseSecs = append(baseSecs, secPair{s, logical})
			}
		}

		// one pass over a slice of sections; inject scene carry-over inside that pass only
		process := func(list []secPair) error {
			// carry-over across consecutive [Scene N] sections within THIS pass
			var clearcolor, clearalpha, clearlayerno, layerallpos string

			for _, sp := range list {
				section := sp.sec
				sectionName := sp.name // logical (no lang prefix)

				// Inject carry-over keys for scenes (so missing ones inherit within THIS pass)
				if strings.HasPrefix(strings.ToLower(sectionName), "scene ") {
					keysToCheck := map[string]*string{
						"clearcolor":   &clearcolor,
						"clearalpha":   &clearalpha,
						"clearlayerno": &clearlayerno,
						"layerall.pos": &layerallpos,
					}
					existing := make(map[string]bool)
					for _, k := range section.Keys() {
						if ptr, ok := keysToCheck[strings.ToLower(k.Name())]; ok {
							v, dup := iniFirstValue(k)
							if dup > 0 {
								LogMessage("Duplicate key [%s] %s (%d duplicate(s) ignored)", section.Name(), k.Name(), dup)
							}
							*ptr = v
							existing[strings.ToLower(k.Name())] = true
						}
					}
					for k, ptr := range keysToCheck {
						if !existing[k] && *ptr != "" {
							if _, err := section.NewKey(k, *ptr); err != nil {
								return fmt.Errorf("failed to add %s to %q: %w", k, section.Name(), err)
							}
						}
					}
				}

				for _, key := range section.Keys() {
					keyName := key.Name()
					value, dup := iniFirstValue(key)
					if dup > 0 {
						LogMessage("Duplicate key [%s] %s (%d duplicate(s) ignored)", section.Name(), keyName, dup)
					}
					fullKey := strings.ReplaceAll(sectionName, " ", "_") + "." + strings.ReplaceAll(keyName, " ", "_")
					keyParts := parseQueryPath(fullKey)
					if err := assignField(&s, keyParts, value, def); err != nil {
						// keep going; soft-fail like before
						//fmt.Printf("Warning: Failed to assign key [%s]: %v\n", fullKey, err)
					}
				}
			}
			return nil
		}

		// Apply base (unprefixed + en.*) first, then language overrides
		if err := process(baseSecs); err != nil {
			return err
		}
		// reset of carry-over between passes is handled by new local vars in process()
		if err := process(langSecs); err != nil {
			return err
		}
		return nil
	}

	// Apply precedence: struct zero/default tags < defaultStoryboard.ini < user storyboard
	if err := assignFrom(defaultOnlyIni); err != nil {
		return nil, err
	}
	if err := assignFrom(userIniFile); err != nil {
		return nil, err
	}

	s.IniFile = iniFile

	resolveInlineFonts(s.IniFile, s.Def, s.Fnt, s.fntIndexByKey, s.SetValueUpdate)
	syncFontsMap(&s.SceneDef.Font, s.Fnt, s.fntIndexByKey)

	s.loadFiles()

	str, err := LoadText(def)
	if err != nil {
		return nil, err
	}
	lines, i := SplitAndTrim(str, "\n"), 0
	s.AnimTable = ReadAnimationTable(s.Def, s.Sff, &s.Sff.palList, lines, &i, true)
	i = 0

	// Storyboard-specific quirk:
	// enable phantom pixel adjustment on all storyboard animations so that
	// frames flipped via H or V get an extra pixel of offset.
	for _, anim := range s.AnimTable.anims {
		if anim != nil {
			anim.phantomPixel = true
		}
	}

	s.populateDataPointers()

	for scene, sceneProps := range s.Scene {
		sceneName := strings.Replace(scene, "scene_", "scene ", 1)
		sceneProps.Music = parseMusicSection(pickLangSectionMerged(iniFile, sceneName))
		sceneProps.Music.DebugDump(fmt.Sprintf("Storyboard %s [%s]", def, sceneName))
	}

	return &s, nil
}

func (s *Storyboard) loadFiles() {
	LoadFile(&s.SceneDef.Spr, []string{s.SceneDef.Spr}, func(filename string) error {
		if filename != "" {
			var err error
			s.Sff, err = loadSff(filename, false, true, false)
			if err != nil {
				LogMessage("Failed to load %v: %v", filename, err)
			}
		}
		if s.Sff == nil {
			s.Sff = newSff()
		}
		return nil
	})

	LoadFile(&s.SceneDef.Model, []string{s.SceneDef.Model}, func(filename string) error {
		if filename != "" {
			var err error
			s.Model, err = loadglTFModel(filename)
			if err != nil {
				LogMessage("Failed to load %v: %v", filename, err)
			}
			sys.mainThreadTask <- func() {
				gfx.SetModelVertexData(1, s.Model.vertexBuffer)
				gfx.SetModelIndexData(1, s.Model.elementBuffer...)
			}
			sys.runMainThreadTask()
		}
		return nil
	})

	for _, scene := range s.Scene {
		if scene.Bg.Name != "" {
			var err error
			scene.Bg.BGDef, err = loadBGDef(s.Sff, s.Model, s.Def, scene.Bg.Name, 0)
			if err != nil {
				LogMessage("Failed to load %v (%v): %v", scene.Bg.Name, s.Def, err.Error())
			}
		}
		if scene.Bg.BGDef == nil {
			scene.Bg.BGDef = newBGDef(s.Def)
		}
		//scene.Bg.BgClearColor = scene.Bg.BGDef.bgclearcolor
	}

	LoadFile(&s.SceneDef.Snd, []string{s.SceneDef.Snd}, func(filename string) error {
		if filename != "" {
			var err error
			s.Snd, err = LoadSnd(filename)
			if err != nil {
				LogMessage("Failed to load %v: %v", filename, err)
			}
		}
		if s.Snd == nil {
			s.Snd = newSnd()
		}
		return nil
	})

	for key, fnt := range s.SceneDef.Font {
		LoadFile(&fnt.Font, []string{fnt.Font}, func(filename string) error {
			re := regexp.MustCompile(`\d+`)
			i := int(Atoi(re.FindString(key)))

			if filename != "" {
				var err error
				s.Fnt[i], err = loadFnt(filename, fnt.Height)
				registerFontIndex(s.fntIndexByKey, filename, fnt.Height, i)
				if err != nil {
					LogMessage("Failed to load %v: %v", filename, err)
				}
			}
			if s.Fnt[i] == nil {
				s.Fnt[i] = newFnt()
			}

			// Populate extended properties from the loaded font
			fnt.Type = s.Fnt[i].Type
			fnt.Size = s.Fnt[i].Size
			fnt.Spacing = s.Fnt[i].Spacing
			fnt.Offset = s.Fnt[i].offset

			return nil
		})
	}
}

// Initialize struct
func (s *Storyboard) initStruct() {
	initMaps(reflect.ValueOf(s).Elem())
	applyDefaultsToValue(reflect.ValueOf(s).Elem())
	s.fntIndexByKey = make(map[string]int)
}

// GetValue retrieves the value based on the query string.
func (s *Storyboard) GetValue(query string) (interface{}, error) {
	return GetValue(s, query)
}

// SetValue sets the value based on the query string and updates the IniFile.
func (s *Storyboard) SetValueUpdate(query string, value interface{}) error {
	return SetValueUpdate(s, s.IniFile, query, value)
}

// Save writes the current IniFile to disk, preserving comments and syntax.
func (s *Storyboard) Save(file string) error {
	return SaveINI(s.IniFile, file)
}

func (s *Storyboard) populateDataPointers() {
	PopulateDataPointers(s, s.Info.Localcoord)
}

func (s *Storyboard) reset() {
	if s.SceneDef.StopMusic {
		sys.bgm.Stop()
	}
	s.active = false
	s.initialized = false
	s.endTimer = -1
	s.currentSceneIndex = s.SceneDef.StartScene
	s.cancel = false
	for _, sceneProps := range s.Scene {
		if sceneProps.Bg.Name != "" {
			sceneProps.Bg.BGDef.Reset()
		}
		for _, layerProps := range sceneProps.Layer {
			if layerProps.AnimData != nil {
				layerProps.AnimData.Reset()
				layerProps.AnimData.AddPos(sceneProps.Layerall.Pos[0], sceneProps.Layerall.Pos[1])
				// Don't alias PalFX pointers between Anim/Text: they have independent runtime state.
				// Copy layer PalFX settings into the anim's own instance, then reset runtime state.
				if layerProps.PalFx.PalFxData != nil {
					if layerProps.AnimData.palfx == nil {
						layerProps.AnimData.palfx = newPalFX()
					}
					*layerProps.AnimData.palfx = *layerProps.PalFx.PalFxData
					layerProps.AnimData.palfx.clear()
				} else if layerProps.AnimData.palfx != nil {
					layerProps.AnimData.palfx.clear()
				}
				layerProps.AnimData.Update(false)
			}
			if layerProps.TextSpriteData != nil {
				layerProps.TextSpriteData.Reset()
				// Storyboard uses uses its own typewriter logic, so disable the internal TextSprite typing.
				layerProps.TextSpriteData.textDelay = 0
				// Apply layerall.pos so storyboard [layerall.pos] affects text too
				layerProps.TextSpriteData.AddPos(sceneProps.Layerall.Pos[0], sceneProps.Layerall.Pos[1])
				if layerProps.PalFx.PalFxData != nil {
					layerProps.TextSpriteData.palfx = layerProps.PalFx.PalFxData
					if layerProps.TextSpriteData.palfx != nil {
						layerProps.TextSpriteData.palfx.clear()
					}
				}
				// Re-apply font tuple RGBA after reset.
				layerProps.TextSpriteData.SetColor(layerProps.Font[3], layerProps.Font[4], layerProps.Font[5], layerProps.Font[6])
			}
			// Reset per-layer typing state
			layerProps.typedLen = 0
			layerProps.charDelayCounter = 0
			layerProps.lineFullyRendered = false
		}
	}
	s.fadePolicy = FadeStop
}

// Builds the ordered list of text layers for skip-to-advance. Order: StartTime ascending, then layer numeric index.
func (s *Storyboard) buildDialogueQueue(sceneProps *SceneProperties) {
	type entry struct {
		key   string
		lp    *LayerProperties
		start int32
		idx   int
	}

	parseLayerIndex := func(key string) int {
		// Accept "layer0", "layer_0", "Layer10", etc.
		re := regexp.MustCompile(`\d+`)
		num := re.FindString(key)
		if num == "" {
			return 1<<30 - 1
		}
		return int(Atoi(num))
	}

	var items []entry
	for _, key := range SortedKeys(sceneProps.Layer) {
		lp := sceneProps.Layer[key]
		if lp == nil || lp.TextSpriteData == nil || lp.Text == "" {
			continue
		}
		items = append(items, entry{
			key:   key,
			lp:    lp,
			start: lp.StartTime,
			idx:   parseLayerIndex(key),
		})
	}

	sort.SliceStable(items, func(i, j int) bool {
		if items[i].start != items[j].start {
			return items[i].start < items[j].start
		}
		if items[i].idx != items[j].idx {
			return items[i].idx < items[j].idx
		}
		return strings.ToLower(items[i].key) < strings.ToLower(items[j].key)
	})

	s.dialogueLayers = s.dialogueLayers[:0]
	for _, it := range items {
		s.dialogueLayers = append(s.dialogueLayers, it.lp)
	}
	s.dialoguePos = 0
}

// Advances the dialogue cursor past layers that are already over due to time.
// This prevents skip from trying to act on old layers if the scene has naturally progressed.
func (s *Storyboard) syncDialoguePosToTime() {
	for s.dialoguePos < len(s.dialogueLayers) {
		lp := s.dialogueLayers[s.dialoguePos]
		if lp == nil {
			s.dialoguePos++
			continue
		}
		// If the layer's time window has ended, advance past it.
		if lp.EndTime > 0 && s.counter >= lp.EndTime {
			s.dialoguePos++
			continue
		}
		break
	}
}

func (s *Storyboard) revealTextLayer(lp *LayerProperties) {
	if lp == nil || lp.TextSpriteData == nil || lp.Text == "" {
		return
	}
	// Only reveal once its StartTime is active (or we're at/after it).
	if s.counter < lp.StartTime {
		return
	}
	if lp.EndTime > 0 && s.counter >= lp.EndTime {
		return
	}
	runeCount := utf8.RuneCountInString(lp.Text)
	lp.typedLen = runeCount
	lp.lineFullyRendered = true
	lp.charDelayCounter = 0
	// Apply immediately so it draws this frame.
	lp.TextSpriteData.wrapText(lp.Text, lp.typedLen)
}

func (s *Storyboard) init() {
	//if !s.enabled {
	//	co.initialized = true
	//	return
	//}
	s.sceneKeys = SortedKeys(s.Scene)
	s.reset()
	s.counter = 0
	s.active = true
	s.initialized = true
}

func (s *Storyboard) step() {
	sys.stepCommandLists()
	sceneKey := s.sceneKeys[s.currentSceneIndex]
	sceneProps := s.Scene[sceneKey]

	// because skip may fast-forward s.counter, we must not rely on (s.counter == 0)
	// later in the frame. Capture scene start state up front.
	sceneJustStarted := (s.counter == 0)

	// Scene entry: build dialogue queue for this scene.
	if sceneJustStarted {
		s.buildDialogueQueue(sceneProps)
	}

	// Cancel handling
	if !s.SceneDef.DisableCancel && (sys.esc ||
		(!sys.motif.AttractMode.Enabled && sys.uiRawInput(s.SceneDef.Key.Cancel, -1)) ||
		(!sys.gameRunning && sys.motif.AttractMode.Enabled && sys.credits > 0)) {
		sys.esc = false
		s.cancel = true
	}

	// Skip handling
	skipPressed := sys.uiRawInput(s.SceneDef.Key.Skip, -1)

	// Keep dialogue cursor aligned with time progression.
	s.syncDialoguePosToTime()

	skipAdvancesScene := false
	if skipPressed {
		// Dialogue progression:
		// 1) If current layer hasn't started yet, fast-forward to its start.
		// 2) If it's typing, finish only that layer.
		// 3) If it's fully rendered, advance to next layer (fast-forward if needed).
		// 4) If no layers remain, skip advances the scene.
		if s.dialoguePos < len(s.dialogueLayers) {
			lp := s.dialogueLayers[s.dialoguePos]
			if lp != nil {
				// If we're before this layer's window, jump to just before StartTime so it begins immediately.
				if s.counter < lp.StartTime {
					s.counter = lp.StartTime - 1
				} else if (lp.EndTime <= 0 || s.counter < lp.EndTime) && !lp.lineFullyRendered {
					// Finish only the currently-typing layer.
					s.revealTextLayer(lp)
				} else {
					// Already fully rendered (or effectively done): advance to next layer.
					s.dialoguePos++
					s.syncDialoguePosToTime()
					if s.dialoguePos < len(s.dialogueLayers) {
						next := s.dialogueLayers[s.dialoguePos]
						if next != nil && s.counter < next.StartTime {
							s.counter = next.StartTime - 1
						}
					} else {
						// All dialogue layers handled: next skip should advance scene immediately.
						skipAdvancesScene = true
					}
				}
			} else {
				// Defensive: nil entry
				s.dialoguePos++
			}
		} else {
			// No dialogue layers (or already finished): skip advances the scene.
			skipAdvancesScene = true
		}
	}

	// Auto-end logic must use >= because skip can fast-forward time past end.time.
	endTime := sceneProps.End.Time
	startFadeAt := int32(-1)
	if endTime > 0 {
		fadeLen := int32(0)
		if sceneProps.FadeOut.FadeData != nil {
			fadeLen = sceneProps.FadeOut.FadeData.duration()
		}
		if fadeLen > 0 {
			startFadeAt = endTime - fadeLen
		} else {
			startFadeAt = endTime - 1
		}
		if startFadeAt < 0 {
			startFadeAt = 0
		}
	}
	reachedEndTime := endTime <= 0 || (startFadeAt >= 0 && s.counter >= startFadeAt)

	if s.endTimer == -1 && (reachedEndTime || s.cancel || skipAdvancesScene) {
		userInterrupt := s.cancel || skipAdvancesScene
		startFadeOut(sceneProps.FadeOut.FadeData, sys.motif.fadeOut, userInterrupt, s.fadePolicy)
		s.endTimer = s.counter + sys.motif.fadeOut.timeRemaining
	}

	// Run "scene start" init even if skip fast-forwarded s.counter this frame.
	if sceneJustStarted {
		if ok := sceneProps.Music.Play("", s.Def); ok {
			s.musicPlaying = true
		}
		sceneProps.FadeIn.FadeData.init(sys.motif.fadeIn, true)
	}

	if s.Snd != nil {
		for _, key := range SortedKeys(sceneProps.Sound) {
			soundProps := sceneProps.Sound[key]
			if s.counter == soundProps.StartTime {
				s.Snd.play(
					[...]int32{soundProps.Value[0], soundProps.Value[1]},
					soundProps.VolumeScale,
					soundProps.Pan,
					soundProps.LoopStart,
					soundProps.LoopEnd,
					soundProps.StartPosition,
				)
			}
		}
	}

	for _, layerProps := range sceneProps.Layer {
		// Update animations while the layer is actually visible.
		if s.counter >= layerProps.StartTime && (s.counter < layerProps.EndTime || layerProps.EndTime <= 0) {
			if layerProps.AnimData != nil {
				layerProps.AnimData.Update(false)
			}
		}
		if layerProps.TextSpriteData != nil && layerProps.Text != "" {
			nextCounter := s.counter + 1
			if nextCounter >= layerProps.StartTime && (nextCounter < layerProps.EndTime || layerProps.EndTime <= 0) {
				runeCount := utf8.RuneCountInString(layerProps.Text)
				if !layerProps.lineFullyRendered {
					StepTypewriter(
						layerProps.Text,
						&layerProps.typedLen,
						&layerProps.charDelayCounter,
						&layerProps.lineFullyRendered,
						layerProps.TextDelay,
					)
					if layerProps.typedLen > runeCount {
						layerProps.typedLen = runeCount
					}
					if layerProps.typedLen >= runeCount {
						layerProps.lineFullyRendered = true
					}
				}
				layerProps.TextSpriteData.wrapText(layerProps.Text, layerProps.typedLen)
				layerProps.TextSpriteData.Update()
			}
		}
	}

	// Only leave the storyboard once the fade-out has finished.
	if s.endTimer != -1 && s.counter >= s.endTimer {
		// Ensure no leftover storyboard fade decks bleed into the next screen.
		if sys.motif.fadeOut != nil {
			sys.motif.fadeOut.reset()
		}
		s.counter = 0
		s.endTimer = -1
		if s.cancel {
			s.currentSceneIndex = len(s.sceneKeys)
		} else {
			if sceneProps.Jump != nil && *sceneProps.Jump != s.currentSceneIndex {
				s.currentSceneIndex = *sceneProps.Jump
			} else {
				s.currentSceneIndex++
			}
		}
		if s.currentSceneIndex >= len(s.sceneKeys) {
			if s.musicPlaying && s.SceneDef.StopMusic {
				sys.bgm.Stop()
			}
			s.active = false
			// Do not force-reset fadeOut here; it will self-reset after its last draw.
			return
		}
		// Start the next scene's fade-in immediately (same step) so there is no gap frame.
		nextKey := s.sceneKeys[s.currentSceneIndex]
		if nextProps, ok := s.Scene[nextKey]; ok && nextProps.FadeIn.FadeData != nil {
			nextProps.FadeIn.FadeData.init(sys.motif.fadeIn, true)
		}
		return
	}

	s.counter++
}

func (s *Storyboard) draw(layerno int16) {
	sceneKey := s.sceneKeys[s.currentSceneIndex]
	sceneProps := s.Scene[sceneKey]

	if sceneProps.ClearColor[0] >= 0 {
		sceneProps.RectData.Draw(layerno)
	}

	if sceneProps.Bg.Name != "" {
		//if sceneProps.Bg.BgClearColor[0] >= 0 {
		//	sceneProps.Bg.RectData.Draw(layerno)
		//}
		sceneProps.Bg.BGDef.Draw(int32(layerno), 0, 0, 1)
	}

	for _, key := range SortedKeys(sceneProps.Layer) {
		layerProps := sceneProps.Layer[key]
		// Draw only within the layer's actual time window.
		// (Pre-start drawing is intentionally disallowed to prevent overlap when skipping.)
		if s.counter >= layerProps.StartTime && (s.counter < layerProps.EndTime || layerProps.EndTime <= 0) {
			if layerProps.AnimData != nil {
				layerProps.AnimData.Draw(layerno)
			}
			if layerProps.TextSpriteData != nil && layerProps.Text != "" {
				layerProps.TextSpriteData.Draw(layerno)
			}
		}
	}
}

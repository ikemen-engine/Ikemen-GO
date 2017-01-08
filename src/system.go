package main

import (
	"fmt"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/timshannon/go-openal/openal"
	"github.com/yuin/gopher-lua"
	"runtime"
	"strings"
	"sync"
	"time"
)

const MaxSimul = 4

var sys = System{
	randseed:  int32(time.Now().UnixNano()),
	scrrect:   [4]int32{0, 0, 320, 240},
	gameWidth: 320, gameHeight: 240,
	widthScale: 1, heightScale: 1,
	brightness: 256,
	roundTime:  -1,
	lifeMul:    1, team1VS2Life: 1,
	turnsRecoveryRate: 1.0 / 300,
	zoomMin:           1, zoomMax: 1, zoomSpeed: 1,
	lifebarFontScale:       1,
	mixer:                  *newMixer(),
	bgm:                    *newVorbis(),
	sounds:                 newSounds(),
	allPalFX:               *NewPalFX(),
	bgPalFX:                *NewPalFX(),
	sel:                    *newSelect(),
	match:                  1,
	listenPort:             "7500",
	loader:                 *newLoader(),
	numSimul:               [2]int{2, 2},
	numTurns:               [2]int{2, 2},
	afterImageMax:          8,
	attack_LifeToPowerMul:  0.7,
	getHit_LifeToPowerMul:  0.6,
	superpmap:              *NewPalFX(),
	super_TargetDefenceMul: 1.5,
	helperMax:              56}

type TeamMode int32

const (
	TM_Single TeamMode = iota
	TM_Simul
	TM_Turns
	TM_LAST = TM_Turns
)

type System struct {
	randseed                    int32
	scrrect                     [4]int32
	gameWidth, gameHeight       int32
	widthScale, heightScale     float32
	window                      *glfw.Window
	gameEnd, frameSkip          bool
	redrawWait                  struct{ nextTime, lastDraw time.Time }
	brightness                  int32
	introTime, roundTime        int32
	lifeMul, team1VS2Life       float32
	turnsRecoveryRate           float32
	zoomEnable                  bool
	zoomMin, zoomMax, zoomSpeed float32
	lifebarFontScale            float32
	debugFont                   *Fnt
	debugScript                 string
	mixer                       Mixer
	bgm                         Vorbis
	audioContext                *openal.Context
	nullSndBuf                  [audioOutLen * 2]int16
	sounds                      Sounds
	allPalFX                    PalFX
	bgPalFX                     PalFX
	lifebar                     Lifebar
	sel                         Select
	netInput                    *NetInput
	fileInput                   *FileInput
	aiInput                     [MaxSimul * 2]AiInput
	keyConfig                   []*KeyConfig
	com                         [MaxSimul * 2]int32
	autolevel                   bool
	home                        int
	gameTime                    int32
	match                       int32
	inputRemap                  [MaxSimul * 2]int
	listenPort                  string
	round                       int32
	intro                       int32
	time                        int32
	winTeam                     int
	matchWins                   [2]int32
	wins                        [2]int32
	roundsExisted               [2]int32
	draws                       int32
	loader                      Loader
	chars                       [MaxSimul * 2][]*Char
	charList                    CharList
	cgi                         [MaxSimul * 2]CharGlobalInfo
	tmode                       [2]TeamMode
	numSimul                    [2]int
	numTurns                    [2]int
	esc                         bool
	loadMutex                   sync.Mutex
	ignoreMostErrors            bool
	stringPool                  [MaxSimul * 2]StringPool
	bcStack, bcVarStack         BytecodeStack
	bcVar                       []BytecodeValue
	workingChar                 *Char
	specialFlag                 GlobalSpecialFlag
	afterImageMax               int
	attack_LifeToPowerMul       float32
	getHit_LifeToPowerMul       float32
	cameraPos                   [2]float32
	envShake                    EnvShake
	pause                       int32
	pausetime                   int32
	pausebg                     bool
	pauseendcmdbuftime          int32
	pauseplayer                 int
	super                       int32
	supertime                   int32
	superpausebg                bool
	superendcmdbuftime          int32
	superplayer                 int
	superdarken                 bool
	superanim                   *Animation
	superpmap                   PalFX
	superpos                    [2]float32
	superfacing                 float32
	superp2defmul               float32
	superunhittable             bool
	super_TargetDefenceMul      float32
	envcol                      [3]int32
	envcol_time                 int32
	envcol_under                bool
	clipboardText               [MaxSimul * 2][]string
	stage                       *Stage
	helperMax                   int
	nextCharId                  int32
}

func (s *System) init(w, h int32) *lua.LState {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	s.setWindowSize(w, h)
	var err error
	s.window, err = glfw.CreateWindow(int(s.scrrect[2]), int(s.scrrect[3]),
		"Ikemen GO", nil, nil)
	chk(err)
	s.window.MakeContextCurrent()
	s.window.SetKeyCallback(keyCallback)
	glfw.SwapInterval(1)
	chk(gl.Init())
	s.keyConfig = append(s.keyConfig, &KeyConfig{-1,
		int(glfw.KeyUp), int(glfw.KeyDown), int(glfw.KeyLeft), int(glfw.KeyRight),
		int(glfw.KeyZ), int(glfw.KeyX), int(glfw.KeyC),
		int(glfw.KeyA), int(glfw.KeyS), int(glfw.KeyD), int(glfw.KeyEnter)})
	RenderInit()
	s.audioOpen()
	l := lua.NewState()
	l.OpenLibs()
	for i := range s.inputRemap {
		s.inputRemap[i] = i
	}
	for i := range s.stringPool {
		s.stringPool[i] = *NewStringPool()
	}
	systemScriptInit(l)
	return l
}
func (s *System) setWindowSize(w, h int32) {
	s.scrrect[2], s.scrrect[3] = w, h
	if s.scrrect[2]*3 > s.scrrect[3]*4 {
		s.gameWidth, s.gameHeight = s.scrrect[2]*3*320/(s.scrrect[3]*4), 240
	} else {
		s.gameWidth, s.gameHeight = 320, s.scrrect[3]*4*240/(s.scrrect[2]*3)
	}
	s.widthScale = float32(s.scrrect[2]) / float32(s.gameWidth)
	s.heightScale = float32(s.scrrect[3]) / float32(s.gameHeight)
}
func (s *System) await(fps int) {
	s.playSound()
	if !s.frameSkip {
		s.window.SwapBuffers()
	}
	now := time.Now()
	diff := s.redrawWait.nextTime.Sub(now)
	wait := time.Second / time.Duration(fps)
	s.redrawWait.nextTime = s.redrawWait.nextTime.Add(wait)
	switch {
	case diff >= 0 && diff < wait+2*time.Millisecond:
		time.Sleep(diff)
		fallthrough
	case now.Sub(s.redrawWait.lastDraw) > 250*time.Millisecond:
		fallthrough
	case diff >= -17*time.Millisecond:
		s.redrawWait.lastDraw = now
		s.frameSkip = false
	default:
		if diff < -150*time.Millisecond {
			s.redrawWait.nextTime = now.Add(wait)
		}
		s.frameSkip = true
	}
	s.esc = false
	glfw.PollEvents()
	s.gameEnd = s.window.ShouldClose()
	if !s.frameSkip {
		gl.Viewport(0, 0, int32(s.scrrect[2]), int32(s.scrrect[3]))
		gl.Clear(gl.COLOR_BUFFER_BIT)
	}
}
func (s *System) resetRemapInput() {
	for i := range s.inputRemap {
		s.inputRemap[i] = i
	}
}
func (s *System) loaderReset() {
	s.round, s.wins, s.roundsExisted = 1, [2]int32{0, 0}, [2]int32{0, 0}
	s.loader.reset()
}
func (s *System) loadStart() {
	s.loaderReset()
	s.loader.runTread()
}
func (s *System) synchronize() error {
	if s.fileInput != nil {
		return s.fileInput.Synchronize()
	} else if s.netInput != nil {
		return s.netInput.Synchronize()
	}
	return nil
}
func (s *System) matchOver() bool {
	return s.wins[0] >= s.matchWins[0] || s.wins[1] >= s.matchWins[1]
}
func (s *System) sf(gsf GlobalSpecialFlag) bool {
	return s.specialFlag&gsf != 0
}
func (s *System) setSF(gsf GlobalSpecialFlag) {
	s.specialFlag |= gsf
}
func (s *System) unsetSF(gsf GlobalSpecialFlag) {
	s.specialFlag &^= gsf
}
func (s *System) appendToClipboard(pn, sn int, a ...interface{}) {
	spl := s.stringPool[pn].List
	if sn >= 0 && sn < len(spl) {
		s.clipboardText[pn] = append(s.clipboardText[pn],
			strings.Split(fmt.Sprintf(spl[sn], a...), "\n")...)
	}
}
func (s *System) newCharId() int32 {
	s.nextCharId++
	return s.nextCharId - 1
}
func (s *System) fight() (reload bool) {
	unimplemented()
	return false
}

type SelectChar struct {
	def, name, sprite    string
	sportrait, lportrait *Sprite
}
type SelectStage struct {
	def, name string
}
type Select struct {
	columns, rows   int
	cellsize        [2]float32
	cellscale       [2]float32
	randomspr       *Sprite
	randomscl       [2]float32
	charlist        []SelectChar
	stagelist       []SelectStage
	curStageNo      int
	selected        [2][][2]int
	selectedStageNo int
}

func newSelect() *Select {
	return &Select{columns: 5, rows: 2, randomscl: [2]float32{1, 1},
		cellsize: [2]float32{29, 29}, cellscale: [2]float32{1, 1},
		selectedStageNo: -1}
}
func (s *Select) GetCharNo(i int) int {
	n := i
	if len(s.charlist) > 0 {
		n %= len(s.charlist)
		if n < 0 {
			n += len(s.charlist)
		}
	}
	return n
}
func (s *Select) GetChar(i int) *SelectChar {
	if len(s.charlist) == 0 {
		return nil
	}
	n := s.GetCharNo(i)
	return &s.charlist[n]
}
func (s *Select) SetStageNo(n int) int {
	s.curStageNo = n % (len(s.stagelist) + 1)
	if s.curStageNo < 0 {
		s.curStageNo += len(s.stagelist) + 1
	}
	return s.curStageNo
}
func (s *Select) SelectStage(n int) { s.selectedStageNo = n }
func (s *Select) GetStageName(n int) string {
	n %= len(s.stagelist) + 1
	if n < 0 {
		n += len(s.stagelist) + 1
	}
	if n == 0 {
		return "Random"
	}
	return s.stagelist[n-1].name
}
func (s *Select) AddCahr(def string) {
	s.charlist = append(s.charlist, SelectChar{})
	sc := &s.charlist[len(s.charlist)-1]
	def = strings.Replace(strings.TrimSpace(strings.Split(def, ",")[0]),
		"\\", "/", -1)
	if strings.ToLower(def) == "randomselect" {
		sc.def, sc.name = "randomselect", "Random"
		return
	}
	idx := strings.Index(def, "/")
	if len(def) >= 4 && strings.ToLower(def[len(def)-4:]) == ".def" {
		if idx < 0 {
			return
		}
	} else if idx < 0 {
		def += "/" + def + ".def"
	} else {
		def += ".def"
	}
	if def[0] != '/' || idx > 0 && strings.Index(def[:idx], ":") < 0 {
		def = "chars/" + def
	}
	if def = FileExist(def); len(def) == 0 {
		return
	}
	str, err := LoadText(def)
	if err != nil {
		return
	}
	sc.def = def
	lines, i, info, files, sprite := SplitAndTrim(str, "\n"), 0, true, true, ""
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "info":
			if info {
				info = false
				var ok bool
				sc.name, ok, _ = is.getText("displayname")
				if !ok {
					sc.name, _, _ = is.getText("name")
				}
			}
		case "files":
			if files {
				files = false
				sprite = is["sprite"]
			}
		}
	}
	sc.sprite = sprite
	LoadFile(&sprite, def, func(file string) error {
		var err error
		sc.sportrait, err = LoadFromSff(file, 9000, 0)
		return err
	})
	sprite = sc.sprite
	LoadFile(&sprite, def, func(file string) error {
		var err error
		sc.lportrait, err = LoadFromSff(file, 9000, 1)
		return err
	})
}
func (s *Select) AddStage(def string) error {
	var lines []string
	if err := LoadFile(&def, "stages/", func(file string) error {
		str, err := LoadText(file)
		if err != nil {
			return err
		}
		lines = SplitAndTrim(str, "\n")
		return nil
	}); err != nil {
		return err
	}
	i, info := 0, true
	s.stagelist = append(s.stagelist, SelectStage{})
	ss := &s.stagelist[len(s.stagelist)-1]
	ss.def = def
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "info":
			if info {
				info = false
				var ok bool
				ss.name, ok, _ = is.getText("displayname")
				if !ok {
					ss.name, ok, _ = is.getText("name")
					if !ok {
						ss.name = def
					}
				}
			}
		}
	}
	return nil
}
func (s *Select) AddSelectedChar(tn, cn, pl int) bool {
	m, n := 0, s.GetCharNo(cn)
	if len(s.charlist) == 0 || len(s.charlist[n].def) == 0 {
		return false
	}
	for s.charlist[n].def == "randomselect" || len(s.charlist[n].def) == 0 {
		m++
		if m > 100000 {
			return false
		}
		n = int(Rand(0, int32(len(s.charlist))-1))
		pl = int(Rand(1, 12))
	}
	sys.loadMutex.Lock()
	s.selected[tn] = append(s.selected[tn], [2]int{n, pl})
	sys.loadMutex.Unlock()
	return true
}
func (s *Select) ClearSelected() {
	sys.loadMutex.Lock()
	s.selected = [2][][2]int{}
	sys.loadMutex.Unlock()
	s.selectedStageNo = -1
}

type LoaderState int32

const (
	LS_NotYet LoaderState = iota
	LS_Loading
	LS_Complete
	LS_Error
	LS_Cancel
)

type Loader struct {
	state    LoaderState
	loadExit chan LoaderState
	err      error
}

func newLoader() *Loader {
	return &Loader{state: LS_NotYet, loadExit: make(chan LoaderState, 1)}
}
func (l *Loader) loadChar(pn int) int {
	sys.loadMutex.Lock()
	result := -1
	nsel := len(sys.sel.selected[pn&1])
	if sys.tmode[pn&1] == TM_Simul {
		if pn>>1 >= sys.numSimul[pn&1] {
			sys.cgi[pn].states = nil
			sys.chars[pn] = nil
			result = 1
		}
	} else if pn >= 2 {
		result = 0
	}
	if sys.tmode[pn&1] == TM_Turns && nsel < sys.numTurns[pn&1] {
		result = 0
	}
	memberNo := pn >> 1
	if sys.tmode[pn&1] == TM_Turns {
		memberNo = int(sys.wins[^pn&1])
	}
	if nsel <= memberNo {
		result = 0
	}
	if result >= 0 {
		sys.loadMutex.Unlock()
		return result
	}
	pal, idx := int32(sys.sel.selected[pn&1][memberNo][1]), make([]int, nsel)
	for i := range idx {
		idx[i] = sys.sel.selected[pn&1][i][0]
	}
	sys.loadMutex.Unlock()
	cdef := sys.sel.charlist[idx[memberNo]].def
	var p *Char
	if len(sys.chars) > 0 && cdef == sys.cgi[pn].def {
		p = sys.chars[pn][0]
		p.key = pn
		if sys.com[pn] != 0 {
			p.key ^= -1
		}
	} else {
		p = newChar(pn, 0)
		sys.cgi[pn].sff = nil
	}
	sys.chars[pn] = make([]*Char, 1)
	sys.chars[pn][0] = p
	if sys.roundsExisted[pn&1] == 0 {
		sys.cgi[pn].palno = sys.cgi[pn].palkeymap[pal-1] + 1
	}
	if sys.cgi[pn].sff == nil {
		if sys.cgi[pn].states, l.err =
			newCompiler().Compile(p.playerNo, cdef); l.err != nil {
			sys.chars[pn] = nil
			return -1
		}
		if l.err = p.load(cdef); l.err != nil {
			sys.chars[pn] = nil
			return -1
		}
	}
	if pn < len(sys.lifebar.fa[sys.tmode[pn&1]]) {
		fa := sys.lifebar.fa[sys.tmode[pn&1]][pn]
		fa.face = sys.cgi[pn].sff.GetOwnPalSprite(
			int16(fa.face_spr[0]), int16(fa.face_spr[1]))
		if sys.tmode[pn&1] == TM_Turns && sys.round == 1 {
			fa.numko = 0
			fa.teammate_face = make([]*Sprite, nsel)
			for i, ci := range idx {
				sprite := sys.sel.charlist[ci].sprite
				LoadFile(&sprite, sys.sel.charlist[ci].def, func(file string) error {
					var err error
					fa.teammate_face[i], err = LoadFromSff(file,
						int16(fa.teammate_face_spr[0]), int16(fa.teammate_face_spr[1]))
					return err
				})
			}
		}
	}
	return 1
}
func (l *Loader) loadStage() bool {
	var def string
	if sys.sel.selectedStageNo == 0 {
		def = sys.sel.stagelist[Rand(0, int32(len(sys.sel.stagelist))-1)].def
	} else {
		def = sys.sel.stagelist[sys.sel.selectedStageNo-1].def
	}
	if sys.stage != nil && sys.stage.def == def {
		return true
	}
	sys.stage, l.err = LoadStage(def)
	return l.err == nil
}
func (l *Loader) load() {
	defer func() { l.loadExit <- l.state }()
	charDone, stageDone := make([]bool, len(sys.chars)), false
	allCharDone := func() bool {
		for _, b := range charDone {
			if !b {
				return false
			}
		}
		return true
	}
	for !stageDone || !allCharDone() {
		runtime.LockOSThread()
		for i, b := range charDone {
			if !b {
				result := l.loadChar(i)
				if result > 0 {
					charDone[i] = true
				} else if result < 0 {
					l.state = LS_Error
					return
				}
			}
		}
		for i := 0; i < 2; i++ {
			if !charDone[i+2] && len(sys.sel.selected[i]) > 0 &&
				sys.tmode[i] != TM_Simul {
				for j := i + 2; j < len(sys.chars); j += 2 {
					sys.chars[j], sys.cgi[j].states, charDone[j] = nil, nil, true
					sys.cgi[j].wakewakaLength = 0
				}
			}
		}
		if !stageDone && sys.sel.selectedStageNo >= 0 {
			if !l.loadStage() {
				l.state = LS_Error
				return
			}
			stageDone = true
		}
		runtime.UnlockOSThread()
		time.Sleep(10 * time.Millisecond)
		if sys.gameEnd {
			l.state = LS_Cancel
		}
		if l.state == LS_Cancel {
			return
		}
	}
	l.state = LS_Complete
}
func (l *Loader) reset() {
	if l.state != LS_NotYet {
		l.state = LS_Cancel
		<-l.loadExit
		l.state = LS_NotYet
	}
	l.err = nil
	for i := range sys.cgi {
		if sys.roundsExisted[i&1] == 0 {
			sys.cgi[i].drawpalno = -1
		}
	}
}
func (l *Loader) runTread() bool {
	if l.state != LS_NotYet {
		return false
	}
	l.state = LS_Loading
	go l.load()
	return true
}

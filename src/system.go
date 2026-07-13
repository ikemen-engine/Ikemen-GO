package main

import (
	"arena"
	"bufio"
	"errors"
	"fmt"
	"image"
	"io"

	//"log"
	"math"
	"os"
	"path"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gopxl/beep/v2"

	//glfont "github.com/ikemen-engine/glfont"
	lua "github.com/yuin/gopher-lua"
)

const (
	MaxSimul        = 4
	MaxAttachedChar = 4
	MaxPlayerNo     = MaxSimul*2 + MaxAttachedChar
)

// The variables placed in this struct will be saved/loaded automatically by game states
// Note: Still need to deep copy pointers etc like usual
// TODO: Testing the changes and cleaning up
type SystemStateVars struct {
	randseed          int32
	matchTime         int32
	curRoundTime      int32
	persistRoundCount int32
	curPlayTime       int32

	aiInput    [MaxPlayerNo]AiInput
	ffbparams  [MaxPlayerNo]ForceFeedbackParams
	inputRemap [MaxPlayerNo]int
	aiLevel    [MaxPlayerNo]float32
	cam        Camera

	pausetime          int32
	pausebg            bool
	pauseendcmdbuftime int32
	pausetimebuffer    int32
	pauseplayerno      int
	supertimebuffer    int32
	supertime          int32
	superpausebg       bool
	superendcmdbuftime int32
	superplayerno      int
	superbrightness    float32

	envShake    EnvShake
	specialFlag GlobalSpecialFlag
	envcol      [3]int32
	envcol_time int32

	scrrect                 [4]int32
	gameWidth, gameHeight   float32
	widthScale, heightScale float32
	gameEnd, frameSkip      bool
	paused, frameStepFlag   bool
	brightness              float32
	brightnessOld           float32
	maxRoundTime            int32
	curFramesPerCount       int32 // The mutatable value for the current match
	matchNo                 int32
	roundNo                 int32
	intro                   int32
	lastHitter              [2]int
	winTeam                 int
	winType                 [2]WinType
	winTrigger              [2]WinType
	matchWins, wins         [2]int32 // Required wins, current wins
	roundsExisted           [2]int32
	draws                   int32
	maxDraws                [2]int32
	effectiveLoss           [2]bool
	tmode                   [2]TeamMode
	numSimul, numTurns      [2]int32
	esc                     bool
	envcol_under            bool
	lastCharId              int32
	tickCount               int
	oldTickCount            int
	tickCountF              float32
	lastTick                float32
	nextAddTime             float32
	oldNextAddTime          float32
	xmin, xmax              float32
	zmin, zmax              float32
	winskipped              bool
	roundResetFlg           bool
	roundResetMatchStart    bool
	reloadFlg               bool
	reloadStageFlg          bool
	reloadFightScreenFlg    bool
	reloadCharSlot          [MaxPlayerNo]bool
	turbo                   float32
	zoom                    ZoomEffect
	finishType              FinishType
	winwaittime             int32
	slowtime                int32
	changeStateNest         int32
	nomusic                 bool
	timerStart              int32
	teamLeader              [2]int
	postMatchFlg            bool
	scoreStart              [2]float32
	scorePoints             [2]float32
	comboCount              [2]int32
	decisiveRound           [2]bool
	gameMode                string
	credits                 int32

	consecutiveWins  [2]int32
	firstAttack      [3]int
	home             int
	stageLoop        bool
	dialogueHideBars bool
	dialogueForce    int
	playBgmFlg       bool

	keyInput            Key
	keyString           string
	lastInputController int
	uiLastInputToken    string
	uiConsumeInputFrame int32
	uiRepeatToken       string
	uiRepeatController  int
	uiRepeatFrame       int32

	endMatch       bool
	noCharSoundFlg bool
	fightLoopEnd   bool
	continueFlg    bool
	matchResetFlg  bool
	stageLoopNo    int
	introSkipCall  bool
	preMatchTime   int32
	loopBreak      bool
	loopContinue   bool
	winposetime    int32
}

// sys
// The only instance of a System struct.
// Do not create more than 1.
var sys = System{
	soundMixer: &beep.Mixer{},
	bgm:        *newBgm(),
	//soundChannels: newSoundChannels(16), // Lazy allocation in Request()
	allPalFX: newPalFX(),
	bgPalFX:  newPalFX(),
	ffx:      make(map[string]*FightFx),
	//ffxRegexp:         "^(f)|^(s)|^(go)", // https://github.com/ikemen-engine/Ikemen-GO/issues/1620
	sel:              *newSelect(),
	keyState:         make(map[Key]bool),
	loader:           *newLoader(),
	ignoreMostErrors: true,
	stageList:        make(map[int32]*Stage),
	stageLocalcoords: make(map[string][2]int32),
	commandLine:      make(chan string),
	mainThreadTask:   make(chan func(), 65536),
	workpal:          make([]uint32, 256),
	saveState:        NewGameState(),
	statePool:        NewGameStatePool(),
	savePool:         NewGameStatePool(),
	loadPool:         NewGameStatePool(),
	commandLists:     make([]*CommandList, 0),
	arenaSaveMap:     make(map[int]*arena.Arena),
	arenaLoadMap:     make(map[int]*arena.Arena),
	debugAccel:       1, // TODO: We probably shouldn't rely on this being initialized to 1
	charVarsBackup:   make(map[int]CharVarBackup),
	SystemStateVars: SystemStateVars{
		randseed:            int32(time.Now().UnixNano()),
		scrrect:             [...]int32{0, 0, 320, 240},
		gameWidth:           320,
		gameHeight:          240,
		widthScale:          1,
		heightScale:         1,
		brightness:          1,
		maxRoundTime:        -1,
		matchNo:             1,
		numSimul:            [...]int32{2, 2},
		numTurns:            [...]int32{2, 2},
		oldNextAddTime:      1,
		cam:                 *newCamera(),
		keyInput:            KeyUnknown,
		lastInputController: -1,
		uiRepeatController:  -1,
	},
}

type TeamMode int32

const (
	TM_Single TeamMode = iota
	TM_Simul
	TM_Turns
	TM_Tag
	TM_LAST = TM_Tag
)

// System struct, holds most of the data that is accessed globally through the program.
type System struct {
	SystemStateVars

	window              *Window
	redrawWait          struct{ nextTime, lastDraw time.Time }
	debugFont           *TextSprite
	debugDisplay        bool
	debugRef            [2]int // player number, helper index
	debugLastID         int32
	soundMixer          *beep.Mixer
	bgm                 Bgm
	matchMusicSel       []*bgMusic
	pauseVolumeApplied  bool
	soundChannels       SoundChannels // System sounds. Lifebars etc
	charSoundChannels   [MaxPlayerNo]SoundChannels
	allPalFX            *PalFX
	bgPalFX             *PalFX
	fightScreen         FightScreen
	motif               Motif
	storyboard          Storyboard
	cfg                 Config
	ffx                 map[string]*FightFx
	sel                 Select
	keyState            map[Key]bool
	netConnection       *NetConnection
	replayFile          *ReplayFile
	keyConfig           []KeyConfig
	joystickConfig      []KeyConfig
	loader              Loader
	chars               [MaxPlayerNo][]*Char
	charList            CharList
	cgi                 [MaxPlayerNo]CharGlobalInfo
	turnsPreloadMember  [2]int // -1 = none; otherwise selected Turns member index to load into side+2
	selMutex            sync.RWMutex
	loadMutex           sync.Mutex
	ignoreMostErrors    bool
	stringPool          [MaxPlayerNo]StringPool
	bcStack, bcVarStack BytecodeStack
	bcVar               []BytecodeValue
	workingChar         *Char          // Char currently running its states
	workingState        *StateBytecode // State currently running
	stage               *Stage
	stageList           map[int32]*Stage
	stageLocalcoords    map[string][2]int32
	wireframeDisplay    bool
	shortcutScripts     map[ShortcutKey]*ShortcutScript
	commandLine         chan string
	debugWC             *Char
	projs               [MaxPlayerNo][]*Projectile
	explods             [MaxPlayerNo][]*Explod
	explodDrawOrder     []*Explod
	chartexts           [MaxPlayerNo][]*TextSprite // From Text sctrl
	spriteList          DrawList
	shadowList          ShadowList
	reflectionList      ReflectionList
	afterImageCount     [MaxPlayerNo]int32
	debugc1hit          DebugClsn
	debugc1rev          DebugClsn
	debugc1not          DebugClsn
	debugc2             DebugClsn
	debugc2hb           DebugClsn
	debugc2mtk          DebugClsn
	debugc2grd          DebugClsn
	debugc2stb          DebugClsn
	debugcsize          DebugClsn
	debugch             DebugClsn
	debugAccel          float32
	clsnSpr             Sprite
	clsnDisplay         bool
	lifebarHide         bool
	mainThreadTask      chan func()
	workpal             []uint32
	workBe              []BytecodeExp
	timerCount          []int32
	cmdFlags            map[string]string
	whitePalTex         Texture
	usePalette          bool
	gameRunning         bool

	msaa               int32
	externalShaders    [][][]byte
	windowMainIcon     []image.Image
	frameCounter       int32
	captureNum         int
	timerRounds        []int32
	scoreRounds        [][2]float32
	statsLog           StatsLog
	maxPowerMode       bool
	debugClsnText      []DebugClsnText
	consoleText        []string
	luaLState          *lua.LState
	statusLFunc        *lua.LFunction
	listLFunc          []*lua.LFunction
	reloadPreserveVars [MaxPlayerNo]bool
	charVarsBackup     map[int]CharVarBackup
	shaderRefCount     map[string]int

	statePool       GameStatePool
	commandLists    []*CommandList
	arenaSaveMap    map[int]*arena.Arena
	arenaLoadMap    map[int]*arena.Arena
	rollbackStateID int
	savePool        GameStatePool
	loadPool        GameStatePool
	rollback        RollbackSystem
	rollbackConfig  RollbackProperties
	saveState       *GameState
	saveStateFlag   bool
	loadStateFlag   bool

	// Match loop variables
	roundBackup RoundStartBackup
	matchBackup RoundStartBackup

	// for avg. FPS calculations
	gameFPS          float32
	gameFPSprevcount uint64

	// screenshot deferral
	isTakingScreenshot bool

	// keepAlive profiling (debug only)
	keepAliveProfile bool
	keepAliveOnce    sync.Once
	keepAlivePrev    time.Time
	keepAliveStart   time.Time
	keepAliveCount   int

	luaDrawPreOps   []func()
	luaDrawLayerOps [3][]func()

	// UI repeat guard: avoid firing the same token multiple times in the same frame

	// UI command registry. Ensures new players get the full command set.
	uiCommandRegistry map[string]CommandSpec

	// For Android support
	baseDir string

	// Session-wide sync config override for netplay/replay.
	// Must live on System because rollback temporarily steals sys.netConnection.
	netplayOverride SessionConfigOverride
	sessionWarning  string
}

type drawAspectState struct {
	gameWidth, gameHeight   float32
	widthScale, heightScale float32
}

// Check if the application is running inside a macOS app bundle
func isRunningInsideAppBundle(exePath string) bool {
	// Check if we're on Darwin and the executable path contains .app (macOS application bundle)
	return runtime.GOOS == "darwin" && strings.Contains(exePath, ".app")
}

// Initialize stuff, this is called after the config int at main.go
func (s *System) init(w, h int32) *lua.LState {
	var err error

	s.setGameSize(w, h)
	for i := range sys.cgi {
		sys.cgi[i].palInfo = make(map[int]PalInfo)
	}

	// Select the renderer first so we can fall back to default before doing anything
	// This is the only time we should check s.cfg.Video.RenderMode
	Logcat("Check A: Selecting Renderer")
	gfx, gfxFont = selectRenderer(s.cfg.Video.RenderMode)

	// Check the actual render name in case the config file was set up incorrectly
	renderName := gfx.GetName()

	// Create a system window.
	s.window, err = s.newWindow(int(s.scrrect[2]), int(s.scrrect[3]))
	chk(err)

	// Check if it's actually fullscreen. For some reason the constructor doesn't handle this properly.
	_, forceWindowed := s.cmdFlags["-windowed"]
	s.window.fullscreen = s.cfg.Video.Fullscreen && !forceWindowed

	if strings.HasPrefix(renderName, "OpenGL") {
		if ctx, err := s.window.GLCreateContext(); err != nil {
			panic(fmt.Sprintf("Could not initialize context :( Reason? %s", err))
		} else {
			s.window.GLMakeCurrent(ctx)
		}
	}

	if runtime.GOOS != "android" {
		exePath, err := os.Executable()
		if err != nil {
			fmt.Println("Error getting executable path:", err)
		} else {
			// Change the context for Darwin if we're in an app bundle
			if isRunningInsideAppBundle(exePath) {
				os.Chdir(path.Dir(exePath))
				os.Chdir("../../../")
			}
		}
		// Update the gamepad mappings with user mappings, if present.
		input.UpdateGamepadMappings(sys.cfg.Config.GamepadMappings)
	}

	// Loading of external shader data.
	// We need to do this before the render initialization at "gfx.Init()"
	if len(s.cfg.Video.ExternalShaders) > 0 {
		// First we initialize arrays.
		s.externalShaders = make([][][]byte, 2)
		s.externalShaders[0] = make([][]byte, len(s.cfg.Video.ExternalShaders))
		s.externalShaders[1] = make([][]byte, len(s.cfg.Video.ExternalShaders))

		// Then we load.
		for i, shaderLocation := range s.cfg.Video.ExternalShaders {
			// Create names.
			shaderLocation = strings.Replace(shaderLocation, "\\", "/", -1)

			if strings.HasPrefix(renderName, "OpenGL") {
				// Load vert shaders.
				s.externalShaders[0][i], err = os.ReadFile(shaderLocation + ".vert")
				if err != nil {
					chk(err)
				}

				// Load frag shaders.
				s.externalShaders[1][i], err = os.ReadFile(shaderLocation + ".frag")
				if err != nil {
					chk(err)
				}
			} else if strings.HasPrefix(renderName, "Vulkan") {
				// Load spv shaders
				s.externalShaders[0][i], err = os.ReadFile(shaderLocation + ".vert.spv")
				if err != nil {
					chk(err)
				}

				s.externalShaders[1][i], err = os.ReadFile(shaderLocation + ".frag.spv")
				if err != nil {
					chk(err)
				}
			}
		}
	}
	// PS: The "\x00" is what is know as Null Terminator.

	// Init renderer
	Logcat("Check B: Initializing GFX")
	gfx.Init()
	s.window.SetSwapInterval(s.cfg.Video.VSync) // VSync must be set after Init(), or our config.ini will be ignored

	Logcat("Check C: Initializing GFX Font")
	gfxFont.Init(gfx)
	Logcat("Check D: We are GOOD")
	gfx.BeginFrame(false)

	// And the audio.
	speaker = &SDLSpeaker{}
	speaker.Init(beep.SampleRate(sys.cfg.Sound.SampleRate), audioOutLen)
	speaker.Play(NewNormalizer(s.soundMixer))
	l := lua.NewState()
	l.Options.IncludeGoStackTrace = true
	l.OpenLibs()
	s.resetRemapInput()
	for i := range s.stringPool {
		s.stringPool[i] = *NewStringPool()
	}
	s.clsnSpr = *newSprite()
	s.clsnSpr.Size, s.clsnSpr.Pal = [...]uint16{1, 1}, make([]uint32, 256)
	s.clsnSpr.SetPxl([]byte{0})
	// Create a reusable white palette texture for shadows
	whitepal := make([]uint32, 256)
	for i := 1; i < 256; i++ {
		whitepal[i] = 0xffffffff // White (and full alpha)
	}
	s.whitePalTex = gfx.newPaletteTexture()
	s.whitePalTex.SetData(Pal32ToBytes(whitepal))

	systemScriptInit(l)
	s.shortcutScripts = make(map[ShortcutKey]*ShortcutScript)
	s.shaderRefCount = make(map[string]int)
	if runtime.GOOS != "android" {
		// So now that we have a window we add an icon.
		if len(s.cfg.Config.WindowIcon) > 0 {
			// First we initialize arrays.
			var f = make([]io.ReadCloser, len(s.cfg.Config.WindowIcon))
			s.windowMainIcon = make([]image.Image, len(s.cfg.Config.WindowIcon))
			// And then we load them.
			for i, iconLocation := range s.cfg.Config.WindowIcon {
				f[i], err = os.Open(iconLocation)
				if err != nil {
					var dErr = "Icon file can not be found.\nPanic: " + err.Error()
					ShowErrorDialog(dErr)
					panic(Error(dErr))
				}
				s.windowMainIcon[i], _, err = image.Decode(f[i])
			}
			s.window.SetIcon(s.windowMainIcon)
			chk(err)
		}

		// Error print?
		SafeGo(func() {
			stdin := bufio.NewScanner(os.Stdin)
			for stdin.Scan() {
				if err := stdin.Err(); err != nil {
					LogMessage(err.Error())
					return
				}
				s.commandLine <- stdin.Text()
			}
		})
	}
	return l
}

func (s *System) shutdown() {
	if !sys.gameEnd {
		sys.gameEnd = true
	}
	if sys.rollback.session != nil && sys.rollback.session.recording != nil {
		sys.rollback.session.SaveReplay()
	}
	gfx.Close()
	s.window.Close()
	if speaker != nil {
		speaker.Close()
	}
}

func getViewport(srcW, srcH, dstW, dstH float64) [4]float64 {
	fromRatio := srcW * dstH
	toRatio := srcH * dstW
	if fromRatio > toRatio {
		// Source is wider than target aspect ratio
		w := srcH * dstW / dstH
		h := srcH
		x := (srcW - w) / 2
		return [4]float64{x, 0, w, h}
	} else if fromRatio < toRatio {
		// Source is taller than target aspect ratio
		w := srcW
		h := srcW * dstH / dstW
		y := (srcH - h) / 2
		return [4]float64{0, y, w, h}
	}
	// Source and target have the same aspect ratio
	return [4]float64{0, 0, srcW, srcH}
}

func (s *System) middleOfMatch() bool {
	return !s.fightLoopEnd && s.matchTime != 0 && !s.postMatchFlg
}

func (s *System) skipMotifScaling() bool {
	var lc [2]int32
	if (!s.middleOfMatch() && !s.postMatchFlg) || s.stage == nil {
		lc = s.motif.Info.Localcoord
	} else {
		lc = s.stage.stageCamera.localcoord
	}
	return CalculateAspect(lc[0], lc[1]) > s.getMotifAspect()
}

func (s *System) getFightAspect() float32 {
	// Stage aspect ratio
	if s.cfg.Video.FightAspectWidth < 0 && s.cfg.Video.FightAspectHeight < 0 && s.stage != nil {
		coord := s.stage.stageCamera.localcoord
		if coord[0] > 0 && coord[1] > 0 {
			return CalculateAspect(coord[0], coord[1])
		}
	}

	// Custom aspect ratio
	if s.cfg.Video.FightAspectWidth > 0 && s.cfg.Video.FightAspectHeight > 0 {
		return CalculateAspect(s.cfg.Video.FightAspectWidth, s.cfg.Video.FightAspectHeight)
	}

	// Default
	// Using video options directly has unwanted behavior if those options are changed without restarting Ikemen
	//return float32(s.cfg.Video.GameWidth) / float32(s.cfg.Video.GameHeight)
	return s.getMotifAspect()
}

func (s *System) getMotifAspect() float32 {
	// Using options directly makes aspect change as soon as options are changed
	//return float32(s.cfg.Video.GameWidth) / float32(s.cfg.Video.GameHeight)
	return CalculateAspect(s.scrrect[2], s.scrrect[3])
}

func (s *System) getCurrentAspect() float32 {
	skip := s.skipMotifScaling()
	if (s.postMatchFlg && skip) || (s.middleOfMatch() && (skip || (!s.motif.me.active && !s.motif.di.active))) {
		return s.getFightAspect()
	}
	return s.getMotifAspect()
}

func (s *System) setGameSize(w, h int32) {
	s.scrrect[2], s.scrrect[3] = w, h

	// TODO: These ought to be system constants maybe
	baseWidth := int32(320)
	baseHeight := int32(240)

	screenAspect := CalculateAspect(w, h)
	targetAspect := CalculateAspect(baseWidth, baseHeight)

	if screenAspect > targetAspect {
		// Screen is wider than 4:3 - scale based on height
		s.gameWidth = float32(baseHeight) * screenAspect
		s.gameHeight = float32(baseHeight)
	} else {
		// Screen is taller than 4:3 - scale based on width
		s.gameWidth = float32(baseWidth)
		s.gameHeight = float32(baseWidth) / screenAspect
	}

	// Update scale
	s.widthScale = float32(s.scrrect[2]) / s.gameWidth
	s.heightScale = float32(s.scrrect[3]) / s.gameHeight
}

// Change aspect ratio at match start
func (s *System) applyFightAspect() {
	baseHeight := float32(240)
	var aspectGame float32

	// Select which aspect ratio to use
	if s.cfg.Video.FightAspectWidth < 0 && s.cfg.Video.FightAspectHeight < 0 {
		// Stage aspect
		// Get the next stage's localcoord
		// We need this branch because here we check next stage instead of current one like getFightAspect()
		var stageWidth, stageHeight int32
		if s.sel.selectedStageNo > 0 && s.sel.selectedStageNo <= len(s.sel.stagelist) {
			def := strings.ToLower(filepath.Base(s.sel.stagelist[s.sel.selectedStageNo-1].def))
			if coord, ok := s.stageLocalcoords[def]; ok && coord[0] > 0 && coord[1] > 0 {
				stageWidth = int32(coord[0])
				stageHeight = int32(coord[1])
			}
		}

		// Calculate the stage's aspect ratio
		if stageWidth > 0 && stageHeight > 0 {
			aspectGame = CalculateAspect(stageWidth, stageHeight)
		} else {
			// Fallback
			//aspectGame = CalculateAspect(s.cfg.Video.GameWidth, s.cfg.Video.GameHeight)
			aspectGame = s.getMotifAspect()
		}
	} else {
		aspectGame = s.getFightAspect()
	}

	// Compute new game dimensions while maintaining the same base height
	gameWidth := baseHeight * aspectGame
	s.gameWidth = gameWidth
	s.gameHeight = baseHeight

	// Scale to fit current screen size
	s.widthScale = float32(s.scrrect[2]) / float32(s.gameWidth)
	s.heightScale = float32(s.scrrect[3]) / float32(s.gameHeight)
}

func (s *System) captureAspectState() drawAspectState {
	return drawAspectState{
		gameWidth:   s.gameWidth,
		gameHeight:  s.gameHeight,
		widthScale:  s.widthScale,
		heightScale: s.heightScale,
	}
}

func (s *System) restoreAspectState(st drawAspectState) {
	s.gameWidth = st.gameWidth
	s.gameHeight = st.gameHeight
	s.widthScale = st.widthScale
	s.heightScale = st.heightScale
}

func (s *System) wrapDrawWithAspectState(fn func()) func() {
	if fn == nil {
		return nil
	}
	st := s.captureAspectState()
	return func() {
		prev := s.captureAspectState()
		s.restoreAspectState(st)
		defer s.restoreAspectState(prev)
		fn()
	}
}

func (s *System) shouldPersistMotifAspect() bool {
	return s.cfg.Video.KeepAspect && !s.skipMotifScaling()
}

func (s *System) enterMotifAspect() {
	if !s.shouldPersistMotifAspect() {
		return
	}
	s.setGameSize(s.scrrect[2], s.scrrect[3])
}

func (s *System) leaveMotifAspect() {
	if !s.shouldPersistMotifAspect() {
		return
	}
	s.applyFightAspect()
}

func (s *System) setGameAspect() {
	sys.applyFightAspect()
	if sys.stage != nil {
		sys.stage.localscl = float32(sys.gameWidth) / float32(sys.stage.stageCamera.localcoord[0])
		sys.stage.stageCamera.localscl = sys.stage.localscl
	}
	for _, p := range sys.chars {
		if len(p) > 0 {
			p[0].localcoord = float32(p[0].gi().localcoord[0]) / (float32(sys.gameWidth) / 320)
			p[0].localscl = 320 / p[0].localcoord
		}
	}
	sys.fightScreen.setScale()
}

func (s *System) eventUpdate() bool {
	s.esc = false
	for _, v := range s.shortcutScripts {
		v.Activate = false
	}
	s.window.pollEvents()
	s.gameEnd = s.window.shouldClose()
	return !s.gameEnd
}

func (s *System) runMainThreadTask() {
	for {
		select {
		case f := <-s.mainThreadTask:
			f()
		default:
			return
		}
	}
}

func (s *System) keepAlive() {
	//s.keepAliveProfile = true
	// Log elapsed time since previous keepAlive and where this call came from.
	if s.keepAliveProfile {
		now := time.Now()
		if s.keepAliveCount == 0 {
			s.keepAliveStart = now
		} else {
			delta := now.Sub(s.keepAlivePrev)
			total := now.Sub(s.keepAliveStart)
			_, file, line, ok := runtime.Caller(1)
			loc := "unknown"
			if ok {
				loc = fmt.Sprintf("%s:%d", filepath.Base(file), line)
			}
			LogMessage("[keepAlive] #%d Δ=%.3fms total=%.3fs at %s",
				s.keepAliveCount,
				float64(delta)/float64(time.Millisecond),
				float64(total)/float64(time.Second),
				loc)
		}
		s.keepAlivePrev = now
		s.keepAliveCount++
	}
	s.runMainThreadTask()
	s.window.pollEvents()
}

func (s *System) await(fps int) bool {
	if !s.frameSkip {
		// Render the finished frame
		gfx.EndFrame()
		if gfx.GetName()[:6] == "OpenGL" {
			s.window.SwapBuffers()
		} else {
			gfx.Await()
		}
		s.window.UpdateDebugFPS()
		if s.isTakingScreenshot {
			defer captureScreen()
			s.isTakingScreenshot = false
		}
		// Begin the next frame after events have been processed. Do not clear
		// the screen if network input is present.
		defer gfx.BeginFrame(sys.netConnection == nil)
	}

	s.runMainThreadTask()

	now := time.Now()
	diff := s.redrawWait.nextTime.Sub(now)

	var waitDuration time.Duration

	if s.rollback.session != nil {
		// Use rollback backend value
		waitDuration = s.rollback.session.loopTimer.usToWaitThisLoop()
	} else {
		// Use the given argument
		waitDuration = time.Second / time.Duration(fps)
	}

	// Increment the deadline
	s.redrawWait.nextTime = s.redrawWait.nextTime.Add(waitDuration)

	switch {
	case diff >= 0 && diff < waitDuration+2*time.Millisecond:
		time.Sleep(diff)
		fallthrough
	case now.Sub(s.redrawWait.lastDraw) > 250*time.Millisecond:
		fallthrough
	case diff >= -17*time.Millisecond:
		s.redrawWait.lastDraw = now
		s.frameSkip = false
	default:
		// If we are lagging behind by more than 150ms, don't try to speed up. Just reset the schedule to now
		if diff < -150*time.Millisecond {
			s.redrawWait.nextTime = now.Add(waitDuration)
		}
		s.frameSkip = true
	}

	s.eventUpdate()

	return !s.gameEnd
}

func (s *System) renderFrame() {
	if !s.frameSkip {
		x, y, scl := s.cam.Pos[0], s.cam.Pos[1], s.cam.Scale/s.cam.BaseScale()
		dx, dy, dscl := s.zoom.apply(x, y, scl)
		s.draw(dx, dy, dscl)
	}

	// Lua
	if !s.frameSkip {
		s.luaFlushDrawQueue()
	} else {
		// Keep pause-menu logic responsive even when this render frame is skipped.
		// Any queued draw ops are discarded below because this frame is not being rendered.
		if s.motif.me.active {
			s.motif.me.runLua(&s.motif)
		}
		// On skipped frames, discard queued draws to avoid buildup.
		s.luaDiscardDrawQueue()
	}

	// Render top elements
	if !s.frameSkip {
		s.drawTop()
	}

	// Render debug elements
	if !s.frameSkip && s.debugDisplay {
		s.drawDebugText()
	}
}

func (s *System) update() bool {
	s.frameCounter++

	// preMatchTime normally follows the local UI clock until gameplay begins.
	// Preserve the synchronized prematch time while rollback waits for its first frame.
	rollbackMatchStarting := s.rollback.session != nil && s.rollback.netConnection != nil
	if s.matchTime == 0 && !rollbackMatchStarting && s.replayFile == nil {
		s.preMatchTime = s.frameCounter
	}

	// Correct the joystick mappings (macOS)
	for i := 0; i < len(sys.joystickConfig); i++ {
		if runtime.GOOS == "darwin" && !sys.joystickConfig[i].isInitialized {
			joyS := i
			if joyS < len(sys.joystickConfig) {
				if input.IsJoystickPresent(joyS) {
					guid := input.GetJoystickGUID(joyS)

					// Correct the inner config
					if sys.joystickConfig[joyS].GUID != guid && !sys.joystickConfig[joyS].isInitialized {
						// Swap those that don't match
						for i := 0; i < len(sys.joystickConfig); i++ {
							if i != joyS && sys.joystickConfig[i].GUID == guid {
								sys.joystickConfig[joyS].swap(&sys.joystickConfig[i])
								logicalPlayerA := sys.joystickConfig[joyS].Joy
								logicalPlayerB := sys.joystickConfig[i].Joy
								sys.inputRemap[logicalPlayerA] = joyS
								sys.inputRemap[logicalPlayerB] = i
								// cs := *input.controllerstate[joyS]
								// *input.controllerstate[joyS] = *input.controllerstate[i]
								// *input.controllerstate[i] = cs
								// c := input.controllers[joyS]
								// input.controllers[joyS] = input.controllers[i]
								// input.controllers[i] = c
								// fmt.Printf("system.go: inputremap[%v] = %v, inputRemap[%v] = %v\n", joyS, sys.inputRemap[joyS], i, sys.inputRemap[i])
								// fmt.Printf("system.go: %v, Joy = %v, RealJoy = %v, GUID = %v\n", input.GetJoystickGUID(joyS), joyS, sys.joystickConfig[joyS].Joy, guid)
								break
							}
						}
					}
				}
			}
		}
	}

	if s.replayFile != nil {
		if s.anyHardButton() {
			s.await(s.gameRenderSpeed() * 4)
		} else {
			s.await(s.gameRenderSpeed())
		}
		return s.replayFile.Update()
	}

	if s.netConnection != nil {
		s.await(s.gameRenderSpeed())
		return s.netConnection.Update()
	}

	return s.await(s.gameRenderSpeed())
}

func (s *System) tickSound() {
	s.soundChannels.Tick()
	if !s.noCharSoundFlg {
		for i := range sys.charSoundChannels {
			sys.charSoundChannels[i].Tick()
		}
	}

	// Always pause if noMusic flag set, pause master volume is 0, or freqmul is 0.
	s.bgm.SetPaused(s.nomusic || (s.paused && s.cfg.Sound.PauseMasterVolume == 0) || (s.bgm.freqmul == 0))

	if s.paused {
		// Apply BGM pause volume once per pause, even when the original BGM volume is 0.
		// volRestore cannot be used as the latch because 0 is a valid volume.
		if !s.bgm.pauseVolumeApplied {
			s.bgm.volRestore = s.bgm.bgmVolume
			s.bgm.bgmVolume = int(s.cfg.Sound.PauseMasterVolume * s.bgm.bgmVolume / 100.0)
			s.bgm.UpdateVolume()
			s.bgm.pauseVolumeApplied = true
		}

		// Run every paused tick so sounds started while paused are also softened.
		s.pauseVolumeApplied = true
		s.softenAllSound()
	} else if s.pauseVolumeApplied || s.bgm.pauseVolumeApplied {
		s.restorePauseVolume()
	}
}

func (s *System) restorePauseVolume() {
	if s.bgm.pauseVolumeApplied {
		s.bgm.bgmVolume = s.bgm.volRestore
		s.bgm.volRestore = 0
		s.bgm.pauseVolumeApplied = false
		s.bgm.UpdateVolume()
	}
	if s.pauseVolumeApplied {
		s.restoreAllVolume()
		s.pauseVolumeApplied = false
	}
}

func (s *System) resetRemapInput() {
	for i := range s.inputRemap {
		s.inputRemap[i] = i
	}
}

func (s *System) loaderReset() {
	s.roundNo, s.wins, s.roundsExisted, s.decisiveRound = 1, [2]int32{}, [2]int32{}, [2]bool{}
	s.turnsPreloadMember = [2]int{-1, -1}
	s.loader.reset()
}

func (s *System) loadStart() {
	s.loaderReset()
	s.loader.runTread()
}

// Drop everything that might have been partially produced by the pre-match async loader.
func (s *System) dropCanceledLoadData() {
	for {
		select {
		case <-s.mainThreadTask:
		default:
			goto drained
		}
	}
drained:
	s.loadMutex.Lock()
	defer s.loadMutex.Unlock()

	for prefix, ffx := range s.ffx {
		if ffx == nil || !ffx.isCharFX {
			continue
		}
		ffx.sff = nil
		delete(s.ffx, prefix)
	}

	for i := range s.cgi {
		s.cgi[i].sff = nil
		s.cgi[i].palettedata = nil
		s.cgi[i].snd = nil
		s.cgi[i].states = nil
		s.chars[i] = nil
		s.cgi[i].palno = -1
	}

	if s.stage != nil {
		s.stage.destroy()
		s.stage = nil
	}
	s.stageList = make(map[int32]*Stage)
	s.stageLoop = false
}

func (s *System) loadCancel() {
	s.loader.reset()
	s.dropCanceledLoadData()
	if s.netConnection != nil {
		s.netConnection.ResetLoadingPhase()
	}
}

func (s *System) synchronize() error {
	if s.replayFile != nil {
		s.replayFile.Synchronize()
	} else if s.netConnection != nil {
		return s.netConnection.Synchronize()
	}
	return nil
}

/*
func (s *System) anyHardButton() bool {
	for _, kc := range s.keyConfig {
		if kc.a() || kc.b() || kc.c() || kc.x() || kc.y() || kc.z() {
			return true
		}
	}
	for _, kc := range s.joystickConfig {
		if kc.a() || kc.b() || kc.c() || kc.x() || kc.y() || kc.z() {
			return true
		}
	}
	return false
}
*/

// Joysticks were already refactored to be polled less times, but having these functions still makes them be polled twice as often during intros/outros
// We're already polling them about 10 times less so that should be enough anyway
// In Mugen, intro/outro skipping only happens on button press, not button hold
func (s *System) anyHardButton() bool {
	// Button indices for a, b, c, x, y, z
	hardButtonIdx := []int{4, 5, 6, 7, 8, 9}

	for _, kc := range s.keyConfig {
		buttons := GetKeyboardState(kc)
		for _, idx := range hardButtonIdx {
			if buttons[idx] {
				return true
			}
		}
	}

	for _, kc := range s.joystickConfig {
		buttons := GetJoystickState(kc)
		for _, idx := range hardButtonIdx {
			if buttons[idx] {
				return true
			}
		}
	}

	return false
}

func (s *System) anyButton() bool {
	if s.replayFile != nil {
		return s.replayFile.AnyButton()
	}
	if s.netConnection != nil {
		return s.netConnection.AnyButton()
	}
	if s.rollback.session != nil {
		return s.rollback.anyButton()
	}
	return s.anyHardButton()
}

func (s *System) buttonController(btns []string) int {
	for _, btn := range btns {
		for i, cl := range s.commandLists {
			if cl != nil && cl.GetState(btn) {
				return i
			}
		}
	}
	return -1
}

func (s *System) stepCommandLists() {
	for i, cl := range s.commandLists {
		if cl == nil || cl.Buffer == nil {
			continue
		}
		// Step commands only if the buffer has already stepped. Prevents rapid fire inputs
		if cl.InputUpdate(nil, i) {
			cl.Step(false, false, false, false, 0)
		}
	}
}

// Avoid immediate UI token reuse right after defaults/config/player changes.
func (s *System) uiResetTokenGuard() {
	s.uiConsumeInputFrame = s.frameCounter + 1
	s.uiLastInputToken = ""
	s.lastInputController = -1
}

func (s *System) uiRepeatShouldFire(held int32) bool {
	// held: 1 = pressed this frame, 2+ = held
	if held <= 1 {
		return false
	}
	// Wait for initial delay, then repeat every UiRepeatRate frames.
	if held < 1+s.cfg.Input.UiRepeatDelay {
		return false
	}
	return (held-(1+s.cfg.Input.UiRepeatDelay))%s.cfg.Input.UiRepeatRate == 0
}

func (s *System) uiControllerKey(controllerIdx int) int {
	if controllerIdx < 0 {
		return controllerIdx
	}
	if controllerIdx < len(s.inputRemap) && s.inputRemap[controllerIdx] >= 0 {
		return s.inputRemap[controllerIdx]
	}
	return controllerIdx
}

func (s *System) uiConsumeTokenThisFrame(controller int, tok string) bool {
	if s.uiRepeatFrame == s.frameCounter && s.uiRepeatController == controller && s.uiRepeatToken == tok {
		return true
	}
	s.uiRepeatFrame = s.frameCounter
	s.uiRepeatController = controller
	s.uiRepeatToken = tok
	return false
}

// Checks the per-controller raw input buffer (InputBuffer) instead of command states.
// Returns true if any of the provided keys were pressed this frame (buffer time == 1).
// Supported keys: B, D, F, U, L, R, a, b, c, x, y, z, s, d, w, m
// D/U/B/F also recognize LS axis tokens: D -> LS_Y+, U -> LS_Y-, B -> LS_X-, F -> LS_X+.
func (s *System) uiRawInput(btns []string, controllerIdx int) bool {
	checkOne := func(i int, cl *CommandList) bool {
		if cl == nil || cl.Buffer == nil {
			return false
		}
		controllerKey := s.uiControllerKey(i)

		// True for raw controller axis tokens (LS_*, RS_*, LT/RT).
		// Used when a screenpack uses e.g. getInput(pn, "LS_Y+") directly.
		isAxisToken := func(tok string) bool {
			idx, ok := StringToButtonLUT[tok]
			return ok && idx != 25 && idx >= 15 && idx <= 24
		}

		// Repeat only for navigation-style inputs (directions).
		repeatable := func(tok string) bool {
			switch tok {
			case "U", "D", "L", "R", "B", "F":
				return true
			default:
				return false
			}
		}

		ib := cl.Buffer
		for _, raw := range btns {
			raw = strings.TrimSpace(raw)
			if raw == "" {
				continue
			}
			// Canonicalize UI stick-direction tokens to the digital directions.
			tok := raw
			switch tok {
			case "LS_Y+":
				tok = "D"
			case "LS_Y-":
				tok = "U"
			case "LS_X-":
				tok = "B"
			case "LS_X+":
				tok = "F"
			}
			// Digital/raw-buffer check (pressed this frame == 1) + UI repeat while held
			trigger := false
			switch tok {
			case "B":
				trigger = ib.Bb == 1 || (repeatable(tok) && s.uiRepeatShouldFire(int32(ib.Bb)))
			case "D":
				trigger = ib.Db == 1 || (repeatable(tok) && s.uiRepeatShouldFire(int32(ib.Db)))
			case "F":
				trigger = ib.Fb == 1 || (repeatable(tok) && s.uiRepeatShouldFire(int32(ib.Fb)))
			case "U":
				trigger = ib.Ub == 1 || (repeatable(tok) && s.uiRepeatShouldFire(int32(ib.Ub)))
			case "L":
				trigger = ib.Lb == 1 || (repeatable(tok) && s.uiRepeatShouldFire(int32(ib.Lb)))
			case "R":
				trigger = ib.Rb == 1 || (repeatable(tok) && s.uiRepeatShouldFire(int32(ib.Rb)))
			case "a":
				trigger = ib.ab == 1
			case "b":
				trigger = ib.bb == 1
			case "c":
				trigger = ib.cb == 1
			case "x":
				trigger = ib.xb == 1
			case "y":
				trigger = ib.yb == 1
			case "z":
				trigger = ib.zb == 1
			case "s":
				trigger = ib.sb == 1
			case "d":
				trigger = ib.db == 1
			case "w":
				trigger = ib.wb == 1
			case "m":
				trigger = ib.mb == 1
			}
			// Raw axis tokens (LS_*/RS_*/LT/RT) when used directly in Lua.
			// LS axis directions are handled above as D/U/B/F aliases.
			// Other axis tokens are treated as "held" here.
			if !trigger && isAxisToken(tok) && cl.IsControllerButtonPressed(tok, i) {
				trigger = true
			}
			if trigger {
				if s.uiConsumeInputFrame != 0 && s.frameCounter == s.uiConsumeInputFrame {
					continue
				}
				// Guard: don't fire same token more than once per frame (repeat-safe).
				if s.uiConsumeTokenThisFrame(controllerKey, tok) {
					continue
				}
				s.lastInputController = controllerKey
				s.uiLastInputToken = tok
				return true
			}
		}
		return false
	}
	if controllerIdx >= 0 {
		if controllerIdx < len(s.commandLists) {
			return checkOne(controllerIdx, s.commandLists[controllerIdx])
		}
		return false
	}
	//found := false
	for i := 0; i < len(s.commandLists); i++ {
		if checkOne(i, s.commandLists[i]) {
			//found = true
			return true
		}
	}
	//return found
	return false
}

func (s *System) uiIsKeyAction(name string) bool {
	switch name {
	case "up", "down", "left", "right",
		"a", "b", "c", "x", "y", "z",
		"start", "d", "w", "menu":
		return true
	default:
		return false
	}
}

func (s *System) uiDefaultTokens(pn int, joystick bool) []string {
	defaultNoKey := s.cfg.DefaultOnlyIni.Section("Keys_P3").Key("menu").String()
	sectionName := "Keys_P3"
	if pn == 1 || pn == 2 {
		sectionName = "Keys_P" + Itoa(pn)
	}
	if joystick {
		sectionName = "Joystick_P1"
	}
	get := func(k string) string {
		v := s.cfg.DefaultOnlyIni.Section(sectionName).Key(k).String()
		if v == defaultNoKey {
			return s.motif.OptionInfo.Menu.Valuename["nokey"]
		}
		return v
	}
	return []string{
		get("up"), get("down"), get("left"), get("right"),
		get("a"), get("b"), get("c"),
		get("x"), get("y"), get("z"),
		get("start"), get("d"), get("w"), get("menu"),
	}
}

func (s *System) uiSetConfigDefaults(pn int, joystick bool, enabled map[string]bool) {
	if pn < 1 {
		return
	}
	var (
		kc     *KeyConfig
		base   string
		joyVal int
		toInt  func(string) int
	)
	if joystick {
		if pn > len(s.joystickConfig) {
			return
		}
		joyVal = pn - 1
		kc = &s.joystickConfig[pn-1]
		base = fmt.Sprintf("Joystick_P%d.", pn)
		toInt = func(tok string) int { return int(StringToButtonLUT[tok]) }
	} else {
		if pn > len(s.keyConfig) {
			return
		}
		joyVal = -1
		kc = &s.keyConfig[pn-1]
		base = fmt.Sprintf("Keys_P%d.", pn)
		toInt = func(tok string) int { return int(StringToKey(tok)) }
	}
	kc.Joy = joyVal
	kc.GUID = ""
	kc.rumbleOn = false
	// Build & apply tokens (respect enabled set).
	def := s.uiDefaultTokens(pn, joystick)
	var uiPlayerFields = [...]string{
		"up", "down", "left", "right",
		"a", "b", "c",
		"x", "y", "z",
		"start", "d", "w", "menu",
	}
	var v [len(uiPlayerFields)]int
	for i, f := range uiPlayerFields {
		tok := def[i]
		if enabled != nil && !enabled[f] {
			tok = s.motif.OptionInfo.Menu.Valuename["nokey"]
		}
		v[i] = toInt(tok)
		_ = s.cfg.SetValueUpdate(base+f, tok)
	}
	kc.set(v)
	_ = s.cfg.SetValueUpdate(base+"Joystick", joyVal)
	_ = s.cfg.SetValueUpdate(base+"GUID", "")
	if joystick {
		_ = s.cfg.SetValueUpdate(base+"RumbleOn", false)
	}
	s.uiResetTokenGuard()
}

func (s *System) uiRegisterCommand(name string, spec CommandSpec) error {
	name = strings.TrimSpace(name)
	spec.Cmd = strings.TrimSpace(spec.Cmd)
	if name == "" || spec.Cmd == "" {
		return nil
	}
	if s.uiCommandRegistry == nil {
		s.uiCommandRegistry = make(map[string]CommandSpec, 32)
	}
	// Already registered -> no-op.
	if _, ok := s.uiCommandRegistry[name]; ok {
		return nil
	}
	s.uiCommandRegistry[name] = spec
	// Add to existing UI command lists (skip if already present).
	for _, cl := range s.commandLists {
		if cl == nil {
			continue
		}
		if _, exists := cl.Names[name]; exists {
			continue
		}
		if err := cl.AddCommand(name, spec); err != nil {
			return err
		}
	}
	return nil
}

func (s *System) uiApplyCommandRegistry(cl *CommandList) error {
	if cl == nil || len(s.uiCommandRegistry) == 0 {
		return nil
	}
	for name, spec := range s.uiCommandRegistry {
		if _, exists := cl.Names[name]; exists {
			continue
		}
		if err := cl.AddCommand(name, spec); err != nil {
			return err
		}
	}
	return nil
}

func (s *System) uiEnsureCommandLists(total int) error {
	// Ensure UI commands exist.
	def := []string{"D", "U", "L", "R", "a", "b", "c", "x", "y", "z", "s", "d", "w", "m", "/s"}
	dcl := NewCommandList(nil)
	defSpec := CommandSpec{
		Time:           dcl.DefaultTime,
		BufTime:        dcl.DefaultBufferTime,
		BufferHitpause: dcl.DefaultBufferHitpause,
		BufferPauseend: dcl.DefaultBufferPauseEnd,
		StepTime:       dcl.DefaultStepTime,
	}
	for _, k := range def {
		spec := defSpec
		spec.Cmd = k
		if err := s.uiRegisterCommand(k, spec); err != nil {
			LogMessage("uiSeedDefaultCommands: %q: %v", k, err)
		}
	}
	// Trim if needed.
	if total < len(s.commandLists) {
		s.commandLists = s.commandLists[:total]
	}
	// Grow.
	for len(s.commandLists) < total {
		cl := NewCommandList(NewInputBuffer())
		if err := s.uiApplyCommandRegistry(cl); err != nil {
			return fmt.Errorf("command registry apply failed: %w", err)
		}
		s.commandLists = append(s.commandLists, cl)
	}
	// Repair nil/bufferless entries.
	for i := 0; i < total; i++ {
		if s.commandLists[i] == nil || s.commandLists[i].Buffer == nil {
			cl := NewCommandList(NewInputBuffer())
			if err := s.uiApplyCommandRegistry(cl); err != nil {
				return fmt.Errorf("command registry apply failed: %w", err)
			}
			s.commandLists[i] = cl
		}
	}
	return nil
}

func (s *System) netplay() bool {
	return s.rollback.session != nil || s.netConnection != nil || s.replayFile != nil
}

func (s *System) escExit() bool {
	return s.esc && (!s.sel.gameParams.PauseMenu || !s.cfg.Config.EscOpensMenu ||
		(s.motif.AttractMode.Enabled && s.credits == 0))
}

func (s *System) matchOver() bool {
	// We must check if wins are greater than 0 because modes like Training may have "0 rounds to win"
	return s.wins[0] > 0 && s.wins[0] >= s.matchWins[0] ||
		s.wins[1] > 0 && s.wins[1] >= s.matchWins[1]
}

func (s *System) anyChar() *Char {
	for i := range s.chars {
		for j := range s.chars[i] {
			if s.chars[i][j] != nil {
				return s.chars[i][j]
			}
		}
	}
	return nil
}

func (s *System) playerID(id int32) *Char {
	if id < 0 {
		return nil
	}

	// Invalid ID
	c, ok := s.charList.idMap[id]
	if !ok {
		return nil
	}

	// Mugen skips destroyed helpers and disabled players here
	// Note: This will also make them be skipped in the engine's internal ID lookups
	if c.csf(CSF_destroy) || c.scf(SCF_disabled) {
		return nil
	}

	return c
}

func (s *System) playerIDExist(id BytecodeValue) BytecodeValue {
	if id.IsUndefined() {
		return BytecodeUndefined()
	}
	c := s.playerID(id.ToI())
	return BytecodeBool(c != nil)
}

func (s *System) playerIndex(idx int32) *Char {
	if idx < 0 || int(idx) >= len(sys.charList.runOrder) {
		return nil
	}

	// We will also ignore destroyed/disabled players here, like Mugen redirections do
	var searchIdx int32
	for _, c := range sys.charList.creationOrder {
		if c != nil && !c.csf(CSF_destroy) && !c.scf(SCF_disabled) {
			if searchIdx == idx {
				return c
			}
			searchIdx++
		}
	}

	//if idx >= 0 && int(idx) < len(s.charList.runOrder) {
	//	return s.charList.runOrder[idx]
	//}

	return nil
}

// TODO: This is redundant since the index always exists if "NumPlayer >= idx-1"
// Maybe remove it or make it ignore destroyed helpers at least
func (s *System) playerIndexExist(idx BytecodeValue) BytecodeValue {
	if idx.IsUndefined() {
		return BytecodeUndefined()
	}
	char := s.playerIndex(idx.ToI())
	return BytecodeBool(char != nil)
}

// For player number triggers
func (s *System) getCharRoot(idx int) *Char {
	if idx < 0 || idx >= len(sys.chars) {
		return nil
	}
	p := sys.chars[idx]
	if len(p) == 0 || p[0] == nil || p[0].scf(SCF_disabled) {
		return nil
	}
	return p[0]
}

func (s *System) playerNoExist(pn BytecodeValue) BytecodeValue {
	if pn.IsUndefined() {
		return BytecodeUndefined()
	}
	idx := int(pn.ToI() - 1)
	c := s.getCharRoot(idx)
	return BytecodeBool(c != nil)
}

func (s *System) palfxvar(x int32, y int32) int32 {
	n := int32(0)
	if x >= 4 {
		n = 256
	}
	pfx := s.bgPalFX
	if y == 2 {
		pfx = s.allPalFX
	}
	if pfx.enable {
		switch x {
		case -2:
			n = pfx.eInvertblend
		case -1:
			n = Btoi(pfx.eInvertall)
		case 0:
			n = pfx.time
		case 1:
			n = pfx.eAdd[0]
		case 2:
			n = pfx.eAdd[1]
		case 3:
			n = pfx.eAdd[2]
		case 4:
			n = pfx.eMul[0]
		case 5:
			n = pfx.eMul[1]
		case 6:
			n = pfx.eMul[2]
		default:
			n = 0
		}
	}
	return n
}

func (s *System) palfxvar2(x int32, y int32) float32 {
	n := float32(1)
	if x > 1 {
		n = 0
	}
	pfx := s.bgPalFX
	if y == 2 {
		pfx = s.allPalFX
	}
	if pfx.enable {
		switch x {
		case 1:
			n = pfx.eColor
		case 2:
			n = pfx.eHue
		default:
			n = 0
		}
	}
	return n * 256
}

// Only Lua uses these currently
func (s *System) screenHeight() float32 {
	//return 240
	return float32(s.gameHeight)
}

func (s *System) screenWidth() float32 {
	return float32(s.gameWidth)
}

func (s *System) roundEnded() bool {
	return s.intro < -s.fightScreen.round.over_hittime
}

// Characters cannot hurt each other between fight screen timers over.hittime and over.waittime
func (s *System) roundNoDamage() bool {
	return sys.intro < 0 && sys.intro <= -sys.fightScreen.round.over_hittime && sys.intro >= -sys.fightScreen.round.over_waittime
}

// Gametime is the sum of the match time and the screenpack time
func (s *System) gameTime() int32 {
	// Select the appropriate offset
	var pmTime int32
	if sys.netConnection != nil {
		pmTime = sys.netConnection.preMatchTime
	} else if sys.replayFile != nil {
		pmTime = sys.replayFile.preMatchTime
	} else {
		pmTime = sys.preMatchTime
	}

	return sys.matchTime + pmTime
}

// In Mugen, RoundState 2 begins as soon as the "Fight" screen appears, before players have control
// That causes more harm than good and is not clearly stated in the documentation, so Ikemen changes it
func (s *System) roundState() int32 {
	switch {
	case sys.intro > sys.fightScreen.round.ctrl_time+1 || sys.postMatchFlg:
		return 0
	//case sys.fightScreen.round.current == 0:
	case sys.intro > 0:
		return 1
	//case sys.intro >= 0 || sys.finishType == FT_NotYet:
	case sys.intro == 0 || sys.finishType == FT_NotYet:
		return 2
	case sys.intro < -sys.fightScreen.round.over_waittime:
		return 4
	default:
		return 3
	}
}

func (s *System) introState() int32 {
	switch {
	case s.intro > s.fightScreen.round.ctrl_time+1:
		// Pre-intro [RoundState = 0]
		return 1
	case (s.motif.di.active && s.dialogueForce == 0) || s.intro == s.fightScreen.round.ctrl_time+1:
		// Player intros [RoundState = 1]
		return 2
	case s.intro == s.fightScreen.round.ctrl_time:
		// Dialogue detection (s.motif.di.active is detectable 1 frame later)
		if s.motif.isDialogueSet() {
			return 2
		}
		// Round announcement
		return 3
	case s.intro > 0 && s.intro < s.fightScreen.round.ctrl_time:
		// Fight called
		// Used to check both these conditions, but refactors revealed that was probably wrong
		//s.fightScreen.round.fight_timing.animTimer == -1 || (s.intro > 0 && s.intro < s.fightScreen.round.ctrl_time)
		return 4
	default:
		// Not applicable
		return 0
	}
}

func (s *System) outroState() int32 {
	switch {
	case s.intro >= 0:
		// Not applicable
		return 0
	case s.roundOver():
		// Round over
		return 5
	case s.winposetime <= 0:
		// Player win states
		return 4
	case sys.intro <= -sys.fightScreen.round.over_waittime && sys.winposetime > 0:
		// Players lose control, but the round has not yet entered win states
		return 3
	case s.intro < -s.fightScreen.round.over_hittime || sys.fightScreen.round.over_hittime == 1:
		// Players still have control, but the match outcome can no longer be changed
		return 2
	case s.intro < 0:
		// Payers can still act, allowing a possible double KO
		return 1
	default:
		// Fallback case, shouldn't be reached
		return 0
	}
}

func (s *System) roundOver() bool {
	return s.intro < -(s.fightScreen.round.over_waittime + s.fightScreen.round.overTime())
}

func (s *System) roundStateTicks() int32 {
	return s.intro + s.fightScreen.round.over_waittime + s.fightScreen.round.overTime()
}

// Check if the match consists of a single round
func (s *System) roundIsSingle() bool {
	return !s.sel.gameParams.PersistRounds && s.roundNo == 1 && s.decisiveRound[0] && s.decisiveRound[1]
}

func (s *System) maxDrawsReached(team int) bool {
	limit := s.maxDraws[team]
	return limit >= 0 && s.draws >= limit
}

// This checks if a round is eligible for "Final Round" behavior, not if it's literally the final round
// https://github.com/ikemen-engine/Ikemen-GO/issues/1659
func (s *System) roundIsFinal() bool {
	if s.sel.gameParams.PersistRounds {
		return false
	}
	// The first round is never already final
	if s.roundNo <= 1 {
		return false
	}
	// Both teams must be on their decisive round
	if !s.decisiveRound[0] || !s.decisiveRound[1] {
		return false
	}
	// Both teams must have reached their maximum draw limits
	return s.maxDrawsReached(0) && s.maxDrawsReached(1)
}

func (s *System) winnerTeam() int32 {
	var winp int32 = -1
	if !s.endMatch {
		if s.matchOver() && s.roundOver() {
			w1 := s.wins[0] >= s.matchWins[0]
			w2 := s.wins[1] >= s.matchWins[1]
			if w1 != w2 {
				winp = Btoi(w1) + Btoi(w2)*2
			} else {
				winp = 0
			}
		} else if s.winTeam >= 0 || s.roundState() >= 3 {
			winp = int32(s.winTeam) + 1
		}
	}
	return winp
}

func (s *System) gsf(gsf GlobalSpecialFlag) bool {
	return s.specialFlag&gsf != 0
}

func (s *System) setGSF(gsf GlobalSpecialFlag) {
	s.specialFlag |= gsf
}

func (s *System) unsetGSF(gsf GlobalSpecialFlag) {
	s.specialFlag &^= gsf
}

func (s *System) appendToConsole(str string) {
	s.consoleText = append(s.consoleText, str)
	if len(s.consoleText) > s.cfg.Debug.ConsoleRows {
		s.consoleText = s.consoleText[len(s.consoleText)-s.cfg.Debug.ConsoleRows:]
	}
}

func (s *System) printToConsole(pn, sn int, a ...interface{}) {
	spl := s.stringPool[pn].List
	if sn >= 0 && sn < len(spl) {
		for _, str := range strings.Split(OldSprintf(spl[sn], a...), "\n") {
			fmt.Printf("%s\n", str)
			s.appendToConsole(str)
		}
	}
}

// Lua-driven deferred draw queue
// Used by embedded Lua helpers (animDraw/bgDraw/textImgDraw/etc).
func (s *System) luaQueuePreDraw(fn func()) {
	if fn == nil {
		return
	}
	fn = s.wrapDrawWithAspectState(fn)
	s.luaDrawPreOps = append(s.luaDrawPreOps, fn)
}
func (s *System) luaQueueLayerDraw(layer int, fn func()) {
	if fn == nil {
		return
	}
	fn = s.wrapDrawWithAspectState(fn)
	// Negative layers behave like a "pre" pass (e.g. clearColor).
	if layer < 0 {
		s.luaQueuePreDraw(fn)
		return
	}
	// Clamp to last known Lua layer bucket.
	if layer >= len(s.luaDrawLayerOps) {
		layer = len(s.luaDrawLayerOps) - 1
	}
	s.luaDrawLayerOps[layer] = append(s.luaDrawLayerOps[layer], fn)
}

func (s *System) luaFlushDrawQueue() {
	// Pre-pass
	for _, fn := range s.luaDrawPreOps {
		fn()
	}
	s.luaDrawPreOps = s.luaDrawPreOps[:0]
	// Layered passes
	for i := range s.luaDrawLayerOps {
		for _, fn := range s.luaDrawLayerOps[i] {
			fn()
		}
		s.luaDrawLayerOps[i] = s.luaDrawLayerOps[i][:0]
	}
}
func (s *System) luaDiscardDrawQueue() {
	s.luaDrawPreOps = s.luaDrawPreOps[:0]
	for i := range s.luaDrawLayerOps {
		s.luaDrawLayerOps[i] = s.luaDrawLayerOps[i][:0]
	}
}

// Print an error directly from bytecode.go
// Printing from char.go is preferable, but not always possible
// Most of them come from invalid math operations
func (s *System) printBytecodeError(str string) {
	if s.loader.state == LS_Complete && s.workingChar != nil {
		// Print during matches
		s.appendToConsole(sys.workingChar.warn() + str)
	} else if !sys.ignoreMostErrors {
		// Print outside matches (compiling)
		LogMessage(str)
	}
}

func (s *System) loadTime(start time.Time, str string, shell, console bool) {
	elapsed := time.Since(start)
	str = fmt.Sprintf("%v; Load time: %v", str, elapsed)
	if shell {
		fmt.Printf("%s\n", str)
	}
	if console {
		s.appendToConsole(str)
	}
}

// Update Z scale
// TODO: See if this still works correctly with Winmugen stages that scaled chars with Z
func (s *System) updateZScale(pos, localscale float32) float32 {
	topz := sys.stage.stageCamera.topz / localscale
	botz := sys.stage.stageCamera.botz / localscale
	scale := float32(1)
	if topz != botz {
		ztopscale, zbotscale := sys.stage.stageCamera.ztopscale, sys.stage.stageCamera.zbotscale
		d := (pos - topz) / (botz - topz)
		scale = ztopscale + d*(zbotscale-ztopscale)
		if scale <= 0 {
			scale = 0
		}
	}
	return scale
}

func (s *System) zEnabled() bool {
	return s.zmin != s.zmax
}

// Convert X and Y drawing position to Z perspective
func (s *System) drawposXYfromZ(inpos [2]float32, localscl, zpos, zscale float32) (outpos [2]float32) {
	outpos[0] = (inpos[0]-s.cam.Pos[0])*zscale + s.cam.Pos[0]
	outpos[1] = inpos[1] * zscale
	outpos[1] += s.posZtoYoffset(zpos, localscl) // "Z" position
	return
}

// Convert Z logic position to Y drawing offset
// This is separate from the above because shadows only need this part
func (s *System) posZtoYoffset(zpos, localscl float32) float32 {
	return zpos * localscl * s.stage.stageCamera.depthtoscreen
}

// Z axis check
// Changed to no longer check z enable constant, depends on stage now
func (s *System) zAxisOverlap(posz1, top1, bot1, localscl1, posz2, top2, bot2, localscl2 float32) bool {
	if s.zEnabled() {
		if (posz1+bot1)*localscl1 < (posz2-top2)*localscl2 ||
			(posz1-top1)*localscl1 > (posz2+bot2)*localscl2 {
			return false
		}
	}
	return true
}

func (s *System) clsnOverlap(clsn1 [][4]float32, scl1, pos1 [2]float32, facing1 float32, angle1 float32,
	clsn2 [][4]float32, scl2, pos2 [2]float32, facing2 float32, angle2 float32) bool {

	// Skip function if any boxes are missing
	if clsn1 == nil || clsn2 == nil {
		return false
	}
	anface1 := facing1
	anface2 := facing2

	// Flip boxes if scale < 0
	if scl1[0] < 0 {
		facing1 *= -1
		scl1[0] *= -1
	}
	if scl2[0] < 0 {
		facing2 *= -1
		scl2[0] *= -1
	}

	// Loop through first set of boxes
	for i := 0; i < len(clsn1); i++ {
		// Calculate positions
		l1 := clsn1[i][0]
		r1 := clsn1[i][2]
		if facing1 < 0 {
			l1, r1 = -r1, -l1
		}
		left1 := l1 * scl1[0]
		right1 := r1 * scl1[0]
		top1 := clsn1[i][1] * scl1[1]
		bottom1 := clsn1[i][3] * scl1[1]

		// Loop through second set of boxes
		for j := 0; j < len(clsn2); j++ {
			// Calculate positions
			l2 := clsn2[j][0]
			r2 := clsn2[j][2]
			if facing2 < 0 {
				l2, r2 = -r2, -l2
			}
			left2 := l2 * scl2[0]
			right2 := r2 * scl2[0]
			top2 := clsn2[j][1] * scl2[1]
			bottom2 := clsn2[j][3] * scl2[1]

			// Check for overlap
			if angle1 != 0 || angle2 != 0 {
				if RectIntersect(left1+pos1[0], top1+pos1[1], right1-left1, bottom1-top1,
					left2+pos2[0], top2+pos2[1], right2-left2, bottom2-top2, pos1[0], pos1[1], pos2[0], pos2[1],
					-Rad(angle1*anface1), -Rad(angle2*anface2)) {
					return true
				}
			} else {
				if left1+pos1[0] <= right2+pos2[0] &&
					left2+pos2[0] <= right1+pos1[0] &&
					top1+pos1[1] <= bottom2+pos2[1] &&
					top2+pos2[1] <= bottom1+pos1[1] {
					return true
				}
			}
		}
	}

	return false
}

// Assign starting player ID's in a way similar to Mugen
// This isn't strictly necessary but might improve backward compatibility
// TODO: We may be going through too much work for nothing here. A natural ID order of each loaded player having "ID + 1" would probably be better
func (s *System) initPlayerID() {
	// Assign a new player ID only if needed
	assignID := func(i int) {
		if i < 0 || i >= len(sys.chars) || len(sys.chars[i]) == 0 || sys.chars[i][0] == nil {
			return
		}

		c := sys.chars[i][0]

		if sys.roundNo == 1 || c.roundsExisted() == 0 {
			// In BG-loaded Turns, promoted fighters already have a valid ID
			if sys.cfg.Config.TurnsLoading && sys.tmode[i&1] == TM_Turns && c.id >= 0 {
				return
			}
			c.id = sys.newCharId()
		}
	}

	// Free some ID's in subsequent rounds
	if sys.roundNo > 1 {
		sys.pruneCharId()
	}

	// Odd player number ID's
	for i := 0; i < MaxSimul*2; i += 2 {
		assignID(i)
	}

	// Even player number ID's
	for i := 1; i < MaxSimul*2; i += 2 {
		assignID(i)
	}

	// Extra player ID's
	for i := MaxSimul * 2; i < MaxPlayerNo; i++ {
		assignID(i)
	}
}

// Prune player ID's above the last active root ID
// Mugen doesn't do this, but it avoids having to work with very high ID's in later rounds
func (s *System) pruneCharId() {
	s.lastCharId = Max(0, sys.cfg.Config.HelperMax-1)

	// At this point we're still checking the previous round's player ID's
	// However that works out well for Turns mode because it means a joining player won't reuse the ID of a defeated player
	for _, p := range s.chars {
		if len(p) > 0 && p[0] != nil && p[0].id > s.lastCharId {
			s.lastCharId = p[0].id
		}
	}
}

// Determine the next available character ID while keeping track of the last number used
func (s *System) newCharId() int32 {
	newid := s.lastCharId + 1

	// Check if the next ID is already being used
	// This is needed because helpers may be preserved between rounds
	// Directly check the idMap as it is the source of truth for player ID's
	for {
		if _, conflict := s.charList.idMap[newid]; !conflict {
			break
		}
		newid++
	}

	/*
		// This method scanned sys.chars instead but that's worse
		for {
			conflict := false
			for _, p := range s.chars {
				for _, c := range p {
					if c != nil && c.id == newid && !c.csf(CSF_destroy) {
						// Note: We only recycle destroyed helper ID's because the ID's refresh each round, unlike Mugen
						conflict = true
						newid++
						break
					}
				}
				if conflict {
					break
				}
			}
			if !conflict {
				break
			}
		}
	*/

	s.lastCharId = newid
	return newid
}

func (s *System) resetGblEffect() {
	s.allPalFX.clear()
	s.bgPalFX.clear()
	s.envShake.clear()
	s.zoom.reset()
	s.pausetime, s.pausetimebuffer = 0, 0
	s.supertime, s.supertimebuffer = 0, 0
	s.envcol_time = 0
	s.specialFlag = 0
}

// Hard reset. Used between rounds
func (s *System) clearAllCharSounds() {
	for i := range s.charSoundChannels {
		s.charSoundChannels[i].SetSize(0)
	}
}

// Soft reset. Used during gameplay
func (s *System) stopAllCharSounds() {
	for i := range s.charSoundChannels {
		s.charSoundChannels[i].StopAll()
	}
}

func (s *System) softenAllSound() {
	for i := range s.charSoundChannels {
		for j := range s.charSoundChannels[i] {
			ch := &s.charSoundChannels[i][j]

			// Temporarily store the volume so it can be recalled later.
			if ch.IsPlaying() && ch.sfx != nil && ch.ctrl != nil && !ch.pauseVolumeApplied {
				ch.volResume = ch.sfx.volume
				softVolume := ch.sfx.volume * (float32(s.cfg.Sound.PauseMasterVolume) / 100.0)
				ch.SetVolume(softVolume)
				ch.pauseVolumeApplied = true

				// Pause if pause master volume is 0
				if s.cfg.Sound.PauseMasterVolume == 0 {
					ch.SetPaused(true)
				}
			}
		}
	}
	// Don't pause motif sounds
}

func (s *System) restoreAllVolume() {
	for i := range s.charSoundChannels {
		for j := range s.charSoundChannels[i] {
			ch := &s.charSoundChannels[i][j]

			// Restore the volume we had.
			if ch.sfx != nil && ch.ctrl != nil && ch.pauseVolumeApplied {
				ch.SetVolume(ch.volResume)
				ch.pauseVolumeApplied = false

				// Unpause only those whose freqmul > 0
				if ch.ctrl.Paused && ch.sfx.freqmul > 0 {
					ch.SetPaused(false)
				}
			}
		}
	}
}

func (s *System) clearMatchSound() {
	s.clearAllCharSounds()
	// Quiesce stage videos so no background decoding continues while mixer is empty,
	// and mark them as detached so SetPlaying(true) can re-attach next frame.
	if s.stage != nil {
		for _, b := range s.stage.bg {
			if b != nil && b._type == BG_Video {
				b.video.SetPlaying(false)
				b.video.SetVisible(false)
				b.video.MixerCleared()
			}
		}
	}
}

func (s *System) clearAllSound() {
	s.soundChannels.StopAll()
	s.soundMixer.Clear()
	s.clearMatchSound()
}

// Remove the player's explods, projectiles and (optionally) helpers as well as stopping their sounds
func (s *System) clearPlayerAssets(pn int, forceDestroy bool) {
	if len(s.chars[pn]) > 0 {
		// These aren't "assets" but we'll do it here
		for _, c := range s.chars[pn] {
			c.targets = c.targets[:0]
		}

		// Destroy helpers
		for _, h := range s.chars[pn][1:] {
			// F4 now destroys "preserve" helpers when reloading round start backup
			if !h.preserve || forceDestroy {
				h.destroy()
			}
		}
	}

	// Clear sounds
	s.charSoundChannels[pn].SetSize(0)

	// Clear projectiles, explods and text sprites
	s.projs[pn] = PointerSliceReset(s.projs[pn])
	s.explods[pn] = PointerSliceReset(s.explods[pn])
	s.chartexts[pn] = PointerSliceReset(s.chartexts[pn])

	// Clear explod draw order so that no reference is left to this char
	s.explodDrawOrder = PointerSliceReset(s.explodDrawOrder)
}

// Select correct stage for current round
func (s *System) handleStageSwap() bool {
	var roundRef int32
	if s.roundNo == 1 {
		s.stageLoopNo = 0
	} else {
		roundRef = s.roundNo
	}

	if s.stageLoop && !s.roundResetFlg {
		var keys []int
		for k := range s.stageList {
			keys = append(keys, int(k))
		}
		sort.Ints(keys)
		roundRef = int32(keys[s.stageLoopNo])
		s.stageLoopNo++
		if s.stageLoopNo >= len(s.stageList) {
			s.stageLoopNo = 0
		}
	}

	swap := false
	if _, ok := s.stageList[roundRef]; ok {
		s.stage = s.stageList[roundRef]
		if s.roundNo > 1 && !s.roundResetFlg {
			swap = true
		}
		if s.stage.model != nil {
			sys.mainThreadTask <- func() {
				gfx.SetModelVertexData(0, s.stage.model.vertexBuffer)
				gfx.SetModelIndexData(0, s.stage.model.elementBuffer...)
			}
		}
	}
	return swap
}

func (s *System) updateMusicMaps() {
	newMatchMusic := s.roundNo == 1 && !s.roundResetFlg

	if newMatchMusic {
		s.matchMusicSel = s.matchMusicSel[:0]
	}

	for i, p := range s.chars {
		if len(p) == 0 {
			continue
		}
		isSelectableChar := i < MaxSimul*2
		// Reset music map
		s.cgi[i].music = make(Music)
		// Append stage def file music parameters
		s.cgi[i].music.Append(sys.stage.music)
		// Append select.def stage music parameters
		s.cgi[i].music.Append(sys.stage.si().music)
		// Override with select.def char music parameters
		useCharMusic := false
		if sys.sel.gameParams != nil {
			useCharMusic = sys.sel.gameParams.CharParamMusic
		}
		if isSelectableChar {
			// In Versus modes, round/final/life music shouldn't override stage music; victory.music still can.
			if useCharMusic {
				s.cgi[i].music.Override(p[0].si().music)
			} else if v, ok := p[0].si().music[normalizeMusicPrefix("victory")]; ok {
				s.cgi[i].music.Override(Music{normalizeMusicPrefix("victory"): v})
			}
		}
		// Override with music with launchFight parameters
		s.cgi[i].music.Override(sys.sel.music)
		// Debug dump of merged music.
		s.cgi[i].music.DebugDump(fmt.Sprintf("cgi[%d].music after updateMusicMaps()", i))
	}
}

// TODO: This function is still a bit overloaded because it's handling some selections instead of doing a pure reset
func (s *System) resetRound() {
	if s.sel.gameParams.PersistRounds && !s.roundResetFlg {
		s.persistRoundCount++
	}

	s.resetFrameTime()

	s.restorePauseVolume()
	s.paused = false
	s.introSkipCall = false
	s.roundResetFlg = false
	s.reloadFlg, s.reloadStageFlg, s.reloadFightScreenFlg = false, false, false

	// https://github.com/ikemen-engine/Ikemen-GO/issues/1400
	s.noCharSoundFlg = false

	s.resetGblEffect()
	s.scorePoints = [2]float32{}
	s.comboCount = [2]int32{}
	s.fightScreen.reset()
	s.motif.reset()
	s.saveStateFlag = false
	s.loadStateFlag = false
	s.firstAttack = [3]int{-1, -1, 0}
	s.finishType = FT_NotYet
	s.winTeam = -1
	s.winType = [...]WinType{WT_Normal, WT_Normal}
	s.winTrigger = [...]WinType{WT_Normal, WT_Normal}
	s.effectiveLoss = [2]bool{false, false}
	s.lastHitter = [2]int{-1, -1}
	s.slowtime = s.fightScreen.round.slow_time
	s.winposetime = s.fightScreen.round.over_wintime
	s.winwaittime = s.fightScreen.round.over_waittime + s.fightScreen.round.over_forcewintime
	s.winskipped = false
	s.intro = s.fightScreen.round.start_waittime + s.fightScreen.round.ctrl_time + 1
	s.curRoundTime = s.maxRoundTime
	s.curPlayTime = 0
	// Mugen resets the starting ID between matches but not between rounds
	// Previously Ikemen reset it between rounds, but that creates the odd scenario where a new player in Turns mode will have the same ID as a previous player
	//s.nextCharId = s.cfg.Config.HelperMax

	// Decisive round check
	for i := range s.decisiveRound {
		s.decisiveRound[i] = s.wins[i] >= s.matchWins[i]-1
	}

	stageSwap := s.handleStageSwap()

	s.cam.stageCamera = s.stage.stageCamera
	s.cam.Init()

	if s.stage.resetbg || stageSwap {
		s.stage.reset()
	}

	// Reset characters
	for i, p := range s.chars {
		if len(p) == 0 {
			continue
		}
		s.clearPlayerAssets(i, false)
		p[0].posReset()
		p[0].setCtrl(false)
		p[0].clearState()
		p[0].prepareNextRound()
		p[0].varRangeSet(0, s.cgi[i].data.intpersistindex-1, 0)
		p[0].fvarRangeSet(0, s.cgi[i].data.floatpersistindex-1, 0)
		if s.reloadPreserveVars[i] {
			s.restoreCharVars(p[0])
			s.reloadPreserveVars[i] = false
		}

		for j := range p[0].cmd {
			p[0].cmd[j].BufReset()
		}
		if s.roundsExisted[i&1] == 0 {
			s.cgi[i].palettedata.palList.ResetRemap()
			if s.cgi[i].sff.header.Version[0] == 1 {
				p[0].remapPal(p[0].getPalfx(),
					[...]int32{1, 1}, [...]int32{1, s.cgi[i].palno})
			}
		}
	}

	// Place characters in state 5900
	for _, p := range s.chars {
		if len(p) == 0 {
			continue
		}
		// Select anim 0. If it is missing, choose the lowest available action
		// instead of depending on Go map iteration order.
		firstAnim := int32(0)
		if p[0].gi().animTable.anims[0] == nil {
			found := false
			for k := range p[0].gi().animTable.anims {
				if !found || k < firstAnim {
					firstAnim = k
					found = true
				}
			}
		}
		p[0].selfState(5900, firstAnim, -1, 0, "")
	}

	s.updateMusicMaps()

	// Make a new backup reflecting the post-swap roundXdef stage, or F4 will restore the prior round's stage
	// TODO: The fact that the reset() function needs to make a new backup shows that this function is currently overloaded
	// Update: Save operations moved outside
	//s.roundBackup.Save()

	// Ensure main thread catches up
	s.runMainThreadTask()
	gfx.Await()
}

func (s *System) debugPaused() bool {
	return s.paused && !s.frameStepFlag && s.oldTickCount < s.tickCount
}

// "Tick frames" are the frames where most of the game logic happens
func (s *System) tickFrame() bool {
	return (!s.paused || s.frameStepFlag) && s.oldTickCount < s.tickCount
}

// "Tick next frame" is right after the "tick frame"
// Where for instance the collision detections happen
func (s *System) tickNextFrame() bool {
	return int(s.tickCountF+s.nextAddTime) > s.tickCount &&
		(!s.paused || s.frameStepFlag || s.oldTickCount >= s.tickCount)
}

// This divides a frame into fractions for the purpose of drawing position interpolation
func (s *System) tickInterpolation() float32 {
	// Always synchronize here
	if s.tickNextFrame() {
		return 1
	}
	// Apply interpolation if enabled
	if sys.cfg.Config.TickInterpolation {
		progress := s.tickCountF - s.lastTick + s.nextAddTime
		return Clamp(progress, 0, 1)
	}
	return 1
}

func (s *System) addFrameTime(t float32) bool {
	if s.debugPaused() {
		s.oldNextAddTime = 0
		return true
	}
	s.oldTickCount = s.tickCount
	if int(s.tickCountF) > s.tickCount {
		s.tickCount++
		return false
	}
	s.tickCountF += s.nextAddTime
	if int(s.tickCountF) > s.tickCount {
		s.tickCount++
		s.lastTick = s.tickCountF
	}
	s.oldNextAddTime = s.nextAddTime
	s.nextAddTime = t
	return true
}

func (s *System) resetFrameTime() {
	s.tickCount, s.oldTickCount, s.tickCountF, s.lastTick = 0, -1, 0, 0
	s.nextAddTime, s.oldNextAddTime = 1, 1

	s.redrawWait.nextTime = time.Now()
	s.redrawWait.lastDraw = time.Now()
}

func (s *System) resetMatchData(forceDestroy bool) {
	sys.allPalFX = newPalFX()
	sys.bgPalFX = newPalFX()
	sys.resetGblEffect()
	for i, p := range sys.chars {
		if len(p) > 0 {
			sys.clearPlayerAssets(i, forceDestroy)
		}
	}
}

func (s *System) charUpdate() {
	s.charList.update()

	s.projectileUpdate()

	// Set global First Attack flag if either team got it
	if s.firstAttack[0] >= 0 || s.firstAttack[1] >= 0 {
		s.firstAttack[2] = 1
	}
}

// Run collision detection for chars and projectiles
func (s *System) globalCollision() {
	for i := range s.projs {
		for j, p := range s.projs[i] {
			p.tradeDetection(i, j)
		}
	}

	s.charList.collisionDetection()
}

/*
// This is never called
func (s *System) posReset() {
	for _, p := range s.chars {
		if len(p) > 0 {
			p[0].posReset()
		}
	}
}
*/

// Skip character intros on button press and play the shutter effect
func (s *System) runIntroSkip() {
	// If no intros to skip or not allowed to
	if !s.gsf(GSF_intro) || s.gsf(GSF_roundnotskip) {
		return
	}

	// If too late to skip intros
	if s.intro <= s.fightScreen.round.ctrl_time || s.fightScreen.round.roundDisplayPhase >= 1 { // Latter is probably redundant
		return
	}

	// Start shutter effect on button press
	if s.fightScreen.round.shutterTimer == 0 && s.anyButton() {
		s.fightScreen.round.shutterTimer = s.fightScreen.round.shutter_time * 2 // Open + close time
	}

	// Skip intros when signal from shutter animation arrives
	if s.introSkipCall {
		s.introSkipCall = false
		s.intro = s.fightScreen.round.ctrl_time

		// SkipRoundDisplay and SkipFightDisplay flags must be preserved during intro skip frame
		kept := (s.specialFlag & GSF_skiprounddisplay) | (s.specialFlag & GSF_skipfightdisplay)
		s.resetGblEffect()
		s.specialFlag = kept

		// Reset all characters
		for i, p := range s.chars {
			if len(p) > 0 {
				s.clearPlayerAssets(i, false)
				p[0].posReset()
				p[0].selfState(0, -1, -1, 0, "")
			}
		}
	}
}

// TODO: Maybe these could get a more thorough clear between matches, to prevent holding onto previous character sprites
func (s *System) clearSpriteData() {
	// Main sprites
	s.spriteList = s.spriteList[:0]

	// Shadows and reflections
	s.shadowList = s.shadowList[:0]
	s.reflectionList = s.reflectionList[:0]

	// Debug sprites
	s.debugc1hit.rects = s.debugc1hit.rects[:0]
	s.debugc1rev.rects = s.debugc1rev.rects[:0]
	s.debugc1not.rects = s.debugc1not.rects[:0]
	s.debugc2.rects = s.debugc2.rects[:0]
	s.debugc2hb.rects = s.debugc2hb.rects[:0]
	s.debugc2mtk.rects = s.debugc2mtk.rects[:0]
	s.debugc2grd.rects = s.debugc2grd.rects[:0]
	s.debugc2stb.rects = s.debugc2stb.rects[:0]
	s.debugcsize.rects = s.debugcsize.rects[:0]
	s.debugch.rects = s.debugch.rects[:0]
	s.debugClsnText = nil

	// Reset afterimage tracker
	for i := range s.afterImageCount {
		s.afterImageCount[i] = 0
	}
}

func (s *System) screenleft() float32 {
	return float32(s.stage.screenleft) * s.stage.localscl
}

func (s *System) screenright() float32 {
	return float32(s.stage.screenright) * s.stage.localscl
}

func (s *System) action() {
	s.clearSpriteData()

	var x, y, scl float32 = s.cam.Pos[0], s.cam.Pos[1], s.cam.Scale / s.cam.BaseScale()
	s.cam.ResetTracking()

	// Update round state
	// This is also reflected on characters (intros, win poses)
	// It's important that this is placed after the tickFrame logic, or characters will not see every step of the sys.intro timer
	// Update: In Mugen, the state change to win pose happens before the character code runs. It's evidence this should be placed before charList.action()
	s.stepRoundState()

	// In version 0.99 this was moved to before sys.action() for undocumented reasons
	// Moving it here preserves whatever behavior that implemented while not keeping the stage oddly outside the main loop
	s.stage.action()

	// Run "tick frame"
	if s.tickFrame() {
		// X axis player limits
		s.xmin = s.cam.ScreenPos[0] + s.cam.Offset[0] + s.screenleft()
		s.xmax = s.cam.ScreenPos[0] + s.cam.Offset[0] + float32(s.gameWidth)/s.cam.Scale - s.screenright()
		if s.xmin > s.xmax {
			s.xmin = (s.xmin + s.xmax) / 2
			s.xmax = s.xmin
		}
		if Abs(s.cam.maxRight-s.xmax) < 0.0001 {
			s.xmax = s.cam.maxRight
		}
		if Abs(s.cam.minLeft-s.xmin) < 0.0001 {
			s.xmin = s.cam.minLeft
		}
		// Z axis player limits
		s.zmin = s.stage.topbound * s.stage.localscl
		s.zmax = s.stage.botbound * s.stage.localscl
		//s.bgPalFX.step()

		if s.envcol_time > 0 {
			s.envcol_time--
		}

		// Step pause timers
		if s.supertime > 0 {
			s.supertime--
		} else if s.pausetime > 0 {
			s.pausetime--
		}

		// Start pause timers
		if s.supertimebuffer < 0 {
			s.supertimebuffer = ^s.supertimebuffer
			s.supertime = s.supertimebuffer
		}
		if s.pausetimebuffer < 0 {
			s.pausetimebuffer = ^s.pausetimebuffer
			s.pausetime = s.pausetimebuffer
		}

		// In Mugen 1.1, few global AssertSpecial flags persist during pauses. Seemingly only TimerFreeze
		if s.supertime <= 0 && s.pausetime <= 0 {
			s.specialFlag = 0
		} else {
			// These flags persist even during pauses
			// "NoKOSlow" added to facilitate custom slowdown. In Mugen that flag only needs to be asserted in first frame of KO slowdown
			s.specialFlag = (s.specialFlag&GSF_nokoslow | s.specialFlag&GSF_timerfreeze)
		}

		// Run the main character logic
		s.charList.action()

		// The following must be placed after char action or they will lag behind 1 frame
		s.allPalFX.step()
		s.bgPalFX.step()
		s.envShake.update()
		s.zoom.update()
		s.nomusic = s.gsf(GSF_nomusic) && !sys.postMatchFlg
	}

	// This function runs every tick
	// It should be placed between "tick frame" and "tick next frame"
	s.charUpdate()

	// Update the fight screen
	// Lifebar and combo must update after character states but before hit detection for accurate detection
	// So that it allows a combo to still end if a character is hit in the same frame where it exits movetype H
	s.fightScreen.step()

	if s.tickNextFrame() {
		s.globalCollision() // This could perhaps happen during "tick frame" instead? Would need more testing
		s.globalTick()
	}

	// Run camera
	x, y, scl = s.cam.action(x, y, scl, s.supertime > 0 || s.pausetime > 0)

	// Character intro skipping
	if s.tickNextFrame() {
		s.runIntroSkip()
	}

	if !s.cam.ZoomEnable {
		// Lower the precision to prevent errors in Pos X.
		x = float32(math.Ceil(float64(x)*4-0.5) / 4)
	}
	s.cam.Update(scl, x, y)
	s.xmin = s.cam.ScreenPos[0] + s.cam.Offset[0] + s.screenleft()
	s.xmax = s.cam.ScreenPos[0] + s.cam.Offset[0] + float32(s.gameWidth)/s.cam.Scale - s.screenright()
	if s.xmin > s.xmax {
		s.xmin = (s.xmin + s.xmax) / 2
		s.xmax = s.xmin
	}
	if Abs(s.cam.maxRight-s.xmax) < 0.0001 {
		s.xmax = s.cam.maxRight
	}
	if Abs(s.cam.minLeft-s.xmin) < 0.0001 {
		s.xmin = s.cam.minLeft
	}
	s.charList.xScreenBound()

	for i := range s.projs {
		for _, p := range s.projs[i] {
			p.cueDraw()
		}
	}

	s.charList.cueDraw()

	// Note: Explod update must happen after hit detection. Because hit sparks are also explods
	s.explodUpdate()
	s.charTextsUpdate()
	s.explodCueDraw()

	// Adjust game speed
	if s.tickNextFrame() {
		spd := float32(s.gameLogicSpeed()) / float32(s.gameRenderSpeed())

		// KO slowdown
		if st := s.getSlowtime(); st > 0 {
			if !s.gsf(GSF_nokoslow) {
				base := s.fightScreen.round.slow_speed
				fade := s.fightScreen.round.slow_fadetime
				spd *= base
				if st < fade {
					ratio := float32(fade-st) / float32(fade)
					spd = base + (1-base)*ratio
				}
			}
			s.slowtime--
		}

		// Outside match or while frame stepping
		if s.postMatchFlg || s.frameStepFlag {
			spd = 1
		}

		// Save result
		s.turbo = spd
	}

	// Force Feedback (legacy)
	if s.tickNextFrame() {
		for i := 0; i < len(s.ffbparams); i++ {
			if s.ffbparams[i].timer > 0 {
				start := s.ffbparams[i].start
				t := s.ffbparams[i].timer
				d1, d2, d3 := s.ffbparams[i].d1, s.ffbparams[i].d2, s.ffbparams[i].d3
				d1 = float32(math.Pow(float64(d1), float64(t)))
				d2 = float32(math.Pow(float64(d2), float64(t*t)))
				d3 = float32(math.Pow(float64(d2), float64(t*t*t)))
				ampl := (start + d1 + d2 + d3) / 255.0
				intensity := uint16(ampl * 0xFFFF)
				switch s.ffbparams[i].waveform {
				case waveform_off:
					input.RumbleController(i, 0, 0, 1)
				case waveform_sine:
					input.RumbleController(i, intensity, 0, 1)
				case waveform_square:
					input.RumbleController(i, 0, intensity, 1)
				case waveform_sinesquare:
					input.RumbleController(i, intensity, intensity, 1)
				}
				s.ffbparams[i].timer--
			}
		}
	}

	// Update motif
	// Needs to happen at the very end or pause toggles will get out of sync
	// https://github.com/ikemen-engine/Ikemen-GO/issues/3080
	s.motif.step()

	// Run motif
	s.motif.act()

	// Common Lua calls
	// Needs to happens after motif update or motif inputs will lag 1 frame
	for _, key := range SortedKeys(sys.cfg.Common.Lua) {
		for _, v := range sys.cfg.Common.Lua[key] {
			if err := sys.luaLState.DoString(v); err != nil {
				sys.luaLState.RaiseError("Error executing Lua code: %s\n%v", v, err.Error())
			}
		}
	}

	s.tickSound()
	return
}

// Update all projectiles for all players
func (s *System) projectileUpdate() {
	// Because sys.projs has actual values rather than pointers like sys.chars does, it's important to not copy its contents with range
	// https://github.com/ikemen-engine/Ikemen-GO/discussions/1707
	// Update: Projectiles now work based on pointers, so we can go back to old loop format
	for i := range s.projs {
		for _, p := range s.projs[i] {
			p.update()
		}
	}

	// Cleanup
	for i := range s.projs {
		s.projectilePrune(i)
	}
}

// Remove a player's inactive projectiles
// This method avoids pointer duplication bugs caused by the old "tempSlice" method
func (s *System) projectilePrune(pn int) {
	playerProjs := s.projs[pn]
	writeIdx := 0

	for j := 0; j < len(playerProjs); j++ {
		// Keep only valid projectiles
		if playerProjs[j].id >= 0 {
			// Stable swap. Move live projectiles forward while respecting the original order
			playerProjs[writeIdx], playerProjs[j] = playerProjs[j], playerProjs[writeIdx]
			writeIdx++
		}
	}

	// Shrink the slice to the new length
	s.projs[pn] = playerProjs[:writeIdx]
}

// Update all explods for all players
func (s *System) explodUpdate() {
	// Update each explod in storage order
	for i := range s.explods {
		for _, e := range s.explods[i] {
			e.update()
		}
	}

	// Cleanup
	for i := range s.explods {
		s.explodPrune(i)
	}
}

// Sort explods by age, ontop, etc before queuing them for drawing
func (s *System) explodCueDraw() {
	// Reset sorting list
	s.explodDrawOrder = s.explodDrawOrder[:0]

	// Flatten all explods into the sorting list
	// Using this list makes the drawing order fairer instead of prioritizing player 1
	// Mugen probably just used a single array for explods instead of organizing them by player, so it skipped this issue
	// https://github.com/ikemen-engine/Ikemen-GO/issues/1099
	count := 0
	for i := range s.explods {
		for j := range s.explods[i] {
			e := s.explods[i][j]
			e.sortindex = count // Unique reference for this frame
			s.explodDrawOrder = append(s.explodDrawOrder, e)
			count++
		}
	}

	// Sort explods by age
	// For the sake of backward compatibility we must also open exceptions for the ontop parameter
	// The sortindex variable saves us from needing the more expensive SliceStable
	sort.Slice(s.explodDrawOrder, func(i, j int) bool {
		a, b := s.explodDrawOrder[i], s.explodDrawOrder[j]

		// All ontop explods come before the rest
		if a.ontop != b.ontop {
			return a.ontop
		}
		// If both are ontop, the normal logic is inverted in order to emulate Mugen's memory layout
		// However, it's impossible to cover all of Mugen's quirks without making our layout worse
		// https://github.com/ikemen-engine/Ikemen-GO/issues/3737
		// https://github.com/ikemen-engine/Ikemen-GO/issues/3749
		if a.ontop && b.ontop {
			if a.timestamp != b.timestamp {
				return a.timestamp >= b.timestamp
			}
			return a.sortindex >= b.sortindex
		}
		// Normal case: older timestamps come first
		if a.timestamp != b.timestamp {
			return a.timestamp < b.timestamp
		}
		// Same timestamp: keep original order
		return a.sortindex < b.sortindex
	})

	// Queue in the sorted order
	for _, e := range s.explodDrawOrder {
		e.cueDraw()
	}
}

// Remove a player's inactive explods
func (s *System) explodPrune(pn int) {
	playerExplods := s.explods[pn]
	writeIdx := 0

	for j := 0; j < len(playerExplods); j++ {
		// Keep only valid explods
		if playerExplods[j].id != IErr {
			playerExplods[writeIdx], playerExplods[j] = playerExplods[j], playerExplods[writeIdx]
			writeIdx++
		}
	}

	s.explods[pn] = playerExplods[:writeIdx]
}

// Update all texts for all players
func (s *System) charTextsUpdate() {
	// Explod timers update at this time, so we'll do the same here
	//if !sys.tickNextFrame() {
	// It looks like updating them at that time makes them flicker at lower game speeds
	if !sys.tickFrame() {
		return
	}

	// With texts we run the cleanup first. Otherwise a removetime of 1 disappears immediately
	// The reason this is different from explods is that texts don't use a DrawList to buffer the sprites
	for i := range s.chartexts {
		s.charTextsPrune(i)
	}

	// Logic
	for i := range s.chartexts {
		for _, ts := range s.chartexts[i] {
			ts.Update()
			if ts.removetime > 0 {
				ts.removetime--
			}
		}
	}
}

// Remove a player's inactive texts
func (s *System) charTextsPrune(pn int) {
	playerTexts := s.chartexts[pn]
	writeIdx := 0

	for j := 0; j < len(playerTexts); j++ {
		// Keep only valid texts
		if playerTexts[j].removetime != 0 {
			playerTexts[writeIdx], playerTexts[j] = playerTexts[j], playerTexts[writeIdx]
			writeIdx++
		}
	}

	s.chartexts[pn] = playerTexts[:writeIdx]
}

func (s *System) globalTick() {
	s.stage.tick()
	s.charList.tick()

	for i := range s.projs {
		for _, p := range s.projs[i] {
			p.tick()
		}
	}

	s.matchTime++
}

func (s *System) getSlowtime() int32 {
	if s.slowtime > 0 && s.intro < 0 && s.curRoundTime != 0 {
		return s.slowtime
	}
	return 0
}

func (s *System) timeElapsed() int32 {
	// Timed rounds
	if s.maxRoundTime > 0 {
		return s.maxRoundTime - s.curRoundTime
	}
	// Unlimited rounds
	return s.curPlayTime
}

func (s *System) timeRemaining() int32 {
	if s.curRoundTime >= 0 {
		return s.curRoundTime
	}
	return -1
}

func (s *System) timeTotal() int32 {
	t := s.timerStart
	for _, v := range s.timerRounds {
		t += v
	}
	if s.fightScreen.round.timerActive {
		t += s.timeElapsed()
	}
	return t
}

func (s *System) matchEndDialoguePending() bool {
	if s.postMatchFlg || s.dialogueForce != 0 || s.motif.di.matchEndDone {
		return false
	}
	if !s.motif.di.enabled || !s.motif.isDialogueSet() {
		return false
	}
	if s.matchOver() {
		return true
	}
	if s.finishType == FT_NotYet {
		return false
	}
	if s.winTeam >= 0 {
		return s.decisiveRound[s.winTeam]
	}
	// Draw-based match end is determined before draws is incremented for this round.
	return s.maxDrawsReached(0) || s.maxDrawsReached(1) // TODO: Maybe this should be &&
}

func (s *System) shouldStartMatchEndDialogue() bool {
	if !s.matchEndDialoguePending() || !s.matchOver() || s.postMatchFlg {
		return false
	}
	if s.motif.di.active || s.motif.di.initialized {
		return false
	}
	return s.intro+s.fightScreen.round.over_waittime+s.fightScreen.round.over_time <= 0
}

func (s *System) holdPostMatchForDialogue() bool {
	if s.dialogueForce != 0 || s.postMatchFlg {
		return false
	}
	if s.motif.di.active && s.motif.di.atMatchEnd {
		return true
	}
	if s.shouldStartMatchEndDialogue() {
		return true
	}
	if s.motif.di.matchEndDone && s.fightScreen.round.fadeOut.isActive() {
		return true
	}
	return false
}

// Step sys.intro timer and execute related tasks
func (s *System) stepRoundState() {
	// Freeze round state if round animations cannot advance
	if !s.fightScreen.round.act() {
		return
	}

	// Fading
	if !(s.fightScreen.round.fadeOut.isActive() || s.fightScreen.round.fadeIn.isActive()) {
		if s.motif.fadeOut.isActive() {
			s.motif.fadeOut.step()
		} else if s.motif.fadeIn.isActive() {
			s.motif.fadeIn.step()
		}
	}

	// Intros
	if s.intro > s.fightScreen.round.ctrl_time {
		s.intro--
		if s.gsf(GSF_intro) && s.intro <= s.fightScreen.round.ctrl_time {
			s.intro = s.fightScreen.round.ctrl_time + 1
		}
	} else if s.intro > 0 {
		if s.intro == s.fightScreen.round.ctrl_time {
			for _, p := range s.chars {
				if len(p) > 0 {
					if p[0].activelyFighting() && !p[0].asf(ASF_nointroreset) {
						p[0].posReset() // Technically we don't really need this step
					}
				}
			}
		}
		s.intro--
		if s.intro == 0 {
			for _, p := range s.chars {
				if len(p) > 0 {
					if p[0].alive() {
						p[0].unsetSCF(SCF_over_alive)
						//if !p[0].scf(SCF_standby) || p[0].teamside == -1 {
						if p[0].activelyFighting() {
							// In Mugen this ctrlSet happens in beginning of frame. More evidence all this code should run before character states
							p[0].setCtrl(true)
							if p[0].ss.no != 0 && !p[0].asf(ASF_nointroreset) {
								p[0].selfState(0, -1, -1, 1, "") // Nor this one
							}
						}
					}
				}
			}
		}
	}

	// Ongoing round
	// Handle remaining time limit
	if s.intro == 0 && !s.gsf(GSF_timerfreeze) && s.supertime <= 0 && s.pausetime <= 0 {
		if s.maxRoundTime > 0 && s.curRoundTime > 0 {
			s.curRoundTime--
		}
		s.curPlayTime++
	}

	// Post round
	if s.roundEnded() || s.roundEndDecision() {
		rs4t := -s.fightScreen.round.over_waittime
		fadeOutDuration := s.fightScreen.round.fadeOut.duration()
		fadeoutStart := rs4t - 2 - s.fightScreen.round.overTime() + fadeOutDuration
		matchEndDialoguePending := s.matchEndDialoguePending()

		s.intro--

		if s.intro == -s.fightScreen.round.over_hittime && s.finishType != FT_NotYet {
			// Consecutive wins counter
			winner := [2]bool{s.effectiveLoss[1], s.effectiveLoss[0]}
			if !winner[0] || !winner[1] ||
				s.tmode[0] == TM_Turns || s.tmode[1] == TM_Turns || s.maxDrawsReached(0) || s.maxDrawsReached(1) {
				for i, win := range winner {
					if win {
						s.wins[i]++
						if s.matchOver() && s.wins[^i&1] == 0 {
							s.consecutiveWins[i]++
						}
						s.consecutiveWins[^i&1] = 0
					}
				}
			}
		}

		// Check if player skipped win pose time
		if !s.winskipped && s.winposetime < 0 && s.anyButton() &&
			!s.gsf(GSF_roundnotskip) && !matchEndDialoguePending {
			s.intro = Min(s.intro, fadeoutStart)
			s.winskipped = true
		}
		// Match-end dialogue takes priority over the round transition until it finishes.
		if matchEndDialoguePending && !s.motif.di.active && !s.motif.di.matchEndDone {
			if s.fightScreen.round.fadeOut.isActive() {
				s.fightScreen.round.fadeOut.reset()
			}
		}

		// Start the normal round fade only when no match-end dialogue is pending.
		// If the user skipped winposes, don't let RoundNotOver swallow the single fadeoutStart tick.
		if !matchEndDialoguePending &&
			s.intro == fadeoutStart &&
			(!s.gsf(GSF_roundnotover) || s.winskipped) &&
			!s.motif.di.active && !s.fightScreen.round.fadeOut.isActive() {
			s.fightScreen.round.fadeOut.init(s.fightScreen.round.fadeOut, false)
		}

		// Before win poses
		if s.winposetime > 0 {
			// Check if game can proceed into roundstate 4
			if s.winwaittime > 0 {
				if s.intro == rs4t-1 {
					for _, p := range s.chars {
						if len(p) == 0 {
							continue
						}
						// Check if this player is ready to proceed to roundstate 4
						// Maybe the "activelyFighting()" could be replaced with an AssertSpecial flag or such to ignore the win/lose states
						if !p[0].activelyFighting() {
							continue
						}
						if p[0].scf(SCF_over_alive) || p[0].scf(SCF_over_ko) {
							continue
						}
						// Mugen seems to skip this anim 5 check on time overs
						// It also seems a bit pointless here to begin with because the char has already turned by the time anim 5 starts
						if p[0].scf(SCF_ctrl) && p[0].ss.moveType == MT_I && p[0].ss.stateType == ST_S && p[0].animNo != 5 {
							continue
						}
						// Freeze timer if any player is not ready to proceed yet
						s.intro = rs4t
						break
					}
				}
			}

			// Disable ctrl (once) at the first frame of roundstate 4
			if s.intro == rs4t-1 {
				for _, p := range s.chars {
					if len(p) > 0 {
						p[0].setCtrl(false)
					}
				}
			}
		}

		// Start running wintime counter only after getting into roundstate 4
		if s.intro < rs4t {
			s.winposetime--
		}

		// Set timer to start win/lose poses. Update win counters
		// In the first frame of win poses only
		if s.winposetime == 0 {
			// Attribute the win icon
			winner := [2]bool{s.effectiveLoss[1], s.effectiveLoss[0]}

			if winner[0] || winner[1] {
				for i, win := range winner {
					if win {
						s.fightScreen.winIcons[i].add(s.winType[i])
						if s.matchOver() {
							// In a draw game both players go back to 0 wins
							if winner[0] == winner[1] {
								s.fightScreen.winCounts[0].wins = 0
								s.fightScreen.winCounts[1].wins = 0
							} else {
								if s.wins[i] >= s.matchWins[i] {
									s.fightScreen.winCounts[i].wins++
								}
							}
						}
					}
				}
			} else {
				s.draws++
			}

			// Fast forward wait timer since we're now done waiting
			s.winwaittime = 0
		}

		// Send characters to win/lose poses
		// In Mugen this loop can run at any point after the win poses have started, hence "<="
		if s.winposetime <= 0 {
			for _, p := range s.chars {
				if len(p) == 0 {
					continue
				}
				// TODO: These changestates ought to be unhardcoded
				// TODO: They also happen 1 frame too early compared to Mugen
				// Mugen only checks for readiness in the RoundState 4 loop above. It doesn't care in this one and forces characters into win poses no matter what
				// HitPause is checked here but not in the other loop. Perhaps because changing states during hitpause in Mugen isn't quite safe
				// These selfStates should happen in beginning of frame. Previously, Ikemen had them at end of frame
				if !p[0].scf(SCF_over_alive) && p[0].alive() && p[0].activelyFighting() && !p[0].hitPause() {
					p[0].setSCF(SCF_over_alive)
					if p[0].win() {
						p[0].selfState(180, -1, -1, -1, "")
					} else if p[0].lose() {
						p[0].selfState(170, -1, -1, -1, "")
					} else {
						p[0].selfState(175, -1, -1, -1, "")
					}
				}
			}
		}

		if s.winwaittime > 0 {
			s.winwaittime--
		}
		// If the game can't proceed to the fadeout screen, we turn back the counter 1 tick
		if !matchEndDialoguePending && !s.winskipped && s.gsf(GSF_roundnotover) && s.intro == fadeoutStart {
			s.intro++
		}
	} else if s.intro < 0 {
		s.intro = 0
	}
}

// Check if the round ended by KO or time over and set win types
func (s *System) roundEndDecision() bool {
	checkPerfect := func(team int) bool {
		for i := team; i < MaxSimul*2; i += 2 {
			if len(s.chars[i]) > 0 &&
				s.chars[i][0].teamside != -1 &&
				s.chars[i][0].life < s.chars[i][0].lifeMax {
				return false
			}
		}
		return true
	}
	if s.intro > 0 {
		return false
	}

	checkClutch := func(team int) bool {
		clutchRatio := float32(s.fightScreen.round.clutch_threshold) / 100.0
		for i := team; i < MaxSimul*2; i += 2 {
			if len(s.chars[i]) > 0 {
				char := s.chars[i][0]
				if char.teamside != -1 && float32(char.life) > float32(char.lifeMax)*clutchRatio {
					return false
				}
			}
		}
		return true
	}

	// KO check
	ko := [2]bool{true, true}
	for loser := range ko {
		// Check if all players or leader on one side are KO
		for i := loser; i < MaxSimul*2; i += 2 {
			if len(s.chars[i]) > 0 && s.chars[i][0].teamside != -1 {
				if s.chars[i][0].alive() {
					ko[loser] = false
				} else if (s.tmode[i&1] == TM_Simul && s.cfg.Options.Simul.LoseOnKO && s.aiLevel[i] == 0) ||
					(s.tmode[i&1] == TM_Tag && s.cfg.Options.Tag.LoseOnKO) {
					ko[loser] = true
					break
				}
			}
		}
		if ko[loser] {
			if checkPerfect(loser ^ 1) {
				s.winType[loser^1].SetPerfect()
			} else if checkClutch(loser ^ 1) {
				s.winType[loser^1].SetClutch()
			}
		}
	}

	// Time over
	ft := s.finishType
	if s.curRoundTime == 0 {
		s.winType[0], s.winType[1] = WT_Time, WT_Time
		l := [2]float32{}
		for i := 0; i < 2; i++ { // Check life percentage of each team
			for j := i; j < MaxSimul*2; j += 2 {
				if len(s.chars[j]) > 0 {
					if s.chars[j][0].teamside == -1 {
						continue
					}
					if s.tmode[i] == TM_Simul || s.tmode[i] == TM_Tag {
						l[i] += (float32(s.chars[j][0].life) / float32(s.numSimul[i])) / float32(s.chars[j][0].lifeMax)
					} else {
						l[i] += float32(s.chars[j][0].life) / float32(s.chars[j][0].lifeMax)
					}
				}
			}
		}
		// Some other methods were considered to make the winner decision more fair, like a minimum % difference
		// But ultimately a direct comparison seems to be the fairest method
		if math.Round(float64(l[0]*1000)) != math.Round(float64(l[1]*1000)) || // Convert back to 1000 life points scale then round it to reduce calculation errors
			((l[0] >= float32(1.0)) != (l[1] >= float32(1.0))) { // But make sure the rounding doesn't turn a perfect into a draw game
			winner := 0
			if l[0] < l[1] {
				winner = 1
			}
			if checkPerfect(winner) {
				s.winType[winner].SetPerfect()
			} else if checkClutch(winner) {
				s.winType[winner].SetClutch()
			}
			s.finishType = FT_TO
			s.winTeam = winner
		} else { // Draw game
			s.finishType = FT_TODraw
			s.winTeam = -1
		}
	}

	// KO
	if s.intro >= -1 && (ko[0] || ko[1]) {
		if ko[0] && ko[1] {
			s.finishType = FT_DKO
			s.winTeam = -1
		} else {
			s.finishType = FT_KO
			s.winTeam = int(Btoi(ko[0]))
		}
	}

	// Effective loss check
	// Accounts for max draws unlike plain win/lose logic. Used for round progression
	if s.winTeam < 0 {
		// TODO: Turns mode could have special handling for balancing team sizes here
		// For instance only allow draws if both teams are the same size
		// In the meantime treating everything the same is the most impartial
		for i := 0; i < 2; i++ {
			s.effectiveLoss[i] = s.maxDrawsReached(i)
		}
	} else {
		for i := range s.effectiveLoss {
			s.effectiveLoss[i] = s.winTeam != i
		}
	}

	// Update win triggers if finish type was changed
	if ft != s.finishType {
		for i, p := range s.chars {
			if len(p) > 0 && p[0].teamside != -1 && ko[^i&1] {
				for _, h := range p {
					for _, tid := range h.targets {
						if t := s.playerID(tid); t != nil {
							if t.ghv.attr&int32(AT_AH) != 0 {
								s.winTrigger[i&1] = WT_Hyper
							} else if t.ghv.attr&int32(AT_AS) != 0 && s.winTrigger[i&1] == WT_Normal {
								s.winTrigger[i&1] = WT_Special
							}
						}
					}
				}
			}
		}
	}

	return ko[0] || ko[1] || s.curRoundTime == 0
}

func (s *System) draw(x, y, scl float32) {
	s.brightnessOld = s.brightness
	//s.brightness = 0x100 >> uint(Btoi(s.supertime > 0 && s.superdarken))
	s.brightness = 1.0
	if s.supertime > 0 && s.superbrightness >= 0 && s.superbrightness < 1 {
		s.brightness = s.superbrightness
	}

	bgx, bgy := x/s.stage.localscl, y/s.stage.localscl

	//fade := func(rect [4]int32, color uint32, alpha int32) {
	//	FillRect(rect, color, alpha>>uint(Btoi(s.clsnDisplay))+Btoi(s.clsnDisplay)*128, nil)
	//}

	if s.envcol_time == 0 {
		// Determine fill color
		var fcol uint32
		if s.gsf(GSF_nobg) {
			fcol = 0
		} else if s.stage.debugbg {
			fcol = 0xff00ff
		} else {
			fcol = uint32(s.stage.bgclearcolor[2]&0xff | s.stage.bgclearcolor[1]&0xff<<8 | s.stage.bgclearcolor[0]&0xff<<16)
		}

		// Apply BGPalFX to fill color
		// Mugen doesn't do this, but it makes sense
		//if !s.gsf(GSF_nobg) && s.bgPalFX.enable {
		//	fcol = s.bgPalFX.SynthesizeColor(TransType(TT_none), fcol)
		//}

		// Draw the background fill
		FillRect(s.scrrect, fcol, [2]int32{255, 0}, nil)

		// Draw stage elements with layerNo == -1
		if !s.gsf(GSF_nobg) {
			if s.stage.ikemenver[0] != 0 || s.stage.ikemenver[1] != 0 { // This layer did not render in Mugen
				s.stage.draw(-1, bgx, bgy, scl)
			}
		}

		// Draw reflections on layer -1
		if !s.gsf(GSF_globalnoshadow) {
			if s.stage.reflection.layerno < 0 {
				s.reflectionList.draw(x, y, scl*s.cam.BaseScale())
			}
		}

		// Draw character sprites with layerNo == -1
		s.spriteList.draw(-1, false, x, y, scl*s.cam.BaseScale())

		// Draw stage elements with layerNo == 0
		if !s.gsf(GSF_nobg) {
			s.stage.draw(0, bgx, bgy, scl)
		}

		// Draw character sprites with special under flag
		s.spriteList.draw(0, true, x, y, scl*s.cam.BaseScale())

		// Draw fight screen layer -1
		s.fightScreen.draw(-1)

		// Draw char texts layer -1
		s.drawCharTexts(-1)

		// Draw motif layer -1
		s.motif.draw(-1)

		// Draw shadows
		// Draw reflections on layer 0
		// TODO: Make shadows render in same layers as their sources?
		if !s.gsf(GSF_globalnoshadow) {
			if s.stage.reflection.layerno >= 0 {
				s.reflectionList.draw(x, y, scl*s.cam.BaseScale())
			}
			s.shadowList.draw(x, y, scl*s.cam.BaseScale())
		}

		//off := s.envShake.getOffset()
		//yofs, yofs2 := float32(s.gameHeight), float32(0)
		//if scl > 1 && s.cam.verticalfollow > 0 {
		//	yofs = s.cam.screenZoff + float32(s.gameHeight-240)
		//	yofs2 = (240 - s.cam.screenZoff) * (1 - 1/scl)
		//}
		//yofs *= 1/scl - 1
		//rect := s.scrrect
		//if off < (yofs-y+s.cam.boundH)*scl {
		//	rect[3] = (int32(math.Ceil(float64(((yofs-y+s.cam.boundH)*scl-off)*
		//		float32(s.scrrect[3])))) + s.gameHeight - 1) / s.gameHeight
		//	fade(rect, 0, 255)
		//}
		//if off > (-y+yofs2)*scl {
		//	rect[3] = (int32(math.Ceil(float64(((y-yofs2)*scl+off)*
		//		float32(s.scrrect[3])))) + s.gameHeight - 1) / s.gameHeight
		//	rect[1] = s.scrrect[3] - rect[3]
		//	fade(rect, 0, 255)
		//}
		//bl, br := Min(x, s.cam.boundL), Max(x, s.cam.boundR)
		//xofs := float32(s.gameWidth) * (1/scl - 1) / 2
		//rect = s.scrrect
		//if x-xofs < bl {
		//	rect[2] = (int32(math.Ceil(float64((bl-(x-xofs))*scl*
		//		float32(s.scrrect[2])))) + s.gameWidth - 1) / s.gameWidth
		//	fade(rect, 0, 255)
		//}
		//if x+xofs > br {
		//	rect[2] = (int32(math.Ceil(float64(((x+xofs)-br)*scl*
		//		float32(s.scrrect[2])))) + s.gameWidth - 1) / s.gameWidth
		//	rect[0] = s.scrrect[2] - rect[2]
		//	fade(rect, 0, 255)
		//}

		// Draw fight screen layer 0
		s.fightScreen.draw(0)

		// Draw char texts layer 0
		s.drawCharTexts(0)

		// Draw motif layer 0
		s.motif.draw(0)
	}

	// Draw EnvColor effect
	if s.envcol_time != 0 {
		ecol := uint32(s.envcol[2]&0xff | s.envcol[1]&0xff<<8 | s.envcol[0]&0xff<<16)
		FillRect(s.scrrect, ecol, [2]int32{255, 0}, nil)
	}

	// Draw character sprites in layer 0
	if s.envcol_time == 0 || s.envcol_under {
		s.spriteList.draw(0, false, x, y, scl*s.cam.BaseScale())
		if s.envcol_time == 0 && !s.gsf(GSF_nofg) {
			s.stage.draw(1, bgx, bgy, scl)
		}
	}

	// Draw fight screen layer 1
	s.fightScreen.draw(1)

	// Draw char texts layer 1
	s.drawCharTexts(1)

	// Draw motif layer 1
	s.motif.draw(1)

	// Draw character sprites in layer 1 (old "ontop")
	s.spriteList.draw(1, false, x, y, scl*s.cam.BaseScale())

	// Draw fight screen layer 2
	s.fightScreen.draw(2)

	// Draw char texts layer 2
	s.drawCharTexts(2)

	// Draw motif layer 2
	s.motif.draw(2)

	// Draw system fade/shutter over top-layer texts
	s.fightScreen.drawFade()

	// Draw motif layer 3
	s.motif.draw(3)
}

func (s *System) drawCharTexts(layerno int16) {
	for _, playerTexts := range s.chartexts {
		for _, ts := range playerTexts {
			ts.Draw(layerno)
		}
	}
}

func (s *System) drawTop() {
	s.brightness = s.brightnessOld

	// Draw Clsn boxes
	if s.clsnDisplay {
		alpha := [2]int32{255, 255}

		// Clsn1 HitDef
		s.debugc1hit.draw(0xff0000ff, alpha)
		// Clsn1 ReversalDef
		s.debugc1rev.draw(0xff0040c0, alpha)
		// Clsn1 Inactive
		s.debugc1not.draw(0xff000080, alpha)
		// Clsn2
		s.debugc2.draw(0xffff0000, alpha)
		// Clsn2 HitBy
		s.debugc2hb.draw(0xff808000, alpha)
		// Clsn2 Invincible
		s.debugc2mtk.draw(0xff004000, alpha)
		// Clsn2 Guarding
		s.debugc2grd.draw(0xffc00040, alpha)
		// Clsn2 Standby
		s.debugc2stb.draw(0xff404040, alpha)
		// Size
		s.debugcsize.draw(0xff303030, alpha)
		// Crosshair
		s.debugch.draw(0xffffffff, alpha)
	}
}

func (s *System) drawDebugText() {
	put := func(x, y *float32, txt string) {
		for txt != "" {
			w, drawTxt := int32(0), ""
			for i, r := range txt {
				w += s.debugFont.fnt.CharWidth(r, 0) + s.debugFont.fnt.Spacing[0]
				if w > s.scrrect[2] {
					drawTxt, txt = txt[:i], txt[i:]
					break
				}
			}
			if drawTxt == "" {
				drawTxt, txt = txt, ""
			}
			*y += float32(s.debugFont.fnt.Size[1]) * s.debugFont.yscl / s.heightScale
			s.debugFont.fnt.Print(drawTxt, *x, *y, s.debugFont.xscl/s.widthScale,
				s.debugFont.yscl/s.heightScale, 0, Rotation{0, 0, 0}, 0, 0, 0, 1, &s.scrrect,
				s.debugFont.palfx, s.debugFont.frgba)
		}
	}
	if s.debugDisplay {
		// Player Info on top of screen
		x := (320-float32(s.gameWidth))/2 + 1
		y := 240 - float32(s.gameHeight)
		if s.statusLFunc != nil {
			s.debugFont.SetColor(255, 255, 255, 255)
			for i, p := range s.chars {
				if len(p) > 0 {
					top := s.luaLState.GetTop()
					if s.luaLState.CallByParam(lua.P{Fn: s.statusLFunc, NRet: 1,
						Protect: true}, lua.LNumber(i+1)) == nil {
						l, ok := s.luaLState.Get(-1).(lua.LString)
						if ok && len(l) > 0 {
							put(&x, &y, string(l))
						}
					}
					s.luaLState.SetTop(top)
				}
			}
		}
		// Console
		y = Max(y, 48+240-float32(s.gameHeight))
		s.debugFont.SetColor(255, 255, 255, 255)
		for _, s := range s.consoleText {
			put(&x, &y, s)
		}
		// Data
		y = float32(s.gameHeight) - float32(s.debugFont.fnt.Size[1])*sys.debugFont.yscl/s.heightScale*
			(float32(len(s.listLFunc))+float32(s.cfg.Debug.ClipboardRows)) - 1*s.heightScale
		// Get debug char reference. Default to player 1 if out of bounds
		pn := s.debugRef[0]
		hn := s.debugRef[1]
		if pn >= len(s.chars) || hn >= len(s.chars[pn]) {
			s.debugRef[0] = 0
			s.debugRef[1] = 0
		}
		s.debugWC = s.chars[s.debugRef[0]][s.debugRef[1]]
		for i, f := range s.listLFunc {
			if f != nil {
				if i == 1 {
					s.debugFont.SetColor(199, 199, 219, 255)
				} else if (i == 2 && s.debugWC.animPN != s.debugWC.playerNo) ||
					(i == 3 && s.debugWC.ss.sb.playerNo != s.debugWC.playerNo) {
					s.debugFont.SetColor(255, 255, 127, 255)
				} else {
					s.debugFont.SetColor(255, 255, 255, 255)
				}
				top := s.luaLState.GetTop()
				if s.luaLState.CallByParam(lua.P{Fn: f, NRet: 1,
					Protect: true}) == nil {
					s, ok := s.luaLState.Get(-1).(lua.LString)
					if ok && len(s) > 0 {
						if i == 1 && (sys.debugWC == nil || sys.debugWC.csf(CSF_destroy)) { // TODO: A nil check this late would still crash
							put(&x, &y, string(s)+" disabled")
							break
						}
						put(&x, &y, string(s))
					}
				}
				s.luaLState.SetTop(top)
			}
		}
		// Clipboard
		s.debugFont.SetColor(255, 255, 255, 255)
		for _, s := range s.debugWC.clipboardText {
			put(&x, &y, s)
		}
	}
	// Draw Clsn text
	// Unlike Mugen, this is drawn separately from the Clsn boxes themselves, making debug more flexible
	//if s.clsnDisplay {
	for _, t := range s.debugClsnText {
		s.debugFont.SetColor(t.r, t.g, t.b, t.a)
		s.debugFont.fnt.Print(t.text, t.x, t.y, s.debugFont.xscl/s.widthScale,
			s.debugFont.yscl/s.heightScale, 0, Rotation{0, 0, 0}, 0, 0, 0, 0, &s.scrrect,
			s.debugFont.palfx, s.debugFont.frgba)
	}
	//}
}

// Starts and runs gameplay
// Called to start each match, on hard reset with shift+F4,
// and at the start of any round where a new character tags in for turns mode
func (s *System) runMatch() (reload bool) {
	// Reset variables
	s.matchTime = 0
	s.fightLoopEnd = false
	s.aiInput = [len(s.aiInput)]AiInput{}
	s.saveState = NewGameState()

	// Disable debug during netplay (but not during replays)
	if !s.debugModeAllowed() {
		s.debugDisplay = false
		s.clsnDisplay = false
		s.lifebarHide = false
		s.debugAccel = 1
	}

	// Defer resetting variables on return
	defer func() {
		s.oldNextAddTime = 1
		s.nomusic = false
		s.allPalFX.clear()
		s.allPalFX.enable = false
		for i, p := range s.chars {
			if len(p) > 0 {
				forceDestroy := s.matchOver() ||
					(s.tmode[i&1] == TM_Turns && p[0].teamside != -1 && p[0].life <= 0)
				s.clearPlayerAssets(i, forceDestroy)
			}
		}
	}()

	// Synchronize with external inputs (netplay, replays, etc)
	if err := s.synchronize(); err != nil {
		LogMessage(err.Error())
		s.esc = true
	}
	if s.netConnection != nil {
		defer func() {
			// Keep delay-netplay alive across Turns character swaps;
			// stopping here can desync nc.time before the next Synchronize().
			if s.matchOver() || s.esc || s.endMatch || s.gameEnd || s.fightLoopEnd {
				s.netConnection.Stop()
			}
		}()
	}

	// Setup characters
	s.SetupCharRoundStart()

	// Default debug/scripts to any found char
	s.debugWC = s.anyChar()
	debugInput := func() {
		select {
		case cl := <-s.commandLine:
			if err := s.luaLState.DoString(cl); err != nil {
				s.luaLState.RaiseError("Error during Debug Input Lua execution:\nCode: %s\nDetails: %s", cl, err.Error())
			}
		default:
		}
	}

	// Reset round state
	s.resetRound()

	// Make a first backup once everything is initialized
	s.roundBackup.Save()

	// Save round 1 backup separately
	if s.roundNo == 1 {
		s.matchBackup = s.roundBackup
	}

	if s.cfg.Config.TurnsLoading {
		s.startNextTurnsPreload()
		if (s.rollback.session != nil || s.cfg.Netplay.Rollback.DesyncTestFrames > 0) && !s.finishTurnsPreloadForRollback() {
			return false
		}
	}

	// Reset the clock right before entering the loop to ensure we start with a clean timeline
	// Handled by resetRound()
	//s.resetFrameTime()

	// Now switch to rollback if applicable
	// TODO: More merging so we don't hijack this function at all
	if s.rollback.session != nil || s.cfg.Netplay.Rollback.DesyncTestFrames > 0 {
		return s.rollback.hijackRunMatch(s)
	}

	// Loop until end of match
	for !s.endMatch {
		s.frameStepFlag = false

		for _, v := range s.shortcutScripts {
			if v.Activate {
				if err := s.luaLState.DoString(v.Script); err != nil {
					s.luaLState.RaiseError("Error executing Lua code: %s\n%v", v.Script, err.Error())
				}
			}
		}

		// Save/load state
		// TODO: Confirm at which exact point rollback does its own save/restore and match that
		if s.saveStateFlag {
			s.saveState.SaveState(0)
		} else if s.loadStateFlag {
			s.saveState.LoadState(0)
		}
		s.saveStateFlag = false
		s.loadStateFlag = false

		// If next round
		if !s.runNextRound() {
			break
		}

		// Update game state
		s.action()

		debugInput()

		if !s.addFrameTime(s.turbo) {
			if !s.eventUpdate() {
				return false
			}
			continue
		}

		if s.roundResetMatchStart && sys.tickCount == 1 {
			s.roundResetMatchStart = false
		}

		if s.matchResetFlg {
			s.matchResetFlg = false

			// If a member has already been changed in Turns mode
			// the char in memory is different from the char in the backup, so restoration with reload=0 is impossible
			canReset := true
			for i := 0; i < 2; i++ {
				if s.tmode[i] == TM_Turns && s.wins[i] > 0 {
					canReset = false
					break
				}
			}

			if !canReset {
				s.appendToConsole("Warning: MatchRestart with resetmatch=1 is disabled in Turns mode after character switch")
			} else {
				for i := 0; i < MaxPlayerNo; i++ {
					if s.reloadPreserveVars[i] {
						s.saveCharVars(i)
					}
				}
				s.roundNo = 1
				s.roundsExisted = [2]int32{}

				s.statsLog.abortMatch()
				s.statsLog.startMatch()

				// Recover the round 1 backup
				s.roundBackup = s.matchBackup
				s.roundBackup.Restore()
				s.resetRound()
			}
		}

		// F4 pressed to reset round
		if s.roundResetFlg && !s.postMatchFlg {
			for i := 0; i < MaxPlayerNo; i++ {
				if s.reloadPreserveVars[i] {
					s.saveCharVars(i)
				}
			}
			s.roundBackup.Restore()
			s.resetRound()
		}

		// Shift+F4 pressed to restart match
		if s.reloadFlg {
			return true
		}

		// Break if finished
		if s.fightLoopEnd && !s.postMatchFlg {
			break
		}

		if s.endMatch && !s.fightScreen.round.fadeOut.isActive() {
			break
		}

		// Render frame
		s.renderFrame()

		// Update system. Break if update returns false (engine shutdown).
		if !s.update() {
			break
		}

		// Exit the replay match loop before EOF can reuse the last input sample.
		if s.replayFile != nil && s.replayFile.file == nil {
			break
		}

		// If end match selected from menu
		if s.endMatch {
			if !s.motif.fadeOut.isActive() {
				break
			}
		}

		// If player pressed esc during netplay
		if s.esc && s.netConnection != nil {
			break
		}
	}

	return false
}

func (s *System) SetupCharRoundStart() {
	// Prepare next round for all players
	for _, p := range s.chars {
		if len(p) > 0 {
			p[0].prepareNextRound()
		}
	}

	// For power sharing, set maximum power to the highest one in the team
	if s.cfg.Options.Team.PowerShare {
		for i, p := range s.chars {
			if len(p) > 0 && p[0].teamside != -1 {
				pmax := Max(s.cgi[i&1].data.power, s.cgi[i].data.power)
				for j := i & 1; j < MaxSimul*2; j += 2 {
					if len(s.chars[j]) > 0 {
						s.chars[j][0].powerMax = pmax
					}
				}
			}
		}
	}

	// Calculate maximum life for all characters
	for i, p := range s.chars {
		if len(p) > 0 {
			// Get max life, and adjust based on team mode
			var lmax float32
			if p[0].ocd().lifeMax > 0 {
				lmax = float32(p[0].ocd().lifeMax)
			} else {
				lmax = float32(p[0].gi().data.life)
			}

			// Apply life options
			lmax *= s.cfg.Options.Life / 100

			// Adjust life by team mode
			if p[0].teamside != -1 {
				switch s.tmode[i&1] {
				case TM_Single:
					// Single mode gets the explicitly configured bonus
					switch s.tmode[(i+1)&1] {
					case TM_Simul, TM_Turns, TM_Tag:
						lmax *= s.cfg.Options.Team.SingleVsTeamLife / 100
					}
				case TM_Simul, TM_Tag:
					// For Simul/Tag life sharing, use average life of the team
					if s.cfg.Options.Team.LifeShare {
						totalTeamLife := float32(0)
						teamSize := 0
						for j := i & 1; j < MaxSimul*2; j += 2 {
							if len(s.chars[j]) > 0 {
								var charLm float32
								if s.chars[j][0].ocd().lifeMax > 0 {
									charLm = float32(s.chars[j][0].ocd().lifeMax) * s.cfg.Options.Life / 100
								} else {
									charLm = float32(s.chars[j][0].gi().data.life) * s.cfg.Options.Life / 100
								}
								totalTeamLife += charLm
								teamSize++
							}
						}
						if teamSize > 0 {
							lmax = totalTeamLife / float32(teamSize)
						}
					}
				case TM_Turns:
					// For Turns life sharing, divide life by number of characters
					if s.cfg.Options.Team.LifeShare && s.numTurns[i&1] > 0 {
						lmax /= float32(s.numTurns[i&1])
					}
				}
			}

			// Set lifemax
			p[0].lifeMax = Max(1, int32(math.Floor(float64(lmax))))
		}
	}

	// Initialize each character's dizzy and guard points
	for _, p := range s.chars {
		if len(p) > 0 {
			if p[0].ocd().dizzyPoints > 0 {
				p[0].dizzyPoints = p[0].ocd().dizzyPoints
			} else {
				p[0].dizzyPoints = p[0].dizzyPointsMax
			}
			if p[0].ocd().guardPoints > 0 {
				p[0].guardPoints = p[0].ocd().guardPoints
			} else {
				p[0].guardPoints = p[0].guardPointsMax
			}
		}
	}

	// Initialize each character's state
	for i, p := range s.chars {
		if len(p) > 0 {
			if p[0].roundsExisted() == 0 && (s.roundNo == 1 || s.tmode[i&1] == TM_Turns) {
				// If round 1 or a new character in Turns mode, initialize values
				if p[0].ocd().life != -1 {
					p[0].life = Clamp(p[0].ocd().life, 0, p[0].lifeMax)
					p[0].redLife = p[0].life
				} else {
					p[0].life = p[0].lifeMax
					p[0].redLife = p[0].lifeMax
				}
				if s.roundNo == 1 {
					if s.maxPowerMode {
						p[0].power = p[0].powerMax
					} else if p[0].ocd().power != -1 {
						p[0].power = Clamp(p[0].ocd().power, 0, p[0].powerMax)
					} else if !sys.sel.gameParams.PersistRounds || sys.consecutiveWins[0] == 0 {
						p[0].power = 0
					}
				}
				p[0].power = Clamp(p[0].power, 0, p[0].powerMax) // Because of previous partner in Turns mode
				p[0].dialogue = []string{}
				p[0].remapSpr = make(RemapPreset)
			}
		}
	}
}

func (s *System) runNextRound() bool {
	if s.roundOver() && !s.fightLoopEnd {
		if s.holdPostMatchForDialogue() {
			return true
		}
		s.clearAllSound()
		s.statsLog.nextRound()
		s.scoreRounds = append(s.scoreRounds, s.scorePoints)

		if !s.matchOver() &&
			!(s.tmode[0] == TM_Turns && s.effectiveLoss[0]) &&
			!(s.tmode[1] == TM_Turns && s.effectiveLoss[1]) {
			// Prepare for the next round
			s.roundNo++
			for i := range s.roundsExisted {
				s.roundsExisted[i]++
			}
			for i, p := range s.chars {
				if len(p) > 0 {
					if s.tmode[i&1] != TM_Turns || !s.effectiveLoss[i&1] {
						p[0].life = p[0].lifeMax
					} else if p[0].life <= 0 {
						p[0].life = 1
					}
					p[0].redLife = p[0].life // TODO: This doesn't truly need to be hardcoded
				}
			}
			s.resetRound()
			s.roundBackup.Save()
		} else {
			// End match, or prepare for a new character in turns mode
			for i, tm := range s.tmode {
				if s.chars[i][0].win() || (!s.chars[i][0].lose() && tm != TM_Turns) {
					for j := i; j < len(s.chars); j += 2 {
						if len(s.chars[j]) > 0 {
							if tm == TM_Turns && j != i {
								continue
							}
							if !s.chars[j][0].win() {
								s.chars[j][0].life = Max(1, s.cgi[j].data.life)
							}
						}
					}
				}
			}

			// If match isn't over, presumably this is turns mode,
			// so break to restart fight for the next character
			if !s.matchOver() {
				s.roundNo++
				for i := range s.roundsExisted {
					s.roundsExisted[i]++
				}
				return false
			}

			// Otherwise match is over
			s.postMatchFlg = true
			s.fightLoopEnd = true
			s.resetMatchData(false)
		}
	}

	// Not last round
	return true
}

func (s *System) gameLogicSpeed() int32 {
	base := int32(60 + s.cfg.Options.GameSpeed*s.cfg.Options.GameSpeedStep)
	spd := int32(float32(base) * s.debugAccel)
	return Max(1, spd)
}

func (s *System) gameRenderSpeed() int {
	var spd int32
	if !s.gameRunning || s.rollback.session != nil {
		// Currently, rendering the motif above 60fps breaks many things, so we'll patch it here
		// https://github.com/ikemen-engine/Ikemen-GO/issues/2131
		// Rollback is likewise locked to 60fps anyway, so we'll make it consistent here
		// TODO: Fix both properly. Motif could interpolate. Rollback should render at Framerate but sync at Gamespeed
		spd = 60
	} else {
		spd = int32(s.cfg.Video.Framerate)
	}
	return int(Max(1, spd))
}

func (s *System) debugModeAllowed() bool {
	if s.netConnection != nil || s.rollback.session != nil {
		return false
	}
	return s.cfg.Debug.AllowDebugMode
}

// Returns true if the current frame is inside a rollback
func (s *System) inRollback() bool {
	if s.rollback.session != nil {
		return s.rollback.session.inRollback
	}
	return false
}

func (s *System) isSpeculativeFrame() bool {
	if s.rollback.session != nil {
		return !s.rollback.session.inRollback
	}
	return false
}

// The backup to be used when F4 is pressed
type RoundStartBackup struct {
	charBackup    [MaxPlayerNo][]Char
	cgiBackup     [MaxPlayerNo]CharGlobalInfo
	stageBackup   Stage
	oldWins       [2]int32
	oldDraws      int32
	oldTeamLeader [2]int
	lastCharId    int32
}

func (bk *RoundStartBackup) Save() {
	// Save characters
	// We save helpers as well because of "preserve" parameter
	for i, chars := range sys.chars {
		if len(chars) == 0 {
			continue
		}

		// Allocate slice for backup
		bk.charBackup[i] = make([]Char, 0, len(chars))

		for _, c := range chars {
			// Shallow copy whole struct
			bkup := *c

			// Deep copy maps
			bkup.cnsvar = make(map[int32]int32, len(c.cnsvar))
			for k, v := range c.cnsvar {
				bkup.cnsvar[k] = v
			}
			bkup.cnsfvar = make(map[int32]float32, len(c.cnsfvar))
			for k, v := range c.cnsfvar {
				bkup.cnsfvar[k] = v
			}
			bkup.mapArray = make(map[string]float32, len(c.mapArray))
			for k, v := range c.mapArray {
				bkup.mapArray[k] = v
			}

			// Deep copy dialogue slice
			bkup.dialogue = append([]string{}, c.dialogue...)

			// Deep copy remap preset
			bkup.remapSpr = make(RemapPreset)
			for k, v := range c.remapSpr {
				bkup.remapSpr[k] = v
			}

			bk.charBackup[i] = append(bk.charBackup[i], bkup)
		}
	}

	// CharGlobalInfo backup
	for i := range sys.cgi {
		bk.cgiBackup[i] = sys.cgi[i]
	}

	// Stage backup
	bk.stageBackup = *sys.stage

	// Deep copy stage maps/slices
	bk.stageBackup.constants = make(map[string]float32, len(sys.stage.constants))
	for k, v := range sys.stage.constants {
		bk.stageBackup.constants[k] = v
	}
	bk.stageBackup.attachedchardef = append([]string{}, sys.stage.attachedchardef...)

	// Match info
	bk.oldWins, bk.oldDraws = sys.wins, sys.draws
	bk.oldTeamLeader = sys.teamLeader

	// Save last char ID so F4 won't keep incrementing helper ID's
	bk.lastCharId = sys.lastCharId
}

func (bk *RoundStartBackup) Restore() {
	// Restore characters
	for i, chars := range sys.chars {
		if len(chars) == 0 {
			continue
		}

		// Staged Turns preload can create standby P3/P4 after roundBackup.Save()
		if len(bk.charBackup[i]) == 0 {
			sys.clearPlayerAssets(i, true)
			sys.removePlayerFromCharList(i)
			sys.chars[i] = nil
			continue
		}

		for j, c := range chars {
			// Find the backup corresponding to this index
			var bkup *Char
			for k := range bk.charBackup[i] {
				if bk.charBackup[i][k].helperIndex == c.helperIndex {
					bkup = &bk.charBackup[i][k]
					break
				}
			}

			// Safeguard: if no backup exists for this slot and it’s not the root, destroy the helper
			if bkup == nil {
				if j != 0 && c.helperIndex != 0 {
					c.destroy()
				}
				continue
			}

			// We no longer need the live sounds workaround because sounds now belong to system
			// Save live sounds before overwriting
			//liveSounds := c.soundChannels

			// Restore shallow copy from backup
			*c = *bkup

			// Restore live sounds
			//c.soundChannels = liveSounds

			// Remake the CNS variable maps
			// Then restore only var and fvar (losing sysvar and sysfvar)
			c.initCnsVar()
			for k, v := range bkup.cnsvar {
				c.cnsvar[k] = v
			}
			for k, v := range bkup.cnsfvar {
				c.cnsfvar[k] = v
			}

			// Restore maps
			c.mapArray = make(map[string]float32, len(bkup.mapArray))
			for k, v := range bkup.mapArray {
				c.mapArray[k] = v
			}

			c.dialogue = append([]string{}, bkup.dialogue...)

			c.remapSpr = make(RemapPreset)
			for k, v := range bkup.remapSpr {
				c.remapSpr[k] = v
			}
		}
	}

	// Restore CharGlobalInfo
	for i := range sys.cgi {
		sys.cgi[i] = bk.cgiBackup[i]
	}

	// Restore stage
	// We preserve the stage time as a cosmetic thing just to match Mugen. Might as well restore it to where it was when we use F4
	// NOTE: If reloading stage time we'd need backups of the BGCtrls as well
	// NOTE: This save and restore of stage variables makes ModifyStageVar not persist. Maybe that should not be the case?
	stageTime := sys.stage.stageTime
	*sys.stage = bk.stageBackup
	sys.stage.stageTime = stageTime

	// Restore stage maps/slices
	sys.stage.constants = make(map[string]float32, len(bk.stageBackup.constants))
	for k, v := range bk.stageBackup.constants {
		sys.stage.constants[k] = v
	}
	sys.stage.attachedchardef = append([]string{}, bk.stageBackup.attachedchardef...)

	// Restore match info
	sys.wins, sys.draws = bk.oldWins, bk.oldDraws
	sys.teamLeader = bk.oldTeamLeader

	// Restore other info
	sys.lastCharId = bk.lastCharId
}

type SelectChar struct {
	def           string
	name          string
	lifebarname   string
	author        string
	sound         string
	intro         string
	ending        string
	arcadepath    string
	pal           []int32
	pal_defaults  []int32
	pal_keymap    []int32
	pal_files     []string
	localcoord    [2]int32
	portraitscale float32
	cns_scale     [2]float32
	anims         PreloadedAnims
	sff           *Sff
	music         Music
	scp           *SelectCharParams
	preloadAnim   string
	preloadSprite string
}

func newSelectChar() *SelectChar {
	return &SelectChar{
		localcoord:    [2]int32{320, 240},
		portraitscale: 1,
		cns_scale:     [2]float32{1, 1},
		anims:         NewPreloadedAnims(),
		music:         make(Music),
		scp:           newSelectCharParams(),
	}
}

type SelectStage struct {
	def             string
	name            string
	attachedchardef []string
	localcoord      [2]int32
	portraitscale   float32
	anims           PreloadedAnims
	sff             *Sff
	music           Music
	ssp             *SelectStageParams
	preloadSprite   string
}

func newSelectStage() *SelectStage {
	return &SelectStage{
		localcoord:    [2]int32{320, 240},
		portraitscale: 1,
		anims:         NewPreloadedAnims(),
		music:         make(Music),
		ssp:           newSelectStageParams(),
	}
}

type Select struct {
	charlist           []SelectChar
	stagelist          []SelectStage
	selected           [2][][2]int
	selectedStageNo    int
	charAnimPreload    map[int32]bool
	stageAnimPreload   map[int32]bool
	charSpritePreload  map[[2]uint16]bool
	stageSpritePreload map[[2]uint16]bool
	cdefOverwrite      map[int]string
	palOverwrite       map[int]int
	sdefOverwrite      string
	music              Music
	gameParams         *GameParams
	preloadMu          sync.Mutex
	preloadCond        *sync.Cond
	preloadWorkerRun   bool
	preloadOrder       int64
	charPreload        []PreloadEntry
	stagePreload       []PreloadEntry
}

type PreloadState int32

const (
	PS_Idle PreloadState = iota
	PS_Queued
	PS_Loading
	PS_Ready
	//PS_Error // We simply crash when preloading invalid assets
)

func (ps PreloadState) String() string {
	switch ps {
	case PS_Idle:
		return "idle"
	case PS_Queued:
		return "queued"
	case PS_Loading:
		return "loading"
	case PS_Ready:
		return "ready"
		// case PS_Error:
		// return "error"
	}
	return "idle"
}

const (
	PreloadPriorityLow  = 1
	PreloadPriorityHigh = 2
)

type PreloadEntry struct {
	State    PreloadState
	Priority int
	Order    int64
	//Err      string
}

func newSelect() *Select {
	return &Select{
		selectedStageNo:  -1,
		charAnimPreload:  make(map[int32]bool),
		stageAnimPreload: make(map[int32]bool),
		charSpritePreload: map[[2]uint16]bool{[...]uint16{9000, 0}: true,
			[...]uint16{9000, 1}: true},
		stageSpritePreload: make(map[[2]uint16]bool),
		palOverwrite:       make(map[int]int),
		cdefOverwrite:      make(map[int]string),
		music:              make(Music),
		gameParams:         newGameParams(),
	}
}

func (s *Select) initPreloadSync() {
	s.preloadMu.Lock()
	defer s.preloadMu.Unlock()
	if s.preloadCond == nil {
		s.preloadCond = sync.NewCond(&s.preloadMu)
	}
}

func (s *Select) ensurePreloadWorker() {
	s.initPreloadSync()
	s.preloadMu.Lock()
	if s.preloadWorkerRun {
		s.preloadMu.Unlock()
		return
	}
	s.preloadWorkerRun = true
	s.preloadMu.Unlock()
	SafeGo(func() {
		s.preloadWorkerLoop()
	})
}

func (s *Select) ensureCharPreloadSlot(ref int) {
	if ref < 0 {
		return
	}
	for len(s.charPreload) <= ref {
		s.charPreload = append(s.charPreload, PreloadEntry{})
	}
}

func (s *Select) ensureStagePreloadSlot(ref int) {
	if ref <= 0 {
		return
	}
	idx := ref - 1
	for len(s.stagePreload) <= idx {
		s.stagePreload = append(s.stagePreload, PreloadEntry{})
	}
}

func (s *Select) QueueCharPreload(ref int, priority int) {
	if sys.cfg.Config.BootLoadingMode == 0 {
		return
	}
	if ref < 0 || ref >= len(s.charlist) {
		return
	}
	s.initPreloadSync()
	s.ensurePreloadWorker()
	s.preloadMu.Lock()
	defer func() {
		s.preloadCond.Broadcast()
		s.preloadMu.Unlock()
	}()
	s.ensureCharPreloadSlot(ref)
	e := &s.charPreload[ref]
	if e.State == PS_Ready {
		return
	}
	if e.State == PS_Loading {
		e.Priority = priority
		return
	}
	if e.State == PS_Idle { // || e.State == PS_Error {
		e.State = PS_Queued
	}
	if e.Priority == priority {
		return
	}
	e.Priority = priority
	if priority <= PreloadPriorityLow {
		e.Order = 0
	} else {
		s.preloadOrder++
		e.Order = s.preloadOrder
	}
}

func (s *Select) QueueStagePreload(ref int, priority int) {
	if sys.cfg.Config.BootLoadingMode == 0 {
		return
	}
	if ref <= 0 || ref > len(s.stagelist) {
		return
	}
	s.initPreloadSync()
	s.ensurePreloadWorker()
	s.preloadMu.Lock()
	defer func() {
		s.preloadCond.Broadcast()
		s.preloadMu.Unlock()
	}()
	s.ensureStagePreloadSlot(ref)
	e := &s.stagePreload[ref-1]
	if e.State == PS_Ready {
		return
	}
	if e.State == PS_Loading {
		e.Priority = priority
		return
	}
	if e.State == PS_Idle { // || e.State == PS_Error {
		e.State = PS_Queued
	}
	if e.Priority == priority {
		return
	}
	e.Priority = priority
	if priority <= PreloadPriorityLow {
		e.Order = 0
	} else {
		s.preloadOrder++
		e.Order = s.preloadOrder
	}
}

func (s *Select) CharPreloadStatus(ref int) PreloadState { // (PreloadState, string) {
	if sys.cfg.Config.BootLoadingMode == 0 {
		return PS_Ready //, ""
	}
	s.initPreloadSync()
	s.preloadMu.Lock()
	defer s.preloadMu.Unlock()
	if ref < 0 || ref >= len(s.charPreload) {
		return PS_Idle //, ""
	}
	return s.charPreload[ref].State //, s.charPreload[ref].Err
}

func (s *Select) StagePreloadStatus(ref int) PreloadState { // (PreloadState, string) {
	if sys.cfg.Config.BootLoadingMode == 0 {
		return PS_Ready //, ""
	}
	s.initPreloadSync()
	s.preloadMu.Lock()
	defer s.preloadMu.Unlock()
	if ref <= 0 || ref-1 >= len(s.stagePreload) {
		return PS_Idle //, ""
	}
	return s.stagePreload[ref-1].State //, s.stagePreload[ref-1].Err
}

func (s *Select) AllPreloadsReady() bool {
	if sys.cfg.Config.BootLoadingMode == 0 {
		return true
	}
	s.initPreloadSync()
	s.preloadMu.Lock()
	defer s.preloadMu.Unlock()
	for _, e := range s.charPreload {
		switch e.State {
		case PS_Queued, PS_Loading:
			return false
		}
	}
	for _, e := range s.stagePreload {
		switch e.State {
		case PS_Queued, PS_Loading:
			return false
		}
	}
	return true
}

func (s *Select) preloadWorkerLoop() {
	s.initPreloadSync()
	for {
		s.preloadMu.Lock()
		for {
			kind := ""
			ref := -1
			bestPriority := -1
			bestOrder := int64(-1)
			for i, e := range s.charPreload {
				if e.State != PS_Queued {
					continue
				}
				if e.Priority > bestPriority || (e.Priority == bestPriority && e.Order > bestOrder) {
					bestPriority = e.Priority
					bestOrder = e.Order
					kind = "char"
					ref = i
				}
			}
			for i, e := range s.stagePreload {
				if e.State != PS_Queued {
					continue
				}
				if e.Priority > bestPriority || (e.Priority == bestPriority && e.Order > bestOrder) {
					bestPriority = e.Priority
					bestOrder = e.Order
					kind = "stage"
					ref = i + 1
				}
			}
			if ref >= 0 {
				if kind == "char" {
					s.charPreload[ref].State = PS_Loading
					//s.charPreload[ref].Err = ""
				} else {
					s.stagePreload[ref-1].State = PS_Loading
					//s.stagePreload[ref-1].Err = ""
				}
				s.preloadMu.Unlock()

				var err error
				if kind == "char" {
					err = s.preloadCharAssets(ref)
				} else {
					err = s.preloadStageAssets(ref)
				}

				s.preloadMu.Lock()

				if err != nil {
					panic(fmt.Sprintf("Preloading error: %s", err))
				} else if kind == "char" {
					s.charPreload[ref].State = PS_Ready
				} else {
					s.stagePreload[ref-1].State = PS_Ready
				}
				s.preloadCond.Broadcast()
				break
			}
			s.preloadCond.Wait()
		}
		s.preloadMu.Unlock()
	}
}

func (s *Select) preloadCharAssets(ref int) error {
	if ref < 0 || ref >= len(s.charlist) {
		return nil
	}
	sc := s.charlist[ref]
	localAnims := NewPreloadedAnims()
	listSpr := make(map[[2]uint16]bool)
	for k := range s.charSpritePreload {
		listSpr[k] = true
	}
	tempSff := newSff()

	if sc.preloadAnim != "" {
		animPath := sc.preloadAnim
		if err := LoadFile(&animPath, []string{sc.def}, "", func(filename string) error {
			str, err := LoadText(filename)
			if err != nil {
				return err
			}
			lines, lnidx := SplitAndTrim(str, "\n"), 0
			at := ReadAnimationTable(sc.def, tempSff, &tempSff.palList, lines, &lnidx, false)
			for vAnim := range s.charAnimPreload {
				if animation := at.get(vAnim); animation != nil {
					localAnims.addAnim(animation, vAnim)
					for _, fr := range animation.frames {
						if fr.Group < 0 || fr.Number < 0 {
							continue
						}
						listSpr[[2]uint16{uint16(fr.Group), uint16(fr.Number)}] = true
					}
				}
			}
			return nil
		}); err != nil {
			return err
		}
	}

	var localSff *Sff
	if sc.preloadSprite != "" {
		spritePath := sc.preloadSprite
		var selPal []int32
		if err := LoadFile(&spritePath, []string{sc.def, "", "data/"}, "", func(file string) error {
			var err error
			localSff, selPal, err = preloadSff(file, true, listSpr)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		localAnims.updateSff(localSff)
		for k := range s.charSpritePreload {
			localAnims.addSprite(localSff, k[0], k[1])
		}
		// Determine the character's selectable palettes based on the return from preloadSff()
		if localSff.header.Version[0] != 1 {
			// Merge SFF and ACT indexes
			defPals := make(map[int32]bool)
			for _, p := range sc.pal {
				defPals[p] = true
			}
			maxPalSlots := sys.cfg.Config.PaletteMax
			newPalFiles := make([]string, maxPalSlots)
			for i, pIdx := range sc.pal {
				if pIdx >= 1 && int(pIdx) <= maxPalSlots && i < len(sc.pal_files) {
					newPalFiles[pIdx-1] = sc.pal_files[i]
				}
			}
			newPal := make([]int32, 0, maxPalSlots)
			newPalFilesOut := make([]string, 0, maxPalSlots)
			for i := 1; i <= maxPalSlots; i++ {
				existsInSff := false
				for _, sIdx := range selPal {
					if sIdx == int32(i) {
						existsInSff = true
						break
					}
				}
				if defPals[int32(i)] || existsInSff {
					newPal = append(newPal, int32(i))
					newPalFilesOut = append(newPalFilesOut, newPalFiles[i-1])
				}
			}
			sc.pal = newPal
			sc.pal_files = newPalFilesOut
		} else if len(sc.pal) == 0 {
			// Just use selPal directly
			sc.pal = selPal
		}
	} else {
		// Make a dummy SFF
		localSff = newSff()
		localAnims.updateSff(localSff)
		for k := range s.charSpritePreload {
			localAnims.addSprite(localSff, k[0], k[1])
		}
	}

	s.preloadMu.Lock()
	dst := &s.charlist[ref]
	dst.sff = localSff
	dst.anims = localAnims
	dst.pal = sc.pal
	dst.pal_files = sc.pal_files
	s.preloadMu.Unlock()
	return nil
}

func (s *Select) preloadStageAssets(ref int) error {
	if ref <= 0 || ref > len(s.stagelist) {
		return nil
	}
	ss := s.stagelist[ref-1]
	lines := []string{}
	defPath := ss.def
	if err := LoadFile(&defPath, nil, "", func(file string) error {
		str, err := LoadText(file)
		if err != nil {
			return err
		}
		lines = SplitAndTrim(str, "\n")
		return nil
	}); err != nil {
		return err
	}

	localAnims := NewPreloadedAnims()
	listSpr := make(map[[2]uint16]bool)
	for k := range s.stageSpritePreload {
		listSpr[k] = true
	}

	tempSff := newSff()
	atidx := 0
	at := ReadAnimationTable(ss.def, tempSff, &tempSff.palList, lines, &atidx, false)
	for v := range s.stageAnimPreload {
		if anim := at.get(v); anim != nil {
			localAnims.addAnim(anim, v)
			for _, fr := range anim.frames {
				if fr.Group >= 0 && fr.Number >= 0 {
					listSpr[[2]uint16{uint16(fr.Group), uint16(fr.Number)}] = true
				}
			}
		}
	}

	var localSff *Sff
	if ss.preloadSprite != "" {
		spritePath := ss.preloadSprite
		if err := LoadFile(&spritePath, []string{ss.def, "", "data/"}, "", func(file string) error {
			var err error
			localSff, _, err = preloadSff(file, false, listSpr)
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			return err
		}
		localAnims.updateSff(localSff)
		for k := range s.stageSpritePreload {
			localAnims.addSprite(localSff, k[0], k[1])
		}
	} else {
		localSff = newSff()
		localAnims.updateSff(localSff)
	}

	s.preloadMu.Lock()
	dst := &s.stagelist[ref-1]
	dst.sff = localSff
	dst.anims = localAnims
	s.preloadMu.Unlock()
	return nil
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

// Validates a palette index for the palette select
func (s *Select) ValidatePalette(charRef, requested int) int {
	if charRef < 0 || charRef >= len(s.charlist) {
		return 1
	}
	sc := &s.charlist[charRef]
	if len(sc.pal) == 0 {
		return 1
	}
	// If the requested index exists, return it
	for _, real := range sc.pal {
		if int(real) == requested {
			return requested
		}
	}
	// Otherwise, return the next valid one (circular)
	for _, real := range sc.pal {
		if int(real) > requested {
			return int(real)
		}
	}
	// Fallback: return the first available
	return int(sc.pal[0])
}

func (s *Select) SelectStage(n int) {
	sys.selMutex.Lock()
	s.selectedStageNo = n
	sys.selMutex.Unlock()
}

func (s *Select) GetStage(n int) *SelectStage {
	if len(s.stagelist) == 0 {
		return nil
	}
	n %= len(s.stagelist) + 1
	if n < 0 {
		n += len(s.stagelist) + 1
	}
	return &s.stagelist[n-1]
}

func getDefaultDefPathInZip(zipFilePathOnDisk string) (path1 string, path2 string) {
	zipBaseName := LowercaseNoExtension(filepath.Base(zipFilePathOnDisk))
	path1 = zipBaseName + ".def"
	path2 = filepath.ToSlash(filepath.Join(zipBaseName, zipBaseName+".def"))
	return path1, path2
}

func (s *Select) AddChar(def string) *SelectChar {
	tnow := time.Now()
	s.charlist = append(s.charlist, *newSelectChar())
	sc := &s.charlist[len(s.charlist)-1]

	parts := strings.Split(def, ",")
	defPathFromSelect := strings.TrimSpace(parts[0])
	defPathFromSelect = filepath.ToSlash(defPathFromSelect)

	// Early exits for known keywords
	if strings.ToLower(defPathFromSelect) == "randomselect" {
		sc.def, sc.name = "randomselect", "Random"
		return nil
	}
	if strings.ToLower(defPathFromSelect) == "dummyslot" {
		sc.name = "dummyslot"
		return nil
	}
	if strings.ToLower(defPathFromSelect) == "skipslot" {
		sc.name = "skipslot"
		return nil
	}

	// Helper to set missing characters to dummy slots and (always) print a warning
	useDummy := func(reason string) *SelectChar {
		sc.name = "dummyslot"
		LogMessage("Failed to add char: %v (%s)", defPathFromSelect, reason)
		return nil
	}

	// Print a message for every loaded character. Just muted at the moment
	defer func() {
		tstr := fmt.Sprintf("Char: %v", defPathFromSelect)
		sys.loadTime(tnow, tstr, false, false)
	}()

	var finalDefPath string
	isZipChar := strings.HasSuffix(strings.ToLower(defPathFromSelect), ".zip")

	if isZipChar {
		zipSearchDirs := []string{"chars/", "data/", ""}
		var actualZipPathOnDisk string

		if filepath.IsAbs(defPathFromSelect) {
			if foundPath := FileExist(defPathFromSelect); foundPath != "" && strings.HasSuffix(strings.ToLower(foundPath), ".zip") {
				actualZipPathOnDisk = foundPath
			}
		} else {
			for _, dir := range zipSearchDirs {
				candidateZipPath := filepath.ToSlash(filepath.Join(dir, defPathFromSelect))
				if foundPath := FileExist(candidateZipPath); foundPath != "" && strings.HasSuffix(strings.ToLower(foundPath), ".zip") {
					actualZipPathOnDisk = foundPath
					break
				}
			}
		}

		if actualZipPathOnDisk == "" {
			return useDummy("ZIP not found")
		}

		defInZip1, defInZip2 := getDefaultDefPathInZip(actualZipPathOnDisk)

		// Construct logical paths for FileExist to check *inside* the zip
		candidateLogicalPath1 := filepath.ToSlash(actualZipPathOnDisk + "/" + defInZip1)
		if FileExist(candidateLogicalPath1) != "" { // FileExist checks inside the zip now
			finalDefPath = candidateLogicalPath1
		} else {
			candidateLogicalPath2 := filepath.ToSlash(actualZipPathOnDisk + "/" + defInZip2)
			if FileExist(candidateLogicalPath2) != "" {
				finalDefPath = candidateLogicalPath2
			} else {
				return useDummy(fmt.Sprintf("DEF in ZIP missing: %s or %s", defInZip1, defInZip2))
			}
		}
	} else {
		charDefPathGuess := defPathFromSelect
		if !strings.HasSuffix(strings.ToLower(charDefPathGuess), ".def") {
			if !strings.Contains(charDefPathGuess, "/") {
				baseName := filepath.Base(charDefPathGuess)
				charDefPathGuess = filepath.ToSlash(filepath.Join(charDefPathGuess, baseName+".def"))
			} else {
				charDefPathGuess += ".def"
			}
		}

		foundDiskPath := SearchFile(charDefPathGuess, []string{"", "data/"}, "chars/")
		if foundDiskPath == "" || !strings.HasSuffix(strings.ToLower(foundDiskPath), ".def") {
			return useDummy("DEF not found")
		}
		finalDefPath = foundDiskPath
	}

	sc.def = finalDefPath
	if sc.def == "" {
		return useDummy("Empty DEF path")
	}

	charDefContent, err := LoadText(sc.def)
	if err != nil {
		// Intercept simple "file not found" errors so the message isn't too long
		if os.IsNotExist(err) {
			return useDummy("DEF not found")
		}
		// Print full message if it's an actual read error
		return useDummy("DEF read error: " + err.Error())
	}

	var cns_orig, sprite_orig, anim_orig string
	var fnt_orig [10][2]string

	lines, lnidx := SplitAndTrim(charDefContent, "\n"), 0
	info, files, keymap, arcade := true, true, true, true
	lanInfo, lanFiles, lanKeymap, lanArcade := true, true, true, true
	langPrefix := sys.cfg.Config.Language + "."

	for lnidx < len(lines) {
		isec, name, subname := ReadIniSection(lines, &lnidx)

		isLan := strings.HasPrefix(name, langPrefix)
		baseName := name
		if isLan {
			baseName = name[len(langPrefix):]
		}

		switch baseName {
		case "info":
			if (isLan && lanInfo) || (!isLan && info) {
				if isLan {
					lanInfo = false
				}
				info = false

				var ok bool
				if sc.name, ok, _ = isec.getText("displayname"); !ok {
					sc.name, _, _ = isec.getText("name")
				}
				if sc.lifebarname, ok, _ = isec.getText("lifebarname"); !ok {
					sc.lifebarname = sc.name
				}
				sc.author, _, _ = isec.getText("author")
				sc.pal_defaults = isec.readI32CsvForStage("pal.defaults")
				isec.ReadI32("localcoord", &sc.localcoord[0], &sc.localcoord[1])
				isec.ReadF32("portraitscale", &sc.portraitscale)
			}

		case "files":
			if (isLan && lanFiles) || (!isLan && files) {
				if isLan {
					lanFiles = false
				}
				files = false

				cns_orig = decodeShiftJIS(isec["cns"])
				sprite_orig = decodeShiftJIS(isec["sprite"])
				anim_orig = decodeShiftJIS(isec["anim"])
				sc.sound = decodeShiftJIS(isec["sound"])

				// Clear and rebuild palettes to ensure localized files can overwrite defaults
				sc.pal = nil
				sc.pal_files = nil
				for pIdx := 1; pIdx <= sys.cfg.Config.PaletteMax; pIdx++ {
					if palFile := isec[fmt.Sprintf("pal%v", pIdx)]; palFile != "" {
						sc.pal = append(sc.pal, int32(pIdx))
						sc.pal_files = append(sc.pal_files, palFile)
					}
				}

				for fIdx := range fnt_orig {
					fnt_orig[fIdx][0] = isec[fmt.Sprintf("font%v", fIdx)]
					fnt_orig[fIdx][1] = isec[fmt.Sprintf("fnt_height%v", fIdx)]
				}
			}

		case "palette ": // Note space
			isKeymap := len(subname) >= 6 && strings.ToLower(subname[:6]) == "keymap"
			if isKeymap && ((isLan && lanKeymap) || (!isLan && keymap)) {
				if isLan {
					lanKeymap = false
				}
				keymap = false

				sc.pal_keymap = make([]int32, 12)
				for kIdx, v := range [12]string{"a", "b", "c", "x", "y", "z",
					"a2", "b2", "c2", "x2", "y2", "z2"} {
					var i32 int32
					if isec.ReadI32(v, &i32) {
						sc.pal_keymap[kIdx] = i32
					} else {
						sc.pal_keymap[kIdx] = int32(kIdx + 1) // default
					}
				}
			}

		case "arcade":
			if (isLan && lanArcade) || (!isLan && arcade) {
				if isLan {
					lanArcade = false
				}
				arcade = false

				sc.intro, _, _ = isec.getText("intro.storyboard")
				sc.ending, _, _ = isec.getText("ending.storyboard")
				sc.arcadepath, _, _ = isec.getText("arcadepath")
			}
		}
	}

	LoadFile(&cns_orig, []string{sc.def, "", "data/"}, "", func(filename string) error {
		str, err := LoadText(filename)
		if err != nil {
			return err
		}
		lines, lnidx := SplitAndTrim(str, "\n"), 0
		for lnidx < len(lines) {
			is, name, _ := ReadIniSection(lines, &lnidx)
			switch name {
			case "size":
				if ok := is.ReadF32("xscale", &sc.cns_scale[0]); !ok {
					sc.cns_scale[0] = 320 / float32(sc.localcoord[0])
				}
				if ok := is.ReadF32("yscale", &sc.cns_scale[1]); !ok {
					sc.cns_scale[1] = 320 / float32(sc.localcoord[0])
				}
				return nil
			}
		}
		return nil
	})

	// preload animations
	if len(anim_orig) > 0 {
		sc.preloadAnim = anim_orig
	}

	// Try to use the "_preload.sff" file if available
	fp := fmt.Sprintf("%v_preload.sff", strings.TrimSuffix(sc.def, filepath.Ext(sc.def)))
	if fp = FileExist(fp); len(fp) == 0 {
		// Fall back to normal SFF
		fp = sprite_orig
	}

	// preload portion of sff file
	if len(fp) > 0 {
		sc.preloadSprite = fp
	}
	if sys.cfg.Config.BootLoadingMode > 0 {
		sc.sff = newSff()
		sc.anims.updateSff(sc.sff)
		for k := range s.charSpritePreload {
			sc.anims.addSprite(sc.sff, k[0], k[1])
		}
	} else {
		if err := s.preloadCharAssets(len(s.charlist) - 1); err != nil {
			panic(fmt.Errorf("failed to preload %v: %v", sc.def, err))
		}
	}

	return sc
}

func (s *Select) AddStage(def string) (*SelectStage, error) {
	var tstr string
	tnow := time.Now()

	// Print a message for every loaded stage. Just muted at the moment
	defer func() {
		sys.loadTime(tnow, tstr, false, false)
	}()

	defPathFromSelect := filepath.ToSlash(def)
	tstr = fmt.Sprintf("Stage added: %v", defPathFromSelect)

	var finalDefPath string
	isZipStage := strings.HasSuffix(strings.ToLower(defPathFromSelect), ".zip")

	if isZipStage {
		zipSearchDirs := []string{"stages/", "data/", ""}
		var actualZipPathOnDisk string

		if filepath.IsAbs(defPathFromSelect) {
			if foundPath := FileExist(defPathFromSelect); foundPath != "" && strings.HasSuffix(strings.ToLower(foundPath), ".zip") {
				actualZipPathOnDisk = foundPath
			}
		} else {
			for _, dir := range zipSearchDirs {
				candidateZipPath := filepath.ToSlash(filepath.Join(dir, defPathFromSelect))
				if foundPath := FileExist(candidateZipPath); foundPath != "" && strings.HasSuffix(strings.ToLower(foundPath), ".zip") {
					actualZipPathOnDisk = foundPath
					break
				}
			}
		}

		if actualZipPathOnDisk == "" {
			err := fmt.Errorf("stage ZIP not found: %s", defPathFromSelect)
			LogMessage("Failed to add stage. File not found: %v", defPathFromSelect)
			return nil, err
		}

		defInZip1, defInZip2 := getDefaultDefPathInZip(actualZipPathOnDisk)

		candidateLogicalPath1 := filepath.ToSlash(actualZipPathOnDisk + "/" + defInZip1)
		if FileExist(candidateLogicalPath1) != "" {
			finalDefPath = candidateLogicalPath1
		} else {
			candidateLogicalPath2 := filepath.ToSlash(actualZipPathOnDisk + "/" + defInZip2)
			if FileExist(candidateLogicalPath2) != "" {
				finalDefPath = candidateLogicalPath2
			} else {
				err := fmt.Errorf("DEF file not found in ZIP: %s or %s", defInZip1, defInZip2)
				LogMessage("Failed to add stage. DEF file not found in %v: %v or %v", defPathFromSelect, defInZip1, defInZip2)
				return nil, err
			}
		}
	} else {
		if !strings.HasSuffix(strings.ToLower(def), ".def") {
			def += ".def"
		}
		if err := LoadFile(&def, []string{"", "data/"}, "stages/", func(file string) error {
			finalDefPath = file
			return nil
		}); err != nil {
			LogMessage("Failed to add stage. File not found: %v", def)
			return nil, err
		}
	}

	var lines []string
	var err error
	if err = LoadFile(&finalDefPath, nil, "", func(file string) error {
		var str string
		str, err = LoadText(file)
		if err != nil {
			return err
		}
		lines = SplitAndTrim(str, "\n")
		return nil
	}); err != nil {
		LogMessage("Failed to add stage. File read error: %s: %v", finalDefPath, err)
		return nil, err
	}

	tstr = fmt.Sprintf("Stage added: %v", finalDefPath)

	lnidx, info, bgdef, stageinfo := 0, true, true, true
	lanInfo, lanBgdef, lanStageinfo := true, true, true
	langPrefix := sys.cfg.Config.Language + "."
	var spr string

	s.stagelist = append(s.stagelist, *newSelectStage())
	ss := &s.stagelist[len(s.stagelist)-1]
	ss.def = finalDefPath

	for lnidx < len(lines) {
		isec, name, _ := ReadIniSection(lines, &lnidx)

		isLan := strings.HasPrefix(name, langPrefix)
		baseName := name
		if isLan {
			baseName = name[len(langPrefix):]
		}

		switch baseName {
		case "info":
			if (isLan && lanInfo) || (!isLan && info) {
				if isLan {
					lanInfo = false
				}
				info = false

				var ok bool
				if ss.name, ok, _ = isec.getText("displayname"); !ok {
					if ss.name, ok, _ = isec.getText("name"); !ok {
						ss.name = def
					}
				}

				for idx := 0; idx < MaxAttachedChar; idx++ {
					key := "attachedchar"
					if idx > 0 {
						key += fmt.Sprint(idx + 1) // attachedchar2, attachedchar3, attachedchar4
					}

					if err := isec.LoadFile(key, []string{finalDefPath, "", "data/"}, "chars/", func(filename string) error {
						// Ensure slice has correct length
						for len(ss.attachedchardef) <= idx {
							ss.attachedchardef = append(ss.attachedchardef, "")
						}
						ss.attachedchardef[idx] = filename
						return nil
					}); err != nil {
						continue
					}
				}
			}

		case "bgdef":
			if (isLan && lanBgdef) || (!isLan && bgdef) {
				if isLan {
					lanBgdef = false
				}
				bgdef = false
				spr = isec["spr"]
			}

		case "stageinfo":
			if (isLan && lanStageinfo) || (!isLan && stageinfo) {
				if isLan {
					lanStageinfo = false
				}
				stageinfo = false

				isec.ReadI32("localcoord", &ss.localcoord[0], &ss.localcoord[1])
				isec.ReadF32("portraitscale", &ss.portraitscale)

				// Cache localcoords for stage fitting
				if _, ok := sys.stageLocalcoords[ss.def]; !ok {
					key := strings.ToLower(filepath.Base(ss.def))
					sys.stageLocalcoords[key] = ss.localcoord
				}
			}
		}
	}
	if spr != "" {
		ss.preloadSprite = spr
	}
	if sys.cfg.Config.BootLoadingMode > 0 {
		ss.sff = newSff()
		ss.anims.updateSff(ss.sff)
		for k := range s.stageSpritePreload {
			ss.anims.addSprite(ss.sff, k[0], k[1])
		}
		s.ensureStagePreloadSlot(len(s.stagelist))
	} else {
		if err := s.preloadStageAssets(len(s.stagelist)); err != nil {
			panic(fmt.Errorf("failed to preload %v: %v", def, err))
		}
	}
	return ss, nil
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
		pl = int(Rand(1, int32(sys.cfg.Config.PaletteMax)))
	}
	sys.selMutex.Lock()
	s.selected[tn] = append(s.selected[tn], [...]int{n, pl})
	if s.gameParams == nil {
		s.gameParams = newGameParams()
	}
	// ensure per-member override slot exists (needed for existed flag / persistence)
	dst := len(s.selected[tn]) - 1
	_ = s.gameParams.ensureOverride(tn, dst)
	sys.selMutex.Unlock()
	return true
}

func (s *Select) ClearSelected() {
	sys.selMutex.Lock()
	s.selected = [2][][2]int{}
	s.selectedStageNo = -1
	s.music = make(Music)
	s.gameParams = newGameParams()
	sys.selMutex.Unlock()
}

type LoaderState int32

const (
	LS_NotYet LoaderState = iota
	LS_Loading
	LS_Complete
	LS_Error
	LS_Cancel
)

// Returned by heavy loaders (notably SFF parsing) to signal cooperative cancellation.
var ErrLoadingCanceled = errors.New("loading canceled")

type Loader struct {
	state      LoaderState
	loadExit   chan LoaderState
	err        error
	cancelCh   chan struct{}
	cancelOnce sync.Once
}

func newLoader() *Loader {
	return &Loader{state: LS_NotYet, loadExit: make(chan LoaderState, 1)}
}

func (l *Loader) requestCancel() {
	if l.cancelCh == nil {
		return
	}
	l.cancelOnce.Do(func() { close(l.cancelCh) })
}

func (l *Loader) cancelRequested() bool {
	if l.cancelCh == nil {
		return false
	}
	select {
	case <-l.cancelCh:
		return true
	default:
		return false
	}
}

/*
func (l *Loader) loadPlayerChar(pn int) int {
	return l.loadCharacter(pn, false)
}

func (l *Loader) loadAttachedChar(pn int) int {
	return l.loadCharacter(pn, true)
}
*/

func (s *System) turnsPreloadActive() bool {
	return s.cfg.Config.TurnsLoading && (s.turnsPreloadMember[0] >= 0 || s.turnsPreloadMember[1] >= 0)
}

func (s *System) startNextTurnsPreload() {
	if !s.cfg.Config.TurnsLoading || s.loader.state == LS_Loading || s.loader.state == LS_Error {
		return
	}
	next := [2]int{-1, -1}
	s.selMutex.RLock()
	for team := 0; team < 2; team++ {
		if s.tmode[team] != TM_Turns {
			continue
		}
		member := int(s.wins[team^1]) + 1
		if member <= 0 || member >= int(s.numTurns[team]) || member >= len(s.sel.selected[team]) {
			continue
		}
		pn := team + 2 // one hidden standby slot per Turns side
		if len(s.chars[pn]) > 0 && s.chars[pn][0] != nil &&
			s.chars[pn][0].memberNo == member &&
			s.chars[pn][0].selectNo == s.sel.selected[team][member][0] {
			continue
		}
		next[team] = member
	}
	s.selMutex.RUnlock()
	if next == [2]int{-1, -1} {
		s.turnsPreloadMember = next
		return
	}
	s.turnsPreloadMember = next
	if s.loader.state == LS_Complete || s.loader.state == LS_Cancel {
		select {
		case <-s.loader.loadExit:
		default:
		}
		s.loader.state = LS_NotYet
		s.loader.err = nil
		s.loader.cancelCh = nil
		s.loader.cancelOnce = sync.Once{}
	}
	if s.loader.state == LS_NotYet {
		s.loader.runTread()
	}
}

// Rollback snapshots must not race the asynchronous Turns loader. The loader mutates
// character slots and compilation globals that are not safe to update between save/load.
func (s *System) finishTurnsPreloadForRollback() bool {
	if !s.turnsPreloadActive() {
		return true
	}
	for s.loader.state != LS_Complete {
		switch s.loader.state {
		case LS_Error:
			if s.loader.err != nil {
				panic(s.loader.err)
			}
			return false
		case LS_Cancel:
			return false
		case LS_NotYet:
			if !s.loader.runTread() {
				return false
			}
		}
		if !s.await(s.gameRenderSpeed()) {
			return false
		}
	}
	return true
}

// Rewrite slot-indexed fields inside chars when a preloaded Turns member is promoted into P1/P2.
func (s *System) remapCharSlotRefs(chars []*Char, oldSlot, newSlot int) {
	oldCPU := oldSlot ^ -1
	newCPU := newSlot ^ -1
	for _, ch := range chars {
		if ch == nil {
			continue
		}
		ch.playerNo = newSlot
		ch.ss.sb.playerNo = newSlot
		if ch.controller == oldSlot {
			ch.controller = newSlot
		} else if ch.controller == oldCPU {
			ch.controller = newCPU
		}
		if ch.animPN == oldSlot {
			ch.animPN = newSlot
		}
		if ch.spritePN == oldSlot {
			ch.spritePN = newSlot
		}
		// Commands are slot-indexed.
		if oldSlot >= 0 && newSlot >= 0 && oldSlot < len(ch.cmd) && newSlot < len(ch.cmd) {
			ch.cmd[oldSlot], ch.cmd[newSlot] = ch.cmd[newSlot], ch.cmd[oldSlot]
		}
	}
}

func (s *System) rebindCgiStateOwners(pn int) {
	if pn < 0 || pn >= len(s.cgi) || s.cgi[pn].states == nil {
		return
	}
	keys := make([]int32, 0, len(s.cgi[pn].states))
	for k := range s.cgi[pn].states {
		keys = append(keys, k)
	}
	for _, k := range keys {
		sb := s.cgi[pn].states[k]
		if sb.playerNo != pn {
			sb.playerNo = pn
			s.cgi[pn].states[k] = sb
		}
	}
}

func (s *System) removePlayerFromCharList(pn int) {
	for {
		removed := false
		for _, c := range s.charList.creationOrder {
			if c != nil && c.playerNo == pn {
				s.charList.delete(c)
				removed = true
				break
			}
		}
		if !removed {
			return
		}
	}
}

func (s *System) setBGTurnsSlotState(chars []*Char, slot int, active bool) {
	team := slot & 1
	for _, ch := range chars {
		if ch == nil {
			continue
		}
		if active {
			ch.teamside = team
			ch.unsetSCF(SCF_disabled)
			ch.unsetSCF(SCF_standby)
			if ch.helperIndex == 0 {
				ch.controller = slot
				if s.aiLevel[slot] != 0 {
					ch.controller ^= -1
				}
				// Defensive: the real round initialization will run next, but
				// never let a promoted preloaded fighter enter as already dead.
				if ch.life <= 0 {
					ch.life = Max(1, ch.lifeMax)
					ch.redLife = ch.life
				}
			}
		} else {
			ch.teamside = -1
			ch.setSCF(SCF_disabled)
			ch.setSCF(SCF_standby)
			ch.setCtrl(false)
			if ch.helperIndex == 0 {
				ch.controller = slot
			}
		}
	}
}

// Promote the next preloaded Turns member into the active P1/P2 slot after a KO.
func (s *System) activateNextTurnsFighters() {
	for team := 0; team < 2; team++ {
		if s.tmode[team] != TM_Turns || !s.effectiveLoss[team] {
			continue
		}
		nextMember := int(s.wins[team^1])
		if nextMember < 0 || nextMember >= int(s.numTurns[team]) {
			continue
		}
		dst := team
		src := team + 2
		if src < 0 || src >= MaxSimul*2 {
			continue
		}
		if len(s.chars[src]) == 0 || s.chars[src][0] == nil ||
			s.chars[src][0].memberNo != nextMember {
			continue
		}
		s.removePlayerFromCharList(dst)
		s.removePlayerFromCharList(src)
		oldDst, oldSrc := dst, src
		s.chars[dst], s.chars[src] = s.chars[src], s.chars[dst]
		s.cgi[dst], s.cgi[src] = s.cgi[src], s.cgi[dst]
		s.stringPool[dst], s.stringPool[src] = s.stringPool[src], s.stringPool[dst]
		s.remapCharSlotRefs(s.chars[dst], oldSrc, dst)
		s.remapCharSlotRefs(s.chars[src], oldDst, src)
		s.rebindCgiStateOwners(dst)
		s.rebindCgiStateOwners(src)
		s.workingChar = nil
		s.workingState = nil
		s.setBGTurnsSlotState(s.chars[dst], dst, true)
		s.setBGTurnsSlotState(s.chars[src], src, false)
		if s.chars[dst][0].id < 0 {
			s.chars[dst][0].id = s.newCharId()
		}
		for _, ch := range s.chars[dst] {
			if ch != nil {
				s.charList.add(ch)
			}
		}
		s.charList.enemyNearChanged = true
	}
}

func (l *Loader) loadCharacter(pn int, attached bool) int {
	// Quick abort between expensive steps.
	if l.cancelRequested() || l.state == LS_Cancel || sys.gameEnd {
		return 0
	}

	sys.selMutex.RLock()
	team := pn & 1
	tm := sys.tmode[team]
	nSim := sys.numSimul[team]
	nTurns := sys.numTurns[team]
	teamSel := make([][2]int, len(sys.sel.selected[team]))
	copy(teamSel, sys.sel.selected[team])
	sys.selMutex.RUnlock()
	if l.cancelRequested() || l.state == LS_Cancel || sys.gameEnd {
		return 0
	}
	if attached && sys.roundNo != 1 {
		return 1
	}

	// Get number of selected characters in team
	memberNo := pn >> 1
	nsel := len(teamSel)

	if !attached && tm == TM_Turns && sys.cfg.Config.TurnsLoading {
		if sys.turnsPreloadActive() {
			// During the match only the fixed hidden standby slot is eligible.
			// Active P1/P2 must not be touched by the background loader.
			if pn < 2 || pn != team+2 {
				return 1
			}
			memberNo = sys.turnsPreloadMember[team]
			if memberNo < 0 {
				return 1
			}
		} else if pn >= 2 {
			// Initial pre-match load: only the first Turns member is loaded.
			// Clear any stale resident char from previous matches/rounds.
			sys.chars[pn] = nil
			sys.cgi[pn].states = nil
			sys.cgi[pn].palettedata = nil
			sys.cgi[pn].hitPauseToggleFlagCount = 0
			return 1
		}
		if memberNo >= nsel {
			return 0
		}
		if sys.turnsPreloadActive() &&
			len(sys.chars[pn]) > 0 &&
			sys.chars[pn][0] != nil &&
			sys.chars[pn][0].memberNo == memberNo &&
			sys.chars[pn][0].selectNo == teamSel[memberNo][0] {
			return 1
		}
	} else if !attached && sys.roundsExisted[team] > 0 {
		return 1
	}

	// Check if player number is acceptable for selected team mode
	if !attached {
		// BG Turns uses a staged standby slot instead of loading the whole team.
		if tm == TM_Turns && sys.cfg.Config.TurnsLoading {
			if memberNo >= int(nTurns) || pn >= 4 {
				sys.cgi[pn].states = nil
				sys.chars[pn] = nil
				return 1
			}
		} else if tm == TM_Simul || tm == TM_Tag {
			if memberNo >= int(nSim) {
				sys.cgi[pn].states = nil
				sys.chars[pn] = nil
				return 1
			}
		} else if pn >= 2 {
			return 0
		}

		if tm == TM_Turns && !sys.cfg.Config.TurnsLoading && nsel < int(nTurns) {
			return 0
		}

		if tm == TM_Turns && !sys.cfg.Config.TurnsLoading {
			off := int32(0)
			if sys.sel.gameParams != nil {
				off = sys.sel.gameParams.TurnsOffset[pn&1]
			}
			if off < 0 {
				off = 0
			}
			// Clamp numTurns if Lua sent inconsistent values.
			// In Survival Turns we expect: totalSelected == off + numTurns.
			if int(off) > nsel {
				off = int32(nsel)
			}
			if nTurns < 0 {
				nTurns = 0
			}
			if int(off)+int(nTurns) > nsel {
				nTurns = int32(nsel) - off
			}
			memberNo = int(off) + int(sys.wins[^pn&1])
		}

		if nsel <= memberNo {
			return 0
		}
	}

	teamChars := make([]int, nsel)
	for i := range teamChars {
		teamChars[i] = teamSel[i][0]
	}

	// Prepare loading time clipboard message
	var tstr string
	tnow := time.Now()
	defer func() {
		sys.loadTime(tnow, tstr, false, true)
		// Mugen compatibility mode indicator
		if sys.cgi[pn].ikemenver[0] == 0 && sys.cgi[pn].ikemenver[1] == 0 {
			if sys.cgi[pn].mugenver[0] == 1 && sys.cgi[pn].mugenver[1] == 1 {
				sys.appendToConsole("Using Mugen 1.1 compatibility mode.")
			} else if sys.cgi[pn].mugenver[0] == 1 && sys.cgi[pn].mugenver[1] == 0 {
				sys.appendToConsole("Using Mugen 1.0 compatibility mode.")
			} else if sys.cgi[pn].mugenver[0] != 1 {
				sys.appendToConsole("Using WinMugen compatibility mode.")
			} else {
				sys.appendToConsole("Character with unknown engine version.")
			}
		}
	}()

	var cdef string
	var cdefOWnumber int

	if attached {
		atcpn := pn - MaxSimul*2
		cdef = sys.stageList[0].attachedchardef[atcpn]
	} else {
		if tm == TM_Turns {
			cdefOWnumber = memberNo*2 + pn&1
		} else {
			cdefOWnumber = pn
		}
		sys.selMutex.RLock()
		cdefOW := sys.sel.cdefOverwrite[cdefOWnumber]
		sys.selMutex.RUnlock()
		if cdefOW != "" {
			cdef = cdefOW
		} else {
			cdef = sys.sel.charlist[teamChars[memberNo]].def
		}
	}

	for _, ffx := range sys.ffx {
		prefixToDecrement := true
		for _, fxPath := range sys.cgi[pn].fxPath {
			if ffx.fileName == fxPath {
				prefixToDecrement = false
				break
			}
		}
		if prefixToDecrement && ffx.isCharFX {
			if ffx.refCount > 0 {
				ffx.refCount--
			}
		}
	}

	var p *Char
	sys.workingChar = p // This should help compiler and bytecode stay consistent

	// Same character from a previous match
	sameChar := sys.chars[pn] != nil && len(sys.chars[pn]) > 0 && sys.chars[pn][0].gi().def == cdef && sys.chars[pn][0].ocd().existed

	if sameChar {
		p = sys.chars[pn][0]

		// Restore values that ModifyPlayer may have mutated.
		// TODO: Should ModifyPlayer persist or not across matches?
		p.resetModifyPlayer()

		// Prepare success message
		if attached {
			tstr = fmt.Sprintf("Same attached char kept: %v", cdef)
		} else {
			tstr = fmt.Sprintf("Same char kept: %v", cdef)
		}
	}

	// Brand new character
	if p == nil {
		p = newChar(pn, 0)

		// These are already in char.load()
		//sys.cgi[pn].sff = nil
		//sys.cgi[pn].palettedata = nil

		// TODO: Why do we do this here?
		if len(sys.chars[pn]) > 0 {
			p.power = sys.chars[pn][0].power
			p.guardPoints = sys.chars[pn][0].guardPoints // And why do these two persist?
			p.dizzyPoints = sys.chars[pn][0].dizzyPoints
		}

		// Prepare success message
		if attached {
			tstr = fmt.Sprintf("New attached char loaded: %v", cdef)
		} else {
			tstr = fmt.Sprintf("New char loaded: %v", cdef)
		}
	}

	// Set new character parameters
	if attached {
		atcpn := pn - MaxSimul*2
		p.memberNo = atcpn
		p.selectNo = -atcpn
		p.teamside = -1
		sys.aiLevel[pn] = 0
		p.controller = pn
	} else {
		p.memberNo = memberNo
		p.selectNo = teamSel[memberNo][0]
		p.teamside = p.playerNo & 1
	}

	// Commit character to system
	sys.chars[pn] = make([]*Char, 1)
	sys.chars[pn][0] = p
	if !attached && tm == TM_Turns && sys.cfg.Config.TurnsLoading {
		sys.setBGTurnsSlotState(sys.chars[pn], pn, pn < 2 && !sys.turnsPreloadActive())
	}

	// Load character
	if !sameChar {
		if l.cancelRequested() || l.state == LS_Cancel || sys.gameEnd {
			return 0
		}
		if l.err = p.load(cdef); l.err != nil {
			if errors.Is(l.err, ErrLoadingCanceled) {
				sys.chars[pn] = nil
				l.state = LS_Cancel
				return 0
			}
			sys.chars[pn] = nil
			if attached {
				tstr = fmt.Sprintf("WARNING: Failed to load new attached char: %v", cdef)
			} else {
				tstr = fmt.Sprintf("WARNING: Failed to load new char: %v", cdef)
			}
			return -1
		}
		if l.cancelRequested() || l.state == LS_Cancel || sys.gameEnd {
			sys.chars[pn] = nil
			l.state = LS_Cancel
			return 0
		}

		// Compile character states
		if sys.cgi[pn].states, l.err = newCharCompiler().Compile(p.playerNo, cdef, p.gi().constants); l.err != nil {
			sys.chars[pn] = nil
			if attached {
				tstr = fmt.Sprintf("WARNING: Failed to compile new attached char states: %v", cdef)
			} else {
				tstr = fmt.Sprintf("WARNING: Failed to compile new char states: %v", cdef)
			}
			return -1
		}
	}

	// Setup selected palette
	selectPalno := 1
	if pal, ok := sys.sel.palOverwrite[pn]; ok && pal > 0 {
		selectPalno = pal
	} else if !attached {
		// Get palette number from select screen choice
		selectPalno = teamSel[memberNo][1]
	}
	sys.cgi[pn].palno = int32(selectPalno)

	// Apply per-launch map overrides prepared from Lua/loadStart.
	if !attached {
		p.applyMapOverrides()
	}

	// Flag "existed" just in case
	// Update: This is Lua's responsibility. Enabling this flag made duplicate Turns characters misbehave
	//sys.chars[pn][0].ocd().existed = true

	return 1
}

func (l *Loader) prepareTurnsFaces(pn int, fa *FightScreenFace, nm *FightScreenName, teamChars []int, teamSel [][2]int) {
	// Reset face and name KO's
	off := int32(0)
	if sys.sel.gameParams != nil {
		off = sys.sel.gameParams.TurnsOffset[pn&1]
	}
	if off < 0 {
		off = 0
	}
	// Clamp to valid range so fight screen teammate rotation doesn't misbehave
	if len(teamChars) == 0 {
		fa.numko = 0
		nm.numko = 0
	} else {
		maxOff := int32(len(teamChars) - 1)
		if off > maxOff {
			off = maxOff
		}
		fa.numko = off
		nm.numko = off
	}

	// Pre-allocate
	nsel := len(teamChars)
	nm.teammate_name_strings = make([]string, nsel)
	fa.teammate_face = make([]*Sprite, nsel)
	fa.teammate_scale = make([]float32, nsel)
	fa.teammate_face_pfx = make([]*PalFX, nsel)

	// Wait for texture uploads before we clone the portraits. Fixes quick VS
	sys.runMainThreadTask()

	// Iterate all selected characters
	for i, charIdx := range teamChars {
		sc := &sys.sel.charlist[charIdx]

		// Save the name
		nm.teammate_name_strings[i] = sc.lifebarname

		// Calculate portrait scale
		fa.teammate_scale[i] = sc.portraitscale * 320 / float32(sc.localcoord[0])

		// Get the sprite from the teammate's SFF
		origSpr := sc.sff.GetSprite(uint16(fa.teammate_face_spr[0]), uint16(fa.teammate_face_spr[1]))
		if origSpr == nil {
			continue
		}

		// Create an independent clone of the sprite to avoid mutating the SFF
		spr := *origSpr

		// Run palette replacement if applicable
		if fa.teammate_face_palshare {
			// Check if palettes are already loaded
			// They won't be unless the select screen used "applypal" (or if the character was already used before maybe)
			// https://github.com/ikemen-engine/Ikemen-GO/issues/3300
			palIdx := teamSel[i][1]
			_, hasTarget := sc.sff.palList.PalTable[[...]uint16{1, uint16(palIdx)}]
			_, has11 := sc.sff.palList.PalTable[[...]uint16{1, 1}]

			// Only load palettes if necessary
			if !hasTarget || !has11 {
				sc.sff.loadActPalettes(charIdx)
			}

			// Check if the sprite uses or shares palette 1, 1
			usesPal11 := false
			if spr.coldepth <= 8 {
				pal11Idx, ok := sc.sff.palList.PalTable[[...]uint16{1, 1}]
				if ok && spr.palidx >= 0 && int(spr.palidx) < len(sc.sff.palList.paletteMap) && pal11Idx < len(sc.sff.palList.paletteMap) {
					if sc.sff.palList.paletteMap[spr.palidx] == sc.sff.palList.paletteMap[pal11Idx] {
						usesPal11 = true
					}
				}
			}

			// Apply selected color only if the sprite shares the base palette
			if usesPal11 {
				// Pull selected palette index (1-based)
				targetPal := sc.sff.palList.Get(int(palIdx) - 1)
				if targetPal != nil {
					// Decouple clone from global SFF palettes
					spr.Pal = make([]uint32, len(targetPal))
					copy(spr.Pal, targetPal)

					spr.paltemp = make([]uint32, len(targetPal))
					copy(spr.paltemp, targetPal)

					// Force lazy loading for unique recolored texture
					spr.PalTex = nil
					spr.palidx = -1
				}
			}
		}

		// Commit sprite and init PalFX
		fa.teammate_face[i] = &spr
		fa.teammate_face_pfx[i] = newPalFX()
	}
}

func (l *Loader) loadStage() bool {
	if sys.roundNo == 1 {
		if l.cancelRequested() || l.state == LS_Cancel || sys.gameEnd {
			l.state = LS_Cancel
			return false
		}
		var tstr string
		tnow := time.Now()
		defer func() {
			if sys.stage != nil {
				sys.loadTime(tnow, tstr, false, true)
				// Mugen compatibility mode indicator
				if sys.stage.ikemenver[0] == 0 && sys.stage.ikemenver[1] == 0 {
					if sys.stage.mugenver[0] == 1 && sys.stage.mugenver[1] == 1 {
						sys.appendToConsole("Using Mugen 1.1 compatibility mode.")
					} else if sys.stage.mugenver[0] == 1 && sys.stage.mugenver[1] == 0 {
						sys.appendToConsole("Using Mugen 1.0 compatibility mode.")
					} else if sys.stage.mugenver[0] != 1 {
						sys.appendToConsole("Using WinMugen compatibility mode.")
					} else {
						sys.appendToConsole("Stage with unknown engine version.")
					}
				}
			}
		}()
		// Snapshot stage selection quickly (avoid races with Lua).
		sys.selMutex.RLock()
		stageNo := sys.sel.selectedStageNo
		sdefOW := sys.sel.sdefOverwrite
		sys.selMutex.RUnlock()

		var def string
		if stageNo == 0 {
			randomstageno := Rand(0, int32(len(sys.sel.stagelist))-1)
			def = sys.sel.stagelist[randomstageno].def
		} else {
			def = sys.sel.stagelist[stageNo-1].def
		}
		if sdefOW != "" {
			def = sdefOW
		}
		if sys.stage != nil && sys.stage.def == def && sys.stage.mainstage && !sys.stage.reload {
			tstr = fmt.Sprintf("Cached stage loaded: %v", def)
			return true
		}
		// We're switching stages (or reloading): tear down background media in the old stage.
		if sys.stage != nil && (sys.stage.def != def || !sys.stage.mainstage || sys.stage.reload) {
			sys.stage.destroy()
		}
		sys.stageList = make(map[int32]*Stage)
		sys.stageLoop = false
		sys.stageList[0], l.err = loadStage(def, true)

		// Add the stage's name to the error stack
		if l.err != nil {
			l.err = fmt.Errorf("\nError loading %v: %v", def, l.err)
			return false
		}

		sys.stage = sys.stageList[0]
		tstr = fmt.Sprintf("New stage loaded: %v", def)
	}

	return l.err == nil
}

func (l *Loader) load() {
	defer func() {
		l.loadExit <- l.state
	}()

	stagedTurns := sys.turnsPreloadActive()

	// Load any config-driven Common FX not already cached. External modules may
	// append to Common.Fx before a match; once loaded, non-char FX are kept.
	if !stagedTurns {
		if err := loadCommonFightFx(sys.fightScreen.def, false); err != nil {
			l.err = err
			l.state = LS_Error
			return
		}
	}

	/*
		// This should now be handled by loadSff()
		sys.loadMutex.Lock()
		for prefix, ffx := range sys.ffx {
			if !ffx.isCharFX {
				continue
			}
			if ffx.refCount <= 0 {
				if ffx.sff != nil {
					removeSFFCache(ffx.sff.filename)
				}
				delete(sys.ffx, prefix)
			}
		}
		sys.loadMutex.Unlock()
	*/

	playerSlotsEnd := len(sys.chars) - MaxAttachedChar
	charDone, stageDone := make([]bool, len(sys.chars)), stagedTurns

	if stagedTurns {
		// In-match Turns preload must not touch active gameplay slots or the other side's normal slots. It only fills one standby slot per Turns side: P3 for side 1, P4 for side 2.
		for i := range charDone {
			charDone[i] = true
		}
		for team, member := range sys.turnsPreloadMember {
			if member >= 0 {
				pn := team + 2
				if pn >= 0 && pn < playerSlotsEnd {
					charDone[pn] = false
				}
			}
		}
	}

	// Check if all chars are loaded
	allCharDone := func() bool {
		for _, b := range charDone {
			if !b {
				return false
			}
		}
		return true
	}

	for !stageDone || !allCharDone() {
		// Fast exit on cancel without waiting for the whole load to complete.
		if l.cancelRequested() || l.state == LS_Cancel || sys.gameEnd {
			l.state = LS_Cancel
			return
		}
		// Load stage
		sys.selMutex.RLock()
		stageNo := sys.sel.selectedStageNo
		sys.selMutex.RUnlock()
		if !stageDone && stageNo >= 0 {
			if !l.loadStage() {
				if l.state == LS_Cancel || l.cancelRequested() {
					l.state = LS_Cancel
					return
				}
				l.state = LS_Error
				return
			}
			stageDone = true
		}
		// Load characters that aren't already loaded
		for i, b := range charDone {
			if !b {
				if l.cancelRequested() || l.state == LS_Cancel || sys.gameEnd {
					l.state = LS_Cancel
					return
				}
				if stagedTurns && i >= playerSlotsEnd {
					charDone[i] = true
					continue
				}
				result := -1
				// Attached stage chars depend on the loaded stage definition.
				if i >= playerSlotsEnd {
					if !stageDone || sys.stageList[0] == nil {
						continue
					}
					atcpn := i - MaxSimul*2
					if atcpn < 0 || atcpn >= len(sys.stageList[0].attachedchardef) || sys.stageList[0].attachedchardef[atcpn] == "" {
						sys.chars[i] = nil
						sys.cgi[i].states = nil
						sys.cgi[i].hitPauseToggleFlagCount = 0
						charDone[i] = true
						continue
					}
					result = l.loadCharacter(i, true)
				} else {
					result = l.loadCharacter(i, false)
				}
				if result > 0 {
					charDone[i] = true
				} else if result < 0 {
					l.state = LS_Error
					return
				}
			}
		}
		for i := 0; i < 2; i++ {
			// Snapshot selection + team mode under selMutex.
			sys.selMutex.RLock()
			selLen := len(sys.sel.selected[i])
			tm := sys.tmode[i]
			sys.selMutex.RUnlock()
			if !charDone[i+2] && selLen > 0 &&
				tm != TM_Simul && tm != TM_Tag &&
				!(tm == TM_Turns && sys.cfg.Config.TurnsLoading) {
				for j := i + 2; j < playerSlotsEnd; j += 2 {
					if !charDone[j] {
						sys.chars[j] = nil
						sys.cgi[j].states = nil
						sys.cgi[j].hitPauseToggleFlagCount = 0
						charDone[j] = true
					}
				}
			}
		}
		if l.cancelRequested() || l.state == LS_Cancel || sys.gameEnd {
			l.state = LS_Cancel
			return
		}
		time.Sleep(10 * time.Millisecond)
		if sys.gameEnd {
			l.state = LS_Cancel
		}
		if l.state == LS_Cancel {
			return
		}
	}
	sys.cleanCustomShaders()

	// Flag loading state as complete
	l.state = LS_Complete
}

func (l *Loader) reset() {
	if l.state != LS_NotYet {
		// Ensure the loader goroutine gets a cooperative cancel signal.
		l.requestCancel()
		l.state = LS_Cancel
		<-l.loadExit
		l.state = LS_NotYet
	}
	l.err = nil
	l.cancelCh = nil
	l.cancelOnce = sync.Once{}
	for i := range sys.cgi {
		keepPreloadedTurnsPal := sys.cfg.Config.TurnsLoading && sys.roundNo > 1 && sys.tmode[i&1] == TM_Turns
		if sys.roundsExisted[i&1] == 0 && !keepPreloadedTurnsPal {
			sys.cgi[i].palno = -1
		}
	}
}

func (l *Loader) runTread() bool {
	if l.state != LS_NotYet {
		return false
	}
	// Fresh cancel signal for this run.
	l.cancelCh = make(chan struct{})
	l.cancelOnce = sync.Once{}
	l.state = LS_Loading
	SafeGo(func() {
		l.load()
	})
	return true
}

type EnvShake struct {
	time      int32
	freq      float32
	ampl      float32
	phase     float32
	mul       float32
	dir       float32
	diradd    float32
	decay     float32
	curTime   int32
	curOffset [2]float32
}

func (es *EnvShake) clear() {
	*es = EnvShake{
		freq:  60,
		phase: float32(math.NaN()),
		ampl:  -4,
		mul:   1.0,
	}
}

func (es *EnvShake) restart() {
	if es.time <= 0 {
		es.clear() // Just in case
		return
	}
	es.curTime = es.time
}

func (es *EnvShake) setDefaultPhase() {
	if math.IsNaN(float64(es.phase)) {
		if es.freq >= 90.0 {
			es.phase = 90.0
		} else {
			es.phase = 0
		}
	}
}

func (es *EnvShake) update() {
	if es.curTime <= 0 {
		// Only run clear() if necessary
		if es.time > 0 {
			es.clear()
		}
		return
	}

	// Recompute offset once per frame and cache it
	elapsed := float32(es.time - es.curTime)
	currentPhase := es.phase + es.freq*elapsed
	currentDir := es.dir + es.diradd*elapsed

	var factor float64 = 1.0

	// Mul is applied once per cycle
	if es.mul != 0 && es.mul != 1 {
		cycles := math.Floor(float64(es.freq*elapsed) / 360.0)
		factor *= math.Pow(float64(es.mul), cycles)
	}

	// Apply decay
	t := float64(es.curTime) / float64(es.time)
	if es.decay != 0 && t > 0 {
		factor *= math.Pow(t, float64(es.decay))
	}

	ampl := float64(es.ampl) * factor
	if ampl == 0 {
		es.curOffset = [2]float32{0, 0}
	} else {
		phaseRad := float64(currentPhase) * math.Pi / 180.0
		dirRad := float64(currentDir) * math.Pi / 180.0
		offset := ampl * math.Sin(phaseRad)
		x := float32(offset * math.Sin(-dirRad))
		y := float32(offset * math.Cos(-dirRad))
		es.curOffset = [2]float32{x, y}
	}

	// Step timer only after computing the offset
	es.curTime--
}

func (es *EnvShake) getOffset() [2]float32 {
	return es.curOffset
}

type ZoomEffect struct {
	active      bool
	time        int32
	lag         float32
	endLag      float32
	scale       float32
	curScale    float32
	pos         [2]float32 // Defined parameters
	curPos      [2]float32 // Current values with lag
	cameraBound bool
	stageBound  bool
}

func (z *ZoomEffect) reset() {
	z.active = false
	z.time = 0
	z.pos = [2]float32{0, 0}
	z.curPos = [2]float32{0, 0}
	z.scale = 1
	z.curScale = 1
	z.lag = 0
	z.endLag = 0
	z.cameraBound = true
	z.stageBound = true
}

// Step the zoom variables
func (z *ZoomEffect) update() {
	if !z.active {
		return
	}

	// Active phase target
	targetPos := z.pos
	targetScale := z.scale
	lag := z.lag

	// Release phase target
	if z.time <= 0 {
		targetPos = [2]float32{0, 0}
		targetScale = 1
		lag = z.endLag
	}

	// Manage timer
	if z.time > 0 {
		z.time--
	}

	// Threshold for snapping to target
	const eps = 0.001

	// Apply smoothing
	if lag == 0 {
		// Fast path for no lag
		z.curPos = targetPos
		z.curScale = targetScale
	} else if lag < 1 {
		// Position
		for i := 0; i < 2; i++ {
			if Abs(z.curPos[i]-targetPos[i]) < eps {
				z.curPos[i] = targetPos[i]
			} else {
				z.curPos[i] += (targetPos[i] - z.curPos[i]) * (1 - lag)
			}
		}

		// Scale
		if Abs(z.curScale-targetScale) < eps {
			z.curScale = targetScale
		} else {
			z.curScale += (targetScale - z.curScale) * (1 - lag)
		}
	}
	// lag >= 1 freezes current zoom state. Maybe that shouldn't be allowed either?

	// Finish release
	if z.time <= 0 {
		if Abs(z.curPos[0]) < eps &&
			Abs(z.curPos[1]) < eps &&
			Abs(z.curScale-1) < eps {
			z.reset()
		}
	}
}

// Apply current zoom to sys.draw() parameters
func (z *ZoomEffect) apply(x, y, scl float32) (dx, dy, dscl float32) {
	dx, dy, dscl = x, y, scl

	if !z.active {
		return
	}

	finalScale := z.curScale * scl

	// Apply position limits
	if z.stageBound {
		dscl = Max(sys.cam.MinScale, finalScale/sys.cam.BaseScale())

		if z.cameraBound {
			zoomedViewWidth := float32(sys.gameWidth) / finalScale
			minCamX := x - (sys.cam.halfWidth/scl - zoomedViewWidth/2)
			maxCamX := x + (sys.cam.halfWidth/scl - zoomedViewWidth/2)
			intermediateTargetX := x + z.curPos[0]/scl
			dx = Clamp(intermediateTargetX, minCamX, maxCamX)
		} else {
			dx = x + z.curPos[0]/scl
		}

		dx = sys.cam.XBound(dscl, dx)
	} else {
		dscl = finalScale / sys.cam.BaseScale()
		dx = x + z.curPos[0]/scl
	}

	dy = y + z.curPos[1]/scl

	return
}

type CharVarBackup struct {
	cnsvar   map[int32]int32
	cnsfvar  map[int32]float32
	mapArray map[string]float32
}

func (s *System) saveCharVars(pn int) {
	if len(s.chars[pn]) == 0 {
		return
	}
	c := s.chars[pn][0]

	bk := CharVarBackup{
		cnsvar:   make(map[int32]int32),
		cnsfvar:  make(map[int32]float32),
		mapArray: make(map[string]float32),
	}

	for k, v := range c.cnsvar {
		bk.cnsvar[k] = v
	}
	for k, v := range c.cnsfvar {
		bk.cnsfvar[k] = v
	}
	for k, v := range c.mapArray {
		bk.mapArray[k] = v
	}

	s.charVarsBackup[pn] = bk
}

func (s *System) restoreCharVars(c *Char) {
	bk, ok := s.charVarsBackup[c.playerNo]
	if !ok {
		return
	}

	for k, v := range bk.cnsvar {
		c.cnsvar[k] = v
	}
	for k, v := range bk.cnsfvar {
		c.cnsfvar[k] = v
	}
	c.mapArray = make(map[string]float32, len(bk.mapArray))
	for k, v := range bk.mapArray {
		c.mapArray[k] = v
	}

	delete(s.charVarsBackup, c.playerNo)
}

func (s *System) isValidCustomShader(name string) bool {
	if _, ok := sys.shaderRefCount[name]; ok {
		return true
	}
	return false
}

func (s *System) cleanCustomShaders() {
	activeShaders := make(map[string]bool)
	for i := 0; i < len(s.cgi); i++ {
		for _, sName := range s.cgi[i].customShaders {
			activeShaders[sName] = true
		}
	}

	for sName, count := range s.shaderRefCount {
		if activeShaders[sName] {
			s.shaderRefCount[sName] = 3
		} else {
			count--
			if count <= 0 {
				s.mainThreadTask <- func(name string) func() {
					return func() {
						gfx.UnloadCustomSpriteShader(name)
					}
				}(sName)

				delete(s.shaderRefCount, sName)
			} else {
				s.shaderRefCount[sName] = count
			}
		}
	}
}

func (s *System) shouldHideWithBars() bool {
	return !s.fightScreen.visible() || s.gsf(GSF_nobardisplay) || !s.fightScreen.bars
}

package main

import (
	"arena"
	"bufio"
	"fmt"
	"image"
	"io"
	"log"
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

// sys
// The only instance of a System struct.
// Do not create more than 1.
var sys = System{
	randseed:      int32(time.Now().UnixNano()),
	scrrect:       [...]int32{0, 0, 320, 240},
	gameWidth:     320,
	gameHeight:    240,
	widthScale:    1,
	heightScale:   1,
	brightness:    1,
	maxRoundTime:  -1,
	soundMixer:    &beep.Mixer{},
	bgm:           *newBgm(),
	soundChannels: newSoundChannels(16),
	allPalFX:      newPalFX(),
	bgPalFX:       newPalFX(),
	ffx:           make(map[string]*FightFx),
	//ffxRegexp:         "^(f)|^(s)|^(go)", // https://github.com/ikemen-engine/Ikemen-GO/issues/1620
	sel:      *newSelect(),
	keyState: make(map[Key]bool),
	match:    1,
	loader:   *newLoader(),
	numSimul: [...]int32{2, 2}, numTurns: [...]int32{2, 2},
	ignoreMostErrors:    true,
	stageList:           make(map[int32]*Stage),
	stageLocalcoords:    make(map[string][2]float32),
	oldNextAddTime:      1,
	commandLine:         make(chan string),
	cam:                 *newCamera(),
	mainThreadTask:      make(chan func(), 65536),
	workpal:             make([]uint32, 256),
	errLog:              log.New(NewLogWriter(), "", log.LstdFlags),
	keyInput:            KeyUnknown,
	saveState:           NewGameState(),
	statePool:           NewGameStatePool(),
	savePool:            NewGameStatePool(),
	loadPool:            NewGameStatePool(),
	luaStringVars:       make(map[string]string),
	luaNumVars:          make(map[string]float32),
	luaTables:           make([]*lua.LTable, 0),
	commandLists:        make([]*CommandList, 0),
	arenaSaveMap:        make(map[int]*arena.Arena),
	arenaLoadMap:        make(map[int]*arena.Arena),
	debugAccel:          1, // TODO: We probably shouldn't rely on this being initialized to 1
	lastInputController: -1,
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
	randseed                int32
	scrrect                 [4]int32
	gameWidth, gameHeight   int32
	widthScale, heightScale float32
	window                  *Window
	gameEnd, frameSkip      bool
	redrawWait              struct{ nextTime, lastDraw time.Time }
	brightness              float32
	brightnessOld           float32
	maxRoundTime            int32
	debugFont               *TextSprite
	debugDisplay            bool
	debugRef                [2]int // player number, helper index
	debugLastID             int32
	soundMixer              *beep.Mixer
	bgm                     Bgm
	soundChannels           *SoundChannels
	allPalFX                *PalFX
	bgPalFX                 *PalFX
	lifebar                 Lifebar
	motif                   Motif
	storyboard              Storyboard
	cfg                     Config
	ffx                     map[string]*FightFx
	sel                     Select
	keyState                map[Key]bool
	netConnection           *NetConnection
	replayFile              *ReplayFile
	aiInput                 [MaxPlayerNo]AiInput
	ffbparams               [MaxPlayerNo]ForceFeedbackParams
	keyConfig               []KeyConfig
	joystickConfig          []KeyConfig
	aiLevel                 [MaxPlayerNo]float32
	home                    int
	matchTime               int32
	match                   int32
	inputRemap              [MaxPlayerNo]int // Ingame player number (index) to human player number (value) pairing
	commandInputSource      [MaxPlayerNo]int // command list index to controller index to read inputs from
	round                   int32
	intro                   int32
	curRoundTime            int32
	lastHitter              [2]int
	winTeam                 int
	winType                 [2]WinType
	winTrigger              [2]WinType
	matchWins, wins         [2]int32
	roundsExisted           [2]int32
	draws                   int32
	effectiveLoss           [2]bool
	loader                  Loader
	chars                   [MaxPlayerNo][]*Char
	charList                CharList
	cgi                     [MaxPlayerNo]CharGlobalInfo
	tmode                   [2]TeamMode
	numSimul, numTurns      [2]int32
	esc                     bool
	loadMutex               sync.Mutex
	ignoreMostErrors        bool
	stringPool              [MaxPlayerNo]StringPool
	bcStack, bcVarStack     BytecodeStack
	bcVar                   []BytecodeValue
	workingChar             *Char
	workingState            *StateBytecode
	specialFlag             GlobalSpecialFlag
	envShake                EnvShake
	pausetime               int32
	pausetimebuffer         int32
	pausebg                 bool
	pauseendcmdbuftime      int32
	pauseplayerno           int
	supertime               int32
	supertimebuffer         int32
	superpausebg            bool
	superendcmdbuftime      int32
	superplayerno           int
	superbrightness         float32
	superp2defmul           float32
	envcol                  [3]int32
	envcol_time             int32
	envcol_under            bool
	stage                   *Stage
	stageList               map[int32]*Stage
	stageLoop               bool
	stageLoopNo             int
	stageLocalcoords        map[string][2]float32
	wireframeDisplay        bool
	lastCharId              int32
	tickCount               int
	oldTickCount            int
	tickCountF              float32
	lastTick                float32
	nextAddTime             float32
	oldNextAddTime          float32
	screenleft              float32
	screenright             float32
	xmin, xmax              float32
	zmin, zmax              float32
	winskipped              bool
	paused, frameStepFlag   bool
	roundResetFlg           bool
	roundResetMatchStart    bool
	reloadFlg               bool
	reloadStageFlg          bool
	reloadLifebarFlg        bool
	reloadCharSlot          [MaxPlayerNo]bool
	shortcutScripts         map[ShortcutKey]*ShortcutScript
	turbo                   float32
	commandLine             chan string
	drawScale               float32
	zoomlag                 float32
	zoomScale               float32
	zoomPosXLag             float32
	zoomPosYLag             float32
	enableZoomtime          int32
	zoomCameraBound         bool
	zoomStageBound          bool
	zoomPos                 [2]float32
	debugWC                 *Char
	cam                     Camera
	finishType              FinishType
	winwaittime             int32
	slowtime                int32
	winposetime             int32
	projs                   [MaxPlayerNo][]*Projectile
	explods                 [MaxPlayerNo][]*Explod
	chartexts               [MaxPlayerNo][]*TextSprite // From Text sctrl
	changeStateNest         int32
	spritesLayerN1          DrawList
	spritesLayerU           DrawList
	spritesLayer0           DrawList
	spritesLayer1           DrawList
	shadows                 ShadowList
	reflections             ReflectionList
	afterImageCount         [MaxPlayerNo]int32
	debugc1hit              ClsnRect
	debugc1rev              ClsnRect
	debugc1not              ClsnRect
	debugc2                 ClsnRect
	debugc2hb               ClsnRect
	debugc2mtk              ClsnRect
	debugc2grd              ClsnRect
	debugc2stb              ClsnRect
	debugcsize              ClsnRect
	debugch                 ClsnRect
	debugAccel              float32
	clsnSpr                 Sprite
	clsnDisplay             bool
	lifebarHide             bool
	mainThreadTask          chan func()
	workpal                 []uint32
	errLog                  *log.Logger
	nomusic                 bool
	workBe                  []BytecodeExp
	keyInput                Key
	keyString               string
	timerCount              []int32
	cmdFlags                map[string]string
	whitePalTex             Texture
	usePalette              bool
	credits                 int32
	gameRunning             bool

	msaa            int32
	externalShaders [][][]byte
	windowMainIcon  []image.Image
	gameMode        string
	frameCounter    int32
	preMatchTime    int32
	captureNum      int
	decisiveRound   [2]bool
	timerStart      int32
	timerRounds     []int32
	curPlayTime     int32
	scoreStart      [2]float32
	scoreRounds     [][2]float32
	statsLog        StatsLog
	consecutiveWins [2]int32
	firstAttack     [3]int
	teamLeader      [2]int
	maxPowerMode    bool
	clsnText        []ClsnText
	consoleText     []string
	luaLState       *lua.LState
	statusLFunc     *lua.LFunction
	listLFunc       []*lua.LFunction
	introSkipCall   bool
	endMatch        bool
	continueFlg     bool
	dialogueForce   int
	dialogueBarsFlg bool
	noSoundFlg      bool
	postMatchFlg    bool
	playBgmFlg      bool
	loopBreak       bool
	loopContinue    bool

	statePool       GameStatePool
	luaStringVars   map[string]string
	luaNumVars      map[string]float32
	luaTables       []*lua.LTable
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
	fightLoopEnd bool
	roundBackup  RoundStartBackup

	// for avg. FPS calculations
	gameFPS       float32
	prevTimestamp uint64
	absTickCountF float32

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

	lastInputController int
	uiLastInputToken    string

	// For Android support
	baseDir string
}

// Check if the application is running inside a macOS app bundle
func isRunningInsideAppBundle(exePath string) bool {
	// Check if we're on Darwin and the executable path contains .app (macOS application bundle)
	return runtime.GOOS == "darwin" && strings.Contains(exePath, ".app")
}

// Initialize stuff, this is called after the config int at main.go
func (s *System) init(w, h int32) *lua.LState {
	s.setGameSize(w, h)
	for i := range sys.cgi {
		sys.cgi[i].palInfo = make(map[int]PalInfo)
	}
	var err error
	// Create a system window.
	s.window, err = s.newWindow(int(s.scrrect[2]), int(s.scrrect[3]))
	chk(err)

	// Check if it's actually fullscreen. For some reason the constructor doesn't handle this properly.
	_, forceWindowed := s.cmdFlags["-windowed"]
	s.window.fullscreen = s.cfg.Video.Fullscreen && !forceWindowed

	if strings.Contains(s.cfg.Video.RenderMode, "OpenGL") {
		if ctx, err := s.window.GLCreateContext(); err != nil {
			Logcat("GL Context Creation Failed: " + err.Error())
			s.errLog.Fatalf("Could not initialize context :( Reason? %s", err)
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

			if s.cfg.Video.RenderMode != "Vulkan 1.3" {
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
			} else {
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

	Logcat("Check A: Selecting Renderer")
	// Now we need to init the renderer
	gfx, gfxFont = selectRenderer(s.cfg.Video.RenderMode)
	Logcat("Check B: Initializing GFX")
	gfx.Init()
	s.window.SetSwapInterval(s.cfg.Video.VSync) // VSync must be set after gfx, or our config will be ignored
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
		go func() {
			stdin := bufio.NewScanner(os.Stdin)
			for stdin.Scan() {
				if err := stdin.Err(); err != nil {
					s.errLog.Println(err.Error())
					return
				}
				s.commandLine <- stdin.Text()
			}
		}()
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

func (s *System) aspect(w, h int32) float32 {
	return float32(w) / float32(h)
}

func (s *System) middleOfMatch() bool {
	return !s.fightLoopEnd && s.matchTime != 0 && !s.postMatchFlg
}

func (s *System) skipMotifScaling() bool {
	var local [2]int32
	if (!s.middleOfMatch() && !s.postMatchFlg) || s.stage == nil {
		local = s.motif.Info.Localcoord
	} else {
		local = s.stage.stageCamera.localcoord
	}
	return s.aspect(local[0], local[1]) > s.getMotifAspect()
}

func (s *System) getFightAspect() float32 {
	// Stage aspect ratio
	if s.cfg.Video.FightAspectWidth < 0 && s.cfg.Video.FightAspectHeight < 0 && s.stage != nil {
		coord := s.stage.stageCamera.localcoord
		if coord[0] > 0 && coord[1] > 0 {
			return s.aspect(coord[0], coord[1])
		}
	}

	// Custom aspect ratio
	if s.cfg.Video.FightAspectWidth > 0 && s.cfg.Video.FightAspectHeight > 0 {
		return s.aspect(s.cfg.Video.FightAspectWidth, s.cfg.Video.FightAspectHeight)
	}

	// Default
	// Using video options directly has unwanted behavior if those options are changed without restarting Ikemen
	//return float32(s.cfg.Video.GameWidth) / float32(s.cfg.Video.GameHeight)
	return s.getMotifAspect()
}

func (s *System) getMotifAspect() float32 {
	// Using options directly makes aspect change as soon as options are changed
	//return float32(s.cfg.Video.GameWidth) / float32(s.cfg.Video.GameHeight)
	return s.aspect(s.scrrect[2], s.scrrect[3])
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

	screenAspect := s.aspect(w, h)
	targetAspect := s.aspect(baseWidth, baseHeight)

	if screenAspect > targetAspect {
		// Screen is wider than 4:3 - scale based on height
		s.gameWidth = int32(float32(baseHeight) * screenAspect)
		s.gameHeight = baseHeight
	} else {
		// Screen is taller than 4:3 - scale based on width
		s.gameWidth = baseWidth
		s.gameHeight = int32(float32(baseWidth) / screenAspect)
	}

	// Update scale
	s.widthScale = float32(s.scrrect[2]) / float32(s.gameWidth)
	s.heightScale = float32(s.scrrect[3]) / float32(s.gameHeight)
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
			aspectGame = s.aspect(stageWidth, stageHeight)
		} else {
			// Fallback
			//aspectGame = s.aspect(s.cfg.Video.GameWidth, s.cfg.Video.GameHeight)
			aspectGame = s.getMotifAspect()
		}
	} else {
		aspectGame = s.getFightAspect()
	}

	// Compute new game dimensions while maintaining the same base height
	gameWidth := baseHeight * aspectGame
	s.gameWidth = int32(gameWidth)
	s.gameHeight = int32(baseHeight)

	// Scale to fit current screen size
	s.widthScale = float32(s.scrrect[2]) / float32(s.gameWidth)
	s.heightScale = float32(s.scrrect[3]) / float32(s.gameHeight)
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
			s.errLog.Printf("[keepAlive] #%d Δ=%.3fms total=%.3fs at %s",
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
		waitDuration = s.rollback.session.loopTimer.usToWaitThisLoop()
	} else {
		waitDuration = time.Second / time.Duration(fps)
	}

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
		dx, dy, dscl := x, y, scl
		if s.enableZoomtime > 0 {
			if !s.debugPaused() {
				s.zoomPosXLag += ((s.zoomPos[0] - s.zoomPosXLag) * (1 - s.zoomlag))
				s.zoomPosYLag += ((s.zoomPos[1] - s.zoomPosYLag) * (1 - s.zoomlag))
				s.drawScale = s.drawScale / (s.drawScale + (s.zoomScale*scl-s.drawScale)*s.zoomlag) * s.zoomScale * scl
			}
			if s.zoomStageBound {
				dscl = MaxF(s.cam.MinScale, s.drawScale/s.cam.BaseScale())
				if s.zoomCameraBound {
					zoomedViewWidth := float32(s.gameWidth) / s.drawScale
					minCamX := x - (s.cam.halfWidth/scl - zoomedViewWidth/2)
					maxCamX := x + (s.cam.halfWidth/scl - zoomedViewWidth/2)
					intermediateTargetX := x + s.zoomPosXLag/scl
					dx = ClampF(intermediateTargetX, minCamX, maxCamX)
				} else {
					dx = x + s.zoomPosXLag/scl
				}
				dx = s.cam.XBound(dscl, dx)
			} else {
				dscl = s.drawScale / s.cam.BaseScale()
				dx = x + s.zoomPosXLag/scl
			}
			dy = y + s.zoomPosYLag/scl
		} else {
			s.zoomlag = 0
			s.zoomPosXLag = 0
			s.zoomPosYLag = 0
			s.zoomScale = 1
			s.zoomPos = [2]float32{0, 0}
			s.drawScale = s.cam.Scale
		}
		s.draw(dx, dy, dscl)
	}

	if !s.frameSkip {
		s.luaFlushDrawQueue()
	} else {
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

	if s.matchTime == 0 {
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
			s.await(s.cfg.Video.Framerate * 4)
		} else {
			s.await(s.cfg.Video.Framerate)
		}
		return s.replayFile.Update()
	}

	if s.netConnection != nil {
		s.await(s.cfg.Video.Framerate)
		return s.netConnection.Update()
	}

	return s.await(s.cfg.Video.Framerate)
}

func (s *System) tickSound() {
	s.soundChannels.Tick()
	if !s.noSoundFlg {
		for _, ch := range s.chars {
			for _, c := range ch {
				c.soundChannels.Tick()
			}
		}
	}

	// Always pause if noMusic flag set, pause master volume is 0, or freqmul is 0.
	s.bgm.SetPaused(s.nomusic || (s.paused && s.cfg.Sound.PauseMasterVolume == 0) || (s.bgm.freqmul == 0))

	// Set BGM volume if paused
	if s.paused && s.bgm.volRestore == 0 {
		s.bgm.volRestore = s.bgm.bgmVolume
		s.bgm.bgmVolume = int(s.cfg.Sound.PauseMasterVolume * s.bgm.bgmVolume / 100.0)
		s.bgm.UpdateVolume()
		s.softenAllSound()
	} else if !s.paused && s.bgm.volRestore > 0 {
		// Restore all volume
		s.bgm.bgmVolume = s.bgm.volRestore
		s.bgm.volRestore = 0
		s.bgm.UpdateVolume()
		s.restoreAllVolume()
	}
}

func (s *System) resetCommandInputSource() {
	for i := range s.commandInputSource {
		s.commandInputSource[i] = i
	}
}

func (s *System) resetRemapInput() {
	for i := range s.inputRemap {
		s.inputRemap[i] = i
	}
}

func (s *System) loaderReset() {
	s.round, s.wins, s.roundsExisted, s.decisiveRound = 1, [2]int32{}, [2]int32{}, [2]bool{}
	s.loader.reset()
}

func (s *System) loadStart() {
	s.loaderReset()
	s.loader.runTread()
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

func (s *System) button(btns []string, controllerNo int) bool {
	// Direction token -> matching LS axis token.
	dirOf := func(tok string) byte {
		if len(tok) == 2 && tok[0] == '$' {
			tok = tok[1:]
		}
		if len(tok) != 1 {
			return 0
		}
		switch tok[0] {
		case 'D', 'U', 'B', 'F':
			return tok[0]
		default:
			return 0
		}
	}
	axisOf := func(dir byte) string {
		switch dir {
		case 'D':
			return "LS_Y+"
		case 'U':
			return "LS_Y-"
		case 'B':
			return "LS_X-"
		case 'F':
			return "LS_X+"
		default:
			return ""
		}
	}

	lastAxis := ""
	switch s.uiLastInputToken {
	case "LS_Y+", "LS_Y-", "LS_X-", "LS_X+":
		lastAxis = s.uiLastInputToken
	}
	lastDir := dirOf(s.uiLastInputToken)

	checkOne := func(i int, cl *CommandList) bool {
		if cl == nil {
			return false
		}
		// Resolve which controller slot should feed this command list.
		src := i
		if i >= 0 && i < len(s.commandInputSource) {
			src = s.commandInputSource[i]
		}
		for _, btn := range btns {
			// Command-system state
			dir := dirOf(btn)
			if cl.GetState(btn) {
				// Avoid "same direction, different type" conflict (axis->dir) on the same controller.
				if dir != 0 && s.lastInputController == src && lastAxis == axisOf(dir) {
					return false
				}
				s.lastInputController = src
				s.uiLastInputToken = btn
				return true
			}

			// Also allow LS axis for D/U/B/F (and $D/$U/$B/$F).
			if dir != 0 {
				axisTok := axisOf(dir)
				if axisTok != "" {
					// Avoid "same direction, different type" conflict (dir->axis) on the same controller.
					if s.lastInputController == src && lastDir == dir {
						return false
					}
					if cl.IsControllerButtonPressed(axisTok, src) {
						return true
					}
				}
			}
		}
		return false
	}

	if controllerNo >= 0 {
		if controllerNo < len(s.commandLists) {
			return checkOne(controllerNo, s.commandLists[controllerNo])
		}
		return false
	}

	for i := 0; i < len(s.commandLists); i++ {
		if checkOne(i, s.commandLists[i]) {
			return true
		}
	}
	return false
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
		// controller index is 0-based here.
		// Allow overriding the input source per command list
		controller := i
		if i >= 0 && i < len(s.commandInputSource) {
			controller = s.commandInputSource[i]
		}
		// Step commands only if the buffer has already stepped. Prevents rapid fire inputs
		if cl.InputUpdate(nil, controller) {
			cl.Step(false, false, false, false, 0)
		}
	}
}

func (s *System) netplay() bool {
	return sys.rollback.session != nil || sys.netConnection != nil || sys.replayFile != nil
}

func (s *System) escExit() bool {
	return (sys.gameMode == "demo" && (sys.esc || sys.button([]string{"m"}, -1))) ||
		(sys.esc && (sys.netplay() || !sys.cfg.Config.EscOpensMenu || sys.gameMode == "" || (sys.motif.AttractMode.Enabled && sys.credits == 0)))
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
	ch, ok := s.charList.idMap[id]
	if !ok {
		return nil
	}

	// Mugen skips DestroySelf helpers here
	if ch.csf(CSF_destroy) {
		return nil
	}

	return ch
}

func (s *System) playerIndex(idx int32) *Char {
	if idx < 0 || int(idx) >= len(sys.charList.runOrder) {
		return nil
	}

	// We will ignore destroyed helpers here, like Mugen redirections do
	var searchIdx int32
	for _, p := range sys.charList.creationOrder {
		if p != nil && !p.csf(CSF_destroy) {
			if searchIdx == idx {
				return p
			}
			searchIdx++
		}
	}

	//if idx >= 0 && int(idx) < len(s.charList.runOrder) {
	//	return s.charList.runOrder[idx]
	//}

	return nil
}

// We must check if wins are greater than 0 because modes like Training may have "0 rounds to win"
func (s *System) matchOver() bool {
	return s.wins[0] > 0 && s.wins[0] >= s.matchWins[0] ||
		s.wins[1] > 0 && s.wins[1] >= s.matchWins[1]
}

func (s *System) playerIDExist(id BytecodeValue) BytecodeValue {
	if id.IsSF() {
		return BytecodeSF()
	}
	return BytecodeBool(s.playerID(id.ToI()) != nil)
}

// TODO: This is redundant since the index always exists if "NumPlayer >= idx-1"
// Maybe remove it or make it ignore destroyed helpers at least
func (s *System) playerIndexExist(idx BytecodeValue) BytecodeValue {
	if idx.IsSF() {
		return BytecodeSF()
	}
	return BytecodeBool(s.playerIndex(idx.ToI()) != nil)
}

func (s *System) playerNoExist(no BytecodeValue) BytecodeValue {
	if no.IsSF() {
		return BytecodeSF()
	}
	exist := false
	number := int(no.ToI() - 1)
	if number >= 0 && number < len(sys.chars) {
		exist = len(sys.chars[number]) > 0
	}
	return BytecodeBool(exist)
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
	return s.intro < -s.lifebar.ro.over_hittime
}

// Characters cannot hurt each other between lifebar timers over.hittime and over.waittime
func (s *System) roundNoDamage() bool {
	return sys.intro < 0 && sys.intro <= -sys.lifebar.ro.over_hittime && sys.intro >= -sys.lifebar.ro.over_waittime
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
	case sys.intro > sys.lifebar.ro.ctrl_time+1 || sys.postMatchFlg:
		return 0
	//case sys.lifebar.ro.current == 0:
	case sys.intro > 0:
		return 1
	//case sys.intro >= 0 || sys.finishType == FT_NotYet:
	case sys.intro == 0 || sys.finishType == FT_NotYet:
		return 2
	case sys.intro < -sys.lifebar.ro.over_waittime:
		return 4
	default:
		return 3
	}
}

func (s *System) introState() int32 {
	switch {
	case s.intro > s.lifebar.ro.ctrl_time+1:
		// Pre-intro [RoundState = 0]
		return 1
	case (s.motif.di.active && s.dialogueForce == 0) || s.intro == s.lifebar.ro.ctrl_time+1:
		// Player intros [RoundState = 1]
		return 2
	case s.intro == s.lifebar.ro.ctrl_time:
		// Dialogue detection (s.motif.di.active is detectable 1 frame later)
		if s.motif.isDialogueSet() {
			return 2
		}
		// Round announcement
		return 3
	case s.lifebar.ro.waitTimer[1] == -1 || (s.intro > 0 && s.intro < s.lifebar.ro.ctrl_time):
		// Fight called
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
	case sys.intro <= -sys.lifebar.ro.over_waittime && sys.winposetime > 0:
		// Players lose control, but the round has not yet entered win states
		return 3
	case s.intro < -s.lifebar.ro.over_hittime || sys.lifebar.ro.over_hittime == 1:
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
	return s.intro < -(s.lifebar.ro.over_waittime + s.lifebar.ro.over_time)
}

func (s *System) roundStateTicks() int32 {
	return s.intro + s.lifebar.ro.over_waittime + s.lifebar.ro.over_time
}

// Check if the match consists of a single round
func (s *System) roundIsSingle() bool {
	return !s.sel.gameParams.PersistRounds && s.round == 1 && s.decisiveRound[0] && s.decisiveRound[1]
}

// This checks if a round is eligible for "Final Round" behavior, not if it's literally the final round
// https://github.com/ikemen-engine/Ikemen-GO/issues/1659
func (s *System) roundIsFinal() bool {
	return !s.sel.gameParams.PersistRounds && s.round > 1 && s.decisiveRound[0] && s.decisiveRound[1] &&
		(s.draws >= s.lifebar.ro.match_maxdrawgames[0] || s.draws >= s.lifebar.ro.match_maxdrawgames[1])
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
	s.luaDrawPreOps = append(s.luaDrawPreOps, fn)
}
func (s *System) luaQueueLayerDraw(layer int, fn func()) {
	if fn == nil {
		return
	}
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
func (s *System) printBytecodeError(str string) {
	if s.loader.state == LS_Complete && s.workingChar != nil {
		// Print during matches
		s.appendToConsole(sys.workingChar.warn() + str)
	} else if !sys.ignoreMostErrors {
		// Print outside matches (compiling)
		sys.errLog.Println(str)
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

		if sys.round == 1 || c.roundsExisted() == 0 {
			c.id = sys.newCharId()
		}
	}

	// Free some ID's in subsequent rounds
	if sys.round > 1 {
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

	for _, p := range s.chars {
		if len(p) > 0 && p[0] != nil && p[0].id > s.lastCharId {
			s.lastCharId = p[0].id
		}
		// At this point we're still checking the previous round's player ID's
		// However that works out well for Turns mode because it means a joining player won't reuse the ID of a defeated player
	}
}

// Determine the next available character ID while keeping track of the last number used
func (s *System) newCharId() int32 {
	newid := s.lastCharId + 1

	// Check if the next ID is already being used
	// This is needed because helpers may be preserved between rounds
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

	s.lastCharId = newid
	return newid
}

func (s *System) resetGblEffect() {
	s.allPalFX.clear()
	s.bgPalFX.clear()
	s.envShake.clear()
	s.pausetime, s.pausetimebuffer = 0, 0
	s.supertime, s.supertimebuffer = 0, 0
	s.envcol_time = 0
	s.specialFlag = 0
}

func (s *System) stopAllCharSound() {
	for _, p := range s.chars {
		for _, c := range p {
			c.soundChannels.SetSize(0)
		}
	}
}

func (s *System) softenAllSound() {
	for _, p := range s.chars {
		for _, c := range p {
			for i := 0; i < int(c.soundChannels.count()); i++ {
				// Temporarily store the volume so it can be recalled later.
				if c.soundChannels.channels[i].sfx != nil && c.soundChannels.channels[i].ctrl != nil {
					c.soundChannels.volResume[i] = c.soundChannels.channels[i].sfx.volume
					c.soundChannels.channels[i].SetVolume(float32(c.gi().data.volume * int32(s.cfg.Sound.PauseMasterVolume) / 100))

					// Pause if pause master volume is 0
					if s.cfg.Sound.PauseMasterVolume == 0 {
						c.soundChannels.channels[i].SetPaused(true)
					}
				}
			}
		}
	}
	// Don't pause motif sounds
}

func (s *System) restoreAllVolume() {
	for _, p := range s.chars {
		for _, c := range p {
			for i := 0; i < int(c.soundChannels.count()); i++ {
				// Restore the volume we had.
				if c.soundChannels.channels[i].sfx != nil && c.soundChannels.channels[i].ctrl != nil {
					c.soundChannels.channels[i].SetVolume(c.soundChannels.volResume[i])

					// Unpause only those whose freqmul > 0
					if c.soundChannels.channels[i].ctrl.Paused && c.soundChannels.channels[i].sfx.freqmul > 0 {
						c.soundChannels.channels[i].SetPaused(false)
					}
				}
			}
		}
	}
}

func (s *System) clearMatchSound() {
	s.stopAllCharSound()
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
			c.soundChannels.SetSize(0)
		}

		// Destroy helpers
		for _, h := range s.chars[pn][1:] {
			// F4 now destroys "preserve" helpers when reloading round start backup
			//if forceDestroy || h.preserve == 0 || (s.roundResetFlg && h.preserve <= s.round) {
			if !h.preserve || forceDestroy {
				h.destroy()
			}
		}

		// Clear children "family tree"
		// We only do this because helpers may be preserved
		for _, c := range s.chars[pn] {
			// Iterate backwards since we're removing items
			for i := len(c.children) - 1; i >= 0; i-- {
				if c.children[i].helperIndex < 0 {
					c.children = SliceDelete(c.children, i)
				}
			}
		}
	}

	// Clear projectiles, explods and text sprites
	s.projs[pn] = PointerSliceReset(s.projs[pn])
	s.explods[pn] = PointerSliceReset(s.explods[pn])
	s.chartexts[pn] = PointerSliceReset(s.chartexts[pn])
}

func (s *System) resetRoundState() {
	s.roundBackup.Restore()
	s.resetFrameTime()

	s.paused = false
	s.introSkipCall = false
	s.roundResetFlg = false
	s.reloadFlg, s.reloadStageFlg, s.reloadLifebarFlg = false, false, false

	s.resetGblEffect()
	s.lifebar.reset()
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
	s.slowtime = s.lifebar.ro.slow_time
	s.winposetime = s.lifebar.ro.over_wintime
	s.winwaittime = s.lifebar.ro.over_waittime + s.lifebar.ro.over_forcewintime
	s.winskipped = false
	s.intro = s.lifebar.ro.start_waittime + s.lifebar.ro.ctrl_time + 1
	s.curRoundTime = s.maxRoundTime
	s.curPlayTime = 0
	// Mugen resets the starting ID between matches but not between rounds
	// Previously Ikemen reset it between rounds, but that creates the odd scenario where a new player in Turns mode will have the same ID as a previous player
	//s.nextCharId = s.cfg.Config.HelperMax

	// Decisive round check
	for i := range s.decisiveRound {
		neededWins := s.lifebar.ro.match_wins[i]
		// If enemy is in Turns, then we must win "team size" rounds
		if s.tmode[1-i] == TM_Turns {
			neededWins = s.numTurns[1-i]
		}
		if s.wins[i] >= neededWins-1 {
			s.decisiveRound[i] = true
		}
	}

	var roundRef int32
	if s.round == 1 {
		s.stageLoopNo = 0
	} else {
		roundRef = s.round
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

	var swap bool
	if _, ok := s.stageList[roundRef]; ok {
		s.stage = s.stageList[roundRef]
		if s.round > 1 && !s.roundResetFlg {
			swap = true
		}
		if s.stage.model != nil {
			sys.mainThreadTask <- func() {
				gfx.SetModelVertexData(0, s.stage.model.vertexBuffer)
				gfx.SetModelIndexData(0, s.stage.model.elementBuffer...)
			}
		}
	}

	s.cam.stageCamera = s.stage.stageCamera
	s.cam.Init()
	s.screenleft = float32(s.stage.screenleft) * s.stage.localscl
	s.screenright = float32(s.stage.screenright) * s.stage.localscl

	if s.stage.resetbg || swap {
		s.stage.reset()
	}
	s.cam.ResetZoomdelay()

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
		s.cgi[i].clearPCTime()

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
		// In Versus modes, round/final/life music shouldn't override stage music; victory.music still can.
		if useCharMusic {
			s.cgi[i].music.Override(p[0].si().music)
		} else if v, ok := p[0].si().music[normalizeMusicPrefix("victory")]; ok {
			s.cgi[i].music.Override(Music{normalizeMusicPrefix("victory"): v})
		}
		// Override with music with launchFight parameters
		s.cgi[i].music.Override(sys.sel.music)
		// Debug dump of merged music.
		s.cgi[i].music.DebugDump(fmt.Sprintf("cgi[%d].music after resetRoundState", i))
	}

	// Place characters in state 5900
	for _, p := range s.chars {
		if len(p) == 0 {
			continue
		}
		// Select anim 0
		firstAnim := int32(0)
		// Default to first anim in .AIR if 0 was not found
		if p[0].gi().animTable[0] == nil {
			for k := range p[0].gi().animTable {
				firstAnim = k
				break
			}
		}
		p[0].selfState(5900, firstAnim, -1, 0, "")
	}

	// Backup must reflect the post-swap roundXdef stage, or F4 restores the prior round's stage.
	if s.stage != nil {
		s.roundBackup.Save()
	}
}

func (s *System) resetRound() {
	s.resetRoundState()
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
		return ClampF(progress, 0, 1)
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
	s.tickCount, s.oldTickCount, s.tickCountF, s.lastTick, s.absTickCountF = 0, -1, 0, 0, 0
	s.nextAddTime, s.oldNextAddTime = 1, 1
}

func (s *System) resetMatchData(assets bool) {
	sys.allPalFX = newPalFX()
	sys.bgPalFX = newPalFX()
	sys.resetGblEffect()
	for i, p := range sys.chars {
		if len(p) > 0 {
			sys.clearPlayerAssets(i, assets)
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

func (s *System) posReset() {
	for _, p := range s.chars {
		if len(p) > 0 {
			p[0].posReset()
		}
	}
}

// Skip character intros on button press and play the shutter effect
func (s *System) runIntroSkip() {
	// If no intros to skip or not allowed to
	if !s.gsf(GSF_intro) || s.gsf(GSF_roundnotskip) {
		return
	}

	// If too late to skip intros
	if s.intro <= s.lifebar.ro.ctrl_time || s.lifebar.ro.current >= 1 {
		return
	}

	// Start shutter effect on button press
	if s.lifebar.ro.shutterTimer == 0 && s.anyButton() {
		s.lifebar.ro.shutterTimer = s.lifebar.ro.shutter_time * 2 // Open + close time
	}

	// Skip intros when signal from shutter animation arrives
	if s.introSkipCall {
		s.introSkipCall = false
		s.intro = s.lifebar.ro.ctrl_time

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
	s.spritesLayerN1 = s.spritesLayerN1[:0]
	s.spritesLayerU = s.spritesLayerU[:0]
	s.spritesLayer0 = s.spritesLayer0[:0]
	s.spritesLayer1 = s.spritesLayer1[:0]

	// Shadows and reflections
	s.shadows = s.shadows[:0]
	s.reflections = s.reflections[:0]

	// Debug sprites
	s.debugc1hit = s.debugc1hit[:0]
	s.debugc1rev = s.debugc1rev[:0]
	s.debugc1not = s.debugc1not[:0]
	s.debugc2 = s.debugc2[:0]
	s.debugc2hb = s.debugc2hb[:0]
	s.debugc2mtk = s.debugc2mtk[:0]
	s.debugc2grd = s.debugc2grd[:0]
	s.debugc2stb = s.debugc2stb[:0]
	s.debugcsize = s.debugcsize[:0]
	s.debugch = s.debugch[:0]
	s.clsnText = nil

	// Reset afterimage tracker
	for i := range s.afterImageCount {
		s.afterImageCount[i] = 0
	}
}

func (s *System) action() {
	s.clearSpriteData()

	var x, y, scl float32 = s.cam.Pos[0], s.cam.Pos[1], s.cam.Scale / s.cam.BaseScale()
	s.cam.ResetTracking()

	// Run "tick frame"
	if s.tickFrame() {
		// X axis player limits
		s.xmin = s.cam.ScreenPos[0] + s.cam.Offset[0] + s.screenleft
		s.xmax = s.cam.ScreenPos[0] + s.cam.Offset[0] + float32(s.gameWidth)/s.cam.Scale - s.screenright
		if s.xmin > s.xmax {
			s.xmin = (s.xmin + s.xmax) / 2
			s.xmax = s.xmin
		}
		if AbsF(s.cam.maxRight-s.xmax) < 0.0001 {
			s.xmax = s.cam.maxRight
		}
		if AbsF(s.cam.minLeft-s.xmin) < 0.0001 {
			s.xmin = s.cam.minLeft
		}
		// Z axis player limits
		s.zmin = s.stage.topbound * s.stage.localscl
		s.zmax = s.stage.botbound * s.stage.localscl
		s.allPalFX.step()
		//s.bgPalFX.step()
		s.envShake.next()
		if s.envcol_time > 0 {
			s.envcol_time--
		}
		if s.enableZoomtime > 0 {
			s.enableZoomtime--
		} else {
			s.zoomCameraBound = true
			s.zoomStageBound = true
		}
		if s.supertime > 0 {
			s.supertime--
		} else if s.pausetime > 0 {
			s.pausetime--
		}
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
			// "Intro" seems to have been deliberately added. Does not persist in Mugen 1.1
			// "NoKOSlow" added to facilitate custom slowdown. In Mugen that flag only needs to be asserted in first frame of KO slowdown
			s.specialFlag = (s.specialFlag&GSF_intro | s.specialFlag&GSF_nokoslow | s.specialFlag&GSF_timerfreeze)
		}
		s.charList.action()
		s.nomusic = s.gsf(GSF_nomusic) && !sys.postMatchFlg
	}

	// This function runs every tick
	// It should be placed between "tick frame" and "tick next frame"
	s.charUpdate()

	// Update round state
	// This is also reflected on characters (intros, win poses)
	// It's important that this is placed after the tickFrame logic, or characters will not see every step of the sys.intro timer
	s.stepRoundState()

	// Update lifebars
	// This must happen before hit detection for accurate display
	// Allows a combo to still end if a character is hit in the same frame where it exits movetype H
	s.lifebar.step()

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
	s.xmin = s.cam.ScreenPos[0] + s.cam.Offset[0] + s.screenleft
	s.xmax = s.cam.ScreenPos[0] + s.cam.Offset[0] +
		float32(s.gameWidth)/s.cam.Scale - s.screenright
	if s.xmin > s.xmax {
		s.xmin = (s.xmin + s.xmax) / 2
		s.xmax = s.xmin
	}
	if AbsF(s.cam.maxRight-s.xmax) < 0.0001 {
		s.xmax = s.cam.maxRight
	}
	if AbsF(s.cam.minLeft-s.xmin) < 0.0001 {
		s.xmin = s.cam.minLeft
	}
	s.charList.xScreenBound()

	for i := range s.projs {
		for _, p := range s.projs[i] {
			p.cueDraw()
		}
	}

	s.charList.cueDraw()
	s.explodUpdate()
	s.charTextsUpdate()

	// Adjust game speed
	if s.tickNextFrame() {
		spd := float32(s.gameLogicSpeed()) / float32(s.gameRenderSpeed())

		// KO slowdown
		if st := s.getSlowtime(); st > 0 {
			if !s.gsf(GSF_nokoslow) {
				base := s.lifebar.ro.slow_speed
				fade := s.lifebar.ro.slow_fadetime
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

		s.turbo = spd

		// Force Feedback (legacy)
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
	// Logic
	for i := range s.explods {
		for _, e := range s.explods[i] {
			e.update(i)
		}
	}

	// Cleanup
	for i := range s.explods {
		s.explodPrune(i)
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
	if s.lifebar.ro.timerActive {
		t += s.timeElapsed()
	}
	return t
}

// Step sys.intro timer and execute related tasks
func (s *System) stepRoundState() {
	// Freeze round state if round animations cannot advance
	if !s.lifebar.ro.act() {
		return
	}

	// Fading
	if !(s.lifebar.ro.fadeOut.isActive() || s.lifebar.ro.fadeIn.isActive()) {
		if s.motif.fadeOut.isActive() {
			s.motif.fadeOut.step()
		} else if s.motif.fadeIn.isActive() {
			s.motif.fadeIn.step()
		}
	}

	// Intros
	if s.intro > s.lifebar.ro.ctrl_time {
		s.intro--
		if s.gsf(GSF_intro) && s.intro <= s.lifebar.ro.ctrl_time {
			s.intro = s.lifebar.ro.ctrl_time + 1
		}
	} else if s.intro > 0 {
		if s.intro == s.lifebar.ro.ctrl_time {
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
		rs4t := -s.lifebar.ro.over_waittime
		fadeoutStart := rs4t - 2 - s.lifebar.ro.over_time + s.lifebar.ro.fadeOut.time

		s.intro--

		if s.intro == -s.lifebar.ro.over_hittime && s.finishType != FT_NotYet {
			// Consecutive wins counter
			winner := [2]bool{s.effectiveLoss[1], s.effectiveLoss[0]}
			if !winner[0] || !winner[1] ||
				s.tmode[0] == TM_Turns || s.tmode[1] == TM_Turns ||
				s.draws >= s.lifebar.ro.match_maxdrawgames[0] || s.draws >= s.lifebar.ro.match_maxdrawgames[1] {
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
		if !s.winskipped && s.winposetime < 0 && s.anyButton() && !s.gsf(GSF_roundnotskip) {
			s.intro = Min(s.intro, fadeoutStart)
			s.winskipped = true
		}
		// If the user skipped winposes, don't let RoundNotOver swallow the single fadeoutStart tick.
		if s.intro == fadeoutStart && (!s.gsf(GSF_roundnotover) || s.winskipped) &&
			!s.motif.di.active && !s.lifebar.ro.fadeOut.isActive() {
			s.lifebar.ro.fadeOut.init(s.lifebar.ro.fadeOut, false)
		}

		// Before win poses
		if s.winposetime > 0 {
			// Check if game can proceed into roundstate 4
			if s.winwaittime > 0 {
				if s.intro == rs4t-1 {
					for _, p := range s.chars {
						if len(p) > 0 {
							// Check if this player is ready to proceed to roundstate 4
							// Maybe the "activelyFighting()" could be replaced with an AssertSpecial flag or such to ignore the win/lose states
							// Mugen seems to skip this anim 5 check on time overs
							// It also seems a bit pointless to begin with because the char has already turned by the time anim 5 starts
							if p[0].scf(SCF_over_alive) || p[0].scf(SCF_over_ko) || !p[0].activelyFighting() ||
								(p[0].scf(SCF_ctrl) && p[0].ss.moveType == MT_I && p[0].ss.stateType != ST_A && p[0].ss.stateType != ST_L && p[0].animNo != 5) {
								continue
							}
							// Freeze timer if any player is not ready to proceed yet
							s.intro = rs4t
							break
						}
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
						s.lifebar.wi[i].add(s.winType[i])
						if s.matchOver() {
							// In a draw game both players go back to 0 wins
							if winner[0] == winner[1] {
								s.lifebar.wc[0].wins = 0
								s.lifebar.wc[1].wins = 0
							} else {
								if s.wins[i] >= s.matchWins[i] {
									s.lifebar.wc[i].wins++
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
				// Mugen only checks for readiness in the RoundState 4 loop above. It doesn't care in this one and forces characters into win poses no matter what
				// HitPause is checked here but not in the other loop. Perhaps because changing states during hitpause in Mugen isn't quite safe
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
		if !s.winskipped && s.gsf(GSF_roundnotover) && s.intro == fadeoutStart {
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
		clutchRatio := float32(s.lifebar.ro.clutch_threshold) / 100.0
		for i := team; i < MaxSimul*2; i += 2 {
			if len(s.chars[i]) > 0 {
				char := s.chars[i][0]
				if float32(char.life) > float32(char.lifeMax)*clutchRatio {
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
		for i := 0; i < 2; i++ {
			// TODO: Turns mode could have special handling for balancing team sizes here
			// For instance only allow draws if both teams are the same size
			// In the meantime treating everything the same is the most impartial
			if s.draws >= s.lifebar.ro.match_maxdrawgames[i] {
				s.effectiveLoss[i] = true
			} else {
				s.effectiveLoss[i] = false
			}
		}
	} else {
		for i := range s.effectiveLoss {
			s.effectiveLoss[i] = s.winTeam != i
		}
	}

	// Update win triggers if finish type was changed
	if ft != s.finishType {
		for i, p := range s.chars {
			if len(p) > 0 && ko[^i&1] {
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
	//	FillRect(rect, color, alpha>>uint(Btoi(s.clsnDisplay))+Btoi(s.clsnDisplay)*128)
	//}

	if s.envcol_time == 0 {
		fcol := uint32(0)

		// Draw stage background fill if stage is disabled
		if s.gsf(GSF_nobg) {
			if s.allPalFX.enable {
				var rgb [3]int32
				if s.allPalFX.eInvertall {
					rgb = [...]int32{0xff, 0xff, 0xff}
				}
				for i, v := range rgb {
					rgb[i] = Clamp((v+s.allPalFX.eAdd[i])*s.allPalFX.eMul[i]>>8, 0, 0xff)
				}
				fcol = uint32(rgb[2] | rgb[1]<<8 | rgb[0]<<16)
			}
			FillRect(s.scrrect, fcol, [2]int32{255, 0})
		}

		// Draw normal stage background fill and elements with layerNo == -1
		if !s.gsf(GSF_nobg) {
			if s.stage.debugbg {
				FillRect(s.scrrect, 0xff00ff, [2]int32{255, 0})
			} else {
				fcol = uint32(s.stage.bgclearcolor[2]&0xff | s.stage.bgclearcolor[1]&0xff<<8 | s.stage.bgclearcolor[0]&0xff<<16)
				FillRect(s.scrrect, fcol, [2]int32{255, 0})
			}
			if s.stage.ikemenver[0] != 0 || s.stage.ikemenver[1] != 0 { // This layer did not render in Mugen
				s.stage.draw(-1, bgx, bgy, scl)
			}
		}

		// Draw reflections on layer -1
		if !s.gsf(GSF_globalnoshadow) {
			if s.stage.reflection.layerno < 0 {
				s.reflections.draw(x, y, scl*s.cam.BaseScale())
			}
		}

		// Draw character sprites with layerNo == -1
		s.spritesLayerN1.draw(x, y, scl*s.cam.BaseScale())

		// Draw stage elements with layerNo == 0
		if !s.gsf(GSF_nobg) {
			s.stage.draw(0, bgx, bgy, scl)
		}

		// Draw character sprites with special under flag
		s.spritesLayerU.draw(x, y, scl*s.cam.BaseScale())

		// Draw lifebar layer -1
		s.lifebar.draw(-1)

		// Draw char texts layer -1
		s.drawCharTexts(-1)

		// Draw motif layer -1
		s.motif.draw(-1)

		// Draw shadows
		// Draw reflections on layer 0
		// TODO: Make shadows render in same layers as their sources?
		if !s.gsf(GSF_globalnoshadow) {
			if s.stage.reflection.layerno >= 0 {
				s.reflections.draw(x, y, scl*s.cam.BaseScale())
			}
			s.shadows.draw(x, y, scl*s.cam.BaseScale())
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
		//bl, br := MinF(x, s.cam.boundL), MaxF(x, s.cam.boundR)
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

		// Draw lifebar layer 0
		s.lifebar.draw(0)

		// Draw char texts layer 0
		s.drawCharTexts(0)

		// Draw motif layer 0
		s.motif.draw(0)
	}

	// Draw EnvColor effect
	if s.envcol_time != 0 {
		ecol := uint32(s.envcol[2]&0xff | s.envcol[1]&0xff<<8 | s.envcol[0]&0xff<<16)
		FillRect(s.scrrect, ecol, [2]int32{255, 0})
	}

	// Draw character sprites in layer 0
	if s.envcol_time == 0 || s.envcol_under {
		s.spritesLayer0.draw(x, y, scl*s.cam.BaseScale())
		if s.envcol_time == 0 && !s.gsf(GSF_nofg) {
			s.stage.draw(1, bgx, bgy, scl)
		}
	}

	// Draw lifebar layer 1
	s.lifebar.draw(1)

	// Draw char texts layer 1
	s.drawCharTexts(1)

	// Draw motif layer 1
	s.motif.draw(1)

	// Draw character sprites in layer 1 (old "ontop")
	s.spritesLayer1.draw(x, y, scl*s.cam.BaseScale())

	// Draw lifebar layer 2
	s.lifebar.draw(2)

	// Draw char texts layer 2
	s.drawCharTexts(2)

	// Draw motif layer 2
	s.motif.draw(2)

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
		// Change the first color of the Clsn sprite
		setColor := func(color uint32) {
			s.clsnSpr.Pal[0] = color
			s.clsnSpr.PalTex = s.clsnSpr.CachePalTex(s.clsnSpr.Pal)
		}
		// Clsn1 HitDef
		setColor(0xff0000ff)
		s.debugc1hit.draw(alpha)
		// Clsn1 ReversalDef
		setColor(0xff0040c0)
		s.debugc1rev.draw(alpha)
		// Clsn1 Inactive
		setColor(0xff000080)
		s.debugc1not.draw(alpha)
		// Clsn2
		setColor(0xffff0000)
		s.debugc2.draw(alpha)
		// Clsn2 HitBy
		setColor(0xff808000)
		s.debugc2hb.draw(alpha)
		// Clsn2 Invincible
		setColor(0xff004000)
		s.debugc2mtk.draw(alpha)
		// Clsn2 Guarding
		setColor(0xffc00040)
		s.debugc2grd.draw(alpha)
		// Clsn2 Standby
		setColor(0xff404040)
		s.debugc2stb.draw(alpha)
		// Size
		setColor(0xff303030)
		s.debugcsize.draw(alpha)
		// Crosshair
		setColor(0xffffffff)
		s.debugch.draw(alpha)
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
		y = MaxF(y, 48+240-float32(s.gameHeight))
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
	for _, t := range s.clsnText {
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
				forceDestroy := s.matchOver() || (s.tmode[i&1] == TM_Turns && p[0].life <= 0)
				s.clearPlayerAssets(i, forceDestroy)
			}
		}
	}()

	// Synchronize with external inputs (netplay, replays, etc)
	if err := s.synchronize(); err != nil {
		s.errLog.Println(err.Error())
		s.esc = true
	}
	if s.netConnection != nil {
		defer s.netConnection.Stop()
	}

	// Setup characters
	s.SetupCharRoundStart()

	// Make a new backup once everything is initialized
	s.roundBackup.Save()

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

	s.resetRound()

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
		// TODO: Confirm at which exaact point rollback does its own save/restore and match that
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

		s.bgPalFX.step()
		s.stage.action()

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

		// F4 pressed to reset round
		if s.roundResetFlg && !s.postMatchFlg {
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

		if s.endMatch && !s.lifebar.ro.fadeOut.isActive() {
			break
		}

		// Render frame
		s.renderFrame()

		// Update system. Break if update returns false (game ended)
		if !s.update() {
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
			lmax *= p[0].ocd().lifeRatio * s.cfg.Options.Life / 100

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
									charLm = float32(s.chars[j][0].ocd().lifeMax) * s.chars[j][0].ocd().lifeRatio * s.cfg.Options.Life / 100
								} else {
									charLm = float32(s.chars[j][0].gi().data.life) * s.chars[j][0].ocd().lifeRatio * s.cfg.Options.Life / 100
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
			if p[0].roundsExisted() == 0 && (s.round == 1 || s.tmode[i&1] == TM_Turns) {
				// If round 1 or a new character in Turns mode, initialize values
				if p[0].ocd().life != -1 {
					p[0].life = Clamp(p[0].ocd().life, 0, p[0].lifeMax)
					p[0].redLife = p[0].life
				} else {
					p[0].life = p[0].lifeMax
					p[0].redLife = p[0].lifeMax
				}
				if s.round == 1 {
					if s.maxPowerMode {
						p[0].power = p[0].powerMax
					} else if p[0].ocd().power != -1 {
						p[0].power = Clamp(p[0].ocd().power, 0, p[0].powerMax)
					} else if !sys.sel.gameParams.PersistRounds || sys.consecutiveWins[0] == 0 {
						p[0].power = 0
					}
				}
				p[0].power = Clamp(p[0].power, 0, p[0].powerMax) // Because of previous partner in Turns mode
				p[0].mapArray = make(map[string]float32)
				for k, v := range p[0].mapDefault {
					p[0].mapArray[k] = v
				}
				p[0].dialogue = []string{}
				p[0].remapSpr = make(RemapPreset)
			}
		}
	}
}

func (s *System) runNextRound() bool {
	if s.roundOver() && !s.fightLoopEnd {
		s.clearAllSound()
		s.statsLog.nextRound()
		s.scoreRounds = append(s.scoreRounds, [2]float32{s.lifebar.sc[0].scorePoints, s.lifebar.sc[1].scorePoints})

		if !s.matchOver() &&
			!(s.tmode[0] == TM_Turns && s.effectiveLoss[0]) &&
			!(s.tmode[1] == TM_Turns && s.effectiveLoss[1]) {
			// Prepare for the next round
			s.round++
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
			s.roundBackup.Save()
			s.resetRound()
		} else {
			// End match, or prepare for a new character in turns mode
			for i, tm := range s.tmode {
				if s.chars[i][0].win() || (!s.chars[i][0].lose() && tm != TM_Turns) {
					for j := i; j < len(s.chars); j += 2 {
						if len(s.chars[j]) > 0 {
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
				s.round++
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

func (s *System) gameRenderSpeed() int32 {
	spd := int32(s.cfg.Video.Framerate)
	return Max(1, spd)
}

func (s *System) debugModeAllowed() bool {
	if s.netConnection != nil || s.rollback.session != nil {
		return false
	}
	return s.cfg.Debug.AllowDebugMode
}

func (s *System) IsRollback() bool {
	if s.rollback.session != nil {
		return s.rollback.session.inRollback
	}
	return false
}

type RoundStartBackup struct {
	charBackup    [MaxPlayerNo][]Char
	cgiBackup     [MaxPlayerNo]CharGlobalInfo
	stageBackup   Stage
	oldWins       [2]int32
	oldDraws      int32
	oldTeamLeader [2]int
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
}

func (bk *RoundStartBackup) Restore() {
	// Restore characters
	for i, chars := range sys.chars {
		if len(chars) == 0 {
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

			// Save live sounds before overwriting
			liveSounds := c.soundChannels

			// Restore shallow copy from backup
			*c = *bkup

			// Restore live sounds
			c.soundChannels = liveSounds

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
	ratiopath     string
	movelist      string
	pal           []int32
	pal_defaults  []int32
	pal_keymap    []int32
	pal_files     []string
	localcoord    [2]float32
	portraitscale float32
	cns_scale     [2]float32
	anims         PreloadedAnims
	sff           *Sff
	music         Music
	scp           *SelectCharParams
}

func newSelectChar() *SelectChar {
	return &SelectChar{
		localcoord:    [2]float32{320, 240},
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
	localcoord      [2]float32
	portraitscale   float32
	anims           PreloadedAnims
	sff             *Sff
	music           Music
	ssp             *SelectStageParams
}

func newSelectStage() *SelectStage {
	return &SelectStage{
		localcoord:    [2]float32{320, 240},
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
	sdefOverwrite      string
	music              Music
	gameParams         *GameParams
}

func newSelect() *Select {
	return &Select{
		selectedStageNo:  -1,
		charAnimPreload:  make(map[int32]bool),
		stageAnimPreload: make(map[int32]bool),
		charSpritePreload: map[[2]uint16]bool{[...]uint16{9000, 0}: true,
			[...]uint16{9000, 1}: true},
		stageSpritePreload: make(map[[2]uint16]bool),
		cdefOverwrite:      make(map[int]string),
		music:              make(Music),
		gameParams:         newGameParams(),
	}
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

func (s *Select) SelectStage(n int) { s.selectedStageNo = n }

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

	tstr := fmt.Sprintf("Char: %v", defPathFromSelect)
	defer func() {
		sys.loadTime(tnow, tstr, false, false)
	}()

	if strings.ToLower(defPathFromSelect) == "randomselect" {
		sc.def, sc.name = "randomselect", "Random"
		return nil
	}
	if strings.ToLower(defPathFromSelect) == "dummyslot" {
		sc.name = "dummyslot"
		return nil
	}

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
			sc.name = "dummyslot"
			tstr = fmt.Sprintf("Char: %v (ZIP NOT FOUND)", defPathFromSelect)
			return nil
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
				sc.name = "dummyslot"
				tstr = fmt.Sprintf("Char: %v (DEF IN ZIP MISSING: %s or %s)", defPathFromSelect, defInZip1, defInZip2)
				return nil
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

		foundDiskPath := SearchFile(charDefPathGuess, []string{"chars/", "data/", ""})
		if foundDiskPath == "" || !strings.HasSuffix(strings.ToLower(foundDiskPath), ".def") {
			sc.name = "dummyslot"
			tstr = fmt.Sprintf("Char: %v (DEF NOT FOUND)", defPathFromSelect)
			return nil
		}
		finalDefPath = foundDiskPath
	}

	sc.def = finalDefPath
	if sc.def == "" {
		sc.name = "dummyslot"
		return nil
	}

	charDefContent, err := LoadText(sc.def)
	if err != nil {
		sc.name = "dummyslot"
		tstr = fmt.Sprintf("Char: %v (DEF READ ERROR: %s)", defPathFromSelect, err.Error())
		return nil
	}

	resolvePathRelativeToDef := func(pathInDefFile string) string {
		isZipDef, zipArchiveOfDef, defSubPathInZip := IsZipPath(sc.def)
		pathInDefFile = filepath.ToSlash(pathInDefFile)

		if filepath.IsAbs(pathInDefFile) {
			return pathInDefFile
		}

		// Check if pathInDefFile itself looks like a zip-internal path.
		if isZipRel, _, _ := IsZipPath(pathInDefFile); isZipRel {
			return pathInDefFile // Assume it's a correct logical path
		}

		isEngineRootRelative := strings.HasPrefix(pathInDefFile, "data/") ||
			strings.HasPrefix(pathInDefFile, "font/") ||
			strings.HasPrefix(pathInDefFile, "stages/")

		if isZipDef {
			if isEngineRootRelative {
				return pathInDefFile
			}
			baseDirWithinZip := filepath.ToSlash(filepath.Dir(defSubPathInZip))
			if baseDirWithinZip == "." || baseDirWithinZip == "" { // .def is at zip root
				return filepath.ToSlash(filepath.Join(zipArchiveOfDef, pathInDefFile))
			}
			return filepath.ToSlash(filepath.Join(zipArchiveOfDef, baseDirWithinZip, pathInDefFile))
		}

		return pathInDefFile
	}

	var cns_orig, sprite_orig, anim_orig, movelist_orig string
	var fnt_orig [10][2]string

	lines, i, info, files, keymap, arcade, lanInfo, lanFiles, lanKeymap, lanArcade := SplitAndTrim(charDefContent, "\n"), 0, true, true, true, true, true, true, true, true

	for i < len(lines) {
		isec, name, subname := ReadIniSection(lines, &i)
		switch name {
		case "info":
			if info {
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
				isec.ReadF32("localcoord", &sc.localcoord[0], &sc.localcoord[1])
				isec.ReadF32("portraitscale", &sc.portraitscale)
			}
		case fmt.Sprintf("%v.info", sys.cfg.Config.Language):
			if lanInfo {
				info = false
				lanInfo = false
				var ok bool
				if sc.name, ok, _ = isec.getText("displayname"); !ok {
					sc.name, _, _ = isec.getText("name")
				}
				if sc.lifebarname, ok, _ = isec.getText("lifebarname"); !ok {
					sc.lifebarname = sc.name
				}
				sc.author, _, _ = isec.getText("author")
				sc.pal_defaults = isec.readI32CsvForStage("pal.defaults")
				isec.ReadF32("localcoord", &sc.localcoord[0], &sc.localcoord[1])
				isec.ReadF32("portraitscale", &sc.portraitscale)
			}
		case "files":
			if files {
				files = false
				cns_orig = decodeShiftJIS(isec["cns"])
				sprite_orig = decodeShiftJIS(isec["sprite"])
				anim_orig = decodeShiftJIS(isec["anim"])
				sc.sound = decodeShiftJIS(isec["sound"])
				for i := 1; i <= sys.cfg.Config.PaletteMax; i++ {
					if isec[fmt.Sprintf("pal%v", i)] != "" {
						sc.pal = append(sc.pal, int32(i))
						sc.pal_files = append(sc.pal_files, isec[fmt.Sprintf("pal%v", i)])
					}
				}
				movelist_orig = decodeShiftJIS(isec["movelist"])
				for i_fnt := range fnt_orig {
					fnt_orig[i_fnt][0] = isec[fmt.Sprintf("font%v", i_fnt)]
					fnt_orig[i_fnt][1] = isec[fmt.Sprintf("fnt_height%v", i_fnt)]
				}
			}
		case fmt.Sprintf("%v.files", sys.cfg.Config.Language):
			if lanFiles {
				files = false
				lanFiles = false
				cns_orig = decodeShiftJIS(isec["cns"])
				sprite_orig = decodeShiftJIS(isec["sprite"])
				anim_orig = decodeShiftJIS(isec["anim"])
				sc.sound = decodeShiftJIS(isec["sound"])
				for i := 1; i <= sys.cfg.Config.PaletteMax; i++ {
					if isec[fmt.Sprintf("pal%v", i)] != "" {
						sc.pal = append(sc.pal, int32(i))
						sc.pal_files = append(sc.pal_files, isec[fmt.Sprintf("pal%v", i)])
					}
				}
				movelist_orig = decodeShiftJIS(isec["movelist"])
				for i := range fnt_orig {
					fnt_orig[i][0] = isec[fmt.Sprintf("font%v", i)]
					fnt_orig[i][1] = isec[fmt.Sprintf("fnt_height%v", i)]
				}
			}
		case "palette ": // Note space
			if keymap && len(subname) >= 6 && strings.ToLower(subname[:6]) == "keymap" {
				keymap = false
				sc.pal_keymap = make([]int32, 12)
				for i, v := range [12]string{"a", "b", "c", "x", "y", "z",
					"a2", "b2", "c2", "x2", "y2", "z2"} {
					var i32 int32
					if isec.ReadI32(v, &i32) {
						sc.pal_keymap[i] = i32
					} else {
						sc.pal_keymap[i] = int32(i + 1) // default
					}
				}
			}
		case fmt.Sprintf("%v.palette ", sys.cfg.Config.Language):
			if lanKeymap &&
				len(subname) >= 6 && strings.ToLower(subname[:6]) == "keymap" {
				keymap = false
				sc.pal_keymap = make([]int32, 12)
				for i, v := range [12]string{"a", "b", "c", "x", "y", "z",
					"a2", "b2", "c2", "x2", "y2", "z2"} {
					var i32 int32
					if isec.ReadI32(v, &i32) {
						sc.pal_keymap[i] = i32
					} else {
						sc.pal_keymap[i] = int32(i + 1)
					}
				}
			}
		case "arcade":
			if arcade {
				arcade = false
				sc.intro, _, _ = isec.getText("intro.storyboard")
				sc.ending, _, _ = isec.getText("ending.storyboard")
				sc.arcadepath, _, _ = isec.getText("arcadepath")
				sc.ratiopath, _, _ = isec.getText("ratiopath")
			}
		case fmt.Sprintf("%v.arcade", sys.cfg.Config.Language):
			if lanArcade {
				arcade = false
				lanArcade = false
				sc.intro, _, _ = isec.getText("intro.storyboard")
				sc.ending, _, _ = isec.getText("ending.storyboard")
				sc.arcadepath, _, _ = isec.getText("arcadepath")
				sc.ratiopath, _, _ = isec.getText("ratiopath")
			}
		}
	}
	listSpr := make(map[[2]uint16]bool)
	for k := range s.charSpritePreload {
		listSpr[k] = true
	}

	tempSff := newSff()
	LoadFile(&cns_orig, []string{sc.def, "", "data/"}, func(filename string) error {
		str, err := LoadText(filename)
		if err != nil {
			return err
		}
		lines, i := SplitAndTrim(str, "\n"), 0
		for i < len(lines) {
			is, name, _ := ReadIniSection(lines, &i)
			switch name {
			case "size":
				if ok := is.ReadF32("xscale", &sc.cns_scale[0]); !ok {
					sc.cns_scale[0] = 320 / sc.localcoord[0]
				}
				if ok := is.ReadF32("yscale", &sc.cns_scale[1]); !ok {
					sc.cns_scale[1] = 320 / sc.localcoord[0]
				}
				return nil
			}
		}
		return nil
	})
	// preload animations
	if len(anim_orig) > 0 {
		resolvedAnimPath := resolvePathRelativeToDef(anim_orig)
		LoadFile(&resolvedAnimPath, []string{sc.def}, func(filename string) error {
			str, err := LoadText(filename) // LoadText is zip-aware
			if err != nil {
				return err
			}
			lines, i := SplitAndTrim(str, "\n"), 0
			at := ReadAnimationTable(tempSff, &tempSff.palList, lines, &i) // SFF here is temporary
			for v_anim := range s.charAnimPreload {
				if animation := at.get(v_anim); animation != nil {
					sc.anims.addAnim(animation, v_anim)
					for _, fr := range animation.frames {
						if fr.Group < 0 || fr.Number < 0 {
							continue
						}
						listSpr[[2]uint16{uint16(fr.Group), uint16(fr.Number)}] = true
					}
				}
			}
			return nil
		})
	}
	// preload portion of sff file
	fp := fmt.Sprintf("%v_preload.sff", strings.TrimSuffix(sc.def, filepath.Ext(sc.def)))
	if fp = FileExist(fp); len(fp) == 0 {
		fp = sprite_orig
	}
	if len(fp) > 0 {
		resolvedSpritePath := resolvePathRelativeToDef(fp)
		LoadFile(&resolvedSpritePath, []string{sc.def, "", "data/"}, func(file string) error {
			var selPal []int32
			var err_sff error
			sc.sff, selPal, err_sff = preloadSff(file, true, listSpr)
			if err_sff != nil {
				return fmt.Errorf("failed to preload SFF %s for %s: %w", file, sc.def, err_sff)
			}
			sc.anims.updateSff(sc.sff)
			for k_spr := range s.charSpritePreload {
				sc.anims.addSprite(sc.sff, k_spr[0], k_spr[1])
			}
			// Synchronize SFFv2 internal palettes with DEF declarations
			if sc.sff.header.Version[0] != 1 {
				defPals := make(map[int32]bool)
				for _, p := range sc.pal {
					defPals[p] = true
				}
				// Map DEF palette files to their intended slots
				maxPalSlots := sys.cfg.Config.PaletteMax
				newPalFiles := make([]string, maxPalSlots)
				for i, pIdx := range sc.pal {
					if pIdx >= 1 && int(pIdx) <= maxPalSlots {
						newPalFiles[pIdx-1] = sc.pal_files[i]
					}
				}
				// Rebuild sc.pal and sc.pal_files in sequential order
				sc.pal = nil
				sc.pal_files = nil
				for i := 1; i <= maxPalSlots; i++ {
					existsInSff := false
					for _, sIdx := range selPal {
						if sIdx == int32(i) {
							existsInSff = true
							break
						}
					}
					// Include palette if it exists in either the SFFv2 or the DEF
					if defPals[int32(i)] || existsInSff {
						sc.pal = append(sc.pal, int32(i))
						sc.pal_files = append(sc.pal_files, newPalFiles[i-1])
					}
				}
			} else if len(sc.pal) == 0 {
				sc.pal = selPal
			}
			return nil
		})
	} else {
		sc.sff = newSff()
		sc.anims.updateSff(sc.sff)
		for k := range s.charSpritePreload {
			sc.anims.addSprite(sc.sff, k[0], k[1])
		}
	}
	// read movelist
	if len(movelist_orig) > 0 {
		resolvedMovelistPath := resolvePathRelativeToDef(movelist_orig)
		// Movelist is text, can be loaded now
		LoadFile(&resolvedMovelistPath, []string{sc.def, "", "data/"}, func(filename string) error {
			sc.movelist, _ = LoadText(filename)
			return nil
		})
	}

	return sc
}

func (s *Select) AddStage(def string) (*SelectStage, error) {
	var tstr string
	tnow := time.Now()
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
			err := fmt.Errorf("stage zip not found: %s", defPathFromSelect)
			sys.errLog.Printf("Failed to add stage, file not found: %v\n", defPathFromSelect)
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
				err := fmt.Errorf("def file not found in zip: %s or %s", defInZip1, defInZip2)
				sys.errLog.Printf("Failed to add stage, def file not found in %v: %v or %v\n", defPathFromSelect, defInZip1, defInZip2)
				return nil, err
			}
		}
	} else {
		if !strings.HasSuffix(strings.ToLower(def), ".def") {
			def += ".def"
		}
		if err := LoadFile(&def, []string{"stages/", "data/", ""}, func(file string) error {
			finalDefPath = file
			return nil
		}); err != nil {
			sys.errLog.Printf("Failed to add stage, file not found: %v\n", def)
			return nil, err
		}
	}

	var lines []string
	var err error
	if err = LoadFile(&finalDefPath, nil, func(file string) error {
		var str string
		str, err = LoadText(file)
		if err != nil {
			return err
		}
		lines = SplitAndTrim(str, "\n")
		return nil
	}); err != nil {
		sys.errLog.Printf("Failed to add stage, file not found: %s: %v\n", finalDefPath, err)
		return nil, err
	}
	tstr = fmt.Sprintf("Stage added: %v", finalDefPath)
	i, info, bgdef, stageinfo, lanInfo, lanBgdef, lanStageinfo := 0, true, true, true, true, true, true
	var spr string
	s.stagelist = append(s.stagelist, *newSelectStage())
	ss := &s.stagelist[len(s.stagelist)-1]
	ss.def = finalDefPath
	for i < len(lines) {
		is, name, _ := ReadIniSection(lines, &i)
		switch name {
		case "info":
			if info {
				info = false
				var ok bool
				if ss.name, ok, _ = is.getText("displayname"); !ok {
					if ss.name, ok, _ = is.getText("name"); !ok {
						ss.name = def
					}
				}
				for i := 0; i < MaxAttachedChar; i++ {
					key := "attachedchar"
					if i > 0 {
						key += fmt.Sprint(i + 1) // attachedchar2, attachedchar3, attachedchar4
					}
					if err := is.LoadFile(key, []string{def, "", sys.motif.Def, "data/"}, func(filename string) error {
						// Ensure slice has correct length
						for len(ss.attachedchardef) <= i {
							ss.attachedchardef = append(ss.attachedchardef, "")
						}
						ss.attachedchardef[i] = filename
						return nil
					}); err != nil {
						continue
					}
				}
			}
		case fmt.Sprintf("%v.info", sys.cfg.Config.Language):
			if lanInfo {
				info = false
				lanInfo = false
				var ok bool
				if ss.name, ok, _ = is.getText("displayname"); !ok {
					if ss.name, ok, _ = is.getText("name"); !ok {
						ss.name = def
					}
				}
				for i := 0; i < MaxAttachedChar; i++ {
					key := "attachedchar"
					if i > 0 {
						key += fmt.Sprint(i + 1) // attachedchar2, attachedchar3, attachedchar4
					}
					if err := is.LoadFile(key, []string{def, "", sys.motif.Def, "data/"}, func(filename string) error {
						// Ensure slice has correct length (fill gaps)
						for len(ss.attachedchardef) <= i {
							ss.attachedchardef = append(ss.attachedchardef, "")
						}
						ss.attachedchardef[i] = filename
						return nil
					}); err != nil {
						continue
					}
				}
			}
		case "bgdef":
			if bgdef {
				bgdef = false
				spr = is["spr"]
			}
		case fmt.Sprintf("%v.bgdef", sys.cfg.Config.Language):
			if lanBgdef {
				bgdef = false
				lanBgdef = false
				spr = is["spr"]
			}
		case "stageinfo":
			if stageinfo {
				stageinfo = false
				is.ReadF32("localcoord", &ss.localcoord[0], &ss.localcoord[1])
				is.ReadF32("portraitscale", &ss.portraitscale)
				if _, ok := sys.stageLocalcoords[ss.def]; !ok {
					key := strings.ToLower(filepath.Base(ss.def))
					sys.stageLocalcoords[key] = ss.localcoord // Store localcoords for StageFit
				}
			}
		case fmt.Sprintf("%v.stageinfo", sys.cfg.Config.Language):
			if lanStageinfo {
				stageinfo = false
				lanStageinfo = false
				is.ReadF32("localcoord", &ss.localcoord[0], &ss.localcoord[1])
				is.ReadF32("portraitscale", &ss.portraitscale)
				if _, ok := sys.stageLocalcoords[ss.def]; !ok {
					key := strings.ToLower(filepath.Base(ss.def))
					sys.stageLocalcoords[key] = ss.localcoord
				}
			}
		}
	}
	if len(s.stageSpritePreload) > 0 || len(s.stageAnimPreload) > 0 {
		listSpr := make(map[[2]uint16]bool)
		for k := range s.stageSpritePreload {
			listSpr[[...]uint16{k[0], k[1]}] = true
		}
		sff := newSff()
		// preload animations
		i = 0
		at := ReadAnimationTable(sff, &sff.palList, lines, &i)
		for v := range s.stageAnimPreload {
			if anim := at.get(v); anim != nil {
				ss.anims.addAnim(anim, v)
				for _, fr := range anim.frames {
					if fr.Group >= 0 && fr.Number >= 0 {
						listSpr[[2]uint16{uint16(fr.Group), uint16(fr.Number)}] = true
					}
				}
			}
		}
		// preload portion of sff file
		LoadFile(&spr, []string{def, "", "data/"}, func(file string) error {
			var err error
			ss.sff, _, err = preloadSff(file, false, listSpr)
			if err != nil {
				panic(fmt.Errorf("failed to load %v: %v\nerror preloading %v", file, err, def))
			}
			ss.anims.updateSff(ss.sff)
			for k := range s.stageSpritePreload {
				ss.anims.addSprite(ss.sff, k[0], k[1])
			}
			return nil
		})
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
	sys.loadMutex.Lock()
	s.selected[tn] = append(s.selected[tn], [...]int{n, pl})
	if s.gameParams == nil {
		s.gameParams = newGameParams()
	}
	// ensure per-member override slot exists (needed for existed flag / persistence)
	_ = s.gameParams.ensureOverride(tn, len(s.selected[tn])-1)
	sys.loadMutex.Unlock()
	return true
}

func (s *Select) ClearSelected() {
	sys.loadMutex.Lock()
	s.selected = [2][][2]int{}
	sys.loadMutex.Unlock()
	s.selectedStageNo = -1
	s.music = make(Music)
	s.gameParams = newGameParams()
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

/*
func (l *Loader) loadPlayerChar(pn int) int {
	return l.loadCharacter(pn, false)
}

func (l *Loader) loadAttachedChar(pn int) int {
	return l.loadCharacter(pn, true)
}
*/

func (l *Loader) loadCharacter(pn int, attached bool) int {
	if !attached && sys.roundsExisted[pn&1] > 0 {
		return 1
	}
	if attached && sys.round != 1 {
		return 1
	}

	sys.loadMutex.Lock()
	defer sys.loadMutex.Unlock()

	// Get number of selected characters in team
	memberNo := pn >> 1
	nsel := len(sys.sel.selected[pn&1])

	// Check if player number is acceptable for selected team mode
	if !attached {
		if sys.tmode[pn&1] == TM_Simul || sys.tmode[pn&1] == TM_Tag {
			if memberNo >= int(sys.numSimul[pn&1]) {
				sys.cgi[pn].states = nil
				sys.chars[pn] = nil
				return 1
			}
		} else if pn >= 2 {
			return 0
		}

		if sys.tmode[pn&1] == TM_Turns && nsel < int(sys.numTurns[pn&1]) {
			return 0
		}

		if sys.tmode[pn&1] == TM_Turns {
			memberNo = int(sys.wins[^pn&1])
		}

		if nsel <= memberNo {
			return 0
		}
	}

	idx := make([]int, nsel)
	for i := range idx {
		idx[i] = sys.sel.selected[pn&1][i][0]
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
		if sys.tmode[pn&1] == TM_Turns {
			cdefOWnumber = memberNo*2 + pn&1
		} else {
			cdefOWnumber = pn
		}
		if sys.sel.cdefOverwrite[cdefOWnumber] != "" {
			cdef = sys.sel.cdefOverwrite[cdefOWnumber]
		} else {
			cdef = sys.sel.charlist[idx[memberNo]].def
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
		if prefixToDecrement && !ffx.isGlobal {
			if ffx.refCount > 0 {
				ffx.refCount--
			}
		}
	}

	var p *Char
	sys.workingChar = p // This should help compiler and bytecode stay consistent

	// Reuse existing character or create a new one
	if len(sys.chars[pn]) > 0 && cdef == sys.cgi[pn].def {
		p = sys.chars[pn][0]
		if !attached {
			p.controller = pn
			if sys.aiLevel[pn] != 0 {
				p.controller ^= -1
			}
		}
		p.clearCachedData()
		if l.err = p.loadFx(cdef); l.err != nil {
			sys.errLog.Printf("Error reloading FX for %s: %v", cdef, l.err)
		}
	} else {
		p = newChar(pn, 0)
		sys.cgi[pn].sff = nil
		sys.cgi[pn].palettedata = nil
		if len(sys.chars[pn]) > 0 {
			p.power = sys.chars[pn][0].power
			p.guardPoints = sys.chars[pn][0].guardPoints
			p.dizzyPoints = sys.chars[pn][0].dizzyPoints
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
		p.selectNo = sys.sel.selected[pn&1][memberNo][0]
		p.teamside = p.playerNo & 1
	}

	if !p.ocd().existed {
		p.initCnsVar()
		p.ocd().existed = true
	}

	sys.chars[pn] = make([]*Char, 1)
	sys.chars[pn][0] = p

	// Load new SFF if previous one was not cached
	if sys.cgi[pn].sff == nil {
		if l.err = p.load(cdef); l.err != nil {
			sys.chars[pn] = nil
			if attached {
				tstr = fmt.Sprintf("WARNING: Failed to load new attached char: %v", cdef)
			} else {
				tstr = fmt.Sprintf("WARNING: Failed to load new char: %v", cdef)
			}
			return -1
		}
		if sys.cgi[pn].states, l.err = newCompiler().Compile(p.playerNo, cdef, p.gi().constants); l.err != nil {
			sys.chars[pn] = nil
			if attached {
				tstr = fmt.Sprintf("WARNING: Failed to compile new attached char states: %v", cdef)
			} else {
				tstr = fmt.Sprintf("WARNING: Failed to compile new char states: %v", cdef)
			}
			return -1
		}
		if attached {
			tstr = fmt.Sprintf("New attached char loaded: %v", cdef)
		} else {
			tstr = fmt.Sprintf("New char loaded: %v", cdef)
		}
	} else {
		if attached {
			tstr = fmt.Sprintf("Cached attached char loaded: %v", cdef)
		} else {
			tstr = fmt.Sprintf("Cached char loaded: %v", cdef)
		}
	}

	if attached {
		sys.cgi[pn].palno = 1
	} else {
		// Get palette number from select screen choice
		sys.cgi[pn].palno = int32(sys.sel.selected[pn&1][memberNo][1])
		// Prepare lifebar portraits
		if pn < len(sys.lifebar.fa[sys.tmode[pn&1]]) && sys.tmode[pn&1] == TM_Turns && sys.round == 1 {
			fa := sys.lifebar.fa[sys.tmode[pn&1]][pn]
			fa.numko = 0
			fa.teammate_face = make([]*Sprite, nsel)
			fa.teammate_scale = make([]float32, nsel)
			sys.lifebar.nm[sys.tmode[pn&1]][pn].numko = 0
			for i, ci := range idx {
				fa.teammate_scale[i] = sys.sel.charlist[ci].portraitscale * 320 / sys.sel.charlist[ci].localcoord[0]
				fa.teammate_face[i] = sys.sel.charlist[ci].sff.GetSprite(uint16(fa.teammate_face_spr[0]), uint16(fa.teammate_face_spr[1]))
			}
		}
	}

	return 1
}

func (l *Loader) loadStage() bool {
	if sys.round == 1 {
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
		var def string
		if sys.sel.selectedStageNo == 0 {
			randomstageno := Rand(0, int32(len(sys.sel.stagelist))-1)
			def = sys.sel.stagelist[randomstageno].def
		} else {
			def = sys.sel.stagelist[sys.sel.selectedStageNo-1].def
		}
		if sys.sel.sdefOverwrite != "" {
			def = sys.sel.sdefOverwrite
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
		sys.stage = sys.stageList[0]
		tstr = fmt.Sprintf("New stage loaded: %v", def)
	}
	return l.err == nil
}

func (l *Loader) load() {
	defer func() {
		l.loadExit <- l.state
	}()

	// Update aspect ratio
	sys.applyFightAspect()

	// Update cached stage scaling
	// In case FightAspect option was changed between matches
	if sys.stage != nil {
		sys.stage.localscl = float32(sys.gameWidth) / float32(sys.stage.stageCamera.localcoord[0])
		sys.stage.stageCamera.localscl = sys.stage.localscl
	}

	// Update cached character scaling
	for _, p := range sys.chars {
		if len(p) > 0 {
			p[0].localcoord = p[0].gi().localcoord[0] / (float32(sys.gameWidth) / 320)
			p[0].localscl = 320 / p[0].localcoord
		}
	}

	// Update lifebar scale
	sys.lifebar.setLifebarScale()
	//sys.motif.setMotifScale()
	sys.loadMutex.Lock()
	for prefix, ffx := range sys.ffx {
		if ffx.isGlobal {
			continue
		}
		if ffx.refCount <= 0 {
			if ffx.sff != nil {
				removeSFFCache(ffx.sff.filename)
			}
			delete(sys.ffx, prefix)
			//sys.errLog.Printf("Unloaded CommonFX: %s (prefix: %s)", ffx.fileName, prefix)
		}
	}
	sys.loadMutex.Unlock()

	charDone, stageDone := make([]bool, len(sys.chars)), false

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
		// Load stage
		if !stageDone && sys.sel.selectedStageNo >= 0 {
			if !l.loadStage() {
				l.state = LS_Error
				return
			}
			stageDone = true
		}
		// Load characters that aren't already loaded
		for i, b := range charDone {
			if !b {
				result := -1
				if i < len(sys.chars)-MaxAttachedChar ||
					len(sys.stageList[0].attachedchardef) <= i-MaxSimul*2 {
					result = l.loadCharacter(i, false)
				} else {
					result = l.loadCharacter(i, true)
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
			if !charDone[i+2] && len(sys.sel.selected[i]) > 0 &&
				sys.tmode[i] != TM_Simul && sys.tmode[i] != TM_Tag {
				for j := i + 2; j < len(sys.chars); j += 2 {
					if !charDone[j] {
						sys.chars[j] = nil
						sys.cgi[j].states = nil
						sys.cgi[j].hitPauseToggleFlagCount = 0
						charDone[j] = true
					}
				}
			}
		}
		time.Sleep(10 * time.Millisecond)
		if sys.gameEnd {
			l.state = LS_Cancel
		}
		if l.state == LS_Cancel {
			return
		}
	}

	// Flag loading state as complete
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
			sys.cgi[i].palno = -1
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

type EnvShake struct {
	time  int32
	freq  float32
	ampl  float32
	phase float32
	mul   float32
	dir   float32 // rad, for ampl=-4:  0: down first, 90: left first, 180: up first, 270: right first
}

func (es *EnvShake) clear() {
	*es = EnvShake{
		freq:  float32(math.Pi / 3),
		ampl:  -4.0,
		phase: float32(math.NaN()),
		mul:   1.0,
		dir:   0.0}
}

func (es *EnvShake) setDefaultPhase() {
	if math.IsNaN(float64(es.phase)) {
		if es.freq >= math.Pi/2 {
			es.phase = math.Pi / 2
		} else {
			es.phase = 0
		}
	}
}

func (es *EnvShake) next() {
	if es.time > 0 {
		es.time--
		es.phase += es.freq
		if es.phase > math.Pi*2 {
			es.ampl *= es.mul
			es.phase -= math.Pi * 2
		}
	} else {
		es.ampl = 0
	}
}

func (es *EnvShake) getOffset() [2]float32 {
	if es.time > 0 {
		offset := (es.ampl * float32(math.Sin(float64(es.phase))))
		return [2]float32{offset * float32(math.Sin(float64(-es.dir))),
			offset * float32(math.Cos(float64(-es.dir)))}
	}
	return [2]float32{0, 0}
}

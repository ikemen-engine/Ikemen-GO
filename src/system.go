package main

import (
	"bufio"
	"fmt"
	"image"
	"io"
	"io/ioutil"
	"log"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	gl "github.com/fyne-io/gl-js"
	glfw "github.com/fyne-io/glfw-js"
	lua "github.com/yuin/gopher-lua"
)

const (
	MaxSimul        = 4
	MaxAttachedChar = 1
)

var (
	FPS           = 60
	Mp3SampleRate = 44100
)

// sys
// The only instance of a System struct.
// Do not create more than 1.
var sys = System{
	randseed:          int32(time.Now().UnixNano()),
	scrrect:           [...]int32{0, 0, 320, 240},
	gameWidth:         320,
	gameHeight:        240,
	widthScale:        1,
	heightScale:       1,
	brightness:        256,
	roundTime:         -1,
	lifeMul:           1,
	team1VS2Life:      1,
	turnsRecoveryRate: 1.0 / 300,
	soundMixer:        &beep.Mixer{},
	bgm:               *newBgm(),
	soundChannels:     newSoundChannels(16),
	allPalFX:          *newPalFX(),
	bgPalFX:           *newPalFX(),
	sel:               *newSelect(),
	keyState:          make(map[glfw.Key]bool),
	match:             1,
	listenPort:        "7500",
	loader:            *newLoader(),
	numSimul:          [...]int32{2, 2}, numTurns: [...]int32{2, 2},
	ignoreMostErrors:      true,
	superpmap:             *newPalFX(),
	stageList:             make(map[int32]*Stage),
	wincnt:                wincntMap(make(map[string][]int32)),
	wincntFileName:        "save/autolevel.save",
	powerShare:            [...]bool{true, true},
	oldNextAddTime:        1,
	commandLine:           make(chan string),
	cam:                   *newCamera(),
	statusDraw:            true,
	mainThreadTask:        make(chan func(), 65536),
	workpal:               make([]uint32, 256),
	errLog:                log.New(NewLogWriter(), "", log.LstdFlags),
	keyInput:              glfw.KeyUnknown,
	wavChannels:           256,
	comboExtraFrameWindow: 1,
	fontShaderVer:         120,
	//FLAC_FrameWait:          -1,
	luaSpriteScale:       1,
	luaPortraitScale:     1,
	lifebarScale:         1,
	lifebarPortraitScale: 1,
	vRetrace:             1,
	consoleRows:          15,
	clipboardRows:        2,
	pngFilter:            false,
	clsnDarken:           true,
	maxBgmVolume:         100,
	stereoEffects:        true,
	panningRange:         30,
	windowCentered:       true,
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
	brightness              int32
	roundTime               int32
	lifeMul                 float32
	team1VS2Life            float32
	turnsRecoveryRate       float32
	debugFont               *TextSprite
	debugDraw               bool
	debugRef                [2]int
	soundMixer              *beep.Mixer
	bgm                     Bgm
	soundChannels           *SoundChannels
	allPalFX, bgPalFX       PalFX
	lifebar                 Lifebar
	sel                     Select
	keyState                map[glfw.Key]bool
	netInput                *NetInput
	fileInput               *FileInput
	aiInput                 [MaxSimul*2 + MaxAttachedChar]AiInput
	keyConfig               []KeyConfig
	joystickConfig          []KeyConfig
	com                     [MaxSimul*2 + MaxAttachedChar]float32
	autolevel               bool
	home                    int
	gameTime                int32
	match                   int32
	inputRemap              [MaxSimul*2 + MaxAttachedChar]int
	listenPort              string
	round                   int32
	intro                   int32
	time                    int32
	lastHitter              [2]int
	winTeam                 int
	winType                 [2]WinType
	winTrigger              [2]WinType
	matchWins, wins         [2]int32
	roundsExisted           [2]int32
	draws                   int32
	loader                  Loader
	chars                   [MaxSimul*2 + MaxAttachedChar][]*Char
	charList                CharList
	cgi                     [MaxSimul*2 + MaxAttachedChar]CharGlobalInfo
	tmode                   [2]TeamMode
	numSimul, numTurns      [2]int32
	esc                     bool
	loadMutex               sync.Mutex
	ignoreMostErrors        bool
	stringPool              [MaxSimul*2 + MaxAttachedChar]StringPool
	bcStack, bcVarStack     BytecodeStack
	bcVar                   []BytecodeValue
	workingChar             *Char
	workingState            *StateBytecode
	specialFlag             GlobalSpecialFlag
	afterImageMax           int32
	comboExtraFrameWindow   int32
	envShake                EnvShake
	pause                   int32
	pausetime               int32
	pausebg                 bool
	pauseendcmdbuftime      int32
	pauseplayer             int
	super                   int32
	supertime               int32
	superpausebg            bool
	superendcmdbuftime      int32
	superplayer             int
	superdarken             bool
	superanim               *Animation
	superpmap               PalFX
	superpos                [2]float32
	superfacing             float32
	superp2defmul           float32
	envcol                  [3]int32
	envcol_time             int32
	envcol_under            bool
	stage                   *Stage
	stageList               map[int32]*Stage
	stageLoop               bool
	stageLoopNo             int
	helperMax               int32
	nextCharId              int32
	wincnt                  wincntMap
	wincntFileName          string
	powerShare              [2]bool
	tickCount               int
	oldTickCount            int
	tickCountF              float32
	lastTick                float32
	nextAddTime             float32
	oldNextAddTime          float32
	screenleft              float32
	screenright             float32
	xmin, xmax              float32
	winskipped              bool
	paused, step            bool
	roundResetFlg           bool
	reloadFlg               bool
	reloadStageFlg          bool
	reloadLifebarFlg        bool
	reloadCharSlot          [MaxSimul*2 + MaxAttachedChar]bool
	shortcutScripts         map[ShortcutKey]*ShortcutScript
	turbo                   float32
	commandLine             chan string
	drawScale               float32
	zoomlag                 float32
	zoomScale               float32
	zoomPosXLag             float32
	zoomPosYLag             float32
	enableZoomstate         bool
	zoomCameraBound         bool
	zoomPos                 [2]float32
	debugWC                 *Char
	cam                     Camera
	finish                  FinishType
	waitdown                int32
	slowtime                int32
	shuttertime             int32
	fadeintime              int32
	fadeouttime             int32
	projs                   [MaxSimul*2 + MaxAttachedChar][]Projectile
	explods                 [MaxSimul*2 + MaxAttachedChar][]Explod
	explDrawlist            [MaxSimul*2 + MaxAttachedChar][]int
	topexplDrawlist         [MaxSimul*2 + MaxAttachedChar][]int
	underexplDrawlist       [MaxSimul*2 + MaxAttachedChar][]int
	changeStateNest         int32
	sprites                 DrawList
	topSprites              DrawList
	bottomSprites           DrawList
	shadows                 ShadowList
	drawc1                  ClsnRect
	drawc2                  ClsnRect
	drawc2sp                ClsnRect
	drawc2mtk               ClsnRect
	drawwh                  ClsnRect
	autoguard               [MaxSimul*2 + MaxAttachedChar]bool
	accel                   float32
	clsnSpr                 Sprite
	clsnDraw                bool
	statusDraw              bool
	mainThreadTask          chan func()
	explodMax               int
	workpal                 []uint32
	playerProjectileMax     int
	errLog                  *log.Logger
	nomusic                 bool
	workBe                  []BytecodeExp
	lifeShare               [2]bool
	loseSimul               bool
	loseTag                 bool
	allowDebugKeys          bool
	allowDebugMode          bool
	commonAir               string
	commonCmd               string
	keyInput                glfw.Key
	keyString               string
	timerCount              []int32
	cmdFlags                map[string]string
	wavChannels             int32
	masterVolume            int
	wavVolume               int
	bgmVolume               int
	audioDucking            bool
	windowTitle             string
	screenshotFolder        string
	//FLAC_FrameWait          int

	// Resolution variables
	fullscreen            bool
	fullscreenRefreshRate int32
	fullscreenWidth       int32
	fullscreenHeight      int32

	controllerStickSensitivity float32
	xinputTriggerSensitivity   float32

	// Localcoord sceenpack
	luaLocalcoord    [2]int32
	luaSpriteScale   float32
	luaPortraitScale float32
	luaSpriteOffsetX float32

	// Localcoord lifebar
	lifebarScale         float32
	lifebarOffsetX       float32
	lifebarPortraitScale float32
	lifebarLocalcoord    [2]int32

	// Shader Vars
	postProcessingShader    int32
	multisampleAntialiasing bool
	fontShaderVer           uint

	// External Shader Vars
	externalShaderList  []string
	externalShaderNames []string
	externalShaders     [][]string

	// Icon
	windowMainIcon         []image.Image
	windowMainIconLocation []string

	// Rendering
	borderless bool
	vRetrace   int
	pngFilter  bool // Controls the GL_TEXTURE_MAG_FILTER on 32bit sprites

	gameMode        string
	frameCounter    int32
	motifDir        string
	captureNum      int
	roundType       [2]RoundType
	timerStart      int32
	timerRounds     []int32
	scoreStart      [2]float32
	scoreRounds     [][2]float32
	matchData       *lua.LTable
	consecutiveWins [2]int32
	teamLeader      [2]int
	commonConst     string
	commonLua       []string
	commonStates    []string
	gameSpeed       float32
	maxPowerMode    bool
	clsnText        []ClsnText
	consoleText     []string
	consoleRows     int
	clipboardRows   int
	luaLState       *lua.LState
	statusLFunc     *lua.LFunction
	listLFunc       []*lua.LFunction
	introSkipped    bool
	endMatch        bool
	continueFlg     bool
	dialogueFlg     bool
	dialogueForce   int
	dialogueBarsFlg bool
	noSoundFlg      bool
	postMatchFlg    bool
	playBgmFlg      bool
	brightnessOld   int32
	clsnDarken      bool
	maxBgmVolume    int
	stereoEffects   bool
	panningRange    float32
	windowCentered  bool
}

type Window struct {
	*glfw.Window
	title      string
	fullscreen bool
	x, y, w, h int
}

func (s *System) newWindow(w, h int) (*Window, error) {
	var err error
	var window *glfw.Window
	var monitor *glfw.Monitor

	if monitor = glfw.GetPrimaryMonitor(); monitor == nil {
		return nil, fmt.Errorf("failed to obtain primary monitor")
	}

	var mode = monitor.GetVideoMode()
	var x, y = (mode.Width - w) / 2, (mode.Height - h) / 2

	// "-windowed" overrides the configuration setting but does not change it
	_, forceWindowed := sys.cmdFlags["-windowed"]
	fullscreen := s.fullscreen && !forceWindowed

	// Create main window.
	// NOTE: Borderless fullscreen is in reality just a window without borders.
	if fullscreen && !s.borderless {
		window, err = glfw.CreateWindow(w, h, s.windowTitle, monitor, nil)
	} else {
		window, err = glfw.CreateWindow(w, h, s.windowTitle, nil, nil)
	}
	if err != nil {
		return nil, fmt.Errorf("failed to create window: %w", err)
	}

	// Set windows atributes
	if fullscreen {
		window.SetPos(0, 0)
		if s.borderless {
			window.SetAttrib(glfw.Decorated, 0)
			window.SetSize(mode.Width, mode.Height)
		}
		window.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	} else {
		window.SetSize(w, h)
		window.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
		if s.windowCentered {
			window.SetPos(x, y)
		}
	}

	window.MakeContextCurrent()
	window.SetKeyCallback(keyCallback)
	window.SetCharModsCallback(charCallback)
	ret := &Window{window, s.windowTitle, fullscreen, x, y, w, h}
	return ret, err
}

func (w *Window) toggleFullscreen() {
	var mode = glfw.GetPrimaryMonitor().GetVideoMode()

	if w.fullscreen {
		w.SetAttrib(glfw.Decorated, 1)
		w.SetMonitor(&glfw.Monitor{}, w.x, w.y, w.w, w.h, mode.RefreshRate)
		w.SetInputMode(glfw.CursorMode, glfw.CursorNormal)
	} else {
		w.SetAttrib(glfw.Decorated, 0)
		if sys.borderless {
			w.SetSize(mode.Width, mode.Height)
			w.SetMonitor(&glfw.Monitor{}, 0, 0, mode.Width, mode.Height, mode.RefreshRate)
		} else {
			w.x, w.y = w.GetPos()
			w.SetMonitor(glfw.GetPrimaryMonitor(), w.x, w.y, w.w, w.h, mode.RefreshRate)
		}
		w.SetInputMode(glfw.CursorMode, glfw.CursorHidden)
	}
	if sys.vRetrace != -1 {
		glfw.SwapInterval(sys.vRetrace)
	}
	w.fullscreen = !w.fullscreen
}

// Initialize stuff, this is called after the config int at main.go
func (s *System) init(w, h int32) *lua.LState {
	s.setWindowSize(w, h)
	var err error
	// Create a GLFW window.
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	s.window, err = s.newWindow(int(s.scrrect[2]), int(s.scrrect[3]))
	chk(err)

	// V-Sync
	if s.vRetrace >= 0 {
		glfw.SwapInterval(s.vRetrace)
	}

	// Check if the shader selected is currently available.
	if s.postProcessingShader < int32(len(s.externalShaderList)) {
		s.postProcessingShader = 0
	}

	// Loading of external shader data.
	// We need to do this before the render initialization at "RenderInit()"
	if len(s.externalShaderList) > 0 {
		// First we initialize arrays.
		s.externalShaders = make([][]string, 2)
		s.externalShaderNames = make([]string, len(s.externalShaderList))
		s.externalShaders[0] = make([]string, len(s.externalShaderList))
		s.externalShaders[1] = make([]string, len(s.externalShaderList))

		// Then we load.
		for i, shaderLocation := range s.externalShaderList {
			// Create names.
			shaderLocation = strings.Replace(shaderLocation, "\\", "/", -1)
			splitDir := strings.Split(shaderLocation, "/")
			s.externalShaderNames[i] = splitDir[len(splitDir)-1]

			// Load vert shaders.
			content, err := ioutil.ReadFile(shaderLocation + ".vert")
			if err != nil {
				chk(err)
			}
			s.externalShaders[0][i] = string(content) + "\x00"

			// Load frag shaders.
			content, err = ioutil.ReadFile(shaderLocation + ".frag")
			if err != nil {
				chk(err)
			}
			s.externalShaders[1][i] = string(content) + "\x00"
		}
	}
	// PS: The "\x00" is what is know as Null Terminator.

	// Now we proceed to int the render.
	RenderInit()
	// And the audio.
	speaker.Init(audioFrequency, audioOutLen)
	speaker.Play(NewNormalizer(s.soundMixer))
	l := lua.NewState()
	l.Options.IncludeGoStackTrace = true
	l.OpenLibs()
	for i := range s.inputRemap {
		s.inputRemap[i] = i
	}
	for i := range s.stringPool {
		s.stringPool[i] = *NewStringPool()
	}
	s.clsnSpr = *newSprite()
	s.clsnSpr.Size, s.clsnSpr.Pal = [...]uint16{1, 1}, make([]uint32, 256)
	s.clsnSpr.SetPxl([]byte{0})
	systemScriptInit(l)
	s.shortcutScripts = make(map[ShortcutKey]*ShortcutScript)
	// So now that we have a window we add a icon.
	if len(s.windowMainIconLocation) > 0 {
		// First we initialize arrays.
		var f = make([]io.ReadCloser, len(s.windowMainIconLocation))
		s.windowMainIcon = make([]image.Image, len(s.windowMainIconLocation))
		// And then we load them.
		for i, iconLocation := range s.windowMainIconLocation {
			f[i], err = os.Open(iconLocation)
			if err != nil {
				var dErr = "Icon file can not be found.\nPanic: " + err.Error()
				ShowErrorDialog(dErr)
				panic(Error(dErr))
			}
			s.windowMainIcon[i], _, err = image.Decode(f[i])
		}
		s.window.Window.SetIcon(s.windowMainIcon)
		chk(err)
	}
	// [Icon add end]

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
	return l
}
func (s *System) shutdown() {
	if !sys.gameEnd {
		sys.gameEnd = true
	}
	speaker.Close()
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
func (s *System) eventUpdate() bool {
	s.esc = false
	for _, v := range s.shortcutScripts {
		v.Activate = false
	}
	glfw.PollEvents()
	s.gameEnd = s.window.Window.ShouldClose()
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
func (s *System) await(fps int) bool {
	if !s.frameSkip {
		// Render the finished frame
		unbindFB()
		s.window.Window.SwapBuffers()
		// Begin the next frame
		bindFB()
	}
	s.runMainThreadTask()
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
	s.eventUpdate()
	if !s.frameSkip {
		//var width, height = glfw.GetCurrentContext().GetFramebufferSize()
		//gl.Viewport(0, 0, int32(width), int32(height))
		gl.Viewport(0, 0, int(s.scrrect[2]), int(s.scrrect[3]))
		if s.netInput == nil {
			gl.Clear(gl.COLOR_BUFFER_BIT)
		}
	}
	return !s.gameEnd
}
func (s *System) update() bool {
	s.frameCounter++
	if s.fileInput != nil {
		if s.anyHardButton() {
			s.await(FPS * 4)
		} else {
			s.await(FPS)
		}
		return s.fileInput.Update()
	}
	if s.netInput != nil {
		s.await(FPS)
		return s.netInput.Update()
	}
	return s.await(FPS)
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

	if !s.nomusic {
		speaker.Lock()
		if s.bgm.ctrl != nil && s.bgm.streamer != nil {
			s.bgm.ctrl.Paused = false
			if s.bgm.bgmLoopEnd > 0 && s.bgm.streamer.Position() >= s.bgm.bgmLoopEnd {
				s.bgm.streamer.Seek(s.bgm.bgmLoopStart)
			}
		}
		speaker.Unlock()
	} else {
		s.bgm.Pause()
	}

	//if s.FLAC_FrameWait >= 0 {
	//	if s.FLAC_FrameWait == 0 {
	//		s.bgm.PlayMemAudio(s.bgm.loop, s.bgm.bgmVolume)
	//	}
	//	s.FLAC_FrameWait--
	//}
}
func (s *System) resetRemapInput() {
	for i := range s.inputRemap {
		s.inputRemap[i] = i
	}
}
func (s *System) loaderReset() {
	s.round, s.wins, s.roundsExisted, s.roundType = 1, [2]int32{}, [2]int32{}, [2]RoundType{}
	s.loader.reset()
}
func (s *System) loadStart() {
	s.loaderReset()
	s.loader.runTread()
}
func (s *System) synchronize() error {
	if s.fileInput != nil {
		s.fileInput.Synchronize()
	} else if s.netInput != nil {
		return s.netInput.Synchronize()
	}
	return nil
}
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
func (s *System) anyButton() bool {
	if s.fileInput != nil {
		return s.fileInput.AnyButton()
	}
	if s.netInput != nil {
		return s.netInput.AnyButton()
	}
	return s.anyHardButton()
}
func (s *System) playerID(id int32) *Char {
	return s.charList.get(id)
}
func (s *System) matchOver() bool {
	return s.wins[0] >= s.matchWins[0] || s.wins[1] >= s.matchWins[1]
}
func (s *System) playerIDExist(id BytecodeValue) BytecodeValue {
	if id.IsSF() {
		return BytecodeSF()
	}
	return BytecodeBool(s.playerID(id.ToI()) != nil)
}
func (s *System) screenHeight() float32 {
	return 240
}
func (s *System) screenWidth() float32 {
	return float32(s.gameWidth)
}
func (s *System) roundEnd() bool {
	return s.intro < -s.lifebar.ro.over_hittime
}
func (s *System) roundWinTime() bool {
	return s.intro < -(s.lifebar.ro.over_hittime+s.lifebar.ro.over_waittime+s.lifebar.ro.over_wintime)
}
func (s *System) roundOver() bool {
	return s.intro < -(s.lifebar.ro.over_hittime + s.lifebar.ro.over_waittime +
		s.lifebar.ro.over_time)
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
func (s *System) appendToConsole(str string) {
	s.consoleText = append(s.consoleText, str)
	if len(s.consoleText) > s.consoleRows {
		s.consoleText = s.consoleText[len(s.consoleText)-s.consoleRows:]
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
func (s *System) clsnHantei(clsn1 []float32, scl1, pos1 [2]float32,
	facing1 float32, clsn2 []float32, scl2, pos2 [2]float32,
	facing2 float32) bool {
	if scl1[0] < 0 {
		facing1 *= -1
		scl1[0] *= -1
	}
	if scl2[0] < 0 {
		facing2 *= -1
		scl2[0] *= -1
	}
	for i1 := 0; i1+3 < len(clsn1); i1 += 4 {
		l1, r1 := clsn1[i1], clsn1[i1+2]+1
		if facing1 < 0 {
			l1, r1 = -r1, -l1
		}
		for i2 := 0; i2+3 < len(clsn2); i2 += 4 {
			l2, r2 := clsn2[i2], clsn2[i2+2]+1
			if facing2 < 0 {
				l2, r2 = -r2, -l2
			}
			if l1*scl1[0]+pos1[0] < r2*scl2[0]+pos2[0] &&
				l2*scl2[0]+pos2[0] < r1*scl1[0]+pos1[0] &&
				clsn1[i1+1]*scl1[1]+pos1[1] < (clsn2[i2+3]+1)*scl2[1]+pos2[1] &&
				clsn2[i2+1]*scl2[1]+pos2[1] < (clsn1[i1+3]+1)*scl1[1]+pos1[1] {
				return true
			}
		}
	}
	return false
}
func (s *System) newCharId() int32 {
	s.nextCharId++
	return s.nextCharId - 1
}
func (s *System) resetGblEffect() {
	s.allPalFX.clear()
	s.bgPalFX.clear()
	s.envShake.clear()
	s.pause, s.pausetime = 0, 0
	s.super, s.supertime = 0, 0
	s.superanim = nil
	s.envcol_time = 0
	s.specialFlag = 0
}
func (s *System) stopAllSound() {
	for _, p := range s.chars {
		for _, c := range p {
			c.soundChannels.SetSize(0)
		}
	}
}
func (s *System) clearAllSound() {
	s.soundChannels.StopAll()
	s.stopAllSound()
}
func (s *System) playerClear(pn int, destroy bool) {
	if len(s.chars[pn]) > 0 {
		p := s.chars[pn][0]
		for _, h := range s.chars[pn][1:] {
			if destroy || h.preserve == 0 || (s.roundResetFlg && h.preserve == s.round) {
				h.destroy()
			}
			h.soundChannels.SetSize(0)
		}
		if destroy {
			p.children = p.children[:0]
		} else {
			for i, ch := range p.children {
				if ch != nil {
					if ch.preserve == 0 || (s.roundResetFlg && ch.preserve == s.round) {
						p.children[i] = nil
					}
				}
			}
		}
		p.targets = p.targets[:0]
		p.soundChannels.SetSize(0)
	}
	s.projs[pn] = s.projs[pn][:0]
	s.explods[pn] = s.explods[pn][:0]
	s.explDrawlist[pn] = s.explDrawlist[pn][:0]
	s.topexplDrawlist[pn] = s.topexplDrawlist[pn][:0]
	s.underexplDrawlist[pn] = s.underexplDrawlist[pn][:0]
}
func (s *System) nextRound() {
	s.resetGblEffect()
	s.lifebar.reset()
	s.finish = FT_NotYet
	s.winTeam = -1
	s.winType = [...]WinType{WT_N, WT_N}
	s.winTrigger = [...]WinType{WT_N, WT_N}
	s.lastHitter = [2]int{-1, -1}
	s.waitdown = s.lifebar.ro.over_hittime*s.lifebar.ro.over_waittime + 900
	s.slowtime = s.lifebar.ro.slow_time
	s.shuttertime = 0
	s.fadeintime = s.lifebar.ro.fadein_time
	s.fadeouttime = s.lifebar.ro.fadeout_time
	s.winskipped = false
	s.intro = s.lifebar.ro.start_waittime + s.lifebar.ro.ctrl_time + 1
	s.time = s.roundTime
	s.nextCharId = s.helperMax
	if (s.tmode[0] == TM_Turns && s.wins[1] == s.numTurns[0]-1) ||
		(s.tmode[0] != TM_Turns && s.wins[1] == s.lifebar.ro.match_wins[0]-1) {
		s.roundType[0] = RT_Deciding
	}
	if (s.tmode[1] == TM_Turns && s.wins[0] == s.numTurns[1]-1) ||
		(s.tmode[1] != TM_Turns && s.wins[0] == s.lifebar.ro.match_wins[1]-1) {
		s.roundType[1] = RT_Deciding
	}
	if s.roundType[0] == RT_Deciding && s.roundType[1] == RT_Deciding {
		s.roundType = [2]RoundType{RT_Final, RT_Final}
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
	}
	s.cam.stageCamera = s.stage.stageCamera
	s.cam.Init()
	s.screenleft = float32(s.stage.screenleft) * s.stage.localscl
	s.screenright = float32(s.stage.screenright) * s.stage.localscl
	if s.stage.resetbg || swap {
		s.stage.reset()
	}
	s.cam.ResetZoomdelay()
	s.cam.Update(1, 0, 0)
	for i, p := range s.chars {
		if len(p) > 0 {
			s.nextCharId = Max(s.nextCharId, p[0].id+1)
			s.playerClear(i, false)
			p[0].posReset()
			p[0].setCtrl(false)
			p[0].clearState()
			p[0].clear2()
			p[0].varRangeSet(0, s.cgi[i].data.intpersistindex-1, 0)
			p[0].fvarRangeSet(0, s.cgi[i].data.floatpersistindex-1, 0)
			for j := range p[0].cmd {
				p[0].cmd[j].BufReset()
			}
			if s.roundsExisted[i&1] == 0 {
				s.cgi[i].sff.palList.ResetRemap()
				if s.cgi[i].sff.header.Ver0 == 1 {
					p[0].remapPal(p[0].getPalfx(),
						[...]int32{1, 1}, [...]int32{1, s.cgi[i].drawpalno})
				}
			}
			s.cgi[i].clearPCTime()
			s.cgi[i].unhittable = 0
		}
	}
	for _, p := range s.chars {
		if len(p) > 0 {
			p[0].selfState(5900, 0, -1, 0, false)
		}
	}
}
func (s *System) debugPaused() bool {
	return s.paused && !s.step && s.oldTickCount < s.tickCount
}
func (s *System) tickFrame() bool {
	return (!s.paused || s.step) && s.oldTickCount < s.tickCount
}
func (s *System) tickNextFrame() bool {
	return int(s.tickCountF+s.nextAddTime) > s.tickCount &&
		(!s.paused || s.step || s.oldTickCount >= s.tickCount)
}
func (s *System) tickInterpola() float32 {
	if s.tickNextFrame() {
		return 1
	}
	return s.tickCountF - s.lastTick + s.nextAddTime
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
}
func (s *System) commandUpdate() {
	for i, p := range s.chars {
		if len(p) > 0 {
			r := p[0]
			if (r.ctrlOver() && !r.sf(CSF_postroundinput)) || r.sf(CSF_noinput) ||
				(r.aiLevel() > 0 && !r.alive()) {
				for j := range r.cmd {
					r.cmd[j].BufReset()
				}
				continue
			}
			act := true
			if s.super > 0 {
				act = r.superMovetime != 0
			} else if s.pause > 0 && r.pauseMovetime == 0 {
				act = false
			}
			if act && !r.sf(CSF_noautoturn) &&
				(r.ss.no == 0 || r.ss.no == 11 || r.ss.no == 20) {
				r.turn()
			}
			for _, c := range p {
				if (c.helperIndex == 0 ||
					c.helperIndex > 0 && &c.cmd[0] != &r.cmd[0]) &&
					c.cmd[0].Input(c.key, int32(c.facing), sys.com[i], c.inputFlag) {
					hp := c.hitPause() && c.gi().constants["input.pauseonhitpause"] != 0
					buftime := Btoi(hp && c.gi().ver[0] != 1)
					if s.super > 0 {
						if !act && s.super <= s.superendcmdbuftime {
							hp = true
						}
					} else if s.pause > 0 {
						if !act && s.pause <= s.pauseendcmdbuftime {
							hp = true
						}
					}
					for j := range c.cmd {
						c.cmd[j].Step(int32(c.facing), c.key < 0, hp, buftime+Btoi(hp))
					}
				}
			}
			if r.key < 0 {
				cc := int32(-1)
				// AI Scaling
				// TODO: Balance AI Scaling
				if r.roundState() == 2 && RandF32(0, sys.com[i]/2+32) > 32 {
					cc = Rand(0, int32(len(r.cmd[r.ss.sb.playerNo].Commands))-1)
				} else {
					cc = -1
				}
				for j := range p {
					if p[j].helperIndex >= 0 {
						p[j].cpucmd = cc
					}
				}
			}
		}
	}
}
func (s *System) charUpdate(cvmin, cvmax,
	highest, lowest, leftest, rightest *float32) {
	s.charList.update(cvmin, cvmax, highest, lowest, leftest, rightest)
	for i, pr := range s.projs {
		for j, p := range pr {
			if p.id >= 0 {
				s.projs[i][j].update(i)
			}
		}
	}
	if s.tickNextFrame() {
		for i, pr := range s.projs {
			for j, p := range pr {
				if p.id >= 0 {
					s.projs[i][j].clsn(i)
				}
			}
		}
		s.charList.getHit()
		for i, pr := range s.projs {
			for j, p := range pr {
				if p.id != IErr {
					s.projs[i][j].tick(i)
				}
			}
		}
		s.charList.tick()
	}
}
func (s *System) posReset() {
	for _, p := range s.chars {
		if len(p) > 0 {
			p[0].posReset()
		}
	}
}
func (s *System) action(x, y, scl *float32) {
	s.sprites = s.sprites[:0]
	s.topSprites = s.topSprites[:0]
	s.bottomSprites = s.bottomSprites[:0]
	s.shadows = s.shadows[:0]
	s.drawc1 = s.drawc1[:0]
	s.drawc2 = s.drawc2[:0]
	s.drawc2sp = s.drawc2sp[:0]
	s.drawc2mtk = s.drawc2mtk[:0]
	s.drawwh = s.drawwh[:0]
	s.clsnText = nil
	var cvmin, cvmax, highest, lowest, leftest, rightest float32 = 0, 0, 0, 0, 0, 0
	leftest, rightest = *x, *x
	
	if s.tickFrame() {
		s.xmin = s.cam.ScreenPos[0] + s.cam.Offset[0] + s.screenleft
		s.xmax = s.cam.ScreenPos[0] + s.cam.Offset[0] +
			float32(s.gameWidth)/s.cam.Scale - s.screenright
		if s.xmin > s.xmax {
			s.xmin = (s.xmin + s.xmax) / 2
			s.xmax = s.xmin
		}
		s.allPalFX.step()
		s.bgPalFX.step()
		s.envShake.next()
		if s.envcol_time > 0 {
			s.envcol_time--
		}
		s.enableZoomstate = false
		s.zoomCameraBound = true
		if s.super > 0 {
			s.super--
		} else if s.pause > 0 {
			s.pause--
		}
		if s.supertime < 0 {
			s.supertime = ^s.supertime
			s.super = s.supertime
		}
		if s.pausetime < 0 {
			s.pausetime = ^s.pausetime
			s.pause = s.pausetime
		}
		// in mugen 1.1 most global assertspecial flags are reset during pause
		// TODO: test if roundnotover should reset (keep intro and noko active)
		if s.super <= 0 && s.pause <= 0 {
			s.specialFlag = 0
		} else {
			s.unsetSF(GSF_assertspecialpause)
		}
		if s.superanim != nil {
			s.superanim.Action()
		}
		s.charList.action(*x, &cvmin, &cvmax,
			&highest, &lowest, &leftest, &rightest)
		s.nomusic = s.sf(GSF_nomusic) && !sys.postMatchFlg
	} else {
		s.charUpdate(&cvmin, &cvmax, &highest, &lowest, &leftest, &rightest)
	}
	s.lifebar.step()
	
	// Action camera
	var newx, newy float32 = *x, *y
	var sclMul float32
	leftest -= *x
	rightest -= *x
	sclMul = s.cam.action(&newx, &newy, leftest, rightest, lowest, highest,
		cvmin, cvmax, s.super > 0 || s.pause > 0)
	
	// Update camera
	introSkip := false
	if s.tickNextFrame() {
		if s.lifebar.ro.cur < 1 && !s.introSkipped {
			if s.shuttertime > 0 ||
				s.anyButton() && !s.sf(GSF_roundnotskip) && s.intro > s.lifebar.ro.ctrl_time {
				s.shuttertime++
				if s.shuttertime == s.lifebar.ro.shutter_time {
					s.fadeintime = 0
					s.resetGblEffect()
					s.intro = s.lifebar.ro.ctrl_time
					for i, p := range s.chars {
						if len(p) > 0 {
							s.playerClear(i, false)
							p[0].selfState(0, -1, -1, 0, false)
						}
					}
					ox := newx
					newx = 0
					leftest = MaxF(float32(Min(s.stage.p[0].startx,
						s.stage.p[1].startx))*s.stage.localscl,
						-(float32(s.gameWidth)/2)/s.cam.BaseScale()+s.screenleft) - ox
					rightest = MinF(float32(Max(s.stage.p[0].startx,
						s.stage.p[1].startx))*s.stage.localscl,
						(float32(s.gameWidth)/2)/s.cam.BaseScale()-s.screenright) - ox
					introSkip = true
					s.introSkipped = true
				}
			}
		} else {
			if s.shuttertime > 0 {
				s.shuttertime--
			}
		}
	}
	if introSkip {
		sclMul = 1 / *scl
	}
	leftest = (leftest - s.screenleft) * s.cam.BaseScale()
	rightest = (rightest + s.screenright) * s.cam.BaseScale()
	*scl = s.cam.ScaleBound(*scl, sclMul)
	tmp := (float32(s.gameWidth) / 2) / *scl
	if AbsF((leftest+rightest)-(newx-*x)*2) >= tmp/2 {
		tmp = MaxF(0, MinF(tmp, MaxF((newx-*x)-leftest, rightest-(newx-*x))))
	}
	*x = s.cam.XBound(*scl, MinF(*x+leftest+tmp, MaxF(*x+rightest-tmp, newx)))
	if !s.cam.ZoomEnable {
		// Pos X の誤差が出ないように精度を落とす
		*x = float32(math.Ceil(float64(*x)*4-0.5) / 4)
	}
	*y = s.cam.YBound(*scl, newy)
	s.cam.Update(*scl, *x, *y)
	
	if s.superanim != nil {
		s.topSprites.add(&SprData{s.superanim, &s.superpmap, s.superpos,
			[...]float32{s.superfacing, 1}, [2]int32{-1}, 5, Rotation{}, [2]float32{},
			false, true, s.cgi[s.superplayer].ver[0] != 1, 1, 1, 0, 0, [4]float32{0, 0, 0, 0}}, 0, 0, 0, 0)
		if s.superanim.loopend {
			s.superanim = nil
		}
	}
	for i, pr := range s.projs {
		for j, p := range pr {
			if p.id >= 0 {
				s.projs[i][j].cueDraw(s.cgi[i].ver[0] != 1, i)
			}
		}
	}
	s.charList.cueDraw()
	explUpdate := func(edl *[len(s.chars)][]int, drop bool) {
		for i, el := range *edl {
			for j := len(el) - 1; j >= 0; j-- {
				if el[j] >= 0 {
					s.explods[i][el[j]].update(s.cgi[i].ver[0] != 1, i)
					if s.explods[i][el[j]].id == IErr {
						if drop {
							el = append(el[:j], el[j+1:]...)
							(*edl)[i] = el
						} else {
							el[j] = -1
						}
					}
				}
			}
		}
	}
	explUpdate(&s.explDrawlist, true)
	explUpdate(&s.topexplDrawlist, false)
	explUpdate(&s.underexplDrawlist, true)
	
	if s.lifebar.ro.act() {
		if s.intro > s.lifebar.ro.ctrl_time {
			s.intro--
			if s.sf(GSF_intro) && s.intro <= s.lifebar.ro.ctrl_time {
				s.intro = s.lifebar.ro.ctrl_time + 1
			}
		} else if s.intro > 0 {
			if s.intro == s.lifebar.ro.ctrl_time {
				s.posReset()
			}
			s.intro--
			if s.intro == 0 {
				for _, p := range s.chars {
					if len(p) > 0 {
						p[0].unsetSCF(SCF_over)
						if !p[0].scf(SCF_standby) || p[0].teamside == -1 {
							if p[0].ss.no == 0 {
								p[0].setCtrl(true)
							} else {
								p[0].selfState(0, -1, -1, 1, false)
							}
						}
					}
				}
			}
		}
		if s.intro == 0 && s.time > 0 && !s.sf(GSF_timerfreeze) &&
			(s.super <= 0 || !s.superpausebg) && (s.pause <= 0 || !s.pausebg) {
			s.time--
		}
		fin := func() bool {
			if s.intro > 0 {
				return false
			}
			ko := [...]bool{true, true}
			for ii := range ko {
				for i := ii; i < MaxSimul*2; i += 2 {
					if len(s.chars[i]) > 0 && s.chars[i][0].teamside != -1 {
						if s.chars[i][0].alive() {
							ko[ii] = false
						} else if (s.tmode[i&1] == TM_Simul && s.loseSimul && s.com[i] == 0) ||
							(s.tmode[i&1] == TM_Tag && s.loseTag) {
							ko[ii] = true
							break
						}
					}
				}
				if ko[ii] {
					i := ii ^ 1
					for ; i < MaxSimul*2; i += 2 {
						if len(s.chars[i]) > 0 && s.chars[i][0].life <
							s.chars[i][0].lifeMax {
							break
						}
					}
					if i >= MaxSimul*2 {
						s.winType[ii^1].SetPerfect()
					}
				}
			}
			ft := s.finish
			if s.time == 0 {
				l := [2]float32{}
				for i := 0; i < 2; i++ {
					for j := i; j < MaxSimul*2; j += 2 {
						if len(s.chars[j]) > 0 {
							if s.tmode[i] == TM_Simul || s.tmode[i] == TM_Tag {
								l[i] += (float32(s.chars[j][0].life) /
									float32(s.numSimul[i])) /
									float32(s.chars[j][0].lifeMax)
							} else {
								l[i] += float32(s.chars[j][0].life) /
									float32(s.chars[j][0].lifeMax)
							}
						}
					}
				}
				if l[0] > l[1] {
					p := true
					for i := 0; i < MaxSimul*2; i += 2 {
						if len(s.chars[i]) > 0 &&
							s.chars[i][0].life < s.chars[i][0].lifeMax {
							p = false
							break
						}
					}
					if p {
						s.winType[0].SetPerfect()
					}
					s.finish = FT_TO
					s.winTeam = 0
				} else if l[0] < l[1] {
					p := true
					for i := 1; i < MaxSimul*2; i += 2 {
						if len(s.chars[i]) > 0 &&
							s.chars[i][0].life < s.chars[i][0].lifeMax {
							p = false
							break
						}
					}
					if p {
						s.winType[1].SetPerfect()
					}
					s.finish = FT_TO
					s.winTeam = 1
				} else {
					s.finish = FT_TODraw
					s.winTeam = -1
				}
				if !(ko[0] || ko[1]) {
					s.winType[0], s.winType[1] = WT_T, WT_T
				}
			}
			if s.intro >= -1 && (ko[0] || ko[1]) {
				if ko[0] && ko[1] {
					s.finish, s.winTeam = FT_DKO, -1
				} else {
					s.finish, s.winTeam = FT_KO, int(Btoi(ko[0]))
				}
			}
			if ft != s.finish {
				for i, p := range sys.chars {
					if len(p) > 0 && ko[^i&1] {
						for _, h := range p {
							for _, tid := range h.targets {
								if t := sys.playerID(tid); t != nil {
									if t.ghv.attr&int32(AT_AH) != 0 {
										s.winTrigger[i&1] = WT_H
									} else if t.ghv.attr&int32(AT_AS) != 0 &&
										s.winTrigger[i&1] == WT_N {
										s.winTrigger[i&1] = WT_S
									}
								}
							}
						}
					}
				}
			}
			return ko[0] || ko[1] || s.time == 0
		}
		if s.roundEnd() || fin() {
			inclWinCount := func() {
				w := [...]bool{!s.chars[1][0].win(), !s.chars[0][0].win()}
				if !w[0] || !w[1] ||
					s.tmode[0] == TM_Turns || s.tmode[1] == TM_Turns ||
					s.draws >= s.lifebar.ro.match_maxdrawgames[0] ||
					s.draws >= s.lifebar.ro.match_maxdrawgames[1] {
					for i, win := range w {
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
			if s.intro == -s.lifebar.ro.over_hittime && s.finish != FT_NotYet {
				inclWinCount()
			}
			// Check if player skipped win pose time
			if s.tickFrame() && s.roundWinTime() && (s.anyButton() && !s.sf(GSF_roundnotskip)) {
				s.intro = Min(s.intro, -(s.lifebar.ro.over_hittime +
					s.lifebar.ro.over_waittime + s.lifebar.ro.over_time -
					s.lifebar.ro.start_waittime))
				s.winskipped = true
			}
			rs4t := -(s.lifebar.ro.over_hittime + s.lifebar.ro.over_waittime)
			if s.winskipped || !s.sf(GSF_roundnotover) ||
				s.intro >= rs4t-s.lifebar.ro.over_wintime {
				s.intro--
				if s.intro == rs4t-1 {
					if s.waitdown > 0 {
						for _, p := range s.chars {
							if len(p) > 0 && !p[0].over() {
								s.intro = rs4t
							}
						}
					}
				}
				if s.waitdown <= 0 || s.intro < rs4t-s.lifebar.ro.over_wintime {
					if s.waitdown >= 0 {
						w := [...]bool{!s.chars[1][0].win(), !s.chars[0][0].win()}
						if !w[0] || !w[1] ||
							s.tmode[0] == TM_Turns || s.tmode[1] == TM_Turns ||
							s.draws >= s.lifebar.ro.match_maxdrawgames[0] ||
							s.draws >= s.lifebar.ro.match_maxdrawgames[1] {
							for i, win := range w {
								if win {
									s.lifebar.wi[i].add(s.winType[i])
									if s.matchOver() && s.wins[i] >= s.matchWins[i] {
										s.lifebar.wc[i].wins += 1
									}
								}
							}
						} else {
							s.draws++
						}
					}
					for _, p := range s.chars {
						if len(p) > 0 {
							//default life recovery, used only if externalized Lua implementaion is disabled
							if len(sys.commonLua) == 0 && s.waitdown >= 0 && s.time > 0 && p[0].win() &&
								p[0].alive() && !s.matchOver() &&
								(s.tmode[0] == TM_Turns || s.tmode[1] == TM_Turns) {
								p[0].life += int32((float32(p[0].lifeMax) *
									float32(s.time) / 60) * s.turnsRecoveryRate)
								if p[0].life > p[0].lifeMax {
									p[0].life = p[0].lifeMax
								}
							}
							if !p[0].scf(SCF_over) && !p[0].hitPause() && p[0].alive() {
								p[0].setSCF(SCF_over)
								if p[0].win() {
									p[0].selfState(180, -1, -1, 1, false)
								} else if p[0].lose() {
									p[0].selfState(170, -1, -1, 1, false)
								} else {
									p[0].selfState(175, -1, -1, 1, false)
								}
							}
						}
					}
					s.waitdown = 0
				}
				s.waitdown--
			}
		} else if s.intro < 0 {
			s.intro = 0
		}
	}
	
	if s.tickNextFrame() {
		spd := s.gameSpeed * s.accel
		if s.postMatchFlg {
			spd = 1
		} else if !s.sf(GSF_nokoslow) && s.time != 0 && s.intro < 0 && s.slowtime > 0 {
			spd *= s.lifebar.ro.slow_speed
			if s.slowtime < s.lifebar.ro.slow_fadetime {
				spd += (float32(1) - s.lifebar.ro.slow_speed) * float32(s.lifebar.ro.slow_fadetime-s.slowtime) / float32(s.lifebar.ro.slow_fadetime)
			}
			s.slowtime--
		}
		s.turbo = spd
	}
	s.tickSound()
	return
}
func (s *System) draw(x, y, scl float32) {
	ecol := uint32(s.envcol[2]&0xff | s.envcol[1]&0xff<<8 |
		s.envcol[0]&0xff<<16)
	s.brightnessOld = s.brightness
	s.brightness = 0x100 >> uint(Btoi(s.super > 0 && s.superdarken))
	bgx, bgy := x/s.stage.localscl, y/s.stage.localscl
	//fade := func(rect [4]int32, color uint32, alpha int32) {
	//	FillRect(rect, color, alpha>>uint(Btoi(s.clsnDraw))+Btoi(s.clsnDraw)*128)
	//}
	if s.envcol_time == 0 {
		c := uint32(0)
		if s.sf(GSF_nobg) {
			if s.allPalFX.enable {
				var rgb [3]int32
				if s.allPalFX.eInvertall {
					rgb = [...]int32{0xff, 0xff, 0xff}
				}
				for i, v := range rgb {
					rgb[i] = Clamp((v+s.allPalFX.eAdd[i])*s.allPalFX.eMul[i]>>8, 0, 0xff)
				}
				c = uint32(rgb[2] | rgb[1]<<8 | rgb[0]<<16)
			}
			FillRect(s.scrrect, c, 0xff)
		} else {
			if s.stage.debugbg {
				FillRect(s.scrrect, 0xff00ff, 0xff)
			} else {
				c = uint32(s.stage.bgclearcolor[2]&0xff | s.stage.bgclearcolor[1]&0xff<<8 | s.stage.bgclearcolor[0]&0xff<<16)
				FillRect(s.scrrect, c, 0xff)
			}
			s.stage.draw(false, bgx, bgy, scl)
		}
		s.bottomSprites.draw(x, y, scl*s.cam.BaseScale())
		if !s.sf(GSF_globalnoshadow) {
			if s.stage.reflection > 0 {
				s.shadows.drawReflection(x, y, scl*s.cam.BaseScale())
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
		s.lifebar.draw(-1)
		s.lifebar.draw(0)
	} else {
		FillRect(s.scrrect, ecol, 255)
	}
	if s.envcol_time == 0 || s.envcol_under {
		s.sprites.draw(x, y, scl*s.cam.BaseScale())
		if s.envcol_time == 0 && !s.sf(GSF_nofg) {
			s.stage.draw(true, bgx, bgy, scl)
		}
	}
	s.lifebar.draw(1)
	s.topSprites.draw(x, y, scl*s.cam.BaseScale())
	s.lifebar.draw(2)
}
func (s *System) drawTop() {
	fade := func(rect [4]int32, color uint32, alpha int32) {
		FillRect(rect, color, alpha>>uint(Btoi(s.clsnDraw))+Btoi(s.clsnDraw)*128)
	}
	fadeout := sys.intro + sys.lifebar.ro.over_hittime + sys.lifebar.ro.over_waittime + sys.lifebar.ro.over_time
	if fadeout == s.fadeouttime-1 && len(sys.commonLua) > 0 && sys.matchOver() && !s.dialogueFlg {
		for _, p := range sys.chars {
			if len(p) > 0 {
				if len(p[0].dialogue) > 0 {
					sys.lifebar.ro.cur = 3
					sys.dialogueFlg = true
					break
				}
			}
		}
	}
	if s.fadeintime > 0 {
		fade(s.scrrect, s.lifebar.ro.fadein_col, 256*s.fadeintime/s.lifebar.ro.fadein_time)
		s.fadeintime--
	} else if s.fadeouttime > 0 && fadeout < s.fadeouttime-1 && !s.dialogueFlg {
		fade(s.scrrect, s.lifebar.ro.fadeout_col, 256*(s.lifebar.ro.fadeout_time-s.fadeouttime)/s.lifebar.ro.fadeout_time)
		s.fadeouttime--
	} else if s.clsnDraw && s.clsnDarken {
		fade(s.scrrect, 0, 0)
	}
	if s.shuttertime > 0 {
		rect := s.scrrect
		rect[3] = s.shuttertime * ((s.scrrect[3] + 1) >> 1) / s.lifebar.ro.shutter_time
		fade(rect, s.lifebar.ro.shutter_col, 255)
		rect[1] = s.scrrect[3] - rect[3]
		fade(rect, s.lifebar.ro.shutter_col, 255)
	}
	s.brightness = s.brightnessOld
	if s.clsnDraw {
		s.clsnSpr.Pal[0] = 0xff0000ff
		s.drawc1.draw(0x3feff)
		s.clsnSpr.Pal[0] = 0xffff0000
		s.drawc2.draw(0x3feff)
		s.clsnSpr.Pal[0] = 0xff00ff00
		s.drawc2sp.draw(0x3feff)
		s.clsnSpr.Pal[0] = 0xff002000
		s.drawc2mtk.draw(0x3feff)
		s.clsnSpr.Pal[0] = 0xff404040
		s.drawwh.draw(0x3feff)
	}
}
func (s *System) drawDebug() {
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
				s.debugFont.yscl/s.heightScale, 0, 1, &s.scrrect,
				s.debugFont.palfx, s.debugFont.frgba)
		}
	}
	if s.debugDraw {
		//Player Info
		x := (320-float32(s.gameWidth))/2 + 1
		y := 240 - float32(s.gameHeight)
		if s.statusLFunc != nil {
			s.debugFont.SetColor(255, 255, 255)
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
		//Console
		y = MaxF(y, 48+240-float32(s.gameHeight))
		s.debugFont.SetColor(255, 255, 255)
		for _, s := range s.consoleText {
			put(&x, &y, s)
		}
		//Data
		pn := s.debugRef[0]
		hn := s.debugRef[1]
		if pn >= len(s.chars) || hn >= len(s.chars[pn]) {
			s.debugRef[0] = 0
			s.debugRef[1] = 0
		}
		s.debugWC = s.chars[s.debugRef[0]][s.debugRef[1]]
		y = float32(s.gameHeight) - float32(s.debugFont.fnt.Size[1])*sys.debugFont.yscl/s.heightScale*
			(float32(len(s.listLFunc))+float32(s.clipboardRows)) - 1*s.heightScale
		for i, f := range s.listLFunc {
			if f != nil {
				if i == 1 {
					s.debugFont.SetColor(199, 199, 219)
				} else if (i == 2 && s.debugWC.animPN != s.debugWC.playerNo) ||
					(i == 3 && s.debugWC.ss.sb.playerNo != s.debugWC.playerNo) {
					s.debugFont.SetColor(255, 255, 127)
				} else {
					s.debugFont.SetColor(255, 255, 255)
				}
				top := s.luaLState.GetTop()
				if s.luaLState.CallByParam(lua.P{Fn: f, NRet: 1,
					Protect: true}) == nil {
					s, ok := s.luaLState.Get(-1).(lua.LString)
					if ok && len(s) > 0 {
						if i == 1 && (sys.debugWC == nil || sys.debugWC.sf(CSF_destroy)) {
							put(&x, &y, string(s)+" disabled")
							break
						}
						put(&x, &y, string(s))
					}
				}
				s.luaLState.SetTop(top)
			}
		}
		//Clipboard
		s.debugFont.SetColor(255, 255, 255)
		for _, s := range s.debugWC.clipboardText {
			put(&x, &y, s)
		}
	}
	//Clsn
	if s.clsnDraw {
		for _, t := range s.clsnText {
			s.debugFont.SetColor(t.r, t.g, t.b)
			s.debugFont.fnt.Print(t.text, t.x, t.y, s.debugFont.xscl/s.widthScale,
				s.debugFont.yscl/s.heightScale, 0, 0, &s.scrrect,
				s.debugFont.palfx, s.debugFont.frgba)
		}
	}
}

// Starts and runs gameplay
// Called to start each match, on hard reset with shift+F4, and
// at the start of any round where a new character tags in for turns mode
func (s *System) fight() (reload bool) {
	// Reset variables
	s.gameTime, s.paused, s.accel = 0, false, 1
	s.aiInput = [len(s.aiInput)]AiInput{}
	// Defer resetting variables on return
	defer func() {
		s.oldNextAddTime = 1
		s.nomusic = false
		s.allPalFX.clear()
		s.allPalFX.enable = false
		for i, p := range s.chars {
			if len(p) > 0 {
				s.playerClear(i, s.matchOver() || (s.tmode[i&1] == TM_Turns && p[0].life <= 0))
			}
		}
		s.wincnt.update()
	}()
	var oldStageVars Stage
	oldStageVars.copyStageVars(s.stage)
	var life, pow, gpow, spow, rlife [len(s.chars)]int32
	var ivar [len(s.chars)][]int32
	var fvar [len(s.chars)][]float32
	var dialogue [len(s.chars)][]string
	var mapArray [len(s.chars)]map[string]float32
	var remapSpr [len(s.chars)]RemapPreset
	// Anonymous function to assign initial character values
	copyVar := func(pn int) {
		life[pn] = s.chars[pn][0].life
		pow[pn] = s.chars[pn][0].power
		gpow[pn] = s.chars[pn][0].guardPoints
		spow[pn] = s.chars[pn][0].dizzyPoints
		rlife[pn] = s.chars[pn][0].redLife
		if len(ivar[pn]) < len(s.chars[pn][0].ivar) {
			ivar[pn] = make([]int32, len(s.chars[pn][0].ivar))
		}
		copy(ivar[pn], s.chars[pn][0].ivar[:])
		if len(fvar[pn]) < len(s.chars[pn][0].fvar) {
			fvar[pn] = make([]float32, len(s.chars[pn][0].fvar))
		}
		copy(fvar[pn], s.chars[pn][0].fvar[:])
		copy(dialogue[pn], s.chars[pn][0].dialogue[:])
		mapArray[pn] = make(map[string]float32)
		for k, v := range s.chars[pn][0].mapArray {
			mapArray[pn][k] = v
		}
		remapSpr[pn] = make(RemapPreset)
		for k, v := range s.chars[pn][0].remapSpr {
			remapSpr[pn][k] = v
		}

		// Reset hitScale.
		s.chars[pn][0].defaultHitScale = newHitScaleArray()
		s.chars[pn][0].activeHitScale = make(map[int32][3]*HitScale)
		s.chars[pn][0].nextHitScale = make(map[int32][3]*HitScale)
	}

	s.debugWC = sys.chars[0][0]
	debugInput := func() {
		select {
		case cl := <-s.commandLine:
			if err := s.luaLState.DoString(cl); err != nil {
				s.errLog.Println(err.Error())
			}
		default:
		}
	}

	// Synchronize with external inputs (netplay, replays, etc)
	if err := s.synchronize(); err != nil {
		s.errLog.Println(err.Error())
		s.esc = true
	}
	if s.netInput != nil {
		defer s.netInput.Stop()
	}
	s.wincnt.init()

	// Initialize super meter values, and max power for teams sharing meter
	var level [len(s.chars)]int32
	for i, p := range s.chars {
		if len(p) > 0 {
			p[0].clear2()
			level[i] = s.wincnt.getLevel(i)
			if s.powerShare[i&1] && p[0].teamside != -1 {
				pmax := Max(s.cgi[i&1].data.power, s.cgi[i].data.power)
				for j := i & 1; j < MaxSimul*2; j += 2 {
					if len(s.chars[j]) > 0 {
						s.chars[j][0].powerMax = pmax
					}
				}
			}
		}
	}
	minlv, maxlv := level[0], level[0]
	for i, lv := range level[1:] {
		if len(s.chars[i+1]) > 0 {
			minlv = Min(minlv, lv)
			maxlv = Max(maxlv, lv)
		}
	}
	if minlv > 0 {
		for i := range level {
			level[i] -= minlv
		}
	} else if maxlv < 0 {
		for i := range level {
			level[i] -= maxlv
		}
	}

	// Initialize each character
	lvmul := math.Pow(2, 1.0/12)
	for i, p := range s.chars {
		if len(p) > 0 {
			// Get max life, and adjust based on team mode
			var lm float32
			if p[0].ocd().lifeMax != -1 {
				lm = float32(p[0].ocd().lifeMax) * p[0].ocd().lifeRatio * s.lifeMul
			} else {
				lm = float32(p[0].gi().data.life) * p[0].ocd().lifeRatio * s.lifeMul
			}
			if p[0].teamside != -1 {
				switch s.tmode[i&1] {
				case TM_Single:
					switch s.tmode[(i+1)&1] {
					case TM_Simul, TM_Tag:
						lm *= s.team1VS2Life
					case TM_Turns:
						if s.numTurns[(i+1)&1] < s.matchWins[(i+1)&1] && s.lifeShare[i&1] {
							lm = lm * float32(s.numTurns[(i+1)&1]) /
								float32(s.matchWins[(i+1)&1])
						}
					}
				case TM_Simul, TM_Tag:
					switch s.tmode[(i+1)&1] {
					case TM_Simul, TM_Tag:
						if s.numSimul[(i+1)&1] < s.numSimul[i&1] && s.lifeShare[i&1] {
							lm = lm * float32(s.numSimul[(i+1)&1]) / float32(s.numSimul[i&1])
						}
					case TM_Turns:
						if s.numTurns[(i+1)&1] < s.numSimul[i&1]*s.matchWins[(i+1)&1] && s.lifeShare[i&1] {
							lm = lm * float32(s.numTurns[(i+1)&1]) /
								float32(s.numSimul[i&1]*s.matchWins[(i+1)&1])
						}
					default:
						if s.lifeShare[i&1] {
							lm /= float32(s.numSimul[i&1])
						}
					}
				case TM_Turns:
					switch s.tmode[(i+1)&1] {
					case TM_Single:
						if s.matchWins[i&1] < s.numTurns[i&1] && s.lifeShare[i&1] {
							lm = lm * float32(s.matchWins[i&1]) / float32(s.numTurns[i&1])
						}
					case TM_Simul, TM_Tag:
						if s.numSimul[(i+1)&1]*s.matchWins[i&1] < s.numTurns[i&1] && s.lifeShare[i&1] {
							lm = lm * s.team1VS2Life *
								float32(s.numSimul[(i+1)&1]*s.matchWins[i&1]) /
								float32(s.numTurns[i&1])
						}
					case TM_Turns:
						if s.numTurns[(i+1)&1] < s.numTurns[i&1] && s.lifeShare[i&1] {
							lm = lm * float32(s.numTurns[(i+1)&1]) / float32(s.numTurns[i&1])
						}
					}
				}
			}
			foo := math.Pow(lvmul, float64(-level[i]))
			p[0].lifeMax = Max(1, int32(math.Floor(foo*float64(lm))))

			if p[0].roundsExisted() > 0 {
				/* If character already existed for a round, presumably because of turns mode, just update life */
				p[0].life = Min(p[0].lifeMax, int32(math.Ceil(foo*float64(p[0].life))))
			} else if s.round == 1 || s.tmode[i&1] == TM_Turns {
				/* If round 1 or a new character in turns mode, initialize values */
				if p[0].ocd().life != -1 {
					p[0].life = p[0].ocd().life
				} else {
					p[0].life = p[0].lifeMax
				}
				if s.round == 1 {
					if s.maxPowerMode {
						p[0].power = p[0].powerMax
					} else if p[0].ocd().power != -1 {
						p[0].power = p[0].ocd().power
					} else {
						p[0].power = 0
					}
				}
				p[0].dialogue = []string{}
				p[0].mapArray = make(map[string]float32)
				for k, v := range p[0].mapDefault {
					p[0].mapArray[k] = v
				}
				p[0].remapSpr = make(RemapPreset)

				// Reset hitScale
				p[0].defaultHitScale = newHitScaleArray()
				p[0].activeHitScale = make(map[int32][3]*HitScale)
				p[0].nextHitScale = make(map[int32][3]*HitScale)
			}

			if p[0].ocd().guardPoints != -1 {
				p[0].guardPoints = p[0].ocd().guardPoints
			} else {
				p[0].guardPoints = p[0].guardPointsMax
			}
			if p[0].ocd().dizzyPoints != -1 {
				p[0].dizzyPoints = p[0].ocd().dizzyPoints
			} else {
				p[0].dizzyPoints = p[0].dizzyPointsMax
			}
			p[0].redLife = 0
			copyVar(i)
		}
	}

	//default bgm playback, used only in Quick VS or if externalized Lua implementaion is disabled
	if s.round == 1 && (s.gameMode == "" || len(sys.commonLua) == 0) {
		s.bgm.Open(s.stage.bgmusic, 1, 100, 0, 0)
	}

	oldWins, oldDraws := s.wins, s.draws
	oldTeamLeader := s.teamLeader
	var x, y, scl float32
	// Anonymous function to reset values, called at the start of each round
	reset := func() {
		s.wins, s.draws = oldWins, oldDraws
		s.teamLeader = oldTeamLeader
		for i, p := range s.chars {
			if len(p) > 0 {
				p[0].life = life[i]
				p[0].power = pow[i]
				p[0].guardPoints = gpow[i]
				p[0].dizzyPoints = spow[i]
				p[0].redLife = rlife[i]
				copy(p[0].ivar[:], ivar[i])
				copy(p[0].fvar[:], fvar[i])
				copy(p[0].dialogue[:], dialogue[i])
				p[0].mapArray = make(map[string]float32)
				for k, v := range mapArray[i] {
					p[0].mapArray[k] = v
				}
				p[0].remapSpr = make(RemapPreset)
				for k, v := range remapSpr[i] {
					p[0].remapSpr[k] = v
				}

				// Reset hitScale
				p[0].defaultHitScale = newHitScaleArray()
				p[0].activeHitScale = make(map[int32][3]*HitScale)
				p[0].nextHitScale = make(map[int32][3]*HitScale)
			}
		}
		s.stage.copyStageVars(&oldStageVars)
		s.resetFrameTime()
		s.nextRound()
		s.roundResetFlg, s.introSkipped = false, false
		s.reloadFlg, s.reloadStageFlg, s.reloadLifebarFlg = false, false, false
		x, y = 0, 0
		scl = s.cam.startzoom
		s.cam.Update(scl, x, y)
	}
	reset()

	// Loop until end of match
	fin := false
	for !s.endMatch {
		s.step = false
		for _, v := range s.shortcutScripts {
			if v.Activate {
				if err := s.luaLState.DoString(v.Script); err != nil {
					s.errLog.Println(err.Error())
				}
			}
		}

		// If next round
		if s.roundOver() && !fin {
			s.round++
			for i := range s.roundsExisted {
				s.roundsExisted[i]++
			}
			s.clearAllSound()
			tbl_roundNo := s.luaLState.NewTable()
			for _, p := range s.chars {
				if len(p) > 0 && p[0].teamside != -1 {
					tmp := s.luaLState.NewTable()
					tmp.RawSetString("name", lua.LString(p[0].name))
					tmp.RawSetString("id", lua.LNumber(p[0].id))
					tmp.RawSetString("memberNo", lua.LNumber(p[0].memberNo))
					tmp.RawSetString("selectNo", lua.LNumber(p[0].selectNo))
					tmp.RawSetString("teamside", lua.LNumber(p[0].teamside))
					tmp.RawSetString("life", lua.LNumber(p[0].life))
					tmp.RawSetString("lifeMax", lua.LNumber(p[0].lifeMax))
					tmp.RawSetString("winquote", lua.LNumber(p[0].winquote))
					tmp.RawSetString("aiLevel", lua.LNumber(p[0].aiLevel()))
					tmp.RawSetString("palno", lua.LNumber(p[0].palno()))
					tmp.RawSetString("ratiolevel", lua.LNumber(p[0].ocd().ratioLevel))
					tmp.RawSetString("win", lua.LBool(p[0].win()))
					tmp.RawSetString("winKO", lua.LBool(p[0].winKO()))
					tmp.RawSetString("winTime", lua.LBool(p[0].winTime()))
					tmp.RawSetString("winPerfect", lua.LBool(p[0].winPerfect()))
					tmp.RawSetString("winSpecial", lua.LBool(p[0].winType(WT_S)))
					tmp.RawSetString("winHyper", lua.LBool(p[0].winType(WT_H)))
					tmp.RawSetString("drawgame", lua.LBool(p[0].drawgame()))
					tmp.RawSetString("ko", lua.LBool(p[0].scf(SCF_ko)))
					tmp.RawSetString("ko_round_middle", lua.LBool(p[0].scf(SCF_ko_round_middle)))
					tmp.RawSetString("firstAttack", lua.LBool(p[0].firstAttack))
					tbl_roundNo.RawSetInt(p[0].playerNo+1, tmp)
					p[0].firstAttack = false
				}
			}
			s.matchData.RawSetInt(int(s.round-1), tbl_roundNo)
			s.scoreRounds = append(s.scoreRounds, [2]float32{s.lifebar.sc[0].scorePoints, s.lifebar.sc[1].scorePoints})
			oldTeamLeader = s.teamLeader

			if !s.matchOver() && (s.tmode[0] != TM_Turns || s.chars[0][0].win()) &&
				(s.tmode[1] != TM_Turns || s.chars[1][0].win()) {
				/* Prepare for the next round */
				for i, p := range s.chars {
					if len(p) > 0 {
						if s.tmode[i&1] != TM_Turns || !p[0].win() {
							p[0].life = p[0].lifeMax
						} else if p[0].life <= 0 {
							p[0].life = 1
						}
						p[0].redLife = 0
						copyVar(i)
					}
				}
				oldWins, oldDraws = s.wins, s.draws
				oldStageVars.copyStageVars(s.stage)
				reset()
			} else {
				/* End match, or prepare for a new character in turns mode */
				for i, tm := range s.tmode {
					if s.chars[i][0].win() || !s.chars[i][0].lose() && tm != TM_Turns {
						for j := i; j < len(s.chars); j += 2 {
							if len(s.chars[j]) > 0 {
								if s.chars[j][0].win() {
									s.chars[j][0].life = Max(1, int32(math.Ceil(math.Pow(lvmul,
										float64(level[i]))*float64(s.chars[j][0].life))))
								} else {
									s.chars[j][0].life = Max(1, s.cgi[j].data.life)
								}
							}
						}
						//} else {
						//	s.chars[i][0].life = 0
					}
				}
				// If match isn't over, presumably this is turns mode,
				// so break to restart fight for the next character
				if !s.matchOver() {
					break
				}

				// Otherwise match is over
				s.postMatchFlg = true
				fin = true
			}
		}

		// If frame is ready to tick and not paused
		if s.tickFrame() && (s.super <= 0 || !s.superpausebg) &&
			(s.pause <= 0 || !s.pausebg) {
			// Update stage
			s.stage.action()
		}

		// Update game state
		s.action(&x, &y, &scl)

		// F4 pressed to restart round
		if s.roundResetFlg && !s.postMatchFlg {
			reset()
		}
		// Shift+F4 pressed to restart match
		if s.reloadFlg {
			return true
		}

		debugInput()
		if !s.addFrameTime(s.turbo) {
			if !s.eventUpdate() {
				return false
			}
			continue
		}
		// Render frame
		if !s.frameSkip {
			dx, dy, dscl := x, y, scl
			if s.enableZoomstate {
				if !s.debugPaused() {
					s.zoomPosXLag += ((s.zoomPos[0] - s.zoomPosXLag) * (1 - s.zoomlag))
					s.zoomPosYLag += ((s.zoomPos[1] - s.zoomPosYLag) * (1 - s.zoomlag))
					s.drawScale = s.drawScale / (s.drawScale + (s.zoomScale*scl-s.drawScale)*s.zoomlag) * s.zoomScale * scl
				}
				if s.zoomCameraBound {
					dscl = MaxF(s.cam.MinScale, s.drawScale/s.cam.BaseScale())
					dx = s.cam.XBound(dscl, x+s.zoomPosXLag/scl)
				} else {
					dscl = s.drawScale / s.cam.BaseScale()
					dx = x + s.zoomPosXLag/scl
				}
				dy = y + s.zoomPosYLag
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
		//Lua code executed before drawing fade, clsns and debug
		for _, str := range s.commonLua {
			if err := s.luaLState.DoString(str); err != nil {
				s.luaLState.RaiseError(err.Error())
			}
		}
		// Render debug elements
		if !s.frameSkip {
			s.drawTop()
			s.drawDebug()
		}

		// Break if finished
		if fin && (!s.postMatchFlg || len(sys.commonLua) == 0) {
			break
		}

		// Update system; break if update returns false (game ended)
		if !s.update() {
			break
		}

		// If end match selected from menu/end of attract mode match/etc
		if s.endMatch {
			s.esc = true
		} else if s.esc {
			s.endMatch = s.netInput != nil || len(sys.commonLua) == 0
		}
	}

	return false
}

type wincntMap map[string][]int32

func (wm *wincntMap) init() {
	if sys.autolevel {
		b, err := ioutil.ReadFile(sys.wincntFileName)
		if err != nil {
			return
		}
		str := string(b)
		if len(str) < 3 {
			return
		}
		if str[:3] == "\ufeff" {
			str = str[3:]
		}
		toint := func(strAry []string) (intAry []int32) {
			for _, s := range strAry {
				i, _ := strconv.ParseInt(s, 10, 32)
				intAry = append(intAry, int32(i))
			}
			return
		}
		for _, l := range strings.Split(str, "\n") {
			tmp := strings.Split(l, ",")
			if len(tmp) >= 2 {
				item := toint(strings.Split(strings.TrimSpace(tmp[1]), " "))
				if len(item) < MaxPalNo {
					item = append(item, make([]int32, MaxPalNo-len(item))...)
				}
				(*wm)[tmp[0]] = item
			}
		}
	}
}
func (wm *wincntMap) update() {
	winPoint := func(i int) int32 {
		if sys.tmode[(i+1)&1] == TM_Simul || sys.tmode[(i+1)&1] == TM_Tag {
			if sys.tmode[i&1] != TM_Simul && sys.tmode[i&1] != TM_Tag {
				return sys.numSimul[(i+1)&1]
			} else if sys.numSimul[(i+1)&1] > sys.numSimul[i&1] {
				return sys.numSimul[(i+1)&1] / sys.numSimul[i&1]
			}
		}
		return 1
	}
	win := func(i int) {
		item := wm.getItem(sys.cgi[i].def)
		item[sys.cgi[i].palno-1] += winPoint(i)
		wm.setItem(i, item)
	}
	lose := func(i int) {
		item := wm.getItem(sys.cgi[i].def)
		item[sys.cgi[i].palno-1] -= winPoint(i)
		wm.setItem(i, item)
	}
	if sys.autolevel && sys.matchOver() {
		for i, p := range sys.chars {
			if len(p) > 0 {
				if p[0].win() {
					win(i)
				} else if p[0].lose() {
					lose(i)
				}
			}
		}
		var str string
		for k, v := range *wm {
			str += k + ","
			for _, w := range v {
				str += fmt.Sprintf(" %v", w)
			}
			str += "\r\n"
		}
		f, err := os.Create(sys.wincntFileName)
		if err == nil {
			f.Write([]byte(str))
			chk(f.Close())
		}
	}
}
func (wm wincntMap) getItem(def string) []int32 {
	lv := wm[def]
	if len(lv) < MaxPalNo {
		lv = append(lv, make([]int32, MaxPalNo-len(lv))...)
	}
	return lv
}
func (wm wincntMap) setItem(pn int, item []int32) {
	var ave, palcnt int32 = 0, 0
	for i, v := range item {
		if sys.cgi[pn].palSelectable[i] {
			ave += v
			palcnt++
		}
	}
	ave /= palcnt
	for i := range item {
		if !sys.cgi[pn].palSelectable[i] {
			item[i] = ave
		}
	}
	wm[sys.cgi[pn].def] = item
}
func (wm wincntMap) getLevel(p int) int32 {
	return wm.getItem(sys.cgi[p].def)[sys.cgi[p].palno-1]
}

type SelectChar struct {
	def            string
	name           string
	lifebarname    string
	author         string
	sound          string
	intro          string
	ending         string
	arcadepath     string
	ratiopath      string
	movelist       string
	pal            []int32
	pal_defaults   []int32
	pal_keymap     []int32
	localcoord     int32
	portrait_scale float32
	cns_scale      [2]float32
	anims          PreloadedAnims
	sff            *Sff
	fnt            [10]*Fnt
}

func newSelectChar() *SelectChar {
	return &SelectChar{
		localcoord:     320,
		portrait_scale: 1,
		cns_scale:      [...]float32{1, 1},
		anims:          NewPreloadedAnims(),
	}
}

type SelectStage struct {
	def             string
	name            string
	attachedchardef string
	stagebgm        IniSection
	portrait_scale  float32
	anims           PreloadedAnims
	sff             *Sff
}

func newSelectStage() *SelectStage {
	return &SelectStage{portrait_scale: 1, anims: NewPreloadedAnims()}
}

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

type Select struct {
	charlist           []SelectChar
	stagelist          []SelectStage
	selected           [2][][2]int
	selectedStageNo    int
	charAnimPreload    []int32
	stageAnimPreload   []int32
	charSpritePreload  map[[2]int16]bool
	stageSpritePreload map[[2]int16]bool
	cdefOverwrite      map[int]string
	sdefOverwrite      string
	ocd                [3][]OverrideCharData
}

func newSelect() *Select {
	return &Select{selectedStageNo: -1,
		charSpritePreload: map[[2]int16]bool{[...]int16{9000, 0}: true,
			[...]int16{9000, 1}: true}, stageSpritePreload: make(map[[2]int16]bool),
		cdefOverwrite: make(map[int]string)}
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
func (s *Select) addChar(def string) {
	var tstr string
	tnow := time.Now()
	defer func() {
		sys.loadTime(tnow, tstr, false, false)
	}()
	s.charlist = append(s.charlist, *newSelectChar())
	sc := &s.charlist[len(s.charlist)-1]
	def = strings.Replace(strings.TrimSpace(strings.Split(def, ",")[0]),
		"\\", "/", -1)
	tstr = fmt.Sprintf("Char added: %v", def)
	if strings.ToLower(def) == "dummyslot" {
		sc.name = "dummyslot"
		return
	}
	if strings.ToLower(def) == "randomselect" {
		sc.def, sc.name = "randomselect", "Random"
		return
	}
	idx := strings.Index(def, "/")
	if len(def) >= 4 && strings.ToLower(def[len(def)-4:]) == ".def" {
		if idx < 0 {
			sc.name = "dummyslot"
			return
		}
	} else if idx < 0 {
		def += "/" + def + ".def"
	} else {
		def += ".def"
	}
	if chk := FileExist(def); len(chk) != 0 {
		def = chk
	} else {
		if strings.ToLower(def[0:6]) != "chars/" && strings.ToLower(def[1:3]) != ":/" && (def[0] != '/' || idx > 0 && !strings.Contains(def[:idx], ":")) {
			def = "chars/" + def
		}
		if def = FileExist(def); len(def) == 0 {
			sc.name = "dummyslot"
			return
		}
	}
	str, err := LoadText(def)
	if err != nil {
		sc.name = "dummyslot"
		return
	}
	sc.def = def
	lines, i, info, files, keymap, arcade := SplitAndTrim(str, "\n"), 0, true, true, true, true
	var cns, sprite, anim, movelist string
	var fnt [10][2]string
	for i < len(lines) {
		is, name, subname := ReadIniSection(lines, &i)
		switch name {
		case "info":
			if info {
				info = false
				var ok bool
				if sc.name, ok, _ = is.getText("displayname"); !ok {
					sc.name, _, _ = is.getText("name")
				}
				if sc.lifebarname, ok, _ = is.getText("lifebarname"); !ok {
					sc.lifebarname = sc.name
				}
				sc.author, _, _ = is.getText("author")
				sc.pal_defaults = is.readI32CsvForStage("pal.defaults")
				is.ReadI32("localcoord", &sc.localcoord)
				if ok = is.ReadF32("portraitscale", &sc.portrait_scale); !ok {
					sc.portrait_scale = 320 / float32(sc.localcoord)
				}
			}
		case "files":
			if files {
				files = false
				cns = is["cns"]
				sprite = is["sprite"]
				anim = is["anim"]
				sc.sound = is["sound"]
				for i := 1; i <= MaxPalNo; i++ {
					if is[fmt.Sprintf("pal%v", i)] != "" {
						sc.pal = append(sc.pal, int32(i))
					}
				}
				movelist = is["movelist"]
				for i := range fnt {
					fnt[i][0] = is[fmt.Sprintf("font%v", i)]
					fnt[i][1] = is[fmt.Sprintf("fnt_height%v", i)]
				}
			}
		case "palette ":
			if keymap &&
				len(subname) >= 6 && strings.ToLower(subname[:6]) == "keymap" {
				keymap = false
				for _, v := range [12]string{"a", "b", "c", "x", "y", "z",
					"a2", "b2", "c2", "x2", "y2", "z2"} {
					var i32 int32
					if is.ReadI32(v, &i32) {
						sc.pal_keymap = append(sc.pal_keymap, i32)
					}
				}
			}
		case "arcade":
			if arcade {
				arcade = false
				sc.intro, _, _ = is.getText("intro.storyboard")
				sc.ending, _, _ = is.getText("ending.storyboard")
				sc.arcadepath, _, _ = is.getText("arcadepath")
				sc.ratiopath, _, _ = is.getText("ratiopath")
			}
		}
	}
	listSpr := make(map[[2]int16]bool)
	for k := range s.charSpritePreload {
		listSpr[[...]int16{k[0], k[1]}] = true
	}
	sff := newSff()
	//read size values
	LoadFile(&cns, []string{def, "", "data/"}, func(filename string) error {
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
					sc.cns_scale[0] = 320 / float32(sc.localcoord)
				}
				if ok := is.ReadF32("yscale", &sc.cns_scale[1]); !ok {
					sc.cns_scale[1] = 320 / float32(sc.localcoord)
				}
				return nil
			}
		}
		return nil
	})
	//preload animations
	LoadFile(&anim, []string{def, "", "data/"}, func(filename string) error {
		str, err := LoadText(filename)
		if err != nil {
			return err
		}
		lines, i := SplitAndTrim(str, "\n"), 0
		at := ReadAnimationTable(sff, lines, &i)
		for _, v := range s.charAnimPreload {
			if anim := at.get(v); anim != nil {
				sc.anims.addAnim(anim, v)
				for _, fr := range anim.frames {
					listSpr[[...]int16{fr.Group, fr.Number}] = true
				}
			}
		}
		return nil
	})
	//preload portion of sff file
	fp := fmt.Sprintf("%v_preload.sff", strings.TrimSuffix(def, filepath.Ext(def)))
	if fp = FileExist(fp); len(fp) == 0 {
		fp = sprite
	}
	LoadFile(&fp, []string{def, "", "data/"}, func(file string) error {
		var selPal []int32
		var err error
		sc.sff, selPal, err = preloadSff(file, true, listSpr)
		if err != nil {
			panic(fmt.Errorf("failed to load %v: %v\nerror preloading %v", file, err, def))
		}
		sc.anims.updateSff(sc.sff)
		for k := range s.charSpritePreload {
			sc.anims.addSprite(sc.sff, k[0], k[1])
		}
		if len(sc.pal) == 0 {
			sc.pal = selPal
		}
		return nil
	})
	//read movelist
	if len(movelist) > 0 {
		LoadFile(&movelist, []string{def, "", "data/"}, func(file string) error {
			sc.movelist, _ = LoadText(file)
			return nil
		})
	}
	//preload fonts
	for i, f := range fnt {
		if len(f[0]) > 0 {
			LoadFile(&f[0], []string{def, sys.motifDir, "", "data/", "font/"}, func(filename string) error {
				var err error
				var height int32 = -1
				if len(f[1]) > 0 {
					height = Atoi(f[1])
				}
				if sc.fnt[i], err = loadFnt(filename, height); err != nil {
					sys.errLog.Printf("failed to load %v (char font): %v", filename, err)
				}
				return nil
			})
		}
	}
}
func (s *Select) AddStage(def string) error {
	var tstr string
	tnow := time.Now()
	defer func() {
		sys.loadTime(tnow, tstr, false, false)
	}()
	var lines []string
	if err := LoadFile(&def, []string{"", "data/"}, func(file string) error {
		str, err := LoadText(file)
		if err != nil {
			return err
		}
		lines = SplitAndTrim(str, "\n")
		return nil
	}); err != nil {
		sys.errLog.Printf("Failed to add stage, file not found: %v\n", def)
		return err
	}
	tstr = fmt.Sprintf("Stage added: %v", def)
	i, info, music, bgdef, stageinfo := 0, true, true, true, true
	var spr string
	s.stagelist = append(s.stagelist, *newSelectStage())
	ss := &s.stagelist[len(s.stagelist)-1]
	ss.def = def
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
				if err := is.LoadFile("attachedchar", []string{def, "", sys.motifDir, "data/"}, func(filename string) error {
					ss.attachedchardef = filename
					return nil
				}); err != nil {
					return nil
				}
			}
		case "music":
			if music {
				music = false
				ss.stagebgm = is
			}
		case "bgdef":
			if bgdef {
				bgdef = false
				spr = is["spr"]
			}
		case "stageinfo":
			if stageinfo {
				stageinfo = false
				if ok := is.ReadF32("portraitscale", &ss.portrait_scale); !ok {
					localcoord := float32(320)
					is.ReadF32("localcoord", &localcoord)
					ss.portrait_scale = 320 / localcoord
				}
			}
		}
	}
	if len(s.stageSpritePreload) > 0 || len(s.stageAnimPreload) > 0 {
		listSpr := make(map[[2]int16]bool)
		for k := range s.stageSpritePreload {
			listSpr[[...]int16{k[0], k[1]}] = true
		}
		sff := newSff()
		//preload animations
		i = 0
		at := ReadAnimationTable(sff, lines, &i)
		for _, v := range s.stageAnimPreload {
			if anim := at.get(v); anim != nil {
				ss.anims.addAnim(anim, v)
				for _, fr := range anim.frames {
					listSpr[[...]int16{fr.Group, fr.Number}] = true
				}
			}
		}
		//preload portion of sff file
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
		pl = int(Rand(1, MaxPalNo))
	}
	sys.loadMutex.Lock()
	s.selected[tn] = append(s.selected[tn], [...]int{n, pl})
	s.ocd[tn] = append(s.ocd[tn], *newOverrideCharData())
	sys.loadMutex.Unlock()
	return true
}
func (s *Select) ClearSelected() {
	sys.loadMutex.Lock()
	s.selected = [2][][2]int{}
	s.ocd = [3][]OverrideCharData{}
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
	if sys.roundsExisted[pn&1] > 0 {
		return 1
	}
	sys.loadMutex.Lock()
	result := -1
	nsel := len(sys.sel.selected[pn&1])
	if sys.tmode[pn&1] == TM_Simul || sys.tmode[pn&1] == TM_Tag {
		if pn>>1 >= int(sys.numSimul[pn&1]) {
			sys.cgi[pn].states = nil
			sys.chars[pn] = nil
			result = 1
		}
	} else if pn >= 2 {
		result = 0
	}
	if sys.tmode[pn&1] == TM_Turns && nsel < int(sys.numTurns[pn&1]) {
		result = 0
	}
	memberNo := pn >> 1
	if sys.tmode[pn&1] == TM_Turns {
		memberNo = int(sys.wins[^pn&1])
	}
	if result < 0 && nsel <= memberNo {
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
	var tstr string
	tnow := time.Now()
	defer func() {
		sys.loadTime(tnow, tstr, false, true)
	}()
	var cdef string
	var cdefOWnumber int
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
	var p *Char
	if len(sys.chars[pn]) > 0 && cdef == sys.cgi[pn].def {
		p = sys.chars[pn][0]
		p.key = pn
		if sys.com[pn] != 0 {
			p.key ^= -1
		}
		p.clearCachedData()
	} else {
		p = newChar(pn, 0)
		if sys.cgi[pn].sff != nil {
			sys.cgi[pn].sff.sprites = nil
		}
		sys.cgi[pn].sff = nil
		if len(sys.chars[pn]) > 0 {
			p.power = sys.chars[pn][0].power
			p.guardPoints = sys.chars[pn][0].guardPoints
			p.dizzyPoints = sys.chars[pn][0].dizzyPoints
		}
	}
	p.memberNo = memberNo
	p.selectNo = sys.sel.selected[pn&1][memberNo][0]
	p.teamside = p.playerNo & 1
	if !p.ocd().existed {
		p.varRangeSet(0, int32(NumVar)-1, 0)
		p.fvarRangeSet(0, int32(NumFvar)-1, 0)
		p.ocd().existed = true
	}
	sys.chars[pn] = make([]*Char, 1)
	sys.chars[pn][0] = p
	if sys.cgi[pn].sff == nil {
		if l.err = p.load(cdef); l.err != nil {
			sys.chars[pn] = nil
			tstr = fmt.Sprintf("WARNING: Failed to load new char: %v", cdef)
			return -1
		}
		if sys.cgi[pn].states, l.err =
			newCompiler().Compile(p.playerNo, cdef, p.gi().constants); l.err != nil {
			sys.chars[pn] = nil
			tstr = fmt.Sprintf("WARNING: Failed to compile new char states: %v", cdef)
			return -1
		}
		tstr = fmt.Sprintf("New char loaded: %v", cdef)
	} else {
		tstr = fmt.Sprintf("Cached char loaded: %v", cdef)
	}
	sys.cgi[pn].palno = pal //sys.cgi[pn].palkeymap[pal-1] + 1
	if pn < len(sys.lifebar.fa[sys.tmode[pn&1]]) &&
		sys.tmode[pn&1] == TM_Turns && sys.round == 1 {
		fa := sys.lifebar.fa[sys.tmode[pn&1]][pn]
		fa.numko, fa.teammate_face, fa.teammate_scale = 0, make([]*Sprite, nsel), make([]float32, nsel)
		sys.lifebar.nm[sys.tmode[pn&1]][pn].numko = 0
		for i, ci := range idx {
			fa.teammate_scale[i] = sys.sel.charlist[ci].portrait_scale
			fa.teammate_face[i] = sys.sel.charlist[ci].sff.GetSprite(int16(fa.teammate_face_spr[0]),
				int16(fa.teammate_face_spr[1]))
		}
	}
	return 1
}

func (l *Loader) loadAttachedChar(pn int) int {
	if sys.round != 1 {
		return 1
	}
	atcpn := pn - MaxSimul*2
	var tstr string
	tnow := time.Now()
	defer func() {
		sys.loadTime(tnow, tstr, false, true)
	}()
	sys.sel.ocd[2] = append(sys.sel.ocd[2], *newOverrideCharData())
	cdef := sys.stageList[0].attachedchardef[atcpn]
	var p *Char
	if len(sys.chars[pn]) > 0 && cdef == sys.cgi[pn].def {
		p = sys.chars[pn][0]
		//p.key = -pn
		p.clearCachedData()
	} else {
		p = newChar(pn, 0)
		sys.cgi[pn].sff = nil
		if len(sys.chars[pn]) > 0 {
			p.power = sys.chars[pn][0].power
			p.guardPoints = sys.chars[pn][0].guardPoints
			p.dizzyPoints = sys.chars[pn][0].dizzyPoints
		}
	}
	p.memberNo = -atcpn
	p.selectNo = -atcpn
	p.teamside = -1
	if !p.ocd().existed {
		p.varRangeSet(0, int32(NumVar)-1, 0)
		p.fvarRangeSet(0, int32(NumFvar)-1, 0)
		p.ocd().existed = true
	}
	sys.com[pn] = 8
	sys.chars[pn] = make([]*Char, 1)
	sys.chars[pn][0] = p
	if sys.cgi[pn].sff == nil {
		if l.err = p.load(cdef); l.err != nil {
			sys.chars[pn] = nil
			tstr = fmt.Sprintf("WARNING: Failed to load new attachedchar: %v", cdef)
			return -1
		}
		if sys.cgi[pn].states, l.err =
			newCompiler().Compile(p.playerNo, cdef, p.gi().constants); l.err != nil {
			sys.chars[pn] = nil
			tstr = fmt.Sprintf("WARNING: Failed to compile new attachedchar states: %v", cdef)
			return -1
		}
		tstr = fmt.Sprintf("New attachedchar loaded: %v", cdef)
	} else {
		tstr = fmt.Sprintf("Cached attachedchar loaded: %v", cdef)
	}
	sys.cgi[pn].palno = 1
	return 1
}

func (l *Loader) loadStage() bool {
	if sys.round == 1 {
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
			return true
		}
		sys.stageList = make(map[int32]*Stage)
		sys.stageLoop = false
		sys.stageList[0], l.err = loadStage(def, true)
		sys.stage = sys.stageList[0]
	}
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
		if !stageDone && sys.sel.selectedStageNo >= 0 {
			if !l.loadStage() {
				l.state = LS_Error
				return
			}
			stageDone = true
		}
		for i, b := range charDone {
			if !b {
				result := -1
				if i < len(sys.chars)-MaxAttachedChar ||
					len(sys.stageList[0].attachedchardef) <= i-MaxSimul*2 {
					result = l.loadChar(i)
				} else {
					result = l.loadAttachedChar(i)
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
						sys.chars[j], sys.cgi[j].states, charDone[j] = nil, nil, true
						sys.cgi[j].wakewakaLength = 0
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

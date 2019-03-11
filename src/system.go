package main

import (
	"bufio"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/faiface/beep"
	"github.com/faiface/beep/speaker"
	"github.com/go-gl/gl/v2.1/gl"
	"github.com/go-gl/glfw/v3.2/glfw"
	"github.com/timshannon/go-openal/openal"
	lua "github.com/yuin/gopher-lua"
)

const (
	MaxSimul        = 8
	MaxAttachedChar = 2
	FPS             = 60
	P1P3Dist        = 25
	Mp3SampleRate   = 44100
)

var sys = System{
	randseed:  int32(time.Now().UnixNano()),
	scrrect:   [...]int32{0, 0, 320, 240},
	gameWidth: 320, gameHeight: 240,
	widthScale: 1, heightScale: 1,
	brightness: 256,
	roundTime:  -1,
	lifeMul:    1, team1VS2Life: 1,
	turnsRecoveryRate: 1.0 / 300,
	mixer:             *newMixer(),
	bgm:               *newBgm(),
	sounds:            newSounds(16),
	allPalFX:          *newPalFX(),
	bgPalFX:           *newPalFX(),
	sel:               *newSelect(),
	keySatate:         make(map[glfw.Key]bool),
	match:             1,
	listenPort:        "7500",
	loader:            *newLoader(),
	numSimul:          [...]int32{2, 2}, numTurns: [...]int32{2, 2},
	ignoreMostErrors: true,
	superpmap:        *newPalFX(),
	wincnt:           wincntMap(make(map[string][]int32)),
	wincntFileName:   "autolevel.txt",
	powerShare:       [...]bool{true, true},
	oldNextAddTime:   1,
	commandLine:      make(chan string),
	cam:              *newCamera(),
	mainThreadTask:   make(chan func(), 65536),
	workpal:          make([]uint32, 256),
	errLog:           log.New(os.Stderr, "", 0),
	audioClose:       make(chan bool, 1),
	keyInput:         glfw.KeyUnknown,
	keyString:        "",
	// Localcoord sceenpack
	luaSpriteScale:        1,
	luaSmallPortraitScale: 1,
	luaBigPortraitScale:   1,
	luaSpriteOffsetX:      0,
}

type TeamMode int32

const (
	TM_Single TeamMode = iota
	TM_Simul
	TM_Turns
	TM_LAST = TM_Turns
)

type System struct {
	randseed                int32
	scrrect                 [4]int32
	gameWidth, gameHeight   int32
	widthScale, heightScale float32
	window                  *glfw.Window
	gameEnd, frameSkip      bool
	redrawWait              struct{ nextTime, lastDraw time.Time }
	brightness              int32
	roundTime               int32
	lifeMul, team1VS2Life   float32
	turnsRecoveryRate       float32
	lifebarFontScale        float32
	debugFont               *Fnt
	debugScript             string
	debugDraw               bool
	mixer                   Mixer
	bgm                     Bgm
	audioContext            *openal.Context
	nullSndBuf              [audioOutLen * 2]int16
	sounds                  Sounds
	allPalFX, bgPalFX       PalFX
	lifebar                 Lifebar
	sel                     Select
	keySatate               map[glfw.Key]bool
	netInput                *NetInput
	fileInput               *FileInput
	aiInput                 [MaxSimul*2 + MaxAttachedChar]AiInput
	keyConfig               []KeyConfig
	JoystickConfig          []KeyConfig
	com                     [MaxSimul*2 + MaxAttachedChar]int32
	autolevel               bool
	home                    int
	gameTime                int32
	match                   int32
	inputRemap              [MaxSimul*2 + MaxAttachedChar]int
	listenPort              string
	round                   int32
	intro                   int32
	time                    int32
	winTeam                 int
	winType                 [2]WinType
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
	attack_LifeToPowerMul   float32
	getHit_LifeToPowerMul   float32
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
	super_TargetDefenceMul  float32
	envcol                  [3]int32
	envcol_time             int32
	envcol_under            bool
	clipboardText           [MaxSimul*2 + MaxAttachedChar][]string
	stage                   *Stage
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
	shortcutScripts         map[ShortcutKey]*ShortcutScript
	turbo                   float32
	commandLine             chan string
	drawScale               float32
	zoomlag                 float32
	zoomPos                 [2]float32
	debugWC                 *Char
	cam                     Camera
	finish                  FinishType
	waitdown                int32
	shuttertime             int32
	projs                   [MaxSimul*2 + MaxAttachedChar][]Projectile
	explods                 [MaxSimul*2 + MaxAttachedChar][]Explod
	explDrawlist            [MaxSimul*2 + MaxAttachedChar][]int
	topexplDrawlist         [MaxSimul*2 + MaxAttachedChar][]int
	changeStateNest         int32
	sprites                 DrawList
	topSprites              DrawList
	shadows                 ShadowList
	drawc1                  ClsnRect
	drawc2                  ClsnRect
	drawc2sp                ClsnRect
	drawc2mtk               ClsnRect
	drawwh                  ClsnRect
	autoguard               [MaxSimul*2 + MaxAttachedChar]bool
	clsnDraw                bool
	accel                   float32
	statusDraw              bool
	clsnSpr                 Sprite
	mainThreadTask          chan func()
	explodMax               int
	workpal                 []uint32
	playerProjectileMax     int
	errLog                  *log.Logger
	audioClose              chan bool
	nomusic                 bool
	workBe                  []BytecodeExp
	teamLifeShare           bool
	fullscreen              bool
	aiRandomColor           bool
	allowDebugKeys          bool
	commonAir               string
	commonCmd               string
	keyInput                glfw.Key
	keyString               string
	timerCount              []int32
	cmdFlags                map[string]string
	quickLaunch             bool
	// Localcoord sceenpack
	luaSpriteScale        float64
	luaSmallPortraitScale float32
	luaBigPortraitScale   float32
	luaSpriteOffsetX      float64
}

func (s *System) init(w, h int32) *lua.LState {
	glfw.WindowHint(glfw.Resizable, glfw.False)
	glfw.WindowHint(glfw.ContextVersionMajor, 2)
	glfw.WindowHint(glfw.ContextVersionMinor, 1)
	s.setWindowSize(w, h)
	var err error
	if s.fullscreen {
		s.window, err = glfw.CreateWindow(int(s.scrrect[2]), int(s.scrrect[3]),
			"Ikemen GO", glfw.GetPrimaryMonitor(), nil)
	} else {
		s.window, err = glfw.CreateWindow(int(s.scrrect[2]), int(s.scrrect[3]),
			"Ikemen GO", nil, nil)
	}
	chk(err)
	s.window.MakeContextCurrent()
	s.window.SetKeyCallback(keyCallback)
	s.window.SetCharModsCallback(charCallback)
	glfw.SwapInterval(1)
	chk(gl.Init())
	RenderInit()
	s.audioOpen()
	sr := beep.SampleRate(Mp3SampleRate)
	speaker.Init(sr, sr.N(time.Second/10))
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
	s.gameEnd = s.window.ShouldClose()
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
		s.window.SwapBuffers()
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
		gl.Viewport(0, 0, int32(s.scrrect[2]), int32(s.scrrect[3]))
		gl.Clear(gl.COLOR_BUFFER_BIT)
	}
	return !s.gameEnd
}
func (s *System) update() bool {
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
func (s *System) audioOpen() {
	if s.audioContext == nil {
		device := openal.OpenDevice("")
		if device == nil {
			chk(openal.Err())
		}
		s.audioContext = device.CreateContext()
		if err := device.Err(); err != nil {
			s.errLog.Println(err.Error())
		}
		s.audioContext.Activate()
		go s.soundWrite()
	}
}
func (s *System) soundWrite() {
	defer func() { s.audioClose <- true }()
	src := NewAudioSource()
	bgmSrc := NewAudioSource()
	processed := false
	for !s.gameEnd {
		if src.Src.State() != openal.Playing {
			src.Src.Play()
		}
		if bgmSrc.Src.State() != openal.Playing {
			bgmSrc.Src.Play()
		}
		if !processed {
			time.Sleep(10 * time.Millisecond)
		}
		processed = false
		if src.Src.BuffersProcessed() > 0 {
			var out []int16
			select {
			case out = <-s.mixer.out:
			default:
				out = s.nullSndBuf[:]
			}
			buf := src.Src.UnqueueBuffer()
			buf.SetDataInt16(openal.FormatStereo16, out, audioFrequency)
			src.Src.QueueBuffer(buf)
			if err := openal.Err(); err != nil {
				s.errLog.Println(err.Error())
			}
			processed = true
		}
		if bgmSrc.Src.BuffersProcessed() > 0 {
			out := s.nullSndBuf[:]
			if !s.nomusic {
				if s.bgm.IsVorbis() {
					out = s.bgm.ReadVorbis()
				} else if s.bgm.IsMp3() && s.bgm.ctrlmp3 != nil {
					s.bgm.ctrlmp3.Paused = false
				}
			} else {
				if s.bgm.IsMp3() && s.bgm.ctrlmp3 != nil {
					s.bgm.Mp3Paused()
				}
			}
			buf := bgmSrc.Src.UnqueueBuffer()
			buf.SetDataInt16(openal.FormatStereo16, out, audioFrequency)
			bgmSrc.Src.QueueBuffer(buf)
			if err := openal.Err(); err != nil {
				s.errLog.Println(err.Error())
			}
			processed = true
		}
	}
	bgmSrc.Delete()
	src.Delete()
	openal.NullContext.Activate()
	device := s.audioContext.GetDevice()
	s.audioContext.Destroy()
	s.audioContext = nil
	device.CloseDevice()
}
func (s *System) playSound() {
	if s.mixer.write() {
		s.sounds.mixSounds()
		for _, ch := range s.chars {
			for _, c := range ch {
				c.sounds.mixSounds()
			}
		}
	}
}
func (s *System) resetRemapInput() {
	for i := range s.inputRemap {
		s.inputRemap[i] = i
	}
}
func (s *System) loaderReset() {
	s.round, s.wins, s.roundsExisted = 1, [2]int32{}, [2]int32{}
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
		if kc.A() || kc.B() || kc.C() || kc.X() || kc.Y() || kc.Z() {
			return true
		}
	}
	for _, kc := range s.JoystickConfig {
		if kc.A() || kc.B() || kc.C() || kc.X() || kc.Y() || kc.Z() {
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
func (s *System) roundOver() bool {
	if s.intro < -(s.lifebar.ro.over_hittime+s.lifebar.ro.over_waittime+
		s.lifebar.ro.over_wintime) && s.tickFrame() && s.anyButton() {
		s.intro = Min(s.intro, -(s.lifebar.ro.over_hittime +
			s.lifebar.ro.over_waittime + s.lifebar.ro.over_time -
			s.lifebar.ro.start_waittime))
		s.winskipped = true
	}
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
func (s *System) appendToClipboard(pn, sn int, a ...interface{}) {
	spl := s.stringPool[pn].List
	if sn >= 0 && sn < len(spl) {
		s.clipboardText[pn] = append(s.clipboardText[pn],
			strings.Split(OldSprintf(spl[sn], a...), "\n")...)
		if len(s.clipboardText[pn]) > 10 {
			s.clipboardText[pn] = s.clipboardText[pn][len(s.clipboardText[pn])-10:]
		}
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
		for i2 := 0; i2+3 < len(clsn2); i2 += 4 {
			var l1, r1, l2, r2 float32
			if facing1 > 0 {
				l1, r1 = clsn1[i1], clsn1[i1+2]+1
			} else {
				l1, r1 = -clsn1[i1+2], -clsn1[i1]+1
			}
			if facing2 > 0 {
				l2, r2 = clsn2[i2], clsn2[i2+2]+1
			} else {
				l2, r2 = -clsn2[i2+2], -clsn2[i2]+1
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
			c.sounds = c.sounds[:0]
		}
	}
}
func (s *System) playerClear(pn int) {
	if len(s.chars[pn]) > 0 {
		for _, h := range s.chars[pn][1:] {
			h.destroy()
			h.sounds = h.sounds[:0]
		}
		p := s.chars[pn][0]
		p.children = p.children[:0]
		p.targets = p.targets[:0]
		p.sounds = p.sounds[:0]
	}
	s.projs[pn] = s.projs[pn][:0]
	s.explods[pn] = s.explods[pn][:0]
	s.explDrawlist[pn] = s.explDrawlist[pn][:0]
	s.topexplDrawlist[pn] = s.topexplDrawlist[pn][:0]
}
func (s *System) nextRound() {
	s.resetGblEffect()
	s.lifebar.reset()
	s.finish = FT_NotYet
	s.winTeam = -1
	s.winType = [...]WinType{WT_N, WT_N}
	s.cam.ResetZoomdelay()
	s.waitdown = s.lifebar.ro.over_hittime*s.lifebar.ro.over_waittime + 900
	s.shuttertime = 0
	s.winskipped = false
	s.intro = s.lifebar.ro.start_waittime + s.lifebar.ro.ctrl_time + 1
	s.time = s.roundTime
	s.nextCharId = s.helperMax
	if s.stage.resetbg {
		s.stage.reset()
	}
	s.cam.Update(1, 0, 0)
	for i, p := range s.chars {
		if len(p) > 0 {
			s.nextCharId = Max(s.nextCharId, p[0].id+1)
			s.playerClear(i)
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
			p[0].selfState(5900, 0, 0)
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
		!s.paused || s.step || s.oldTickCount >= s.tickCount
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
			act := true
			if s.super > 0 {
				act = r.superMovetime != 0
			} else if s.pause > 0 && r.pauseMovetime == 0 {
				act = false
			}
			if act && !r.sf(CSF_noautoturn) &&
				(r.ss.no == 0 || r.ss.no == 11 || r.ss.no == 20) {
				r.furimuki()
			}
			for _, c := range p {
				if (c.helperIndex == 0 ||
					c.helperIndex > 0 && &c.cmd[0] != &r.cmd[0]) &&
					c.cmd[0].Input(c.key, int32(c.facing)) {
					hp := c.hitPause()
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
				if r.roundState() == 2 && Rand(0, s.com[i]+16) > 16 {
					cc = Rand(0, int32(len(r.cmd[r.ss.sb.playerNo].Commands))-1)
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
func (s *System) action(x, y *float32, scl float32) (leftest, rightest,
	sclMul float32) {
	s.sprites = s.sprites[:0]
	s.topSprites = s.topSprites[:0]
	s.shadows = s.shadows[:0]
	s.drawc1 = s.drawc1[:0]
	s.drawc2 = s.drawc2[:0]
	s.drawc2sp = s.drawc2sp[:0]
	s.drawc2mtk = s.drawc2mtk[:0]
	s.drawwh = s.drawwh[:0]
	s.cam.Update(scl, *x, *y)
	var cvmin, cvmax, highest, lowest float32 = 0, 0, 0, 0
	leftest, rightest = *x, *x
	if s.cam.verticalfollow > 0 {
		lowest = s.cam.ScreenPos[1]
	}
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
		s.drawScale, s.zoomPos = float32(math.NaN()), [2]float32{}
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
		if s.super <= 0 && s.pause <= 0 {
			s.specialFlag = 0
		} else {
			s.unsetSF(GSF_roundnotover)
		}
		if s.superanim != nil {
			s.superanim.Action()
		}
		s.charList.action(*x, &cvmin, &cvmax,
			&highest, &lowest, &leftest, &rightest)
		s.nomusic = s.sf(GSF_nomusic)
	} else {
		s.charUpdate(&cvmin, &cvmax, &highest, &lowest, &leftest, &rightest)
	}
	s.lifebar.step()
	if s.superanim != nil {
		s.topSprites.add(&SprData{s.superanim, &s.superpmap, s.superpos,
			[...]float32{s.superfacing, 1}, [2]int32{-1}, 5, 0, 0, 0, [2]float32{},
			false, true, s.cgi[s.superplayer].ver[0] != 1, 1}, 0, 0, 0, 0)
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
	leftest -= *x
	rightest -= *x
	sclMul = s.cam.action(x, y, leftest, rightest, lowest, highest,
		cvmin, cvmax, s.super > 0 || s.pause > 0)
	introSkip := false
	if s.tickNextFrame() {
		if s.lifebar.ro.cur < 1 {
			if s.shuttertime > 0 ||
				s.anyButton() && s.intro > s.lifebar.ro.ctrl_time {
				s.shuttertime++
				if s.shuttertime == 15 {
					s.resetGblEffect()
					s.intro = s.lifebar.ro.ctrl_time
					for i, p := range s.chars {
						if len(p) > 0 {
							s.playerClear(i)
							p[0].selfState(0, -1, 0)
						}
					}
					ox := *x
					*x = 0
					leftest = MaxF(float32(Min(s.stage.p[0].startx,
						s.stage.p[1].startx))*s.stage.localscl,
						-(float32(s.gameWidth)/2)/s.cam.BaseScale()+s.screenleft) - ox
					rightest = MinF(float32(Max(s.stage.p[0].startx,
						s.stage.p[1].startx))*s.stage.localscl,
						(float32(s.gameWidth)/2)/s.cam.BaseScale()-s.screenright) - ox
					introSkip = true
					s.lifebar.ro.callFight()
				}
			}
		} else {
			if s.shuttertime > 0 {
				s.shuttertime--
			}
		}
	}
	if s.lifebar.ro.act() {
		if s.intro > s.lifebar.ro.ctrl_time {
			s.intro--
			if s.sf(GSF_intro) && s.intro <= s.lifebar.ro.ctrl_time {
				s.intro = s.lifebar.ro.ctrl_time + 1
			}
		} else if s.intro > 0 {
			if s.intro == s.lifebar.ro.ctrl_time {
				for _, p := range s.chars {
					if len(p) > 0 {
						p[0].posReset()
					}
				}
			}
			s.intro--
			if s.intro == 0 {
				for _, p := range s.chars {
					if len(p) > 0 {
						p[0].unsetSCF(SCF_over)
						if p[0].ss.no == 0 {
							p[0].setCtrl(true)
						} else {
							p[0].selfState(0, -1, 1)
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
				for i := ii; i < len(s.chars); i += 2 {
					if len(s.chars[i]) > 0 && s.chars[i][0].alive() && s.chars[i][0].teamside < 2 {
						ko[ii] = false
						break
					}
				}
				if ko[ii] {
					i := ii ^ 1
					for ; i < len(s.chars); i += 2 {
						if len(s.chars[i]) > 0 && s.chars[i][0].life <
							s.chars[i][0].lifeMax {
							break
						}
					}
					if i >= len(s.chars) {
						s.winType[ii^1].SetPerfect()
					}
				}
			}
			if s.time == 0 {
				s.intro = -s.lifebar.ro.over_hittime
				if !(ko[0] || ko[1]) {
					s.winType[0], s.winType[1] = WT_T, WT_T
				}
			}
			if s.intro == -s.lifebar.ro.over_hittime && (ko[0] || ko[1]) {
				if ko[0] && ko[1] {
					s.finish, s.winTeam = FT_DKO, -1
				} else {
					s.finish = FT_KO
					if ko[0] {
						s.winTeam = 1
					} else {
						s.winTeam = 0
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
					s.draws >= s.lifebar.ro.match_maxdrawgames {
					for i, win := range w {
						if win {
							s.wins[i]++
						}
					}
				}
			}
			if s.intro == -s.lifebar.ro.over_hittime && s.finish != FT_NotYet {
				inclWinCount()
			}
			rs4t := -(s.lifebar.ro.over_hittime + s.lifebar.ro.over_waittime)
			if s.winskipped || !s.sf(GSF_roundnotover) ||
				s.intro >= rs4t-s.lifebar.ro.over_wintime {
				s.intro--
				if s.intro == rs4t-1 {
					if s.time == 0 {
						s.intro -= s.lifebar.ro.over_wintime
					}
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
						if s.finish == FT_NotYet {
							l := [2]float32{}
							for i := 0; i < 2; i++ {
								for j := i; j < len(s.chars); j += 2 {
									if len(s.chars[j]) > 0 {
										if s.tmode[i] == TM_Simul {
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
								for i := 0; i < len(s.chars); i += 2 {
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
								for i := 1; i < len(s.chars); i += 2 {
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
							inclWinCount()
						}
						w := [...]bool{!s.chars[1][0].win(), !s.chars[0][0].win()}
						if !w[0] || !w[1] ||
							s.tmode[0] == TM_Turns || s.tmode[1] == TM_Turns ||
							s.draws >= s.lifebar.ro.match_maxdrawgames {
							for i, win := range w {
								if win {
									s.lifebar.wi[i].add(s.winType[i])
								}
							}
						} else {
							s.draws++
						}
					}
					for _, p := range s.chars {
						if len(p) > 0 {
							if s.waitdown >= 0 && s.time > 0 && p[0].win() && p[0].alive() &&
								!s.matchOver() &&
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
									p[0].selfState(180, -1, 1)
								} else if p[0].lose() {
									p[0].selfState(170, -1, 1)
								} else {
									p[0].selfState(175, -1, 1)
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
		spd := s.accel
		_else := s.sf(GSF_nokoslow) || s.time == 0
		if !_else {
			slowt := -(s.lifebar.ro.over_hittime + (s.lifebar.ro.slow_time+3)>>2)
			if s.intro >= slowt && s.intro < -s.lifebar.ro.over_hittime {
				s.turbo = spd * 0.25
			} else {
				slowfade := s.lifebar.ro.slow_time * 2 / 5
				if s.intro >= slowt-slowfade && s.intro < slowt {
					s.turbo = spd *
						(0.25 + 0.75*float32(slowt-s.intro)/float32(slowfade))
				} else {
					_else = true
				}
			}
		}
		if _else {
			s.turbo = spd
		}
	}
	s.playSound()
	if introSkip {
		sclMul = 1 / scl
	}
	leftest = (leftest - s.screenleft) * s.cam.BaseScale()
	rightest = (rightest + s.screenright) * s.cam.BaseScale()
	return
}
func (s *System) draw(x, y, scl float32) {
	ecol := uint32(s.envcol[2]&0xff | s.envcol[1]&0xff<<8 |
		s.envcol[0]&0xff<<16)
	ob := s.brightness
	s.brightness = 0x100 >> uint(Btoi(s.super > 0 && s.superdarken))
	bgx, bgy := x/s.stage.localscl, y/s.stage.localscl
	fade := func(rect [4]int32, alpha int32) {
		FillRect(rect, 0, alpha>>uint(Btoi(s.clsnDraw))+Btoi(s.clsnDraw)*128)
	}
	if s.envcol_time == 0 {
		if s.sf(GSF_nobg) {
			c := uint32(0)
			if s.allPalFX.enable {
				var rgb [3]int32
				if s.allPalFX.eInvertall {
					rgb = [...]int32{0xff, 0xff, 0xff}
				}
				for i, v := range rgb {
					rgb[i] = Max(0, Min(0xff,
						(v+s.allPalFX.eAdd[i])*s.allPalFX.eMul[i]>>8))
				}
				c = uint32(rgb[2] | rgb[1]<<8 | rgb[0]<<16)
			}
			FillRect(s.scrrect, c, 0xff)
		} else {
			if s.stage.debugbg {
				FillRect(s.scrrect, 0xff00ff, 0xff)
			}
			s.stage.draw(false, bgx, bgy, scl)
		}
		if !s.sf(GSF_globalnoshadow) {
			if s.stage.reflection > 0 {
				s.shadows.drawReflection(x, y, scl*s.cam.BaseScale())
			}
			s.shadows.draw(x, y, scl*s.cam.BaseScale())
		}
		off := s.envShake.getOffset()
		yofs, yofs2 := float32(s.gameHeight), float32(0)
		if scl > 1 && s.cam.verticalfollow > 0 {
			yofs = s.cam.screenZoff + float32(s.gameHeight-240)
			yofs2 = (240 - s.cam.screenZoff) * (1 - 1/scl)
		}
		yofs *= 1/scl - 1
		rect := s.scrrect
		if off < (yofs-y+s.cam.boundH)*scl {
			rect[3] = (int32(math.Ceil(float64(((yofs-y+s.cam.boundH)*scl-off)*
				float32(s.scrrect[3])))) + s.gameHeight - 1) / s.gameHeight
			fade(rect, 255)
		}
		if off > (-y+yofs2)*scl {
			rect[3] = (int32(math.Ceil(float64(((y-yofs2)*scl+off)*
				float32(s.scrrect[3])))) + s.gameHeight - 1) / s.gameHeight
			rect[1] = s.scrrect[3] - rect[3]
			fade(rect, 255)
		}
		bl, br := MinF(x, s.cam.boundL), MaxF(x, s.cam.boundR)
		xofs := float32(s.gameWidth) * (1/scl - 1) / 2
		rect = s.scrrect
		if x-xofs < bl {
			rect[2] = (int32(math.Ceil(float64((bl-(x-xofs))*scl*
				float32(s.scrrect[2])))) + s.gameWidth - 1) / s.gameWidth
			fade(rect, 255)
		}
		if x+xofs > br {
			rect[2] = (int32(math.Ceil(float64(((x+xofs)-br)*scl*
				float32(s.scrrect[2])))) + s.gameWidth - 1) / s.gameWidth
			rect[0] = s.scrrect[2] - rect[2]
			fade(rect, 255)
		}
		s.lifebar.draw(0)
		s.lifebar.ro.draw(0)
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
	s.lifebar.ro.draw(1)
	s.topSprites.draw(x, y, scl*s.cam.BaseScale())
	s.lifebar.draw(2)
	s.lifebar.ro.draw(2)
	tmp := s.lifebar.ro.over_hittime + s.lifebar.ro.over_waittime +
		s.lifebar.ro.over_time - s.lifebar.ro.start_waittime
	if s.intro > s.lifebar.ro.ctrl_time+1 {
		fade(s.scrrect, 256*(s.intro-(s.lifebar.ro.ctrl_time+1))/
			s.lifebar.ro.start_waittime)
	} else if s.lifebar.ro.over_time >= s.lifebar.ro.start_waittime &&
		s.intro < -tmp {
		fade(s.scrrect, 256*(-tmp-s.intro)/s.lifebar.ro.start_waittime)
	} else if s.clsnDraw {
		fade(s.scrrect, 0)
	}
	if s.shuttertime > 0 {
		rect := s.scrrect
		rect[3] = s.shuttertime * ((s.scrrect[3] + 1) >> 1) / 15
		fade(rect, 255)
		rect[1] = s.scrrect[3] - rect[3]
		fade(rect, 255)
	}
	s.brightness = ob
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
func (s *System) fight() (reload bool) {
	s.gameTime, s.paused, s.accel, s.statusDraw = 0, false, 1, true
	for i := range s.clipboardText {
		s.clipboardText[i] = nil
	}
	s.aiInput = [len(s.aiInput)]AiInput{}
	s.shortcutScripts = make(map[ShortcutKey]*ShortcutScript)
	defer func() {
		s.oldNextAddTime = 1
		s.nomusic = false
		s.allPalFX.clear()
		s.allPalFX.enable = false
		for i, p := range s.chars {
			if len(p) > 0 {
				s.playerClear(i)
			}
		}
		s.wincnt.update()
	}()
	var life, pow [len(s.chars)]int32
	var ivar [len(s.chars)][]int32
	var fvar [len(s.chars)][]float32
	copyVar := func(pn int) {
		life[pn] = s.chars[pn][0].life
		pow[pn] = s.chars[pn][0].power
		if len(ivar[pn]) < len(s.chars[pn][0].ivar) {
			ivar[pn] = make([]int32, len(s.chars[pn][0].ivar))
		}
		copy(ivar[pn], s.chars[pn][0].ivar[:])
		if len(fvar[pn]) < len(s.chars[pn][0].fvar) {
			fvar[pn] = make([]float32, len(s.chars[pn][0].fvar))
		}
		copy(fvar[pn], s.chars[pn][0].fvar[:])
	}
	s.debugWC = nil
	dL := lua.NewState()
	defer dL.Close()
	var statusLFunc *lua.LFunction
	if len(s.debugScript) > 0 {
		if err := debugScriptInit(dL, s.debugScript); err != nil {
			s.errLog.Println(err.Error())
		} else {
			statusLFunc, _ = dL.GetGlobal("status").(*lua.LFunction)
		}
	}
	debugInput := func() {
		select {
		case cl := <-s.commandLine:
			if err := dL.DoString(cl); err != nil {
				s.errLog.Println(err.Error())
			}
		default:
		}
	}
	put := func(y *float32, txt string) {
		tmp := s.allPalFX.enable
		s.allPalFX.enable = false
		for txt != "" {
			w, drawTxt := int32(0), ""
			for i, r := range txt {
				w += s.debugFont.CharWidth(r) + s.debugFont.Spacing[0]
				if w > s.scrrect[2] {
					drawTxt, txt = txt[:i], txt[i:]
					break
				}
			}
			if drawTxt == "" {
				drawTxt, txt = txt, ""
			}
			*y += float32(s.debugFont.Size[1]) / s.heightScale
			s.debugFont.DrawText(drawTxt, (320-float32(s.gameWidth))/2, *y,
				1/s.widthScale, 1/s.heightScale, 0, 1)
		}
		s.allPalFX.enable = tmp
	}
	drawDebug := func() {
		if s.debugDraw && s.debugFont != nil {
			y := 240 - float32(s.gameHeight)
			if statusLFunc != nil {
				for i, p := range s.chars {
					if len(p) > 0 {
						top := dL.GetTop()
						if dL.CallByParam(lua.P{Fn: statusLFunc, NRet: 1,
							Protect: true}, lua.LNumber(i+1)) == nil {
							s, ok := dL.Get(-1).(lua.LString)
							if ok && len(s) > 0 {
								put(&y, string(s))
							}
						}
						dL.SetTop(top)
					}
				}
			}
			y = MaxF(y, 48+240-float32(s.gameHeight))
			for i, p := range s.chars {
				if len(p) > 0 {
					put(&y, s.cgi[i].def)
				}
			}
			put(&y, s.stage.def)
			if s.debugWC != nil {
				put(&y, fmt.Sprintf("<P%v:%v>", s.debugWC.playerNo+1, s.debugWC.name))
			}
			for i, p := range s.chars {
				if len(p) > 0 {
					for _, s := range s.clipboardText[i] {
						put(&y, s)
					}
				}
				y += float32(s.debugFont.Size[1]) / s.heightScale
			}
		}
	}
	if err := s.synchronize(); err != nil {
		s.errLog.Println(err.Error())
		s.esc = true
	}
	if s.netInput != nil {
		defer s.netInput.Stop()
	}
	s.wincnt.init()
	var level [len(s.chars)]int32
	for i, p := range s.chars {
		if len(p) > 0 {
			p[0].clear2()
			level[i] = s.wincnt.getLevel(i)
			if s.powerShare[i&1] {
				pmax := Max(s.cgi[i&1].data.power, s.cgi[i].data.power)
				for j := i & 1; j < len(s.chars); j += 2 {
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
	lvmul := math.Pow(2, 1.0/12)
	for i, p := range s.chars {
		if len(p) > 0 {
			lm := float32(p[0].gi().data.life) * s.lifeMul
			switch s.tmode[i&1] {
			case TM_Single:
				switch s.tmode[(i+1)&1] {
				case TM_Simul:
					lm *= s.team1VS2Life
				case TM_Turns:
					if s.numTurns[(i+1)&1] < s.matchWins[(i+1)&1] && sys.teamLifeShare {
						lm = lm * float32(s.numTurns[(i+1)&1]) /
							float32(s.matchWins[(i+1)&1])
					}
				}
			case TM_Simul:
				switch s.tmode[(i+1)&1] {
				case TM_Simul:
					if s.numSimul[(i+1)&1] < s.numSimul[i&1] && sys.teamLifeShare {
						lm = lm * float32(s.numSimul[(i+1)&1]) / float32(s.numSimul[i&1])
					}
				case TM_Turns:
					if s.numTurns[(i+1)&1] < s.numSimul[i&1]*s.matchWins[(i+1)&1] && sys.teamLifeShare {
						lm = lm * float32(s.numTurns[(i+1)&1]) /
							float32(s.numSimul[i&1]*s.matchWins[(i+1)&1])
					}
				default:
					if sys.teamLifeShare {
						lm /= float32(s.numSimul[i&1])
					}
				}
			case TM_Turns:
				switch s.tmode[(i+1)&1] {
				case TM_Single:
					if s.matchWins[i&1] < s.numTurns[i&1] && sys.teamLifeShare {
						lm = lm * float32(s.matchWins[i&1]) / float32(s.numTurns[i&1])
					}
				case TM_Simul:
					if s.numSimul[(i+1)&1]*s.matchWins[i&1] < s.numTurns[i&1] && sys.teamLifeShare {
						lm = lm * s.team1VS2Life *
							float32(s.numSimul[(i+1)&1]*s.matchWins[i&1]) /
							float32(s.numTurns[i&1])
					}
				case TM_Turns:
					if s.numTurns[(i+1)&1] < s.numTurns[i&1] && sys.teamLifeShare {
						lm = lm * float32(s.numTurns[(i+1)&1]) / float32(s.numTurns[i&1])
					}
				}
			}
			foo := math.Pow(lvmul, float64(-level[i]))
			p[0].lifeMax = Max(1, int32(math.Floor(foo*float64(lm))))
			if s.roundsExisted[i&1] > 0 {
				p[0].life = Min(p[0].lifeMax, int32(math.Ceil(foo*float64(p[0].life))))
			} else if s.round == 1 || s.tmode[i&1] == TM_Turns {
				p[0].life = p[0].lifeMax
				if s.round == 1 {
					p[0].power = 0
				}
			}
			copyVar(i)
		}
	}
	if s.round == 1 {
		s.bgm.Open(s.stage.bgmusic)
	}
	s.cam.Init()
	s.screenleft = float32(s.stage.screenleft) * s.stage.localscl
	s.screenright = float32(s.stage.screenright) * s.stage.localscl
	oldWins, oldDraws := s.wins, s.draws
	var x, y, newx, newy, l, r float32
	var scl, sclmul float32
	reset := func() {
		s.wins, s.draws = oldWins, oldDraws
		for i, p := range s.chars {
			if len(p) > 0 {
				p[0].life = life[i]
				p[0].power = pow[i]
				copy(p[0].ivar[:], ivar[i])
				copy(p[0].fvar[:], fvar[i])
			}
		}
		s.resetFrameTime()
		s.nextRound()
		x, y, newx, newy, l, r, scl, sclmul = 0, 0, 0, 0, 0, 0, 1, 1
		s.cam.Update(scl, x, y)
	}
	reset()
	for !s.esc {
		s.step, s.roundResetFlg, s.reloadFlg = false, false, false
		for _, v := range s.shortcutScripts {
			if v.Activate {
				if err := dL.DoString(v.Script); err != nil {
					s.errLog.Println(err.Error())
				}
			}
		}
		if s.roundResetFlg {
			reset()
		}
		if s.reloadFlg {
			return true
		}
		if s.roundOver() {
			s.round++
			for i := range s.roundsExisted {
				s.roundsExisted[i]++
			}
			if !s.matchOver() && (s.tmode[0] != TM_Turns || s.chars[0][0].win()) &&
				(s.tmode[1] != TM_Turns || s.chars[1][0].win()) {
				for i, p := range s.chars {
					if len(p) > 0 {
						if s.tmode[i&1] != TM_Turns || !p[0].win() {
							p[0].life = p[0].lifeMax
						} else if p[0].life <= 0 {
							p[0].life = 1
						}
						copyVar(i)
					}
				}
				oldWins, oldDraws = s.wins, s.draws
				reset()
			} else {
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
					} else {
						s.chars[i][0].life = 0
					}
				}
				break
			}
		}
		scl = s.cam.ScaleBound(scl, sclmul)
		tmp := (float32(s.gameWidth) / 2) / scl
		if AbsF((l+r)-(newx-x)*2) >= tmp/2 {
			tmp = MaxF(0, MinF(tmp, MaxF((newx-x)-l, r-(newx-x))))
		}
		x = s.cam.XBound(scl, MinF(x+l+tmp, MaxF(x+r-tmp, newx)))
		if !s.cam.ZoomEnable {
			// Pos X の誤差が出ないように精度を落とす
			x = float32(math.Ceil(float64(x)*4-0.5) / 4)
		}
		y = s.cam.YBound(scl, newy)
		if s.tickFrame() && (s.super <= 0 || !s.superpausebg) &&
			(s.pause <= 0 || !s.pausebg) {
			s.stage.action()
		}
		newx, newy = x, y
		l, r, sclmul = s.action(&newx, &newy, scl)
		debugInput()
		if !s.addFrameTime(s.turbo) {
			if !s.eventUpdate() {
				return false
			}
			continue
		}
		if !s.frameSkip {
			dx, dy, dscl := x, y, scl
			if !math.IsNaN(float64(s.drawScale)) &&
				!math.IsNaN(float64(s.zoomPos[0])) &&
				!math.IsNaN(float64(s.zoomPos[1])) {
				dscl = MaxF(s.cam.MinScale, s.drawScale/s.cam.BaseScale())
				dx = s.cam.XBound(dscl, x+s.zoomPos[0]/scl*s.drawScale)
				dy = y + s.zoomPos[1]
			} else {
				s.zoomlag = 1
			}
			s.draw(dx, dy, dscl)
			drawDebug()
		}
		if !s.update() {
			break
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
		if sys.tmode[(i+1)&1] == TM_Simul {
			if sys.tmode[i&1] != TM_Simul {
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
	lv, _ := wm[def]
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
	def, name, sprite, intro_storyboard, ending_storyboard string
	pal_defaults                                           []int32
	pal                                                    []int32
	portrait_scale                                         float32
	sportrait, lportrait, vsportrait, vportrait            *Sprite
}
type SelectStage struct {
	def, name, zoomout, zoomin, bgmusic, bgmvolume, attachedchardef string
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
	sportrait       [2]int16
	lportrait       [2]int16
	vsportrait      [2]int16
	vportrait       [2]int16
}

func newSelect() *Select {
	return &Select{columns: 5, rows: 2, randomscl: [...]float32{1, 1},
		cellsize: [...]float32{29, 29}, cellscale: [...]float32{1, 1},
		selectedStageNo: -1, sportrait: [...]int16{9000, 0}, lportrait: [...]int16{9000, 1},
		vsportrait: [...]int16{9000, 1}, vportrait: [...]int16{9000, 2}}
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
func (s *Select) GetStageInfo(n int) (zoomin, zoomout, bgmusic, bgmvolume string) {
	n %= len(s.stagelist) + 1
	if n < 0 {
		n += len(s.stagelist) + 1
	}
	return s.stagelist[n-1].zoomin, s.stagelist[n-1].zoomout, s.stagelist[n-1].bgmusic, s.stagelist[n-1].bgmvolume
}
func (s *Select) addCahr(def string) {
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
	if strings.ToLower(def[0:6]) != "chars/" && strings.ToLower(def[1:3]) != ":/" && (def[0] != '/' || idx > 0 && strings.Index(def[:idx], ":") < 0) {
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
	lines, i, info, files, arcade, sprite := SplitAndTrim(str, "\n"), 0, true, true, true, ""
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
				sc.pal_defaults = is.readI32CsvForStage("pal.defaults")
				ok = is.ReadF32("localcoord", &sc.portrait_scale)
				if !ok {
					sc.portrait_scale = 1
				} else {
					sc.portrait_scale = (320 / sc.portrait_scale)
				}
				is.ReadF32("portraitscale", &sc.portrait_scale)
			}
		case "files":
			if files {
				files = false
				sprite = is["sprite"]
				for i := 1; i <= MaxPalNo; i++ {
					if is[fmt.Sprintf("pal%v", i)] != "" {
						sc.pal = append(sc.pal, int32(i))
					}
				}
			}
		case "arcade":
			if arcade {
				arcade = false
				sc.intro_storyboard, _ = is.getString("intro.storyboard")
				sc.ending_storyboard, _ = is.getString("ending.storyboard")
			}
		}
	}
	sc.sprite = sprite
	LoadFile(&sprite, def, func(file string) error {
		var err error
		sc.sportrait, err = loadFromSff(file, sys.sel.sportrait[0], sys.sel.sportrait[1])
		if sys.quickLaunch {
			sc.lportrait = sc.sportrait
			sc.vsportrait, sc.vportrait = sc.lportrait, sc.lportrait
		} else {
			sc.lportrait, err = loadFromSff(file, sys.sel.lportrait[0], sys.sel.lportrait[1])
			sc.vsportrait, err = loadFromSff(file, sys.sel.vsportrait[0], sys.sel.vsportrait[1])
			if err != nil {
				sc.vsportrait = sc.lportrait
			}
			sc.vportrait, err = loadFromSff(file, sys.sel.vportrait[0], sys.sel.vportrait[1])
			if err != nil {
				sc.vportrait = sc.lportrait
			}
		}
		if len(sc.pal) == 0 {
			sc.pal, _ = selectablePalettes(file)
		}
		return nil
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
	i, info, camera, music := 0, true, true, true
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
				ss.attachedchardef, ok = is.getString("attachedchar")
			}
		case "camera":
			if camera {
				camera = false
				var ok bool
				ss.zoomout, ok = is.getString("setzoommax")
				if !ok {
					ss.zoomout = ""
				}
				ss.zoomin, ok = is.getString("setzoommin")
				if !ok {
					ss.zoomin = ""
				}
			}
		case "music":
			if music {
				music = false
				var ok bool
				ss.bgmusic, ok = is.getString("bgmusic")
				if !ok {
					ss.bgmusic = "100"
				}
				ss.bgmvolume, ok = is.getString("bgmvolume")
				if !ok {
					ss.bgmvolume = "100"
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
		pl = int(Rand(1, MaxPalNo))
	}
	sys.loadMutex.Lock()
	s.selected[tn] = append(s.selected[tn], [...]int{n, pl})
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
	cdef := sys.sel.charlist[idx[memberNo]].def
	var p *Char
	if len(sys.chars[pn]) > 0 && cdef == sys.cgi[pn].def {
		p = sys.chars[pn][0]
		p.key = pn
		if sys.com[pn] != 0 {
			p.key ^= -1
		}
	} else {
		p = newChar(pn, 0)
		sys.cgi[pn].sff = nil
		if len(sys.chars[pn]) > 0 {
			p.power = sys.chars[pn][0].power
		}
	}
	p.memberNo = memberNo
	p.selectNo = sys.sel.selected[pn&1][memberNo][0]
	p.teamside = p.playerNo & 1
	sys.chars[pn] = make([]*Char, 1)
	sys.chars[pn][0] = p
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
	if sys.roundsExisted[pn&1] == 0 {
		sys.cgi[pn].palno = sys.cgi[pn].palkeymap[pal-1] + 1
	}
	if pn < len(sys.lifebar.fa[sys.tmode[pn&1]]) &&
		sys.tmode[pn&1] == TM_Turns && sys.round == 1 {
		fa := sys.lifebar.fa[sys.tmode[pn&1]][pn]
		fa.numko, fa.teammate_face, fa.teammate_scale = 0, make([]*Sprite, nsel), make([]float32, nsel)
		for i, ci := range idx {
			sprite := sys.sel.charlist[ci].sprite
			LoadFile(&sprite, sys.sel.charlist[ci].def, func(file string) error {
				var err error
				fa.teammate_face[i], err = loadFromSff(file,
					int16(fa.teammate_face_spr[0]), int16(fa.teammate_face_spr[1]))
				fa.teammate_scale[i] = sys.sel.charlist[ci].portrait_scale
				return err
			})
		}
	}
	return 1
}

func (l *Loader) loadAttachedChar(atcpn int, def string) int {
	pn := atcpn + MaxSimul*2
	cdef := def
	var p *Char
	if len(sys.chars[pn]) > 0 && cdef == sys.cgi[pn].def {
		p = sys.chars[pn][0]
		p.key = -pn
	} else {
		p = newChar(pn, 0)
		sys.cgi[pn].sff = nil
		if len(sys.chars[pn]) > 0 {
			p.power = sys.chars[pn][0].power
		}
	}
	p.memberNo = -atcpn
	p.selectNo = -atcpn
	p.teamside = 2
	sys.com[pn] = 8
	sys.chars[pn] = make([]*Char, 1)
	sys.chars[pn][0] = p
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
	if sys.roundsExisted[pn&1] == 0 {
		sys.cgi[pn].palno = 1
	}
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
			l.loadAttachedChar(0, sys.sel.stagelist[sys.sel.selectedStageNo-1].attachedchardef)
		}
		if sys.stage != nil && sys.stage.def == def {
			return true
		}
		sys.stage, l.err = loadStage(def)
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

package main

import (
	"arena"
	"fmt"
	"hash/fnv"
	"strconv"
	"sync"
	"time"

	glfw "github.com/go-gl/glfw/v3.3/glfw"
	lua "github.com/yuin/gopher-lua"
)

func (cs Char) String() string {
	str := fmt.Sprintf(`Char %s 
	RedLife             :%d 
	Juggle              :%d 
	Life                :%d 
	Key                 :%d  
	Localcoord          :%f 
	Localscl            :%f 
	Pos                 :%v 
	DrawPos             :%v 
	OldPos              :%v 
	Vel                 :%v  
	Facing              :%f
	Id                  :%d
	HelperId            :%d
	HelperIndex         :%d
	ParentIndex         :%d
	PlayerNo            :%d
	Teamside            :%d
	AnimPN              :%d
	AnimNo              :%d
	LifeMax             :%d
	PowerMax            :%d
	DizzyPoints         :%d
	GuardPoints         :%d
	FallTime            :%d
	ClsnScale           :%v
	HoIdx               :%d
	Mctime              :%d
	Targets             :%v
	TargetsOfHitdef     :%v
	Atktmp              :%d
	Hittmp              :%d
	Acttmp              :%d
	Minus               :%d
	GroundAngle          :%f
	InheritJuggle         :%d
	Preserve              :%d
	Cnsvar              :%v
	Cnsfvar             :%v
	Offset              :%v`,
		cs.name, cs.redLife, cs.juggle, cs.life, cs.controller, cs.localcoord,
		cs.localscl, cs.pos, cs.interPos, cs.oldPos, cs.vel, cs.facing,
		cs.id, cs.helperId, cs.helperIndex, cs.parentIndex, cs.playerNo,
		cs.teamside, cs.animPN, cs.animNo, cs.lifeMax, cs.powerMax, cs.dizzyPoints,
		cs.guardPoints, cs.fallTime, cs.clsnScale, cs.hoverIdx, cs.mctime, cs.targets, cs.hitdefTargetsBuffer,
		cs.atktmp, cs.hittmp, cs.acttmp, cs.minus, cs.groundAngle, cs.inheritJuggle,
		cs.preserve, cs.cnsvar, cs.cnsfvar, cs.offset)
	return str
}

func (gs *GameState) getID() string {
	return strconv.Itoa(int(gs.id))
}

func (gs *GameState) Checksum() int {
	//	buf := bytes.Buffer{}
	//	enc := gob.NewEncoder(&buf)
	//	err := enc.Encode(gs)
	//	if err != nil {
	//		panic(err)
	//	}
	//	gs.bytes = buf.Bytes()
	gs.bytes = []byte(gs.String())
	h := fnv.New32a()
	h.Write(gs.bytes)
	return int(h.Sum32())
}

func (gs *GameState) String() (str string) {
	str = fmt.Sprintf("GameTime %d CurRoundTime: %d\n", gs.gameTime, gs.curRoundTime)
	str += fmt.Sprintf("bcStack: %v\n", gs.bcStack)
	str += fmt.Sprintf("bcVarStack: %v\n", gs.bcVarStack)
	str += fmt.Sprintf("bcVar: %v\n", gs.bcVar)
	str += fmt.Sprintf("workBe: %v\n", gs.workBe)
	for i := 0; i < len(gs.charData); i++ {
		for j := 0; j < len(gs.charData[i]); j++ {
			str += gs.charData[i][j].String()
			str += "\n"
		}
	}
	return
}

const MaxSaveStates = 8

type GameState struct {
	// Identifiers
	bytes []byte
	id    int
	saved bool
	frame int32

	// Selective copy of the system struct
	randseed     int32
	gameTime     int32
	curRoundTime int32

	chars      [MaxPlayerNo][]*Char
	charData   [MaxPlayerNo][]Char
	projs      [MaxPlayerNo][]*Projectile
	explods    [MaxPlayerNo][]*Explod
	aiInput    [MaxPlayerNo]AiInput
	inputRemap [MaxPlayerNo]int
	charList   CharList

	aiLevel            [MaxPlayerNo]float32 // UIT
	cam                Camera
	allPalFX           *PalFX
	bgPalFX            *PalFX
	pause              int32
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
	superdarken        bool

	envShake            EnvShake
	specialFlag         GlobalSpecialFlag // UIT
	envcol              [3]int32
	envcol_time         int32
	bcStack, bcVarStack BytecodeStack
	bcVar               []BytecodeValue
	workBe              []BytecodeExp

	scrrect                 [4]int32
	gameWidth, gameHeight   int32 // UIT
	widthScale, heightScale float32
	gameEnd, frameSkip      bool
	brightness              int32
	maxRoundTime            int32 // UIT
	team1VS2Life            float32
	turnsRecoveryRate       float32
	match                   int32 // UIT
	round                   int32 // UIT
	intro                   int32
	lastHitter              [2]int
	winTeam                 int // UIT
	winType                 [2]WinType
	winTrigger              [2]WinType // UIT
	matchWins, wins         [2]int32   // UIT
	roundsExisted           [2]int32
	draws                   int32
	tmode                   [2]TeamMode // UIT
	numSimul, numTurns      [2]int32    // UIT
	esc                     bool
	envcol_under            bool
	nextCharId              int32
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
	paused, frameStepFlag   bool
	roundResetFlg           bool
	reloadFlg               bool
	reloadStageFlg          bool
	reloadLifebarFlg        bool
	reloadCharSlot          [MaxPlayerNo]bool
	turbo                   float32
	drawScale               float32
	zoomlag                 float32
	zoomScale               float32
	zoomPosXLag             float32
	zoomPosYLag             float32
	enableZoomtime          int32
	zoomCameraBound         bool
	zoomPos                 [2]float32
	finishType              FinishType // UIT
	waitdown                int32
	slowtime                int32

	changeStateNest int32
	workpal         []uint32
	nomusic         bool
	keyConfig       []KeyConfig
	joystickConfig  []KeyConfig
	lifebar         Lifebar
	cgi             [MaxPlayerNo]CharGlobalInfo

	//accel                   float32
	//clsnDisplay             bool
	//debugDisplay            bool

	// New 11/04/2022 all UIT
	timerStart      int32
	timerRounds     []int32
	teamLeader      [2]int
	stage           *Stage
	postMatchFlg    bool
	scoreStart      [2]float32
	scoreRounds     [][2]float32
	decisiveRound   [2]bool
	sel             Select
	stringPool      [MaxPlayerNo]StringPool
	dialogueFlg     bool
	gameMode        string
	consecutiveWins [2]int32
	home            int

	// Non UIT, but adding them anyway just because
	// Used in Stage.go
	stageLoop bool

	// ByteCode
	dialogueBarsFlg bool
	dialogueForce   int
	playBgmFlg      bool

	// Input
	keyInput  glfw.Key
	keyString string

	// LifeBar
	timerCount []int32

	// Script
	endMatch    bool
	matchData   *lua.LTable
	noSoundFlg  bool
	continueFlg bool

	stageLoopNo int

	// 11/5/2022
	fight        Fight
	introSkipped bool
	preFightTime int32

	commandLists []*CommandList
	luaTables    []*lua.LTable

	loopBreak     bool
	loopContinue  bool
	brightnessOld int32
	wintime       int32

	// Rollback
	netTime int32
}

func NewGameState() *GameState {
	return &GameState{
		id: int(time.Now().UnixMilli()),
	}
}

func (gs *GameState) LoadState(stateID int) {
	// No state to load
	if gs == nil || !gs.saved {
		sys.appendToConsole(fmt.Sprintf("%v: No game state available for loading", sys.tickCount))
		return
	}

	if sys.rollback.session != nil {
		sys.rollback.session.netTime = gs.netTime
	}

	sys.arenaLoadMap[stateID] = arena.NewArena()
	a := sys.arenaLoadMap[stateID]
	gsp := &sys.loadPool

	sys.randseed = gs.randseed
	sys.gameTime = gs.gameTime
	sys.curRoundTime = gs.curRoundTime // UIT

	gs.loadCharData(a, gsp)
	gs.loadExplodData(a, gsp)
	sys.cam = gs.cam

	gs.loadPauseData()
	gs.loadSuperPauseData()

	gs.loadPalFX(a)
	gs.loadProjectileData(a, gsp)
	sys.aiLevel = gs.aiLevel
	sys.envShake = gs.envShake
	sys.envcol_time = gs.envcol_time
	sys.specialFlag = gs.specialFlag
	sys.envcol = gs.envcol

	sys.bcStack = arena.MakeSlice[BytecodeValue](a, len(gs.bcStack), len(gs.bcStack))
	copy(sys.bcStack, gs.bcStack)
	sys.bcVarStack = arena.MakeSlice[BytecodeValue](a, len(gs.bcVarStack), len(gs.bcVarStack))
	copy(sys.bcVarStack, gs.bcVarStack)
	sys.bcVar = arena.MakeSlice[BytecodeValue](a, len(gs.bcVar), len(gs.bcVar))
	copy(sys.bcVar, gs.bcVar)

	if sys.rollback.session != nil || sys.cfg.Netplay.Rollback.DesyncTestFrames > 0 {
		if sys.cfg.Netplay.Rollback.SaveStageData {
			sys.stage = gs.stage.Clone(a, gsp)
		}
	} else {
		sys.stage = gs.stage.Clone(a, gsp)
	}

	sys.aiInput = gs.aiInput
	sys.inputRemap = gs.inputRemap

	sys.workBe = arena.MakeSlice[BytecodeExp](a, len(gs.workBe), len(gs.workBe))
	for i := 0; i < len(gs.workBe); i++ {
		sys.workBe[i] = arena.MakeSlice[OpCode](a, len(gs.workBe[i]), len(gs.workBe[i]))
		copy(sys.workBe[i], gs.workBe[i])
	}

	sys.finishType = gs.finishType
	sys.winTeam = gs.winTeam
	sys.winType = gs.winType
	sys.winTrigger = gs.winTrigger
	sys.lastHitter = gs.lastHitter
	sys.waitdown = gs.waitdown
	sys.slowtime = gs.slowtime

	sys.winskipped = gs.winskipped

	sys.intro = gs.intro
	sys.nextCharId = gs.nextCharId

	sys.scrrect = gs.scrrect
	sys.gameWidth = gs.gameWidth
	sys.gameHeight = gs.gameHeight
	sys.widthScale = gs.widthScale
	sys.heightScale = gs.heightScale
	sys.gameEnd = gs.gameEnd
	sys.frameSkip = gs.frameSkip
	sys.brightness = gs.brightness
	sys.maxRoundTime = gs.maxRoundTime
	sys.turnsRecoveryRate = gs.turnsRecoveryRate

	sys.changeStateNest = gs.changeStateNest

	//sys.accel = gs.accel
	//sys.clsnDisplay = gs.clsnDisplay
	//sys.debugDisplay = gs.debugDisplay

	// Things that directly or indirectly get put into CGO can't go into arenas
	sys.workpal = make([]uint32, len(gs.workpal)) //arena.MakeSlice[uint32](a, len(gs.workpal), len(gs.workpal))
	copy(sys.workpal, gs.workpal)

	sys.nomusic = gs.nomusic

	sys.turbo = gs.turbo
	sys.drawScale = gs.drawScale
	sys.zoomlag = gs.zoomlag
	sys.zoomScale = gs.zoomScale
	sys.zoomPosXLag = gs.zoomPosXLag
	sys.zoomPosYLag = gs.zoomPosYLag
	sys.enableZoomtime = gs.enableZoomtime
	sys.zoomCameraBound = gs.zoomCameraBound
	sys.zoomPos = gs.zoomPos

	sys.reloadCharSlot = gs.reloadCharSlot
	sys.turbo = gs.turbo
	sys.drawScale = gs.drawScale
	sys.zoomlag = gs.zoomlag
	sys.zoomScale = gs.zoomScale
	sys.zoomPosXLag = gs.zoomPosXLag

	sys.matchWins = gs.matchWins
	sys.wins = gs.wins
	sys.roundsExisted = gs.roundsExisted
	sys.draws = gs.draws
	sys.tmode = gs.tmode
	sys.numSimul = gs.numSimul
	sys.numTurns = gs.numTurns
	sys.esc = gs.esc
	sys.envcol_under = gs.envcol_under
	sys.nextCharId = gs.nextCharId
	sys.tickCount = gs.tickCount
	sys.oldTickCount = gs.oldTickCount
	sys.tickCountF = gs.tickCountF
	sys.lastTick = gs.lastTick
	sys.nextAddTime = gs.nextAddTime
	sys.oldNextAddTime = gs.oldNextAddTime
	sys.screenleft = gs.screenleft
	sys.screenright = gs.screenright
	sys.xmin = gs.xmin
	sys.xmax = gs.xmax
	sys.winskipped = gs.winskipped
	sys.paused = gs.paused
	sys.frameStepFlag = gs.frameStepFlag
	sys.roundResetFlg = gs.roundResetFlg
	sys.reloadFlg = gs.reloadFlg
	sys.reloadStageFlg = gs.reloadStageFlg
	sys.reloadLifebarFlg = gs.reloadLifebarFlg

	sys.match = gs.match
	sys.round = gs.round

	sys.lifebar = gs.lifebar.Clone(a)

	sys.cgi = gs.cgi

	sys.timerStart = gs.timerStart

	sys.timerRounds = arena.MakeSlice[int32](a, len(gs.timerRounds), len(gs.timerRounds))
	copy(sys.timerRounds, gs.timerRounds)

	sys.teamLeader = gs.teamLeader
	sys.postMatchFlg = gs.postMatchFlg
	sys.scoreStart = gs.scoreStart

	sys.scoreRounds = arena.MakeSlice[[2]float32](a, len(gs.scoreRounds), len(gs.scoreRounds))
	copy(sys.scoreRounds, gs.scoreRounds)

	sys.decisiveRound = gs.decisiveRound

	sys.sel = gs.sel.Clone(a)
	for i := 0; i < len(sys.stringPool); i++ {
		sys.stringPool[i] = gs.stringPool[i].Clone(a, gsp)
	}

	sys.dialogueFlg = gs.dialogueFlg
	sys.gameMode = gs.gameMode
	sys.consecutiveWins = gs.consecutiveWins
	sys.home = gs.home

	// Not UIT
	sys.stageLoop = gs.stageLoop
	sys.dialogueBarsFlg = gs.dialogueBarsFlg
	sys.dialogueForce = gs.dialogueForce
	sys.playBgmFlg = gs.playBgmFlg
	//sys.keyState = gs.keyState
	sys.keyInput = gs.keyInput
	sys.keyString = gs.keyString

	sys.timerCount = arena.MakeSlice[int32](a, len(gs.timerCount), len(gs.timerCount))
	copy(sys.timerCount, gs.timerCount)

	sys.endMatch = gs.endMatch

	// theoretically this shouldn't do anything.
	sys.matchData = gs.cloneLuaTable(gs.matchData)

	sys.noSoundFlg = gs.noSoundFlg
	sys.continueFlg = gs.continueFlg
	sys.stageLoopNo = gs.stageLoopNo

	// gotta keep these pointers around because they are userdata
	for i := 0; i < len(sys.commandLists); i++ {
		gs.commandLists[i].CopyTo(sys.commandLists[i], a)
	}

	// sys.luaTables = gs.luaTables

	// This won't be around if we aren't in a proper rollback session.
	if sys.rollback.session != nil {
		sys.rollback.currentFight = gs.fight.Clone(a, gsp)
	}

	sys.introSkipped = gs.introSkipped

	sys.preFightTime = gs.preFightTime

	sys.loopBreak = gs.loopBreak
	sys.loopContinue = gs.loopContinue
	sys.brightnessOld = gs.brightnessOld

	sys.wintime = gs.wintime

	// Log state load
	if sys.rollback.session == nil {
		sys.appendToConsole(fmt.Sprintf("%v: Game state loaded", sys.tickCount))
	}
}

func (gs *GameState) SaveState(stateID int) {
	if sys.rollback.session != nil {
		gs.netTime = sys.rollback.session.netTime
	}

	sys.arenaSaveMap[stateID] = arena.NewArena()
	a := sys.arenaSaveMap[stateID]
	gsp := &sys.savePool

	gs.cgi = sys.cgi
	gs.saved = true
	gs.frame = sys.frameCounter

	gs.randseed = sys.randseed
	gs.gameTime = sys.gameTime
	gs.curRoundTime = sys.curRoundTime

	gs.saveCharData(a, gsp)
	gs.saveExplodData(a, gsp)
	gs.cam = sys.cam

	gs.savePauseData()
	gs.saveSuperPauseData()

	gs.savePalFX(a)
	gs.saveProjectileData(a, gsp)

	gs.aiLevel = sys.aiLevel
	gs.envShake = sys.envShake
	gs.envcol_time = sys.envcol_time
	gs.specialFlag = sys.specialFlag
	gs.envcol = sys.envcol

	gs.bcStack = arena.MakeSlice[BytecodeValue](a, len(sys.bcStack), len(sys.bcStack))
	copy(gs.bcStack, sys.bcStack)
	gs.bcVarStack = arena.MakeSlice[BytecodeValue](a, len(sys.bcVarStack), len(sys.bcVarStack))
	copy(gs.bcVarStack, sys.bcVarStack)
	gs.bcVar = arena.MakeSlice[BytecodeValue](a, len(sys.bcVar), len(sys.bcVar))
	copy(gs.bcVar, sys.bcVar)

	if sys.rollback.session != nil || sys.cfg.Netplay.Rollback.DesyncTestFrames > 0 {
		if sys.cfg.Netplay.Rollback.SaveStageData {
			gs.stage = sys.stage.Clone(a, gsp)
		}
	} else {
		gs.stage = sys.stage.Clone(a, gsp)
	}

	gs.aiInput = sys.aiInput
	gs.inputRemap = sys.inputRemap
	gs.workBe = arena.MakeSlice[BytecodeExp](a, len(sys.workBe), len(sys.workBe))
	for i := 0; i < len(sys.workBe); i++ {
		gs.workBe[i] = arena.MakeSlice[OpCode](a, len(sys.workBe[i]), len(sys.workBe[i]))
		copy(gs.workBe[i], sys.workBe[i])
	}

	gs.finishType = sys.finishType
	gs.winTeam = sys.winTeam
	gs.winType = sys.winType
	gs.winTrigger = sys.winTrigger
	gs.lastHitter = sys.lastHitter
	gs.waitdown = sys.waitdown
	gs.slowtime = sys.slowtime
	gs.winskipped = sys.winskipped
	gs.intro = sys.intro
	gs.nextCharId = sys.nextCharId

	gs.scrrect = sys.scrrect
	gs.gameWidth = sys.gameWidth
	gs.gameHeight = sys.gameHeight
	gs.widthScale = sys.widthScale
	gs.heightScale = sys.heightScale
	gs.gameEnd = sys.gameEnd
	gs.frameSkip = sys.frameSkip
	gs.brightness = sys.brightness
	gs.maxRoundTime = sys.maxRoundTime
	gs.turnsRecoveryRate = sys.turnsRecoveryRate

	gs.changeStateNest = sys.changeStateNest

	//gs.accel = sys.accel
	//gs.clsnDisplay = sys.clsnDisplay
	//gs.debugDisplay = sys.debugDisplay

	// Things that directly or indirectly get put into CGO can't go into arenas
	gs.workpal = make([]uint32, len(sys.workpal)) //arena.MakeSlice[uint32](a, len(sys.workpal), len(sys.workpal))
	copy(gs.workpal, sys.workpal)
	gs.nomusic = sys.nomusic

	gs.turbo = sys.turbo
	gs.drawScale = sys.drawScale
	gs.zoomlag = sys.zoomlag
	gs.zoomScale = sys.zoomScale
	gs.zoomPosXLag = sys.zoomPosXLag
	gs.zoomPosYLag = sys.zoomPosYLag
	gs.enableZoomtime = sys.enableZoomtime
	gs.zoomCameraBound = sys.zoomCameraBound
	gs.zoomPos = sys.zoomPos

	gs.reloadCharSlot = sys.reloadCharSlot
	gs.turbo = sys.turbo
	gs.drawScale = sys.drawScale
	gs.zoomlag = sys.zoomlag
	gs.zoomScale = sys.zoomScale
	gs.zoomPosXLag = sys.zoomPosXLag

	gs.matchWins = sys.matchWins
	gs.wins = sys.wins
	gs.roundsExisted = sys.roundsExisted
	gs.draws = sys.draws
	gs.tmode = sys.tmode
	gs.numSimul = sys.numSimul
	gs.numTurns = sys.numTurns
	gs.esc = sys.esc
	gs.envcol_under = sys.envcol_under
	gs.nextCharId = sys.nextCharId
	gs.tickCount = sys.tickCount
	gs.oldTickCount = sys.oldTickCount
	gs.tickCountF = sys.tickCountF
	gs.lastTick = sys.lastTick
	gs.nextAddTime = sys.nextAddTime
	gs.oldNextAddTime = sys.oldNextAddTime
	gs.screenleft = sys.screenleft
	gs.screenright = sys.screenright
	gs.xmin = sys.xmin
	gs.xmax = sys.xmax
	gs.winskipped = sys.winskipped
	gs.paused = sys.paused
	gs.frameStepFlag = sys.frameStepFlag
	gs.roundResetFlg = sys.roundResetFlg
	gs.reloadFlg = sys.reloadFlg
	gs.reloadStageFlg = sys.reloadStageFlg
	gs.reloadLifebarFlg = sys.reloadLifebarFlg

	gs.match = sys.match
	gs.round = sys.round

	gs.lifebar = sys.lifebar.Clone(a)

	gs.timerStart = sys.timerStart
	gs.timerRounds = arena.MakeSlice[int32](a, len(sys.timerRounds), len(sys.timerRounds))
	copy(gs.timerRounds, sys.timerRounds)
	gs.teamLeader = sys.teamLeader
	gs.postMatchFlg = sys.postMatchFlg
	gs.scoreStart = sys.scoreStart
	gs.scoreRounds = arena.MakeSlice[[2]float32](a, len(sys.scoreRounds), len(sys.scoreRounds))
	copy(gs.scoreRounds, sys.scoreRounds)
	gs.decisiveRound = sys.decisiveRound
	gs.sel = sys.sel.Clone(a)
	for i := 0; i < len(sys.stringPool); i++ {
		gs.stringPool[i] = sys.stringPool[i].Clone(a, gsp)
	}

	gs.dialogueFlg = sys.dialogueFlg
	gs.gameMode = sys.gameMode
	gs.consecutiveWins = sys.consecutiveWins

	gs.stageLoop = sys.stageLoop
	gs.dialogueBarsFlg = sys.dialogueBarsFlg
	gs.dialogueForce = sys.dialogueForce
	gs.playBgmFlg = sys.playBgmFlg

	gs.keyInput = sys.keyInput
	gs.keyString = sys.keyString

	gs.timerCount = arena.MakeSlice[int32](a, len(sys.timerCount), len(sys.timerCount))
	copy(gs.timerCount, sys.timerCount)

	gs.endMatch = sys.endMatch

	// can't deep copy because its members are private
	//matchData := *sys.matchData
	gs.matchData = gs.cloneLuaTable(sys.matchData)

	gs.noSoundFlg = sys.noSoundFlg
	gs.continueFlg = sys.continueFlg
	gs.stageLoopNo = sys.stageLoopNo

	gs.commandLists = arena.MakeSlice[*CommandList](a, len(sys.commandLists), len(sys.commandLists))
	for i := 0; i < len(sys.commandLists); i++ {
		cl := sys.commandLists[i].Clone(a)
		gs.commandLists[i] = &cl
	}
	gs.luaTables = arena.MakeSlice[*lua.LTable](a, len(sys.luaTables), len(sys.luaTables))
	for i := 0; i < len(sys.luaTables); i++ {
		gs.luaTables[i] = gs.cloneLuaTable(sys.luaTables[i])
	}

	// This won't be around if we aren't in a proper rollback session.
	if sys.rollback.session != nil {
		gs.fight = sys.rollback.currentFight.Clone(a, gsp)
	}

	gs.introSkipped = sys.introSkipped
	gs.preFightTime = sys.preFightTime

	gs.loopBreak = sys.loopBreak
	gs.loopContinue = sys.loopContinue
	gs.brightnessOld = sys.brightnessOld

	gs.wintime = sys.wintime

	// Log save state
	if sys.rollback.session == nil {
		sys.appendToConsole(fmt.Sprintf("%v: Game state saved", sys.tickCount))
	}
}

func (gs *GameState) cloneLuaTable(s *lua.LTable) *lua.LTable {
	tbl := sys.luaLState.NewTable()
	s.ForEach(func(key lua.LValue, value lua.LValue) {
		switch value.Type() {
		case lua.LTTable:
			innerTbl := value.(*lua.LTable)
			tbl.RawSet(key, gs.cloneLuaTable(innerTbl))
		default:
			tbl.RawSet(key, value)
		}
	})
	return tbl
}

func (src *CommandList) CopyTo(dst *CommandList, a *arena.Arena) {
	clone := src.Clone(a)
	*dst = clone
}

func (gs *GameState) savePalFX(a *arena.Arena) {
	gs.allPalFX = sys.allPalFX.Clone(a)
	gs.bgPalFX = sys.bgPalFX.Clone(a)
}

func (gs *GameState) saveCharData(a *arena.Arena, gsp *GameStatePool) {
	for i := range sys.chars {
		gs.charData[i] = arena.MakeSlice[Char](a, len(sys.chars[i]), len(sys.chars[i]))
		gs.chars[i] = arena.MakeSlice[*Char](a, len(sys.chars[i]), len(sys.chars[i]))

		for j, c := range sys.chars[i] {
			gs.charData[i][j] = c.Clone(a, gsp)
			gs.chars[i][j] = c
		}
	}

	// Update command sharing for chars without keyctrl
	for i := range gs.chars {
		for _, c := range gs.chars[i] {
			if !c.keyctrl[0] {
				c.cmd = gs.chars[c.playerNo][0].cmd
			}
		}
	}

	// Clone charList
	gs.charList = sys.charList.Clone(a, gsp)
}

func (gs *GameState) saveProjectileData(a *arena.Arena, gsp *GameStatePool) {
	for i := range sys.projs {
		gs.projs[i] = arena.MakeSlice[*Projectile](a, len(sys.projs[i]), len(sys.projs[i]))
		for j := 0; j < len(sys.projs[i]); j++ {
			gs.projs[i][j] = sys.projs[i][j].clone(a, gsp)
		}
	}
}

func (gs *GameState) savePauseData() {
	gs.pausetimebuffer = sys.pausetimebuffer
	gs.pausetime = sys.pausetime
	gs.pausebg = sys.pausebg
	gs.pauseendcmdbuftime = sys.pauseendcmdbuftime
	gs.pauseplayerno = sys.pauseplayerno
}

func (gs *GameState) saveSuperPauseData() {
	gs.supertimebuffer = sys.supertimebuffer
	gs.supertime = sys.supertime
	gs.superpausebg = sys.superpausebg
	gs.superendcmdbuftime = sys.superendcmdbuftime
	gs.superplayerno = sys.superplayerno
	gs.superdarken = sys.superdarken
}

func (gs *GameState) saveExplodData(a *arena.Arena, gsp *GameStatePool) {
	for i := range sys.explods {
		gs.explods[i] = arena.MakeSlice[*Explod](a, len(sys.explods[i]), len(sys.explods[i]))
		for j := 0; j < len(sys.explods[i]); j++ {
			gs.explods[i][j] = sys.explods[i][j].Clone(a, gsp)
		}
	}
}

func (gs *GameState) loadPalFX(a *arena.Arena) {
	sys.allPalFX = gs.allPalFX.Clone(a)
	sys.bgPalFX = gs.bgPalFX.Clone(a)
}

func (gs *GameState) loadCharData(a *arena.Arena, gsp *GameStatePool) {
	for i := 0; i < len(sys.chars); i++ {
		sys.chars[i] = arena.MakeSlice[*Char](a, len(gs.chars[i]), len(gs.chars[i]))
		copy(sys.chars[i], gs.chars[i])
	}

	for i := 0; i < len(sys.chars); i++ {
		for j := 0; j < len(sys.chars[i]); j++ {
			*sys.chars[i][j] = gs.charData[i][j].Clone(a, gsp)
		}
	}

	for i := range sys.chars {
		for _, c := range sys.chars[i] {
			if !c.keyctrl[0] {
				c.cmd = sys.chars[c.playerNo][0].cmd
			}
		}
	}

	// Set workingChar and debugWC to the first char we find, just in case
	if c := sys.anyChar(); c != nil {
		sys.workingChar = c
		sys.workingState = &c.ss.sb
		sys.debugWC = c
	}

	sys.charList = gs.charList.Clone(a, gsp)
}

func (gs *GameState) loadSuperPauseData() {
	sys.supertimebuffer = gs.supertimebuffer
	sys.supertime = gs.supertime
	sys.superpausebg = gs.superpausebg
	sys.superendcmdbuftime = gs.superendcmdbuftime
	sys.superplayerno = gs.superplayerno
	sys.superdarken = gs.superdarken
}

func (gs *GameState) loadPauseData() {
	sys.pausetimebuffer = gs.pausetimebuffer
	sys.pausetime = gs.pausetime
	sys.pausebg = gs.pausebg
	sys.pauseendcmdbuftime = gs.pauseendcmdbuftime
	sys.pauseplayerno = gs.pauseplayerno
}

func (gs *GameState) loadExplodData(a *arena.Arena, gsp *GameStatePool) {
	for i := range gs.explods {
		sys.explods[i] = arena.MakeSlice[*Explod](a, len(gs.explods[i]), len(gs.explods[i]))
		for j := 0; j < len(gs.explods[i]); j++ {
			sys.explods[i][j] = gs.explods[i][j].Clone(a, gsp)
		}
	}
}

func (gs *GameState) loadProjectileData(a *arena.Arena, gsp *GameStatePool) {
	for i := range gs.projs {
		sys.projs[i] = arena.MakeSlice[*Projectile](a, len(gs.projs[i]), len(gs.projs[i]))
		for j := range gs.projs[i] {
			sys.projs[i][j] = gs.projs[i][j].clone(a, gsp)
		}
	}
}

func (gsp *GameStatePool) Get(item interface{}) (result interface{}) {
	objs, ok := gsp.poolObjs[gsp.curStateID]
	if !ok {
		gsp.poolObjs[gsp.curStateID] = make([]interface{}, 0, 50)
		objs = gsp.poolObjs[gsp.curStateID]
	}

	switch item.(type) {
	case (map[string]float32):
		objs = append(objs, gsp.stringFloat32MapPool.Get())
		return objs[len(objs)-1]
	case (map[string]int):
		objs = append(objs, gsp.stringIntMapPool.Get())
		return objs[len(objs)-1]
	case (AnimationTable):
		objs = append(objs, gsp.animationTablePool.Get())
		return objs[len(objs)-1]
	case (map[int32]*Char):
		objs = append(objs, gsp.int32CharPointerMapPool.Get())
		return objs[len(objs)-1]
	case ([]AnimFrame):
		objs = append(objs, gsp.animFrameSlicePool.Get())
		return objs[len(objs)-1]
	case (map[int32]int32):
		objs = append(objs, gsp.int32int32MapPool.Get())
		return objs[len(objs)-1]
	case (map[int32]float32):
		objs = append(objs, gsp.int32float32MapPool.Get())
		return objs[len(objs)-1]
	default:
		return nil
	}
}

func (gsp *GameStatePool) Put(item interface{}) {
	switch item.(type) {
	case (*map[string]float32):
		gsp.stringFloat32MapPool.Put(item)
	case (*map[string]int):
		gsp.stringIntMapPool.Put(item)
	case (*AnimationTable):
		gsp.animationTablePool.Put(item)
	case (*map[int32]*Char):
		gsp.int32CharPointerMapPool.Put(item)
	case (*[]AnimFrame):
		gsp.animFrameSlicePool.Put(item)
	case (*map[int32]int32):
		gsp.int32int32MapPool.Put(item)
	case (*map[int32]float32):
		gsp.int32float32MapPool.Put(item)
	default:
	}
}

func (gsp *GameStatePool) Free(stateID int) {
	objs, ok := gsp.poolObjs[stateID]
	if ok {
		for i := 0; i < len(objs); i++ {
			gsp.Put(objs[i])
		}
	}
	delete(gsp.poolObjs, stateID)
}

func NewGameStatePool() GameStatePool {
	return GameStatePool{
		gameStatePool: sync.Pool{
			New: func() interface{} {
				return NewGameState()
			},
		},
		stringIntMapPool: sync.Pool{
			New: func() interface{} {
				si := make(map[string]int)
				return &si
			},
		},
		stringFloat32MapPool: sync.Pool{
			New: func() interface{} {
				sf := make(map[string]float32)
				return &sf
			},
		},
		animationTablePool: sync.Pool{
			New: func() interface{} {
				at := make(AnimationTable)
				return &at
			},
		},
		int32CharPointerMapPool: sync.Pool{
			New: func() interface{} {
				ic := make(map[int32]*Char)
				return &ic
			},
		},
		animFrameSlicePool: sync.Pool{
			New: func() interface{} {
				af := make([]AnimFrame, 0, 8)
				return &af
			},
		},
		int32int32MapPool: sync.Pool{
			New: func() interface{} {
				ii := make(map[int32]int32)
				return &ii
			},
		},
		int32float32MapPool: sync.Pool{
			New: func() interface{} {
				if3 := make(map[int32]float32)
				return &if3
			},
		},
		poolObjs: make(map[int][]interface{}),
	}
}

type GameStatePool struct {
	gameStatePool           sync.Pool
	stringIntMapPool        sync.Pool
	hitscaleMapPool         sync.Pool
	stringFloat32MapPool    sync.Pool
	animationTablePool      sync.Pool
	mapArraySlicePool       sync.Pool
	int32CharPointerMapPool sync.Pool
	int32int32MapPool       sync.Pool
	int32float32MapPool     sync.Pool

	animFrameSlicePool sync.Pool
	poolObjs           map[int][]interface{}
	curStateID         int
}

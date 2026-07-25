package main

import (
	"arena"
	"fmt"
	"hash/fnv"
	"strconv"
	"sync"
	"time"
)

const MaxSaveStates = 8

type GameState struct {
	// Identifiers
	bytes              []byte
	id                 int
	saved              bool
	frame              int32
	isSpeculativeFrame bool

	SystemStateVars

	chars     [MaxPlayerNo][]*Char
	charData  [MaxPlayerNo][]Char
	projs     [MaxPlayerNo][]*Projectile
	explods   [MaxPlayerNo][]*Explod
	chartexts [MaxPlayerNo][]*TextSprite
	charList  CharList

	allPalFX *PalFX
	bgPalFX  *PalFX

	bcStack, bcVarStack BytecodeStack
	bcVar               []BytecodeValue
	workBe              []BytecodeExp

	workpal        []uint32
	keyConfig      []KeyConfig
	joystickConfig []KeyConfig
	fightScreen    FightScreen
	motif          Motif
	storyboard     Storyboard
	cgi            [MaxPlayerNo]CharGlobalInfo

	//accel                   float32
	//clsnDisplay             bool
	//debugDisplay            bool

	timerRounds []int32
	stageRef    *Stage
	stage       *Stage
	stageList   map[int32]*Stage
	stageState  stageRollbackState
	statsState  statsRollbackState
	scoreRounds [][2]float32
	sel         Select
	//stringPool      [MaxPlayerNo]StringPool // Only mutated while compiling
	dialogueFlg bool

	// FightScreen
	timerCount []int32

	commandLists  []*CommandList
	matchMusicSel []*bgMusic

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
		// Replay recording follows the rollback timeline.
		// Any frames from the abandoned speculative future must be discarded before restoring the frame cursor.
		sys.rollback.session.TruncateReplayFrom(gs.netTime)
		sys.rollback.session.netTime = gs.netTime
	}

	sys.arenaLoadMap[stateID] = arena.NewArena()
	a := sys.arenaLoadMap[stateID]
	gsp := &sys.loadPool

	sys.SystemStateVars = gs.SystemStateVars
	sys.frameCounter = gs.frame
	sys.matchMusicSel = arena.MakeSlice[*bgMusic](a, len(gs.matchMusicSel), len(gs.matchMusicSel))
	copy(sys.matchMusicSel, gs.matchMusicSel)

	gs.loadCharData(a, gsp)
	gs.loadProjectileData(a, gsp)
	gs.loadExplodData(a, gsp)
	gs.loadCharTextData(a)
	gs.loadPalFX(a)

	sys.bcStack = arena.MakeSlice[BytecodeValue](a, len(gs.bcStack), len(gs.bcStack))
	copy(sys.bcStack, gs.bcStack)
	sys.bcVarStack = arena.MakeSlice[BytecodeValue](a, len(gs.bcVarStack), len(gs.bcVarStack))
	copy(sys.bcVarStack, gs.bcVarStack)
	sys.bcVar = arena.MakeSlice[BytecodeValue](a, len(gs.bcVar), len(gs.bcVar))
	copy(sys.bcVar, gs.bcVar)

	// Load the full stage only when character code can mutate its definition.
	// Otherwise restore the small gameplay-visible runtime subset.
	if gs.stage != nil {
		sys.stageList, sys.stage = cloneStageList(a, gsp, gs.stageList, gs.stage)
	} else {
		// A speculative round transition may have switched sys.stage to a different roundXdef entry.
		// Restore the saved stage identity before applying its lightweight runtime snapshot.
		sys.stage = gs.stageRef
		gs.stageState.Load(sys.stage)
	}

	sys.workBe = arena.MakeSlice[BytecodeExp](a, len(gs.workBe), len(gs.workBe))
	for i := 0; i < len(gs.workBe); i++ {
		sys.workBe[i] = arena.MakeSlice[OpCode](a, len(gs.workBe[i]), len(gs.workBe[i]))
		copy(sys.workBe[i], gs.workBe[i])
	}

	//sys.accel = gs.accel
	//sys.clsnDisplay = gs.clsnDisplay
	//sys.debugDisplay = gs.debugDisplay

	// Things that directly or indirectly get put into CGO can't go into arenas
	sys.workpal = make([]uint32, len(gs.workpal)) //arena.MakeSlice[uint32](a, len(gs.workpal), len(gs.workpal))
	copy(sys.workpal, gs.workpal)

	sys.fightScreen = gs.fightScreen.Clone(a)
	sys.motif = gs.motif.Clone(a, gs.postMatchFlg)

	// Storyboard: only rollback-touch it when it was actually running.
	if gs.storyboard.active {
		sys.storyboard = gs.storyboard.Clone(a)
	} else {
		// If storyboard started after the save point, prevent it from continuing after rollback.
		sys.storyboard.active = false
		sys.storyboard.initialized = false
		sys.storyboard.dialogueLayers = nil
		sys.storyboard.dialoguePos = 0
	}

	sys.cgi = gs.cgi

	sys.timerRounds = arena.MakeSlice[int32](a, len(gs.timerRounds), len(gs.timerRounds))
	copy(sys.timerRounds, gs.timerRounds)

	sys.scoreRounds = arena.MakeSlice[[2]float32](a, len(gs.scoreRounds), len(gs.scoreRounds))
	copy(sys.scoreRounds, gs.scoreRounds)
	gs.statsState.load(&sys.statsLog)

	//sys.sel = gs.sel.Clone(a)
	// for i := 0; i < len(sys.stringPool); i++ {
	// 	sys.stringPool[i] = gs.stringPool[i].Clone(a, gsp)
	// }

	sys.motif.di.active = gs.dialogueFlg

	sys.timerCount = arena.MakeSlice[int32](a, len(gs.timerCount), len(gs.timerCount))
	copy(sys.timerCount, gs.timerCount)

	// gotta keep these pointers around because they are userdata
	for i := 0; i < len(sys.commandLists); i++ {
		gs.commandLists[i].CopyTo(sys.commandLists[i], a)
	}

	// Stop all sounds if they started playing after the point of the save state
	for i := range sys.soundChannels {
		ch := &sys.soundChannels[i]
		if ch.IsPlaying() && ch.timeStamp >= sys.gameTime() {
			ch.Reset()
		}
	}
	for i := range sys.charSoundChannels {
		for j := range sys.charSoundChannels[i] {
			ch := &sys.charSoundChannels[i][j]
			if ch.IsPlaying() && ch.timeStamp >= sys.gameTime() {
				ch.Reset()
			}
		}
	}

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
	gs.isSpeculativeFrame = sys.isSpeculativeFrame()
	gs.SystemStateVars = sys.SystemStateVars
	gs.matchMusicSel = arena.MakeSlice[*bgMusic](a, len(sys.matchMusicSel), len(sys.matchMusicSel))
	copy(gs.matchMusicSel, sys.matchMusicSel)

	gs.saveCharData(a, gsp)
	gs.saveProjectileData(a, gsp)
	gs.saveExplodData(a, gsp)
	gs.saveCharTextData(a)
	gs.savePalFX(a)

	gs.bcStack = arena.MakeSlice[BytecodeValue](a, len(sys.bcStack), len(sys.bcStack))
	copy(gs.bcStack, sys.bcStack)
	gs.bcVarStack = arena.MakeSlice[BytecodeValue](a, len(sys.bcVarStack), len(sys.bcVarStack))
	copy(gs.bcVarStack, sys.bcVarStack)
	gs.bcVar = arena.MakeSlice[BytecodeValue](a, len(sys.bcVar), len(sys.bcVar))
	copy(gs.bcVar, sys.bcVar)

	// A pooled GameState may still contain stage data from an older save.
	gs.stageRef = sys.stage
	gs.stage = nil
	gs.stageList = nil
	gs.stageState = stageRollbackState{}

	// Character-controlled stage mutations need the full clone.
	if sys.rollback.session != nil || sys.cfg.Netplay.Rollback.DesyncTestFrames > 0 {
		if gs.stageCanMutate() || sys.cfg.Netplay.Rollback.SaveStageData {
			gs.stageList, gs.stage = cloneStageList(a, gsp, sys.stageList, sys.stage)
		} else {
			gs.stageState = sys.stage.CloneRollbackState(a)
		}
	} else {
		// Save anyway if using debug keys
		gs.stageList, gs.stage = cloneStageList(a, gsp, sys.stageList, sys.stage)
	}

	gs.workBe = arena.MakeSlice[BytecodeExp](a, len(sys.workBe), len(sys.workBe))
	for i := 0; i < len(sys.workBe); i++ {
		gs.workBe[i] = arena.MakeSlice[OpCode](a, len(sys.workBe[i]), len(sys.workBe[i]))
		copy(gs.workBe[i], sys.workBe[i])
	}

	//gs.accel = sys.accel
	//gs.clsnDisplay = sys.clsnDisplay
	//gs.debugDisplay = sys.debugDisplay

	// Things that directly or indirectly get put into CGO can't go into arenas
	gs.workpal = make([]uint32, len(sys.workpal)) //arena.MakeSlice[uint32](a, len(sys.workpal), len(sys.workpal))
	copy(gs.workpal, sys.workpal)

	gs.fightScreen = sys.fightScreen.Clone(a)
	gs.motif = sys.motif.Clone(a, sys.postMatchFlg)

	// Storyboard: only rollback-save while active.
	if sys.storyboard.active {
		gs.storyboard = sys.storyboard.Clone(a)
	} else {
		gs.storyboard = Storyboard{}
		gs.storyboard.active = false
	}

	gs.timerRounds = arena.MakeSlice[int32](a, len(sys.timerRounds), len(sys.timerRounds))
	copy(gs.timerRounds, sys.timerRounds)
	gs.scoreRounds = arena.MakeSlice[[2]float32](a, len(sys.scoreRounds), len(sys.scoreRounds))
	copy(gs.scoreRounds, sys.scoreRounds)
	gs.statsState = sys.statsLog.saveRollbackState()

	//gs.sel = sys.sel.Clone(a)
	// for i := 0; i < len(sys.stringPool); i++ {
	//		gs.stringPool[i] = sys.stringPool[i].Clone(a, gsp)
	// }

	gs.dialogueFlg = sys.motif.di.active

	gs.timerCount = arena.MakeSlice[int32](a, len(sys.timerCount), len(sys.timerCount))
	copy(gs.timerCount, sys.timerCount)

	gs.commandLists = arena.MakeSlice[*CommandList](a, len(sys.commandLists), len(sys.commandLists))
	for i := 0; i < len(sys.commandLists); i++ {
		cl := sys.commandLists[i].Clone(a)
		gs.commandLists[i] = &cl
	}

	// Log save state
	if sys.rollback.session == nil {
		sys.appendToConsole(fmt.Sprintf("%v: Game state saved", sys.tickCount))
	}
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

func (gs *GameState) saveExplodData(a *arena.Arena, gsp *GameStatePool) {
	for i := range sys.explods {
		gs.explods[i] = arena.MakeSlice[*Explod](a, len(sys.explods[i]), len(sys.explods[i]))
		for j := 0; j < len(sys.explods[i]); j++ {
			gs.explods[i][j] = sys.explods[i][j].Clone(a, gsp)
		}
	}
}

func (gs *GameState) saveCharTextData(a *arena.Arena) {
	for i := range sys.chartexts {
		gs.chartexts[i] = arena.MakeSlice[*TextSprite](a, len(sys.chartexts[i]), len(sys.chartexts[i]))
		for j := range sys.chartexts[i] {
			gs.chartexts[i][j] = cloneTextSprite(a, sys.chartexts[i][j])
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

func (gs *GameState) loadProjectileData(a *arena.Arena, gsp *GameStatePool) {
	for i := range gs.projs {
		sys.projs[i] = arena.MakeSlice[*Projectile](a, len(gs.projs[i]), len(gs.projs[i]))
		for j := range gs.projs[i] {
			sys.projs[i][j] = gs.projs[i][j].clone(a, gsp)
		}
	}
}

func (gs *GameState) loadExplodData(a *arena.Arena, gsp *GameStatePool) {
	for i := range gs.explods {
		sys.explods[i] = arena.MakeSlice[*Explod](a, len(gs.explods[i]), len(gs.explods[i]))
		for j := 0; j < len(gs.explods[i]); j++ {
			sys.explods[i][j] = gs.explods[i][j].Clone(a, gsp)
		}
	}
}

func (gs *GameState) loadCharTextData(a *arena.Arena) {
	for i := range gs.chartexts {
		sys.chartexts[i] = arena.MakeSlice[*TextSprite](a, len(gs.chartexts[i]), len(gs.chartexts[i]))
		for j := range gs.chartexts[i] {
			sys.chartexts[i][j] = cloneTextSprite(a, gs.chartexts[i][j])
		}
	}
}

func (gs *GameState) stageCanMutate() bool {
	for i := range sys.cgi {
		if sys.cgi[i].canMutateStage {
			return true
		}
	}
	return false
}

func (gs *GameState) getID() string {
	return strconv.Itoa(int(gs.id))
}

// Not to be confused with the live checksum. This one's for debugging
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

// Returns some state variables as a string for debugging
func (gs *GameState) String() (str string) {
	// Add match data
	str = fmt.Sprintf("MatchTime %d CurRoundTime: %d ScorePoints: %v ComboCount: %v\n",
		gs.matchTime, gs.curRoundTime, gs.scorePoints, gs.comboCount)

	// Add bytecode data
	// TODO: Every log seems to have these empty. May not be needed
	str += fmt.Sprintf("bcStack: %v\n", gs.bcStack)
	str += fmt.Sprintf("bcVarStack: %v\n", gs.bcVarStack)
	str += fmt.Sprintf("bcVar: %v\n", gs.bcVar)
	str += fmt.Sprintf("workBe: %v\n", gs.workBe)

	// Add char data
	for i := 0; i < len(gs.charData); i++ {
		for j := 0; j < len(gs.charData[i]); j++ {
			str += gs.charData[i][j].String()
			str += "\n"
		}
	}

	return
}

// Returns char status as a string for debugging
func (cs Char) String() string {
	// Save button states if char has keyctrl
	inputBufStr := "none"
	if cs.keyctrl[0] && len(cs.cmd) > 0 && cs.cmd[0].Buffer != nil {
		ib := cs.cmd[0].Buffer
		inputBufStr = fmt.Sprintf(
			"U:%d D:%d L:%d R:%d B:%d F:%d N:%d a:%d b:%d c:%d x:%d y:%d z:%d s:%d d:%d w:%d m:%d",
			ib.Ub, ib.Db, ib.Lb, ib.Rb, ib.Bb, ib.Fb, ib.Nb,
			ib.ab, ib.bb, ib.cb, ib.xb, ib.yb, ib.zb,
			ib.sb, ib.db, ib.wb, ib.mb,
		)
	}

	str := fmt.Sprintf(`Char %s
	Controller          :%d
	PlayerNo            :%d
	HelperIndex         :%d
	Life                :%d
	RedLife             :%d
	DizzyPoints         :%d
	GuardPoints         :%d
	Power               :%d
	Localcoord          :%f
	Localscl            :%f
	Pos                 :%v
	Vel                 :%v
	Facing              :%f
	Id                  :%d
	HelperId            :%d
	ParentId            :%d
	StateNo             :%d
	StateTime           :%d
	AnimNo              :%d
	Mctime              :%d
	Targets             :%v
	Preserve            :%t
	MapsActive          :%d
	CnsVar              :%v
	CnsFvar             :%v
	InputBuffer         :%s`,
		cs.name, cs.controller, cs.playerNo, cs.helperIndex,
		cs.life, cs.redLife, cs.dizzyPoints, cs.guardPoints, cs.power,
		cs.localcoord, cs.localscl,
		cs.pos, cs.vel, cs.facing,
		cs.id, cs.helperId, cs.parentId,
		cs.ss.no, cs.ss.time, cs.animNo, // Move/Statetype would require interpreting the flags so they're not worth it
		cs.mctime, cs.targets,
		cs.preserve,
		len(cs.mapArray), cs.cnsvar, cs.cnsfvar, inputBufStr) // Dumping entire map is too verbose so we'll just log how many are active

	return str
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
	remapPresetPool         sync.Pool
	remapTablePool          sync.Pool

	animFrameSlicePool sync.Pool
	poolObjs           map[int][]interface{}
	curStateID         int
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
				at := AnimationTable{
					anims: make(map[int32]*Animation),
				}
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
		remapPresetPool: sync.Pool{
			New: func() interface{} {
				rp := make(RemapPreset)
				return &rp
			},
		},
		remapTablePool: sync.Pool{
			New: func() interface{} {
				rt := make(RemapTable)
				return &rt
			},
		},
		poolObjs: make(map[int][]interface{}),
	}
}

func (gsp *GameStatePool) Get(item interface{}) (result interface{}) {
	stateID := gsp.curStateID
	objs, ok := gsp.poolObjs[stateID]
	if !ok {
		gsp.poolObjs[stateID] = make([]interface{}, 0, 50)
		objs = gsp.poolObjs[stateID]
	}

	// A map stores a copy of a slice header, so appending only to the local variable would leave the map entry at its old length.
	defer func() {
		gsp.poolObjs[stateID] = objs
	}()

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
	case (RemapPreset):
		objs = append(objs, gsp.remapPresetPool.Get())
		return objs[len(objs)-1]
	case (RemapTable):
		objs = append(objs, gsp.remapTablePool.Get())
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
	case (*RemapPreset):
		gsp.remapPresetPool.Put(item)
	case (*RemapTable):
		gsp.remapTablePool.Put(item)
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

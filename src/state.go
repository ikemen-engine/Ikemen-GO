package main

import (
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

	workpal          []uint32
	keyConfig        []KeyConfig
	joystickConfig   []KeyConfig
	fightScreen      FightScreen
	motif            Motif
	storyboard       Storyboard
	matchStoryboards []*Storyboard
	cgi              [MaxPlayerNo]CharGlobalInfo

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

	gsp := &sys.loadPool

	sys.SystemStateVars = gs.SystemStateVars
	sys.frameCounter = gs.frame
	sys.matchMusicSel = make([]*bgMusic, len(gs.matchMusicSel), len(gs.matchMusicSel))
	copy(sys.matchMusicSel, gs.matchMusicSel)

	gs.loadCharData(gsp)
	gs.loadProjectileData(gsp)
	gs.loadExplodData(gsp)
	gs.loadCharTextData()
	gs.loadPalFX()

	sys.bcStack = make([]BytecodeValue, len(gs.bcStack), len(gs.bcStack))
	copy(sys.bcStack, gs.bcStack)
	sys.bcVarStack = make([]BytecodeValue, len(gs.bcVarStack), len(gs.bcVarStack))
	copy(sys.bcVarStack, gs.bcVarStack)
	sys.bcVar = make([]BytecodeValue, len(gs.bcVar), len(gs.bcVar))
	copy(sys.bcVar, gs.bcVar)

	// Load the full stage only when character code can mutate its definition.
	// Otherwise restore the small gameplay-visible runtime subset.
	if gs.stage != nil {
		sys.stageList, sys.stage = cloneStageList(gsp, gs.stageList, gs.stage)
	} else {
		// A speculative round transition may have switched sys.stage to a different roundXdef entry.
		// Restore the saved stage identity before applying its lightweight runtime snapshot.
		sys.stage = gs.stageRef
		gs.stageState.Load(sys.stage)
	}

	sys.workBe = make([]BytecodeExp, len(gs.workBe), len(gs.workBe))
	for i := 0; i < len(gs.workBe); i++ {
		sys.workBe[i] = make([]OpCode, len(gs.workBe[i]), len(gs.workBe[i]))
		copy(sys.workBe[i], gs.workBe[i])
	}

	//sys.accel = gs.accel
	//sys.clsnDisplay = gs.clsnDisplay
	//sys.debugDisplay = gs.debugDisplay

	// Bound to CGO; kept as an explicit allocation.
	sys.workpal = make([]uint32, len(gs.workpal))
	copy(sys.workpal, gs.workpal)

	sys.fightScreen = gs.fightScreen.Clone()
	sys.motif = gs.motif.Clone(gs.postMatchFlg)

	// Storyboard: only rollback-touch it when it was actually running.
	if gs.storyboard.active {
		sys.storyboard = gs.storyboard.Clone()
	} else {
		// If storyboard started after the save point, prevent it from continuing after rollback.
		sys.storyboard.active = false
		sys.storyboard.initialized = false
		sys.storyboard.dialogueLayers = nil
		sys.storyboard.dialoguePos = 0
	}
	sys.matchStoryboards.clear(-1)
	// Restore playback state without replacing the retained decoder set.
	sys.matchStoryboards.instances = cloneStoryboards(gs.matchStoryboards)

	sys.cgi = gs.cgi

	sys.timerRounds = make([]int32, len(gs.timerRounds), len(gs.timerRounds))
	copy(sys.timerRounds, gs.timerRounds)

	sys.scoreRounds = make([][2]float32, len(gs.scoreRounds), len(gs.scoreRounds))
	copy(sys.scoreRounds, gs.scoreRounds)
	gs.statsState.load(&sys.statsLog)

	//sys.sel = gs.sel.Clone()
	// for i := 0; i < len(sys.stringPool); i++ {
	// 	sys.stringPool[i] = gs.stringPool[i].Clone(gsp)
	// }

	sys.motif.di.active = gs.dialogueFlg

	sys.timerCount = make([]int32, len(gs.timerCount), len(gs.timerCount))
	copy(sys.timerCount, gs.timerCount)

	// gotta keep these pointers around because they are userdata
	for i := 0; i < len(sys.commandLists); i++ {
		gs.commandLists[i].CopyTo(sys.commandLists[i])
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

	gsp := &sys.savePool

	gs.cgi = sys.cgi
	gs.saved = true
	gs.frame = sys.frameCounter
	gs.isSpeculativeFrame = sys.isSpeculativeFrame()
	gs.SystemStateVars = sys.SystemStateVars
	gs.matchMusicSel = make([]*bgMusic, len(sys.matchMusicSel), len(sys.matchMusicSel))
	copy(gs.matchMusicSel, sys.matchMusicSel)

	gs.saveCharData(gsp)
	gs.saveProjectileData(gsp)
	gs.saveExplodData(gsp)
	gs.saveCharTextData()
	gs.savePalFX()

	gs.bcStack = make([]BytecodeValue, len(sys.bcStack), len(sys.bcStack))
	copy(gs.bcStack, sys.bcStack)
	gs.bcVarStack = make([]BytecodeValue, len(sys.bcVarStack), len(sys.bcVarStack))
	copy(gs.bcVarStack, sys.bcVarStack)
	gs.bcVar = make([]BytecodeValue, len(sys.bcVar), len(sys.bcVar))
	copy(gs.bcVar, sys.bcVar)

	// A pooled GameState may still contain stage data from an older save.
	gs.stageRef = sys.stage
	gs.stage = nil
	gs.stageList = nil
	gs.stageState = stageRollbackState{}

	// Character-controlled stage mutations need the full clone.
	if sys.rollback.session != nil || sys.cfg.Netplay.Rollback.DesyncTestFrames > 0 {
		if gs.stageCanMutate() || sys.cfg.Netplay.Rollback.SaveStageData {
			gs.stageList, gs.stage = cloneStageList(gsp, sys.stageList, sys.stage)
		} else {
			gs.stageState = sys.stage.CloneRollbackState()
		}
	} else {
		// Save anyway if using debug keys
		gs.stageList, gs.stage = cloneStageList(gsp, sys.stageList, sys.stage)
	}

	gs.workBe = make([]BytecodeExp, len(sys.workBe), len(sys.workBe))
	for i := 0; i < len(sys.workBe); i++ {
		gs.workBe[i] = make([]OpCode, len(sys.workBe[i]), len(sys.workBe[i]))
		copy(gs.workBe[i], sys.workBe[i])
	}

	//gs.accel = sys.accel
	//gs.clsnDisplay = sys.clsnDisplay
	//gs.debugDisplay = sys.debugDisplay

	// Bound to CGO; kept as an explicit allocation.
	gs.workpal = make([]uint32, len(sys.workpal))
	copy(gs.workpal, sys.workpal)

	gs.fightScreen = sys.fightScreen.Clone()
	gs.motif = sys.motif.Clone(sys.postMatchFlg)

	// Storyboard: only rollback-save while active.
	if sys.storyboard.active {
		gs.storyboard = sys.storyboard.Clone()
	} else {
		gs.storyboard = Storyboard{}
		gs.storyboard.active = false
	}
	gs.matchStoryboards = cloneStoryboards(sys.matchStoryboards.instances)

	gs.timerRounds = make([]int32, len(sys.timerRounds), len(sys.timerRounds))
	copy(gs.timerRounds, sys.timerRounds)
	gs.scoreRounds = make([][2]float32, len(sys.scoreRounds), len(sys.scoreRounds))
	copy(gs.scoreRounds, sys.scoreRounds)
	gs.statsState = sys.statsLog.saveRollbackState()

	//gs.sel = sys.sel.Clone()
	// for i := 0; i < len(sys.stringPool); i++ {
	//		gs.stringPool[i] = sys.stringPool[i].Clone(gsp)
	// }

	gs.dialogueFlg = sys.motif.di.active

	gs.timerCount = make([]int32, len(sys.timerCount), len(sys.timerCount))
	copy(gs.timerCount, sys.timerCount)

	gs.commandLists = make([]*CommandList, len(sys.commandLists), len(sys.commandLists))
	for i := 0; i < len(sys.commandLists); i++ {
		cl := sys.commandLists[i].Clone()
		gs.commandLists[i] = &cl
	}

	// Log save state
	if sys.rollback.session == nil {
		sys.appendToConsole(fmt.Sprintf("%v: Game state saved", sys.tickCount))
	}
}

func (src *CommandList) CopyTo(dst *CommandList) {
	clone := src.Clone()
	*dst = clone
}

func (gs *GameState) savePalFX() {
	gs.allPalFX = sys.allPalFX.Clone()
	gs.bgPalFX = sys.bgPalFX.Clone()
}

func (gs *GameState) saveCharData(gsp *GameStatePool) {
	for i := range sys.chars {
		gs.charData[i] = make([]Char, len(sys.chars[i]), len(sys.chars[i]))
		gs.chars[i] = make([]*Char, len(sys.chars[i]), len(sys.chars[i]))

		for j, c := range sys.chars[i] {
			gs.charData[i][j] = c.Clone(gsp)
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
	gs.charList = sys.charList.Clone(gsp)
}

func (gs *GameState) saveProjectileData(gsp *GameStatePool) {
	for i := range sys.projs {
		gs.projs[i] = make([]*Projectile, len(sys.projs[i]), len(sys.projs[i]))
		for j := 0; j < len(sys.projs[i]); j++ {
			gs.projs[i][j] = sys.projs[i][j].clone(gsp)
		}
	}
}

func (gs *GameState) saveExplodData(gsp *GameStatePool) {
	for i := range sys.explods {
		gs.explods[i] = make([]*Explod, len(sys.explods[i]), len(sys.explods[i]))
		for j := 0; j < len(sys.explods[i]); j++ {
			gs.explods[i][j] = sys.explods[i][j].Clone(gsp)
		}
	}
}

func (gs *GameState) saveCharTextData() {
	for i := range sys.chartexts {
		gs.chartexts[i] = make([]*TextSprite, len(sys.chartexts[i]), len(sys.chartexts[i]))
		for j := range sys.chartexts[i] {
			gs.chartexts[i][j] = cloneTextSprite(sys.chartexts[i][j])
		}
	}
}

func (gs *GameState) loadPalFX() {
	sys.allPalFX = gs.allPalFX.Clone()
	sys.bgPalFX = gs.bgPalFX.Clone()
}

func (gs *GameState) loadCharData(gsp *GameStatePool) {
	for i := 0; i < len(sys.chars); i++ {
		sys.chars[i] = make([]*Char, len(gs.chars[i]), len(gs.chars[i]))
		copy(sys.chars[i], gs.chars[i])
	}

	for i := 0; i < len(sys.chars); i++ {
		for j := 0; j < len(sys.chars[i]); j++ {
			*sys.chars[i][j] = gs.charData[i][j].Clone(gsp)
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
		sys.workingState = c.ss.sb
		sys.debugWC = c
	}

	sys.charList = gs.charList.Clone(gsp)
}

func (gs *GameState) loadProjectileData(gsp *GameStatePool) {
	for i := range gs.projs {
		sys.projs[i] = make([]*Projectile, len(gs.projs[i]), len(gs.projs[i]))
		for j := range gs.projs[i] {
			sys.projs[i][j] = gs.projs[i][j].clone(gsp)
		}
	}
}

func (gs *GameState) loadExplodData(gsp *GameStatePool) {
	for i := range gs.explods {
		sys.explods[i] = make([]*Explod, len(gs.explods[i]), len(gs.explods[i]))
		for j := 0; j < len(gs.explods[i]); j++ {
			sys.explods[i][j] = gs.explods[i][j].Clone(gsp)
		}
	}
}

func (gs *GameState) loadCharTextData() {
	for i := range gs.chartexts {
		sys.chartexts[i] = make([]*TextSprite, len(gs.chartexts[i]), len(gs.chartexts[i]))
		for j := range gs.chartexts[i] {
			sys.chartexts[i][j] = cloneTextSprite(gs.chartexts[i][j])
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
	str = fmt.Sprintf("MatchTime %d CurRoundTime: %d RandSeed: %d\n",
		gs.matchTime, gs.curRoundTime, gs.randseed)
	str = fmt.Sprintf("ScorePoints: %v ComboCount: %v\n",
		gs.scorePoints, gs.comboCount)
	str += fmt.Sprintf("RoundState RoundNo:%d Intro:%d WinSkipped:%t WinPoseTime:%d WinWaitTime:%d FinishType:%v SpecialFlag:0x%08x Wins:%v EffectiveLoss:%v SlowTime:%d WinTeam:%d\n",
		gs.roundNo, gs.intro, gs.winskipped, gs.winposetime, gs.winwaittime, gs.finishType,
		uint32(gs.specialFlag), gs.wins, gs.effectiveLoss, gs.slowtime, gs.winTeam)

	if round := gs.fightScreen.round; round != nil {
		fadeoutActive := false
		fadeoutRemain := int32(0)
		if round.fadeOut != nil {
			fadeoutActive = round.fadeOut.isActive()
			fadeoutRemain = round.fadeOut.timeRemaining
		}
		str += fmt.Sprintf("RoundUI Round:%d/%d Fight:%d/%d KO:%d/%d Win:%d/%d TimerActive:%t OutroStarted:%t OutroFrameAcc:%v Shutter:%d FadeOutActive:%t FadeOutRemaining:%d\n",
			round.roundDisplayPhase, round.roundDisplayTimer,
			round.fightDisplayPhase, round.fightDisplayTimer,
			round.koDisplayPhase, round.koDisplayTimer,
			round.winDisplayPhase, round.winDisplayTimer,
			round.timerActive, round.outroStarted, round.outroFrameAcc,
			round.shutterTimer, fadeoutActive, fadeoutRemain)
	}

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
	stringMapValueMapPool   sync.Pool
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
		stringMapValueMapPool: sync.Pool{
			New: func() interface{} {
				sm := make(map[string]MapValue)
				return &sm
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
	case (map[string]MapValue):
		objs = append(objs, gsp.stringMapValueMapPool.Get())
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
	case (*map[string]MapValue):
		gsp.stringMapValueMapPool.Put(item)
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

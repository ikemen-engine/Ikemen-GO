package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"log"
	"math"
	"os"
	"strings"
	"time"

	ggpo "github.com/ikemen-engine/ggpo"
)

type RollbackSystem struct {
	session          *RollbackSession
	netConnection    *NetConnection
	ggpoInputs       []InputBits
	ggpoAnalogInputs [][6]int8
}

type RollbackProperties struct {
	FrameDelay            int  `ini:"FrameDelay"`
	DisconnectNotifyStart int  `ini:"DisconnectNotifyStart"`
	DisconnectTimeout     int  `ini:"DisconnectTimeout"`
	LogsEnabled           bool `ini:"LogsEnabled"`
	SaveStageData         bool `ini:"SaveStageData"`
	DesyncTest            bool `ini:"DesyncTest"`
	DesyncTestFrames      int  `ini:"DesyncTestFrames"`
	DesyncTestAI          bool `ini:"DesyncTestAI"`
}

// TODO: Merge with system.go
func (rs *RollbackSystem) hijackRunMatch(s *System) bool {
	// Shared part up until this point already handled by sys.runMatch()

	// Reset variables
	rs.ggpoInputs = make([]InputBits, 2)
	rs.ggpoAnalogInputs = make([][6]int8, 2)

	// Initialize rollback network session and synchronize state
	rs.preMatchSetup()

	var running bool

	// Loop until end of match
	for !s.endMatch {
		rs.session.now = time.Now().UnixMilli()
		err := rs.session.backend.Idle(
			int(math.Max(0, float64(rs.session.next-rs.session.now-1))))
		if err != nil {
			panic(err)
		}

		// Desync/disconnect callbacks may request a session abort outside the normal input path.
		if s.esc {
			break
		}

		// Sync speculative inputs and run a speculative frame
		running = rs.runFrame(s)

		if s.fightLoopEnd && !s.postMatchFlg {
			break
		}

		rs.session.next = rs.session.now + 1000/60

		if s.esc || !running {
			break
		}

		s.renderFrame()

		//rs.session.loopTimer.usToWaitThisLoop()
		running = s.update()

		if s.esc || !running {
			break
		}
	}

	rs.postMatchSetup()

	return false
}

func (rs *RollbackSystem) preMatchSetup() {
	if rs.session != nil && sys.netConnection != nil {
		if rs.session.host != "" {
			// Initialize client as P2
			rs.session.InitP2(2, 7550, 7600, rs.session.host)
			rs.session.playerNo = 2
		} else {
			// Initialize host as P1
			rs.session.InitP1(2, 7600, 7550, rs.session.remoteIp)
			rs.session.playerNo = 1
		}

		// Synchronize matchTime at match start
		sys.matchTime = rs.session.netTime //s.time = rs.session.netTime // Old typo?
		sys.preMatchTime = sys.netConnection.preMatchTime

		// Wait until both peers have fully synchronized?
		//if !rs.session.IsConnected() {
		//	for !rs.session.synchronized {
		//		rs.session.backend.Idle(0)
		//	}
		//}
		//sys.netConnection.Close()

		// Borrow netConnection replay recording
		rs.session.recording = sys.netConnection.recording

		// Transfer the active netConnection to the rollback system
		rs.netConnection = sys.netConnection
		sys.netConnection = nil

	} else if sys.netConnection == nil && rs.session == nil {
		// If offline, initialize a local rollback sync test session
		session := NewRollbackSession(sys.cfg.Netplay.Rollback)
		rs.session = &session
		rs.session.InitSyncTest(2)
	}

	// Reset rollback session timer
	rs.session.netTime = 0
}

func (rs *RollbackSystem) postMatchSetup() {
	rs.session.SaveReplay()

	// sys.esc = true

	sys.netConnection = rs.netConnection
	rs.session.backend.Close()

	// Prep for the next match.
	if sys.netConnection != nil {
		newSession := NewRollbackSession(sys.cfg.Netplay.Rollback)
		host := rs.session.host
		remoteIp := rs.session.remoteIp

		rs.session = &newSession
		rs.session.host = host
		rs.session.remoteIp = remoteIp
	} else {
		rs.session = nil
	}
}

// Called once per frame by the main game loop
// Responsible for collecting local inputs and driving the GGPO backend forward
func (rs *RollbackSystem) runFrame(s *System) bool {
	var buffer []byte
	var ggpoerr error

	if s.esc {
		return false
	}

	if rs.session.syncTest && rs.session.netTime == 0 {
		if rs.session.config.DesyncTestAI {
			buffer = getAIInputs(0)
			ggpoerr = rs.session.backend.AddLocalInput(ggpo.PlayerHandle(0), buffer, len(buffer))
			buffer = getAIInputs(1)
			ggpoerr = rs.session.backend.AddLocalInput(ggpo.PlayerHandle(1), buffer, len(buffer))
		} else {
			buffer = rs.getInputs(0)
			ggpoerr = rs.session.backend.AddLocalInput(ggpo.PlayerHandle(0), buffer, len(buffer))
			buffer = rs.getInputs(1)
			ggpoerr = rs.session.backend.AddLocalInput(ggpo.PlayerHandle(1), buffer, len(buffer))
		}
	} else {
		buffer = rs.getInputs(0)
		ggpoerr = rs.session.backend.AddLocalInput(rs.session.currentPlayerHandle, buffer, len(buffer))
	}

	if ggpoerr == nil {
		// Get speculative inputs for the local player for this frame
		var values [][]byte
		disconnectFlags := 0
		values, ggpoerr = rs.session.backend.SyncInput(&disconnectFlags)

		if ggpoerr == nil {
			rs.ggpoInputs, rs.ggpoAnalogInputs = decodeInputs(values)

			// Record inputs against the rollback timeline. netTime is part of the saved state,
			// so any confirmed re-simulation overwrites earlier speculative data for the same frame.
			if rs.session.recording != nil {
				rs.session.RecordReplayFrame(rs.session.netTime, rs.ggpoInputs, rs.ggpoAnalogInputs)
				rs.session.netTime++
			}

			// Commit this frame to GGPO even if this frame exits gameplay.
			keepRunning := rs.simulateFrame(s)

			defer func() {
				if re := recover(); re != nil {
					if rs.session.config.DesyncTest {
						rs.session.log.updateLogs()
						rs.session.log.saveLogs()
						panic("RaiseDesyncError")
					}
				}
			}()

			err := rs.session.backend.AdvanceFrame(rs.session.LiveChecksum())
			if err != nil {
				panic(err)
			}
			if s.esc || !keepRunning {
				return false
			}
		}
	}

	return true
}

// Contains the logic for a single frame of the game
// Called by both runFrame for speculative execution, and AdvanceFrame for confirmed execution
func (rs *RollbackSystem) simulateFrame(s *System) bool {
	s.frameStepFlag = false

	// If next round
	if !sys.runNextRound() {
		return false
	}

	// Update game state
	s.action()

	// if rs.handleFlags(s) {
	//	return true
	// }

	if !rs.updateEvents(s) {
		return false
	}

	// Break if finished
	if s.fightLoopEnd && !s.postMatchFlg {
		return false
	}

	// Update system; break if update returns false (game ended)
	//if !s.update() {
	//	return false
	//}

	// If end match selected from menu/end of attract mode match/etc
	if s.endMatch {
		s.esc = true
		return false
	} else if s.esc {
		s.endMatch = s.netConnection != nil
		return false
	}
	return true
}

func (rs *RollbackSystem) updateEvents(s *System) bool {
	if !s.addFrameTime(s.turbo) {
		if !s.eventUpdate() {
			return false
		}
		return false
	}
	return true
}

func getAIInputs(player int) []byte {
	var ib InputBits
	ib.SetInputAI(player)
	return writeI32(int32(ib))
}

func (ib *InputBits) SetInputAI(in int) {
	*ib = InputBits(Btoi(sys.aiInput[in].U()) |
		Btoi(sys.aiInput[in].D())<<1 |
		Btoi(sys.aiInput[in].L())<<2 |
		Btoi(sys.aiInput[in].R())<<3 |
		Btoi(sys.aiInput[in].a())<<4 |
		Btoi(sys.aiInput[in].b())<<5 |
		Btoi(sys.aiInput[in].c())<<6 |
		Btoi(sys.aiInput[in].x())<<7 |
		Btoi(sys.aiInput[in].y())<<8 |
		Btoi(sys.aiInput[in].z())<<9 |
		Btoi(sys.aiInput[in].s())<<10 |
		Btoi(sys.aiInput[in].d())<<11 |
		Btoi(sys.aiInput[in].w())<<12 |
		Btoi(sys.aiInput[in].m())<<13)
}

func readI16(b []byte) int16 {
	if len(b) < 2 {
		return 0
	}
	return int16(b[0]) | int16(b[1])<<8
}

func readI32(b []byte) int32 {
	if len(b) < 4 {
		return 0
	}
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
}

func decodeInputs(buffer [][]byte) ([]InputBits, [][6]int8) {
	var inputs = make([]InputBits, len(buffer))
	var analogInputs = make([][6]int8, len(buffer))
	for i, b := range buffer {
		inputs[i] = InputBits(readI16(b))
		for j := 0; j < len(analogInputs[i]); j++ {
			if len(b) < 8 {
				analogInputs[i][j] = 0
			} else {
				analogInputs[i][j] = int8(b[2+j])
			}
		}
	}
	return inputs, analogInputs
}

func writeI16(i16 int16) []byte {
	b := []byte{byte(i16), byte((i16 >> 8) & 0xFF)}
	return b
}

func writeI32(i32 int32) []byte {
	b := []byte{byte(i32), byte(i32 >> 8), byte(i32 >> 16), byte(i32 >> 24)}
	return b
}

func (rs *RollbackSystem) getInputs(player int) []byte {
	// Digital inputs
	var ib InputBits
	ib.KeysToBits(rs.netConnection.buf[player].InputReader.LocalInput(0))
	bytes := writeI16(int16(ib))

	// Analog inputs
	sbyteAxes := rs.netConnection.buf[player].InputReader.LocalAnalogInput(0)
	for i := 0; i < len(sbyteAxes); i++ {
		bytes = append(bytes, byte(sbyteAxes[i]))
	}

	return bytes
}

func (rs *RollbackSystem) readRollbackInput(controller int) [14]bool {
	if controller < 0 || controller >= len(sys.inputRemap) {
		return [14]bool{}
	}

	remap := sys.inputRemap[controller]
	if remap < 0 || remap >= len(rs.ggpoInputs) {
		return [14]bool{}
	}

	return rs.ggpoInputs[remap].BitsToKeys()
}

func (rs *RollbackSystem) readRollbackInputAnalog(controller int) [6]int8 {
	if controller < 0 || controller >= len(sys.inputRemap) {
		return [6]int8{}
	}

	remap := sys.inputRemap[controller]
	if remap < 0 || remap >= len(rs.ggpoInputs) {
		return [6]int8{}
	}

	return rs.ggpoAnalogInputs[remap]
}

func (rs *RollbackSystem) anyButton() bool {
	for _, b := range rs.ggpoInputs {
		if b&IB_anybutton != 0 {
			return true
		}
	}
	return false
}

type RollbackLogger struct {
	filename   string
	currentLog strings.Builder
	logs       [(MaxSaveStates + 3) * 3]string
}

func NewRollbackLogger(timestamp string) RollbackLogger {
	gl := RollbackLogger{filename: "save/logs/Rollback-State-" + timestamp + ".log"}
	return gl
}

func (g *RollbackLogger) logState(action string, stateID int, state string, checksum int) {
	fmt.Fprintf(&g.currentLog, "State Logger: \n")
	fmt.Fprintf(&g.currentLog, "%s state for stateID: %d\n", action, stateID)
	fmt.Fprintf(&g.currentLog, state+"\n")
	fmt.Fprintf(&g.currentLog, "Checksum: %d\n", checksum)
}

func (g *RollbackLogger) Write(p []byte) (n int, err error) {
	g.currentLog.WriteString(string(p))
	return len(p), nil
}

func (g *RollbackLogger) updateLogs() {
	tmp := g.logs
	for i := 0; i < len(g.logs)-1; i++ {
		g.logs[i] = tmp[i+1]
	}
	g.logs[len(g.logs)-1] = g.currentLog.String()
	g.currentLog.Reset()
}

func (g *RollbackLogger) saveLogs() {
	var fullLog string
	for i := 0; i < len(g.logs); i++ {
		fullLog += g.logs[i]
	}
	err := os.WriteFile(g.filename, []byte(fullLog), 0666)
	if err != nil {
		fmt.Println(err)
	}
}

type RollbackSession struct {
	backend             ggpo.Backend
	saveStates          map[int]*GameState
	now                 int64
	next                int64
	players             []ggpo.Player
	handles             []ggpo.PlayerHandle
	recording           *os.File
	connected           bool
	host                string
	playerNo            int
	syncProgress        int
	synchronized        bool
	syncTest            bool
	run                 int
	remoteIp            string
	currentPlayer       int
	currentPlayerHandle ggpo.PlayerHandle
	remotePlayerHandle  ggpo.PlayerHandle
	loopTimer           LoopTimer
	inputs              map[int][MaxPlayerNo]InputBits
	analogInputs        map[int][MaxPlayerNo][6]int8
	config              RollbackProperties
	log                 RollbackLogger
	timestamp           string
	netTime             int32
	replayInputs        [][REPLAY_NUM_INPUTS]InputBits
	replayAnalogInputs  [][REPLAY_NUM_INPUTS][6]int8
	lastConfirmedInput  [MaxPlayerNo]InputBits
	inputBits           []InputBits
	inRollback          bool
	replaySaved         bool
}

func (rs *RollbackSession) SetInput(time int32, player int, input InputBits, axes [6]int8) {
	if _, ok := rs.inputs[int(time)]; !ok {
		rs.inputs[int(time)] = [MaxPlayerNo]InputBits{}
		rs.analogInputs[int(time)] = [MaxPlayerNo][6]int8{}
	}
	inputArr := rs.inputs[int(time)]
	analogInputs := rs.analogInputs[int(time)]
	inputArr[player] = input
	analogInputs[player] = axes
	rs.inputs[int(time)] = inputArr
	rs.analogInputs[int(time)] = analogInputs
}

func (rs *RollbackSession) ensureReplayFrame(frame int) {
	if frame < len(rs.replayInputs) {
		return
	}
	if frame > len(rs.replayInputs) {
		log.Printf("Rollback replay frame gap detected: expected frame %d, got %d; padding missing frames with zeros", len(rs.replayInputs), frame)
	}
	missing := frame - len(rs.replayInputs) + 1
	rs.replayInputs = append(rs.replayInputs, make([][REPLAY_NUM_INPUTS]InputBits, missing)...)
	rs.replayAnalogInputs = append(rs.replayAnalogInputs, make([][REPLAY_NUM_INPUTS][6]int8, missing)...)
}

func (rs *RollbackSession) RecordReplayFrame(time int32, inputs []InputBits, axes [][6]int8) {
	if rs == nil || time < 0 {
		return
	}
	frame := int(time)
	rs.ensureReplayFrame(frame)

	// Always clear the frame first so any overwritten rollback frame keeps zeroes in slots not present in the current input payload.
	rs.replayInputs[frame] = [REPLAY_NUM_INPUTS]InputBits{}
	rs.replayAnalogInputs[frame] = [REPLAY_NUM_INPUTS][6]int8{}

	for i := 0; i < REPLAY_NUM_INPUTS; i++ {
		if i < len(inputs) {
			rs.replayInputs[frame][i] = inputs[i]
		}

		if i < len(axes) {
			rs.replayAnalogInputs[frame][i] = axes[i]
		}
	}
}

func (rs *RollbackSession) TruncateReplayFrom(time int32) {
	if rs == nil {
		return
	}
	frame := int(time)
	if frame < 0 {
		frame = 0
	}
	if frame < len(rs.replayInputs) {
		rs.replayInputs = rs.replayInputs[:frame]
	}
	if frame < len(rs.replayAnalogInputs) {
		rs.replayAnalogInputs = rs.replayAnalogInputs[:frame]
	}
}

func (rs *RollbackSession) SaveReplay() {
	if rs == nil || rs.recording == nil || rs.replaySaved {
		return
	}

	// Never append the same rollback replay twice.
	rs.replaySaved = true

	frameCount := int(rs.netTime)
	if frameCount <= 0 {
		return
	}

	if len(rs.replayInputs) < frameCount || len(rs.replayAnalogInputs) < frameCount {
		frameCount = len(rs.replayInputs)
		if len(rs.replayAnalogInputs) < frameCount {
			frameCount = len(rs.replayAnalogInputs)
		}
		log.Printf("Missing rollback replay frames after frame %d; truncating replay at this point", frameCount)
	}

	for frame := 0; frame < frameCount; frame++ {
		inputFrame := rs.replayInputs[frame]
		axisFrame := rs.replayAnalogInputs[frame]
		for i := 0; i < REPLAY_NUM_INPUTS; i++ {
			if err := writeReplayInput(rs.recording, inputFrame[i], axisFrame[i]); err != nil {
				log.Printf("Error while writing rollback replay frame %d controller %d: %v", frame, i, err)
				return
			}
		}
	}
}

type LoopTimer struct {
	lastAdvantage      float32
	usPergameLoop      time.Duration
	usExtraToWait      int
	framesToSpreadWait int
	waitCount          time.Duration
	timeWait           time.Duration
	waitTotal          time.Duration
}

func NewLoopTimer(fps uint32, framesToSpread uint32) LoopTimer {
	return LoopTimer{
		framesToSpreadWait: int(framesToSpread),
		usPergameLoop:      time.Second / time.Duration(fps),
	}
}

func (lt *LoopTimer) OnGGPOTimeSyncEvent(framesAhead float32) {
	lt.waitTotal = time.Duration(float32(time.Second/60) * framesAhead)
	lt.lastAdvantage = float32(time.Second/60) * framesAhead

	if sys.intro > 0 && sys.tickCount == 0 {
		// Wait longer during the start of each round, allowing both players to load assets
		if lt.lastAdvantage < 0 {
			lt.timeWait = time.Duration(lt.lastAdvantage) / time.Duration(lt.framesToSpreadWait)
			lt.waitCount = time.Duration(lt.framesToSpreadWait)
		}
	} else {
		// Normal waiting time
		lt.lastAdvantage /= 4
		lt.timeWait = time.Duration(lt.lastAdvantage) / time.Duration(lt.framesToSpreadWait)
		lt.waitCount = time.Duration(lt.framesToSpreadWait)
	}
}

func (lt *LoopTimer) usToWaitThisLoop() time.Duration {
	if lt.waitCount > 0 {
		lt.waitCount--
		return lt.usPergameLoop + lt.timeWait
	}
	return lt.usPergameLoop
}

func (r *RollbackSession) Close() {
	if r.backend != nil {
		r.backend.Close()
	}
}

func (r *RollbackSession) IsConnected() bool {
	return r.connected
}

func (r *RollbackSession) SaveGameState(stateID int) int {
	sys.savePool.curStateID = stateID
	sys.rollbackStateID = stateID
	oldest := stateID + 1%(MaxSaveStates+2)
	if _, ok := sys.arenaSaveMap[oldest]; ok {
		sys.arenaSaveMap[oldest].Free()
		sys.arenaSaveMap[oldest] = nil
		delete(sys.arenaSaveMap, oldest)
	}
	sys.savePool.Free(oldest)
	if _, ok := sys.arenaSaveMap[stateID]; ok {
		sys.arenaSaveMap[stateID].Free()
		sys.arenaSaveMap[stateID] = nil
		delete(sys.arenaSaveMap, stateID)
	}
	sys.savePool.Free(stateID)

	r.saveStates[stateID] = sys.statePool.gameStatePool.Get().(*GameState)
	r.saveStates[stateID].SaveState(stateID)

	if r.config.DesyncTest {
		checksum := r.saveStates[stateID].Checksum()
		if r.config.LogsEnabled {
			r.log.logState("Saving", stateID, r.saveStates[stateID].String(), checksum)
			r.log.updateLogs()
		}
		return checksum
	} else {
		if r.config.LogsEnabled {
			checksum := r.saveStates[stateID].Checksum()
			r.log.logState("Saving", stateID, r.saveStates[stateID].String(), checksum)
			r.log.updateLogs()
		}
		return ggpo.DefaultChecksum
	}
}

var lastLoadedFrame int = -1

func (r *RollbackSession) LoadGameState(stateID int) {
	sys.loadPool.curStateID = stateID
	if _, ok := sys.arenaLoadMap[stateID]; ok {
		sys.arenaLoadMap[stateID].Free()
		sys.arenaLoadMap[stateID] = nil
		delete(sys.arenaLoadMap, stateID)
	}
	for sid := range sys.arenaLoadMap {
		if sid != lastLoadedFrame {
			sys.arenaLoadMap[sid].Free()
			sys.arenaLoadMap[sid] = nil
			delete(sys.arenaLoadMap, sid)
		}
	}
	sys.loadPool.Free(stateID)

	r.saveStates[stateID].LoadState(stateID)

	if r.config.DesyncTest && r.config.LogsEnabled {
		checksum := r.saveStates[stateID].Checksum()
		r.log.logState("Loaded", stateID, r.saveStates[stateID].String(), checksum)
	}

	sys.statePool.gameStatePool.Put(r.saveStates[stateID])

	lastLoadedFrame = stateID
}

// Called when the GGPO backend needs the game to simulate a single frame
// This can happen multiple times per displayed frame during a rollback
func (r *RollbackSession) AdvanceFrame(flags int) {
	// This flag allows the game logic to re-run while knowing it's in a rollback
	// Will be useful later
	r.inRollback = true
	defer func() {
		r.inRollback = false
	}()

	// Make sure we fetch the inputs from GGPO and use these to update
	// the game state instead of reading from the keyboard.
	// Get the confirmed inputs from the GGPO backend for the frame being simulated
	var disconnectFlags int
	inputs, ggpoerr := r.backend.SyncInput(&disconnectFlags)

	// Run frame again using confirmed inputs
	if ggpoerr == nil {
		sys.rollback.ggpoInputs, sys.rollback.ggpoAnalogInputs = decodeInputs(inputs)
		if r.recording != nil {
			r.RecordReplayFrame(r.netTime, sys.rollback.ggpoInputs, sys.rollback.ggpoAnalogInputs)
			r.netTime++
		}
		// As in runFrame(): commit rollback re-simulated frames to GGPO even if this frame exits gameplay.
		_ = sys.rollback.simulateFrame(&sys)
		defer func() {
			if re := recover(); re != nil {
				if r.config.DesyncTest {
					r.log.updateLogs()
					r.log.saveLogs()
					panic("RaiseDesyncError")
				}
			}
		}()

		// Notify GGPO that frame has advanced
		err := r.backend.AdvanceFrame(r.LiveChecksum())
		if err != nil {
			panic(err)
		}
	}
}

func (r *RollbackSession) OnEvent(info *ggpo.Event) {
	switch info.Code {
	case ggpo.EventCodeConnectedToPeer:
		r.connected = true
	case ggpo.EventCodeSynchronizingWithPeer:
		r.syncProgress = 100 * (info.Count / info.Total)
	case ggpo.EventCodeSynchronizedWithPeer:
		r.syncProgress = 100
		r.synchronized = true
	case ggpo.EventCodeRunning:
		fmt.Println("EventCodeRunning")
	case ggpo.EventCodeDisconnectedFromPeer:
		if sys.postMatchFlg {
			return
		}
		if r.config.LogsEnabled {
			r.log.saveLogs()
		}
		fmt.Println("EventCodeDisconnectedFromPeer")
		sys.endMatch = true
		r.SaveReplay()
		if sys.sessionWarning == "" {
			sys.sessionWarning = fmt.Sprintf(sys.motif.WarningInfo.Text.Text["disconnect"], int(info.Player))
		}
	case ggpo.EventCodeTimeSync:
		fmt.Printf("EventCodeTimeSync: FramesAhead %f TimeSyncPeriodInFrames: %d\n", info.FramesAhead, info.TimeSyncPeriodInFrames)
		r.loopTimer.OnGGPOTimeSyncEvent(info.FramesAhead)
	case ggpo.EventCodeDesync:
		if r.config.LogsEnabled {
			r.log.saveLogs()
		}
		fmt.Println("EventCodeDesync")
		log.Printf("Rollback desync detected")
		sys.esc = true
		r.SaveReplay()
		sys.sessionWarning = sys.motif.WarningInfo.Text.Text["desync"]
	case ggpo.EventCodeConnectionInterrupted:
		fmt.Println("EventCodeconnectionInterrupted")
	case ggpo.EventCodeConnectionResumed:
		fmt.Println("EventCodeconnectionInterrupted")
	}
}

func NewRollbackSession(config RollbackProperties) RollbackSession {
	r := RollbackSession{}
	r.saveStates = make(map[int]*GameState)
	r.players = make([]ggpo.Player, 2)       // MaxPlayerNo
	r.handles = make([]ggpo.PlayerHandle, 2) // MaxPlayerNo
	r.config = config
	r.loopTimer = NewLoopTimer(60, 100)
	r.timestamp = time.Now().Format("2006-01-02_03-04PM-05s")
	r.log = NewRollbackLogger(r.timestamp)
	r.replayInputs = make([][REPLAY_NUM_INPUTS]InputBits, 0)
	r.replayAnalogInputs = make([][REPLAY_NUM_INPUTS][6]int8, 0)
	r.inputs = make(map[int][MaxPlayerNo]InputBits)
	r.analogInputs = make(map[int][MaxPlayerNo][6]int8)
	return r

}

func encodeInputs(inputs InputBits) []byte {
	return writeI16(int16(inputs))
}

func (rs *RollbackSession) LiveChecksum() uint32 {
	// System variables. Check always
	buf := writeI32(sys.randseed)
	buf = append(buf, writeI32(sys.matchTime)...)
	buf = append(buf, writeI32(sys.curRoundTime)...)

	// Round start checks. Ensure both players have the same selection
	if sys.roundState() == 1 {
		// Stage
		stageHash := crc32.ChecksumIEEE([]byte(sys.stage.name))
		buf = binary.BigEndian.AppendUint32(buf, stageHash)

		// Characters
		for i := range sys.chars {
			if len(sys.chars[i]) == 0 {
				continue
			}
			c := sys.chars[i][0]
			nameHash := crc32.ChecksumIEEE([]byte(c.name))
			buf = binary.BigEndian.AppendUint32(buf, nameHash)
		}

		// CharGlobalInfo
		// Checking selected palette is a little overzealous and makes palette modules desync
		//for i := range sys.cgi {
		//	buf = binary.BigEndian.AppendUint32(buf, uint32(sys.cgi[i].palno))
		//}
	}

	// During fight checks
	// Checking life during intros may cause trouble with Turns mode life refill
	if sys.roundState() == 2 || sys.roundState() == 3 {
		// Character data
		for i := range sys.chars {
			if len(sys.chars[i]) == 0 {
				continue
			}
			c := sys.chars[i][0]
			buf = binary.BigEndian.AppendUint32(buf, uint32(c.life))
			buf = binary.BigEndian.AppendUint32(buf, uint32(c.redLife))
			buf = binary.BigEndian.AppendUint32(buf, uint32(c.dizzyPoints))
			buf = binary.BigEndian.AppendUint32(buf, uint32(c.guardPoints))
			buf = binary.BigEndian.AppendUint32(buf, uint32(c.power))
			buf = binary.BigEndian.AppendUint32(buf, uint32(c.animNo))
			//buf = binary.BigEndian.AppendUint32(buf, math.Float32bits(cc.pos[0])) // These might add float operation errors
			//buf = binary.BigEndian.AppendUint32(buf, math.Float32bits(cc.pos[1]))
			//buf = binary.BigEndian.AppendUint32(buf, math.Float32bits(cc.pos[2]))
		}
	}

	return crc32.ChecksumIEEE(buf)
}

func (rs *RollbackSession) Input(time int32, player int) (input InputBits) {
	inputs := rs.inputs[int(time)]
	input = inputs[player]
	return
}

func (rs *RollbackSession) AnyButton() bool {
	for i := 0; i < len(rs.inputs[len(rs.inputs)-1]); i++ {
		if rs.Input(sys.matchTime, i)&IB_anybutton != 0 {
			return true
		}
	}
	return false
}

func (rs *RollbackSession) InitP1(numPlayers int, localPort int, remotePort int, remoteIp string) {
	if rs.config.LogsEnabled {
		logFileName := fmt.Sprintf("save/logs/Rollback-%s.log", rs.timestamp)
		f, err := os.OpenFile(logFileName, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
		logger := log.New(f, "Logger:", log.Ldate|log.Ltime|log.Lshortfile)
		ggpo.EnableLogs()
		ggpo.SetLogger(logger)
	}

	var inputBits InputBits = 0
	var inputAxes [6]int8 = [6]int8{}
	var inputSize int = len(encodeInputs(inputBits)) + len(inputAxes)

	player := ggpo.NewLocalPlayer(20, 1)
	player2 := ggpo.NewRemotePlayer(20, 2, remoteIp, remotePort)
	rs.players = append(rs.players, player)
	rs.players = append(rs.players, player2)

	peer := ggpo.NewPeer(rs, localPort, numPlayers, inputSize)
	rs.backend = &peer

	peer.InitializeConnection()

	var handle ggpo.PlayerHandle
	result := peer.AddPlayer(&player, &handle)
	if result != nil {
		panic("panic")
	}
	rs.handles = append(rs.handles, handle)
	var handle2 ggpo.PlayerHandle
	result = peer.AddPlayer(&player2, &handle2)
	if result != nil {
		panic("panic")
	}
	rs.handles = append(rs.handles, handle2)
	rs.currentPlayer = int(handle)
	rs.currentPlayerHandle = handle
	rs.remotePlayerHandle = handle2

	peer.SetDisconnectTimeout(rs.config.DisconnectTimeout)
	peer.SetDisconnectNotifyStart(rs.config.DisconnectNotifyStart)
	peer.SetFrameDelay(handle, rs.config.FrameDelay)

	peer.Start()
}

func (rs *RollbackSession) InitP2(numPlayers int, localPort int, remotePort int, remoteIp string) {
	if rs.config.LogsEnabled {
		logFileName := fmt.Sprintf("save/logs/Rollback-%s.log", rs.timestamp)
		f, err := os.OpenFile(logFileName, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
		logger := log.New(f, "Logger:", log.Ldate|log.Ltime|log.Lshortfile)
		ggpo.EnableLogs()
		ggpo.SetLogger(logger)
	}

	var inputBits InputBits = 0
	var inputAxes [6]int8 = [6]int8{}
	var inputSize int = len(encodeInputs(inputBits)) + len(inputAxes)

	player := ggpo.NewRemotePlayer(20, 1, remoteIp, remotePort)
	player2 := ggpo.NewLocalPlayer(20, 2)
	rs.players = append(rs.players, player)
	rs.players = append(rs.players, player2)

	peer := ggpo.NewPeer(rs, localPort, numPlayers, inputSize)
	rs.backend = &peer

	peer.InitializeConnection()

	var handle ggpo.PlayerHandle
	result := peer.AddPlayer(&player, &handle)
	if result != nil {
		panic("panic")
	}
	rs.handles = append(rs.handles, handle)
	var handle2 ggpo.PlayerHandle
	result = peer.AddPlayer(&player2, &handle2)
	if result != nil {
		panic("panic")
	}
	rs.handles = append(rs.handles, handle2)
	rs.currentPlayer = int(handle2)
	rs.currentPlayerHandle = handle2
	rs.remotePlayerHandle = handle

	peer.SetDisconnectTimeout(rs.config.DisconnectTimeout)
	peer.SetDisconnectNotifyStart(rs.config.DisconnectNotifyStart)
	peer.SetFrameDelay(handle2, rs.config.FrameDelay)

	peer.Start()
}

func (rs *RollbackSession) InitSyncTest(numPlayers int) {
	rs.syncTest = true
	if rs.config.LogsEnabled {
		logFileName := fmt.Sprintf("save/logs/Rollback-Desync-Test-%s.log", rs.timestamp)
		f, err := os.OpenFile(logFileName, os.O_CREATE|os.O_RDWR, 0666)
		if err != nil {
			panic(err)
		}
		logger := log.New(f, "Logger:", log.Ldate|log.Ltime|log.Lshortfile)
		ggpo.EnableLogs()
		ggpo.SetLogger(logger)
	}

	var inputBits InputBits = 0
	var inputAxes [6]int8 = [6]int8{}
	var inputSize int = len(encodeInputs(inputBits)) + len(inputAxes)

	player := ggpo.NewLocalPlayer(20, 1)
	player2 := ggpo.NewLocalPlayer(20, 2)
	rs.players = append(rs.players, player)
	rs.players = append(rs.players, player2)

	//peer := ggpo.NewPeer(sys.rollbackNetwork, localPort, numPlayers, inputSize)
	peer := ggpo.NewSyncTest(rs, numPlayers, rs.config.DesyncTestFrames, inputSize, true)
	rs.backend = &peer

	//
	peer.InitializeConnection()
	peer.Start()

	var handle ggpo.PlayerHandle
	result := peer.AddPlayer(&player, &handle)
	if result != nil {
		panic("panic")
	}
	rs.handles = append(rs.handles, handle)
	var handle2 ggpo.PlayerHandle
	result = peer.AddPlayer(&player2, &handle2)
	if result != nil {
		panic("panic")
	}
	rs.handles = append(rs.handles, handle2)

	peer.SetDisconnectTimeout(3000)
	peer.SetDisconnectNotifyStart(1000)
}

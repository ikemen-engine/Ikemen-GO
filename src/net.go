package main

import (
	"encoding/binary"
	"fmt"
	"hash/crc32"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	ggpo "github.com/assemblaj/ggpo"
	"golang.org/x/exp/maps"
)

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
	config              RollbackProperties
	log                 RollbackLogger
	timestamp           string
	netTime             int32
	replayBuffer        [][MaxPlayerNo]InputBits
	lastConfirmedInput  [MaxPlayerNo]InputBits
	inputBits           []InputBits
	inRollback          bool
}

func (rs *RollbackSession) SetInput(time int32, player int, input InputBits) {
	if _, ok := rs.inputs[int(time)]; !ok {
		rs.inputs[int(time)] = [MaxPlayerNo]InputBits{}
	}
	inputArr := rs.inputs[int(time)]
	inputArr[player] = input
	rs.inputs[int(time)] = inputArr
}

func (rs *RollbackSession) SaveReplay() {
	if rs.recording != nil && len(rs.inputs) > 0 {
		frames := maps.Keys(rs.inputs)
		sort.Ints(frames)
		lastFrame := frames[len(frames)-1]

		size := lastFrame * (MaxPlayerNo) * 4
		buf := make([]byte, size)
		offset := 0

		for i := 0; i < lastFrame; i++ {
			if _, ok := rs.inputs[i]; ok {
				inputBuf := rs.inputToBytes(i)
				copy(buf[offset:offset+len(inputBuf)], inputBuf)
				offset += len(inputBuf)
			} else {
				inputBuf := []byte{}
				for i := 0; i < MaxSimul*2+MaxAttachedChar; i++ {
					inputBuf = append(inputBuf, writeI32(int32(0))...)
				}
				copy(buf[offset:offset+len(inputBuf)], inputBuf)
				offset += len(inputBuf)
			}
		}
		rs.recording.Write(buf)
	}
}

func (rs *RollbackSession) inputToBytes(time int) []byte {
	buf := []byte{}
	for i := 0; i < MaxSimul*2+MaxAttachedChar; i++ {
		buf = append(buf, writeI32(int32(rs.inputs[time][i]))...)
	}
	return buf
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
	if sys.intro > 0 && sys.tickCount == 0 {
		lt.waitTotal = time.Duration(float32(time.Second/60) * framesAhead)
		lt.lastAdvantage = float32(time.Second/60) * framesAhead
		if lt.lastAdvantage < float32(0) {
			lt.timeWait = time.Duration(lt.lastAdvantage) / time.Duration(lt.framesToSpreadWait)
			lt.waitCount = time.Duration(lt.framesToSpreadWait)
		}
	} else {
		lt.waitTotal = time.Duration(float32(time.Second/60) * framesAhead)
		lt.lastAdvantage = float32(time.Second/60) * framesAhead
		lt.lastAdvantage = lt.lastAdvantage / 4
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
	sys.rollback.ggpoInputs = decodeInputs(inputs)

	if r.recording != nil {
		r.SetInput(r.netTime, 0, sys.rollback.ggpoInputs[0])
		r.SetInput(r.netTime, 1, sys.rollback.ggpoInputs[1])
		r.netTime++
	}

	// Run frame again using confirmed inputs
	if ggpoerr == nil {
		if !sys.rollback.simulateFrame(&sys) {
			return
		}
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
		sys.rollback.currentFight.fin = true
		sys.endMatch = true
		disconnectMessage := fmt.Sprintf("Player %d disconnected.", info.Player)
		r.SaveReplay()
		ShowInfoDialog(disconnectMessage, "Disconnection")
	case ggpo.EventCodeTimeSync:
		fmt.Printf("EventCodeTimeSync: FramesAhead %f TimeSyncPeriodInFrames: %d\n", info.FramesAhead, info.TimeSyncPeriodInFrames)
		r.loopTimer.OnGGPOTimeSyncEvent(info.FramesAhead)
	case ggpo.EventCodeDesync:
		if r.config.LogsEnabled {
			r.log.saveLogs()
		}
		fmt.Println("EventCodeDesync")
		sys.rollback.currentFight.fin = true
		sys.endMatch = true
		r.SaveReplay()
		ShowInfoDialog("Desync error.\nIf the problem persists, please report it at:\nhttps://github.com/ikemen-engine/Ikemen-GO/issues.\nThank you for your patience", "Desync Error")
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
	r.replayBuffer = make([][MaxPlayerNo]InputBits, 0)
	r.inputs = make(map[int][MaxPlayerNo]InputBits)
	return r

}

func encodeInputs(inputs InputBits) []byte {
	return writeI32(int32(inputs))
}

func (rs *RollbackSession) LiveChecksum() uint32 {
	// System variables. Check always
	buf := writeI32(sys.randseed)
	buf = append(buf, writeI32(sys.gameTime)...)
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
		for i := range sys.cgi {
			buf = binary.BigEndian.AppendUint32(buf, uint32(sys.cgi[i].palno))
		}
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
		if rs.Input(sys.gameTime, i)&IB_anybutton != 0 {
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
	var inputSize int = len(encodeInputs(inputBits))

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
	var inputSize int = len(encodeInputs(inputBits))

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
	var inputSize int = len(encodeInputs(inputBits))

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

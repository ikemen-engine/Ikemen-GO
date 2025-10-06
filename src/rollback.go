package main

import (
	"math"
	"time"

	ggpo "github.com/assemblaj/ggpo"
)

type RollbackSystem struct {
	session       *RollbackSession
	currentFight  Fight
	netConnection *NetConnection
	ggpoInputs    []InputBits
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

	// Initialize rollback network session and synchronize state
	rs.preMatchSetup()

	var didTryLoadBGM bool
	var running bool

	// Loop until end of match
	for !s.endMatch {
		rs.session.now = time.Now().UnixMilli()
		err := rs.session.backend.Idle(
			int(math.Max(0, float64(rs.session.next-rs.session.now-1))))
		if err != nil {
			panic(err)
		}

		// Sync speculative inputs and run a speculative frame
		running = rs.runFrame(s)

		// Default BGM playback. Used only in Quick VS or if externalized Lua implementaion is disabled
		if !didTryLoadBGM && s.round == 1 && (s.gameMode == "" || len(sys.cfg.Common.Lua) == 0) && sys.stage.stageTime > 0 {
			didTryLoadBGM = true
			// Need to search first
			LoadFile(&s.stage.bgmusic, []string{s.stage.def, "", "sound/"}, func(path string) error {
				s.bgm.Open(path, 1, int(s.stage.bgmvolume), int(s.stage.bgmloopstart), int(s.stage.bgmloopend), int(s.stage.bgmstartposition), s.stage.bgmfreqmul, -1)
				return nil
			})
		}

		if s.fightLoopEnd && (!s.postMatchFlg || len(s.cfg.Common.Lua) == 0) {
			break
		}

		rs.session.next = rs.session.now + 1000/60

		if !running {
			break
		}

		s.renderFrame()

		//rs.session.loopTimer.usToWaitThisLoop()
		running = s.update()

		if !running {
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

		// Synchronize gameTime at match start
		sys.gameTime = rs.session.netTime //s.time = rs.session.netTime // Old typo?
		sys.preFightTime = sys.netConnection.preFightTime

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
	//sys.rollback.currentFight.fin = true

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
		rs.ggpoInputs = decodeInputs(values)

		// TODO: Why does this depend on the replay?
		if rs.session.recording != nil {
			rs.session.SetInput(rs.session.netTime, 0, rs.ggpoInputs[0])
			rs.session.SetInput(rs.session.netTime, 1, rs.ggpoInputs[1])
			rs.session.netTime++
		}

		if ggpoerr == nil {
			if !rs.simulateFrame(s) {
				return false
			}

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

	s.bgPalFX.step()
	s.stage.action()

	// If frame is ready to tick and not paused
	//sys.stage.action()

	// Update game state
	s.action()

	// if rs.handleFlags(s) {
	//	return true
	// }

	if !rs.updateEvents(s) {
		return false
	}

	// Break if finished
	if s.fightLoopEnd && (!s.postMatchFlg || len(s.cfg.Common.Lua) == 0) {
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
		s.endMatch = s.netConnection != nil || len(s.cfg.Common.Lua) == 0
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

func (rs *RollbackSystem) updateCamera(s *System) {
	if !s.frameSkip {
		scl := s.cam.Scale / s.cam.BaseScale()
		if s.enableZoomtime > 0 {
			if !s.debugPaused() {
				s.zoomPosXLag += ((s.zoomPos[0] - s.zoomPosXLag) * (1 - s.zoomlag))
				s.zoomPosYLag += ((s.zoomPos[1] - s.zoomPosYLag) * (1 - s.zoomlag))
				s.drawScale = s.drawScale / (s.drawScale + (s.zoomScale*scl-s.drawScale)*s.zoomlag) * s.zoomScale * scl
			}
		} else {
			s.zoomlag = 0
			s.zoomPosXLag = 0
			s.zoomPosYLag = 0
			s.zoomScale = 1
			s.zoomPos = [2]float32{0, 0}
			s.drawScale = s.cam.Scale
		}
	}
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

// system.go already refactored to use a local var to save all this data
type Fight struct {
	fin                                                               bool
	oldTeamLeader                                                     [2]int
	oldWins                                                           [2]int32
	oldDraws                                                          int32
	oldStageVars                                                      Stage
	level                                                             []int32
	lvmul                                                             float64
	life, pow, gpow, spow, rlife                                      []int32
	cnsvar                                                            []map[int32]int32
	cnsfvar                                                           []map[int32]float32
	dialogue                                                          [][]string
	mapArray                                                          []map[string]float32
	remapSpr                                                          []RemapPreset
	lifeMax, power, powerMax                                          []int32
	guardPoints, guardPointsMax, dizzyPoints, dizzyPointsMax, redLife []int32
	teamside                                                          []int
}

func NewFight() Fight {
	f := Fight{}
	return f
}

func readI32(b []byte) int32 {
	if len(b) < 4 {
		return 0
	}
	return int32(b[0]) | int32(b[1])<<8 | int32(b[2])<<16 | int32(b[3])<<24
}

func decodeInputs(buffer [][]byte) []InputBits {
	var inputs = make([]InputBits, len(buffer))
	for i, b := range buffer {
		inputs[i] = InputBits(readI32(b))
	}
	return inputs
}

func writeI32(i32 int32) []byte {
	b := []byte{byte(i32), byte(i32 >> 8), byte(i32 >> 16), byte(i32 >> 24)}
	return b
}

func (rs *RollbackSystem) getInputs(player int) []byte {
	var ib InputBits
	ib.KeysToBits(rs.netConnection.buf[player].InputReader.LocalInput(0, false))
	return writeI32(int32(ib))
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

func (rs *RollbackSystem) anyButton() bool {
	for _, b := range rs.ggpoInputs {
		if b&IB_anybutton != 0 {
			return true
		}
	}
	return false
}

package main

import (
	"fmt"
	"math"
	"time"

	ggpo "github.com/assemblaj/ggpo"
	lua "github.com/yuin/gopher-lua"
)

type RollbackSystem struct {
	session       *RollbackSession
	currentFight  Fight
	active        bool
	netConnection *NetConnection
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

func (rs *RollbackSystem) fight(s *System) bool {

	// Reset variables
	s.gameTime, s.paused, s.accel = 0, false, 1
	s.aiInput = [len(s.aiInput)]AiInput{}
	rs.currentFight = NewFight()
	// Defer resetting variables on return
	defer rs.currentFight.endFight()

	s.debugWC = sys.chars[0][0]
	// debugInput := func() {
	// 	select {
	// 	case cl := <-s.commandLine:
	// 		if err := s.luaLState.DoString(cl); err != nil {
	// 			s.errLog.Println(err.Error())
	// 		}
	// 	default:
	// 	}
	// }

	// Synchronize with external inputs (netplay, replays, etc)
	if err := s.synchronize(); err != nil {
		s.errLog.Println(err.Error())
		s.esc = true
	}
	if s.netConnection != nil {
		defer s.netConnection.Stop()
	}
	s.wincnt.init()

	rs.currentFight.initSuperMeter()
	rs.currentFight.initTeamsLevels()

	rs.currentFight.initChars()

	var didTryLoadBGM bool

	// default bgm playback, used only in Quick VS or if externalized Lua implementaion is disabled
	if s.round == 1 && (s.gameMode == "" || len(sys.cfg.Common.Lua) == 0) && sys.stage.stageTime > 0 && !didTryLoadBGM {
		// Need to search first
		LoadFile(&s.stage.bgmusic, []string{s.stage.def, "", "sound/"}, func(path string) error {
			s.bgm.Open(path, 1, int(s.stage.bgmvolume), int(s.stage.bgmloopstart), int(s.stage.bgmloopend), int(s.stage.bgmstartposition), s.stage.bgmfreqmul, -1)
			didTryLoadBGM = true
			return nil
		})
	}

	rs.currentFight.oldWins, rs.currentFight.oldDraws = s.wins, s.draws
	rs.currentFight.oldTeamLeader = s.teamLeader

	var running bool
	if rs.session != nil && s.netConnection != nil {
		if rs.session.host != "" {
			rs.session.InitP2(2, 7550, 7600, rs.session.host)
			rs.session.playerNo = 2
		} else {
			rs.session.InitP1(2, 7600, 7550, rs.session.remoteIp)
			rs.session.playerNo = 1
		}
		s.time = rs.session.netTime
		s.preFightTime = s.netConnection.preFightTime
		//if !rs.session.IsConnected() {
		// for !rs.session.synchronized {
		// 	rs.session.backend.Idle(0)
		// }
		//}
		// s.netConnection.Close()
		rs.session.recording = s.netConnection.recording
		rs.netConnection = s.netConnection
		s.netConnection = nil
	} else if s.netConnection == nil && rs.session == nil {
		session := NewRollbackSesesion(s.cfg.Netplay.Rollback)
		rs.session = &session
		rs.session.InitSyncTest(2)
	}
	rs.session.netTime = 0
	rs.currentFight.reset()

	// These empty frames help the netcode stabilize. Without them, the chances of it desyncing at match start increase a lot
	for i := 0; i < 120; i++ {
		err := rs.session.backend.Idle(
			int(math.Max(0, float64(120))))
		fmt.Printf("difference: %d\n", rs.session.next-rs.session.now-1)
		if err != nil {
			panic(err)
		}

		s.renderFrame() // Do we need to render at this point? Is there anything to render?
		frameTime := rs.session.loopTimer.usToWaitThisLoop()
		running = rs.update(s, frameTime)

		if !running {
			break
		}
	}

	// Loop until end of match
	///fin := false
	for !s.endMatch {
		rs.session.now = time.Now().UnixMilli()
		err := rs.session.backend.Idle(
			int(math.Max(0, float64(rs.session.next-rs.session.now-1))))
		if err != nil {
			panic(err)
		}

		running = rs.runFrame(s)

		if rs.currentFight.fin && (!s.postMatchFlg || len(s.cfg.Common.Lua) == 0) {
			break
		}

		rs.session.next = rs.session.now + 1000/60

		if !running {
			break
		}
		s.renderFrame()
		frameTime := rs.session.loopTimer.usToWaitThisLoop()
		running = rs.update(s, frameTime)

		if !running {
			break
		}
	}
	rs.session.SaveReplay()
	// sys.esc = true
	//sys.rollback.currentFight.fin = true
	s.netConnection = rs.netConnection
	rs.session.backend.Close()

	// Prep for the next match.
	if s.netConnection != nil {
		newSession := NewRollbackSesesion(s.cfg.Netplay.Rollback)
		host := rs.session.host
		remoteIp := rs.session.remoteIp

		rs.session = &newSession
		rs.session.host = host
		rs.session.remoteIp = remoteIp
	} else {
		rs.session = nil
	}
	return false
}

func (rs *RollbackSystem) runFrame(s *System) bool {
	var buffer []byte
	var result error
	if sys.cfg.Netplay.Rollback.DesyncTestFrames > 0 {
		buffer = getAIInputs(1)
		result = rs.session.backend.AddLocalInput(ggpo.PlayerHandle(1), buffer, len(buffer))
	}
	buffer = rs.getTestInputs(0)
	result = rs.session.backend.AddLocalInput(rs.session.currentPlayerHandle, buffer, len(buffer))

	if result == nil {
		var values [][]byte
		disconnectFlags := 0
		values, result = rs.session.backend.SyncInput(&disconnectFlags)
		inputs := decodeInputs(values)
		if rs.session.recording != nil {
			rs.session.SetInput(rs.session.netTime, 0, inputs[0])
			rs.session.SetInput(rs.session.netTime, 1, inputs[1])
			rs.session.netTime++
		}

		if result == nil {

			s.step = false
			//rs.runShortcutScripts(s)

			// If next round
			if !rs.runNextRound(s) {
				return false
			}

			s.bgPalFX.step()
			s.stage.action()

			// If frame is ready to tick and not paused
			//rs.updateStage(s)

			//for i := 0; i < len(inputs) && i < len(sys.commandLists); i++ {
			//	myChar := sys.chars[rs.session.currentPlayerHandle][0]
			//	sys.commandLists[i].Input(myChar.controller, int32(myChar.facing), 0, inputs[i], false)
			//	sys.commandLists[i].Step(int32(myChar.facing), false, false, 0)
			//}

			// Update game state
			rs.action(s, inputs)

			// if rs.handleFlags(s) {
			// 	return true
			// }

			if !rs.updateEvents(s) {
				return false
			}

			// Break if finished
			if rs.currentFight.fin && (!s.postMatchFlg || len(s.cfg.Common.Lua) == 0) {
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

			defer func() {
				if re := recover(); re != nil {
					if rs.session.config.DesyncTest {
						rs.session.log.updateLogs()
						rs.session.log.saveLogs()
						panic("RaiseDesyncError")
					}
				}
			}()

			err := rs.session.backend.AdvanceFrame(rs.session.LiveChecksum(s))
			if err != nil {
				panic(err)
			}

		}
	}
	return true

}

func (rs *RollbackSystem) runShortcutScripts(s *System) {
	for _, v := range s.shortcutScripts {
		if v.Activate {
			if err := s.luaLState.DoString(v.Script); err != nil {
				s.errLog.Println(err.Error())
			}
		}
	}
}

func (rs *RollbackSystem) runNextRound(s *System) bool {
	if s.roundOver() && !rs.currentFight.fin {
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
				tmp.RawSetString("aiLevel", lua.LNumber(p[0].getAILevel()))
				tmp.RawSetString("palno", lua.LNumber(p[0].gi().palno))
				tmp.RawSetString("ratiolevel", lua.LNumber(p[0].ocd().ratioLevel))
				tmp.RawSetString("win", lua.LBool(p[0].win()))
				tmp.RawSetString("winKO", lua.LBool(p[0].winKO()))
				tmp.RawSetString("winTime", lua.LBool(p[0].winTime()))
				tmp.RawSetString("winPerfect", lua.LBool(p[0].winPerfect()))
				tmp.RawSetString("winSpecial", lua.LBool(p[0].winType(WT_Special)))
				tmp.RawSetString("winHyper", lua.LBool(p[0].winType(WT_Hyper)))
				tmp.RawSetString("drawgame", lua.LBool(p[0].drawgame()))
				tmp.RawSetString("ko", lua.LBool(p[0].scf(SCF_ko)))
				tmp.RawSetString("over_ko", lua.LBool(p[0].scf(SCF_over_ko)))
				tbl_roundNo.RawSetInt(p[0].playerNo+1, tmp)
			}
		}
		s.matchData.RawSetInt(int(s.round-1), tbl_roundNo)
		s.scoreRounds = append(s.scoreRounds, [2]float32{s.lifebar.sc[0].scorePoints, s.lifebar.sc[1].scorePoints})
		rs.currentFight.oldTeamLeader = s.teamLeader

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
					rs.currentFight.copyVar(i)
				}
			}
			rs.currentFight.oldWins, rs.currentFight.oldDraws = s.wins, s.draws
			rs.currentFight.oldStageVars.copyStageVars(s.stage)
			rs.currentFight.reset()
		} else {
			/* End match, or prepare for a new character in turns mode */
			for i, tm := range s.tmode {
				if s.chars[i][0].win() || !s.chars[i][0].lose() && tm != TM_Turns {
					for j := i; j < len(s.chars); j += 2 {
						if len(s.chars[j]) > 0 {
							if s.chars[j][0].win() {
								s.chars[j][0].life = Max(1, int32(math.Ceil(math.Pow(rs.currentFight.lvmul,
									float64(rs.currentFight.level[i]))*float64(s.chars[j][0].life))))
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
				return false
			}

			// Otherwise match is over
			s.postMatchFlg = true
			rs.currentFight.fin = true
		}
	}

	return true

}

func (rs *RollbackSystem) updateStage(s *System) {
	// Update stage
	s.stage.action()
}

func (rs *RollbackSystem) action(s *System, input []InputBits) {
	// Clear sprite data
	s.spritesLayerN1 = s.spritesLayerN1[:0]
	s.spritesLayerU = s.spritesLayerU[:0]
	s.spritesLayer0 = s.spritesLayer0[:0]
	s.spritesLayer1 = s.spritesLayer1[:0]
	s.shadows = s.shadows[:0]
	s.reflections = s.reflections[:0]
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

	var x, y, scl float32 = s.cam.Pos[0], s.cam.Pos[1], s.cam.Scale / s.cam.BaseScale()
	s.cam.ResetTracking()

	// Run fight screen
	if s.lifebar.ro.act() {
		if s.intro > s.lifebar.ro.ctrl_time {
			s.intro--
			if s.gsf(GSF_intro) && s.intro <= s.lifebar.ro.ctrl_time {
				s.intro = s.lifebar.ro.ctrl_time + 1
			}
		} else if s.intro > 0 {
			if s.intro == s.lifebar.ro.ctrl_time {
				for _, p := range s.chars {
					if len(p) > 0 {
						if !p[0].asf(ASF_nointroreset) {
							p[0].posReset()
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
							if !p[0].scf(SCF_standby) || p[0].teamside == -1 {
								p[0].setCtrl(true)
								if p[0].ss.no != 0 && !p[0].asf(ASF_nointroreset) {
									p[0].selfState(0, -1, -1, 1, "")
								}
							}
						}
					}
				}
			}
		}
		if s.intro == 0 && s.time > 0 && !s.gsf(GSF_timerfreeze) &&
			(s.supertime <= 0 || !s.superpausebg) && (s.pausetime <= 0 || !s.pausebg) {
			s.time--
		}

		// Check if round ended by KO or time over and set win types
		fin := func() bool {
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
			// KO
			ko := [...]bool{true, true}
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
					}
				}
			}
			// Time over
			ft := s.finishType
			if s.time == 0 {
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
					}
					s.finishType = FT_TO
					s.winTeam = winner
				} else { // Draw game
					s.finishType = FT_TODraw
					s.winTeam = -1
				}
			}
			if s.intro >= -1 && (ko[0] || ko[1]) {
				if ko[0] && ko[1] {
					s.finishType = FT_DKO
					s.winTeam = -1
				} else {
					s.finishType = FT_KO
					s.winTeam = int(Btoi(ko[0]))
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
			return ko[0] || ko[1] || s.time == 0
		}

		// Post round
		if s.roundEnd() || fin() {
			rs4t := -s.lifebar.ro.over_waittime
			fadeoutStart := rs4t - 2 - s.lifebar.ro.over_time + s.lifebar.ro.rt.fadeout_time
			s.intro--
			if s.intro == -s.lifebar.ro.over_hittime && s.finishType != FT_NotYet {
				// Consecutive wins counter
				winner := [...]bool{!s.chars[1][0].win(), !s.chars[0][0].win()}
				if !winner[0] || !winner[1] ||
					s.tmode[0] == TM_Turns || s.tmode[1] == TM_Turns ||
					s.draws >= s.lifebar.ro.match_maxdrawgames[0] ||
					s.draws >= s.lifebar.ro.match_maxdrawgames[1] {
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
			if s.intro > fadeoutStart && s.roundWinTime() && (rs.session.AnyButtonIB(input) && !s.gsf(GSF_roundnotskip)) {
				s.intro = fadeoutStart
				s.winskipped = true
			}
			if s.winskipped || !s.roundWinTime() {
				// Check if game can proceed into roundstate 4
				if s.waitdown > 0 {
					if s.intro == rs4t-1 {
						for _, p := range s.chars {
							if len(p) > 0 {
								// Check if this player is ready to proceed to roundstate 4
								// TODO: The game should normally only wait for players that are active in the fight // || p[0].teamside == -1 || p[0].scf(SCF_standby)
								// TODO: This could be manageable from the char's side with an AssertSpecial or such
								if p[0].scf(SCF_over_alive) || p[0].scf(SCF_over_ko) ||
									(p[0].scf(SCF_ctrl) && p[0].ss.moveType == MT_I && p[0].ss.stateType != ST_A && p[0].ss.stateType != ST_L) {
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
				// Start running wintime counter only after getting into roundstate 4
				if s.intro < rs4t && !s.roundWinTime() {
					s.wintime--
				}
				// Set characters into win/lose poses, update win counters
				if s.roundWinStates() {
					if s.waitdown >= 0 {
						winner := [...]bool{!s.chars[1][0].win(), !s.chars[0][0].win()}
						if !winner[0] || !winner[1] ||
							s.tmode[0] == TM_Turns || s.tmode[1] == TM_Turns ||
							s.draws >= s.lifebar.ro.match_maxdrawgames[0] ||
							s.draws >= s.lifebar.ro.match_maxdrawgames[1] {
							for i, win := range winner {
								if win {
									s.lifebar.wi[i].add(s.winType[i])
									if s.matchOver() {
										// In a draw game both players go back to 0 wins
										if winner[0] == winner[1] { // sys.winTeam < 0
											s.lifebar.wc[0].wins = 0
											s.lifebar.wc[1].wins = 0
										} else {
											if s.wins[i] >= s.matchWins[i] {
												s.lifebar.wc[i].wins += 1
											}
										}
									}
								}
							}
						} else {
							s.draws++
						}
					}
					for _, p := range s.chars {
						if len(p) > 0 {
							// Default life recovery. Used only if externalized Lua implementation is disabled
							if len(sys.cfg.Common.Lua) == 0 && s.waitdown >= 0 && s.time > 0 && p[0].win() &&
								p[0].alive() && !s.matchOver() &&
								(s.tmode[0] == TM_Turns || s.tmode[1] == TM_Turns) {
								p[0].life += int32((float32(p[0].lifeMax) *
									float32(s.time) / 60) * s.turnsRecoveryRate)
								if p[0].life > p[0].lifeMax {
									p[0].life = p[0].lifeMax
								}
							}
							// TODO: These changestates ought to be unhardcoded
							if !p[0].scf(SCF_over_alive) && !p[0].hitPause() && p[0].alive() && p[0].animNo != 5 {
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
					s.waitdown = 0
				}
				s.waitdown--
			}
			// If the game can't proceed to the fadeout screen, we turn back the counter 1 tick
			if !s.winskipped && s.gsf(GSF_roundnotover) &&
				s.intro == rs4t-2-s.lifebar.ro.over_time+s.lifebar.ro.rt.fadeout_time {
				s.intro++
			}
			// Start fadeout effect
			if s.intro == fadeoutStart {
				s.lifebar.ro.rt.fadeoutTimer = s.lifebar.ro.rt.fadeout_time
			}
		} else if s.intro < 0 {
			s.intro = 0
		}
	}

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
		if s.superanim != nil {
			s.superanim.Action()
		}
		rs.rollbackAction(s, &s.charList, input)
		s.nomusic = s.gsf(GSF_nomusic) && !sys.postMatchFlg
	}

	// This function runs every tick
	// It should be placed between "tick frame" and "tick next frame"
	s.charUpdate()

	// Update lifebars
	// This must happen before hit detection for accurate display
	// Allows a combo to still end if a character is hit in the same frame where it exits movetype H
	s.lifebar.step()
	if s.tickNextFrame() {
		s.globalCollision() // This could perhaps happen during "tick frame" instead? Would need more testing
		s.charList.tick()
	}

	// Run camera
	x, y, scl = s.cam.action(x, y, scl, s.supertime > 0 || s.pausetime > 0)

	// Skip character intros on button press and play the shutter effect
	if s.tickNextFrame() {
		if s.lifebar.ro.current < 1 && !s.introSkipped {
			// Checking the intro flag prevents skipping intros when they don't exist
			if s.lifebar.ro.rt.shutterTimer == 0 &&
				rs.session.AnyButtonIB(input) && s.gsf(GSF_intro) && !s.gsf(GSF_roundnotskip) && s.intro > s.lifebar.ro.ctrl_time {
				// Start shutter effect
				s.lifebar.ro.rt.shutterTimer = s.lifebar.ro.rt.shutter_time * 2 // Open + close time
			}
			// Do the actual skipping halfway into the shutter animation, when it's closed
			if s.lifebar.ro.rt.shutterTimer == s.lifebar.ro.rt.shutter_time {
				// SkipRoundDisplay and SkipFightDisplay flags must be preserved during intro skip frame
				skipround := (s.specialFlag&GSF_skiprounddisplay | s.specialFlag&GSF_skipfightdisplay)
				s.resetGblEffect()
				s.specialFlag = skipround
				s.intro = s.lifebar.ro.ctrl_time
				for i, p := range s.chars {
					if len(p) > 0 {
						s.clearPlayerAssets(i, false)
						p[0].posReset()
						p[0].selfState(0, -1, -1, 0, "")
					}
				}
				s.introSkipped = true
			}
		}
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
	// Superpause effect
	if s.superanim != nil {
		s.spritesLayer1.add(&SprData{
			anim:         s.superanim,
			fx:           &s.superpmap,
			pos:          s.superpos,
			scl:          s.superscale,
			alpha:        [2]int32{-1},
			priority:     5,
			rot:          Rotation{},
			screen:       false,
			undarken:     true,
			oldVer:       s.cgi[s.superplayerno].mugenver[0] != 1,
			facing:       1,
			airOffsetFix: [2]float32{1, 1},
			projection:   0,
			fLength:      0,
			window:       [4]float32{0, 0, 0, 0},
		})
		if s.superanim.loopend {
			s.superanim = nil // Not allowed to loop
		}
	}
	for i := range s.projs {
		for j := range s.projs[i] {
			if s.projs[i][j].id >= 0 {
				s.projs[i][j].cueDraw(s.cgi[i].mugenver[0] != 1)
			}
		}
	}
	s.charList.cueDraw()
	explUpdate := func(edl *[len(s.chars)][]int, drop bool) {
		for i, el := range *edl {
			for j := len(el) - 1; j >= 0; j-- {
				if el[j] >= 0 {
					s.explods[i][el[j]].update(s.cgi[i].mugenverF, i)
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
	explUpdate(&s.explodsLayerN1, true)
	explUpdate(&s.explodsLayer0, true)
	explUpdate(&s.explodsLayer1, false)
	// Adjust game speed
	if s.tickNextFrame() {
		spd := (60 + s.cfg.Options.GameSpeed*5) / float32(s.cfg.Config.Framerate) * s.accel
		// KO slowdown
		s.slowtimeTrigger = 0
		if s.intro < 0 && s.time != 0 && s.slowtime > 0 {
			if !s.gsf(GSF_nokoslow) {
				spd *= s.lifebar.ro.slow_speed
				if s.slowtime < s.lifebar.ro.slow_fadetime {
					spd += (float32(1) - s.lifebar.ro.slow_speed) * float32(s.lifebar.ro.slow_fadetime-s.slowtime) / float32(s.lifebar.ro.slow_fadetime)
				}
			}
			s.slowtimeTrigger = s.slowtime
			s.slowtime--
		}
		// Outside match or while frame stepping
		if s.postMatchFlg || s.step {
			spd = 1
		}
		s.turbo = spd
	}
	s.tickSound()
	return
}

func (rs *RollbackSystem) handleFlags(s *System) bool {
	// F4 pressed to restart round
	if s.roundResetFlg && !s.postMatchFlg {
		rs.currentFight.reset()
	}
	// Shift+F4 pressed to restart match
	if s.reloadFlg {
		return true
	}
	return false

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

func (rs *RollbackSystem) update(s *System, wait time.Duration) bool {
	s.frameCounter++
	return rs.await(s, wait)
}

func (rs *RollbackSystem) await(s *System, wait time.Duration) bool {
	if !s.frameSkip {
		// Render the finished frame
		gfx.EndFrame()
		s.window.SwapBuffers()
		// Begin the next frame after events have been processed. Do not clear
		// the screen if network input is present.
		defer gfx.BeginFrame(sys.netConnection == nil)
	}
	s.runMainThreadTask()
	now := time.Now()
	diff := s.redrawWait.nextTime.Sub(now)
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
	return !s.gameEnd
}

func (rs *RollbackSystem) commandUpdate(ib []InputBits, sys *System) {
	// Iterate players
	for i, p := range sys.chars {
		if len(p) > 0 {
			root := p[0]
			// Select a random command for AI cheating
			// The way this only allows one command to be cheated at a time may be the cause of issue #2022
			cheat := int32(-1)
			if root.controller < 0 {
				if sys.roundState() == 2 && RandF32(0, sys.aiLevel[i]/2+32) > 32 { // TODO: Balance AI scaling
					cheat = Rand(0, int32(len(root.cmd[root.ss.sb.playerNo].Commands))-1)
				}
			}
			// Iterate root and helpers
			for _, c := range p {
				act := true
				if sys.supertime > 0 {
					act = c.superMovetime != 0
				} else if sys.pausetime > 0 && c.pauseMovetime == 0 {
					act = false
				}
				// Auto turning check for the root
				// Having this here makes B and F inputs reverse the same instant the character turns
				if act && c.helperIndex == 0 && (c.scf(SCF_ctrl) || sys.roundState() > 2) &&
					(c.ss.no == 0 || c.ss.no == 11 || c.ss.no == 20 || c.ss.no == 52) {
					c.autoTurn()
				}

				// Update Forward/Back flipping flag
				c.updateFBFlip()

				// Rollback part
				if c.helperIndex == 0 || c.helperIndex > 0 && &c.cmd[0] != &root.cmd[0] {

					if i < len(ib) {
						if sys.gameMode == "watch" && (c.controller < 0 && ^c.controller < len(sys.aiInput)) {
							sys.aiInput[^c.controller].Update(sys.aiLevel[i])
						}
						// if we have an input from the players
						// update the command buffer based on that.
						ib[i].RollbackBitsToKeys(c.cmd[0].Buffer, int32(c.facing))
					} else if (sys.tmode[0] == TM_Tag || sys.tmode[1] == TM_Tag) && (c.teamside != -1) {
						ib[c.teamside].RollbackBitsToKeys(c.cmd[0].Buffer, int32(c.facing))
					} else {
						// Otherwise, this will ostensibly update the buffers based on AIInput
						c.cmd[0].InputUpdate(c.controller, c.fbFlip, sys.aiLevel[i], c.inputFlag, c.inputShift, false)
					}
				}

				// Clear input buffers and skip the rest of the loop
				// This used to apply only to the root, but that caused some issues with helper-based custom input systems
				if c.inputWait() || c.asf(ASF_noinput) {
					for i := range c.cmd {
						c.cmd[i].BufReset()
					}
					continue
				}
				hpbuf := false
				pausebuf := false
				winbuf := false
				// Buffer during hitpause
				if c.hitPause() && c.gi().constants["input.pauseonhitpause"] != 0 { // TODO: Deprecated constant
					hpbuf = true
					// In Winmugen, commands were buffered for one extra frame after hitpause (but not after Pause/SuperPause)
					// This was fixed in Mugen 1.0
					if c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0 && c.stWgi().mugenver[0] != 1 {
						winbuf = true
					}
				}
				// Buffer during Pause and SuperPause
				if sys.supertime > 0 {
					if !act && sys.supertime <= sys.superendcmdbuftime {
						pausebuf = true
					}
				} else if sys.pausetime > 0 {
					if !act && sys.pausetime <= sys.pauseendcmdbuftime {
						pausebuf = true
					}
				}
				// Update commands
				for i := range c.cmd {
					extratime := Btoi(hpbuf || pausebuf) + Btoi(winbuf)
					helperbug := c.helperIndex != 0 && c.stWgi().ikemenver[0] == 0 && c.stWgi().ikemenver[1] == 0
					c.cmd[i].Step(c.controller < 0, helperbug, hpbuf, pausebuf, extratime)
				}
				// Enable AI cheated command
				c.cpucmd = cheat
			}
		}
	}
}

func (rs *RollbackSystem) rollbackAction(sys *System, cl *CharList, ib []InputBits) {
	// Update commands for all chars
	rs.commandUpdate(ib, sys)

	// Prepare characters before performing their actions
	for i := 0; i < len(cl.runOrder); i++ {
		cl.runOrder[i].actionPrepare()
	}

	// Run actions for each character in the sorted list
	// Sorting the characters first makes new helpers wait for their turn and allows RunOrder trigger accuracy
	sortedOrder := cl.sortActionRunOrder()
	for i := 0; i < len(sortedOrder); i++ {
		if sortedOrder[i] < len(cl.runOrder) {
			cl.runOrder[sortedOrder[i]].actionRun()
		}
	}

	// Run actions for anyone missed (new helpers)
	extra := len(sortedOrder) + 1
	for i := 0; i < len(cl.runOrder); i++ {
		if cl.runOrder[i].runorder < 0 {
			cl.runOrder[i].runorder = int32(extra)
			cl.runOrder[i].actionRun()
			extra++
		}
	}

	// Finish performing character actions
	for i := 0; i < len(cl.runOrder); i++ {
		cl.runOrder[i].actionFinish()
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

func (f *Fight) copyVar(pn int) {
	f.life[pn] = sys.chars[pn][0].life
	f.lifeMax[pn] = sys.chars[pn][0].lifeMax
	f.power[pn] = sys.chars[pn][0].power
	f.powerMax[pn] = sys.chars[pn][0].powerMax
	f.guardPoints[pn] = sys.chars[pn][0].guardPoints
	f.guardPointsMax[pn] = sys.chars[pn][0].guardPointsMax
	f.dizzyPoints[pn] = sys.chars[pn][0].dizzyPoints
	f.dizzyPointsMax[pn] = sys.chars[pn][0].dizzyPointsMax
	f.redLife[pn] = sys.chars[pn][0].redLife
	f.teamside[pn] = sys.chars[pn][0].teamside

	f.cnsvar[pn] = make(map[int32]int32)
	for k, v := range sys.chars[pn][0].cnsvar {
		f.cnsvar[pn][k] = v
	}

	f.cnsfvar[pn] = make(map[int32]float32)
	for k, v := range sys.chars[pn][0].cnsfvar {
		f.cnsfvar[pn][k] = v
	}

	f.mapArray[pn] = make(map[string]float32)
	for k, v := range sys.chars[pn][0].mapArray {
		f.mapArray[pn][k] = v
	}

	copy(f.dialogue[pn], sys.chars[pn][0].dialogue[:])

	f.remapSpr[pn] = make(RemapPreset)
	for k, v := range sys.chars[pn][0].remapSpr {
		f.remapSpr[pn][k] = v
	}
}

func (f *Fight) reset() {
	sys.wins, sys.draws = f.oldWins, f.oldDraws
	sys.teamLeader = f.oldTeamLeader
	for i, p := range sys.chars {
		if len(p) > 0 {
			p[0].life = f.life[i]
			p[0].lifeMax = f.lifeMax[i]
			p[0].power = f.power[i]
			p[0].powerMax = f.powerMax[i]
			p[0].guardPoints = f.guardPoints[i]
			p[0].guardPointsMax = f.guardPointsMax[i]
			p[0].dizzyPoints = f.dizzyPoints[i]
			p[0].dizzyPointsMax = f.dizzyPointsMax[i]
			p[0].redLife = f.redLife[i]
			p[0].teamside = f.teamside[i]
			p[0].cnsvar = make(map[int32]int32)
			for k, v := range f.cnsvar[i] {
				p[0].cnsvar[k] = v
			}
			p[0].cnsfvar = make(map[int32]float32)
			for k, v := range f.cnsfvar[i] {
				p[0].cnsfvar[k] = v
			}
			copy(p[0].dialogue[:], f.dialogue[i])
			p[0].mapArray = make(map[string]float32)
			for k, v := range f.mapArray[i] {
				p[0].mapArray[k] = v
			}
			p[0].remapSpr = make(RemapPreset)
			for k, v := range f.remapSpr[i] {
				p[0].remapSpr[k] = v
			}
		}
	}
	sys.stage.copyStageVars(&f.oldStageVars)
	sys.resetFrameTime()
	sys.nextRound()
	sys.roundResetFlg, sys.introSkipped = false, false
	sys.reloadFlg, sys.reloadStageFlg, sys.reloadLifebarFlg = false, false, false
	sys.runMainThreadTask()
	gfx.Await()
}

func (f *Fight) endFight() {
	sys.oldNextAddTime = 1
	sys.nomusic = false
	sys.allPalFX.clear()
	sys.allPalFX.enable = false
	for i, p := range sys.chars {
		if len(p) > 0 {
			sys.clearPlayerAssets(i, sys.matchOver() || (sys.tmode[i&1] == TM_Turns && p[0].life <= 0))
		}
	}
	sys.wincnt.update()
}

func (f *Fight) initChars() {
	// Initialize each character
	lvmul := math.Pow(2, 1.0/12)
	for i, p := range sys.chars {
		if len(p) > 0 {
			// Get max life, and adjust based on team mode
			var lm float32
			if p[0].ocd().lifeMax != -1 {
				lm = float32(p[0].ocd().lifeMax) * p[0].ocd().lifeRatio * sys.cfg.Options.Life / 100
			} else {
				lm = float32(p[0].gi().data.life) * p[0].ocd().lifeRatio * sys.cfg.Options.Life / 100
			}
			if p[0].teamside != -1 {
				switch sys.tmode[i&1] {
				case TM_Single:
					switch sys.tmode[(i+1)&1] {
					case TM_Simul, TM_Tag:
						lm *= sys.cfg.Options.Team.SingleVsTeamLife / 100
					case TM_Turns:
						if sys.numTurns[(i+1)&1] < sys.matchWins[(i+1)&1] && sys.cfg.Options.Team.LifeShare {
							lm = lm * float32(sys.numTurns[(i+1)&1]) /
								float32(sys.matchWins[(i+1)&1])
						}
					}
				case TM_Simul, TM_Tag:
					switch sys.tmode[(i+1)&1] {
					case TM_Simul, TM_Tag:
						if sys.numSimul[(i+1)&1] < sys.numSimul[i&1] && sys.cfg.Options.Team.LifeShare {
							lm = lm * float32(sys.numSimul[(i+1)&1]) / float32(sys.numSimul[i&1])
						}
					case TM_Turns:
						if sys.numTurns[(i+1)&1] < sys.numSimul[i&1]*sys.matchWins[(i+1)&1] && sys.cfg.Options.Team.LifeShare {
							lm = lm * float32(sys.numTurns[(i+1)&1]) /
								float32(sys.numSimul[i&1]*sys.matchWins[(i+1)&1])
						}
					default:
						if sys.cfg.Options.Team.LifeShare {
							lm /= float32(sys.numSimul[i&1])
						}
					}
				case TM_Turns:
					switch sys.tmode[(i+1)&1] {
					case TM_Single:
						if sys.matchWins[i&1] < sys.numTurns[i&1] && sys.cfg.Options.Team.LifeShare {
							lm = lm * float32(sys.matchWins[i&1]) / float32(sys.numTurns[i&1])
						}
					case TM_Simul, TM_Tag:
						if sys.numSimul[(i+1)&1]*sys.matchWins[i&1] < sys.numTurns[i&1] && sys.cfg.Options.Team.LifeShare {
							lm = lm * sys.cfg.Options.Team.SingleVsTeamLife / 100 *
								float32(sys.numSimul[(i+1)&1]*sys.matchWins[i&1]) /
								float32(sys.numTurns[i&1])
						}
					case TM_Turns:
						if sys.numTurns[(i+1)&1] < sys.numTurns[i&1] && sys.cfg.Options.Team.LifeShare {
							lm = lm * float32(sys.numTurns[(i+1)&1]) / float32(sys.numTurns[i&1])
						}
					}
				}
			}
			foo := math.Pow(lvmul, float64(-f.level[i]))
			p[0].lifeMax = Max(1, int32(math.Floor(foo*float64(lm))))

			if p[0].roundsExisted() > 0 {
				// If character already existed for a round, presumably because of turns mode, just update life
				p[0].life = Min(p[0].lifeMax, int32(math.Ceil(foo*float64(p[0].life))))
			} else if sys.round == 1 || sys.tmode[i&1] == TM_Turns {
				// If round 1 or a new character in turns mode, initialize values
				if p[0].ocd().life != -1 {
					p[0].life = Clamp(p[0].ocd().life, 0, p[0].lifeMax)
					p[0].redLife = p[0].life
				} else {
					p[0].life = p[0].lifeMax
					p[0].redLife = p[0].lifeMax
				}
				if sys.round == 1 {
					if sys.maxPowerMode {
						p[0].power = p[0].powerMax
					} else if p[0].ocd().power != -1 {
						p[0].power = Clamp(p[0].ocd().power, 0, p[0].powerMax)
					} else if !sys.consecutiveRounds || sys.consecutiveWins[0] == 0 {
						p[0].power = 0
					}
				}
				p[0].power = Clamp(p[0].power, 0, p[0].powerMax) // Because of previous partner in Turns mode
				p[0].dialogue = []string{}
				p[0].mapArray = make(map[string]float32)
				for k, v := range p[0].mapDefault {
					p[0].mapArray[k] = v
				}
				p[0].remapSpr = make(RemapPreset)
			}

			if p[0].ocd().guardPoints != -1 {
				p[0].guardPoints = Clamp(p[0].ocd().guardPoints, 0, p[0].guardPointsMax)
			} else {
				p[0].guardPoints = p[0].guardPointsMax
			}
			if p[0].ocd().dizzyPoints != -1 {
				p[0].dizzyPoints = Clamp(p[0].ocd().dizzyPoints, 0, p[0].dizzyPointsMax)
			} else {
				p[0].dizzyPoints = p[0].dizzyPointsMax
			}
			f.copyVar(i)
		}
	}
}

func (f *Fight) initSuperMeter() {
	// Prepare next round for all players
	for _, p := range sys.chars {
		if len(p) > 0 {
			p[0].prepareNextRound()
		}
	}

	for i, p := range sys.chars {
		if len(p) > 0 && p[0].teamside != -1 {
			f.level[i] = sys.wincnt.getLevel(i)
			if sys.cfg.Options.Team.PowerShare {
				pmax := Max(sys.cgi[i&1].data.power, sys.cgi[i].data.power)
				for j := i & 1; j < MaxSimul*2; j += 2 {
					if len(sys.chars[j]) > 0 {
						sys.chars[j][0].powerMax = pmax
					}
				}
			}
		}
	}
}

func (f *Fight) initTeamsLevels() {
	minlv, maxlv := f.level[0], f.level[0]
	for i, lv := range f.level[1:] {
		if len(sys.chars[i+1]) > 0 {
			minlv = Min(minlv, lv)
			maxlv = Max(maxlv, lv)
		}
	}
	if minlv > 0 {
		for i := range f.level {
			f.level[i] -= minlv
		}
	} else if maxlv < 0 {
		for i := range f.level {
			f.level[i] -= maxlv
		}
	}
}

func NewFight() Fight {
	f := Fight{}
	f.oldStageVars.copyStageVars(sys.stage)
	f.life = make([]int32, len(sys.chars))
	f.lifeMax = make([]int32, len(sys.chars))
	f.pow = make([]int32, len(sys.chars))
	f.gpow = make([]int32, len(sys.chars))
	f.spow = make([]int32, len(sys.chars))
	f.rlife = make([]int32, len(sys.chars))
	f.cnsvar = make([]map[int32]int32, len(sys.chars))
	f.cnsfvar = make([]map[int32]float32, len(sys.chars))
	f.power = make([]int32, len(sys.chars))
	f.powerMax = make([]int32, len(sys.chars))
	f.guardPoints = make([]int32, len(sys.chars))
	f.guardPointsMax = make([]int32, len(sys.chars))
	f.dizzyPoints = make([]int32, len(sys.chars))
	f.dizzyPointsMax = make([]int32, len(sys.chars))
	f.redLife = make([]int32, len(sys.chars))
	f.teamside = make([]int, len(sys.chars))
	f.dialogue = make([][]string, len(sys.chars))
	f.mapArray = make([]map[string]float32, len(sys.chars))
	f.remapSpr = make([]RemapPreset, len(sys.chars))
	f.level = make([]int32, len(sys.chars))
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

// HACK: So you won't be playing eachothers characters
func reverseInputs(inputs []InputBits) []InputBits {
	for i, j := 0, len(inputs)-1; i < j; i, j = i+1, j-1 {
		inputs[i], inputs[j] = inputs[j], inputs[i]
	}
	return inputs
}

func writeI32(i32 int32) []byte {
	b := []byte{byte(i32), byte(i32 >> 8), byte(i32 >> 16), byte(i32 >> 24)}
	return b
}

func (rs *RollbackSystem) getInputs(player int) []byte {
	var ib InputBits
	ib.KeysToBits(rs.netConnection.buf[player].InputReader.LocalInput(player, false))
	return writeI32(int32(ib))
}

func (rs *RollbackSystem) getTestInputs(player int) []byte {
	var ib InputBits
	ib.KeysToBits(sys.chars[0][0].cmd[0].Buffer.InputReader.LocalInput(player, false))
	return writeI32(int32(ib))
}

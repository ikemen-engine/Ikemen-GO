local start = {}

--team side specific data storage
start.p = {{}, {}}
--cell data storage
start.c = {}
for i = 1, --[[config.Players]]8 do
	table.insert(start.c, {selX = 0, selY = 0, cell = -1, randCnt = 0, randRef = nil})
end
--globally accessible temp data
start.challenger = 0
--local variables
local restoreCursor = false
local selScreenEnd = false
local stageEnd = false
local stageRandom = false
local stageListNo = 0
local t_aiRamp = {}
local t_gameStats = {}
local t_recordText = {}
local t_reservedChars = {{}, {}}
local timerSelect = 0

--;===========================================================
--; COMMON FUNCTIONS
--;===========================================================
--; ROSTER
--converts '.maxmatches' style table (key = order, value = max matches) to the same structure as '.ratiomatches' (key = match number, value = subtable with char num and order data)
function start.f_unifySettings(t, t_chars)
	local ret = {}
	for i = 1, #t do --for each order number
		if t_chars[i] ~= nil then --only if there are any characters available with this order
			local infinite = false
			local num = t[i]
			if num == -1 then --infinite matches
				num = #t_chars[i] --assign max amount of characters with this order
				infinite = true
			end
			for j = 1, num do --iterate up to max amount of matches versus characters with this order
				--[[if j * start.p[2].numChars > #t_chars[i] and #ret > 0 then --if there are not enough characters to fill all slots and at least 1 fight is already assigned
					local stop = true
					for k = (j - 1) * start.p[2].numChars + 1, #t_chars[i] do --loop through characters left for this match
						if start.f_getCharData(t_chars[i][k]).single == 1 then --and allow appending if any of the remaining characters has 'single' flag set
							stop = false
						end
					end
					if stop then
						break
					end
				end]]
				table.insert(ret, {['rmin'] = start.p[2].numChars, ['rmax'] = start.p[2].numChars, ['order'] = i})
			end
			if infinite then
				table.insert(ret, {['rmin'] = start.p[2].numChars, ['rmax'] = start.p[2].numChars, ['order'] = -1})
				break --no point in appending additional matches
			end
		end
	end
	return ret
end

-- start.t_makeRoster is a table storing functions returning table data used
-- by start.f_makeRoster function, depending on game mode. Can be appended via
-- external module, without conflicting with default scripts.
start.t_makeRoster = {}
start.t_makeRoster.arcade = function()
	if start.p[2].ratio then --Ratio
		if start.f_getCharData(start.p[1].t_selected[1].ref).ratiomatches ~= nil and main.t_selOptions[start.f_getCharData(start.p[1].t_selected[1].ref).ratiomatches .. '_arcaderatiomatches'] ~= nil then --custom settings exists as char param
			return main.t_selOptions[start.f_getCharData(start.p[1].t_selected[1].ref).ratiomatches .. '_arcaderatiomatches'], main.t_orderChars
		else --default settings
			return main.t_selOptions.arcaderatiomatches, main.t_orderChars
		end
	elseif start.p[2].teamMode == 0 then --Single
		if start.f_getCharData(start.p[1].t_selected[1].ref).maxmatches ~= nil and main.t_selOptions[start.f_getCharData(start.p[1].t_selected[1].ref).maxmatches .. '_arcademaxmatches'] ~= nil then --custom settings exists as char param
			return start.f_unifySettings(main.t_selOptions[start.f_getCharData(start.p[1].t_selected[1].ref).maxmatches .. '_arcademaxmatches'], main.t_orderChars), main.t_orderChars
		else --default settings
			return start.f_unifySettings(main.t_selOptions.arcademaxmatches, main.t_orderChars), main.t_orderChars
		end
	else --Simul / Turns / Tag
		if start.f_getCharData(start.p[1].t_selected[1].ref).maxmatches ~= nil and main.t_selOptions[start.f_getCharData(start.p[1].t_selected[1].ref).maxmatches .. '_teammaxmatches'] ~= nil then --custom settings exists as char param
			return start.f_unifySettings(main.t_selOptions[start.f_getCharData(start.p[1].t_selected[1].ref).maxmatches .. '_teammaxmatches'], main.t_orderChars), main.t_orderChars
		else --default settings
			return start.f_unifySettings(main.t_selOptions.teammaxmatches, main.t_orderChars), main.t_orderChars
		end
	end
end
start.t_makeRoster.teamcoop = start.t_makeRoster.arcade
start.t_makeRoster.netplayteamcoop = start.t_makeRoster.arcade
start.t_makeRoster.timeattack = start.t_makeRoster.arcade
start.t_makeRoster.survival = function()
	if start.f_getCharData(start.p[1].t_selected[1].ref).maxmatches ~= nil and main.t_selOptions[start.f_getCharData(start.p[1].t_selected[1].ref).maxmatches .. '_survivalmaxmatches'] ~= nil then --custom settings exists as char param
		return start.f_unifySettings(main.t_selOptions[start.f_getCharData(start.p[1].t_selected[1].ref).maxmatches .. '_survivalmaxmatches'], main.t_orderSurvival), main.t_orderSurvival
	else --default settings
		return start.f_unifySettings(main.t_selOptions.survivalmaxmatches, main.t_orderSurvival), main.t_orderSurvival
	end
end
start.t_makeRoster.survivalcoop = start.t_makeRoster.survival
start.t_makeRoster.netplaysurvivalcoop = start.t_makeRoster.survival

-- generates roster table
function start.f_makeRoster(t_ret)
	t_ret = t_ret or {}
	--prepare correct settings tables
	if start.t_makeRoster[gamemode()] == nil then
		panicError("\n" .. gamemode() .. " game mode unrecognized by start.f_makeRoster()\n")
	end
	local t, t_static = start.t_makeRoster[gamemode()]()
	--generate roster
	local t_removable = main.f_tableCopy(t_static) --copy into editable order table
	for i = 1, #t do --for each match number
		if t[i].order == -1 then --infinite matches for this order detected
			table.insert(t_ret, {-1}) --append infinite matches flag at the end
			break
		end
		if t_removable[t[i].order] ~= nil then
			if #t_removable[t[i].order] == 0 and main.forceRosterSize then
				t_removable = main.f_tableCopy(t_static) --allows character repetition, if needed to fill whole roster
			end
			if #t_removable[t[i].order] >= 1 then --there is at least 1 character with this order available
				local remaining = t[i].rmin - #t_removable[t[i].order]
				table.insert(t_ret, {}) --append roster table with new subtable
				local t_toinsert = {}
				local t_removableTemp = main.f_tableCopy(t_removable)
				for j = 1, math.random(math.min(t[i].rmin, #t_removableTemp[t[i].order]), math.min(t[i].rmax, #t_removableTemp[t[i].order])) do --for randomized characters count
					local rand = math.random(1, #t_removableTemp[t[i].order]) --randomize which character will be taken
					local ref = t_removableTemp[t[i].order][rand]
					if not main.charparam.single or not start.f_getCharData(ref).single then
						table.insert(t_toinsert, ref) --append character if 'single' param is not blocking larger team size
						table.remove(t_removableTemp[t[i].order], rand) --remove it from the t_removableTemp table
					else --otherwise only this character is added to roster
						t_toinsert = {ref}
						remaining = 0
						break
					end
				end
				for _, v in ipairs(t_toinsert) do
					table.insert(t_ret[#t_ret], v) --add such character into roster subtable
					main.f_tableRemove(t_removable[t[i].order], v) --and remove it from the available character pool
				end
				--fill the remaining slots randomly if there are not enough players available with this order
				while remaining > 0 do
					table.insert(t_ret[#t_ret], t_static[t[i].order][math.random(1, #t_static[t[i].order])])
					remaining = remaining - 1
				end
			end
		end
	end
	if main.debugLog then main.f_printTable(t_ret, 'debug/t_roster.txt') end
	return t_ret
end

--;===========================================================
--; AI RAMPING
-- start.t_aiRampData is a table storing functions returning variable data used
-- by start.f_aiRamp function, depending on game mode. Can be appended via
-- external module, without conflicting with default scripts.
start.t_aiRampData = {}
start.t_aiRampData.arcade = function()
	if start.p[2].teamMode == 0 then --Single
		return main.t_selOptions.arcadestart.wins, main.t_selOptions.arcadestart.offset, main.t_selOptions.arcadeend.wins, main.t_selOptions.arcadeend.offset
	elseif start.p[2].ratio then --Ratio
		return main.t_selOptions.ratiostart.wins, main.t_selOptions.ratiostart.offset, main.t_selOptions.ratioend.wins, main.t_selOptions.ratioend.offset
	else --Simul / Turns / Tag
		return main.t_selOptions.teamstart.wins, main.t_selOptions.teamstart.offset, main.t_selOptions.teamend.wins, main.t_selOptions.teamend.offset
	end
end
start.t_aiRampData.teamcoop = start.t_aiRampData.arcade
start.t_aiRampData.netplayteamcoop = start.t_aiRampData.arcade
start.t_aiRampData.timeattack = start.t_aiRampData.arcade
start.t_aiRampData.survival = function()
	return main.t_selOptions.survivalstart.wins, main.t_selOptions.survivalstart.offset, main.t_selOptions.survivalend.wins, main.t_selOptions.survivalend.offset
end
start.t_aiRampData.survivalcoop = start.t_aiRampData.survival
start.t_aiRampData.netplaysurvivalcoop = start.t_aiRampData.survival

-- generates AI ramping table
function start.f_aiRamp(currentMatch)
	if start.t_aiRampData[gamemode()] == nil then
		panicError("\n" .. gamemode() .. " game mode unrecognized by start.f_aiRamp()\n")
	end
	local start_match, start_diff, end_match, end_diff = start.t_aiRampData[gamemode()]()
	local startAI = config.Difficulty + start_diff
	if startAI > 8 then
		startAI = 8
	elseif startAI < 1 then
		startAI = 1
	end
	local endAI = config.Difficulty + end_diff
	if endAI > 8 then
		endAI = 8
	elseif endAI < 1 then
		endAI = 1
	end
	if currentMatch == 1 then
		t_aiRamp = {}
	end
	for i = math.min(#t_aiRamp, currentMatch), math.max(#start.t_roster, currentMatch) do
		if i - 1 <= start_match then
			table.insert(t_aiRamp, startAI)
		elseif i - 1 <= end_match then
			local curMatch = i - (start_match + 1)
			table.insert(t_aiRamp, curMatch * (endAI - startAI) / (end_match - start_match) + startAI)
		else
			table.insert(t_aiRamp, endAI)
		end
	end
	if main.debugLog then main.f_printTable(t_aiRamp, 'debug/t_aiRamp.txt') end
end
--;===========================================================

--calculates AI level
function start.f_difficulty(player, offset)
	local t = {}
	if main.f_playerSide(player) == 1 then
		t = start.f_getCharData(start.p[1].t_selected[math.floor(player / 2 + 0.5)].ref)
	else
		t = start.f_getCharData(start.p[2].t_selected[math.floor(player / 2)].ref)
	end
	if t.ai ~= nil then
		return t.ai
	else
		return config.Difficulty + offset
	end
end

--assigns AI level, remaps input
function start.f_remapAI(ai)
	--Offset
	local offset = 0
	if config.AIRamping and main.aiRamp then
		if t_aiRamp[matchno()] == nil then
			start.f_aiRamp(matchno())
		end
		offset = t_aiRamp[matchno()] - config.Difficulty
	end
	local t_ex = {}
	for side = 1, 2 do
		if main.coop then
			for k, v in ipairs(start.p[side].t_selCmd) do
				if gamemode('versuscoop') then
					remapInput(v.player, v.cmd)
					setCom(v.player, 0)
					t_ex[v.player] = true
				else
					local pn = v.player * 2 - 1
					remapInput(pn, v.cmd)
					setCom(pn, 0)
					t_ex[pn] = true
				end
			end
		end
		if start.p[side].teamMode == 0 or start.p[side].teamMode == 2 then --Single or Turns
			if (main.t_pIn[side] == side and not main.cpuSide[side] and not main.coop) or start.challenger > 0 or gamemode('training') then
				setCom(side, 0)
			else
				setCom(side, ai or start.f_difficulty(side, offset))
			end
		elseif start.p[side].teamMode == 1 then --Simul
			if not t_ex[side] then
				if (main.t_pIn[side] == side and not main.cpuSide[side] and not main.coop) or start.challenger > 0 then
					setCom(side, 0)
				else
					setCom(side, ai or start.f_difficulty(side, offset))
				end
			end
			for i = side + 2, #start.p[side].t_selected * 2 do
				if not t_ex[i] and (i - 1) % 2 + 1 == side then
					remapInput(i, side) --P3/5/7 => P1 controls, P4/6/8 => P2 controls
					setCom(i, ai or start.f_difficulty(i, offset))
				end
			end
		else --Tag
			for i = side, #start.p[side].t_selected * 2 do
				if not t_ex[i] and (i - 1) % 2 + 1 == side then
					if (main.t_pIn[side] == side and not main.cpuSide[side] and not main.coop) or start.challenger > 0 then
						remapInput(i, main.t_remaps[side]) --P1/3/5/7 => P1 controls, P2/4/6/8 => P2 controls
						setCom(i, 0)
					else
						setCom(i, ai or start.f_difficulty(i, offset))
					end
				end
			end
		end
	end
end

--sets lifebar elements, round time, rounds to win
function start.f_setRounds(roundTime, t_rounds)
	setLifebarElements(main.lifebar)
	--round time
	local frames = main.timeFramesPerCount
	local p1FramesMul = 1
	local p2FramesMul = 1
	if start.p[1].teamMode == 3 then --Tag
		p1FramesMul = start.p[1].numChars
	end
	if start.p[2].teamMode == 3 then --Tag
		p2FramesMul = start.p[2].numChars
	end
	frames = frames * math.max(p1FramesMul, p2FramesMul)
	setTimeFramesPerCount(frames)
	if roundTime ~= nil then
		setRoundTime(math.max(-1, roundTime * frames)) --round time predefined
	elseif main.charparam.time and start.f_getCharData(start.p[2].t_selected[1].ref).time ~= nil then --round time assigned as character param
		setRoundTime(math.max(-1, start.f_getCharData(start.p[2].t_selected[1].ref).time * frames))
	else --default round time
		setRoundTime(math.max(-1, main.roundTime * frames))
	end
	--rounds to win
	for side = 1, 2 do
		if t_rounds[side] ~= nil then
			setMatchWins(side, t_rounds[side])
			setMatchMaxDrawGames(side, t_rounds[side])
		else
			if side == 2 and main.charparam.rounds and start.f_getCharData(start.p[2].t_selected[1].ref).rounds ~= nil then --round num assigned as character param
				setMatchWins(side, start.f_getCharData(start.p[2].t_selected[1].ref).rounds)
			elseif start.p[side].teamMode == 1 then --default rounds num (Simul)
				setMatchWins(side, main.matchWins.simul[side])
			elseif start.p[side].teamMode == 3 then --default rounds num (Tag)
				setMatchWins(side, main.matchWins.tag[side])
			else --default rounds num (Single)
				setMatchWins(side, main.matchWins.single[side])
			end
			setMatchMaxDrawGames(side, main.matchWins.draw[side])
		end
	end
	--timer / score counter
	local timer = start.t_savedData.time.total
	local t_score = {start.t_savedData.score.total[1], start.t_savedData.score.total[2]}
	if start.challenger > 0 and gamemode('teamversus') then
		timer = 0
		t_score = {0, 0}
	end
	setLifebarTimer(timer)
	setLifebarScore(t_score[1], t_score[2])
end

--save data between matches
start.t_savedData = {}
function start.f_saveData()
	if main.debugLog then main.f_printTable(t_gameStats, 'debug/t_gameStats.txt') end
	if winnerteam() == -1 then
		return
	end
	--win/lose matches count, total score
	if winnerteam() == 1 then
		start.t_savedData.win[1] = start.t_savedData.win[1] + 1
		start.t_savedData.lose[2] = start.t_savedData.lose[2] + 1
		start.t_savedData.score.total[1] = t_gameStats.p1score
	else --if winnerteam() == 2 then
		start.t_savedData.win[2] = start.t_savedData.win[2] + 1
		start.t_savedData.lose[1] = start.t_savedData.lose[1] + 1
		if main.resetScore and matchno() ~= -1 then --loosing sets score for the next match to lose count
			start.t_savedData.score.total[1] = start.t_savedData.lose[1]
			start.t_savedData.debugflag[1] = false
		else
			start.t_savedData.score.total[1] = t_gameStats.p1score
		end
	end
	start.t_savedData.score.total[2] = t_gameStats.p2score
	--total time
	start.t_savedData.time.total = start.t_savedData.time.total + t_gameStats.matchTime
	--time in each round
	table.insert(start.t_savedData.time.matches, t_gameStats.timerRounds)
	--score in each round
	table.insert(start.t_savedData.score.matches, t_gameStats.scoreRounds)
	--max consecutive wins
	for side = 1, 2 do
		if getConsecutiveWins(side) > start.t_savedData.consecutive[side] then
			start.t_savedData.consecutive[side] = getConsecutiveWins(side)
		end
	end
	if main.debugLog then main.f_printTable(start.t_savedData, 'debug/t_savedData.txt') end
end

--return sorted and capped table
local function f_formattedTable(t, append, f, size)
	local t = t or {}
	local size = size or -1
	local t_tmp = {}
	local tmp = 0
	table.insert(t, append)
	for _, v in main.f_sortKeys(t, f) do
		tmp = tmp + 1
		table.insert(t_tmp, v)
		if tmp == size then
			break
		end
	end
	return t_tmp
end

local function f_listCharRefs(t)
	local ret = {}
	for i = 1, #t do
		table.insert(ret, start.f_getCharData(t[i].ref).char:lower())
	end
	return ret
end

--;===========================================================
--; Ranking
-- start.t_sortRanking is a table storing functions with ranking sorting logic
-- used by start.f_storeStats function, depending on game mode. Can be appended
-- via external module, without conflicting with default scripts.
start.t_sortRanking = {}
start.t_sortRanking.arcade = function(t, a, b) return t[b].score < t[a].score end
start.t_sortRanking.teamcoop = start.t_sortRanking.arcade
start.t_sortRanking.netplayteamcoop = start.t_sortRanking.arcade
start.t_sortRanking.timeattack = function(t, a, b) return t[b].time > t[a].time end
start.t_sortRanking.survival = function(t, a, b) return t[b].win < t[a].win or (t[b].win == t[a].win and t[b].score < t[a].score) end
start.t_sortRanking.survivalcoop = start.t_sortRanking.survival
start.t_sortRanking.netplaysurvivalcoop = start.t_sortRanking.survival

-- as above but the functions return if game mode should be considered "cleared"
start.t_clearCondition = {
	arcade = function() return winnerteam() == 1 end,
	netplaysurvivalcoop = function() return winnerteam() == 1 or start.winCnt >= main.resultsTable.roundstowin end,
	netplayteamcoop = function() return winnerteam() == 1 end,
	survival = function() return winnerteam() == 1 or start.winCnt >= main.resultsTable.roundstowin end,
	survivalcoop = function() return winnerteam() == 1 or start.winCnt >= main.resultsTable.roundstowin end,
	timeattack = function() return winnerteam() == 1 end,
	teamcoop = function() return winnerteam() == 1 end,
}

--data saving to stats.json
local function f_saveStats()
	if main.debugLog then main.f_printTable(stats, 'debug/t_stats.txt') end
	main.f_fileWrite(main.flags['-stats'], json.encode(stats, {indent = 2}))
end

--stats data
function start.f_storeStats()
	local cleared = false
	if start.t_clearCondition[gamemode()] ~= nil then
		cleared = start.t_clearCondition[gamemode()]()
	end
	if stats.modes == nil then
		stats.modes = {}
	end
	--play time
	stats.playtime = main.f_round((stats.playtime or 0) + start.t_savedData.time.total / 60, 2)
	--mode play time
	if stats.modes[gamemode()] == nil then
		stats.modes[gamemode()] = {}
	end
	local t = stats.modes[gamemode()]
	t.playtime = main.f_round((t.playtime or 0) + start.t_savedData.time.total / 60, 2)
	if start.t_sortRanking[gamemode()] == nil or main.t_hiscoreData[gamemode()] == nil then
		return cleared, -1 --mode can't be cleared
	end
	--number times cleared
	if cleared then
		t.clear = (t.clear or 0) + 1
	elseif t.clear == nil then
		t.clear = 0
	end
	--team leader mode cleared count
	if t.clearcount == nil then
		t.clearcount = {}
	end
	if cleared then
		local leader = start.f_getCharData(start.p[1].t_selected[1].ref).char:lower()
		t.clearcount[leader] = (t.clearcount[leader] or 0) + 1
	end
	--ranking data exceptions
	if main.t_hiscoreData[gamemode()].data == 'score' and start.t_savedData.score.total[1] == 0 then
		return cleared, -1
	end
	if main.t_hiscoreData[gamemode()].data == 'win' and start.t_savedData.win[1] == 0 then
		return cleared, -1
	end
	if not cleared and main.rankingCondition then
		return cleared, -1 --only winning produces ranking data
	end
	if start.t_savedData.debugflag[1] then
		return cleared, -1 --using debug keys disables high score table registering
	end
	--rankings
	t.ranking = f_formattedTable(
		t.ranking,
		{
			score = start.t_savedData.score.total[1],
			time = main.f_round(start.t_savedData.time.total / 60, 2),
			name = start.t_savedData.name or '',
			chars = f_listCharRefs(start.p[1].t_selected),
			tmode = start.p[1].teamMode,
			ailevel = config.Difficulty,
			win = start.t_savedData.win[1],
			lose = start.t_savedData.lose[1],
			consecutive = start.t_savedData.consecutive[1],
			flag = true,
		},
		start.t_sortRanking[gamemode()],
		motif.hiscore_info.window_visibleitems
	)
	local place = 0
	for k, v in ipairs(t.ranking) do
		if v.flag then
			place = k
			v.flag = nil
			break
		end
	end
	return cleared, place
end
--;===========================================================

--sets stage
function start.f_setStage(num, assigned)
	if main.stageMenu then
		num = main.t_selectableStages[stageListNo]
		if stageListNo == 0 then
			num = main.t_selectableStages[math.random(1, #main.t_selectableStages)]
			stageListNo = num -- comment out to randomize stage after each fight in survival mode, when random stage is chosen
			stageRandom = true
		else
			num = main.t_selectableStages[stageListNo]
		end
		assigned = true
	end
	if not assigned then
		if main.charparam.stage and start.f_getCharData(start.p[2].t_selected[1].ref).stage ~= nil then --stage assigned as character param
			num = math.random(1, #start.f_getCharData(start.p[2].t_selected[1].ref).stage)
			num = start.f_getCharData(start.p[2].t_selected[1].ref).stage[num]
		elseif main.stageOrder and main.t_orderStages[start.f_getCharData(start.p[2].t_selected[1].ref).order] ~= nil then --stage assigned as stage order param
			num = math.random(1, #main.t_orderStages[start.f_getCharData(start.p[2].t_selected[1].ref).order])
			num = main.t_orderStages[start.f_getCharData(start.p[2].t_selected[1].ref).order][num]
		else --stage randomly selected
			num = main.t_includeStage[1][math.random(1, #main.t_includeStage[1])]
		end
	end
	selectStage(num)
	return num
end

--sets music
function start.f_setMusic(num, data)
	start.bgmround = 0
	start.t_music = {}
	local side = 2
	for _, v in ipairs({'music', 'musicfinal', 'musiclife', 'musicvictory', 'musicvictory'}) do
		if start.t_music[v] == nil then
			start.t_music[v] = {}
		end
		local t_ref = nil
		-- music assigned by launchFight
		if data ~= nil and data[v] ~= nil then
			t_ref = data[v]
		-- game modes other than demo (or demo with stage BGM param enabled)
		elseif not gamemode('demo') or motif.demo_mode.fight_playbgm == 1 then
			-- music assigned as character param
			if (main.charparam.music or (v == 'musicvictory' and main.victoryScreen)) and start.f_getCharData(start.p[side].t_selected[1].ref)[v] ~= nil then
				t_ref = start.f_getCharData(start.p[side].t_selected[1].ref)[v]
			-- music assigned as stage param
			elseif main.t_selStages[num] ~= nil and main.t_selStages[num][v] ~= nil then
				t_ref = main.t_selStages[num][v]
			end
		end
		-- append t_music table
		if t_ref ~= nil then
			-- musicX tracks are nested using round numbers as table keys
			if v == 'music' then
				for k2, v2 in pairs(t_ref) do
					local track = math.random(1, #v2)
					start.t_music[v][k2] = {
						bgmusic = v2[track].bgmusic,
						bgmvolume = v2[track].bgmvolume,
						bgmloopstart = v2[track].bgmloopstart,
						bgmloopend = v2[track].bgmloopend
					}
				end
			else
				local track = math.random(1, #t_ref)
				-- musicvictory tracks are nested using team side as table keys
				if v == 'musicvictory' then
					start.t_music[v][side] = {
						bgmusic = t_ref[track].bgmusic,
						bgmvolume = t_ref[track].bgmvolume,
						bgmloopstart = t_ref[track].bgmloopstart,
						bgmloopend = t_ref[track].bgmloopend
					}
				-- musicfinal and musiclife tracks are stored without additional nesting
				else
					start.t_music[v] = {
						bgmusic = t_ref[track].bgmusic,
						bgmvolume = t_ref[track].bgmvolume,
						bgmloopstart = t_ref[track].bgmloopstart,
						bgmloopend = t_ref[track].bgmloopend
					}
				end
			end
		end
		if v == 'musicvictory' then
			side = 1
		end
	end
	-- bgmratio.life, bgmtrigger.life
	for k, v in pairs({bgmratio_life = 30, bgmtrigger_life = 1}) do
		if main.t_selStages[num] ~= nil and main.t_selStages[num][k] ~= nil then
			start.t_music[k] = main.t_selStages[num][k]
		else
			start.t_music[k] = v
		end
	end
end

--remaps palette based on button press and character's keymap settings
function start.f_reampPal(ref, num)
	return start.f_getCharData(ref).pal_keymap[num] or num
end

-- returns palette number
function start.f_selectPal(ref, palno)
	-- generate table with palette entries already used by this char ref
	local t_assignedPals = {}
	for side = 1, 2 do
		for k, v in pairs(start.p[side].t_selected) do
			if v.ref == ref then
				t_assignedPals[start.p[side].t_selected[k].pal] = true
			end
		end
	end
	-- selected palette
	if palno ~= nil and palno > 0 then
		if not t_assignedPals[start.f_reampPal(ref, palno)] then
			return start.f_reampPal(ref, palno)
		else
			for _, v in ipairs(start.f_getCharData(ref).pal) do
				if not t_assignedPals[start.f_reampPal(ref, v)] then
					return start.f_reampPal(ref, v)
				end
			end
		end
	-- default palette
	elseif (not main.rotationChars and not config.AIRandomColor) or (main.rotationChars and not config.AISurvivalColor) then
		for _, v in ipairs(start.f_getCharData(ref).pal_defaults) do
			if not t_assignedPals[v] then
				return v
			end
		end
	end
	-- random palette
	t = main.f_tableCopy(start.f_getCharData(ref).pal)
	if #t_assignedPals >= #t then -- not enough palettes for unique selection
		return t[math.random(1, #t)]
	end
	main.f_tableShuffle(t)
	for k, v in ipairs(t) do
		if not t_assignedPals[v] then
			return v
		end
	end
	panicError("\n" .. start.f_getCharData(ref).name .. " palette was not selected\n")
end

--returns ratio level
local t_ratioArray = {
	{2, 1, 1},
	{1, 2, 1},
	{1, 1, 2},
	{2, 2},
	{3, 1},
	{1, 3},
	{4}
}
function start.f_getRatio(player, ratio)
	if player == 1 then
		if not start.p[1].ratio and ratio == nil then
			return nil
		end
		return ratio or t_ratioArray[start.p[1].numRatio][#start.p[1].t_selected + 1]
	end
	if not start.p[2].ratio and ratio == nil then
		return nil
	end
	if ratio ~= nil then
		return ratio
	end
	if not continue() and not main.selectMenu[2] and #start.p[2].t_selected == 0 then
		if start.p[2].numChars == 3 then
			start.p[2].numRatio = math.random(1, 3)
		elseif start.p[2].numChars == 2 then
			start.p[2].numRatio = math.random(4, 6)
		else
			start.p[2].numRatio = 7
		end
	end
	return t_ratioArray[start.p[2].numRatio][#start.p[2].t_selected + 1]
end

--returns player number
function start.f_getPlayerNo(side, num)
	if main.coop and not gamemode('versuscoop') then
		return side + num - 1
	end
	if side == 1 then
		return num * 2 - 1
	end
	return num * 2
end

--Convert number to name and get rid of the ""
function start.f_getName(ref, side)
	if ref == nil or start.f_getCharData(ref).hidden == 2 then
		return ''
	end
	if start.f_getCharData(ref).char == 'randomselect' or start.f_getCharData(ref).hidden == 3 then
		return motif.select_info['p' .. (side or 1) .. '_name_random_text']
	end
	return start.f_getCharData(ref).name
end

--reset temp data values
function start.f_resetTempData(t, subname)
	for k, v in pairs(t) do
		if k:match('_data$') then
			animReset(v)
			animUpdate(v)
		end
	end
	for side = 1, 2 do
		if #start.p[side].t_selTemp == 0 then
			for member, v in ipairs(start.p[side].t_selected) do
				table.insert(start.p[side].t_selTemp, {ref = v.ref})
			end
		end
		for member, v in ipairs(start.p[side].t_selTemp) do
			v.anim = t['p' .. side .. '_member' .. member .. subname .. '_anim'] or t['p' .. side .. subname .. '_anim']
			v.anim_data = start.f_animGet(v.ref, side, member, t, subname, '', true)
			v.face2_data = start.f_animGet(v.ref, side, member, t, '_face2', '', true)
			v.slide_dist = {0, 0}
		end
		start.p[side].screenDelay = 0
	end
end

function start.f_animGet(ref, side, member, t, subname, prefix, loop, default)
	if ref == nil then
		return nil
	end
	for _, v in pairs({
		{t['p' .. side .. '_member' .. member .. subname .. prefix .. '_anim'], -1},
		{t['p' .. side .. subname .. prefix .. '_anim'], -1},
		t['p' .. side .. '_member' .. member .. subname .. prefix .. '_spr'],
		t['p' .. side .. subname .. prefix .. '_spr'],
		default
	}) do
		if v[1] ~= nil and v[1] ~= -1 then
			local a = animGetPreloadedData('char', ref, v[1], v[2], loop)
			if a ~= nil then
				local xscale = start.f_getCharData(ref).portrait_scale / (main.SP_Viewport43[3] / main.SP_Localcoord[1])
				local yscale = xscale
				if v[2] == -1 then
					xscale = xscale * (start.f_getCharData(ref).cns_scale[1] or 1)
					yscale = yscale * (start.f_getCharData(ref).cns_scale[2] or 1)
				end
				animSetScale(
					a,
					t['p' .. side .. subname .. '_scale'][1] * (main.f_tableExists(t['p' .. side .. '_member' .. member .. subname .. '_scale'])[1] or 1) * xscale,
					t['p' .. side .. subname .. '_scale'][2] * (main.f_tableExists(t['p' .. side .. '_member' .. member .. subname .. '_scale'])[2] or 1) * yscale,
					false
				)
				animSetWindow(
					a,
					t['p' .. side .. subname .. '_window'][1],
					t['p' .. side .. subname .. '_window'][2],
					t['p' .. side .. subname .. '_window'][3],
					t['p' .. side .. subname .. '_window'][4]
				)
				animUpdate(a)
				return a
			end
		end
	end
	return nil
end

--calculate portraits slide.dist offset
local function f_slideDistCalc(slide_dist, t_dist, t_speed)
	if t_dist == nil or t_speed == nil then
		return
	end
	for i = 1, 2 do
		if (t_dist[i] or 0) > 0 then
			if slide_dist[i] < (t_dist[i] or 0) then
				slide_dist[i] = math.min(slide_dist[i] + (t_speed[i] or 0), t_dist[i] or 0)
			end
		elseif (t_dist[i] or 0) < 0 then
			if slide_dist[i] > (t_dist[i] or 0) then
				slide_dist[i] = math.max(slide_dist[i] - (t_speed[i] or 0), t_dist[i] or 0)
			end
		end
	end
end

--calculate portraits x pos
local function f_portraitsXCalc(side, t, subname, member)
	local x = t['p' .. side .. subname .. '_pos'][1] + t['p' .. side .. subname .. '_offset'][1] + (main.f_tableExists(t['p' .. side .. '_member' .. member .. subname .. '_offset'])[1] or 0)
	if t['p' .. side .. subname .. '_padding'] == 1 then
		return x + (2 * member - 1) * t['p' .. side .. subname .. '_spacing'][1] * t['p' .. side .. subname .. '_num'] / (2 * math.min(t['p' .. side .. subname .. '_num'], math.max(start.p[side].numChars, #start.p[side].t_selected)))
	end
	return x + (member - 1) * t['p' .. side .. subname .. '_spacing'][1]
end

--draw portraits
function start.f_drawPortraits(t_portraits, side, t, subname, last, icon)
	if #t_portraits == 0 then
		return
	end
	-- draw background portrait
	local member = 1
	if last then
		member = #t_portraits
	end
	if t_portraits[member].face2_data ~= nil then
		main.f_animPosDraw(
			t_portraits[member].face2_data,
			t['p' .. side .. subname .. '_pos'][1] + t['p' .. side .. '_face2_offset'][1],
			t['p' .. side .. subname .. '_pos'][2] + t['p' .. side .. '_face2_offset'][2],
			t['p' .. side .. '_face2_facing'],
			true
		)
	end
	-- if next player portrait should replace previous one
	if t['p' .. side .. subname .. '_num'] == 1 and last and not main.coop then
		if t_portraits[#t_portraits].anim_data ~= nil then
			local v = t_portraits[#t_portraits]
			f_slideDistCalc(v.slide_dist, t['p' .. side .. '_member1' .. subname .. '_slide_dist'], t['p' .. side .. '_member1' .. subname .. '_slide_speed'])
			main.f_animPosDraw(
				v.anim_data,
				f_portraitsXCalc(side, t, subname, 1) + main.f_round(v.slide_dist[1]),
				t['p' .. side .. subname .. '_pos'][2] + t['p' .. side .. subname .. '_offset'][2] + (main.f_tableExists(t['p' .. side .. '_member1' .. subname .. '_offset'])[2] or 0) + main.f_round(v.slide_dist[2]),
				t['p' .. side .. subname .. '_facing'],
				true
			)
		end
		return
	end
	-- otherwise render portraits in order, up to the 'num' limit
	for member = #t_portraits, 1, -1 do
		if member <= t['p' .. side .. subname .. '_num'] --[[or (last and main.coop)]] then
			if t_portraits[member].anim_data ~= nil then
				local v = t_portraits[member]
				f_slideDistCalc(v.slide_dist, t['p' .. side .. '_member' .. member .. subname .. '_slide_dist'], t['p' .. side .. '_member' .. member .. subname .. '_slide_speed'])
					main.f_animPosDraw(
					v.anim_data,
					f_portraitsXCalc(side, t, subname, member) + main.f_round(v.slide_dist[1]),
					t['p' .. side .. subname .. '_pos'][2] + t['p' .. side .. subname .. '_offset'][2] + (main.f_tableExists(t['p' .. side .. '_member' .. member .. subname .. '_offset'])[2] or 0) + (member - 1) * t['p' .. side .. subname .. '_spacing'][2] + main.f_round(v.slide_dist[2]),
					t['p' .. side .. subname .. '_facing'],
					true
				)
			end
		end
	end
	-- draw order icons
	if icon == nil then
		return
	end
	for member = 1, #t_portraits do
		if t['p' .. side .. '_member' .. member .. subname .. icon .. '_data'] ~= nil then
			main.f_animPosDraw(
				t['p' .. side .. '_member' .. member .. subname .. icon .. '_data'],
				f_portraitsXCalc(side, t, subname, member),
				t['p' .. side .. subname .. '_pos'][2] + t['p' .. side .. subname .. '_offset'][2] + (main.f_tableExists(t['p' .. side .. '_member' .. member .. subname .. '_offset'])[2] or 0) + (member - 1) * t['p' .. side .. subname .. '_spacing'][2]
			)
		end
	end
end

--returns cell_<col>_<row>_offset values
function start.f_faceOffset(col, row, key)
	if motif.select_info['cell_' .. col .. '_' .. row .. '_offset'] ~= nil then
		return motif.select_info['cell_' .. col .. '_' .. row .. '_offset'][key] or 0
	end
	return 0
end

--returns correct cell position after moving the cursor
function start.f_cellMovement(selX, selY, cmd, side, snd, dir)
	local tmpX = selX
	local tmpY = selY
	local found = false
	if main.f_input({cmd}, {'$U'}) or dir == 'U' then
		for i = 1, motif.select_info.rows do
			selY = selY - 1
			if selY < 0 then
				if motif.select_info.wrapping == 1 or dir ~= nil then
					selY = motif.select_info.rows - 1
				else
					selY = tmpY
				end
			end
			if dir ~= nil then
				found, selX = start.f_searchEmptyBoxes(selX, selY, side, -1)
			elseif (start.t_grid[selY + 1][selX + 1].char ~= nil or motif.select_info.moveoveremptyboxes == 1) and start.t_grid[selY + 1][selX + 1].skip ~= 1 and (config.TeamDuplicates or start.t_grid[selY + 1][selX + 1].char == 'randomselect' or not t_reservedChars[side][start.t_grid[selY + 1][selX + 1].char_ref]) and start.t_grid[selY + 1][selX + 1].hidden ~= 2 then
				break
			elseif motif.select_info.searchemptyboxesup ~= 0 then
				found, selX = start.f_searchEmptyBoxes(selX, selY, side, motif.select_info.searchemptyboxesup)
			end
			if found then
				break
			end
		end
	elseif main.f_input({cmd}, {'$D'}) or dir == 'D' then
		for i = 1, motif.select_info.rows do
			selY = selY + 1
			if selY >= motif.select_info.rows then
				if motif.select_info.wrapping == 1 or dir ~= nil then
					selY = 0
				else
					selY = tmpY
				end
			end
			if dir ~= nil then
				found, selX = start.f_searchEmptyBoxes(selX, selY, side, 1)
			elseif (start.t_grid[selY + 1][selX + 1].char ~= nil or motif.select_info.moveoveremptyboxes == 1) and start.t_grid[selY + 1][selX + 1].skip ~= 1 and (config.TeamDuplicates or start.t_grid[selY + 1][selX + 1].char == 'randomselect' or not t_reservedChars[side][start.t_grid[selY + 1][selX + 1].char_ref]) and start.t_grid[selY + 1][selX + 1].hidden ~= 2 then
				break
			elseif motif.select_info.searchemptyboxesdown ~= 0 then
				found, selX = start.f_searchEmptyBoxes(selX, selY, side, motif.select_info.searchemptyboxesdown)
			end
			if found then
				break
			end
		end
	elseif main.f_input({cmd}, {'$B'}) or dir == 'B' then
		if dir ~= nil then
			found, selX = start.f_searchEmptyBoxes(selX, selY, side, -1)
		else
			for i = 1, motif.select_info.columns do
				selX = selX - 1
				if selX < 0 then
					if motif.select_info.wrapping == 1 then
						selX = motif.select_info.columns - 1
					else
						selX = tmpX
					end
				end
				if (start.t_grid[selY + 1][selX + 1].char ~= nil or motif.select_info.moveoveremptyboxes == 1) and start.t_grid[selY + 1][selX + 1].skip ~= 1 and (config.TeamDuplicates or start.t_grid[selY + 1][selX + 1].char == 'randomselect' or not t_reservedChars[side][start.t_grid[selY + 1][selX + 1].char_ref]) and start.t_grid[selY + 1][selX + 1].hidden ~= 2 then
					break
				end
			end
		end
	elseif main.f_input({cmd}, {'$F'}) or dir == 'F' then
		if dir ~= nil then
			found, selX = start.f_searchEmptyBoxes(selX, selY, side, 1)
		else
			for i = 1, motif.select_info.columns do
				selX = selX + 1
				if selX >= motif.select_info.columns then
					if motif.select_info.wrapping == 1 then
						selX = 0
					else
						selX = tmpX
					end
				end
				if (start.t_grid[selY + 1][selX + 1].char ~= nil or motif.select_info.moveoveremptyboxes == 1) and start.t_grid[selY + 1][selX + 1].skip ~= 1 and (config.TeamDuplicates or start.t_grid[selY + 1][selX + 1].char == 'randomselect' or not t_reservedChars[side][start.t_grid[selY + 1][selX + 1].char_ref]) and start.t_grid[selY + 1][selX + 1].hidden ~= 2 then
					break
				end
			end
		end
	end
	if (tmpX ~= selX or tmpY ~= selY) then
		if dir == nil then
			sndPlay(motif.files.snd_data, snd[1], snd[2])
		end
	end
	return selX, selY
end

--used by above function to find valid cell in case of dummy character entries
function start.f_searchEmptyBoxes(x, y, side, direction)
	if direction > 0 then --right
		while true do
			x = x + 1
			if x >= motif.select_info.columns then
				return false, 0
			elseif start.t_grid[y + 1][x + 1].skip ~= 1 and start.t_grid[y + 1][x + 1].char ~= nil and (start.t_grid[y + 1][x + 1].char == 'randomselect' or not t_reservedChars[side][start.t_grid[y + 1][x + 1].char_ref]) and start.t_grid[y + 1][x + 1].hidden ~= 2 then
				return true, x
			end
		end
	elseif direction < 0 then --left
		while true do
			x = x - 1
			if x < 0 then
				return false, motif.select_info.columns - 1
			elseif start.t_grid[y + 1][x + 1].skip ~= 1 and start.t_grid[y + 1][x + 1].char ~= nil and (start.t_grid[y + 1][x + 1].char == 'randomselect' or not t_reservedChars[side][start.t_grid[y + 1][x + 1].char_ref]) and start.t_grid[y + 1][x + 1].hidden ~= 2 then
				return true, x
			end
		end
	end
end

--returns player cursor data
function start.f_getCursorData(pn, suffix)
	if main.coop and motif.select_info['p' .. pn .. suffix] ~= nil then
		return motif.select_info['p' .. pn .. suffix]
	end
	return motif.select_info['p' .. (pn - 1) % 2 + 1 .. suffix]
end

--draw cursor
function start.f_drawCursor(pn, x, y, param)
	-- in non-coop modes only p1 and p2 cursors are used
	if not main.coop then
		pn = (pn - 1) % 2 + 1
	end
	local prefix = 'p' .. pn .. param .. '_' .. x + 1 .. '_' .. y + 1
	-- create spr/anim data, if not existing yet
	if motif.select_info[prefix .. '_data'] == nil then
		-- if cell based variants are not defined we're defaulting to standard pX parameters
		for _, v in ipairs({'_anim', '_spr', '_offset', '_scale', '_facing'}) do
			if motif.select_info[prefix .. v] == nil then
				motif.select_info[prefix .. v] = start.f_getCursorData(pn, param .. v)
			end
		end
		motif.f_loadSprData(motif.select_info, {s = prefix .. '_'})
	end
	-- draw
	main.f_animPosDraw(
		motif.select_info[prefix .. '_data'],
		motif.select_info.pos[1] + x * (motif.select_info.cell_size[1] + motif.select_info.cell_spacing[1]) + start.f_faceOffset(x + 1, y + 1, 1),
		motif.select_info.pos[2] + y * (motif.select_info.cell_size[2] + motif.select_info.cell_spacing[2]) + start.f_faceOffset(x + 1, y + 1, 2),
		(motif.select_info['cell_' .. x + 1 .. '_' .. y + 1 .. '_facing'] or motif.select_info['p' .. pn .. param .. '_facing'])
	)
end
--returns t_selChars table out of cell number
function start.f_selGrid(cell, slot)
	if main.t_selGrid[cell] == nil or #main.t_selGrid[cell].chars == 0 then
		local csCol = ((cell - 1) % motif.select_info.columns) + 1
		local csRow = math.floor((cell - 1) / motif.select_info.columns) + 1
		if motif.select_info['cell_' .. csCol .. '_' .. csRow .. '_skip'] == 1 then
			return {skip = 1}
		end
		return {}
	end
	return main.t_selChars[main.t_selGrid[cell].chars[(slot or main.t_selGrid[cell].slot)]]
end

--returns t_selChars table out of char ref
function start.f_getCharData(ref)
	return main.t_selChars[ref + 1]
end

--returns stage ref out of def filename
function start.f_getStageRef(def)
	if def == '' then
		return getStageNo()
	end
	if main.t_stageDef[def:lower()] == nil then
		 main.f_addStage(def)
	end
	return main.t_stageDef[def:lower()]
end

--returns char ref out of def filename
function start.f_getCharRef(def)
	if main.t_charDef[def:lower()] == nil then
		if not main.f_addChar(def .. ', order = 0, ordersurvival = 0, exclude = 1', true, false) then
			panicError("\nUnable to add character. No such file or directory: " .. def .. "\n")
		end
	end
	return main.t_charDef[def:lower()]
end

--returns teammode int from string
function start.f_stringToTeamMode(tm)
	if tm == 'single' then
		return 0
	elseif tm == 'simul' then
		return 1
	elseif tm == 'turns' then
		return 2
	elseif tm == 'tag' then
		return 3
	end
	return nil
end

--returns formatted clear time string
function start.f_clearTimeText(text, totalSec)
	local h = tostring(math.floor(totalSec / 3600))
	local m = tostring(math.floor((totalSec / 3600 - h) * 60))
	local s = tostring(math.floor(((totalSec / 3600 - h) * 60 - m) * 60))
	local x = tostring(math.floor((((totalSec / 3600 - h) * 60 - m) * 60 - s) * 100))
	if string.len(m) < 2 then
		m = '0' .. m
	end
	if string.len(s) < 2 then
		s = '0' .. s
	end
	if string.len(x) < 2 then
		x = '0' .. x
	end
	return text:gsub('%%h', h):gsub('%%m', m):gsub('%%s', s):gsub('%%x', x)
end

--returns formatted record text table
function start.f_getRecordText()
	if motif.select_info['record_' .. gamemode() .. '_text'] == nil or stats.modes == nil or stats.modes[gamemode()] == nil or stats.modes[gamemode()].ranking == nil or stats.modes[gamemode()].ranking[1] == nil then
		return {}
	end
	local text = motif.select_info['record_' .. gamemode() .. '_text']
	--time
	text = start.f_clearTimeText(text, stats.modes[gamemode()].ranking[1].time)
	--score
	text = text:gsub('%%p', tostring(stats.modes[gamemode()].ranking[1].score))
	--char name
	local name = '?' --in case character being removed from roster
	if main.t_charDef[stats.modes[gamemode()].ranking[1].chars[1]] ~= nil then
		name = start.f_getCharData(main.t_charDef[stats.modes[gamemode()].ranking[1].chars[1]]).name
	end
	text = text:gsub('%%c', name)
	--player name
	text = text:gsub('%%n', stats.modes[gamemode()].ranking[1].name)
	return main.f_extractText(text)
end

--cursor sound data, play cursor sound
function start.f_playWave(ref, name, g, n, loops)
	if g < 0 or n < 0 then return 0 end
	if name == 'stage' then
		local a = main.t_selStages[ref].attachedChar
		if a == nil or a.sound == nil then
			return 0
		end
		if main.t_selStages[ref][name .. '_wave_data'] == nil then
			main.t_selStages[ref][name .. '_wave_data'] = getWaveData(a.dir .. a.sound, g, n, loops or -1)
		end
		wavePlay(main.t_selStages[ref][name .. '_wave_data'])
	else
		local sound = start.f_getCharData(ref).sound
		if sound == nil or sound == '' then
			return 0
		end
		if start.f_getCharData(ref)[name .. '_wave_data'] == nil then
			start.f_getCharData(ref)[name .. '_wave_data'] = getWaveData(start.f_getCharData(ref).dir .. sound, g, n, loops or -1)
		end
		wavePlay(start.f_getCharData(ref)[name .. '_wave_data'])
	end
end

--removes char with particular ref from table
function start.f_excludeChar(t, ref)
	for _, sel in ipairs(main.t_selChars) do
		if sel.char_ref == ref then
			if t[sel.order] ~= nil then
				for k, v in ipairs(t[sel.order]) do
					if v == ref then
						table.remove(t[sel.order], k)
					end
				end
			end
			break
		end
	end
	return t
end

--returns random char ref
function start.f_randomChar(pn)
	if #main.t_randomChars == 0 then
		return nil
	end
	if config.TeamDuplicates then
		return main.t_randomChars[math.random(1, #main.t_randomChars)]
	end
	local t = {}
	for k, v in ipairs(main.t_randomChars) do
		if not t_reservedChars[pn][v] then
			table.insert(t, v)
		end
	end
	if #t > 0 then
		return t[math.random(1, #t)]
	end
	return main.t_randomChars[math.random(1, #main.t_randomChars)]
end

--return true if slot is selected, update start.t_grid
function start.f_slotSelected(cell, side, cmd, player, x, y)
	if cmd == nil then
		return false
	end
	if #main.t_selGrid[cell].chars > 0 then
		-- select.def 'slot' parameter special keys detection
		for _, cmdType in ipairs({'select', 'next', 'previous'}) do
			if main.t_selGrid[cell][cmdType] ~= nil then
				for k, v in pairs(main.t_selGrid[cell][cmdType]) do
					if main.f_input({cmd}, main.f_extractKeys(k)) then
						if cmdType == 'next' then
							local ok = false
							for i = main.t_selGrid[cell].slot + 1, #v do
								if start.f_getCharData(start.f_selGrid(cell, v[i]).char_ref).hidden < 2 then
									main.t_selGrid[cell].slot = v[i]
									ok = true
									break
								end
							end
							if not ok then
								for i = 1, main.t_selGrid[cell].slot - 1 do
									if start.f_getCharData(start.f_selGrid(cell, v[i]).char_ref).hidden < 2 then
										main.t_selGrid[cell].slot = v[i]
										ok = true
										break
									end
								end
							end
							if ok then
								sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_swap_snd'][1], motif.select_info['p' .. side .. '_swap_snd'][2])
							end
						elseif cmdType == 'previous' then
							local ok = false
							for i = main.t_selGrid[cell].slot -1, 1, -1 do
								if start.f_getCharData(start.f_selGrid(cell, v[i]).char_ref).hidden < 2 then
									main.t_selGrid[cell].slot = v[i]
									ok = true
									break
								end
							end
							if not ok then
								for i = #v, main.t_selGrid[cell].slot + 1, -1 do
									if start.f_getCharData(start.f_selGrid(cell, v[i]).char_ref).hidden < 2 then
										main.t_selGrid[cell].slot = v[i]
										ok = true
										break
									end
								end
							end
							if ok then
								sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_swap_snd'][1], motif.select_info['p' .. side .. '_swap_snd'][2])
							end
						else --select
							main.t_selGrid[cell].slot = v[math.random(1, #v)]
							start.c[player].selRef = start.f_selGrid(cell).char_ref
						end
						start.t_grid[y + 1][x + 1].char = start.f_selGrid(cell).char
						start.t_grid[y + 1][x + 1].char_ref = start.f_selGrid(cell).char_ref
						start.t_grid[y + 1][x + 1].hidden = start.f_selGrid(cell).hidden
						start.t_grid[y + 1][x + 1].skip = start.f_selGrid(cell).skip
						return cmdType == 'select'
					end
				end
			end
		end
	end
	-- returns true on pressed key if current slot is not blocked by TeamDuplicates feature
	return main.f_btnPalNo(cmd) > 0 and (not t_reservedChars[side][start.t_grid[y + 1][x + 1].char_ref] or start.t_grid[start.c[player].selY + 1][start.c[player].selX + 1].char == 'randomselect')
end

--generate start.t_grid table, assign row and cell to main.t_selChars
local cnt = motif.select_info.columns + 1
local row = 1
local col = 0
start.t_grid = {[row] = {}}
for i = 1, motif.select_info.rows * motif.select_info.columns do
	if i == cnt then
		row = row + 1
		cnt = cnt + motif.select_info.columns
		start.t_grid[row] = {}
	end
	col = #start.t_grid[row] + 1
	start.t_grid[row][col] = {
		x = (col - 1) * (motif.select_info.cell_size[1] + motif.select_info.cell_spacing[1]) + start.f_faceOffset(col, row, 1),
		y = (row - 1) * (motif.select_info.cell_size[2] + motif.select_info.cell_spacing[2]) + start.f_faceOffset(col, row, 2)
	}
	if start.f_selGrid(i).char ~= nil then
		start.t_grid[row][col].char = start.f_selGrid(i).char
		start.t_grid[row][col].char_ref = start.f_selGrid(i).char_ref
		start.t_grid[row][col].hidden = start.f_selGrid(i).hidden
		for j = 1, #main.t_selGrid[i].chars do
			start.f_selGrid(i, j).row = row
			start.f_selGrid(i, j).col = col
		end
	end
	if start.f_selGrid(i).skip == 1 then
		start.t_grid[row][col].skip = 1
	end
end
if main.debugLog then main.f_printTable(start.t_grid, 'debug/t_grid.txt') end

-- return amount of life to recover
local function f_lifeRecovery(lifeMax, ratioLevel)
	local bonus = lifeMax * config.TurnsRecoveryBonus / 100
	local base = lifeMax * config.TurnsRecoveryBase / 100
	if ratioLevel > 0 then
		bonus = lifeMax * config.RatioRecoveryBonus / 100
		base = lifeMax * config.RatioRecoveryBase / 100
	end
	return base + main.f_round(timeremaining() / (timeremaining() + timeelapsed()) * bonus)
end

-- turns mode life recovery or mode with life persistence between matches
function start.f_turnsRecovery()
	if start.turnsRecoveryInit then
		return
	end
	start.turnsRecoveryInit = true
	player(winnerteam())
	for i = 1, teamsize() * 2 do
		if player(i) and win() and alive() then --assign sys.debugWC if player i exists, member of winning team, alive
			if (not matchover() and teammode() == 'turns') or main.lifePersistence then
				setLife(math.min(lifemax(), life() + f_lifeRecovery(lifemax(), ratiolevel())))
			end
		end
	end
end

-- match persistence
function start.f_matchPersistence()
	-- checked only after at least 1 match
	if matchno() >= 2 then
		-- set 'existed' flag (decides if var/fvar should be persistent between matches)
		for _, v in ipairs(t_gameStats.match) do
			for _, t in pairs(v) do
				if start.p[t.teamside + 1].t_selected[t.memberNo + 1] ~= nil then
					start.p[t.teamside + 1].t_selected[t.memberNo + 1].existed = true
				end
			end
		end
		-- if defeated members should be removed from team, or if life should be maintained
		if main.dropDefeated or main.lifePersistence then
			local t_removeMembers = {}
			-- Turns
			if start.p[1].teamMode == 2 then
				--for each round in the last match
				for _, v in ipairs(t_gameStats.match) do
					-- if defeated
					if v[1].ko and v[1].life <= 0 then
						-- remove character from team
						if main.dropDefeated then
							t_removeMembers[v[1].memberNo + 1] = true
						-- or resurrect and recover character's life
						elseif main.lifePersistence then
							start.p[1].t_selected[v[1].memberNo + 1].life = math.max(1, f_lifeRecovery(v[1].lifeMax, v[1].ratiolevel))
						end
					-- otherwise maintain character's life
					elseif main.lifePersistence then
						start.p[1].t_selected[v[1].memberNo + 1].life = v[1].life
					end
				end
			-- Single / Simul / Tag
			else
				-- for each player data in the last round
				for _, v in pairs(t_gameStats.match[#t_gameStats.match]) do
					-- only check player controlled characters
					if not main.cpuSide[v.teamside + 1] then
						-- if defeated
						if v.ko and v.life <= 0 then
							-- remove character from team
							if main.dropDefeated then
								t_removeMembers[v.memberNo + 1] = true
							-- or resurrect and recover character's life
							elseif main.lifePersistence then
								start.p[1].t_selected[v.memberNo + 1].life = math.max(1, f_lifeRecovery(v.lifeMax, v.ratiolevel))
							end
						-- otherwise maintain character's life
						elseif main.lifePersistence then
							start.p[1].t_selected[v.memberNo + 1].life = v.life
						end
					end
				end
			end
			-- drop defeated characters
			for i = #start.p[1].t_selected, 1, -1 do
				if t_removeMembers[i] then
					table.remove(start.p[1].t_selected, i)
					table.remove(start.p[1].t_selTemp, i)
					start.p[1].numChars = start.p[1].numChars - 1
				end
			end
		end
	end
	return start.p[1].numChars
end

--upcoming match character data adjustment
function start.f_overrideCharData()
	for side = 1, 2 do
		for member, v in ipairs(start.p[side].t_selected) do
			overrideCharData(side, member, {
				['life'] = v.life,
				['lifeMax'] = v.lifeMax,
				['power'] = v.power,
				['dizzyPoints'] = v.dizzyPoints,
				['guardPoints'] = v.guardPoints,
				['ratioLevel'] = v.ratioLevel,
				['lifeRatio'] = v.lifeRatio or config.RatioLife[v.ratioLevel],
				['attackRatio'] = v.attackRatio or config.RatioAttack[v.ratioLevel],
				['existed'] = v.existed,
			})
		end
	end
end

--start game
function start.f_game(lua)
	clearColor(0, 0, 0)
	if main.debugLog and start ~= nil then main.f_printTable(start.p, 'debug/t_p.txt') end
	local p2In = main.t_pIn[2]
	main.t_pIn[2] = 2
	if lua ~= '' then commonLuaInsert(lua) end
	local winner, tbl = game()
	main.f_restoreInput()
	if lua ~= '' then commonLuaDelete(lua) end
	if gameend() then
		clearColor(0, 0, 0)
		os.exit()
	end
	main.t_pIn[2] = p2In
	return winner, tbl
end

--;===========================================================
--; MODES LOOP
--;===========================================================
function start.f_selectMode()
	start.f_selectReset(true)
	while true do
		--select screen
		if not start.f_selectScreen() then
			sndPlay(motif.files.snd_data, motif.select_info.cancel_snd[1], motif.select_info.cancel_snd[2])
			main.f_bgReset(motif[main.background].bg)
			main.f_fadeReset('fadein', motif[main.group])
			main.f_playBGM(false, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
			return
		end
		--first match
		if start.reset then
			main.t_availableChars = main.f_tableCopy(main.t_orderChars)
			--generate default roster
			if main.makeRoster then
				start.t_roster = start.f_makeRoster()
			end
			--generate AI ramping table
			if main.aiRamp then
				start.f_aiRamp(1)
			end
			start.reset = false
		end
		--lua file with custom arcade path detection
		local path = main.luaPath
		if main.charparam.arcadepath then
			if start.p[2].ratio and start.f_getCharData(start.p[1].t_selected[1].ref).ratiopath ~= '' then
				path = start.f_getCharData(start.p[1].t_selected[1].ref).ratiopath
				if not main.f_fileExists(path) then
					panicError("\n" .. start.f_getCharData(start.p[1].t_selected[1].ref).name .. " ratiopath doesn't exist: " .. path .. "\n")
				end
			elseif not start.p[2].ratio and start.f_getCharData(start.p[1].t_selected[1].ref).arcadepath ~= '' then
				path = start.f_getCharData(start.p[1].t_selected[1].ref).arcadepath
				if not main.f_fileExists(path) then
					panicError("\n" .. start.f_getCharData(start.p[1].t_selected[1].ref).name .. " arcadepath doesn't exist: " .. path .. "\n")
				end
			end
		end
		--external script execution
		assert(loadfile(path))()
		--infinite matches flag detected
		if main.makeRoster and start.t_roster[matchno()] ~= nil and start.t_roster[matchno()][1] == -1 then
			table.remove(start.t_roster, matchno())
			start.t_roster = start.f_makeRoster(start.t_roster)
			if main.aiRamp then
				start.f_aiRamp(matchno())
			end
		--otherwise
		else
			if matchno() == -1 then --no more matches left
				--hiscore and stats data
				local cleared, place = start.f_storeStats()
				if main.hiscoreScreen and main.t_hiscoreData[gamemode()] ~= nil and motif.hiscore_info.enabled == 1 and place > 0 then
					start.hiscoreInit = false
					while start.f_hiscore(main.t_hiscoreData[gamemode()], true, place) do
						main.f_refresh()
					end
				end
				f_saveStats()
				--credits
				if cleared and main.storyboard.credits and motif.end_credits.enabled == 1 and main.f_fileExists(motif.end_credits.storyboard) then
					storyboard.f_storyboard(motif.end_credits.storyboard)
				end
				--game over
				if main.storyboard.gameover and motif.game_over_screen.enabled == 1 and main.f_fileExists(motif.game_over_screen.storyboard) then
					if not main.continueScreen or (not continue() and motif.continue_screen.gameover_enabled == 1) then
						storyboard.f_storyboard(motif.game_over_screen.storyboard)
					end
				end
				--exit to main menu
				if main.exitSelect then
					if motif.files.intro_storyboard ~= '' and motif.attract_mode.enabled == 0 then
						storyboard.f_storyboard(motif.files.intro_storyboard)
					end
				end
				start.exit = start.exit or main.exitSelect or not main.selectMenu[1]
			end
			if start.exit then
				main.f_bgReset(motif[main.background].bg)
				main.f_fadeReset('fadein', motif[main.group])
				main.f_playBGM(false, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
				start.exit = false
				return
			end
			if not continue() or esc() then
				start.f_selectReset(false)
			else
				t_reservedChars = {{}, {}}
			end
		end
	end
end

--resets various data
function start.f_selectReset(hardReset)
	esc(false)
	setMatchNo(1)
	setConsecutiveWins(1, 0)
	setConsecutiveWins(2, 0)
	setContinue(false)
	main.f_cmdInput()
	local col = 1
	local row = 1
	for i = 1, #main.t_selGrid do
		if i > motif.select_info.columns * row then
			row = row + 1
			col = 1
		end
		if main.t_selGrid[i].slot ~= 1 then
			main.t_selGrid[i].slot = 1
			start.t_grid[row][col].char = start.f_selGrid(i).char
			start.t_grid[row][col].char_ref = start.f_selGrid(i).char_ref
			start.t_grid[row][col].hidden = start.f_selGrid(i).hidden
			start.t_grid[row][col].skip = start.f_selGrid(i).skip
		end
		col = col + 1
	end
	if hardReset then
		stageListNo = 0
		restoreCursor = false
		--cursor start cell
		for i = 1, config.Players do
			if start.f_getCursorData(i, '_cursor_startcell')[1] < motif.select_info.rows then
				start.c[i].selY = start.f_getCursorData(i, '_cursor_startcell')[1]
			else
				start.c[i].selY = 0
			end
			if start.f_getCursorData(i, '_cursor_startcell')[2] < motif.select_info.columns then
				start.c[i].selX = start.f_getCursorData(i, '_cursor_startcell')[2]
			else
				start.c[i].selX = 0
			end
			start.c[i].cell = -1
			start.c[i].randCnt = 0
			start.c[i].randRef = nil
		end
	end
	if stageRandom then
		stageListNo = 0
		stageRandom = false
	end
	for side = 1, 2 do
		if hardReset then
			start.p[side].numSimul = math.max(2, config.NumSimul[1])
			start.p[side].numTag = math.max(2, config.NumTag[1])
			start.p[side].numTurns = math.max(2, config.NumTurns[1])
			start.p[side].numRatio = 1
			start.p[side].teamMenu = 1
			start.p[side].t_cursor = {}
			start.p[side].teamMode = 0
		end
		start.p[side].numSimul = math.min(start.p[side].numSimul, main.numSimul[2])
		start.p[side].numTag = math.min(start.p[side].numTag, main.numTag[2])
		start.p[side].numTurns = math.min(start.p[side].numTurns, main.numTurns[2])
		start.p[side].numChars = 1
		start.p[side].teamEnd = main.cpuSide[side] and (side == 2 or not main.cpuSide[1]) and main.forceChar[side] == nil
		start.p[side].selEnd = not main.selectMenu[side]
		start.p[side].ratio = false
		start.p[side].t_selected = {}
		start.p[side].t_selTemp = {}
		start.p[side].t_selCmd = {}
	end
	for _, v in ipairs(start.c) do
		v.cell = -1
	end
	selScreenEnd = false
	stageEnd = false
	t_reservedChars = {{}, {}}
	start.winCnt = 0
	start.loseCnt = 0
	if start.challenger == 0 then
		start.t_savedData = {
			win = {0, 0},
			lose = {0, 0},
			time = {total = 0, matches = {}},
			score = {total = {0, 0}, matches = {}},
			consecutive = {0, 0},
			debugflag = {false, false},
		}
		start.t_roster = {}
		start.reset = true
	end
	t_recordText = start.f_getRecordText()
	menu.movelistChar = 1
end

function start.f_selectChallenger()
	esc(false)
	--save values
	local t_p_sav = main.f_tableCopy(start.p)
	local t_c_sav = main.f_tableCopy(start.c)
	local winCnt_sav = start.winCnt
	local loseCnt_sav = start.loseCnt
	local matchNo_sav = matchno()
	local p1cmd = main.t_remaps[1]
	local p2cmd = main.t_remaps[start.challenger]
	local p1ConsecutiveWins = getConsecutiveWins(1)
	local p2ConsecutiveWins = getConsecutiveWins(2)
	--start challenger match
	main.f_default()
	main.f_playerInput(p1cmd, 1)
	remapInput(2, p2cmd)
	main.t_itemname.versus()
	start.f_selectReset(false)
	if not start.f_selectScreen() then
		start.exit = true
		return false
	end
	local ok = launchFight{challenger = true}
	--restore values
	main.f_default()
	main.playerInput = p1cmd -- main.f_playerInput called via main.t_itemname.arcade()
	main.t_itemname.arcade()
	if not ok then
		return false
	end
	start.p = t_p_sav
	start.c = t_c_sav
	start.winCnt = winCnt_sav
	start.loseCnt = loseCnt_sav
	setMatchNo(matchNo_sav)
	setConsecutiveWins(1, p1ConsecutiveWins)
	setConsecutiveWins(2, p2ConsecutiveWins)
	return true
end

function launchFight(data)
	local t = {}
	if continue() then -- on rematch all arguments are ignored and values are restored from last match
		t = main.f_tableCopy(start.launchFightSav)
		start.p[2].t_selTemp = {} -- in case it's not cleaned already (preserved p2 side during select screen)
	else -- otherwise take all arguments and settings into account
		t.p1numchars = start.p[1].numChars
		t.p1teammode = start.p[1].teamMode
		t.p2numchars = start.p[2].numChars
		t.p2teammode = start.p[2].teamMode
		t.challenger = main.f_arg(data.challenger, false)
		t.continue = main.f_arg(data.continue, main.continueScreen)
		t.quickcontinue = (not main.selectMenu[1] and not main.selectMenu[2]) or main.f_arg(data.quickcontinue, main.quickContinue or config.QuickContinue)
		t.order = data.order or 1
		t.orderselect = {main.f_arg(data.p1orderselect, main.orderSelect[1]), main.f_arg(data.p2orderselect, main.orderSelect[2])}
		t.p1char = data.p1char or {}
		t.p1numratio = data.p1numratio or {}
		t.p1rounds = data.p1rounds or nil
		t.p2char = data.p2char or {}
		t.p2numratio = data.p2numratio or {}
		t.p2rounds = data.p2rounds or nil
		t.exclude = data.exclude or {}
		t.musicData = {}
		-- Parse musicX / musicfinal / musiclife / musicvictory arguments
		for k, v in pairs(data) do
			if k:match('^music') then
				-- old syntax with only string argument maintained for backward compatibility with previous builds
				if type(v) == "string" then
					v = {v}
				end
				local bgtype, round = k:match('^(music[a-z]*)([0-9]*)$')
				if t.musicData[bgtype] == nil then
					t.musicData[bgtype] = {}
				end
				local t_ref = t.musicData[bgtype]
				-- musicX parameters are nested using round numbers as table keys
				if bgtype == 'music' or round ~= '' then
					round = tonumber(round) or 1
					if t.musicData[bgtype][round] == nil then t.musicData[bgtype][round] = {} end
					t_ref = t.musicData[bgtype][round]
				end
				table.insert(t_ref, {bgmusic = (v[1] or ''), bgmvolume = (v[2] or 100), bgmloopstart = (v[3] or 0), bgmloopend = (v[4] or 0)})
			end
		end
		t.stage = data.stage or ''
		t.ai = data.ai or nil
		t.vsscreen = main.f_arg(data.vsscreen, main.versusScreen)
		t.victoryscreen = main.f_arg(data.victoryscreen, main.victoryScreen)
		--t.frames = data.frames or framespercount()
		t.roundtime = data.time or nil
		t.lua = data.lua or ''
		t.stageNo = start.f_getStageRef(t.stage)
		start.p[1].numChars = data.p1numchars or math.max(start.p[1].numChars, #t.p1char)
		start.p[1].teamMode = start.f_stringToTeamMode(data.p1teammode) or start.p[1].teamMode
		start.p[2].numChars = data.p2numchars or math.max(start.p[2].numChars, #t.p2char)
		start.p[2].teamMode = start.f_stringToTeamMode(data.p2teammode) or start.p[2].teamMode
		t.p1numchars = start.f_matchPersistence()
		-- add P1 chars forced via function arguments (ignore char param restrictions)
		local reset = false
		local cnt = 0
		for _, v in main.f_sortKeys(t.p1char) do
			if not reset then
				start.p[1].t_selected = {}
				start.p[1].t_selTemp = {}
				reset = true
			end
			cnt = cnt + 1
			local ref = start.f_getCharRef(v)
			table.insert(start.p[1].t_selected, {
				ref = ref,
				pal = start.f_selectPal(ref),
				pn = start.f_getPlayerNo(1, #start.p[1].t_selected + 1),
				--cursor = {},
				ratioLevel = start.f_getRatio(1, t.p1numratio[cnt]),
			})
			main.t_availableChars = start.f_excludeChar(main.t_availableChars, ref)
		end
		if #start.p[1].t_selected == 0 then
			panicError("\n" .. "launchFight(): no valid P1 characters\n")
			start.exit = true
			return false -- return to main menu
		end
		-- add P2 chars forced via function arguments (ignore char param restrictions)
		local onlyme = false
		cnt = 0
		for _, v in main.f_sortKeys(t.p2char) do
			cnt = cnt + 1
			local ref = start.f_getCharRef(v)
			table.insert(start.p[2].t_selected, {
				ref = ref,
				pal = start.f_selectPal(ref),
				pn = start.f_getPlayerNo(2, #start.p[2].t_selected + 1),
				--cursor = {},
				ratioLevel = start.f_getRatio(2, t.p2numratio[cnt]),
			})
			main.t_availableChars = start.f_excludeChar(main.t_availableChars, ref)
			if not onlyme then onlyme = start.f_getCharData(ref).single end
		end
		-- add remaining P2 chars of particular order if there are still free slots in the selected team mode
		if main.cpuSide[2] and #start.p[2].t_selected < start.p[2].numChars and not onlyme then
			-- get list of available chars
			local t_chars = main.f_tableCopy(main.t_availableChars)
			-- remove chars temporary excluded from this match
			for _, v in ipairs(t.exclude) do
				t_chars = start.f_excludeChar(t_chars, start.f_getCharRef(v))
			end
			-- remove chars with 'single' param if some characters are forced into team
			if #start.p[2].t_selected > 0 then
				for _, v in ipairs(t_chars[t.order]) do
					if start.f_getCharData(v).single then
						t_chars = start.f_excludeChar(t_chars, v)
					end
				end
			end
			-- fill free slots
			local t_remaining = main.f_tableCopy(t_chars)
			local t_tmp = {}
			for i = #start.p[2].t_selected, start.p[2].numChars - 1 do
				if t_chars[t.order] ~= nil and #t_chars[t.order] > 0 then
					local rand = math.random(1, #t_chars[t.order])
					local ref = t_chars[t.order][rand]
					if not start.f_getCharData(ref).single then
						table.remove(t_chars[t.order], rand)
						table.insert(t_tmp, ref)
					else --one entry if 'single' param is detected on any opponent
						t_tmp = {ref}
						onlyme = true
						break
					end
				end
			end
			-- not enough unique characters of particular order, take into account only if skiporder parameter = false
			while not t.skiporder and #t_tmp + #start.p[2].t_selected < start.p[2].numChars and not onlyme and t_remaining[t.order] ~= nil and #t_remaining[t.order] > 0 do
				table.insert(t_tmp, t_remaining[t.order][math.random(1, #t_remaining[t.order])])
			end
			-- append remaining characters
			for _, v in ipairs(t_tmp) do
				table.insert(start.p[2].t_selected, {
					ref = v,
					pal = start.f_selectPal(v),
					pn = start.f_getPlayerNo(2, #start.p[2].t_selected + 1),
					--cursor = {},
					ratioLevel = start.f_getRatio(2, t.p2numratio[cnt]),
				})
				main.t_availableChars = start.f_excludeChar(main.t_availableChars, v)
			end
			-- team conversion if 'single' param is set on randomly added chars
			if onlyme and #start.p[2].t_selected > 1 then
				panicError("Unexpected launchFight state.\nPlease write down everything that lead to this error and report it to K4thos.\n")
				--[[for i = 1, #start.p[2].t_selected do
					if not start.f_getCharData(start.p[2].t_selected[i].ref).single then
						table.insert(main.t_availableChars[t.order], start.p[2].t_selected[i].ref)
						table.remove(start.p[2].t_selected, k)
					end
				end]]
			end
		end
		if onlyme then
			start.p[2].numChars = #start.p[2].t_selected
		end
		-- skip match if needed
		if #start.p[2].t_selected < start.p[2].numChars then
			start.p[2].t_selected = {}
			start.p[2].t_selTemp = {}
			printConsole("launchFight(): not enough P2 characters, skipping execution")
			setMatchNo(matchno() + 1)
			return true --continue lua code execution
		end
	end
	--TODO: fix config.BackgroundLoading setting
	--if config.BackgroundLoading then
	--	selectStart()
	--else
		clearSelected()
	--end
	local ok = false
	local saveData = false
	local loopCount = 0
	while true do
		-- fight initialization
		setTeamMode(1, start.p[1].teamMode, start.p[1].numChars)
		setTeamMode(2, start.p[2].teamMode, start.p[2].numChars)
		start.f_remapAI(t.ai)
		start.f_setRounds(t.roundtime, {t.p1rounds, t.p2rounds})
		t.stageNo = start.f_setStage(t.stageNo, t.stage ~= '' or continue() or loopCount > 0)
		start.f_setMusic(t.stageNo, t.musicData)
		if not start.f_selectVersus(t.vsscreen, t.orderselect) then break end
		start.f_selectLoading()
		start.f_overrideCharData()
		saveData = true
		local continueScreen = main.continueScreen
		local victoryScreen = main.victoryScreen
		main.continueScreen = t.continue
		main.victoryScreen = t.victoryscreen
		hook.run("launchFight")
		_, t_gameStats = start.f_game(t.lua)
		main.continueScreen = continueScreen
		main.victoryScreen = victoryScreen
		clearColor(motif.selectbgdef.bgclearcolor[1], motif.selectbgdef.bgclearcolor[2], motif.selectbgdef.bgclearcolor[3])
		-- here comes a new challenger
		if start.challenger > 0 then
			saveData = false
			if t.challenger then -- end function called by f_arcadeChallenger() regardless of outcome
				ok = not start.exit and not esc()
				break
			elseif not start.f_selectChallenger() then
				start.challenger = 0
				break
			end
		-- player exit the game via ESC
		elseif winnerteam() == -1 then
			if not main.selectMenu[1] and not main.selectMenu[2] then
				setMatchNo(-1)
			end
			break
		-- player lost in modes that ends after 1 lose
		elseif winnerteam() ~= 1 and main.elimination then
			setMatchNo(-1)
			break
		-- player won or continuing is disabled
		elseif winnerteam() == 1 or not t.continue then
			start.p[2].t_selected = {}
			start.p[2].t_selTemp = {}
			setMatchNo(matchno() + 1)
			setContinue(false)
			ok = true -- continue lua code execution
			break
		-- continue = no
		elseif not continue() then
			setMatchNo(-1)
			break
		-- continue = yes
		elseif not t.quickcontinue then -- if 'Quick Continue' is disabled
			start.p[1].t_selected = {}
			start.p[1].t_selTemp = {}
			start.p[1].selEnd = false
			start.launchFightSav = main.f_tableCopy(t)
			--start.p[2].t_selTemp = {} -- uncomment to disable enemy team showing up in select screen
			selScreenEnd = false
			start.f_saveData()
			return
		else
			start.f_saveData()
		end
		start.challenger = 0
		loopCount = loopCount + 1
	end
	if saveData then
		start.f_saveData()
	end
	-- restore original values
	start.p[1].numChars = t.p1numchars
	start.p[1].teamMode = t.p1teammode
	start.p[2].numChars = t.p2numchars
	start.p[2].teamMode = t.p2teammode
	return ok
end

function launchStoryboard(path)
	if path == nil or not main.f_fileExists(path) then
		return false
	end
	storyboard.f_storyboard(path)
	return true
end

function codeInput(name)
	if main.t_commands[name] == nil then
		return false
	end
	if commandGetState(main.t_cmd[main.playerInput], name) then
		return true
	end
	return false
end

--;===========================================================
--; SELECT SCREEN
--;===========================================================
local txt_recordSelect = main.f_createTextImg(motif.select_info, 'record')
local txt_timerSelect = main.f_createTextImg(motif.select_info, 'timer')
local txt_selStage = main.f_createTextImg(motif.select_info, 'stage_active')
local t_txt_name = {}
for i = 1, 2 do
	table.insert(t_txt_name, main.f_createTextImg(motif.select_info, 'p' .. i .. '_name'))
end

if main.t_sort.select_info.teammenu == nil then
	main.t_sort.select_info.teammenu = {'single', 'simul', 'turns'}
end

function start.f_selectScreen()
	if (not main.selectMenu[1] and not main.selectMenu[2]) or selScreenEnd then
		return true
	end
	main.f_bgReset(motif.selectbgdef.bg)
	main.f_fadeReset('fadein', motif.select_info)
	main.f_playBGM(false, motif.music.select_bgm, motif.music.select_bgm_loop, motif.music.select_bgm_volume, motif.music.select_bgm_loopstart, motif.music.select_bgm_loopend)
	start.f_resetTempData(motif.select_info, '_face')
	local stageActiveCount = 0
	local stageActiveType = 'stage_active'
	timerSelect = 0
	local escFlag = false
	local t_teamMenu = {{}, {}}
	local blinkCount = 0
	local counter = 0 - motif.select_info.fadein_time
	-- generate team mode items table
	for side = 1, 2 do
		-- start with all default teammode entires
		local str = 'teammenu_itemname_' .. gamemode() .. '_'
		local t = {
			{data = text:create({}), itemname = 'single', displayname = (motif.select_info[str .. 'single'] or motif.select_info.teammenu_itemname_single), mode = 0, insert = true},
			{data = text:create({}), itemname = 'simul', displayname = (motif.select_info[str .. 'simul'] or motif.select_info.teammenu_itemname_simul), mode = 1, insert = true},
			{data = text:create({}), itemname = 'turns', displayname = (motif.select_info[str .. 'turns'] or motif.select_info.teammenu_itemname_turns), mode = 2, insert = true},
			{data = text:create({}), itemname = 'tag', displayname = (motif.select_info[str .. 'tag'] or motif.select_info.teammenu_itemname_tag), mode = 3, insert = true},
			{data = text:create({}), itemname = 'ratio', displayname = (motif.select_info[str .. 'ratio'] or motif.select_info.teammenu_itemname_ratio), mode = 2, insert = true},
		}
		local activeNum = #t
		-- keep team mode allowed by game mode declaration, but only if it hasn't been disabled by screenpack parameter
		for i = #t, 1, -1 do
			local itemname = t[i].itemname
			if not main.teamMenu[side][itemname]
				or (motif.select_info[str .. itemname] ~= nil and motif.select_info[str .. itemname] == '')
				or (motif.select_info[str .. itemname] == nil and motif.select_info['teammenu_itemname_' .. itemname] == '') then
				t[i].insert = false
				activeNum = activeNum - 1 --track disabled items
			end
		end
		-- first we insert all entries existing in screenpack file in correct order
		for _, name in ipairs(main.f_tableExists(main.t_sort.select_info).teammenu) do
			for k, v in ipairs(t) do
				if v.insert and (name == v.itemname or name == gamemode() .. '_' .. v.itemname) then
					table.insert(t_teamMenu[side], v)
					v.insert = false
					break
				end
			end
		end
		-- then we insert remaining default entries
		for k, v in ipairs(t) do
			if v.insert or (activeNum == 0 and main.teamMenu[side][v.itemname]) then
				table.insert(t_teamMenu[side], v)
				-- if all items are disabled by screenpack add only first default item
				if activeNum == 0 then
					break
				end
			end
		end
	end
	while not selScreenEnd do
		counter = counter + 1
		--credits
		if main.credits ~= -1 and getKey(motif.attract_mode.credits_key) then
			sndPlay(motif.files.snd_data, motif.attract_mode.credits_snd[1], motif.attract_mode.credits_snd[2])
			main.credits = main.credits + 1
			resetKey()
		end
		--draw clearcolor
		clearColor(motif.selectbgdef.bgclearcolor[1], motif.selectbgdef.bgclearcolor[2], motif.selectbgdef.bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(motif.selectbgdef.bg, false)
		--draw title
		main.txt_mainSelect:draw()
		--draw portraits
		for side = 1, 2 do
			if #start.p[side].t_selTemp > 0 then
				start.f_drawPortraits(start.p[side].t_selTemp, side, motif.select_info, '_face', true)
			end
		end
		--draw cell art
		for row = 1, motif.select_info.rows do
			for col = 1, motif.select_info.columns do
				local t = start.t_grid[row][col]
				if t.skip ~= 1 then
					--draw cell background
					if (t.char ~= nil and (t.hidden == 0 or t.hidden == 3)) or motif.select_info.showemptyboxes == 1 then
						main.f_animPosDraw(
							motif.select_info.cell_bg_data,
							motif.select_info.pos[1] + t.x,
							motif.select_info.pos[2] + t.y,
							(motif.select_info['cell_' .. col .. '_' .. row .. '_facing'] or motif.select_info.cell_bg_facing)
						)
					end
					--draw random cell
					if t.char == 'randomselect' or t.hidden == 3 then
						main.f_animPosDraw(
							motif.select_info.cell_random_data,
							motif.select_info.pos[1] + t.x + motif.select_info.portrait_offset[1],
							motif.select_info.pos[2] + t.y + motif.select_info.portrait_offset[2],
							(motif.select_info['cell_' .. col .. '_' .. row .. '_facing'] or motif.select_info.cell_random_facing)
						)
					--draw face cell
					elseif t.char ~= nil and t.hidden == 0 then
						main.f_animPosDraw(
							start.f_getCharData(t.char_ref).cell_data,
							motif.select_info.pos[1] + t.x + motif.select_info.portrait_offset[1],
							motif.select_info.pos[2] + t.y + motif.select_info.portrait_offset[2],
							(motif.select_info['cell_' .. col .. '_' .. row .. '_facing'] or motif.select_info.portrait_facing)
						)
					end
				end
			end
		end
		--draw done cursors
		for side = 1, 2 do
			for _, v in pairs(start.p[side].t_selected) do
				if v.cursor ~= nil then
					--get cell coordinates
					local x = v.cursor[1]
					local y = v.cursor[2]
					local t = start.t_grid[y + 1][x + 1]
					--retrieve proper cell coordinates in case of random selection
					--TODO: doesn't work with slot feature
					--if (t.char == 'randomselect' or t.hidden == 3) --[[and not config.TeamDuplicates]] then
					--	x = start.f_getCharData(v.ref).col - 1
					--	y = start.f_getCharData(v.ref).row - 1
					--	t = start.t_grid[y + 1][x + 1]
					--end
					--render only if cell is not hidden
					if t.hidden ~= 1 and t.hidden ~= 2 then
						start.f_drawCursor(v.pn, x, y, '_cursor_done')
					end
				end
			end
		end
		--team and select menu
		if blinkCount < motif.select_info.p2_cursor_switchtime then
			blinkCount = blinkCount + 1
		else
			blinkCount = 0
		end
		for side = 1, 2 do
			if not start.p[side].teamEnd then
				start.f_teamMenu(side, t_teamMenu[side])
			elseif not start.p[side].selEnd then
				--for each player with active controls
				for k, v in ipairs(start.p[side].t_selCmd) do
					local member = main.f_tableLength(start.p[side].t_selected) + k
					if main.coop and (side == 1 or gamemode('versuscoop')) then
						member = k
					end
					--member selection
					v.selectState = start.f_selectMenu(side, v.cmd, v.player, member, v.selectState)
					--draw active cursor
					if side == 2 and motif.select_info.p2_cursor_blink == 1 then
						local sameCell = false
						for _, v2 in ipairs(start.p[1].t_selCmd) do							
							if start.c[v.player].cell == start.c[v2.player].cell and v.selectState == 0 and v2.selectState == 0 then
								if blinkCount == 0 then
									start.c[v.player].blink = not start.c[v.player].blink
								end
								sameCell = true
								break
							end
						end
						if not sameCell then
							start.c[v.player].blink = false
						end
					end
					if v.selectState < 4 and start.f_selGrid(start.c[v.player].cell + 1).hidden ~= 1 and not start.c[v.player].blink then
						start.f_drawCursor(v.player, start.c[v.player].selX, start.c[v.player].selY, '_cursor_active')
					end
				end
			end
			--delayed screen transition for the duration of face_done_anim or selection sound
			if start.p[side].screenDelay > 0 then
				if main.f_input(main.t_players, {'pal', 's'}) then
					start.p[side].screenDelay = 0
				else
					start.p[side].screenDelay = start.p[side].screenDelay - 1
				end
			end
		end
		--exit select screen
		if not escFlag and (esc() or main.f_input(main.t_players, {'m'})) then
			main.f_fadeReset('fadeout', motif.select_info)
			escFlag = true
		end
		--draw names
		for side = 1, 2 do
			if #start.p[side].t_selTemp > 0 then
				for i = 1, #start.p[side].t_selTemp do
					if i <= motif.select_info['p' .. side .. '_name_num'] or main.coop then
						local name = ''
						if motif.select_info['p' .. side .. '_name_num'] == 1 then
							name = start.f_getName(start.p[side].t_selTemp[#start.p[side].t_selTemp].ref, side)
						else
							name = start.f_getName(start.p[side].t_selTemp[i].ref, side)
						end
						t_txt_name[side]:update({
							font =   motif.select_info['p' .. side .. '_name_font'][1],
							bank =   motif.select_info['p' .. side .. '_name_font'][2],
							align =  motif.select_info['p' .. side .. '_name_font'][3],
							text =   name,
							x =      motif.select_info['p' .. side .. '_name_offset'][1] + (i - 1) * motif.select_info['p' .. side .. '_name_spacing'][1],
							y =      motif.select_info['p' .. side .. '_name_offset'][2] + (i - 1) * motif.select_info['p' .. side .. '_name_spacing'][2],
							scaleX = motif.select_info['p' .. side .. '_name_scale'][1],
							scaleY = motif.select_info['p' .. side .. '_name_scale'][2],
							r =      motif.select_info['p' .. side .. '_name_font'][4],
							g =      motif.select_info['p' .. side .. '_name_font'][5],
							b =      motif.select_info['p' .. side .. '_name_font'][6],
							height = motif.select_info['p' .. side .. '_name_font'][7],
						})
						t_txt_name[side]:draw()
					end
				end
			end
		end
		--team and character selection complete
		if start.p[1].selEnd and start.p[2].selEnd and start.p[1].teamEnd and start.p[2].teamEnd then
			restoreCursor = true
			if main.stageMenu and not stageEnd then --Stage select
				start.f_stageMenu()
			elseif start.p[1].screenDelay <= 0 and start.p[2].screenDelay <= 0 and main.fadeType == 'fadein' then
				main.f_fadeReset('fadeout', motif.select_info)
			end
			--draw stage portrait
			if main.stageMenu then
				--draw stage portrait background
				main.f_animPosDraw(motif.select_info.stage_portrait_bg_data)
				--draw stage portrait (random)
				if stageListNo == 0 then
					main.f_animPosDraw(motif.select_info.stage_portrait_random_data)
				--draw stage portrait loaded from stage SFF
				else
					main.f_animPosDraw(
						main.t_selStages[main.t_selectableStages[stageListNo]].anim_data,
						motif.select_info.stage_pos[1] + motif.select_info.stage_portrait_offset[1],
						motif.select_info.stage_pos[2] + motif.select_info.stage_portrait_offset[2]
					)
				end
				if not stageEnd then
					if main.f_input(main.t_players, {'pal', 's'}) or timerSelect == -1 then
						sndPlay(motif.files.snd_data, motif.select_info.stage_done_snd[1], motif.select_info.stage_done_snd[2])
						stageActiveType = 'stage_done'
						stageEnd = true
					elseif stageActiveCount < motif.select_info.stage_active_switchtime then --delay change
						stageActiveCount = stageActiveCount + 1
					else
						if stageActiveType == 'stage_active' then
							stageActiveType = 'stage_active2'
						else
							stageActiveType = 'stage_active'
						end
						stageActiveCount = 0
					end
				end
				--draw stage name
				local t_txt = {}
				if stageListNo == 0 then
					t_txt[1] = motif.select_info.stage_random_text
				else
					t = motif.select_info.stage_text:gsub('%%i', tostring(stageListNo))
					t = t:gsub('\n', '\\n')
					t = t:gsub('%%s', main.t_selStages[main.t_selectableStages[stageListNo]].name)
					for i, c in ipairs(main.f_strsplit('\\n', t)) do --split string using "\n" delimiter
						t_txt[i] = c
					end
				end
				for i = 1, #t_txt do
					txt_selStage:update({
						font =   motif.select_info[stageActiveType .. '_font'][1],
						bank =   motif.select_info[stageActiveType .. '_font'][2],
						align =  motif.select_info[stageActiveType .. '_font'][3],
						text =   t_txt[i],
						x =      motif.select_info.stage_pos[1] + motif.select_info[stageActiveType .. '_offset'][1],
						y =      motif.select_info.stage_pos[2] + motif.select_info[stageActiveType .. '_offset'][2] + main.f_ySpacing(motif.select_info, stageActiveType) * (i - 1),
						scaleX = motif.select_info[stageActiveType .. '_scale'][1],
						scaleY = motif.select_info[stageActiveType .. '_scale'][2],
						r =      motif.select_info[stageActiveType .. '_font'][4],
						g =      motif.select_info[stageActiveType .. '_font'][5],
						b =      motif.select_info[stageActiveType .. '_font'][6],
						height = motif.select_info[stageActiveType .. '_font'][7],
					})
					txt_selStage:draw()
				end
			end
		end
		--draw timer
		if motif.select_info.timer_count ~= -1 and (not start.p[1].teamEnd or not start.p[2].teamEnd or not start.p[1].selEnd or not start.p[2].selEnd or (main.stageMenu and not stageEnd)) and counter >= 0 then
			timerSelect = main.f_drawTimer(timerSelect, motif.select_info, 'timer_', txt_timerSelect)
		end
		--draw record text
		for i = 1, #t_recordText do
			txt_recordSelect:update({
				text = t_recordText[i],
				y = motif.select_info.record_offset[2] + main.f_ySpacing(motif.select_info, 'record') * (i - 1),
			})
			txt_recordSelect:draw()
		end
		-- hook
		hook.run("start.f_selectScreen")
		--draw layerno = 1 backgrounds
		bgDraw(motif.selectbgdef.bg, true)
		--draw fadein / fadeout
		main.f_fadeAnim(motif.select_info)
		--frame transition
		if not main.f_frameChange() then
			selScreenEnd = true
			break --skip last frame rendering
		end
		main.f_refresh()
	end
	return not escFlag
end

--;===========================================================
--; TEAM MENU
--;===========================================================
local t_txt_teamSelfTitle = {}
local t_txt_teamEnemyTitle = {}
for i = 1, 2 do
	table.insert(t_txt_teamSelfTitle, main.f_createTextImg(motif.select_info, 'p' .. i .. '_teammenu_selftitle', {x = motif.select_info['p' .. i .. '_teammenu_pos'][1], y = motif.select_info['p' .. i .. '_teammenu_pos'][2]}))
	table.insert(t_txt_teamEnemyTitle, main.f_createTextImg(motif.select_info, 'p' .. i .. '_teammenu_enemytitle', {x = motif.select_info['p' .. i .. '_teammenu_pos'][1], y = motif.select_info['p' .. i .. '_teammenu_pos'][2]}))
end
local t_teamActiveCount = {0, 0}
local t_teamActiveType = {'p1_teammenu_item_active', 'p2_teammenu_item_active'}

function start.f_teamMenu(side, t)
	if #t == 0 then
		start.p[side].teamEnd = true
		return
	end
	--skip selection if only 1 team mode is available and team size is fixed
	if #t == 1 and (t[1].itemname == 'single' or (t[1].itemname == 'simul' and main.numSimul[1] == main.numSimul[2]) or (t[1].itemname == 'turns' and main.numTurns[1] == main.numTurns[2]) or (t[1].itemname == 'tag' and main.numTag[1] == main.numTag[2])) then
		if t[1].itemname == 'single' then
			start.p[side].numChars = 1
		elseif t[1].itemname == 'simul' then
			start.p[side].numChars = start.p[side].numSimul
		elseif t[1].itemname == 'turns' then
			start.p[side].numChars = start.p[side].numTurns
		elseif t[1].itemname == 'tag' then
			start.p[side].numChars = start.p[side].numTag
		end
		start.p[side].teamMode = t[1].mode
		start.p[side].teamEnd = true
	--otherwise display team mode selection
	else
		--Commands
		local t_cmd = {}
		if main.coop then
			for i = 1, config.Players do
				if not gamemode('versuscoop') or (i - 1) % 2 + 1 == side then
					table.insert(t_cmd, i)
				end
			end
		else
			t_cmd = {side}
		end
		--Calculate team cursor position
		if start.p[side].teamMenu > #t then
			start.p[side].teamMenu = 1
		end
		if #t > 1 and main.f_input(t_cmd, main.f_extractKeys(motif.select_info['p' .. side .. '_teammenu_previous_key'])) then
			if start.p[side].teamMenu > 1 then
				sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_move_snd'][1], motif.select_info['p' .. side .. '_teammenu_move_snd'][2])
				start.p[side].teamMenu = start.p[side].teamMenu - 1
			elseif motif.select_info.teammenu_move_wrapping == 1 then
				sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_move_snd'][1], motif.select_info['p' .. side .. '_teammenu_move_snd'][2])
				start.p[side].teamMenu = #t
			end
		elseif #t > 1 and main.f_input(t_cmd, main.f_extractKeys(motif.select_info['p' .. side .. '_teammenu_next_key'])) then
			if start.p[side].teamMenu < #t then
				sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_move_snd'][1], motif.select_info['p' .. side .. '_teammenu_move_snd'][2])
				start.p[side].teamMenu = start.p[side].teamMenu + 1
			elseif motif.select_info.teammenu_move_wrapping == 1 then
				sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_move_snd'][1], motif.select_info['p' .. side .. '_teammenu_move_snd'][2])
				start.p[side].teamMenu = 1
			end
		else
			if t[start.p[side].teamMenu].itemname == 'simul' then
				if main.f_input(t_cmd, main.f_extractKeys(motif.select_info['p' .. side .. '_teammenu_subtract_key'])) then
					if start.p[side].numSimul > main.numSimul[1] then
						sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_value_snd'][1], motif.select_info['p' .. side .. '_teammenu_value_snd'][2])
						start.p[side].numSimul = start.p[side].numSimul - 1
					end
				elseif main.f_input(t_cmd, main.f_extractKeys(motif.select_info['p' .. side .. '_teammenu_add_key'])) then
					if start.p[side].numSimul < main.numSimul[2] then
						sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_value_snd'][1], motif.select_info['p' .. side .. '_teammenu_value_snd'][2])
						start.p[side].numSimul = start.p[side].numSimul + 1
					end
				end
			elseif t[start.p[side].teamMenu].itemname == 'turns' then
				if main.f_input(t_cmd, main.f_extractKeys(motif.select_info['p' .. side .. '_teammenu_subtract_key'])) then
					if start.p[side].numTurns > main.numTurns[1] then
						sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_value_snd'][1], motif.select_info['p' .. side .. '_teammenu_value_snd'][2])
						start.p[side].numTurns = start.p[side].numTurns - 1
					end
				elseif main.f_input(t_cmd, main.f_extractKeys(motif.select_info['p' .. side .. '_teammenu_add_key'])) then
					if start.p[side].numTurns < main.numTurns[2] then
						sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_value_snd'][1], motif.select_info['p' .. side .. '_teammenu_value_snd'][2])
						start.p[side].numTurns = start.p[side].numTurns + 1
					end
				end
			elseif t[start.p[side].teamMenu].itemname == 'tag' then
				if main.f_input(t_cmd, main.f_extractKeys(motif.select_info['p' .. side .. '_teammenu_subtract_key'])) then
					if start.p[side].numTag > main.numTag[1] then
						sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_value_snd'][1], motif.select_info['p' .. side .. '_teammenu_value_snd'][2])
						start.p[side].numTag = start.p[side].numTag - 1
					end
				elseif main.f_input(t_cmd, main.f_extractKeys(motif.select_info['p' .. side .. '_teammenu_add_key'])) then
					if start.p[side].numTag < main.numTag[2] then
						sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_value_snd'][1], motif.select_info['p' .. side .. '_teammenu_value_snd'][2])
						start.p[side].numTag = start.p[side].numTag + 1
					end
				end
			elseif t[start.p[side].teamMenu].itemname == 'ratio' then
				if main.f_input(t_cmd, main.f_extractKeys(motif.select_info['p' .. side .. '_teammenu_subtract_key'])) and main.selectMenu[side] then
					sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_value_snd'][1], motif.select_info['p' .. side .. '_teammenu_value_snd'][2])
					if start.p[side].numRatio > 1 then
						start.p[side].numRatio = start.p[side].numRatio - 1
					else
						start.p[side].numRatio = 7
					end
				elseif main.f_input(t_cmd, main.f_extractKeys(motif.select_info['p' .. side .. '_teammenu_add_key'])) and main.selectMenu[side] then
					sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_value_snd'][1], motif.select_info['p' .. side .. '_teammenu_value_snd'][2])
					if start.p[side].numRatio < 7 then
						start.p[side].numRatio = start.p[side].numRatio + 1
					else
						start.p[side].numRatio = 1
					end
				end
			end
		end
		--Draw team background
		main.f_animPosDraw(motif.select_info['p' .. side .. '_teammenu_bg_data'])
		--Draw team title
		if side == 2 and main.cpuSide[2] then
			main.f_animPosDraw(motif.select_info['p' .. side .. '_teammenu_enemytitle_data'])
			t_txt_teamEnemyTitle[side]:draw()
		else
			main.f_animPosDraw(motif.select_info['p' .. side .. '_teammenu_selftitle_data'])
			t_txt_teamSelfTitle[side]:draw()
		end
		--Draw team cursor
		main.f_animPosDraw(
			motif.select_info['p' .. side .. '_teammenu_item_cursor_data'],
			(start.p[side].teamMenu - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][1],
			(start.p[side].teamMenu - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][2]
		)
		for i = 1, #t do
			--Draw team items
			if i == start.p[side].teamMenu then
				if t_teamActiveCount[side] < motif.select_info['p' .. side .. '_teammenu_item_active_switchtime'] then --delay change
					t_teamActiveCount[side] = t_teamActiveCount[side] + 1
				else
					if t_teamActiveType[side] == 'p' .. side .. '_teammenu_item_active' then
						t_teamActiveType[side] = 'p' .. side .. '_teammenu_item_active2'
					else
						t_teamActiveType[side] = 'p' .. side .. '_teammenu_item_active'
					end
					t_teamActiveCount[side] = 0
				end
				--Draw team active item background
				main.f_animPosDraw(motif.select_info['p' .. side .. '_teammenu_bg_active_' .. gamemode() .. '_' .. t[i].itemname .. '_data'] or motif.select_info['p' .. side .. '_teammenu_bg_active_' .. t[i].itemname .. '_data'])
				--Draw team active item font
				t[i].data:update({
					font =   motif.select_info[t_teamActiveType[side] .. '_font'][1],
					bank =   motif.select_info[t_teamActiveType[side] .. '_font'][2],
					align =  motif.select_info[t_teamActiveType[side] .. '_font'][3], --winmugen ignores active font facing? Fixed in mugen 1.0
					text =   t[i].displayname,
					x =      motif.select_info['p' .. side .. '_teammenu_pos'][1] + motif.select_info['p' .. side .. '_teammenu_item_offset'][1] + motif.select_info['p' .. side .. '_teammenu_item_spacing'][1] * (i - 1),
					y =      motif.select_info['p' .. side .. '_teammenu_pos'][2] + motif.select_info['p' .. side .. '_teammenu_item_offset'][2] + motif.select_info['p' .. side .. '_teammenu_item_spacing'][2] * (i - 1),
					scaleX = motif.select_info[t_teamActiveType[side] .. '_scale'][1],
					scaleY = motif.select_info[t_teamActiveType[side] .. '_scale'][2],
					r =      motif.select_info[t_teamActiveType[side] .. '_font'][4],
					g =      motif.select_info[t_teamActiveType[side] .. '_font'][5],
					b =      motif.select_info[t_teamActiveType[side] .. '_font'][6],
					height = motif.select_info[t_teamActiveType[side] .. '_font'][7],
				})
				t[i].data:draw()
			else
				--Draw team not active item background
				main.f_animPosDraw(motif.select_info['p' .. side .. '_teammenu_bg_' .. gamemode() .. '_' .. t[i].itemname .. '_data'] or motif.select_info['p' .. side .. '_teammenu_bg_' .. t[i].itemname .. '_data'])
				--Draw team not active item font
				t[i].data:update({
					font =   motif.select_info['p' .. side .. '_teammenu_item_font'][1],
					bank =   motif.select_info['p' .. side .. '_teammenu_item_font'][2],
					align =  motif.select_info['p' .. side .. '_teammenu_item_font'][3], --winmugen ignores active font facing? Fixed in mugen 1.0
					text =   t[i].displayname,
					x =      motif.select_info['p' .. side .. '_teammenu_pos'][1] + motif.select_info['p' .. side .. '_teammenu_item_offset'][1] + motif.select_info['p' .. side .. '_teammenu_item_spacing'][1] * (i - 1),
					y =      motif.select_info['p' .. side .. '_teammenu_pos'][2] + motif.select_info['p' .. side .. '_teammenu_item_offset'][2] + motif.select_info['p' .. side .. '_teammenu_item_spacing'][2] * (i - 1),
					scaleX = motif.select_info['p' .. side .. '_teammenu_item_scale'][1],
					scaleY = motif.select_info['p' .. side .. '_teammenu_item_scale'][2],
					r =      motif.select_info['p' .. side .. '_teammenu_item_font'][4],
					g =      motif.select_info['p' .. side .. '_teammenu_item_font'][5],
					b =      motif.select_info['p' .. side .. '_teammenu_item_font'][6],
					height = motif.select_info['p' .. side .. '_teammenu_item_font'][7],
				})
				t[i].data:draw()
			end
			--Draw team icons
			if t[i].itemname == 'simul' then
				for j = 1, main.numSimul[2] do
					if j <= start.p[side].numSimul then
						main.f_animPosDraw(
							motif.select_info['p' .. side .. '_teammenu_value_icon_data'],
							(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][1] + (j - 1) * motif.select_info['p' .. side .. '_teammenu_value_spacing'][1],
							(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][2] + (j - 1) * motif.select_info['p' .. side .. '_teammenu_value_spacing'][2]
						)
					else
						main.f_animPosDraw(
							motif.select_info['p' .. side .. '_teammenu_value_empty_icon_data'],
							(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][1] + (j - 1) * motif.select_info['p' .. side .. '_teammenu_value_spacing'][1],
							(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][2] + (j - 1) * motif.select_info['p' .. side .. '_teammenu_value_spacing'][2]
						)
					end
				end
			elseif t[i].itemname == 'turns' then
				for j = 1, main.numTurns[2] do
					if j <= start.p[side].numTurns then
						main.f_animPosDraw(
							motif.select_info['p' .. side .. '_teammenu_value_icon_data'],
							(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][1] + (j - 1) * motif.select_info['p' .. side .. '_teammenu_value_spacing'][1],
							(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][2] + (j - 1) * motif.select_info['p' .. side .. '_teammenu_value_spacing'][2]
						)
					else
						main.f_animPosDraw(
							motif.select_info['p' .. side .. '_teammenu_value_empty_icon_data'],
							(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][1] + (j - 1) * motif.select_info['p' .. side .. '_teammenu_value_spacing'][1],
							(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][2] + (j - 1) * motif.select_info['p' .. side .. '_teammenu_value_spacing'][2]
						)
					end
				end
			elseif t[i].itemname == 'tag' then
				for j = 1, main.numTag[2] do
					if j <= start.p[side].numTag then
						main.f_animPosDraw(
							motif.select_info['p' .. side .. '_teammenu_value_icon_data'],
							(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][1] + (j - 1) * motif.select_info['p' .. side .. '_teammenu_value_spacing'][1],
							(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][2] + (j - 1) * motif.select_info['p' .. side .. '_teammenu_value_spacing'][2]
						)
					else
						main.f_animPosDraw(
							motif.select_info['p' .. side .. '_teammenu_value_empty_icon_data'],
							(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][1] + (j - 1) * motif.select_info['p' .. side .. '_teammenu_value_spacing'][1],
							(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][2] + (j - 1) * motif.select_info['p' .. side .. '_teammenu_value_spacing'][2]
						)
					end
				end
			elseif t[i].itemname == 'ratio' and start.p[side].teamMenu == i and main.selectMenu[side] then
				main.f_animPosDraw(
					motif.select_info['p' .. side .. '_teammenu_ratio' .. start.p[side].numRatio .. '_icon_data'],
					(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][1],
					(i - 1) * motif.select_info['p' .. side .. '_teammenu_item_spacing'][2]
				)
			end
		end
		--Confirmed team selection
		if main.f_input(t_cmd, main.f_extractKeys(motif.select_info['p' .. side .. '_teammenu_accept_key'])) or timerSelect == -1 then
			timerSelect = motif.select_info.timer_displaytime
			sndPlay(motif.files.snd_data, motif.select_info['p' .. side .. '_teammenu_done_snd'][1], motif.select_info['p' .. side .. '_teammenu_done_snd'][2])
			if t[start.p[side].teamMenu].itemname == 'single' then
				start.p[side].teamMode = t[start.p[side].teamMenu].mode
				start.p[side].numChars = 1
			elseif t[start.p[side].teamMenu].itemname == 'simul' then
				start.p[side].teamMode = t[start.p[side].teamMenu].mode
				start.p[side].numChars = start.p[side].numSimul
			elseif t[start.p[side].teamMenu].itemname == 'turns' then
				start.p[side].teamMode = t[start.p[side].teamMenu].mode
				start.p[side].numChars = start.p[side].numTurns
			elseif t[start.p[side].teamMenu].itemname == 'tag' then
				start.p[side].teamMode = t[start.p[side].teamMenu].mode
				start.p[side].numChars = start.p[side].numTag
			elseif t[start.p[side].teamMenu].itemname == 'ratio' then
				start.p[side].teamMode = t[start.p[side].teamMenu].mode
				if start.p[side].numRatio <= 3 then
					start.p[side].numChars = 3
				elseif start.p[side].numRatio <= 6 then
					start.p[side].numChars = 2
				else
					start.p[side].numChars = 1
				end
				start.p[side].ratio = true
			end
			start.p[side].teamEnd = true
			main.f_cmdBufReset(side)
		end
	end
	--t_selCmd table appending once team mode selection is finished
	if start.p[side].teamEnd then
		if main.coop and (side == 1 or gamemode('versuscoop')) then
			for i = 1, start.p[side].numChars do
				if gamemode('versuscoop') then
					if side == 1 then
						table.insert(start.p[side].t_selCmd, {cmd = i * 2 - 1, player = start.f_getPlayerNo(side, #start.p[side].t_selCmd + 1), selectState = 0})
					else
						table.insert(start.p[side].t_selCmd, {cmd = i * 2, player = start.f_getPlayerNo(side, #start.p[side].t_selCmd + 1), selectState = 0})
					end
				else
					table.insert(start.p[1].t_selCmd, {cmd = i, player = start.f_getPlayerNo(side, #start.p[1].t_selCmd + 1), selectState = 0})
				end
			end
		else
			table.insert(start.p[side].t_selCmd, {cmd = side, player = start.f_getPlayerNo(side, #start.p[side].t_selCmd + 1), selectState = 0})
		end
	end
end

--;===========================================================
--; SELECT MENU
--;===========================================================
function start.f_selectMenu(side, cmd, player, member, selectState)
	--predefined selection
	if main.forceChar[side] ~= nil then
		local t = {}
		for _, v in ipairs(main.forceChar[side]) do
			if t[v] == nil then
				t[v] = ''
			end
			table.insert(start.p[side].t_selected, {
				ref = v,
				pal = start.f_selectPal(v),
				--pn = start.f_getPlayerNo(side, #start.p[side].t_selected + 1),
				--cursor = = {},
				--ratioLevel = start.f_getRatio(side),
			})
		end
		start.p[side].selEnd = true
		return 0
	--manual selection
	elseif not start.p[side].selEnd then
		--cell not selected yet
		if selectState == 0 then
			--restore cursor coordinates
			if restoreCursor then
				-- remove entries if stored cursors exceeds team size
				if #start.p[side].t_cursor > start.p[side].numChars then
					for i = #start.p[side].t_cursor, start.p[side].numChars + 1, -1 do
						start.p[side].t_cursor[i] = nil
					end
				end
				-- restore saved position
				if start.p[side].t_cursor[member] ~= nil then
					local selX = start.p[side].t_cursor[member].x
					local selY = start.p[side].t_cursor[member].y
					if config.TeamDuplicates or t_reservedChars[side][start.t_grid[selY + 1][selX + 1].char_ref] == nil then
						start.c[player].selX = selX
						start.c[player].selY = selY
					end
					start.p[side].t_cursor[member] = nil
				end
			end
			--calculate current position
			start.c[player].selX, start.c[player].selY = start.f_cellMovement(start.c[player].selX, start.c[player].selY, cmd, side, start.f_getCursorData(player, '_cursor_move_snd'))
			start.c[player].cell = start.c[player].selX + motif.select_info.columns * start.c[player].selY
			start.c[player].selRef = start.f_selGrid(start.c[player].cell + 1).char_ref
			-- temp data not existing yet
			if start.p[side].t_selTemp[member] == nil then
				table.insert(start.p[side].t_selTemp, {
					ref = start.c[player].selRef,
					cell = start.c[player].cell,
					anim = motif.select_info['p' .. side .. '_member' .. member .. '_face_anim'] or motif.select_info['p' .. side .. '_face_anim'],
					anim_data = start.f_animGet(start.c[player].selRef, side, member, motif.select_info, '_face', '', true),
					face2_data = start.f_animGet(start.c[player].selRef, side, member, motif.select_info, '_face2', '', true),
					slide_dist = {0, 0},
				})
			else
				local updateAnim = false
				local slotSelected = start.f_slotSelected(start.c[player].cell + 1, side, cmd, player, start.c[player].selX, start.c[player].selY)
				-- cursor changed position or character change within current slot
				if start.p[side].t_selTemp[member].cell ~= start.c[player].cell or start.p[side].t_selTemp[member].ref ~= start.c[player].selRef then
					--start.p[side].t_selTemp[member].pal = 1
					start.p[side].t_selTemp[member].ref = start.c[player].selRef
					start.p[side].t_selTemp[member].cell = start.c[player].cell
					start.p[side].t_selTemp[member].anim = motif.select_info['p' .. side .. '_member' .. member .. '_face_anim'] or motif.select_info['p' .. side .. '_face_anim']
					start.p[side].t_selTemp[member].slide_dist = {0, 0}
					updateAnim = true
				end
				-- cursor at randomselect cell
				if start.f_selGrid(start.c[player].cell + 1).char == 'randomselect' or start.f_selGrid(start.c[player].cell + 1).hidden == 3 then
					if start.c[player].randCnt > 0 then
						start.c[player].randCnt = start.c[player].randCnt - 1
						start.c[player].selRef = start.c[player].randRef
					else
						if motif.select_info.random_move_snd_cancel == 1 then
							sndStop(motif.files.snd_data, start.f_getCursorData(player, '_random_move_snd')[1], start.f_getCursorData(player, '_random_move_snd')[2])
						end
						sndPlay(motif.files.snd_data, start.f_getCursorData(player, '_random_move_snd')[1], start.f_getCursorData(player, '_random_move_snd')[2])
						start.c[player].randCnt = motif.select_info.cell_random_switchtime
						start.c[player].selRef = start.f_randomChar(side)
						if start.c[player].randRef ~= start.c[player].selRef or start.p[side].t_selTemp[member].anim_data == nil then
							updateAnim = true
							start.c[player].randRef = start.c[player].selRef
						end
					end
				end
				-- update anim data
				if updateAnim then
					start.p[side].t_selTemp[member].anim_data = start.f_animGet(start.c[player].selRef, side, member, motif.select_info, '_face', '', true)
					start.p[side].t_selTemp[member].face2_data = start.f_animGet(start.c[player].selRef, side, member, motif.select_info, '_face2', '', true)
				end
				-- cell selected or select screen timer reached 0
				if (slotSelected and start.f_selGrid(start.c[player].cell + 1).char ~= nil and start.f_selGrid(start.c[player].cell + 1).hidden ~= 2) or (motif.select_info.timer_count ~= -1 and timerSelect == -1) then
					sndPlay(motif.files.snd_data, start.f_getCursorData(player, '_cursor_done_snd')[1], start.f_getCursorData(player, '_cursor_done_snd')[2])
					start.f_playWave(start.c[player].selRef, 'cursor', motif.select_info['p' .. side .. '_select_snd'][1], motif.select_info['p' .. side .. '_select_snd'][2])
					start.p[side].t_selTemp[member].pal = main.f_btnPalNo(cmd)
					if start.p[side].t_selTemp[member].pal == nil or start.p[side].t_selTemp[member].pal == 0 then
						start.p[side].t_selTemp[member].pal = 1
					end
					-- if select anim differs from done anim and coop or pX.face.num allows to display more than 1 portrait or it's the last team member
					local done_anim = motif.select_info['p' .. side .. '_member' .. member .. '_face_done_anim'] or motif.select_info['p' .. side .. '_face_done_anim']
					if done_anim ~= -1 and start.p[side].t_selTemp[member].anim ~= done_anim and (main.coop or motif.select_info['p' .. side .. '_face_num'] > 1 or main.f_tableLength(start.p[side].t_selected) + 1 == start.p[side].numChars) then
						start.p[side].t_selTemp[member].anim_data = start.f_animGet(start.c[player].selRef, side, member, motif.select_info, '_face', '_done', false) or start.p[side].t_selTemp[member].anim_data
						if start.p[side].t_selTemp[member].anim_data ~= nil then
							start.p[side].screenDelay = math.min(120, math.max(start.p[side].screenDelay, animGetLength(start.p[side].t_selTemp[member].anim_data)))
						end
					end
					start.p[side].t_selTemp[member].ref = start.c[player].selRef
					main.f_cmdBufReset(cmd)
					selectState = 1
				end
			end
		--selection menu
		elseif selectState == 1 then
			--TODO: hook left for optional menu that shows up after selecting character (groove, palette selection etc.)
			--once everything is ready set selectState to 3 to confirm character selection
			selectState = 3
		--confirm selection
		elseif selectState == 3 then
			start.p[side].t_selected[member] = {
				ref = start.c[player].selRef,
				pal = start.f_selectPal(start.c[player].selRef, start.p[side].t_selTemp[member].pal),
				pn = start.f_getPlayerNo(side, member),
				cursor = {start.c[player].selX, start.c[player].selY},
				ratioLevel = start.f_getRatio(side),
			}
			if not config.TeamDuplicates then
				t_reservedChars[side][start.c[player].selRef] = true
			end
			start.p[side].t_cursor[member] = {x = start.c[player].selX, y = start.c[player].selY}
			if main.f_tableLength(start.p[side].t_selected) == start.p[side].numChars then --if all characters have been chosen
				if side == 1 and main.cpuSide[2] and start.reset then --if player1 is allowed to select p2 characters
					if timerSelect == -1 then
						start.p[2].teamMode = start.p[1].teamMode
						start.p[2].numChars = start.p[1].numChars
						start.c[2].cell = start.c[1].cell
						start.c[2].selX = start.c[1].selX
						start.c[2].selY = start.c[1].selY
					else
						start.p[2].teamEnd = false
					end
				end
				start.p[side].selEnd = true
			elseif not config.TeamDuplicates and start.t_grid[start.c[player].selY + 1][start.c[player].selX + 1].char ~= 'randomselect' then
				local t_dirs = {'F', 'B', 'D', 'U'}
				if start.c[player].selY + 1 >= motif.select_info.rows then --next row not visible on the screen
					t_dirs = {'F', 'B', 'U', 'D'}
				end
				for _, v in ipairs(t_dirs) do
					local selX, selY = start.f_cellMovement(start.c[player].selX, start.c[player].selY, cmd, side, start.f_getCursorData(player, '_cursor_move_snd'), v)
					if start.t_grid[selY + 1][selX + 1].char ~= nil and (selX ~= start.c[player].selX or selY ~= start.c[player].selY) then
						start.c[player].selX, start.c[player].selY = selX, selY
						break
					end
				end
			end
			if not start.p[1].teamEnd or not start.p[2].teamEnd or not start.p[1].selEnd or not start.p[2].selEnd then
				timerSelect = motif.select_info.timer_displaytime
			end
			if main.coop and (side == 1 or gamemode('versuscoop')) then --remaining members are controlled by different players
				selectState = 4
			elseif not start.p[side].selEnd then --next member controlled by this player should become selectable
				selectState = 0
			end
		end
	end
	return selectState
end

--;===========================================================
--; STAGE MENU
--;===========================================================
function start.f_stageMenu()
	local n = stageListNo
	if timerSelect == -1 then
		stageEnd = true
		return
	elseif main.f_input(main.t_players, {'$B'}) then
		sndPlay(motif.files.snd_data, motif.select_info.stage_move_snd[1], motif.select_info.stage_move_snd[2])
		stageListNo = stageListNo - 1
		if stageListNo < 0 then stageListNo = #main.t_selectableStages end
	elseif main.f_input(main.t_players, {'$F'}) then
		sndPlay(motif.files.snd_data, motif.select_info.stage_move_snd[1], motif.select_info.stage_move_snd[2])
		stageListNo = stageListNo + 1
		if stageListNo > #main.t_selectableStages then stageListNo = 0 end
	elseif main.f_input(main.t_players, {'$U'}) then
		sndPlay(motif.files.snd_data, motif.select_info.stage_move_snd[1], motif.select_info.stage_move_snd[2])
		for i = 1, 10 do
			stageListNo = stageListNo - 1
			if stageListNo < 0 then stageListNo = #main.t_selectableStages end
		end
	elseif main.f_input(main.t_players, {'$D'}) then
		sndPlay(motif.files.snd_data, motif.select_info.stage_move_snd[1], motif.select_info.stage_move_snd[2])
		for i = 1, 10 do
			stageListNo = stageListNo + 1
			if stageListNo > #main.t_selectableStages then stageListNo = 0 end
		end
	end
	if n ~= stageListNo and stageListNo > 0 then
		animReset(main.t_selStages[main.t_selectableStages[stageListNo]].anim_data)
		animUpdate(main.t_selStages[main.t_selectableStages[stageListNo]].anim_data)
	end
end

--;===========================================================
--; VERSUS SCREEN / ORDER SELECTION
--;===========================================================
local txt_matchNo = main.f_createTextImg(motif.vs_screen, 'match')
local t_txt_nameVS = {}
for i = 1, 2 do
	table.insert(t_txt_nameVS, main.f_createTextImg(motif.vs_screen, 'p' .. i .. '_name'))
end
txt_timerVS = main.f_createTextImg(motif.vs_screen, 'timer')

function start.f_selectVersus(active, t_orderSelect)
	start.t_orderRemap = {{}, {}}
	for side = 1, 2 do
		-- populate order remap table with default values
		for i = 1, #start.p[side].t_selected do
			table.insert(start.t_orderRemap[side], i)
		end
		-- prevent order select if not enabled in screenpack or if team size = 1
		if t_orderSelect[side] then
			t_orderSelect[side] = motif.vs_screen.orderselect_enabled == 1 and #start.p[side].t_selected > 1
		end
		-- reset loading flags
		for _, v in ipairs(start.p[side].t_selected) do
			v.loading = false
		end
	end
	-- skip versus screen if vs screen is disabled or p2 side char has vsscreen select.def flag set to 0
	for _, v in ipairs(start.p[2].t_selected) do
		if start.f_getCharData(v.ref).vsscreen == 0 then
			active = false
			break
		end
	end
	if not active then
		clearColor(motif.versusbgdef.bgclearcolor[1], motif.versusbgdef.bgclearcolor[2], motif.versusbgdef.bgclearcolor[3])
		return true
	end
	local text = main.f_extractText(motif.vs_screen.match_text, matchno())
	txt_matchNo:update({text = text[1]})
	main.f_bgReset(motif.versusbgdef.bg)
	main.f_fadeReset('fadein', motif.vs_screen)
	main.f_playBGM(false, motif.music.vs_bgm, motif.music.vs_bgm_loop, motif.music.vs_bgm_volume, motif.music.vs_bgm_loopstart, motif.music.vs_bgm_loopend)
	start.f_resetTempData(motif.vs_screen, '')
	start.f_playWave(getStageNo(), 'stage', motif.vs_screen.stage_snd[1], motif.vs_screen.stage_snd[2])
	local counter = 0 - motif.vs_screen.fadein_time
	local done = (not t_orderSelect[1] and not t_orderSelect[2]) -- both sides having order disabled
		or (not t_orderSelect[1] and main.cpuSide[2]) -- left side with disabled order, right side controlled by CPU
		or (not t_orderSelect[2] and main.cpuSide[1]) -- right side with disabled order, left side controlled by CPU
		or (main.cpuSide[1] and main.cpuSide[2]) -- both sides controlled by CPU
	local timerActive = not done
	local timerCount = 0
	local escFlag = false
	local t_order = {{}, {}}
	local t_icon = {'_icon', '_icon'}
	while true do
		local snd = false
		-- for each team side member
		for side = 1, 2 do
			for k, v in ipairs(start.p[side].t_selected) do
				-- until loading flag is set
				if not v.loading then
					-- if not valid for order selection or CPU or doesn't have key for this member assigned, or order timer run out
					if not t_orderSelect[side] or main.cpuSide[side] or (motif.vs_screen['p' .. side .. '_member' .. k .. '_key'] == nil and #t_order[side] == k - 1) or timerCount == -1 then
						table.insert(t_order[side], k)
						-- if it's the last unordered team member
						if #start.p[side].t_selected == #t_order[side] then
							-- randomize CPU side team order (if valid for order selection)
							if main.cpuSide[side] and t_orderSelect[side] then
								main.f_tableShuffle(t_order[side])
							end
							-- confirm char selection (starts loading immediately if config.BackgroundLoading is true)
							for _, member in ipairs(t_order[side]) do
								if not start.p[side].t_selected[member].loading then
									selectChar(side, start.p[side].t_selected[member].ref, start.p[side].t_selected[member].pal)
									start.p[side].t_selected[member].loading = true
								end
							end
							t_icon[side] = nil
							-- play sound if timer run out
							if not snd and timerCount == -1 then
								sndPlay(motif.files.snd_data, motif.vs_screen['p' .. side .. '_value_snd'][1], motif.vs_screen['p' .. side .. '_value_snd'][2])
								snd = true
							end
						end
					elseif motif.vs_screen['p' .. side .. '_member' .. k .. '_key'] ~= nil and main.f_input({side}, main.f_extractKeys(motif.vs_screen['p' .. side .. '_member' .. k .. '_key'])) or (#start.p[side].t_selected == #t_order[side] + 1) then
						table.insert(t_order[side], k)
						-- confirm char selection (starts loading immediately if config.BackgroundLoading is true)
						selectChar(side, v.ref, v.pal)
						v.loading = true
						-- if it's the last unordered team member
						if #start.p[side].t_selected == #t_order[side] then
							t_icon[side] = nil
						end
						-- play sound only once in particular frame
						if not snd then
							sndPlay(motif.files.snd_data, motif.vs_screen['p' .. side .. '_value_snd'][1], motif.vs_screen['p' .. side .. '_value_snd'][2])
							snd = true
						end
						-- reset pressed button to prevent remapped P2 from registering P1 input
						main.f_cmdBufReset(side)
					end
				end
			end
		end
		-- do once if both sides confirmed order selection
		if not done and #start.p[1].t_selected == #t_order[1] and #start.p[2].t_selected == #t_order[2] then
			for side = 1, 2 do
				-- rearrange characters in selection order
				for k, v in ipairs(t_order[side]) do
					start.t_orderRemap[side][k] = v
				end
				-- update spr/anim data
				for member, v in ipairs(start.p[side].t_selected) do
					local done_anim = motif.vs_screen['p' .. side .. '_member' .. member .. '_done_anim'] or motif.vs_screen['p' .. side .. '_done_anim']
					if done_anim ~= -1 then
						if start.p[side].t_selTemp[member].anim ~= done_anim then
							start.p[side].t_selTemp[member].anim_data = start.f_animGet(v.ref, side, member, motif.vs_screen, '', '_done', false) or start.p[side].t_selTemp[member].anim_data
						end
					end
				end
				if t_orderSelect[side] then
					t_icon[side] = '_icon_done'
				end
			end
			counter = motif.vs_screen.time - motif.vs_screen.done_time
			done = true
		end
		counter = counter + 1
		--draw clearcolor
		clearColor(motif.versusbgdef.bgclearcolor[1], motif.versusbgdef.bgclearcolor[2], motif.versusbgdef.bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(motif.versusbgdef.bg, false)
		--draw portraits and order icons
		for side = 1, 2 do
			start.f_drawPortraits(main.f_remapTable(start.p[side].t_selTemp, start.t_orderRemap[side]), side, motif.vs_screen, '', false, t_icon[side])
		end
		--draw order values
		for side = 1, 2 do
			if t_orderSelect[side] then
				for i = 1, #start.p[side].t_selected do
					local prefix = '_icon'
					if i > #t_order[side] and #start.p[side].t_selected > #t_order[side] then
						prefix = '_empty_icon'
					end
					if motif.vs_screen['p' .. side .. '_value' .. prefix .. '_member' .. i .. '_data'] ~= nil then
						prefix = prefix .. '_member' .. i
					end
					main.f_animPosDraw(
						motif.vs_screen['p' .. side .. '_value' .. prefix .. '_data'],
						(i - 1) * motif.vs_screen['p' .. side .. '_value_icon_spacing'][1],
						(i - 1) * motif.vs_screen['p' .. side .. '_value_icon_spacing'][2],
						motif.vs_screen['p' .. side .. '_value' .. prefix .. '_facing']
					)
				end
			end
		end
		--draw names
		for side = 1, 2 do
			for k, v in ipairs(main.f_remapTable(start.p[side].t_selTemp, start.t_orderRemap[side])) do
				if k <= motif.vs_screen['p' .. side .. '_name_num'] or main.coop then
					t_txt_nameVS[side]:update({
						font =   motif.vs_screen['p' .. side .. '_name_font'][1],
						bank =   motif.vs_screen['p' .. side .. '_name_font'][2],
						align =  motif.vs_screen['p' .. side .. '_name_font'][3],
						text =   start.f_getName(v.ref, side),
						x =      motif.vs_screen['p' .. side .. '_name_pos'][1] + motif.vs_screen['p' .. side .. '_name_offset'][1] + (k - 1) * motif.vs_screen['p' .. side .. '_name_spacing'][1],
						y =      motif.vs_screen['p' .. side .. '_name_pos'][2] + motif.vs_screen['p' .. side .. '_name_offset'][2] + (k - 1) * motif.vs_screen['p' .. side .. '_name_spacing'][2],
						scaleX = motif.vs_screen['p' .. side .. '_name_scale'][1],
						scaleY = motif.vs_screen['p' .. side .. '_name_scale'][2],
						r =      motif.vs_screen['p' .. side .. '_name_font'][4],
						g =      motif.vs_screen['p' .. side .. '_name_font'][5],
						b =      motif.vs_screen['p' .. side .. '_name_font'][6],
						height = motif.vs_screen['p' .. side .. '_name_font'][7],
					})
					t_txt_nameVS[side]:draw()
				end
			end
		end
		--draw match counter
		if main.versusMatchNo then
			txt_matchNo:draw()
		end
		--draw timer
		if not done and motif.vs_screen.timer_count ~= -1 and timerActive and counter >= 0 then
			timerCount, timerActive = main.f_drawTimer(timerCount, motif.vs_screen, 'timer_', txt_timerVS)
		end
		-- hook
		hook.run("start.f_selectVersus")
		--draw layerno = 1 backgrounds
		bgDraw(motif.versusbgdef.bg, true)
		--draw fadein / fadeout
		for side = 1, 2 do
			if main.fadeType == 'fadein' and (
				counter >= motif.vs_screen.time
				or (not main.cpuSide[side] and main.f_input({side}, main.f_extractKeys(motif.vs_screen['p' .. side .. '_skip_key'])))
				or (done and main.f_input({side}, main.f_extractKeys(motif.vs_screen['p' .. side .. '_accept_key'])))
				) then
				main.f_fadeReset('fadeout', motif.vs_screen)
				break
			end
		end
		main.f_fadeAnim(motif.vs_screen)
		--frame transition
		if not escFlag and (esc() or main.f_input(main.t_players, {'m'})) then
			esc(false)
			main.f_fadeReset('fadeout', motif.vs_screen)
			escFlag = true
		end
		if not main.f_frameChange() then
			clearColor(motif.versusbgdef.bgclearcolor[1], motif.versusbgdef.bgclearcolor[2], motif.versusbgdef.bgclearcolor[3])
			break --skip last frame rendering
		end
		main.f_refresh()
	end
	esc(escFlag) --force Esc detection
	return not escFlag
end

--loading loop called after versus screen is finished
function start.f_selectLoading()
	clearAllSound()
	for side = 1, 2 do
		for _, v in ipairs(start.p[side].t_selected) do
			if not v.loading then
				selectChar(side, v.ref, v.pal)
				v.loading = true
			end
		end
	end
	--TODO: fix config.BackgroundLoading setting
	--if not config.BackgroundLoading then
		loadStart()
	--end
	-- calling refresh() during netplay data loading can lead to synchronization error
	--while motif.vs_screen.loading_data ~= nil and loading() and not network() do
	--	animDraw(motif.vs_screen.loading_data)
	--	animUpdate(motif.vs_screen.loading_data)
	--	refresh()
	--end
end

--;===========================================================
--; RESULT SCREEN
--;===========================================================
local txt_winscreen = main.f_createTextImg(motif.win_screen, 'wintext')
local txt_resultSurvival = main.f_createTextImg(motif.survival_results_screen, 'winstext')
local txt_resultTimeAttack = main.f_createTextImg(motif.time_attack_results_screen, 'winstext')

local function f_drawTextAtLayerNo(t, prefix, t_text, txt, layerNo)
	if t[prefix .. '_layerno'] ~= layerNo then
		return
	end
	for i = 1, #t_text do
		txt:update({
			text = t_text[i],
			y =    t[prefix .. '_offset'][2] + main.f_ySpacing(t, prefix) * (i - 1),
		})
		txt:draw()
	end
end

function start.f_lowestRankingData(data)
	if stats.modes == nil or stats.modes[gamemode()] == nil or stats.modes[gamemode()].ranking == nil or #stats.modes[gamemode()].ranking < motif.hiscore_info.window_visibleitems then
		if data == 'score' then
			return 0
		else --time
			return 99
		end
	end
	local ret = 0
	for k, v in ipairs(stats.modes[gamemode()].ranking) do
		if k == 1 or (data == 'score' and ret > v[data]) or (data == 'time' and ret < v[data]) then
			ret = v[data]
		end
	end
	return ret
end

-- start.t_resultData is a table storing functions used for setting variables
-- stored in start.t_result table, returning boolean depending on various
-- factors. It's used by start.f_resultInit function, depending on game mode.
-- Can be appended via external module, without conflicting with default scripts.
start.t_resultData = {}
start.t_resultData.arcade = function()
	if winnerteam() ~= 1 or matchno() < #start.t_roster or motif.win_screen.enabled == 0 then
		return false
	end
	if main.f_fileExists(start.f_getCharData(start.p[1].t_selected[1].ref).ending) then --not displayed if the team leader has an ending
		return false
	end
	start.t_result.prefix = 'wintext'
	start.t_result.resultText = main.f_extractText(main.resultsTable[start.t_result.prefix .. '_text'])
	start.t_result.txt = txt_winscreen
	start.t_result.bgdef = 'winbgdef'
	return true
end
start.t_resultData.teamcoop = start.t_resultData.arcade
start.t_resultData.netplayteamcoop = start.t_resultData.arcade
start.t_resultData.survival = function()
	if winnerteam() == 1 and (matchno() < #start.t_roster or (start.t_roster[matchno() + 1] ~= nil and start.t_roster[matchno() + 1][1] == -1)) or motif.survival_results_screen.enabled == 0 then
		return false
	end
	start.t_result.resultText = main.f_extractText(main.resultsTable[start.t_result.prefix .. '_text'], start.winCnt)
	start.t_result.txt = txt_resultSurvival
	start.t_result.bgdef = 'survivalresultsbgdef'
	if start.winCnt < main.resultsTable.roundstowin and matchno() < #start.t_roster then
		start.t_result.stateType = ''
		start.t_result.winBgm = false
	else
		start.t_result.stateType = '_win'
	end
	return true
end
start.t_resultData.survivalcoop = start.t_resultData.survival
start.t_resultData.netplaysurvivalcoop = start.t_resultData.survival
start.t_resultData.timeattack = function()
	if winnerteam() ~= 1 or matchno() < #start.t_roster or motif.time_attack_results_screen.enabled == 0 then
		return false
	end
	start.t_result.resultText = main.f_extractText(start.f_clearTimeText(main.resultsTable[start.t_result.prefix .. '_text'], timetotal() / 60))
	start.t_result.txt = txt_resultTimeAttack
	start.t_result.bgdef = 'timeattackresultsbgdef'
	if matchtime() / 60 >= start.f_lowestRankingData('time') then
		start.t_result.stateType = ''
		start.t_result.winBgm = false
	else
		start.t_result.stateType = '_win'
	end
	return true
end

start.resultInit = false
function start.f_resultInit()
	if start.resultInit then
		return start.t_result.active
	end
	start.resultInit = true
	start.t_result = {
		active = false,
		escFlag = false,
		prefix = 'winstext',
		stateType = '',
		winBgm = true,
		resultText = {},
		txt = nil,
		bgdef = 'winbgdef',
		counter = 0,
	}
	if main.resultsTable == nil then
		return false
	end
	start.t_result.counter = 0 - main.resultsTable.fadein_time
	local t = main.resultsTable
	start.t_result.overlay = main.f_createOverlay(t, 'overlay')
	if winnerteam() == 1 then
		start.winCnt = start.winCnt + 1
	else
		start.loseCnt = start.loseCnt + 1
	end
	if start.t_resultData[gamemode()] == nil or not start.t_resultData[gamemode()]() then
		return false
	end
	for i = 1, 2 do
		for k, v in ipairs(t['p' .. i .. start.t_result.stateType .. '_state']) do
			if charChangeState(i, v) then
				break
			end
		end
		player(i) --assign sys.debugWC to player i
		for j = 1, numpartner() do
			for _, v in ipairs(t['p' .. i .. '_teammate' .. start.t_result.stateType .. '_state']) do
				if charChangeState(j * 2 + i, v) then
					break
				end
			end
		end
	end
	if main.resultsTable.sounds_enabled == 0 then
		clearAllSound()
		toggleNoSound(true)
	end
	main.f_bgReset(motif[start.t_result.bgdef].bg)
	main.f_fadeReset('fadein', t)
	if start.t_result.winBgm and motif.music.results_bgm ~= '' then
		main.f_playBGM(false, motif.music.results_bgm, motif.music.results_bgm_loop, motif.music.results_bgm_volume, motif.music.results_bgm_loopstart, motif.music.results_bgm_loopend)
	elseif motif.music.results_lose_bgm ~= '' then
		main.f_playBGM(false, motif.music.results_lose_bgm, motif.music.results_lose_bgm_loop, motif.music.results_lose_bgm_volume, motif.music.results_lose_bgm_loopstart, motif.music.results_lose_bgm_loopend)
	end
	start.t_result.active = true
	return true
end

function start.f_result()
	if not start.f_resultInit() then
		return false
	end
	local t = main.resultsTable
	start.t_result.counter = start.t_result.counter + 1
	--draw overlay
	start.t_result.overlay:draw()
	--draw text at layerno = 0
	f_drawTextAtLayerNo(t, start.t_result.prefix, start.t_result.resultText, start.t_result.txt, 0)
	--draw layerno = 0 backgrounds
	bgDraw(motif[start.t_result.bgdef].bg, false)
	--draw text at layerno = 1
	f_drawTextAtLayerNo(t, start.t_result.prefix, start.t_result.resultText, start.t_result.txt, 1)
	-- hook
	hook.run("start.f_result")
	--draw layerno = 1 backgrounds
	bgDraw(motif[start.t_result.bgdef].bg, true)
	--draw text at layerno = 2
	f_drawTextAtLayerNo(t, start.t_result.prefix, start.t_result.resultText, start.t_result.txt, 2)
	--draw fadein / fadeout
	if main.fadeType == 'fadein' and (start.t_result.counter >= (t.show_time or t.pose_time) or main.f_input({1}, {'pal', 's'})) then
		main.f_fadeReset('fadeout', t)
	end
	main.f_fadeAnim(t)
	--frame transition
	if not start.t_result.escFlag and (esc() or main.f_input(main.t_players, {'m'})) then
		esc(false)
		main.f_fadeReset('fadeout', t)
		start.t_result.escFlag = true
	end
	if not main.f_frameChange() then
		start.t_result.active = false
		toggleNoSound(false)
		return false
	end
	return true
end

--;===========================================================
--; VICTORY SCREEN
--;===========================================================
local txt_winquote = main.f_createTextImg(motif.victory_screen, 'winquote')
local overlay_winquote = main.f_createOverlay(motif.victory_screen, 'overlay')
local t_txt_winquoteName = {}
for i = 1, 2 do
	table.insert(t_txt_winquoteName, main.f_createTextImg(motif.victory_screen, 'p' .. i .. '_name'))
end

function start.f_victoryOrder(side, paramSide, allow_ko, num)
	local allow_ko = allow_ko or 0
	local t = {}
	local t_matchList = {}
	local t_teamList = {}
	local playerNo = -1
	local selectNo = -1
	local memberNo = -1
	local foundLeader = false
	local f_appendTable = function(ref, memberNo)
		if #t >= num then return false end
		table.insert(t, {
			ref = ref,
			anim = motif.victory_screen['p' .. paramSide .. '_member' .. #t + 1 .. '_anim'] or motif.victory_screen['p' .. paramSide .. '_anim'],
			anim_data = start.f_animGet(ref, paramSide, #t + 1, motif.victory_screen, '', '', true, {9000, 1}),
			face2_data = start.f_animGet(ref, paramSide, #t + 1, motif.victory_screen, '_face2', '', true),
			slide_dist = {0, 0},
		})
		t_matchList[ref] = (t_matchList[ref] or 0) + 1
		t_teamList[memberNo] = true
		return true
	end
	--winner who made last hit takes priority
	local lastHitter = lasthitter(winnerteam())
	if lastHitter == 0 then
		lastHitter = side
	end
	if player(lastHitter) and winnerteam() == side then --assign sys.debugWC
		playerNo = lastHitter
		selectNo = selectno()
		memberNo = memberno()
		foundLeader = true
		f_appendTable(selectNo, memberNo)
	end
	--generate table out of remaining characters present in the last match
	for i = 1, math.max(#start.p[1].t_selected, #start.p[2].t_selected) * 2 do
		if player(i) and teamside() == side then --assign sys.debugWC
			if side ~= winnerteam() then --member of lose team
				if not foundLeader then
					playerNo = i
					selectNo = selectno()
					memberNo = memberno()
					foundLeader = true
				end
				if not f_appendTable(selectno(), memberNo) then break end
			elseif i ~= lastHitter then --member of win team (but skip winner who made last hit)
				if alive() and not foundLeader then --first not KOed member
					playerNo = i
					selectNo = selectno()
					memberNo = memberno()
					foundLeader = true
					if not f_appendTable(selectNo, memberNo) then break end
				elseif alive() or allow_ko == 1 then --other team members
					if not f_appendTable(selectno(), memberno()) then break end
				end
			end
		end
	end
	--append turns team mode characters not loaded during last match
	if #t < num and #t < #start.p[side].t_selected then
		for k, v in ipairs(main.f_remapTable(start.p[side].t_selected, start.t_orderRemap[side])) do
			if not t_teamList[k] and (allow_ko == 1 or k > memberNo) then
				f_appendTable(v.ref, k)
			end
		end
	end
	return playerNo, selectNo, t
end

start.victoryInit = false
function start.f_victoryInit()
	if start.victoryInit then
		return start.t_victory.active
	end
	start.victoryInit = true
	start.t_victory = {
		active = false,
		escFlag = false,
		winquote = '',
		textcnt = 0,
		textend = false,
		winnerNo = -1,
		winnerRef = -1,
		loserNo = -1,
		loserRef = -1,
		team1 = {},
		team2 = {},
		counter = 0 - motif.victory_screen.fadein_time,
	}
	if winnerteam() < 1 or not main.victoryScreen or motif.victory_screen.enabled == 0 then
		return false
	elseif gamemode('versus') or gamemode('netplayversus') then
		if motif.victory_screen.vs_enabled == 0 then
			return false
		end
	elseif winnerteam() == 2 and motif.victory_screen.cpu_enabled == 0 then
		return false
	end
	for side = 1, 2 do
		if winnerteam() == side then
			start.t_victory.winnerNo, start.t_victory.winnerRef, start.t_victory.team1 = start.f_victoryOrder(side, 1, motif.victory_screen.winner_teamko_enabled, motif.victory_screen.p1_num)
		else
			start.t_victory.loserNo, start.t_victory.loserRef, start.t_victory.team2 = start.f_victoryOrder(side, 2, true, motif.victory_screen.p2_num)
		end
	end
	if start.t_victory.winnerNo == -1 or start.t_victory.winnerRef == -1 then
		return false
	elseif start.f_getCharData(start.t_victory.winnerRef).victoryscreen == 0 then
		return false
	end
	for i = 1, 2 do
		for k, v in ipairs(motif.victory_screen['p' .. i .. '_state']) do
			if charChangeState(i, v) then
				break
			end
		end
		player(i) --assign sys.debugWC to player i
		for j = 1, numpartner() do
			for _, v in ipairs(motif.victory_screen['p' .. i .. '_teammate_state']) do
				if charChangeState(j * 2 + i, v) then
					break
				end
			end
		end
	end
	if motif.victory_screen.sounds_enabled == 0 then
		clearAllSound()
		toggleNoSound(true)
	end
	main.f_bgReset(motif.victorybgdef.bg)
	main.f_fadeReset('fadein', motif.victory_screen)
	if start.t_music.musicvictory[winnerteam()] == nil and motif.music.victory_bgm ~= '' then
		main.f_playBGM(false, motif.music.victory_bgm, motif.music.victory_bgm_loop, motif.music.victory_bgm_volume, motif.music.victory_bgm_loopstart, motif.music.victory_bgm_loopend)
	end
	start.f_resetTempData(motif.victory_screen, '')
	start.t_victory.winquote = getCharVictoryQuote(start.t_victory.winnerNo)
	if start.t_victory.winquote == '' then
		start.t_victory.winquote = motif.victory_screen.winquote_text
	end
	t_txt_winquoteName[1]:update({text = start.f_getName(start.t_victory.winnerRef)})
	t_txt_winquoteName[2]:update({text = start.f_getName(start.t_victory.loserRef)})
	start.t_victory.active = true
	return true
end

function start.f_victory()
	if not start.f_victoryInit() then
		return false
	end
	start.t_victory.counter = start.t_victory.counter + 1
	--draw overlay
	overlay_winquote:draw()
	--draw layerno = 0 backgrounds
	bgDraw(motif.victorybgdef.bg, false)
	--draw portraits (starting from losers)
	for side = 2, 1, -1 do
		start.f_drawPortraits(start.t_victory['team' .. side], side, motif.victory_screen, '', false)
	end
	--draw winner name
	t_txt_winquoteName[1]:draw()
	--draw loser name
	t_txt_winquoteName[2]:draw()
	--draw winquote
	if start.t_victory.counter + motif.victory_screen.fadein_time >= motif.victory_screen.winquote_displaytime then
		if not start.t_victory.textend then
			start.t_victory.textcnt = start.t_victory.textcnt + 1
		end
		start.t_victory.textend = main.f_textRender(
			txt_winquote,
			start.t_victory.winquote,
			start.t_victory.textcnt,
			motif.victory_screen.winquote_offset[1],
			motif.victory_screen.winquote_offset[2],
			motif.victory_screen.winquote_spacing[1],
			motif.victory_screen.winquote_spacing[2],
			main.font_def[motif.victory_screen.winquote_font[1] .. motif.victory_screen.winquote_font[7]],
			motif.victory_screen.winquote_delay,
			main.f_lineLength(
				motif.victory_screen.winquote_offset[1],
				motif.info.localcoord[1],
				motif.victory_screen.winquote_font[3],
				motif.victory_screen.winquote_window,
				motif.victory_screen.winquote_textwrap:match('[wl]')
			)
		)
	end
	-- hook
	hook.run("start.f_victory")
	--draw layerno = 1 backgrounds
	bgDraw(motif.victorybgdef.bg, true)
	--draw fadein / fadeout
	if main.fadeType == 'fadein' and ((start.t_victory.textend and start.t_victory.counter - start.t_victory.textcnt >= motif.victory_screen.time) or main.f_input(main.t_players, {'pal', 's'})) then
		main.f_fadeReset('fadeout', motif.victory_screen)
	end
	main.f_fadeAnim(motif.victory_screen)
	--frame transition
	if not start.t_victory.escFlag and (esc() or main.f_input(main.t_players, {'m'})) then
		esc(false)
		main.f_fadeReset('fadeout', motif.victory_screen)
		start.t_victory.escFlag = true
	end
	if not main.f_frameChange() then
		start.t_victory.active = false
		toggleNoSound(false)
		return false
	end
	return true
end

--;===========================================================
--; CONTINUE SCREEN
--;===========================================================
local txt_credits = main.f_createTextImg(motif.continue_screen, 'credits')
local txt_continue = main.f_createTextImg(motif.continue_screen, 'continue', {x = motif.continue_screen.pos[1], y = motif.continue_screen.pos[2]})
local txt_yes = text:create({})
local txt_no = text:create({})
local overlay_continue = main.f_createOverlay(motif.continue_screen, 'overlay')

start.t_continueCounts = {}
for k, v in pairs(motif.continue_screen) do
	local n = k:match('^counter_([0-9]+)_skiptime$')
	if n ~= nil then
		start.t_continueCounts[tonumber(n)] = {
			skiptime = v,
			snd = {motif.continue_screen.counter_default_snd[1], motif.continue_screen.counter_default_snd[2]}
		}
		if motif.continue_screen['counter_' .. n .. '_snd'] ~= nil then
			start.t_continueCounts[tonumber(n)].snd = motif.continue_screen['counter_' .. n .. '_snd']
		end
	end
end

start.continueInit = false
function start.f_continueInit()
	if start.continueInit then
		return start.t_continue.active
	end
	setContinue(false)
	start.continueInit = true
	start.t_continue = {
		active = false,
		escFlag = false,
		continue = false,
		flag = true,
		selected = false,
		counter = 0,-- - motif.continue_screen.fadein_time
	}
	if motif.continue_screen.enabled == 0 or not main.continueScreen or winnerteam() == 1 or (motif.continue_screen.legacymode_enabled == 1 and main.credits == 0) or start.challenger > 0 then
		return false
	end
	if motif.continue_screen.legacymode_enabled == 0 then
		start.t_continue.t_btnSkip = {'s'}
	else
		start.t_continue.t_btnSkip = {'pal', 's'}
	end
	if motif.continue_screen.sounds_enabled == 0 then
		clearAllSound()
		toggleNoSound(true)
	end
	if motif.music.continue_bgm ~= '' then
		main.f_playBGM(false, motif.music.continue_bgm, motif.music.continue_bgm_loop, motif.music.continue_bgm_volume, motif.music.continue_bgm_loopstart, motif.music.continue_bgm_loopend)
	end
	main.f_bgReset(motif.continuebgdef.bg)
	main.f_fadeReset('fadein', motif.continue_screen)
	animReset(motif.continue_screen.counter_data)
	animUpdate(motif.continue_screen.counter_data)
	for i = 1, 2 do
		for _, v in ipairs(start.p[i].t_selCmd) do
			v.selectState = 0
		end
		for _, v in ipairs(motif.continue_screen['p' .. i .. '_state']) do
			if charChangeState(i, v) then
				break
			end
		end
		player(i) --assign sys.debugWC to player i
		for j = 1, numpartner() do
			for _, v in ipairs(motif.continue_screen['p' .. i .. '_teammate_state']) do
				if charChangeState(j * 2 + i, v) then
					break
				end
			end
		end
	end
	start.t_continue.active = true
	return true
end

function start.f_continue()
	if not start.f_continueInit() then
		return false
	end
	--draw overlay
	overlay_continue:draw()
	--draw layerno = 0 backgrounds
	bgDraw(motif.continuebgdef.bg, false)
	if motif.continue_screen.legacymode_enabled == 0 then --extended continue screen parameters
		if not start.t_continue.selected then
			if start.t_continue.counter < motif.continue_screen.counter_end_skiptime then
				--start pressed, continue = yes
				if (main.credits == -1 or main.credits > 0) and main.f_input({1}, {'s'}) then
					start.t_continue.continue = true
					sndPlay(motif.files.snd_data, motif.continue_screen.done_snd[1], motif.continue_screen.done_snd[2])
					for i = 1, 2 do
						for _, v in ipairs(motif.continue_screen['p' .. i .. '_yes_state']) do
							if charChangeState(i, v) then
								break
							end
						end
						player(i) --assign sys.debugWC to player i
						for j = 1, numpartner() do
							for _, v in ipairs(motif.continue_screen['p' .. i .. '_teammate_yes_state']) do
								if charChangeState(j * 2 + i, v) then
									break
								end
							end
						end
					end
					start.t_continue.selected = true
					if main.credits > 0 then
						main.credits = main.credits - 1
					end
				--counter anim time skip on button press
				elseif main.f_input({1}, {'pal'}) and start.t_continue.counter >= motif.continue_screen.counter_starttime + motif.continue_screen.counter_skipstart then
					for _, v in main.f_sortKeys(start.t_continueCounts, function(t, a, b) return a > b end) do --iterate over the table in descending order
						if start.t_continue.counter < v.skiptime then
							while start.t_continue.counter < v.skiptime do
								start.t_continue.counter = start.t_continue.counter + 1
								animUpdate(motif.continue_screen.counter_data)
							end
							break
						end
					end
				end
				--counter anim snd play
				for _, v in main.f_sortKeys(start.t_continueCounts, function(t, a, b) return a > b end) do --iterate over the table in descending order
					if start.t_continue.counter == v.skiptime then
						sndPlay(motif.files.snd_data, v.snd[1], v.snd[2])
						break
					end
				end
			elseif start.t_continue.counter == motif.continue_screen.counter_end_skiptime then
				if motif.music.continue_end_bgm ~= '' then
					main.f_playBGM(false, motif.music.continue_end_bgm, motif.music.continue_end_bgm_loop, motif.music.continue_end_bgm_volume, motif.music.continue_end_bgm_loopstart, motif.music.continue_end_bgm_loopend)
				end
				sndPlay(motif.files.snd_data, motif.continue_screen.counter_end_snd[1], motif.continue_screen.counter_end_snd[2])
				for i = 1, 2 do
					for _, v in ipairs(motif.continue_screen['p' .. i .. '_no_state']) do
						if charChangeState(i, v) then
							break
						end
					end
					player(i) --assign sys.debugWC to player i
					for j = 1, numpartner() do
						for _, v in ipairs(motif.continue_screen['p' .. i .. '_teammate_no_state']) do
							if charChangeState(j * 2 + i, v) then
								break
							end
						end
					end
				end
			end
			--draw counter
			animUpdate(motif.continue_screen.counter_data)
			animDraw(motif.continue_screen.counter_data)
		end
		--draw credits text
		if main.credits ~= -1 and start.t_continue.counter >= motif.continue_screen.counter_skipstart then --show when counter starts counting down
			txt_credits:update({text = main.f_extractText(motif.continue_screen.credits_text, main.credits)[1]})
			txt_credits:draw()
		end
		start.t_continue.counter = start.t_continue.counter + 1
	else --legacy mugen continue screen parameters
		if not start.t_continue.selected and main.f_input({1}, {'$F', '$B'}) then
			sndPlay(motif.files.snd_data, motif.continue_screen.move_snd[1], motif.continue_screen.move_snd[2])
			if start.t_continue.flag then
				start.t_continue.flag = false
			else
				start.t_continue.flag = true
			end
		elseif not start.t_continue.selected and main.f_input({1}, {'pal', 's'}) then
			start.t_continue.continue = start.t_continue.flag
			if start.t_continue.continue then
				sndPlay(motif.files.snd_data, motif.continue_screen.done_snd[1], motif.continue_screen.done_snd[2])
				for i = 1, 2 do
					for _, v in ipairs(motif.continue_screen['p' .. i .. '_yes_state']) do
						if charChangeState(i, v) then
							break
						end
					end
					player(i) --assign sys.debugWC to player i
					for j = 1, numpartner() do
						for _, v in ipairs(motif.continue_screen['p' .. i .. '_teammate_yes_state']) do
							if charChangeState(j * 2 + i, v) then
								break
							end
						end
					end
				end
				start.t_continue.selected = true
				main.credits = main.credits - 1
			else
				sndPlay(motif.files.snd_data, motif.continue_screen.cancel_snd[1], motif.continue_screen.cancel_snd[2])
				for i = 1, 2 do
					for _, v in ipairs(motif.continue_screen['p' .. i .. '_no_state']) do
						if charChangeState(i, v) then
							break
						end
					end
					player(i) --assign sys.debugWC to player i
					for j = 1, numpartner() do
						for _, v in ipairs(motif.continue_screen['p' .. i .. '_teammate_no_state']) do
							if charChangeState(j * 2 + i, v) then
								break
							end
						end
					end
				end
			end
			start.t_continue.counter = motif.continue_screen.counter_endtime + 1
		end
		txt_continue:draw()
		--draw yes/no text
		for i = 1, 2 do
			local txt = ''
			local var = ''
			if i == 1 then
				txt = txt_yes
				if start.t_continue.flag then
					var = 'yes_active'
				else
					var = 'yes'
				end
			else
				txt = txt_no
				if start.t_continue.flag then
					var = 'no'
				else
					var = 'no_active'
				end
			end
			txt:update({
				font =   motif.continue_screen[var .. '_font'][1],
				bank =   motif.continue_screen[var .. '_font'][2],
				align =  motif.continue_screen[var .. '_font'][3],
				text =   motif.continue_screen[var .. '_text'],
				x =      motif.continue_screen.pos[1] + motif.continue_screen[var .. '_offset'][1],
				y =      motif.continue_screen.pos[2] + motif.continue_screen[var .. '_offset'][2],
				scaleX = motif.continue_screen[var .. '_scale'][1],
				scaleY = motif.continue_screen[var .. '_scale'][2],
				r =      motif.continue_screen[var .. '_font'][4],
				g =      motif.continue_screen[var .. '_font'][5],
				b =      motif.continue_screen[var .. '_font'][6],
				height = motif.continue_screen[var .. '_font'][7],
			})
			txt:draw()
		end
		--draw credits text
		if main.credits ~= -1 then
			txt_credits:update({text = main.f_extractText(motif.continue_screen.credits_text, main.credits)[1]})
			txt_credits:draw()
		end
	end
	-- hook
	hook.run("start.f_continue")
	--draw layerno = 1 backgrounds
	bgDraw(motif.continuebgdef.bg, true)
	--draw fadein / fadeout
	if main.fadeType == 'fadein' and (start.t_continue.counter > motif.continue_screen.counter_endtime or start.t_continue.continue or (main.f_input({1}, start.t_continue.t_btnSkip) and (motif.continue_screen.legacymode_enabled == 1 or start.t_continue.counter >= motif.continue_screen.counter_end_skiptime))) then
		main.f_fadeReset('fadeout', motif.continue_screen)
	end
	main.f_fadeAnim(motif.continue_screen)
	--frame transition
	if not start.t_continue.escFlag and (esc() or main.f_input(main.t_players, {'m'})) then
		esc(false)
		main.f_fadeReset('fadeout', motif.continue_screen)
		start.t_continue.escFlag = true
	end
	if not main.f_frameChange() then
		start.t_continue.active = false
		setContinue(start.t_continue.continue)
		toggleNoSound(false)
		return false
	end
	return true
end

--;===========================================================
--; HISCORE
--;===========================================================
local overlay_hiscore = main.f_createOverlay(motif.hiscore_info, 'overlay')
for _, v in ipairs({'title', 'title_rank', 'title_data', 'title_name', 'title_face', 'item_rank', 'item_rank_active', 'item_rank_active2', 'item_data', 'item_data_active', 'item_data_active2', 'item_name', 'item_name_active', 'item_name_active2', 'timer'}) do
	start['txt_hiscore_' .. v] = main.f_createTextImg(motif.hiscore_info, v, {x = motif.hiscore_info.pos[1], y = motif.hiscore_info.pos[2]})
end

start.hiscoreInit = false
function start.f_hiscoreInit(gameMode, playMusic, input)
	if start.hiscoreInit then
		return start.t_hiscore.active
	end
	start.hiscoreInit = true
	start.t_hiscore = {
		active = false,
		escFlag = false,
		rankActiveCount = 0,
		dataActiveCount = 0,
		nameActiveCount = 0,
		rankActiveType = '_active',
		dataActiveType = '_active',
		nameActiveType = '_active',
		faces = {},
		letters = {},
		input = input,
		timer = 0,
		counter = 0 - motif.hiscore_info.fadein_time,
	}
	if input then
		table.insert(start.t_hiscore.letters, 1)
	end
	if motif.hiscore_info.enabled == 0 or stats.modes == nil or stats.modes[gameMode] == nil or stats.modes[gameMode].ranking == nil then
		return false
	end
	main.f_cmdBufReset()
	clearColor(motif.hiscorebgdef.bgclearcolor[1], motif.hiscorebgdef.bgclearcolor[2], motif.hiscorebgdef.bgclearcolor[3])
	if playMusic and motif.music.hiscore_bgm ~= '' then
		main.f_playBGM(false, motif.music.hiscore_bgm, motif.music.hiscore_bgm_loop, motif.music.hiscore_bgm_volume, motif.music.hiscore_bgm_loopstart, motif.music.hiscore_bgm_loopend)
	end
	main.f_bgReset(motif.hiscorebgdef.bg)
	main.f_fadeReset('fadein', motif.hiscore_info)
	for i = 1, motif.hiscore_info.window_visibleitems do
		table.insert(start.t_hiscore.faces, {})
		local t = stats.modes[gameMode].ranking
		if t[i] == nil then
			break
		end
		for _, def in ipairs(t[i].chars) do
			if main.t_charDef[def] ~= nil then
				for _, v in pairs({
					{motif.hiscore_info.item_face_anim, -1},
					motif.hiscore_info.item_face_spr,
				}) do
					if v[1] ~= -1 then
						local a = animGetPreloadedData('char', main.t_charDef[def], v[1], v[2], true)
						if a ~= nil then
							animSetScale(
								a,
								motif.hiscore_info.item_face_scale[1] * start.f_getCharData(start.f_getCharRef(def)).portrait_scale / (main.SP_Viewport43[3] / main.SP_Localcoord[1]),
								motif.hiscore_info.item_face_scale[2] * start.f_getCharData(start.f_getCharRef(def)).portrait_scale / (main.SP_Viewport43[3] / main.SP_Localcoord[1]),
								false
							)
							animUpdate(a)
							table.insert(start.t_hiscore.faces[#start.t_hiscore.faces], {anim_data = a, chardata = true})
							break
						end
					end
				end
			else
				table.insert(start.t_hiscore.faces[#start.t_hiscore.faces], {anim_data = motif.hiscore_info.item_face_unknown_data, chardata = false})
			end
		end
	end
	start.t_hiscore.active = true
	return true
end

function start.f_hiscore(t, playMusic, place, infinite)
	if not start.f_hiscoreInit(t.mode, playMusic, place > 0) then
		return false
	end
	start.t_hiscore.counter = start.t_hiscore.counter + 1
	--draw layerno = 0 backgrounds
	bgDraw(motif.hiscorebgdef.bg, false)
	--draw overlay
	overlay_hiscore:draw()
	--draw title
	start.txt_hiscore_title:update({text = main.f_itemnameUpper(motif.hiscore_info.title_text:gsub('%%s', t.title), motif.hiscore_info.title_uppercase == 1)})
	start.txt_hiscore_title:draw()
	--draw hiscore
	local dataActiveType = ''
	local t_ranking = stats.modes[t.mode].ranking
	--draw portraits subtitle
	start.txt_hiscore_title_face:draw()
	--draw portraits
	for i, subt in ipairs(start.t_hiscore.faces) do
		for j, v in ipairs(subt) do
			if j > motif.hiscore_info.item_face_num then
				break
			end
			main.f_animPosDraw(
				motif.hiscore_info.item_face_bg_data,
				motif.hiscore_info.pos[1] + motif.hiscore_info.item_offset[1] + motif.hiscore_info.item_face_offset[1] + (i - 1) * motif.hiscore_info.item_spacing[1] + (j - 1) * motif.hiscore_info.item_face_spacing[1],
				motif.hiscore_info.pos[2] + motif.hiscore_info.item_offset[2] + motif.hiscore_info.item_face_offset[2] + (i - 1) * (motif.hiscore_info.item_spacing[2] + motif.hiscore_info.item_face_spacing[2]),
				motif.hiscore_info.item_face_facing,
				false
			)
			main.f_animPosDraw(
				v.anim_data,
				motif.hiscore_info.pos[1] + motif.hiscore_info.item_offset[1] + motif.hiscore_info.item_face_offset[1] + (i - 1) * motif.hiscore_info.item_spacing[1] + (j - 1) * motif.hiscore_info.item_face_spacing[1],
				motif.hiscore_info.pos[2] + motif.hiscore_info.item_offset[2] + motif.hiscore_info.item_face_offset[2] + (i - 1) * (motif.hiscore_info.item_spacing[2] + motif.hiscore_info.item_face_spacing[2]),
				motif.hiscore_info.item_face_facing,
				v.chardata
			)
		end
	end
	for _, v in ipairs({'rank', 'data', 'name'}) do
		--draw subtitle
		start['txt_hiscore_title_' .. v]:draw()
		if start.t_hiscore[v .. 'ActiveCount'] < motif.hiscore_info['item_' .. v .. '_active_switchtime'] then --delay change
			start.t_hiscore[v .. 'ActiveCount'] = start.t_hiscore[v .. 'ActiveCount'] + 1
		else
			if start.t_hiscore[v .. 'ActiveType'] == '_active' then
				start.t_hiscore[v .. 'ActiveType'] = '_active2'
			else
				start.t_hiscore[v .. 'ActiveType'] = '_active'
			end
			start.t_hiscore[v .. 'ActiveCount'] = 0
		end
		for i = 1, motif.hiscore_info.window_visibleitems do
			if t_ranking[i] == nil then
				break
			end
			if i == place then
				dataActiveType = start.t_hiscore[v .. 'ActiveType']
				if v == 'name' and start.t_hiscore.input then
					local t_letters = start.t_hiscore.letters
					if main.f_input(main.t_players, {'$B'}) then
						sndPlay(motif.files.snd_data, motif.hiscore_info.move_snd[1], motif.hiscore_info.move_snd[2])
						t_letters[#t_letters] = t_letters[#t_letters] - 1
						if t_letters[#t_letters] <= 0 then
							t_letters[#t_letters] = #motif.hiscore_info.glyphs
						end
					elseif main.f_input(main.t_players, {'$F'}) then
						sndPlay(motif.files.snd_data, motif.hiscore_info.move_snd[1], motif.hiscore_info.move_snd[2])
						t_letters[#t_letters] = t_letters[#t_letters] + 1
						if t_letters[#t_letters] > #motif.hiscore_info.glyphs then
							t_letters[#t_letters] = 1
						end
					elseif main.f_input(main.t_players, {'pal'}) then
						if motif.hiscore_info.glyphs[t_letters[#t_letters]] == '<' then
							sndPlay(motif.files.snd_data, motif.hiscore_info.cancel_snd[1], motif.hiscore_info.cancel_snd[2])
							if #t_letters > 1 then
								table.remove(t_letters, #t_letters)
							else
								t_letters[1] = 1
							end
						elseif #t_letters < (tonumber(motif.hiscore_info.item_name_text:match('%%([0-9]+)s')) or 3) then
							sndPlay(motif.files.snd_data, motif.hiscore_info.done_snd[1], motif.hiscore_info.done_snd[2])
							table.insert(t_letters, 1)
						else
							sndPlay(motif.files.snd_data, motif.hiscore_info.done_snd[1], motif.hiscore_info.done_snd[2])
							start.t_hiscore.counter = motif.hiscore_info.time - 30
							start.t_hiscore.input = false
						end
						main.f_cmdBufReset()
					end
					local name = ''
					for _, v in ipairs(t_letters) do
						name = name .. tostring(motif.hiscore_info.glyphs[v]):gsub('>', ' ')
					end
					t_ranking[i].name = name
				end
			else
				dataActiveType = ''
			end
			--draw rank
			local text = ''
			if v == 'rank' then
				text = (motif.hiscore_info['item_' .. v .. '_' .. i .. '_text'] or motif.hiscore_info['item_' .. v .. '_text']):gsub('%%s', tostring(i))
			--draw text
			elseif v == 'data' then
				local subText = t_ranking[i][t.data]
				text = (motif.hiscore_info['item_' .. v .. '_' .. t.data .. '_' .. i .. '_text'] or motif.hiscore_info['item_' .. v .. '_' .. t.data .. '_text'] or motif.hiscore_info['item_' .. v .. '_' .. i .. '_text'] or motif.hiscore_info['item_' .. v .. '_text'])
				if t.data == 'score' then
					local length = tonumber(text:match('%%([0-9]+)s'))
					while string.len(tostring(subText)) < (length or 0) do
						subText = 0 .. tostring(subText)
					end
					text = text:gsub('%%([0-9]*)s', tostring(subText))
				elseif t.data == 'time' then
					text = start.f_clearTimeText(text, tostring(subText))
				else --if t.data == 'win' then
					text = text:gsub('%%s', tostring(subText))
				end
			--draw name
			elseif v == 'name' and t_ranking[i].name ~= '' then
				text = (motif.hiscore_info['item_' .. v .. '_' .. i .. '_text'] or motif.hiscore_info['item_' .. v .. '_text']):gsub('%%([0-9]*)s', main.f_itemnameUpper(t_ranking[i].name, motif.hiscore_info.item_name_uppercase == 1))
			end
			local font_def = main.font_def[motif.hiscore_info['item_' .. v .. dataActiveType .. '_font'][1] .. motif.hiscore_info['item_' .. v .. dataActiveType .. '_font'][7]]
			start['txt_hiscore_item_' .. v .. dataActiveType]:update({
				text = text,
				x = motif.hiscore_info.pos[1] + motif.hiscore_info.item_offset[1] + motif.hiscore_info['item_' .. v .. '_offset'][1] + (motif.hiscore_info.item_spacing[1] + motif.hiscore_info['item_' .. v .. '_spacing'][1]) * (i - 1),
				y = motif.hiscore_info.pos[2] + motif.hiscore_info.item_offset[2] + motif.hiscore_info['item_' .. v .. '_offset'][2] + main.f_round((font_def.Size[2] + font_def.Spacing[2]) * start['txt_hiscore_item_' .. v .. dataActiveType].scaleY + (motif.hiscore_info.item_spacing[2] + motif.hiscore_info['item_' .. v .. '_spacing'][2])) * (i - 1),
			})
			start['txt_hiscore_item_' .. v .. dataActiveType]:draw()
		end
	end
	--draw timer
	if motif.hiscore_info.timer_count ~= -1 and start.t_hiscore.input and start.t_hiscore.counter >= 0 then
		start.t_hiscore.timer, start.t_hiscore.input = main.f_drawTimer(start.t_hiscore.timer, motif.hiscore_info, 'timer_', start.txt_hiscore_timer)
		if not start.t_hiscore.input then
			sndPlay(motif.files.snd_data, motif.hiscore_info.done_snd[1], motif.hiscore_info.done_snd[2])
		end
	end
	--credits
	if main.credits ~= -1 and getKey(motif.attract_mode.credits_key) then
		sndPlay(motif.files.snd_data, motif.attract_mode.credits_snd[1], motif.attract_mode.credits_snd[2])
		main.credits = main.credits + 1
		resetKey()
	end
	-- hook
	hook.run("start.f_hiscore")
	--draw layerno = 1 backgrounds
	bgDraw(motif.hiscorebgdef.bg, true)
	--draw fadein / fadeout
	if main.fadeType == 'fadein' and not main.fadeActive and not start.t_hiscore.input and (((not infinite and start.t_hiscore.counter >= motif.hiscore_info.time) or (motif.attract_mode.enabled == 0 and main.f_input(main.t_players, {'pal', 's'}))) or (motif.attract_mode.enabled == 1 and main.credits > 0)) then
		main.f_fadeReset('fadeout', motif.hiscore_info)
	end
	main.f_fadeAnim(motif.hiscore_info)
	--frame transition
	if not start.t_hiscore.escFlag and (esc() or main.f_input(main.t_players, {'m'})) then
		esc(false)
		main.f_fadeReset('fadeout', motif.continue_screen)
		start.t_hiscore.escFlag = true
	end
	if not main.f_frameChange() then
		start.t_hiscore.active = false
		return false
	end
	return true
end

--;===========================================================
--; CHALLENGER
--;===========================================================
local txt_challenger = main.f_createTextImg(motif.challenger_info, 'text')
local overlay_challenger = main.f_createOverlay(motif.challenger_info, 'overlay')

start.challengerInit = false
function start.f_challengerInit()
	if start.challengerInit then
		return start.t_challenger.active
	end
	start.challengerInit = true
	start.t_challenger = {
		active = false,
		counter = 0 - motif.challenger_info.fadein_time,
	}
	if motif.challenger_info.enabled == 0 then
		return false
	end
	if motif.attract_mode.enabled == 1 and main.credits > 0 then
		main.credits = main.credits - 1
	end
	main.f_playBGM(true)
	main.f_bgReset(motif.challengerbgdef.bg)
	main.f_fadeReset('fadein', motif.challenger_info)
	animReset(motif.challenger_info.bg_data)
	animUpdate(motif.challenger_info.bg_data)
	start.t_challenger.active = true
	return true
end

function start.f_challenger()
	if not start.f_challengerInit() then
		return false
	end
	if start.t_challenger.counter == motif.challenger_info.pause_time then
		togglePause(true)
	end
	if start.t_challenger.counter == motif.challenger_info.snd_time then
		sndPlay(motif.files.snd_data, motif.challenger_info.snd[1], motif.challenger_info.snd[2])
	end
	--draw overlay
	overlay_challenger:draw()
	--draw text at layerno = 0
	if start.t_challenger.counter >= motif.challenger_info.text_displaytime then
		f_drawTextAtLayerNo(motif.challenger_info, 'text', {motif.challenger_info.text_text}, txt_challenger, 0)
	end
	--draw layerno = 0 backgrounds
	bgDraw(motif.challengerbgdef.bg, false)
	--draw bg
	if start.t_challenger.counter >= motif.challenger_info.bg_displaytime then
		animUpdate(motif.challenger_info.bg_data)
		animDraw(motif.challenger_info.bg_data)
	end
	--draw text at layerno = 1
	if start.t_challenger.counter >= motif.challenger_info.text_displaytime then
		f_drawTextAtLayerNo(motif.challenger_info, 'text', {motif.challenger_info.text_text}, txt_challenger, 1)
	end
	-- hook
	hook.run("start.f_challenger")
	--draw layerno = 1 backgrounds
	bgDraw(motif.challengerbgdef.bg, true)
	--draw text at layerno = 2
	if start.t_challenger.counter >= motif.challenger_info.text_displaytime then
		f_drawTextAtLayerNo(motif.challenger_info, 'text', {motif.challenger_info.text_text}, txt_challenger, 2)
	end
	--draw fadein / fadeout
	if main.fadeType == 'fadein' and start.t_challenger.counter >= motif.challenger_info.time then
		main.f_fadeReset('fadeout', motif.challenger_info)
	end
	main.f_fadeAnim(motif.challenger_info)
	--frame transition
	if not main.fadeActive and main.fadeType == 'fadeout' then
		clearColor(motif.selectbgdef.bgclearcolor[1], motif.selectbgdef.bgclearcolor[2], motif.selectbgdef.bgclearcolor[3])
		start.challengerInit = false
		endMatch()
	end
	start.t_challenger.counter = start.t_challenger.counter + 1
end

--;===========================================================
--; STAGE MUSIC
--;===========================================================

-- Function checking conditions for music triggering, called each frame during match by loop() function in global.lua
function start.f_stageMusic()
	if main.flags['-nomusic'] ~= nil or gamemode('') then
		return
	end
	if gamemode('demo') and motif.demo_mode.fight_playbgm == 0 then
		return
	end
	-- bgmusic / bgmusic.roundX / bgmusic.final
	if roundstart() then
		-- only if the round is not restarted
		if start.bgmround ~= roundno() then
			start.bgmround = roundno()
			local roundNo = start.bgmround
			-- lookup first valid track if life bgm was triggered in the previous round
			if start.bgmstate == 1 then
				for i = roundNo, 1, -1 do
					if start.t_music.music[i] ~= nil then
						roundNo = i
						break
					end
				end
			end
			-- final round music assigned
			if roundNo > 1 and roundtype() == 3 and start.t_music.musicfinal.bgmusic ~= nil then
				main.f_playBGM(false, start.t_music.musicfinal.bgmusic, 1, start.t_music.musicfinal.bgmvolume, start.t_music.musicfinal.bgmloopstart, start.t_music.musicfinal.bgmloopend)
			-- music exists for this round
			elseif start.t_music.music[roundNo] ~= nil then
				-- interrupt same track playing only on round 1 of first match (skips continuous survival etc.)
				main.f_playBGM(matchno() == 1 and roundNo == 1, start.t_music.music[roundNo].bgmusic, 1, start.t_music.music[roundNo].bgmvolume, start.t_music.music[roundNo].bgmloopstart, start.t_music.music[roundNo].bgmloopend)
			-- stop versus screen track or life bgm even if stage music is not assigned
			elseif start.bgmround == 1 or start.bgmstate == 1 then
				main.f_playBGM(true)
			end
		end
		start.bgmstate = 0
	-- bgmusic.life
	elseif start.t_music.musiclife.bgmusic ~= nil and start.bgmstate == 0 and roundstate() == 2 then
		for i = 1, 2 do
			player(i) --assign sys.debugWC to player i
			-- continue only if p1/p2 life meets life ratio criteria
			if life() / lifemax() * 100 <= start.t_music.bgmratio_life then
				local ok = true
				for j = 1, numpartner() do
					player(j * 2 + i) --assign sys.debugWC to member j
					-- skip music playback if any of the team members doesn't meet life ratio criteria
					if life() / lifemax() * 100 > start.t_music.bgmratio_life then
						ok = false
						break
					end
				end
				if ok then
					if start.t_music.bgmtrigger_life == 1 or roundtype() >= 2 then
						main.f_playBGM(true, start.t_music.musiclife.bgmusic, 1, start.t_music.musiclife.bgmvolume, start.t_music.musiclife.bgmloopstart, start.t_music.musiclife.bgmloopend)
						start.bgmstate = 1
						break
					end
				end
			end
		end
	-- bgmusic.victory
	elseif #start.t_music.musicvictory > 0 and start.bgmstate ~= -1 and roundstate() == 3 then
		for i = 1, 2 do
			if start.t_music.musicvictory[i] ~= nil and player(i) and win() and (roundtype() == 1 or roundtype() == 3) then --assign sys.debugWC to player i
				main.f_playBGM(true, start.t_music.musicvictory[i].bgmusic, 1, start.t_music.musicvictory[i].bgmvolume, start.t_music.musicvictory[i].bgmloopstart, start.t_music.musicvictory[i].bgmloopend)
				start.bgmstate = -1
				break
			end
		end
	end
end

--;===========================================================
--; DIALOGUE
--;===========================================================
for i = 1, 2 do
	start['txt_dialogue_p' .. i .. '_name'] = main.f_createTextImg(motif.dialogue_info, 'p' .. i .. '_name')
	start['txt_dialogue_p' .. i .. '_text'] = main.f_createTextImg(motif.dialogue_info, 'p' .. i .. '_text')
end

start.dialogueInit = false
function start.f_dialogueInit()
	if start.dialogueInit then
		return start.t_dialogue.active
	end
	start.dialogueInit = true
	start.t_dialogue = {
		active = false,
		switch = false,
		player = -1,
		parsed = {},
		face = {
			{spr = motif.dialogue_info.p1_face_spr, pn = -1},
			{spr = motif.dialogue_info.p2_face_spr, pn = -1},
		},
		textNum = 1,
		activeSide = -1,
		endtime = -1,
		wait = 0,
		checktoken = 0,
		counter = 0,
	}
	if motif.dialogue_info.enabled == 0 then
		start.dialogueInit = false
		dialogueReset()
		return false
	end
	toggleDialogueBars(true)
	start.f_dialogueParse()
	for side = 1, 2 do
		animReset(motif.dialogue_info['p' .. side .. '_bg_data'])
		animUpdate(motif.dialogue_info['p' .. side .. '_bg_data'])
		animReset(motif.dialogue_info['p' .. side .. '_active_data'])
		animUpdate(motif.dialogue_info['p' .. side .. '_active_data'])
	end
	player(start.t_dialogue.player)
	start.txt_dialogue_p1_name:update({text = name()})
	enemy(0)
	start.txt_dialogue_p2_name:update({text = name()})
	start.t_dialogue.active = true
	return true
end

function start.f_dialogueRedirection(str)
	local redirection, val = str:match('([A-Za-z]+)%(?([^%)]*)%)?$')
	redirection = redirection:lower()
	player(start.t_dialogue.player)
	if redirection == 'self' then
		return start.t_dialogue.player
	elseif redirection == 'playerno' then
		if player(tonumber(val)) then
			return tonumber(val)
		end
	elseif redirection == 'partner' then
		if val == '' then val = 0 end
		if partner(tonumber(val)) then
			return playerno()
		end
	elseif redirection == 'enemy' then
		if val == '' then val = 0 end
		if enemy(tonumber(val)) then
			return playerno()
		end
	elseif redirection == 'enemyname' then
		for i = 1, numenemy() do
			player(start.t_dialogue.player)
			if enemy(i - 1) and name():lower() == val:lower() then
				return playerno()
			end
		end
	elseif redirection == 'partnername' then
		for i = 1, numpartner() do
			player(start.t_dialogue.player)
			if partner(i - 1) and name():lower() == val:lower() then
				return playerno()
			end
		end
	end
	return -1
end

function start.f_dialogueParse()
	local t_text, pn = getCharDialogue()
	start.t_dialogue.player = pn
	start.t_dialogue.face[1].pn = start.t_dialogue.player
	if main.f_playerSide(start.t_dialogue.player) == 1 then
		start.t_dialogue.face[2].pn = 2
	else
		start.t_dialogue.face[2].pn = 1
	end
	for _, v in ipairs(t_text) do
		--TODO: split string using "<p[1-2]>" delimiter
		local t = {side = 1, text = '', colors = {{}, {}}, tokens = {}, cnt = 0}
		v = v .. '<#>'
		local length = 0
		local text = ''
		--in-text names
		v = v:gsub('<(d?i?s?p?l?a?y?name)s*=s*([^>]+)>', function(m1, m2)
			if player(start.f_dialogueRedirection(m2)) then
				if m1 == 'displayname' then
					return displayname()
				elseif m1 == 'name' then
					return name()
				end
			end
			return ''
		end)
		for m1, m2 in v:gmatch('(.-)<([^>]+)>') do
			--text
			if m1 ~= '' then
				length = length + string.len(m1:gsub('\\n', ''))
				text = text .. m1
			end
			if not m2:match('^#$') then
				--colors (TODO: currently not functional)
				if m2:match('^#[a-z0-9]+$') or m2:match('^/$') then
					for i = 1, 2 do
						if t.colors[i][length] == nil then
							t.colors[i][length] = {}
						end
						if m2:match('^/$') then
							t.colors[i][length] = {
								r = motif.dialogue_info['p' .. i .. '_text_font'][4],
								g = motif.dialogue_info['p' .. i .. '_text_font'][5],
								b = motif.dialogue_info['p' .. i .. '_text_font'][6],
							}
						else
							t.colors[i][length] = color:fromHex(m2)
						end
					end
				--side
				elseif m2:match('^p[1-2]$') then
					t.side = tonumber(m2:match('^p([1-2])$'))
				--other tokens
				else
					if t.tokens[length] == nil then
						t.tokens[length] = {}
					end
					local param, val = m2:match('^([a-z1-2]+)%s*=%s*(.+)$')
					if param ~= nil then
						local t_token = {param = param, side = -1, redirection = '', pn = -1, value = {}}
						if param:match('^p[1-2]') then
							t_token.side, t_token.param = param:match('^p([1-2])([a-z]+)$')
							t_token.side = tonumber(t_token.side)
						end
						for i, str in ipairs(main.f_strsplit(',', val)) do --split using "," delimiter
							local strCase = str:lower()
							if i == 1 and (strCase:match('^self') or strCase:match('^playerno') or strCase:match('^partner') or strCase:match('^enemy') or strCase:match('^enemyname') or strCase:match('^partnername')) then
								t_token.redirection = strCase
								t_token.pn = start.f_dialogueRedirection(strCase)
							else
								table.insert(t_token.value, main.f_dataType(str))
							end
						end
						table.insert(t.tokens[length], t_token)
					end
				end
			end
		end
		t.text = text
		table.insert(start.t_dialogue.parsed, t)
	end
	if main.debugLog then main.f_printTable(start.t_dialogue, 'debug/t_dialogue.txt') end
end

function start.f_dialogueTokens(key, t)
	if t.parsed[t.textNum].tokens[key] == nil then
		return
	end
	local rem = 0
	for k, v in ipairs(t.parsed[t.textNum].tokens[key]) do
		rem = k
		--clear text
		if v.param == 'clear' then
			t.parsed[t.textNum].text = ''
		--wait x frames
		elseif v.param == 'wait' then
			t.wait = v.value[1] or 0
			t.checktoken = key
			break
		elseif v.pn ~= -1 or v.param == 'name' then
			--change portrait
			if v.param == 'face' then
				t.face[v.side].pn = v.pn
				t.face[v.side].spr = {v.value[1] or -1, v.value[2] or 0}
			--change name
			elseif v.param == 'name' and start['txt_dialogue_p' .. v.side .. '_name'] ~= nil then
				if v.pn ~= -1 then
					player(v.pn)
					start['txt_dialogue_p' .. v.side .. '_name']:update({text = displayname()})
				else
					start['txt_dialogue_p' .. v.side .. '_name']:update({text = v.value[1] or ''})
				end
			--play sound
			elseif v.param == 'sound' then --pn, group_no, sound_no, volumescale
				charSndPlay(v.pn, v.value[1] or -1, v.value[2] or 0, v.value[3] or 100)
			--change anim
			elseif v.param == 'anim' then --pn, anim_no, anim_elem
				charChangeAnim(v.pn, v.value[1] or 0, v.value[2] or 0)
			--change state
			elseif v.param == 'state' then --pn, state_no
				charChangeState(v.pn, v.value[1] or 0)
			--map operation
			elseif v.param == 'map' then --pn, map_name, value, map_type
				charMapSet(v.pn, v.value[1] or 'dummy', v.value[2] or 0, v.value[3] or 'set')
			end
		end
	end
	for i = rem, 1, -1 do
		table.remove(t.parsed[t.textNum].tokens[key], 1)
	end
end

function start.f_dialogue()
	if not start.f_dialogueInit() then
		return false
	end
	local t = start.t_dialogue
	--draw bg
	for side = 1, 2 do
		if not paused() then
			animUpdate(motif.dialogue_info['p' .. side .. '_bg_data'])
		end
		animDraw(motif.dialogue_info['p' .. side .. '_bg_data'])
	end
	if not paused() then
		t.counter = t.counter + 1
	end
	if t.counter < motif.dialogue_info.starttime then
		return true
	end
	--draw text
	if t.checktoken ~= -1 and t.wait <= 1 then
		start.f_dialogueTokens(t.checktoken, t)
		t.checktoken = -1
	end
	local ok = false
	local length = -1
	for i = 1, 0, -1 do
		local t_parsed = t.parsed[t.textNum - i]
		if i == 0 or (t_parsed ~= nil and t_parsed.side ~= t.parsed[t.textNum].side) then
			ok, length = main.f_textRender(
				start['txt_dialogue_p' .. t_parsed.side .. '_text'],
				t_parsed.text,
				t_parsed.cnt,
				motif.dialogue_info['p' .. t_parsed.side .. '_text_offset'][1],
				motif.dialogue_info['p' .. t_parsed.side .. '_text_offset'][2],
				motif.dialogue_info['p' .. t_parsed.side .. '_text_spacing'][1],
				motif.dialogue_info['p' .. t_parsed.side .. '_text_spacing'][2],
				main.font_def[motif.dialogue_info['p' .. t_parsed.side .. '_text_font'][1] .. motif.dialogue_info['p' .. t_parsed.side .. '_text_font'][7]],
				motif.dialogue_info['p' .. t_parsed.side .. '_text_delay'],
				main.f_lineLength(
					motif.dialogue_info['p' .. t_parsed.side .. '_text_offset'][1],
					motif.info.localcoord[1],
					motif.dialogue_info['p' .. t_parsed.side .. '_text_font'][3],
					motif.dialogue_info['p' .. t_parsed.side .. '_text_window'],
					motif.dialogue_info['p' .. t_parsed.side .. '_text_textwrap']:match('[wl]')
				),
				t_parsed.colors[t_parsed.side]
			)
		end
	end
	if not paused() then
		if t.wait > 0 then
			t.wait = t.wait - 1
		else
			start.f_dialogueTokens(length, t)
			t.parsed[t.textNum].cnt = t.parsed[t.textNum].cnt + 1
		end
	end
	for side = 1, 2 do
		--draw faces
		charSpriteDraw(
			t.face[side].pn,
			{
				t.face[side].spr[1], t.face[side].spr[2],
				motif.dialogue_info['p' .. side .. '_face_spr'][1], motif.dialogue_info['p' .. side .. '_face_spr'][2]
			},
			motif.dialogue_info['p' .. side .. '_face_offset'][1],
			motif.dialogue_info['p' .. side .. '_face_offset'][2],
			motif.dialogue_info['p' .. side .. '_face_scale'][1],
			motif.dialogue_info['p' .. side .. '_face_scale'][2],
			motif.dialogue_info['p' .. side .. '_face_facing'],
			motif.dialogue_info['p' .. side .. '_face_window'][1],
			motif.dialogue_info['p' .. side .. '_face_window'][2],
			motif.dialogue_info['p' .. side .. '_face_window'][3] * config.GameWidth / main.SP_Localcoord[1],
			motif.dialogue_info['p' .. side .. '_face_window'][4] * config.GameHeight / main.SP_Localcoord[2]
		)
		--draw names
		start['txt_dialogue_p' .. side .. '_name']:draw()
	end
	--draw active element
	for side = 1, 2 do
		if t.parsed[t.textNum].side == side then
			if t.activeSide ~= side then
				t.activeSide = side
				animReset(motif.dialogue_info['p' .. side .. '_active_data'])
			end
			animUpdate(motif.dialogue_info['p' .. side .. '_active_data'])
			animDraw(motif.dialogue_info['p' .. side .. '_active_data'])
		end
	end
	if main.f_input(main.t_players, main.f_extractKeys(motif.dialogue_info.skip_key)) then
		charSndStop()
		t.parsed[t.textNum].cnt = 9999
		t.parsed[t.textNum].tokens = {}
		t.wait = 0
		t.switch = true
	elseif ok and not t.switch then
		if #t.parsed > t.textNum then
			t.wait = t.wait + motif.dialogue_info.switchtime
		end
		t.switch = true
	end
	if t.switch and t.wait == 0 then
		if #t.parsed > t.textNum then
			t.textNum = t.textNum + 1
			t.checktoken = 0
			t.switch = false
		elseif t.endtime == -1 then
			t.endtime = t.counter + motif.dialogue_info.endtime
		end
	end
	local key_cancel = main.f_input(main.t_players, main.f_extractKeys(motif.dialogue_info.cancel_key))
	if (t.endtime ~= -1 and t.counter > t.endtime) or (t.counter > motif.dialogue_info.skiptime and key_cancel) then
		if key_cancel then
			charSndStop()
		end
		dialogueReset()
		start.dialogueInit = false
		t.active = false
		return false
	end
	return true
end

return start

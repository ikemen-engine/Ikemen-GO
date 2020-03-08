
local start = {}

setSelColRow(motif.select_info.columns, motif.select_info.rows)

--not used for now
--setShowEmptyCells(motif.select_info.showemptyboxes)
--setRandomSpr(motif.selectbgdef.spr_data, motif.select_info.cell_random_spr[1], motif.select_info.cell_random_spr[2], motif.select_info.cell_random_scale[1], motif.select_info.cell_random_scale[2])
--setCellSpr(motif.selectbgdef.spr_data, motif.select_info.cell_bg_spr[1], motif.select_info.cell_bg_spr[2], motif.select_info.cell_bg_scale[1], motif.select_info.cell_bg_scale[2])

setSelCellSize(motif.select_info.cell_size[1] + motif.select_info.cell_spacing[1], motif.select_info.cell_size[2] + motif.select_info.cell_spacing[2])
setSelCellScale(motif.select_info.portrait_scale[1], motif.select_info.portrait_scale[2])

--default team count after starting the game
local p1NumTurns = math.max(2, config.NumTurns[1])
local p1NumSimul = math.max(2, config.NumSimul[1])
local p1NumTag = math.max(2, config.NumTag[1])
local p1NumRatio = 1
local p2NumTurns = math.max(2, config.NumTurns[1])
local p2NumSimul = math.max(2, config.NumSimul[1])
local p2NumTag = math.max(2, config.NumTag[1])
local p2NumRatio = 1
--default team mode after starting the game
local p1TeamMenu = 1
local p2TeamMenu = 1
--let cursor wrap around
local wrappingX = false
local wrappingY = false
if motif.select_info.wrapping == 1 then
	if motif.select_info.wrapping_x == 1 then
		wrappingX = true
	end
	if motif.select_info.wrapping_y == 1 then
		wrappingY = true
	end
end
--initialize other local variables
local t_victoryBGM = {false, false}
local t_roster = {}
local t_aiRamp = {}
local t_p1Selected = {}
local t_p2Selected = {}
local t_p1Cursor = {}
local t_p2Cursor = {}
local p1RestoreCursor = false
local p2RestoreCursor = false
local p1Cell = false
local p2Cell = false
local p1TeamEnd = false
local p1SelEnd = false
local p1Ratio = false
local p2TeamEnd = false
local p2SelEnd = false
local p2Ratio = false
local selScreenEnd = false
local stageEnd = false
local coopEnd = false
local restoreTeam = false
local resetgrid = false
local continueData = false
local p1NumChars = 0
local p2NumChars = 0
local matchNo = 0
local p1SelX = 0
local p1SelY = 0
local p2SelX = 0
local p2SelY = 0
local p1FaceOffset = 0
local p2FaceOffset = 0
local p1RowOffset = 0
local p2RowOffset = 0
local winner = 0
local t_gameStats = {}
local t_recordText = {}
local winCnt = 0
local loseCnt = 0
local p1FaceX = 0
local p1FaceY = 0
local p2FaceX = 0
local p2FaceY = 0
local p1TeamMode = 0
local p2TeamMode = 0
local lastMatch = 0
local stageNo = 0
local stageList = 0
local timerSelect = 0
local t_savedData = {
	['win'] = {0, 0},
	['lose'] = {0, 0},
	['time'] = {['total'] = 0, ['matches'] = {}},
	['score'] = {['total'] = {0, 0}, ['matches'] = {}},
	['consecutive'] = {0, 0},
}
local fadeType = 'fadein'
local challenger = false

--;===========================================================
--; COMMON FUNCTIONS
--;===========================================================

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
				if j * p2NumChars > #t_chars[i] and #ret > 0 then --if there are not enough characters to fill all slots and at least 1 fight is already assigned
					local stop = true
					for k = (j - 1) * p2NumChars + 1, #t_chars[i] do --loop through characters left for this match
						if main.t_selChars[t_chars[i][k] + 1].onlyme == 1 then --and allow appending if any of the remaining characters has 'onlyme' flag set
							stop = false
						end
					end
					if stop then
						break
					end
				end
				table.insert(ret, {['rmin'] = p2NumChars, ['rmax'] = p2NumChars, ['order'] = i})
			end
			if infinite then
				table.insert(ret, {['rmin'] = p2NumChars, ['rmax'] = p2NumChars, ['order'] = -1})
				break --no point in appending additional matches
			end
		end
	end
	return ret
end

--generates roster table
function start.f_makeRoster(t_ret)
	t_ret = t_ret or {}
	--prepare correct settings tables
	local t = {}
	local t_static = {}
	local t_removable = {}
	--Arcade / Time Attack
	if gameMode('arcade') or gameMode('teamcoop') or gameMode('netplayteamcoop') or gameMode('timeattack') then
		t_static = main.t_orderChars
		if p2Ratio then --Ratio
			if main.t_selChars[t_p1Selected[1].ref + 1].ratiomatches ~= nil and main.t_selOptions[main.t_selChars[t_p1Selected[1].ref + 1].ratiomatches .. "_arcaderatiomatches"] ~= nil then --custom settings exists as char param
				t = main.t_selOptions[main.t_selChars[t_p1Selected[1].ref + 1].ratiomatches .. "_arcaderatiomatches"]
			else --default settings
				t = main.t_selOptions.arcaderatiomatches
			end
		elseif p2TeamMode == 0 then --Single
			if main.t_selChars[t_p1Selected[1].ref + 1].maxmatches ~= nil and main.t_selOptions[main.t_selChars[t_p1Selected[1].ref + 1].maxmatches .. "_arcademaxmatches"] ~= nil then --custom settings exists as char param
				t = start.f_unifySettings(main.t_selOptions[main.t_selChars[t_p1Selected[1].ref + 1].maxmatches .. "_arcademaxmatches"], t_static)
			else --default settings
				t = start.f_unifySettings(main.t_selOptions.arcademaxmatches, t_static)
			end
		else --Simul / Turns / Tag
			if main.t_selChars[t_p1Selected[1].ref + 1].maxmatches ~= nil and main.t_selOptions[main.t_selChars[t_p1Selected[1].ref + 1].maxmatches .. "_teammaxmatches"] ~= nil then --custom settings exists as char param
				t = start.f_unifySettings(main.t_selOptions[main.t_selChars[t_p1Selected[1].ref + 1].maxmatches .. "_teammaxmatches"], t_static)
			else --default settings
				t = start.f_unifySettings(main.t_selOptions.teammaxmatches, t_static)
			end
		end
	--Survival
	elseif gameMode('survival') or gameMode('survivalcoop') or gameMode('netplaysurvivalcoop') then
		t_static = main.t_orderSurvival
		if main.t_selChars[t_p1Selected[1].ref + 1].maxmatches ~= nil and main.t_selOptions[main.t_selChars[t_p1Selected[1].ref + 1].maxmatches .. "_survivalmaxmatches"] ~= nil then --custom settings exists as char param
			t = start.f_unifySettings(main.t_selOptions[main.t_selChars[t_p1Selected[1].ref + 1].maxmatches .. "_survivalmaxmatches"], t_static)
		else --default settings
			t = start.f_unifySettings(main.t_selOptions.survivalmaxmatches, t_static)
		end
	--Boss Rush
	elseif gameMode('bossrush') then
		t_static = {main.t_bossChars}
		for i = 1, math.ceil(#main.t_bossChars / p2NumChars) do --generate ratiomatches style table
			table.insert(t, {['rmin'] = p2NumChars, ['rmax'] = p2NumChars, ['order'] = 1})
		end
	--VS 100 Kumite
	elseif gameMode('vs100kumite') then
		t_static = {main.t_randomChars}
		for i = 1, 100 do --generate ratiomatches style table for 100 matches
			table.insert(t, {['rmin'] = p2NumChars, ['rmax'] = p2NumChars, ['order'] = 1})
		end
	else
		panicError('LUA ERROR: ' .. gameMode() .. ' game mode unrecognized by start.f_makeRoster()')
	end
	--generate roster
	t_removable = main.f_copyTable(t_static) --copy into editable order table
	for i = 1, #t do --for each match number
		if t[i].order == -1 then --infinite matches for this order detected
			table.insert(t_ret, {-1}) --append infinite matches flag at the end
			break
		end
		if t_removable[t[i].order] ~= nil then
			if #t_removable[t[i].order] == 0 and gameMode('vs100kumite') then
				t_removable = main.f_copyTable(t_static) --ensure that there will be at least 100 matches in VS 100 Kumite mode
			end
			if #t_removable[t[i].order] >= 1 then --there is at least 1 character with this order available
				local remaining = t[i].rmin - #t_removable[t[i].order]
				table.insert(t_ret, {}) --append roster table with new subtable
				for j = 1, math.random(math.min(t[i].rmin, #t_removable[t[i].order]), math.min(t[i].rmax, #t_removable[t[i].order])) do --for randomized characters count
					local rand = math.random(1, #t_removable[t[i].order]) --randomize which character will be taken
					table.insert(t_ret[#t_ret], t_removable[t[i].order][rand]) --add such character into roster subtable
					table.remove(t_removable[t[i].order], rand) --and remove it from the available character pool
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

--generates AI ramping table
function start.f_aiRamp(currentMatch)
	local start_match = 0
	local start_diff = 0
	local end_match = 0
	local end_diff = 0
	if currentMatch == 1 then
		t_aiRamp = {}
	end
	--Arcade
	if gameMode('arcade') or gameMode('teamcoop') or gameMode('netplayteamcoop') or gameMode('timeattack') then
		if p2TeamMode == 0 then --Single
			start_match = main.t_selOptions.arcadestart.wins
			start_diff = main.t_selOptions.arcadestart.offset
			end_match =  main.t_selOptions.arcadeend.wins
			end_diff = main.t_selOptions.arcadeend.offset
		elseif p2Ratio then --Ratio
			start_match = main.t_selOptions.ratiostart.wins
			start_diff = main.t_selOptions.ratiostart.offset
			end_match =  main.t_selOptions.ratioend.wins
			end_diff = main.t_selOptions.ratioend.offset
		else --Simul / Turns / Tag
			start_match = main.t_selOptions.teamstart.wins
			start_diff = main.t_selOptions.teamstart.offset
			end_match =  main.t_selOptions.teamend.wins
			end_diff = main.t_selOptions.teamend.offset
		end
	elseif gameMode('survival') or gameMode('survivalcoop') or gameMode('netplaysurvivalcoop') then
		start_match = main.t_selOptions.survivalstart.wins
		start_diff = main.t_selOptions.survivalstart.offset
		end_match =  main.t_selOptions.survivalend.wins
		end_diff = main.t_selOptions.survivalend.offset
	end
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
	for i = currentMatch, lastMatch do
		if i - 1 <= start_match then
			table.insert(t_aiRamp, startAI)
		elseif i - 1 <= end_match then
			local curMatch = i - (start_match + 1)
			table.insert(t_aiRamp, math.floor(curMatch * (endAI - startAI) / (end_match - start_match) + startAI))
		else
			table.insert(t_aiRamp, endAI)
		end
	end
	if main.debugLog then main.f_printTable(t_aiRamp, 'debug/t_aiRamp.txt') end
end

--returns bool depending of rivals match validity
function start.f_rivalsMatch(param, value)
	if main.t_selChars[t_p1Selected[1].ref + 1].rivals ~= nil and main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo] ~= nil then
		if param == nil then --check only if rivals assignment for this match exists at all
			return true
		elseif main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo][param] ~= nil then
			if value == nil then --check only if param is assigned for this rival
				return true
			else --check if param equals value
				return main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo][param] == value
			end
		end
	end
	return false
end

--calculates AI level
function start.f_difficulty(player, offset)
	local t = {}
	if player % 2 ~= 0 then --odd value (Player1 side)
		local pos = math.floor(player / 2 + 0.5)
		t = main.t_selChars[t_p1Selected[pos].ref + 1]
	else --even value (Player2 side)
		local pos = math.floor(player / 2)
		if pos == 1 and start.f_rivalsMatch('ai') then --player2 team leader and arcade mode and ai rivals param exists
			t = main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo]
		else
			t = main.t_selChars[t_p2Selected[pos].ref + 1]
		end
	end
	if t.ai ~= nil then
		return t.ai
	else
		return config.Difficulty + offset
	end
end

--assigns AI level
function start.f_remapAI()
	--Offset
	local offset = 0
	if config.AIRamping and (gameMode('arcade') or gameMode('teamcoop') or gameMode('netplayteamcoop') or gameMode('survival') or gameMode('survivalcoop') or gameMode('netplaysurvivalcoop')) then
		offset = t_aiRamp[matchNo] - config.Difficulty
	end
	--Player 1
	if main.coop then
		remapInput(3, 2) --P3 character uses P2 controls
		setCom(1, 0)
		setCom(3, 0)
	elseif p1TeamMode == 0 then --Single
		if main.p1In == 1 and not main.aiFight then
			setCom(1, 0)
		else
			setCom(1, start.f_difficulty(1, offset))
		end
	elseif p1TeamMode == 1 then --Simul
		if main.p1In == 1 and not main.aiFight then
			setCom(1, 0)
		else
			setCom(1, start.f_difficulty(1, offset))
		end
		for i = 3, p1NumChars * 2 do
			if i % 2 ~= 0 then --odd value
				remapInput(i, 1) --P3/5/7 character uses P1 controls
				setCom(i, start.f_difficulty(i, offset))
			end
		end
	elseif p1TeamMode == 2 then --Turns
		for i = 1, p1NumChars * 2 do
			if i % 2 ~= 0 then --odd value
				if main.p1In == 1 and not main.aiFight then
					remapInput(i, 1) --P1/3/5/7 character uses P1 controls
					setCom(i, 0)
				else
					setCom(i, start.f_difficulty(i, offset))
				end
			end
		end
	else --Tag
		for i = 1, p1NumChars * 2 do
			if i % 2 ~= 0 then --odd value
				if main.p1In == 1 and not main.aiFight then
					remapInput(i, 1) --P1/3/5/7 character uses P1 controls
					setCom(i, 0)
				else
					setCom(i, start.f_difficulty(i, offset))
				end
			end
		end
	end
	--Player 2
	if p2TeamMode == 0 then --Single
		if main.p2In == 2 and not main.aiFight and not main.coop then
			setCom(2, 0)
		else
			setCom(2, start.f_difficulty(2, offset))
		end
	elseif p2TeamMode == 1 then --Simul
		if main.p2In == 2 and not main.aiFight and not main.coop then
			setCom(2, 0)
		else
			setCom(2, start.f_difficulty(2, offset))
		end
		for i = 4, p2NumChars * 2 do
			if i % 2 == 0 then --even value
				remapInput(i, 2) --P4/6/8 character uses P2 controls
				setCom(i, start.f_difficulty(i, offset))
			end
		end
	elseif p2TeamMode == 2 then --Turns
		for i = 2, p2NumChars * 2 do
			if i % 2 == 0 then --even value
				if main.p2In == 2 and not main.aiFight and not main.coop then
					remapInput(i, 2) --P2/4/6/8 character uses P2 controls
					setCom(i, 0)
				else
					setCom(i, start.f_difficulty(i, offset))
				end
			end
		end
	else --Tag
		for i = 2, p2NumChars * 2 do
			if i % 2 == 0 then --even value
				if main.p2In == 2 and not main.aiFight and not main.coop then
					remapInput(i, 2) --P2/4/6/8 character uses P2 controls
					setCom(i, 0)
				else
					setCom(i, start.f_difficulty(i, offset))
				end
			end
		end
	end
end

--sets lifebar, round time, rounds to win
local lifebar = motif.files.fight
function start.f_setRounds()
	--lifebar
	if main.t_charparam.lifebar and main.t_charparam.rivals and start.f_rivalsMatch('lifebar') then --lifebar assigned as rivals param
		lifebar = main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo].lifebar:gsub('\\', '/')
	elseif main.t_charparam.lifebar and main.t_selChars[t_p2Selected[1].ref + 1].lifebar ~= nil then --lifebar assigned as character param
		lifebar = main.t_selChars[t_p2Selected[1].ref + 1].lifebar:gsub('\\', '/')
	else --default lifebar
		lifebar = motif.files.fight
	end
	if lifebar:lower() ~= main.currentLifebar:lower() then
		main.currentLifebar = lifebar
		loadLifebar(lifebar)
		main.framesPerCount = getFramesPerCount()
	end
	setLifeBarElements(main.t_lifebar)
	--round time
	local frames = main.framesPerCount
	local p1FramesMul = 1
	local p2FramesMul = 1
	if p1TeamMode == 3 then
		p1FramesMul = p1NumChars
	end
	if p2TeamMode == 3 then
		p2FramesMul = p2NumChars
	end
	frames = frames * math.max(p1FramesMul, p2FramesMul)
	setFramesPerCount(frames)
	if main.t_charparam.time and main.t_charparam.rivals and start.f_rivalsMatch('time') then --round time assigned as rivals param
		setRoundTime(math.max(-1, main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo].time * frames))
	elseif main.t_charparam.time and main.t_selChars[t_p2Selected[1].ref + 1].time ~= nil then --round time assigned as character param
		setRoundTime(math.max(-1, main.t_selChars[t_p2Selected[1].ref + 1].time * frames))
	else --default round time
		setRoundTime(math.max(-1, main.roundTime * frames))
	end
	--rounds to win
	if main.t_charparam.rounds and main.t_charparam.rivals and start.f_rivalsMatch('rounds') then --round num assigned as rivals param
		setMatchWins(main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo].rounds)
	elseif main.t_charparam.rounds and main.t_selChars[t_p2Selected[1].ref + 1].rounds ~= nil then --round num assigned as character param
		setMatchWins(main.t_selChars[t_p2Selected[1].ref + 1].rounds)
	elseif p2TeamMode == 0 then --default rounds num (Single mode)
		setMatchWins(main.matchWins[1])
	else --default rounds num (Team mode)
		setMatchWins(main.matchWins[2])
	end
	setMatchMaxDrawGames(main.matchWins[3])
	--timer / score counter
	local timer = 0
	local t_score = {0, 0}
	if not challenger then
		t_score = {t_savedData.score.total[1], t_savedData.score.total[2]}
		timer = t_savedData.time.total
	end
	setLifeBarTimer(timer)
	setLifeBarScore(t_score[1], t_score[2])
end

--save data between matches
function start.f_saveData()
	if main.debugLog then main.f_printTable(t_gameStats, 'debug/t_gameStats.txt') end
	if winner == -1 then
		return
	end
	--win/lose matches count, total score
	if winner == 1 then
		t_savedData.win[1] = t_savedData.win[1] + 1
		t_savedData.lose[2] = t_savedData.lose[2] + 1
		t_savedData.score.total[1] = t_gameStats.p1score
	else --if winner == 2 then
		t_savedData.win[2] = t_savedData.win[2] + 1
		t_savedData.lose[1] = t_savedData.lose[1] + 1
		if main.resetScore then --loosing sets score for the next match to lose count
			t_savedData.score.total[1] = t_savedData.lose[1]
		else
			t_savedData.score.total[1] = t_gameStats.p1score
		end
	end
	t_savedData.score.total[2] = t_gameStats.p2score
	--total time
	t_savedData.time.total = t_savedData.time.total + t_gameStats.time
	--time in each round
	table.insert(t_savedData.time.matches, t_gameStats.timerRounds)
	--score in each round
	table.insert(t_savedData.score.matches, t_gameStats.scoreRounds)
	--individual characters
	local t_cheat = {false, false}
	for round = 1, #t_gameStats.match do
		for c, v in ipairs(t_gameStats.match[round]) do
			--cheat flag
			if v.cheated then
				t_cheat[v.teamside + 1] = true
			end
		end
	end
	--max consecutive wins
	for i = 1, #t_cheat do
		if t_cheat[i] then
			setConsecutiveWins(i, 0)
		elseif getConsecutiveWins(i) > t_savedData.consecutive[i] then
			t_savedData.consecutive[i] = getConsecutiveWins(i)
		end
	end
	if main.debugLog then main.f_printTable(t_savedData, 'debug/t_savedData.txt') end
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
		table.insert(ret, main.t_selChars[t[i].ref + 1].char:lower())
	end
	return ret
end

local t_sortRanking = {}
t_sortRanking['arcade'] = function(t, a, b) return t[b].score < t[a].score end
t_sortRanking['teamcoop'] = t_sortRanking.arcade
t_sortRanking['netplayteamcoop'] = t_sortRanking.arcade
t_sortRanking['scorechallenge'] = t_sortRanking.arcade
t_sortRanking['timeattack'] = function(t, a, b) return t[b].time > t[a].time end
t_sortRanking['timechallenge'] = t_sortRanking.timeattack
t_sortRanking['survival'] = function(t, a, b) return t[b].win < t[a].win or (t[b].win == t[a].win and t[b].score < t[a].score) end
t_sortRanking['survivalcoop'] = t_sortRanking.survival
t_sortRanking['netplaysurvivalcoop'] = t_sortRanking.survival
t_sortRanking['bossrush'] = t_sortRanking.survival
t_sortRanking['vs100kumite'] = t_sortRanking.survival

--data saving to stats.json
function f_saveStats()
	local file = io.open("save/stats.json","w+")
	file:write(json.encode(stats, {indent = true}))
	file:close()
end

--store saved data in save/stats.json
function start.f_storeSavedData(mode, cleared)
	if stats.modes == nil then
		stats.modes = {}
	end
	stats.playtime = (stats.playtime or 0) + t_savedData.time.total / 60 --play time
	if stats.modes[mode] == nil then
		stats.modes[mode] = {}
	end
	local t = stats.modes[mode] --mode play time
	t.playtime = (t.playtime or 0) + t_savedData.time.total / 60
	if t_sortRanking[mode] == nil then
		f_saveStats()
		return --mode can't be cleared, so further data collecting is not needed 
	end
	if cleared then
		t.clear = (t.clear or 0) + 1 --number times cleared
	elseif t.clear == nil then
		t.clear = 0
	end
	if not cleared and (mode == 'bossrush' or mode == 'scorechallenge' or mode == 'timechallenge') then
		return --only winning these modes produces ranking data
	end
	--rankings
	t.ranking = f_formattedTable(
		t.ranking,
		{
			['score'] = t_savedData.score.total[1],
			['time'] = t_savedData.time.total / 60,
			['name'] = t_savedData.name or '',
			['chars'] = f_listCharRefs(t_p1Selected),
			['tmode'] = p1TeamMode,
			['ailevel'] = config.Difficulty,
			['win'] = t_savedData.win[1],
			['lose'] = t_savedData.lose[1],
			['consecutive'] = t_savedData.consecutive[1]
		},
		t_sortRanking[mode],
		motif.rankings.max_entries
	)
	f_saveStats()
end

--set end match clearance flags (currently used to prevent end match fading based on various factors, if smooth screen transition is needed)
function start.setLastMatchFlags(mode)
	local t = {{'sound'}, {'sound'}} --I can't think about any situation where preserving match sounds is necessary
	local victoryCheck = main.victoryScreen and motif.victory_screen.enabled == 1
	if  mode == 'versus' or mode == 'netplayversus' then
		victoryCheck = victoryCheck and motif.victory_screen.vs_enabled == 1
	end
	if not victoryCheck and main.resultsTable == nil then
		lastMatchClearance(1, t[1])
		lastMatchClearance(2, t[2])
		return
	end
	local t_noFading = {false, false}
	for i = 1, 2 do
		--victory screen transition
		if victoryCheck then
			if motif.victory_screen.fadein_time > 0 then
				lastMatchClearance(1, t[1])
				lastMatchClearance(2, t[2])
				return
			elseif motif.victory_screen.cpu_enabled == 1 or mode == 'versus' or mode == 'netplayversus' then
				t_noFading[i] = true
			elseif i == 1 and (mode == 'arcade' or mode == 'teamcoop' or mode == 'netplayteamcoop') then
				t_noFading[1] = true
			elseif i == 2 then
				t_noFading[2] = true
			end
		end
		--results screen transition
		if main.resultsTable ~= nil and main.resultsTable.enabled and main.resultsTable.fadein_time == 0 and not t_noFading[i] then
			if i == 1 or mode == 'vs100kumite' then --p1 won the match or lost/draw in vs 100 kumite
				if matchNo == lastMatch or (mode == 'timechallenge' or mode == 'scorechallenge') then --enable fading if it's the last match available or the mode ends after 1 fight
					t_noFading[i] = true
				end
			elseif i == 2 and (mode == 'survival' or mode == 'survivalcoop' or mode == 'netplaysurvivalcoop') then --p1 lost the match in arranged modes
				t_noFading[2] = true
			end
		end
		--continue screen transition
		if motif.continue_screen.enabled == 1 and motif.continue_screen.fadein_time == 0 and not t_noFading[i] then
			if i == 2 and main.credits ~= nil and main.credits > 0 and (mode == 'arcade' or mode == 'teamcoop' or mode == 'timeattack') then
				t_noFading[2] = true
			end
		end
		--set flags
		if t_noFading[i] then
			--print('p' .. i .. ' win = match fading out disabled')
			table.insert(t[i], 'fading')
		end
		lastMatchClearance(i, t[i])
	end
end

--sets stage
function start.f_setStage(num)
	num = num or 0
	--stage
	if not main.stageMenu and not continueData then
		if main.t_charparam.stage and main.t_charparam.rivals and start.f_rivalsMatch('stage') then --stage assigned as rivals param
			num = math.random(1, #main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo].stage)
			num = main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo].stage[num]
		elseif main.t_charparam.stage and main.t_selChars[t_p2Selected[1].ref + 1].stage ~= nil then --stage assigned as character param
			num = math.random(1, #main.t_selChars[t_p2Selected[1].ref + 1].stage)
			num = main.t_selChars[t_p2Selected[1].ref + 1].stage[num]
		elseif (gameMode('arcade') or gameMode('teamcoop') or gameMode('netplayteamcoop')) and main.t_orderStages[main.t_selChars[t_p2Selected[1].ref + 1].order] ~= nil then --stage assigned as stage order param
			num = math.random(1, #main.t_orderStages[main.t_selChars[t_p2Selected[1].ref + 1].order])
			num = main.t_orderStages[main.t_selChars[t_p2Selected[1].ref + 1].order][num]
		else --stage randomly selected
			num = main.t_includeStage[1][math.random(1, #main.t_includeStage[1])]
		end
	end
	setStage(num)
	selectStage(num)
	--music
	t_victoryBGM = {false, false}
	local t = {'music', 'musicalt', 'musiclife', 'musicvictory', 'musicvictory'}
	for i = 1, #t do
		local track = 0
		local music = ''
		local volume = 100
		local loopstart = 0
		local loopend = 0
		if main.stageMenu then --game modes with stage selection screen
			if main.t_selStages[num] ~= nil and main.t_selStages[num][t[i]] ~= nil then --music assigned as stage param
				track = math.random(1, #main.t_selStages[num][t[i]])
				music = main.t_selStages[num][t[i]][track].bgmusic
				volume = main.t_selStages[num][t[i]][track].bgmvolume
				loopstart = main.t_selStages[num][t[i]][track].bgmloopstart
				loopend = main.t_selStages[num][t[i]][track].bgmloopend
			end
		elseif not gameMode('demo') or motif.demo_mode.fight_playbgm == 1 then --game modes other than demo (or demo with stage BGM param enabled)
			if main.t_charparam.music and main.t_charparam.rivals and start.f_rivalsMatch(t[i]) then --music assigned as rivals param
				track = math.random(1, #main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo][t[i]])
				music = main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo][t[i]][track].bgmusic
				volume = main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo][t[i]][track].bgmvolume
				loopstart = main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo][t[i]][track].bgmloopstart
				loopend = main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo][t[i]][track].bgmloopend
			elseif main.t_charparam.music and main.t_selChars[t_p2Selected[1].ref + 1][t[i]] ~= nil then --music assigned as character param
				track = math.random(1, #main.t_selChars[t_p2Selected[1].ref + 1][t[i]])
				music = main.t_selChars[t_p2Selected[1].ref + 1][t[i]][track].bgmusic
				volume = main.t_selChars[t_p2Selected[1].ref + 1][t[i]][track].bgmvolume
				loopstart = main.t_selChars[t_p2Selected[1].ref + 1][t[i]][track].bgmloopstart
				loopend = main.t_selChars[t_p2Selected[1].ref + 1][t[i]][track].bgmloopend
			elseif main.t_selStages[num] ~= nil and main.t_selStages[num][t[i]] ~= nil then --music assigned as stage param
				track = math.random(1, #main.t_selStages[num][t[i]])
				music = main.t_selStages[num][t[i]][track].bgmusic
				volume = main.t_selStages[num][t[i]][track].bgmvolume
				loopstart = main.t_selStages[num][t[i]][track].bgmloopstart
				loopend = main.t_selStages[num][t[i]][track].bgmloopend
			end
		end
		if music ~= '' then
			setStageBGM(i - 1, music, volume, loopstart, loopend)
			if i >= 4 then
				t_victoryBGM[i - 3] = true
			end
		end
	end
	return num
end

--remaps palette based on button press and character's keymap settings
function start.f_reampPal(ref, num)
	if main.t_selChars[ref + 1].pal_keymap[num] ~= nil then
		return main.t_selChars[ref + 1].pal_keymap[num]
	end
	return num
end

--returns palette number
function start.f_selectPal(ref)
	--prepare palette tables
	local t_assignedVals = {} --values = pal numbers already assigned
	local t_assignedKeys = {} --keys = pal numbers already assigned
	for i = 1, #t_p1Selected do
		if t_p1Selected[i].ref == ref then
			table.insert(t_assignedVals, start.f_reampPal(ref, t_p1Selected[i].pal))
			t_assignedKeys[t_assignedVals[#t_assignedVals]] = ''
		end
	end
	for i = 1, #t_p2Selected do
		if t_p2Selected[i].ref == ref then
			table.insert(t_assignedVals, start.f_reampPal(ref, t_p2Selected[i].pal))
			t_assignedKeys[t_assignedVals[#t_assignedVals]] = ''
		end
	end
	--return random palette
	if config.AIRandomColor then
		local t_uniqueVals = {} --values = pal numbers not assigned yet (or all if there are not enough pals for unique appearance of all characters)
		for i = 1, #main.t_selChars[ref + 1].pal do
			if t_assignedKeys[main.t_selChars[ref + 1].pal[i]] == nil or #t_assignedVals >= #main.t_selChars[ref + 1].pal then
				table.insert(t_uniqueVals, main.t_selChars[ref + 1].pal[i])
			end
		end
		if #t_uniqueVals > 0 then --return random unique palette
			return start.f_reampPal(ref, t_uniqueVals[math.random(1, #t_uniqueVals)])
		else --no unique palettes available, randomize from all palettes
			return main.t_selChars[ref + 1].pal[math.random(1, #main.t_selChars[ref + 1].pal)]
		end
	end
	--return first available default palette
	for i = 1, #main.t_selChars[ref + 1].pal_defaults do
		local d = main.t_selChars[ref + 1].pal_defaults[i]
		if t_assignedKeys[d] == nil then
			return start.f_reampPal(ref, d)
		end
	end
	--no default palettes available, force first default palette
	return start.f_reampPal(ref, main.t_selChars[ref + 1].pal_defaults[1])
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
function start.f_setRatio(player)
	if player == 1 then
		if not p1Ratio then
			return nil
		end
		return t_ratioArray[p1NumRatio][#t_p1Selected + 1]
	end
	if not p2Ratio then
		return nil
	end
	if not continueData and not main.p2SelectMenu and #t_p2Selected == 0 then
		if p2NumChars == 3 then
			p2NumRatio = math.random(1, 3)
		elseif p2NumChars == 2 then
			p2NumRatio = math.random(4, 6)
		else
			p2NumRatio = 7
		end
	end
	return t_ratioArray[p2NumRatio][#t_p2Selected + 1]
end

--sets life recovery and ratio level
function start.f_overrideCharData()
	--round 2+ in survival mode
	if matchNo >= 2 and (gameMode('survival') or gameMode('survivalcoop') or gameMode('netplaysurvivalcoop')) then
		local lastRound = #t_gameStats.match
		local removedNum = 0
		local p1Count = 0
		--Turns
		if p1TeamMode == 2 then
			local t_p1Keys = {}
			--for each round in the last match
			for round = 1, #t_gameStats.match do
				--remove character from team if he/she has been defeated
				if not t_gameStats.match[round][1].win or t_gameStats.match[round][1].ko then
					table.remove(t_p1Selected, t_gameStats.match[round][1].memberNo + 1 - removedNum)
					removedNum = removedNum + 1
					p1NumChars = p1NumChars - 1
				--otherwise override character's next match life (done after all rounds have been checked)
				else
					t_p1Keys[t_gameStats.match[round][1].memberNo] = t_gameStats.match[round][1].life
				end
			end
			for k, v in pairs(t_p1Keys) do
				p1Count = p1Count + 1
				overrideCharData(p1Count, {['life'] = v})
			end
		--Single / Simul / Tag
		else
			--for each player data in the last round
			for player = 1, #t_gameStats.match[lastRound] do
				--only check P1 side characters
				if player % 2 ~= 0 and player <= (p1NumChars + removedNum) * 2 then --odd value, team size check just in case
					--in normal survival remove character from team if he/she has been defeated
					if gameMode('survival') and (not t_gameStats.match[lastRound][player].win or t_gameStats.match[lastRound][player].ko) then
						table.remove(t_p1Selected, t_gameStats.match[lastRound][player].memberNo + 1 - removedNum)
						removedNum = removedNum + 1
						p1NumChars = p1NumChars - 1
					--in coop modes defeated character can still fight
					elseif gameMode('survivalcoop') or gameMode('netplaysurvivalcoop') then
						local life = t_gameStats.match[lastRound][player].life
						if life <= 0 then
							life = math.max(1, t_gameStats.match[lastRound][player].lifeMax * config.TurnsRecoveryBase)
						end
						overrideCharData(player, {['life'] = life})
					--otherwise override character's next match life
					else
						if p1Count == 0 then
							p1Count = 1
						else
							p1Count = p1Count + 2
						end
						overrideCharData(p1Count, {['life'] = t_gameStats.match[lastRound][player].life})
					end
				end
			end
		end
		if removedNum > 0 then
			setTeamMode(1, p1TeamMode, p1NumChars)
		end
	end
	--ratio level
	if p1Ratio then
		for i = 1, #t_p1Selected do
			setRatioLevel(i * 2 - 1, t_p1Selected[i].ratio)
			overrideCharData(i * 2 - 1, {['lifeRatio'] = config.RatioLife[t_p1Selected[i].ratio]})
			overrideCharData(i * 2 - 1, {['attackRatio'] = config.RatioAttack[t_p1Selected[i].ratio]})
		end
	end
	if p2Ratio then
		for i = 1, #t_p2Selected do
			setRatioLevel(i * 2, t_p2Selected[i].ratio)
			overrideCharData(i * 2, {['lifeRatio'] = config.RatioLife[t_p2Selected[i].ratio]})
			overrideCharData(i * 2, {['attackRatio'] = config.RatioAttack[t_p2Selected[i].ratio]})
		end
	end
end

--Convert number to name and get rid of the ""
function start.f_getName(ref)
	local tmp = getCharName(ref)
	if main.t_selChars[ref + 1].hidden == 3 then
		tmp = 'Random'
	elseif main.t_selChars[ref + 1].hidden == 2 then
		tmp = ''
	end
	return tmp
end

--draws character names
function start.f_drawName(t, data, font, offsetX, offsetY, scaleX, scaleY, spacingX, spacingY, active_font, active_row)
	for i = 1, #t do
		local x = offsetX
		local f = font
		if active_font ~= nil and active_row ~= nil then
			if i == active_row then
				f = active_font
			else
				f = font
			end
		end
		if motif.font_data[f[1]] ~= -1 then
			data:update({
				font =   motif.font_data[f[1]],
				bank =   f[2],
				align =  f[3],
				text =   start.f_getName(t[i].ref),
				x =      x + (i - 1) * spacingX,
				y =      offsetY + (i - 1) * spacingY,
				scaleX = scaleX,
				scaleY = scaleY,
				r =      f[4],
				g =      f[5],
				b =      f[6],
				src =    f[7],
				dst =    f[8],
			})
			data:draw()
		end
	end
end

--returns correct cell position after moving the cursor
function start.f_cellMovement(selX, selY, cmd, faceOffset, rowOffset, snd)
	local tmpX = selX
	local tmpY = selY
	local tmpFace = faceOffset
	local tmpRow = rowOffset
	local found = false
	if main.input({cmd}, {'$U'}) then
		for i = 1, motif.select_info.rows + motif.select_info.rows_scrolling do
			selY = selY - 1
			if selY < 0 then
				if wrappingY then
					faceOffset = motif.select_info.rows_scrolling * motif.select_info.columns
					rowOffset = motif.select_info.rows_scrolling
					selY = motif.select_info.rows + motif.select_info.rows_scrolling - 1
				else
					faceOffset = tmpFace
					rowOffset = tmpRow
					selY = tmpY
				end
			elseif selY < rowOffset then
				faceOffset = faceOffset - motif.select_info.columns
				rowOffset = rowOffset - 1
			end
			if (start.t_grid[selY + 1][selX + 1].char ~= nil and start.t_grid[selY + 1][selX + 1].hidden ~= 2) or motif.select_info.moveoveremptyboxes == 1 then
				break
			elseif motif.select_info.searchemptyboxesup ~= 0 then
				found, selX = start.f_searchEmptyBoxes(motif.select_info.searchemptyboxesup, selX, selY)
				if found then
					break
				end
			end
		end
	elseif main.input({cmd}, {'$D'}) then
		for i = 1, motif.select_info.rows + motif.select_info.rows_scrolling do
			selY = selY + 1
			if selY >= motif.select_info.rows + motif.select_info.rows_scrolling then
				if wrappingY then
					faceOffset = 0
					rowOffset = 0
					selY = 0
				else
					faceOffset = tmpFace
					rowOffset = tmpRow
					selY = tmpY
				end
			elseif selY >= motif.select_info.rows + rowOffset then
				faceOffset = faceOffset + motif.select_info.columns
				rowOffset = rowOffset + 1
			end
			if (start.t_grid[selY + 1][selX + 1].char ~= nil and start.t_grid[selY + 1][selX + 1].hidden ~= 2) or motif.select_info.moveoveremptyboxes == 1 then
				break
			elseif motif.select_info.searchemptyboxesdown ~= 0 then
				found, selX = start.f_searchEmptyBoxes(motif.select_info.searchemptyboxesdown, selX, selY)
				if found then
					break
				end
			end
		end
	elseif main.input({cmd}, {'$B'}) then
		for i = 1, motif.select_info.columns do
			selX = selX - 1
			if selX < 0 then
				if wrappingX then
					selX = motif.select_info.columns - 1
				else
					selX = tmpX
				end
			end
			if (start.t_grid[selY + 1][selX + 1].char ~= nil and start.t_grid[selY + 1][selX + 1].hidden ~= 2) or motif.select_info.moveoveremptyboxes == 1 then
				break
			end
		end
	elseif main.input({cmd}, {'$F'}) then
		for i = 1, motif.select_info.columns do
			selX = selX + 1
			if selX >= motif.select_info.columns then
				if wrappingX then
					selX = 0
				else
					selX = tmpX
				end
			end
			if (start.t_grid[selY + 1][selX + 1].char ~= nil and start.t_grid[selY + 1][selX + 1].hidden ~= 2) or motif.select_info.moveoveremptyboxes == 1 then
				break
			end
		end
	end
	if tmpX ~= selX or tmpY ~= selY then
		resetgrid = true
		--if tmpRow ~= rowOffset then
			--start.f_resetGrid()
		--end
		sndPlay(motif.files.snd_data, snd[1], snd[2])
	end
	return selX, selY, faceOffset, rowOffset
end

--used by above function to find valid cell in case of dummy character entries
function start.f_searchEmptyBoxes(direction, x, y)
	local selX = x
	local selY = y
	local tmpX = x
	local found = false
	if direction > 0 then --right
		while true do
			x = x + 1
			if x >= motif.select_info.columns then
				x = tmpX
				break
			elseif start.t_grid[y + 1][x + 1].char ~= nil and start.t_grid[selY + 1][selX + 1].hidden ~= 2 then
				found = true
				break
			end
		end
	elseif direction < 0 then --left
		while true do
			x = x - 1
			if x < 0 then
				x = tmpX
				break
			elseif start.t_grid[y + 1][x + 1].char ~= nil and start.t_grid[selY + 1][selX + 1].hidden ~= 2 then
				found = true
				break
			end
		end
	end
	return found, x
end

--generates table with cell coordinates
function start.f_resetGrid()
	start.t_drawFace = {}
	for row = 1, motif.select_info.rows do
		for col = 1, motif.select_info.columns do
			-- Note to anyone editing this function:
			-- The "elseif" chain is important if a "end" is added in the middle it could break the character icon display.
			
			--1Pのランダムセル表示位置 / 1P random cell display position
			if start.t_grid[row + p1RowOffset][col].char == 'randomselect' or start.t_grid[row + p1RowOffset][col].hidden == 3 then
				table.insert(start.t_drawFace, {d = 1, p1 = start.t_grid[row + p1RowOffset][col].char_ref, p2 = start.t_grid[row + p2RowOffset][col].char_ref, x1 = p1FaceX + start.t_grid[row][col].x, x2 = p2FaceX + start.t_grid[row][col].x, y1 = p1FaceY + start.t_grid[row][col].y, y2 = p2FaceY + start.t_grid[row][col].y})
			--1Pのキャラ表示位置 / 1P character display position
			elseif start.t_grid[row + p1RowOffset][col].char ~= nil and start.t_grid[row + p1RowOffset][col].hidden == 0 then
				table.insert(start.t_drawFace, {d = 2, p1 = start.t_grid[row + p1RowOffset][col].char_ref, p2 = start.t_grid[row + p2RowOffset][col].char_ref, x1 = p1FaceX + start.t_grid[row][col].x, x2 = p2FaceX + start.t_grid[row][col].x, y1 = p1FaceY + start.t_grid[row][col].y, y2 = p2FaceY + start.t_grid[row][col].y})
			--Empty boxes display position
			elseif motif.select_info.showemptyboxes == 1 then
				table.insert(start.t_drawFace, {d = 0, p1 = start.t_grid[row + p1RowOffset][col].char_ref, p2 = start.t_grid[row + p2RowOffset][col].char_ref, x1 = p1FaceX + start.t_grid[row][col].x, x2 = p2FaceX + start.t_grid[row][col].x, y1 = p1FaceY + start.t_grid[row][col].y, y2 = p2FaceY + start.t_grid[row][col].y})
			end
			
			--2Pのランダムセル表示位置 / 2P random cell display position
			if start.t_grid[row + p2RowOffset][col].char == 'randomselect' or start.t_grid[row + p2RowOffset][col].hidden == 3 then
				table.insert(start.t_drawFace, {d = 11, p1 = start.t_grid[row + p1RowOffset][col].char_ref, p2 = start.t_grid[row + p2RowOffset][col].char_ref, x1 = p1FaceX + start.t_grid[row][col].x, x2 = p2FaceX + start.t_grid[row][col].x, y1 = p1FaceY + start.t_grid[row][col].y, y2 = p2FaceY + start.t_grid[row][col].y}		)
			--2Pのキャラ表示位置 / 2P character display position
			elseif start.t_grid[row + p2RowOffset][col].char ~= nil and start.t_grid[row + p2RowOffset][col].hidden == 0 then
				table.insert(start.t_drawFace, {d = 12, p1 = start.t_grid[row + p1RowOffset][col].char_ref, p2 = start.t_grid[row + p2RowOffset][col].char_ref, x1 = p1FaceX + start.t_grid[row][col].x, x2 = p2FaceX + start.t_grid[row][col].x, y1 = p1FaceY + start.t_grid[row][col].y, y2 = p2FaceY + start.t_grid[row][col].y})
			--Empty boxes display position
			elseif motif.select_info.showemptyboxes == 1 then
				table.insert(start.t_drawFace, {d = 10, p1 = start.t_grid[row + p1RowOffset][col].char_ref, p2 = start.t_grid[row + p2RowOffset][col].char_ref, x1 = p1FaceX + start.t_grid[row][col].x, x2 = p2FaceX + start.t_grid[row][col].x, y1 = p1FaceY + start.t_grid[row][col].y, y2 = p2FaceY + start.t_grid[row][col].y})
			end
		end
	end
	--if main.debugLog then main.f_printTable(start.t_drawFace, 'debug/t_drawFace.txt') end
end

--sets correct start cell
function start.f_startCell()
	--starting row
	if motif.select_info.p1_cursor_startcell[1] < motif.select_info.rows then
		p1SelY = motif.select_info.p1_cursor_startcell[1]
	else
		p1SelY = 0
	end
	if motif.select_info.p2_cursor_startcell[1] < motif.select_info.rows then
		p2SelY = motif.select_info.p2_cursor_startcell[1]
	else
		p2SelY = 0
	end
	--starting column
	if motif.select_info.p1_cursor_startcell[2] < motif.select_info.columns then
		p1SelX = motif.select_info.p1_cursor_startcell[2]
	else
		p1SelX = 0
	end
	if motif.select_info.p2_cursor_startcell[2] < motif.select_info.columns then
		p2SelX = motif.select_info.p2_cursor_startcell[2]
	else
		p2SelX = 0
	end
end

--unlocks characters on select screen
function start.f_unlockChar(char, flag)
	--setHiddenFlag(char, flag) --not used for now
	main.t_selChars[char + 1].hidden = flag
	start.t_grid[main.t_selChars[char + 1].row][main.t_selChars[char + 1].col].hidden = flag
	start.f_resetGrid()
end

--return t_selChars table out of cell number
function start.f_selGrid(cell)
	if #main.t_selGrid[cell].chars == 0 then
		return {}
	end
	return main.t_selChars[main.t_selGrid[cell].chars[main.t_selGrid[cell].slot]]
end

--return true if slot is selected, update start.t_grid
function start.f_slotSelected(cell, cmd, x, y)
	if #main.t_selGrid[cell].chars > 1 then
		for _, cmdType in ipairs({'swap', 'select'}) do
			if main.t_selGrid[cell][cmdType] ~= nil then
				for k, v in pairs(main.t_selGrid[cell][cmdType]) do
					if main.input({cmd}, {k}) then
						if cmdType == 'swap' then
							local ok = false
							for i, s in ipairs(v) do
								if s > main.t_selGrid[cell].slot then
									main.t_selGrid[cell].slot = s
									ok = true
									break
								end
							end
							if not ok then
								main.t_selGrid[cell].slot = v[1]
								ok = true
							end
						else --select
							main.t_selGrid[cell].slot = v[math.random(1, #v)]
						end
						start.t_grid[y + 1][x + 1].char = start.f_selGrid(cell).char
						start.t_grid[y + 1][x + 1].char_ref = start.f_selGrid(cell).char_ref
						start.f_resetGrid()
						return cmdType == 'select'
					end
				end
			end
		end
	end
	if main.f_btnPalNo(main.cmd[cmd]) == 0 then
		return false
	end
	return true
end

--
local cnt = motif.select_info.columns + 1
local row = 1
local col = 0
start.t_grid = {[row] = {}}
for i = 1, (motif.select_info.rows + motif.select_info.rows_scrolling) * motif.select_info.columns do
	if i == cnt then
		row = row + 1
		cnt = cnt + motif.select_info.columns
		start.t_grid[row] = {}
	end
	col = #start.t_grid[row] + 1
	start.t_grid[row][col] = {x = (col - 1) * (motif.select_info.cell_size[1] + motif.select_info.cell_spacing[1]), y = (row - 1) * (motif.select_info.cell_size[2] + motif.select_info.cell_spacing[2])}
	if start.f_selGrid(i).char ~= nil then
		start.t_grid[row][col].char = start.f_selGrid(i).char
		start.t_grid[row][col].char_ref = start.f_selGrid(i).char_ref
		start.t_grid[row][col].hidden = start.f_selGrid(i).hidden
	end
end
if main.debugLog then main.f_printTable(start.t_grid, 'debug/t_grid.txt') end

--return formatted clear time string
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
	text = text:gsub('%%h', h)
	text = text:gsub('%%m', m)
	text = text:gsub('%%s', s)
	text = text:gsub('%%x', x)
	return text
end

--return formatted record text table
function start.f_getRecordText()
	if motif.select_info['record_text_' .. gameMode()] == nil or stats.modes == nil or stats.modes[gameMode()] == nil or stats.modes[gameMode()].ranking == nil or stats.modes[gameMode()].ranking[1] == nil then
		return {}
	end
	local text = motif.select_info['record_text_' .. gameMode()]
	--time
	text = start.f_clearTimeText(text, stats.modes[gameMode()].ranking[1].time)
	--score
	text = text:gsub('%%p', tostring(stats.modes[gameMode()].ranking[1].score))
	--char name
	local name = '?' --in case character being removed from roster
	if main.t_charDef[stats.modes[gameMode()].ranking[1].chars[1]] ~= nil then
		name = main.t_selChars[main.t_charDef[stats.modes[gameMode()].ranking[1].chars[1]] + 1].displayname
	end
	text = text:gsub('%%c', name)
	--player name
	text = text:gsub('%%n', stats.modes[gameMode()].ranking[1].name)
	return main.f_extractText(text)
end

--cursor sound data, play cursor sound
function start.f_playWave(ref, name, g, n, loops)
	if name == 'stage' then
		local a = main.t_selStages[ref].attachedChar
		if a == nil or a.sound == nil then
			return
		end
		if main.t_selStages[ref][name .. '_wave_data'] == nil then
			main.t_selStages[ref][name .. '_wave_data'] = getWaveData(a.dir .. a.sound, g, n, loops or -1)
		end
		wavePlay(main.t_selStages[ref][name .. '_wave_data'])
	else
		local sound = getCharSnd(ref)
		if sound == nil or sound == '' then
			return
		end
		if main.t_selChars[ref + 1][name .. '_wave_data'] == nil then
			main.t_selChars[ref + 1][name .. '_wave_data'] = getWaveData(main.t_selChars[ref + 1].dir .. sound, g, n, loops or -1)
		end
		wavePlay(main.t_selChars[ref + 1][name .. '_wave_data'])
	end
end

--resets various data
function start.f_selectReset()
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
			start.f_resetGrid()
		end
		col = col + 1
	end
	if main.p2Faces and motif.select_info.double_select == 1 then
		p1FaceX = motif.select_info.pos_p1_double_select[1]
		p1FaceY = motif.select_info.pos_p1_double_select[2]
		p2FaceX = motif.select_info.pos_p2_double_select[1]
		p2FaceY = motif.select_info.pos_p2_double_select[2]
	else
		p1FaceX = motif.select_info.pos[1]
		p1FaceY = motif.select_info.pos[2]
		p2FaceX = motif.select_info.pos[1]
		p2FaceY = motif.select_info.pos[2]
	end
	start.f_resetGrid()
	if gameMode('netplayversus') or gameMode('netplayteamcoop') or gameMode('netplaysurvivalcoop') then
		p1TeamMode = 0
		p2TeamMode = 0
		stageNo = 0
		stageList = 0
	end
	p1Cell = nil
	p2Cell = nil
	t_p1Selected = {}
	t_p2Selected = {}
	p1TeamEnd = false
	p1SelEnd = false
	p1Ratio = false
	p2TeamEnd = false
	p2SelEnd = false
	p2Ratio = false
	if main.p2In == 1 then
		p2TeamEnd = true
		p2SelEnd = true
	elseif main.coop then
		p1TeamEnd = true
		p2TeamEnd = true
	end
	if not main.p2SelectMenu then
		p2SelEnd = true
	end
	selScreenEnd = false
	stageEnd = false
	coopEnd = false
	restoreTeam = false
	continueData = false
	p1NumChars = 1
	p2NumChars = 1
	winner = 0
	matchNo = 0
	if not challenger then
		t_savedData = {
			['win'] = {0, 0},
			['lose'] = {0, 0},
			['time'] = {['total'] = 0, ['matches'] = {}},
			['score'] = {['total'] = {0, 0}, ['matches'] = {}},
			['consecutive'] = {0, 0},
		}
	end
	t_recordText = start.f_getRecordText()
	setMatchNo(matchNo)
end

--;===========================================================
--; SIMPLE LOOP (VS MODE, TEAM VERSUS, TRAINING, WATCH, BONUS GAMES, TIME CHALLENGE, SCORE CHALLENGE)
--;===========================================================
function start.f_selectSimple()
	start.f_startCell()
	t_p1Cursor = {}
	t_p2Cursor = {}
	p1RestoreCursor = false
	p2RestoreCursor = false
	p1TeamMenu = 1
	p2TeamMenu = 1
	p1FaceOffset = 0
	p2FaceOffset = 0
	p1RowOffset = 0
	p2RowOffset = 0
	stageList = 0
	while true do --outer loop (moved back here after pressing ESC)
		start.f_selectReset()
		while true do --inner loop
			fadeType = 'fadein'
			selectStart()
			if not start.f_selectScreen() then
				sndPlay(motif.files.snd_data, motif.select_info.cancel_snd[1], motif.select_info.cancel_snd[2])
				main.f_bgReset(motif.titlebgdef.bg)
				main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
				return
			end
			--fight initialization
			start.f_overrideCharData()
			start.f_remapAI()
			start.f_setRounds()
			stageNo = start.f_setStage(stageNo)
			if start.f_selectVersus() == nil then break end
			start.setLastMatchFlags(gameMode())
			clearColor(motif.selectbgdef.bgclearcolor[1], motif.selectbgdef.bgclearcolor[2], motif.selectbgdef.bgclearcolor[3])
			loadStart()
			winner, t_gameStats = game()
			start.f_saveData()
			if start.f_selectVictory() == nil then break end
			if challenger then
				return
			end
			if winner == -1 then break end --player exit the game via ESC
			start.f_storeSavedData(gameMode(), winner == 1)
			if start.f_result(gameMode()) == nil then break end
			start.f_selectReset()
			--main.f_cmdInput()
			refresh()
		end
		esc(false) --reset ESC
		if gameMode() == 'netplayversus' then
			--resetRemapInput()
			--main.reconnect = winner == -1
		end
	end
end

--;===========================================================
--; ARRANGED LOOP (SURVIVAL, SURVIVAL CO-OP, VS 100 KUMITE, BOSS RUSH)
--;===========================================================
function start.f_selectArranged()
	start.f_startCell()
	t_p1Cursor = {}
	t_p2Cursor = {}
	p1RestoreCursor = false
	p2RestoreCursor = false
	challenger = false
	p1TeamMenu = 1
	p2TeamMenu = 1
	p1FaceOffset = 0
	p2FaceOffset = 0
	p1RowOffset = 0
	p2RowOffset = 0
	winCnt = 0
	loseCnt = 0
	stageList = 0
	while true do --outer loop (moved back here after pressing ESC)
		start.f_selectReset()
		while true do --inner loop
			fadeType = 'fadein'
			selectStart()
			if not start.f_selectScreen() then
				sndPlay(motif.files.snd_data, motif.select_info.cancel_snd[1], motif.select_info.cancel_snd[2])
				main.f_bgReset(motif.titlebgdef.bg)
				main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
				return
			end
			--first match
			if matchNo == 0 then
				--generate roster
				t_roster = start.f_makeRoster()
				lastMatch = #t_roster
				matchNo = 1
				--generate AI ramping table
				start.f_aiRamp(1)
			end
			--assign enemy team
			if #t_p2Selected == 0 then
				local shuffle = true
				for i = 1, #t_roster[matchNo] do
					table.insert(t_p2Selected, {ref = t_roster[matchNo][i], pal = start.f_selectPal(t_roster[matchNo][i]), ratio = start.f_setRatio(2)})
					if shuffle then
						main.f_shuffleTable(t_p2Selected)
					end
				end
			end
			--fight initialization
			setMatchNo(matchNo)
			start.f_overrideCharData()
			start.f_remapAI()
			start.f_setRounds()
			stageNo = start.f_setStage(stageNo)
			if start.f_selectVersus() == nil then break end
			start.setLastMatchFlags(gameMode())
			clearColor(motif.selectbgdef.bgclearcolor[1], motif.selectbgdef.bgclearcolor[2], motif.selectbgdef.bgclearcolor[3])
			loadStart()
			winner, t_gameStats = game()
			start.f_saveData()
			if winner == -1 then break end --player exit the game via ESC
			--player won in any mode or lost/draw in VS 100 Kumite mode
			if winner == 1 or gameMode('vs100kumite') then
				--counter
				if winner == 1 then
					winCnt = winCnt + 1
				else
					loseCnt = loseCnt + 1
				end
				--infinite matches flag detected
				if t_roster[matchNo + 1] ~= nil and t_roster[matchNo + 1][1] == -1 then
					--remove flag
					table.remove(t_roster, matchNo + 1)
					--append entries to existing roster table
					t_roster = start.f_makeRoster(t_roster)
					local lastMatchRamp = lastMatch + 1
					lastMatch = #t_roster
					--append new entries to existing AI ramping table
					start.f_aiRamp(lastMatchRamp)
				end
				--victory screen
				if start.f_selectVictory() == nil then break end
				--no more matches left
				if matchNo == lastMatch then
					--store saved data in save/stats.json
					start.f_storeSavedData(gameMode(), true)
					--result
					if start.f_result(gameMode()) == nil then break end
					--credits
					if motif.end_credits.enabled == 1 and main.f_fileExists(motif.end_credits.storyboard) then
						storyboard.f_storyboard(motif.end_credits.storyboard)
					end
					--game over
					if motif.game_over_screen.enabled == 1 and main.f_fileExists(motif.game_over_screen.storyboard) then
						storyboard.f_storyboard(motif.game_over_screen.storyboard)
					end
					--intro
					if motif.files.intro_storyboard ~= '' then
						storyboard.f_storyboard(motif.files.intro_storyboard)
					end
					main.f_bgReset(motif.titlebgdef.bg)
					main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
					return
				--next match available
				else
					matchNo = matchNo + 1
					t_p2Selected = {}
				end
			--player lost
			elseif winner ~= -1 then
				--counter
				loseCnt = loseCnt + 1
				--victory screen
				if start.f_selectVictory() == nil then break end
				--store saved data in save/stats.json
				start.f_storeSavedData(gameMode(), true and gameMode() ~= 'bossrush')
				--result
				if start.f_result(gameMode()) == nil then break end
				--game over
				if motif.game_over_screen.enabled == 1 and main.f_fileExists(motif.game_over_screen.storyboard) then
					storyboard.f_storyboard(motif.game_over_screen.storyboard)
				end
				--intro
				if motif.files.intro_storyboard ~= '' then
					storyboard.f_storyboard(motif.files.intro_storyboard)
				end
				main.f_bgReset(motif.titlebgdef.bg)
				main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
				return
			end
			--main.f_cmdInput()
			refresh()
		end
		esc(false) --reset ESC
		if gameMode() == 'netplaysurvivalcoop' then
			--resetRemapInput()
			--main.reconnect = winner == -1
		end
	end
end

--;===========================================================
--; ARCADE LOOP (ARCADE, TEAM ARCADE, TEAM CO-OP, TIME ATTACK)
--;===========================================================
function start.f_selectArcade()
	start.f_startCell()
	t_p1Cursor = {}
	t_p2Cursor = {}
	p1RestoreCursor = false
	p2RestoreCursor = false
	challenger = false
	p1TeamMenu = 1
	p2TeamMenu = 1
	p1FaceOffset = 0
	p2FaceOffset = 0
	p1RowOffset = 0
	p2RowOffset = 0
	winCnt = 0
	loseCnt = 0
	--stageEnd = true
	local teamMode = 0
	local numChars = 0
	while true do --outer loop (moved back here after pressing ESC)
		start.f_selectReset()
		while true do --inner loop
			fadeType = 'fadein'
			selectStart()
			--select screen
			if not start.f_selectScreen() then
				sndPlay(motif.files.snd_data, motif.select_info.cancel_snd[1], motif.select_info.cancel_snd[2])
				main.f_bgReset(motif.titlebgdef.bg)
				main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
				return
			end
			--first match
			if matchNo == 0 then
				--generate roster
				t_roster = start.f_makeRoster()
				lastMatch = #t_roster
				matchNo = 1
				--generate AI ramping table
				start.f_aiRamp(1)
				--intro
				if gameMode('arcade') or gameMode('teamcoop') or gameMode('netplayteamcoop') then --not timeattack
					local tPos = main.t_selChars[t_p1Selected[1].ref + 1]
					if tPos.intro ~= nil and main.f_fileExists(tPos.intro) then
						storyboard.f_storyboard(tPos.intro)
					end
				end
			end
			--assign enemy team
			local enemy_ref = 0
			if #t_p2Selected == 0 then
				if p2NumChars ~= #t_roster[matchNo] then
					p2NumChars = #t_roster[matchNo]
					setTeamMode(2, p2TeamMode, p2NumChars)
				end
				local shuffle = true
				for i = 1, #t_roster[matchNo] do
					if i == 1 and start.f_rivalsMatch('char_ref') then --enemy assigned as rivals param
						enemy_ref = main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo].char_ref
						shuffle = false
					else
						enemy_ref = t_roster[matchNo][i]
					end
					table.insert(t_p2Selected, {ref = enemy_ref, pal = start.f_selectPal(enemy_ref), ratio = start.f_setRatio(2)})
					if shuffle then
						main.f_shuffleTable(t_p2Selected)
					end
				end
			end
			--Team conversion to Single match if onlyme paramvalue on any opponents is detected
			if p2NumChars > 1 then
				for i = 1, #t_p2Selected do
					local onlyme = false
					if start.f_rivalsMatch('char_ref') and start.f_rivalsMatch('onlyme', 1) then --team conversion assigned as rivals param
						enemy_ref = main.t_selChars[t_p1Selected[1].ref + 1].rivals[matchNo].char_ref
						onlyme = true
					elseif main.t_selChars[t_p2Selected[i].ref + 1].onlyme == 1 then --team conversion assigned as character param
						enemy_ref = t_p2Selected[i].ref
						onlyme = true
					end
					if onlyme then
						teamMode = p2TeamMode
						numChars = p2NumChars
						p2TeamMode = 0
						p2NumChars = 1
						setTeamMode(2, p2TeamMode, p2NumChars)
						t_p2Selected = {}
						t_p2Selected[1] = {ref = enemy_ref, pal = start.f_selectPal(enemy_ref)}
						restoreTeam = true
						break
					end
				end
			end
			--fight initialization
			challenger = false
			setMatchNo(matchNo)
			start.f_overrideCharData()
			start.f_remapAI()
			start.f_setRounds()
			stageNo = start.f_setStage(stageNo)
			if start.f_selectVersus() == nil then break end
			start.setLastMatchFlags(gameMode())
			clearColor(motif.selectbgdef.bgclearcolor[1], motif.selectbgdef.bgclearcolor[2], motif.selectbgdef.bgclearcolor[3])
			loadStart()
			winner, t_gameStats = game()
			start.f_saveData()
			if t_gameStats.challenger > 0 then --here comes a new challenger
				start.f_challenger()
			elseif winner == -1 then --player exit the game via ESC
				break
			elseif winner == 1 then --player won
				--counter
				winCnt = winCnt + 1
				--victory screen
				if start.f_selectVictory() == nil then break end
				--no more matches left
				if matchNo == lastMatch then
					--store saved data in save/stats.json
					start.f_storeSavedData(gameMode(), true)
					--result
					if start.f_result(gameMode()) == nil then break end
					--ending
					if gameMode('arcade') or gameMode('teamcoop') or gameMode('netplayteamcoop') then --not timeattack
						local tPos = main.t_selChars[t_p1Selected[1].ref + 1]
						if tPos.ending ~= nil and main.f_fileExists(tPos.ending) then
							storyboard.f_storyboard(tPos.ending)
						elseif motif.default_ending.enabled == 1 and main.f_fileExists(motif.default_ending.storyboard) then
							storyboard.f_storyboard(motif.default_ending.storyboard)
						end
					end
					--credits
					if motif.end_credits.enabled == 1 and main.f_fileExists(motif.end_credits.storyboard) then
						storyboard.f_storyboard(motif.end_credits.storyboard)
					end
					--game over
					if motif.game_over_screen.enabled == 1 and main.f_fileExists(motif.game_over_screen.storyboard) then
						storyboard.f_storyboard(motif.game_over_screen.storyboard)
					end
					--intro
					if motif.files.intro_storyboard ~= '' then
						storyboard.f_storyboard(motif.files.intro_storyboard)
					end
					main.f_bgReset(motif.titlebgdef.bg)
					main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
					return
				--next match available
				else
					matchNo = matchNo + 1
					continueData = false
					t_p2Selected = {}
				end
			--player lost and doesn't have any credits left
			elseif main.credits == 0 then
				--counter
				loseCnt = loseCnt + 1
				--victory screen
				if start.f_selectVictory() == nil then break end
				--store saved data in save/stats.json
				start.f_storeSavedData(gameMode(), false)
				--game over
				if motif.game_over_screen.enabled == 1 and main.f_fileExists(motif.game_over_screen.storyboard) then
					storyboard.f_storyboard(motif.game_over_screen.storyboard)
				end
				--intro
				if motif.files.intro_storyboard ~= '' then
					storyboard.f_storyboard(motif.files.intro_storyboard)
				end
				main.f_bgReset(motif.titlebgdef.bg)
				main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
				return
			--player lost but can continue
			else
				--counter
				loseCnt = loseCnt + 1
				--victory screen
				if start.f_selectVictory() == nil then break end
				--continue screen
				if not gameMode('netplayteamcoop') then
					local continueFlag = start.f_continue()
					if continueFlag == nil then
						break
					elseif not continueFlag then
						--store saved data in save/stats.json
						start.f_storeSavedData(gameMode(), false)
						--game over
						if motif.continue_screen.external_gameover == 1 and main.f_fileExists(motif.game_over_screen.storyboard) then
							storyboard.f_storyboard(motif.game_over_screen.storyboard)
						end
						--intro
						if motif.files.intro_storyboard ~= '' then
							storyboard.f_storyboard(motif.files.intro_storyboard)
						end
						main.f_bgReset(motif.titlebgdef.bg)
						main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
						return
					end
				end
				--character selection
				if (not main.quickContinue and not config.QuickContinue) or gameMode('netplayteamcoop') then --true if 'Quick Continue' is disabled or we're playing online
					t_p1Selected = {}
					p1SelEnd = false
					selScreenEnd = false
				end
				continueData = true
			end
			--restore P2 Team settings if needed
			if restoreTeam then
				p2TeamMode = teamMode
				p2NumChars = numChars
				setTeamMode(2, p2TeamMode, p2NumChars)
				restoreTeam = false
			end
			--main.f_cmdInput()
			refresh()
		end
		esc(false) --reset ESC
		if gameMode() == 'netplayteamcoop' then
			--resetRemapInput()
			--main.reconnect = winner == -1
		end
	end
end

function start.f_challenger()
	esc(false)
	challenger = true
	--save values
	local t_p1Selected_sav = main.f_copyTable(t_p1Selected)
	local t_p2Selected_sav = main.f_copyTable(t_p2Selected)
	local p1TeamMenu_sav = main.f_copyTable(main.p1TeamMenu)
	local p2TeamMenu_sav = main.f_copyTable(main.p2TeamMenu)
	local t_charparam_sav = main.f_copyTable(main.t_charparam)
	local p1Ratio_sav = p1Ratio
	local p2Ratio_sav = p2Ratio
	local p1NumRatio_sav = p1NumRatio
	local p2NumRatio_sav = p2NumRatio
	local p1Cell_sav = p1Cell
	local p2Cell_sav = p2Cell
	local matchNo_sav = matchNo
	local stageNo_sav = stageNo
	local restoreTeam_sav = restoreTeam
	local p1TeamMode_sav = p1TeamMode
	local p1NumChars_sav = p1NumChars
	local p2TeamMode_sav = p2TeamMode
	local p2NumChars_sav = p2NumChars
	local gameMode = gameMode()
	local p1score_sav = main.t_lifebar.p1score
	local p2score_sav = main.t_lifebar.p2score
	--temp mode data
	main.txt_mainSelect:update({text = motif.select_info.title_text_teamversus})
	setHomeTeam(1)
	p2NumRatio = 1
	main.p2In = 2
	main.p2SelectMenu = true
	main.stageMenu = true
	main.p2Faces = true
	main.p1TeamMenu = nil
	main.p2TeamMenu = nil
	main.t_lifebar.p1score = true
	main.t_lifebar.p2score = true
	main.f_resetCharparam()
	setGameMode('teamversus')
	--start challenger match
	start.f_selectSimple()
	--restore mode data
	main.txt_mainSelect:update({text = motif.select_info.title_text_arcade})
	setHomeTeam(2)
	main.p2In = 1
	main.p2SelectMenu = false
	main.stageMenu = false
	main.p2Faces = false
	main.p1TeamMenu = p1TeamMenu_sav
	main.p2TeamMenu = p2TeamMenu_sav
	main.t_lifebar.p1score = p1score_sav
	main.t_lifebar.p2score = p2score_sav
	main.t_charparam = t_charparam_sav
	setGameMode(gameMode)
	if esc() then
		challenger = false
		start.f_selectReset()
		return
	end
	if getConsecutiveWins(1) > 0 then
		setConsecutiveWins(1, getConsecutiveWins(1) - 1)
	end
	if winner == 2 then
		--TODO: when player1 team lose continue playing the arcade mode as player2 team
	end
	--restore values
	p1TeamEnd = true
	p2TeamEnd = true
	p1SelEnd = true
	p2SelEnd = true
	t_p1Selected = t_p1Selected_sav
	t_p2Selected = t_p2Selected_sav
	p1Ratio = p1Ratio_sav
	p2Ratio = p2Ratio_sav
	p1NumRatio = p1NumRatio_sav
	p2NumRatio = p2NumRatio_sav
	p1Cell = p1Cell_sav
	p2Cell = p2Cell_sav
	matchNo = matchNo_sav
	stageNo = stageNo_sav
	restoreTeam = restoreTeam_sav
	p1TeamMode = p1TeamMode_sav
	p1NumChars = p1NumChars_sav
	setTeamMode(1, p1TeamMode, p1NumChars)
	p2TeamMode = p2TeamMode_sav
	p2NumChars = p2NumChars_sav
	setTeamMode(2, p2TeamMode, p2NumChars)
	continueData = true
end

--;===========================================================
--; TOURNAMENT LOOP
--;===========================================================
function start.f_selectTournament(size)
	return
end

--;===========================================================
--; TOURNAMENT SCREEN
--;===========================================================
function start.f_selectTournamentScreen(size)
	--draw clearcolor
	clearColor(motif.tournamentbgdef.bgclearcolor[1], motif.tournamentbgdef.bgclearcolor[2], motif.tournamentbgdef.bgclearcolor[3])
	--draw layerno = 0 backgrounds
	bgDraw(motif.tournamentbgdef.bg, false)
	
	--draw layerno = 1 backgrounds
	bgDraw(motif.tournamentbgdef.bg, true)
	--draw fadein / fadeout
	main.fadeActive = fadeScreen(
		fadeType,
		main.fadeStart,
		motif.vs_screen[fadeType .. '_time'],
		motif.vs_screen[fadeType .. '_col'][1],
		motif.vs_screen[fadeType .. '_col'][2],
		motif.vs_screen[fadeType .. '_col'][3]
	)
	--frame transition
	if main.fadeActive then
		commandBufReset(main.cmd[1])
	elseif fadeType == 'fadeout' then
		commandBufReset(main.cmd[1])
		return --skip last frame rendering
	else
		main.f_cmdInput()
	end
	refresh()
end

--;===========================================================
--; SELECT SCREEN
--;===========================================================
local txt_recordSelect = text:create({
	font =   motif.font_data[motif.select_info.record_font[1]],
	bank =   motif.select_info.record_font[2],
	align =  motif.select_info.record_font[3],
	text =   '',
	x =      motif.select_info.record_offset[1],
	y =      motif.select_info.record_offset[2],
	scaleX = motif.select_info.record_font_scale[1],
	scaleY = motif.select_info.record_font_scale[2],
	r =      motif.select_info.record_font[4],
	g =      motif.select_info.record_font[5],
	b =      motif.select_info.record_font[6],
	src =    motif.select_info.record_font[7],
	dst =    motif.select_info.record_font[8],
})
local txt_timerSelect = text:create({
	font =   motif.font_data[motif.select_info.timer_font[1]],
	bank =   motif.select_info.timer_font[2],
	align =  motif.select_info.timer_font[3],
	text =   '',
	x =      motif.select_info.timer_offset[1],
	y =      motif.select_info.timer_offset[2],
	scaleX = motif.select_info.timer_font_scale[1],
	scaleY = motif.select_info.timer_font_scale[2],
	r =      motif.select_info.timer_font[4],
	g =      motif.select_info.timer_font[5],
	b =      motif.select_info.timer_font[6],
	src =    motif.select_info.timer_font[7],
	dst =    motif.select_info.timer_font[8],
})
local txt_p1Name = text:create({
	font =   motif.font_data[motif.select_info.p1_name_font[1]],
	bank =   motif.select_info.p1_name_font[2],
	align =  motif.select_info.p1_name_font[3],
	text =   '',
	x =      0,
	y =      0,
	scaleX = motif.select_info.p1_name_font_scale[1],
	scaleY = motif.select_info.p1_name_font_scale[2],
	r =      motif.select_info.p1_name_font[4],
	g =      motif.select_info.p1_name_font[5],
	b =      motif.select_info.p1_name_font[6],
	src =    motif.select_info.p1_name_font[7],
	dst =    motif.select_info.p1_name_font[8],
})
local p1RandomCount = 0
local p1RandomPortrait = 0
p1RandomPortrait = main.t_randomChars[math.random(1, #main.t_randomChars)]
local txt_p2Name = text:create({
	font =   motif.font_data[motif.select_info.p2_name_font[1]],
	bank =   motif.select_info.p2_name_font[2],
	align =  motif.select_info.p2_name_font[3],
	text =   '',
	x =      0,
	y =      0,
	scaleX = motif.select_info.p2_name_font_scale[1],
	scaleY = motif.select_info.p2_name_font_scale[2],
	r =      motif.select_info.p2_name_font[4],
	g =      motif.select_info.p2_name_font[5],
	b =      motif.select_info.p2_name_font[6],
	src =    motif.select_info.p2_name_font[7],
	dst =    motif.select_info.p2_name_font[8],
})
local p2RandomCount = 0
local p2RandomPortrait = 0
p2RandomPortrait = main.t_randomChars[math.random(1, #main.t_randomChars)]

function start.f_alignOffset(align)
	if align == -1 then
		return 1 --fix for wrong offset after flipping sprites
	end
	return 0
end

function start.f_selectScreen()
	if selScreenEnd then
		return true
	end
	main.f_bgReset(motif.selectbgdef.bg)
	main.f_playBGM(true, motif.music.select_bgm, motif.music.select_bgm_loop, motif.music.select_bgm_volume, motif.music.select_bgm_loopstart, motif.music.select_bgm_loopend)
	local t_enemySelected = {}
	local numChars = p2NumChars
	if main.coop and matchNo > 0 then --coop swap after first match
		t_enemySelected = main.f_copyTable(t_p2Selected)
		p1NumChars = 1
		p2NumChars = 1
		t_p2Selected = {}
		p2SelEnd = false
	end
	timerSelect = 0
	while not selScreenEnd do
		if esc() then
			return false
		end
		--draw clearcolor
		clearColor(motif.selectbgdef.bgclearcolor[1], motif.selectbgdef.bgclearcolor[2], motif.selectbgdef.bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(motif.selectbgdef.bg, false)
		--draw title
		main.txt_mainSelect:draw()
		if p1Cell then
			--draw p1 portrait
			local t_portrait = {}
			if #t_p1Selected < p1NumChars then
				if start.f_selGrid(p1Cell + 1).char == 'randomselect' or start.f_selGrid(p1Cell + 1).hidden == 3 then
					if p1RandomCount < motif.select_info.cell_random_switchtime then
						p1RandomCount = p1RandomCount + 1
					else
						p1RandomPortrait = main.t_randomChars[math.random(1, #main.t_randomChars)]
						p1RandomCount = 0
					end
					sndPlay(motif.files.snd_data, motif.select_info.p1_random_move_snd[1], motif.select_info.p1_random_move_snd[2])
					t_portrait[1] = p1RandomPortrait
				elseif start.f_selGrid(p1Cell + 1).hidden ~= 2 then
					t_portrait[1] = start.f_selGrid(p1Cell + 1).char_ref
				end
			end
			for i = #t_p1Selected, 1, -1 do
				if #t_portrait < motif.select_info.p1_face_num then
					table.insert(t_portrait, t_p1Selected[i].ref)
				end
			end
			t_portrait = main.f_reversedTable(t_portrait)
			for n = #t_portrait, 1, -1 do
				drawPortrait(
					t_portrait[n],
					motif.select_info.p1_face_offset[1] + motif.select_info['p1_c' .. n .. '_face_offset'][1] + (n - 1) * motif.select_info.p1_face_spacing[1] + start.f_alignOffset(motif.select_info.p1_face_facing),
					motif.select_info.p1_face_offset[2] + motif.select_info['p1_c' .. n .. '_face_offset'][2] + (n - 1) * motif.select_info.p1_face_spacing[2],
					motif.select_info.p1_face_facing * motif.select_info.p1_face_scale[1] * motif.select_info['p1_c' .. n .. '_face_scale'][1],
					motif.select_info.p1_face_scale[2] * motif.select_info['p1_c' .. n .. '_face_scale'][2]
				)
			end
		end
		if p2Cell then
			--draw p2 portrait
			local t_portrait = {}
			if #t_p2Selected < p2NumChars then
				if start.f_selGrid(p2Cell + 1).char == 'randomselect' or start.f_selGrid(p2Cell + 1).hidden == 3 then
					if p2RandomCount < motif.select_info.cell_random_switchtime then
						p2RandomCount = p2RandomCount + 1
					else
						p2RandomPortrait = main.t_randomChars[math.random(1, #main.t_randomChars)]
						p2RandomCount = 0
					end
					sndPlay(motif.files.snd_data, motif.select_info.p2_random_move_snd[1], motif.select_info.p2_random_move_snd[2])
					t_portrait[1] = p2RandomPortrait
				elseif start.f_selGrid(p2Cell + 1).hidden ~= 2 then
					t_portrait[1] = start.f_selGrid(p2Cell + 1).char_ref
				end
			end
			for i = #t_p2Selected, 1, -1 do
				if #t_portrait < motif.select_info.p2_face_num then
					table.insert(t_portrait, t_p2Selected[i].ref)
				end
			end
			t_portrait = main.f_reversedTable(t_portrait)
			for n = #t_portrait, 1, -1 do
				drawPortrait(
					t_portrait[n],
					motif.select_info.p2_face_offset[1] + motif.select_info['p2_c' .. n .. '_face_offset'][1] + (n - 1) * motif.select_info.p2_face_spacing[1] + start.f_alignOffset(motif.select_info.p2_face_facing),
					motif.select_info.p2_face_offset[2] + motif.select_info['p2_c' .. n .. '_face_offset'][2] + (n - 1) * motif.select_info.p2_face_spacing[2],
					motif.select_info.p2_face_facing * motif.select_info.p2_face_scale[1] * motif.select_info['p2_c' .. n .. '_face_scale'][1],
					motif.select_info.p2_face_scale[2] * motif.select_info['p2_c' .. n .. '_face_scale'][2]
				)
			end
		end
		--draw cell art (slow for large rosters, this will be likely moved to 'drawFace' function in future)
		for i = 1, #start.t_drawFace do
			-- Let's check if is on the P1 side before drawing the background.
			if start.t_drawFace[i].d == 1 or start.t_drawFace[i].d == 2 or start.t_drawFace[i].d == 0 then
				main.f_animPosDraw(motif.select_info.cell_bg_data, start.t_drawFace[i].x1, start.t_drawFace[i].y1) --draw cell background
			end
			
			if start.t_drawFace[i].d == 1 then --draw random cell
				main.f_animPosDraw(motif.select_info.cell_random_data, start.t_drawFace[i].x1, start.t_drawFace[i].y1)
			elseif start.t_drawFace[i].d == 2 then --draw face cell
				drawSmallPortrait(start.t_drawFace[i].p1, start.t_drawFace[i].x1, start.t_drawFace[i].y1, motif.select_info.portrait_scale[1], motif.select_info.portrait_scale[2])
			end
			--P2 side grid enabled
			if main.p2Faces and motif.select_info.double_select == 1 then
				-- Let's check if is on the P2 side before drawing the background.
				if start.t_drawFace[i].d == 11 or start.t_drawFace[i].d == 12 or start.t_drawFace[i].d == 10 then 
					main.f_animPosDraw(motif.select_info.cell_bg_data, start.t_drawFace[i].x2, start.t_drawFace[i].y2) --draw cell background
				end

				if start.t_drawFace[i].d == 11 then --draw random cell
					main.f_animPosDraw(motif.select_info.cell_random_data, start.t_drawFace[i].x2, start.t_drawFace[i].y2)
				elseif start.t_drawFace[i].d == 12 then --draw face cell
					drawSmallPortrait(start.t_drawFace[i].p2, start.t_drawFace[i].x2, start.t_drawFace[i].y2, motif.select_info.portrait_scale[1], motif.select_info.portrait_scale[2])
				end
			end
		end
		--drawFace(p1FaceX, p1FaceY, p1FaceOffset)
		--if main.p2Faces and motif.select_info.double_select == 1 then
		--	drawFace(p2FaceX, p2FaceY, p2FaceOffset)
		--end
		--draw p1 done cursor
		for i = 1, #t_p1Selected do
			if t_p1Selected[i].cursor ~= nil then
				main.f_animPosDraw(motif.select_info.p1_cursor_done_data, t_p1Selected[i].cursor[1], t_p1Selected[i].cursor[2])
			end
		end
		--draw p2 done cursor
		for i = 1, #t_p2Selected do
			if t_p2Selected[i].cursor ~= nil then
				main.f_animPosDraw(motif.select_info.p2_cursor_done_data, t_p2Selected[i].cursor[1], t_p2Selected[i].cursor[2])
			end
		end
		--Player1 team menu
		if not p1TeamEnd then
			start.f_p1TeamMenu()
		--Player1 select
		elseif main.p1In > 0 or main.p1Char ~= nil then
			start.f_p1SelectMenu()
		end
		--Player2 team menu
		if not p2TeamEnd then
			start.f_p2TeamMenu()
		--Player2 select
		elseif main.p2In > 0 or main.p2Char ~= nil then
			start.f_p2SelectMenu()
		end
		if p1Cell then
			--draw p1 name
			if #t_p1Selected < p1NumChars then
				if start.f_selGrid(p1Cell + 1).char_ref ~= nil then
					txt_p1Name:update({
						align = motif.select_info.p1_name_font[3],
						text =  start.f_getName(start.f_selGrid(p1Cell + 1).char_ref),
						x =     motif.select_info.p1_name_offset[1] + #t_p1Selected * motif.select_info.p1_name_spacing[1],
						y =     motif.select_info.p1_name_offset[2] + #t_p1Selected * motif.select_info.p1_name_spacing[2],
					})
					txt_p1Name:draw()
				end
			end
			start.f_drawName(
				t_p1Selected,
				txt_p1Name,
				motif.select_info.p1_name_font,
				motif.select_info.p1_name_offset[1],
				motif.select_info.p1_name_offset[2],
				motif.select_info.p1_name_font_scale[1],
				motif.select_info.p1_name_font_scale[2],
				motif.select_info.p1_name_spacing[1],
				motif.select_info.p1_name_spacing[2]
			)
		end
		if p2Cell then
			--draw p2 name
			if #t_p2Selected < p2NumChars then
				if start.f_selGrid(p2Cell + 1).char_ref ~= nil then
					txt_p2Name:update({
						align = motif.select_info.p2_name_font[3],
						text =  start.f_getName(start.f_selGrid(p2Cell + 1).char_ref),
						x =     motif.select_info.p2_name_offset[1] + #t_p2Selected * motif.select_info.p2_name_spacing[1],
						y =     motif.select_info.p2_name_offset[2] + #t_p2Selected * motif.select_info.p2_name_spacing[2],
					})
					txt_p2Name:draw()
				end
			end
			start.f_drawName(
				t_p2Selected,
				txt_p2Name,
				motif.select_info.p2_name_font,
				motif.select_info.p2_name_offset[1],
				motif.select_info.p2_name_offset[2],
				motif.select_info.p2_name_font_scale[1],
				motif.select_info.p2_name_font_scale[2],
				motif.select_info.p2_name_spacing[1],
				motif.select_info.p2_name_spacing[2]
			)
		end
		--draw timer
		if motif.select_info.timer_enabled == 1 and p1TeamEnd and (p2TeamEnd or not main.p2SelectMenu) then
			local num = math.floor((motif.select_info.timer_count * motif.select_info.timer_framespercount - timerSelect + motif.select_info.timer_displaytime) / motif.select_info.timer_framespercount + 0.5)
			if num <= -1 then
				timerSelect = -1
				txt_timerSelect:update({text = 0})
			else
				timerSelect = timerSelect + 1
				txt_timerSelect:update({text = math.max(0, num)})
			end
			if timerSelect >= motif.select_info.timer_displaytime then
				txt_timerSelect:draw()
			end
		end
		--draw record text
		for i = 1, #t_recordText do
			txt_recordSelect:update({
				text = t_recordText[i],
				y = motif.select_info.record_offset[2] + (motif.font_def[motif.select_info.record_font[1]].Size[2] + motif.font_def[motif.select_info.record_font[1]].Spacing[2]) * (i - 1),
			})
			txt_recordSelect:draw()
		end
		--team and character selection complete
		if p1SelEnd and p2SelEnd and p1TeamEnd and p2TeamEnd then
			p1RestoreCursor = true
			p2RestoreCursor = true
			if main.stageMenu and not stageEnd then --Stage select
				start.f_stageMenu()
			elseif main.coop and not coopEnd then
				coopEnd = true
				p2TeamEnd = false
			elseif fadeType == 'fadein' then
				main.fadeStart = getFrameCount()
				fadeType = 'fadeout'
			end
		end
		--draw layerno = 1 backgrounds
		bgDraw(motif.selectbgdef.bg, true)
		--draw fadein / fadeout
		main.fadeActive = fadeScreen(
			fadeType,
			main.fadeStart,
			motif.select_info[fadeType .. '_time'],
			motif.select_info[fadeType .. '_col'][1],
			motif.select_info[fadeType .. '_col'][2],
			motif.select_info[fadeType .. '_col'][3]
		)
		--frame transition
		if main.fadeActive then
			commandBufReset(main.cmd[1])
		elseif fadeType == 'fadeout' then
			commandBufReset(main.cmd[1])
			selScreenEnd = true
			break --skip last frame rendering
		else
			main.f_cmdInput()
		end
		refresh()
	end
	if main.coop then
		if matchNo == 0 then --coop swap before first match
			p1TeamMode = 1
			p1NumChars = 2
			setTeamMode(1, p1TeamMode, p1NumChars)
			t_p1Selected[2] = {ref = t_p2Selected[1].ref, pal = t_p2Selected[1].pal}
			t_p2Selected = {}
		else --coop swap after first match
			p1NumChars = 2
			p2NumChars = numChars
			t_p1Selected[2] = {ref = t_p2Selected[1].ref, pal = t_p2Selected[1].pal}
			t_p2Selected = t_enemySelected
		end
	end
	return true
end

--;===========================================================
--; PLAYER 1 TEAM MENU
--;===========================================================
local txt_p1TeamSelfTitle = text:create({
	font =   motif.font_data[motif.select_info.p1_teammenu_selftitle_font[1]],
	bank =   motif.select_info.p1_teammenu_selftitle_font[2],
	align =  motif.select_info.p1_teammenu_selftitle_font[3],
	text =   motif.select_info.p1_teammenu_selftitle_text,
	x =      motif.select_info.p1_teammenu_pos[1] + motif.select_info.p1_teammenu_selftitle_offset[1],
	y =      motif.select_info.p1_teammenu_pos[2] + motif.select_info.p1_teammenu_selftitle_offset[2],
	scaleX = motif.select_info.p1_teammenu_selftitle_font_scale[1],
	scaleY = motif.select_info.p1_teammenu_selftitle_font_scale[2],
	r =      motif.select_info.p1_teammenu_selftitle_font[4],
	g =      motif.select_info.p1_teammenu_selftitle_font[5],
	b =      motif.select_info.p1_teammenu_selftitle_font[6],
	src =    motif.select_info.p1_teammenu_selftitle_font[7],
	dst =    motif.select_info.p1_teammenu_selftitle_font[8],
})
local txt_p1TeamEnemyTitle = text:create({
	font =   motif.font_data[motif.select_info.p1_teammenu_enemytitle_font[1]],
	bank =   motif.select_info.p1_teammenu_enemytitle_font[2],
	align =  motif.select_info.p1_teammenu_enemytitle_font[3],
	text =   motif.select_info.p1_teammenu_enemytitle_text,
	x =      motif.select_info.p1_teammenu_pos[1] + motif.select_info.p1_teammenu_enemytitle_offset[1],
	y =      motif.select_info.p1_teammenu_pos[2] + motif.select_info.p1_teammenu_enemytitle_offset[2],
	scaleX = motif.select_info.p1_teammenu_enemytitle_font_scale[1],
	scaleY = motif.select_info.p1_teammenu_enemytitle_font_scale[2],
	r =      motif.select_info.p1_teammenu_enemytitle_font[4],
	g =      motif.select_info.p1_teammenu_enemytitle_font[5],
	b =      motif.select_info.p1_teammenu_enemytitle_font[6],
	src =    motif.select_info.p1_teammenu_enemytitle_font[7],
	dst =    motif.select_info.p1_teammenu_enemytitle_font[8],
})
local p1TeamActiveCount = 0
local p1TeamActiveFont = 'p1_teammenu_item_active_font'

local t_p1TeamMenu = {
	{data = text:create({}), itemname = 'single', displayname = motif.select_info.teammenu_itemname_single},
	{data = text:create({}), itemname = 'simul', displayname = motif.select_info.teammenu_itemname_simul},
	{data = text:create({}), itemname = 'turns', displayname = motif.select_info.teammenu_itemname_turns},
	{data = text:create({}), itemname = 'tag', displayname = motif.select_info.teammenu_itemname_tag},
	{data = text:create({}), itemname = 'ratio', displayname = motif.select_info.teammenu_itemname_ratio},
}
t_p1TeamMenu = main.f_cleanTable(t_p1TeamMenu, main.t_sort.select_info)

function start.f_p1TeamMenu()
	if #t_p1Cursor > 0 then
		t_p1Cursor = {}
	end
	if main.p1TeamMenu ~= nil then --Predefined team
		p1TeamMode = main.p1TeamMenu.mode
		p1NumChars = main.p1TeamMenu.chars
		setTeamMode(1, p1TeamMode, p1NumChars)
		if main.p1TeamMenu.ratio ~= nil and p1TeamMode == 2 then
			p1NumRatio = main.p1TeamMenu.ratio
			p1Ratio = true
		end
		p1TeamEnd = true
	else
		--Calculate team cursor position
		if main.input({1}, main.f_extractKeys(motif.select_info.teammenu_key_previous)) then
			if p1TeamMenu > 1 then
				sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_move_snd[1], motif.select_info.p1_teammenu_move_snd[2])
				p1TeamMenu = p1TeamMenu - 1
			elseif motif.select_info.teammenu_move_wrapping == 1 then
				sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_move_snd[1], motif.select_info.p1_teammenu_move_snd[2])
				p1TeamMenu = #t_p1TeamMenu
			end
		elseif main.input({1}, main.f_extractKeys(motif.select_info.teammenu_key_next)) then
			if p1TeamMenu < #t_p1TeamMenu then
				sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_move_snd[1], motif.select_info.p1_teammenu_move_snd[2])
				p1TeamMenu = p1TeamMenu + 1
			elseif motif.select_info.teammenu_move_wrapping == 1 then
				sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_move_snd[1], motif.select_info.p1_teammenu_move_snd[2])
				p1TeamMenu = 1
			end
		elseif t_p1TeamMenu[p1TeamMenu].itemname == 'simul' then
			if main.input({1}, main.f_extractKeys(motif.select_info.teammenu_key_subtract)) then
				if p1NumSimul > config.NumSimul[1] then
					sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
					p1NumSimul = p1NumSimul - 1
				end
			elseif main.input({1}, main.f_extractKeys(motif.select_info.teammenu_key_add)) then
				if p1NumSimul < config.NumSimul[2] then
					sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
					p1NumSimul = p1NumSimul + 1
				end
			end
		elseif t_p1TeamMenu[p1TeamMenu].itemname == 'turns' then
			if main.input({1}, main.f_extractKeys(motif.select_info.teammenu_key_subtract)) then
				if p1NumTurns > config.NumTurns[1] then
					sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
					p1NumTurns = p1NumTurns - 1
				end
			elseif main.input({1}, main.f_extractKeys(motif.select_info.teammenu_key_add)) then
				if p1NumTurns < config.NumTurns[2] then
					sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
					p1NumTurns = p1NumTurns + 1
				end
			end
		elseif t_p1TeamMenu[p1TeamMenu].itemname == 'tag' then
			if main.input({1}, main.f_extractKeys(motif.select_info.teammenu_key_subtract)) then
				if p1NumTag > config.NumTag[1] then
					sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
					p1NumTag = p1NumTag - 1
				end
			elseif main.input({1}, main.f_extractKeys(motif.select_info.teammenu_key_add)) then
				if p1NumTag < config.NumTag[2] then
					sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
					p1NumTag = p1NumTag + 1
				end
			end
		elseif t_p1TeamMenu[p1TeamMenu].itemname == 'ratio' then
			if main.input({1}, main.f_extractKeys(motif.select_info.teammenu_key_subtract)) then
				sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
				if p1NumRatio > 1 then
					p1NumRatio = p1NumRatio - 1
				else
					p1NumRatio = 7
				end
			elseif main.input({1}, main.f_extractKeys(motif.select_info.teammenu_key_add)) then
				sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
				if p1NumRatio < 7 then
					p1NumRatio = p1NumRatio + 1
				else
					p1NumRatio = 1
				end
			end
		end
		--Draw team background
		animUpdate(motif.select_info.p1_teammenu_bg_data)
		animDraw(motif.select_info.p1_teammenu_bg_data)
		--Draw team active element background
		animUpdate(motif.select_info['p1_teammenu_bg_' .. t_p1TeamMenu[p1TeamMenu].itemname .. '_data'])
		animDraw(motif.select_info['p1_teammenu_bg_' .. t_p1TeamMenu[p1TeamMenu].itemname .. '_data'])
		--Draw team cursor
		main.f_animPosDraw(
			motif.select_info.p1_teammenu_item_cursor_data,
			(p1TeamMenu - 1) * motif.select_info.p1_teammenu_item_spacing[1],
			(p1TeamMenu - 1) * motif.select_info.p1_teammenu_item_spacing[2]
		)
		--Draw team title
		animUpdate(motif.select_info.p1_teammenu_selftitle_data)
		animDraw(motif.select_info.p1_teammenu_selftitle_data)
		txt_p1TeamSelfTitle:draw()
		for i = 1, #t_p1TeamMenu do
			if i == p1TeamMenu then
				if p1TeamActiveCount < 2 then --delay change
					p1TeamActiveCount = p1TeamActiveCount + 1
				elseif p1TeamActiveFont == 'p1_teammenu_item_active_font' then
					p1TeamActiveFont = 'p1_teammenu_item_active2_font'
					p1TeamActiveCount = 0
				else
					p1TeamActiveFont = 'p1_teammenu_item_active_font'
					p1TeamActiveCount = 0
				end
				--Draw team active font
				t_p1TeamMenu[i].data:update({
					font =   motif.font_data[motif.select_info[p1TeamActiveFont][1]],
					bank =   motif.select_info[p1TeamActiveFont][2],
					align =  motif.select_info[p1TeamActiveFont][3], --p1_teammenu_item_font (winmugen ignores active font facing? Fixed in mugen 1.0)
					text =   t_p1TeamMenu[i].displayname,
					x =      motif.select_info.p1_teammenu_pos[1] + motif.select_info.p1_teammenu_item_offset[1] + motif.select_info.p1_teammenu_item_font_offset[1] + motif.select_info.p1_teammenu_item_spacing[1] * (i - 1),
					y =      motif.select_info.p1_teammenu_pos[2] + motif.select_info.p1_teammenu_item_offset[2] + motif.select_info.p1_teammenu_item_font_offset[2] + motif.select_info.p1_teammenu_item_spacing[2] * (i - 1),
					scaleX = motif.select_info[p1TeamActiveFont .. '_scale'][1],
					scaleY = motif.select_info[p1TeamActiveFont .. '_scale'][2],
					r =      motif.select_info[p1TeamActiveFont][4],
					g =      motif.select_info[p1TeamActiveFont][5],
					b =      motif.select_info[p1TeamActiveFont][6],
					src =    motif.select_info[p1TeamActiveFont][7],
					dst =    motif.select_info[p1TeamActiveFont][8],
				})
				t_p1TeamMenu[i].data:draw()
			else
				--Draw team not active font
				t_p1TeamMenu[i].data:update({
					font =   motif.font_data[motif.select_info.p1_teammenu_item_font[1]],
					bank =   motif.select_info.p1_teammenu_item_font[2],
					align =  motif.select_info.p1_teammenu_item_font[3], --p1_teammenu_item_font (winmugen ignores active font facing? Fixed in mugen 1.0)
					text =   t_p1TeamMenu[i].displayname,
					x =      motif.select_info.p1_teammenu_pos[1] + motif.select_info.p1_teammenu_item_offset[1] + motif.select_info.p1_teammenu_item_font_offset[1] + motif.select_info.p1_teammenu_item_spacing[1] * (i - 1),
					y =      motif.select_info.p1_teammenu_pos[2] + motif.select_info.p1_teammenu_item_offset[2] + motif.select_info.p1_teammenu_item_font_offset[2] + motif.select_info.p1_teammenu_item_spacing[2] * (i - 1),
					scaleX = motif.select_info.p1_teammenu_item_font_scale[1],
					scaleY = motif.select_info.p1_teammenu_item_font_scale[2],
					r =      motif.select_info.p1_teammenu_item_font[4],
					g =      motif.select_info.p1_teammenu_item_font[5],
					b =      motif.select_info.p1_teammenu_item_font[6],
					src =    motif.select_info.p1_teammenu_item_font[7],
					dst =    motif.select_info.p1_teammenu_item_font[8],
				})
				t_p1TeamMenu[i].data:draw()
			end
			--Draw team icons
			if t_p1TeamMenu[i].itemname == 'simul' then
				for j = 1, config.NumSimul[2] do
					if j <= p1NumSimul then
						main.f_animPosDraw(
							motif.select_info.p1_teammenu_value_icon_data,
							(i - 1) * motif.select_info.p1_teammenu_item_spacing[1] + (j - 1) * motif.select_info.p1_teammenu_value_spacing[1],
							(i - 1) * motif.select_info.p1_teammenu_item_spacing[2] + (j - 1) * motif.select_info.p1_teammenu_value_spacing[2]
						)
					else
						main.f_animPosDraw(
							motif.select_info.p1_teammenu_value_empty_icon_data,
							(i - 1) * motif.select_info.p1_teammenu_item_spacing[1] + (j - 1) * motif.select_info.p1_teammenu_value_spacing[1],
							(i - 1) * motif.select_info.p1_teammenu_item_spacing[2] + (j - 1) * motif.select_info.p1_teammenu_value_spacing[2]
						)
					end
				end
			elseif t_p1TeamMenu[i].itemname == 'turns' then
				for j = 1, config.NumTurns[2] do
					if j <= p1NumTurns then
						main.f_animPosDraw(
							motif.select_info.p1_teammenu_value_icon_data,
							(i - 1) * motif.select_info.p1_teammenu_item_spacing[1] + (j - 1) * motif.select_info.p1_teammenu_value_spacing[1],
							(i - 1) * motif.select_info.p1_teammenu_item_spacing[2] + (j - 1) * motif.select_info.p1_teammenu_value_spacing[2]
						)
					else
						main.f_animPosDraw(
							motif.select_info.p1_teammenu_value_empty_icon_data,
							(i - 1) * motif.select_info.p1_teammenu_item_spacing[1] + (j - 1) * motif.select_info.p1_teammenu_value_spacing[1],
							(i - 1) * motif.select_info.p1_teammenu_item_spacing[2] + (j - 1) * motif.select_info.p1_teammenu_value_spacing[2]
						)
					end
				end
			elseif t_p1TeamMenu[i].itemname == 'tag' then
				for j = 1, config.NumTag[2] do
					if j <= p1NumTag then
						main.f_animPosDraw(
							motif.select_info.p1_teammenu_value_icon_data,
							(i - 1) * motif.select_info.p1_teammenu_item_spacing[1] + (j - 1) * motif.select_info.p1_teammenu_value_spacing[1],
							(i - 1) * motif.select_info.p1_teammenu_item_spacing[2] + (j - 1) * motif.select_info.p1_teammenu_value_spacing[2]
						)
					else
						main.f_animPosDraw(
							motif.select_info.p1_teammenu_value_empty_icon_data,
							(i - 1) * motif.select_info.p1_teammenu_item_spacing[1] + (j - 1) * motif.select_info.p1_teammenu_value_spacing[1],
							(i - 1) * motif.select_info.p1_teammenu_item_spacing[2] + (j - 1) * motif.select_info.p1_teammenu_value_spacing[2]
						)
					end
				end
			elseif t_p1TeamMenu[i].itemname == 'ratio' and p1TeamMenu == i then
				animUpdate(motif.select_info['p1_teammenu_ratio' .. p1NumRatio .. '_icon_data'])
				animDraw(motif.select_info['p1_teammenu_ratio' .. p1NumRatio .. '_icon_data'])
			end
		end
		--Confirmed team selection
		if main.input({1}, main.f_extractKeys(motif.select_info.teammenu_key_accept)) then
			sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_done_snd[1], motif.select_info.p1_teammenu_done_snd[2])
			if t_p1TeamMenu[p1TeamMenu].itemname == 'single' then
				p1TeamMode = 0
				p1NumChars = 1
			elseif t_p1TeamMenu[p1TeamMenu].itemname == 'simul' then
				p1TeamMode = 1
				p1NumChars = p1NumSimul
			elseif t_p1TeamMenu[p1TeamMenu].itemname == 'turns' then
				p1TeamMode = 2
				p1NumChars = p1NumTurns
			elseif t_p1TeamMenu[p1TeamMenu].itemname == 'tag' then
				p1TeamMode = 3
				p1NumChars = p1NumTag
			elseif t_p1TeamMenu[p1TeamMenu].itemname == 'ratio' then
				p1TeamMode = 2
				if p1NumRatio <= 3 then
					p1NumChars = 3
				elseif p1NumRatio <= 6 then
					p1NumChars = 2
				else
					p1NumChars = 1
				end
				p1Ratio = true
			end
			setTeamMode(1, p1TeamMode, p1NumChars)
			p1TeamEnd = true
			--main.f_cmdInput()
		end
	end
end

--;===========================================================
--; PLAYER 2 TEAM MENU
--;===========================================================
local txt_p2TeamSelfTitle = text:create({
	font =   motif.font_data[motif.select_info.p2_teammenu_selftitle_font[1]],
	bank =   motif.select_info.p2_teammenu_selftitle_font[2],
	align =  motif.select_info.p2_teammenu_selftitle_font[3],
	text =   motif.select_info.p2_teammenu_selftitle_text,
	x =      motif.select_info.p2_teammenu_pos[1] + motif.select_info.p2_teammenu_selftitle_offset[1],
	y =      motif.select_info.p2_teammenu_pos[2] + motif.select_info.p2_teammenu_selftitle_offset[2],
	scaleX = motif.select_info.p2_teammenu_selftitle_font_scale[1],
	scaleY = motif.select_info.p2_teammenu_selftitle_font_scale[2],
	r =      motif.select_info.p2_teammenu_selftitle_font[4],
	g =      motif.select_info.p2_teammenu_selftitle_font[5],
	b =      motif.select_info.p2_teammenu_selftitle_font[6],
	src =    motif.select_info.p2_teammenu_selftitle_font[7],
	dst =    motif.select_info.p2_teammenu_selftitle_font[8],
})
local txt_p2TeamEnemyTitle = text:create({
	font =   motif.font_data[motif.select_info.p2_teammenu_enemytitle_font[1]],
	bank =   motif.select_info.p2_teammenu_enemytitle_font[2],
	align =  motif.select_info.p2_teammenu_enemytitle_font[3],
	text =   motif.select_info.p2_teammenu_enemytitle_text,
	x =      motif.select_info.p2_teammenu_pos[1] + motif.select_info.p2_teammenu_enemytitle_offset[1],
	y =      motif.select_info.p2_teammenu_pos[2] + motif.select_info.p2_teammenu_enemytitle_offset[2],
	scaleX = motif.select_info.p2_teammenu_enemytitle_font_scale[1],
	scaleY = motif.select_info.p2_teammenu_enemytitle_font_scale[2],
	r =      motif.select_info.p2_teammenu_enemytitle_font[4],
	g =      motif.select_info.p2_teammenu_enemytitle_font[5],
	b =      motif.select_info.p2_teammenu_enemytitle_font[6],
	src =    motif.select_info.p2_teammenu_enemytitle_font[7],
	dst =    motif.select_info.p2_teammenu_enemytitle_font[8],
})
local p2TeamActiveCount = 0
local p2TeamActiveFont = 'p2_teammenu_item_active_font'

local t_p2TeamMenu = {
	{data = text:create({}), itemname = 'single', displayname = motif.select_info.teammenu_itemname_single},
	{data = text:create({}), itemname = 'simul', displayname = motif.select_info.teammenu_itemname_simul},
	{data = text:create({}), itemname = 'turns', displayname = motif.select_info.teammenu_itemname_turns},
	{data = text:create({}), itemname = 'tag', displayname = motif.select_info.teammenu_itemname_tag},
	{data = text:create({}), itemname = 'ratio', displayname = motif.select_info.teammenu_itemname_ratio},
}
t_p2TeamMenu = main.f_cleanTable(t_p2TeamMenu, main.t_sort.select_info)

function start.f_p2TeamMenu()
	if #t_p2Cursor > 0 then
		t_p2Cursor = {}
	end
	if main.p2TeamMenu ~= nil then --Predefined team
		p2TeamMode = main.p2TeamMenu.mode
		p2NumChars = main.p2TeamMenu.chars
		setTeamMode(2, p2TeamMode, p2NumChars)
		if main.p2TeamMenu.ratio ~= nil and p2TeamMode == 2 then
			p2NumRatio = main.p2TeamMenu.ratio
			p2Ratio = true
		end
		p2TeamEnd = true
	else
		--Command swap
		local cmd = 2
		if main.coop then
			cmd = 1
		end
		--Calculate team cursor position
		if main.input({cmd}, main.f_extractKeys(motif.select_info.teammenu_key_previous)) then
			if p2TeamMenu > 1 then
				sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_move_snd[1], motif.select_info.p2_teammenu_move_snd[2])
				p2TeamMenu = p2TeamMenu - 1
			elseif motif.select_info.teammenu_move_wrapping == 1 then
				sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_move_snd[1], motif.select_info.p2_teammenu_move_snd[2])
				p2TeamMenu = #t_p2TeamMenu
			end
		elseif main.input({cmd}, main.f_extractKeys(motif.select_info.teammenu_key_next)) then
			if p2TeamMenu < #t_p2TeamMenu then
				sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_move_snd[1], motif.select_info.p2_teammenu_move_snd[2])
				p2TeamMenu = p2TeamMenu + 1
			elseif motif.select_info.teammenu_move_wrapping == 1 then
				sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_move_snd[1], motif.select_info.p2_teammenu_move_snd[2])
				p2TeamMenu = 1
			end
		elseif t_p2TeamMenu[p2TeamMenu].itemname == 'simul' then
			if main.input({cmd}, main.f_extractKeys(motif.select_info.teammenu_key_add)) then
				if p2NumSimul > config.NumSimul[1] then
					sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
					p2NumSimul = p2NumSimul - 1
				end
			elseif main.input({cmd}, main.f_extractKeys(motif.select_info.teammenu_key_subtract)) then
				if p2NumSimul < config.NumSimul[2] then
					sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
					p2NumSimul = p2NumSimul + 1
				end
			end
		elseif t_p2TeamMenu[p2TeamMenu].itemname == 'turns' then
			if main.input({cmd}, main.f_extractKeys(motif.select_info.teammenu_key_add)) then
				if p2NumTurns > config.NumTurns[1] then
					sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
					p2NumTurns = p2NumTurns - 1
				end
			elseif main.input({cmd}, main.f_extractKeys(motif.select_info.teammenu_key_subtract)) then
				if p2NumTurns < config.NumTurns[2] then
					sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
					p2NumTurns = p2NumTurns + 1
				end
			end
		elseif t_p2TeamMenu[p2TeamMenu].itemname == 'tag' then
			if main.input({cmd}, main.f_extractKeys(motif.select_info.teammenu_key_add)) then
				if p2NumTag > config.NumTag[1] then
					sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
					p2NumTag = p2NumTag - 1
				end
			elseif main.input({cmd}, main.f_extractKeys(motif.select_info.teammenu_key_subtract)) then
				if p2NumTag < config.NumTag[2] then
					sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
					p2NumTag = p2NumTag + 1
				end
			end
		elseif t_p2TeamMenu[p2TeamMenu].itemname == 'ratio' then
			if main.input({cmd}, main.f_extractKeys(motif.select_info.teammenu_key_subtract)) and main.p2SelectMenu then
				sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
				if p2NumRatio > 1 then
					p2NumRatio = p2NumRatio - 1
				else
					p2NumRatio = 7
				end
			elseif main.input({cmd}, main.f_extractKeys(motif.select_info.teammenu_key_add)) and main.p2SelectMenu then
				sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
				if p2NumRatio < 7 then
					p2NumRatio = p2NumRatio + 1
				else
					p2NumRatio = 1
				end
			end
		end
		--Draw team background
		animUpdate(motif.select_info.p2_teammenu_bg_data)
		animDraw(motif.select_info.p2_teammenu_bg_data)
		--Draw team active element background
		animUpdate(motif.select_info['p2_teammenu_bg_' .. t_p2TeamMenu[p2TeamMenu].itemname .. '_data'])
		animDraw(motif.select_info['p2_teammenu_bg_' .. t_p2TeamMenu[p2TeamMenu].itemname .. '_data'])
		--Draw team cursor
		main.f_animPosDraw(
			motif.select_info.p2_teammenu_item_cursor_data,
			(p2TeamMenu - 1) * motif.select_info.p2_teammenu_item_spacing[1],
			(p2TeamMenu - 1) * motif.select_info.p2_teammenu_item_spacing[2]
		)
		--Draw team title
		if main.coop or main.p2In == 1 then
			animUpdate(motif.select_info.p2_teammenu_enemytitle_data)
			animDraw(motif.select_info.p2_teammenu_enemytitle_data)
			txt_p2TeamEnemyTitle:draw()
		else
			animUpdate(motif.select_info.p2_teammenu_selftitle_data)
			animDraw(motif.select_info.p2_teammenu_selftitle_data)
			txt_p2TeamSelfTitle:draw()
		end
		for i = 1, #t_p2TeamMenu do
			if i == p2TeamMenu then
				if p2TeamActiveCount < 2 then --delay change
					p2TeamActiveCount = p2TeamActiveCount + 1
				elseif p2TeamActiveFont == 'p2_teammenu_item_active_font' then
					p2TeamActiveFont = 'p2_teammenu_item_active2_font'
					p2TeamActiveCount = 0
				else
					p2TeamActiveFont = 'p2_teammenu_item_active_font'
					p2TeamActiveCount = 0
				end
				--Draw team active font
				t_p2TeamMenu[i].data:update({
					font =   motif.font_data[motif.select_info[p2TeamActiveFont][1]],
					bank =   motif.select_info[p2TeamActiveFont][2],
					align =  motif.select_info[p2TeamActiveFont][3], --p2_teammenu_item_font (winmugen ignores active font facing? Fixed in mugen 1.0)
					text =   t_p2TeamMenu[i].displayname,
					x =      motif.select_info.p2_teammenu_pos[1] + motif.select_info.p2_teammenu_item_offset[1] + motif.select_info.p2_teammenu_item_font_offset[1] + motif.select_info.p2_teammenu_item_spacing[1] * (i - 1),
					y =      motif.select_info.p2_teammenu_pos[2] + motif.select_info.p2_teammenu_item_offset[2] + motif.select_info.p2_teammenu_item_font_offset[2] + motif.select_info.p2_teammenu_item_spacing[2] * (i - 1),
					scaleX = motif.select_info[p2TeamActiveFont .. '_scale'][1],
					scaleY = motif.select_info[p2TeamActiveFont .. '_scale'][2],
					r =      motif.select_info[p2TeamActiveFont][4],
					g =      motif.select_info[p2TeamActiveFont][5],
					b =      motif.select_info[p2TeamActiveFont][6],
					src =    motif.select_info[p2TeamActiveFont][7],
					dst =    motif.select_info[p2TeamActiveFont][8],
				})
				t_p2TeamMenu[i].data:draw()
			else
				--Draw team not active font
				t_p2TeamMenu[i].data:update({
					font =   motif.font_data[motif.select_info.p2_teammenu_item_font[1]],
					bank =   motif.select_info.p2_teammenu_item_font[2],
					align =  motif.select_info.p2_teammenu_item_font[3], --p2_teammenu_item_font (winmugen ignores active font facing? Fixed in mugen 1.0)
					text =   t_p2TeamMenu[i].displayname,
					x =      motif.select_info.p2_teammenu_pos[1] + motif.select_info.p2_teammenu_item_offset[1] + motif.select_info.p2_teammenu_item_font_offset[1] + motif.select_info.p2_teammenu_item_spacing[1] * (i - 1),
					y =      motif.select_info.p2_teammenu_pos[2] + motif.select_info.p2_teammenu_item_offset[2] + motif.select_info.p2_teammenu_item_font_offset[2] + motif.select_info.p2_teammenu_item_spacing[2] * (i - 1),
					scaleX = motif.select_info.p2_teammenu_item_font_scale[1],
					scaleY = motif.select_info.p2_teammenu_item_font_scale[2],
					r =      motif.select_info.p2_teammenu_item_font[4],
					g =      motif.select_info.p2_teammenu_item_font[5],
					b =      motif.select_info.p2_teammenu_item_font[6],
					src =    motif.select_info.p2_teammenu_item_font[7],
					dst =    motif.select_info.p2_teammenu_item_font[8],
				})
				t_p2TeamMenu[i].data:draw()
			end
			--Draw team icons
			if t_p2TeamMenu[i].itemname == 'simul' then
				for j = 1, config.NumSimul[2] do
					if j <= p2NumSimul then
						main.f_animPosDraw(
							motif.select_info.p2_teammenu_value_icon_data,
							(i - 1) * motif.select_info.p2_teammenu_item_spacing[1] + (j - 1) * motif.select_info.p2_teammenu_value_spacing[1],
							(i - 1) * motif.select_info.p2_teammenu_item_spacing[2] + (j - 1) * motif.select_info.p2_teammenu_value_spacing[2]
						)
					else
						main.f_animPosDraw(
							motif.select_info.p2_teammenu_value_empty_icon_data,
							(i - 1) * motif.select_info.p2_teammenu_item_spacing[1] + (j - 1) * motif.select_info.p2_teammenu_value_spacing[1],
							(i - 1) * motif.select_info.p2_teammenu_item_spacing[2] + (j - 1) * motif.select_info.p2_teammenu_value_spacing[2]
						)
					end
				end
			elseif t_p2TeamMenu[i].itemname == 'turns' then
				for j = 1, config.NumTurns[2] do
					if j <= p2NumTurns then
						main.f_animPosDraw(
							motif.select_info.p2_teammenu_value_icon_data,
							(i - 1) * motif.select_info.p2_teammenu_item_spacing[1] + (j - 1) * motif.select_info.p2_teammenu_value_spacing[1],
							(i - 1) * motif.select_info.p2_teammenu_item_spacing[2] + (j - 1) * motif.select_info.p2_teammenu_value_spacing[2]
						)
					else
						main.f_animPosDraw(
							motif.select_info.p2_teammenu_value_empty_icon_data,
							(i - 1) * motif.select_info.p2_teammenu_item_spacing[1] + (j - 1) * motif.select_info.p2_teammenu_value_spacing[1],
							(i - 1) * motif.select_info.p2_teammenu_item_spacing[2] + (j - 1) * motif.select_info.p2_teammenu_value_spacing[2]
						)
					end
				end
			elseif t_p2TeamMenu[i].itemname == 'tag' then
				for j = 1, config.NumTag[2] do
					if j <= p2NumTag then
						main.f_animPosDraw(
							motif.select_info.p2_teammenu_value_icon_data,
							(i - 1) * motif.select_info.p2_teammenu_item_spacing[1] + (j - 1) * motif.select_info.p2_teammenu_value_spacing[1],
							(i - 1) * motif.select_info.p2_teammenu_item_spacing[2] + (j - 1) * motif.select_info.p2_teammenu_value_spacing[2]
						)
					else
						main.f_animPosDraw(
							motif.select_info.p2_teammenu_value_empty_icon_data,
							(i - 1) * motif.select_info.p2_teammenu_item_spacing[1] + (j - 1) * motif.select_info.p2_teammenu_value_spacing[1],
							(i - 1) * motif.select_info.p2_teammenu_item_spacing[2] + (j - 1) * motif.select_info.p2_teammenu_value_spacing[2]
						)
					end
				end
			elseif t_p2TeamMenu[i].itemname == 'ratio' and p2TeamMenu == i and main.p2SelectMenu then
				animUpdate(motif.select_info['p2_teammenu_ratio' .. p2NumRatio .. '_icon_data'])
				animDraw(motif.select_info['p2_teammenu_ratio' .. p2NumRatio .. '_icon_data'])
			end
		end
		--Confirmed team selection
		if main.input({cmd}, main.f_extractKeys(motif.select_info.teammenu_key_accept)) then
			sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_done_snd[1], motif.select_info.p2_teammenu_done_snd[2])
			if t_p2TeamMenu[p2TeamMenu].itemname == 'single' then
				p2TeamMode = 0
				p2NumChars = 1
			elseif t_p2TeamMenu[p2TeamMenu].itemname == 'simul' then
				p2TeamMode = 1
				p2NumChars = p2NumSimul
			elseif t_p2TeamMenu[p2TeamMenu].itemname == 'turns' then
				p2TeamMode = 2
				p2NumChars = p2NumTurns
			elseif t_p2TeamMenu[p2TeamMenu].itemname == 'tag' then
				p2TeamMode = 3
				p2NumChars = p2NumTag
			elseif t_p2TeamMenu[p2TeamMenu].itemname == 'ratio' then
				p2TeamMode = 2
				if p2NumRatio <= 3 then
					p2NumChars = 3
				elseif p2NumRatio <= 6 then
					p2NumChars = 2
				else
					p2NumChars = 1
				end
				p2Ratio = true
			end
			setTeamMode(2, p2TeamMode, p2NumChars)
			p2TeamEnd = true
			--main.f_cmdInput()
		end
	end
end

--;===========================================================
--; PLAYER 1 SELECT MENU
--;===========================================================
function start.f_p1SelectMenu()
	--predefined selection
	if main.p1Char ~= nil then
		local t = {}
		for i = 1, #main.p1Char do
			if t[main.p1Char[i]] == nil then
				t[main.p1Char[i]] = ''
			end
			t_p1Selected[i] = {
				ref = main.p1Char[i],
				pal = start.f_selectPal(main.p1Char[i])
			}
		end
		p1SelEnd = true
		return
	--manual selection
	elseif not p1SelEnd then
		resetgrid = false
		--cell movement
		if p1RestoreCursor and t_p1Cursor[p1NumChars - #t_p1Selected] ~= nil then --restore saved position
			p1SelX = t_p1Cursor[p1NumChars - #t_p1Selected][1]
			p1SelY = t_p1Cursor[p1NumChars - #t_p1Selected][2]
			p1FaceOffset = t_p1Cursor[p1NumChars - #t_p1Selected][3]
			p1RowOffset = t_p1Cursor[p1NumChars - #t_p1Selected][4]
			t_p1Cursor[p1NumChars - #t_p1Selected] = nil
		else --calculate current position
			p1SelX, p1SelY, p1FaceOffset, p1RowOffset = start.f_cellMovement(p1SelX, p1SelY, 1, p1FaceOffset, p1RowOffset, motif.select_info.p1_cursor_move_snd)
		end
		p1Cell = p1SelX + motif.select_info.columns * p1SelY
		--draw active cursor
		local cursorX = p1FaceX + p1SelX * (motif.select_info.cell_size[1] + motif.select_info.cell_spacing[1])
		local cursorY = p1FaceY + (p1SelY - p1RowOffset) * (motif.select_info.cell_size[2] + motif.select_info.cell_spacing[2])
		if resetgrid == true then
			start.f_resetGrid()
		end
		if start.f_selGrid(p1Cell + 1).hidden ~= 1 then
			main.f_animPosDraw(motif.select_info.p1_cursor_active_data, cursorX, cursorY)
		end
		--cell selected
		if start.f_slotSelected(p1Cell + 1, 1, p1SelX, p1SelY) and start.f_selGrid(p1Cell + 1).char ~= nil and start.f_selGrid(p1Cell + 1).hidden ~= 2 then
			sndPlay(motif.files.snd_data, motif.select_info.p1_cursor_done_snd[1], motif.select_info.p1_cursor_done_snd[2])
			local selected = start.f_selGrid(p1Cell + 1).char_ref
			if main.t_selChars[selected + 1].char == 'randomselect' or main.t_selChars[selected + 1].hidden == 3 then
				selected = main.t_randomChars[math.random(1, #main.t_randomChars)]
			end
			start.f_playWave(selected, 'cursor', motif.select_info.p1_select_snd[1], motif.select_info.p1_select_snd[2])
			table.insert(t_p1Selected, {
				ref = selected,
				pal = main.f_btnPalNo(main.cmd[1]),
				cursor = {cursorX, cursorY, p1RowOffset},
				ratio = start.f_setRatio(1)
			})
			t_p1Cursor[p1NumChars - #t_p1Selected + 1] = {p1SelX, p1SelY, p1FaceOffset, p1RowOffset}
			if #t_p1Selected == p1NumChars then --if all characters have been chosen
				if main.p2In == 1 and matchNo == 0 then --if player1 is allowed to select p2 characters
					p2TeamEnd = false
					p2SelEnd = false
					--commandBufReset(main.cmd[2])
				end
				p1SelEnd = true
			end
			main.f_cmdInput()
		--select screen timer reached 0
		elseif motif.select_info.timer_enabled == 1 and timerSelect == -1 then
			sndPlay(motif.files.snd_data, motif.select_info.p1_cursor_done_snd[1], motif.select_info.p1_cursor_done_snd[2])
			local selected = start.f_selGrid(p1Cell + 1).char_ref
			local rand = false
			for i = #t_p1Selected + 1, p1NumChars do
				if rand or main.t_selChars[selected + 1].char == 'randomselect' or main.t_selChars[selected + 1].hidden == 3 then
					selected = main.t_randomChars[math.random(1, #main.t_randomChars)]
				end
				if not rand then --play it just for the first character
					start.f_playWave(selected, 'cursor', motif.select_info.p1_select_snd[1], motif.select_info.p1_select_snd[2])
				end
				rand = true
				table.insert(t_p1Selected, {
					ref = selected,
					pal = math.random(1, 12),
					cursor = {cursorX, cursorY, p1RowOffset},
					ratio = start.f_setRatio(1)
				})
				t_p1Cursor[p1NumChars - #t_p1Selected + 1] = {p1SelX, p1SelY, p1FaceOffset, p1RowOffset}
			end
			if main.p2SelectMenu and main.p2In == 1 and matchNo == 0 then --if player1 is allowed to select p2 characters
				p2TeamMode = p1TeamMode
				p2NumChars = p1NumChars
				setTeamMode(2, p2TeamMode, p2NumChars)
				p2Cell = p1Cell
				p2SelX = p1SelX
				p2SelY = p1SelY
				p2FaceOffset = p1FaceOffset
				p2RowOffset = p1RowOffset
				for i = 1, p2NumChars do
					selected = main.t_randomChars[math.random(1, #main.t_randomChars)]
					table.insert(t_p2Selected, {
						ref = selected,
						pal = math.random(1, 12),
						cursor = {cursorX, cursorY, p2RowOffset},
						ratio = start.f_setRatio(2)
					})
					t_p2Cursor[p2NumChars - #t_p2Selected + 1] = {p2SelX, p2SelY, p2FaceOffset, p2RowOffset}
				end
			end
			if main.stageMenu then
				stageNo = main.t_includeStage[2][math.random(1, #main.t_includeStage[2])]
				stageEnd = true
			end
			p1SelEnd = true
		end
	end
end

--;===========================================================
--; PLAYER 2 SELECT MENU
--;===========================================================
function start.f_p2SelectMenu()
	--predefined selection
	if main.p2Char ~= nil then
		local t = {}
		for i = 1, #main.p2Char do
			if t[main.p2Char[i]] == nil then
				t[main.p2Char[i]] = ''
			end
			t_p2Selected[i] = {
				ref = main.p2Char[i],
				pal = start.f_selectPal(main.p2Char[i])
			}
		end
		p2SelEnd = true
		return
	--p2 selection disabled
	elseif not main.p2SelectMenu then
		p2SelEnd = true
		return
	--manual selection
	elseif not p2SelEnd then
		resetgrid = false
		--cell movement
		if p2RestoreCursor and t_p2Cursor[p2NumChars - #t_p2Selected] ~= nil then --restore saved position
			p2SelX = t_p2Cursor[p2NumChars - #t_p2Selected][1]
			p2SelY = t_p2Cursor[p2NumChars - #t_p2Selected][2]
			p2FaceOffset = t_p2Cursor[p2NumChars - #t_p2Selected][3]
			p2RowOffset = t_p2Cursor[p2NumChars - #t_p2Selected][4]
			t_p2Cursor[p2NumChars - #t_p2Selected] = nil
		else --calculate current position
			p2SelX, p2SelY, p2FaceOffset, p2RowOffset = start.f_cellMovement(p2SelX, p2SelY, 2, p2FaceOffset, p2RowOffset, motif.select_info.p2_cursor_move_snd)
		end
		p2Cell = p2SelX + motif.select_info.columns * p2SelY
		--draw active cursor
		local cursorX = p2FaceX + p2SelX * (motif.select_info.cell_size[1] + motif.select_info.cell_spacing[1])
		local cursorY = p2FaceY + (p2SelY - p2RowOffset) * (motif.select_info.cell_size[2] + motif.select_info.cell_spacing[2])
		if resetgrid == true then
			start.f_resetGrid()
		end
		main.f_animPosDraw(motif.select_info.p2_cursor_active_data, cursorX, cursorY)
		--cell selected
		if start.f_slotSelected(p2Cell + 1, 2, p2SelX, p2SelY) and start.f_selGrid(p2Cell + 1).char ~= nil and start.f_selGrid(p2Cell + 1).hidden ~= 2 then
			sndPlay(motif.files.snd_data, motif.select_info.p2_cursor_done_snd[1], motif.select_info.p2_cursor_done_snd[2])
			local selected = start.f_selGrid(p2Cell + 1).char_ref
			if main.t_selChars[selected + 1].char == 'randomselect' or main.t_selChars[selected + 1].hidden == 3 then
				selected = main.t_randomChars[math.random(1, #main.t_randomChars)]
			end
			start.f_playWave(selected, 'cursor', motif.select_info.p2_select_snd[1], motif.select_info.p2_select_snd[2])
			table.insert(t_p2Selected, {
				ref = selected,
				pal = main.f_btnPalNo(main.cmd[2]),
				cursor = {cursorX, cursorY, p2RowOffset},
				ratio = start.f_setRatio(2)
			})
			t_p2Cursor[p2NumChars - #t_p2Selected + 1] = {p2SelX, p2SelY, p2FaceOffset, p2RowOffset}
			if #t_p2Selected == p2NumChars then
				p2SelEnd = true
			end
			main.f_cmdInput()
		--select screen timer reached 0
		elseif motif.select_info.timer_enabled == 1 and timerSelect == -1 then
			sndPlay(motif.files.snd_data, motif.select_info.p2_cursor_done_snd[1], motif.select_info.p2_cursor_done_snd[2])
			local selected = start.f_selGrid(p2Cell + 1).char_ref
			local rand = false
			for i = #t_p2Selected + 1, p2NumChars do
				if rand or main.t_selChars[selected + 1].char == 'randomselect' or main.t_selChars[selected + 1].hidden == 3 then
					selected = main.t_randomChars[math.random(1, #main.t_randomChars)]
				end
				if not rand then --play it just for the first character
					start.f_playWave(selected, 'cursor', motif.select_info.p2_select_snd[1], motif.select_info.p2_select_snd[2])
				end
				rand = true
				table.insert(t_p2Selected, {
					ref = selected,
					pal = math.random(1, 12),
					cursor = {cursorX, cursorY, p2RowOffset},
					ratio = start.f_setRatio(2)
				})
				t_p2Cursor[p2NumChars - #t_p2Selected + 1] = {p2SelX, p2SelY, p2FaceOffset, p2RowOffset}
			end
			p2SelEnd = true
		end
	end
end

--;===========================================================
--; STAGE MENU
--;===========================================================
local txt_selStage = text:create({})

local stageActiveCount = 0
local stageActiveFont = 'stage_active_font'

function start.f_stageMenu()
	if main.input({1, 2}, {'$B'}) then
		sndPlay(motif.files.snd_data, motif.select_info.stage_move_snd[1], motif.select_info.stage_move_snd[2])
		stageList = stageList - 1
		if stageList < 0 then stageList = #main.t_includeStage[2] end
	elseif main.input({1, 2}, {'$F'}) then
		sndPlay(motif.files.snd_data, motif.select_info.stage_move_snd[1], motif.select_info.stage_move_snd[2])
		stageList = stageList + 1
		if stageList > #main.t_includeStage[2] then stageList = 0 end
	elseif main.input({1, 2}, {'$U'}) then
		sndPlay(motif.files.snd_data, motif.select_info.stage_move_snd[1], motif.select_info.stage_move_snd[2])
		for i = 1, 10 do
			stageList = stageList - 1
			if stageList < 0 then stageList = #main.t_includeStage[2] end
		end
	elseif main.input({1, 2}, {'$D'}) then
		sndPlay(motif.files.snd_data, motif.select_info.stage_move_snd[1], motif.select_info.stage_move_snd[2])
		for i = 1, 10 do
			stageList = stageList + 1
			if stageList > #main.t_includeStage[2] then stageList = 0 end
		end
	end
	if stageList == 0 then --draw random stage portrait loaded from screenpack SFF
		animUpdate(motif.select_info.stage_portrait_random_data)
		animDraw(motif.select_info.stage_portrait_random_data)	
	else --draw stage portrait loaded from stage SFF
		drawStagePortrait(
			stageList,
			motif.select_info.stage_pos[1] + motif.select_info.stage_portrait_offset[1],
			motif.select_info.stage_pos[2] + motif.select_info.stage_portrait_offset[2],
			--[[motif.select_info.stage_portrait_facing * ]]motif.select_info.stage_portrait_scale[1],
			motif.select_info.stage_portrait_scale[2]
		)
	end
	if main.input({1, 2}, {'pal'}) then
		sndPlay(motif.files.snd_data, motif.select_info.stage_done_snd[1], motif.select_info.stage_done_snd[2])
		if stageList == 0 then
			stageNo = main.t_includeStage[2][math.random(1, #main.t_includeStage[2])]
		else
			stageNo = main.t_includeStage[2][stageList]
		end
		stageActiveFont = 'stage_done_font'
		stageEnd = true
		--main.f_cmdInput()
	else
		if stageActiveCount < 2 then --delay change
			stageActiveCount = stageActiveCount + 1
		elseif stageActiveFont == 'stage_active_font' then
			stageActiveFont = 'stage_active2_font'
			stageActiveCount = 0
		else
			stageActiveFont = 'stage_active_font'
			stageActiveCount = 0
		end
	end
	local t_txt = {}
	if stageList == 0 then
		t_txt[1] = motif.select_info.stage_random_text
	else
		t_txt = main.f_extractText(motif.select_info.stage_text, stageList, getStageName(main.t_includeStage[2][stageList]))
	end
	for i = 1, #t_txt do
		txt_selStage:update({
			font =   motif.font_data[motif.select_info[stageActiveFont][1]],
			bank =   motif.select_info[stageActiveFont][2],
			align =  motif.select_info[stageActiveFont][3],
			text =   t_txt[i],
			x =      motif.select_info.stage_pos[1],
			y =      motif.select_info.stage_pos[2] + (motif.font_def[motif.select_info[stageActiveFont][1]].Size[2] + motif.font_def[motif.select_info[stageActiveFont][1]].Spacing[2]) * (i - 1),
			scaleX = motif.select_info[stageActiveFont .. '_scale'][1],
			scaleY = motif.select_info[stageActiveFont .. '_scale'][2],
			r =      motif.select_info[stageActiveFont][4],
			g =      motif.select_info[stageActiveFont][5],
			b =      motif.select_info[stageActiveFont][6],
			src =    motif.select_info[stageActiveFont][7],
			dst =    motif.select_info[stageActiveFont][8],
		})
		txt_selStage:draw()
	end
end

--;===========================================================
--; VERSUS SCREEN
--;===========================================================
local txt_p1NameVS = text:create({
	font =   motif.font_data[motif.vs_screen.p1_name_font[1]],
	bank =   motif.vs_screen.p1_name_font[2],
	align =  motif.vs_screen.p1_name_font[3],
	text =   '',
	x =      0,
	y =      0,
	scaleX = motif.vs_screen.p1_name_font_scale[1],
	scaleY = motif.vs_screen.p1_name_font_scale[2],
	r =      motif.vs_screen.p1_name_font[4],
	g =      motif.vs_screen.p1_name_font[5],
	b =      motif.vs_screen.p1_name_font[6],
	src =    motif.vs_screen.p1_name_font[7],
	dst =    motif.vs_screen.p1_name_font[8],
})
local txt_p2NameVS = text:create({
	font =   motif.font_data[motif.vs_screen.p2_name_font[1]],
	bank =   motif.vs_screen.p2_name_font[2],
	align =  motif.vs_screen.p2_name_font[3],
	text =   '',
	x =      0,
	y =      0,
	scaleX = motif.vs_screen.p2_name_font_scale[1],
	scaleY = motif.vs_screen.p2_name_font_scale[2],
	r =      motif.vs_screen.p2_name_font[4],
	g =      motif.vs_screen.p2_name_font[5],
	b =      motif.vs_screen.p2_name_font[6],
	src =    motif.vs_screen.p2_name_font[7],
	dst =    motif.vs_screen.p2_name_font[8],
})
local txt_matchNo = text:create({
	font =   motif.font_data[motif.vs_screen.match_font[1]],
	bank =   motif.vs_screen.match_font[2],
	align =  motif.vs_screen.match_font[3],
	text =   '',
	x =      motif.vs_screen.match_offset[1],
	y =      motif.vs_screen.match_offset[2],
	scaleX = motif.vs_screen.match_font_scale[1],
	scaleY = motif.vs_screen.match_font_scale[2],
	r =      motif.vs_screen.match_font[4],
	g =      motif.vs_screen.match_font[5],
	b =      motif.vs_screen.match_font[6],
	src =    motif.vs_screen.match_font[7],
	dst =    motif.vs_screen.match_font[8],
})

function start.f_selectChar(player, t)
	for i = 1, #t do
		selectChar(player, t[i].ref, t[i].pal)
	end
end

function start.f_selectVersus()
	if not main.versusScreen or not main.t_charparam.vsscreen or (main.t_charparam.rivals and start.f_rivalsMatch('vsscreen', 0)) or main.t_selChars[t_p1Selected[1].ref + 1].vsscreen == 0 then
		start.f_selectChar(1, t_p1Selected)
		start.f_selectChar(2, t_p2Selected)
		return true
	else
		local text = main.f_extractText(motif.vs_screen.match_text, matchNo)
		txt_matchNo:update({text = text[1]})
		main.f_bgReset(motif.versusbgdef.bg)
		main.f_playBGM(true, motif.music.vs_bgm, motif.music.vs_bgm_loop, motif.music.vs_bgm_volume, motif.music.vs_bgm_loopstart, motif.music.vs_bgm_loopend)
		local p1Confirmed = false
		local p2Confirmed = false
		local p1Row = 1
		local p2Row = 1
		local t_tmp = {}
		local t_p1_slide_dist = {0, 0}
		local t_p2_slide_dist = {0, 0}
		local orderTime = 0
		if main.p1In == 1 and main.p2In == 2 and (#t_p1Selected > 1 or #t_p2Selected > 1) and not main.coop then
			orderTime = math.max(#t_p1Selected, #t_p2Selected) - 1 * motif.vs_screen.time_order
			if #t_p1Selected == 1 then
				start.f_selectChar(1, t_p1Selected)
				p1Confirmed = true
			end
			if #t_p2Selected == 1 then
				start.f_selectChar(2, t_p2Selected)
				p2Confirmed = true
			end
		elseif #t_p1Selected > 1 and not main.coop then
			orderTime = #t_p1Selected - 1 * motif.vs_screen.time_order
		else
			start.f_selectChar(1, t_p1Selected)
			p1Confirmed = true
			start.f_selectChar(2, t_p2Selected)
			p2Confirmed = true
		end
		--main.f_cmdInput()
		main.fadeStart = getFrameCount()
		local counter = 0 - motif.vs_screen.fadein_time
		fadeType = 'fadein'
		while true do
			if counter == motif.vs_screen.stage_time then
				start.f_playWave(stageNo, 'stage', motif.vs_screen.stage_snd[1], motif.vs_screen.stage_snd[2])
			end
			if esc() then
				--main.f_cmdInput()
				return nil
			elseif p1Confirmed and p2Confirmed then
				if fadeType == 'fadein' and (counter >= motif.vs_screen.time or main.input({1}, {'pal'})) then
					main.fadeStart = getFrameCount()
					fadeType = 'fadeout'
				end
			elseif counter >= motif.vs_screen.time + orderTime then
				if not p1Confirmed then
					start.f_selectChar(1, t_p1Selected)
					p1Confirmed = true
				end
				if not p2Confirmed then
					start.f_selectChar(2, t_p2Selected)
					p2Confirmed = true
				end
			else
				--if Player1 has not confirmed the order yet
				if not p1Confirmed then
					if main.input({1}, {'pal'}) then
						if not p1Confirmed then
							sndPlay(motif.files.snd_data, motif.vs_screen.p1_cursor_done_snd[1], motif.vs_screen.p1_cursor_done_snd[2])
							start.f_selectChar(1, t_p1Selected)
							p1Confirmed = true
						end
						if main.p2In ~= 2 then
							if not p2Confirmed then
								start.f_selectChar(2, t_p2Selected)
								p2Confirmed = true
							end
						end
					elseif main.input({1}, {'$U'}) then
						if #t_p1Selected > 1 then
							sndPlay(motif.files.snd_data, motif.vs_screen.p1_cursor_move_snd[1], motif.vs_screen.p1_cursor_move_snd[2])
							p1Row = p1Row - 1
							if p1Row == 0 then p1Row = #t_p1Selected end
						end
					elseif main.input({1}, {'$D'}) then
						if #t_p1Selected > 1 then
							sndPlay(motif.files.snd_data, motif.vs_screen.p1_cursor_move_snd[1], motif.vs_screen.p1_cursor_move_snd[2])
							p1Row = p1Row + 1
							if p1Row > #t_p1Selected then p1Row = 1 end
						end
					elseif main.input({1}, {'$B'}) then
						if p1Row - 1 > 0 then
							sndPlay(motif.files.snd_data, motif.vs_screen.p1_cursor_move_snd[1], motif.vs_screen.p1_cursor_move_snd[2])
							p1Row = p1Row - 1
							t_tmp = {}
							t_tmp[p1Row] = t_p1Selected[p1Row + 1]
							for i = 1, #t_p1Selected do
								for j = 1, #t_p1Selected do
									if t_tmp[j] == nil and i ~= p1Row + 1 then
										t_tmp[j] = t_p1Selected[i]
										break
									end
								end
							end
							t_p1Selected = t_tmp
						end
					elseif main.input({1}, {'$F'}) then
						if p1Row + 1 <= #t_p1Selected then
							sndPlay(motif.files.snd_data, motif.vs_screen.p1_cursor_move_snd[1], motif.vs_screen.p1_cursor_move_snd[2])
							p1Row = p1Row + 1
							t_tmp = {}
							t_tmp[p1Row] = t_p1Selected[p1Row - 1]
							for i = 1, #t_p1Selected do
								for j = 1, #t_p1Selected do
									if t_tmp[j] == nil and i ~= p1Row - 1 then
										t_tmp[j] = t_p1Selected[i]
										break
									end
								end
							end
							t_p1Selected = t_tmp
						end
					end
				end
				--if Player2 has not confirmed the order yet and is not controlled by Player1
				if not p2Confirmed and main.p2In ~= 1 then
					if main.input({2}, {'pal'}) then
						if not p2Confirmed then
							sndPlay(motif.files.snd_data, motif.vs_screen.p2_cursor_done_snd[1], motif.vs_screen.p2_cursor_done_snd[2])
							start.f_selectChar(2, t_p2Selected)
							p2Confirmed = true
						end
					elseif main.input({2}, {'$U'}) then
						if #t_p2Selected > 1 then
							sndPlay(motif.files.snd_data, motif.vs_screen.p2_cursor_move_snd[1], motif.vs_screen.p2_cursor_move_snd[2])
							p2Row = p2Row - 1
							if p2Row == 0 then p2Row = #t_p2Selected end
						end
					elseif main.input({2}, {'$D'}) then
						if #t_p2Selected > 1 then
							sndPlay(motif.files.snd_data, motif.vs_screen.p2_cursor_move_snd[1], motif.vs_screen.p2_cursor_move_snd[2])
							p2Row = p2Row + 1
							if p2Row > #t_p2Selected then p2Row = 1 end
						end
					elseif main.input({2}, {'$B'}) then
						if p2Row + 1 <= #t_p2Selected then
							sndPlay(motif.files.snd_data, motif.vs_screen.p2_cursor_move_snd[1], motif.vs_screen.p2_cursor_move_snd[2])
							p2Row = p2Row + 1
							t_tmp = {}
							t_tmp[p2Row] = t_p2Selected[p2Row - 1]
							for i = 1, #t_p2Selected do
								for j = 1, #t_p2Selected do
									if t_tmp[j] == nil and i ~= p2Row - 1 then
										t_tmp[j] = t_p2Selected[i]
										break
									end
								end
							end
							t_p2Selected = t_tmp
						end
					elseif main.input({2}, {'$F'}) then
						if p2Row - 1 > 0 then
							sndPlay(motif.files.snd_data, motif.vs_screen.p2_cursor_move_snd[1], motif.vs_screen.p2_cursor_move_snd[2])
							p2Row = p2Row - 1
							t_tmp = {}
							t_tmp[p2Row] = t_p2Selected[p2Row + 1]
							for i = 1, #t_p2Selected do
								for j = 1, #t_p2Selected do
									if t_tmp[j] == nil and i ~= p2Row + 1 then
										t_tmp[j] = t_p2Selected[i]
										break
									end
								end
							end
							t_p2Selected = t_tmp
						end
					end
				end
			end
			counter = counter + 1
			--draw clearcolor
			clearColor(motif.versusbgdef.bgclearcolor[1], motif.versusbgdef.bgclearcolor[2], motif.versusbgdef.bgclearcolor[3])
			--draw layerno = 0 backgrounds
			bgDraw(motif.versusbgdef.bg, false)
			--draw p1 portraits
			local t_portrait = {}
			for i = #t_p1Selected, 1, -1 do
				if #t_portrait < motif.vs_screen.p1_num then
					table.insert(t_portrait, t_p1Selected[i].ref)
				end
			end
			t_portrait = main.f_reversedTable(t_portrait)
			for i = #t_portrait, 1, -1 do
				for j = 1, 2 do
					if t_p1_slide_dist[j] < motif.vs_screen['p1_c' .. i .. '_slide_dist'][j] then
						t_p1_slide_dist[j] = math.min(t_p1_slide_dist[j] + motif.vs_screen['p1_c' .. i .. '_slide_speed'][j], motif.vs_screen['p1_c' .. i .. '_slide_dist'][j])
					end
				end
				drawVersusPortrait(
					t_portrait[i],
					motif.vs_screen.p1_pos[1] + motif.vs_screen.p1_offset[1] + motif.vs_screen['p1_c' .. i .. '_offset'][1] + (i - 1) * motif.vs_screen.p1_spacing[1] + start.f_alignOffset(motif.vs_screen.p1_facing) + math.floor(t_p1_slide_dist[1] + 0.5),
					motif.vs_screen.p1_pos[2] + motif.vs_screen.p1_offset[2] + motif.vs_screen['p1_c' .. i .. '_offset'][2] + (i - 1) * motif.vs_screen.p1_spacing[2] +  math.floor(t_p1_slide_dist[2] + 0.5),
					motif.vs_screen.p1_facing * motif.vs_screen.p1_scale[1] * motif.vs_screen['p1_c' .. i .. '_scale'][1],
					motif.vs_screen.p1_scale[2] * motif.vs_screen['p1_c' .. i .. '_scale'][2]
				)
			end
			--draw p2 portraits
			t_portrait = {}
			for i = #t_p2Selected, 1, -1 do
				if #t_portrait < motif.vs_screen.p2_num then
					table.insert(t_portrait, t_p2Selected[i].ref)
				end
			end
			t_portrait = main.f_reversedTable(t_portrait)
			for i = #t_portrait, 1, -1 do
				for j = 1, 2 do
					if t_p2_slide_dist[j] < motif.vs_screen['p2_c' .. i .. '_slide_dist'][j] then
						t_p2_slide_dist[j] = math.min(t_p2_slide_dist[j] + motif.vs_screen['p2_c' .. i .. '_slide_speed'][j], motif.vs_screen['p2_c' .. i .. '_slide_dist'][j])
					end
				end
				drawVersusPortrait(
					t_portrait[i],
					motif.vs_screen.p2_pos[1] + motif.vs_screen.p2_offset[1] + motif.vs_screen['p2_c' .. i .. '_offset'][1] + (i - 1) * motif.vs_screen.p2_spacing[1] + start.f_alignOffset(motif.vs_screen.p2_facing) + math.floor(t_p2_slide_dist[1] + 0.5),
					motif.vs_screen.p2_pos[2] + motif.vs_screen.p2_offset[2] + motif.vs_screen['p2_c' .. i .. '_offset'][2] + (i - 1) * motif.vs_screen.p2_spacing[2] + math.floor(t_p2_slide_dist[2] + 0.5),
					motif.vs_screen.p2_facing * motif.vs_screen.p2_scale[1] * motif.vs_screen['p2_c' .. i .. '_scale'][1],
					motif.vs_screen.p2_scale[2] * motif.vs_screen['p2_c' .. i .. '_scale'][2]
				)
			end
			--draw names
			start.f_drawName(
				t_p1Selected,
				txt_p1NameVS,
				motif.vs_screen.p1_name_font,
				motif.vs_screen.p1_name_pos[1] + motif.vs_screen.p1_name_offset[1],
				motif.vs_screen.p1_name_pos[2] + motif.vs_screen.p1_name_offset[2],
				motif.vs_screen.p1_name_font_scale[1],
				motif.vs_screen.p1_name_font_scale[2],
				motif.vs_screen.p1_name_spacing[1],
				motif.vs_screen.p1_name_spacing[2],
				motif.vs_screen.p1_name_active_font,
				p1Row
			)
			start.f_drawName(
				t_p2Selected,
				txt_p2NameVS,
				motif.vs_screen.p2_name_font,
				motif.vs_screen.p2_name_pos[1] + motif.vs_screen.p2_name_offset[1],
				motif.vs_screen.p2_name_pos[2] + motif.vs_screen.p2_name_offset[2],
				motif.vs_screen.p2_name_font_scale[1],
				motif.vs_screen.p2_name_font_scale[2],
				motif.vs_screen.p2_name_spacing[1],
				motif.vs_screen.p2_name_spacing[2],
				motif.vs_screen.p2_name_active_font,
				p2Row
			)
			--draw match counter
			if matchNo > 0 then
				txt_matchNo:draw()
			end
			--draw layerno = 1 backgrounds
			bgDraw(motif.versusbgdef.bg, true)
			--draw fadein / fadeout
			main.fadeActive = fadeScreen(
				fadeType,
				main.fadeStart,
				motif.vs_screen[fadeType .. '_time'],
				motif.vs_screen[fadeType .. '_col'][1],
				motif.vs_screen[fadeType .. '_col'][2],
				motif.vs_screen[fadeType .. '_col'][3]
			)
			--frame transition
			if main.fadeActive then
				commandBufReset(main.cmd[1])
				commandBufReset(main.cmd[2])
			elseif fadeType == 'fadeout' then
				commandBufReset(main.cmd[1])
				commandBufReset(main.cmd[2])
				clearColor(motif.versusbgdef.bgclearcolor[1], motif.versusbgdef.bgclearcolor[2], motif.versusbgdef.bgclearcolor[3]) --skip last frame rendering
				break
			else
				main.f_cmdInput()
			end
			refresh()
		end
		return true
	end
end

--;===========================================================
--; RESULT SCREEN
--;===========================================================
local txt_winscreen = text:create({
	font =   motif.font_data[motif.win_screen.wintext_font[1]],
	bank =   motif.win_screen.wintext_font[2],
	align =  motif.win_screen.wintext_font[3],
	text =   motif.win_screen.wintext_text,
	x =      motif.win_screen.wintext_offset[1],
	y =      motif.win_screen.wintext_offset[2],
	scaleX = motif.win_screen.wintext_font_scale[1],
	scaleY = motif.win_screen.wintext_font_scale[2],
	r =      motif.win_screen.wintext_font[4],
	g =      motif.win_screen.wintext_font[5],
	b =      motif.win_screen.wintext_font[6],
	src =    motif.win_screen.wintext_font[7],
	dst =    motif.win_screen.wintext_font[8],
})
local txt_resultSurvival = text:create({
	font =   motif.font_data[motif.survival_results_screen.wintext_font[1]],
	bank =   motif.survival_results_screen.wintext_font[2],
	align =  motif.survival_results_screen.wintext_font[3],
	text =   '',
	x =      motif.survival_results_screen.wintext_offset[1],
	y =      motif.survival_results_screen.wintext_offset[2],
	scaleX = motif.survival_results_screen.wintext_font_scale[1],
	scaleY = motif.survival_results_screen.wintext_font_scale[2],
	r =      motif.survival_results_screen.wintext_font[4],
	g =      motif.survival_results_screen.wintext_font[5],
	b =      motif.survival_results_screen.wintext_font[6],
	src =    motif.survival_results_screen.wintext_font[7],
	dst =    motif.survival_results_screen.wintext_font[8],
})
local txt_resultVS100 = text:create({
	font =   motif.font_data[motif.vs100kumite_results_screen.wintext_font[1]],
	bank =   motif.vs100kumite_results_screen.wintext_font[2],
	align =  motif.vs100kumite_results_screen.wintext_font[3],
	text =   '',
	x =      motif.vs100kumite_results_screen.wintext_offset[1],
	y =      motif.vs100kumite_results_screen.wintext_offset[2],
	scaleX = motif.vs100kumite_results_screen.wintext_font_scale[1],
	scaleY = motif.vs100kumite_results_screen.wintext_font_scale[2],
	r =      motif.vs100kumite_results_screen.wintext_font[4],
	g =      motif.vs100kumite_results_screen.wintext_font[5],
	b =      motif.vs100kumite_results_screen.wintext_font[6],
	src =    motif.vs100kumite_results_screen.wintext_font[7],
	dst =    motif.vs100kumite_results_screen.wintext_font[8],
})
local txt_resultTimeAttack = text:create({
	font =   motif.font_data[motif.timeattack_results_screen.wintext_font[1]],
	bank =   motif.timeattack_results_screen.wintext_font[2],
	align =  motif.timeattack_results_screen.wintext_font[3],
	text =   '',
	x =      motif.timeattack_results_screen.wintext_offset[1],
	y =      motif.timeattack_results_screen.wintext_offset[2],
	scaleX = motif.timeattack_results_screen.wintext_font_scale[1],
	scaleY = motif.timeattack_results_screen.wintext_font_scale[2],
	r =      motif.timeattack_results_screen.wintext_font[4],
	g =      motif.timeattack_results_screen.wintext_font[5],
	b =      motif.timeattack_results_screen.wintext_font[6],
	src =    motif.timeattack_results_screen.wintext_font[7],
	dst =    motif.timeattack_results_screen.wintext_font[8],
})
local txt_resultTimeChallenge = text:create({
	font =   motif.font_data[motif.timechallenge_results_screen.wintext_font[1]],
	bank =   motif.timechallenge_results_screen.wintext_font[2],
	align =  motif.timechallenge_results_screen.wintext_font[3],
	text =   '',
	x =      motif.timechallenge_results_screen.wintext_offset[1],
	y =      motif.timechallenge_results_screen.wintext_offset[2],
	scaleX = motif.timechallenge_results_screen.wintext_font_scale[1],
	scaleY = motif.timechallenge_results_screen.wintext_font_scale[2],
	r =      motif.timechallenge_results_screen.wintext_font[4],
	g =      motif.timechallenge_results_screen.wintext_font[5],
	b =      motif.timechallenge_results_screen.wintext_font[6],
	src =    motif.timechallenge_results_screen.wintext_font[7],
	dst =    motif.timechallenge_results_screen.wintext_font[8],
})
local txt_resultScoreChallenge = text:create({
	font =   motif.font_data[motif.scorechallenge_results_screen.wintext_font[1]],
	bank =   motif.scorechallenge_results_screen.wintext_font[2],
	align =  motif.scorechallenge_results_screen.wintext_font[3],
	text =   '',
	x =      motif.scorechallenge_results_screen.wintext_offset[1],
	y =      motif.scorechallenge_results_screen.wintext_offset[2],
	scaleX = motif.scorechallenge_results_screen.wintext_font_scale[1],
	scaleY = motif.scorechallenge_results_screen.wintext_font_scale[2],
	r =      motif.scorechallenge_results_screen.wintext_font[4],
	g =      motif.scorechallenge_results_screen.wintext_font[5],
	b =      motif.scorechallenge_results_screen.wintext_font[6],
	src =    motif.scorechallenge_results_screen.wintext_font[7],
	dst =    motif.scorechallenge_results_screen.wintext_font[8],
})
local txt_resultBossRush = text:create({
	font =   motif.font_data[motif.bossrush_results_screen.wintext_font[1]],
	bank =   motif.bossrush_results_screen.wintext_font[2],
	align =  motif.bossrush_results_screen.wintext_font[3],
	text =   motif.bossrush_results_screen.wintext_text,
	x =      motif.bossrush_results_screen.wintext_offset[1],
	y =      motif.bossrush_results_screen.wintext_offset[2],
	scaleX = motif.bossrush_results_screen.wintext_font_scale[1],
	scaleY = motif.bossrush_results_screen.wintext_font_scale[2],
	r =      motif.bossrush_results_screen.wintext_font[4],
	g =      motif.bossrush_results_screen.wintext_font[5],
	b =      motif.bossrush_results_screen.wintext_font[6],
	src =    motif.bossrush_results_screen.wintext_font[7],
	dst =    motif.bossrush_results_screen.wintext_font[8],
})

local function f_drawTextAtLayerNo(t, t_resultText, txt, layerNo)
	if t.wintext_layerno ~= layerNo then
		return
	end
	for i = 1, #t_resultText do
		txt:update({
			text = t_resultText[i],
			y =    t.wintext_offset[2] + (motif.font_def[t.wintext_font[1]].Size[2] + motif.font_def[t.wintext_font[1]].Spacing[2]) * (i - 1),
		})
		txt:draw()
	end
end

local function f_lowestRankingData(data)
	if stats.modes == nil or stats.modes[gameMode()] == nil or stats.modes[gameMode()].ranking == nil or #stats.modes[gameMode()].ranking < motif.rankings.max_entries then
		if data == 'score' then
			return 0
		else --time
			return 99
		end
	end
	local ret = 0
	for k, v in ipairs(stats.modes[gameMode()].ranking) do
		if k == 1 or (data == 'score' and ret > v[data]) or (data == 'time' and ret < v[data]) then
			ret = v[data]
		end
	end
	return ret
end

function start.f_result(mode)
	if main.resultsTable == nil then
		return false
	end
	local t = main.resultsTable
	local t_resultText = {}
	local txt = ''
	local stateType = ''
	local winBgm = true
	if mode == 'arcade' or mode == 'teamcoop' or mode == 'netplayteamcoop' then
		local tPos = main.t_selChars[t_p1Selected[1].ref + 1]
		if tPos.ending ~= nil and main.f_fileExists(tPos.ending) then --not displayed if the team leader has an ending
			return false
		end
		t_resultText = main.f_extractText(t.wintext_text)
		txt = txt_winscreen
	elseif mode == 'bossrush' then
		if winner ~= 1 then
			return false
		end
		t_resultText = main.f_extractText(t.wintext_text)
		txt = txt_resultBossRush
	elseif mode == 'survival' or mode == 'survivalcoop' or mode == 'netplaysurvivalcoop' then
		t_resultText = main.f_extractText(t.wintext_text, winCnt)
		txt = txt_resultSurvival
		if winCnt < t.roundstowin then
			stateType = '_lose'
			winBgm = false
		else
			stateType = '_win'
		end
	elseif mode == 'vs100kumite' then
		t_resultText = main.f_extractText(t.wintext_text, winCnt, loseCnt)
		txt = txt_resultVS100
		if winCnt < t.roundstowin then
			stateType = '_lose'
			winBgm = false
		else
			stateType = '_win'
		end
	elseif mode == 'timeattack' then
		t_resultText = main.f_extractText(start.f_clearTimeText(t.wintext_text, t_savedData.time.total / 60))
		txt = txt_resultTimeAttack
		if t_gameStats.time / 60 >= f_lowestRankingData('time') then
			stateType = '_lose'
			winBgm = false
		else
			stateType = '_win'
		end
	elseif mode == 'timechallenge' then
		if winner ~= 1 then
			return false
		end
		t_resultText = main.f_extractText(start.f_clearTimeText(t.wintext_text, t_savedData.time.total / 60))
		txt = txt_resultTimeChallenge
		if t_gameStats.time / 60 >= f_lowestRankingData('time') then
			stateType = '_lose'
			winBgm = false
		else
			stateType = '_win'
		end
	elseif mode == 'scorechallenge' then
		if winner ~= 1 then
			return false
		end
		t_resultText = main.f_extractText(t.wintext_text, t_gameStats.p1score)
		txt = txt_resultScoreChallenge
		if t_gameStats.p1score <= f_lowestRankingData('score') then
			stateType = '_lose'
			winBgm = false
		else
			stateType = '_win'
		end
	else
		panicError('LUA ERROR: ' .. gameMode() .. ' game mode unrecognized by start.f_result()')
	end
	main.f_bgReset(motif.resultsbgdef.bg)
	if winBgm then
		main.f_playBGM(false, motif.music.results_bgm, motif.music.results_bgm_loop, motif.music.results_bgm_volume, motif.music.results_bgm_loopstart, motif.music.results_bgm_loopend)
	else
		main.f_playBGM(false, motif.music.results_lose_bgm, motif.music.results_lose_bgm_loop, motif.music.results_lose_bgm_volume, motif.music.results_lose_bgm_loopstart, motif.music.results_lose_bgm_loopend)
	end
	--main.f_cmdInput()
	main.fadeStart = getFrameCount()
	local counter = 0 - t.fadein_time
	fadeType = 'fadein'
	for i = 1, 2 do
		for k, v in ipairs(t['p' .. i .. '_statedef' .. stateType]) do
			if charChangeState(i, v) then
				break
			end
		end
	end
	while true do
		if esc() then
			lastMatchClear()
			--main.f_cmdInput()
			return nil
		elseif fadeType == 'fadein' and (counter >= t.pose_time or main.input({1}, {'pal'})) then
			main.fadeStart = getFrameCount()
			fadeType = 'fadeout'
		end
		counter = counter + 1
		--draw clearcolor
		clearColor(motif.resultsbgdef.bgclearcolor[1], motif.resultsbgdef.bgclearcolor[2], motif.resultsbgdef.bgclearcolor[3])
		--draw previous match data
		lastMatchRender(false)
		--draw menu box
		fillRect(
			t.boxbg_coords[1],
			t.boxbg_coords[2],
			t.boxbg_coords[3] - t.boxbg_coords[1] + 1,
			t.boxbg_coords[4] - t.boxbg_coords[2] + 1,
			t.boxbg_col[1],
			t.boxbg_col[2],
			t.boxbg_col[3],
			t.boxbg_alpha[1],
			t.boxbg_alpha[2],
			false
		)
		--draw text at layerno = 0
		f_drawTextAtLayerNo(t, t_resultText, txt, 0)
		--draw layerno = 0 backgrounds
		bgDraw(motif.resultsbgdef.bg, false)
		--draw text at layerno = 1
		f_drawTextAtLayerNo(t, t_resultText, txt, 1)		
		--draw layerno = 1 backgrounds
		bgDraw(motif.resultsbgdef.bg, true)
		--draw text at layerno = 1
		f_drawTextAtLayerNo(t, t_resultText, txt, 2)
		--draw fadein / fadeout
		main.fadeActive = fadeScreen(
			fadeType,
			main.fadeStart,
			t[fadeType .. '_time'],
			t[fadeType .. '_col'][1],
			t[fadeType .. '_col'][2],
			t[fadeType .. '_col'][3]
		)
		--frame transition
		if main.fadeActive then
			commandBufReset(main.cmd[1])
		elseif fadeType == 'fadeout' then
			commandBufReset(main.cmd[1])
			clearColor(motif.resultsbgdef.bgclearcolor[1], motif.resultsbgdef.bgclearcolor[2], motif.resultsbgdef.bgclearcolor[3]) --skip last frame rendering
			break
		else
			main.f_cmdInput()
		end
		refresh()
	end
	lastMatchClear()
	return true
end

--;===========================================================
--; VICTORY SCREEN
--;===========================================================
local txt_winquote = text:create({
	font =   motif.font_data[motif.victory_screen.winquote_font[1]],
	bank =   motif.victory_screen.winquote_font[2],
	align =  motif.victory_screen.winquote_font[3],
	text =   '',
	x =      0,
	y =      0,
	scaleX = motif.victory_screen.winquote_font_scale[1],
	scaleY = motif.victory_screen.winquote_font_scale[2],
	r =      motif.victory_screen.winquote_font[4],
	g =      motif.victory_screen.winquote_font[5],
	b =      motif.victory_screen.winquote_font[6],
	src =    motif.victory_screen.winquote_font[7],
	dst =    motif.victory_screen.winquote_font[8],
})
local txt_p1_winquoteName = text:create({
	font =   motif.font_data[motif.victory_screen.p1_name_font[1]],
	bank =   motif.victory_screen.p1_name_font[2],
	align =  motif.victory_screen.p1_name_font[3],
	text =   '',
	x =      motif.victory_screen.p1_name_offset[1],
	y =      motif.victory_screen.p1_name_offset[2],
	scaleX = motif.victory_screen.p1_name_font_scale[1],
	scaleY = motif.victory_screen.p1_name_font_scale[2],
	r =      motif.victory_screen.p1_name_font[4],
	g =      motif.victory_screen.p1_name_font[5],
	b =      motif.victory_screen.p1_name_font[6],
	src =    motif.victory_screen.p1_name_font[7],
	dst =    motif.victory_screen.p1_name_font[8],
})
local txt_p2_winquoteName = text:create({
	font =   motif.font_data[motif.victory_screen.p2_name_font[1]],
	bank =   motif.victory_screen.p2_name_font[2],
	align =  motif.victory_screen.p2_name_font[3],
	text =   '',
	x =      motif.victory_screen.p2_name_offset[1],
	y =      motif.victory_screen.p2_name_offset[2],
	scaleX = motif.victory_screen.p2_name_font_scale[1],
	scaleY = motif.victory_screen.p2_name_font_scale[2],
	r =      motif.victory_screen.p2_name_font[4],
	g =      motif.victory_screen.p2_name_font[5],
	b =      motif.victory_screen.p2_name_font[6],
	src =    motif.victory_screen.p2_name_font[7],
	dst =    motif.victory_screen.p2_name_font[8],
})

function start.f_teamOrder(teamNo, allow_ko, num)
	local allow_ko = allow_ko or 0
	local t = {}
	local playerNo = -1
	local selectNo = -1
	local ok = false
	for k, v in ipairs(t_gameStats.match[t_gameStats.lastRound]) do --loop through all last round participants
		if k % 2 ~= teamNo then --only if character belongs to selected team
			if v.win then --win team
				if not v.ko and not ok then --first not KOed win team member
					playerNo = k
					selectNo = v.selectNo
					if #t >= num then break end
					table.insert(t, {['pn'] = k, ['ref'] = v.selectNo})
					ok = true
				elseif not v.ko or allow_ko == 1 then --other win team members
					if #t >= num then break end
					table.insert(t, {['pn'] = k, ['ref'] = v.selectNo})
				end
			else --lose team
				if not ok then
					playerNo = k
					selectNo = v.selectNo
					ok = true
				end
				if #t >= num then break end
				table.insert(t, {['pn'] = k, ['ref'] = v.selectNo})
			end
		end
	end
	return playerNo, selectNo, t
end

function start.f_selectVictory()
	if winner < 1 or not main.victoryScreen or motif.victory_screen.enabled == 0 then
		return false
	elseif gameMode('versus') or gameMode('netplayversus') then
		if motif.victory_screen.vs_enabled == 0 then
			return false
		end
	elseif winner == 2 and motif.victory_screen.cpu_enabled == 0 then
		return false
	end
	local t_p1_slide_dist = {0, 0}
	local t_p2_slide_dist = {0, 0}
	local winnerNo = -1
	local winnerRef = -1
	local loserNo = -1
	local loserRef = -1
	local t = {}
	local t2 = {}
	for i = 0, 1 do
		if i == t_gameStats.winTeam then
			winnerNo, winnerRef, t = start.f_teamOrder(i, motif.victory_screen.winner_teamko_enabled, motif.victory_screen.p1_num)
		else
			loserNo, loserRef, t2 = start.f_teamOrder(i, true, motif.victory_screen.p2_num)
		end
	end
	if winnerNo == -1 or winnerRef == -1 then
		return false
	elseif not main.t_charparam.winscreen then
		return false
	elseif main.t_charparam.rivals and start.f_rivalsMatch('winscreen', 0) then --winscreen assigned as rivals param
		return false
	elseif main.t_selChars[winnerRef + 1].winscreen == 0 then --winscreen assigned as character param
		return false
	end
	main.f_bgReset(motif.victorybgdef.bg)
	if not t_victoryBGM[t_gameStats.winTeam + 1] then
		main.f_playBGM(false, motif.music.victory_bgm, motif.music.victory_bgm_loop, motif.music.victory_bgm_volume, motif.music.victory_bgm_loopstart, motif.music.victory_bgm_loopend)
	end
	local winquote = getCharVictoryQuote(winnerNo)
	if winquote == '' then
		winquote = motif.victory_screen.winquote_text
	end
	txt_p1_winquoteName:update({text = start.f_getName(winnerRef)})
	txt_p2_winquoteName:update({text = start.f_getName(loserRef)})
	local cnt = 0
	--main.f_cmdInput()
	main.fadeStart = getFrameCount()
	local counter = 0 - motif.victory_screen.fadein_time
	local finishedText = false
	fadeType = 'fadein'
	while true do
		if esc() then
			lastMatchClear()
			--main.f_cmdInput()
			return nil
		elseif fadeType == 'fadein' and (counter >= motif.victory_screen.time or main.input({1}, {'pal'})) then
			main.fadeStart = getFrameCount()
			fadeType = 'fadeout'
		end
		if finishedText then
			counter = counter + 1
		end
		--draw clearcolor
		clearColor(motif.victorybgdef.bgclearcolor[1], motif.victorybgdef.bgclearcolor[2], motif.victorybgdef.bgclearcolor[3])
		--draw previous match data
		lastMatchRender(false)
		--draw menu box
		fillRect(
			motif.victory_screen.boxbg_coords[1],
			motif.victory_screen.boxbg_coords[2],
			motif.victory_screen.boxbg_coords[3] - motif.victory_screen.boxbg_coords[1] + 1,
			motif.victory_screen.boxbg_coords[4] - motif.victory_screen.boxbg_coords[2] + 1,
			motif.victory_screen.boxbg_col[1],
			motif.victory_screen.boxbg_col[2],
			motif.victory_screen.boxbg_col[3],
			motif.victory_screen.boxbg_alpha[1],
			motif.victory_screen.boxbg_alpha[2],
			false
		)
		--draw layerno = 0 backgrounds
		bgDraw(motif.victorybgdef.bg, false)
		--draw portraits
		-- loser team portraits
		for i = #t2, 1, -1 do
			for j = 1, 2 do
				if t_p2_slide_dist[j] < motif.victory_screen['p2_c' .. i .. '_slide_dist'][j] then
					t_p2_slide_dist[j] = math.min(t_p2_slide_dist[j] + motif.victory_screen['p2_c' .. i .. '_slide_speed'][j], motif.victory_screen['p2_c' .. i .. '_slide_dist'][j])
				end
			end
			drawCharSprite(
				t2[i].pn,
				{
					motif.victory_screen['p2_c' .. i .. '_spr'][1], motif.victory_screen['p2_c' .. i .. '_spr'][2],
					motif.victory_screen.p2_spr[1], motif.victory_screen.p2_spr[2],
					9000, 1
				},
				motif.victory_screen.p2_pos[1] + motif.victory_screen.p2_offset[1] + motif.victory_screen['p2_c' .. i .. '_offset'][1] + math.floor(t_p2_slide_dist[1] + 0.5),
				motif.victory_screen.p2_pos[2] + motif.victory_screen.p2_offset[2] + motif.victory_screen['p2_c' .. i .. '_offset'][2] + math.floor(t_p2_slide_dist[2] + 0.5),
				motif.victory_screen.p2_scale[1] * motif.victory_screen['p2_c' .. i .. '_scale'][1],
				motif.victory_screen.p2_scale[2] * motif.victory_screen['p2_c' .. i .. '_scale'][2],
				motif.victory_screen.p2_facing
			)
		end
		-- winner team portraits
		for i = #t, 1, -1 do
			for j = 1, 2 do
				if t_p1_slide_dist[j] < motif.victory_screen['p1_c' .. i .. '_slide_dist'][j] then
					t_p1_slide_dist[j] = math.min(t_p1_slide_dist[j] + motif.victory_screen['p1_c' .. i .. '_slide_speed'][j], motif.victory_screen['p1_c' .. i .. '_slide_dist'][j])
				end
			end
			drawCharSprite(
				t[i].pn,
				{
					motif.victory_screen['p1_c' .. i .. '_spr'][1], motif.victory_screen['p1_c' .. i .. '_spr'][2],
					motif.victory_screen.p1_spr[1], motif.victory_screen.p1_spr[2],
					9000, 1
				},
				motif.victory_screen.p1_pos[1] + motif.victory_screen.p1_offset[1] + motif.victory_screen['p1_c' .. i .. '_offset'][1] + math.floor(t_p1_slide_dist[1] + 0.5),
				motif.victory_screen.p1_pos[2] + motif.victory_screen.p1_offset[2] + motif.victory_screen['p1_c' .. i .. '_offset'][2] + math.floor(t_p1_slide_dist[2] + 0.5),
				motif.victory_screen.p1_scale[1] * motif.victory_screen['p1_c' .. i .. '_scale'][1],
				motif.victory_screen.p1_scale[2] * motif.victory_screen['p1_c' .. i .. '_scale'][2],
				motif.victory_screen.p1_facing
			)
		end
		--draw winner name
		txt_p1_winquoteName:draw()
		--draw loser name
		if motif.victory_screen.loser_name_enabled == 1 then
			txt_p2_winquoteName:draw()
		end
		--draw winquote
		cnt = cnt + 1
		local pxLimit = motif.victory_screen.winquote_length
		if pxLimit == 0 and motif.victory_screen.winquote_textwrap:match('[wl]') then
			pxLimit = main.f_pxLimit(motif.victory_screen.winquote_offset[1], motif.info.localcoord[1], motif.victory_screen.winquote_font[3])
		end
		finishedText = main.f_textRender(
			txt_winquote,
			winquote,
			cnt,
			motif.victory_screen.winquote_offset[1],
			motif.victory_screen.winquote_offset[2],
			motif.font_def[motif.victory_screen.winquote_font[1]],
			motif.victory_screen.winquote_delay,
			pxLimit
		)
		--draw layerno = 1 backgrounds
		bgDraw(motif.victorybgdef.bg, true)
		--draw fadein / fadeout
		main.fadeActive = fadeScreen(
			fadeType,
			main.fadeStart,
			motif.victory_screen[fadeType .. '_time'],
			motif.victory_screen[fadeType .. '_col'][1],
			motif.victory_screen[fadeType .. '_col'][2],
			motif.victory_screen[fadeType .. '_col'][3]
		)
		--frame transition
		if main.fadeActive then
			commandBufReset(main.cmd[1])
		elseif fadeType == 'fadeout' then
			commandBufReset(main.cmd[1])
			clearColor(motif.victorybgdef.bgclearcolor[1], motif.victorybgdef.bgclearcolor[2], motif.victorybgdef.bgclearcolor[3]) --skip last frame rendering
			break
		else
			main.f_cmdInput()
		end
		refresh()
	end
	lastMatchClear()
	return true
end

--;===========================================================
--; CONTINUE SCREEN
--;===========================================================
local txt_credits = text:create({
	font =   motif.font_data[motif.continue_screen.credits_font[1]],
	bank =   motif.continue_screen.credits_font[2],
	align =  motif.continue_screen.credits_font[3],
	text =   '',
	x =      motif.continue_screen.credits_offset[1],
	y =      motif.continue_screen.credits_offset[2],
	scaleX = motif.continue_screen.credits_font_scale[1],
	scaleY = motif.continue_screen.credits_font_scale[2],
	r =      motif.continue_screen.credits_font[4],
	g =      motif.continue_screen.credits_font[5],
	b =      motif.continue_screen.credits_font[6],
	src =    motif.continue_screen.credits_font[7],
	dst =    motif.continue_screen.credits_font[8],
})
local txt_continue = text:create({
	font =   motif.font_data[motif.continue_screen.continue_font[1]],
	bank =   motif.continue_screen.continue_font[2],
	align =  motif.continue_screen.continue_font[3],
	text =   motif.continue_screen.continue_text,
	x =      motif.continue_screen.pos[1] + motif.continue_screen.continue_offset[1],
	y =      motif.continue_screen.pos[2] + motif.continue_screen.continue_offset[2],
	scaleX = motif.continue_screen.continue_font_scale[1],
	scaleY = motif.continue_screen.continue_font_scale[2],
	r =      motif.continue_screen.continue_font[4],
	g =      motif.continue_screen.continue_font[5],
	b =      motif.continue_screen.continue_font[6],
	src =    motif.continue_screen.continue_font[7],
	dst =    motif.continue_screen.continue_font[8],
})
local txt_yes = text:create({})
local txt_no = text:create({})

function start.f_continue()
	if motif.continue_screen.enabled == 0 then
		return false
	end
	main.f_bgReset(motif.continuebgdef.bg)
	main.f_playBGM(false, motif.music.continue_bgm, motif.music.continue_bgm_loop, motif.music.continue_bgm_volume, motif.music.continue_bgm_loopstart, motif.music.continue_bgm_loopend)
	--animReset(motif.continue_screen.continue_anim_data)
	--animUpdate(motif.continue_screen.continue_anim_data)
	local continue = false
	local text = main.f_extractText(motif.continue_screen.credits_text, main.credits)
	txt_credits:update({text = text[1]})
	--main.f_cmdInput()
	main.fadeStart = getFrameCount()
	local counter = 0-- - motif.victory_screen.fadein_time
	local yesActive = true
	fadeType = 'fadein'
	for i = 1, 2 do
		for k, v in ipairs(motif.continue_screen['p' .. i .. '_statedef_continue']) do
			if charChangeState(i, v) then
				break
			end
		end
	end
	while true do
		--draw clearcolor
		clearColor(motif.continuebgdef.bgclearcolor[1], motif.continuebgdef.bgclearcolor[2], motif.continuebgdef.bgclearcolor[3])
		--draw previous match data
		lastMatchRender(false)
		--draw menu box
		fillRect(
			motif.continue_screen.boxbg_coords[1],
			motif.continue_screen.boxbg_coords[2],
			motif.continue_screen.boxbg_coords[3] - motif.continue_screen.boxbg_coords[1] + 1,
			motif.continue_screen.boxbg_coords[4] - motif.continue_screen.boxbg_coords[2] + 1,
			motif.continue_screen.boxbg_col[1],
			motif.continue_screen.boxbg_col[2],
			motif.continue_screen.boxbg_col[3],
			motif.continue_screen.boxbg_alpha[1],
			motif.continue_screen.boxbg_alpha[2],
			false
		)
		--draw layerno = 0 backgrounds
		bgDraw(motif.continuebgdef.bg, false)
		--continue screen state
		if esc() then
			lastMatchClear()
			--main.f_cmdInput()
			return nil
		elseif fadeType == 'fadein' and (counter > motif.continue_screen.endtime or continue) then
			main.fadeStart = getFrameCount()
			fadeType = 'fadeout'
		elseif motif.continue_screen.animated_continue == 1 then --advanced continue screen parameters
			if counter < motif.continue_screen.continue_end_skiptime then
				if main.input({1}, {'s'}) then
					continue = true
					for i = 1, 2 do
						for k, v in ipairs(motif.continue_screen['p' .. i .. '_statedef_yes']) do
							if charChangeState(i, v) then
								break
							end
						end
					end
					main.credits = main.credits - 1
					text = main.f_extractText(motif.continue_screen.credits_text, main.credits)
					txt_credits:update({text = text[1]})
				elseif main.input({1}, {'pal'}) and counter >= motif.continue_screen.continue_starttime + motif.continue_screen.continue_skipstart then
					local cnt = 0
					if counter < motif.continue_screen.continue_9_skiptime then
						cnt = motif.continue_screen.continue_9_skiptime
					elseif counter <= motif.continue_screen.continue_8_skiptime then
						cnt = motif.continue_screen.continue_8_skiptime
					elseif counter < motif.continue_screen.continue_7_skiptime then
						cnt = motif.continue_screen.continue_7_skiptime
					elseif counter < motif.continue_screen.continue_6_skiptime then
						cnt = motif.continue_screen.continue_6_skiptime
					elseif counter < motif.continue_screen.continue_5_skiptime then
						cnt = motif.continue_screen.continue_5_skiptime
					elseif counter < motif.continue_screen.continue_4_skiptime then
						cnt = motif.continue_screen.continue_4_skiptime
					elseif counter < motif.continue_screen.continue_3_skiptime then
						cnt = motif.continue_screen.continue_3_skiptime
					elseif counter < motif.continue_screen.continue_2_skiptime then
						cnt = motif.continue_screen.continue_2_skiptime
					elseif counter < motif.continue_screen.continue_1_skiptime then
						cnt = motif.continue_screen.continue_1_skiptime
					elseif counter < motif.continue_screen.continue_0_skiptime then
						cnt = motif.continue_screen.continue_0_skiptime
					end
					while counter < cnt do
						counter = counter + 1
						animUpdate(motif.continue_screen.continue_anim_data)
					end
				end
				if counter == motif.continue_screen.continue_9_skiptime then
					sndPlay(motif.files.snd_data, motif.continue_screen.continue_9_snd[1], motif.continue_screen.continue_9_snd[2])
				elseif counter == motif.continue_screen.continue_8_skiptime then
					sndPlay(motif.files.snd_data, motif.continue_screen.continue_8_snd[1], motif.continue_screen.continue_8_snd[2])
				elseif counter == motif.continue_screen.continue_7_skiptime then
					sndPlay(motif.files.snd_data, motif.continue_screen.continue_7_snd[1], motif.continue_screen.continue_7_snd[2])
				elseif counter == motif.continue_screen.continue_6_skiptime then
					sndPlay(motif.files.snd_data, motif.continue_screen.continue_6_snd[1], motif.continue_screen.continue_6_snd[2])
				elseif counter == motif.continue_screen.continue_5_skiptime then
					sndPlay(motif.files.snd_data, motif.continue_screen.continue_5_snd[1], motif.continue_screen.continue_5_snd[2])
				elseif counter == motif.continue_screen.continue_4_skiptime then
					sndPlay(motif.files.snd_data, motif.continue_screen.continue_4_snd[1], motif.continue_screen.continue_4_snd[2])
				elseif counter == motif.continue_screen.continue_3_skiptime then
					sndPlay(motif.files.snd_data, motif.continue_screen.continue_3_snd[1], motif.continue_screen.continue_3_snd[2])
				elseif counter == motif.continue_screen.continue_2_skiptime then
					sndPlay(motif.files.snd_data, motif.continue_screen.continue_2_snd[1], motif.continue_screen.continue_2_snd[2])
				elseif counter == motif.continue_screen.continue_1_skiptime then
					sndPlay(motif.files.snd_data, motif.continue_screen.continue_1_snd[1], motif.continue_screen.continue_1_snd[2])
				elseif counter == motif.continue_screen.continue_0_skiptime then
					sndPlay(motif.files.snd_data, motif.continue_screen.continue_0_snd[1], motif.continue_screen.continue_0_snd[2])
				end
			elseif counter == motif.continue_screen.continue_end_skiptime then
				playBGM(motif.music.continue_end_bgm, true, motif.music.continue_end_bgm_loop, motif.music.continue_end_bgm_volume, motif.music.continue_end_bgm_loopstart, motif.music.continue_end_bgm_loopend)
				sndPlay(motif.files.snd_data, motif.continue_screen.continue_end_snd[1], motif.continue_screen.continue_end_snd[2])
				for i = 1, 2 do
					for k, v in ipairs(motif.continue_screen['p' .. i .. '_statedef_no']) do
						if charChangeState(i, v) then
							break
						end
					end
				end
			end
			--draw credits text
			if counter >= motif.continue_screen.continue_skipstart then --show when counter starts counting down
				txt_credits:draw()
			end
			counter = counter + 1
			--draw counter
			animUpdate(motif.continue_screen.continue_anim_data)
			animDraw(motif.continue_screen.continue_anim_data)
		else --vanilla mugen continue screen parameters
			if main.input({1}, {'$F', '$B'}) then
				sndPlay(motif.files.snd_data, motif.continue_screen.move_snd[1], motif.continue_screen.move_snd[2])
				if yesActive then
					yesActive = false
				else
					yesActive = true
				end
			elseif main.input({1}, {'pal', 's'}) then
				continue = yesActive
				if continue then
					sndPlay(motif.files.snd_data, motif.continue_screen.done_snd[1], motif.continue_screen.done_snd[2])
					for i = 1, 2 do
						for k, v in ipairs(motif.continue_screen['p' .. i .. '_statedef_yes']) do
							if charChangeState(i, v) then
								break
							end
						end
					end
					main.credits = main.credits - 1
					--text = main.f_extractText(motif.continue_screen.credits_text, main.credits)
					--txt_credits:update({text = text[1]})
				else
					sndPlay(motif.files.snd_data, motif.continue_screen.cancel_snd[1], motif.continue_screen.cancel_snd[2])
					for i = 1, 2 do
						for k, v in ipairs(motif.continue_screen['p' .. i .. '_statedef_no']) do
							if charChangeState(i, v) then
								break
							end
						end
					end
				end
				counter = motif.continue_screen.endtime + 1
			end
			txt_continue:draw()
			for i = 1, 2 do
				local txt = ''
				local var = ''
				if i == 1 then
					txt = txt_yes
					if yesActive then
						var = 'yes_active'
					else
						var = 'yes'
					end
				else
					txt = txt_no
					if yesActive then
						var = 'no'
					else
						var = 'no_active'
					end
				end
				txt:update({
					font =   motif.font_data[motif.continue_screen[var .. '_font'][1]],
					bank =   motif.continue_screen[var .. '_font'][2],
					align =  motif.continue_screen[var .. '_font'][3],
					text =   motif.continue_screen[var .. '_text'],
					x =      motif.continue_screen.pos[1] + motif.continue_screen[var .. '_offset'][1],
					y =      motif.continue_screen.pos[2] + motif.continue_screen[var .. '_offset'][2],
					scaleX = motif.continue_screen[var .. '_font_scale'][1],
					scaleY = motif.continue_screen[var .. '_font_scale'][2],
					r =      motif.continue_screen[var .. '_font'][4],
					g =      motif.continue_screen[var .. '_font'][5],
					b =      motif.continue_screen[var .. '_font'][6],
					src =    motif.continue_screen[var .. '_font'][7],
					dst =    motif.continue_screen[var .. '_font'][8],
				})
				txt:draw()
			end
		end
		--draw layerno = 1 backgrounds
		bgDraw(motif.continuebgdef.bg, true)
		--draw fadein / fadeout
		main.fadeActive = fadeScreen(
			fadeType,
			main.fadeStart,
			motif.continue_screen[fadeType .. '_time'],
			motif.continue_screen[fadeType .. '_col'][1],
			motif.continue_screen[fadeType .. '_col'][2],
			motif.continue_screen[fadeType .. '_col'][3]
		)
		--frame transition
		if main.fadeActive then
			commandBufReset(main.cmd[1])
		elseif fadeType == 'fadeout' then
			commandBufReset(main.cmd[1])
			clearColor(motif.continuebgdef.bgclearcolor[1], motif.continuebgdef.bgclearcolor[2], motif.continuebgdef.bgclearcolor[3]) --skip last frame rendering
			break
		else
			main.f_cmdInput()
		end
		refresh()
	end
	lastMatchClear()
	return continue
end

return start

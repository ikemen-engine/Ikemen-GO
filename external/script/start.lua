local start = {}

--team side specific data storage
start.p = {{}, {}}
--cell data storage
start.c = {}
for i = 1, --[[gameOption('Config.Players')]]8 do
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
local t_reservedChars = {{}, {}}
local timerSelect = 0
local cursorActive = {}
local cursorDone = {}

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
	if gameOption('Debug.DumpLuaTables') then main.f_printTable(t_ret, 'debug/t_roster.txt') end
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
		return gameOption('Arcade.arcade.AIramp.start')[1], gameOption('Arcade.arcade.AIramp.start')[2], gameOption('Arcade.arcade.AIramp.end')[1], gameOption('Arcade.arcade.AIramp.end')[2]
	elseif start.p[2].ratio then --Ratio
		return gameOption('Arcade.ratio.AIramp.start')[1], gameOption('Arcade.ratio.AIramp.start')[2], gameOption('Arcade.ratio.AIramp.end')[1], gameOption('Arcade.ratio.AIramp.end')[2]
	else --Simul / Turns / Tag
		return gameOption('Arcade.team.AIramp.start')[1], gameOption('Arcade.team.AIramp.start')[2], gameOption('Arcade.team.AIramp.end')[1], gameOption('Arcade.team.AIramp.end')[2]
	end
end
start.t_aiRampData.teamcoop = start.t_aiRampData.arcade
start.t_aiRampData.netplayteamcoop = start.t_aiRampData.arcade
start.t_aiRampData.timeattack = start.t_aiRampData.arcade
start.t_aiRampData.survival = function()
	return gameOption('Arcade.survival.AIramp.start')[1], gameOption('Arcade.survival.AIramp.start')[2], gameOption('Arcade.survival.AIramp.end')[1], gameOption('Arcade.survival.AIramp.end')[2]
end
start.t_aiRampData.survivalcoop = start.t_aiRampData.survival
start.t_aiRampData.netplaysurvivalcoop = start.t_aiRampData.survival

-- generates AI ramping table
function start.f_aiRamp(currentMatch)
	if start.t_aiRampData[gamemode()] == nil then
		panicError("\n" .. gamemode() .. " game mode unrecognized by start.f_aiRamp()\n")
	end
	local start_match, start_diff, end_match, end_diff = start.t_aiRampData[gamemode()]()
	local startAI = gameOption('Options.Difficulty') + start_diff
	if startAI > 8 then
		startAI = 8
	elseif startAI < 1 then
		startAI = 1
	end
	local endAI = gameOption('Options.Difficulty') + end_diff
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
	if gameOption('Debug.DumpLuaTables') then main.f_printTable(t_aiRamp, 'debug/t_aiRamp.txt') end
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
		return gameOption('Options.Difficulty') + offset
	end
end

--assigns AI level, remaps input
function start.f_remapAI(ai)
	--Offset
	local offset = 0
	if gameOption('Arcade.AI.Ramping') and main.aiRamp then
		if t_aiRamp[matchno()] == nil then
			start.f_aiRamp(matchno())
		end
		offset = t_aiRamp[matchno()] - gameOption('Options.Difficulty')
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
			if (getCommandInputSource(side) == side and not main.cpuSide[side] and not main.coop) or start.challenger > 0 or gamemode('training') then
				setCom(side, 0)
			else
				setCom(side, ai or start.f_difficulty(side, offset))
			end
		elseif start.p[side].teamMode == 1 then --Simul
			if not t_ex[side] then
				if (getCommandInputSource(side) == side and not main.cpuSide[side] and not main.coop) or start.challenger > 0 then
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
					if (getCommandInputSource(side) == side and not main.cpuSide[side] and not main.coop) or start.challenger > 0 then
						remapInput(i, getRemapInput(side)) --P1/3/5/7 => P1 controls, P2/4/6/8 => P2 controls
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
	-- disable winscreen if another match exists
	local winscreen = main.motif.winscreen
	if winscreen and main.makeRoster and start.t_roster[matchno() + 1] ~= nil then
		main.motif.winscreen = false
	end
	setMotifElements(main.motif)
	main.motif.winscreen = winscreen
	setLifebarElements(main.lifebar)
	-- Round time
	local frames = main.timeFramesPerCount
	local p1FramesMul = 1
	local p2FramesMul = 1
	if start.p[1].teamMode == 3 then -- Tag
		p1FramesMul = start.p[1].numChars
	end
	if start.p[2].teamMode == 3 then -- Tag
		p2FramesMul = start.p[2].numChars
	end
	if (start.p[1].teamMode == 3 or start.p[2].teamMode == 3) and gameOption('Options.Tag.TimeScaling') > 0 then
		-- Calculate the maximum team size
		local maxTeamSize = math.max(p1FramesMul, p2FramesMul)
		-- Apply a base multiplier for team size
		local adjustedFrames = frames * (1 + (maxTeamSize - 1) * gameOption('Options.Tag.TimeScaling'))
		-- Enforce a minimum threshold to avoid overly short rounds
		frames = main.f_round(math.max(adjustedFrames, frames), 0)
	end
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
	local timer, t_score = start.f_prefightHUD()
	setLifebarTimer(timer)
	setLifebarScore(t_score[1], t_score[2])
end

local function f_listCharRefs(t)
	local ret = {}
	for i = 1, #t do
		table.insert(ret, start.f_getCharData(t[i].ref).char:lower())
	end
	return ret
end

--;===========================================================
-- Accumulators derived from game stats
--;===========================================================
-- Fold matches up to 'upto' (inclusive)
function start.f_accStats(upto)
	local ret = {
		win = {0,0}, lose = {0,0},
		time = { total = 0, matches = {} },
		score = { total = {0,0}, matches = {} },
		consecutive = {0,0},
	}
	local gameStats = readGameStats()
	local matches = (gameStats and gameStats.Matches) or {}
	local n = math.min(upto or #matches, #matches)
	local streak = {0,0}
	for i = 1, n do
		local m = matches[i] or {}
		ret.time.total = ret.time.total + (m.MatchTime or 0)
		if m.TotalScore then
			ret.score.total[1] = m.TotalScore[1] or ret.score.total[1]
			ret.score.total[2] = m.TotalScore[2] or ret.score.total[2]
		end
		local tr, sr = {}, {}
		if m.Rounds then
			for j, r in ipairs(m.Rounds) do
				tr[j] = r.Timer or 0
				sr[j] = { [1] = (r.Score and r.Score[1]) or 0, [2] = (r.Score and r.Score[2]) or 0 }
			end
		end
		table.insert(ret.time.matches, tr)
		table.insert(ret.score.matches, sr)
		if m.WinSide == 1 then
			ret.win[1] = ret.win[1] + 1; ret.lose[2] = ret.lose[2] + 1
			streak[1] = streak[1] + 1; streak[2] = 0
		elseif m.WinSide == 2 then
			ret.win[2] = ret.win[2] + 1; ret.lose[1] = ret.lose[1] + 1
			streak[2] = streak[2] + 1; streak[1] = 0
		end
		for s = 1, 2 do
			if streak[s] > ret.consecutive[s] then ret.consecutive[s] = streak[s] end
		end
	end
	return ret
end

-- Compute HUD timer/score to show at the beginning of the *next* match
function start.f_prefightHUD()
	-- "Next match index" is current matchno(); we want totals of already-finished matches.
	local prev = math.max((matchno() or 1) - 1, 0)
	local acc = start.f_accStats(prev)
	local t_score = {acc.score.total[1], acc.score.total[2]}
	local timer = acc.time.total
	if start.challenger > 0 and gamemode('versus') then
		return 0, {0, 0}
	end
	-- emulate resetScore-on-loss behavior for the next match HUD
	local gameStats = readGameStats()
	local last = (gameStats and gameStats.Matches and gameStats.Matches[prev]) or nil
	if last and main.resetScore and matchno() ~= -1 then
		if last.WinSide == 2 then
			t_score[1] = acc.lose[1]
		end
	end
	return timer, t_score
end

--;===========================================================

--returns the next stage path from the given pool
function start.stageShuffleBag(id, pool)
	-- safety check: prevent nil or invalid pools
	if not pool or type(pool) ~= 'table' or #pool == 0 then
		return nil
	end

	-- safety check: prevent nil id
	id = id or 'defaultStageBag'
	start.shuffleStages = start.shuffleStages or {}
	start.shuffleStages[id] = start.shuffleStages[id] or {}

	if #start.shuffleStages[id] == 0 then
		local t = {}
		for i = 1, #pool do
			table.insert(t, i)
		end
		start.f_shuffleTable(t)
		-- prevent immediate repetition if the bag was just refilled
		if start.lastStageIdx and #pool > 1 and t[#t] == start.lastStageIdx then
			table.insert(t, 1, table.remove(t)) -- rotate
		end
		start.shuffleStages[id] = t
	end

	local idx = table.remove(start.shuffleStages[id])
	start.lastStageIdx = idx
	local result = pool[idx]

	-- ensure result is a valid stage string (handles numeric refs)
	if type(result) == "number" and main.t_selectableStages and main.t_selectableStages[result] then
		result = main.t_selectableStages[result]
	end
	return result
end

--sets stage
function start.f_setStage(num, assigned)
	if main.stageMenu then
		local pool = main.t_selectableStages
		if stageListNo == 0 then
			num = start.stageShuffleBag('stageMenu', pool)
			stageListNo = num -- comment out to randomize stage after each fight in survival mode, when random stage is chosen
			stageRandom = true
		else
			num = pool[stageListNo]
		end
		assigned = true
	end
	if not assigned then
		local sel = start.p[2] and start.p[2].t_selected and start.p[2].t_selected[1]
		local charData = sel and sel.ref and start.f_getCharData(sel.ref)
		if charData and charData.stage and #charData.stage > 0 and not (gamemode('training') and gameOption('Config.TrainingStage')) then --stage assigned as character param
			num = start.stageShuffleBag(charData.ref, charData.stage)
		elseif charData and main.stageOrder and main.t_orderStages[charData.order] then --stage assigned as stage order param
			num = start.stageShuffleBag(charData.order, main.t_orderStages[charData.order])
		elseif gamemode('training') and gameOption('Config.TrainingStage') ~= '' then --training stage
			num = start.f_getStageRef(gameOption('Config.TrainingStage'))
		else
			num = start.stageShuffleBag('includeStage', main.t_includeStage[1])
		end
	end
	selectStage(num)
	return num
end

-- generate table with palette entries already used by this char ref
function start.f_setAssignedPal(ref, t_assignedPals)
	for side = 1, 2 do
		for k, v in pairs(start.p[side].t_selected) do
			if v.ref == ref then
				t_assignedPals[start.p[side].t_selected[k].pal] = true
			end
		end
	end
end

--remaps palette based on button press and character's keymap settings
function start.f_keyPalMap(ref, num)
	return start.f_getCharData(ref).pal_keymap[num] or num
end

-- returns palette number
function start.f_selectPal(ref, palno)
	-- generate table with palette entries already used by this char ref
	local t_assignedPals = {}
	start.f_setAssignedPal(ref, t_assignedPals)

	local charData = start.f_getCharData(ref)
	local availablePals = charData.pal

	-- selected palette by player input
	if palno ~= nil and palno > 0 then
		local mappedPal = start.f_keyPalMap(ref, palno)

		-- Check if the mapped palette is defined and not already used. (MUGEN doesn't do this)
		-- This leads to issues with certain characters who don't have the entire group 1's indices
		-- filled out, so it's been commented out for compatibility.

		-- local isDefined = false
		-- for _, p in ipairs(availablePals) do
		--     if p == mappedPal then
		--         isDefined = true
		--         break
		--     end
		-- end

		if not t_assignedPals[mappedPal] then
			return mappedPal
		end

		-- If the desired palette is not available, find the next available one.

		-- 1. Dynamically build the list of palettes to cycle through
		local cycleList = {1, 2, 3, 4, 5, 6}
		local customDefaults = false

		if charData.pal_defaults then
			local defaultsSet = {}
			for _, p_val in ipairs(charData.pal_defaults) do
				if p_val > 6 then
					-- To avoid duplicates in cycleList
					if not defaultsSet[p_val] then
						table.insert(cycleList, p_val)
						defaultsSet[p_val] = true
						customDefaults = true
					end
				end
			end
			if customDefaults then
				table.sort(cycleList) -- Ensure a consistent cycle order
			end
		end

		-- Exception: If a palette from 7 to 12 was chosen directly, cycle through all 12
		if mappedPal > 6 and not customDefaults then
			cycleList = {1, 2, 3, 4, 5, 6, 7, 8, 9, 10, 11, 12}
		end

		-- 2. Find the starting index for our search in the cycleList
		local startIndex = 1
		for i, p_val in ipairs(cycleList) do
			if p_val == mappedPal then
				startIndex = i
				break
			end
		end

		-- 3. Search for the next available palette in a circular manner
		for i = 1, #cycleList do
			-- Get the index for the next palette in the cycle
			local nextIndex = (startIndex - 1 + i) % #cycleList + 1
			local nextPal = cycleList[nextIndex]

			-- Check if this next palette is defined for the character
			local isNextDefined = false
			for _, p in ipairs(availablePals) do
				if p == nextPal then
					isNextDefined = true
					break
				end
			end

			-- If it's defined and not used, assign it.
			if isNextDefined and not t_assignedPals[nextPal] then
				return nextPal
			end
		end

		-- If all palettes in the cycle list are taken, return the originally mapped one as a fallback.
		return mappedPal

	-- default palette for AI or no-input selection
	elseif (not main.rotationChars and not gameOption('Arcade.AI.RandomColor')) or (main.rotationChars and not gameOption('Arcade.AI.SurvivalColor')) then
		for _, v in ipairs(charData.pal_defaults) do
			if not t_assignedPals[v] then
				return v
			end
		end
	end

	-- random palette
	local t = main.f_tableCopy(availablePals)
	if #t_assignedPals >= #t then -- not enough palettes for unique selection
		if #t > 0 then
			return t[math.random(1, #t)]
		else
			return 1
		end
	end
	main.f_tableShuffle(t)
	for _, v in ipairs(t) do
		if not t_assignedPals[v] then
			return v
		end
	end
	panicError("\n" .. charData.name .. " palette was not selected\n")
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
function start.f_getPlayerNo(side, member)
	if main.coop and not gamemode('versuscoop') then
		return side + member - 1
	end
	if side == 1 then
		return member * 2 - 1
	end
	return member * 2
end

--Convert number to name and get rid of the ""
function start.f_getName(ref, side)
	if ref == nil or start.f_getCharData(ref).hidden == 2 then
		return ''
	end
	if start.f_getCharData(ref).char == 'randomselect' or start.f_getCharData(ref).hidden == 3 then
		return motif.select_info['p' .. (side or 1)].name.random.text
	end
	return start.f_getCharData(ref).name
end

-- Resolve per-member motif tables (p3/p4/..), fallback to p1/p2 when undefined.
local function f_getMotifP(t, pn, side)
	if type(t) ~= "table" then return nil end
	local p = t["p" .. pn]
	if p ~= nil then return p end
	return t["p" .. side]
end

--reset temp data values
function start.f_resetTempData(t, subname)
	-- reset AnimData
	local seenTables = {}
	local seenAnim   = {}
	local function walk(t)
		if type(t) ~= "table" or seenTables[t] then return end
		seenTables[t] = true
		for k, v in pairs(t) do
			if k == "AnimData" and type(v) == "userdata" then
				if not seenAnim[v] then
					animReset(v)
					animUpdate(v)
					seenAnim[v] = true
				end
			elseif type(v) == "table" then
				walk(v)
			end
		end
	end
	walk(t)
	-- generate t_selTemp
	for side = 1, 2 do
		if #start.p[side].t_selTemp == 0 then
			for member, v in ipairs(start.p[side].t_selected) do
				table.insert(start.p[side].t_selTemp, {ref = v.ref})
			end
		end
		for member, v in ipairs(start.p[side].t_selTemp) do
			local pn = 2 * (member - 1) + side
			local pCfg = f_getMotifP(t, pn, side)
			if subname == '' then
				v.face_anim = pCfg.anim
				v.face_data = start.f_animGet(v.ref, side, member, pCfg, nil, true)
			else
				v.face_anim = pCfg[subname].anim
				v.face_data = start.f_animGet(v.ref, side, member, pCfg[subname], nil, true)
			end
			v.face2_anim = pCfg.face2.anim
			v.face2_data = start.f_animGet(v.ref, side, member, pCfg.face2, nil, true)
		end
		start.p[side].screenDelay = 0
	end
end

function start.f_animGet(ref, side, member, params, velParams, loop, srcAnim)
	if not ref then return nil end
	local velParams = velParams or params
	local pn = 2 * (member - 1) + side
	-- Animation/sprite priority order
	for _, v in ipairs({{params.anim, -1}, params.spr}) do
		local anim = v[1]
		if anim ~= nil and anim ~= -1 then
			-- Determine whether to apply palette
			local usePal = params.applypal or false
			-- Try to load the animation
			local a = animGetPreloadedCharData(ref, anim, v[2], loop)
			if a then
				local charData = start.f_getCharData(ref)
				local xscale = start.f_getCharData(ref).portraitscale * motif.info.localcoord[1] / start.f_getCharData(ref).localcoord
				local yscale = xscale
				if v[2] == -1 then
					xscale = xscale * (charData.cns_scale[1] or 1)
					yscale = yscale * (charData.cns_scale[2] or 1)
				end
				animSetLocalcoord(a, motif.info.localcoord[1], motif.info.localcoord[2])
				animSetLayerno(a, params.layerno)
				animSetVelocity(a, velParams.velocity[1], velParams.velocity[2])
				animSetMaxDist(a, velParams.maxdist[1], velParams.maxdist[2])
				animSetAccel(a, velParams.accel[1], velParams.accel[2])
				animSetFriction(a, velParams.friction[1], velParams.friction[2])
				animSetPos(a, 0, 0)
				animSetScale(a, params.scale[1] * xscale, params.scale[2] * yscale)
				animSetFacing(a, params.facing)
				animSetXShear(a, params.xshear)
				animSetAngle(a, params.angle)
				animSetXAngle(a, params.xangle)
				animSetYAngle(a, params.yangle)
				animSetProjection(a, params.projection)
				animSetFocalLength(a, params.focallength)
				animSetWindow(a, params.window[1], params.window[2], params.window[3], params.window[4])
				if srcAnim ~= nil then
					animApplyVel(a, srcAnim)
				end
				-- Apply palette if needed
				local sel = start.p[side].t_selected[member]
				if usePal and not gamemode('netplayteamcoop') then
					local sel = start.p[side].t_selected[member]
					if sel and sel.ref then
						a = start.loadPalettes(a, ref, sel.pal)
					end
				end
				animUpdate(a)
				return a
			end
		end
	end
	return nil
end

--calculate portraits x pos
local function f_portraitsXCalc(side, member, paramsSide, params)
	local x = paramsSide.pos[1] + params.offset[1]
	if paramsSide.padding then
		return x + (2 * member - 1) * paramsSide.spacing[1] * paramsSide.num / (2 * math.min(paramsSide.num, math.max(start.p[side].numChars, #start.p[side].t_selected)))
	end
	return x + (member - 1) * paramsSide.spacing[1]
end

local function getParams(side, member, t, subname)
	local paramsSide = t['p' .. side]
	local pn = 2 * (member - 1) + side
	local params = f_getMotifP(t, pn, side)
	if subname and subname ~= '' then
		paramsSide = paramsSide[subname]
		params = params[subname]
	end
	return paramsSide, params
end

local function drawPortraitRandom(randomCfg)
	if not randomCfg then
		return false
	end
	local spr = randomCfg.spr
	if randomCfg.anim >= 0 or (spr and spr[1] >= 0 and spr[2] >= 0) then
		main.f_animPosDraw(randomCfg.AnimData)
		return true
	end
	return false
end

local function drawPortraitLayer(t_portraits, side, t, subname, last, dataField)
	local lastIdx = #t_portraits
	-- "next player replaces previous one" case
	local paramsSide, params = getParams(side, lastIdx, t, subname)
	if paramsSide.num == 1 and last and not main.coop then
		local v = t_portraits[lastIdx]
		local data = v[dataField]
		if not v.skipCurrent and data ~= nil then
			main.f_animPosDraw(
				data,
				f_portraitsXCalc(side, 1, paramsSide, params),
				paramsSide.pos[2] + params.offset[2]
			)
		end
		-- we're done for this layer in this mode
		return
	end
	-- stacked portraits up to num
	local order = {}
	for member = 1, lastIdx do
		local paramsSide, params = getParams(side, member, t, subname)
		order[#order + 1] = {m = member, o = params.draworder}
	end
	table.sort(order, function(a, b)
		if a.o == b.o then
			return a.m < b.m -- stable legacy tie-break (invertorder=0 behavior)
		end
		return a.o < b.o
	end)
	for _, it in ipairs(order) do
		local member = it.m
		local paramsSide, params = getParams(side, member, t, subname)
		local v = t_portraits[member]
		local data = v[dataField]
		if member <= paramsSide.num and not v.skipCurrent and data ~= nil then
			main.f_animPosDraw(
				data,
				f_portraitsXCalc(side, member, paramsSide, params),
				paramsSide.pos[2] + params.offset[2] + (member - 1) * paramsSide.spacing[2]
			)
		end
	end
end

-- draw portraits
function start.f_drawPortraits(t_portraits, side, t, subname, last, iconDone)
	if #t_portraits == 0 then
		return
	end
	-- reset skip flags
	for m = 1, #t_portraits do
		t_portraits[m].skipCurrent = false
	end
	-- draw random portraits (per member; required for co-op)
	for m = 1, #t_portraits do
		if t_portraits[m].inRandom then
			local pn = 2 * (m - 1) + side
			local pData = f_getMotifP(t, pn, side)
			-- face2 layer random portrait
			if pData.face2.random and drawPortraitRandom(pData.face2.random) then
				t_portraits[m].skipCurrent = true
			end
			-- primary face random portrait
			local baseFace = pData
			if subname and subname ~= '' then
				baseFace = baseFace[subname]
			end
			if baseFace.random and drawPortraitRandom(baseFace.random) then
				t_portraits[m].skipCurrent = true
			end
		end
	end
	-- face2 layer (if present)
	drawPortraitLayer(t_portraits, side, t, 'face2', last, 'face2_data')
	-- primary face layer
	drawPortraitLayer(t_portraits, side, t, subname, last, 'face_data')
	-- draw order icons (unchanged, still using main face params)
	if iconDone == nil then
		return
	end
	for member = 1, #t_portraits do
		local paramsSide, params = getParams(side, member, t, subname)
		if member > paramsSide.num then
			break
		end
		local animData = params.icon.AnimData
		if iconDone then
			animData = params.icon.done.AnimData
		end
		main.f_animPosDraw(
			animData,
			f_portraitsXCalc(side, member, paramsSide, params),
			paramsSide.pos[2] + params.offset[2] + (member - 1) * paramsSide.spacing[2]
		)
	end
end

--returns correct cell position after moving the cursor
function start.f_cellMovement(selX, selY, cmd, side, snd, dir)
	local tmpX = selX
	local tmpY = selY
	local found = false
	if getInput({cmd}, motif.select_info.cell.up.key) or dir == 'U' then
		for i = 1, motif.select_info.rows do
			selY = selY - 1
			if selY < 0 then
				if motif.select_info.wrapping or dir ~= nil then
					selY = motif.select_info.rows - 1
				else
					selY = tmpY
				end
			end
			if dir ~= nil then
				found, selX = start.f_searchEmptyBoxes(selX, selY, side, -1)
			elseif (start.t_grid[selY + 1][selX + 1].char ~= nil or motif.select_info.moveoveremptyboxes) and start.t_grid[selY + 1][selX + 1].skip ~= 1 and (gameOption('Options.Team.Duplicates') or start.t_grid[selY + 1][selX + 1].char == 'randomselect' or not t_reservedChars[side][start.t_grid[selY + 1][selX + 1].char_ref]) and start.t_grid[selY + 1][selX + 1].hidden ~= 2 then
				break
			elseif motif.select_info.searchemptyboxesup then
				found, selX = start.f_searchEmptyBoxes(selX, selY, side, 1)
			end
			if found then
				break
			end
		end
	elseif getInput({cmd}, motif.select_info.cell.down.key) or dir == 'D' then
		for i = 1, motif.select_info.rows do
			selY = selY + 1
			if selY >= motif.select_info.rows then
				if motif.select_info.wrapping or dir ~= nil then
					selY = 0
				else
					selY = tmpY
				end
			end
			if dir ~= nil then
				found, selX = start.f_searchEmptyBoxes(selX, selY, side, 1)
			elseif (start.t_grid[selY + 1][selX + 1].char ~= nil or motif.select_info.moveoveremptyboxes) and start.t_grid[selY + 1][selX + 1].skip ~= 1 and (gameOption('Options.Team.Duplicates') or start.t_grid[selY + 1][selX + 1].char == 'randomselect' or not t_reservedChars[side][start.t_grid[selY + 1][selX + 1].char_ref]) and start.t_grid[selY + 1][selX + 1].hidden ~= 2 then
				break
			elseif motif.select_info.searchemptyboxesdown then
				found, selX = start.f_searchEmptyBoxes(selX, selY, side, 1)
			end
			if found then
				break
			end
		end
	elseif getInput({cmd}, motif.select_info.cell.left.key) or dir == 'B' then
		if dir ~= nil then
			found, selX = start.f_searchEmptyBoxes(selX, selY, side, -1)
		else
			for i = 1, motif.select_info.columns do
				selX = selX - 1
				if selX < 0 then
					if motif.select_info.wrapping then
						selX = motif.select_info.columns - 1
					else
						selX = tmpX
					end
				end
				if (start.t_grid[selY + 1][selX + 1].char ~= nil or motif.select_info.moveoveremptyboxes) and start.t_grid[selY + 1][selX + 1].skip ~= 1 and (gameOption('Options.Team.Duplicates') or start.t_grid[selY + 1][selX + 1].char == 'randomselect' or not t_reservedChars[side][start.t_grid[selY + 1][selX + 1].char_ref]) and start.t_grid[selY + 1][selX + 1].hidden ~= 2 then
					break
				end
			end
		end
	elseif getInput({cmd}, motif.select_info.cell.right.key) or dir == 'F' then
		if dir ~= nil then
			found, selX = start.f_searchEmptyBoxes(selX, selY, side, 1)
		else
			for i = 1, motif.select_info.columns do
				selX = selX + 1
				if selX >= motif.select_info.columns then
					if motif.select_info.wrapping then
						selX = 0
					else
						selX = tmpX
					end
				end
				if (start.t_grid[selY + 1][selX + 1].char ~= nil or motif.select_info.moveoveremptyboxes) and start.t_grid[selY + 1][selX + 1].skip ~= 1 and (gameOption('Options.Team.Duplicates') or start.t_grid[selY + 1][selX + 1].char == 'randomselect' or not t_reservedChars[side][start.t_grid[selY + 1][selX + 1].char_ref]) and start.t_grid[selY + 1][selX + 1].hidden ~= 2 then
					break
				end
			end
		end
	end
	if (tmpX ~= selX or tmpY ~= selY) then
		if dir == nil then
			sndPlay(motif.Snd, snd[1], snd[2])
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

-- Returns player cursor data
function start.f_getCursorData(pn)
	-- In coop/multi, p3+ may be undefined in motif; fallback to p1/p2 by parity.
	if main.coop then
		return motif.select_info['p' .. pn] or motif.select_info['p' .. ((pn - 1) % 2 + 1)]
	end
	return motif.select_info['p' .. ((pn - 1) % 2 + 1)]
end

-- Reset cursor animation for a specific slot only
local function resetCursorData(pn, store, param)
	local pData = start.f_getCursorData(pn)
	local cursorCfg = pData.cursor[param]
	local key = start.c[pn].selX .. '-' .. start.c[pn].selY
	local cursorParams = cursorCfg.default
	if cursorCfg[key] then
		cursorParams = cursorCfg[key]
	end
	local src = cursorParams.AnimData
	if not src then
		return
	end
	store[pn] = store[pn] or {}
	local cd = store[pn]
	cd.animCache = cd.animCache or {}
	local cache = cd.animCache[param]
	if cache == nil or cache.src ~= src then
		cache = {src = src, anim = animCopy(src)}
		cd.animCache[param] = cache
		if cache.anim then
			animReset(cache.anim)
			animUpdate(cache.anim)
		end
	end
	if cache.anim then
		animReset(cache.anim)
		animUpdate(cache.anim)
	end
end

-- Calculate cursor.tween
local function cursorTween(val, target, factor)
	if not factor or not target then
		return val
	end
	for i = 1, 2 do
		local t = target[i] or 0
		local f = math.min(math.abs(factor[i] or 0.5), 1)
		val[i] = val[i] + (t - val[i]) * f
	end
	return val
end

local function getCellOverride(col, row)
    local cells = motif.select_info.cell
    local exact = col .. '-' .. row
	local colWild = col .. '-*'
	local rowWild = '*-' .. row
    if cells[exact] then 
		return cells[exact] 
	end
    if cells[colWild] then 
		return cells[colWild] 
	end
    if cells[rowWild] then 
		return cells[rowWild] 
	end
    if cells['*-*'] then 
		return cells['*-*'] 
	end
    return nil
end

function getCellFacing(default, col, row)
	local override = getCellOverride(col, row)
	if override ~= nil and override.facing ~= 0 then
		return override.facing
	end
	return default
end

function getCellOffset(col, row)
	local override = getCellOverride(col, row)
	if override ~= nil and override.offset ~= nil then
		return override.offset
	end
	return {0, 0}
end

function getCellSpacing(col, row)
	local override = getCellOverride(col, row)
	if override ~= nil and override.spacing ~= nil then
		if override.spacing[1] ~= 0 then
			if override.spacing[2] == 0 then
				return {override.spacing[1], override.spacing[1]}
			end
			return override.spacing
		end
		if override.spacing[2] ~= 0 then
			return override.spacing
		end
	end
	return motif.select_info.cell.spacing
end

function getCellSkip(col, row)
	local override = getCellOverride(col, row)
	if override ~= nil then
		return override.skip
	end
	return false
end

function getCellTransform(col, row, paramName, default)
	local override = getCellOverride(col, row)
	if override ~= nil then
		local val = override[paramName]
		-- Table Validation
		if type(val) == "table" then
			if paramName == "scale" then
				if val[1] ~= 0 or val[2] ~= 0 then return val end
			else
				return val
			end
		elseif type(val) == "string" then
			if val ~= "" then 
				return val 
			end
		elseif type(val) == "number" then
			if val ~= 0 then 
				return val 
			end
		end
	end
	return default
end

--draw cursor
function start.f_drawCursor(pn, x, y, param, done)
	local pData = start.f_getCursorData(pn)
	-- select appropriate cursor table and initialize if needed
	local store = done and cursorDone or cursorActive
	store[pn] = store[pn] or {}
	local cd = store[pn]
	cd.currentPos  = cd.currentPos  or {0, 0}
	cd.targetPos   = cd.targetPos   or {0, 0}
	cd.startPos    = cd.startPos    or {0, 0}
	cd.slideOffset = cd.slideOffset or {0, 0}
	cd.init        = cd.init or false
	if not done then
		cd.snap = cd.snap or false -- only used by active cursors
	end
	-- calculate target cell coordinates using the pre-calculated grid
	local cellData = start.t_grid[y + 1] and start.t_grid[y + 1][x + 1]
	local baseX, baseY
	if cellData then
		-- cellData already includes all spacing and offsets
		baseX = motif.select_info.pos[1] + cellData.x
		baseY = motif.select_info.pos[2] + cellData.y
	end
	-- initialization or snap: set cursor directly
	if not cd.init or done or cd.snap then
		for i = 1, 2 do
			cd.currentPos[i] = (i == 1) and baseX or baseY
			cd.targetPos[i]  = cd.currentPos[i]
			cd.startPos[i]   = cd.currentPos[i]
			cd.slideOffset[i]= 0
		end
		cd.init, cd.snap = true, false
	-- new cell selected: recalc tween offsets
	elseif cd.targetPos[1] ~= baseX or cd.targetPos[2] ~= baseY then
		cd.startPos[1], cd.startPos[2] = cd.currentPos[1], cd.currentPos[2]
		cd.targetPos[1], cd.targetPos[2] = baseX, baseY
		cd.slideOffset[1] = cd.startPos[1] - baseX
		cd.slideOffset[2] = cd.startPos[2] - baseY
	end
	local t_factor = {
		pData.cursor.tween.factor[1],
		pData.cursor.tween.factor[2]
	}
	-- apply tween if enabled, otherwise snap to target
	if not done and t_factor[1] > 0 and t_factor[2] > 0 then
		cursorTween(cd.slideOffset, {0, 0}, t_factor)
	else
		cd.slideOffset[1], cd.slideOffset[2] = 0, 0
	end
	if pData.cursor.tween.wrap.snap then
		local dx = cd.targetPos[1] - cd.startPos[1]
		local dy = cd.targetPos[2] - cd.startPos[2]
		if math.abs(dx) > motif.select_info.cell.size[1] * (motif.select_info.columns - 1) or math.abs(dy) > motif.select_info.cell.size[2] * (motif.select_info.rows - 1) then
		cd.slideOffset[1], cd.slideOffset[2] = 0, 0	
		end
	end
	-- update final cursor position
	cd.currentPos[1] = cd.targetPos[1] + cd.slideOffset[1]
	cd.currentPos[2] = cd.targetPos[2] + cd.slideOffset[2]
	-- draw
	local params = pData.cursor[param].default
	local key = x .. '-' .. y
	if pData.cursor[param][key] ~= nil then
		params = pData.cursor[param][key]
	end
	local a = params.AnimData
	cd.animCache = cd.animCache or {}
	local cache = cd.animCache[param]
	if cache == nil or cache.src ~= a then
		cache = {src = a, anim = animCopy(a)}
		cd.animCache[param] = cache
		if cache.anim then
			animReset(cache.anim)
		end
	end
	a = cache.anim
	animSetFacing(a, getCellFacing(params.facing, x, y))
	local scale = getCellTransform(x, y, "scale", params.scale)
	animSetScale(a, scale[1], scale[2])
	animSetXShear(a, getCellTransform(x, y, "xshear", params.xshear))
	animSetAngle(a, getCellTransform(x, y, "angle", params.angle))
	animSetXAngle(a, getCellTransform(x, y, "xangle", params.xangle))
	animSetYAngle(a, getCellTransform(x, y, "yangle", params.yangle))
	animSetProjection(a, getCellTransform(x, y, "projection", params.projection))
	animSetFocalLength(a, getCellTransform(x, y, "focallength", params.focallength))
	animUpdate(a)
	main.f_animPosDraw(a, cd.currentPos[1], cd.currentPos[2], getCellFacing(params.facing, x, y))
end

-- snaps the cursor instantly to its target cell
local function f_snapCursor()
	for k, v in pairs(cursorActive) do
		v.snap = true
	end
end

--returns t_selChars table out of cell number
function start.f_selGrid(cell, slot)
	if main.t_selGrid[cell] == nil or #main.t_selGrid[cell].chars == 0 then
		local csCol = ((cell - 1) % motif.select_info.columns) + 1
		local csRow = math.floor((cell - 1) / motif.select_info.columns) + 1
		local cellCfg = motif.select_info.cell[(csCol - 1) .. '-' .. (csRow - 1)]
		if cellCfg ~= nil and cellCfg.skip then
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
	local text = motif.select_info.record.text[gamemode()]
	if text == nil then
		return ""
	end
	local stats = jsonDecode('save/stats.json')
	if stats.modes == nil or stats.modes[gamemode()] == nil or stats.modes[gamemode()].ranking == nil or stats.modes[gamemode()].ranking[1] == nil then
		return ""
	end
	local t = stats.modes[gamemode()].ranking[1]
	--time
	text = start.f_clearTimeText(text, t.time)
	--score
	text = text:gsub('%%p', tostring(t.score))
	--win
	text = text:gsub('%%r', tostring(t.win))
	--char name
	local name = '?' --in case character being removed from roster
	if main.t_charDef[t.chars[1]] ~= nil then
		name = start.f_getCharData(main.t_charDef[t.chars[1]]).name
	end
	text = text:gsub('%%c', name)
	--player name
	text = text:gsub('%%n', t.name)
	return text
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
		wavePlay(main.t_selStages[ref][name .. '_wave_data'], g, n)
	else
		local sound = start.f_getCharData(ref).sound
		if sound == nil or sound == '' then
			return 0
		end
		local key = name .. '_wave_data_' .. g .. '_' .. n
		if start.f_getCharData(ref)[key] == nil then
			start.f_getCharData(ref)[key] =
				getWaveData(start.f_getCharData(ref).dir .. start.f_getCharData(ref).sound, g, n, loops or -1)
		end
		wavePlay(start.f_getCharData(ref)[key], g, n)
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

--shuffles a table in-place (using synced RNG)
function start.f_shuffleTable(t, last)
	for i = #t, 2, -1 do
		local j = (sszRandom() % i) + 1
		t[i], t[j] = t[j], t[i]
	end
	-- prevent first element from repeating the last of previous cycle
	if last and #t > 1 and t[#t] == last then
		-- swap the first element with a random other position
		local swap = (sszRandom() % (#t - 1)) + 1
		t[#t], t[swap] = t[swap], t[#t]
	end
end

--returns random char ref
function start.f_randomChar(pn)
	if #main.t_randomChars == 0 then
		return nil
	end
	start.shuffleChars = start.shuffleChars or {}

	if not start.shuffleChars[pn] or #start.shuffleChars[pn] == 0 then
		local last = start.lastRandomChar and start.lastRandomChar[pn]
		local t = {}
		for _, v in ipairs(main.t_randomChars) do
			if gameOption('Options.Team.Duplicates') or not t_reservedChars[pn][v] then
				table.insert(t, v)
			end
		end
		start.f_shuffleTable(t, last)
		start.shuffleChars[pn] = t
	end
	-- draw one char from the bag
	local result = table.remove(start.shuffleChars[pn])
	-- store the last drawn value
	start.lastRandomChar = start.lastRandomChar or {}
	start.lastRandomChar[pn] = result
	return result
end

--return true if slot is selected, update start.t_grid
function start.f_slotSelected(cell, side, cmd, player, x, y)
	if cmd == nil then
		return false, false
	end
	if #main.t_selGrid[cell].chars > 0 then
		-- select.def 'slot' parameter special keys detection
		for _, cmdType in ipairs({'select', 'next', 'previous'}) do
			if main.t_selGrid[cell][cmdType] ~= nil then
				for k, v in pairs(main.t_selGrid[cell][cmdType]) do
					if getInput({cmd}, {k}) then
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
								sndPlay(motif.Snd, motif.select_info['p' .. side].swap.snd[1], motif.select_info['p' .. side].swap.snd[2])
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
								sndPlay(motif.Snd, motif.select_info['p' .. side].swap.snd[1], motif.select_info['p' .. side].swap.snd[2])
							end
						else --select
							main.t_selGrid[cell].slot = v[(sszRandom() % #v) + 1]
							start.c[player].selRef = start.f_selGrid(cell).char_ref
						end
						start.t_grid[y + 1][x + 1].char = start.f_selGrid(cell).char
						start.t_grid[y + 1][x + 1].char_ref = start.f_selGrid(cell).char_ref
						start.t_grid[y + 1][x + 1].hidden = start.f_selGrid(cell).hidden
						start.t_grid[y + 1][x + 1].skip = start.f_selGrid(cell).skip
						return cmdType == 'select', (main.t_selGrid[cell].slot ~= original_slot)
					end
				end
			end
		end
	end
	-- returns true on pressed key if current slot is not blocked by TeamDuplicates feature
	return main.f_btnPalNo(cmd) > 0 and (not t_reservedChars[side][start.t_grid[y + 1][x + 1].char_ref] or start.t_grid[start.c[player].selY + 1][start.c[player].selX + 1].char == 'randomselect'),false
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
	local cell_spacing = getCellSpacing(col - 1, row - 1)
	local cell_offset = getCellOffset(col - 1, row - 1)
	start.t_grid[row][col] = {
		x = (col - 1) * (motif.select_info.cell.size[1] + cell_spacing[1]) + cell_offset[1],
		y = (row - 1) * (motif.select_info.cell.size[2] + cell_spacing[2]) + cell_offset[2]
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
	local overrideSkip = getCellSkip(col - 1, row - 1)
	if start.f_selGrid(i).skip == 1 or overrideSkip then
		start.t_grid[row][col].skip = 1
	end
end
if gameOption('Debug.DumpLuaTables') then main.f_printTable(start.t_grid, 'debug/t_grid.txt') end

-- return amount of life to recover
local function f_lifeRecovery(lifeMax, ratioLevel)
	local bonus = lifeMax * gameOption('Options.Turns.Recovery.Bonus') / 100
	local base = lifeMax * gameOption('Options.Turns.Recovery.Base') / 100
	if ratioLevel > 0 then
		bonus = lifeMax * gameOption('Options.Ratio.Recovery.Bonus') / 100
		base = lifeMax * gameOption('Options.Ratio.Recovery.Base') / 100
	end
	return base + main.f_round(timeremaining() / (timeremaining() + timeelapsed()) * bonus)
end

-- match persistence
function start.f_matchPersistence()
	-- checked only after at least 1 match
	if matchno() >= 2 then
		local gameStats = readGameStats()
		local roundStats = gameStats.Matches[matchno()-1].Rounds
		-- set 'existed' flag (decides if var/fvar should be persistent between matches)
		if roundStats then
			for _, round in ipairs(roundStats) do
				for side = 1, 2 do
					local fighters = (round.Fighters and round.Fighters[side]) or {}
					for _, f in ipairs(fighters) do
						local memberIdx = (f.MemberNo or 0) + 1
						if start.p[side].t_selected[memberIdx] ~= nil then
							start.p[side].t_selected[memberIdx].existed = true
						end
					end
				end
			end
		end

		-- if defeated members should be removed from team, or if life should be maintained
		if main.dropDefeated or main.persistLife then
			local t_removeMembers = {}
			-- Turns
			if start.p[1].teamMode == 2 then
				--for each round in the last match
				if roundStats then
					for _, round in ipairs(roundStats) do
						-- P1 active fighter snapshot for the round
						local f1 = (round.Fighters and round.Fighters[1] and round.Fighters[1][1]) or nil
						if f1 then
							local memberIdx = (f1.MemberNo or 0) + 1
							-- if defeated
							if f1.KO and (f1.Life or 0) <= 0 then
								-- remove character from team
								if main.dropDefeated then
									t_removeMembers[memberIdx] = true
								-- or resurrect and recover character's life
								elseif main.persistLife then
									start.p[1].t_selected[memberIdx].life = math.max(1, f_lifeRecovery(f1.LifeMax or 0, f1.RatioLevel or 0))
								end
							-- otherwise maintain character's life
							elseif main.persistLife then
								start.p[1].t_selected[memberIdx].life = f1.Life or start.p[1].t_selected[memberIdx].life
							end
						end
					end
				end
			-- Single / Simul / Tag
			else
				-- for each player data in the last round (new format)
				if roundStats and #roundStats > 0 then
					local lastRound = roundStats[#roundStats]
					for side = 1, 2 do
						local fighters = (lastRound.Fighters and lastRound.Fighters[side]) or {}
						-- only check player-controlled side
						if not main.cpuSide[side] then
							for _, f in ipairs(fighters) do
								local memberIdx = (f.MemberNo or 0) + 1
								-- if defeated
								if f.KO and (f.Life or 0) <= 0 then
									-- remove character from team
									if main.dropDefeated then
										t_removeMembers[memberIdx] = true
									-- or resurrect and recover character's life
									elseif main.persistLife then
										start.p[1].t_selected[memberIdx].life = math.max(1, f_lifeRecovery(f.LifeMax or 0, f.RatioLevel or 0))
									end
								-- otherwise maintain character's life
								elseif main.persistLife then
									start.p[1].t_selected[memberIdx].life = f.Life or start.p[1].t_selected[memberIdx].life
								end
							end
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

--start game
function start.f_game(lua)
	clearColor(0, 0, 0)
	if gameOption('Debug.DumpLuaTables') and start ~= nil then main.f_printTable(start.p, 'debug/t_p.txt') end
	local p2In = getCommandInputSource(2)
	setCommandInputSource(2, 2)
	if lua ~= '' then
		local t = gameOption('Common.Lua')
		local ok = false
		for _, v in ipairs(t) do
			if v == lua then
				ok = true
				break
			end
		end
		if not ok then
			table.insert(t, lua)
		end
		modifyGameOption('Common.Lua', t)
	end
	if gamemode('training') then
		menu.f_trainingReset()
	end
	local winner = -1
	winner, start.challenger = game()

	if gameOption('Debug.DumpLuaTables') then main.f_printTable(readGameStats(), 'debug/t_gameStats.txt') end

	main.f_restoreInput()
	if lua ~= '' then
		local t = gameOption('Common.Lua')
		for i, v in ipairs(t) do
			if v == lua then
				table.remove(t, i)
				break
			end
		end
		modifyGameOption('Common.Lua', t)
	end
	if gameend() then
		clearColor(0, 0, 0)
		os.exit()
	end
	setCommandInputSource(2, p2In)
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
			sndPlay(motif.Snd, motif.select_info.cancel.snd[1], motif.select_info.cancel.snd[2])
			bgReset(motif[main.background].BGDef)
			main.f_fadeReset('fadein', motif[main.group])
			playBgm({source = "motif.title", interrupt = true})
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
				-- hiscore & stats handled in Go; returns (cleared, place)
				local cleared, place = computeRanking(gamemode())
				if main.motif.hiscore and place > 0 then
					main.f_hiscore(gamemode(), place)
				end
				--credits
				if cleared and main.storyboard.credits and motif.end_credits.enabled and main.f_fileExists(motif.end_credits.storyboard) then
					main.f_storyboard(motif.end_credits.storyboard)
				end
				--game over
				if main.storyboard.gameover and motif.game_over_screen.enabled and main.f_fileExists(motif.game_over_screen.storyboard) then
					if cleared or not main.motif.continuescreen or (not continue() and motif.continue_screen.gameover.enabled) then
						main.f_storyboard(motif.game_over_screen.storyboard)
					end
				end
				--exit to main menu
				if main.exitSelect then
					if motif.files.intro.storyboard ~= '' and not motif.attract_mode.enabled then
						main.f_storyboard(motif.files.intro.storyboard)
					end
				end
				start.exit = start.exit or main.exitSelect or not main.selectMenu[1]
			end
			if start.exit then
				bgReset(motif[main.background].BGDef)
				main.f_fadeReset('fadein', motif[main.group])
				playBgm({source = "motif.title", interrupt = true})
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
	main.f_cmdBufReset()
	resetGameStats()
	setMatchNo(1)
	setConsecutiveWins(1, 0)
	setConsecutiveWins(2, 0)
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
		if motif.select_info.stage.randomselect == 0 or motif.select_info.stage.randomselect == 2 then
			stageListNo = 1
		else
			stageListNo = 0
		end
		restoreCursor = false
		--cursor start cell
		for i = 1, gameOption('Config.Players') do
			if start.f_getCursorData(i).cursor.startcell[1] < motif.select_info.rows then
				start.c[i].selY = start.f_getCursorData(i).cursor.startcell[1]
			else
				start.c[i].selY = 0
			end
			if start.f_getCursorData(i).cursor.startcell[2] < motif.select_info.columns then
				start.c[i].selX = start.f_getCursorData(i).cursor.startcell[2]
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
			start.p[side].numSimul = math.max(2, gameOption('Options.Simul.Min'))
			start.p[side].numTag = math.max(2, gameOption('Options.Tag.Min'))
			start.p[side].numTurns = math.max(2, gameOption('Options.Turns.Min'))
			start.p[side].numRatio = 1
			start.p[side].teamMenu = 1
			start.p[side].t_cursor = {}
			start.p[side].teamMode = 0
		end
		start.p[side].numSimul = math.min(start.p[side].numSimul, gameOption('Options.Simul.Max'))
		start.p[side].numTag = math.min(start.p[side].numTag, gameOption('Options.Tag.Max'))
		start.p[side].numTurns = math.min(start.p[side].numTurns, gameOption('Options.Turns.Max'))
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
	cursorActive = {}
	cursorDone = {}
	if start.challenger == 0 then
		start.t_roster = {}
		start.reset = true
	end
	menu.movelistChar = 1
	hook.run("start.f_selectReset")
end

function start.f_selectChallenger()
	esc(false)
	--save values
	local t_p_sav = main.f_tableCopy(start.p)
	local t_c_sav = main.f_tableCopy(start.c)
	local matchNo_sav = matchno()
	local p1ConsecutiveWins = getConsecutiveWins(1)
	local p2ConsecutiveWins = getConsecutiveWins(2)
	--start challenger match
	main.f_default()
	remapInput(2, getRemapInput(start.challenger))
	main.t_itemname.versus()
	start.f_selectReset(false)
	if not start.f_selectScreen() then
		start.exit = true
		return false
	end
	local ok = launchFight{challenger = true}
	--restore values
	main.f_default()
	setLastInputController(getRemapInput(1))
	main.t_itemname.arcade()
	if not ok then
		return false
	end
	start.p = t_p_sav
	start.c = t_c_sav
	setMatchNo(matchNo_sav)
	setConsecutiveWins(1, p1ConsecutiveWins)
	setConsecutiveWins(2, p2ConsecutiveWins)
	return true
end

local function buildMusicParams(data)
	local out = {}
	for k, v in pairs(data) do
		if type(k) == "string" and k:match("music$") then
			if type(v) == "string" then
				out[#out + 1] = k .. "=" .. v
			elseif type(v) == "table" and #v > 0 then
				local first = v[1]
				-- If table looks like { "path.mp3", 100, 123, 456 } => positional args
				local positional = (type(first) == "string") and (#v == 1 or type(v[2]) ~= "string")
				if positional then
					local pieces = {}
					for i = 1, #v do
						pieces[i] = tostring(v[i])
					end
					-- space-separated to avoid commas inside the value
					out[#out + 1] = k .. "=" .. table.concat(pieces, " ")
				else
					-- Treat as multiple candidate tracks: {"a.mp3","b.mp3",...}
					for i = 1, #v do
						out[#out + 1] = k .. "=" .. tostring(v[i])
					end
				end
			end
		end
	end
	return table.concat(out, ", ")
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
		t.continue = main.f_arg(data.continue, main.motif.continuescreen)
		t.quickcontinue = (not main.selectMenu[1] and not main.selectMenu[2]) or main.f_arg(data.quickcontinue, main.quickContinue or gameOption('Options.QuickContinue'))
		t.order = data.order or 1
		t.orderselect = {main.f_arg(data.p1orderselect, main.orderSelect[1]), main.f_arg(data.p2orderselect, main.orderSelect[2])}
		t.p1char = data.p1char or {}
		t.p1pal = data.p1pal
		t.p1numratio = data.p1numratio or {}
		t.p1rounds = data.p1rounds or nil
		t.p2char = data.p2char or {}
		t.p2pal = data.p2pal
		t.p2numratio = data.p2numratio or {}
		t.p2rounds = data.p2rounds or nil
		t.exclude = data.exclude or {}
		t.musicParams = buildMusicParams(data)
		t.stage = data.stage or ''
		t.ai = data.ai or nil
		t.vsscreen = main.f_arg(data.vsscreen, main.motif.versusscreen)
		t.victoryscreen = main.f_arg(data.victoryscreen, main.motif.victoryscreen)
		--t.frames = data.frames or fightscreenvar("time.framespercount")
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
				pal = t.p1pal or start.f_selectPal(ref),
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
				pal = t.p2pal or start.f_selectPal(ref),
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
	--TODO: fix gameOption('Config.BackgroundLoading') setting
	--if gameOption('Config.BackgroundLoading') then
	--	selectStart()
	--else
		clearSelected()
	--end
	local ok = false
	local loopCount = 0
	while true do
		-- fight initialization
		setTeamMode(1, start.p[1].teamMode, start.p[1].numChars)
		setTeamMode(2, start.p[2].teamMode, start.p[2].numChars)
		start.f_remapAI(t.ai)
		start.f_setRounds(t.roundtime, {t.p1rounds, t.p2rounds})
		t.stageNo = start.f_setStage(t.stageNo, t.stage ~= '' or continue() or loopCount > 0)
		if not start.f_selectVersus(t.vsscreen, t.orderselect) then break end
		start.f_selectLoading(t.musicParams)
		local continueScreen = main.motif.continuescreen
		local victoryScreen = main.motif.victoryscreen
		main.motif.continuescreen = t.continue
		main.motif.victoryscreen = t.victoryscreen
		hook.run("launchFight")
		start.f_game(t.lua)
		main.motif.continuescreen = continueScreen
		main.motif.victoryscreen = victoryScreen
		clearColor(motif.selectbgdef.bgclearcolor[1], motif.selectbgdef.bgclearcolor[2], motif.selectbgdef.bgclearcolor[3])
		if start.exit or start.characterchange then
			start.characterchange = false
			break
		-- here comes a new challenger
		elseif start.challenger > 0 then
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
			ok = true -- continue lua code execution
			break
		-- continue = no
		elseif not continue() then
			setMatchNo(-1)
			break
		-- continue = yes
		elseif not t.quickcontinue then -- if 'Quick Continue' is disabled
			for i = 1, 2 do
				for _, v in ipairs(start.p[i].t_selCmd) do
					v.selectState = 0
				end
			end
			start.p[1].t_selected = {}
			start.p[1].t_selTemp = {}
			start.p[1].selEnd = false
			start.launchFightSav = main.f_tableCopy(t)
			--start.p[2].t_selTemp = {} -- uncomment to disable enemy team showing up in select screen
			selScreenEnd = false
			return
		end
		start.challenger = 0
		loopCount = loopCount + 1
	end
	-- restore original values
	start.p[1].numChars = t.p1numchars
	start.p[1].teamMode = t.p1teammode
	start.p[2].numChars = t.p2numchars
	start.p[2].teamMode = t.p2teammode
	return ok
end

function launchStoryboard(path)
	if path == nil or path == '' then
		return false
	end
	main.f_storyboard(path)
	return true
end

function codeInput(name)
	if main.t_commands[name] == nil then
		return false
	end
	if commandGetState(main.t_cmd[getLastInputController()], name) then
		return true
	end
	return false
end

--;===========================================================
--; SELECT SCREEN
--;===========================================================
function start.updateDrawList()
	local drawList = {}

	for row = 1, motif.select_info.rows do
		for col = 1, motif.select_info.columns do
			local cellIndex = (row - 1) * motif.select_info.columns + col
			local t = start.t_grid[row][col]
			local c = col - 1
			local r = row - 1

			if t.skip ~= 1 then
				local charData = start.f_selGrid(cellIndex)
				local function getTransforms(base)
					return {
						facing      = getCellFacing(base.facing, c, r),
						scale       = getCellTransform(c, r, "scale", base.scale),
						xshear      = getCellTransform(c, r, "xshear", base.xshear),
						angle       = getCellTransform(c, r, "angle", base.angle),
						xangle      = getCellTransform(c, r, "xangle", base.xangle),
						yangle      = getCellTransform(c, r, "yangle", base.yangle),
						projection  = getCellTransform(c, r, "projection", base.projection),
						focallength = getCellTransform(c, r, "focallength", base.focallength)
					}
				end

				if (charData and charData.char ~= nil and (charData.hidden == 0 or charData.hidden == 3)) or motif.select_info.showemptyboxes then
					local item = getTransforms(motif.select_info.cell.bg)
					item.anim = motif.select_info.cell.bg.AnimData
					item.x = motif.select_info.pos[1] + t.x
					item.y = motif.select_info.pos[2] + t.y
					table.insert(drawList, item)
				end

				if charData and (charData.char == 'randomselect' or charData.hidden == 3) then
					local item = getTransforms(motif.select_info.cell.random)
					item.anim = motif.select_info.cell.random.AnimData
					item.x = motif.select_info.pos[1] + t.x + motif.select_info.portrait.offset[1]
					item.y = motif.select_info.pos[2] + t.y + motif.select_info.portrait.offset[2]
					table.insert(drawList, item)
				end

				if charData and charData.char_ref ~= nil and charData.hidden == 0 then
					local item = getTransforms(motif.select_info.portrait)
					item.anim = charData.cell_data
					item.x = motif.select_info.pos[1] + t.x + motif.select_info.portrait.offset[1]
					item.y = motif.select_info.pos[2] + t.y + motif.select_info.portrait.offset[2]
					-- apply cell scale override while preserving portrait resolution factor
					if item.scale ~= nil then
						local charInfo = main.t_selChars[charData.char_ref + 1]
						if charInfo then
							local portraitScale = charInfo.portraitscale or 1
							local charLocalcoord = charInfo.localcoord or motif.info.localcoord[1]
							-- recompute resolution compensation factor
							local resFix = portraitScale * motif.info.localcoord[1] / charLocalcoord
							item.scale = {
								item.scale[1] * resFix,
								item.scale[2] * resFix
							}
						end
					end
					table.insert(drawList, item)
				end
			end
		end
	end

	return drawList
end

start.needUpdateDrawList = false
function start.f_selectScreen()
	if (not main.selectMenu[1] and not main.selectMenu[2]) or selScreenEnd then
		return true
	end
	bgReset(motif.selectbgdef.BGDef)
	main.f_fadeReset('fadein', motif.select_info)
	local fadeOutStarted = false
	playBgm({source = "motif.select", interrupt = true})
	start.f_resetTempData(motif.select_info, 'face')
	f_snapCursor()
	local stageActiveCount = 0
	local stageActiveState = false
	timerSelect = 0
	start.escFlag = false
	local t_teamMenu = {{}, {}}
	local blinkCount = 0
	local counter = 0 - motif.select_info.fadein.time
	local timerReset = false
	local stageTextData = motif.select_info.stage.active.TextSpriteData
	-- generate team mode items table
	for side = 1, 2 do
		-- read display names for the current gamemode (or default)
		local params = motif.select_info.teammenu.itemname.default
		if motif.select_info.teammenu.itemname[gamemode()] ~= nil then
			params = motif.select_info.teammenu.itemname[gamemode()]
		end
		-- read itemname_order for the current gamemode (or default)
		local itemname_order = motif.select_info.teammenu.itemname_order.default
		if motif.select_info.teammenu.itemname_order[gamemode()] ~= nil then
			itemname_order = motif.select_info.teammenu.itemname_order[gamemode()]
		end
		-- map itemname -> mode (kept from old defaults)
		local modeByName = {
			single = 0,
			simul  = 1,
			turns  = 2,
			tag    = 3,
			ratio  = 2,
		}
		-- itemname_order lists exactly what to render, in the correct order
		for _, name in ipairs(itemname_order) do
			local itemname = name
			local mode = modeByName[itemname]
			if mode ~= nil and main.teamMenu[side][itemname] then
				table.insert(t_teamMenu[side], {
					itemname    = itemname,
					displayname = params[itemname],
					mode        = mode,
				})
			end
		end
	end

	textImgReset(motif.select_info.record.TextSpriteData)
	textImgSetText(motif.select_info.record.TextSpriteData, start.f_getRecordText())

	local staticDrawList = start.updateDrawList()
	local stageResetInput = false
	start.needUpdateDrawList = false

	while not selScreenEnd do
		counter = counter + 1
		--draw clearcolor
		clearColor(motif.selectbgdef.bgclearcolor[1], motif.selectbgdef.bgclearcolor[2], motif.selectbgdef.bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(motif.selectbgdef.BGDef, 0)
		--draw title
		textImgDraw(motif.select_info.title.TextSpriteData)
		--draw portraits
		for side = 1, 2 do
			if #start.p[side].t_selTemp > 0 then
				start.f_drawPortraits(start.p[side].t_selTemp, side, motif.select_info, 'face', true)
			end
		end
		--draw cell art
		if start.needUpdateDrawList then
			staticDrawList = start.updateDrawList()
			start.needUpdateDrawList = false 
		end
		batchDraw(staticDrawList)
		--draw done cursors
		for side = 1, 2 do
			local persist = motif.select_info['p' .. side].cursor.persist 
			local totalSelected = #start.p[side].t_selected
			for k, v in pairs(start.p[side].t_selected) do
				if v.cursor ~= nil then
					--get cell coordinates
					local x = v.cursor[1]
					local y = v.cursor[2]
					local t = start.t_grid[y + 1][x + 1]
					--retrieve proper cell coordinates in case of random selection
					--TODO: doesn't work with slot feature
					--if (t.char == 'randomselect' or t.hidden == 3) --[[and not gameOption('Options.Team.Duplicates')]] then
					--	x = start.f_getCharData(v.ref).col - 1
					--	y = start.f_getCharData(v.ref).row - 1
					--	t = start.t_grid[y + 1][x + 1]
					--end
					--render only if cell is not hidden
					if t.hidden ~= 1 and t.hidden ~= 2 then
						local shouldDraw = false
						if main.coop or persist then
							shouldDraw = true
						end
						if start.p[side].selEnd and k == totalSelected then
							shouldDraw = true
						end
						if shouldDraw then
							start.f_drawCursor(v.pn, x, y, 'done', true)
						end
					end
				end
			end
		end
		--team and select menu
		if blinkCount < motif.select_info.p2.cursor.switchtime then
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
					v.selectState, DrawUpdateflag = start.f_selectMenu(side, v.cmd, v.player, member, v.selectState)
					if start.needUpdateDrawList == false then
						start.needUpdateDrawList = DrawUpdateflag
					end
					--not in palmenu
					if v.selectState == 0 and not getInput(-1, motif.select_info.cancel.key) then
						start.p[side].inPalMenu = false
					end
					--draw active cursor
					if side == 2 and motif.select_info.p2.cursor.blink then
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
					local cursorState = 'active'
					if v.selectState > 0 and motif.select_info.paletteselect > 0 then
						--cursorState when palmenu is active
						cursorState = 'done'
					end
					if v.selectState < 4 and start.f_selGrid(start.c[v.player].cell + 1).hidden ~= 1 and not start.c[v.player].blink then
						start.f_drawCursor(v.player, start.c[v.player].selX, start.c[v.player].selY, cursorState, false)
					end
				end
			end
			--delayed screen transition for the duration of face_done_anim or selection sound
			if start.p[side].screenDelay > 0 then
				if getInput(-1, motif.select_info.done.key) and not start.p[side].inPalMenu then
					start.p[side].screenDelay = 0
				else
					start.p[side].screenDelay = start.p[side].screenDelay - 1
				end
			end
			--exit select screen
			for _, v in ipairs(start.p[side].t_selCmd) do
				if not start.escFlag and (esc() or (getInput({v.cmd}, motif.select_info.cancel.key) and not start.p[side].inPalMenu)) then
					main.f_fadeReset('fadeout', motif.select_info)
					fadeOutStarted = true
					start.escFlag = true
				end
			end
		end
		--draw names
		for side = 1, 2 do
			if #start.p[side].t_selTemp > 0 then
				for i = 1, #start.p[side].t_selTemp do
					if i <= motif.select_info['p' .. side].name.num or main.coop then
						local name = ''
						if motif.select_info['p' .. side].name.num == 1 then
							name = start.f_getName(start.p[side].t_selTemp[#start.p[side].t_selTemp].ref, side)
						else
							name = start.f_getName(start.p[side].t_selTemp[i].ref, side)
						end
						textImgReset(motif.select_info['p' .. side].name.TextSpriteData)
						textImgAddPos(
							motif.select_info['p' .. side].name.TextSpriteData,
							(i - 1) * motif.select_info['p' .. side].name.spacing[1],
							(i - 1) * motif.select_info['p' .. side].name.spacing[2]
						)
						textImgSetText(motif.select_info['p' .. side].name.TextSpriteData, name)
						textImgDraw(motif.select_info['p' .. side].name.TextSpriteData)
					end
				end
			end
		end
		--team and character selection complete
		if start.p[1].selEnd and start.p[2].selEnd and start.p[1].teamEnd and start.p[2].teamEnd then
			restoreCursor = true
			if main.stageMenu and not stageEnd then --Stage select
				if not stageResetInput then
					main.f_cmdBufReset()
					stageResetInput = true
				end
				start.f_stageMenu()
				if not timerReset then
					timerSelect = motif.select_info.timer.displaytime
					timerReset = true
				end
			elseif start.p[1].screenDelay <= 0 and start.p[2].screenDelay <= 0 and not fadeOutStarted then
				main.f_fadeReset('fadeout', motif.select_info)
				fadeOutStarted = true
			end
			--draw stage portrait
			if main.stageMenu then
				--draw stage portrait background
				main.f_animPosDraw(motif.select_info.stage.portrait.bg.AnimData)
				--draw stage portrait (random)
				if stageListNo == 0 then
					main.f_animPosDraw(motif.select_info.stage.portrait.random.AnimData)
				--draw stage portrait loaded from stage SFF
				else
					main.f_animPosDraw(
						main.t_selStages[main.t_selectableStages[stageListNo]].anim_data,
						motif.select_info.stage.pos[1] + motif.select_info.stage.portrait.offset[1],
						motif.select_info.stage.pos[2] + motif.select_info.stage.portrait.offset[2]
					)
				end
				if not stageEnd then
					if getInput(-1, motif.select_info.done.key) or timerSelect == -1 then
						sndPlay(motif.Snd, motif.select_info.stage.done.snd[1], motif.select_info.stage.done.snd[2])
						stageTextData = motif.select_info.stage.done.TextSpriteData
						stageEnd = true
					elseif stageActiveCount < motif.select_info.stage.active.switchtime then --delay change
						stageActiveCount = stageActiveCount + 1
					else
						if stageActiveState then
							stageActiveState = false
							stageTextData = motif.select_info.stage.active2.TextSpriteData
						else
							stageActiveState = true
							stageTextData = motif.select_info.stage.active.TextSpriteData
						end
						stageActiveCount = 0
					end
				end
				--draw stage name
				local stage_text = motif.select_info.stage.random.text
				if stageListNo ~= 0 then
					stage_text = main.f_formatBySpec(motif.select_info.stage.text, {i = stageListNo, s = main.t_selStages[main.t_selectableStages[stageListNo]].name})
				end
				textImgReset(stageTextData)
				textImgSetText(stageTextData, stage_text)
				textImgDraw(stageTextData)
			end
		else
			--draw record text
			textImgDraw(motif.select_info.record.TextSpriteData)
		end
		--draw timer
		if motif.select_info.timer.count ~= -1 and (not start.p[1].teamEnd or not start.p[2].teamEnd or not start.p[1].selEnd or not start.p[2].selEnd or (main.stageMenu and not stageEnd)) and counter >= 0 then
			timerSelect = main.f_drawTimer(timerSelect, motif.select_info.timer)
		end
		-- hook
		hook.run("start.f_selectScreen")
		--draw layerno = 1 backgrounds
		bgDraw(motif.selectbgdef.BGDef, 1)
		--draw fadein / fadeout
		main.f_fadeAnim(motif.select_info)
		--frame transition
		if main.fadeActive or main.fadeCnt > 0 then
			main.f_cmdBufReset()
		elseif fadeOutStarted or start.escFlag then
			main.f_cmdBufReset()
			selScreenEnd = true
			break --skip last frame rendering
		end
		refresh()
	end
	return not start.escFlag
end

--;===========================================================
--; TEAM MENU
--;===========================================================
local t_teamActiveCount = {0, 0}
local t_teamActiveState = {false, false}

function start.f_teamMenu(side, t)
	if #t == 0 then
		start.p[side].teamEnd = true
		-- Team menu has no renderable entries (e.g. itemname_order hides them).
		-- Still allow character selection for this side if enabled.
		if not start.p[side].selEnd and #start.p[side].t_selCmd == 0 then
			table.insert(start.p[side].t_selCmd, {cmd = side, player = side, selectState = 0})
		end
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
			for i = 1, gameOption('Config.Players') do
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
		if #t > 1 and getInput(t_cmd, motif.select_info['p' .. side].teammenu.previous.key) then
			if start.p[side].teamMenu > 1 then
				sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.move.snd[1], motif.select_info['p' .. side].teammenu.move.snd[2])
				start.p[side].teamMenu = start.p[side].teamMenu - 1
			elseif motif.select_info.teammenu.move.wrapping then
				sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.move.snd[1], motif.select_info['p' .. side].teammenu.move.snd[2])
				start.p[side].teamMenu = #t
			end
		elseif #t > 1 and getInput(t_cmd, motif.select_info['p' .. side].teammenu.next.key) then
			if start.p[side].teamMenu < #t then
				sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.move.snd[1], motif.select_info['p' .. side].teammenu.move.snd[2])
				start.p[side].teamMenu = start.p[side].teamMenu + 1
			elseif motif.select_info.teammenu.move.wrapping then
				sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.move.snd[1], motif.select_info['p' .. side].teammenu.move.snd[2])
				start.p[side].teamMenu = 1
			end
		else
			if t[start.p[side].teamMenu].itemname == 'simul' then
				if getInput(t_cmd, motif.select_info['p' .. side].teammenu.subtract.key) then
					if start.p[side].numSimul > main.numSimul[1] then
						sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.value.snd[1], motif.select_info['p' .. side].teammenu.value.snd[2])
						start.p[side].numSimul = start.p[side].numSimul - 1
					end
				elseif getInput(t_cmd, motif.select_info['p' .. side].teammenu.add.key) then
					if start.p[side].numSimul < main.numSimul[2] then
						sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.value.snd[1], motif.select_info['p' .. side].teammenu.value.snd[2])
						start.p[side].numSimul = start.p[side].numSimul + 1
					end
				end
			elseif t[start.p[side].teamMenu].itemname == 'turns' then
				if getInput(t_cmd, motif.select_info['p' .. side].teammenu.subtract.key) then
					if start.p[side].numTurns > main.numTurns[1] then
						sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.value.snd[1], motif.select_info['p' .. side].teammenu.value.snd[2])
						start.p[side].numTurns = start.p[side].numTurns - 1
					end
				elseif getInput(t_cmd, motif.select_info['p' .. side].teammenu.add.key) then
					if start.p[side].numTurns < main.numTurns[2] then
						sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.value.snd[1], motif.select_info['p' .. side].teammenu.value.snd[2])
						start.p[side].numTurns = start.p[side].numTurns + 1
					end
				end
			elseif t[start.p[side].teamMenu].itemname == 'tag' then
				if getInput(t_cmd, motif.select_info['p' .. side].teammenu.subtract.key) then
					if start.p[side].numTag > main.numTag[1] then
						sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.value.snd[1], motif.select_info['p' .. side].teammenu.value.snd[2])
						start.p[side].numTag = start.p[side].numTag - 1
					end
				elseif getInput(t_cmd, motif.select_info['p' .. side].teammenu.add.key) then
					if start.p[side].numTag < main.numTag[2] then
						sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.value.snd[1], motif.select_info['p' .. side].teammenu.value.snd[2])
						start.p[side].numTag = start.p[side].numTag + 1
					end
				end
			elseif t[start.p[side].teamMenu].itemname == 'ratio' then
				if getInput(t_cmd, motif.select_info['p' .. side].teammenu.subtract.key) and main.selectMenu[side] then
					sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.value.snd[1], motif.select_info['p' .. side].teammenu.value.snd[2])
					if start.p[side].numRatio > 1 then
						start.p[side].numRatio = start.p[side].numRatio - 1
					else
						start.p[side].numRatio = 7
					end
				elseif getInput(t_cmd, motif.select_info['p' .. side].teammenu.add.key) and main.selectMenu[side] then
					sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.value.snd[1], motif.select_info['p' .. side].teammenu.value.snd[2])
					if start.p[side].numRatio < 7 then
						start.p[side].numRatio = start.p[side].numRatio + 1
					else
						start.p[side].numRatio = 1
					end
				end
			end
		end
		--Exit during team menu
		if not start.escFlag and (esc() or getInput(-1, motif.select_info.cancel.key)) then
			esc(false)
			main.f_fadeReset('fadeout', motif.select_info)
			fadeOutStarted = true
			start.escFlag = true
		end
		--Draw team background
		main.f_animPosDraw(motif.select_info['p' .. side].teammenu.bg.default.AnimData)
		--Draw team title
		if side == 2 and main.cpuSide[2] then
			main.f_animPosDraw(motif.select_info['p' .. side].teammenu.enemytitle.AnimData)
			textImgDraw(motif.select_info['p' .. side].teammenu.enemytitle.TextSpriteData)
		else
			main.f_animPosDraw(motif.select_info['p' .. side].teammenu.selftitle.AnimData)
			textImgDraw(motif.select_info['p' .. side].teammenu.selftitle.TextSpriteData)
		end
		--Draw team cursor
		main.f_animPosDraw(
			motif.select_info['p' .. side].teammenu.item.cursor.AnimData,
			(start.p[side].teamMenu - 1) * motif.select_info['p' .. side].teammenu.item.spacing[1],
			(start.p[side].teamMenu - 1) * motif.select_info['p' .. side].teammenu.item.spacing[2]
		)
		local teammenu = motif.select_info['p' .. side].teammenu
		local itemCfg = teammenu.item
		local valueCfg = teammenu.value
		local spacingX, spacingY = itemCfg.spacing[1], itemCfg.spacing[2]
		local uppercase = itemCfg.uppercase
		local gm = gamemode()
		for i = 1, #t do
			local x = (i - 1) * spacingX
			local y = (i - 1) * spacingY
			local itemname = t[i].itemname
			local itemText = main.f_itemnameUpper(t[i].displayname, uppercase)
			local textData = itemCfg.TextSpriteData
			--Draw team items
			if i == start.p[side].teamMenu then
				if t_teamActiveCount[side] < itemCfg.active.switchtime then --delay change
					t_teamActiveCount[side] = t_teamActiveCount[side] + 1
				else
					if t_teamActiveState[side] then
						t_teamActiveState[side] = false
					else
						t_teamActiveState[side] = true
					end
					t_teamActiveCount[side] = 0
				end
				--Draw team active item background
				if teammenu.active.bg[gm .. '-' .. itemname] ~= nil then
					main.f_animPosDraw(teammenu.active.bg[gm .. '-' .. itemname].AnimData)
				elseif teammenu.active.bg[itemname] ~= nil then
					main.f_animPosDraw(teammenu.active.bg[itemname].AnimData)
				end
				--Draw team active item font
				if t_teamActiveState[side] then
					textData = itemCfg.active2.TextSpriteData
				else
					textData = itemCfg.active.TextSpriteData
				end
			else
				--Draw team not active item background
				if teammenu.bg[gm .. '-' .. itemname] ~= nil then
					main.f_animPosDraw(teammenu.bg[gm .. '-' .. itemname].AnimData)
				elseif teammenu.bg[itemname] ~= nil then
					main.f_animPosDraw(teammenu.bg[itemname].AnimData)
				end
			end
			--Draw item font
			textImgReset(textData)
			textImgAddPos(textData, x, y)
			textImgSetText(textData, itemText)
			textImgDraw(textData)
			--Draw team icons
			if itemname == 'simul' then
				for j = 1, main.numSimul[2] do
					local vx = x + (j - 1) * valueCfg.spacing[1]
					local vy = y + (j - 1) * valueCfg.spacing[2]
					if j <= start.p[side].numSimul then
						main.f_animPosDraw(valueCfg.icon.AnimData, vx, vy)
					else
						main.f_animPosDraw(valueCfg.empty.icon.AnimData, vx, vy)
					end
				end
			elseif itemname == 'turns' then
				for j = 1, main.numTurns[2] do
					local vx = x + (j - 1) * valueCfg.spacing[1]
					local vy = y + (j - 1) * valueCfg.spacing[2]
					if j <= start.p[side].numTurns then
						main.f_animPosDraw(valueCfg.icon.AnimData, vx, vy)
					else
						main.f_animPosDraw(valueCfg.empty.icon.AnimData, vx, vy)
					end
				end
			elseif itemname == 'tag' then
				for j = 1, main.numTag[2] do
					local vx = x + (j - 1) * valueCfg.spacing[1]
					local vy = y + (j - 1) * valueCfg.spacing[2]
					if j <= start.p[side].numTag then
						main.f_animPosDraw(valueCfg.icon.AnimData, vx, vy)
					else
						main.f_animPosDraw(valueCfg.empty.icon.AnimData, vx, vy)
					end
				end
			elseif itemname == 'ratio' and start.p[side].teamMenu == i and main.selectMenu[side] then
				main.f_animPosDraw(teammenu['ratio' .. start.p[side].numRatio].icon.AnimData, x, y)
			end
		end
		--Confirmed team selection
		if getInput(t_cmd, motif.select_info['p' .. side].teammenu.done.key) or timerSelect == -1 then
			timerSelect = motif.select_info.timer.displaytime
			sndPlay(motif.Snd, motif.select_info['p' .. side].teammenu.done.snd[1], motif.select_info['p' .. side].teammenu.done.snd[2])
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

--===========================================================
--; PALETTE SELECT
--===========================================================
LoadedPals = {}
-- Tracks which characters have already had their palettes loaded to avoid redundant loading
local function ifCharPalsLoaded(ref)
	for _, v in ipairs(LoadedPals) do
		if v == ref then
			return true
		end
	end
	table.insert(LoadedPals, ref)
	return false
end
-- Loads palettes for a character if needed, prepares the animation, and applies the palette
function start.loadPalettes(a, ref, pal)
	if not ifCharPalsLoaded(ref) then
		animLoadPalettes(a, ref)
	end
	local srcAnim = a
	a = animPrepare(a, ref)
	animApplyVel(a, srcAnim)
	a = changeColorPalette(a, pal)
	return a
end

--===========================================================
-- Draw Palette Menu
--===========================================================
function start.f_palMenuDraw(side, member, curIdx, validIdx ,maxIdx)
	local charData = start.f_getCharData(start.p[side].t_selTemp[member].ref)
	if not charData or not charData.pal then return end
	local palIndex = start.p[side].t_selTemp[member].pal
	local totalPals = #charData.pal
	local pn = 2 * (member - 1) + side
	local pCfg = f_getMotifP(motif.select_info, pn, side)
	local displayText = (curIdx == maxIdx) and pCfg.palmenu.random.text or tostring(validIdx)
	-- bg
	main.f_animPosDraw(pCfg.palmenu.bg.AnimData)
	-- draw number
	textImgReset(pCfg.palmenu.number.TextSpriteData)
	textImgSetText(pCfg.palmenu.number.TextSpriteData, displayText)
	textImgDraw(pCfg.palmenu.number.TextSpriteData)
	-- draw text
	textImgReset(pCfg.palmenu.text.TextSpriteData)
	textImgDraw(pCfg.palmenu.text.TextSpriteData)
end

--returns a random palette (using synced RNG)
function start.f_randomPal(charRef, validPals)
	start.shufflePals = start.shufflePals or {}
	start.shufflePals[charRef] = start.shufflePals[charRef] or {}

	if #start.shufflePals[charRef] == 0 then
		local last = start.lastRandomPal and start.lastRandomPal[charRef]
		local t = {}
		for _, v in ipairs(validPals) do
			table.insert(t, v)
		end
		start.f_shuffleTable(t, last)
		start.shufflePals[charRef] = t
	end
	-- draw one palette from the bag
	local result = table.remove(start.shufflePals[charRef])
	-- store last drawn palette
	start.lastRandomPal = start.lastRandomPal or {}
	start.lastRandomPal[charRef] = result
	return result
end

local function resolvePalConflict(side, charRef, pal)
	local charData = start.f_getCharData(charRef)
	if not charData or not charData.pal then
		return pal
	end
	local usedPals = {}
	for s = 1, 2 do
		for _, sel in ipairs(start.p[s].t_selected) do
			if sel.ref == charRef and sel.pal then
				usedPals[sel.pal] = true
			end
		end
	end
	-- if the chosen palette is not used, keep it
	if not usedPals[pal] then
		return validatePal(pal, charRef)
	end
	-- if it's in use, try to find the next free one
	local maxPal = gameOption('Config.PaletteMax')
	for i = pal + 1, maxPal do
	if not usedPals[i] then
			return validatePal(i, charRef)
		end
	end
	for i = 1, pal - 1 do
		if not usedPals[i] then
			return validatePal(i, charRef)
		end
	end

	return validatePal(pal, charRef)
end

local function applyPalette(sel, charData, palIndex)
	if sel.face_data then
		local srcAnim = sel.face_data
		sel.face_data = changeColorPalette(sel.face_data, palIndex)
	end
	if sel.face2_data then
		local srcAnim = sel.face2_data
		sel.face2_data = changeColorPalette(sel.face2_data, palIndex)
	end
end

-- palette select menu
function start.f_palMenu(side, cmd, player, member, selectState)
	local st = start.p[side].t_selTemp[member]
	local charRef = st.ref
	local charData = start.f_getCharData(charRef)
	local pn = 2 * (member - 1) + side
	local pCfg = f_getMotifP(motif.select_info, pn, side)
	-- initialize palette list and index if character changed or not yet set
	if st.validPalsCharRef ~= charRef or not st.validPals then
		local valid, seen, cur = {}, {}, validatePal(1, charRef)
		valid[1], seen[cur] = cur, true
		for i = 1, #charData.pal do
			local nextp = validatePal(cur + 1, charRef)
			if seen[nextp] then break end
			table.insert(valid, nextp)
			seen[nextp], cur = true, nextp
		end
		st.validPals, st.validPalsCharRef = valid, charRef
		-- set current index to match current palette (or default to first)
		local curPal = st.pal or valid[1]
		st.currentIdx = 1
		for i, p in ipairs(valid) do
			if p == curPal then st.currentIdx = i; break end
		end
	end

	local validPals = st.validPals
	local curIdx = st.currentIdx or 1
	local pal = st.pal or validPals[curIdx]
	local maxIdx = #validPals + 1
	start.p[side].inPalMenu = true

	-- accept selection
	if getInput({cmd}, motif.select_info['p' .. side].palmenu.done.key) or timerSelect == -1 then
		pal = (curIdx == maxIdx) and (start.c[player].randPalPreview or start.f_randomPal(charRef, validPals)) or validPals[curIdx]
		st.pal, st.currentIdx = pal, curIdx

		-- done anim after pal confirmation - primary face
		local done_anim = pCfg.face.done.anim
		local preview_anim = pCfg.palmenu.preview.anim
		if done_anim ~= preview_anim then
			if st.face_anim ~= done_anim and (main.coop or motif.select_info['p' .. side].face.num > 1 or main.f_tableLength(start.p[side].t_selected) + 1 == start.p[side].numChars) then
				local a = start.f_animGet(start.c[player].selRef, side, member, pCfg.face.done, pCfg.face, false, st.face_data)
				if a then
					st.face_data = start.loadPalettes(a, charRef, pal)
					animUpdate(st.face_data)
					start.p[side].screenDelay = math.min(120, math.max(start.p[side].screenDelay, animGetLength(st.face_data)))
				end
			end
		end

		-- face2 "done" anim after pal confirmation
		local done_anim2 = pCfg.face2.done.anim
		if st.face2_anim ~= done_anim2 and (main.coop or motif.select_info['p' .. side].face2.num > 1 or main.f_tableLength(start.p[side].t_selected) + 1 == start.p[side].numChars) then
			local a = start.f_animGet(start.c[player].selRef, side, member, pCfg.face2.done, pCfg.face2, false, st.face2_data)
			if a then
				st.face2_data = start.loadPalettes(a, charRef, pal)
				animUpdate(st.face2_data)
				start.p[side].screenDelay = math.min(120, math.max(start.p[side].screenDelay, animGetLength(st.face2_data)))
			end
		end
		selectState = 3
		start.f_playWave(start.c[player].selRef, 'cursor', motif.select_info['p' .. side].select.snd[1], motif.select_info['p' .. side].select.snd[2])
		sndPlay(motif.Snd, motif.select_info['p' .. side].palmenu.done.snd[1], motif.select_info['p' .. side].palmenu.done.snd[2])
	 -- next palette
	elseif getInput({cmd}, motif.select_info['p' .. side].palmenu.next.key) then
		curIdx = (curIdx == maxIdx) and 1 or curIdx + 1
		st.currentIdx = curIdx
		if curIdx < maxIdx then
			applyPalette(st, charData, validPals[curIdx])
		end
		sndPlay(motif.Snd, motif.select_info['p' .. side].palmenu.value.snd[1], motif.select_info['p' .. side].palmenu.value.snd[2])
	-- previous palette
	elseif getInput({cmd}, motif.select_info['p' .. side].palmenu.previous.key) then
		curIdx = (curIdx == 1) and maxIdx or curIdx - 1
		st.currentIdx = curIdx
		if curIdx < maxIdx then
			applyPalette(st, charData, validPals[curIdx])
		end
		sndPlay(motif.Snd, motif.select_info['p' .. side].palmenu.value.snd[1], motif.select_info['p' .. side].palmenu.value.snd[2])
	-- cancel
	elseif getInput({cmd}, motif.select_info['p' .. side].palmenu.cancel.key) then
		st.face_data = start.f_animGet(start.c[player].selRef, side, member, motif.select_info['p' .. pn].face, nil, true, st.face_data)
		st.face2_data = start.f_animGet(start.c[player].selRef, side, member, motif.select_info['p' .. pn].face2, nil, true, st.face2_data)
		selectState = 0
		st.currentIdx = nil
		st.validPals = nil
		sndPlay(motif.Snd, motif.select_info['p' .. side].palmenu.cancel.snd[1], motif.select_info['p' .. side].palmenu.cancel.snd[2])
	end
	-- random hotkey
	if getInput({cmd}, motif.select_info['p' .. side].palmenu.random.key) then
		curIdx, st.currentIdx = maxIdx, maxIdx
	end
	-- random preview update
	if st.currentIdx == maxIdx then
		if not start.c[player].randPalCnt or start.c[player].randPalCnt <= 0 then
			start.c[player].randPalCnt = motif.select_info.palmenu.random.switchtime
			start.c[player].randPalPreview = start.f_randomPal(charRef, validPals)
			if motif.select_info.palmenu.random.applypal then
				applyPalette(st, charData, start.c[player].randPalPreview)
			else
				applyPalette(st, charData, 1)
			end
			sndPlay(motif.Snd, motif.select_info['p' .. side].palmenu.value.snd[1], motif.select_info['p' .. side].palmenu.value.snd[2])
		else
			start.c[player].randPalCnt = start.c[player].randPalCnt - 1
		end
	end
	start.f_palMenuDraw(side, member, curIdx, validPals[curIdx], maxIdx)
	return selectState
end

--;===========================================================
--; SELECT MENU
--;===========================================================
function start.f_selectMenu(side, cmd, player, member, selectState)
	local needUpdateDrawList = false
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
		local pn = 2 * (member - 1) + side
		local pCfg = f_getMotifP(motif.select_info, pn, side)
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
					if gameOption('Options.Team.Duplicates') or t_reservedChars[side][start.t_grid[selY + 1][selX + 1].char_ref] == nil then
						start.c[player].selX = selX
						start.c[player].selY = selY
					end
					start.p[side].t_cursor[member] = nil
				end
			end
			--calculate current position
			start.c[player].selX, start.c[player].selY = start.f_cellMovement(start.c[player].selX, start.c[player].selY, cmd, side, start.f_getCursorData(player).cursor.move.snd)
			start.c[player].cell = start.c[player].selX + motif.select_info.columns * start.c[player].selY
			start.c[player].selRef = start.f_selGrid(start.c[player].cell + 1).char_ref
			-- temp data not existing yet
			if start.p[side].t_selTemp[member] == nil then
				table.insert(start.p[side].t_selTemp, {
					ref = start.c[player].selRef,
					cell = start.c[player].cell,
					inRandom = false,
					face_anim = pCfg.face.anim,
					face_data = start.f_animGet(start.c[player].selRef, side, member, pCfg.face, nil, true),
					face2_anim = pCfg.face2.anim,
					face2_data = start.f_animGet(start.c[player].selRef, side, member, pCfg.face2, nil, true),
				})
			else
				local updateAnim = false
				local slotSelected,slotChanged = start.f_slotSelected(start.c[player].cell + 1, side, cmd, player, start.c[player].selX, start.c[player].selY)
				local timerExpired = motif.select_info.timer.count ~= -1 and timerSelect == -1
				needUpdateDrawList = slotChanged
				local velCopy = false
				if timerExpired then
					if start.c[player].selRef == nil or main.t_selChars[start.c[player].selRef + 1] == nil then
						start.c[player].selRef = start.f_randomChar(side)
					end
				end
				-- cursor changed position or character change within current slot
				if start.p[side].t_selTemp[member].cell ~= start.c[player].cell or start.p[side].t_selTemp[member].ref ~= start.c[player].selRef then
					--start.p[side].t_selTemp[member].pal = 1
					start.p[side].t_selTemp[member].ref = start.c[player].selRef
					start.p[side].t_selTemp[member].cell = start.c[player].cell
					start.p[side].t_selTemp[member].face_anim = pCfg.face.anim
					start.p[side].t_selTemp[member].face2_anim = pCfg.face2.anim
					if start.f_getCursorData(player).cursor.reset then
						resetCursorData(player, cursorActive, 'active')
					end
					updateAnim = true
				end
				-- cursor at randomselect cell
				if start.f_selGrid(start.c[player].cell + 1).char == 'randomselect' or start.f_selGrid(start.c[player].cell + 1).hidden == 3 then
					local wasRandom = start.p[side].t_selTemp[member].inRandom
					start.p[side].t_selTemp[member].inRandom = true
					velCopy = wasRandom
					-- re-entering random slot: reset random overlay anims so sliding restarts
					if not wasRandom then
						local pData = pCfg
						if pData.face2.random then
							animReset(pData.face2.random.AnimData)
							animUpdate(pData.face2.random.AnimData)
						end
						if pData.face.random then
							animReset(pData.face.random.AnimData)
							animUpdate(pData.face.random.AnimData)
						end
					end
					if start.c[player].randCnt > 0 then
						start.c[player].randCnt = start.c[player].randCnt - 1
						start.c[player].selRef = start.c[player].randRef
					else
						if motif.select_info.random.move.snd.cancel then
							sndStop(motif.Snd, start.f_getCursorData(player).random.move.snd[1], start.f_getCursorData(player).random.move.snd[2])
						end
						sndPlay(motif.Snd, start.f_getCursorData(player).random.move.snd[1], start.f_getCursorData(player).random.move.snd[2])
						start.c[player].randCnt = motif.select_info.cell.random.switchtime
						start.c[player].selRef = start.f_randomChar(side)
						if start.c[player].randRef ~= start.c[player].selRef or start.p[side].t_selTemp[member].face_data == nil then
							updateAnim = true
							start.c[player].randRef = start.c[player].selRef
						end
					end
				else
					start.p[side].t_selTemp[member].inRandom = false
				end
				-- update anim data
				if updateAnim then
					local face_data = velCopy and start.p[side].t_selTemp[member].face_data or nil
					local face2_data = velCopy and start.p[side].t_selTemp[member].face2_data or nil
					start.p[side].t_selTemp[member].face_data = start.f_animGet(start.c[player].selRef, side, member, pCfg.face, nil, true, face_data)
					start.p[side].t_selTemp[member].face2_data = start.f_animGet(start.c[player].selRef, side, member, pCfg.face2, nil, true, face2_data)
				end
				-- cell selected or select screen timer reached 0
				if (slotSelected and start.f_selGrid(start.c[player].cell + 1).char ~= nil and start.f_selGrid(start.c[player].cell + 1).hidden ~= 2) or timerExpired then
					if motif.select_info.paletteselect ~= 0 then
						timerSelect = motif.select_info.timer.displaytime
					end
					sndPlay(motif.Snd, start.f_getCursorData(player).cursor.done.default.snd[1], start.f_getCursorData(player).cursor.done.default.snd[2])
					if motif.select_info.paletteselect == 0 then
						start.f_playWave(start.c[player].selRef, 'cursor', motif.select_info['p' .. side].select.snd[1], motif.select_info['p' .. side].select.snd[2])
					end
					if motif.select_info.paletteselect > 0 then
						resetCursorData(player, cursorActive, 'done')
					end
					start.p[side].t_selTemp[member].pal = main.f_btnPalNo(cmd)
					start.p[side].t_selTemp[member].inRandom = false
					if start.p[side].t_selTemp[member].pal == nil or start.p[side].t_selTemp[member].pal == 0 then
						start.p[side].t_selTemp[member].pal = 1
					end
					-- done anim helper
					local function setDoneAnim(ref, side, member, params, velParams, targetField)
						local a = start.f_animGet(ref, side, member, params, velParams, false, start.p[side].t_selTemp[member][targetField])
						if a then
							start.p[side].t_selTemp[member][targetField] = a
							start.p[side].screenDelay = math.min(120, math.max(start.p[side].screenDelay, animGetLength(a)))
						end
					end
					-- if select anim differs from done anim and coop or pX.face.num allows to display more than 1 portrait or it's the last team member
					local done_anim = pCfg.face.done.anim
					local done_anim2 = pCfg.face2.done.anim
					local palmenu_preview_anim = pCfg.palmenu.preview.anim
					local face_anim = start.p[side].t_selTemp[member].face_anim
					local face2_anim = start.p[side].t_selTemp[member].face2_anim
					local canShow = main.coop or motif.select_info['p' .. side].face.num > 1 or main.f_tableLength(start.p[side].t_selected) + 1 == start.p[side].numChars
					local canShow2 = main.coop or motif.select_info['p' .. side].face2.num > 1 or main.f_tableLength(start.p[side].t_selected) + 1 == start.p[side].numChars
					-- primary face "done" / preview
					if face_anim ~= done_anim and canShow then
						if motif.select_info.paletteselect == 0 and done_anim ~= -1 then
							setDoneAnim(start.c[player].selRef, side, member, pCfg.face.done, pCfg.face, 'face_data')
						elseif palmenu_preview_anim ~= -1 and motif.select_info.paletteselect ~= 0 then
							start.f_playWave(start.c[player].selRef, 'cursor', motif.select_info['p' .. side].palmenu.preview.snd[1], motif.select_info['p' .. side].palmenu.preview.snd[2])
							setDoneAnim(start.c[player].selRef, side, member, pCfg.palmenu.preview, pCfg.face, 'face_data')
						end
					end
					-- face2 "done" anim
					if face2_anim ~= done_anim and canShow2 and done_anim2 ~= -1 then
						setDoneAnim(start.c[player].selRef, side, member, pCfg.face2.done, pCfg.face2, 'face2_data')
					end

					start.p[side].t_selTemp[member].ref = start.c[player].selRef
					local charRef = start.p[side].t_selTemp[member].ref
					local charData = start.f_getCharData(charRef)
					local pal = start.p[side].t_selTemp[member].pal
					local finalPal

					if motif.select_info.paletteselect == 1 then
						finalPal = 1
					elseif motif.select_info.paletteselect == 2 then
						finalPal = pal
					elseif motif.select_info.paletteselect == 3 then
						finalPal = start.f_keyPalMap(charRef, pal)
					else
						finalPal = start.f_keyPalMap(charRef, pal)
					end

					-- resolve visual palette conflict
					finalPal = resolvePalConflict(side, charRef, finalPal)

					if motif.select_info.paletteselect > 0 then
						start.p[side].t_selTemp[member].pal = finalPal
					end

					if start.p[side].t_selTemp[member].face_data ~= nil then
						local applyFlag = pCfg.face.applypal
						if applyFlag then
							start.p[side].t_selTemp[member].face_data = start.loadPalettes(start.p[side].t_selTemp[member].face_data, charRef, finalPal)
							animUpdate(start.p[side].t_selTemp[member].face_data)
						end
					end
					if start.p[side].t_selTemp[member].face2_data ~= nil then
						local applyFlag = pCfg.face2.applypal
						if applyFlag then
							start.p[side].t_selTemp[member].face2_data = start.loadPalettes(start.p[side].t_selTemp[member].face2_data, charRef, finalPal)
							animUpdate(start.p[side].t_selTemp[member].face2_data)
						end
					end
					main.f_cmdBufReset(cmd)
					selectState = 1
				end
			end
		--selection menu
		elseif selectState == 1 then
			if motif.select_info.paletteselect and motif.select_info.paletteselect > 0 then
				selectState = start.f_palMenu(side, cmd, player, member, selectState)
			else
				selectState = 3
			end
		--confirm selection
		elseif selectState == 3 then
			local valid = {1,2,3}
			local finalPal
			for _, v in ipairs(valid) do
				if motif.select_info.paletteselect == v then
					finalPal = start.p[side].t_selTemp[member].pal
					start.p[side].inPalMenu = false
					break
				end
			end
			finalPal = finalPal or start.f_selectPal(start.c[player].selRef, start.p[side].t_selTemp[member].pal)
			finalPal = resolvePalConflict(side, start.c[player].selRef, finalPal)
			applyPalette(start.p[side].t_selTemp[member], start.f_getCharData(start.c[player].selRef), finalPal)
			start.p[side].t_selected[member] = {
				ref = start.c[player].selRef,
				pal = finalPal,
				pn = start.f_getPlayerNo(side, member),
				cursor = {start.c[player].selX, start.c[player].selY},
				ratioLevel = start.f_getRatio(side),
			}
			if not gameOption('Options.Team.Duplicates') then
				t_reservedChars[side][start.c[player].selRef] = true
			end
			-- If we just confirmed a pick while hovering Random Select, force the random preview to advance for the next team member.
			-- This prevents mashing select from picking the same random character repeatedly.
			local cellIdx = (start.c[player].cell or -1) + 1
			if cellIdx > 0 then
				local g = start.f_selGrid(cellIdx)
				if g.char == 'randomselect' or g.hidden == 3 then
					start.c[player].randCnt = 0
					start.c[player].randRef = nil
				end
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
						start.p[2].teamEnd = false
					else
						start.p[2].teamEnd = false
					end
				end
				start.p[side].selEnd = true
			elseif not gameOption('Options.Team.Duplicates') and start.t_grid[start.c[player].selY + 1][start.c[player].selX + 1].char ~= 'randomselect' then
				local t_dirs = {'F', 'B', 'D', 'U'}
				if start.c[player].selY + 1 >= motif.select_info.rows then --next row not visible on the screen
					t_dirs = {'F', 'B', 'U', 'D'}
				end
				for _, v in ipairs(t_dirs) do
					local selX, selY = start.f_cellMovement(start.c[player].selX, start.c[player].selY, cmd, side, start.f_getCursorData(player).cursor.move.snd, v)
					if start.t_grid[selY + 1][selX + 1].char ~= nil and (selX ~= start.c[player].selX or selY ~= start.c[player].selY) then
						start.c[player].selX, start.c[player].selY = selX, selY
						break
					end
				end
			end
			if not start.p[1].teamEnd or not start.p[2].teamEnd or not start.p[1].selEnd or not start.p[2].selEnd then
				timerSelect = motif.select_info.timer.displaytime
			end
			if main.coop and (side == 1 or gamemode('versuscoop')) then --remaining members are controlled by different players
				selectState = 4
			elseif not start.p[side].selEnd then --next member controlled by this player should become selectable
				selectState = 0
			end
		end
	end
	return selectState, needUpdateDrawList
end

--;===========================================================
--; STAGE MENU
--;===========================================================
function start.f_stageMenu()
	local n = stageListNo
	local randomMode = {
		[0] = { init = 1, min = 1 }, -- disabled
		[1] = { init = 0, min = 0 }, -- default
		[2] = { init = 1, min = 0 }, -- random at the 'end'
	}

	local r = randomMode[motif.select_info.stage_randomselect] or randomMode[1]
	local stageListIdx = r.init
	local stageListMinIdx = r.min
	local stageListMaxIdx = #main.t_selectableStages

	if getInput(-1, motif.select_info.cell.left.key) then
		sndPlay(motif.Snd, motif.select_info.stage.move.snd[1], motif.select_info.stage.move.snd[2])
		stageListNo = stageListNo - 1
		if stageListNo < stageListMinIdx then stageListNo = stageListMaxIdx end
	elseif getInput(-1, motif.select_info.cell.right.key) then
		sndPlay(motif.Snd, motif.select_info.stage.move.snd[1], motif.select_info.stage.move.snd[2])
		stageListNo = stageListNo + 1
		if stageListNo > stageListMaxIdx then stageListNo = stageListMinIdx end
	elseif getInput(-1, motif.select_info.cell.up.key) then
		sndPlay(motif.Snd, motif.select_info.stage.move.snd[1], motif.select_info.stage.move.snd[2])
		for i = 1, 10 do
			stageListNo = stageListNo - 1
			if stageListNo < stageListMinIdx then stageListNo = stageListMaxIdx end
		end
	elseif getInput(-1, motif.select_info.cell.down.key) then
		sndPlay(motif.Snd, motif.select_info.stage.move.snd[1], motif.select_info.stage.move.snd[2])
		for i = 1, 10 do
			stageListNo = stageListNo + 1
			if stageListNo > stageListMaxIdx then stageListNo = stageListMinIdx end
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
function start.f_selectVersus(active, t_orderSelect)
	start.t_orderRemap = {{}, {}}
	for side = 1, 2 do
		-- populate order remap table with default values
		for i = 1, #start.p[side].t_selected do
			table.insert(start.t_orderRemap[side], i)
		end
		-- prevent order select if not enabled in screenpack or if team size = 1
		if t_orderSelect[side] then
			t_orderSelect[side] = motif.vs_screen.orderselect.enabled and #start.p[side].t_selected > 1
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
	textImgReset(motif.vs_screen.match.TextSpriteData)
	textImgSetText(motif.vs_screen.match.TextSpriteData, string.format(motif.vs_screen.match.text, matchno()))
	bgReset(motif.versusbgdef.BGDef)
	main.f_fadeReset('fadein', motif.vs_screen)
	local fadeOutStarted = false
	playBgm({source = "motif.vs"})
	start.f_resetTempData(motif.vs_screen, '')
	start.f_playWave(getStageNo(), 'stage', motif.vs_screen.stage.snd[1], motif.vs_screen.stage.snd[2])
	local counter = 0 - motif.vs_screen.fadein.time
	local done = (not t_orderSelect[1] and not t_orderSelect[2]) -- both sides having order disabled
		or (not t_orderSelect[1] and main.cpuSide[2]) -- left side with disabled order, right side controlled by CPU
		or (not t_orderSelect[2] and main.cpuSide[1]) -- right side with disabled order, left side controlled by CPU
		or (main.cpuSide[1] and main.cpuSide[2]) -- both sides controlled by CPU
	local timerActive = not done
	local timerCount = 0
	local escFlag = false
	local t_order = {{}, {}}
	local t_icon = {false, false}
	local selStageNo = getStageNo()
	while true do
		local snd = false
		-- for each team side member
		for side = 1, 2 do
			for k, v in ipairs(start.p[side].t_selected) do
				local pn = 2 * (k - 1) + side
				local pCfg = f_getMotifP(motif.vs_screen, pn, side)
				-- until loading flag is set
				if not v.loading then
					-- if not valid for order selection or CPU or doesn't have key for this member assigned, or order timer run out
					if not t_orderSelect[side] or main.cpuSide[side] or (#pCfg.key == 0 and #t_order[side] == k - 1) or timerCount == -1 then
						table.insert(t_order[side], k)
						-- if it's the last unordered team member
						if #start.p[side].t_selected == #t_order[side] then
							-- randomize CPU side team order (if valid for order selection)
							if main.cpuSide[side] and t_orderSelect[side] then
								main.f_tableShuffle(t_order[side])
							end
							-- confirm char selection (starts loading immediately if gameOption('Config.BackgroundLoading') is true)
							for _, member in ipairs(t_order[side]) do
								if not start.p[side].t_selected[member].loading then
									selectChar(side, start.p[side].t_selected[member].ref, start.p[side].t_selected[member].pal)
									start.p[side].t_selected[member].loading = true
								end
							end
							t_icon[side] = nil
							-- play sound if timer run out
							if not snd and timerCount == -1 then
								sndPlay(motif.Snd, motif.vs_screen['p' .. side].value.snd[1], motif.vs_screen['p' .. side].value.snd[2])
								snd = true
							end
						end
					elseif getInput({side}, pCfg.key) or (#start.p[side].t_selected == #t_order[side] + 1) then
						table.insert(t_order[side], k)
						-- confirm char selection (starts loading immediately if gameOption('Config.BackgroundLoading') is true)
						selectChar(side, v.ref, v.pal)
						v.loading = true
						-- if it's the last unordered team member
						if #start.p[side].t_selected == #t_order[side] then
							t_icon[side] = nil
						end
						-- play sound only once in particular frame
						if not snd then
							sndPlay(motif.Snd, motif.vs_screen['p' .. side].value.snd[1], motif.vs_screen['p' .. side].value.snd[2])
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
					local pn = 2 * (member - 1) + side
					local pCfg = f_getMotifP(motif.vs_screen, pn, side)
					-- primary face "done" anim
					local done_anim = pCfg.done.anim
					if done_anim ~= -1 and start.p[side].t_selTemp[member].face_anim ~= done_anim then
						start.p[side].t_selTemp[member].face_data = start.f_animGet(v.ref, side, member, pCfg.done, pCfg, false, start.p[side].t_selTemp[member].face_data)
					end
					-- face2 "done" anim
					local done_anim2 = pCfg.face2.done.anim
					if done_anim2 ~= -1 and start.p[side].t_selTemp[member].face2_anim ~= done_anim2 then
						start.p[side].t_selTemp[member].face2_data = start.f_animGet(v.ref, side, member, pCfg.face2.done, pCfg.face2, false, start.p[side].t_selTemp[member].face2_data)
					end
				end
				if t_orderSelect[side] then
					t_icon[side] = true
				end
			end
			counter = motif.vs_screen.time - motif.vs_screen.done.time
			done = true
		end
		counter = counter + 1
		--draw clearcolor
		clearColor(motif.versusbgdef.bgclearcolor[1], motif.versusbgdef.bgclearcolor[2], motif.versusbgdef.bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(motif.versusbgdef.BGDef, 0)
		--draw portraits and order icons
		for side = 1, 2 do
			start.f_drawPortraits(main.f_remapTable(start.p[side].t_selTemp, start.t_orderRemap[side]), side, motif.vs_screen, '', false, t_icon[side])
		end
		--draw order values
		for side = 1, 2 do
			if t_orderSelect[side] then
				for i = 1, math.min(#start.p[side].t_selected, motif.vs_screen['p' .. side].num) do
					local pn = 2 * (i - 1) + side
					local pCfg = f_getMotifP(motif.vs_screen, pn, side)
					if i > #t_order[side] and #start.p[side].t_selected > #t_order[side] then
						main.f_animPosDraw(
							pCfg.value.empty.icon.AnimData,
							(i - 1) * motif.vs_screen['p' .. side].value.icon.spacing[1],
							(i - 1) * motif.vs_screen['p' .. side].value.icon.spacing[2]
						)
					else
						main.f_animPosDraw(
							pCfg.value.icon.AnimData,
							(i - 1) * motif.vs_screen['p' .. side].value.icon.spacing[1],
							(i - 1) * motif.vs_screen['p' .. side].value.icon.spacing[2]
						)
					end
				end
			end
		end
		--draw names
		for side = 1, 2 do
			for k, v in ipairs(main.f_remapTable(start.p[side].t_selTemp, start.t_orderRemap[side])) do
				if k <= motif.vs_screen['p' .. side].name.num or main.coop then
					textImgReset(motif.vs_screen['p' .. side].name.TextSpriteData)
					textImgAddPos(
						motif.vs_screen['p' .. side].name.TextSpriteData,
						(k - 1) * motif.vs_screen['p' .. side].name.spacing[1],
						(k - 1) * motif.vs_screen['p' .. side].name.spacing[2]
					)
					textImgSetText(motif.vs_screen['p' .. side].name.TextSpriteData, start.f_getName(v.ref, side))
					textImgDraw(motif.vs_screen['p' .. side].name.TextSpriteData)
				end
			end
		end
		--draw stage portrait
		if selStageNo then
			--draw stage portrait background
			main.f_animPosDraw(motif.vs_screen.stage.portrait.bg.AnimData)
			--draw stage portrait loaded from stage SFF
			if main.t_selStages[selStageNo].vs_anim_data then
				main.f_animPosDraw(
					main.t_selStages[selStageNo].vs_anim_data,
					motif.vs_screen.stage.pos[1] + motif.vs_screen.stage.portrait.offset[1],
					motif.vs_screen.stage.pos[2] + motif.vs_screen.stage.portrait.offset[2]
				)
			end
		end
		--draw stage name
		if selStageNo and main.t_selStages[selStageNo] then
			textImgReset(motif.vs_screen.stage.TextSpriteData)
			local stage_text = main.f_formatBySpec(motif.vs_screen.stage.text, {i = selStageNo, s = main.t_selStages[selStageNo].name})
			textImgSetText(motif.vs_screen.stage.TextSpriteData, stage_text)
			textImgDraw(motif.vs_screen.stage.TextSpriteData)
		end
		--draw match counter
		if main.motif.versusmatchno then
			textImgDraw(motif.vs_screen.match.TextSpriteData)
		end
		--draw timer
		if not done and motif.vs_screen.timer.count ~= -1 and timerActive and counter >= 0 then
			timerCount, timerActive = main.f_drawTimer(timerCount, motif.vs_screen.timer)
		end
		--draw layerno = 1 backgrounds
		bgDraw(motif.versusbgdef.BGDef, 1)
		-- hook
		hook.run("start.f_selectVersus")
		--draw fadein / fadeout
		for side = 1, 2 do
			if not fadeOutStarted and (
				-- Wait for order select to finish before vs_screen.time can end the screen.
				(counter >= motif.vs_screen.time and (not (t_orderSelect[1] or t_orderSelect[2]) or done))
				or (not main.cpuSide[side] and getInput({side}, motif.vs_screen.skip.key))
				or (done and getInput({side}, motif.vs_screen.done.key))
				) then
				main.f_fadeReset('fadeout', motif.vs_screen)
				fadeOutStarted = true
				break
			end
		end
		main.f_fadeAnim(motif.vs_screen)
		--frame transition
		if not escFlag and (esc() or getInput(-1, motif.vs_screen.cancel.key)) then
			esc(false)
			main.f_fadeReset('fadeout', motif.vs_screen)
			fadeOutStarted = true
			escFlag = true
		end
		if main.fadeActive or main.fadeCnt > 0 then
			main.f_cmdBufReset()
		elseif fadeOutStarted or start.escFlag then
			main.f_cmdBufReset()
			clearColor(motif.versusbgdef.bgclearcolor[1], motif.versusbgdef.bgclearcolor[2], motif.versusbgdef.bgclearcolor[3])
			break --skip last frame rendering
		end
		refresh()
	end
	esc(escFlag) --force Esc detection
	return not escFlag
end

--loading loop called after versus screen is finished
function start.f_selectLoading(musicParams)
	clearAllSound()
	local parts = {}
	if musicParams and musicParams ~= "" then
		parts[#parts + 1] = musicParams
	end
	local function addParam(k, v)
		if v == nil then return end
		parts[#parts + 1] = k .. "=" .. tostring(v)
	end
	addParam("charparam.ai", main.charparam.ai)
	addParam("charparam.arcadepath", main.charparam.arcadepath)
	addParam("charparam.music", main.charparam.music)
	addParam("charparam.rounds", main.charparam.rounds)
	addParam("charparam.single", main.charparam.single)
	addParam("charparam.stage", main.charparam.stage)
	addParam("charparam.time", main.charparam.time)
	for side = 1, 2 do
		for member, v in ipairs(start.p[side].t_selected) do
			if not v.loading then
				selectChar(side, v.ref, v.pal)
				v.loading = true
			end
			-- fold overrideCharData() payload into loadStart() params
			local lifeRatio, attackRatio
			if v.ratioLevel then
				lifeRatio = gameOption("Options.Ratio.Level" .. v.ratioLevel .. ".Life")
				attackRatio = gameOption("Options.Ratio.Level" .. v.ratioLevel .. ".Attack")
			end
			local pfx = "p" .. side .. "." .. member .. "."
			addParam(pfx .. "life", v.life)
			addParam(pfx .. "lifemax", v.lifeMax)
			addParam(pfx .. "power", v.power)
			addParam(pfx .. "dizzypoints", v.dizzyPoints)
			addParam(pfx .. "guardpoints", v.guardPoints)
			addParam(pfx .. "ratiolevel", v.ratioLevel)
			addParam(pfx .. "liferatio", v.lifeRatio or lifeRatio)
			addParam(pfx .. "attackratio", v.attackRatio or attackRatio)
			addParam(pfx .. "existed", v.existed)
		end
	end
	addParam("persistlife", main.persistLife)
	addParam("persistmusic", main.persistMusic)
	addParam("persistrounds", main.persistRounds)
	local params = table.concat(parts, ", ")
	if gameOption('Debug.DumpLuaTables') then main.f_printTable(params, "debug/loadStartParams.txt") end
	loadStart(params)
end

return start


local select = {}

setSelColRow(motif.select_info.columns, motif.select_info.rows)

--not used for now
--setShowEmptyCells(motif.select_info.showemptyboxes)
--setRandomSpr(motif.selectbgdef.spr_data, motif.select_info.cell_random_spr[1], motif.select_info.cell_random_spr[2], motif.select_info.cell_random_scale[1], motif.select_info.cell_random_scale[2])
--setCellSpr(motif.selectbgdef.spr_data, motif.select_info.cell_bg_spr[1], motif.select_info.cell_bg_spr[2], motif.select_info.cell_bg_scale[1], motif.select_info.cell_bg_scale[2])

setSelCellSize(motif.select_info.cell_size[1] + motif.select_info.cell_spacing, motif.select_info.cell_size[2] + motif.select_info.cell_spacing)
setSelCellScale(motif.select_info.portrait_scale[1], motif.select_info.portrait_scale[2])

--default team count after starting the game
local p1NumTurns = 2
local p1NumSimul = 2
local p1NumTag = 2
local p2NumTurns = 2
local p2NumSimul = 2
local p2NumTag = 2
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
local t_p1Selected = {}
local t_p2Selected = {}
local t_roster = {}
local t_aiRamp = {}
local continue = false
local p1Cell = false
local p2Cell = false
local p1TeamEnd = false
local p1SelEnd = false
local p2TeamEnd = false
local p2SelEnd = false
local selScreenEnd = false
local stageEnd = false
local coopEnd = false
local restoreTeam = false
local teamMode = 0
local numChars = 0
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
local winCnt = 0
local looseCnt = 0
local clearTime = 0
local matchTime = 0
local p1FaceX = 0
local p1FaceY = 0
local p2FaceX = 0
local p2FaceY = 0
local p1TeamMode = 0
local p2TeamMode = 0
local lastMatch = 0
local stageNo = 0
local stageList = 0

local cnt = motif.select_info.columns + 1
local row = 1
local col = 0
local t_grid = {}
t_grid[row] = {}
for i = 1, (motif.select_info.rows + motif.select_info.rows_scrolling) * motif.select_info.columns do
	if i == cnt then
		row = row + 1
		cnt = cnt + motif.select_info.columns
		t_grid[row] = {}
	end
	col = #t_grid[row] + 1
	t_grid[row][col] = {num = i - 1, x = (col - 1) * (motif.select_info.cell_size[1] + motif.select_info.cell_spacing), y = (row - 1) * (motif.select_info.cell_size[2] + motif.select_info.cell_spacing)}
	if main.t_selChars[i].char ~= nil then
		t_grid[row][col].char = main.t_selChars[i].char
		t_grid[row][col].hidden = main.t_selChars[i].hidden
		main.t_selChars[i].row = row
		main.t_selChars[i].col = col
	end
end

--;===========================================================
--; COMMON FUNCTIONS
--;===========================================================
function select.f_setZoom()
	local zoom = config.ZoomActive
	if main.t_selChars[t_p2Selected[1].cel + 1].zoom ~= nil then
		if main.t_selChars[t_p2Selected[1].cel + 1].zoom == 1 then
			zoom = true
		else
			zoom = false
		end
	elseif main.t_selStages[stageNo].zoom ~= nil then
		if main.t_selChars[stageNo].zoom == 1 then
			zoom = true
		else
			zoom = false
		end
	end
	setZoom(zoom)
	local zoomMin = config.ZoomMin
	if main.t_selStages[stageNo].zoommin ~= nil then
		zoomMin = main.t_selStages[stageNo].zoommin
	end
	setZoomMin(zoomMin)
	local zoomMax = config.ZoomMax
	if main.t_selStages[stageNo].zoommax ~= nil then
		zoomMax = main.t_selStages[stageNo].zoommax
	end
	setZoomMax(zoomMax)
	local zoomSpeed = config.ZoomSpeed
	if main.t_selStages[stageNo].zoomspeed ~= nil then
		zoomSpeed = main.t_selStages[stageNo].zoomspeed
	end
	setZoomSpeed(zoomSpeed)
end

function select.f_makeRoster()
	t_roster = {}
	local t = {}
	local cnt = 0
	--Arcade
	if main.gameMode == 'arcade' or main.gameMode == 'teamcoop' or main.gameMode == 'netplayteamcoop' then
		if p2TeamMode == 0 then --Single
			t = main.t_selOptions.arcademaxmatches
		else --Team
			t = main.t_selOptions.teammaxmatches
		end
		for i = 1, #t do --for each order number
			cnt = t[i] * p2NumChars --set amount of matches to get from the table
			if cnt > 0 and main.t_orderChars[i] ~= nil then --if it's more than 0 and there are characters with such order
				while cnt > 0 do --do the following until amount of matches for particular order is reached
					main.f_shuffleTable(main.t_orderChars[i]) --randomize characters table
					for j = 1, #main.t_orderChars[i] do --loop through chars associated with that particular order
						t_roster[#t_roster + 1] = main.t_orderChars[i][j] --and add such character into new table
						cnt = cnt - 1
						if cnt == 0 then --but only if amount of matches for particular order has not been reached yet
							break
						end
					end
				end
			end
		end
	--Survival / Boss Rush / VS 100 Kumite
	else
		if main.gameMode == 'survival' or main.gameMode == 'survivalcoop' or main.gameMode == 'netplaysurvivalcoop' then
			t = main.t_randomChars
			cnt = #t
			local i = 0
			while cnt / p2NumChars ~= math.ceil(cnt / p2NumChars) do --not integer
				i = i + 1
				cnt = #t + i
			end
		elseif main.gameMode == 'bossrush' then
			t = main.t_bossChars
			cnt = #t
			local i = 0
			while cnt / p2NumChars ~= math.ceil(cnt / p2NumChars) do
				i = i + 1
				cnt = #t + i
			end
		elseif main.gameMode == '100kumite' then
			t = main.t_randomChars
			cnt = 100 * p2NumChars
		end
		while cnt > 0 do
			main.f_shuffleTable(t)
			for i = 1, #t do
				t_roster[#t_roster + 1] = t[i]
				cnt = cnt - 1
				if cnt == 0 then
					break
				end
			end
		end
	end
	main.f_printTable(t_roster, 'debug/t_roster.txt')
end

function select.f_aiRamp()
	local start_match = 0
	local start_diff = 0
	local end_match = 0
	local end_diff = 0
	t_aiRamp = {}
	--Arcade
	if main.gameMode == 'arcade' or main.gameMode == 'teamcoop' or main.gameMode == 'netplayteamcoop' then
		if p2TeamMode == 0 then --Single
			start_match = main.t_selOptions.arcadestart.wins
			start_diff = main.t_selOptions.arcadestart.offset
			end_match =  main.t_selOptions.arcadeend.wins
			end_diff = main.t_selOptions.arcadeend.offset
		else --Team
			start_match = main.t_selOptions.teamstart.wins
			start_diff = main.t_selOptions.teamstart.offset
			end_match =  main.t_selOptions.teamend.wins
			end_diff = main.t_selOptions.teamend.offset
		end
	elseif main.gameMode == 'survival' or main.gameMode == 'survivalcoop' or main.gameMode == 'netplaysurvivalcoop' then
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
	for i = 1, lastMatch do
		if i - 1 <= start_match then
			t_aiRamp[#t_aiRamp + 1] = startAI
		elseif i - 1 <= end_match then
			local curMatch = i - (start_match + 1)
			t_aiRamp[#t_aiRamp + 1] = math.floor(curMatch * (endAI - startAI) / (end_match - start_match) + startAI)
		else
			t_aiRamp[#t_aiRamp + 1] = endAI
		end
	end
	main.f_printTable(t_aiRamp, 'debug/t_aiRamp.txt')
end

function select.f_difficulty(player, offset)
	local t = {}
	if player % 2 ~= 0 then --odd value
		local pos = math.floor(player / 2 + 0.5)
		t = main.t_selChars[t_p1Selected[pos].cel + 1]
	else --even value
		local pos = math.floor(player / 2)
		t = main.t_selChars[t_p2Selected[pos].cel + 1]
	end
	if t.ai ~= nil then
		return t.ai
	else
		return config.Difficulty + offset
	end
end

function select.f_aiLevel()
	--Offset
	local offset = 0
	if config.AIRamping and main.gameMode == 'arcade' or main.gameMode == 'teamcoop' or main.gameMode == 'netplayteamcoop' or main.gameMode == 'survival' or main.gameMode == 'survivalcoop' or main.gameMode == 'netplaysurvivalcoop' then
		offset = t_aiRamp[matchNo] - config.Difficulty
	end
	--Player 1
	if main.coop then
		setCom(1, 0)
		setCom(3, 0)
	elseif p1TeamMode == 0 then --Single
		if main.p1In == 1 and not main.aiFight then
			setCom(1, 0)
		else
			setCom(1, select.f_difficulty(1, offset))
		end
	elseif p1TeamMode == 1 and config.SimulMode then --Simul
		if main.p1In == 1 and not main.aiFight then
			setCom(1, 0)
		else
			setCom(1, select.f_difficulty(1, offset))
		end
		for i = 3, p1NumChars * 2 do
			if i % 2 ~= 0 then --odd value
				setCom(i, select.f_difficulty(i, offset))
			end
		end
	elseif p1TeamMode == 2 then --Turns
		for i = 1, p1NumChars * 2 do
			if i % 2 ~= 0 then
				if main.p1In == 1 and not main.aiFight then
					setCom(i, 0)
				else
					setCom(i, select.f_difficulty(i, offset))
				end
			end
		end
	else --Tag
		for i = 1, p1NumChars * 2 do
			if i % 2 ~= 0 then --odd value
				if main.p1In == 1 and not main.aiFight then
					setCom(i, 0)
				else
					setCom(i, select.f_difficulty(i, offset))
				end
			end
		end
	end
	--Player 2
	if p2TeamMode == 0 then --Single
		if main.p2In == 2 and not main.aiFight and not main.coop then
			setCom(2, 0)
		else
			setCom(2, select.f_difficulty(2, offset))
		end
	elseif p2TeamMode == 1 and config.SimulMode then --Simul
		if main.p2In == 2 and not main.aiFight and not main.coop then
			setCom(2, 0)
		else
			setCom(2, select.f_difficulty(2, offset))
		end
		for i = 4, p2NumChars * 2 do
			if i % 2 == 0 then --even value
				setCom(i, select.f_difficulty(i, offset))
			end
		end
	elseif p2TeamMode == 2 then --Turns
		for i = 2, p2NumChars * 2 do
			if i % 2 == 0 then
				if main.p2In == 2 and not main.aiFight and not main.coop then
					setCom(i, 0)
				else
					setCom(i, select.f_difficulty(i, offset))
				end
			end
		end
	else --Tag
		for i = 2, p2NumChars * 2 do
			if i % 2 == 0 then --even value
				if main.p2In == 2 and not main.aiFight and not main.coop then
					setCom(i, 0)
				else
					setCom(i, select.f_difficulty(i, offset))
				end
			end
		end
	end
end

function select.f_assignMusic()
	local track = ''
	if main.stageMenu then
		if main.t_selStages[stageNo].music ~= nil then
			track = math.random(1, #main.t_selStages[stageNo].music)
			track = main.t_selStages[stageNo].music[track].bgmusic
		end
	else
		if main.t_selChars[t_p2Selected[1].cel + 1].music ~= nil then
			track = math.random(1, #main.t_selChars[t_p2Selected[1].cel + 1].music)
			track = main.t_selChars[t_p2Selected[1].cel + 1].music[track].bgmusic
		elseif main.t_selStages[stageNo].music ~= nil then
			track = math.random(1, #main.t_selStages[stageNo].music)
			track = main.t_selStages[stageNo].music[track].bgmusic
		end
		stageEnd = true
	end
	playBGM(track)
end

function select.f_selectStage()
	if main.t_selChars[t_p2Selected[1].cel + 1].stage ~= nil then
		stageNo = math.random(1, #main.t_selChars[t_p2Selected[1].cel + 1].stage)
		stageNo = main.t_selChars[t_p2Selected[1].cel + 1].stage[stageNo]
	else
		stageNo = main.t_includeStage[math.random(1, #main.t_includeStage)]
	end
	setStage(stageNo)
	selectStage(stageNo)
end

function select.f_randomPal(cell)
	--table with pal numbers already assigned
	local t = {}
	for i = 1, #t_p1Selected do
		if t_p1Selected[i].cel == cell then
			t[#t + 1] = t_p1Selected[i].pal
		end
	end
	for i = 1, #t_p2Selected do
		if t_p2Selected[i].cel == cell then
			t[#t + 1] = t_p2Selected[i].pal
		end
	end
	--table with pal numbers not assigned yet (or all if there are not enough pals for unique appearance of all characters)
	local t2 = {}
	for i = 1, #main.t_selChars[cell + 1].pal do
		if t[main.t_selChars[cell + 1].pal[i]] == nil or #t >= #main.t_selChars[cell + 1].pal then
			t2[#t2 + 1] = main.t_selChars[cell + 1].pal[i]
		end
	end
	return t2[math.random(1, #t2)]
end

function select.f_drawName(t, data, font, offsetX, offsetY, scaleX, scaleY, spacingX, spacingY, active_font, active_row)
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
			main.f_updateTextImg(
				data,
				motif.font_data[f[1]],
				f[2],
				f[3],
				main.f_getName(t[i].cel),
				x + (i - 1) * spacingX,
				offsetY + (i - 1) * spacingY,
				scaleX,
				scaleY
			)
			textImgDraw(data)
		end
	end
end

function select.f_drawPortrait(t, offsetX, offsetY, facing, scaleX, scaleY, spacingX, spacingY, limit, func)
	if facing == -1 then offsetX = offsetX + 1 end --fix for wrong offset after flipping sprites
	for i = #t, 1, -1 do
		if i <= limit then
			if func == 'select' then
				drawPortrait(t[i].cel, offsetX + (i - 1) * spacingX, offsetY + (i - 1) * spacingY, facing * scaleX, scaleY)
			elseif func == 'versus' then
				drawVersusPortrait(t[i].cel, offsetX + (i - 1) * spacingX, offsetY + (i - 1) * spacingY, facing * scaleX, scaleY)
			elseif func == 'victory' then
				drawVictoryPortrait(t[i].cel, offsetX + (i - 1) * spacingX, offsetY + (i - 1) * spacingY, facing * scaleX, scaleY)
			end
		end
	end
end

function select.f_cellMovement(selX, selY, cmd, faceOffset, rowOffset, snd)
	local tmpX = selX
	local tmpY = selY
	local tmpFace = faceOffset
	local tmpRow = rowOffset
	local found = false
	if commandGetState(cmd, 'u') then
		for i = 1, motif.select_info.rows do
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
			if (t_grid[selY + 1][selX + 1].char ~= nil and t_grid[selY + 1][selX + 1].hidden ~= 2) or motif.select_info.moveoveremptyboxes == 1 then
				break
			elseif motif.select_info.searchemptyboxesup ~= 0 then
				found, selX = select.f_searchEmptyBoxes(motif.select_info.searchemptyboxesup, selX, selY)
				if found then
					break
				end
			end
		end
	elseif commandGetState(cmd, 'd') then
		for i = 1, motif.select_info.rows do
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
			if (t_grid[selY + 1][selX + 1].char ~= nil and t_grid[selY + 1][selX + 1].hidden ~= 2) or motif.select_info.moveoveremptyboxes == 1 then
				break
			elseif motif.select_info.searchemptyboxesdown ~= 0 then
				found, selX = select.f_searchEmptyBoxes(motif.select_info.searchemptyboxesdown, selX, selY)
				if found then
					break
				end
			end
		end
	elseif commandGetState(cmd, 'l') then
		for i = 1, motif.select_info.columns do
			selX = selX - 1
			if selX < 0 then
				if wrappingX then
					selX = motif.select_info.columns - 1
				else
					selX = tmpX
				end
			end
			if (t_grid[selY + 1][selX + 1].char ~= nil and t_grid[selY + 1][selX + 1].hidden ~= 2) or motif.select_info.moveoveremptyboxes == 1 then
				break
			end
		end
	elseif commandGetState(cmd, 'r') then
		for i = 1, motif.select_info.columns do
			selX = selX + 1
			if selX >= motif.select_info.columns then
				if wrappingX then
					selX = 0
				else
					selX = tmpX
				end
			end
			if (t_grid[selY + 1][selX + 1].char ~= nil and t_grid[selY + 1][selX + 1].hidden ~= 2) or motif.select_info.moveoveremptyboxes == 1 then
				break
			end
		end
	end
	if tmpX ~= selX or tmpY ~= selY then
		if tmpRow ~= rowOffset then
			select.f_resetGrid()
		end
		sndPlay(motif.files.snd_data, snd[1], snd[2])
	end
	return selX, selY, faceOffset, rowOffset
end

function select.f_searchEmptyBoxes(direction, x, y)
	local tmpX = x
	local found = false
	if direction > 0 then --right
		while true do
			x = x + 1
			if x >= motif.select_info.columns then
				x = tmpX
				break
			elseif t_grid[y + 1][x + 1].char ~= nil and t_grid[selY + 1][selX + 1].hidden ~= 2 then
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
			elseif t_grid[y + 1][x + 1].char ~= nil and t_grid[selY + 1][selX + 1].hidden ~= 2 then
				found = true
				break
			end
		end
	end
	return found, x
end

function select.f_unlockChar(char, flag)
	--setHiddenFlag(char, flag) --not used for now
	main.t_selChars[char + 1].hidden = flag
	t_grid[main.t_selChars[char + 1].row][main.t_selChars[char + 1].col].hidden = flag
	select.f_resetGrid()
end

function select.f_resetGrid()
	select.t_drawFace = {}
	for row = 1, motif.select_info.rows do
		for col = 1, motif.select_info.columns do
			if t_grid[row + p1RowOffset][col].char == 'randomselect' or t_grid[row + p1RowOffset][col].hidden == 3 then
				select.t_drawFace[#select.t_drawFace + 1] = {d = 1, p1 = t_grid[row + p1RowOffset][col].num, p2 = t_grid[row + p2RowOffset][col].num, x1 = p1FaceX + t_grid[row][col].x, x2 = p2FaceX + t_grid[row][col].x, y1 = p1FaceY + t_grid[row][col].y, y2 = p2FaceY + t_grid[row][col].y}
			elseif t_grid[row + p1RowOffset][col].char ~= nil and t_grid[row + p1RowOffset][col].hidden == 0 then
				select.t_drawFace[#select.t_drawFace + 1] = {d = 2, p1 = t_grid[row + p1RowOffset][col].num, p2 = t_grid[row + p2RowOffset][col].num, x1 = p1FaceX + t_grid[row][col].x, x2 = p2FaceX + t_grid[row][col].x, y1 = p1FaceY + t_grid[row][col].y, y2 = p2FaceY + t_grid[row][col].y}
			elseif motif.select_info.showemptyboxes == 1 then
				select.t_drawFace[#select.t_drawFace + 1] = {d = 0, p1 = t_grid[row + p1RowOffset][col].num, p2 = t_grid[row + p2RowOffset][col].num, x1 = p1FaceX + t_grid[row][col].x, x2 = p2FaceX + t_grid[row][col].x, y1 = p1FaceY + t_grid[row][col].y, y2 = p2FaceY + t_grid[row][col].y}
			end
		end
	end
end

function select.f_selectReset()
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
	select.f_resetGrid()
	if main.gameMode == 'netplayversus' or main.gameMode == 'netplayteamcoop' or main.gameMode == 'netplaysurvivalcoop' then
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
	p2TeamEnd = false
	p2SelEnd = false
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
	p1NumChars = 1
	p2NumChars = 1
	matchNo = 0
	setMatchNo(matchNo)
end

--;===========================================================
--; SIMPLE LOOP (VS MODE, TRAINING, WATCH, BONUS GAMES)
--;===========================================================
function select.f_selectSimple()
	p1SelX = motif.select_info.p1_cursor_startcell[2]
	p1SelY = motif.select_info.p1_cursor_startcell[1]
	p2SelX = motif.select_info.p2_cursor_startcell[2]
	p2SelY = motif.select_info.p2_cursor_startcell[1]
	p1FaceOffset = 0
	p2FaceOffset = 0
	p1RowOffset = 0
	p2RowOffset = 0
	stageList = 0
	main.f_cmdInput()
	while true do
		main.f_resetBG(motif.select_info, motif.selectbgdef, motif.music.select_bgm)
		select.f_selectReset()
		selectStart()
		while not selScreenEnd do
			if esc() then
				sndPlay(motif.files.snd_data, motif.select_info.cancel_snd[1], motif.select_info.cancel_snd[2])
				main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm)
				return
			end
			select.f_selectScreen()
		end
		select.f_aiLevel()
		if not main.stageMenu then
			select.f_selectStage()
		end
		select.f_selectVersus()
		if esc() then break end
		select.f_setZoom()
		select.f_assignMusic()
		loadStart()
		winner = game()
		main.f_cmdInput()
		refresh()
	end
end

--;===========================================================
--; ADVANCE LOOP (ARCADE, TEAM CO-OP, SURVIVAL, SURVIVAL CO-OP, VS 100 KUMITE, BOSS RUSH)
--;===========================================================
function select.f_selectAdvance()
	p1SelX = motif.select_info.p1_cursor_startcell[2]
	p1SelY = motif.select_info.p1_cursor_startcell[1]
	p2SelX = motif.select_info.p2_cursor_startcell[2]
	p2SelY = motif.select_info.p2_cursor_startcell[1]
	p1FaceOffset = 0
	p2FaceOffset = 0
	p1RowOffset = 0
	p2RowOffset = 0
	winner = 0
	winCnt = 0
	looseCnt = 0
	clearTime = 0
	matchTime = 0
	main.f_cmdInput()
	select.f_selectReset()
	stageEnd = true
	while true do
		main.f_resetBG(motif.select_info, motif.selectbgdef, motif.music.select_bgm)
		selectStart()
		while not selScreenEnd do
			if esc() then
				sndPlay(motif.files.snd_data, motif.select_info.cancel_snd[1], motif.select_info.cancel_snd[2])
				main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm)
				return
			end
			select.f_selectScreen()
		end
		--first match
		if matchNo == 0 then
			--coop swap
			if main.coop then
				p1TeamMode = 1
				p1NumChars = 2
				setTeamMode(1, p1TeamMode, p1NumChars)
				t_p1Selected[2] = {cel = t_p2Selected[1].cel, pal = t_p2Selected[1].pal}
			end
			--generate roster
			select.f_makeRoster()
			lastMatch = #t_roster / p2NumChars
			matchNo = 1
			--generate AI ramping table
			select.f_aiRamp()
			--intro
			if main.gameMode == 'arcade' or main.gameMode == 'teamcoop' or main.gameMode == 'netplayteamcoop' then
				local tPos = main.t_selChars[t_p1Selected[1].cel + 1]
				if tPos.intro ~= nil and main.f_fileExists(tPos.intro) then
					storyboard.f_storyboard(tPos.intro)
				end
			end
		--player exit the match via ESC in VS 100 Kumite mode
		elseif winner == -1 and main.gameMode == '100kumite' then
			--counter
			looseCnt = looseCnt + 1
			--result
			select.f_result('lost')
			--game over
			if motif.game_over_screen.enabled == 1 and motif.game_over_screen.storyboard ~= '' then
				storyboard.f_storyboard(motif.game_over_screen.storyboard)
			end
			--intro
			if motif.files.intro_storyboard ~= '' then
				storyboard.f_storyboard(motif.files.intro_storyboard)
			end
			main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm)
			return
		--player won (also if lost in VS 100 Kumite)
		elseif winner == 1 or main.gameMode == '100kumite' then
			--counter
			if winner == 1 then
				winCnt = winCnt + 1
			else --only true in VS 100 Kumite mode
				looseCnt = looseCnt + 1
			end
			--victory screen
			if main.gameMode == 'arcade' or main.gameMode == 'teamcoop' or main.gameMode == 'netplayteamcoop' then
				if main.t_selChars[t_p2Selected[1].cel + 1].winscreen == nil or main.t_selChars[t_p2Selected[1].cel + 1].winscreen == 1 then
					select.f_selectVictory()
				end
			end
			--no more matches left
			if matchNo == lastMatch then
				--ending
				if main.gameMode == 'arcade' or main.gameMode == 'teamcoop' or main.gameMode == 'netplayteamcoop' then
					local tPos = main.t_selChars[t_p1Selected[1].cel + 1]
					if tPos.ending ~= nil and main.f_fileExists(tPos.ending) then
						storyboard.f_storyboard(tPos.ending)
					elseif motif.default_ending.enabled == 1 and motif.default_ending.storyboard ~= '' then
						storyboard.f_storyboard(motif.default_ending.storyboard)
					end
				end
				--result
				select.f_result(true)
				--credits
				if motif.end_credits.enabled == 1 and motif.end_credits.storyboard ~= '' then
					storyboard.f_storyboard(motif.end_credits.storyboard)
				end
				--game over
				if motif.game_over_screen.enabled == 1 and motif.game_over_screen.storyboard ~= '' then
					storyboard.f_storyboard(motif.game_over_screen.storyboard)
				end
				--intro
				if motif.files.intro_storyboard ~= '' then
					storyboard.f_storyboard(motif.files.intro_storyboard)
				end
				main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm)
				return
			--next match available
			else
				matchNo = matchNo + 1
			end
		--player lost and doesn't have any credits left
		elseif main.credits == 0 then
			--counter
			looseCnt = looseCnt + 1
			--victory screen
			if main.gameMode == 'arcade' or main.gameMode == 'teamcoop' or main.gameMode == 'netplayteamcoop' then
				if winner >= 1 and (main.t_selChars[t_p2Selected[1].cel + 1].winscreen == nil or main.t_selChars[t_p2Selected[1].cel + 1].winscreen == 1) then
					select.f_selectVictory()
				end
			end
			--result
			select.f_result(false)
			--game over
			if motif.game_over_screen.enabled == 1 and motif.game_over_screen.storyboard ~= '' then
				storyboard.f_storyboard(motif.game_over_screen.storyboard)
			end
			--intro
			if motif.files.intro_storyboard ~= '' then
				storyboard.f_storyboard(motif.files.intro_storyboard)
			end
			main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm)
			return
		--player lost but can continue
		else
			--counter
			looseCnt = looseCnt + 1
			--victory screen
			if winner >= 1 and (main.t_selChars[t_p2Selected[1].cel + 1].winscreen == nil or main.t_selChars[t_p2Selected[1].cel + 1].winscreen == 1) then
				select.f_selectVictory()
			end
			--continue screen
			select.f_continue()
			if not continue then
				--game over
				if motif.continue_screen.external_gameover == 1 and motif.game_over_screen.storyboard ~= '' then
					storyboard.f_storyboard(motif.game_over_screen.storyboard)
				end
				--intro
				if motif.files.intro_storyboard ~= '' then
					storyboard.f_storyboard(motif.files.intro_storyboard)
				end
				main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm)
				return
			end
			if config.ContSelection then --true if 'Char change at Continue' option is enabled
				t_p1Selected = {}
				p1SelEnd = false
				if main.coop then
					p1NumChars = 1
					numChars = p2NumChars
					p2NumChars = 1
					t_p2Selected = {}
					p2SelEnd = false
				end
				selScreenEnd = false
				while not selScreenEnd do
					if esc() then
						sndPlay(motif.files.snd_data, motif.select_info.cancel_snd[1], motif.select_info.cancel_snd[2])
						main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm)
						return
					end
					select.f_selectScreen()
				end
			elseif esc() then
				sndPlay(motif.files.snd_data, motif.select_info.cancel_snd[1], motif.select_info.cancel_snd[2])
				main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm)
				return
			end
		end
		--coop swap
		if main.coop then
			remapInput(3,2) --P2 controls assigned to P3 character
			if winner == -1 or winner == 2 then
				p1NumChars = 2
				p2NumChars = numChars
				t_p1Selected[2] = {cel = t_p2Selected[1].cel, pal = t_p2Selected[1].pal}
			end
		end
		--assign enemy team
		t_p2Selected = {}
		local shuffle = true
		for i = 1, p2NumChars do
			if i == 1 and (main.gameMode == 'arcade' or main.gameMode == 'teamcoop' or main.gameMode == 'netplayteamcoop') and main.t_selChars[t_p1Selected[1].cel + 1][matchNo] ~= nil then
				p2Cell = main.t_charDef[main.t_selChars[t_p1Selected[1].cel + 1][matchNo]]
				shuffle = false
			else
				p2Cell = t_roster[matchNo * p2NumChars - i + 1]
			end
			local updateAnim = true
			for j = 1, #t_p2Selected do
				if t_p2Selected[j].cel == p2Cell then
					updateAnim = false
				end
			end
			t_p2Selected[#t_p2Selected + 1] = {cel = p2Cell, pal = select.f_randomPal(p2Cell), up = updateAnim}
			if shuffle then
				main.f_shuffleTable(t_p2Selected)
			end
		end
		--Team conversion to Single match if bonus paramvalue on any opponents is detected
		if p2NumChars > 1 then
			for i = 1, #t_p2Selected do
				if main.t_selChars[t_p2Selected[i].cel + 1].bonus ~= nil and main.t_selChars[t_p2Selected[i].cel + 1].bonus == 1 then
					teamMode = p2TeamMode
					numChars = p2NumChars
					p2TeamMode = 0
					p2NumChars = 1
					setTeamMode(2, 0, 1)
					p2Cell = main.t_charDef[main.t_selChars[t_p2Selected[i].cel + 1].char]
					t_p2Selected = {}
					t_p2Selected[1] = {cel = p2Cell, pal = select.f_randomPal(p2Cell), up = true}
					restoreTeam = true
					break
				end
			end
		end
		setMatchNo(matchNo)
		select.f_aiLevel()
		if not main.stageMenu then
			select.f_selectStage()
		end
		select.f_selectVersus()
		if esc() then break end
		select.f_setZoom()
		matchTime = os.clock()
		select.f_assignMusic()
		loadStart()
		winner = game()
		matchTime = os.clock() - matchTime
		clearTime = clearTime + matchTime
		--restore P2 Team settings if needed
		if restoreTeam then
			p2TeamMode = teamMode
			p2NumChars = numChars
			setTeamMode(2, p2TeamMode, p2NumChars)
			restoreTeam = false
		end
		resetRemapInput()
		main.f_cmdInput()
		--main.f_printTable(_G)
		refresh()
	end
end

--;===========================================================
--; TOURNAMENT LOOP
--;===========================================================
function select.f_selectTournament()
	p1SelX = motif.select_info.p1_cursor_startcell[2]
	p1SelY = motif.select_info.p1_cursor_startcell[1]
	p2SelX = motif.select_info.p2_cursor_startcell[2]
	p2SelY = motif.select_info.p2_cursor_startcell[1]
	p1FaceOffset = 0
	p2FaceOffset = 0
	p1RowOffset = 0
	p2RowOffset = 0
	stageList = 0
	main.f_cmdInput()
	while true do
		main.f_resetBG(motif.tournament_info, motif.tournamentbgdef, motif.music.tournament_bgm)
		select.f_selectReset()
		while not selScreenEnd do
			if esc() then
				sndPlay(motif.files.snd_data, motif.select_info.cancel_snd[1], motif.select_info.cancel_snd[2])
				main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm)
				return
			end
			select.f_selectTournamentScreen()
		end
		select.f_aiLevel()
		select.f_selectVersus()
		select.f_setZoom()
		select.f_assignMusic()
		loadStart()
		winner = game()
		main.f_cmdInput()
		refresh()
	end
end

--;===========================================================
--; TOURNAMENT SCREEN
--;===========================================================
function select.f_selectTournamentScreen()
	--draw clearcolor
	animDraw(motif.tournamentbgdef.bgclearcolor_data)
	--draw layerno = 0 backgrounds
	main.f_drawBG(motif.tournamentbgdef.bg_data, motif.tournamentbgdef.bg, 0, motif.tournamentbgdef.timer)
	
	--draw layerno = 1 backgrounds
	main.f_drawBG(motif.tournamentbgdef.bg_data, motif.tournamentbgdef.bg, 1, motif.tournamentbgdef.timer)
	--draw fadein
	animDraw(motif.tournament_info.fadein_data)
	animUpdate(motif.tournament_info.fadein_data)
	--update timer
	motif.tournamentbgdef.timer = motif.tournamentbgdef.timer + 1
	--end loop
	main.f_cmdInput()
	refresh()
end

--;===========================================================
--; SELECT SCREEN
--;===========================================================
local txt_p1Name = main.f_createTextImg(motif.font_data[motif.select_info.p1_name_font[1]], motif.select_info.p1_name_font[2], motif.select_info.p1_name_font[3], '', 0, 0, motif.select_info.p1_name_font_scale[1], motif.select_info.p1_name_font_scale[2])
local p1RandomCount = 0
local p1RandomPortrait = 0
if #main.t_randomChars > 0 then p1RandomPortrait = main.t_randomChars[math.random(1, #main.t_randomChars)] end
local txt_p2Name = main.f_createTextImg(motif.font_data[motif.select_info.p2_name_font[1]], motif.select_info.p2_name_font[2], motif.select_info.p2_name_font[3], '', 0, 0, motif.select_info.p2_name_font_scale[1], motif.select_info.p2_name_font_scale[2])
local p2RandomCount = 0
local p2RandomPortrait = 0
if #main.t_randomChars > 0 then p2RandomPortrait = main.t_randomChars[math.random(1, #main.t_randomChars)] end

function select.f_selectScreen()
	--draw clearcolor
	animDraw(motif.selectbgdef.bgclearcolor_data)
	--draw layerno = 0 backgrounds
	main.f_drawBG(motif.selectbgdef.bg_data, motif.selectbgdef.bg, 0, motif.selectbgdef.timer)
	--draw title
	textImgDraw(main.txt_mainSelect)
	if p1Cell then
		--draw p1 portrait
		local t_portrait = {}
		if #t_p1Selected < p1NumChars then
			if main.t_selChars[p1Cell + 1].char == 'randomselect' or main.t_selChars[p1Cell + 1].hidden == 3 then
				if p1RandomCount < motif.select_info.cell_random_switchtime then
					p1RandomCount = p1RandomCount + 1
				elseif #main.t_randomChars > 0 then
					p1RandomPortrait = main.t_randomChars[math.random(1, #main.t_randomChars)]
					p1RandomCount = 0
				end
				sndPlay(motif.files.snd_data, motif.select_info.p1_random_move_snd[1], motif.select_info.p1_random_move_snd[2])
				t_portrait[1] = {cel = p1RandomPortrait}
			elseif main.t_selChars[p1Cell + 1].hidden ~= 2 then
				t_portrait[1] = {cel = p1Cell}
			end
		end
		for i = #t_p1Selected, 1, -1 do
			if motif.select_info.p1_face_num > #t_portrait then
				t_portrait[#t_portrait + 1] = {cel = t_p1Selected[i].cel}
			end
		end
		select.f_drawPortrait(
			main.f_reversedTable(t_portrait),
			motif.select_info.p1_face_offset[1],
			motif.select_info.p1_face_offset[2],
			motif.select_info.p1_face_facing,
			motif.select_info.p1_face_scale[1],
			motif.select_info.p1_face_scale[2],
			motif.select_info.p1_face_spacing[1],
			motif.select_info.p1_face_spacing[2],
			#t_portrait,
			'select'
		)
	end
	if p2Cell then
		--draw p2 portrait
		local t_portrait = {}
		if #t_p2Selected < p2NumChars then
			if main.t_selChars[p2Cell + 1].char == 'randomselect' or main.t_selChars[p2Cell + 1].hidden == 3 then
				if p2RandomCount < motif.select_info.cell_random_switchtime then
					p2RandomCount = p2RandomCount + 1
				elseif #main.t_randomChars > 0 then
					p2RandomPortrait = main.t_randomChars[math.random(1, #main.t_randomChars)]
					p2RandomCount = 0
				end
				sndPlay(motif.files.snd_data, motif.select_info.p2_random_move_snd[1], motif.select_info.p2_random_move_snd[2])
				t_portrait[1] = {cel = p2RandomPortrait}
			elseif main.t_selChars[p2Cell + 1].hidden ~= 2 then
				t_portrait[1] = {cel = p2Cell}
			end
		end
		for i = #t_p2Selected, 1, -1 do
			if motif.select_info.p2_face_num > #t_portrait then
				t_portrait[#t_portrait + 1] = {cel = t_p2Selected[i].cel}
			end
		end
		select.f_drawPortrait(
			main.f_reversedTable(t_portrait),
			motif.select_info.p2_face_offset[1],
			motif.select_info.p2_face_offset[2],
			motif.select_info.p2_face_facing,
			motif.select_info.p2_face_scale[1],
			motif.select_info.p2_face_scale[2],
			motif.select_info.p2_face_spacing[1],
			motif.select_info.p2_face_spacing[2],
			#t_portrait,
			'select'
		)
	end
	--draw cell art (slow for large rosters, this will be likely moved to 'drawFace' function in future)
	for i = 1, #select.t_drawFace do
		main.f_animPosDraw(motif.select_info.cell_bg_data, select.t_drawFace[i].x1, select.t_drawFace[i].y1) --draw cell background
		if select.t_drawFace[i].d == 1 then --draw random cell
			main.f_animPosDraw(motif.select_info.cell_random_data, select.t_drawFace[i].x1, select.t_drawFace[i].y1)
		elseif select.t_drawFace[i].d == 2 then --draw face cell
			drawSmallPortrait(select.t_drawFace[i].p1, select.t_drawFace[i].x1, select.t_drawFace[i].y1, motif.select_info.portrait_scale[1], motif.select_info.portrait_scale[2])
		end
		if main.p2Faces and motif.select_info.double_select == 1 then --P2 side grid enabled
			main.f_animPosDraw(motif.select_info.cell_bg_data, select.t_drawFace[i].x2, select.t_drawFace[i].y2) --draw cell background
			if select.t_drawFace[i].d == 1 then --draw random cell
				main.f_animPosDraw(motif.select_info.cell_random_data, select.t_drawFace[i].x2, select.t_drawFace[i].y2)
			elseif select.t_drawFace[i].d == 2 then --draw face cell
				drawSmallPortrait(select.t_drawFace[i].p2, select.t_drawFace[i].x2, select.t_drawFace[i].y2, motif.select_info.portrait_scale[1], motif.select_info.portrait_scale[2])
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
		select.f_p1TeamMenu()
	--Player1 select
	elseif main.p1In > 0 or main.p1Char ~= nil then
		select.f_p1SelectMenu()
	end
	--Player2 team menu
	if not p2TeamEnd then
		select.f_p2TeamMenu()
	--Player2 select
	elseif main.p2In > 0 or main.p2Char ~= nil then
		select.f_p2SelectMenu()
	end
	if p1Cell then
		--draw p1 name
		if #t_p1Selected < p1NumChars then
			textImgSetText(txt_p1Name, main.f_getName(p1Cell))
			main.f_textImgPosDraw(
				txt_p1Name,
				motif.select_info.p1_name_offset[1] + #t_p1Selected * motif.select_info.p1_name_spacing[1],
				motif.select_info.p1_name_offset[2] + #t_p1Selected * motif.select_info.p1_name_spacing[2],
				motif.select_info.p1_name_font[3]
			)
		end
		select.f_drawName(
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
			textImgSetText(txt_p2Name, main.f_getName(p2Cell))
			main.f_textImgPosDraw(
				txt_p2Name,
				motif.select_info.p2_name_offset[1] + #t_p2Selected * motif.select_info.p2_name_spacing[1],
				motif.select_info.p2_name_offset[2] + #t_p2Selected * motif.select_info.p2_name_spacing[2],
				motif.select_info.p2_name_font[3]
			)
		end
		select.f_drawName(
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
	if p1SelEnd and p2SelEnd and p1TeamEnd and p2TeamEnd then
		if main.stageMenu and not stageEnd then --Stage select
			select.f_stageMenu()
		elseif main.coop and not coopEnd then
			coopEnd = true
			p2TeamEnd = false
		else
			selScreenEnd = true
		end
	end
	--draw layerno = 1 backgrounds
	main.f_drawBG(motif.selectbgdef.bg_data, motif.selectbgdef.bg, 1, motif.selectbgdef.timer)
	--draw fadein
	animDraw(motif.select_info.fadein_data)
	animUpdate(motif.select_info.fadein_data)
	--update timer
	motif.selectbgdef.timer = motif.selectbgdef.timer + 1
	--end loop
	main.f_cmdInput()
	refresh()
end

--;===========================================================
--; PLAYER 1 TEAM MENU
--;===========================================================
local txt_p1TeamSelfTitle = main.f_createTextImg(
	motif.font_data[motif.select_info.p1_teammenu_selftitle_font[1]],
	motif.select_info.p1_teammenu_selftitle_font[2],
	motif.select_info.p1_teammenu_selftitle_font[3],
	motif.select_info.p1_teammenu_selftitle_text,
	motif.select_info.p1_teammenu_pos[1] + motif.select_info.p1_teammenu_selftitle_offset[1],
	motif.select_info.p1_teammenu_pos[2] + motif.select_info.p1_teammenu_selftitle_offset[2],
	motif.select_info.p1_teammenu_selftitle_font_scale[1],
	motif.select_info.p1_teammenu_selftitle_font_scale[2]
)
local txt_p1TeamEnemyTitle = main.f_createTextImg(
	motif.font_data[motif.select_info.p1_teammenu_enemytitle_font[1]],
	motif.select_info.p1_teammenu_enemytitle_font[2],
	motif.select_info.p1_teammenu_enemytitle_font[3],
	motif.select_info.p1_teammenu_enemytitle_text,
	motif.select_info.p1_teammenu_pos[1] + motif.select_info.p1_teammenu_enemytitle_offset[1],
	motif.select_info.p1_teammenu_pos[2] + motif.select_info.p1_teammenu_enemytitle_offset[2],
	motif.select_info.p1_teammenu_enemytitle_font_scale[1],
	motif.select_info.p1_teammenu_enemytitle_font_scale[2]
)
local t_p1TeamMenu = {
	{data = textImgNew(), itemname = 'single', displayname = motif.select_info.teammenu_itemname_single},
	{data = textImgNew(), itemname = 'simul', displayname = motif.select_info.teammenu_itemname_simul},
	{data = textImgNew(), itemname = 'turns', displayname = motif.select_info.teammenu_itemname_turns},
	--{data = textImgNew(), itemname = 'tag', displayname = motif.select_info.teammenu_itemname_tag},
}
t_p1TeamMenu = main.f_cleanTable(t_p1TeamMenu)

local p1TeamActiveCount = 0
local p1TeamActiveFont = 'p1_teammenu_item_active_font'

function select.f_p1TeamMenu()
	if main.p1TeamMenu ~= nil then --Predefined team
		p1NumChars = main.p1TeamMenu.chars
		p1TeamMode = main.p1TeamMenu.mode
		setTeamMode(1, p1TeamMode, p1NumChars)
		p1TeamEnd = true
	else
		--Calculate team cursor position
		if commandGetState(main.p1Cmd, 'u') then
			if p1TeamMenu - 1 >= 1 then
				sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_move_snd[1], motif.select_info.p1_teammenu_move_snd[2])
				p1TeamMenu = p1TeamMenu - 1
			elseif motif.select_info.teammenu_move_wrapping == 1 then
				sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_move_snd[1], motif.select_info.p1_teammenu_move_snd[2])
				p1TeamMenu = #t_p1TeamMenu
			end
		elseif commandGetState(main.p1Cmd, 'd') then
			if p1TeamMenu + 1 <= #t_p1TeamMenu then
				sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_move_snd[1], motif.select_info.p1_teammenu_move_snd[2])
				p1TeamMenu = p1TeamMenu + 1
			elseif motif.select_info.teammenu_move_wrapping == 1 then
				sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_move_snd[1], motif.select_info.p1_teammenu_move_snd[2])
				p1TeamMenu = 1
			end
		elseif t_p1TeamMenu[p1TeamMenu].itemname == 'simul' then
			if commandGetState(main.p1Cmd, 'l') then
				if p1NumSimul - 1 >= 2 then
					sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
					p1NumSimul = p1NumSimul - 1
				end
			elseif commandGetState(main.p1Cmd, 'r') then
				if p1NumSimul + 1 <= config.NumSimul then
					sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
					p1NumSimul = p1NumSimul + 1
				end
			end
		elseif t_p1TeamMenu[p1TeamMenu].itemname == 'turns' then
			if commandGetState(main.p1Cmd, 'l') then
				if p1NumTurns - 1 >= 2 then
					sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
					p1NumTurns = p1NumTurns - 1
				end
			elseif commandGetState(main.p1Cmd, 'r') then
				if p1NumTurns + 1 <= config.NumTurns then
					sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
					p1NumTurns = p1NumTurns + 1
				end
			end
		elseif t_p1TeamMenu[p1TeamMenu].itemname == 'tag' then
			if commandGetState(main.p1Cmd, 'l') then
				if p1NumTag - 1 >= 2 then
					sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
					p1NumTag = p1NumTag - 1
				end
			elseif commandGetState(main.p1Cmd, 'r') then
				if p1NumTag + 1 <= config.NumTag then
					sndPlay(motif.files.snd_data, motif.select_info.p1_teammenu_value_snd[1], motif.select_info.p1_teammenu_value_snd[2])
					p1NumTag = p1NumTag + 1
				end
			end
		end
		--Draw team background
		animUpdate(motif.select_info.p1_teammenu_bg_data)
		animDraw(motif.select_info.p1_teammenu_bg_data)
		--Draw team cursor
		main.f_animPosDraw(
			motif.select_info.p1_teammenu_item_cursor_data,
			(p1TeamMenu - 1) * motif.select_info.p1_teammenu_item_spacing[1],
			(p1TeamMenu - 1) * motif.select_info.p1_teammenu_item_spacing[2]
		)
		--Draw team title
		animUpdate(motif.select_info.p1_teammenu_selftitle_data)
		animDraw(motif.select_info.p1_teammenu_selftitle_data)
		textImgDraw(txt_p1TeamSelfTitle)
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
				textImgDraw(main.f_updateTextImg(
					t_p1TeamMenu[i].data,
					motif.font_data[motif.select_info[p1TeamActiveFont][1]],
					motif.select_info[p1TeamActiveFont][2],
					motif.select_info.p1_teammenu_item_font[3], --mugen ignores active font facing
					t_p1TeamMenu[i].displayname,
					motif.select_info.p1_teammenu_pos[1] + motif.select_info.p1_teammenu_item_offset[1] + motif.select_info.p1_teammenu_item_font_offset[1] + (i - 1) * motif.select_info.p1_teammenu_item_spacing[1],
					motif.select_info.p1_teammenu_pos[2] + motif.select_info.p1_teammenu_item_offset[2] + motif.select_info.p1_teammenu_item_font_offset[2] + (i - 1) * motif.select_info.p1_teammenu_item_spacing[2],
					motif.select_info[p1TeamActiveFont .. '_scale'][1],
					motif.select_info[p1TeamActiveFont .. '_scale'][2]
				))
			else
				--Draw team not active font
				textImgDraw(main.f_updateTextImg(
					t_p1TeamMenu[i].data,
					motif.font_data[motif.select_info.p1_teammenu_item_font[1]],
					motif.select_info.p1_teammenu_item_font[2],
					motif.select_info.p1_teammenu_item_font[3],
					t_p1TeamMenu[i].displayname,
					motif.select_info.p1_teammenu_pos[1] + motif.select_info.p1_teammenu_item_offset[1] + motif.select_info.p1_teammenu_item_font_offset[1] + (i - 1) * motif.select_info.p1_teammenu_item_spacing[1],
					motif.select_info.p1_teammenu_pos[2] + motif.select_info.p1_teammenu_item_offset[2] + motif.select_info.p1_teammenu_item_font_offset[2] + (i - 1) * motif.select_info.p1_teammenu_item_spacing[2],
					motif.select_info.p1_teammenu_item_font_scale[1],
					motif.select_info.p1_teammenu_item_font_scale[2]
				))
			end
			--Draw team icons
			if t_p1TeamMenu[i].itemname == 'simul' then
				for j = 1, config.NumSimul do
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
				for j = 1, config.NumTurns do
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
				for j = 1, config.NumTag do
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
			end
		end
		--Confirmed team selection
		if main.f_btnPalNo(main.p1Cmd) > 0 then
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
			end
			setTeamMode(1, p1TeamMode, p1NumChars)
			p1TeamEnd = true
			main.f_cmdInput()
		end
	end
end

--;===========================================================
--; PLAYER 2 TEAM MENU
--;===========================================================
local txt_p2TeamSelfTitle = main.f_createTextImg(
	motif.font_data[motif.select_info.p2_teammenu_selftitle_font[1]],
	motif.select_info.p2_teammenu_selftitle_font[2],
	motif.select_info.p2_teammenu_selftitle_font[3],
	motif.select_info.p2_teammenu_selftitle_text,
	motif.select_info.p2_teammenu_pos[1] + motif.select_info.p2_teammenu_selftitle_offset[1],
	motif.select_info.p2_teammenu_pos[2] + motif.select_info.p2_teammenu_selftitle_offset[2],
	motif.select_info.p2_teammenu_selftitle_font_scale[1],
	motif.select_info.p2_teammenu_selftitle_font_scale[2]
)
local txt_p2TeamEnemyTitle = main.f_createTextImg(
	motif.font_data[motif.select_info.p2_teammenu_enemytitle_font[1]],
	motif.select_info.p2_teammenu_enemytitle_font[2],
	motif.select_info.p2_teammenu_enemytitle_font[3],
	motif.select_info.p2_teammenu_enemytitle_text,
	motif.select_info.p2_teammenu_pos[1] + motif.select_info.p2_teammenu_enemytitle_offset[1],
	motif.select_info.p2_teammenu_pos[2] + motif.select_info.p2_teammenu_enemytitle_offset[2],
	motif.select_info.p2_teammenu_enemytitle_font_scale[1],
	motif.select_info.p2_teammenu_enemytitle_font_scale[2]
)
local t_p2TeamMenu = {
	{data = textImgNew(), itemname = 'single', displayname = motif.select_info.teammenu_itemname_single},
	{data = textImgNew(), itemname = 'simul', displayname = motif.select_info.teammenu_itemname_simul},
	{data = textImgNew(), itemname = 'turns', displayname = motif.select_info.teammenu_itemname_turns},
	--{data = textImgNew(), itemname = 'tag', displayname = motif.select_info.teammenu_itemname_tag},
}
t_p2TeamMenu = main.f_cleanTable(t_p2TeamMenu)

local p2TeamActiveCount = 0
local p2TeamActiveFont = 'p2_teammenu_item_active_font'

function select.f_p2TeamMenu()
	if main.p2TeamMenu ~= nil then --Predefined team
		p2NumChars = main.p2TeamMenu.chars
		p2TeamMode = main.p2TeamMenu.mode
		setTeamMode(2, p2TeamMode, p2NumChars)
		p2TeamEnd = true
	else
		--Command swap
		local cmd = main.p2Cmd
		if main.coop then
			cmd = main.p1Cmd
		end
		--Calculate team cursor position
		if commandGetState(cmd, 'u') then
			if p2TeamMenu - 1 >= 1 then
				sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_move_snd[1], motif.select_info.p2_teammenu_move_snd[2])
				p2TeamMenu = p2TeamMenu - 1
			elseif motif.select_info.teammenu_move_wrapping == 1 then
				sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_move_snd[1], motif.select_info.p2_teammenu_move_snd[2])
				p2TeamMenu = #t_p2TeamMenu
			end
		elseif commandGetState(cmd, 'd') then
			if p2TeamMenu + 1 <= #t_p2TeamMenu then
				sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_move_snd[1], motif.select_info.p2_teammenu_move_snd[2])
				p2TeamMenu = p2TeamMenu + 1
			elseif motif.select_info.teammenu_move_wrapping == 1 then
				sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_move_snd[1], motif.select_info.p2_teammenu_move_snd[2])
				p2TeamMenu = 1
			end
		elseif t_p2TeamMenu[p2TeamMenu].itemname == 'simul' then
			if commandGetState(cmd, 'r') then
				if p2NumSimul - 1 >= 2 then
					sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
					p2NumSimul = p2NumSimul - 1
				end
			elseif commandGetState(cmd, 'l') then
				if p2NumSimul + 1 <= config.NumSimul then
					sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
					p2NumSimul = p2NumSimul + 1
				end
			end
		elseif t_p2TeamMenu[p2TeamMenu].itemname == 'turns' then
			if commandGetState(cmd, 'r') then
				if p2NumTurns - 1 >= 2 then
					sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
					p2NumTurns = p2NumTurns - 1
				end
			elseif commandGetState(cmd, 'l') then
				if p2NumTurns + 1 <= config.NumTurns then
					sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
					p2NumTurns = p2NumTurns + 1
				end
			end
		elseif t_p2TeamMenu[p2TeamMenu].itemname == 'tag' then
			if commandGetState(cmd, 'r') then
				if p2NumTag - 1 >= 2 then
					sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
					p2NumTag = p2NumTag - 1
				end
			elseif commandGetState(cmd, 'l') then
				if p2NumTag + 1 <= config.NumTag then
					sndPlay(motif.files.snd_data, motif.select_info.p2_teammenu_value_snd[1], motif.select_info.p2_teammenu_value_snd[2])
					p2NumTag = p2NumTag + 1
				end
			end
		end
		--Draw team background
		animUpdate(motif.select_info.p2_teammenu_bg_data)
		animDraw(motif.select_info.p2_teammenu_bg_data)
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
			textImgDraw(txt_p2TeamEnemyTitle)
		else
			animUpdate(motif.select_info.p2_teammenu_selftitle_data)
			animDraw(motif.select_info.p2_teammenu_selftitle_data)
			textImgDraw(txt_p2TeamSelfTitle)
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
				textImgDraw(main.f_updateTextImg(
					t_p2TeamMenu[i].data,
					motif.font_data[motif.select_info[p2TeamActiveFont][1]],
					motif.select_info[p2TeamActiveFont][2],
					motif.select_info.p2_teammenu_item_font[3], --mugen ignores active font facing
					t_p2TeamMenu[i].displayname,
					motif.select_info.p2_teammenu_pos[1] + motif.select_info.p2_teammenu_item_offset[1] + motif.select_info.p2_teammenu_item_font_offset[1] + (i - 1) * motif.select_info.p2_teammenu_item_spacing[1],
					motif.select_info.p2_teammenu_pos[2] + motif.select_info.p2_teammenu_item_offset[2] + motif.select_info.p2_teammenu_item_font_offset[2] + (i - 1) * motif.select_info.p2_teammenu_item_spacing[2],
					motif.select_info[p2TeamActiveFont .. '_scale'][1],
					motif.select_info[p2TeamActiveFont .. '_scale'][2]
				))
			else
				--Draw team not active font
				textImgDraw(main.f_updateTextImg(
					t_p2TeamMenu[i].data,
					motif.font_data[motif.select_info.p2_teammenu_item_font[1]],
					motif.select_info.p2_teammenu_item_font[2],
					motif.select_info.p2_teammenu_item_font[3],
					t_p2TeamMenu[i].displayname,
					motif.select_info.p2_teammenu_pos[1] + motif.select_info.p2_teammenu_item_offset[1] + motif.select_info.p2_teammenu_item_font_offset[1] + (i - 1) * motif.select_info.p2_teammenu_item_spacing[1],
					motif.select_info.p2_teammenu_pos[2] + motif.select_info.p2_teammenu_item_offset[2] + motif.select_info.p2_teammenu_item_font_offset[2] + (i - 1) * motif.select_info.p2_teammenu_item_spacing[2],
					motif.select_info.p2_teammenu_item_font_scale[1],
					motif.select_info.p2_teammenu_item_font_scale[2]
				))
			end
			--Draw team icons
			if t_p2TeamMenu[i].itemname == 'simul' then
				for j = 1, config.NumSimul do
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
				for j = 1, config.NumTurns do
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
				for j = 1, config.NumTag do
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
			end
		end
		--Confirmed team selection
		if main.f_btnPalNo(cmd) > 0 then
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
			end
			setTeamMode(2, p2TeamMode, p2NumChars)
			p2TeamEnd = true
			main.f_cmdInput()
		end
	end
end

--;===========================================================
--; PLAYER 1 SELECT MENU
--;===========================================================
function select.f_p1SelectMenu()
	--predefined selection
	if main.p1Char ~= nil then
		local t = {}
		for i = 1, #main.p1Char do
			if t[main.p1Char[i]] == nil then
				t[main.p1Char[i]] = ''
			end
			t_p1Selected[i] = {cel = main.p1Char[i], pal = select.f_randomPal(main.p1Char[i])}
		end
		p1SelEnd = true
		return
	--manual selection
	elseif not p1SelEnd then
		--cell movement
		p1SelX, p1SelY, p1FaceOffset, p1RowOffset = select.f_cellMovement(p1SelX, p1SelY, main.p1Cmd, p1FaceOffset, p1RowOffset, motif.select_info.p1_cursor_move_snd)
		p1Cell = p1SelX + motif.select_info.columns * p1SelY
		--draw active cursor
		local cursorX = p1FaceX + p1SelX * (motif.select_info.cell_size[1] + motif.select_info.cell_spacing)
		local cursorY = p1FaceY + (p1SelY - p1RowOffset) * (motif.select_info.cell_size[2] + motif.select_info.cell_spacing)
		if main.t_selChars[p1Cell + 1].hidden ~= 1 then
			main.f_animPosDraw(motif.select_info.p1_cursor_active_data, cursorX, cursorY)
		end
		--cell selected
		if main.f_btnPalNo(main.p1Cmd) > 0 and main.t_selChars[p1Cell + 1].char ~= nil and main.t_selChars[p1Cell + 1].hidden ~= 2 and #main.t_randomChars > 0 then
			sndPlay(motif.files.snd_data, motif.select_info.p1_cursor_done_snd[1], motif.select_info.p1_cursor_done_snd[2])
			local selected = p1Cell
			if main.t_selChars[selected + 1].char == 'randomselect' or main.t_selChars[selected + 1].hidden == 3 then
				selected = main.t_randomChars[math.random(1, #main.t_randomChars)]
			end
			t_p1Selected[#t_p1Selected + 1] = {cel = selected, pal = main.f_btnPalNo(main.p1Cmd), cursor = {cursorX, cursorY, p1RowOffset}}
			if #t_p1Selected == p1NumChars then
				if main.p2In == 1 and matchNo == 0 then
					p2TeamEnd = false
					p2SelEnd = false
					--commandBufReset(main.p2Cmd)
				end
				p1SelEnd = true
			end
			main.f_cmdInput()
		end
	end
end

--;===========================================================
--; PLAYER 2 SELECT MENU
--;===========================================================
function select.f_p2SelectMenu()
	--predefined selection
	if main.p2Char ~= nil then
		local t = {}
		for i = 1, #main.p2Char do
			if t[main.p2Char[i]] == nil then
				t[main.p2Char[i]] = ''
			end
			t_p2Selected[i] = {cel = main.p2Char[i], pal = select.f_randomPal(main.p2Char[i])}
		end
		p2SelEnd = true
		return
	--p2 selection disabled
	elseif not main.p2SelectMenu then
		p2SelEnd = true
		return
	--manual selection
	elseif not p2SelEnd then
		--cell movement
		p2SelX, p2SelY, p2FaceOffset, p2RowOffset = select.f_cellMovement(p2SelX, p2SelY, main.p2Cmd, p2FaceOffset, p2RowOffset, motif.select_info.p2_cursor_move_snd)
		p2Cell = p2SelX + motif.select_info.columns * p2SelY
		--draw active cursor
		local cursorX = p2FaceX + p2SelX * (motif.select_info.cell_size[1] + motif.select_info.cell_spacing)
		local cursorY = p2FaceY + (p2SelY - p2RowOffset) * (motif.select_info.cell_size[2] + motif.select_info.cell_spacing)
		main.f_animPosDraw(motif.select_info.p2_cursor_active_data, cursorX, cursorY)
		--cell selected
		if main.f_btnPalNo(main.p2Cmd) > 0 and main.t_selChars[p2Cell + 1].char ~= nil and main.t_selChars[p2Cell + 1].hidden ~= 2 and #main.t_randomChars > 0 then
			sndPlay(motif.files.snd_data, motif.select_info.p2_cursor_done_snd[1], motif.select_info.p2_cursor_done_snd[2])
			local selected = p2Cell
			if main.t_selChars[selected + 1].char == 'randomselect' or main.t_selChars[selected + 1].hidden == 3 then
				selected = main.t_randomChars[math.random(1, #main.t_randomChars)]
			end
			t_p2Selected[#t_p2Selected + 1] = {cel = selected, pal = main.f_btnPalNo(main.p2Cmd), cursor = {cursorX, cursorY, p2RowOffset}}
			if #t_p2Selected == p2NumChars then
				p2SelEnd = true
			end
			main.f_cmdInput()
		end
	end
end

--;===========================================================
--; STAGE MENU
--;===========================================================
local txt_selStage = textImgNew()

local stageActiveCount = 0
local stageActiveFont = 'stage_active_font'

function select.f_stageMenu()
	if commandGetState(main.p1Cmd, 'l') then
		sndPlay(motif.files.snd_data, motif.select_info.stage_move_snd[1], motif.select_info.stage_move_snd[2])
		stageList = stageList - 1
		if stageList < 0 then stageList = #main.t_includeStage end
	elseif commandGetState(main.p1Cmd, 'r') then
		sndPlay(motif.files.snd_data, motif.select_info.stage_move_snd[1], motif.select_info.stage_move_snd[2])
		stageList = stageList + 1
		if stageList > #main.t_includeStage then stageList = 0 end
	elseif commandGetState(main.p1Cmd, 'u') then
		sndPlay(motif.files.snd_data, motif.select_info.stage_move_snd[1], motif.select_info.stage_move_snd[2])
		for i = 1, 10 do
			stageList = stageList - 1
			if stageList < 0 then stageList = #main.t_includeStage end
		end
	elseif commandGetState(main.p1Cmd, 'd') then
		sndPlay(motif.files.snd_data, motif.select_info.stage_move_snd[1], motif.select_info.stage_move_snd[2])
		for i = 1, 10 do
			stageList = stageList + 1
			if stageList > #main.t_includeStage then stageList = 0 end
		end
	end
	if main.f_btnPalNo(main.p1Cmd) > 0 then
		sndPlay(motif.files.snd_data, motif.select_info.stage_done_snd[1], motif.select_info.stage_done_snd[2])
		if stageList == 0 then
			stageNo = main.t_includeStage[math.random(1, #main.t_includeStage)]
			setStage(stageNo)
			selectStage(stageNo)
		else
			stageNo = main.t_includeStage[stageList]
			setStage(stageNo)
			selectStage(stageNo)
		end
		stageActiveFont = 'stage_done_font'
		stageEnd = true
		main.f_cmdInput()
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
		t_txt = main.f_extractText(motif.select_info.stage_text, '', 'Random')
	else
		t_txt = main.f_extractText(motif.select_info.stage_text, stageList, getStageName(main.t_includeStage[stageList]):match('^["%s]*(.-)["%s]*$'))
	end
	for i = 1, #t_txt do
		textImgDraw(main.f_updateTextImg(
			txt_selStage,
			motif.font_data[motif.select_info[stageActiveFont][1]],
			motif.select_info[stageActiveFont][2],
			motif.select_info[stageActiveFont][3],
			t_txt[i],
			motif.select_info.stage_pos[1] + (i - 1) * motif.select_info.stage_text_spacing[1],
			motif.select_info.stage_pos[2] + (i - 1) * motif.select_info.stage_text_spacing[2],
			motif.select_info[stageActiveFont .. '_scale'][1],
			motif.select_info[stageActiveFont .. '_scale'][2]
		))
	end
end

--;===========================================================
--; VERSUS SCREEN
--;===========================================================
local txt_p1NameVS = main.f_createTextImg(motif.font_data[motif.vs_screen.p1_name_font[1]], motif.vs_screen.p1_name_font[2], motif.vs_screen.p1_name_font[3], '', 0, 0, motif.vs_screen.p1_name_font_scale[1], motif.vs_screen.p1_name_font_scale[2])
local txt_p2NameVS = main.f_createTextImg(motif.font_data[motif.vs_screen.p2_name_font[1]], motif.vs_screen.p2_name_font[2], motif.vs_screen.p2_name_font[3], '', 0, 0, motif.vs_screen.p2_name_font_scale[1], motif.vs_screen.p2_name_font_scale[2])
local txt_matchNo = main.f_createTextImg(motif.font_data[motif.vs_screen.match_font[1]], motif.vs_screen.match_font[2], motif.vs_screen.match_font[3], '', motif.vs_screen.match_offset[1], motif.vs_screen.match_offset[2], motif.vs_screen.match_font_scale[1], motif.vs_screen.match_font_scale[2])

function select.f_selectVersus()
	local text = main.f_extractText(motif.vs_screen.match_text, matchNo)
	textImgSetText(txt_matchNo, text[1])
	local delay = 0
	local minTime = 15 --let's reserve few extra ticks in case selectChar function needs time to load data, also prevents sound from being interrupted
	main.f_resetBG(motif.vs_screen, motif.versusbgdef, motif.music.vs_bgm)
	if not main.versusScreen then
		delay = minTime
		select.f_selectChar(1, t_p1Selected)
		select.f_selectChar(2, t_p2Selected)
		while true do
			if delay > 0 then
				delay = delay - 1
			else
				main.f_cmdInput()
				break
			end
			main.f_cmdInput()
			refresh()
		end
	else
		local p1Confirmed = false
		local p2Confirmed = false
		local p1Row = 1
		local p2Row = 1
		local t_tmp = {}
		local orderTime = motif.vs_screen.time
		if main.p1In == 1 and main.p2In == 2 and (#t_p1Selected > 1 or #t_p2Selected > 1) and not main.coop then
			orderTime = orderTime + (math.max(#t_p1Selected, #t_p2Selected) - 1) * motif.vs_screen.time_order
			if #t_p1Selected == 1 then
				select.f_selectChar(1, t_p1Selected)
				p1Confirmed = true
			end
			if #t_p2Selected == 1 then
				select.f_selectChar(2, t_p2Selected)
				p2Confirmed = true
			end
		elseif #t_p1Selected > 1 and not main.coop then
			orderTime = orderTime + (#t_p1Selected - 1) * motif.vs_screen.time_order
		else
			select.f_selectChar(1, t_p1Selected)
			p1Confirmed = true
			select.f_selectChar(2, t_p2Selected)
			p2Confirmed = true
			delay = motif.vs_screen.time
			orderTime = -1
		end
		main.f_cmdInput()
		while true do
			if esc() then
				sndPlay(motif.files.snd_data, motif.select_info.cancel_snd[1], motif.select_info.cancel_snd[2])
				main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm)
				break
			elseif p1Confirmed and p2Confirmed then
				if orderTime == -1 and main.f_btnPalNo(main.p1Cmd) > 0 and delay > motif.versusbgdef.timer + minTime then
					delay = motif.versusbgdef.timer + minTime
				elseif delay < motif.versusbgdef.timer then
					break
				end
			elseif orderTime <= motif.versusbgdef.timer then
				if not p1Confirmed then
					select.f_selectChar(1, t_p1Selected)
					p1Confirmed = true
					delay = motif.versusbgdef.timer + minTime
				end
				if not p2Confirmed then
					select.f_selectChar(2, t_p2Selected)
					p2Confirmed = true
					delay = motif.versusbgdef.timer + minTime
				end
			else
				local sndRef = ''
				--if Player1 has not confirmed the order yet
				if not p1Confirmed then
					if main.f_btnPalNo(main.p1Cmd) > 0 then
						if not p1Confirmed then
							sndRef = 'p1_cursor_done_snd'
							select.f_selectChar(1, t_p1Selected)
							p1Confirmed = true
						end
						if main.p2In ~= 2 then
							if not p2Confirmed then
								select.f_selectChar(2, t_p2Selected)
								p2Confirmed = true
							end
						end
					elseif commandGetState(main.p1Cmd, 'u') then
						if #t_p1Selected > 1 then
							sndRef = 'p1_cursor_move_snd'
							p1Row = p1Row - 1
							if p1Row == 0 then p1Row = #t_p1Selected end
						end
					elseif commandGetState(main.p1Cmd, 'd') then
						if #t_p1Selected > 1 then
							sndRef = 'p1_cursor_move_snd'
							p1Row = p1Row + 1
							if p1Row > #t_p1Selected then p1Row = 1 end
						end
					elseif commandGetState(main.p1Cmd, 'l') then
						if p1Row - 1 > 0 then
							sndRef = 'p1_cursor_move_snd'
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
					elseif commandGetState(main.p1Cmd, 'r') then
						if p1Row + 1 <= #t_p1Selected then
							sndRef = 'p1_cursor_move_snd'
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
					if main.f_btnPalNo(main.p2Cmd) > 0 then
						if not p2Confirmed then
							sndRef = 'p2_cursor_done_snd'
							select.f_selectChar(2, t_p2Selected)
							p2Confirmed = true
						end
					elseif commandGetState(main.p2Cmd, 'u') then
						if #t_p2Selected > 1 then
							sndRef = 'p2_cursor_move_snd'
							p2Row = p2Row - 1
							if p2Row == 0 then p2Row = #t_p2Selected end
						end
					elseif commandGetState(main.p2Cmd, 'd') then
						if #t_p2Selected > 1 then
							sndRef = 'p2_cursor_move_snd'
							p2Row = p2Row + 1
							if p2Row > #t_p2Selected then p2Row = 1 end
						end
					elseif commandGetState(main.p2Cmd, 'l') then
						if p2Row + 1 <= #t_p2Selected then
							sndRef = 'p2_cursor_move_snd'
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
					elseif commandGetState(main.p2Cmd, 'r') then
						if p2Row - 1 > 0 then
							sndRef = 'p2_cursor_move_snd'
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
				--sndPlay separated to not play more than 1 sound at once
				if sndRef ~= '' then
					sndPlay(motif.files.snd_data, motif.vs_screen[sndRef][1], motif.vs_screen[sndRef][2])
					delay = motif.versusbgdef.timer + minTime
				end
			end
			--draw clearcolor
			animDraw(motif.versusbgdef.bgclearcolor_data)
			--draw clearcolor
			animDraw(motif.versusbgdef.bgclearcolor_data)
			--draw layerno = 0 backgrounds
			main.f_drawBG(motif.versusbgdef.bg_data, motif.versusbgdef.bg, 0, motif.versusbgdef.timer)
			--draw portraits
			select.f_drawPortrait(
				t_p1Selected,
				motif.vs_screen.p1_pos[1] + motif.vs_screen.p1_offset[1],
				motif.vs_screen.p1_pos[2] + motif.vs_screen.p1_offset[2],
				motif.vs_screen.p1_facing,
				motif.vs_screen.p1_scale[1],
				motif.vs_screen.p1_scale[2],
				motif.vs_screen.p1_spacing[1],
				motif.vs_screen.p1_spacing[2],
				motif.vs_screen.p1_num,
				'versus'
			)
			select.f_drawPortrait(
				t_p2Selected,
				motif.vs_screen.p2_pos[1] + motif.vs_screen.p2_offset[1],
				motif.vs_screen.p2_pos[2] + motif.vs_screen.p2_offset[2],
				motif.vs_screen.p2_facing,
				motif.vs_screen.p2_scale[1],
				motif.vs_screen.p2_scale[2],
				motif.vs_screen.p2_spacing[1],
				motif.vs_screen.p2_spacing[2],
				motif.vs_screen.p2_num,
				'versus'
			)
			--draw names
			select.f_drawName(
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
			select.f_drawName(
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
				textImgDraw(txt_matchNo)
			end
			--draw layerno = 1 backgrounds
			main.f_drawBG(motif.versusbgdef.bg_data, motif.versusbgdef.bg, 1, motif.versusbgdef.timer)
			--draw fadein
			animDraw(motif.vs_screen.fadein_data)
			animUpdate(motif.vs_screen.fadein_data)
			--update timer
			motif.versusbgdef.timer = motif.versusbgdef.timer + 1
			--end loop
			main.f_cmdInput()
			refresh()
		end
	end
end

function select.f_selectChar(player, t)
	for i = 1, #t do
		selectChar(player, t[i].cel, t[i].pal)
	end
end

--;===========================================================
--; RESULT SCREEN
--;===========================================================
local txt_resultSurvival = main.f_createTextImg(
	motif.font_data[motif.survival_results_screen.winstext_font[1]],
	motif.survival_results_screen.winstext_font[2],
	motif.survival_results_screen.winstext_font[3],
	'',
	motif.survival_results_screen.winstext_offset[1],
	motif.survival_results_screen.winstext_offset[2],
	motif.survival_results_screen.winstext_font_scale[1],
	motif.survival_results_screen.winstext_font_scale[2]
)
local txt_resultVS100 = main.f_createTextImg(
	motif.font_data[motif.vs100kumite_results_screen.winstext_font[1]],
	motif.vs100kumite_results_screen.winstext_font[2],
	motif.vs100kumite_results_screen.winstext_font[3],
	'',
	motif.vs100kumite_results_screen.winstext_offset[1],
	motif.vs100kumite_results_screen.winstext_offset[2],
	motif.vs100kumite_results_screen.winstext_font_scale[1],
	motif.vs100kumite_results_screen.winstext_font_scale[2]
)

function select.f_result(state)
	--if state == true then --win
	--elseif state == false then --loose
	--end
	local t = {}
	local t_resultText = {}
	local txt = ''
	if main.gameMode == 'survival' or main.gameMode == 'survivalcoop' or main.gameMode == 'netplaysurvivalcoop' then
		t = motif.survival_results_screen
		t_resultText = main.f_extractText(t.winstext_text, winCnt)
		txt = txt_resultSurvival
	elseif main.gameMode == '100kumite' then
		t = motif.vs100kumite_results_screen
		t_resultText = main.f_extractText(t.winstext_text, winCnt, looseCnt)
		txt = txt_resultVS100
	else
		return
	end
	main.f_resetBG(t, motif.resultsbgdef, motif.music.results_bgm)
	main.f_cmdInput()
	while true do
		if esc() or main.f_btnPalNo(main.p1Cmd) > 0 then
			main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm)
			break
		elseif motif.resultsbgdef.timer >= t.show_time then
			--add fadeout code here
			main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm)
			break
		end
		--draw clearcolor
		--animDraw(motif.resultsbgdef.bgclearcolor_data) --disabled to not cover game screen
		--draw layerno = 0 backgrounds
		main.f_drawBG(motif.resultsbgdef.bg_data, motif.resultsbgdef.bg, 0, motif.resultsbgdef.timer)
		--draw text
		for i = 1, #t_resultText do
			textImgSetText(txt, t_resultText[i])
			textImgSetPos(
				txt,
				t.winstext_offset[1] - t.winstext_spacing[1] + i * t.winstext_spacing[1],
				t.winstext_spacing[2] - t.winstext_spacing[2] + i * t.winstext_spacing[2]
			)
			textImgDraw(txt)
		end
		--draw layerno = 1 backgrounds
		main.f_drawBG(motif.resultsbgdef.bg_data, motif.resultsbgdef.bg, 1, motif.resultsbgdef.timer)
		--draw fadein
		animDraw(t.fadein_data)
		animUpdate(t.fadein_data)
		--update timer
		motif.resultsbgdef.timer = motif.resultsbgdef.timer + 1
		--end loop
		main.f_cmdInput()
		refresh()
	end
end

--;===========================================================
--; VICTORY SCREEN
--;===========================================================
local txt_winquote = main.f_createTextImg(motif.font_data[motif.victory_screen.winquote_font[1]], motif.victory_screen.winquote_font[2], motif.victory_screen.winquote_font[3], '', 0, 0, motif.victory_screen.winquote_font_scale[1], motif.victory_screen.winquote_font_scale[2])
local txt_p1_winquoteName = main.f_createTextImg(
	motif.font_data[motif.victory_screen.p1_name_font[1]],
	motif.victory_screen.p1_name_font[2],
	motif.victory_screen.p1_name_font[3],
	'',
	motif.victory_screen.p1_name_offset[1],
	motif.victory_screen.p1_name_offset[2],
	motif.victory_screen.p1_name_font_scale[1],
	motif.victory_screen.p1_name_font_scale[2]
)
local txt_p2_winquoteName = main.f_createTextImg(
	motif.font_data[motif.victory_screen.p2_name_font[1]],
	motif.victory_screen.p2_name_font[2],
	motif.victory_screen.p2_name_font[3],
	'',
	motif.victory_screen.p2_name_offset[1],
	motif.victory_screen.p2_name_offset[2],
	motif.victory_screen.p2_name_font_scale[1],
	motif.victory_screen.p2_name_font_scale[2]
)

function select.f_selectVictory()
	if motif.music.victory_bgm == '' then
		main.f_resetBG(motif.victory_screen, motif.victorybgdef)
	else
		main.f_resetBG(motif.victory_screen, motif.victorybgdef, motif.music.victory_bgm)
	end
	textImgSetText(txt_p1_winquoteName, main.f_getName(t_p1Selected[1].cel))
	textImgSetText(txt_p2_winquoteName, main.f_getName(t_p2Selected[1].cel))
	local winquote = ''
	local winnerNum = 0
	local txt_winquoteName = ''
	if winner == 1 then
		winquote = select.f_winquote()
		txt_winquoteName = txt_p1_winquoteName
		winnerNum = t_p1Selected[1].cel
	else--if winner == 2 then
		winquote = select.f_winquote()
		txt_winquoteName = txt_p2_winquoteName
		winnerNum = t_p2Selected[1].cel
	end
	local i = 0
	main.f_cmdInput()
	while true do
		if esc() or main.f_btnPalNo(main.p1Cmd) > 0 then
			main.f_cmdInput()
			break
		elseif motif.victorybgdef.timer >= motif.victory_screen.time then
			--add fadeout code here
			main.f_cmdInput()
			break
		end
		--draw clearcolor
		animDraw(motif.victorybgdef.bgclearcolor_data)
		--draw layerno = 0 backgrounds
		main.f_drawBG(motif.victorybgdef.bg_data, motif.victorybgdef.bg, 0, motif.victorybgdef.timer)
		--draw portraits
		if motif.victory_screen.p2_display == 0 then
			drawVictoryPortrait(
				winnerNum,
				motif.victory_screen.p1_offset[1],
				motif.victory_screen.p1_offset[2],
				motif.victory_screen.p1_facing,
				motif.victory_screen.p1_scale[1],
				motif.victory_screen.p1_scale[2]
			)
		else
			drawVictoryPortrait(
				t_p1Selected[1].cel,
				motif.victory_screen.p1_offset[1],
				motif.victory_screen.p1_offset[2],
				motif.victory_screen.p1_facing,
				motif.victory_screen.p1_scale[1],
				motif.victory_screen.p1_scale[2]
			)
			drawVictoryPortrait(
				t_p2Selected[1].cel,
				motif.victory_screen.p2_offset[1],
				motif.victory_screen.p2_offset[2],
				motif.victory_screen.p2_facing,
				motif.victory_screen.p2_scale[1],
				motif.victory_screen.p2_scale[2]
			)
		end
		--draw winner's name
		textImgDraw(txt_winquoteName)
		--draw winquote
		i = i + 1
		main.f_textRender(
			txt_winquote,
			winquote,
			i,
			motif.victory_screen.winquote_offset[1],
			motif.victory_screen.winquote_offset[2],
			motif.victory_screen.winquote_spacing[2],
			motif.victory_screen.winquote_delay,
			motif.victory_screen.winquote_length
		)
		--draw layerno = 1 backgrounds
		main.f_drawBG(motif.victorybgdef.bg_data, motif.victorybgdef.bg, 1, motif.victorybgdef.timer)
		--draw fadein
		animDraw(motif.victory_screen.fadein_data)
		animUpdate(motif.victory_screen.fadein_data)
		--update timer
		motif.victorybgdef.timer = motif.victorybgdef.timer + 1
		--end loop
		main.f_cmdInput()
		refresh()
	end
end

function select.f_winquote()
	--in future code that reads data from characters will be added here
	return motif.victory_screen.winquote_text
end

--;===========================================================
--; CONTINUE SCREEN
--;===========================================================
local txt_credits = main.f_createTextImg(
	motif.font_data[motif.continue_screen.credits_font[1]],
	motif.continue_screen.credits_font[2],
	motif.continue_screen.credits_font[3],
	'',
	motif.continue_screen.credits_offset[1],
	motif.continue_screen.credits_offset[2],
	motif.continue_screen.credits_font_scale[1],
	motif.continue_screen.credits_font_scale[2]
)

function select.f_continue()
	main.f_resetBG(motif.continue_screen, motif.continuebgdef, motif.music.continue_bgm)
	animReset(motif.continue_screen.continue_anim_data)
	animUpdate(motif.continue_screen.continue_anim_data)
	continue = false
	local text = main.f_extractText(motif.continue_screen.credits_text, main.credits)
	textImgSetText(txt_credits, text[1])
	main.f_cmdInput()
	while true do
		--draw clearcolor (disabled to not cover area)
		--animDraw(motif.continuebgdef.bgclearcolor_data)
		--draw layerno = 0 backgrounds
		main.f_drawBG(motif.continuebgdef.bg_data, motif.continuebgdef.bg, 0, motif.continuebgdef.timer)
		--continue screen state
		if esc() or motif.continuebgdef.timer > motif.continue_screen.endtime then
			main.f_cmdInput()
			break
		elseif motif.continuebgdef.timer < motif.continue_screen.continue_end_skiptime then
			if commandGetState(main.p1Cmd, 'holds') then
				continue = true
				main.credits = main.credits - 1
				text = main.f_extractText(motif.continue_screen.credits_text, main.credits)
				textImgSetText(txt_credits, text[1])
				main.f_cmdInput()
				main.f_resetBG(motif.select_info, motif.selectbgdef, motif.music.select_bgm)
				break
			elseif main.f_btnPalNo(main.p1Cmd) > 0 and motif.continuebgdef.timer >= motif.continue_screen.continue_starttime + motif.continue_screen.continue_skipstart then
				local cnt = 0
				if motif.continuebgdef.timer < motif.continue_screen.continue_9_skiptime then
					cnt = motif.continue_screen.continue_9_skiptime
				elseif motif.continuebgdef.timer <= motif.continue_screen.continue_8_skiptime then
					cnt = motif.continue_screen.continue_8_skiptime
				elseif motif.continuebgdef.timer < motif.continue_screen.continue_7_skiptime then
					cnt = motif.continue_screen.continue_7_skiptime
				elseif motif.continuebgdef.timer < motif.continue_screen.continue_6_skiptime then
					cnt = motif.continue_screen.continue_6_skiptime
				elseif motif.continuebgdef.timer < motif.continue_screen.continue_5_skiptime then
					cnt = motif.continue_screen.continue_5_skiptime
				elseif motif.continuebgdef.timer < motif.continue_screen.continue_4_skiptime then
					cnt = motif.continue_screen.continue_4_skiptime
				elseif motif.continuebgdef.timer < motif.continue_screen.continue_3_skiptime then
					cnt = motif.continue_screen.continue_3_skiptime
				elseif motif.continuebgdef.timer < motif.continue_screen.continue_2_skiptime then
					cnt = motif.continue_screen.continue_2_skiptime
				elseif motif.continuebgdef.timer < motif.continue_screen.continue_1_skiptime then
					cnt = motif.continue_screen.continue_1_skiptime
				elseif motif.continuebgdef.timer < motif.continue_screen.continue_0_skiptime then
					cnt = motif.continue_screen.continue_0_skiptime
				end
				while motif.continuebgdef.timer < cnt do
					motif.continuebgdef.timer = motif.continuebgdef.timer + 1
					animUpdate(motif.continue_screen.continue_anim_data)
				end
			end
			if motif.continuebgdef.timer == motif.continue_screen.continue_9_skiptime then
				sndPlay(motif.files.continue_snd_data, motif.continue_screen.continue_9_snd[1], motif.continue_screen.continue_9_snd[2])
			elseif motif.continuebgdef.timer == motif.continue_screen.continue_8_skiptime then
				sndPlay(motif.files.continue_snd_data, motif.continue_screen.continue_8_snd[1], motif.continue_screen.continue_8_snd[2])
			elseif motif.continuebgdef.timer == motif.continue_screen.continue_7_skiptime then
				sndPlay(motif.files.continue_snd_data, motif.continue_screen.continue_7_snd[1], motif.continue_screen.continue_7_snd[2])
			elseif motif.continuebgdef.timer == motif.continue_screen.continue_6_skiptime then
				sndPlay(motif.files.continue_snd_data, motif.continue_screen.continue_6_snd[1], motif.continue_screen.continue_6_snd[2])
			elseif motif.continuebgdef.timer == motif.continue_screen.continue_5_skiptime then
				sndPlay(motif.files.continue_snd_data, motif.continue_screen.continue_5_snd[1], motif.continue_screen.continue_5_snd[2])
			elseif motif.continuebgdef.timer == motif.continue_screen.continue_4_skiptime then
				sndPlay(motif.files.continue_snd_data, motif.continue_screen.continue_4_snd[1], motif.continue_screen.continue_4_snd[2])
			elseif motif.continuebgdef.timer == motif.continue_screen.continue_3_skiptime then
				sndPlay(motif.files.continue_snd_data, motif.continue_screen.continue_3_snd[1], motif.continue_screen.continue_3_snd[2])
			elseif motif.continuebgdef.timer == motif.continue_screen.continue_2_skiptime then
				sndPlay(motif.files.continue_snd_data, motif.continue_screen.continue_2_snd[1], motif.continue_screen.continue_2_snd[2])
			elseif motif.continuebgdef.timer == motif.continue_screen.continue_1_skiptime then
				sndPlay(motif.files.continue_snd_data, motif.continue_screen.continue_1_snd[1], motif.continue_screen.continue_1_snd[2])
			elseif motif.continuebgdef.timer == motif.continue_screen.continue_0_skiptime then
				sndPlay(motif.files.continue_snd_data, motif.continue_screen.continue_0_snd[1], motif.continue_screen.continue_0_snd[2])
			end
		elseif motif.continuebgdef.timer == motif.continue_screen.continue_end_skiptime then
			playBGM(motif.music.continue_end_bgm)
			sndPlay(motif.files.continue_snd_data, motif.continue_screen.continue_end_snd[1], motif.continue_screen.continue_end_snd[2])
		end
		--draw credits text
		if motif.continuebgdef.timer >= motif.continue_screen.continue_skipstart then --show when counter starts counting down
			textImgDraw(txt_credits)
		end
		--draw counter
		animUpdate(motif.continue_screen.continue_anim_data)
		animDraw(motif.continue_screen.continue_anim_data)
		--draw layerno = 1 backgrounds
		main.f_drawBG(motif.continuebgdef.bg_data, motif.continuebgdef.bg, 1, motif.continuebgdef.timer)
		--draw fadein
		animDraw(motif.continue_screen.fadein_data)
		animUpdate(motif.continue_screen.fadein_data)
		--update timer
		motif.continuebgdef.timer = motif.continuebgdef.timer + 1
		--end loop
		main.f_cmdInput()
		refresh()
	end
end

return select

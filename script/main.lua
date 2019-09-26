-- Disable GC during the initial load so it does not crash.
SetGCPercent(-1)
-------------------------------------------------------------
main = {}

refresh()
math.randomseed(os.time())

--;===========================================================
--; COMMON SECTION
--;===========================================================

--One-time load of the json routines
json = (loadfile 'script/dkjson.lua')()

-- Data loading from config.json
local file = io.open("save/config.json","r")
config = json.decode(file:read("*all"))
file:close()

-- Int localcoord
require "script/screenpack"
main.IntLocalcoordValues()
require "script/localcoord"
main.CalculateLocalcoordValues()
main.IntLifebarScale()
main.SetScaleValues()

main.p1In = 1
main.p2In = 2
--main.inputDialog = inputDialogNew()

function main.f_setCommand(c)
	commandAdd(c, 'u', '$U')
	commandAdd(c, 'd', '$D')
	commandAdd(c, 'l', '$B')
	commandAdd(c, 'r', '$F')
	commandAdd(c, 'a', 'a')
	commandAdd(c, 'b', 'b')
	commandAdd(c, 'c', 'c')
	commandAdd(c, 'x', 'x')
	commandAdd(c, 'y', 'y')
	commandAdd(c, 'z', 'z')
	commandAdd(c, 's', 's')
	commandAdd(c, 'v', 'v')
	commandAdd(c, 'w', 'w')
	commandAdd(c, 'holds', '/s')
	commandAdd(c, 'su', '/s, U')
	commandAdd(c, 'sd', '/s, D')
end

main.p1Cmd = commandNew()
main.f_setCommand(main.p1Cmd)

main.p2Cmd = commandNew()
main.f_setCommand(main.p2Cmd)

function main.f_cmdInput()
	commandInput(main.p1Cmd, main.p1In)
	commandInput(main.p2Cmd, main.p2In)
end

--returns value depending on button pressed (a = 1; a + start = 7 etc.)
function main.f_btnPalNo(cmd)
	local s = 0
	if commandGetState(cmd, 'holds') then s = 6 end
	if commandGetState(cmd, 'a') then return 1 + s end
	if commandGetState(cmd, 'b') then return 2 + s end
	if commandGetState(cmd, 'c') then return 3 + s end
	if commandGetState(cmd, 'x') then return 4 + s end
	if commandGetState(cmd, 'y') then return 5 + s end
	if commandGetState(cmd, 'z') then return 6 + s end
	if commandGetState(cmd, 'v') then return 1 + s end
	if commandGetState(cmd, 'w') then return 2 + s end
	return 0
end

--animDraw at specified coordinates
function main.f_animPosDraw(a, x, y)
	animSetPos(a, x, y)
	animUpdate(a)
	animDraw(a)
end

-- Check if files exists.
function main.file_exists(name)
   local f=io.open(name,"r")
   if f~=nil then io.close(f) return true else return false end
end

--textImgDraw at specified coordinates
function main.f_textImgPosDraw(ti, x, y, align)
	align = align or 0
	textImgSetAlign(ti, align)
	if align == -1 then x = x + 1 end --fix for wrong offset after flipping text
	textImgSetPos(ti, x, y)
	textImgDraw(ti)
end

--shortcut for creating new text with several parameters
function main.f_createTextImg(font, bank, align, text, x, y, scaleX, scaleY, colorR, colorG, colorB)
	local ti = textImgNew()
	if font ~= nil then
		textImgSetFont(ti, font)
		textImgSetBank(ti, bank)
		textImgSetAlign(ti, align)
		textImgSetText(ti, text)
		if colorR ~= nil and colorG ~= nil and colorB ~= nil then
			textImgSetColor(ti, colorR, colorG, colorB)
		end
		if align == -1 then x = x + 1 end --fix for wrong offset after flipping text
		textImgSetPos(ti, x, y)
		textImgSetScale(ti, scaleX, scaleY)
	end
	return ti
end

--shortcut for updating text with several parameters
function main.f_updateTextImg(animName, font, bank, align, text, x, y, scaleX, scaleY, colorR, colorG, colorB)
	if font ~= nil then
		textImgSetFont(animName, font)
		textImgSetBank(animName, bank)
		textImgSetAlign(animName, align)
		textImgSetText(animName, text)
		if colorR ~= nil and colorG ~= nil and colorB ~= nil then
			textImgSetColor(animName, colorR, colorG, colorB)
		end
		if align == -1 then x = x + 1 end --fix for wrong offset after flipping text
		textImgSetPos(animName, x, y)
		textImgSetScale(animName, scaleX, scaleY)
	end
	return animName
end

--dynamically adjusts alpha blending each time called based on specified values
local alpha1cur = 0
local alpha2cur = 0
local alpha1add = true
local alpha2add = true
function main.f_dynamicAlpha(animName, r1min, r1max, r1step, r2min, r2max, r2step)
	if r1step == 0 then alpha1cur = r1max end
	if alpha1cur < r1max and alpha1add then
		alpha1cur = alpha1cur + r1step
		if alpha1cur >= r1max then
			alpha1add = false
		end
	elseif alpha1cur > r1min and not alpha1add then
		alpha1cur = alpha1cur - r1step
		if alpha1cur <= r1min then
			alpha1add = true
		end
	end
	if r2step == 0 then alpha2cur = r2max end
	if alpha2cur < r2max and alpha2add then
		alpha2cur = alpha2cur + r2step
		if alpha2cur >= r2max then
			alpha2add = false
		end
	elseif alpha2cur > r2min and not alpha2add then
		alpha2cur = alpha2cur - r2step
		if alpha2cur <= r2min then
			alpha2add = true
		end
	end
	animSetAlpha(animName, alpha1cur, alpha2cur)
end

--generate fading animation
--currently only black when average of RGB <= 128 and white if >128 colors supported
function main.f_fadeAnim(fadeType, ticks, r, g, b)
	local anim = ''
	if (r + g + b)/3 > 128 then --white
		if fadeType == 0 then --fadeout
			for i = 1, ticks do
				anim = anim .. '0,1, 0,0, 1, 0, AS' .. math.floor(256 / ticks * i) .. 'D256\n'
			end
			anim = anim .. '0,1, 0,0, -1, 0, AS256D256'
		elseif fadeType == 1 then --fadein
			for i = ticks, 1, -1 do
				anim = anim .. '0,1, 0,0, 1, 0, AS' .. math.floor(256 / ticks * i) .. 'D256\n'
			end
			anim = anim .. '0,1, 0,0, -1, 0, AS0D256'
		end
	else --black
		if fadeType == 0 then --fadeout
			for i = ticks, 1, -1 do
				anim = anim .. '0,0, 0,0, 1, 0, AS256D' .. math.floor(256 / ticks * i) .. '\n'
			end
			anim = anim .. '0,0, 0,0, -1, 0, AS256D0'
		elseif fadeType == 1 then --fadein
			for i = 1, ticks do
				anim = anim .. '0,0, 0,0, 1, 0, AS256D' .. math.floor(256 / ticks * i) .. '\n'
			end
			anim = anim .. '0,0, 0,0, -1, 0, AS256D256'
		end
	end
	anim = animNew(main.fadeSff, anim)
	animUpdate(anim)
	return anim, ticks
end

--generate clearcolor animation
--currently only black when average of RGB <= 128 and white if >128 colors supported
function main.f_clearColor(r, g, b)
	local anim = ''
	if (r + g + b)/3 > 128 then
		anim = '0,1, 0,0, -1' --white
	else
		anim = '0,0, 0,0, -1' --black
	end
	anim = animNew(main.fadeSff, anim)
	animUpdate(anim)
	return anim
end

--generate anim from table
function main.f_animFromTable(t, sff, x, y, scaleX, scaleY, facing, infFrame)
	x = x or 0
	y = y or 0
	scaleX = scaleX or 1.0
	scaleY = scaleY or 1.0
	facing = facing or '0'
	infFrame = infFrame or 1
	local facing_sav = ''
	local anim = ''
	local length = 0
	for i = 1, #t do
		local t_anim = {}
		for j, c in ipairs(main.f_strsplit(',', t[i])) do --split using "," delimiter
			t_anim[#t_anim + 1] = c
		end
		if #t_anim > 1 then
			--required parameters
			t_anim[3] = tonumber(t_anim[3]) + x
			t_anim[4] = tonumber(t_anim[4]) + y
			if tonumber(t_anim[5]) == -1 then
				length = length + infFrame
			else
				length = length + tonumber(t_anim[5])
			end
			--optional parameters
			if t_anim[6] ~= nil and not t_anim[6]:match(facing) then --flip parameter not negated by repeated flipping
				if t_anim[6]:match('[Hh]') then t_anim[3] = t_anim[3] + 1 end --fix for wrong offset after flipping sprites
				if t_anim[6]:match('[Vv]') then t_anim[4] = t_anim[4] + 1 end --fix for wrong offset after flipping sprites
				t_anim[6] = facing .. t_anim[6]
			end
		end
		for j = 1, #t_anim do
			if j == 1 then
				anim = anim .. t_anim[j]
			else
				anim = anim .. ', ' .. t_anim[j]
			end
		end
		anim = anim .. '\n'
	end
	local data = animNew(sff, anim)
	animSetScale(data, scaleX, scaleY)
	animUpdate(data)
	return data, length
end

--Convert number to name and get rid of the ""
function main.f_getName(cell)
	local tmp = getCharName(cell)
	tmp = tmp:gsub('^["%s]*(.-)["%s]*$', '%1') --needed for s-size ikemen
	if main.t_selChars[cell + 1].hidden == 3 then
		tmp = 'Random'
	elseif main.t_selChars[cell + 1].hidden == 2 then
		tmp = ''
	end
	return tmp
end

--randomizes table content
function main.f_shuffleTable(t)
	local rand = math.random
	assert(t, "main.f_shuffleTable() expected a table, got nil")
	local iterations = #t
	local j
	for i = iterations, 2, -1 do
		j = rand(i)
		t[i], t[j] = t[j], t[i]
	end
end

--iterate over the table in order
-- basic usage, just sort by the keys:
--for k, v in main.f_sortKeys(t) do
--    print(k,v)
--end
-- this uses an custom sorting function ordering by score descending
--for k, v in  main.f_sortKeys(t, function(t,a,b) return t[b] < t[a] end) do
--    print(k,v)
--end
function main.f_sortKeys(t, order)
	-- collect the keys
	local keys = {}
	for k in pairs(t) do keys[#keys + 1] = k end
	-- if order function given, sort it by passing the table and keys a, b,
	-- otherwise just sort the keys 
	if order then
		table.sort(keys, function(a,b) return order(t, a, b) end)
	else
		table.sort(keys)
	end
	-- return the iterator function
	local i = 0
	return function()
		i = i + 1
		if keys[i] then
			return keys[i], t[keys[i]]
		end
	end
end

--prints "t" table content into "toFile" file
function main.f_printTable(t, toFile)
	local toFile = toFile or 'debug/table_print.txt'
	local txt = ''
	local print_t_cache = {}
	local function sub_print_t(t, indent)
		if print_t_cache[tostring(t)] then
			txt = txt .. indent .. '*' .. tostring(t) .. '\n'
		else
			print_t_cache[tostring(t)] = true
			if type(t) == 'table' then
				for pos, val in pairs(t) do
					if type(val) == 'table' then
						txt = txt .. indent .. '[' .. pos .. '] => ' .. tostring(t) .. ' {' .. '\n'
						sub_print_t(val, indent .. string.rep(' ', string.len(tostring(pos)) + 8))
						txt = txt .. indent .. string.rep(' ', string.len(tostring(pos)) + 6) .. '}' .. '\n'
					elseif type(val) == 'string' then
						txt = txt .. indent .. '[' .. pos .. '] => "' .. val .. '"' .. '\n'
					else
						txt = txt .. indent .. '[' .. pos .. '] => ' .. tostring(val) ..'\n'
					end
				end
			else
				txt = txt .. indent .. tostring(t) .. '\n'
			end
		end
	end
	if type(t) == 'table' then
		txt = txt .. tostring(t) .. ' {' .. '\n'
		sub_print_t(t, '  ')
		txt = txt .. '}' .. '\n'
	else
		sub_print_t(t, '  ')
	end
	local file = io.open(toFile,"w+")
	if file == nil then return end
	file:write(txt)
	file:close()
end

--prints "v" variable into "toFile" file
function main.f_printVar(v, toFile)
	local toFile = toFile or 'debug/var_print.txt'
	local file = io.open(toFile,"w+")
	file:write(v)
	file:close()
end

--remove duplicated string pattern
function main.f_uniq(str, pattern, subpattern)
	local out = {}
	for s in str:gmatch(pattern) do
		local s2 = s:match(subpattern)
		if not main.f_contains(out, s2) then out[#out + 1] = s end
	end
	return table.concat(out)
end

function main.f_contains(t, val)
	for k, v in pairs(t) do
		--if v == val then
		if v:match(val) then
			return true
		end
	end
	return false
end

--- Draw string letter by letter + wrap lines.
-- @data: text data
-- @str: string (text you want to draw)
-- @counter: external counter (values should be increased each frame by 1 starting from 1)
-- @x: first line X position
-- @y: first line Y position
-- @spacing: spacing between lines (rendering Y position increase for each line)
-- @delay (optional): ticks (frames) delay between each letter is rendered, defaults to 0 (all text rendered immediately)
-- @limit (optional): maximum line length (string wraps when reached), if omitted line wraps only if string contains '\n'
function main.f_textRender(data, str, counter, x, y, spacing, delay, limit)
	local delay = delay or 0
	local limit = limit or -1
	str = tostring(str)
	if limit == -1 then
		str = str:gsub('\\n', '\n')
	else
		str = str:gsub('%s*\\n%s*', ' ')
		if math.floor(#str / limit) + 1 > 1 then
			str = main.f_wrap(str, limit, indent, indent1)
		end
	end
	local subEnd = math.floor(#str - (#str - counter / delay))
	local t = {}
	for line in str:gmatch('([^\r\n]*)[\r\n]?') do
		t[#t + 1] = line
	end
	local lengthCnt = 0
	for i = 1, #t do
		if subEnd < #str then
			local length = #t[i]
			if i > 1 and i <= #t then
				length = length + 1
			end
			lengthCnt = lengthCnt + length
			if subEnd < lengthCnt then
				t[i] = t[i]:sub(0, subEnd - lengthCnt)
			end
		end
		textImgSetText(data, t[i])
		textImgSetPos(data, x, y + spacing * (i - 1))
		textImgDraw(data)
	end
end

--- Wrap a long string.
-- source: http://lua-users.org/wiki/StringRecipes
-- @str: string to wrap
-- @limit: maximum line length
-- @indent: regular indentation
-- @indent1: indentation of first line
function main.f_wrap(str, limit, indent, indent1)
	indent = indent or ''
	indent1 = indent1 or indent
	limit = limit or 72
	local here = 1 - #indent1
	return indent1 .. str:gsub("(%s+)()(%S+)()",
		function(sp, st, word, fi)
			if fi - here > limit then
				here = st - #indent
				return '\n' .. indent .. word
			end
		end
	)
end

--Convert DEF string to table (each line = next item; %i, %s swapped with variable values)
function main.f_extractText(txt, v1, v2, v3, v4)
	local t = {v1 or '', v2 or '', v3 or '', v4 or ''}
	local tmp = ''
	txt = txt:gsub('%%[is]', '%%')
	for i, c in ipairs(main.f_strsplit('%%', txt)) do --split string using "%" delimiter
		if t[i] == '' then
			c = c:gsub('%s$', '')
		end
		tmp = tmp .. c .. t[i]
	end
	t = {}
	for i, c in ipairs(main.f_strsplit('\n', tmp)) do --split string using "\n" delimiter
		t[i] = c
	end
	if #t == 0 then
		t[1] = tmp
	end
	return t
end

--ensure that correct data type is set
function main.f_dataType(arg)
	arg = arg:gsub('^%s*(.-)%s*$', '%1')
	if tonumber(arg) then
		arg = tonumber(arg)
	elseif arg == 'true' then
		arg = true
	elseif arg == 'false' then
		arg = false
	else
		arg = tostring(arg)
	end
	return arg
end

--merge 2 tables into 1 overwriting values
function main.f_tableMerge(t1, t2)
	for k, v in pairs(t2) do
		if type(v) == "table" then
			if type(t1[k] or false) == "table" then
				main.f_tableMerge(t1[k] or {}, t2[k] or {})
			else
				t1[k] = v
			end
		elseif type(t1[k] or false) == "table" then
			t1[k][1] = v
		else
			t1[k] = v
		end
	end
	return t1
end

--check if file exists
function main.f_fileExists(name)
	local f = io.open(name,'r')
	if f ~= nil then
		io.close(f)
		return true
	else
		return false
	end
end

--split strings
function main.f_strsplit(delimiter, text)
	local list = {}
	local pos = 1
	if string.find('', delimiter, 1) then
		if string.len(text) == 0 then
			table.insert(list, text)
		else
			for i = 1, string.len(text) do
				table.insert(list, string.sub(text, i, i))
			end
		end
	else
		while true do
			local first, last = string.find(text, delimiter, pos)
			if first then
				table.insert(list, string.sub(text, pos, first - 1))
				pos = last + 1
			else
				table.insert(list, string.sub(text, pos))
				break
			end
		end
	end
	return list
end

--expand background table with background controllers data
function main.f_ctrlBG(t_bg, t_ctrl)
	local t = {}
	for k, v in pairs(t_ctrl) do
		local t1 = t_ctrl[k].ctrl
		for i = 1, #t1 do
			for k2, v2 in pairs(t1[i].ctrlid) do
				local t2 = t1[i].type
				if t1[i].ctrlid[k2] == t_bg.id and t2 ~= 'null' then
					if t[t2] == nil then
						t[t2] = {}
					end
					t[t2][#t[t2] + 1] = {}
					local t3 = t[t2][#t[t2]]
					t3.time = {t1[i].time[1], t1[i].time[2], t1[i].time[3]}
					t3.timer = {t1[i].time[1], t1[i].time[2], t1[i].time[3]}
					if t2 == 'visible' or t2 == 'enable' or t2 == 'anim' or t2 == 'gamemode' then
						t3.value = t1[i].value
					elseif t2 == 'sinx' or t2 == 'siny' then
						t3.value = {t1[i].value[1], t1[i].value[2], t1[i].value[3]}
					elseif t2 == 'velset' or t2 == 'veladd' or t2 == 'posset' or t2 == 'posadd' then
						if t1[i].x ~= nil then
							t3.x = t1[i].x
						end
						if t1[i].y ~= nil then
							t3.y = t1[i].y
						end
					end
				end
			end
		end
	end
	return t
end

--draw background layers
function main.f_drawBG(data, info, layerno, timer, localcoord)
	timer = timer or 0
	--loop through all backgrounds
	for i = 1, #data do
		--skip if layerno is not matched
		if info[i].layerno == layerno then
			local x = 0
			local y = 0
			--loop through controllers
			for k, v in pairs(info[i].ctrl) do
				for j = 1, #info[i].ctrl[k] do
					local t = info[i].ctrl[k][j]
					--reset internal timer
					if t.time[3] ~= -1 and timer >= t.timer[3] then
						t.timer[1] = timer + t.time[1]
						t.timer[2] = timer + t.time[2]
						t.timer[3] = timer + t.time[3]
					end
					--check if time condition allows controller to run
					if timer >= t.timer[1] and timer <= t.timer[2] then
						x = 0
						y = 0
						if k == 'gamemode' then --Ikemen feature
							for m = 1, #t.enabled do
								info[i].ctrl_flags.enabled = 0
								if t.enabled[m] == main.gameMode then
									info[i].ctrl_flags.enabled = 1
									m = #t.enabled
								end
							end
						elseif k == 'visible' then
							info[i].ctrl_flags.visible = t.value
						elseif k == 'enable' then
							info[i].ctrl_flags.enabled = t.value
						elseif k == 'velset' or k == 'posset' then
							if t.x ~= nil then
								info[i].ctrl_flags.velx = 0 - info[i].velocity[1] + t.x * 320/localcoord[1]
							end
							if t.y ~= nil then
								info[i].ctrl_flags.vely = 0 - info[i].velocity[2] + t.y * 240/localcoord[2]
							end
						elseif k == 'veladd'  then --or k == 'posadd' then
							if t.x ~= nil then
								info[i].ctrl_flags.velx = info[i].ctrl_flags.velx + t.x * 320/localcoord[1]
							end
							if t.y ~= nil then
								info[i].ctrl_flags.vely = info[i].ctrl_flags.vely + t.y * 240/localcoord[2]
							end
						elseif k == 'posadd' then
							if t.x ~= nil then x = t.x * 320/localcoord[1] end
							if t.y ~= nil then y = t.y * 240/localcoord[2] end
							animAddPos(data[i], x, y)
						--[[elseif k == 'posset' then
							if t.x ~= nil then
								x = t.x * 320/localcoord[1]
							else
								x = info[i].start[1]
							end
							if t.y ~= nil then
								y = t.y * 240/localcoord[2]
							else
								y = info[i].start[2]
							end
							animSetPos(data[i], x, y)
							animAddPos(data[i], 160, 0) --for some reason needed in ikemen
						elseif k == 'posadd' then
							if t.x ~= nil then x = t.x * 320/localcoord[1] end
							if t.y ~= nil then y = t.y * 240/localcoord[2] end
							animAddPos(data[i], x, y)]]
						--elseif k == 'anim' then --not supported yet
						--elseif k == 'sinx' then --not supported yet
						--elseif k == 'siny' then --not supported yet
						end
					end
				end
			end
			--animation
			if info[i].ctrl_flags.enabled == 1 then
				x = (info[i].ctrl_flags.velx + info[i].velocity[1]) * info[i].delta[1]
				y = (info[i].ctrl_flags.vely + info[i].velocity[2]) * info[i].delta[2]
				info[i].ctrl_flags.x = info[i].ctrl_flags.x + x
				info[i].ctrl_flags.y = info[i].ctrl_flags.y + y
				animAddPos(data[i], x, y)
				animUpdate(data[i])
				if info[i].ctrl_flags.visible == 1 then
					animDraw(data[i])
				end
			end
		end
	end
end

--reset screenpack data
function main.f_resetBG(info, bgdef, bgm, bgmLoop, bgmVolume, bgmloopstart, bgmloopend)
	bgm = bgm or nil
	animReset(info.fadein_data)
	animUpdate(info.fadein_data)
	animReset(info.fadeout_data)
	animUpdate(info.fadeout_data)
	bgdef.timer = 0
	for i = 1, #bgdef.bg do
		bgdef.bg[i].ctrl_flags.visible = 1
		bgdef.bg[i].ctrl_flags.enabled = 1
		bgdef.bg[i].ctrl_flags.velx = 0
		bgdef.bg[i].ctrl_flags.vely = 0
		animReset(bgdef.bg_data[i])
		animAddPos(bgdef.bg_data[i], 0 - bgdef.bg[i].ctrl_flags.x, 0 - bgdef.bg[i].ctrl_flags.y)
		animUpdate(bgdef.bg_data[i])
		bgdef.bg[i].ctrl_flags.x = 0
		bgdef.bg[i].ctrl_flags.y = 0
		for k, v in pairs(bgdef.bg[i].ctrl) do
			for j = 1, #bgdef.bg[i].ctrl[k] do
				bgdef.bg[i].ctrl[k][j].timer[1] = bgdef.bg[i].ctrl[k][j].time[1]
				bgdef.bg[i].ctrl[k][j].timer[2] = bgdef.bg[i].ctrl[k][j].time[2]
				bgdef.bg[i].ctrl[k][j].timer[3] = bgdef.bg[i].ctrl[k][j].time[3]
			end
		end
	end
	if bgm ~= nil then
		playBGM(bgm, true, bgmLoop, bgmVolume, bgmloopstart or "0", bgmloopend or "0")
	end
end

--return table with reversed keys
function main.f_reversedTable(t)
	local reversedTable = {}
	local itemCount = #t
	for k, v in ipairs(t) do
		reversedTable[itemCount + 1 - k] = v
	end
	return reversedTable
end

--return table without rows disabled in screenpack
function main.f_cleanTable(t)
	local t_clean = {}
	for i = 1, #t do
		if t[i].displayname ~= '' then
			t_clean[#t_clean + 1] = t[i]
		end
	end
	return t_clean
end

--warning display
local txt_warning = textImgNew()
function main.f_warning(t, info, background, font_info, title, box)
	font_info = font_info or motif.warning_info
	title = title or main.txt_warningTitle
	box = box or main.warningBox
	main.f_cmdInput()
	box = box or false
	while true do
		if main.f_btnPalNo(main.p1Cmd) > 0 or esc() then
			sndPlay(motif.files.snd_data, info.cursor_move_snd[1], info.cursor_move_snd[2])
			break
		end
		--draw clearcolor
		animDraw(background.bgclearcolor_data)
		--draw layerno = 0 backgrounds
		main.f_drawBG(background.bg_data, background.bg, 0, background.timer, {320,240})
		--draw layerno = 1 backgrounds
		main.f_drawBG(background.bg_data, background.bg, 1, background.timer, {320,240})
		--draw menu box
		animDraw(box)
		--draw title
		textImgDraw(title)
		--draw text
		for i = 1, #t do
			main.f_updateTextImg(
				txt_warning,
				motif.font_data[font_info.text_font[1]],
				font_info.text_font[2],
				font_info.text_font[3],
				t[i],
				font_info.text_pos[1],
				font_info.text_pos[2] - font_info.text_spacing[2] + i * font_info.text_spacing[2],
				font_info.text_font_scale[1],
				font_info.text_font_scale[2],
				font_info.text_font[4],
				font_info.text_font[5],
				font_info.text_font[6]
			)
			textImgDraw(txt_warning)
		end
		--update timer
		background.timer = background.timer + 1
		--end loop
		main.f_cmdInput()
		refresh()
	end
end

--input display
local txt_input = textImgNew()
function main.f_input(t, info, background, type)
	main.f_cmdInput()
	type = type or 'string'
	if type == 'string' then
		t[#t + 1] = ''
	end
	local input = ''
	resetKey()
	while true do
		if esc() then
			input = ''
			break
		end
		if type == 'key' then
			input = getKey()
			if input ~= '' then
				main.f_cmdInput()
				break
			end
		else --string
			if getKey() == 'RETURN' then
				main.f_cmdInput()
				break
			elseif getKey() == 'BACKSPACE' then
				input = input:match('^(.-).?$')
			else
				input = input .. getKeyText()
			end
			t[#t] = input
			resetKey()
		end
		--draw clearcolor
		animDraw(background.bgclearcolor_data)
		--draw layerno = 0 backgrounds
		main.f_drawBG(background.bg_data, background.bg, 0, background.timer, {320,240})
		--draw layerno = 1 backgrounds
		main.f_drawBG(background.bg_data, background.bg, 1, background.timer, {320,240})
		--draw menu box
		animDraw(main.warningBox)
		--draw text
		for i = 1, #t do
			main.f_updateTextImg(
				txt_input,
				motif.font_data[motif.warning_info.text_font[1]],
				motif.warning_info.text_font[2],
				motif.warning_info.text_font[3],
				t[i],
				motif.warning_info.text_pos[1],
				motif.warning_info.text_pos[2] - motif.warning_info.text_spacing[2] + i * motif.warning_info.text_spacing[2],
				motif.warning_info.text_font_scale[1],
				motif.warning_info.text_font_scale[2],
				motif.warning_info.text_font[4],
				motif.warning_info.text_font[5],
				motif.warning_info.text_font[6]
			)
			textImgDraw(txt_input)
		end
		--update timer
		background.timer = background.timer + 1
		--end loop
		main.f_cmdInput()
		refresh()
	end
	return input
end

main.nextRefresh = os.clock() + 0.02
function main.loadingRefresh(txt)
	if os.clock() >= main.nextRefresh then
		if txt ~= nil then
			textImgDraw(txt)
		end
		refresh()
		main.nextRefresh = os.clock() + 0.02
	end
end

--;===========================================================
--; COMMAND LINE QUICK VS
--;===========================================================
main.flags = getCommandLineFlags()
if main.flags['-p1'] ~= nil and main.flags['-p2'] ~= nil then
	--load lifebar
	local sp = config.Motif
	if main.flags['-r'] ~= nil then
		if main.f_fileExists(main.flags['-r']) then
			sp = main.flags['-r']
		elseif main.f_fileExists('data/' .. main.flags['-r'] .. '/system.def') then
			sp = 'data/' .. main.flags['-r'] .. '/system.def'
		end
	end
	local fileDir = sp:match('^(.-)[^/\\]+$')
	local file = io.open(sp,"r")
	local s = file:read("*all")
	file:close()
	local lifebar = s:match('fight%s*=%s*(.-%.def)%s*')
	if main.f_fileExists(lifebar) then
		loadLifebar(lifebar)
	elseif main.f_fileExists(fileDir .. lifebar) then
		loadLifebar(fileDir .. lifebar)
	elseif main.f_fileExists('data/' .. lifebar) then
		loadLifebar('data/' .. lifebar)
	else
		loadLifebar('data/fight.def')
	end
	refresh()
	--add chars
	local p1NumChars = 0
	local p2NumChars = 0
	local t = {}
	for k, v in pairs(main.flags) do
		if k:match('^-p[1-8]$') then
			addChar(v)
			local num = tonumber(k:match('^-p([1-8])'))
			local player = 1
			if num % 2 == 0 then --even value
				player = 2
				p2NumChars = p2NumChars + 1
			else
				p1NumChars = p1NumChars + 1
			end
			local pal = 1
			if main.flags['-p' .. num .. '.pal'] ~= nil then
				pal = main.flags['-p' .. num .. '.pal']
			end
			local ai = 0
			if main.flags['-p' .. num .. '.ai'] ~= nil then
				ai = main.flags['-p' .. num .. '.ai']
			end
			t[#t + 1] = {player = player, num = num - 1, pal = tonumber(pal), ai = tonumber(ai)}
			refresh()
		end
	end
	local p1TeamMode = 0
	if p1NumChars > 1 then
		p1TeamMode = 1
	end
	local p2TeamMode = 0
	if p2NumChars > 1 then
		p2TeamMode = 1
	end
	--add stage
	local stage = 'stages/stage0.def'
	if main.flags['-s'] ~= nil then
		if main.f_fileExists(main.flags['-s']) then
			stage = main.flags['-s']
		elseif main.f_fileExists('stages/' .. main.flags['-s'] .. '.def') then
			stage = 'stages/' .. main.flags['-s'] .. '.def'
		end
	end
	addStage(stage)
	--load data
	loadDebugFont('font/f-6x9.fnt')
	setDebugScript('script/debug.lua')
	setMatchNo(1)
	selectStart()
	setStage(0)
	selectStage(0)
	setTeamMode(1, p1TeamMode, p1NumChars)
	setTeamMode(2, p2TeamMode, p2NumChars)
	for i = 1, #t do
		selectChar(t[i].player, t[i].num, t[i].pal)
		setCom(t[i].player, t[i].ai)
	end
	local winner, t_gameStats = game()
	if main.flags['-log'] ~= nil then
		main.f_printTable(t_gameStats, main.flags['-log'])
	end
	--exit ikemen
	return
end

--;===========================================================
--; LOAD DATA
--;===========================================================
main.fadeSff = sffNew('data/fade.sff')

motif = require('script.motif')

setPortrait(motif.select_info.p1_face_spr[1], motif.select_info.p1_face_spr[2], 1) --Big portrait
setPortrait(motif.select_info.portrait_spr[1], motif.select_info.portrait_spr[2], 2) --Small portrait
setPortrait(motif.vs_screen.p1_spr[1], motif.vs_screen.p1_spr[2], 3) --Versus portrait
setPortrait(motif.victory_screen.p1_spr[1], motif.victory_screen.p1_spr[2], 4) --Victory portrait

main.cursorBox = animNew(main.fadeSff, '0,1, 0,0, -1')
animSetTile(main.cursorBox, 1, 1)
animUpdate(main.cursorBox)

main.warningBox = animNew(main.fadeSff, '0,0, 0,0, -1')
animSetTile(main.warningBox, 1, 1)
animSetAlpha(main.warningBox, motif.warning_info.background_alpha[1], motif.warning_info.background_alpha[2])
animSetWindow(main.warningBox, 0, 0, motif.info.localcoord[1], motif.info.localcoord[2])
animUpdate(main.warningBox)

main.infoBox = animNew(main.fadeSff, '0,0, 0,0, -1')
animSetTile(main.infoBox, 1, 1)
animSetAlpha(main.infoBox, motif.infobox.background_alpha[1], motif.infobox.background_alpha[2])
animSetWindow(main.infoBox, 0, 0, motif.info.localcoord[1], motif.info.localcoord[2])
animUpdate(main.infoBox)


main.txt_warningTitle = main.f_createTextImg(
	motif.font_data[motif.warning_info.title_font[1]],
	motif.warning_info.title_font[2],
	motif.warning_info.title_font[3],
	motif.warning_info.title,
	motif.warning_info.title_pos[1],
	motif.warning_info.title_pos[2],
	motif.warning_info.title_font_scale[1],
	motif.warning_info.title_font_scale[2],
	motif.warning_info.title_font[4],
	motif.warning_info.title_font[5],
	motif.warning_info.title_font[6]
)

main.SetDefaultScale()

local footerBox = animNew(main.fadeSff, '0,2, 0,0, -1')
animSetTile(footerBox, 1, 1)
animSetWindow(
	footerBox,
	motif.title_info.footer_boxbackground_coords[1],
	motif.title_info.footer_boxbackground_coords[2],
	motif.title_info.footer_boxbackground_coords[3] - motif.title_info.footer_boxbackground_coords[1] + 1,
	motif.title_info.footer_boxbackground_coords[4] - motif.title_info.footer_boxbackground_coords[2] + 1
)
animSetAlpha(footerBox, motif.title_info.footer_boxbackground_alpha[1], motif.title_info.footer_boxbackground_alpha[2])
animUpdate(footerBox)

main.SetScaleValues()

--add characters and stages using select.def instead of select.lua
local txt_loading = main.f_createTextImg(
	motif.font_data[motif.title_info.loading_font[1]],
	motif.title_info.loading_font[2],
	motif.title_info.loading_font[3],
	motif.title_info.loading_text,
	motif.title_info.loading_offset[1],
	motif.title_info.loading_offset[2],
	motif.title_info.loading_font_scale[1],
	motif.title_info.loading_font_scale[2],
	motif.title_info.loading_font[4],
	motif.title_info.loading_font[5],
	motif.title_info.loading_font[6]
)
textImgDraw(txt_loading)
refresh()

function main.f_addChar(line, row)
	local tmp = ''
	local order = false
	local hidden = 0
	for i, c in ipairs(main.f_strsplit(',', line)) do
		c = c:match('^%s*(.-)%s*$')
		if i == 1 then
			c = c:gsub('\\', '/')
			c = tostring(c)
			main.t_selChars[row] = {}
			addChar(c)
			tmp = getCharName(row - 1):match('^["%s]*(.-)["%s]*$') --needed for s-size ikemen
			if tmp == '' then
				break
			end
			main.t_charDef[c] = row - 1
			main.t_selChars[row].char = c
			if tmp ~= 'Random' then
				main.t_selChars[row].displayname = tmp
				main.t_selChars[row].def = getCharFileName(row - 1)
				main.t_selChars[row].dir = main.t_selChars[row].def:gsub('[^/]+%.def$', '')
				main.t_selChars[row].pal = getCharPalettes(row - 1)
				if tmp ~= 'Training' then
					tmp = getCharIntro(row - 1)
					if tmp ~= '' then
						main.t_selChars[row].intro = main.t_selChars[row].dir .. tmp:gsub('\\', '/')
					end
					tmp = getCharEnding(row - 1)
					if tmp ~= '' then
						main.t_selChars[row].ending = main.t_selChars[row].dir .. tmp:gsub('\\', '/')
					end
					main.t_selChars[row].order = 1
					order = true
				end
			else
				break
			end
		elseif c:match('music%s*=%s*') then
			c = c:gsub('\\', '/')
			local bgmvolume = c:match('%s([0-9]+)$')
			if bgmvolume == nil then
				bgmvolume = 100
			else
				bgmvolume = tonumber(bgmvolume)
				c = c:gsub('%s*[0-9]+$','')
			end
			local bgmusic = c:match('^music%s*=%s*(.-)%s*$')
			if main.t_selChars[row].music == nil then
				main.t_selChars[row].music = {}
			end
			main.t_selChars[row].music[#main.t_selChars[row].music + 1] = {}
			main.t_selChars[row].music[#main.t_selChars[row].music].bgmusic = bgmusic
			main.t_selChars[row].music[#main.t_selChars[row].music].bgmvolume = bgmvolume
		elseif c:match('[0-9]+%s*=%s*[^%s]') then
			local var1, var2 = c:match('([0-9]+)%s*=%s*(.+)%s*$')
			main.t_selChars[row][tonumber(var1)] = var2:lower()
		elseif c:match('%.def') or c:match('^random$') then
			c = c:gsub('\\', '/')
			if main.t_selChars[row].stage == nil then
				main.t_selChars[row].stage = {}
			end
			main.t_selChars[row].stage[#main.t_selChars[row].stage + 1] = c
		else
			local param, value = c:match('^(.-)%s*=%s*(.-)$')
			if param ~= '' and value ~= '' and param ~= nil and value ~= nil then
				main.t_selChars[row][param] = tonumber(value)
			end
		end
	end
	if main.t_selChars[row].exclude == nil then
		main.t_selChars[row].hidden = hidden
	end
	if order then
		if main.t_orderChars[main.t_selChars[row].order] == nil then
			main.t_orderChars[main.t_selChars[row].order] = {}
		end
		main.t_orderChars[main.t_selChars[row].order][#main.t_orderChars[main.t_selChars[row].order] + 1] = row - 1
	end
	main.loadingRefresh(txt_loading)
end

--start_time = os.time()
main.t_includeStage = {}
main.t_orderChars = {}
main.t_stageDef = {['random'] = 0}
main.t_charDef = {}
local t_exlude = {}
local chars = 0
local stages = 0
local tmp = ''
local section = 0
local row = 0
local file = io.open(motif.files.select,"r")
local content = file:read("*all")
file:close()
content = content:gsub('([^\r\n;]*)%s*;[^\r\n]*', '%1')
content = content:gsub('\n%s*\n', '\n')
for line in content:gmatch('[^\r\n]+') do
	if chars + stages == 100 then
		SetGCPercent(100)
	end
--for line in io.lines("data/select.def") do
	line = line:lower()
	if line:match('^%s*%[%s*characters%s*%]') then
		main.t_selChars = {}
		row = 0
		section = 1
	elseif line:match('^%s*%[%s*extrastages%s*%]') then
		main.t_selStages = {}
		row = 0
		section = 2
	elseif line:match('^%s*%[%s*options%s*%]') then
		main.t_selOptions = {
			arcadestart = {wins = 0, offset = 0},
			arcadeend = {wins = 0, offset = 0},
			teamstart = {wins = 0, offset = 0},
			teamend = {wins = 0, offset = 0},
			survivalstart = {wins = 0, offset = 0},
			survivalend = {wins = 0, offset = 0},
		}
		row = 0
		section = 3
	elseif section == 1 then --[Characters]
		if line:match(',%s*exclude%s*=%s*1') then --character should be added after all slots are filled
			t_exlude[#t_exlude + 1] = line
		else
			chars = chars + 1
			main.f_addChar(line, chars)
		end
	elseif section == 2 then --[ExtraStages]
		row = #main.t_selStages + 1
		for i, c in ipairs(main.f_strsplit(',', line)) do
			c = c:gsub('^%s*(.-)%s*$', '%1')
			if i == 1 then
				c = c:gsub('\\', '/')
				if not main.f_fileExists(c) or c:match('^stages/$') then
					break
				end
				addStage(c)
				tmp = getStageName(row):match('^["%s]*(.-)["%s]*$') --needed for s-size ikemen
				if tmp == '' then
					break
				end
				main.t_selStages[row] = {name = tmp, stage = c}
				local zoomout, zoomin, bgmusic, bgmvolume, bgmloopstart, bgmloopend  = getStageInfo(row)
				if zoomout ~= '' then
					main.t_selStages[row].zoommin = tonumber(zoomout)
				end
				if zoomin ~= '' then
					main.t_selStages[row].zoommax = tonumber(zoomin)
				end
				if bgmusic ~= '' then
					bgmusic = bgmusic:gsub('\\', '/')
					main.t_selStages[row].music = {[1] = {bgmusic = bgmusic, bgmvolume = tonumber(bgmvolume), bgmloopstart = bgmloopstart, bgmloopend = bgmloopend}}
				end
				main.t_includeStage[#main.t_includeStage + 1] = row
				main.t_stageDef[c] = row
			elseif c:match('music%s*=%s*') then
				c = c:gsub('\\', '/')
				local bgmvolume = c:match('%s([0-9]+)$')
				if bgmvolume == nil then
					bgmvolume = 100
				else
					bgmvolume = tonumber(bgmvolume)
					c = c:gsub('%s*[0-9]+$','')
				end
				local bgmusic = c:match('^music%s*=%s*(.-)%s*$')
				if main.t_selStages[row].music == nil then
					main.t_selStages[row].music = {}
				end
				main.t_selStages[row].music[#main.t_selStages[row].music+1] = {bgmusic = bgmusic, bgmvolume = bgmvolume}
			else
				local param, value = c:match('^(.-)%s*=%s*(.-)$')
				main.t_selStages[row][param] = tonumber(value)
			end
		end
	elseif section == 3 then --[Options]
		if line:match('^%s*.-%.maxmatches%s*=%s*') then
			local rowName, line = line:match('^%s*(.-)%.maxmatches%s*=%s*(.+)')
			main.t_selOptions[rowName .. 'maxmatches'] = {}
			for i, c in ipairs(main.f_strsplit(',', line:gsub('%s*(.-)%s*', '%1'))) do
				main.t_selOptions[rowName .. 'maxmatches'][i] = tonumber(c)
			end
		elseif line:match('^%s*.-%.airamp%..-%s*=%s*') then
			local rowName, rowName2, wins, offset = line:match('^%s*(.-)%.airamp%.(.-)%s*=%s*([0-9]+)%s*,%s*([0-9-]+)')
			main.t_selOptions[rowName .. rowName2] = {wins = tonumber(wins), offset = tonumber(offset)}
		end
	end
end
--main.f_printVar(os.difftime(os.time(), start_time))

--add excluded characters once all slots are filled
for i = chars, (motif.select_info.rows + motif.select_info.rows_scrolling) * motif.select_info.columns - 1 do
	chars = chars + 1
	main.t_selChars[chars] = {}
	addChar('dummyChar')
end
for i = 1, #t_exlude do
	chars = chars + 1
	main.f_addChar(t_exlude[i], chars)
end
--add Training by stupa if not included in select.def
if main.t_charDef.training == nil and main.f_fileExists('chars/training/training.def') then
	chars = chars + 1
	main.f_addChar('training, exclude = 1', chars)
end

--add remaining character parameters
main.t_bossChars = {}
main.t_bonusChars = {}
main.t_randomChars = {}
--for each character loaded
for i = 1, #main.t_selChars do
	if main.t_selChars[i].stage ~= nil then
		for j = 1, #main.t_selChars[i].stage do
			if main.t_stageDef[main.t_selChars[i].stage[j]] == nil and main.file_exists(main.t_selChars[i].stage[j]) then
				row = #main.t_selStages + 1
				addStage(main.t_selChars[i].stage[j])
				tmp = getStageName(row):match('^["%s]*(.-)["%s]*$') --needed for s-size ikemen
				if tmp == '' then
					break
				end
				main.t_selStages[row] = {name = tmp, stage = main.t_selChars[i].stage[j]}
				local zoomout, zoomin, bgmusic, bgmvolume, bgmloopstart, bgmloopend = getStageInfo(row)
				if zoomout ~= '' then
					main.t_selStages[row].zoommin = tonumber(zoomout)
				end
				if zoomin ~= '' then
					main.t_selStages[row].zoommax = tonumber(zoomin)
				end
				if bgmusic ~= '' then
					bgmusic = bgmusic:gsub('\\', '/')
					main.t_selStages[row].music = {}
					main.t_selStages[row].music[1] = {bgmusic = bgmusic, bgmvolume = tonumber(bgmvolume), bgmloopstart = bgmloopstart, bgmloopend = bgmloopend}
				end
				if main.t_selChars[i].includestage == nil or main.t_selChars[i].includestage == 1 then
					main.t_includeStage[#main.t_includeStage + 1] = row
				end
				main.t_selChars[i].stage[j] = row
				--main.t_stageDef[main.t_selChars[i].stage[j]] = row
			else
				main.t_selChars[i].stage[j] = main.t_stageDef[main.t_selChars[i].stage[j]]
			end
		end
	end
	--if character's name has been stored
	if main.t_selChars[i].displayname ~= nil then
		--generate table for boss rush mode
		if main.t_selChars[i].boss ~= nil and main.t_selChars[i].boss == 1 then
			main.t_bossChars[#main.t_bossChars + 1] = i - 1
		end
		--generate table for bonus games mode
		if main.t_selChars[i].bonus ~= nil and main.t_selChars[i].bonus == 1 then
			main.t_bonusChars[#main.t_bonusChars + 1] = i - 1
		end
		--generate table with characters allowed to be random selected
		if main.t_selChars[i].hidden ~= nil and main.t_selChars[i].hidden <= 1 then
			main.t_randomChars[#main.t_randomChars + 1] = i - 1
		end
	end
end

--Save debug tables
main.f_printTable(main.t_selChars, "debug/t_selChars.txt")
main.f_printTable(main.t_selStages, "debug/t_selStages.txt")
main.f_printTable(main.t_selOptions, "debug/t_selOptions.txt")
main.f_printTable(main.t_orderChars, "debug/t_orderChars.txt")
main.f_printTable(main.t_randomChars, "debug/t_randomChars.txt")
main.f_printTable(main.t_bossChars, "debug/t_bossChars.txt")
main.f_printTable(main.t_bonusChars, "debug/t_bonusChars.txt")
main.f_printTable(main.t_stageDef, "debug/t_stageDef.txt")
main.f_printTable(main.t_charDef, "debug/t_charDef.txt")
main.f_printTable(main.t_includeStage, "debug/t_includeStage.txt")

--Debug stuff
loadDebugFont(motif.files.debug_font)
setDebugScript(motif.files.debug_script)

--Assign Lifebar
textImgDraw(txt_loading)
refresh()
loadLifebar(motif.files.fight)
main.loadingRefresh(txt_loading)

--warnings
if main.t_charDef.training == nil then
	main.f_warning(main.f_extractText(motif.warning_info.text_training), motif.title_info, motif.titlebgdef)
	os.exit()
end

if #main.t_includeStage == 0 then
	main.f_warning(main.f_extractText(motif.warning_info.text_stages), motif.title_info, motif.titlebgdef)
	os.exit()
end

local t = {'arcademaxmatches', 'teammaxmatches'}
for i = 1, #t do
	local orderOK = false
	for j = 1, #main.t_selOptions[t[i]] do
		if main.t_selOptions[t[i]][j] > 0 and main.t_orderChars[j] ~= nil and #main.t_orderChars[j] > 0 then
			orderOK = true
			break
		end
	end
	if not orderOK then
		main.f_warning(main.f_extractText(motif.warning_info.text_order), motif.title_info, motif.titlebgdef)
		os.exit()
	end
end

--Load additional scripts
randomtest = require('script.randomtest')
options = require('script.options')
select = require('script.select')
storyboard = require('script.storyboard')

--;===========================================================
--; MAIN MENU
--;===========================================================

--Disable screenpack scale on the text for showing them corectly.
main.SetDefaultScale()

local txt_titleFooter1 = main.f_createTextImg(
	motif.font_data[motif.title_info.footer1_font[1]],
	motif.title_info.footer1_font[2],
	motif.title_info.footer1_font[3],
	motif.title_info.footer1_text,
	motif.title_info.footer1_offset[1],
	motif.title_info.footer1_offset[2],
	motif.title_info.footer1_font_scale[1],
	motif.title_info.footer1_font_scale[2],
	motif.title_info.footer1_font[4],
	motif.title_info.footer1_font[5],
	motif.title_info.footer1_font[6]
)
local txt_titleFooter2 = main.f_createTextImg(
	motif.font_data[motif.title_info.footer2_font[1]],
	motif.title_info.footer2_font[2],
	motif.title_info.footer2_font[3],
	motif.title_info.footer2_text,
	motif.title_info.footer2_offset[1],
	motif.title_info.footer2_offset[2],
	motif.title_info.footer2_font_scale[1],
	motif.title_info.footer2_font_scale[2],
	motif.title_info.footer2_font[4],
	motif.title_info.footer2_font[5],
	motif.title_info.footer2_font[6]
)
local txt_titleFooter3 = main.f_createTextImg(
	motif.font_data[motif.title_info.footer3_font[1]],
	motif.title_info.footer3_font[2],
	motif.title_info.footer3_font[3],
	motif.title_info.footer3_text,
	motif.title_info.footer3_offset[1],
	motif.title_info.footer3_offset[2],
	motif.title_info.footer3_font_scale[1],
	motif.title_info.footer3_font_scale[2],
	motif.title_info.footer3_font[4],
	motif.title_info.footer3_font[5],
	motif.title_info.footer3_font[6]
)
local txt_infoboxTitle = main.f_createTextImg(
	motif.font_data[motif.infobox.title_font[1]],
	motif.infobox.title_font[2],
	motif.infobox.title_font[3],
	motif.infobox.title,
	motif.infobox.title_pos[1],
	motif.infobox.title_pos[2],
	motif.infobox.title_font_scale[1],
	motif.infobox.title_font_scale[2],
	motif.infobox.title_font[4],
	motif.infobox.title_font[5],
	motif.infobox.title_font[6]
)

--Enable screenpack scale again.
main.SetScaleValues()

main.txt_mainSelect = main.f_createTextImg(
	motif.font_data[motif.select_info.title_font[1]],
	motif.select_info.title_font[2],
	motif.select_info.title_font[3],
	'',
	motif.select_info.title_offset[1],
	motif.select_info.title_offset[2],
	motif.select_info.title_font_scale[1],
	motif.select_info.title_font_scale[2],
	motif.select_info.title_font[4],
	motif.select_info.title_font[5],
	motif.select_info.title_font[6]
)

--itemname: names used to distinguish modes in lua code (keep it as it is)
--displayname: names for each of the items in the menu
--selectname: names that will show up in select screen
local t_mainMenu = {
	{data = textImgNew(), itemname = 'arcade', displayname = motif.title_info.menu_itemname_arcade, selectname = motif.select_info.title_text_arcade},
	{data = textImgNew(), itemname = 'versus', displayname = motif.title_info.menu_itemname_versus, selectname = motif.select_info.title_text_versus},
	{data = textImgNew(), itemname = 'online', displayname = motif.title_info.menu_itemname_online},
	{data = textImgNew(), itemname = 'teamcoop', displayname = motif.title_info.menu_itemname_teamcoop, selectname = motif.select_info.title_text_teamcoop},
	{data = textImgNew(), itemname = 'survival', displayname = motif.title_info.menu_itemname_survival, selectname = motif.select_info.title_text_survival},
	{data = textImgNew(), itemname = 'survivalcoop', displayname = motif.title_info.menu_itemname_survivalcoop, selectname = motif.select_info.title_text_survivalcoop},
	--{data = textImgNew(), itemname = 'storymode', displayname = motif.title_info.menu_itemname_storymode, selectname = motif.select_info.title_text_storymode},
	--{data = textImgNew(), itemname = 'timeattack', displayname = motif.title_info.menu_itemname_timeattack, selectname = motif.select_info.title_text_timeattack},
	--{data = textImgNew(), itemname = 'tournament', displayname = motif.title_info.menu_itemname_tournament},
	{data = textImgNew(), itemname = 'training', displayname = motif.title_info.menu_itemname_training, selectname = motif.select_info.title_text_training},
	{data = textImgNew(), itemname = 'watch', displayname = motif.title_info.menu_itemname_watch, selectname = motif.select_info.title_text_watch},
	{data = textImgNew(), itemname = 'extras', displayname = motif.title_info.menu_itemname_extras},
	{data = textImgNew(), itemname = 'options', displayname = motif.title_info.menu_itemname_options},
	{data = textImgNew(), itemname = 'exit', displayname = motif.title_info.menu_itemname_exit},
}
t_mainMenu = main.f_cleanTable(t_mainMenu)

function main.f_default()
	setAutoLevel(false) --generate autolevel.txt in game dir
	setHomeTeam(2) --P2 side considered the home team: http://mugenguild.com/forum/topics/ishometeam-triggers-169132.0.html
	resetRemapInput()
	--settings adjustable via options
	setAutoguard(1, config.AutoGuard)
	setAutoguard(2, config.AutoGuard)
	setPowerShare(1, config.TeamPowerShare)
	setPowerShare(2, config.TeamPowerShare)
	setLifeShare(config.TeamLifeShare)
	setRoundTime(config.RoundTime * options.framespercount)
	setLifeMul(config.LifeMul / 100)
	setTeam1VS2Life(config.Team1VS2Life / 100)
	setTurnsRecoveryRate(1.0 / config.TurnsRecoveryRate)
	--default values for all modes
	main.p1Char = nil --no predefined P1 character (assigned via table: {X, Y, (...)})
	main.p2Char = nil --no predefined P2 character (assigned via table: {X, Y, (...)})
	main.p1TeamMenu = nil --no predefined P1 team mode (assigned via table: {mode = X, chars = Y})
	main.p2TeamMenu = nil --no predefined P2 team mode (assigned via table: {mode = X, chars = Y})
	main.aiFight = false --AI = config.Difficulty for all characters disabled
	main.stageMenu = false --stage selection disabled
	main.p2Faces = false --additional window with P2 select screen small portraits (faces) disabled
	main.coop = false --P2 fighting on P1 side disabled
	main.p2SelectMenu = true --P2 character selection enabled
	main.versusScreen = true --versus screen enabled
	main.p1In = 1 --P1 controls P1 side of the select screen
	main.p2In = 2 --P2 controls P2 side of the select screen
	main.gameMode = '' --additional variable used to distinguish modes in select screen
end

function main.f_menuCommon1(cursorPosY, moveTxt, item, t)
	if commandGetState(main.p1Cmd, 'u') or commandGetState(main.p2Cmd, 'u') then
		sndPlay(motif.files.snd_data, motif.title_info.cursor_move_snd[1], motif.title_info.cursor_move_snd[2])
		item = item - 1
	elseif commandGetState(main.p1Cmd, 'd') or commandGetState(main.p2Cmd, 'd') then
		sndPlay(motif.files.snd_data, motif.title_info.cursor_move_snd[1], motif.title_info.cursor_move_snd[2])
		item = item + 1
	end
	--cursor position calculation
	if item < 1 then
		item = #t
		if #t > motif.title_info.menu_window_visibleitems then
			cursorPosY = motif.title_info.menu_window_visibleitems
		else
			cursorPosY = #t
		end
	elseif item > #t then
		item = 1
		cursorPosY = 1
	elseif (commandGetState(main.p1Cmd, 'u') or commandGetState(main.p2Cmd, 'u') ) and cursorPosY > 1 then
		cursorPosY = cursorPosY - 1
	elseif (commandGetState(main.p1Cmd, 'd') or commandGetState(main.p2Cmd, 'd') ) and cursorPosY < motif.title_info.menu_window_visibleitems then
		cursorPosY = cursorPosY + 1
	end
	if cursorPosY == motif.title_info.menu_window_visibleitems then
		moveTxt = (item - motif.title_info.menu_window_visibleitems) * motif.title_info.menu_item_spacing[2]
	elseif cursorPosY == 1 then
		moveTxt = (item - 1) * motif.title_info.menu_item_spacing[2]
	end
	return cursorPosY, moveTxt, item
end

function main.f_menuCommon2(cursorPosY, moveTxt, item, t)
	--draw clearcolor
	animDraw(motif.titlebgdef.bgclearcolor_data)
	--draw layerno = 0 backgrounds
	main.f_drawBG(motif.titlebgdef.bg_data, motif.titlebgdef.bg, 0, motif.titlebgdef.timer, {320,240})
	--draw menu items
	local items_shown = item + motif.title_info.menu_window_visibleitems - cursorPosY
	if motif.title_info.menu_window_margins_y[2] ~= 0 and items_shown < #t then
		items_shown = items_shown + 1
	end
	if items_shown > #t then
		items_shown = #t
	end
	for i = 1, items_shown do
		if i > item - cursorPosY then
			if i == item then
				textImgDraw(main.f_updateTextImg(
					t[i].data,
					motif.font_data[motif.title_info.menu_item_active_font[1]],
					motif.title_info.menu_item_active_font[2],
					motif.title_info.menu_item_active_font[3],
					t[i].displayname,
					motif.title_info.menu_pos[1],
					motif.title_info.menu_pos[2] + (i - 1) * motif.title_info.menu_item_spacing[2] - moveTxt,
					motif.title_info.menu_item_active_font_scale[1],
					motif.title_info.menu_item_active_font_scale[2],
					motif.title_info.menu_item_active_font[4],
					motif.title_info.menu_item_active_font[5],
					motif.title_info.menu_item_active_font[6]
				))
			else
				textImgDraw(main.f_updateTextImg(
					t[i].data,
					motif.font_data[motif.title_info.menu_item_font[1]],
					motif.title_info.menu_item_font[2],
					motif.title_info.menu_item_font[3],
					t[i].displayname,
					motif.title_info.menu_pos[1],
					motif.title_info.menu_pos[2] + (i - 1) * motif.title_info.menu_item_spacing[2] - moveTxt,
					motif.title_info.menu_item_font_scale[1],
					motif.title_info.menu_item_font_scale[2],
					motif.title_info.menu_item_font[4],
					motif.title_info.menu_item_font[5],
					motif.title_info.menu_item_font[6]
				))
			end
		end
	end
	--draw menu cursor
	if motif.title_info.menu_boxcursor_visible == 1 then
		animSetWindow(
			main.cursorBox,
			motif.title_info.menu_pos[1] + motif.title_info.menu_boxcursor_coords[1],
			motif.title_info.menu_pos[2] + motif.title_info.menu_boxcursor_coords[2] + (cursorPosY - 1) * motif.title_info.menu_item_spacing[2],
			motif.title_info.menu_boxcursor_coords[3] - motif.title_info.menu_boxcursor_coords[1] + 1,
			motif.title_info.menu_boxcursor_coords[4] - motif.title_info.menu_boxcursor_coords[2] + 1
		)
		main.f_dynamicAlpha(main.cursorBox, 10,40,2, 255,255,0)
		animDraw(main.cursorBox)
	end
	--draw layerno = 1 backgrounds
	main.f_drawBG(motif.titlebgdef.bg_data, motif.titlebgdef.bg, 1, motif.titlebgdef.timer, {320,240})
	--footer draw
	if motif.title_info.footer_boxbackground_visible == 1 then
		animDraw(footerBox)
	end
	textImgDraw(txt_titleFooter1)
	textImgDraw(txt_titleFooter2)
	textImgDraw(txt_titleFooter3)
	--draw fadein
	animDraw(motif.title_info.fadein_data)
	animUpdate(motif.title_info.fadein_data)
	--update timer
	motif.titlebgdef.timer = motif.titlebgdef.timer + 1
	--end loop
	main.f_cmdInput()
	refresh()
end

function main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
	--[[for i = 1, motif.title_info.fadeout_time do
		main.f_menuCommon2(cursorPosY, moveTxt, item, t)
		animDraw(motif.title_info.fadeout_data)
		animUpdate(motif.title_info.fadeout_data)
	end]]
end

function main.f_mainMenu()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_mainMenu
	if motif.files.logo_storyboard ~= '' then
		storyboard.f_storyboard(motif.files.logo_storyboard)
	end
	if motif.files.intro_storyboard ~= '' then
		storyboard.f_storyboard(motif.files.intro_storyboard)
	end
	main.f_resetBG(motif.title_info, motif.titlebgdef, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
	while true do
		cursorPosY, moveTxt, item = main.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			break
		elseif getKey() == 'F1' then
			main.SetDefaultScale()
			main.f_warning(main.f_extractText(motif.infobox.text), motif.title_info, motif.titlebgdef, motif.infobox, txt_infoboxTitle, main.infoBox)
			main.SetScaleValues()
		elseif main.f_btnPalNo(main.p1Cmd) > 0 then
			main.f_default()
			--ARCADE
			if t[item].itemname == 'arcade' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.p2In = 1 --P1 controls P2 side of the select screen
				main.p2SelectMenu = false --P2 character selection disabled
				main.credits = config.Credits - 1 --amount of continues
				main.gameMode = t[item].itemname --mode recognized in select screen as 'arcade'
				textImgSetText(main.txt_mainSelect, t[item].selectname) --message displayed on top of select screen
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectAdvance() --start f_selectAdvance() function from script/select.lua
			end
			--VS MODE
			if t[item].itemname == 'versus' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				setHomeTeam(1) --P1 side considered the home team
				main.p2In = 2 --P2 controls P2 side of the select screen
				main.stageMenu = true --stage selection enabled
				main.p2Faces = true --additional window with P2 select screen small portraits (faces) enabled
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectSimple() --start f_selectSimple() function from script/select.lua
			end
			--ONLINE
			if t[item].itemname == 'online' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.f_mainNetplay()
			end
			--TEAM CO-OP
			if t[item].itemname == 'teamcoop' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.p2In = 2
				main.p2Faces = true
				main.coop = true --P2 fighting on P1 side enabled
				main.credits = config.Credits - 1
				main.gameMode = t[item].itemname
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectAdvance()
			end
			--SURVIVAL
			if t[item].itemname == 'survival' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.p2In = 1
				main.p2SelectMenu = false
				main.credits = 0
				main.gameMode = t[item].itemname
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectAdvance()
			end
			--SURVIVAL CO-OP
			if t[item].itemname == 'survivalcoop' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.p2In = 2
				main.p2Faces = true
				main.coop = true
				main.credits = 0
				main.gameMode = t[item].itemname
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectAdvance()
			end
			--TOURNAMENT
			if t[item].itemname == 'tournament' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.f_mainTournament()
			end
			--TRAINING
			if t[item].itemname == 'training' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				setRoundTime(-1) --round time disabled
				main.p2In = 2
				main.stageMenu = true
				main.versusScreen = false --versus screen disabled
				main.p2TeamMenu = {mode = 0, chars = 1} --predefined P2 team mode as Single, 1 Character
				main.p2Char = {main.t_charDef.training} --predefined P2 character as Training by stupa
				main.gameMode = t[item].itemname
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectSimple()
			end
			--WATCH
			if t[item].itemname == 'watch' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.p2In = 1
				main.aiFight = true --AI = config.Difficulty for all characters enabled
				main.stageMenu = true
				main.p2Faces = true
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectSimple()
			end
			--EXTRAS
			if t[item].itemname == 'extras' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.f_mainExtras()
			end
			--OPTIONS
			if t[item].itemname == 'options' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)				
				--Disable screenpack scale on the menu text for showing the menu corectly.
				main.SetDefaultScale()
				options.f_mainCfg() --start f_mainCfg() function from script/options.lua
				--Enable screenpack scale again.
				main.SetScaleValues()
			end
			--EXIT
			if t[item].itemname == 'exit' then
				break
			end
		end
		main.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; NETPLAY MENU
--;===========================================================
local txt_connection = main.f_createTextImg(
	motif.font_data[motif.title_info.connecting_font[1]],
	motif.title_info.connecting_font[2],
	motif.title_info.connecting_font[3],
	"",
	motif.title_info.connecting_offset[1],
	motif.title_info.connecting_offset[2],
	motif.title_info.connecting_font_scale[1],
	motif.title_info.connecting_font_scale[2],
	motif.title_info.connecting_font[4],
	motif.title_info.connecting_font[5],
	motif.title_info.connecting_font[6]
)
function main.f_connect(server, t)
	local cancel = false
	enterNetPlay(server)
	while not connected() do
		if esc() then
			sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
			cancel = true
			break
		end
		--draw clearcolor
		animDraw(motif.titlebgdef.bgclearcolor_data)
		--draw layerno = 0 backgrounds
		main.f_drawBG(motif.titlebgdef.bg_data, motif.titlebgdef.bg, 0, motif.titlebgdef.timer, {320,240})
		--draw layerno = 1 backgrounds
		main.f_drawBG(motif.titlebgdef.bg_data, motif.titlebgdef.bg, 1, motif.titlebgdef.timer, {320,240})
		--draw menu box
		animDraw(main.warningBox)
		--draw text
		for i = 1, #t do
			textImgSetText(txt_connection, t[i])
			textImgDraw(txt_connection)
		end
		--update timer
		motif.titlebgdef.timer = motif.titlebgdef.timer + 1
		--end loop
		refresh()
	end
	main.f_cmdInput()
	if not cancel then
		synchronize()
		math.randomseed(sszRandom())
		main.f_netplayMode()
	end
end

local t_mainNetplay = {
	{data = textImgNew(), itemname = 'serverhost', displayname = motif.title_info.menu_itemname_serverhost},
	{data = textImgNew(), itemname = 'serverjoin', displayname = motif.title_info.menu_itemname_serverjoin},
	{data = textImgNew(), itemname = 'serverback', displayname = motif.title_info.menu_itemname_serverback},
}
t_mainNetplay = main.f_cleanTable(t_mainNetplay)

function main.f_mainNetplay()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_mainNetplay
	while true do
		cursorPosY, moveTxt, item = main.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
			break
		elseif main.f_btnPalNo(main.p1Cmd) > 0 then
			main.f_default()
			--HOST
			if t[item].itemname == 'serverhost' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.f_connect("", main.f_extractText(motif.title_info.connecting_host, getListenPort()))
				exitNetPlay()
				exitReplay()

				-- Save replay with a new name
				tpmFileRpl1 = io.open("save/replays/netplay.replay","r")
				tpmFileRpl1C = tpmFileRpl1:read("*all")
				io.close(tpmFileRpl1)

				tpmFileRpl2name = os.date("%Y-%m(%b)-%d %I-%M%p-%Ss")
				tpmFileRpl2 = io.open("save/replays/" .. tpmFileRpl2name .. ".replay","w+")
				tpmFileRpl2:write(tpmFileRpl1C)
				io.close(tpmFileRpl2)
			end
			--JOIN
			if t[item].itemname == 'serverjoin' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.f_netplayJoin()
			end
			--BACK
			if t[item].itemname == 'serverback' then
				sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
				break
			end
		end
		main.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; NETPLAY JOIN
--;===========================================================
local t_netplayJoin = {}
t_netplayJoin[#t_netplayJoin + 1] = {data = textImgNew(), itemname = 'joinadd', displayname = motif.title_info.menu_itemname_joinadd}
for k, v in pairs(config.IP) do
	t_netplayJoin[#t_netplayJoin + 1] = {
		data = textImgNew(),
		itemname = k,
		displayname = k,
		address = v
	}
end
t_netplayJoin[#t_netplayJoin + 1] = {data = textImgNew(), itemname = 'joinback', displayname = motif.title_info.menu_itemname_joinback}

function main.f_netplayJoin()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_netplayJoin
	local t_tmp = {}
	while true do
		cursorPosY, moveTxt, item = main.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
			break
		--DELETE ENTRY
		elseif getKey() == 'DELETE' and item ~= 1 and item ~= #t then
			sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
			resetKey()
			config.IP[t[item].itemname] = nil
			t_tmp = {}
			for i = 1, #t do
				if i ~= item then
					t_tmp[#t_tmp + 1] = t[i]
				end
			end
			t_netplayJoin = t_tmp
			t = t_netplayJoin
			local file = io.open("save/config.json","w+")
			file:write(json.encode(config, {indent = true}))
			file:close()
		elseif main.f_btnPalNo(main.p1Cmd) > 0 then
			main.f_default()
			--NEW ADDRESS
			if t[item].itemname == 'joinadd' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_move_snd[1], motif.title_info.cursor_move_snd[2])
				local name = main.f_input(main.f_extractText(motif.title_info.input_ip_name), motif.title_info, motif.titlebgdef, 'string')
				if name ~= '' then
					sndPlay(motif.files.snd_data, motif.title_info.cursor_move_snd[1], motif.title_info.cursor_move_snd[2])
					local address = main.f_input(main.f_extractText(motif.title_info.input_ip_address), motif.title_info, motif.titlebgdef, 'string')
					if address:match('^[0-9%.]+$') then
						sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
						config.IP[name] = address
						t_tmp = {}
						for i = 1, #t do
							if i < #t then
								t_tmp[i] = t[i]
							else
								t_tmp[i] = {data = textImgNew(), itemname = name, displayname = name, address = address}
								t_tmp[i + 1] = t[i]
							end
						end
						t_netplayJoin = t_tmp
						t = t_netplayJoin
						local file = io.open("save/config.json","w+")
						file:write(json.encode(config, {indent = true}))
						file:close()
					else
						sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
					end
				else
					sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
				end
			--BACK
			elseif t[item].itemname == 'joinback' then
				sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
				break
			--CONNECTION
			else
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.f_connect(t[item].address, main.f_extractText(motif.title_info.connecting_join, t[item].name, t[item].address))
				exitNetPlay()
				exitReplay()
			end
		end
		main.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; NETPLAY MODE
--;===========================================================
local t_netplayMode = {
	{data = textImgNew(), itemname = 'netplayversus', displayname = motif.title_info.menu_itemname_netplayversus, selectname = motif.select_info.title_text_netplayversus},
	{data = textImgNew(), itemname = 'netplayteamcoop', displayname = motif.title_info.menu_itemname_netplayteamcoop, selectname = motif.select_info.title_text_netplayteamcoop},
	{data = textImgNew(), itemname = 'netplaysurvivalcoop', displayname = motif.title_info.menu_itemname_netplaysurvivalcoop, selectname = motif.select_info.title_text_netplaysurvivalcoop},
	{data = textImgNew(), itemname = 'netplayback', displayname = motif.title_info.menu_itemname_netplayback},
}
t_netplayMode = main.f_cleanTable(t_netplayMode)

function main.f_netplayMode()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_netplayMode
	while true do
		cursorPosY, moveTxt, item = main.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
			break
		elseif main.f_btnPalNo(main.p1Cmd) > 0 then
			main.f_default()
			--VS MODE
			if t[item].itemname == 'netplayversus' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				setHomeTeam(1)
				main.p2In = 2
				main.stageMenu = true
				main.p2Faces = true
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				select.f_selectSimple()
			end
			--TEAM CO-OP
			if t[item].itemname == 'netplayteamcoop' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.p2In = 2
				main.p2Faces = true
				main.coop = true
				main.credits = config.Credits - 1
				main.gameMode = t[item].itemname
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				select.f_selectAdvance()
			end
			--SURVIVAL CO-OP
			if t[item].itemname == 'netplaysurvivalcoop' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.p2In = 2
				main.p2Faces = true
				main.coop = true
				main.credits = 0
				main.gameMode = t[item].itemname
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				select.f_selectAdvance()
			end
			--BACK
			if t[item].itemname == 'netplayback' then
				sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
				break
			end
		end
		main.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; EXTRAS MENU
--;===========================================================
local t_mainExtras = {
	{data = textImgNew(), itemname = 'freebattle', displayname = motif.title_info.menu_itemname_freebattle, selectname = motif.select_info.title_text_freebattle},
	--{data = textImgNew(), itemname = 'timechallenge', displayname = motif.title_info.menu_itemname_timechallenge, selectname = motif.select_info.title_text_timechallenge},
	--{data = textImgNew(), itemname = 'scorechallenge', displayname = motif.title_info.menu_itemname_scorechallenge, selectname = motif.select_info.title_text_scorechallenge},
	{data = textImgNew(), itemname = '100kumite', displayname = motif.title_info.menu_itemname_100kumite, selectname = motif.select_info.title_text_100kumite},
	{data = textImgNew(), itemname = 'bossrush', displayname = motif.title_info.menu_itemname_bossrush, selectname = motif.select_info.title_text_bossrush},
	{data = textImgNew(), itemname = 'bonusgames', displayname = motif.title_info.menu_itemname_bonusgames},
	--{data = textImgNew(), itemname = 'scoreranking', displayname = motif.title_info.menu_itemname_scoreranking},
	{data = textImgNew(), itemname = 'replay', displayname = motif.title_info.menu_itemname_replay, selectname = motif.select_info.title_text_replay},
	{data = textImgNew(), itemname = 'demo', displayname = motif.title_info.menu_itemname_demo},
	{data = textImgNew(), itemname = 'extrasback', displayname = motif.title_info.menu_itemname_extrasback},
}
for i = 1, #t_mainExtras do
	if t_mainExtras[i].itemname == 'bossrush' and #main.t_bossChars == 0 then
		t_mainExtras[i].displayname = ''
	elseif t_mainExtras[i].itemname == 'bonusgames' and #main.t_bonusChars == 0 then
		t_mainExtras[i].displayname = ''
	elseif t_mainExtras[i].itemname == 'demo' and #main.t_randomChars < 2 then
		t_mainExtras[i].displayname = ''
	end
end
t_mainExtras = main.f_cleanTable(t_mainExtras)

function main.f_mainExtras()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_mainExtras
	while true do
		cursorPosY, moveTxt, item = main.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
			break
		elseif main.f_btnPalNo(main.p1Cmd) > 0 then
			main.f_default()
			--FREE BATTLE
			if t[item].itemname == 'freebattle' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.p2In = 1
				main.stageMenu = true
				main.p2Faces = true
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectSimple()
			end
			--VS 100 KUMITE
			if t[item].itemname == '100kumite' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.p2In = 1
				main.p2SelectMenu = false
				main.gameMode = t[item].itemname
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectAdvance()
			end
			--BOSS RUSH
			if t[item].itemname == 'bossrush' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.p2In = 1
				main.p2SelectMenu = false
				main.credits = 0
				main.gameMode = t[item].itemname
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectAdvance()
			end
			--BONUS GAMES
			if t[item].itemname == 'bonusgames' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.f_bonusExtras()
			end
			--REPLAY
			if t[item].itemname == 'replay' then
				if main.f_fileExists('save/replays/netplay.replay') then
					sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
					enterReplay('save/replays/netplay.replay')
					synchronize()
					math.randomseed(sszRandom())
					main.f_netplayMode()
					exitNetPlay()
					exitReplay()
				end
			end
			--DEMO
			if t[item].itemname == 'demo' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				randomtest.run()
			end
			--BACK
			if t[item].itemname == 'extrasback' then
				sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
				break
			end
		end
		main.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; BONUS GAMES
--;===========================================================
local t_bonusExtras = {}
for i = 1, #main.t_bonusChars do
	local name = getCharName(main.t_bonusChars[i])
	t_bonusExtras[i] = {
		data = textImgNew(),
		itemname = name,
		displayname = name:upper(),
		selectname = name
	}
end
if motif.title_info.menu_itemname_bonusback ~= '' then
	t_bonusExtras[#t_bonusExtras + 1] = {data = textImgNew(), itemname = 'bonusback', displayname = motif.title_info.menu_itemname_bonusback}
end

function main.f_bonusExtras()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_bonusExtras
	while true do
		cursorPosY, moveTxt, item = main.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
			break
		elseif main.f_btnPalNo(main.p1Cmd) > 0 then
			main.f_default()
			--BACK
			if t[item].itemname == 'bonusback' then
				sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
				break
			--BONUS CHAR NAME
			else
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.versusScreen = false
				main.p2In = 1
				main.p1TeamMenu = {mode = 0, chars = 1}
				main.p2TeamMenu = {mode = 0, chars = 1}
				main.p2Char = {main.t_bonusChars[item]}
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectSimple()
			end
		end
		main.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; TOURNAMENT MENU
--;===========================================================
local t_mainTournament = {
	{data = textImgNew(), itemname = 'tourney32', displayname = motif.title_info.menu_itemname_tourney32, selectname = motif.select_info.title_text_tourney32},
	{data = textImgNew(), itemname = 'tourney16', displayname = motif.title_info.menu_itemname_tourney16, selectname = motif.select_info.title_text_tourney16},
	{data = textImgNew(), itemname = 'tourney8', displayname = motif.title_info.menu_itemname_tourney8, selectname = motif.select_info.title_text_tourney8},
	{data = textImgNew(), itemname = 'tourney4', displayname = motif.title_info.menu_itemname_tourney4, selectname = motif.select_info.title_text_tourney4},
	{data = textImgNew(), itemname = 'tourneyback', displayname = motif.title_info.menu_itemname_tourneyback},
}
t_mainTournament = main.f_cleanTable(t_mainTournament)

function main.f_mainTournament()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_mainTournament
	while true do
		cursorPosY, moveTxt, item = main.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
			break
		elseif main.f_btnPalNo(main.p1Cmd) > 0 then
			main.f_default()
			--ROUND OF 32
			if t[item].itemname == 'tourney32' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.gameMode = t[item].itemname
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectTournament()
			end
			--ROUND OF 16
			if t[item].itemname == 'tourney16' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.gameMode = t[item].itemname
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectTournament()
			end
			--QUARTERFINALS
			if t[item].itemname == 'tourney8' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.gameMode = t[item].itemname
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectTournament()
			end
			--SEMIFINALS
			if t[item].itemname == 'tourney4' then
				sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
				main.gameMode = t[item].itemname
				textImgSetText(main.txt_mainSelect, t[item].selectname)
				main.f_menuFadeOut(cursorPosY, moveTxt, item, t)
				select.f_selectTournament()
			end
			--BACK
			if t[item].itemname == 'tourneyback' then
				sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
				break
			end
		end
		main.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; INITIALIZE LOOPS
--;===========================================================

-- Now that everithig is loaded we can enable GC back.
SetGCPercent(100)
-------------------------------------------------------------

main.f_mainMenu()


-- Debug Info
main.f_printTable(main, "debug/t_main.txt")

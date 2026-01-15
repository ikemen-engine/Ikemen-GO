main = {}
--nClock = os.clock()
--print("Elapsed time: " .. os.clock() - nClock)
--;===========================================================
--; INITIALIZE DATA
--;===========================================================
math.randomseed(os.time())

--;===========================================================
--; COMMON FUNCTIONS
--;===========================================================

--return file content
function main.f_fileRead(path, mode, noError)
	local file = io.open(path, mode or 'r')
	if not file then
		if not noError then
			panicError("\nFile doesn't exist: " .. path)
		end
		return nil
	end

	local str = file:read("*all")
	file:close()
	return str
end

--write to file
function main.f_fileWrite(path, str, mode)
	if str == nil then
		return
	end
	local file = io.open(path, mode or 'w+')
	if file == nil then
		panicError("\nFile doesn't exist: " .. path)
		return
	end
	file:write(str)
	file:close()
end

--add default commands
main.t_commands = {
	{name = "$U", input = "$U"},
	{name = "$D", input = "$D"},
	{name = "$B", input = "$B"},
	{name = "$F", input = "$F"},
	{name = "a",  input = "a"},
	{name = "b",  input = "b"},
	{name = "c",  input = "c"},
	{name = "x",  input = "x"},
	{name = "y",  input = "y"},
	{name = "z",  input = "z"},
	{name = "s",  input = "s"},
	{name = "d",  input = "d"},
	{name = "w",  input = "w"},
	{name = "m",  input = "m"},
	{name = "/s", input = "/s"},
	{name = "/d", input = "/d"},
	{name = "/w", input = "/w"},
}
function main.f_commandNew(controllerNo)
	local c = commandNew(controllerNo)
	for _, v in ipairs(main.t_commands) do
		commandAdd(c, v.name, v.input)
	end
	return c
end

main.t_defaultKeysMapping = {
	up = 'Not used',
	down = 'Not used',
	left = 'Not used',
	right = 'Not used',
	a = 'Not used',
	b = 'Not used',
	c = 'Not used',
	x = 'Not used',
	y = 'Not used',
	z = 'Not used',
	start = 'Not used',
	d = 'Not used',
	w = 'Not used',
	menu = 'Not used',
}

main.t_defaultJoystickMapping = {
	up = 'DP_U',
	down = 'DP_D',
	left = 'DP_L',
	right = 'DP_R',
	a = 'A',
	b = 'B',
	c = 'RT',
	x = 'X',
	y = 'Y',
	z = 'RB',
	start = 'START',
	d = 'LB',
	w = 'LT',
	menu = 'BACK',
}

--prepare players/command tables
function main.f_setPlayers()
	local n = gameOption('Config.Players')
	setPlayers(n)
	for i = 3, n do
		if gameOption('Keys_P' .. i .. '.Joystick') == 0 then
			local c = {Joystick = -1, GUID = ""}
			for k, v in pairs(main.t_defaultKeysMapping) do
				c[k] = v
			end
			setKeyConfig(i, c.Joystick, {c.up, c.down, c.left, c.right, c.a, c.b, c.c, c.x, c.y, c.z, c.start, c.d, c.w, c.menu})
			for k, v in pairs(c) do
				modifyGameOption('Keys_P' .. i .. '.' .. k, v)
			end
		end
		if getCommandLineValue("-nojoy") == nil then
			if gameOption('Joystick_P' .. i .. '.Joystick') == 0 then
				local c = {Joystick = i - 1, GUID = ""}
				for k, v in pairs(main.t_defaultJoystickMapping) do
					c[k] = v
				end
				setKeyConfig(i, c.Joystick, {c.up, c.down, c.left, c.right, c.a, c.b, c.c, c.x, c.y, c.z, c.start, c.d, c.w, c.menu})
				for k, v in pairs(c) do
					modifyGameOption('Joystick_P' .. i .. '.' .. k, v)
				end
			end
		end
	end
	main.t_players = {}
	main.t_remaps = {}
	main.t_lastInputs = {}
	main.t_cmd = {}
	main.t_pIn = {}
	for i = 1, n do
		table.insert(main.t_players, i)
		table.insert(main.t_remaps, i)
		table.insert(main.t_lastInputs, {})
		table.insert(main.t_cmd, main.f_commandNew(i))
		table.insert(main.t_pIn, i)
	end
end
main.f_setPlayers()

--add new commands
function main.f_commandAdd(name, cmd, tim, buf)
	if main.t_commands[name] ~= nil then
		return
	end
	for i = 1, #main.t_cmd do
		commandAdd(main.t_cmd[i], name, cmd, tim or 15, buf or 1)
	end
	main.t_commands[name] = 0
end
--main.f_commandAdd("KonamiCode", "~U,U,D,D,B,F,B,F,b,a,s", 300, 1)

--sends inputs to buffer
function main.f_cmdInput()
	for i = 1, gameOption('Config.Players') do
		if main.t_pIn[i] > 0 then
			commandInput(main.t_cmd[i], main.t_pIn[i])
		end
	end
end

--resets command buffer
function main.f_cmdBufReset(pn)
	esc(false)
	if pn ~= nil then
		commandBufReset(main.t_cmd[pn])
		main.f_cmdInput()
		return
	end
	for i = 1, gameOption('Config.Players') do
		commandBufReset(main.t_cmd[i])
	end
	main.f_cmdInput()
end

--returns value depending on button pressed (a = 1; a + start = 7 etc.)
function main.f_btnPalNo(p)
	local s = 0
	if commandGetState(main.t_cmd[p], '/s') then s = 6 end
	for i, k in pairs({'a', 'b', 'c', 'x', 'y', 'z'}) do
		if commandGetState(main.t_cmd[p], k) then return i + s end
	end
	return 0
end

--return bool based on command input
main.playerInput = 1
function main.f_input(p, ...)
	-- Centralized input read (Go-side), shared with motif.button()
	local ok = getInput(p, ...)
	if ok then
		local li = getLastInputController()
		if li ~= nil and li > 0 then
			main.playerInput = li
		end
	end
	return ok
end

--remap active players input
function main.f_playerInput(src, dst)
	main.t_remaps[src] = dst
	main.t_remaps[dst] = src
	remapInput(src, dst)
	remapInput(dst, src)
end

--restore screenpack remapped inputs
function main.f_restoreInput()
	if start.challenger > 0 then
		return
	end
	resetRemapInput()
	for k, v in ipairs(main.t_remaps) do
		if k ~= v then
			remapInput(k, v)
			remapInput(v, k)
		end
	end
end

--check if file exists
function main.f_fileExists(file)
	if file == '' then
		return false
	end
	return fileExists(file)
end

--prints "t" table content into "toFile" file
function main.f_printTable(t, toFile)
	local txt = ''
	local print_t_cache = {}
	local function sub_print_t(t, indent)
		if print_t_cache[tostring(t)] then
			txt = txt .. indent .. '*' .. tostring(t) .. '\n'
		else
			print_t_cache[tostring(t)] = true
			if type(t) == 'table' then
				for pos, val in pairs(t) do
					local k = (type(pos) == 'string') and ('[' .. string.format('%q', pos) .. ']') or  ('[' .. tostring(pos) .. ']')
					if type(val) == 'table' then
						txt = txt .. indent .. k .. ' => ' .. tostring(val) .. ' {' .. '\n'
						sub_print_t(val, indent .. string.rep(' ', string.len(k) + 6))
						txt = txt .. indent .. string.rep(' ', string.len(k) + 4) .. '}' .. '\n'
					elseif type(val) == 'string' then
						txt = txt .. indent .. k .. ' => "' .. val .. '"' .. '\n'
					else
						txt = txt .. indent .. k .. ' => ' .. tostring(val) ..'\n'
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
	main.f_fileWrite(toFile or 'debug/table_print.txt', txt)
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

--escape ().%+-*?[^$ characters
function main.f_escapePattern(str)
	return str:gsub('([^%w])', '%%%1')
end

--return argument or default value
function main.f_arg(arg, default)
	if arg ~= nil then
		return arg
	end
	return default
end

--command line global flags
if getCommandLineValue("-ailevel") ~= nil then
	modifyGameOption('Options.Difficulty', math.max(1, math.min(tonumber(getCommandLineValue("-ailevel")), 8)))
end
if getCommandLineValue("-speed") ~= nil then
	local speed_input = tonumber(getCommandLineValue("-speed"))
	if speed_input ~= nil and speed_input >= -9 and speed_input <= 9 then
		local target_game_speed
		if speed_input > 0 then
			target_game_speed = speed_input
		elseif speed_input == 0 then
			target_game_speed = 0
		else
			if speed_input >= -1 then
				target_game_speed = speed_input * 5
			else
				target_game_speed = speed_input * 0.5 - 4.5
			end
		end
		setGameSpeed(target_game_speed)
	end
end
if getCommandLineValue("-speedtest") ~= nil then
	setGameSpeed(100)
end
if getCommandLineValue("-nosound") ~= nil then
	modifyGameOption('Sound.MasterVolume', 0)
end
if getCommandLineValue("-togglelifebars") ~= nil then
	toggleLifebarDisplay()
end
if getCommandLineValue("-maxpowermode") ~= nil then
	toggleMaxPowerMode()
end
if getCommandLineValue("-debug") ~= nil then
	toggleDebugDisplay()
end
if getCommandLineValue("-setport") ~= nil then
	setListenPort(getCommandLineValue("-setport"))
end
if getCommandLineValue("-setvolume") ~= nil and getCommandLineValue("-nosound") == nil then
	modifyGameOption('Sound.MasterVolume', getCommandLineValue("-setvolume"))
end
if getCommandLineValue("-windowed") ~= nil then
	modifyGameOption('Video.Fullscreen', false)
end
if getCommandLineValue("-width") ~= nil then
	modifyGameOption('Video.GameWidth', getCommandLineValue("-width"))
end
if getCommandLineValue("-height") ~= nil then 
	modifyGameOption('Video.GameHeight', getCommandLineValue("-height"))
end

-- Lua Hook System
-- Allows hooking additional code into existing functions, from within external
-- modules, without having to worry as much about your code being removed by
-- engine update.
-- * hook.run(list, ...): Runs all the functions within a certain list.
--   It won't do anything if the list doesn't exist or is empty. ... is any
--   number of arguments, which will be passed to every function in the list.
-- * hook.add(list, name, function): Adds a function to a hook list with a name.
--   It will replace anything in the list with the same name.
-- * hook.stop(list, name): Removes a hook from a list, if it's not needed.
-- Currently there are only few hooks available by default:
-- * loop: global.lua 'loop' function start (called by CommonLua)
-- * loop#[gamemode]: global.lua 'loop' function, limited to the gamemode
-- * main.f_commandLine: main.lua 'f_commandLine' function (before loading)
-- * main.f_default: main.lua 'f_default' function
-- * main.t_itemname: main.lua table entries (modes configuration)
-- * main.menu.loop: main.lua menu loop function (each submenu loop start)
-- * menu.menu.loop: menu.lua menu loop function (each submenu loop start)
-- * options.menu.loop: options.lua menu loop function (each submenu loop start)
-- * launchFight: start.lua 'launchFight' function (right before match starts)
-- * start.f_selectScreen: start.lua 'f_selectScreen' function (pre layerno=1)
-- * start.f_selectVersus: start.lua 'f_selectVersus' function (pre layerno=1)
-- * start.f_selectReset: start.lua 'f_selectReset' function (before returning)
-- * game.result_init: hook executed on the 1st frame of win/results screen
-- * game.result: hook executed from 2nd frame onward on win/results screen
-- * game.victory_init: hook executed on the 1st frame of victory screen
-- * game.victory: hook executed from 2nd frame onward on victory screen
-- * game.continue_init: hook executed on the 1st frame of continue screen
-- * game.continue: hook executed at the 2nd and later frames of continue screen
-- * game.hiscore_init: hook executed on the 1st frame of hiscore screen
-- * game.hiscore: hook executed from 2nd frame onward on hiscore screen
-- * game.challenger_init: hook executed on the 1st frame of challenger screen
-- * game.challenger: hook executed from 2nd frame onward on challenger screen
-- More entry points may be added in future - let us know if your external
-- module needs to hook code in place where it's not allowed yet.

hook = {
	lists = {}
}
function hook.add(list, name, func)
	if hook.lists[list] == nil then
		hook.lists[list] = {}
	end
	hook.lists[list][name] = func
end
function hook.run(list, ...)
	if hook.lists[list] then
		for i, k in pairs(hook.lists[list]) do
			k(...)
		end
	end
end
function hook.stop(list, name)
	hook.lists[list][name] = nil
end

--animDraw at specified coordinates
function main.f_animPosDraw(a, x, y, f)
	if a == nil then
		return
	end
	if x ~= nil then
		animReset(a, {"pos"})
		animAddPos(a, x, y)
	end
	if f ~= nil then
		animSetFacing(a, f)
	end
	animDraw(a)
	animUpdate(a)
end

--screen fade animation
function main.f_fadeAnim(fadeGroup)
	--draw fade anim
	if main.fadeCnt > 0 then
		if fadeGroup[main.fadeType].AnimData ~= nil then
			animDraw(fadeGroup[main.fadeType].AnimData)
			animUpdate(fadeGroup[main.fadeType].AnimData)
		end
		main.fadeCnt = main.fadeCnt - 1
	end
	--play fade snd
	if main.fadeStart == getFrameCount() then
		sndPlay(motif.Snd, fadeGroup[main.fadeType].snd[1], fadeGroup[main.fadeType].snd[2])
	end
	--draw fadein / fadeout
	main.fadeActive = fadeColor(
		main.fadeType,
		main.fadeStart,
		fadeGroup[main.fadeType].time,
		fadeGroup[main.fadeType].col[1],
		fadeGroup[main.fadeType].col[2],
		fadeGroup[main.fadeType].col[3]
	)
end

--reset fade
main.fadeActive = false
function main.f_fadeReset(fadeType, fadeGroup)
	main.fadeType = fadeType
	main.fadeGroup = fadeGroup
	main.fadeStart = getFrameCount()
	main.fadeCnt = 0
	if fadeGroup[fadeType].AnimData ~= nil then
		animReset(fadeGroup[fadeType].AnimData)
		animUpdate(fadeGroup[fadeType].AnimData)
		main.fadeCnt = animGetLength(fadeGroup[fadeType].AnimData) - 1
		if fadeType == 'fadeout' and main.fadeCnt > fadeGroup[fadeType].time then
			main.fadeStart = main.fadeStart + main.fadeCnt - fadeGroup[fadeType].time
		end
	end
end

--copy table content into new table
function main.f_tableCopy(t)
	if t == nil then
		return nil
	end
	t = t or {}
	local t2 = {}
	for k, v in pairs(t) do
		if type(v) == "table" then
			t2[k] = main.f_tableCopy(v)
		else
			t2[k] = v
		end
	end
	return t2
end

--returns table length
function main.f_tableLength(t)
	local n = 0
	for _ in pairs(t) do
		n = n + 1
	end
	return n
end

--randomizes table content
function main.f_tableShuffle(t)
	local rand = math.random
	assert(t, "main.f_tableShuffle() expected a table, got nil")
	local iterations = #t
	local j
	for i = iterations, 2, -1 do
		j = rand(i)
		t[i], t[j] = t[j], t[i]
	end
end

--remove from table
function main.f_tableRemove(t, value)
	for k, v in pairs(t) do
		if v == value then
			table.remove(t, k)
			break
		end
	end
end

--merge 2 tables into 1 overwriting values
local function f_printValue(arg)
	if type(arg) == "table" then
		return arg[1]
	end
	return arg
end
function main.f_tableMerge(t1, t2, key)
	for k, v in pairs(t2) do
		if type(v) == "table" then
			if type(t1[k] or false) == "table" then
				main.f_tableMerge(t1[k] or {}, t2[k] or {}, k)
			elseif (t1[k] ~= nil and type(t1[k]) ~= type(v)) then
				--panicError("\n" .. (k or ''):gsub('_', '.') .. ": Incorrect data type (" .. type(t1[k]) .. " expected, got " .. type(v) .. "): " .. f_printValue(v))
				print((k or ''):gsub('_', '.') .. ": Incorrect data type (" .. type(t1[k]) .. " expected, got " .. type(v) .. "): " .. f_printValue(v))
			else
				t1[k] = v
			end
		elseif type(t1[k] or false) == "table" then
			if v ~= '' then
				t1[k][1] = v
			end
		elseif t1[k] ~= nil and type(t1[k]) ~= type(v) then
			if type(t1[k]) == "string" then
				t1[k] = tostring(v)
			else
				--panicError("\n" .. (k or ''):gsub('_', '.') .. ": Incorrect data type (" .. type(t1[k]) .. " expected, got " .. type(v) .. "): " .. f_printValue(v))
				print((k or ''):gsub('_', '.') .. ": Incorrect data type (" .. type(t1[k]) .. " expected, got " .. type(v) .. "): " .. f_printValue(v))
			end
		else
			t1[k] = v
		end
	end
	return t1
end

--returns bool if table contains value
function main.f_tableHasValue(t, val)
	for k, v in pairs(t) do
		--if v == val then
		if v:match(val) then
			return true
		end
	end
	return false
end

-- rearrange array table indexes based on index numbers stored in a second array table
function main.f_remapTable(src, remap)
	local t = {}
	for i = 1, #remap do
		table.insert(t, src[remap[i]])
	end
	return t
end

--iterate over the table in order
-- basic usage, just sort by the keys:
--for k, v in main.f_sortKeys(t) do
--	print(k, v)
--end
-- this uses an custom sorting function ordering by score descending
--for k, v in main.f_sortKeys(t, function(t, a, b) return t[b] < t[a] end) do
--	print(k, v)
--end
function main.f_sortKeys(t, order)
	-- collect the keys
	local keys = {}
	for k in pairs(t) do table.insert(keys, k) end
	-- if order function given, sort it by passing the table and keys a, b,
	-- otherwise just sort the keys
	if order then
		table.sort(keys, function(a, b) return order(t, a, b) end)
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

--round value
function main.f_round(num, places)
	if places ~= nil and places > 0 then
		local mult = 10 ^ places
		return math.floor(num * mult + 0.5) / mult
	end
	return math.floor(num + 0.5)
end

--return playerno teamside
function main.f_playerSide(pn)
	if pn % 2 ~= 0 then --odd value (Player1 side)
		return 1
	end
	return 2
end

--count occurrences of a substring
function main.f_countSubstring(s1, s2)
	return select(2, s1:gsub(s2, ""))
end

-- formats using %s / %i in the order they appear in fmt
function main.f_formatBySpec(fmt, specMap)
	if fmt == nil then return "" end
	if type(fmt) ~= "string" then fmt = tostring(fmt) end
	local args = {}
	local pos = 1
	while true do
		local p = fmt:find("%%", pos)
		if not p then break end
		local nextc = fmt:sub(p + 1, p + 1)
		if nextc == "%" then
			-- literal %%
			pos = p + 2
		else
			-- match a printf-like directive
			local _, e, verb = fmt:find("%%[%-%+ #0]*%d*%.?%d*([A-Za-z])", p)
			if not e then break end
			if verb == "s" then
				args[#args + 1] = tostring(specMap.s or "")
			elseif verb == "i" or verb == "d" or verb == "u" then
				args[#args + 1] = tonumber(specMap.i) or 0
			end
			pos = e + 1
		end
	end
	if #args == 0 then return fmt end
	local fmt2 = fmt
	local ok, out = pcall(string.format, fmt2, unpack(args))
	if not ok then
		print("main.f_formatBySpec ERROR: " .. tostring(out))
		return fmt
	end
	return out
end

--update rounds to win variables
main.roundsNumSingle = {}
main.roundsNumSimul = {}
main.roundsNumTag = {}
main.maxDrawGames = {}
function main.f_updateRoundsNum()
	for i = 1, 2 do
		if gameOption('Options.Match.Wins') == -1 then
			main.roundsNumSingle[i] = getMatchWins(i)
		else
			main.roundsNumSingle[i] = gameOption('Options.Match.Wins')
		end
		if gameOption('Options.Simul.Match.Wins') == -1 then
			main.roundsNumSimul[i] = getMatchWins(i)
		else
			main.roundsNumSimul[i] = gameOption('Options.Simul.Match.Wins')
		end
		if gameOption('Options.Tag.Match.Wins') == -1 then
			main.roundsNumTag[i] = getMatchWins(i)
		else
			main.roundsNumTag[i] = gameOption('Options.Tag.Match.Wins')
		end
		if gameOption('Options.Match.MaxDrawGames') == -2 then
			main.maxDrawGames[i] = getMatchMaxDrawGames(i)
		else
			main.maxDrawGames[i] = gameOption('Options.Match.MaxDrawGames')
		end
	end
end

--refresh screen every 0.02 during initial loading
main.nextRefresh = os.clock() + 0.02
function main.f_loadingRefresh()
	if os.clock() >= main.nextRefresh then
		textImgDraw(motif.title_info.loading.TextSpriteData)
		refresh()
		main.nextRefresh = os.clock() + 0.02
	end
end

main.pauseMenu = false
require('external.script.debug')

loadDebugFont(gameOption('Debug.Font'), gameOption('Debug.FontScale'))

main.t_stageDef = {['random'] = 0}
main.t_charDef = {}
main.t_selChars = {}
main.t_selStages = {}

--;===========================================================
--; COMMAND LINE QUICK VS
--;===========================================================
function main.f_commandLine()
	setCredits(-1)
	local ref = #main.t_selChars
	local t_teamMode = {0, 0}
	local t_numChars = {0, 0}
	local t_matchWins = {single = main.roundsNumSingle, simul = main.roundsNumSimul, tag = main.roundsNumTag, draw = main.maxDrawGames}
	local roundTime = gameOption('Options.Time')
	if getCommandLineValue("-loadmotif") == nil then
		loadLifebar()
	end
	setLifebarElements({guardbar = gameOption('Options.GuardBreak'), stunbar = gameOption('Options.Dizzy'), redlifebar = gameOption('Options.RedLife')})
	local frames = fightscreenvar("time.framespercount")
	main.f_updateRoundsNum()
	local t = {}
	local t_assignedPals = {}
	local flags = getCommandLineFlags()
	for k, v in pairs(flags) do
		if k:match('^-p[0-9]+$') then
			local num = tonumber(k:match('^-p([0-9]+)'))
			local player = main.f_playerSide(num)
			t_numChars[player] = t_numChars[player] + 1
			local pal = 1
			if flags['-p' .. num .. '.color'] ~= nil or flags['-p' .. num .. '.pal'] ~= nil then
				pal = tonumber(flags['-p' .. num .. '.color']) or tonumber(flags['-p' .. num .. '.pal'])
			elseif t_assignedPals[v] ~= nil then
				for i = 1, 12 do
					if t_assignedPals[v][i] == nil then
						pal = i
						break
					end
				end
			end
			if t_assignedPals[v] == nil then
				t_assignedPals[v] = {}
			end
			t_assignedPals[v][pal] = true
			local ai = 0
			if flags['-p' .. num .. '.ai'] ~= nil then
				ai = tonumber(flags['-p' .. num .. '.ai'])
			end
			local input = player
			if flags['-p' .. num .. '.input'] ~= nil then
				input = tonumber(flags['-p' .. num .. '.input'])
			end
			table.insert(t, {character = v, player = player, num = num, pal = pal, ai = ai, input = input, override = {}})
			if flags['-p' .. num .. '.life'] ~= nil then
				t[#t].override['life'] = tonumber(flags['-p' .. num .. '.life'])
			end
			if flags['-p' .. num .. '.lifeMax'] ~= nil then
				t[#t].override['lifemax'] = tonumber(flags['-p' .. num .. '.lifeMax'])
			end
			if flags['-p' .. num .. '.power'] ~= nil then
				t[#t].override['power'] = tonumber(flags['-p' .. num .. '.power'])
			end
			if flags['-p' .. num .. '.dizzyPoints'] ~= nil then
				t[#t].override['dizzypoints'] = tonumber(flags['-p' .. num .. '.dizzyPoints'])
			end
			if flags['-p' .. num .. '.guardPoints'] ~= nil then
				t[#t].override['guardpoints'] = tonumber(flags['-p' .. num .. '.guardPoints'])
			end
			if flags['-p' .. num .. '.lifeRatio'] ~= nil then
				t[#t].override['liferatio'] = tonumber(flags['-p' .. num .. '.lifeRatio'])
			end
			if flags['-p' .. num .. '.attackRatio'] ~= nil then
				t[#t].override['attackratio'] = tonumber(flags['-p' .. num .. '.attackRatio'])
			end
			refresh()
		elseif k:match('^-tmode1$') then
			t_teamMode[1] = tonumber(v)
		elseif k:match('^-tmode2$') then
			t_teamMode[2] = tonumber(v)
		elseif k:match('^-time$') then
			roundTime = tonumber(v)
		elseif k:match('^-rounds$') then
			for i = 1, 2 do
				t_matchWins.single[i] = tonumber(v)
				t_matchWins.simul[i] = tonumber(v)
				t_matchWins.tag[i] = tonumber(v)
			end
		elseif k:match('^-draws$') then
			for i = 1, 2 do
				t_matchWins.draw[i] = tonumber(v)
			end
		end
	end
	local t_framesMul = {1, 1}
	for i = 1, 2 do
		if t_teamMode[i] == 0 and t_numChars[i] > 1 then
			t_teamMode[i] = 1
		end
		if t_teamMode[i] == 1 then --Simul
			setMatchWins(i, t_matchWins.simul[i])
		elseif t_teamMode[i] == 3 then --Tag
			t_framesMul[i] = t_numChars[i]
			setMatchWins(i, t_matchWins.tag[i])
		else
			setMatchWins(i, t_matchWins.single[i])
		end
		setMatchMaxDrawGames(i, t_matchWins.draw[i])
	end
	frames = frames * math.max(t_framesMul[1], t_framesMul[2])
	setTimeFramesPerCount(frames)
	setRoundTime(math.max(-1, roundTime * frames))
	local stage = gameOption('Debug.StartStage')
	if flags['-s'] ~= nil then
		for _, v in ipairs({flags['-s'], 'stages/' .. flags['-s'], 'stages/' .. flags['-s'] .. '.def'}) do
			if main.f_fileExists(v) then
				stage = v
				break
			end
		end
	end
	if main.t_stageDef[stage:lower()] == nil then
		if addStage(stage) == 0 then
			panicError("\nUnable to add stage: " .. stage .. "\n")
		end
		main.t_stageDef[stage:lower()] = #main.t_selStages + 1
	end
	clearSelected()
	setMatchNo(1)
	selectStage(main.t_stageDef[stage:lower()])
	setTeamMode(1, t_teamMode[1], t_numChars[1])
	setTeamMode(2, t_teamMode[2], t_numChars[2])
	if gameOption('Debug.DumpLuaTables') then main.f_printTable(t, 'debug/t_quickvs.txt') end
	local t_params = {}
	--iterate over the table in -p order ascending
	for _, v in main.f_sortKeys(t, function(t, a, b) return t[b].num > t[a].num end) do
		if main.t_charDef[v.character:lower()] == nil then
			if flags['-loadmotif'] ~= nil then
				main.f_addChar(v.character, true, true)
			else
				addChar(v.character)
				main.t_charDef[v.character:lower()] = ref
				ref = ref + 1
			end
		end
		if main.t_charDef[v.character:lower()] == nil then
			panicError("\nUnable to add character. No such file or directory: " .. v.character .. "\n")
		end
		selectChar(v.player, main.t_charDef[v.character:lower()], v.pal)
		setCom(v.num, v.ai)
		remapInput(v.num, v.input)
		-- fold overrides into loadStart() params (p1.<member>.<field>=...)
		local member = math.ceil(v.num / 2)
		for k2, v2 in pairs(v.override) do
			table.insert(t_params, 'p' .. v.player .. '.' .. member .. '.' .. k2 .. '=' .. tostring(v2))
		end
		if start ~= nil then
			if start.p[v.player].t_selected == nil then
				start.p[v.player].t_selected = {}
			end
			table.insert(start.p[v.player].t_selected, {
				ref = main.t_charDef[v.character:lower()],
				pal = v.pal,
				pn = start.f_getPlayerNo(v.player, #start.p[v.player].t_selected + 1)
			})
		end
	end
	hook.run("main.f_commandLine")
	if flags['-ip'] ~= nil then
		enterNetPlay(flags['-ip'])
		while not connected() do
			if esc() then
				exitNetPlay()
				os.exit()
			end
			refresh()
		end
		refresh()
		synchronize()
		main.f_clearShuffleTables()
		math.randomseed(sszRandom())
		main.f_cmdBufReset()
		refresh()
	end
	local params = table.concat(t_params, ", ")
	if params == '' then
		loadStart()
	else
		loadStart(params)
	end
	while loading() do
		--do nothing
	end
	local winner = game()
	if flags['-log'] ~= nil then
		main.f_printTable(readGameStats().Matches[matchno()], flags['-log'])
	end
	os.exit()
end

--initiate quick match only if -loadmotif flag is missing
if getCommandLineValue("-p1") ~= nil and getCommandLineValue("-p2") ~= nil and getCommandLineValue("-loadmotif") == nil then
	main.f_commandLine()
end

--;===========================================================
--; LOAD DATA
--;===========================================================
main.t_unlockLua = {chars = {}, stages = {}, modes = {}}

motif = loadMotif()
if gameOption('Debug.DumpLuaTables') then main.f_printTable(motif, "debug/loadMotif.txt") end

loadLifebar()
main.f_loadingRefresh()
main.timeFramesPerCount = fightscreenvar("time.framespercount")
main.f_updateRoundsNum()

--warning display
function main.f_warning(text, sec, background, overlay, titleData, textData, cancel_snd, done_snd)
	local overlay = overlay or motif.warning_info.overlay.RectData
	local titleData = titleData or motif.warning_info.title.TextSpriteData
	local textData = textData or motif.warning_info.text.TextSpriteData
	local cancel_snd = cancel_snd or motif.warning_info.cancel.snd
	local done_snd = done_snd or motif.warning_info.done.snd
	textImgReset(textData)
	textImgSetText(textData, text)
	resetKey()
	esc(false)
	while true do
		main.f_cmdInput()
		if esc() or main.f_input(main.t_players, sec.menu.cancel.key) then
			esc(false)
			sndPlay(motif.Snd, cancel_snd[1], cancel_snd[2])
			return false
		elseif getKey() ~= '' or main.f_input(main.t_players, sec.menu.done.key) then
			sndPlay(motif.Snd, done_snd[1], done_snd[2])
			resetKey()
			return true
		end
		--draw clearcolor
		clearColor(background.bgclearcolor[1], background.bgclearcolor[2], background.bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(background.BGDef, 0)
		--draw layerno = 1 backgrounds
		bgDraw(background.BGDef, 1)
		--draw overlay
		rectDraw(overlay)
		--draw title
		textImgDraw(titleData)
		--draw text
		textImgDraw(textData)
		--end loop
		refresh()
	end
end

function main.f_drawInput(textData, text, sec, background, overlay)
	local input = ''
	resetKey()
	while true do
		if esc() or main.f_input(main.t_players, sec.menu.cancel.key) then
			input = ''
			break
		end
		if getKey('RETURN') then
			break
		elseif getKey('BACKSPACE') then
			input = input:match('^(.-).?$')
		else
			input = input .. getKeyText()
		end
		resetKey()
		--draw clearcolor
		clearColor(background.bgclearcolor[1], background.bgclearcolor[2], background.bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(background.BGDef, 0)
		--draw layerno = 1 backgrounds
		bgDraw(background.BGDef, 1)
		--draw overlay
		rectDraw(overlay)
		--draw text
		textImgReset(textData)
		textImgSetText(textData, text)
		textImgAddText(textData, '\\n\\n' .. input)
		textImgDraw(textData)
		--end loop
		main.f_cmdInput()
		refresh()
	end
	main.f_cmdInput()
	return input
end

--add characters and stages using select.def
function main.f_charParam(t, c)
	if c:match('%.[Dd][Ee][Ff]$') then --stage
		c = c:gsub('\\', '/')
		if main.f_fileExists(c) then
			if t.stage == nil then
				t.stage = {}
			end
			table.insert(t.stage, c)
		else
			print("Stage doesn't exist: " .. c)
		end
	else --param = value
		local param, value = c:match('^(.-)%s*=%s*(.-)$')
		if param ~= nil and value ~= nil and param ~= '' and value ~= '' then
			t[param] = tonumber(value)
			if t[param] == nil then
				t[param] = value
			end
		end
	end
end

function main.f_addChar(line, playable, loading, slot)
	table.insert(main.t_selChars, {})
	local row = #main.t_selChars
	local slot = slot or false
	local valid = false
	--store 'unlock' param and get rid of everything that follows it
	local unlock = ''
	line = line:gsub(',%s*unlock%s*=%s*(.-)%s*$', function(m1)
		unlock = (m1 or ''):match('^%s*(.-)%s*$')
		return ''
	end)
	--parse rest of the line
	for i, c in ipairs(main.f_strsplit(',', line)) do --split using "," delimiter
		c = c:match('^%s*(.-)%s*$')
		if i == 1 then
			if c == '' then
				playable = false
				break
			end
			c = c:gsub('\\', '/')
			c = tostring(c)
			--nClock = os.clock()
			addChar(c, line)
			--print(c .. ": " .. os.clock() - nClock)
			if c:lower() == 'skipslot' then
				main.t_selChars[row].skip = 1
				playable = false
				break
			end
			if getCharName(row - 1) == 'dummyslot' then
				playable = false
				break
			end
			main.t_charDef[c:lower()] = row - 1
			if c:lower() == 'randomselect' then
				main.t_selChars[row].char = c:lower()
				playable = false
				break
			end
			main.t_selChars[row].char = c
			valid = true
			main.t_selChars[row].playable = playable
			local t_info = getCharInfo(row - 1)
			main.t_selChars[row] = main.f_tableMerge(main.t_selChars[row], t_info)
			main.t_selChars[row].dir = main.t_selChars[row].def:gsub('[^/]+%.def$', '')
			if playable then
				for _, v in ipairs({'intro', 'ending', 'arcadepath', 'ratiopath'}) do
					if main.t_selChars[row][v] ~= '' then
						main.t_selChars[row][v] = searchFile(main.t_selChars[row][v], {main.t_selChars[row].dir, '', motif.def, 'data/'})
					end
				end
				main.t_selChars[row].order = 1
			end
		else
			main.f_charParam(main.t_selChars[row], c)
		end
	end
	if main.t_selChars[row].hidden == nil then
		main.t_selChars[row].hidden = 0
	end
	if main.t_selChars[row].char ~= nil then
		main.t_selChars[row].char_ref = main.t_charDef[main.t_selChars[row].char:lower()]
	end
	if playable then
		--order param
		if main.t_orderChars[main.t_selChars[row].order] == nil then
			main.t_orderChars[main.t_selChars[row].order] = {}
		end
		table.insert(main.t_orderChars[main.t_selChars[row].order], row - 1)
		--ordersurvival param
		local num = main.t_selChars[row].ordersurvival or 1
		if main.t_orderSurvival[num] == nil then
			main.t_orderSurvival[num] = {}
		end
		table.insert(main.t_orderSurvival[num], row - 1)
		--bonus games mode
		if main.t_selChars[row].bonus ~= nil and main.t_selChars[row].bonus == 1 then
			table.insert(main.t_bonusChars, row - 1)
		end
		--unlock
		if unlock ~= '' then
			--main.t_selChars[row].unlock = unlock
			main.t_unlockLua.chars[row] = unlock
		end
		--cell data
		local params = motif.select_info.portrait
		for _, v in pairs({{params.anim, -1}, params.spr}) do
			if v[1] ~= -1 then
				local a = animGetPreloadedCharData(main.t_selChars[row].char_ref, v[1], v[2])
				if a then
					animSetLocalcoord(a, motif.info.localcoord[1], motif.info.localcoord[2])
					animSetLayerno(a, params.layerno)
					--animSetVelocity(a, params.velocity[1], params.velocity[2])
					--animSetMaxDist(a, params.maxdist[1], params.maxdist[2])
					--animSetAccel(a, params.accel[1], params.accel[2])
					--animSetFriction(a, params.friction[1], params.friction[2])
					animSetPos(a, 0, 0)
					animSetScale(
						a,
						params.scale[1] * main.t_selChars[row].portraitscale * motif.info.localcoord[1] / main.t_selChars[row].localcoord,
						params.scale[2] * main.t_selChars[row].portraitscale * motif.info.localcoord[1] / main.t_selChars[row].localcoord
					)
					animSetFacing(a, params.facing)
					animSetXShear(a, params.xshear)
					animSetAngle(a, params.angle)
					animSetXAngle(a, params.xangle)
					animSetYAngle(a, params.yangle)
					animSetProjection(a, params.projection)
					animSetFocalLength(a, params.focallength)
					animSetWindow(a, params.window[1], params.window[2], params.window[3], params.window[4])
					animUpdate(a)
					main.t_selChars[row].cell_data = a
					break
				end
			end
		end
		if main.t_selChars[row].cell_data == nil then
			main.t_selChars[row].cell_data = animNew(nil, '-1,0, 0,0, -1')
		end
	end
	--slots
	if not slot then
		table.insert(main.t_selGrid, {['chars'] = {row}, ['slot'] = 1})
	else
		table.insert(main.t_selGrid[#main.t_selGrid].chars, row)
	end
	for _, v in ipairs({'next', 'previous', 'select'}) do
		if main.t_selChars[row][v] ~= nil then
			main.f_commandAdd(main.t_selChars[row][v], main.t_selChars[row][v])
			if main.t_selGrid[#main.t_selGrid][v] == nil then
				main.t_selGrid[#main.t_selGrid][v] = {}
			end
			if main.t_selGrid[#main.t_selGrid][v][main.t_selChars[row][v]] == nil then
				main.t_selGrid[#main.t_selGrid][v][main.t_selChars[row][v]] = {}
			end
			table.insert(main.t_selGrid[#main.t_selGrid][v][main.t_selChars[row][v]], #main.t_selGrid[#main.t_selGrid].chars)
		end
	end
	if loading then
		main.f_loadingRefresh()
	end
	return valid
end

function main.f_addStage(file, hidden, line)
	file = file:gsub('\\', '/')
	if file:match('/$') then
		return
	end
	if not addStage(file, line) then
		return
	end
	local stageNo = #main.t_selStages + 1
	local t_info = getStageInfo(stageNo)
	table.insert(main.t_selStages, {
		name = t_info.name,
		def = file,
		dir = t_info.def:gsub('[^/]+%.def$', ''),
		portraitscale = t_info.portraitscale,
		localcoord = t_info.localcoord
	})
	main.t_stageDef[file:lower()] = stageNo
	--attachedchar
	if t_info.attachedchardef ~= '' then
		local attachedList = t_info.attachedchardef
		if type(attachedList) ~= 'table' then
			attachedList = {attachedList} -- Convert string to list
		end
		main.t_selStages[stageNo].attachedChar = {}
		for i = 1, #attachedList do
			local acInfo = getCharAttachedInfo(attachedList[i])
			if acInfo ~= nil then
				acInfo.dir = acInfo.def:gsub('[^/]+%.def$', '')
				table.insert(main.t_selStages[stageNo].attachedChar, acInfo)
			end
		end
	end
	--anim data
	local function f_makeStageAnim(stageNo, params, fieldName)
		for _, v in pairs({{params.anim, -1}, params.spr}) do
			if #v > 0 and v[1] ~= -1 then
				local a = animGetPreloadedStageData(stageNo, v[1], v[2])
				if a then
					animSetLocalcoord(a, motif.info.localcoord[1], motif.info.localcoord[2])
					animSetLayerno(a, params.layerno)
					--animSetVelocity(a, params.velocity[1], params.velocity[2])
					--animSetMaxDist(a, params.maxdist[1], params.maxdist[2])
					--animSetAccel(a, params.accel[1], params.accel[2])
					--animSetFriction(a, params.friction[1], params.friction[2])
					animSetPos(a, 0, 0)
					animSetScale(
						a,
						params.scale[1] * main.t_selStages[stageNo].portraitscale * motif.info.localcoord[1] / t_info.localcoord,
						params.scale[2] * main.t_selStages[stageNo].portraitscale * motif.info.localcoord[1] / t_info.localcoord
					)
					animSetFacing(a, params.facing)
					animSetXShear(a, params.xshear)
					animSetAngle(a, params.angle)
					animSetXAngle(a, params.xangle)
					animSetYAngle(a, params.yangle)
					animSetProjection(a, params.projection)
					animSetFocalLength(a, params.focallength)
					if params.window == nil or #params.window < 4 then
						params.window = {0, 0, motif.info.localcoord[1], motif.info.localcoord[2]}
					end
					animSetWindow(
						a,
						params.window[1],
						params.window[2],
						params.window[3],
						params.window[4]
					)
					animUpdate(a)
					main.t_selStages[stageNo][fieldName] = a
					break
				end
			end
		end
	end
	--select screen anim data
	f_makeStageAnim(stageNo, motif.select_info.stage.portrait, "anim_data")
	--vs screen anim data
	f_makeStageAnim(stageNo, motif.vs_screen.stage.portrait, "vs_anim_data")
	if hidden ~= nil and hidden ~= 0 then
		main.t_selStages[stageNo].hidden = hidden
	end
	if main.t_selStages[stageNo].anim_data == nil then
		main.t_selStages[stageNo].anim_data = animNew(nil, '-1,0, 0,0, -1')
	end
	return stageNo
end

main.t_includeStage = {{}, {}} --includestage = 1, includestage = -1
main.t_orderChars = {}
main.t_orderStages = {}
main.t_orderSurvival = {}
main.t_bonusChars = {}
main.t_selGrid = {}
main.t_selOptions = {}
main.t_selStoryMode = {}
local t_storyModeList = {}
local t_addExluded = {}
local tmp = ''
local section = 0
local row = 0
local slot = false
local csCell = 0
local content = main.f_fileRead(motif.files.select)
content = content:gsub('([^\r\n;]*)%s*;[^\r\n]*', '%1')
content = content:gsub('\n%s*\n', '\n')

lanChars = false
lanStages = false
lanOptions = false
lanStory = false
for line in content:gmatch('[^\r\n]+') do
	local lineCase = line:lower()
	if lineCase:match('^%s*%[%s*' .. gameOption('Config.Language') .. '.characters' .. '%s*%]') then
		lanChars = true
	elseif lineCase:match('^%s*%[%s*' .. gameOption('Config.Language') .. '.extrastages' .. '%s*%]') then
		lanStages = true
	elseif lineCase:match('^%s*%[%s*' .. gameOption('Config.Language') .. '.options' .. '%s*%]') then
		lanOptions = true
	elseif lineCase:match('^%s*%[%s*' .. gameOption('Config.Language') .. '.storymode' .. '%s*%]') then
		lanStory = true
	end
end


for line in content:gmatch('[^\r\n]+') do
--for line in io.lines("data/select.def") do
	local lineCase = line:lower()
	if lineCase:match('^%s*%[%s*characters%s*%]') then
		row = 0
		section = 1
	elseif lineCase:match('^%s*%[%s*' .. gameOption('Config.Language') .. '.characters' .. '%s*%]') then
		if lanChars then
			row = 0
			section = 1
		else 
			section = -1
		end
	elseif lineCase:match('^%s*%[%s*extrastages%s*%]') then
		row = 0
		section = 2
	elseif lineCase:match('^%s*%[%s*' .. gameOption('Config.Language') .. '.extrastages' .. '%s*%]') then
		if lanStages then
			row = 0
			section = 2
		else 
			section = -1
		end
	elseif lineCase:match('^%s*%[%s*options%s*%]') then
		row = 0
		section = 3
	elseif lineCase:match('^%s*%[%s*' .. gameOption('Config.Language') .. '.options' .. '%s*%]') then
		if lanOptions then
			row = 0
			section = 3
		else
			section = -1
		end
	elseif lineCase:match('^%s*%[%s*storymode%s*%]') then
		row = 0
		section = 4
	elseif lineCase:match('^%s*%[%s*' .. gameOption('Config.Language') .. '.storymode' .. '%s*%]') then
		if lanStory then
			row = 0
			section = 4
		else
			section = -1
		end
	elseif lineCase:match('^%s*%[%w+%]$') then
		section = -1
	elseif section == 1 then --[Characters]
		local csCol = (csCell % motif.select_info.columns) + 1
		local csRow = math.floor(csCell / motif.select_info.columns) + 1
		local cellKey = (csCol - 1) .. '-' .. (csRow - 1)
		while not slot and motif.select_info.cell[cellKey] ~= nil and motif.select_info.cell[cellKey].skip do
			main.f_addChar('skipslot', true, true, false, line)
			csCell = csCell + 1
			csCol = (csCell % motif.select_info.columns) + 1
			csRow = math.floor(csCell / motif.select_info.columns) + 1
			cellKey = (csCol - 1) .. '-' .. (csRow - 1)
		end
		if lineCase:match(',%s*exclude%s*=%s*1') then --character should be added after all slots are filled
			table.insert(t_addExluded, line)
		elseif lineCase:match('^%s*slot%s*=%s*{%s*$') then --start of the 'multiple chars in one slot' assignment
			table.insert(main.t_selGrid, {['chars'] = {}, ['slot'] = 1})
			slot = true
		elseif slot and lineCase:match('^%s*}%s*$') then --end of 'multiple chars in one slot' assignment
			slot = false
			csCell = csCell + 1
		else
			main.f_addChar(line, true, true, slot)
			if not slot then
				csCell = csCell + 1
			end
		end
	elseif section == 2 then --[ExtraStages]
		--store 'unlock' param and get rid of everything that follows it
		local unlock = ''
		local hidden = 0 --TODO: temporary flag, won't be used once stage selection screen is ready
		line = line:gsub(',%s*unlock%s*=%s*(.-)%s*$', function(m1)
			unlock = (m1 or ''):match('^%s*(.-)%s*$')
			hidden = 1
			return ''
		end)
		--parse rest of the line
		for i, c in ipairs(main.f_strsplit(',', line)) do --split using "," delimiter
			c = c:gsub('^%s*(.-)%s*$', '%1')
			if i == 1 then
				row = main.f_addStage(c, hidden, line)
				if row == nil then
					break
				end
				table.insert(main.t_includeStage[1], row)
				table.insert(main.t_includeStage[2], row)
			else
				local param, value = c:match('^(.-)%s*=%s*(.-)$')
				if param ~= nil and value ~= nil and param ~= '' and value ~= '' then
					main.t_selStages[row][param] = tonumber(value)
					--order (more than 1 order param can be set at the same time)
					if param:match('order') then
						if main.t_orderStages[main.t_selStages[row].order] == nil then
							main.t_orderStages[main.t_selStages[row].order] = {}
						end
						table.insert(main.t_orderStages[main.t_selStages[row].order], row)
					end
				end
			end
		end
		if row ~= nil then
			--default order (only if no explicit order was set)
			if main.t_selStages[row].order == nil then
				main.t_selStages[row].order = 1
				if main.t_orderStages[main.t_selStages[row].order] == nil then
					main.t_orderStages[main.t_selStages[row].order] = {}
				end
				table.insert(main.t_orderStages[main.t_selStages[row].order], row)
			end
			--unlock param
			if unlock ~= '' then
				--main.t_selStages[row].unlock = unlock
				main.t_unlockLua.stages[row] = unlock
			end
		end
	elseif section == 3 then --[Options]
		if lineCase:match('%.maxmatches%s*=') then
			local rowName, line = lineCase:match('^%s*(.-)%.maxmatches%s*=%s*(.+)')
			rowName = rowName:gsub('%.', '_')
			main.t_selOptions[rowName .. 'maxmatches'] = {}
			for i, c in ipairs(main.f_strsplit(',', line:gsub('%s*(.-)%s*', '%1'))) do --split using "," delimiter
				main.t_selOptions[rowName .. 'maxmatches'][i] = tonumber(c)
			end
		elseif lineCase:match('%.ratiomatches%s*=') then
			local rowName, line = lineCase:match('^%s*(.-)%.ratiomatches%s*=%s*(.+)')
			rowName = rowName:gsub('%.', '_')
			main.t_selOptions[rowName .. 'ratiomatches'] = {}
			for i, c in ipairs(main.f_strsplit(',', line:gsub('%s*(.-)%s*', '%1'))) do --split using "," delimiter
				local rmin, rmax, order = c:match('^%s*([0-9]+)-?([0-9]*)%s*:%s*([0-9]+)%s*$')
				rmin = tonumber(rmin)
				rmax = tonumber(rmax) or rmin
				order = tonumber(order)
				if rmin == nil or order == nil or rmin < 1 or rmin > 4 or rmax < 1 or rmax > 4 or rmin > rmax then
					main.f_warning(motif.warning_info.text.text.ratio, motif.title_info, motif.titlebgdef)
					main.t_selOptions[rowName .. 'ratiomatches'] = nil
					break
				end
				if rmax == '' then
					rmax = rmin
				end
				table.insert(main.t_selOptions[rowName .. 'ratiomatches'], {rmin = rmin, rmax = rmax, order = order})
			end
		end
	elseif section == 4 then --[StoryMode]
		local param, value = line:match('^%s*(.-)%s*=%s*(.-)%s*$')
		if param ~= nil and value ~= nil and param ~= '' and value ~= '' then
			if param:match('^name$') then
				table.insert(main.t_selStoryMode, {name = value, displayname = '', path = '', unlock = 'true'})
				t_storyModeList[value] = true
			elseif main.t_selStoryMode[#main.t_selStoryMode][param] ~= nil then
				main.t_selStoryMode[#main.t_selStoryMode][param] = value
			end
		end
	end
end

for k, v in ipairs(main.t_selStoryMode) do
	main.t_unlockLua.modes[v.name] = v.unlock
end

--add excluded characters once all slots are filled
for i = #main.t_selGrid, motif.select_info.rows * motif.select_info.columns - 1 do
	table.insert(main.t_selChars, {})
	table.insert(main.t_selGrid, {['chars'] = {}, ['slot'] = 1})
	addChar('dummyChar')
end
for i = 1, #t_addExluded do
	main.f_addChar(t_addExluded[i], true, true)
end

--add Training char if defined and not included in select.def
if gameOption('Config.TrainingChar') ~= '' and main.t_charDef[gameOption('Config.TrainingChar'):lower()] == nil then
	main.f_addChar(gameOption('Config.TrainingChar') .. ', order = 0, ordersurvival = 0, exclude = 1', false, true)
end

--add remaining character parameters
--for each character loaded
for i = 1, #main.t_selChars do
	--character stage param
	if main.t_selChars[i].stage ~= nil then
		for j, v in ipairs(main.t_selChars[i].stage) do
			--add 'stage' param stages if needed or reference existing ones
			if main.t_stageDef[v:lower()] == nil then
				main.t_selChars[i].stage[j] = main.f_addStage(v)
				if main.t_selChars[i].includestage == nil or main.t_selChars[i].includestage == 1 then --stage available all the time
					table.insert(main.t_includeStage[1], main.t_selChars[i].stage[j])
					table.insert(main.t_includeStage[2], main.t_selChars[i].stage[j])
				elseif main.t_selChars[i].includestage == -1 then --excluded stage that can be still manually selected
					table.insert(main.t_includeStage[2], main.t_selChars[i].stage[j])
				end
			else --already added
				main.t_selChars[i].stage[j] = main.t_stageDef[v:lower()]
			end
		end
	end
end

-- (Re)build list of characters allowed to be randomly selected. Must be called after any unlock/hidden state changes.
main.t_randomChars = {}
function main.f_updateRandomChars()
	local t = {}
	for i = 1, #main.t_selChars do
		local ch = main.t_selChars[i]
		-- only real character entries
		if ch ~= nil and ch.name ~= nil then
			-- generate table with characters allowed to be randomly selected
			if ch.playable
				and (ch.hidden == nil or ch.hidden <= 1)
				and (ch.exclude == nil or ch.exclude == 0) then
				table.insert(t, i - 1)
			end
		end
	end
	main.t_randomChars = t
	if gameOption('Debug.DumpLuaTables') then main.f_printTable(main.t_randomChars, "debug/t_randomChars.txt") end
end

-- build initial pool (may be refreshed later after unlock() runs)
main.f_updateRandomChars()

--add default starting stage if no stages have been added via select.def
if #main.t_includeStage[1] == 0 or #main.t_includeStage[2] == 0 then
	local row = main.f_addStage(gameOption('Debug.StartStage'))
	table.insert(main.t_includeStage[1], row)
	table.insert(main.t_includeStage[2], row)
end

--update selectableStages table
function main.f_updateSelectableStages()
	main.t_selectableStages = {}
	for _, v in ipairs(main.t_includeStage[2]) do
		if main.t_selStages[v].hidden == nil or main.t_selStages[v].hidden == 0 then
			table.insert(main.t_selectableStages, v)
		end
	end
end
main.f_updateSelectableStages()

--add default maxmatches / ratiomatches values if config is missing in select.def
if main.t_selOptions.arcademaxmatches == nil then main.t_selOptions.arcademaxmatches = {6, 1, 1, 0, 0, 0, 0, 0, 0, 0} end
if main.t_selOptions.teammaxmatches == nil then main.t_selOptions.teammaxmatches = {4, 1, 1, 0, 0, 0, 0, 0, 0, 0} end
if main.t_selOptions.timeattackmaxmatches == nil then main.t_selOptions.timeattackmaxmatches = {6, 1, 1, 0, 0, 0, 0, 0, 0, 0} end
if main.t_selOptions.survivalmaxmatches == nil then main.t_selOptions.survivalmaxmatches = {-1, 0, 0, 0, 0, 0, 0, 0, 0, 0} end
if main.t_selOptions.arcaderatiomatches == nil then
	main.t_selOptions.arcaderatiomatches = {
		{rmin = 1, rmax = 3, order = 1},
		{rmin = 3, rmax = 3, order = 1},
		{rmin = 2, rmax = 2, order = 1},
		{rmin = 2, rmax = 2, order = 1},
		{rmin = 1, rmax = 1, order = 2},
		{rmin = 3, rmax = 3, order = 1},
		{rmin = 1, rmax = 2, order = 3},
	}
end

--uppercase title
function main.f_itemnameUpper(title, uppercase)
	if title == nil then
		return ''
	end
	if uppercase then
		return title:upper()
	end
	return title
end

--returns table storing menu window coordinates
function main.f_menuWindow(t, offset)
	local offset = offset or {0, 0}
	-- If margins are set, keep legacy vertical-only clamp.
	if t.window.margins.y[1] ~= 0 or t.window.margins.y[2] ~= 0 then
		return {
			0,
			math.max(0, t.pos[2] + offset[2] - t.window.margins.y[1]),
			motif.info.localcoord[1],
			t.pos[2] + offset[2] + (t.window.visibleitems - 1) * t.item.spacing[2] + t.window.margins.y[2]
		}
	end
	-- Margins 0,0 => clamp tightly to the menu box (both axes).
	local x1 = t.pos[1] + offset[1] + t.boxcursor.coords[1]
	local y1 = t.pos[2] + offset[2] + t.boxcursor.coords[2]
	local w  = t.boxcursor.coords[3] - t.boxcursor.coords[1] + 1
	local h  = t.boxcursor.coords[4] - t.boxcursor.coords[2] + 1
	-- Height grows with visible rows; using visibleitems is enough for clipping.
	local winLeft   = math.max(0, x1)
	local winTop    = math.max(0, y1)
	local winRight  = math.min(x1 + w, motif.info.localcoord[1])
	local winBottom = math.min(y1 + h + (t.window.visibleitems - 1) * t.item.spacing[2], motif.info.localcoord[2])
	return {winLeft, winTop, winRight, winBottom}
end

function main.f_storyboard(path)
	local s = loadStoryboard(path)
	if s == nil then
		return
	end
	if gameOption('Debug.DumpLuaTables') then
		-- get filename without extension from full path
		local name = path:match("([^/\\]+)$") or "unknown" -- last path segment
		name = name:gsub("%.[^%.]+$", "") -- strip last extension
		main.f_printTable(s, 'debug/loadStoryboard_' .. name .. '.txt')
	end
	while true do
		if not runStoryboard() then
			break
		end
		refresh()
	end
end

function main.f_hiscore(mode, place)
	while true do
		if not runHiscore(mode, place) then
			break
		end
		main.f_cmdInput()
		refresh()
	end
end

--Load additional scripts
start = require('external.script.start')
options = require('external.script.options')
menu = require('external.script.menu')

if getCommandLineValue("-storyboard") ~= nil then
	main.f_storyboard(getCommandLineValue("-storyboard"))
	os.exit()
end

--;===========================================================
--; MENUS
--;===========================================================
if motif.attract_mode.enabled then
	main.group = 'attract_mode'
	main.background = 'attractbgdef'
else
	main.group = 'title_info'
	main.background = 'titlebgdef'
end

function main.f_default()
	for i = 1, gameOption('Config.Players') do
		main.t_pIn[i] = i
		main.t_remaps[i] = i
	end
	main.aiRamp = false --if AI ramping should be active
	main.charparam = { --which select.def charparam should be used
		ai = false,
		arcadepath = false,
		music = false,
		rounds = false,
		single = false,
		stage = false,
		time = false,
	}
	main.coop = false --if mode should be recognized as coop
	main.cpuSide = {false, true} --which side is controlled by CPU
	main.dropDefeated = false --if defeated members should be removed from team
	main.elimination = false --if single lose should stop further lua execution
	main.exitSelect = false --if "clearing" the mode (matchno == -1) should go back to main menu
	main.forceChar = {nil, nil} --predefined P1/P2 characters
	main.forceRosterSize = false --if roster size should be enforced even if there are not enough characters to fill it (not used but may be useful for external modules)
	main.lifebar = { --which lifebar elements should be rendered (these defaults are overwritten by fight.def, depending on game mode)
		active = true,
		bars = true,
		match = false,
		mode = true,
		p1ailevel = false,
		p1score = false,
		p1wincount = false,
		p2ailevel = false,
		p2score = false,
		p2wincount = false,
		timer = false,
		guardbar = gameOption('Options.GuardBreak'),
		stunbar = gameOption('Options.Dizzy'),
		redlifebar = gameOption('Options.RedLife'),
	}
	main.luaPath = 'external/script/default.lua' --path to script executed by start.f_selectMode()
	main.makeRoster = false --if default roster for each match should be generated before first match
	main.matchWins = { --amount of rounds to win for each team side and team mode
		draw = main.maxDrawGames,
		simul = main.roundsNumSimul,
		single = main.roundsNumSingle,
		tag = main.roundsNumTag,
	}
	main.motif = { --which motif elements should be rendered
		challenger = false,
		continuescreen = false,
		demo = true,
		dialogue = true,
		hiscore = false,
		versusscreen = false,
		versusmatchno = false,
		victoryscreen = false,
		winscreen = false,
		losescreen = false,
		menu = true,
	}
	main.numSimul = {gameOption('Options.Simul.Min'), gameOption('Options.Simul.Max')} --min/max number of simul characters
	main.numTag = {gameOption('Options.Tag.Min'), gameOption('Options.Tag.Max')} --min/max number of tag characters
	main.numTurns = {gameOption('Options.Turns.Min'), gameOption('Options.Turns.Max')} --min/max number of turn characters
	main.orderSelect = {false, false} --if versus screen order selection should be active
	main.persistLife = false --if life should be maintained after match
	main.persistMusic = false --if the music that was playing previously should be stopped at the start of the match
	main.persistRounds = false --if lifebar should use consecutive wins for round numbers
	main.quickContinue = false --if by default continuing should skip player selection
	main.rankingCondition = false --if winning (clearing) whole mode is needed for rankings to be saved
	main.resetScore = false --if loosing should set score for the next match to lose count
	main.rotationChars = false --flags modes where gameOption('Arcade.AI.SurvivalColor') should be used instead of gameOption('Arcade.AI.RandomColor')
	main.roundTime = gameOption('Options.Time') --sets round time
	main.selectMenu = {true, false} --which team side should be allowed to select players
	main.stageMenu = false --if manual stage selection is allowed
	main.stageOrder = false --if select.def stage order param should be used
	main.storyboard = {intro = false, ending = false, credits = false, gameover = false} --which storyboards should be active
	main.teamMenu = {
		{ratio = false, simul = false, single = false, tag = false, turns = false}, --which team modes should be selectable by P1 side
		{ratio = false, simul = false, single = false, tag = false, turns = false}, --which team modes should be selectable by P2 side
	}
	resetAILevel()
	resetRemapInput()
	if not motif.attract_mode.enabled and start.challenger == 0 then
		setCredits(-1) --amount of credits from the start (-1 = disabled)
	end
	setConsecutiveWins(1, 0)
	setConsecutiveWins(2, 0)
	setGameMode('')
	setHomeTeam(2) --http://mugenguild.com/forum/topics/ishometeam-triggers-169132.0.html
	setLifebarElements(main.lifebar)
	setMotifElements(main.motif)
	setRoundTime(math.max(-1, main.roundTime * main.timeFramesPerCount))
	setTimeFramesPerCount(main.timeFramesPerCount)
	setWinCount(1, 0)
	setWinCount(2, 0)
	textImgReset(motif.select_info.title.TextSpriteData)
	main.f_cmdBufReset()
	demoFrameCounter = 0
	hook.run("main.f_default")
end

-- Associative elements table storing functions controlling behaviour of each
-- menu item (modes configuration). Can be appended via external module.
main.t_itemname = {
	--ARCADE / TEAM ARCADE
	['arcade'] = function(t, item)
		main.f_playerInput(main.playerInput, 1)
		main.t_pIn[2] = 1
		main.aiRamp = true
		main.charparam.ai = true
		main.charparam.arcadepath = true
		main.charparam.music = true
		main.charparam.rounds = true
		main.charparam.single = true
		main.charparam.stage = true
		main.charparam.time = true
		main.exitSelect = true
		--main.lifebar.p1score = true
		--main.lifebar.p2ailevel = true
		main.makeRoster = true
		main.motif.challenger = true
		main.motif.continuescreen = true
		main.motif.hiscore = true
		main.motif.versusscreen = true
		main.motif.versusmatchno = true
		main.motif.victoryscreen = true
		main.motif.winscreen = true
		main.orderSelect[1] = true
		main.orderSelect[2] = true
		main.resetScore = true
		main.stageOrder = true
		main.storyboard.credits = true
		main.storyboard.ending = true
		main.storyboard.gameover = true
		main.storyboard.intro = true
		if (t ~= nil and t[item].itemname == 'arcade') or (t == nil and not main.teamarcade) then
			main.teamMenu[1].single = true
			main.teamMenu[2].single = true
			textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.arcade)
			main.teamarcade = false
		else --teamarcade
			main.teamMenu[1].ratio = true
			main.teamMenu[1].simul = true
			main.teamMenu[1].single = true
			main.teamMenu[1].tag = true
			main.teamMenu[1].turns = true
			main.teamMenu[2].ratio = true
			main.teamMenu[2].simul = true
			main.teamMenu[2].single = true
			main.teamMenu[2].tag = true
			main.teamMenu[2].turns = true
			textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.teamarcade)
			main.teamarcade = true
			
		end
		main.f_setCredits()
		setGameMode('arcade')
		hook.run("main.t_itemname")
		if start.challenger == 0 then
			return start.f_selectMode
		end
		return nil
	end,
	--BONUS CHAR
	['bonus'] = function(t, item)
		main.f_playerInput(main.playerInput, 1)
		main.charparam.ai = true
		main.charparam.music = true
		main.charparam.rounds = true
		main.charparam.single = true
		main.charparam.stage = true
		main.charparam.time = true
		main.forceChar[2] = {main.t_bonusChars[item]}
		main.selectMenu[2] = true
		main.teamMenu[1].single = true
		main.teamMenu[2].single = true
		textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.bonus)
		setGameMode('bonus')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--DEMO
	['demo'] = function()
		return main.f_demoStart
	end,
	--FREE BATTLE (QUICK VS)
	['freebattle'] = function()
		main.f_playerInput(main.playerInput, 1)
		main.t_pIn[2] = 1
		--main.lifebar.p1score = true
		--main.lifebar.p2ailevel = true
		main.motif.versusscreen = true
		main.motif.victoryscreen = true
		main.orderSelect[1] = true
		main.orderSelect[2] = true
		main.selectMenu[2] = true
		main.stageMenu = true
		main.teamMenu[1].ratio = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].single = true
		main.teamMenu[1].tag = true
		main.teamMenu[1].turns = true
		main.teamMenu[2].ratio = true
		main.teamMenu[2].simul = true
		main.teamMenu[2].single = true
		main.teamMenu[2].tag = true
		main.teamMenu[2].turns = true
		textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.freebattle)
		setGameMode('freebattle')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--JOIN (NEW ADDRESS)
	['joinadd'] = function(t, item)
		sndPlay(motif.Snd, motif[main.group].cursor.move.snd[1], motif[main.group].cursor.move.snd[2])
		local name = main.f_drawInput(
			motif[main.group].textinput.TextSpriteData,
			motif[main.group].textinput.text.name,
			motif[main.group],
			motif[main.background],
			motif[main.group].textinput.overlay.RectData
		)
		if name ~= '' then
			sndPlay(motif.Snd, motif[main.group].cursor.move.snd[1], motif[main.group].cursor.move.snd[2])
			local address = main.f_drawInput(
				motif[main.group].textinput.TextSpriteData,
				motif[main.group].textinput.text.address,
				motif[main.group],
				motif[main.background],
				motif[main.group].textinput.overlay.RectData
			)
			if address:match('^[0-9%.]+$') then
				sndPlay(motif.Snd, motif[main.group].cursor.done.snd[1], motif[main.group].cursor.done.snd[2])
				modifyGameOption('Netplay.IP.' .. name, address)
				table.insert(t, #t, {itemname = 'ip_' .. name, displayname = name})
				saveGameOption(getCommandLineValue("-config"))
			else
				sndPlay(motif.Snd, motif[main.group].cancel.snd[1], motif[main.group].cancel.snd[2])
			end
		else
			sndPlay(motif.Snd, motif[main.group].cancel.snd[1], motif[main.group].cancel.snd[2])
		end
		main.f_cmdBufReset()
		return t
	end,
	--NETPLAY SURVIVAL
	['netplaysurvivalcoop'] = function()
		main.aiRamp = true
		main.charparam.ai = true
		main.charparam.music = true
		main.charparam.single = true
		main.charparam.stage = true
		main.charparam.time = true
		main.coop = true
		main.elimination = true
		main.exitSelect = true
		--main.lifebar.match = true
		--main.lifebar.p2ailevel = true
		main.makeRoster = true
		main.motif.losescreen = true
		main.motif.winscreen = true
		main.matchWins.draw = {0, 0}
		main.matchWins.simul = {1, 1}
		main.matchWins.single = {1, 1}
		main.matchWins.tag = {1, 1}
		main.numSimul = {2, 2}
		main.numTag = {2, 2}
		main.persistLife = true
		main.persistMusic = true
		main.persistRounds = true
		main.stageMenu = true
		main.storyboard.credits = true
		main.storyboard.gameover = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].tag = true
		main.teamMenu[2].ratio = true
		main.teamMenu[2].simul = true
		main.teamMenu[2].single = true
		main.teamMenu[2].tag = true
		main.teamMenu[2].turns = true
		textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.netplaysurvivalcoop)
		setGameMode('netplaysurvivalcoop')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--NETPLAY CO-OP
	['netplayteamcoop'] = function()
		main.aiRamp = true
		main.charparam.ai = true
		main.charparam.arcadepath = true
		main.charparam.music = true
		main.charparam.rounds = true
		main.charparam.single = true
		main.charparam.stage = true
		main.charparam.time = true
		main.coop = true
		main.exitSelect = true
		--main.lifebar.p1score = true
		--main.lifebar.p2ailevel = true
		main.makeRoster = true
		main.motif.continuescreen = true
		main.motif.versusscreen = true
		main.motif.versusmatchno = true
		main.motif.victoryscreen = true
		main.motif.winscreen = true
		main.numSimul = {2, 2}
		main.numTag = {2, 2}
		main.resetScore = true
		main.stageOrder = true
		main.storyboard.credits = true
		main.storyboard.ending = true
		main.storyboard.gameover = true
		main.storyboard.intro = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].tag = true
		main.teamMenu[2].ratio = true
		main.teamMenu[2].simul = true
		main.teamMenu[2].single = true
		main.teamMenu[2].tag = true
		main.teamMenu[2].turns = true
		main.f_setCredits()
		textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.netplayteamcoop)
		setGameMode('netplayteamcoop')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--NETPLAY VERSUS
	['netplayversus'] = function()
		setHomeTeam(1)
		main.cpuSide[2] = false
		--main.lifebar.p1wincount = true
		--main.lifebar.p2wincount = true
		main.motif.versusscreen = true
		main.motif.victoryscreen = true
		main.orderSelect[1] = true
		main.orderSelect[2] = true
		main.selectMenu[2] = true
		main.stageMenu = true
		main.teamMenu[1].ratio = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].single = true
		main.teamMenu[1].tag = true
		main.teamMenu[1].turns = true
		main.teamMenu[2].ratio = true
		main.teamMenu[2].simul = true
		main.teamMenu[2].single = true
		main.teamMenu[2].tag = true
		main.teamMenu[2].turns = true
		textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.netplayversus)
		setGameMode('netplayversus')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--OPTIONS
	['options'] = function()
		hook.run("main.t_itemname")
		return options.menu.loop
	end,
	--REPLAY
	['replay'] = function()
		return main.f_replay
	end,
	--SERVER CONNECT
	['serverconnect'] = function(t, item)
		sndPlay(motif.Snd, motif[main.group].cursor.done.snd[1], motif[main.group].cursor.done.snd[2]) -- Needs manual sndPlay due to special menu behavior
		if main.f_connect(gameOption('Netplay.IP.' .. t[item].displayname), t[item].displayname) then
			synchronize()
			main.f_clearShuffleTables()
			math.randomseed(sszRandom())
			main.f_cmdBufReset()
			main.menu.submenu.server.loop()
			replayStop()
			exitNetPlay()
			exitReplay()
		end
		return nil
	end,
	--SERVER HOST
	['serverhost'] = function(t, item)
		sndPlay(motif.Snd, motif[main.group].cursor.done.snd[1], motif[main.group].cursor.done.snd[2]) -- Needs manual sndPlay due to special menu behavior
		if main.f_connect("", gameOption('Netplay.ListenPort')) then
			synchronize()
			main.f_clearShuffleTables()
			math.randomseed(sszRandom())
			main.f_cmdBufReset()
			main.menu.submenu.server.loop()
			replayStop()
			exitNetPlay()
			exitReplay()
		end
		return nil
	end,
	--STORY MODE ARC
	['storyarc'] = function(t, item)
		main.f_playerInput(main.playerInput, 1)
		main.motif.continuescreen = true
		main.selectMenu[1] = false
		for _, v in ipairs(main.t_selStoryMode) do
			if v.name == t[item].itemname then
				main.luaPath = v.path
				break
			end
		end
		setGameMode(t[item].itemname)
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--SURVIVAL
	['survival'] = function()
		main.f_playerInput(main.playerInput, 1)
		main.t_pIn[2] = 1
		main.aiRamp = true
		main.charparam.ai = true
		main.charparam.music = true
		main.charparam.single = true
		main.charparam.stage = true
		main.charparam.time = true
		main.dropDefeated = true
		main.elimination = true
		main.exitSelect = true
		--main.lifebar.match = true
		--main.lifebar.p2ailevel = true
		main.makeRoster = true
		main.motif.hiscore = true
		main.motif.losescreen = true
		main.motif.winscreen = true
		main.matchWins.draw = {0, 0}
		main.matchWins.simul = {1, 1}
		main.matchWins.single = {1, 1}
		main.matchWins.tag = {1, 1}
		main.orderSelect[1] = true
		main.orderSelect[2] = true
		main.persistLife = true
		main.persistMusic = true
		main.persistRounds = true
		main.rotationChars = true
		main.stageMenu = true
		main.storyboard.credits = true
		main.storyboard.gameover = true
		main.teamMenu[1].ratio = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].single = true
		main.teamMenu[1].tag = true
		main.teamMenu[1].turns = true
		main.teamMenu[2].ratio = true
		main.teamMenu[2].simul = true
		main.teamMenu[2].single = true
		main.teamMenu[2].tag = true
		main.teamMenu[2].turns = true
		textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.survival)
		setGameMode('survival')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--SURVIVAL CO-OP
	['survivalcoop'] = function()
		main.aiRamp = true
		main.charparam.ai = true
		main.charparam.music = true
		main.charparam.single = true
		main.charparam.stage = true
		main.charparam.time = true
		main.coop = true
		main.elimination = true
		main.exitSelect = true
		--main.lifebar.match = true
		--main.lifebar.p2ailevel = true
		main.makeRoster = true
		main.motif.hiscore = true
		main.motif.winscreen = true
		main.matchWins.draw = {0, 0}
		main.matchWins.simul = {1, 1}
		main.matchWins.single = {1, 1}
		main.matchWins.tag = {1, 1}
		main.numSimul = {2, math.min(4, gameOption('Config.Players'))}
		main.numTag = {2, math.min(4, gameOption('Config.Players'))}
		main.persistLife = true
		main.persistMusic = true
		main.persistRounds = true
		main.rotationChars = true
		main.stageMenu = true
		main.storyboard.credits = true
		main.storyboard.gameover = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].tag = true
		main.teamMenu[2].ratio = true
		main.teamMenu[2].simul = true
		main.teamMenu[2].single = true
		main.teamMenu[2].tag = true
		main.teamMenu[2].turns = true
		textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.survivalcoop)
		setGameMode('survivalcoop')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--TEAM CO-OP
	['teamcoop'] = function()
		main.aiRamp = true
		main.charparam.ai = true
		main.charparam.arcadepath = true
		main.charparam.music = true
		main.charparam.rounds = true
		main.charparam.single = true
		main.charparam.stage = true
		main.charparam.time = true
		main.coop = true
		main.exitSelect = true
		--main.lifebar.p1score = true
		--main.lifebar.p2ailevel = true
		main.makeRoster = true
		main.motif.continuescreen = true
		main.motif.hiscore = true
		main.motif.versusscreen = true
		main.motif.versusmatchno = true
		main.motif.victoryscreen = true
		main.motif.winscreen = true
		main.numSimul = {2, math.min(4, gameOption('Config.Players'))}
		main.numTag = {2, math.min(4, gameOption('Config.Players'))}
		main.resetScore = true
		main.stageOrder = true
		main.storyboard.credits = true
		main.storyboard.ending = true
		main.storyboard.gameover = true
		main.storyboard.intro = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].tag = true
		main.teamMenu[2].ratio = true
		main.teamMenu[2].simul = true
		main.teamMenu[2].single = true
		main.teamMenu[2].tag = true
		main.teamMenu[2].turns = true
		main.f_setCredits()
		textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.teamcoop)
		setGameMode('teamcoop')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--TIME ATTACK
	['timeattack'] = function()
		main.f_playerInput(main.playerInput, 1)
		main.t_pIn[2] = 1
		main.aiRamp = true
		main.charparam.ai = true
		main.charparam.music = true
		main.charparam.rounds = true
		main.charparam.single = true
		main.charparam.stage = true
		main.charparam.time = true
		main.exitSelect = true
		--main.lifebar.p2ailevel = true
		--main.lifebar.timer = true
		main.makeRoster = true
		main.motif.continuescreen = true
		main.motif.hiscore = true
		main.motif.versusscreen = true
		main.motif.versusmatchno = true
		main.motif.winscreen = true
		main.quickContinue = true
		main.orderSelect[1] = true
		main.orderSelect[2] = true
		main.resetScore = true
		if main.roundTime == -1 then
			main.roundTime = 99
		end
		main.stageOrder = true
		main.storyboard.credits = true
		main.storyboard.gameover = true
		main.teamMenu[1].ratio = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].single = true
		main.teamMenu[1].tag = true
		main.teamMenu[1].turns = true
		main.teamMenu[2].ratio = true
		main.teamMenu[2].simul = true
		main.teamMenu[2].single = true
		main.teamMenu[2].tag = true
		main.teamMenu[2].turns = true
		main.f_setCredits()
		textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.timeattack)
		setGameMode('timeattack')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--TRAINING
	['training'] = function()
		setHomeTeam(1)
		main.f_playerInput(main.playerInput, 1)
		main.t_pIn[2] = 1
		if main.t_charDef[gameOption('Config.TrainingChar'):lower()] ~= nil then
			main.forceChar[2] = {main.t_charDef[gameOption('Config.TrainingChar'):lower()]}
		end
		--main.lifebar.p1score = true
		--main.lifebar.p2ailevel = true
		main.roundTime = -1
		main.selectMenu[2] = true
		if gameOption('Config.TrainingStage') == '' then
			main.stageMenu = true
		end
		main.teamMenu[1].ratio = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].single = true
		main.teamMenu[1].tag = true
		main.teamMenu[1].turns = true
		main.teamMenu[2].single = true
		main.matchWins.draw = {0, 0}
		main.matchWins.simul = {0, 0}
		main.matchWins.single = {0, 0}
		main.matchWins.tag = {0, 0}
		textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.training)
		setGameMode('training')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--TRIALS
	['trials'] = function()
	end,
	--VS MODE / TEAM VERSUS
	['versus'] = function(t, item)
		setHomeTeam(1)
		if start.challenger > 0 then
			main.t_pIn[2] = start.challenger
		end
		main.cpuSide[2] = false
		--main.lifebar.p1wincount = true
		--main.lifebar.p2wincount = true
		main.motif.versusscreen = true
		main.motif.victoryscreen = true
		main.orderSelect[1] = true
		main.orderSelect[2] = true
		main.selectMenu[2] = true
		main.stageMenu = true
		if (start.challenger == 0 and t[item].itemname == 'versus') or (start.challenger ~= 0 and not main.teamarcade) then
			main.teamMenu[1].single = true
			main.teamMenu[2].single = true
			textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.versus)
		else --teamversus
			main.teamMenu[1].ratio = true
			main.teamMenu[1].simul = true
			main.teamMenu[1].single = true
			main.teamMenu[1].tag = true
			main.teamMenu[1].turns = true
			main.teamMenu[2].ratio = true
			main.teamMenu[2].simul = true
			main.teamMenu[2].single = true
			main.teamMenu[2].tag = true
			main.teamMenu[2].turns = true
			textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.teamversus)
		end
		if start.challenger > 0 then
			setGameMode('challenger')
		else
			setGameMode('versus')
		end
		hook.run("main.t_itemname")
		if start.challenger == 0 then
			return start.f_selectMode
		end
		return nil
	end,
	--VERSUS CO-OP
	['versuscoop'] = function()
		setHomeTeam(1)
		main.coop = true
		main.cpuSide[2] = false
		--main.lifebar.p1wincount = true
		--main.lifebar.p2wincount = true
		main.motif.versusscreen = true
		main.motif.victoryscreen = true
		main.numSimul = {2, math.min(4, math.max(2, math.ceil(gameOption('Config.Players') / 2)))}
		main.numTag = {2, math.min(4, math.max(2, math.ceil(gameOption('Config.Players') / 2)))}
		main.selectMenu[2] = true
		main.stageMenu = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].tag = true
		main.teamMenu[2].simul = true
		main.teamMenu[2].tag = true
		textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.versuscoop)
		setGameMode('versuscoop')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--WATCH
	['watch'] = function()
		setHomeTeam(1)
		main.f_playerInput(main.playerInput, 1)
		main.t_pIn[2] = 1
		main.cpuSide[1] = true
		--main.lifebar.p1ailevel = true
		--main.lifebar.p2ailevel = true
		main.motif.versusscreen = true
		main.motif.victoryscreen = true
		main.selectMenu[2] = true
		main.stageMenu = true
		main.teamMenu[1].ratio = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].single = true
		main.teamMenu[1].tag = true
		main.teamMenu[1].turns = true
		main.teamMenu[2].ratio = true
		main.teamMenu[2].simul = true
		main.teamMenu[2].single = true
		main.teamMenu[2].tag = true
		main.teamMenu[2].turns = true
		textImgSetText(motif.select_info.title.TextSpriteData, motif.select_info.title.text.watch)
		setGameMode('watch')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
}
main.t_itemname.teamarcade = main.t_itemname.arcade
main.t_itemname.teamversus = main.t_itemname.versus
if gameOption('Debug.DumpLuaTables') then main.f_printTable(main.t_itemname, 'debug/t_mainItemname.txt') end

function main.f_deleteIP(item, t)
	if t[item].itemname:match('^ip_') then
		sndPlay(motif.Snd, motif.title_info.cancel.snd[1], motif.title_info.cancel.snd[2])
		resetKey()
		modifyGameOption('Netplay.IP.' .. t[item].itemname:gsub('^ip_', ''), nil)
		saveGameOption(getCommandLineValue("-config"))
		for i = 1, #t do
			if t[i].itemname == t[item].itemname then
				table.remove(t, i)
				break
			end
		end
	end
	return t
end

--return table without hidden modes (present in main.t_unlockLua.modes table)
function main.f_hiddenItems(t_items)
	local t = {}
	for _, v in ipairs(t_items) do
		if main.t_unlockLua.modes[v.itemname] == nil then
			table.insert(t, v)
		end
	end
	return t
end

local demoFrameCounter = 0
local introWaitCycles = 0

-- Returns the action function name and visible items table if the given menu table has exactly one actionable entry.
function main.f_getSingleMenuAction(tbl)
	if tbl == nil or tbl.items == nil then
		return nil
	end
	local cnt = 0
	local f = nil
	for _, v in ipairs(tbl.items) do
		if tbl.name == 'bonusgames' --[[or tbl.name == 'storymode']] or v.itemname == 'joinadd' then
			return nil
		elseif v.itemname ~= 'back' and main.t_unlockLua.modes[v.itemname] == nil then
			local fname = v.itemname
			if main.t_itemname[fname] == nil then
				if t_storyModeList[fname] then
					fname = 'storyarc'
				elseif fname:match('^bonus_') then
					fname = 'bonus'
				elseif fname:match('^ip_') then
					fname = 'serverconnect'
				end
			end
			cnt = cnt + 1
			f = fname
			if cnt > 1 then
				return nil
			end
		end
	end
	if f ~= nil and main.t_itemname[f] ~= nil and cnt == 1 then
		local t = main.f_hiddenItems(tbl.items)
		return f, t
	end
	return nil
end

-- Shared menu loop logic
function main.f_createMenu(tbl, bool_bgreset, bool_main, bool_f1, bool_del)
	return function()
		hook.run("main.menu.loop")
		local cursorPosY = 1
		local moveTxt = 0
		local item = 1
		local t = main.f_hiddenItems(tbl.items)
		local call_t_override = nil
		local call_item_override = nil
		--skip showing menu if there is only 1 valid item
		main.f_menuSnap(motif[main.group])
		-- Only auto-run here for menus that are entered directly (no parent submenu), so the fadeout is handled in the caller for submenus.
		local single_f, single_t = main.f_getSingleMenuAction(tbl)
		if single_f ~= nil and bool_bgreset then
			main.f_default()
			main.menu.f = main.t_itemname[single_f](single_t, item)
			main.f_unlock(false)
			if main.menu.f ~= nil then
				main.menu.f()
			end
			main.f_default()
			main.f_unlock(false)
			local itemNum = #t
			t = main.f_hiddenItems(tbl.items)
			main.menu.f = nil
			if itemNum == #t then
				return
			end
		end
		--more than 1 item (or newly unlocked ones), continue loop
		if bool_main then
			if motif.files.logo.storyboard ~= '' then
				main.f_storyboard(motif.files.logo.storyboard)
			end
			if motif.files.intro.storyboard ~= '' then
				main.f_storyboard(motif.files.intro.storyboard)
			end
		end
		if bool_bgreset then
			if not motif.attract_mode.enabled then
				bgReset(motif[main.background].BGDef)
				playBgm({source = "motif.title", interrupt = true})
			end
			main.f_fadeReset('fadein', motif[main.group])
		end
		main.menu.f = nil
		while true do
			if tbl.reset then
				tbl.reset = false
				main.f_cmdInput()
			else
				main.f_menuCommonDraw(t, item, cursorPosY, moveTxt, motif[main.group], motif[main.background], false)
			end
			if main.menu.f ~= nil and not main.fadeActive then
				main.f_unlock(false)
				main.menu.f()
				main.f_default()
				main.f_unlock(false)
				t = main.f_hiddenItems(tbl.items)
				main.menu.f = nil
			else
				if bool_main then
					main.f_demo()
				end
				local item_sav = item
				cursorPosY, moveTxt, item = main.f_menuCommonCalc(t, item, cursorPosY, moveTxt, motif[main.group], motif[main.group].cursor)
				textImgSetText(motif[main.group].title.TextSpriteData, tbl.title)
				if item_sav ~= item then
					demoFrameCounter = 0
					introWaitCycles = 0
				end
				if esc() or main.f_input(main.t_players, motif[main.group].menu.cancel.key) then
					if not bool_main then
						sndPlay(motif.Snd, motif[main.group].cancel.snd[1], motif[main.group].cancel.snd[2])
					elseif not esc() and t[item].itemname ~= 'exit' then
						--menu key moves cursor to exit without exiting the game
						for i = 1, #t do
							if t[i].itemname == 'exit' then
								sndPlay(motif.Snd, motif[main.group].cancel.snd[1], motif[main.group].cancel.snd[2])
								item = i
								cursorPosY = math.min(item, motif[main.group].menu.window.visibleitems)
								if cursorPosY >= motif[main.group].menu.window.visibleitems then
									moveTxt = (item - motif[main.group].menu.window.visibleitems) * motif[main.group].menu.item.spacing[2]
								end
								break
							end
						end
					end
					if not bool_main or esc() then
						break
					end
				elseif bool_f1 and (getKey('F1') or gameOption('Config.FirstRun')) then
					if gameOption('Config.FirstRun') then
						modifyGameOption('Config.FirstRun', false)
						options.f_saveCfg(false)
					end
					main.f_warning(
						motif.infobox.text.text,
						motif[main.group],
						motif[main.background],
						motif.infobox.overlay.RectData,
						motif.infobox.title.TextSpriteData,
						motif.infobox.text.TextSpriteData
					)
					main.f_cmdBufReset()
				elseif motif.attract_mode.enabled and getKey(motif.attract_mode.options.keycode) then
					main.f_default()
					main.menu.f = main.t_itemname.options()
					sndPlay(motif.Snd, motif[main.group].cursor.done.snd[1], motif[main.group].cursor.done.snd[2])
					main.f_fadeReset('fadeout', motif[main.group])
					resetKey()
				elseif bool_del and getKey('DELETE') then
					tbl.items = main.f_deleteIP(item, t)
				elseif main.f_input(main.t_players, motif[main.group].menu.hiscore.key) and main.f_hiscoreDisplay(t[item].itemname) then
					demoFrameCounter = 0
				elseif main.f_input(main.t_players, motif[main.group].menu.done.key) then
					demoFrameCounter = 0
					local f = t[item].itemname
					if f == 'back' then
						sndPlay(motif.Snd, motif[main.group].cancel.snd[1], motif[main.group].cancel.snd[2])
						break
					elseif f == 'exit' then
						break
					elseif main.t_itemname[f] == nil then
						if t_storyModeList[f] then
							f = 'storyarc'
						elseif f:match('^bonus_') then
							f = 'bonus'
						elseif f:match('^ip_') then
							f = 'serverconnect'
						elseif tbl.submenu[f].loop ~= nil and #tbl.submenu[f].items > 0 then
							-- Check if the target submenu would immediately auto-run a single actionable item.
							-- If so, treat that as if it was selected directly here, so fadeout uses the current menu.
							local single_f2, single_t2 = main.f_getSingleMenuAction(tbl.submenu[f])
							if single_f2 ~= nil then
								f = single_f2
								call_t_override = single_t2
								call_item_override = 1
							else
								if motif.title_info.cursor[f] ~= nil and motif.title_info.cursor[f].snd ~= nil then
									sndPlay(motif.Snd, motif.title_info.cursor[f].snd[1], motif.title_info.cursor[f].snd[2])
								else
									sndPlay(motif.Snd, motif.title_info.cursor.done.snd[1], motif.title_info.cursor.done.snd[2])
								end
								tbl.submenu[f].loop()
								f = ''
								main.f_menuSnap(motif[main.group])
							end
						else
							break
						end
					end
					if f ~= '' then
						main.f_default()
						if f == 'joinadd' then
							tbl.items = main.t_itemname[f](t, item)
						elseif main.t_itemname[f] ~= nil then
							local call_t = call_t_override or t
							local call_item = call_item_override or item
							main.menu.f = main.t_itemname[f](call_t, call_item)
						end
						call_t_override = nil
						call_item_override = nil
						if main.menu.f ~= nil then
							if motif.title_info.cursor[f] ~= nil and motif.title_info.cursor[f].snd ~= nil then
								sndPlay(motif.Snd, motif.title_info.cursor[f].snd[1], motif.title_info.cursor[f].snd[2])
							else
								sndPlay(motif.Snd, motif.title_info.cursor.done.snd[1], motif.title_info.cursor.done.snd[2])
							end
							main.f_fadeReset('fadeout', motif[main.group])
						end
					end
				end
			end
		end
	end
end

-- Dynamically generates all menus and submenus
function main.f_start()
	main.menu = {title = main.f_itemnameUpper(motif[main.group].title.text, motif[main.group].menu.title.uppercase), submenu = {}, items = {}}
	main.menu.loop = main.f_createMenu(main.menu, true, main.group == 'title_info', main.group == 'title_info', false)
	local w = main.f_menuWindow(motif[main.group].menu)
	local t_pos = {} --for storing current main.menu table position
	local t_skipGroup = {}
	local lastNum = 0
	local bonusUpper = true
	for i, suffix in ipairs(motif[main.group].menu.itemname_order) do
		for j, c in ipairs(main.f_strsplit('_', suffix)) do --split using "_" delimiter
			--exceptions for expanding the menu table
			if motif[main.group].menu.itemname[suffix] == '' and c ~= 'server' then --items and groups without displayname are skipped
				t_skipGroup[c] = true
				break
			elseif t_skipGroup[c] then --named item but inside a group without displayname
				break
			elseif c == 'bonusgames' and #main.t_bonusChars == 0 then --skip bonus mode if there are no characters with bonus param set to 1
				t_skipGroup[c] = true
				break
			elseif c == 'storymode' and #main.t_selStoryMode == 0 then --skip story mode if there are no story arc declared
				t_skipGroup[c] = true
				break
			elseif c == 'versuscoop' and gameOption('Config.Players') < 4 then --skip versus coop if there are not enough players
				t_skipGroup[c] = true
				break
			end
			--appending the menu table
			if j == 1 then --first string after menu.itemname (either reserved one or custom submenu assignment)
				if main.menu.submenu[c] == nil then
					main.menu.submenu[c] = {title = main.f_itemnameUpper(motif[main.group].menu.itemname[suffix], motif[main.group].menu.title.uppercase), submenu = {}, items = {}}
					main.menu.submenu[c].loop = main.f_createMenu(main.menu.submenu[c], false, false, true, c == 'serverjoin')
					if not suffix:match(c .. '_') then
						table.insert(main.menu.items, {
							itemname = c,
							displayname = motif[main.group].menu.itemname[suffix],
							paramname = suffix,
						})
						if c == 'bonusgames' then bonusUpper = main.menu.items[#main.menu.items].displayname == main.menu.items[#main.menu.items].displayname:upper() end
					end
				end
				t_pos = main.menu.submenu[c]
				t_pos.name = c
			else --following strings
				if t_pos.submenu[c] == nil then
					t_pos.submenu[c] = {title = main.f_itemnameUpper(motif[main.group].menu.itemname[suffix], motif[main.group].menu.title.uppercase), submenu = {}, items = {}}
					t_pos.submenu[c].loop = main.f_createMenu(t_pos.submenu[c], false, false, true, c == 'serverjoin')
					table.insert(t_pos.items, {
						itemname = c,
						displayname = motif[main.group].menu.itemname[suffix],
						paramname = suffix,
					})
					if c == 'bonusgames' then bonusUpper = t_pos.items[#t_pos.items].displayname == t_pos.items[#t_pos.items].displayname:upper() end
				end
				if j > lastNum then
					t_pos = t_pos.submenu[c]
					t_pos.name = c
				end
			end
			lastNum = j
			--add bonus character names to bonusgames submenu
			if suffix:match('bonusgames_back$') and c == 'bonusgames' then --j == main.f_countSubstring(suffix, '_') then
				for k = 1, #main.t_bonusChars do
					local name = start.f_getCharData(main.t_bonusChars[k]).name
					local itemname = 'bonus_' .. name:gsub('%s+', '_')
					table.insert(t_pos.items, {
						itemname = itemname,
						displayname = main.f_itemnameUpper(name, bonusUpper),
						paramname = suffix:gsub('back$', itemname),
					})
				end
			end
			--add story arcs to storymode submenu
			if suffix:match('storymode_back$') and c == 'storymode' then --j == main.f_countSubstring(suffix, '_') then
				for k, v in ipairs(main.t_selStoryMode) do
					local itemname = v.name:gsub('%s+', '_')
					table.insert(t_pos.items, {
						itemname = itemname,
						displayname = v.displayname,
						paramname = suffix:gsub('back$', itemname),
					})
				end
			end
			--add IP addresses for serverjoin submenu
			if suffix:match('_serverjoin_back$') and c == 'serverjoin' then --j == main.f_countSubstring(suffix, '_') then
				for k, v in pairs(gameOption('Netplay.IP')) do
					local itemname = 'ip_' .. k
					table.insert(t_pos.items, {
						itemname = itemname,
						displayname = k,
						paramname = suffix:gsub('back$', itemname),
					})
				end
			end
		end
	end
	textImgSetWindow(motif[main.group].menu.item.selected.active.TextSpriteData, w[1], w[2], w[3], w[4])
	textImgSetWindow(motif[main.group].menu.item.active.TextSpriteData, w[1], w[2], w[3], w[4])
	textImgSetWindow(motif[main.group].menu.item.value.active.TextSpriteData, w[1], w[2], w[3], w[4])
	textImgSetWindow(motif[main.group].menu.item.selected.TextSpriteData, w[1], w[2], w[3], w[4])
	textImgSetWindow(motif[main.group].menu.item.TextSpriteData, w[1], w[2], w[3], w[4])
	textImgSetWindow(motif[main.group].menu.item.value.TextSpriteData, w[1], w[2], w[3], w[4])
	--for _, v in pairs(motif[main.group].menu.item.bg) do
	--	animSetWindow(v.AnimData, w[1], w[2], w[3], w[4])
	--end
	--for _, v in pairs(motif[main.group].menu.item.active.bg) do
	--	animSetWindow(v.AnimData, w[1], w[2], w[3], w[4])
	--end
	if gameOption('Debug.DumpLuaTables') then main.f_printTable(main.menu, 'debug/t_mainMenu.txt') end
end

function main.f_clearShuffleTables()
	start.shuffleChars = nil
	start.shuffleStages = nil
	start.shufflePals = nil
	start.lastRandomChar = nil
	start.lastRandomPal = nil
	start.lastStageIdx = nil
end

--replay menu
function main.f_replay()
	local w = main.f_menuWindow(motif.replay_info.menu)
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = {}
	for k, v in ipairs(getDirectoryFiles('save/replays')) do
		v:gsub('^(.-)([^\\/]+)%.([^%.\\/]-)$', function(path, filename, ext)
			path = path:gsub('\\', '/')
			ext = ext:lower()
			if ext == 'replay' then
				table.insert(t, {itemname = path .. filename .. '.' .. ext, displayname = filename})
			end
		end)
	end
	table.insert(t, {itemname = 'back', displayname = motif.replay_info.menu.itemname.back})
	bgReset(motif.replaybgdef.BGDef)
	main.f_fadeReset('fadein', motif.replay_info)
	playBgm({source = "motif.replay"})
	main.close = false
	while true do
		main.f_menuCommonDraw(t, item, cursorPosY, moveTxt, motif.replay_info, motif.replaybgdef, false)
		cursorPosY, moveTxt, item = main.f_menuCommonCalc(t, item, cursorPosY, moveTxt, motif.replay_info, motif.replay_info.cursor)
		if main.close and not main.fadeActive then
			bgReset(motif[main.background].BGDef)
			main.f_fadeReset('fadein', motif[main.group])
			playBgm({source = "motif.title", interrupt = true})
			main.close = false
			break
		elseif esc() or main.f_input(main.t_players, motif[main.group].menu.cancel.key) or (t[item].itemname == 'back' and main.f_input(main.t_players, motif[main.group].menu.done.key)) then
			sndPlay(motif.Snd, motif.replay_info.cancel.snd[1], motif.replay_info.cancel.snd[2])
			main.f_fadeReset('fadeout', motif.replay_info)
			main.close = true
		elseif main.f_input(main.t_players, motif[main.group].menu.done.key) then
			sndPlay(motif.Snd, motif[main.group].cursor.done.snd[1], motif[main.group].cursor.done.snd[2])
			enterReplay(t[item].itemname)
			synchronize()
			main.f_clearShuffleTables()
			math.randomseed(sszRandom())
			main.f_cmdBufReset()
			main.menu.submenu.server.loop()
			replayStop()
			exitNetPlay()
			exitReplay()
		end
	end
end

function main.f_connect(server, str)
	enterNetPlay(server)
	while not connected() do
		if esc() or main.f_input(main.t_players, motif.title_info.menu.cancel.key) then
			sndPlay(motif.Snd, motif.title_info.cancel.snd[1], motif.title_info.cancel.snd[2])
			exitNetPlay()
			return false
		end
		--draw clearcolor
		clearColor(motif[main.background].bgclearcolor[1], motif[main.background].bgclearcolor[2], motif[main.background].bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(motif[main.background].BGDef, 0)
		--draw layerno = 1 backgrounds
		bgDraw(motif[main.background].BGDef, 1)
		--draw overlay
		rectDraw(motif.title_info.connecting.overlay.RectData)
		--draw text
		local txt = ''
		if server == '' then
			txt = string.format(motif.title_info.connecting.text.host, str)
		else
			txt = string.format(motif.title_info.connecting.text.join, server, str)
		end
		textImgReset(motif.title_info.connecting.TextSpriteData)
		textImgSetText(motif.title_info.connecting.TextSpriteData, txt)
		textImgDraw(motif.title_info.connecting.TextSpriteData)
		main.f_cmdInput()
		refresh()
	end
	replayRecord('save/replays/' .. os.date("%Y-%m-%d_%Hh%Mm%Ss") .. '.replay')
	return true
end

--asserts content unlock conditions
function main.f_unlock(permanent)
	local refreshRandom = false
	for group, t in pairs(main.t_unlockLua) do
		local t_del = {}
		for k, v in pairs(t) do
			local bool = assert(loadstring('return ' .. v))()
			if type(bool) == 'boolean' then
				if group == 'chars' then
					if main.f_unlockChar(k, bool, false) then
						refreshRandom = true
					end
				elseif group == 'stages' then
					main.f_unlockStage(k, bool)
				elseif group == 'modes' then
					--already handled via t_del cleaning
				end
				if bool and (permanent or group == 'modes') then
					table.insert(t_del, k)
				end
			else
				panicError("\nmain.t_unlockLua." .. group .. "[" .. k .. "]\n" .. "Following Lua code does not return boolean value: \n" .. v .. "\n")
			end
		end
		--clean lua code that already returned true
		for k, v in ipairs(t_del) do
			t[v] = nil
		end
	end
	-- If any character visibility changed, rebuild random pool and clear shuffle bags so RandomSelect immediately reflects the updated roster.
	if refreshRandom then
		main.f_updateRandomChars()
		if start ~= nil then
			main.f_clearShuffleTables()
		end
	end
end

--unlock characters (select screen grid only)
function main.f_unlockChar(num, bool, reset)
	local changed = false
	if bool then
		if main.t_selChars[num].hidden ~= 0 then
			main.t_selChars[num].hidden_default = main.t_selChars[num].hidden
			main.t_selChars[num].hidden = 0
			for k, t in pairs({order = main.t_orderChars, ordersurvival = main.t_orderSurvival}) do
				if main.t_selChars[num][k] ~= nil and main.t_selChars[num][k] < 0 then
					main.t_selChars[num][k] = 0 - main.t_selChars[num][k]
					if t[main.t_selChars[num][k]] == nil then
						t[main.t_selChars[num][k]] = {}
					end
					table.insert(t[main.t_selChars[num][k]], main.t_selChars[num].char_ref)
				end
			end
			start.t_grid[main.t_selChars[num].row][main.t_selChars[num].col].hidden = main.t_selChars[num].hidden
			if reset then start.f_resetGrid() end
			changed = true
		end
	elseif main.t_selChars[num].hidden_default == nil then
		return false
	elseif main.t_selChars[num].hidden ~= main.t_selChars[num].hidden_default then
		main.t_selChars[num].hidden = main.t_selChars[num].hidden_default
		start.t_grid[main.t_selChars[num].row][main.t_selChars[num].col].hidden = main.t_selChars[num].hidden
		if reset then start.f_resetGrid() end
		changed = true
	end
	return changed
end

--unlock stages (stage selection menu only)
function main.f_unlockStage(num, bool)
	if bool then
		if main.t_selStages[num].hidden ~= 0 then
			main.t_selStages[num].hidden_default = main.t_selStages[num].hidden
			main.t_selStages[num].hidden = 0
			main.f_updateSelectableStages()
		end
	elseif main.t_selStages[num].hidden_default == nil then
		return
	elseif main.t_selStages[num].hidden ~= main.t_selStages[num].hidden_default then
		main.t_selStages[num].hidden = main.t_selStages[num].hidden_default
		main.f_updateSelectableStages()
	end
end

function main.f_hiscoreDisplay(itemname)
	local stats = jsonDecode(getCommandLineValue("-stats"))
	if not motif.hiscore_info.enabled or stats.modes == nil or stats.modes[itemname] == nil or stats.modes[itemname].ranking == nil then
		return false
	end
	main.f_cmdBufReset()
	sndPlay(motif.Snd, motif[main.group].cursor.done.snd[1], motif[main.group].cursor.done.snd[2])
	main.f_hiscore(itemname, -1)
	--main.f_fadeReset('fadein', motif[main.group])
	playBgm({source = "motif.title", interrupt = true})
	return true
end

--attract mode start screen
function main.f_attractStart()
	local timerActive = credits() ~= 0
	local timer = 0
	local counter = 0 - motif.attract_mode.fadein.time
	local press_blinktime, insert_blinktime = 0, 0
	local press_switched, insert_switched = false, false
	local drawPress, drawInsert = true, true
	main.f_cmdBufReset()
	clearColor(motif.attractbgdef.bgclearcolor[1], motif.attractbgdef.bgclearcolor[2], motif.attractbgdef.bgclearcolor[3])
	bgReset(motif.attractbgdef.BGDef)
	main.f_fadeReset('fadein', motif.attract_mode)
	local fadeOutStarted = false
	playBgm({source = "motif.title", interrupt = true})
	local creditsCnt = credits()
	while true do
		counter = counter + 1
		--draw layerno = 0 backgrounds
		bgDraw(motif.attractbgdef.BGDef, 0)
		--draw layerno = 1 backgrounds
		bgDraw(motif.attractbgdef.BGDef, 1)
		--draw text
		if credits() ~= 0 then
			if motif.attract_mode.start.press.blinktime > 0 and not fadeOutStarted then
				if press_blinktime < motif.attract_mode.start.press.blinktime then
					press_blinktime = press_blinktime + 1
				elseif press_switched then
					drawPress = true
					press_switched = false
					press_blinktime = 0
				else
					drawPress = false
					press_switched = true
					press_blinktime = 0
				end
			end
			if drawPress then
				textImgDraw(motif.attract_mode.start.press.TextSpriteData)
			end
		else
			if motif.attract_mode.start.insert.blinktime > 0 and not fadeOutStarted then
				if insert_blinktime < motif.attract_mode.start.insert.blinktime then
					insert_blinktime = insert_blinktime + 1
				elseif insert_switched then
					drawInsert = true
					insert_switched = false
					insert_blinktime = 0
				else
					drawInsert = false
					insert_switched = true
					insert_blinktime = 0
				end
			end
			if drawInsert then
				textImgDraw(motif.attract_mode.start.insert.TextSpriteData)
			end
		end
		--draw timer
		if motif.attract_mode.start.timer.count ~= -1 and timerActive then
			timer, timerActive = main.f_drawTimer(timer, motif.attract_mode.start.timer)
		end
		--draw credits text
		if credits() ~= -1 then
			textImgReset(motif.attract_mode.credits.TextSpriteData)
			textImgSetText(motif.attract_mode.credits.TextSpriteData, string.format(motif.attract_mode.credits.text, credits()))
			textImgDraw(motif.attract_mode.credits.TextSpriteData)
		end
		--credits
		if creditsCnt ~= credits() then
			timerActive = true
			timer = motif.attract_mode.start.timer.displaytime
			creditsCnt = credits()
		end
		--options
		if motif.attract_mode.enabled and getKey(motif.attract_mode.options.keycode) then
			main.f_default()
			main.menu.f = main.t_itemname.options()
			sndPlay(motif.Snd, motif[main.group].cursor.done.snd[1], motif[main.group].cursor.done.snd[2])
			main.f_fadeReset('fadeout', motif[main.group])
			fadeOutStarted = true
			resetKey()
			main.menu.f()
			return false
		end
		--draw fadein / fadeout
		if not fadeOutStarted and not main.fadeActive and ((credits() ~= 0 and main.f_input(main.t_players, motif.attract_mode.start.press.key)) or (not timerActive and counter >= motif.attract_mode.start.time)) then
			if credits() ~= 0 then
				sndPlay(motif.Snd, motif.attract_mode.start.done.snd[1], motif.attract_mode.start.done.snd[2])
			end
			main.f_fadeReset('fadeout', motif.attract_mode)
			fadeOutStarted = true
		end
		main.f_fadeAnim(motif.attract_mode)
		--frame transition
		main.f_cmdInput()
		if esc() --[[or main.f_input(main.t_players, motif.attract_mode.menu.cancel.key)]] then
			esc(false)
			return false
		end
		if fadeOutStarted and not main.fadeActive then
			return credits() ~= 0
		end
		refresh()
	end
end

--attract mode loop
function main.f_attractMode()
	setCredits(0)
	while true do --outer loop
		local startScreen = false
		while true do --inner loop (attract mode)
			--logo storyboard
			if motif.attract_mode.logo.storyboard ~= '' and main.f_storyboard(motif.attract_mode.logo.storyboard) then
				break
			end
			--intro storyboard
			if motif.attract_mode.intro.storyboard ~= '' and main.f_storyboard(motif.attract_mode.intro.storyboard) then
				break
			end
			--demo
			main.f_demoStart()
			if credits() > 0 then break end
			--hiscores
			main.f_hiscore("arcade", -1)
			if credits() > 0 then break end
			--start
			if main.f_attractStart() then
				startScreen = true
				break
			end
			--demo
			main.f_demoStart()
			if credits() > 0 then break end
			--hiscores
			main.f_hiscore("arcade", -1)
			if credits() > 0 then break end
		end
		if startScreen or main.f_attractStart() then
			--attract storyboard
			if motif.attract_mode.start.storyboard ~= '' then
				main.f_storyboard(motif.attract_mode.start.storyboard)
			end
			--eat credit
			if credits() > 0 then
				setCredits(credits() - 1)
			end
			--enter menu
			main.menu.loop()
		elseif credits() > 0 then
			setCredits(credits() - 1)
		end
	end
end

setCredits(-1)
function main.f_setCredits()
	if motif.attract_mode.enabled or start.challenger ~= 0 then
		return
	end
	setCredits(gameOption('Options.Credits') - 1)
end

--demo mode
function main.f_demo()
	if #main.t_randomChars == 0 then
		return
	end
	if main.fadeActive or not motif.demo_mode.enabled then
		demoFrameCounter = 0
		return
	end
	demoFrameCounter = demoFrameCounter + 1
	if demoFrameCounter < motif.demo_mode.title.waittime then
		return
	end
	main.f_fadeReset('fadeout', motif.demo_mode)
	main.menu.f = main.t_itemname.demo()
end

--prevents mirrored palette in demo mode mirror matches
local function getUniquePalette(ch, prev)
	local charData = start.f_getCharData(ch)
	local pals = charData and charData.pal or {1}

	if not prev or ch ~= prev.ch then
		return pals[sszRandom() % #pals + 1]
	end

	local available = {}
	for _, p in ipairs(pals) do
		if p ~= prev.pal then
			table.insert(available, p)
		end
	end

	if #available > 0 then
		return available[sszRandom() % #available + 1]
	else
		return prev.pal
	end
end

function main.f_demoStart()
	main.f_default()
	setLifebarElements({bars = motif.demo_mode.fight.bars.display})
	setGameMode('demo')
	for i = 1, 2 do
		setCom(i, 8)
		setTeamMode(i, 0, 1)
		local ch = main.t_randomChars[math.random(1, #main.t_randomChars)]
		local pal = getUniquePalette(ch, prev)

		selectChar(i, ch, pal)
		prev = {ch = ch, pal = pal}
	end
	start.f_setStage()
	if motif.demo_mode.fight.stopbgm then
		stopBgm()
	end
	hook.run("main.t_itemname")
	--clearColor(motif[main.background].bgclearcolor[1], motif[main.background].bgclearcolor[2], motif[main.background].bgclearcolor[3])
	loadStart()
	game()
	if not motif.attract_mode.enabled then
		if introWaitCycles >= motif.demo_mode.intro.waitcycles then
			main.f_hiscore("arcade", -1)
			if motif.files.intro.storyboard ~= '' then
				main.f_storyboard(motif.files.intro.storyboard)
			end
			introWaitCycles = 0
		else
			introWaitCycles = introWaitCycles + 1
		end
		bgReset(motif[main.background].BGDef)
		--start title BGM only if it has been interrupted
		if motif.demo_mode.fight.stopbgm or motif.demo_mode.fight.playbgm or (introWaitCycles == 0 and motif.files.intro.storyboard ~= '') then
			playBgm({source = "motif.title", interrupt = true})
		end
	end
	main.f_fadeReset('fadein', motif.demo_mode)
end

--calculate menu.tween and boxcursor.tween
local function f_tweenStep(val, target, factor)
	if not factor or factor <= 0 then
		return target
	end
	local newVal = val + (target - val) * math.min(factor, 1)
	if math.abs(newVal - target) < 1 then
		return target
	end
	return newVal
end

--common menu calculations
function main.f_menuCommonCalc(t, item, cursorPosY, moveTxt, sec, cursorParams, forcedDir)
	-- persistent scroll tween per section
	if not sec.menuTweenData then
		sec.menuTweenData = {
			currentPos = 0,
			targetPos = 0,
			slideOffset = 0
		}
	end
	local startItem = 1
	for _, v in ipairs(t) do
		if not v.itemname:match("^spacer%d*$") then
			break
		end
		startItem = startItem + 1
	end
	-- effective visible-items: treat 0 / nil as "all items"
	local visible = #t
	if sec.menu and sec.menu.window and sec.menu.window.visibleitems ~= nil then
		if sec.menu.window.visibleitems > 0 then
			visible = sec.menu.window.visibleitems
		end
	end
	-- movement: forcedDir: 1 = next (down), -1 = previous (up), 0/nil = no forced move
	local moveDir = 0
	if forcedDir ~= nil then
		moveDir = forcedDir
	elseif main.f_input(main.t_players, sec.menu.next.key) then
		moveDir = 1
	elseif main.f_input(main.t_players, sec.menu.previous.key) then
		moveDir = -1
	end
	if moveDir == 1 then
		sndPlay(motif.Snd, cursorParams.move.snd[1], cursorParams.move.snd[2])
		while true do
			item = item + 1
			if cursorPosY < visible then
				cursorPosY = cursorPosY + 1
			end
			if t[item] == nil or not t[item].itemname:match("^spacer%d*$") then
				break
			end
		end
	elseif moveDir == -1 then
		sndPlay(motif.Snd, cursorParams.move.snd[1], cursorParams.move.snd[2])
		while true do
			item = item - 1
			if cursorPosY > startItem then
				cursorPosY = cursorPosY - 1
			end
			if t[item] == nil or not t[item].itemname:match("^spacer%d*$") then
				break
			end
		end
	end
	main.menuWrapped = false
	if item > #t or (item == 1 and t[item].itemname:match("^spacer%d*$")) then
		item = 1
		while true do
			if not t[item].itemname:match("^spacer%d*$") or item >= #t then break end
			item = item + 1
		end
		cursorPosY = item
		if sec.menu.tween.wrap.snap then
			main.menuSnap = true
		end
		main.menuWrapped = true
	elseif item < 1 then
		item = #t
		while true do
			if not t[item].itemname:match("^spacer%d*$") or item <= 1 then break end
			item = item - 1
		end
		if item > visible then
			cursorPosY = visible
		else
			cursorPosY = item
		end
		if sec.menu.tween.wrap.snap then
			main.menuSnap = true
		end
		main.menuWrapped = true
	end
	-- compute target: determine first visible item to keep cursor at row `cursorPosY`, clamp to valid range, and convert to pixel offset
	local spacing = sec.menu.item.spacing[2]
	-- max index that can appear at the top of the window
	local maxFirst = math.max(1, #t - visible + 1)
	local t_factor = sec.menu.tween.factor

	-- which list index should be drawn on the very first row
	local desiredFirst = item - cursorPosY + 1
	-- clamp so we never scroll before the first row or past the end
	if desiredFirst < 1 then
		desiredFirst = 1
	elseif desiredFirst > maxFirst then
		desiredFirst = maxFirst
	end

	-- Measure scroll offset from the first drawn row.
	local targetMove = (desiredFirst - 1) * spacing
	-- update target and offset if changed, snap immediately if requested, otherwise apply tween or direct move
	if sec.menuTweenData.targetPos ~= targetMove then
		sec.menuTweenData.targetPos = targetMove
		sec.menuTweenData.slideOffset = sec.menuTweenData.currentPos - sec.menuTweenData.targetPos
	end

	if main.menuSnap then
		sec.menuTweenData.currentPos = sec.menuTweenData.targetPos
		sec.menuTweenData.slideOffset = 0
		main.menuSnap = false
		moveTxt = sec.menuTweenData.currentPos
		return cursorPosY, moveTxt, item
	end

	if t_factor[1] > 0 then
		sec.menuTweenData.slideOffset = f_tweenStep(sec.menuTweenData.slideOffset, 0, t_factor[1])
		sec.menuTweenData.currentPos = sec.menuTweenData.targetPos + sec.menuTweenData.slideOffset
	else
		sec.menuTweenData.currentPos = sec.menuTweenData.targetPos
		sec.menuTweenData.slideOffset = 0
	end
	moveTxt = sec.menuTweenData.currentPos
	return cursorPosY, moveTxt, item
end

--common menu draw
function main.f_menuCommonDraw(t, item, cursorPosY, moveTxt, sec, bg, skipClear, opts)
	-- opts:
	--   offx, offy               : per-call offsets
	--   skipBG0, skipBG1         : skip bg layer 0 / 1
	--   skipTitle                : skip drawing the title
	--   forceInactive            : treat "selected" row as inactive (no highlight, no cursor)
	--   skipInput                : do not call main.f_cmdInput() inside this function
	opts = opts or {}
	local offx = opts.offx or 0
	local offy = opts.offy or 0
	local forceInactive = (opts.forceInactive == true)
	local skipInput = (opts.skipInput == true)

	-- effective visible-items: treat 0 or 'unlimitedItems' as "all"
	local visible = (sec.menu and sec.menu.window and sec.menu.window.visibleitems) or #t
	if not visible or visible <= 0 then
		visible = #t
	end

	if not skipClear then
		clearColor(bg.bgclearcolor[1], bg.bgclearcolor[2], bg.bgclearcolor[3])
	end
	--draw layerno = 0 backgrounds
	if not opts.skipBG0 then
		bgDraw(bg.BGDef, 0)
	end
	--draw layerno = 1 backgrounds
	if not opts.skipBG1 then
		bgDraw(bg.BGDef, 1)
	end

	--draw menu box
	if sec.menu.boxbg.visible then
		local x1 = offx + sec.menu.pos[1] + sec.menu.boxcursor.coords[1]
		local y1 = offy + sec.menu.pos[2] + sec.menu.boxcursor.coords[2]
		local w  = sec.menu.boxcursor.coords[3] - sec.menu.boxcursor.coords[1] + 1
		local h  = sec.menu.boxcursor.coords[4] - sec.menu.boxcursor.coords[2] + 1
			+ (math.min(#t, visible) - 1) * sec.menu.item.spacing[2]
		rectSetWindow(sec.menu.boxbg.RectData, x1, y1, x1 + w, y1 + h)
		rectUpdate(sec.menu.boxbg.RectData)
		rectDraw(sec.menu.boxbg.RectData)
	end
	--draw title
	if not opts.skipTitle and sec.title and sec.title.TextSpriteData then
		textImgDraw(sec.title.TextSpriteData)
	end
	--draw menu items
	local cur = moveTxt
	local tgt = moveTxt
	if sec.menuTweenData then
		cur = sec.menuTweenData.currentPos or sec.menuTweenData.slideOffset or cur
		tgt = sec.menuTweenData.targetPos or tgt
	end
	local tweenDone = math.abs((cur or 0) - (tgt or 0)) < 1

	local items_shown = item + visible - cursorPosY
	if items_shown > #t or (visible > 0 and items_shown < #t and (sec.menu.window.margins.y[1] ~= 0 or sec.menu.window.margins.y[2] ~= 0)) then
		items_shown = #t
	end

	-- helper that draws a single item either as active or inactive
	local function drawItem(i, isActive)
		local itemData = t[i]
		-- display name
		local displayname = main.f_itemnameUpper(itemData.displayname, sec.menu.item.uppercase)
		if itemData.itemname:match("^spacer%d*$") then
			displayname = ""
		end
		-- shared position
		local posX = offx + (i - 1) * sec.menu.item.spacing[1]
		local posY = offy + (i - 1) * sec.menu.item.spacing[2] - moveTxt
		-- background params
		local bgTable = isActive and sec.menu.item.active.bg or sec.menu.item.bg
		local params = bgTable[itemData.paramname] or bgTable.default
		-- draw background
		local bgPosX = offx 
		local bgPosY = offy
		if params.spacing[1] ~= 0 or params.spacing[2] ~= 0 then
			bgPosX = bgPosY + (i - 1) * params.spacing[1]
			bgPosY = bgPosY + (i - 1) * params.spacing[2] - moveTxt
		end
		main.f_animPosDraw(params.AnimData, bgPosX, bgPosY)
		-- text sprite for label
		local labelSprite
		if itemData.selected then
			labelSprite = isActive and sec.menu.item.selected.active.TextSpriteData or sec.menu.item.selected.TextSpriteData
		else
			labelSprite = isActive and sec.menu.item.active.TextSpriteData or sec.menu.item.TextSpriteData
		end
		textImgReset(labelSprite)
		textImgAddPos(labelSprite, posX, posY)
		textImgSetText(labelSprite, displayname)
		textImgDraw(labelSprite)
		-- value / info sprites
		if itemData.vardisplay ~= nil then
			if itemData.conflict and sec.menu.item.value.conflict and sec.menu.item.value.conflict.TextSpriteData then
				local spr = sec.menu.item.value.conflict.TextSpriteData
				textImgReset(spr)
				textImgAddPos(spr, posX, posY)
				textImgSetText(spr, itemData.vardisplay)
				textImgDraw(spr)
			else
				local spr = isActive and sec.menu.item.value.active.TextSpriteData or sec.menu.item.value.TextSpriteData
				textImgReset(spr)
				textImgAddPos(spr, posX, posY)
				textImgSetText(spr, itemData.vardisplay)
				textImgDraw(spr)
			end
		elseif itemData.infodisplay ~= nil then
			local spr = isActive and sec.menu.item.info.active.TextSpriteData or sec.menu.item.info.TextSpriteData
			textImgReset(spr)
			textImgAddPos(spr, posX, posY)
			textImgSetText(spr, itemData.infodisplay)
			textImgDraw(spr)
		end
	end
	-- draw not active items
	for i = 1, items_shown do
		if i > item - cursorPosY or not tweenDone then
			local isSelected = (i == item) and (not forceInactive)
			if not isSelected then
				drawItem(i, false)
			end
		end
	end
	-- draw active items
	for i = 1, items_shown do
		if i > item - cursorPosY or not tweenDone then
			local isSelected = (i == item) and (not forceInactive)
			if isSelected then
				drawItem(i, true)
			end
		end
	end
	--initialize storage for boxcursor per section if missing
	if not sec.boxCursorData then
		sec.boxCursorData = {
			offsetY = 0, 
			snap = 0, 
			init = false
		}
	end
	--calculate target Y position for cursor
	local targetY = offy + sec.menu.pos[2] + sec.menu.boxcursor.coords[2] + (cursorPosY - 1) * sec.menu.item.spacing[2]
	local t_factor = sec.menu.boxcursor.tween.factor
	--snap cursor immediately if first use or snap enabled
	if sec.boxCursorData.snap == 1 or not sec.boxCursorData.init then
		sec.boxCursorData.offsetY = targetY
		sec.boxCursorData.init = true
		sec.boxCursorData.snap = -1
	end
	if sec.menu.boxcursor.tween.wrap.snap and main.menuWrapped then
		sec.boxCursorData.offsetY = targetY
	end
	--apply tween if enabled, otherwise snap to target
	if t_factor[1] > 0 then
		sec.boxCursorData.offsetY = f_tweenStep(sec.boxCursorData.offsetY, targetY, t_factor[1])
	else
		sec.boxCursorData.offsetY = targetY
	end
	--draw menu cursor
	if sec.menu.boxcursor.visible and not main.fadeActive and not forceInactive then
		local x1 = offx + sec.menu.pos[1] + sec.menu.boxcursor.coords[1] + (cursorPosY - 1) * sec.menu.item.spacing[1]
		local y1 = sec.boxCursorData.offsetY
		local w  = sec.menu.boxcursor.coords[3] - sec.menu.boxcursor.coords[1] + 1
		local h  = sec.menu.boxcursor.coords[4] - sec.menu.boxcursor.coords[2] + 1
		rectSetWindow(sec.menu.boxcursor.RectData, x1, y1, x1 + w, y1 + h)
		rectUpdate(sec.menu.boxcursor.RectData)
		rectDraw(sec.menu.boxcursor.RectData)
	end
	--draw scroll arrows
	if #t > visible then
		if item > cursorPosY then
			animReset(sec.menu.arrow.up.AnimData, {'pos'})
			animAddPos(sec.menu.arrow.up.AnimData, offx, offy)
			animUpdate(sec.menu.arrow.up.AnimData)
			animDraw(sec.menu.arrow.up.AnimData)
		end
		if item >= cursorPosY and item + visible - cursorPosY < #t then
			animReset(sec.menu.arrow.down.AnimData, {'pos'})
			animAddPos(sec.menu.arrow.down.AnimData, offx, offy)
			animUpdate(sec.menu.arrow.down.AnimData)
			animDraw(sec.menu.arrow.down.AnimData)
		end
	end
	--draw credits text
	if motif.attract_mode.enabled and credits() ~= -1 then
		textImgReset(motif.attract_mode.credits.TextSpriteData)
		textImgSetText(motif.attract_mode.credits.TextSpriteData, string.format(motif.attract_mode.credits.text, credits()))
		textImgDraw(motif.attract_mode.credits.TextSpriteData)
	end
	--draw footer
	if sec.footer ~= nil then
		rectDraw(sec.footer.overlay.RectData)
		textImgDraw(sec.footer.title.TextSpriteData)
		textImgDraw(sec.footer.info.TextSpriteData)
		textImgDraw(sec.footer.version.TextSpriteData)
	end
	--draw fadein / fadeout
	main.f_fadeAnim(main.fadeGroup)
	--frame transition
	if not skipInput then
		if main.fadeActive or main.fadeCnt > 0 then
			main.f_cmdBufReset()
		elseif main.fadeType == 'fadeout' then
			main.f_cmdBufReset()
			return --skip last frame rendering
		else
			main.f_cmdInput()
		end
	end
	if not skipClear then
		refresh()
	end
end

--force menu to snap, boxcursor snaps if enabled
function main.f_menuSnap(sec)
	main.menuSnap = true
	if sec.boxCursorData then
		sec.boxCursorData.snap = sec.menu.boxcursor.tween.snap
	end
end

--common timer draw code
function main.f_drawTimer(timer, params)
	local num = main.f_round((params.count * params.framespercount - timer + params.displaytime) / params.framespercount)
	local active = true
	if num <= -1 then
		active = false
		timer = -1
		textImgReset(params.TextSpriteData)
		textImgSetText(params.TextSpriteData, string.format(params.text, 0))		
	elseif timer ~= -1 then
		timer = timer + 1
		textImgReset(params.TextSpriteData)
		textImgSetText(params.TextSpriteData, string.format(params.text, math.max(0, num)))
	end
	if timer == -1 or timer >= params.displaytime then
		textImgDraw(params.TextSpriteData)
	end
	return timer, active
end

--;===========================================================
--; EXTERNAL LUA CODE
--;===========================================================
local t_modules = {}
for _, v in ipairs(getDirectoryFiles('external/mods')) do
	if v:lower():match('%.([^%.\\/]-)$') == 'lua' then
		table.insert(t_modules, v)
	end
end

for _, v in ipairs(gameOption('Common.Modules')) do
	table.insert(t_modules, v)
end
if motif.files.module ~= '' then table.insert(t_modules, motif.files.module) end
for _, v in ipairs(t_modules) do
	print('Loading module: ' .. v)
	v = v:gsub('^%s*[%./\\]*', '')
	v = v:gsub('%.[^%.]+$', '')
	require(v:gsub('[/\\]+', '.'))
end

main.f_unlock(false)

--;===========================================================
--; INITIALIZE LOOPS
--;===========================================================
if gameOption('Debug.DumpLuaTables') then
	main.f_printTable(main.t_selChars, "debug/t_selChars.txt")
	main.f_printTable(main.t_selStages, "debug/t_selStages.txt")
	main.f_printTable(main.t_selOptions, "debug/t_selOptions.txt")
	main.f_printTable(main.t_selStoryMode, "debug/t_selStoryMode.txt")
	main.f_printTable(main.t_orderChars, "debug/t_orderChars.txt")
	main.f_printTable(main.t_orderStages, "debug/t_orderStages.txt")
	main.f_printTable(main.t_orderSurvival, "debug/t_orderSurvival.txt")
	main.f_printTable(main.t_randomChars, "debug/t_randomChars.txt")
	main.f_printTable(main.t_bonusChars, "debug/t_bonusChars.txt")
	main.f_printTable(main.t_stageDef, "debug/t_stageDef.txt")
	main.f_printTable(main.t_charDef, "debug/t_charDef.txt")
	main.f_printTable(main.t_includeStage, "debug/t_includeStage.txt")
	main.f_printTable(main.t_selectableStages, "debug/t_selectableStages.txt")
	main.f_printTable(main.t_selGrid, "debug/t_selGrid.txt")
	main.f_printTable(main.t_unlockLua, "debug/t_unlockLua.txt")
	main.f_printTable(loadGameOption(), "debug/config.txt")
end

main.f_start()
menu.f_start()
options.f_start()

if getCommandLineValue("-p1") ~= nil and getCommandLineValue("-p2") ~= nil then
	main.f_default()
	main.f_commandLine()
end

main.f_loadingRefresh()

if motif.attract_mode.enabled then
	main.f_attractMode()
else
	main.menu.loop()
end

-- Debug Info
--if gameOption('Debug.DumpLuaTables') then main.f_printTable(main, "debug/t_main.txt") end

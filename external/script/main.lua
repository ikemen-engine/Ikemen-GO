main = {}
--nClock = os.clock()
--print("Elapsed time: " .. os.clock() - nClock)
--;===========================================================
--; INITIALIZE DATA
--;===========================================================
math.randomseed(os.time())

main.flags = getCommandLineFlags()
if main.flags['-config'] == nil then main.flags['-config'] = 'save/config.json' end
if main.flags['-stats'] == nil then main.flags['-stats'] = 'save/stats.json' end

--One-time load of the json routines
json = (loadfile 'external/script/json.lua')()

--;===========================================================
--; COMMON FUNCTIONS
--;===========================================================

--return file content
function main.f_fileRead(path, mode)
	local file = io.open(path, mode or 'r')
	if file == nil then
		panicError("\nFile doesn't exist: " .. path)
		return
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

--Data loading from config.json
config = json.decode(main.f_fileRead(main.flags['-config']))

--Data loading from stats.json
stats = json.decode(main.f_fileRead(main.flags['-stats']))

--add default commands
main.t_commands = {
	['$U'] = 0, ['$D'] = 0, ['$B'] = 0, ['$F'] = 0, ['a'] = 0, ['b'] = 0, ['c'] = 0, ['x'] = 0, ['y'] = 0, ['z'] = 0, ['s'] = 0, ['d'] = 0, ['w'] = 0, ['m'] = 0, ['/s'] = 0, ['/d'] = 0, ['/w'] = 0}
function main.f_commandNew()
	local c = commandNew()
	for k, _ in pairs(main.t_commands) do
		commandAdd(c, k, k)
	end
	return c
end

--prepare players/command tables
function main.f_setPlayers(num, default)
	setPlayers(num)
	main.t_players = {}
	main.t_remaps = {}
	main.t_lastInputs = {}
	main.t_cmd = {}
	main.t_pIn = {}
	for i = 1, num do
		table.insert(main.t_players, i)
		table.insert(main.t_remaps, i)
		table.insert(main.t_lastInputs, {})
		table.insert(main.t_cmd, main.f_commandNew())
		table.insert(main.t_pIn, i)
		local new = false
		if i > #config.KeyConfig then
			table.insert(config.KeyConfig, {Joystick = -1, Buttons = {'', '', '', '', '', '', '', '', '', '', '', '', '', ''}})
			new = true
		end
		if i > #config.JoystickConfig then
			table.insert(config.JoystickConfig, {Joystick = i - 1, Buttons = {'', '', '', '', '', '', '', '', '', '', '', '', '', ''}})
			new = true
		end
		if new and default then
			options.f_keyDefault(i)
		end
	end
	for i = 1, #config.KeyConfig - num do
		table.remove(config.KeyConfig, #config.KeyConfig)
	end
	for i = 1, #config.JoystickConfig - num do
		table.remove(config.JoystickConfig, #config.JoystickConfig)
	end
end
main.f_setPlayers(config.Players, false)

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
	for i = 1, config.Players do
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
	for i = 1, config.Players do
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
function main.f_input(p, b)
	for _, pn in ipairs(p) do
		for _, btn in ipairs(b) do
			if btn == 'pal' then
				if main.f_btnPalNo(pn) > 0 then
					main.playerInput = pn
					return true
				end
			elseif commandGetState(main.t_cmd[pn], btn) then
				main.playerInput = pn
				return true
			end
		end
	end
	return false
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

--return table with key names
function main.f_extractKeys(str)
	local t = {}
	if str ~= nil then
		for i, c in ipairs(main.f_strsplit('%s*&%s*', str)) do --split string using "%s*&%s*" delimiter
			t[i] = c
		end
	end
	return t
end

--check if a file or directory exists in this path
function main.f_exists(file)
	local ok, err, code = os.rename(file, file)
	if not ok then
		if code == 13 then
			--permission denied, but it exists
			return true
		end
	end
	return ok, err
end
--check if a directory exists in this path
function  main.f_isdir(path)
	-- "/" works on both Unix and Windows
	return main.f_exists(path .. '/')
end

main.debugLog = false
if main.f_isdir('debug') then
	main.debugLog = true
end

--check if file exists
function main.f_fileExists(file)
	if file == '' then
		return false
	end
	local f = io.open(file,'r')
	if f ~= nil then
		io.close(f)
		return true
	end
	return false
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
	main.f_fileWrite(toFile or 'debug/table_print.txt', txt)
end

--prints "v" variable into "toFile" file
function main.f_printVar(v, toFile)
	main.f_fileWrite(toFile or 'debug/var_print.txt', v)
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
if main.flags['-ailevel'] ~= nil then
	config.Difficulty = math.max(1, math.min(tonumber(main.flags['-ailevel']), 8))
end
if main.flags['-speed'] ~= nil and tonumber(main.flags['-speed']) > 0 then
	setGameSpeed(tonumber(main.flags['-speed']) * config.Framerate / 100)
end
if main.flags['-speedtest'] ~= nil then
	setGameSpeed(100 * config.Framerate)
end
if main.flags['-nosound'] ~= nil then
	setVolumeMaster(0)
end
if main.flags['-togglelifebars'] ~= nil then
	toggleStatusDraw()
end
if main.flags['-maxpowermode'] ~= nil then
	toggleMaxPowerMode()
end
if main.flags['-debug'] ~= nil then
	toggleDebugDraw()
end
if main.flags['-setport'] ~= nil then
	setListenPort(main.flags['-setport'])
end

--motif
main.motifDef = config.Motif
if main.flags['-r'] ~= nil or main.flags['-rubric'] ~= nil then
	local case = main.flags['-r']:lower() or main.flags['-rubric']:lower()
	if case:match('^data[/\\]') and main.f_fileExists(main.flags['-r']) then
		main.motifDef = main.flags['-r'] or main.flags['-rubric']
	elseif case:match('%.def$') and main.f_fileExists('data/' .. main.flags['-r']) then
		main.motifDef = 'data/' .. (main.flags['-r'] or main.flags['-rubric'])
	elseif main.f_fileExists('data/' .. main.flags['-r'] .. '/system.def') then
		main.motifDef = 'data/' .. (main.flags['-r'] or main.flags['-rubric']) .. '/system.def'
	end
end
main.motifDir, main.motifFile = main.motifDef:match('^(.-)[^/\\]+$')
setMotifDir(main.motifDir)

--lifebar
main.motifData = main.f_fileRead(main.motifDef)
local fileDir = main.motifDef:match('^(.-)[^/\\]+$')
if main.flags['-lifebar'] ~= nil then
	main.lifebarDef = main.flags['-lifebar']
else
	main.lifebarDef = main.motifData:match('\n%s*fight%s*=%s*(.-%.def)%s*')
end
if main.f_fileExists(main.lifebarDef) then
	--do nothing
elseif main.f_fileExists(fileDir .. main.lifebarDef) then
	main.lifebarDef = fileDir .. main.lifebarDef
elseif main.f_fileExists('data/' .. main.lifebarDef) then
	main.lifebarDef = 'data/' .. main.lifebarDef
else
	main.lifebarDef = 'data/fight.def'
end
main.lifebarData = main.f_fileRead(main.lifebarDef)
refresh()

--localcoord
require('external.script.screenpack')

--"phantom pixel" adjustment to match mugen flipping behavior (extra pixel)
function main.f_alignOffset(align)
	if align == -1 then
		return 1
	end
	return 0
end

main.font = {}
main.font_def = {}

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
-- * motif.setBaseTitleInfo: motif.lua default game mode items assignment
-- * motif.setBaseOptionInfo: motif.lua default option items assignment
-- * motif.setBaseMenuInfo: motif.lua default pause menu items assignment
-- * motif.setBaseTrainingInfo: motif.lua default training menu items assignment
-- * launchFight: start.lua 'launchFight' function (right before match starts)
-- * start.f_selectScreen: start.lua 'f_selectScreen' function (pre layerno=1)
-- * start.f_selectVersus: start.lua 'f_selectVersus' function (pre layerno=1)
-- * start.f_result: start.lua 'f_result' function (pre layerno=1)
-- * start.f_victory: start.lua 'f_victory' function (pre layerno=1)
-- * start.f_continue: start.lua 'f_continue' function (pre layerno=1)
-- * start.f_hiscore: start.lua 'f_hiscore' function (pre layerno=1)
-- * start.f_challenger: start.lua 'f_challenger' function (pre layerno=1)
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

text = {}
color = {}
rect = {}
--create text
function text:create(t)
	local t = t or {}
	t.font = t.font or -1
	t.bank = t.bank or 0
	t.align = t.align or 0
	t.text = t.text or ''
	t.x = t.x or 0
	t.y = t.y or 0
	t.scaleX = t.scaleX or 1
	t.scaleY = t.scaleY or 1
	t.r = t.r or 255
	t.g = t.g or 255
	t.b = t.b or 255
	t.height = t.height or -1
	if t.window == nil then t.window = {} end
	t.window[1] = t.window[1] or 0
	t.window[2] = t.window[2] or 0
	t.window[3] = t.window[3] or motif.info.localcoord[1]
	t.window[4] = t.window[4] or motif.info.localcoord[2]
	t.defsc = t.defsc or false
	t.ti = textImgNew()
	setmetatable(t, self)
	self.__index = self
	if t.font ~= -1 then
		if main.font[t.font .. t.height] == nil then
			--main.f_loadingRefresh(main.txt_loading)
			main.font[t.font .. t.height] = fontNew(t.font, t.height)
			main.f_loadingRefresh(main.txt_loading)
		end
		if main.font_def[t.font .. t.height] == nil then
			main.font_def[t.font .. t.height] = fontGetDef(main.font[t.font .. t.height])
		end
		textImgSetFont(t.ti, main.font[t.font .. t.height])
	end
	textImgSetBank(t.ti, t.bank)
	textImgSetAlign(t.ti, t.align)
	textImgSetText(t.ti, t.text)
	textImgSetColor(t.ti, t.r, t.g, t.b)
	if t.defsc then main.f_disableLuaScale() end
	textImgSetPos(t.ti, t.x + main.f_alignOffset(t.align), t.y)
	textImgSetScale(t.ti, t.scaleX, t.scaleY)
	textImgSetWindow(t.ti, t.window[1], t.window[2], t.window[3] - t.window[1], t.window[4] - t.window[2])
	if t.defsc then main.f_setLuaScale() end
	return t
end

text.new = text.create

--align text
function text:setAlign(align)
	if align:lower() == "left" then
		self.align = -1
	elseif align:lower() == "center" or align:lower() == "middle" then
		self.align = 0
	elseif align:lower() == "right" then
		self.align = 1
	end
	textImgSetAlign(self.ti,self.align)
	return self
end

--update text
function text:update(t)
	if type(t) == "table" then
		local ok = false
		local fontChange = false
		for k, v in pairs(t) do
			if self[k] ~= v then
				if k == 'font' or k == 'height' then
					fontChange = true
				end
				self[k] = v
				ok = true
			end
		end
		if not ok then return end
		if fontChange and self.font ~= -1 then
			if main.font[self.font .. self.height] == nil then
				main.font[self.font .. self.height] = fontNew(self.font, self.height)
			end
			if main.font_def[self.font .. self.height] == nil then
				main.font_def[self.font .. self.height] = fontGetDef(main.font[self.font .. self.height])
			end
			textImgSetFont(self.ti, main.font[self.font .. self.height])
		end
		textImgSetBank(self.ti, self.bank)
		textImgSetAlign(self.ti, self.align)
		textImgSetText(self.ti, self.text)
		textImgSetColor(self.ti, self.r, self.g, self.b)
		if self.defsc then main.f_disableLuaScale() end
		textImgSetPos(self.ti, self.x + main.f_alignOffset(self.align), self.y)
		textImgSetScale(self.ti, self.scaleX, self.scaleY)
		textImgSetWindow(self.ti, self.window[1], self.window[2], self.window[3] - self.window[1], self.window[4] - self.window[2])
		if self.defsc then main.f_setLuaScale() end
	else
		self.text = t
		textImgSetText(self.ti, self.text)
	end

	return self
end

--draw text
function text:draw()
	if self.font == -1 then return end
	textImgDraw(self.ti)
	return self
end

--create color
function color:new(r, g, b, src, dst)
	local n = {r = r or 255, g = g or 255, b = b or 255, src = src or 255, dst = dst or 0}
	setmetatable(n, self)
	self.__index = self
	return n
end

--adds rgb (color + color)
function color.__add(a, b)
	local r = math.max(0, math.min(a.r + b.r, 255))
	local g = math.max(0, math.min(a.g + b.g, 255))
	local b = math.max(0, math.min(a.b + b.b, 255))
	return color:new(r, g, b, a.src, a.dst)
end

--substracts rgb (color - color)
function color.__sub(a, b)
	local r = math.max(0, math.min(a.r - b.r, 255))
	local g = math.max(0, math.min(a.g - b.g, 255))
	local b = math.max(0, math.min(a.b - b.b, 255))
	return color:new(r, g, b, a.src, a.dst)
end

--multiply blend (color * color)
function color.__mul(a, b)
	local r = (a.r / 255) * (b.r / 255) * 255
	local g = (a.g / 255) * (b.g / 255) * 255
	local b = (a.b / 255) * (b.b / 255) * 255
	return color:new(r, g, b, a.src, a.dst)
end

--compares r, g, b, src, and dst (color == color)
function color.__eq(a, b)
	if a.r == b.r and a.g == b.g and a.b == b.b and a.src == b.src and a.dst == b.dst then
		return true
	else
		return false
	end
end

--create color from hex value
function color:fromHex(h)
	h = tostring(h)
	if h:sub(0, 1) =="#" then h = h:sub(2, -1) end
	if h:sub(0, 2) =="0x" then h = h:sub(3, -1) end
	local r = tonumber(h:sub(1, 2), 16)
	local g = tonumber(h:sub(3, 4), 16)
	local b = tonumber(h:sub(5, 6), 16)
	local src = tonumber(h:sub(7, 8), 16) or 255
	local dst = tonumber(h:sub(9, 10), 16) or 0
	return color:new(r, g, b, src, dst)
end

--create string of color converted to hex
function color:toHex(lua)
	local r = string.format("%x", self.r)
	local g = string.format("%x", self.g)
	local b = string.format("%x", self.b)
	local src = string.format("%x", self.src)
	local dst = string.format("%x", self.dst)
	local hex = tostring((r:len() < 2 and "0") .. r .. (g:len() < 2 and "0") .. g .. (b:len() < 2 and "0") .. b ..(src:len() < 2 and "0") .. src .. (dst:len() < 2 and "0") .. dst)
	return hex
end

--returns r, g, b, src, dst
function color:unpack()
	return tonumber(self.r), tonumber(self.g), tonumber(self.b), tonumber(self.src), tonumber(self.dst)
end

--create rect
function rect:create(t)
	local t = t or {}
	t.x1 = t.x1 or 0
	t.y1 = t.y1 or 0
	t.x2 = t.x2 or 0
	t.y2 = t.y2 or 0
	t.color = t.color or color:new(t.r, t.g, t.b, t.src, t.dst)
	t.r, t.g, t.b, t.src, t.dst = t.color:unpack()
	t.defsc = t.defsc or false
	setmetatable(t, self)
	self.__index = self
	return t
end

rect.new = rect.create

--modify rect
function rect:update(t)
	for i, k in pairs(t) do
		self[i] = k
	end
	if t.r or t.g or t.b or t.src or t.dst then
		self.color = color:new(t.r or self.r, t.g or self.g, t.b or self.b, t.src or self.src, t.dst or self.dst)
	end
	return self
end

--draw rect
function rect:draw()
	if self.defsc then main.f_disableLuaScale() end
	fillRect(self.x1, self.y1, self.x2, self.y2, self.r, self.g, self.b, self.src, self.dst)
	if self.defsc then main.f_setLuaScale() end
	return self
end

--create textImg based on usual motif parameters
function main.f_createTextImg(t, prefix, mod)
	local mod = mod or {}
	if t[prefix .. '_font'] == nil then t[prefix .. '_font'] = {} end
	if t[prefix .. '_offset'] == nil then t[prefix .. '_offset'] = {} end
	if t[prefix .. '_scale'] == nil then t[prefix .. '_scale'] = {} end
	return text:create({
		font =   t[prefix .. '_font'][1],
		bank =   t[prefix .. '_font'][2],
		align =  t[prefix .. '_font'][3],
		text =   t[prefix .. '_text'],
		x =      (t[prefix .. '_offset'][1] or 0) + (mod.x or 0),
		y =      (t[prefix .. '_offset'][2] or 0) + (mod.y or 0),
		scaleX = (t[prefix .. '_scale'][1] or 1) * (mod.scaleX or 1),
		scaleY = (t[prefix .. '_scale'][2] or 1) * (mod.scaleY or 1),
		r =      t[prefix .. '_font'][4],
		g =      t[prefix .. '_font'][5],
		b =      t[prefix .. '_font'][6],
		height = t[prefix .. '_font'][7],
		window = t[prefix .. '_window'],
		defsc = mod.defsc or false,
	})
end

--create overlay based on usual motif parameters
function main.f_createOverlay(t, prefix, mod)
	local mod = mod or {}
	if t[prefix .. '_window'] == nil then t[prefix .. '_window'] = {} end
	if t[prefix .. '_col'] == nil then t[prefix .. '_col'] = {} end
	if t[prefix .. '_alpha'] == nil then t[prefix .. '_alpha'] = {} end
	return rect:create({
		x1 =    t[prefix .. '_window'][1],
		y1 =    t[prefix .. '_window'][2],
		x2 =    t[prefix .. '_window'][3] - t[prefix .. '_window'][1] + 1,
		y2 =    t[prefix .. '_window'][4] - t[prefix .. '_window'][2] + 1,
		r =     t[prefix .. '_col'][1],
		g =     t[prefix .. '_col'][2],
		b =     t[prefix .. '_col'][3],
		src =   t[prefix .. '_alpha'][1],
		dst =   t[prefix .. '_alpha'][2],
		defsc = mod.defsc or false,
	})
end

--refreshing screen after delayed animation progression to next frame
main.t_animUpdate = {}
function main.f_refresh()
	for k, v in pairs(main.t_animUpdate) do
		for i = 1, v do
			animUpdate(k)
		end
	end
	main.t_animUpdate = {}
	refresh()
end

--animDraw at specified coordinates
function main.f_animPosDraw(a, x, y, f, instant)
	if a == nil then
		return
	end
	if x ~= nil then animSetPos(a, x, y) end
	if f ~= nil then animSetFacing(a, f) end
	animDraw(a)
	if instant then
		animUpdate(a)
	else
		main.t_animUpdate[a] = 1
	end
end

--screen fade animation
function main.f_fadeAnim(t)
	--draw fade anim
	if main.fadeCnt > 0 then
		if t[main.fadeType .. '_data'] ~= nil then
			animDraw(t[main.fadeType .. '_data'])
			animUpdate(t[main.fadeType .. '_data'])
		end
		main.fadeCnt = main.fadeCnt - 1
	end
	--draw fadein / fadeout
	main.fadeActive = fadeColor(
		main.fadeType,
		main.fadeStart,
		t[main.fadeType .. '_time'],
		t[main.fadeType .. '_col'][1],
		t[main.fadeType .. '_col'][2],
		t[main.fadeType .. '_col'][3]
	)
end

--dynamically adjusts alpha blending each time called based on specified values
local alpha1cur = 0
local alpha2cur = 0
local alpha1add = true
local alpha2add = true
function main.f_boxcursorAlpha(r1min, r1max, r1step, r2min, r2max, r2step)
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
	return alpha1cur, alpha2cur
end

--generate anim from table
function main.f_animFromTable(t, sff, x, y, scaleX, scaleY, facing, infFrame, defsc)
	local t = t or {}
	local x = x or 0
	local y = y or 0
	local scaleX = scaleX or 1.0
	local scaleY = scaleY or 1.0
	local facing = facing or '0'
	local infFrame = infFrame or 1
	local facing_sav = ''
	local anim = ''
	local length = 0
	for i = 1, #t do
		local t_anim = {}
		for j, c in ipairs(main.f_strsplit(',', t[i])) do --split using "," delimiter
			table.insert(t_anim, c)
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
	if defsc then main.f_disableLuaScale() end
	if anim == '' then
		anim = '-1,0, 0,0, -1'
	end
	local data = animNew(sff, anim)
	animSetScale(data, scaleX, scaleY)
	animUpdate(data)
	if defsc then main.f_setLuaScale() end
	return data, length
end

--print array
function main.f_arrayPrint(t)
	print('{' .. table.concat(t, ',') .. '}')
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

--return table with reversed keys
function main.f_tableReverse(t)
	local reversedTable = {}
	local itemCount = #t
	for k, v in ipairs(t) do
		reversedTable[itemCount + 1 - k] = v
	end
	return reversedTable
end

--rotate table elements
function main.f_tableRotate(t, num)
	for i = 1, math.abs(num) do
		if num < 0 then
			table.insert(t, 1, table.remove(t))
		else
			table.insert(t, table.remove(t, 1))
		end
	end
end

--shift table elements
function main.f_tableShift(t, old, new)
	table.insert(t, new, table.remove(t, old))
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
		elseif t1[k] ~= nil and type(t1[k]) ~= type(v) and (not (key or k):match('_font$') --[[or (type(k) == "number" and k > 1)]]) then
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

--return table with proper order and without rows disabled in screenpack
function main.f_tableClean(t, t_sort)
	if t_sort == nil or #t_sort == 0 then
		return t
	end
	local t_clean = {}
	local t_added = {}
	--first we add all entries existing in screenpack file in correct order
	for i = 1, #t_sort do
		for j = 1, #t do
			if t_sort[i] == t[j].itemname and t[j].displayname ~= '' then
				table.insert(t_clean, t[j])
				t_added[t[j].itemname] = 1
				break
			end
		end
	end
	--then we add remaining default entries if not existing yet and not disabled (by default or via screenpack)
	for i = 1, #t do
		if t_sort[t[i].itemname] ~= nil and t_added[t[i].itemname] == nil and t[i].displayname ~= '' then
			table.insert(t_clean, t[i])
		end
	end
	--exception for input menu
	if t[1].itemname == 'empty' and t[#t].itemname == 'page' then
		table.insert(t_clean, 1, t[1])
		table.insert(t_clean, t[#t])
	end
	return t_clean
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

--ensure table existence
function main.f_tableExists(t)
	if t == nil then
		return {}
	end
	return t
end

--initialize table array size
function main.f_tableArray(size, val)
	local t = {}
	for i = 1, size do
		table.insert(t, val or i)
	end
	return t
end

-- append table array right after index having matching key value
function main.f_tableAppendAtKey(t, mKey, nValue)
	for k, v in ipairs(t) do
		if v == mKey then
			table.insert(t, k + 1, nValue)
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

--remove duplicated string pattern
function main.f_uniq(str, pattern, subpattern)
	local out = {}
	for s in str:gmatch(pattern) do
		local s2 = s:match(subpattern)
		if not main.f_tableHasValue(out, s2) then table.insert(out, s) end
	end
	return table.concat(out)
end

--calculates text line length (in pixels) for main.f_textRender
function main.f_lineLength(startX, maxWidth, align, window, windowWrap)
	if window == nil or #window == 0 then
		return 0
	end
	local w = maxWidth
	if windowWrap then
		w = window[3]
	end
	if align == 1 then --left
		return w - startX
	elseif align == 0 then --center
		return main.f_round(math.min(startX - (window[1] or 0), w - startX) * 2)
	else --right
		return startX - (window[1] or 0)
	end
end

--draw string letter by letter + wrap lines. Returns true after finishing rendering last letter.
function main.f_textRender(data, str, counter, x, y, spacingX, spacingY, font_def, delay, length, t_colors)
	if data.font == -1 then return end
	local delay = delay or 0
	local length = length or 0
	local t_colors = t_colors or {}
	str = tostring(str)
	local t = {}
	if length <= 0 then --auto wrapping disabled
		for line in str:gsub('\\n', '\n'):gmatch('([^\r\n]*)[\r\n]?') do
			table.insert(t, line)
		end
	else
		str = str:gsub('\n', '\\n')
		-- for each new line
		for _, line in ipairs(main.f_strsplit('\\n', str)) do --split string using "\n" delimiter
			local text = ''
			local word = ''
			local pxLeft = length
			local word_px = 0
			-- for each character in current line
			for i = 1, string.len(line) do
				local symbol = string.sub(line, i, i)
				-- store symbol length in global table for faster counting
				if font_def[symbol] == nil then
					font_def[symbol] = fontGetTextWidth(main.font[data.font .. data.height], symbol, data.bank)
				end
				local px = (font_def[symbol] + font_def.Spacing[1]) * data.scaleX
				-- continue counting if character fits in the line length
				if pxLeft - px >= 0 or symbol:match('%s') or text == '' then
					-- word valid for line appending on whitespace character (or if it's first word in line)
					if symbol:match('%s') or text == '' then
						text = text .. word .. symbol
						word = ''
						word_px = 0
					-- otherwise add character to the current word
					else
						word = word .. symbol
						word_px = word_px + px
					end
					pxLeft = pxLeft - px
				-- otherwise append current words to table and reset line counting
				else
					table.insert(t, text)
					text = ''
					word = word .. symbol
					word_px = word_px + px
					pxLeft = length - word_px
					word_px = 0
				end
			end
			-- append remaining text in last line
			text = text .. word
			table.insert(t, text)
		end
	end
	-- render text
	local retDone = false
	local retLength = 0
	local lengthCnt = 0
	local subEnd = math.floor(#text - (#text - counter / delay))
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
		elseif i == #t then
			retDone = true
		end
		--TODO: colors support
		--[[if t_colors[subEnd - 1] ~= nil then
			data:update({
				r = t_colors[subEnd - 1].r,
				g = t_colors[subEnd - 1].g,
				b = t_colors[subEnd - 1].b,
			})
		end]]
		data:update({
			text = t[i],
			x = x + spacingX * (i - 1),
			y = y + (main.f_round((font_def.Size[2] + font_def.Spacing[2]) * data.scaleY) + spacingY) * (i - 1),
		})
		data:draw()
		retLength = retLength + string.len(t[i])
	end
	return retDone, retLength
end

--Convert DEF string to table
function main.f_extractText(txt, var1, var2, var3, var4)
	local t = {var1 or '', var2 or '', var3 or '', var4 or ''}
	local str = ''
	--replace %s, %i with variables
	local cnt = 0
	str = txt:gsub('%%([0-9]*)[is]', function(m1)
		cnt = cnt + 1
		if t[cnt] ~= nil then
			if m1 ~= '' then
				while string.len(t[cnt]) < tonumber(m1) do
					t[cnt] = '0' .. t[cnt]
				end
			end
			return t[cnt]
		end
	end)
	--store each line in different row
	t = {}
	str = str:gsub('\n', '\\n')
	for i, c in ipairs(main.f_strsplit('%c?\\n', str)) do --split string using "\n" delimiter
		t[i] = c
	end
	if #t == 0 then
		t[1] = str
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

--y spacing calculation
function main.f_ySpacing(t, key)
	local font_def = main.font_def[t[key .. '_font'][1] .. t[key .. '_font'][7]]
	if font_def == nil then return 0 end
	return main.f_round(font_def.Size[2] * t[key .. '_scale'][2] + font_def.Spacing[2] * t[key .. '_scale'][2])
end

--count occurrences of a substring
function main.f_countSubstring(s1, s2)
    return select(2, s1:gsub(s2, ""))
end

--update rounds to win variables
main.roundsNumSingle = {}
main.roundsNumSimul = {}
main.roundsNumTag = {}
main.maxDrawGames = {}
function main.f_updateRoundsNum()
	for i = 1, 2 do
		if config.RoundsNumSingle == -1 then
			main.roundsNumSingle[i] = getMatchWins(i)
		else
			main.roundsNumSingle[i] = config.RoundsNumSingle
		end
		if config.RoundsNumSimul == -1 then
			main.roundsNumSimul[i] = getMatchWins(i)
		else
			main.roundsNumSimul[i] = config.RoundsNumSimul
		end
		if config.RoundsNumTag == -1 then
			main.roundsNumTag[i] = getMatchWins(i)
		else
			main.roundsNumTag[i] = config.RoundsNumTag
		end
		if config.MaxDrawGames == -2 then
			main.maxDrawGames[i] = getMatchMaxDrawGames(i)
		else
			main.maxDrawGames[i] = config.MaxDrawGames
		end
	end
end

--refresh screen every 0.02 during initial loading
main.nextRefresh = os.clock() + 0.02
function main.f_loadingRefresh(txt)
	if os.clock() >= main.nextRefresh then
		if txt ~= nil then
			txt:draw()
		end
		refresh()
		main.nextRefresh = os.clock() + 0.02
	end
end

--play music
main.lastBgm = ''
function main.f_playBGM(interrupt, bgm, bgmLoop, bgmVolume, bgmLoopstart, bgmLoopend)
	if main.flags['-nomusic'] ~= nil then
		return
	end
	local bgm = bgm or ''
	if interrupt or bgm:gsub('^%./', '') ~= main.lastBgm then
		playBGM(bgm, bgmLoop or 1, bgmVolume or 100, bgmLoopstart or 0, bgmLoopend or 0)
		main.lastBgm = bgm:gsub('^%./', '')
	end
end

main.pauseMenu = false
require('external.script.global')

if main.debugLog then main.f_printTable(main.flags, "debug/flags.txt") end

loadDebugFont(config.DebugFont, config.DebugFontScale)

--;===========================================================
--; COMMAND LINE QUICK VS
--;===========================================================
function main.f_commandLine()
	if main.t_charDef == nil then
		main.t_charDef = {}
	end
	if main.t_stageDef == nil then
		main.t_stageDef = {}
	end
	local ref = #main.f_tableExists(main.t_selChars)
	local t_teamMode = {0, 0}
	local t_numChars = {0, 0}
	local t_matchWins = {single = main.roundsNumSingle, simul = main.roundsNumSimul, tag = main.roundsNumTag, draw = main.maxDrawGames}
	local roundTime = config.RoundTime
	if main.flags['-loadmotif'] == nil then
		loadLifebar(main.lifebarDef)
	end
	setLifebarElements({guardbar = config.BarGuard, stunbar = config.BarStun, redlifebar = config.BarRedLife})
	local frames = framespercount()
	main.f_updateRoundsNum()
	local t = {}
	local t_assignedPals = {}
	for k, v in pairs(main.flags) do
		if k:match('^-p[0-9]+$') then
			local num = tonumber(k:match('^-p([0-9]+)'))
			local player = main.f_playerSide(num)
			t_numChars[player] = t_numChars[player] + 1
			local pal = 1
			if main.flags['-p' .. num .. '.color'] ~= nil or main.flags['-p' .. num .. '.pal'] ~= nil then
				pal = tonumber(main.flags['-p' .. num .. '.color']) or tonumber(main.flags['-p' .. num .. '.pal'])
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
			if main.flags['-p' .. num .. '.ai'] ~= nil then
				ai = tonumber(main.flags['-p' .. num .. '.ai'])
			end
			local input = player
			if main.flags['-p' .. num .. '.input'] ~= nil then
				input = tonumber(main.flags['-p' .. num .. '.input'])
			end
			table.insert(t, {character = v, player = player, num = num, pal = pal, ai = ai, input = input, override = {}})
			if main.flags['-p' .. num .. '.life'] ~= nil then
				t[#t].override['life'] = tonumber(main.flags['-p' .. num .. '.life'])
			end
			if main.flags['-p' .. num .. '.lifeMax'] ~= nil then
				t[#t].override['lifeMax'] = tonumber(main.flags['-p' .. num .. '.lifeMax'])
			end
			if main.flags['-p' .. num .. '.power'] ~= nil then
				t[#t].override['power'] = tonumber(main.flags['-p' .. num .. '.power'])
			end
			if main.flags['-p' .. num .. '.dizzyPoints'] ~= nil then
				t[#t].override['dizzyPoints'] = tonumber(main.flags['-p' .. num .. '.dizzyPoints'])
			end
			if main.flags['-p' .. num .. '.guardPoints'] ~= nil then
				t[#t].override['guardPoints'] = tonumber(main.flags['-p' .. num .. '.guardPoints'])
			end
			if main.flags['-p' .. num .. '.lifeRatio'] ~= nil then
				t[#t].override['lifeRatio'] = tonumber(main.flags['-p' .. num .. '.lifeRatio'])
			end
			if main.flags['-p' .. num .. '.attackRatio'] ~= nil then
				t[#t].override['attackRatio'] = tonumber(main.flags['-p' .. num .. '.attackRatio'])
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
		setAutoguard(i, config.AutoGuard)
	end
	frames = frames * math.max(t_framesMul[1], t_framesMul[2])
	setTimeFramesPerCount(frames)
	setRoundTime(math.max(-1, roundTime * frames))
	local stage = config.StartStage
	if main.flags['-s'] ~= nil then
		for _, v in ipairs({main.flags['-s'], 'stages/' .. main.flags['-s'], 'stages/' .. main.flags['-s'] .. '.def'}) do
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
		main.t_stageDef[stage:lower()] = #main.f_tableExists(main.t_selStages) + 1
	end
	clearSelected()
	setMatchNo(1)
	selectStage(main.t_stageDef[stage:lower()])
	setTeamMode(1, t_teamMode[1], t_numChars[1])
	setTeamMode(2, t_teamMode[2], t_numChars[2])
	if main.debugLog then main.f_printTable(t, 'debug/t_quickvs.txt') end
	--iterate over the table in -p order ascending
	for _, v in main.f_sortKeys(t, function(t, a, b) return t[b].num > t[a].num end) do
		if main.t_charDef[v.character:lower()] == nil then
			if main.flags['-loadmotif'] ~= nil then
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
		overrideCharData(v.player, math.ceil(v.num / 2), v.override)
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
	if main.flags['-ip'] ~= nil then
		enterNetPlay(main.flags['-ip'])
		while not connected() do
			if esc() then
				exitNetPlay()
				os.exit()
			end
			refresh()
		end
		refresh()
		synchronize()
		math.randomseed(sszRandom())
		main.f_cmdBufReset()
		refresh()
	end
	loadStart()
	while loading() do
		--do nothing
	end
	local winner, t_gameStats = game()
	if main.flags['-log'] ~= nil then
		main.f_printTable(t_gameStats, main.flags['-log'])
	end
	os.exit()
end

--initiate quick match only if -loadmotif flag is missing
if main.flags['-p1'] ~= nil and main.flags['-p2'] ~= nil and main.flags['-loadmotif'] == nil then
	main.f_commandLine()
end

--;===========================================================
--; LOAD DATA
--;===========================================================
main.t_unlockLua = {chars = {}, stages = {}, modes = {}}

motif = require('external.script.motif')

main.txt_loading = main.f_createTextImg(motif.title_info, 'loading')
main.txt_loading:draw()
refresh()
loadLifebar(main.lifebarDef)
main.f_loadingRefresh(main.txt_loading)
main.timeFramesPerCount = framespercount()
main.f_updateRoundsNum()

-- generate preload character spr/anim list
local t_preloadList = {}
local function f_preloadList(v)
	if v == nil then
		return
	end
	-- sprite
	if type(v) == 'table' then
		if #v >= 2 and v[1] >= 0 and not t_preloadList[tostring(v[1]) .. ',' .. tostring(v[2])] then
			preloadListChar(v[1], v[2])
			t_preloadList[tostring(v[1]) .. ',' .. tostring(v[2])] = true
		end
	-- anim
	elseif v >= 0 and not t_preloadList[v] then
		preloadListChar(v)
		t_preloadList[v] = true
	end
end
f_preloadList(motif.select_info.portrait_anim)
f_preloadList(motif.select_info.portrait_spr)
f_preloadList(motif.select_info.p1_face_anim)
f_preloadList(motif.select_info.p1_face_spr)
f_preloadList(motif.select_info.p2_face_anim)
f_preloadList(motif.select_info.p2_face_spr)
f_preloadList(motif.select_info.p1_face_done_anim)
f_preloadList(motif.select_info.p1_face_done_spr)
f_preloadList(motif.select_info.p2_face_done_anim)
f_preloadList(motif.select_info.p2_face_done_spr)
f_preloadList(motif.select_info.p1_face2_anim)
f_preloadList(motif.select_info.p1_face2_spr)
f_preloadList(motif.select_info.p2_face2_anim)
f_preloadList(motif.select_info.p2_face2_spr)
f_preloadList(motif.vs_screen.p1_anim)
f_preloadList(motif.vs_screen.p1_spr)
f_preloadList(motif.vs_screen.p2_anim)
f_preloadList(motif.vs_screen.p2_spr)
f_preloadList(motif.vs_screen.p1_done_anim)
f_preloadList(motif.vs_screen.p1_done_spr)
f_preloadList(motif.vs_screen.p2_done_anim)
f_preloadList(motif.vs_screen.p2_done_spr)
f_preloadList(motif.vs_screen.p1_face2_anim)
f_preloadList(motif.vs_screen.p1_face2_spr)
f_preloadList(motif.vs_screen.p2_face2_anim)
f_preloadList(motif.vs_screen.p2_face2_spr)
f_preloadList(motif.victory_screen.p1_anim)
f_preloadList(motif.victory_screen.p1_spr)
f_preloadList(motif.victory_screen.p2_anim)
f_preloadList(motif.victory_screen.p2_spr)
f_preloadList(motif.victory_screen.p1_face2_anim)
f_preloadList(motif.victory_screen.p1_face2_spr)
f_preloadList(motif.victory_screen.p2_face2_anim)
f_preloadList(motif.victory_screen.p2_face2_spr)
f_preloadList(motif.hiscore_info.item_face_anim)
f_preloadList(motif.hiscore_info.item_face_spr)
for i = 1, 2 do
	for _, v in ipairs({{sec = 'select_info', sn = '_face'}, {sec = 'vs_screen', sn = ''}, {sec = 'victory_screen', sn = ''}}) do
		for j = 1, motif[v.sec]['p' .. i .. v.sn .. '_num'] do
			f_preloadList(motif[v.sec]['p' .. i .. '_member' .. j .. v.sn .. '_anim'])
			f_preloadList(motif[v.sec]['p' .. i .. '_member' .. j .. v.sn .. '_spr'])
			f_preloadList(motif[v.sec]['p' .. i .. '_member' .. j .. v.sn .. '_done_anim'])
			f_preloadList(motif[v.sec]['p' .. i .. '_member' .. j .. v.sn .. '_done_spr'])
		end
	end
end

-- generate preload stage spr/anim list
if #motif.select_info.stage_portrait_spr >= 2 and motif.select_info.stage_portrait_spr[1] >= 0 then
	preloadListStage(motif.select_info.stage_portrait_spr[1], motif.select_info.stage_portrait_spr[2])
end
if motif.select_info.stage_portrait_anim >= 0 then
	preloadListStage(motif.select_info.stage_portrait_anim)
end

--warning display
local txt_warning = main.f_createTextImg(motif.warning_info, 'text', {defsc = motif.defaultWarning})
local txt_warningTitle = main.f_createTextImg(motif.warning_info, 'title', {defsc = motif.defaultWarning})
local overlay_warning = main.f_createOverlay(motif.warning_info, 'overlay')
function main.f_warning(t, background, info, title, txt, overlay)
	local info = info or motif.warning_info
	local title = title or txt_warningTitle
	local txt = txt or txt_warning
	local overlay = overlay or overlay_warning
	local cancel_snd = info.cancel_snd or motif.warning_info.cancel_snd
	local done_snd = info.done_snd or motif.warning_info.done_snd
	resetKey()
	esc(false)
	while true do
		main.f_cmdInput()
		if esc() or main.f_input(main.t_players, {'m'}) then
			sndPlay(motif.files.snd_data, cancel_snd[1], cancel_snd[2])
			return false
		elseif getKey() ~= '' then
			sndPlay(motif.files.snd_data, done_snd[1], done_snd[2])
			resetKey()
			return true
		end
		--draw clearcolor
		clearColor(background.bgclearcolor[1], background.bgclearcolor[2], background.bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(background.bg, false)
		--draw overlay
		overlay:draw()
		--draw title
		title:draw()
		--draw text
		for i = 1, #t do
			txt:update({
				text = t[i],
				y = info.text_offset[2] + main.f_ySpacing(info, 'text') * (i - 1),
			})
			txt:draw()
		end
		--draw layerno = 1 backgrounds
		bgDraw(background.bg, true)
		--end loop
		refresh()
	end
end

--input display
local txt_textinput = main.f_createTextImg(motif.title_info, 'textinput')
local overlay_textinput = main.f_createOverlay(motif.title_info, 'textinput_overlay')
function main.f_drawInput(t, txt, overlay, offsetY, spacingY, background, category, controllerNo, keyBreak)
	local category = category or 'string'
	local controllerNo = controllerNo or 0
	local keyBreak = keyBreak or ''
	if category == 'string' then
		table.insert(t, '')
	end
	local input = ''
	local btnReleased = 0
	resetKey()
	while true do
		if esc() --[[or main.f_input(main.t_players, {'m'})]] then
			input = ''
			break
		end
		if category == 'keyboard' then
			input = getKey()
			if input ~= '' then
				break
			end
		elseif category == 'gamepad' then
			if getJoystickPresent(controllerNo) == false then
				break
			end
			if getKey() == keyBreak then
				input = keyBreak
				break
			end
			local tmp = getKey()
			if tonumber(tmp) == nil then --button released
				if btnReleased == 0 then
					btnReleased = 1
				elseif btnReleased == 2 then
					break
				end
			elseif btnReleased == 1 then --button pressed after releasing button once
				input = tmp
				btnReleased = 2
			end
		else --string
			if getKey('RETURN') then
				break
			elseif getKey('BACKSPACE') then
				input = input:match('^(.-).?$')
			else
				input = input .. getKeyText()
			end
			t[#t] = input
			resetKey()
		end
		--draw clearcolor
		clearColor(background.bgclearcolor[1], background.bgclearcolor[2], background.bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(background.bg, false)
		--draw overlay
		overlay:draw()
		--draw text
		for i = 1, #t do
			txt:update({
				text = t[i],
				y = offsetY + spacingY * (i - 1),
			})
			txt:draw()
		end
		--draw layerno = 1 backgrounds
		bgDraw(background.bg, true)
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
	elseif c:match('^music') then --musicX / musiclife / musicvictory
		local bgmvolume, bgmloopstart, bgmloopend = 100, 0, 0
		c = c:gsub('%s+([0-9%s]+)$', function(m1)
			for i, c in ipairs(main.f_strsplit('%s+', m1)) do --split using whitespace delimiter
				if i == 1 then
					bgmvolume = tonumber(c)
				elseif i == 2 then
					bgmloopstart = tonumber(c)
				elseif i == 3 then
					bgmloopend = tonumber(c)
				else
					break
				end
			end
			return ''
		end)
		c = c:gsub('\\', '/')
		local bgtype, round, bgmusic = c:match('^(music[a-z]*)([0-9]*)%s*=%s*(.-)%s*$')
		if t[bgtype] == nil then t[bgtype] = {} end
		local t_ref = t[bgtype]
		if bgtype == 'music' or round ~= '' then
			round = tonumber(round) or 1
			if t[bgtype][round] == nil then t[bgtype][round] = {} end
			t_ref = t[bgtype][round]
		end
		table.insert(t_ref, {bgmusic = bgmusic, bgmvolume = bgmvolume, bgmloopstart = bgmloopstart, bgmloopend = bgmloopend})
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

main.dummySff = sffNew()
function main.f_addChar(line, playable, loading, slot)
	table.insert(main.t_selChars, {})
	local row = #main.t_selChars
	local slot = slot or false
	local valid = false
	--store 'unlock' param and get rid of everything that follows it
	local unlock = ''
	line = line:gsub(',%s*unlock%s*=%s*(.-)s*$', function(m1)
		unlock = m1
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
			addChar(c)
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
						main.t_selChars[row][v] = searchFile(main.t_selChars[row][v], {main.t_selChars[row].dir, '', motif.fileDir, 'data/'})
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
		for _, v in pairs({{motif.select_info.portrait_anim, -1}, motif.select_info.portrait_spr}) do
			if v[1] ~= -1 then
				main.t_selChars[row].cell_data = animGetPreloadedData('char', main.t_selChars[row].char_ref, v[1], v[2])
				if main.t_selChars[row].cell_data ~= nil then
					animSetScale(
						main.t_selChars[row].cell_data,
						motif.select_info.portrait_scale[1] * main.t_selChars[row].portrait_scale / (main.SP_Viewport43[3] / main.SP_Localcoord[1]),
						motif.select_info.portrait_scale[2] * main.t_selChars[row].portrait_scale / (main.SP_Viewport43[3] / main.SP_Localcoord[1]),
						false
					)
					animUpdate(main.t_selChars[row].cell_data)
					break
				end
			end
		end
		if main.t_selChars[row].cell_data == nil then
			main.t_selChars[row].cell_data = animNew(main.dummySff, '-1,0, 0,0, -1')
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
			main.t_selChars[row][v] = main.t_selChars[row][v]:gsub('/(.)%s*+', '/%1,') --convert '+' to ',' for button holding
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
		main.f_loadingRefresh(main.txt_loading)
	end
	return valid
end

function main.f_addStage(file, hidden)
	file = file:gsub('\\', '/')
	if file:match('/$') then
		return
	end
	if addStage(file) == 0 then
		return
	end
	local stageNo = #main.t_selStages + 1
	local t_info = getStageInfo(stageNo)
	table.insert(main.t_selStages, {
		name = t_info.name,
		def = file,
		dir = t_info.def:gsub('[^/]+%.def$', ''),
		portrait_scale = t_info.portrait_scale,
	})
	--attachedchar
	if t_info.attachedchardef ~= '' then
		main.t_selStages[stageNo].attachedChar = getCharAttachedInfo(t_info.attachedchardef)
		if main.t_selStages[stageNo].attachedChar ~= nil then
			main.t_selStages[stageNo].attachedChar.dir = main.t_selStages[stageNo].attachedChar.def:gsub('[^/]+%.def$', '')
		end
	end
	--music
	for k, v in pairs(t_info.stagebgm) do
		if k:match('^bgmusic') or k:match('^bgmvolume') or k:match('^bgmloop') then
			if t_info.stagebgm[k] ~= '' then
				local prefix, dot, suffix, round = k:match('^([^%.]+)(%.?)([A-Za-z]*)([0-9]*)$')
				local bgtype = 'music' .. suffix
				if suffix == '' or suffix == 'round' then
					bgtype = 'music'
					round = tonumber(round) or 1
				end
				if main.t_selStages[stageNo][bgtype] == nil then main.t_selStages[stageNo][bgtype] = {} end
				local t_ref = main.t_selStages[stageNo][bgtype]
				if bgtype == 'music' then
					if main.t_selStages[stageNo][bgtype][round] == nil then main.t_selStages[stageNo][bgtype][round] = {} end
					t_ref = main.t_selStages[stageNo][bgtype][round]
				end
				if #t_ref == 0 then
					table.insert(t_ref, {bgmusic = '', bgmvolume = 100, bgmloopstart = 0, bgmloopend = 0})
				end
				if k:match('^bgmusic') then
					t_ref[1][prefix] = searchFile(tostring(v), {file, "", "data/", "sound/"})
				elseif tonumber(v) then
					t_ref[1][prefix] = tonumber(v)
				end
			end
		elseif v ~= '' then
			main.t_selStages[stageNo][k:gsub('%.', '_')] = main.f_dataType(v)
		end
	end
	main.t_stageDef[file:lower()] = stageNo
	--anim data
	for _, v in pairs({{motif.select_info.stage_portrait_anim, -1}, motif.select_info.stage_portrait_spr}) do
		if #v > 0 and v[1] ~= -1 then
			main.t_selStages[stageNo].anim_data = animGetPreloadedData('stage', stageNo, v[1], v[2])
			if main.t_selStages[stageNo].anim_data ~= nil then
				animSetScale(
					main.t_selStages[stageNo].anim_data,
					motif.select_info.stage_portrait_scale[1] * main.t_selStages[stageNo].portrait_scale / (main.SP_Viewport43[3] / main.SP_Localcoord[1]),
					motif.select_info.stage_portrait_scale[2] * main.t_selStages[stageNo].portrait_scale / (main.SP_Viewport43[3] / main.SP_Localcoord[1]),
					false
				)
				animSetWindow(
					main.t_selStages[stageNo].anim_data,
					motif.select_info.stage_portrait_window[1],
					motif.select_info.stage_portrait_window[2],
					motif.select_info.stage_portrait_window[3],
					motif.select_info.stage_portrait_window[4]
				)
				animUpdate(main.t_selStages[stageNo].anim_data)
				break
			end
		end
	end
	if hidden ~= nil and hidden ~= 0 then
		main.t_selStages[stageNo].hidden = hidden
	end
	if main.t_selStages[stageNo].anim_data == nil then
		main.t_selStages[stageNo].anim_data = animNew(main.dummySff, '-1,0, 0,0, -1')
	end
	return stageNo
end

main.t_includeStage = {{}, {}} --includestage = 1, includestage = -1
main.t_orderChars = {}
main.t_orderStages = {}
main.t_orderSurvival = {}
main.t_bonusChars = {}
main.t_stageDef = {['random'] = 0}
main.t_charDef = {}
main.t_selChars = {}
main.t_selGrid = {}
main.t_selStages = {}
main.t_selOptions = {}
main.t_selStoryMode = {}
local t_storyModeList = {}
local t_addExluded = {}
local tmp = ''
local section = 0
local row = 0
local slot = false
local content = main.f_fileRead(motif.files.select)
local csCell = 0
content = content:gsub('([^\r\n;]*)%s*;[^\r\n]*', '%1')
content = content:gsub('\n%s*\n', '\n')
for line in content:gmatch('[^\r\n]+') do
--for line in io.lines("data/select.def") do
	local lineCase = line:lower()
	if lineCase:match('^%s*%[%s*characters%s*%]') then
		row = 0
		section = 1
	elseif lineCase:match('^%s*%[%s*extrastages%s*%]') then
		row = 0
		section = 2
	elseif lineCase:match('^%s*%[%s*options%s*%]') then
		main.t_selOptions = {
			arcadestart = {wins = 0, offset = 0},
			arcadeend = {wins = 0, offset = 0},
			teamstart = {wins = 0, offset = 0},
			teamend = {wins = 0, offset = 0},
			survivalstart = {wins = 0, offset = 0},
			survivalend = {wins = 0, offset = 0},
			ratiostart = {wins = 0, offset = 0},
			ratioend = {wins = 0, offset = 0},
		}
		row = 0
		section = 3
	elseif lineCase:match('^%s*%[%s*storymode%s*%]') then
		row = 0
		section = 4
	elseif lineCase:match('^%s*%[%w+%]$') then
		section = -1
	elseif section == 1 then --[Characters]
		local csCol = (csCell % motif.select_info.columns) + 1
		local csRow = math.floor(csCell / motif.select_info.columns) + 1
		while not slot and motif.select_info['cell_' .. csCol .. '_' .. csRow .. '_skip'] == 1 do
			main.f_addChar('skipslot', true, true, false)
			csCell = csCell + 1
			csCol = (csCell % motif.select_info.columns) + 1
			csRow = math.floor(csCell / motif.select_info.columns) + 1
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
		line = line:gsub(',%s*unlock%s*=%s*(.-)s*$', function(m1)
			unlock = m1
			hidden = 1
			return ''
		end)
		--parse rest of the line
		for i, c in ipairs(main.f_strsplit(',', line)) do --split using "," delimiter
			c = c:gsub('^%s*(.-)%s*$', '%1')
			if i == 1 then
				row = main.f_addStage(c, hidden)
				if row == nil then
					break
				end
				table.insert(main.t_includeStage[1], row)
				table.insert(main.t_includeStage[2], row)
			elseif c:match('^music') then --musicX / musiclife / musicvictory
				local bgmvolume, bgmloopstart, bgmloopend = 100, 0, 0
				c = c:gsub('%s+([0-9%s]+)$', function(m1)
					for i, c in ipairs(main.f_strsplit('%s+', m1)) do --split using whitespace delimiter
						if i == 1 then
							bgmvolume = tonumber(c)
						elseif i == 2 then
							bgmloopstart = tonumber(c)
						elseif i == 3 then
							bgmloopend = tonumber(c)
						else
							break
						end
					end
					return ''
				end)
				c = c:gsub('\\', '/')
				local bgtype, round, bgmusic = c:match('^(music[a-z]*)([0-9]*)%s*=%s*(.-)%s*$')
				if main.t_selStages[row][bgtype] == nil then main.t_selStages[row][bgtype] = {} end
				local t_ref = main.t_selStages[row][bgtype]
				if bgtype == 'music' or round ~= '' then
					round = tonumber(round) or 1
					if main.t_selStages[row][bgtype][round] == nil then main.t_selStages[row][bgtype][round] = {} end
					t_ref = main.t_selStages[row][bgtype][round]
				end
				table.insert(t_ref, {bgmusic = bgmusic, bgmvolume = bgmvolume, bgmloopstart = bgmloopstart, bgmloopend = bgmloopend})
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
			--default order
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
					main.f_warning(main.f_extractText(motif.warning_info.text_ratio_text), motif.titlebgdef)
					main.t_selOptions[rowName .. 'ratiomatches'] = nil
					break
				end
				if rmax == '' then
					rmax = rmin
				end
				table.insert(main.t_selOptions[rowName .. 'ratiomatches'], {rmin = rmin, rmax = rmax, order = order})
			end
		elseif lineCase:match('%.airamp%.') then
			local rowName, rowName2, wins, offset = lineCase:match('^%s*(.-)%.airamp%.(.-)%s*=%s*([%.0-9-]+)%s*,%s*([%.0-9-]+)')
			main.t_selOptions[rowName .. rowName2] = {wins = tonumber(wins), offset = tonumber(offset)}
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
if config.TrainingChar ~= '' and main.t_charDef[config.TrainingChar:lower()] == nil then
	main.f_addChar(config.TrainingChar .. ', order = 0, ordersurvival = 0, exclude = 1', false, true)
end

--add remaining character parameters
main.t_randomChars = {}
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
	--if character's name has been stored
	if main.t_selChars[i].name ~= nil then
		--generate table with characters allowed to be randomly selected
		if main.t_selChars[i].playable and (main.t_selChars[i].hidden == nil or main.t_selChars[i].hidden <= 1) and (main.t_selChars[i].exclude == nil or main.t_selChars[i].exclude == 0) then
			table.insert(main.t_randomChars, i - 1)
		end
	end
end

--add default starting stage if no stages have been added via select.def
if #main.t_includeStage[1] == 0 or #main.t_includeStage[2] == 0 then
	local row = main.f_addStage(config.StartStage)
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
function main.f_menuWindow(t)
	if t.menu_window_margins_y[1] ~= 0 or t.menu_window_margins_y[2] ~= 0 then
		return {
			0,
			math.max(0, t.menu_pos[2] - t.menu_window_margins_y[1]),
			motif.info.localcoord[1],
			t.menu_pos[2] + (t.menu_window_visibleitems - 1) * t.menu_item_spacing[2] + t.menu_window_margins_y[2]
		}
	end
	return {0, 0, main.SP_Localcoord[1], math.max(240, main.SP_Localcoord[2])}
end

--Load additional scripts
start = require('external.script.start')
randomtest = require('external.script.randomtest')
options = require('external.script.options')
storyboard = require('external.script.storyboard')
menu = require('external.script.menu')

if main.flags['-storyboard'] ~= nil then
	storyboard.f_storyboard(main.flags['-storyboard'])
	os.exit()
end

--;===========================================================
--; MENUS
--;===========================================================
if motif.attract_mode.enabled == 1 then
	main.group = 'attract_mode'
	main.background = 'attractbgdef'
else
	main.group = 'title_info'
	main.background = 'titlebgdef'
end

main.txt_title = main.f_createTextImg(motif[main.group], 'title')
main.txt_mainSelect = main.f_createTextImg(motif.select_info, 'title')
local t_footer = {}
if motif.attract_mode.enabled == 0 then
	for i = 1, 3 do
		table.insert(t_footer, main.f_createTextImg(motif.title_info, 'footer' .. i))
	end
end

local txt_infoboxTitle = main.f_createTextImg(motif.infobox, 'title')
local txt_infobox = main.f_createTextImg(motif.infobox, 'text')
local overlay_infobox = main.f_createOverlay(motif.infobox, 'overlay')
local overlay_footer = main.f_createOverlay(motif.title_info, 'footer_overlay')

function main.f_default()
	for i = 1, config.Players do
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
	main.continueScreen = false --if continue screen should be shown
	main.coop = false --if mode should be recognized as coop
	main.cpuSide = {false, true} --which side is controlled by CPU
	if motif.attract_mode.enabled == 0 and start.challenger == 0 then
		main.credits = -1 --amount of credits from the start (-1 = disabled)
	end
	main.dropDefeated = false --if defeated members should be removed from team
	main.elimination = false --if single lose should stop further lua execution
	main.exitSelect = false --if "clearing" the mode (matchno == -1) should go back to main menu
	main.forceChar = {nil, nil} --predefined P1/P2 characters
	main.forceRosterSize = false --if roster size should be enforced even if there are not enough characters to fill it (not used but may be useful for external modules)
	main.hiscoreScreen = false --if hiscore screen should be shown
	main.lifebar = { --which lifebar elements should be rendered
		active = true,
		bars = true,
		match = false,
		mode = true,
		p1aiLevel = false,
		p1score = false,
		p1winCount = false,
		p2aiLevel = false,
		p2score = false,
		p2winCount = false,
		timer = false,
		guardbar = config.BarGuard,
		stunbar = config.BarStun,
		redlifebar = config.BarRedLife,
		hidebars = motif.dialogue_info.enabled == 1,
	}
	main.lifePersistence = false --if life should be maintained after match
	main.luaPath = 'external/script/default.lua' --path to script executed by start.f_selectMode()
	main.makeRoster = false --if default roster for each match should be generated before first match
	main.matchWins = { --amount of rounds to win for each team side and team mode
		draw = main.maxDrawGames,
		simul = main.roundsNumSimul,
		single = main.roundsNumSingle,
		tag = main.roundsNumTag,
	}
	main.numSimul = {config.NumSimul[1], config.NumSimul[2]} --min/max number of simul characters
	main.numTag = {config.NumTag[1], config.NumTag[2]} --min/max number of tag characters
	main.numTurns = {config.NumTurns[1], config.NumTurns[2]} --min/max number of turn characters
	main.orderSelect = {false, false} --if versus screen order selection should be active
	main.quickContinue = false --if by default continuing should skip player selection
	main.rankingCondition = false --if winning (clearing) whole mode is needed for rankings to be saved
	main.resetScore = false --if loosing should set score for the next match to lose count
	main.resultsTable = nil --which motif section should be used for result screen rendering
	main.rotationChars = false --flags modes where config.AISurvivalColor should be used instead of config.AIRandomColor
	main.roundTime = config.RoundTime --sets round time
	main.selectMenu = {true, false} --which team side should be allowed to select players
	main.stageMenu = false --if manual stage selection is allowed
	main.stageOrder = false --if select.def stage order param should be used
	main.storyboard = {intro = false, ending = false, credits = false, gameover = false} --which storyboards should be active
	main.teamMenu = {
		{ratio = false, simul = false, single = false, tag = false, turns = false}, --which team modes should be selectable by P1 side
		{ratio = false, simul = false, single = false, tag = false, turns = false}, --which team modes should be selectable by P2 side
	}
	main.versusScreen = false --if versus screen should be shown
	main.versusMatchNo = false --if versus screen should render screenpack match element
	main.victoryScreen = false --if victory screen should be shown
	resetAILevel()
	resetRemapInput()
	setAutoguard(1, config.AutoGuard)
	setAutoguard(2, config.AutoGuard)
	setAutoLevel(false)
	setConsecutiveWins(1, 0)
	setConsecutiveWins(2, 0)
	setConsecutiveRounds(false)
	setContinue(false)
	setGameMode('')
	setHomeTeam(2) --http://mugenguild.com/forum/topics/ishometeam-triggers-169132.0.html
	setLifebarElements(main.lifebar)
	setRoundTime(math.max(-1, main.roundTime * main.timeFramesPerCount))
	setTimeFramesPerCount(main.timeFramesPerCount)
	setWinCount(1, 0)
	setWinCount(2, 0)
	main.txt_mainSelect:update({text = ''})
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
		main.continueScreen = true
		main.exitSelect = true
		main.hiscoreScreen = true
		--main.lifebar.p1score = true
		--main.lifebar.p2aiLevel = true
		main.makeRoster = true
		main.orderSelect[1] = true
		main.orderSelect[2] = true
		main.resetScore = true
		main.resultsTable = motif.win_screen
		main.stageOrder = true
		main.storyboard.credits = true
		main.storyboard.ending = true
		main.storyboard.gameover = true
		main.storyboard.intro = true
		if (t ~= nil and t[item].itemname == 'arcade') or (t == nil and not main.teamarcade) then
			main.teamMenu[1].single = true
			main.teamMenu[2].single = true
			main.txt_mainSelect:update({text = motif.select_info.title_arcade_text})
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
			main.txt_mainSelect:update({text = motif.select_info.title_teamarcade_text})
			main.teamarcade = true
		end
		main.versusScreen = true
		main.versusMatchNo = true
		main.victoryScreen = true
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
		main.txt_mainSelect:update({text = motif.select_info.title_bonus_text})
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
		--main.lifebar.p2aiLevel = true
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
		main.versusScreen = true
		main.victoryScreen = true
		main.txt_mainSelect:update({text = motif.select_info.title_freebattle_text})
		setGameMode('freebattle')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--JOIN (NEW ADDRESS)
	['joinadd'] = function(t, item)
		sndPlay(motif.files.snd_data, motif[main.group].cursor_move_snd[1], motif[main.group].cursor_move_snd[2])
		local name = main.f_drawInput(
			main.f_extractText(motif.title_info.textinput_name_text),
			txt_textinput,
			overlay_textinput,
			motif[main.group].textinput_offset[2],
			main.f_ySpacing(motif.title_info, 'textinput'),
			motif[main.background]
		)
		if name ~= '' then
			sndPlay(motif.files.snd_data, motif[main.group].cursor_move_snd[1], motif[main.group].cursor_move_snd[2])
			local address = main.f_drawInput(
				main.f_extractText(motif.title_info.textinput_address_text),
				txt_textinput,
				overlay_textinput,
				motif[main.group].textinput_offset[2],
				main.f_ySpacing(motif.title_info, 'textinput'),
				motif[main.background]
			)
			if address:match('^[0-9%.]+$') then
				sndPlay(motif.files.snd_data, motif[main.group].cursor_done_snd[1], motif[main.group].cursor_done_snd[2])
				config.IP[name] = address
				table.insert(t, #t, {data = text:create({}), itemname = 'ip_' .. name, displayname = name})
				main.f_fileWrite(main.flags['-config'], json.encode(config, {indent = 2}))
			else
				sndPlay(motif.files.snd_data, motif[main.group].cancel_snd[1], motif[main.group].cancel_snd[2])
			end
		else
			sndPlay(motif.files.snd_data, motif[main.group].cancel_snd[1], motif[main.group].cancel_snd[2])
		end
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
		--main.lifebar.p2aiLevel = true
		main.lifePersistence = true
		main.makeRoster = true
		main.matchWins.draw = {0, 0}
		main.matchWins.simul = {1, 1}
		main.matchWins.single = {1, 1}
		main.matchWins.tag = {1, 1}
		main.numSimul = {2, 2}
		main.numTag = {2, 2}
		main.resultsTable = motif.survival_results_screen
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
		main.txt_mainSelect:update({text = motif.select_info.title_netplaysurvivalcoop_text})
		setConsecutiveRounds(true)
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
		main.continueScreen = true
		main.coop = true
		main.exitSelect = true
		--main.lifebar.p1score = true
		--main.lifebar.p2aiLevel = true
		main.makeRoster = true
		main.numSimul = {2, 2}
		main.numTag = {2, 2}
		main.resetScore = true
		main.resultsTable = motif.win_screen
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
		main.versusScreen = true
		main.versusMatchNo = true
		main.victoryScreen = true
		main.f_setCredits()
		main.txt_mainSelect:update({text = motif.select_info.title_netplayteamcoop_text})
		setGameMode('netplayteamcoop')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--NETPLAY VERSUS
	['netplayversus'] = function()
		setHomeTeam(1)
		main.cpuSide[2] = false
		--main.lifebar.p1winCount = true
		--main.lifebar.p2winCount = true
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
		main.versusScreen = true
		main.victoryScreen = true
		main.txt_mainSelect:update({text = motif.select_info.title_netplayversus_text})
		setGameMode('netplayversus')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--OPTIONS
	['options'] = function()
		hook.run("main.t_itemname")
		return options.menu.loop
	end,
	--RANDOMTEST
	['randomtest'] = function()
		setGameMode('randomtest')
		hook.run("main.t_itemname")
		return randomtest.run
	end,
	--REPLAY
	['replay'] = function()
		return main.f_replay
	end,
	--SERVER CONNECT
	['serverconnect'] = function(t, item)
		if main.f_connect(config.IP[t[item].displayname], main.f_extractText(motif.title_info.connecting_join_text, t[item].displayname, config.IP[t[item].displayname])) then
			synchronize()
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
		if main.f_connect("", main.f_extractText(motif.title_info.connecting_host_text, getListenPort())) then
			synchronize()
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
		main.continueScreen = true
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
		main.hiscoreScreen = true
		--main.lifebar.match = true
		--main.lifebar.p2aiLevel = true
		main.lifePersistence = true
		main.makeRoster = true
		main.matchWins.draw = {0, 0}
		main.matchWins.simul = {1, 1}
		main.matchWins.single = {1, 1}
		main.matchWins.tag = {1, 1}
		main.orderSelect[1] = true
		main.orderSelect[2] = true
		main.resultsTable = motif.survival_results_screen
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
		main.txt_mainSelect:update({text = motif.select_info.title_survival_text})
		setConsecutiveRounds(true)
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
		main.hiscoreScreen = true
		--main.lifebar.match = true
		--main.lifebar.p2aiLevel = true
		main.lifePersistence = true
		main.makeRoster = true
		main.matchWins.draw = {0, 0}
		main.matchWins.simul = {1, 1}
		main.matchWins.single = {1, 1}
		main.matchWins.tag = {1, 1}
		main.numSimul = {2, math.min(4, config.Players)}
		main.numTag = {2, math.min(4, config.Players)}
		main.resultsTable = motif.survival_results_screen
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
		main.txt_mainSelect:update({text = motif.select_info.title_survivalcoop_text})
		setConsecutiveRounds(true)
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
		main.continueScreen = true
		main.coop = true
		main.exitSelect = true
		main.hiscoreScreen = true
		--main.lifebar.p1score = true
		--main.lifebar.p2aiLevel = true
		main.makeRoster = true
		main.numSimul = {2, math.min(4, config.Players)}
		main.numTag = {2, math.min(4, config.Players)}
		main.resetScore = true
		main.resultsTable = motif.win_screen
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
		main.versusScreen = true
		main.versusMatchNo = true
		main.victoryScreen = true
		main.f_setCredits()
		main.txt_mainSelect:update({text = motif.select_info.title_teamcoop_text})
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
		main.continueScreen = true
		main.exitSelect = true
		main.hiscoreScreen = true
		--main.lifebar.p2aiLevel = true
		--main.lifebar.timer = true
		main.makeRoster = true
		main.quickContinue = true
		main.orderSelect[1] = true
		main.orderSelect[2] = true
		main.resetScore = true
		main.resultsTable = motif.time_attack_results_screen
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
		main.versusScreen = true
		main.versusMatchNo = true
		main.f_setCredits()
		main.txt_mainSelect:update({text = motif.select_info.title_timeattack_text})
		setGameMode('timeattack')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--TRAINING
	['training'] = function()
		setHomeTeam(1)
		main.f_playerInput(main.playerInput, 1)
		main.t_pIn[2] = 1
		if main.t_charDef[config.TrainingChar:lower()] ~= nil then
			main.forceChar[2] = {main.t_charDef[config.TrainingChar:lower()]}
		end
		--main.lifebar.p1score = true
		--main.lifebar.p2aiLevel = true
		main.roundTime = -1
		main.selectMenu[2] = true
		main.stageMenu = true
		main.teamMenu[1].ratio = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].single = true
		main.teamMenu[1].tag = true
		main.teamMenu[1].turns = true
		main.teamMenu[2].single = true
		main.txt_mainSelect:update({text = motif.select_info.title_training_text})
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
		--main.lifebar.p1winCount = true
		--main.lifebar.p2winCount = true
		main.orderSelect[1] = true
		main.orderSelect[2] = true
		main.selectMenu[2] = true
		main.stageMenu = true
		if (start.challenger == 0 and t[item].itemname == 'versus') or (start.challenger ~= 0 and not main.teamarcade) then
			main.teamMenu[1].single = true
			main.teamMenu[2].single = true
			main.txt_mainSelect:update({text = motif.select_info.title_versus_text})
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
			main.txt_mainSelect:update({text = motif.select_info.title_teamversus_text})
		end
		main.versusScreen = true
		main.victoryScreen = true
		setGameMode('versus')
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
		--main.lifebar.p1winCount = true
		--main.lifebar.p2winCount = true
		main.numSimul = {2, math.min(4, math.max(2, math.ceil(config.Players / 2)))}
		main.numTag = {2, math.min(4, math.max(2, math.ceil(config.Players / 2)))}
		main.selectMenu[2] = true
		main.stageMenu = true
		main.teamMenu[1].simul = true
		main.teamMenu[1].tag = true
		main.teamMenu[2].simul = true
		main.teamMenu[2].tag = true
		main.versusScreen = true
		main.victoryScreen = true
		main.txt_mainSelect:update({text = motif.select_info.title_versuscoop_text})
		setGameMode('versuscoop')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
	--WATCH
	['watch'] = function()
		main.f_playerInput(main.playerInput, 1)
		main.t_pIn[2] = 1
		main.cpuSide[1] = true
		--main.lifebar.p1aiLevel = true
		--main.lifebar.p2aiLevel = true
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
		main.versusScreen = true
		main.victoryScreen = true
		main.txt_mainSelect:update({text = motif.select_info.title_watch_text})
		setGameMode('watch')
		hook.run("main.t_itemname")
		return start.f_selectMode
	end,
}
main.t_itemname.teamarcade = main.t_itemname.arcade
main.t_itemname.teamversus = main.t_itemname.versus
if main.debugLog then main.f_printTable(main.t_itemname, 'debug/t_mainItemname.txt') end

function main.f_deleteIP(item, t)
	if t[item].itemname:match('^ip_') then
		sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
		resetKey()
		config.IP[t[item].itemname:gsub('^ip_', '')] = nil
		main.f_fileWrite(main.flags['-config'], json.encode(config, {indent = 2}))
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

main.fadeActive = false
local demoFrameCounter = 0
local introWaitCycles = 0
-- Shared menu loop logic
function main.f_createMenu(tbl, bool_bgreset, bool_main, bool_f1, bool_del)
	return function()
		hook.run("main.menu.loop")
		local cursorPosY = 1
		local moveTxt = 0
		local item = 1
		local t = main.f_hiddenItems(tbl.items)
		--skip showing menu if there is only 1 valid item
		local cnt = 0
		local f = ''
		for _, v in ipairs(tbl.items) do
			if tbl.name == 'bonusgames' --[[or tbl.name == 'storymode']] or v.itemname == 'joinadd' then
				skip = true
				break
			elseif v.itemname ~= 'back' and main.t_unlockLua.modes[v.itemname] == nil then
				f = v.itemname
				if main.t_itemname[f] == nil and t_storyModeList[f] then
					f = 'storyarc'
				end
				cnt = cnt + 1
			end
		end
		if main.t_itemname[f] ~= nil and cnt == 1 --[[and motif.attract_mode.enabled == 0]] then
			main.f_default()
			main.menu.f = main.t_itemname[f](t, item)
			main.f_unlock(false)
			main.menu.f()
			main.f_default()
			main.f_unlock(false)
			local itemNum = #t
			t = main.f_hiddenItems(tbl.items)
			main.menu.f = nil
			if itemNum == #t then
				return
			end
		end
		--more than 1 item, continue loop
		if bool_main then
			if motif.files.logo_storyboard ~= '' then
				storyboard.f_storyboard(motif.files.logo_storyboard)
			end
			if motif.files.intro_storyboard ~= '' then
				storyboard.f_storyboard(motif.files.intro_storyboard)
			end
		end
		if bool_bgreset then
			if motif.attract_mode.enabled == 0 then
				main.f_bgReset(motif[main.background].bg)
				main.f_playBGM(false, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
			end
			main.f_fadeReset('fadein', motif[main.group])
		end
		main.menu.f = nil
		while true do
			if tbl.reset then
				tbl.reset = false
				main.f_cmdInput()
			else
				main.f_menuCommonDraw(t, item, cursorPosY, moveTxt, main.group, main.background, main.txt_title, false, t_footer)
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
				cursorPosY, moveTxt, item = main.f_menuCommonCalc(t, item, cursorPosY, moveTxt, main.group, main.f_extractKeys(motif[main.group].menu_previous_key), main.f_extractKeys(motif[main.group].menu_next_key))
				main.txt_title:update({text = tbl.title})
				if item_sav ~= item then
					demoFrameCounter = 0
					introWaitCycles = 0
				end
				if esc() or main.f_input(main.t_players, {'m'}) then
					if not bool_main then
						sndPlay(motif.files.snd_data, motif[main.group].cancel_snd[1], motif[main.group].cancel_snd[2])
					elseif not esc() and t[item].itemname ~= 'exit' then
						--menu key moves cursor to exit without exiting the game
						for i = 1, #t do
							if t[i].itemname == 'exit' then
								sndPlay(motif.files.snd_data, motif[main.group].cancel_snd[1], motif[main.group].cancel_snd[2])
								item = i
								cursorPosY = math.min(item, motif[main.group].menu_window_visibleitems)
								if cursorPosY >= motif[main.group].menu_window_visibleitems then
									moveTxt = (item - motif[main.group].menu_window_visibleitems) * motif[main.group].menu_item_spacing[2]
								end
								break
							end
						end
					end
					if not bool_main or esc() then
						break
					end
				elseif bool_f1 and (getKey('F1') or config.FirstRun) then
					if config.FirstRun then
						config.FirstRun = false
						options.f_saveCfg(false)
					end
					main.f_warning(
						main.f_extractText(motif.infobox_text),
						motif[main.background],
						motif.infobox,
						txt_infoboxTitle,
						txt_infobox,
						overlay_infobox
					)
				elseif main.credits ~= -1 and getKey(motif.attract_mode.credits_key) then
					sndPlay(motif.files.snd_data, motif.attract_mode.credits_snd[1], motif.attract_mode.credits_snd[2])
					main.credits = main.credits + 1
					resetKey()
				elseif motif.attract_mode.enabled == 1 and getKey(motif.attract_mode.options_key) then
					main.f_default()
					main.menu.f = main.t_itemname.options()
					sndPlay(motif.files.snd_data, motif[main.group].cursor_done_snd[1], motif[main.group].cursor_done_snd[2])
					main.f_fadeReset('fadeout', motif[main.group])
					resetKey()
				elseif bool_del and getKey('DELETE') then
					tbl.items = main.f_deleteIP(item, t)
				elseif main.f_input(main.t_players, main.f_extractKeys(motif[main.group].menu_hiscore_key)) and main.f_hiscoreDisplay(t[item].itemname) then
					demoFrameCounter = 0
				elseif main.f_input(main.t_players, main.f_extractKeys(motif[main.group].menu_accept_key)) then
					demoFrameCounter = 0
					local f = t[item].itemname
					if f == 'back' then
						sndPlay(motif.files.snd_data, motif[main.group].cancel_snd[1], motif[main.group].cancel_snd[2])
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
							if motif.title_info['cursor_' .. f .. '_snd'] ~= nil then
								sndPlay(motif.files.snd_data, motif.title_info['cursor_' .. f .. '_snd'][1], motif.title_info['cursor_' .. f .. '_snd'][2])
							else
								sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
							end
							tbl.submenu[f].loop()
							f = ''
						else
							break
						end
					end
					if f ~= '' then
						main.f_default()
						if f == 'joinadd' then
							tbl.items = main.t_itemname[f](t, item)
						elseif main.t_itemname[f] ~= nil then
							main.menu.f = main.t_itemname[f](t, item)
						end
						if main.menu.f ~= nil then
							if motif.title_info['cursor_' .. f .. '_snd'] ~= nil then
								sndPlay(motif.files.snd_data, motif.title_info['cursor_' .. f .. '_snd'][1], motif.title_info['cursor_' .. f .. '_snd'][2])
							else
								sndPlay(motif.files.snd_data, motif.title_info.cursor_done_snd[1], motif.title_info.cursor_done_snd[2])
							end
							main.f_fadeReset('fadeout', motif[main.group])
						end
					end
				end
			end
		end
	end
end

-- Dynamically generates all menus and submenus, iterating over values stored in
-- main.t_sort table (in order that they're present in system.def).
function main.f_start()
	if main.t_sort.title_info == nil or main.t_sort.title_info.menu == nil or #main.t_sort.title_info.menu == 0 then
		motif.setBaseTitleInfo()
	end
	main.menu = {title = main.f_itemnameUpper(motif[main.group].title_text, motif[main.group].menu_title_uppercase == 1), submenu = {}, items = {}}
	main.menu.loop = main.f_createMenu(main.menu, true, main.group == 'title_info', main.group == 'title_info', false)
	local t_menuWindow = main.f_menuWindow(motif[main.group])
	local t_pos = {} --for storing current main.menu table position
	local t_skipGroup = {}
	local lastNum = 0
	local bonusUpper = true
	for i, suffix in ipairs(main.f_tableExists(main.t_sort[main.group]).menu) do
		for j, c in ipairs(main.f_strsplit('_', suffix)) do --split using "_" delimiter
			--exceptions for expanding the menu table
			if motif[main.group]['menu_itemname_' .. suffix] == '' and c ~= 'server' then --items and groups without displayname are skipped
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
			end
			--appending the menu table
			if j == 1 then --first string after menu.itemname (either reserved one or custom submenu assignment)
				if main.menu.submenu[c] == nil then
					main.menu.submenu[c] = {title = main.f_itemnameUpper(motif[main.group]['menu_itemname_' .. suffix], motif[main.group].menu_title_uppercase == 1), submenu = {}, items = {}}
					main.menu.submenu[c].loop = main.f_createMenu(main.menu.submenu[c], false, false, true, c == 'serverjoin')
					if not suffix:match(c .. '_') then
						table.insert(main.menu.items, {
							data = text:create({window = t_menuWindow}),
							itemname = c,
							displayname = motif[main.group]['menu_itemname_' .. suffix],
							paramname = 'menu_itemname_' .. suffix,
						})
						if c == 'bonusgames' then bonusUpper = main.menu.items[#main.menu.items].displayname == main.menu.items[#main.menu.items].displayname:upper() end
					end
				end
				t_pos = main.menu.submenu[c]
				t_pos.name = c
			else --following strings
				if t_pos.submenu[c] == nil then
					t_pos.submenu[c] = {title = main.f_itemnameUpper(motif[main.group]['menu_itemname_' .. suffix], motif[main.group].menu_title_uppercase == 1), submenu = {}, items = {}}
					t_pos.submenu[c].loop = main.f_createMenu(t_pos.submenu[c], false, false, true, c == 'serverjoin')
					table.insert(t_pos.items, {
						data = text:create({window = t_menuWindow}),
						itemname = c,
						displayname = motif[main.group]['menu_itemname_' .. suffix],
						paramname = 'menu_itemname_' .. suffix,
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
						data = text:create({window = t_menuWindow}),
						itemname = itemname,
						displayname = main.f_itemnameUpper(name, bonusUpper),
						paramname = 'menu_itemname_' .. suffix:gsub('back$', itemname),
					})
					--creating anim data out of appended menu items
					motif.f_loadSprData(motif[main.group], {s = 'menu_bg_' .. suffix:gsub('back$', itemname) .. '_', x = motif[main.group].menu_pos[1], y = motif[main.group].menu_pos[2]})
					motif.f_loadSprData(motif[main.group], {s = 'menu_bg_active_' .. suffix:gsub('back$', itemname) .. '_', x = motif[main.group].menu_pos[1], y = motif[main.group].menu_pos[2]})
				end
			end
			--add story arcs to storymode submenu
			if suffix:match('storymode_back$') and c == 'storymode' then --j == main.f_countSubstring(suffix, '_') then
				for k, v in ipairs(main.t_selStoryMode) do
					local itemname = v.name:gsub('%s+', '_')
					table.insert(t_pos.items, {
						data = text:create({window = t_menuWindow}),
						itemname = itemname,
						displayname = v.displayname,
						paramname = 'menu_itemname_' .. suffix:gsub('back$', itemname),
					})
					--creating anim data out of appended menu items
					motif.f_loadSprData(motif[main.group], {s = 'menu_bg_' .. suffix:gsub('back$', itemname) .. '_', x = motif[main.group].menu_pos[1], y = motif[main.group].menu_pos[2]})
					motif.f_loadSprData(motif[main.group], {s = 'menu_bg_active_' .. suffix:gsub('back$', itemname) .. '_', x = motif[main.group].menu_pos[1], y = motif[main.group].menu_pos[2]})
				end
			end
			--add IP addresses for serverjoin submenu
			if suffix:match('_serverjoin_back$') and c == 'serverjoin' then --j == main.f_countSubstring(suffix, '_') then
				for k, v in pairs(config.IP) do
					local itemname = 'ip_' .. k
					table.insert(t_pos.items, {
						data = text:create({window = t_menuWindow}),
						itemname = itemname,
						displayname = k,
						--paramname = 'menu_itemname_' .. suffix:gsub('back$', itemname),
					})
					--motif.f_loadSprData(motif[main.group], {s = 'menu_bg_' .. suffix:gsub('back$', itemname) .. '_', x = motif[main.group].menu_pos[1], y = motif[main.group].menu_pos[2]})
					--motif.f_loadSprData(motif[main.group], {s = 'menu_bg_active_' .. suffix:gsub('back$', itemname) .. '_', x = motif[main.group].menu_pos[1], y = motif[main.group].menu_pos[2]})
				end
			end
		end
	end
	if main.debugLog then main.f_printTable(main.menu, 'debug/t_mainMenu.txt') end
end

--replay menu
local txt_titleReplay = main.f_createTextImg(motif.replay_info, 'title', {defsc = motif.defaultReplay})
local t_menuWindowReplay = main.f_menuWindow(motif.replay_info)
function main.f_replay()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = {}
	for k, v in ipairs(getDirectoryFiles('save/replays')) do
		v:gsub('^(.-)([^\\/]+)%.([^%.\\/]-)$', function(path, filename, ext)
			path = path:gsub('\\', '/')
			ext = ext:lower()
			if ext == 'replay' then
				table.insert(t, {data = text:create({window = t_menuWindowReplay}), itemname = path .. filename .. '.' .. ext, displayname = filename})
			end
		end)
	end
	table.insert(t, {data = text:create({window = t_menuWindowReplay}), itemname = 'back', displayname = motif.replay_info.menu_itemname_back})
	main.f_bgReset(motif.replaybgdef.bg)
	main.f_fadeReset('fadein', motif.replay_info)
	if motif.music.replay_bgm ~= '' then
		main.f_playBGM(false, motif.music.replay_bgm, motif.music.replay_bgm_loop, motif.music.replay_bgm_volume, motif.music.replay_bgm_loopstart, motif.music.replay_bgm_loopend)
	end
	main.close = false
	while true do
		main.f_menuCommonDraw(t, item, cursorPosY, moveTxt, 'replay_info', 'replaybgdef', txt_titleReplay, motif.defaultReplay, {})
		cursorPosY, moveTxt, item = main.f_menuCommonCalc(t, item, cursorPosY, moveTxt, 'replay_info', {'$U'}, {'$D'})
		if main.close and not main.fadeActive then
			main.f_bgReset(motif[main.background].bg)
			main.f_fadeReset('fadein', motif[main.group])
			main.f_playBGM(false, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
			main.close = false
			break
		elseif esc() or main.f_input(main.t_players, {'m'}) or (t[item].itemname == 'back' and main.f_input(main.t_players, {'pal', 's'})) then
			sndPlay(motif.files.snd_data, motif.replay_info.cancel_snd[1], motif.replay_info.cancel_snd[2])
			main.f_fadeReset('fadeout', motif.replay_info)
			main.close = true
		elseif main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif[main.group].cursor_done_snd[1], motif[main.group].cursor_done_snd[2])
			enterReplay(t[item].itemname)
			synchronize()
			math.randomseed(sszRandom())
			main.f_cmdBufReset()
			main.menu.submenu.server.loop()
			replayStop()
			exitNetPlay()
			exitReplay()
		end
	end
end

local txt_connecting = main.f_createTextImg(motif.title_info, 'connecting')
local overlay_connecting = main.f_createOverlay(motif.title_info, 'connecting_overlay')
function main.f_connect(server, t)
	enterNetPlay(server)
	while not connected() do
		if esc() or main.f_input(main.t_players, {'m'}) then
			sndPlay(motif.files.snd_data, motif.title_info.cancel_snd[1], motif.title_info.cancel_snd[2])
			exitNetPlay()
			return false
		end
		--draw clearcolor
		clearColor(motif[main.background].bgclearcolor[1], motif[main.background].bgclearcolor[2], motif[main.background].bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(motif[main.background].bg, false)
		--draw overlay
		overlay_connecting:draw()
		--draw text
		for i = 1, #t do
			txt_connecting:update({
				text = t[i],
				y = motif[main.group].connecting_offset[2] + main.f_ySpacing(motif.title_info, 'connecting') * (i - 1),
			})
			txt_connecting:draw()
		end
		--draw layerno = 1 backgrounds
		bgDraw(motif[main.background].bg, true)
		main.f_cmdInput()
		refresh()
	end
	replayRecord('save/replays/' .. os.date("%Y-%m-%d %I-%M%p-%Ss") .. '.replay')
	return true
end

--asserts content unlock conditions
function main.f_unlock(permanent)
	for group, t in pairs(main.t_unlockLua) do
		local t_del = {}
		for k, v in pairs(t) do
			local bool = assert(loadstring('return ' .. v))()
			if type(bool) == 'boolean' then
				if group == 'chars' then
					main.f_unlockChar(k, bool, false)
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
end

--unlock characters (select screen grid only)
function main.f_unlockChar(num, bool, reset)
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
		end
	elseif main.t_selChars[num].hidden_default == nil then
		return
	elseif main.t_selChars[num].hidden ~= main.t_selChars[num].hidden_default then
		main.t_selChars[num].hidden = main.t_selChars[num].hidden_default
		start.t_grid[main.t_selChars[num].row][main.t_selChars[num].col].hidden = main.t_selChars[num].hidden
		if reset then start.f_resetGrid() end
	end
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

--hiscore rendering
main.t_hiscoreData = {
	arcade = {mode = 'arcade', data = 'score', title = motif.select_info.title_arcade_text},
	survival = {mode = 'survival', data = 'win', title = motif.select_info.title_survival_text},
	survivalcoop = {mode = 'survivalcoop', data = 'win', title = motif.select_info.title_survivalcoop_text},
	teamcoop = {mode = 'teamcoop', data = 'score', title = motif.select_info.title_teamcoop_text},
	timeattack = {mode = 'timeattack', data = 'time', title = motif.select_info.title_timeattack_text},
}
main.t_hiscoreData.teamarcade = main.t_hiscoreData.arcade

function main.f_hiscoreDisplay(itemname)
	if main.t_hiscoreData[itemname] == nil or motif.hiscore_info.enabled == 0 or stats.modes == nil or stats.modes[main.t_hiscoreData[itemname].mode] == nil or stats.modes[main.t_hiscoreData[itemname].mode].ranking == nil then
		return false
	end
	main.f_cmdBufReset()
	sndPlay(motif.files.snd_data, motif[main.group].cursor_done_snd[1], motif[main.group].cursor_done_snd[2])
	start.hiscoreInit = false
	while start.f_hiscore(main.t_hiscoreData[itemname], true, -1, true) do
		main.f_refresh()
	end
	main.f_fadeReset('fadein', motif[main.group])
	main.f_playBGM(false, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
	return true
end

--attract mode start screen
local txt_attract_credits = main.f_createTextImg(motif.attract_mode, 'credits')
local txt_attract_timer = main.f_createTextImg(motif.attract_mode, 'start_timer')
local txt_attract_insert = main.f_createTextImg(motif.attract_mode, 'start_insert')
local txt_attract_press = main.f_createTextImg(motif.attract_mode, 'start_press')
function main.f_attractStart()
	local timerActive = main.credits ~= 0
	local timer = 0
	local counter = 0 - motif.attract_mode.fadein_time
	local press_blinktime, insert_blinktime = 0, 0
	local press_switched, insert_switched = false, false
	txt_attract_insert:update({text = motif.attract_mode.start_insert_text})
	txt_attract_press:update({text = motif.attract_mode.start_press_text})
	main.f_cmdBufReset()
	clearColor(motif.attractbgdef.bgclearcolor[1], motif.attractbgdef.bgclearcolor[2], motif.attractbgdef.bgclearcolor[3])
	main.f_bgReset(motif.attractbgdef.bg)
	main.f_fadeReset('fadein', motif.attract_mode)
	main.f_playBGM(false, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
	while true do
		counter = counter + 1
		--draw layerno = 0 backgrounds
		bgDraw(motif.attractbgdef.bg, false)
		--draw text
		if main.credits ~= 0 then
			if motif.attract_mode.start_press_blinktime > 0 and main.fadeType == 'fadein' then
				if press_blinktime < motif.attract_mode.start_press_blinktime then
					press_blinktime = press_blinktime + 1
				elseif press_switched then
					txt_attract_press:update({text = motif.attract_mode.start_press_text})
					press_switched = false
					press_blinktime = 0
				else
					txt_attract_press:update({text = ''})
					press_switched = true
					press_blinktime = 0
				end
			end
			txt_attract_press:draw()
		else
			if motif.attract_mode.start_insert_blinktime > 0 and main.fadeType == 'fadein' then
				if insert_blinktime < motif.attract_mode.start_insert_blinktime then
					insert_blinktime = insert_blinktime + 1
				elseif insert_switched then
					txt_attract_insert:update({text = motif.attract_mode.start_insert_text})
					insert_switched = false
					insert_blinktime = 0
				else
					txt_attract_insert:update({text = ''})
					insert_switched = true
					insert_blinktime = 0
				end
			end
			txt_attract_insert:draw()
		end
		--draw timer
		if motif.attract_mode.start_timer_count ~= -1 and timerActive then
			timer, timerActive = main.f_drawTimer(timer, motif.attract_mode, 'start_timer_', txt_attract_timer)
		end
		--draw credits text
		if main.credits ~= -1 then
			txt_attract_credits:update({text = main.f_extractText(motif.attract_mode.credits_text, main.credits)[1]})
			txt_attract_credits:draw()
		end
		--credits
		if main.credits ~= -1 and getKey(motif.attract_mode.credits_key) then
			sndPlay(motif.files.snd_data, motif.attract_mode.credits_snd[1], motif.attract_mode.credits_snd[2])
			main.credits = main.credits + 1
			resetKey()
			timerActive = true
			timer = motif.attract_mode.start_timer_displaytime
		end
		--options
		if motif.attract_mode.enabled == 1 and getKey(motif.attract_mode.options_key) then
			main.f_default()
			main.menu.f = main.t_itemname.options()
			sndPlay(motif.files.snd_data, motif[main.group].cursor_done_snd[1], motif[main.group].cursor_done_snd[2])
			main.f_fadeReset('fadeout', motif[main.group])
			resetKey()
			main.menu.f()
			return false
		end
		--draw layerno = 1 backgrounds
		bgDraw(motif.attractbgdef.bg, true)
		--draw fadein / fadeout
		if main.fadeType == 'fadein' and not main.fadeActive and ((main.credits ~= 0 and main.f_input(main.t_players, {'s'})) or (not timerActive and counter >= motif.attract_mode.start_time)) then
			if main.credits ~= 0 then
				sndPlay(motif.files.snd_data, motif.attract_mode.start_done_snd[1], motif.attract_mode.start_done_snd[2])
			end
			main.f_fadeReset('fadeout', motif.attract_mode)
		end
		main.f_fadeAnim(motif.attract_mode)
		--frame transition
		main.f_cmdInput()
		if esc() --[[or main.f_input(main.t_players, {'m'})]] then
			esc(false)
			return false
		end
		if not main.fadeActive and main.fadeType == 'fadeout' then
			return main.credits ~= 0
		end
		main.f_refresh()
	end
end

--attract mode loop
function main.f_attractMode()
	main.credits = 0
	while true do --outer loop
		local startScreen = false
		while true do --inner loop (attract mode)
			--logo storyboard
			if motif.attract_mode.logo_storyboard ~= '' and storyboard.f_storyboard(motif.attract_mode.logo_storyboard, true) then
				break
			end
			--intro storyboard
			if motif.attract_mode.intro_storyboard ~= '' and storyboard.f_storyboard(motif.attract_mode.intro_storyboard, true) then
				break
			end
			--demo
			main.f_demoStart()
			if main.credits > 0 then break end
			--hiscores
			start.hiscoreInit = false
			while start.f_hiscore(main.t_hiscoreData.arcade, true, -1, false) do
				main.f_refresh()
			end
			if main.credits > 0 then break end
			--start
			if main.f_attractStart() then
				startScreen = true
				break
			end
			--demo
			main.f_demoStart()
			if main.credits > 0 then break end
			--hiscores
			start.hiscoreInit = false
			while start.f_hiscore(main.t_hiscoreData.arcade, true, -1, false) do
				main.f_refresh()
			end
			if main.credits > 0 then break end
		end
		if startScreen or main.f_attractStart() then
			--attract storyboard
			if motif.attract_mode.start_storyboard ~= '' then
				storyboard.f_storyboard(motif.attract_mode.start_storyboard, false)
			end
			--eat credit
			if main.credits > 0 then
				main.credits = main.credits - 1
			end
			--enter menu
			main.menu.loop()
		elseif main.credits > 0 then
			main.credits = main.credits - 1
		end
	end
end

main.credits = -1
function main.f_setCredits()
	if motif.attract_mode.enabled == 1 or start.challenger ~= 0 then
		return
	end
	main.credits = config.Credits - 1
end

--demo mode
function main.f_demo()
	if #main.t_randomChars == 0 then
		return
	end
	if main.fadeActive or motif.demo_mode.enabled == 0 then
		demoFrameCounter = 0
		return
	end
	demoFrameCounter = demoFrameCounter + 1
	if demoFrameCounter < motif.demo_mode.title_waittime then
		return
	end
	main.f_fadeReset('fadeout', motif.demo_mode)
	main.menu.f = main.t_itemname.demo()
end

function main.f_demoStart()
	main.f_default()
	if motif.demo_mode.debuginfo == 0 and config.DebugKeys then
		setAllowDebugKeys(false)
		setAllowDebugMode(false)
	end
	main.lifebar.bars = motif.demo_mode.fight_bars_display == 1
	setGameMode('demo')
	for i = 1, 2 do
		setCom(i, 8)
		setTeamMode(i, 0, 1)
		local ch = main.t_randomChars[math.random(1, #main.t_randomChars)]
		selectChar(i, ch, getCharRandomPalette(ch))
	end
	local stage = start.f_setStage()
	start.f_setMusic(stage)
	if motif.demo_mode.fight_stopbgm == 1 then
		main.f_playBGM(true) --stop music
	end
	hook.run("main.t_itemname")
	clearColor(motif[main.background].bgclearcolor[1], motif[main.background].bgclearcolor[2], motif[main.background].bgclearcolor[3])
	loadStart()
	game()
	setAllowDebugKeys(config.DebugKeys)
	setAllowDebugMode(config.DebugMode)
	if motif.attract_mode.enabled == 0 then
		if introWaitCycles >= motif.demo_mode.intro_waitcycles then
			start.hiscoreInit = false
			while start.f_hiscore(main.t_hiscoreData.arcade, true, -1, false) do
				main.f_refresh()
			end
			if motif.files.intro_storyboard ~= '' then
				storyboard.f_storyboard(motif.files.intro_storyboard)
			end
			introWaitCycles = 0
		else
			introWaitCycles = introWaitCycles + 1
		end
		main.f_bgReset(motif[main.background].bg)
		--start title BGM only if it has been interrupted
		if motif.demo_mode.fight_stopbgm == 1 or motif.demo_mode.fight_playbgm == 1 or (introWaitCycles == 0 and motif.files.intro_storyboard ~= '') then
			main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
		end
	end
	main.f_fadeReset('fadein', motif.demo_mode)
end

--common menu calculations
function main.f_menuCommonCalc(t, item, cursorPosY, moveTxt, section, keyPrev, keyNext)
	local startItem = 1
	for _, v in ipairs(t) do
		if v.itemname ~= 'empty' then
			break
		end
		startItem = startItem + 1
	end
	if main.f_input(main.t_players, keyNext) then
		sndPlay(motif.files.snd_data, motif[section].cursor_move_snd[1], motif[section].cursor_move_snd[2])
		while true do
			item = item + 1
			if cursorPosY < motif[section].menu_window_visibleitems then
				cursorPosY = cursorPosY + 1
			end
			if t[item] == nil or t[item].itemname ~= 'empty' then
				break
			end
		end
	elseif main.f_input(main.t_players, keyPrev) then
		sndPlay(motif.files.snd_data, motif[section].cursor_move_snd[1], motif[section].cursor_move_snd[2])
		while true do
			item = item - 1
			if cursorPosY > startItem then
				cursorPosY = cursorPosY - 1
			end
			if t[item] == nil or t[item].itemname ~= 'empty' then
				break
			end
		end
	end
	if item > #t or (item == 1 and t[item].itemname == 'empty') then
		item = 1
		while true do
			if t[item].itemname ~= 'empty' or item >= #t then
				break
			else
				item = item + 1
			end
		end
		cursorPosY = item
	elseif item < 1 then
		item = #t
		while true do
			if t[item].itemname ~= 'empty' or item <= 1 then
				break
			else
				item = item - 1
			end
		end
		if item > motif[section].menu_window_visibleitems then
			cursorPosY = motif[section].menu_window_visibleitems
		else
			cursorPosY = item
		end
	end
	if cursorPosY >= motif[section].menu_window_visibleitems then
		moveTxt = (item - motif[section].menu_window_visibleitems) * motif[section].menu_item_spacing[2]
	elseif cursorPosY <= startItem then
		moveTxt = (item - startItem) * motif[section].menu_item_spacing[2]
	end
	return cursorPosY, moveTxt, item
end

--frame change command buffer and fadeout signal
function main.f_frameChange()
	if main.fadeActive or main.fadeCnt > 0 then
		main.f_cmdBufReset()
	elseif main.fadeType == 'fadeout' then
		main.f_cmdBufReset()
		return false --fadeout ended
	else
		main.f_cmdInput()
	end
	return true
end

--common menu draw
local rect_boxcursor = rect:create({})
local rect_boxbg = rect:create({})
function main.f_menuCommonDraw(t, item, cursorPosY, moveTxt, section, bgdef, title, defsc, footer_txt, skipClear)
	--draw clearcolor
	if not skipClear then
		clearColor(motif[bgdef].bgclearcolor[1], motif[bgdef].bgclearcolor[2], motif[bgdef].bgclearcolor[3])
	end
	--draw layerno = 0 backgrounds
	bgDraw(motif[bgdef].bg, false)
	--draw menu box
	if motif[section].menu_boxbg_visible == 1 then
		rect_boxbg:update({
			x1 =    motif[section].menu_pos[1] + motif[section].menu_boxcursor_coords[1],
			y1 =    motif[section].menu_pos[2] + motif[section].menu_boxcursor_coords[2],
			x2 =    motif[section].menu_boxcursor_coords[3] - motif[section].menu_boxcursor_coords[1] + 1,
			y2 =    motif[section].menu_boxcursor_coords[4] - motif[section].menu_boxcursor_coords[2] + 1 + (math.min(#t, motif[section].menu_window_visibleitems) - 1) * motif[section].menu_item_spacing[2],
			r =     motif[section].menu_boxbg_col[1],
			g =     motif[section].menu_boxbg_col[2],
			b =     motif[section].menu_boxbg_col[3],
			src =   motif[section].menu_boxbg_alpha[1],
			dst =   motif[section].menu_boxbg_alpha[2],
			defsc = defsc,
		})
		rect_boxbg:draw()
	end
	--draw title
	title:draw()
	--draw menu items
	local items_shown = item + motif[section].menu_window_visibleitems - cursorPosY
	if items_shown > #t or (motif[section].menu_window_visibleitems > 0 and items_shown < #t and (motif[section].menu_window_margins_y[1] ~= 0 or motif[section].menu_window_margins_y[2] ~= 0)) then
		items_shown = #t
	end
	for i = 1, items_shown do
		if i > item - cursorPosY then
			if i == item then
				--Draw active item background
				if t[i].paramname ~= nil then
					animDraw(motif[section][t[i].paramname:gsub('menu_itemname_', 'menu_bg_active_') .. '_data'])
					animUpdate(motif[section][t[i].paramname:gsub('menu_itemname_', 'menu_bg_active_') .. '_data'])
				end
				--Draw active item font
				if t[i].selected then
					t[i].data:update({
						font =   motif[section].menu_item_selected_active_font[1],
						bank =   motif[section].menu_item_selected_active_font[2],
						align =  motif[section].menu_item_selected_active_font[3],
						text =   t[i].displayname,
						x =      motif[section].menu_pos[1] + motif[section].menu_item_offset[1] + (i - 1) * motif[section].menu_item_spacing[1],
						y =      motif[section].menu_pos[2] + motif[section].menu_item_offset[2] + (i - 1) * motif[section].menu_item_spacing[2] - moveTxt,
						scaleX = motif[section].menu_item_selected_active_scale[1],
						scaleY = motif[section].menu_item_selected_active_scale[2],
						r =      motif[section].menu_item_selected_active_font[4],
						g =      motif[section].menu_item_selected_active_font[5],
						b =      motif[section].menu_item_selected_active_font[6],
						height = motif[section].menu_item_selected_active_font[7],
						defsc =  defsc,
					})
					t[i].data:draw()
				else
					t[i].data:update({
						font =   motif[section].menu_item_active_font[1],
						bank =   motif[section].menu_item_active_font[2],
						align =  motif[section].menu_item_active_font[3],
						text =   t[i].displayname,
						x =      motif[section].menu_pos[1] + motif[section].menu_item_active_offset[1] + (i - 1) * motif[section].menu_item_spacing[1],
						y =      motif[section].menu_pos[2] + motif[section].menu_item_active_offset[2] + (i - 1) * motif[section].menu_item_spacing[2] - moveTxt,
						scaleX = motif[section].menu_item_active_scale[1],
						scaleY = motif[section].menu_item_active_scale[2],
						r =      motif[section].menu_item_active_font[4],
						g =      motif[section].menu_item_active_font[5],
						b =      motif[section].menu_item_active_font[6],
						height = motif[section].menu_item_active_font[7],
						defsc =  defsc,
					})
					t[i].data:draw()
				end
				if t[i].vardata ~= nil then
					t[i].vardata:update({
						font =   motif[section].menu_item_value_active_font[1],
						bank =   motif[section].menu_item_value_active_font[2],
						align =  motif[section].menu_item_value_active_font[3],
						text =   t[i].vardisplay,
						x =      motif[section].menu_pos[1] + motif[section].menu_item_value_active_offset[1] + (i - 1) * motif[section].menu_item_spacing[1],
						y =      motif[section].menu_pos[2] + motif[section].menu_item_value_active_offset[2] + (i - 1) * motif[section].menu_item_spacing[2] - moveTxt,
						scaleX = motif[section].menu_item_value_active_scale[1],
						scaleY = motif[section].menu_item_value_active_scale[2],
						r =      motif[section].menu_item_value_active_font[4],
						g =      motif[section].menu_item_value_active_font[5],
						b =      motif[section].menu_item_value_active_font[6],
						height = motif[section].menu_item_value_active_font[7],
						defsc =  defsc,
					})
					t[i].vardata:draw()
				end
			else
				--Draw not active item background
				if t[i].paramname ~= nil then
					animDraw(motif[section][t[i].paramname:gsub('menu_itemname_', 'menu_bg_') .. '_data'])
					animUpdate(motif[section][t[i].paramname:gsub('menu_itemname_', 'menu_bg_') .. '_data'])
				end
				--Draw not active item font
				if t[i].selected then
					t[i].data:update({
						font =   motif[section].menu_item_selected_font[1],
						bank =   motif[section].menu_item_selected_font[2],
						align =  motif[section].menu_item_selected_font[3],
						text =   t[i].displayname,
						x =      motif[section].menu_pos[1] + motif[section].menu_item_selected_offset[1] + (i - 1) * motif[section].menu_item_spacing[1],
						y =      motif[section].menu_pos[2] + motif[section].menu_item_selected_offset[2] + (i - 1) * motif[section].menu_item_spacing[2] - moveTxt,
						scaleX = motif[section].menu_item_selected_scale[1],
						scaleY = motif[section].menu_item_selected_scale[2],
						r =      motif[section].menu_item_selected_font[4],
						g =      motif[section].menu_item_selected_font[5],
						b =      motif[section].menu_item_selected_font[6],
						height = motif[section].menu_item_selected_font[7],
						defsc =  defsc,
					})
					t[i].data:draw()
				else
					t[i].data:update({
						font =   motif[section].menu_item_font[1],
						bank =   motif[section].menu_item_font[2],
						align =  motif[section].menu_item_font[3],
						text =   t[i].displayname,
						x =      motif[section].menu_pos[1] + motif[section].menu_item_offset[1] + (i - 1) * motif[section].menu_item_spacing[1],
						y =      motif[section].menu_pos[2] + motif[section].menu_item_offset[2] + (i - 1) * motif[section].menu_item_spacing[2] - moveTxt,
						scaleX = motif[section].menu_item_scale[1],
						scaleY = motif[section].menu_item_scale[2],
						r =      motif[section].menu_item_font[4],
						g =      motif[section].menu_item_font[5],
						b =      motif[section].menu_item_font[6],
						height = motif[section].menu_item_font[7],
						defsc =  defsc,
					})
					t[i].data:draw()
				end
				if t[i].vardata ~= nil then
					t[i].vardata:update({
						font =   motif[section].menu_item_value_font[1],
						bank =   motif[section].menu_item_value_font[2],
						align =  motif[section].menu_item_value_font[3],
						text =   t[i].vardisplay,
						x =      motif[section].menu_pos[1] + motif[section].menu_item_value_offset[1] + (i - 1) * motif[section].menu_item_spacing[1],
						y =      motif[section].menu_pos[2] + motif[section].menu_item_value_offset[2] + (i - 1) * motif[section].menu_item_spacing[2] - moveTxt,
						scaleX = motif[section].menu_item_value_scale[1],
						scaleY = motif[section].menu_item_value_scale[2],
						r =      motif[section].menu_item_value_font[4],
						g =      motif[section].menu_item_value_font[5],
						b =      motif[section].menu_item_value_font[6],
						height = motif[section].menu_item_value_font[7],
						defsc =  defsc,
					})
					t[i].vardata:draw()
				end
			end
		end
	end
	--draw menu cursor
	if motif[section].menu_boxcursor_visible == 1 and not main.fadeActive then
		local src, dst = main.f_boxcursorAlpha(
			motif[section].menu_boxcursor_alpharange[1],
			motif[section].menu_boxcursor_alpharange[2],
			motif[section].menu_boxcursor_alpharange[3],
			motif[section].menu_boxcursor_alpharange[4],
			motif[section].menu_boxcursor_alpharange[5],
			motif[section].menu_boxcursor_alpharange[6]
		)
		rect_boxcursor:update({
			x1 =    motif[section].menu_pos[1] + motif[section].menu_boxcursor_coords[1] + (cursorPosY - 1) * motif[section].menu_item_spacing[1],
			y1 =    motif[section].menu_pos[2] + motif[section].menu_boxcursor_coords[2] + (cursorPosY - 1) * motif[section].menu_item_spacing[2],
			x2 =    motif[section].menu_boxcursor_coords[3] - motif[section].menu_boxcursor_coords[1] + 1,
			y2 =    motif[section].menu_boxcursor_coords[4] - motif[section].menu_boxcursor_coords[2] + 1,
			r =     motif[section].menu_boxcursor_col[1],
			g =     motif[section].menu_boxcursor_col[2],
			b =     motif[section].menu_boxcursor_col[3],
			src =   src,
			dst =   dst,
			defsc = defsc,
		})
		rect_boxcursor:draw()
	end
	--draw scroll arrows
	if #t > motif[section].menu_window_visibleitems then
		if item > cursorPosY then
			animUpdate(motif[section].menu_arrow_up_data)
			animDraw(motif[section].menu_arrow_up_data)
		end
		if item >= cursorPosY and item + motif[section].menu_window_visibleitems - cursorPosY < #t then
			animUpdate(motif[section].menu_arrow_down_data)
			animDraw(motif[section].menu_arrow_down_data)
		end
	end
	--draw credits text
	if motif.attract_mode.enabled == 1 and main.credits ~= -1 then
		txt_attract_credits:update({text = main.f_extractText(motif.attract_mode.credits_text, main.credits)[1]})
		txt_attract_credits:draw()
	end
	--draw layerno = 1 backgrounds
	bgDraw(motif[bgdef].bg, true)
	--draw footer overlay
	if motif[section].footer_overlay_window ~= nil then
		overlay_footer:draw()
	end
	--draw footer text
	for i = 1, #footer_txt do
		footer_txt[i]:draw()
	end
	--draw fadein / fadeout
	main.f_fadeAnim(main.fadeGroup)
	--frame transition
	if not main.f_frameChange() then
		return --skip last frame rendering
	end
	if not skipClear then
		refresh()
	end
end

--common timer draw code
function main.f_drawTimer(timer, t, prefix, txt)
	local num = main.f_round((t[prefix .. 'count'] * t[prefix .. 'framespercount'] - timer + t[prefix .. 'displaytime']) / t[prefix .. 'framespercount'])
	local active = true
	if num <= -1 then
		active = false
		timer = -1
		txt:update({text = t[prefix .. 'text']:gsub('%%i', tostring(0))})
	elseif timer ~= -1 then
		timer = timer + 1
		txt:update({text = t[prefix .. 'text']:gsub('%%i', tostring(math.max(0, num)))})
	end
	if timer == -1 or timer >= t[prefix .. 'displaytime'] then
		txt:draw()
	end
	return timer, active
end

--reset background
function main.f_bgReset(data)
	main.t_animUpdate = {}
	alpha1cur = 0
	alpha2cur = 0
	alpha1add = true
	alpha2add = true
	bgReset(data)
end

--reset fade
function main.f_fadeReset(fadeType, fadeGroup)
	main.fadeType = fadeType
	main.fadeGroup = fadeGroup
	main.fadeStart = getFrameCount()
	main.fadeCnt = 0
	if fadeGroup[fadeType .. '_data'] ~= nil then
		animReset(fadeGroup[fadeType .. '_data'])
		animUpdate(fadeGroup[fadeType .. '_data'])
		main.fadeCnt = animGetLength(fadeGroup[fadeType .. '_data'])
		if fadeType == 'fadeout' and main.fadeCnt > fadeGroup[fadeType .. '_time'] then
			main.fadeStart = main.fadeStart + main.fadeCnt - fadeGroup[fadeType .. '_time']
		end
	end
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
for _, v in ipairs(config.Modules) do
	table.insert(t_modules, v)
end
if motif.files.module ~= '' then table.insert(t_modules, motif.files.module) end
for _, v in ipairs(t_modules) do
	print('Loading module: ' .. v)
	v = v:gsub('^%s*[%./\\]*', '')
	v = v:gsub('%.[^%.]+$', '')
	require(v:gsub('[/\\]+', '.'))
	--assert(loadfile(v))()
end

--assert(loadstring(main.lua))()
main.f_unlock(false)

--;===========================================================
--; INITIALIZE LOOPS
--;===========================================================
if main.debugLog then
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
	main.f_printTable(config, "debug/config.txt")
end

main.f_start()
menu.f_start()
options.f_start()
motif.f_start()

if main.flags['-p1'] ~= nil and main.flags['-p2'] ~= nil then
	main.f_default()
	main.f_commandLine()
end

if main.flags['-stresstest'] ~= nil then
	main.f_default()
	local frameskip = tonumber(main.flags['-stresstest'])
	if frameskip >= 1 then
		setGameSpeed((frameskip + 1) * config.Framerate)
	end
	setGameMode('randomtest')
	randomtest.run()
	os.exit()
end

main.f_loadingRefresh(main.txt_loading)
main.txt_loading = nil
--sleep(1)

if motif.attract_mode.enabled == 1 then
	main.f_attractMode()
else
	main.menu.loop()
end

-- Debug Info
--main.motifData = nil
--if main.debugLog then main.f_printTable(main, "debug/t_main.txt") end

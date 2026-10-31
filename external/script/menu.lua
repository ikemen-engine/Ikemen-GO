local menu = {}

--;===========================================================
--; PAUSE MENU
--;===========================================================

-- Associative elements table storing arrays with training menu option names.
-- Can be appended via external module.
menu.t_valuename = {
	dummycontrol = {
		{itemname = 'cooperative', displayname = motif.pause_menu.training_pause_menu.menu.valuename.dummycontrol_cooperative},
		{itemname = 'ai', displayname = motif.pause_menu.training_pause_menu.menu.valuename.dummycontrol_ai},
		{itemname = 'manual', displayname = motif.pause_menu.training_pause_menu.menu.valuename.dummycontrol_manual},
	},
	ailevel = {
		{itemname = '1', displayname = motif.pause_menu.training_pause_menu.menu.valuename.ailevel_1},
		{itemname = '2', displayname = motif.pause_menu.training_pause_menu.menu.valuename.ailevel_2},
		{itemname = '3', displayname = motif.pause_menu.training_pause_menu.menu.valuename.ailevel_3},
		{itemname = '4', displayname = motif.pause_menu.training_pause_menu.menu.valuename.ailevel_4},
		{itemname = '5', displayname = motif.pause_menu.training_pause_menu.menu.valuename.ailevel_5},
		{itemname = '6', displayname = motif.pause_menu.training_pause_menu.menu.valuename.ailevel_6},
		{itemname = '7', displayname = motif.pause_menu.training_pause_menu.menu.valuename.ailevel_7},
		{itemname = '8', displayname = motif.pause_menu.training_pause_menu.menu.valuename.ailevel_8},
	},
	dummymode = {
		{itemname = 'stand', displayname = motif.pause_menu.training_pause_menu.menu.valuename.dummymode_stand},
		{itemname = 'crouch', displayname = motif.pause_menu.training_pause_menu.menu.valuename.dummymode_crouch},
		{itemname = 'jump', displayname = motif.pause_menu.training_pause_menu.menu.valuename.dummymode_jump},
		{itemname = 'wjump', displayname = motif.pause_menu.training_pause_menu.menu.valuename.dummymode_wjump},
	},
	guardmode = {
		{itemname = 'none', displayname = motif.pause_menu.training_pause_menu.menu.valuename.guardmode_none},
		{itemname = 'auto', displayname = motif.pause_menu.training_pause_menu.menu.valuename.guardmode_auto},
		{itemname = 'all', displayname = motif.pause_menu.training_pause_menu.menu.valuename.guardmode_all},
		{itemname = 'random', displayname = motif.pause_menu.training_pause_menu.menu.valuename.guardmode_random},
	},
	fallrecovery = {
		{itemname = 'none', displayname = motif.pause_menu.training_pause_menu.menu.valuename.fallrecovery_none},
		{itemname = 'ground', displayname = motif.pause_menu.training_pause_menu.menu.valuename.fallrecovery_ground},
		{itemname = 'air', displayname = motif.pause_menu.training_pause_menu.menu.valuename.fallrecovery_air},
		{itemname = 'random', displayname = motif.pause_menu.training_pause_menu.menu.valuename.fallrecovery_random},
	},
	distance = {
		{itemname = 'any', displayname = motif.pause_menu.training_pause_menu.menu.valuename.distance_any},
		{itemname = 'close', displayname = motif.pause_menu.training_pause_menu.menu.valuename.distance_close},
		{itemname = 'medium', displayname = motif.pause_menu.training_pause_menu.menu.valuename.distance_medium},
		{itemname = 'far', displayname = motif.pause_menu.training_pause_menu.menu.valuename.distance_far},
	},
	buttonjam = {
		{itemname = 'none', displayname = motif.pause_menu.training_pause_menu.menu.valuename.buttonjam_none},
		{itemname = 'a', displayname = motif.pause_menu.training_pause_menu.menu.valuename.buttonjam_a},
		{itemname = 'b', displayname = motif.pause_menu.training_pause_menu.menu.valuename.buttonjam_b},
		{itemname = 'c', displayname = motif.pause_menu.training_pause_menu.menu.valuename.buttonjam_c},
		{itemname = 'x', displayname = motif.pause_menu.training_pause_menu.menu.valuename.buttonjam_x},
		{itemname = 'y', displayname = motif.pause_menu.training_pause_menu.menu.valuename.buttonjam_y},
		{itemname = 'z', displayname = motif.pause_menu.training_pause_menu.menu.valuename.buttonjam_z},
		{itemname = 's', displayname = motif.pause_menu.training_pause_menu.menu.valuename.buttonjam_s},
		{itemname = 'd', displayname = motif.pause_menu.training_pause_menu.menu.valuename.buttonjam_d},
		{itemname = 'w', displayname = motif.pause_menu.training_pause_menu.menu.valuename.buttonjam_w},
	},
}

-- Shared logic for pause menu option change, returns 2 values:
-- * boolean depending if option has changed (via right/left button press)
-- * itemname read from t_valuename table based on currently active option
--   (or nil, if there was no option change in this frame)
function menu.f_valueChanged(t, sec)
	local valueitem = menu[t.itemname] or 1
	local chk = valueitem
	if getAction(-1, sec.menu.add.action) then
		valueitem = valueitem + 1
	elseif getAction(-1, sec.menu.subtract.action) then
		valueitem = valueitem - 1
	end
	if valueitem > #menu.t_valuename[t.itemname] then
		valueitem = 1
	elseif valueitem < 1 then
		valueitem = #menu.t_valuename[t.itemname]
	end
	-- true upon option change
	if chk ~= valueitem then
		sndPlay(motif.Snd, sec.cursor.move.snd[1], sec.cursor.move.snd[2])
		t.vardisplay = menu.t_valuename[t.itemname][valueitem].displayname
		menu[t.itemname] = valueitem
		menu.itemname = t.itemname
		return true, menu.t_valuename[t.itemname][valueitem].itemname
	end
	return false, nil
end

-- Current pause menu itemname for internal use (key from menu.t_itemname table)
menu.itemname = ''

-- Associative elements table storing functions controlling behaviour of each
-- pause menu item. Can be appended via external module.
menu.t_itemname = {
	--Back
	['back'] = function(t, item, cursorPosY, moveTxt, sec)
		if getAction(-1, sec.menu.done.action) then
			if menu.currentMenu[1] == menu.currentMenu[2] then
				sndPlay(motif.Snd, sec.exit.snd[1], sec.exit.snd[2])
				menu.pauseExitDelay = gameOption('Input.PauseExitDelay')
			else
				sndPlay(motif.Snd, sec.cancel.snd[1], sec.cancel.snd[2])
			end
			if menu.currentMenu[1] ~= menu.currentMenu[2] then
				main.f_menuSnap(sec)
			end
			menu.currentMenu[1] = menu.currentMenu[2]
			return false
		end
		return true
	end,
	--Dummy Control
	['dummycontrol'] = function(t, item, cursorPosY, moveTxt, sec)
		local ok, name = menu.f_valueChanged(t.items[item], sec)
		if ok then
			if name == 'cooperative' or name == 'manual' then
				player(main.opponentSide)
				setAILevel(0)
			elseif name == 'ai' then
				player(main.opponentSide)
				setAILevel(menu.ailevel)
			end
			player(main.opponentSide)
			mapSet('_iksys_trainingDummyControl', menu.dummycontrol - 1)
		end
		return true
	end,
	--AI Level
	['ailevel'] = function(t, item, cursorPosY, moveTxt, sec)
		if menu.f_valueChanged(t.items[item], sec) then
			if menu.t_valuename.dummycontrol[menu.dummycontrol or 1].itemname == 'ai' then
				player(main.opponentSide)
				setAILevel(menu.ailevel)
			end
		end
		return true
	end,
	--Dummy Mode
	['dummymode'] = function(t, item, cursorPosY, moveTxt, sec)
		if menu.f_valueChanged(t.items[item], sec) then
			player(main.opponentSide)
			mapSet('_iksys_trainingDummyMode', menu.dummymode - 1)
		end
		return true
	end,
	--Guard Mode
	['guardmode'] = function(t, item, cursorPosY, moveTxt, sec)
		if menu.f_valueChanged(t.items[item], sec) then
			player(main.opponentSide)
			mapSet('_iksys_trainingGuardMode', menu.guardmode - 1)
		end
		return true
	end,
	--Fall Recovery
	['fallrecovery'] = function(t, item, cursorPosY, moveTxt, sec)
		if menu.f_valueChanged(t.items[item], sec) then
			player(main.opponentSide)
			mapSet('_iksys_trainingFallRecovery', menu.fallrecovery - 1)
		end
		return true
	end,
	--Distance
	['distance'] = function(t, item, cursorPosY, moveTxt, sec)
		if menu.f_valueChanged(t.items[item], sec) then
			player(main.opponentSide)
			mapSet('_iksys_trainingDistance', menu.distance - 1)
		end
		return true
	end,
	--Button Jam
	['buttonjam'] = function(t, item, cursorPosY, moveTxt, sec)
		if menu.f_valueChanged(t.items[item], sec) then
			player(main.opponentSide)
			mapSet('_iksys_trainingButtonJam', menu.buttonjam - 1)
		end
		return true
	end,
	-- Controller Setup of the player who paused, the same editor as in Options
	['menuinput'] = function(t, item, cursorPosY, moveTxt, sec, bg)
		if getAction(-1, sec.menu.done.action) then
			sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
			options.f_pauseSetup(bg)
			menu.itemname = 'inputconfig'
		end
		return true
	end,
	--Round Reset
	['reset'] = function(t, item, cursorPosY, moveTxt, sec)
		if getAction(-1, sec.menu.done.action) then
			sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
			--togglePause(false)
			resetRound()
			main.pauseMenuActive = false
			return false
		end
		return true
	end,
	--Reload (Rematch)
	['reload'] = function(t, item, cursorPosY, moveTxt, sec)
		if getAction(-1, sec.menu.done.action) then
			sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
			--togglePause(false)
			reload()
			main.pauseMenuActive = false
			return false
		end
		return true
	end,
	--Command List
	['commandlist'] = function(t, item, cursorPosY, moveTxt, sec)
		if getAction(-1, sec.menu.done.action) then
			sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
			menu.f_commandlistParse(sec)
			menu.itemname = t.items[item].itemname
		end
		return true
	end,
	--Character Change
	['characterchange'] = function(t, item, cursorPosY, moveTxt, sec)
		if getAction(-1, sec.menu.done.action) then
			sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2], 100, 0, 0, 0, 0, true)
			--togglePause(false)
			endMatch()
			start.characterchange = true
			main.pauseMenuActive = false
			return false
		end
		return true
	end,
	--Exit
	['exit'] = function(t, item, cursorPosY, moveTxt, sec)
		if getAction(-1, sec.menu.done.action) then
			sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2], 100, 0, 0, 0, 0, true)
			--togglePause(false)
			endMatch()
			start.exit = true
			main.pauseMenuActive = false
			return false
		end
		return true
	end,
}
-- options.t_itemname table functions are also appended to this table, to make
-- option screen settings logic accessible from within pause menu.
for k, v in pairs(options.t_itemname) do
	if menu.t_itemname[k] == nil then
		menu.t_itemname[k] = v
	end
end

-- Shared menu loop logic
function menu.f_createMenu(tbl, sec, bg, bool_main)
	return function()
		hook.run("menu.menu.loop")
		if menu._lastMenuTbl ~= tbl then
			menu._lastMenuTbl = tbl
			main.f_menuItemBgAnimReset(sec)
		end
		local t = tbl.items
		if tbl.reset then
			tbl.reset = false
		else
			local name = t[tbl.item].itemname
			local hintContext = options.f_hintContext(t[tbl.item])
			if menu.t_valuename[name] ~= nil then
				hintContext = 'adjust'
			end
			local actions = nil
			if options.t_itemname[name] ~= nil and menu.t_itemname[name] == options.t_itemname[name] then
				-- Imported Options handlers still read their own action keys.
				actions = {
					['menu.done'] = motif.option_info.menu.done.action,
					['menu.add'] = motif.option_info.menu.add.action,
					['menu.subtract'] = motif.option_info.menu.subtract.action,
				}
			end
			main.f_menuCommonDraw(t, tbl.item, tbl.cursorPosY, tbl.moveTxt, sec, bg, true,
				{hintContext = hintContext, hintActions = actions})
		end
		-- Draw during fades, but don't accept menu input until they finish.
		if not main.pauseMenuActive or fadeActive() or menu.pauseExitDelay >= 0 then
			return
		end
		tbl.cursorPosY, tbl.moveTxt, tbl.item = main.f_menuCommonCalc(t, tbl.item, tbl.cursorPosY, tbl.moveTxt, sec, sec.cursor)
		textImgReset(sec.title.TextSpriteData)
		textImgSetText(sec.title.TextSpriteData, tbl.title)
		if esc() or getAction(-1, sec.menu.cancel.action) then
			if bool_main then
				sndPlay(motif.Snd, sec.exit.snd[1], sec.exit.snd[2])
				menu.pauseExitDelay = gameOption('Input.PauseExitDelay')
			else
				sndPlay(motif.Snd, sec.cancel.snd[1], sec.cancel.snd[2])
			end
			if menu.currentMenu[1] ~= menu.currentMenu[2] then
				main.f_menuSnap(sec)
			end
			menu.currentMenu[1] = menu.currentMenu[2]
			return
		elseif menu.t_itemname[t[tbl.item].itemname] ~= nil then
			if not menu.t_itemname[t[tbl.item].itemname](tbl, tbl.item, tbl.cursorPosY, tbl.moveTxt, sec, bg) then
				return
			end
		elseif getAction(-1, sec.menu.done.action) then
			local f = t[tbl.item].itemname
			if tbl.submenu[f].loop ~= nil then
				sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
				menu.currentMenu[1] = tbl.submenu[f].loop
				main.f_menuSnap(sec)
			elseif not menu.t_itemname[f](tbl, tbl.item, tbl.cursorPosY, tbl.moveTxt, sec, bg) then
				return
			end
		end
	end
end

menu.t_vardisplayPointers = {}

-- Associative elements table storing functions returning current setting values
-- rendered alongside menu item name. Can be appended via external module.
menu.t_vardisplay = {
	['dummycontrol'] = function()
		return menu.t_valuename.dummycontrol[menu.dummycontrol or 1].displayname
	end,
	['ailevel'] = function()
		return menu.t_valuename.ailevel[menu.ailevel or gameOption('Options.Difficulty')].displayname
	end,
	['dummymode'] = function()
		return menu.t_valuename.dummymode[menu.dummymode or 1].displayname
	end,
	['guardmode'] = function()
		return menu.t_valuename.guardmode[menu.guardmode or 1].displayname
	end,
	['fallrecovery'] = function()
		return menu.t_valuename.fallrecovery[menu.fallrecovery or 1].displayname
	end,
	['distance'] = function()
		return menu.t_valuename.distance[menu.distance or 1].displayname
	end,
	['buttonjam'] = function()
		return menu.t_valuename.buttonjam[menu.buttonjam or 1].displayname
	end,
}

-- Returns setting value rendered alongside menu item name (calls appropriate
-- function from menu or options t_vardisplay table)
function menu.f_vardisplay(itemname)
	if menu.t_vardisplay[itemname] ~= nil then
		return menu.t_vardisplay[itemname]()
	end
	if options.t_vardisplay[itemname] ~= nil then
		return options.t_vardisplay[itemname]()
	end
	return ''
end

-- Table storing arrays with data used for different pause menu types generation.
-- Can be appended via external module but not necessary - can specify gamemode-specific [<gamemode> Pause Menu] sections in `system.def` instead.
menu.t_menus = {
	{id = 'menu', sec = motif.pause_menu.pause_menu, bg = motif.pausebgdef.pausebgdef, movelist = true},
	{id = 'training', sec = motif.pause_menu.training_pause_menu, bg = motif.pausebgdef.trainingpausebgdef, movelist = true},
}
menu.t_menuIndex = {}

local function f_pauseMenuKey(gamemodeName)
	gamemodeName = tostring(gamemodeName or ''):lower()
	gamemodeName = gamemodeName:gsub('%s+', '_')
	if gamemodeName == '' then
		return ''
	end
	return gamemodeName .. '_pause_menu'
end

local function f_pauseMenuIdFromKey(key)
	key = tostring(key or ''):lower()
	if key:sub(-11) == '_pause_menu' then
		key = key:sub(1, -12)
	end
	return key
end

local function f_pauseMenuBgKey(gamemodeName)
	gamemodeName = tostring(gamemodeName or ''):lower()
	gamemodeName = gamemodeName:gsub('%s+', '_')
	if gamemodeName == '' then
		return ''
	end
	return gamemodeName .. 'pausebgdef'
end

local function f_registerPauseMenu(id, sec, bg, movelist)
	if id == '' or sec == nil then
		return
	end
	for i = 1, #menu.t_menus do
		if menu.t_menus[i].id == id then
			return
		end
	end
	table.insert(menu.t_menus, {id = id, sec = sec, bg = bg, movelist = movelist})
end

local function f_indexPauseMenus()
	menu.t_menuIndex = {}
	for i = 1, #menu.t_menus do
		menu.t_menuIndex[menu.t_menus[i].id] = menu.t_menus[i]
	end
end

-- Dynamically generates all menus and submenus
function menu.f_start()
	if motif.pause_menu ~= nil then
		for k, sec in pairs(motif.pause_menu) do
			local id = f_pauseMenuIdFromKey(k)
			if id ~= '' and id ~= 'pause_menu' then
				local bg = motif.pausebgdef.pausebgdef
				local bgKey = f_pauseMenuBgKey(id)
				if bgKey ~= '' and motif.pausebgdef ~= nil and motif.pausebgdef[bgKey] ~= nil then
					bg = motif.pausebgdef[bgKey]
				end
				f_registerPauseMenu(id, sec, bg, true)
			end
		end
	end
	f_indexPauseMenus()
	for k, v in ipairs(menu.t_menus) do
		menu[v.id] = {
			title = main.f_itemnameUpper(v.sec.title.text, v.sec.menu.title.uppercase),
			cursorPosY = 1,
			moveTxt = 0,
			item = 1,
			submenu = {},
			items = {}
		}
		menu[v.id].loop = menu.f_createMenu(menu[v.id], v.sec, v.bg, true)
		local t_pos = {} --for storing current table position
		local lastNum = 0
		for i, suffix in ipairs(v.sec.menu.itemname_order) do
			for j, c in ipairs(main.f_strsplit('_', suffix)) do --split using "_" delimiter
				--appending the menu table
				if j == 1 then --first string after menu.itemname (either reserved one or custom submenu assignment)
					if menu[v.id].submenu[c] == nil or c:match("^spacer%d*$") then
						menu[v.id].submenu[c] = {}
						menu[v.id].submenu[c].title = main.f_itemnameUpper(v.sec.menu.itemname[suffix], v.sec.menu.title.uppercase)
						if menu.t_itemname[c] == nil and not c:match("^spacer%d*$") then
							menu[v.id].submenu[c].cursorPosY = 1
							menu[v.id].submenu[c].moveTxt = 0
							menu[v.id].submenu[c].item = 1
							menu[v.id].submenu[c].submenu = {}
							menu[v.id].submenu[c].items = {}
							menu[v.id].submenu[c].loop = menu.f_createMenu(menu[v.id].submenu[c], v.sec, v.bg, false)
						end
						if not suffix:match(c .. '_') then
							table.insert(menu[v.id].items, {
								itemname = c,
								displayname = v.sec.menu.itemname[suffix],
								paramname = suffix,
								vardisplay = menu.f_vardisplay(c),
								selected = false,
							})
							table.insert(menu.t_vardisplayPointers, menu[v.id].items[#menu[v.id].items])
						end
					end
					t_pos = menu[v.id].submenu[c]
					t_pos.name = c
				else --following strings
					if t_pos.submenu[c] == nil or c:match("^spacer%d*$") then
						t_pos.submenu[c] = {}
						t_pos.submenu[c].title = main.f_itemnameUpper(v.sec.menu.itemname[suffix], v.sec.menu.title.uppercase)
						if menu.t_itemname[c] == nil and not c:match("^spacer%d*$") then
							t_pos.submenu[c].cursorPosY = 1
							t_pos.submenu[c].moveTxt = 0
							t_pos.submenu[c].item = 1
							t_pos.submenu[c].submenu = {}
							t_pos.submenu[c].items = {}
							t_pos.submenu[c].loop = menu.f_createMenu(t_pos.submenu[c], v.sec, v.bg, false)
						end
						table.insert(t_pos.items, {
							itemname = c,
							displayname = v.sec.menu.itemname[suffix],
							paramname = suffix,
							vardisplay = menu.f_vardisplay(c),
							selected = false,
						})
						table.insert(menu.t_vardisplayPointers, t_pos.items[#t_pos.items])
					end
					if j > lastNum then
						t_pos = t_pos.submenu[c]
						t_pos.name = c
					end
				end
				lastNum = j
			end
		end
		-- Runtime platform filtering
		if getRuntimeOS() == 'android' then
			local excluded = {keyboard = true}
			main.f_pruneMenu(menu[v.id], excluded)
			main.f_prunePointers(menu.t_vardisplayPointers, excluded)
		end
		-- Menu windows
		main.f_menuWindowSet(v.sec.menu)
		if gameOption('Debug.DumpLuaTables') then main.f_printTable(menu[v.id], 'debug/t_' .. v.id .. 'Menu.txt') end
		if v.movelist then
			main.f_menuWindowSet(v.sec.movelist.menu, true)
		end
	end
end

-- Called from start.lua right before game() function to reset training menu values and p2 settings for a new match
function menu.f_trainingReset()
	for k, _ in pairs(menu.t_valuename) do
		menu[k] = 1
	end
	menu.ailevel = gameOption('Options.Difficulty')
	for _, v in ipairs(menu.t_vardisplayPointers) do
		v.vardisplay = menu.f_vardisplay(v.itemname)
	end
	player(main.opponentSide)
	setAILevel(0)
end

hook.add('loop#training', 'trainingSettings', function()
	if not roundStart() then return end
	player(main.opponentSide)
	mapSet('_iksys_trainingDummy', 1)
	setAILevel(menu.dummycontrol == 2 and menu.ailevel or 0)
	mapSet('_iksys_trainingDummyControl', menu.dummycontrol - 1)
	mapSet('_iksys_trainingDummyMode', menu.dummymode - 1)
	mapSet('_iksys_trainingGuardMode', menu.guardmode - 1)
	mapSet('_iksys_trainingFallRecovery', menu.fallrecovery - 1)
	mapSet('_iksys_trainingDistance', menu.distance - 1)
	mapSet('_iksys_trainingButtonJam', menu.buttonjam - 1)
end)

menu.movelistChar = 1
-- player: 1-based player who opened the menu. The title can show it with %s.
function menu.f_init(player)
	esc(false)
	main.pauseMenuActive = true
	local id = f_pauseMenuIdFromKey(f_pauseMenuKey(gameMode()))
	if id == '' or menu.t_menuIndex == nil or menu.t_menuIndex[id] == nil then
		id = 'menu'
	end
	if menu.t_menuIndex ~= nil and menu.t_menuIndex[id] ~= nil then
		local entry = menu.t_menuIndex[id]
		if menu[id] ~= nil then
			menu[id].title = main.f_itemnameUpper(entry.sec.title.text:gsub('%%s', tostring(player or 1)), entry.sec.menu.title.uppercase)
		end
		sndPlay(motif.Snd, entry.sec.enter.snd[1], entry.sec.enter.snd[2])
		bgReset(entry.bg.BGDef)
		if menu[id] ~= nil and menu[id].loop ~= nil then
			menu.currentMenu = {menu[id].loop, menu[id].loop}
			menu.currentMenuId = id
		else
			menu.currentMenu = {menu.menu.loop, menu.menu.loop}
			menu.currentMenuId = 'menu'
		end
	else
		sndPlay(motif.Snd, motif.pause_menu.pause_menu.enter.snd[1], motif.pause_menu.pause_menu.enter.snd[2])
		bgReset(motif.pausebgdef.pausebgdef.BGDef)
		menu.currentMenu = {menu.menu.loop, menu.menu.loop}
		menu.currentMenuId = 'menu'
	end
end

menu.pauseExitDelay = -1

function menu.f_run()
	local entry = nil
	if menu.t_menuIndex ~= nil then
		if menu.currentMenuId ~= nil and menu.t_menuIndex[menu.currentMenuId] ~= nil then
			entry = menu.t_menuIndex[menu.currentMenuId]
		else
			local id = f_pauseMenuIdFromKey(f_pauseMenuKey(gameMode()))
			if id ~= '' and menu.t_menuIndex[id] ~= nil then
				entry = menu.t_menuIndex[id]
			elseif menu.t_menuIndex['menu'] ~= nil then
				entry = menu.t_menuIndex['menu']
			end
		end
	end
	local sec = motif.pause_menu.pause_menu
	local bg = motif.pausebgdef.pausebgdef
	if entry ~= nil and entry.sec ~= nil and entry.bg ~= nil then
		sec = entry.sec
		bg = entry.bg
	end
	--draw overlay
	rectDraw(sec.overlay.RectData)
	--pause exit delay
	if menu.pauseExitDelay >= 0 then
		if menu.pauseExitDelay > 0 then
			menu.pauseExitDelay = menu.pauseExitDelay - 1
		else
			-- Keep the game paused. Go owns the fade-out/fade-in handoff and unpauses afterwards.
			main.pauseMenuActive = false
		end
	end
	-- Input configuration uses the active pause background and the overlay above.
	if menu.itemname == 'inputconfig' then
		if options.inputContext == nil or not options.f_inputConfigFrame(options.inputContext) then
			options.inputContext = nil
			options.f_saveCfg(false)
			menu.itemname = ''
		end
	--Command List
	elseif menu.itemname == 'commandlist' then
		menu.f_commandlistRender(sec, bg)
	--Menu
	else
		menu.currentMenu[1]()
	end
	local active = main.pauseMenuActive
	if not active then
		menu.pauseExitDelay = -1
	end
	return active
end

-- Reset selection/scroll state recursively for a menu table (root + submenus)
local function f_resetMenuState(tbl)
	if type(tbl) ~= "table" then return end
	-- Reset cursor/scroll
	tbl.cursorPosY = 1
	tbl.moveTxt = 0
	tbl.item = 1
	-- Clear any "selected" flags on items (if used by custom menus)
	if type(tbl.items) == "table" then
		for i = 1, #tbl.items do
			if type(tbl.items[i]) == "table" then
				tbl.items[i].selected = false
			end
		end
	end
	-- Recurse into submenus
	if type(tbl.submenu) == "table" then
		for _, sub in pairs(tbl.submenu) do
			f_resetMenuState(sub)
		end
	end
end

function menu.f_reset()
	-- exit any special screens rendered by menu.f_run()
	if options.inputContext ~= nil then
		options.inputContext.closed = true
		options.f_inputConfigFrame(options.inputContext)
		options.inputContext = nil
	end
	menu.itemname = ''
	-- Reset movelist selector (command list screen)
	menu.movelistChar = 1
	menu._lastMovelist = nil
	-- Reset tween caches
	if menu.t_menus ~= nil then
		for i = 1, #menu.t_menus do
			local sec = menu.t_menus[i].sec
			if sec ~= nil then
				sec.menuTweenData = nil
				sec.boxCursorData = nil
				if sec.movelist ~= nil then
					sec.movelist.menuTweenData = nil
					sec.movelist.boxCursorData = nil
				end
			end
		end
	end
	-- Determine which root menu is active and reset its cursor/item/scroll (and all submenus)
	if menu.currentMenu ~= nil and menu.currentMenu[2] ~= nil then
		if menu.currentMenuId ~= nil and menu[menu.currentMenuId] ~= nil then
			f_resetMenuState(menu[menu.currentMenuId])
		elseif menu.training ~= nil and menu.training.loop ~= nil and menu.currentMenu[2] == menu.training.loop then
			f_resetMenuState(menu.training)
		else
			f_resetMenuState(menu.menu)
		end
	end
	-- Force snap on next calc so we don't keep old scroll offsets.
	main.menuSnap = true
	-- restore root menu loop
	if menu.currentMenu ~= nil and menu.currentMenu[2] ~= nil then
		menu.currentMenu[1] = menu.currentMenu[2]
	end
end

function menuInit(player)
	return menu.f_init(player)
end

function menuRun()
	return menu.f_run()
end

function menuReset()
	return menu.f_reset()
end

--;===========================================================
--; COMMAND LIST
--;===========================================================
function menu.f_commandlistParse(sec)
	menu.t_movelists = {}
	for side, tbl in ipairs({start.p[1].t_selected, start.p[2].t_selected}) do
		for member, sel in ipairs(tbl) do
			local pn = side
			if member > 1 then
				pn = pn + (member - 1) * 2
			end
			if player(pn) and aiLevel() == 0 then
				local movelist = getMovelist()
				if sel.movelistText ~= movelist then
					sel.movelistText = movelist
					sel.commandlist = nil
					sel.movelistState = {item = 1, cursorPosY = 1, moveTxt = 0}
					if movelist:find('%S') then
						-- wrap the plain glyph names in <>, but leave existing tags alone
						for k, v in main.f_sortKeys(motif.glyphs, function(t, a, b) return string.len(a) > string.len(b) end) do
							if k:sub(1, 1) ~= '<' then
								local escaped = k:gsub('&', '&amp;'):gsub('<', '&lt;'):gsub('>', '&gt;')
								movelist = main.f_textTransform(movelist, function(text)
									return (text:gsub(main.f_escapePattern(k), function() return '<' .. escaped .. '>' end))
								end)
							end
						end
						local t = {}
						movelist = movelist:gsub('\r\n', '\n'):gsub('\n$', '')
						for line in (movelist .. '\n'):gmatch('(.-)\n') do
							line = line:gsub('%s+$', '')
							local spacer
							line, spacer = line:gsub('^<spacer>', '')
							local subt = {}
							for m in line:gmatch('(	*[^	]+)') do
								local tabs = 0
								m = m:gsub('^(	*)', function(m1)
									tabs = string.len(m1)
									return ''
								end)
								local align = 1 --left align
								if tabs == 1 then
									align = 0 --center align
								elseif tabs > 1 then
									align = -1 --right align
								end
								table.insert(subt, {text = m, align = align})
							end
							table.insert(t, {
								itemname = ((line == '' or spacer > 0) and 'spacer' or '') .. tostring(#t + 1),
								paramname = tostring(#t + 1),
								displayname = line,
								columns = subt,
							})
						end
						sel.commandlist = t
					end
				end
				table.insert(menu.t_movelists, {
					pn = pn,
					name = displayName(),
					state = sel.movelistState,
					items = sel.commandlist,
				})
			end
		end
	end
	if #menu.t_movelists == 0 then
		table.insert(menu.t_movelists, {
			pn = 1,
			name = "",
			state = {item = 1, cursorPosY = 1, moveTxt = 0},
		})
	end
	for _, t in ipairs(menu.t_movelists) do
		if t.items == nil or #t.items == 0 then
			t.items = {{itemname = 'empty', paramname = 'empty', displayname = sec.movelist.menu.item.text}}
		end
	end
	menu._lastMovelist = nil
	if menu.movelistChar > #menu.t_movelists then
		menu.movelistChar = 1
	end
	if gameOption('Debug.DumpLuaTables') then main.f_printTable(menu.t_movelists, "debug/t_movelists.txt") end
end

function menu.f_commandlistRender(sec, bg)
	local mv = sec.movelist
	local m = mv.menu
	local acceptInput = main.pauseMenuActive and not fadeActive() and menu.pauseExitDelay < 0
	local switched = false
	if acceptInput then
		if esc() or getAction(-1, m.cancel.action, m.done.action) then
			sndPlay(motif.Snd, sec.cancel.snd[1], sec.cancel.snd[2])
			menu.itemname = ''
			main.f_menuSnap(sec)
			return
		elseif #menu.t_movelists > 1 then
			local direction = 0
			if getAction(-1, m.subtract.action) then
				direction = -1
			elseif getAction(-1, m.add.action) then
				direction = 1
			end
			if direction ~= 0 then
				sndPlay(motif.Snd, sec.cursor.move.snd[1], sec.cursor.move.snd[2])
				menu.movelistChar = (menu.movelistChar - 1 + direction) % #menu.t_movelists + 1
				switched = true
			end
		end
	end
	local t = menu.t_movelists[menu.movelistChar]
	local state = t.state
	if menu._lastMovelist ~= t then
		menu._lastMovelist = t
		mv.boxCursorData = nil
		main.f_menuSnap(mv)
		main.f_menuItemBgAnimReset(mv)
	end
	local forcedDir = nil
	if not acceptInput or switched then
		forcedDir = 0
	end
	state.cursorPosY, state.moveTxt, state.item = main.f_menuCommonCalc(
		t.items, state.item, state.cursorPosY, state.moveTxt, mv, sec.cursor, forcedDir)
	textImgReset(mv.title.TextSpriteData)
	textImgSetText(mv.title.TextSpriteData, main.f_itemnameUpper(string.format(mv.title.text, t.name), m.title.uppercase))
	main.f_menuCommonDraw(t.items, state.item, state.cursorPosY, state.moveTxt, mv, bg, true)
end

return menu

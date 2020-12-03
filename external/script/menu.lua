local menu = {}

--;===========================================================
--; PAUSE MENU
--;===========================================================
menu.itemname = ''
menu.t_itemname = {
	--Back
	['back'] = function(t, item, cursorPosY, moveTxt, section)
		if main.f_input(main.t_players, {'pal', 's'}) then
			if menu.currentMenu[1] == menu.currentMenu[2] then
				sndPlay(motif.files.snd_data, motif[section].cursor_done_snd[1], motif[section].cursor_done_snd[2])
				togglePause(false)
				main.pauseMenu = false
			else
				sndPlay(motif.files.snd_data, motif[section].cancel_snd[1], motif[section].cancel_snd[2])
			end
			menu.currentMenu[1] = menu.currentMenu[2]
			return false
		end
		return true
	end,
	--Key Config
	['keyboard'] = function(t, item, cursorPosY, moveTxt, section)
		if main.f_input(main.t_players, {'pal', 's'}) --[[or getKey('F1')]] then
			sndPlay(motif.files.snd_data, motif[section].cursor_done_snd[1], motif[section].cursor_done_snd[2])
			options.f_keyCfgInit('KeyConfig', t.submenu[t.items[item].itemname].title)
			menu.itemname = t.items[item].itemname
		end
		return true
	end,
	--Joystick Config
	['gamepad'] = function(t, item, cursorPosY, moveTxt, section)
		if main.f_input(main.t_players, {'pal', 's'}) --[[or getKey('F2')]] then
			sndPlay(motif.files.snd_data, motif[section].cursor_done_snd[1], motif[section].cursor_done_snd[2])
			options.f_keyCfgInit('JoystickConfig', t.submenu[t.items[item].itemname].title)
			menu.itemname = t.items[item].itemname
		end
		return true
	end,
	--Default
	['inputdefault'] = function(t, item, cursorPosY, moveTxt, section)
		if main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif[section].cursor_done_snd[1], motif[section].cursor_done_snd[2])
			options.f_keyDefault()
			for pn = 1, #config.KeyConfig do
				setKeyConfig(pn, config.KeyConfig[pn].Joystick, config.KeyConfig[pn].Buttons)
			end
			if main.flags['-nojoy'] == nil then
				for pn = 1, #config.JoystickConfig do
					setKeyConfig(pn, config.JoystickConfig[pn].Joystick, config.JoystickConfig[pn].Buttons)
				end
			end
			options.f_saveCfg(false)
		end
		return true
	end,
	--Round Reset
	['reset'] = function(t, item, cursorPosY, moveTxt, section)
		if main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif[section].cursor_done_snd[1], motif[section].cursor_done_snd[2])
			togglePause(false)
			roundReset()
			main.pauseMenu = false
			return false
		end
		return true
	end,
	--Reload (Rematch)
	['reload'] = function(t, item, cursorPosY, moveTxt, section)
		if main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif[section].cursor_done_snd[1], motif[section].cursor_done_snd[2])
			togglePause(false)
			reload()
			main.pauseMenu = false
			return false
		end
		return true
	end,
	--Command List
	['commandlist'] = function(t, item, cursorPosY, moveTxt, section)
		if main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif[section].cursor_done_snd[1], motif[section].cursor_done_snd[2])
			menu.f_commandlistParse()
			menu.itemname = t.items[item].itemname
		end
		return true
	end,
	--Character Change
	['characterchange'] = function(t, item, cursorPosY, moveTxt, section)
		if main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif[section].cursor_done_snd[1], motif[section].cursor_done_snd[2])
			togglePause(false)
			endMatch()
			main.pauseMenu = false
			return false
		end
		return true
	end,
	--Exit
	['exit'] = function(t, item, cursorPosY, moveTxt, section)
		if main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif[section].cursor_done_snd[1], motif[section].cursor_done_snd[2])
			togglePause(false)
			endMatch()
			start.exit = true
			main.pauseMenu = false
			return false
		end
		return true
	end,
}

function menu.f_createMenu(tbl, section, bgdef, txt_title, bool_main)
	return function()
		local t = tbl.items
		if tbl.reset then
			tbl.reset = false
			main.f_cmdInput()
		else
			main.f_menuCommonDraw(t, tbl.item, tbl.cursorPosY, tbl.moveTxt, section, bgdef, txt_title, motif.defaultMenu, {}, true)
		end
		tbl.cursorPosY, tbl.moveTxt, tbl.item = main.f_menuCommonCalc(t, tbl.item, tbl.cursorPosY, tbl.moveTxt, section, {'$U'}, {'$D'})
		txt_title:update({text = tbl.title})
		if esc() or main.f_input(main.t_players, {'m'}) then
			if bool_main then
				togglePause(false)
				main.pauseMenu = false
			else
				sndPlay(motif.files.snd_data, motif[section].cancel_snd[1], motif[section].cancel_snd[2])
			end
			menu.currentMenu[1] = menu.currentMenu[2]
			return
		elseif menu.t_itemname[t[tbl.item].itemname] ~= nil then
			if not menu.t_itemname[t[tbl.item].itemname](tbl, tbl.item, tbl.cursorPosY, tbl.moveTxt, section) then
				return
			end
		elseif main.f_input(main.t_players, {'pal', 's'}) then
			local f = t[tbl.item].itemname
			if tbl.submenu[f].loop ~= nil then
				sndPlay(motif.files.snd_data, motif[section].cursor_done_snd[1], motif[section].cursor_done_snd[2])
				menu.currentMenu[1] = tbl.submenu[f].loop
			elseif not menu.t_itemname[f](tbl, tbl.item, tbl.cursorPosY, tbl.moveTxt, section) then
				return
			end
		end
	end
end

function menu.f_vardisplay(itemname)
	return ''
end

--dynamically generates all menus and submenus using itemname data stored in main.t_sort table
for k, v in pairs(
	{
		{id = 'menu', section = 'menu_info', bgdef = 'menubgdef', txt_title = 'txt_title_menu'},
		{id = 'training', section = 'training_info', bgdef = 'trainingbgdef', txt_title = 'txt_title_training'},
	}
) do
	menu[v.txt_title] = main.f_createTextImg(motif[v.section], 'title', {defsc = motif.defaultMenu})
	menu[v.id] = {
		title = main.f_itemnameUpper(motif[v.section].title_text, motif[v.section].menu_title_uppercase == 1),
		cursorPosY = 1,
		moveTxt = 0,
		item = 1,
		submenu = {},
		items = {}
	}
	menu[v.id].loop = menu.f_createMenu(menu[v.id], v.section, v.bgdef, menu[v.txt_title], true)
	local t_menuWindow = main.f_menuWindow(motif[v.section])
	local t_pos = {} --for storing current table position
	local lastNum = 0
	for i, suffix in ipairs(main.f_tableExists(main.t_sort[v.section]).menu) do
		for j, c in ipairs(main.f_strsplit('_', suffix)) do --split using "_" delimiter
			--appending the menu table
			if j == 1 then --first string after menu.itemname (either reserved one or custom submenu assignment)
				if menu[v.id].submenu[c] == nil or c == 'empty' then
					menu[v.id].submenu[c] = {}
					menu[v.id].submenu[c].title = main.f_itemnameUpper(motif[v.section]['menu_itemname_' .. suffix], motif[v.section].menu_title_uppercase == 1)
					if menu.t_itemname[c] == nil and c ~= 'empty' then
						menu[v.id].submenu[c].cursorPosY = 1
						menu[v.id].submenu[c].moveTxt = 0
						menu[v.id].submenu[c].item = 1
						menu[v.id].submenu[c].submenu = {}
						menu[v.id].submenu[c].items = {}
						menu[v.id].submenu[c].loop = menu.f_createMenu(menu[v.id].submenu[c], v.section, v.bgdef, menu[v.txt_title], false)
					end
					if not suffix:match(c .. '_') then
						table.insert(menu[v.id].items, {
							data = text:create({window = t_menuWindow}),
							itemname = c,
							displayname = motif[v.section]['menu_itemname_' .. suffix],
							paramname = 'menu_itemname_' .. suffix,
							vardata = text:create({window = t_menuWindow}),
							vardisplay = menu.f_vardisplay(c),
							selected = false,
						})
					end
				end
				t_pos = menu[v.id].submenu[c]
				t_pos.name = c
			else --following strings
				if t_pos.submenu[c] == nil or c == 'empty' then
					t_pos.submenu[c] = {}
					t_pos.submenu[c].title = main.f_itemnameUpper(motif[v.section]['menu_itemname_' .. suffix], motif[v.section].menu_title_uppercase == 1)
					if menu.t_itemname[c] == nil and c ~= 'empty' then
						t_pos.submenu[c].cursorPosY = 1
						t_pos.submenu[c].moveTxt = 0
						t_pos.submenu[c].item = 1
						t_pos.submenu[c].submenu = {}
						t_pos.submenu[c].items = {}
						t_pos.submenu[c].loop = menu.f_createMenu(t_pos.submenu[c], v.section, v.bgdef, menu[v.txt_title], false)
					end
					table.insert(t_pos.items, {
						data = text:create({window = t_menuWindow}),
						itemname = c,
						displayname = motif[v.section]['menu_itemname_' .. suffix],
						paramname = 'menu_itemname_' .. suffix,
						vardata = text:create({window = t_menuWindow}),
						vardisplay = menu.f_vardisplay(c),
						selected = false,
					})
				end
				if j > lastNum then
					t_pos = t_pos.submenu[c]
					t_pos.name = c
				end
			end
			lastNum = j
		end
	end
	if main.debugLog then main.f_printTable(menu[v.id], 'debug/t_' .. v.id .. 'Menu.txt') end
end

menu.movelistChar = 1
function menu.f_init()
	esc(false)
	togglePause(true)
	main.pauseMenu = true
	main.f_bgReset(motif.optionbgdef.bg)
	if gamemode('training') then
		sndPlay(motif.files.snd_data, motif.training_info.enter_snd[1], motif.training_info.enter_snd[2])
		main.f_bgReset(motif.trainingbgdef.bg)
		main.f_fadeReset('fadein', motif.training_info)
		menu.currentMenu = {menu.training.loop, menu.training.loop}
	else
		sndPlay(motif.files.snd_data, motif.menu_info.enter_snd[1], motif.menu_info.enter_snd[2])
		main.f_bgReset(motif.menubgdef.bg)
		main.f_fadeReset('fadein', motif.menu_info)
		--menu.menu.cursorPosY = 1
		--menu.menu.moveTxt = 0
		--menu.menu.item = 1
		menu.currentMenu = {menu.menu.loop, menu.menu.loop}
	end
end

function menu.f_run()
	local section = 'menu_info'
	local bgdef = 'menubgdef'
	if gamemode('training') then
		section = 'training_info'
		bgdef = 'trainingbgdef'
	end
	--draw overlay
	menu[section .. '_overlay']:draw()
	--Button Config
	if menu.itemname == 'keyboard' or menu.itemname == 'gamepad' then
		if menu.itemname == 'keyboard' then
			options.f_keyCfg('KeyConfig', menu.itemname, bgdef, true)
		else
			options.f_keyCfg('JoystickConfig', menu.itemname, bgdef, true)
		end
	--Command List
	elseif menu.itemname == 'commandlist' then
		menu.f_commandlistRender(section, menu.t_movelists[menu.movelistChar])
	--Menu
	else
		menu.currentMenu[1]()
	end
end

--;===========================================================
--; COMMAND LIST
--;===========================================================
local function f_commandlistData(t, str, align, col)
	local t_insert = {}
	str = str .. '<#>'
	for m1, m2 in str:gmatch('(.-)<([^%g <>]+)>') do
		if m1 ~= '' then
			table.insert(t_insert, {glyph = false, text = m1, align = align, col = col})
		end
		if not m2:match('^#[A-Za-z0-9]+$') and not m2:match('^/$') and not m2:match('^#$') then
			table.insert(t_insert, {glyph = true, text = m2, align = align, col = col})
		elseif m2:match('^#[A-Za-z0-9]+$') then
			col = color:fromHex(m2)
		elseif m2:match('^/$') then
			col = {}
		end
	end
	if align == -1 then
		for i = #t_insert, 1, -1 do
			table.insert(t, t_insert[i])
		end
	else
		for i = 1, #t_insert do
			table.insert(t, t_insert[i])
		end
	end
	return t, col
end

function menu.f_commandlistParse()
	menu.t_movelists = {}
	local t_uniqueRefs = {}
	for player, tbl in ipairs({start.p[1].t_selected, start.p[2].t_selected}) do
		for member, sel in ipairs(tbl) do
			if t_uniqueRefs[sel.ref] == nil then
				t_uniqueRefs[sel.ref] = true
				if sel.movelistLine == nil then
					sel.movelistLine = 1
				end
				if start.f_getCharData(sel.ref).commandlist == nil then
					local movelist = getCharMovelist(sel.ref)
					if movelist ~= '' then
						for k, v in main.f_sortKeys(motif.glyphs, function(t, a, b) return string.len(a) > string.len(b) end) do
							movelist = movelist:gsub(main.f_escapePattern(k), '<' .. numberToRune(v[1] + 0xe000) .. '>')
						end
						local t = {}
						local col = {}
						for line in movelist:gmatch('([^\n]*)\n?') do
							line = line:gsub('%s+$', '')
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
								subt, col = f_commandlistData(subt, m, align, col)
							end
							table.insert(t, subt)
						end
						t[#t] = nil --blank line produced by regexp matching
						start.f_getCharData(sel.ref).commandlist = t
					end
				end
				local pn = player
				if member > 1 then
					pn = pn + (member - 1) * 2
				end
				table.insert(menu.t_movelists, {
					pn = pn,
					name = start.f_getCharData(sel.ref).name,
					tbl = sel,
					commandlist = start.f_getCharData(sel.ref).commandlist,
				})
			end
		end
	end
	if menu.movelistChar > #menu.t_movelists then
		menu.movelistChar = 1
	end
	if main.debugLog then main.f_printTable(menu.t_movelists, "debug/t_movelists.txt") end
end

for _, v in ipairs({'menu_info', 'training_info'}) do
	menu[v .. '_txt_title'] = main.f_createTextImg(motif[v], 'movelist_title', {defsc = motif.defaultMenu, addX = motif[v].movelist_pos[1], addY = motif[v].movelist_pos[2]})
	menu[v .. '_txt_text'] = main.f_createTextImg(motif[v], 'movelist_text', {defsc = motif.defaultMenu, addX = motif[v].movelist_pos[1], addY = motif[v].movelist_pos[2]})
	menu[v .. '_overlay'] = main.f_createOverlay(motif[v], 'overlay')
	menu[v .. '_movelist_overlay'] = main.f_createOverlay(motif[v], 'movelist_overlay')
	--menu[v .. '_t_movelistWindow'] = {0, 0, main.SP_Localcoord[1], main.SP_Localcoord[2]}
	if motif[v].movelist_window_margins_y[1] ~= 0 or motif[v].movelist_window_margins_y[2] ~= 0 then
		local data = menu[v .. '_txt_text']
		local font_def = main.font_def[motif[v].movelist_text_font[1] .. motif[v].movelist_text_font_height]
		menu[v .. '_t_movelistWindow'] = {
			0,
			math.max(0, motif[v].movelist_pos[2] + motif[v].movelist_text_offset[2] - motif[v].movelist_window_margins_y[1]),
			motif[v].movelist_pos[1] + motif[v].movelist_text_offset[1] + motif[v].movelist_window_width,
			motif[v].movelist_pos[2] + motif[v].movelist_text_offset[2] + (motif[v].movelist_window_visibleitems - 1) * main.f_round((font_def.Size[2] + font_def.Spacing[2]) * data.scaleY + motif[v].movelist_text_spacing[2]) + motif[v].movelist_window_margins_y[2] + math.max(0, motif[v].movelist_glyphs_offset[2])
		}

	end
	menu[v .. '_txt_text']:update({window = menu[v .. '_t_movelistWindow']})
end

function menu.f_commandlistRender(section, t)
	main.f_cmdInput()
	local cmdList = {}
	if t.commandlist ~= nil then
		cmdList = t.commandlist
	else
		table.insert(cmdList, {{glyph = false, text = motif[section].movelist_text_text, align = 1, col = {}}})
	end
	if esc() or main.f_input(main.t_players, {'m'}) then
		sndPlay(motif.files.snd_data, motif[section].cancel_snd[1], motif[section].cancel_snd[2])
		menu.itemname = ''
		return
	elseif main.f_input(main.t_players, {'pal', 's'}) then
		sndPlay(motif.files.snd_data, motif[section].cursor_done_snd[1], motif[section].cursor_done_snd[2])
		menu.itemname = ''
		togglePause(false)
		main.pauseMenu = false
		menu.currentMenu[1] = menu.currentMenu[2]
		return
	elseif main.f_input(main.t_players, {'$B'}) and #menu.t_movelists > 1 then
		sndPlay(motif.files.snd_data, motif[section].cursor_move_snd[1], motif[section].cursor_move_snd[2])
		menu.movelistChar = menu.movelistChar - 1
		if menu.movelistChar < 1 then
			menu.movelistChar = #menu.t_movelists
		end
	elseif main.f_input(main.t_players, {'$F'}) and #menu.t_movelists > 1 then
		sndPlay(motif.files.snd_data, motif[section].cursor_move_snd[1], motif[section].cursor_move_snd[2])
		menu.movelistChar = menu.movelistChar + 1
		if menu.movelistChar > #menu.t_movelists then
			menu.movelistChar = 1
		end
	elseif main.f_input(main.t_players, {'$U'}) and t.tbl.movelistLine > 1 then
		sndPlay(motif.files.snd_data, motif[section].cursor_move_snd[1], motif[section].cursor_move_snd[2])
		t.tbl.movelistLine = t.tbl.movelistLine - 1
	elseif main.f_input(main.t_players, {'$D'}) and t.tbl.movelistLine <= #cmdList - motif[section].movelist_window_visibleitems then
		sndPlay(motif.files.snd_data, motif[section].cursor_move_snd[1], motif[section].cursor_move_snd[2])
		t.tbl.movelistLine = t.tbl.movelistLine + 1
	end
	--draw overlay
	menu[section .. '_movelist_overlay']:draw()
	--draw title
	menu[section .. '_txt_title']:update({text = main.f_itemnameUpper(motif[section].movelist_title_text:gsub('%%s', t.name), motif[section].movelist_title_uppercase == 1)})
	menu[section .. '_txt_title']:draw()
	--draw commands
	local i = 0
	for n = t.tbl.movelistLine, math.min(t.tbl.movelistLine + motif[section].movelist_window_visibleitems + 1, #cmdList) do
		i = i + 1
		local alignOffset = 0
		local lengthOffset = 0
		local align = 1
		local width = 0
		for k, v in ipairs(cmdList[n]) do
			if v.text ~= '' then
				alignOffset = 0
				if v.align == 0 then --center align
					alignOffset = motif[section].movelist_window_width * 0.5
				elseif v.align == -1 then --right align
					alignOffset = motif[section].movelist_window_width
				end
				if v.align ~= align then
					lengthOffset = 0
					align = v.align
				end
				local data = menu[section .. '_txt_text']
				local font_def = main.font_def[motif[section].movelist_text_font[1] .. motif[section].movelist_text_font_height]
				--render glyph
				if v.glyph and motif.glyphs_data[v.text] ~= nil then
					local scaleX = font_def.Size[2] * motif[section].movelist_text_font_scale[2] / motif.glyphs_data[v.text].info.Size[2] * motif[section].movelist_glyphs_scale[1]
					local scaleY = font_def.Size[2] * motif[section].movelist_text_font_scale[2] / motif.glyphs_data[v.text].info.Size[2] * motif[section].movelist_glyphs_scale[2]
					if v.align == -1 then
						alignOffset = alignOffset - motif.glyphs_data[v.text].info.Size[1] * scaleX
					end
					if motif.defaultMenu then main.f_disableLuaScale() end
					animSetScale(motif.glyphs_data[v.text].anim, scaleX, scaleY)
					animSetPos(
						motif.glyphs_data[v.text].anim,
						math.floor(motif[section].movelist_pos[1] + motif[section].movelist_text_offset[1] + motif[section].movelist_glyphs_offset[1] + alignOffset + lengthOffset),
						motif[section].movelist_pos[2] + motif[section].movelist_text_offset[2] + motif[section].movelist_glyphs_offset[2] + main.f_round((font_def.Size[2] + font_def.Spacing[2]) * data.scaleY + motif[section].movelist_text_spacing[2]) * (i - 1)
					)
					animSetWindow(
						motif.glyphs_data[v.text].anim,
						menu[section .. '_t_movelistWindow'][1],
						menu[section .. '_t_movelistWindow'][2],
						menu[section .. '_t_movelistWindow'][3] - menu[section .. '_t_movelistWindow'][1],
						menu[section .. '_t_movelistWindow'][4] - menu[section .. '_t_movelistWindow'][2]
					)
					--animUpdate(motif.glyphs_data[v.text].anim)
					animDraw(motif.glyphs_data[v.text].anim)
					if motif.defaultMenu then main.f_setLuaScale() end
					if k < #cmdList[n] then
						width = motif.glyphs_data[v.text].info.Size[1] * scaleX + motif[section].movelist_glyphs_spacing[1]
					end
				--render text
				else
					data:update({
						text = v.text,
						align = v.align,
						x = math.floor(motif[section].movelist_pos[1] + motif[section].movelist_text_offset[1] + alignOffset + lengthOffset),
						y = motif[section].movelist_pos[2] + motif[section].movelist_text_offset[2] + main.f_round((font_def.Size[2] + font_def.Spacing[2]) * data.scaleY + motif[section].movelist_text_spacing[2]) * (i - 1),
						r = v.col.r or motif[section].movelist_text_font[4],
						g = v.col.g or motif[section].movelist_text_font[5],
						b = v.col.b or motif[section].movelist_text_font[6],
					})
					data:draw()
					if k < #cmdList[n] then
						width = fontGetTextWidth(main.font[data.font .. data.height], v.text) * motif[section].movelist_text_font_scale[1] + motif[section].movelist_text_spacing[1]
					end
				end
				if v.align == 0 then
					lengthOffset = lengthOffset + width / 2
				elseif v.align == -1 then
					lengthOffset = lengthOffset - width
				else
					lengthOffset = lengthOffset + width
				end
			end
		end
	end
	--draw scroll arrows
	if #cmdList > motif[section].movelist_window_visibleitems then
		if t.tbl.movelistLine > 1 then
			animUpdate(motif[section].movelist_arrow_up_data)
			animDraw(motif[section].movelist_arrow_up_data)
		end
		if t.tbl.movelistLine <= #cmdList - motif[section].movelist_window_visibleitems then
			animUpdate(motif[section].movelist_arrow_down_data)
			animDraw(motif[section].movelist_arrow_down_data)
		end
	end
end

return menu

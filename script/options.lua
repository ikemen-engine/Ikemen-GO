
local options = {}

--;===========================================================
--; LOAD DATA
--;===========================================================

-- Data loading from lifebar
local file = io.open(motif.files.fight,"r")
local s_lifebar = file:read("*all")
file:close()
local roundsNum = tonumber(s_lifebar:match('match.wins%s*=%s*(%d+)'))
options.framespercount = tonumber(s_lifebar:match('framespercount%s*=%s*(%d+)'))

main.f_printTable(config, "debug/config.txt")

--;===========================================================
--; COMMON
--;===========================================================
local modified = 0
local needReload = 0

local windowBox = animNew(main.fadeSff, '0,0, 0,0, -1')
animSetTile(windowBox, 1, 1)
animSetAlpha(windowBox, motif.option_info.menu_boxbackground_alpha[1], motif.option_info.menu_boxbackground_alpha[2])
animUpdate(windowBox)

--return string depending on bool
function options.f_boolDisplay(bool, t, f)
	t = t or motif.option_info.menu_itemname_yes
	f = f or motif.option_info.menu_itemname_no
	if bool == true then
		return t
	else
		return f
	end
end

--return string depending on int
function options.f_intDisplay(bool, t, f)
	t = t or motif.option_info.menu_itemname_yes
	f = f or motif.option_info.menu_itemname_no
	if bool == 1 then
		return t
	else
		return f
	end
end

--return correct precision
function options.f_precision(v, decimal)
	return tonumber(string.format(decimal, v))
end

function options.f_saveCfg()
	--Data saving to config.json
	local file = io.open("data/config.json","w+")
	file:write(json.encode(config, {indent = true}))
	file:close()
	--Data saving to lifebar
	s_lifebar = s_lifebar:gsub('match.wins%s*=%s*%d+', 'match.wins = ' .. roundsNum)
	file = io.open(motif.files.fight,"w+")
	file:write(s_lifebar)
	file:close()
	--Reload lifebar
	loadLifebar(motif.files.fight)
	--Reload game if needed
	if needReload == 1 then
		main.f_warning(main.f_extractText(motif.warning_info.text_reload), motif.option_info, motif.optionbgdef)
		os.exit()
	end
end

function options.f_keyDefault()
	config.KeyConfig[1].Buttons[1] = 'UP'
	config.KeyConfig[1].Buttons[2] = 'DOWN'
	config.KeyConfig[1].Buttons[3] = 'LEFT'
	config.KeyConfig[1].Buttons[4] = 'RIGHT'
	config.KeyConfig[1].Buttons[5] = 'z'
	config.KeyConfig[1].Buttons[6] = 'x'
	config.KeyConfig[1].Buttons[7] = 'c'
	config.KeyConfig[1].Buttons[8] = 'a'
	config.KeyConfig[1].Buttons[9] = 's'
	config.KeyConfig[1].Buttons[10] = 'd'
	config.KeyConfig[1].Buttons[11] = 'RETURN'
	config.KeyConfig[1].Buttons[12] = 'q'
	config.KeyConfig[1].Buttons[13] = 'w'
	config.KeyConfig[2].Buttons[1] = 't'
	config.KeyConfig[2].Buttons[2] = 'g'
	config.KeyConfig[2].Buttons[3] = 'f'
	config.KeyConfig[2].Buttons[4] = 'h'
	config.KeyConfig[2].Buttons[5] = 'j'
	config.KeyConfig[2].Buttons[6] = 'k'
	config.KeyConfig[2].Buttons[7] = 'l'
	config.KeyConfig[2].Buttons[8] = 'u'
	config.KeyConfig[2].Buttons[9] = 'i'
	config.KeyConfig[2].Buttons[10] = 'o'
	config.KeyConfig[2].Buttons[11] = 'RSHIFT'
	config.KeyConfig[2].Buttons[12] = 'LEFTBRACKET'
	config.KeyConfig[2].Buttons[13] = 'RIGHTBRACKET'
	config.JoystickConfig[1].Buttons[1] = '-7'
	config.JoystickConfig[1].Buttons[2] = '-8'
	config.JoystickConfig[1].Buttons[3] = '-5'
	config.JoystickConfig[1].Buttons[4] = '-6'
	config.JoystickConfig[1].Buttons[5] = '0'
	config.JoystickConfig[1].Buttons[6] = '1'
	config.JoystickConfig[1].Buttons[7] = '4'
	config.JoystickConfig[1].Buttons[8] = '2'
	config.JoystickConfig[1].Buttons[9] = '3'
	config.JoystickConfig[1].Buttons[10] = '5'
	config.JoystickConfig[1].Buttons[11] = '7'
	config.JoystickConfig[1].Buttons[12] = '6'
	config.JoystickConfig[1].Buttons[13] = '8'
	config.JoystickConfig[2].Buttons[1] = '-7'
	config.JoystickConfig[2].Buttons[2] = '-8'
	config.JoystickConfig[2].Buttons[3] = '-5'
	config.JoystickConfig[2].Buttons[4] = '-6'
	config.JoystickConfig[2].Buttons[5] = '0'
	config.JoystickConfig[2].Buttons[6] = '1'
	config.JoystickConfig[2].Buttons[7] = '4'
	config.JoystickConfig[2].Buttons[8] = '2'
	config.JoystickConfig[2].Buttons[9] = '3'
	config.JoystickConfig[2].Buttons[10] = '5'
	config.JoystickConfig[2].Buttons[11] = '7'
	config.JoystickConfig[2].Buttons[12] = '6'
	config.JoystickConfig[2].Buttons[13] = '8'
end

function options.f_menuCommon1(cursorPosY, moveTxt, item, t)
	if commandGetState(main.p1Cmd, 'u') then
		sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
		item = item - 1
		if t[item] ~= nil and t[item].itemname == 'empty' then
			item = item - 1
		end
	elseif commandGetState(main.p1Cmd, 'd') then
		sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
		item = item + 1
		if t[item] ~= nil and t[item].itemname == 'empty' then
			item = item + 1
		end
	end
	--cursor position calculation
	if item < 1 then
		item = #t
		if #t > motif.option_info.menu_window_visibleitems then
			cursorPosY = motif.option_info.menu_window_visibleitems
		else
			cursorPosY = #t
		end
	elseif item > #t then
		item = 1
		cursorPosY = 1
	elseif commandGetState(main.p1Cmd, 'u') and cursorPosY > 1 then
		cursorPosY = cursorPosY - 1
		if t[cursorPosY] ~= nil and t[cursorPosY].itemname == 'empty' then
			cursorPosY = cursorPosY - 1
		end
	elseif commandGetState(main.p1Cmd, 'd') and cursorPosY < motif.option_info.menu_window_visibleitems then
		cursorPosY = cursorPosY + 1
		if t[cursorPosY] ~= nil and t[cursorPosY].itemname == 'empty' then
			cursorPosY = cursorPosY + 1
		end
	end
	if cursorPosY == motif.option_info.menu_window_visibleitems then
		moveTxt = (item - motif.option_info.menu_window_visibleitems) * motif.option_info.menu_item_spacing[2]
	elseif cursorPosY == 1 then
		moveTxt = (item - 1) * motif.option_info.menu_item_spacing[2]
	end
	return cursorPosY, moveTxt, item
end

local txt_title = main.f_createTextImg(
	motif.font_data[motif.option_info.title_font[1]],
	motif.option_info.title_font[2],
	motif.option_info.title_font[3],
	"",
	motif.option_info.title_offset[1],
	motif.option_info.title_offset[2],
	motif.option_info.title_font_scale[1],
	motif.option_info.title_font_scale[2],
	motif.option_info.title_font[4],
	motif.option_info.title_font[5],
	motif.option_info.title_font[6]
)
function options.f_menuCommon2(cursorPosY, moveTxt, item, t)
	--draw clearcolor
	animDraw(motif.optionbgdef.bgclearcolor_data)
	--draw layerno = 0 backgrounds
	main.f_drawBG(motif.optionbgdef.bg_data, motif.optionbgdef.bg, 0, motif.optionbgdef.timer)
	--draw menu box
	if motif.option_info.menu_boxbackground_visible == 1 then
		if #t > motif.option_info.menu_window_visibleitems and moveTxt == (#t - motif.option_info.menu_window_visibleitems) * motif.option_info.menu_item_spacing[2] then
			animSetWindow(
				windowBox,
				motif.option_info.menu_pos[1] + motif.option_info.menu_boxcursor_coords[1],
				motif.option_info.menu_pos[2] + motif.option_info.menu_boxcursor_coords[2],
				motif.option_info.menu_boxcursor_coords[3] - motif.option_info.menu_boxcursor_coords[1] + 1,
				motif.option_info.menu_window_visibleitems * (motif.option_info.menu_boxcursor_coords[4] - motif.option_info.menu_boxcursor_coords[2] + 1)
			)
		else
			animSetWindow(
				windowBox,
				motif.option_info.menu_pos[1] + motif.option_info.menu_boxcursor_coords[1],
				motif.option_info.menu_pos[2] + motif.option_info.menu_boxcursor_coords[2],
				motif.option_info.menu_boxcursor_coords[3] - motif.option_info.menu_boxcursor_coords[1] + 1,
				#t * (motif.option_info.menu_boxcursor_coords[4] - motif.option_info.menu_boxcursor_coords[2] + 1)
			)
		end
		animDraw(windowBox)
	end
	--draw title
	textImgDraw(txt_title)
	--draw menu items
	for i = 1, #t do
		if i > item - cursorPosY then
			if i == item then
				textImgDraw(main.f_updateTextImg(
					t[i].data,
					motif.font_data[motif.option_info.menu_item_active_font[1]],
					motif.option_info.menu_item_active_font[2],
					motif.option_info.menu_item_active_font[3],
					t[i].displayname,
					motif.option_info.menu_pos[1],
					motif.option_info.menu_pos[2] + (i - 1) * motif.option_info.menu_item_spacing[2] - moveTxt,
					motif.option_info.menu_item_active_font_scale[1],
					motif.option_info.menu_item_active_font_scale[2],
					motif.option_info.menu_item_active_font[4],
					motif.option_info.menu_item_active_font[5],
					motif.option_info.menu_item_active_font[6]
				))
				if t[i].vardata ~= nil then
					textImgDraw(main.f_updateTextImg(
						t[i].vardata,
						motif.font_data[motif.option_info.menu_item_value_active_font[1]],
						motif.option_info.menu_item_value_active_font[2],
						motif.option_info.menu_item_value_active_font[3],
						t[i].vardisplay,
						motif.option_info.menu_pos[1] + motif.option_info.menu_item_spacing[1],
						motif.option_info.menu_pos[2] + (i - 1) * motif.option_info.menu_item_spacing[2] - moveTxt,
						motif.option_info.menu_item_value_active_font_scale[1],
						motif.option_info.menu_item_value_active_font_scale[2],
						motif.option_info.menu_item_value_active_font[4],
						motif.option_info.menu_item_value_active_font[5],
						motif.option_info.menu_item_value_active_font[6]
					))
				end
			else
				textImgDraw(main.f_updateTextImg(
					t[i].data,
					motif.font_data[motif.option_info.menu_item_font[1]],
					motif.option_info.menu_item_font[2],
					motif.option_info.menu_item_font[3],
					t[i].displayname,
					motif.option_info.menu_pos[1],
					motif.option_info.menu_pos[2] + (i - 1) * motif.option_info.menu_item_spacing[2] - moveTxt,
					motif.option_info.menu_item_font_scale[1],
					motif.option_info.menu_item_font_scale[2],
					motif.option_info.menu_item_font[4],
					motif.option_info.menu_item_font[5],
					motif.option_info.menu_item_font[6]
				))
				if t[i].vardata ~= nil then
					textImgDraw(main.f_updateTextImg(
						t[i].vardata,
						motif.font_data[motif.option_info.menu_item_value_font[1]],
						motif.option_info.menu_item_value_font[2],
						motif.option_info.menu_item_value_font[3],
						t[i].vardisplay,
						motif.option_info.menu_pos[1] + motif.option_info.menu_item_spacing[1],
						motif.option_info.menu_pos[2] + (i - 1) * motif.option_info.menu_item_spacing[2] - moveTxt,
						motif.option_info.menu_item_value_font_scale[1],
						motif.option_info.menu_item_value_font_scale[2],
						motif.option_info.menu_item_value_font[4],
						motif.option_info.menu_item_value_font[5],
						motif.option_info.menu_item_value_font[6]
					))
				end
			end
		end
	end
	--draw menu cursor
	if motif.option_info.menu_boxcursor_visible == 1 then
		animSetWindow(
			main.cursorBox,
			motif.option_info.menu_pos[1] + motif.option_info.menu_boxcursor_coords[1],
			motif.option_info.menu_pos[2] + motif.option_info.menu_boxcursor_coords[2] + (cursorPosY - 1) * motif.option_info.menu_item_spacing[2],
			motif.option_info.menu_boxcursor_coords[3] - motif.option_info.menu_boxcursor_coords[1] + 1,
			motif.option_info.menu_boxcursor_coords[4] - motif.option_info.menu_boxcursor_coords[2] + 1
		)
		main.f_dynamicAlpha(main.cursorBox, 10,40,2, 255,255,0)
		animDraw(main.cursorBox)
	end
	--draw layerno = 1 backgrounds
	main.f_drawBG(motif.optionbgdef.bg_data, motif.optionbgdef.bg, 1, motif.optionbgdef.timer)
	--draw fadein
	animDraw(motif.option_info.fadein_data)
	animUpdate(motif.option_info.fadein_data)
	--update timer
	motif.optionbgdef.timer = motif.optionbgdef.timer + 1
	--end loop
	main.f_cmdInput()
	refresh()
end

--;===========================================================
--; MAIN LOOP
--;===========================================================
local t_mainCfg = {
	{data = textImgNew(), itemname = 'arcadesettings', displayname = motif.option_info.menu_itemname_main_arcade},
	{data = textImgNew(), itemname = 'videosettings', displayname = motif.option_info.menu_itemname_main_video},
	{data = textImgNew(), itemname = 'audiosettings', displayname = motif.option_info.menu_itemname_main_audio},
	{data = textImgNew(), itemname = 'inputsettings', displayname = motif.option_info.menu_itemname_main_input},
	{data = textImgNew(), itemname = 'gameplaysettings', displayname = motif.option_info.menu_itemname_main_gameplay},
	{data = textImgNew(), itemname = 'enginesettings', displayname = motif.option_info.menu_itemname_main_engine},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'quicklaunch', displayname = motif.option_info.menu_itemname_engine_quicklaunch, vardata = textImgNew(), vardisplay = options.f_intDisplay(config.QuickLaunch, motif.option_info.menu_itemname_yes, motif.option_info.menu_itemname_no)},
	{data = textImgNew(), itemname = 'portchange', displayname = motif.option_info.menu_itemname_main_port, vardata = textImgNew(), vardisplay = getListenPort()},
	{data = textImgNew(), itemname = 'defaultvalues', displayname = motif.option_info.menu_itemname_main_default},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'save', displayname = motif.option_info.menu_itemname_main_save},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_main_back},
}
t_mainCfg = main.f_cleanTable(t_mainCfg)

function options.f_mainCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_mainCfg
	textImgSetText(txt_title, motif.option_info.title_text_main)
	main.f_resetBG(motif.option_info, motif.optionbgdef)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			main.f_resetBG(motif.title_info, motif.titlebgdef)
			break
		--Port Change
		elseif t[item].itemname == 'portchange' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			local port = main.f_input(main.f_extractText(motif.option_info.input_text_port), motif.option_info, motif.optionbgdef, 'string')
			if tonumber(port) ~= nil then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				setListenPort(port)
			else
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			end
			t[item].vardisplay = getListenPort()
			modified = 1
		-- Quick Launch
		elseif t[item].itemname == 'quicklaunch' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.QuickLaunch == 1 then
				config.QuickLaunch = 0
			else
				config.QuickLaunch = 1
			end
			t[item].vardisplay = options.f_intDisplay(config.QuickLaunch, motif.option_info.menu_itemname_yes, motif.option_info.menu_itemname_no)
			modified = 1
		elseif main.f_btnPalNo(main.p1Cmd) > 0 then
			--Arcade Settings
			if t[item].itemname == 'arcadesettings' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				options.f_arcadeCfg()
			--Gameplay Settings
			elseif t[item].itemname == 'gameplaysettings' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				options.f_gameplayCfg()
			--Video Settings
			elseif t[item].itemname == 'videosettings' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				options.f_videoCfg()
			elseif t[item].itemname == 'audiosettings' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				options.f_audioCfg()
			--Engine Settings
				elseif t[item].itemname == 'enginesettings' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				options.f_engineCfg()
			--Input Settings
			elseif t[item].itemname == 'inputsettings' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				options.f_inputCfg()
			--Default Values
			elseif t[item].itemname == 'defaultvalues' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				config.HelperMax = 56
				config.PlayerProjectileMax = 50
				config.ExplodMax = 256
				config.AfterImageMax = 8
				config['Attack.LifeToPowerMul'] = 0.7
				config['GetHit.LifeToPowerMul'] = 0.6
				config.Width = 640
				config.Height = 480
				config['Super.TargetDefenceMul'] = 1.5
				config.LifebarFontScale = 1
				--config.System = 'script/main.lua'
				options.f_keyDefault()
				--config.Motif = 'data/system.def'
				config.SimulMode = true
				config.LifeMul = 100
				config.Team1VS2Life = 120
				config.TurnsRecoveryRate = 300
				config.ZoomActive = false
				config.ZoomMin = 0.75
				config.ZoomMax = 1.1
				config.ZoomSpeed = 1.0
				config.AIRandomColor = true
				config.RoundTime = 99
				config.NumTurns = 4
				config.NumSimul = 4
				config.NumTag = 4
				config.Difficulty = 8
				config.Credits = 10
				setListenPort(7500)
				config.ContSelection = true
				config.AIRamping = true
				config.AutoGuard = false
				config.TeamPowerShare = false
				config.TeamLifeShare = false
				config.Fullscreen = false
				--lifebar
				roundsNum = 2
				modified = 1
				needReload = 1
			-- Save
			elseif t[item].itemname == 'save' then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				if modified == 1 then
					options.f_saveCfg()
				end
				main.f_resetBG(motif.title_info, motif.titlebgdef)
				break
			-- Back
			elseif t[item].itemname == 'back' then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				main.f_resetBG(motif.title_info, motif.titlebgdef)
				break
			end
		end
		options.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; ARCADE SETTINGS
--;===========================================================
local t_arcadeCfg = {
	{data = textImgNew(), itemname = 'roundstowin', displayname = motif.option_info.menu_itemname_arcade_roundstowin, vardata = textImgNew(), vardisplay = roundsNum},
	{data = textImgNew(), itemname = 'roundtime', displayname = motif.option_info.menu_itemname_arcade_roundtime, vardata = textImgNew(), vardisplay = config.RoundTime},
	{data = textImgNew(), itemname = 'difficulty', displayname = motif.option_info.menu_itemname_arcade_difficulty, vardata = textImgNew(), vardisplay = config.Difficulty},
	{data = textImgNew(), itemname = 'credits', displayname = motif.option_info.menu_itemname_arcade_credits, vardata = textImgNew(), vardisplay = config.Credits},
	{data = textImgNew(), itemname = 'charchange', displayname = motif.option_info.menu_itemname_arcade_charchange, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.ContSelection)},
	{data = textImgNew(), itemname = 'airamping', displayname = motif.option_info.menu_itemname_arcade_airamping, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.AIRamping)},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_arcade_back},
}
t_arcadeCfg = main.f_cleanTable(t_arcadeCfg)

function options.f_arcadeCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_arcadeCfg
	textImgSetText(txt_title, motif.option_info.title_text_arcade)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		--Rounds to Win
		elseif t[item].itemname == 'roundstowin' then
			if commandGetState(main.p1Cmd, 'r') and roundsNum < 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				roundsNum = roundsNum + 1
				t[item].vardisplay = roundsNum
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and roundsNum > 1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				roundsNum = roundsNum - 1
				t[item].vardisplay = roundsNum
				modified = 1
			end
		--Round Time
		elseif t[item].itemname == 'roundtime' then
			if commandGetState(main.p1Cmd, 'r') and config.RoundTime < 1000 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.RoundTime = config.RoundTime + 1
				t[item].vardisplay = config.RoundTime
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.RoundTime > -2 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.RoundTime = config.RoundTime - 1
				t[item].vardisplay = config.RoundTime
				modified = 1
			end
		--Difficulty level
		elseif t[item].itemname == 'difficulty' then
			if commandGetState(main.p1Cmd, 'r') and config.Difficulty < 8 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.Difficulty = config.Difficulty + 1
				t[item].vardisplay = config.Difficulty
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.Difficulty > 1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.Difficulty = config.Difficulty - 1
				t[item].vardisplay = config.Difficulty
				modified = 1
			end
		--Credits
		elseif t[item].itemname == 'credits' then
			if commandGetState(main.p1Cmd, 'r') and config.Credits < 99 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.Credits = config.Credits + 1
				t[item].vardisplay = config.Credits
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.Credits > 1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.Credits = config.Credits - 1
				t[item].vardisplay = config.Credits
				modified = 1
			end
		--Char change at Continue
		elseif t[item].itemname == 'charchange' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.ContSelection then
				config.ContSelection = false
			else
				config.ContSelection = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.ContSelection)
			modified = 1
		--AI ramping
		elseif t[item].itemname == 'airamping' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AIRamping then
				config.AIRamping = false
			else
				config.AIRamping = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.AIRamping)
			modified = 1
		--Back
		elseif t[item].itemname == 'back' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		end
		options.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; GAMEPLAY SETTINGS
--;===========================================================
local t_gameplayCfg = {
	{data = textImgNew(), itemname = 'lifemul', displayname = motif.option_info.menu_itemname_gameplay_lifemul, vardata = textImgNew(), vardisplay = config.LifeMul .. '%'},
	{data = textImgNew(), itemname = 'autoguard', displayname = motif.option_info.menu_itemname_gameplay_autoguard, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.AutoGuard)},
	{data = textImgNew(), itemname = 'team1vs2life', displayname = motif.option_info.menu_itemname_gameplay_team1vs2life, vardata = textImgNew(), vardisplay = config.Team1VS2Life},
	{data = textImgNew(), itemname = 'turnsrecoveryrate', displayname = motif.option_info.menu_itemname_gameplay_turnsrecoveryrate, vardata = textImgNew(), vardisplay = config.TurnsRecoveryRate},
	{data = textImgNew(), itemname = 'teampowershare', displayname = motif.option_info.menu_itemname_gameplay_teampowershare, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.TeamPowerShare)},
	{data = textImgNew(), itemname = 'teamlifeshare', displayname = motif.option_info.menu_itemname_gameplay_teamlifeshare, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.TeamLifeShare)},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'advancedGameplaySettings', displayname = motif.option_info.menu_itemname_gameplay_advanced},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_gameplay_back},
}
t_gameplayCfg = main.f_cleanTable(t_gameplayCfg)

function options.f_gameplayCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_gameplayCfg
	textImgSetText(txt_title, motif.option_info.title_text_gameplay)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		--Life
		elseif t[item].itemname == 'lifemul' then
			if commandGetState(main.p1Cmd, 'r') and config.LifeMul < 300 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.LifeMul = config.LifeMul + 10
				t[item].vardisplay = config.LifeMul .. '%'
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.LifeMul > 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.LifeMul = config.LifeMul - 10
				t[item].vardisplay = config.LifeMul .. '%'
				modified = 1
			end
		--Auto-Guard
		elseif t[item].itemname == 'autoguard' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AutoGuard then
				config.AutoGuard = false
			else
				config.AutoGuard = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.AutoGuard)
			modified = 1
		--1P Vs Team Life
		elseif t[item].itemname == 'team1vs2life' then
			if commandGetState(main.p1Cmd, 'r') and config.Team1VS2Life < 3000 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.Team1VS2Life = config.Team1VS2Life + 10
				t[item].vardisplay = config.Team1VS2Life
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.Team1VS2Life > 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.Team1VS2Life = config.Team1VS2Life - 10
				t[item].vardisplay = config.Team1VS2Life
				modified = 1
			end
		--Turns HP Recovery
		elseif t[item].itemname == 'turnsrecoveryrate' then
			if commandGetState(main.p1Cmd, 'r') and config.TurnsRecoveryRate < 3000 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.TurnsRecoveryRate = config.TurnsRecoveryRate + 10
				t[item].vardisplay = config.TurnsRecoveryRate
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.TurnsRecoveryRate > 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.TurnsRecoveryRate = config.TurnsRecoveryRate - 10
				t[item].vardisplay = config.TurnsRecoveryRate
				modified = 1
			end
		--Team Power Share
		elseif t[item].itemname == 'teampowershare' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.TeamPowerShare then
				config.TeamPowerShare = false
			else
				config.TeamPowerShare = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.TeamPowerShare)
			modified = 1
		--Team Life Share
		elseif t[item].itemname == 'teamlifeshare' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.TeamLifeShare then
				config.TeamLifeShare = false
			else
				config.TeamLifeShare = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.TeamLifeShare)
			modified = 1
		-- Advanced settings
		elseif t[item].itemname == 'advancedGameplaySettings' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
			options.f_advGameplayCfg()
		--Back
		elseif t[item].itemname == 'back' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		end
		options.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end


--;===========================================================
--; ADVANCED GAMEPLAY SETTINGS
--;===========================================================
local t_advGameplayCfg = {
	{data = textImgNew(), itemname = 'attackpowermul', displayname = motif.option_info.menu_itemname_gameplay_attackpowermul, vardata = textImgNew(), vardisplay = config['Attack.LifeToPowerMul']},
	{data = textImgNew(), itemname = 'gethitpowermul', displayname = motif.option_info.menu_itemname_gameplay_gethitpowermul, vardata = textImgNew(), vardisplay = config['GetHit.LifeToPowerMul']},
	{data = textImgNew(), itemname = 'superdefencemul', displayname = motif.option_info.menu_itemname_gameplay_superdefencemul, vardata = textImgNew(), vardisplay = config['Super.TargetDefenceMul']},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'numturns', displayname = motif.option_info.menu_itemname_gameplay_numturns, vardata = textImgNew(), vardisplay = config.NumTurns},
	{data = textImgNew(), itemname = 'numsimul', displayname = motif.option_info.menu_itemname_gameplay_numsimul, vardata = textImgNew(), vardisplay = config.NumSimul},
	{data = textImgNew(), itemname = 'numtag', displayname = motif.option_info.menu_itemname_gameplay_numtag, vardata = textImgNew(), vardisplay = config.NumTag},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_gameplay_back},
}
t_advGameplayCfg = main.f_cleanTable(t_advGameplayCfg)

function options.f_advGameplayCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_advGameplayCfg
	textImgSetText(txt_title, motif.option_info.title_text_advgameplay)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_gameplay)
			break
		--Life
		elseif t[item].itemname == 'lifemul' then
			if commandGetState(main.p1Cmd, 'r') and config.LifeMul < 300 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.LifeMul = config.LifeMul + 10
				t[item].vardisplay = config.LifeMul .. '%'
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.LifeMul > 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.LifeMul = config.LifeMul - 10
				t[item].vardisplay = config.LifeMul .. '%'
				modified = 1
			end
		--Auto-Guard
		elseif t[item].itemname == 'autoguard' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AutoGuard then
				config.AutoGuard = false
			else
				config.AutoGuard = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.AutoGuard)
			modified = 1
		--Attack.LifeToPowerMul
		elseif t[item].itemname == 'attackpowermul' then
			if commandGetState(main.p1Cmd, 'r') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['Attack.LifeToPowerMul'] = options.f_precision(config['Attack.LifeToPowerMul'] + 0.1, '%.01f')
				t[item].vardisplay = config['Attack.LifeToPowerMul']
				modified = 1
				needReload = 1
			elseif commandGetState(main.p1Cmd, 'l') and config['Attack.LifeToPowerMul'] > 0.1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['Attack.LifeToPowerMul'] = options.f_precision(config['Attack.LifeToPowerMul'] - 0.1, '%.01f')
				t[item].vardisplay = config['Attack.LifeToPowerMul']
				modified = 1
				needReload = 1
			end
		--GetHit.LifeToPowerMul
		elseif t[item].itemname == 'gethitpowermul' then
			if commandGetState(main.p1Cmd, 'r') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['GetHit.LifeToPowerMul'] = options.f_precision(config['GetHit.LifeToPowerMul'] + 0.1, '%.01f')
				t[item].vardisplay = config['GetHit.LifeToPowerMul']
				modified = 1
				needReload = 1
			elseif commandGetState(main.p1Cmd, 'l') and config['GetHit.LifeToPowerMul'] > 0.1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['GetHit.LifeToPowerMul'] = options.f_precision(config['GetHit.LifeToPowerMul'] - 0.1, '%.01f')
				t[item].vardisplay = config['GetHit.LifeToPowerMul']
				modified = 1
				needReload = 1
			end
		--Super.TargetDefenceMul
		elseif t[item].itemname == 'superdefencemul' then
			if commandGetState(main.p1Cmd, 'r') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['Super.TargetDefenceMul'] = options.f_precision(config['Super.TargetDefenceMul'] + 0.1, '%.01f')
				t[item].vardisplay = config['Super.TargetDefenceMul']
				modified = 1
				needReload = 1
			elseif commandGetState(main.p1Cmd, 'l') and config['Super.TargetDefenceMul'] > 0.1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['Super.TargetDefenceMul'] = options.f_precision(config['Super.TargetDefenceMul'] - 0.1, '%.01f')
				t[item].vardisplay = config['Super.TargetDefenceMul']
				modified = 1
				needReload = 1
			end
		--1P Vs Team Life
		elseif t[item].itemname == 'team1vs2life' then
			if commandGetState(main.p1Cmd, 'r') and config.Team1VS2Life < 3000 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.Team1VS2Life = config.Team1VS2Life + 10
				t[item].vardisplay = config.Team1VS2Life
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.Team1VS2Life > 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.Team1VS2Life = config.Team1VS2Life - 10
				t[item].vardisplay = config.Team1VS2Life
				modified = 1
			end
		--Turns HP Recovery
		elseif t[item].itemname == 'turnsrecoveryrate' then
			if commandGetState(main.p1Cmd, 'r') and config.TurnsRecoveryRate < 3000 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.TurnsRecoveryRate = config.TurnsRecoveryRate + 10
				t[item].vardisplay = config.TurnsRecoveryRate
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.TurnsRecoveryRate > 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.TurnsRecoveryRate = config.TurnsRecoveryRate - 10
				t[item].vardisplay = config.TurnsRecoveryRate
				modified = 1
			end
		--Team Power Share
		elseif t[item].itemname == 'teampowershare' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.TeamPowerShare then
				config.TeamPowerShare = false
			else
				config.TeamPowerShare = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.TeamPowerShare)
			modified = 1
		--Team Life Share
		elseif t[item].itemname == 'teamlifeshare' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.TeamLifeShare then
				config.TeamLifeShare = false
			else
				config.TeamLifeShare = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.TeamLifeShare)
			modified = 1
		--Turns Limit
		elseif t[item].itemname == 'numturns' then
			if commandGetState(main.p1Cmd, 'r') and config.NumTurns < 8 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.NumTurns = config.NumTurns + 1
				t[item].vardisplay = config.NumTurns
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.NumTurns > 1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.NumTurns = config.NumTurns - 1
				t[item].vardisplay = config.NumTurns
				modified = 1
			end
		--Simul Limit
		elseif t[item].itemname == 'numsimul' then
			if commandGetState(main.p1Cmd, 'r') and config.NumSimul < 4 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.NumSimul = config.NumSimul + 1
				t[item].vardisplay = config.NumSimul
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.NumSimul > 2 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.NumSimul = config.NumSimul - 1
				t[item].vardisplay = config.NumSimul
				modified = 1
			end
		--Tag Limit
		elseif t[item].itemname == 'numtag' then
			if commandGetState(main.p1Cmd, 'r') and config.NumTag < 4 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.NumTag = config.NumTag + 1
				t[item].vardisplay = config.NumTag
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.NumTag > 2 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.NumTag = config.NumTag - 1
				t[item].vardisplay = config.NumTag
				modified = 1
			end
		--Back
		elseif t[item].itemname == 'back' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_gameplay)
			break
		end
		options.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; VIDEO SETTINGS
--;===========================================================
local t_shaderNames = {}
t_shaderNames[0] = "No shader"
t_shaderNames[1] = "hqx2"
t_shaderNames[2] = "hqx4"

local t_videoCfg = {
	{data = textImgNew(), itemname = 'resolution', displayname = motif.option_info.menu_itemname_video_resolution, vardata = textImgNew(), vardisplay = config.Width .. 'x' .. config.Height},
	{data = textImgNew(), itemname = 'fullscreen', displayname = motif.option_info.menu_itemname_video_fullscreen, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.Fullscreen)},
	{data = textImgNew(), itemname = 'airandomcolor', displayname = motif.option_info.menu_itemname_video_aipalette, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.AIRandomColor, motif.option_info.menu_itemname_video_aipalette_random, motif.option_info.menu_itemname_video_aipalette_default)},
	{data = textImgNew(), itemname = 'postprocessingshader', displayname = "Shader", vardata = textImgNew(), vardisplay = t_shaderNames[config.PostProcessingShader]},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_video_back},
}
t_videoCfg = main.f_cleanTable(t_videoCfg)

function options.f_videoCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_videoCfg
	textImgSetText(txt_title, motif.option_info.title_text_video)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		--Resolution
		elseif t[item].itemname == 'resolution' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			options.f_resCfg()
			t[item].vardisplay = config.Width .. 'x' .. config.Height
		--Fullscreen
		elseif t[item].itemname == 'fullscreen' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.Fullscreen then
				config.Fullscreen = false
			else
				config.Fullscreen = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.Fullscreen)
			modified = 1
			needReload = 1
		--AI Palette
		elseif t[item].itemname == 'airandomcolor' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AIRandomColor then
				config.AIRandomColor = false
			else
				config.AIRandomColor = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.AIRandomColor, motif.option_info.menu_itemname_video_aipalette_random, motif.option_info.menu_itemname_video_aipalette_default)
			modified = 1
		-- Postprocessing
		elseif t[item].itemname == 'postprocessingshader' then
			if commandGetState(main.p1Cmd, 'r') and config.PostProcessingShader < #t_shaderNames then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.PostProcessingShader = config.PostProcessingShader + 1
				modified = 1
				needReload = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.PostProcessingShader > 0 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.PostProcessingShader = config.PostProcessingShader - 1
				modified = 1
				needReload = 1
			end
			t[item].vardisplay = t_shaderNames[config.PostProcessingShader]
		-- Back
		elseif t[item].itemname == 'back' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		end
		options.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; RESOLUTION SETTINGS
--;===========================================================
local t_resCfg = {
	{data = textImgNew(), x = 320,  y = 240, displayname = motif.option_info.menu_itemname_res_320x240},
	{data = textImgNew(), x = 640,  y = 480, displayname = motif.option_info.menu_itemname_res_640x480},
	{data = textImgNew(), x = 1280, y = 960, displayname = motif.option_info.menu_itemname_res_1280x960},
	{data = textImgNew(), x = 1600, y = 1200, displayname = motif.option_info.menu_itemname_res_1600x1200},
	{data = textImgNew(), x = 960,  y = 720, displayname = motif.option_info.menu_itemname_res_960x720},
	{data = textImgNew(), x = 1200, y = 900, displayname = motif.option_info.menu_itemname_res_1200x900},
	{data = textImgNew(), x = 1440, y = 1080, displayname = motif.option_info.menu_itemname_res_1440x1080},
	{data = textImgNew(), x = 1280, y = 720, displayname = motif.option_info.menu_itemname_res_1280x720},
	{data = textImgNew(), x = 1600, y = 900, displayname = motif.option_info.menu_itemname_res_1600x900},
	{data = textImgNew(), x = 1920, y = 1080, displayname = motif.option_info.menu_itemname_res_1920x1080},
	{data = textImgNew(), x = 2560, y = 1440, displayname = motif.option_info.menu_itemname_res_2560x1440},
	{data = textImgNew(), x = 3840, y = 2160, displayname = motif.option_info.menu_itemname_res_3840x2160},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'custom', displayname = motif.option_info.menu_itemname_res_custom},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_res_back},
}
t_resCfg = main.f_cleanTable(t_resCfg)

function options.f_resCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_resCfg
	textImgSetText(txt_title, motif.option_info.title_text_res)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_video)
			break
		elseif main.f_btnPalNo(main.p1Cmd) > 0 then
			--Back
			if t[item].itemname == 'back' then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				textImgSetText(txt_title, motif.option_info.title_text_video)
				break
			--Custom
			elseif t[item].itemname == 'custom' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				local width = tonumber(main.f_input(main.f_extractText(motif.option_info.input_text_reswidth), motif.option_info, motif.optionbgdef, 'string'))
				if width ~= nil then
					sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
					local height = tonumber(main.f_input(main.f_extractText(motif.option_info.input_text_resheight), motif.option_info, motif.optionbgdef, 'string'))
					if height ~= nil then
						config.Width = width
						config.Height = height
						sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
						if (height / 3 * 4) ~= width then
							main.f_warning(main.f_extractText(motif.warning_info.text_res), motif.option_info, motif.optionbgdef)
						end
						modified = 1
						needReload = 1
					else
						sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
					end
				else
					sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				end
				break
			--Resolution
			else
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				config.Width = t[item].x
				config.Height = t[item].y
				if (config.Height / 3 * 4) ~= config.Width then
					main.f_warning(main.f_extractText(motif.warning_info.text_res), motif.option_info, motif.optionbgdef)
				end
				modified = 1
				needReload = 1
				break
			end
		end
		options.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; ENGINE SETTINGS
--;===========================================================
local t_engineCfg = {
	{data = textImgNew(), itemname = 'allowdebugkeys', displayname = motif.option_info.menu_itemname_engine_allowdebugkeys, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.AllowDebugKeys, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled)},
	{data = textImgNew(), itemname = 'simulmode', displayname = motif.option_info.menu_itemname_gameplay_simulmode, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.SimulMode, motif.option_info.menu_itemname_disabled, motif.option_info.menu_itemname_enabled)},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'helpermax', displayname = motif.option_info.menu_itemname_video_helpermax, vardata = textImgNew(), vardisplay = config.HelperMax},
	{data = textImgNew(), itemname = 'playerprojectilemax', displayname = motif.option_info.menu_itemname_video_playerprojectilemax, vardata = textImgNew(), vardisplay = config.PlayerProjectileMax},
	{data = textImgNew(), itemname = 'explodmax', displayname = motif.option_info.menu_itemname_video_explodmax, vardata = textImgNew(), vardisplay = config.ExplodMax},
	{data = textImgNew(), itemname = 'afterimagemax', displayname = motif.option_info.menu_itemname_video_afterimagemax, vardata = textImgNew(), vardisplay = config.AfterImageMax},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'zoomactive', displayname = motif.option_info.menu_itemname_video_zoomactive, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.ZoomActive)},
	{data = textImgNew(), itemname = 'maxzoomout', displayname = motif.option_info.menu_itemname_video_maxzoomout, vardata = textImgNew(), vardisplay = config.ZoomMin},
	{data = textImgNew(), itemname = 'maxzoomin', displayname = motif.option_info.menu_itemname_video_maxzoomin, vardata = textImgNew(), vardisplay = config.ZoomMax},
	{data = textImgNew(), itemname = 'zoomspeed', displayname = motif.option_info.menu_itemname_video_zoomspeed, vardata = textImgNew(), vardisplay = config.ZoomSpeed},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_video_back},
}
t_engineCfg = main.f_cleanTable(t_engineCfg)

function options.f_engineCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_engineCfg
	textImgSetText(txt_title, motif.option_info.title_text_engine)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		--HelperMax
		elseif t[item].itemname == 'helpermax' then
			if commandGetState(main.p1Cmd, 'r') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.HelperMax = config.HelperMax + 10
				t[item].vardisplay = config.HelperMax
				modified = 1
				needReload = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.HelperMax > 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.HelperMax = config.HelperMax - 10
				t[item].vardisplay = config.HelperMax
				modified = 1
				needReload = 1
			end
		--PlayerProjectileMax
		elseif t[item].itemname == 'playerprojectilemax' then
			if commandGetState(main.p1Cmd, 'r') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.PlayerProjectileMax = config.PlayerProjectileMax + 10
				t[item].vardisplay = config.PlayerProjectileMax
				modified = 1
				needReload = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.PlayerProjectileMax > 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.PlayerProjectileMax = config.PlayerProjectileMax - 10
				t[item].vardisplay = config.PlayerProjectileMax
				modified = 1
				needReload = 1
			end
		--ExplodMax
		elseif t[item].itemname == 'explodmax' then
			if commandGetState(main.p1Cmd, 'r') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.ExplodMax = config.ExplodMax + 10
				t[item].vardisplay = config.ExplodMax
				modified = 1
				needReload = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.ExplodMax > 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.ExplodMax = config.ExplodMax - 10
				t[item].vardisplay = config.ExplodMax
				modified = 1
				needReload = 1
			end
		--AfterImageMax
		elseif t[item].itemname == 'afterimagemax' then
			if commandGetState(main.p1Cmd, 'r') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.AfterImageMax = config.AfterImageMax + 1
				t[item].vardisplay = config.AfterImageMax
				modified = 1
				needReload = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.AfterImageMax > 1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.AfterImageMax = config.AfterImageMax - 1
				t[item].vardisplay = config.AfterImageMax
				modified = 1
				needReload = 1
			end
		--Zoom Active
		elseif t[item].itemname == 'zoomactive' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.ZoomActive then
				config.ZoomActive = false
			else
				config.ZoomActive = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.ZoomActive)
			modified = 1
		--Default Max Zoom Out
		elseif t[item].itemname == 'maxzoomout' then
			if commandGetState(main.p1Cmd, 'r') and config.ZoomMin < 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.ZoomMin = options.f_precision(config.ZoomMin + 0.05, '%.02f')
				t[item].vardisplay = config.ZoomMin
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.ZoomMin > 0.05 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.ZoomMin = options.f_precision(config.ZoomMin - 0.05, '%.02f')
				t[item].vardisplay = config.ZoomMin
				modified = 1
			end
		--Default Max Zoom In
		elseif t[item].itemname == 'maxzoomin' then
			if commandGetState(main.p1Cmd, 'r') and config.ZoomMax < 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.ZoomMax = options.f_precision(config.ZoomMax + 0.05, '%.02f')
				t[item].vardisplay = config.ZoomMax
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.ZoomMax > 0.05 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.ZoomMax = options.f_precision(config.ZoomMax - 0.05, '%.02f')
				t[item].vardisplay = config.ZoomMax
				modified = 1
			end
		--Default Zoom Speed
		elseif t[item].itemname == 'zoomspeed' then
			if commandGetState(main.p1Cmd, 'r') and config.ZoomSpeed < 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.ZoomSpeed = options.f_precision(config.ZoomSpeed + 0.1, '%.01f')
				t[item].vardisplay = config.ZoomSpeed
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.ZoomSpeed > 0.1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.ZoomSpeed = options.f_precision(config.ZoomSpeed - 0.1, '%.01f')
				t[item].vardisplay = config.ZoomSpeed
				modified = 1
			end
		-- Leagcy TAG Mode
		elseif t[item].itemname == 'simulmode' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.SimulMode then
				config.SimulMode = false
			else
				config.SimulMode = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.SimulMode, motif.option_info.menu_itemname_disabled, motif.option_info.menu_itemname_enabled)
			modified = 1
			needReload = 1
		-- Allow Debug Keys
		elseif t[item].itemname == 'allowdebugkeys' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AllowDebugKeys then
				config.AllowDebugKeys = false
			else
				config.AllowDebugKeys = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.AllowDebugKeys, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled)
			modified = 1
		--Back
		elseif t[item].itemname == 'back' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		end
		options.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; INPUT SETTINGS
--;===========================================================
local t_inputCfg = {
	{data = textImgNew(), itemname = 'p1keyboard', displayname = motif.option_info.menu_itemname_input_p1keyboard},
	{data = textImgNew(), itemname = 'p1gamepad', displayname = motif.option_info.menu_itemname_input_p1gamepad},
	{data = textImgNew(), itemname = 'p2keyboard', displayname = motif.option_info.menu_itemname_input_p2keyboard},
	{data = textImgNew(), itemname = 'p2gamepad', displayname = motif.option_info.menu_itemname_input_p2gamepad},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'defaultvalues', displayname = motif.option_info.menu_itemname_input_default},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_input_back},
}
t_inputCfg = main.f_cleanTable(t_inputCfg)

function options.f_inputCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_inputCfg
	textImgSetText(txt_title, motif.option_info.title_text_input)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		elseif main.f_btnPalNo(main.p1Cmd) > 0 then
			--P1 Keyboard
			if t[item].itemname == 'p1keyboard' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.f_keyCfg(1, -1)
			--P1 Gamepad
			elseif t[item].itemname == 'p1gamepad' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.f_keyCfg(1, 0)
			--P2 Keyboard
			elseif t[item].itemname == 'p2keyboard' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.f_keyCfg(2, -1)
			--P2 Gamepad
			elseif t[item].itemname == 'p2gamepad' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.f_keyCfg(2, 1)
			--Default Values
			elseif t[item].itemname == 'defaultvalues' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				options.f_keyDefault()
				modified = 1
				needReload = 1
			--Back
			elseif t[item].itemname == 'back' then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				textImgSetText(txt_title, motif.option_info.title_text_main)
				break
			end
		end
		options.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; KEY SETTINGS
--;===========================================================
local t_keyCfg = {
	{data = textImgNew(), itemname = 'up', displayname = motif.option_info.menu_itemname_key_up, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'down', displayname = motif.option_info.menu_itemname_key_down, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'left', displayname = motif.option_info.menu_itemname_key_left, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'right', displayname = motif.option_info.menu_itemname_key_right, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'a', displayname = motif.option_info.menu_itemname_key_a, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'b', displayname = motif.option_info.menu_itemname_key_b, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'c', displayname = motif.option_info.menu_itemname_key_c, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'x', displayname = motif.option_info.menu_itemname_key_x, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'y', displayname = motif.option_info.menu_itemname_key_y, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'z', displayname = motif.option_info.menu_itemname_key_z, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'start', displayname = motif.option_info.menu_itemname_key_start, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'v', displayname = motif.option_info.menu_itemname_key_v, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'w', displayname = motif.option_info.menu_itemname_key_w, vardata = textImgNew(), vardisplay = ''},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_key_back},
}
t_keyCfg = main.f_cleanTable(t_keyCfg)

function options.f_keyCfg(playerNo, controller)
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_keyCfg
	textImgSetText(txt_title, motif.option_info.title_text_key)
	for i = 1, #t_keyCfg do
		if controller == -1 then
			t_keyCfg[i].vardisplay = config.KeyConfig[playerNo].Buttons[i]
		else
			t_keyCfg[i].vardisplay = config.JoystickConfig[playerNo].Buttons[i]
		end
	end
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_input)
			break
		elseif main.f_btnPalNo(main.p1Cmd) > 0 then
			--Back
			if t[item].itemname == 'back' then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				textImgSetText(txt_title, motif.option_info.title_text_input)
				break
			--Buttons
			else
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				local key = main.f_input(main.f_extractText(motif.option_info.input_text_key), motif.option_info, motif.optionbgdef, 'key')
				if (controller == -1 and key ~= '') or (controller ~= -1 and tonumber(key) ~= nil) then
					sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
					t[item].vardisplay = key
					if controller == -1 then
						config.KeyConfig[playerNo].Buttons[item] = key
					else
						config.JoystickConfig[playerNo].Buttons[item] = key
					end
				else
					sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				end
			end
			modified = 1
			needReload = 1
		end
		options.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; AUDIO SETTINGS
--;===========================================================
local t_audioCfg = {
	{data = textImgNew(), itemname = 'mastervolume', displayname = motif.option_info.menu_itemname_audio_mastervolume, vardata = textImgNew(), vardisplay = config.MasterVolume .. '%'},
	{data = textImgNew(), itemname = 'bgmvolume', displayname = motif.option_info.menu_itemname_audio_bgmvolume, vardata = textImgNew(), vardisplay = config.BgmVolume .. '%'},
	{data = textImgNew(), itemname = 'sfxvolume', displayname = motif.option_info.menu_itemname_audio_sfxvolume, vardata = textImgNew(), vardisplay = config.WavVolume .. '%'},
	{data = textImgNew(), itemname = 'audioducking', displayname = motif.option_info.menu_itemname_audio_audioducking, vardata = textImgNew(), vardisplay = options.f_intDisplay(config.AudioDucking, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled)},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_video_back},
}
t_audioCfg = main.f_cleanTable(t_audioCfg)

function options.f_audioCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_audioCfg
	textImgSetText(txt_title, motif.option_info.title_text_audio)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommon1(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		-- Master Volume
		elseif t[item].itemname == 'mastervolume' then
			if commandGetState(main.p1Cmd, 'r') and config.MasterVolume < 200 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.MasterVolume = config.MasterVolume + 1
				t[item].vardisplay = config.MasterVolume .. '%'
				modified = 1
				needReload = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.MasterVolume > 0 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.MasterVolume = config.MasterVolume - 1
				t[item].vardisplay = config.MasterVolume .. '%'
				modified = 1
				needReload = 1
			end
		-- BGM Volume
		elseif t[item].itemname == 'bgmvolume' then
			if commandGetState(main.p1Cmd, 'r') and config.BgmVolume < 100 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.BgmVolume = config.BgmVolume + 1
				t[item].vardisplay = config.BgmVolume .. '%'
				modified = 1
				needReload = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.BgmVolume > 0 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.BgmVolume = config.BgmVolume - 1
				t[item].vardisplay = config.BgmVolume .. '%'
				modified = 1
				needReload = 1
			end
		-- SFX Volume
		elseif t[item].itemname == 'sfxvolume' then
			if commandGetState(main.p1Cmd, 'r') and config.WavVolume < 100 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.WavVolume = config.WavVolume + 1
				t[item].vardisplay = config.WavVolume .. '%'
				modified = 1
				needReload = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.WavVolume > 0 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.WavVolume = config.WavVolume - 1
				t[item].vardisplay = config.WavVolume .. '%'
				modified = 1
				needReload = 1
			end
		-- Allow Debug Keys
		elseif t[item].itemname == 'audioducking' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AudioDucking then
				config.AudioDucking = false
			else
				config.AudioDucking = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.AudioDucking, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled)
			modified = 1
			needReload = 1
		--Back
		elseif t[item].itemname == 'back' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		end
		options.f_menuCommon2(cursorPosY, moveTxt, item, t)
	end
end

return options


local options = {}

--;===========================================================
--; COMMON
--;===========================================================
local modified = 0
local needReload = 0

if config.RoundsNumSingle == -1 then
	options.roundsNumSingle = getMatchWins()
else
	options.roundsNumSingle = config.RoundsNumSingle
end
if config.RoundsNumTeam == -1 then
	options.roundsNumTeam = getMatchWins()
else
	options.roundsNumTeam = config.RoundsNumTeam
end
if config.MaxDrawGames == -2 then
	options.maxDrawGames = getMatchMaxDrawGames()
else
	options.maxDrawGames = config.MaxDrawGames
end

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

--return correct menu_itemname_controller string
function options.f_itemnameController(player)
	local kb = 0
	local pad = 0
	local itemname = motif.option_info.menu_itemname_controller_disabled
	for i = 1, #config.KeyConfig do
		if config.KeyConfig[i].Joystick == -1 then
			kb = kb + 1
		else
			pad = pad + 1
		end
		if config.KeyConfig[i].Player == player then
			if config.KeyConfig[i].Joystick == -1 then
				itemname = main.f_extractText(motif.option_info.menu_itemname_controller_keyboard, kb)[1]
			else
				itemname = main.f_extractText(motif.option_info.menu_itemname_controller_gamepad, pad)[1]
			end
			break
		end
	end
	return itemname
end

--return correct precision
function options.f_precision(v, decimal)
	return tonumber(string.format(decimal, v))
end

--save configuration
function options.f_saveCfg()
	--Data saving to config.json
	local file = io.open("save/config.json","w+")
	file:write(json.encode(config, {indent = true}))
	file:close()
	--Reload game if needed
	if needReload == 1 then
		main.f_warning(main.f_extractText(motif.warning_info.text_reload), motif.option_info, motif.optionbgdef)
		os.exit()
	end
end

--reset key settings
function options.f_keyDefault()
	local cnt = 0
	for i = 1, #config.KeyConfig do
		config.KeyConfig[i].Player = 0
		if config.KeyConfig[i].Joystick == -1 then
			cnt = cnt + 1
			if cnt == 1 then
				config.KeyConfig[i].Player = 1
				config.KeyConfig[i].Buttons[1] = 'UP'
				config.KeyConfig[i].Buttons[2] = 'DOWN'
				config.KeyConfig[i].Buttons[3] = 'LEFT'
				config.KeyConfig[i].Buttons[4] = 'RIGHT'
				config.KeyConfig[i].Buttons[5] = 'z'
				config.KeyConfig[i].Buttons[6] = 'x'
				config.KeyConfig[i].Buttons[7] = 'c'
				config.KeyConfig[i].Buttons[8] = 'a'
				config.KeyConfig[i].Buttons[9] = 's'
				config.KeyConfig[i].Buttons[10] = 'd'
				config.KeyConfig[i].Buttons[11] = 'RETURN'
				config.KeyConfig[i].Buttons[12] = 'q'
				config.KeyConfig[i].Buttons[13] = 'w'
			elseif cnt == 2 then
				config.KeyConfig[i].Player = 2
				config.KeyConfig[i].Buttons[1] = 't'
				config.KeyConfig[i].Buttons[2] = 'g'
				config.KeyConfig[i].Buttons[3] = 'f'
				config.KeyConfig[i].Buttons[4] = 'h'
				config.KeyConfig[i].Buttons[5] = 'j'
				config.KeyConfig[i].Buttons[6] = 'k'
				config.KeyConfig[i].Buttons[7] = 'l'
				config.KeyConfig[i].Buttons[8] = 'u'
				config.KeyConfig[i].Buttons[9] = 'i'
				config.KeyConfig[i].Buttons[10] = 'o'
				config.KeyConfig[i].Buttons[11] = 'RSHIFT'
				config.KeyConfig[i].Buttons[12] = 'LEFTBRACKET'
				config.KeyConfig[i].Buttons[13] = 'RIGHTBRACKET'
			else
				for j = 1, #config.KeyConfig[i].Buttons do
					config.KeyConfig[i].Buttons[j] = motif.option_info.menu_itemname_info_disable
				end
			end
		else
			config.KeyConfig[i].Buttons[1] = 10
			config.KeyConfig[i].Buttons[2] = 12
			config.KeyConfig[i].Buttons[3] = 13
			config.KeyConfig[i].Buttons[4] = 11
			config.KeyConfig[i].Buttons[5] = 0
			config.KeyConfig[i].Buttons[6] = 1
			config.KeyConfig[i].Buttons[7] = 4
			config.KeyConfig[i].Buttons[8] = 2
			config.KeyConfig[i].Buttons[9] = 3
			config.KeyConfig[i].Buttons[10] = 5
			config.KeyConfig[i].Buttons[11] = 7
			config.KeyConfig[i].Buttons[12] = 8
			config.KeyConfig[i].Buttons[13] = 9
		end
	end
	resetRemapInput()
end

--reset vardisplay in tables
function options.f_resetTables()
	local t_displaynameReset = {
		t_mainCfg = {
			portchange = getListenPort(),
		},
		t_arcadeCfg = {
			roundtime = config.RoundTime,
			roundsnumsingle = options.roundsNumSingle,
			roundsnumteam = options.roundsNumTeam,
			maxdrawgames = options.maxDrawGames,
			difficulty = config.Difficulty,
			credits = config.Credits,
			charchange = options.f_boolDisplay(config.ContSelection),
			airamping = options.f_boolDisplay(config.AIRamping),
		},
		t_gameplayCfg = {
			lifemul = config.LifeMul .. '%',
			autoguard = options.f_boolDisplay(config.AutoGuard),
			attackpowermul = config['Attack.LifeToPowerMul'],
			gethitpowermul = config['GetHit.LifeToPowerMul'],
			superdefencemul = config['Super.TargetDefenceMul'],
			team1vs2life = config.Team1VS2Life,
			turnsrecoveryrate = config.TurnsRecoveryRate,
			teampowershare = options.f_boolDisplay(config.TeamPowerShare),
			teamlifeshare = options.f_boolDisplay(config.TeamLifeShare),
			numturns = config.NumTurns,
			numsimul = config.NumSimul,
			numtag = config.NumTag,
			simulmode = options.f_boolDisplay(config.SimulMode, motif.option_info.menu_itemname_gameplay_simulmode_simul, motif.option_info.menu_itemname_gameplay_simulmode_tag),
		},
		t_videoCfg = {
			resolution = config.Width .. 'x' .. config.Height,
			fullscreen = options.f_boolDisplay(config.Fullscreen),
			helpermax = config.HelperMax,
			playerprojectilemax = config.PlayerProjectileMax,
			explodmax = config.ExplodMax,
			afterimagemax = config.AfterImageMax,
			zoomactive = options.f_boolDisplay(config.ZoomActive),
			maxzoomout = config.ZoomMin,
			maxzoomin = config.ZoomMax,
			zoomspeed = config.ZoomSpeed,
			lifebarfontscale = config.LifebarFontScale,
			airandomcolor = options.f_boolDisplay(config.AIRandomColor, motif.option_info.menu_itemname_video_aipalette_random, motif.option_info.menu_itemname_video_aipalette_default),
		},
		t_audioCfg = {
			wavvolume = config.WavVolume,
		},
		t_inputCfg = {
			p1controller = options.f_itemnameController(1),
			p2controller = options.f_itemnameController(2),
			p3controller = options.f_itemnameController(3),
			p4controller = options.f_itemnameController(4),
		},
	}
	for k1, v1 in pairs(t_displaynameReset) do
		for k2, v2 in pairs(t_displaynameReset[k1]) do
			for i = 1, #options[k1] do
				if options[k1][i].itemname == k2 then
					options[k1][i].vardisplay = v2
				end
			end
		end
	end
end

function options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
	if commandGetState(main.p1Cmd, 'u') then
		sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
		item = item - 1
	elseif commandGetState(main.p1Cmd, 'd') then
		sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
		item = item + 1
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
	elseif commandGetState(main.p1Cmd, 'd') and cursorPosY < motif.option_info.menu_window_visibleitems then
		cursorPosY = cursorPosY + 1
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
	motif.option_info.title_font[6],
	motif.option_info.title_font[7],
	motif.option_info.title_font[8]
)
function options.f_menuCommonDraw(cursorPosY, moveTxt, item, t, fadeType)
	fadeType = fadeType or 'fadein'
	--draw clearcolor
	clearColor(motif.optionbgdef.bgclearcolor[1], motif.optionbgdef.bgclearcolor[2], motif.optionbgdef.bgclearcolor[3])
	--draw layerno = 0 backgrounds
	bgDraw(motif.optionbgdef.bg, false)
	--draw menu box
	if motif.option_info.menu_boxbg_visible == 1 then
		local coord4 = 0
		if #t > motif.option_info.menu_window_visibleitems and moveTxt == (#t - motif.option_info.menu_window_visibleitems) * motif.option_info.menu_item_spacing[2] then
			coord4 = motif.option_info.menu_window_visibleitems * (motif.option_info.menu_boxcursor_coords[4] - motif.option_info.menu_boxcursor_coords[2] + 1) + main.f_oddRounding(motif.option_info.menu_boxcursor_coords[2])
		else
			coord4 = #t * (motif.option_info.menu_boxcursor_coords[4] - motif.option_info.menu_boxcursor_coords[2] + 1) + main.f_oddRounding(motif.option_info.menu_boxcursor_coords[2])
		end
		fillRect(
			motif.option_info.menu_pos[1] + motif.option_info.menu_boxcursor_coords[1],
			motif.option_info.menu_pos[2] + motif.option_info.menu_boxcursor_coords[2],
			motif.option_info.menu_boxcursor_coords[3] - motif.option_info.menu_boxcursor_coords[1] + 1,
			coord4,
			motif.option_info.menu_boxbg_col[1],
			motif.option_info.menu_boxbg_col[2],
			motif.option_info.menu_boxbg_col[3],
			motif.option_info.menu_boxbg_alpha[1],
			motif.option_info.menu_boxbg_alpha[2]
		)
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
					motif.option_info.menu_item_active_font[6],
					motif.option_info.menu_item_active_font[7],
					motif.option_info.menu_item_active_font[8]
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
						motif.option_info.menu_item_value_active_font[6],
						motif.option_info.menu_item_value_active_font[7],
						motif.option_info.menu_item_value_active_font[8]
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
					motif.option_info.menu_item_font[6],
					motif.option_info.menu_item_font[7],
					motif.option_info.menu_item_font[8]
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
						motif.option_info.menu_item_value_font[6],
						motif.option_info.menu_item_value_font[7],
						motif.option_info.menu_item_value_font[8]
					))
				end
			end
		end
	end
	--draw menu cursor
	if motif.option_info.menu_boxcursor_visible == 1 and not main.fadeActive then
		local src, dst = main.f_boxcursorAlpha(
			motif.option_info.menu_boxcursor_alpharange[1],
			motif.option_info.menu_boxcursor_alpharange[2],
			motif.option_info.menu_boxcursor_alpharange[3],
			motif.option_info.menu_boxcursor_alpharange[4],
			motif.option_info.menu_boxcursor_alpharange[5],
			motif.option_info.menu_boxcursor_alpharange[6]
		)
		fillRect(
			motif.option_info.menu_pos[1] + motif.option_info.menu_boxcursor_coords[1],
			motif.option_info.menu_pos[2] + motif.option_info.menu_boxcursor_coords[2] + (cursorPosY - 1) * motif.option_info.menu_item_spacing[2],
			motif.option_info.menu_boxcursor_coords[3] - motif.option_info.menu_boxcursor_coords[1] + 1,
			motif.option_info.menu_boxcursor_coords[4] - motif.option_info.menu_boxcursor_coords[2] + 1 + main.f_oddRounding(motif.option_info.menu_boxcursor_coords[2]),
			motif.option_info.menu_boxcursor_col[1],
			motif.option_info.menu_boxcursor_col[2],
			motif.option_info.menu_boxcursor_col[3],
			src,
			dst
		)
	end
	--draw layerno = 1 backgrounds
	bgDraw(motif.optionbgdef.bg, true)
	--draw fadein / fadeout
	main.fadeActive = fadeScreen(
		fadeType,
		main.fadeStart,
		motif.option_info[fadeType .. '_time'],
		motif.option_info[fadeType .. '_col'][1],
		motif.option_info[fadeType .. '_col'][2],
		motif.option_info[fadeType .. '_col'][3]
	)
	--frame transition
	if main.fadeActive then
		commandBufReset(main.p1Cmd)
	elseif fadeType == 'fadeout' then
		commandBufReset(main.p1Cmd)
		return --skip last frame rendering
	else
		main.f_cmdInput()
	end
	refresh()
end

--;===========================================================
--; MAIN LOOP
--;===========================================================
options.t_mainCfg = {
	{data = textImgNew(), itemname = 'arcadesettings', displayname = motif.option_info.menu_itemname_main_arcade},
	{data = textImgNew(), itemname = 'gameplaysettings', displayname = motif.option_info.menu_itemname_main_gameplay},
	{data = textImgNew(), itemname = 'videosettings', displayname = motif.option_info.menu_itemname_main_video},
	{data = textImgNew(), itemname = 'audiosettings', displayname = motif.option_info.menu_itemname_main_audio},
	{data = textImgNew(), itemname = 'inputsettings', displayname = motif.option_info.menu_itemname_main_input},
	{data = textImgNew(), itemname = 'portchange', displayname = motif.option_info.menu_itemname_main_port, vardata = textImgNew(), vardisplay = getListenPort()},
	{data = textImgNew(), itemname = 'defaultvalues', displayname = motif.option_info.menu_itemname_main_default},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_main_back},
}
options.t_mainCfg = main.f_cleanTable(options.t_mainCfg, main.t_sort.option_info)

function options.f_mainCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = options.t_mainCfg
	textImgSetText(txt_title, motif.option_info.title_text_main)
	if motif.music.options_bgm == '' then
		main.f_menuReset(motif.optionbgdef.bg)
	else
		main.f_menuReset(motif.optionbgdef.bg, motif.music.option_bgm, motif.music.option_bgm_loop, motif.music.option_bgm_volume, motif.music.option_bgm_loopstart, motif.music.option_bgm_loopend)
	end
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			if modified == 1 then
				options.f_saveCfg()
			end
			main.f_menuFadeOut('option_info', cursorPosY, moveTxt, item, t)
			if motif.music.option_bgm == '' then
				main.f_menuReset(motif.titlebgdef.bg)
			else
				main.f_menuReset(motif.titlebgdef.bg, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
			end
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
			--Audio Settings
			elseif t[item].itemname == 'audiosettings' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				options.f_audioCfg()
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
				config.ZoomActive = true
				config.ZoomMin = 0.75
				config.ZoomMax = 1.0
				config.ZoomSpeed = 1.0
				config.AIRandomColor = true
				config.WavVolume = 80
				config.RoundTime = 99
				config.RoundsNumSingle = -1
				config.RoundsNumTeam = -1
				config.MaxDrawGames = -2
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
				loadLifebar(motif.files.fight)
				options.roundsNumSingle = getMatchWins()
				options.roundsNumTeam = getMatchWins()
				options.maxDrawGames = getMatchMaxDrawGames()
				options.f_resetTables()
				modified = 1
				needReload = 1
			--Back
			elseif t[item].itemname == 'back' then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				if modified == 1 then
					options.f_saveCfg()
				end
				main.f_menuFadeOut('option_info', cursorPosY, moveTxt, item, t)
				if motif.music.option_bgm == '' then
					main.f_menuReset(motif.titlebgdef.bg)
				else
					main.f_menuReset(motif.titlebgdef.bg, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
				end
				break
			end
		end
		options.f_menuCommonDraw(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; ARCADE SETTINGS
--;===========================================================
options.t_arcadeCfg = {
	{data = textImgNew(), itemname = 'roundtime', displayname = motif.option_info.menu_itemname_arcade_roundtime, vardata = textImgNew(), vardisplay = config.RoundTime},
	{data = textImgNew(), itemname = 'roundsnumsingle', displayname = motif.option_info.menu_itemname_arcade_roundsnumsingle, vardata = textImgNew(), vardisplay = options.roundsNumSingle},
	{data = textImgNew(), itemname = 'roundsnumteam', displayname = motif.option_info.menu_itemname_arcade_roundsnumteam, vardata = textImgNew(), vardisplay = options.roundsNumTeam},
	{data = textImgNew(), itemname = 'maxdrawgames', displayname = motif.option_info.menu_itemname_arcade_maxdrawgames, vardata = textImgNew(), vardisplay = options.maxDrawGames},
	{data = textImgNew(), itemname = 'difficulty', displayname = motif.option_info.menu_itemname_arcade_difficulty, vardata = textImgNew(), vardisplay = config.Difficulty},
	{data = textImgNew(), itemname = 'credits', displayname = motif.option_info.menu_itemname_arcade_credits, vardata = textImgNew(), vardisplay = config.Credits},
	{data = textImgNew(), itemname = 'charchange', displayname = motif.option_info.menu_itemname_arcade_charchange, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.ContSelection)},
	{data = textImgNew(), itemname = 'airamping', displayname = motif.option_info.menu_itemname_arcade_airamping, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.AIRamping)},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_arcade_back},
}
options.t_arcadeCfg = main.f_cleanTable(options.t_arcadeCfg, main.t_sort.option_info)

function options.f_arcadeCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = options.t_arcadeCfg
	textImgSetText(txt_title, motif.option_info.title_text_arcade)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
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
		--Rounds to Win Single
		elseif t[item].itemname == 'roundsnumsingle' then
			if commandGetState(main.p1Cmd, 'r') and options.roundsNumSingle < 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.roundsNumSingle = options.roundsNumSingle + 1
				t[item].vardisplay = options.roundsNumSingle
				config.RoundsNumSingle = options.roundsNumSingle
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and options.roundsNumSingle > 1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.roundsNumSingle = options.roundsNumSingle - 1
				t[item].vardisplay = options.roundsNumSingle
				config.RoundsNumSingle = options.roundsNumSingle
				modified = 1
			end
		--Rounds to Win Simul/Tag
		elseif t[item].itemname == 'roundsnumteam' then
			if commandGetState(main.p1Cmd, 'r') and options.roundsNumTeam < 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.roundsNumTeam = options.roundsNumTeam + 1
				t[item].vardisplay = options.roundsNumTeam
				config.RoundsNumTeam = options.roundsNumTeam
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and options.roundsNumTeam > 1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.roundsNumTeam = options.roundsNumTeam - 1
				t[item].vardisplay = options.roundsNumTeam
				config.RoundsNumTeam = options.roundsNumTeam
				modified = 1
			end
		--Max Draw Games
		elseif t[item].itemname == 'maxdrawgames' then
			if commandGetState(main.p1Cmd, 'r') and options.maxDrawGames < 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.maxDrawGames = options.maxDrawGames + 1
				t[item].vardisplay = options.maxDrawGames
				config.MaxDrawGames = options.maxDrawGames
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and options.maxDrawGames > -1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.maxDrawGames = options.maxDrawGames - 1
				t[item].vardisplay = options.maxDrawGames
				config.MaxDrawGames = options.maxDrawGames
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
		options.f_menuCommonDraw(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; GAMEPLAY SETTINGS
--;===========================================================
options.t_gameplayCfg = {
	{data = textImgNew(), itemname = 'lifemul', displayname = motif.option_info.menu_itemname_gameplay_lifemul, vardata = textImgNew(), vardisplay = config.LifeMul .. '%'},
	{data = textImgNew(), itemname = 'autoguard', displayname = motif.option_info.menu_itemname_gameplay_autoguard, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.AutoGuard)},
	{data = textImgNew(), itemname = 'attackpowermul', displayname = motif.option_info.menu_itemname_gameplay_attackpowermul, vardata = textImgNew(), vardisplay = config['Attack.LifeToPowerMul']},
	{data = textImgNew(), itemname = 'gethitpowermul', displayname = motif.option_info.menu_itemname_gameplay_gethitpowermul, vardata = textImgNew(), vardisplay = config['GetHit.LifeToPowerMul']},
	{data = textImgNew(), itemname = 'superdefencemul', displayname = motif.option_info.menu_itemname_gameplay_superdefencemul, vardata = textImgNew(), vardisplay = config['Super.TargetDefenceMul']},
	{data = textImgNew(), itemname = 'team1vs2life', displayname = motif.option_info.menu_itemname_gameplay_team1vs2life, vardata = textImgNew(), vardisplay = config.Team1VS2Life},
	{data = textImgNew(), itemname = 'turnsrecoveryrate', displayname = motif.option_info.menu_itemname_gameplay_turnsrecoveryrate, vardata = textImgNew(), vardisplay = config.TurnsRecoveryRate},
	{data = textImgNew(), itemname = 'teampowershare', displayname = motif.option_info.menu_itemname_gameplay_teampowershare, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.TeamPowerShare)},
	{data = textImgNew(), itemname = 'teamlifeshare', displayname = motif.option_info.menu_itemname_gameplay_teamlifeshare, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.TeamLifeShare)},
	{data = textImgNew(), itemname = 'numturns', displayname = motif.option_info.menu_itemname_gameplay_numturns, vardata = textImgNew(), vardisplay = config.NumTurns},
	{data = textImgNew(), itemname = 'numsimul', displayname = motif.option_info.menu_itemname_gameplay_numsimul, vardata = textImgNew(), vardisplay = config.NumSimul},
	{data = textImgNew(), itemname = 'numtag', displayname = motif.option_info.menu_itemname_gameplay_numtag, vardata = textImgNew(), vardisplay = config.NumTag},
	{data = textImgNew(), itemname = 'simulmode', displayname = motif.option_info.menu_itemname_gameplay_simulmode, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.SimulMode, motif.option_info.menu_itemname_gameplay_simulmode_simul, motif.option_info.menu_itemname_gameplay_simulmode_tag)},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_gameplay_back},
}
options.t_gameplayCfg = main.f_cleanTable(options.t_gameplayCfg, main.t_sort.option_info)

function options.f_gameplayCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = options.t_gameplayCfg
	textImgSetText(txt_title, motif.option_info.title_text_gameplay)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
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
			if commandGetState(main.p1Cmd, 'r') and config.NumTurns < 4 then
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
		--Assist Mode
		elseif t[item].itemname == 'simulmode' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.SimulMode then
				config.SimulMode = false
			else
				config.SimulMode = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.SimulMode, motif.option_info.menu_itemname_gameplay_simulmode_simul, motif.option_info.menu_itemname_gameplay_simulmode_tag)
			main.f_warning(main.f_extractText(motif.warning_info.text_simul), motif.option_info, motif.optionbgdef)
			modified = 1
		--Back
		elseif t[item].itemname == 'back' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		end
		options.f_menuCommonDraw(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; VIDEO SETTINGS
--;===========================================================
options.t_videoCfg = {
	{data = textImgNew(), itemname = 'resolution', displayname = motif.option_info.menu_itemname_video_resolution, vardata = textImgNew(), vardisplay = config.Width .. 'x' .. config.Height},
	{data = textImgNew(), itemname = 'fullscreen', displayname = motif.option_info.menu_itemname_video_fullscreen, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.Fullscreen)},
	{data = textImgNew(), itemname = 'helpermax', displayname = motif.option_info.menu_itemname_video_helpermax, vardata = textImgNew(), vardisplay = config.HelperMax},
	{data = textImgNew(), itemname = 'playerprojectilemax', displayname = motif.option_info.menu_itemname_video_playerprojectilemax, vardata = textImgNew(), vardisplay = config.PlayerProjectileMax},
	{data = textImgNew(), itemname = 'explodmax', displayname = motif.option_info.menu_itemname_video_explodmax, vardata = textImgNew(), vardisplay = config.ExplodMax},
	{data = textImgNew(), itemname = 'afterimagemax', displayname = motif.option_info.menu_itemname_video_afterimagemax, vardata = textImgNew(), vardisplay = config.AfterImageMax},
	{data = textImgNew(), itemname = 'zoomactive', displayname = motif.option_info.menu_itemname_video_zoomactive, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.ZoomActive)},
	{data = textImgNew(), itemname = 'maxzoomout', displayname = motif.option_info.menu_itemname_video_maxzoomout, vardata = textImgNew(), vardisplay = config.ZoomMin},
	{data = textImgNew(), itemname = 'maxzoomin', displayname = motif.option_info.menu_itemname_video_maxzoomin, vardata = textImgNew(), vardisplay = config.ZoomMax},
	{data = textImgNew(), itemname = 'zoomspeed', displayname = motif.option_info.menu_itemname_video_zoomspeed, vardata = textImgNew(), vardisplay = config.ZoomSpeed},
	{data = textImgNew(), itemname = 'lifebarfontscale', displayname = motif.option_info.menu_itemname_video_lifebarfontscale, vardata = textImgNew(), vardisplay = config.LifebarFontScale},
	{data = textImgNew(), itemname = 'airandomcolor', displayname = motif.option_info.menu_itemname_video_aipalette, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.AIRandomColor, motif.option_info.menu_itemname_video_aipalette_random, motif.option_info.menu_itemname_video_aipalette_default)},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_video_back},
}
options.t_videoCfg = main.f_cleanTable(options.t_videoCfg, main.t_sort.option_info)

function options.f_videoCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = options.t_videoCfg
	textImgSetText(txt_title, motif.option_info.title_text_video)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
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
		--Default Lifebar Font Scale
		elseif t[item].itemname == 'lifebarfontscale' then
			if commandGetState(main.p1Cmd, 'r') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.LifebarFontScale = options.f_precision(config.LifebarFontScale + 0.1, '%.01f')
				t[item].vardisplay = config.LifebarFontScale
				modified = 1
				needReload = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.LifebarFontScale > 0.1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.LifebarFontScale = options.f_precision(config.LifebarFontScale - 0.1, '%.01f')
				t[item].vardisplay = config.LifebarFontScale
				modified = 1
				needReload = 1
			end
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
		--Back
		elseif t[item].itemname == 'back' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		end
		options.f_menuCommonDraw(cursorPosY, moveTxt, item, t)
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
	{data = textImgNew(), x = 1280, y = 720, displayname = motif.option_info.menu_itemname_res_1280x720},
	{data = textImgNew(), x = 1600, y = 900, displayname = motif.option_info.menu_itemname_res_1600x900},
	{data = textImgNew(), x = 1920, y = 1080, displayname = motif.option_info.menu_itemname_res_1920x1080},
	{data = textImgNew(), x = 2560, y = 1440, displayname = motif.option_info.menu_itemname_res_2560x1440},
	{data = textImgNew(), x = 3840, y = 2160, displayname = motif.option_info.menu_itemname_res_3840x2160},
	{data = textImgNew(), itemname = 'custom', displayname = motif.option_info.menu_itemname_res_custom},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_res_back},
}
t_resCfg = main.f_cleanTable(t_resCfg, main.t_sort.option_info)

function options.f_resCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_resCfg
	textImgSetText(txt_title, motif.option_info.title_text_res)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
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
				textImgSetText(txt_title, motif.option_info.title_text_video)
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
				textImgSetText(txt_title, motif.option_info.title_text_video)
				break
			end
		end
		options.f_menuCommonDraw(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; AUDIO SETTINGS
--;===========================================================
options.t_audioCfg = {
	{data = textImgNew(), itemname = 'wavvolume', displayname = motif.option_info.menu_itemname_audio_wavvolume, vardata = textImgNew(), vardisplay = config.WavVolume},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_audio_back},
}
options.t_audioCfg = main.f_cleanTable(options.t_audioCfg, main.t_sort.option_info)

function options.f_audioCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = options.t_audioCfg
	textImgSetText(txt_title, motif.option_info.title_text_audio)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		--1P Vs Team Life
		elseif t[item].itemname == 'wavvolume' then
			if commandGetState(main.p1Cmd, 'r') and config.WavVolume < 100 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.WavVolume = config.WavVolume + 1
				t[item].vardisplay = config.WavVolume
				setWavVolume(config.WavVolume)
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.WavVolume > 0 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.WavVolume = config.WavVolume - 1
				t[item].vardisplay = config.WavVolume
				setWavVolume(config.WavVolume)
				modified = 1
			end
		--Back
		elseif t[item].itemname == 'back' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		end
		options.f_menuCommonDraw(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; INPUT SETTINGS
--;===========================================================
options.t_inputCfg = {
	{data = textImgNew(), itemname = 'keyboard', displayname = motif.option_info.menu_itemname_input_keyboard},
	{data = textImgNew(), itemname = 'gamepad', displayname = motif.option_info.menu_itemname_input_gamepad},
	--{data = textImgNew(), itemname = 'system', displayname = motif.option_info.menu_itemname_input_system},
	{data = textImgNew(), itemname = 'p1controller', displayname = motif.option_info.menu_itemname_input_p1controller, vardata = textImgNew(), vardisplay = options.f_itemnameController(1)},
	{data = textImgNew(), itemname = 'p2controller', displayname = motif.option_info.menu_itemname_input_p2controller, vardata = textImgNew(), vardisplay = options.f_itemnameController(2)},
	{data = textImgNew(), itemname = 'p3controller', displayname = motif.option_info.menu_itemname_input_p3controller, vardata = textImgNew(), vardisplay = options.f_itemnameController(3)},
	{data = textImgNew(), itemname = 'p4controller', displayname = motif.option_info.menu_itemname_input_p4controller, vardata = textImgNew(), vardisplay = options.f_itemnameController(4)},
	{data = textImgNew(), itemname = 'defaultvalues', displayname = motif.option_info.menu_itemname_input_default},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_input_back},
}
options.t_inputCfg = main.f_cleanTable(options.t_inputCfg, main.t_sort.option_info)

function options.f_inputCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = options.t_inputCfg
	textImgSetText(txt_title, motif.option_info.title_text_input)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		elseif main.f_btnPalNo(main.p1Cmd) > 0 then
			--Keyboard / Gamepad
			if t[item].itemname == 'keyboard' or t[item].itemname == 'gamepad' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.f_keyCfg(t[item].itemname)
			--System (not implemented yet)
			elseif t[item].itemname == 'system' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			--P1-4 Controller
			elseif t[item].itemname:match('p[1-4]controller') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				local player = tonumber(t[item].itemname:match('p([1-4])controller'))
				options.f_controllerCfg(player)
				t[item].vardisplay = options.f_itemnameController(player)
			--Default Values
			elseif t[item].itemname == 'defaultvalues' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				options.f_keyDefault()
				for i = 1, #t do
					local player = t[i].itemname:match('p([1-4])controller')
					if player ~= nil then
						t[i].vardisplay = options.f_itemnameController(tonumber(player))
					end
				end
				modified = 1
				needReload = 1
			--Back
			elseif t[item].itemname == 'back' then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				textImgSetText(txt_title, motif.option_info.title_text_main)
				break
			end
		end
		options.f_menuCommonDraw(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; KEY SETTINGS
--;===========================================================
local t_keyCfg = {
	{data = textImgNew(), itemname = 'dummy', displayname = ''},
	{data = textImgNew(), itemname = 'configall', displayname = motif.option_info.menu_itemname_key_all, infodata = textImgNew(), infodisplay = ''},
	{data = textImgNew(), itemname = 'up', displayname = motif.option_info.menu_itemname_key_up, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'down', displayname = motif.option_info.menu_itemname_key_down, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'left', displayname = motif.option_info.menu_itemname_key_left, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'right', displayname = motif.option_info.menu_itemname_key_right, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'a', displayname = motif.option_info.menu_itemname_key_a, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'b', displayname = motif.option_info.menu_itemname_key_b, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'c', displayname = motif.option_info.menu_itemname_key_c, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'x', displayname = motif.option_info.menu_itemname_key_x, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'y', displayname = motif.option_info.menu_itemname_key_y, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'z', displayname = motif.option_info.menu_itemname_key_z, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'start', displayname = motif.option_info.menu_itemname_key_start, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'v', displayname = motif.option_info.menu_itemname_key_v, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'w', displayname = motif.option_info.menu_itemname_key_w, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'move', displayname = '', infodata = textImgNew(), infodisplay = ''},
}
--t_keyCfg = main.f_cleanTable(t_keyCfg, main.t_sort.option_info)

local txt_keyController = textImgNew()
function options.f_keyCfg(controller)
	main.f_cmdInput()
	local cursorPosY = 2
	local moveTxt = 0
	local item = 2
	local item_start = 2
	local t = t_keyCfg
	local t_pos = {motif.option_info.menu_key_p1_pos, motif.option_info.menu_key_p2_pos}
	local configall = false
	local key = ''
	local t_keyList = {[-1] = {}, [0] = {}, [1] = {}, [2] = {}, [3] = {}}
	local t_conflict = {[-1] = false, [0] = false, [1] = false, [2] = false, [3] = false}
	local conflict = false
	local btnReleased = 0
	local side = 1
	local add = 0
	local num = main.t_controllers[controller][1]
	local s_btn = tostring(config.KeyConfig[num].Buttons[item - item_start])
	local itemname = ''
	textImgSetText(txt_title, motif.option_info.title_text_key)
	for i = 1, #config.KeyConfig[1].Buttons do
		for j = 1, #main.t_controllers[controller] do
			t_keyCfg[i + item_start]['vardisplay' .. j] = config.KeyConfig[main.t_controllers[controller][j]].Buttons[i]
			if tostring(config.KeyConfig[main.t_controllers[controller][j]].Buttons[i]) ~= tostring(motif.option_info.menu_itemname_info_disable) then
				if t_keyList[config.KeyConfig[main.t_controllers[controller][j]].Joystick][tostring(config.KeyConfig[main.t_controllers[controller][j]].Buttons[i])] == nil then
					t_keyList[config.KeyConfig[main.t_controllers[controller][j]].Joystick][tostring(config.KeyConfig[main.t_controllers[controller][j]].Buttons[i])] = 1
				else
					t_keyList[config.KeyConfig[main.t_controllers[controller][j]].Joystick][tostring(config.KeyConfig[main.t_controllers[controller][j]].Buttons[i])] = t_keyList[config.KeyConfig[main.t_controllers[controller][j]].Joystick][tostring(config.KeyConfig[main.t_controllers[controller][j]].Buttons[i])] + 1
				end
			end
		end
	end
	while true do
		if configall then
			if controller == 'keyboard' then
				key = getKey()
			elseif getKey() == 'SPACE' then
				key = 'SPACE'
			else
				local tmp = getJoystickKey(config.KeyConfig[num].Joystick)
				if tonumber(tmp) == nil then --button released
					btnReleased = 1
				elseif btnReleased == 1 then --button pressed after releasing button once
					key = tmp
					btnReleased = 0
				end
			end
			if esc() then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				configall = false
				commandBufReset(main.p1Cmd)
			elseif key ~= '' then
				if key == 'SPACE' then
					sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
					if t_keyList[config.KeyConfig[num].Joystick][s_btn] ~= nil and t_keyList[config.KeyConfig[num].Joystick][s_btn] > 1 then
						t_keyList[config.KeyConfig[num].Joystick][s_btn] = t_keyList[config.KeyConfig[num].Joystick][s_btn] - 1
					else
						t_keyList[config.KeyConfig[num].Joystick][s_btn] = nil
					end
					t[item]['vardisplay' .. side + add] = motif.option_info.menu_itemname_info_disable
					config.KeyConfig[num].Buttons[item - item_start] = motif.option_info.menu_itemname_info_disable
					modified = 1
					needReload = 1
				elseif controller == 'keyboard' or (controller == 'gamepad' and tonumber(key) ~= nil) then
					sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
					if t_keyList[config.KeyConfig[num].Joystick][s_btn] ~= nil and t_keyList[config.KeyConfig[num].Joystick][s_btn] > 1 then
						t_keyList[config.KeyConfig[num].Joystick][s_btn] = t_keyList[config.KeyConfig[num].Joystick][s_btn] - 1
					else
						t_keyList[config.KeyConfig[num].Joystick][s_btn] = nil
					end
					if t_keyList[config.KeyConfig[num].Joystick][tostring(key)] == nil then
						t_keyList[config.KeyConfig[num].Joystick][tostring(key)] = 1
					else
						t_keyList[config.KeyConfig[num].Joystick][tostring(key)] = t_keyList[config.KeyConfig[num].Joystick][tostring(key)] + 1
					end
					t[item]['vardisplay' .. side + add] = key
					if controller == 'keyboard' then
						config.KeyConfig[num].Buttons[item - item_start] = tostring(key)
					else
						config.KeyConfig[num].Buttons[item - item_start] = tonumber(key)
					end
					modified = 1
					needReload = 1
				else
					sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				end
				item = item + 1
				if cursorPosY < motif.option_info.menu_window_visibleitems then
					cursorPosY = cursorPosY + 1
				end
				if item > 15 then
					item = item_start
					cursorPosY = item_start
					configall = false
					commandBufReset(main.p1Cmd)
				end
			end
			resetKey()
			key = ''
		elseif commandGetState(main.p1Cmd, 'u') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			item = item - 1
		elseif commandGetState(main.p1Cmd, 'd') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			item = item + 1
		elseif commandGetState(main.p1Cmd, 'l') or commandGetState(main.p1Cmd, 'r') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if side == 1 then
				side = 2
			else
				side = 1
			end
			num = main.t_controllers[controller][side + add]
		end
		--cursor position calculation
		if item < item_start then
			item = #t
			if #t > motif.option_info.menu_window_visibleitems then
				cursorPosY = motif.option_info.menu_window_visibleitems
			else
				cursorPosY = #t
			end
		elseif item > #t then
			item = item_start
			cursorPosY = item_start
		elseif configall == false then
			if commandGetState(main.p1Cmd, 'u') and cursorPosY > item_start then
				cursorPosY = cursorPosY - 1
			elseif commandGetState(main.p1Cmd, 'd') and cursorPosY < motif.option_info.menu_window_visibleitems then
				cursorPosY = cursorPosY + 1
			end
		end
		if cursorPosY == motif.option_info.menu_window_visibleitems then
			moveTxt = (item - motif.option_info.menu_window_visibleitems) * motif.option_info.menu_item_spacing[2]
		elseif cursorPosY == item_start then
			moveTxt = (item - item_start) * motif.option_info.menu_item_spacing[2]
		end
		s_btn = tostring(config.KeyConfig[num].Buttons[item - item_start])
		if controller == 'keyboard' then
			conflict = t_conflict[-1]
		elseif t_conflict[0] or t_conflict[1] or t_conflict[2] or t_conflict[3] then
			conflict = true
		else
			conflict = false
		end
		if configall == false then
			if esc() and conflict == false then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				textImgSetText(txt_title, motif.option_info.title_text_input)
				break
			--Config all
			elseif (t[item].itemname == 'configall' and main.f_btnPalNo(main.p1Cmd) > 0) or getKey() == 'F1' or getKey() == 'F2' or getKey() == 'F3' or getKey() == 'F4' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				if getKey() == 'F1' then
					add = 0
					side = 1
					num = main.t_controllers[controller][side + add]
				elseif getKey() == 'F2' then
					add = 0
					side = 2
					num = main.t_controllers[controller][side + add]
				elseif getKey() == 'F3' then
					add = 2
					side = 1
					num = main.t_controllers[controller][side + add]
				elseif getKey() == 'F4' then
					add = 2
					side = 2
					num = main.t_controllers[controller][side + add]
				end
				if controller == 'gamepad' and getJoystickPresent(config.KeyConfig[num].Joystick) == false then
					main.f_warning(main.f_extractText(motif.warning_info.text_pad), motif.option_info, motif.optionbgdef)
					item = item_start
					cursorPosY = item_start
				else
					resetKey()
					item = item_start + 1
					cursorPosY = item_start + 1
					btnReleased = 0
					configall = true
				end
			--Back / Next / Previous
			elseif (t[item].itemname == 'move' and main.f_btnPalNo(main.p1Cmd) > 0) or getKey() == 'TAB' then
				if side == 1 and getKey() ~= 'TAB' then --Back
					if conflict then
						main.f_warning(main.f_extractText(motif.warning_info.text_keys), motif.option_info, motif.optionbgdef)
					else
						sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
						textImgSetText(txt_title, motif.option_info.title_text_input)
						break
					end
				elseif add == 0 then --Next
					sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
					add = 2
					num = main.t_controllers[controller][side + add]
				else --Previous
					sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
					add = 0
					num = main.t_controllers[controller][side + add]
				end
				resetKey()
			--Buttons
			elseif main.f_btnPalNo(main.p1Cmd) > 0 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				if controller == 'gamepad' and getJoystickPresent(config.KeyConfig[num].Joystick) == false then
					main.f_warning(main.f_extractText(motif.warning_info.text_pad), motif.option_info, motif.optionbgdef)
				else
					key = main.f_input(main.f_extractText(motif.option_info.input_text_key), motif.option_info, motif.optionbgdef, controller, config.KeyConfig[num].Joystick, 'SPACE')
					if key == 'SPACE' then
						sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
						if t_keyList[config.KeyConfig[num].Joystick][s_btn] ~= nil and t_keyList[config.KeyConfig[num].Joystick][s_btn] > 1 then
							t_keyList[config.KeyConfig[num].Joystick][s_btn] = t_keyList[config.KeyConfig[num].Joystick][s_btn] - 1
						else
							t_keyList[config.KeyConfig[num].Joystick][s_btn] = nil
						end
						t[item]['vardisplay' .. side + add] = motif.option_info.menu_itemname_info_disable
						config.KeyConfig[num].Buttons[item - item_start] = motif.option_info.menu_itemname_info_disable
						modified = 1
						needReload = 1
					elseif (controller == 'keyboard' and key ~= '') or (controller == 'gamepad' and tonumber(key) ~= nil) then
						sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
						if t_keyList[config.KeyConfig[num].Joystick][s_btn] ~= nil and t_keyList[config.KeyConfig[num].Joystick][s_btn] > 1 then
							t_keyList[config.KeyConfig[num].Joystick][s_btn] = t_keyList[config.KeyConfig[num].Joystick][s_btn] - 1
						else
							t_keyList[config.KeyConfig[num].Joystick][s_btn] = nil
						end
						if t_keyList[config.KeyConfig[num].Joystick][tostring(key)] == nil then
							t_keyList[config.KeyConfig[num].Joystick][tostring(key)] = 1
						else
							t_keyList[config.KeyConfig[num].Joystick][tostring(key)] = t_keyList[config.KeyConfig[num].Joystick][tostring(key)] + 1
						end
						t[item]['vardisplay' .. side + add] = key
						if controller == 'keyboard' then
							config.KeyConfig[num].Buttons[item - item_start] = tostring(key)
						else
							config.KeyConfig[num].Buttons[item - item_start] = tonumber(key)
						end
						modified = 1
						needReload = 1
					else
						sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
					end
					resetKey()
					key = ''
				end
			end
		end
		--draw clearcolor
		clearColor(motif.optionbgdef.bgclearcolor[1], motif.optionbgdef.bgclearcolor[2], motif.optionbgdef.bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(motif.optionbgdef.bg, false)
		--draw controller itemname, reset conflict
		for i = 1, 2 do
			if controller == 'keyboard' then
				t_conflict[-1] = false
				itemname = main.f_extractText(motif.option_info.menu_itemname_key_keyboard, i + add)[1]
			else
				t_conflict[config.KeyConfig[main.t_controllers[controller][i + add]].Joystick] = false
				itemname = main.f_extractText(motif.option_info.menu_itemname_key_gamepad, i + add)[1]
			end
			textImgDraw(main.f_updateTextImg(
				txt_keyController,
				motif.font_data[motif.option_info['menu_item_controller' .. i .. '_font'][1]],
				motif.option_info['menu_item_controller' .. i .. '_font'][2],
				motif.option_info['menu_item_controller' .. i .. '_font'][3],
				itemname,
				motif.option_info['menu_item_p' .. i .. '_pos'][1],
				motif.option_info['menu_item_p' .. i .. '_pos'][2],
				motif.option_info['menu_item_controller' .. i .. '_font_scale'][1],
				motif.option_info['menu_item_controller' .. i .. '_font_scale'][2],
				motif.option_info['menu_item_controller' .. i .. '_font'][4],
				motif.option_info['menu_item_controller' .. i .. '_font'][5],
				motif.option_info['menu_item_controller' .. i .. '_font'][6],
				motif.option_info['menu_item_controller' .. i .. '_font'][7],
				motif.option_info['menu_item_controller' .. i .. '_font'][8]
			))
		end
		--draw menu box
		if motif.option_info.menu_boxbg_visible == 1 then
			local coord4 = 0
			for i = 1, #t_pos do
				if #t > motif.option_info.menu_window_visibleitems and moveTxt == (#t - motif.option_info.menu_window_visibleitems) * motif.option_info.menu_key_item_spacing[2] then
					coord4 = motif.option_info.menu_window_visibleitems * (motif.option_info.menu_key_boxcursor_coords[4] - motif.option_info.menu_key_boxcursor_coords[2] + 1) + main.f_oddRounding(motif.option_info.menu_key_boxcursor_coords[2])
				else
					coord4 = #t * (motif.option_info.menu_key_boxcursor_coords[4] - motif.option_info.menu_key_boxcursor_coords[2] + 1) + main.f_oddRounding(motif.option_info.menu_key_boxcursor_coords[2])
				end
				fillRect(
					t_pos[i][1] + motif.option_info.menu_key_boxcursor_coords[1],
					t_pos[i][2] + motif.option_info.menu_key_boxcursor_coords[2],
					motif.option_info.menu_key_boxcursor_coords[3] - motif.option_info.menu_key_boxcursor_coords[1] + 1,
					coord4,
					motif.option_info.menu_boxbg_col[1],
					motif.option_info.menu_boxbg_col[2],
					motif.option_info.menu_boxbg_col[3],
					motif.option_info.menu_boxbg_alpha[1],
					motif.option_info.menu_boxbg_alpha[2]
				)
			end
		end
		--draw title
		textImgDraw(txt_title)
		--draw menu items
		for i = 1, #t do
			for j = 1, #t_pos do
				if i > item - cursorPosY then
					if t[i].itemname == 'configall' then
						if add == 0 then --page 1
							if j == 1 then --left side
								t[i].infodisplay = motif.option_info.menu_itemname_info_f1
							else --right side
								t[i].infodisplay = motif.option_info.menu_itemname_info_f2
							end
						else --page 2
							if j == 1 then --left side
								t[i].infodisplay = motif.option_info.menu_itemname_info_f3
							else --right side
								t[i].infodisplay = motif.option_info.menu_itemname_info_f4
							end
						end
					elseif i == #t then --last row
						if j == 1 then --left side
							t[i].displayname = motif.option_info.menu_itemname_key_back
							t[i].infodisplay = motif.option_info.menu_itemname_info_esc
						elseif add == 0 then --right side, page 1
							t[i].displayname = motif.option_info.menu_itemname_key_next
							t[i].infodisplay = motif.option_info.menu_itemname_info_tab
						else --right side, page 2
							t[i].displayname = motif.option_info.menu_itemname_key_previous
							t[i].infodisplay = motif.option_info.menu_itemname_info_tab
						end
					end
					if i == item and j == side then --active item
						--draw displayname
						textImgDraw(main.f_updateTextImg(
							t[i].data,
							motif.font_data[motif.option_info.menu_item_active_font[1]],
							motif.option_info.menu_item_active_font[2],
							motif.option_info.menu_item_active_font[3],
							t[i].displayname,
							t_pos[j][1],
							t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
							motif.option_info.menu_item_active_font_scale[1],
							motif.option_info.menu_item_active_font_scale[2],
							motif.option_info.menu_item_active_font[4],
							motif.option_info.menu_item_active_font[5],
							motif.option_info.menu_item_active_font[6],
							motif.option_info.menu_item_active_font[7],
							motif.option_info.menu_item_active_font[8]
						))
						--draw vardata
						if t[i].vardata ~= nil then
							if t_keyList[config.KeyConfig[main.t_controllers[controller][j + add]].Joystick][tostring(t[i]['vardisplay' .. j + add])] ~= nil and t_keyList[config.KeyConfig[main.t_controllers[controller][j + add]].Joystick][tostring(t[i]['vardisplay' .. j + add])] > 1 then
								textImgDraw(main.f_updateTextImg(
									t[i].vardata,
									motif.font_data[motif.option_info.menu_item_value_conflict_font[1]],
									motif.option_info.menu_item_value_conflict_font[2],
									motif.option_info.menu_item_value_conflict_font[3],
									t[i]['vardisplay' .. j + add],
									t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
									t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
									motif.option_info.menu_item_value_conflict_font_scale[1],
									motif.option_info.menu_item_value_conflict_font_scale[2],
									motif.option_info.menu_item_value_conflict_font[4],
									motif.option_info.menu_item_value_conflict_font[5],
									motif.option_info.menu_item_value_conflict_font[6],
									motif.option_info.menu_item_value_conflict_font[7],
									motif.option_info.menu_item_value_conflict_font[8]
								))
								t_conflict[config.KeyConfig[main.t_controllers[controller][j + add]].Joystick] = true
							else
								textImgDraw(main.f_updateTextImg(
									t[i].vardata,
									motif.font_data[motif.option_info.menu_item_value_active_font[1]],
									motif.option_info.menu_item_value_active_font[2],
									motif.option_info.menu_item_value_active_font[3],
									t[i]['vardisplay' .. j + add],
									t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
									t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
									motif.option_info.menu_item_value_active_font_scale[1],
									motif.option_info.menu_item_value_active_font_scale[2],
									motif.option_info.menu_item_value_active_font[4],
									motif.option_info.menu_item_value_active_font[5],
									motif.option_info.menu_item_value_active_font[6],
									motif.option_info.menu_item_value_active_font[7],
									motif.option_info.menu_item_value_active_font[8]
								))
							end
						--draw infodata
						elseif t[i].infodata ~= nil then
							textImgDraw(main.f_updateTextImg(
								t[i].infodata,
								motif.font_data[motif.option_info.menu_item_info_active_font[1]],
								motif.option_info.menu_item_info_active_font[2],
								motif.option_info.menu_item_info_active_font[3],
								t[i].infodisplay,
								t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
								t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
								motif.option_info.menu_item_value_active_font_scale[1],
								motif.option_info.menu_item_value_active_font_scale[2],
								motif.option_info.menu_item_info_active_font[4],
								motif.option_info.menu_item_info_active_font[5],
								motif.option_info.menu_item_info_active_font[6],
								motif.option_info.menu_item_info_active_font[7],
								motif.option_info.menu_item_info_active_font[8]
							))
						end
					else --inactive item
						--draw displayname
						textImgDraw(main.f_updateTextImg(
							t[i].data,
							motif.font_data[motif.option_info.menu_item_font[1]],
							motif.option_info.menu_item_font[2],
							motif.option_info.menu_item_font[3],
							t[i].displayname,
							t_pos[j][1],
							t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
							motif.option_info.menu_item_font_scale[1],
							motif.option_info.menu_item_font_scale[2],
							motif.option_info.menu_item_font[4],
							motif.option_info.menu_item_font[5],
							motif.option_info.menu_item_font[6],
							motif.option_info.menu_item_font[7],
							motif.option_info.menu_item_font[8]
						))
						--draw vardata
						if t[i].vardata ~= nil then
							if t_keyList[config.KeyConfig[main.t_controllers[controller][j + add]].Joystick][tostring(t[i]['vardisplay' .. j + add])] ~= nil and t_keyList[config.KeyConfig[main.t_controllers[controller][j + add]].Joystick][tostring(t[i]['vardisplay' .. j + add])] > 1 then
								textImgDraw(main.f_updateTextImg(
									t[i].vardata,
									motif.font_data[motif.option_info.menu_item_value_conflict_font[1]],
									motif.option_info.menu_item_value_conflict_font[2],
									motif.option_info.menu_item_value_conflict_font[3],
									t[i]['vardisplay' .. j + add],
									t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
									t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
									motif.option_info.menu_item_value_conflict_font_scale[1],
									motif.option_info.menu_item_value_conflict_font_scale[2],
									motif.option_info.menu_item_value_conflict_font[4],
									motif.option_info.menu_item_value_conflict_font[5],
									motif.option_info.menu_item_value_conflict_font[6],
									motif.option_info.menu_item_value_conflict_font[7],
									motif.option_info.menu_item_value_conflict_font[8]
								))
								t_conflict[config.KeyConfig[main.t_controllers[controller][j + add]].Joystick] = true
							else
								textImgDraw(main.f_updateTextImg(
									t[i].vardata,
									motif.font_data[motif.option_info.menu_item_value_font[1]],
									motif.option_info.menu_item_value_font[2],
									motif.option_info.menu_item_value_font[3],
									t[i]['vardisplay' .. j + add],
									t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
									t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
									motif.option_info.menu_item_value_font_scale[1],
									motif.option_info.menu_item_value_font_scale[2],
									motif.option_info.menu_item_value_font[4],
									motif.option_info.menu_item_value_font[5],
									motif.option_info.menu_item_value_font[6],
									motif.option_info.menu_item_value_font[7],
									motif.option_info.menu_item_value_font[8]
								))
							end
						--draw infodata
						elseif t[i].infodata ~= nil then
							textImgDraw(main.f_updateTextImg(
								t[i].infodata,
								motif.font_data[motif.option_info.menu_item_info_font[1]],
								motif.option_info.menu_item_info_font[2],
								motif.option_info.menu_item_info_font[3],
								t[i].infodisplay,
								t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
								t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
								motif.option_info.menu_item_value_active_font_scale[1],
								motif.option_info.menu_item_value_active_font_scale[2],
								motif.option_info.menu_item_info_font[4],
								motif.option_info.menu_item_info_font[5],
								motif.option_info.menu_item_info_font[6],
								motif.option_info.menu_item_info_font[7],
								motif.option_info.menu_item_info_font[8]
							))
						end
					end
				end
			end
		end
		--draw menu cursor
		if motif.option_info.menu_boxcursor_visible == 1 then
			local src, dst = main.f_boxcursorAlpha(
				motif.option_info.menu_boxcursor_alpharange[1],
				motif.option_info.menu_boxcursor_alpharange[2],
				motif.option_info.menu_boxcursor_alpharange[3],
				motif.option_info.menu_boxcursor_alpharange[4],
				motif.option_info.menu_boxcursor_alpharange[5],
				motif.option_info.menu_boxcursor_alpharange[6]
			)
			for i = 1, #t_pos do
				if i == side then
					fillRect(
						t_pos[i][1] + motif.option_info.menu_key_boxcursor_coords[1],
						t_pos[i][2] + motif.option_info.menu_key_boxcursor_coords[2] + (cursorPosY - 1) * motif.option_info.menu_key_item_spacing[2],
						motif.option_info.menu_key_boxcursor_coords[3] - motif.option_info.menu_key_boxcursor_coords[1] + 1,
						motif.option_info.menu_key_boxcursor_coords[4] - motif.option_info.menu_key_boxcursor_coords[2] + 1 + main.f_oddRounding(motif.option_info.menu_key_boxcursor_coords[2]),
						motif.option_info.menu_boxcursor_col[1],
						motif.option_info.menu_boxcursor_col[2],
						motif.option_info.menu_boxcursor_col[3],
						src,
						dst
					)
				end
			end
		end
		--draw layerno = 1 backgrounds
		bgDraw(motif.optionbgdef.bg, true)
		main.f_cmdInput()
		refresh()
	end
end

--;===========================================================
--; CONTROLLER SETTINGS
--;===========================================================
local t_controllerCfg = {
	{data = textImgNew(), itemname = 'keyboard1', displayname = main.f_extractText(motif.option_info.menu_itemname_controller_keyboard, 1)[1]},
	{data = textImgNew(), itemname = 'keyboard2', displayname = main.f_extractText(motif.option_info.menu_itemname_controller_keyboard, 2)[1]},
	{data = textImgNew(), itemname = 'keyboard3', displayname = main.f_extractText(motif.option_info.menu_itemname_controller_keyboard, 3)[1]},
	{data = textImgNew(), itemname = 'keyboard4', displayname = main.f_extractText(motif.option_info.menu_itemname_controller_keyboard, 4)[1]},
	{data = textImgNew(), itemname = 'gamepad1', displayname = main.f_extractText(motif.option_info.menu_itemname_controller_gamepad, 1)[1]},
	{data = textImgNew(), itemname = 'gamepad2', displayname = main.f_extractText(motif.option_info.menu_itemname_controller_gamepad, 2)[1]},
	{data = textImgNew(), itemname = 'gamepad3', displayname = main.f_extractText(motif.option_info.menu_itemname_controller_gamepad, 3)[1]},
	{data = textImgNew(), itemname = 'gamepad4', displayname = main.f_extractText(motif.option_info.menu_itemname_controller_gamepad, 4)[1]},
	{data = textImgNew(), itemname = 'disable', displayname = motif.option_info.menu_itemname_controller_disabled},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_controller_back},
}
t_controllerCfg = main.f_cleanTable(t_controllerCfg, main.t_sort.option_info)

function options.f_controllerCfg(player)
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_controllerCfg
	textImgSetText(txt_title, motif.option_info.title_text_controller)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
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
			--Not assigned
			elseif t[item].itemname == 'disable' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				for i = 1, #config.KeyConfig do
					if config.KeyConfig[i].Player == player then
						config.KeyConfig[i].Player = 0
						main.t_controllers.player[i] = 0
					end
				end
				modified = 1
				textImgSetText(txt_title, motif.option_info.title_text_input)
				break
			--Keyboard 1-4 / Gamepad 1-4
			else
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				local controller, cnt = t[item].itemname:match('^([^0-9]+)([0-9]+)$')
				local num = main.t_controllers[controller][tonumber(cnt)]
				for i = 1, #config.KeyConfig do
					if i == num then
						config.KeyConfig[i].Player = player
						main.t_controllers.player[i] = player
					elseif config.KeyConfig[i].Player == player then
						config.KeyConfig[i].Player = 0
						main.t_controllers.player[i] = 0
					end
				end
				modified = 1
				textImgSetText(txt_title, motif.option_info.title_text_input)
				break
			end
		end
		options.f_menuCommonDraw(cursorPosY, moveTxt, item, t)
	end
end

return options

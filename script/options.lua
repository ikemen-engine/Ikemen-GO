
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

--return table entry (or ret if specified) if provided key exists in the table, otherwise return default argument
function options.f_definedDisplay(key, t, default, ret)
	if t[key] ~= nil then
		return ret or t[key]
	end
	return default
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
	for i = 1, #config.KeyConfig do
		if i == 1 then
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
		elseif i == 2 then
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
				config.KeyConfig[i].Buttons[j] = tostring(motif.option_info.menu_itemname_info_disable)
			end
		end
	end
	for i = 1, #config.JoystickConfig do
		config.JoystickConfig[i].Buttons[1] = '-3'
		config.JoystickConfig[i].Buttons[2] = '-4'
		config.JoystickConfig[i].Buttons[3] = '-1'
		config.JoystickConfig[i].Buttons[4] = '-2'
		config.JoystickConfig[i].Buttons[5] = '0'
		config.JoystickConfig[i].Buttons[6] = '1'
		config.JoystickConfig[i].Buttons[7] = '4'
		config.JoystickConfig[i].Buttons[8] = '2'
		config.JoystickConfig[i].Buttons[9] = '3'
		config.JoystickConfig[i].Buttons[10] = '5'
		config.JoystickConfig[i].Buttons[11] = '7'
		config.JoystickConfig[i].Buttons[12] = '-10'
		config.JoystickConfig[i].Buttons[13] = '-12'
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
			roundtime = options.f_definedDisplay(config.RoundTime, {[-1] = motif.option_info.menu_itemname_arcade_roundtime_none}, config.RoundTime),
			roundsnumsingle = options.roundsNumSingle,
			roundsnumteam = options.roundsNumTeam,
			maxdrawgames = options.maxDrawGames,
			difficulty = config.Difficulty,
			credits = config.Credits,
			quickcontinue = options.f_boolDisplay(config.QuickContinue),
			airamping = options.f_boolDisplay(config.AIRamping),
			airandomcolor = options.f_boolDisplay(config.AIRandomColor, motif.option_info.menu_itemname_arcade_aipalette_random, motif.option_info.menu_itemname_arcade_aipalette_default),
		},
		t_videoCfg = {
			resolution = config.Width .. 'x' .. config.Height,
			fullscreen = options.f_boolDisplay(config.Fullscreen),
			msaa = options.f_boolDisplay(config.MSAA, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled),
			externalshaders = motif.option_info.menu_itemname_disabled,
		},
		t_audioCfg = {
			mastervolume = config.MasterVolume .. '%',
			bgmvolume = config.BgmVolume .. '%',
			sfxvolume = config.WavVolume .. '%',
			audioducking = options.f_boolDisplay(config.AudioDucking, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled),
		},
		t_gameplayCfg = {
			lifemul = config.LifeMul .. '%',
			autoguard = options.f_boolDisplay(config.AutoGuard),
			team1vs2life = config.Team1VS2Life .. '%',
			turnsrecoverybase = config.TurnsRecoveryBase .. '%',
			turnsrecoverybonus = config.TurnsRecoveryBonus .. '%',
			teampowershare = options.f_boolDisplay(config.TeamPowerShare),
			teamlifeshare = options.f_boolDisplay(config.TeamLifeShare),
		},
		t_ratioCfg = {
			ratio1Life = options.f_displayRatio(config.LifeRatio[1]),
			ratio1Attack = options.f_displayRatio(config.AttackRatio[1]),
			ratio2Life = options.f_displayRatio(config.LifeRatio[2]),
			ratio2Attack = options.f_displayRatio(config.AttackRatio[2]),
			ratio3Life = options.f_displayRatio(config.LifeRatio[3]),
			ratio3Attack = options.f_displayRatio(config.AttackRatio[3]),
			ratio4Life = options.f_displayRatio(config.LifeRatio[4]),
			ratio4Attack = options.f_displayRatio(config.AttackRatio[4]),
		},
		t_advGameplayCfg = {
			attackpowermul = config['Attack.LifeToPowerMul'],
			gethitpowermul = config['GetHit.LifeToPowerMul'],
			superdefencemul = config['Super.TargetDefenceMul'],
			singlemode = options.f_boolDisplay(config.SingleTeamMode, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled),
			numturns = options.f_checkTeamAmount(config.NumTurns, 1, motif.option_info.menu_itemname_disabled),
			numsimul = options.f_checkTeamAmount(config.NumSimul, 1, motif.option_info.menu_itemname_disabled),
			numtag = options.f_checkTeamAmount(config.NumTag, 1, motif.option_info.menu_itemname_disabled),
		},
		t_engineCfg = {
			allowdebugkeys = options.f_boolDisplay(config.AllowDebugKeys, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled),
			simulmode = options.f_boolDisplay(config.SimulMode, motif.option_info.menu_itemname_disabled, motif.option_info.menu_itemname_enabled),
			lifebarfontscale = config.LifebarFontScale,
			helpermax = config.HelperMax,
			playerprojectilemax = config.PlayerProjectileMax,
			explodmax = config.ExplodMax,
			afterimagemax = config.AfterImageMax,
			zoomactive = options.f_boolDisplay(config.ZoomActive),
			maxzoomout = config.ZoomMin,
			maxzoomin = config.ZoomMax,
			zoomspeed = config.ZoomSpeed,
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
	motif.option_info.title_font[6],
	motif.option_info.title_font[7],
	motif.option_info.title_font[8]
	--motif.defaultOptions --title font assignment exists in mugen
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
			motif.option_info.menu_boxbg_alpha[2],
			motif.defaultOptions
		)
	end
	--draw title
	textImgDraw(txt_title)
	--draw menu items
	for i = 1, #t do
		if i > item - cursorPosY then
			if i == item then
				if t[i].selected then
					textImgDraw(main.f_updateTextImg(
						t[i].data,
						motif.font_data[motif.option_info.menu_item_selected_active_font[1]],
						motif.option_info.menu_item_selected_active_font[2],
						motif.option_info.menu_item_selected_active_font[3],
						t[i].displayname,
						motif.option_info.menu_pos[1],
						motif.option_info.menu_pos[2] + (i - 1) * motif.option_info.menu_item_spacing[2] - moveTxt,
						motif.option_info.menu_item_selected_active_font_scale[1],
						motif.option_info.menu_item_selected_active_font_scale[2],
						motif.option_info.menu_item_selected_active_font[4],
						motif.option_info.menu_item_selected_active_font[5],
						motif.option_info.menu_item_selected_active_font[6],
						motif.option_info.menu_item_selected_active_font[7],
						motif.option_info.menu_item_selected_active_font[8],
						motif.defaultOptions
					))
				else
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
						motif.option_info.menu_item_active_font[8],
						motif.defaultOptions
					))
				end
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
						motif.option_info.menu_item_value_active_font[8],
						motif.defaultOptions
					))
				end
			else
				if t[i].selected then
					textImgDraw(main.f_updateTextImg(
						t[i].data,
						motif.font_data[motif.option_info.menu_item_selected_font[1]],
						motif.option_info.menu_item_selected_font[2],
						motif.option_info.menu_item_selected_font[3],
						t[i].displayname,
						motif.option_info.menu_pos[1],
						motif.option_info.menu_pos[2] + (i - 1) * motif.option_info.menu_item_spacing[2] - moveTxt,
						motif.option_info.menu_item_selected_font_scale[1],
						motif.option_info.menu_item_selected_font_scale[2],
						motif.option_info.menu_item_selected_font[4],
						motif.option_info.menu_item_selected_font[5],
						motif.option_info.menu_item_selected_font[6],
						motif.option_info.menu_item_selected_font[7],
						motif.option_info.menu_item_selected_font[8],
						motif.defaultOptions
					))
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
						motif.option_info.menu_item_font[8],
						motif.defaultOptions
					))
				end
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
						motif.option_info.menu_item_value_font[8],
						motif.defaultOptions
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
			dst,
			motif.defaultOptions
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
	{data = textImgNew(), itemname = 'videosettings', displayname = motif.option_info.menu_itemname_main_video},
	{data = textImgNew(), itemname = 'audiosettings', displayname = motif.option_info.menu_itemname_main_audio},
	{data = textImgNew(), itemname = 'inputsettings', displayname = motif.option_info.menu_itemname_main_input},
	{data = textImgNew(), itemname = 'gameplaysettings', displayname = motif.option_info.menu_itemname_main_gameplay},
	{data = textImgNew(), itemname = 'enginesettings', displayname = motif.option_info.menu_itemname_main_engine},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'portchange', displayname = motif.option_info.menu_itemname_main_port, vardata = textImgNew(), vardisplay = getListenPort()},
	{data = textImgNew(), itemname = 'defaultvalues', displayname = motif.option_info.menu_itemname_main_default},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'save', displayname = motif.option_info.menu_itemname_main_save},
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
	if motif.music.option_bgm == '' then
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
			main.f_menuFade('option_info', 'fadeout', cursorPosY, moveTxt, item, t)
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
			--Gameplay Settings
			elseif t[item].itemname == 'gameplaysettings' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				options.f_gameplayCfg()
			--Engine Settings
			elseif t[item].itemname == 'enginesettings' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				options.f_engineCfg()
			--Default Values
			elseif t[item].itemname == 'defaultvalues' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				config.HelperMax = 56
				config.PlayerProjectileMax = 256
				config.ExplodMax = 512
				config.AfterImageMax = 128
				config.MasterVolume = 80
				config.WavVolume = 80
				config.BgmVolume = 80
				config['Attack.LifeToPowerMul'] = 0.7
				config['GetHit.LifeToPowerMul'] = 0.6
				config.Width = 640
				config.Height = 480
				config['Super.TargetDefenceMul'] = 1.5
				config.LifebarFontScale = 1
				--config.System = 'script/main.lua'
				options.f_keyDefault()
				--config.ControllerStickSensitivity = 0.4
				--config.XinputTriggerSensitivity = 0
				--config.Motif = 'data/system.def'
				--config.CommonAir = 'data/common.air'
				--config.CommonCmd = 'data/common.cmd'
				config.SimulMode = true
				config.LifeMul = 100
				config.Team1VS2Life = 100
				config.TurnsRecoveryBase = 12.5
				config.TurnsRecoveryBonus = 27.5
				config.ZoomActive = false
				config.ZoomMin = 0.75
				config.ZoomMax = 1.1
				config.ZoomSpeed = 1.0
				config.RoundTime = 99
				config.RoundsNumSingle = -1
				config.RoundsNumTeam = -1
				config.MaxDrawGames = -2
				config.SingleTeamMode = true
				config.NumTurns = 4
				config.NumSimul = 4
				config.NumTag = 4
				config.Difficulty = 8
				config.Credits = 10
				setListenPort(7500)
				config.QuickContinue = false
				config.AIRandomColor = true
				config.AIRamping = true
				config.AutoGuard = false
				config.TeamPowerShare = false
				config.TeamLifeShare = false
				config.Fullscreen = false
				config.AudioDucking = false
				config.QuickLaunch = 0
				config.AllowDebugKeys = true
				config.ComboExtraFrameWindow = 1
				config.ExternalShaders = {}
				config.LocalcoordScalingType = 1
				config.MSAA = false
				config.LifeRatio = {0.80, 1.0, 1.17, 1.40}
				config.AttackRatio = {0.82, 1.0, 1.17, 1.30}
				loadLifebar(motif.files.fight)
				options.roundsNumSingle = getMatchWins()
				options.roundsNumTeam = getMatchWins()
				options.maxDrawGames = getMatchMaxDrawGames()
				options.f_resetTables()
				modified = 1
				needReload = 1
			--Save and Return
			elseif t[item].itemname == 'save' then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				if modified == 1 then
					options.f_saveCfg()
				end
				main.f_menuFade('option_info', 'fadeout', cursorPosY, moveTxt, item, t)
				if motif.music.option_bgm == '' then
					main.f_menuReset(motif.titlebgdef.bg)
				else
					main.f_menuReset(motif.titlebgdef.bg, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
				end
				break
			--Return Without Saving
			elseif t[item].itemname == 'back' then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				if needReload == 1 then
					main.f_warning(main.f_extractText(motif.warning_info.text_noreload), motif.option_info, motif.optionbgdef)
				end
				main.f_menuFade('option_info', 'fadeout', cursorPosY, moveTxt, item, t)
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
	{data = textImgNew(), itemname = 'roundtime', displayname = motif.option_info.menu_itemname_arcade_roundtime, vardata = textImgNew(), vardisplay = options.f_definedDisplay(config.RoundTime, {[-1] = motif.option_info.menu_itemname_arcade_roundtime_none}, config.RoundTime)},
	{data = textImgNew(), itemname = 'roundsnumsingle', displayname = motif.option_info.menu_itemname_arcade_roundsnumsingle, vardata = textImgNew(), vardisplay = options.roundsNumSingle},
	{data = textImgNew(), itemname = 'roundsnumteam', displayname = motif.option_info.menu_itemname_arcade_roundsnumteam, vardata = textImgNew(), vardisplay = options.roundsNumTeam},
	{data = textImgNew(), itemname = 'maxdrawgames', displayname = motif.option_info.menu_itemname_arcade_maxdrawgames, vardata = textImgNew(), vardisplay = options.maxDrawGames},
	{data = textImgNew(), itemname = 'difficulty', displayname = motif.option_info.menu_itemname_arcade_difficulty, vardata = textImgNew(), vardisplay = config.Difficulty},
	{data = textImgNew(), itemname = 'credits', displayname = motif.option_info.menu_itemname_arcade_credits, vardata = textImgNew(), vardisplay = config.Credits},
	{data = textImgNew(), itemname = 'quickcontinue', displayname = motif.option_info.menu_itemname_arcade_quickcontinue, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.QuickContinue)},
	{data = textImgNew(), itemname = 'airamping', displayname = motif.option_info.menu_itemname_arcade_airamping, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.AIRamping)},
	{data = textImgNew(), itemname = 'airandomcolor', displayname = motif.option_info.menu_itemname_arcade_aipalette, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.AIRandomColor, motif.option_info.menu_itemname_arcade_aipalette_random, motif.option_info.menu_itemname_arcade_aipalette_default)},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
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
		--Time Limit
		elseif t[item].itemname == 'roundtime' then
			if commandGetState(main.p1Cmd, 'r') and config.RoundTime < 1000 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.RoundTime = config.RoundTime + 1
				t[item].vardisplay = config.RoundTime
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.RoundTime > -1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.RoundTime = config.RoundTime - 1
				t[item].vardisplay = options.f_definedDisplay(config.RoundTime, {[-1] = motif.option_info.menu_itemname_arcade_roundtime_none}, config.RoundTime)
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
		elseif t[item].itemname == 'quickcontinue' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.QuickContinue then
				config.QuickContinue = false
			else
				config.QuickContinue = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.QuickContinue)
			modified = 1
		--AI Ramping
		elseif t[item].itemname == 'airamping' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AIRamping then
				config.AIRamping = false
			else
				config.AIRamping = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.AIRamping)
			modified = 1
		--AI Palette
		elseif t[item].itemname == 'airandomcolor' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AIRandomColor then
				config.AIRandomColor = false
			else
				config.AIRandomColor = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.AIRandomColor, motif.option_info.menu_itemname_arcade_aipalette_random, motif.option_info.menu_itemname_arcade_aipalette_default)
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
	{data = textImgNew(), itemname = 'msaa', displayname = motif.option_info.menu_itemname_video_msaa, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.MSAA, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled)},
	{data = textImgNew(), itemname = 'externalshaders', displayname = motif.option_info.menu_itemname_video_externalshaders, vardata = textImgNew(), vardisplay = options.f_definedDisplay(1, config.ExternalShaders, motif.option_info.menu_itemname_disabled, #config.ExternalShaders)},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
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
		--MSAA
		elseif t[item].itemname == 'msaa' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.MSAA then
				config.MSAA = false
			else
				config.MSAA = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.MSAA, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled)
			modified = 1
			needReload = 1
		--Shaders
		elseif t[item].itemname == 'externalshaders' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			options.f_shaderCfg()
			t[item].vardisplay = options.f_definedDisplay(1, config.ExternalShaders, motif.option_info.menu_itemname_disabled, #config.ExternalShaders)
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
	{data = textImgNew(), x = 320,  y = 240, displayname = motif.option_info.menu_itemname_video_res_320x240},
	{data = textImgNew(), x = 640,  y = 480, displayname = motif.option_info.menu_itemname_video_res_640x480},
	{data = textImgNew(), x = 1280, y = 960, displayname = motif.option_info.menu_itemname_video_res_1280x960},
	{data = textImgNew(), x = 1600, y = 1200, displayname = motif.option_info.menu_itemname_video_res_1600x1200},
	{data = textImgNew(), x = 960,  y = 720, displayname = motif.option_info.menu_itemname_video_res_960x720},
	{data = textImgNew(), x = 1280, y = 720, displayname = motif.option_info.menu_itemname_video_res_1280x720},
	{data = textImgNew(), x = 1600, y = 900, displayname = motif.option_info.menu_itemname_video_res_1600x900},
	{data = textImgNew(), x = 1920, y = 1080, displayname = motif.option_info.menu_itemname_video_res_1920x1080},
	{data = textImgNew(), x = 2560, y = 1440, displayname = motif.option_info.menu_itemname_video_res_2560x1440},
	{data = textImgNew(), x = 3840, y = 2160, displayname = motif.option_info.menu_itemname_video_res_3840x2160},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'custom', displayname = motif.option_info.menu_itemname_video_res_custom},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_video_res_back},
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
--; SHADER SETTINGS
--;===========================================================
local t_shaderCfg = {}
local t_shaders = {}
local t_files = GetDirectoryFiles('shaders')
for i = 1, #t_files do
	t_files[i]:gsub('^(.-)([^\\/]+)%.([^%.\\/]-)$', function(path, filename, ext)
		path = path:gsub('\\', '/')
		ext = ext:lower()
		if ext:match('vert') or ext:match('frag') --[[or ext:match('shader')]] then
			if t_shaders[path .. filename] == nil then
				table.insert(t_shaderCfg, {data = textImgNew(), itemname = path .. filename, displayname = filename, selected = false})
				t_shaders[path .. filename] = ''
			end
		end
	end)
end
if #t_shaderCfg > 0 then
	table.insert(t_shaderCfg, {data = textImgNew(), itemname = 'empty', displayname = ' '})
	table.insert(t_shaderCfg, {data = textImgNew(), itemname = 'disableall', displayname = motif.option_info.menu_itemname_video_externalshaders_disableall})
end
table.insert(t_shaderCfg, {data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_video_externalshaders_back})
t_shaderCfg = main.f_cleanTable(t_shaderCfg, main.t_sort.option_info)

function options.f_shaderCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = t_shaderCfg
	textImgSetText(txt_title, motif.option_info.title_text_externalshaders)
	if #t_shaderCfg == 1 then --only 'Back' option exists
		main.f_warning(main.f_extractText(motif.warning_info.text_shaders), motif.option_info, motif.optionbgdef)
	end
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
			--Disable all
			elseif t[item].itemname == 'disableall' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				if #config.ExternalShaders > 0 then
					config.ExternalShaders = {}
					for i = 1, #t do
						if t[i].selected then
							t[i].selected = false
						end
					end
					modified = 1
					needReload = 1
				end
				textImgSetText(txt_title, motif.option_info.title_text_video)
				break
			--Shader
			else
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				local found = false
				--get rid of shader reference if it exists in config.ExternalShaders
				for i = 1, #config.ExternalShaders do
					if config.ExternalShaders[i]:lower() == t[item].itemname:lower() then
						table.remove(config.ExternalShaders, i)
						t[item].selected = false
						found = true
						break
					end
				end
				--or add it if not
				if not found then
					table.insert(config.ExternalShaders, t[item].itemname)
					t[item].selected = true
				end
				modified = 1
				needReload = 1
				--textImgSetText(txt_title, motif.option_info.title_text_video)
				--break
			end
		end
		options.f_menuCommonDraw(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; AUDIO SETTINGS
--;===========================================================
options.t_audioCfg = {
	{data = textImgNew(), itemname = 'mastervolume', displayname = motif.option_info.menu_itemname_audio_mastervolume, vardata = textImgNew(), vardisplay = config.MasterVolume .. '%'},
	{data = textImgNew(), itemname = 'bgmvolume', displayname = motif.option_info.menu_itemname_audio_bgmvolume, vardata = textImgNew(), vardisplay = config.BgmVolume .. '%'},
	{data = textImgNew(), itemname = 'sfxvolume', displayname = motif.option_info.menu_itemname_audio_sfxvolume, vardata = textImgNew(), vardisplay = config.WavVolume .. '%'},
	{data = textImgNew(), itemname = 'audioducking', displayname = motif.option_info.menu_itemname_audio_audioducking, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.AudioDucking, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled)},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
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
		--Master Volume
		elseif t[item].itemname == 'mastervolume' then
			if commandGetState(main.p1Cmd, 'r') and config.MasterVolume < 200 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.MasterVolume = config.MasterVolume + 1
				t[item].vardisplay = config.MasterVolume .. '%'
				setMasterVolume(config.MasterVolume)
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.MasterVolume > 0 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.MasterVolume = config.MasterVolume - 1
				t[item].vardisplay = config.MasterVolume  .. '%'
				setMasterVolume(config.MasterVolume)
				modified = 1
			end
		--BGM Volume
		elseif t[item].itemname == 'bgmvolume' then
			if commandGetState(main.p1Cmd, 'r') and config.BgmVolume < 100 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.BgmVolume = config.BgmVolume + 1
				t[item].vardisplay = config.BgmVolume .. '%'
				setBgmVolume(config.BgmVolume)
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.BgmVolume > 0 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.BgmVolume = config.BgmVolume - 1
				t[item].vardisplay = config.BgmVolume .. '%'
				setBgmVolume(config.BgmVolume)
				modified = 1
			end
		--SFX Volume
		elseif t[item].itemname == 'sfxvolume' then
			if commandGetState(main.p1Cmd, 'r') and config.WavVolume < 100 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.WavVolume = config.WavVolume + 1
				t[item].vardisplay = config.WavVolume .. '%'
				setWavVolume(config.WavVolume)
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.WavVolume > 0 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.WavVolume = config.WavVolume - 1
				t[item].vardisplay = config.WavVolume .. '%'
				setWavVolume(config.WavVolume)
				modified = 1
			end
		--Audio Ducking
		elseif t[item].itemname == 'audioducking' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AudioDucking then
				config.AudioDucking = false
			else
				config.AudioDucking = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.AudioDucking, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled)
			setAudioDucking(config.AudioDucking)
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
	{data = textImgNew(), itemname = 'team1vs2life', displayname = motif.option_info.menu_itemname_gameplay_team1vs2life, vardata = textImgNew(), vardisplay = config.Team1VS2Life .. '%'},
	{data = textImgNew(), itemname = 'turnsrecoverybase', displayname = motif.option_info.menu_itemname_gameplay_turnsrecoverybase, vardata = textImgNew(), vardisplay = config.TurnsRecoveryBase .. '%'},
	{data = textImgNew(), itemname = 'turnsrecoverybonus', displayname = motif.option_info.menu_itemname_gameplay_turnsrecoverybonus, vardata = textImgNew(), vardisplay = config.TurnsRecoveryBonus .. '%'},
	{data = textImgNew(), itemname = 'teampowershare', displayname = motif.option_info.menu_itemname_gameplay_teampowershare, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.TeamPowerShare)},
	{data = textImgNew(), itemname = 'teamlifeshare', displayname = motif.option_info.menu_itemname_gameplay_teamlifeshare, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.TeamLifeShare)},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'ratioSettings', displayname = motif.option_info.menu_itemname_gameplay_ratio},
	{data = textImgNew(), itemname = 'advancedGameplaySettings', displayname = motif.option_info.menu_itemname_gameplay_advanced},
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
		--1P Vs Team Life
		elseif t[item].itemname == 'team1vs2life' then
			if commandGetState(main.p1Cmd, 'r') and config.Team1VS2Life < 300 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.Team1VS2Life = config.Team1VS2Life + 10
				t[item].vardisplay = config.Team1VS2Life .. '%'
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.Team1VS2Life > 10 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.Team1VS2Life = config.Team1VS2Life - 10
				t[item].vardisplay = config.Team1VS2Life .. '%'
				modified = 1
			end
		--Turns Recovery Base
		elseif t[item].itemname == 'turnsrecoverybase' then
			if commandGetState(main.p1Cmd, 'r') and config.TurnsRecoveryBase < 100 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.TurnsRecoveryBase = config.TurnsRecoveryBase + 0.5
				t[item].vardisplay = config.TurnsRecoveryBase .. '%'
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.TurnsRecoveryBase > 0 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.TurnsRecoveryBase = config.TurnsRecoveryBase - 0.5
				t[item].vardisplay = config.TurnsRecoveryBase .. '%'
				modified = 1
			end
		--Turns Recovery Bonus
		elseif t[item].itemname == 'turnsrecoverybonus' then
			if commandGetState(main.p1Cmd, 'r') and config.TurnsRecoveryBonus < 100 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.TurnsRecoveryBonus = config.TurnsRecoveryBonus + 0.5
				t[item].vardisplay = config.TurnsRecoveryBonus .. '%'
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config.TurnsRecoveryBonus > 0 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.TurnsRecoveryBonus = config.TurnsRecoveryBonus - 0.5
				t[item].vardisplay = config.TurnsRecoveryBonus .. '%'
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
		--Ratio Settings
		elseif t[item].itemname == 'ratioSettings' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
			options.f_ratioCfg()
		--Advanced Settings
		elseif t[item].itemname == 'advancedGameplaySettings' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
			options.f_advGameplayCfg()
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
--; RATIO SETTINGS
--;===========================================================
function options.f_displayRatio(value)
	local ret = options.f_precision((value - 1) * 100, '%.01f')
	if ret >= 0 then
		return '+' .. ret .. '%'
	end
	return ret .. '%'
end

options.t_ratioCfg = {
	{data = textImgNew(), itemname = 'ratio1Life', displayname = motif.option_info.menu_itemname_gameplay_ratio1life, vardata = textImgNew(), vardisplay = options.f_displayRatio(config.LifeRatio[1])},
	{data = textImgNew(), itemname = 'ratio1Attack', displayname = motif.option_info.menu_itemname_gameplay_ratio1attack, vardata = textImgNew(), vardisplay = options.f_displayRatio(config.AttackRatio[1])},
	{data = textImgNew(), itemname = 'ratio2Life', displayname = motif.option_info.menu_itemname_gameplay_ratio2life, vardata = textImgNew(), vardisplay = options.f_displayRatio(config.LifeRatio[2])},
	{data = textImgNew(), itemname = 'ratio2Attack', displayname = motif.option_info.menu_itemname_gameplay_ratio2attack, vardata = textImgNew(), vardisplay = options.f_displayRatio(config.AttackRatio[2])},
	{data = textImgNew(), itemname = 'ratio3Life', displayname = motif.option_info.menu_itemname_gameplay_ratio3life, vardata = textImgNew(), vardisplay = options.f_displayRatio(config.LifeRatio[3])},
	{data = textImgNew(), itemname = 'ratio3Attack', displayname = motif.option_info.menu_itemname_gameplay_ratio3attack, vardata = textImgNew(), vardisplay = options.f_displayRatio(config.AttackRatio[3])},
	{data = textImgNew(), itemname = 'ratio4Life', displayname = motif.option_info.menu_itemname_gameplay_ratio4life, vardata = textImgNew(), vardisplay = options.f_displayRatio(config.LifeRatio[4])},
	{data = textImgNew(), itemname = 'ratio4Attack', displayname = motif.option_info.menu_itemname_gameplay_ratio4attack, vardata = textImgNew(), vardisplay = options.f_displayRatio(config.AttackRatio[4])},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_gameplay_back},
}
options.t_ratioCfg = main.f_cleanTable(options.t_ratioCfg, main.t_sort.option_info)

function options.f_ratioCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = options.t_ratioCfg
	textImgSetText(txt_title, motif.option_info.title_text_ratio)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_gameplay)
			break
		--Back
		elseif t[item].itemname == 'back' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_gameplay)
			break
		--Ratio 1-4 Life / Damage
		else
			local ratioLevel, ratioType = t[item].itemname:match('^ratio([1-4])(.+)$')
			ratioLevel = tonumber(ratioLevel)
			if commandGetState(main.p1Cmd, 'r') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config[ratioType .. 'Ratio'][ratioLevel] = options.f_precision(config[ratioType .. 'Ratio'][ratioLevel] + 0.01, '%.02f')
				t[item].vardisplay = options.f_displayRatio(config[ratioType .. 'Ratio'][ratioLevel])
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config[ratioType .. 'Ratio'][ratioLevel] > 0.01 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config[ratioType .. 'Ratio'][ratioLevel] = options.f_precision(config[ratioType .. 'Ratio'][ratioLevel] - 0.01, '%.02f')
				t[item].vardisplay = options.f_displayRatio(config[ratioType .. 'Ratio'][ratioLevel])
				modified = 1
			end
		end
		options.f_menuCommonDraw(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; ADVANCED GAMEPLAY SETTINGS
--;===========================================================
function options.f_checkTeamAmount(arg1, arg2, arg3)
	local ret = arg1
	if arg1 <= arg2 then
		ret = arg3
	end
	return ret
end

options.t_advGameplayCfg = {
	{data = textImgNew(), itemname = 'attackpowermul', displayname = motif.option_info.menu_itemname_gameplay_attackpowermul, vardata = textImgNew(), vardisplay = config['Attack.LifeToPowerMul']},
	{data = textImgNew(), itemname = 'gethitpowermul', displayname = motif.option_info.menu_itemname_gameplay_gethitpowermul, vardata = textImgNew(), vardisplay = config['GetHit.LifeToPowerMul']},
	{data = textImgNew(), itemname = 'superdefencemul', displayname = motif.option_info.menu_itemname_gameplay_superdefencemul, vardata = textImgNew(), vardisplay = config['Super.TargetDefenceMul']},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'singlemode', displayname = motif.option_info.menu_itemname_gameplay_singlemode, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.SingleTeamMode, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled)},
	{data = textImgNew(), itemname = 'numturns', displayname = motif.option_info.menu_itemname_gameplay_numturns, vardata = textImgNew(), vardisplay = options.f_checkTeamAmount(config.NumTurns, 1, motif.option_info.menu_itemname_disabled)},
	{data = textImgNew(), itemname = 'numsimul', displayname = motif.option_info.menu_itemname_gameplay_numsimul, vardata = textImgNew(), vardisplay = options.f_checkTeamAmount(config.NumSimul, 1, motif.option_info.menu_itemname_disabled)},
	{data = textImgNew(), itemname = 'numtag', displayname = motif.option_info.menu_itemname_gameplay_numtag, vardata = textImgNew(), vardisplay = options.f_checkTeamAmount(config.NumTag, 1, motif.option_info.menu_itemname_disabled)},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_gameplay_back},
}
options.t_advGameplayCfg = main.f_cleanTable(options.t_advGameplayCfg, main.t_sort.option_info)

function options.f_advGameplayCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = options.t_advGameplayCfg
	textImgSetText(txt_title, motif.option_info.title_text_advgameplay)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_gameplay)
			break
		--Attack.LifeToPowerMul
		elseif t[item].itemname == 'attackpowermul' then
			if commandGetState(main.p1Cmd, 'r') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['Attack.LifeToPowerMul'] = options.f_precision(config['Attack.LifeToPowerMul'] + 0.1, '%.01f')
				t[item].vardisplay = config['Attack.LifeToPowerMul']
				setAttackLifeToPowerMul(config['Attack.LifeToPowerMul'])
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config['Attack.LifeToPowerMul'] > 0.1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['Attack.LifeToPowerMul'] = options.f_precision(config['Attack.LifeToPowerMul'] - 0.1, '%.01f')
				t[item].vardisplay = config['Attack.LifeToPowerMul']
				setAttackLifeToPowerMul(config['Attack.LifeToPowerMul'])
				modified = 1
			end
		--GetHit.LifeToPowerMul
		elseif t[item].itemname == 'gethitpowermul' then
			if commandGetState(main.p1Cmd, 'r') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['GetHit.LifeToPowerMul'] = options.f_precision(config['GetHit.LifeToPowerMul'] + 0.1, '%.01f')
				t[item].vardisplay = config['GetHit.LifeToPowerMul']
				setGetHitLifeToPowerMul(config['GetHit.LifeToPowerMul'])
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config['GetHit.LifeToPowerMul'] > 0.1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['GetHit.LifeToPowerMul'] = options.f_precision(config['GetHit.LifeToPowerMul'] - 0.1, '%.01f')
				t[item].vardisplay = config['GetHit.LifeToPowerMul']
				setGetHitLifeToPowerMul(config['GetHit.LifeToPowerMul'])
				modified = 1
			end
		--Super.TargetDefenceMul
		elseif t[item].itemname == 'superdefencemul' then
			if commandGetState(main.p1Cmd, 'r') then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['Super.TargetDefenceMul'] = options.f_precision(config['Super.TargetDefenceMul'] + 0.1, '%.01f')
				t[item].vardisplay = config['Super.TargetDefenceMul']
				setSuperTargetDefenceMul(config['Super.TargetDefenceMul'])
				modified = 1
			elseif commandGetState(main.p1Cmd, 'l') and config['Super.TargetDefenceMul'] > 0.1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['Super.TargetDefenceMul'] = options.f_precision(config['Super.TargetDefenceMul'] - 0.1, '%.01f')
				t[item].vardisplay = config['Super.TargetDefenceMul']
				setSuperTargetDefenceMul(config['Super.TargetDefenceMul'])
				modified = 1
			end
		--Single Mode
		elseif t[item].itemname == 'singlemode' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.SingleTeamMode then
				config.SingleTeamMode = false
			else
				config.SingleTeamMode = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.SingleTeamMode, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled)
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
				t[item].vardisplay = options.f_checkTeamAmount(config.NumTurns, 1, motif.option_info.menu_itemname_disabled)
				modified = 1
			end
		--Simul Limit
		elseif t[item].itemname == 'numsimul' then
			if commandGetState(main.p1Cmd, 'r') and config.NumSimul < 8 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.NumSimul = config.NumSimul + 1
				t[item].vardisplay = config.NumSimul
				modified = 1
				needReload = 1 --TODO: won't be needed if we add a function that can extend sys.keyConfig and sys.JoystickConfig from lua
			elseif commandGetState(main.p1Cmd, 'l') and config.NumSimul > 1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.NumSimul = config.NumSimul - 1
				t[item].vardisplay = options.f_checkTeamAmount(config.NumSimul, 1, motif.option_info.menu_itemname_disabled)
				modified = 1
			end
		--Tag Limit
		elseif t[item].itemname == 'numtag' then
			if commandGetState(main.p1Cmd, 'r') and config.NumTag < 4 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.NumTag = config.NumTag + 1
				t[item].vardisplay = config.NumTag
				modified = 1
				needReload = 1 --TODO: won't be needed if we add a function that can extend sys.keyConfig and sys.JoystickConfig from lua
			elseif commandGetState(main.p1Cmd, 'l') and config.NumTag > 1 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config.NumTag = config.NumTag - 1
				t[item].vardisplay = options.f_checkTeamAmount(config.NumTag, 1, motif.option_info.menu_itemname_disabled)
				modified = 1
			end
		--Back
		elseif t[item].itemname == 'back' and main.f_btnPalNo(main.p1Cmd) > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_gameplay)
			break
		end
		options.f_menuCommonDraw(cursorPosY, moveTxt, item, t)
	end
end

--;===========================================================
--; ENGINE SETTINGS
--;===========================================================
local t_quicklaunchNames = {}
t_quicklaunchNames[0] = "Disabled"
t_quicklaunchNames[1] = "Level1"
t_quicklaunchNames[2] = "Level2"

options.t_engineCfg = {
	{data = textImgNew(), itemname = 'allowdebugkeys', displayname = motif.option_info.menu_itemname_engine_allowdebugkeys, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.AllowDebugKeys, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled)},
	{data = textImgNew(), itemname = 'simulmode', displayname = motif.option_info.menu_itemname_engine_simulmode, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.SimulMode, motif.option_info.menu_itemname_disabled, motif.option_info.menu_itemname_enabled)},
	{data = textImgNew(), itemname = 'quicklaunch', displayname = motif.option_info.menu_itemname_engine_quicklaunch, vardata = textImgNew(), vardisplay = t_quicklaunchNames[config.QuickLaunch]},
	{data = textImgNew(), itemname = 'lifebarfontscale', displayname = motif.option_info.menu_itemname_engine_lifebarfontscale, vardata = textImgNew(), vardisplay = config.LifebarFontScale},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'helpermax', displayname = motif.option_info.menu_itemname_engine_helpermax, vardata = textImgNew(), vardisplay = config.HelperMax},
	{data = textImgNew(), itemname = 'playerprojectilemax', displayname = motif.option_info.menu_itemname_engine_playerprojectilemax, vardata = textImgNew(), vardisplay = config.PlayerProjectileMax},
	{data = textImgNew(), itemname = 'explodmax', displayname = motif.option_info.menu_itemname_engine_explodmax, vardata = textImgNew(), vardisplay = config.ExplodMax},
	{data = textImgNew(), itemname = 'afterimagemax', displayname = motif.option_info.menu_itemname_engine_afterimagemax, vardata = textImgNew(), vardisplay = config.AfterImageMax},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'zoomactive', displayname = motif.option_info.menu_itemname_engine_zoomactive, vardata = textImgNew(), vardisplay = options.f_boolDisplay(config.ZoomActive)},
	{data = textImgNew(), itemname = 'maxzoomout', displayname = motif.option_info.menu_itemname_engine_maxzoomout, vardata = textImgNew(), vardisplay = config.ZoomMin},
	{data = textImgNew(), itemname = 'maxzoomin', displayname = motif.option_info.menu_itemname_engine_maxzoomin, vardata = textImgNew(), vardisplay = config.ZoomMax},
	{data = textImgNew(), itemname = 'zoomspeed', displayname = motif.option_info.menu_itemname_engine_zoomspeed, vardata = textImgNew(), vardisplay = config.ZoomSpeed},
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_video_back},
}
options.t_engineCfg = main.f_cleanTable(options.t_engineCfg, main.t_sort.option_info)

function options.f_engineCfg()
	main.f_cmdInput()
	local cursorPosY = 1
	local moveTxt = 0
	local item = 1
	local t = options.t_engineCfg
	textImgSetText(txt_title, motif.option_info.title_text_engine)
	while true do
		cursorPosY, moveTxt, item = options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
		if esc() then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			textImgSetText(txt_title, motif.option_info.title_text_main)
			break
		--Allow Debug Keys
		elseif t[item].itemname == 'allowdebugkeys' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AllowDebugKeys then
				config.AllowDebugKeys = false
			else
				config.AllowDebugKeys = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.AllowDebugKeys, motif.option_info.menu_itemname_enabled, motif.option_info.menu_itemname_disabled)
			setAllowDebugKeys(config.AllowDebugKeys)
			modified = 1
		--Legacy Tag Mode
		elseif t[item].itemname == 'simulmode' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.SimulMode then
				config.SimulMode = false
			else
				config.SimulMode = true
			end
			t[item].vardisplay = options.f_boolDisplay(config.SimulMode, motif.option_info.menu_itemname_disabled, motif.option_info.menu_itemname_enabled)
			main.f_warning(main.f_extractText(motif.warning_info.text_simul), motif.option_info, motif.optionbgdef)
			modified = 1
		-- Quick Launch
		elseif t[item].itemname == 'quicklaunch' and (commandGetState(main.p1Cmd, 'r') or commandGetState(main.p1Cmd, 'l') or main.f_btnPalNo(main.p1Cmd) > 0) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if commandGetState(main.p1Cmd, 'r') and config.QuickLaunch < #t_quicklaunchNames then
				config.QuickLaunch = config.QuickLaunch + 1
			elseif commandGetState(main.p1Cmd, 'l') and config.QuickLaunch > 0 then
				config.QuickLaunch = config.QuickLaunch - 1
			end
			t[item].vardisplay = t_quicklaunchNames[config.QuickLaunch]
			modified = 1
		--Lifebar Font Scale
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
	{data = textImgNew(), itemname = 'empty', displayname = ' '},
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
			--Key Config
			if t[item].itemname == 'keyboard' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.f_keyCfg('KeyConfig', t[item].itemname)
			--Joystick Config
			elseif t[item].itemname == 'gamepad' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				options.f_keyCfg('JoystickConfig', t[item].itemname)
			--System Keys (not implemented yet)
			elseif t[item].itemname == 'system' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			--Default Values
			elseif t[item].itemname == 'defaultvalues' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				options.f_keyDefault()
				modified = 1
				needReload = 1 --TODO: won't be needed if we add a function that can edit sys.keyConfig and sys.JoystickConfig from lua
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
	{data = textImgNew(), itemname = 'd', displayname = motif.option_info.menu_itemname_key_d, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'w', displayname = motif.option_info.menu_itemname_key_w, vardata = textImgNew()},
	{data = textImgNew(), itemname = 'back', displayname = motif.option_info.menu_itemname_key_back, infodata = textImgNew(), infodisplay = motif.option_info.menu_itemname_info_esc},
}
--t_keyCfg = main.f_cleanTable(t_keyCfg, main.t_sort.option_info)

local txt_keyController = textImgNew()
function options.f_keyCfg(cfgType, controller)
	main.f_cmdInput()
	local cursorPosY = 2
	local moveTxt = 0
	local item = 2
	local item_start = 2
	local t = t_keyCfg
	local t_pos = {motif.option_info.menu_key_p1_pos, motif.option_info.menu_key_p2_pos}
	local configall = false
	local key = ''
	local t_keyList = {}
	local t_conflict = {}
	local btnReleased = 0
	local player = 1
	local btn = tostring(config[cfgType][player].Buttons[item - item_start])
	local joyNum = 0
	textImgSetText(txt_title, motif.option_info.title_text_key)
	--count all button assignments on the same controller
	for i = 1, #config[cfgType] do
		joyNum = config[cfgType][i].Joystick
		if t_keyList[joyNum] == nil then
			t_keyList[joyNum] = {} --creates subtable for each controller (1 for keyboard or at least 2 for gamepads)
			t_conflict[joyNum] = false --set default conflict flag for each controller
		end
		for k, v in pairs(config[cfgType][i].Buttons) do
			v = tostring(v)
			t_keyCfg[k + item_start]['vardisplay' .. i] = v --assign vardisplay entry (assigned button name) in t_keyCfg table
			if v ~= tostring(motif.option_info.menu_itemname_info_disable) then --if button is not disabled
				if t_keyList[joyNum][v] == nil then
					t_keyList[joyNum][v] = 1
				else
					t_keyList[joyNum][v] = t_keyList[joyNum][v] + 1
				end
			end
		end
	end
	joyNum = config[cfgType][player].Joystick
	while true do
		--Config all
		if configall then
			if cfgType == 'KeyConfig' then --detect keyboard key
				key = getKey()
			elseif getJoystickPresent(joyNum) == false then --ensure that gamepad is connected
				main.f_warning(main.f_extractText(motif.warning_info.text_pad), motif.option_info, motif.optionbgdef)
				configall = false
				commandBufReset(main.p1Cmd)
			else --detect gamepad key
				local tmp = getKey()
				if tonumber(tmp) == nil then --button released
					btnReleased = 1
				elseif btnReleased == 1 then --button pressed after releasing button once
					key = tmp
					btnReleased = 0
				end
			end
			key = tostring(key)
			if esc() then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				configall = false
				commandBufReset(main.p1Cmd)
			--some key detected
			elseif key ~= '' then
				--spacebar (disable key)
				if key == 'SPACE' then
					sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
					--decrease old button count
					if t_keyList[joyNum][btn] ~= nil and t_keyList[joyNum][btn] > 1 then
						t_keyList[joyNum][btn] = t_keyList[joyNum][btn] - 1
					else
						t_keyList[joyNum][btn] = nil
					end
					--update vardisplay / config data
					t[item]['vardisplay' .. player] = motif.option_info.menu_itemname_info_disable
					config[cfgType][player].Buttons[item - item_start] = tostring(motif.option_info.menu_itemname_info_disable)
					modified = 1
					needReload = 1 --TODO: won't be needed if we add a function that can edit sys.keyConfig and sys.JoystickConfig from lua
				--other keyboard or gamepad key
				elseif cfgType == 'KeyConfig' or (cfgType == 'JoystickConfig' and tonumber(key) ~= nil) then
					sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
					--decrease old button count
					if t_keyList[joyNum][btn] ~= nil and t_keyList[joyNum][btn] > 1 then
						t_keyList[joyNum][btn] = t_keyList[joyNum][btn] - 1
					else
						t_keyList[joyNum][btn] = nil
					end
					--increase new button count
					if t_keyList[joyNum][key] == nil then
						t_keyList[joyNum][key] = 1
					else
						t_keyList[joyNum][key] = t_keyList[joyNum][key] + 1
					end
					--update vardisplay / config data
					t[item]['vardisplay' .. player] = key
					config[cfgType][player].Buttons[item - item_start] = tostring(key)
					modified = 1
					needReload = 1 --TODO: won't be needed if we add a function that can edit sys.keyConfig and sys.JoystickConfig from lua
				--non gamepad key on gamepad controller
				else
					sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				end
				--move to the next position
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
		--move up / down / left / right
		elseif commandGetState(main.p1Cmd, 'u') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			item = item - 1
		elseif commandGetState(main.p1Cmd, 'd') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			item = item + 1
		elseif commandGetState(main.p1Cmd, 'l') or commandGetState(main.p1Cmd, 'r') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if player == 1 then
				player = 2
			else
				player = 1
			end
			joyNum = config[cfgType][player].Joystick
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
		btn = tostring(config[cfgType][player].Buttons[item - item_start])
		if configall == false then
			if esc() and not t_conflict[joyNum] then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				textImgSetText(txt_title, motif.option_info.title_text_input)
				break
			--Config all
			elseif (t[item].itemname == 'configall' and main.f_btnPalNo(main.p1Cmd) > 0) or getKey() == 'F1' or getKey() == 'F2' then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				if getKey() == 'F1' then
					player = 1
				elseif getKey() == 'F2' then
					player = 2
				end
				if cfgType == 'JoystickConfig' and getJoystickPresent(joyNum) == false then
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
			--Back
			elseif (t[item].itemname == 'back' and main.f_btnPalNo(main.p1Cmd) > 0) then
				if t_conflict[joyNum] then
					main.f_warning(main.f_extractText(motif.warning_info.text_keys), motif.option_info, motif.optionbgdef)
				else
					sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
					textImgSetText(txt_title, motif.option_info.title_text_input)
					break
				end
			--individual buttons
			elseif main.f_btnPalNo(main.p1Cmd) > 0 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				if cfgType == 'JoystickConfig' and getJoystickPresent(joyNum) == false then
					main.f_warning(main.f_extractText(motif.warning_info.text_pad), motif.option_info, motif.optionbgdef)
				else
					key = main.f_input(main.f_extractText(motif.option_info.input_text_key), motif.option_info, motif.optionbgdef, controller, joyNum, 'SPACE')
					--spacebar (disable key)
					if key == 'SPACE' then
						sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
						--decrease old button count
						if t_keyList[joyNum][btn] ~= nil and t_keyList[joyNum][btn] > 1 then
							t_keyList[joyNum][btn] = t_keyList[joyNum][btn] - 1
						else
							t_keyList[joyNum][btn] = nil
						end
						--update vardisplay / config data
						t[item]['vardisplay' .. player] = motif.option_info.menu_itemname_info_disable
						config[cfgType][player].Buttons[item - item_start] = motif.option_info.menu_itemname_info_disable
						modified = 1
						needReload = 1 --TODO: won't be needed if we add a function that can edit sys.keyConfig and sys.JoystickConfig from lua
					--other keyboard or gamepad key
					elseif (cfgType == 'KeyConfig' and key ~= '') or (cfgType == 'JoystickConfig' and tonumber(key) ~= nil) then
						sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
						--decrease old button count
						if t_keyList[joyNum][btn] ~= nil and t_keyList[joyNum][btn] > 1 then
							t_keyList[joyNum][btn] = t_keyList[joyNum][btn] - 1
						else
							t_keyList[joyNum][btn] = nil
						end
						--increase new button count
						if t_keyList[joyNum][key] == nil then
							t_keyList[joyNum][key] = 1
						else
							t_keyList[joyNum][key] = t_keyList[joyNum][key] + 1
						end
						--update vardisplay / config data
						t[item]['vardisplay' .. player] = key
						config[cfgType][player].Buttons[item - item_start] = tostring(key)
						modified = 1
						needReload = 1 --TODO: won't be needed if we add a function that can edit sys.keyConfig and sys.JoystickConfig from lua
					else
						sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
					end
					resetKey()
					key = ''
				end
			end
		end
		t_conflict[joyNum] = false
		--draw clearcolor
		clearColor(motif.optionbgdef.bgclearcolor[1], motif.optionbgdef.bgclearcolor[2], motif.optionbgdef.bgclearcolor[3])
		--draw layerno = 0 backgrounds
		bgDraw(motif.optionbgdef.bg, false)
		--draw player num
		for i = 1, 2 do
			textImgDraw(main.f_updateTextImg(
				txt_keyController,
				motif.font_data[motif.option_info['menu_item_key_p' .. i .. '_font'][1]],
				motif.option_info['menu_item_key_p' .. i .. '_font'][2],
				motif.option_info['menu_item_key_p' .. i .. '_font'][3],
				motif.option_info['menu_itemname_key_p' .. i],
				motif.option_info['menu_item_p' .. i .. '_pos'][1],
				motif.option_info['menu_item_p' .. i .. '_pos'][2],
				motif.option_info['menu_item_key_p' .. i .. '_font_scale'][1],
				motif.option_info['menu_item_key_p' .. i .. '_font_scale'][2],
				motif.option_info['menu_item_key_p' .. i .. '_font'][4],
				motif.option_info['menu_item_key_p' .. i .. '_font'][5],
				motif.option_info['menu_item_key_p' .. i .. '_font'][6],
				motif.option_info['menu_item_key_p' .. i .. '_font'][7],
				motif.option_info['menu_item_key_p' .. i .. '_font'][8],
				motif.defaultOptions
			))
		end
		--draw menu box
		if motif.option_info.menu_boxbg_visible == 1 then
			local coord4 = 0
			for i = 1, 2 do
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
					motif.option_info.menu_boxbg_alpha[2],
					motif.defaultOptions
				)
			end
		end
		--draw title
		textImgDraw(txt_title)
		--draw menu items
		for i = 1, #t do
			for j = 1, 2 do
				if i > item - cursorPosY then
					if t[i].itemname == 'configall' then
						if j == 1 then --player1 side (left)
							t[i].infodisplay = motif.option_info.menu_itemname_info_f1
						else --player2 side (right)
							t[i].infodisplay = motif.option_info.menu_itemname_info_f2
						end
					end
					if i == item and j == player then --active item
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
							motif.option_info.menu_item_active_font[8],
							motif.defaultOptions
						))
						--draw vardata
						if t[i].vardata ~= nil then
							if t_keyList[joyNum][tostring(t[i]['vardisplay' .. j])] ~= nil and t_keyList[joyNum][tostring(t[i]['vardisplay' .. j])] > 1 then
								textImgDraw(main.f_updateTextImg(
									t[i].vardata,
									motif.font_data[motif.option_info.menu_item_value_conflict_font[1]],
									motif.option_info.menu_item_value_conflict_font[2],
									motif.option_info.menu_item_value_conflict_font[3],
									t[i]['vardisplay' .. j],
									t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
									t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
									motif.option_info.menu_item_value_conflict_font_scale[1],
									motif.option_info.menu_item_value_conflict_font_scale[2],
									motif.option_info.menu_item_value_conflict_font[4],
									motif.option_info.menu_item_value_conflict_font[5],
									motif.option_info.menu_item_value_conflict_font[6],
									motif.option_info.menu_item_value_conflict_font[7],
									motif.option_info.menu_item_value_conflict_font[8],
									motif.defaultOptions
								))
								t_conflict[joyNum] = true
							else
								textImgDraw(main.f_updateTextImg(
									t[i].vardata,
									motif.font_data[motif.option_info.menu_item_value_active_font[1]],
									motif.option_info.menu_item_value_active_font[2],
									motif.option_info.menu_item_value_active_font[3],
									t[i]['vardisplay' .. j],
									t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
									t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
									motif.option_info.menu_item_value_active_font_scale[1],
									motif.option_info.menu_item_value_active_font_scale[2],
									motif.option_info.menu_item_value_active_font[4],
									motif.option_info.menu_item_value_active_font[5],
									motif.option_info.menu_item_value_active_font[6],
									motif.option_info.menu_item_value_active_font[7],
									motif.option_info.menu_item_value_active_font[8],
									motif.defaultOptions
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
								motif.option_info.menu_item_info_active_font[8],
								motif.defaultOptions
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
							motif.option_info.menu_item_font[8],
							motif.defaultOptions
						))
						--draw vardata
						if t[i].vardata ~= nil then
							if t_keyList[joyNum][tostring(t[i]['vardisplay' .. j])] ~= nil and t_keyList[joyNum][tostring(t[i]['vardisplay' .. j])] > 1 then
								textImgDraw(main.f_updateTextImg(
									t[i].vardata,
									motif.font_data[motif.option_info.menu_item_value_conflict_font[1]],
									motif.option_info.menu_item_value_conflict_font[2],
									motif.option_info.menu_item_value_conflict_font[3],
									t[i]['vardisplay' .. j],
									t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
									t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
									motif.option_info.menu_item_value_conflict_font_scale[1],
									motif.option_info.menu_item_value_conflict_font_scale[2],
									motif.option_info.menu_item_value_conflict_font[4],
									motif.option_info.menu_item_value_conflict_font[5],
									motif.option_info.menu_item_value_conflict_font[6],
									motif.option_info.menu_item_value_conflict_font[7],
									motif.option_info.menu_item_value_conflict_font[8],
									motif.defaultOptions
								))
								t_conflict[joyNum] = true
							else
								textImgDraw(main.f_updateTextImg(
									t[i].vardata,
									motif.font_data[motif.option_info.menu_item_value_font[1]],
									motif.option_info.menu_item_value_font[2],
									motif.option_info.menu_item_value_font[3],
									t[i]['vardisplay' .. j],
									t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
									t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
									motif.option_info.menu_item_value_font_scale[1],
									motif.option_info.menu_item_value_font_scale[2],
									motif.option_info.menu_item_value_font[4],
									motif.option_info.menu_item_value_font[5],
									motif.option_info.menu_item_value_font[6],
									motif.option_info.menu_item_value_font[7],
									motif.option_info.menu_item_value_font[8],
									motif.defaultOptions
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
								motif.option_info.menu_item_info_font[8],
								motif.defaultOptions
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
			for i = 1, 2 do
				if i == player then
					fillRect(
						t_pos[i][1] + motif.option_info.menu_key_boxcursor_coords[1],
						t_pos[i][2] + motif.option_info.menu_key_boxcursor_coords[2] + (cursorPosY - 1) * motif.option_info.menu_key_item_spacing[2],
						motif.option_info.menu_key_boxcursor_coords[3] - motif.option_info.menu_key_boxcursor_coords[1] + 1,
						motif.option_info.menu_key_boxcursor_coords[4] - motif.option_info.menu_key_boxcursor_coords[2] + 1 + main.f_oddRounding(motif.option_info.menu_key_boxcursor_coords[2]),
						motif.option_info.menu_boxcursor_col[1],
						motif.option_info.menu_boxcursor_col[2],
						motif.option_info.menu_boxcursor_col[3],
						src,
						dst,
						motif.defaultOptions
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

return options

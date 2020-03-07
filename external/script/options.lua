
local options = {}

--;===========================================================
--; COMMON
--;===========================================================
local modified = 0
local needReload = 0

main.framesPerCount = getFramesPerCount()
if config.RoundsNumSingle == -1 then
	main.roundsNumSingle = getMatchWins()
else
	main.roundsNumSingle = config.RoundsNumSingle
end
if config.RoundsNumTeam == -1 then
	main.roundsNumTeam = getMatchWins()
else
	main.roundsNumTeam = config.RoundsNumTeam
end
if config.MaxDrawGames == -2 then
	main.maxDrawGames = getMatchMaxDrawGames()
else
	main.maxDrawGames = config.MaxDrawGames
end

--return string depending on bool
function options.f_boolDisplay(bool, t, f)
	t = t or motif.option_info.menu_valuename_yes
	f = f or motif.option_info.menu_valuename_no
	if bool == true then
		return t
	else
		return f
	end
end

--return table entry (or ret if specified) if provided key exists in the table, otherwise return default argument
function options.f_definedDisplay(key, t, default, ret)
	if key ~= nil and t[key] ~= nil then
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
				config.KeyConfig[i].Buttons[j] = tostring(motif.option_info.menu_valuename_nokey)
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

function options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
	if main.input({1, 2}, {'$U'}) then
		sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
		item = item - 1
		if t[item] ~= nil and t[item].itemname == 'empty' then
			item = item - 1
		end
	elseif main.input({1, 2}, {'$D'}) then
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
	elseif main.input({1, 2}, {'$U'}) and cursorPosY > 1 then
		cursorPosY = cursorPosY - 1
		if t[cursorPosY] ~= nil and t[cursorPosY].itemname == 'empty' then
			cursorPosY = cursorPosY - 1
		end
	elseif main.input({1, 2}, {'$D'}) and cursorPosY < motif.option_info.menu_window_visibleitems then
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

local txt_title = text:create({
	font =   motif.font_data[motif.option_info.title_font[1]],
	bank =   motif.option_info.title_font[2],
	align =  motif.option_info.title_font[3],
	text =   '',
	x =      motif.option_info.title_offset[1],
	y =      motif.option_info.title_offset[2],
	scaleX = motif.option_info.title_font_scale[1],
	scaleY = motif.option_info.title_font_scale[2],
	r =      motif.option_info.title_font[4],
	g =      motif.option_info.title_font[5],
	b =      motif.option_info.title_font[6],
	src =    motif.option_info.title_font[7],
	dst =    motif.option_info.title_font[8],
	--defsc =  motif.defaultOptions --title font assignment exists in mugen
})
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
	txt_title:draw()
	--draw menu items
	for i = 1, #t do
		if i > item - cursorPosY then
			if i == item then
				if t[i].selected then
					t[i].data:update({
						font =   motif.font_data[motif.option_info.menu_item_selected_active_font[1]],
						bank =   motif.option_info.menu_item_selected_active_font[2],
						align =  motif.option_info.menu_item_selected_active_font[3],
						text =   t[i].displayname,
						x =      motif.option_info.menu_pos[1],
						y =      motif.option_info.menu_pos[2] + (i - 1) * motif.option_info.menu_item_spacing[2] - moveTxt,
						scaleX = motif.option_info.menu_item_selected_active_font_scale[1],
						scaleY = motif.option_info.menu_item_selected_active_font_scale[2],
						r =      motif.option_info.menu_item_selected_active_font[4],
						g =      motif.option_info.menu_item_selected_active_font[5],
						b =      motif.option_info.menu_item_selected_active_font[6],
						src =    motif.option_info.menu_item_selected_active_font[7],
						dst =    motif.option_info.menu_item_selected_active_font[8],
						defsc = motif.defaultOptions
					})
					t[i].data:draw()
				else
					t[i].data:update({
						font =   motif.font_data[motif.option_info.menu_item_active_font[1]],
						bank =   motif.option_info.menu_item_active_font[2],
						align =  motif.option_info.menu_item_active_font[3],
						text =   t[i].displayname,
						x =      motif.option_info.menu_pos[1],
						y =      motif.option_info.menu_pos[2] + (i - 1) * motif.option_info.menu_item_spacing[2] - moveTxt,
						scaleX = motif.option_info.menu_item_active_font_scale[1],
						scaleY = motif.option_info.menu_item_active_font_scale[2],
						r =      motif.option_info.menu_item_active_font[4],
						g =      motif.option_info.menu_item_active_font[5],
						b =      motif.option_info.menu_item_active_font[6],
						src =    motif.option_info.menu_item_active_font[7],
						dst =    motif.option_info.menu_item_active_font[8],
						defsc =  motif.defaultOptions
					})
					t[i].data:draw()
				end
				if t[i].vardata ~= nil then
					t[i].vardata:update({
						font =   motif.font_data[motif.option_info.menu_item_value_active_font[1]],
						bank =   motif.option_info.menu_item_value_active_font[2],
						align =  motif.option_info.menu_item_value_active_font[3],
						text =   t[i].vardisplay,
						x =      motif.option_info.menu_pos[1] + motif.option_info.menu_item_spacing[1],
						y =      motif.option_info.menu_pos[2] + (i - 1) * motif.option_info.menu_item_spacing[2] - moveTxt,
						scaleX = motif.option_info.menu_item_value_active_font_scale[1],
						scaleY = motif.option_info.menu_item_value_active_font_scale[2],
						r =      motif.option_info.menu_item_value_active_font[4],
						g =      motif.option_info.menu_item_value_active_font[5],
						b =      motif.option_info.menu_item_value_active_font[6],
						src =    motif.option_info.menu_item_value_active_font[7],
						dst =    motif.option_info.menu_item_value_active_font[8],
						defsc =  motif.defaultOptions
					})
					t[i].vardata:draw()
				end
			else
				if t[i].selected then
					t[i].data:update({
						font =   motif.font_data[motif.option_info.menu_item_selected_font[1]],
						bank =   motif.option_info.menu_item_selected_font[2],
						align =  motif.option_info.menu_item_selected_font[3],
						text =   t[i].displayname,
						x =      motif.option_info.menu_pos[1],
						y =      motif.option_info.menu_pos[2] + (i - 1) * motif.option_info.menu_item_spacing[2] - moveTxt,
						scaleX = motif.option_info.menu_item_selected_font_scale[1],
						scaleY = motif.option_info.menu_item_selected_font_scale[2],
						r =      motif.option_info.menu_item_selected_font[4],
						g =      motif.option_info.menu_item_selected_font[5],
						b =      motif.option_info.menu_item_selected_font[6],
						src =    motif.option_info.menu_item_selected_font[7],
						dst =    motif.option_info.menu_item_selected_font[8],
						defsc =  motif.defaultOptions
					})
					t[i].data:draw()
				else
					t[i].data:update({
						font =   motif.font_data[motif.option_info.menu_item_font[1]],
						bank =   motif.option_info.menu_item_font[2],
						align =  motif.option_info.menu_item_font[3],
						text =   t[i].displayname,
						x =      motif.option_info.menu_pos[1],
						y =      motif.option_info.menu_pos[2] + (i - 1) * motif.option_info.menu_item_spacing[2] - moveTxt,
						scaleX = motif.option_info.menu_item_font_scale[1],
						scaleY = motif.option_info.menu_item_font_scale[2],
						r =      motif.option_info.menu_item_font[4],
						g =      motif.option_info.menu_item_font[5],
						b =      motif.option_info.menu_item_font[6],
						src =    motif.option_info.menu_item_font[7],
						dst =    motif.option_info.menu_item_font[8],
						defsc =  motif.defaultOptions
					})
					t[i].data:draw()
				end
				if t[i].vardata ~= nil then
					t[i].vardata:update({
						font =   motif.font_data[motif.option_info.menu_item_value_font[1]],
						bank =   motif.option_info.menu_item_value_font[2],
						align =  motif.option_info.menu_item_value_font[3],
						text =   t[i].vardisplay,
						x =      motif.option_info.menu_pos[1] + motif.option_info.menu_item_spacing[1],
						y =      motif.option_info.menu_pos[2] + (i - 1) * motif.option_info.menu_item_spacing[2] - moveTxt,
						scaleX = motif.option_info.menu_item_value_font_scale[1],
						scaleY = motif.option_info.menu_item_value_font_scale[2],
						r =      motif.option_info.menu_item_value_font[4],
						g =      motif.option_info.menu_item_value_font[5],
						b =      motif.option_info.menu_item_value_font[6],
						src =    motif.option_info.menu_item_value_font[7],
						dst =    motif.option_info.menu_item_value_font[8],
						defsc =  motif.defaultOptions
					})
					t[i].vardata:draw()
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
		commandBufReset(main.cmd[1])
		commandBufReset(main.cmd[2])
	elseif fadeType == 'fadeout' then
		commandBufReset(main.cmd[1])
		commandBufReset(main.cmd[2])
		return --skip last frame rendering
	else
		main.f_cmdInput()
	end
	refresh()
end

--;===========================================================
--; LOOPS
--;===========================================================
function options.f_displayRatio(value)
	local ret = options.f_precision((value - 1) * 100, '%.01f')
	if ret >= 0 then
		return '+' .. ret .. '%'
	end
	return ret .. '%'
end

local t_quicklaunchNames = {
	[0] = motif.option_info.menu_valuename_disabled,
	[1] = motif.option_info.menu_valuename_level1,
	[2] = motif.option_info.menu_valuename_level2,
}

local function f_externalShaderName()
	if #config.ExternalShaders > 0 and config.PostProcessingShader ~= 0 then
		return config.ExternalShaders[1]:gsub('^.+/', '')
	end
	return motif.option_info.menu_valuename_disabled
end

options.t_itemname = {
	--Back
	['back'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.title_info.cancel_snd[2])
			return false
		end
		return true
	end,
	--Port Change
	['portchange'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			local port = main.f_input(main.f_extractText(motif.option_info.input_text_port), motif.option_info, motif.optionbgdef, 'string')
			if tonumber(port) ~= nil then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				config.ListenPort = tostring(port)
				setListenPort(port)
				t.items[item].vardisplay = getListenPort()
				modified = 1
			else
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			end
		end
		return true
	end,
	--Default Values
	['default'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
			config.AIRamping = true
			config.AIRandomColor = true
			config.AudioDucking = false
			config.AutoGuard = false
			config.ComboExtraFrameWindow = 1
			--config.CommonAir = "data/common.air"
			--config.CommonCmd = "data/common.cmd"
			--config.CommonScore = "data/score.zss"
			--config.CommonTag = "data/tag.zss"
			--config.ControllerStickSensitivity = 0.4
			config.Credits = 10
			config.DebugKeys = true
			config.Difficulty = 8
			config.ExternalShaders = {}
			config.Fullscreen = false
			config.GameWidth = 640
			config.GameHeight = 480
			config.GameSpeed = 100
			--config.IP = {}
			config.LifebarFontScale = 1
			config.LifeMul = 100
			config.ListenPort = "7500"
			config.LocalcoordScalingType = 1
			config.MaxDrawGames = -2
			--config.Motif = "data/system.def"
			config.MaxHelper = 56
			config.MaxPlayerProjectile = 256
			config.MaxExplod = 512
			config.MaxAfterImage = 128
			config.MSAA = false
			config.MulAttackLifeToPower = 0.7
			config.MulGetHitLifeToPower = 0.6
			config.MulSuperTargetDefence = 1.5
			config.NumSimul = {2, 4}
			config.NumTag = {2, 4}
			config.NumTurns = {2, 4}
			config.PostProcessingShader = 0
			config.QuickContinue = false
			config.QuickLaunch = 0
			config.RatioLife = {0.80, 1.0, 1.17, 1.40}
			config.RatioAttack = {0.82, 1.0, 1.17, 1.30}
			config.RoundsNumSingle = 2
			config.RoundsNumTeam = 2
			config.RoundTime = 99
			config.SimulLoseKO = true
			config.SingleVsTeamLife = 100
			--config.System = "external/script/main.lua"
			config.TagLoseKO = false
			config.TeamLifeAdjustment = false
			config.TeamPowerShare = true
			config.TurnsRecoveryBase = 0
			config.TurnsRecoveryBonus = 20
			config.VolumeBgm = 80
			config.VolumeMaster = 80
			config.VolumeSfx = 80
			--config.WindowMainIconLocation = {}
			--config.WindowTitle = "Ikemen GO"
			--config.XinputTriggerSensitivity = 0
			config.ZoomActive = false
			config.ZoomMax = 1.1
			config.ZoomMin = 0.75
			config.ZoomSpeed = 1
			loadLifebar(motif.files.fight)
			main.roundsNumSingle = getMatchWins()
			main.roundsNumTeam = getMatchWins()
			main.maxDrawGames = getMatchMaxDrawGames()
			options.f_resetVardisplay(options.menu)
			setListenPort(config.ListenPort)
			modified = 1
			needReload = 1
		end
		return true
	end,
	--Save and Return
	['savereturn'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			if modified == 1 then
				options.f_saveCfg()
			end
			main.f_menuFade('option_info', 'fadeout', cursorPosY, moveTxt, item, t)
			main.f_bgReset(motif.titlebgdef.bg)
			if motif.music.option_bgm ~= '' then
				main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
			end
			return false
		end
		return true
	end,
	--Return Without Saving
	['return'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			if needReload == 1 then
				main.f_warning(main.f_extractText(motif.warning_info.text_noreload), motif.option_info, motif.optionbgdef)
			end
			main.f_menuFade('option_info', 'fadeout', cursorPosY, moveTxt, item, t)
			main.f_bgReset(motif.titlebgdef.bg)
			if motif.music.option_bgm ~= '' then
				main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
			end
			return false
		end
		return true
	end,
	--Time Limit
	['roundtime'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.RoundTime < 1000 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.RoundTime = config.RoundTime + 1
			t.items[item].vardisplay = config.RoundTime
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.RoundTime > -1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.RoundTime = config.RoundTime - 1
			t.items[item].vardisplay = options.f_definedDisplay(config.RoundTime, {[-1] = motif.option_info.menu_valuename_none}, config.RoundTime)
			modified = 1
		end
		return true
	end,
	--Rounds to Win Single
	['roundsnumsingle'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and main.roundsNumSingle < 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			main.roundsNumSingle = main.roundsNumSingle + 1
			t.items[item].vardisplay = main.roundsNumSingle
			config.RoundsNumSingle = main.roundsNumSingle
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and main.roundsNumSingle > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			main.roundsNumSingle = main.roundsNumSingle - 1
			t.items[item].vardisplay = main.roundsNumSingle
			config.RoundsNumSingle = main.roundsNumSingle
			modified = 1
		end
		return true
	end,
	--Rounds to Win Simul/Tag
	['roundsnumteam'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and main.roundsNumTeam < 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			main.roundsNumTeam = main.roundsNumTeam + 1
			t.items[item].vardisplay = main.roundsNumTeam
			config.RoundsNumTeam = main.roundsNumTeam
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and main.roundsNumTeam > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			main.roundsNumTeam = main.roundsNumTeam - 1
			t.items[item].vardisplay = main.roundsNumTeam
			config.RoundsNumTeam = main.roundsNumTeam
			modified = 1
		end
		return true
	end,
	--Max Draw Games
	['maxdrawgames'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and main.maxDrawGames < 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			main.maxDrawGames = main.maxDrawGames + 1
			t.items[item].vardisplay = main.maxDrawGames
			config.MaxDrawGames = main.maxDrawGames
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and main.maxDrawGames > -1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			main.maxDrawGames = main.maxDrawGames - 1
			t.items[item].vardisplay = main.maxDrawGames
			config.MaxDrawGames = main.maxDrawGames
			modified = 1
		end
		return true
	end,
	--Difficulty Level
	['difficulty'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.Difficulty < 8 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.Difficulty = config.Difficulty + 1
			t.items[item].vardisplay = config.Difficulty
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.Difficulty > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.Difficulty = config.Difficulty - 1
			t.items[item].vardisplay = config.Difficulty
			modified = 1
		end
		return true
	end,
	--Credits
	['credits'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.Credits < 99 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.Credits = config.Credits + 1
			t.items[item].vardisplay = config.Credits
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.Credits > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.Credits = config.Credits - 1
			t.items[item].vardisplay = config.Credits
			modified = 1
		end
		return true
	end,
	--Quick Continue
	['quickcontinue'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.QuickContinue then
				config.QuickContinue = false
			else
				config.QuickContinue = true
				end
			t.items[item].vardisplay = options.f_boolDisplay(config.QuickContinue)
			modified = 1
		end
		return true
	end,
	--AI Ramping
	['airamping'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AIRamping then
				config.AIRamping = false
			else
				config.AIRamping = true
			end
			t.items[item].vardisplay = options.f_boolDisplay(config.AIRamping)
			modified = 1
		end
		return true
	end,
	--AI Palette
	['aipalette'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AIRandomColor then
				config.AIRandomColor = false
			else
				config.AIRandomColor = true
			end
			t.items[item].vardisplay = options.f_boolDisplay(config.AIRandomColor, motif.option_info.menu_valuename_random, motif.option_info.menu_valuename_default)
			modified = 1
		end
		return true
	end,
	--Resolution (submenu)
	['resolution'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			local t_pos = {}
			local ok = false
			for k, v in ipairs(t.submenu[t.items[item].itemname].items) do
				local width, height = v.itemname:match('^([0-9]+)x([0-9]+)$')
				if tonumber(width) == config.GameWidth and tonumber(height) == config.GameHeight then
					v.selected = true
					ok = true
				else
					v.selected = false
				end
				if v.itemname == 'customres' then
					t_pos = v
				end
			end
			if not ok and t_pos.selected ~= nil then
				t_pos.selected = true
			end
			t.submenu[t.items[item].itemname].loop()
			t.items[item].vardisplay = config.GameWidth .. 'x' .. config.GameHeight
		end
		return true
	end,
	--Fullscreen
	['fullscreen'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.Fullscreen then
				config.Fullscreen = false
			else
				config.Fullscreen = true
			end
			t.items[item].vardisplay = options.f_boolDisplay(config.Fullscreen)
			modified = 1
			needReload = 1
		end
		return true
	end,
	--MSAA
	['msaa'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.MSAA then
				config.MSAA = false
			else
				config.MSAA = true
			end
			t.items[item].vardisplay = options.f_boolDisplay(config.MSAA, motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
			modified = 1
			needReload = 1
		end
		return true
	end,
	--Shaders (submenu)
	['shaders'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if #options.t_shaders == 0 then
				main.f_warning(main.f_extractText(motif.warning_info.text_shaders), motif.option_info, motif.optionbgdef)
				return true
			end
			for k, v in ipairs(t.submenu[t.items[item].itemname].items) do
				if config.ExternalShaders[1] == v.itemname then
					v.selected = true
				else
					v.selected = false
				end
			end
			t.submenu[t.items[item].itemname].loop()
			t.items[item].vardisplay = f_externalShaderName()
			modified = 1
			needReload = 1
		end
		return true
	end,
	--Disable (shader)
	['noshader'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			config.ExternalShaders = {}
			config.PostProcessingShader = 0
			modified = 1
			needReload = 1
			return false
		end
		return true
	end,
	--Custom resolution
	['customres'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			local width = tonumber(main.f_input(main.f_extractText(motif.option_info.input_text_reswidth), motif.option_info, motif.optionbgdef, 'string'))
			if width ~= nil then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				local height = tonumber(main.f_input(main.f_extractText(motif.option_info.input_text_resheight), motif.option_info, motif.optionbgdef, 'string'))
				if height ~= nil then
					config.GameWidth = width
					config.GameHeight = height
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
			return false
		end
		return true
	end,
	--Master Volume
	['mastervolume'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.VolumeMaster < 200 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.VolumeMaster = config.VolumeMaster + 1
			t.items[item].vardisplay = config.VolumeMaster .. '%'
			setMasterVolume(config.VolumeMaster)
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.VolumeMaster > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.VolumeMaster = config.VolumeMaster - 1
			t.items[item].vardisplay = config.VolumeMaster  .. '%'
			setMasterVolume(config.VolumeMaster)
			modified = 1
		end
		return true
	end,
	--BGM Volume
	['bgmvolume'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.VolumeBgm < 100 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.VolumeBgm = config.VolumeBgm + 1
			t.items[item].vardisplay = config.VolumeBgm .. '%'
			setBgmVolume(config.VolumeBgm)
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.VolumeBgm > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.VolumeBgm = config.VolumeBgm - 1
			t.items[item].vardisplay = config.VolumeBgm .. '%'
			setBgmVolume(config.VolumeBgm)
			modified = 1
		end
		return true
	end,
	--SFX Volume
	['sfxvolume'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.VolumeSfx < 100 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.VolumeSfx = config.VolumeSfx + 1
			t.items[item].vardisplay = config.VolumeSfx .. '%'
			setWavVolume(config.VolumeSfx)
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.VolumeSfx > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.VolumeSfx = config.VolumeSfx - 1
			t.items[item].vardisplay = config.VolumeSfx .. '%'
			setWavVolume(config.VolumeSfx)
			modified = 1
		end
		return true
	end,
	--Audio Ducking
	['audioducking'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AudioDucking then
				config.AudioDucking = false
			else
				config.AudioDucking = true
			end
			t.items[item].vardisplay = options.f_boolDisplay(config.AudioDucking, motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
			setAudioDucking(config.AudioDucking)
			modified = 1
		end
		return true
	end,
	--Life
	['lifemul'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.LifeMul < 300 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.LifeMul = config.LifeMul + 10
			t.items[item].vardisplay = config.LifeMul .. '%'
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.LifeMul > 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.LifeMul = config.LifeMul - 10
			t.items[item].vardisplay = config.LifeMul .. '%'
			modified = 1
		end
		return true
	end,
	--Game Speed
	['gamespeed'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.GameSpeed < 200 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.GameSpeed = config.GameSpeed + 1
			t.items[item].vardisplay = config.GameSpeed .. '%'
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.GameSpeed > 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.GameSpeed = config.GameSpeed - 1
			t.items[item].vardisplay = config.GameSpeed .. '%'
			modified = 1
		end
		return true
	end,
	--Auto-Guard
	['autoguard'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.AutoGuard then
				config.AutoGuard = false
			else
				config.AutoGuard = true
			end
			t.items[item].vardisplay = options.f_boolDisplay(config.AutoGuard)
			modified = 1
		end
		return true
	end,
	--Single VS Team Life
	['singlevsteamlife'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.SingleVsTeamLife < 300 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.SingleVsTeamLife = config.SingleVsTeamLife + 10
			t.items[item].vardisplay = config.SingleVsTeamLife .. '%'
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.SingleVsTeamLife > 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.SingleVsTeamLife = config.SingleVsTeamLife - 10
			t.items[item].vardisplay = config.SingleVsTeamLife .. '%'
			modified = 1
		end
		return true
	end,
	--Team Life Adjustment
	['teamlifeadjustment'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.TeamLifeAdjustment then
				config.TeamLifeAdjustment = false
			else
				config.TeamLifeAdjustment = true
			end
			t.items[item].vardisplay = options.f_boolDisplay(config.TeamLifeAdjustment)
			modified = 1
		end
		return true
	end,
	--Team Power Share
	['teampowershare'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.TeamPowerShare then
				config.TeamPowerShare = false
			else
				config.TeamPowerShare = true
			end
			t.items[item].vardisplay = options.f_boolDisplay(config.TeamPowerShare)
			modified = 1
		end
		return true
	end,
	--Simul Player KOed Lose
	['simulloseko'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.SimulLoseKO then
				config.SimulLoseKO = false
			else
				config.SimulLoseKO = true
			end
			t.items[item].vardisplay = options.f_boolDisplay(config.SimulLoseKO)
			modified = 1
		end
		return true
	end,
	--Tag Partner KOed Lose
	['tagloseko'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.TagLoseKO then
				config.TagLoseKO = false
			else
				config.TagLoseKO = true
			end
			t.items[item].vardisplay = options.f_boolDisplay(config.TagLoseKO)
			modified = 1
		end
		return true
	end,
	--Turns Recovery Base
	['turnsrecoverybase'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.TurnsRecoveryBase < 100 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.TurnsRecoveryBase = config.TurnsRecoveryBase + 0.5
			t.items[item].vardisplay = config.TurnsRecoveryBase .. '%'
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.TurnsRecoveryBase > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.TurnsRecoveryBase = config.TurnsRecoveryBase - 0.5
			t.items[item].vardisplay = config.TurnsRecoveryBase .. '%'
			modified = 1
		end
		return true
	end,
	--Turns Recovery Bonus
	['turnsrecoverybonus'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.TurnsRecoveryBonus < 100 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.TurnsRecoveryBonus = config.TurnsRecoveryBonus + 0.5
			t.items[item].vardisplay = config.TurnsRecoveryBonus .. '%'
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.TurnsRecoveryBonus > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.TurnsRecoveryBonus = config.TurnsRecoveryBonus - 0.5
			t.items[item].vardisplay = config.TurnsRecoveryBonus .. '%'
			modified = 1
		end
		return true
	end,
	--Attack.LifeToPowerMul
	['attackpowermul'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MulAttackLifeToPower = options.f_precision(config.MulAttackLifeToPower + 0.1, '%.01f')
			t.items[item].vardisplay = config.MulAttackLifeToPower
			setAttackLifeToPowerMul(config.MulAttackLifeToPower)
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.MulAttackLifeToPower > 0.1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MulAttackLifeToPower = options.f_precision(config.MulAttackLifeToPower - 0.1, '%.01f')
			t.items[item].vardisplay = config.MulAttackLifeToPower
			setAttackLifeToPowerMul(config.MulAttackLifeToPower)
			modified = 1
		end
		return true
	end,
	--GetHit.LifeToPowerMul
	['gethitpowermul'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MulGetHitLifeToPower = options.f_precision(config.MulGetHitLifeToPower + 0.1, '%.01f')
			t.items[item].vardisplay = config.MulGetHitLifeToPower
			setGetHitLifeToPowerMul(config.MulGetHitLifeToPower)
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.MulGetHitLifeToPower > 0.1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MulGetHitLifeToPower = options.f_precision(config.MulGetHitLifeToPower - 0.1, '%.01f')
			t.items[item].vardisplay = config.MulGetHitLifeToPower
			setGetHitLifeToPowerMul(config.MulGetHitLifeToPower)
			modified = 1
		end
		return true
	end,
	--Super.TargetDefenceMul
	['superdefencemul'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MulSuperTargetDefence = options.f_precision(config.MulSuperTargetDefence + 0.1, '%.01f')
			t.items[item].vardisplay = config.MulSuperTargetDefence
			setSuperTargetDefenceMul(config.MulSuperTargetDefence)
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.MulSuperTargetDefence > 0.1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MulSuperTargetDefence = options.f_precision(config.MulSuperTargetDefence - 0.1, '%.01f')
			t.items[item].vardisplay = config.MulSuperTargetDefence
			setSuperTargetDefenceMul(config.MulSuperTargetDefence)
			modified = 1
		end
		return true
	end,
	--Min Turns Chars
	['minturns'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.NumTurns[1] < config.NumTurns[2] then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.NumTurns[1] = config.NumTurns[1] + 1
			t.items[item].vardisplay = config.NumTurns[1]
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.NumTurns[1] > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.NumTurns[1] = config.NumTurns[1] - 1
			t.items[item].vardisplay = config.NumTurns[1]
			modified = 1
		end
		return true
	end,
	--Max Turns Chars
	['maxturns'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.NumTurns[2] < 8 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.NumTurns[2] = config.NumTurns[2] + 1
			t.items[item].vardisplay = config.NumTurns[2]
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.NumTurns[2] > config.NumTurns[1] then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.NumTurns[2] = config.NumTurns[2] - 1
			t.items[item].vardisplay = config.NumTurns[2]
			modified = 1
		end
		return true
	end,
	--Min Simul Chars
	['minsimul'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.NumSimul[1] < config.NumSimul[2] then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.NumSimul[1] = config.NumSimul[1] + 1
			t.items[item].vardisplay = config.NumSimul[1]
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.NumSimul[1] > 2 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.NumSimul[1] = config.NumSimul[1] - 1
			t.items[item].vardisplay = config.NumSimul[1]
			modified = 1
		end
		return true
	end,
	--Max Simul Chars
	['maxsimul'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.NumSimul[2] < 8 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.NumSimul[2] = config.NumSimul[2] + 1
			t.items[item].vardisplay = config.NumSimul[2]
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.NumSimul[2] > config.NumSimul[1] then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.NumSimul[2] = config.NumSimul[2] - 1
			t.items[item].vardisplay = config.NumSimul[2]
			modified = 1
		end
		return true
	end,
	--Min Tag Chars
	['mintag'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.NumTag[1] < config.NumTag[2] then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.NumTag[1] = config.NumTag[1] + 1
			t.items[item].vardisplay = config.NumTag[1]
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.NumTag[1] > 2 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.NumTag[1] = config.NumTag[1] - 1
			t.items[item].vardisplay = config.NumTag[1]
			modified = 1
		end
		return true
	end,
	--Max Tag Chars
	['maxtag'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.NumTag[2] < 4 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.NumTag[2] = config.NumTag[2] + 1
			t.items[item].vardisplay = config.NumTag[2]
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.NumTag[2] > config.NumTag[1] then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.NumTag[2] = config.NumTag[2] - 1
			t.items[item].vardisplay = config.NumTag[2]
			modified = 1
		end
		return true
	end,
	--Debug Keys
	['debugkeys'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.DebugKeys then
				config.DebugKeys = false
			else
				config.DebugKeys = true
			end
			t.items[item].vardisplay = options.f_boolDisplay(config.DebugKeys, motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
			setAllowDebugKeys(config.DebugKeys)
			modified = 1
		end
		return true
	end,
	--Quick Launch
	['quicklaunch'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if main.input({1, 2}, {'$F'}) and config.QuickLaunch < #t_quicklaunchNames then
				config.QuickLaunch = config.QuickLaunch + 1
			elseif main.input({1, 2}, {'$B'}) and config.QuickLaunch > 0 then
				config.QuickLaunch = config.QuickLaunch - 1
			end
			t.items[item].vardisplay = t_quicklaunchNames[config.QuickLaunch]
			modified = 1
		end
		return true
	end,
	--Lifebar Font Scale
	['lifebarfontscale'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.LifebarFontScale = options.f_precision(config.LifebarFontScale + 0.1, '%.01f')
			t.items[item].vardisplay = config.LifebarFontScale
			modified = 1
			needReload = 1
		elseif main.input({1, 2}, {'$B'}) and config.LifebarFontScale > 0.1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.LifebarFontScale = options.f_precision(config.LifebarFontScale - 0.1, '%.01f')
			t.items[item].vardisplay = config.LifebarFontScale
			modified = 1
			needReload = 1
		end
		return true
	end,
	--HelperMax
	['helpermax'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MaxHelper = config.MaxHelper + 1
			t.items[item].vardisplay = config.MaxHelper
			modified = 1
			needReload = 1
		elseif main.input({1, 2}, {'$B'}) and config.MaxHelper > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MaxHelper = config.MaxHelper - 1
			t.items[item].vardisplay = config.MaxHelper
			modified = 1
			needReload = 1
		end
		return true
	end,
	--PlayerProjectileMax
	['projectilemax'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MaxPlayerProjectile = config.MaxPlayerProjectile + 1
			t.items[item].vardisplay = config.MaxPlayerProjectile
			modified = 1
			needReload = 1
		elseif main.input({1, 2}, {'$B'}) and config.MaxPlayerProjectile > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MaxPlayerProjectile = config.MaxPlayerProjectile - 1
			t.items[item].vardisplay = config.MaxPlayerProjectile
			modified = 1
			needReload = 1
		end
		return true
	end,
	--ExplodMax
	['explodmax'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MaxExplod = config.MaxExplod + 1
			t.items[item].vardisplay = config.MaxExplod
			modified = 1
			needReload = 1
		elseif main.input({1, 2}, {'$B'}) and config.MaxExplod > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MaxExplod = config.MaxExplod - 1
			t.items[item].vardisplay = config.MaxExplod
			modified = 1
			needReload = 1
		end
		return true
	end,
	--AfterImageMax
	['afterimagemax'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MaxAfterImage = config.MaxAfterImage + 1
			t.items[item].vardisplay = config.MaxAfterImage
			modified = 1
			needReload = 1
		elseif main.input({1, 2}, {'$B'}) and config.MaxAfterImage > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.MaxAfterImage = config.MaxAfterImage - 1
			t.items[item].vardisplay = config.MaxAfterImage
			modified = 1
			needReload = 1
		end
		return true
	end,
	--Zoom Active
	['zoomactive'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F', '$B', 'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if config.ZoomActive then
				config.ZoomActive = false
			else
				config.ZoomActive = true
			end
			t.items[item].vardisplay = options.f_boolDisplay(config.ZoomActive)
			modified = 1
		end
		return true
	end,
	--Default Max Zoom Out
	['maxzoomout'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.ZoomMin < 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.ZoomMin = options.f_precision(config.ZoomMin + 0.05, '%.02f')
			t.items[item].vardisplay = config.ZoomMin
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.ZoomMin > 0.05 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.ZoomMin = options.f_precision(config.ZoomMin - 0.05, '%.02f')
			t.items[item].vardisplay = config.ZoomMin
			modified = 1
		end
		return true
	end,
	--Default Max Zoom In
	['maxzoomin'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.ZoomMax < 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.ZoomMax = options.f_precision(config.ZoomMax + 0.05, '%.02f')
			t.items[item].vardisplay = config.ZoomMax
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.ZoomMax > 0.05 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.ZoomMax = options.f_precision(config.ZoomMax - 0.05, '%.02f')
			t.items[item].vardisplay = config.ZoomMax
			modified = 1
		end
		return true
	end,
	--Default Zoom Speed
	['zoomspeed'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'$F'}) and config.ZoomSpeed < 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.ZoomSpeed = options.f_precision(config.ZoomSpeed + 0.1, '%.01f')
			t.items[item].vardisplay = config.ZoomSpeed
			modified = 1
		elseif main.input({1, 2}, {'$B'}) and config.ZoomSpeed > 0.1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			config.ZoomSpeed = options.f_precision(config.ZoomSpeed - 0.1, '%.01f')
			t.items[item].vardisplay = config.ZoomSpeed
			modified = 1
		end
		return true
	end,
	--Key Config
	['keyboard'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			--t.submenu[t.items[item].itemname].loop()
			--options.menu.submenu.input.loop{}
			options.f_keyCfg('KeyConfig', t.items[item].itemname, t.submenu[t.items[item].itemname].title)
		end
		return true
	end,
	--Joystick Config
	['gamepad'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			--t.submenu[t.items[item].itemname].loop()
			--options.menu.submenu.input.loop{}
			options.f_keyCfg('JoystickConfig', t.items[item].itemname, t.submenu[t.items[item].itemname].title)
		end
		return true
	end,
	--Default
	['inputdefault'] = function(cursorPosY, moveTxt, item, t)
		if main.input({1, 2}, {'pal'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
			options.f_keyDefault()
			modified = 1
			needReload = 1 --TODO: won't be needed if we add a function that can edit sys.keyConfig and sys.JoystickConfig from lua
		end
		return true
	end,
}
--external shaders
options.t_shaders = {}
for k, v in ipairs(GetDirectoryFiles('external/shaders')) do
	v:gsub('^(.-)([^\\/]+)%.([^%.\\/]-)$', function(path, filename, ext)
		path = path:gsub('\\', '/')
		ext = ext:lower()
		if ext == 'frag' then
			table.insert(options.t_shaders, {['path'] = path, ['filename'] = filename})
		end
		if ext:match('vert') or ext:match('frag') --[[or ext:match('shader')]] then
			options.t_itemname[path .. filename] = function(cursorPosY, moveTxt, item, t)
				if main.input({1, 2}, {'pal'}) then
					sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
					config.ExternalShaders = {path .. filename}
					config.PostProcessingShader = 4
					return false
				end
				return true
			end
		end
	end)
end
for k, v in ipairs(main.t_sort.option_info) do
	--resolution
	if v:match('_[0-9]+x[0-9]+$') then
		local width, height = v:match('_([0-9]+)x([0-9]+)$')
		options.t_itemname[width .. 'x' .. height] = function(cursorPosY, moveTxt, item, t)
			if main.input({1, 2}, {'pal'}) then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				config.GameWidth = tonumber(width)
				config.GameHeight = tonumber(height)
				if (config.GameHeight / 3 * 4) ~= config.GameWidth then
					main.f_warning(main.f_extractText(motif.warning_info.text_res), motif.option_info, motif.optionbgdef)
				end
				modified = 1
				needReload = 1
				return false
			end
			return true
		end
	--ratio
	elseif v:match('_ratio[1-4]+[al].-$') then
		local ratioLevel, tmp1, tmp2 = v:match('_ratio([1-4])([al])(.-)$')
		options.t_itemname['ratio' .. ratioLevel .. tmp1 .. tmp2] = function(cursorPosY, moveTxt, item, t)
			local ratioType = tmp1:upper() .. tmp2
			ratioLevel = tonumber(ratioLevel)
			if main.input({1, 2}, {'$F'}) then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['Ratio' .. ratioType][ratioLevel] = options.f_precision(config['Ratio' .. ratioType][ratioLevel] + 0.01, '%.02f')
				t.items[item].vardisplay = options.f_displayRatio(config['Ratio' .. ratioType][ratioLevel])
				modified = 1
			elseif main.input({1, 2}, {'$B'}) and config['Ratio' .. ratioType][ratioLevel] > 0.01 then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				config['Ratio' .. ratioType][ratioLevel] = options.f_precision(config['Ratio' .. ratioType][ratioLevel] - 0.01, '%.02f')
				t.items[item].vardisplay = options.f_displayRatio(config['Ratio' .. ratioType][ratioLevel])
				modified = 1
			end
			return true
		end
	end
end
if main.debugLog then main.f_printTable(options.t_itemname, 'debug/t_optionsItemname.txt') end

function options.createMenu(tbl, bool_bgreset, bool_main)
	return function()
		--main.f_cmdInput()
		local cursorPosY = 1
		local moveTxt = 0
		local item = 1
		local t = tbl.items
		if bool_bgreset then
			main.f_bgReset(motif.optionbgdef.bg)
			main.f_playBGM(false, motif.music.option_bgm, motif.music.option_bgm_loop, motif.music.option_bgm_volume, motif.music.option_bgm_loopstart, motif.music.option_bgm_loopend)
			if #main.t_sort.option_info == 0 then
				main.f_warning(main.f_extractText(motif.warning_info.text_options), motif.option_info, motif.optionbgdef)
				return
			end
		end
		while true do
			options.f_menuCommonDraw(cursorPosY, moveTxt, item, t)
			cursorPosY, moveTxt, item = options.f_menuCommonCalc(cursorPosY, moveTxt, item, t)
			txt_title:update({text = tbl.title})
			if esc() then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				if bool_main then
					if modified == 1 then
						options.f_saveCfg()
					end
					main.f_menuFade('option_info', 'fadeout', cursorPosY, moveTxt, item, t)
					main.f_bgReset(motif.titlebgdef.bg)
					if motif.music.option_bgm ~= '' then
						main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
					end
				end
				break
			elseif options.t_itemname[t[item].itemname] ~= nil then
				if not options.t_itemname[t[item].itemname](cursorPosY, moveTxt, item, tbl) then
					break
				end
			elseif main.input({1, 2}, {'pal'}) then
				local f = main.f_checkSubmenu(tbl.submenu[t[item].itemname])
				if f ~= '' and not options.t_itemname[f](cursorPosY, moveTxt, item, tbl) then
					break
				end
			end
		end
	end
end

--reset vardisplay in tables
function options.f_resetVardisplay(t)
	for k, v in pairs(t) do
		if k == 'items' and type(v) == "table" and #v > 0 then
			for i, v2 in ipairs(v) do
				if v2.vardisplay ~= nil then
					v2.vardisplay = options.f_vardisplay(v2.itemname)
				end
			end
		elseif k == 'submenu' and type(v) == "table" then
			for k2, v2 in pairs(v) do
				options.f_resetVardisplay(v2)
			end
		end
	end
end

function options.f_vardisplay(itemname)
	if itemname == 'portchange' then return config.ListenPort end
	if itemname == 'roundtime' then return options.f_definedDisplay(config.RoundTime, {[-1] = motif.option_info.menu_valuename_none}, config.RoundTime) end
	if itemname == 'roundsnumsingle' then return main.roundsNumSingle end
	if itemname == 'roundsnumteam' then return main.roundsNumTeam end
	if itemname == 'maxdrawgames' then return main.maxDrawGames end
	if itemname == 'difficulty' then return config.Difficulty end
	if itemname == 'credits' then return config.Credits end
	if itemname == 'quickcontinue' then return options.f_boolDisplay(config.QuickContinue) end
	if itemname == 'airamping' then return options.f_boolDisplay(config.AIRamping) end
	if itemname == 'aipalette' then return options.f_boolDisplay(config.AIRandomColor, motif.option_info.menu_valuename_random, motif.option_info.menu_valuename_default) end
	if itemname == 'resolution' then return config.GameWidth .. 'x' .. config.GameHeight end
	if itemname == 'fullscreen' then return options.f_boolDisplay(config.Fullscreen) end
	if itemname == 'msaa' then return options.f_boolDisplay(config.MSAA, motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled) end
	if itemname == 'shaders' then return f_externalShaderName() end
	if itemname == 'mastervolume' then return config.VolumeMaster .. '%' end
	if itemname == 'bgmvolume' then return config.VolumeBgm .. '%' end
	if itemname == 'sfxvolume' then return config.VolumeSfx .. '%' end
	if itemname == 'audioducking' then return options.f_boolDisplay(config.AudioDucking, motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled) end
	if itemname == 'lifemul' then return config.LifeMul .. '%' end
	if itemname == 'gamespeed' then return config.GameSpeed .. '%' end
	if itemname == 'autoguard' then return options.f_boolDisplay(config.AutoGuard) end
	if itemname == 'singlevsteamlife' then return config.SingleVsTeamLife .. '%' end
	if itemname == 'teamlifeadjustment' then return options.f_boolDisplay(config.TeamLifeAdjustment) end
	if itemname == 'teampowershare' then return options.f_boolDisplay(config.TeamPowerShare) end
	if itemname == 'simulloseko' then return options.f_boolDisplay(config.SimulLoseKO) end
	if itemname == 'tagloseko' then return options.f_boolDisplay(config.TagLoseKO) end
	if itemname == 'turnsrecoverybase' then return config.TurnsRecoveryBase .. '%' end
	if itemname == 'turnsrecoverybonus' then return config.TurnsRecoveryBonus .. '%' end
	if itemname == 'ratio1life' then return options.f_displayRatio(config.RatioLife[1]) end
	if itemname == 'ratio1attack' then return options.f_displayRatio(config.RatioAttack[1]) end
	if itemname == 'ratio2life' then return options.f_displayRatio(config.RatioLife[2]) end
	if itemname == 'ratio2attack' then return options.f_displayRatio(config.RatioAttack[2]) end
	if itemname == 'ratio3life' then return options.f_displayRatio(config.RatioLife[3]) end
	if itemname == 'ratio3attack' then return options.f_displayRatio(config.RatioAttack[3]) end
	if itemname == 'ratio4life' then return options.f_displayRatio(config.RatioLife[4]) end
	if itemname == 'ratio4attack' then return options.f_displayRatio(config.RatioAttack[4]) end
	if itemname == 'attackpowermul' then return config.MulAttackLifeToPower end
	if itemname == 'gethitpowermul' then return config.MulGetHitLifeToPower end
	if itemname == 'superdefencemul' then return config.MulSuperTargetDefence end
	if itemname == 'minturns' then return config.NumTurns[1] end
	if itemname == 'maxturns' then return config.NumTurns[2] end
	if itemname == 'minsimul' then return config.NumSimul[1] end
	if itemname == 'maxsimul' then return config.NumSimul[2] end
	if itemname == 'mintag' then return config.NumTag[1] end
	if itemname == 'maxtag' then return config.NumTag[2] end
	if itemname == 'debugkeys' then return options.f_boolDisplay(config.DebugKeys, motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled) end
	if itemname == 'quicklaunch' then return t_quicklaunchNames[config.QuickLaunch] end
	if itemname == 'lifebarfontscale' then return config.LifebarFontScale end
	if itemname == 'helpermax' then return config.MaxHelper end
	if itemname == 'projectilemax' then return config.MaxPlayerProjectile end
	if itemname == 'explodmax' then return config.MaxExplod end
	if itemname == 'afterimagemax' then return config.MaxAfterImage end
	if itemname == 'zoomactive' then return options.f_boolDisplay(config.ZoomActive) end
	if itemname == 'maxzoomout' then return config.ZoomMin end
	if itemname == 'maxzoomin' then return config.ZoomMax end
	if itemname == 'zoomspeed' then return config.ZoomSpeed end
	return ''
end

local function f_itemnameUpper(title)
	if motif.option_info.menu_title_uppercase == 1 then
		return title:upper()
	end
	return title
end

--dynamically generates all option screen menus and submenus using itemname data stored in main.t_sort table
options.menu = {['submenu'] = {}, ['items'] = {}, ['title'] = f_itemnameUpper(motif.title_info.menu_itemname_options)}
options.menu.loop = options.createMenu(options.menu, true, true)
local t_pos = {} --for storing current options.menu table position
local lastNum = 0
for i = 1, #main.t_sort.option_info do
	for j, c in ipairs(main.f_strsplit('_', main.t_sort.option_info[i])) do --split using "_" delimiter
		--populate shaders submenu
		if main.t_sort.option_info[i]:match('_shaders_back$') and c == 'back' then
			for k = #options.t_shaders, 1, -1 do
				table.insert(t_pos.items, 1, {data = text:create({}), itemname = options.t_shaders[k].path .. options.t_shaders[k].filename, displayname = options.t_shaders[k].filename, vardata = text:create({}), vardisplay = options.f_vardisplay(c), selected = false})
			end
		end
		--appending the menu table
		if j == 1 then --first string after menu.itemname (either reserved one or custom submenu assignment)
			if options.menu.submenu[c] == nil or c == 'empty' then
				options.menu.submenu[c] = {['submenu'] = {}, ['items'] = {}, ['title'] = f_itemnameUpper(motif.option_info['menu_itemname_' .. main.t_sort.option_info[i]])}
				options.menu.submenu[c].loop = options.createMenu(options.menu.submenu[c], false, false)
				if not main.t_sort.option_info[i]:match(c .. '_') then
					table.insert(options.menu.items, {data = text:create({}), itemname = c, displayname = motif.option_info['menu_itemname_' .. main.t_sort.option_info[i]], vardata = text:create({}), vardisplay = options.f_vardisplay(c), selected = false})
				end
			end
			t_pos = options.menu.submenu[c]
		else --following strings
			if t_pos.submenu[c] == nil or c == 'empty' then
				t_pos.submenu[c] = {['submenu'] = {}, ['items'] = {}, ['title'] = f_itemnameUpper(motif.option_info['menu_itemname_' .. main.t_sort.option_info[i]])}
				t_pos.submenu[c].loop = options.createMenu(t_pos.submenu[c], false, false)
				table.insert(t_pos.items, {data = text:create({}), itemname = c, displayname = motif.option_info['menu_itemname_' .. main.t_sort.option_info[i]], vardata = text:create({}), vardisplay = options.f_vardisplay(c), selected = false})
			end
			if j > lastNum then
				t_pos = t_pos.submenu[c]
			end
		end
		lastNum = j
	end
end
if main.debugLog then main.f_printTable(options.menu, 'debug/t_optionsMenu.txt') end

--;===========================================================
--; KEY SETTINGS
--;===========================================================
local t_keyCfg = {
	{data = {text:create({}), text:create({})}, itemname = 'empty', displayname = ''},
	{data = {text:create({}), text:create({})}, itemname = 'configall', displayname = motif.option_info.menu_itemname_key_all, infodata = {text:create({}), text:create({})}, infodisplay = ''},
	{data = {text:create({}), text:create({})}, itemname = 'up', displayname = motif.option_info.menu_itemname_key_up, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'down', displayname = motif.option_info.menu_itemname_key_down, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'left', displayname = motif.option_info.menu_itemname_key_left, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'right', displayname = motif.option_info.menu_itemname_key_right, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'a', displayname = motif.option_info.menu_itemname_key_a, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'b', displayname = motif.option_info.menu_itemname_key_b, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'c', displayname = motif.option_info.menu_itemname_key_c, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'x', displayname = motif.option_info.menu_itemname_key_x, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'y', displayname = motif.option_info.menu_itemname_key_y, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'z', displayname = motif.option_info.menu_itemname_key_z, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'start', displayname = motif.option_info.menu_itemname_key_start, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'd', displayname = motif.option_info.menu_itemname_key_d, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'w', displayname = motif.option_info.menu_itemname_key_w, vardata = {text:create({}), text:create({})}},
	{data = {text:create({}), text:create({})}, itemname = 'back', displayname = motif.option_info.menu_itemname_key_back, infodata = {text:create({}), text:create({})}, infodisplay = motif.option_info.menu_valuename_esc},
}
--t_keyCfg = main.f_cleanTable(t_keyCfg, main.t_sort.option_info)

local txt_keyController = {text:create({}), text:create({})}
function options.f_keyCfg(cfgType, controller, title)
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
	txt_title:update({text = title})
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
			if v ~= tostring(motif.option_info.menu_valuename_nokey) then --if button is not disabled
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
				commandBufReset(main.cmd[1])
				commandBufReset(main.cmd[2])
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
				commandBufReset(main.cmd[1])
				commandBufReset(main.cmd[2])
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
					t[item]['vardisplay' .. player] = motif.option_info.menu_valuename_nokey
					config[cfgType][player].Buttons[item - item_start] = tostring(motif.option_info.menu_valuename_nokey)
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
					commandBufReset(main.cmd[1])
					commandBufReset(main.cmd[2])
				end
			end
			resetKey()
			key = ''
		--move up / down / left / right
		elseif main.input({1, 2}, {'$U'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			item = item - 1
		elseif main.input({1, 2}, {'$D'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			item = item + 1
		elseif main.input({1, 2}, {'$F', '$B'}) then
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
			if main.input({1, 2}, {'$U'}) and cursorPosY > item_start then
				cursorPosY = cursorPosY - 1
			elseif main.input({1, 2}, {'$D'}) and cursorPosY < motif.option_info.menu_window_visibleitems then
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
				txt_title:update({text = motif.option_info.title_text_input})
				break
			--Config all
			elseif (t[item].itemname == 'configall' and main.input({1, 2}, {'pal'})) or getKey() == 'F1' or getKey() == 'F2' then
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
			elseif t[item].itemname == 'back' and main.input({1, 2}, {'pal'}) then
				if t_conflict[joyNum] then
					main.f_warning(main.f_extractText(motif.warning_info.text_keys), motif.option_info, motif.optionbgdef)
				else
					sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
					txt_title:update({text = motif.option_info.title_text_input})
					break
				end
			--individual buttons
			elseif main.input({1, 2}, {'pal'}) then
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
						t[item]['vardisplay' .. player] = motif.option_info.menu_valuename_nokey
						config[cfgType][player].Buttons[item - item_start] = motif.option_info.menu_valuename_nokey
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
			txt_keyController[i]:update({
				font =   motif.font_data[motif.option_info['menu_item_key_p' .. i .. '_font'][1]],
				bank =   motif.option_info['menu_item_key_p' .. i .. '_font'][2],
				align =  motif.option_info['menu_item_key_p' .. i .. '_font'][3],
				text =   motif.option_info['menu_itemname_key_p' .. i],
				x =      motif.option_info['menu_item_p' .. i .. '_pos'][1],
				y =      motif.option_info['menu_item_p' .. i .. '_pos'][2],
				scaleX = motif.option_info['menu_item_key_p' .. i .. '_font_scale'][1],
				scaleY = motif.option_info['menu_item_key_p' .. i .. '_font_scale'][2],
				r =      motif.option_info['menu_item_key_p' .. i .. '_font'][4],
				g =      motif.option_info['menu_item_key_p' .. i .. '_font'][5],
				b =      motif.option_info['menu_item_key_p' .. i .. '_font'][6],
				src =    motif.option_info['menu_item_key_p' .. i .. '_font'][7],
				dst =    motif.option_info['menu_item_key_p' .. i .. '_font'][8],
				defsc =  motif.defaultOptions
			})
			txt_keyController[i]:draw()
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
		txt_title:draw()
		--draw menu items
		for i = 1, #t do
			for j = 1, 2 do
				if i > item - cursorPosY then
					if t[i].itemname == 'configall' then
						if j == 1 then --player1 side (left)
							t[i].infodisplay = motif.option_info.menu_valuename_f1
						else --player2 side (right)
							t[i].infodisplay = motif.option_info.menu_valuename_f2
						end
					end
					if i == item and j == player then --active item
						--draw displayname
						t[i].data[j]:update({
							font =   motif.font_data[motif.option_info.menu_item_active_font[1]],
							bank =   motif.option_info.menu_item_active_font[2],
							align =  motif.option_info.menu_item_active_font[3],
							text =   t[i].displayname,
							x =      t_pos[j][1],
							y =      t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
							scaleX = motif.option_info.menu_item_active_font_scale[1],
							scaleY = motif.option_info.menu_item_active_font_scale[2],
							r =      motif.option_info.menu_item_active_font[4],
							g =      motif.option_info.menu_item_active_font[5],
							b =      motif.option_info.menu_item_active_font[6],
							src =    motif.option_info.menu_item_active_font[7],
							dst =    motif.option_info.menu_item_active_font[8],
							defsc =  motif.defaultOptions
						})
						t[i].data[j]:draw()
						--draw vardata
						if t[i].vardata ~= nil then
							if t_keyList[joyNum][tostring(t[i]['vardisplay' .. j])] ~= nil and t_keyList[joyNum][tostring(t[i]['vardisplay' .. j])] > 1 then
								t[i].vardata[j]:update({
									font =   motif.font_data[motif.option_info.menu_item_value_conflict_font[1]],
									bank =   motif.option_info.menu_item_value_conflict_font[2],
									align =  motif.option_info.menu_item_value_conflict_font[3],
									text =   t[i]['vardisplay' .. j],
									x =      t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
									y =      t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
									scaleX = motif.option_info.menu_item_value_conflict_font_scale[1],
									scaleY = motif.option_info.menu_item_value_conflict_font_scale[2],
									r =      motif.option_info.menu_item_value_conflict_font[4],
									g =      motif.option_info.menu_item_value_conflict_font[5],
									b =      motif.option_info.menu_item_value_conflict_font[6],
									src =    motif.option_info.menu_item_value_conflict_font[7],
									dst =    motif.option_info.menu_item_value_conflict_font[8],
									defsc =  motif.defaultOptions
								})
								t[i].vardata[j]:draw()
								t_conflict[joyNum] = true
							else
								t[i].vardata[j]:update({
									font =   motif.font_data[motif.option_info.menu_item_value_active_font[1]],
									bank =   motif.option_info.menu_item_value_active_font[2],
									align =  motif.option_info.menu_item_value_active_font[3],
									text =   t[i]['vardisplay' .. j],
									x =      t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
									y =      t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
									scaleX = motif.option_info.menu_item_value_active_font_scale[1],
									scaleY = motif.option_info.menu_item_value_active_font_scale[2],
									r =      motif.option_info.menu_item_value_active_font[4],
									g =      motif.option_info.menu_item_value_active_font[5],
									b =      motif.option_info.menu_item_value_active_font[6],
									src =    motif.option_info.menu_item_value_active_font[7],
									dst =    motif.option_info.menu_item_value_active_font[8],
									defsc =  motif.defaultOptions
								})
								t[i].vardata[j]:draw()
							end
						--draw infodata
						elseif t[i].infodata ~= nil then
							t[i].infodata[j]:update({
								font =   motif.font_data[motif.option_info.menu_item_info_active_font[1]],
								bank =   motif.option_info.menu_item_info_active_font[2],
								align =  motif.option_info.menu_item_info_active_font[3],
								text =   t[i].infodisplay,
								x =      t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
								y =      t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
								scaleX = motif.option_info.menu_item_value_active_font_scale[1],
								scaleY = motif.option_info.menu_item_value_active_font_scale[2],
								r =      motif.option_info.menu_item_info_active_font[4],
								g =      motif.option_info.menu_item_info_active_font[5],
								b =      motif.option_info.menu_item_info_active_font[6],
								src =    motif.option_info.menu_item_info_active_font[7],
								dst =    motif.option_info.menu_item_info_active_font[8],
								defsc =  motif.defaultOptions
							})
							t[i].infodata[j]:draw()
						end
					else --inactive item
						--draw displayname
						t[i].data[j]:update({
							font =   motif.font_data[motif.option_info.menu_item_font[1]],
							bank =   motif.option_info.menu_item_font[2],
							align =  motif.option_info.menu_item_font[3],
							text =   t[i].displayname,
							x =      t_pos[j][1],
							y =      t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
							scaleX = motif.option_info.menu_item_font_scale[1],
							scaleY = motif.option_info.menu_item_font_scale[2],
							r =      motif.option_info.menu_item_font[4],
							g =      motif.option_info.menu_item_font[5],
							b =      motif.option_info.menu_item_font[6],
							src =    motif.option_info.menu_item_font[7],
							dst =    motif.option_info.menu_item_font[8],
							defsc =  motif.defaultOptions
						})
						t[i].data[j]:draw()
						--draw vardata
						if t[i].vardata ~= nil then
							if t_keyList[joyNum][tostring(t[i]['vardisplay' .. j])] ~= nil and t_keyList[joyNum][tostring(t[i]['vardisplay' .. j])] > 1 then
								t[i].vardata[j]:update({
									font =   motif.font_data[motif.option_info.menu_item_value_conflict_font[1]],
									bank =   motif.option_info.menu_item_value_conflict_font[2],
									align =  motif.option_info.menu_item_value_conflict_font[3],
									text =   t[i]['vardisplay' .. j],
									x =      t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
									y =      t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
									scaleX = motif.option_info.menu_item_value_conflict_font_scale[1],
									scaleY = motif.option_info.menu_item_value_conflict_font_scale[2],
									r =      motif.option_info.menu_item_value_conflict_font[4],
									g =      motif.option_info.menu_item_value_conflict_font[5],
									b =      motif.option_info.menu_item_value_conflict_font[6],
									src =    motif.option_info.menu_item_value_conflict_font[7],
									dst =    motif.option_info.menu_item_value_conflict_font[8],
									defsc =  motif.defaultOptions
								})
								t[i].vardata[j]:draw()
								t_conflict[joyNum] = true
							else
								t[i].vardata[j]:update({
									font =   motif.font_data[motif.option_info.menu_item_value_font[1]],
									bank =   motif.option_info.menu_item_value_font[2],
									align =  motif.option_info.menu_item_value_font[3],
									text =   t[i]['vardisplay' .. j],
									x =      t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
									y =      t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
									scaleX = motif.option_info.menu_item_value_font_scale[1],
									scaleY = motif.option_info.menu_item_value_font_scale[2],
									r =      motif.option_info.menu_item_value_font[4],
									g =      motif.option_info.menu_item_value_font[5],
									b =      motif.option_info.menu_item_value_font[6],
									src =    motif.option_info.menu_item_value_font[7],
									dst =    motif.option_info.menu_item_value_font[8],
									defsc =  motif.defaultOptions
								})
								t[i].vardata[j]:draw()
							end
						--draw infodata
						elseif t[i].infodata ~= nil then
							t[i].infodata[j]:update({
								font =   motif.font_data[motif.option_info.menu_item_info_font[1]],
								bank =   motif.option_info.menu_item_info_font[2],
								align =  motif.option_info.menu_item_info_font[3],
								text =   t[i].infodisplay,
								x =      t_pos[j][1] + motif.option_info.menu_key_item_spacing[1],
								y =      t_pos[j][2] + (i - 1) * motif.option_info.menu_key_item_spacing[2] - moveTxt,
								scaleX = motif.option_info.menu_item_value_active_font_scale[1],
								scaleY = motif.option_info.menu_item_value_active_font_scale[2],
								r =      motif.option_info.menu_item_info_font[4],
								g =      motif.option_info.menu_item_info_font[5],
								b =      motif.option_info.menu_item_info_font[6],
								src =    motif.option_info.menu_item_info_font[7],
								dst =    motif.option_info.menu_item_info_font[8],
								defsc =  motif.defaultOptions
							})
							t[i].infodata[j]:draw()
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

local options = {}
--;===========================================================
--; COMMON
--;===========================================================
options.modified = false
options.needReload = false

--return string depending on bool
function options.f_boolDisplay(bool, t, f)
	if bool == true then
		return t or motif.option_info.menu_valuename_yes
	end
	return f or motif.option_info.menu_valuename_no
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

--- Save the current configuration to the config file and handle common file modifications
local t_commonFilesOriginal = gameOption('Common')
function options.f_saveCfg(reload)
    -- Restore the original content of the common files
	local t_commonFiles = gameOption('Common')
	for _, k in ipairs({'Air', 'Cmd', 'Const', 'States', 'Fx', 'Modules', 'Lua'}) do
		modifyGameOption('Common.' .. k, t_commonFilesOriginal[k][k:lower()] or {})
	end
    -- Save the current configuration to 'config.ini'
	saveGameOption(main.flags['-config'])
    -- Reload the game if the reload parameter is true
	if reload then
		main.f_warning(main.f_extractText(motif.warning_info.text_reload_text), motif.optionbgdef)
		os.exit()
	end
    -- Reapply modified common file arrays after saving
	for _, k in ipairs({'Air', 'Cmd', 'Const', 'States', 'Fx', 'Modules', 'Lua'}) do
		modifyGameOption('Common.' .. k, t_commonFiles[k][k:lower()] or {})
	end
end

options.txt_title = main.f_createTextImg(motif.option_info, 'title', {defsc = motif.defaultOptionsTitle})

--;===========================================================
--; LOOPS
--;===========================================================
local txt_textinput = main.f_createTextImg(motif.option_info, 'textinput', {defsc = motif.defaultOptions})
local overlay_textinput = main.f_createOverlay(motif.option_info, 'textinput_overlay')

function options.f_displayRatio(value)
	local ret = options.f_precision((value - 1) * 100, '%.01f')
	if ret >= 0 then
		return '+' .. ret .. '%'
	end
	return ret .. '%'
end

motif.languages.languages = {}
for k, _ in pairs(motif.languages) do
	if k ~= "languages" then
		table.insert(motif.languages.languages, k)
	end
end
local function changeLanguageSetting(val)
	sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
	languageCounter = 0
	currentLanguage = -1
	for x, c in ipairs(motif.languages.languages) do
		if c == gameOption('Config.Language') then
			currentLanguage = x
		end
		languageCounter = languageCounter + 1
	end
	if currentLanguage > 0 then
		modifyGameOption('Config.Language', motif.languages.languages[((currentLanguage + val) % languageCounter) + 1])
	else
		modifyGameOption('Config.Language', motif.languages.languages[1] or "en")
	end
	options.modified = true
	options.needReload = true
end

-- Associative elements table storing functions controlling behaviour of each
-- option screen item. Can be appended via external module.
options.t_itemname = {
	--Back
	['back'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			return false
		end
		return true
	end,
	--Port Change
	['portchange'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			local port = main.f_drawInput(
				main.f_extractText(motif.option_info.textinput_port_text),
				txt_textinput,
				overlay_textinput,
				motif.option_info.textinput_offset[2],
				main.f_ySpacing(motif.option_info, 'textinput'),
				motif.optionbgdef
			)
			if tonumber(port) ~= nil then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				modifyGameOption('Netplay.ListenPort', tostring(port))
				t.items[item].vardisplay = gameOption('Netplay.ListenPort')
				options.modified = true
			else
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			end
		end
		return true
	end,
	--Default Values
	['default'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
			--modifyGameOption('Common.Air', {"data/common.air"})
			--modifyGameOption('Common.Cmd', {"data/common.cmd"})
			--modifyGameOption('Common.Const', {"data/common.const"})
			--modifyGameOption('Common.States', {"data/action.zss", "data/dizzy.zss", "data/guardbreak.zss", "data/score.zss", "data/system.zss", "data/tag.zss", "data/training.zss"})
			--modifyGameOption('Common.Fx', {})
			--modifyGameOption('Common.Modules', {})
			--modifyGameOption('Common.Lua', {"loop()"})
			modifyGameOption('Options.Difficulty', 5)
			modifyGameOption('Options.Life', 100)
			modifyGameOption('Options.Time', 99)
			modifyGameOption('Options.GameSpeed', 0)
			modifyGameOption('Options.Match.Wins', 2)
			--modifyGameOption('Options.GameSpeedStep', 5)
			modifyGameOption('Options.Match.MaxDrawGames', -2) -- -2: match.maxdrawgames
			modifyGameOption('Options.Credits', 10)
			modifyGameOption('Options.QuickContinue', false)
			modifyGameOption('Options.AutoGuard', false)
			modifyGameOption('Options.GuardBreak', false)
			modifyGameOption('Options.Dizzy', false)
			modifyGameOption('Options.RedLife', true)
			modifyGameOption('Options.Team.Duplicates', true)
			modifyGameOption('Options.Team.LifeShare', false)
			modifyGameOption('Options.Team.PowerShare', true)
			modifyGameOption('Options.Team.SingleVsTeamLife', 100)
			modifyGameOption('Options.Simul.Min', 2)
			modifyGameOption('Options.Simul.Max', 4)
			modifyGameOption('Options.Simul.Match.Wins', 2)
			modifyGameOption('Options.Simul.LoseOnKO', true)
			modifyGameOption('Options.Tag.Min', 2)
			modifyGameOption('Options.Tag.Max', 4)
			modifyGameOption('Options.Tag.Match.Wins', 2)
			modifyGameOption('Options.Tag.LoseOnKO', false)	
			modifyGameOption('Options.Tag.TimeScaling', 1)
			modifyGameOption('Options.Turns.Min', 2)
			modifyGameOption('Options.Turns.Max', 4)
			modifyGameOption('Options.Turns.Recovery.Base', 0)
			modifyGameOption('Options.Turns.Recovery.Bonus', 20)
			modifyGameOption('Options.Ratio.Recovery.Base', 0)
			modifyGameOption('Options.Ratio.Recovery.Bonus', 20)
			modifyGameOption('Options.Ratio.Level1.Attack', 0.82)
			modifyGameOption('Options.Ratio.Level2.Attack', 1.0)
			modifyGameOption('Options.Ratio.Level3.Attack', 1.17)
			modifyGameOption('Options.Ratio.Level4.Attack', 1.30)
			modifyGameOption('Options.Ratio.Level1.Life', 0.80)
			modifyGameOption('Options.Ratio.Level2.Life', 1.0)
			modifyGameOption('Options.Ratio.Level3.Life', 1.17)
			modifyGameOption('Options.Ratio.Level4.Life', 1.40)
			--modifyGameOption('Config.Motif', "data/system.def")
			modifyGameOption('Config.Players', 4)
			--modifyGameOption('Config.Framerate', 60)
			modifyGameOption('Config.Language', "en")
			modifyGameOption('Config.AfterImageMax', 128)
			modifyGameOption('Config.ExplodMax', 512)
			modifyGameOption('Config.HelperMax', 56)
			modifyGameOption('Config.PlayerProjectileMax', 256)
			--modifyGameOption('Config.ZoomActive', true)
			--modifyGameOption('Config.EscOpensMenu', true)
			--modifyGameOption('Config.BackgroundLoading', false) --TODO: not implemented
			--modifyGameOption('Config.FirstRun', false)
			--modifyGameOption('Config.WindowTitle', "Ikemen GO")
			--modifyGameOption('Config.WindowIcon', {"external/icons/IkemenCylia_256.png", "external/icons/IkemenCylia_96.png", "external/icons/IkemenCylia_48.png"})
			--modifyGameOption('Config.System', "external/script/main.lua")
			--modifyGameOption('Config.ScreenshotFolder', "")
			--modifyGameOption('Config.TrainingChar', "")
			modifyGameOption('Config.GamepadMappings', "external/gamecontrollerdb.txt")
			modifyGameOption('Debug.AllowDebugMode', true)
			modifyGameOption('Debug.AllowDebugKeys', true)
			--modifyGameOption('Debug.ClipboardRows', 2)
			--modifyGameOption('Debug.ConsoleRows', 15)
			--modifyGameOption('Debug.ClsnDarken', true)
			--modifyGameOption('Debug.Font', "font/debug.def")
			--modifyGameOption('Debug.FontScale', 1.0)
			--modifyGameOption('Debug.StartStage', "stages/stage0-720.def")
			--modifyGameOption('Debug.ForceStageZoomout', 0)
			--modifyGameOption('Debug.ForceStageZoomin', 0)
			modifyGameOption('Video.RenderMode', "OpenGL 3.2")
			modifyGameOption('Video.GameWidth', 640)
			modifyGameOption('Video.GameHeight', 480)
			--modifyGameOption('Video.WindowWidth', 0)
			--modifyGameOption('Video.WindowHeight', 0)
			modifyGameOption('Video.Fullscreen', false)
			--modifyGameOption('Video.Borderless', false)
			--modifyGameOption('Video.RGBSpriteBilinearFilter', true)
			modifyGameOption('Video.VSync', 1)
			modifyGameOption('Video.MSAA', 0)
			--modifyGameOption('Video.WindowCentered', true)
			modifyGameOption('Video.ExternalShaders', {})
			modifyGameOption('Video.WindowScaleMode', true)
			modifyGameOption('Video.KeepAspect', true)
			modifyGameOption('Video.EnableModel', true)
			modifyGameOption('Video.EnableModelShadow', true)
			--modifyGameOption('Sound.SampleRate', 44100)
			modifyGameOption('Sound.StereoEffects', true)
			modifyGameOption('Sound.PanningRange', 30)
			--modifyGameOption('Sound.WavChannels', 32)
			modifyGameOption('Sound.MasterVolume', 80)
			--modifyGameOption('Sound.PauseMasterVolume', 100)
			modifyGameOption('Sound.WavVolume', 80)
			modifyGameOption('Sound.BGMVolume', 80)
			--modifyGameOption('Sound.MaxBGMVolume', 100)
			modifyGameOption('Sound.AudioDucking', false)	
			modifyGameOption('Arcade.AI.RandomColor', false)
			modifyGameOption('Arcade.AI.SurvivalColor', true)
			modifyGameOption('Arcade.AI.Ramping', true)
			modifyGameOption('Netplay.ListenPort', "7500")
			--modifyGameOption('Netplay.IP.<>', "")
			--modifyGameOption('Input.ButtonAssist', true)
			--modifyGameOption('Input.SOCDResolution', 4)
			--modifyGameOption('Input.ControllerStickSensitivity', 0.4)
			--modifyGameOption('Input.XinputTriggerSensitivity', 0.5)

			loadLifebar(motif.files.fight)
			main.timeFramesPerCount = fightscreenvar("time.framespercount")
			main.f_updateRoundsNum()
			main.f_setPlayers()
			for _, v in ipairs(options.t_vardisplayPointers) do
				v.vardisplay = options.f_vardisplay(v.itemname)
			end
			toggleFullscreen(gameOption('Video.Fullscreen'))
			toggleVSync(gameOption('Video.VSync'))
			updateVolume()
			options.modified = true
			options.needReload = true
		end
		return true
	end,
	--Difficulty Level
	['difficulty'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Difficulty') < 8 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Difficulty', gameOption('Options.Difficulty') + 1)
			t.items[item].vardisplay = gameOption('Options.Difficulty')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Difficulty') > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Difficulty', gameOption('Options.Difficulty') - 1)
			t.items[item].vardisplay = gameOption('Options.Difficulty')
			options.modified = true
		end
		return true
	end,
	--Time Limit
	['roundtime'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Time') < 1000 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Time', gameOption('Options.Time') + 1)
			t.items[item].vardisplay = gameOption('Options.Time')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Time') > -1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Time', gameOption('Options.Time') - 1)
			t.items[item].vardisplay = options.f_definedDisplay(gameOption('Options.Time'), {[-1] = motif.option_info.menu_valuename_none}, gameOption('Options.Time'))
			options.modified = true
		end
		return true
	end,
	--Language Setting
	['language'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) then
			changeLanguageSetting(0)
			t.items[item].vardisplay = motif.languages[gameOption('Config.Language')] or gameOption('Config.Language')
		elseif main.f_input(main.t_players, {'$B'}) then
			changeLanguageSetting(-2)
			t.items[item].vardisplay = motif.languages[gameOption('Config.Language')] or gameOption('Config.Language')
		end
		return true
	end,
	--Life
	['lifemul'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Life') < 300 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Life', gameOption('Options.Life') + 10)
			t.items[item].vardisplay = gameOption('Options.Life') .. '%'
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Life') > 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Life', gameOption('Options.Life') - 10)
			t.items[item].vardisplay = gameOption('Options.Life') .. '%'
			options.modified = true
		end
		return true
	end,
	--Single VS Team Life
	['singlevsteamlife'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Team.SingleVsTeamLife') < 300 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Team.SingleVsTeamLife', gameOption('Options.Team.SingleVsTeamLife') + 10)
			t.items[item].vardisplay = gameOption('Options.Team.SingleVsTeamLife') .. '%'
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Team.SingleVsTeamLife') > 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Team.SingleVsTeamLife', gameOption('Options.Team.SingleVsTeamLife') - 10)
			t.items[item].vardisplay = gameOption('Options.Team.SingleVsTeamLife') .. '%'
			options.modified = true
		end
		return true
	end,
	-- Game Speed
	['gamespeed'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.GameSpeed') < 9 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.GameSpeed', gameOption('Options.GameSpeed') + 1)
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Options.GameSpeed') == 0, motif.option_info.menu_valuename_normal, options.f_boolDisplay(gameOption('Options.GameSpeed') < 0, motif.option_info.menu_valuename_slow:gsub('%%i', tostring(0-gameOption('Options.GameSpeed'))), motif.option_info.menu_valuename_fast:gsub('%%i', tostring(gameOption('Options.GameSpeed')))))
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.GameSpeed') > -9 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.GameSpeed', gameOption('Options.GameSpeed') - 1)
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Options.GameSpeed') == 0, motif.option_info.menu_valuename_normal, options.f_boolDisplay(gameOption('Options.GameSpeed') < 0, motif.option_info.menu_valuename_slow:gsub('%%i', tostring(0-gameOption('Options.GameSpeed'))), motif.option_info.menu_valuename_fast:gsub('%%i', tostring(gameOption('Options.GameSpeed')))))
			options.modified = true
		end
		return true
	end,
	--Rounds to Win (Single)
	['roundsnumsingle'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and main.roundsNumSingle[1] < 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Match.Wins', main.roundsNumSingle[1] + 1)
			main.roundsNumSingle = {gameOption('Options.Match.Wins'), gameOption('Options.Match.Wins')}
			t.items[item].vardisplay = gameOption('Options.Match.Wins')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and main.roundsNumSingle[1] > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Match.Wins', main.roundsNumSingle[1] - 1)
			main.roundsNumSingle = {gameOption('Options.Match.Wins'), gameOption('Options.Match.Wins')}
			t.items[item].vardisplay = gameOption('Options.Match.Wins')
			options.modified = true
		end
		return true
	end,
	--Max Draw Games
	['maxdrawgames'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and main.maxDrawGames[1] < 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Match.MaxDrawGames', main.maxDrawGames[1] + 1)
			main.maxDrawGames = {gameOption('Options.Match.MaxDrawGames'), gameOption('Options.Match.MaxDrawGames')}
			t.items[item].vardisplay = gameOption('Options.Match.MaxDrawGames')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and main.maxDrawGames[1] > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Match.MaxDrawGames', main.maxDrawGames[1] - 1)
			main.maxDrawGames = {gameOption('Options.Match.MaxDrawGames'), gameOption('Options.Match.MaxDrawGames')}
			t.items[item].vardisplay = gameOption('Options.Match.MaxDrawGames')
			options.modified = true
		end
		return true
	end,
	--Credits
	['credits'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Credits') < 99 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Credits', gameOption('Options.Credits') + 1)
			t.items[item].vardisplay = gameOption('Options.Credits')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Credits') > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Credits', gameOption('Options.Credits') - 1)
			t.items[item].vardisplay = options.f_definedDisplay(gameOption('Options.Credits'), {[0] = motif.option_info.menu_valuename_disabled}, gameOption('Options.Credits'))
			options.modified = true
		end
		return true
	end,
	--Arcade Palette
	['aipalette'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Arcade.AI.RandomColor') then
				modifyGameOption('Arcade.AI.RandomColor', false)
			else
				modifyGameOption('Arcade.AI.RandomColor', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Arcade.AI.RandomColor'), motif.option_info.menu_valuename_random, motif.option_info.menu_valuename_default)
			options.modified = true
		end
		return true
	end,
	--Survival Palette
	['aisurvivalpalette'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Arcade.AI.SurvivalColor') then
				modifyGameOption('Arcade.AI.SurvivalColor', false)
			else
				modifyGameOption('Arcade.AI.SurvivalColor', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Arcade.AI.SurvivalColor'), motif.option_info.menu_valuename_random, motif.option_info.menu_valuename_default)
			options.modified = true
		end
		return true
	end,
	--AI Ramping
	['airamping'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Arcade.AI.Ramping') then
				modifyGameOption('Arcade.AI.Ramping', false)
			else
				modifyGameOption('Arcade.AI.Ramping', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Arcade.AI.Ramping'))
			options.modified = true
		end
		return true
	end,
	--Quick Continue
	['quickcontinue'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Options.QuickContinue') then
				modifyGameOption('Options.QuickContinue', false)
			else
				modifyGameOption('Options.QuickContinue', true)
				end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Options.QuickContinue'))
			options.modified = true
		end
		return true
	end,
	--Auto-Guard
	['autoguard'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Options.AutoGuard') then
				modifyGameOption('Options.AutoGuard', false)
			else
				modifyGameOption('Options.AutoGuard', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Options.AutoGuard'))
			options.modified = true
		end
		return true
	end,
	--Dizzy
	['dizzy'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Options.Dizzy') then
				modifyGameOption('Options.Dizzy', false)
			else
				modifyGameOption('Options.Dizzy', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Options.Dizzy'))
			options.modified = true
		end
		return true
	end,
	--Guard Break
	['guardbreak'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Options.GuardBreak') then
				modifyGameOption('Options.GuardBreak', false)
			else
				modifyGameOption('Options.GuardBreak', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Options.GuardBreak'))
			options.modified = true
		end
		return true
	end,
	--Red Life
	['redlife'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Options.RedLife') then
				modifyGameOption('Options.RedLife', false)
			else
				modifyGameOption('Options.RedLife', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Options.RedLife'))
			options.modified = true
		end
		return true
	end,
	--Team Duplicates
	['teamduplicates'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Options.Team.Duplicates') then
				modifyGameOption('Options.Team.Duplicates', false)
			else
				modifyGameOption('Options.Team.Duplicates', true)
				end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Options.Team.Duplicates'))
			options.modified = true
		end
		return true
	end,
	--Team Life Share
	['teamlifeshare'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Options.Team.LifeShare') then
				modifyGameOption('Options.Team.LifeShare', false)
			else
				modifyGameOption('Options.Team.LifeShare', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Options.Team.LifeShare'))
			options.modified = true
		end
		return true
	end,
	--Team Power Share
	['teampowershare'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Options.Team.PowerShare') then
				modifyGameOption('Options.Team.PowerShare', false)
			else
				modifyGameOption('Options.Team.PowerShare', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Options.Team.PowerShare'))
			options.modified = true
		end
		return true
	end,
	--Rounds to Win (Tag)
	['roundsnumtag'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and main.roundsNumTag[1] < 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Tag.Match.Wins', main.roundsNumTag[1] + 1)
			main.roundsNumTag = {gameOption('Options.Tag.Match.Wins'), gameOption('Options.Tag.Match.Wins')}
			t.items[item].vardisplay = gameOption('Options.Tag.Match.Wins')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and main.roundsNumTag[1] > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Tag.Match.Wins', main.roundsNumTag[1] - 1)
			main.roundsNumTag = {gameOption('Options.Tag.Match.Wins'), gameOption('Options.Tag.Match.Wins')}
			t.items[item].vardisplay = gameOption('Options.Tag.Match.Wins')
			options.modified = true
		end
		return true
	end,
	--Partner KOed Lose
	['losekotag'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Options.Tag.LoseOnKO') then
				modifyGameOption('Options.Tag.LoseOnKO', false)
			else
				modifyGameOption('Options.Tag.LoseOnKO', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Options.Tag.LoseOnKO'))
			options.modified = true
		end
		return true
	end,
	--Min Tag Chars
	['mintag'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Tag.Min') < gameOption('Options.Tag.Max') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Tag.Min', gameOption('Options.Tag.Min') + 1)
			t.items[item].vardisplay = gameOption('Options.Tag.Min')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Tag.Min') > 2 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Tag.Min', gameOption('Options.Tag.Min') - 1)
			t.items[item].vardisplay = gameOption('Options.Tag.Min')
			options.modified = true
		end
		return true
	end,
	--Max Tag Chars
	['maxtag'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Tag.Max') < 4 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Tag.Max', gameOption('Options.Tag.Max') + 1)
			t.items[item].vardisplay = gameOption('Options.Tag.Max')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Tag.Max') > gameOption('Options.Tag.Min') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Tag.Max', gameOption('Options.Tag.Max') - 1)
			t.items[item].vardisplay = gameOption('Options.Tag.Max')
			options.modified = true
		end
		return true
	end,
	--Rounds to Win (Simul)
	['roundsnumsimul'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and main.roundsNumSimul[1] < 10 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Simul.Match.Wins', main.roundsNumSimul[1] + 1)
			main.roundsNumSimul = {gameOption('Options.Simul.Match.Wins'), gameOption('Options.Simul.Match.Wins')}
			t.items[item].vardisplay = gameOption('Options.Simul.Match.Wins')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and main.roundsNumSimul[1] > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Simul.Match.Wins', main.roundsNumSimul[1] - 1)
			main.roundsNumSimul = {gameOption('Options.Simul.Match.Wins'), gameOption('Options.Simul.Match.Wins')}
			t.items[item].vardisplay = gameOption('Options.Simul.Match.Wins')
			options.modified = true
		end
		return true
	end,
	--Simul Player KOed Lose
	['losekosimul'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Options.Simul.LoseOnKO') then
				modifyGameOption('Options.Simul.LoseOnKO', false)
			else
				modifyGameOption('Options.Simul.LoseOnKO', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Options.Simul.LoseOnKO'))
			options.modified = true
		end
		return true
	end,
	--Min Simul Chars
	['minsimul'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Simul.Min') < gameOption('Options.Simul.Max') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Simul.Min', gameOption('Options.Simul.Min') + 1)
			t.items[item].vardisplay = gameOption('Options.Simul.Min')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Simul.Min') > 2 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Simul.Min', gameOption('Options.Simul.Min') - 1)
			t.items[item].vardisplay = gameOption('Options.Simul.Min')
			options.modified = true
		end
		return true
	end,
	--Max Simul Chars
	['maxsimul'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Simul.Max') < 4 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Simul.Max', gameOption('Options.Simul.Max') + 1)
			t.items[item].vardisplay = gameOption('Options.Simul.Max')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Simul.Max') > gameOption('Options.Simul.Min') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Simul.Max', gameOption('Options.Simul.Max') - 1)
			t.items[item].vardisplay = gameOption('Options.Simul.Max')
			options.modified = true
		end
		return true
	end,
	--Turns Recovery Base
	['turnsrecoverybase'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Turns.Recovery.Base') < 100 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Turns.Recovery.Base', gameOption('Options.Turns.Recovery.Base') + 0.5)
			t.items[item].vardisplay = gameOption('Options.Turns.Recovery.Base') .. '%'
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Turns.Recovery.Base') > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Turns.Recovery.Base', gameOption('Options.Turns.Recovery.Base') - 0.5)
			t.items[item].vardisplay = gameOption('Options.Turns.Recovery.Base') .. '%'
			options.modified = true
		end
		return true
	end,
	--Turns Recovery Bonus
	['turnsrecoverybonus'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Turns.Recovery.Bonus') < 100 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Turns.Recovery.Bonus', gameOption('Options.Turns.Recovery.Bonus') + 0.5)
			t.items[item].vardisplay = gameOption('Options.Turns.Recovery.Bonus') .. '%'
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Turns.Recovery.Bonus') > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Turns.Recovery.Bonus', gameOption('Options.Turns.Recovery.Bonus') - 0.5)
			t.items[item].vardisplay = gameOption('Options.Turns.Recovery.Bonus') .. '%'
			options.modified = true
		end
		return true
	end,
	--Min Turns Chars
	['minturns'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Turns.Min') < gameOption('Options.Turns.Max') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Turns.Min', gameOption('Options.Turns.Min') + 1)
			t.items[item].vardisplay = gameOption('Options.Turns.Min')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Turns.Min') > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Turns.Min', gameOption('Options.Turns.Min') - 1)
			t.items[item].vardisplay = gameOption('Options.Turns.Min')
			options.modified = true
		end
		return true
	end,
	--Max Turns Chars
	['maxturns'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Turns.Max') < 8 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Turns.Max', gameOption('Options.Turns.Max') + 1)
			t.items[item].vardisplay = gameOption('Options.Turns.Max')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Turns.Max') > gameOption('Options.Turns.Min') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Turns.Max', gameOption('Options.Turns.Max') - 1)
			t.items[item].vardisplay = gameOption('Options.Turns.Max')
			options.modified = true
		end
		return true
	end,
	--Ratio Recovery Base
	['ratiorecoverybase'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Ratio.Recovery.Base') < 100 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Ratio.Recovery.Base', gameOption('Options.Ratio.Recovery.Base') + 0.5)
			t.items[item].vardisplay = gameOption('Options.Ratio.Recovery.Base') .. '%'
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Ratio.Recovery.Base') > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Ratio.Recovery.Base', gameOption('Options.Ratio.Recovery.Base') - 0.5)
			t.items[item].vardisplay = gameOption('Options.Ratio.Recovery.Base') .. '%'
			options.modified = true
		end
		return true
	end,
	--Ratio Recovery Bonus
	['ratiorecoverybonus'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Options.Ratio.Recovery.Bonus') < 100 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Ratio.Recovery.Bonus', gameOption('Options.Ratio.Recovery.Bonus') + 0.5)
			t.items[item].vardisplay = gameOption('Options.Ratio.Recovery.Bonus') .. '%'
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Options.Ratio.Recovery.Bonus') > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Options.Ratio.Recovery.Bonus', gameOption('Options.Ratio.Recovery.Bonus') - 0.5)
			t.items[item].vardisplay = gameOption('Options.Ratio.Recovery.Bonus') .. '%'
			options.modified = true
		end
		return true
	end,
	--Renderer (submenu)
	['renderer'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			for k, v in ipairs(t.submenu[t.items[item].itemname].items) do
				if gameOption('Video.RenderMode') == v.itemname then
					v.selected = true
				else
					v.selected = false
				end
			end
			t.submenu[t.items[item].itemname].loop()
			t.items[item].vardisplay = gameOption('Video.RenderMode')
			options.modified = true
			options.needReload = true
		end
		return true
	end,
	--gl32
	['gl32'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Video.RenderMode', "OpenGL 3.2")
			options.modified = true
			options.needReload = true
			return false
		end
		return true
	end,
	--gl21
	['gl21'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Video.RenderMode', "OpenGL 2.1")
			options.modified = true
			options.needReload = true
			return false
		end
		return true
	end,
	--Resolution (submenu)
	['resolution'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			local t_pos = {}
			local ok = false
			for k, v in ipairs(t.submenu[t.items[item].itemname].items) do
				local width, height = v.itemname:match('^([0-9]+)x([0-9]+)$')
				if tonumber(width) == gameOption('Video.GameWidth') and tonumber(height) == gameOption('Video.GameHeight') then
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
			t.items[item].vardisplay = gameOption('Video.GameWidth') .. 'x' .. gameOption('Video.GameHeight')
		end
		return true
	end,
	--Custom resolution
	['customres'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			local width = tonumber(main.f_drawInput(
				main.f_extractText(motif.option_info.textinput_reswidth_text),
					txt_textinput,
					overlay_textinput,
					motif.option_info.textinput_offset[2],
					main.f_ySpacing(motif.option_info, 'textinput'),
					motif.optionbgdef
				))
			if width ~= nil then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				local height = tonumber(main.f_drawInput(
					main.f_extractText(motif.option_info.textinput_resheight_text),
					txt_textinput,
					overlay_textinput,
					motif.option_info.textinput_offset[2],
					main.f_ySpacing(motif.option_info, 'textinput'),
					motif.optionbgdef
				))
				if height ~= nil then
					modifyGameOption('Video.GameWidth', width)
					modifyGameOption('Video.GameHeight', height)
					sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
					options.modified = true
					options.needReload = true
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
	--Fullscreen
	['fullscreen'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Video.Fullscreen') then
				modifyGameOption('Video.Fullscreen', false)
			else
				modifyGameOption('Video.Fullscreen', true)
			end
			toggleFullscreen(gameOption('Video.Fullscreen'))
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Video.Fullscreen'))
			options.modified = true
		end
		return true
	end,
	--VSync
	['vsync'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Video.VSync') == 1 then
				modifyGameOption('Video.VSync', 0)
			else
				modifyGameOption('Video.VSync', 1)
			end
			toggleVSync(gameOption('Video.VSync'))
			t.items[item].vardisplay = options.f_definedDisplay(gameOption('Video.VSync'), {[1] = motif.option_info.menu_valuename_enabled}, motif.option_info.menu_valuename_disabled)
			options.modified = true
		end
		return true
	end,
	--MSAA
	['msaa'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Video.MSAA') < 32 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Video.MSAA') == 0 then
				modifyGameOption('Video.MSAA', 2)
			else
				modifyGameOption('Video.MSAA', gameOption('Video.MSAA') * 2)
			end
			t.items[item].vardisplay = gameOption('Video.MSAA') .. 'x'
			options.modified = true
			options.needReload = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Video.MSAA') > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Video.MSAA') == 2 then
				modifyGameOption('Video.MSAA', 0)
			else
				modifyGameOption('Video.MSAA', gameOption('Video.MSAA') / 2)
			end
			t.items[item].vardisplay = options.f_definedDisplay(gameOption('Video.MSAA'), {[0] = motif.option_info.menu_valuename_disabled}, gameOption('Video.MSAA') .. 'x')
			options.modified = true
			options.needReload = true
		end
		return true
	end,
	--Window scaling mode
	['windowscalemode'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Video.WindowScaleMode') then
				modifyGameOption('Video.WindowScaleMode', false)
			else
				modifyGameOption('Video.WindowScaleMode', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['windowscalemode']()
			options.modified = true
		end
		return true
	end,
	--Keep Aspect Ratio
	['keepaspect'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Video.KeepAspect') then
				modifyGameOption('Video.KeepAspect', false)
			else
				modifyGameOption('Video.KeepAspect', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Video.KeepAspect'), motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
			options.modified = true
		end
		return true
	end,
	--Shaders (submenu)
	['shaders'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if #options.t_shaders == 0 then
				main.f_warning(main.f_extractText(motif.warning_info.text_shaders_text), motif.optionbgdef)
				return true
			end
			t.submenu[t.items[item].itemname].loop()
			t.items[item].vardisplay = options.f_boolDisplay(#gameOption('Video.ExternalShaders') > 0, motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
			options.modified = true
			options.needReload = true
		end
		return true
	end,
	--Disable (shader)
	['noshader'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			modifyGameOption('Video.ExternalShaders', {})
			options.modified = true
			options.needReload = true
			return false
		end
		return true
	end,
	--Enable Model
	['enablemodel'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Video.EnableModel') then
				modifyGameOption('Video.EnableModel', false)
			else
				modifyGameOption('Video.EnableModel', true)
			end
			t.items[item].vardisplay = options.f_definedDisplay(gameOption('Video.EnableModel'), {[true] = motif.option_info.menu_valuename_enabled}, motif.option_info.menu_valuename_disabled)
			options.modified = true
			options.needReload = true
		end
		return true
	end,
	--Enable Model Shadow
	['enablemodelshadow'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Video.EnableModelShadow') then
				modifyGameOption('Video.EnableModelShadow', false)
			else
				modifyGameOption('Video.EnableModelShadow', true)
			end
			t.items[item].vardisplay = options.f_definedDisplay(gameOption('Video.EnableModelShadow'), {[true] = motif.option_info.menu_valuename_enabled}, motif.option_info.menu_valuename_disabled)
			options.modified = true
			options.needReload = true
		end
		return true
	end,
	--Master Volume
	['mastervolume'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Sound.MasterVolume') < 200 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Sound.MasterVolume', gameOption('Sound.MasterVolume') + 1)
			t.items[item].vardisplay = gameOption('Sound.MasterVolume') .. '%'
			options.modified = true
			updateVolume()
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Sound.MasterVolume') > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Sound.MasterVolume', gameOption('Sound.MasterVolume') - 1)
			t.items[item].vardisplay = gameOption('Sound.MasterVolume')  .. '%'
			options.modified = true
			updateVolume()
		end
		return true
	end,
	--BGM Volume
	['bgmvolume'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Sound.BGMVolume') < 100 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Sound.BGMVolume', gameOption('Sound.BGMVolume') + 1)
			t.items[item].vardisplay = gameOption('Sound.BGMVolume') .. '%'
			options.modified = true
			updateVolume()
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Sound.BGMVolume') > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Sound.BGMVolume', gameOption('Sound.BGMVolume') - 1)
			t.items[item].vardisplay = gameOption('Sound.BGMVolume') .. '%'
			options.modified = true
			updateVolume()
		end
		return true
	end,
	--SFX Volume
	['sfxvolume'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Sound.WavVolume') < 100 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Sound.WavVolume', gameOption('Sound.WavVolume') + 1)
			t.items[item].vardisplay = gameOption('Sound.WavVolume') .. '%'
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Sound.WavVolume') > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Sound.WavVolume', gameOption('Sound.WavVolume') - 1)
			t.items[item].vardisplay = gameOption('Sound.WavVolume') .. '%'
			options.modified = true
		end
		return true
	end,
	--Audio Ducking
	['audioducking'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Sound.AudioDucking') then
				modifyGameOption('Sound.AudioDucking', false)
			else
				modifyGameOption('Sound.AudioDucking', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Sound.AudioDucking'), motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
			options.modified = true
		end
		return true
	end,
	--Stereo Effects
	['stereoeffects'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Sound.StereoEffects') then
				modifyGameOption('Sound.StereoEffects', false)
			else
				modifyGameOption('Sound.StereoEffects', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Sound.StereoEffects'), motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
			options.modified = true
		end
		return true
	end,
	--Panning Range
	['panningrange'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Sound.PanningRange') < 100 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Sound.PanningRange', gameOption('Sound.PanningRange') + 1)
			t.items[item].vardisplay = gameOption('Sound.PanningRange') .. '%'
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Sound.PanningRange') > 0 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Sound.PanningRange', gameOption('Sound.PanningRange') - 1)
			t.items[item].vardisplay = gameOption('Sound.PanningRange') .. '%'
			options.modified = true
		end
		return true
	end,
	--Key Config
	['keyboard'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'pal', 's'}) --[[or getKey():match('^F[0-9]+$')]] then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
			options.f_keyCfgInit('Keys', t.submenu[t.items[item].itemname].title)
			while true do
				if not options.f_keyCfg('Keys', t.items[item].itemname, 'optionbgdef', false) then
					break
				end
			end
		end
		return true
	end,
	--Joystick Config
	['gamepad'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'pal', 's'}) --[[or getKey():match('^F[0-9]+$')]] then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
			if main.flags['-nojoy'] == nil then
				options.f_keyCfgInit('Joystick', t.submenu[t.items[item].itemname].title)
				while true do
					if not options.f_keyCfg('Joystick', t.items[item].itemname, 'optionbgdef', false) then
						break
					end
				end
			end
		end
		return true
	end,
	--Default
	['inputdefault'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
			options.f_keyDefault()
			options.f_setKeyConfig('Keys')
			if main.flags['-nojoy'] == nil then
				options.f_setKeyConfig('Joystick')
			end
			options.modified = true
		end
		return true
	end,
	--Players
	['players'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) and gameOption('Config.Players') < 8 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Config.Players', math.min(8, gameOption('Config.Players') + 2))
			t.items[item].vardisplay = gameOption('Config.Players')
			main.f_setPlayers()
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Config.Players') > 2 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Config.Players', math.max(2, gameOption('Config.Players') - 2))
			t.items[item].vardisplay = gameOption('Config.Players')
			main.f_setPlayers()
			options.modified = true
		end
		return true
	end,
	--Debug Keys
	['debugkeys'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Debug.AllowDebugKeys') then
				modifyGameOption('Debug.AllowDebugKeys', false)
			else
				modifyGameOption('Debug.AllowDebugKeys', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Debug.AllowDebugKeys'), motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
			options.modified = true
		end
		return true
	end,
	--Debug Mode
	['debugmode'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Debug.AllowDebugMode') then
				modifyGameOption('Debug.AllowDebugMode', false)
			else
				modifyGameOption('Debug.AllowDebugMode', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Debug.AllowDebugMode'), motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
			options.modified = true
		end
		return true
	end,
	--Background Loading
	--[[['backgroundloading'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if gameOption('Config.BackgroundLoading') then
				modifyGameOption('Config.BackgroundLoading', false)
			else
				modifyGameOption('Config.BackgroundLoading', true)
			end
			t.items[item].vardisplay = options.f_boolDisplay(gameOption('Config.BackgroundLoading'), motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
			options.modified = true
		end
		return true
	end,]]
	--HelperMax
	['helpermax'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Config.HelperMax', gameOption('Config.HelperMax') + 1)
			t.items[item].vardisplay = gameOption('Config.HelperMax')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Config.HelperMax') > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Config.HelperMax', gameOption('Config.HelperMax') - 1)
			t.items[item].vardisplay = gameOption('Config.HelperMax')
			options.modified = true
		end
		return true
	end,
	--PlayerProjectileMax
	['projectilemax'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Config.PlayerProjectileMax', gameOption('Config.PlayerProjectileMax') + 1)
			t.items[item].vardisplay = gameOption('Config.PlayerProjectileMax')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Config.PlayerProjectileMax') > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Config.PlayerProjectileMax', gameOption('Config.PlayerProjectileMax') - 1)
			t.items[item].vardisplay = gameOption('Config.PlayerProjectileMax')
			options.modified = true
		end
		return true
	end,
	--ExplodMax
	['explodmax'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Config.ExplodMax', gameOption('Config.ExplodMax') + 1)
			t.items[item].vardisplay = gameOption('Config.ExplodMax')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Config.ExplodMax') > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Config.ExplodMax', gameOption('Config.ExplodMax') - 1)
			t.items[item].vardisplay = gameOption('Config.ExplodMax')
			options.modified = true
		end
		return true
	end,
	--AfterImageMax
	['afterimagemax'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Config.AfterImageMax', gameOption('Config.AfterImageMax') + 1)
			t.items[item].vardisplay = gameOption('Config.AfterImageMax')
			options.modified = true
		elseif main.f_input(main.t_players, {'$B'}) and gameOption('Config.AfterImageMax') > 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			modifyGameOption('Config.AfterImageMax', gameOption('Config.AfterImageMax') - 1)
			t.items[item].vardisplay = gameOption('Config.AfterImageMax')
			options.modified = true
		end
		return true
	end,
	--Save and Return
	['savereturn'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			if options.modified then
				options.f_saveCfg(options.needReload)
			end
			main.close = true
			--return false
		end
		return true
	end,
	--Return Without Saving
	['return'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			if options.needReload then
				main.f_warning(main.f_extractText(motif.warning_info.text_noreload_text), motif.optionbgdef)
			end
			main.close = true
			--return false
		end
		return true
	end,
	--Save Settings
	['savesettings'] = function(t, item, cursorPosY, moveTxt)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
			if options.modified then
				options.f_saveCfg(options.needReload)
			end
		end
		return true
	end,
}

-- Shared menu loop logic
function options.f_createMenu(tbl, bool_main)
	return function()
		hook.run("options.menu.loop")
		local cursorPosY = 1
		local moveTxt = 0
		local item = 1
		local t = tbl.items
		main.f_menuSnap('option_info')
		if bool_main then
			main.f_bgReset(motif.optionbgdef.bg)
			main.f_fadeReset('fadein', motif.option_info)
			if motif.music.option_bgm ~= '' then
				main.f_playBGM(false, motif.music.option_bgm, motif.music.option_bgm_loop, motif.music.option_bgm_volume, motif.music.option_bgm_loopstart, motif.music.option_bgm_loopend)
			end
			main.close = false
		end
		while true do
			if tbl.reset then
				tbl.reset = false
				main.f_cmdInput()
			else
				main.f_menuCommonDraw(t, item, cursorPosY, moveTxt, 'option_info', 'optionbgdef', options.txt_title, motif.defaultOptions, {})
			end
			cursorPosY, moveTxt, item = main.f_menuCommonCalc(t, item, cursorPosY, moveTxt, 'option_info', {'$U'}, {'$D'})
			options.txt_title:update({text = tbl.title})
			if main.close and not main.fadeActive then
				main.f_bgReset(motif[main.background].bg)
				main.f_fadeReset('fadein', motif[main.group])
				main.f_playBGM(false, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
				main.close = false
				break
			elseif esc() or main.f_input(main.t_players, {'m'}) then
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				if bool_main then
					if options.modified then
						--options.f_saveCfg(options.needReload)
					end
					if options.needReload then
						main.f_warning(main.f_extractText(motif.warning_info.text_noreload_text), motif.optionbgdef)
					end
					main.f_fadeReset('fadeout', motif.option_info)
					main.close = true
				else
					break
				end
			elseif options.t_itemname[t[item].itemname] ~= nil then
				if not options.t_itemname[t[item].itemname](tbl, item, cursorPosY, moveTxt) then
					break
				end
			elseif main.f_input(main.t_players, {'pal', 's'}) then
				local f = t[item].itemname
				if tbl.submenu[f].loop ~= nil then
					sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
					tbl.submenu[f].loop()
					main.f_menuSnap('option_info')
				elseif not options.t_itemname[f](tbl, item, cursorPosY, moveTxt) then
					break
				end
			end
		end
	end
end

options.t_vardisplayPointers = {}

-- Associative elements table storing functions returning current setting values
-- rendered alongside menu item name. Can be appended via external module.
options.t_vardisplay = {
	['afterimagemax'] = function()
		return gameOption('Config.AfterImageMax')
	end,
	['aipalette'] = function()
		return options.f_boolDisplay(gameOption('Arcade.AI.RandomColor'), motif.option_info.menu_valuename_random, motif.option_info.menu_valuename_default)
	end,
	['aisurvivalpalette'] = function()
		return options.f_boolDisplay(gameOption('Arcade.AI.SurvivalColor'), motif.option_info.menu_valuename_random, motif.option_info.menu_valuename_default)
	end,
	['airamping'] = function()
		return options.f_boolDisplay(gameOption('Arcade.AI.Ramping'))
	end,
	['audioducking'] = function()
		return options.f_boolDisplay(gameOption('Sound.AudioDucking'), motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
	end,
	['autoguard'] = function()
		return options.f_boolDisplay(gameOption('Options.AutoGuard'))
	end,
	--['backgroundloading'] = function()
	--	return options.f_boolDisplay(gameOption('Config.BackgroundLoading'), motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
	--end,
	['bgmvolume'] = function()
		return gameOption('Sound.BGMVolume') .. '%'
	end,
	['credits'] = function()
		return options.f_definedDisplay(gameOption('Options.Credits'), {[0] = motif.option_info.menu_valuename_disabled}, gameOption('Options.Credits'))
	end,
	['debugkeys'] = function()
		return options.f_boolDisplay(gameOption('Debug.AllowDebugKeys'), motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
	end,
	['debugmode'] = function()
		return options.f_boolDisplay(gameOption('Debug.AllowDebugMode'), motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
	end,
	['difficulty'] = function()
		return gameOption('Options.Difficulty')
	end,
	['enablemodel'] = function()
		return options.f_definedDisplay(gameOption('Video.EnableModel'), {[true] = motif.option_info.menu_valuename_enabled}, motif.option_info.menu_valuename_disabled)
	end,
	['enablemodelshadow'] = function()
		return options.f_definedDisplay(gameOption('Video.EnableModelShadow'), {[true] = motif.option_info.menu_valuename_enabled}, motif.option_info.menu_valuename_disabled)
	end,
	['explodmax'] = function()
		return gameOption('Config.ExplodMax')
	end,
	['fullscreen'] = function()
		return options.f_boolDisplay(gameOption('Video.Fullscreen'))
	end,
	['gamespeed'] = function()
		return options.f_boolDisplay(gameOption('Options.GameSpeed') == 0, motif.option_info.menu_valuename_normal, options.f_boolDisplay(gameOption('Options.GameSpeed') < 0, motif.option_info.menu_valuename_slow:gsub('%%i', tostring(0-gameOption('Options.GameSpeed'))), motif.option_info.menu_valuename_fast:gsub('%%i', tostring(gameOption('Options.GameSpeed')))))
	end,
	['guardbreak'] = function()
		return options.f_boolDisplay(gameOption('Options.GuardBreak'))
	end,
	['helpermax'] = function()
		return gameOption('Config.HelperMax')
	end,
	['keepaspect'] = function()
		return options.f_boolDisplay(gameOption('Video.KeepAspect'), motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
	end,
	['language'] = function()
		sfs = motif.languages[gameOption('Config.Language')]
		return sfs or gameOption('Config.Language')
	end,
	['lifemul'] = function()
		return gameOption('Options.Life') .. '%'
	end,
	['losekosimul'] = function()
		return options.f_boolDisplay(gameOption('Options.Simul.LoseOnKO'))
	end,
	['losekotag'] = function()
		return options.f_boolDisplay(gameOption('Options.Tag.LoseOnKO'))
	end,
	['mastervolume'] = function()
		return gameOption('Sound.MasterVolume') .. '%'
	end,
	['maxdrawgames'] = function()
		return main.maxDrawGames[1]
	end,
	['maxsimul'] = function()
		return gameOption('Options.Simul.Max')
	end,
	['maxtag'] = function()
		return gameOption('Options.Tag.Max')
	end,
	['maxturns'] = function()
		return gameOption('Options.Turns.Max')
	end,
	['minsimul'] = function()
		return gameOption('Options.Simul.Min')
	end,
	['mintag'] = function()
		return gameOption('Options.Tag.Min')
	end,
	['minturns'] = function()
		return gameOption('Options.Turns.Min')
	end,
	['msaa'] = function()
		return options.f_definedDisplay(gameOption('Video.MSAA'), {[0] = motif.option_info.menu_valuename_disabled}, gameOption('Video.MSAA') .. 'x')
	end,
	['panningrange'] = function()
		return gameOption('Sound.PanningRange') .. '%'
	end,
	['players'] = function()
		return gameOption('Config.Players')
	end,
	['portchange'] = function()
		return gameOption('Netplay.ListenPort')
	end,
	['projectilemax'] = function()
		return gameOption('Config.PlayerProjectileMax')
	end,
	['quickcontinue'] = function()
		return options.f_boolDisplay(gameOption('Options.QuickContinue'))
	end,
	['ratio1attack'] = function()
		return options.f_displayRatio(gameOption('Options.Ratio.Level1.Attack'))
	end,
	['ratio1life'] = function()
		return options.f_displayRatio(gameOption('Options.Ratio.Level1.Life'))
	end,
	['ratio2attack'] = function()
		return options.f_displayRatio(gameOption('Options.Ratio.Level2.Attack'))
	end,
	['ratio2life'] = function()
		return options.f_displayRatio(gameOption('Options.Ratio.Level2.Life'))
	end,
	['ratio3attack'] = function()
		return options.f_displayRatio(gameOption('Options.Ratio.Level3.Attack'))
	end,
	['ratio3life'] = function()
		return options.f_displayRatio(gameOption('Options.Ratio.Level3.Life'))
	end,
	['ratio4attack'] = function()
		return options.f_displayRatio(gameOption('Options.Ratio.Level4.Attack'))
	end,
	['ratio4life'] = function()
		return options.f_displayRatio(gameOption('Options.Ratio.Level4.Life'))
	end,
	['ratiorecoverybase'] = function()
		return gameOption('Options.Ratio.Recovery.Base') .. '%'
	end,
	['ratiorecoverybonus'] = function()
		return gameOption('Options.Ratio.Recovery.Bonus') .. '%'
	end,
	['redlife'] = function()
		return options.f_boolDisplay(gameOption('Options.RedLife'))
	end,
	['renderer'] = function()
		return gameOption('Video.RenderMode')
	end,
	['resolution'] = function()
		return gameOption('Video.GameWidth') .. 'x' .. gameOption('Video.GameHeight')
	end,
	['roundsnumsimul'] = function()
		return main.roundsNumSimul[1]
	end,
	['roundsnumsingle'] = function()
		return main.roundsNumSingle[1]
	end,
	['roundsnumtag'] = function()
		return main.roundsNumTag[1]
	end,
	['roundtime'] = function()
		return options.f_definedDisplay(gameOption('Options.Time'), {[-1] = motif.option_info.menu_valuename_none}, gameOption('Options.Time'))
	end,
	['sfxvolume'] = function()
		return gameOption('Sound.WavVolume') .. '%'
	end,
	['shaders'] = function()
		return options.f_boolDisplay(#gameOption('Video.ExternalShaders') > 0, motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
	end,
	['singlevsteamlife'] = function()
		return gameOption('Options.Team.SingleVsTeamLife') .. '%'
	end,
	['stereoeffects'] = function()
		return options.f_boolDisplay(gameOption('Sound.StereoEffects'), motif.option_info.menu_valuename_enabled, motif.option_info.menu_valuename_disabled)
	end,
	['dizzy'] = function()
		return options.f_boolDisplay(gameOption('Options.Dizzy'))
	end,
	['teamduplicates'] = function()
		return options.f_boolDisplay(gameOption('Options.Team.Duplicates'))
	end,
	['teamlifeshare'] = function()
		return options.f_boolDisplay(gameOption('Options.Team.LifeShare'))
	end,
	['teampowershare'] = function()
		return options.f_boolDisplay(gameOption('Options.Team.PowerShare'))
	end,
	['turnsrecoverybase'] = function()
		return gameOption('Options.Turns.Recovery.Base') .. '%'
	end,
	['turnsrecoverybonus'] = function()
		return gameOption('Options.Turns.Recovery.Bonus') .. '%'
	end,
	['vsync'] = function()
		return options.f_definedDisplay(gameOption('Video.VSync'), {[1] = motif.option_info.menu_valuename_enabled}, motif.option_info.menu_valuename_disabled)
	end,
	['windowscalemode'] = function()
		return options.f_boolDisplay(gameOption('Video.WindowScaleMode'), "Bilinear", "Nearest")
	end,
}

-- Returns setting value rendered alongside menu item name (calls appropriate
-- function from t_vardisplay table)
function options.f_vardisplay(itemname)
	if options.t_vardisplay[itemname] ~= nil then
		return options.t_vardisplay[itemname]()
	end
	return ''
end

-- Dynamically generates all menus and submenus, iterating over values stored in
-- main.t_sort table (in order that they're present in system.def).
function options.f_start()
	-- default menus
	if main.t_sort.option_info == nil or main.t_sort.option_info.menu == nil or #main.t_sort.option_info.menu == 0 then
		motif.setBaseOptionInfo()
	end
	-- external shaders
	options.t_shaders = {}
	for _, v in ipairs(getDirectoryFiles('external/shaders')) do
		v:gsub('^(.-)([^\\/]+)%.([^%.\\/]-)$', function(path, filename, ext)
			path = path:gsub('\\', '/')
			ext = ext:lower()
			if ext == 'frag' then
				table.insert(options.t_shaders, {path = path, filename = filename})
			end
			if ext:match('vert') or ext:match('frag') --[[or ext:match('shader')]] then
				options.t_itemname[path .. filename] = function(t, item, cursorPosY, moveTxt)
					if main.f_input(main.t_players, {'pal', 's'}) then
						sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
						local t_externalShaders = gameOption('Video.ExternalShaders')
						for k, v in ipairs(t.items) do
							if v.itemname == path .. filename then
								v.selected = not v.selected
								if v.selected then
									table.insert(t_externalShaders, v.itemname)
								else
									for k2, v2 in ipairs(t_externalShaders) do
										if v2 == v.itemname then
											table.remove(t_externalShaders, k2)
											v.vardisplay = options.f_boolDisplay(v.selected, tostring(k2), '')
											break
										end
									end
								end
							end
						end
						-- Need to correct ALL indices
						for k, v in ipairs(t.items) do
							for k2, v2 in ipairs(t_externalShaders) do
								if v2 == v.itemname then
									v.vardisplay = options.f_boolDisplay(v.selected, tostring(k2), '')
								end
							end
						end
						modifyGameOption('Video.ExternalShaders', t_externalShaders)
						return true
					end
					return true
				end
			end
		end)
	end
	for _, v in ipairs(main.f_tableExists(main.t_sort.option_info).menu) do
		-- resolution
		if v:match('_[0-9]+x[0-9]+$') then
			local width, height = v:match('_([0-9]+)x([0-9]+)$')
			options.t_itemname[width .. 'x' .. height] = function(t, item, cursorPosY, moveTxt)
				if main.f_input(main.t_players, {'pal', 's'}) then
					sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
					modifyGameOption('Video.GameWidth', tonumber(width))
					modifyGameOption('Video.GameHeight', tonumber(height))
					options.modified = true
					options.needReload = true
					return false
				end
				return true
			end
		-- ratio
		elseif v:match('_ratio[1-4]+[al].-$') then
			local ratioLevel, tmp1, tmp2 = v:match('_ratio([1-4])([al])(.-)$')
			options.t_itemname['ratio' .. ratioLevel .. tmp1 .. tmp2] = function(t, item, cursorPosY, moveTxt)
				local ratioKey = 'Options.Ratio.Level' .. tonumber(ratioLevel) .. '.' .. tmp1:upper() .. tmp2
				if main.f_input(main.t_players, {'$F'}) then
					sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
					modifyGameOption(ratioKey, gameOption(ratioKey) + 0.01)
					t.items[item].vardisplay = options.f_displayRatio(gameOption(ratioKey))
					options.modified = true
				elseif main.f_input(main.t_players, {'$B'}) and gameOption(ratioKey) > 0.01 then
					sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
					modifyGameOption(ratioKey, gameOption(ratioKey) - 0.01)
					t.items[item].vardisplay = options.f_displayRatio(gameOption(ratioKey))
					options.modified = true
				end
				return true
			end
		end
	end
	if main.debugLog then main.f_printTable(options.t_itemname, 'debug/t_optionsItemname.txt') end
	-- create menu
	options.menu = {title = main.f_itemnameUpper(motif.option_info.title_text, motif.option_info.menu_title_uppercase == 1), submenu = {}, items = {}}
	options.menu.loop = options.f_createMenu(options.menu, true)
	local t_menuWindow = main.f_menuWindow(motif.option_info)
	local t_pos = {} --for storing current options.menu table position
	local lastNum = 0
	for i, suffix in ipairs(main.f_tableExists(main.t_sort.option_info).menu) do
		for j, c in ipairs(main.f_strsplit('_', suffix)) do --split using "_" delimiter
			--populate shaders submenu
			if suffix:match('_shaders_back$') and c == 'back' then
				for k = #options.t_shaders, 1, -1 do
					local itemname = options.t_shaders[k].path .. options.t_shaders[k].filename
					local idx = 0
					-- Has the shader been enabled?
					local isSelected = false
					for i, v in ipairs(gameOption('Video.ExternalShaders')) do
						if itemname == v then
							isSelected = true
							idx = i
							break
						end
					end
					table.insert(t_pos.items, 1, {
						data = text:create({window = t_menuWindow}),
						itemname = itemname,
						displayname = options.t_shaders[k].filename,
						paramname = 'menu_itemname_' .. suffix:gsub('back$', itemname),
						vardata = text:create({window = t_menuWindow}),
						vardisplay = options.f_boolDisplay(idx > 0, tostring(idx), ''),
						selected = isSelected,
					})
					table.insert(options.t_vardisplayPointers, t_pos.items[#t_pos.items])
					--creating anim data out of appended menu items
					motif.f_loadSprData(motif.option_info, {s = 'menu_bg_' .. suffix:gsub('back$', itemname) .. '_', x = motif.option_info.menu_pos[1], y = motif.option_info.menu_pos[2]})
					motif.f_loadSprData(motif.option_info, {s = 'menu_bg_active_' .. suffix:gsub('back$', itemname) .. '_', x = motif.option_info.menu_pos[1], y = motif.option_info.menu_pos[2]})
				end
			end
			--appending the menu table
			if j == 1 then --first string after menu.itemname (either reserved one or custom submenu assignment)
				if options.menu.submenu[c] == nil or c == 'empty' then
					options.menu.submenu[c] = {title = main.f_itemnameUpper(motif.option_info['menu_itemname_' .. suffix], motif.option_info.menu_title_uppercase == 1), submenu = {}, items = {}}
					options.menu.submenu[c].loop = options.f_createMenu(options.menu.submenu[c], false)
					if not suffix:match(c .. '_') then
						table.insert(options.menu.items, {
							data = text:create({window = t_menuWindow}),
							itemname = c,
							displayname = motif.option_info['menu_itemname_' .. suffix],
							paramname = 'menu_itemname_' .. suffix,
							vardata = text:create({window = t_menuWindow}),
							vardisplay = options.f_vardisplay(c),
							selected = false,
						})
						table.insert(options.t_vardisplayPointers, options.menu.items[#options.menu.items])
					end
				end
				t_pos = options.menu.submenu[c]
				t_pos.name = c
			else --following strings
				if t_pos.submenu[c] == nil or c == 'empty' then
					t_pos.submenu[c] = {title = main.f_itemnameUpper(motif.option_info['menu_itemname_' .. suffix], motif.option_info.menu_title_uppercase == 1), submenu = {}, items = {}}
					t_pos.submenu[c].loop = options.f_createMenu(t_pos.submenu[c], false)
					table.insert(t_pos.items, {
						data = text:create({window = t_menuWindow}),
						itemname = c,
						displayname = motif.option_info['menu_itemname_' .. suffix],
						paramname = 'menu_itemname_' .. suffix,
						vardata = text:create({window = t_menuWindow}),
						vardisplay = options.f_vardisplay(c),
						selected = false,
					})
					table.insert(options.t_vardisplayPointers, t_pos.items[#t_pos.items])
				end
				if j > lastNum then
					t_pos = t_pos.submenu[c]
					t_pos.name = c
				end
			end
			lastNum = j
		end
	end
	motif.f_loadSprData(motif.option_info, {s = 'menu_item_bg_', x = 0, y = 0})
	motif.f_loadSprData(motif.option_info, {s = 'menu_item_active_bg_', x = 0, y = 0})
	animSetWindow(motif.option_info.menu_item_bg_data, t_menuWindow[1], t_menuWindow[2], t_menuWindow[3] - t_menuWindow[1], t_menuWindow[4] - t_menuWindow[2])
	animSetWindow(motif.option_info.menu_item_active_bg_data, t_menuWindow[1], t_menuWindow[2], t_menuWindow[3] - t_menuWindow[1], t_menuWindow[4] - t_menuWindow[2])
	-- log
	if main.debugLog then main.f_printTable(options.menu, 'debug/t_optionsMenu.txt') end
end

--;===========================================================
--; KEY SETTINGS
--;===========================================================
local function f_keyCfgText()
	return {text:create({}), text:create({})}
end
local t_keyCfg = {
	{data = f_keyCfgText(), itemname = 'empty', displayname = ''},
	{data = f_keyCfgText(), itemname = 'configall', displayname = motif.option_info.keymenu_itemname_configall, infodata = f_keyCfgText(), infodisplay = ''},
	{data = f_keyCfgText(), itemname = 'Up', displayname = motif.option_info.keymenu_itemname_up, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'Down', displayname = motif.option_info.keymenu_itemname_down, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'Left', displayname = motif.option_info.keymenu_itemname_left, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'Right', displayname = motif.option_info.keymenu_itemname_right, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'A', displayname = motif.option_info.keymenu_itemname_a, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'B', displayname = motif.option_info.keymenu_itemname_b, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'C', displayname = motif.option_info.keymenu_itemname_c, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'X', displayname = motif.option_info.keymenu_itemname_x, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'Y', displayname = motif.option_info.keymenu_itemname_y, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'Z', displayname = motif.option_info.keymenu_itemname_z, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'Start', displayname = motif.option_info.keymenu_itemname_start, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'D', displayname = motif.option_info.keymenu_itemname_d, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'W', displayname = motif.option_info.keymenu_itemname_w, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'Menu', displayname = motif.option_info.keymenu_itemname_menu, vardata = f_keyCfgText()},
	{data = f_keyCfgText(), itemname = 'page', displayname = '', infodata = f_keyCfgText(), infodisplay = ''},
}
t_keyCfg = main.f_tableClean(t_keyCfg, main.f_tableExists(main.t_sort.option_info).keymenu)

local rect_boxbg = rect:create({
	r =     motif.option_info.menu_boxbg_col[1],
	g =     motif.option_info.menu_boxbg_col[2],
	b =     motif.option_info.menu_boxbg_col[3],
	src =   motif.option_info.menu_boxbg_alpha[1],
	dst =   motif.option_info.menu_boxbg_alpha[2],
	defsc = motif.defaultOptions,
})
local rect_boxcursor = rect:create({
	r =     motif.option_info.menu_boxcursor_col[1],
	g =     motif.option_info.menu_boxcursor_col[2],
	b =     motif.option_info.menu_boxcursor_col[3],
	defsc = motif.defaultOptions,
})

local txt_keyController = f_keyCfgText()
local cursorPosY = 2
local item = 2
local item_start = 2
local t_pos = {}
local configall = false
local key = ''
local t_keyList = {}
local t_conflict = {}
local t_savedConfig = {}
local btnReleased = false
local player = 1
local side = 1
local btn = ''
local joyNum = 0
local t_btnEnabled = {Up = false, Down = false, Left = false, Right = false, A = false, B = false, C = false, X = false, Y = false, Z = false, Start = false, D = false, W = false, Menu = false}
for k, v in ipairs(t_keyCfg) do
	if t_btnEnabled[v.itemname] ~= nil then
		t_btnEnabled[v.itemname] = true
	end
end

function options.f_keyDefault()
	for i = 1, gameOption('Config.Players') do
		local defaultKeys = main.t_defaultKeysMapping
		if i == 1 then
			defaultKeys = {
				Up = 'UP',
				Down = 'DOWN',
				Left = 'LEFT',
				Right = 'RIGHT',
				A = 'z',
				B = 'x',
				C = 'c',
				X = 'a',
				Y = 's',
				Z = 'd',
				Start = 'RETURN',
				D = 'q',
				W = 'w',
				Menu = 'Not used',
			}
		elseif i == 2 then
			defaultKeys = {
				Up = 'i',
				Down = 'k',
				Left = 'j',
				Right = 'l',
				A = 'f',
				B = 'g',
				C = 'h',
				X = 'r',
				Y = 't',
				Z = 'y',
				Start = 'RSHIFT',
				D = 'LEFTBRACKET',
				W = 'RIGHTBRACKET',
				Menu = 'Not used',
			}
		end
		for action, button in pairs(defaultKeys) do
			if not t_btnEnabled[action] then
				modifyGameOption('Keys_P' .. i .. '.' .. action, tostring(motif.option_info.menu_valuename_nokey))
			else
				modifyGameOption('Keys_P' .. i .. '.' .. action, button)
			end
		end
		for action, button in pairs(main.t_defaultJoystickMapping) do
			if not t_btnEnabled[action] then
				modifyGameOption('Joystick_P' .. i .. '.' .. action, tostring(motif.option_info.menu_valuename_nokey))
			else
				modifyGameOption('Joystick_P' .. i .. '.' .. action, button)
			end
		end
	end
	resetRemapInput()
end

if gameOption('Config.FirstRun') then
	modifyGameOption('Config.Language', motif.languages.languages[1] or "en")
	options.f_keyDefault()
end

function options.f_keyCfgReset(cfgType)
	t_keyList = {}
	for i = 1, gameOption('Config.Players') do
		local c = gameOption(cfgType .. '_P' .. i)
		if t_keyList[c.Joystick] == nil then
			t_keyList[c.Joystick] = {} --creates subtable for each controller (1 for all keyboard configs, new one for each gamepad)
			t_conflict[c.Joystick] = false --set default conflict flag for each controller
		end
		for k, v in ipairs(t_keyCfg) do
			if t_btnEnabled[v.itemname] ~= nil then
				local btn = c[v.itemname]
				t_keyCfg[k]['vardisplay' .. i] = btn
				if btn ~= tostring(motif.option_info.menu_valuename_nokey) then --if button is not disabled
					t_keyList[c.Joystick][btn] = (t_keyList[c.Joystick][btn] or 0) + 1
				end
			end
		end
	end
end

function options.f_setKeyConfig(cfgType)
	for i = 1, gameOption('Config.Players') do
		local c = gameOption(cfgType .. '_P' .. i)
		setKeyConfig(i, c.Joystick, {c.Up, c.Down, c.Left, c.Right, c.A, c.B, c.C, c.X, c.Y, c.Z, c.Start, c.D, c.W, c.Menu})
	end
end

function options.f_keyCfgInit(cfgType, title)
	resetKey()
	main.f_cmdInput()
	cursorPosY = 2
	item = 2
	item_start = 2
	t_pos = {motif.option_info.keymenu_p1_pos, motif.option_info.keymenu_p2_pos}
	configall = false
	key = ''
	t_conflict = {}
	t_savedConfig = {}
	for i = 1, gameOption('Config.Players') do
		table.insert(t_savedConfig, gameOption(cfgType .. '_P' .. i))
	end
	btnReleased = false
	player = 1
	side = 1
	btn = ''
	options.txt_title:update({text = title})
	options.f_keyCfgReset(cfgType)
	joyNum = gameOption(cfgType .. '_P' .. player .. '.Joystick')
end

function options.f_keyCfg(cfgType, controller, bgdef, skipClear)
	local t = t_keyCfg
	--Config all
	if configall then
		--esc (reset mapping)
		if esc() --[[or main.f_input(main.t_players, {'m'})]] then
			sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
			esc(false)
			for i = 1, gameOption('Config.Players') do
				if i == player then
					modifyGameOption(cfgType .. '_P' .. i .. '.Joystick', t_savedConfig[i].Joystick)
					modifyGameOption(cfgType .. '_P' .. i .. '.Up', t_savedConfig[i].Up)
					modifyGameOption(cfgType .. '_P' .. i .. '.Down', t_savedConfig[i].Down)
					modifyGameOption(cfgType .. '_P' .. i .. '.Left', t_savedConfig[i].Left)
					modifyGameOption(cfgType .. '_P' .. i .. '.Right', t_savedConfig[i].Right)
					modifyGameOption(cfgType .. '_P' .. i .. '.A', t_savedConfig[i].A)
					modifyGameOption(cfgType .. '_P' .. i .. '.B', t_savedConfig[i].B)
					modifyGameOption(cfgType .. '_P' .. i .. '.C', t_savedConfig[i].C)
					modifyGameOption(cfgType .. '_P' .. i .. '.X', t_savedConfig[i].X)
					modifyGameOption(cfgType .. '_P' .. i .. '.Y', t_savedConfig[i].Y)
					modifyGameOption(cfgType .. '_P' .. i .. '.Z', t_savedConfig[i].Z)
					modifyGameOption(cfgType .. '_P' .. i .. '.Start', t_savedConfig[i].Start)
					modifyGameOption(cfgType .. '_P' .. i .. '.D', t_savedConfig[i].D)
					modifyGameOption(cfgType .. '_P' .. i .. '.W', t_savedConfig[i].W)
					modifyGameOption(cfgType .. '_P' .. i .. '.Menu', t_savedConfig[i].Menu)
					modifyGameOption(cfgType .. '_P' .. i .. '.GUID', t_savedConfig[i].GUID)
				end
				options.f_setKeyConfig(cfgType)
			end
			options.f_keyCfgReset(cfgType)
			item = item_start
			cursorPosY = item_start
			configall = false
			main.f_cmdBufReset()
		--spacebar (disable key)
		elseif getKey('SPACE') then
			key = 'SPACE'
		--keyboard key detection
		elseif cfgType == 'Keys' then
			key = getKey()
		--gamepad key detection
		else
			local tmp = getJoystickKey(joyNum)
			local guid = getJoystickGUID(joyNum)

			-- Fix the GUID so that configs are preserved between boots for macOS
			if gameOption(cfgType .. '_P' .. player .. '.GUID') ~= guid and guid ~= '' then
				modifyGameOption(cfgType .. '_P' .. player .. '.GUID', guid)
				options.modified = true
			end

			if tonumber(tmp) == nil then
				btnReleased = true
			elseif btnReleased then
				key = tmp
				btnReleased = false
			end
			key = tostring(key)
		end
		--other keyboard or gamepad key
		if key ~= '' then
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
				modifyGameOption(cfgType .. '_P' .. player .. '.' .. t[item].itemname, tostring(motif.option_info.menu_valuename_nokey))
				options.modified = true
			elseif cfgType == 'Keys' or (cfgType == 'Joystick' and tonumber(key) ~= nil) then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
				--decrease old button count
				if t_keyList[joyNum][btn] ~= nil and t_keyList[joyNum][btn] > 1 then
					t_keyList[joyNum][btn] = t_keyList[joyNum][btn] - 1
				else
					t_keyList[joyNum][btn] = nil
				end
				--remove previous button assignment if already set
				for k, v in ipairs(t) do
					if v['vardisplay' .. player] == key then
						v['vardisplay' .. player] = tostring(motif.option_info.menu_valuename_nokey)
						modifyGameOption(cfgType .. '_P' .. player .. '.' .. v.itemname, tostring(motif.option_info.menu_valuename_nokey))
						if t_keyList[joyNum][key] ~= nil and t_keyList[joyNum][key] > 1 then
							t_keyList[joyNum][key] = t_keyList[joyNum][key] - 1
						else
							t_keyList[joyNum][key] = nil
						end
					end
				end
				--increase new button count
				if t_keyList[joyNum][key] == nil then
					t_keyList[joyNum][key] = 1
				else
					t_keyList[joyNum][key] = t_keyList[joyNum][key] + 1
				end
				--update vardisplay / config data
				t[item]['vardisplay' .. player] = key
				modifyGameOption(cfgType .. '_P' .. player .. '.' .. t[item].itemname, key)
				options.modified = true
			end
			--move to the next position
			item = item + 1
			cursorPosY = cursorPosY + 1
			if item > #t or t[item].itemname == 'page' then
				item = item_start
				cursorPosY = item_start
				configall = false
				options.f_setKeyConfig(cfgType)
				main.f_cmdBufReset()
			end
			key = ''
		end
		if t_btnEnabled[t[item].itemname] ~= nil then
			btn = gameOption(cfgType .. '_P' .. player .. '.' .. t[item].itemname)
		end
		resetKey()
	else
		key = getKey()
		--back
		if esc() or main.f_input(main.t_players, {'m'}) or (t[item].itemname == 'page' and (side == 1 or gameOption('Config.Players') <= 2) and main.f_input(main.t_players, {'pal', 's'})) then
			if t_conflict[joyNum] then
				if not main.f_warning(main.f_extractText(motif.warning_info.text_keys_text), motif.optionbgdef) then
					options.txt_title:update({text = motif.option_info.title_input_text})
					for i = 1, gameOption('Config.Players') do
						modifyGameOption(cfgType .. '_P' .. i .. '.Joystick', t_savedConfig[i].Joystick)
						modifyGameOption(cfgType .. '_P' .. i .. '.Up', t_savedConfig[i].Up)
						modifyGameOption(cfgType .. '_P' .. i .. '.Down', t_savedConfig[i].Down)
						modifyGameOption(cfgType .. '_P' .. i .. '.Left', t_savedConfig[i].Left)
						modifyGameOption(cfgType .. '_P' .. i .. '.Right', t_savedConfig[i].Right)
						modifyGameOption(cfgType .. '_P' .. i .. '.A', t_savedConfig[i].A)
						modifyGameOption(cfgType .. '_P' .. i .. '.B', t_savedConfig[i].B)
						modifyGameOption(cfgType .. '_P' .. i .. '.C', t_savedConfig[i].C)
						modifyGameOption(cfgType .. '_P' .. i .. '.X', t_savedConfig[i].X)
						modifyGameOption(cfgType .. '_P' .. i .. '.Y', t_savedConfig[i].Y)
						modifyGameOption(cfgType .. '_P' .. i .. '.Z', t_savedConfig[i].Z)
						modifyGameOption(cfgType .. '_P' .. i .. '.Start', t_savedConfig[i].Start)
						modifyGameOption(cfgType .. '_P' .. i .. '.D', t_savedConfig[i].D)
						modifyGameOption(cfgType .. '_P' .. i .. '.W', t_savedConfig[i].W)
						modifyGameOption(cfgType .. '_P' .. i .. '.Menu', t_savedConfig[i].Menu)
						modifyGameOption(cfgType .. '_P' .. i .. '.GUID', t_savedConfig[i].GUID)
					end
					options.f_setKeyConfig(cfgType)
					menu.itemname = ''
					return false
				end
			else
				sndPlay(motif.files.snd_data, motif.option_info.cancel_snd[1], motif.option_info.cancel_snd[2])
				options.txt_title:update({text = motif.option_info.title_input_text})
				options.f_setKeyConfig(cfgType)
				menu.itemname = ''
				return false
			end
		--switch page
		elseif gameOption('Config.Players') > 2 and ((t[item].itemname == 'page' and side == 2 and main.f_input(main.t_players, {'pal', 's'})) or key == 'TAB') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			player = player + 2
			if player > gameOption('Config.Players') then
				player = side
			else
				side = main.f_playerSide(player)
			end
			joyNum = gameOption(cfgType .. '_P' .. player .. '.Joystick')
		--move right
		elseif main.f_input(main.t_players, {'$F'}) and player + 1 <= gameOption('Config.Players') then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			player = player + 1
			side = main.f_playerSide(player)
			joyNum = gameOption(cfgType .. '_P' .. player .. '.Joystick')
		--move left
		elseif main.f_input(main.t_players, {'$B'}) and player - 1 >= 1 then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			player = player - 1
			side = main.f_playerSide(player)
			joyNum = gameOption(cfgType .. '_P' .. player .. '.Joystick')
		--move up / down
		elseif main.f_input(main.t_players, {'$U', '$D'}) then
			sndPlay(motif.files.snd_data, motif.option_info.cursor_move_snd[1], motif.option_info.cursor_move_snd[2])
			if cursorPosY == item_start then
				cursorPosY = #t
				item = #t
			else
				cursorPosY = item_start
				item = item_start
			end
		--Config all
		elseif t[item].itemname == 'configall' or key:match('^F[0-9]+$') then
			local pn = key:match('^F([0-9]+)$')
			if pn ~= nil then
				pn = tonumber(pn)
				key = ''
			end
			if main.f_input(main.t_players, {'pal', 's'}) or (pn ~= nil and pn >= 1 and pn <= gameOption('Config.Players')) then
				sndPlay(motif.files.snd_data, motif.option_info.cursor_done_snd[1], motif.option_info.cursor_done_snd[2])
				if pn ~= nil then
					player = pn
					side = main.f_playerSide(player)
					joyNum = gameOption(cfgType .. '_P' .. player .. '.Joystick')
				end
				if cfgType == 'Joystick' and getJoystickPresent(joyNum) == false then
					main.f_warning(main.f_extractText(motif.warning_info.text_pad_text), motif.optionbgdef)
					item = item_start
					cursorPosY = item_start
				else
					item = item_start + 1
					cursorPosY = item_start + 1
					btnReleased = false
					configall = true
				end
			end
		end
		resetKey()
	end
	t_conflict[joyNum] = false
	--draw clearcolor
	if not skipClear then
		clearColor(motif[bgdef].bgclearcolor[1], motif[bgdef].bgclearcolor[2], motif[bgdef].bgclearcolor[3])
	end
	--draw layerno = 0 backgrounds
	bgDraw(motif[bgdef].bg, 0)
	--draw menu box
	if motif.option_info.menu_boxbg_visible == 1 then
		for i = 1, 2 do
			rect_boxbg:update({
				x1 = t_pos[i][1] + motif.option_info.keymenu_boxcursor_coords[1],
				y1 = t_pos[i][2] + motif.option_info.keymenu_boxcursor_coords[2],
				x2 = motif.option_info.keymenu_boxcursor_coords[3] - motif.option_info.keymenu_boxcursor_coords[1] + 1,
				y2 = motif.option_info.keymenu_boxcursor_coords[4] - motif.option_info.keymenu_boxcursor_coords[2] + 1 + (#t - 1) * motif.option_info.keymenu_item_spacing[2],
			})
			rect_boxbg:draw()
		end
	end
	--draw title
	options.txt_title:draw()
	--draw player num
	for i = 1, 2 do
		txt_keyController[i]:update({
			font =   motif.option_info['keymenu_item_p' .. i .. '_font'][1],
			bank =   motif.option_info['keymenu_item_p' .. i .. '_font'][2],
			align =  motif.option_info['keymenu_item_p' .. i .. '_font'][3],
			text =   motif.option_info.keymenu_itemname_playerno:gsub('%%i', tostring(i + player - side)),
			x =      motif.option_info['keymenu_p' .. i .. '_pos'][1] + motif.option_info['keymenu_item_p' .. i .. '_offset'][1],
			y =      motif.option_info['keymenu_p' .. i .. '_pos'][2] + motif.option_info['keymenu_item_p' .. i .. '_offset'][2],
			scaleX = motif.option_info['keymenu_item_p' .. i .. '_scale'][1],
			scaleY = motif.option_info['keymenu_item_p' .. i .. '_scale'][2],
			r =      motif.option_info['keymenu_item_p' .. i .. '_font'][4],
			g =      motif.option_info['keymenu_item_p' .. i .. '_font'][5],
			b =      motif.option_info['keymenu_item_p' .. i .. '_font'][6],
			height = motif.option_info['keymenu_item_p' .. i .. '_font'][7],
			xshear = motif.option_info['keymenu_item_p' .. i .. '_xshear'],
			angle  = motif.option_info['keymenu_item_p' .. i .. '_angle'],
			defsc =  motif.defaultOptions,
		})
		txt_keyController[i]:draw()
	end
	--draw menu items
	for i = 1, #t do
		for j = 1, 2 do
			if i > item - cursorPosY then
				if j == 1 then --left side
					if t[i].itemname == 'configall' then
						t[i].infodisplay = motif.option_info.menu_valuename_f:gsub('%%i', tostring(j + player - side))
					elseif t[i].itemname == 'page' then
						t[i].displayname = motif.option_info.keymenu_itemname_back
						t[i].infodisplay = motif.option_info.menu_valuename_esc
					end
				else --right side
					if t[i].itemname == 'configall' then
						t[i].infodisplay = motif.option_info.menu_valuename_f:gsub('%%i', tostring(j + player - side))
					elseif t[i].itemname == 'page' then
						if gameOption('Config.Players') > 2 then
							t[i].displayname = motif.option_info.keymenu_itemname_page
							t[i].infodisplay = motif.option_info.menu_valuename_page
						else
							t[i].displayname = motif.option_info.keymenu_itemname_back
							t[i].infodisplay = motif.option_info.menu_valuename_esc
						end
					end
				end
				if i == item and j == side then --active item
					--draw active item background
					if t[i].paramname ~= nil then
						animDraw(motif.option_info['keymenu_bg_active_' .. t[i].itemname .. '_data'])
						animUpdate(motif.option_info['keymenu_bg_active_' .. t[i].itemname .. '_data'])
					end
					--draw displayname
					t[i].data[j]:update({
						font =   motif.option_info.keymenu_item_active_font[1],
						bank =   motif.option_info.keymenu_item_active_font[2],
						align =  motif.option_info.keymenu_item_active_font[3],
						text =   t[i].displayname,
						x =      t_pos[j][1] + motif.option_info.keymenu_item_active_offset[1] + (i - 1) * motif.option_info.keymenu_item_spacing[1],
						y =      t_pos[j][2] + motif.option_info.keymenu_item_active_offset[2] + (i - 1) * motif.option_info.keymenu_item_spacing[2],
						scaleX = motif.option_info.keymenu_item_active_scale[1],
						scaleY = motif.option_info.keymenu_item_active_scale[2],
						r =      motif.option_info.keymenu_item_active_font[4],
						g =      motif.option_info.keymenu_item_active_font[5],
						b =      motif.option_info.keymenu_item_active_font[6],
						height = motif.option_info.keymenu_item_active_font[7],
						xshear = motif.option_info.keymenu_item_active_xshear,
						angle  = motif.option_info.keymenu_item_active_angle,
						defsc =  motif.defaultOptions,
					})
					t[i].data[j]:draw()
					--draw vardata
					if t[i].vardata ~= nil then
						if t_keyList[joyNum][tostring(t[i]['vardisplay' .. j + player - side])] ~= nil and t_keyList[joyNum][tostring(t[i]['vardisplay' .. j + player - side])] > 1 then
							t[i].vardata[j]:update({
								font =   motif.option_info.keymenu_item_value_conflict_font[1],
								bank =   motif.option_info.keymenu_item_value_conflict_font[2],
								align =  motif.option_info.keymenu_item_value_conflict_font[3],
								text =   t[i]['vardisplay' .. j + player - side],
								x =      t_pos[j][1] + motif.option_info.keymenu_item_value_conflict_offset[1] + (i - 1) * motif.option_info.keymenu_item_spacing[1],
								y =      t_pos[j][2] + motif.option_info.keymenu_item_value_conflict_offset[2] + (i - 1) * motif.option_info.keymenu_item_spacing[2],
								scaleX = motif.option_info.keymenu_item_value_conflict_scale[1],
								scaleY = motif.option_info.keymenu_item_value_conflict_scale[2],
								r =      motif.option_info.keymenu_item_value_conflict_font[4],
								g =      motif.option_info.keymenu_item_value_conflict_font[5],
								b =      motif.option_info.keymenu_item_value_conflict_font[6],
								height = motif.option_info.keymenu_item_value_conflict_font[7],
								xshear = motif.option_info.keymenu_item_value_conflict_xshear,
								angle  = motif.option_info.keymenu_item_value_conflict_angle,
								defsc =  motif.defaultOptions,
							})
							t[i].vardata[j]:draw()
							t_conflict[joyNum] = true
						else
							t[i].vardata[j]:update({
								font =   motif.option_info.keymenu_item_value_active_font[1],
								bank =   motif.option_info.keymenu_item_value_active_font[2],
								align =  motif.option_info.keymenu_item_value_active_font[3],
								text =   t[i]['vardisplay' .. j + player - side],
								x =      t_pos[j][1] + motif.option_info.keymenu_item_value_active_offset[1] + (i - 1) * motif.option_info.keymenu_item_spacing[1],
								y =      t_pos[j][2] + motif.option_info.keymenu_item_value_active_offset[2] + (i - 1) * motif.option_info.keymenu_item_spacing[2],
								scaleX = motif.option_info.keymenu_item_value_active_scale[1],
								scaleY = motif.option_info.keymenu_item_value_active_scale[2],
								r =      motif.option_info.keymenu_item_value_active_font[4],
								g =      motif.option_info.keymenu_item_value_active_font[5],
								b =      motif.option_info.keymenu_item_value_active_font[6],
								height = motif.option_info.keymenu_item_value_active_font[7],
								xshear = motif.option_info.keymenu_item_value_active_xshear,
								angle  = motif.option_info.keymenu_item_value_active_angle,
								defsc =  motif.defaultOptions,
							})
							t[i].vardata[j]:draw()
						end
					--draw infodata
					elseif t[i].infodata ~= nil then
						t[i].infodata[j]:update({
							font =   motif.option_info.keymenu_item_info_active_font[1],
							bank =   motif.option_info.keymenu_item_info_active_font[2],
							align =  motif.option_info.keymenu_item_info_active_font[3],
							text =   t[i].infodisplay,
							x =      t_pos[j][1] + motif.option_info.keymenu_item_info_active_offset[1] + (i - 1) * motif.option_info.keymenu_item_spacing[1],
							y =      t_pos[j][2] + motif.option_info.keymenu_item_info_active_offset[2] + (i - 1) * motif.option_info.keymenu_item_spacing[2],
							scaleX = motif.option_info.keymenu_item_value_active_scale[1],
							scaleY = motif.option_info.keymenu_item_value_active_scale[2],
							r =      motif.option_info.keymenu_item_info_active_font[4],
							g =      motif.option_info.keymenu_item_info_active_font[5],
							b =      motif.option_info.keymenu_item_info_active_font[6],
							height = motif.option_info.keymenu_item_info_active_font[7],
							xshear = motif.option_info.keymenu_item_info_active_xshear,
							angle  = motif.option_info.keymenu_item_info_active_angle,
							defsc =  motif.defaultOptions,
						})
						t[i].infodata[j]:draw()
					end
				else --inactive item
					--draw active item background
					if t[i].paramname ~= nil then
						animDraw(motif.option_info['keymenu_bg_' .. t[i].itemname .. '_data'])
						animUpdate(motif.option_info['keymenu_bg_' .. t[i].itemname .. '_data'])
					end
					--draw displayname
					t[i].data[j]:update({
						font =   motif.option_info.keymenu_item_font[1],
						bank =   motif.option_info.keymenu_item_font[2],
						align =  motif.option_info.keymenu_item_font[3],
						text =   t[i].displayname,
						x =      t_pos[j][1] + motif.option_info.keymenu_item_offset[1] + (i - 1) * motif.option_info.keymenu_item_spacing[1],
						y =      t_pos[j][2] + motif.option_info.keymenu_item_offset[2] + (i - 1) * motif.option_info.keymenu_item_spacing[2],
						scaleX = motif.option_info.keymenu_item_scale[1],
						scaleY = motif.option_info.keymenu_item_scale[2],
						r =      motif.option_info.keymenu_item_font[4],
						g =      motif.option_info.keymenu_item_font[5],
						b =      motif.option_info.keymenu_item_font[6],
						height = motif.option_info.keymenu_item_font[7],
						xshear = motif.option_info.keymenu_item_xshear,
						angle  = motif.option_info.keymenu_item_angle,
						defsc =  motif.defaultOptions,
					})
					t[i].data[j]:draw()
					--draw vardata
					if t[i].vardata ~= nil then
						if t_keyList[joyNum][tostring(t[i]['vardisplay' .. j + player - side])] ~= nil and t_keyList[joyNum][tostring(t[i]['vardisplay' .. j + player - side])] > 1 then
							t[i].vardata[j]:update({
								font =   motif.option_info.keymenu_item_value_conflict_font[1],
								bank =   motif.option_info.keymenu_item_value_conflict_font[2],
								align =  motif.option_info.keymenu_item_value_conflict_font[3],
								text =   t[i]['vardisplay' .. j + player - side],
								x =      t_pos[j][1] + motif.option_info.keymenu_item_value_conflict_offset[1] + (i - 1) * motif.option_info.keymenu_item_spacing[1],
								y =      t_pos[j][2] + motif.option_info.keymenu_item_value_conflict_offset[2] + (i - 1) * motif.option_info.keymenu_item_spacing[2],
								scaleX = motif.option_info.keymenu_item_value_conflict_scale[1],
								scaleY = motif.option_info.keymenu_item_value_conflict_scale[2],
								r =      motif.option_info.keymenu_item_value_conflict_font[4],
								g =      motif.option_info.keymenu_item_value_conflict_font[5],
								b =      motif.option_info.keymenu_item_value_conflict_font[6],
								height = motif.option_info.keymenu_item_value_conflict_font[7],
								xshear = motif.option_info.keymenu_item_value_conflict_xshear,
								angle  = motif.option_info.keymenu_item_value_conflict_angle,
								defsc =  motif.defaultOptions,
							})
							t[i].vardata[j]:draw()
							t_conflict[joyNum] = true
						else
							t[i].vardata[j]:update({
								font =   motif.option_info.keymenu_item_value_font[1],
								bank =   motif.option_info.keymenu_item_value_font[2],
								align =  motif.option_info.keymenu_item_value_font[3],
								text =   t[i]['vardisplay' .. j + player - side],
								x =      t_pos[j][1] + motif.option_info.keymenu_item_value_offset[1] + (i - 1) * motif.option_info.keymenu_item_spacing[1],
								y =      t_pos[j][2] + motif.option_info.keymenu_item_value_offset[2] + (i - 1) * motif.option_info.keymenu_item_spacing[2],
								scaleX = motif.option_info.keymenu_item_value_scale[1],
								scaleY = motif.option_info.keymenu_item_value_scale[2],
								r =      motif.option_info.keymenu_item_value_font[4],
								g =      motif.option_info.keymenu_item_value_font[5],
								b =      motif.option_info.keymenu_item_value_font[6],
								height = motif.option_info.keymenu_item_value_font[7],
								xshear = motif.option_info.keymenu_item_value_xshear,
								angle  = motif.option_info.keymenu_item_value_angle,
								defsc =  motif.defaultOptions,
							})
							t[i].vardata[j]:draw()
						end
					--draw infodata
					elseif t[i].infodata ~= nil then
						t[i].infodata[j]:update({
							font =   motif.option_info.keymenu_item_info_font[1],
							bank =   motif.option_info.keymenu_item_info_font[2],
							align =  motif.option_info.keymenu_item_info_font[3],
							text =   t[i].infodisplay,
							x =      t_pos[j][1] + motif.option_info.keymenu_item_info_offset[1] + (i - 1) * motif.option_info.keymenu_item_spacing[1],
							y =      t_pos[j][2] + motif.option_info.keymenu_item_info_offset[2] + (i - 1) * motif.option_info.keymenu_item_spacing[2],
							scaleX = motif.option_info.keymenu_item_value_active_scale[1],
							scaleY = motif.option_info.keymenu_item_value_active_scale[2],
							r =      motif.option_info.keymenu_item_info_font[4],
							g =      motif.option_info.keymenu_item_info_font[5],
							b =      motif.option_info.keymenu_item_info_font[6],
							height = motif.option_info.keymenu_item_info_font[7],
							xshear = motif.option_info.keymenu_item_info_xshear,
							angle  = motif.option_info.keymenu_item_info_angle,
							defsc =  motif.defaultOptions,
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
			if i == side then
				rect_boxcursor:update({
					x1 = t_pos[i][1] + motif.option_info.keymenu_boxcursor_coords[1] + (cursorPosY - 1) * motif.option_info.keymenu_item_spacing[1],
					y1 = t_pos[i][2] + motif.option_info.keymenu_boxcursor_coords[2] + (cursorPosY - 1) * motif.option_info.keymenu_item_spacing[2],
					x2 = motif.option_info.keymenu_boxcursor_coords[3] - motif.option_info.keymenu_boxcursor_coords[1] + 1,
					y2 = motif.option_info.keymenu_boxcursor_coords[4] - motif.option_info.keymenu_boxcursor_coords[2] + 1,
					src = src,
					dst = dst,
				})
				rect_boxcursor:draw()
			end
		end
	end
	--draw layerno = 1 backgrounds
	bgDraw(motif[bgdef].bg, 1)
	main.f_cmdInput()
	if not skipClear then
		refresh()
	end
	return true
end

return options

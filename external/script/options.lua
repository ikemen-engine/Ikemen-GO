local options = {}
--;===========================================================
--; COMMON
--;===========================================================
options.modified = false
options.needReload = false

--return string depending on bool
function options.f_boolDisplay(bool, t, f)
	if bool == true then
		return t or motif.option_info.menu.valuename.yes
	end
	return f or motif.option_info.menu.valuename.no
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
	saveGameOption(getCommandLineValue("-config"))
	hook.run("options.save", reload)
	-- Reload the game if the reload parameter is true
	if reload then
		main.f_warning(motif.warning_info.text.text.reload, motif.option_info, motif.optionbgdef)
		os.exit()
	end
	-- Reapply modified common file arrays after saving
	for _, k in ipairs({'Air', 'Cmd', 'Const', 'States', 'Fx', 'Modules', 'Lua'}) do
		modifyGameOption('Common.' .. k, t_commonFiles[k][k:lower()] or {})
	end
end

--;===========================================================
--; LOOPS
--;===========================================================
local function f_switchLanguage(dir)
	sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
	-- collect available language keys
	local codes = {}
	for code in pairs(motif.languages) do
		table.insert(codes, code)
	end
	-- current language
	local cur = gameOption('Config.Language')
	-- find current index (fallback to 1 if not found)
	local curIndex = 1
	for i, code in ipairs(codes) do
		if code == cur then
			curIndex = i
			break
		end
	end
	-- move index with wrap
	local newIndex = curIndex + dir
	if newIndex < 1 then
		newIndex = #codes
	elseif newIndex > #codes then
		newIndex = 1
	end
	-- apply
	local newCode = codes[newIndex]
	modifyGameOption('Config.Language', newCode)
	options.modified = true
	options.needReload = true
end

-- Associative elements table storing functions controlling behaviour of each
-- option screen item. Can be appended via external module.
--Render mode values as stored in the configuration.
options.t_renderModes = {
	gles32 = 'OpenGL ES 3.2',
	gl33 = 'OpenGL 3.3',
	vk13 = 'Vulkan 1.3',
}

options.t_itemname = {
	--Back
	['back'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
			return false
		end
		return true
	end,
	--Default Values
	['default'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
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
			--modifyGameOption('Options.GameSpeedStep', 5)
			modifyGameOption('Options.Match.Wins', 2)
			modifyGameOption('Options.Match.MaxDrawGames', 1)
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
			modifyGameOption('Options.Turns.Recovery.Base', 12.5)
			modifyGameOption('Options.Turns.Recovery.Bonus', 27.5)
			--modifyGameOption('Config.Motif', "data/system.def")
			modifyGameOption('Config.Players', 4)
			modifyGameOption('Config.Language', "en")
			--modifyGameOption('Config.AfterImageMax', 128)
			--modifyGameOption('Config.ExplodMax', 512)
			--modifyGameOption('Config.HelperMax', 56)
			--modifyGameOption('Config.ProjectileMax', 256)
			--modifyGameOption('Config.PaletteMax', 100)
			--modifyGameOption('Config.TextMax', 128)
			--modifyGameOption('Config.TickInterpolation', true)
			--modifyGameOption('Config.ZoomActive', true)
			--modifyGameOption('Config.EscOpensMenu', true)
			modifyGameOption('Config.VsScreenLoading', false)
			modifyGameOption('Config.TurnsLoading', false)
			--modifyGameOption('Config.BootLoadingMode', 0)
			--modifyGameOption('Config.FirstRun', false)
			--modifyGameOption('Config.WindowTitle', "Ikemen GO")
			--modifyGameOption('Config.WindowIcon', {"external/icons/IkemenCylia_256.png", "external/icons/IkemenCylia_96.png", "external/icons/IkemenCylia_48.png"})
			--modifyGameOption('Config.System', "external/script/main.lua")
			--modifyGameOption('Config.ScreenshotFolder', "")
			--modifyGameOption('Config.TrainingChar', "")
			--modifyGameOption('Config.TrainingStage', "")
			modifyGameOption('Config.GamepadMappings', "external/gamecontrollerdb.txt")
			--modifyGameOption('Config.LegacyTime', 0)
			modifyGameOption('Debug.AllowDebugMode', true)
			modifyGameOption('Debug.AllowDebugKeys', true)
			--modifyGameOption('Debug.ClipboardRows', 2)
			--modifyGameOption('Debug.ConsoleRows', 15)
			--modifyGameOption('Debug.ClsnDarken', true)
			--modifyGameOption('Debug.DumpLuaTables', false)
			--modifyGameOption('Debug.Font', "font/debug.def")
			--modifyGameOption('Debug.FontScale', 1.0)
			--modifyGameOption('Debug.StartStage', "stages/stage0-720.def")
			--modifyGameOption('Debug.ForceStageZoomout', 0)
			--modifyGameOption('Debug.ForceStageZoomin', 0)
			--modifyGameOption('Debug.KeepSpritesOnReload', 0)
			--modifyGameOption('Debug.MacOSUseCommandKey', 0)
			--modifyGameOption('Debug.SpeedTest', 100)
			-- This is platform dependent now
			local runtimeOS = getRuntimeOS()
			if runtimeOS == 'android' then
				modifyGameOption('Video.RenderMode', 'OpenGL ES 3.2')
			elseif runtimeOS == 'ios' then
				modifyGameOption('Video.RenderMode', 'Vulkan 1.3')
			else
				modifyGameOption('Video.RenderMode', "OpenGL 3.3")
			end
			modifyGameOption('Video.GameWidth', 1280)
			modifyGameOption('Video.GameHeight', 720)
			--modifyGameOption('Video.WindowWidth', 0)
			--modifyGameOption('Video.WindowHeight', 0)
			modifyGameOption('Video.Fullscreen', false)
			--modifyGameOption('Video.Borderless', false)
			--modifyGameOption('Video.RGBSpriteBilinearFilter', true)
			--modifyGameOption('Video.Framerate', 60)
			modifyGameOption('Video.VSync', 1)
			modifyGameOption('Video.MSAA', 0)
			--modifyGameOption('Video.WindowCentered', true)
			modifyGameOption('Video.ExternalShaders', {})
			modifyGameOption('Video.WindowScaleMode', true)
			modifyGameOption('Video.FightAspectWidth', -1)
			modifyGameOption('Video.FightAspectHeight', -1)
			modifyGameOption('Video.KeepAspect', true)
			modifyGameOption('Video.EnableModel', true)
			modifyGameOption('Video.EnableModelShadow', true)
			--modifyGameOption('Sound.SampleRate', 44100)
			--modifyGameOption('Sound.SoundFont', "sound/soundfont.sf2")
			modifyGameOption('Sound.StereoEffects', true)
			modifyGameOption('Sound.PanningRange', 30)
			--modifyGameOption('Sound.WavChannels', 32)
			modifyGameOption('Sound.MasterVolume', 80)
			--modifyGameOption('Sound.PauseMasterVolume', 100)
			modifyGameOption('Sound.WavVolume', 80)
			modifyGameOption('Sound.BGMVolume', 80)
			--modifyGameOption('Sound.MaxBGMVolume', 100)
			modifyGameOption('Sound.AudioDucking', false)	
			--modifyGameOption('Sound.BGMRAMBuffer', true)	
			--modifyGameOption('Sound.AudioResampleQuality', 4)	
			modifyGameOption('Arcade.AI.RandomColor', false)
			modifyGameOption('Arcade.AI.SurvivalColor', true)
			modifyGameOption('Arcade.AI.Ramping', true)
			modifyGameOption('Netplay.RollbackNetcode', true)
			modifyGameOption('Netplay.ListenPort', 7500)
			modifyGameOption('Netplay.Rollback.Port', 7500)
			modifyGameOption('Netplay.Rollback.FrameDelay', 2)
			--modifyGameOption('Netplay.Rollback.DisconnectNotifyStart', 1000)
			--modifyGameOption('Netplay.Rollback.DisconnectTimeout', 3000)
			--modifyGameOption('Netplay.Rollback.StateLogsEnabled', false)
			--modifyGameOption('Netplay.Rollback.GgpoLogsEnabled', false)
			--modifyGameOption('Netplay.Rollback.SaveStageData', false)
			--modifyGameOption('Netplay.Rollback.DesyncTest', false)
			--modifyGameOption('Netplay.Rollback.DesyncTestFrames', false)
			--modifyGameOption('Netplay.Rollback.DesyncTestAI', false)
			--modifyGameOption('Netplay.IP.<>', "")
			--modifyGameOption('Input.ButtonAssist', true)
			--modifyGameOption('Input.SOCDResolution', 4)
			--modifyGameOption('Input.ControllerStickSensitivity', 0.4)
			--modifyGameOption('Input.XinputTriggerSensitivity', 0.5)
			--modifyGameOption('Input.UiRepeatDelay', 30)
			--modifyGameOption('Input.UiRepeatRate', 4)
			--modifyGameOption('Input.PauseExitDelay', 10)

			loadFightScreen(motif.files.fight)
			setPlayers()
			for _, v in ipairs(options.t_vardisplayPointers) do
				v.vardisplay = options.f_vardisplay(v.itemname)
			end
			toggleFullscreen(gameOption('Video.Fullscreen'))
			toggleVSync(gameOption('Video.VSync'))
			updateVolume()
			options.modified = true
			options.needReload = true
			hook.run("options.default")
		end
		return true
	end,
	--Difficulty Level
	['difficulty'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Difficulty') < 8 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Difficulty', gameOption('Options.Difficulty') + 1)
			t.items[item].vardisplay = options.t_vardisplay['difficulty']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Difficulty') > 1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Difficulty', gameOption('Options.Difficulty') - 1)
			t.items[item].vardisplay = options.t_vardisplay['difficulty']()
			options.modified = true
		end
		return true
	end,
	--Time Limit
	['roundtime'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Time') < 1000 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Time', gameOption('Options.Time') + 1)
			t.items[item].vardisplay = options.t_vardisplay['roundtime']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Time') > -1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Time', gameOption('Options.Time') - 1)
			t.items[item].vardisplay = options.t_vardisplay['roundtime']()
			options.modified = true
		end
		return true
	end,
	--Language Setting
	['language'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and (main.f_tableLength(motif.languages) > 1 or motif.languages[gameOption('Config.Language')] == nil) then
			f_switchLanguage(1)
			t.items[item].vardisplay = options.t_vardisplay['language']()
		elseif getInput(-1, motif.option_info.menu.subtract.key) and (main.f_tableLength(motif.languages) > 1 or motif.languages[gameOption('Config.Language')] == nil) then
			f_switchLanguage(-1)
			t.items[item].vardisplay = options.t_vardisplay['language']()
		end
		return true
	end,
	--Life
	['lifemul'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Life') < 300 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Life', gameOption('Options.Life') + 10)
			t.items[item].vardisplay = options.t_vardisplay['lifemul']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Life') > 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Life', gameOption('Options.Life') - 10)
			t.items[item].vardisplay = options.t_vardisplay['lifemul']()
			options.modified = true
		end
		return true
	end,
	--Single VS Team Life
	['singlevsteamlife'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Team.SingleVsTeamLife') < 300 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Team.SingleVsTeamLife', gameOption('Options.Team.SingleVsTeamLife') + 10)
			t.items[item].vardisplay = options.t_vardisplay['singlevsteamlife']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Team.SingleVsTeamLife') > 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Team.SingleVsTeamLife', gameOption('Options.Team.SingleVsTeamLife') - 10)
			t.items[item].vardisplay = options.t_vardisplay['singlevsteamlife']()
			options.modified = true
		end
		return true
	end,
	-- Game Speed
	['gamespeed'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.GameSpeed') < 9 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.GameSpeed', gameOption('Options.GameSpeed') + 1)
			t.items[item].vardisplay = options.t_vardisplay['gamespeed']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.GameSpeed') > -9 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.GameSpeed', gameOption('Options.GameSpeed') - 1)
			t.items[item].vardisplay = options.t_vardisplay['gamespeed']()
			options.modified = true
		end
		return true
	end,
	--Rounds to Win (Single)
	['roundsnumsingle'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Match.Wins') < 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Match.Wins', gameOption('Options.Match.Wins') + 1)
			t.items[item].vardisplay = options.t_vardisplay['roundsnumsingle']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Match.Wins') > 1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Match.Wins', gameOption('Options.Match.Wins') - 1)
			t.items[item].vardisplay = options.t_vardisplay['roundsnumsingle']()
			options.modified = true
		end
		return true
	end,
	--Max Draw Games
	['maxdrawgames'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Match.MaxDrawGames') < 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Match.MaxDrawGames', gameOption('Options.Match.MaxDrawGames') + 1)
			t.items[item].vardisplay = options.t_vardisplay['maxdrawgames']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Match.MaxDrawGames') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Match.MaxDrawGames', gameOption('Options.Match.MaxDrawGames') - 1)
			t.items[item].vardisplay = options.t_vardisplay['maxdrawgames']()
			options.modified = true
		end
		return true
	end,
	--Credits
	['credits'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Credits') < 99 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Credits', gameOption('Options.Credits') + 1)
			t.items[item].vardisplay = options.t_vardisplay['credits']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Credits') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Credits', gameOption('Options.Credits') - 1)
			t.items[item].vardisplay = options.t_vardisplay['credits']()
			options.modified = true
		end
		return true
	end,
	--Arcade Palette
	['aipalette'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Arcade.AI.RandomColor') then
				modifyGameOption('Arcade.AI.RandomColor', false)
			else
				modifyGameOption('Arcade.AI.RandomColor', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['aipalette']()
			options.modified = true
		end
		return true
	end,
	--Survival Palette
	['aisurvivalpalette'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Arcade.AI.SurvivalColor') then
				modifyGameOption('Arcade.AI.SurvivalColor', false)
			else
				modifyGameOption('Arcade.AI.SurvivalColor', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['aisurvivalpalette']()
			options.modified = true
		end
		return true
	end,
	--AI Ramping
	['airamping'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Arcade.AI.Ramping') then
				modifyGameOption('Arcade.AI.Ramping', false)
			else
				modifyGameOption('Arcade.AI.Ramping', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['airamping']()
			options.modified = true
		end
		return true
	end,
	--Quick Continue
	['quickcontinue'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Options.QuickContinue') then
				modifyGameOption('Options.QuickContinue', false)
			else
				modifyGameOption('Options.QuickContinue', true)
				end
			t.items[item].vardisplay = options.t_vardisplay['quickcontinue']()
			options.modified = true
		end
		return true
	end,
	--Auto-Guard
	['autoguard'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Options.AutoGuard') then
				modifyGameOption('Options.AutoGuard', false)
			else
				modifyGameOption('Options.AutoGuard', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['autoguard']()
			options.modified = true
		end
		return true
	end,
	--Dizzy
	['dizzy'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Options.Dizzy') then
				modifyGameOption('Options.Dizzy', false)
			else
				modifyGameOption('Options.Dizzy', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['dizzy']()
			options.modified = true
		end
		return true
	end,
	--Guard Break
	['guardbreak'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Options.GuardBreak') then
				modifyGameOption('Options.GuardBreak', false)
			else
				modifyGameOption('Options.GuardBreak', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['guardbreak']()
			options.modified = true
		end
		return true
	end,
	--Red Life
	['redlife'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Options.RedLife') then
				modifyGameOption('Options.RedLife', false)
			else
				modifyGameOption('Options.RedLife', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['redlife']()
			options.modified = true
		end
		return true
	end,
	--Team Duplicates
	['teamduplicates'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Options.Team.Duplicates') then
				modifyGameOption('Options.Team.Duplicates', false)
			else
				modifyGameOption('Options.Team.Duplicates', true)
				end
			t.items[item].vardisplay = options.t_vardisplay['teamduplicates']()
			options.modified = true
		end
		return true
	end,
	--Team Life Share
	['teamlifeshare'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Options.Team.LifeShare') then
				modifyGameOption('Options.Team.LifeShare', false)
			else
				modifyGameOption('Options.Team.LifeShare', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['teamlifeshare']()
			options.modified = true
		end
		return true
	end,
	--Team Power Share
	['teampowershare'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Options.Team.PowerShare') then
				modifyGameOption('Options.Team.PowerShare', false)
			else
				modifyGameOption('Options.Team.PowerShare', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['teampowershare']()
			options.modified = true
		end
		return true
	end,
	--Rounds to Win (Tag)
	['roundsnumtag'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Tag.Match.Wins') < 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Tag.Match.Wins', gameOption('Options.Tag.Match.Wins') + 1)
			t.items[item].vardisplay = options.t_vardisplay['roundsnumtag']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Tag.Match.Wins') > 1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Tag.Match.Wins', gameOption('Options.Tag.Match.Wins') - 1)
			t.items[item].vardisplay = options.t_vardisplay['roundsnumtag']()
			options.modified = true
		end
		return true
	end,
	--Partner KOed Lose
	['losekotag'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Options.Tag.LoseOnKO') then
				modifyGameOption('Options.Tag.LoseOnKO', false)
			else
				modifyGameOption('Options.Tag.LoseOnKO', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['losekotag']()
			options.modified = true
		end
		return true
	end,
	--Min Tag Chars
	['mintag'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Tag.Min') < gameOption('Options.Tag.Max') then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Tag.Min', gameOption('Options.Tag.Min') + 1)
			t.items[item].vardisplay = options.t_vardisplay['mintag']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Tag.Min') > 2 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Tag.Min', gameOption('Options.Tag.Min') - 1)
			t.items[item].vardisplay = options.t_vardisplay['mintag']()
			options.modified = true
		end
		return true
	end,
	--Max Tag Chars
	['maxtag'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Tag.Max') < 4 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Tag.Max', gameOption('Options.Tag.Max') + 1)
			t.items[item].vardisplay = options.t_vardisplay['maxtag']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Tag.Max') > gameOption('Options.Tag.Min') then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Tag.Max', gameOption('Options.Tag.Max') - 1)
			t.items[item].vardisplay = options.t_vardisplay['maxtag']()
			options.modified = true
		end
		return true
	end,
	--Rounds to Win (Simul)
	['roundsnumsimul'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Simul.Match.Wins') < 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Simul.Match.Wins', gameOption('Options.Simul.Match.Wins') + 1)
			t.items[item].vardisplay = options.t_vardisplay['roundsnumsimul']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Simul.Match.Wins') > 1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Simul.Match.Wins', gameOption('Options.Simul.Match.Wins') - 1)
			t.items[item].vardisplay = options.t_vardisplay['roundsnumsimul']()
			options.modified = true
		end
		return true
	end,
	--Simul Player KOed Lose
	['losekosimul'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Options.Simul.LoseOnKO') then
				modifyGameOption('Options.Simul.LoseOnKO', false)
			else
				modifyGameOption('Options.Simul.LoseOnKO', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['losekosimul']()
			options.modified = true
		end
		return true
	end,
	--Min Simul Chars
	['minsimul'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Simul.Min') < gameOption('Options.Simul.Max') then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Simul.Min', gameOption('Options.Simul.Min') + 1)
			t.items[item].vardisplay = options.t_vardisplay['minsimul']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Simul.Min') > 2 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Simul.Min', gameOption('Options.Simul.Min') - 1)
			t.items[item].vardisplay = options.t_vardisplay['minsimul']()
			options.modified = true
		end
		return true
	end,
	--Max Simul Chars
	['maxsimul'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Simul.Max') < 4 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Simul.Max', gameOption('Options.Simul.Max') + 1)
			t.items[item].vardisplay = options.t_vardisplay['maxsimul']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Simul.Max') > gameOption('Options.Simul.Min') then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Simul.Max', gameOption('Options.Simul.Max') - 1)
			t.items[item].vardisplay = options.t_vardisplay['maxsimul']()
			options.modified = true
		end
		return true
	end,
	--Turns Recovery Base
	['turnsrecoverybase'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Turns.Recovery.Base') < 100 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Recovery.Base', gameOption('Options.Turns.Recovery.Base') + 0.5)
			t.items[item].vardisplay = options.t_vardisplay['turnsrecoverybase']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Turns.Recovery.Base') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Recovery.Base', gameOption('Options.Turns.Recovery.Base') - 0.5)
			t.items[item].vardisplay = options.t_vardisplay['turnsrecoverybase']()
			options.modified = true
		end
		return true
	end,
	--Turns Recovery Bonus
	['turnsrecoverybonus'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Turns.Recovery.Bonus') < 100 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Recovery.Bonus', gameOption('Options.Turns.Recovery.Bonus') + 0.5)
			t.items[item].vardisplay = options.t_vardisplay['turnsrecoverybonus']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Turns.Recovery.Bonus') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Recovery.Bonus', gameOption('Options.Turns.Recovery.Bonus') - 0.5)
			t.items[item].vardisplay = options.t_vardisplay['turnsrecoverybonus']()
			options.modified = true
		end
		return true
	end,
	--Min Turns Chars
	['minturns'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Turns.Min') < gameOption('Options.Turns.Max') then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Min', gameOption('Options.Turns.Min') + 1)
			t.items[item].vardisplay = options.t_vardisplay['minturns']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Turns.Min') > 1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Min', gameOption('Options.Turns.Min') - 1)
			t.items[item].vardisplay = options.t_vardisplay['minturns']()
			options.modified = true
		end
		return true
	end,
	--Max Turns Chars
	['maxturns'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Options.Turns.Max') < 8 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Max', gameOption('Options.Turns.Max') + 1)
			t.items[item].vardisplay = options.t_vardisplay['maxturns']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Options.Turns.Max') > gameOption('Options.Turns.Min') then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Max', gameOption('Options.Turns.Max') - 1)
			t.items[item].vardisplay = options.t_vardisplay['maxturns']()
			options.modified = true
		end
		return true
	end,
	--Renderer (submenu)
	['renderer'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			for k, v in ipairs(t.submenu[t.items[item].itemname].items) do
				if gameOption('Video.RenderMode') == options.t_renderModes[v.itemname] then
					v.selected = true
				else
					v.selected = false
				end
			end
			t.submenu[t.items[item].itemname].loop()
			t.items[item].vardisplay = options.t_vardisplay['renderer']()
			options.modified = true
			options.needReload = true
		end
		return true
	end,
	--gles32
	['gles32'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Video.RenderMode', options.t_renderModes.gles32)
			options.modified = true
			options.needReload = true
			return false
		end
		return true
	end,
	--gl33
	['gl33'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Video.RenderMode', options.t_renderModes.gl33)
			options.modified = true
			options.needReload = true
			return false
		end
		return true
	end,
	--vk13
	['vk13'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Video.RenderMode', options.t_renderModes.vk13)
			options.modified = true
			options.needReload = true
			return false
		end
		return true
	end,
	--Resolution (submenu)
	['resolution'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
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
			t.items[item].vardisplay = options.t_vardisplay['resolution']()
		end
		return true
	end,
	--Custom resolution
	['customres'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			local reswidth = tonumber(main.f_drawInput(
				motif.option_info.textinput.TextSpriteData,
				motif.option_info.textinput.text.reswidth,
				motif.option_info,
				motif.optionbgdef,
				motif.option_info.textinput.overlay.RectData
			))
			if reswidth ~= nil then
				sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
				local resheight = tonumber(main.f_drawInput(
					motif.option_info.textinput.TextSpriteData,
					motif.option_info.textinput.text.resheight,
					motif.option_info,
					motif.optionbgdef,
					motif.option_info.textinput.overlay.RectData
				))
				if resheight ~= nil then
					modifyGameOption('Video.GameWidth', reswidth)
					modifyGameOption('Video.GameHeight', resheight)
					sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
					options.modified = true
					options.needReload = true
				else
					sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
				end
			else
				sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
			end
			return false
		end
		return true
	end,
	--Fullscreen
	['fullscreen'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Video.Fullscreen') then
				modifyGameOption('Video.Fullscreen', false)
			else
				modifyGameOption('Video.Fullscreen', true)
			end
			toggleFullscreen(gameOption('Video.Fullscreen'))
			t.items[item].vardisplay = options.t_vardisplay['fullscreen']()
			options.modified = true
		end
		return true
	end,
	--VSync
	['vsync'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Video.VSync') == 1 then
				modifyGameOption('Video.VSync', 0)
			else
				modifyGameOption('Video.VSync', 1)
			end
			toggleVSync(gameOption('Video.VSync'))
			t.items[item].vardisplay = options.t_vardisplay['vsync']()
			options.modified = true
		end
		return true
	end,
	--MSAA
	['msaa'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Video.MSAA') < 32 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Video.MSAA') == 0 then
				modifyGameOption('Video.MSAA', 2)
			else
				modifyGameOption('Video.MSAA', gameOption('Video.MSAA') * 2)
			end
			t.items[item].vardisplay = options.t_vardisplay['msaa']()
			options.modified = true
			options.needReload = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Video.MSAA') > 1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Video.MSAA') == 2 then
				modifyGameOption('Video.MSAA', 0)
			else
				modifyGameOption('Video.MSAA', gameOption('Video.MSAA') / 2)
			end
			t.items[item].vardisplay = options.t_vardisplay['msaa']()
			options.modified = true
			options.needReload = true
		end
		return true
	end,
	--Window scaling mode
	['windowscalemode'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
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
	-- Match Aspect Ratio (submenu)
	['aspectratio'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			local t_pos = {}
			local ok = false
			local currentWidth = gameOption('Video.FightAspectWidth')
			local currentHeight = gameOption('Video.FightAspectHeight')
			for k, v in ipairs(t.submenu[t.items[item].itemname].items) do
				if v.itemname == 'defaultaspect' and currentWidth == 0 and currentHeight == 0 then
					v.selected = true
					ok = true
				elseif v.itemname == 'stageaspect' and currentWidth == -1 and currentHeight == -1 then
					v.selected = true
					ok = true
				else
					local width, height = v.itemname:match('^([0-9]+)x([0-9]+)$')
					if width and height and tonumber(width) == currentWidth and tonumber(height) == currentHeight then
						v.selected = true
						ok = true
					else
						v.selected = false
					end
				end
				if v.itemname == 'customaspect' then
					t_pos = v
				end
			end
			if not ok and t_pos.selected ~= nil then
				t_pos.selected = true
			end
			t.submenu[t.items[item].itemname].loop()
			t.items[item].vardisplay = options.t_vardisplay['aspectratio']()
		end
		return true
	end,
	--Custom aspect ratio
	['customaspect'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			local aspectwidth = tonumber(main.f_drawInput(
				motif.option_info.textinput.TextSpriteData,
				motif.option_info.textinput.text.aspectwidth,
				motif.option_info,
				motif.optionbgdef,
				motif.option_info.textinput.overlay.RectData
			))
			if aspectwidth ~= nil and aspectwidth > 0 then
				sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
				local aspectheight = tonumber(main.f_drawInput(
					motif.option_info.textinput.TextSpriteData,
					motif.option_info.textinput.text.aspectheight,
					motif.option_info,
					motif.optionbgdef,
					motif.option_info.textinput.overlay.RectData
				))
				if aspectheight ~= nil and aspectheight > 0 then
					modifyGameOption('Video.FightAspectWidth', aspectwidth)
					modifyGameOption('Video.FightAspectHeight', aspectheight)
					sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
					options.modified = true
					options.needReload = true
				else
					sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
				end
			else
				sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
			end
			return false
		end
		return true
	end,
	--Keep Aspect Ratio
	['keepaspect'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Video.KeepAspect') then
				modifyGameOption('Video.KeepAspect', false)
			else
				modifyGameOption('Video.KeepAspect', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['keepaspect']()
			options.modified = true
		end
		return true
	end,
	--Shaders (submenu)
	['shaders'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if #options.t_shaders == 0 then
				main.f_warning(motif.warning_info.text.text.shaders, motif.option_info, motif.optionbgdef)
				return true
			end
			t.submenu[t.items[item].itemname].loop()
			t.items[item].vardisplay = options.t_vardisplay['shaders']()
			options.modified = true
			options.needReload = true
		end
		return true
	end,
	--Disable (shader)
	['noshader'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
			modifyGameOption('Video.ExternalShaders', {})
			options.modified = true
			options.needReload = true
			return false
		end
		return true
	end,
	--Enable Model
	['enablemodel'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Video.EnableModel') then
				modifyGameOption('Video.EnableModel', false)
			else
				modifyGameOption('Video.EnableModel', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['enablemodel']()
			options.modified = true
			options.needReload = true
		end
		return true
	end,
	--Enable Model Shadow
	['enablemodelshadow'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Video.EnableModelShadow') then
				modifyGameOption('Video.EnableModelShadow', false)
			else
				modifyGameOption('Video.EnableModelShadow', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['enablemodelshadow']()
			options.modified = true
			options.needReload = true
		end
		return true
	end,
	--Master Volume
	['mastervolume'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Sound.MasterVolume') < 200 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.MasterVolume', gameOption('Sound.MasterVolume') + 1)
			t.items[item].vardisplay = options.t_vardisplay['mastervolume']()
			options.modified = true
			updateVolume()
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Sound.MasterVolume') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.MasterVolume', gameOption('Sound.MasterVolume') - 1)
			t.items[item].vardisplay = options.t_vardisplay['mastervolume']()
			options.modified = true
			updateVolume()
		end
		return true
	end,
	--BGM Volume
	['bgmvolume'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Sound.BGMVolume') < 100 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.BGMVolume', gameOption('Sound.BGMVolume') + 1)
			t.items[item].vardisplay = options.t_vardisplay['bgmvolume']()
			options.modified = true
			updateVolume()
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Sound.BGMVolume') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.BGMVolume', gameOption('Sound.BGMVolume') - 1)
			t.items[item].vardisplay = options.t_vardisplay['bgmvolume']()
			options.modified = true
			updateVolume()
		end
		return true
	end,
	--SFX Volume
	['sfxvolume'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Sound.WavVolume') < 100 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.WavVolume', gameOption('Sound.WavVolume') + 1)
			t.items[item].vardisplay = options.t_vardisplay['sfxvolume']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Sound.WavVolume') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.WavVolume', gameOption('Sound.WavVolume') - 1)
			t.items[item].vardisplay = options.t_vardisplay['sfxvolume']()
			options.modified = true
		end
		return true
	end,
	--Audio Ducking
	['audioducking'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Sound.AudioDucking') then
				modifyGameOption('Sound.AudioDucking', false)
			else
				modifyGameOption('Sound.AudioDucking', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['audioducking']()
			options.modified = true
		end
		return true
	end,
	--Stereo Effects
	['stereoeffects'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Sound.StereoEffects') then
				modifyGameOption('Sound.StereoEffects', false)
			else
				modifyGameOption('Sound.StereoEffects', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['stereoeffects']()
			options.modified = true
		end
		return true
	end,
	--Panning Range
	['panningrange'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Sound.PanningRange') < 100 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.PanningRange', gameOption('Sound.PanningRange') + 1)
			t.items[item].vardisplay = options.t_vardisplay['panningrange']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Sound.PanningRange') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.PanningRange', gameOption('Sound.PanningRange') - 1)
			t.items[item].vardisplay = options.t_vardisplay['panningrange']()
			options.modified = true
		end
		return true
	end,
	--Key Config
	['keyboard'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			options.f_inputCfg('keyboard', t.submenu[t.items[item].itemname])
			textImgSetText(motif.option_info.title.TextSpriteData, t.title)
		end
		return true
	end,
	--Joystick Config
	['gamepad'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			if getCommandLineValue("-nojoy") == nil then
				options.f_inputCfg('gamepad', t.submenu[t.items[item].itemname])
				textImgSetText(motif.option_info.title.TextSpriteData, t.title)
			end
		end
		return true
	end,
	--Default
	['inputdefault'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			inputDefaults()
			options.modified = true
		end
		return true
	end,
	--Players
	['players'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Config.Players') < 8 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Config.Players', math.min(8, gameOption('Config.Players') + 2))
			t.items[item].vardisplay = options.t_vardisplay['players']()
			setPlayers()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Config.Players') > 2 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Config.Players', math.max(2, gameOption('Config.Players') - 2))
			t.items[item].vardisplay = options.t_vardisplay['players']()
			setPlayers()
			options.modified = true
		end
		return true
	end,
	--Debug Keys
	['debugkeys'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Debug.AllowDebugKeys') then
				modifyGameOption('Debug.AllowDebugKeys', false)
			else
				modifyGameOption('Debug.AllowDebugKeys', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['debugkeys']()
			options.modified = true
		end
		return true
	end,
	--Debug Mode
	['debugmode'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Debug.AllowDebugMode') then
				modifyGameOption('Debug.AllowDebugMode', false)
			else
				modifyGameOption('Debug.AllowDebugMode', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['debugmode']()
			options.modified = true
		end
		return true
	end,
	--Background Loading
	['vsscreenloading'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Config.VsScreenLoading') then
				modifyGameOption('Config.VsScreenLoading', false)
			else
				modifyGameOption('Config.VsScreenLoading', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['vsscreenloading']()
			options.modified = true
		end
		return true
	end,
	--Turns Preloading
	['turnsloading'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Config.TurnsLoading') then
				modifyGameOption('Config.TurnsLoading', false)
			else
				modifyGameOption('Config.TurnsLoading', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['turnsloading']()
			options.modified = true
		end
		return true
	end,
	--Rollback Netcode
	['netcode'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Netplay.RollbackNetcode') then
				modifyGameOption('Netplay.RollbackNetcode', false)
			else
				modifyGameOption('Netplay.RollbackNetcode', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['netcode']()
			options.modified = true
		end
		return true
	end,
	--Host Port (TCP)
	['portchange'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			local port = tonumber(main.f_drawInput(
				motif.option_info.textinput.TextSpriteData,
				motif.option_info.textinput.text.port,
				motif.option_info,
				motif.optionbgdef,
				motif.option_info.textinput.overlay.RectData
			))
			if port ~= nil and port == math.floor(port) and port >= 1 and port <= 65535 then
				sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
				modifyGameOption('Netplay.ListenPort', port)
				t.items[item].vardisplay = options.t_vardisplay['portchange']()
				options.modified = true
			else
				sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
			end
		end
		return true
	end,
	--Rollback Port (UDP)
	['rollbackport'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			local port = tonumber(main.f_drawInput(
				motif.option_info.textinput.TextSpriteData,
				motif.option_info.textinput.text.port,
				motif.option_info,
				motif.optionbgdef,
				motif.option_info.textinput.overlay.RectData
			))
			if port ~= nil and port == math.floor(port) and port >= 1 and port <= 65535 then
				sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
				modifyGameOption('Netplay.Rollback.Port', port)
				t.items[item].vardisplay = options.t_vardisplay['rollbackport']()
				options.modified = true
			else
				sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
			end
		end
		return true
	end,
	--Rollback Frame Delay
	['framedelay'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key) and gameOption('Netplay.Rollback.FrameDelay') < 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Netplay.Rollback.FrameDelay', gameOption('Netplay.Rollback.FrameDelay') + 1)
			t.items[item].vardisplay = options.t_vardisplay['framedelay']()
			options.modified = true
		elseif getInput(-1, motif.option_info.menu.subtract.key) and gameOption('Netplay.Rollback.FrameDelay') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Netplay.Rollback.FrameDelay', gameOption('Netplay.Rollback.FrameDelay') - 1)
			t.items[item].vardisplay = options.t_vardisplay['framedelay']()
			options.modified = true
		end
		return true
	end,
	--Save and Return
	['savereturn'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			if options.modified then
				options.f_saveCfg(options.needReload)
			end
			fadeOutInit(motif.option_info.fadeout.FadeData)
			main.close = true
			--return false
		end
		return true
	end,
	--Return Without Saving
	['return'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
			if options.needReload then
				main.f_warning(motif.warning_info.text.text.noreload, motif.option_info, motif.optionbgdef)
			end
			fadeOutInit(motif.option_info.fadeout.FadeData)
			main.close = true
			--return false
		end
		return true
	end,
	--Save Settings
	['savesettings'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			if options.modified then
				options.f_saveCfg(options.needReload)
			end
		end
		return true
	end,
}

-- Shared menu loop logic
function options.f_createMenu(tbl, bool_main)
	local cursorPosY, moveTxt, item = 1, 0, 1
	local initialized = false
	return function(context)
		hook.run("options.menu.loop")
		if not initialized or context == nil then
			cursorPosY, moveTxt, item = 1, 0, 1
			main.f_menuSnap(motif.option_info)
			main.f_menuItemBgAnimReset(motif.option_info)
			initialized = true
		end
		local t = tbl.items
		if #t == 0 then
			return false
		end
		if bool_main then
			bgReset(motif.optionbgdef.BGDef)
			fadeInInit(motif.option_info.fadein.FadeData)
			playBgm({source = "motif.option"})
			main.close = false
		end
		while true do
			if tbl.inputMenu ~= nil and tbl.inputMenu.finished then
				break
			end
			if tbl.inputSource ~= nil and tbl.inputSource.kind == 'gamepad' and not getJoystickPresent(tbl.inputSource.joy) then
				break
			end
			textImgReset(motif.option_info.title.TextSpriteData)
			textImgSetText(motif.option_info.title.TextSpriteData, tbl.title)
			if tbl.inputFocus ~= nil then
				-- Forced movement uses the same cursor and scrolling rules as normal navigation.
				while item ~= tbl.inputFocus do
					local direction = -1
					if item < tbl.inputFocus then
						direction = 1
					end
					cursorPosY, moveTxt, item = main.f_menuCommonCalc(t, item, cursorPosY, moveTxt, motif.option_info, motif.option_info.cursor, direction)
				end
				tbl.inputFocus = nil
			end
			if tbl.reset and context == nil then
				tbl.reset = false
			else
				main.f_menuCommonDraw(t, item, cursorPosY, moveTxt, motif.option_info,
					context and context.background or motif.optionbgdef, context ~= nil)
			end
			-- While fading, ignore normal option-menu inputs, but still allow ESC / menu-cancel to skip the fade.
			if context ~= nil and (context.capture ~= nil or context.popup ~= nil or context.skipInput) then
				cursorPosY, moveTxt, item = main.f_menuCommonCalc(t, item, cursorPosY, moveTxt, motif.option_info, motif.option_info.cursor, 0)
			elseif fadeActive() then
				if esc() or getInput(-1, motif.option_info.menu.cancel.key) then
					fadeSkip()
				end
			else
				cursorPosY, moveTxt, item = main.f_menuCommonCalc(t, item, cursorPosY, moveTxt, motif.option_info, motif.option_info.cursor)
				if main.close then
					bgReset(motif[main.background].BGDef)
					fadeInInit(motif[main.group].fadein.FadeData)
					playBgm({source = "motif.title", interrupt = true})
					main.close = false
					break
				elseif esc() or getInput(-1, motif.option_info.menu.cancel.key) then
					sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
					if bool_main then
						if options.modified then
							--options.f_saveCfg(options.needReload)
						end
						if options.needReload then
							main.f_warning(motif.warning_info.text.text.noreload, motif.option_info, motif.optionbgdef)
						end
						fadeOutInit(motif.option_info.fadeout.FadeData)
						main.close = true
					else
						break
					end
				elseif tbl.inputSource ~= nil and getKey('F1') then
					options.f_inputConfigAll(tbl)
				elseif tbl == tbl.inputMenu and getInput(-1, motif.option_info.menu.save.key) then
					sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
					tbl.finished = true
					break
				elseif options.t_itemname[t[item].itemname] ~= nil then
					if not options.t_itemname[t[item].itemname](tbl, item, cursorPosY, moveTxt) then
						break
					end
				elseif getInput(-1, motif.option_info.menu.done.key) then
					local f = t[item].itemname
					if tbl.submenu[f].loop ~= nil then
						sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
						if context ~= nil then
							table.insert(context.parents, tbl)
							context.current = tbl.submenu[f]
						else
							tbl.submenu[f].loop()
						end
						main.f_menuSnap(motif.option_info)
						main.f_menuItemBgAnimReset(motif.option_info)
					elseif not options.t_itemname[f](tbl, item, cursorPosY, moveTxt) then
						break
					end
				end
			end
			if context ~= nil then
				return true
			end
		end
		initialized = false
		return false
	end
end

options.t_vardisplayPointers = {}

local function findItemnameBySuffix(sfx)
	for _, k in ipairs(motif.option_info.menu.itemname_order) do
		if k:match(sfx) then
			return motif.option_info.menu.itemname[k]
		end
	end
	return ''
end

-- Associative elements table storing functions returning current setting values
-- rendered alongside menu item name. Can be appended via external module.
options.t_vardisplay = {
	['aipalette'] = function()
		return options.f_boolDisplay(gameOption('Arcade.AI.RandomColor'), motif.option_info.menu.valuename.random, motif.option_info.menu.valuename.default)
	end,
	['aisurvivalpalette'] = function()
		return options.f_boolDisplay(gameOption('Arcade.AI.SurvivalColor'), motif.option_info.menu.valuename.random, motif.option_info.menu.valuename.default)
	end,
	['airamping'] = function()
		return options.f_boolDisplay(gameOption('Arcade.AI.Ramping'))
	end,
	['aspectratio'] = function()
		local width = gameOption('Video.FightAspectWidth')
		local height = gameOption('Video.FightAspectHeight')
		if width > 0 and height > 0 then
			return width .. ':' .. height
		elseif width < 0 and height < 0 then
			return findItemnameBySuffix('stageaspect$')
		else
			return findItemnameBySuffix('defaultaspect$')
		end
	end,
	['audioducking'] = function()
		return options.f_boolDisplay(gameOption('Sound.AudioDucking'), motif.option_info.menu.valuename.enabled, motif.option_info.menu.valuename.disabled)
	end,
	['autoguard'] = function()
		return options.f_boolDisplay(gameOption('Options.AutoGuard'))
	end,
	['vsscreenloading'] = function()
		return options.f_boolDisplay(gameOption('Config.VsScreenLoading'), motif.option_info.menu.valuename.enabled, motif.option_info.menu.valuename.disabled)
	end,
	['turnsloading'] = function()
		return options.f_boolDisplay(gameOption('Config.TurnsLoading'), motif.option_info.menu.valuename.enabled, motif.option_info.menu.valuename.disabled)
	end,
	['bgmvolume'] = function()
		return gameOption('Sound.BGMVolume') .. '%'
	end,
	['credits'] = function()
		return options.f_definedDisplay(gameOption('Options.Credits'), {[0] = motif.option_info.menu.valuename.disabled}, gameOption('Options.Credits'))
	end,
	['debugkeys'] = function()
		return options.f_boolDisplay(gameOption('Debug.AllowDebugKeys'), motif.option_info.menu.valuename.enabled, motif.option_info.menu.valuename.disabled)
	end,
	['debugmode'] = function()
		return options.f_boolDisplay(gameOption('Debug.AllowDebugMode'), motif.option_info.menu.valuename.enabled, motif.option_info.menu.valuename.disabled)
	end,
	['difficulty'] = function()
		return gameOption('Options.Difficulty')
	end,
	['enablemodel'] = function()
		return options.f_definedDisplay(gameOption('Video.EnableModel'), {[true] = motif.option_info.menu.valuename.enabled}, motif.option_info.menu.valuename.disabled)
	end,
	['enablemodelshadow'] = function()
		return options.f_definedDisplay(gameOption('Video.EnableModelShadow'), {[true] = motif.option_info.menu.valuename.enabled}, motif.option_info.menu.valuename.disabled)
	end,
	['fullscreen'] = function()
		return options.f_boolDisplay(gameOption('Video.Fullscreen'))
	end,
	['gamespeed'] = function()
		return options.f_boolDisplay(gameOption('Options.GameSpeed') == 0, motif.option_info.menu.valuename.normal, options.f_boolDisplay(gameOption('Options.GameSpeed') < 0, string.format(motif.option_info.menu.valuename.slow, 0 - gameOption('Options.GameSpeed')), string.format(motif.option_info.menu.valuename.fast, gameOption('Options.GameSpeed'))))
	end,
	['guardbreak'] = function()
		return options.f_boolDisplay(gameOption('Options.GuardBreak'))
	end,
	['keepaspect'] = function()
		return options.f_boolDisplay(gameOption('Video.KeepAspect'), motif.option_info.menu.valuename.enabled, motif.option_info.menu.valuename.disabled)
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
		return gameOption('Options.Match.MaxDrawGames')
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
		return options.f_definedDisplay(gameOption('Video.MSAA'), {[0] = motif.option_info.menu.valuename.disabled}, gameOption('Video.MSAA') .. 'x')
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
	['quickcontinue'] = function()
		return options.f_boolDisplay(gameOption('Options.QuickContinue'))
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
	['rollbackport'] = function()
		return gameOption('Netplay.Rollback.Port')
	end,
	['roundsnumsimul'] = function()
		return gameOption('Options.Simul.Match.Wins')
	end,
	['roundsnumsingle'] = function()
		return gameOption('Options.Match.Wins')
	end,
	['roundsnumtag'] = function()
		return gameOption('Options.Tag.Match.Wins')
	end,
	['roundtime'] = function()
		return options.f_definedDisplay(gameOption('Options.Time'), {[-1] = motif.option_info.menu.valuename.none}, gameOption('Options.Time'))
	end,
	['sfxvolume'] = function()
		return gameOption('Sound.WavVolume') .. '%'
	end,
	['shaders'] = function()
		return options.f_boolDisplay(#gameOption('Video.ExternalShaders') > 0, motif.option_info.menu.valuename.enabled, motif.option_info.menu.valuename.disabled)
	end,
	['singlevsteamlife'] = function()
		return gameOption('Options.Team.SingleVsTeamLife') .. '%'
	end,
	['stereoeffects'] = function()
		return options.f_boolDisplay(gameOption('Sound.StereoEffects'), motif.option_info.menu.valuename.enabled, motif.option_info.menu.valuename.disabled)
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
		return options.f_definedDisplay(gameOption('Video.VSync'), {[1] = motif.option_info.menu.valuename.enabled}, motif.option_info.menu.valuename.disabled)
	end,
	['windowscalemode'] = function()
		return options.f_boolDisplay(gameOption('Video.WindowScaleMode'), motif.option_info.menu.valuename.bilinear, motif.option_info.menu.valuename.nearest)
	end,
	['netcode'] = function()
		return options.f_boolDisplay(gameOption('Netplay.RollbackNetcode'), motif.option_info.menu.valuename.rollbacknetcode, motif.option_info.menu.valuename.delaynetcode)
	end,
	['framedelay'] = function()
		return gameOption('Netplay.Rollback.FrameDelay')
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

-- Dynamically generates all menus and submenus
function options.f_start()
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
					if getInput(-1, motif.option_info.menu.done.key) then
						sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
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
	for _, v in ipairs(motif.option_info.menu.itemname_order) do
		-- resolution
		if v:match('_resolution_[0-9]+x[0-9]+$') then
			local width, height = v:match('_resolution_([0-9]+)x([0-9]+)$')
			options.t_itemname[width .. 'x' .. height] = function(t, item, cursorPosY, moveTxt)
				if getInput(-1, motif.option_info.menu.done.key) then
					sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
					modifyGameOption('Video.GameWidth', tonumber(width))
					modifyGameOption('Video.GameHeight', tonumber(height))
					options.modified = true
					options.needReload = true
					return false
				end
				return true
			end
		-- aspect ratio
		elseif v:match('_aspectratio_') then
			-- aspect ratio default
			if v:match('_aspectratio_defaultaspect$') then
				options.t_itemname['defaultaspect'] = function(t, item, cursorPosY, moveTxt)
					if getInput(-1, motif.option_info.menu.done.key) then
						sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
						modifyGameOption('Video.FightAspectWidth', 0)
						modifyGameOption('Video.FightAspectHeight', 0)
						options.modified = true
						return false
					end
					return true
				end
			-- aspect ratio stage
			elseif v:match('_aspectratio_stageaspect$') then
				options.t_itemname['stageaspect'] = function(t, item, cursorPosY, moveTxt)
					if getInput(-1, motif.option_info.menu.done.key) then
						sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
						modifyGameOption('Video.FightAspectWidth', -1)
						modifyGameOption('Video.FightAspectHeight', -1)
						options.modified = true
						return false
					end
					return true
				end
			-- aspect ratio presets (4x3, 16x9, etc.)
			elseif v:match('_aspectratio_[0-9]+x[0-9]+$') then
				local width, height = v:match('_aspectratio_([0-9]+)x([0-9]+)$')
				options.t_itemname[width .. 'x' .. height] = function(t, item, cursorPosY, moveTxt)
					if getInput(-1, motif.option_info.menu.done.key) then
						sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
						modifyGameOption('Video.FightAspectWidth', tonumber(width))
						modifyGameOption('Video.FightAspectHeight', tonumber(height))
						options.modified = true
						return false
					end
					return true
				end
			end
		end
	end
	if gameOption('Debug.DumpLuaTables') then main.f_printTable(options.t_itemname, 'debug/t_optionsItemname.txt') end
	-- create menu
	options.menu = {title = main.f_itemnameUpper(motif.option_info.title.text, motif.option_info.menu.title.uppercase), submenu = {}, items = {}}
	options.menu.loop = options.f_createMenu(options.menu, true)
	local w = main.f_menuWindow(motif.option_info.menu)
	local t_pos = {} --for storing current options.menu table position
	local lastNum = 0
	for i, suffix in ipairs(motif.option_info.menu.itemname_order) do
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
						itemname = itemname,
						displayname = options.t_shaders[k].filename,
						paramname = suffix:gsub('back$', itemname),
						vardisplay = options.f_boolDisplay(idx > 0, tostring(idx), ''),
						selected = isSelected,
					})
					table.insert(options.t_vardisplayPointers, t_pos.items[#t_pos.items])
				end
			end
			--appending the menu table
			if j == 1 then --first string after menu.itemname (either reserved one or custom submenu assignment)
				if options.menu.submenu[c] == nil or c:match("^spacer%d*$") then
					options.menu.submenu[c] = {title = main.f_itemnameUpper(motif.option_info.menu.itemname[suffix], motif.option_info.menu.title.uppercase), submenu = {}, items = {}}
					options.menu.submenu[c].loop = options.f_createMenu(options.menu.submenu[c], false)
					if not suffix:match(c .. '_') then
						table.insert(options.menu.items, {
							itemname = c,
							displayname = motif.option_info.menu.itemname[suffix],
							paramname = suffix,
							vardisplay = options.f_vardisplay(c),
							selected = false,
						})
						table.insert(options.t_vardisplayPointers, options.menu.items[#options.menu.items])
					end
				end
				t_pos = options.menu.submenu[c]
				t_pos.name = c
			else --following strings
				if t_pos.submenu[c] == nil or c:match("^spacer%d*$") then
					t_pos.submenu[c] = {title = main.f_itemnameUpper(motif.option_info.menu.itemname[suffix], motif.option_info.menu.title.uppercase), submenu = {}, items = {}}
					t_pos.submenu[c].loop = options.f_createMenu(t_pos.submenu[c], false)
					table.insert(t_pos.items, {
						itemname = c,
						displayname = motif.option_info.menu.itemname[suffix],
						paramname = suffix,
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
	-- Runtime platform filtering
	local runtimeOS = getRuntimeOS()
	local excluded = {gles32 = true}
	if runtimeOS == 'android' then
		excluded = {keyboard = true, gl33 = true}
	elseif runtimeOS == 'ios' then
		excluded = {keyboard = true, gl33 = true, gles32 = true}
	end
	main.f_pruneMenu(options.menu, excluded)
	main.f_prunePointers(options.t_vardisplayPointers, excluded)
	-- Menu windows
	textImgSetWindow(motif.option_info.menu.item.selected.active.TextSpriteData, w[1], w[2], w[3], w[4])
	textImgSetWindow(motif.option_info.menu.item.active.TextSpriteData, w[1], w[2], w[3], w[4])
	textImgSetWindow(motif.option_info.menu.item.value.active.TextSpriteData, w[1], w[2], w[3], w[4])
	textImgSetWindow(motif.option_info.menu.item.selected.TextSpriteData, w[1], w[2], w[3], w[4])
	textImgSetWindow(motif.option_info.menu.item.TextSpriteData, w[1], w[2], w[3], w[4])
	textImgSetWindow(motif.option_info.menu.item.value.TextSpriteData, w[1], w[2], w[3], w[4])
	for _, v in pairs(motif.option_info.menu.item.bg) do
		animSetWindow(v.AnimData, w[1], w[2], w[3], w[4])
	end
	for _, v in pairs(motif.option_info.menu.item.active.bg) do
		animSetWindow(v.AnimData, w[1], w[2], w[3], w[4])
	end
	-- log
	if gameOption('Debug.DumpLuaTables') then main.f_printTable(options.menu, 'debug/t_optionsMenu.txt') end
end

--;===========================================================
--; KEY SETTINGS
--;===========================================================
local t_profileFields = {'up', 'down', 'left', 'right', 'a', 'b', 'c', 'x', 'y', 'z', 'start', 'd', 'w', 'menu'}
local t_profileButtons = {}
for _, field in ipairs(t_profileFields) do
	t_profileButtons[field] = true
end

local function findInputItem(tbl, name)
	for i, row in ipairs(tbl.items) do
		if row.itemname == name then
			return row, tbl.submenu[row.itemname], tbl, i
		end
		local sub = tbl.submenu[row.itemname]
		if sub ~= nil then
			local row, menu, parent, item = findInputItem(sub, name)
			if row ~= nil then
				return row, menu, parent, item
			end
		end
	end
end

local function inputPopup(context, text, getValue, onClose)
	local sec = motif.option_info
	context.popup = main.f_warning(text, sec, context.background, nil, sec.title.TextSpriteData, nil,
		sec.cancel.snd, sec.cursor.done.snd, getValue, true)
	context.onClose = onClose
end

-- Expand the profile placeholder in place, retaining the order the motif declares.
local function populateInputMenu(tbl, kind, selected)
	for i = #tbl.items, 1, -1 do
		local row = tbl.items[i]
		if row.itemname == 'profile' then
			table.remove(tbl.items, i)
			local entries = inputSources()
			if selected ~= nil then
				entries = inputProfiles(selected.id)
			end
			for j = #entries, 1, -1 do
				local data = entries[j]
				local source = selected or data
				if source.connected and source.kind == kind then
					local entry = main.f_tableCopy(row)
					entry.source = source
					entry.profile = nil
					if selected ~= nil then
						entry.profile = data.id
					end
					entry.displayname = string.format(row.displayname, data.name)
					entry.vardisplay = nil
					entry.selected = data.selected or false
					table.insert(tbl.items, i, entry)
				end
			end
		elseif tbl.submenu[row.itemname] ~= nil then
			if row.itemname == 'profiles' and selected ~= nil then
				row.vardisplay = inputProfile(selected.id).name
			end
			populateInputMenu(tbl.submenu[row.itemname], kind, selected)
		end
	end
	for i, row in ipairs(tbl.items) do
		if row.selected then
			tbl.inputFocus = i
			break
		end
	end
	tbl.loop = options.f_createMenu(tbl, false)
end

local function updateProfileMenu(tbl, root, source, values)
	if tbl.inputMenu == nil then
		tbl.loop = options.f_createMenu(tbl, false)
	end
	tbl.inputContext = root.inputContext
	tbl.inputSource = source
	tbl.inputMenu = root
	for _, row in ipairs(tbl.items) do
		local field = row.itemname
		row.vardisplay = values[field]
		if field == 'rumble' then
			row.vardisplay = options.f_boolDisplay(values.rumble)
		elseif field == 'configall' then
			row.infodisplay = string.format(motif.option_info.menu.valuename.f, 1)
		end
	end
	for _, sub in pairs(tbl.submenu) do
		updateProfileMenu(sub, root, source, values)
	end
end

local function applyProfile(tbl, values, copy)
	local sec = motif.option_info
	if not inputSaveProfile(tbl.inputSource.id, values.name, values, values.rumble, copy or false, sec.menu.valuename.copy) then
		inputPopup(tbl.inputContext, sec.textinput.text.profilesave)
		return false
	end
	options.modified = true
	local values = inputProfile(tbl.inputSource.id)
	local root = tbl.inputMenu
	root.title = main.f_itemnameUpper(values.name, sec.menu.title.uppercase)
	updateProfileMenu(root, root, tbl.inputSource, values)
	return true
end

local function cancelProfile(root)
	if root.previousProfile == nil then
		applyProfile(root, root.savedBindings)
		return
	end
	inputDeleteProfile(root.inputSource.id)
	if root.previousProfile ~= '' then
		inputSetProfile(root.inputSource.id, root.previousProfile)
	end
end

function options.f_inputConfigFrame(context)
	local sec = motif.option_info
	-- Tracks whether this frame rendered a menu, so the caller only flushes a
	-- frame that drew something (the closing frame may only clear the screen).
	local drew = false
	if not context.pause then
		local bg = context.background
		clearColor(bg.bgclearcolor[1], bg.bgclearcolor[2], bg.bgclearcolor[3])
	end
	if not context.closed and context.popup ~= nil then
		-- The popup overlays the options screen, so draw that background behind it
		-- instead of leaving the cleared (black) frame.
		if not context.pause and context.background ~= nil then
			bgDraw(context.background.BGDef, 0)
			bgDraw(context.background.BGDef, 1)
		end
		local finished, value = context.popup()
		if not finished then
			return true
		end
		local onClose = context.onClose
		context.popup, context.onClose = nil, nil
		context.skipInput = true
		if onClose ~= nil then
			onClose(value)
		end
	end
	local source = context.source or (context.current and context.current.inputSource)
	if source ~= nil and source.kind == 'gamepad' and not getJoystickPresent(source.joy) then
		context.closed = true
	end
	if not context.closed and context.current ~= nil then
		drew = true
		if context.capture ~= nil then
			options.f_inputCapture(context)
		end
		local tbl = context.current
		tbl.inputContext = context
		if not tbl.loop(context) then
			if tbl == tbl.inputMenu and not tbl.finished and (tbl.inputSource.kind ~= 'gamepad' or getJoystickPresent(tbl.inputSource.joy)) then
				cancelProfile(tbl)
			end
			if tbl.inputMenu ~= nil and (tbl == tbl.inputMenu or tbl.inputMenu.finished) then
				context.current, context.parents = nil, {}
				if context.openPicker ~= nil then
					context.openPicker()
				end
			else
				context.current = table.remove(context.parents)
			end
			context.closed = context.current == nil
			main.f_menuSnap(sec)
			main.f_menuItemBgAnimReset(sec)
			resetTokenGuard()
			esc(false)
		end
	end
	context.skipInput = false
	if context.closed then
		if context.previousSource ~= nil then
			inputConfigSource(context.previousSource)
		end
		main.f_menuSnap(sec)
		main.f_menuItemBgAnimReset(sec)
		resetKey()
		resetTokenGuard()
		esc(false)
		return false, drew
	end
	return true
end

local function runInputConfig(context)
	resetKey()
	resetTokenGuard()
	if context.pause then
		options.inputContext = context
	else
		local continue, drew = options.f_inputConfigFrame(context)
		while continue do
			refresh()
			continue, drew = options.f_inputConfigFrame(context)
		end
		-- Flush the final editor frame on its own only if it drew a menu, so it is
		-- not composited with the options menu, without presenting a blank frame.
		if drew then
			refresh()
		end
	end
end

function options.f_editProfile(source, template, context, newProfile)
	if template == nil then
		local row, picker = findInputItem(options.menu, source.kind)
		if picker ~= nil then
			row, template = findInputItem(picker, 'profile')
		end
	end
	if template == nil then
		return
	end
	local ownContext = context == nil
	context = context or {background = motif.optionbgdef, parents = {}}
	local tbl = main.f_tableCopy(template)
	if source.kind == 'keyboard' then
		main.f_pruneMenu(tbl, {keyboardrumble = true, gamepadrumble = true})
	end
	-- A profile made by New Profile is already a separate copy.
	if newProfile then
		main.f_pruneMenu(tbl, {savecopy = true})
	end
	local values = inputProfile(source.id)
	tbl.title = main.f_itemnameUpper(values.name, motif.option_info.menu.title.uppercase)
	tbl.savedBindings = main.f_tableCopy(values)
	tbl.inputContext = context
	updateProfileMenu(tbl, tbl, source, values)
	context.current, context.parents = tbl, {}
	-- Capture targets the edited source; menu navigation keeps the source that opened it.
	local menuSource = source.id
	if not ownContext then
		menuSource = inputMenuSource()
		if context.source ~= nil then
			menuSource = context.source.id
		end
	end
	local previousSource = inputConfigSource(menuSource)
	if context.previousSource == nil then
		context.previousSource = previousSource
	end
	resetKey()
	resetTokenGuard()
	if ownContext then
		bgReset(motif.optionbgdef.BGDef)
		runInputConfig(context)
	end
end

local function startInputCapture(tbl, item, all)
	local root = tbl.inputMenu
	local entries, seen = {}, {}
	local temporary = 0
	local function collect(menu)
		for i, row in ipairs(menu.items) do
			local field = row.itemname
			if t_profileButtons[field] then
				seen[field] = true
				table.insert(entries, {menu = menu, item = i, field = field})
			elseif menu.submenu[row.itemname] ~= nil then
				collect(menu.submenu[row.itemname])
			end
		end
	end
	if all then
		collect(root)
		-- Config all also covers bindings whose normal menu row the motif hides.
		for _, field in ipairs(t_profileFields) do
			if not seen[field] then
				table.insert(root.items, {itemname = field, displayname = field, paramname = ''})
				table.insert(entries, {menu = root, item = #root.items, field = field})
				temporary = temporary + 1
			end
		end
	else
		entries = {{menu = tbl, item = item, field = tbl.items[item].itemname}}
	end
	if #entries == 0 then
		return
	end
	tbl.inputContext.capture = {
		root = root, entries = entries, index = 1, released = false, temporary = temporary,
		values = inputProfile(tbl.inputSource.id), menu = tbl, item = item,
	}
	resetKey()
	resetTokenGuard()
end

function options.f_inputConfigAll(tbl)
	local row, sub, menu, item = findInputItem(tbl.inputMenu, 'configall')
	if row ~= nil then
		menu.inputContext = tbl.inputContext
		if menu == tbl.inputMenu then
			tbl.inputContext.parents = {}
		end
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		startInputCapture(menu, item, true)
	end
	resetKey()
end

function options.f_inputCapture(context)
	local sec = motif.option_info
	local capture = context.capture
	local root, source = capture.root, capture.root.inputSource
	local entry = capture.entries[capture.index]
	local finished = false
	if esc() then
		-- Esc drops this capture only.
		sndPlay(motif.Snd, sec.cancel.snd[1], sec.cancel.snd[2])
		esc(false)
		finished = true
	else
		local token = ''
		if getKey(sec.unbind.keycode) then
			token = '-'
		elseif source.kind == 'keyboard' then
			token = getKey()
		else
			local key = getJoystickKey(source.joy)
			if key == '' then
				capture.released = true
			elseif capture.released then
				token = key
				capture.released = false
			end
		end
		if token ~= '' and token ~= 'nil' then
			for _, field in ipairs(t_profileFields) do
				if capture.values[field] == token then
					capture.values[field] = '-'
				end
			end
			capture.values[entry.field] = token
			sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
			capture.index = capture.index + 1
			if capture.index > #capture.entries then
				applyProfile(root, capture.values)
				finished = true
			end
		end
	end
	resetKey()
	if finished then
		for i = 1, capture.temporary do
			table.remove(root.items)
		end
		updateProfileMenu(root, root, source, inputProfile(source.id))
		context.current, context.capture = capture.menu, nil
		context.skipInput = true
		capture.menu.inputFocus = capture.item
		resetTokenGuard()
	else
		updateProfileMenu(root, root, source, capture.values)
		entry = capture.entries[capture.index]
		context.current = entry.menu
		entry.menu.inputFocus = entry.item
		entry.menu.items[entry.item].vardisplay = sec.menu.valuename.presskey
	end
end

-- Input editor handlers. One handler serves both kinds: it reads the source
-- from the menu it is given.
options.t_itemname.configall = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		options.f_inputConfigAll(t)
	end
	return true
end

for _, field in ipairs(t_profileFields) do
	options.t_itemname[field] = function(t, item)
		local values = inputProfile(t.inputSource.id)
		if getKey(motif.option_info.unbind.keycode) or getInput(-1, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			values[field] = '-'
			applyProfile(t, values)
		elseif getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			startInputCapture(t, item, false)
		end
		resetKey()
		return true
	end
end

options.t_itemname.rumble = function(t, item)
	if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
		sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
		local values = inputProfile(t.inputSource.id)
		values.rumble = not values.rumble
		applyProfile(t, values)
	end
	return true
end

local function profileName(tbl, copy)
	local sec, context = motif.option_info, tbl.inputContext
	local values = inputProfile(tbl.inputSource.id)
	context.popup = main.f_drawInput(sec.textinput.TextSpriteData, sec.textinput.text.profilename,
		sec, context.background, sec.textinput.overlay.RectData, values.name, true)
	context.onClose = function(name)
		if name ~= '' then
			values.name = name
			if copy then
				-- The copy takes the edits; the original profile keeps its saved state.
				local root = tbl.inputMenu
				if applyProfile(tbl, root.savedBindings) and applyProfile(tbl, values, true) then
					root.finished = true
				end
			else
				applyProfile(tbl, values)
			end
		end
		esc(false)
	end
end

options.t_itemname.renameprofile = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		profileName(t, false)
	end
	return true
end

options.t_itemname.savecopy = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		profileName(t, true)
	end
	return true
end

options.t_itemname.deleteprofile = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		if inputDeleteProfile(t.inputSource.id) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			options.modified = true
			t.inputMenu.finished = true
			return false
		end
		inputPopup(t.inputContext, motif.option_info.textinput.text.profiledelete)
	end
	return true
end

options.t_itemname.save = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		t.inputMenu.finished = true
		return false
	end
	return true
end

options.t_itemname.cancel = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
		cancelProfile(t.inputMenu)
		t.inputMenu.finished = true
		return false
	end
	return true
end

options.t_itemname.newprofile = function(t, item)
	local sec, context = motif.option_info, t.inputContext
	if getInput(-1, sec.menu.done.key) then
		local source = context.source
		local menuSource = inputMenuSource()
		if source ~= nil then
			menuSource = source.id
		else
			-- Keyboard layouts have separate sources. Keep the existing one intact,
			-- preferring the one navigating the menu.
			for _, entry in ipairs(inputSources()) do
				if entry.connected and entry.kind == context.kind then
					if source == nil or entry.id == menuSource then
						source = entry
					end
				end
			end
		end
		if source == nil then
			return true
		end
		local previous = inputConfigSource(menuSource)
		if context.previousSource == nil then
			context.previousSource = previous
		end
		local id = inputNewProfile(source.id)
		if id == 0 then
			inputPopup(context, sec.textinput.text.profilesave)
			return true
		end
		sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
		options.modified = true
		for _, entry in ipairs(inputSources()) do
			if entry.id == id then
				options.f_editProfile(entry, nil, context, true)
				context.current.previousProfile = previous
				break
			end
		end
	end
	return true
end

options.t_itemname.profile = function(t, item)
	local sec = motif.option_info
	if getInput(-1, sec.menu.done.key) then
		local row = t.items[item]
		if row.profile ~= nil and not inputSetProfile(row.source.id, row.profile) then
			sndPlay(motif.Snd, sec.cancel.snd[1], sec.cancel.snd[2])
			return true
		end
		sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
		options.f_editProfile(row.source, t.submenu[row.itemname], t.inputContext)
	end
	return true
end

function options.f_inputCfg(kind, template, background)
	local sec = motif.option_info
	if main.pauseMenuActive then
		-- The menu already knows who opened Button Config; edit that source directly.
		for _, source in ipairs(inputSources()) do
			if source.id == inputMenuSource() then
				local context = {background = background or motif.optionbgdef, pause = true, parents = {}}
				options.f_editProfile(source, nil, context)
				runInputConfig(context)
				return
			end
		end
		return
	end
	if template == nil then
		local row
		row, template = findInputItem(options.menu, kind)
	end
	if template == nil then
		return
	end
	local context = {kind = kind, background = background or motif.optionbgdef, pause = main.pauseMenuActive, parents = {}}
	local function openMenu(selected)
		if selected ~= nil and context.previousSource == nil then
			context.previousSource = inputConfigSource(selected.id)
		elseif selected == nil and context.previousSource ~= nil then
			inputConfigSource(context.previousSource)
		end
		local tbl = main.f_tableCopy(template)
		if selected ~= nil then
			tbl.title = main.f_itemnameUpper(selected.name, sec.menu.title.uppercase)
		end
		populateInputMenu(tbl, kind, selected)
		context.current, context.source = tbl, selected
		-- Rebuild after editing so renamed, copied and deleted profiles stay current.
		context.openPicker = function() openMenu(selected) end
	end
	if kind == 'gamepad' then
		textImgReset(sec.title.TextSpriteData)
		textImgSetText(sec.title.TextSpriteData, template.title)
		local released = false
		inputPopup(context, sec.textinput.text.gamepad, function()
			local token, joy = getJoystickKey()
			if token == '' then
				released = true
			elseif released then
				for _, source in ipairs(inputSources()) do
					-- getJoystickKey returns a 1-based index; inputSources uses SDL's 0-based index.
					if source.connected and source.kind == 'gamepad' and source.joy == joy - 1 then
						return source
					end
				end
			end
		end, function(source)
			if source then
				openMenu(source)
			else
				context.closed = true
			end
		end)
	else
		openMenu()
	end
	runInputConfig(context)
end

return options

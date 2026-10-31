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
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
			return false
		end
		return true
	end,
	--Default Values
	['default'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Difficulty') < 8 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Difficulty', gameOption('Options.Difficulty') + 1)
			t.items[item].vardisplay = options.t_vardisplay['difficulty']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Difficulty') > 1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Difficulty', gameOption('Options.Difficulty') - 1)
			t.items[item].vardisplay = options.t_vardisplay['difficulty']()
			options.modified = true
		end
		return true
	end,
	--Time Limit
	['roundtime'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Time') < 1000 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Time', gameOption('Options.Time') + 1)
			t.items[item].vardisplay = options.t_vardisplay['roundtime']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Time') > -1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Time', gameOption('Options.Time') - 1)
			t.items[item].vardisplay = options.t_vardisplay['roundtime']()
			options.modified = true
		end
		return true
	end,
	--Language Setting
	['language'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and (main.f_tableLength(motif.languages) > 1 or motif.languages[gameOption('Config.Language')] == nil) then
			f_switchLanguage(1)
			t.items[item].vardisplay = options.t_vardisplay['language']()
		elseif getAction(-1, motif.option_info.menu.subtract.action) and (main.f_tableLength(motif.languages) > 1 or motif.languages[gameOption('Config.Language')] == nil) then
			f_switchLanguage(-1)
			t.items[item].vardisplay = options.t_vardisplay['language']()
		end
		return true
	end,
	--Life
	['lifemul'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Life') < 300 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Life', gameOption('Options.Life') + 10)
			t.items[item].vardisplay = options.t_vardisplay['lifemul']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Life') > 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Life', gameOption('Options.Life') - 10)
			t.items[item].vardisplay = options.t_vardisplay['lifemul']()
			options.modified = true
		end
		return true
	end,
	--Single VS Team Life
	['singlevsteamlife'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Team.SingleVsTeamLife') < 300 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Team.SingleVsTeamLife', gameOption('Options.Team.SingleVsTeamLife') + 10)
			t.items[item].vardisplay = options.t_vardisplay['singlevsteamlife']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Team.SingleVsTeamLife') > 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Team.SingleVsTeamLife', gameOption('Options.Team.SingleVsTeamLife') - 10)
			t.items[item].vardisplay = options.t_vardisplay['singlevsteamlife']()
			options.modified = true
		end
		return true
	end,
	-- Game Speed
	['gamespeed'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.GameSpeed') < 9 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.GameSpeed', gameOption('Options.GameSpeed') + 1)
			t.items[item].vardisplay = options.t_vardisplay['gamespeed']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.GameSpeed') > -9 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.GameSpeed', gameOption('Options.GameSpeed') - 1)
			t.items[item].vardisplay = options.t_vardisplay['gamespeed']()
			options.modified = true
		end
		return true
	end,
	--Rounds to Win (Single)
	['roundsnumsingle'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Match.Wins') < 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Match.Wins', gameOption('Options.Match.Wins') + 1)
			t.items[item].vardisplay = options.t_vardisplay['roundsnumsingle']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Match.Wins') > 1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Match.Wins', gameOption('Options.Match.Wins') - 1)
			t.items[item].vardisplay = options.t_vardisplay['roundsnumsingle']()
			options.modified = true
		end
		return true
	end,
	--Max Draw Games
	['maxdrawgames'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Match.MaxDrawGames') < 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Match.MaxDrawGames', gameOption('Options.Match.MaxDrawGames') + 1)
			t.items[item].vardisplay = options.t_vardisplay['maxdrawgames']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Match.MaxDrawGames') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Match.MaxDrawGames', gameOption('Options.Match.MaxDrawGames') - 1)
			t.items[item].vardisplay = options.t_vardisplay['maxdrawgames']()
			options.modified = true
		end
		return true
	end,
	--Credits
	['credits'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Credits') < 99 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Credits', gameOption('Options.Credits') + 1)
			t.items[item].vardisplay = options.t_vardisplay['credits']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Credits') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Credits', gameOption('Options.Credits') - 1)
			t.items[item].vardisplay = options.t_vardisplay['credits']()
			options.modified = true
		end
		return true
	end,
	--Arcade Palette
	['aipalette'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Tag.Match.Wins') < 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Tag.Match.Wins', gameOption('Options.Tag.Match.Wins') + 1)
			t.items[item].vardisplay = options.t_vardisplay['roundsnumtag']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Tag.Match.Wins') > 1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Tag.Match.Wins', gameOption('Options.Tag.Match.Wins') - 1)
			t.items[item].vardisplay = options.t_vardisplay['roundsnumtag']()
			options.modified = true
		end
		return true
	end,
	--Partner KOed Lose
	['losekotag'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Tag.Min') < gameOption('Options.Tag.Max') then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Tag.Min', gameOption('Options.Tag.Min') + 1)
			t.items[item].vardisplay = options.t_vardisplay['mintag']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Tag.Min') > 2 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Tag.Min', gameOption('Options.Tag.Min') - 1)
			t.items[item].vardisplay = options.t_vardisplay['mintag']()
			options.modified = true
		end
		return true
	end,
	--Max Tag Chars
	['maxtag'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Tag.Max') < 4 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Tag.Max', gameOption('Options.Tag.Max') + 1)
			t.items[item].vardisplay = options.t_vardisplay['maxtag']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Tag.Max') > gameOption('Options.Tag.Min') then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Tag.Max', gameOption('Options.Tag.Max') - 1)
			t.items[item].vardisplay = options.t_vardisplay['maxtag']()
			options.modified = true
		end
		return true
	end,
	--Rounds to Win (Simul)
	['roundsnumsimul'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Simul.Match.Wins') < 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Simul.Match.Wins', gameOption('Options.Simul.Match.Wins') + 1)
			t.items[item].vardisplay = options.t_vardisplay['roundsnumsimul']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Simul.Match.Wins') > 1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Simul.Match.Wins', gameOption('Options.Simul.Match.Wins') - 1)
			t.items[item].vardisplay = options.t_vardisplay['roundsnumsimul']()
			options.modified = true
		end
		return true
	end,
	--Simul Player KOed Lose
	['losekosimul'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Simul.Min') < gameOption('Options.Simul.Max') then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Simul.Min', gameOption('Options.Simul.Min') + 1)
			t.items[item].vardisplay = options.t_vardisplay['minsimul']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Simul.Min') > 2 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Simul.Min', gameOption('Options.Simul.Min') - 1)
			t.items[item].vardisplay = options.t_vardisplay['minsimul']()
			options.modified = true
		end
		return true
	end,
	--Max Simul Chars
	['maxsimul'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Simul.Max') < 4 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Simul.Max', gameOption('Options.Simul.Max') + 1)
			t.items[item].vardisplay = options.t_vardisplay['maxsimul']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Simul.Max') > gameOption('Options.Simul.Min') then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Simul.Max', gameOption('Options.Simul.Max') - 1)
			t.items[item].vardisplay = options.t_vardisplay['maxsimul']()
			options.modified = true
		end
		return true
	end,
	--Turns Recovery Base
	['turnsrecoverybase'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Turns.Recovery.Base') < 100 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Recovery.Base', gameOption('Options.Turns.Recovery.Base') + 0.5)
			t.items[item].vardisplay = options.t_vardisplay['turnsrecoverybase']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Turns.Recovery.Base') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Recovery.Base', gameOption('Options.Turns.Recovery.Base') - 0.5)
			t.items[item].vardisplay = options.t_vardisplay['turnsrecoverybase']()
			options.modified = true
		end
		return true
	end,
	--Turns Recovery Bonus
	['turnsrecoverybonus'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Turns.Recovery.Bonus') < 100 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Recovery.Bonus', gameOption('Options.Turns.Recovery.Bonus') + 0.5)
			t.items[item].vardisplay = options.t_vardisplay['turnsrecoverybonus']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Turns.Recovery.Bonus') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Recovery.Bonus', gameOption('Options.Turns.Recovery.Bonus') - 0.5)
			t.items[item].vardisplay = options.t_vardisplay['turnsrecoverybonus']()
			options.modified = true
		end
		return true
	end,
	--Min Turns Chars
	['minturns'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Turns.Min') < gameOption('Options.Turns.Max') then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Min', gameOption('Options.Turns.Min') + 1)
			t.items[item].vardisplay = options.t_vardisplay['minturns']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Turns.Min') > 1 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Min', gameOption('Options.Turns.Min') - 1)
			t.items[item].vardisplay = options.t_vardisplay['minturns']()
			options.modified = true
		end
		return true
	end,
	--Max Turns Chars
	['maxturns'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Options.Turns.Max') < 8 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Max', gameOption('Options.Turns.Max') + 1)
			t.items[item].vardisplay = options.t_vardisplay['maxturns']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Options.Turns.Max') > gameOption('Options.Turns.Min') then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Options.Turns.Max', gameOption('Options.Turns.Max') - 1)
			t.items[item].vardisplay = options.t_vardisplay['maxturns']()
			options.modified = true
		end
		return true
	end,
	--Renderer (submenu)
	['renderer'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Video.MSAA') < 32 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Video.MSAA') == 0 then
				modifyGameOption('Video.MSAA', 2)
			else
				modifyGameOption('Video.MSAA', gameOption('Video.MSAA') * 2)
			end
			t.items[item].vardisplay = options.t_vardisplay['msaa']()
			options.modified = true
			options.needReload = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Video.MSAA') > 1 then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Sound.MasterVolume') < 200 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.MasterVolume', gameOption('Sound.MasterVolume') + 1)
			t.items[item].vardisplay = options.t_vardisplay['mastervolume']()
			options.modified = true
			updateVolume()
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Sound.MasterVolume') > 0 then
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
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Sound.BGMVolume') < 100 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.BGMVolume', gameOption('Sound.BGMVolume') + 1)
			t.items[item].vardisplay = options.t_vardisplay['bgmvolume']()
			options.modified = true
			updateVolume()
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Sound.BGMVolume') > 0 then
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
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Sound.WavVolume') < 100 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.WavVolume', gameOption('Sound.WavVolume') + 1)
			t.items[item].vardisplay = options.t_vardisplay['sfxvolume']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Sound.WavVolume') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.WavVolume', gameOption('Sound.WavVolume') - 1)
			t.items[item].vardisplay = options.t_vardisplay['sfxvolume']()
			options.modified = true
		end
		return true
	end,
	--Audio Ducking
	['audioducking'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Sound.PanningRange') < 100 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.PanningRange', gameOption('Sound.PanningRange') + 1)
			t.items[item].vardisplay = options.t_vardisplay['panningrange']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Sound.PanningRange') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Sound.PanningRange', gameOption('Sound.PanningRange') - 1)
			t.items[item].vardisplay = options.t_vardisplay['panningrange']()
			options.modified = true
		end
		return true
	end,
	-- Button Mapping
	['buttonmapping'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			options.f_inputCfg('buttonmapping', t.submenu[t.items[item].itemname])
			textImgSetText(motif.option_info.title.TextSpriteData, t.title)
		end
		return true
	end,
	-- Controller Setup
	['controllersetup'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			options.f_inputCfg('controllersetup', t.submenu[t.items[item].itemname])
			textImgSetText(motif.option_info.title.TextSpriteData, t.title)
		end
		return true
	end,
	-- Rumble
	['rumble'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Input.ControllerRumble', not gameOption('Input.ControllerRumble'))
			t.items[item].vardisplay = options.t_vardisplay['rumble']()
			options.modified = true
		end
		return true
	end,
	--Default
	['inputdefault'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			inputRestoreDefaults()
			options.modified = true
		end
		return true
	end,
	--Players
	['players'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Config.Players') < 8 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Config.Players', math.min(8, gameOption('Config.Players') + 2))
			t.items[item].vardisplay = options.t_vardisplay['players']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Config.Players') > 2 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Config.Players', math.max(2, gameOption('Config.Players') - 2))
			t.items[item].vardisplay = options.t_vardisplay['players']()
			options.modified = true
		end
		return true
	end,
	--Debug Keys
	['debugkeys'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.add.action) and gameOption('Netplay.Rollback.FrameDelay') < 10 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Netplay.Rollback.FrameDelay', gameOption('Netplay.Rollback.FrameDelay') + 1)
			t.items[item].vardisplay = options.t_vardisplay['framedelay']()
			options.modified = true
		elseif getAction(-1, motif.option_info.menu.subtract.action) and gameOption('Netplay.Rollback.FrameDelay') > 0 then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			modifyGameOption('Netplay.Rollback.FrameDelay', gameOption('Netplay.Rollback.FrameDelay') - 1)
			t.items[item].vardisplay = options.t_vardisplay['framedelay']()
			options.modified = true
		end
		return true
	end,
	--Save and Return
	['savereturn'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
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
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			if options.modified then
				options.f_saveCfg(options.needReload)
			end
		end
		return true
	end,
}

-- Rows that show a value but open a list or text prompt instead of changing it.
options.t_hintDefault = {
	aspectratio = true,
	renderer = true,
	resolution = true,
	shaders = true,
	portchange = true,
	rollbackport = true,
}

-- Picks the hint for an Options row.
function options.f_hintContext(row)
	if row.hintContext ~= nil then
		return row.hintContext
	end
	local name = row.itemname
	if options.t_hintDefault[name] then
		return 'default'
	end
	if options.t_vardisplay[name] ~= nil then
		return 'adjust'
	end
	return 'default'
end

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
				local hintContext = options.f_hintContext(t[item])
				if context ~= nil and context.capture ~= nil then
					hintContext = 'capture'
				end
				local menuBackground = motif.optionbgdef
				if context ~= nil and context.background ~= nil then
					menuBackground = context.background
				end
				main.f_menuCommonDraw(t, item, cursorPosY, moveTxt, motif.option_info,
					menuBackground, context ~= nil, {hintContext = hintContext})
			end
			-- While fading, ignore normal option-menu inputs, but still allow ESC / menu-cancel to skip the fade.
			if context ~= nil and (context.capture ~= nil or context.popup ~= nil or context.skipInput) then
				cursorPosY, moveTxt, item = main.f_menuCommonCalc(t, item, cursorPosY, moveTxt, motif.option_info, motif.option_info.cursor, 0)
			elseif fadeActive() then
				if esc() or getAction(-1, motif.option_info.menu.cancel.action) then
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
				elseif esc() or getAction(-1, motif.option_info.menu.cancel.action) then
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
				elseif options.t_itemname[t[item].itemname] ~= nil then
					if not options.t_itemname[t[item].itemname](tbl, item, cursorPosY, moveTxt) then
						break
					end
				elseif getAction(-1, motif.option_info.menu.done.action) then
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
	['rumble'] = function()
		return options.f_boolDisplay(gameOption('Input.ControllerRumble'))
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
					if getAction(-1, motif.option_info.menu.done.action) then
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
				if getAction(-1, motif.option_info.menu.done.action) then
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
					if getAction(-1, motif.option_info.menu.done.action) then
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
					if getAction(-1, motif.option_info.menu.done.action) then
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
					if getAction(-1, motif.option_info.menu.done.action) then
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
		excluded = {gl33 = true}
	elseif runtimeOS == 'ios' then
		excluded = {gl33 = true, gles32 = true}
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
--; INPUT SETTINGS
--;===========================================================
-- Button Mapping: physical input -> virtual control. Controller Setup: virtual control -> engine button.
-- Both editors change a draft and apply it when the editor is left.
local t_virtualControls = {'DP_U', 'DP_D', 'DP_L', 'DP_R', 'A', 'B', 'X', 'Y', 'START', 'BACK',
	'LB', 'RB', 'LT', 'RT', 'LS_Y-', 'LS_Y+', 'LS_X-', 'LS_X+', 'RS_Y-', 'RS_Y+', 'RS_X-', 'RS_X+', 'LS', 'RS'}
local t_engineButtons = {'up', 'down', 'left', 'right', 'a', 'b', 'c', 'x', 'y', 'z', 'start', 'd', 'w', 'menu', 'l', 'r'}

-- Menu rows of the mapping editor are named after the virtual controls: "LS_Y-" -> "vclsyn".
-- Menu item names can't use "_", it separates submenus.
local t_mappingRows = {}
for _, control in ipairs(t_virtualControls) do
	t_mappingRows['vc' .. control:lower():gsub('_', ''):gsub('%-', 'n'):gsub('%+', 'p')] = control
end
local t_setupRows = {}
for _, button in ipairs(t_engineButtons) do
	t_setupRows[button] = button
end
for path in pairs(motif.option_info.menu.itemname) do
	local button = path:match('controllersetup_(.+)$')
	if button ~= nil then
		button = button:match('([^_]+)$')
		if inputIsSetupButton(button) then
			t_setupRows[button] = button
		end
	end
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

local function inputPopup(context, text, getValue, onClose, hintContext)
	local sec = motif.option_info
	context.popup = main.f_warning(text, sec, context.background, nil, sec.title.TextSpriteData, nil,
		sec.cancel.snd, sec.cursor.done.snd, getValue, true, hintContext)
	context.onClose = onClose
end

local function deviceById(id)
	for _, device in ipairs(inputGetDevices()) do
		if device.id == id then
			return device
		end
	end
end

-- Value shown for a binding: a key name, a virtual control name, or the "no key" text.
-- With glyphs on, keys show as keyboard art and controls as controller art.
local function bindingDisplay(token, kind)
	if token == nil or token == '' or token == '-' then
		return motif.option_info.menu.valuename.nokey
	end
	if not motif.option_info.useglyphs then
		return token
	elseif kind == 'keyboard' then
		return '<key:' .. token .. '>'
	end
	return '<virtual:' .. token .. '>'
end

-- Refreshes the values shown in an editor and links its submenus to the root.
local function updateEditor(tbl, root)
	if tbl.inputMenu == nil then
		tbl.loop = options.f_createMenu(tbl, false)
	end
	tbl.inputContext = root.inputContext
	tbl.device = root.device
	tbl.inputMenu = root
	for _, row in ipairs(tbl.items) do
		local field = row.itemname
		row.glyphDevice = root.device and root.device.id or nil
		row.hintContext = 'editor'
		if root.kind == 'mapping' and t_mappingRows[field] then
			row.hintContext = 'binding'
			row.vardisplay = bindingDisplay(root.draft[t_mappingRows[field]], root.device.kind)
		elseif root.kind == 'setup' and t_setupRows[field] then
			row.hintContext = 'binding'
			row.vardisplay = bindingDisplay(root.draft[field], 'virtual')
		elseif field == 'activesetup' then
			row.hintContext = 'setup'
			row.vardisplay = root.draft.name
		elseif field == 'device' then
			row.hintContext = 'adjust'
			row.vardisplay = root.device.name
		elseif field == 'glyphs' and root.device ~= nil then
			row.hintContext = 'adjust'
			for _, set in ipairs(inputGetGlyphPacks(root.device.id)) do
				if set.selected then
					row.vardisplay = set.glyph .. ' ' .. set.name
					break
				end
			end
		end
	end
	for _, sub in pairs(tbl.submenu) do
		updateEditor(sub, root)
	end
end

-- True if another keyboard controller maps this key. Keyboard controllers can't share
-- keys, since two players may use them at the same time.
local function keyTaken(device, key)
	if device.kind ~= 'keyboard' then
		return false
	end
	for _, other in ipairs(inputGetDevices()) do
		if other.kind == 'keyboard' and other.id ~= device.id then
			for _, k in pairs(inputGetMapping(other.id)) do
				if k == key then
					return true
				end
			end
		end
	end
	return false
end

-- True if two setups bind the same controls.
local function sameSetup(a, b)
	for _, values in ipairs({a, b}) do
		for button in pairs(values) do
			if button ~= 'name' and button ~= 'id' and (a[button] or '-') ~= (b[button] or '-') then
				return false
			end
		end
	end
	return true
end

-- A device edits its own setup. Before the first change to a shipped setup, or to one
-- another device uses, it is copied to "Custom N" and the device switches to the copy.
local function ownSetup(root)
	if root.kind ~= 'setup' or inputSetupEditable(root.setupId, root.device.id) then
		return
	end
	local values = main.f_tableCopy(root.saved)
	values.name = motif.option_info.menu.valuename.custom
	local id = inputNewSetup(values, motif.option_info.menu.valuename.copy, root.device.id)
	if id ~= '' then
		options.modified = true
		root.copyOf = root.copyOf or root.setupId
		root.setupId = id
		root.draft.name = inputGetSetup(id).name
	end
end

-- Drops a copy made by this editor and goes back to setupId.
local function undoCopy(root, setupId)
	if root.kind ~= 'setup' or root.setupId == setupId then
		return
	end
	inputDeleteSetup(root.setupId)
	inputSetDeviceSetup(root.device.id, setupId)
	root.setupId = setupId
	if setupId == root.copyOf then
		root.copyOf = nil
	end
end

-- Applies the draft. Returns false and explains why if it can't be saved.
local function applyEditor(root)
	local sec = motif.option_info
	local ok = true
	if root.kind == 'mapping' then
		ok = inputSaveMapping(root.device.id, root.draft)
	else
		ok = inputSaveSetup(root.setupId, root.draft)
	end
	if not ok then
		inputPopup(root.inputContext, root.kind == 'mapping' and sec.textinput.text.mappingsave or sec.textinput.text.setupsave)
		return false
	end
	options.modified = true
	return true
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
		-- The popup overlays the options screen, so draw that background behind it.
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
	-- An unplugged gamepad switches the mapping editor to another device.
	local device = context.current and context.current.inputMenu and context.current.inputMenu.kind == 'mapping' and context.current.device
	if device and device.kind == 'gamepad' and not getJoystickPresent(device.joy) then
		context.closed = not options.f_editMapping(nil, context)
	end
	if not context.closed and context.mapping ~= nil then
		options.f_gamepadMappingFrame(context)
		if not context.closed then
			context.skipInput = false
			return true
		end
	end
	if not context.closed and context.current ~= nil then
		drew = true
		if context.capture ~= nil then
			options.f_inputCapture(context)
		end
		local tbl = context.current
		tbl.inputContext = context
		if not tbl.loop(context) then
			local root = tbl.inputMenu
			-- Leaving an editor keeps its edits, like Confirm and Exit.
			if root ~= nil and tbl == root and not root.finished and not applyEditor(root) then
				-- Stay in the editor, the popup explains the problem.
			elseif root ~= nil and (tbl == root or root.finished) then
				context.current, context.parents = nil, {}
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
		main.f_menuSnap(sec)
		main.f_menuItemBgAnimReset(sec)
		resetKey()
		resetTokenGuard()
		esc(false)
		options.editingInput = false
		return false, drew
	end
	return true
end

local function runInputConfig(context)
	options.editingInput = true
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
		-- Show the last editor frame on its own, but only if it drew something.
		if drew then
			refresh()
		end
	end
	if not context.pause then options.editingInput = false end
end

-- Devices Button Mapping can edit: keyboard controllers (except on phones) and connected gamepads.
local function mappingDevices()
	local platform = getRuntimeOS()
	local keyboards = platform ~= 'android' and platform ~= 'ios'
	local t = {}
	for _, device in ipairs(inputGetDevices()) do
		if (device.kind == 'keyboard' and keyboards) or (device.kind == 'gamepad' and device.connected) then
			table.insert(t, device)
		end
	end
	return t
end

-- Opens the Button Mapping editor of a device. Without one, or if it can't be
-- edited, the first device is used. Returns false if there is none.
function options.f_editMapping(device, context)
	local row, template = findInputItem(options.menu, 'buttonmapping')
	local devices = mappingDevices()
	if template == nil or #devices == 0 then
		return false
	end
	local found = devices[1]
	for _, v in ipairs(devices) do
		if device ~= nil and v.id == device.id then
			found = v
		end
	end
	device = found
	local tbl = main.f_tableCopy(template)
	tbl.kind = 'mapping'
	tbl.device = device
	tbl.draft = inputGetMapping(device.id)
	tbl.saved = main.f_tableCopy(tbl.draft)
	tbl.inputContext = context
	updateEditor(tbl, tbl)
	context.current, context.parents = tbl, {}
	resetKey()
	resetTokenGuard()
	return true
end

-- Loads a setup into the Controller Setup editor.
local function loadSetup(root, setupId)
	local values = inputGetSetup(setupId)
	if values == nil then
		return false
	end
	root.setupId, root.copyOf = setupId, nil
	root.draft = values
	root.saved = main.f_tableCopy(values)
	updateEditor(root, root)
	return true
end

-- Opens the Controller Setup editor on the setup this device uses.
function options.f_editSetup(device, context)
	local row, template = findInputItem(options.menu, 'controllersetup')
	if template == nil or device == nil then
		return false
	end
	local tbl = main.f_tableCopy(template)
	tbl.kind = 'setup'
	tbl.device = device
	-- Every device is a virtual controller, so Options reads them all. Side select and
	-- the pause menu only read the player who opened the editor.
	tbl.readDevice = context.allDevices and 0 or device.id
	tbl.inputContext = context
	if not loadSetup(tbl, device.setup) then
		return false
	end
	context.current, context.parents = tbl, {}
	resetKey()
	resetTokenGuard()
	return true
end

local function configAllButtons(kind)
	local mapping = kind == 'mapping'
	local names = mapping and motif.option_info.buttonmapping.configall or motif.option_info.controllersetup.configall
	local buttons = {}
	for _, name in ipairs(names) do
		buttons[mapping and name:upper() or name:lower()] = true
	end
	return buttons
end

-- Config All follows menu order and skips hidden or unlisted buttons.
local function startInputCapture(tbl, item, all)
	local root = tbl.inputMenu
	local rows = root.kind == 'mapping' and t_mappingRows or t_setupRows
	local allowed = all and configAllButtons(root.kind)
	local entries = {}
	local function collect(menu)
		for i, row in ipairs(menu.items) do
			if rows[row.itemname] and allowed[rows[row.itemname]] then
				table.insert(entries, {menu = menu, item = i, field = rows[row.itemname]})
			elseif menu.submenu[row.itemname] ~= nil then
				collect(menu.submenu[row.itemname])
			end
		end
	end
	if all then
		collect(root)
	else
		entries = {{menu = tbl, item = item, field = rows[tbl.items[item].itemname]}}
	end
	if #entries == 0 then
		return
	end
	-- Cancel puts back the bindings from before the capture, also after Config all assigned some.
	tbl.inputContext.capture = {root = root, entries = entries, index = 1, released = false, menu = tbl, item = item,
		saved = main.f_tableCopy(root.draft), setupId = root.setupId}
	resetKey()
	resetTokenGuard()
end

-- Reads one input for a capture, counted when let go, so holding cancel never binds it.
-- 'virtual' reads virtual controls (id 0: every device), 'keyboard'/'gamepad' physical ones.
local function releasedInput(state, id, kind)
	local pressed, held = '', ''
	if kind == 'virtual' then
		pressed = inputGetVirtualPressed(id)
		if state.pending ~= nil and inputDeviceActionTime(id, {state.pending}) > 0 then
			held = state.pending
		end
	elseif kind == 'keyboard' then
		pressed = getKey()
		if pressed == 'nil' or pressed == 'ESCAPE' then
			pressed = ''
		end
		if state.pending ~= nil and inputDeviceActionTime(id, {'KEY:' .. state.pending}) > 0 then
			held = state.pending
		end
	else
		-- Only a press after everything was let go, so one press can't fill two rows.
		held = inputGetPhysicalHeld(id)
		if held == '' then
			state.released = true
		elseif state.released and state.pending == nil then
			pressed = held
		end
	end
	if state.pending == nil then
		if pressed ~= '' then
			state.pending, state.released = pressed, false
		end
	elseif held ~= state.pending then
		local token = state.pending
		state.pending = nil
		return token
	end
	return ''
end

-- One frame of capture. The cursor moves to each captured row in turn.
function options.f_inputCapture(context)
	local sec = motif.option_info
	local capture = context.capture
	local root = capture.root
	local device = root.device
	local entry = capture.entries[capture.index]
	local finished = false
	-- Holding cancel (through the saved mapping) leaves, so every control can still be bound.
	if esc() or inputDeviceActionTime(root.readDevice or device.id, sec.menu.cancel.action) >= sec.holdtime then
		sndPlay(motif.Snd, sec.cancel.snd[1], sec.cancel.snd[2])
		esc(false)
		undoCopy(root, capture.setupId)
		root.draft = capture.saved
		finished = true
	else
		-- Setups bind virtual controls, read through the device's Button Mapping.
		local token
		if root.kind == 'setup' then
			token = releasedInput(capture, root.readDevice, 'virtual')
		else
			token = releasedInput(capture, device.id, device.kind)
		end
		if token ~= '' then
			-- A key of the other keyboard controller stays there and leaves this control unused.
			local taken = root.kind == 'mapping' and keyTaken(device, token)
			if taken then
				token = '-'
			end
			-- An input binds one row only; the row that had it becomes unused.
			for field, value in pairs(root.draft) do
				if field ~= 'name' and field ~= 'id' and field ~= entry.field and token ~= '-' and value == token then
					ownSetup(root)
					root.draft[field] = '-'
				end
			end
			if root.draft[entry.field] ~= token then
				ownSetup(root)
			end
			root.draft[entry.field] = token
			local snd = taken and sec.cancel.snd or sec.cursor.done.snd
			sndPlay(motif.Snd, snd[1], snd[2])
			capture.index = capture.index + 1
			finished = capture.index > #capture.entries
		end
	end
	resetKey()
	updateEditor(root, root)
	if finished then
		context.current, context.capture = capture.menu, nil
		context.skipInput = true
		capture.menu.inputFocus = capture.item
		resetTokenGuard()
	else
		entry = capture.entries[capture.index]
		context.current = entry.menu
		entry.menu.inputFocus = entry.item
		entry.menu.items[entry.item].vardisplay = sec.menu.valuename.presskey
	end
end

--;===========================================================
--; MAPPING SCREEN ([Mapping Info])
--;===========================================================
-- The controller diagram, used by Config All (assigns each control) and Test Input (shows them).
local t_mappingArt = {
	DP_U = 'dpup', DP_D = 'dpdown', DP_L = 'dpleft', DP_R = 'dpright',
	A = 'a', B = 'b', X = 'x', Y = 'y', START = 'start', BACK = 'back',
	LB = 'leftshoulder', RB = 'rightshoulder', LT = 'lefttrigger', RT = 'righttrigger',
	['LS_Y-'] = 'leftynegative', ['LS_Y+'] = 'leftypositive', ['LS_X-'] = 'leftxnegative', ['LS_X+'] = 'leftxpositive',
	['RS_Y-'] = 'rightynegative', ['RS_Y+'] = 'rightypositive', ['RS_X-'] = 'rightxnegative', ['RS_X+'] = 'rightxpositive',
	LS = 'leftstick', RS = 'rightstick',
}

-- Leave the mapping screen.
local function mappingClose(context, finish)
	if context.pause then
		finish()
		return
	end
	fadeOutInit(motif.mapping_info.fadeout.FadeData)
	context.mapping.finish = function()
		finish()
		fadeInInit(motif.option_info.fadein.FadeData)
	end
end

-- Mapping editor rows name the virtual controls, so the guided screen reuses those names.
local function mappingLabels(root)
	local labels = {}
	local function collect(menu)
		for _, row in ipairs(menu.items) do
			if t_mappingRows[row.itemname] then
				labels[t_mappingRows[row.itemname]] = row.displayname
			elseif menu.submenu[row.itemname] ~= nil then
				collect(menu.submenu[row.itemname])
			end
		end
	end
	collect(root)
	return labels
end

-- Opens the mapping screen for a Button Mapping editor root.
-- The screen uses [Mapping Info] and [MappingBgDef], or the pause background in a match.
function options.f_mappingScreen(context, mode, root)
	local mapping = {mode = mode, index = 1, root = root, name = root.device.name, device = root.device.id}
	local allowed = configAllButtons('mapping')
	mapping.controls = {}
	for _, control in ipairs(t_virtualControls) do
		if allowed[control] then
			table.insert(mapping.controls, control)
		end
	end
	if mode == 'guided' and #mapping.controls == 0 then
		return
	end
	mapping.labels = mappingLabels(root)
	mapping.saved = main.f_tableCopy(root.draft)
	if mode == 'test' then
		mapping.index = 0
	end
	context.mapping = mapping
	if not context.pause then
		bgReset(motif.mappingbgdef.BGDef)
		fadeInInit(motif.mapping_info.fadein.FadeData)
	end
	resetKey()
	resetTokenGuard()
end

function options.f_gamepadMappingFrame(context)
	local sec = motif.option_info
	local cfg, mapping = motif.mapping_info, context.mapping
	local background = context.background
	if not context.pause then
		background = motif.mappingbgdef
		clearColor(background.bgclearcolor[1], background.bgclearcolor[2], background.bgclearcolor[3])
	end
	local name = mapping.controls[mapping.index]
	local control = name and cfg.control[t_mappingArt[name]]
	local state = mapping.mode == 'test' and 'test' or 'default'
	local pressed = {}
	if mapping.mode == 'test' then
		local held = inputGetVirtualState(mapping.device, mapping.root.draft)
		for _, vc in ipairs(t_virtualControls) do
			pressed[t_mappingArt[vc]] = held[vc]
		end
	end
	local rear = control ~= nil and control.rear
	for key, held in pairs(pressed) do
		if held and cfg.control[key] ~= nil and cfg.control[key].rear then
			rear = true
		end
	end
	bgDraw(background.BGDef, 0)
	rectDraw(cfg.overlay.RectData)
	textImgReset(cfg.title.TextSpriteData)
	textImgSetText(cfg.title.TextSpriteData, main.f_itemnameUpper(mapping.name, cfg.title.uppercase))
	textImgDraw(cfg.title.TextSpriteData)
	if rear then
		main.f_animPosDraw(cfg.back.AnimData)
		main.f_animPosDraw(cfg.rotate.AnimData)
	else
		main.f_animPosDraw(cfg.front.AnimData)
	end
	for _, vc in ipairs(t_virtualControls) do
		local key = t_mappingArt[vc]
		local art = cfg.control[key]
		if art ~= nil and (vc == name or (pressed[key] and art.rear == rear)) then
			main.f_animPosDraw(art.AnimData)
		end
	end
	local label = mapping.labels[name or ''] or name or ''
	-- Config all has nothing left to ask while it fades out.
	if name ~= nil or state ~= 'default' then
		textImgReset(cfg.prompt.TextSpriteData)
		textImgSetText(cfg.prompt.TextSpriteData, main.f_formatBySpec(cfg.prompt.text[state], {
			s = label, i = mapping.index,
		}))
		textImgDraw(cfg.prompt.TextSpriteData)
	end
	bgDraw(background.BGDef, 1)
	-- A refused key shows hint.warn until another key is accepted or the control is left.
	if mapping.warn ~= nil then
		textImgReset(cfg.hint.warn.TextSpriteData)
		textImgSetText(cfg.hint.warn.TextSpriteData, cfg.hint.warn.text[mapping.warn] or '')
		textImgDraw(cfg.hint.warn.TextSpriteData)
	else
		main.f_hintDraw(cfg.hint, state, nil, mapping.device)
	end
	if mapping.finish ~= nil then
		if not fadeActive() then
			mapping.finish()
		end
		return
	end
	if context.skipInput or fadeActive() then
		return
	end
	local function close(finish)
		mappingClose(context, function()
			context.mapping = nil
			if finish ~= nil then
				finish()
			end
			context.closed = context.current == nil
		end)
	end
	-- Held actions read the device being set up through its saved mapping, and the device
	-- running the menu.
	local menuDevice = inputGetMenuDevice()
	local function holdTime(action)
		local time = inputDeviceActionTime(mapping.device, action)
		if menuDevice ~= mapping.device then
			time = math.max(time, inputDeviceActionTime(menuDevice, action))
		end
		return time
	end
	local function held(action)
		return holdTime(action) == cfg.holdtime
	end
	if esc() or holdTime(cfg.cancel.action) >= cfg.holdtime then
		-- Cancel keeps the mapping as it was before the screen opened.
		sndPlay(motif.Snd, cfg.cancel.snd[1], cfg.cancel.snd[2])
		esc(false)
		if mapping.mode == 'guided' then
			mapping.root.draft = mapping.saved
			updateEditor(mapping.root, mapping.root)
		end
		close()
	elseif mapping.mode == 'test' then
		-- Read-only: inputs only highlight the diagram.
	elseif held(cfg.previous.action) and mapping.index > 1 then
		mapping.index, mapping.pending, mapping.released, mapping.warn = mapping.index - 1, nil, false, nil
		sndPlay(motif.Snd, cfg.move.snd[1], cfg.move.snd[2])
	elseif name ~= nil and held(cfg.skip.action) then
		mapping.index, mapping.pending, mapping.released, mapping.warn = mapping.index + 1, nil, false, nil
		sndPlay(motif.Snd, cfg.move.snd[1], cfg.move.snd[2])
	elseif name ~= nil then
		-- Guided Button Mapping: one physical input for each virtual control.
		local root = mapping.root
		local token = releasedInput(mapping, mapping.device, root.device.kind)
		if token ~= '' then
			-- A key of the other keyboard controller is refused: wait for another one.
			if keyTaken(root.device, token) then
				mapping.warn = 'keytaken'
				sndPlay(motif.Snd, cfg.cancel.snd[1], cfg.cancel.snd[2])
			else
				for _, vc in ipairs(t_virtualControls) do
					if root.draft[vc] == token then
						root.draft[vc] = '-'
					end
				end
				root.draft[name] = token
				mapping.index, mapping.warn = mapping.index + 1, nil
				sndPlay(motif.Snd, cfg.done.snd[1], cfg.done.snd[2])
			end
		end
	end
	if mapping.mode == 'guided' and mapping.index > #mapping.controls and mapping.finish == nil then
		updateEditor(mapping.root, mapping.root)
		close()
	end
	resetKey()
	resetTokenGuard()
end

-- Editor rows. One handler serves both editors: it reads the kind from the menu root.
-- Button Mapping uses the guided controller screen; Controller Setup moves the cursor row by row.
options.t_itemname.configall = function(t, item)
	if getAction(-1, motif.option_info.menu.done.action) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		if t.inputMenu.kind == 'mapping' then
			options.f_mappingScreen(t.inputContext, 'guided', t.inputMenu)
		else
			startInputCapture(t, item, true)
		end
	end
	resetKey()
	return true
end

options.t_itemname.testinput = function(t, item)
	if getAction(-1, motif.option_info.menu.done.action) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		options.f_mappingScreen(t.inputContext, 'test', t.inputMenu)
	end
	return true
end

local function bindingRow(t, item)
	local sec = motif.option_info
	local root = t.inputMenu
	if getAction(-1, sec.menu.subtract.action) then
		sndPlay(motif.Snd, sec.cursor.move.snd[1], sec.cursor.move.snd[2])
		local rows = root.kind == 'mapping' and t_mappingRows or t_setupRows
		local field = rows[t.items[item].itemname]
		if (root.draft[field] or '-') ~= '-' then
			ownSetup(root)
		end
		root.draft[field] = '-'
		updateEditor(root, root)
	elseif getAction(-1, sec.menu.done.action) then
		sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
		startInputCapture(t, item, false)
	end
	resetKey()
	return true
end
for name in pairs(t_mappingRows) do
	options.t_itemname[name] = bindingRow
end
for name in pairs(t_setupRows) do
	options.t_itemname[name] = bindingRow
end

-- The glyph pack belongs to the device rather than the mapping, so it is saved straight away.
options.t_itemname.glyphs = function(t, item)
	local step = 0
	if getAction(-1, motif.option_info.menu.add.action) then
		step = 1
	elseif getAction(-1, motif.option_info.menu.subtract.action) then
		step = -1
	end
	if step ~= 0 then
		sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
		local root = t.inputMenu
		local sets = inputGetGlyphPacks(root.device.id)
		local current = 1
		for i, set in ipairs(sets) do
			if set.selected then
				current = i
				break
			end
		end
		if inputSetGlyphPack(root.device.id, sets[(current - 1 + step) % #sets + 1].id) then
			updateEditor(root, root)
		end
	end
	return true
end

-- Renames the setup. One this device doesn't own is copied first.
local function setupName(tbl)
	local sec, context, root = motif.option_info, tbl.inputContext, tbl.inputMenu
	context.popup = main.f_drawInput(sec.textinput.TextSpriteData, sec.textinput.text.setupname,
		sec, context.background, sec.textinput.overlay.RectData, root.draft.name, true)
	context.onClose = function(name)
		esc(false)
		if name == '' or name == root.draft.name then
			return
		end
		ownSetup(root)
		root.draft.name = name
		updateEditor(root, root)
		applyEditor(root)
	end
end

-- Restore and defaults only change the draft.
options.t_itemname.restoreprevious = function(t, item)
	if getAction(-1, motif.option_info.menu.done.action) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		local root = t.inputMenu
		undoCopy(root, root.copyOf)
		root.draft = main.f_tableCopy(root.saved)
		updateEditor(root, root)
	end
	return true
end

options.t_itemname.restoredefault = function(t, item)
	if getAction(-1, motif.option_info.menu.done.action) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		local root = t.inputMenu
		local values
		if root.kind == 'mapping' then
			values = inputGetMappingDefaults(root.device.id)
		else
			values = inputGetSetupDefaults(root.setupId)
			if values ~= nil and not sameSetup(values, root.draft) then
				ownSetup(root)
				values.name = root.draft.name
			end
		end
		if values ~= nil then
			root.draft = values
			updateEditor(root, root)
		end
	end
	return true
end

options.t_itemname.confirmexit = function(t, item)
	if getAction(-1, motif.option_info.menu.done.action) then
		local root = t.inputMenu
		if applyEditor(root) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			root.finished = true
			return false
		end
	end
	return true
end

-- Device switches the device being mapped. The edits of the device being left are
-- kept, like leaving the editor.
options.t_itemname.device = function(t, item)
	local sec = motif.option_info
	local root = t.inputMenu
	local step = 0
	if getAction(-1, sec.menu.add.action) then
		step = 1
	elseif getAction(-1, sec.menu.subtract.action) then
		step = -1
	end
	local devices = mappingDevices()
	if step ~= 0 and #devices > 1 and applyEditor(root) then
		local index = 1
		for i, device in ipairs(devices) do
			if device.id == root.device.id then
				index = i
			end
		end
		sndPlay(motif.Snd, sec.cursor.move.snd[1], sec.cursor.move.snd[2])
		options.f_editMapping(devices[(index - 1 + step) % #devices + 1], t.inputContext)
	end
	return true
end

-- Use Setup switches the device to another setup. Edits to the old one are kept.
-- Confirm renames it, delete removes it.
options.t_itemname.activesetup = function(t, item)
	local sec = motif.option_info
	local root = t.inputMenu
	local step = 0
	if getAction(-1, sec.menu.add.action) then
		step = 1
	elseif getAction(-1, sec.menu.subtract.action) then
		step = -1
	end
	if step ~= 0 then
		if applyEditor(root) and inputCycleSetup(root.device.id, step) then
			sndPlay(motif.Snd, sec.cursor.move.snd[1], sec.cursor.move.snd[2])
			root.device = deviceById(root.device.id)
			loadSetup(root, root.device.setup)
		end
	elseif getAction(-1, sec.menu.done.action) then
		sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
		setupName(t)
	elseif getAction(-1, sec.menu.item.delete.action) then
		-- Shipped setups stay. The device falls back to another setup and edits that one.
		if inputDeleteSetup(root.setupId) then
			sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
			options.modified = true
			root.device = deviceById(root.device.id)
			loadSetup(root, root.device.setup)
		else
			inputPopup(t.inputContext, sec.textinput.text.setupdelete)
		end
	end
	return true
end

-- Opens Button Mapping or Controller Setup from Options or the pause menu.
-- Both go straight to the editor of the device that opened them.
function options.f_inputCfg(itemname, template, background, pause)
	local context = {background = background or motif.optionbgdef, pause = pause or main.pauseMenuActive, parents = {}}
	context.allDevices = not context.pause
	local device = deviceById(inputGetMenuDevice())
	if itemname == 'buttonmapping' then
		if not options.f_editMapping(device, context) then
			return
		end
	elseif not options.f_editSetup(device, context) then
		return
	end
	runInputConfig(context)
end

-- Side select: edit the Controller Setup of the device that pressed the button.
-- It has no player slot yet, so the editor reads that device directly.
function options.f_setupEditor(device)
	local context = {background = motif.optionbgdef, parents = {}}
	-- The press that opened the editor must not act in it too.
	context.skipInput = true
	bgReset(motif.optionbgdef.BGDef)
	local previous = inputSetMenuFocus(device.id)
	if options.f_editSetup(device, context) then
		runInputConfig(context)
	end
	inputSetMenuFocus(previous)
end

-- Pause menu: the same Controller Setup editor, in the pause background.
function options.f_pauseSetup(background)
	options.f_inputCfg('controllersetup', nil, background, true)
end

return options

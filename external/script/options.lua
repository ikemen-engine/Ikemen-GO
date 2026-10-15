local options = {}
--;===========================================================
--; COMMON
--;===========================================================
options.modified = false

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

-- Video options that are applied without restarting.
local t_videoOptions = {
	'GameWidth', 'GameHeight', 'WindowWidth', 'WindowHeight', 'Fullscreen',
	'Borderless', 'VSync', 'MSAA', 'ExternalShaders', 'WindowScaleMode',
	'FightAspectWidth', 'FightAspectHeight', 'KeepAspect', 'EnableModel',
	'EnableModelShadow',
}

-- Returns a copy of the current video options.
function options.f_videoSnapshot()
	local t = {}
	for _, name in ipairs(t_videoOptions) do
		local value = gameOption('Video.' .. name)
		if type(value) == 'table' then
			value = main.f_tableCopy(value)
		end
		t['Video.' .. name] = value
	end
	return t
end

local function f_videoWarning(text, err)
	main.f_warning(main.f_formatBySpec(text, {s = tostring(err)}), motif.option_info, motif.optionbgdef)
end

-- Puts the given video options back and applies them.
local function f_restoreVideo(values)
	for name, value in pairs(values) do
		modifyGameOption(name, value)
	end
	local ok, err = applyVideoSettings()
	if not ok then
		f_videoWarning(motif.warning_info.text.text.videorestore, err)
	end
	return ok
end

-- Applies the edited video options, restoring the previous ones if that fails
-- or, with confirm, if the player does not keep the change.
function options.f_finishVideo(previous, confirm)
	local ok, err = applyVideoSettings()
	if not ok then
		f_restoreVideo(previous)
		f_videoWarning(motif.warning_info.text.text.videoapply, err)
		return false
	end
	if confirm and not main.f_warning(motif.warning_info.text.text.videokeep,
		motif.option_info, motif.optionbgdef, nil, nil, nil, nil,
		motif.option_info.cursor.done.snd, nil, nil, 'confirm') then
		f_restoreVideo(previous)
		return false
	end
	options.modified = true
	return true
end

-- Sets and applies video options.
function options.f_applyVideo(changes, confirm)
	local previous = options.f_videoSnapshot()
	for name, value in pairs(changes) do
		modifyGameOption(name, value)
	end
	return options.f_finishVideo(previous, confirm)
end

-- Choosing a resolution changes both the internal game size and the windowed SDL size.
function options.f_applyResolution(width, height)
	return options.f_applyVideo({
		['Video.GameWidth'] = width, ['Video.GameHeight'] = height,
		['Video.WindowWidth'] = 0, ['Video.WindowHeight'] = 0,
	}, true)
end

--- Save the current configuration to the config file and handle common file modifications
local t_commonFilesOriginal = gameOption('Common')
function options.f_saveCfg(reload)
	if not reload then
		local ok, err = applyVideoSettings()
		if not ok then
			f_videoWarning(motif.warning_info.text.text.videoapply, err)
			return false
		end
	end
	-- Restore the original content of the common files
	local t_commonFiles = gameOption('Common')
	for _, k in ipairs({'Air', 'Cmd', 'Const', 'States', 'Fx', 'Modules', 'Lua'}) do
		modifyGameOption('Common.' .. k, t_commonFilesOriginal[k][k:lower()] or {})
	end
	-- Save the current configuration to 'config.ini'
	saveGameOption(getCommandLineValue("-config"))
	hook.run("options.save", reload)
	-- Options that need a restart, such as the renderer, restart the game.
	-- Only called with reload on desktop, see f_restartOption.
	if reload then
		local ok, err = restartGame()
		if not ok then
			f_videoWarning(motif.warning_info.text.text.restart, err)
			return false
		end
	end
	-- Reapply modified common file arrays after saving
	for _, k in ipairs({'Air', 'Cmd', 'Const', 'States', 'Fx', 'Modules', 'Lua'}) do
		modifyGameOption('Common.' .. k, t_commonFiles[k][k:lower()] or {})
	end
	options.modified = false
	options.savedVideo = options.f_videoSnapshot()
	return true
end

-- Only settings exposed in Options participate in screen-wide restore/default.
-- Profiles and SDL controller mappings have their own reset paths.
options.t_editableOptions = {
	'Options.Difficulty',
	'Options.Life',
	'Options.Time',
	'Options.GameSpeed',
	'Options.Match.Wins',
	'Options.Match.MaxDrawGames',
	'Options.Credits',
	'Options.QuickContinue',
	'Options.AutoGuard',
	'Options.GuardBreak',
	'Options.Dizzy',
	'Options.RedLife',
	'Options.Team.Duplicates',
	'Options.Team.LifeShare',
	'Options.Team.PowerShare',
	'Options.Team.SingleVsTeamLife',
	'Options.Simul.Min',
	'Options.Simul.Max',
	'Options.Simul.Match.Wins',
	'Options.Simul.LoseOnKO',
	'Options.Tag.Min',
	'Options.Tag.Max',
	'Options.Tag.Match.Wins',
	'Options.Tag.LoseOnKO',
	'Options.Tag.TimeScaling',
	'Options.Turns.Min',
	'Options.Turns.Max',
	'Options.Turns.Recovery.Base',
	'Options.Turns.Recovery.Bonus',
	'Config.Players',
	'Config.VsScreenLoading',
	'Config.TurnsLoading',
	'Config.Language',
	'Debug.AllowDebugMode',
	'Debug.AllowDebugKeys',
	'Arcade.AI.RandomColor',
	'Arcade.AI.SurvivalColor',
	'Arcade.AI.Ramping',
	'Netplay.RollbackNetcode',
	'Netplay.ListenPort',
	'Netplay.Rollback.Port',
	'Netplay.Rollback.FrameDelay',
	'Video.RenderMode',
	'Video.GameWidth',
	'Video.GameHeight',
	'Video.WindowWidth',
	'Video.WindowHeight',
	'Video.Fullscreen',
	'Video.Borderless',
	'Video.VSync',
	'Video.MSAA',
	'Video.ExternalShaders',
	'Video.WindowScaleMode',
	'Video.FightAspectWidth',
	'Video.FightAspectHeight',
	'Video.KeepAspect',
	'Video.EnableModel',
	'Video.EnableModelShadow',
	'Sound.MasterVolume',
	'Sound.BGMVolume',
	'Sound.WavVolume',
	'Sound.AudioDucking',
	'Sound.StereoEffects',
	'Sound.PanningRange',
}

-- Copies the current or default values of every setting shown in Options.
function options.f_settingsSnapshot(defaults)
	local values = {}
	for _, key in ipairs(options.t_editableOptions) do
		local value = defaults and gameOptionDefault(key) or gameOption(key)
		if value ~= nil then
			values[key] = type(value) == 'table' and main.f_tableCopy(value) or value
		end
	end
	if defaults then
		-- The shipped shader list is empty. Use an array, not an INI string.
		values['Video.ExternalShaders'] = {}
	end
	return values
end

function options.f_restoreSettings(values)
	local previous = options.f_settingsSnapshot(false)
	for key, value in pairs(values) do
		modifyGameOption(key, value)
	end
	local ok, err = applyVideoSettings()
	if not ok then
		for key, value in pairs(previous) do modifyGameOption(key, value) end
		applyVideoSettings()
		f_videoWarning(motif.warning_info.text.text.videoapply, err)
		return false
	end
	setPlayers()
	updateVolume()
	options.f_refreshVardisplay()
	options.modified = true
	return true
end

function options.f_refreshVardisplay()
	for _, row in ipairs(options.t_vardisplayPointers) do
		row.vardisplay = options.f_vardisplay(row.itemname)
	end
end

-- Language and renderer only change after a restart. A restore can change them
-- silently, so the restart is offered when saving or leaving Options.
local t_restartOptions = {'Config.Language', 'Video.RenderMode'}
local t_restartBaseline = {}
for _, key in ipairs(t_restartOptions) do
	t_restartBaseline[key] = gameOption(key)
end

-- True if language or renderer changed since the game started.
function options.f_restartPending()
	for _, key in ipairs(t_restartOptions) do
		if gameOption(key) ~= t_restartBaseline[key] then
			return true
		end
	end
	return false
end

-- Confirm saves everything and restarts (on mobile it applies at next launch).
-- Cancel restores the running values. Returns true if it saved.
function options.f_finishRestartOptions()
	if not options.f_restartPending() then
		return false
	end
	local mobile = getRuntimeOS() == 'android' or getRuntimeOS() == 'ios'
	local text = motif.warning_info.text.text.restartoption
	if mobile then
		text = motif.warning_info.text.text.restartlater
	end
	local accepted = main.f_warning(text, motif.option_info, motif.optionbgdef, nil, nil, nil, nil, nil, nil, nil, 'confirm')
	if accepted and options.f_saveCfg(not mobile) then
		-- Mobile keeps running the old values, but they are already saved.
		for _, key in ipairs(t_restartOptions) do
			t_restartBaseline[key] = gameOption(key)
		end
		return true
	end
	for _, key in ipairs(t_restartOptions) do
		modifyGameOption(key, t_restartBaseline[key])
	end
	options.f_refreshVardisplay()
	if accepted then
		-- The restart failed after saving: keep the running values saved instead.
		options.f_saveCfg(false)
		return true
	end
	return false
end

-- Back/ESC keeps unsaved changes in the running session.
function options.f_close()
	options.f_finishRestartOptions()
	fadeOutInit(motif.option_info.fadeout.FadeData)
	main.close = true
end

--;===========================================================
--; LOOPS
--;===========================================================
-- Associative elements table storing functions controlling behaviour of each
-- option screen item. Can be appended via external module.
--Render mode values as stored in the configuration.
options.t_renderModes = {
	gles32 = 'OpenGL ES 3.2',
	gl33 = 'OpenGL 3.3',
	vk13 = 'Vulkan 1.3',
}

-- Sets an option that only takes effect after a restart. Once the player
-- confirms, all options are saved and the game restarts (desktop only).
function options.f_restartOption(name, value)
	-- Android and iOS cannot restart the app, so restart-only options take effect
	-- the next time the game starts there.
	local mobile = getRuntimeOS() == 'android' or getRuntimeOS() == 'ios'
	local old = gameOption(name)
	if old == value then
		return
	end
	-- Going back to the running value after a restore needs no restart.
	if t_restartBaseline[name] == value then
		modifyGameOption(name, value)
		options.modified = true
		options.f_refreshVardisplay()
		return
	end
	local text = motif.warning_info.text.text.restartoption
	if mobile then
		text = motif.warning_info.text.text.restartlater
	end
	if not main.f_warning(text, motif.option_info, motif.optionbgdef, nil, nil, nil, nil, nil, nil, nil, 'confirm') then
		return
	end
	modifyGameOption(name, value)
	options.modified = true
	if not options.f_saveCfg(not mobile) then
		-- Keep the old value saved if the game could not restart.
		modifyGameOption(name, old)
		options.f_saveCfg(false)
	elseif mobile then
		-- Every restart-only value is now saved for the next start.
		for _, key in ipairs(t_restartOptions) do
			t_restartBaseline[key] = gameOption(key)
		end
	end
end

options.t_itemname = {
	--Back
	['back'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
			if t == options.menu then
				options.f_close()
				return true
			end
			return false
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
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			for _, v in ipairs(t.submenu[t.items[item].itemname].items) do
				v.selected = v.language ~= nil and v.language == gameOption('Config.Language')
			end
			t.submenu[t.items[item].itemname].loop()
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
				v.selected = gameOption('Video.RenderMode') == options.t_renderModes[v.itemname]
			end
			t.submenu[t.items[item].itemname].loop()
			t.items[item].vardisplay = options.t_vardisplay['renderer']()
		end
		return true
	end,
	--gles32
	['gles32'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			options.f_restartOption('Video.RenderMode', options.t_renderModes.gles32)
			return false
		end
		return true
	end,
	--gl33
	['gl33'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			options.f_restartOption('Video.RenderMode', options.t_renderModes.gl33)
			return false
		end
		return true
	end,
	--vk13
	['vk13'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			options.f_restartOption('Video.RenderMode', options.t_renderModes.vk13)
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
			local reswidth = tonumber(main.f_drawInput(motif.option_info.textinput.text.reswidth))
			local validWidth = reswidth ~= nil and reswidth > 0 and reswidth == math.floor(reswidth)
			main.f_inputSound(validWidth)
			if validWidth then
				local resheight = tonumber(main.f_drawInput(motif.option_info.textinput.text.resheight))
				local validHeight = resheight ~= nil and resheight > 0 and resheight == math.floor(resheight)
				main.f_inputSound(validHeight)
				if validHeight then
					options.f_applyResolution(reswidth, resheight)
				end
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
		local old = gameOption('Video.MSAA')
		local nextValue
		if getInput(-1, motif.option_info.menu.add.key) and old < 32 then
			nextValue = old == 0 and 2 or old * 2
		elseif getInput(-1, motif.option_info.menu.subtract.key) and old > 1 then
			nextValue = old == 2 and 0 or old / 2
		end
		if nextValue and options.f_applyVideo({['Video.MSAA'] = nextValue}) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			t.items[item].vardisplay = options.t_vardisplay['msaa']()
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
			local aspectwidth = tonumber(main.f_drawInput(motif.option_info.textinput.text.aspectwidth))
			local validWidth = aspectwidth ~= nil and aspectwidth > 0
			main.f_inputSound(validWidth)
			if validWidth then
				local aspectheight = tonumber(main.f_drawInput(motif.option_info.textinput.text.aspectheight))
				local validHeight = aspectheight ~= nil and aspectheight > 0
				main.f_inputSound(validHeight)
				if validHeight then
					options.f_applyVideo({['Video.FightAspectWidth'] = aspectwidth, ['Video.FightAspectHeight'] = aspectheight})
				end
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
		end
		return true
	end,
	--Disable (shader)
	['noshader'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
			options.f_applyVideo({['Video.ExternalShaders'] = {}})
			return false
		end
		return true
	end,
	--Enable Model
	['enablemodel'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if options.f_applyVideo({['Video.EnableModel'] = not gameOption('Video.EnableModel')}) then
				t.items[item].vardisplay = options.t_vardisplay['enablemodel']()
			end
		end
		return true
	end,
	--Enable Model Shadow
	['enablemodelshadow'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if options.f_applyVideo({['Video.EnableModelShadow'] = not gameOption('Video.EnableModelShadow')}) then
				t.items[item].vardisplay = options.t_vardisplay['enablemodelshadow']()
			end
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
			inputResetProfiles()
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
			local port = tonumber(main.f_drawInput(motif.option_info.textinput.text.port))
			local valid = port ~= nil and port == math.floor(port) and port >= 1 and port <= 65535
			main.f_inputSound(valid)
			if valid then
				modifyGameOption('Netplay.ListenPort', port)
				t.items[item].vardisplay = options.t_vardisplay['portchange']()
				options.modified = true
			end
		end
		return true
	end,
	--Rollback Port (UDP)
	['rollbackport'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			local port = tonumber(main.f_drawInput(motif.option_info.textinput.text.port))
			local valid = port ~= nil and port == math.floor(port) and port >= 1 and port <= 65535
			main.f_inputSound(valid)
			if valid then
				modifyGameOption('Netplay.Rollback.Port', port)
				t.items[item].vardisplay = options.t_vardisplay['rollbackport']()
				options.modified = true
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
	--Save Settings
	['savesettings'] = function(t, item, cursorPosY, moveTxt)
		if getInput(-1, motif.option_info.menu.done.key) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			-- A pending language/renderer change saves everything itself when confirmed.
			if not options.f_finishRestartOptions() and options.modified then
				options.f_saveCfg(false)
			end
		end
		return true
	end,
}

-- Rows that show a value but open a list or text prompt instead of changing it.
options.t_hintDefault = {
	language = true,
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
			options.entryValues = options.f_settingsSnapshot(false)
			options.savedVideo = options.f_videoSnapshot()
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
						options.f_close()
					else
						break
					end
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
		return motif.languages[gameOption('Config.Language')] or gameOption('Config.Language')
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
			-- Vulkan lists compiled name.vert.spv / name.frag.spv files.
			if gameOption('Video.RenderMode'):match('^Vulkan') ~= nil then
				local kind = filename:match('%.([^.]+)$')
				if ext ~= 'spv' or (kind ~= 'frag' and kind ~= 'vert') then
					return
				end
				filename = filename:sub(1, -(#kind + 2))
				ext = kind
			end
			if ext == 'frag' then
				table.insert(options.t_shaders, {path = path, filename = filename})
			end
			if ext:match('vert') or ext:match('frag') --[[or ext:match('shader')]] then
				options.t_itemname[path .. filename] = function(t, item, cursorPosY, moveTxt)
					if getInput(-1, motif.option_info.menu.done.key) then
						sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
						local t_externalShaders = main.f_tableCopy(gameOption('Video.ExternalShaders'))
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
						-- On failure, show the shaders that are still active.
						if not options.f_applyVideo({['Video.ExternalShaders'] = t_externalShaders}) then
							for _, shader in ipairs(t.items) do
								shader.selected, shader.vardisplay = false, ''
								for index, name in ipairs(gameOption('Video.ExternalShaders')) do
									if shader.itemname == name then
										shader.selected, shader.vardisplay = true, tostring(index)
									end
								end
							end
						end
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
					options.f_applyResolution(tonumber(width), tonumber(height))
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
						options.f_applyVideo({['Video.FightAspectWidth'] = 0, ['Video.FightAspectHeight'] = 0})
						return false
					end
					return true
				end
			-- aspect ratio stage
			elseif v:match('_aspectratio_stageaspect$') then
				options.t_itemname['stageaspect'] = function(t, item, cursorPosY, moveTxt)
					if getInput(-1, motif.option_info.menu.done.key) then
						sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
						options.f_applyVideo({['Video.FightAspectWidth'] = -1, ['Video.FightAspectHeight'] = -1})
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
			-- populate language submenu
			if suffix:match('_language_back$') and c == 'back' then
				local codes = {}
				for code in pairs(motif.languages) do
					table.insert(codes, code)
				end
				-- Rows are inserted at the top, so sort by name in reverse.
				table.sort(codes, function(a, b)
					return motif.languages[a] > motif.languages[b]
				end)
				for _, code in ipairs(codes) do
					local itemname = 'language_' .. code
					options.t_itemname[itemname] = function(t, item, cursorPosY, moveTxt)
						if getInput(-1, motif.option_info.menu.done.key) then
							sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
							options.f_restartOption('Config.Language', code)
							return false
						end
						return true
					end
					table.insert(t_pos.items, 1, {
						itemname = itemname,
						displayname = motif.languages[code],
						paramname = suffix:gsub('back$', itemname),
						vardisplay = '',
						selected = false,
						language = code,
					})
				end
			end
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
		excluded = {renderer = true, keyboard = true, gl33 = true, gles32 = true}
	end
	-- Vulkan cannot change MSAA without a restart yet.
	if gameOption('Video.RenderMode'):match('^Vulkan') then
		excluded.msaa = true
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

local function inputPopup(context, text, getValue, onClose, hintContext)
	local sec = motif.option_info
	context.popup = main.f_warning(text, sec, context.background, nil, sec.title.TextSpriteData, nil,
		sec.cancel.snd, sec.cursor.done.snd, getValue, true, hintContext)
	context.onClose = onClose
end

-- Expand the profile placeholder in place, retaining the order the motif declares.
local function populateInputMenu(tbl, kind, selected)
	for i = #tbl.items, 1, -1 do
		local row = tbl.items[i]
		if row.itemname == 'profile' then
			table.remove(tbl.items, i)
			local entries = inputGetSources()
			if selected ~= nil then
				entries = inputGetProfiles(selected.id)
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
				row.vardisplay = inputGetProfile(selected.id).name
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
		if values[field] == '-' then
			row.vardisplay = motif.option_info.menu.valuename.nokey
		end
		row.glyphSource = source.id
		row.hintContext = 'profile'
		if t_profileButtons[field] then
			row.hintContext = 'binding'
		elseif field == 'rumble' then
			row.hintContext = 'adjust'
		end
		if t_profileButtons[field] and motif.option_info.useglyphs and values[field] ~= '-' then
			row.vardisplay = '<binding:' .. values[field] .. '>'
		end
		if field == 'rumble' then
			row.vardisplay = options.f_boolDisplay(values.rumble)
		elseif field == 'glyphs' then
			row.hintContext = 'adjust'
			for _, set in ipairs(inputGetGlyphPacks(source.id)) do
				if set.selected then
					row.vardisplay = set.glyph .. ' ' .. set.name
					break
				end
			end
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
	local values = inputGetProfile(tbl.inputSource.id)
	local root = tbl.inputMenu
	root.title = main.f_itemnameUpper(values.name, sec.menu.title.uppercase)
	updateProfileMenu(root, root, tbl.inputSource, values)
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
	while not context.closed and context.popup ~= nil do
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
	if context.device ~= nil then
		if gamepadMappingGetDevice(context.device.instance) == nil then
			context.closed = true
		end
	elseif source ~= nil and source.kind == 'gamepad' and not getJoystickPresent(source.joy) then
		context.closed = true
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
			-- Leaving the profile editor keeps its edits, like Confirm and Exit.
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
		textInputCancel()
		gamepadMappingCancel()
		if context.previousSource ~= nil then
			inputSetConfigSource(context.previousSource)
		end
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
	local values = inputGetProfile(source.id)
	tbl.title = main.f_itemnameUpper(values.name, motif.option_info.menu.title.uppercase)
	tbl.savedBindings = main.f_tableCopy(values)
	for _, set in ipairs(inputGetGlyphPacks(source.id)) do
		if set.selected then tbl.savedGlyph = set.id; break end
	end
	tbl.inputContext = context
	updateProfileMenu(tbl, tbl, source, values)
	context.current, context.parents = tbl, {}
	-- Capture targets the edited source; menu navigation keeps the source that opened it.
	local menuSource = source.id
	if not ownContext then
		menuSource = inputGetMenuSource()
		if context.source ~= nil then
			menuSource = context.source.id
		end
	end
	local previousSource = inputSetConfigSource(menuSource)
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
		values = inputGetProfile(tbl.inputSource.id), menu = tbl, item = item,
	}
	resetKey()
	resetTokenGuard()
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
		updateProfileMenu(root, root, source, inputGetProfile(source.id))
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

-- Leave the mapping screen.
local function gamepadMappingClose(context, finish)
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

-- Calibration uses [Mapping Info] and [MappingBgDef], or the pause background in a match.
-- Keyboard controls keep working even if the pad has no mapping.
function options.f_gamepadMapping(context, device, testOnly)
	if not device or not gamepadMappingBegin(device.instance) then
		context.closed = true
		return
	end
	local previous = inputSetConfigSource(0)
	if context.previousSource == nil then
		context.previousSource = previous
	end
	context.device = device
	context.mapping = {device = device, name = device.name,
		index = testOnly and 0 or 1, testOnly = testOnly or false}
	if not context.pause then
		bgReset(motif.mappingbgdef.BGDef)
		fadeInInit(motif.mapping_info.fadein.FadeData)
	end
	if not device.recognized and not testOnly then
		local sec = motif.option_info
		context.popup = main.f_drawInput(sec.textinput.text.gamepadname, device.name, true)
		context.onClose = function(name)
			main.f_inputSound(name ~= '')
			if name == '' then
				gamepadMappingCancel()
				context.closed = true
			else
				context.mapping.name = name
				-- The pad sat idle while it was named: record its resting axes again.
				gamepadMappingCalibrate()
			end
		end
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
	local name = cfg.order[mapping.index]
	local control = cfg.control[name]
	local state = 'default'
	if mapping.testOnly then
		state = 'test'
	elseif mapping.index > #cfg.order then
		state = 'save'
	end
	local pressed = {}
	if mapping.index == 0 then
		pressed = gamepadMappingGetState()
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
	for _, key in ipairs(cfg.order) do
		local art = cfg.control[key]
		if art ~= nil and (key == name or (pressed[key] and art.rear == rear)) then
			main.f_animPosDraw(art.AnimData)
		end
	end
	local label = name or ''
	if control ~= nil and control.text ~= nil then
		label = control.text
	end
	textImgReset(cfg.prompt.TextSpriteData)
	textImgSetText(cfg.prompt.TextSpriteData, main.f_formatBySpec(cfg.prompt.text[state], {
		s = label, i = mapping.index,
	}))
	textImgDraw(cfg.prompt.TextSpriteData)
	bgDraw(background.BGDef, 1)
	main.f_hintDraw(cfg.hint, state, nil, mapping.device.id)
	if mapping.finish ~= nil then
		if not fadeActive() then
			mapping.finish()
		end
		return
	end
	if context.skipInput or fadeActive() then
		return
	end
	-- Test mode never captures or saves a mapping.
	if esc() then
		gamepadMappingCancel()
		sndPlay(motif.Snd, cfg.cancel.snd[1], cfg.cancel.snd[2])
		esc(false)
		gamepadMappingClose(context, function()
			context.mapping = nil
			context.closed = not mapping.device.recognized and not mapping.testOnly
			if not context.closed then
				inputSetConfigSource(mapping.device.id)
			end
		end)
	elseif mapping.testOnly then
		-- Read-only test mode: inputs only highlight the controller diagram.
	elseif getKey(cfg.previous.keycode) and mapping.index > 1 then
		mapping.index = mapping.index - 1
		sndPlay(motif.Snd, cfg.move.snd[1], cfg.move.snd[2])
	elseif state == 'save' and getKey(cfg.confirm.keycode) then
		local message = gamepadMappingSave(mapping.name)
		if message == '' then
			sndPlay(motif.Snd, cfg.done.snd[1], cfg.done.snd[2])
			gamepadMappingClose(context, function()
				context.mapping = nil
				if context.openDevice ~= nil then
					context.openDevice(gamepadMappingGetDevice(mapping.device.instance))
				end
			end)
		else
			inputPopup(context, main.f_formatBySpec(sec.textinput.text.gamepadmappingfailed, {s = message}))
		end
	elseif name ~= nil then
		local skip = getKey(cfg.skip.keycode)
		if skip or gamepadMappingCapture(name) ~= '' then
			if skip then
				gamepadMappingSkip(name)
			end
			mapping.index = mapping.index + 1
			sndPlay(motif.Snd, cfg.done.snd[1], cfg.done.snd[2])
		end
	end
	resetKey()
	resetTokenGuard()
end

-- Input editor handlers. One handler serves both kinds: it reads the source
-- from the menu it is given.
options.t_itemname.configall = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		startInputCapture(t, item, true)
		resetKey()
	end
	return true
end

for _, field in ipairs(t_profileFields) do
	options.t_itemname[field] = function(t, item)
		local values = inputGetProfile(t.inputSource.id)
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

options.t_itemname.glyphs = function(t, item)
	local step = 0
	if getInput(-1, motif.option_info.menu.add.key) then
		step = 1
	elseif getInput(-1, motif.option_info.menu.subtract.key) then
		step = -1
	end
	if step ~= 0 then
		sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
		local sets = inputGetGlyphPacks(t.inputSource.id)
		local current = 1
		for i, set in ipairs(sets) do
			if set.selected then
				current = i
				break
			end
		end
		local chosen = sets[(current - 1 + step) % #sets + 1]
		-- The pack belongs to the device rather than the bindings, so it is saved
		-- straight away and Cancel does not undo it.
		if inputSetGlyphPack(t.inputSource.id, chosen.id) then
			local root = t.inputMenu
			updateProfileMenu(root, root, t.inputSource, inputGetProfile(t.inputSource.id))
		end
	end
	return true
end

options.t_itemname.rumble = function(t, item)
	if getInput(-1, motif.option_info.menu.add.key, motif.option_info.menu.subtract.key) then
		sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
		local values = inputGetProfile(t.inputSource.id)
		values.rumble = not values.rumble
		applyProfile(t, values)
	end
	return true
end

local function profileName(tbl, copy)
	local sec, context = motif.option_info, tbl.inputContext
	local values = inputGetProfile(tbl.inputSource.id)
	context.popup = main.f_drawInput(sec.textinput.text.profilename, values.name, true)
	context.onClose = function(name)
		local saved = false
		if name ~= '' then
			values.name = name
			if copy then
				-- The copy takes the edits; the original profile keeps its saved state.
				local root = tbl.inputMenu
				if applyProfile(tbl, root.savedBindings) and applyProfile(tbl, values, true) then
					root.finished = true
					saved = true
				end
			else
				saved = applyProfile(tbl, values)
			end
		end
		main.f_inputSound(saved)
		esc(false)
	end
end

options.t_itemname.renameprofile = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
		profileName(t, false)
	end
	return true
end

options.t_itemname.savecopy = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
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

-- Restore, defaults and saving are separate operations. Back retains runtime edits.
-- In the profile editor they act on the edited profile, elsewhere on Options settings.
options.t_itemname.restoreprevious = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		local root = t.inputMenu
		if root ~= nil then
			applyProfile(root, root.savedBindings)
			if root.savedGlyph ~= nil then inputSetGlyphPack(root.inputSource.id, root.savedGlyph) end
			updateProfileMenu(root, root, root.inputSource, inputGetProfile(root.inputSource.id))
		else
			options.f_restoreSettings(options.entryValues)
		end
	end
	return true
end
options.t_itemname.restoredefault = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		local root = t.inputMenu
		if root ~= nil then
			local values = inputGetProfileDefaults(root.inputSource.id)
			if values then applyProfile(root, values) end
		else
			options.f_restoreSettings(options.f_settingsSnapshot(true))
			hook.run("options.default")
		end
	end
	return true
end
options.t_itemname.confirmexit = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		t.inputMenu.finished = true
		return false
	end
	return true
end

options.t_itemname.calibrate = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		local context, sec = t.inputContext, motif.option_info
		inputPopup(context, sec.textinput.text.gamepadcalibrate, nil, function(accepted)
			if accepted then options.f_gamepadMapping(context, context.device) end
		end, 'confirm')
	end
	return true
end

options.t_itemname.checkinput = function(t, item)
	if getInput(-1, motif.option_info.menu.done.key) then
		local context, source = t.inputContext, t.inputSource
		for _, device in ipairs(gamepadMappingGetDevices()) do
			if device.joy == source.joy and device.recognized then
				options.f_gamepadMapping(context, device, true)
				break
			end
		end
	end
	return true
end

options.t_itemname.newprofile = function(t, item)
	local sec, context = motif.option_info, t.inputContext
	if getInput(-1, sec.menu.done.key) then
		local source = context.source
		local menuSource = inputGetMenuSource()
		if source ~= nil then
			menuSource = source.id
		else
			-- Keyboard layouts have separate sources. Keep the existing one intact,
			-- preferring the one navigating the menu.
			for _, entry in ipairs(inputGetSources()) do
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
		local previous = inputSetConfigSource(menuSource)
		if context.previousSource == nil then
			context.previousSource = previous
		end
		local name = sec.menu.valuename.keyboardprofile
		if source.kind == 'gamepad' then
			name = sec.menu.valuename.gamepadprofile
		end
		-- Name the profile before opening its editor.
		context.popup = main.f_drawInput(sec.textinput.text.profilename,
			inputGetProfileName(name), true)
		context.onClose = function(value)
			esc(false)
			if value == '' then
				main.f_inputSound(false)
				return
			end
			local id = inputNewProfile(source.id, value, sec.menu.valuename.copy)
			if id == 0 then
				main.f_inputSound(false)
				inputPopup(context, sec.textinput.text.profilesave)
				return
			end
			main.f_inputSound(true)
			options.modified = true
			for _, entry in ipairs(inputGetSources()) do
				if entry.id == id then
					options.f_editProfile(entry, nil, context, true)
					break
				end
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
		for _, source in ipairs(inputGetSources()) do
			if source.id == inputGetMenuSource() then
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
		-- Select the navigating source again, so bindings saved or chosen in the
		-- profile editor take effect in this menu.
		local previous = inputSetConfigSource(0)
		if context.previousSource == nil then
			context.previousSource = previous
		end
		if selected ~= nil then
			inputSetConfigSource(selected.id)
		else
			inputSetConfigSource(context.previousSource)
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
	context.openDevice = function(device)
		context.device = device
		openMenu(device)
	end
	if kind == 'gamepad' then
		textImgReset(sec.title.TextSpriteData)
		textImgSetText(sec.title.TextSpriteData, template.title)
		local released = false
		inputPopup(context, sec.textinput.text.gamepad, function()
			local instance = getJoystickRawId()
			if instance == nil then
				released = true
			elseif released then
				local device = gamepadMappingGetDevice(instance)
				-- The controller's own cancel button backs out, instead of being
				-- read as the button that picks a device to configure.
				if device ~= nil and device.recognized then
					local profile = inputGetProfile(device.id)
					if profile ~= nil and getJoystickKey(device.joy) == profile.menu then
						return false
					end
				end
				return device
			end
		end, function(device)
			if not device then
				context.closed = true
			elseif device.recognized then
				context.openDevice(device)
			else
				-- An unmapped pad cannot navigate a profile picker yet.
				context.device = device
				inputPopup(context, sec.textinput.text.gamepadunrecognized, nil, function(accepted)
					if accepted then
						options.f_gamepadMapping(context, device)
					else
						context.closed = true
					end
				end, 'confirm')
			end
		end)
	else
		openMenu()
	end
	runInputConfig(context)
end

-- Prompt once per physical connection. A canceled prompt is not retried until
-- the pad is disconnected and reconnected. This is called by normal Lua menus.
local t_unknownPrompted = {}
function options.f_unrecognizedControllerPrompt()
	if options.promptingUnknown or options.editingInput then return end
	local connected = {}
	local first = nil
	for _, device in ipairs(gamepadMappingGetDevices()) do
		connected[device.instance] = true
		if not device.recognized and not t_unknownPrompted[device.instance] and first == nil then
			first = device
		end
	end
	for instance in pairs(t_unknownPrompted) do
		if not connected[instance] then t_unknownPrompted[instance] = nil end
	end
	if first == nil then return end
	t_unknownPrompted[first.instance] = true
	options.promptingUnknown = true
	local sec = motif.option_info
	local accepted = main.f_warning(sec.textinput.text.gamepadunrecognized,
		sec, motif.optionbgdef, nil, nil, nil, nil, nil, nil, nil, 'confirm')
	if accepted and gamepadMappingGetDevice(first.instance) ~= nil then
		local context = {background = motif.optionbgdef, parents = {}, device = first}
		options.f_gamepadMapping(context, first)
		runInputConfig(context)
	end
	options.promptingUnknown = false
	resetKey()
	resetTokenGuard()
end

return options

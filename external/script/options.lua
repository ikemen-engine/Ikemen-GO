local options = {}
--;===========================================================
--; COMMON
--;===========================================================
options.modified = false

-- return string depending on bool
function options.f_boolDisplay(bool, t, f)
	if bool == true then
		return t or motif.option_info.menu.valuename.on
	end
	return f or motif.option_info.menu.valuename.off
end

-- return table entry (or ret if specified) if provided key exists in the table, otherwise return default argument
function options.f_definedDisplay(key, t, default, ret)
	if key ~= nil and t[key] ~= nil then
		return ret or t[key]
	end
	return default
end

-- return correct precision
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
	local ok, err = applyVideoSettings(true)
	if not ok then
		f_videoWarning(motif.warning_info.text.text.videorestore, err)
	end
	return ok
end

-- Applies the edited video options, restoring the previous ones if that fails
-- or, with confirm, if the player does not keep the change.
function options.f_finishVideo(previous, confirm)
	local ok, err = applyVideoSettings(true)
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
		local ok, err = applyVideoSettings(true)
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
-- Controller Mappings and Button Setups have their own reset.
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
	'Options.PauseDelay',
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
	local ok, err = applyVideoSettings(true)
	if not ok then
		for key, value in pairs(previous) do modifyGameOption(key, value) end
		applyVideoSettings(true)
		f_videoWarning(motif.warning_info.text.text.videoapply, err)
		return false
	end
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

-- A restore can change restart-only settings silently. Apply the other settings
-- now, and offer the restart when saving or leaving Options.
local t_restartOptions = {'Config.Language', 'Video.RenderMode'}
if gameOption('Video.RenderMode'):match('^Vulkan') then
	table.insert(t_restartOptions, 'Video.MSAA')
end
local t_restartBaseline = {}
for _, key in ipairs(t_restartOptions) do
	t_restartBaseline[key] = gameOption(key)
end

-- True if a restart-only setting changed since startup or the last deferred save.
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
-- Render mode values as stored in the configuration.
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
	-- Back
	['back'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
			if t == options.menu then
				options.f_close()
				return true
			end
			return false
		end
		return true
	end,
	-- Difficulty Level
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
	-- Time Limit
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
	-- Language Setting
	['language'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			for _, v in ipairs(t.submenu[t.items[item].itemname].items) do
				v.selected = v.language ~= nil and v.language == gameOption('Config.Language')
			end
			t.submenu[t.items[item].itemname].loop()
			t.items[item].vardisplay = options.t_vardisplay['language']()
		end
		return true
	end,
	-- Life
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
	-- Single VS Team Life
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
	-- Rounds to Win (Single)
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
	-- Max Draw Games
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
	-- Credits
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
	-- Arcade Palette
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
	-- Survival Palette
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
	-- AI Ramping
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
	-- Quick Continue
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
	-- Auto-Guard
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
	-- Dizzy
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
	-- Guard Break
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
	-- Red Life
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
	-- Team Duplicates
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
	-- Team Life Share
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
	-- Team Power Share
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
	-- Pause Delay
	['pausedelay'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if gameOption('Options.PauseDelay') then
				modifyGameOption('Options.PauseDelay', false)
			else
				modifyGameOption('Options.PauseDelay', true)
			end
			t.items[item].vardisplay = options.t_vardisplay['pausedelay']()
			options.modified = true
		end
		return true
	end,
	-- Rounds to Win (Tag)
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
	-- Partner KOed Lose
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
	-- Min Tag Chars
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
	-- Max Tag Chars
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
	-- Rounds to Win (Simul)
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
	-- Simul Player KOed Lose
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
	-- Min Simul Chars
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
	-- Max Simul Chars
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
	-- Turns Recovery Base
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
	-- Turns Recovery Bonus
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
	-- Min Turns Chars
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
	-- Max Turns Chars
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
	-- Renderer (submenu)
	['renderer'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			for k, v in ipairs(t.submenu[t.items[item].itemname].items) do
				v.selected = gameOption('Video.RenderMode') == options.t_renderModes[v.itemname]
			end
			t.submenu[t.items[item].itemname].loop()
			t.items[item].vardisplay = options.t_vardisplay['renderer']()
		end
		return true
	end,
	-- gles32
	['gles32'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			options.f_restartOption('Video.RenderMode', options.t_renderModes.gles32)
			return false
		end
		return true
	end,
	-- gl33
	['gl33'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			options.f_restartOption('Video.RenderMode', options.t_renderModes.gl33)
			return false
		end
		return true
	end,
	-- vk13
	['vk13'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			options.f_restartOption('Video.RenderMode', options.t_renderModes.vk13)
			return false
		end
		return true
	end,
	-- Resolution (submenu)
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
	-- Custom resolution
	['customres'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
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
	-- Fullscreen
	['fullscreen'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action)
			and options.f_applyVideo({['Video.Fullscreen'] = not gameOption('Video.Fullscreen')}) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			t.items[item].vardisplay = options.t_vardisplay['fullscreen']()
		end
		return true
	end,
	-- VSync
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
	-- MSAA
	['msaa'] = function(t, item, cursorPosY, moveTxt)
		local old = gameOption('Video.MSAA')
		local nextValue
		if getAction(-1, motif.option_info.menu.add.action) and old < 32 then
			nextValue = old == 0 and 2 or old * 2
		elseif getAction(-1, motif.option_info.menu.subtract.action) and old > 1 then
			nextValue = old == 2 and 0 or old / 2
		end
		if nextValue and options.f_applyVideo({['Video.MSAA'] = nextValue}) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			t.items[item].vardisplay = options.t_vardisplay['msaa']()
		end
		return true
	end,
	-- Window scaling mode
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
	-- Custom aspect ratio
	['customaspect'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
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
	-- Keep Aspect Ratio
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
	-- Shaders (submenu)
	['shaders'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
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
	-- Disable (shader)
	['noshader'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cancel.snd[1], motif.option_info.cancel.snd[2])
			options.f_applyVideo({['Video.ExternalShaders'] = {}})
			return false
		end
		return true
	end,
	-- Enable Model
	['enablemodel'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if options.f_applyVideo({['Video.EnableModel'] = not gameOption('Video.EnableModel')}) then
				t.items[item].vardisplay = options.t_vardisplay['enablemodel']()
			end
		end
		return true
	end,
	-- Enable Model Shadow
	['enablemodelshadow'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
			if options.f_applyVideo({['Video.EnableModelShadow'] = not gameOption('Video.EnableModelShadow')}) then
				t.items[item].vardisplay = options.t_vardisplay['enablemodelshadow']()
			end
		end
		return true
	end,
	-- Master Volume
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
	-- BGM Volume
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
	-- SFX Volume
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
	-- Audio Ducking
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
	-- Stereo Effects
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
	-- Panning Range
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
	-- Input Settings edit the device that opened them; Device switches to another one
	['menuinput'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			options.f_inputSettings(t.submenu[t.items[item].itemname])
			main.f_menuSnap(motif.option_info)
			main.f_menuItemBgAnimReset(motif.option_info)
		end
		return true
	end,
	-- Battle Controls
	['buttonsetup'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			options.f_inputCfg('buttonsetup')
			textImgSetText(motif.option_info.title.TextSpriteData, t.title)
		end
		return true
	end,
	-- Menu Controls
	['keyboardmenu'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			options.f_inputCfg('mapping')
			textImgSetText(motif.option_info.title.TextSpriteData, t.title)
		end
		return true
	end,
	-- Controller Mapping
	['controllermapping'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			options.f_inputCfg('mapping')
			textImgSetText(motif.option_info.title.TextSpriteData, t.title)
		end
		return true
	end,
	-- Restore All Input Defaults
	['inputdefault'] = function(t, item, cursorPosY, moveTxt)
		local sec = motif.option_info
		if getAction(-1, sec.menu.done.action) then
			sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
			if main.f_warning(sec.textinput.text.inputdefault, sec, motif.optionbgdef, nil, sec.title.TextSpriteData, nil, sec.cancel.snd, sec.cursor.done.snd) then
				inputRestoreDefaults()
				options.modified = true
				options.f_refreshVardisplay()
				options.f_refreshInputSettings()
			end
		end
		return true
	end,
	-- Players
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
	-- Debug Keys
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
	-- Debug Mode
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
	-- Background Loading
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
	-- Turns Preloading
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
	-- Rollback Netcode
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
	-- Host Port (TCP)
	['portchange'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
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
	-- Rollback Port (UDP)
	['rollbackport'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
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
	-- Rollback Frame Delay
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
	-- Save Settings
	['savesettings'] = function(t, item, cursorPosY, moveTxt)
		if getAction(-1, motif.option_info.menu.done.action) then
			sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
			-- A pending restart-only change saves everything itself when confirmed.
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
				local hintDevice = nil
				if context ~= nil and context.capture ~= nil then
					hintContext = 'capture'
					-- Only the edited device can cancel a capture, so the hint shows its buttons.
					local device = context.capture.root.device
					if device ~= nil and device.id > 0 then
						hintDevice = device.id
					end
				end
				local menuBackground = motif.optionbgdef
				if context ~= nil and context.background ~= nil then
					menuBackground = context.background
				end
				main.f_menuCommonDraw(t, item, cursorPosY, moveTxt, motif.option_info,
					menuBackground, context ~= nil, {hintContext = hintContext, hintDevice = hintDevice})
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
						options.f_close()
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
		return options.f_boolDisplay(gameOption('Sound.AudioDucking'), motif.option_info.menu.valuename.on, motif.option_info.menu.valuename.off)
	end,
	['autoguard'] = function()
		return options.f_boolDisplay(gameOption('Options.AutoGuard'))
	end,
	['vsscreenloading'] = function()
		return options.f_boolDisplay(gameOption('Config.VsScreenLoading'), motif.option_info.menu.valuename.on, motif.option_info.menu.valuename.off)
	end,
	['turnsloading'] = function()
		return options.f_boolDisplay(gameOption('Config.TurnsLoading'), motif.option_info.menu.valuename.on, motif.option_info.menu.valuename.off)
	end,
	['bgmvolume'] = function()
		return gameOption('Sound.BGMVolume') .. '%'
	end,
	['credits'] = function()
		return options.f_definedDisplay(gameOption('Options.Credits'), {[0] = motif.option_info.menu.valuename.off}, gameOption('Options.Credits'))
	end,
	['debugkeys'] = function()
		return options.f_boolDisplay(gameOption('Debug.AllowDebugKeys'), motif.option_info.menu.valuename.on, motif.option_info.menu.valuename.off)
	end,
	['debugmode'] = function()
		return options.f_boolDisplay(gameOption('Debug.AllowDebugMode'), motif.option_info.menu.valuename.on, motif.option_info.menu.valuename.off)
	end,
	['difficulty'] = function()
		return gameOption('Options.Difficulty')
	end,
	['enablemodel'] = function()
		return options.f_definedDisplay(gameOption('Video.EnableModel'), {[true] = motif.option_info.menu.valuename.on}, motif.option_info.menu.valuename.off)
	end,
	['enablemodelshadow'] = function()
		return options.f_definedDisplay(gameOption('Video.EnableModelShadow'), {[true] = motif.option_info.menu.valuename.on}, motif.option_info.menu.valuename.off)
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
		return options.f_boolDisplay(gameOption('Video.KeepAspect'), motif.option_info.menu.valuename.on, motif.option_info.menu.valuename.off)
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
		return options.f_definedDisplay(gameOption('Video.MSAA'), {[0] = motif.option_info.menu.valuename.off}, gameOption('Video.MSAA') .. 'x')
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
		return options.f_boolDisplay(#gameOption('Video.ExternalShaders') > 0, motif.option_info.menu.valuename.on, motif.option_info.menu.valuename.off)
	end,
	['singlevsteamlife'] = function()
		return gameOption('Options.Team.SingleVsTeamLife') .. '%'
	end,
	['stereoeffects'] = function()
		return options.f_boolDisplay(gameOption('Sound.StereoEffects'), motif.option_info.menu.valuename.on, motif.option_info.menu.valuename.off)
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
	['pausedelay'] = function()
		return options.f_boolDisplay(gameOption('Options.PauseDelay'))
	end,
	['turnsrecoverybase'] = function()
		return gameOption('Options.Turns.Recovery.Base') .. '%'
	end,
	['turnsrecoverybonus'] = function()
		return gameOption('Options.Turns.Recovery.Bonus') .. '%'
	end,
	['vsync'] = function()
		return options.f_definedDisplay(gameOption('Video.VSync'), {[1] = motif.option_info.menu.valuename.on}, motif.option_info.menu.valuename.off)
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
					if getAction(-1, motif.option_info.menu.done.action) then
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
				if getAction(-1, motif.option_info.menu.done.action) then
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
					if getAction(-1, motif.option_info.menu.done.action) then
						sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
						options.f_applyVideo({['Video.FightAspectWidth'] = 0, ['Video.FightAspectHeight'] = 0})
						return false
					end
					return true
				end
			-- aspect ratio stage
			elseif v:match('_aspectratio_stageaspect$') then
				options.t_itemname['stageaspect'] = function(t, item, cursorPosY, moveTxt)
					if getAction(-1, motif.option_info.menu.done.action) then
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
	local t_pos = {} -- for storing current options.menu table position
	local lastNum = 0
	for i, suffix in ipairs(motif.option_info.menu.itemname_order) do
		for j, c in ipairs(main.f_strsplit('_', suffix)) do -- split using "_" delimiter
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
						if getAction(-1, motif.option_info.menu.done.action) then
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
			-- populate shaders submenu
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
			-- appending the menu table
			if j == 1 then -- first string after menu.itemname (either reserved one or custom submenu assignment)
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
			else -- following strings
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
		excluded = {renderer = true, gl33 = true, gles32 = true}
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
--; INPUT SETTINGS
--;===========================================================
-- Battle Controls: keys (keyboards) or virtual controls (gamepads) -> engine buttons.
-- Menu Controls and Controller Mapping: physical input -> virtual control.
-- The editors change a draft and apply it when the editor is left.
local function controlRow(prefix, control)
	return prefix .. control:lower():gsub('_', ''):gsub('%-', 'n'):gsub('%+', 'p')
end
local t_mappingRows = {}
for _, control in ipairs(main.virtualControls) do
	t_mappingRows[controlRow('vc', control)] = control
end
local t_setupRows = {}
for _, button in ipairs(main.engineButtons) do
	t_setupRows[button] = button
end
for path in pairs(motif.option_info.menu.itemname) do
	local button = path:match('buttonsetup_(.+)$')
	if button ~= nil then
		button = button:match('([^_]+)$')
		if inputIsSetupButton(button) then
			t_setupRows[button] = button
		end
	end
end
-- Keyboard Battle Controls also bind keys to sticks and triggers (Analog Emulation).
local t_keyboardSetupRows = main.f_tableCopy(t_setupRows)
for _, control in ipairs({'LS_X-', 'LS_X+', 'LS_Y-', 'LS_Y+', 'RS_X-', 'RS_X+', 'RS_Y-', 'RS_Y+', 'LT', 'RT'}) do
	t_keyboardSetupRows[controlRow('axis', control)] = control
end

-- Binding rows of an editor, by row name.
local function editorRows(root)
	if root.kind == 'mapping' then
		return t_mappingRows
	end
	return root.keyboard and t_keyboardSetupRows or t_setupRows
end

-- Hidden Input Settings rows still count, so editors can always find their templates.
local function findInputItem(tbl, name)
	for i, row in ipairs(tbl.allItems or tbl.items) do
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

-- Value shown for a binding.
local function bindingDisplay(token, kind)
	if token == nil or token == '' or token == '-' then
		return motif.option_info.menu.valuename.nokey
	end
	if kind == 'keyboard' then
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
	local rows = editorRows(root)
	local kind = root.device.kind
	if root.kind == 'setup' and not root.keyboard then
		kind = 'virtual'
	end
	for _, row in ipairs(tbl.items) do
		local field = row.itemname
		row.glyphDevice = root.device.id
		row.hintContext = 'editor'
		if rows[field] then
			row.hintContext = 'binding'
			row.vardisplay = bindingDisplay(root.draft[rows[field]], kind)
		elseif field == 'activesetup' then
			row.hintContext = 'setup'
			row.vardisplay = root.draft.name
		end
	end
	for _, sub in pairs(tbl.submenu) do
		updateEditor(sub, root)
	end
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
	if setupId == nil or root.kind ~= 'setup' or root.setupId == setupId then
		return
	end
	inputDeleteSetup(root.setupId)
	inputSetDeviceSetup(root.device.id, setupId)
	root.setupId = setupId
	if setupId == root.copyOf then
		root.copyOf = nil
	end
end

-- Applies the draft.
local function applyEditor(root)
	if root.kind == 'mapping' and not root.saveDefaults and sameSetup(root.draft, root.saved) then
		return true
	end
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
	-- Tracks if this frame rendered a menu, so the caller only flushes a frame that drew something.
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
	-- An unplugged gamepad closes its editor. Input Settings then select another device.
	local root = context.current and context.current.inputMenu
	local device = root and root.device
	if not context.closed and device and device.kind == 'gamepad' and not getJoystickPresent(device.joy) then
		context.capture, context.diagram, context.current, context.parents = nil, nil, nil, {}
		context.closed = true
	end
	if not context.closed and context.diagram ~= nil then
		options.f_inputDiagramFrame(context)
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
		textInputCancel()
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

-- Devices Input Settings can edit: keyboard controllers and connected gamepads.
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

-- Opens Menu Controls (keyboards) or Controller Mapping (gamepads) for a device.
function options.f_editMapping(device, context)
	if device == nil then
		return false
	end
	local itemname = device.kind == 'keyboard' and 'keyboardmenu' or 'controllermapping'
	local row, template = findInputItem(options.menu, itemname)
	if template == nil then
		return false
	end
	local tbl = main.f_tableCopy(template)
	tbl.kind = 'mapping'
	tbl.template = itemname
	tbl.device = device
	tbl.draft = inputGetMapping(device.id) or {}
	tbl.saved = main.f_tableCopy(tbl.draft)
	tbl.inputContext = context
	updateEditor(tbl, tbl)
	context.current, context.parents = tbl, {}
	resetKey()
	resetTokenGuard()
	return true
end

-- Loads a setup into the Battle Controls editor.
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

-- Opens the Battle Controls editor on the setup this device uses.
function options.f_editSetup(device, context)
	local row, template = findInputItem(options.menu, 'buttonsetup')
	if template == nil or device == nil then
		return false
	end
	local tbl = main.f_tableCopy(template)
	tbl.kind = 'setup'
	tbl.device = device
	tbl.keyboard = device.kind == 'keyboard'
	-- Gamepads have real sticks and triggers.
	if not tbl.keyboard then
		main.f_pruneMenu(tbl, {analogemulation = true})
	end
	tbl.inputContext = context
	if not loadSetup(tbl, device.setup) then
		return false
	end
	context.current, context.parents = tbl, {}
	resetKey()
	resetTokenGuard()
	return true
end

local function configAllButtons(root)
	local mapping = root.kind == 'mapping'
	local names = motif.option_info[mapping and root.template or 'buttonsetup'].configall
	local buttons = {}
	for _, name in ipairs(names) do
		buttons[name:upper()] = true
	end
	return buttons
end

-- Config All follows menu order and skips hidden or unlisted buttons.
local function startInputCapture(tbl, item, all)
	local root = tbl.inputMenu
	local rows = editorRows(root)
	local allowed = all and configAllButtons(root)
	local entries = {}
	local function collect(menu)
		for i, row in ipairs(menu.items) do
			local field = rows[row.itemname]
			if field and allowed[field:upper()] then
				table.insert(entries, {menu = menu, item = i, field = field})
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
	-- Holding cancel leaves, so every control can still be bound.
	if esc() or inputDeviceActionTime(device.id, sec.menu.cancel.action) >= sec.holdtime then
		sndPlay(motif.Snd, sec.cancel.snd[1], sec.cancel.snd[2])
		esc(false)
		undoCopy(root, capture.setupId)
		root.draft = capture.saved
		finished = true
	else
		-- Gamepad Battle Controls bind virtual controls, read through the device's Controller Mapping.
		-- Everything else binds keys or physical inputs.
		local token
		if root.kind == 'setup' and not root.keyboard then
			token = releasedInput(capture, device.id, 'virtual')
		else
			token = releasedInput(capture, device.id, device.kind)
		end
		if token ~= '' then
			-- An input binds one row only. The row that had it becomes unused. Other keyboard
			-- controllers may share keys, they just can't play at the same time.
			for field, value in pairs(root.draft) do
				if field ~= 'name' and field ~= 'id' and field ~= 'type' and field ~= entry.field and token ~= '-' and value == token then
					ownSetup(root)
					root.draft[field] = '-'
				end
			end
			if root.draft[entry.field] ~= token then
				ownSetup(root)
			end
			root.draft[entry.field] = token
			sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
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
--; INPUT DIAGRAMS ([Mapping Info], [Setup Info])
--;===========================================================
-- Virtual control names used by [Mapping Info].
local t_mappingArt = {
	DP_U = 'dpup', DP_D = 'dpdown', DP_L = 'dpleft', DP_R = 'dpright',
	A = 'a', B = 'b', X = 'x', Y = 'y', START = 'start', BACK = 'back',
	LB = 'lb', RB = 'rb', LT = 'lt', RT = 'rt',
	['LS_Y-'] = 'lsup', ['LS_Y+'] = 'lsdown', ['LS_X-'] = 'lsleft', ['LS_X+'] = 'lsright',
	['RS_Y-'] = 'rsup', ['RS_Y+'] = 'rsdown', ['RS_X-'] = 'rsleft', ['RS_X+'] = 'rsright',
	LS = 'ls', RS = 'rs',
}

-- The direction display lights one arrow, including diagonals, or its center.
local function diagramDirection(pressed, up, down, left, right)
	local vertical, horizontal = '', ''
	if up ~= down then vertical = up and 'up' or 'down' end
	if left ~= right then horizontal = left and 'left' or 'right' end
	local direction = vertical .. horizontal
	pressed[direction == '' and 'neutral' or direction] = true
end

local function setupDiagramState(root, deviceId)
	local held, analog = inputGetSetupState(deviceId, root.draft)
	local pressed = {}
	for i = 5, #main.engineButtons do
		local button = main.engineButtons[i]
		pressed[button] = held[button]
	end
	diagramDirection(pressed, held.up or analog['LS_Y-'], held.down or analog['LS_Y+'],
		held.left or analog['LS_X-'], held.right or analog['LS_X+'])
	return pressed
end

-- Leave the diagram screen.
local function closeInputDiagram(context, finish)
	if context.pause then
		finish()
		return
	end
	fadeOutInit(context.diagram.cfg.fadeout.FadeData)
	context.diagram.finish = function()
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

-- The input editors share the diagram screen and keep their current drafts.
function options.f_inputDiagram(context, mode, root)
	local device = root.device
	local diagram = {mode = mode, index = 1, root = root, name = device.name, device = device.id}
	-- A raw joystick's buttons may mean nothing yet, so Controller Mapping hints and held
	-- actions also use the first keyboard controller.
	if root.kind == 'mapping' and device.raw then
		for _, other in ipairs(inputGetDevices()) do
			if other.kind == 'keyboard' then
				diagram.keyboard = other.id
				break
			end
		end
	end
	diagram.setup = root.kind == 'setup'
	diagram.cfg = diagram.setup and motif.setup_info or motif.mapping_info
	diagram.background = diagram.setup and motif.setupbgdef or motif.mappingbgdef
	diagram.order = {}
	for key in main.f_sortKeys(diagram.cfg.control) do
		table.insert(diagram.order, key)
	end
	diagram.controls = {}
	if not diagram.setup then
		local allowed = configAllButtons(root)
		for _, control in ipairs(main.virtualControls) do
			if allowed[control] then
				table.insert(diagram.controls, control)
			end
		end
	end
	if mode == 'guided' and #diagram.controls == 0 then
		return
	end
	diagram.labels = mappingLabels(root)
	diagram.saved = main.f_tableCopy(root.draft)
	if mode == 'test' then
		diagram.index = 0
	end
	context.diagram = diagram
	if not context.pause then
		bgReset(diagram.background.BGDef)
		fadeInInit(diagram.cfg.fadein.FadeData)
	end
	resetKey()
	resetTokenGuard()
end

function options.f_inputDiagramFrame(context)
	local sec = motif.option_info
	local diagram = context.diagram
	local cfg = diagram.cfg
	local background = context.background
	if not context.pause then
		background = diagram.background
		clearColor(background.bgclearcolor[1], background.bgclearcolor[2], background.bgclearcolor[3])
	end
	local name = diagram.controls[diagram.index]
	local state = diagram.mode == 'test' and 'test' or 'default'
	local pressed = {}
	if diagram.mode == 'test' then
		if diagram.setup then
			pressed = setupDiagramState(diagram.root, diagram.device)
		else
			local held = inputGetVirtualState(diagram.device, diagram.root.draft)
			for _, vc in ipairs(main.virtualControls) do
				pressed[t_mappingArt[vc]] = held[vc]
			end
		end
	end
	bgDraw(background.BGDef, 0)
	rectDraw(cfg.overlay.RectData)
	textImgReset(cfg.title.TextSpriteData)
	textImgSetText(cfg.title.TextSpriteData, main.f_itemnameUpper(diagram.name, cfg.title.uppercase))
	textImgDraw(cfg.title.TextSpriteData)
	main.f_animPosDraw(cfg.bg.AnimData)
	for _, key in ipairs(diagram.order) do
		local art = cfg.control[key]
		local active = key == t_mappingArt[name] or pressed[key]
		local anim = active and art.active.AnimData or art.AnimData
		if anim ~= nil then
			main.f_animPosDraw(anim)
		end
	end
	local label = diagram.setup and diagram.root.draft.name or (diagram.labels[name or ''] or name or '')
	-- Config all has nothing left to ask while it fades out.
	if name ~= nil or state ~= 'default' then
		textImgReset(cfg.prompt.TextSpriteData)
		textImgSetText(cfg.prompt.TextSpriteData, main.f_formatBySpec(cfg.prompt.text[state], {
			s = label, i = diagram.index,
		}))
		textImgDraw(cfg.prompt.TextSpriteData)
	end
	bgDraw(background.BGDef, 1)
	main.f_hintDraw(cfg.hint, state, nil, diagram.keyboard or diagram.device)
	if diagram.finish ~= nil then
		if not fadeActive() then
			diagram.finish()
		end
		return
	end
	if context.skipInput or fadeActive() then
		return
	end
	local function close(finish)
		closeInputDiagram(context, function()
			context.diagram = nil
			if finish ~= nil then
				finish()
			end
			context.closed = context.current == nil
		end)
	end
	-- Held actions use the saved mapping of the edited device, the device running the menu
	-- and the keyboard controller of a raw joystick.
	local menuDevice = inputGetMenuDevice()
	local function holdTime(action)
		local time = inputDeviceActionTime(diagram.device, action)
		for _, id in ipairs({menuDevice, diagram.keyboard}) do
			if id ~= diagram.device then
				time = math.max(time, inputDeviceActionTime(id, action))
			end
		end
		return time
	end
	local function held(action)
		return holdTime(action) == cfg.holdtime
	end
	if esc() or holdTime(cfg.cancel.action) >= cfg.holdtime then
		-- Cancel restores the draft from before guided capture.
		sndPlay(motif.Snd, cfg.cancel.snd[1], cfg.cancel.snd[2])
		esc(false)
		if diagram.mode == 'guided' then
			diagram.root.draft = diagram.saved
			updateEditor(diagram.root, diagram.root)
		end
		close()
	elseif diagram.mode == 'test' then
		-- Read-only: inputs only highlight the diagram.
	elseif held(cfg.previous.action) and diagram.index > 1 then
		diagram.index, diagram.pending, diagram.released = diagram.index - 1, nil, false
		sndPlay(motif.Snd, cfg.move.snd[1], cfg.move.snd[2])
	elseif name ~= nil and held(cfg.next.action) then
		diagram.index, diagram.pending, diagram.released = diagram.index + 1, nil, false
		sndPlay(motif.Snd, cfg.move.snd[1], cfg.move.snd[2])
	elseif name ~= nil and held(cfg.clear.action) then
		diagram.root.draft[name] = '-'
		diagram.index, diagram.pending, diagram.released = diagram.index + 1, nil, false
		sndPlay(motif.Snd, cfg.move.snd[1], cfg.move.snd[2])
	elseif name ~= nil then
		-- Guided Controller Mapping: one physical input for each virtual control.
		local root = diagram.root
		local token = releasedInput(diagram, diagram.device, root.device.kind)
		if token ~= '' then
			for _, vc in ipairs(main.virtualControls) do
				if root.draft[vc] == token then
					root.draft[vc] = '-'
				end
			end
			root.draft[name] = token
			diagram.index = diagram.index + 1
			sndPlay(motif.Snd, cfg.done.snd[1], cfg.done.snd[2])
		end
	end
	if diagram.mode == 'guided' and diagram.index > #diagram.controls and diagram.finish == nil then
		updateEditor(diagram.root, diagram.root)
		close()
	end
	resetKey()
	resetTokenGuard()
end

-- Editor rows. One handler serves every editor: it reads the kind from the menu root.
-- Controller Mapping uses the guided controller screen, the other editors move the cursor row by row.
options.t_itemname.configall = function(t, item)
	if getAction(-1, motif.option_info.menu.done.action) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		if t.inputMenu.kind == 'mapping' and t.inputMenu.device.kind ~= 'keyboard' then
			options.f_inputDiagram(t.inputContext, 'guided', t.inputMenu)
		else
			startInputCapture(t, item, true)
		end
	end
	resetKey()
	return true
end

options.t_itemname.testmapping = function(t, item)
	if getAction(-1, motif.option_info.menu.done.action) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		options.f_inputDiagram(t.inputContext, 'test', t.inputMenu)
	end
	return true
end
options.t_itemname.testsetup = options.t_itemname.testmapping

local function bindingRow(t, item)
	local sec = motif.option_info
	local root = t.inputMenu
	if getAction(-1, sec.menu.subtract.action) then
		sndPlay(motif.Snd, sec.cursor.move.snd[1], sec.cursor.move.snd[2])
		local field = editorRows(root)[t.items[item].itemname]
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
for _, rows in ipairs({t_mappingRows, t_keyboardSetupRows}) do
	for name in pairs(rows) do
		options.t_itemname[name] = bindingRow
	end
end

-- Renames the setup. One this device doesn't own is copied first.
local function setupName(tbl)
	local sec, context, root = motif.option_info, tbl.inputContext, tbl.inputMenu
	context.popup = main.f_drawInput(sec.textinput.text.setupname, root.draft.name, true)
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

-- In an input editor, restore and defaults only change the draft, elsewhere they act on Options settings.
options.t_itemname.restoreprevious = function(t, item)
	if getAction(-1, motif.option_info.menu.done.action) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		local root = t.inputMenu
		if root == nil then
			options.f_restoreSettings(options.entryValues)
			return true
		end
		undoCopy(root, root.copyOf)
		root.saveDefaults = nil
		root.draft = main.f_tableCopy(root.saved)
		updateEditor(root, root)
	end
	return true
end

options.t_itemname.restoredefault = function(t, item)
	if getAction(-1, motif.option_info.menu.done.action) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		local root = t.inputMenu
		if root == nil then
			options.f_restoreSettings(options.f_settingsSnapshot(true))
			hook.run("options.default")
			return true
		end
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
			root.saveDefaults = root.kind == 'mapping'
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

-- Use Setup switches the device to another setup of its kind. Edits to the old one are kept.
-- Rename and delete act on the setup itself.
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
	elseif getAction(-1, sec.menu.item.rename.action) then
		sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
		setupName(t)
	elseif getAction(-1, sec.menu.item.delete.action) then
		-- The last setup of a device type stays. The device falls back to another setup and edits that one.
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

--;===========================================================
--; INPUT SETTINGS MENU
--;===========================================================
-- Input Settings edit one device at a time. Rows that don't apply to its kind are hidden.
function options.f_refreshInputSettings()
	local tbl = options.inputSettings
	if tbl == nil then
		return
	end
	local devices = mappingDevices()
	local device = devices[1]
	for _, v in ipairs(devices) do
		if options.inputDevice ~= nil and v.id == options.inputDevice.id then
			device = v
		end
	end
	options.inputDevice = device
	local kind = device and device.kind
	local hidden = {
		device = device == nil,
		buttonsetup = device == nil,
		keyboardmenu = kind ~= 'keyboard',
		controllermapping = kind ~= 'gamepad',
		preferredside = device == nil,
		glyphs = device == nil,
		vibration = device == nil or not device.hasRumble,
	}
	-- Menu loops keep their item lists, so they are refilled in place. Editor templates are copied whole.
	local editors = {buttonsetup = true, keyboardmenu = true, controllermapping = true}
	local function filter(menu)
		menu.allItems = menu.allItems or {unpack(menu.items)}
		for i = #menu.items, 1, -1 do
			menu.items[i] = nil
		end
		for _, row in ipairs(menu.allItems) do
			if not hidden[row.itemname] then
				table.insert(menu.items, row)
			end
			if menu.submenu[row.itemname] ~= nil and not editors[row.itemname] then
				filter(menu.submenu[row.itemname])
			end
		end
	end
	filter(tbl)
	local valuename = motif.option_info.menu.valuename
	for _, row in ipairs(tbl.items) do
		local field = row.itemname
		if field == 'device' then
			row.hintContext, row.vardisplay = 'device', device.name
		elseif field == 'preferredside' then
			row.hintContext, row.vardisplay = 'adjust', options.f_preferredSideDisplay(device.id)
		elseif field == 'vibration' then
			row.hintContext, row.vardisplay = 'adjust', options.f_boolDisplay(device.rumble)
		elseif field == 'glyphs' then
			for _, set in ipairs(inputGetGlyphPacks(device.id)) do
				if set.selected then
					row.vardisplay = set.automatic and valuename.glyphauto or (valuename.glyphpack:gsub('%%(%a+)', {id = set.id, name = set.name}))
					break
				end
			end
		end
	end
end

-- Preferred Side of a device, as shown in Input Settings.
function options.f_preferredSideDisplay(deviceId)
	local side = 0
	for _, device in ipairs(inputGetDevices()) do
		if device.id == deviceId then
			side = device.preferredSide
			break
		end
	end
	if side == 0 then
		return motif.option_info.menu.valuename.any
	end
	return main.f_formatBySpec(motif.option_info.menu.valuename.team, {i = side})
end

-- Opens Input Settings on the device that opened them.
function options.f_inputSettings(tbl)
	options.inputSettings = tbl
	options.inputDevice = deviceById(inputGetMenuDevice())
	options.f_refreshInputSettings()
	tbl.loop()
end

-- Opens Battle Controls or the device's mapping editor for the device selected in Input Settings.
function options.f_inputCfg(itemname)
	local context = {background = motif.optionbgdef, parents = {}}
	local device = options.inputDevice
	local ok
	if itemname == 'buttonsetup' then
		ok = options.f_editSetup(device, context)
	else
		ok = options.f_editMapping(device, context)
	end
	if ok then
		runInputConfig(context)
	end
	options.f_refreshInputSettings()
end

-- Device switches the device Input Settings edit. Rename changes the device's name.
options.t_itemname.device = function(t, item)
	local sec = motif.option_info
	local device = options.inputDevice
	if device == nil then
		return true
	end
	local step = 0
	if getAction(-1, sec.menu.add.action) then
		step = 1
	elseif getAction(-1, sec.menu.subtract.action) then
		step = -1
	end
	local devices = mappingDevices()
	if step ~= 0 and #devices > 1 then
		local index = 1
		for i, v in ipairs(devices) do
			if v.id == device.id then
				index = i
			end
		end
		sndPlay(motif.Snd, sec.cursor.move.snd[1], sec.cursor.move.snd[2])
		options.inputDevice = devices[(index - 1 + step) % #devices + 1]
		options.f_refreshInputSettings()
	elseif getAction(-1, sec.menu.item.rename.action) then
		sndPlay(motif.Snd, sec.cursor.done.snd[1], sec.cursor.done.snd[2])
		local name = main.f_drawInput(sec.textinput.text.devicename, device.name)
		esc(false)
		if name ~= '' and name ~= device.name and inputRenameDevice(device.id, name) then
			options.modified = true
			options.f_refreshInputSettings()
		end
	end
	return true
end

-- Preferred Side, Glyphs and Vibration belong to the device, so they are saved straight away.
options.t_itemname.preferredside = function(t, item)
	local step = 0
	if getAction(-1, motif.option_info.menu.add.action) then
		step = 1
	elseif getAction(-1, motif.option_info.menu.subtract.action) then
		step = -1
	end
	local device = options.inputDevice
	if step ~= 0 and device ~= nil then
		sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
		-- Any, Team 1, Team 2
		if inputSetPreferredSide(device.id, (device.preferredSide + step) % 3) then
			options.f_refreshInputSettings()
		end
	end
	return true
end

-- Opens the list of glyph packs this device can use, with the cursor on the current one.
options.t_itemname.glyphs = function(t, item)
	local device, list = options.inputDevice, t.submenu.glyphs
	if getAction(-1, motif.option_info.menu.done.action) and device ~= nil and list ~= nil then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		local valuename = motif.option_info.menu.valuename
		while list.items[1] ~= nil and list.items[1].pack ~= nil do
			table.remove(list.items, 1)
		end
		for i, set in ipairs(inputGetGlyphPacks(device.id)) do
			table.insert(list.items, i, {
				itemname = 'glyphpack',
				displayname = set.automatic and valuename.glyphauto or (valuename.glyphpack:gsub('%%(%a+)', {id = set.id, name = set.name})),
				paramname = t.items[item].paramname .. '_' .. (set.automatic and 'auto' or set.id),
				vardisplay = '',
				selected = set.selected,
				pack = set.id,
			})
			if set.selected then
				list.inputFocus = i
			end
		end
		list.loop()
		options.f_refreshInputSettings()
		main.f_menuSnap(motif.option_info)
		main.f_menuItemBgAnimReset(motif.option_info)
	end
	return true
end

-- The pack under the cursor is used and saved straight away, so glyphs change while moving.
options.t_itemname.glyphpack = function(t, item)
	local row, device = t.items[item], options.inputDevice
	if device ~= nil and not row.selected and inputSetGlyphPack(device.id, row.pack) then
		for _, v in ipairs(t.items) do
			v.selected = v == row
		end
	end
	if getAction(-1, motif.option_info.menu.done.action) then
		sndPlay(motif.Snd, motif.option_info.cursor.done.snd[1], motif.option_info.cursor.done.snd[2])
		return false
	end
	return true
end

-- Pads SDL can't vibrate keep it unchanged.
options.t_itemname.vibration = function(t, item)
	local device = options.inputDevice
	if getAction(-1, motif.option_info.menu.add.action, motif.option_info.menu.subtract.action) and device ~= nil and device.hasRumble then
		sndPlay(motif.Snd, motif.option_info.cursor.move.snd[1], motif.option_info.cursor.move.snd[2])
		if inputSetRumble(device.id, not device.rumble) then
			options.f_refreshInputSettings()
		end
	end
	return true
end

-- Side select: edit the Button Setup of the device that pressed the button.
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

-- Pause menu: the Battle Controls of the device in use, in the pause background.
function options.f_pauseSetup(background)
	local context = {background = background, pause = true, parents = {}}
	if options.f_editSetup(deviceById(inputGetMenuDevice()), context) then
		runInputConfig(context)
	end
end

return options

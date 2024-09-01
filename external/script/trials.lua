-- IKEMEN GO TRIALS MODE EXTERNAL MODULE --------------------------------
-- Last tested on Ikemen GO v0.99
-- Module developed by two4teezee
-------------------------------------------------------------------------
-- This external module implements TRIALS game mode (defeat all opponents
-- that are consider bosses). Features full screenpack integration via
-- system.def, ability to create and read trails for any character, and a
-- trials menu option, as well as a timer for the speed demons out there.
-- The trials mode and verification thresholds can be modified to suit your
-- custome game if needed. For more info on lua external modules:
-- https://github.com/K4thos/Ikemen_GO/wiki/Miscellaneous-Info#lua_modules
-- This mode is detectable by GameMode trigger as trials.
-- Only characters with a trials.def in their character folder will have
-- trials available for them; the character's def file also needs to be
-- modified to point to that trials.def. Documentation on how to use trials
-- mode is in README.md.
-------------------------------------------------------------------------

--;===========================================================
--; Local Functions
--;===========================================================

local function f_timeConvert(value)
	-- converts ticks to time
	local totalSec = value / config.GameFramerate
	local h = tostring(math.floor(totalSec / 3600))
	local m = tostring(math.floor((totalSec / 3600 - h) * 60))
	local s = tostring(math.floor(((totalSec / 3600 - h) * 60 - m) * 60))
	local x = tostring(math.floor((((totalSec / 3600 - h) * 60 - m) * 60 - s) *100))
	if string.len(m) < 2 then
		m = 0 .. m
	end
	if string.len(s) < 2 then
		s = 0 .. s
	end
	if string.len(x) < 2 then
		x = 0 .. x
	end
	return m, s, x
end

local function f_trimafterchar(line, char)
	-- trims a string after a specified character.
	-- also trims leading and trailing whitespace
	x = string.find(line, char)
	if x ~= nil then
		line = string.sub(line, x+1, #line)
		line = string.gsub(line, '^%s*(.-)%s*$', '%1')
		line = string.gsub(line, '[ \t]+%f[\r\n%z]', '')
	else
		line = ""
	end
	return line
end

local function f_strtoboolean(str)
	-- converts a table of "true" and "false" strings to bool
    local bool = {}
	for x = 1, #str, 1 do
		if string.lower(str[x]) == "true" then
			bool[x] = true
		else
			bool[x] = false
		end
	end
    return bool
end

local function f_strtonumber(str)
	-- converts a table of strings to numbers
    local array = {}
	for x = 1, #str, 1 do
		array[x] = tonumber(str[x])
	end
    return array
end

local function f_deepCopy(orig)
	-- copies a table into a local instance that can be modified freely
    local orig_type = type(orig)
    local copy
    if orig_type == 'table' then
        copy = {}
        for orig_key, orig_value in next, orig, nil do
            copy[f_deepCopy(orig_key)] = f_deepCopy(orig_value)
        end
        setmetatable(copy, f_deepCopy(getmetatable(orig)))
    else -- number, string, boolean, etc
        copy = orig
    end
    return copy
end

--;===========================================================
--; motif.lua
--;===========================================================
-- if motif.select_info.title_trials_text == nil then
-- 	motif.select_info.title_trials_text = 'Trials'
-- end

-- This code creates data out of optional [trialsbgdef] sff file.
-- Defaults to motif.files.spr_data, defined in screenpack, if not declared.
-- if motif.trialsbgdef.spr ~= nil and motif.trialsbgdef.spr ~= '' then
-- 	motif.trialsbgdef.spr = searchFile(motif.trialsbgdef.spr, {motif.fileDir, '', 'data/'})
-- 	motif.trialsbgdef.spr_data = sffNew(motif.trialsbgdef.spr)
-- else
-- 	motif.trialsbgdef.spr = motif.files.spr
-- 	motif.trialsbgdef.spr_data = motif.files.spr_data
-- end

-- Background data generation.
-- Refer to official Elecbyte docs for information how to define backgrounds.
-- http://www.elecbyte.com/mugendocs/bgs.html#description-of-background-elements
-- motif.trialsbgdef.bg = bgNew(motif.trialsbgdef.spr_data, motif.def, 'trialsbg')

-- fadein/fadeout anim data generation.
-- if motif.trials_mode.fadein_anim ~= -1 then
-- 	motif.f_loadSprData(motif.trials_mode, {s = 'fadein_'})
-- end
-- if motif.trials_mode.fadeout_anim ~= -1 then
-- 	motif.f_loadSprData(motif.trials_mode, {s = 'fadeout_'})
-- end

--;===========================================================
--; start.lua
--;===========================================================

function start.f_inittrialsData()
	start.trialsdata = {
		trialsExist = true,
		trialsInitialized = false,
		trialsPaused = false,
		trialAdvancement = true,
		trialsRemovalIndex = {},
		active = false,
		allclear = false,
		currenttrial = 1,
		currenttrialstep = 1,
		currenttrialmicrostep = 1,
		pauseuntilnexthit = false,
		combocounter = 0,
		maxsteps = 0,
		starttick = tickcount(),
		elapsedtime = 0,
		trial = f_deepCopy(start.f_getCharData(start.p[1].t_selected[1].ref).trialsdata),
		displaytimers = {
			totaltimer = true,
			trialtimer = true,
		},
	}

	-- Initialize trialAdvancement based on last-left menu value
	if menu.t_valuename.trialAdvancement[menu.trialAdvancement or 1].itemname == "Auto-Advance" then
		start.trialsdata.trialAdvancement = true
	else
		start.trialsdata.trialAdvancement = false
	end
end

function start.f_trialsBuilder()
	--start.f_trialsParser()
	--This function will initialize once to build all the trial tables based on the motif information and the trials information loaded when the char was selected
	--Populate background elements information
	start.trialsdata.bgelemdata = {
		currentbgsize = animGetSpriteInfo(motif.trials_mode.currentstep_bg_data),
		upcomingbgsize = animGetSpriteInfo(motif.trials_mode.upcomingstep_bg_data),
		completedbgsize = animGetSpriteInfo(motif.trials_mode.completedstep_bg_data),
		currentbgtailwidth = animGetSpriteInfo(motif.trials_mode.currentstep_bg_tail_data),
		currentbgheadwidth = animGetSpriteInfo(motif.trials_mode.currentstep_bg_head_data),
		upcomingbgtailwidth = animGetSpriteInfo(motif.trials_mode.upcomingstep_bg_tail_data),
		upcomingbgheadwidth = animGetSpriteInfo(motif.trials_mode.upcomingstep_bg_head_data),
		completedbgtailwidth = animGetSpriteInfo(motif.trials_mode.completedstep_bg_tail_data),
		completedbgheadwidth = animGetSpriteInfo(motif.trials_mode.completedstep_bg_head_data),
	}
	
	-- thin out trials data according to showforvarvalpairs
	for i = 1, #start.trialsdata.trial, 1 do
		--player(1)
		if #start.trialsdata.trial[i].showforvarvalpairs > 1 then
			valvarcheck = true
			for ii = 1, #start.trialsdata.trial[i].showforvarvalpairs, 2 do
				player(1)
				if var(start.trialsdata.trial[i].showforvarvalpairs[ii]) ~= start.trialsdata.trial[i].showforvarvalpairs[ii+1] then
					valvarcheck = false
				end
			end
			if not valvarcheck then
				start.trialsdata.trialsRemovalIndex[#start.trialsdata.trialsRemovalIndex+1] = i
			end
		end
	end
	for i = #start.trialsdata.trialsRemovalIndex, 1, -1 do
		table.remove(start.trialsdata.trial,start.trialsdata.trialsRemovalIndex[i])
	end

	--Obtain all of the trials information, to include the offset positions based on whether the display layout is horizontal or vertical
	for i = 1, #start.trialsdata.trial, 1 do
		
		if #start.trialsdata.trial[i].trialstep > start.trialsdata.maxsteps then
			start.trialsdata.maxsteps = #start.trialsdata.trial[i].trialstep
		end

		for j = 1, #start.trialsdata.trial[i].trialstep, 1 do
			--var-val pairs for each trialstep
			if #start.trialsdata.trial[i].trialstep[j].validforvarvalpairs > 1 then
				for ii = 1, #start.trialsdata.trial[i].trialstep[j].validforvarvalpairs, 2 do
					table.insert(start.trialsdata.trial[i].trialstep[j].validforvar,start.trialsdata.trial[i].trialstep[j].validforvarvalpairs[ii])
					table.insert(start.trialsdata.trial[i].trialstep[j].validforval,start.trialsdata.trial[i].trialstep[j].validforvarvalpairss[ii+1])
				end
			end

			local movelistline = start.trialsdata.trial[i].trialstep[j].glyphs
			for kk, v in main.f_sortKeys(motif.glyphs, function(t, a, b) return string.len(a) > string.len(b) end) do
				movelistline = movelistline:gsub(main.f_escapePattern(kk), '<' .. numberToRune(v[1] + 0xe000) .. '>')
			end
			movelistline = movelistline:gsub('%s+$', '')
			for moves in movelistline:gmatch('(	*[^	]+)') do
				moves = moves .. '<#>'
				tempglyphs = {}
				for m1, m2 in moves:gmatch('(.-)<([^%g <>]+)>') do
					if not m2:match('^#[A-Za-z0-9]+$') and not m2:match('^/$') and not m2:match('^#$') then
						tempglyphs[#tempglyphs+1] = m2
					end
				end
				if motif.trials_mode.glyphs_align == -1 then
					for ii = #tempglyphs, 1, -1 do
						start.trialsdata.trial[i].trialstep[j].glyphline.glyph[#start.trialsdata.trial[i].trialstep[j].glyphline.glyph+1] = tempglyphs[ii]
						start.trialsdata.trial[i].trialstep[j].glyphline.pos[#start.trialsdata.trial[i].trialstep[j].glyphline.glyph+1] = {0,0}
						start.trialsdata.trial[i].trialstep[j].glyphline.width[#start.trialsdata.trial[i].trialstep[j].glyphline.glyph+1] = 0
						start.trialsdata.trial[i].trialstep[j].glyphline.alignOffset[#start.trialsdata.trial[i].trialstep[j].glyphline.glyph+1] = 0
						start.trialsdata.trial[i].trialstep[j].glyphline.lengthOffset[#start.trialsdata.trial[i].trialstep[j].glyphline.glyph+1] = 0
						start.trialsdata.trial[i].trialstep[j].glyphline.scale[#start.trialsdata.trial[i].trialstep[j].glyphline.glyph+1] = {1,1}
					end
				else
					for ii = 1, #tempglyphs do
						start.trialsdata.trial[i].trialstep[j].glyphline.glyph[ii] = tempglyphs[ii]
						start.trialsdata.trial[i].trialstep[j].glyphline.pos[ii] = {0,0}
						start.trialsdata.trial[i].trialstep[j].glyphline.width[ii] = 0
						start.trialsdata.trial[i].trialstep[j].glyphline.alignOffset[ii] = 0
						start.trialsdata.trial[i].trialstep[j].glyphline.lengthOffset[ii] = 0
						start.trialsdata.trial[i].trialstep[j].glyphline.scale[ii] = {1,1}
					end
				end
			end
			--This glyphs section is more or less wholesale borrowed from the movelist section with minor tweaks
			local lengthOffset = 0
			local alignOffset = 0
			local align = 1
			local width = 0
			local font_def = 0
			--Some fonts won't give us the data we need to scale glyphs from, but sometimes that doesn't matter anyway
			if motif.trials_mode.currentstep_text_font[7] == nil and motif.trials_mode.glyphs_scalewithtext == "true" then
				font_def = main.font_def[motif.trials_mode.currentstep_text_font[1] .. motif.trials_mode.currentstep_text_font_height]
			elseif motif.trials_mode.glyphs_scalewithtext == "true" then
				font_def = main.font_def[motif.trials_mode.currentstep_text_font[1] .. motif.trials_mode.currentstep_text_font[7]]
			end
			for m in pairs(start.trialsdata.trial[i].trialstep[j].glyphline.glyph) do
				if motif.glyphs_data[start.trialsdata.trial[i].trialstep[j].glyphline.glyph[m]] ~= nil then
					if motif.trials_mode.trialslayout == "vertical" then
						if motif.trials_mode.glyphs_align == 0 then --center align
							alignOffset = motif.trials_mode.glyphs_offset[1] * 0.5
						elseif motif.trials_mode.glyphs_align == -1 then --right align
							alignOffset = motif.trials_mode.glyphs_offset[1]
						end
						if motif.trials_mode.glyphs_align ~= align then
							lengthOffset = 0
							align = motif.trials_mode.glyphs_align
						end
					end
					local scaleX = motif.trials_mode.glyphs_scale[1]
					local scaleY = motif.trials_mode.glyphs_scale[2]
					if motif.trials_mode.trialslayout == "vertical" and motif.trials_mode.glyphs_scalewithtext == "true" then
						scaleX = font_def.Size[2] * motif.trials_mode.currentstep_text_scale[2] / motif.glyphs_data[start.trialsdata.trial[i].trialstep[j].glyphline.glyph[m]].info.Size[2] * motif.trials_mode.glyphs_scale[1]
						scaleY = font_def.Size[2] * motif.trials_mode.currentstep_text_scale[2] / motif.glyphs_data[start.trialsdata.trial[i].trialstep[j].glyphline.glyph[m]].info.Size[2] * motif.trials_mode.glyphs_scale[2]
					end
					if motif.trials_mode.glyphs_align == -1 then
						alignOffset = alignOffset - motif.glyphs_data[start.trialsdata.trial[i].trialstep[j].glyphline.glyph[m]].info.Size[1] * scaleX
					end
					start.trialsdata.trial[i].trialstep[j].glyphline.alignOffset[m] = alignOffset
					start.trialsdata.trial[i].trialstep[j].glyphline.scale[m] = {scaleX, scaleY}
					start.trialsdata.trial[i].trialstep[j].glyphline.pos[m] = {
						math.floor(motif.trials_mode.trialsteps_pos[1] + motif.trials_mode.glyphs_offset[1] + alignOffset + lengthOffset),
						motif.trials_mode.trialsteps_pos[2] + motif.trials_mode.glyphs_offset[2]
					}
					start.trialsdata.trial[i].trialstep[j].glyphline.width[m] = math.floor(motif.glyphs_data[start.trialsdata.trial[i].trialstep[j].glyphline.glyph[m]].info.Size[1] * scaleX + motif.trials_mode.glyphs_spacing[1])
					if motif.trials_mode.glyphs_align == 1 then
						lengthOffset = lengthOffset + start.trialsdata.trial[i].trialstep[j].glyphline.width[m]
					elseif motif.trials_mode.glyphs_align == -1 then
						lengthOffset = lengthOffset - start.trialsdata.trial[i].trialstep[j].glyphline.width[m]
					else
						lengthOffset = lengthOffset + start.trialsdata.trial[i].trialstep[j].glyphline.width[m] / 2
					end
					start.trialsdata.trial[i].trialstep[j].glyphline.lengthOffset[m] = lengthOffset
				end
			end
		end
		if #start.trialsdata.trial[i].trialstep > start.trialsdata.maxsteps then
			start.trialsdata.maxsteps = #start.trialsdata.trial[i].trialstep
		end
	end
	--Pre-populate the draw table
	start.trialsdata.draw = {
		upcomingtextline = {},
		currenttextline = {},
		completedtextline = {},
		success = 0,
		fade = 0,
		fadein = 0,
		fadeout = 0,
		success_text = main.f_createTextImg(motif.trials_mode, 'success_text'),
		allclear = math.max(animGetLength(motif.trials_mode.allclear_front_data), animGetLength(motif.trials_mode.allclear_bg_data), motif.trials_mode.allclear_text_displaytime),
		allclear_text = main.f_createTextImg(motif.trials_mode, 'allclear_text'),
		trialcounter = main.f_createTextImg(motif.trials_mode, 'trialcounter'),
		totaltrialtimer = main.f_createTextImg(motif.trials_mode, 'totaltrialtimer'),
		currenttrialtimer = main.f_createTextImg(motif.trials_mode, 'currenttrialtimer'),
		trialtitle = math.max(animGetLength(motif.trials_mode.trialtitle_front_data), animGetLength(motif.trials_mode.trialtitle_bg_data)),
		trialtitle_text = main.f_createTextImg(motif.trials_mode, 'trialtitle_text'),
		windowXrange = motif.trials_mode.trialsteps_window[3] - motif.trials_mode.trialsteps_window[1],
		windowYrange = motif.trials_mode.trialsteps_window[4] - motif.trials_mode.trialsteps_window[2],
	}
	start.trialsdata.draw.success_text:update({x = motif.trials_mode.success_pos[1], y = motif.trials_mode.success_pos[2]+motif.trials_mode.success_text_offset[2],})
	start.trialsdata.draw.allclear_text:update({x = motif.trials_mode.allclear_pos[1]+motif.trials_mode.allclear_text_offset[1], y = motif.trials_mode.allclear_pos[2]+motif.trials_mode.allclear_text_offset[2],})
	start.trialsdata.draw.trialcounter:update({x = motif.trials_mode.trialcounter_pos[1], y = motif.trials_mode.trialcounter_pos[2],})
	start.trialsdata.draw.totaltrialtimer:update({x = motif.trials_mode.totaltrialtimer_pos[1], y = motif.trials_mode.totaltrialtimer_pos[2],})
	start.trialsdata.draw.currenttrialtimer:update({x = motif.trials_mode.currenttrialtimer_pos[1], y = motif.trials_mode.currenttrialtimer_pos[2],})
	start.trialsdata.draw.trialtitle_text:update({x = motif.trials_mode.trialtitle_pos[1]+motif.trials_mode.trialtitle_text_offset[1], y = motif.trials_mode.trialtitle_pos[2]+motif.trials_mode.trialtitle_text_offset[2],})
	for i = 1, start.trialsdata.maxsteps, 1 do
		start.trialsdata.draw.upcomingtextline[i] = main.f_createTextImg(motif.trials_mode, 'upcomingstep_text')
		start.trialsdata.draw.currenttextline[i] = main.f_createTextImg(motif.trials_mode, 'currentstep_text')
		start.trialsdata.draw.completedtextline[i] = main.f_createTextImg(motif.trials_mode, 'completedstep_text')
	end

	-- Build list out all of the available trials for Pause menu
	menu.t_valuename.trialsList = {}
	for i = 1, #start.trialsdata.trial, 1 do
		table.insert(menu.t_valuename.trialsList, {itemname = tostring(i), displayname = start.trialsdata.trial[i].name})
	end

	start.trialsdata.trialsInitialized = true
	if main.debugLog then main.f_printTable(start.trialsdata, "debug/t_trialsdata.txt") end
end

function start.f_trialsDummySetup()
	--If the trials initializer was successful and the round animation is completed, we will start drawing trials on the screen
	player(2)
	setAILevel(0)
	player(1)
	charMapSet(2, '_iksys_trialsDummyControl', 0)
	if not start.trialsdata.allclear and not start.trialsdata.trial[start.trialsdata.currenttrial].active then
		if start.trialsdata.trial[start.trialsdata.currenttrial].dummymode == 'stand' then
			charMapSet(2, '_iksys_trialsDummyMode', 0)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].dummymode == 'crouch' then
			charMapSet(2, '_iksys_trialsDummyMode', 1)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].dummymode == 'jump' then
			charMapSet(2, '_iksys_trialsDummyMode', 2)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].dummymode == 'wjump' then
			charMapSet(2, '_iksys_trialsDummyMode', 3)
		end
		if start.trialsdata.trial[start.trialsdata.currenttrial].guardmode == 'none' then
			charMapSet(2, '_iksys_trialsGuardMode', 0)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].guardmode == 'auto' then
			charMapSet(2, '_iksys_trialsGuardMode', 1)
		end
		if start.trialsdata.trial[start.trialsdata.currenttrial].buttonjam == 'none' then
			charMapSet(2, '_iksys_trialsButtonJam', 0)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].buttonjam == 'a' then
			charMapSet(2, '_iksys_trialsButtonJam', 1)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].buttonjam == 'b' then
			charMapSet(2, '_iksys_trialsButtonJam', 2)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].buttonjam == 'c' then
			charMapSet(2, '_iksys_trialsButtonJam', 3)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].buttonjam == 'x' then
			charMapSet(2, '_iksys_trialsButtonJam', 4)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].buttonjam == 'y' then
			charMapSet(2, '_iksys_trialsButtonJam', 5)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].buttonjam == 'z' then
			charMapSet(2, '_iksys_trialsButtonJam', 6)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].buttonjam == 'start' then
			charMapSet(2, '_iksys_trialsButtonJam', 7)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].buttonjam == 'd' then
			charMapSet(2, '_iksys_trialsButtonJam', 8)
		elseif start.trialsdata.trial[start.trialsdata.currenttrial].buttonjam == 'w' then
			charMapSet(2, '_iksys_trialsButtonJam', 9)
		end
		start.trialsdata.trial[start.trialsdata.currenttrial].active = true
	end
end

function start.f_trialsDrawer()
	if start.trialsdata.trialsInitialized and roundstate() == 2 and not start.trialsdata.active and start.trialsdata.draw.fade == 0 then
		start.f_trialsDummySetup()
		start.trialsdata.active = true
	end

	-- Check if game is paused - if so, set pause menu loop
	if paused() and not start.trialsdata.trialsPaused then
		start.trialsdata.trialsPaused = true
		menu.currentMenu = {menu.trials.loop, menu.trials.loop}
	elseif not paused() then
		start.trialsdata.trialsPaused = false
	end

	local accwidth = 0
	local addrow = 0
	-- Initialize abbreviated values for readability
	ct = start.trialsdata.currenttrial
	cts = start.trialsdata.currenttrialstep
	ctms = start.trialsdata.currenttrialmicrostep

	if start.trialsdata.active then
		if ct <= #start.trialsdata.trial and start.trialsdata.draw.success == 0 then

			--According to motif instructions, draw trials counter on screen
			local trtext = motif.trials_mode.trialcounter_text
			trtext = trtext:gsub('%%s', tostring(ct)):gsub('%%t', tostring(#start.trialsdata.trial))
			start.trialsdata.draw.trialcounter:update({text = trtext})
			start.trialsdata.draw.trialcounter:draw()
			--Logic for the stopwatches: total time spent in trial, and time spent on this current trial
			if start.trialsdata.displaytimers.totaltimer then
				local totaltimertext = motif.trials_mode.totaltrialtimer_text
				start.trialsdata.elapsedtime = tickcount() - start.trialsdata.starttick
				local m, s, x = f_timeConvert(start.trialsdata.elapsedtime)
				totaltimertext = totaltimertext:gsub('%%s', m .. ":" .. s .. ":" .. x)
				start.trialsdata.draw.totaltrialtimer:update({text = totaltimertext})
				start.trialsdata.draw.totaltrialtimer:draw()
			else
				--start.trialsdata.draw.totaltrialtimer:update({text = "Timer Disabled"})
				--start.trialsdata.draw.totaltrialtimer:draw()
			end
			if start.trialsdata.displaytimers.trialtimer then
				local currenttimertext = motif.trials_mode.currenttrialtimer_text
				start.trialsdata.trial[ct].elapsedtime = tickcount() - start.trialsdata.trial[ct].starttick
				local m, s, x = f_timeConvert(start.trialsdata.trial[ct].elapsedtime)
				currenttimertext = currenttimertext:gsub('%%s', m .. ":" .. s .. ":" .. x)
				start.trialsdata.draw.currenttrialtimer:update({text = currenttimertext})
				start.trialsdata.draw.currenttrialtimer:draw()
			else
				--start.trialsdata.draw.currenttrialtimer:update({text = "Timer Disabled"})
				--start.trialsdata.draw.currenttrialtimer:draw()
			end

			start.trialsdata.draw.trialtitle_text:update({text = start.trialsdata.trial[ct].name})
			start.trialsdata.draw.trialtitle_text:draw()
			animUpdate(motif.trials_mode.trialtitle_bg_data)
			animDraw(motif.trials_mode.trialtitle_bg_data)
			animUpdate(motif.trials_mode.trialtitle_front_data)
			animDraw(motif.trials_mode.trialtitle_front_data)

			local startonstep = 1
			local drawtothisstep = #start.trialsdata.trial[ct].trialstep

			--For vertical trial layouts, determine if all assets will be drawn within the trials window range, or if scrolling needs to be enabled. For horizontal layouts, we will figure it out
			--when we determine glyph and incrementor widths (see notes below). We do this step outside of the draw loop to speed things up.
			if #start.trialsdata.trial[ct].trialstep*motif.trials_mode.trialsteps_spacing[2] > start.trialsdata.draw.windowYrange and motif.trials_mode.trialslayout == "vertical" then
				startonstep = math.max(cts-2, 1)
				if (drawtothisstep - startonstep)*motif.trials_mode.trialsteps_spacing[2] > start.trialsdata.draw.windowYrange then
					drawtothisstep = math.min(startonstep+math.floor(start.trialsdata.draw.windowYrange/motif.trials_mode.trialsteps_spacing[2]),#start.trialsdata.trial[ct].trialstep)
				end
			end

			--This is the draw loop
			for i = startonstep, drawtothisstep, 1 do
				local tempoffset = {motif.trials_mode.trialsteps_spacing[1]*(i-startonstep),motif.trials_mode.trialsteps_spacing[2]*(i-startonstep)}
				--sub = 'current'
				if i < cts then
					sub = 'completed'
				elseif i == cts then
					sub = 'current'
				else
					sub = 'upcoming'
				end

				local bgtargetscale = {1,1}
				local bgcomponentposX = 0
				local padding = 0
				local totalglyphlength = 0
				local bgtailwidth = 0 --only used for horizontal layouts
				local bgheadwidth = 0 --only used for horizontal layouts

				if motif.trials_mode.trialslayout == "vertical" then
					--Vertical layouts are the simplest - they have a constant width sprite or anim that the text is drawn on top of, and the glyphs are displayed wherever specified.
					--The vertical layouts do NOT support incrementors (see notes below for horizontal layout).
					animSetPos(
						motif.trials_mode[sub .. 'step_bg_data'],
						motif.trials_mode.trialsteps_pos[1] + motif.trials_mode[sub .. 'step_bg_offset'][1] + tempoffset[1],
						motif.trials_mode.trialsteps_pos[2] + motif.trials_mode[sub .. 'step_bg_offset'][2] + tempoffset[2]
					)
					start.trialsdata.draw[sub .. 'textline'][i]:update({
						x = motif.trials_mode.trialsteps_pos[1]+motif.trials_mode.upcomingstep_text_offset[1]+motif.trials_mode.trialsteps_spacing[1]*(i-startonstep),
						y = motif.trials_mode.trialsteps_pos[2]+motif.trials_mode.upcomingstep_text_offset[2]+motif.trials_mode.trialsteps_spacing[2]*(i-startonstep),
						text = start.trialsdata.trial[ct].trialstep[i].text
					})
					animSetPalFX(motif.trials_mode[sub .. 'step_bg_data'], {
						time = 1,
						add = motif.trials_mode[sub .. 'step_bg_palfx_add'],
						mul = motif.trials_mode[sub .. 'step_bg_palfx_mul'],
						sinadd = motif.trials_mode[sub .. 'step_bg_palfx_sinadd'],
						invertall = motif.trials_mode[sub .. 'step_bg_palfx_invertall'],
						color = motif.trials_mode[sub .. 'step_bg_palfx_color']
					})
					animReset(motif.trials_mode[sub .. 'step_bg_data'])
					animUpdate(motif.trials_mode[sub .. 'step_bg_data'])
					animDraw(motif.trials_mode[sub .. 'step_bg_data'])
					start.trialsdata.draw[sub .. 'textline'][i]:draw()
				elseif motif.trials_mode.trialslayout == "horizontal" then
					--Horizontal layouts are much more complicated. Text is not drawn in horizontal mode, instead we only display the glyphs. A small sprite is dynamically tiled to the width of the
					--glyphs, and an optional background element called an incrementor (bginc) can be used to link the pieces together (think of an arrow where the body of the arrow is where the
					--glyphs are being drawn and that's the dynamically sized part, and the head of the arrow is the incrementor which is a fixed width sprite). There's quite a bit more work that
					--goes into displaying the horizontal layouts because the code needs to figure out the window size, and determine when it needs to "go to the next line" and create a return so
					--that trials can be displayed dynamically. Back to the arrow analogy, you always want an arrow body to have an arrow head, so the incrementor width is added to the glyphs length
					--and the padding factor specified in the motif data, it's all added together until the window width is met or exceeded, then a line return occurs and the next line is drawn.
					local bgsize = {0,0}
					if start.trialsdata.bgelemdata[sub .. 'bgtailwidth'] ~= nil then bgtailwidth = math.floor(start.trialsdata.bgelemdata[sub .. 'bgtailwidth'].Size[1]) end
					if start.trialsdata.bgelemdata[sub .. 'bgheadwidth'] ~= nil then bgheadwidth = math.floor(start.trialsdata.bgelemdata[sub .. 'bgheadwidth'].Size[1]) end
					if start.trialsdata.bgelemdata[sub .. 'bgsize'] ~= nil then bgsize = start.trialsdata.bgelemdata[sub .. 'bgsize'].Size end

					totalglyphlength = start.trialsdata.trial[ct].trialstep[i].glyphline.lengthOffset[#start.trialsdata.trial[ct].trialstep[i].glyphline.lengthOffset]
					local tailoffset = motif.trials_mode[sub .. 'step_bg_tail_offset'][1]
					padding = motif.trials_mode.trialsteps_horizontal_padding
					spacing = motif.trials_mode.trialsteps_spacing[1]

					local tempwidth = spacing + bgtailwidth + tailoffset + padding + totalglyphlength + padding + bgheadwidth + accwidth
					if tempwidth - motif.trials_mode.trialsteps_spacing[1] > start.trialsdata.draw.windowXrange then
						accwidth = 0
						addrow = addrow + 1
					end

					tempoffset[2] = motif.trials_mode.trialsteps_spacing[2]*(addrow)

					-- Calculate initial positions
					if accwidth == 0 then
						bgcomponentposX = motif.trials_mode.trialsteps_pos[1] + motif.trials_mode[sub .. 'step_bg_tail_offset'][1]
					else
						bgcomponentposX = accwidth + spacing - bgheadwidth + bgtailwidth + motif.trials_mode[sub .. 'step_bg_tail_offset'][1]
					end
					
					-- Draw tail
					animSetPos(motif.trials_mode[sub .. 'step_bg_tail_data'], 
						bgcomponentposX, 
						start.trialsdata.trial[ct].trialstep[i].glyphline.pos[1][2] + motif.trials_mode[sub .. 'step_bg_tail_offset'][2] + tempoffset[2]
					)
					animSetPalFX(motif.trials_mode[sub .. 'step_bg_tail_data'], {
						time = 1,
						add = motif.trials_mode[sub .. 'step_bg_palfx_add'],
						mul = motif.trials_mode[sub .. 'step_bg_palfx_mul'],
						sinadd = motif.trials_mode[sub .. 'step_bg_palfx_sinadd'],
						invertall = motif.trials_mode[sub .. 'step_bg_palfx_invertall'],
						color = motif.trials_mode[sub .. 'step_bg_palfx_color']
					})
					animReset(motif.trials_mode[sub .. 'step_bg_tail_data'])
					animUpdate(motif.trials_mode[sub .. 'step_bg_tail_data'])
					animDraw(motif.trials_mode[sub .. 'step_bg_tail_data'])
					
					-- Draw BG for Glyphs - scale to length, start from tail pos
					bgtargetscale = {(padding + totalglyphlength + padding)/bgsize[1], 1}
					bgcomponentposX = bgcomponentposX + bgtailwidth + motif.trials_mode[sub .. 'step_bg_offset'][1]
					local gpoffset = 0
					for m in pairs(start.trialsdata.trial[ct].trialstep[i].glyphline.glyph) do
						if m > 1 then gpoffset = start.trialsdata.trial[ct].trialstep[i].glyphline.lengthOffset[m-1] end
						start.trialsdata.trial[ct].trialstep[i].glyphline.pos[m][1] = bgcomponentposX + padding + gpoffset -- motif.trials_mode.trialsteps_pos[1] + start.trialsdata.trial[ct].trialstep[i].glyphline.alignOffset[m] +
					end

					animSetScale(motif.trials_mode[sub .. 'step_bg_data'], bgtargetscale[1], bgtargetscale[2])
					animSetPos(motif.trials_mode[sub .. 'step_bg_data'], 
						bgcomponentposX, 
						start.trialsdata.trial[ct].trialstep[i].glyphline.pos[1][2] + motif.trials_mode[sub .. 'step_bg_offset'][2] + tempoffset[2]
					)
					animSetPalFX(motif.trials_mode[sub .. 'step_bg_data'], {
						time = 1,
						add = motif.trials_mode[sub .. 'step_bg_palfx_add'],
						mul = motif.trials_mode[sub .. 'step_bg_palfx_mul'],
						sinadd = motif.trials_mode[sub .. 'step_bg_palfx_sinadd'],
						invertall = motif.trials_mode[sub .. 'step_bg_palfx_invertall'],
						color = motif.trials_mode[sub .. 'step_bg_palfx_color']
					})
					animReset(motif.trials_mode[sub .. 'step_bg_data'])
					animUpdate(motif.trials_mode[sub .. 'step_bg_data'])
					animDraw(motif.trials_mode[sub .. 'step_bg_data'])
					
					-- Draw head
					bgcomponentposX = bgcomponentposX + start.trialsdata.trial[ct].trialstep[i].glyphline.alignOffset[1] + (totalglyphlength + 2*padding) + motif.trials_mode[sub .. 'step_bg_head_offset'][1]
					animSetPos(motif.trials_mode[sub .. 'step_bg_head_data'], 
						bgcomponentposX, 
						start.trialsdata.trial[ct].trialstep[i].glyphline.pos[1][2] + motif.trials_mode[sub .. 'step_bg_head_offset'][2] + tempoffset[2]
					)
					animSetPalFX(motif.trials_mode[sub .. 'step_bg_head_data'], {
						time = 1,
						add = motif.trials_mode[sub .. 'step_bg_palfx_add'],
						mul = motif.trials_mode[sub .. 'step_bg_palfx_mul'],
						sinadd = motif.trials_mode[sub .. 'step_bg_palfx_sinadd'],
						invertall = motif.trials_mode[sub .. 'step_bg_palfx_invertall'],
						color = motif.trials_mode[sub .. 'step_bg_palfx_color']
					})
					animReset(motif.trials_mode[sub .. 'step_bg_head_data'])
					animUpdate(motif.trials_mode[sub .. 'step_bg_head_data'])
					animDraw(motif.trials_mode[sub .. 'step_bg_head_data'])
				end
				for m = 1, #start.trialsdata.trial[ct].trialstep[i].glyphline.glyph, 1 do
					animSetScale(motif.glyphs_data[start.trialsdata.trial[ct].trialstep[i].glyphline.glyph[m]].anim, start.trialsdata.trial[ct].trialstep[i].glyphline.scale[m][1], start.trialsdata.trial[ct].trialstep[i].glyphline.scale[m][2])
					animSetPos(motif.glyphs_data[start.trialsdata.trial[ct].trialstep[i].glyphline.glyph[m]].anim, 
						start.trialsdata.trial[ct].trialstep[i].glyphline.pos[m][1], 
						start.trialsdata.trial[ct].trialstep[i].glyphline.pos[m][2] + tempoffset[2] + motif.trials_mode.glyphs_offset[2]
					)
					animSetPalFX(motif.glyphs_data[start.trialsdata.trial[ct].trialstep[i].glyphline.glyph[m]].anim, {
						time = 1,
						add = motif.trials_mode[sub .. 'step_glyphs_palfx_add'],
						mul = motif.trials_mode[sub .. 'step_glyphs_palfx_mul'],
						sinadd = motif.trials_mode[sub .. 'step_glyphs_palfx_sinadd'],
						invertall = motif.trials_mode[sub .. 'step_glyphs_palfx_invertall'],
						color = motif.trials_mode[sub .. 'step_glyphs_palfx_color']
					})
					animReset(motif.glyphs_data[start.trialsdata.trial[ct].trialstep[i].glyphline.glyph[m]].anim)
					animUpdate(motif.glyphs_data[start.trialsdata.trial[ct].trialstep[i].glyphline.glyph[m]].anim)
					animDraw(motif.glyphs_data[start.trialsdata.trial[ct].trialstep[i].glyphline.glyph[m]].anim)
				end
				accwidth = bgcomponentposX
			end
		elseif ct > #start.trialsdata.trial then
			-- All trials have been completed, draw the all clear and freeze the timer
			if start.trialsdata.draw.allclear ~= 0 then
				start.f_trialsSuccess('allclear', ct-1)
				main.f_createTextImg(motif.trials_mode, 'allclear_text')
			end

			start.trialsdata.allclear = true
			start.trialsdata.draw.success = 0
			start.trialsdata.draw.trialcounter:update({text = motif.trials_mode.trialcounter_allclear_text})
			start.trialsdata.draw.trialcounter:draw()

			if start.trialsdata.displaytimers.totaltimer then
				local totaltimertext = motif.trials_mode.totaltrialtimer_text
				local m, s, x = f_timeConvert(start.trialsdata.elapsedtime)
				totaltimertext = totaltimertext:gsub('%%s', m .. ":" .. s .. ":" .. x)
				start.trialsdata.draw.totaltrialtimer:update({text = totaltimertext})
				start.trialsdata.draw.totaltrialtimer:draw()
			else
				--start.trialsdata.draw.totaltrialtimer:update({text = "Timer Disabled"})
				--start.trialsdata.draw.totaltrialtimer:draw()
			end
			if start.trialsdata.displaytimers.trialtimer then
				local currenttimertext = motif.trials_mode.currenttrialtimer_text
				local m, s, x = f_timeConvert(start.trialsdata.trial[ct-1].elapsedtime)
				currenttimertext = currenttimertext:gsub('%%s', m .. ":" .. s .. ":" .. x)
				start.trialsdata.draw.currenttrialtimer:update({text = currenttimertext})
				start.trialsdata.draw.currenttrialtimer:draw()
			else
				--start.trialsdata.draw.currenttrialtimer:update({text = "Timer Disabled"})
				--start.trialsdata.draw.currenttrialtimer:draw()
			end
		end
	end
end

function start.f_trialsChecker()
	--This function sets dummy actions according to the character trials info and validates trials attempts
	--To help follow along, ct = current trial, cts = current trial step, ncts = next current trial step
	if ct <= #start.trialsdata.trial and start.trialsdata.draw.success == 0 and start.trialsdata.draw.fade == 0 and start.trialsdata.active then
		local helpercheck = false
		local projcheck = false
		local maincharcheck = false
		player(2)
		local attackerid = gethitvar('id')
		player(1)
		local attackerstate = nil
		local attackeranim = nil
		if attackerid > 0 then
			playerid(attackerid)
			attackerstate = stateno()
			attackeranim = anim()
			player(1)
			-- Can uncomment this section to debug helper/proj data
			-- print("ID: " .. attackerid)
			-- print("State: " .. attackerstate)
			-- print("Anim: " .. attackeranim)
		end

		if (start.trialsdata.trial[ct].trialstep[cts].ishelper[ctms] and start.trialsdata.trial[ct].trialstep[cts].stateno[ctms] == attackerstate) and (attackeranim == start.trialsdata.trial[ct].trialstep[cts].animno[ctms] or start.trialsdata.trial[ct].trialstep[cts].animno[ctms] == nil) then
			helpercheck = true
		end

		if (start.trialsdata.trial[ct].trialstep[cts].isproj[ctms] and start.trialsdata.trial[ct].trialstep[cts].stateno[ctms] == attackerstate) and (attackeranim == start.trialsdata.trial[ct].trialstep[cts].animno[ctms] or start.trialsdata.trial[ct].trialstep[cts].animno[ctms] == nil) then
			projcheck = true
		end

		maincharcheck = (stateno() == start.trialsdata.trial[ct].trialstep[cts].stateno[ctms] and not(start.trialsdata.trial[ct].trialstep[cts].isproj[ctms]) and not(start.trialsdata.trial[ct].trialstep[cts].ishelper[ctms]) and (anim() == start.trialsdata.trial[ct].trialstep[cts].animno[ctms] or start.trialsdata.trial[ct].trialstep[cts].animno[ctms] == nil) and ((hitpausetime() > 1 and movehit() and combocount() > start.trialsdata.combocounter) or start.trialsdata.trial[ct].trialstep[cts].isthrow[ctms] or start.trialsdata.trial[ct].trialstep[cts].isnohit[ctms]))
		
		--Check val-var pairs if specified
		if start.trialsdata.trial[ct].trialstep[cts].validforvarvalpairs ~= nil and maincharcheck then
			for i = 1, #start.trialsdata.trial[ct].trialstep[cts].validforvar, 1 do
				if maincharcheck then
					maincharcheck = var(start.trialsdata.trial[ct].trialstep[cts].validforvar[i]) == start.trialsdata.trial[ct].trialstep[cts].validforval[i]
				end
			end
		end
		
		if maincharcheck or projcheck or helpercheck then
			if start.trialsdata.trial[ct].trialstep[cts].numofhits[ctms] >= 1 then
				if start.trialsdata.trial[ct].trialstep[cts].stephitscount[ctms] == 0 then
					start.trialsdata.trial[ct].trialstep[cts].combocountonstep[ctms] = combocount()
				end
				if combocount() - start.trialsdata.trial[ct].trialstep[cts].stephitscount[ctms] == start.trialsdata.trial[ct].trialstep[cts].combocountonstep[ctms] then
					start.trialsdata.trial[ct].trialstep[cts].stephitscount[ctms] = start.trialsdata.trial[ct].trialstep[cts].stephitscount[ctms] + 1
				end
			elseif start.trialsdata.trial[ct].trialstep[cts].numofhits[ctms] == 0 then
				start.trialsdata.trial[ct].trialstep[cts].stephitscount[ctms] = 0
			end

			if start.trialsdata.trial[ct].trialstep[cts].numofhits[ctms] == start.trialsdata.trial[ct].trialstep[cts].stephitscount[ctms] then
				nctms = ctms + 1
				-- First, check that the microstep has passed
				if nctms >= 1 and ((combocount() > 0 and (start.trialsdata.trial[ct].trialstep[cts].iscounterhit[ctms] and movecountered() > 0) or not start.trialsdata.trial[ct].trialstep[cts].iscounterhit[ctms]) or start.trialsdata.trial[ct].trialstep[cts].isnohit[ctms]) then
					if nctms >= 1 and ((start.trialsdata.trial[ct].trialstep[cts].numofhits[ctms] > 1 and combocount() == start.trialsdata.trial[ct].trialstep[cts].stephitscount[ctms] + start.trialsdata.trial[ct].trialstep[cts].combocountonstep[ctms] - 1) or start.trialsdata.trial[ct].trialstep[cts].numofhits[ctms] == 1 or start.trialsdata.trial[ct].trialstep[cts].isnohit[ctms]) then
						start.trialsdata.currenttrialmicrostep = nctms
						start.trialsdata.pauseuntilnexthit = start.trialsdata.trial[ct].trialstep[cts].validuntilnexthit[ctms]
						start.trialsdata.combocounter = combocount()
					elseif ((combocount() == 0 and not start.trialsdata.trial[ct].trialstep[cts].isnohit[ctms]) and not start.trialsdata.pauseuntilnexthit) or (start.trialsdata.pauseuntilnexthit and combocount() > start.trialsdata.combocounter) then
						start.trialsdata.currenttrialstep = 1
						start.trialsdata.currenttrialmicrostep = 1
						start.trialsdata.trial[ct].trialstep[cts].stephitscount[ctms] = 0
						start.trialsdata.trial[ct].trialstep[cts].combocountonstep[ctms] = 0
						start.trialsdata.combocounter = 0
					end
				end
				-- Next, if microstep is exceeded, go to next trial step
				if start.trialsdata.currenttrialmicrostep > start.trialsdata.trial[ct].trialstep[cts].numofmicrosteps then
					start.trialsdata.currenttrialmicrostep = 1
					start.trialsdata.currenttrialstep = cts + 1
					start.trialsdata.combocounter = combocount()
					start.trialsdata.pauseuntilnexthit = start.trialsdata.trial[ct].trialstep[cts].validuntilnexthit[ctms]
					if start.trialsdata.currenttrialstep > #start.trialsdata.trial[ct].trialstep then
						-- If trial step was last, go to next trial and display success banner
						if start.trialsdata.trialAdvancement then
							start.trialsdata.currenttrial = ct + 1
						end
						start.trialsdata.currenttrialstep = 1
						start.trialsdata.combocounter = 0
						if ct < #start.trialsdata.trial or (not start.trialsdata.trialAdvancement and ct == #start.trialsdata.trial) then
							if (motif.trials_mode.success_front_displaytime == -1) and (motif.trials_mode.success_bg_displaytime == -1) then
								start.trialsdata.draw.success = math.max(animGetLength(motif.trials_mode.success_front_data), animGetLength(motif.trials_mode.success_bg_data), motif.trials_mode.success_text_displaytime)
							else
								start.trialsdata.draw.success = math.max(motif.trials_mode.success_front_displaytime, motif.trials_mode.success_bg_displaytime, motif.trials_mode.success_text_displaytime)
							end
							if motif.trials_mode.resetonsuccess == "true" then
								start.trialsdata.draw.fadein = motif.trials_mode.fadein_time
								start.trialsdata.draw.fadeout = motif.trials_mode.fadeout_time
								start.trialsdata.draw.fade = start.trialsdata.draw.fadein + start.trialsdata.draw.fadeout
							end
						end
					end
				end
			end
		elseif ((combocount() == 0 and not start.trialsdata.trial[ct].trialstep[cts].isnohit[ctms]) and not start.trialsdata.pauseuntilnexthit) or (start.trialsdata.pauseuntilnexthit and combocount() > start.trialsdata.combocounter) then
			start.trialsdata.currenttrialstep = 1
			start.trialsdata.currenttrialmicrostep = 1
			start.trialsdata.combocounter = 0
			start.trialsdata.trial[ct].trialstep[cts].stephitscount[ctms] = 0
			start.trialsdata.trial[ct].trialstep[cts].combocountonstep[ctms] = 0
			start.trialsdata.pauseuntilnexthit = false
		end
	end
	--If the trial was completed successfully, draw the trials success
	if start.trialsdata.draw.success > 0 then
		start.f_trialsSuccess('success', ct)
	elseif start.trialsdata.draw.fade > 0 and motif.trials_mode.resetonsuccess == "true" then
		if start.trialsdata.draw.fade < start.trialsdata.draw.fadein + start.trialsdata.draw.fadeout then
			start.f_trialsFade()
		else
			player(2)
			if stateno() == 0 then
				start.f_trialsFade()
			end
			player(1)
		end
	end
end

function start.f_trialsSuccess(successstring, index)
	-- This function is responsible for drawing the Success or All Clear banners after a trial is completed successfully.
	charMapSet(2, '_iksys_trialsDummyMode', 0)
	charMapSet(2, '_iksys_trialsGuardMode', 0)
	charMapSet(2, '_iksys_trialsButtonJam', 0)
	if not start.trialsdata.trial[index].complete or (successstring == "allclear" and not start.trialsdata.allclear) then
		-- Play sound only once
		sndPlay(motif.files.snd_data, motif.trials_mode[successstring .. '_snd'][1], motif.trials_mode[successstring .. '_snd'][2])
	end
	animUpdate(motif.trials_mode[successstring .. '_bg_data'])
	animDraw(motif.trials_mode[successstring .. '_bg_data'])
	animUpdate(motif.trials_mode[successstring .. '_front_data'])
	animDraw(motif.trials_mode[successstring .. '_front_data'])
	start.trialsdata.draw[successstring .. '_text']:draw()
	start.trialsdata.draw[successstring] = start.trialsdata.draw[successstring] - 1
	start.trialsdata.trial[index].complete = true
	start.trialsdata.trial[index].active = false
	start.trialsdata.active = false
	if not start.trialsdata.trialAdvancement then
		start.trialsdata.trial[index].starttick = tickcount()
	end
	if index ~= #start.trialsdata.trial then
		start.trialsdata.trial[index+1].starttick = tickcount()
	end
end

function start.f_trialsFade()
	-- This function is responsible for fadein/fadeout if resetonsuccess is set to true.
	if start.trialsdata.draw.fadeout > 0 then
		if not main.fadeActive then
			main.f_fadeReset('fadeout',motif.trials_mode)
		end
		main.f_fadeAnim(motif.trials_mode)
		start.trialsdata.draw.fadeout = start.trialsdata.draw.fadeout - 1
	elseif start.trialsdata.draw.fadein > 0 then
		if main.fadeType == 'fadeout' then
			charMapSet(2, '_iksys_trialsReposition', 1)
			main.f_fadeReset('fadein',motif.trials_mode)
		elseif main.fadeType == 'fadein' then
			charMapSet(2, '_iksys_trialsCameraReset', 1)
		end
		main.f_fadeAnim(motif.trials_mode)
		start.trialsdata.draw.fadein = start.trialsdata.draw.fadein - 1
	end
	
	start.trialsdata.draw.fade = start.trialsdata.draw.fade - 1
end

function start.f_trialsMode()
	if roundstart() then
		start.trialsdata = nil
		-- Check if there's a trials file - if so, parse it
		if start.f_getCharData(start.p[1].t_selected[1].ref).trialsdef ~= "" then
			start.f_inittrialsData()
			trialsExist = true
 		else
			trialsExist = false
		end
	end

	if trialsExist and roundstate() == 2 and not start.trialsdata.trialsInitialized then
		-- Initialize the trials based on parsed file and char state at roundstate() == 2
		start.f_trialsBuilder()
		menu.f_trialsReset()
	elseif trialsExist and roundstate() == 2 and start.trialsdata.trialsInitialized then
		-- If trials initialized, draw elements and check for success!
		start.f_trialsDrawer()
		start.f_trialsChecker()
	elseif roundstate() == 2 then
		-- No trials present!
		player(2)
		setAILevel(0)
		player(1)
		charMapSet(2, '_iksys_trialsDummyControl', 0)
		trialcounter = main.f_createTextImg(motif.trials_mode, 'trialcounter')
		trialcounter:update({x = motif.trials_mode.trialcounter_pos[1], y = motif.trials_mode.trialcounter_pos[2], text = motif.trials_mode.trialcounter_notrialsdata_text})
		trialcounter:draw()
	end
end

--;===========================================================
--; trials.lua
--;===========================================================

-- Find trials files and parse them; append t_selChars table
for row = 1, #main.t_selChars, 1 do
	if main.t_selChars[row].def ~= nil then
		main.t_selChars[row].trialsdef = ""
		local deffile = io.open(main.t_selChars[row].def, "r")
		for line in deffile:lines() do
			line = line:gsub('%s*;.*$', '')
			lcline = string.lower(line)
			if lcline:match('trials') then
				main.t_selChars[row].trialsdef = main.t_selChars[row].dir .. "/" .. f_trimafterchar(line, "=")
			end
		end
		deffile:close()
	end
	if  main.t_selChars[row].def ~= nil and main.t_selChars[row].trialsdef ~= "" then
		i = 0 --Trial number
		j = 0 --TrialStep number
		trial = {}

		local trialsFile = io.open(main.t_selChars[row].trialsdef, "r")

		for line in trialsFile:lines() do
			line = line:gsub('%s*;.*$', '')
			lcline = string.lower(line)

			if lcline:find("trialstep." .. j+1 .. ".") then
				j = j + 1
				trial[i].trialstep[j] = {
					numofmicrosteps = 1,
					text = "",
					glyphs = "",
					stateno = {},
					animno = {},
					numofhits = {},
					stephitscount = {},
					combocountonstep = {},
					isthrow = {},
					isnohit = {},
					ishelper = {},
					isproj = {},
					iscounterhit = {},
					validuntilnexthit = {},
					validforvarvalpairs = {nil},
					validforvar = {},
					validforval = {},
					glyphline = {
						glyph = {},
						pos = {},
						width = {},
						alignOffset = {},
						lengthOffset = {},
						scale = {},
					},
				}
			end 

			if line:match('^%s*%[.-%s*%]%s*$') then --matched [] group
				line = line:match('^%s*%[(.-)%s*%]%s*$') --match text between []
				lcline = string.lower(line)
				if lcline:match('^trialdef') then --matched trialdef block
					i = i + 1 -- increment Trial number
					j = 0 -- reset trialstep number
					trial[i] = {
						name = "",
						dummymode = "stand",
						guardmode = "none",
						buttonjam = "none",
						active = false,
						complete = false,
						showforvarvalpairs = {nil},
						elapsedtime = 0,
						starttick = tickcount(),
						trialstep = {},
					}
					line = f_trimafterchar(line, ",")
					if line == "" then
						line = "Trial " .. tostring(i)
					end
					trial[i].name = line
				end
			elseif lcline:find("dummymode") then
				trial[i].dummymode = f_trimafterchar(lcline, "=")
			elseif lcline:find("guardmode") then
				trial[i].guardmode = f_trimafterchar(lcline, "=")
			elseif lcline:find("dummybuttonjam") then
				trial[i].buttonjam = f_trimafterchar(lcline, "=")
			elseif lcline:find("showforvarvalpairs") then
				trial[i].showforvarvalpairs = f_strtonumber(main.f_strsplit(',', string.gsub(f_trimafterchar(lcline, "="),"%s+", "")))
			elseif lcline:find("trialstep." .. j .. ".text") then
				trial[i].trialstep[j].text = f_trimafterchar(line, "=")
			elseif lcline:find("trialstep." .. j .. ".glyphs") then
				trial[i].trialstep[j].glyphs = f_trimafterchar(line, "=")
			elseif lcline:find("trialstep." .. j .. ".stateno") then
				trial[i].trialstep[j].stateno = f_strtonumber(main.f_strsplit(',', string.gsub(f_trimafterchar(lcline, "="),"%s+", "")))
				trial[i].trialstep[j].numofmicrosteps = #trial[i].trialstep[j].stateno
				for k = 1, trial[i].trialstep[j].numofmicrosteps, 1 do
					trial[i].trialstep[j].stephitscount[k] = 0
					trial[i].trialstep[j].combocountonstep[k] = 0
					trial[i].trialstep[j].numofhits[k] = 1
					trial[i].trialstep[j].isthrow[k] = false
					trial[i].trialstep[j].isnohit[k] = false
					trial[i].trialstep[j].ishelper[k] = false
					trial[i].trialstep[j].isproj[k] = false
					trial[i].trialstep[j].iscounterhit[k] = false
					trial[i].trialstep[j].validuntilnexthit[k] = false
				end
			elseif lcline:find("trialstep." .. j .. ".anim") then
				if string.gsub(f_trimafterchar(lcline, "="),"%s+", "") ~= "" then
					trial[i].trialstep[j].anim = f_strtonumber(main.f_strsplit(',', string.gsub(f_trimafterchar(lcline, "="),"%s+", "")))
				end
			elseif lcline:find("trialstep." .. j .. ".numofhits") then
				if string.gsub(f_trimafterchar(lcline, "="),"%s+", "") ~= "" then
					trial[i].trialstep[j].numofhits = f_strtonumber(main.f_strsplit(',', string.gsub(f_trimafterchar(lcline, "="),"%s+", "")))
				end
			elseif lcline:find("trialstep." .. j .. ".isthrow") then
				if string.gsub(f_trimafterchar(lcline, "="),"%s+", "") ~= "" then
					trial[i].trialstep[j].isthrow = f_strtoboolean(main.f_strsplit(',', string.gsub(f_trimafterchar(lcline, "="),"%s+", "")))
				end
			elseif lcline:find("trialstep." .. j .. ".isnohit") then
				if string.gsub(f_trimafterchar(lcline, "="),"%s+", "") ~= "" then
					trial[i].trialstep[j].isnohit = f_strtoboolean(main.f_strsplit(',', string.gsub(f_trimafterchar(lcline, "="),"%s+", "")))
				end
			elseif lcline:find("trialstep." .. j .. ".iscounterhit") then
				if string.gsub(f_trimafterchar(lcline, "="),"%s+", "") ~= "" then
					trial[i].trialstep[j].iscounterhit = f_strtoboolean(main.f_strsplit(',', string.gsub(f_trimafterchar(lcline, "="),"%s+", "")))
				end
			elseif lcline:find("trialstep." .. j .. ".ishelper") then
				if string.gsub(f_trimafterchar(lcline, "="),"%s+", "") ~= "" then
					trial[i].trialstep[j].ishelper = f_strtoboolean(main.f_strsplit(',', string.gsub(f_trimafterchar(lcline, "="),"%s+", "")))
				end
			elseif lcline:find("trialstep." .. j .. ".isproj") then
				if string.gsub(f_trimafterchar(lcline, "="),"%s+", "") ~= "" then
					trial[i].trialstep[j].isproj = f_strtoboolean(main.f_strsplit(',', string.gsub(f_trimafterchar(lcline, "="),"%s+", "")))
				end
			elseif lcline:find("trialstep." .. j .. ".validuntilnexthit") then
				if string.gsub(f_trimafterchar(lcline, "="),"%s+", "") ~= "" then
					trial[i].trialstep[j].validuntilnexthit = f_strtoboolean(main.f_strsplit(',', string.gsub(f_trimafterchar(lcline, "="),"%s+", "")))
				end
			end
		end
		trialsFile:close()

		main.t_selChars[row].trialsdata = trial
	end
end

--;===========================================================
--; global.lua
--;===========================================================
hook.add("loop#trials", "f_trialsMode", start.f_trialsMode)
hook.add("start.f_selectScreen", "f_trialsSelectScreen", start.f_trialsSelectScreen)
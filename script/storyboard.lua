
local storyboard = {}

--http://www.elecbyte.com/mugendocs/storyboard.html

storyboard.t_storyboard = {} --stores all parsed storyboards (we parse each of them only once)

local function f_reset(t)
	for k, v in pairs(t.scene) do
		if t.scene[k].fadein_data ~= nil then
			animReset(t.scene[k].fadein_data)
			animUpdate(t.scene[k].fadein_data)
		end
		if t.scene[k].fadeout_data ~= nil then
			animReset(t.scene[k].fadeout_data)
			animUpdate(t.scene[k].fadeout_data)
		end
		if t.scene[k].bg_name ~= '' then
			local t_bgdef = t[t.scene[k].bg_name .. 'def']
			for i = 1, #t_bgdef do
				t_bgdef[i].ctrl_flags.visible = 1
				t_bgdef[i].ctrl_flags.enabled = 1
				t_bgdef[i].ctrl_flags.velx = 0
				t_bgdef[i].ctrl_flags.vely = 0
				animReset(t.scene[k].bg_data[i])
				animAddPos(t.scene[k].bg_data[i], 0 - t_bgdef[i].ctrl_flags.x, 0 - t_bgdef[i].ctrl_flags.y)
				animUpdate(t.scene[k].bg_data[i])
				t_bgdef[i].ctrl_flags.x = 0
				t_bgdef[i].ctrl_flags.y = 0
				for k2, v2 in pairs(t_bgdef[i].ctrl) do
					for j = 1, #t_bgdef[i].ctrl[k2] do
						t_bgdef[i].ctrl[k2][j].timer[1] = t_bgdef[i].ctrl[k2][j].time[1]
						t_bgdef[i].ctrl[k2][j].timer[2] = t_bgdef[i].ctrl[k2][j].time[2]
						t_bgdef[i].ctrl[k2][j].timer[3] = t_bgdef[i].ctrl[k2][j].time[3]
					end
				end
			end
		end
		for k2, v2 in pairs(t.scene[k].layer) do
			if t.scene[k].layer[k2].anim_data ~= nil then
				animReset(t.scene[k].layer[k2].anim_data)
				animUpdate(t.scene[k].layer[k2].anim_data)
			end
			t.scene[k].layer[k2].text_timer = 0
		end
	end
end

local function f_play(t)
	main.f_printTable(t, 'debug/t_storyboard.txt')
	--loop through scenes in order
	for k, v in main.f_sortKeys(t.scene) do
		--scene >= startscene
		if k >= t.scenedef.startscene then
			for i = 0, t.scene[k].end_time do
				--end storyboard
				if esc() or main.f_btnPalNo(main.p1Cmd) > 0 and t.scenedef.skipbutton > 0 then
					main.f_cmdInput()
					refresh()
					return
				end
				--play bgm
				if i == 0 and t.scene[k].bgm ~= nil then
					playBGM(t.scene[k].bgm, true, t.scene[k].bgm_loop, t.scene[k].bgm_volume, t.scene[k].bgm_loopstart or "0", t.scene[k].bgm_loopend or "0")
				end
				--play snd
				if t.scenedef.snd_data ~= nil then
					for j = 1, #t.scene[k].sound do
						if i == t.scene[k].sound[j].starttime then
							sndPlay(t.scenedef.snd_data, t.scene[k].sound[j].value[1], t.scene[k].sound[j].value[2])
						end
					end
				end
				--clearcolor
				animDraw(t.scene[k].clearcolor_data)
				--draw layerno = 0 backgrounds
				if t.scene[k].bg_name ~= '' then
					main.f_drawBG(t.scene[k].bg_data, t[t.scene[k].bg_name .. 'def'], 0, i, t.info.localcoord)
				end
				--loop through layers in order
				for k2, v2 in main.f_sortKeys(t.scene[k].layer) do
					if i >= t.scene[k].layer[k2].starttime and i <= t.scene[k].layer[k2].endtime then
						--layer anim
						if t.scene[k].layer[k2].anim_data ~= nil then
							animDraw(t.scene[k].layer[k2].anim_data)
							animUpdate(t.scene[k].layer[k2].anim_data)
						end
						--layer text
						if t.scene[k].layer[k2].text_data ~= nil then
							t.scene[k].layer[k2].text_timer = t.scene[k].layer[k2].text_timer + 1
							main.f_textRender(
								t.scene[k].layer[k2].text_data,
								t.scene[k].layer[k2].text,
								t.scene[k].layer[k2].text_timer,
								t.scene[k].layerall_pos[1] + t.scene[k].layer[k2].offset[1],
								t.scene[k].layerall_pos[2] + t.scene[k].layer[k2].offset[2],
								t.scene[k].layer[k2].text_spacing[2],
								t.scene[k].layer[k2].textdelay,
								t.scene[k].layer[k2].text_length
							)
							end
					end
				end
				--draw layerno = 1 backgrounds
				if t.scene[k].bg_name ~= '' then
					main.f_drawBG(t.scene[k].bg_data, t[t.scene[k].bg_name .. 'def'], 1, i, t.info.localcoord)
				end
				--fadein
				if i <= t.scene[k].fadein_time then
					animDraw(t.scene[k].fadein_data)
					animUpdate(t.scene[k].fadein_data)
				end
				--fadeout
				if i >= t.scene[k].end_time - t.scene[k].fadeout_time then
					animDraw(t.scene[k].fadeout_data)
					animUpdate(t.scene[k].fadeout_data)
				end
				if main.f_btnPalNo(main.p1Cmd) > 0 and t.scenedef.skipbutton <= 0 then
					main.f_cmdInput()
					refresh()
					do
						break
					end
				end
				main.f_cmdInput()
				refresh()
			end
		end
	end
	playBGM('', true, 1, 100, "0", "0")
end

local function f_parse(path)
	-- Intro haves his own localcoord function
	-- So we disable it
	main.SetDefaultScale()
	
	local file = io.open(path, 'r')
	local fileDir, fileName = path:match('^(.-)([^/\\]+)$')
	local t = {}
	local pos = t
	local pos_default = {}
	local pos_val = {}
	t.anim = {}
	t.ctrldef = {}
	t.scene = {}
	t.fileDir = fileDir
	t.fileName = fileName
	local bgdef = 'dummyUntilSet'
	local bgctrl = ''
	local bgctrl_match = 'dummyUntilSet'
	local tmp = ''
	local t_default =
	{
		info = {localcoord = {320, 240}},
		scenedef = {spr = '', snd = '', font = {[1] = 'font/f-6x9.fnt'}, font_height = {}, startscene = 0, skipbutton = 1, font_data = {}},
		scene = {},
		ctrldef = {}
	}
	for line in file:lines() do
		line = line:gsub('%s*;.*$', '')
		if line:match('^%s*%[.-%s*%]%s*$') then --matched [] group
			line = line:match('^%s*%[(.-)%s*%]%s*$') --match text between []
			line = line:gsub('[%. ]', '_') --change . and space to _
			line = line:lower() --lowercase line
			local row = tostring(line:lower()) --just in case it's a number (not really needed)
			if row:match('.+ctrldef') then --matched ctrldef start
				bgctrl = row
				bgctrl_match = bgctrl:match('^(.-ctrl)def')
				if t.ctrldef[bgdef .. 'def'][bgctrl] ~= nil then --Ctrldef名の重複を避ける
					bgctrl = bgctrl..tostring(os.clock())
				end
				t.ctrldef[bgdef .. 'def'][bgctrl] = {}
				t.ctrldef[bgdef .. 'def'][bgctrl].ctrl = {}
				pos = t.ctrldef[bgdef .. 'def'][bgctrl]
				t_default.ctrldef[bgdef .. 'def'][bgctrl] = {
					looptime = -1,
					ctrlid = {0},
					ctrl = {}
				}
			elseif row:match('^' .. bgctrl_match) then --matched ctrldef content
				tmp = t.ctrldef[bgdef .. 'def'][bgctrl].ctrl
				tmp[#tmp + 1] = {}
				pos = tmp[#tmp]
				t_default.ctrldef[bgdef .. 'def'][bgctrl].ctrl[#tmp] = {
					type = 'null',
					time = {0, -1, -1},
					ctrlid = {}
				}
			elseif row:match('.+def$') and not row:match('^scenedef$') --[[and not row:match('^' .. bgdef .. '.*$')]] then --matched bgdef start
				t[row] = {}
				pos = t[row]
				bgdef = row:match('(.+)def$')
				t_default[row] = {}
				t.ctrldef[bgdef .. 'def'] = {}
				t_default.ctrldef[bgdef .. 'def'] = {}
			elseif row:match('^' .. bgdef) then --matched bgdef content
				tmp = t[bgdef .. 'def']
				tmp[#tmp + 1] = {}
				pos = tmp[#tmp]
				t_default[bgdef .. 'def'][#tmp] =
				{
					type = 'normal',
					spriteno = {0, 0},
					id = 0,
					layerno = 0,
					start = {0, 0},
					delta = {1, 1},
					trans = '',
					mask = 0,
					tile = {0, 0},
					tilespacing = {0, nil},
					--window = {0, 0, 0, 0},
					--windowdelta = {0, 0}, --not supported yet
					--width = {0, 0}, --not supported yet (parallax)
					--xscale = {1.0, 1.0}, --not supported yet (parallax)
					--yscalestart = 100, --not supported yet (parallax)
					--yscaledelta = 1, --not supported yet (parallax)
					positionlink = 0,
					velocity = {0, 0},
					--sin_x = {0, 0, 0}, --not supported yet
					--sin_y = {0, 0, 0}, --not supported yet
					ctrl = {},
					ctrl_flags = {
						visible = 1,
						enabled = 1,
						velx = 0,
						vely = 0,
						x = 0,
						y = 0
					}
				}
			elseif row:match('^scene_[0-9]+$') then --matched scene
				row = tonumber(row:match('^scene_([0-9]+)$'))
				t.scene[row] = {}
				pos = t.scene[row]
				pos.layer = {}
				pos.sound = {}
				t_default.scene[row] =
				{
					end_time = 0,
					fadein_time = 0,
					fadein_col = {0, 0, 0},
					fadeout_time = 0,
					fadeout_col = {0, 0, 0},
					clearcolor = {},
					layerall_pos = {},
					layer = {},
					sound = {},
					--bgm = '',
					bgm_loop = 0,
					bgm_volume = 100,
					bgm_loopstart = nil,
					bgm_loopend = nil,
					--window = {0, 0, 0, 0},
					bg_name = ''
				}
				pos_default = t_default.scene[row]
			elseif row:match('^begin_action_[0-9]+$') then --matched anim
				row = tonumber(row:match('^begin_action_([0-9]+)$'))
				t.anim[row] = {}
				pos = t.anim[row]
			else --matched other []
				t[row] = {}
				pos = t[row]
			end
		else --matched non [] line
			local param, value = line:match('^%s*([^=]-)%s*=%s*(.-)%s*$')
			if param ~= nil and value ~= nil and not value:match('^%s*$') then --param = value pattern matched
				param = param:gsub('[%. ]', '_') --change param . and space to _
				param = param:lower() --lowercase param
				value = value:gsub('"', '') --remove brackets from value
				value = value:gsub('^(%.[0-9])', '0%1') --add 0 before dot if missing at the beginning of matched string
				value = value:gsub('([^0-9])(%.[0-9])', '%10%2') --add 0 before dot if missing anywhere else
				if param:match('^font[0-9]+$') then --font param matched
					local num = tonumber(param:match('font([0-9]+)'))
					if param:match('_height$') then
						if pos.font_height == nil then
							pos.font_height = {}
						end
						pos.font_height[num] = main.f_dataType(value)
					else
						if pos.font == nil then
							pos.font = {}
						end
						pos.font[num] = tostring(value)
					end
				else
					if param:match('^layer[0-9]+_') then --scene layer param matched
						local num = tonumber(param:match('^layer([0-9]+)_'))
						param = param:gsub('layer[0-9]+_', '')
						if pos.layer[num] == nil then
							pos.layer[num] = {}
							pos_default.layer[num] =
							{
								anim = -1,
								text = '',
								font = {1, 0, 0, nil, nil, nil},
								text_spacing = {0, 15}, --Ikemen feature
								textdelay = 2,
								text_length = 50, --Ikemen feature
								text_timer = 0, --Ikemen feature
								offset = {0, 0},
								starttime = 0,
								--endtime = 0
							}
						end
						pos_val = pos.layer[num]
					elseif param:match('^sound[0-9]+_') then --sound param matched
						local num = tonumber(param:match('^sound([0-9]+)_'))
						param = param:gsub('sound[0-9]+_', '')
						if pos.sound[num] == nil then
							pos.sound[num] = {}
							pos_default.sound[num] =
							{
								value = {-1, -1},
								starttime = 0,
								volumescale = 100, --not supported yet
								pan = 0 --not supported yet
							}
						end
						pos_val = pos.sound[num]
					else
						pos_val = pos
					end
					if pos_val[param] == nil then --mugen takes into account only first occurrence
						if value:match('.+,.+') then --multiple values
							for i, c in ipairs(main.f_strsplit(',', value)) do --split value using "," delimiter
								if pos_val[param] == nil then
									pos_val[param] = {}
								end
								if c == '' then
									pos_val[param][#pos_val[param] + 1] = 0
								else
									pos_val[param][#pos_val[param] + 1] = main.f_dataType(c)
								end
							end
						else --single value
							pos_val[param] = main.f_dataType(value)
						end
					end
				end
			else --only valid lines left are animations
				line = line:lower()
				local value = line:match('^%s*([0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+.-)[,%s]*$') or line:match('^%s*loopstart') or line:match('^%s*interpolate offset') or line:match('^%s*interpolate angle') or line:match('^%s*interpolate scale') or line:match('^%s*interpolate blend')
				if value ~= nil then
					value = value:gsub(',%s*,', ',0,') --add missing values
					value = value:gsub(',%s*$', '')
					pos[#pos + 1] = value
				end
			end
		end
	end
	file:close()
	--;===========================================================
	--; FIX REFERENCES, LOAD DATA
	--;===========================================================
	--merge tables
	t = main.f_tableMerge(t_default, t)
	--ctrldef table adjustment
	for k, v in pairs(t.ctrldef) do
		for k2, v2 in pairs(t.ctrldef[k]) do
			tmp = t.ctrldef[k][k2].ctrl
			for i = 1, #tmp do
				--if END_TIME is omitted it should default to the same value as START_TIME
				if tmp[i].time[2] == -1 then
					tmp[i].time[2] = tmp[i].time[1]
				end
				--if LOOPTIME is omitted or set to -1, the background controller will not reset its own timer. In such case use GLOBAL_LOOPTIME
				if tmp[i].time[3] == -1 then
					tmp[i].time[3] = t.ctrldef[k][k2].looptime
				end
				--lowercase type name
				tmp[i].type = tmp[i].type:lower()
				--this list, if specified, overrides the default list specified in the BGCtrlDef
				if #tmp[i].ctrlid == 0 then
					for j = 1, #t.ctrldef[k][k2].ctrlid do
						tmp[i].ctrlid[#tmp[i].ctrlid + 1] = t.ctrldef[k][k2].ctrlid[j]
					end
				end
			end
		end
	end
	--scenedef spr
	if not t.scenedef.spr:match('^data/') then
		if main.f_fileExists(t.fileDir .. t.scenedef.spr) then
			t.scenedef.spr = t.fileDir .. t.scenedef.spr
		elseif main.f_fileExists('data/' .. t.scenedef.spr) then
			t.scenedef.spr = 'data/' .. t.scenedef.spr
		end
	end
	t.scenedef.spr_data = sffNew(t.scenedef.spr)
	--scenedef snd
	if t.scenedef.snd ~= '' then
		if not t.scenedef.snd:match('^data/') then
			if main.f_fileExists(t.fileDir .. t.scenedef.snd) then
				t.scenedef.snd = t.fileDir .. t.scenedef.snd
			elseif main.f_fileExists('data/' .. t.scenedef.snd) then
				t.scenedef.snd = 'data/' .. t.scenedef.snd
			end
		end
		t.scenedef.snd_data = sndNew(t.scenedef.snd)
	end
	--scenedef fonts
	for k, v in pairs(t.scenedef.font) do --loop through table keys
		if t.scenedef.font[k] ~= '' then
			if not t.scenedef.font[k]:match('^data/') then
				if main.f_fileExists(t.fileDir .. t.scenedef.font[k]) then
					t.scenedef.font[k] = t.fileDir .. t.scenedef.font[k]
				elseif main.f_fileExists('font/' .. t.scenedef.font[k]) then
					t.scenedef.font[k] = 'font/' .. t.scenedef.font[k]
				end
				t.scenedef.font_data[k] = fontNew(t.scenedef.font[k])
				t.scenedef.font[k] = {}
				t.scenedef.font[k][1] = k
				t.scenedef.font[k][2] = 0
				t.scenedef.font[k][3] = 0
			end
		end
	end
	--loop through scenes
	local prev_k = ''
	for k, v in main.f_sortKeys(t.scene) do
		--bgm
		if t.scene[k].bgm ~= nil and not t.scene[k].bgm:match('^data/') then
			if main.f_fileExists(t.fileDir .. t.scene[k].bgm) then
				t.scene[k].bgm = t.fileDir .. t.scene[k].bgm
			elseif main.f_fileExists('music/' .. t.scene[k].bgm) then
				t.scene[k].bgm = 'music/' .. t.scene[k].bgm
			end
		end
		--default values
		if #t.scene[k].clearcolor == 0 then
			if prev_k ~= '' and #t.scene[prev_k].clearcolor > 0 then
				t.scene[k].clearcolor[1], t.scene[k].clearcolor[2], t.scene[k].clearcolor[3] = t.scene[prev_k].clearcolor[1], t.scene[prev_k].clearcolor[2], t.scene[prev_k].clearcolor[3]
			else
				t.scene[k].clearcolor[1], t.scene[k].clearcolor[2], t.scene[k].clearcolor[3] = 0, 0, 0
			end
		end
		if #t.scene[k].layerall_pos == 0 then
			if prev_k ~= '' and #t.scene[prev_k].layerall_pos > 0 then
				t.scene[k].layerall_pos[1], t.scene[k].layerall_pos[2] = t.scene[prev_k].layerall_pos[1], t.scene[prev_k].layerall_pos[2]
			else
				t.scene[k].layerall_pos[1], t.scene[k].layerall_pos[2] = 0, 0
			end
		end
		prev_k = k
		--backgrounds data
		local anim = ''
		if t.scene[k].bg_name ~= '' then
			t.scene[k].bg_data = {}
			t.scene[k].bg_name = t.scene[k].bg_name:lower()
			local t_bgdef = t[t.scene[k].bg_name .. 'def']
			local prev_k2 = ''
			for k2, v2 in pairs(t_bgdef) do --loop through table keys
				if type(k2) == "number" and t_bgdef[k2].type ~= nil then
					t_bgdef[k2].type = t_bgdef[k2].type:lower()
					--mugen ignores delta = 0 (defaults to 1)
					if t_bgdef[k2].delta[1] == 0 then t_bgdef[k2].delta[1] = 1 end
					if t_bgdef[k2].delta[2] == 0 then t_bgdef[k2].delta[2] = 1 end
					--add ctrl data
					t[t.scene[k].bg_name .. 'def'][k2].ctrl = main.f_ctrlBG(t_bgdef[k2], t.ctrldef[t.scene[k].bg_name .. 'def'])
					--positionlink adjustment
					if t_bgdef[k2].positionlink == 1 and prev_k2 ~= '' then
						t_bgdef[k2].start[1] = t_bgdef[prev_k2].start[1]
						t_bgdef[k2].start[2] = t_bgdef[prev_k2].start[2]
						t_bgdef[k2].delta[1] = t_bgdef[prev_k2].delta[1]
						t_bgdef[k2].delta[2] = t_bgdef[prev_k2].delta[2]
					end
					prev_k2 = k2
					--generate anim data
					local sizeX, sizeY, offsetX, offsetY = 0, 0, 0, 0
					if t_bgdef[k2].type == 'anim' then
						anim = main.f_animFromTable(t.anim[t_bgdef[k2].actionno], t.scenedef.spr_data, t_bgdef[k2].start[1], t_bgdef[k2].start[2])
					else --normal, parallax
						anim = t_bgdef[k2].spriteno[1] .. ', ' .. t_bgdef[k2].spriteno[2] .. ', ' .. t_bgdef[k2].start[1] .. ', ' .. t_bgdef[k2].start[2] .. ', ' .. -1
						anim = animNew(t.scenedef.spr_data, anim)
						sizeX, sizeY, offsetX, offsetY = getSpriteInfo(t.scenedef.spr, t_bgdef[k2].spriteno[1], t_bgdef[k2].spriteno[2])
					end
					if t_bgdef[k2].trans == 'add1' then
						animSetAlpha(anim, 255, 128)
					elseif t_bgdef[k2].trans == 'add' then
						animSetAlpha(anim, 255, 255)
					elseif t_bgdef[k2].trans == 'sub' then
						animSetAlpha(anim, 1, 255)
					end
					animAddPos(anim, 160, 0) --for some reason needed in ikemen
					if t_bgdef[k2].window ~= nil then
						animSetWindow(
							anim,
							t_bgdef[k2].window[1] * 320/t.info.localcoord[1],
							t_bgdef[k2].window[2] * 240/t.info.localcoord[2],
							(t_bgdef[k2].window[3] - t_bgdef[k2].window[1] + 1)* 320/t.info.localcoord[1],
							(t_bgdef[k2].window[4] - t_bgdef[k2].window[2] + 1) * 240/t.info.localcoord[2]
						)
					else
						animSetWindow(anim, 0, 0, t.info.localcoord[1], t.info.localcoord[2])
					end
					if t_bgdef[k2].tilespacing[2] == nil then t_bgdef[k2].tilespacing[2] = t_bgdef[k2].tilespacing[1] end
					if t_bgdef[k2].type == 'parallax' then
						animSetTile(anim, t_bgdef[k2].tile[1], 0, t_bgdef[k2].tilespacing[1] + sizeX, t_bgdef[k2].tilespacing[2] + sizeY)
					else
						animSetTile(anim, t_bgdef[k2].tile[1], t_bgdef[k2].tile[2], t_bgdef[k2].tilespacing[1] + sizeX, t_bgdef[k2].tilespacing[2] + sizeY)
					end
					animSetScale(anim, 320/t.info.localcoord[1], 240/t.info.localcoord[2])
					if t_bgdef[k2].mask == 1 or t_bgdef[k2].type ~= 'normal' or (t_bgdef[k2].trans ~= '' and t_bgdef[k2].trans ~= 'none') then
						animSetColorKey(anim, 0)
					else
						animSetColorKey(anim, -1)
					end
					--animUpdate(anim)
					t.scene[k].bg_data[k2] = anim
				end
			end
		end
		--clearcolor data
		t.scene[k].clearcolor_data = main.f_clearColor(t.scene[k].clearcolor[1], t.scene[k].clearcolor[2], t.scene[k].clearcolor[3])
		animSetWindow(t.scene[k].clearcolor_data, 0, 0, t.info.localcoord[1], t.info.localcoord[2])
		--fadein data
		t.scene[k].fadein_data = main.f_fadeAnim(1, t.scene[k].fadein_time, t.scene[k].fadein_col[1], t.scene[k].fadein_col[2], t.scene[k].fadein_col[3])
		animSetWindow(t.scene[k].fadein_data, 0, 0, t.info.localcoord[1], t.info.localcoord[2])
		--fadeout data
		t.scene[k].fadeout_data = main.f_fadeAnim(0, t.scene[k].fadeout_time, t.scene[k].fadeout_col[1], t.scene[k].fadeout_col[2], t.scene[k].fadeout_col[3])
		animSetWindow(t.scene[k].fadeout_data, 0, 0, t.info.localcoord[1], t.info.localcoord[2])
		--loop through scene layers
		local t_layer = t.scene[k].layer
		for k2, v2 in pairs(t_layer) do
			--anim
			if t_layer[k2].anim ~= -1 and t.anim[t_layer[k2].anim] ~= nil then
				t.scene[k].layer[k2].anim_data = main.f_animFromTable(
					t.anim[t_layer[k2].anim],
					t.scenedef.spr_data,
					t.scene[k].layerall_pos[1] + t_layer[k2].offset[1],
					t.scene[k].layerall_pos[2] + t_layer[k2].offset[2]
				)
				animSetScale(t.scene[k].layer[k2].anim_data, 320/t.info.localcoord[1], 240/t.info.localcoord[2])
			end
			--text
			if t_layer[k2].text ~= '' then
				t.scene[k].layer[k2].text_data = main.f_createTextImg(
					t.scenedef.font_data[t.scenedef.font[t_layer[k2].font[1]][1]],
					t_layer[k2].font[2],
					t_layer[k2].font[3],
					t_layer[k2].text,
					t.scene[k].layerall_pos[1] + t_layer[k2].offset[1],
					t.scene[k].layerall_pos[2] + t_layer[k2].offset[2],
					320/t.info.localcoord[1],
					240/t.info.localcoord[2],
					t_layer[k2].font[4],
					t_layer[k2].font[5],
					t_layer[k2].font[6]
				)
			end
			--endtime
			if t_layer[k2].endtime == nil then
				t_layer[k2].endtime = t.scene[k].end_time
			end
		end
	end
	--t.ctrldef = nil
	
	-- Finished loading intro
	-- Re-enabled custom scaling
	main.SetScaleValues()
	
	return t
end

function storyboard.f_storyboard(path)
	path = path:gsub('\\', '/')
	if storyboard.t_storyboard[path] == nil then
		storyboard.t_storyboard[path] = f_parse(path)
	else
		f_reset(storyboard.t_storyboard[path])
	end
	f_play(storyboard.t_storyboard[path])
end

return storyboard

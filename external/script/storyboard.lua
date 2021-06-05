local storyboard = {}

--http://www.elecbyte.com/mugendocs/storyboard.html

storyboard.t_storyboard = {} --stores all parsed storyboards (we parse each of them only once)

local function f_reset(t)
	main.f_setStoryboardScale(t.info.localcoord)
	for k, v in pairs(t.scene) do
		if t.scene[k].bg_name ~= '' then
			bgReset(t.scene[k].bg)
		end
		for k2, v2 in pairs(t.scene[k].layer) do
			if t.scene[k].layer[k2].anim_data ~= nil then
				animReset(t.scene[k].layer[k2].anim_data)
				animUpdate(t.scene[k].layer[k2].anim_data)
				animSetPalFX(t.scene[k].layer[k2].anim_data, {
					time = t.scene[k].layer[k2].palfx_time,
					add = t.scene[k].layer[k2].palfx_add,
					mul = t.scene[k].layer[k2].palfx_mul,
					sinadd = t.scene[k].layer[k2].palfx_sinadd,
					invertall = t.scene[k].layer[k2].palfx_invertall,
					color = t.scene[k].layer[k2].palfx_color
				})
			end
			t.scene[k].layer[k2].counter = 0
		end
	end
end

local function f_play(t, attract)
	playBGM('')
	if main.debugLog then main.f_printTable(t, 'debug/t_storyboard.txt') end
	--loop through scenes in order
	for k, v in ipairs(t.sceneOrder) do
		if k >= t.scenedef.startscene then
			local fadeType = 'fadein'
			local fadeStart = getFrameCount()
			for i = 0, t.scene[v].end_time do
				--end storyboard
				if esc() or (attract and main.credits > 0) or (not attract and main.f_input(main.t_players, {'pal', 's', 'm'})) and t.scenedef.skipbutton > 0 then
					return
				end
				--credits
				if getKey(motif.attract_mode.credits_key) then
					sndPlay(motif.files.snd_data, motif.attract_mode.credits_snd[1], motif.attract_mode.credits_snd[2])
					main.credits = main.credits + 1
					resetKey()
				end
				--play bgm
				if i == 0 and t.scene[v].bgm ~= nil then
					playBGM(t.scene[v].bgm, true, t.scene[v].bgm_loop, t.scene[v].bgm_volume, t.scene[v].bgm_loopstart, t.scene[v].bgm_loopend)
				end
				--play snd
				if t.scenedef.snd_data ~= nil then
					for k2, v2 in main.f_sortKeys(t.scene[v].sound) do
						if i == v2.starttime then
							sndPlay(t.scenedef.snd_data, v2.value[1], v2.value[2], v2.volumescale, v2.pan)
						end
					end
				end
				--draw clearcolor
				clearColor(t.scene[v].clearcolor[1], t.scene[v].clearcolor[2], t.scene[v].clearcolor[3])
				--draw layerno = 0 backgrounds
				if t.scene[v].bg_name ~= '' then
					bgDraw(t.scene[v].bg, false)
				end
				--loop through layers in order
				for k2, v2 in main.f_sortKeys(t.scene[v].layer) do
					if i >= t.scene[v].layer[k2].starttime and i <= t.scene[v].layer[k2].endtime then
						--layer anim
						if t.scene[v].layer[k2].anim_data ~= nil then
							animDraw(t.scene[v].layer[k2].anim_data)
							animUpdate(t.scene[v].layer[k2].anim_data)
						end
						--layer text
						if t.scene[v].layer[k2].text_data ~= nil then
							t.scene[v].layer[k2].counter = t.scene[v].layer[k2].counter + 1
							main.f_textRender(
								t.scene[v].layer[k2].text_data,
								t.scene[v].layer[k2].text,
								t.scene[v].layer[k2].counter,
								t.scene[v].layerall_pos[1] + t.scene[v].layer[k2].offset[1],
								t.scene[v].layerall_pos[2] + t.scene[v].layer[k2].offset[2],
								t.scene[v].layer[k2].spacing[1],
								t.scene[v].layer[k2].spacing[2],
								main.font_def[t.scene[v].layer[k2].font[1] .. t.scene[v].layer[k2].font_height],
								t.scene[v].layer[k2].textdelay,
								main.f_lineLength(
									t.scene[v].layerall_pos[1] + t.scene[v].layer[k2].offset[1],
									t.info.localcoord[1],
									t.scene[v].layer[k2].font[3],
									t.scene[v].layer[k2].textwindow,
									true
								)
							)
						end
					end
				end
				--draw layerno = 1 backgrounds
				if t.scene[v].bg_name ~= '' then
					bgDraw(t.scene[v].bg, true)
				end
				--draw fadein / fadeout
				if i == t.scene[v].end_time - t.scene[v].fadeout_time then
					fadeType = 'fadeout'
					fadeStart = getFrameCount()
				end
				main.fadeActive = fadeColor(
					fadeType,
					fadeStart,
					t.scene[v][fadeType .. '_time'],
					t.scene[v][fadeType .. '_col'][1],
					t.scene[v][fadeType .. '_col'][2],
					t.scene[v][fadeType .. '_col'][3]
				)
				main.f_cmdInput()
				refresh()
			end
		end
	end
end

local function f_parse(path)
	local file = io.open(path, 'r')
	local fileDir, fileName = path:match('^(.-)([^/\\]+)$')
	local t = {}
	local pos = t
	local pos_default = {}
	local pos_val = {}
	t.sceneOrder = {}
	t.anim = {}
	t.scene = {}
	t.def = fileDir .. fileName
	t.fileDir = fileDir
	t.fileName = fileName
	local tmp = ''
	local t_default =
	{
		info = {localcoord = {320, 240}},
		scenedef = {
			spr = '',
			snd = '',
			font = {},
			font_height = {},
			startscene = 0,
			skipbutton = 1, --Ikemen feature
		},
		scene = {},
	}
	for line in file:lines() do
		line = line:gsub('%s*;.*$', '')
		if line:match('^%s*%[.-%s*%]%s*$') then --matched [] group
			line = line:match('^%s*%[(.-)%s*%]%s*$') --match text between []
			line = line:gsub('[%. ]', '_') --change . and space to _
			local row = tostring(line:lower())
			if row:match('^scene$') or row:match('^scene_') then --matched scene
				table.insert(t.sceneOrder, row)
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
					bgm_volume = 100,  --Ikemen feature
					bgm_loopstart = 0, --Ikemen feature
					bgm_loopend = 0, --Ikemen feature
					--window = {0, 0, 0, 0},
					bg_name = ''
				}
				pos_default = t_default.scene[row]
			elseif row:match('^begin_action_[0-9]+$') then --matched anim
				row = tonumber(row:match('^begin_action_([0-9]+)$'))
				t.anim[row] = {}
				pos = t.anim[row]
			else --matched other []
				if t[row] == nil then
					t[row] = {}
				end
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
				if param:match('^font[0-9]+') then --font declaration param matched
					if pos.font == nil then
						pos.font = {}
						pos.font_height = {}
					end
					local num = tonumber(param:match('font([0-9]+)'))
					if param:match('_height$') then
						pos.font_height[num] = main.f_dataType(value)
					else
						value = value:gsub('\\', '/')
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
								font = {'f-6x9.def', 0, 0, 255, 255, 255, 255, 0},
								font_scale = {1.0, 1.0}, --Ikemen feature
								font_height = -1, --Ikemen feature
								palfx_time = -1, --Ikemen feature
								palfx_add = {0, 0, 0}, --Ikemen feature
								palfx_mul = {256, 256, 256}, --Ikemen feature
								palfx_sinadd = {0, 0, 0}, --Ikemen feature
								palfx_invertall = 0, --Ikemen feature
								palfx_color = 256, --Ikemen feature
								textdelay = 2,
								textwindow = nil, --Ikemen feature
								offset = {0, 0},
								spacing = {0, 0}, --Ikemen feature
								starttime = 0,
								--endtime = 0,
								counter = 0, --used internally by main.f_textRender
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
								volumescale = 100,
								pan = 0 --TODO: not implemented yet
							}
						end
						pos_val = pos.sound[num]
					else
						pos_val = pos
					end
					if pos_val[param] == nil or param:match('_font_height$') then --mugen takes into account only first occurrence
						if param:match('^font$') then --assign default font values if needed (also ensure that there are multiple values in the first place)
							local _, n = value:gsub(',%s*[0-9]*', '')
							for i = n + 1, #main.t_fntDefault do
								value = value:gsub(',?%s*$', ',' .. main.t_fntDefault[i])
							end
						end
						if param:match('^text$') then --skip commas detection for strings
							pos_val[param] = value
						elseif value:match('.+,.+') then --multiple values
							for i, c in ipairs(main.f_strsplit(',', value)) do --split value using "," delimiter
								if i == 1 then
									--t_layer[k2].font
									pos_val[param] = {}
									if param:match('^font$') and tonumber(c) ~= -1 then
										if t.scenedef ~= nil and t.scenedef.font ~= nil and t.scenedef.font[tonumber(c)] ~= nil then
											if pos_val[param .. '_height'] == nil and t.scenedef.font_height[tonumber(c)] ~= nil then
												pos_val[param .. '_height'] = t.scenedef.font_height[tonumber(c)]
											end
											c = t.scenedef.font[tonumber(c)]
										else
											break --use default font values
										end
									end
								end
								if c == nil or c == '' then
									table.insert(pos_val[param], 0)
								else
									table.insert(pos_val[param], main.f_dataType(c))
								end
							end
						else --single value
							pos_val[param] = main.f_dataType(value)
						end
					end
				end
			else --only valid lines left are animations
				line = line:lower()
				local value = line:match('^%s*([0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+%s*,%s*[0-9%-]+.-)[,%s]*$') or line:match('^%s*loopstart') or line:match('^%s*interpolate [oasb][fncl][fgae][sln][ed]t?')
				if value ~= nil then
					value = value:gsub(',%s*,', ',0,') --add missing values
					value = value:gsub(',%s*$', '')
					table.insert(pos, value)
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
	--localcoord
	main.f_setStoryboardScale(t.info.localcoord)
	--scenedef spr
	t.scenedef.spr = main.f_filePath(t.scenedef.spr, t.fileDir, 'data/')
	t.spr_data = {[t.scenedef.spr] = sffNew(t.scenedef.spr)}
	--scenedef snd
	if t.scenedef.snd ~= '' then
		t.scenedef.snd = main.f_filePath(t.scenedef.snd, t.fileDir, 'data/')
		t.scenedef.snd_data = sndNew(t.scenedef.snd)
	end
	--loop through scenes
	local prev_k = ''
	for k, v in main.f_sortKeys(t.scene) do
		--bgm
		if t.scene[k].bgm ~= nil then
			t.scene[k].bgm = main.f_filePath(t.scene[k].bgm, t.fileDir, 'music/')
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
		--backgrounds
		if t.scene[k].bg_name ~= '' then
			local spr_def = t.scene[k].bg_name .. 'def'
			if t[spr_def] ~= nil and t[spr_def].spr ~= nil then --custom spr associated with bg.name is declared
				t[spr_def].spr = main.f_filePath(t[spr_def].spr, t.fileDir, 'data/')
				if t.spr_data[t[spr_def].spr] == nil then --sff data not created yet
					t.spr_data[t[spr_def].spr] = sffNew(t[spr_def].spr)
				end
				t.scene[k].bg = bgNew(t.spr_data[t[spr_def].spr], t.def, t.scene[k].bg_name:lower())
			else
				t.scene[k].bg = bgNew(t.spr_data[t.scenedef.spr], t.def, t.scene[k].bg_name:lower())
			end
			bgReset(t.scene[k].bg)
		end
		--loop through scene layers
		local t_layer = t.scene[k].layer
		for k2, v2 in pairs(t_layer) do
			--anim
			if t_layer[k2].anim ~= -1 and t.anim[t_layer[k2].anim] ~= nil then
				t.scene[k].layer[k2].anim_data = main.f_animFromTable(
					t.anim[t_layer[k2].anim],
					t.spr_data[t.scenedef.spr],
					t.scene[k].layerall_pos[1] + t_layer[k2].offset[1] + main.storyboardOffsetX,
					t.scene[k].layerall_pos[2] + t_layer[k2].offset[2]
				)
				--palfx
				animSetPalFX(t.scene[k].layer[k2].anim_data, {
					time = t.scene[k].layer[k2].palfx_time,
					add = t.scene[k].layer[k2].palfx_add,
					mul = t.scene[k].layer[k2].palfx_mul,
					sinadd = t.scene[k].layer[k2].palfx_sinadd,
					invertall = t.scene[k].layer[k2].palfx_invertall,
					color = t.scene[k].layer[k2].palfx_color
				})
			end
			--text
			if t_layer[k2].text ~= '' then
				t.scene[k].layer[k2].text_data = text:create({
					font =   t_layer[k2].font[1],
					bank =   t_layer[k2].font[2],
					align =  t_layer[k2].font[3],
					text =   t_layer[k2].text,
					x =      t.scene[k].layerall_pos[1] + t_layer[k2].offset[1],
					y =      t.scene[k].layerall_pos[2] + t_layer[k2].offset[2],
					scaleX = t_layer[k2].font_scale[1],
					scaleY = t_layer[k2].font_scale[2],
					r =      t_layer[k2].font[4],
					g =      t_layer[k2].font[5],
					b =      t_layer[k2].font[6],
					height = t_layer[k2].font_height,
					window = t_layer[k2].textwindow,
				})
			end
			--endtime
			if t_layer[k2].endtime == nil then
				t_layer[k2].endtime = t.scene[k].end_time
			end
		end
	end
	return t
end

function storyboard.f_storyboard(path, attract)
	path = path:gsub('\\', '/')
	if not main.f_fileExists(path) then
		return
	end
	main.f_cmdBufReset()
	main.f_disableLuaScale()
	if storyboard.t_storyboard[path] == nil then
		storyboard.t_storyboard[path] = f_parse(path)
	else
		f_reset(storyboard.t_storyboard[path])
	end
	f_play(storyboard.t_storyboard[path], attract or false)
	main.f_cmdBufReset()
	main.f_setLuaScale()
	if attract and main.credits > 0 then
		return true
	end
end

return storyboard

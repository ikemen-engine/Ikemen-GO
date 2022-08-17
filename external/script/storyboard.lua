local storyboard = {}

--http://www.elecbyte.com/mugendocs/storyboard.html

storyboard.t_storyboard = {} --stores all parsed storyboards (we parse each of them only once)

local function f_reset(t)
	main.f_setStoryboardScale(t.info.localcoord)
	for _, scene in pairs(t.scene) do
		if scene.bg_name ~= '' then
			bgReset(scene.bg)
		end
		for _, layer in pairs(scene.layer) do
			if layer.anim_data ~= nil then
				animReset(layer.anim_data)
				animUpdate(layer.anim_data)
				animSetPalFX(layer.anim_data, {
					time =      layer.palfx_time,
					add =       layer.palfx_add,
					mul =       layer.palfx_mul,
					sinadd =    layer.palfx_sinadd,
					invertall = layer.palfx_invertall,
					color =     layer.palfx_color
				})
			end
		end
	end
end

local function f_play(t, attract)
	if main.debugLog then main.f_printTable(t, 'debug/t_storyboard.txt') end
	--loop through scenes in order
	for k, v in ipairs(t.sceneOrder) do
		if k >= t.scenedef.startscene then
			local scene = t.scene[v]
			local fadeType = 'fadein'
			local fadeStart = getFrameCount()
			for i = 0, scene.end_time do
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
				if i == 0 and (k - 1 == t.scenedef.startscene or scene.bgm ~= '') then
					main.f_playBGM(k - 1 == t.scenedef.startscene, scene.bgm, scene.bgm_loop, scene.bgm_volume, scene.bgm_loopstart, scene.bgm_loopend)
				end
				--play snd
				if t.scenedef.snd_data ~= nil then
					for _, sound in main.f_sortKeys(scene.sound) do
						if i == sound.starttime then
							sndPlay(t.scenedef.snd_data, sound.value[1], sound.value[2], sound.volumescale, sound.pan)
						end
					end
				end
				--draw clearcolor
				clearColor(scene.clearcolor[1], scene.clearcolor[2], scene.clearcolor[3])
				--draw layerno = 0 backgrounds
				if scene.bg_name ~= '' then
					bgDraw(scene.bg, false)
				end
				--loop through layers in order
				for _, layer in main.f_sortKeys(scene.layer) do
					if i >= layer.starttime and i <= layer.endtime then
						--layer anim
						if layer.anim_data ~= nil then
							animDraw(layer.anim_data)
							animUpdate(layer.anim_data)
						end
						--layer text
						if layer.text_data ~= nil then
							local counter = i - layer.starttime
							main.f_textRender(
								layer.text_data,
								layer.text,
								counter + 1,
								scene.layerall_pos[1] + layer.offset[1] + layer.vel[1] * counter,
								scene.layerall_pos[2] + layer.offset[2] + layer.vel[2] * counter,
								layer.spacing[1],
								layer.spacing[2],
								main.font_def[layer.font[1] .. layer.font[7]],
								layer.textdelay,
								main.f_lineLength(
									scene.layerall_pos[1] + layer.offset[1] + layer.vel[1] * counter,
									t.info.localcoord[1],
									layer.font[3],
									layer.textwindow,
									true
								)
							)
						end
					end
				end
				--draw layerno = 1 backgrounds
				if scene.bg_name ~= '' then
					bgDraw(scene.bg, true)
				end
				--draw fadein / fadeout
				if i == scene.end_time - scene.fadeout_time then
					fadeType = 'fadeout'
					fadeStart = getFrameCount()
				end
				main.fadeActive = fadeColor(
					fadeType,
					fadeStart,
					scene[fadeType .. '_time'],
					scene[fadeType .. '_col'][1],
					scene[fadeType .. '_col'][2],
					scene[fadeType .. '_col'][3]
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
	local t = {info = {localcoord = {320, 240}}}
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
				if not t.scene[row] then --mugen skips duplicated scenes
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
						bgm = '',
						bgm_loop = 0,
						bgm_volume = 100,  --Ikemen feature
						bgm_loopstart = 0, --Ikemen feature
						bgm_loopend = 0, --Ikemen feature
						bg_name = ''
					}
					pos_default = t_default.scene[row]
				end
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
				value = value:gsub(',%s*$', '') --remove dummy ','
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
								font = {'f-6x9.def', 0, 0, 255, 255, 255, -1},
								scale = {1.0, 1.0}, --Ikemen feature
								palfx_time = -1, --Ikemen feature
								palfx_add = {0, 0, 0}, --Ikemen feature
								palfx_mul = {256, 256, 256}, --Ikemen feature
								palfx_sinadd = {0, 0, 0}, --Ikemen feature
								palfx_invertall = 0, --Ikemen feature
								palfx_color = 256, --Ikemen feature
								textdelay = 2,
								textwindow = {0, 0, math.max(config.GameWidth, t.info.localcoord[1]), math.max(config.GameHeight, t.info.localcoord[2])}, --Ikemen feature
								offset = {0, 0},
								vel = {0, 0}, --Ikemen feature
								spacing = {0, 0}, --Ikemen feature
								starttime = 0,
								--endtime = 0,
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
								pan = 0,
							}
						end
						pos_val = pos.sound[num]
					else
						pos_val = pos
					end
					if pos_val[param] == nil or param == 'localcoord' then --mugen takes into account only first occurrence
						if param:match('^font$') then --assign default font values if needed (also ensure that there are multiple values in the first place)
							local _, n = value:gsub(',', '')
							for i = n + 1, #main.t_fntDefault do
								value = value:gsub(',?%s*$', ',' .. main.t_fntDefault[i])
							end
						end
						if param:match('^text$') then --skip commas detection for strings
							pos_val[param] = value
						elseif value:match('.+,.+') then --multiple values
							local fontRef = -1
							for i, c in ipairs(main.f_strsplit(',', value)) do --split value using "," delimiter
								if param:match('_anim$') then --mugen recognizes animations even if there are more values
									pos_val[param] = main.f_dataType(c)
									break
								else
									if i == 1 then
										pos_val[param] = {}
									end
									if param:match('^font$') then
										-- Change font number reference to font string
										if i == 1 then
											if t.scenedef ~= nil and t.scenedef.font ~= nil and t.scenedef.font[tonumber(c)] ~= nil then
												fontRef = tonumber(c)
												c = t.scenedef.font[fontRef]
											end
										-- Assign default ttf font height, if custom value is not set
										elseif i == 7 and tonumber(c) == -1 and t.scenedef ~= nil and t.scenedef.font_height ~= nil and t.scenedef.font_height[fontRef] ~= nil then
											c = tostring(t.scenedef.font_height[fontRef])
										-- Otherwise validate data
										elseif not tonumber(c) then
											c = nil
										end
									end
								end
								-- Append values
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
	t.scenedef.spr = searchFile(t.scenedef.spr, {t.fileDir})
	if not main.f_fileExists(t.scenedef.spr) then
		print("failed to load " .. path .. " (storyboard): SFF file not found: " .. t.scenedef.spr)
		return nil
	end
	t.spr_data = {[t.scenedef.spr] = sffNew(t.scenedef.spr)}
	--scenedef snd
	if t.scenedef.snd ~= '' then
		t.scenedef.snd = searchFile(t.scenedef.snd, {t.fileDir})
		if not main.f_fileExists(t.scenedef.snd) then
			print("failed to load " .. path .. " (storyboard): SND file not found: " .. t.scenedef.snd)
		end
		t.scenedef.snd_data = sndNew(t.scenedef.snd)
	end
	--loop through scenes
	local prev_s = nil
	for _, scene in main.f_sortKeys(t.scene) do
		--bgm
		if scene.bgm ~= nil then
			scene.bgm = searchFile(scene.bgm, {t.fileDir, 'sound/'})
		end
		--default values
		if #scene.clearcolor == 0 then
			local r, g, b = 0, 0, 0
			if prev_s ~= nil and #prev_s.clearcolor > 0 then
				r, g, b = prev_s.clearcolor[1], prev_s.clearcolor[2], prev_s.clearcolor[3]
			end
			scene.clearcolor[1], scene.clearcolor[2], scene.clearcolor[3] = r, g, b
		end
		if #scene.layerall_pos == 0 then
			local x, y = 0, 0
			if prev_s ~= nil and #prev_s.layerall_pos > 0 then
				x, y = prev_s.layerall_pos[1], prev_s.layerall_pos[2]
			end
			scene.layerall_pos[1], scene.layerall_pos[2] = x, y
		end
		prev_s = scene
		--backgrounds
		if scene.bg_name ~= '' then
			local spr_def = scene.bg_name .. 'def'
			if t[spr_def] ~= nil and t[spr_def].spr ~= nil then --custom spr associated with bg.name is declared
				t[spr_def].spr = searchFile(t[spr_def].spr, {t.fileDir})
				if not main.f_fileExists(t[spr_def].spr) then
					print("failed to load " .. path .. " (storyboard): SFF file not found: " .. t[spr_def].spr)
				end
				if t.spr_data[t[spr_def].spr] == nil then --sff data not created yet
					t.spr_data[t[spr_def].spr] = sffNew(t[spr_def].spr)
				end
				scene.bg = bgNew(t.spr_data[t[spr_def].spr], t.def, scene.bg_name:lower())
			else
				scene.bg = bgNew(t.spr_data[t.scenedef.spr], t.def, scene.bg_name:lower())
			end
			bgReset(scene.bg)
		end
		--loop through scene layers
		for _, layer in pairs(scene.layer) do
			--anim
			if layer.anim ~= -1 and t.anim[layer.anim] ~= nil then
				layer.anim_data = main.f_animFromTable(
					t.anim[layer.anim],
					t.spr_data[t.scenedef.spr],
					scene.layerall_pos[1] + layer.offset[1],
					scene.layerall_pos[2] + layer.offset[2]
				)
				--palfx
				animSetPalFX(layer.anim_data, {
					time =      layer.palfx_time,
					add =       layer.palfx_add,
					mul =       layer.palfx_mul,
					sinadd =    layer.palfx_sinadd,
					invertall = layer.palfx_invertall,
					color =     layer.palfx_color
				})
			end
			--text
			if layer.text ~= '' then
				layer.text_data = text:create({
					font =   layer.font[1],
					bank =   layer.font[2],
					align =  layer.font[3],
					text =   layer.text,
					x =      scene.layerall_pos[1] + layer.offset[1],
					y =      scene.layerall_pos[2] + layer.offset[2],
					scaleX = layer.scale[1],
					scaleY = layer.scale[2],
					r =      layer.font[4],
					g =      layer.font[5],
					b =      layer.font[6],
					height = layer.font[7],
					window = layer.textwindow,
				})
			end
			--endtime
			if layer.endtime == nil then
				layer.endtime = scene.end_time
			end
		end
	end
	return t
end

function storyboard.f_preload(path)
    path = path:gsub('\\', '/')
    if storyboard.t_storyboard[path] ~= nil or not main.f_fileExists(path) then
        return
    end
    main.f_disableLuaScale()
    storyboard.t_storyboard[path] = f_parse(path)
    main.f_setLuaScale()
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
	if storyboard.t_storyboard[path] ~= nil then
		f_play(storyboard.t_storyboard[path], attract or false)
	end
	main.f_cmdBufReset()
	main.f_setLuaScale()
	if attract and main.credits > 0 then
		return true
	end
end

return storyboard

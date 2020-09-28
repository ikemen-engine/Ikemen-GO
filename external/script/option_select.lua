option_select = {}

option_select.should_load_select = true
function option_select.f_load_select()
	-- load characters data
	local file_def = (loadfile 'external/script/file_def.lua')()
	local section = 0
	local slot = false
	option_select.select_lines = {}
	option_select.select_characters = {}
	option_select.last_character_line = nil

	for select_line in io.lines(motif.files.select) do
		local parsed = file_def.parse_line(select_line)
		if parsed["kind"] == "section" then
			if parsed["section"]:lower() == "characters" then
				section = 1
			else
				if section == 1 then option_select.last_character_line = #option_select.select_lines + 1 end
				section = 0
			end
		elseif section == 1 then
			if parsed["kind"] == "data" then
				data = parsed["data"]
				if data:match('^%s*slot%s*=%s*{%s*$') then --start of the 'multiple chars in one slot' assignment
					table.insert(main.t_selGrid, {['chars'] = {}, ['slot'] = 1})
					slot = true
				elseif slot and data:match('^%s*}%s*$') then --end of 'multiple chars in one slot' assignment
					slot = false
				elseif slot == false then -- ignore multiple character slot at the moment ;TODO: do not do this
					local char_data = file_def.parse_char_line(data)
					char_data["user_enabled"] = true --TODO: deduplicate those lines
					char_data["line"] = parsed
					table.insert(option_select.select_characters, char_data)
				end
			elseif parsed["kind"] == "empty" then
				if parsed["comment"] ~= nil then
					if #parsed.comment > 13 then
						if parsed["comment"]:sub(1, 13) == "CHARDISABLED:" then
							local char_data = file_def.parse_char_line(parsed["comment"]:sub(14))
							char_data["user_enabled"] = false
							char_data["line"] = parsed
							table.insert(option_select.select_characters, char_data)
						end
					end
				end
			end
		end
		table.insert(option_select.select_lines, parsed)
	end

	-- check if the file is in fact a folder that point to a custom .def file
	local char_registred_by_folder_name = {}
	for k, character_data in ipairs(option_select.select_characters) do
		-- look for */*.def like line
		local name = character_data["name"]
		local splited = main.f_strsplit("/", name)
		local folder_name = nil
		if #splited == 2 then
			folder_name = splited[1]
		else
			folder_name = name
		end
		character_data["folder_name"] = folder_name
		char_registred_by_folder_name[folder_name:lower()] = true
	end

	-- look for character in chars/ but not in the
	for k, char_dir in ipairs(listSubDirectory("chars/")) do
		if char_registred_by_folder_name[char_dir:lower()] == nil and char_dir ~= "training" then
			-- check for .def files in the subfolder TODO: do not include the .def file if useless
			local other_char_in_dir = {}
			for k, file_name in ipairs(listFiles("chars/" .. char_dir)) do
				if #file_name > 5 then
					if file_name ~= "ending.def" and file_name ~= "intro.def" then
						if file_name:sub(-4, -1) == ".def" then
							-- add one char per variation
							local data = {user_enabled = false, name=char_dir .. "/" .. file_name, folder_name=char_dir, config={}}
							data["other_char_in_dir"] = other_char_in_dir
							table.insert(other_char_in_dir, data)
							table.insert(option_select.select_characters, data)
						end
					end
				end
			end
		end
	end

	-- generate the display_text value for characters
	local number_of_variation_in_folder = {}
	for k, char in ipairs(option_select.select_characters) do
		if char["folder_name"] ~= nil then
			if number_of_variation_in_folder[char["folder_name"]] == nil then
				number_of_variation_in_folder[char["folder_name"]] = 0
			end
			number_of_variation_in_folder[char["folder_name"]] = number_of_variation_in_folder[char["folder_name"]] + 1
		end
	end

	for k, char in ipairs(option_select.select_characters) do
		if char["display_text"] == nil then
			if number_of_variation_in_folder[char["folder_name"]] == 1 then
				char["display_text"] = char["folder_name"]
			else
				char["display_text"] = char["name"]
			end
		end
	end

	option_select.should_load_select = false
end

--TODO: reuse in option.lua
function option_select.option_numerical_plage(data, min, max, accept_nil)
	if accept_nil ~= true and data == nil then
		data = min
	end
	local changed = false
	if main.f_input(main.t_players, {'$F'}) then
		changed = true
		if data == nil then
			data = min
		else
			data = data + 1
			if data > max then
				if accept_nil then
					data = nil
				else
					data = min
				end
			end
		end
	elseif main.f_input(main.t_players, {'$B'}) then
		changed = true
		if data == nil then
			data = max
		else
			data = data - 1
			if data < min then
				if accept_nil then
					data = nil
				else
					data = max
				end
			end
		end
	end
	return data, changed
end

--TODO:
function option_select.f_generate_option_data(char_data)
	local char_option_data = {option = {}}

	function get_feedback_color(enabled)
		color = {255, 0, 0}
		if enabled then
			color = {0, 255, 0}
		end
		return color
	end

	table.insert(char_option_data.option, {
		displayname=options.f_boolDisplay(char_data.user_enabled, motif.character_edit_info.text_character_enabled, motif.character_edit_info.text_character_disabled),
		data=text:create({}),
		color = get_feedback_color(char_data.user_enabled),
		onselected = function(entry)
		if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
			sndPlay(motif.files.snd_data, motif.character_edit_info.cursor_move_snd[1], motif.character_edit_info.cursor_move_snd[2])
			if char_data.user_enabled == false then
				char_data.user_enabled = true
			else
				char_data.user_enabled = false
			end
			entry.displayname = options.f_boolDisplay(char_data.user_enabled, motif.character_edit_info.text_character_enabled, motif.character_edit_info.text_character_disabled)
			entry.color = get_feedback_color(char_data.user_enabled)
			char_data.changed = true
		end
	end})

	table.insert(char_option_data.option, {
		displayname = motif.character_edit_info.text_this_ai_level,
		data=text:create({}),
		vardisplay = char_data.option.named.ai or motif.character_edit_info.text_undefined,
		vardata = text:create({}),
		onselected = function(entry)
			char_data.option.named.ai, modified = option_select.option_numerical_plage(char_data.option.named.ai, 1, 8, true)
			if modified then
				char_data.changed = true
			end
			entry.vardisplay = char_data.option.named.ai or motif.character_edit_info.text_undefined
		end
	})

	table.insert(char_option_data.option, {
		displayname = motif.character_edit_info.text_return,
		data=text:create({}),
		onselected = function(entry)
			if main.f_input(main.t_players, {'$F', '$B', 'pal', 's'}) then
				return true
			end
		end
	})

	--TODO:
	-- some other binary option

	char_option_data["cursorPosY"] = 1
	char_option_data["moveTxt"] = 0
	char_option_data["item"] = 1
	char_option_data["first_frame"] = true
	char_option_data.char_data = char_data
	return char_option_data
end

function option_select.f_displayCharacterOption(base_x, base_y, option_char_data)
	motif.character_edit_info.menu_pos = {base_x, base_y} --TODO: maybe not use base_x and base_y
	motif.character_edit_info.is_absolute = true --TODO: maybe not use base_x and base_y

	local t = {}
	main.f_menuCommonDraw(
		option_char_data.cursorPosY,
		option_char_data.moveTxt,
		option_char_data.item,
	 	option_char_data.option,
		'fadein',
		'character_edit_info',
		'character_edit_info',
		'optionbgdef',
	 	nil,
	 	motif.defaultOptions,
	 	false,
	 	false,
	 	{},
	 	true,
		true,
		true
 	)

	option_char_data.cursorPosY, option_char_data.moveTxt, option_char_data.item = main.f_menuCommonCalc(
		option_char_data.cursorPosY,
		option_char_data.moveTxt,
		option_char_data.item,
		option_char_data.option,
		"character_edit_info",
		{"$U"},
		{"$D"}
	)

	if option_char_data.first_frame == true then -- skip the first frame, when the button is still pressed
		option_char_data.first_frame = false
	else
		local selected = option_char_data.option[option_char_data.item]
		if selected.onselected ~= nil then
			if selected.onselected(selected) == true then
				return false
			end
		end
	end

	return not esc()
end


option_select.char_ref = 0
--TODO: add scrolling animation
function option_select.f_loop_character_edit()
	--TODO: show the shortcut here
	--TODO: display character data in the right (mainly path to .def file). Maybe disable this on lower resolution

	if option_select.should_load_select then
		option_select.char_ref = 0
		resetSelect()
		option_select.f_load_select()
	end


	local portrait_scale = {heightscale(), heightscale()}

	local big_portrait_scale = {widthscale(), widthscale()}
	local space_for_data_in_right = config.GameWidth/3
	local portrait_size = {space_for_data_in_right, space_for_data_in_right*1.3}
	local big_portrait_pos = {0, 0} --TODO: center ?

	local char_display_base = {7.5*portrait_scale[1] + space_for_data_in_right, 7.5*portrait_scale[1]}
	local tile_size = {24*portrait_scale[1], 24*portrait_scale[2]} --was 75
	local space_between_portrait = {(7.5+24)*portrait_scale[1], (7.5+24)*portrait_scale[2]}
	local displayable_element = {
		math.floor((config.GameWidth - char_display_base[1]) / space_between_portrait[1]),
		math.floor((config.GameHeight - char_display_base[2]) / space_between_portrait[2])
	}
	-- optimise the worst case time of navigation
	--TODO: try to only use one screen (no scrolling)
	local char_by_line = math.floor(math.sqrt(#option_select.select_characters))
	if char_by_line > displayable_element[1] then
		char_by_line = displayable_element[1]
	end
	local selected_char_id = 1 -- the currently select id in the list, starting by 1
	local first_line_to_display = 1
	local char_by_screen = displayable_element[2] * char_by_line
	local background_enabled = {{2.5*portrait_scale[1], 2.5*portrait_scale[2]}, {170, 255, 170, 128, 0}}
	local background_disabled = {{2.5*portrait_scale[1], 2.5*portrait_scale[2]}, {255, 170, 170, 128, 0}}
	local background_selected_enabled = {{4*portrait_scale[1], 4*portrait_scale[2]}, {170, 255, 170, 200, 0}}
	local background_selected_disabled = {{4*portrait_scale[1], 4*portrait_scale[2]}, {255, 170, 170, 200, 0}}
	local in_sub_menu = false
	local editing_character = nil
	local continue = true

	local big_portrait_transition_ongoing = true
	local big_portrait_transition_new = selected_char_id
	local big_portrait_transition_old = nil
	local big_portrait_transition_progress = 0 -- float, from 0 to 1

	while continue do
		main.f_disableLuaScale()
		main.f_cmdInput()
		if in_sub_menu == false then
			-- update cursor
			local selected_another_portrait = false
			if main.f_input(main.t_players, {"$D"}) then
				if selected_char_id + char_by_line <= #option_select.select_characters then
					selected_char_id = selected_char_id + char_by_line
					selected_another_portrait = true
				end
			elseif main.f_input(main.t_players, {"$U"}) then
				if selected_char_id - char_by_line >= 1 then
					selected_char_id = selected_char_id - char_by_line
					selected_another_portrait = true
				end
			elseif main.f_input(main.t_players, {"$B"}) then
				if selected_char_id > 1 then
					selected_char_id = selected_char_id - 1
					selected_another_portrait = true
				end
			elseif main.f_input(main.t_players, {"$F"}) then
				if selected_char_id < #option_select.select_characters then
					selected_char_id = selected_char_id + 1
					selected_another_portrait = true
				end
			elseif main.f_input(main.t_players, {'b'}) then
				option_select.select_characters[selected_char_id].user_enabled = not option_select.select_characters[selected_char_id].user_enabled
				option_select.select_characters[selected_char_id].changed = true
			elseif main.f_input(main.t_players, {'pal', 's'}) then
				in_sub_menu = true
				editing_character = option_select.f_generate_option_data(option_select.select_characters[selected_char_id])
			elseif esc() then
				continue = false
			end
		end

		local first_visible_char = ((first_line_to_display - 1) * char_by_line) + 1
		if selected_char_id >= first_visible_char + char_by_screen then
			first_line_to_display = first_line_to_display + 1
		elseif selected_char_id < first_visible_char then
			first_line_to_display = first_line_to_display - 1
		end
		first_visible_char = ((first_line_to_display - 1) * char_by_line) + 1


		-- draw
		bgDraw(motif["optionbgdef"].bg, false)
		local char_pos = {char_display_base[1], char_display_base[2]}
		local char_place = {0, 0}
		for char_ref = first_visible_char, math.min(#option_select.select_characters, first_visible_char + char_by_screen - 1 + char_by_line) do --TODO: math.max
			char = option_select.select_characters[char_ref]
			if char["loaded_id"] == nil then --TODO: randomselect
				addChar(char.name)
				char.loaded_id = option_select.char_ref
				option_select.char_ref = option_select.char_ref + 1
			end

			-- draw background for each tile
			local background_to_draw = nil
			if char_ref == selected_char_id then
				if char.user_enabled == true then
					background_to_draw = background_selected_enabled
				else
					background_to_draw = background_selected_disabled
				end
			else
				if char.user_enabled == true then
					background_to_draw = background_enabled
				else
					background_to_draw = background_disabled
				end
			end

			fillRect(
				char_pos[1] - background_to_draw[1][1],
				char_pos[2] - background_to_draw[1][2],
				background_to_draw[1][1]*2+tile_size[1],
				background_to_draw[1][2]*2+tile_size[2],
				background_to_draw[2][1],
				background_to_draw[2][2],
				background_to_draw[2][3],
				background_to_draw[2][4],
				background_to_draw[2][5]
			)

			drawPortraitChar(
				char.loaded_id,
				motif.select_info.portrait_spr[1],
				motif.select_info.portrait_spr[2],
				char_pos[1],
				char_pos[2],
				portrait_scale[1],
				portrait_scale[2],
				char_pos[1],
				char_pos[2],
				tile_size[1],
				tile_size[2],
				false
			)

			char_pos[1] = char_pos[1] + space_between_portrait[1]
			char_place[1] = char_place[1] + 1
			if char_place[1] >= char_by_line then
				char_place[1] = 0
				char_place[2] = char_place[2] + 1
				char_pos[1] = char_display_base[1]
				char_pos[2] = char_pos[2] + space_between_portrait[2]
			end
		end

		-- big portrait
		local big_portrait_temp_pos = big_portrait_pos
		if big_portrait_transition_ongoing then
			big_portrait_temp_pos = {
				big_portrait_temp_pos[1],
				big_portrait_temp_pos[2] + config.GameHeight * (1-(math.sin((big_portrait_transition_progress * math.pi) / 2))) --easeOutSine
			}
			if big_portrait_transition_old ~= nil then
				local old_portrait_pos = {
					big_portrait_pos[1],
					big_portrait_pos[2] - config.GameHeight * (1 - math.cos((big_portrait_transition_progress * math.pi) / 2)) --easeInSine
				}
				drawPortraitChar(
					big_portrait_transition_old - 1,
					motif.select_info.p1_face_spr[1],
					motif.select_info.p1_face_spr[2],
					old_portrait_pos[1],
					old_portrait_pos[2],
					big_portrait_scale[1],
					big_portrait_scale[2],
					old_portrait_pos[1],
					old_portrait_pos[2],
					portrait_size[1],
					portrait_size[2],
					false
				)
			end
			big_portrait_transition_progress = big_portrait_transition_progress + 0.04
			if big_portrait_transition_progress >= 1 then
				big_portrait_transition_progress = 1
				big_portrait_transition_ongoing = false
			end
		end

		drawPortraitChar(
			big_portrait_transition_new - 1,
			motif.select_info.p1_face_spr[1],
			motif.select_info.p1_face_spr[2],
			big_portrait_temp_pos[1],
			big_portrait_temp_pos[2],
			big_portrait_scale[1],
			big_portrait_scale[2],
			big_portrait_temp_pos[1],
			big_portrait_temp_pos[2],
			portrait_size[1],
			portrait_size[2],
			false
		)

		if selected_char_id ~= big_portrait_transition_new and big_portrait_transition_ongoing == false then
			big_portrait_transition_old = big_portrait_transition_new
			big_portrait_transition_new = selected_char_id
			big_portrait_transition_progress = 0
			big_portrait_transition_ongoing = true
		end



		if in_sub_menu == true then
			if not option_select.f_displayCharacterOption(300, 100, editing_character) then
				in_sub_menu = false
			end
		end
		bgDraw(motif["optionbgdef"].bg, true)
		refresh()
	end

	main.f_setLuaScale()
end

function option_select.reload_base_character()
	if option_select.should_load_select == false then
		option_select.should_load_select = true
		resetSelect()
		load_select_def()
		start.f_generateGrid()
	end
end

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
	local unused_characted_added = false
	for k, char_dir in ipairs(listSubDirectory("chars/")) do
		if char_registred_by_folder_name[char_dir:lower()] == nil and char_dir ~= "training" then
			if unused_characted_added == false then
				table.insert(option_select.select_characters, {special = "marker", display_text = "never included characters"})
				unused_characted_added = true
			end
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

-- this module allow to parse .def file, and modify it while keeping the
-- original template.

local file_def = {}

--split strings, from main.lua ;TODO: consider the use of a common.lua file
function f_strsplit(delimiter, text)
	local list = {}
	local pos = 1
	if string.find('', delimiter, 1) then
		if string.len(text) == 0 then
			table.insert(list, text)
		else
			for i = 1, string.len(text) do
				table.insert(list, string.sub(text, i, i))
			end
		end
	else
		while true do
			local first, last = string.find(text, delimiter, pos)
			if first then
				table.insert(list, string.sub(text, pos, first - 1))
				pos = last + 1
			else
				table.insert(list, string.sub(text, pos))
				break
			end
		end
	end
	return list
end

-- parse a line of a .def file. Return a dictionary with key:
--
-- "kind": common for every line, can either be "empty" (contain no data),
--   "section", contain a section, and "data", contain data
-- "initial_withespace": the whitespace before the beggining of the data/comment
--   /end of line.
-- "have_comment": true if this line contain a comment
-- "comment": if this line have a comment, its stored here (without the ";")
-- "section": if this line is a section header, this is the contained section
-- "end_well": if this line is a section header, it is set to false if the line
--   doesn't end by "]" as expected
-- "data": if this line contain data, this is the data it contain
function file_def.parse_line(line)
	local result = {}
	local line_without_initial_space = ""
	local initial_space = ""
	local comment = ""
	local in_comment = false
	local first_char_meet = false
	for char_id = 1, #line do
		local char = line:sub(char_id, char_id)
		if not first_char_meet then
			if char == " " or char == "\t" then
				initial_space = initial_space .. char
			else
				first_char_meet = true
			end
		end
		if first_char_meet then
			if char == ";" then
				in_comment = true
			elseif in_comment then
				comment = comment .. char
			else
				line_without_initial_space = line_without_initial_space .. char
			end
		end
	end
	result["initial_whitespace"] = initial_space
	line = line_without_initial_space

	if in_comment then
		result["have_comment"] = true
		result["comment"] = comment
	else
		result["have_comment"] = false
	end

	if line:len() == 0 then
		result["kind"] = "empty"
	elseif line:sub(1, 1) == "[" then
		result["kind"] = "section"
		if not (line:sub(-1, -1) == "]") then
			print("warning: a .def file have the section header " .. line .. "that doesn't end with \"]\". considering it as " .. line:sub(2))
			result["section"] = line:sub(2)
			result["end_well"] = false
		else
			result["section"]  = line:sub(2, -2)
			result["end_well"] = true
		end
	else
		result["kind"] = "data"
		result["data"] = line
	end
	return result
end

-- recreate the .def file from a list of dictionary returned by def_parse_line
function file_def.rebuild_source_file(parsed)
	local result = ""
	for line_id = 1, #parsed do
		local line = parsed[line_id]
		result = result .. line["initial_whitespace"]
		if line["kind"] == "section" then
			result = result .. "[" .. line["section"]
			if line["end_well"] then
				result = result .. "]"
			end
		elseif line["kind"] == "data" then
			result = result .. line["data"]
		end
		if line["have_comment"] then
			result = result .. ";" .. line["comment"]
		end
		if not (line_id == #parsed) then
			result = result .. "\n"
		end
	end
	return result
end

function file_def.parse_char_line(data)
	local result = {}
	result["config"] = {}

	for j, c in ipairs(f_strsplit(',', data)) do
		if j == 1 then
			result["name"] = c
		else
			table.insert(result.config, data)
		end
	end
	return result
end

function file_def.rebuild_char(char)
	local result = char["name"]
	for k, v in ipairs(char.config) do
		result = result .. "," .. v
	end
	return result
end

return file_def

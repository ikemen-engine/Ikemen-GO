gui_tool = {}

function gui_tool.create_message(title, body)
	character_width = math.ceil(gui_tool.message_body_font:get_font_width() * gui_tool.message_body_font.scaleX)

	lines = {}
	actual_line = ""
	actual_line_length = 0
	maximal_line_length = GameWidth/2
	if GameWidth < 500 then
		maximal_line_length = GameWidth
	end
	longest_line_length = 0
	for char_id = 1, #body do
		char = body:sub(char_id, char_id)
		actual_line = actual_line .. char
		actual_line_length = actual_line_length + character_width
		if actual_line_length >= maximal_line_length then
			table.insert(lines, actual_line)
			--TODO: softwrap
			if actual_line_length > longest_line_length then
				longest_line_length = actual_line_length
			end
			actual_line_length = 0
			actual_line = ""
		end
	end

	if actual_line_length ~= 0 then
		table.insert(lines, actual_line)
		if actual_line_length > longest_line_length then
			longest_line_length = actual_line_length
		end
	end

	height_pixel = (gui_tool.message_body_font:get_font_height() + 2*2) * gui_tool.message_body_font.scaleX * #lines
	title_height = (gui_tool.message_title_font:get_font_height() + 4) * gui_tool.message_title_font.scaleX

	return {
		title=title,
		body=body,
		body_height=height_pixel,
		title_height=title_height,
		width=longest_line_length,
		lines=lines,
	}
end

gui_tool.message_title_font = text:create({
	font = motif.message.title_font[1],
	bank = motif.message.title_font[2],
	align = 0,
	text = "placeholder",
	x = GameWidth/2,
	y = nil,
	scaleX = motif.message.title_font_scale[1],
	scaleY = motif.message.title_font_scale[2],
	r = motif.message.title_font[4],
	g = motif.message.title_font[5],
	b = motif.message.title_font[6],
	src = motif.message.title_font[7],
	dst = motif.message.title_font[8],
	defsc = true
})

gui_tool.message_body_font = text:create({
	font = motif.message.body_font[1],
	bank = motif.message.body_font[2],
	align = 0,
	text = "placeholder",
	x = GameWidth/2,
	y = nil,
	scaleX = motif.message.body_font_scale[1],
	scaleY = motif.message.body_font_scale[2],
	r = motif.message.body_font[4],
	g = motif.message.body_font[5],
	b = motif.message.body_font[6],
	src = motif.message.body_font[7],
	dst = motif.message.body_font[8],
	defsc = true
})

function gui_tool.display_message(message)
	backed_lua_scale = main.f_backupLuaScale()
	main.f_disableLuaScale()

	y_delta = GameHeight/2

	-- contain the text body
	fillRect(
		GameWidth/2-message.width/2,
		y_delta - message.body_height/2,
		message.width,
		message.body_height,
		motif.message.body_color[1],
		motif.message.body_color[2],
		motif.message.body_color[3],
		motif.message.body_color[4],
		motif.message.body_color[5]
	)

	-- contain the title
	fillRect(
		GameWidth/2-message.width/2,
		y_delta - message.body_height/2 - message.title_height,
		message.width,
		message.title_height,
		motif.message.title_color[1],
		motif.message.title_color[2],
		motif.message.title_color[3],
		motif.message.title_color[4],
		motif.message.title_color[5]
	)

	-- the title
	gui_tool.message_title_font:update({
		y = y_delta - message.body_height/2 - gui_tool.message_title_font.scaleX * 3,
		text = message.title,
	})
	gui_tool.message_title_font:draw(true)


	body_font_change = (gui_tool.message_body_font:get_font_height() + 2) * gui_tool.message_body_font.scaleX

	-- the body
	for k, line in ipairs(message.lines) do
		gui_tool.message_body_font:update({
			y = y_delta - message.body_height/2 + body_font_change * k,
			text = line,
		})
		gui_tool.message_body_font:draw(true)
	end

	main.f_restoreLuaScale(backed_lua_scale)
end


return gui_tool

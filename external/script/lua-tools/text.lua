--[[
Lua Text System (v2)
Allows the creation of text objects, encouraging the reuse and updating of
text on screen.

This update recreates most of the system, reducing the amount of unused
values and expanding on possible uses.

Changes from v1:
! Creating a color is done with text.new or text.create instead of text:new or text:create
* Text objects use a shared metatable with default values, decreasing 
  memory usage.

Constructors:
* text.new(table): Text
* text.create(table): Text
Creates a new text object, using the values passed to it by the table argument.

* Text:setAlign(alignment): Text
Sets the alignment of the text object of the argument is a string.
Accepted values are 'left', 'center' or 'middle', and 'right'. Anything else
is ignored.
* Text:update(newValue): Text
Updates the text object to use the new value(s) passed in.
If newValue is a table, every value is changed to use what that table contains.
If newValue is a string, only the contents of the text are updated.
* Text:draw(): Text
Draws the text object onto the screen, if using a valid font.

]]

text = {
    mText = { --these are the default values, excluding window
        font = -1,
        bank = 0,
        align = 0,
        text = '',
        x = 0, y = 0,
        scaleX = 1, scaleY = 1,
        r = 255, g = 255, b = 255,
        height = -1,
        defsc = false
    },
    mWindow = {
        0, 0, 0,0
    }
}
text.mText.__index = text.mText
text.mWindow.__index = text.mWindow

function text.create(t)
	local o = t or {}
	setmetatable(o, text.mText)

    o.window = t.window or {}
    o.window[3] = o.window[3] or motif.info.localcoord[1]
    o.window[4] = o.window[4] or motif.info.localcoord[2]
    setmetatable(o.window, text.mWindow)

	o.ti = textImgNew()
	if o.font ~= -1 then
		if main.font[o.font .. o.height] == nil then
			--main.f_loadingRefresh(main.txt_loading)
			main.font[o.font .. o.height] = fontNew(o.font, o.height)
			main.f_loadingRefresh(main.txt_loading)
		end
		if main.font_def[o.font .. o.height] == nil then
			main.font_def[o.font .. o.height] = fontGetDef(main.font[o.font .. o.height])
		end
		textImgSetFont(o.ti, main.font[o.font .. o.height])
	end
	textImgSetBank(o.ti, o.bank)
	textImgSetAlign(o.ti, o.align)
	textImgSetText(o.ti, o.text)
	textImgSetColor(o.ti, o.r, o.g, o.b)
	if o.defsc then main.f_disableLuaScale() end
	textImgSetPos(o.ti, o.x + main.f_alignOffset(o.align), o.y)
	textImgSetScale(o.ti, o.scaleX, o.scaleY)
	textImgSetWindow(o.ti, o.window[1], o.window[2], o.window[3] - o.window[1], o.window[4] - o.window[2])
	if o.defsc then main.f_setLuaScale() end
	return o
end

text.new = text.create

function text.mText:setAlign(align)
	if align:lower() == "left" then
		self.align = -1
	elseif align:lower() == "center" or align:lower() == "middle" then
		self.align = 0
	elseif align:lower() == "right" then
		self.align = 1
	end
	textImgSetAlign(self.ti,self.align)
	return self
end

function text.mText:update(t)
	if type(t) == "table" then
		local ok = false
		local fontChange = false
		for k, v in pairs(t) do
			if self[k] ~= v then
				if k == 'font' or k == 'height' then
					fontChange = true
				end
				self[k] = v
				ok = true
			end
		end
		if not ok then return end
		if fontChange and self.font ~= -1 then
			if main.font[self.font .. self.height] == nil then
				main.font[self.font .. self.height] = fontNew(self.font, self.height)
			end
			if main.font_def[self.font .. self.height] == nil then
				main.font_def[self.font .. self.height] = fontGetDef(main.font[self.font .. self.height])
			end
			textImgSetFont(self.ti, main.font[self.font .. self.height])
		end
		textImgSetBank(self.ti, self.bank)
		textImgSetAlign(self.ti, self.align)
		textImgSetText(self.ti, self.text)
		textImgSetColor(self.ti, self.r, self.g, self.b)
		if self.defsc then main.f_disableLuaScale() end
		textImgSetPos(self.ti, self.x + main.f_alignOffset(self.align), self.y)
		textImgSetScale(self.ti, self.scaleX, self.scaleY)
		textImgSetWindow(self.ti, self.window[1], self.window[2], self.window[3] - self.window[1], self.window[4] - self.window[2])
		if self.defsc then main.f_setLuaScale() end
	else
		self.text = t
		textImgSetText(self.ti, self.text)
	end

	return self
end

function text.mText:draw()
	if self.font == -1 then return end
	textImgDraw(self.ti)
	return self
end

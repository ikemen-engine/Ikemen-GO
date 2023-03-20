--[[
Lua Rectangle System (v2)
Allows the creation of rect objects, encouraging the reuse and updating of
rectangles on screen.

This update recreates most of the system, reducing the amount of unused
values and expanding on possible uses.

Changes from v1:
* Rect objects use a shared metatable with default values, decreasing 
  memory usage.
* Support for w(idth) and h(eight) as alternatives to x2 and y2.

Color objects support various operations:
    Addition and Subtraction
    Multiplication (acts like multiply blend mode)
    Equality Check (Color == Color)

Constructors:
* rect.new(table): Rectangle
Creates a new rect object, using values passed in through the first argument.

* Rectangle:update(table): Rectangle
Updates the values of the rect based on values passed in by the first argument.
* Rectangle:draw(): Rectangle
Draws the rectangle onto the screen.

]]

rect = {
    mRect = {
        x1 = 0, y1 = 0,
        x2 = 0, y2 = 0
    }
}
rect.mRect.__index = rect.mRect

--create rect
function rect.create(t)
	local o = t or {}
    if o.w or o.width then
        o.x2 = o.w or o.width
    end
    if o.h or o.height then
        o.y2 = o.h or o.height
    end
	o.color = o.color or color.new(o.r, o.g, o.b, o.src, o.dst)
	o.r, o.g, o.b, o.src, o.dst = o.color:unpack()
	o.defsc = o.defsc or false
	setmetatable(o, rect.mRect)
	return o
end

rect.new = rect.create

--modify rect
function rect.mRect:update(t)
	for i, k in pairs(t) do
		self[i] = k
	end
    if t.w or t.width then
        self.x2 = t.w or t.width
    end
    if t.h or t.height then
        self.y2 = t.h or ot.height
    end
	if t.r or t.g or t.b or t.src or t.dst then
		self.color = color.new(t.r or self.r, t.g or self.g, t.b or self.b, t.src or self.src, t.dst or self.dst)
	end
	return self
end

--draw rect
function rect.mRect:draw()
	if self.defsc then main.f_disableLuaScale() end
	fillRect(self.x1, self.y1, self.x2, self.y2, self.r, self.g, self.b, self.src, self.dst)
	if self.defsc then main.f_setLuaScale() end
	return self
end

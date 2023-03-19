--[[
Lua Color System (v2)
Allows the creation of color objects, encouraging the reuse and updating of
colors on screen.

This update recreates most of the system, reducing the amount of unused
values and expanding on possible uses.

Changes from v1:
* Colors now range from 0.0 to 1.0 instead of 0 to 255.
* Color objects use a shared metatable with default values, decreasing 
  memory usage.

Color objects support various operations:
    Addition and Subtraction
    Multiplication (acts like multiply blend mode)
    Equality Check (Color == Color)

Constructors:
* color.new(r, g, b, src, dst): Color
Creates a new color object. All arguments are optional, all defaulting to
1 except for dst, which defaults to 0.
* color.fromHex(hex): Color
Creates a new color object from a hex string.
All sections are optional, but previous ones are required to add latter ones.
Ex. #RRGGBB works, #RG works, but #SSDD does not work.

* Color:toHex(): string
Returns a string containing the color converted to hex (RRGGBBSSDD).
* Color:unpack(): r,g,b,src,dst
Returns every value from the color separately, for use in things like arguments.

]]

color = {
    mColor = {r = 1, g = 1, b = 1, src = 1, dst = 0}
}
color.mColor.__index = color.mColor

--create color
function color.new(r, g, b, src, dst)
	local n = {r = r, g = g, b = b, src = src, dst = dst}
	setmetatable(n, color.mColor)
	return n
end

--adds rgb (color + color)
function color.mColor.__add(a, b)
	local nR = math.max(0, math.min(a.r + b.r, 255))
	local nG = math.max(0, math.min(a.g + b.g, 255))
	local nB = math.max(0, math.min(a.b + b.b, 255))
	return color.new(nR, nG, nB, a.src, a.dst)
end

--substracts rgb (color - color)
function color.mColor.__sub(a, b)
	local nR = math.max(0, math.min(a.r - b.r, 255))
	local nG = math.max(0, math.min(a.g - b.g, 255))
	local nB = math.max(0, math.min(a.b - b.b, 255))
	return color.new(nR, nG, nB, a.src, a.dst)
end

--multiply blend (color * color)
function color.mColor.__mul(a, b)
	local nR = (a.r / 255) * (b.r / 255) * 255
	local nG = (a.g / 255) * (b.g / 255) * 255
	local nB = (a.b / 255) * (b.b / 255) * 255
	return color.new(nR, nG, nB, a.src, a.dst)
end

--compares r, g, b, src, and dst (color == color)
function color.mColor.__eq(a, b)
	if a.r == b.r and a.g == b.g and a.b == b.b and a.src == b.src and a.dst == b.dst then
		return true
	else
		return false
	end
end

function color.mColor:toHex()
    return ('%02x%02x%02x%02x%02x'):format(self.r, self.g, self.b, self.src, self.dst)
end
function color.mColor:unpack()
	return tonumber(self.r), tonumber(self.g), tonumber(self.b), tonumber(self.src), tonumber(self.dst)
end

function color.fromHex(h)
	h = tostring(h)
	if h:sub(0, 1) =="#" then h = h:sub(2, -1) end
	if h:sub(0, 2) =="0x" then h = h:sub(3, -1) end
	local r = tonumber(h:sub(1, 2), 16)
	local g = tonumber(h:sub(3, 4), 16)
	local b = tonumber(h:sub(5, 6), 16)
	local src = tonumber(h:sub(7, 8), 16) or 255
	local dst = tonumber(h:sub(9, 10), 16) or 0
	return color.new(r, g, b, src, dst)
end
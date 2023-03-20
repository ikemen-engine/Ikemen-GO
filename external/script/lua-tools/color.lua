--[[
Lua Color System (v2)
Allows the creation of color objects, encouraging the reuse and updating of
colors on screen.

This update recreates most of the system, reducing the amount of unused
values and expanding on possible uses.

Changes from v1:
* Color objects use a shared metatable with default values, decreasing 
  memory usage.

Color objects support various operations:
    Addition and Subtraction
    Multiplication (acts like multiply blend mode)
    Equality Check (Color == Color)

Constructors:
* color.new(r, g, b, src, dst): Color
Creates a new color object. All arguments are optional, all defaulting to
255 except for dst, which defaults to 0.
* color.fromHex(hex): Color
Creates a new color object from a hex string.
All sections are optional, but previous ones are required to add latter ones.
Ex. #RRGGBB works, #RG works, but #SSDD does not work.
* color.fromHSL(h, s, l, src, dst): Color
Creates a new color object from HSL color.
H is a number from 0 to 360, while s and l are from 0 to 100.

* Color:toHex(): string
Returns a string containing the color converted to hex (RRGGBBSSDD).
* Color:toHSL(): h,s,l,src,dst
Returns HSL, src and dst color values separately.
* Color:unpack(): r,g,b,src,dst
Returns every value from the color separately, for use in things like arguments.

]]

color = {
    mColor = {r = 255, g = 255, b = 255, src = 255, dst = 0}
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

local function h2r(p, q, t)
	if t < 0 then t=t+1 end
	if t > 1 then t=t-1 end
	if t < 1/6 then return p+(q-p)*6*t end
	if t < 1/2 then return q end
	if t < 2/3 then return p+(q-p)*(2/3-t)*6 end
	return p
end
function color.fromHSL(h,s,l, src, dst)
	local hue, sat, lum = h/360, s/100, l/100
	local r,g,b

	if sat == 0 then
		r,g,b = lum,lum,lum
	else
		local q = lum < 0.5 and lum*(1+s) or lum+s-lum*s
		local p = 2*lum-q

		r = h2r(p,q,h + 1/3)
		g = h2r(p,q,h)
		b = h2r(p,q,h - 1/3)
	end

	return color.new(r * 255,g * 255,b * 255,src,dst)
end

function color.mColor:toHSL()
	local r,g,b = self.r/255, self.g/255, self.b/255
	local min, max = math.min(self.r, self.g, self.b), math.max(self.r, self.g, self.b)
	local h,s,l

	l = (max + min) / 2

	if max == min then return 0, 0, l*100, self.src, self.dst end

	
	local d = max - min
	s = l > .5 and d/(2-max-min) or d/(max+min)

	if max == r then
		h = (g - b) / d
		if g < b then h = h + 6 end
	elseif max == g then
		h = (b - r) / d + 2
	else
		h = (r - g) / d + 4
	end
	h = h / 6

	return h * 360, s * 100, l * 100, src, dst
end
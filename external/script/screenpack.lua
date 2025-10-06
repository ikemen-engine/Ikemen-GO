--;===========================================================
--; LOCALCOORD
--;===========================================================
--viewport calculation (offsetX, offsetY, width, height)
--from 4:3 to 16:9 => 640, 480, 1280, 720 => 0, 60, 640, 360
--from 16:9 to 4:3 => 1280, 720, 640, 480 => 160, 0, 960, 720
function main.f_getViewport(fromAspectW, fromAspectH, toAspectW, toAspectH)
	local t = {0, 0, 0, 0}
	if fromAspectW * toAspectH > fromAspectH * toAspectW then
		t[3] = fromAspectH * toAspectW / toAspectH
		t[4] = fromAspectH
		t[1] = (fromAspectW - t[3]) / 2
	elseif fromAspectW * toAspectH < fromAspectH * toAspectW then
		t[3] = fromAspectW
		t[4] = fromAspectW * toAspectH / toAspectW
		t[2] = (fromAspectH - t[4]) / 2
	else
		t = {0, 0, fromAspectW, fromAspectH}
	end
	return t
end

--enable localcoord screenpack scaling
function main.f_setLuaScale()
	setLuaSpriteScale(main.SP_Viewport43[3] / 320)
	setLuaSpriteOffsetX(main.SP_OffsetX)
	setLuaPortraitScale(main.SP_Viewport43[3] / main.SP_Localcoord[1])
end

--disable localcoord screenpack scaling
function main.f_disableLuaScale()
	setLuaSpriteScale(1)
	setLuaSpriteOffsetX(0)
	setLuaPortraitScale(1)
end

--assign storyboard localcoord scaling
function main.f_setStoryboardScale(localcoord)
	local viewport43 = main.f_getViewport(localcoord[1], localcoord[2], 4, 3)
	setLuaSpriteScale(viewport43[3] / 320)
	setLuaSpriteOffsetX(-math.floor(((localcoord[1] / (viewport43[3] / 320)) - 320) / 2))
	setLuaPortraitScale(viewport43[3] / localcoord[1])
end

--calculate and set screenpack / lifebar localcoord
function main.f_localcoord()
	--default values
	main.SP_Localcoord = {320, 240}
	main.SP_Viewport43 = {0, 0, 320, 240}
	main.SP_OffsetX = 0
	main.LB_Localcoord = {320, 240}
	main.LB_Viewport43 = {0, 0, 320, 240}
	main.LB_Viewport = {0, 0, 320, 240}
	main.LB_Scale = 1
	main.LB_OffsetX = 0
	--motif localcoord
	local tmp1, tmp2 = main.motifData:match('\n%s*localcoord%s*=%s*([0-9]+)%s*,%s*([0-9]+)')
	if tmp1 ~= nil and tmp2 ~= nil then
		main.SP_Localcoord = {tonumber(tmp1), tonumber(tmp2)}
	end
	main.SP_Viewport43 = main.f_getViewport(main.SP_Localcoord[1], main.SP_Localcoord[2], 4, 3)
	main.SP_OffsetX = -math.floor(((main.SP_Localcoord[1] / (main.SP_Viewport43[3] / 320)) - 320) / 2)
	--lifebar localcoord
	local tmp1, tmp2 = main.lifebarData:match('\n%s*localcoord%s*=%s*([0-9]+)%s*,%s*([0-9]+)')
	if tmp1 ~= nil and tmp2 ~= nil then
		main.LB_Localcoord = {tonumber(tmp1), tonumber(tmp2)}
	else
		main.LB_Localcoord = main.SP_Localcoord
	end
	main.LB_Viewport43 = main.f_getViewport(main.LB_Localcoord[1], main.LB_Localcoord[2], 4, 3)
	main.LB_Viewport = main.f_getViewport(main.LB_Localcoord[1], main.LB_Localcoord[2], gameOption('Video.GameWidth'), gameOption('Video.GameHeight'))
	main.LB_Scale = main.LB_Viewport[3] / main.LB_Localcoord[1]
	if main.LB_Scale == 1 then
		main.LB_OffsetX = (main.LB_Viewport43[3] - main.LB_Localcoord[1]) / 2
	end
	--update system vars
	setLuaLocalcoord(main.SP_Localcoord[1], main.SP_Localcoord[2])
	main.f_setLuaScale()
	setLifebarLocalcoord(main.LB_Localcoord[1], main.LB_Localcoord[2])
	setLifebarScale(320 / main.LB_Viewport43[3] * main.LB_Scale)
	setLifebarPortraitScale(main.LB_Localcoord[1] / main.LB_Viewport43[3] * main.LB_Scale)
	setLifebarOffsetX(main.LB_OffsetX * main.LB_Scale)
	--debug print
	--[[
	print('main.SP_Localcoord = ' .. main.SP_Localcoord[1] .. ', ' .. main.SP_Localcoord[2])
	print('main.SP_Viewport43 = ' .. main.SP_Viewport43[1] .. ', ' .. main.SP_Viewport43[2] .. ', ' .. main.SP_Viewport43[3] .. ', ' .. main.SP_Viewport43[4])
	print('main.SP_OffsetX = ' .. main.SP_OffsetX)
	print('main.LB_Localcoord = ' .. main.LB_Localcoord[1] .. ', ' .. main.LB_Localcoord[2])
	print('main.LB_Viewport43 = ' .. main.LB_Viewport43[1] .. ', ' .. main.LB_Viewport43[2] .. ', ' .. main.LB_Viewport43[3] .. ', ' .. main.LB_Viewport43[4])
	print('main.LB_Viewport = ' .. main.LB_Viewport[1] .. ', ' .. main.LB_Viewport[2] .. ', ' .. main.LB_Viewport[3] .. ', ' .. main.LB_Viewport[4])
	print('main.LB_Scale = ' .. main.LB_Scale)
	print('main.LB_OffsetX = ' .. main.LB_OffsetX)
	--]]
end

main.f_localcoord()

-- Screenpack config file

function main.IntLocalcoordValues()

main.SP_Localcoord = {}
main.LB_Localcoord = {}
main.SP_Localcoord43 = {}
main.LB_Localcoord43 = {}

main.SP_Localcoord[0] = 320
main.SP_Localcoord[1] = 240
main.LB_Localcoord[0] = 320
main.LB_Localcoord[1] = 240

main.LB_ScreenWidth = 320
main.LB_ScreenDiference = 0

main.screenOverscan = 0
main.normalSpriteCenter = 0

main.SP_Localcoord43[0] = 320
main.LB_Localcoord43[0] = 240

main.SP_Center = 0

end

function main.CalculateLocalcoordValues()
	if main.SP_Localcoord[0] >= main.SP_Localcoord[1] then
		main.SP_Localcoord43[0] = (main.SP_Localcoord[1] / 3) * 4
	else
		main.SP_Localcoord43[0] = (main.SP_Localcoord[0] / 4) * 3
	end
	
	if main.LB_Localcoord[0] >= main.LB_Localcoord[1] then
		main.LB_Localcoord43[0] = (main.LB_Localcoord[1] / 3) * 4
	else
		main.LB_Localcoord43[0] = (main.LB_Localcoord[0] / 4) * 3
	end
	
	main.SP_Localcoord_X_Dif = -math.floor( (( main.SP_Localcoord[0] / (main.SP_Localcoord43[0] / 320) ) - 320) / 2 )
		
	main.LB_ScreenWidth = config.Width / (config.Height / 240)
	main.LB_ScreenDiference = (main.LB_ScreenWidth - 320) / (main.LB_ScreenWidth / 320)

end

function main.IntLifebarScale()
	if config.LocalcoordScalingType == 0 then
		setLifebarOffsetX( - main.LB_ScreenDiference / 2)
		setLuaLifebarScale(main.LB_ScreenWidth / main.LB_Localcoord43[0])
	else
		setLifebarOffsetX((main.LB_Localcoord43[0] - main.LB_Localcoord[0]) / 2)
		setLuaLifebarScale(320 / main.LB_Localcoord43[0])
	end
end

function main.SetScaleValues()
	setLuaSpriteScale(main.SP_Localcoord43[0] / 320)
	setLuaSpriteOffsetX(main.SP_Localcoord_X_Dif)
	setLuaSmallPortraitScale(main.SP_Localcoord43[0] / main.SP_Localcoord[0])
	setLuaBigPortraitScale(main.SP_Localcoord43[0] / main.SP_Localcoord[0])
	main.normalSpriteCenter = main.SP_Center
	main.screenOverscan = 0
end


function main.SetDefaultScale()
	setLuaSpriteScale(1)
	setLuaSpriteOffsetX(0)
	setLuaSmallPortraitScale(1)
	setLuaBigPortraitScale(1)
	main.normalSpriteCenter = 0
	main.screenOverscan = 0
end
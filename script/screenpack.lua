-- Screenpack config file

function main.IntLocalcoordValues()
	main.SP_Localcoord = {}
	main.LB_Localcoord = {}
	main.SP_Localcoord43 = {}
	main.LB_Localcoord43 = {}

	main.SP_Localcoord[1] = 320
	main.SP_Localcoord[2] = 240
	main.LB_Localcoord[1] = 320
	main.LB_Localcoord[2] = 240

	main.LB_ScreenWidth = 320
	main.LB_ScreenDiference = 0

	main.screenOverscan = 0
	main.normalSpriteCenter = 0

	main.SP_Localcoord43[1] = 320
	main.LB_Localcoord43[1] = 240

	main.SP_Center = 0
end

function main.CalculateLocalcoordValues()
	-- We load the libar localcoord from the motif file
	main.SP_Localcoord = main.ParseDefFileValue(config.Motif, "info", "localcoord", true)
	local spOriginTemp = main.ParseDefFileValue(config.Motif, "info", "localcoord_origin", true)
	local spCenterTemp = main.ParseDefFileValue(config.Motif, "info", "localcoord_center", false)
	
	-- We check if we got a valid value
	if spCenterTemp  == nil then
		spCenterTemp = "default"
	else
		spCenterTemp = spCenterTemp:lower()
	end
	
	-- We check if what we got is valid
	if main.SP_Localcoord == nil then
		main.SP_Localcoord = {320, 240}
	end

	-- We now have to search for the config file.
	local motifFileFolder = ""
	local lbFileName = main.ParseDefFileValue(config.Motif, "files", "fight", false)
	-- Get the morif file folder.
	local tempMFF = main.f_strsplit("/",config.Motif)
	local i = 1
	-- We skip the last object on the table (The file iteself) to get only the directory.
	while i < table.getn(tempMFF) do
		motifFileFolder = motifFileFolder .. tempMFF[i] .. "/"
		i = i + 1
	end
	tempMFF = nil
	
	-- We seach for the file
	if main.file_exists(motifFileFolder .. lbFileName) then
		main.LB_Localcoord = main.ParseDefFileValue(motifFileFolder .. lbFileName, "info", "localcoord", true)
	elseif main.file_exists("data/" .. lbFileName) then
		main.LB_Localcoord = main.ParseDefFileValue("data/" .. lbFileName, "info", "localcoord", true)
	elseif main.file_exists(lbFileName) then
		main.LB_Localcoord = main.ParseDefFileValue(lbFileName, "info", "localcoord", true)
	else
		main.LB_Localcoord = main.SP_Localcoord
	end

	-- We check if what we got is valid
	if main.LB_Localcoord == nil then
		main.LB_Localcoord = main.SP_Localcoord
	end
	
	-- And we calculate some extra stuff.
	if main.SP_Localcoord[1] >= main.SP_Localcoord[2] then
		main.SP_Localcoord43[1] = (main.SP_Localcoord[2] / 3) * 4
	else
		main.SP_Localcoord43[1] = (main.SP_Localcoord[1] / 4) * 3
	end
	
	if main.LB_Localcoord[1] >= main.LB_Localcoord[2] then
		main.LB_Localcoord43[1] = (main.LB_Localcoord[2] / 3) * 4
	else
		main.LB_Localcoord43[1] = (main.LB_Localcoord[1] / 4) * 3
	end
	
	main.SP_Localcoord_X_Dif = -math.floor( (( main.SP_Localcoord[1] / (main.SP_Localcoord43[1] / 320) ) - 320) / 2 )
		
	main.LB_ScreenWidth = config.Width / (config.Height / 240)
	main.LB_ScreenDiference = (main.LB_ScreenWidth - 320) / (main.LB_ScreenWidth / 320)
	--setLifebarPortaitScale(main.SP_Localcoord[1] / main.SP_Localcoord43[1])

	-- Now we load posible values of main.SP_Center
	if spOriginTemp == nil then
		if spCenterTemp == "center" then
			main.SP_Center = main.SP_Localcoord[1] / 2
		elseif spCenterTemp == "left" then
			main.SP_Center = 0
		elseif spCenterTemp == "right" then
			main.SP_Center = main.SP_Localcoord[1]
		else
			main.SP_Center = main.SP_Localcoord[1] - main.SP_Localcoord43[1]
		end
	else 
		main.SP_Center = spOriginTemp
	end
end

function main.IntLifebarScale()
	if config.LocalcoordScalingType == 0 then
		setLifebarOffsetX( - main.LB_ScreenDiference / 2)
		setLuaLifebarScale(main.LB_ScreenWidth / main.LB_Localcoord43[1])
	else
		setLifebarOffsetX((main.LB_Localcoord43[1] - main.LB_Localcoord[1]) / 2)
		setLuaLifebarScale(320 / main.LB_Localcoord43[1])
	end
end

function main.SetScaleValues()
	setLuaSpriteScale(main.SP_Localcoord43[1] / 320)
	setLuaSpriteOffsetX(main.SP_Localcoord_X_Dif)
	setLuaSmallPortraitScale(main.SP_Localcoord43[1] / main.SP_Localcoord[1])
	setLuaBigPortraitScale(main.SP_Localcoord43[1] / main.SP_Localcoord[1])
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

-- Edited version of the parser in motif.lua, made to parse only a single value and end once it steps outside [searchBlock]
function main.ParseDefFileValue(argFile, searchBlock, searchParam, isNumber)
	-- We use 'arg' inestead of 'config.Motif' because we also want the option to parse the lifebar
	local file = io.open(argFile)
	local weAreInInfo = 0
	local ret = {}
	local ipos = 0

	for line in file:lines() do
		ipos = ipos +1
		if weAreInInfo ~= 2 then
			local line = line:gsub('%s*;.*$', '')
			if line:match('^%s*%[.-%s*%]%s*$') then -- matched [] group
				line = line:match('^%s*%[(.-)%s*%]%s*$') -- match text between []
				line = line:gsub('[%. ]', '_') -- change . and space to _
				line = line:lower() -- lowercase line
				local row = tostring(line:lower()) -- just in case it's a number (not really needed)

				if row == searchBlock then -- matched info
					weAreInInfo = 1
				else
					if weAreInInfo == 1 then weAreInInfo = 2 end
				end
			elseif weAreInInfo == 1 then -- matched non [] line inside [Info]
				local param, value = line:match('^%s*([^=]-)%s*=%s*(.-)%s*$')
				if param ~= nil then
					param = param:gsub('[%. ]', '_') -- change param . and space to _
					param = param:lower() -- lowercase param
				end
				if param ~= nil and value ~= nil and param == searchParam then -- param = value pattern matched
					value = value:gsub('"', '') -- remove brackets from value
					if value:match('.+,.+') then -- multiple values
						for i, c in ipairs(main.f_strsplit(',', value)) do -- split value using "," delimiter
							if c == nil or c == '' then
								ret[i] = nil
							else
								if isNumber == true then
									ret[i] = tonumber(c)
								else
									ret[i] = c
								end
							end
						end
					else --single value
						if isNumber == true then
							ret = tonumber(value)
						else
							ret = value
						end
					end
				end
			end
		end
	end
	file:close()

	-- Let's check if the table values are valid
	if type(ret) == "table" and (ret[1] == nil or ret[2] == nil) then
		-- If not we return nil
		ret = nil
	end

	-- Return what we parsed
	return ret
end
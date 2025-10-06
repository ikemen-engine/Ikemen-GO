-- This script handles the logic for a game mode where random battles are run, and AI difficulty is adjusted based on character performance.
-- The script reads from the 'autolevel.save' file, which contains win/loss data for each character per palette.
-- The 'autolevel.save' file is updated by the engine's backend (Golang code) after each match.
-- This script generates AI rank data to adjust the difficulty of opponents, creating a dynamic and challenging gameplay experience.
-- It selects characters for battles based on their AI ranks, ensuring a varied and balanced roster.

local randomtest = {}

-- Trims leading and trailing whitespace from a string
function randomtest.trimString(s)
	return string.match(s, '^()%s*$') and '' or string.match(s, '^%s*(.*%S)')
end

-- Applies a function to each element in a table
function randomtest.mapTable(func, tbl)
	local result = {}
	for k, v in pairs(tbl) do
		result[k] = func(v)
	end
	return result
end

-- Integer variables
local strongThreshold = 0  -- Threshold for strong characters based on win counts
local paletteCount = 12    -- Number of palettes per character
local recentRanks = {}     -- List of recent ranks to avoid repetition
local currentRank = 0      -- Current AI rank being used
local isStrongest = false  -- Flag to indicate if the strongest characters are being used
local characterRoster = {} -- List of characters selected for the roster
local debugText = ''       -- Debug information
local numCharacters = 0    -- Total number of characters available
local nextCharacterIndex = 1  -- Index for the next character to be selected from the roster

-- Adds a rank to recentRanks and limits its size
function randomtest.addRecentRank(rank)
	recentRanks[#recentRanks + 1] = rank
	local maxRecent = math.floor(numCharacters / (math.min(numCharacters / (paletteCount * 10) + 3, paletteCount) * paletteCount))
	while #recentRanks > maxRecent do
		table.remove(recentRanks, 1)
	end
end

-- Generates a random rank not close to recent ranks
function randomtest.getRandomRank()
	local rank = 0
	while true do
		rank = math.random(1, strongThreshold + paletteCount - 2)
		local isValid = true
		for i = 1, #recentRanks do
			if math.abs(recentRanks[i] - rank) <= math.floor(paletteCount / 3) then
				isValid = false
				break
			end
		end
		if isValid then
			break
		end
	end
	return rank
end

-- Applies a function to all characters in the roster
function randomtest.forAllCharacters(func)
	for index = 1, #main.t_randomChars do
		func(main.t_randomChars[index])
	end
end

-- Reads win counts from file, updates character rankings, and builds roster
function randomtest.updateWinCounts()
	local autoLevelFile = 'save/autolevel.save'
	local characterWins = {}  -- Stores win counts per character per palette
	local winCounts = {}      -- Temporary storage for parsed win counts
	local buffer = '\239\187\191'  -- Byte Order Mark for UTF-8
	local file = io.open(autoLevelFile, 'r')
	if file then
		for line in file:lines() do
			local tmp = main.f_strsplit(',', line)
			if #tmp >= 2 then
				-- Remove BOM if present
				for i = 1, 4 do
					if i == 4 then
						tmp[1] = string.sub(tmp[1], 4)
					else
						if string.byte(tmp[1], i) ~= string.byte(buffer, i) then break end
					end
				end
				-- tmp[1] is character definition (e.g., file path)
				-- tmp[2] is win counts per palette as a string
				winCounts[tmp[1]] = randomtest.mapTable(tonumber, main.f_strsplit(' ', randomtest.trimString(tmp[2])))
			end
		end
		io.close(file)
	end
	numCharacters = 0
	randomtest.forAllCharacters(function()
		numCharacters = numCharacters + 1
	end)
	local strongCharacterCount = math.floor(numCharacters / (paletteCount * 10))
	if strongCharacterCount < paletteCount - 1 then
		strongThreshold = math.floor(numCharacters / (strongCharacterCount + 1))
		strongCharacterCount = paletteCount - 1
	else
		strongThreshold = math.floor(numCharacters / paletteCount)
	end
	local totalWins = 0
	local zeroWinsList = {}       -- Characters with zero wins
	local strongCharacters = {}   -- Strong characters with high win counts
	local randomRankCharacters = {}  -- Characters within a random rank range
	local negativeWinsList = {}   -- Characters with negative win counts (more losses)
	local averageCharacters = {}  -- Characters with average win counts
	local strongCount = 0         -- Count of strong characters
	local randomRank = randomtest.getRandomRank()
	randomtest.forAllCharacters(function(characterIndex)
		-- Ensure characterWins is large enough
		if #characterWins < characterIndex * paletteCount then
			for i = #characterWins + 1, characterIndex * paletteCount do
				characterWins[i] = 0
			end
		end
		local wins = winCounts[getCharFileName(characterIndex)]
		local totalCharacterWins = 0
		for paletteIndex = 1, paletteCount do
			if wins and paletteIndex <= #wins then
				totalWins = totalWins + wins[paletteIndex]
				characterWins[characterIndex * paletteCount + paletteIndex] = wins[paletteIndex]
				totalCharacterWins = totalCharacterWins + wins[paletteIndex]
			else
				characterWins[characterIndex * paletteCount + paletteIndex] = 0
			end
		end
		if totalCharacterWins >= strongThreshold then strongCount = strongCount + 1 end
		if totalCharacterWins >= strongThreshold - paletteCount then table.insert(strongCharacters, characterIndex) end
		if totalCharacterWins >= 1 and totalCharacterWins <= paletteCount then table.insert(averageCharacters, characterIndex) end
		if totalCharacterWins > randomRank - paletteCount and totalCharacterWins <= randomRank then table.insert(randomRankCharacters, characterIndex) end
		if totalCharacterWins == 0 then table.insert(zeroWinsList, characterIndex) end
		if totalCharacterWins < 0 then table.insert(negativeWinsList, characterIndex) end
	end)
	-- Function to add characters from a list to the roster
	local function addCharactersToRoster(characterList, numToAdd)
		if numToAdd <= 0 then return end
		for i = 1, numToAdd do
			if #characterList == 0 then break end
			local index = math.random(1, #characterList)
			table.insert(characterRoster, characterList[index])
			table.remove(characterList, index)
		end
	end
	characterRoster = {}
	nextCharacterIndex = 1
	debugText = ''
	local numZeroWins = #zeroWinsList
	if numZeroWins > 0 then
		addCharactersToRoster(zeroWinsList, numZeroWins)
		addCharactersToRoster(negativeWinsList, strongCharacterCount - numZeroWins)
		currentRank = 0
	elseif #averageCharacters >= math.max(strongCharacterCount * 20, math.floor((numCharacters * 3) / 20)) then
		addCharactersToRoster(averageCharacters, #averageCharacters)
		currentRank = paletteCount
	else
		for _ = 1, 3 do
			if #randomRankCharacters >= strongCharacterCount then break end
			randomRankCharacters = {}
			randomRank = randomtest.getRandomRank()
			randomtest.forAllCharacters(function(characterIndex)
				local totalCharacterWins = 0
				for paletteIndex = 1, paletteCount do
					totalCharacterWins = totalCharacterWins + characterWins[characterIndex * paletteCount + paletteIndex]
				end
				if totalCharacterWins > randomRank - paletteCount and totalCharacterWins <= randomRank then
					table.insert(randomRankCharacters, characterIndex)
				end
			end)
		end
		debugText = randomRank .. ' ' .. #randomRankCharacters
		if #randomRankCharacters >= strongCharacterCount then
			addCharactersToRoster(randomRankCharacters, #randomRankCharacters)
			currentRank = randomRank
			randomtest.addRecentRank(currentRank)
		elseif strongCount >= strongCharacterCount then
			addCharactersToRoster(strongCharacters, #strongCharacters)
			currentRank = strongThreshold + paletteCount - 1
		else
			randomtest.addRecentRank(strongThreshold + (paletteCount - 2) - math.floor(paletteCount / 3))
			addCharactersToRoster(negativeWinsList, #negativeWinsList)
			currentRank = -1
		end
	end
	if numZeroWins == 0 then
		while totalWins ~= 0 do
			local i = math.random(1, #characterWins)
			if totalWins > 0 then
				characterWins[i] = characterWins[i] - 1
				totalWins = totalWins - 1
			else
				characterWins[i] = characterWins[i] + 1
				totalWins = totalWins + 1
			end
		end
	end
	-- Write the updated win counts back to the file
	-- File format:
	-- Each line represents a character and their win counts per palette
	-- Format: <character_def>, <win1> <win2> <win3> ... <win12>
	-- Example: chars/kfm/kfm.def, 5 3 2 0 -1 -2 -3 1 0 0 0 0
	randomtest.forAllCharacters(function(characterIndex)
		buffer = buffer .. getCharFileName(characterIndex) .. ','
		for paletteIndex = 1, paletteCount do
			buffer = buffer .. ' ' .. characterWins[characterIndex * paletteCount + paletteIndex]
		end
		buffer = buffer .. '\r\n'
	end)
	local outputFile = io.open(autoLevelFile, 'wb')
	outputFile:write(buffer)
	io.close(outputFile)
	-- Debug print to console
	print("Updated win counts written to " .. autoLevelFile)
end

-- Randomly selects a character for a player, considering rank and winner
function randomtest.randomSelect(playerNumber, winner)
	if winner > 0 and (playerNumber == winner) == not isStrongest then return end
	local teamMode
	if currentRank == 0 or currentRank == paletteCount or isStrongest then
		teamMode = 0  -- Single
	elseif currentRank < 0 then
		teamMode = math.random(0, 2)  -- Any team mode
	else
		teamMode = math.random(0, 1) * 2  -- Single or Turns
	end
	setTeamMode(playerNumber, teamMode, math.random(1, 4))
	start.p[playerNumber].teamMode = teamMode
	local tmp = 0
	while tmp < 2 do
		tmp = selectChar(playerNumber, characterRoster[nextCharacterIndex], getCharRandomPalette(characterRoster[nextCharacterIndex]))
		nextCharacterIndex = nextCharacterIndex + 1
		if nextCharacterIndex > #characterRoster then nextCharacterIndex = 1 end
	end
end

-- Writes the current rank and character roster to a save file
function randomtest.writeRosterFile()
	local str = "Rank: " .. currentRank .. ' ' .. debugText
	for i = 1, #characterRoster do
		str = str .. '\n' .. getCharFileName(characterRoster[i])
	end
	local file = io.open('save/AI_Rank.save', 'w')
	file:write(str)
	io.close(file)
end

-- Initializes AI levels, AutoLevel, and prepares for random battle
function randomtest.init()
	for i = 1, 8 do
		setCom(i, 8)  -- Set AI level to maximum
	end
	setAutoLevel(true)
	setMatchNo(1)
	randomtest.updateWinCounts()
	winner = 0
	wins = 0
	randomtest.writeRosterFile()
	nextCharacterIndex = 1
	isStrongest = currentRank == strongThreshold + paletteCount - 1
end

-- Main loop for running random battles
function randomtest.run()
	clearColor(0, 0, 0)
	randomtest.init()
	refresh()
	clearSelected()
	while not esc() do
		randomtest.randomSelect(1, winner)
		randomtest.randomSelect(2, winner)
		local stage = start.f_setStage()
		start.f_setMusic(stage)
		loadStart()
		local previousWinner = winner
		winner = game()
		clearColor(0, 0, 0)
		if winner < 0 or esc() then break end
		local previousWins = wins
		wins = wins + 1
		if winner ~= previousWinner then
			wins = 1
			setHomeTeam(winner == 1 and 2 or 1)
		end
		setMatchNo(wins)
		if winner <= 0 or wins >= 20 or wins == previousWins then
			randomtest.init()
		end
		refresh()
	end
	main.f_bgReset(motif[main.background].bg)
	main.f_fadeReset('fadein', motif[main.group])
	main.f_playBGM(true, motif.music.title_bgm, motif.music.title_bgm_loop, motif.music.title_bgm_volume, motif.music.title_bgm_loopstart, motif.music.title_bgm_loopend)
end

return randomtest

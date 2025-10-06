--========================================================
--  A.N.I.M.E  |  Random Menu Music Selector
--  Powered by Straw Hat Devs 🏴‍☠️
--========================================================

-- Folder where the OGG files are stored (relative path)
local musicFolder = "./sound/menu_music/"

-- List of available tracks
local tracks = {
    "showdown.ogg",
    "plot_amor.ogg",
    "zenitsu_uk_drill.ogg"
}

-- Seed the random number generator
math.randomseed(os.time())

-- Pick a random track from the list
local chosen = tracks[math.random(#tracks)]
local selectedTrack = musicFolder .. chosen

-- Debug print in console
print("[A.N.I.M.E] Now playing: " .. selectedTrack)

-- Function to check if file exists
local function fileExists(name)
    local f = io.open(name, "r")
    if f then f:close() return true else return false end
end

-- Play the soundtrack if the file exists
if fileExists(selectedTrack) then
    bgmPlay(selectedTrack, true, 1.0) -- loop = true, volume = 1.0
else
    print("[A.N.I.M.E] ERROR: File not found -> " .. selectedTrack)
end

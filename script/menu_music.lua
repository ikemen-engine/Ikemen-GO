--========================================================
--  A.N.I.M.E  |  Robust Random Menu Music for Android & PC
--  Powered by Straw Hat Devs 🏴‍☠️
--========================================================

-- Folder where the OGG files are stored
local musicFolder = "sound/menu_music/"

-- List of available tracks
local tracks = {
    "showdown.ogg",
    "plot_armor.ogg",
    "zenitsu_uk_drill.ogg"
}

-- Seed RNG
math.randomseed(os.time())

-- Function to check if file exists
local function fileExists(path)
    local f = io.open(path, "r")
    if f then f:close() return true else return false end
end

-- Function to pick a random valid track
local function pickRandomTrack()
    local validTracks = {}
    for _, track in ipairs(tracks) do
        local fullPath = musicFolder .. track
        if fileExists(fullPath) then
            table.insert(validTracks, fullPath)
        else
            print("[A.N.I.M.E] WARNING: Missing file -> " .. fullPath)
        end
    end
    if #validTracks == 0 then
        return nil
    end
    return validTracks[math.random(#validTracks)]
end

-- Pick and play
local selectedTrack = pickRandomTrack()
if selectedTrack then
    print("[A.N.I.M.E] Now playing: " .. selectedTrack)
    bgmPlay(selectedTrack, true, 1.0)
else
    print("[A.N.I.M.E] ERROR: No valid music files found!")
end

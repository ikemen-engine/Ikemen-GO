--========================================================
--  A.N.I.M.E  |  Random Menu Music Selector
--  Powered by Straw Hat Devs 🏴‍☠️
--========================================================

-- Folder where the OGG files are stored
local musicFolder = "sound/menu_music/"

-- List of available tracks
local tracks = {
    "showdown.ogg",
    "plot_amor.ogg",
    "zenitsu_uk_drill.ogg"
}

-- Seed RNG and pick a random track
math.randomseed(os.time())
local chosen = tracks[math.random(#tracks)]
local selectedTrack = musicFolder .. chosen

-- Debug print (optional, shows in console)
print("[A.N.I.M.E] Now playing: " .. selectedTrack)

-- Play the soundtrack on loop
bgmPlay(selectedTrack, true, 1.0)

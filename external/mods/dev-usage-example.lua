-- The point of this file is to show how a GAME DEVELOPER WOULD USE IT

-- Function for hte purpose of testing config such that it is correctly sent to server
local function testConfigPrint()
  local config_test_response = sblib.init("external/mods/config-example")
  print("Server response: ", config_test_response)
end
hook.add("launchFight","test", testConfigPrint);


-- Function that runs every frame and serves sb-lib with game_state variables
-- runs every n steps (temporarily hardcode)
local frame = 0
local function stepWithGameState()
  frame = frame + 1
  -- Run step which mutates the game state with activations from server
  sblib.step(frame)
  -- print("Ikemon go Players game_state ", sblib.json.encode(mutated_game_state))
end

hook.add("loop#watch","state", stepWithGameState);
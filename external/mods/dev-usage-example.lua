-- The point of this file is to show how a GAME DEVELOPER WOULD USE IT

-- Function that sets up config and sends it to the server
local function testConfigPrint()
  local config_test_response = sblib.init("external/mods/config-example")
  print("Server response: ", config_test_response)
end
hook.add("launchFight","test", testConfigPrint);


-- Function that runs every frame and serves sb-lib with game_state variables this runs every frame Interval
-- See config for frame interval
local frame = 0
local function stepWithGameState()
  frame = frame + 1
  
  -- Run step which mutates the game state with activations from server
  sblib.step(frame)

end

hook.add("loop#watch","state", stepWithGameState);
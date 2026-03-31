-- The point of this file is to show how a GAME DEVELOPER WOULD USE IT

-- Function for hte purpose of testing config such that it is correctly sent to server
local function testConfigPrint()
  local config_test_response = sblib.init("external/mods/config-example")
  print("Server response: ", config_test_response)
end
hook.add("launchFight","test", testConfigPrint);


-- Function that runs every frame and serves sb-lib with game_state variables
-- runs every n steps (temporarily hardcode)
local frames = 0
local frameInterval = 15
local function stepWithGameState()
  local game_state = {}
  -- Insert player 1 and player 2 variables into game_state
  if player(1) then
    p1life = life()
    p1attackmul = attackmul()
    game_state.p1life = life()
    game_state.p1attackmul  = attackmul()
  end
  if player(2) then
    p2life = life()
    p2attackmul = attackmul()
    game_state.p2life = life()
    game_state.p2attackmul  = attackmul()
  end

  frames = frames+1
  if frames % frameInterval ~= 0 then
    return
  end
  
  -- Run step which mutates the game state with activations from server
  local mutated_game_state = sblib.step(game_state)
  if player(1) then
    setAttackMul(mutated_game_state.p1attackmul)
  end
  if player(2) then
    setAttackMul(mutated_game_state.p2attackmul)
  end
  print("Ikemon go Players game_state ", sblib.json.encode(mutated_game_state))

end

hook.add("loop#watch","state", stepWithGameState);
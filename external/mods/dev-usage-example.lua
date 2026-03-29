-- The point of this file is to show how a GAME DEVELOPER WOULD USE IT

-- Function for hte purpose of testing config such that it is correctly sent to server
local function testConfigPrint()
  local config_test_response = sblib.init("external/mods/config-example")
  print("Server response: ", config_test_response)
end
hook.add("launchFight","test", testConfigPrint);


-- Function that runs every frame and serves sb-lib with state variables
local function stepWithState()
  local state = {}
  if player(1) then
    p1redlife = redlife()
    p1attackmul = attackmul()
    -- Inserts player 1 state into state
    state.p1redlife = redlife()
    state.p1attackmul  = attackmul()
  end

  if player(2) then
    p2redlife = redlife()
    p2attackmul = attackmul()
    -- Inserts player 1 state into state
    state.p2redlife = redlife()
    state.p2attackmul  = attackmul()
  end

  
  mutated_state = sblib.step(state)
  -- print("Ikemon go Players State ", sblib.json.encode(mutated_state))
  if player(1) then
    setAttackMul(mutated_state.p1attackmul)
  end
  if player(2) then
    setAttackMul(mutated_state.p2attackmul)
  end
  print("Ikemon go Players State ", sblib.json.encode(mutated_state))

end

hook.add("loop#watch","state", stepWithState);
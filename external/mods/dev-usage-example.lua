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
    if (redlife() == 0 and attack()== 0) then return end
    -- print("Player 1 redlife " .. redlife())
    -- print("Player 1 attack " .. attack())

    p1redlife = redlife()
    p1attack = attack()
    -- Inserts player 1 state into state
    table.insert(state, p1redlife)
    table.insert(state, p1attack)
  end

  if player(2) then
    if (redlife() == 0 and attack()== 0) then return end
    -- print("Player 2 redlife " .. redlife())
    -- print("Player 2 attack " .. attack())

    -- Inserts player 2 state into state
    p2redlife = redlife()
    p2attack = attack()
    -- Inserts player 1 state into state
    table.insert(state, p2redlife)
    table.insert(state, p2attack)
  end

  -- Prints json encoded Ikemon go state
  print("Ikemon go Players State ", sblib.json.encode(state))

  -- Calls step from sblib with state
  -- Step doesn't work right now so it's not being called
  sblib.step(state)
end

hook.add("loop#watch","state", stepWithState);
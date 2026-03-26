-- The point of this file is to show how a GAME DEVELOPER WOULD USE IT

-- Add hooks and stuff that runs every frame here
-- Add the config for sblib
-- Skillbalancer has a bunch of good examples

-- 
local function testConfigPrint()
  local config_test_response = sblib.init("external/mods/config-example")
  print("Server response: " .. tostring(config_test_response))
end
hook.add("launchFight","test", testConfigPrint);
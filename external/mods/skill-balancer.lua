-- send http request

local function blocking_httpget(url)
  local handle = io.popen("curl -s " .. url) -- Is blocking and may cause stutter
  if not handle then
    print("Failed to execute curl command")
    return nil
  end
  local result = handle:read("*a")
  handle:close()
  return result
end

local response = blocking_httpget("http://localhost:5656/hifromlua")
print(response)

response = httpget("http://localhost:5656/hifromgo") -- Using custom go patch function
print(response)

response = httppost("http://localhost:5656/hifromgo", "application/json", '{"name":"Alfred"}') -- Using custom go patch function
print(response)

function setLevels(p1, p2)
  if player(1) then
    setAILevel(p1)
    print("Player 1 AI level set to " .. ailevel())
  end
  if player(2) then
    setAILevel(p2)
    print("Player 2 AI level set to " .. ailevel())
  end
end

function printLife()
  if player(1) then
    print("Player 1 Life: " .. life())
  end
  if player(2) then
    print("Player 2 Life: " .. life())
  end
end

addHotkey('p', true, false, false, true, false, 'setCom(1,0)')
addHotkey('o', true, false, false, true, false, 'setCom(2,0)')
addHotkey('u', true, false, false, true, false, 'setLevels(1,6)')
addHotkey('l', true, false, false, true, false, 'printLife(1,6)')

local frames = 0
local function testPrint()
  frames = frames + 1
  print("frame: ".. frames)
end
hook.add("loop#watch","test", testPrint);




-- setGameSpeed(10000)
print("SB Loaded")

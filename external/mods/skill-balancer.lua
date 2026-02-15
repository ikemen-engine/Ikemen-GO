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

response = httppost("http://localhost:5656/hifromgo","application/json",'{"name":"Alfred"}') -- Using custom go patch function
print(response)

print("SB Loaded")

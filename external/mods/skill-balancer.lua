
-- send http request

function http_request(url)
    local handle = io.popen("curl -s " .. url) -- Is blocking and may cause stutter
    local result = handle:read("*a")
    handle:close()
    return result
end

local data = http_request("http://localhost:5656/hifromlua")

print("SB Loaded")

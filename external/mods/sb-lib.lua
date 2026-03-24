sblib =  {}

-- Set to the reward function from the config table
-- This reward function should be used to calculate reward client side
-- This function IS the one to be used to calculate reward (So call this variable in step)
local reward_function;

-- Call action names in step
local action_names;


-- Sends and setups config
-- Sets reward
-- Saves action names as well
-- Can maybe
function sblib.setup_config (config_table)
end


-- Sends ikemon go data to the server as per the contract (Decide with MO)
-- Returns
-- Needs to give the current state.
function sblib.step (state)
    -- If statement that checks if all game values are in the right format

    -- Calculate actions based on the http call
    adjustments = httppost(baseurl + "/step", state)
    sblib.apply_adjustments(adjustments)
end

function sblib.apply_adjustments (adjustments)
    -- Apply the adjustments as per the config apply functions 
end
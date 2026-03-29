-- Reward encourage equal redlife. For balanced game
local function reward_function(current_state)
    local diff = math.abs(current_state.p1redlife - current_state.p2redlife)

    -- Negative penalty for imbalance
    return -diff
end

local function apply_attack_mul_p1 (state, value) 
    state.p1attackmul = state.p1attackmul + (value * 0.01)
end

local function apply_attack_mul_p2 (state, value) 
    state.p2attackmul = state.p2attackmul + (value * 0.01)
end



return {
    name = "ikemon-test",
    endpoint = "http://localhost:3000",
    description = "Sample RL config",
    reward_function = reward_function,
    state = {
        "p1redlife",
        "p1attackmul",
        "p2redlife",
        "p2attackmul",
    },
    actions = {
        apply_attack_mul_p1 = apply_attack_mul_p1,
        apply_attack_mul_p2 = apply_attack_mul_p2,
    },
    
    hyperparameters = {
        learning_rate = 0.01,
        discount_factor = 0.99
    }
}
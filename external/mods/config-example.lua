-- Reward encourage equal redlife. For balanced game
local function reward_function(current_game_state)
    local diff = math.abs(current_game_state.p1redlife - current_game_state.p2redlife)

    -- Negative penalty for imbalance
    return -diff
end

local function apply_attack_mul_p1 (game_state, value) 
    game_state.p1attackmul = game_state.p1attackmul + (value * 0.01)
end

local function apply_attack_mul_p2 (game_state, value) 
    game_state.p2attackmul = game_state.p2attackmul + (value * 0.01)
end



return {
    name = "ikemon-test",
    endpoint = "http://localhost:3000",
    description = "Sample RL config",
    reward_function = reward_function,
    game_state = {
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
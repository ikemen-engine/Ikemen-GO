-- Reward: Encourages Player 1 to win by maximizing damage to Player 2 (minimizing p2 redlife)
-- while avoiding damage to itself.
local function reward_function(prev, curr)
    if not prev then return 0 end

    local damage_to_enemy = prev.p2redlife - curr.p2redlife
    local damage_to_self  = prev.p1redlife - curr.p1redlife

    local attack_gain = curr.p1attack - prev.p1attack

    return
        (damage_to_enemy * 5)   -- main goal
      - (damage_to_self * 3)    -- don't die
      + (attack_gain * 0.5)     -- encourage scaling
end

-- Example action functions all these try
local function increase_attack(state, value)
    state.variables.p1attack =
        state.variables.p1attack + (value * 2)
end

local function decrease_attack(state, value)
    state.variables.p1attack =
        state.variables.p1attack - (value * 1)
end

local function heal(state, value)
    state.variables.p1redlife =
        state.variables.p1redlife + (value * 2)
end



return {
    name = "ikemon-test",
    version = "1.0",
    endpoint = "http://localhost:3000",
    description = "Sample RL config",
    reward_function = reward_function,
    state = {
        "p1redlife",
        "p1attack",
        "p2redlife",
        "p2attack",
    },

    actions = {
        buff_attack = {
            application_function = increase_attack,
        },

        weaken_attack = {
            application_function = decrease_attack,
        },

        heal = {
            application_function = heal,
        }
    },
    
    hyperparameters = {
        learning_rate = 0.01,
        discount_factor = 0.99
    }
}
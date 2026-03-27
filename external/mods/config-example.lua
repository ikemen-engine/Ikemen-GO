local function reward_function(state)
    -- Example: reward based on health
    return state.variables.health
end

-- Example action functions
local function increase_attack(state, value, impact)
    state.variables.attack = state.variables.attack + (value * impact)
end

local function heal(state, value, impact)
    state.variables.health = state.variables.health + (value * impact)
end

local function decrease_stamina(state, value, impact)
    state.variables.stamina = state.variables.stamina - (value * impact)
end


return {
    name = "ikemon-test",
    version = "1.0",
    endpoint = "http://localhost:3000",
    description = "Sample RL config",
    reward_function = reward_function,
    state = {
        "p1life",
        "p1attack",
        "p2life",
        "p2attack",
    },

    actions = {
        attack = {
            application_function = increase_attack,
            impact = 2
        },

        heal = {
            application_function = heal,
            impact = 5
        },

        rest = {
            application_function = decrease_stamina,
            impact = 1
        }
    },

    hyperparameters = {
        learning_rate = 0.01,
        discount_factor = 0.99
    }
}
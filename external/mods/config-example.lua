-- Reward encourage equal life. For balanced game

-- Every variable here needs a getter
local function reward_function(current_game_state)
    local diff = math.abs(current_game_state.p1life - current_game_state.p2life)
    return ((diff-1000)/10)-50
end

-- Every variable used here needs a setter
local function apply_attack_mul_p1 (game_state, value) 
    game_state.p1attackmul = game_state.p1attackmul + (value * 0.01)
end
local function apply_attack_mul_p2 (game_state, value) 
    game_state.p2attackmul = game_state.p2attackmul + (value * 0.01)
end

local function get_p1_life () 
  if player(1) then
    return life()
  end
end
local function get_p2_life () 
  if player(2) then
    return life()
  end
end

local function get_p1_attackMul()
  if player(1) then
    return attackmul()
  end
end

local function get_p2_attackMul()
  if player(2) then
    return attackmul()
  end
end


local function set_p1_attackMul (game_state) 
  if player(1) then
    setAttackMul(game_state.p1attackmul)
  end
end

local function set_p2_attackMul (game_state) 
  if player(2) then
    setAttackMul(game_state.p2attackmul)
  end
end

-- Game state variable order is needed since the server must know the variable order
-- The game state variables are used for setters and getters. Sometimes both are needed sometimes not.
-- All getters are needed but setters are only needed for reward function
return {
    name = "ikemon-test",
    endpoint = "http://localhost:3000",
    description = "Sample RL config",
    reward_function = reward_function,
    frameStepInterval = 15,
    game_state_variables_order = { "p1life", "p1attackmul", "p2life", "p2attackmul" },
    game_state_variables = {
        p1life = {
            get = get_p1_life,
            },
        p1attackmul = { 
            get = get_p1_attackMul,
            set = set_p1_attackMul 
            },
        p2life = {
            get = get_p2_life, 
            },
        p2attackmul = { 
            get = get_p2_attackMul,
            set = set_p2_attackMul 
        },
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
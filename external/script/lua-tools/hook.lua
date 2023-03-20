--[[
Lua Hook System (v2)
Allows hooking additional code into existing functions, from within external
modules, without having to worry as much about your code being removed by
engine update.
* hook.run(list, ...): Runs all the functions within a certain list.
  It won't do anything if the list doesn't exist or is empty. ... is any
  number of arguments, which will be passed to every function in the list.
* hook.add(list, name, function): Adds a function to a hook list with a name.
  It will replace anything in the list with the same name.
* hook.stop(list, name): Removes a hook from a list, if it's not needed.

New in v2:
* hook.get(listName): Returns the hook list if it exists.
* hook.getOrCreate(listName): Returns the hook list, creating one if it
  doesn't exist.
* hook.once(list, callback): Adds a function to a hook list that runs once.
  One time hooks are run before ones that repeat.
* hook.removeOnce(list, callback): Removes a function from the hook's 
  once list.
* hook.on(list, callback): Adds a function to a hook list that will always run.
* hook.removeOn(list, callback): Removes a function from the hook's on list.
]]

-- Currently there are only few hooks available by default:
-- * loop: global.lua 'loop' function start (called by CommonLua)
-- * loop#[gamemode]: global.lua 'loop' function, limited to the gamemode
-- * main.f_commandLine: main.lua 'f_commandLine' function (before loading)
-- * main.f_default: main.lua 'f_default' function
-- * main.t_itemname: main.lua table entries (modes configuration)
-- * main.menu.loop: main.lua menu loop function (each submenu loop start)
-- * menu.menu.loop: menu.lua menu loop function (each submenu loop start)
-- * options.menu.loop: options.lua menu loop function (each submenu loop start)
-- * motif.setBaseTitleInfo: motif.lua default game mode items assignment
-- * motif.setBaseOptionInfo: motif.lua default option items assignment
-- * motif.setBaseMenuInfo: motif.lua default pause menu items assignment
-- * motif.setBaseTrainingInfo: motif.lua default training menu items assignment
-- * launchFight: start.lua 'launchFight' function (right before match starts)
-- * start.f_selectScreen: start.lua 'f_selectScreen' function (pre layerno=1)
-- * start.f_selectVersus: start.lua 'f_selectVersus' function (pre layerno=1)
-- * start.f_result: start.lua 'f_result' function (pre layerno=1)
-- * start.f_victory: start.lua 'f_victory' function (pre layerno=1)
-- * start.f_continue: start.lua 'f_continue' function (pre layerno=1)
-- * start.f_hiscore: start.lua 'f_hiscore' function (pre layerno=1)
-- * start.f_challenger: start.lua 'f_challenger' function (pre layerno=1)
-- More entry points may be added in future - let us know if your external
-- module needs to hook code in place where it's not allowed yet.

hook = {
	lists = {}
}

function hook.get(listName)
	return hook.lists[listName]
end
function hook.getOrCreate(listName)
	if hook.lists[listName] == nil then
		hook.lists[listName] = {
			on = {},
			once = {}
		}
	end
	return hook.lists[listName]
end

function hook.add(list, name, func)
	hook.getOrCreate(list).on[name] = func
end
function hook.stop(list, name)
	if hook.get(list) then
		hook.get(list).on[name] = nil
	end
end

function hook.on(list, func)
	table.insert(hook.getOrCreate(list).on, func)
end
function hook.once(list, func)
	table.insert(hook.getOrCreate(list).once, func)
end

function hook.removeOn(list, func)
    local curHook = hook.get(list)
	if curHook then
        local key = table.getKey(curHook.on, func)
        if key then
            if type(key) == 'number' then
                -- only do this with number keys, as table.remove
                -- renumbers indices
                table.remove(curHook.on, key)
            else
                curHook.on[key] = nil
            end
        end
    end
end

function hook.removeOnce(list, func)
    local curHook = hook.get(list)
	if curHook then
        local key = table.getKey(curHook.once, func)
        if key then
            if type(key) == 'number' then
                table.remove(curHook.once, key)
            else
                curHook.once[key] = nil
            end
        end
    end
end

function hook.run(list, ...)
	local curHook = hook.getOrCreate(list)
	for i, k in pairs(curHook.once) do
		k(...)
		curHook.once[i] = nil
	end
	for _, k in pairs(curHook.on) do
		k(...)
	end
end

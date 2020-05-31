--;===========================================================
--; HOTKEYS
--;===========================================================
--One-time load of the json routines
json = (loadfile 'external/script/dkjson.lua')()

-- Data loading from config.json
local file = io.open("save/config.json","r")
config = json.decode(file:read("*all"))
file:close()
-- This is done to get the AI level config
if getAllowDebugKeys() then
	--key, ctrl, alt, shift, function
	addHotkey('c', true, false, false, 'toggleClsnDraw()')
	addHotkey('d', true, false, false, 'toggleDebugDraw()')
	addHotkey('s', true, false, false, 'changeSpeed()')
	addHotkey('KP_PLUS', true, false, false, 'changeSpeed(1)')
	addHotkey('KP_MINUS', true, false, false, 'changeSpeed(-1)')
	addHotkey('l', true, false, false, 'toggleStatusDraw()')
	addHotkey('v', true, false, false, 'toggleVsync()')
	addHotkey('1', true, false, false, 'toggleAI(1)')
	addHotkey('1', true, true, false, 'togglePlayer(1)')
	addHotkey('2', true, false, false, 'toggleAI(2)')
	addHotkey('2', true, true, false, 'togglePlayer(2)')
	addHotkey('3', true, false, false, 'toggleAI(3)')
	addHotkey('3', true, true, false, 'togglePlayer(3)')
	addHotkey('4', true, false, false, 'toggleAI(4)')
	addHotkey('4', true, true, false, 'togglePlayer(4)')
	addHotkey('5', true, false, false, 'toggleAI(5)')
	addHotkey('5', true, true, false, 'togglePlayer(5)')
	addHotkey('6', true, false, false, 'toggleAI(6)')
	addHotkey('6', true, true, false, 'togglePlayer(6)')
	addHotkey('7', true, false, false, 'toggleAI(7)')
	addHotkey('7', true, true, false, 'togglePlayer(7)')
	addHotkey('8', true, false, false, 'toggleAI(8)')
	addHotkey('8', true, true, false, 'togglePlayer(8)')
	addHotkey('F1', false, false, false, 'kill(2);kill(4);kill(6);kill(8);markCheat(1);resetScore(1)')
	addHotkey('F1', true, false, false, 'kill(1);kill(3);kill(5);kill(7);markCheat(2);resetScore(2)')
	addHotkey('F2', false, false, false, 'kill(1,1);kill(2,1);kill(3,1);kill(4,1);kill(5,1);kill(6,1);kill(7,1);kill(8,1);markCheat(1);resetScore(1);markCheat(2);resetScore(2)')
	addHotkey('F2', true, false, false, 'kill(1,1);kill(3,1);kill(5,1);kill(7,1);markCheat(2);resetScore(2)')
	addHotkey('F2', false, false, true, 'kill(2,1);kill(4,1);kill(6,1);kill(8,1);markCheat(1);resetScore(1)')
	addHotkey('F3', false, false, false, 'powMax(1);powMax(2);markCheat(1);resetScore(1);markCheat(2);resetScore(2)')
	addHotkey('F3', true, false, true, 'toggleMaxPowerMode();markCheat(1);resetScore(1);markCheat(2);resetScore(2)')
	addHotkey('F4', false, false, false, 'roundReset()')
	addHotkey('F4', false, false, true, 'reload()')
	addHotkey('F5', false, false, false, 'setTime(0);markCheat(1);resetScore(1);markCheat(2);resetScore(2)')
	addHotkey('SPACE', false, false, false,
	'full(1);full(2);full(3);full(4);full(5);full(6);full(7);full(8);setTime(getRoundTime());markCheat(1);resetScore(1);markCheat(2);resetScore(2)')
	addHotkey('i', true, false, false, 'stand(1);stand(2);stand(3);stand(4);stand(5);stand(6);stand(7);stand(8)')
end
addHotkey('PAUSE', false, false, false, 'togglePause()')
addHotkey('SCROLLLOCK', false, false, false, 'step()')

speedMul = 1.0
speedAdd = 0
function changeSpeed(add)
	if add ~= nil then
		speedAdd = speedAdd + add / 100
	elseif speedMul >= 4 then
		speedMul = 0.25
	else
		speedMul = speedMul * 2.0
	end
	setAccel(math.max(0.01, speedMul + speedAdd))
end

function toggleAI(p)
	local oldid = id()
	if player(p) then
		if ailevel() > 0 then
			setAILevel(0)
		else
			setAILevel(config.Difficulty)
		end
		playerid(oldid)
	end
end

function kill(p, ...)
	local oldid = id()
	if player(p) then
		local n = ...
		if not n then n = 0 end
		setLife(n)
		playerid(oldid)
	end
end

function powMax(p)
	local oldid = id()
	if player(p) then
		setPower(powermax())
		setGuardPoints(guardpointsmax())
		setDizzyPoints(dizzypointsmax())
		playerid(oldid)
	end
end

function full(p)
	local oldid = id()
	if player(p) then
		setLife(lifemax())
		setPower(powermax())
		setGuardPoints(guardpointsmax())
		setDizzyPoints(dizzypointsmax())
		setRedLife(0)
		removeDizzy()
		playerid(oldid)
	end
end

function stand(p)
	local oldid = id()
	if player(p) then
		selfState(0)
		playerid(oldid)
	end
end

--;===========================================================
--; STATUS INFO
--;===========================================================
function statusInfo(p)
	local oldid = id()
	if not player(p) then return false end
	local ret = string.format(
		'P%d: %d; LIF:%5d; POW:%5d; RED:%5d; GRD:%5d; STN:%5d',
		playerno(), id(), life(), power(), redlife(), guardpoints(), dizzypoints()
	)
	playerid(oldid)
	return ret
end

--;===========================================================
--; PLAYER/HELPER INFO
--;===========================================================
function customState()
	if not incustomstate() then return "" end
	return " (in " .. stateownername() .. " " .. stateownerid() .. "'s state)"
end

function boolToInt(bool)
	if bool then return 1 end
	return 0
end

function engineInfo()
	return string.format('VSync: %d; Speed: %d%%', vsync(), gamespeed())
end

function playerInfo()
	return string.format('%s %d%s', name(), id(), customState())
end

function statsInfo()
	return string.format('HP: %d; ATK: %s; DEF: %s', 
		life(), tostring(attack()), tostring(defence()))
end

function actionInfo()
	return string.format(
		'ActionID: %d (P%d); SPR: %d,%d; ElemNo: %d/%d; Time: %d/%d (%d/%d)',
		anim(), animowner(), spritegroup(), spritenumber(), animelemno(0), animelemcount(), animelemtimesum(), animelemlength(), animtimesum(), animlength()
	)
end

function stateInfo()
	return string.format(
		'State No: %d (P%d); CTRL: %s; Type: %s; MoveType: %s; Physics: %s; Time: %d',
		stateno(), stateowner(), boolToInt(ctrl()), statetype(), movetype(), physics(), time()
	)
end

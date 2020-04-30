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
	addHotkey('l', true, false, false, 'toggleStatusDraw()')
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
	addHotkey('F3', false, false, false, 'powMax(1);powMax(2);markCheat(1);markCheat(2)')
	addHotkey('F3', true, false, true, 'toggleMaxPowerMode()')
	addHotkey('F4', false, false, false, 'roundReset()')
	addHotkey('F4', false, false, true, 'reload()')
	addHotkey('F5', false, false, false, 'setTime(0);markCheat(1);markCheat(2)')
	addHotkey('SPACE', false, false, false,
	'full(1);full(2);full(3);full(4);full(5);full(6);full(7);full(8);setTime(getRoundTime());markCheat(1);resetScore(1);markCheat(2);resetScore(2)')
	addHotkey('i', true, false, false, 'stand(1);stand(2);stand(3);stand(4);stand(5);stand(6);stand(7);stand(8)')
end
addHotkey('PAUSE', false, false, false, 'togglePause()')
addHotkey('SCROLLLOCK', false, false, false, 'step()')

speed = 1.0

function changeSpeed()
	if speed >= 4 then
		speed = 0.25
	else
		speed = speed * 2.0
	end
	setAccel(speed)
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

function info()
	puts(string.format(
		'name:%s state:%d>%d %s move:%s physics:%s',
		name(), prevstateno(), stateno(), statetype(), movetype(), physics())
	)
	puts(string.format(
		'anim:%d %d elem:%d %d pos:%.3f,%.3f vel:%.3f,%.3f',
		anim(), animtime(), animelemno(0), animelemtime(animelemno(0)),
		posX(), posY(), velX(), velY())
	)
end

function status(p)
	local oldid = id()
	if not player(p) then return false end
	ret = string.format(
		'ID:%d STA:%s%s%s%6d(%d) ANI:%6d(%d)%2d LIF:%5d POW:%5d GRD:%5d STN:%5d TIM:%d',
		id(), statetype(), movetype(), physics(), stateno(), stateOwner(),
		anim(), animOwner(), animelemno(0), life(), power(), guardpoints(), dizzypoints(), time()
	)
	playerid(oldid)
	return ret
end

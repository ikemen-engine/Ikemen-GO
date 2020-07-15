--;===========================================================
--; DEBUG HOTKEYS
--;===========================================================
--key, ctrl, alt, shift, pause, function
addHotkey('c', true, false, false, true, 'toggleClsnDraw()')
addHotkey('d', true, false, false, true, 'toggleDebugDraw()')
addHotkey('s', true, false, false, true, 'changeSpeed()')
addHotkey('KP_PLUS', true, false, false, true, 'changeSpeed(1)')
addHotkey('KP_MINUS', true, false, false, true, 'changeSpeed(-1)')
addHotkey('l', true, false, false, true, 'toggleStatusDraw()')
addHotkey('v', true, false, false, true, 'toggleVsync()')
addHotkey('1', true, false, false, true, 'toggleAI(1)')
addHotkey('1', true, true, false, true, 'togglePlayer(1)')
addHotkey('2', true, false, false, true, 'toggleAI(2)')
addHotkey('2', true, true, false, true, 'togglePlayer(2)')
addHotkey('3', true, false, false, true, 'toggleAI(3)')
addHotkey('3', true, true, false, true, 'togglePlayer(3)')
addHotkey('4', true, false, false, true, 'toggleAI(4)')
addHotkey('4', true, true, false, true, 'togglePlayer(4)')
addHotkey('5', true, false, false, true, 'toggleAI(5)')
addHotkey('5', true, true, false, true, 'togglePlayer(5)')
addHotkey('6', true, false, false, true, 'toggleAI(6)')
addHotkey('6', true, true, false, true, 'togglePlayer(6)')
addHotkey('7', true, false, false, true, 'toggleAI(7)')
addHotkey('7', true, true, false, true, 'togglePlayer(7)')
addHotkey('8', true, false, false, true, 'toggleAI(8)')
addHotkey('8', true, true, false, true, 'togglePlayer(8)')
addHotkey('F1', false, false, false, false, 'kill(2);kill(4);kill(6);kill(8);markCheat(1);resetScore(1)')
addHotkey('F1', true, false, false, false, 'kill(1);kill(3);kill(5);kill(7);markCheat(2);resetScore(2)')
addHotkey('F2', false, false, false, false, 'kill(1,1);kill(2,1);kill(3,1);kill(4,1);kill(5,1);kill(6,1);kill(7,1);kill(8,1);markCheat(1);resetScore(1);markCheat(2);resetScore(2)')
addHotkey('F2', true, false, false, false, 'kill(1,1);kill(3,1);kill(5,1);kill(7,1);markCheat(2);resetScore(2)')
addHotkey('F2', false, false, true, false, 'kill(2,1);kill(4,1);kill(6,1);kill(8,1);markCheat(1);resetScore(1)')
addHotkey('F3', false, false, false, false, 'powMax(1);powMax(2);markCheat(1);resetScore(1);markCheat(2);resetScore(2)')
addHotkey('F3', true, false, true, false, 'toggleMaxPowerMode();markCheat(1);resetScore(1);markCheat(2);resetScore(2)')
addHotkey('F4', false, false, false, false, 'roundReset()')
addHotkey('F4', false, false, true, false, 'reload()')
addHotkey('F5', false, false, false, false, 'setTime(0);markCheat(1);resetScore(1);markCheat(2);resetScore(2)')
addHotkey('SPACE', false, false, false, false, 'full(1);full(2);full(3);full(4);full(5);full(6);full(7);full(8);setTime(getRoundTime());markCheat(1);resetScore(1);markCheat(2);resetScore(2)')
addHotkey('i', true, false, false, true, 'stand(1);stand(2);stand(3);stand(4);stand(5);stand(6);stand(7);stand(8)')
addHotkey('PAUSE', false, false, false, true, 'togglePause()')
addHotkey('PAUSE', true, false, false, true, 'step()')
addHotkey('SCROLLLOCK', false, false, false, true, 'step()')

local speedMul = 1
local speedAdd = 0
function changeSpeed(add)
	if add ~= nil then
		speedAdd = speedAdd + add / 100
	elseif speedMul >= 4 then
		speedMul = 0.25
	else
		speedMul = speedMul * 2
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
--; DEBUG STATUS INFO
--;===========================================================
function statusInfo(p)
	local oldid = id()
	if not player(p) then return false end
	local ret = string.format(
		'P%d: %d; LIF:%4d; POW:%4d; ATK:%4d; DEF:%4d; RED:%4d; GRD:%4d; STN:%4d',
		playerno(), id(), life(), power(), attack(), defence(), redlife(), guardpoints(), dizzypoints()
	)
	playerid(oldid)
	return ret
end

loadDebugStatus('statusInfo')

--;===========================================================
--; DEBUG PLAYER/HELPER INFO
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

loadDebugInfo({'engineInfo', 'playerInfo', 'actionInfo', 'stateInfo'})

--;===========================================================
--; MATCH LOOP
--;===========================================================
local endFlag = false

--function called during match via config.json CommonLua
function loop()
	if start == nil then --match started via command line
		togglePostMatch(false)
		return
	end
	--music
	start.f_stageMusic()
	--match start
	if roundstart() then
		if roundno() == 1 then
			speedMul = 1
			speedAdd = 0
			start.victoryInit = false
			start.resultInit = false
			start.continueInit = false
			endFlag = false
		end
	end
	--match end
	if matchover() and roundover() then
		if not endFlag then
			resetMatchData()
			endFlag = true
		end
		--victory screen
		if start.f_victory() then
			return
		--result screen
		elseif start.f_result() then
			return
		--continue screen
		elseif start.f_continue() then
			return
		end
		togglePostMatch(false)
	end
	--escMenu
	if main.escMenu then
		playerBufReset()
		menu.run()
	else
		main.f_cmdInput()
		if esc() or main.f_input(main.t_players, {'m'}) then
			if gamemode('') or gamemode('demo') or gamemode('randomtest') then
				endMatch()
			else
				menu.init()
			end
		end
	end
end

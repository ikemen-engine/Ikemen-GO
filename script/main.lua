
setRoundTime(999 * 6)--frames
setLifeMul(1.0)
setTeam1VS2Life(1.0)
setTurnsRecoveryRate(1.0 / 300.0)

setZoom(false)
setZoomMin(0.25)
setZoomMax(1.0)
setZoomSpeed(1.0)

loadLifebar('data/gms_lifebar/fight.def')
loadDebugFont('data/gms_lifebar/font2.fnt')
setDebugScript('script/debug.lua')

selectColumns = 10


require('script.randomtest')

function addWithRefresh(addFn, text)
  local nextRefresh = os.clock() + 0.02
  for i, c
    in ipairs(script.randomtest.strsplit('\n',
                                         text:gsub('^%s*(.-)%s*$', '%1')))
  do
    addFn(c)
    if os.clock() >= nextRefresh then
      refresh()
      nextRefresh = os.clock() + 0.02
    end
  end
end

orgAddChar = addChar
orgAddStage = addStage

function addChar(text)
  addWithRefresh(orgAddChar, text)
end

function addStage(text)
  addWithRefresh(orgAddStage, text)
end

assert(loadfile('script/select.lua'))()


math.randomseed(os.time())

------------------------------------------------------------
sysSff = sffNew('script/system.sff')
sysSnd = sndNew('script/system.snd')
jgFnt = fontNew('data/gms_lifebar/font2.fnt')

bgm = ''
playBGM(bgm)

------------------------------------------------------------
function setCommand(c)
  commandAdd(c, 'u', '$U')
  commandAdd(c, 'd', '$D')
  commandAdd(c, 'l', '$B')
  commandAdd(c, 'r', '$F')
  commandAdd(c, 'a', 'a')
  commandAdd(c, 'b', 'b')
  commandAdd(c, 'c', 'c')
  commandAdd(c, 'x', 'x')
  commandAdd(c, 'y', 'y')
  commandAdd(c, 'z', 'z')
  commandAdd(c, 's', 's')
  commandAdd(c, 'holds', '/s')
  commandAdd(c, 'su', '/s, U')
  commandAdd(c, 'sd', '/s, D')
end

p1Cmd = commandNew()
setCommand(p1Cmd)

p2Cmd = commandNew()
setCommand(p2Cmd)

------------------------------------------------------------
selectRows = math.floor(selectColumns * 2 / 5.0)

setRandomSpr(sysSff, 151, 0, 5.0/selectColumns, 5.0/selectColumns)
setSelColRow(selectColumns, selectRows)
setSelCellSize(29*5.0/selectColumns, 29*5.0/selectColumns)
setSelCellScale(5.0/selectColumns, 5.0/selectColumns)

function init()
  p1TeamMode = 0
  p1NumTurns  = 2
  p1SelOffset = 0
  p1SelX = 0
  p1SelY = 0
  p1Portrait = nil

  p2TeamMode = 0
  p2NumTurns  = 2
  p2SelOffset = 0
  p2SelX = 0
  p2SelY = 0
  p2Portrait = nil

  stageNo = 0
  setStage(0)
end

init()

function noTask()
end


function animPosDraw(a, x, y)
  animSetPos(a, x, y)
  animDraw(a)
end

function textImgPosDraw(ti, x, y)
  textImgSetPos(ti, x, y)
  textImgDraw(ti)
end

function createTextImg(font, bank, aline, text, x, y)
  local ti = textImgNew()
  textImgSetFont(ti, font)
  textImgSetBank(ti, bank)
  textImgSetAlign(ti, aline)
  textImgSetText(ti, text)
  textImgSetPos(ti, x, y)
  return ti
end

function btnPalNo(cmd)
  local s = 0
  if commandGetState(cmd, 'holds') then s = 6 end
  if commandGetState(cmd, 'a') then return 1 + s end
  if commandGetState(cmd, 'b') then return 2 + s end
  if commandGetState(cmd, 'c') then return 3 + s end
  if commandGetState(cmd, 'x') then return 4 + s end
  if commandGetState(cmd, 'y') then return 5 + s end
  if commandGetState(cmd, 'z') then return 6 + s end
  return 0
end

------------------------------------------------------------
p1SelTmTxt = createTextImg(jgFnt, 0, 1, 'Team Mode', 20, 30)
p1SingleTxt = createTextImg(jgFnt, 0, 1, 'Single', 20, 50)
p1SimulTxt = createTextImg(jgFnt, 0, 1, 'Simul', 20, 65)
p1TurnsTxt = createTextImg(jgFnt, 0, 1, 'Turns', 20, 80)

p1TmCursor = animNew(sysSff, [[
180,0, 0,0, -1
]])

p1TmIcon = animNew(sysSff, [[
181,0, 0,0, -1
]])

function p1TmSub()
  if commandGetState(p1Cmd, 'u') then
    sndPlay(sysSnd, 100, 0)
    p1TeamMode = p1TeamMode - 1
    if p1TeamMode < 0 then p1TeamMode = 2 end
  elseif commandGetState(p1Cmd, 'd') then
    sndPlay(sysSnd, 100, 0)
    p1TeamMode = p1TeamMode + 1
    if p1TeamMode > 2 then p1TeamMode = 0 end
  elseif p1TeamMode == 2 then
    if commandGetState(p1Cmd, 'l') then
      sndPlay(sysSnd, 100, 0)
      p1NumTurns = p1NumTurns - 1
      if p1NumTurns < 1 then p1NumTurns = 1 end
    elseif commandGetState(p1Cmd, 'r') then
      sndPlay(sysSnd, 100, 0)
      p1NumTurns = p1NumTurns + 1
      if p1NumTurns > 4 then p1NumTurns = 4 end
    end
  end
  textImgDraw(p1SelTmTxt)
  textImgDraw(p1SingleTxt)
  textImgDraw(p1SimulTxt)
  textImgDraw(p1TurnsTxt)
  animUpdate(p1TmIcon)
  animPosDraw(p1TmIcon, 80, 66)
  animPosDraw(p1TmIcon, 86, 66)
  for i = 1, p1NumTurns do
    animPosDraw(p1TmIcon, 74 + i*6, 81)
  end
  animUpdate(p1TmCursor)
  animPosDraw(p1TmCursor, 10, 47 + p1TeamMode*15)
  if btnPalNo(p1Cmd) > 0 then
    sndPlay(sysSnd, 100, 1)
    setTeamMode(1, p1TeamMode, p1NumTurns)
    p1Selected = {}
    p1SelEnd = false
    p1Task = p1SelSub
  end
end


------------------------------------------------------------
p2SelTmTxt = createTextImg(jgFnt, 0, -1, 'Team Mode', 300, 30)
p2SingleTxt = createTextImg(jgFnt, 0, -1, 'Single', 300, 50)
p2SimulTxt = createTextImg(jgFnt, 0, -1, 'Simul', 300, 65)
p2TurnsTxt = createTextImg(jgFnt, 0, -1, 'Turns', 300, 80)

p2TmCursor = animNew(sysSff, [[
190,0, 0,0, -1
]])

p2TmIcon = animNew(sysSff, [[
191,0, 0,0, -1
]])

function p2TmSub()
  if commandGetState(p2Cmd, 'u') then
    sndPlay(sysSnd, 100, 0)
    p2TeamMode = p2TeamMode - 1
    if p2TeamMode < 0 then p2TeamMode = 2 end
  elseif commandGetState(p2Cmd, 'd') then
    sndPlay(sysSnd, 100, 0)
    p2TeamMode = p2TeamMode + 1
    if p2TeamMode > 2 then p2TeamMode = 0 end
  elseif p2TeamMode == 2 then
    if commandGetState(p2Cmd, 'r') then
      sndPlay(sysSnd, 100, 0)
      p2NumTurns = p2NumTurns - 1
      if p2NumTurns < 1 then p2NumTurns = 1 end
    elseif commandGetState(p2Cmd, 'l') then
      sndPlay(sysSnd, 100, 0)
      p2NumTurns = p2NumTurns + 1
      if p2NumTurns > 4 then p2NumTurns = 4 end
    end
  end
  textImgDraw(p2SelTmTxt)
  textImgDraw(p2SingleTxt)
  textImgDraw(p2SimulTxt)
  textImgDraw(p2TurnsTxt)
  animUpdate(p2TmIcon)
  animPosDraw(p2TmIcon, 240, 66)
  animPosDraw(p2TmIcon, 234, 66)
  for i = 1, p2NumTurns do
    animPosDraw(p2TmIcon, 246 - i*6, 81)
  end
  animUpdate(p2TmCursor)
  animPosDraw(p2TmCursor, 310, 47 + p2TeamMode*15)
  if btnPalNo(p2Cmd) > 0 then
    sndPlay(sysSnd, 100, 1)
    setTeamMode(2, p2TeamMode, p2NumTurns)
    p2Selected = {}
    p2SelEnd = false
    p2Task = p2SelSub
  end
end


------------------------------------------------------------
p1Cursor = animNew(sysSff, [[
160,0, 0,0, -1
]])
animSetScale(p1Cursor, 5.0/selectColumns, 5.0/selectColumns)

p1NameTxt = createTextImg(jgFnt, 0, 1, '', 0, 0)
textImgSetScale(p1NameTxt, 0.5, 0.5)

function p1DrawSelectName()
  local y = 162
  for i = 1, #p1Selected do
    textImgSetText(p1NameTxt, getCharName(p1Selected[i]))
    textImgPosDraw(p1NameTxt, 10, y)
    y = y + 7
  end
  return y
end

function p1SelSub()
  local n = p1SelOffset + p1SelX + selectColumns*p1SelY
  p1Portrait = n
  local y = p1DrawSelectName()
  if not p1SelEnd then
    if commandGetState(p1Cmd, 'su') then
      sndPlay(sysSnd, 100, 0)
      p1SelY = p1SelY - 20
    elseif commandGetState(p1Cmd, 'sd') then
      sndPlay(sysSnd, 100, 0)
      p1SelY = p1SelY + 20
    elseif commandGetState(p1Cmd, 'u') then
      sndPlay(sysSnd, 100, 0)
      p1SelY = p1SelY - 1
    elseif commandGetState(p1Cmd, 'd') then
      sndPlay(sysSnd, 100, 0)
      p1SelY = p1SelY + 1
    elseif commandGetState(p1Cmd, 'l') then
      sndPlay(sysSnd, 100, 0)
      p1SelX = p1SelX - 1
    elseif commandGetState(p1Cmd, 'r') then
      sndPlay(sysSnd, 100, 0)
      p1SelX = p1SelX + 1
    end
    if p1SelY < 0 then
      p1SelOffset = p1SelOffset + selectColumns*p1SelY
      p1SelY = 0
    elseif p1SelY >= selectRows then
      p1SelOffset = p1SelOffset + selectColumns*(p1SelY - (selectRows - 1))
      p1SelY = selectRows - 1
    end
    if p1SelX < 0 then
      p1SelX = selectColumns - 1
    elseif p1SelX >= selectColumns then
      p1SelX = 0
    end
    animUpdate(p1Cursor)
    animPosDraw(
      p1Cursor, 10 + 29*p1SelX*5.0/selectColumns,
      170 + 29*p1SelY*5.0/selectColumns)
    textImgSetText(p1NameTxt, getCharName(n))
    textImgPosDraw(p1NameTxt, 10, y)
    local selval = selectChar(1, n, btnPalNo(p1Cmd))
    if selval > 0 then
      sndPlay(sysSnd, 100, 1)
      p1Selected[#p1Selected+1] = n
    end
    if selval == 2 then
      p1SelEnd = true
      if p2In == 1 then
        p2Task = p2TmSub
        commandBufReset(p2Cmd)
      end
    end
  end
end


------------------------------------------------------------
p2Cursor = animNew(sysSff, [[
170,0, 0,0, -1
]])
animSetScale(p2Cursor, 5.0/selectColumns, 5.0/selectColumns)

p2NameTxt = createTextImg(jgFnt, 0, -1, '', 0, 0)
textImgSetScale(p2NameTxt, 0.5, 0.5)

function p2DrawSelectName()
  local y = 162
  for i = 1, #p2Selected do
    textImgSetText(p2NameTxt, getCharName(p2Selected[i]))
    textImgPosDraw(p2NameTxt, 310, y)
    y = y + 7
  end
  return y
end

function p2SelSub()
  local n = p2SelOffset + p2SelX + selectColumns*p2SelY
  p2Portrait = n
  local y = p2DrawSelectName()
  if not p2SelEnd then
    if commandGetState(p2Cmd, 'su') then
      sndPlay(sysSnd, 100, 0)
      p2SelY = p2SelY - 20
    elseif commandGetState(p2Cmd, 'sd') then
      sndPlay(sysSnd, 100, 0)
      p2SelY = p2SelY + 20
    elseif commandGetState(p2Cmd, 'u') then
      sndPlay(sysSnd, 100, 0)
      p2SelY = p2SelY - 1
    elseif commandGetState(p2Cmd, 'd') then
      sndPlay(sysSnd, 100, 0)
      p2SelY = p2SelY + 1
    elseif commandGetState(p2Cmd, 'l') then
      sndPlay(sysSnd, 100, 0)
      p2SelX = p2SelX - 1
    elseif commandGetState(p2Cmd, 'r') then
      sndPlay(sysSnd, 100, 0)
      p2SelX = p2SelX + 1
    end
    if p2SelY < 0 then
      p2SelOffset = p2SelOffset + selectColumns*p2SelY
      p2SelY = 0
    elseif p2SelY >= selectRows then
      p2SelOffset = p2SelOffset + selectColumns*(p2SelY - (selectRows - 1))
      p2SelY = selectRows - 1
    end
    if p2SelX < 0 then
      p2SelX = selectColumns - 1
    elseif p2SelX >= selectColumns then
      p2SelX = 0
    end
    animUpdate(p2Cursor)
    animPosDraw(
      p2Cursor, 169 + 29*p2SelX*5.0/selectColumns,
      170 + 29*p2SelY*5.0/selectColumns)
    textImgSetText(p2NameTxt, getCharName(n))
    textImgPosDraw(p2NameTxt, 310, y)
    local selval = selectChar(2, n, btnPalNo(p2Cmd))
    if selval > 0 then
      sndPlay(sysSnd, 100, 1)
      p2Selected[#p2Selected+1] = n
    end
    if selval == 2 then
      p2SelEnd = true
      if p1In == 2 then
        p1Task = p1TmSub
        commandBufReset(p1Cmd)
      end
    end
  end
end


------------------------------------------------------------
selStageTxt = createTextImg(jgFnt, 0, 0, '', 160, 237)
textImgSetScale(selStageTxt, 0.5, 0.5)

function selStageSub()
  if commandGetState(p1Cmd, 'l') then
    sndPlay(sysSnd, 100, 0)
    stageNo = setStage(stageNo - 1)
  elseif commandGetState(p1Cmd, 'r') then
    sndPlay(sysSnd, 100, 0)
    stageNo = setStage(stageNo + 1)
  elseif commandGetState(p1Cmd, 'u') then
    sndPlay(sysSnd, 100, 0)
    stageNo = setStage(stageNo - 10)
  elseif commandGetState(p1Cmd, 'd') then
    sndPlay(sysSnd, 100, 0)
    stageNo = setStage(stageNo + 10)
  end
  textImgSetText(
    selStageTxt, 'Stage ' .. stageNo .. ': ' .. getStageName(stageNo))
  textImgDraw(selStageTxt)
  if btnPalNo(p1Cmd) > 0 then
    selectStage(stageNo)
    selMode = false
  end
end


------------------------------------------------------------
selBG = animNew(sysSff, [[
100,0, 0,0, -1
]])
animSetTile(selBG, 1, 1)
animSetColorKey(selBG, -1)
animSetScale(selBG, 0.5, 0.5)

selBox = animNew(sysSff, [[
100,1, 0,0, -1
]])
animSetTile(selBox, 1, 0)
animSetColorKey(selBox, -1)
animSetAlpha(selBox, 1, 255)
animSetPos(selBox, 0, 166)
animSetWindow(selBox, 5, 0, 151, 240)

selBox2 = animNew(sysSff, [[
100,1, 0,0, -1
]])
animSetTile(selBox2, 1, 0)
animSetColorKey(selBox2, -1)
animSetAlpha(selBox2, 1, 255)
animSetPos(selBox2, 0, 166)
animSetWindow(selBox2, 164, 0, 151, 240)

function bgSub()
  animAddPos(selBG, 1, 1)
  animUpdate(selBG)
  animDraw(selBG)
  animAddPos(selBox, 1, 0)
  animUpdate(selBox)
  animDraw(selBox)
  animAddPos(selBox2, 1, 0)
  animUpdate(selBox2)
  animDraw(selBox2)
end


------------------------------------------------------------
watchMode = createTextImg(jgFnt, 0, 1, 'Watch Mode', 100, 80)
p1VsComTxt = createTextImg(jgFnt, 0, 1, '1P vs. Com', 100, 100)
p1VsP2 = createTextImg(jgFnt, 0, 1, '1P vs. 2P', 100, 120)
netplay = createTextImg(jgFnt, 0, 1, 'Netplay', 100, 140)
portChange = createTextImg(jgFnt, 0, 1, '', 100, 160)
replay = createTextImg(jgFnt, 0, 1, 'Replay', 100, 180)
comVsP1 = createTextImg(jgFnt, 0, 1, 'Com vs. 1P', 100, 200)
autoRandomTest = createTextImg(jgFnt, 0, 1, 'Auto Random Test', 100, 220)

connecting = createTextImg(jgFnt, 0, 1, '', 10, 140)
loading = createTextImg(jgFnt, 0, 1, 'Loading...', 100, 210)

inputdia = inputDialogNew()

function cmdInput()
  commandInput(p1Cmd, p1In)
  commandInput(p2Cmd, p2In)
end

function main()
  while true do
    p1Selected = {}
    p1SelEnd = false
    p1Portrait = nil

    p2Selected = {}
    p2SelEnd = false
    p2Portrait = nil

    if gameMode == 6 then
      p1Task = noTask
      p2Task = p2TmSub
    else
      p1Task = p1TmSub
      p2Task = noTask
      if gameMode > 1 then p2Task = p2TmSub end
    end

    refresh()

    commandBufReset(p1Cmd)
    commandBufReset(p2Cmd)

    selMode = true
    selectStart()

    ------------------------------------------------------------
    --メインループ
    ------------------------------------------------------------
    while selMode do
      if esc() then return end
      bgSub()
      if p1Portrait then drawPortrait(p1Portrait, 18, 13, 1, 1) end
      if p2Portrait then drawPortrait(p2Portrait, 302, 13, -1, 1) end
      drawFace(10, 170, p1SelOffset)
      drawFace(169, 170, p2SelOffset)
      if p1SelEnd and p2SelEnd then selStageSub() end
      p2Task()
      p1Task()
      cmdInput()
      refresh()
    end
    for i = 1, 20 do
      animDraw(selBG)
      local k = 0
      for j = 1, #p1Selected do
        local scl = 10000.0 / (10000.0 - k*i)
        local tmp = i*k / 20
        drawPortrait(p1Selected[j], 18 - tmp, 13 + tmp/3, scl, scl)
        k = k + 48
      end
      k = 0
      for j = 1, #p2Selected do
        local scl = 10000.0 / (10000.0 - k*i)
        local tmp = i*k / 20
        drawPortrait(p2Selected[j], 302 + tmp, 13 + tmp/3, -scl, scl)
        k = k + 48
      end
      p1DrawSelectName()
      p2DrawSelectName()
      textImgDraw(loading)
      refresh()
    end
    game()
    playBGM(bgm)
  end
end

function modeSel()
  while true do
    exitNetPlay()
    exitReplay()

    gameMode = 0
    p1In = 1
    p2In = 1

    for i = 1, 8 do
      setCom(i, 8)
    end
    setAutoLevel(false)
    setMatchNo(1)
    setHomeTeam(1)
    resetRemapInput()

    textImgSetText(portChange, 'Port Change(' .. getListenPort() .. ')')

    refresh()
    commandBufReset(p1Cmd)

    while btnPalNo(p1Cmd) <= 0 do
      if commandGetState(p1Cmd, 'u') then
        sndPlay(sysSnd, 100, 0)
        gameMode = gameMode - 1
      elseif commandGetState(p1Cmd, 'd') then
        sndPlay(sysSnd, 100, 0)
        gameMode = gameMode + 1
      end
      if gameMode < 0 then
        gameMode = 7
      elseif gameMode > 7 then
        gameMode = 0
      end
      textImgDraw(watchMode)
      textImgDraw(p1VsComTxt)
      textImgDraw(p1VsP2)
      textImgDraw(netplay)
      textImgDraw(portChange)
      textImgDraw(replay)
      textImgDraw(comVsP1)
      textImgDraw(autoRandomTest)
      animUpdate(p1TmCursor)
      animPosDraw(p1TmCursor, 95, 77 + 20*gameMode)
      cmdInput()
      refresh()
    end
    sndPlay(sysSnd, 100, 1)

    local cancel = false

    if gameMode == 0 then
    elseif gameMode == 1 then
      setCom(1, 0)
    elseif gameMode == 2 then
      p2In = 2
      setCom(1, 0)
      setCom(2, 0)
    elseif gameMode == 3 then
      p2In = 2
      setCom(1, 0)
      setCom(2, 0)
      inputDialogPopup(inputdia, 'Input Server')
      while not inputDialogIsDone(inputdia) do
        refresh()
      end
      textImgSetText(
        connecting,
        'Now connecting.. ' .. inputDialogGetStr(inputdia)
        .. ' ' .. getListenPort())
      enterNetPlay(inputDialogGetStr(inputdia))
      while not connected() do
        if esc() then
          cancel = true
          break
        end
        textImgDraw(connecting)
        refresh()
      end
      if not cancel then
        init()
        synchronize()
        math.randomseed(sszRandom())
      end
    elseif gameMode == 4 then
      inputDialogPopup(inputdia, 'Input Port')
      while not inputDialogIsDone(inputdia) do
        refresh()
      end
      setListenPort(inputDialogGetStr(inputdia))
      cancel = true
    elseif gameMode == 5 then
      p2In = 2
      setCom(1, 0)
      setCom(2, 0)
      enterReplay('replay/netplay.replay')
      init()
      synchronize()
      math.randomseed(sszRandom())
    elseif gameMode == 6 then
      remapInput(1, 2)
      remapInput(2, 1)
      p1In = 2
      p2In = 2
      setCom(2, 0)
    elseif gameMode == 7 then
      script.randomtest.run()
      cancel = true
    end
    if not cancel then
      main()
    end
  end
end

modeSel()


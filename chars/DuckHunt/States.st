; ____________________________________
;| Duck Hunt by Phantom.of.the.Server |
; ¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯
;========================< FLYING >=============================
[Statedef 1000]
type = A
physics = N
movetype = I
anim = 1000
ctrl = 0
juggle = 16
sprpriority = -4

[State 1000, Burn]
type = hitoverride
trigger1 = 1
attr = SCA, AP
stateno = 1020
time = -1

[State 1000, timevar]
type = varadd
trigger1 = time
var(0) = 1

[State 1000, flap]
type = playsnd
trigger1 = numhelper(1000) <= 6
lowpriority = 1
value = 0, 0
channel = 0
loop = 1
[State 1000, flap]
type = sndpan
trigger1 = 1
channel = 0
pan = 0

[State 1000, colorvar]
type = varset
trigger1 = !time && !prevstateno
var(1) = ifelse(random < 500, 0, ifelse(random < 666, 1, 2))

[State 1000, anglevar]
type = varset
trigger1 = !time
fvar(0) = -(pi / 6.0) -(random / 250.0) * (pi / 6.0)

[State 1000, sts]
type = statetypeset
trigger1 = pos y >= -60
statetype = S
[State 1000, sts]
type = statetypeset
trigger1 = pos y < -60
statetype = A

[State 1000, Velvar]
type = varset
trigger1 = !time && var(1) = 0
fvar(1) = (1 + (random / 1000.0)) * (root, fvar(30))
[State 1000, Velvar]
type = varset
trigger1 = !time && var(1) = 1
fvar(1) = (1.5 + (random / 666.7)) * (root, fvar(30))
[State 1000, Velvar]
type = varset
trigger1 = !time && var(1) = 2
fvar(1) = (2 + (random / 500.0)) * (root, fvar(30))

[State 1000, Velbounce]
type = varset
trigger1 = var(0) <= 300 && vel x > 0
trigger1 = (facing = 1 && pos x > 144) || (facing = -1 && pos x < -144)
trigger2 = movecontact = 1 || movereversed = 1
fvar(0) = pi -fvar(0) -(pi / 9.0) + (random / 500.0) * (pi / 9.0)
[State 1000, Velbounce]
type = varset
trigger1 = var(0) <= 300 && vel y < 0 && screenpos y < 16 
trigger2 = vel y > 0 && pos y > -21
trigger3 = (movecontact = 1 || movereversed = 1)
trigger3 = (vel y > 0 && pos y >= -132) || (vel y < 0 && pos y < -132)
fvar(0) = -fvar(0) -(pi / 9.0) + (random / 500.0) * (pi / 9.0)

[State 1000, Vel]
type = velset
trigger1 = 1
x = cos(fvar(0)) * fvar(1) * ifelse((pos x != [ -160, 160]), 2, 1)
y = sin(fvar(0)) * fvar(1) * ifelse((screenpos y != [0, 240]), 2, 1)

[State 1000, turn]
type = turn
trigger1 = vel x < 0
trigger1 = fvar(0) := (pi -fvar(0))

[State 1000, Anim]
type = changeanim
trigger1 = anim != (1000 + 100 * var(1)) && abs(vel x / vel y) <= 0.33 && vel y <= 0
value = 1000 + 100 * var(1)
elem = animelemno(0)
[State 1000, Anim]
type = changeanim
trigger1 = anim != (1001 + 100 * var(1)) && (abs(vel x / vel y) = [.33, 0.66])
trigger2 = anim != (1001 + 100 * var(1)) && abs(vel x / vel y) <= 0.33 && vel y > 0
value = 1001 + 100 * var(1)
elem = animelemno(0)
[State 1000, Anim]
type = changeanim
trigger1 = anim != (1002 + 100 * var(1)) && abs(vel x / vel y) > 0.66
value = 1002 + 100 * var(1)
elem = animelemno(0)

;--- Attacking ---
[State 1000, Attack]
type = statetypeset
trigger1 = root, var(40)
movetype = A

[State 1000, Attack]
type = hitdef
trigger1 = time >= 60 && root, var(40)
trigger1 = !((movecontact + movereversed) % 60)
attr = A, NA
hitflag = MA-
guardflag = MA
guarddist = 320
priority = 1, Hit
damage = 0, 0
getpower = 0, 0
givepower = 36, 36
pausetime = 0, 12
animtype = Light
air.animtype = Light
fall.animtype = Light
ground.type = High
air.type = High
ground.hittime = 10
ground.slidetime = 10
guard.hittime = 10
air.hittime = 10
airguard.ctrltime = 10
down.hittime = 30
ground.velocity = 0, 0
ground.cornerpush.veloff = 0
guard.velocity = 0
guard.cornerpush.veloff = 0
air.velocity = 0, -4
air.cornerpush.veloff = 0
airguard.velocity = 0, -4
airguard.cornerpush.veloff = 0
down.velocity = 0, 0
down.cornerpush.veloff = 0
yaccel = 0.5
fall = 0
fall.recover = 1
sparkno = 0
guard.sparkno = 40
sparkxy = -10, 0
hitsound = 5, (random % 2)
guardsound = 6, 0
sprpriority = -4

[State 1000, Attack]
type = hitdef
trigger1 = root, var(40)
trigger1 = p2movetype = H || (p2stateno = [150, 155]) || (p2stateno = [5020, 5040])
attr = A, NA
hitflag = 
sparkxy = -10, 0
sprpriority = -4

[State 1000, Attack]
type = targetfacing
trigger1 = numtarget
trigger1 = (target, statetype = A) && !(target, hitshakeover) && !(target, time)
value = 1
ignorehitpause = 1
;--- ------------

[State 1000, Shot]
type = changestate
trigger1 = (root, stateno = 5150) && (root, time)
trigger2 = (root, stateno = 2)
value = 1010
[State 1000, Fly Away]
type = changestate
triggerall = var(0) > 300
trigger1 = (pos x != [ -176, 176]) || screenpos y < -16
value = 1005


;========================< FLY AWAY >=============================
[Statedef 1005]
type = A
physics = N
movetype = I
ctrl = 0
sprpriority = -4

[State 1005, score]
type = parentvaradd
trigger1 = !time
var(8) = 1

[State 1005, dog]
type = helper
trigger1 = !time
trigger1 = !numhelper(400) && roundstate = 2
helpertype = normal
stateno = 410
ID = 400
name = "Dog"
postype = left
pos = 160 + ifelse(pos x > 144, 144, ifelse(pos x < -144, -144, ceil(pos x))), 0
size.shadowoffset = -7
ownpal = 1

[State 1005, snd]
type = stopsnd
trigger1 = 1
channel = 0
[State 1005, End]
type = destroyself
trigger1 = 1


;========================< HIT >=============================
[Statedef 1010]
type = A
physics = N
movetype = H
anim = 1010 + 100 * var(1)
ctrl = 0
velset = 0, 0
sprpriority = -4
[State 1010, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
[State 1010, Pos]
type = posset
trigger1 = pos y >= -16
y = -16

[State 1010, score]
type = varset
trigger1 = !time
var(2) = ceil(ifelse(var(1) = 2, 20, ifelse(var(1) = 1, 15, 10)) * (lifemax -life) * abs(root, fvar(30)))
[State 1010, score]
type = varset
trigger1 = !time && !var(2)
var(2) = ceil(ifelse(var(1) = 2, 20, ifelse(var(1) = 1, 15, 10)) * gethitvar(damage) * abs(root, fvar(30)))

[State 1010, score]
type = parentvaradd
trigger1 = !time
var(6) = var(2)
[State 1010, score]
type = parentvaradd
trigger1 = !time
var(7) = 1

[State 1010, points]
type = helper
trigger1 = time = 22
helpertype = normal
stateno = 650
ID = 650
name = "Points"
postype = p1
pos = 0, 0
facing = facing
ownpal = 1

[State 1010, snd]
type = stopsnd
trigger1 = !time
channel = 0
[State 1010, snd]
type = stopsnd
trigger1 = !time
channel = 1
[State 1010, Stop]
type = velset
trigger1 = 1
x = 0
y = 0
[State 1010, Stop]
type = posfreeze
trigger1 = 1
value = 1
[State 1010, End]
type = changestate
trigger1 = time >= 24
value = 1011


;--- Falling ---
[Statedef 1011]
type = A
physics = N
movetype = I
anim = 1011 + 100 * var(1)
ctrl = 0
velset = 0, 2
sprpriority = -4
[State 1011, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
[State 1011, snd]
type = playsnd
trigger1 = !time
value = 0, 2
channel = 0
[State 1011, End]
type = changestate
trigger1 = pos y >= -21
value = 1012


;--- Hit ground ---
[Statedef 1012]
type = A
physics = N
movetype = I
ctrl = 0
velset = 0, 0
sprpriority = -4
[State 1012, inv]
type = assertspecial
trigger1 = 1
flag = invisible

[State 1012, dog]
type = helper
trigger1 = !time
trigger1 = !numhelper(400) && roundstate = 2
helpertype = normal
stateno = 450
ID = 400
name = "Dog"
postype = left
pos = 160 + ifelse(pos x > 144, 144, ifelse(pos x < -144, -144, ceil(pos x))), 0
size.shadowoffset = -7
ownpal = 1

[State 1012, snd]
type = playsnd
trigger1 = !time
value = 0, 3
channel = 0
[State 1012, End]
type = destroyself
trigger1 = time


;========================< FEATHERLESS >=============================
[Statedef 1020]
type = A
physics = N
movetype = H
anim = 1020 + 100 * var(1)
ctrl = 0
velset = 0, 0
sprpriority = -4
[State 1020, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
[State 1020, Pos]
type = posset
trigger1 = pos y >= -16
y = -16

[State 1020, score]
type = varset
trigger1 = !time
var(2) = ceil(ifelse(var(1) = 2, 20, ifelse(var(1) = 1, 15, 10)) * (lifemax -life) * abs(root, fvar(30)))
[State 1020, score]
type = varset
trigger1 = !time && !var(2)
var(2) = ceil(ifelse(var(1) = 2, 20, ifelse(var(1) = 1, 15, 10)) * gethitvar(damage) * abs(root, fvar(30)))

[State 1020, score]
type = parentvaradd
trigger1 = !time
var(6) = var(2)
[State 1020, score]
type = parentvaradd
trigger1 = !time
var(7) = 1

[State 1020, points]
type = helper
trigger1 = time = 22
helpertype = normal
stateno = 650
ID = 650
name = "Points"
postype = p1
pos = 0, 0
facing = facing
ownpal = 1

[State 1020, snd]
type = stopsnd
trigger1 = !time
channel = 0
[State 1020, snd]
type = stopsnd
trigger1 = !time
channel = 1
[State 1020, Stop]
type = velset
trigger1 = 1
x = 0
y = 0
[State 1020, Stop]
type = posfreeze
trigger1 = 1
value = 1
[State 1020, End]
type = changestate
trigger1 = time >= 24
value = 1021


;--- Falling ---
[Statedef 1021]
type = A
physics = N
movetype = I
anim = 1021 + 100 * var(1)
ctrl = 0
velset = 0, 2
sprpriority = -4
[State 1021, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
[State 1021, snd]
type = playsnd
trigger1 = !time
value = 0, 2
channel = 0
[State 1021, End]
type = changestate
trigger1 = pos y >= -21
value = 1022


;--- Hit ground ---
[Statedef 1022]
type = A
physics = N
movetype = I
ctrl = 0
velset = 0, 0
sprpriority = -4
[State 1022, inv]
type = assertspecial
trigger1 = 1
flag = invisible

[State 1022, snd]
type = playsnd
trigger1 = !time
value = 0, 3
channel = 0
[State 1022, End]
type = destroyself
trigger1 = time


;========================< FLY AWAY >=============================
[Statedef 1021]
type = A
physics = N
movetype = I
ctrl = 0
anim = 1000 + 100 * var(1)
sprpriority = -4

[State 1021, Vel]
type = velset
trigger1 = 1
x = 0
y = -.5 * fvar(1)

[State 1021, snd]
type = stopsnd
trigger1 = 1
channel = 0
[State 1021, End]
type = destroyself
trigger1 = screenpos y <= -16


;========================< GRASS + SCORE >=============================
[Statedef 500]
movetype = I
physics = N
sprpriority = -3
[State 500, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
ignorehitpause = 1

[State 500, No Bars]
type = assertspecial
trigger1 = !numpartner
flag = nobardisplay
[State 500, No Bars]
type = assertspecial
triggerall = numpartner
trigger1 = partner, name = "Duck Hunt" && partner, authorname = "Phantom.of.the.Server"
trigger2 = partner, name = "Catwalk" && partner, authorname = "Phantom.of.the.Server"
trigger3 = partner, name = "SS1 Bonus" && partner, authorname = "Phantom.of.the.Server"
trigger4 = partner, name = "Sky Chase" && partner, authorname = "Phantom.of.the.Server"
flag = nobardisplay

[State 500, Pos]
type = posadd
trigger1 = pos x > 256
x = -256
[State 500, Pos]
type = posadd
trigger1 = pos x < -256
x = 256

[State 500, off]
type = offset
trigger1 = 1
y = -5
[State 500, Anim]
type = changeanim
trigger1 = anim != 500
value = 500

[State 500, 1]
type = varset
trigger1 = root, var(6) < 9999999
var(0) = floor(root, var(6) * 0.000001)
[State 500, 10]
type = varset
trigger1 = root, var(6) < 9999999
var(1) = floor(root, var(6) * 0.00001) -(var(0) * 10)
[State 500, 100]
type = varset
trigger1 = root, var(6) < 9999999
var(2) = floor(root, var(6) * 0.0001) -(var(1) * 10) -(var(0) * 100)
[State 500, 1000]
type = varset
trigger1 = root, var(6) < 9999999
var(3) = floor(root, var(6) * 0.001) -(var(2) * 10) -(var(1) * 100) -(var(0) * 1000)
[State 500, 10000]
type = varset
trigger1 = root, var(6) < 9999999
var(4) = floor(root, var(6) * 0.01) -(var(3) * 10) -(var(2) * 100) -(var(1) * 1000) -(var(0) * 10000)
[State 500, 100000]
type = varset
trigger1 = root, var(6) < 9999999
var(5) = floor(root, var(6) * 0.1) -(var(4) * 10) -(var(3) * 100) -(var(2) * 1000) -(var(1) * 10000) -(var(0) * 100000)
[State 500, 1000000]
type = varset
trigger1 = root, var(6) < 9999999
var(6) = root, var(6) -(var(5) * 10) -(var(4) * 100) -(var(3) * 1000) -(var(2) * 10000) -(var(1) * 100000) -(var(0) * 1000000)
[State 500, max]
type = varrangeset
trigger1 = root, var(6) >= 9999999
first = 0
last = 6
value = 9

[State 500, board]
type = explod
trigger1 = !numexplod(600)
anim = 600
ID = 600
sprpriority = 15
postype = left
pos = ifelse(teamside = 1, 40, 280), 19
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 500, 1]
type = explod
trigger1 = 1
anim = 610 + var(6)
ID = 610
sprpriority = 16
postype = left
pos = ifelse(teamside = 1, 61, 301), 13
bindtime = -1
removetime = 1
pausemovetime = -1
supermovetime = -1
[State 500, 10]
type = explod
trigger1 = 1
anim = 610 + var(5)
ID = 611
sprpriority = 16
postype = left
pos = ifelse(teamside = 1, 53, 293), 13
bindtime = -1
removetime = 1
pausemovetime = -1
supermovetime = -1
[State 500, 100]
type = explod
trigger1 = 1
anim = 610 + var(4)
ID = 612
sprpriority = 16
postype = left
pos = ifelse(teamside = 1, 45, 285), 13
bindtime = -1
removetime = 1
pausemovetime = -1
supermovetime = -1
[State 500, 1000]
type = explod
trigger1 = 1
anim = 610 + var(3)
ID = 613
sprpriority = 16
postype = left
pos = ifelse(teamside = 1, 37, 277), 13
bindtime = -1
removetime = 1
pausemovetime = -1
supermovetime = -1
[State 500, 10000]
type = explod
trigger1 = 1
anim = 610 + var(2)
ID = 614
sprpriority = 16
postype = left
pos = ifelse(teamside = 1, 29, 269), 13
bindtime = -1
removetime = 1
pausemovetime = -1
supermovetime = -1
[State 500, 100000]
type = explod
trigger1 = 1
anim = 610 + var(1)
ID = 615
sprpriority = 16
postype = left
pos = ifelse(teamside = 1, 21, 261), 13
bindtime = -1
removetime = 1
pausemovetime = -1
supermovetime = -1
[State 500, 1000000]
type = explod
trigger1 = 1
anim = 610 + var(0)
ID = 616
sprpriority = 16
postype = left
pos = ifelse(teamside = 1, 13, 253), 13
bindtime = -1
removetime = 1
pausemovetime = -1
supermovetime = -1

[State 500, End]
type = destroyself
trigger1 = root, var(17) = 2


;========================< POINTS >=============================
[Statedef 650]
movetype = I
physics = N
anim = 999
[State 650, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
ignorehitpause = 1
[State 650, Stop]
type = posfreeze
trigger1 = 1
value = 1
[State 650, turn]
type = turn
trigger1 = facing != 1

[State 650, 1]
type = varset
trigger1 = !time && parent, var(2) < 9999999
var(0) = floor(parent, var(2) * 0.000001)
[State 650, 10]
type = varset
trigger1 = !time && parent, var(2) < 9999999
var(1) = floor(parent, var(2) * 0.00001) -(var(0) * 10)
[State 650, 100]
type = varset
trigger1 = !time && parent, var(2) < 9999999
var(2) = floor(parent, var(2) * 0.0001) -(var(1) * 10) -(var(0) * 100)
[State 650, 1000]
type = varset
trigger1 = !time && parent, var(2) < 9999999
var(3) = floor(parent, var(2) * 0.001) -(var(2) * 10) -(var(1) * 100) -(var(0) * 1000)
[State 650, 10000]
type = varset
trigger1 = !time && parent, var(2) < 9999999
var(4) = floor(parent, var(2) * 0.01) -(var(3) * 10) -(var(2) * 100) -(var(1) * 1000) -(var(0) * 10000)
[State 650, 100000]
type = varset
trigger1 = !time && parent, var(2) < 9999999
var(5) = floor(parent, var(2) * 0.1) -(var(4) * 10) -(var(3) * 100) -(var(2) * 1000) -(var(1) * 10000) -(var(0) * 100000)
[State 650, 1000000]
type = varset
trigger1 = !time && parent, var(2) < 9999999
var(6) = parent, var(2) -(var(5) * 10) -(var(4) * 100) -(var(3) * 1000) -(var(2) * 10000) -(var(1) * 100000) -(var(0) * 1000000)
[State 650, max]
type = varrangeset
trigger1 = !time && parent, var(2) >= 9999999
first = 0
last = 6
value = 9

[State 650, 1]
type = explod
trigger1 = !time
anim = 650 + var(6)
ID = 650
sprpriority = -5
postype = p1
pos = 12, -4
removetime = -1
[State 650, 10]
type = explod
trigger1 = !time && (var(5) > 0 || var(4) > 0 || var(3) > 0 || var(2) > 0 || var(1) > 0 || var(0) > 0)
anim = 650 + var(5)
ID = 611
sprpriority = -5
postype = p1
pos = 8, -4
removetime = -1
[State 650, 100]
type = explod
trigger1 = !time && (var(4) > 0 || var(3) > 0 || var(2) > 0 || var(1) > 0 || var(0) > 0)
anim = 650 + var(4)
ID = 612
sprpriority = -5
postype = p1
pos = 4, -4
removetime = -1
[State 650, 1000]
type = explod
trigger1 = !time && (var(3) > 0 || var(2) > 0 || var(1) > 0 || var(0) > 0)
anim = 650 + var(3)
ID = 613
sprpriority = -5
postype = p1
pos = 0, -4
removetime = -1
[State 650, 10000]
type = explod
trigger1 = !time && (var(2) > 0 || var(1) > 0 || var(0) > 0)
anim = 650 + var(2)
ID = 614
sprpriority = -5
postype = p1
pos = -4, -4
removetime = -1
[State 650, 100000]
type = explod
trigger1 = !time && (var(1) > 0 || var(0) > 0)
anim = 650 + var(1)
ID = 615
sprpriority = -5
postype = p1
pos = -8, -4
removetime = -1
[State 650, 1000000]
type = explod
trigger1 = !time && var(0) > 0
anim = 650 + var(0)
ID = 616
sprpriority = -5
postype = p1
pos = -12, -4
removetime = -1

[State 650, End]
type = removeexplod
trigger1 = time >= 50 || frontedgedist < -16 || backedgedist < -16
[State 650, End]
type = destroyself
trigger1 = !numexplod


;==========< Stand >==========
[Statedef 0]
type = A
physics = N
movetype = I
anim = 999
ctrl = 0
[State 0, override]
type = changestate
trigger1 = ishelper(1000)
value = 1010

[State 0, duck]
type = helper
trigger1 = !(time % 90)
trigger1 = roundstate = 2 && roundno < 2 && numhelper(1000) < 8
trigger2 = var(19) && numhelper(1000) < 8
helpertype = normal
stateno = 1000
ID = 1000
name = "Duck"
postype = back
pos = 16 + (random % 289), -14
ownpal = 1
[State 0, duck]
type = helper
trigger1 = !(time % 90)
trigger1 = numenemy
trigger1 = (enemy, teammode = simul)
trigger1 = roundstate = 2 && roundno < 2 && numhelper(1000) < 8
helpertype = normal
stateno = 1000
ID = 1000
name = "Duck"
postype = back
pos = 16 + (random % 289), -14
ownpal = 1

[State 0, duck]
type = helper
trigger1 = !(time % 120)
trigger1 = roundstate = 2 && roundno >= 2 && numhelper(1000) < 12
helpertype = normal
stateno = 1000
ID = 1000
name = "Duck"
postype = back
pos = 16 + (random % 289), -14
ownpal = 1
[State 0, duck]
type = helper
trigger1 = !(time % 120)
trigger1 = roundstate = 2 && roundno >= 2 && numhelper(1000) < 12
helpertype = normal
stateno = 1000
ID = 1000
name = "Duck"
postype = back
pos = 16 + (random % 289), -14
ownpal = 1
[State 0, duck]
type = helper
trigger1 = !(time % 120)
trigger1 = numenemy
trigger1 = (enemy, teammode = simul)
trigger1 = roundstate = 2 && roundno >= 2 && numhelper(1000) < 12
helpertype = normal
stateno = 1000
ID = 1000
name = "Duck"
postype = back
pos = 16 + (random % 289), -14
ownpal = 1
[State 0, duck]
type = helper
trigger1 = !(time % 120)
trigger1 = numenemy
trigger1 = (enemy, teammode = simul)
trigger1 = roundstate = 2 && roundno >= 2 && numhelper(1000) < 12
helpertype = normal
stateno = 1000
ID = 1000
name = "Duck"
postype = back
pos = 16 + (random % 289), -14
ownpal = 1

[State 0, kill!]
type = changestate
trigger1 = var(9)
value = 2

[State 0, Ctrl]
type = ctrlset
trigger1 = 1
value = roundstate >= 3
[State 0, sts]
type = statetypeset
trigger1 = roundstate >= 3
statetype = S
[State 0, dead]
type = changestate
trigger1 = !alive
value = 5150


;==========< Wait for Round >==========
[Statedef 1]
type = A
physics = N
movetype = I
anim = 999
ctrl = 0
[State 1, Assert]
type = assertspecial
trigger1 = roundno = 1
flag = nomusic
[State 1, End]
type = changestate
trigger1 = roundstate = 2
value = 0


;==========< Kill All Ducks>==========
[Statedef 2]
type = A
physics = N
movetype = I
anim = 999
ctrl = 0
[State 2, kill!]
type = varset
trigger1 = !time
var(9) = 0
[State 2, End]
type = changestate
trigger1 = time && var(9)
value = 2
[State 2, End]
type = changestate
trigger1 = time >= 60
value = 0


;==========< Dead >==========
[Statedef 5150]
type = A
movetype = H
physics = N
anim = 999
[State 5150, override]
type = changestate
trigger1 = ishelper(1000)
value = 1010

[State 5150, snd]
type = playsnd
trigger1 = !time
value = 0, 4
channel = 1
[State 5150, Flash]
type = allpalfx
trigger1 = !time
time = 8
add = 255, 255, 255
sinadd = -255, -255, -255, 32
[State 5150, hurt]
type = lifeadd
trigger1 = !time
value = -life * 2
kill = 1


;========================< END BGM >=============================
[Statedef 990]
movetype = I
physics = N
anim = 999
[State 990, asserts]
type = assertspecial
trigger1 = time >= 60
flag = nomusic
[State 990, asserts]
type = assertspecial
trigger1 = time <= 270
trigger2 = !(root, var(8)) && (root, var(7)) && time <= 480
flag = roundnotover

[State 990, snd]
type = playsnd
trigger1 = time = 60
value = 0, 9
channel = 1

[State 990, Perfect]
type = parentvaradd
trigger1 = !time
trigger1 = !(root, var(8)) && (root, var(7)) 
var(6) = 10000
[State 990, Perfect]
type = playsnd
trigger1 = time = 300
trigger1 = !(root, var(8)) && (root, var(7)) 
value = 0, 10
channel = 1
[State 990, Perfect]
type = explod
trigger1 = time = 300
trigger1 = !(root, var(8)) && (root, var(7)) && !numexplod(600)
anim = 990
ID = 990
sprpriority = 15
postype = left
pos = 160, 80
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 990, End]
type = stopsnd
trigger1 = roundstate != 4
channel = 1
[State 990, End]
type = destroyself
trigger1 = roundstate != 4


;==========< COMMON ERASE >==========
[Statedef 10]
[State 10, End]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 11]
[State 11, End]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 12]
[State 12, End]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)

[Statedef 20]
[State 20, End]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)

[Statedef 40]
[State 40, End]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 50]
[State 50, End]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 51]
[State 51, End]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 52]
[State 52, End]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)

[Statedef 45]
[State 45, End]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)

[Statedef 100]
[State 100, End]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 105]
[State 105, End]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)


;==========< Initialize >==========
[Statedef 5900]
type = A

[State 5900, varreset]
type = varrangeset
trigger1 = !roundsexisted
value = 0
[State 5900, varreset]
type = varrangeset
trigger1 = !roundsexisted
fvalue = 0

[State 5900, Palette]
type = remappal
trigger1 = 1
source = 1, 1
dest = 1, palno

[State 5900, Reset Missed]
type = varset
trigger1 = 1
var(8) = 0
[State 5900, Reset Time]
type = varset
trigger1 = 1
fvar(11) = 0

[State 5900, End]
type = changestate
trigger1 = 1
value = 5901


;==========< INTRO >==========
[Statedef 190]
type = S
ctrl = 0
velset = 0, 0
[State 190, intro]
type = playsnd
trigger1 = !time && roundno = 1
value = 0, 5
channel = 1
[State 190, dog]
type = helper
trigger1 = !numhelper(400)
helpertype = normal
stateno = 400
ID = 400
name = "Dog"
postype = left
pos = 160, 0
facing = ifelse(teamside = 1, -1, 1)
size.shadowoffset = -3
ownpal = 1
[State 190, End]
type = changestate
trigger1 = 1
value = 1


;==========< TIME OVER >==========
[Statedef 170]
physics = N
movetype = I
ctrl = 1
[State 170, End]
type = changestate
trigger1 = 1
value = 5150

[Statedef 175]
physics = N
movetype = I
ctrl = 1
[State 175, End]
type = changestate
trigger1 = 1
value = 5150


;==========< WIN >==========
[Statedef 180]
physics = N
movetype = I
ctrl = 0

[State 180, Quote]
type = victoryquote
trigger1 = 1
value = 1


;========================< DOG INTRO >=============================
[Statedef 400]
physics = N
movetype = I
anim = ifelse(roundno = 1, 400, 402)
ctrl = 0
sprpriority = -2

[State 400, fix]
type = destroyself
trigger1 = !time && numhelper(400) > 1
time = -1
value = SCA
ignorehitpause = 1

[State 400, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
ignorehitpause = 1
[State 400, asserts]
type = assertspecial
trigger1 = 1
flag = intro
[State 400, off]
type = offset
trigger1 = 1
y = -3

[State 400, Anim]
type = changeanim
trigger1 = (anim = 400 || anim = 402) && !animtime
value = 401
[State 400, bark]
type = playsnd
trigger1 = anim = 401
value = 0, 6
persistent = 0

[State 400, jump]
type = velset
trigger1 = anim = 401
x = 1
y = -3
persistent = 0
[State 400, jump]
type = veladd
trigger1 = anim = 401
y = 0.167

[State 400, Anim]
type = changeanim
trigger1 = anim = 401 && animelemtime(2) < 0 && vel y > 0
value = 401
elem = 2

[State 400, Anim]
type = sprpriority
trigger1 = anim = 401 && animelemtime(2) >= 0
value = -4

[State 400, End]
type = destroyself
trigger1 = anim = 401 && vel y > 0 && pos y > 0


;========================< DOG LAUGH >=============================
[Statedef 410]
physics = N
movetype = I
anim = 410
ctrl = 0
sprpriority = -4

[State 410, fix]
type = destroyself
trigger1 = !time && numhelper(400) > 1
time = -1
value = SCA
ignorehitpause = 1

[State 410, NHB]
type = nothitby
trigger1 = anim != 410
time = -1
value = SCA
ignorehitpause = 1
[State 410, hit]
type = hitoverride
trigger1 = 1
attr = SCA, AA, AP, AT
stateno = 420
time = -1

[State 410, HB]
type = hitby
trigger1 = anim = 410
time = -1
value = SCA, AA, AP
ignorehitpause = 1

[State 410, off]
type = offset
trigger1 = 1
y = -7
[State 410, laugh]
type = playsnd
trigger1 = !time
value = 0, 7
channel = 0

[State 410, Pos]
type = posset
trigger1 = !time
y = 17
[State 410, Vel]
type = velset
trigger1 = time < 120
x = 0
y = ifelse(time < 40, -1, ifelse(time >= 80, 1, 0))
[State 410, Pos]
type = posset
trigger1 = time >= 120
y = 17
[State 410, Vel]
type = velset
trigger1 = time >= 120
y = 0

[State 410, Anim]
type = changeanim
trigger1 = anim != 411 && pos y >= 0
trigger2 = anim != 410 && pos y < 0
value = 410 + (pos y >= 0)
elem = animelemno(0)

[State 410, hit]
type = changestate
trigger1 = (root, stateno = 2)
value = 420

[State 410, inv]
type = assertspecial
trigger1 = time >= 120
flag = invisible
[State 410, End]
type = destroyself
trigger1 = time >= 180


;========================< DOG HIT >=============================
[Statedef 420]
physics = N
movetype = I
anim = 420
ctrl = 0
sprpriority = -4

[State 420, fix]
type = destroyself
trigger1 = !time && numhelper(400) > 1
time = -1
value = SCA
ignorehitpause = 1

[State 420, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
ignorehitpause = 1

[State 420, off]
type = offset
trigger1 = 1
y = -7
[State 420, laugh]
type = stopsnd
trigger1 = !time
channel = 0

[State 420, Pos]
type = posset
trigger1 = !time
y = -23
[State 420, Vel]
type = velset
trigger1 = 1
x = 0
y = ifelse((time < 80 || time >= 120), 0, 1)
[State 420, Pos]
type = posset
trigger1 = time >= 120
y = 17

[State 420, Anim]
type = changeanim
trigger1 = anim != 421 && pos y >= 0
trigger2 = anim != 420 && pos y < 0
value = 420 + (pos y >= 0)
elem = animelemno(0)

[State 420, inv]
type = assertspecial
trigger1 = time >= 120
flag = invisible
[State 420, End]
type = destroyself
trigger1 = time >= 180


;========================< DOG CATCH >=============================
[Statedef 450]
physics = N
movetype = I
anim = 999
ctrl = 0
sprpriority = -4

[State 450, fix]
type = destroyself
trigger1 = !time && numhelper(400) > 1
time = -1
value = SCA
ignorehitpause = 1

[State 450, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
ignorehitpause = 1
[State 450, off]
type = offset
trigger1 = 1
y = -7
[State 450, yay]
type = playsnd
trigger1 = time = 30
value = 0, 8

[State 450, colorvar]
type = varset
trigger1 = !time
trigger1 = playerIDexist(parent, ID)
var(1) = parent, var(1)
[State 450, Anim]
type = changeanim
trigger1 = time >= 30
value = (450 + var(1)) + (10 * (pos y >= 0))

[State 450, Pos]
type = posset
trigger1 = !time
y = 17
[State 450, Vel]
type = velset
trigger1 = time = [30, 139]
x = 0
y = ifelse(time < 70, -1, ifelse(time >= 100, 1, 0))
[State 410, Pos]
type = posset
trigger1 = time >= 140
y = 17
[State 450, Vel]
type = velset
trigger1 = time >= 140
y = 0

[State 450, inv]
type = assertspecial
trigger1 = time >= 140
flag = invisible
[State 450, End]
type = destroyself
trigger1 = time >= 180


;==========< HIT ERASE >==========
[Statedef 5000]
movetype = H
[State 5000, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5001]
movetype = H
[State 5001, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5002]
movetype = H
[State 5002, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5010]
movetype = H
[State 5010, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5020]
movetype = H
[State 5020, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5030]
movetype = H
[State 5030, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5035]
movetype = H
[State 5035, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5040]
movetype = H
[State 5040, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5050]
movetype = H
[State 5050, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5070]
movetype = H
[State 5070, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5071]
movetype = H
[State 5071, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5080]
movetype = H
[State 5080, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5081]
movetype = H
[State 5081, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5100]
movetype = H
[State 5100, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5101]
movetype = H
[State 5101, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5110]
movetype = H
[State 5110, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5120]
movetype = H
[State 5120, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5200]
movetype = H
[State 5200, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5201]
movetype = H
[State 5201, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)
[Statedef 5210]
movetype = H
[State 5210, override]
type = changestate
trigger1 = 1
value = ifelse(ishelper(1000), 1010, 0)


;===========================================================================
;===============================< -2 STATES >=================================
;===========================================================================
[Statedef -2]

[State -2, Time]
type = varadd
trigger1 = roundstate = 2
fvar(11) = 0.01666667
[State -2, Time]
type = assertspecial
trigger1 = fvar(10) < 0
flag = timerfreeze

[State -2, Anim]
type = changeanim
trigger1 = 1
value = 999
[State -2, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
ignorehitpause = 1
[State -2, Life]
type = lifeset
trigger1 = roundstate >= 3 && life > 1
value = 1

[State -2, No Push]
type = playerpush
trigger1 = 1
value = 0
[State -2, Not Bound]
type = screenbound
trigger1 = roundstate >= 2
value = 0
movecamera = 0, 0
[State -2, Pos]
type = posset
trigger1 = roundstate >= 2
x = 0
y = -64
[State -2, Pos]
type = posset
trigger1 = roundstate >= 2
trigger1 = numenemy = 1
x = ifelse((enemynear, facing = -1), -20, 20)
y = -64

[State -2, Assisting?]
type = varset
trigger1 = numpartner
trigger1 = partner, name = "Duck Hunt" && partner, authorname = "Phantom.of.the.Server"
var(17) = ifelse((ID > partner, ID), 2, 1)
[State -2, Assisting?]
type = varset
trigger1 = !numpartner
var(17) = 0

[State -2, Score]
type = varset
trigger1 = var(6) < 0
var(6) = 0

[State -2, Share Score]
type = varadd
trigger1 = numpartner && var(17) = 1
var(6) = partner, var(6)
[State -2, Share Score]
type = varset
trigger1 = var(6)
trigger1 = numpartner && var(17) = 2
var(6) = 0

[State -2, Share Hit Ducks]
type = varadd
trigger1 = numpartner && var(17) = 1
var(7) = partner, var(7)
[State -2, Share Hit Ducks]
type = varset
trigger1 = var(7)
trigger1 = numpartner && var(17) = 2
var(7) = 0

[State -2, Share Missed Ducks]
type = varadd
trigger1 = numpartner && var(17) = 1
var(8) = partner, var(8)
[State -2, Share Missed Ducks]
type = varset
trigger1 = var(8)
trigger1 = numpartner && var(17) = 2
var(8) = 0

[State -2, Grass]
type = helper
trigger1 = var(17) != 2
trigger1 = !numhelper(500)
helpertype = normal
stateno = 500
ID = 500
name = "Grass + Score"
postype = left
pos = 160, 0
size.shadowoffset = -4
pausemovetime = 2147483647
supermovetime = 2147483647
ownpal = 1

[State -2, End BGM]
type = helper
trigger1 = var(17) != 2
trigger1 = roundstate = 4 && !numhelper(990)
helpertype = normal
stateno = 990
ID = 990
name = "BGM"
postype = left
pos = 160, 0
pausemovetime = 2147483647
supermovetime = 2147483647

[State -2, Kill]
type = varset
triggerall = numenemy
trigger1 = enemy, name = "Shin Gouki" && enemy, authorname = "Phantom.of.the.Server"
trigger1 = enemy, stateno = 4001 && enemy, animelemtime(35) >= 0 && enemy, animelemtime(36) < 0
trigger2 = enemy, name = "Ryu" && enemy, authorname = "Phantom.of.the.Server"
trigger2 = enemy, stateno = 4001 && enemy, animelemtime(17) >= 0 && enemy, animelemtime(22) < 0
trigger3 = enemy, name = "Dan Hibiki" && enemy, authorname = "Phantom.of.the.Server"
trigger3 = (enemy, stateno = 4001 && enemy, animelemtime(6) >= 0 && enemy, animelemtime(7) < 0) || (enemy, stateno = 4101)
trigger4 = enemy, name = "Sakura Kasugano" && enemy, authorname = "Phantom.of.the.Server"
trigger4 = enemy, stateno = 4001 && enemy, animelemtime(8) >= 0 && enemy, animelemtime(23) < 0
trigger5 = enemy, name = "Yuri Sakazaki" && enemy, authorname = "Phantom.of.the.Server"
trigger5 = (enemy, stateno = 3301 || enemy, stateno = 3311) && enemy, animelemtime(2) >= 0
trigger6 = enemy, name = "Nash" && enemy, authorname = "Phantom.of.the.Server"
trigger6 = enemy, stateno = 3402
trigger7 = enemy, name = "Rare Akuma" && enemy, authorname = "Phantom.of.the.Server"
trigger7 = enemy, stateno = 4010 && enemy, animelemtime(3) >= 0
var(9) = 1

[State -2, Kill]
type = varset
triggerall = numenemy
trigger1 = enemy, name = "ss3goku" && enemy, authorname = "Warusaki3"
trigger1 = enemy, stateno = 3602 && enemy, animelemtime(24) >= 0 && enemy, animelemtime(25) < 0
trigger2 = enemy, name = "Noroko" && enemy, authorname = "Andrés Borghi"
trigger2 = enemy, stateno = 44 || enemy, stateno = 3511
trigger3 = enemy, name = "MJJ" && enemy, authorname = "RyuUi SinYongKwon"
trigger3 = enemy, stateno = 3201 || enemy, stateno = 3205
var(9) = 1

[State -2, Extra Ducks]
type = varset
trigger1 = 1
var(19) = 0
[State -2, Extra Ducks]
type = varset
triggerall = numenemy
trigger1 = enemy, stateno = 195 && !(enemy, prevstateno = 195) && enemy, time = 1
trigger2 = enemy, name = "Nakoruru" && enemy, authorname = "Phantom.of.the.Server"
trigger2 = enemy, stateno = 3201
trigger3 = enemy, name = "Rare Akuma" && enemy, authorname = "Phantom.of.the.Server"
var(19) = 1

[State -2, Debug]
type = displaytoclipboard
trigger1 = 1
text = "Score = %d Hit = %d Miss = %d Time = %f / %f"
params = var(6), var(7), var(8), fvar(11), fvar(10)
ignorehitpause = 1

[State -2, Debug]
type = displaytoclipboard
trigger1 = var(17) = 2
text = "Assisting..."
ignorehitpause = 1

[State -2, Debug]
type = appendtoclipboard
trigger1 = 1
text = "\nBy PotS - http://www.mugenguild.com/pots/"
ignorehitpause = 1

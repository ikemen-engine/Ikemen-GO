; ___________________________________
;| Cat Walk by Phantom.of.the.Server |
; ¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯
;====================< MAIN STATE >====================
[Statedef 0]
type = A
physics = N
movetype = I
anim = 0

[State 0, OK]
type = changestate
trigger1 = numtarget(1)
value = 2
[State 0, Attack]
type = changestate
trigger1 = ctrl
trigger1 = roundstate = 2
trigger1 = !numtarget(1)
trigger1 = fvar(2) != 0 && var(2) > 0
value = 200


;====================< PLAYER ACQUIRED >====================
[Statedef 1]
type = A
physics = N
movetype = I
anim = 0
ctrl = 0

[State 1, Snd]
type = playsnd
trigger1 = !time
value = 4, (random % 4)
channel = 2
abspan = 0
[State 1, Confirmation Sound]
type = playsnd
trigger1 = !time
value = 5, 0

[State 1, Text]
type = explod
trigger1 = !time
anim = 1
ontop = 1
postype = left
bindtime = -1
removetime = -2
pausemovetime = -1
supermovetime = -1

[State 1, End]
type = changestate
trigger1 = time >= 60
trigger2 = numtarget(1)
trigger2 = target(1), stateno != [2030, 2031]
value = 2


;====================< PLAYING >====================
[Statedef 2]
type = A
physics = N
movetype = I
anim = 0
ctrl = 0

[State 2, Time]
type = varadd
trigger1 = roundstate = 2
fvar(2) = -.01666667
[State 2, Player ID]
type = varset
trigger1 = numtarget(1)
var(1) = target(1), ID

[State 2, Dizzy]
type = targetstate
trigger1 = var(2) <= 30
trigger1 = numtarget(1)
trigger1 = target(1), stateno = 2000
ID = 1
value = 2001

[State 2, Miss]
type = targetstate
trigger1 = numtarget(1)
trigger1 = var(10) = -1
ID = 1
value = 2010
[State 2, Hit]
type = targetstate
trigger1 = numtarget(1)
trigger1 = var(10) = 1
trigger1 = var(11) = [2000, 2199]
ID = 1
value = var(11)

[State 2, Life]
type = varadd
trigger1 = var(10) = 1
var(2) = 1
[State 2, Life]
type = varadd
trigger1 = var(10) = -1
var(2) = -20
[State 2, Life Range]
type = varset
trigger1 = var(2) > 144
var(2) = 144
[State 2, Life Range]
type = varset
trigger1 = var(2) < 0
var(2) = 0

[State 2, Combo Reset]
type = varset
trigger1 = var(12) > 0 && var(10) < 0
trigger2 = var(12) < 0 && var(10) > 0
var(12) = 0
[State 2, Combo + ]
type = varadd
trigger1 = var(10) = 1
var(12) = 1
[State 2, Combo -]
type = varadd
trigger1 = var(10) = -1
var(12) = -1

[State 2, Number of Hits]
type = varadd
trigger1 = var(10) = 1
var(20) = 1
[State 2, Number of Missed]
type = varadd
trigger1 = var(10) = -1
var(21) = 1

[State 2, Score - All]
type = varadd
trigger1 = var(10) = 1
var(24) = ceil((1 + (var(50) != 0) * 0.25 + (var(52) != 0) * 0.25) * (abs(fvar(4)) * (5 + var(13) * 5)))
[State 2, Score - + 5 Chain]
type = varadd
trigger1 = var(10) = 1
trigger1 = var(12) > 0 && !(var(12) % 5)
var(24) = ceil(abs(fvar(4)) * 10)
[State 2, Score - Speed Up]
type = varadd
trigger1 = var(10) = 1
trigger1 = var(12) = 50 || (var(12) >= 100 && !(var(12) % 100))
var(24) = ceil(abs(fvar(4)) * 40)

[State 2, Snd]
type = playsnd
trigger1 = var(10) = 1
trigger1 = var(13) = [1, 4]
value = 1, var(13)
channel = 1
abspan = ifelse(teamside = 1, 160, -160)
[State 2, Snd]
type = playsnd
trigger1 = var(10) = 1
trigger1 = var(12) > 0 && !(var(12) % 5)
value = 2, (random % 23)
channel = 2
abspan = 0
[State 2, Snd]
type = playsnd
trigger1 = var(10) = -1
trigger1 = var(12) < 0
value = 3, (random % 2)
channel = 3
lowpriority = 1
abspan = 0

[State 2, Applause]
type = playsnd
trigger1 = var(10) = 1
trigger1 = var(12) = 50 || (var(12) >= 100 && !(var(12) % 100))
value = 5, 1
[State 2, Boo]
type = playsnd
trigger1 = var(10) = -1 && var(2) <= 0
value = 5, 2

[State 2, Message]
type = helper
triggerall = var(10) = 1
trigger1 = var(12) >= 5 && !(var(12) % 5)
helpertype = normal
stateno = 800
ID = 800
name = "Message"
postype = left
pos = 0, 0
pausemovetime = 65535
supermovetime = 65535
ownpal = 1

[State 2, Hit]
type = envshake
trigger1 = var(10) = 1
trigger1 = var(13) = 4
time = 24
ampl = 4
[State 2, Hit]
type = explod
trigger1 = var(10) = 1
anim = 590 + (random % 6)
sprpriority = 8
postype = left
pos = ifelse(teamside = 1, 66, 254), 24
scale = ifelse(random < 500, -.5, 0.5), ifelse(random < 500, -.5, 0.5)
bindtime = -1
pausemovetime = -1
supermovetime = -1
[State 2, Hit]
type = explod
trigger1 = var(10) = 1
anim = 590 + (random % 6)
sprpriority = 8
postype = left
pos = ifelse(teamside = 1, 66, 254), 24
scale = ifelse(random < 500, -.3, 0.3), ifelse(random < 500, -.3, 0.3)
bindtime = -1
pausemovetime = -1
supermovetime = -1

[State 2, Speed Change]
type = varadd
trigger1 = var(10) = 1
trigger1 = var(12) = 50 || (var(12) >= 100 && !(var(12) % 100))
fvar(2) = 5.0
[State 2, Speed Change]
type = varset
trigger1 = var(12) < 50
var(4) = 0
[State 2, Speed Change]
type = varadd
trigger1 = var(10) = 1
trigger1 = var(12) = 50 || (var(12) >= 100 && !(var(12) % 100))
var(4) = 1

[State 2, Speed]
type = varset
trigger1 = 1
fvar(4) = fvar(3) * (1 + 0.25 * (var(4) + roundsexisted))
[State 2, Interval]
type = varset
trigger1 = fvar(4)
var(3) = floor(24 / fvar(4))

[State 2, Clean Timer]
type = varset
trigger1 = var(10) = 1
trigger1 = var(12) = 50 || (var(12) >= 100 && !(var(12) % 100))
trigger2 = var(10) = -1
var(8) = 8

[State 2, Sequence Length Decrease]
type = varadd
trigger1 = var(6) > 0
trigger1 = !(gametime % var(3))
var(6) = -1
[State 2, Sequence Type Reset]
type = varset
trigger1 = var(6) <= 0
var(5) = 0

[State 2, ABC /xYZ]
type = varset
trigger1 = var(6) <= 0
trigger2 = var(5) = 0 || var(5) = 1
var(7) = ifelse(var(50) = 0, 0, ifelse(var(7) = 0, (random < 333), (random < 666)))

[State 2, Sequence Type]
type = varset
trigger1 = var(6) <= 0
trigger1 = !(gametime % var(3))
var(5) = ifelse(random < ifelse(var(12) >= 100, 0, ifelse(var(12) >= 50, 250, 0)), 1, (random % 33) + 2)
[State 2, Blank Sequence]
type = varset
trigger1 = var(6) <= 0 && var(5) != 0
trigger1 = !(gametime % var(3))
trigger1 = ifelse(numhelper(1000) < 4, 0, ifelse(numhelper(1000) < 6, random < 250, random < 500))
var(5) = 0
[State 2, No Blank Sequence]
type = varset
trigger1 = !numhelper(1000) && var(5) = 0
var(5) = 1

[State 2, Speed Change - > Blank Sequence]
type = varset
trigger1 = var(10) = 1
trigger1 = var(12) = 50 || (var(12) >= 100 && !(var(12) % 100))
trigger2 = var(10) = -1
var(5) = 0
[State 2, Speed Change - > Blank Sequence]
type = varset
trigger1 = var(10) = 1
trigger1 = var(12) = 50 || (var(12) >= 100 && !(var(12) % 100))
trigger2 = var(10) = -1
var(6) = 2

[State 2, Blank Sequence Length]
type = varset
triggerall = var(6) <= 0
trigger1 = var(5) = 0
var(6) = (random % 3) + 2
[State 2, Random Sequence Length]
type = varset
triggerall = var(6) <= 0
trigger1 = var(5) = 1
var(6) = (random % 4) + 3

[State 2, Sequence Length of 5]
type = varset
triggerall = var(6) <= 0
trigger1 = var(5) = 27 || var(5) = 30
var(6) = 5
[State 2, Sequence Length of 4]
type = varset
triggerall = var(6) <= 0
trigger1 = var(5) = 7 || var(5) = 8 || var(5) = 9 || var(5) = 14 || var(5) = 15 || var(5) = 16 || var(5) = 17
trigger2 = var(5) = 18 || var(5) = 19 || var(5) = 20 || var(5) = 21 || var(5) = 28 || var(5) = 29
trigger3 = var(5) = 31 || var(5) = 32 || var(5) = 33 || var(5) = 34
var(6) = 4
[State 2, Sequence Length of 3]
type = varset
triggerall = var(6) <= 0
trigger1 = var(5) = 2 || var(5) = 3 || var(5) = 4 || var(5) = 5 || var(5) = 6 || var(5) = 10 || var(5) = 11
trigger2 = var(5) = 12 || var(5) = 13 || var(5) = 22 || var(5) = 23 || var(5) = 24 || var(5) = 25
var(6) = 3

[State 2, Hit Time]
type = varadd
trigger1 = var(10) > 0
var(10) = 1
[State 2, Miss Time]
type = varadd
trigger1 = var(10) < 0
var(10) = -1

[State 2, Commands]
type = helper
triggerall = roundstate = 2
triggerall = !(gametime % var(3))
triggerall = numenemy && numhelper(1000) < 16
trigger1 = 1
helpertype = normal
stateno = 1000
ID = 1000
name = "Commands"
postype = left
pos = 0, 0
ownpal = 1

[State 2, Attack]
type = changestate
trigger1 = roundstate = 2
trigger1 = !numtarget(1)
value = 5150
[State 2, Over]
type = changestate
trigger1 = roundstate != 2
value = 0


;====================< GET PLAYERS >====================
[Statedef 200]
type = S
physics = N
movetype = A
anim = 200
velset = 0, 0
ctrl = 0

[State 200, hit]
type = hitdef
trigger1 = 1
attr = S, NT
hitflag = MAFD
guardflag = 
priority = 7, Hit
damage = 0, 0
getpower = 0, 0
givepower = 0, 0
yaccel = 0.5
fall.recover = 0
sparkno = -1
guard.sparkno = -1
kill = 0
ID = 1
nochainID = 1
numhits = 0
hitonce = 1
p1stateno = 1
p2stateno = 2030

[State 200, OK]
type = changestate
trigger1 = numtarget(1)
value = 1
[State 200, No Player]
type = changestate
trigger1 = !numtarget(1)
trigger1 = time >= 180
value = 5151


;====================< DISPLAY >====================
[Statedef 500]
movetype = I
physics = N
anim = 0

[State 500, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
ignorehitpause = 1

[State 500, No Bars]
type = assertspecial
trigger1 = !numpartner && numenemy <= 1
flag = nobardisplay
[State 500, No Bars]
type = assertspecial
triggerall = numpartner
trigger1 = partner, name = "Duck Hunt" && partner, authorname = "Phantom.of.the.Server"
trigger2 = partner, name = "Catwalk" && partner, authorname = "Phantom.of.the.Server"
trigger3 = partner, name = "SS1 Bonus" && partner, authorname = "Phantom.of.the.Server"
trigger4 = partner, name = "Sky Chase" && partner, authorname = "Phantom.of.the.Server"
flag = nobardisplay

[State 500, Bar]
type = explod
trigger1 = !numexplod(500)
anim = 500
ID = 500
sprpriority = 7
postype = left
pos = ifelse(teamside = 2, 196, 6), 24
scale = 1, 1
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 500, Frame]
type = explod
trigger1 = !numexplod(501)
anim = 501
ID = 501
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 202, 12), 120
scale = 1, 24
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 500, Fill]
type = explod
trigger1 = !numexplod(505)
anim = 505
ID = 505
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 207, 17), -120
scale = 4, 20
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 500, Life]
type = explod
trigger1 = !numexplod(610)
anim = 610
ID = 610
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 15, 233), 15
scale = 1, 1
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 500, Time]
type = explod
trigger1 = !numexplod(620)
anim = 620
ID = 620
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 15, 233), 32
scale = 1, 1
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 500, Lifebar Frame]
type = explod
trigger1 = !numexplod(600)
anim = 600
ID = 600
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 15, 233), 15
scale = 1, 1
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 500, Life Scale]
type = modifyexplod
trigger1 = numexplod(610)
ID = 610
scale = root, var(2) / 144.0 , 1
[State 500, Life Scale]
type = modifyexplod
trigger1 = numexplod(610)
trigger1 = root, var(2) <= 0
ID = 610
scale = 0, 1
[State 500, Life Scale]
type = modifyexplod
trigger1 = numexplod(610)
trigger1 = root, var(2) >= 144
ID = 610
scale = 1, 1

[State 500, Time Scale]
type = modifyexplod
trigger1 = numexplod(620)
trigger1 = root, fvar(1) > 0
ID = 620
scale = root, fvar(2) / root, fvar(1) , 1
[State 500, Time Scale]
type = modifyexplod
trigger1 = numexplod(620)
trigger1 = (root, fvar(2) / root, fvar(1)) <= 0
ID = 620
scale = 0, 1
[State 500, Time Scale]
type = modifyexplod
triggerall = numexplod(620)
trigger1 = (root, fvar(2) / root, fvar(1)) >= 1
trigger2 = root, fvar(1) = -1
ID = 620
scale = 1, 1


;--- Chain ---
[State 500, Chain!]
type = explod
trigger1 = !numexplod(650)
anim = 650
ID = 650
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 102, 137), 5
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 500, 100]
type = varset
trigger1 = root, var(12) = [1, 998]
var(12) = floor(root, var(12) * 0.01)
[State 500, 10]
type = varset
trigger1 = root, var(12) = [1, 998]
var(11) = floor(root, var(12) * 0.1) - (var(12) * 10)
[State 500, 1]
type = varset
trigger1 = root, var(12) = [1, 998]
var(10) = root, var(12) - (var(11) * 10) - (var(12) * 100)

[State 500, Min]
type = varrangeset
trigger1 = root, var(12) <= 0
first = 10
last = 12
value = 0
[State 500, Max]
type = varrangeset
trigger1 = root, var(12) >= 999
first = 10
last = 12
value = 9

[State 500, 1]
type = explod
trigger1 = 1
anim = 660 + var(10)
ID = 660
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 130, 165) + ifelse(var(12), 28, ifelse(var(11), 14, 0)), 15
bindtime = -1
removetime = 1
pausemovetime = -1
supermovetime = -1
[State 500, 10]
type = explod
trigger1 = var(11) || var(12)
anim = 660 + var(11)
ID = 660
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 116, 151) + ifelse(var(12), 14, 0), 15
bindtime = -1
removetime = 1
pausemovetime = -1
supermovetime = -1
[State 500, 100]
type = explod
trigger1 = var(12)
anim = 660 + var(12)
ID = 660
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 102, 137), 15
bindtime = -1
removetime = 1
pausemovetime = -1
supermovetime = -1

[State 500, End]
type = changestate
trigger1 = roundstate = 4
value = 506


;--- Intro ---
[Statedef 505]
movetype = I
physics = N
anim = 0

[State 506, End]
type = removeexplod
trigger1 = !time

[State 505, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
ignorehitpause = 1

[State 505, No Bars]
type = assertspecial
trigger1 = !numpartner && numenemy <= 1
flag = nobardisplay
[State 505, No Bars]
type = assertspecial
triggerall = numpartner
trigger1 = partner, name = "Duck Hunt" && partner, authorname = "Phantom.of.the.Server"
trigger2 = partner, name = "Catwalk" && partner, authorname = "Phantom.of.the.Server"
trigger3 = partner, name = "SS1 Bonus" && partner, authorname = "Phantom.of.the.Server"
trigger4 = partner, name = "Sky Chase" && partner, authorname = "Phantom.of.the.Server"
flag = nobardisplay

[State 505, Bar]
type = explod
trigger1 = !numexplod(500)
anim = 500
ID = 500
sprpriority = 7
postype = left
pos = ifelse(teamside = 2, 196, 6) + ifelse(teamside = 2, -10 * 32, 10 * 32), 24
vel = ifelse(teamside = 2, 10, -10)
scale = 1, 1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 505, Frame]
type = explod
trigger1 = !numexplod(501)
anim = 501
ID = 501
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 202, 12), 120 + 32 * 20
vel = 0, -20
scale = 1, 24
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 505, Lifebar Frame]
type = explod
trigger1 = !numexplod(600)
anim = 600
ID = 600
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 15, 233) + ifelse(teamside = 2, 10 * 32, -10 * 32), 15
vel = ifelse(teamside = 2, -10, 10)
scale = 1, 1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 505, Fill]
type = explod
trigger1 = !numexplod(505)
anim = 505
ID = 505
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 207, 17), -120 - 32 * 20
vel = 0, 20
scale = 4, 20
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 505, End]
type = removeexplod
trigger1 = time >= 32
[State 505, End]
type = changestate
trigger1 = time >= 32
value = 500


;--- Outro ---
[Statedef 506]
movetype = I
physics = N
anim = 0

[State 506, End]
type = removeexplod
trigger1 = !time

[State 506, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
ignorehitpause = 1

[State 506, No Bars]
type = assertspecial
trigger1 = !numpartner && numenemy <= 1
flag = nobardisplay
[State 506, No Bars]
type = assertspecial
triggerall = numpartner
trigger1 = partner, name = "Duck Hunt" && partner, authorname = "Phantom.of.the.Server"
trigger2 = partner, name = "Catwalk" && partner, authorname = "Phantom.of.the.Server"
trigger3 = partner, name = "SS1 Bonus" && partner, authorname = "Phantom.of.the.Server"
trigger4 = partner, name = "Sky Chase" && partner, authorname = "Phantom.of.the.Server"
flag = nobardisplay

[State 506, Bar]
type = explod
trigger1 = !numexplod(500)
anim = 500
ID = 500
sprpriority = 7
postype = left
pos = ifelse(teamside = 2, 196, 6), 24
vel = -10, 0
scale = 1, 1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 506, Frame]
type = explod
trigger1 = !numexplod(501)
anim = 501
ID = 501
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 202, 12), 120
vel = 0, 20
scale = 1, 24
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 506, Lifebar Frame]
type = explod
trigger1 = !numexplod(600)
anim = 600
ID = 600
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 15, 233), 15
vel = 10, 0
scale = 1, 1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 506, Fill]
type = explod
trigger1 = !numexplod(505)
anim = 505
ID = 505
sprpriority = 5
postype = left
pos = ifelse(teamside = 2, 207, 17), -120
vel = 0, -20
scale = 4, 20
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 506, End]
type = removeexplod
trigger1 = time >= 32
[State 506, End]
type = destroyself
trigger1 = time >= 32


;====================< COMBO MESSAGE >====================
[Statedef 800]
movetype = I
physics = N
anim = 800 + (random % 16)
sprpriority = 8

[State 800, Assert]
type = assertspecial
trigger1 = 1
flag = noshadow

[State 800, Pos]
type = posset
trigger1 = !time
x = ceil(40 + (random / 12.5)) * ifelse(teamside = 1, 1, -1)
y = ceil(pos y -screenpos y) + ceil(60 + (random / 8.33))

[State 800, Vel]
type = velmul
trigger1 = 1
x = 0.95
y = 0.95
[State 800, Vel]
type = velset
trigger1 = !time
x = (random / 250.0) * ifelse((teamside = 1 && pos x > -80) || (teamside = 2 && pos x > 80), -1, 1)
y = (random / 250.0) * ifelse((pos y -screenpos y) > 120, -1, 1)

[State 800, trans]
type = trans
trigger1 = time < 64
trans = addalpha
alpha = 256, 0
ignorehitpause = 1
[State 800, trans]
type = trans
trigger1 = time >= 64
trans = addalpha
alpha = 256 -(16 * (time -64)), (16 * (time -64))
ignorehitpause = 1

[State 800, angle]
type = angledraw
trigger1 = time < 64
scale = 0.5, 0.5
[State 800, angle]
type = angledraw
trigger1 = time >= 64
scale = 0.5 + ((time -64) * 0.03125), 0.5 + ((time -64) * 0.03125)

[State 800, End]
type = destroyself
trigger1 = time >= 80


;====================< OUTCOME >====================
[Statedef 700]
movetype = I
physics = N
anim = 0

[State 700, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
ignorehitpause = 1

[State 700, No Bars]
type = assertspecial
trigger1 = !numpartner && numenemy <= 1
flag = nobardisplay
[State 700, No Bars]
type = assertspecial
triggerall = numpartner
trigger1 = partner, name = "Duck Hunt" && partner, authorname = "Phantom.of.the.Server"
trigger2 = partner, name = "Catwalk" && partner, authorname = "Phantom.of.the.Server"
trigger3 = partner, name = "SS1 Bonus" && partner, authorname = "Phantom.of.the.Server"
trigger4 = partner, name = "Sky Chase" && partner, authorname = "Phantom.of.the.Server"
flag = nobardisplay

[State 700, Frame]
type = explod
trigger1 = !numexplod(700)
anim = 700
ID = 700
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 10, 204), 10
scale = 1, 1
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 700, Fill]
type = explod
trigger1 = !numexplod(701)
anim = 505
ID = 701
sprpriority = 4
postype = left
pos = ifelse(teamside = 1, 10, 204), 10
scale = 4.417, 4.292
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1


;--- Max Chain ---
[State 700, 100]
type = varset
trigger1 = !time
var(12) = floor(root, var(23) * 0.01)
[State 700, 10]
type = varset
trigger1 = !time
var(11) = floor(root, var(23) * 0.1) - (var(12) * 10)
[State 700, 1]
type = varset
trigger1 = !time
var(10) = root, var(23) - (var(11) * 10) - (var(12) * 100)

[State 700, Min]
type = varrangeset
trigger1 = root, var(23) <= 0
first = 10
last = 12
value = 0
[State 700, Max]
type = varrangeset
trigger1 = root, var(23) >= 999
first = 10
last = 12
value = 9

[State 700, 1]
type = explod
trigger1 = !numexplod(710)
anim = 710 + var(10)
ID = 710
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 94, 288), 59
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 700, 10]
type = explod
trigger1 = !numexplod(711)
trigger1 = var(11) || var(12)
anim = 710 + var(11)
ID = 711
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 79, 273), 59
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 700, 100]
type = explod
trigger1 = !numexplod(712)
trigger1 = var(12)
anim = 710 + var(12)
ID = 712
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 64, 258), 59
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1


;--- Success Rate ---
[State 700, 100]
type = varset
trigger1 = !time
var(22) = floor(root, var(22) * 0.01)
[State 700, 10]
type = varset
trigger1 = !time
var(21) = floor(root, var(22) * 0.1) - (var(22) * 10)
[State 700, 1]
type = varset
trigger1 = !time
var(20) = root, var(22) - (var(21) * 10) - (var(22) * 100)

[State 700, Min]
type = varrangeset
trigger1 = root, var(22) <= 0
first = 20
last = 22
value = 0
[State 700, Max]
type = varset
trigger1 = root, var(22) >= 100
var(20) = 0
[State 700, Max]
type = varset
trigger1 = root, var(22) >= 100
var(21) = 0
[State 700, Max]
type = varset
trigger1 = root, var(22) >= 100
var(22) = 1

[State 700, 1]
type = explod
trigger1 = !numexplod(720)
anim = 710 + var(20)
ID = 720
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 94, 288), 88
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 700, 10]
type = explod
trigger1 = !numexplod(721)
trigger1 = var(21) || var(22)
anim = 710 + var(21)
ID = 721
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 79, 273), 88
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 700, 100]
type = explod
trigger1 = !numexplod(722)
trigger1 = var(22)
anim = 710 + var(22)
ID = 722
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 64, 258), 88
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1

;--- Score ---
[State 700, 100000]
type = varset
trigger1 = !time
var(35) = floor(root, var(24) * 0.00001)
[State 700, 10000]
type = varset
trigger1 = !time
var(34) = floor(root, var(24) * 0.0001) - (var(35) * 10)
[State 700, 1000]
type = varset
trigger1 = !time
var(33) = floor(root, var(24) * 0.001) - (var(34) * 10) -(var(35) * 100)
[State 700, 100]
type = varset
trigger1 = !time
var(32) = floor(root, var(24) * 0.01) - (var(33) * 10) -(var(34) * 100) -(var(35) * 1000)
[State 700, 10]
type = varset
trigger1 = !time
var(31) = floor(root, var(24) * 0.1) - (var(32) * 10) -(var(33) * 100) -(var(34) * 1000) -(var(35) * 10000)
[State 700, 1]
type = varset
trigger1 = !time
var(30) = root, var(24) - (var(31) * 10) -(var(32) * 100) -(var(33) * 1000) -(var(34) * 10000) -(var(35) * 100000)

[State 700, Min]
type = varrangeset
trigger1 = root, var(24) <= 0
first = 30
last = 35
value = 0
[State 700, Max]
type = varrangeset
trigger1 = root, var(24) >= 999999
first = 30
last = 35
value = 9

[State 700, 1]
type = explod
trigger1 = !numexplod(730)
anim = 710 + var(30)
ID = 730
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 94, 288), 30
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 700, 10]
type = explod
trigger1 = !numexplod(731)
trigger1 = var(31) || var(32) || var(33) || var(34) || var(35)
anim = 710 + var(31)
ID = 731
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 79, 273), 30
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 700, 100]
type = explod
trigger1 = !numexplod(732)
trigger1 = var(32) || var(33) || var(34) || var(35)
anim = 710 + var(32)
ID = 732
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 64, 258), 30
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 700, 1000]
type = explod
trigger1 = !numexplod(733)
trigger1 = var(33) || var(34) || var(35)
anim = 710 + var(33)
ID = 733
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 49, 243), 30
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 700, 10000]
type = explod
trigger1 = !numexplod(734)
trigger1 = var(34) || var(35)
anim = 710 + var(34)
ID = 734
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 34, 228), 30
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 700, 100000]
type = explod
trigger1 = !numexplod(735)
trigger1 = var(35)
anim = 710 + var(35)
ID = 735
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 19, 213), 30
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 700, End]
type = removeexplod
trigger1 = roundstate != 4
[State 700, End]
type = destroyself
trigger1 = roundstate != 4
trigger1 = !numexplod


;--- Intro ---
[Statedef 705]
movetype = I
physics = N
anim = 0

[State 705, NHB]
type = nothitby
trigger1 = 1
time = -1
value = SCA
ignorehitpause = 1

[State 705, No Bars]
type = assertspecial
trigger1 = !numpartner && numenemy <= 1
flag = nobardisplay
[State 705, No Bars]
type = assertspecial
triggerall = numpartner
trigger1 = partner, name = "Duck Hunt" && partner, authorname = "Phantom.of.the.Server"
trigger2 = partner, name = "Catwalk" && partner, authorname = "Phantom.of.the.Server"
trigger3 = partner, name = "SS1 Bonus" && partner, authorname = "Phantom.of.the.Server"
trigger4 = partner, name = "Sky Chase" && partner, authorname = "Phantom.of.the.Server"
flag = nobardisplay

[State 705, Frame]
type = explod
trigger1 = !numexplod(700)
anim = 700
ID = 700
sprpriority = 5
postype = left
pos = ifelse(teamside = 1, 12, 174) + ifelse(teamside = 1, 10 * 32, -10 * 32), 12 + 10 * 32
vel = ifelse(teamside = 1, -10, 10), -10
scale = 1, 1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 705, Fill]
type = explod
trigger1 = !numexplod(701)
anim = 505
ID = 701
sprpriority = 4
postype = left
pos = ifelse(teamside = 1, 12, 174) + ifelse(teamside = 1, 10 * 32, -10 * 32), 12 + 10 * 32
vel = ifelse(teamside = 1, -10, 10), -10
scale = 4.417, 4.292
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 705, End]
type = removeexplod
trigger1 = time >= 32
[State 705, End]
type = changestate
trigger1 = time >= 32
value = 700


;====================< ARROWS >====================
[Statedef 1000]
type = A
physics = N
movetype = I
anim = 0
ctrl = 0
velset = 0, 0

[State 1000, NHB]
type = nothitby
trigger1 = 1
value = SCA
time = -1
ignorehitpause = 1

[State 1000, AX]
type = varset
triggerall = !time
trigger1 = (root, var(5) = 1) && random < 200
trigger2 = (root, var(5) = 2)
trigger3 = (root, var(5) = 5) && (root, var(6) % 3) = 0
trigger4 = (root, var(5) = 6) && (root, var(6) % 3) = 1
trigger5 = (root, var(5) = 7) && (root, var(6) % 2) = 0
trigger6 = (root, var(5) = 9) && (root, var(6) % 2) = 0
trigger7 = (root, var(5) = 20) && ((root, var(6) % 4) = 0 || (root, var(6) % 4) = 1)
trigger8 = (root, var(5) = 21) && (root, var(6) % 2) = 0
trigger9 = (root, var(5) = 22)
trigger10 = (root, var(5) = 23) && (root, var(6) % 3) = 1
trigger11 = (root, var(5) = 24)
trigger12 = (root, var(5) = 25) && (root, var(6) % 3) = 0
trigger13 = (root, var(5) = 26)
trigger14 = (root, var(5) = 27)
trigger15 = (root, var(5) = 28)
trigger16 = (root, var(5) = 29) && (root, var(6) % 4) = 0
trigger17 = (root, var(5) = 30) && ((root, var(6) % 5) = 0 || (root, var(6) % 5) = 1 || (root, var(6) % 5) = 3 || (root, var(6) % 5) = 4)
var(10) = ifelse(root, var(7), 1001, 1000)
[State 1000, BY]
type = varset
triggerall = !time
trigger1 = (root, var(5) = 1) && random < 200
trigger2 = (root, var(5) = 3)
trigger3 = (root, var(5) = 5) && (root, var(6) % 3) = 2
trigger4 = (root, var(5) = 6) && (root, var(6) % 3) = 2
trigger5 = (root, var(5) = 7) && (root, var(6) % 2) = 1
trigger6 = (root, var(5) = 8) && (root, var(6) % 2) = 0
trigger7 = (root, var(5) = 20) && ((root, var(6) % 4) = 2 || (root, var(6) % 4) = 3)
trigger8 = (root, var(5) = 21)
trigger9 = (root, var(5) = 22) && ((root, var(6) % 3) = 1 || (root, var(6) % 3) = 2)
trigger10 = (root, var(5) = 23) && ((root, var(6) % 3) = 1 || (root, var(6) % 3) = 2)
trigger11 = (root, var(5) = 24) && ((root, var(6) % 3) = 0 || (root, var(6) % 3) = 2)
trigger12 = (root, var(5) = 25) && ((root, var(6) % 3) = 0 || (root, var(6) % 3) = 2)
trigger13 = (root, var(5) = 26)
trigger14 = (root, var(5) = 27) && ((root, var(6) % 5) = 0 || (root, var(6) % 5) = 3)
trigger15 = (root, var(5) = 28) && ((root, var(6) % 4) = 0 || (root, var(6) % 4) = 1)
trigger16 = (root, var(5) = 29)
trigger17 = (root, var(5) = 30) && ((root, var(6) % 5) = 0 || (root, var(6) % 5) = 1 || (root, var(6) % 5) = 3)
var(11) = ifelse(root, var(7), 1011, 1010)
[State 1000, CZ]
type = varset
triggerall = !time
trigger1 = (root, var(5) = 1) && random < 200
trigger2 = (root, var(5) = 4)
trigger3 = (root, var(5) = 5) && (root, var(6) % 3) = 1
trigger4 = (root, var(5) = 6) && (root, var(6) % 3) = 0
trigger5 = (root, var(5) = 8) && (root, var(6) % 2) = 1
trigger6 = (root, var(5) = 9) && (root, var(6) % 2) = 1
trigger7 = (root, var(5) = 20) && ((root, var(6) % 4) = 0 || (root, var(6) % 4) = 1)
trigger8 = (root, var(5) = 21) && (root, var(6) % 2) = 1
trigger9 = (root, var(5) = 22) && (root, var(6) % 3) = 1
trigger10 = (root, var(5) = 23)
trigger11 = (root, var(5) = 24) && (root, var(6) % 3) = 0
trigger12 = (root, var(5) = 25)
trigger13 = (root, var(5) = 26)
trigger14 = (root, var(5) = 27) && ((root, var(6) % 5) = 0 || (root, var(6) % 5) = 3 || (root, var(6) % 5) = 4)
trigger15 = (root, var(5) = 28)
trigger16 = (root, var(5) = 29) && (root, var(6) % 4) = 0
trigger17 = (root, var(5) = 30) && ((root, var(6) % 5) = 0 || (root, var(6) % 5) = 1 || (root, var(6) % 5) = 2 || (root, var(6) % 5) = 3)
var(12) = ifelse(root, var(7), 1021, 1020)

[State 1000, Random Direction]
type = varset
triggerall = !time
trigger1 = (root, var(5) = 1) && random < 333
var(13) = 1030 + (random % 4)
[State 1000, Up]
type = varset
triggerall = !time
trigger1 = (root, var(5) = 10)
trigger2 = (root, var(5) = 14) && (root, var(6) % 2) = 0
trigger3 = (root, var(5) = 16) && ((root, var(6) % 4) = 0 || (root, var(6) % 4) = 3)
trigger4 = (root, var(5) = 18) && (root, var(6) % 4) = 0
trigger5 = (root, var(5) = 19) && (root, var(6) % 4) = 3
trigger6 = (root, var(5) = 33) && (root, var(6) % 2) = 0
trigger7 = (root, var(5) = 34) && (root, var(6) % 2) = 0
var(13) = 1030
[State 1000, Down]
type = varset
triggerall = !time
trigger1 = (root, var(5) = 11)
trigger2 = (root, var(5) = 14) && (root, var(6) % 2) = 1
trigger3 = (root, var(5) = 16) && ((root, var(6) % 4) = 1 || (root, var(6) % 4) = 2)
trigger4 = (root, var(5) = 18) && (root, var(6) % 4) = 2
trigger5 = (root, var(5) = 19) && (root, var(6) % 4) = 3
trigger6 = (root, var(5) = 31) && (root, var(6) % 2) = 0
trigger7 = (root, var(5) = 32) && (root, var(6) % 2) = 0
var(13) = 1031
[State 1000, Left]
type = varset
triggerall = !time
trigger1 = (root, var(5) = 12)
trigger2 = (root, var(5) = 15) && (root, var(6) % 2) = 0
trigger3 = (root, var(5) = 17) && ((root, var(6) % 4) = 0 || (root, var(6) % 4) = 3)
trigger4 = (root, var(5) = 18) && (root, var(6) % 4) = 1
trigger5 = (root, var(5) = 19) && (root, var(6) % 4) = 0
trigger6 = (root, var(5) = 31) && (root, var(6) % 2) = 1
trigger7 = (root, var(5) = 33) && (root, var(6) % 2) = 1
var(13) = 1032
[State 1000, Right]
type = varset
triggerall = !time
trigger1 = (root, var(5) = 13)
trigger2 = (root, var(5) = 15) && (root, var(6) % 2) = 1
trigger3 = (root, var(5) = 17) && ((root, var(6) % 4) = 1 || (root, var(6) % 4) = 2)
trigger4 = (root, var(5) = 18) && (root, var(6) % 4) = 3
trigger5 = (root, var(5) = 19) && (root, var(6) % 4) = 2
trigger6 = (root, var(5) = 32) && (root, var(6) % 2) = 1
trigger7 = (root, var(5) = 34) && (root, var(6) % 2) = 1
var(13) = 1033

[State 1000, No Direction]
type = varset
trigger1 = root, var(53)
trigger1 = ((var(10) != 0) + (var(11) != 0) + (var(12) != 0)) >= 3
var(13) = 0

[State 1000, None]
type = destroyself
trigger1 = !var(10) && !var(11) && !var(12) && !var(13)

[State 1000, AX]
type = explod
trigger1 = var(10) = [1000, 1001]
trigger1 = !numexplod(1000)
anim = var(10)
ID = 1000
sprpriority = 6
postype = p1
pos = ifelse(root, var(51), 24, 0), 0
scale = 1, 1
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 1000, BY]
type = explod
trigger1 = var(11) = [1010, 1011]
trigger1 = !numexplod(1010)
anim = var(11)
ID = 1010
sprpriority = 6
postype = p1
pos = ifelse(root, var(51), 48, 24), 0
scale = 1, 1
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 1000, CZ]
type = explod
trigger1 = var(12) = [1020, 1021]
trigger1 = !numexplod(1020)
anim = var(12)
ID = 1020
sprpriority = 6
postype = p1
pos = ifelse(root, var(51), 72, 48), 0
scale = 1, 1
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 1000, Dir]
type = explod
trigger1 = var(13) = [1030, 1033]
trigger1 = !numexplod(1030)
anim = var(13)
ID = 1030
sprpriority = 6
postype = p1
pos = ifelse(root, var(51), 0, 72), 0
scale = 1, 1
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1

[State 1000, No Indicator]
type = varset
trigger1 = !numexplod(1000)
var(10) = 0
[State 1000, No Indicator]
type = varset
trigger1 = !numexplod(1010)
var(11) = 0
[State 1000, No Indicator]
type = varset
trigger1 = !numexplod(1020)
var(12) = 0
[State 1000, No Indicator]
type = varset
trigger1 = !numexplod(1030)
var(13) = 0

[State 1000, Pos]
type = varset
trigger1 = !time
fvar(0) = 240
[State 1000, Pos]
type = varadd
trigger1 = roundstate = 2
fvar(0) = (root, fvar(4))
[State 1000, Pos]
type = posset
trigger1 = 1
x = ifelse(teamside = 2, 47, -143)
y = ceil(pos y -screenpos y) + fvar(0)

[State 1000, Clean]
type = changestate
trigger1 = roundstate = 4
trigger2 = root, var(8) && fvar(0) <= 126
value = 1007

[State 1000, Check Input]
type = changestate
trigger1 = fvar(0) <= 25
value = 1001


;--- Input Check ---
[Statedef 1001]
type = A
physics = N
movetype = I
anim = 0
ctrl = 0
velset = 0, 0

[State 1001, NHB]
type = nothitby
trigger1 = 1
value = SCA
time = -1
ignorehitpause = 1

[State 1001, Wrong]
type = changestate
triggerall = root, var(52)
triggerall = playerIDexist(root, var(1))
trigger1 = var(10) != 1000 || var(20) = 1000
trigger1 = playerID(root, var(1)), command = "a"
trigger2 = var(11) != 1010 || var(21) = 1010
trigger2 = playerID(root, var(1)), command = "b"
trigger3 = var(12) != 1020 || var(22) = 1020
trigger3 = playerID(root, var(1)), command = "c"
trigger4 = var(10) != 1001 || var(20) = 1001
trigger4 = playerID(root, var(1)), command = "x"
trigger5 = var(11) != 1011 || var(21) = 1011
trigger5 = playerID(root, var(1)), command = "y"
trigger6 = var(12) != 1021 || var(22) = 1021
trigger6 = playerID(root, var(1)), command = "z"
trigger7 = var(13) != 1030 || var(23) = 1030
trigger7 = playerID(root, var(1)), command = "up"
trigger8 = var(13) != 1031 || var(23) = 1031
trigger8 = playerID(root, var(1)), command = "down"
trigger9 = var(13) != 1032 || var(23) = 1032
trigger9 = playerID(root, var(1)), command = "back" && playerID(root, var(1)), facing = 1
trigger10 = var(13) != 1032 || var(23) = 1032
trigger10 = playerID(root, var(1)), command = "fwd" && playerID(root, var(1)), facing = -1
trigger11 = var(13) != 1033 || var(23) = 1033
trigger11 = playerID(root, var(1)), command = "fwd" && playerID(root, var(1)), facing = 1
trigger12 = var(13) != 1033 || var(23) = 1033
trigger12 = playerID(root, var(1)), command = "back" && playerID(root, var(1)), facing = -1
value = 1006

[State 1001, A]
type = varset
trigger1 = var(10)
trigger1 = playerIDexist(root, var(1))
trigger1 = playerID(root, var(1)), command = "a"
var(20) = 1000
[State 1001, B]
type = varset
trigger1 = var(11)
trigger1 = playerIDexist(root, var(1))
trigger1 = playerID(root, var(1)), command = "b"
var(21) = 1010
[State 1001, C]
type = varset
trigger1 = var(12)
trigger1 = playerIDexist(root, var(1))
trigger1 = playerID(root, var(1)), command = "c"
var(22) = 1020

[State 1001, X]
type = varset
trigger1 = var(10)
trigger1 = playerIDexist(root, var(1))
trigger1 = playerID(root, var(1)), command = "x"
var(20) = 1001
[State 1001, Y]
type = varset
trigger1 = var(11)
trigger1 = playerIDexist(root, var(1))
trigger1 = playerID(root, var(1)), command = "y"
var(21) = 1011
[State 1001, Z]
type = varset
trigger1 = var(12)
trigger1 = playerIDexist(root, var(1))
trigger1 = playerID(root, var(1)), command = "z"
var(22) = 1021

[State 1001, Up]
type = varset
trigger1 = var(13)
trigger1 = playerIDexist(root, var(1))
trigger1 = playerID(root, var(1)), command = "up"
var(23) = 1030
[State 1001, Down]
type = varset
trigger1 = var(13)
trigger1 = playerIDexist(root, var(1))
trigger1 = playerID(root, var(1)), command = "down"
var(23) = 1031
[State 1001, Left]
type = varset
triggerall = var(13)
triggerall = playerIDexist(root, var(1))
trigger1 = playerID(root, var(1)), command = "back" && playerID(root, var(1)), facing = 1
trigger2 = playerID(root, var(1)), command = "fwd" && playerID(root, var(1)), facing = -1
var(23) = 1032
[State 1001, Right]
type = varset
triggerall = var(13)
triggerall = playerIDexist(root, var(1))
trigger1 = playerID(root, var(1)), command = "fwd" && playerID(root, var(1)), facing = 1
trigger2 = playerID(root, var(1)), command = "back" && playerID(root, var(1)), facing = -1
var(23) = 1033

[State 1001, Miss]
type = changestate
trigger1 = roundstate != 2
value = 1006

;[State 1001, Cheat]
;type = varset
;trigger1 = 1
;var(20) = var(10)
;[State 1001, Cheat]
;type = varset
;trigger1 = 1
;var(21) = var(11)
;[State 1001, Cheat]
;type = varset
;trigger1 = 1
;var(22) = var(12)
;[State 1001, Cheat]
;type = varset
;trigger1 = 1
;var(23) = var(13)

[State 1001, Hit]
type = changestate
trigger1 = var(10) || var(11) || var(12) || var(13)
trigger1 = var(10) = var(20)
trigger1 = var(11) = var(21)
trigger1 = var(12) = var(22)
trigger1 = var(13) = var(23)
value = 1005

[State 1001, Pos]
type = varadd
trigger1 = roundstate = 2
fvar(0) = (root, fvar(4))
[State 1001, Pos]
type = posset
trigger1 = 1
x = ifelse(teamside = 2, 47, -143)
y = ceil(pos y -screenpos y) + fvar(0)

[State 1001, Miss]
type = changestate
trigger1 = time
trigger1 = fvar(0) < 6
value = 1006

[State 1001, Clean]
type = changestate
trigger1 = roundstate = 4
trigger2 = root, var(8) && fvar(0) <= 126
value = 1007


;--- Hit ---
[Statedef 1005]
type = A
physics = N
movetype = I
anim = 0
ctrl = 0
velset = 0, 0

[State 1005, Hit]
type = parentvarset
trigger1 = !time
trigger1 = var(10) || var(11) || var(12) || var(13)
trigger1 = var(10) = var(20)
trigger1 = var(11) = var(21)
trigger1 = var(12) = var(22)
trigger1 = var(13) = var(23)
var(10) = 1

[State 1005, Hit]
type = parentvarset
trigger1 = !time
var(13) = (var(20) != 0) + (var(21) != 0) + (var(22) != 0) + (var(23) != 0)

[State 1005, P2 Anim]
type = parentvarset
trigger1 = !time
trigger1 = var(23)
var(11) = ifelse(var(23) = 1030, 2100, ifelse(var(23) = 1031, 2101, ifelse(var(23) = 1032, 2102, 2103)))

[State 1005, P2 Anim]
type = parentvarset
trigger1 = !time
trigger1 = var(20) || var(21) || var(22)
var(11) = ifelse(var(20), 2110, ifelse(var(21), 2111, 2112))

[State 1005, P2 Anim]
type = parentvarset
trigger1 = !time
trigger1 = var(23) = 1030
trigger1 = ((var(20) != 0) + (var(21) != 0) + (var(22) != 0)) = 1
var(11) = ifelse(var(20), 2120, ifelse(var(21), 2121, 2122))
[State 1005, P2 Anim]
type = parentvarset
trigger1 = !time
trigger1 = var(23) = 1031
trigger1 = ((var(20) != 0) + (var(21) != 0) + (var(22) != 0)) = 1
var(11) = ifelse(var(20), 2130, ifelse(var(21), 2131, 2132))

[State 1005, P2 Anim]
type = parentvarset
trigger1 = !time
trigger1 = ((var(20) != 0) + (var(21) != 0) + (var(22) != 0)) = 2
var(11) = 2140
[State 1005, P2 Anim]
type = parentvarset
trigger1 = !time
trigger1 = ((var(20) != 0) + (var(21) != 0) + (var(22) != 0) + (var(23) != 0)) = 3
var(11) = 2150
[State 1005, P2 Anim]
type = parentvarset
trigger1 = !time
trigger1 = ((var(20) != 0) + (var(21) != 0) + (var(22) != 0) + (var(23) != 0)) = 4
var(11) = 2160

[State 1005, FX]
type = modifyexplod
trigger1 = 1
scale = 1, cos(time / 12.0 * pi)
trans = addalpha
alpha = 256 -(10 * time), (10 * time)

[State 1005, End]
type = removeexplod
trigger1 = time >= 24
[State 1005, End]
type = destroyself
trigger1 = !numexplod


;--- Miss ---
[Statedef 1006]
type = A
physics = N
movetype = I
anim = 0
ctrl = 0
velset = 0, 0

[State 1006, Hit]
type = parentvarset
trigger1 = !time
var(10) = -1

[State 1006, Pos]
type = varadd
trigger1 = 1
fvar(0) = (root, fvar(4))
[State 1006, Pos]
type = posset
trigger1 = 1
x = ifelse(teamside = 2, 47, -143)
y = ceil(pos y -screenpos y) + fvar(0)

[State 1006, FX]
type = modifyexplod
trigger1 = 1
trans = addalpha
alpha = 256 -(10 * time), (10 * time)

[State 1006, End]
type = removeexplod
trigger1 = time >= 24
[State 1006, End]
type = destroyself
trigger1 = !numexplod


;--- Clean ---
[Statedef 1007]
type = A
physics = N
movetype = I
anim = 0
ctrl = 0
velset = 0, 0

[State 1007, FX]
type = modifyexplod
trigger1 = 1
trans = addalpha
alpha = 256 -(32 * time), (32 * time)

[State 1007, End]
type = removeexplod
trigger1 = time >= 8
[State 1007, End]
type = destroyself
trigger1 = !numexplod


;====================< PLAYER >====================
;--- Idle ---
[Statedef 2000]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2000, Turn]
type = turn
trigger1 = teamside = 1 && facing != 1
trigger2 = teamside = 2 && facing != -1
[State 2000, Pos]
type = posset
trigger1 = 1
x = ifelse(teamside = 1, -60, 60)
y = 0

[State 2000, OK]
type = changeanim
trigger1 = anim != 0
value = 0


;--- Dizzy ---
[Statedef 2001]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2001, Turn]
type = turn
trigger1 = teamside = 1 && facing != 1
trigger2 = teamside = 2 && facing != -1
[State 2001, Pos]
type = posset
trigger1 = 1
x = ifelse(teamside = 1, -60, 60)
y = 0

[State 2001, Dizzy]
type = changeanim
trigger1 = anim != 5300
trigger1 = selfanimexist(5300)
value = 5300
[State 2001, Dizzy]
type = changeanim2
trigger1 = anim != 5301
trigger1 = !selfanimexist(5300)
value = 5301


;--- Miss ---
[Statedef 2010]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2010, Anim]
type = changeanim
trigger1 = !time
value = ifelse(selfanimexist(5910) && random < 333, 5910, ifelse(random < 500, 5012, 5002))
[State 2010, Anim]
type = changeanim
trigger1 = anim = 5002 || anim = 5012
trigger1 = !animtime
value = anim + 5

[State 2010, Snd]
type = playsnd
triggerall = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger2 = authorname = "Warusaki3"
trigger3 = authorname = "Sepp"
trigger4 = authorname = "Mr.Fong"
trigger5 = authorname = "GM"
trigger6 = authorname = "Jesuszilla"
value = 5000, (random % 3)
channel = 0
[State 2010, Snd]
type = playsnd
trigger1 = !time
trigger1 = name = "Kung Fu Man"
value = 100, 0
channel = 0
[State 2010, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "N -Mario" || authorname = "N64Mario"
value = 12, (random % 3)
channel = 0

[State 2010, End]
type = changestate
trigger1 = time >= 16
value = 2000


;--- Intro ---
[Statedef 2030]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2030, Camera]
type = screenbound
trigger1 = 1
value = 1
movecamera = 1, 0

[State 2030, Vel]
type = velset
trigger1 = !time && teamside = 1
x = (-60 -pos x) / 40.0
y = -10.25
[State 2030, Vel]
type = velset
trigger1 = !time && teamside = 2
x = (60 -pos x) / 40.0
y = -10.25
[State 2030, Vel]
type = veladd
trigger1 = 1
y = 0.5

[State 2030, Anim]
type = changeanim
trigger1 = !time
value = ifelse(vel x > 0, 42, ifelse(vel x < 0, 43, 41))
[State 2030, Anim]
type = changeanim
trigger1 = anim = 41 || anim = 42 || anim = 43
trigger1 = vel y > -2 && selfanimexist(anim + 3)
value = anim + 3

[State 2030, Land]
type = changestate
trigger1 = time
trigger1 = pos y > -vel y
value = 2031


[Statedef 2031]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2031, Pos]
type = posset
trigger1 = 1
x = ifelse(teamside = 1, -60, 60)
y = 0

[State 2031, Anim]
type = changeanim
trigger1 = anim != 47
value = 47

[State 2031, End]
type = changestate
trigger1 = anim = 47 && !animtime
trigger2 = time >= 15
value = 2000


;--- Win ---
[Statedef 2050]
type = A
physics = N
movetype = H
ctrl = 0
[State 2050, End]
type = selfstate
trigger1 = 1
value = 0
ctrl = 1


;--- Lose ---
[Statedef 2060]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2060, Collapse]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(5950)
value = 5950
[State 2060, Collapse]
type = changeanim2
trigger1 = !time
trigger1 = !selfanimexist(5950)
value = 5951

[State 2060, Anim]
type = changeanim
triggerall = anim = 5950 || anim = 5951
trigger1 = !animtime
trigger2 = animelemtime(animelemno(0)) >= 4 && !animelemtime(animelemno(0) + 1)
value = ifelse(selfanimexist(5111), 5111, 5110)

[State 2060, End]
type = selfstate
trigger1 = time >= 120
value = 5120


;--- Up ---
[Statedef 2100]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2100, Anim]
type = changeanim
trigger1 = !time
value = ifelse(random < 333, 43, ifelse(random < 500, 42, 41))

[State 2100, End]
type = changestate
trigger1 = 1
value = 2020


;--- Down ---
[Statedef 2101]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2101, Anim]
type = changeanim
trigger1 = !time
value = ifelse(random < 500, 12, 5120)

[State 2101, End]
type = changestate
trigger1 = 1
value = 2020


;--- Left ---
[Statedef 2102]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2102, Anim]
type = changeanim
trigger1 = !time
value = 105

[State 2102, End]
type = changestate
trigger1 = 1
value = 2020


;--- Right ---
[Statedef 2103]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2103, Anim]
type = changeanim
trigger1 = !time
value = 100
[State 2103, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(102) && random < 500
value = 102

[State 2103, End]
type = changestate
trigger1 = 1
value = 2020


;--- Weak Attack ---
[Statedef 2110]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2110, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(200)
value = 200
[State 2110, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(230) && random < 500
value = 230

[State 2110, Anim]
type = changeanim
triggerall = !time
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(200)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 200
[State 2110, Anim]
type = changeanim
triggerall = !time && random < 500
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(215)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 215

[State 2110, End]
type = changestate
trigger1 = 1
value = 2020


;--- Medium Attack ---
[Statedef 2111]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2111, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(210)
value = 210
[State 2111, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(240) && random < 500
value = 240

[State 2111, Anim]
type = changeanim
triggerall = !time
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(205)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 205
[State 2111, Anim]
type = changeanim
triggerall = !time && random < 500
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(220)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 220

[State 2111, End]
type = changestate
trigger1 = 1
value = 2020


;--- Strong Attack ---
[Statedef 2112]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2112, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(220)
value = 220
[State 2112, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(250) && random < 500
value = 250

[State 2112, Anim]
type = changeanim
triggerall = !time
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(210)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 210
[State 2112, Anim]
type = changeanim
triggerall = !time && random < 500
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(225)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 225

[State 2112, End]
type = changestate
trigger1 = 1
value = 2020


;--- Jumping Weak Attack ---
[Statedef 2120]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2120, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(600)
value = 600
[State 2120, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(630) && random < 500
value = 630

[State 2120, Anim]
type = changeanim
triggerall = !time
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(260)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 260
[State 2120, Anim]
type = changeanim
triggerall = !time && random < 500
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(275)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 275

[State 2120, End]
type = changestate
trigger1 = 1
value = 2020


;--- Jumping Medium Attack ---
[Statedef 2121]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2121, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(610)
value = 610
[State 2121, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(640) && random < 500
value = 640

[State 2121, Anim]
type = changeanim
triggerall = !time
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(265)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 265
[State 2121, Anim]
type = changeanim
triggerall = !time && random < 500
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(280)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 280

[State 2121, End]
type = changestate
trigger1 = 1
value = 2020


;--- Jumping Strong Attack ---
[Statedef 2122]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2122, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(620)
value = 620
[State 2122, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(650) && random < 500
value = 650

[State 2122, Anim]
type = changeanim
triggerall = !time
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(270)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 270
[State 2122, Anim]
type = changeanim
triggerall = !time && random < 500
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(285)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 285

[State 2122, End]
type = changestate
trigger1 = 1
value = 2020


;--- Crouching Weak Attack ---
[Statedef 2130]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2130, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(400)
value = 40
[State 2130, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(430) && random < 500
value = 430

[State 2130, Anim]
type = changeanim
triggerall = !time
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(230)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 230
[State 2130, Anim]
type = changeanim
triggerall = !time && random < 500
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(245)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 245

[State 2130, End]
type = changestate
trigger1 = 1
value = 2020


;--- Crouching Medium Attack ---
[Statedef 2131]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2131, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(410)
value = 410
[State 2131, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(440) && random < 500
value = 440

[State 2131, Anim]
type = changeanim
triggerall = !time
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(235)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 235
[State 2131, Anim]
type = changeanim
triggerall = !time && random < 500
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(250)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 250

[State 2131, End]
type = changestate
trigger1 = 1
value = 2020


;--- Crouching Strong Attack ---
[Statedef 2132]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2132, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(420)
value = 420
[State 2132, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(450) && random < 500
value = 450

[State 2132, Anim]
type = changeanim
triggerall = !time
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(240)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 240
[State 2132, Anim]
type = changeanim
triggerall = !time && random < 500
triggerall = authorname = "Phantom.of.the.Server" && selfanimexist(255)
trigger1 = name = "Pocket Shin Gouki" || name = "Guy" || name = "Ryu" || name = "Shin Gouki"
trigger2 = name = "Lei -Lei" || name = "Dan Hibiki" || name = "Morrigan Aensland"
trigger3 = name = "Sakura Kasugano" || name = "Vega" || name = "Rare Akuma"
value = 255

[State 2132, End]
type = changestate
trigger1 = 1
value = 2020


;--- Special ---
[Statedef 2140]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2140, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(1000)
value = 1000
[State 2140, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(1100) && random < 200
value = 1100
[State 2140, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(1200) && random < 200
value = 1200
[State 2140, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(1300) && random < 200
value = 1200
[State 2140, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(1400) && random < 200
value = 1200

[State 2140, End]
type = changestate
trigger1 = 1
value = 2020


;--- Pose ---
[Statedef 2150]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2150, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(195)
value = 195

[State 2150, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(181) && random < 200
value = 181
[State 2150, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(182) && random < 200
value = 182
[State 2150, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(183) && random < 200
value = 183
[State 2150, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(184) && random < 200
value = 184
[State 2150, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(185) && random < 200
value = 185

[State 2150, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(190) && random < 200
value = 190
[State 2150, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(191) && random < 200
value = 191
[State 2150, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(192) && random < 200
value = 192

[State 2150, End]
type = changestate
trigger1 = 1
value = 2020


;--- Super ---
[Statedef 2160]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2160, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(3000)
value = 3000
[State 2160, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(3100) && random < 333
value = 3100
[State 2160, Anim]
type = changeanim
trigger1 = !time
trigger1 = selfanimexist(3200) && random < 333
value = 3200

[State 2160, End]
type = changestate
trigger1 = 1
value = 2020


;--- Hit Shout /safety ---
[Statedef 2020]
type = A
physics = N
movetype = H
ctrl = 0
velset = 0, 0
[State NHB]
type = nothitby
trigger1 = 1
time = 1
value = SCA
ignorehitpause = 1

[State 2020, Anim]
type = changeanim
trigger1 = !time
trigger1 = anim = 0 || anim = 5300 || anim = 5301
value = ifelse(random < 333, 100, ifelse(random < 500, 41, 5120))

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
value = 4, (random % 6)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Pocket Shin Gouki"
value = 3, (random % 15)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 500
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Pocket Shin Gouki"
value = 4, (random % 36)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Guy"
value = 3, (random % 9)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 500
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Guy"
value = 4, (random % 4)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Ryu"
value = 3 + (var(1) = 2) * 10, (random % 8)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 619
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Ryu"
value = 4 + (var(1) = 2) * 10, ifelse((var(1) = 2), (random % 13), (random % 18))
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Shin Gouki"
value = 3, (random % 12)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 571
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Shin Gouki"
value = 4, (random % 16)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Lei -Lei"
value = 3, (random % 13)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 435
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Lei -Lei"
value = 4, (random % 10)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Dan Hibiki"
value = 3, (random % 8)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 837
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Dan Hibiki"
value = 4, (random % 41)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Morrigan Aensland"
value = 3, (random % 6)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 778
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Morrigan Aensland"
value = 4, (random % 21)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Sakura Kasugano"
value = 3, (random % 8)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 742
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Sakura Kasugano"
value = 4, (random % 23)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Vega"
value = 3, (random % 9)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 550
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Vega"
value = 4, (random % 11)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Yuri Sakazaki"
value = 4, (random % 6)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 684
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Yuri Sakazaki"
value = 5, (random % 13)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Nash"
value = 4, (random % 9)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 438
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Nash"
value = 5, (random % 7)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Karin Kanzuki"
value = 4, (random % 13)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 435
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Karin Kanzuki"
value = 5, (random % 10)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Rare Akuma"
value = 3, (random % 12)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 571
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Rare Akuma"
value = 4, (random % 16)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Nakoruru"
value = 5, (random % 11)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 181
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Nakoruru"
value = 6, (random % 6)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 227
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Nakoruru"
value = 4, (random % 5)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Geese Howard"
value = 4, (random % 8)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 333
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Geese Howard"
value = 5, (random % 13)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 462
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Geese Howard"
value = 6, (random % 18)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Cammy"
value = 4, (random % 8)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 158
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Cammy"
value = 5, (random % 13)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time && random < 368
trigger1 = authorname = "Phantom.of.the.Server"
trigger1 = name = "Cammy"
value = 6, (random % 18)
channel = 0

[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = name = "Kung Fu Man" && random < 500
value = 5, 0
channel = 0
[State 2020, Snd]
type = playsnd
triggerall = !time
trigger1 = authorname = "Warusaki3"
trigger2 = authorname = "Sepp"
trigger3 = authorname = "GM"
value = 0, (random % 5)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "The Necromancer" || authorname = "FlowaGirl"
value = 0, (random % 20)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = name = "Homer J Simpson" && authorname = "Warner el Tochix"
value = 0, ifelse(random < 571, (random % 13), (random % 9) + 14)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Baby Bonnie Hood"
value = 1, (random % 15)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Jesuszilla" || authorname = "[Kira]"
value = 8, (random % 10)
channel = 0
[State 2020, Snd]
type = playsnd
triggerall = !time
trigger1 = authorname = "N -Mario" || authorname = "N64Mario"
trigger2 = authorname = "jin"
value = 10, (random % 15)
channel = 0
[State 2020, Snd]
type = playsnd
trigger1 = !time
trigger1 = authorname = "Andrés Borghi"
value = 15, (random % 5)
channel = 0

[State 2020, Weird Anim]
type = changeanim
trigger1 = name = "Kung Fu Man"
trigger1 = anim = 191 || anim = 192
value = 190

[State 2020, Speed Up]
type = changeanim
trigger1 = selfanimexist(anim)
trigger1 = animelemtime(animelemno(0) + 1)
trigger1 = animelemtime(animelemno(0) + 1) < -12
value = anim
elem = animelemno(0) + 1

[State 2020, End]
type = changestate
trigger1 = !animtime
trigger2 = time >= 16
trigger2 = animelemtime(animelemno(0)) >= 5 && !animelemtime(animelemno(0) + 1)
value = 2000


;====================< DEAD >====================
[Statedef 5150]
type = A
movetype = H
physics = N
anim = 0

[State 5150, safety]
type = changestate
trigger1 = numtarget(1)
value = 5151

[State 5150, hurt]
type = lifeset
trigger1 = 1
value = 0
kill = 1


;--- Drop Target ---
[Statedef 5151]
type = A
movetype = I
physics = N
anim = 0

[State 5151, P2 State]
type = targetstate
trigger1 = numtarget(1)
ID = 1
value = ifelse(var(2) <= 0, 2060, 2050)

[State 5151, P2 State]
type = targetdrop
trigger1 = 1
[State 5151, End]
type = changestate
trigger1 = !numtarget(1)
value = 5150


;==========< COMMON ERASE >==========
[Statedef 10]
[State 10, End]
type = changestate
trigger1 = 1
value = 0
[Statedef 11]
[State 11, End]
type = changestate
trigger1 = 1
value = 0
[Statedef 12]
[State 12, End]
type = changestate
trigger1 = 1
value = 0

[Statedef 20]
[State 20, End]
type = changestate
trigger1 = 1
value = 0

[Statedef 40]
[State 40, End]
type = changestate
trigger1 = 1
value = 0
[Statedef 50]
[State 50, End]
type = changestate
trigger1 = 1
value = 0
[Statedef 51]
[State 51, End]
type = changestate
trigger1 = 1
value = 0
[Statedef 52]
[State 52, End]
type = changestate
trigger1 = 1
value = 0

[Statedef 45]
[State 45, End]
type = changestate
trigger1 = 1
value = 0

[Statedef 100]
[State 100, End]
type = changestate
trigger1 = 1
value = 0
[Statedef 105]
[State 105, End]
type = changestate
trigger1 = 1
value = 0


;==========< Initialize >==========
[Statedef 5900]
type = A

[State 5900, var reset]
type = varrangeset
trigger1 = !roundsexisted
value = 0
[State 5900, var reset]
type = varrangeset
trigger1 = !roundsexisted
fvalue = 0

[State 5900, Palette]
type = remappal
trigger1 = 1
source = 1, 1
dest = 1, palno

[State 5900, End]
type = changestate
trigger1 = 1
value = 5901


;--- Second ---
[Statedef 5902]
type = A

[State 5902, Reset Life]
type = varset
trigger1 = 1
var(2) = 60
[State 5902, Reset Time]
type = varset
trigger1 = 1
fvar(2) = fvar(1)
[State 5902, Reset Hit Time]
type = varset
trigger1 = 1
var(10) = 0

[State 5902, Speed Fix]
type = varset
trigger1 = 1
fvar(3) = -abs(fvar(3))

[State 5902, End]
type = changestate
trigger1 = 1
value = 190


;====================< INTRO >====================
[Statedef 190]
type = A
physics = N
movetype = I
anim = 0
ctrl = 0

[State 190, End]
type = changestate
trigger1 = ctrl
value = 0


;====================< TIME OVER >====================
[Statedef 170]
physics = N
movetype = I
ctrl = 1
[State 170, End]
type = changestate
trigger1 = 1
value = 5151

[Statedef 175]
physics = N
movetype = I
ctrl = 1
[State 175, End]
type = changestate
trigger1 = 1
value = 5151


;====================< WIN >====================
[Statedef 180]
physics = N
movetype = I
ctrl = 0

[State 180, Quote]
type = victoryquote
trigger1 = 1
value = 1


;====================< ERASE HITS >====================
[Statedef 5000]
movetype = H
[State 5000, override]
type = changestate
trigger1 = 1
value = 0
[Statedef 5001]
movetype = H
[State 5001, override]
type = changestate
trigger1 = 1
value = 0
[Statedef 5002]
movetype = H
[State 5002, override]
type = changestate
trigger1 = 1
value = 0
[Statedef 5010]
movetype = H
[State 5010, override]
type = changestate
trigger1 = 1
value = 0
[Statedef 5020]
movetype = H
[State 5020, override]
type = changestate
trigger1 = 1
value = 0
[Statedef 5030]
movetype = H
[State 5030, override]
type = changestate
trigger1 = 1
value = 0
[Statedef 5050]
movetype = H
[State 5050, override]
type = changestate
trigger1 = 1
value = 0
[Statedef 5070]
movetype = H
[State 5070, override]
type = changestate
trigger1 = 1
value = 0
[Statedef 5080]
movetype = H
[State 5080, override]
type = changestate
trigger1 = 1
value = 0
[Statedef 5100]
movetype = H
[State 5100, override]
type = changestate
trigger1 = 1
value = 0


;===========================================================================
;===============================< -2 STATES >=================================
;===========================================================================
[Statedef -2]

[State -2, NotHitBy]
type = nothitby
trigger1 = 1
time = -1
value = SCA, AA, AP, AT
ignorehitpause = 1
[State -2, Life]
type = lifeset
trigger1 = 1
trigger1 = roundstate >= 3 && life > 1
value = lifemax

[State -2, Turn]
type = turn
trigger1 = facing != 1
[State -2, Pos]
type = posset
trigger1 = roundstate >= 2
x = 0
y = 0

[State -2, No Push]
type = playerpush
trigger1 = 1
value = 0
[State -2, Not Bound]
type = screenbound
trigger1 = roundstate >= 2
value = 0
movecamera = 1, 0
[State -2, Camera]
type = width
trigger1 = !numpartner && roundstate >= 2
edge = 320, 320

[State -2, Ctrl]
type = changestate
trigger1 = ctrl
value = 200
[State -2, Ctrl]
type = ctrlset
trigger1 = ctrl
value = 0

[State 2, Time]
type = varset
trigger1 = fvar(1) != -1
trigger1 = fvar(2) > (fvar(1) + 5.0)
fvar(2) = fvar(1)
[State -2, Time]
type = varset
trigger1 = fvar(2) = (-1, 0]
fvar(2) = 0
[State -2, Time]
type = assertspecial
trigger1 = alive
flag = timerfreeze

[State -2, Success Rate]
type = varset
trigger1 = var(20) || var(21)
var(22) = floor(100 * (var(20) * 1.0 / (var(20) + var(21))))
[State -2, Max Combo]
type = varset
trigger1 = var(23) < var(12)
var(23) = var(12)

[State -2, Clean Timer]
type = varadd
trigger1 = var(8)
var(8) = ifelse(var(8) > 0, -1, 1)

[State -2, Assert]
type = assertspecial
trigger1 = alive && !numpartner
flag = nokoslow

[State -2, Controlled KO]
type = lifeset
trigger1 = life <= 1
value = 1
kill = 0

[State -2, Display]
type = helper
trigger1 = roundstate != 4
trigger1 = !numhelper(500)
helpertype = normal
stateno = ifelse(roundstate < 2, 505, 500)
ID = 500
name = "Display"
postype = left
pos = 160, 0
pausemovetime = 2147483647
supermovetime = 2147483647
ownpal = 1
[State -2, Outcome]
type = helper
trigger1 = roundstate = 4
trigger1 = !numhelper(700)
helpertype = normal
stateno = 705
ID = 700
name = "Outcome"
postype = left
pos = 160, 0
pausemovetime = 2147483647
supermovetime = 2147483647
ownpal = 1

[State -2, Debug]
type = displaytoclipboard
trigger1 = 1
text = "Hit = %d Miss = %d Success = %d Max = %d"
params = var(20), var(21), var(22), var(23)
ignorehitpause = 1

[State -2, Debug]
type = appendtoclipboard
trigger1 = 1
text = "\nBy PotS - http://www.mugenguild.com/pots/"
ignorehitpause = 1

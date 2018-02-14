; _____________________________________
;| Shin Gouki by Phantom.of.the.Server |
; ¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯
;==============================================================================================
;=======================================< HELPER STATES >========================================
;==============================================================================================

;========================< GOU HADOU KEN >=============================
[Statedef 1005]
type = S
physics = N
movetype = A
anim = 1005
ctrl = 0
velset = 0, 0
sprpriority = 4

[State 1005, HitOverride]
type = hitoverride
trigger1 = 1
time = -1
attr = SCA, AA, AP, AT
stateno = 1006
ignorehitpause = 1
[State 1005, HB]
type = hitby
trigger1 = 1
time = -1
value = SCA, NP, SP, HP
ignorehitpause = 1

[State 1005, Strength]
type = varset
trigger1 = !time
var(10) = root, var(10)
[State 1005, movehitvar]
type = varset
trigger1 = movecontact && numtarget
trigger1 = !(target, ishelper)
var(3) = 1
ignorehitpause = 1
[State 1005, timevar]
type = varadd
trigger1 = time && var(1) < 25
var(1) = 1

[State 1005, Vel]
type = velset
trigger1 = 1
x = ifelse(var(10) = 3, 6, ifelse(var(10) = 2, 4.5, 3))
y = 0

[State 1005, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1

[State 1005, HitSound]
type = playsnd
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
value = 2, 7
channel = 1
persistent = 0
ignorehitpause = 1

[State 1005, Hit]
type = hitdef
trigger1 = 1
priority = 4, Hit
attr = S, SP
damage = ceil(ifelse(time <= 1, 68, 54) * root, fvar(12)), 7
animtype = Hard
hitflag = MAF
guardflag = MA
pausetime = 0, 4
sparkno = -1
guard.sparkno = s7001
guardsound = s2, 0
ground.type = High
ground.slidetime = 16
ground.hittime = 16
ground.velocity = ifelse(var(1) <= 12, -4, -10.52), ifelse(var(1) <= 12, -2.5, 0)
ground.cornerpush.veloff = 0
air.velocity = -.75 * vel x, -2
fall = var(1) <= 12
air.fall = 1
fall.recover = 1
fall.recovertime = 16
getpower = ifelse(p2movetype = H, 72, 144) * !(root, var(20)), 36 * !(root, var(20))
givepower = 36, 36
envshake.time = 14
envshake.freq = 120
envshake.ampl = 4
sprpriority = 4
airguard.ctrltime = 16
airguard.velocity = ifelse((root, var(20)), -4, -8), -1.5
forcestand = 1

[State 1005, Sparks]
type = explod
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
anim = 7020 + (random % 8)
ID = 7000
sprpriority = 5
postype = left
pos = ceil((target, pos x) -((target, const(size.ground.front) -10) * facing) + 160), ceil(screenpos y)
removetime = -2
ownpal = 1
facing = -1
ignorehitpause = 1
persistent = 0
pausemovetime = 100
supermovetime = 100

[State 1005, End]
type = changestate
trigger1 = movecontact
value = 1006
ignorehitpause = 1
[State 1006, End]
type = destroyself
trigger1 = pos x != [ -240, 240]

[Statedef 1006]
type = A
physics = N
movetype = I
anim = 3006
velset = 0, 0
ctrl = 0
sprpriority = 4
movehitpersist = 1
[State 1006, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1
[State 1006, Snd]
type = playsnd
trigger1 = !time
value = 5, 1
ignorehitpause = 1
persistent = 0
[State 1006, End]
type = destroyself
trigger1 = !animtime


;========================< SHAKUNETSU HADOU KEN >=============================
[Statedef 1025]
type = S
physics = N
movetype = A
ctrl = 0
velset = 0, 0
sprpriority = 4

[State 1025, Strength]
type = varset
trigger1 = !time && !prevstateno
var(10) = root, var(10)
[State 1025, Anim]
type = changeanim
trigger1 = !time && !prevstateno
value = 1025

[State 1025, HitOverride]
type = hitoverride
trigger1 = 1
time = -1
attr = SCA, AA, AP, AT
stateno = ifelse(var(2) < var(10), 1025, 1026)
ignorehitpause = 1
[State 1025, hitby]
type = hitby
trigger1 = 1
time = -1
value = SCA, NP, SP, HP
ignorehitpause = 1

[State 1025, hitvar]
type = varadd
trigger1 = !time
var(2) = 1
[State 1025, movehitvar]
type = varset
trigger1 = movecontact && numtarget
trigger1 = !(target, ishelper)
var(3) = 1
ignorehitpause = 1

[State 1025, Vel]
type = velset
trigger1 = 1
x = ifelse((time < 2 && var(2) > 1), 1, ifelse(var(10) = 3, 7, ifelse(var(10) = 2, 5.25, 3.5)))
y = 0

[State 1025, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1

[State 1025, HitSound]
type = playsnd
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
value = 2, 7
channel = 1
persistent = 0
ignorehitpause = 1

[State 1025, Hit]
type = hitdef
trigger1 = 1
priority = 4, Hit
attr = S, SP
damage = ceil(ifelse(var(10) = 1, 68, ifelse(var(10) = 2, 37, 27)) * root, fvar(12)), 7
animtype = Back
hitflag = MAF
guardflag = MA
pausetime = 4 * (var(2) < var(10)), 6
sparkno = -1
sparkxy = -10, 0
guard.sparkno = s7001
guardsound = s2, 0
ground.type = Low
ground.slidetime = 16
ground.hittime = 16
ground.velocity = -4.5, -3
ground.cornerpush.veloff = 0
guard.velocity = -10.52
air.velocity = -4.5, -3
fall = 1
fall.recover = 1
fall.recovertime = 24
getpower = ifelse(p2movetype = H, 72, 144) * !(root, var(20)), 36 * !(root, var(20))
givepower = 36, 36
envshake.time = 14
envshake.freq = 120
envshake.ampl = 4
palfx.color = 32
palfx.sinadd = 80, 25, 0, 8
palfx.mul = 512, 256, 128
palfx.time = 32
sprpriority = 4
airguard.ctrltime = 16
airguard.velocity = ifelse((root, var(20)), -4, -8), -1.5
forcestand = 1

[State 1025, Sparks]
type = explod
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
anim = 7020 + (random % 8)
ID = 7000
sprpriority = 5
postype = left
pos = ceil((target, pos x) -((target, const(size.ground.front) -10) * facing) + 160), ceil(screenpos y)
removetime = -2
ownpal = 1
facing = -1
ignorehitpause = 1
persistent = 0
pausemovetime = 100
supermovetime = 100

[State 1025, flames]
type = helper
trigger1 = movehit && numtarget && numhelper(7180) < 1
trigger1 = !(target, time) && (target, movetype = H) && !var(16)
helpertype = normal
stateno = 7180
ID = 7180
name = "flames"
postype = p2
bindtime = 1000
ownpal = 1
facing = -1
size.ground.front = 0
size.ground.back = 0
ignorehitpause = 1

[State 1025, End]
type = changestate
trigger1 = movecontact
value = ifelse(var(2) < var(10), 1025, 1026)
[State 1025, End]
type = destroyself
trigger1 = pos x != [ -240, 240]

[Statedef 1026]
type = A
physics = N
movetype = I
anim = 1026
ctrl = 0
velset = 0, 0
sprpriority = 4
[State 1026, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1
[State 1026, Snd]
type = playsnd
trigger1 = !time
value = 5, 2
ignorehitpause = 1
persistent = 0
[State 1026, End]
type = destroyself
trigger1 = !animtime


;========================< ZANKUU HADOUKEN >=============================
[Statedef 1055]
type = A
physics = N
movetype = A
anim = 1055
ctrl = 0
velset = 0, 0
sprpriority = 4

[State 1055, HitOverride]
type = hitoverride
trigger1 = 1
time = -1
attr = SCA, AA, AP, AT
stateno = 1056
ignorehitpause = 1
[State 1055, hitby]
type = hitby
trigger1 = 1
time = -1
value = SCA, NP, SP, HP
ignorehitpause = 1

[State 1055, Strength]
type = varset
trigger1 = !time
var(10) = root, var(10)
[State 1055, movehitvar]
type = varset
trigger1 = movecontact && numtarget
trigger1 = !(target, ishelper)
var(3) = 1
ignorehitpause = 1

[State 1055, Vel]
type = velset
trigger1 = 1
x = ifelse(var(10) = 3, 5.7, ifelse(var(10) = 2, 4.3, 2.9))
y = ifelse(var(10) = 3, 4, ifelse(var(10) = 2, 3, 2))

[State 1055, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1

[State 1055, HitSound]
type = playsnd
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
value = 2, 7
channel = 1
persistent = 0
ignorehitpause = 1

[State 1055, Hit]
type = hitdef
trigger1 = 1
priority = 4, Hit
attr = A, SP
damage = ceil(47 * root, fvar(12)), 7
animtype = Back
fall.animtype = Back
hitflag = MAFD
guardflag = MA
pausetime = 0, 8
sparkno = -1
sparkxy = -10, 0
guard.sparkno = s7001
guardsound = s2, 0
ground.type = High
ground.slidetime = 16
ground.hittime = 16
ground.velocity = -7.51, 0
guard.velocity = -10.52
ground.cornerpush.veloff = 0
air.velocity = -vel x, vel y
down.hittime = 20
down.velocity = 0, 0
air.fall = 1
fall.recover = 1
fall.recovertime = 16
getpower = ifelse(p2movetype = H, 72, 144) * !(root, var(20)), 36 * !(root, var(20))
givepower = 36, 36
envshake.time = 14
envshake.freq = 120
envshake.ampl = 4
sprpriority = 4
airguard.ctrltime = 16
airguard.velocity = ifelse((root, var(20)), -4, -8), -1.5
forcestand = 1

[State 1055, Sparks]
type = explod
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
anim = 7020 + (random % 8)
ID = 7000
sprpriority = 5
postype = left
pos = ceil((target, pos x) -((target, const(size.ground.front) -10) * facing) + 160), ceil(screenpos y)
removetime = -2
ownpal = 1
facing = -1
ignorehitpause = 1
persistent = 0
pausemovetime = 100
supermovetime = 100

[State 1055, End]
type = changestate
trigger1 = movecontact
value = 1056
ignorehitpause = 1
[State 1055, End]
type = changestate
trigger1 = pos y >= 0
value = 1057
ignorehitpause = 1
[State 1055, End]
type = destroyself
trigger1 = pos x != [ -240, 240]

[Statedef 1056]
type = A
physics = N
movetype = I
anim = 3056
velset = 0, 0
ctrl = 0
sprpriority = 4
movehitpersist = 1
[State 1056, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1
[State 1056, Snd]
type = playsnd
trigger1 = !time
value = 5, 3
ignorehitpause = 1
persistent = 0
[State 1056, End]
type = destroyself
trigger1 = !animtime

[Statedef 1057]
type = A
physics = N
movetype = I
anim = 3057
velset = 0, 0
ctrl = 0
sprpriority = 4
[State 1057, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1
[State 1057, Pos]
type = posset
trigger1 = 1
y = 0
[State 1057, Snd]
type = playsnd
trigger1 = !time
value = 5, 3
ignorehitpause = 1
persistent = 0
[State 1057, Dust]
type = explod
trigger1 = !time
anim = 7104
ID = 7104
sprpriority = -2
pos = 0, 0
vel = 0, 0
removetime = -2
ownpal = 1
scale = 0.75, 0.75
pausemovetime = -1
supermovetime = -1
[State 1057, End]
type = destroyself
trigger1 = !animtime


;========================< MESSATSU GOU HADOU >=============================
[Statedef 3005]
type = S
physics = N
movetype = A
ctrl = 0
velset = 0, 0
sprpriority = 4

[State 3005, HitOverride]
type = hitoverride
trigger1 = 1
time = -1
attr = SCA, AA, AP, AT
stateno = ifelse(var(2) < 8, 3005, 3006)
ignorehitpause = 1
[State 3005, hitby]
type = hitby
trigger1 = 1
time = -1
value = SCA, NP, SP, HP
ignorehitpause = 1

[State 3005, Strength]
type = varset
trigger1 = !time && !prevstateno
var(10) = 5
[State 3005, Anim]
type = changeanim
trigger1 = !time && !prevstateno
value = 3005

[State 3005, hitvar]
type = varadd
trigger1 = !time
var(2) = 1
[State 3005, movehitvar]
type = varset
trigger1 = movecontact && numtarget
trigger1 = !(target, ishelper)
var(3) = 1
ignorehitpause = 1

[State 3005, Vel]
type = velset
trigger1 = 1
x = ifelse((time < 2 && var(2) > 1), 1, 9)
y = 0

[State 3005, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1

[State 3005, HitSound]
type = playsnd
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
value = 2, 7
channel = 1
persistent = 0
ignorehitpause = 1

[State 3005, Hit]
type = hitdef
trigger1 = 1
priority = 4, Hit
attr = S, HP
damage = ceil(25 * root, fvar(12)), 4
animtype = Back
hitflag = MAF
guardflag = M
pausetime = 4 * (var(2) < 8), 6
sparkno = -1
sparkxy = -10, 0
guard.sparkno = s7001
guardsound = s2, 0
ground.type = High
ground.slidetime = 16
ground.hittime = 16
ground.velocity = ifelse(var(2) < 8, -10.52, -6), ifelse(var(2) < 8, 0, -4)
ground.cornerpush.veloff = 0
air.velocity = -6, -4
guard.velocity = -10.52
fall = (var(2) >= 8)
air.fall = 1
fall.recover = 0
getpower = 0
givepower = 36, 36
envshake.time = 18
envshake.freq = 120
envshake.ampl = 6
palfx.sinadd = 64, 64, 64, 4
palfx.mul = 128, 128, 128
palfx.time = 32
sprpriority = 4
airguard.ctrltime = 16
airguard.velocity = ifelse((root, var(20)), -4, -8), -1.5
forcestand = 1

[State 3005, Sparks]
type = explod
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
anim = 7030 + (random % 8)
ID = 7000
sprpriority = 5
postype = left
pos = ceil((target, pos x) -((target, const(size.ground.front) -10) * facing) + 160), ceil(screenpos y)
removetime = -2
ownpal = 1
facing = -1
ignorehitpause = 1
persistent = 0
pausemovetime = 100
supermovetime = 100

[State 3005, flames]
type = helper
trigger1 = movehit && numtarget && numhelper(7181) < 1
trigger1 = !(target, time) && (target, movetype = H) && !var(16)
helpertype = normal
stateno = 7181
ID = 7181
name = "flames"
postype = p2
bindtime = 1000
ownpal = 1
facing = -1
size.ground.front = 0
size.ground.back = 0
ignorehitpause = 1

[State 3005, End]
type = changestate
trigger1 = movecontact
value = ifelse(var(2) < 8, 3005, 3006)
[State 3005, End]
type = destroyself
trigger1 = pos x != [ -280, 280]

[Statedef 3006]
type = A
physics = N
movetype = I
anim = 3006
ctrl = 0
velset = 0, 0
sprpriority = 4
[State 3006, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1
[State 3006, Snd]
type = playsnd
trigger1 = !time
value = 5, 1
ignorehitpause = 1
persistent = 0
[State 3006, End]
type = destroyself
trigger1 = !animtime


;========================< TENMA GOU ZANKUU >=============================
[Statedef 3055]
type = A
physics = N
movetype = A
ctrl = 0
velset = 0, 0
sprpriority = 4

[State 3055, HitOverride]
type = hitoverride
trigger1 = 1
time = -1
attr = SCA, AA, AP, AT
stateno = ifelse(var(2) < 8, 3055, 3056)
ignorehitpause = 1
[State 3055, hitby]
type = hitby
trigger1 = 1
time = -1
value = SCA, NP, SP, HP
ignorehitpause = 1

[State 3055, Strength]
type = varset
trigger1 = !time && !prevstateno
var(10) = 5
[State 3055, Anim]
type = changeanim
trigger1 = !time && !prevstateno
value = 3055

[State 3055, hitvar]
type = varadd
trigger1 = !time
var(2) = 1
[State 3055, movehitvar]
type = varset
trigger1 = movecontact && numtarget
trigger1 = !(target, ishelper)
var(3) = 1
ignorehitpause = 1

[State 3055, Vel]
type = velset
trigger1 = 1
x = ifelse((time < 2 && var(2) > 1), 1, 7.4)
y = ifelse((time < 2 && var(2) > 1), 1, 5.2)

[State 3055, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1

[State 3055, HitSound]
type = playsnd
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
value = 2, 7
channel = 1
persistent = 0
ignorehitpause = 1

[State 3055, Hit]
type = hitdef
trigger1 = 1
priority = 4, Hit
attr = A, HP
damage = ceil(25 * root, fvar(12)), 4
animtype = Back
hitflag = MAFD
guardflag = M
pausetime = 4 * (var(2) < 8), 6
sparkno = -1
sparkxy = -10, 0
guard.sparkno = s7001
guardsound = s2, 0
ground.type = Low
ground.slidetime = 16
ground.hittime = 16
ground.velocity = ifelse(var(2) < 8, -10.52, -4), ifelse(var(2) < 8, 0, -2)
ground.cornerpush.veloff = 0
air.velocity = -4, -2
down.velocity = 0, 0
down.hittime = 20
guard.velocity = -10.52
fall = (var(2) >= 8)
air.fall = 1
fall.recover = 0
getpower = 0
givepower = 36, 36
envshake.time = 18
envshake.freq = 120
envshake.ampl = 6
palfx.sinadd = 64, 64, 64, 4
palfx.mul = 128, 128, 128
palfx.time = 32
sprpriority = 4
airguard.ctrltime = 16
airguard.velocity = ifelse((root, var(20)), -4, -8), -1.5
forcestand = 1

[State 3055, Sparks]
type = explod
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
anim = 7030 + (random % 8)
ID = 7000
sprpriority = 5
postype = left
pos = ceil((target, pos x) -((target, const(size.ground.front) -10) * facing) + 160), ceil(screenpos y)
removetime = -2
ownpal = 1
facing = -1
ignorehitpause = 1
persistent = 0
pausemovetime = 100
supermovetime = 100

[State 3055, flames]
type = helper
trigger1 = movehit && numtarget && numhelper(7181) < 1 && (root, var(1) != 1)
trigger1 = !(target, time) && (target, movetype = H) && !var(16)
helpertype = normal
stateno = 7181
ID = 7181
name = "flames"
postype = p2
bindtime = 1000
ownpal = 1
facing = -1
size.ground.front = 0
size.ground.back = 0
ignorehitpause = 1

[State 3055, End]
type = changestate
trigger1 = movecontact
value = ifelse(var(2) < 8, 3055, 3056)
[State 3055, End]
type = changestate
trigger1 = pos y >= 0
value = 3057
ignorehitpause = 1
[State 3055, End]
type = destroyself
trigger1 = pos x != [ -280, 280]

[Statedef 3056]
type = A
physics = N
movetype = I
anim = 3056
ctrl = 0
velset = 0, 0
sprpriority = 4
[State 3056, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1
[State 3056, Snd]
type = playsnd
trigger1 = !time
value = 5, 3
ignorehitpause = 1
persistent = 0
[State 3056, End]
type = destroyself
trigger1 = !animtime

[Statedef 3057]
type = A
physics = N
movetype = I
anim = 3057
velset = 0, 0
ctrl = 0
sprpriority = 4
[State 3057, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1
[State 3057, Pos]
type = posset
trigger1 = 1
y = 0
[State 3057, Snd]
type = playsnd
trigger1 = animelem = 1
value = 5, 3
ignorehitpause = 1
persistent = 0
[State 3057, Dust]
type = explod
trigger1 = !time
anim = 7104
ID = 7104
sprpriority = -2
pos = 0, 0
vel = 0, 0
removetime = -2
ownpal = 1
pausemovetime = -1
supermovetime = -1
[State 3057, End]
type = destroyself
trigger1 = !animtime


;========================< ZANKUU HADOUKEN >=============================
[Statedef 3075]
type = A
physics = N
movetype = A
anim = 1055
ctrl = 0
velset = 0, 0
sprpriority = 4

[State 3075, HitOverride]
type = hitoverride
trigger1 = 1
time = -1
attr = SCA, AA, AP, AT
stateno = 1056
ignorehitpause = 1
[State 3075, hitby]
type = hitby
trigger1 = 1
time = -1
value = SCA, NP, SP, HP
ignorehitpause = 1

[State 3075, Strength]
type = varset
trigger1 = !time && !prevstateno
var(10) = 6
[State 3075, movehitvar]
type = varset
triggerall = !var(3)
trigger1 = movecontact && numtarget
trigger1 = !(target, ishelper)
trigger2 = numhelper(3075)
trigger2 = helper(3075), var(3)
var(3) = 1
ignorehitpause = 1

[State 3075, vel1]
type = velset
trigger1 = !time && (rootdist y > 48)
x = 7 + random / 333.3
y = 3.5 + random / 666.7
[State 3075, vel2]
type = velset
trigger1 = !time && (rootdist y <= 48)
x = 4 + random / 333.3
y = 5 + random / 666.7
;x = y * 1.425

[State 3075, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1

[State 3075, HitSound]
type = playsnd
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
value = 2, 7
channel = 1
persistent = 0
ignorehitpause = 1

[State 3075, Hit]
type = hitdef
trigger1 = 1
priority = 4, Hit
attr = A, HP
damage = ceil(23 * root, fvar(12)), 3
animtype = Hard
fall.animtype = Back
hitflag = MAFD
guardflag = M
pausetime = 0, 12
sparkno = -1
sparkxy = -10, 0
guard.sparkno = s7001
guardsound = s2, 0
ground.type = High
ground.slidetime = 16
ground.hittime = 16
ground.velocity = -10, 0
ground.cornerpush.veloff = 0
air.velocity = -1, -.5
down.hittime = 20
down.velocity = 0, 0
air.fall = 1
fall.recover = 0
getpower = 0
givepower = 36, 36
envshake.time = 18
envshake.freq = 120
envshake.ampl = 6
sprpriority = 4
airguard.ctrltime = 16
airguard.velocity = ifelse((root, var(20)), -4, -8), -1.5
forcestand = 1

[State 3075, Sparks]
type = explod
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
anim = 7030 + (random % 8)
ID = 7000
sprpriority = 5
postype = left
pos = ceil((target, pos x) -((target, const(size.ground.front) -10) * facing) + 160), ceil(screenpos y)
removetime = -2
ownpal = 1
facing = -1
ignorehitpause = 1
persistent = 0
pausemovetime = 100
supermovetime = 100

[State 3075, End]
type = changestate
trigger1 = movecontact
value = 1056
ignorehitpause = 1
[State 3075, End]
type = changestate
trigger1 = pos y >= 0
value = 1057
ignorehitpause = 1
[State 3075, End]
type = destroyself
trigger1 = pos x != [ -240, 240]


;========================< KONGOU KOKURETSU ZAN >=============================
;--- LIGHTNING ---
[Statedef 4205]
type = S
physics = N
movetype = A
ctrl = 0
velset = 0, 0
anim = 4205
sprpriority = 4

[State 4205, No Reversals]
type = hitoverride
trigger1 = 1
time = -1
attr = SCA, AT
stateno = 5000
ignorehitpause = 1

[State 4205, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1

[State 4205, Strength]
type = varset
trigger1 = !time && !prevstateno
var(10) = 6

[State 4205, movehitvar]
type = varset
triggerall = !var(3)
trigger1 = movecontact && numtarget
trigger1 = !(target, ishelper)
trigger2 = numhelper(4205)
trigger2 = helper(4205), var(3)
var(3) = 1
ignorehitpause = 1

[State 4205, Shadows]
type = afterimage
trigger1 = !time && !prevstateno
time = 999
timegap = 1
framegap = 2
length = 3
palcontrast = 20, 10, 50
trans = add

[State 4205, Hit]
type = hitdef
trigger1 = animelem = 1
attr = S, HA
damage = ceil(40 * root, fvar(12)), 5
animtype = Up
guardflag = M
hitflag = MAFD
priority = 4, Hit
pausetime = 0, 12
sparkno = -1
guard.sparkno = -1
hitsound = s2, 7
guardsound = s2, 0
ground.type = High
ground.slidetime = 16
ground.hittime = 16
ground.velocity = ifelse(frontedgedist < backedgedist, 6, -6), -3
ground.cornerpush.veloff = 0
air.velocity = ifelse(frontedgedist < backedgedist, 6, -6), -3
down.velocity = ifelse(frontedgedist < backedgedist, 6, -6), -3
getpower = 0
givepower = 36, 36
guard.velocity = -10.52
fall = 1
fall.recover = 0
palfx.sinadd = 64, 64, 64, 4
palfx.mul = 128, 128, 128
palfx.time = 32
sprpriority = 4
forcestand = 1
airguard.ctrltime = 16
airguard.cornerpush.veloff = 0
airguard.velocity = -8, -1.5

[State 4205, flames]
type = helper
triggerall = movehit && numtarget && numhelper(7181) < 1
trigger1 = !(target, time) && (target, movetype = H)
helpertype = normal
stateno = 7181
ID = 7181
name = "flames"
postype = p2
bindtime = 1000
ownpal = 1
facing = -1
size.ground.front = 0
size.ground.back = 0
ignorehitpause = 1
persistent = 0

[State 4205, Sparks]
type = explod
trigger1 = movecontact = 1 && numtarget
trigger1 = (target, movetype = H) && !(target, time)
anim = ifelse(moveguarded, 7001, 7030 + (random % 8))
ID = 7000
sprpriority = 5
postype = p2
pos = (target, const(size.ground.front) -10), (target, const(size.mid.pos.y))
removetime = -2
ownpal = 1
facing = -1
ignorehitpause = 1
pausemovetime = 100
supermovetime = 100

[State 4205, End]
type = destroyself
trigger1 = !animtime


;--- SHOCKWAVES ---
[Statedef 4206]
type = S
physics = N
movetype = A
ctrl = 0
velset = 0, 0
sprpriority = 5

[State 4206, HitOverride]
type = hitoverride
trigger1 = 1
time = -1
attr = SCA, AA, AP, AT
stateno = 4206
ignorehitpause = 1
[State 4206, hitby]
type = hitby
trigger1 = 1
time = -1
value = SCA, NP, SP, HP
ignorehitpause = 1

[State 4206, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1

[State 4206, Strength]
type = varset
trigger1 = !time && !prevstateno
var(10) = 6
[State 4206, Anim]
type = changeanim
trigger1 = !time && !prevstateno
value = 4206

[State 4206, movehitvar]
type = varset
trigger1 = movecontact && numtarget
trigger1 = !(target, ishelper)
var(3) = 1
ignorehitpause = 1
[State 4206, movehitvar2]
type = varset
trigger1 = !time && prevstateno
var(4) = animelemno(0)
ignorehitpause = 1

[State 4206, Shadows]
type = afterimage
trigger1 = !time && !prevstateno
time = 999
timegap = 1
framegap = 2
length = 7
palcontrast = 20, 20, 20
trans = add

[State 4206, HitSound]
type = playsnd
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
value = 2, 7
channel = 1
ignorehitpause = 1

[State 4206, Hit]
type = hitdef
trigger1 = animelemno(0) != var(4)
trigger1 = animelem = 2 || animelem = 3 || animelem = 4 || animelem = 5 || animelem = 6 || animelem = 7 || animelem = 8 || animelem = 9
attr = S, HA
damage = ceil(40 * root, fvar(12)), 5
animtype = Hard
fall.animtype = Up
guardflag = M
hitflag = MAFD
priority = 4, Hit
pausetime = 0, 1
sparkno = -1
guard.sparkno = -1
guardsound = s2, 0
ground.type = High
ground.slidetime = 16
ground.hittime = 16
ground.velocity = ifelse((animelem = 9), -12, -4), 0
ground.cornerpush.veloff = 0
air.velocity = -3, -12
getpower = 0
givepower = 36, 36
fall = 0
air.fall = 1
fall.recover = 0
palfx.sinadd = 64, 64, 64, 4
palfx.mul = 128, 128, 128
palfx.time = ifelse(p2statetype = A, 74, 30)
ID = 4206 + (facing = -1)
nochainID = 4207 -(facing = -1)
sprpriority = 4
forcestand = 1
airguard.ctrltime = 16
airguard.cornerpush.veloff = 0
airguard.velocity = -8, -1.5

[State 4206, flames]
type = helper
trigger1 = movehit && numtarget && numhelper(7181) < 1
trigger1 = !(target, time) && (target, movetype = H)
helpertype = normal
stateno = 7181
ID = 7181
name = "flames"
postype = p2
bindtime = 1000
ownpal = 1
facing = -1
size.ground.front = 0
size.ground.back = 0
ignorehitpause = 1
persistent = 0

[State 4206, Sparks]
type = explod
trigger1 = movecontact = 1 && numtarget
trigger1 = (target, movetype = H) && !(target, time)
anim = ifelse(moveguarded, 7001, 7030 + (random % 8))
ID = 7000
sprpriority = 5
postype = p2
pos = (target, const(size.ground.front) -10), (target, const(size.mid.pos.y))
removetime = -2
ownpal = 1
facing = -1
ignorehitpause = 1
pausemovetime = 100
supermovetime = 100

[State 4206, End]
type = destroyself
trigger1 = !animtime


;========================< TENSHOU KAIREKI JIN >=============================
[Statedef 4305]
type = S
physics = N
movetype = A
ctrl = 0
velset = 0, 0
sprpriority = 5

[State 4305, HitOverride]
type = hitoverride
trigger1 = 1
time = -1
attr = SCA, AA, AP, AT
stateno = 4305
ignorehitpause = 1
[State 4305, hitby]
type = hitby
trigger1 = 1
time = -1
value = SCA, NP, SP, HP
ignorehitpause = 1

[State 4305, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1
[State 4305, Pos]
type = posset
trigger1 = 1
y = 0

[State 4305, Strength]
type = varset
trigger1 = !time && !prevstateno
var(10) = 6
[State 4305, Anim]
type = changeanim
trigger1 = !time && !prevstateno
value = 4305

[State 4305, movehitvar]
type = varset
trigger1 = movecontact && numtarget
trigger1 = !(target, ishelper)
var(3) = 1
ignorehitpause = 1
[State 4305, movehitvar2]
type = varset
trigger1 = !time && prevstateno
var(4) = animelemno(0)
ignorehitpause = 1

[State 4305, HitSound]
type = playsnd
trigger1 = movehit && numtarget
trigger1 = !(target, time) && target, movetype = H
value = 2, 11
channel = 1
ignorehitpause = 1

[State 4305, Hit]
type = hitdef
trigger1 = animelemno(0) != var(4) && animelem = 2
attr = S, HA 
damage = ceil(51 * root, fvar(12)), 5
animtype = Hard
fall.animtype = Up
guardflag = M
hitflag = MAFD 
priority = 4, Hit 
pausetime = 0, 0
sparkno = -1
guard.sparkno = -1
guardsound = s2, 0
ground.type = High
ground.slidetime = 16
ground.hittime = 16
ground.velocity = ifelse(p2dist x >= 30, 1.5, ifelse(p2dist x <= -30, -1.5, ifelse((animelem = 12), -3, 0))), ifelse((animelem = 12), -15, -10)
ground.cornerpush.veloff = 0
air.velocity = ifelse(p2dist x >= 30, 1.5, ifelse(p2dist x <= -30, -1.5, ifelse((animelem = 12), -3, 0))), ifelse((animelem = 12), -15, -10)
getpower = 0
givepower = 36, 36
guard.velocity = -10.52
fall = 1
fall.recover = 0
palfx.sinadd = 64, 64, 64, 4
palfx.mul = 128, 128, 128
palfx.time = 100
sprpriority = 4
forcestand = 1
airguard.ctrltime = 16
airguard.cornerpush.veloff = 0
airguard.velocity = -8, -1.5

[State 4305, Hit]
type = hitdef
triggerall = animelemno(0) != var(4)
trigger1 = animelem = 4 || animelem = 6 || animelem = 8 || animelem = 10 || animelem = 12
attr = S, HA 
damage = ceil(51 * root, fvar(12)), 5
animtype = Hard
fall.animtype = Up
guardflag = M
hitflag = MAFD 
priority = 4, Hit 
pausetime = 0, 0
sparkno = -1
guard.sparkno = -1
guardsound = s2, 0
ground.type = High
ground.slidetime = 16
ground.hittime = 16
ground.velocity = ifelse((animelem = 12), -3, 0), ifelse((animelem = 12), -15, -10)
ground.cornerpush.veloff = 0
air.velocity = ifelse((animelem = 12), -3, 0), ifelse((animelem = 12), -15, -10)
getpower = 0
givepower = 36, 36
guard.velocity = -10.52
fall = 1
fall.recover = 0
palfx.sinadd = 64, 64, 64, 4
palfx.mul = 128, 128, 128
palfx.time = 100
sprpriority = 4
airguard.ctrltime = 16
airguard.velocity = ifelse((root, var(20)), -4, -8), -1.5
forcestand = 1

[State 4305, flames]
type = helper
triggerall = movehit && numtarget && numhelper(7181) < 1
trigger1 = !(target, time) && (target, movetype = H)
helpertype = normal
stateno = 7181
ID = 7181
name = "flames"
postype = p2
bindtime = 1000
ownpal = 1
facing = -1
size.ground.front = 0
size.ground.back = 0
ignorehitpause = 1
persistent = 0

[State 4305, Sparks]
type = explod
trigger1 = movecontact = 1 && numtarget
trigger1 = (target, movetype = H) && !(target, time)
anim = ifelse(moveguarded, 7001, 7030 + (random % 8))
ID = 7000
sprpriority = 5
postype = p2
pos = (target, const(size.ground.front) -10), (target, const(size.mid.pos.y))
removetime = -2
ownpal = 1
facing = -1
ignorehitpause = 1
pausemovetime = 100
supermovetime = 100

[State 4305, End]
type = destroyself
trigger1 = !animtime


;========================< PURPLE DUST >=============================
[Statedef 4210]
physics = N
anim = 9741
[State 4210, Assert]
type = assertspecial
trigger1 = 1
flag = noshadow
ignorehitpause = 1
[State 4210, Dust]
type = explod
trigger1 = !time
anim = 4210
ID = 4210
sprpriority = -2
postype = p1
removetime = 25
[State 4210, Fade]
type = palfx
trigger1 = !time
mul = 384, 384, 384
sinadd = -255, -255, -255, 100
time = 25
[State 4210, Shrink]
type = modifyexplod
trigger1 = 1
ID = 4210
scale = 0.5 + (0.1 * time), 0.1 + (0.02 * time)
[State 4210, End]
type = destroyself
trigger1 = time >= 26


;========================< TENMA SHINZUI WARI F.X. >=============================
[Statedef 3305]
physics = N
velset = 0, 0
anim = 3305
sprpriority = 3
[State 3305, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1
[State 3305, Bind]
type = bindtoroot
trigger1 = root, anim = 3300
ignorehitpause = 1

[State 3305, Shadows]
type = afterimage
trigger1 = !time
time = 999
timegap = 1
framegap = 2
length = 3
palcontrast = 120, 90, 150
trans = add

[State 3305, End]
type = destroyself
trigger1 = !animtime || root, stateno != 3300


;========================< MISOGI ILLUSIONS >=============================
[Statedef 4104]
physics = N
anim = 9741
[State 4111, Assert]
type = assertspecial
trigger1 = 1
flag = invisible

[State 4104, Dust]
type = explod
trigger1 = !time
anim = 4110
ID = 4110
sprpriority = -1
postype = p1
pos = 0, 0
vel = 0, 0
removetime = -2
ownpal = 1
[State 4104, particles]
type = explod
trigger1 = time = [1, 16]
vel = 0, -ceil(random / 200)
accel = 0, -.3
anim = 4111
sprpriority = -1
postype = p1
pos = -2, -45
random = 50, 50
facing = -1 + (random < 500) * 2
removetime = -2
ownpal = 1
scale = 0.25, 0.25
pausemovetime = 300
supermovetime = 300

[State 4104, Fade]
type = palfx
trigger1 = !time
mul = 200, 0, 256
sinadd = -255, -255, -255, 240
time = 60

[State 4104, End]
type = destroyself
trigger1 = time >= 60


;========================< S.G.S. GUARDED >=============================
[Statedef 4012]
physics = N
velset = 0, 0
anim = 9741

[State 4012, BG PalFX]
type = bgpalfx
trigger1 = !time
time = 20
add = -255, -128, -64
sinadd = 255, 128, 64, 80
persistent = 0
[State 4012, palfx]
type = palfx
trigger1 = !time
time = 24
sinadd = -255, -255, -255, 96

[State 4012, thunder]
type = explod
trigger1 = !time
anim = 4012
ID = 4012
sprpriority = -2
postype = p1
pos = 0, -ceil(pos y)
removetime = -2
supermovetime = 30
ownpal = 1
[State 4012, Circle]
type = explod
trigger1 = !time
anim = 7501
ID = 4013
sprpriority = 5
postype = p1
removetime = 20
supermovetime = 30

[State 4012, Shrink]
type = modifyexplod
trigger1 = numexplod(4013)
ID = 4013
scale = 0.2 + (0.09 * time), 0.2 + (0.09 * time)
[State 4012, No Shadow]
type = assertspecial
trigger1 = 1
flag = invisible
ignorehitpause = 1
[State 4012, End]
type = destroyself
trigger1 = time >= 24


;========================< AURA >=============================
[Statedef 6000]
physics = N
velset = 0, 0
anim = 6000
sprpriority = -2
[State 6000, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1
[State 6000, Var]
type = varset
trigger1 = !time
var(0) = root, anim
[State 6000, Shadows]
type = afterimage
trigger1 = !time
time = 999
timegap = 1
framegap = 1
length = 2
palcontrast = 100, 100, 100
trans = add
[State 6000, Bind]
type = bindtoroot
trigger1 = 1
[State 6000, End]
type = destroyself
trigger1 = (root, anim) != var(0)
trigger2 = root, stateno = 195 && root, animelemtime(19) >= 0
trigger3 = root, stateno = 182 && root, animelemtime(26) >= 0
trigger4 = root, stateno = 1020 && root, animelemtime(9) >= 0
trigger5 = root, stateno = 1400 && root, animelemtime(5) >= 0
trigger6 = root, stateno = 3020 && root, anim = 3020 && root, animelemtime(22) >= 0
trigger7 = root, stateno = 193 && root, anim = 193 && root, animelemtime(39) >= 0


;========================< METSU >=============================
[Statedef 6002]
physics = N
velset = 0, -.5
anim = 6002
sprpriority = -1
[State 6002, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1
[State 6002, Shadows]
type = afterimage
trigger1 = !time
time = 999
timegap = 1
framegap = 1
length = 2
palcontrast = 120, 90, 150
trans = add
[State 6002, Shadows]
type = velset
trigger1 = animelem = 6
y = -2
[State 6002, Fade]
type = palfx
trigger1 = animelem = 7
sinadd = -255, -255, -255, 64
time = 16

[State 6002, Bind]
type = posset
trigger1 = root, stateno = 195 && root, movetype != H
x = ceil(root, pos x) + ifelse(root, facing = -1, 5, -5)

[State 6002, End]
type = destroyself
trigger1 = !animtime || root, movetype = H


;========================< KKZ AURA >=============================
[Statedef 4203]
physics = N
velset = 0, 0
anim = 4203
sprpriority = 0
[State 4203, Trans]
type = trans
trigger1 = 1
trans = addalpha
alpha = 256, 256
ignorehitpause = 1
[State 4203, Shadows]
type = afterimage
trigger1 = !time
time = 999
timegap = 1
framegap = 1
length = 2
palcontrast = 100, 100, 100
trans = add
[State 4203, Bind]
type = bindtoroot
trigger1 = 1
time = -1
[State 4203, End]
type = destroyself
trigger1 = root, stateno != 4200 || root, anim != 4200
trigger2 = root, stateno = 4200 && root, animelemtime(19) >= 0
trigger3 = !animtime


;========================< RED PARTICLES >=============================
[Statedef 6021]
physics = N
anim = 6021
sprpriority = 4
[State 6021, Assert]
type = assertspecial
trigger1 = 1
flag = noshadow
[State 6021, Assert]
type = assertspecial
trigger1 = !time
flag = invisible
[State 6021, Fade]
type = palfx
trigger1 = !time
add = -255, -255, -255
sinadd = 255, 255, 255, 40
time = 10
[State 6021, Vel]
type = velset
trigger1 = !time
x = (rootdist x + 16) / 36.0
y = (rootdist y -79) / 36.0
[State 6021, End]
type = destroyself
trigger1 = !animtime


;========================< VEGA >=============================
[Statedef 6100]
physics = N
velset = 0, 0
anim = 6100
sprpriority = 4
[State 6100, exists]
type = parentvarset
trigger1 = playerIDexist(parent, ID)
var(51) = 1
[State 6100, offset]
type = offset
trigger1 = 1
y = 6
[State 6100, Snd]
type = playsnd
trigger1 = time = 5
value = 6, random < 500
channel = 0
[State 6100, blood]
type = explod
trigger1 = animelem = 9
anim = 6105
ID = 6105
sprpriority = 5
postype = p1
bindtime = -1
[State 6100, Snd]
type = playsnd
trigger1 = animelem = 9
value = 2, 8
[State 6100, Snd]
type = playsnd
trigger1 = animelem = 12
value = 6, 11
channel = 0
[State 6100, gone]
type = destroyself
triggerall = numenemy
trigger1 = enemy, name = "Shin Gouki" && enemy, authorname = "Phantom.of.the.Server" && enemy, numhelper(6100)
trigger2 = enemy, name = "Vega" && enemy, authorname = "Phantom.of.the.Server"
trigger3 = enemy, name = "cvsvega" && (enemy, authorname = "Warusaki3" || enemy, authorname != "Gargoyle")
trigger4 = enemy, name = "M. Bison" && enemy, authorname = "VK, Updated by Terry Kuo, adapted by Tenebrous"
trigger5 = enemy, name = "M.Bison" && enemy, authorname = "N -Mario"
trigger6 = enemy, name = "CvSBison" && enemy, authorname = "Team Kai (Saiai and Koga)"
trigger7 = enemy, name = "[Necromancer's] M. Bison" || enemy, name = "Sevil Nahte"


;--- DEAD ---
[Statedef 6101]
physics = N
velset = 0, 0
anim = 6101
sprpriority = 4
[State 6101, offset]
type = offset
trigger1 = 1
y = 6
[State 6101, Snd]
type = stopsnd
trigger1 = !time
channel = 0
[State 6101, exists]
type = parentvarset
trigger1 = playerIDexist(parent, ID)
var(51) = 1
[State 6101, gone]
type = destroyself
triggerall = numenemy
trigger1 = enemy, name = "Shin Gouki" && enemy, authorname = "Phantom.of.the.Server" && enemy, numhelper(6100)
trigger2 = enemy, name = "Vega" && enemy, authorname = "Phantom.of.the.Server"
trigger3 = enemy, name = "cvsvega" && (enemy, authorname = "Warusaki3" || enemy, authorname != "Gargoyle")
trigger4 = enemy, name = "M. Bison" && enemy, authorname = "VK, Updated by Terry Kuo, adapted by Tenebrous"
trigger5 = enemy, name = "M.Bison" && enemy, authorname = "N -Mario"
trigger6 = enemy, name = "CvSBison" && enemy, authorname = "Team Kai (Saiai and Koga)"
trigger7 = enemy, name = "[Necromancer's] M. Bison" || enemy, name = "Sevil Nahte"


;========================< JUMP DUST >=============================
[Statedef 7100]
physics = N
velset = 0, 0
anim = 9741
[State 7100, Bind]
type = bindtoroot
trigger1 = time
[State 7100, Dust]
type = explod
trigger1 = !time
anim = 7100
ID = 7100
sprpriority = 3
postype = p1
scale = 0.5, 0.5
pausemovetime = -1
supermovetime = -1
[State 7100, Dust]
type = explod
trigger1 = !time || time = 2 || time = 6
anim = 7101
ID = 7100
sprpriority = 3
postype = p1
scale = 0.5, 0.5
pausemovetime = -1
supermovetime = -1
[State 7100, Assert]
type = assertspecial
trigger1 = 1
flag = invisible
[State 7100, End]
type = destroyself
trigger1 = time >= 7


;========================< FIRE >=============================
;--- RED ---
[Statedef 7180]
physics = N
velset = 0, 0
anim = 9741
sprpriority = 4
[State 7180, Assert]
type = assertspecial
trigger1 = 1
flag = noshadow
ignorehitpause = 1
[State 7180, flames]
type = explod
trigger1 = 1
anim = 7180
ID = 7180
sprpriority = 3
postype = p2
vel = -.2, -.75
pos = (enemynear, const(size.mid.pos.x)), ifelse((enemynear, statetype = L), 0, (enemynear, const(size.mid.pos.y) + 10))
random = 40, 40
removetime = -2
scale = 0.5, 0.5
ownpal = 1
[State 7180, Pos]
type = posadd
trigger1 = 1
x = ceil(p2dist x)
y = ceil(p2dist y)
[State End]
type = destroyself
triggerall = numenemy
trigger1 = enemynear, ctrl
trigger2 = enemynear, movetype != H || enemynear, stateno = 5300
trigger3 = enemynear, statetype = L && (enemynear, stateno != [5080, 5081])
[State End]
type = destroyself
trigger1 = !numenemy || time >= 120


;--- PURPLE ---
[Statedef 7181]
physics = N
velset = 0, 0
anim = 9741
sprpriority = 4
[State 7181, Assert]
type = assertspecial
trigger1 = 1
flag = noshadow
ignorehitpause = 1
[State 7181, flames]
type = explod
trigger1 = 1
anim = 7181
ID = 7181
sprpriority = 3
postype = p2
vel = -.2, -.75
pos = (enemynear, const(size.mid.pos.x)), ifelse((enemynear, statetype = L), 0, (enemynear, const(size.mid.pos.y) + 10))
random = 40, 40
removetime = -2
scale = 0.5, 0.5
ownpal = 1
[State 7181, Pos]
type = posadd
trigger1 = 1
x = ceil(p2dist x)
y = ceil(p2dist y)
[State End]
type = destroyself
triggerall = numenemy
trigger1 = enemynear, ctrl
trigger2 = enemynear, movetype != H || enemynear, stateno = 5300
trigger3 = enemynear, statetype = L && (enemynear, stateno != [5080, 5081])
[State End]
type = destroyself
trigger1 = !numenemy || time >= 120


;========================< SUPER CHARGE >=============================
[Statedef 7500]
physics = N
velset = 0, 0
anim = 9741

[State 7500, Lv1]
type = bgpalfx
trigger1 = time <= 10 && root, var(10) = 5
time = 1
add = 256 -ceil(25.6 * time), 256 -ceil(25.6 * time), 256 -ceil(25.6 * time)
mul = 256 -floor(25.6 * time), 256 -floor(25.6 * time), 256
color = 256 -floor(25.6 * time)
[State 7500, Lv1]
type = bgpalfx
trigger1 = (time = [11, 39]) && root, var(10) = 5
time = 1
mul = 0, 0, 256
color = 0
[State 7500, Lv1]
type = bgpalfx
trigger1 = time >= 40 && root, var(10) = 5
time = 1
mul = ceil(51.2 * (time -40)), ceil(51.2 * (time -40)), 256
color = floor(51.2 * (time -40))

[State 7500, Lv2]
type = bgpalfx
trigger1 = time <= 15 && root, var(10) = 6
time = 1
add = 256 -ceil(17.07 * time), 256 -ceil(17.07 * time), 256 -ceil(17.07 * time)
mul = 256 -floor(17.07 * time), 256 -floor(17.07 * time), 256
color = 256 -floor(17.07 * time)
[State 7500, Lv2]
type = bgpalfx
trigger1 = (time = [16, 47]) && root, var(10) = 6
time = 1
mul = 0, 0, 256
color = 0
[State 7500, Lv2]
type = bgpalfx
trigger1 = time >= 48 && root, var(10) = 6
time = 1
mul = ceil(51.2 * (time -48)), ceil(51.2 * (time -48)), 256
color = floor(51.2 * (time -48))

[State 7500, Lv3]
type = bgpalfx
trigger1 = time <= 20 && root, var(10) = 7
time = 1
add = 256 -ceil(12.8 * time), 256 -ceil(12.8 * time), 256 -ceil(12.8 * time)
mul = 256 -floor(12.8 * time), 256 -floor(12.8 * time), 256 -floor(12.8 * time)
color = 256 -floor(12.8 * time)
[State 7500, Lv3]
type = bgpalfx
trigger1 = (time = [21, 54]) && root, var(10) = 7
time = 1
mul = 0, 0, 0
color = 0
[State 7500, Lv3]
type = bgpalfx
trigger1 = time >= 55 && root, var(10) = 7
time = 1
mul = ceil(51.2 * (time -55)), ceil(51.2 * (time -55)), ceil(51.2 * (time -55))
color = floor(51.2 * (time -55))

[State 7500, Spiral]
type = helper
trigger1 = !time && root, var(10) >= 6
helpertype = normal
stateno = 7509
ID = 7509
name = "Super Charge Spiral"
postype = p1
pausemovetime = 255
supermovetime = 255
ownpal = 1

[State 7500, Main]
type = explod
trigger1 = !time
anim = 7500
ID = 7500
sprpriority = 5
postype = p1
removetime = -1
pausemovetime = -1
supermovetime = -1
[State 7500, Large Circle]
type = explod
trigger1 = !time
anim = 7501
ID = 7501
sprpriority = 5
postype = p1
pausemovetime = -1
supermovetime = -1

[State 7500, Portrait]
type = explod
trigger1 = time = 1 && root, var(10) = 7
anim = 7510
ID = 7510
postype = back
pos = 160, floor((1.3333333 * gameheight / gamewidth) * 120)
sprpriority = -16
bindtime = -1
pausemovetime = -1
supermovetime = -1

[State 7500, Circle]
type = explod
trigger1 = !time
anim = 7502
ID = 75021
sprpriority = 5
postype = p1
pausemovetime = -1
supermovetime = -1
[State 7500, Circle]
type = explod
trigger1 = time = 7
anim = 7502
ID = 75022
sprpriority = 5
postype = p1
pausemovetime = -1
supermovetime = -1
[State 7500, Circle]
type = explod
trigger1 = time = 14
anim = 7502
ID = 75023
sprpriority = 5
postype = p1
pausemovetime = -1
supermovetime = -1
[State 7500, Circle]
type = explod
trigger1 = time = 21
anim = 7502
ID = 75024
sprpriority = 5
postype = p1
pausemovetime = -1
supermovetime = -1
[State 7500, Circle]
type = explod
trigger1 = time = 28
anim = 7502
ID = 75025
sprpriority = 5
postype = p1
pausemovetime = -1
supermovetime = -1

[State 7500, Scale]
type = modifyexplod
trigger1 = root, var(10) = 5
ID = 7500
scale = 1.5 -(0.0333 * time), 1.5 -(0.0333 * time)
[State 7500, Scale]
type = modifyexplod
trigger1 = root, var(10) = 6
ID = 7500
scale = 1.5 -(0.0283 * time), 1.5 -(0.0283 * time)
[State 7500, Scale]
type = modifyexplod
trigger1 = root, var(10) = 7
ID = 7500
scale = 1.5 -(0.025 * time), 1.5 -(0.025 * time)

[State 7500, Scale]
type = modifyexplod
trigger1 = numexplod(7501)
ID = 7501
scale = 0.25 + (0.075 * time), 0.25 + (0.075 * time)
trans = addalpha
alpha = ceil(256 -6.4 * time), 256

[State 7500, Scale]
type = modifyexplod
trigger1 = numexplod(7510) && time < 8
ID = 7510
postype = back
pos = 16, floor((1.3333333 * gameheight / gamewidth) * 120)
scale = 0.5, 0.5
trans = addalpha
alpha = (32 * time), 256

[State 7500, Scale]
type = modifyexplod
trigger1 = numexplod(7510) && (time = [8, 50])
ID = 7510
postype = back
pos = 160, floor((1.3333333 * gameheight / gamewidth) * 120)
scale = 0.5, 0.5
trans = addalpha
alpha = 256, 256

[State 7500, Scale]
type = modifyexplod
trigger1 = numexplod(7510) && time >= 50
ID = 7510
postype = back
pos = 160, floor((1.3333333 * gameheight / gamewidth) * 120) + (8 * (time -50))
scale = 0.5 + (0.25 * (time -50)), 0.5 + (0.25 * (time -50))
trans = addalpha
alpha = 256 -ceil(25.6 * (time -50)), 256

[State 7500, Scale]
type = modifyexplod
trigger1 = numexplod(75021)
ID = 75021
scale = 0.25 + (0.0375 * time), 0.25 + (0.0375 * time)
[State 7500, Scale]
type = modifyexplod
trigger1 = numexplod(75022)
ID = 75022
scale = 0.25 + (0.0375 * (time -7)), 0.25 + (0.0375 * (time -7))
[State 7500, Scale]
type = modifyexplod
trigger1 = numexplod(75023)
ID = 75023
scale = 0.25 + (0.0375 * (time -14)), 0.25 + (0.0375 * (time -14))
[State 7500, Scale]
type = modifyexplod
trigger1 = numexplod(75024)
ID = 75024
scale = 0.25 + (0.0375 * (time -21)), 0.25 + (0.0375 * (time -21))
[State 7500, Scale]
type = modifyexplod
trigger1 = numexplod(75025)
ID = 75025
scale = 0.25 + (0.0375 * (time -28)), 0.25 + (0.0375 * (time -28))

[State 7500, Safety]
type = assertspecial
trigger1 = 1
flag = noshadow
flag2 = invisible
[State 7500, No bars]
type = assertspecial
trigger1 = root, var(10) = 7
flag = nobardisplay
flag2 = noFG

[State 7500, Safety]
type = removeexplod
trigger1 = root, var(10) = 5 && time >= 45
trigger2 = root, var(10) = 6 && time >= 53
trigger3 = time >= 60
[State 7500, End]
type = destroyself
trigger1 = root, var(10) = 5 && time >= 45
trigger2 = root, var(10) = 6 && time >= 53
trigger3 = time >= 60


;--- Spiral ---
[Statedef 7509]
physics = N
anim = 7509
sprpriority = 5
[State 7509, Assert]
type = assertspecial
trigger1 = 1
flag = noshadow

[State 7509, Spin]
type = angleset
trigger1 = !time
value = -15 * (random % 24)
[State 7509, Spin]
type = angleadd
trigger1 = 1
value = -15

[State 7509, Angle]
type = angledraw
trigger1 = root, var(10) = 6
scale = 1 -(time / 53.0), 1 -(time / 53.0)
[State 7509, Trans]
type = trans
trigger1 = root, var(10) = 6
trans = addalpha
alpha = floor(4.83 * time), 256

[State 7509, Angle]
type = angledraw
trigger1 = root, var(10) = 7
scale = 1 -(time / 60.0), 1 -(time / 60.0)
[State 7509, Trans]
type = trans
trigger1 = root, var(10) = 7
trans = addalpha
alpha = floor(4.267 * time), 256

[State 7509, End]
type = destroyself
trigger1 = root, var(10) = 6 && time >= 54
trigger2 = time >= 61


;========================< POWER CHARGE >=============================
;--- MAIN FX ---
[Statedef 7240]
type = S
anim = 7240
velset = 0, 0
ctrl = 0

[State 7240, Bind]
type = bindtoroot
trigger1 = (root, stateno = 740) && (root, anim = 740)
pos = -4, 0
[State 7240, Snd]
type = playsnd
trigger1 = !time
value = 7, 4
channel = 0

[State 7240, Aura]
type = explod
trigger1 = !numexplod(7242)
ID = 7242
anim = 7242
sprpriority = 3
postype = p1
bindtime = -1
removetime = -1
pausemovetime = -1
supermovetime = -1
scale = 0.5, 0.5
ownpal = 1

[State 7240, FX]
type = modifyexplod
trigger1 = numexplod(7242) && (root, power < 3000)
ID = 7242
scale = 0.5, 0.1 + .4 * (root, power / 3000.0)
[State 7240, FX]
type = modifyexplod
trigger1 = numexplod(7242) && (root, power >= 3000)
ID = 7242
scale = 0.5, 0.5

[State 7240, No Shadow]
type = assertspecial
trigger1 = 1
flag = noshadow

[State 7240, End]
type = stopsnd
trigger1 = (root, stateno != 740) || (root, anim != 740) || (root, movetype = H)
channel = 0
[State 7240, End]
type = removeexplod
trigger1 = (root, stateno != 740) || (root, anim != 740) || (root, movetype = H)
[State 7240, End]
type = destroyself
trigger1 = (root, stateno != 740) || (root, anim != 740) || (root, movetype = H)


;--- PARTICLES ---
[Statedef 7243]
type = S
anim = 9741
velset = 0, 0
ctrl = 0

[State 7243, Bind]
type = bindtoroot
trigger1 = !(time % 4)
pos = ceil(-84 + (random / 6.25)), ceil(24 -(random / 6.25))

[State 7243, Particle]
type = explod
trigger1 = (time % 4) = 1
anim = 7243
sprpriority = ifelse(random < 750, 5, -1)
postype = p1
vel = ((rootdist x -4) / 16.0), ((rootdist y -56) / 16.0)
scale = ifelse(random < 500, 0.5, -.5), ifelse(random < 500, 0.5, -.5)
removetime = 16
pausemovetime = -1
supermovetime = -1

[State 7243, End]
type = removeexplod
trigger1 = (root, stateno != 740) || (root, anim != 740) || (root, movetype = H)
[State 7243, End]
type = destroyself
trigger1 = (root, stateno != 740) || (root, anim != 740) || (root, movetype = H)


;========================< ZERO COUNTER CHARGE >=============================
[Statedef 7250]
physics = N
velset = 0, 0
anim = 9741

[State 7250, BG PalFX]
type = bgpalfx
trigger1 = !time
time = 20
add = -255, -128, -64
sinadd = 255, 128, 64, 80
persistent = 0
[State 7250, palfx]
type = palfx
trigger1 = !time
time = 20
sinadd = -255, -255, -255, 80

[State 7250, thunder]
type = explod
trigger1 = !time
anim = 7250
ID = 7250
sprpriority = -2
postype = p1
pos = 0, -ceil(pos y)
removetime = -2
supermovetime = 30
ownpal = 1
[State 7250, Circle]
type = explod
trigger1 = !time
anim = 7251
ID = 7251
sprpriority = 5
postype = p1
removetime = 20
supermovetime = 30

[State 7250, Shrink]
type = modifyexplod
trigger1 = numexplod(7251)
ID = 7251
scale = 0.2 + (0.09 * time), 0.2 + (0.09 * time)
[State 7250, No Shadow]
type = assertspecial
trigger1 = 1
flag = noshadow
flag2 = invisible
ignorehitpause = 1
[State 7250, End]
type = destroyself
trigger1 = time >= 20


;========================< ORIGINAL COMBO CHARGE >=============================
[Statedef 7260]
physics = N
velset = 0, 0
anim = 9741

[State 7260, BG PalFX]
type = bgpalfx
trigger1 = !time
time = 30
add = -255, -255, -128
sinadd = 255, 255, 128, 120
[State 7260, palfx]
type = palfx
trigger1 = time = 20
time = 11
sinadd = -255, -255, -255, 40
persistent = 0

[State 7260, main]
type = explod
trigger1 = !time
anim = 7260
ID = 7260
sprpriority = 5
postype = p1
removetime = 20
supermovetime = 300
[State 7260, sphere]
type = explod
trigger1 = !time
anim = 7500
ID = 7261
sprpriority = 5
postype = p1
removetime = -1
supermovetime = 300
[State 7260, Circle]
type = explod
trigger1 = time = 20
anim = 7501
ID = 7262
sprpriority = 5
postype = p1
removetime = -1
supermovetime = 300

[State 7260, shrinkcircle]
type = modifyexplod
trigger1 = 1
ID = 7260
scale = 2 -(0.1 * time), 2 -(0.1 * time)
[State 7260, shrinkcircle]
type = modifyexplod
trigger1 = 1
ID = 7261
scale = 0.017 * time, 0.017 * time
[State 7260, shrinkcircle]
type = modifyexplod
trigger1 = time >= 20
ID = 7262
scale = 0.1 + (0.1 * (time -20)), 0.1 + (0.1 * (time -20))

[State 7260, No Shadow]
type = assertspecial
trigger1 = 1
flag = noshadow
flag2 = invisible
ignorehitpause = 1
[State 7260, bound]
type = screenbound
trigger1 = 1
value = 0
[State 7260, Safety]
type = removeexplod
trigger1 = time >= 30
[State 6110, End]
type = destroyself
trigger1 = time >= 30


;========================< COUNTER F.X. >=============================
[Statedef 7400]
type = A
physics = N
anim = 9741

[State 7400, superchargebgpalfx]
type = allpalfx
trigger1 = !time
time = 8
add = 255, 255, 255
sinadd = -255, -255, -255, 32

[State 7400, slow]
type = pause
trigger1 = time < 8
time = 1

[State 7400, End]
type = destroyself
trigger1 = time >= 8


;========================< SUPER FINISH >=============================
[Statedef 7600]
type = A
velset = 0, 0
anim = 9741

[State 7600, Bind]
type = posset
trigger1 = 1
x = 0
y = ceil(pos y -screenpos y) + 120

[State 7600, Snd]
type = playsnd
trigger1 = !time
value = 7, 6
channel = 1

[State 7600, Snd]
type = sndpan
trigger1 = time <= 80
channel = 1
abspan = ceil(sin(gametime / 20.0 * pi) * 160)

[State 7600, Snd]
type = assertspecial
trigger1 = time <= 80
flag = nomusic

[State 7600, BG PalFX]
type = bgpalfx
trigger1 = time < 10
time = 1
mul = 0, 0, 0
[State 7600, BG PalFX]
type = bgpalfx
trigger1 = time = [10, 20]
time = 1
mul = (time -10) * 25, (time -10) * 25, (time -10) * 25

[State 7600, MvC]
type = explod
trigger1 = !time
anim = 7601
ID = 7601
ontop = 1
postype = left
pos = 0, 0
scale = 1, (1.3333333 * gameheight / gamewidth)
bindtime = -1
removetime = 20
pausemovetime = -1
supermovetime = -1

[State 7600, Fade]
type = modifyexplod
trigger1 = numexplod(7601)
ID = 7601
trans = addalpha
alpha = ifelse(time < 10, 256, ceil(256 -(time -10) * 25.6)), 256

[State 7600, Circle]
type = explod
trigger1 = !time
anim = 7600
ID = 7600
sprpriority = -255
postype = left
pos = 160, floor((1.3333333 * gameheight / gamewidth) * 120)
bindtime = -1
removetime = 20
pausemovetime = -1
supermovetime = -1

[State 7600, Scale]
type = modifyexplod
trigger1 = numexplod(7600)
ID = 7600
scale = (0.25 * time) * (1.3333333 * gameheight / gamewidth), (0.25 * time) * (1.3333333 * gameheight / gamewidth)
trans = addalpha
alpha = ceil(256 -12.8 * time), 256

[State 7600, ten]
type = explod
trigger1 = !time
anim = 7605
ID = 7605
sprpriority = -5
postype = left
pos = 160, floor((1.3333333 * gameheight / gamewidth) * 120)
scale = 0.5, 0.5
bindtime = -1
removetime = 40
pausemovetime = -1
supermovetime = -1

[State 7600, Shrink]
type = modifyexplod
trigger1 = numexplod(7605)
ID = 7605
scale = (0.2 * time) * (1.3333333 * gameheight / gamewidth), (0.2 * time) * (1.3333333 * gameheight / gamewidth)
trans = addalpha
alpha = ceil(256 -12.8 * time), 256

[State 7600, Safety]
type = removeexplod
trigger1 = time > 20 && numexplod
trigger2 = roundstate < 3
[State 7600, nobars]
type = assertspecial
trigger1 = time <= 20
flag = nobardisplay
flag2 = noFG
flag3 = globalnoshadow

[State 7600, End]
type = destroyself
trigger1 = roundstate < 3


;========================< KANJI >=============================
[Statedef 6012]
physics = N
velset = 0, 0
anim = 6012
sprpriority = -3

[State 6012, glow]
type = palfx
trigger1 = !time
time = 9999
add = 50, 0, 0
sinadd = 100, 0, 0, 80

[State 6012, Bind]
type = bindtoroot
trigger1 = 1
pos = ifelse((root, facing = 1), 0, 32), 0


;========================< KANJI B.G. >=============================
[Statedef 7610]
physics = N
velset = 0, 0
anim = 9741

[State 7610, Assert]
type = assertspecial
trigger1 = time < 120
flag = nomusic

[State 7610, Snd]
type = playsnd
trigger1 = !time
value = 7, 7

[State 7610, black]
type = bgpalfx
trigger1 = time < 120
time = 2
add = -255, -255, -255

[State 7610, black]
type = bgpalfx
trigger1 = time = 120
time = 60
add = -255, -255, -255
sinadd = 255, 255, 255, 240

[State 6200, Fade]
type = palfx
trigger1 = time = 120
time = 60
sinadd = -255, -255, -255, 240

[State 7610, Kanji]
type = explod
trigger1 = !time
anim = 7610
ID = 7610
sprpriority = -5
postype = left
pos = 160, floor((1.3333333 * gameheight / gamewidth) * 120)
scale = 0.5 * (1.3333333 * gameheight / gamewidth), 0.5 * (1.3333333 * gameheight / gamewidth)
bindtime = -1

[State 7610, Shrink]
type = modifyexplod
trigger1 = time < 6
ID = 7610
scale = (5.5 - time) * (1.3333333 * gameheight / gamewidth), (5.5 - time) * (1.3333333 * gameheight / gamewidth)

[State 7610, Shrink]
type = modifyexplod
trigger1 = time >= 6
ID = 7610
scale = 0.5 * (1.3333333 * gameheight / gamewidth), 0.5 * (1.3333333 * gameheight / gamewidth)

[State 7610, flamebg]
type = explod
trigger1 = !time
anim = 7611
ID = 7611
sprpriority = 5
postype = left
pos = 0, floor((gameheight / gamewidth) * 80) - 60
bindtime = -1
removetime = -1
pausemovetime = 300
supermovetime = 300

[State 7610, Assert]
type = assertspecial
trigger1 = 1
flag = invisible
flag2 = nobardisplay
ignorehitpause = 1

[State 7610, Assert]
type = assertspecial
trigger1 = time < 120
flag = noFG
ignorehitpause = 1

[State 7610, Safety]
type = removeexplod
trigger1 = time >= 180

[State 7610, End]
type = destroyself
trigger1 = time >= 180


;========================< CAMERA CONTROL >=============================
[Statedef 7650]
physics = N
velset = 0, 0
anim = 9741
[State 7650, bound]
type = screenbound
trigger1 = 1
value = 0
movecamera = 1, 1
[State 7650, End]
type = destroyself
trigger1 = roundstate = 2

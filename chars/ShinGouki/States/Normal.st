; _____________________________________
;| Shin Gouki by Phantom.of.the.Server |
; ¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯¯
;============================================================================================
;=====================================< NORMAL STATES >========================================
;============================================================================================

;=========< Stand Light Punch >==========
[Statedef 200]
type = S
movetype = A
physics = S
juggle = ifelse(numtarget, 15, 8)
anim = ifelse(p2bodydist x <= (const(size.xscale) * 25), 201, 200)
velset = 0, 0
poweradd = 0
sprpriority = 1
ctrl = 0
facep2 = 1

[State 200, Snd]
type = playsnd
trigger1 = animelem = 2 && random < 200
value = 3, ifelse(random < 500, 1, 9)
channel = 0
[State 200, Snd]
type = playsnd
trigger1 = animelem = 2
value = 1, 1
channel = 1

[State 200, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = anim = 200 && animelem = 2
attr = S, NA
damage = 20, 0
animtype = Light
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 4, 8), ifelse(var(20), 4, 8)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -84
hitsound = s2, 1
guardsound = s2, 0
ground.type = High
ground.slidetime = 12 + 2 * var(29)
ground.hittime = 12 + 2 * var(29)
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = 72 * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 200, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = anim = 201 && animelem = 2
attr = S, NA
damage = 23, 0
animtype = Light
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 4, 8), ifelse(var(20), 4, 8)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -20, -92
hitsound = s2, 1
guardsound = s2, 0
ground.type = High
ground.slidetime = 12 + 2 * var(29)
ground.hittime = 12 + 2 * var(29)
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = 72 * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State Idle]
type = statetypeset
trigger1 = animelemtime(3) >= 0
movetype = I

[State 200, End]
type = changestate
trigger1 = !animtime
value = 0
ctrl = 1


;=========< Stand Medium Punch >==========
[Statedef 205]
type = S
movetype = A
physics = S
juggle = ifelse(numtarget, 15, 8)
anim = ifelse(p2bodydist x <= (const(size.xscale) * 25), 206, 205)
velset = 0, 0
poweradd = 18 * !var(20)
sprpriority = 1
ctrl = 0
facep2 = 1

[State 205, Snd]
type = playsnd
trigger1 = animelem = 3 && random < 350
value = 3, ifelse(random < 500, 9, 11)
channel = 0
[State 205, Snd]
type = playsnd
trigger1 = animelem = 3
value = 1, 2
channel = 1

[State 205, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 3 && anim = 205
attr = S, NA
damage = 43, 0
animtype = Medium
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 5, 10), ifelse(var(20), 5, 10)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -81
hitsound = s2, 2
guardsound = s2, 0
ground.type = High
ground.slidetime = 15 + 3 * var(29)
ground.hittime = 15 + 3 * var(29)
ground.velocity = ifelse(var(20), -4.30, -7.51), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 205, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 3 && anim = 206
attr = S, NA
damage = 40, 0
animtype = Medium
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 5, 10), ifelse(var(20), 5, 10)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -20, -60
hitsound = s2, 2
guardsound = s2, 0
ground.type = Low
ground.slidetime = 15 + 3 * var(29)
ground.hittime = 15 + 3 * var(29)
ground.velocity = ifelse(var(20), -4.30, -7.51), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State Idle]
type = statetypeset
trigger1 = animelemtime(4) >= 0
movetype = I

[State 205, End]
type = changestate
trigger1 = !animtime
value = 0
ctrl = 1


;=========< Zugai Hasatsu >==========
[Statedef 207]
type = S
movetype = A
physics = S
juggle = ifelse(numtarget, 15, 8)
anim = 207
velset = 0, 0
poweradd = 18 * !var(20)
sprpriority = 1
ctrl = 0
facep2 = 1

[State 207, width]
type = width
trigger1 = 1
value = 20, 0

[State 207, Snd]
type = playsnd
trigger1 = animelem = 4
value = 1, 3
channel = 1
[State 207, Snd]
type = playsnd
trigger1 = animelem = 4
value = 3, 8
channel = 0

[State 207, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 5
attr = S, NA
damage = 25, 0
animtype = Medium
air.animtype = Back
fall.animtype = Back
guardflag = HA
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 5, 10), ifelse(var(20), 5, 10)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -85
hitsound = s2, 2
guardsound = s2, 0
ground.type = High
ground.slidetime = 15 + 3 * var(29)
ground.hittime = 15 + 3 * var(29)
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 207, hit2]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 6
attr = S, NA
damage = 25, 0
animtype = Medium
air.animtype = Back
fall.animtype = Back
guardflag = HA
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 5, 10), ifelse(var(20), 5, 10)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -44
hitsound = s2, 2
guardsound = s2, 0
ground.type = Low
ground.slidetime = 18 + 4 * var(29)
ground.hittime = 18 + 4 * var(29)
ground.velocity = ifelse(var(20), -4.30, -7.51), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State Idle]
type = statetypeset
trigger1 = animelemtime(7) >= 0
movetype = I

[State 207, End]
type = changestate
trigger1 = !animtime
value = 0
ctrl = 1


;=========< Stand Hard Punch >==========
[Statedef 210]
type = S
movetype = A
physics = S
juggle = ifelse(numtarget, 15, 8)
anim = ifelse((prevstateno != 710 && p2bodydist x <= (const(size.xscale) * 25)), 211, 210)
velset = 0, 0
poweradd = 18 * !var(20)
sprpriority = 1
ctrl = 0
facep2 = 1

[State 210, Snd]
type = playsnd
trigger1 = animelem = 3 && random < 500
value = 3, ifelse(random < 500, 9, 11)
channel = 0
[State 210, Snd]
type = playsnd
trigger1 = animelem = 3
value = 1, 3
channel = 1

[State 210, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 4 && anim = 210
attr = S, NA
damage = 65, 0
animtype = Hard
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -68
hitsound = s2, 3
guardsound = s2, 0
ground.type = Low
ground.slidetime = 18 + 4 * var(29)
ground.hittime = 18 + 4 * var(29)
ground.velocity = ifelse(var(20), -5.77, -10.52), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -6.82, -12.74)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 216) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 210, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 4 && anim = 210 && prevstateno = 710
attr = S, NA
damage = 65, 0
animtype = Hard
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -68
hitsound = s2, 3
guardsound = s2, 0
ground.type = Low
ground.slidetime = 18 + 4 * var(29)
ground.hittime = 18 + 4 * var(29)
ground.velocity = -6, -3
air.type = Low
air.hittime = 18
air.velocity = -6, -3
air.cornerpush.veloff = 0
ground.cornerpush.veloff = 0
guard.velocity = ifelse(var(20), -5.77, -10.52)
guard.cornerpush.veloff = 0;ifelse(var(20), -6.82, -12.74)
getpower = ifelse(p2movetype = H, 72, 216) * !var(20), 36 * !var(20)
givepower = 36, 36
fall = 1
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120

[State 210, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 3 && anim = 211
attr = S, NA
damage = 60, 0
animtype = Hard
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -75
hitsound = s2, 3
guardsound = s2, 0
ground.type = Low
ground.slidetime = 18 + 4 * var(29)
ground.hittime = 18 + 4 * var(29)
ground.velocity = ifelse(var(20), -5.77, -10.52), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -6.82, -12.74)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 216) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)
forcestand = 1

[State 210, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 5 && anim = 211
trigger1 = !movecontact && !movereversed
attr = S, NA
damage = 60, 0
animtype = Hard
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -105
hitsound = s2, 3
guardsound = s2, 0
ground.type = High
ground.slidetime = 18 + 4 * var(29)
ground.hittime = 18 + 4 * var(29)
ground.velocity = ifelse(var(20), -5.77, -10.52), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -6.82, -12.74)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 216) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State Idle]
type = statetypeset
trigger1 = animelemtime(5) >= 0
movetype = I

[State 210, End]
type = changestate
trigger1 = !animtime
value = 0
ctrl = 1


;=========< Stand Light Kick >==========
[Statedef 215]
type = S
movetype = A
physics = S
juggle = ifelse(numtarget, 15, 8)
anim = 215
velset = 0, 0
poweradd = 0
sprpriority = 1
ctrl = 0
facep2 = 1

[State 215, Snd]
type = playsnd
trigger1 = animelem = 3 && random < 200
value = 3, ifelse(random < 500, 1, 9)
channel = 0
[State 215, Snd]
type = playsnd
trigger1 = animelem = 3
value = 1, 4
channel = 1

[State 215, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 3
attr = S, NA
damage = 20, 0
animtype = Light
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 4, 8), ifelse(var(20), 4, 8)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -35
hitsound = s2, 4
guardsound = s2, 0
ground.type = Low
ground.slidetime = 12 + 2 * var(29)
ground.hittime = 12 + 2 * var(29)
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = 72 * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State Idle]
type = statetypeset
trigger1 = animelemtime(4) >= 0
movetype = I

[State 215, End]
type = changestate
trigger1 = !animtime
value = 0
ctrl = 1


;=========< Stand Medium Kick >==========
[Statedef 220]
type = S
movetype = A
physics = S
juggle = ifelse(numtarget, 15, 8)
anim = ifelse((prevstateno != 710 && p2bodydist x <= (const(size.xscale) * 25)), 221, 220)
velset = 0, 0
poweradd = 18 * !var(20)
sprpriority = 1
ctrl = 0
facep2 = 1

[State 220, Snd]
type = playsnd
trigger1 = animelem = 3 && random < 350
value = 3, ifelse(random < 500, 9, 11)
channel = 0
[State 220, Snd]
type = playsnd
trigger1 = animelem = 3
value = 1, 5
channel = 1

[State 220, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 3 && anim = 220
attr = S, NA
damage = 40, 0
animtype = Medium
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 5, 10), ifelse(var(20), 5, 10)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -91
hitsound = s2, 5
guardsound = s2, 0
ground.type = High
ground.slidetime = 15 + 3 * var(29)
ground.hittime = 15 + 3 * var(29)
ground.velocity = ifelse(var(20), -4.30, -7.51), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 220, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 3 && anim = 221
attr = S, NA
damage = 40, 0
animtype = Medium
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 5, 10), ifelse(var(20), 5, 10)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -20, -74
hitsound = s2, 5
guardsound = s2, 0
ground.type = Low
ground.slidetime = 15 + 3 * var(29)
ground.hittime = 15 + 3 * var(29)
ground.velocity = ifelse(var(20), -4.30, -7.51), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)
forcestand = 1

[State Idle]
type = statetypeset
trigger1 = animelemtime(4) >= 0
movetype = I

[State 220, End]
type = changestate
trigger1 = !animtime
value = 0
ctrl = 1


;=========< SENPUU KYAKU >==========
[Statedef 222]
type = A
movetype = A
physics = N
juggle = ifelse(numtarget, 15, 8)
anim = 222
velset = vel x, 0
poweradd = 18 * !var(20)
sprpriority = 1
ctrl = 0
facep2 = 1

[State 222, Snd]
type = playsnd
trigger1 = animelem = 1
value = 0, 40
[State 222, Snd]
type = playsnd
trigger1 = animelem = 5
value = 1, 6
channel = 1
[State 222, Snd]
type = playsnd
trigger1 = animelem = 3
value = 3, 11
channel = 0

[State 222, Dust]
type = explod
trigger1 = animelem = 1
anim = 6052
ID = 6052
sprpriority = 3
postype = p1
pos = 0, 0
vel = 0.33, -.5
ownpal = 1
scale = 0.66, 0.66

[State 222, Vel]
type = velset
trigger1 = animelem = 2
x = ifelse(prevstateno = 100, 5, 3)
y = -4.5
[State 222, Accel]
type = gravity
trigger1 = animelemtime(2) > 0

[State 222, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 6
attr = A, NA
damage = 50, 0
animtype = Medium
air.animtype = Back
fall.animtype = Back
guardflag = MA
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 5, 10), ifelse(var(20), 5, 10)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -65
hitsound = s2, 6
guardsound = s2, 0
ground.type = High
ground.slidetime = 15 + 3 * var(29)
ground.hittime = 15 + 3 * var(29)
ground.velocity = ifelse(var(20), -4.30, -7.51), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State Idle]
type = statetypeset
trigger1 = animelemtime(7) >= 0
movetype = I

[State 222, End]
type = changestate
trigger1 = pos y > 10 && vel y > 0
value = 52
ctrl = 1


;=========< Stand Hard Kick >==========
[Statedef 225]
type = S
movetype = A
physics = S
juggle = ifelse(numtarget, 15, 8)
anim = ifelse(p2bodydist x <= (const(size.xscale) * 25), 226, 225)
velset = 0, 0
poweradd = 18 * !var(20)
sprpriority = ifelse(p2bodydist x <= (const(size.xscale) * 25), 1, -1)
ctrl = 0
facep2 = 1

[State 225, Snd]
type = playsnd
trigger1 = animelem = 4 && random < 500
value = 3, ifelse(random < 500, 9, 11)
channel = 0
[State 225, Snd]
type = playsnd
trigger1 = anim = 225 && animelem = 4
trigger2 = anim = 226 && animelem = 5
value = 1, 6
channel = 1

[State 225, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 5 && anim = 225
attr = S, NA
damage = 60, 0
animtype = Hard
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -87
hitsound = s2, 6
guardsound = s2, 0
ground.type = High
ground.slidetime = 18 + 4 * var(29)
ground.hittime = 18 + 4 * var(29)
ground.velocity = ifelse(var(20), -5.77, -10.52), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -6.82, -12.74)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 216) * !var(20), 36 * !var(20)
givepower = 36, 36
sprpriority = -1
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 225, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 3 && anim = 226
attr = S, NA
damage = 33, 0
animtype = Hard
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -97
hitsound = s2, 6
guardsound = s2, 0
ground.type = High
ground.slidetime = 18 + 4 * var(29)
ground.hittime = 18 + 4 * var(29)
ground.velocity = ifelse(var(20), -4.30, -7.51), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 216) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 225, Hit]
type = hitdef
trigger1 = p2stateno != 5040
trigger1 = animelem = 5 && anim = 226
attr = S, NA
damage = 33, 0
animtype = Hard
air.animtype = Hard
fall.animtype = Hard
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -69
hitsound = s2, 6
guardsound = s2, 0
ground.type = Low
ground.slidetime = 18 + 4 * var(29)
ground.hittime = 12 + 2 * var(29)
ground.velocity = ifelse(var(20), -5.77, -10.52), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -6.82, -12.74)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -4, -8), 8
air.fall = 1
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 216) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120

[State Idle]
type = statetypeset
trigger1 = animelemtime(6) >= 0
movetype = I

[State 225, End]
type = changestate
trigger1 = !animtime
value = 0
ctrl = 1


;=========< Crouch Light Punch >==========
[Statedef 230]
type = C
movetype = A
physics = C
juggle = ifelse(numtarget, 15, 8)
anim = 230
velset = 0, 0
poweradd = 0
sprpriority = 1
ctrl = 0
facep2 = 1

[State 230, Snd]
type = playsnd
trigger1 = animelem = 2 && random < 200
value = 3, ifelse(random < 500, 1, 9)
channel = 0
[State 230, Snd]
type = playsnd
trigger1 = animelem = 2
value = 1, 1
channel = 1

[State 230, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 2
attr = C, NA
damage = 20, 0
animtype = Light
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 4, 8), ifelse(var(20), 4, 8)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -50
hitsound = s2, 1
guardsound = s2, 0
ground.type = Low
ground.slidetime = 12 + 2 * var(29)
ground.hittime = 12 + 2 * var(29)
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = 72 * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State Idle]
type = statetypeset
trigger1 = animelemtime(3) >= 0
movetype = I

[State 230, End]
type = changestate
trigger1 = !animtime
value = 11
ctrl = 1


;=========< Crouch Medium Punch >==========
[Statedef 235]
type = C
movetype = A
physics = C
juggle = ifelse(numtarget, 15, 8)
anim = 235
velset = 0, 0
poweradd = 18 * !var(20)
sprpriority = 1
ctrl = 0
facep2 = 1

[State 235, Snd]
type = playsnd
trigger1 = animelem = 3 && random < 350
value = 3, ifelse(random < 500, 9, 11)
channel = 0
[State 235, Snd]
type = playsnd
trigger1 = animelem = 3
value = 1, 2
channel = 1

[State 235, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 3
attr = C, NA
damage = 40, 0
animtype = Medium
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 5, 10), ifelse(var(20), 5, 10)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -46
hitsound = s2, 2
guardsound = s2, 0
ground.type = Low
ground.slidetime = 15 + 3 * var(29)
ground.hittime = 15 + 3 * var(29)
ground.velocity = ifelse(var(20), -4.30, -7.51), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State Idle]
type = statetypeset
trigger1 = animelemtime(4) >= 0
movetype = I

[State 235, End]
type = changestate
trigger1 = !animtime
value = 11
ctrl = 1


;=========< Crouch Hard Punch >==========
[Statedef 240]
type = C
movetype = A
physics = C
juggle = ifelse(numtarget, 15, 8)
anim = 240
velset = 0, 0
poweradd = 18 * !var(20)
sprpriority = 1
ctrl = 0
facep2 = 1

[State 240, Snd]
type = playsnd
trigger1 = animelem = 4 && random < 500
value = 3, ifelse(random < 500, 9, 11)
channel = 0
[State 240, Snd]
type = playsnd
trigger1 = animelem = 4
value = 1, 3
channel = 1

[State 240, Hit]
type = hitdef
triggerall = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 4
trigger2 = animelem = 5 && !movecontact && !movereversed
attr = C, NA
damage = 60, 0
animtype = Hard
air.animtype = Back
fall.animtype = Back
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, ifelse((animelem = 5), -118, -73)
hitsound = s2, 3
guardsound = s2, 0
ground.type = Low
ground.slidetime = 18 + 4 * var(29)
ground.hittime = 18 + 4 * var(29)
ground.velocity = ifelse(var(20), -5.77, -10.52), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -6.82, -12.74)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 216) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)
forcestand = 1

[State 240, Hit]
type = hitdef
triggerall = var(20) && p2stateno != 5040
trigger1 = animelem = 4
trigger2 = animelem = 5 && !movecontact && !movereversed
attr = C, NA
damage = 60, 0
animtype = Up
air.animtype = Up
fall.animtype = Up
guardflag = M
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, ifelse((animelem = 5), -118, -73)
hitsound = s2, 3
guardsound = s2, 0
ground.type = Low
ground.slidetime = 18 + 4 * var(29)
ground.hittime = 18 + 4 * var(29)
ground.velocity = 0, -11
air.type = Low
air.velocity = 0, -11
air.cornerpush.veloff = 0
getpower = 0
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall = 1

[State Idle]
type = statetypeset
trigger1 = animelemtime(6) >= 0
movetype = I

[State 240, End]
type = changestate
trigger1 = !animtime
value = 11
ctrl = 1


;=========< Crouch Light Kick >==========
[Statedef 245]
type = C
movetype = A
physics = C
juggle = ifelse(numtarget, 15, 8)
anim = 245
velset = 0, 0
poweradd = 0
sprpriority = 1
ctrl = 0 
facep2 = 1

[State 245, Snd]
type = playsnd
trigger1 = animelem = 2 && random < 200
value = 3, ifelse(random < 500, 1, 9)
channel = 0
[State 245, Snd]
type = playsnd
trigger1 = animelem = 2
value = 1, 4
channel = 1

[State 245, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 2
attr = C, NA
damage = 20, 0 
animtype = Light
air.animtype = Back
fall.animtype = Back
guardflag = L
hitflag = MAFD
priority = 4, Hit
pausetime = ifelse(var(20), 4, 8), ifelse(var(20), 4, 8)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -4
hitsound = s2, 4
guardsound = s2, 0
ground.type = Low
ground.slidetime = 12 + 2 * var(29)
ground.hittime = 12 + 2 * var(29)
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
down.velocity = -4, -6
down.cornerpush.veloff = -16
air.cornerpush.veloff = 0
getpower = 72 * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State Idle]
type = statetypeset
trigger1 = animelemtime(3) >= 0
movetype = I

[State 245, End]
type = changestate
trigger1 = !animtime
value = 11
ctrl = 1


;=========< Crouch Medium Kick >==========
[Statedef 250]
type = C
movetype = A
physics = C
juggle = ifelse(numtarget, 15, 8)
anim = 250
velset = 0, 0
poweradd = 18 * !var(20)
sprpriority = 1
ctrl = 0 
facep2 = 1

[State 250, Snd]
type = playsnd
trigger1 = animelem = 3 && random < 350
value = 3, ifelse(random < 500, 9, 11)
channel = 0
[State 250, Snd]
type = playsnd
trigger1 = animelem = 3
value = 1, 5
channel = 1

[State 250, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 3
attr = C, NA
damage = 40, 0 
animtype = Medium
air.animtype = Back
fall.animtype = Back
guardflag = L
hitflag = MAFD
priority = 4, Hit
pausetime = ifelse(var(20), 5, 10), ifelse(var(20), 5, 10)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -6
hitsound = s2, 5
guardsound = s2, 0
ground.type = Low
ground.slidetime = 15 + 3 * var(29)
ground.hittime = 15 + 3 * var(29)
ground.velocity = ifelse(var(20), -4.30, -7.51), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = Low
air.hittime = 18
air.velocity = ifelse(var(20), -2.5, -5), -7
down.velocity = -4, -6
down.cornerpush.veloff = -16
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State Idle]
type = statetypeset
trigger1 = animelemtime(4) >= 0
movetype = I

[State 250, End]
type = changestate
trigger1 = !animtime
value = 11
ctrl = 1


;=========< Crouch Hard Kick >==========
[Statedef 255]
type = C
movetype = A
physics = C
juggle = ifelse((p2movetype = H && p2statetype = A), 15, 7)
anim = 255
velset = 0, 0
poweradd = 18 * (!var(20) && prevstateno != 750)
sprpriority = 1
ctrl = 0
facep2 = 1

[State 255, Snd]
type = playsnd
trigger1 = animelem = 3 && random < 500
value = 3, ifelse(random < 500, 9, 11)
channel = 0
[State 255, Snd]
type = playsnd
trigger1 = animelem = 3
value = 1, 6
channel = 1

[State NHB -ZC]
type = nothitby
trigger1 = prevstateno = 750 && animelemtime(3) < 0
value = SCA
time = 1

[State 255, Hit]
type = hitdef
trigger1 = ((p2stateno != [5020, 5050]) || var(20) || roundstate = 3) && p2stateno != 5040
trigger1 = animelem = 3
attr = C, NA
damage = 60, 0 
animtype = Hard
air.animtype = Back
fall.animtype = Back
guardflag = L
hitflag = MAFD
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -6
hitsound = s2, 6
guardsound = s2, 0
ground.type = Trip
ground.slidetime = 18 + 4 * var(29)
ground.hittime = 18 + 4 * var(29)
ground.velocity = -1, -8.24
air.type = Trip
air.velocity = -1, -8.24
down.velocity = -4, -6
down.cornerpush.veloff = -16
air.cornerpush.veloff = 0
ground.cornerpush.veloff = 0
guard.velocity = ifelse(var(20), -5.77, -10.52)
guard.cornerpush.veloff = 0;ifelse(var(20), -6.82, -12.74)
getpower = ifelse(p2movetype = H, 72, 216) * !var(20), 36 * !var(20)
givepower = 36, 36
fall = 1
fall.recover = 1
fall.recovertime = 120
yaccel = 0.677
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0

[State 255, Hit]
type = hitdef
trigger1 = animelem = 3 && prevstateno = 750
attr = C, SA
damage = 60, 0 
animtype = Back
air.animtype = Back
fall.animtype = Back
guardflag = L
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -6
hitsound = s2, 6
guardsound = s2, 0
ground.type = Low
ground.slidetime = 18 + 4 * var(29)
ground.hittime = 18 + 4 * var(29)
ground.velocity = -5, -6.2
air.type = Low
air.velocity = -5, -6.2
down.cornerpush.veloff = -16
air.cornerpush.veloff = 0
ground.cornerpush.veloff = 0
guard.velocity = ifelse(var(20), -5.77, -10.52)
guard.cornerpush.veloff = 0;ifelse(var(20), -6.82, -12.74)
getpower = 0, 0
givepower = 36, 36
fall = 1
fall.recover = 1
fall.recovertime = 120
yaccel = 0.6
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0

[State Idle]
type = statetypeset
trigger1 = animelemtime(4) >= 0
movetype = I

[State 255, End]
type = changestate
trigger1 = !animtime
value = 11
ctrl = 1


;=========< Air Light Punch >==========
[Statedef 260]
type = A
movetype = A
physics = A
juggle = ifelse((prevstateno = [260, 289]), 0, 7)
anim = 260
poweradd = 0
sprpriority = 1
ctrl = 0 

[State 260, Snd]
type = playsnd
trigger1 = animelem = 2 && random < 200
value = 3, ifelse(random < 500, 1, 9)
channel = 0
[State 260, Snd]
type = playsnd
trigger1 = animelem = 2
value = 1, 1
channel = 1

[State 260, Hit]
type = hitdef
trigger1 = animelem = 2
attr = A, NA
damage = 25, 0 
animtype = Light
air.animtype = Back
fall.animtype = Back
guardflag = HA
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 4, 8), ifelse(var(20), 4, 8)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -58
hitsound = s2, 1
guardsound = s2, 0
ground.type = High
ground.hittime = 12
ground.slidetime = 12
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
guard.velocity = ifelse(var(20), -2.66, -4.49)
guard.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
air.type = Low
air.hittime = 18
air.velocity = 0, -2.5
air.cornerpush.veloff = 0
getpower = 72 * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 260, Type]
type = statetypeset
trigger1 = animelemtime(6) >= 0
movetype = I


;=========< Air Medium Punch >==========
[Statedef 265]
type = A
movetype = A
physics = A
juggle = ifelse((prevstateno = [260, 289]), 0, 7)
anim = 265
poweradd = 18 * !var(20)
sprpriority = 1
ctrl = 0 

[State 265, Snd]
type = playsnd
trigger1 = animelem = 3 && random < 350
value = 3, ifelse(random < 500, 9, 11)
channel = 0
[State 265, Snd]
type = playsnd
trigger1 = animelem = 3
value = 1, 2
channel = 1

[State 265, Hit]
type = hitdef
trigger1 = animelem = 4
attr = A, NA
damage = 40, 0 
animtype = Medium 
air.animtype = Back
fall.animtype = Back
guardflag = HA
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 5, 10), ifelse(var(20), 5, 10)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -61
hitsound = s2, 2
guardsound = s2, 0
ground.type = High
ground.hittime = 12
ground.slidetime = 12
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
guard.velocity = ifelse(var(20), -4.30, -7.51)
guard.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = Low
air.hittime = 18
air.velocity = 0, -2.5
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 265, Type]
type = statetypeset
trigger1 = animelemtime(5) >= 0
movetype = I


;=========< Air Hard Punch >==========
[Statedef 270]
type = A
movetype = A
physics = A
juggle = ifelse((prevstateno = [260, 289]), 0, 7)
anim = 270
poweradd = 18 * !var(20)
sprpriority = 1
ctrl = 0

[State 270, Snd]
type = playsnd
trigger1 = animelem = 3 && random < 500
value = 3, ifelse(random < 500, 9, 11)
channel = 0
[State 270, Snd]
type = playsnd
trigger1 = animelem = 3
value = 1, 3
channel = 1

[State 270, Hit]
type = hitdef
trigger1 = animelem = 4
attr = A, NA
damage = 60, 0 
animtype = Hard
air.animtype = Back
fall.animtype = Back
guardflag = HA
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -61
hitsound = s2, 3
guardsound = s2, 0
ground.type = High
ground.hittime = 12
ground.slidetime = 12
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
guard.velocity = ifelse(var(20), -5.77, -10.52)
guard.cornerpush.veloff = 0;ifelse(var(20), -6.82, -12.74)
air.type = Low
air.hittime = 18
air.velocity = -7.5, 4
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 216) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
fall = (prevstateno = [260, 285])
fall.recover = 1
fall.recovertime = 120
airguard.cornerpush.veloff = 0
forcenofall = !(prevstateno = [260, 285])

[State 270, Type]
type = statetypeset
trigger1 = animelemtime(5) >= 0
movetype = I


;=========< Air Light Kick >==========
[Statedef 275]
type = A
movetype = A
physics = A
juggle = ifelse((prevstateno = [260, 289]), 0, 7)
anim = ifelse((vel x || var(9) = 2 || (prevstateno = [260, 285])), 275, 276)
poweradd = 0
sprpriority = 1
ctrl = 0 

[State 275, Snd]
type = playsnd
trigger1 = animelem = 2 && random < 200
value = 3, ifelse(random < 500, 1, 9)
channel = 0
[State 275, Snd]
type = playsnd
trigger1 = animelem = 2
value = 1, 4
channel = 1

[State 275, Hit]
type = hitdef
trigger1 = animelem = 2 && anim = 275
attr = A, NA
damage = 25, 0 
animtype = Light
air.animtype = Back
fall.animtype = Back
guardflag = HA
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 4, 8), ifelse(var(20), 4, 8)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -41
hitsound = s2, 4
guardsound = s2, 0
ground.type = High
ground.hittime = 12
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
guard.velocity = ifelse(var(20), -2.66, -4.49)
guard.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
air.type = Low
air.hittime = 18
air.velocity = 0, -2.5
air.cornerpush.veloff = 0
getpower = 72 * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 275, Hit]
type = hitdef
trigger1 = animelem = 3 && anim = 276
attr = A, NA
damage = 25, 0 
animtype = Light
air.animtype = Back
fall.animtype = Back
guardflag = HA
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 4, 8), ifelse(var(20), 4, 8)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -101
hitsound = s2, 4
guardsound = s2, 0
ground.type = High
ground.hittime = 12
ground.slidetime = 12
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
guard.velocity = ifelse(var(20), -2.66, -4.49)
guard.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
air.type = High
air.hittime = 18
air.velocity = 0, -2.5
air.cornerpush.veloff = 0
getpower = 72 * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 275, Type]
type = statetypeset
trigger1 = anim = 275 && animelemtime(7) >= 0
trigger2 = anim = 276 && animelemtime(4) >= 0
movetype = I


;=========< Air Medium Kick >==========
[Statedef 280]
type = A
movetype = A
physics = A
juggle = ifelse((prevstateno = [260, 289]), 0, 7)
anim = ifelse((vel x || var(9) = 2 || (prevstateno = [260, 285])), 280, 281)
poweradd = 18 * !var(20)
sprpriority = 1
ctrl = 0 

[State 280, Snd]
type = playsnd
trigger1 = animelem = 2 && random < 350
value = 3, ifelse(random < 500, 9, 11)
channel = 0
[State 280, Snd]
type = playsnd
trigger1 = anim = 280 && animelem = 3
trigger2 = anim = 281 && animelem = 2
value = 1, 5
channel = 1

[State 280, Hit]
type = hitdef
trigger1 = anim = 280 && animelem = 3
attr = A, NA
damage = 40, 0 
animtype = Medium 
air.animtype = Back
fall.animtype = Back
guardflag = HA
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 5, 10), ifelse(var(20), 5, 10)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -50
hitsound = s2, 5
guardsound = s2, 0
ground.type = High
ground.hittime = 12
ground.slidetime = 12
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
guard.velocity = ifelse(var(20), -4.30, -7.51)
guard.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = Low
air.hittime = 18
air.velocity = 0, -2.5
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 280, Hit]
type = hitdef
trigger1 = anim = 281 && animelem = 3
attr = A, NA
damage = 40, 0 
animtype = Medium 
air.animtype = Back
fall.animtype = Back
guardflag = HA
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 5, 10), ifelse(var(20), 5, 10)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -105
hitsound = s2, 5
guardsound = s2, 0
ground.type = High
ground.hittime = 12
ground.slidetime = 12
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
guard.velocity = ifelse(var(20), -4.30, -7.51)
guard.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = High
air.hittime = 18
air.velocity = 0, -2.5
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 280, Type]
type = statetypeset
trigger1 = animelemtime(4) >= 0
movetype = I


;======================< TENMA KUUJIN KYAKU >=============================
[Statedef 282]
type = A
physics = N
movetype = A
anim = 282
velset = 0.3 * vel x, 0.3 * vel y
juggle = 0
ctrl = 0
poweradd = 18 * !var(20)
sprpriority = -1


[State 282, Snd]
type = playsnd
trigger1 = animelem = 3
value = 1, 6
channel = 1
[State 282, Snd]
type = playsnd
trigger1 = animelem = 1
value = 3, 9
channel = 0

[State 282, Vel]
type = velset
trigger1 = animelemtime(3) > 0
x = 8.4
y = 9.6

[State 282, Hit]
type = hitdef
trigger1 = !movecontact && !movereversed
attr = A, NA
damage = 60, 0 
animtype = Hard
air.animtype = Back
fall.animtype = Back
guardflag = HA
hitflag = MAFD
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -15
hitsound = s2, 6
guardsound = s2, 0
ground.type = High
ground.hittime = 12
ground.slidetime = 12
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
guard.velocity = ifelse(var(20), -4.30, -7.51)
guard.cornerpush.veloff = 0;ifelse(var(20), -5.78, -10.47)
air.type = Low
air.velocity = -5, 6
down.hittime = 24
down.velocity = 0, 0
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 144) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
airguard.cornerpush.veloff = 0
fall.recover = 1
fall.recovertime = 120
forcenofall = !var(20)

[State 282, End]
type = changestate
trigger1 = pos y >= 0
value = 52


;=========< Air Hard Kick >==========
[Statedef 285]
type = A
movetype = A
physics = A
juggle = ifelse((prevstateno = [260, 289]), 0, 7)
anim = ifelse((vel x || var(9) = 2 || (prevstateno = [260, 285])), 285, 286)
poweradd = 18 * !var(20)
sprpriority = 1
ctrl = 0

[State 285, Snd]
type = playsnd
trigger1 = animelem = 3 && random < 500
value = 3, ifelse(random < 500, 9, 11)
channel = 0
[State 285, Snd]
type = playsnd
trigger1 = anim = 285 && animelem = 3
trigger2 = anim = 286 && animelem = 4
value = 1, 6
channel = 1

[State 285, Hit]
type = hitdef
trigger1 = anim = 285 && animelem = 3
attr = A, NA
damage = 60, 0 
animtype = Hard
air.animtype = Back
fall.animtype = Back
guardflag = HA
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -50
hitsound = s2, 6
guardsound = s2, 0
ground.type = High
ground.hittime = 12
ground.slidetime = 12
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
guard.velocity = ifelse(var(20), -5.77, -10.52)
guard.cornerpush.veloff = 0;ifelse(var(20), -6.82, -12.74)
air.type = Low
air.hittime = 18
air.velocity = -5, 0
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 216) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
fall = (prevstateno = [260, 285])
fall.recover = 1
fall.recovertime = 120
airguard.cornerpush.veloff = 0
forcenofall = !(prevstateno = [260, 285])

[State 285, Hit]
type = hitdef
trigger1 = anim = 286 && animelem = 3
attr = A, NA
damage = 60, 0 
animtype = Hard
air.animtype = Back
fall.animtype = Back
guardflag = HA
hitflag = MAF
priority = 4, Hit
pausetime = ifelse(var(20), 6, 12), ifelse(var(20), 6, 12)
sparkno = s7010 + (random % 8)
guard.sparkno = s7000
sparkxy = -10, -66
hitsound = s2, 6
guardsound = s2, 0
ground.type = High
ground.hittime = 12
ground.slidetime = 12
ground.velocity = ifelse(var(20), -2.66, -4.49), 0
ground.cornerpush.veloff = 0;ifelse(var(20), -4.33, -8.17)
guard.velocity = ifelse(var(20), -5.77, -10.52)
guard.cornerpush.veloff = 0;ifelse(var(20), -6.82, -12.74)
air.type = Low
air.hittime = 18
air.velocity = -4, -3
air.cornerpush.veloff = 0
getpower = ifelse(p2movetype = H, 72, 216) * !var(20), 36 * !var(20)
givepower = 36, 36
yaccel = 0.5
airguard.ctrltime = 16
airguard.velocity = ifelse(var(20), -4, -8), -1.5
fall = (prevstateno = [260, 285])
fall.recover = 1
fall.recovertime = 120
airguard.cornerpush.veloff = 0
forcenofall = !(prevstateno = [260, 285])

[State 285, Type]
type = statetypeset
trigger1 = anim = 285 && animelemtime(4) >= 0
trigger2 = anim = 285 && animelemtime(6) >= 0
movetype = I


;=========================================================================
;===============================< THROWS >==================================
;=========================================================================
;========< THROW CONTROLLER >==========
[Statedef 800]
type = S
movetype = A
physics = S
juggle = 15
velset = 0, 0
ctrl = 0
anim = 800
sprpriority = 1
facep2 = 1
poweradd = 18 * !var(20)

[State 800, Strength]
type = varset
trigger1 = !time && !AIlevel
var(10) = ifelse(command = "c", 4, ifelse(command = "a", 3, ifelse(command = "z", 2, 1)))
[State 800, Strength]
type = varset
trigger1 = !time && AIlevel
var(10) = ceil(random / 250)

[State 800, Tech Hit]
type = varset
trigger1 = 1
var(19) = 0
[State 800, Tech Hit]
type = varset
trigger1 = numenemy
trigger1 = enemynear, statetype != A
trigger1 = (enemynear, stateno = [0, 999]) && (enemynear, stateno != [100, 119]) && (enemynear, stateno != [200, 799])
var(19) = 7

[State 800, Do not guard]
type = attackdist
trigger1 = 1
value = 0

[State 800, Hit]
type = hitdef
trigger1 = animelem = 2
trigger1 = !var(38)
attr = S, NT
hitflag = M- 
priority = 1, Miss 
sparkno = -1
hitsound = s0, 800
p1facing = ifelse((!AIlevel && command = "holdfwd") || (AIlevel && frontedgedist < backedgedist), -1, 1)
p2facing = 1
p1stateno = ifelse((!AIlevel && command = "holdfwd") || (AIlevel && frontedgedist < backedgedist), 803, 801) 
guard.dist = 0
fall = 1
fall.recover = 1
fall.recovertime = 120

[State 800, misssnd]
type = playsnd
trigger1 = animelem = 4
value = 5000, 3
channel = 0

[State Idle]
type = statetypeset
trigger1 = animelemtime(3) >= 0
movetype = I

[State 800, End]
type = changestate
trigger1 = !animtime
value = 0
ctrl = 1


;--- PLAYER ---
[Statedef 801]
type = S
movetype = A
physics = S
anim = 801
ctrl = 0
sprpriority = 1

[State Tech Hit]
type = changestate
trigger1 = numtarget && var(19)
trigger1 = (target, command = "holdfwd" || target, command = "holdback") && target, command != "holdup" && target, command != "holddown"
trigger1 = (target, command = "b" || target, command = "c" || target, command = "y" || target, command = "z")
value = 890
[State Tech Hit]
type = varadd
trigger1 = var(19) > 0
var(19) = -1

[State Throw]
type = targetstate
trigger1 = !time
value = 802

[State Width]
type = width
trigger1 = 1
edge = 30, 0

[State 801, Bind 1]
type = targetbind
trigger1 = 1
pos = 45, 0

[State 801, End]
type = changestate
trigger1 = !animtime
value = ifelse(var(10) = 1 || var(10) = 2, 805, ifelse(var(10) = 3, 810, 815))
ctrl = 0


;--- PLAYER TURNS ---
[Statedef 803]
type = S
movetype = A
physics = S
anim = 803
ctrl = 0
sprpriority = 1

[State Tech Hit]
type = changestate
trigger1 = numtarget && var(19)
trigger1 = (target, command = "holdfwd" || target, command = "holdback") && target, command != "holdup" && target, command != "holddown"
trigger1 = (target, command = "b" || target, command = "c" || target, command = "y" || target, command = "z")
value = 890
[State Tech Hit]
type = varadd
trigger1 = var(19) > 0
var(19) = -1

[State Throw]
type = targetstate
trigger1 = !time
value = 804

[State Width]
type = width
trigger1 = 1
edge = 30, 30

[State 803, Bind 1]
type = targetbind
trigger1 = animelemtime(1) >= 0 && animelemtime(2) < 0
pos = -45, 0
[State 803, Bind 2]
type = targetbind
trigger1 = animelemtime(2) >= 0 && animelemtime(3) < 0
pos = -25, 0
[State 803, Bind 3]
type = targetbind
trigger1 = animelemtime(3) >= 0 && animelemtime(4) < 0
pos = 0, 0
[State 803, Bind 4]
type = targetbind
trigger1 = animelemtime(4) >= 0 && animelemtime(5) < 0
pos = 22, 0
[State 803, Bind 5]
type = targetbind
trigger1 = animelemtime(5) >= 0
pos = 45, 0

[State 803, spr]
type = sprpriority
trigger1 = 1
value = 1 -2 * (animelemtime(2) >= 0 && animelemtime(5) < 0)

[State 803, End]
type = changestate
trigger1 = !animtime
value = ifelse(var(10) = 1 || var(10) = 2, 805, ifelse(var(10) = 3, 810, 815))
ctrl = 0

;--- TARGET ---
[Statedef 802]
type = A
movetype = H
physics = N
velset = 0, 0
sprpriority = 0
[State 802, Anim]
type = changeanim2
trigger1 = !time
value = 802
[State Safety]
type = selfstate
trigger1 = time >= 60
value = ifelse(pos y < 0, 5040, 0)
ctrl = 1


[Statedef 804]
type = A
movetype = H
physics = N
velset = 0, 0
sprpriority = 0
[State 804, Anim]
type = changeanim2
trigger1 = !time
value = 804
[State Safety]
type = selfstate
trigger1 = time >= 60
value = ifelse(pos y < 0, 5040, 0)
ctrl = 1


;========< SEOI NAGE >==========
;--- PLAYER ---
[Statedef 805]
type = S
movetype = A
physics = S
anim = 805
ctrl = 0

[State 805, throw]
type = targetstate
trigger1 = !time
value = ifelse(var(10) = 2, 806, 808)

[State Tech Hit]
type = changestate
trigger1 = numtarget && var(19)
trigger1 = (target, command = "holdfwd" || target, command = "holdback") && target, command != "holdup" && target, command != "holddown"
trigger1 = (target, command = "b" || target, command = "c" || target, command = "y" || target, command = "z")
value = 890
[State Tech Hit]
type = varadd
trigger1 = var(19) > 0
var(19) = -1

[State 805, width]
type = width
trigger1 = animelemtime(4) < 0
edge = 0, 40
[State 805, width]
type = width
trigger1 = animelemtime(4) >= 0 && animelemtime(5) < 0
edge = 0, 70

[State 805, throwsnd]
type = playsnd
trigger1 = animelem = 2
value = 3, 3
channel = 0

[State 805, Bind 1]
type = targetbind
trigger1 = animelemtime(1) >= 0 && animelemtime(2) < 0
pos = 15, -76
[State 805, Bind 2]
type = targetbind
trigger1 = animelemtime(2) >= 0 && animelemtime(3) < 0
pos = -11, -105
[State 805, Bind 3]
type = targetbind
trigger1 = animelemtime(3) >= 0 && animelemtime(4) < 0
pos = -40, -80
[State 805, Bind 4]
type = targetbind
trigger1 = animelem = 4 && var(10) = 1
pos = -70, -45

[State 805, Snd]
type = playsnd
trigger1 = var(10) = 1 && numtarget
trigger1 = target, stateno = 5110 && target, time = 1
value = 5, 4
volume = -100
persistent = 0
[State 805, Dust]
type = explod
trigger1 = var(10) = 1 && numtarget
trigger1 = target, stateno = 5110 && target, time = 1
anim = 1338
ID = 1338
sprpriority = -2
postype = p2
pos = 0, 0
vel = 0, 0
ownpal = 1
scale = 0.5, 0.5
[State 805, Shake]
type = envshake
trigger1 = var(10) = 1 && numtarget
trigger1 = target, stateno = 5110 && target, time = 1
time = 12
freq = 120
ampl = 6
persistent = 0

[State 805, hurt]
type = targetlifeadd
trigger1 = animelem = 4 && var(10) = 2
value = -ceil(118 * fvar(12))
[State 805, hurt]
type = targetlifeadd
trigger1 = var(10) = 1 && numtarget
trigger1 = target, stateno = 5110 && target, time = 1
value = -ceil(125 * fvar(12))

[State 805, hurt]
type = targetpoweradd
trigger1 = animelem = 4
value = 36
[State 805, hurt]
type = poweradd
trigger1 = animelem = 4 && !var(20)
value = 72

[State 805, throw]
type = targetstate
trigger1 = animelem = 4
value = ifelse(var(10) = 2, 807, 809)

[State 805, Pos]
type = posadd
trigger1 = !animtime
x = 3
[State 805, Turn]
type = turn
trigger1 = !animtime
[State 805, End]
type = changestate
trigger1 = !animtime
value = 0 
ctrl = 1

;--- TARGET ---
[Statedef 806]
type = A
movetype = H
physics = N
velset = 0, 0
sprpriority = 0
[State 806, Anim]
type = changeanim2
trigger1 = !time
value = 806
[State 806, Camera]
type = screenbound
trigger1 = 1
value = 1
movecamera = 0, 0

[Statedef 807]
type = A
movetype = H
physics = N
velset = 6, -5

[State 807, Accel]
type = gravity
trigger1 = 1

[State 807, Recover]
type = selfstate
trigger1 = alive
trigger1 = command = "recovery"
trigger1 = vel y > 0 && pos y >= const(movement.air.gethit.groundrecover.ground.threshold)
value = 5200

[State 807, hitground]
type = selfstate
trigger1 = vel y > 0
trigger1 = pos y >= (-const(size.ground.front) -const(size.ground.back))
value = 5100

[Statedef 808]
type = A
movetype = H
physics = N
velset = 0, 0
sprpriority = 0
[State 808, Anim]
type = changeanim2
trigger1 = !time
value = 808
[State 808, Camera]
type = screenbound
trigger1 = 1
value = 1
movecamera = 0, 0

[Statedef 809]
type = A
movetype = H
physics = N
velset = 2, 10
[State 809, Accel]
type = gravity
trigger1 = 1
[State 809, hitground]
type = selfstate
trigger1 = vel y > 0
trigger1 = pos y >= -9
value = 5110


;========< TOMOE NAGE >==========
;--- PLAYER ---
[Statedef 810]
type = S
movetype = A
physics = S
anim = 810
ctrl = 0

[State 810, throw]
type = targetstate
trigger1 = !time
value = 811
[State 810, throw]
type = targetfacing
trigger1 = time <= 1
value = -1

[State Tech Hit]
type = changestate
trigger1 = numtarget && var(19)
trigger1 = (target, command = "holdfwd" || target, command = "holdback") && target, command != "holdup" && target, command != "holddown"
trigger1 = (target, command = "b" || target, command = "c" || target, command = "y" || target, command = "z")
value = 890
[State Tech Hit]
type = varadd
trigger1 = var(19) > 0
var(19) = -1

[State 810, width]
type = width
trigger1 = animelemtime(2) < 0
edge = 0, 20
[State 810, width]
type = width
trigger1 = animelemtime(2) >= 0 && animelemtime(3) < 0
edge = 0, 50
[State 810, width]
type = width
trigger1 = animelemtime(4) >= 0 && animelemtime(5) < 0
edge = 0, 70

[State 810, throwsnd]
type = playsnd
trigger1 = animelem = 2
value = 3, 11
channel = 0

[State 810, Bind 1]
type = targetbind
trigger1 = animelemtime(1) >= 0 && animelemtime(2) < 0
pos = 18, 0
[State 810, Bind 2]
type = targetbind
trigger1 = animelemtime(2) >= 0 && animelemtime(3) < 0
pos = 12, -62
[State 810, Bind 3]
type = targetbind
trigger1 = animelemtime(3) >= 0 && animelemtime(4) < 0
pos = -7, -54
[State 810, Bind 4]
type = targetbind
trigger1 = animelem = 4
pos = -74, -57

[State 810, hurt]
type = targetlifeadd
trigger1 = animelem = 4
value = -ceil(118 * fvar(12))
[State 810, hurt]
type = targetpoweradd
trigger1 = animelem = 4
value = 36
[State 810, hurt]
type = poweradd
trigger1 = animelem = 4 && !var(20)
value = 72

[State 810, throw]
type = targetstate
trigger1 = animelem = 4
value = 812

[State 810, Pos]
type = posadd
trigger1 = animelem = 1 || animelem = 2 || animelem = 3 || animelem = 4 || animelem = 5
x = ifelse((animelem = 1), -17, ifelse((animelem = 2), -15, ifelse((animelem = 3), -5, ifelse((animelem = 4), -14, -3))))
[State 810, Pos]
type = posadd
trigger1 = !animtime
x = 4

[State 810, Turn]
type = turn
trigger1 = !animtime
[State 810, End]
type = changestate
trigger1 = !animtime
value = 0 
ctrl = 1

;--- TARGET ---
[Statedef 811]
type = A
movetype = H
physics = N
velset = 0, 0
sprpriority = 0
[State 811, Anim]
type = changeanim2
trigger1 = !time
value = 811
[State 811, Camera]
type = screenbound
trigger1 = 1
value = 1
movecamera = 0, 0

[Statedef 812]
type = A
movetype = H
physics = N
velset = 5, -9

[State 812, Accel]
type = gravity
trigger1 = 1

[State 812, Recover]
type = selfstate
trigger1 = alive
trigger1 = command = "recovery"
trigger1 = vel y > 0 && pos y >= const(movement.air.gethit.groundrecover.ground.threshold)
value = 5200

[State 812, hitground]
type = selfstate
trigger1 = vel y > 0
trigger1 = pos y >= 0
value = 5100


;========< JIGOKU GURUMA >==========
;--- PLAYER ---
[Statedef 815]
type = S
movetype = A
physics = C
anim = 815
velset = 0, 0
ctrl = 0

[State 815, throw]
type = targetstate
trigger1 = !time
value = 816
[State 815, throw]
type = targetfacing
trigger1 = time <= 1
value = -1

[State Tech Hit]
type = changestate
trigger1 = numtarget && var(19)
trigger1 = (target, command = "holdfwd" || target, command = "holdback") && target, command != "holdup" && target, command != "holddown"
trigger1 = (target, command = "b" || target, command = "c" || target, command = "y" || target, command = "z")
value = 890
[State Tech Hit]
type = varadd
trigger1 = var(19) > 0
var(19) = -1

[State 815, width]
type = width
trigger1 = animelemtime(11) >= 0 && animelemtime(12) < 0
edge = 0, 50
[State 815, width]
type = width
trigger1 = animelemtime(12) >= 0 && animelemtime(13) < 0
edge = 0, 70

[State 815, throwsnd]
type = playsnd
trigger1 = animelem = 12
value = 3, 3
channel = 0
[State 815, spinsnd]
type = playsnd
trigger1 = animelem = 4 || animelem = 8 || animelem = 12
value = 1, ifelse((animelem = 12), 6, 5)
channel = 1

[State 815, Bind 1]
type = targetbind
trigger1 = animelemtime(1) >= 0 && animelemtime(2) < 0
pos = 18, 0
[State 815, Bind 2]
type = targetbind
trigger1 = animelemtime(2) >= 0 && animelemtime(3) < 0
pos = 12, -62
[State 815, Bind 3]
type = targetbind
trigger1 = animelemtime(3) >= 0 && animelemtime(4) < 0
trigger2 = animelemtime(7) >= 0 && animelemtime(8) < 0
pos = 2, -55
[State 815, Bind 4]
type = targetbind
trigger1 = animelemtime(4) >= 0 && animelemtime(5) < 0
trigger2 = animelemtime(8) >= 0 && animelemtime(9) < 0
pos = -14, -36
[State 815, Bind 5]
type = targetbind
trigger1 = animelemtime(5) >= 0 && animelemtime(6) < 0
trigger2 = animelemtime(9) >= 0 && animelemtime(10) < 0
pos = 9, 0
[State 815, Bind 6]
type = targetbind
trigger1 = animelemtime(6) >= 0 && animelemtime(7) < 0
trigger2 = animelemtime(10) >= 0 && animelemtime(11) < 0
pos = 20, -52
[State 815, Bind 11]
type = targetbind
trigger1 = animelemtime(11) >= 0 && animelemtime(12) < 0
pos = -7, -54
[State 815, Bind 12]
type = targetbind
trigger1 = animelem = 12
pos = -74, -57

[State 815, hurt]
type = targetlifeadd
trigger1 = animelem = 12
value = -ceil(139 * fvar(12))
[State 815, hurt]
type = targetpoweradd
trigger1 = animelem = 12
value = 36
[State 815, hurt]
type = poweradd
trigger1 = animelem = 12 && !var(20)
value = 72

[State 815, throw]
type = targetstate
trigger1 = animelem = 12
value = 817

[State 815, Pos]
type = posset
trigger1 = 1
y = 0
[State 815, Vel]
type = velset
trigger1 = animelemtime(3) >= 0 && animelemtime(10) < 0
x = -3

[State 815, Pos]
type = posadd
trigger1 = animelem = 1 || animelem = 2 || animelem = 3
x = ifelse((animelem = 1), -17, ifelse((animelem = 2), -15, -24))

[State 815, Dust]
type = explod
trigger1 = animelemtime(3) >= 0 && animelemtime(10) < 0 && !(animelemtime(3) % 5)
anim = 7103
ID = 7103
sprpriority = 3
postype = p1
vel = 1, 0
ownpal = 1
pausemovetime = -1
supermovetime = -1

[State 815, Turn]
type = turn
trigger1 = !animtime
[State 815, End]
type = changestate
trigger1 = !animtime
value = 0 
ctrl = 1

;--- TARGET ---
[Statedef 816]
type = A
movetype = H
physics = N
velset = 0, 0
sprpriority = 0
[State 816, Anim]
type = changeanim2
trigger1 = !time
value = 816
[State 816, Camera]
type = screenbound
trigger1 = 1
value = 1
movecamera = 0, 0

[Statedef 817]
type = A
movetype = H
physics = N
velset = 5, -9

[State 817, Accel]
type = gravity
trigger1 = 1

[State 817, Recover]
type = selfstate
trigger1 = alive
trigger1 = command = "recovery"
trigger1 = vel y > 0 && pos y >= const(movement.air.gethit.groundrecover.ground.threshold)
value = 5200

[State 817, hitground]
type = selfstate
trigger1 = vel y > 0
trigger1 = pos y >= 0
value = 5100


;========< JIGOKU FUUSHA >==========
[Statedef 850]
type = A
movetype = A
physics = A
juggle = 8
ctrl = 0
anim = 850
sprpriority = 1
poweradd = 18 * !var(20)

[State 850, Do not guard]
type = attackdist
trigger1 = 1
value = 0

[State 850, Hit]
type = hitdef
trigger1 = !time
trigger1 = !var(38)
attr = A, NT 
hitflag = AF
priority = 1, Miss
sparkno = -1
hitsound = s0, 800
p1facing = 1
p1stateno = 860 
guard.dist = 0
fall = 1
fall.recover = 1
fall.recovertime = 120

[State 850, misssnd]
type = playsnd
trigger1 = animelem = 3
value = 5000, 3
channel = 0

[State Idle]
type = statetypeset
trigger1 = animelemtime(3) >= 0
movetype = I


;--- PLAYER ---
[Statedef 860]
type = A
movetype = A
physics = N
anim = 860
velset = ifelse(!vel x, -1.2, vel x), vel y -4
ctrl = 0

[State 860, Turn]
type = turn
trigger1 = !time && vel x > 0

[State 860, throw]
type = targetstate
trigger1 = !time
value = 865
[State 860, throw]
type = targetfacing
trigger1 = time <= 1
value = -1

[State 860, width]
type = width
trigger1 = animelemtime(7) >= 0 && animtime
edge = 0, 55

[State 860, width]
type = width
trigger1 = animelemtime(7) >= 0 && animelemtime(8) < 0
edge = 0, 50
[State 860, width]
type = width
trigger1 = animelemtime(8) >= 0 && animelemtime(9) < 0
edge = 0, 70

[State 860, Accel]
type = gravity
trigger1 = anim = 860

[State 860, throwsnd]
type = playsnd
trigger1 = animelem = 2
value = 3, 5
channel = 0
[State 860, throwsnd]
type = playsnd
trigger1 = animelem = 8
value = 3, 11
channel = 0
[State 860, spinsnd]
type = playsnd
trigger1 = animelem = 4 || animelem = 8
value = 1, ifelse((animelem = 8), 6, 5)
channel = 1

[State 860, Vel]
type = changeanim
trigger1 = pos y < -5 && animelemtime(7) >= 0
value = 860
elem = 3
[State 815, throw]
type = targetstate
trigger1 = animelem = 3
value = 866

[State 865, Bind 1]
type = targetbind
trigger1 = animelemtime(1) >= 0 && animelemtime(2) < 0
pos = 18, 0
[State 865, Bind 2]
type = targetbind
trigger1 = animelemtime(2) >= 0 && animelemtime(3) < 0
pos = 12, -62
[State 865, Bind 3]
type = targetbind
trigger1 = animelemtime(3) >= 0 && animelemtime(4) < 0
pos = 2, -55
[State 865, Bind 4]
type = targetbind
trigger1 = animelemtime(4) >= 0 && animelemtime(5) < 0
pos = -14, -36
[State 865, Bind 5]
type = targetbind
trigger1 = animelemtime(5) >= 0 && animelemtime(6) < 0
pos = 9, 0
[State 865, Bind 6]
type = targetbind
trigger1 = animelemtime(6) >= 0 && animelemtime(7) < 0
pos = 20, -52
[State 865, Bind 7]
type = targetbind
trigger1 = animelemtime(7) >= 0 && animelemtime(8) < 0
pos = -7, -54
[State 865, Bind 8]
type = targetbind
trigger1 = animelem = 8
pos = -74, -57

[State 865, hurt]
type = targetlifeadd
trigger1 = animelem = 8
value = -ceil(139 * fvar(12))
[State 865, hurt]
type = targetpoweradd
trigger1 = animelem = 8
value = 36
[State 865, hurt]
type = poweradd
trigger1 = animelem = 8 && !var(20)
value = 72

[State 865, throw]
type = targetstate
trigger1 = animelem = 8
value = 867

[State 860, Pos]
type = posset
trigger1 = vel y > 0 && pos y >= -5
y = 0
[State 860, Vel]
type = velset
trigger1 = pos y >= -5
y = 0
[State 860, Vel]
type = velset
trigger1 = !pos y && animelemtime(7) < 0
x = -4
[State 860, statetype]
type = statetypeset
trigger1 = pos y >= -5
statetype = C
physics = C

[State 860, Snd]
type = playsnd
trigger1 = !pos y
value = 5, 17
persistent = 0
[State 860, Dust]
type = explod
trigger1 = !pos y
anim = 7102
ID = 7102
sprpriority = -2
postype = p1
scale = 0.75, 0.75
ownpal = 1
persistent = 0
pausemovetime = -1
supermovetime = -1

[State 860, Dust]
type = explod
trigger1 = !pos y && animelemtime(7) < 0 && !(animelemtime(1) % 5)
anim = 7103
ID = 7103
sprpriority = 3
postype = p1
vel = 1, 0
ownpal = 1
pausemovetime = -1
supermovetime = -1

[State 860, Turn]
type = turn
trigger1 = !animtime
[State 860, End]
type = changestate
trigger1 = !animtime
value = 0 
ctrl = 1


;--- TARGET ---
[Statedef 865]
type = A
movetype = H
physics = N
velset = 0, 0
sprpriority = 0
[State 865, Anim]
type = changeanim2
trigger1 = !time
value = 865
[State 865, Camera]
type = screenbound
trigger1 = 1
value = 1
movecamera = 0, 0

[Statedef 866]
type = A
movetype = H
physics = N
velset = 0, 0
sprpriority = 0
[State 866, Anim]
type = changeanim2
trigger1 = !time
value = 866
[State 866, Camera]
type = screenbound
trigger1 = 1
value = 1
movecamera = 0, 0

[Statedef 867]
type = A
movetype = H
physics = N
velset = 6, -9

[State 867, Accel]
type = gravity
trigger1 = 1

[State 867, Recover]
type = selfstate
trigger1 = alive
trigger1 = command = "recovery"
trigger1 = vel y > 0 && pos y >= const(movement.air.gethit.groundrecover.ground.threshold)
value = 5200

[State 867, hitground]
type = selfstate
trigger1 = vel y > 0
trigger1 = pos y >= 0
value = 5100


;===========< Target Escaped >==========
[Statedef 890]
type = S
physics = N
movetype = I
anim = 5945
ctrl = 0
velset = 0, 0
sprpriority = 1

[State 890, Bind]
type = targetbind
trigger1 = !time && numtarget
pos = 0, 0
[State 890, p2state]
type = targetstate
trigger1 = !time && numtarget
value = 895
[State 890, Throw]
type = targetfacing
trigger1 = !time && numtarget
value = -1

[State 890, Pos]
type = posadd
trigger1 = !time
x = 24

[State 890, NHB]
type = nothitby
trigger1 = 1
value = SCA
time = 1
[State 890, Push]
type = playerpush
trigger1 = 1
value = 0

[State 890, Pos]
type = posset
trigger1 = !time
y = 0
[State 890, Vel]
type = velset
trigger1 = 1
x = -1.425
y = 0

[State 890, Snd]
type = playsnd
trigger1 = time = 2
value = 5000, 2
channel = 0

[State 890, FX]
type = playsnd
trigger1 = !time
value = 7, 3

[State 890, FX]
type = explod
trigger1 = !time
anim = 7000
ID = 7000
sprpriority = 5
postype = p1
pos = 0, -55
ownpal = 1
pausemovetime = -1
supermovetime = -1

[State 890, End]
type = changestate
trigger1 = time >= 27
value = 0
ctrl = 1


;--- TARGET ---
[Statedef 895]
type = S
physics = N
movetype = I
ctrl = 0
velset = 0, 0
sprpriority = 1

[State 895, Anim]
type = changeanim
trigger1 = !time
value = ifelse(selfanimexist(5945), 5945, 130)
[State 895, Anim]
type = changeanim
trigger1 = !time && !selfanimexist(5945) && selfanimexist(5910)
value = 5910
[State 895, Anim]
type = changeanim
trigger1 = time >= 20 && anim = 130
value = 140

[State 895, NHB]
type = nothitby
trigger1 = 1
value = SCA
time = 1
[State 895, Push]
type = playerpush
trigger1 = 1
value = 0

[State 895, Pos]
type = posset
trigger1 = !time
y = 0
[State 895, Vel]
type = velset
trigger1 = 1
x = -1.425
y = 0

[State 895, End]
type = selfstate
trigger1 = time >= 27
value = 0
ctrl = 1

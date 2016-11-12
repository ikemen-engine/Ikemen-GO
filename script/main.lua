snd=sndNew('data/test.snd')
playBGM('sound/test.ogg')
sndPlay(snd,2,0)
sndPlay(snd,0,1)
sndPlay(snd,1,0)
sff=sffNew('data/test.sff')
sffv2=sffNew('data/testv2.sff')
anim=animNew(sffv2,[[
Clsn2Default: 2
 Clsn2[0] = -52,0,64,-316
 Clsn2[1] =   20,-316,-28,-372
0,0,0,0, 10
0,1,0,0, 7
0,2,0,0, 7
0,3,0,0, 7
0,4,0,0, 7
0,5,0,0, 45
0,4,0,0, 7
0,3,0,0, 7
0,2,0,0, 7
0,1,0,0, 7
0,0,0,0, 40
]])
animSetPos(anim,80,60)
animAddPos(anim,80,60)
animSetTile(anim,1,1)
animSetColorKey(anim,-1)
animSetAlpha(anim,128,128)
animSetScale(anim,0.5,0.5)
animSetWindow(anim,10,10,300,220)
while true do
  animUpdate(anim)
  animDraw(anim)
  refresh()
end

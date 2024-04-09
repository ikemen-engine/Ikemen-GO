package main

import "math"

type stageCamera struct {
	startx               int32
	starty               int32
	boundleft            int32
	boundright           int32
	boundhigh            int32
	boundlow             int32
	verticalfollow       float32
	floortension         int32
	tensionhigh          int32
	tensionlow           int32
	tension              int32
	tensionvel           float32
	overdrawhigh         int32 //TODO: not implemented
	overdrawlow          int32
	cuthigh              int32
	cutlow               int32
	localcoord           [2]int32
	localscl             float32
	zoffset              int32
	ztopscale            float32
	startzoom            float32
	zoomin               float32
	zoomout              float32
	ytensionenable       bool
	autocenter           bool
	zoomanchor           bool
	zoomindelay          float32
	zoomindelaytime      float32
	fov                  float32
	yshift               float32
	far                  float32
	near                 float32
	aspectcorrection     float32
	zoomanchorcorrection float32
	ywithoutbound        float32
	highest              float32
	lowest               float32
	leftest              float32
	rightest             float32
	leftestvel           float32
	rightestvel          float32
	roundstart           bool
	maxRight             float32
	minLeft              float32
}

func newStageCamera() *stageCamera {
	return &stageCamera{verticalfollow: 0.2, tensionvel: 1, tension: 50,
		cuthigh: 0, cutlow: 0,
		localcoord: [...]int32{320, 240}, localscl: float32(sys.gameWidth / 320),
		ztopscale: 1, startzoom: 1, zoomin: 1, zoomout: 1, ytensionenable: false, fov: 40, yshift: 0, far: 10000, near: 0.1, zoomindelay: 0}
}

type CameraView int

const (
	Fighting_View CameraView = iota
	Follow_View
	Free_View
)

type Camera struct {
	stageCamera
	View                            CameraView
	ZoomEnable, ZoomActive          bool
	ZoomDelayEnable                 bool
	ZoomMin, ZoomMax, ZoomSpeed     float32
	zoomdelay                       float32
	Pos, ScreenPos, Offset          [2]float32
	XMin, XMax                      float32
	Scale, MinScale                 float32
	boundL, boundR, boundH, boundLo float32
	zoff                            float32
	halfWidth                       float32
	FollowChar                      *Char
}

func newCamera() *Camera {
	return &Camera{View: Fighting_View, ZoomMin: 5.0 / 6, ZoomMax: 15.0 / 14, ZoomSpeed: 12}
}
func (c *Camera) Reset() {
	c.ZoomEnable = c.ZoomActive && (c.stageCamera.zoomin != 1 || c.stageCamera.zoomout != 1)
	c.boundL = float32(c.boundleft-c.startx)*c.localscl - ((1-c.zoomout)*100*c.zoomout)*(1/c.zoomout)*(1/c.zoomout)*1.6*(float32(sys.gameWidth)/320)
	c.boundR = float32(c.boundright-c.startx)*c.localscl + ((1-c.zoomout)*100*c.zoomout)*(1/c.zoomout)*(1/c.zoomout)*1.6*(float32(sys.gameWidth)/320)
	c.halfWidth = float32(sys.gameWidth) / 2
	c.XMin = c.boundL - c.halfWidth/c.BaseScale()
	c.XMax = c.boundR + c.halfWidth/c.BaseScale()
	c.aspectcorrection = 0
	c.zoomanchorcorrection = 0
	if float32(c.localcoord[1])*c.localscl-float32(sys.gameHeight) < 0 {
		c.aspectcorrection = MinF(0, (float32(c.localcoord[1])*c.localscl-float32(sys.gameHeight))+MinF((float32(sys.gameHeight)-float32(c.localcoord[1])*c.localscl)/2, float32(c.overdrawlow)*c.localscl))
	} else if float32(c.localcoord[1])*c.localscl-float32(sys.gameHeight) > 0 {
		if c.cuthigh+c.cutlow <= 0 {
			c.aspectcorrection = float32(Ceil(float32(c.localcoord[1])*c.localscl) - sys.gameHeight)
		} else {
			diff := Ceil(float32(c.localcoord[1])*c.localscl) - sys.gameHeight
			tmp := Ceil(float32(c.cuthigh)*c.localscl) * diff / (Ceil(float32(c.cuthigh)*c.localscl) + Ceil(float32(c.cutlow)*c.localscl))
			if diff-tmp <= c.cutlow {
				c.aspectcorrection = float32(tmp)
			} else {
				c.aspectcorrection = float32(diff - Ceil(float32(c.cutlow)*c.localscl))
			}
		}

	}
	c.boundH = float32(c.boundhigh) * c.localscl
	c.boundLo = float32(Max(c.boundhigh, c.boundlow)) * c.localscl
	c.boundlow = Max(c.boundhigh, c.boundlow)

	xminscl := float32(sys.gameWidth) / (float32(sys.gameWidth) - c.boundL +
		c.boundR)
	//yminscl := float32(sys.gameHeight) / (240 - MinF(0, c.boundH))
	c.MinScale = MaxF(c.zoomout, MinF(c.zoomin, xminscl))
	c.maxRight = float32(c.boundright)*c.localscl + c.halfWidth/c.zoomout
	c.minLeft = float32(c.boundleft)*c.localscl - c.halfWidth/c.zoomout
}
func (c *Camera) Init() {
	c.Reset()
	c.View = Fighting_View
	c.roundstart = true
	c.Scale = c.startzoom
	c.Pos[0], c.Pos[1], c.ywithoutbound = float32(c.startx)*c.localscl, float32(c.starty)*c.localscl, float32(c.starty)*c.localscl
	c.zoomindelaytime = c.zoomindelay
}
func (c *Camera) ResetTracking() {
	c.leftest = c.Pos[0]
	c.rightest = c.Pos[0]
	c.highest = math.MaxFloat32
	c.lowest = -math.MaxFloat32
	c.leftestvel = 0
	c.rightestvel = 0
}
func (c *Camera) Update(scl, x, y float32) {
	c.Scale = c.BaseScale() * scl
	c.zoff = float32(c.zoffset) * c.localscl
	if sys.stage.stageCamera.zoomanchor {
		c.zoomanchorcorrection = c.zoff - (float32(sys.gameHeight) + c.aspectcorrection - (float32(sys.gameHeight)-c.zoff+c.aspectcorrection)*scl)
	}
	for i := 0; i < 2; i++ {
		c.Offset[i] = sys.stage.bga.offset[i] * sys.stage.localscl * scl
	}
	c.ScreenPos[0] = x - c.halfWidth/c.Scale - c.Offset[0]
	c.ScreenPos[1] = y - (c.GroundLevel()-float32(sys.gameHeight-240)*scl)/
		c.Scale - c.Offset[1]
	c.Pos[0] = x
	c.Pos[1] = y
}
func (c *Camera) ScaleBound(scl, sclmul float32) float32 {
	if c.ZoomEnable {
		if sys.debugPaused() {
			sclmul = 1
		} else if sys.turbo < 1 {
			sclmul = Pow(sclmul, sys.turbo)
		}
		return MaxF(c.MinScale, MinF(c.zoomin, scl*sclmul))
	}
	return 1
}
func (c *Camera) XBound(scl, x float32) float32 {
	return ClampF(x,
		c.boundL-c.halfWidth+c.halfWidth/scl,
		c.boundR+c.halfWidth-c.halfWidth/scl)
}
func (c *Camera) BaseScale() float32 {
	return c.ztopscale
}
func (c *Camera) GroundLevel() float32 {
	return c.zoff - c.aspectcorrection - c.zoomanchorcorrection
}
func (c *Camera) ResetZoomdelay() {
	c.zoomdelay = 0
}
func (c *Camera) action(x, y, scale float32, pause bool) (newX, newY, newScale float32) {
	newX = x
	newY = y
	newScale = scale
	if !sys.debugPaused() {
		switch c.View {
		case Fighting_View:
			if c.highest != math.MaxFloat32 && c.lowest != -math.MaxFloat32 {
				tension := MaxF(0, float32(c.tension)*c.localscl)
				oldLeft, oldRight := x-c.halfWidth/scale, x+c.halfWidth/scale
				targetLeft, targetRight := oldLeft, oldRight
				if c.autocenter {
					targetLeft = MinF(MaxF((c.leftest+c.rightest)/2-c.halfWidth/scale, c.minLeft), c.maxRight-2*c.halfWidth/scale)
					targetRight = targetLeft + 2*c.halfWidth/scale
				}

				if c.leftest < targetLeft+tension {
					diff := targetLeft - MaxF(c.leftest-tension, c.minLeft)
					targetLeft = MaxF(c.leftest-tension, c.minLeft)
					targetRight = MaxF(oldRight-diff, MinF(c.rightest+tension, c.maxRight))
				} else if c.rightest > targetRight-tension {
					diff := targetRight - MinF(c.rightest+tension, c.maxRight)
					targetRight = MinF(c.rightest+tension, c.maxRight)
					targetLeft = MinF(oldLeft-diff, MaxF(c.leftest-tension, c.minLeft))
				}
				if c.halfWidth*2/(targetRight-targetLeft) < c.zoomout {
					x := (targetRight + targetLeft) / 2
					targetLeft = x - c.halfWidth/c.zoomout
					targetRight = x + c.halfWidth/c.zoomout
					if c.leftest-targetLeft < float32(sys.stage.screenleft)*c.localscl {
						diff := MinF(float32(sys.stage.screenleft)*c.localscl-(c.leftest-targetLeft), targetLeft-c.minLeft)
						if targetRight-c.rightest < float32(sys.stage.screenright)*c.localscl {
							diff = diff + (MinF(float32(sys.stage.screenright)*c.localscl-(targetRight-c.rightest), c.maxRight-targetRight)-diff)/2
						}
						targetLeft -= diff
						targetRight -= diff
					} else if targetRight-c.rightest < float32(sys.stage.screenright)*c.localscl {
						diff := MinF(float32(sys.stage.screenright)*c.localscl-(targetRight-c.rightest), c.maxRight-targetRight)
						targetLeft += diff
						targetRight += diff
					}
				}
				maxScale := c.zoomin
				if c.ytensionenable {
					maxScale = MinF(MaxF(float32(sys.gameHeight)/((c.lowest+float32(c.tensionlow)*c.localscl)-(c.highest-float32(c.tensionhigh)*c.localscl)), c.zoomout), maxScale)
				}
				if c.halfWidth*2/(targetRight-targetLeft) < maxScale {
					if c.zoomindelaytime > 0 {
						c.zoomindelaytime -= 1
					} else {
						diffLeft := MaxF(c.leftest-tension-targetLeft, 0)
						if diffLeft < 0 {
							diffLeft = 0
						}
						diffRight := MinF(c.rightest+tension-targetRight, 0)
						if diffRight > 0 {
							diffRight = 0
						}
						if c.halfWidth*2/((targetRight+diffRight)-(targetLeft+diffLeft)) > maxScale {
							tmp, tmp2 := diffLeft/(diffLeft-diffRight)*((targetRight+diffRight)-(targetLeft+diffLeft)-c.halfWidth*2/maxScale), diffRight/(diffLeft-diffRight)*((targetRight+diffRight)-(targetLeft+diffLeft)-c.halfWidth*2/maxScale)
							diffLeft += tmp
							diffRight += tmp2
						}
						targetLeft += diffLeft
						targetRight += diffRight
					}
				} else {
					c.zoomindelaytime = c.zoomindelay
				}

				targetX := (targetLeft + targetRight) / 2
				targetScale := c.halfWidth * 2 / (targetRight - targetLeft)

				if !c.ytensionenable {
					newY = c.ywithoutbound
					//old*0.85+target* 0.15 if diff > 1
					targetY := (c.highest + float32(c.floortension)*c.localscl) * c.verticalfollow
					if !c.roundstart {
						for i := 0; i < 3; i++ {
							newY = newY*.85 + targetY*.15
							if AbsF(targetY-newY) < 1 {
								newY = targetY
								break
							}
						}
					}
					c.ywithoutbound = newY
				} else {
					targetScale = MinF(MinF(MaxF(float32(sys.gameHeight)/((c.lowest+float32(c.tensionlow)*c.localscl)-(c.highest-float32(c.tensionhigh)*c.localscl)), c.zoomout), c.zoomin), targetScale)
					targetX = MinF(MaxF(targetX, float32(c.boundleft)*c.localscl-c.halfWidth*(1/c.zoomout-1/targetScale)), float32(c.boundright)*c.localscl+c.halfWidth*(1/c.zoomout-1/targetScale))
					targetLeft = targetX - c.halfWidth/targetScale
					targetRight = targetX + c.halfWidth/targetScale

					newY = c.ywithoutbound
					targetY := c.GroundLevel()/targetScale + (c.highest - float32(c.tensionhigh)*c.localscl)
					if !c.roundstart {
						for i := 0; i < 3; i++ {
							newY = newY*.85 + targetY*.15
							if AbsF(targetY-newY) < 1 {
								newY = targetY
								break
							}
						}
					} else {
						newY = targetY
					}
					c.ywithoutbound = newY

				}

				newLeft, newRight := oldLeft, oldRight
				if !c.roundstart {
					for i := 0; i < 3; i++ {
						newLeft, newRight = newLeft+(targetLeft-newLeft)*0.05*sys.turbo*c.tensionvel, newRight+(targetRight-newRight)*0.05*sys.turbo*c.tensionvel
						diffLeft := targetLeft - newLeft
						diffRight := targetRight - newRight
						if AbsF(diffLeft) <= 0.1*sys.turbo {
							newLeft = targetLeft
						} else if diffLeft > 0 {
							newLeft += 0.1 * sys.turbo * c.tensionvel
						} else {
							newLeft -= 0.1 * sys.turbo * c.tensionvel
						}
						if newLeft-oldLeft > 0 && newLeft-oldLeft < c.rightestvel {
							newLeft = MinF(oldLeft+c.rightestvel, targetLeft)
						} else if newLeft-oldLeft < 0 && newLeft-oldLeft > c.leftestvel {
							newLeft = MaxF(oldLeft+c.leftestvel, targetLeft)
						}

						if AbsF(diffRight) <= 0.1*sys.turbo*c.tensionvel {
							newRight = targetRight
						} else if diffRight > 0 {
							newRight += 0.1 * sys.turbo * c.tensionvel
						} else {
							newRight -= 0.1 * sys.turbo * c.tensionvel
						}
						if newRight-oldRight > 0 && newRight-oldRight < c.rightestvel {
							newRight = MinF(oldRight+c.rightestvel, targetRight)
						} else if newRight-oldRight < 0 && newRight-oldRight > c.leftestvel {
							newRight = MaxF(oldRight+c.leftestvel, targetRight)
						}

						newX = (newLeft + newRight) / 2
					}
				} else {
					newLeft, newRight = targetLeft, targetRight
					newX = (newLeft + newRight) / 2
				}
				newScale = c.halfWidth * 2 / (newRight - newLeft)
				newY = MinF(MaxF(newY, float32(c.boundhigh)*c.localscl*newScale), float32(c.boundlow)*c.localscl*newScale)
			} else {
				newScale = MinF(MaxF(newScale, c.zoomout), c.zoomin)
				newX = MinF(MaxF(newX, float32(c.boundleft)*c.localscl*newScale), float32(c.boundright)*c.localscl*newScale)
				newY = MinF(MaxF(newY, float32(c.boundhigh)*c.localscl*newScale), float32(c.boundlow)*c.localscl*newScale)
			}

		case Follow_View:
			newX = c.FollowChar.pos[0]
			newY = c.FollowChar.pos[1] * Pow(c.verticalfollow, MinF(1, 1/Pow(c.Scale, 4)))
			newScale = 1
		case Free_View:
			newX = c.Pos[0]
			newY = c.Pos[1]
			c.ywithoutbound = newY
			newScale = 1
		}
	}
	c.roundstart = false
	return
}

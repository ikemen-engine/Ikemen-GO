// Package main contains camera logic for Ikemen GO.
//
// The camera is responsible for framing the fight scene every frame. It
// reacts to player positions and stage configuration to adjust scroll, zoom and
// boundaries. The update loop is executed once per game tick, keeping the
// action centered while respecting stage limits and player tracking. Many of
// the behaviors mirror those found in the original M.U.G.E.N engine.
package main

import "math"

type stageCamera struct {
	startx                  int32
	starty                  int32
	boundleft               int32
	boundright              int32
	boundhigh               int32
	boundlow                int32
	verticalfollow          float32
	floortension            int32
	tensionhigh             int32
	tensionlow              int32
	lowestcap               bool
	tension                 int32
	tensionvel              float32
	overdrawhigh            int32 // TODO: not implemented
	overdrawlow             int32
	cuthigh                 int32
	cutlow                  int32
	localcoord              [2]int32
	localscl                float32
	zoffset                 int32
	ztopscale               float32
	zbotscale               float32
	depthtoscreen           float32
	topz                    float32
	botz                    float32
	startzoom               float32
	zoomin                  float32
	zoomout                 float32
	ytensionenable          bool
	autocenter              bool
	zoomanchor              bool
	boundhighzoomdelta      float32
	verticalfollowzoomdelta float32
	zoomindelay             float32
	zoomindelaytime         float32
	zoominspeed             float32
	zoomoutspeed            float32
	yscrollspeed            float32
	fov                     float32
	yshift                  float32
	far                     float32
	near                    float32
	aspectcorrection        float32
	zoomanchorcorrection    float32
	ywithoutbound           float32
	highest                 float32
	prevHighest             float32
	lowest                  float32
	prevLowest              float32
	leftest                 float32
	prevLeftest             float32
	rightest                float32
	prevRightest            float32
	leftestvel              float32
	rightestvel             float32
	roundstart              bool
	maxRight                float32
	minLeft                 float32
}

func newStageCamera() *stageCamera {
	return &stageCamera{verticalfollow: 0.2, tensionvel: 1, tension: 50,
		cuthigh: 0, cutlow: math.MinInt32,
		localcoord: [...]int32{320, 240}, localscl: float32(sys.gameWidth / 320),
		topz: 0, botz: 0, ztopscale: 1, zbotscale: 1, depthtoscreen: 1,
		startzoom: 1, zoomin: 1, zoomout: 1,
		ytensionenable: false, tensionhigh: 0, tensionlow: 0,
		fov: 40, yshift: 0, far: 10000, near: 0.1,
		zoomindelay: 0, zoominspeed: 1, zoomoutspeed: 1, yscrollspeed: 1,
		boundhighzoomdelta: 0, verticalfollowzoomdelta: 0}
}

// CameraView describes the active camera mode used by the engine.
//
// Fighting_View is the default behaviour tracking both players. Follow_View
// locks the camera to a specific character, while Free_View allows manual
// control by debug tooling.
type CameraView int

const (
	Fighting_View CameraView = iota // camera tracks all fighters
	Follow_View                     // camera follows a single character
	Free_View                       // camera position is manually controlled
)

// Camera exposes the run‑time camera state used for rendering.
//
// All positions are expressed in stage units (equivalent to pixels at the stage
// local coordinate system). Scale values are multipliers where 1.0 means no
// zoom. The struct embeds stageCamera which holds stage configuration values.
type Camera struct {
	stageCamera
	View                            CameraView
	ZoomEnable                      bool
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
	return &Camera{View: Fighting_View}
}

// Reset recomputes static bounds and zoom parameters based on stage settings.
// It should be called when the stage or configuration changes. All distances
// are in stage pixels.
func (c *Camera) Reset() {
	c.ZoomEnable = sys.cfg.Config.ZoomActive && (c.stageCamera.zoomin != 1 || c.stageCamera.zoomout != 1)
	c.boundL = float32(c.boundleft-c.startx)*c.localscl - ((1-c.zoomout)*100*c.zoomout)*(1/c.zoomout)*(1/c.zoomout)*1.6*(float32(sys.gameWidth)/320)
	c.boundR = float32(c.boundright-c.startx)*c.localscl + ((1-c.zoomout)*100*c.zoomout)*(1/c.zoomout)*(1/c.zoomout)*1.6*(float32(sys.gameWidth)/320)
	c.halfWidth = float32(sys.gameWidth) / 2
	c.XMin = c.boundL - c.halfWidth/c.BaseScale()
	c.XMax = c.boundR + c.halfWidth/c.BaseScale()
	c.aspectcorrection = 0
	c.zoomanchorcorrection = 0
	c.zoomin = MaxF(c.zoomin, c.zoomout)
	if c.cutlow == math.MinInt32 {
		c.cutlow = int32(float32(c.localcoord[1]-c.zoffset) - float32(c.localcoord[1])*0.05)
	}
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
	c.tensionvel = MaxF(MinF(c.tensionvel, 20), 0)
	if c.verticalfollow < 0 {
		c.ytensionenable = true
	}
	xminscl := float32(sys.gameWidth) / (float32(sys.gameWidth) - c.boundL +
		c.boundR)
	//yminscl := float32(sys.gameHeight) / (240 - MinF(0, c.boundH))
	c.MinScale = MaxF(c.zoomout, MinF(c.zoomin, xminscl))
	c.maxRight = float32(c.boundright)*c.localscl + c.halfWidth/c.zoomout
	c.minLeft = float32(c.boundleft)*c.localscl - c.halfWidth/c.zoomout
}

// Init prepares the camera for a new round.
// It resets internal state and positions to stage starting coordinates.
func (c *Camera) Init() {
	c.Reset()
	c.View = Fighting_View
	c.roundstart = true
	c.Scale = c.startzoom
	c.Pos[0], c.Pos[1], c.ywithoutbound = float32(c.startx)*c.localscl, float32(c.starty)*c.localscl, float32(c.starty)*c.localscl
	c.zoomindelaytime = c.zoomindelay
}

// ResetTracking clears extreme position tracking for a new frame cycle.
func (c *Camera) ResetTracking() {
	c.leftest = math.MaxFloat32
	c.rightest = -math.MaxFloat32
	c.highest = math.MaxFloat32
	c.lowest = -math.MaxFloat32
	c.leftestvel = 0
	c.rightestvel = 0
}

// SaveRestoreTracking preserves the last valid extremes when no character
// updated them in the current frame. M.U.G.E.N simply stops camera movement in
// those cases; this implementation keeps previous positions for smoother motion.
func (c *Camera) SaveRestoreTracking() {
	if c.highest == math.MaxFloat32 {
		c.highest = c.prevHighest
	} else {
		c.prevHighest = c.highest
	}

	if c.lowest == -math.MaxFloat32 {
		c.lowest = c.prevLowest
	} else {
		c.prevLowest = c.lowest
	}

	if c.leftest == math.MaxFloat32 {
		c.leftest = c.prevLeftest
	} else {
		c.prevLeftest = c.leftest
	}

	if c.rightest == -math.MaxFloat32 {
		c.rightest = c.prevRightest
	} else {
		c.prevRightest = c.rightest
	}
}

// Update finalizes the camera for the current frame.
//
// scl is the scaling multiplier from action(), and x/y are the target camera
// center in stage units. The method computes screen-space offsets and writes
// the results to the camera state.
func (c *Camera) Update(scl, x, y float32) {
	if sys.gsf(GSF_camerafreeze) {
		return
	}
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

// ScaleBound clamps a proposed zoom level.
//
// Parameters:
//   - scl: current scale factor (1.0 = no zoom).
//   - sclmul: multiplier applied to scl; values in [0,+inf).
//
// Returns the scale limited to [MinScale, zoomin]; if zooming is disabled, 1 is
// returned.
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

// XBound clamps the horizontal camera center.
//
// Parameters:
//   - scl: current scale.
//   - x: desired center position in stage units.
//
// Returns x clamped so that the viewport stays within stage bounds.
func (c *Camera) XBound(scl, x float32) float32 {
	return ClampF(x,
		c.boundL-c.halfWidth+c.halfWidth/scl,
		c.boundR+c.halfWidth-c.halfWidth/scl)
}

// BaseScale returns the stage's top level scale factor used as baseline for
// zoom calculations.
func (c *Camera) BaseScale() float32 {
	return c.ztopscale
}

// GroundLevel reports the Y coordinate of the stage floor in stage units after
// corrections for aspect ratio and zoom anchoring.
func (c *Camera) GroundLevel() float32 {
	return c.zoff - c.aspectcorrection - c.zoomanchorcorrection
}

// ResetZoomdelay clears the internal delay before zooming starts.
func (c *Camera) ResetZoomdelay() {
	c.zoomdelay = 0
}

func (c *Camera) action(x, y, scale float32, pause bool) (newX, newY, newScale float32) {
	newX = x
	newY = y
	newScale = scale
	if !sys.debugPaused() {
		newY = y / scale
		switch c.View {
		case Fighting_View:

			c.SaveRestoreTracking()

			if c.lowestcap {
				c.lowest = MaxF(c.lowest, float32(c.boundhigh)*c.localscl-(float32(sys.gameHeight)-c.GroundLevel()-float32(c.tensionlow))/c.zoomout)
			}
			tension := MaxF(0, float32(c.tension)*c.localscl)
			oldLeft, oldRight := x-c.halfWidth/scale, x+c.halfWidth/scale // previous frame bounds
			targetLeft, targetRight := oldLeft, oldRight                  // desired bounds this frame
			if c.autocenter {
				targetLeft = MinF(MaxF((c.leftest+c.rightest)/2-c.halfWidth/scale, c.minLeft), c.maxRight-2*c.halfWidth/scale)
				targetRight = targetLeft + 2*c.halfWidth/scale
			}

			if c.leftest < targetLeft+tension {
				// Push left bound so leftmost fighter stays within tension margin
				diff := targetLeft - MaxF(c.leftest-tension, c.minLeft)
				targetLeft = MaxF(c.leftest-tension, c.minLeft)
				targetRight = MaxF(oldRight-diff, MinF(c.rightest+tension, c.maxRight))
			} else if c.rightest > targetRight-tension {
				// Same for right edge
				diff := targetRight - MinF(c.rightest+tension, c.maxRight)
				targetRight = MinF(c.rightest+tension, c.maxRight)
				targetLeft = MinF(oldLeft-diff, MaxF(c.leftest-tension, c.minLeft))
			}
			if c.halfWidth*2/(targetRight-targetLeft) < c.zoomout { // prevent zooming out beyond limit
				rLeft := MaxF(targetLeft+tension-c.leftest, 0)
				rRight := MaxF(c.rightest-(targetRight-tension), 0)
				diff := 2 * ((targetRight-targetLeft)/2 - c.halfWidth/c.zoomout)
				if rLeft > rRight {
					diff2 := rLeft - rRight
					targetRight -= MinF(diff2, diff)
					diff -= MinF(diff2, diff)
				} else if rRight > rLeft {
					diff2 := rRight - rLeft
					targetLeft += MinF(diff2, diff)
					diff -= MinF(diff2, diff)
				}
				targetLeft += diff / 2
				targetRight -= diff / 2
				if c.leftest-targetLeft < float32(sys.stage.screenleft)*c.localscl {
					diff := MinF(float32(sys.stage.screenleft)*c.localscl-(c.leftest-targetLeft), targetLeft-c.minLeft)
					if targetRight-c.rightest < float32(sys.stage.screenright)*c.localscl {
						diff2 := MinF(float32(sys.stage.screenright)*c.localscl-(targetRight-c.rightest), c.maxRight-targetRight)
						//diff = diff + (MinF(float32(sys.stage.screenright)*c.localscl-(targetRight-c.rightest), c.maxRight-targetRight)-diff)/2
						diff = diff - diff2
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
						tmp := diffLeft / (diffLeft - diffRight) * ((targetRight + diffRight) - (targetLeft + diffLeft) - c.halfWidth*2/maxScale)
						tmp2 := diffRight / (diffLeft - diffRight) * ((targetRight + diffRight) - (targetLeft + diffLeft) - c.halfWidth*2/maxScale)
						diffLeft += tmp
						diffRight += tmp2
					}
					if c.halfWidth*2/((targetRight+diffRight)-(targetLeft+diffLeft)) > scale {
						targetLeft += diffLeft
						targetRight += diffRight
					} else {
						c.zoomindelaytime = c.zoomindelay
					}
				}
			} else {
				c.zoomindelaytime = c.zoomindelay
			}

			targetX := (targetLeft + targetRight) / 2
			targetScale := MinF(c.halfWidth*2/(targetRight-targetLeft), maxScale) // proposed zoom level

			if !c.ytensionenable {
				//newY = c.ywithoutbound
				ywithoutbound := c.ywithoutbound
				verticalfollow := MaxF(c.verticalfollow, 0.0) + (targetScale-c.zoomout)*MaxF(c.verticalfollowzoomdelta, 0.0)
				targetY := (c.highest + float32(c.floortension)*c.localscl) * verticalfollow
				if !c.roundstart {
					for i := 0; i < 3; i++ {
						ywithoutbound = ywithoutbound*.85 + targetY*.15 // exponential smoothing
						if AbsF(targetY-ywithoutbound)*sys.heightScale < 1 {
							ywithoutbound = targetY
						}
						if AbsF(newY-ywithoutbound) < float32(sys.gameWidth)/320*5.5 {
							newY = ywithoutbound
						} else {
							if newY > ywithoutbound {
								newY -= float32(sys.gameWidth) / 320 * 0.5
								newY -= (newY - ywithoutbound) * verticalfollow / 10
							} else {
								newY += float32(sys.gameWidth) / 320 * 0.5
								newY += (ywithoutbound - newY) * verticalfollow / 10
							}
						}
					}
				} else {
					ywithoutbound = targetY
					newY = ywithoutbound
				}
				c.ywithoutbound = ywithoutbound
			} else {
				targetScale = MinF(MinF(MaxF(float32(sys.gameHeight)/((c.lowest+float32(c.tensionlow)*c.localscl)-(c.highest-float32(c.tensionhigh)*c.localscl)), c.zoomout), c.zoomin), targetScale)
				targetX = MinF(MaxF(targetX, float32(c.boundleft)*c.localscl-c.halfWidth*(1/c.zoomout-1/targetScale)), float32(c.boundright)*c.localscl+c.halfWidth*(1/c.zoomout-1/targetScale))
				targetLeft = targetX - c.halfWidth/targetScale
				targetRight = targetX + c.halfWidth/targetScale

				newY = c.ywithoutbound
				targetY := c.GroundLevel()/targetScale + (c.highest - float32(c.tensionhigh)*c.localscl)
				if !c.roundstart {
					diff := float32(sys.gameWidth) / 320 * 2.5
					for i := 0; i < 3; i++ {
						newY = (newY + targetY) * .5
						if AbsF(targetY-newY) < diff {
							newY = targetY
							break
						} else if targetY-newY > diff {
							newY = newY + diff
						} else {
							newY = newY - diff
						}
					}
				} else {
					newY = targetY
				}
				c.ywithoutbound = newY
			}

			newLeft, newRight := oldLeft, oldRight
			if !c.roundstart {
				diff := float32(sys.gameWidth) / 3200
				for i := 0; i < 3; i++ {
					// iterative smoothing towards target bounds
					newLeft = newLeft + (targetLeft-newLeft)*0.05*sys.turbo*c.tensionvel
					newRight = newRight + (targetRight-newRight)*0.05*sys.turbo*c.tensionvel
					diffLeft := targetLeft - newLeft
					diffRight := targetRight - newRight

					if AbsF(diffLeft) <= diff*sys.turbo*c.tensionvel {
						newLeft = targetLeft
					} else if diffLeft > 0 {
						newLeft += diff * sys.turbo * c.tensionvel
					} else {
						newLeft -= diff * sys.turbo * c.tensionvel
					}
					if newLeft-oldLeft > 0 && newLeft-oldLeft < c.rightestvel {
						newLeft = MinF(oldLeft+c.rightestvel, targetLeft)
					} else if newLeft-oldLeft < 0 && newLeft-oldLeft > c.leftestvel {
						newLeft = MaxF(oldLeft+c.leftestvel, targetLeft)
					}

					if AbsF(diffRight) <= diff*sys.turbo*c.tensionvel {
						newRight = targetRight
					} else if diffRight > 0 {
						newRight += diff * sys.turbo * c.tensionvel
					} else {
						newRight -= diff * sys.turbo * c.tensionvel
					}
					if newRight-oldRight > 0 && newRight-oldRight < c.rightestvel {
						newRight = MinF(oldRight+c.rightestvel, targetRight)
					} else if newRight-oldRight < 0 && newRight-oldRight > c.leftestvel {
						newRight = MaxF(oldRight+c.leftestvel, targetRight)
					}
				}
			} else {
				newLeft, newRight = targetLeft, targetRight
			}
			newScale = MinF(c.halfWidth*2/(newRight-newLeft), c.zoomin)
			newLeft, newRight, newScale = c.reduceZoomSpeed(newLeft, newRight, newScale, oldLeft, oldRight, scale)
			newX = (newLeft + newRight) / 2
			newY = c.reduceYScrollSpeed(newY, y)
			newY = c.boundY(newY, newScale)

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

// reduceZoomSpeed linearly interpolates between the old and new zoom values and
// distributes the required boundary adjustments. It returns the adjusted left,
// right and scale values.
//
// [PITFALL] Setting zoominspeed close to 1 triggers very fast zooming which may
// cause visible jitter compared to M.U.G.E.N's more conservative camera.
func (c *Camera) reduceZoomSpeed(newLeft float32, newRight float32, newScale float32, oldLeft float32, oldRight float32, oldScale float32) (float32, float32, float32) {
	const minBoundDiff float32 = 5e-5
	const minScaleDiff float32 = 5e-4

	var speedFactor float32
	if newScale > oldScale {
		speedFactor = c.zoominspeed
	} else {
		speedFactor = c.zoomoutspeed
	}

	if speedFactor < 0.0 || speedFactor >= 1.0 {
		return newLeft, newRight, newScale
	}

	scaleDiff := newScale - oldScale
	leftAbsDiff, rightAbsDiff := AbsF(newLeft-oldLeft), AbsF(newRight-oldRight)

	if AbsF(scaleDiff) < minScaleDiff || (leftAbsDiff < minBoundDiff && rightAbsDiff < minBoundDiff) {
		return newLeft, newRight, newScale
	}

	// Interpolate scale then adjust bounds proportionally to maintain center
	adjustedNewScale := oldScale + speedFactor*scaleDiff
	scaleAdjustmentFactor := adjustedNewScale / newScale

	width := newRight - newLeft
	widthAdjustmentFactor := 1.0 / scaleAdjustmentFactor
	widthAdjustmentDiff := width*widthAdjustmentFactor - width

	totalAbsDiff := leftAbsDiff + rightAbsDiff
	adjustedNewLeft := newLeft - widthAdjustmentDiff*leftAbsDiff/totalAbsDiff
	adjustedNewRight := newRight + widthAdjustmentDiff*rightAbsDiff/totalAbsDiff

	adjustedNewLeft, adjustedNewRight = c.keepScreenEdge(adjustedNewLeft, adjustedNewRight)
	adjustedNewLeft, adjustedNewRight = c.keepStageEdge(adjustedNewLeft, adjustedNewRight)

	return c.hardLimit(adjustedNewLeft, adjustedNewRight)
}

func (c *Camera) keepScreenEdge(left float32, right float32) (float32, float32) {
	// Ensure both players remain inside their screen edge margins by
	// translating the bounds when necessary.
	screenLeftest := c.leftest - float32(sys.stage.screenleft)*c.localscl
	if left > screenLeftest {
		right += screenLeftest - left
		left = screenLeftest
	}

	screenRightest := c.rightest + float32(sys.stage.screenright)*c.localscl
	if right < screenRightest {
		left += screenRightest - right
		right = screenRightest
	}

	return left, right
}

func (c *Camera) keepStageEdge(left float32, right float32) (float32, float32) {
	// Prevent the viewport from leaving the stage bounds.
	if left < c.minLeft {
		right += c.minLeft - left
		left = c.minLeft
	}
	if right > c.maxRight {
		left += c.maxRight - right
		right = c.maxRight
	}
	return left, right
}

func (c *Camera) hardLimit(left float32, right float32) (float32, float32, float32) {
	// Final clamp ensuring stage limits and recomputing scale for new width.
	left = MaxF(left, c.minLeft)
	right = MinF(right, c.maxRight)
	scale := MaxF(MinF(c.halfWidth*2/(right-left), c.zoomin), c.zoomout)
	return left, right, scale
}

func (c *Camera) reduceYScrollSpeed(newY float32, oldY float32) float32 {
	// Apply smoothing to vertical movement so camera does not snap instantly.
	const minYDiff float32 = 5e-5

	yDiff := newY - oldY
	if AbsF(yDiff) < minYDiff || c.yscrollspeed < 0.0 || c.yscrollspeed >= 1.0 {
		return newY
	}

	return oldY + yDiff*c.yscrollspeed
}

func (c *Camera) boundY(y float32, scale float32) float32 {
	// Compute vertical bounds; when boundhighzoomdelta > 0 the top bound
	// relaxes with zoom to avoid showing outside the stage.
	if c.boundhighzoomdelta > 0 {
		topBound := float32(c.boundhigh)*c.localscl - c.GroundLevel()/c.zoomout
		boundHigh := float32(c.boundhigh)*c.localscl + ((topBound+c.GroundLevel()/scale)-float32(c.boundhigh)*c.localscl)/c.boundhighzoomdelta
		return MinF(MaxF(y, boundHigh), float32(c.boundlow)*c.localscl) * scale
	}
	return MinF(MaxF(y, float32(c.boundhigh)*c.localscl), float32(c.boundlow)*c.localscl) * scale
}

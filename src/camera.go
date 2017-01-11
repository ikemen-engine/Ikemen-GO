package main

type stageCamera struct {
	startx         int32
	boundleft      int32
	boundright     int32
	boundhigh      int32
	verticalfollow float32
	tension        int32
	floortension   int32
	overdrawlow    int32
	localcoord     [2]int32
	localscl       float32
	zoffset        int32
	ztopscale      float32
	drawOffsetY    float32
}

func newStageCamera() *stageCamera {
	return &stageCamera{verticalfollow: 0.2, tension: 50,
		localcoord: [2]int32{320, 240}, localscl: float32(sys.gameWidth / 320),
		ztopscale: 1}
}

type Camera struct {
	stageCamera
	ZoomEnable                  bool
	ZoomMin, ZoomMax, ZoomSpeed float32
	zoomdelay                   float32
	Pos, ScreenPos, Offset      [2]float32
	XMin, XMax                  float32
	Scale, MinScale             float32
	boundL, boundR, boundH      float32
	zoff                        float32
	screenZoff                  float32
	halfWidth                   float32
}

func newCamera() *Camera {
	return &Camera{ZoomMin: 1, ZoomMax: 1, ZoomSpeed: 1}
}
func (c *Camera) Init() {
	c.boundL = float32(c.boundleft-c.startx) * c.localscl
	c.boundR = float32(c.boundright-c.startx) * c.localscl
	if c.verticalfollow > 0 {
		c.boundH = MinF(0, float32(c.boundhigh)*c.localscl+
			float32(sys.gameHeight)-c.drawOffsetY-
			float32(sys.gameWidth)*float32(c.localcoord[1])/float32(c.localcoord[0]))
	} else {
		c.boundH = 0
	}
	if c.boundhigh > 0 {
		c.boundH += float32(c.boundhigh) * c.localscl
	}
	xminscl := float32(sys.gameWidth) / (float32(sys.gameWidth) - c.boundL +
		c.boundR)
	yminscl := float32(sys.gameHeight) / (240 - MinF(0, c.boundH))
	c.MinScale = MaxF(c.ZoomMin, MinF(c.ZoomMax, MaxF(xminscl, yminscl)))
	c.screenZoff = float32(c.zoffset)*c.localscl -
		c.drawOffsetY + 240 - float32(sys.gameWidth)*
		float32(c.localcoord[1])/float32(c.localcoord[0])
	c.halfWidth = float32(sys.gameWidth) / 2
}
func (c *Camera) Update(scl, x, y float32) {
	c.Scale = c.BaseScale() * scl
	c.zoff = scl*(float32(c.zoffset)*c.localscl-c.drawOffsetY+
		(240-float32(sys.gameWidth)*float32(c.localcoord[1])/
			float32(c.localcoord[0]))+float32(sys.gameHeight)-240) +
		(1-scl)*float32(sys.gameHeight)
	for i := 0; i < 2; i++ {
		c.Offset[i] = sys.stage.bga.offset[i] * sys.stage.localscl *
			sys.stage.scale[i] * scl
	}
	c.XMin = c.boundL - c.halfWidth/c.BaseScale()
	c.XMin = c.boundR + c.halfWidth/c.BaseScale()
	c.ScreenPos[0] = x - c.halfWidth/c.Scale - c.Offset[0]
	c.ScreenPos[1] = y - (c.GroundLevel()-float32(sys.gameHeight-240)*scl)/
		c.Scale - c.Offset[1]
	c.Pos[0] = x
	c.Pos[1] = y + float32(sys.gameHeight-240)
}
func (c *Camera) ScaleBound(scl float32) float32 {
	if c.ZoomEnable {
		return MaxF(c.MinScale, MinF(c.ZoomMax, scl))
	}
	return 1
}
func (c *Camera) XBound(scl, x float32) float32 {
	return MaxF(c.boundL-c.halfWidth+c.halfWidth/scl,
		MinF(c.boundR+c.halfWidth-c.halfWidth/scl, x))
}
func (c *Camera) YBound(scl, y float32) float32 {
	if c.verticalfollow <= 0 {
		return 0
	} else {
		tmp := MaxF(0, 240-c.screenZoff)
		return MaxF(0, c.boundH) + MinF(0, tmp*(1/scl-1),
			MaxF(c.boundH-240+MaxF(float32(sys.gameHeight)/scl,
				tmp+c.screenZoff/scl), y+240*(1-MinF(1, scl))))
	}
}
func (c *Camera) BaseScale() float32 {
	return c.ztopscale
}
func (c *Camera) GroundLevel() float32 {
	return c.zoff
}
func (c *Camera) ResetZoomdelay() {
	c.zoomdelay = 0
}

package main

import (
	"math"
)

type Fade struct {
	active        bool
	time          int32
	totalTime     int32
	overlayDelay  int32
	col           [3]int32
	colEncoded    uint32
	animData      *Anim
	isFadeIn      bool
	timeRemaining int32
	snd           [2]int32
}

func newFade() *Fade {
	return &Fade{
		time: 30,
		snd:  [2]int32{-1, 0},
	}
}

func (fa *Fade) reset() {
	if fa == nil {
		return
	}
	fa.active = false
	fa.timeRemaining = 0
	fa.totalTime = 0
	fa.overlayDelay = 0
	if fa.animData != nil {
		fa.animData.Reset()
	}
}

func (fa *Fade) animLength() int32 {
	if fa == nil || fa.animData == nil || fa.animData.anim == nil {
		return 0
	}
	return fa.animData.GetLength()
}

func (fa *Fade) colorFadeTime(animLen int32) int32 {
	if fa == nil {
		return 0
	}
	t := fa.time
	if t == 1 && animLen > 0 {
		t = animLen
	}
	if t < 0 {
		return 0
	}
	return t
}

func (fa *Fade) duration() int32 {
	if fa == nil {
		return 0
	}
	animLen := fa.animLength()
	total := fa.colorFadeTime(animLen)
	if animLen > total {
		total = animLen
	}
	return total
}

func (fa *Fade) init(dest *Fade, isFadeIn bool) {
	if fa == nil || dest == nil {
		return
	}
	animLen := fa.animLength()
	overlayTime := fa.colorFadeTime(animLen)
	totalTime := overlayTime
	if animLen > totalTime {
		totalTime = animLen
	}
	if totalTime <= 0 {
		dest.reset()
		return
	}
	*dest = *fa
	dest.time = overlayTime
	dest.totalTime = totalTime
	dest.overlayDelay = 0
	if !isFadeIn && overlayTime > 0 && animLen > overlayTime {
		// Let fadeout anim finish on the same frame as the color fade.
		dest.overlayDelay = animLen - overlayTime
	}
	dest.colEncoded = uint32(dest.col[0]&0xff<<16 | dest.col[1]&0xff<<8 | dest.col[2]&0xff)
	dest.isFadeIn = isFadeIn
	dest.active = true
	dest.timeRemaining = dest.totalTime
	if dest.animData != nil {
		dest.animData.Reset()
	}
}

func (fa *Fade) step() {
	if !fa.active || (sys.gameRunning && !sys.tickFrame() && !sys.motif.me.active) {
		return
	}
	if fa.timeRemaining == fa.totalTime && fa.snd[0] != -1 && fa.snd[1] != -1 {
		sys.motif.Snd.play(fa.snd, 100, 0, 0, 0, 0)
	}

	if fa.animData != nil {
		fa.animData.Update(false)
	}
	fa.timeRemaining--
	if fa.timeRemaining < 0 {
		fa.reset()
	}
}

func (fa *Fade) drawRect(rect [4]int32, color uint32, alpha int32) {
	alpha = int32(Clamp(alpha, int32(0), int32(255)))
	src := alpha>>uint(Btoi(sys.clsnDisplay)) + Btoi(sys.clsnDisplay)*128
	dst := int32(255) - src
	FillRect(rect, color, [2]int32{src, dst}, nil)
}

func (fa *Fade) draw() {
	if !fa.active || fa.timeRemaining < 0 || fa.totalTime <= 0 {
		return
	}
	if fa.animData != nil && fa.animData.anim != nil {
		fa.animData.Draw(fa.animData.layerno)
	}
	if fa.time <= 0 {
		return
	}
	elapsed := fa.totalTime - fa.timeRemaining - fa.overlayDelay
	if elapsed < 0 || elapsed > fa.time {
		return
	}
	if fa.isFadeIn {
		alpha := int32(math.Floor(float64(255) - 255*float64(elapsed-1)/float64(fa.time)))
		fa.drawRect(sys.scrrect, fa.colEncoded, alpha)
	} else {
		alpha := int32(math.Floor(255 * float64(elapsed) / float64(fa.time)))
		fa.drawRect(sys.scrrect, fa.colEncoded, alpha)
	}
}

func (fa *Fade) isFading() bool {
	return fa != nil && fa.isActive() && fa.timeRemaining > 0
}

func (fa *Fade) isActive() bool {
	return fa != nil && fa.active && fa.timeRemaining >= 0 && fa.totalTime > 0
}

// Policies for starting a new fade when a fade-in might be active.
//   - Continue  : start now, preserving the current visual darkness.
//   - Replace   : start now from full length, cancelling any fade-in immediately.
//   - Wait      : start after the current fade-in completes (no overlap, full length).
//   - Stop      : interrupt both current fade-in and skip fade-out.
type FadeStartPolicy int

const (
	FadeReplace FadeStartPolicy = iota
	FadeContinue
	FadeWait
	FadeStop
)

// startFadeOut starts an outgoing fade using the given policy.
// - overrideBlack: treat caller as a user interruption and force immediate cut for FadeStop.
// - policy: controls how an in-progress fade-in is handled.
func startFadeOut(tmpl *Fade, dest *Fade, overrideBlack bool, policy FadeStartPolicy) {
	if tmpl == nil || dest == nil {
		return
	}

	// Never mutate the template fade (important for rollback and for re-use across screens).
	use := tmpl
	var tmp Fade
	// On user interruption, force black/no-anim fade request.
	if overrideBlack {
		tmp = *tmpl
		tmp.col = [3]int32{0, 0, 0}
		tmp.animData = nil
		use = &tmp
	}
	fi := sys.motif.fadeIn

	// FadeStop semantics:
	// If this is an explicit user interruption OR a fade-in is active, cut immediately.
	if policy == FadeStop && (overrideBlack || (fi != nil && fi.isFading())) {
		if fi != nil {
			fi.reset()
		}
		if dest != nil {
			dest.reset()
			dest.timeRemaining = 0
		}
		return
	}

	// Helper to (re)start a full-length fade-out.
	startFresh := func() {
		use.init(dest, false)
	}

	// If no fade-in is active, all policies behave the same here: start now.
	if fi == nil || !fi.isFading() {
		startFresh()
		return
	}

	switch policy {
	case FadeReplace:
		fi.reset()
		startFresh()
	case FadeWait:
		// Defer: caller should call again after fade-in completes.
	case FadeContinue, FadeStop:
		startFresh()
		// Match current darkness, then cancel fade-in.
		trfi := fi.timeRemaining
		if trfi < 0 {
			trfi = 0
		}
		if fi.time > 0 && dest.time > 0 {
			dest.timeRemaining = dest.time - (dest.time*trfi)/fi.time
			if dest.timeRemaining < 0 {
				dest.timeRemaining = 0
			}
			if dest.timeRemaining > dest.time {
				dest.timeRemaining = dest.time
			}
		}
		fi.reset()
	}
}

type Rect struct {
	window     [4]int32
	col        uint32
	alpha      [2]int32
	time       int32
	layerno    int16
	localScale float32
	offsetX    int32
	// pulse src = clamp(mid + amp * sin(phase))
	pulseMid       float32
	pulseAmp       float32
	pulsePhase     float32
	pulsePhaseStep float32 // radians per frame = 2π / periodFrames
	autoAlpha      bool
	//palfx          *PalFX
	// initial, unscaled values
	windowInit [4]float32
}

func NewRect() *Rect {
	return &Rect{window: sys.scrrect, alpha: [2]int32{255, 0}, localScale: 1}
}

func packAlpha(src, dst int32) [2]int32 {
	return [2]int32{Clamp(src, 0, 255), Clamp(dst, 0, 255)}
}

func (r *Rect) updateAlpha() {
	if !r.autoAlpha {
		return
	}
	r.pulsePhase += r.pulsePhaseStep
	v := r.pulseMid + r.pulseAmp*float32(math.Sin(float64(r.pulsePhase)))
	src := Clamp(int32(math.Round(float64(v))), 0, 255)
	r.alpha = packAlpha(src, 255-src)
}

func (r *Rect) Draw(ln int16) {
	if r.layerno == ln && r != nil {
		FillRect(r.window, r.col, r.alpha, nil)
	}
}

func (r *Rect) Reset() {
	r.SetWindow(r.windowInit)
}

func (r *Rect) SetColor(col [3]int32) {
	r.col = uint32(col[2]&0xff | col[1]&0xff<<8 | col[0]&0xff<<16)
}

func (r *Rect) SetAlpha(alpha [2]int32) {
	r.alpha = packAlpha(alpha[0], alpha[1])
	r.autoAlpha = false
}

func (r *Rect) SetAlphaPulse(sMid, sAmp, sPeriod int32) {
	m := float32(Clamp(sMid, 0, 255))
	a := float32(Clamp(sAmp, 0, 255))
	if sPeriod <= 0 || a == 0 {
		r.pulseMid, r.pulseAmp = m, 0
		r.pulsePhaseStep = 0
		r.autoAlpha = false
		return
	}
	r.pulseMid, r.pulseAmp = m, a
	r.pulsePhaseStep = float32(2 * math.Pi / float64(sPeriod))
	r.autoAlpha = true
}

func (r *Rect) SetLocalcoord(lx, ly float32) {
	if lx <= 0 || ly <= 0 {
		return
	}
	v := lx
	if lx*3 > ly*4 {
		v = ly * 4 / 3
	}
	r.localScale = float32(v / 320)
	r.offsetX = -int32(math.Floor(float64(lx)/(float64(v)/320)-320) / 2)
}

func (r *Rect) SetWindow(window [4]float32) {
	if window == [4]float32{0, 0, 0, 0} {
		return
	}
	r.windowInit = window
	x := window[0]/r.localScale + float32(r.offsetX)
	y := window[1] / r.localScale
	w := (window[2] - window[0]) / r.localScale
	h := (window[3] - window[1]) / r.localScale
	r.window[0] = int32((x + float32(sys.gameWidth-320)/2) * sys.widthScale)
	r.window[1] = int32((y + float32(sys.gameHeight-240)) * sys.heightScale)
	r.window[2] = int32(w*sys.widthScale + 0.5)
	r.window[3] = int32(h*sys.heightScale + 0.5)
}

func (r *Rect) Update() {
	if r != nil {
		r.updateAlpha()
		//if r.palfx != nil {
		//	r.palfx.step()
		//}
	}
}

//func (r *Rect) SetPalFx(p *PalFX) {
//	r.palfx = p
//	if r.palfx != nil && r.palfx.time == 0 {
//		r.palfx.time = -1
//	}
//}

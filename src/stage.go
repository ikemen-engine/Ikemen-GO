package main

import (
	"math"
)

type EnvShake struct {
	time  int32
	freq  float32
	ampl  int32
	phase float32
}

func (es *EnvShake) clear() {
	*es = EnvShake{freq: float32(math.Pi / 3), ampl: -4,
		phase: float32(math.NaN())}
}
func (es *EnvShake) setDefPhase() {
	if math.IsNaN(float64(es.phase)) {
		if es.freq >= math.Pi/2 {
			es.phase = math.Pi / 2
		} else {
			es.phase = 0
		}
	}
}
func (es *EnvShake) next() {
	if es.time > 0 {
		es.time--
		es.phase += es.freq
	}
}
func (es *EnvShake) getOffset() float32 {
	if es.time > 0 {
		return float32(es.ampl) * 0.5 * float32(math.Sin(float64(es.phase)))
	}
	return 0
}
